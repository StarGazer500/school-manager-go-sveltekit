package enrollments

import "time"

// Enrollment is the database entity for the "enrollments" table.
type Enrollment struct {
	ID         uint
	StudentID  uint
	CourseID   uint
	EnrolledAt time.Time
}

// Detail is a read model: an enrollment joined with the student and course
// names, so clients don't need extra requests to display it.
type Detail struct {
	ID          uint
	StudentID   uint
	StudentName string
	CourseID    uint
	CourseCode  string
	CourseName  string
	EnrolledAt  time.Time
}

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	StudentID uint
	CourseID  uint
}
