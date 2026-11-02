package enrollments

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	g := rg.Group("/enrollments")
	g.GET("", h.list)
	g.POST("", h.create)
	g.DELETE("/:id", h.delete)
}
