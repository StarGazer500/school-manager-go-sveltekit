package courses

import "time"

// Course is the database entity for the "courses" table.
// TeacherID is a pointer because a course may have no teacher (NULL).
type Course struct {
	ID          int64     `db:"id"`
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Capacity    int       `db:"capacity"`
	TeacherID   *int64    `db:"teacher_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (Course) TableName() string { return "courses" }

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	Search    string
	TeacherID int64
}
