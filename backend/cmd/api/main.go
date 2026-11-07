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
	cfg := core.LoadConfig()

	log, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // no logger yet, so plain stderr
		os.Exit(1)
	}
	// In JSON mode every line should be JSON, so silence Gin's startup banner
	// and route list (unless GIN_MODE says otherwise).
	if cfg.LogFormat == "json" && os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// The logger is passed explicitly to everything set up here (dependency
	// injection); during requests it travels in the request's context instead.
	ctx := context.Background()

	db, err := core.OpenDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("database") // Fatal logs, then exits with status 1
	}
	defer db.Close()

	// Apply any new SQL files from backend/migrations (see internal/core/migrate.go).
	if err := core.Migrate(ctx, db, log); err != nil {
		log.Fatal().Err(err).Msg("migrations")
	}

	// Redis is only a cache: if it's down, the app still starts and reads go
	// to Postgres (see internal/core/cache.go).
	rdb, err := core.OpenRedis(ctx, cfg.RedisURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("redis")
	}
	defer rdb.Close()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(db, rdb, log),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Stop cleanly on Ctrl+C: finish in-flight requests, then exit.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Claim the port before saying we're listening, so a port that's already
	// taken fails here with a clear message.
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatal().Err(err).Str("port", cfg.Port).Msg("can't listen on port (is something else using it? set PORT=...)")
	}

	go func() {
		log.Info().Str("url", "http://localhost:"+cfg.Port).Msg("API listening")
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server stopped")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown")
	}
}
