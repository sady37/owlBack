// wisefido-ai-health/cmd/health-etl
//
// AI Health 离线 ETL 服务（设计文档 doc/AI_health.md §0.3 / §8.5 / §9）。
// 进程独立于 wisefido-ai，避免批处理 OOM/慢查询拖累实时通路（roomengine / fall_verifier）。
//
// 启动模式：
//   - 默认（守护）：调度器持续运行，按 schedule 配置触发 daily / monthly
//   - --once daily：立即跑一次 daily_etl 然后退出（CI / 回填用）
//   - --once monthly：立即跑一次 monthly_etl 然后退出
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"wisefido-ai-health/internal/config"
	"wisefido-ai-health/internal/etl"
	"wisefido-ai-health/internal/repo"
	"wisefido-ai-health/internal/scheduler"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	once := flag.String("once", "", `run once and exit: "daily" | "monthly"`)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := mustLogger(cfg.Logging.Level)
	defer func() { _ = logger.Sync() }()

	logger.Info("starting wisefido-ai-health",
		zap.String("db_host", cfg.DB.Host),
		zap.String("db_name", cfg.DB.Database),
		zap.String("timezone", cfg.Schedule.Timezone),
		zap.String("daily_at", cfg.Schedule.DailyAt),
		zap.Int("monthly_day", cfg.Schedule.MonthlyDay),
		zap.String("monthly_at", cfg.Schedule.MonthlyAt),
		zap.Bool("dry_run", cfg.ETL.DryRun),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := repo.Open(ctx, cfg.DB)
	if err != nil {
		logger.Fatal("connect owlrd failed", zap.Error(err))
	}
	defer func() { _ = r.Close() }()
	logger.Info("owlrd connected")

	runner := etl.New(cfg, r, logger)

	// --once 模式：立即跑一次然后退出
	if *once != "" {
		now := time.Now()
		switch strings.ToLower(*once) {
		case "daily":
			runDate := now.AddDate(0, 0, -1)
			if err := runner.RunDaily(ctx, runDate); err != nil {
				logger.Fatal("once daily failed", zap.Error(err))
			}
		case "monthly":
			runMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
			if err := runner.RunMonthly(ctx, runMonth); err != nil {
				logger.Fatal("once monthly failed", zap.Error(err))
			}
		default:
			logger.Fatal("unknown --once mode", zap.String("mode", *once))
		}
		logger.Info("--once complete, exit")
		return
	}

	// 守护模式：调度器
	sch := scheduler.New(cfg.Schedule, logger)
	go sch.RunDaily(ctx, runner.RunDaily)
	go sch.RunMonthly(ctx, runner.RunMonthly)

	// 信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Info("signal received, shutting down", zap.String("signal", sig.String()))
	cancel()

	// 给后台 goroutine 1 秒优雅退出
	time.Sleep(1 * time.Second)
}

func mustLogger(level string) *zap.Logger {
	zc := zap.NewProductionConfig()
	zc.EncoderConfig.TimeKey = "timestamp"
	zc.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	if level != "" {
		var lvl zapcore.Level
		if err := lvl.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(level)))); err == nil {
			zc.Level = zap.NewAtomicLevelAt(lvl)
		}
	}
	logger, err := zc.Build()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	return logger
}
