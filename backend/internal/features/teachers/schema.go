package teachers

// Schema creates the table this feature owns.
const Schema = `
CREATE TABLE IF NOT EXISTS teachers (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	first_name TEXT     NOT NULL,
	last_name  TEXT     NOT NULL,
	email      TEXT     NOT NULL UNIQUE,
	subject    TEXT     NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);`
