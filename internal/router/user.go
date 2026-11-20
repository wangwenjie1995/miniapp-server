package router

import (
	"miniapp-server/internal/user"
)

func registerUser(routes groups, h *user.Handler) {
	public := routes.public.Group("/user")
	public.POST("/register", h.Register)
	public.POST("/login", h.Login)

	protected := routes.authenticated.Group("/user")
	protected.GET("/me", h.GetCurrentUser)
	protected.GET("/list", routes.require("user:list"), h.ListUsers)
}
