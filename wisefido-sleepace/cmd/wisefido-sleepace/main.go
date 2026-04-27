package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"owl-common/card"
	"owl-common/database"
	owllogger "owl-common/logger"
	rediscommon "owl-common/redis"

	"wisefido-sleepace/internal/config"
	"wisefido-sleepace/internal/consumer"
	httpapi "wisefido-sleepace/internal/http"
	"wisefido-sleepace/internal/service"
	"wisefido-sleepace/internal/subscriber"
)

func main() {
	env := flag.String("env", "dev", "Environment (dev, prod, test)")
	flag.Parse()

	cfgPath := fmt.Sprintf("sleepace-%s.yaml", *env)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := owllogger.NewLogger(cfg.Log.Level, "console", "wisefido-sleepace")
	if err != nil {
		log.Fatalf("create logger: %v", err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("redis connect failed", zap.Error(err))
	}
	defer redisClient.Close()
	logger.Info("redis connected", zap.String("addr", cfg.Redis.Addr))

	// PostgreSQL
	db, err := database.NewPostgresDB(&cfg.DB)
	if err != nil {
		logger.Fatal("database connect failed", zap.Error(err))
	}
	defer database.Close(db)
	logger.Info("database connected", zap.String("host", cfg.DB.Host))

	// Card mapping: MQTT deviceId (device_code) → device_uid via API
	dataAPIURL := cfg.DataAPIURL
	if dataAPIURL == "" {
		dataAPIURL = "http://127.0.0.1:8080"
	}
	cardAPIClient := card.NewCardAPIClient(dataAPIURL)
	cardMapping := service.NewCardMappingService(cardAPIClient, logger)
	if err := cardMapping.LoadCodeToUIDMap(ctx); err != nil {
		logger.Warn("load device_code→device_uid map", zap.Error(err))
	}

	// CardDB kept only for SetDeviceVersionSync
	cardDB := card.NewCardDB(db)

	// Sleepace API (HTTP proxy to sleepace-service Java)
	sleepaceAPI := service.NewSleepaceAPI(&cfg.Sleepace, logger)

	// Report service
	reportSvc := service.NewReportService(db, sleepaceAPI, cardMapping, logger)

	// Device status tracker (in-memory, updated by MQTT connectionStatus)
	statusTracker := service.NewDeviceStatusTracker(cardMapping, logger)

	// Stream publisher
	streamPub := consumer.NewStreamPublisher(redisClient, cfg, logger)
	streamPub.SetCardMappingService(cardMapping)

	healthCheck := subscriber.NewHealthCheck(cardMapping, sleepaceAPI, streamPub, statusTracker, logger)
	healthCheck.CardDB = cardDB // probe 时顺手 sync device_store.firmware_version
	go healthCheck.Run(ctx)

	const sleepaceCardGroup = "wisefido-sleepace-card"
	if err := rediscommon.CreateConsumerGroup(ctx, redisClient, rediscommon.StreamConfigCard.Name, sleepaceCardGroup); err != nil {
		logger.Warn("create consumer group for config card stream", zap.Error(err))
	}
	cfgSub := subscriber.NewConfigSubscriber(redisClient, cardMapping, healthCheck, logger)
	go subscriber.SubscribeLoop(ctx, logger, redisClient, rediscommon.StreamConfigCard.Name, sleepaceCardGroup, "sleepace-1", cfgSub)

	// MQTT consumer（首次上连时同步 device_store 版本：cardDB + sleepaceAPI）
	mqttConsumer := consumer.NewMQTTConsumer(streamPub, reportSvc, statusTracker, logger)
	mqttConsumer.SetDeviceVersionSync(cardDB, sleepaceAPI)
	mqttConsumer.Start(ctx, 10)
	defer mqttConsumer.Stop()

	// MQTT client
	mqttClient, err := connectMQTT(cfg, mqttConsumer.MessageHandler(), logger)
	if err != nil {
		logger.Fatal("mqtt connect failed", zap.Error(err))
	}
	defer mqttClient.Disconnect(250)
	logger.Info("mqtt connected", zap.String("broker", cfg.MQTT.Broker), zap.String("topic", cfg.MQTT.TopicID))

	// HTTP server (proxy + status APIs)
	router := httpapi.NewRouter(&cfg.Sleepace, sleepaceAPI, cardMapping, statusTracker, logger)
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	if cfg.HTTP.Port == 0 {
		httpAddr = ":8081"
	}
	httpServer := &http.Server{Addr: httpAddr, Handler: router}
	go func() {
		logger.Info("http server starting", zap.String("addr", httpAddr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	logger.Info("wisefido-sleepace 2.0 started",
		zap.String("http", httpAddr),
		zap.String("mqtt_topic", cfg.MQTT.TopicID),
		zap.String("streams", fmt.Sprintf("monitor=%s, event=%s, alarm=%s",
			rediscommon.StreamMonitor.Name, rediscommon.StreamEvent.Name, rediscommon.StreamAlarm.Name)))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info("shutting down...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	httpServer.Shutdown(shutdownCtx)
}

func connectMQTT(cfg *config.Config, handler mqtt.MessageHandler, logger *zap.Logger) (mqtt.Client, error) {
	topic := cfg.MQTT.TopicID

	opts := mqtt.NewClientOptions().
		SetClientID(cfg.MQTT.ClientID).
		SetUsername(cfg.MQTT.Username).
		SetPassword(cfg.MQTT.Password).
		AddBroker(cfg.MQTT.Broker).
		SetAutoReconnect(true).
		SetCleanSession(true).
		SetMaxReconnectInterval(2 * time.Minute).
		SetKeepAlive(30 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			logger.Info("mqtt connected, subscribing", zap.String("topic", topic))
			if token := c.Subscribe(topic, 1, handler); token.Wait() && token.Error() != nil {
				logger.Error("mqtt subscribe failed", zap.Error(token.Error()))
			}
		}).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			logger.Warn("mqtt disconnected", zap.Error(err))
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}
