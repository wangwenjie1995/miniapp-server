package router

import (
	"github.com/gin-gonic/gin"

	"miniapp-server/internal/global"
	"miniapp-server/internal/home"
)

func initHomeRoutes(publicGroup *gin.RouterGroup) {
	handler := home.NewHandler(home.NewRepository(global.DB), global.Logger)
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
