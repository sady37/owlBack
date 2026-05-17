package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/consumer"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/http"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/ota"
	"wisefido-qinglan/internal/publisher"
	"wisefido-qinglan/internal/repository"
	"wisefido-qinglan/internal/service"
	"wisefido-qinglan/internal/subscriber"
	"wisefido-qinglan/internal/tcp"

	"owl-common/card"
	"owl-common/database"
	logpkg "owl-common/logger"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	var cfg *config.Config
	var err error

	if os.Getenv("DB_HOST") != "" || os.Getenv("REDIS_ADDR") != "" || os.Getenv("MQTT_BROKER") != "" {
		cfg, err = config.LoadFromEnv()
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := logpkg.NewLogger(cfg.Logging.Level, cfg.Logging.Format, "wisefido-qinglan")
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("config loaded")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("connect redis", zap.Error(err))
	}
	logger.Info("redis connected", zap.String("addr", cfg.Redis.Addr))

	mqttClient, err := mqtt.NewClient(&cfg.MQTT)
	if err != nil {
		logger.Fatal("create mqtt client", zap.Error(err))
	}
	defer mqttClient.Disconnect()

	db, err := database.NewPostgresDB(&cfg.DB)
	if err != nil {
		logger.Fatal("connect db", zap.Error(err))
	}
	defer database.Close(db)
	logger.Info("db connected",
		zap.String("host", cfg.DB.Host),
		zap.Int("port", cfg.DB.Port),
		zap.String("db", cfg.DB.Database),
	)

	deviceRepo := repository.NewPostgresDeviceRepository(db)
	streamPublisher := consumer.NewStreamPublisher(redisClient, cfg)

	dataAPIURL := cfg.DataAPIURL
	if dataAPIURL == "" {
		dataAPIURL = "http://127.0.0.1:8080"
	}
	cardAPIClient := card.NewCardAPIClient(dataAPIURL)
	cardMappingSvc := service.NewCardMappingService(cardAPIClient, logger)

	streamPublisher.SetCardMappingService(cardMappingSvc)
	streamPublisher.SetLogger(logger)

	mqttConsumer, err := consumer.NewMQTTConsumer(
		cfg,
		mqttClient,
		redisClient,
		deviceRepo,
		cardMappingSvc,
		streamPublisher,
		nil,
		logger,
	)
	if err != nil {
		logger.Fatal("create mqtt consumer", zap.Error(err))
	}

	subscriptionManager := subscriber.NewDeviceSubscriptionManager(
		cfg,
		mqttClient,
		db,
		logger,
		mqttConsumer.GetMessageHandler(),
	)
	subscriptionManager.SetStreamPublisher(streamPublisher)
	subscriptionManager.SetMQTTConsumer(mqttConsumer)
	mqttConsumer.SetSubscriptionManager(subscriptionManager)

	configSub := subscriber.NewConfigSubscriber(redisClient, cfg, logger, cardMappingSvc)
	if err := configSub.Start(ctx); err != nil {
		logger.Warn("config subscriber start", zap.Error(err))
	}
	go runConfigCardStreamReader(ctx, redisClient, logger, configSub)
	go runProbeDeviceStreamReader(ctx, redisClient, logger, subscriptionManager)

	allDevices, err := deviceRepo.GetAllDeviceStoreInfo(ctx)
	if err != nil {
		logger.Fatal("load device_store", zap.Error(err))
	}
	deviceCacheCount := 0
	for _, device := range allDevices {
		if device.Access {
			domain.AllowAccessCache.Store(device.DeviceUID, true)
			deviceCacheCount++
		}
	}
	logger.Info("device cache primed", zap.Int("count", deviceCacheCount))

	radarService, err := service.NewRadarService(cfg, mqttClient, redisClient, deviceRepo, streamPublisher, mqttConsumer, logger)
	if err != nil {
		logger.Fatal("create radar service", zap.Error(err))
	}
	if err := radarService.Start(ctx); err != nil {
		logger.Fatal("start radar service", zap.Error(err))
	}

	subscriptionManager.SetRadarService(radarService)
	radarService.SetHealthRefresher(subscriptionManager)

	if err := subscriptionManager.Start(ctx); err != nil {
		logger.Fatal("start subscription manager", zap.Error(err))
	}
	defer subscriptionManager.Stop(ctx)

	httpServer := http.NewServer(&cfg.HTTP, radarService, cfg, db, deviceRepo, redisClient, logger, subscriptionManager, cardMappingSvc)
	mqttConsumer.SetDB(db)

	var httpsServer *http.HTTPSServer
	var otaScheduler *ota.Scheduler
	if cfg.HTTPS.Port > 0 {
		httpsServer, err = http.NewHTTPSServer(&cfg.HTTPS, cfg, db, deviceRepo, redisClient, logger, subscriptionManager, cardMappingSvc)
		if err != nil {
			logger.Fatal("create https server (need TLS certs)", zap.Error(err))
		}

		httpsServer.TCPServer().SetLogger(logger)
		httpsServer.TCPServer().OnProgress = makeOTAProgressCallback(db, logger)

		httpsServer.TCPServer().OnRegister = func(uid, deviceType, sfVer, hwVer string) {
			logger.Info("tcp register writeback",
				zap.String("uid", uid),
				zap.String("device_type", deviceType),
				zap.String("sf_ver", sfVer),
				zap.String("hw_ver", hwVer),
			)
			_, err := db.ExecContext(context.Background(), `
				UPDATE device_store SET
					firmware_version = COALESCE(NULLIF($2, ''), firmware_version),
					mcu_model = COALESCE(NULLIF($3, ''), mcu_model)
				WHERE device_uid = $1
			`, uid, sfVer, hwVer)
			if err != nil {
				logger.Warn("tcp register writeback failed", zap.String("uid", uid), zap.Error(err))
			}
			var deviceID string
			_ = db.QueryRowContext(context.Background(), `SELECT device_id FROM device_store WHERE device_uid = $1`, uid).Scan(&deviceID)
			subscriptionManager.SetTCPDeviceOnline(uid, deviceID)
		}

		httpsServer.TCPServer().Sessions.OnDisconnect = func(uid string) {
			logger.Info("tcp disconnect", zap.String("uid", uid))
			subscriptionManager.SetTCPDeviceOffline(uid)
		}

		otaHandler := http.NewOTAHandler(httpsServer.OTAManager(), httpsServer.TCPServer())
		otaHandler.SetLogger(logger)
		httpServer.SetOTAHandler(otaHandler)

		mqttPublisher := publisher.NewMQTTPublisher(cfg, mqttClient, logger)
		otaHandler.SetCommander(mqttPublisher)
		otaHandler.SetMQTTOTA(mqttPublisher)

		fwDir := filepath.Join("..", "ota")
		serverAddr := strings.TrimSpace(cfg.MQTT.RadarDeviceMQTT.Server)
		if serverAddr == "" {
			serverAddr = "0.0.0.0"
		}
		fwURL := fmt.Sprintf("https://%s/ota", serverAddr)
		otaScheduler = ota.NewScheduler(
			db,
			func(uid string, data map[string]interface{}) error {
				return mqttPublisher.PublishOTA(context.Background(), uid, data)
			},
			httpsServer.OTAManager().PushToDevice,
			1*time.Minute,
			fwDir,
			fwURL,
			logger,
		)
		otaScheduler.Start()

		go func() {
			if err := httpsServer.Start(); err != nil {
				logger.Fatal("https server", zap.Error(err))
			}
		}()
	}

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Warn("http server", zap.Error(err))
		}
	}()

	logger.Info("wisefido-qinglan started",
		zap.String("mqtt", mqtt.EffectiveBrokerDialString(&cfg.MQTT)),
		zap.String("http", cfg.HTTP.GetAddr()),
		zap.Int("https", cfg.HTTPS.Port),
		zap.String("redis", cfg.Redis.Addr),
		zap.String("monitor_stream", rediscommon.StreamMonitor.Name),
		zap.String("event_stream", rediscommon.StreamEvent.Name),
		zap.String("auth_stream", rediscommon.StreamAuth.Name),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("shutting down", zap.String("signal", sig.String()))

	cancel()
	if otaScheduler != nil {
		otaScheduler.Stop()
	}

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	httpServer.Shutdown(ctxShutdown)
	if httpsServer != nil {
		httpsServer.Stop(ctxShutdown)
	}

	if err := radarService.Stop(ctx); err != nil {
		logger.Warn("radar service shutdown", zap.Error(err))
	}
	logger.Info("service stopped")
}

