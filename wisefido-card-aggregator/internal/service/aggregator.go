package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
	"wisefido-card-aggregator/internal/aggregator"
	"wisefido-card-aggregator/internal/config"
	"wisefido-card-aggregator/internal/consumer"
	"wisefido-card-aggregator/internal/repository"
	
	"owl-common/card"
	"owl-common/database"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// AggregatorService 卡片聚合服务
type AggregatorService struct {
	config        *config.Config
	logger        *zap.Logger
	db            *sql.DB
	redisClient   *redis.Client
	cardRepo      *repository.CardRepository
	cardCreator   *card.CardCreator
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
	
	// 创建 CardCreator（使用 owl-common/card 包）
	cardCreator := card.NewCardCreator(cardRepo, logger)
	
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
// 如果 CARD_POLLING_INTERVAL >= 86400 (24小时)，则使用定时任务（每天8点执行）
// 否则使用固定间隔轮询
func (s *AggregatorService) startPollingMode(ctx context.Context) error {
	interval := time.Duration(s.config.Aggregator.Polling.Interval) * time.Second
	
	// 如果间隔 >= 24小时，使用定时任务（每天8点执行）
	if interval >= 24*time.Hour {
		s.logger.Info("Starting polling mode with scheduled update (daily at 8:00 AM)",
			zap.Duration("interval", interval),
		)
		
		// 首次执行一次全量创建
		if err := s.createAllCards(ctx); err != nil {
			s.logger.Error("Failed to create all cards on startup", zap.Error(err))
		}
		
		// 启动定时任务（每天上午8点）
		return s.startScheduledUpdateAt8AM(ctx)
	}
	
	// 否则使用固定间隔轮询
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	s.logger.Info("Starting polling mode with fixed interval",
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
	
	// 统计更新前的卡片数量
	originalCardCount, err := s.cardRepo.CountCardsByTenant(tenantID)
	if err != nil {
		s.logger.Warn("Failed to count original cards, continuing anyway",
			zap.Error(err),
		)
		originalCardCount = -1 // Use -1 to indicate unknown
	}
	
	// 获取所有 unit
	unitIDs, err := s.cardRepo.GetAllUnits(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get all units: %w", err)
	}
	
	s.logger.Info("Found units to process",
		zap.Int("unit_count", len(unitIDs)),
	)
	
	// 为每个 unit 创建卡片，收集统计信息
	successCount := 0
	errorCount := 0
	totalStats := struct {
		ExistingCount  int
		DeletedCount   int
		CreatedCount   int
		UpdatedCount   int
		UnchangedCount int
	}{}
	
	for _, unitID := range unitIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
			stats, err := s.cardCreator.CreateCardsForUnit(tenantID, unitID)
			if err != nil {
				s.logger.Error("Failed to create cards for unit",
					zap.String("unit_id", unitID),
					zap.Error(err),
				)
				errorCount++
			} else {
				successCount++
				if stats != nil {
					totalStats.ExistingCount += stats.ExistingCount
					totalStats.DeletedCount += stats.DeletedCount
					totalStats.CreatedCount += stats.CreatedCount
					totalStats.UpdatedCount += stats.UpdatedCount
					totalStats.UnchangedCount += stats.UnchangedCount
				}
			}
		}
	}
	
	// 统计更新后的卡片数量
	finalCardCount, err := s.cardRepo.CountCardsByTenant(tenantID)
	if err != nil {
		s.logger.Warn("Failed to count final cards, continuing anyway",
			zap.Error(err),
		)
		finalCardCount = -1 // Use -1 to indicate unknown
	}
	
	// Output statistics to stdout (also logged)
	updateCount := totalStats.DeletedCount + totalStats.CreatedCount + totalStats.UpdatedCount
	summaryMsg := fmt.Sprintf(
		"\n=== Card Check/Update Statistics ===\n"+
			"Original card count: %d\n"+
			"Updated card count: %d (deleted: %d, created: %d, content updated: %d)\n"+
			"Unchanged cards: %d\n"+
			"Final card count: %d\n"+
			"Units processed: %d (success: %d, failed: %d)\n"+
			"===================================\n",
		originalCardCount,
		updateCount,
		totalStats.DeletedCount,
		totalStats.CreatedCount,
		totalStats.UpdatedCount,
		totalStats.UnchangedCount,
		finalCardCount,
		len(unitIDs),
		successCount,
		errorCount,
	)
	
	// Output to stdout (use os.Stdout to ensure proper output)
	os.Stdout.WriteString(summaryMsg)
	
	// 同时记录到日志
	s.logger.Info("Completed creating cards",
		zap.Int("original_count", originalCardCount),
		zap.Int("updated_count", updateCount),
		zap.Int("deleted_count", totalStats.DeletedCount),
		zap.Int("created_count", totalStats.CreatedCount),
		zap.Int("content_updated_count", totalStats.UpdatedCount),
		zap.Int("unchanged_count", totalStats.UnchangedCount),
		zap.Int("final_count", finalCardCount),
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
	)
	
	// 更新所有卡片的报警计数（服务启动时和定时任务时）
	s.logger.Info("Starting to update alarm counts for all cards")
	if err := s.cardRepo.UpdateAllCardsAlarmCounts(ctx, tenantID); err != nil {
		s.logger.Warn("Failed to update alarm counts for all cards",
			zap.Error(err),
		)
		// 不返回错误，卡片创建已成功
	} else {
		s.logger.Info("Completed updating alarm counts for all cards")
	}
	
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
// 用于事件驱动模式
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

// startScheduledUpdateAt8AM 启动定时任务（每天上午8点全量更新）
// 用于轮询模式（当 CARD_POLLING_INTERVAL >= 24小时时）
func (s *AggregatorService) startScheduledUpdateAt8AM(ctx context.Context) error {
	s.logger.Info("Starting scheduled update task (daily at 8:00 AM)")
	
	for {
		// 计算到明天上午8点的时间
		now := time.Now()
		next8AM := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		if next8AM.Before(now) {
			next8AM = next8AM.Add(24 * time.Hour)
		}
		
		duration := next8AM.Sub(now)
		timer := time.NewTimer(duration)
		
		s.logger.Info("Scheduled update will run at",
			zap.Time("next_run", next8AM),
			zap.Duration("wait_duration", duration),
		)
		
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
			// 执行全量更新
			s.logger.Info("Running scheduled full update (daily at 8:00 AM)")
			if err := s.createAllCards(ctx); err != nil {
				s.logger.Error("Failed to create all cards in scheduled update", zap.Error(err))
			} else {
				s.logger.Info("Scheduled full update completed successfully")
			}
			
			// 重置定时器到明天上午8点
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

	s.logger.Debug("Completed aggregating cards",
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

