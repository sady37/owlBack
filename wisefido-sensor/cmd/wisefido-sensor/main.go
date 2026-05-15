package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"wisefido-sensor/internal/config"
	"wisefido-sensor/internal/consumer"
	"wisefido-sensor/internal/service"

	logpkg "owl-common/logger"

	_ "github.com/lib/pq"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-sensor")
	if err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	// 3. 业务租户 ID（可选；内置租户由建库脚本写入，与启动无关）
	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		logger.Info("TENANT_ID unset — wisefido-sensor starts; card polling uses no tenant until set in .env")
	}

	// 4. 创建服务
	alarmService, err := service.NewAlarmService(cfg, logger, tenantID)
	if err != nil {
		logger.Fatal("Failed to create alarm service",
			zap.Error(err),
		)
	}
	defer alarmService.Stop()

	// 5. 创建上下文（支持优雅关闭）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5.1 RoomEngine 实时学习（独立 db/redis 连接，避免与 alarm service 共享状态）
	//      Phase 5（家属反馈接入 RecordGroundTruth）暂未开启 —— winner 用 yaml 默认 balanced
	if engineDB, engineRedis, err := openEngineDeps(cfg); err != nil {
		logger.Warn("roomengine deps init failed; engine disabled", zap.Error(err))
	} else {
		if _, err := startRoomEngine(ctx, cfg, engineDB, engineRedis, logger); err != nil {
			logger.Warn("roomengine startup failed; engine disabled", zap.Error(err))
		}
	}

	// 5.2 PR1 (A7): sensor 端 monitor 流消费者 + buffer。
	//      与 cardagg 用独立 consumer group (wisefido-sensor-monitor)，offset 互不干扰。
	//      本 PR 仅喂 buffer；后续 B/C 组迁移会接 DeriveAndWriteState / DeriveBedStateFromRealtime。
	if _, engineRedis, err := openEngineDeps(cfg); err == nil {
		monitorBuf := service.NewMonitorBuffer()
		monitorConsumer := consumer.NewMonitorConsumer(engineRedis, monitorBuf, logger)
		monitorConsumer.Start(ctx)
	} else {
		logger.Warn("sensor monitor consumer disabled (redis init failed)", zap.Error(err))
	}

	// 6. 启动服务（在 goroutine 中）
	serviceErrChan := make(chan error, 1)
	go func() {
		if err := alarmService.Start(ctx); err != nil {
			serviceErrChan <- err
		}
	}()

	// 7. 等待信号（优雅关闭）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down",
			zap.String("signal", sig.String()),
		)
		cancel() // 取消上下文，停止服务
	case err := <-serviceErrChan:
		logger.Fatal("Service error",
			zap.Error(err),
		)
	}

	logger.Info("Alarm service stopped")
}

// openEngineDeps 为 RoomEngine 单独建 db + redis 连接（与 alarm service 隔离）
func openEngineDeps(cfg *config.Config) (*sql.DB, *redis.Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, nil, fmt.Errorf("ping db: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}
	return db, rdb, nil
}

