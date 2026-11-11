// response 包统一业务 JSON 和服务异常的纯文本响应。
package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 业务码独立于 HTTP 状态码，前端通过 code 判断业务结果。
const (
	CodeSuccess            = 0
	CodeInvalidParams      = 4000
	CodeInvalidCredentials = 5000
	CodeUsernameExists     = 5001
)

func JSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

// Error 返回 HTTP 200 和业务错误。调用方仍需 return 结束当前函数。
func Error(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": code, "message": message, "data": nil})
}

// ServerError 返回 HTTP 500 和纯文本，具体错误只写入服务端日志。
func ServerError(c *gin.Context, message string) {
	c.Abort()
	c.String(http.StatusInternalServerError, "%s", message)
}

func Success(c *gin.Context, data any) {
	JSON(c, http.StatusOK, gin.H{"code": CodeSuccess, "message": "success", "data": data})
}
