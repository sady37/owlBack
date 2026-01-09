package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	logpkg "owl-common/logger"
	"owl-common/database"
	"owl-common/redis"

	"go.uber.org/zap"
	"wisefido-iot-timeseries/internal/config"
	"wisefido-iot-timeseries/internal/consumer"
	httphandler "wisefido-iot-timeseries/internal/http"
	"wisefido-iot-timeseries/internal/publisher"
	"wisefido-iot-timeseries/internal/repository"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化 Logger
	logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-iot-timeseries")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting wisefido-iot-timeseries service",
		zap.String("version", "1.0.0"),
		zap.String("monitor_stream", cfg.Streams.Monitor),
		zap.String("stat_stream", cfg.Streams.Stat),
		zap.String("event_stream", cfg.Streams.Event),
		zap.String("alarm_stream", cfg.Streams.Alarm),
		zap.String("output_stream", cfg.Streams.Output),
	)

	// 初始化数据库连接
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)

	logger.Info("Database connection established")

	// 初始化 Redis 连接
	redisClient := redis.NewRedisClient(&cfg.Redis)
	defer redis.Close(redisClient)

	// 测试 Redis 连接
	if err := redis.Ping(context.Background(), redisClient); err != nil {
		logger.Fatal("Failed to ping Redis", zap.Error(err))
	}

	logger.Info("Redis connection established")

	// 初始化 Repository
	iotRepo := repository.NewIoTTimeSeriesRepository(db, logger)

	// 初始化 Publisher
	streamPublisher := publisher.NewStreamPublisher(redisClient, logger)

	// 初始化 Consumer
	streamConsumer := consumer.NewStreamConsumer(
		cfg,
		redisClient,
		iotRepo,
		streamPublisher,
		logger,
	)

	// 初始化 HTTP Handler（用于内部 API，如清除位置信息缓存）
	httpHandler := httphandler.NewHandler(iotRepo, logger)

	// 启动 HTTP 服务器（用于内部 API）
	httpAddr := getEnv("HTTP_ADDR", ":8083")
	mux := http.NewServeMux()
	mux.HandleFunc("/internal/api/v1/iot-timeseries/cache/invalidate", httpHandler.InvalidateLocationCache)
	
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	go func() {
		logger.Info("Starting HTTP server for internal API",
			zap.String("addr", httpAddr),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 在 goroutine 中启动 Stream Consumer
	go func() {
		if err := streamConsumer.Start(ctx); err != nil {
			logger.Fatal("Failed to start stream consumer", zap.Error(err))
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))

	// 优雅关闭
	cancel()

	// 关闭 HTTP 服务器
	if err := httpServer.Shutdown(context.Background()); err != nil {
		logger.Warn("Failed to shutdown HTTP server", zap.Error(err))
	}

	logger.Info("Service stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
