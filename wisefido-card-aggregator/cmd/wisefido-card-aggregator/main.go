package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/service"

	logpkg "owl-common/logger"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting wisefido-card-aggregator service")

	// 创建服务
	svc, err := service.NewAggregatorService(cfg, log)
	if err != nil {
		log.Fatal("Failed to create aggregator service", zap.Error(err))
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动服务（在 goroutine 中）
	errChan := make(chan error, 1)
	go func() {
		if err := svc.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// 等待信号或错误
	select {
	case sig := <-sigChan:
		log.Info("Received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
	case err := <-errChan:
		log.Error("Service error", zap.Error(err))
		cancel()
	}

	// 停止服务
	if err := svc.Stop(ctx); err != nil {
		log.Error("Error stopping service", zap.Error(err))
	}

	log.Info("Service stopped")
}
