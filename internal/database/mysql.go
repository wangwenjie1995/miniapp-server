// database 包只负责 MySQL 连接，不处理首页或其他业务。
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"miniapp-server/internal/config"
)

// Open 创建连接池并检查连接。调用方必须在退出时关闭返回的 sql.DB。
// gorm.DB 是查询入口，sql.DB 是底层连接池；不是每个请求重新建立一个连接池。
func Open(ctx context.Context, cfg config.Database) (*gorm.DB, *sql.DB, error) {
	// 使用驱动的配置对象生成 DSN，避免手工拼接密码中的特殊字符。
	options := driver.NewConfig()
	options.User, options.Passwd = cfg.User, cfg.Password
	options.Net, options.Addr = "tcp", net.JoinHostPort(cfg.Host, cfg.Port)
	options.DBName, options.ParseTime, options.Loc = cfg.Name, true, time.UTC
	options.Params = map[string]string{"charset": "utf8mb4"}
	options.Timeout, options.ReadTimeout, options.WriteTimeout = 5*time.Second, 5*time.Second, 5*time.Second
	pool, err := sql.Open("mysql", options.FormatDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("创建 MySQL 连接池失败: %w", err)
	}
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(30 * time.Minute)
	pool.SetConnMaxIdleTime(5 * time.Minute)
	if err := pool.PingContext(ctx); err != nil {
		_ = pool.Close()
		return nil, nil, fmt.Errorf("连接 MySQL 失败，请检查 DB_* 配置: %w", err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: pool, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		// SQL 错误由业务层统一写入 slog，避免 GORM 输出重复日志或 SQL 参数。
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		_ = pool.Close()
		return nil, nil, fmt.Errorf("初始化 GORM 失败: %w", err)
	}
	return db, pool, nil
}
