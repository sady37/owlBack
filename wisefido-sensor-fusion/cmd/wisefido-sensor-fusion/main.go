package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	logpkg "owl-common/logger"
	
	"go.uber.org/zap"
	"wisefido-sensor-fusion/internal/config"
	"wisefido-sensor-fusion/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// 初始化Logger
	logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	
	logger.Info("Starting wisefido-sensor-fusion service",
		zap.String("version", "1.5.0"),
		zap.String("input_stream", cfg.Fusion.Stream.Input),
		zap.String("cache_prefix", cfg.Fusion.Cache.RealtimeKeyPrefix),
		zap.Int("cache_ttl", cfg.Fusion.Cache.RealtimeTTL),
	)
	
	// 创建服务
	fusionService, err := service.NewFusionService(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create fusion service", zap.Error(err))
	}
	
	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 在 goroutine 中启动服务
	go func() {
		if err := fusionService.Start(ctx); err != nil {
			logger.Fatal("Failed to start fusion service", zap.Error(err))
		}
	}()
	
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	
	// 优雅关闭
	cancel()
	if err := fusionService.Stop(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}
	
	logger.Info("Service stopped")
}

