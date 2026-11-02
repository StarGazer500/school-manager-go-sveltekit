package enrollments

import (
	"context"
	"time"

	"github.com/martin/school-api/internal/features/courses"
	"github.com/martin/school-api/internal/shared/apperr"
)

// CourseReader is the one thing enrollments needs from the courses feature.
// Declaring the interface here, on the consumer side, keeps the dependency
// narrow: courses.Service satisfies it without knowing enrollments exists,
// and tests can pass a tiny fake instead.
type CourseReader interface {
	Get(ctx context.Context, id uint) (courses.Course, error)
}

type Service struct {
	repo    Repository
	courses CourseReader
}

func NewService(repo Repository, courses CourseReader) *Service {
	return &Service{repo: repo, courses: courses}
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Detail, error) {
	return s.repo.List(ctx, f)
}

// Enroll adds a student to a course if the course has room.
//
// Note: count-then-insert can race under concurrent requests. For a learning
// project that's fine; in production you'd wrap it in a transaction or enforce
// capacity in the database.
func (s *Service) Enroll(ctx context.Context, in EnrollInput) (Detail, error) {
	course, err := s.courses.Get(ctx, in.CourseID)
	if err != nil {
		return Detail{}, err
	}

	taken, err := s.repo.CountByCourse(ctx, course.ID)
	if err != nil {
		return Detail{}, err
	}
	if taken >= course.Capacity {
		return Detail{}, apperr.Conflict("course %s is full (%d/%d)", course.Code, taken, course.Capacity)
	}

	e := Enrollment{StudentID: in.StudentID, CourseID: course.ID, EnrolledAt: time.Now().UTC()}
	if err := s.repo.Create(ctx, &e); err != nil {
		return Detail{}, err
	}
	return s.repo.FindDetail(ctx, e.ID)
}

func (s *Service) Withdraw(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
