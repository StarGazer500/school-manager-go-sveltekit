package teachers

import (
	"context"
	"database/sql"
	"time"

	"github.com/martin/school-api/internal/database"
)

type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Teacher, error)
	FindByID(ctx context.Context, id uint) (Teacher, error)
	Create(ctx context.Context, t *Teacher) error
	Update(ctx context.Context, t *Teacher) error
	Delete(ctx context.Context, id uint) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

const columns = `id, first_name, last_name, email, subject, created_at, updated_at`

func scanTeacher(row database.Scanner) (Teacher, error) {
	var t Teacher
	err := row.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.Subject, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *sqlRepository) List(ctx context.Context, f ListFilter) ([]Teacher, error) {
	query := `SELECT ` + columns + ` FROM teachers WHERE 1=1`
	var args []any
	if f.Search != "" {
		like := "%" + f.Search + "%"
		query += ` AND (first_name LIKE ? OR last_name LIKE ? OR subject LIKE ?)`
		args = append(args, like, like, like)
	}
	query += ` ORDER BY last_name, first_name`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Teacher
	for rows.Next() {
		t, err := scanTeacher(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *sqlRepository) FindByID(ctx context.Context, id uint) (Teacher, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM teachers WHERE id = ?`, id)
	t, err := scanTeacher(row)
	return t, database.MapError(err, "teacher")
}

func (r *sqlRepository) Create(ctx context.Context, t *Teacher) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO teachers (first_name, last_name, email, subject, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.FirstName, t.LastName, t.Email, t.Subject, now, now)
	if err != nil {
		return database.MapError(err, "teacher")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID, t.CreatedAt, t.UpdatedAt = uint(id), now, now
	return nil
}

func (r *sqlRepository) Update(ctx context.Context, t *Teacher) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx,
		`UPDATE teachers
		 SET first_name = ?, last_name = ?, email = ?, subject = ?, updated_at = ?
		 WHERE id = ?`,
		t.FirstName, t.LastName, t.Email, t.Subject, now, t.ID)
	if err != nil {
		return database.MapError(err, "teacher")
	}
	t.UpdatedAt = now
	return database.RequireAffected(res, "teacher")
}

func (r *sqlRepository) Delete(ctx context.Context, id uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM teachers WHERE id = ?`, id)
	if err != nil {
		return database.MapError(err, "teacher")
	}
	return database.RequireAffected(res, "teacher")
}
