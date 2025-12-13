package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"go.uber.org/zap"
	"wisefido-sleepace/internal/config"
	"wisefido-sleepace/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// 初始化Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	
	logger.Info("Starting wisefido-sleepace service",
		zap.String("version", "1.5.0"),
		zap.String("mqtt_broker", cfg.MQTT.Broker),
		zap.String("mqtt_topic", cfg.Sleepace.Topic),
		zap.String("stream", cfg.Sleepace.Stream),
	)
	
	// 创建服务
	sleepaceService, err := service.NewSleepaceService(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create sleepace service", zap.Error(err))
	}
	
	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 在 goroutine 中启动服务
	go func() {
		if err := sleepaceService.Start(ctx); err != nil {
			logger.Fatal("Failed to start sleepace service", zap.Error(err))
		}
	}()
	
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	
	// 优雅关闭
	cancel()
	if err := sleepaceService.Stop(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}
	
	logger.Info("Service stopped")
}

