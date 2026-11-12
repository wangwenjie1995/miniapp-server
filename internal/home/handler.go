package home

import (
	"context"
	"log/slog"
	"time"

	"miniapp-server/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 保存本模块需要的 Service 和日志器，不使用全局数据库变量。
type Handler struct {
	service *Service
	logger  *slog.Logger
}

type Dependencies struct {
	Service *Service
	Logger  *slog.Logger
}

// RegisterRoutes 由总路由调用，把首页接口和数据库服务连接起来。
// 六个 GET 接口均为只读；访问控制由传入路由组的中间件决定。
func RegisterRoutes(api *gin.RouterGroup, deps Dependencies) {
	h := &Handler{service: deps.Service, logger: deps.Logger}
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

// reply 统一处理数据库错误：详细原因只记日志，客户端收到纯文本异常提示。
func (h *Handler) reply(c *gin.Context, data any, err error) {
	if err != nil {
		h.logger.Error("首页数据查询失败", "path", c.Request.URL.Path, "error", err)
		response.ServerError(c, "首页数据暂时无法获取")
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
		response.Error(c, response.CodeInvalidParams, "参数错误：tab 仅支持 discover/news/flash/watchlist/journey，page 必须为正整数，size 必须为 1 至 100 的整数")
		return
	}
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.service.News(ctx, query)
	h.reply(c, data, err)
}
