package courses

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Course, error) {
	f.Search = strings.TrimSpace(f.Search)
	return s.repo.List(ctx, f)
}

func (s *Service) Get(ctx context.Context, id int64) (Course, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, in CourseInput) (Course, error) {
	c := Course{}
	apply(&c, in)
	if err := s.repo.Create(ctx, &c); err != nil {
		return Course{}, err
	}
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
	return c, nil
}

// Delete removes a course and, via ON DELETE CASCADE, its enrollments.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

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
