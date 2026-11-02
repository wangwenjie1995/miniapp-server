package home

import (
	"context"
	"log/slog"
	"miniapp-server/internal/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func RegisterRoutes(api *gin.RouterGroup, service * Service, logger *slog.Logger) {
	h := &Handler{service: service, logger: logger}
	group := api.Group("/home")
	group.GET("", h.overview)
	group.GET("/featured", h.featured)
	group.GET("/tools", h.tools)
	group.GET("/banners", h.banners)
	group.GET("/flash", h.flash)
	group.GET("/news", h.news)
}

func queryContext(c *gin.Context) (context.Context, context.CancelFunc) {
	// context.WithTimeout 会在 5 秒后自动取消，或者客户端断开连接时取消。
	return context.WithTimeout(c.Request.Context(), 5*time.Second)
}

func (h *Handler) reply(c *gin.Context, data any, err error) {
	if err != nil {
		h.logger.Error("首页数据查询失败", "path", c.Request.URL.Path, "error", err)
		response.JSON(c, http.StatusInternalServerError, gin.H{"code": 500, "message": "首页数据暂时无法获取", "data": nil})
		return
	}
	response.Success(c, data)
}

func (h *Handler) overview(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.service.Overview(ctx)
	h.reply(c, data, err)
}