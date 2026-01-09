package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wisefido-card-manage/internal/config"
	httphandler "wisefido-card-manage/internal/http"
	"wisefido-card-manage/internal/service"

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
	log, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-card-manage")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting wisefido-card-manage service",
		zap.String("port", cfg.Server.Port),
	)

	// 创建卡片管理服务
	cardService, err := service.NewCardService(cfg, log)
	if err != nil {
		log.Fatal("Failed to create card service", zap.Error(err))
	}
	defer cardService.Close()

	// 服务启动时创建所有卡片
	ctx := context.Background()
	if err := cardService.CreateAllCards(ctx); err != nil {
		log.Warn("Failed to create all cards on startup", zap.Error(err))
		// 不退出，继续启动服务
	}

	// 创建 HTTP 处理器和路由
	handler := httphandler.NewHandler(cardService, log)
	router := httphandler.NewRouter(handler)
	router.RegisterRoutes()

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// 启动 HTTP 服务器（在 goroutine 中）
	errChan := make(chan error, 1)
	go func() {
		log.Info("HTTP server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case sig := <-sigChan:
		log.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	case err := <-errChan:
		log.Error("Server error", zap.Error(err))
	}

	// 优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Error shutting down server", zap.Error(err))
	}

	log.Info("Service stopped")
}

