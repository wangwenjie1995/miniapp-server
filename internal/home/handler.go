package home

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"miniapp-server/internal/response"
)

// Handler 保存本模块需要的 Service 和日志器，不使用全局数据库变量。
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// RegisterRoutes 由总路由调用，把首页接口和数据库服务连接起来。
// 现有六个 GET 接口保持公开，只读，不新增用户、登录或 AI 功能。
func RegisterRoutes(api *gin.RouterGroup, service *Service, logger *slog.Logger) {
	h := &Handler{service: service, logger: logger}
	group := api.Group("/home")
	group.GET("", h.overview)
	group.GET("/featured", h.featured)
	group.GET("/tools", h.tools)
	group.GET("/banners", h.banners)
	group.GET("/flash", h.flash)
	group.GET("/news", h.news)
}

// queryContext 从当前 HTTP 请求派生超时；客户端断开或 5 秒到期时取消查询。
func queryContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 5*time.Second)
}

// reply 统一处理数据库错误：详细原因只记日志，客户端收到稳定的业务错误结构。
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
func (h *Handler) featured(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.service.Featured(ctx)
	h.reply(c, data, err)
}
func (h *Handler) tools(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.service.Tools(ctx)
	h.reply(c, data, err)
}
func (h *Handler) banners(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.service.Banners(ctx)
	h.reply(c, data, err)
}
func (h *Handler) flash(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.service.Flash(ctx)
	h.reply(c, data, err)
}

// news 保留原参数绑定规则，通过校验后才访问数据库。
func (h *Handler) news(c *gin.Context) {
	var query NewsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：tab 仅支持 discover/news/flash/watchlist/journey，page 必须为正整数，size 必须为 1 至 100 的整数",
			"data":    nil,
		})
		return
	}
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.service.News(ctx, query)
	h.reply(c, data, err)
}
