package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 加载配置
	var cfg *config.Config
	var err error

	// 优先从环境变量加载配置（与现有环境保持一致）
	if os.Getenv("DB_HOST") != "" || os.Getenv("REDIS_ADDR") != "" || os.Getenv("MQTT_BROKER") != "" {
		log.Println("Loading configuration from environment variables...")
		cfg, err = config.LoadFromEnv()
	} else {
		log.Println("Loading configuration from config file...")
		cfg, err = config.Load()
	}

	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化Redis客户端
	log.Println("Initializing Redis client...")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试Redis连接
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connected successfully")

	// 初始化MQTT客户端
	log.Println("Initializing MQTT client...")
	mqttClient, err := mqtt.NewClient(&cfg.MQTT)
	if err != nil {
		log.Fatalf("Failed to create MQTT client: %v", err)
	}
	defer mqttClient.Disconnect()

	// 初始化数据库连接
	log.Println("Initializing database connection...")
	db, err := database.NewPostgresDB(&cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// 创建Repository
	log.Println("Creating repositories...")
	deviceRepo := repository.NewPostgresDeviceRepository(db)

	// 创建 logger（时间格式 hh:mm:ss.ms，与 cardagg 一致）
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
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// 创建Redis Stream发布器
	log.Println("Creating Redis Stream publisher...")
	streamPublisher := consumer.NewStreamPublisher(redisClient, cfg)

	// 创建卡片映射服务
	log.Println("Creating card mapping service...")
	dataAPIURL := cfg.DataAPIURL
	if dataAPIURL == "" {
		dataAPIURL = "http://127.0.0.1:8080"
	}
	cardAPIClient := card.NewCardAPIClient(dataAPIURL)
	cardMappingSvc := service.NewCardMappingService(cardAPIClient, logger)

	// 设置 cardMappingSvc 到 streamPublisher
	streamPublisher.SetCardMappingService(cardMappingSvc)
	streamPublisher.SetLogger(logger)

	// 先创建MQTT消费者（不启动，用于获取messageHandler）
	log.Println("Creating MQTT consumer...")
	mqttConsumer, err := consumer.NewMQTTConsumer(
		cfg,
		mqttClient,
		redisClient,
		deviceRepo,
		cardMappingSvc, // 卡片映射服务
		streamPublisher,
		nil, // 先传nil，后面再设置subscriptionManager
	)
	if err != nil {
		log.Fatalf("Failed to create MQTT consumer: %v", err)
	}

	// 创建设备订阅管理器（传入messageHandler）
	log.Println("Creating device subscription manager...")
	subscriptionManager := subscriber.NewDeviceSubscriptionManager(
		cfg,
		mqttClient,
		db,
		logger,
		mqttConsumer.GetMessageHandler(), // 传入MQTT consumer的messageHandler
	)

	// 设置 Stream 发布器（用于发布设备在线状态到 alarm stream）
	subscriptionManager.SetStreamPublisher(streamPublisher)
	// 设置 MQTT 消费者（用于订阅/取消订阅设备主题，认证后订阅 6 个主题收 monitor/stat/event 等）
	subscriptionManager.SetMQTTConsumer(mqttConsumer)

	// 设置subscriptionManager到mqttConsumer（用于UpdateLastSeen）
	mqttConsumer.SetSubscriptionManager(subscriptionManager)

	configSub := subscriber.NewConfigSubscriber(redisClient, cfg, logger, cardMappingSvc)
	if err := configSub.Start(ctx); err != nil {
		logger.Warn("config subscriber start", zap.Error(err))
	}
	go runConfigCardStreamReader(ctx, redisClient, logger, configSub)
	// 主动探测请求流（前端 refresh 触发）：device_type=Radar 时立刻跑 health_check
	go runProbeDeviceStreamReader(ctx, redisClient, logger, subscriptionManager)

	// 启动时初始化缓存
	log.Println("Initializing device and card mapping caches at startup...")

	// 1. 初始化 device_store 缓存（deviceUID → allow_access）
	log.Println("Loading device_store data into cache...")
	allDevices, err := deviceRepo.GetAllDeviceStoreInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to load device_store data: %v", err)
	}
	deviceCacheCount := 0
	for _, device := range allDevices {
		// 仅缓存允许访问的设备 UID（白名单），避免因为未命中缓存而拒绝认证
		if device.Access {
			domain.AllowAccessCache.Store(device.DeviceUID, true)
			deviceCacheCount++
		}
	}
	log.Printf("✅ Loaded %d device records into cache", deviceCacheCount)

	// cardMapping 使用懒加载：首次 MQTT 消息到达时自动查 DB 并缓存
	log.Printf("✅ Card mapping service ready (lazy-load mode)")

	// 创建服务
	log.Println("Creating radar service...")
	radarService, err := service.NewRadarService(cfg, mqttClient, redisClient, deviceRepo, streamPublisher, mqttConsumer)
	if err != nil {
		log.Fatalf("Failed to create radar service: %v", err)
	}

	// 启动服务
	log.Println("Starting services...")
	if err := radarService.Start(ctx); err != nil {
		log.Fatalf("Failed to start radar service: %v", err)
	}

	// 设置 RadarService 到订阅管理器（用于健康检查时读取设备属性）
	subscriptionManager.SetRadarService(radarService)
	// 反向注入：任何 GetDeviceProperties 成功（install/HTTP/排错）触发"正向恢复"，
	// 几秒内同步 Offline/SignalPoor/AngleAbnormal=0，避免等下一个 10min 健康检查 tick。
	radarService.SetHealthRefresher(subscriptionManager)

	// 启动设备订阅管理器
	log.Println("Starting device subscription manager...")
	if err := subscriptionManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start device subscription manager: %v", err)
	}
	defer subscriptionManager.Stop(ctx)

	// 创建HTTP服务器（实际 Start 延后到 OTA handler 注入之后，否则路由在注册时丢失 OTA）
	log.Println("Creating HTTP server (internal control/query)...")
	httpServer := http.NewServer(&cfg.HTTP, radarService, cfg, db, deviceRepo, redisClient, logger, subscriptionManager, cardMappingSvc)

	// Set DB on MQTT consumer for OTA return handling
	mqttConsumer.SetDB(db)

	// 启动HTTPS服务器（用于设备认证，必须配置证书）
	var httpsServer *http.HTTPSServer
	var otaScheduler *ota.Scheduler
	if cfg.HTTPS.Port > 0 {
		log.Println("Starting HTTPS server (device authentication)...")
		httpsServer, err = http.NewHTTPSServer(&cfg.HTTPS, cfg, db, deviceRepo, redisClient, logger, subscriptionManager, cardMappingSvc)
		if err != nil {
			log.Fatalf("Failed to create HTTPS server: %v (HTTPS server requires TLS certificates)", err)
		}

		// Set TCP OnProgress callback
		httpsServer.TCPServer().OnProgress = makeOTAProgressCallback(db)

		// Set TCP OnRegister callback: writeback firmware + mark online
		httpsServer.TCPServer().OnRegister = func(uid, deviceType, sfVer, hwVer string) {
			log.Printf("[TCP-Register] writeback: uid=%s type=%s sfVer=%s hwVer=%s", uid, deviceType, sfVer, hwVer)
			// Update firmware_version + mcu_model in device_store
			_, err := db.ExecContext(context.Background(), `
				UPDATE device_store SET
					firmware_version = COALESCE(NULLIF($2, ''), firmware_version),
					mcu_model = COALESCE(NULLIF($3, ''), mcu_model)
				WHERE device_uid = $1
			`, uid, sfVer, hwVer)
			if err != nil {
				log.Printf("[TCP-Register] writeback failed uid=%s: %v", uid, err)
			}
			// Mark online in subscription manager (same mechanism as MQTT devices)
			var deviceID string
			_ = db.QueryRowContext(context.Background(), `SELECT device_id FROM device_store WHERE device_uid = $1`, uid).Scan(&deviceID)
			subscriptionManager.SetTCPDeviceOnline(uid, deviceID)
		}

		// Set TCP disconnect callback: mark offline in subscription manager
		httpsServer.TCPServer().Sessions.OnDisconnect = func(uid string) {
			log.Printf("[TCP-Disconnect] setting offline: uid=%s", uid)
			subscriptionManager.SetTCPDeviceOffline(uid)
		}

		// Create OTA handler and inject to HTTP server
		otaHandler := http.NewOTAHandler(httpsServer.OTAManager(), httpsServer.TCPServer())
		httpServer.SetOTAHandler(otaHandler)

		// Create MQTT publisher for OTA push via MQTT
		mqttPublisher := publisher.NewMQTTPublisher(cfg, mqttClient)

		// Inject MQTT publisher into OTA handler for device commands (restart/wifi/iotserver)
		otaHandler.SetCommander(mqttPublisher)
		otaHandler.SetMQTTOTA(mqttPublisher)

		// Create OTA scheduler with MQTT OTA push callback
		fwDir := filepath.Join("..", "ota")
		serverAddr := strings.TrimSpace(cfg.MQTT.RadarDeviceMQTT.Server)
		if serverAddr == "" {
			serverAddr = "0.0.0.0"
		}
		// Firmware download via nginx 443 (Let's Encrypt cert, ESP32 compatible)
		fwURL := fmt.Sprintf("https://%s/ota", serverAddr)
		otaScheduler = ota.NewScheduler(
			db,
			func(uid string, data map[string]interface{}) error {
				return mqttPublisher.PublishOTA(context.Background(), uid, data)
			},
			httpsServer.OTAManager().PushToDevice, // TCP push for old MCU
			1*time.Minute, // 测试期间 1 分钟扫描一次，正式可回 5m
			fwDir,
			fwURL,
		)
		otaScheduler.Start()

		go func() {
			if err := httpsServer.Start(); err != nil {
				log.Fatalf("HTTPS server error: %v", err)
			}
		}()

	}

	// OTA handler 已注入（若启用 HTTPS），此时再启动 HTTP server 以确保 OTA 路由被注册
	log.Println("Starting HTTP server (internal control/query)...")
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Printf("wisefido-qinglan service started successfully")
	log.Printf("MQTT connected to: %s", mqtt.EffectiveBrokerDialString(&cfg.MQTT))
	log.Printf("HTTP server (internal) listening on: %s", cfg.HTTP.GetAddr())
	if httpsServer != nil {
		log.Printf("HTTPS server (auth) listening on: :%d", cfg.HTTPS.Port)
	}
	log.Printf("Redis connected to: %s", cfg.Redis.Addr)
	log.Printf("Database connected to: %s:%d/%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)

	// 输出 Redis Stream 配置信息
	log.Printf("Redis Streams:")
	log.Printf("  - Monitor stream: %s", rediscommon.StreamMonitor.Name)
	log.Printf("  - Stat stream: %s", rediscommon.StreamStat.Name)
	log.Printf("  - Event stream: %s", rediscommon.StreamEvent.Name)
	log.Printf("  - Auth stream: %s", rediscommon.StreamAuth.Name)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v, shutting down...", sig)

	// 优雅关闭
	cancel()

	// Stop OTA scheduler
	if otaScheduler != nil {
		otaScheduler.Stop()
	}

	// 关闭HTTP和HTTPS服务器
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	httpServer.Shutdown(ctxShutdown)
	if httpsServer != nil {
		httpsServer.Stop(ctxShutdown)
	}

	// 停止服务
	if err := radarService.Stop(ctx); err != nil {
		log.Printf("Error during service shutdown: %v", err)
	}

	log.Println("Service stopped gracefully")
}

// makeOTAProgressCallback creates a TCP OTA progress callback that updates device_store
func makeOTAProgressCallback(db *sql.DB) tcp.OTAProgressCallback {
	return func(uid string, progress int, message string) {
		log.Printf("[OTA-Progress] uid=%s progress=%d msg=%s", uid, progress, message)

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
			log.Printf("[OTA-Progress] failed to update ota_status for uid=%s: %v", uid, err)
		}
	}
}

// runProbeDeviceStreamReader 订阅 iot:probe:device:stream，对 device_type=Radar 的请求
// 立刻调用 ProbeDevice → checkDeviceHealth，缩短前端 refresh 到状态更新的延迟。
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

