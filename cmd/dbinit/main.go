// dbinit 是首页数据库初始化命令，与 HTTP 服务分开执行。
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"miniapp-server/internal/config"
	"miniapp-server/internal/database"
	"miniapp-server/internal/home"
	"miniapp-server/migrations"
)

func main() {
	seed := flag.Bool("seed", false, "建表后导入首页演示数据，不覆盖已有记录")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(*seed); err != nil {
		logger.Error("首页数据库初始化失败", "error", err)
		os.Exit(1)
	}
	logger.Info("首页数据库初始化完成")
}

func run(seed bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	db, pool, err := database.Open(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := migrations.Apply(ctx, pool); err != nil {
		return err
	}
	if seed {
		return home.Seed(ctx, db)
	}
	return nil
}
