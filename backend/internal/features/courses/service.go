package courses

import (
	"context"
	"strconv"
	"strings"

	"github.com/StarGazer500/rediscache"

	"github.com/martin/school-api/internal/core"
)

// CacheTag is attached to every cached course. Deleting a teacher sets
// teacher_id to NULL on their courses in the database (ON DELETE SET NULL),
// which this service can't see, so the teachers service invalidates this tag
// to drop every cached course at once.
const CacheTag = "courses"

type Service struct {
	repo  *Repository
	cache *rediscache.Cache[Course]
}

func NewService(repo *Repository, c *rediscache.Cache[Course]) *Service {
	return &Service{repo: repo, cache: c}
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Course, error) {
	f.Search = strings.TrimSpace(f.Search)
	return s.repo.List(ctx, f)
}

// Get returns the cached course, or loads and caches it with CacheTag.
// The enrollments service calls this too, for the capacity check.
func (s *Service) Get(ctx context.Context, id int64) (Course, error) {
	c, err := s.cache.GetOrLoad(ctx, key(id), func(ctx context.Context) (Course, error) {
		c, err := s.repo.FindByID(ctx, id)
		return c, core.ToCache(err)
	}, rediscache.Tags(CacheTag))
	return c, core.FromCache(err, "course not found")
}

func (s *Service) Create(ctx context.Context, in CourseInput) (Course, error) {
	c := Course{}
	apply(&c, in)
	if err := s.repo.Create(ctx, &c); err != nil {
		return Course{}, err
	}
	s.forget(ctx, c.ID)
	return c, nil
}

func (s *Service) Update(ctx context.Context, id int64, in CourseInput) (Course, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Course{}, err
	}
	apply(&c, in)
	if err := s.repo.Update(ctx, &c); err != nil {
		return Course{}, err
	}
	s.forget(ctx, id)
	return c, nil
}

// Delete removes a course and, via ON DELETE CASCADE, its enrollments.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.forget(ctx, id)
	return nil
}

func (s *Service) forget(ctx context.Context, id int64) {
	_, err := s.cache.Delete(ctx, key(id))
	core.WarnInvalidate(ctx, err, "course:"+key(id))
}

func key(id int64) string { return strconv.FormatInt(id, 10) }

func apply(c *Course, in CourseInput) {
	c.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	c.Name = strings.TrimSpace(in.Name)
	c.Description = strings.TrimSpace(in.Description)
	c.Capacity = in.Capacity
	c.TeacherID = in.TeacherID
	if c.TeacherID != nil && *c.TeacherID == 0 {
		c.TeacherID = nil
	}
}