func makeOTAProgressCallback(db *sql.DB, logger *zap.Logger) tcp.OTAProgressCallback {
	return func(uid string, progress int, message string) {
		logger.Info("ota progress",
			zap.String("uid", uid),
			zap.Int("progress", progress),
			zap.String("msg", message),
		)

		var otaStatus string
		switch {
		case progress == -1:
			otaStatus = "failed"
		case progress == 100:
			otaStatus = "success"
		case progress == 0:
			otaStatus = "accepted"
		default:
			otaStatus = fmt.Sprintf("downloading_%d", progress)
		}

		_, err := db.ExecContext(context.Background(), `
			UPDATE device_store SET ota_status = $1, ota_progress = $2, ota_updated_at = NOW()
			WHERE device_uid = $3
		`, otaStatus, progress, uid)
		if err != nil {
			logger.Warn("ota progress update failed", zap.String("uid", uid), zap.Error(err))
		}
	}
}

func runProbeDeviceStreamReader(ctx context.Context, redisClient *redis.Client, logger *zap.Logger, mgr *subscriber.DeviceSubscriptionManager) {
	stream := rediscommon.StreamProbeDevice.Name
	lastID := "$"
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		res, err := redisClient.XRead(ctx, &redis.XReadArgs{
			Streams: []string{stream, lastID},
			Block:   5 * time.Second,
			Count:   16,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			logger.Warn("probe device stream XRead", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, st := range res {
			for _, msg := range st.Messages {
				lastID = msg.ID
				deviceType := streamFieldStr(msg.Values, "device_type")
				if !strings.EqualFold(deviceType, "Radar") {
					continue
				}
				deviceUID := streamFieldStr(msg.Values, "device_uid")
				deviceID := streamFieldStr(msg.Values, "device_id")
				tenantID := streamFieldStr(msg.Values, "tenant_id")
				if deviceUID == "" || deviceID == "" {
					continue
				}
				logger.Info("probe radar requested",
					zap.String("device_uid", deviceUID),
					zap.String("device_id", deviceID),
					zap.String("source", streamFieldStr(msg.Values, "trigger_source")),
				)
				go mgr.ProbeDevice(ctx, deviceUID, deviceID, deviceType, tenantID)
			}
		}
	}
}

func streamFieldStr(values map[string]interface{}, key string) string {
	if v, ok := values[key]; ok && v != nil {
		switch s := v.(type) {
		case string:
			return s
		case []byte:
			return string(s)
		default:
			return fmt.Sprintf("%v", s)
		}
	}
	return ""
}

func runConfigCardStreamReader(ctx context.Context, redisClient *redis.Client, logger *zap.Logger, sub *subscriber.ConfigSubscriber) {
	stream := rediscommon.StreamConfigCard.Name
	lastID := "$"
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		res, err := redisClient.XRead(ctx, &redis.XReadArgs{
			Streams: []string{stream, lastID},
			Block:   5 * time.Second,
			Count:   32,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			logger.Warn("config card stream XRead", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, st := range res {
			for _, msg := range st.Messages {
				if err := sub.HandleConfigChangeMessage(ctx, st.Stream, msg); err != nil {
					logger.Warn("HandleConfigChangeMessage", zap.Error(err))
				}
				lastID = msg.ID
			}
		}
	}
}
