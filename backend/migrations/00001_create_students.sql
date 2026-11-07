-- The first four migrations use IF NOT EXISTS because databases created before
-- goose was added already have these tables; goose then just records them as
-- applied. New migrations don't need it: goose runs each one exactly once.

-- +goose Up
CREATE TABLE IF NOT EXISTS students (
	id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	first_name  TEXT        NOT NULL,
	last_name   TEXT        NOT NULL,
	email       TEXT        NOT NULL UNIQUE,
	grade_level INTEGER     NOT NULL,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE students;
