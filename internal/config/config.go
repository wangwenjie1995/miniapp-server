// config 包集中处理运行配置，避免其他模块各自读取环境变量。
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config 保存服务启动所需配置。struct 是把多个相关字段组合在一起的数据类型。
// 大写开头的类型、字段和函数可以被其他包访问。
// type是啥？ type是Go语言中用于定义新类型的关键字。它允许你创建自定义的数据类型，可以是结构体、接口、函数类型等。在这里，type Config struct {...} 定义了一个名为 Config 的结构体类型，用于保存服务启动所需的配置。
// struct是啥？ struct是Go语言中的一种复合数据类型，用于将多个不同类型的字段组合在一起，形成一个新的数据结构。在这里，struct {...} 定义了一个结构体，其中包含了服务启动所需的配置字段，例如HTTPAddr。
// 还有哪些结构体？在 Go 语言中，结构体（struct）是一种复合数据类型，可以包含多个字段，每个字段可以是不同的数据类型。除了 Config 结构体之外，Go 语言中还有许多其他常见的结构体，例如：
// 1. time.Time：用于表示时间和日期的结构体，包含年、月、日、时、分、秒等字段。
// 2. net.IP：用于表示 IP 地址的结构体，包含 IPv4 和 IPv6 地址的相关信息。
// 3. os.FileInfo：用于表示文件信息的结构体，包含文件名、大小、权限等字段。
// 4. http.Request：用于表示 HTTP 请求的结构体，包含请求方法、URL、头部信息等字段。
// 5. http.Response：用于表示 HTTP 响应的结构体，包含状态码、头部信息、响应体等字段。
// 这些结构体在 Go 标准库中广泛使用，用于处理各种不同类型的数据和操作。
type Config struct {
	// HTTPAddr 是监听地址，例如 :8080（所有本机网卡）或 127.0.0.1:8080（仅本机）。
	HTTPAddr string
	// DB 保存数据库连接参数；密码仅从本地私有文件或环境变量读取。
	DB  Database
	JWT JWTConfig
}

type JWTConfig struct {
	Secret   string
	Duration time.Duration
}

// Database 是连接配置，不是数据库连接对象。
type Database struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// Load 返回两个值：校验通过的配置，以及错误。成功时 error 为 nil。
// 可读取项目根目录 .env.local，进程环境变量优先；不自动加载 .env.example。
// (Config, error)是返回值的类型，表示函数 Load 会返回一个 Config 类型的值和一个 error 类型的值。Config 类型用于保存服务启动所需的配置，而 error 类型用于表示函数执行过程中是否发生了错误。如果函数执行成功，error 将为 nil；如果发生错误，error 将包含错误信息。
func Load() (Config, error) {
	// 可选加载项目根目录的私有文件；已经设置的环境变量优先，不会被文件覆盖。
	// 文件不存在是正常情况，格式错误或权限错误则显式返回。
	if err := godotenv.Load(".env.local"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("读取 .env.local 失败: %w", err)
	}
	// LookupEnv 同时返回值和“变量是否存在”，用于区分未设置与显式设置为空字符串。
	addr, exists := os.LookupEnv("HTTP_ADDR")
	// 只有完全没有设置 HTTP_ADDR 时才使用默认值，空字符串应该暴露为配置错误。
	if !exists {
		addr = ":8080"
	}
	// 拆分 host 和 port，同时检查地址格式。IPv6 地址需要写成 [::1]:8080。
	// 下划线 _ 表示忽略 host 返回值；这里仅验证格式和端口，实际监听由 app 完成。
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		// Config{} 是零值配置；%w 将底层错误包进新的错误，保留错误原因。
		return Config{}, fmt.Errorf("HTTP_ADDR 必须为 host:port 格式: %w", err)
	}
	// 端口最初是字符串；Atoi 将它转为整数，例如 "8080" → 8080。
	number, err := strconv.Atoi(port)
	// || 表示“或者”：不是整数，或超出约定范围，都拒绝启动。
	if err != nil || number < 1 || number > 65535 {
		return Config{}, fmt.Errorf("HTTP_ADDR 端口必须为 1 至 65535: %q", port)
	}
	// 返回已经校验的配置，交给 app 创建 HTTP 服务。
	db := Database{Host: env("DB_HOST", "127.0.0.1"), Port: env("DB_PORT", "3306"),
		Name: env("DB_NAME", "miniapp"), User: env("DB_USER", "root"), Password: os.Getenv("DB_PASSWORD")}
	portNumber, err := strconv.Atoi(db.Port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf("DB_PORT 必须为 1 至 65535 的整数")
	}
	if strings.TrimSpace(db.Host) == "" || strings.TrimSpace(db.Name) == "" || strings.TrimSpace(db.User) == "" {
		return Config{}, fmt.Errorf("DB_HOST、DB_NAME、DB_USER 不能为空")
	}
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET 至少需要 32 字节")
	}

	duration, err := time.ParseDuration(os.Getenv("JWT_DURATION"))
	if err != nil || duration <= 0 {
		return Config{}, fmt.Errorf("JWT_DURATION 必须是大于 0 的有效时长，例如 2h")
	}
	return Config{HTTPAddr: addr, DB: db, JWT: JWTConfig{
		Secret:   secret,
		Duration: duration,
	}}, nil
}

// env 仅在变量未设置时使用默认值，不用默认值掩盖显式空配置。
func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
