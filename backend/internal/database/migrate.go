package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/martin/school-api/migrations"
)

// Migrator returns a goose provider for the SQL files in backend/migrations.
// goose talks to the database through database/sql, so the pgx pool is
// wrapped in a *sql.DB; it borrows the pool's connections instead of opening
// its own. Call close when done (it does not close the pool).
//
// goose records what has run in a goose_db_version table, so each migration
// is applied once, in version order.
func Migrator(pool *pgxpool.Pool) (p *goose.Provider, close func() error, err error) {
	db := stdlib.OpenDBFromPool(pool)
	p, err = goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("load migrations: %w", err)
	}
	return p, db.Close, nil
}

// Migrate applies every migration that hasn't run yet. The API calls it at
// startup, so a fresh database gets all tables and an existing one gets
// only the new migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	p, close, err := Migrator(pool)
	if err != nil {
		return err
	}
	defer close()

	results, err := p.Up(ctx)
	for _, r := range results {
		log.Printf("migration %s", r)
	}
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
