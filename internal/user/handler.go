package user

import (
	"log/slog"
	"miniapp-server/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	logger *slog.Logger
}

func RegisterRoutes(api *gin.RouterGroup, service *Service, logger *slog.Logger) {
	h := &Handler{service: service, logger: logger}
	group := api.Group("/user")
	group.POST("/register", h.register)
}


// 为啥是gin.Context，而不是context.Context？
// gin.Context 是 Gin 框架提供的上下文对象，专门用于处理 HTTP 请求和响应。它封装了请求的详细信息（如请求头、请求体、URL 参数等）以及响应的相关方法（如设置状态码、写入响应体等）。
// context.Context 是 Go 标准库提供的上下文对象，主要用于在不同函数之间传递取消信号、截止时间和其他请求范围的数据。它不包含 HTTP 请求和响应的具体信息。
// 在处理 HTTP 请求时，使用 gin.Context 可以方便地访问请求数据和发送响应，而 context.Context 更适合用于控制函数执行的生命周期和传递跨 API 的数据。因此，在 Gin 的处理函数中，通常会使用 gin.Context 来处理 HTTP 请求。
func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PARAMS", "注册参数错误")
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Error("用户注册失败", "error", err)
		response.Error(c, http.StatusInternalServerError, "REGISTER_FAILED", "注册失败")
		return
	}

	response.Success(c, UserResponse{
		ID: user.ID,
		Username: user.Username,
	})
}