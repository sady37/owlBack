// wisefido-cardagg 薄 adapter：可信源 alarm routing + card:status 投影 + device:status 维护。
//
// 职责（详 owlBack/CLAUDE.md 规则 #1.3 / 规则 #2）：
//   1. iot:alarm:stream → alarm_events PG INSERT + card:status.AlarmState
//   2. iot:monitor:stream → MonitorBuffer 累积 + 1s snapshot 发 card:realtime:stream + device 活跃触刷
//   3. sensor:zone:state:stream → card:status.BedState/RoomState 投影（unit /80 不在此层）
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
	reader := card.NewReader(redisClient)

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

	unitPicker := consumer.NewUnitPicker(db, redisClient, writer, reader, logger)
	alarmRouter := consumer.NewAlarmRouter(db, writer, reader, enablementCache, metaCache, deviceTracker, unitPicker, logger)
	// TargetMerger：per-device → owning card max-merge（LastActive / Standing / WeakBio）。
	// 详 [[target_state_per_device]]。VisitorDeriver 通过 ApplyVisitor 注入 visitor 三字段。
	// 注入 deviceTracker.IsOnline 作为 LastActive merge 的 offline 过滤器。
	targetMerger := service.NewTargetMerger(metaCache)
	targetMerger.SetOnlineChecker(deviceTracker.IsOnline)
	// BedPeopleTracker：firmware NumberPeople event → per /96 bed-bound radar device snapshot。
	// VisitorDeriver bed-level path 读它判定 share/private bed 卡 visitor（doc §4.4）。
	bedPeopleTracker := service.NewBedPeopleTracker(metaCache)
	bedPeopleTracker.SetOnlineChecker(deviceTracker.IsOnline)
	eventHandler := consumer.NewEventHandler(bedPeopleTracker, logger)
	// AIOverrideCache (S5a): sensor 派生 track verdict (ghost 判定等) 缓存 + 合并到
	// realtime publish。mode 按部署环境配置（[[deploy_mode_by_domain]]）：
	//   test.wisefido.com  → sandbox（仅 log，不动 track_confidence）
	//   owl.wisefido.com   → release（覆写 track_confidence + ai_source）
	// 配置优先级：env CARDAGG_AI_OVERRIDE_MODE > config.yaml ai.override_mode > 默认 sandbox
	// 清理触发：a) tid=88 (monitor) b) EnterRoom/ExitRoom (event) c) TTL GC d) (后续) device offline。
	aiOverrides := service.NewAIOverrideCache(cfg.AI.OverrideMode, cfg.AI.OverrideTTLSec, logger)
	monitorHandler.SetAIOverrides(aiOverrides)
	eventHandler.SetAIOverrides(aiOverrides)
	go aiOverrides.RunGCLoop(ctx.Done(), 30*time.Second)
	aiVerdictHandler := consumer.NewAIVerdictHandler(redisClient, aiOverrides, logger)
	aiVerdictHandler.Start(ctx)
	sensorStateProjector := consumer.NewSensorStateProjector(writer, reader, unitPicker, logger)
	sensorStateProjector.SetTargetMerger(targetMerger)
	cardLifecycle := consumer.NewCardLifecycle(db, writer, metaCache, enablementCache, unitPicker, logger)
	cardLifecycle.SetTargetMerger(targetMerger) // 卡变更 / 设备解绑时清 device snapshot
	alarmDeviceHandler := consumer.NewAlarmDeviceHandler(enablementCache, logger)
	alarmProcessHandler := consumer.NewAlarmProcessHandler(db, writer, logger)

	consumer.SubscribeAll(ctx, logger, redisClient, consumer.Handlers{
		Monitor:      monitorHandler,
		Event:        eventHandler,
		Alarm:        alarmRouter,
		SensorState:  sensorStateProjector,
		CardChange:   cardLifecycle,
		AlarmDevice:  alarmDeviceHandler,
		AlarmProcess: alarmProcessHandler,
	})

	// VisitorDeriver 60s tick：双路径（bed-bound radar bed cards 优先 / Private /88 room cards 兜底）。
	// 详 doc/card_display.md §4.4 + [[visitor_belongs_to_cardagg]]。
	visitorDeriver := consumer.NewVisitorDeriver(metaCache, reader, targetMerger, bedPeopleTracker, 0, logger)
	go visitorDeriver.Run(ctx)

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
