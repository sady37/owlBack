package service

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-data-transformer/internal/config"
	"wisefido-data-transformer/internal/consumer"
	"wisefido-data-transformer/internal/repository"
	"wisefido-data-transformer/internal/transformer"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	"owl-common/database"
	rediscommon "owl-common/redis"
)

// TransformerService 数据转换服务
type TransformerService struct {
	config           *config.Config
	logger           *zap.Logger
	db               *sql.DB
	redisClient      *redis.Client
	snomedRepo          *repository.SNOMEDRepository
	iotRepo             *repository.IoTTimeSeriesRepository
	radarTransformer    *transformer.RadarTransformer
	sleepaceTransformer *transformer.SleepaceTransformer
	consumer            *consumer.StreamConsumer
}

// NewTransformerService 创建数据转换服务
func NewTransformerService(cfg *config.Config, logger *zap.Logger) (*TransformerService, error) {
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
	snomedRepo := repository.NewSNOMEDRepository(db, logger)
	iotRepo := repository.NewIoTTimeSeriesRepository(db, logger)
	
	// 创建Transformer
	radarTransformer := transformer.NewRadarTransformer(snomedRepo, logger)
	sleepaceTransformer := transformer.NewSleepaceTransformer(snomedRepo, logger)
	
	// 创建Consumer
	streamConsumer := consumer.NewStreamConsumer(
		cfg,
		redisClient,
		snomedRepo,
		iotRepo,
		radarTransformer,
		sleepaceTransformer,
		logger,
	)
	
	return &TransformerService{
		config:           cfg,
		logger:           logger,
		db:               db,
		redisClient:      redisClient,
		snomedRepo:          snomedRepo,
		iotRepo:             iotRepo,
		radarTransformer:    radarTransformer,
		sleepaceTransformer: sleepaceTransformer,
		consumer:            streamConsumer,
	}, nil
}

// Start 启动服务
func (s *TransformerService) Start(ctx context.Context) error {
	s.logger.Info("Starting data transformer service components")
	
	// 启动Stream消费者
	if err := s.consumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start stream consumer: %w", err)
	}
	
	s.logger.Info("Data transformer service started successfully")
	return nil
}

// Stop 停止服务
func (s *TransformerService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping data transformer service")
	
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
	
	s.logger.Info("Data transformer service stopped")
	return nil
}

