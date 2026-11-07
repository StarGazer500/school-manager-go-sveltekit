-- The join table between students and courses. Deleting a student or a course
-- deletes their enrollments too (ON DELETE CASCADE). UNIQUE (student_id,
-- course_id) also serves lookups by student_id; course_id lookups (the roster,
-- the capacity count) get their own index.

-- +goose Up
CREATE TABLE IF NOT EXISTS enrollments (
	id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	student_id  BIGINT      NOT NULL REFERENCES students(id) ON DELETE CASCADE,
	course_id   BIGINT      NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
	enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (student_id, course_id)
);
CREATE INDEX IF NOT EXISTS enrollments_course_id_idx ON enrollments (course_id);

-- +goose Down
DROP TABLE enrollments;
