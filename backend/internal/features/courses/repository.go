package courses

import (
	"context"
	"time"

	"github.com/StarGazer500/pgrepo"

	"github.com/martin/school-api/internal/database"
	"github.com/martin/school-api/internal/shared/apperr"
)

type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Course, error)
	FindByID(ctx context.Context, id int64) (Course, error)
	Create(ctx context.Context, c *Course) error
	Update(ctx context.Context, c *Course) error
	Delete(ctx context.Context, id int64) error
}

type pgRepository struct {
	rows *pgrepo.Repository[Course]
}

func NewRepository(db pgrepo.DBTX) Repository {
	return &pgRepository{rows: pgrepo.New[Course](db)}
}

func (r *pgRepository) List(ctx context.Context, f ListFilter) ([]Course, error) {
	var where []pgrepo.Filter
	if f.Search != "" {
		like := "%" + f.Search + "%"
		where = append(where, pgrepo.Or(pgrepo.ILike("code", like), pgrepo.ILike("name", like)))
	}
	if f.TeacherID != 0 {
		where = append(where, pgrepo.Eq("teacher_id", f.TeacherID))
	}

	return r.rows.Find(ctx, pgrepo.Query{Where: where, OrderBy: []pgrepo.Order{pgrepo.Asc("code")}})
}

func (r *pgRepository) FindByID(ctx context.Context, id int64) (Course, error) {
	c, err := r.rows.Get(ctx, id)
	if err != nil {
		return Course{}, err
	}
	if c == nil {
		return Course{}, apperr.NotFound("course not found")
	}
	return *c, nil
}

// Create and Update rely on the database's foreign key: an unknown teacherId
// fails with a foreign key violation, which MapError turns into a 400.
// A nil TeacherID is written as NULL.
func (r *pgRepository) Create(ctx context.Context, c *Course) error {
	created, err := r.rows.Create(ctx, map[string]any{
		"code":        c.Code,
		"name":        c.Name,
		"description": c.Description,
		"capacity":    c.Capacity,
		"teacher_id":  c.TeacherID,
	})
	if err != nil {
		return database.MapError(err, "course")
	}
	*c = *created
	return nil
}

func (r *pgRepository) Update(ctx context.Context, c *Course) error {
	updated, err := r.rows.Update(ctx, c.ID, map[string]any{
		"code":        c.Code,
		"name":        c.Name,
		"description": c.Description,
		"capacity":    c.Capacity,
		"teacher_id":  c.TeacherID,
		"updated_at":  time.Now(),
	})
	if err != nil {
		return database.MapError(err, "course")
	}
	if updated == nil {
		return apperr.NotFound("course not found")
	}
	*c = *updated
	return nil
}

func (r *pgRepository) Delete(ctx context.Context, id int64) error {
	deleted, err := r.rows.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return apperr.NotFound("course not found")
	}
	return nil
}
