package courses

// Schema creates the table this feature owns. It must run after the teachers
// schema because of the foreign key.
const Schema = `
CREATE TABLE IF NOT EXISTS courses (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	code        TEXT     NOT NULL UNIQUE,
	name        TEXT     NOT NULL,
	description TEXT     NOT NULL DEFAULT '',
	capacity    INTEGER  NOT NULL,
	teacher_id  INTEGER  REFERENCES teachers(id) ON DELETE SET NULL,
	created_at  DATETIME NOT NULL,
	updated_at  DATETIME NOT NULL
);`
