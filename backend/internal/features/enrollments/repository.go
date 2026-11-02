package enrollments

import (
	"context"
	"database/sql"

	"github.com/martin/school-api/internal/database"
	"github.com/martin/school-api/internal/shared/apperr"
)

type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Detail, error)
	FindDetail(ctx context.Context, id uint) (Detail, error)
	CountByCourse(ctx context.Context, courseID uint) (int, error)
	Create(ctx context.Context, e *Enrollment) error
	Delete(ctx context.Context, id uint) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

// detailQuery joins in the student and course names. List and FindDetail add
// their own WHERE / ORDER BY after it.
const detailQuery = `
SELECT e.id, e.student_id, s.first_name || ' ' || s.last_name,
       e.course_id, c.code, c.name, e.enrolled_at
FROM enrollments e
JOIN students s ON s.id = e.student_id
JOIN courses  c ON c.id = e.course_id`

func scanDetail(row database.Scanner) (Detail, error) {
	var d Detail
	err := row.Scan(&d.ID, &d.StudentID, &d.StudentName, &d.CourseID, &d.CourseCode, &d.CourseName, &d.EnrolledAt)
	return d, err
}

func (r *sqlRepository) List(ctx context.Context, f ListFilter) ([]Detail, error) {
	query := detailQuery + ` WHERE 1=1`
	var args []any
	if f.StudentID != 0 {
		query += ` AND e.student_id = ?`
		args = append(args, f.StudentID)
	}
	if f.CourseID != 0 {
		query += ` AND e.course_id = ?`
		args = append(args, f.CourseID)
	}
	query += ` ORDER BY e.enrolled_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Detail
	for rows.Next() {
		d, err := scanDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (r *sqlRepository) FindDetail(ctx context.Context, id uint) (Detail, error) {
	d, err := scanDetail(r.db.QueryRowContext(ctx, detailQuery+` WHERE e.id = ?`, id))
	return d, database.MapError(err, "enrollment")
}

func (r *sqlRepository) CountByCourse(ctx context.Context, courseID uint) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM enrollments WHERE course_id = ?`, courseID).Scan(&n)
	return n, err
}

func (r *sqlRepository) Create(ctx context.Context, e *Enrollment) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO enrollments (student_id, course_id, enrolled_at) VALUES (?, ?, ?)`,
		e.StudentID, e.CourseID, e.EnrolledAt)
	// More specific messages than MapError's generic ones. The course was
	// already checked by the service, so a FK failure here means the student.
	switch {
	case database.IsUniqueViolation(err):
		return apperr.Conflict("student is already enrolled in this course")
	case database.IsForeignKeyViolation(err):
		return apperr.Invalid("student does not exist")
	case err != nil:
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	e.ID = uint(id)
	return nil
}

func (r *sqlRepository) Delete(ctx context.Context, id uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM enrollments WHERE id = ?`, id)
	if err != nil {
		return database.MapError(err, "enrollment")
	}
	return database.RequireAffected(res, "enrollment")
}
