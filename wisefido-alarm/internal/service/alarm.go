package service

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/consumer"
	"wisefido-alarm/internal/evaluator"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

// AlarmService 报警服务（整合各层）
type AlarmService struct {
	config      *config.Config
	db          *sql.DB
	redisClient *redis.Client
	logger      *zap.Logger
	tenantID    string

	// 各层组件
	cacheManager    *consumer.CacheManager
	stateManager    *consumer.StateManager
	cacheConsumer   *consumer.CacheConsumer
	cardRepo        *repository.CardRepository
	deviceRepo      *repository.DeviceRepository
	roomRepo        *repository.RoomRepository
	alarmCloudRepo  *repository.AlarmCloudRepository
	alarmDeviceRepo *repository.AlarmDeviceRepository
	alarmEventsRepo *repository.AlarmEventsRepository
	evaluator       *evaluator.Evaluator
}

// NewAlarmService 创建报警服务
func NewAlarmService(cfg *config.Config, logger *zap.Logger, tenantID string) (*AlarmService, error) {
	// 1. 连接数据库
	db, err := sql.Open("postgres", buildDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 2. 连接 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试 Redis 连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	// 3. 创建 Repository 层
	cardRepo := repository.NewCardRepository(db, logger)
	deviceRepo := repository.NewDeviceRepository(db, logger)
	roomRepo := repository.NewRoomRepository(db, logger)
	alarmCloudRepo := repository.NewAlarmCloudRepository(db, logger)
	alarmDeviceRepo := repository.NewAlarmDeviceRepository(db, logger)
	alarmEventsRepo := repository.NewAlarmEventsRepository(db, logger)

	// 4. 创建 Consumer 层
	cacheManager := consumer.NewCacheManager(cfg, redisClient, logger)
	stateManager := consumer.NewStateManager(cfg, redisClient, logger)

	// 5. 创建 Evaluator 层
	eval := evaluator.NewEvaluator(
		cfg,
		stateManager,
		cardRepo,
		deviceRepo,
		roomRepo,
		alarmCloudRepo,
		alarmDeviceRepo,
		alarmEventsRepo,
		logger,
	)

	// 6. 创建 CacheConsumer
	cacheConsumer := consumer.NewCacheConsumer(
		cfg,
		cacheManager,
		cardRepo,
		logger,
		tenantID,
	)

	return &AlarmService{
		config:          cfg,
		db:              db,
		redisClient:     redisClient,
		logger:          logger,
		tenantID:        tenantID,
		cacheManager:    cacheManager,
		stateManager:    stateManager,
		cacheConsumer:   cacheConsumer,
		cardRepo:        cardRepo,
		deviceRepo:      deviceRepo,
		roomRepo:        roomRepo,
		alarmCloudRepo:  alarmCloudRepo,
		alarmDeviceRepo: alarmDeviceRepo,
		alarmEventsRepo: alarmEventsRepo,
		evaluator:       eval,
	}, nil
}

// Start 启动服务
func (s *AlarmService) Start(ctx context.Context) error {
	s.logger.Info("Starting alarm service",
		zap.String("tenant_id", s.tenantID),
	)

	// 启动 CacheConsumer（轮询模式）
	if err := s.cacheConsumer.Start(ctx, s.evaluator); err != nil {
		return fmt.Errorf("failed to start cache consumer: %w", err)
	}

	return nil
}

// Stop 停止服务
func (s *AlarmService) Stop() error {
	s.logger.Info("Stopping alarm service")

	// 关闭数据库连接
	if err := s.db.Close(); err != nil {
		s.logger.Error("Failed to close database",
			zap.Error(err),
		)
	}

	// 关闭 Redis 连接
	if err := s.redisClient.Close(); err != nil {
		s.logger.Error("Failed to close redis",
			zap.Error(err),
		)
	}

	return nil
}

// buildDSN 构建数据库连接字符串
func buildDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)
}

