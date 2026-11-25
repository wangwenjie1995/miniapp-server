// Package global 保存启动时初始化的进程级共享资源。
package global

import (
	"log/slog"

	"gorm.io/gorm"

	"miniapp-server/internal/auth"
)

// 在创建路由前赋值，服务运行期间各模块复用这些资源。
var (
	DB     *gorm.DB
	Tokens *auth.TokenManager
	Logger = slog.Default()
)
