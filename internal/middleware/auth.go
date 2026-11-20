package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"miniapp-server/internal/auth"
)

const userIDKey = "userID"

// GetUserIDFromContext 返回鉴权中间件写入的用户 ID。
// 值不存在、类型不匹配或 ID 为零时，返回 false。
func GetUserIDFromContext(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(userIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := value.(uint64)
	return userID, ok && userID != 0
}

func Auth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if header == "" {
			c.Abort()
			c.String(http.StatusUnauthorized, "请先登录或重新登录")
			return
		}
		parts := strings.Split(header, " ")

		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			c.Abort()
			c.String(http.StatusUnauthorized, "请先登录或重新登录")
			return
		}

		token := parts[1]

		if tokens == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		userID, err := tokens.Verify(token)

		if err != nil {
			c.Abort()
			c.String(http.StatusUnauthorized, "请先登录或重新登录")
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}
