package students

import (
	"context"
	"time"

	"github.com/StarGazer500/pgrepo"

	"github.com/martin/school-api/internal/database"
	"github.com/martin/school-api/internal/shared/apperr"
)

// Repository is the persistence contract the service depends on. Because it
// is an interface, tests can swap in a fake without touching a database.
type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Student, error)
	FindByID(ctx context.Context, id int64) (Student, error)
	Create(ctx context.Context, s *Student) error
	Update(ctx context.Context, s *Student) error
	Delete(ctx context.Context, id int64) error
}

// pgRepository implements Repository with pgrepo, which writes the SQL for
// each call from the Student struct's db tags.
type pgRepository struct {
	rows *pgrepo.Repository[Student]
}

func NewRepository(db pgrepo.DBTX) Repository {
	return &pgRepository{rows: pgrepo.New[Student](db)}
}

func (r *pgRepository) List(ctx context.Context, f ListFilter) ([]Student, error) {
	// Each filter that is set becomes one condition; pgrepo joins them with AND
	// and passes values as $1, $2... parameters, never inside the SQL text.
	var where []pgrepo.Filter
	if f.Search != "" {
		like := "%" + f.Search + "%"
		where = append(where, pgrepo.Or(
			pgrepo.ILike("first_name", like),
			pgrepo.ILike("last_name", like),
			pgrepo.ILike("email", like),
		))
	}
	if f.GradeLevel != 0 {
		where = append(where, pgrepo.Eq("grade_level", f.GradeLevel))
	}

	return r.rows.Find(ctx, pgrepo.Query{
		Where:   where,
		OrderBy: []pgrepo.Order{pgrepo.Asc("last_name"), pgrepo.Asc("first_name")},
	})
}

func (r *pgRepository) FindByID(ctx context.Context, id int64) (Student, error) {
	s, err := r.rows.Get(ctx, id)
	if err != nil {
		return Student{}, err
	}
	if s == nil { // pgrepo returns (nil, nil) when no row has this id
		return Student{}, apperr.NotFound("student not found")
	}
	return *s, nil
}

func (r *pgRepository) Create(ctx context.Context, s *Student) error {
	// The database fills id, created_at and updated_at; Create returns the
	// whole new row (INSERT ... RETURNING *), which we copy back into s.
	created, err := r.rows.Create(ctx, map[string]any{
		"first_name":  s.FirstName,
		"last_name":   s.LastName,
		"email":       s.Email,
		"grade_level": s.GradeLevel,
	})
	if err != nil {
		return database.MapError(err, "student") // UNIQUE(email) -> 409
	}
	*s = *created
	return nil
}

func (r *pgRepository) Update(ctx context.Context, s *Student) error {
	updated, err := r.rows.Update(ctx, s.ID, map[string]any{
		"first_name":  s.FirstName,
		"last_name":   s.LastName,
		"email":       s.Email,
		"grade_level": s.GradeLevel,
		"updated_at":  time.Now(),
	})
	if err != nil {
		return database.MapError(err, "student")
	}
	if updated == nil {
		return apperr.NotFound("student not found")
	}
	*s = *updated
	return nil
}

func (r *pgRepository) Delete(ctx context.Context, id int64) error {
	deleted, err := r.rows.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return apperr.NotFound("student not found")
	}
	return nil
}
