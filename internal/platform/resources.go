// Package platform 定义启动时创建、各模块共享的基础资源。
package platform

import (
	"log/slog"

	"gorm.io/gorm"

	"miniapp-server/internal/auth"
)

// Resources 在启动时组装，服务运行期间复用。
type Resources struct {
	DB     *gorm.DB
	Tokens *auth.TokenManager
	Logger *slog.Logger
}

// Log 支持测试使用独立日志器，默认使用进程日志器。
func (r Resources) Log() *slog.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return slog.Default()
}
