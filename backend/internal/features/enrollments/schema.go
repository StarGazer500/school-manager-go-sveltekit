package enrollments

// Schema creates the join table between students and courses. Deleting a
// student or a course deletes their enrollments too (ON DELETE CASCADE).
const Schema = `
CREATE TABLE IF NOT EXISTS enrollments (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	student_id  INTEGER  NOT NULL REFERENCES students(id) ON DELETE CASCADE,
	course_id   INTEGER  NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
	enrolled_at DATETIME NOT NULL,
	UNIQUE (student_id, course_id)
);`
