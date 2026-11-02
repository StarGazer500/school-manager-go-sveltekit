package teachers

import "time"

// Teacher is the database entity for the "teachers" table.
type Teacher struct {
	ID        uint
	FirstName string
	LastName  string
	Email     string
	Subject   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListFilter narrows List results. Zero values mean "no filter".
type ListFilter struct {
	Search string
}
