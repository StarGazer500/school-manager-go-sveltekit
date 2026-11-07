package students

import (
	"context"
	"time"

	"github.com/StarGazer500/pgrepo"

	"github.com/martin/school-api/internal/core"
	"github.com/martin/school-api/internal/shared"
)

// Repository stores students in Postgres with pgrepo, which writes the SQL
// for each call from the Student struct's db tags. The service uses it
// directly, without an interface: add one only when a second implementation
// is needed at the same time (for example a fake for tests).
type Repository struct {
	rows *pgrepo.Repository[Student]
}

func NewRepository(db pgrepo.DBTX) *Repository {
	return &Repository{rows: pgrepo.New[Student](db)}
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]Student, error) {
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

func (r *Repository) FindByID(ctx context.Context, id int64) (Student, error) {
	s, err := r.rows.Get(ctx, id)
	if err != nil {
		return Student{}, err
	}
	if s == nil { // pgrepo returns (nil, nil) when no row has this id
		return Student{}, shared.NotFound("student not found")
	}
	return *s, nil
}

func (r *Repository) Create(ctx context.Context, s *Student) error {
	// The database fills id, created_at and updated_at; Create returns the
	// whole new row (INSERT ... RETURNING *), which we copy back into s.
	created, err := r.rows.Create(ctx, map[string]any{
		"first_name":  s.FirstName,
		"last_name":   s.LastName,
		"email":       s.Email,
		"grade_level": s.GradeLevel,
	})
	if err != nil {
		return core.MapError(err, "student") // UNIQUE(email) -> 409
	}
	*s = *created
	return nil
}

func (r *Repository) Update(ctx context.Context, s *Student) error {
	updated, err := r.rows.Update(ctx, s.ID, map[string]any{
		"first_name":  s.FirstName,
		"last_name":   s.LastName,
		"email":       s.Email,
		"grade_level": s.GradeLevel,
		"updated_at":  time.Now(),
	})
	if err != nil {
		return core.MapError(err, "student")
	}
	if updated == nil {
		return shared.NotFound("student not found")
	}
	*s = *updated
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	deleted, err := r.rows.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return shared.NotFound("student not found")
	}
	return nil
}
