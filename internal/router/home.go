package router

import (
	"github.com/gin-gonic/gin"

	"miniapp-server/internal/home"
	"miniapp-server/internal/platform"
)

func initHomeRoutes(publicGroup *gin.RouterGroup, resources platform.Resources) {
	handler := home.NewHandler(home.NewRepository(resources.DB), resources.Log())
	RegisterHomeRoutes(publicGroup, handler)
}

func RegisterHomeRoutes(publicGroup *gin.RouterGroup, handler *home.Handler) {
	group := publicGroup.Group("/home")
	group.GET("", handler.Overview)
	group.GET("/featured", handler.Featured)
	group.GET("/tools", handler.Tools)
	group.GET("/banners", handler.Banners)
	group.GET("/flash", handler.Flash)
	group.GET("/news", handler.News)
}
