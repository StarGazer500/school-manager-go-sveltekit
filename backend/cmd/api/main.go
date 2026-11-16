package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/martin/school-api/internal/core"
	"github.com/martin/school-api/internal/logger"
	"github.com/martin/school-api/internal/server"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1) // run has already reported the error
	}
}

// run starts everything and blocks until Ctrl+C or SIGTERM. It returns an
// error instead of exiting on the spot, so its deferred cleanup always runs,
// in reverse order: stop serving, close Redis and Postgres, flush the traces
// still buffered, and last of all close the log file.
func run() error {
	cfg := core.LoadConfig()
	ctx := context.Background()

	log, closeLog, err := logger.New(logger.Options{
		Format: cfg.LogFormat, Level: cfg.LogLevel, File: cfg.LogFile,
		Service: cfg.ServiceName, Version: cfg.ServiceVersion, Env: cfg.Environment,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // no logger yet, so plain stderr
		return err
	}
	defer closeLog()

	// In JSON mode every line should be JSON, so silence Gin's startup banner
	// and route list (unless GIN_MODE says otherwise).
	if cfg.LogFormat == "json" && os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	tel, err := core.SetupTelemetry(ctx, cfg)
	if err != nil {
		log.Error().Err(err).Msg("telemetry")
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tel.Shutdown(flushCtx)
	}()
	core.ReportTelemetryErrors(log)

	log.Info().
		Bool("telemetry", cfg.TelemetryEnabled).Str("otlp_endpoint", cfg.OTLPEndpoint).
		Str("log_file", cfg.LogFile).Msg("starting")

	// The logger is passed explicitly to everything set up here (dependency
	// injection); during requests it travels in the request's context instead.
	db, err := core.OpenDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error().Err(err).Msg("database")
		return err
	}
	defer db.Close()

	// Apply any new SQL files from backend/migrations (see internal/core/migrate.go).
	if err := core.Migrate(ctx, db, log); err != nil {
		log.Error().Err(err).Msg("migrations")
		return err
	}

	// Redis is only a cache: if it's down, the app still starts and reads go
	// to Postgres (see internal/core/cache.go).
	rdb, err := core.OpenRedis(ctx, cfg.RedisURL, log)
	if err != nil {
		log.Error().Err(err).Msg("redis")
		return err
	}
	defer rdb.Close()

	// Two servers: the public API, and /metrics on its own port for
	// Prometheus. Keeping metrics off the API port means they're never
	// exposed to the internet along with the API.
	api := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(db, rdb, log, cfg.ServiceName),
		ReadHeaderTimeout: 5 * time.Second,
	}
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", tel.MetricsHandler)
	metrics := &http.Server{Addr: ":" + cfg.MetricsPort, Handler: metricsMux, ReadHeaderTimeout: 5 * time.Second}

	// Stop cleanly on Ctrl+C: finish in-flight requests, then exit.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Claim the ports before saying we're listening, so a port that's already
	// taken fails here with a clear message.
	apiLn, err := net.Listen("tcp", api.Addr)
	if err != nil {
		log.Error().Err(err).Str("port", cfg.Port).Msg("can't listen on port (is something else using it? set PORT=...)")
		return err
	}
	metricsLn, err := net.Listen("tcp", metrics.Addr)
	if err != nil {
		apiLn.Close()
		log.Error().Err(err).Str("port", cfg.MetricsPort).Msg("can't listen on metrics port (set METRICS_PORT=...)")
		return err
	}

	serveErr := make(chan error, 2)
	serve := func(srv *http.Server, ln net.Listener) {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}
	go serve(api, apiLn)
	go serve(metrics, metricsLn)
	log.Info().Str("url", "http://localhost:"+cfg.Port).
		Str("metrics", "http://localhost:"+cfg.MetricsPort+"/metrics").Msg("API listening")

	select {
	case <-ctx.Done():
	case err := <-serveErr:
		log.Error().Err(err).Msg("server stopped")
		return err
	}

	log.Info().Msg("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := api.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown")
	}
	_ = metrics.Shutdown(shutdownCtx)
	return nil
}
