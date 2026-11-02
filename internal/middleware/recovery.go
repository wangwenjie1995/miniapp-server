// 本文件处理请求执行过程中的 panic，避免异常越过当前处理链。
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"miniapp-server/internal/response"
)

// Recovery 创建异常恢复中间件。普通业务失败用 error 或响应表达；
// panic 表示执行被异常中断，recover 用来在同一个 goroutine 内接住这种异常。
// 这里不能捕获处理函数另行启动的其他 goroutine 中发生的 panic。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先注册 defer：无论下面的处理链正常返回还是发生 panic，都会执行此收尾函数。
		defer func() {
			// recover 必须在延迟执行的函数中调用，返回非 nil 表示接住了 panic。
			if recovered := recover(); recovered != nil {
				// 记录异常值和调用堆栈，用于定位出错位置；堆栈只写日志，不返回客户端。
				logger.Error("HTTP 请求异常",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
				// 如果响应头已经发送，就无法改成 500，也不应拼接另一份 JSON。
				if c.Writer.Written() {
					// Abort 阻止后续处理函数执行；下一行 return 才是退出当前延迟函数。
					c.Abort()
					return
				}
				// 响应尚未写出时，返回统一的 500 错误，不把内部异常细节暴露给客户端。
				response.Error(c, http.StatusInternalServerError,
					"INTERNAL_ERROR", "服务内部错误")
			}
		}()
		// 执行后续处理链。发生 panic 时跳转到上面的 defer；
		// 恢复后返回外层 Logging，不会回到触发 panic 的那行继续执行。
		c.Next()
	}
}
