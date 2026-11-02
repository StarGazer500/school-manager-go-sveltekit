package courses

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

// GET /courses?q=math&teacherId=2
func (h *Handler) list(c *gin.Context) {
	teacherID, err := httpx.OptionalIDQuery(c, "teacherId")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	list, err := h.svc.List(c.Request.Context(), ListFilter{Search: c.Query("q"), TeacherID: teacherID})
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponses(list))
}

// GET /courses/:id
func (h *Handler) get(c *gin.Context) {
	id, err := httpx.IDParam(c, "id")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	course, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(course))
}

// POST /courses
func (h *Handler) create(c *gin.Context) {
	var in CourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	course, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, toResponse(course))
}

// PUT /courses/:id
func (h *Handler) update(c *gin.Context) {
	id, err := httpx.IDParam(c, "id")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	var in CourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	course, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(course))
}

// DELETE /courses/:id
func (h *Handler) delete(c *gin.Context) {
	id, err := httpx.IDParam(c, "id")
	if err != nil {
		httpx.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
