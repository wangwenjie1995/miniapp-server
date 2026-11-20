package router

import (
	"github.com/gin-gonic/gin"

	"miniapp-server/internal/home"
	"miniapp-server/internal/platform"
	"miniapp-server/internal/user"
)

// Repositories 允许组装时替换各模块的数据实现。
type Repositories struct {
	Home home.Repository
	User user.Repository
}

// New 为各模块创建共享连接池的 Repository，再统一注册路由。
func New(resources platform.Resources) *gin.Engine {
	return NewWithDependencies(resources, Repositories{
		Home: home.NewRepository(resources.DB),
		User: user.NewRepository(resources.DB),
	})
}
