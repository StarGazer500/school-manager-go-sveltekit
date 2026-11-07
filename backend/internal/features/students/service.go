package students

import (
	"context"
	"strconv"
	"strings"

	"github.com/StarGazer500/rediscache"

	"github.com/martin/school-api/internal/core"
)

// Cache is what the service needs from a cache of students. server.go
// injects a *rediscache.Cache[Student] (Redis); a test can pass a fake that
// just calls load. Declaring the interface here, where it's used, keeps the
// service to the two methods it actually needs.
type Cache interface {
	GetOrLoad(ctx context.Context, key string, load func(context.Context) (Student, error), opts ...rediscache.SetOption) (Student, error)
	Delete(ctx context.Context, keys ...string) (int64, error)
}

// Service holds the business rules. It knows nothing about HTTP or SQL:
// it takes plain Go values, talks to storage through Repository, and keeps
// single students in Cache.
type Service struct {
	repo  *Repository
	cache Cache
}

func NewService(repo *Repository, c Cache) *Service {
	return &Service{repo: repo, cache: c}
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Student, error) {
	f.Search = strings.TrimSpace(f.Search)
	return s.repo.List(ctx, f)
}

// Get returns the cached student, or loads it from the database and caches
// it. Concurrent misses for the same id share one query, and unknown ids are
// remembered as "not found" for a minute (see internal/core/cache.go).
func (s *Service) Get(ctx context.Context, id int64) (Student, error) {
	st, err := s.cache.GetOrLoad(ctx, key(id), func(ctx context.Context) (Student, error) {
		st, err := s.repo.FindByID(ctx, id)
		return st, core.ToCache(err)
	})
	return st, core.FromCache(err, "student not found")
}

// The writes go to the database first, then remove the cached entry, so the
// next Get loads the new data. Create removes it too, in case the id was
// cached as "not found" before the student existed.

func (s *Service) Create(ctx context.Context, in StudentInput) (Student, error) {
	st := Student{}
	apply(&st, in)
	if err := s.repo.Create(ctx, &st); err != nil {
		return Student{}, err
	}
	s.forget(ctx, st.ID)
	return st, nil
}

func (s *Service) Update(ctx context.Context, id int64, in StudentInput) (Student, error) {
	// Read from the database, not the cache, so the update starts from the
	// current row.
	st, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Student{}, err
	}
	apply(&st, in)
	if err := s.repo.Update(ctx, &st); err != nil {
		return Student{}, err
	}
	s.forget(ctx, id)
	return st, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.forget(ctx, id)
	return nil
}

// forget removes a student from the cache. If Redis is down this only logs:
// the database write already succeeded, so the request shouldn't fail.
func (s *Service) forget(ctx context.Context, id int64) {
	_, err := s.cache.Delete(ctx, key(id))
	core.WarnInvalidate(ctx, err, "student:"+key(id))
}

func key(id int64) string { return strconv.FormatInt(id, 10) }

// apply copies normalized input onto the entity.
func apply(st *Student, in StudentInput) {
	st.FirstName = strings.TrimSpace(in.FirstName)
	st.LastName = strings.TrimSpace(in.LastName)
	st.Email = strings.ToLower(strings.TrimSpace(in.Email))
	st.GradeLevel = in.GradeLevel
}
