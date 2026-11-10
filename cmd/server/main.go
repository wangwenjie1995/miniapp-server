// main 包代表可执行程序；运行 go run ./cmd/server 时，从下面的 main() 开始执行。
// 建议阅读顺序：main.go → config.go → app.go → router.go → middleware → response。
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"miniapp-server/internal/app"
	"miniapp-server/internal/config"
	"miniapp-server/internal/database"
	"miniapp-server/internal/home"
	"miniapp-server/internal/user"
)

// main 是整个程序的入口，只负责设置运行模式、创建日志器，以及统一处理退出错误。
func main() {
	// Release 模式关闭 Gin 的调试输出；它不会改变代码的编译优化等级。
	gin.SetMode(gin.ReleaseMode)
	// slog 是 Go 标准库的日志工具。JSONHandler 把日志转成 JSON，便于程序检索。
	// os.Stdout 表示输出到终端；nil 表示使用默认选项，默认记录 INFO 及以上级别。
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// Go 通常通过返回 error 表示失败。nil 表示没有错误，非 nil 表示需要处理。
	// 这里的 err 只在这个 if 语句及其代码块中有效。
	if err := run(logger); err != nil {
		logger.Error("服务退出", "error", err)
		// 退出码 1 表示失败，方便终端或部署系统识别；正常执行完 main 则退出码为 0。
		// os.Exit 不执行 defer，所以资源清理放在 run 中，返回这里之前完成。
		os.Exit(1)
	}
}

// run 串起启动流程，并在服务运行期间等待。*slog.Logger 表示传入日志器的指针，
// 各模块使用同一个日志器；返回 error 让 main 统一记录失败。
func run(logger *slog.Logger) error {
	// 第一步：读取并校验配置。配置错误时直接返回，不继续启动服务。
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	// 第二步：创建退出通知。context 可以理解为在不同函数间传递取消信号的对象。
	// Background 创建没有截止时间的根 context；Ctrl+C 或 SIGTERM 到来时，ctx 被取消。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer 表示在当前函数返回时调用；stop 用于释放信号监听相关资源。
	defer stop()
	// 第三步：New 组装服务但不监听端口；Run 才启动服务，并等待错误或退出信号。
	// app 是 miniapp-server 的核心模块，负责组装路由、注册中间件、启动 HTTP 服务。
	// 在开始监听 HTTP 前检查 MySQL；连接失败则退出，不回退到固定数据。
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, pool, err := database.Open(connectCtx, cfg.DB)
	if err != nil {
		return err
	}
	deps := app.RouterDeps{
		Logger:         logger,
		HomeRepository: home.NewRepository(db),
		UserRepository: user.NewRepository(db),
	}
	// HTTP 服务完成优雅关闭后，才释放数据库连接池。
	defer pool.Close()
	return app.New(cfg, deps).Run(ctx)
}
