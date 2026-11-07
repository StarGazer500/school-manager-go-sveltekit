package teachers

import "time"

// Teacher is the database entity for the "teachers" table.
type Teacher struct {
	ID        int64     `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Subject   string    `db:"subject"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (Teacher) TableName() string { return "teachers" }

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	Search string
}
