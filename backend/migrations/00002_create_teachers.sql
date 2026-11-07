-- +goose Up
CREATE TABLE IF NOT EXISTS teachers (
	id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	first_name TEXT        NOT NULL,
	last_name  TEXT        NOT NULL,
	email      TEXT        NOT NULL UNIQUE,
	subject    TEXT        NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE teachers;
