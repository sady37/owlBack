package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	rediscommon "owl-common/redis"
	"wisefido-cardagg/internal/consumer"
	"wisefido-cardagg/internal/publisher"
	"wisefido-cardagg/internal/repository"
	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	redisClient := redislib.NewClient(&redislib.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Redis connection failed", zap.Error(err))
	}
	logger.Info("Redis connected")

	// 初始化各层组件
	cacheRepo := repository.NewRedisCache(redisClient)
	monitorSvc := service.NewMonitorService(cacheRepo, logger)
	monitorHandler := consumer.NewMonitorHandler(monitorSvc, logger)

	cardPublisher := publisher.NewCardStreamPublisher(redisClient, logger)
	eventAlarmSvc := service.NewEventAlarmService(cacheRepo, cardPublisher, logger)
	eventAlarmHandler := consumer.NewEventAlarmHandler(eventAlarmSvc, logger)
	alarmProcessHandler := consumer.NewAlarmProcessHandler(eventAlarmSvc, logger)

	streams := []string{
		"iot:monitor:stream",
		"iot:event:stream",
		"iot:alarm:stream",
		rediscommon.StreamConfigAlarmProcess.Name,
	}

	logger.Info("Subscribing to streams", zap.Strings("streams", streams))

	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, redisClient, stream, "cardagg-group"); err != nil {
			logger.Warn("Create consumer group failed", zap.String("stream", stream), zap.Error(err))
		}
	}

	// 启动 monitor 消费 goroutine
	go subscribeStream(ctx, logger, redisClient, "iot:monitor:stream", "cardagg-group", "consumer-monitor", func(c context.Context, msg rediscommon.StreamMessage) error {
		return monitorHandler.Handle(c, msg)
	})

	// 启动 event 消费 goroutine
	go subscribeStream(ctx, logger, redisClient, "iot:event:stream", "cardagg-group", "consumer-event", func(c context.Context, msg rediscommon.StreamMessage) error {
		return eventAlarmHandler.Handle(c, msg)
	})

	// 启动 alarm 消费 goroutine
	go subscribeStream(ctx, logger, redisClient, "iot:alarm:stream", "cardagg-group", "consumer-alarm", func(c context.Context, msg rediscommon.StreamMessage) error {
		return eventAlarmHandler.Handle(c, msg)
	})

	// 启动 config:alarmProcess 消费 goroutine（处理 alarmProcess 消息）
	go subscribeStream(ctx, logger, redisClient, rediscommon.StreamConfigAlarmProcess.Name, "cardagg-group", "consumer-config-alarm", func(c context.Context, msg rediscommon.StreamMessage) error {
		return alarmProcessHandler.Handle(c, msg)
	})

	// 启动定期清理过期数据的 goroutine，每5分钟运行一次
	go startPeriodicCleanup(ctx, logger, monitorSvc, cacheRepo)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down")
	cancel()
}

// startPeriodicCleanup 启动定期清理过期数据的任务
func startPeriodicCleanup(ctx context.Context, logger *zap.Logger, monitorSvc *service.MonitorService, cacheRepo repository.CacheRepository) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger.Info("Starting periodic cleanup job", zap.Duration("interval", 5*time.Minute))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping periodic cleanup job")
			return
		case <-ticker.C:
			logger.Info("Running periodic cleanup...")

			// 获取所有卡片的 IDs 并清理过期数据
			cardIDs, err := cacheRepo.GetAllCardIds(ctx)
			if err != nil {
				logger.Warn("Failed to get card IDs for cleanup", zap.Error(err))
				continue
			}

			cleanedCount := 0
			for _, cardID := range cardIDs {
				if err := monitorSvc.CleanupExpired(ctx, cardID); err != nil {
					logger.Warn("Failed to cleanup expired data for card",
						zap.String("card_id", cardID),
						zap.Error(err))
				} else {
					cleanedCount++
				}
			}

			logger.Info("Periodic cleanup completed",
				zap.Int("cards_processed", len(cardIDs)),
				zap.Int("cards_cleaned", cleanedCount))
		}
	}
}

func subscribeStream(ctx context.Context, logger *zap.Logger, redisClient *redislib.Client, stream string, group string, consumer string, handler func(context.Context, rediscommon.StreamMessage) error) {
	logger.Info("Starting subscriber", zap.String("stream", stream), zap.String("consumer", consumer))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, redisClient, stream, group, consumer, 10, 2*time.Second)
		if err != nil {
			logger.Warn("Read stream failed", zap.String("stream", stream), zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		for _, m := range msgs {
			if err := handler(ctx, m); err != nil {
				logger.Warn("Handler error", zap.String("stream", stream), zap.Error(err))
			}

			if err := redisClient.XAck(ctx, stream, group, m.ID).Err(); err != nil {
				logger.Warn("ACK failed", zap.String("stream", stream), zap.String("id", m.ID), zap.Error(err))
			}
		}
	}
}
