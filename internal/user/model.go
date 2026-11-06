package user

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=24"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=4,max=24"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UserResponse struct {
	ID uint64 `json:"id"`
	Username string `json:"username"`
}

// expiresAt 是 token 的过期时间，前端可以根据这个时间判断是否需要刷新 token。
type LoginResponse struct {
	Token string `json:"token"`
	User UserResponse `json:"user"`
	ExpiresAt time.Time `json:"expiresAt"`
}