package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/service"

	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger, err := initLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	// 3. 获取租户ID（从环境变量或配置）
	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		logger.Fatal("TENANT_ID environment variable is required")
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

// initLogger 初始化日志
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Log.Format == "json" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	return logger, nil
}

