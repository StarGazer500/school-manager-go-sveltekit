package students

import "time"

// Student is the database entity: one row of the "students" table. The
// repository sets CreatedAt/UpdatedAt when it inserts or updates.
type Student struct {
	ID         uint
	FirstName  string
	LastName   string
	Email      string
	GradeLevel int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	Search     string
	GradeLevel int
}
