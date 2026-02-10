package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"owl-common/database"
	logpkg "owl-common/logger"
	"owl-common/redis"

	"wisefido-iot/internal/config"
	"wisefido-iot/internal/consumer"
	"wisefido-iot/internal/publisher"
	"wisefido-iot/internal/repository"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化 Logger
	logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-iot")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting wisefido-iot service",
		zap.String("version", "1.0.0"),
		zap.String("monitor_stream", cfg.Streams.Monitor),
		zap.String("stat_stream", cfg.Streams.Stat),
		zap.String("event_stream", cfg.Streams.Event),
		zap.String("alarm_stream", cfg.Streams.Alarm),
		zap.String("auth_stream", cfg.Streams.Auth),
		zap.String("consumer_group", cfg.ConsumerGroup),
		zap.String("consumer_name", cfg.ConsumerName),
		zap.Int64("batch_size", cfg.BatchSize),
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

	logger.Info("Service stopped")
}

