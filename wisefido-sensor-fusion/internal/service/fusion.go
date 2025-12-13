package service

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-sensor-fusion/internal/config"
	"wisefido-sensor-fusion/internal/consumer"
	"wisefido-sensor-fusion/internal/fusion"
	"wisefido-sensor-fusion/internal/repository"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	"owl-common/database"
	rediscommon "owl-common/redis"
)

// FusionService 传感器融合服务
type FusionService struct {
	config      *config.Config
	logger      *zap.Logger
	db          *sql.DB
	redisClient *redis.Client
	consumer    *consumer.StreamConsumer
}

// NewFusionService 创建传感器融合服务
func NewFusionService(cfg *config.Config, logger *zap.Logger) (*FusionService, error) {
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
	
	// 创建Repository
	cardRepo := repository.NewCardRepository(db, logger)
	iotRepo := repository.NewIoTTimeSeriesRepository(db, logger)
	
	// 创建Fusion
	sensorFusion := fusion.NewSensorFusion(cardRepo, iotRepo, logger)
	
	// 创建CacheManager
	cacheManager := consumer.NewCacheManager(cfg, redisClient, logger)
	
	// 创建Consumer
	streamConsumer := consumer.NewStreamConsumer(
		cfg,
		redisClient,
		cardRepo,
		iotRepo,
		sensorFusion,
		cacheManager,
		logger,
	)
	
	return &FusionService{
		config:      cfg,
		logger:      logger,
		db:          db,
		redisClient: redisClient,
		consumer:    streamConsumer,
	}, nil
}

// Start 启动服务
func (s *FusionService) Start(ctx context.Context) error {
	s.logger.Info("Starting sensor fusion service components")
	
	// 启动Stream消费者
	if err := s.consumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start stream consumer: %w", err)
	}
	
	s.logger.Info("Sensor fusion service started successfully")
	return nil
}

// Stop 停止服务
func (s *FusionService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping sensor fusion service")
	
	// 关闭Redis
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			s.logger.Error("Error closing Redis client", zap.Error(err))
		}
	}
	
	// 关闭数据库
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.Error("Error closing database connection", zap.Error(err))
		}
	}
	
	s.logger.Info("Sensor fusion service stopped")
	return nil
}

