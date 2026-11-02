// Package database opens the SQLite connection and holds the small helpers
// every repository uses. Repositories write plain SQL with database/sql.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/martin/school-api/internal/shared/apperr"
)

// Open connects to the SQLite file at path.
//   - foreign_keys(1): SQLite ignores REFERENCES unless this is on, per connection.
//   - busy_timeout: wait instead of failing immediately if the file is locked.
//   - _time_format=sqlite: store time.Time as "2006-01-02 15:04:05.999-07:00",
//     which SQLite's own date functions understand.
func Open(path string) (*sql.DB, error) {
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_time_format=sqlite"
	db, err := sql.Open("sqlite", dsn) // "sqlite" is the name modernc.org/sqlite registers
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// sql.Open doesn't connect yet; Ping makes sure the file is usable.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// Migrate runs each feature's schema in order. Every statement uses
// IF NOT EXISTS, so running it on every startup is safe.
func Migrate(ctx context.Context, db *sql.DB, schemas ...string) error {
	for _, schema := range schemas {
		if _, err := db.ExecContext(ctx, schema); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// Scanner is satisfied by both *sql.Row and *sql.Rows, so one scan function
// per entity works for single-row and multi-row queries.
type Scanner interface {
	Scan(dest ...any) error
}

// MapError converts database errors into apperr kinds. entity is used in the
// message, e.g. "student not found".
func MapError(err error, entity string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return apperr.NotFound("%s not found", entity)
	case IsUniqueViolation(err):
		return apperr.Conflict("%s already exists", entity)
	case IsForeignKeyViolation(err):
		return apperr.Invalid("%s references a record that does not exist", entity)
	default:
		return err
	}
}

// RequireAffected turns "the UPDATE/DELETE matched no rows" into a not-found
// error. Exec doesn't return sql.ErrNoRows, so we check the count ourselves.
func RequireAffected(res sql.Result, entity string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperr.NotFound("%s not found", entity)
	}
	return nil
}

func IsUniqueViolation(err error) bool {
	code := sqliteCode(err)
	return code == sqlite3.SQLITE_CONSTRAINT_UNIQUE || code == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY
}

func IsForeignKeyViolation(err error) bool {
	return sqliteCode(err) == sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY
}

func sqliteCode(err error) int {
	var se *sqlite.Error
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}
