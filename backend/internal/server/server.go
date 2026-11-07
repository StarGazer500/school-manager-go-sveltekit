// Package server is the composition root: the one place that knows about
// every feature and wires repository -> service -> handler -> routes.
package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/martin/school-api/internal/core"
	"github.com/martin/school-api/internal/features/courses"
	"github.com/martin/school-api/internal/features/enrollments"
	"github.com/martin/school-api/internal/features/students"
	"github.com/martin/school-api/internal/features/teachers"
	"github.com/martin/school-api/internal/shared"
)

func New(db *pgxpool.Pool, rdb redis.UniversalClient, log zerolog.Logger) *gin.Engine {
	// gin.New has no middleware; gin.Default would add Gin's own logger and
	// recovery, which these two replace. Order matters: RequestLogger first, so
	// a recovered panic is logged with the request ID and counted as a 500.
	r := gin.New()
	r.Use(shared.RequestLogger(log), shared.Recoverer())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// Each service gets its two dependencies injected: a Postgres repository
	// and a Redis cache for its type. The cache is an interface declared by
	// the feature (so the service doesn't depend on Redis); the repository is
	// the concrete Postgres type.
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
