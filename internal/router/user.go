package router

import (
	"github.com/gin-gonic/gin"

	"miniapp-server/internal/global"
	"miniapp-server/internal/middleware"
	"miniapp-server/internal/user"
)

func initUserRoutes(publicGroup, authenticatedGroup *gin.RouterGroup) {
	repo := user.NewRepository(global.DB)
	handler := user.NewHandler(repo, global.Logger, global.Tokens)
	RegisterUserRoutes(publicGroup, authenticatedGroup, handler, repo)
}

func RegisterUserRoutes(publicGroup, authenticatedGroup *gin.RouterGroup, handler *user.Handler, checker middleware.PermissionChecker) {
	public := publicGroup.Group("/user")
	public.POST("/register", handler.Register)
	public.POST("/login", handler.Login)

	protected := authenticatedGroup.Group("/user")
	protected.GET("/me", handler.GetCurrentUser)
	protected.GET("/list", middleware.RequirePermission(checker, "user:list"), handler.ListUsers)
}
