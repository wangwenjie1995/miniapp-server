package user

import (
	"errors"
	"log/slog"
	"miniapp-server/internal/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
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
		response.Error(c, http.StatusBadRequest, "INVALID_PARAMS", registerValidationMessage(err))
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrUsernameExists) {
			response.Error(c, http.StatusConflict, "USERNAME_EXISTS", "用户名已存在，请更换用户名")
			return
		}
		h.logger.Error("用户注册失败", "error", err)
		response.Error(c, http.StatusInternalServerError, "REGISTER_FAILED", "注册失败")
		return
	}

	response.Success(c, UserResponse{
		ID:       user.ID,
		Username: user.Username,
	})
}

// registerValidationMessage 将校验错误转成用户可以理解的提示。
func registerValidationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	// errors.As(err, &validationErrors) 会检查 err 是否可以被转换为 validator.ValidationErrors 类型，
	// 如果可以，则将其赋值给 validationErrors 变量。
	// 如果不可以，则返回 false，表示 err 不是一个验证错误。
	if !errors.As(err, &validationErrors) {
		return "请求格式错误，请提交包含 username 和 password 的 JSON 对象"
	}
	var messages []string
	for _, field := range validationErrors {
		switch field.Field() {
		case "Username":
			if field.Tag() == "required" {
				messages = append(messages, "请输入用户名")
			} else {
				messages = append(messages, "用户名长度须为 4 到 24 个字符")
			}
		case "Password":
			if field.Tag() == "required" {
				messages = append(messages, "请输入密码")
			} else {
				messages = append(messages, "密码长度须为 8 到 72 个字符")
			}
		}
	}
	if len(messages) == 0 {
		return "注册参数不合法"
	}
	return strings.Join(messages, "；")
}
