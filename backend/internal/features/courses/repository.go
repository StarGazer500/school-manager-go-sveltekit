package courses

import (
	"context"
	"database/sql"
	"time"

	"github.com/martin/school-api/internal/database"
)

type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Course, error)
	FindByID(ctx context.Context, id uint) (Course, error)
	Create(ctx context.Context, c *Course) error
	Update(ctx context.Context, c *Course) error
	Delete(ctx context.Context, id uint) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

const columns = `id, code, name, description, capacity, teacher_id, created_at, updated_at`

// TeacherID is a *uint: database/sql scans NULL into nil and a number into a
// new pointer, and passing a nil pointer as an argument writes NULL.
func scanCourse(row database.Scanner) (Course, error) {
	var c Course
	err := row.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Capacity, &c.TeacherID, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *sqlRepository) List(ctx context.Context, f ListFilter) ([]Course, error) {
	query := `SELECT ` + columns + ` FROM courses WHERE 1=1`
	var args []any
	if f.Search != "" {
		like := "%" + f.Search + "%"
		query += ` AND (code LIKE ? OR name LIKE ?)`
		args = append(args, like, like)
	}
	if f.TeacherID != 0 {
		query += ` AND teacher_id = ?`
		args = append(args, f.TeacherID)
	}
	query += ` ORDER BY code`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Course
	for rows.Next() {
		c, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *sqlRepository) FindByID(ctx context.Context, id uint) (Course, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM courses WHERE id = ?`, id)
	c, err := scanCourse(row)
	return c, database.MapError(err, "course")
}

// Create and Update rely on the database's foreign key: an unknown teacherId
// fails with a FOREIGN KEY constraint error, which MapError turns into a 400.
func (r *sqlRepository) Create(ctx context.Context, c *Course) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO courses (code, name, description, capacity, teacher_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.Code, c.Name, c.Description, c.Capacity, c.TeacherID, now, now)
	if err != nil {
		return database.MapError(err, "course")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	c.ID, c.CreatedAt, c.UpdatedAt = uint(id), now, now
	return nil
}

func (r *sqlRepository) Update(ctx context.Context, c *Course) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`UPDATE courses
		 SET code = ?, name = ?, description = ?, capacity = ?, teacher_id = ?, updated_at = ?
		 WHERE id = ?`,
		c.Code, c.Name, c.Description, c.Capacity, c.TeacherID, now, c.ID)
	if err != nil {
		return database.MapError(err, "course")
	}
	c.UpdatedAt = now
	return database.RequireAffected(res, "course")
}

func (r *sqlRepository) Delete(ctx context.Context, id uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM courses WHERE id = ?`, id)
	if err != nil {
		return database.MapError(err, "course")
	}
	return database.RequireAffected(res, "course")
}
