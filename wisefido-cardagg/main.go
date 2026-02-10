package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"owl-common/card"
	rediscommon "owl-common/redis"
	"wisefido-cardagg/internal/config"
	"wisefido-cardagg/internal/consumer"
	"wisefido-cardagg/internal/publisher"
	"wisefido-cardagg/internal/repository"
	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger, err := zap.NewProduction()
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
	monitorSvc := service.NewMonitorService(cacheRepo, logger)
	monitorHandler := consumer.NewMonitorHandler(monitorSvc, logger)

	cardPublisher := publisher.NewCardStreamPublisher(redisClient, logger)
	eventAlarmSvc := service.NewEventAlarmService(cacheRepo, cardPublisher, db, logger)
	eventAlarmHandler := consumer.NewEventAlarmHandler(eventAlarmSvc, logger)
	alarmProcessHandler := consumer.NewAlarmProcessHandler(eventAlarmSvc, logger)

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

	// 启动 deviceStatus 消费 goroutine（处理设备在线/离线等状态变化）
	go subscribeStream(ctx, logger, redisClient, rediscommon.StreamIoTDeviceStatus.Name, "cardagg-group", "consumer-device-status", func(c context.Context, msg rediscommon.StreamMessage) error {
		return eventAlarmHandler.Handle(c, msg)
	})

	// 启动 config:alarmProcess 消费 goroutine（处理 alarmProcess 消息）
	go subscribeStream(ctx, logger, redisClient, rediscommon.StreamConfigAlarmProcess.Name, "cardagg-group", "consumer-config-alarm", func(c context.Context, msg rediscommon.StreamMessage) error {
		return alarmProcessHandler.Handle(c, msg)
	})

	// 从数据库初始化报警计数到 Redis（可选，需要数据库连接）
	if db != nil {
		if err := initializeAlarmCounts(ctx, db, cacheRepo, logger); err != nil {
			logger.Warn("Failed to initialize alarm counts from database", zap.Error(err))
			// 继续运行，计数可能不准确但不阻塞启动
		}
	}

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

// initializeAlarmCounts 从数据库读取卡片的报警计数，初始化到 Redis
// 查询 cards 表的 unhandled_alarm_0~4 字段，构建 ActiveAlarmState 写入 Redis
func initializeAlarmCounts(ctx context.Context, db *sql.DB, cacheRepo repository.CacheRepository, logger *zap.Logger) error {
	logger.Info("Starting to initialize alarm counts from database")

	// 查询所有卡片及其报警计数
	query := `
		SELECT card_id, unhandled_alarm_0, unhandled_alarm_1, unhandled_alarm_2, unhandled_alarm_3, unhandled_alarm_4
		FROM cards
		WHERE unhandled_alarm_0 > 0 OR unhandled_alarm_1 > 0 OR unhandled_alarm_2 > 0 OR unhandled_alarm_3 > 0 OR unhandled_alarm_4 > 0
		LIMIT 1000
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var cardID string
		var count0, count1, count2, count3, count4 int
		if err := rows.Scan(&cardID, &count0, &count1, &count2, &count3, &count4); err != nil {
			logger.Warn("Failed to scan card alarm count", zap.Error(err))
			continue
		}

		// 构建 ActiveAlarmState
		   activeAlarms := &card.ActiveAlarmState{
			   ActiveEmerg:   count0,
			   ActiveAlert:   count1,
			   ActiveCrit:    count2,
			   ActiveErr:     count3,
			   ActiveWarning: count4,
			   // NOTICE 字段从数据库中没有对应列，保持默认值 0
		   }

		// 从现有 CardStatus 读取，仅更新 ActiveAlarms 部分
		cardStatus, err := cacheRepo.GetCardStatus(ctx, cardID)
		if err != nil || cardStatus == nil {
			// 如果不存在，创建新的
			cardStatus = &card.CardStatus{
				CardID: cardID,
			}
		}

		// 仅在有报警时更新，设置时间戳（当前时间）
		if count0 > 0 || count1 > 0 || count2 > 0 || count3 > 0 || count4 > 0 {
			activeAlarms.Timestamp = time.Now().Unix() * 1000 // 毫秒时间戳
			cardStatus.ActiveAlarms = activeAlarms

			// 写入 Redis（TTL 12H）
			if err := cacheRepo.SetCardStatus(ctx, cardStatus); err != nil {
				logger.Warn("Failed to set card status for alarm initialization",
					zap.String("card_id", cardID),
					zap.Error(err))
				continue
			}

			count++
		}
	}

	if err := rows.Err(); err != nil {
		logger.Warn("Error reading rows", zap.Error(err))
	}

	logger.Info("Alarm count initialization completed", zap.Int("cards_initialized", count))
	return nil
}
