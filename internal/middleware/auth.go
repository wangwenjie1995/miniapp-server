package middleware

import (
	"miniapp-server/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth(tokenManager *auth.TokenManager) gin.HandlerFunc {
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

		userID, err := tokenManager.Verify(token)

		if err != nil {
			c.Abort()
			c.String(http.StatusUnauthorized, "请先登录或重新登录")
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
