package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error)
}

func RequirePermission(permissionChecker PermissionChecker, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 查询当前用户是否拥有指定权限
		allowed, err := permissionChecker.HasPermission(c.Request.Context(), userID, permissionCode)

		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !allowed {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 有权限，继续执行后续中间件或接口处理函数。
		c.Next()
	}
}
