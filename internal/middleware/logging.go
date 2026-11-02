// middleware 包存放多个接口共用的请求处理步骤，例如日志和异常恢复。
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging 接收日志器，返回一个 Gin 中间件。
// 外层函数在注册时执行，返回的内部函数在每次请求时执行。
func Logging(logger *slog.Logger) gin.HandlerFunc {
	// c 是当前请求的上下文，用来访问请求数据和响应状态。
	return func(c *gin.Context) {
		// 在进入处理链时记录起点，后面计算本次请求的总处理耗时。
		start := time.Now()
		// Next 执行后续中间件和路由函数；它们返回后，再继续执行下面的日志代码。
		c.Next()
		// 这里记录处理完成后的状态码，包含异常恢复产生的 500。
		// 参数以“键、值”成对传入；只记录路径，不记录查询参数、请求头和请求体。
		logger.Info("HTTP 请求",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
		)
	}
}
