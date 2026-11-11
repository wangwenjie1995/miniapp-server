package user

import "errors"

// 用户模块的业务错误，由 Handler 转换成接口响应。
var (
	ErrUsernameExists     = errors.New("用户名已存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
)
