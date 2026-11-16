package enrollments

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/martin/school-api/internal/features/courses"
	"github.com/martin/school-api/internal/shared"
)

// CourseReader is the one thing enrollments needs from the courses feature.
// Declaring the interface here, on the consumer side, keeps the dependency
// narrow: courses.Service satisfies it without knowing enrollments exists,
// and tests can pass a tiny fake instead.
type CourseReader interface {
	Get(ctx context.Context, id int64) (courses.Course, error)
}

type Service struct {
	repo    *Repository
	courses CourseReader
	metrics metrics // see telemetry.go
}

func NewService(repo *Repository, courses CourseReader) *Service {
	return &Service{repo: repo, courses: courses, metrics: newMetrics()}
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Detail, error) {
	return s.repo.List(ctx, f)
}

// Enroll adds a student to a course if the course has room.
//
// Note: count-then-insert can race under concurrent requests. For a learning
// project that's fine; in production you'd wrap it in a transaction or enforce
// capacity in the database.
//
// It runs in its own span ("enrollments.Enroll"), so in a trace the course
// lookup, seat count and insert appear grouped under it, with the outcome
// recorded as attributes.
func (s *Service) Enroll(ctx context.Context, in EnrollInput) (Detail, error) {
	ctx, span := tracer.Start(ctx, "enrollments.Enroll", trace.WithAttributes(
		attribute.Int64("student.id", in.StudentID),
		attribute.Int64("course.id", in.CourseID),
	))
	defer span.End()

	course, err := s.courses.Get(ctx, in.CourseID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			s.metrics.reject(ctx, reasonCourseNotFound)
		}
		return Detail{}, err
	}

	taken, err := s.repo.CountByCourse(ctx, course.ID)
	if err != nil {
		return Detail{}, err
	}
	span.SetAttributes(attribute.Int("course.capacity", course.Capacity), attribute.Int("course.taken", taken))
	if taken >= course.Capacity {
		s.metrics.reject(ctx, reasonCourseFull)
		return Detail{}, shared.Conflict("course %s is full (%d/%d)", course.Code, taken, course.Capacity)
	}

	e := Enrollment{StudentID: in.StudentID, CourseID: course.ID, EnrolledAt: time.Now().UTC()}
	if err := s.repo.Create(ctx, &e); err != nil {
		switch {
		case errors.Is(err, shared.ErrConflict):
			s.metrics.reject(ctx, reasonAlreadyEnrolled)
		case errors.Is(err, shared.ErrInvalid):
			s.metrics.reject(ctx, reasonStudentNotFound)
		}
		return Detail{}, err
	}
	s.metrics.created.Add(ctx, 1)
	return s.repo.FindDetail(ctx, e.ID)
}

func (s *Service) Withdraw(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
