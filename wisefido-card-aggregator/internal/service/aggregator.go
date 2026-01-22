package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"wisefido-card-aggregator/internal/aggregator"
	"wisefido-card-aggregator/internal/alarm"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/consumer"
	"wisefido-card-aggregator/internal/fusion"
	"wisefido-card-aggregator/internal/repository"

	"owl-common/database"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// AggregatorService 卡片聚合服务
type AggregatorService struct {
	config            *config.Config
	logger            *zap.Logger
	db                *sql.DB
	redisClient       *redis.Client
	cardRepo          *repository.CardRepository
	dataAggregator    *aggregator.DataAggregator
	cacheManager      *aggregator.CacheManager
	iotStreamConsumer *consumer.IoTStreamConsumer // IoT Stream 消费者（事件驱动）
	configConsumer    *consumer.ConfigConsumer     // 配置变更消费者
}

// NewAggregatorService 创建卡片聚合服务
func NewAggregatorService(cfg *config.Config, logger *zap.Logger) (*AggregatorService, error) {
	// 初始化数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 初始化 Redis（用于事件驱动模式和数据聚合）
	redisClient := rediscommon.NewRedisClient(&cfg.Redis)
	if err := rediscommon.Ping(context.Background(), redisClient); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// 创建 Repository
	cardRepo := repository.NewCardRepository(db, logger)

	// 创建数据聚合器和缓存管理器（如果启用数据聚合）
	var dataAggregator *aggregator.DataAggregator
	var cacheManager *aggregator.CacheManager
	if cfg.Aggregator.Aggregation.Enabled {
		kv := aggregator.NewRedisKVStore(redisClient)
		cacheManager = aggregator.NewCacheManager(cfg, kv, logger)
		dataAggregator = aggregator.NewDataAggregator(cfg, kv, cardRepo, logger)
	}

	// 创建必要的 repositories（用于 IoT Stream Consumer 和 Config Consumer）
	iotRepo := repository.NewIoTTimeSeriesRepository(db, logger)
	alarmEventsRepo := repository.NewAlarmEventsRepository(db, logger)
	alarmDeviceRepo := repository.NewAlarmDeviceRepository(db, logger)

	// 创建 IoT Stream Consumer（如果启用）
	var iotStreamConsumer *consumer.IoTStreamConsumer
	if cfg.Aggregator.IoTStream.Enabled {

		// 创建 SensorFusion
		sensorFusion := fusion.NewSensorFusion(cardRepo, iotRepo, logger)

		// 创建 AlarmHandler
		alarmHandler := alarm.NewAlarmHandler(alarmEventsRepo, alarmDeviceRepo, cardRepo, logger)

		// 创建 IoTStreamConsumer（需要 cacheManager）
		if cacheManager == nil {
			logger.Warn("IoT Stream Consumer requires cache manager, but aggregation is disabled. IoT Stream Consumer will not be created.")
		} else {
			iotStreamConsumer = consumer.NewIoTStreamConsumer(
				cfg,
				redisClient,
				cardRepo,
				iotRepo,
				sensorFusion,
				cacheManager,
				alarmEventsRepo,
				alarmDeviceRepo,
				alarmHandler,
				logger,
			)
		}
	}

	// 创建配置变更消费者（订阅 config:change:stream）
	// 注意：alarmDeviceRepo 已在第64行创建，直接复用
	var configConsumer *consumer.ConfigConsumer
	configConsumer = consumer.NewConfigConsumer(
		cfg,
		redisClient,
		alarmDeviceRepo,
		logger,
	)

	return &AggregatorService{
		config:            cfg,
		logger:            logger,
		db:                db,
		redisClient:       redisClient,
		cardRepo:          cardRepo,
		dataAggregator:    dataAggregator,
		cacheManager:      cacheManager,
		iotStreamConsumer: iotStreamConsumer,
		configConsumer:    configConsumer,
	}, nil
}

// Start 启动服务
func (s *AggregatorService) Start(ctx context.Context) error {
	s.logger.Info("Starting card aggregator service",
		zap.Bool("aggregation_enabled", s.config.Aggregator.Aggregation.Enabled),
		zap.Bool("iot_stream_enabled", s.config.Aggregator.IoTStream.Enabled),
	)

	// 启动数据聚合任务（如果启用）
	if s.config.Aggregator.Aggregation.Enabled {
		go s.startDataAggregation(ctx)
	}

	// 启动 IoT Stream Consumer（如果启用，事件驱动）
	if s.config.Aggregator.IoTStream.Enabled && s.iotStreamConsumer != nil {
		go func() {
			if err := s.iotStreamConsumer.Start(ctx); err != nil {
				s.logger.Error("IoT Stream Consumer failed", zap.Error(err))
			}
		}()
	}

	// 启动配置变更消费者（订阅 config:change:stream）
	if s.configConsumer != nil {
		go func() {
			if err := s.configConsumer.Start(ctx); err != nil {
				s.logger.Error("Config Consumer failed", zap.Error(err))
			}
		}()
	}

	// 阻塞主 goroutine，等待上下文取消
	<-ctx.Done()
	return nil
}

// startDataAggregation 启动数据聚合任务
func (s *AggregatorService) startDataAggregation(ctx context.Context) {
	interval := time.Duration(s.config.Aggregator.Aggregation.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Starting data aggregation",
		zap.Duration("interval", interval),
	)

	// 首次执行一次全量聚合
	if err := s.aggregateAllCards(ctx); err != nil {
		s.logger.Error("Failed to aggregate all cards on startup", zap.Error(err))
	}

	// 定时聚合
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.aggregateAllCards(ctx); err != nil {
				s.logger.Error("Failed to aggregate cards", zap.Error(err))
			}
		}
	}
}

// aggregateAllCards 聚合所有卡片的数据（不分租户）
func (s *AggregatorService) aggregateAllCards(ctx context.Context) error {
	// 获取所有租户的所有卡片
	cards, err := s.cardRepo.GetAllCards("")
	if err != nil {
		return fmt.Errorf("failed to get all cards: %w", err)
	}

	s.logger.Debug("Aggregating cards",
		zap.Int("card_count", len(cards)),
	)

	successCount := 0
	errorCount := 0

	for _, card := range cards {
		select {
		case <-ctx.Done():
			return nil
		default:
			// 聚合单个卡片（使用卡片自己的 tenant_id）
			vitalCard, err := s.dataAggregator.AggregateCard(ctx, card.TenantID, card.CardID)
			if err != nil {
				s.logger.Error("Failed to aggregate card",
					zap.String("card_id", card.CardID),
					zap.String("tenant_id", card.TenantID),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			// 更新缓存
			if err := s.cacheManager.UpdateFullCardCache(ctx, card.CardID, vitalCard); err != nil {
				s.logger.Error("Failed to update full card cache",
					zap.String("card_id", card.CardID),
					zap.String("tenant_id", card.TenantID),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			successCount++
		}
	}

	s.logger.Debug("Completed aggregating cards",
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
	)

	return nil
}

// Stop 停止服务
func (s *AggregatorService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping card aggregator service")

	// 关闭 Redis
	if s.redisClient != nil {
		if err := rediscommon.Close(s.redisClient); err != nil {
			s.logger.Error("Error closing redis connection", zap.Error(err))
		}
	}

	// 关闭数据库
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.Error("Error closing database connection", zap.Error(err))
		}
	}

	s.logger.Info("Card aggregator service stopped")
	return nil
}
