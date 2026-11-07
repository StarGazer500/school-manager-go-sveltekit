// Command migrate manages the database schema with goose, using the SQL files
// in backend/migrations. Run it from the backend folder (the Makefile does):
//
//	go run ./cmd/migrate up            apply all new migrations
//	go run ./cmd/migrate down          undo the most recent migration
//	go run ./cmd/migrate status        list migrations and whether they ran
//	go run ./cmd/migrate new NAME      create migrations/0000N_NAME.sql
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"

	"github.com/martin/school-api/internal/config"
	"github.com/martin/school-api/internal/database"
)

// Where `new` writes files. Relative to the backend folder.
const migrationsDir = "migrations"

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	command, args := os.Args[1], os.Args[2:]

	// `new` only writes a file, so it doesn't need a database.
	if command == "new" {
		if len(args) != 1 {
			usage()
		}
		goose.SetSequential(true) // 00005_name.sql instead of a timestamp
		if err := goose.Create(nil, migrationsDir, args[0], "sql"); err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx := context.Background()
	pool, err := database.Open(ctx, config.Load().DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	p, close, err := database.Migrator(pool)
	if err != nil {
		log.Fatal(err)
	}
	defer close()

	switch command {
	case "up":
		results, err := p.Up(ctx)
		for _, r := range results {
			fmt.Println(r)
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(results) == 0 {
			fmt.Println("nothing to apply: the database is up to date")
		}
	case "down":
		r, err := p.Down(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(r)
	case "status":
		statuses, err := p.Status(ctx)
		if err != nil {
			log.Fatal(err)
		}
		for _, s := range statuses {
			applied := "pending"
			if !s.AppliedAt.IsZero() {
				applied = "applied " + s.AppliedAt.Local().Format("2006-01-02 15:04")
			}
			fmt.Printf("%05d  %-36s %s\n", s.Source.Version, baseName(s.Source.Path), applied)
		}
	default:
		usage()
	}
}

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate up | down | status | new NAME")
	os.Exit(2)
}
