package students

import (
	"context"
	"strings"
)

// Service holds the business rules. It knows nothing about HTTP or SQL:
// it takes plain Go values and talks to storage through Repository.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Student, error) {
	f.Search = strings.TrimSpace(f.Search)
	return s.repo.List(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uint) (Student, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, in StudentInput) (Student, error) {
	st := Student{}
	apply(&st, in)
	if err := s.repo.Create(ctx, &st); err != nil {
		return Student{}, err
	}
	return st, nil
}

func (s *Service) Update(ctx context.Context, id uint, in StudentInput) (Student, error) {
	st, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Student{}, err
	}
	apply(&st, in)
	if err := s.repo.Update(ctx, &st); err != nil {
		return Student{}, err
	}
	return st, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// apply copies normalized input onto the entity.
func apply(st *Student, in StudentInput) {
	st.FirstName = strings.TrimSpace(in.FirstName)
	st.LastName = strings.TrimSpace(in.LastName)
	st.Email = strings.ToLower(strings.TrimSpace(in.Email))
	st.GradeLevel = in.GradeLevel
}
