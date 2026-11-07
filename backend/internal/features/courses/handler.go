package courses

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

// GET /courses?q=math&teacherId=2
func (h *Handler) list(c *gin.Context) {
	teacherID, err := shared.OptionalIDQuery(c, "teacherId")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	list, err := h.svc.List(c.Request.Context(), ListFilter{Search: c.Query("q"), TeacherID: teacherID})
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponses(list))
}

// GET /courses/:id
func (h *Handler) get(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	course, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponse(course))
}

// POST /courses
func (h *Handler) create(c *gin.Context) {
	var in CourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
		return
	}
	course, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, toResponse(course))
}

// PUT /courses/:id
func (h *Handler) update(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	var in CourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
		return
	}
	course, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponse(course))
}

// DELETE /courses/:id
func (h *Handler) delete(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.Status(http.StatusNoContent)
}
