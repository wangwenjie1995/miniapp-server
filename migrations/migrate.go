// migrations 提供显式执行的 SQL 迁移，正常启动 HTTP 服务不会调用这里。
package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"
)

//go:embed *.sql
var scripts embed.FS

// Apply 按文件名顺序执行 SQL 并记录校验和，已执行版本不能再被悄悄修改。
// MySQL 的建表语句会隐式提交，因此不伪装成可整体回滚的事务。
func Apply(ctx context.Context, pool *sql.DB) error {
	// 使用同一条连接持有 MySQL 命名锁，避免两个初始化进程同时迁移。
	conn, err := pool.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var locked int
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(SHA2(CONCAT(DATABASE(), ':home_migrations'), 256), 5)").Scan(&locked); err != nil {
		return err
	}
	if locked != 1 {
		return fmt.Errorf("另一个进程正在迁移，请稍后重试")
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK(SHA2(CONCAT(DATABASE(), ':home_migrations'), 256))")
	}()
	_, err = conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS home_schema_migrations (
        version VARCHAR(100) PRIMARY KEY, checksum CHAR(64) NOT NULL,
        applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		return err
	}
	entries, err := scripts.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		content, err := scripts.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(content))
		var existing string
		err = conn.QueryRowContext(ctx, "SELECT checksum FROM home_schema_migrations WHERE version = ?", entry.Name()).Scan(&existing)
		if err == nil {
			if existing != checksum {
				return fmt.Errorf("迁移 %s 已执行，不能修改原文件，请新增迁移", entry.Name())
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		// 本项目 SQL 迁移只包含普通 DDL，不包含存储过程或字符串中的分号。
		// 按语句执行，不开启驱动 multiStatements，避免放宽运行时查询能力。
		for _, statement := range strings.Split(string(content), ";") {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if _, err := conn.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("执行 %s 失败: %w", entry.Name(), err)
			}
		}
		if _, err := conn.ExecContext(ctx, "INSERT INTO home_schema_migrations(version, checksum) VALUES (?, ?)", entry.Name(), checksum); err != nil {
			return err
		}
	}
	return nil
}
