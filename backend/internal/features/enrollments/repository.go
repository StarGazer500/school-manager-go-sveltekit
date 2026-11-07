package enrollments

import (
	"context"

	"github.com/StarGazer500/pgrepo"
	"golang.org/x/sync/errgroup"

	"github.com/martin/school-api/internal/database"
	"github.com/martin/school-api/internal/features/courses"
	"github.com/martin/school-api/internal/features/students"
	"github.com/martin/school-api/internal/shared/apperr"
)

type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Detail, error)
	FindDetail(ctx context.Context, id int64) (Detail, error)
	CountByCourse(ctx context.Context, courseID int64) (int, error)
	Create(ctx context.Context, e *Enrollment) error
	Delete(ctx context.Context, id int64) error
}

// pgRepository reads students and courses too, to put their names on each
// Detail. It only reads them; writing them is their own features' job.
type pgRepository struct {
	enrollments *pgrepo.Repository[Enrollment]
	students    *pgrepo.Repository[students.Student]
	courses     *pgrepo.Repository[courses.Course]
}

func NewRepository(db pgrepo.DBTX) Repository {
	return &pgRepository{
		enrollments: pgrepo.New[Enrollment](db),
		students:    pgrepo.New[students.Student](db),
		courses:     pgrepo.New[courses.Course](db),
	}
}

func (r *pgRepository) List(ctx context.Context, f ListFilter) ([]Detail, error) {
	var where []pgrepo.Filter
	if f.StudentID != 0 {
		where = append(where, pgrepo.Eq("student_id", f.StudentID))
	}
	if f.CourseID != 0 {
		where = append(where, pgrepo.Eq("course_id", f.CourseID))
	}

	list, err := r.enrollments.Find(ctx, pgrepo.Query{
		Where:   where,
		OrderBy: []pgrepo.Order{pgrepo.Desc("enrolled_at"), pgrepo.Desc("id")},
	})
	if err != nil {
		return nil, err
	}
	return r.withNames(ctx, list)
}

// withNames turns enrollments into Details. Instead of one query per row, it
// loads all the students in one query and all the courses in another
// (pgrepo.ManyToOne: SELECT * FROM students WHERE id IN (...)).
func (r *pgRepository) withNames(ctx context.Context, list []Enrollment) ([]Detail, error) {
	studentIDs := make([]int64, len(list))
	courseIDs := make([]int64, len(list))
	for i, e := range list {
		studentIDs[i], courseIDs[i] = e.StudentID, e.CourseID
	}

	// The two lookups don't depend on each other, so run them at the same time.
	// Each goroutine writes only its own variable, and Wait returns after both
	// finish, so reading them afterwards is safe. If one fails, errgroup
	// cancels gctx, which stops the other query early, and Wait returns that
	// first error. This needs a pool (two connections at once); on a single
	// connection, like a WithTx repository, the second query would fail.
	var (
		studentsByID map[int64]students.Student
		coursesByID  map[int64]courses.Course
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		studentsByID, err = pgrepo.ManyToOne(gctx, r.students, studentIDs)
		return err
	})
	g.Go(func() error {
		var err error
		coursesByID, err = pgrepo.ManyToOne(gctx, r.courses, courseIDs)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	out := make([]Detail, len(list))
	for i, e := range list {
		s, c := studentsByID[e.StudentID], coursesByID[e.CourseID]
		out[i] = Detail{
			ID:          e.ID,
			StudentID:   e.StudentID,
			StudentName: s.FirstName + " " + s.LastName,
			CourseID:    e.CourseID,
			CourseCode:  c.Code,
			CourseName:  c.Name,
			EnrolledAt:  e.EnrolledAt,
		}
	}
	return out, nil
}

func (r *pgRepository) FindDetail(ctx context.Context, id int64) (Detail, error) {
	e, err := r.enrollments.Get(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	if e == nil {
		return Detail{}, apperr.NotFound("enrollment not found")
	}
	details, err := r.withNames(ctx, []Enrollment{*e})
	if err != nil {
		return Detail{}, err
	}
	return details[0], nil
}

func (r *pgRepository) CountByCourse(ctx context.Context, courseID int64) (int, error) {
	n, err := r.enrollments.Count(ctx, pgrepo.Eq("course_id", courseID))
	return int(n), err
}

func (r *pgRepository) Create(ctx context.Context, e *Enrollment) error {
	created, err := r.enrollments.Create(ctx, map[string]any{
		"student_id":  e.StudentID,
		"course_id":   e.CourseID,
		"enrolled_at": e.EnrolledAt,
	})
	// More specific messages than MapError's generic ones. The course was
	// already checked by the service, so a FK failure here means the student.
	switch {
	case database.IsUniqueViolation(err):
		return apperr.Conflict("student is already enrolled in this course")
	case database.IsForeignKeyViolation(err):
		return apperr.Invalid("student does not exist")
	case err != nil:
		return err
	}
	*e = *created
	return nil
}

func (r *pgRepository) Delete(ctx context.Context, id int64) error {
	deleted, err := r.enrollments.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return apperr.NotFound("enrollment not found")
	}
	return nil
}
