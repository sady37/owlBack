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
	enablementCache := service.NewAlarmEnablementCache(db, metaCache, logger)
	alarmSvc := service.NewAlarmService(writer, reader, db, enablementCache, logger)
	alarmSvc.SetRedisPending(&redisPendingAdapter{client: redisClient})

	// Sync alarm state from DB on startup
	if err := alarmSvc.SyncAllCardsAlarmState(ctx); err != nil {
		logger.Warn("alarm sync failed", zap.Error(err))
	}

	resolver := service.NewDeviceCardResolver(db)
	monitorHandler := consumer.NewMonitorHandler(monitorBuf, writer, metaCache, resolver, logger)
	go monitorHandler.RunLoop(ctx)
	go runDeriveLoop(ctx, monitorBuf, stateSvc, metaCache, reader, alarmSvc, logger)
	eventHandler := consumer.NewEventHandler(stateSvc, alarmSvc, monitorBuf, metaCache, enablementCache, resolver, logger)
	alarmHandler := consumer.NewAlarmHandler(alarmSvc, stateSvc, monitorBuf, metaCache, resolver, logger)
	alarmProcessHandler := consumer.NewAlarmProcessHandler(alarmSvc, logger)
	cardChangeHandler := consumer.NewCardChangeHandler(alarmSvc, stateSvc, metaCache, enablementCache, resolver, logger)
	alarmDeviceHandler := consumer.NewAlarmDeviceHandler(enablementCache, logger)

	consumer.SubscribeAll(ctx, logger, redisClient, consumer.Handlers{
		Monitor:      consumer.NewIotPreparedHandler(resolver, stateSvc, metaCache, monitorHandler),
		Event:        consumer.NewIotPreparedHandler(resolver, stateSvc, metaCache, eventHandler),
		Alarm:        consumer.NewIotPreparedHandler(resolver, stateSvc, metaCache, alarmHandler),
		AlarmProcess: alarmProcessHandler,
		CardChange:   cardChangeHandler,
		AlarmDevice:  alarmDeviceHandler,
	})

	go runPendingAlarmScan(ctx, alarmSvc, logger)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down")
	cancel()
}

func runDeriveLoop(ctx context.Context, buf *service.MonitorBuffer, state *service.StateService, metaCache *service.DeviceMetaCache, reader *card.Reader, alarmSvc *service.AlarmService, logger *zap.Logger) {
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
		nowMs := time.Now().UnixMilli()
		snapshots := buf.Flush(nowMs)
		shouldDerive := service.IsDeriveTick(tick)
		currOnline := make(map[string]bool)
		for _, cid := range buf.ActiveCardIDs() {
			currOnline[cid] = true
		}

		for _, snap := range snapshots {
			if shouldDerive {
				meta := metaCache.GetOrLoad(ctx, snap.CardID)
				prev := prevTargets[snap.CardID]
				status, err := state.DeriveAndWriteState(ctx, snap, meta, prev, buf)
				if err != nil {
					logger.Warn("derive state", zap.String("cid", snap.CardID), zap.Error(err))
				}
				if status != nil && status.Target != nil {
					prevTargets[snap.CardID] = status.Target
				}
				state.DeriveBedStateFromRealtime(ctx, snap, meta)
			}
		}
		if shouldDerive {
			snappedCards := make(map[string]bool, len(snapshots))
			for _, s := range snapshots {
				snappedCards[s.CardID] = true
			}
			activeDevs := buf.ActiveDevicesByCard()
			for cid := range activeDevs {
				if snappedCards[cid] {
					continue
				}
				meta := metaCache.GetOrLoad(ctx, cid)
				_ = state.DeriveDeviceOnlineOnly(ctx, cid, meta, buf)
			}
		}
		for cid := range prevOnline {
			if !currOnline[cid] {
				meta := metaCache.GetOrLoad(ctx, cid)
				if meta != nil && len(meta.Devices) > 0 {
					buf.ClearCard(cid, meta.Devices)
				}
				state.SetCardOffline(ctx, cid)
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

func (a *redisPendingAdapter) HSet(ctx context.Context, key string, values ...interface{}) error {
	return a.client.HSet(ctx, key, values...).Err()
}

func (a *redisPendingAdapter) HDel(ctx context.Context, key string, fields ...string) error {
	return a.client.HDel(ctx, key, fields...).Err()
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
