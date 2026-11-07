package teachers

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

// GET /teachers?q=math
func (h *Handler) list(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), ListFilter{Search: c.Query("q")})
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponses(list))
}

// GET /teachers/:id
func (h *Handler) get(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	t, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponse(t))
}

// POST /teachers
func (h *Handler) create(c *gin.Context) {
	var in TeacherInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
		return
	}
	t, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, toResponse(t))
}

// PUT /teachers/:id
func (h *Handler) update(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	var in TeacherInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
		return
	}
	t, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponse(t))
}

// DELETE /teachers/:id
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
