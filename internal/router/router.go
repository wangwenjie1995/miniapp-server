package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"miniapp-server/internal/home"
	"miniapp-server/internal/middleware"
	"miniapp-server/internal/platform"
	"miniapp-server/internal/response"
	"miniapp-server/internal/user"
)

// NewWithDependencies 允许测试构建使用独立仓库、日志器和 Token 的路由。
func NewWithDependencies(resources platform.Resources, repos Repositories) *gin.Engine {
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

	api := router.Group("/api/v1")
	routes := newGroups(api, resources, repos.User)
	users := user.NewHandler(user.NewService(repos.User), resources)
	registerUser(routes, users)

	homepage := home.NewHandler(home.NewService(repos.Home), resources)
	registerHome(routes.public, homepage)

	// 返回配置好的引擎；真正的端口监听在 App.Run 中执行。
	return router
}
