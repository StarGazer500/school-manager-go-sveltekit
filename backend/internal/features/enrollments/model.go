package enrollments

import "time"

// Enrollment is the database entity for the "enrollments" table.
type Enrollment struct {
	ID         int64     `db:"id"`
	StudentID  int64     `db:"student_id"`
	CourseID   int64     `db:"course_id"`
	EnrolledAt time.Time `db:"enrolled_at"`
}

func (Enrollment) TableName() string { return "enrollments" }

// Detail is a read model: an enrollment plus the student and course names,
// so clients don't need extra requests to display it. It isn't a table; the
// repository assembles it.
type Detail struct {
	ID          int64
	StudentID   int64
	StudentName string
	CourseID    int64
	CourseCode  string
	CourseName  string
	EnrolledAt  time.Time
}

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	StudentID int64
	CourseID  int64
}
