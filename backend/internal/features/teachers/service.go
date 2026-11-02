package teachers

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

func (s *Service) List(ctx context.Context, f ListFilter) ([]Teacher, error) {
	f.Search = strings.TrimSpace(f.Search)
	return s.repo.List(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uint) (Teacher, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, in TeacherInput) (Teacher, error) {
	t := Teacher{}
	apply(&t, in)
	if err := s.repo.Create(ctx, &t); err != nil {
		return Teacher{}, err
	}
	return t, nil
}

func (s *Service) Update(ctx context.Context, id uint, in TeacherInput) (Teacher, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Teacher{}, err
	}
	apply(&t, in)
	if err := s.repo.Update(ctx, &t); err != nil {
		return Teacher{}, err
	}
	return t, nil
}

// Delete removes a teacher. Their courses stay, with teacher_id set to NULL
// (see the ON DELETE SET NULL in the courses schema).
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func apply(t *Teacher, in TeacherInput) {
	t.FirstName = strings.TrimSpace(in.FirstName)
	t.LastName = strings.TrimSpace(in.LastName)
	t.Email = strings.ToLower(strings.TrimSpace(in.Email))
	t.Subject = strings.TrimSpace(in.Subject)
}
