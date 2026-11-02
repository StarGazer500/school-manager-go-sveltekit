package courses

import "time"

// Course is the database entity for the "courses" table.
// TeacherID is a pointer because a course may have no teacher (NULL).
type Course struct {
	ID          uint
	Code        string
	Name        string
	Description string
	Capacity    int
	TeacherID   *uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	Search    string
	TeacherID uint
}
