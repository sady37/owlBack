package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"go.uber.org/zap"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/service"
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
	
	logger.Info("Starting wisefido-radar service",
		zap.String("version", "1.5.0"),
		zap.String("mqtt_broker", cfg.MQTT.Broker),
	)
	
	// 创建服务
	radarService, err := service.NewRadarService(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create radar service", zap.Error(err))
	}
	
	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := radarService.Start(ctx); err != nil {
		logger.Fatal("Failed to start radar service", zap.Error(err))
	}
	
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	
	// 优雅关闭
	cancel()
	if err := radarService.Stop(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}
	
	logger.Info("Service stopped")
}

