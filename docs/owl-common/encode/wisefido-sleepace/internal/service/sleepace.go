package service

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-sleepace/internal/config"
	"wisefido-sleepace/internal/consumer"
	"wisefido-sleepace/internal/repository"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	"owl-common/database"
	rediscommon "owl-common/redis"
	mqttcommon "owl-common/mqtt"
)

// SleepaceService Sleepace 服务
type SleepaceService struct {
	config     *config.Config
	logger     *zap.Logger
	db         *sql.DB
	redis      *redis.Client
	mqttClient *mqttcommon.Client
	consumer   *consumer.MQTTConsumer
}

// NewSleepaceService 创建 Sleepace 服务
func NewSleepaceService(cfg *config.Config, logger *zap.Logger) (*SleepaceService, error) {
	// 初始化数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 初始化Redis
	redisClient := rediscommon.NewRedisClient(&cfg.Redis)
	if err := rediscommon.Ping(context.Background(), redisClient); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	
	// 初始化MQTT
	mqttClient, err := mqttcommon.NewClient(&cfg.MQTT, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MQTT: %w", err)
	}
	
	// 创建Repository
	deviceRepo := repository.NewDeviceRepository(db, logger)
	
	// 创建Consumer
	mqttConsumer := consumer.NewMQTTConsumer(cfg, mqttClient, redisClient, deviceRepo, logger)
	
	return &SleepaceService{
		config:     cfg,
		logger:     logger,
		db:         db,
		redis:      redisClient,
		mqttClient: mqttClient,
		consumer:   mqttConsumer,
	}, nil
}

// Start 启动服务
func (s *SleepaceService) Start(ctx context.Context) error {
	s.logger.Info("Starting sleepace service components")
	
	// 启动MQTT消费者
	if err := s.consumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MQTT consumer: %w", err)
	}
	
	s.logger.Info("Sleepace service started successfully")
	return nil
}

// Stop 停止服务
func (s *SleepaceService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping sleepace service")
	
	// 停止Consumer
	if s.consumer != nil {
		if err := s.consumer.Stop(ctx); err != nil {
			s.logger.Error("Error stopping consumer", zap.Error(err))
		}
	}
	
	// 断开MQTT
	if s.mqttClient != nil {
		s.mqttClient.Disconnect()
	}
	
	// 关闭Redis
	if s.redis != nil {
		rediscommon.Close(s.redis)
	}
	
	// 关闭数据库
	if s.db != nil {
		database.Close(s.db)
	}
	
	s.logger.Info("Sleepace service stopped")
	return nil
}

