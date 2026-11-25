package home

import (
	"context"
	"log/slog"
	"time"

	"miniapp-server/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理首页请求。
type Handler struct {
	repo   Repository
	logger *slog.Logger
}

func NewHandler(repo Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
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

func (h *Handler) Overview(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := loadOverview(ctx, h.repo)
	h.reply(c, data, err)
}
func (h *Handler) Featured(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.repo.Featured(ctx)
	h.reply(c, data, err)
}
func (h *Handler) Tools(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.repo.Tools(ctx)
	h.reply(c, data, err)
}
func (h *Handler) Banners(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.repo.Banners(ctx)
	h.reply(c, data, err)
}
func (h *Handler) Flash(c *gin.Context) {
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.repo.Flash(ctx)
	h.reply(c, data, err)
}

// News 校验查询参数后访问数据库。
func (h *Handler) News(c *gin.Context) {
	var query NewsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, response.CodeInvalidParams, "参数错误：tab 仅支持 discover/news/flash/watchlist/journey，page 必须为正整数，size 必须为 1 至 100 的整数")
		return
	}
	ctx, cancel := queryContext(c)
	defer cancel()
	data, err := h.repo.News(ctx, query)
	h.reply(c, data, err)
}
