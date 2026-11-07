-- Runs after teachers because of the foreign key. PostgreSQL doesn't index
-- foreign keys on its own, so the "courses taught by" lookup gets an index.

-- +goose Up
CREATE TABLE IF NOT EXISTS courses (
	id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	code        TEXT        NOT NULL UNIQUE,
	name        TEXT        NOT NULL,
	description TEXT        NOT NULL DEFAULT '',
	capacity    INTEGER     NOT NULL,
	teacher_id  BIGINT      REFERENCES teachers(id) ON DELETE SET NULL,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS courses_teacher_id_idx ON courses (teacher_id);

-- +goose Down
DROP TABLE courses;
