package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"miniapp-server/internal/middleware"
	"miniapp-server/internal/platform"
	"miniapp-server/internal/response"
)

// New 创建 Gin 引擎并注册中间件和业务路由。
func New(resources platform.Resources) *gin.Engine {
	// 创建不带默认日志和异常恢复中间件的引擎，由项目自己安装这些能力。
	router := gin.New()
	// Use 将中间件应用于所有路由。请求顺序：Logging → Recovery → 路由函数。
	// 处理结束时逐层返回，因此外层 Logging 能记录异常恢复后的最终状态码。
	router.Use(
		middleware.Logging(resources.Log()),
		middleware.Recovery(resources.Log()),
	)

	// 健康检查保持公开，只检查服务能否响应，不检查数据库等外部依赖。
	router.GET("/healthz", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
	})

	registerModuleRoutes(router, resources)

	// 返回配置好的引擎；真正的端口监听在 App.Run 中执行。
	return router
}

func registerModuleRoutes(router *gin.Engine, resources platform.Resources) {
	publicGroup := router.Group("/api/v1")
	authenticatedGroup := publicGroup.Group("")
	authenticatedGroup.Use(middleware.Auth(resources.Tokens))

	initUserRoutes(publicGroup, authenticatedGroup, resources)
	initHomeRoutes(publicGroup, resources)
}
