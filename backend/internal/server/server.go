// Package server is the composition root: the one place that knows about
// every feature and wires repository -> service -> handler -> routes.
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/martin/school-api/internal/core"
	"github.com/martin/school-api/internal/features/courses"
	"github.com/martin/school-api/internal/features/enrollments"
	"github.com/martin/school-api/internal/features/students"
	"github.com/martin/school-api/internal/features/teachers"
)

// New builds the HTTP server. serviceName names the service in traces.
func New(db *pgxpool.Pool, rdb redis.UniversalClient, log zerolog.Logger, serviceName string) *gin.Engine {
	// gin.New has no middleware; gin.Default would add Gin's own logger and
	// recovery, which ours replace. Order matters:
	//   1. tracing starts the request's span, so everything after is inside it
	//   2. requestLogger can then put the trace ID in the logs and response
	//   3. recoverer turns a panic into a 500 that the two above still see
	// (all three are in middleware.go)
	r := gin.New()
	r.Use(tracing(serviceName), requestLogger(log), recoverer())

	r.GET("/healthz", liveness)
	r.GET("/readyz", readiness(db, rdb))

	api := r.Group("/api/v1")

	// Each service gets its two dependencies injected: a Postgres repository
	// and a Redis cache for its type. Both are concrete types; the only
	// interface here is enrollments' CourseReader, which limits enrollments
	// to the one courses method it needs.
	studentSvc := students.NewService(
		students.NewRepository(db),
		core.NewCache[students.Student](rdb, "student:", log),
	)
	students.RegisterRoutes(api, students.NewHandler(studentSvc))

	teacherSvc := teachers.NewService(
		teachers.NewRepository(db),
		core.NewCache[teachers.Teacher](rdb, "teacher:", log),
	)
	teachers.RegisterRoutes(api, teachers.NewHandler(teacherSvc))

	courseSvc := courses.NewService(
		courses.NewRepository(db),
		core.NewCache[courses.Course](rdb, "course:", log),
	)
	courses.RegisterRoutes(api, courses.NewHandler(courseSvc))

	// enrollments depends on courses through the narrow CourseReader interface.
	enrollmentSvc := enrollments.NewService(enrollments.NewRepository(db), courseSvc)
	enrollments.RegisterRoutes(api, enrollments.NewHandler(enrollmentSvc))

	return r
}
