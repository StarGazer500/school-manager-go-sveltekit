package students

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/martin/school-api/internal/shared"
)

// Handler translates HTTP requests into service calls and service results
// into JSON. It holds no business logic.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /students?q=ann&grade=10
func (h *Handler) list(c *gin.Context) {
	f := ListFilter{Search: c.Query("q")}
	if g := c.Query("grade"); g != "" {
		grade, err := strconv.Atoi(g)
		if err != nil {
			status, msg := shared.ErrorResponse(c, shared.Invalid("grade must be a number"))
			c.AbortWithStatusJSON(status, gin.H{"error": msg})
			return
		}
		f.GradeLevel = grade
	}

	list, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponses(list))
}

// GET /students/:id
func (h *Handler) get(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	st, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponse(st))
}

// POST /students
func (h *Handler) create(c *gin.Context) {
	var in StudentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
		return
	}
	st, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, toResponse(st))
}

// PUT /students/:id
func (h *Handler) update(c *gin.Context) {
	id, err := shared.IDParam(c, "id")
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	var in StudentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": shared.ValidationMessage(err)})
		return
	}
	st, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		status, msg := shared.ErrorResponse(c, err)
		c.AbortWithStatusJSON(status, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, toResponse(st))
}

// DELETE /students/:id
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
