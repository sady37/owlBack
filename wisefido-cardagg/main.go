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
	rediscommon "owl-common/redis"
	"wisefido-cardagg/internal/config"
	"wisefido-cardagg/internal/consumer"
	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	if cfg.Logging.Level != "" {
		var lvl zapcore.Level
		if err := lvl.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(cfg.Logging.Level)))); err == nil {
			zapCfg.Level = zap.NewAtomicLevelAt(lvl)
		}
	}
	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting wisefido-cardagg",
		zap.String("redis", cfg.Redis.Addr),
		zap.String("db_host", cfg.DB.Host))

	// Redis
	redisClient := redislib.NewClient(&redislib.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		pingCancel()
		logger.Fatal("redis ping failed", zap.Error(err))
	}
	pingCancel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// PostgreSQL
	var db *sql.DB
	if dbURL := cfg.GetDatabaseURL(); dbURL != "" {
		if d, err := sql.Open("postgres", dbURL); err == nil {
			db = d
			defer db.Close()
			dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := db.PingContext(dbCtx); err != nil {
				logger.Warn("db ping failed", zap.Error(err))
				db = nil
			}
			dbCancel()
		}
	}

	// card Writer / Reader
	writer := card.NewWriter(redisClient, rediscommon.StreamCardStatus.MaxLen, rediscommon.StreamCardRealTime.MaxLen)
	reader := card.NewReader(redisClient)

	// Services
	stateSvc := service.NewStateService(writer, reader, logger)
	monitorBuf := service.NewMonitorBuffer()
	metaCache := service.NewDeviceMetaCache(db, logger)
	if err := metaCache.BuildDeviceIndex(ctx); err != nil {
		logger.Warn("device index initial build failed", zap.Error(err))
	}
	enablementCache := service.NewAlarmEnablementCache(db, metaCache, logger)
	alarmSvc := service.NewAlarmService(writer, reader, db, enablementCache, logger)
	alarmSvc.SetRedisPending(&redisPendingAdapter{client: redisClient})

	// Sync alarm state from DB on startup
	if err := alarmSvc.SyncAllCardsAlarmState(ctx); err != nil {
		logger.Warn("alarm sync failed", zap.Error(err))
	}
	// Sync device:status hash alarm flags from alarm_events.active (信号差/倾角/传感器脱落 3 个 flag）
	// 启动 bootstrap：cardagg 重启 + qinglan/sleepace transition dedup 已 cache=1 不再 republish onset 时，
	// 按 DB 真值自动重建 hash 的 3 个 alarm flag。
	if err := alarmSvc.SyncDeviceStatusFromActiveAlarms(ctx); err != nil {
		logger.Warn("device:status flag sync failed", zap.Error(err))
	}

	bedCoord := service.NewBedEventCoordinator()

	// PR6: AI override cache（wisefido-sensor 通过 iot:event:stream + category=track_verdict 喂入）
	// sandbox 默认开，release 模式才把 AI confidence 合并到前端 track 字段。
	// 任何模式都不影响 alarm 路径（"宁可误报不可漏报"）。
	aiOverrides := service.NewAIOverrideCache(cfg.AIOverride.Mode, cfg.AIOverride.TTLSec, logger)
	logger.Info("ai_override cache initialized",
		zap.String("mode", string(aiOverrides.Mode())),
		zap.Int("ttl_sec", cfg.AIOverride.TTLSec),
		zap.Int("gc_sec", cfg.AIOverride.GCSec))
	logger.Info("ai_override runtime switch enabled",
		zap.String("sandbox_signal", "SIGUSR1"),
		zap.String("release_signal", "SIGUSR2"))
	go aiOverrides.RunGCLoop(ctx.Done(), time.Duration(cfg.AIOverride.GCSec)*time.Second)

	monitorHandler := consumer.NewMonitorHandler(monitorBuf, writer, metaCache, bedCoord, stateSvc, alarmSvc, aiOverrides, logger)
	go monitorHandler.RunLoop(ctx)
	go runDeriveLoop(ctx, monitorBuf, stateSvc, metaCache, reader, alarmSvc, bedCoord, logger)
	eventHandler := consumer.NewEventHandler(stateSvc, alarmSvc, monitorBuf, metaCache, enablementCache, bedCoord, aiOverrides, logger)
	alarmHandler := consumer.NewAlarmHandler(alarmSvc, stateSvc, monitorBuf, metaCache, bedCoord, logger)
	alarmProcessHandler := consumer.NewAlarmProcessHandler(alarmSvc, logger)
	cardChangeHandler := consumer.NewCardChangeHandler(alarmSvc, stateSvc, metaCache, enablementCache, bedCoord, db, logger)
	alarmDeviceHandler := consumer.NewAlarmDeviceHandler(enablementCache, alarmSvc, logger)

	consumer.SubscribeAll(ctx, logger, redisClient, consumer.Handlers{
		Monitor:      consumer.NewIotPreparedHandler(stateSvc, metaCache, monitorHandler, logger),
		Event:        consumer.NewIotPreparedHandler(stateSvc, metaCache, eventHandler, logger),
		Alarm:        consumer.NewIotPreparedHandler(stateSvc, metaCache, alarmHandler, logger),
		AlarmProcess: alarmProcessHandler,
		CardChange:   cardChangeHandler,
		AlarmDevice:  alarmDeviceHandler,
	})

	// PR1 (A10): 订阅 ai:track:verdict:stream（sensor 派生 ghost 判定专用流）。
	// 此前 sensor 把 track_verdict 写到 iot:event:stream，cardagg event_handler 兼带消费；
	// 现切到独立流以便 B 组迁移 cardagg 完全退订 iot:event:stream。
	consumer.NewAIVerdictHandler(redisClient, aiOverrides, logger).Start(ctx)

	go runPendingAlarmScan(ctx, alarmSvc, logger)
	go runNightAbsenceCheck(ctx, alarmSvc, metaCache, logger)

	// device:status fail-safe：alarm/event/monitor 三流断时（gateway 故障），把 last_seen_ms
	// 长期未刷新的设备主动 patch offline=1。180s 阈值 = 2× sleepace OfflineRecover 80s 周期。
	deviceWatchdog := service.NewDeviceWatchdog(redisClient, writer, 180*time.Second, 30*time.Second, logger)
	go deviceWatchdog.Run(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		sig := <-sigCh
		switch sig {
		case syscall.SIGUSR1:
			aiOverrides.SetMode(string(service.AIOverrideModeSandbox))
			logger.Info("ai_override mode switched", zap.String("mode", string(aiOverrides.Mode())))
			continue
		case syscall.SIGUSR2:
			aiOverrides.SetMode(string(service.AIOverrideModeRelease))
			logger.Info("ai_override mode switched", zap.String("mode", string(aiOverrides.Mode())))
			continue
		case syscall.SIGINT, syscall.SIGTERM:
			// graceful shutdown
		}
		break
	}

	logger.Info("shutting down")
	cancel()
}

