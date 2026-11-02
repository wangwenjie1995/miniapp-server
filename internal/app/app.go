// app 包负责组装配置、路由和 HTTP 服务，管理从启动到关闭的整个生命周期。
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"miniapp-server/internal/config"
	"miniapp-server/internal/home"
)

// App 保存服务运行所需的对象。字段小写开头，只能由本包直接访问。
type App struct {
	// 标准库服务器管理网络连接、HTTP 协议和超时；Gin 负责具体请求的路由。
	server *http.Server
	logger *slog.Logger
}

// New 根据配置创建 App，但还没有监听端口。
// *App 是指向 App 的指针，让后续方法操作同一个服务实例。
func New(cfg config.Config, logger *slog.Logger, repository home.Repository) *App {
	// & 取得新对象的地址，这里同时构造 App 和内部的 http.Server。
	return &App{
		// 为啥用& &？
		// & 是 Go 语言中的取地址符号，用于获取变量的内存地址。在这里，&App{...} 创建了一个新的 App 结构体实例，并返回其指针。使用指针可以让后续方法操作同一个服务实例，而不是复制一份新的实例，从而节省内存和提高性能。
		server: &http.Server{
			// 监听地址来自已经校验的配置，例如 :8080。
			Addr: cfg.HTTPAddr,
			// 把标准库 HTTP 服务内部的错误日志转交 slog，统一日志格式和输出位置。
			ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
			// Handler 是请求入口，每个请求交给这里创建的 Gin 路由处理。
			Handler: NewRouter(logger, repository),
			// 读取请求头的最长时间，避免客户端长时间占用连接却不发完整的头部。
			ReadHeaderTimeout: 5 * time.Second,
			// 读取整个请求（包含请求体）的最长时间。
			ReadTimeout: 10 * time.Second,
			// 响应写入阶段的超时；它不是强制停止业务函数执行的计时器。
			WriteTimeout: 15 * time.Second,
			// 一个长连接暂时没有请求时，最多等待 60 秒接收下一次请求。
			IdleTimeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// Run 是 App 的方法，a 指向调用该方法的实例，ctx 用于接收外部取消通知。
// 本函数持续等待服务结束或退出信号，返回 nil 表示成功，返回 error 表示失败。
func (a *App) Run(ctx context.Context) error {
	// 先绑定 TCP 端口。端口被占用、主机地址无效等问题在这里立刻返回。
	listener, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return fmt.Errorf("监听 %s 失败: %w", a.server.Addr, err)
	}
	// 只有端口绑定成功后才记录启动日志。
	a.logger.Info("HTTP 服务启动", "addr", listener.Addr().String())
	// channel 是 goroutine 之间传递数据的通道，这里用于传回服务器结束时的错误。
	// 容量为 1，允许发送方先放入结果，不必等待接收方同时读取。
	result := make(chan error, 1)
	// go 开启一个并发执行的 goroutine。Serve 持续接受连接，通常不会立即返回；
	// 放到另一个 goroutine 后，当前流程就能同时等待退出通知。
	go func() {
		// <- 左边是通道时表示发送：先等待 Serve 结束，再把它的返回值放进通道。
		result <- a.server.Serve(listener)
	}()

	// select 等待多个通道，哪个就绪就执行哪个分支；等待时不会持续空转。
	select {
	// <- 右边是通道时表示接收。这里说明服务器已经结束，读取结束原因。
	case err := <-result:
		// ErrServerClosed 表示服务器被主动关闭，是正常结束，不当作故障。
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("HTTP 服务运行失败: %w", err)
	// Ctrl+C 或 SIGTERM 使 context 取消，Done 通道关闭，触发这个分支。
	case <-ctx.Done():
		a.logger.Info("HTTP 服务开始关闭")
		// 原 ctx 已经取消，不能再用它等待收尾；创建一个新的、最多等待 10 秒的 context。
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// defer 在当前函数返回时执行；cancel 释放计时器等资源，不必等满 10 秒。
		defer cancel()
		// Shutdown 停止接受新连接，关闭空闲连接，等待正在处理的请求完成。
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			// 如果优雅关闭超时或失败，则强制关闭连接。
			// _ = 明确忽略 Close 返回值，保留上面的 Shutdown 错误作为最终退出原因。
			_ = a.server.Close()
			return fmt.Errorf("HTTP 服务优雅关闭失败: %w", err)
		}
		// 等待 Serve 所在 goroutine 结束，检查它是否按预期返回 ErrServerClosed。
		if err := <-result; !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP 服务关闭失败: %w", err)
		}
		// 到这里收尾成功，返回 nil，让 main 正常结束，进程退出码为 0。
		a.logger.Info("HTTP 服务已关闭")
		return nil
	}
}
