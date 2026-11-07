// Package migrations holds the database schema as numbered SQL files, applied
// in order by goose. Never edit a migration that has already run somewhere:
// add a new one instead (`make migrate-new name=add_phone_to_students`).
//
// The files are embedded into the binary, so the API can migrate the
// database at startup without the .sql files next to it.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
