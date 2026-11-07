package students

import "time"

// Student is the database entity: one row of the "students" table. The db
// tags tell pgrepo which column each field maps to; the database fills
// ID, CreatedAt and UpdatedAt.
type Student struct {
	ID         int64     `db:"id"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Email      string    `db:"email"`
	GradeLevel int       `db:"grade_level"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// TableName is how pgrepo knows which table Student lives in.
func (Student) TableName() string { return "students" }

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	Search     string
	GradeLevel int
}
