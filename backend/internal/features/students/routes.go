package students

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts this feature's endpoints under rg (e.g. /api/v1).
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	g := rg.Group("/students")
	g.GET("", h.list)
	g.POST("", h.create)
	g.GET("/:id", h.get)
	g.PUT("/:id", h.update)
	g.DELETE("/:id", h.delete)
}
