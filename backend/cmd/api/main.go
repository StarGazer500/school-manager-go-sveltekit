package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/martin/school-api/internal/config"
	"github.com/martin/school-api/internal/database"
	"github.com/martin/school-api/internal/features/courses"
	"github.com/martin/school-api/internal/features/enrollments"
	"github.com/martin/school-api/internal/features/students"
	"github.com/martin/school-api/internal/features/teachers"
	"github.com/martin/school-api/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Order matters: tables with foreign keys come after the tables they reference.
	if err := database.Migrate(context.Background(), db, students.Schema, teachers.Schema, courses.Schema, enrollments.Schema); err != nil {
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

	go func() {
		log.Printf("API listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
