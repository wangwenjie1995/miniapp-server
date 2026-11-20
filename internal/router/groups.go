package router

import (
	"github.com/gin-gonic/gin"

	"miniapp-server/internal/middleware"
	"miniapp-server/internal/platform"
)

// groups 让业务模块选择公开或登录路由，并按需声明业务权限。
type groups struct {
	public        *gin.RouterGroup
	authenticated *gin.RouterGroup
	checker       middleware.PermissionChecker
}

func newGroups(public *gin.RouterGroup, resources platform.Resources, checker middleware.PermissionChecker) groups {
	return groups{
		public:        public,
		authenticated: public.Group("", middleware.Auth(resources.Tokens)),
		checker:       checker,
	}
}

func (g groups) require(code string) gin.HandlerFunc {
	return middleware.RequirePermission(g.checker, code)
}
