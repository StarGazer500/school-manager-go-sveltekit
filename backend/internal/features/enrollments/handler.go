package enrollments

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/martin/school-api/internal/shared/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /enrollments?studentId=1&courseId=2
func (h *Handler) list(c *gin.Context) {
	studentID, err := httpx.OptionalIDQuery(c, "studentId")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	courseID, err := httpx.OptionalIDQuery(c, "courseId")
	if err != nil {
		httpx.Error(c, err)
		return
	}

	list, err := h.svc.List(c.Request.Context(), ListFilter{StudentID: studentID, CourseID: courseID})
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponses(list))
}

// POST /enrollments
func (h *Handler) create(c *gin.Context) {
	var in EnrollInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	d, err := h.svc.Enroll(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, toResponse(d))
}

// DELETE /enrollments/:id
func (h *Handler) delete(c *gin.Context) {
	id, err := httpx.IDParam(c, "id")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.Withdraw(c.Request.Context(), id); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
