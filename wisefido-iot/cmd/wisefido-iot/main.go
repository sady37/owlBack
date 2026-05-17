// wisefido-iot — Redis Streams → PG v2 持久化边界（CLAUDE.md 规则 #2）。
//
//   iot:monitor:stream → monitor_stream
//   iot:event:stream   → event_log
//   (alarm 由 cardagg alarm_router 单 writer 写 alarm_events，本服务不消费)

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
	"wisefido-iot/internal/repository"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-iot")
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting wisefido-iot",
		zap.String("monitor_stream", cfg.Streams.Monitor),
		zap.String("event_stream", cfg.Streams.Event),
		zap.String("consumer_group", cfg.ConsumerGroup),
		zap.String("consumer_name", cfg.ConsumerName),
		zap.Int64("batch_size", cfg.BatchSize),
	)

	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("connect db", zap.Error(err))
	}
	defer database.Close(db)

	redisClient := redis.NewRedisClient(&cfg.Redis)
	defer redis.Close(redisClient)
	if err := redis.Ping(context.Background(), redisClient); err != nil {
		logger.Fatal("ping redis", zap.Error(err))
	}

	streamRepo := repository.NewStreamRepo(db, logger)
	streamConsumer := consumer.NewStreamConsumer(cfg, redisClient, streamRepo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := streamConsumer.Start(ctx); err != nil {
			logger.Fatal("stream consumer exited", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Info("shutting down", zap.String("signal", sig.String()))
	cancel()
}
