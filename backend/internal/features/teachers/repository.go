package teachers

import (
	"context"
	"time"

	"github.com/StarGazer500/pgrepo"

	"github.com/martin/school-api/internal/database"
	"github.com/martin/school-api/internal/shared/apperr"
)

type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Teacher, error)
	FindByID(ctx context.Context, id int64) (Teacher, error)
	Create(ctx context.Context, t *Teacher) error
	Update(ctx context.Context, t *Teacher) error
	Delete(ctx context.Context, id int64) error
}

type pgRepository struct {
	rows *pgrepo.Repository[Teacher]
}

func NewRepository(db pgrepo.DBTX) Repository {
	return &pgRepository{rows: pgrepo.New[Teacher](db)}
}

func (r *pgRepository) List(ctx context.Context, f ListFilter) ([]Teacher, error) {
	var where []pgrepo.Filter
	if f.Search != "" {
		like := "%" + f.Search + "%"
		where = append(where, pgrepo.Or(
			pgrepo.ILike("first_name", like),
			pgrepo.ILike("last_name", like),
			pgrepo.ILike("subject", like),
		))
	}

	return r.rows.Find(ctx, pgrepo.Query{
		Where:   where,
		OrderBy: []pgrepo.Order{pgrepo.Asc("last_name"), pgrepo.Asc("first_name")},
	})
}

func (r *pgRepository) FindByID(ctx context.Context, id int64) (Teacher, error) {
	t, err := r.rows.Get(ctx, id)
	if err != nil {
		return Teacher{}, err
	}
	if t == nil {
		return Teacher{}, apperr.NotFound("teacher not found")
	}
	return *t, nil
}

func (r *pgRepository) Create(ctx context.Context, t *Teacher) error {
	created, err := r.rows.Create(ctx, map[string]any{
		"first_name": t.FirstName,
		"last_name":  t.LastName,
		"email":      t.Email,
		"subject":    t.Subject,
	})
	if err != nil {
		return database.MapError(err, "teacher")
	}
	*t = *created
	return nil
}

func (r *pgRepository) Update(ctx context.Context, t *Teacher) error {
	updated, err := r.rows.Update(ctx, t.ID, map[string]any{
		"first_name": t.FirstName,
		"last_name":  t.LastName,
		"email":      t.Email,
		"subject":    t.Subject,
		"updated_at": time.Now(),
	})
	if err != nil {
		return database.MapError(err, "teacher")
	}
	if updated == nil {
		return apperr.NotFound("teacher not found")
	}
	*t = *updated
	return nil
}

func (r *pgRepository) Delete(ctx context.Context, id int64) error {
	deleted, err := r.rows.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return apperr.NotFound("teacher not found")
	}
	return nil
}
