package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"wisefido-card-aggregator/internal/aggregator"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/consumer"
	"wisefido-card-aggregator/internal/repository"
	
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"owl-common/database"
	rediscommon "owl-common/redis"
)

// AggregatorService 卡片聚合服务
type AggregatorService struct {
	config        *config.Config
	logger        *zap.Logger
	db            *sql.DB
	redisClient   *redis.Client
	cardRepo      *repository.CardRepository
	cardCreator   *aggregator.CardCreator
	eventConsumer *consumer.EventConsumer
	dataAggregator *aggregator.DataAggregator
	cacheManager   *aggregator.CacheManager
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
	
	// 创建 CardCreator
	cardCreator := aggregator.NewCardCreator(cardRepo, logger)
	
	// 创建事件消费者（如果使用事件驱动模式）
	var eventConsumer *consumer.EventConsumer
	if cfg.Aggregator.TriggerMode == "events" {
		eventConsumer = consumer.NewEventConsumer(
			redisClient,
			cardCreator,
			cardRepo,
			logger,
			cfg.Aggregator.EventStream,
			cfg.Aggregator.ConsumerGroup,
			cfg.Aggregator.ConsumerName,
			int64(cfg.Aggregator.BatchSize),
		)
	}

	// 创建数据聚合器和缓存管理器（如果启用数据聚合）
	var dataAggregator *aggregator.DataAggregator
	var cacheManager *aggregator.CacheManager
	if cfg.Aggregator.Aggregation.Enabled {
		kv := aggregator.NewRedisKVStore(redisClient)
		cacheManager = aggregator.NewCacheManager(cfg, kv, logger)
		dataAggregator = aggregator.NewDataAggregator(cfg, kv, cardRepo, logger)
	}
	
	return &AggregatorService{
		config:         cfg,
		logger:         logger,
		db:             db,
		redisClient:    redisClient,
		cardRepo:       cardRepo,
		cardCreator:    cardCreator,
		eventConsumer:  eventConsumer,
		dataAggregator: dataAggregator,
		cacheManager:   cacheManager,
	}, nil
}

// Start 启动服务
func (s *AggregatorService) Start(ctx context.Context) error {
	s.logger.Info("Starting card aggregator service",
		zap.String("trigger_mode", s.config.Aggregator.TriggerMode),
		zap.Bool("aggregation_enabled", s.config.Aggregator.Aggregation.Enabled),
	)
	
	// 启动数据聚合任务（如果启用）
	if s.config.Aggregator.Aggregation.Enabled {
		go s.startDataAggregation(ctx)
	}
	
	// 根据触发模式启动不同的处理逻辑
	if s.config.Aggregator.TriggerMode == "polling" {
		// 📝 当前使用轮询模式（每60秒全量更新）
		//     事件驱动模式待 wisefido-data 服务实现后再启用
		//     详见：docs/PENDING_FEATURES.md
		return s.startPollingMode(ctx)
	} else if s.config.Aggregator.TriggerMode == "events" {
		// ⚠️ 事件驱动模式需要 wisefido-data 服务发布事件
		//     如果 wisefido-data 服务未实现，此模式无法正常工作
		return s.startEventDrivenMode(ctx)
	} else {
		return fmt.Errorf("unsupported trigger mode: %s", s.config.Aggregator.TriggerMode)
	}
}

// startPollingMode 启动轮询模式
func (s *AggregatorService) startPollingMode(ctx context.Context) error {
	interval := time.Duration(s.config.Aggregator.Polling.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	s.logger.Info("Starting polling mode",
		zap.Duration("interval", interval),
	)
	
	// 首次执行一次全量创建
	if err := s.createAllCards(ctx); err != nil {
		s.logger.Error("Failed to create all cards on startup", zap.Error(err))
	}
	
	// 定时轮询
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.createAllCards(ctx); err != nil {
				s.logger.Error("Failed to create cards", zap.Error(err))
			}
		}
	}
}

// createAllCards 为所有 unit 创建卡片
func (s *AggregatorService) createAllCards(ctx context.Context) error {
	s.logger.Info("Starting to create cards for all units")
	
	// 从配置获取 tenant_id
	tenantID := s.config.Aggregator.TenantID
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required, please set TENANT_ID environment variable")
	}
	
	// 获取所有 unit
	unitIDs, err := s.cardRepo.GetAllUnits(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get all units: %w", err)
	}
	
	s.logger.Info("Found units to process",
		zap.Int("unit_count", len(unitIDs)),
	)
	
	// 为每个 unit 创建卡片
	successCount := 0
	errorCount := 0
	
	for _, unitID := range unitIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := s.cardCreator.CreateCardsForUnit(tenantID, unitID); err != nil {
				s.logger.Error("Failed to create cards for unit",
					zap.String("unit_id", unitID),
					zap.Error(err),
				)
				errorCount++
			} else {
				successCount++
			}
		}
	}
	
	s.logger.Info("Completed creating cards",
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
	)
	
	return nil
}

// startEventDrivenMode 启动事件驱动模式
func (s *AggregatorService) startEventDrivenMode(ctx context.Context) error {
	s.logger.Info("Starting event-driven mode")
	
	// 首次执行一次全量创建
	if err := s.createAllCards(ctx); err != nil {
		s.logger.Error("Failed to create all cards on startup", zap.Error(err))
	}
	
	// 启动定时任务（每天上午9点）
	go s.startScheduledUpdate(ctx)
	
	// 启动事件消费者（阻塞）
	if s.eventConsumer != nil {
		return s.eventConsumer.Start(ctx)
	}
	
	return fmt.Errorf("event consumer not initialized")
}

// startScheduledUpdate 启动定时任务（每天上午9点全量更新）
func (s *AggregatorService) startScheduledUpdate(ctx context.Context) {
	s.logger.Info("Starting scheduled update task (daily at 9:00 AM)")
	
	for {
		// 计算到明天上午9点的时间
		now := time.Now()
		next9AM := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
		if next9AM.Before(now) {
			next9AM = next9AM.Add(24 * time.Hour)
		}
		
		duration := next9AM.Sub(now)
		timer := time.NewTimer(duration)
		
		s.logger.Info("Scheduled update will run at",
			zap.Time("next_run", next9AM),
			zap.Duration("wait_duration", duration),
		)
		
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			// 执行全量更新
			s.logger.Info("Running scheduled full update")
			if err := s.createAllCards(ctx); err != nil {
				s.logger.Error("Failed to create all cards in scheduled update", zap.Error(err))
			} else {
				s.logger.Info("Scheduled full update completed successfully")
			}
			
			// 重置定时器到明天上午9点
			timer.Reset(24 * time.Hour)
		}
	}
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

// aggregateAllCards 聚合所有卡片的数据
func (s *AggregatorService) aggregateAllCards(ctx context.Context) error {
	tenantID := s.config.Aggregator.TenantID
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	// 获取所有卡片
	cards, err := s.cardRepo.GetAllCards(tenantID)
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
			// 聚合单个卡片
			vitalCard, err := s.dataAggregator.AggregateCard(ctx, tenantID, card.CardID)
			if err != nil {
				s.logger.Error("Failed to aggregate card",
					zap.String("card_id", card.CardID),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			// 更新缓存
			if err := s.cacheManager.UpdateFullCardCache(ctx, card.CardID, vitalCard); err != nil {
				s.logger.Error("Failed to update full card cache",
					zap.String("card_id", card.CardID),
					zap.Error(err),
				)
				errorCount++
				continue
			}

			successCount++
		}
	}

	s.logger.Info("Completed aggregating cards",
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
		zap.Int("total_count", len(cards)),
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

