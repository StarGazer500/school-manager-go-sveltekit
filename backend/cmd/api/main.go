package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/martin/school-api/internal/config"
	"github.com/martin/school-api/internal/database"
	"github.com/martin/school-api/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Apply any new SQL files from backend/migrations (see internal/database/migrate.go).
	if err := database.Migrate(context.Background(), db); err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(db),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Stop cleanly on Ctrl+C: finish in-flight requests, then exit.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Claim the port before saying we're listening, so a port that's already
	// taken fails here with a clear message.
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalf("can't listen on port %s (is something else using it? set PORT=...): %v", cfg.Port, err)
	}

	go func() {
		log.Printf("API listening on http://localhost:%s", cfg.Port)
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