func runDeriveLoop(ctx context.Context, buf *service.MonitorBuffer, state *service.StateService, metaCache *service.DeviceMetaCache, reader *card.Reader, alarmSvc *service.AlarmService, bedCoord *service.BedEventCoordinator, logger *zap.Logger) {
	_ = reader
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	prevOnline := make(map[string]bool)
	prevTargets := make(map[string]*card.TargetState)
	tick := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		tick++
		bedCoord.Tick(ctx, state, alarmSvc, metaCache, buf, logger)
		nowMs := time.Now().UnixMilli()
		if tick%90 == 0 {
			buf.PruneStaleDevices(nowMs, 90_000)
		}
		snapshots := buf.Flush(nowMs)
		shouldDerive := service.IsDeriveTick(tick)
		currOnline := make(map[string]bool)
		for _, cid := range buf.ActiveCardIDs() {
			currOnline[cid] = true
		}
		for _, snap := range snapshots {
			meta := metaCache.GetOrLoad(ctx, snap.CardID)
			if shouldDerive {
				prev := prevTargets[snap.CardID]
				status, err := state.DeriveAndWriteState(ctx, snap, meta, prev, buf)
				if err != nil {
					logger.Warn("derive state", zap.String("cid", snap.CardID), zap.Error(err))
				}
				if status != nil && status.Target != nil {
					prevTargets[snap.CardID] = status.Target
				}
			}
			state.DeriveBedStateFromRealtime(ctx, snap, meta)
		}
		// card 从 buffer 退场时清 prevTargets，防止 map 长期增长。
		// device:status offline 由 watchdog/alarm Offline 维护，本循环不再插手。
		for cid := range prevOnline {
			if !currOnline[cid] {
				delete(prevTargets, cid)
				logger.Info("card offline", zap.String("cid", cid))
			}
		}
		prevOnline = currOnline
	}
}

type redisPendingAdapter struct {
	client *redislib.Client
}

func (a *redisPendingAdapter) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return a.client.HGetAll(ctx, key).Result()
}

func (a *redisPendingAdapter) HGet(ctx context.Context, key, field string) (string, error) {
	return a.client.HGet(ctx, key, field).Result()
}

func (a *redisPendingAdapter) HSet(ctx context.Context, key string, values ...interface{}) error {
	return a.client.HSet(ctx, key, values...).Err()
}

func (a *redisPendingAdapter) HDel(ctx context.Context, key string, fields ...string) error {
	return a.client.HDel(ctx, key, fields...).Err()
}

func runNightAbsenceCheck(ctx context.Context, alarmSvc *service.AlarmService, metaCache *service.DeviceMetaCache, logger *zap.Logger) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			alarmSvc.CheckNightAbsence(ctx, metaCache)
		}
	}
}

func runPendingAlarmScan(ctx context.Context, alarmSvc *service.AlarmService, logger *zap.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := alarmSvc.ScanPendingAlarms(ctx); err != nil {
				logger.Warn("scan pending alarms", zap.Error(err))
			}
		}
	}
}
