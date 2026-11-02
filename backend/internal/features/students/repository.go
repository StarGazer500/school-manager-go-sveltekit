package students

import (
	"context"
	"database/sql"
	"time"

	"github.com/martin/school-api/internal/database"
)

// Repository is the persistence contract the service depends on. Because it
// is an interface, tests can swap in a fake without touching a database.
type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Student, error)
	FindByID(ctx context.Context, id uint) (Student, error)
	Create(ctx context.Context, s *Student) error
	Update(ctx context.Context, s *Student) error
	Delete(ctx context.Context, id uint) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

// columns and scanStudent must list the fields in the same order.
const columns = `id, first_name, last_name, email, grade_level, created_at, updated_at`

func scanStudent(row database.Scanner) (Student, error) {
	var s Student
	err := row.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.GradeLevel, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *sqlRepository) List(ctx context.Context, f ListFilter) ([]Student, error) {
	// Build the WHERE clause from whichever filters are set. Values always go
	// in args as ? placeholders, never into the SQL string (SQL injection).
	query := `SELECT ` + columns + ` FROM students WHERE 1=1`
	var args []any
	if f.Search != "" {
		like := "%" + f.Search + "%"
		query += ` AND (first_name LIKE ? OR last_name LIKE ? OR email LIKE ?)`
		args = append(args, like, like, like)
	}
	if f.GradeLevel != 0 {
		query += ` AND grade_level = ?`
		args = append(args, f.GradeLevel)
	}
	query += ` ORDER BY last_name, first_name`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Student
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *sqlRepository) FindByID(ctx context.Context, id uint) (Student, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM students WHERE id = ?`, id)
	s, err := scanStudent(row)
	return s, database.MapError(err, "student") // sql.ErrNoRows -> 404
}

func (r *sqlRepository) Create(ctx context.Context, s *Student) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO students (first_name, last_name, email, grade_level, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		s.FirstName, s.LastName, s.Email, s.GradeLevel, now, now)
	if err != nil {
		return database.MapError(err, "student") // UNIQUE(email) -> 409
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	s.ID, s.CreatedAt, s.UpdatedAt = uint(id), now, now
	return nil
}

func (r *sqlRepository) Update(ctx context.Context, s *Student) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`UPDATE students
		 SET first_name = ?, last_name = ?, email = ?, grade_level = ?, updated_at = ?
		 WHERE id = ?`,
		s.FirstName, s.LastName, s.Email, s.GradeLevel, now, s.ID)
	if err != nil {
		return database.MapError(err, "student")
	}
	s.UpdatedAt = now
	return database.RequireAffected(res, "student")
}

func (r *sqlRepository) Delete(ctx context.Context, id uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM students WHERE id = ?`, id)
	if err != nil {
		return database.MapError(err, "student")
	}
	return database.RequireAffected(res, "student")
}
