package enrollments

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/martin/school-api/internal/shared"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /enrollments?studentId=1&courseId=2
func (h *Handler) list(c *gin.Context) {
	studentID, err := shared.OptionalIDQuery(c, "studentId")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	courseID, err := shared.OptionalIDQuery(c, "courseId")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}

	list, err := h.svc.List(c.Request.Context(), ListFilter{StudentID: studentID, CourseID: courseID})
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponses(list))
}

// POST /enrollments
func (h *Handler) create(c *gin.Context) {
	var in EnrollInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
		return
	}
	d, err := h.svc.Enroll(c.Request.Context(), in)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, toResponse(d))
}

// DELETE /enrollments/:id
func (h *Handler) delete(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	if err := h.svc.Withdraw(c.Request.Context(), id); err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.Status(http.StatusNoContent)
}
