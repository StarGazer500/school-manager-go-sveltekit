package teachers

import (
	"context"
	"strconv"
	"strings"

	"github.com/StarGazer500/rediscache"

	"github.com/martin/school-api/internal/core"
	"github.com/martin/school-api/internal/features/courses"
)

type Service struct {
	repo  *Repository
	cache *rediscache.Cache[Teacher]
}

func NewService(repo *Repository, c *rediscache.Cache[Teacher]) *Service {
	return &Service{repo: repo, cache: c}
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Teacher, error) {
	f.Search = strings.TrimSpace(f.Search)
	return s.repo.List(ctx, f)
}

// Get returns the cached teacher, or loads and caches it.
func (s *Service) Get(ctx context.Context, id int64) (Teacher, error) {
	t, err := s.cache.GetOrLoad(ctx, key(id), func(ctx context.Context) (Teacher, error) {
		t, err := s.repo.FindByID(ctx, id)
		return t, core.ToCache(err)
	})
	return t, core.FromCache(err, "teacher not found")
}

func (s *Service) Create(ctx context.Context, in TeacherInput) (Teacher, error) {
	t := Teacher{}
	apply(&t, in)
	if err := s.repo.Create(ctx, &t); err != nil {
		return Teacher{}, err
	}
	s.forget(ctx, t.ID)
	return t, nil
}

func (s *Service) Update(ctx context.Context, id int64, in TeacherInput) (Teacher, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Teacher{}, err
	}
	apply(&t, in)
	if err := s.repo.Update(ctx, &t); err != nil {
		return Teacher{}, err
	}
	s.forget(ctx, id)
	return t, nil
}

// Delete removes a teacher. Their courses stay, with teacher_id set to NULL
// (ON DELETE SET NULL in the courses migration). The courses cache can't see
// that, so this drops every cached course through their shared tag; tags
// work across caches, so this cache can do it.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.forget(ctx, id)
	_, err := s.cache.Invalidate(ctx, courses.CacheTag)
	core.WarnInvalidate(ctx, err, "tag "+courses.CacheTag)
	return nil
}

func (s *Service) forget(ctx context.Context, id int64) {
	_, err := s.cache.Delete(ctx, key(id))
	core.WarnInvalidate(ctx, err, "teacher:"+key(id))
}

func key(id int64) string { return strconv.FormatInt(id, 10) }

func apply(t *Teacher, in TeacherInput) {
	t.FirstName = strings.TrimSpace(in.FirstName)
	t.LastName = strings.TrimSpace(in.LastName)
	t.Email = strings.ToLower(strings.TrimSpace(in.Email))
	t.Subject = strings.TrimSpace(in.Subject)
}
