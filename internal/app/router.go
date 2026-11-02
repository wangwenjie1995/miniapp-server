// 本文件负责注册 URL 与处理函数的对应关系，和 app.go 同属 app 包。
package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"miniapp-server/internal/home"
	"miniapp-server/internal/middleware"
	"miniapp-server/internal/response"
)

// NewRouter 返回 Gin 引擎。它实现 http.Handler 接口，可以直接交给 http.Server 使用。
func NewRouter(logger *slog.Logger, repository home.Repository) *gin.Engine {
	// 创建不带默认日志和异常恢复中间件的引擎，由项目自己安装这些能力。
	router := gin.New()
	// Use 将中间件应用于所有路由。请求顺序：Logging → Recovery → 路由函数。
	// 处理结束时逐层返回，因此外层 Logging 能记录异常恢复后的最终状态码。
	router.Use(middleware.Logging(logger), middleware.Recovery(logger))
	// 只有 GET /healthz 请求会进入这个函数。GET 是方法，/healthz 是路径。
	// *gin.Context 保存本次请求和响应信息，与传递取消信号的 context.Context 不同。
	router.GET("/healthz", func(c *gin.Context) {
		// http.StatusOK 就是 200。gin.H 是 map[string]any 的简写，序列化后是 JSON 对象。
		// 该接口只检查服务能否响应，目前不检查数据库等外部依赖。
		response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
	})
	// 业务接口统一挂在 /api/v1 下，具体首页路由交给 home 模块登记。
	home.RegisterRoutes(router.Group("/api/v1"), home.NewService(repository), logger)
	// 返回配置好的引擎；真正的端口监听在 App.Run 中执行。
	return router
}
