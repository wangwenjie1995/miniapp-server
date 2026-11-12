// 本文件负责注册 URL 与处理函数的对应关系，和 app.go 同属 app 包。
package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"miniapp-server/internal/home"
	"miniapp-server/internal/middleware"
	"miniapp-server/internal/response"
	"miniapp-server/internal/user"
)

// NewRouter 返回 Gin 引擎。它实现 http.Handler 接口，可以直接交给 http.Server 使用。
func NewRouter(deps RouterDeps) *gin.Engine {
	// 创建不带默认日志和异常恢复中间件的引擎，由项目自己安装这些能力。
	router := gin.New()
	// Use 将中间件应用于所有路由。请求顺序：Logging → Recovery → 路由函数。
	// 处理结束时逐层返回，因此外层 Logging 能记录异常恢复后的最终状态码。
	router.Use(
		middleware.Logging(deps.Logger),
		middleware.Recovery(deps.Logger),
	)

	// 健康检查保持公开，只检查服务能否响应，不检查数据库等外部依赖。
	router.GET("/healthz", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")

	// 两个子组共用路径前缀，各自配置中间件；注册和登录保持公开。
	public := api.Group("")
	user.RegisterRoutes(public, user.Dependencies{
		Service:      user.NewService(deps.UserRepository),
		Logger:       deps.Logger,
		TokenManager: deps.TokenManager,
	})

	// 鉴权只作用于此子组，首页接口需要有效的 Token。
	protected := api.Group("", middleware.Auth(deps.TokenManager))
	home.RegisterRoutes(protected, home.Dependencies{
		Service: home.NewService(deps.HomeRepository),
		Logger:  deps.Logger,
	})

	// 返回配置好的引擎；真正的端口监听在 App.Run 中执行。
	return router
}
