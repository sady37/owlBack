package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	rediscommon "owl-common/redis"
	"wisefido-cardagg/internal/config"
	"wisefido-cardagg/internal/consumer"
	"wisefido-cardagg/internal/publisher"
	"wisefido-cardagg/internal/repository"
	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 加载配置
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting wisefido-cardagg service",
		zap.String("redis_addr", cfg.Redis.Addr),
		zap.String("db_host", cfg.DB.Host),
		zap.Int("db_port", cfg.DB.Port))

	// 连接 Redis
	redisClient := redislib.NewClient(&redislib.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 用于短期操作（连接测试）的临时 context
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		pingCancel()
		logger.Fatal("Redis connection failed", zap.Error(err))
	}
	pingCancel()
	logger.Info("Redis connected")

	// 用于长期运行（流订阅）的 context，只在收到关闭信号时才取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 连接 PostgreSQL（可选）
	var db *sql.DB
	dbURL := cfg.GetDatabaseURL()
	if dbURL != "" {
		if d, err := sql.Open("postgres", dbURL); err == nil {
			db = d
			defer db.Close()
			dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := db.PingContext(dbCtx); err == nil {
				logger.Info("PostgreSQL connected",
					zap.String("host", cfg.DB.Host),
					zap.Int("port", cfg.DB.Port),
					zap.String("database", cfg.DB.Database))
			} else {
				logger.Warn("PostgreSQL connection failed", zap.Error(err))
				db = nil
			}
			dbCancel()
		} else {
			logger.Warn("Failed to open PostgreSQL connection", zap.Error(err))
		}
	}

	// 初始化各层组件
	cacheRepo := repository.NewRedisCache(redisClient)
	cardPublisher := publisher.NewCardStreamPublisher(redisClient, logger)
	monitorSvc := service.NewMonitorService(cacheRepo, cardPublisher, logger)
	monitorHandler := consumer.NewMonitorHandler(monitorSvc, logger)
	eventAlarmSvc := service.NewEventAlarmService(cacheRepo, cardPublisher, db, logger)
	eventAlarmHandler := consumer.NewEventAlarmHandler(eventAlarmSvc, logger)
	alarmProcessHandler := consumer.NewAlarmProcessHandler(eventAlarmSvc, logger)

	// 启动时从 DB 同步所有 card 的 alarm_state 到 Redis
	if err := eventAlarmSvc.SyncAllCardsAlarmState(ctx); err != nil {
		logger.Warn("Failed to sync cards alarm state on startup", zap.Error(err))
	}

	streams := []string{
		"iot:monitor:stream",
		"iot:event:stream",
		"iot:alarm:stream",
		rediscommon.StreamIoTDeviceStatus.Name,
		rediscommon.StreamConfigAlarmProcess.Name,
	}

	logger.Info("Subscribing to streams", zap.Strings("streams", streams))

	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, redisClient, stream, "cardagg-group"); err != nil {
			logger.Warn("Create consumer group failed", zap.String("stream", stream), zap.Error(err))
		}
	}

	// 启动 monitor 消费 goroutine
	// 注意：msg.Values 全是 string（Redis Stream 特性），ConvertRedisValues 做 string→native 转换
	go subscribeStream(ctx, logger, redisClient, "iot:monitor:stream", "cardagg-group", "consumer-monitor", func(c context.Context, msg rediscommon.StreamMessage) error {
		return monitorHandler.Handle(c, consumer.ConvertRedisValues(msg.Values))
	})

	// 启动 event 消费 goroutine
	go subscribeStream(ctx, logger, redisClient, "iot:event:stream", "cardagg-group", "consumer-event", func(c context.Context, msg rediscommon.StreamMessage) error {
		return eventAlarmHandler.Handle(c, consumer.ConvertRedisValues(msg.Values))
	})

	// 启动 alarm 消费 goroutine
	go subscribeStream(ctx, logger, redisClient, "iot:alarm:stream", "cardagg-group", "consumer-alarm", func(c context.Context, msg rediscommon.StreamMessage) error {
		return eventAlarmHandler.Handle(c, consumer.ConvertRedisValues(msg.Values))
	})

	// 启动 deviceStatus 消费 goroutine（处理设备在线/离线等状态变化）
	go subscribeStream(ctx, logger, redisClient, rediscommon.StreamIoTDeviceStatus.Name, "cardagg-group", "consumer-device-status", func(c context.Context, msg rediscommon.StreamMessage) error {
		return eventAlarmHandler.Handle(c, consumer.ConvertRedisValues(msg.Values))
	})

	// 启动 config:alarmProcess 消费 goroutine（处理 alarmProcess 消息）
	go subscribeStream(ctx, logger, redisClient, rediscommon.StreamConfigAlarmProcess.Name, "cardagg-group", "consumer-config-alarm", func(c context.Context, msg rediscommon.StreamMessage) error {
		return alarmProcessHandler.Handle(c, consumer.ConvertRedisValues(msg.Values))
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down")
	cancel()
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

