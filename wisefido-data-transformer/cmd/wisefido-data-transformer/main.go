package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"go.uber.org/zap"
	"wisefido-data-transformer/internal/config"
	"wisefido-data-transformer/internal/service"
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
	
	logger.Info("Starting wisefido-data-transformer service",
		zap.String("version", "1.5.0"),
		zap.String("radar_stream", cfg.Transformer.Streams.Radar),
		zap.String("sleepace_stream", cfg.Transformer.Streams.Sleepace),
		zap.String("output_stream", cfg.Transformer.Streams.Output),
	)
	
	// 创建服务
	transformerService, err := service.NewTransformerService(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create transformer service", zap.Error(err))
	}
	
	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 在 goroutine 中启动服务
	go func() {
		if err := transformerService.Start(ctx); err != nil {
			logger.Fatal("Failed to start transformer service", zap.Error(err))
		}
	}()
	
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	
	// 优雅关闭
	cancel()
	if err := transformerService.Stop(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}
	
	logger.Info("Service stopped")
}

