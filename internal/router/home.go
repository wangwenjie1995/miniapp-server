package router

import (
	"github.com/gin-gonic/gin"

	"miniapp-server/internal/home"
)

func registerHome(api *gin.RouterGroup, h *home.Handler) {
	group := api.Group("/home")
	group.GET("", h.Overview)
	group.GET("/featured", h.Featured)
	group.GET("/tools", h.Tools)
	group.GET("/banners", h.Banners)
	group.GET("/flash", h.Flash)
	group.GET("/news", h.News)
}
