// wisefido-cardagg 薄 adapter：可信源 alarm routing + card:status 投影 + device:status 维护。
//
// 职责（详 owlBack/CLAUDE.md 规则 #1.3 / 规则 #2）：
//   1. iot:alarm:stream → alarm_events PG INSERT + card:status.AlarmState
//   2. iot:monitor:stream → MonitorBuffer 累积 + 1s snapshot 发 card:realtime:stream + device 活跃触刷
//   3. sensor:zone:state:stream → card:status.BedState/RoomState/BathRoomState 投影
//   4. config:card / config:alarmDevice / config:alarmProcess → cache invalidate + AlarmState 刷新
//   5. 180s 看门狗 fail-safe device:status offline
//
// 砍掉（已归 sensor 或彻底退役）：bedCoord / Stay FSM / AHI / NightAbsence ticker / Pending 计时 /
// AI override cache (移到 data SSE) / runDeriveLoop / Target 派生 (未来 LogicID 实现时再加).

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"owl-common/card"
	owllogger "owl-common/logger"
	owlredis "owl-common/redis"

	"wisefido-cardagg/internal/config"
	"wisefido-cardagg/internal/consumer"
	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := mustLogger(cfg.Logging.Level)
	defer logger.Sync()
	logger.Info("starting wisefido-cardagg",
		zap.String("redis", cfg.Redis.Addr),
		zap.String("db_host", cfg.DB.Host))

	redisClient := redislib.NewClient(&redislib.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		pingCancel()
		logger.Fatal("redis ping", zap.Error(err))
	}
	pingCancel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sql.Open("postgres", cfg.GetDatabaseURL())
	if err != nil {
		logger.Fatal("open db", zap.Error(err))
	}
	defer db.Close()
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(dbCtx); err != nil {
		dbCancel()
		logger.Fatal("db ping", zap.Error(err))
	}
	dbCancel()

	writer := card.NewWriter(redisClient, owlredis.StreamCardStatus.MaxLen, owlredis.StreamCardRealTime.MaxLen)

	monitorBuf := service.NewMonitorBuffer()
	metaCache := service.NewDeviceMetaCache(db, logger)
	if err := metaCache.BuildDeviceIndex(ctx); err != nil {
		logger.Warn("device index build", zap.Error(err))
	}
	enablementCache := service.NewAlarmEnablementCache(db, metaCache, logger)

	deviceTracker := consumer.NewDeviceStatusTracker(writer, logger)
	go deviceTracker.Run(ctx)

	monitorHandler := consumer.NewMonitorHandler(monitorBuf, writer, deviceTracker, logger)
	go monitorHandler.RunLoop(ctx)

	alarmRouter := consumer.NewAlarmRouter(db, writer, enablementCache, metaCache, deviceTracker, logger)
	sensorStateProjector := consumer.NewSensorStateProjector(writer, logger)
	cardLifecycle := consumer.NewCardLifecycle(db, writer, metaCache, enablementCache, logger)
	alarmDeviceHandler := consumer.NewAlarmDeviceHandler(enablementCache, logger)
	alarmProcessHandler := consumer.NewAlarmProcessHandler(db, writer, logger)

	consumer.SubscribeAll(ctx, logger, redisClient, consumer.Handlers{
		Monitor:      monitorHandler,
		Alarm:        alarmRouter,
		SensorState:  sensorStateProjector,
		CardChange:   cardLifecycle,
		AlarmDevice:  alarmDeviceHandler,
		AlarmProcess: alarmProcessHandler,
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info("shutting down")
	cancel()
	time.Sleep(500 * time.Millisecond)
}

func mustLogger(level string) *zap.Logger {
	l, err := owllogger.NewLogger(strings.ToLower(strings.TrimSpace(level)), "json", "wisefido-cardagg")
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	return l
}
