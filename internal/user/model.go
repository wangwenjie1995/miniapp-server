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
	ID       uint64 `json:"id"`
	Username string `json:"username"`
}

// LoginResponse 保持用户字段平铺，并附带登录凭证及其过期时间。
type LoginResponse struct {
	UserResponse
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
