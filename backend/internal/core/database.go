package core

// PostgreSQL: opening the connection pool and mapping database errors to app
// errors. Migrations are in migrate.go. Repositories use pgrepo on top of the pool.

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/martin/school-api/internal/shared"
)

// PostgreSQL error codes we care about (https://www.postgresql.org/docs/current/errcodes-appendix.html).
const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// OpenDatabase creates a connection pool for url and checks the database is reachable.
func OpenDatabase(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// pgxpool.New connects lazily; Ping makes a bad URL or a stopped server fail at startup.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database (is it running? try `make db-up`): %w", err)
	}
	return pool, nil
}

// MapError converts database errors into shared error kinds. entity is used in the
// message, e.g. "student already exists".
func MapError(err error, entity string) error {
	switch {
	case err == nil:
		return nil
	case IsUniqueViolation(err):
		return shared.Conflict("%s already exists", entity)
	case IsForeignKeyViolation(err):
		return shared.Invalid("%s references a record that does not exist", entity)
	default:
		return err
	}
}

func IsUniqueViolation(err error) bool { return pgCode(err) == uniqueViolation }

func IsForeignKeyViolation(err error) bool { return pgCode(err) == foreignKeyViolation }

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
