// Package server is the composition root: the one place that knows about
// every feature and wires repository -> service -> handler -> routes.
package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/martin/school-api/internal/features/courses"
	"github.com/martin/school-api/internal/features/enrollments"
	"github.com/martin/school-api/internal/features/students"
	"github.com/martin/school-api/internal/features/teachers"
)

func New(db *sql.DB) *gin.Engine {
	r := gin.Default() // includes Logger and Recovery middleware

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	studentSvc := students.NewService(students.NewRepository(db))
	students.RegisterRoutes(api, students.NewHandler(studentSvc))

	teacherSvc := teachers.NewService(teachers.NewRepository(db))
	teachers.RegisterRoutes(api, teachers.NewHandler(teacherSvc))

	courseSvc := courses.NewService(courses.NewRepository(db))
	courses.RegisterRoutes(api, courses.NewHandler(courseSvc))

	// enrollments depends on courses through the narrow CourseReader interface.
	enrollmentSvc := enrollments.NewService(enrollments.NewRepository(db), courseSvc)
	enrollments.RegisterRoutes(api, enrollments.NewHandler(enrollmentSvc))

	return r
}
