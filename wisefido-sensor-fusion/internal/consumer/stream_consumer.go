package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"wisefido-sensor-fusion/internal/config"
	"wisefido-sensor-fusion/internal/fusion"
	"wisefido-sensor-fusion/internal/models"
	"wisefido-sensor-fusion/internal/repository"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	rediscommon "owl-common/redis"
)

// Metrics 监控指标
type Metrics struct {
	mu sync.RWMutex
	
	// 消息处理统计
	MessagesProcessed    int64 // 处理的消息总数
	MessagesSucceeded    int64 // 成功处理的消息数
	MessagesFailed       int64 // 处理失败的消息数
	MessagesSkipped      int64 // 跳过的消息数（设备未绑定等）
	
	// 错误分类统计
	ErrorsParse          int64 // 解析错误
	ErrorsCardNotFound   int64 // 卡片未找到
	ErrorsFusionFailed   int64 // 融合失败
	ErrorsCacheFailed    int64 // 缓存更新失败
	
	// 性能指标
	TotalProcessingTime time.Duration // 总处理时间
	LastProcessTime     time.Time     // 最后处理时间
	
	// 启动时间
	StartTime time.Time
}

// GetSnapshot 获取指标快照（线程安全）
func (m *Metrics) GetSnapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Metrics{
		MessagesProcessed:    m.MessagesProcessed,
		MessagesSucceeded:    m.MessagesSucceeded,
		MessagesFailed:       m.MessagesFailed,
		MessagesSkipped:      m.MessagesSkipped,
		ErrorsParse:          m.ErrorsParse,
		ErrorsCardNotFound:   m.ErrorsCardNotFound,
		ErrorsFusionFailed:   m.ErrorsFusionFailed,
		ErrorsCacheFailed:    m.ErrorsCacheFailed,
		TotalProcessingTime:  m.TotalProcessingTime,
		LastProcessTime:      m.LastProcessTime,
		StartTime:            m.StartTime,
	}
}

// IncrementProcessed 增加处理计数
func (m *Metrics) IncrementProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesProcessed++
}

// IncrementSucceeded 增加成功计数
func (m *Metrics) IncrementSucceeded(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSucceeded++
	m.TotalProcessingTime += duration
	m.LastProcessTime = time.Now()
}

// IncrementFailed 增加失败计数
func (m *Metrics) IncrementFailed(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesFailed++
	switch errorType {
	case "parse":
		m.ErrorsParse++
	case "card_not_found":
		m.ErrorsCardNotFound++
	case "fusion_failed":
		m.ErrorsFusionFailed++
	case "cache_failed":
		m.ErrorsCacheFailed++
	}
}

// IncrementSkipped 增加跳过计数
func (m *Metrics) IncrementSkipped() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSkipped++
}

// StreamConsumer Redis Streams 消费者
type StreamConsumer struct {
	config       *config.Config
	redisClient  *redis.Client
	cardRepo     *repository.CardRepository
	iotRepo      *repository.IoTTimeSeriesRepository
	fusion       *fusion.SensorFusion
	cache        *CacheManager
	logger       *zap.Logger
	metrics      *Metrics
}

// NewStreamConsumer 创建 Streams 消费者
func NewStreamConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	cardRepo *repository.CardRepository,
	iotRepo *repository.IoTTimeSeriesRepository,
	fusion *fusion.SensorFusion,
	cache *CacheManager,
	logger *zap.Logger,
) *StreamConsumer {
	return &StreamConsumer{
		config:      cfg,
		redisClient: redisClient,
		cardRepo:    cardRepo,
		iotRepo:     iotRepo,
		fusion:      fusion,
		cache:       cache,
		logger:      logger,
		metrics: &Metrics{
			StartTime: time.Now(),
		},
	}
}

// Start 启动消费者
func (c *StreamConsumer) Start(ctx context.Context) error {
	// 创建消费者组
	stream := c.config.Fusion.Stream.Input
	if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, c.config.Fusion.ConsumerGroup); err != nil {
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}
	
	c.logger.Info("Stream consumer started",
		zap.String("consumer_group", c.config.Fusion.ConsumerGroup),
		zap.String("consumer_name", c.config.Fusion.ConsumerName),
		zap.String("stream", stream),
	)
	
	// 启动指标报告协程
	metricsCtx, metricsCancel := context.WithCancel(ctx)
	defer metricsCancel()
	go c.reportMetrics(metricsCtx)
	
	// 启动消费循环
	backoffDuration := time.Second // 初始退避时间
	maxBackoff := 30 * time.Second // 最大退避时间
	
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.consumeStream(ctx, stream); err != nil {
				c.logger.Error("Failed to consume stream", 
					zap.Error(err),
					zap.Duration("backoff", backoffDuration),
				)
				
				// 指数退避：等待后重试
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(backoffDuration):
					// 指数退避，但不超过最大值
					backoffDuration *= 2
					if backoffDuration > maxBackoff {
						backoffDuration = maxBackoff
					}
				}
			} else {
				// 成功时重置退避时间
				backoffDuration = time.Second
			}
		}
	}
}

// consumeStream 消费单个 Stream
func (c *StreamConsumer) consumeStream(ctx context.Context, stream string) error {
	// 从 Stream 读取消息
	messages, err := rediscommon.ReadFromStream(
		ctx,
		c.redisClient,
		stream,
		c.config.Fusion.ConsumerGroup,
		c.config.Fusion.ConsumerName,
		c.config.Fusion.BatchSize,
	)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %w", err)
	}
	
	// 处理消息
	for _, msg := range messages {
		c.metrics.IncrementProcessed()
		if err := c.processMessage(ctx, msg); err != nil {
			c.logger.Error("Failed to process message",
				zap.String("stream_id", msg.ID),
				zap.Error(err),
			)
			// 继续处理下一条消息，不中断
		}
	}
	
	return nil
}

// processMessage 处理单条消息
//
// ⚠️ 依赖说明：
// - 本函数依赖 PostgreSQL cards 表，需要 wisefido-card-aggregator 服务先创建卡片
// - 如果 cards 表为空，GetCardByDeviceID 会返回错误，导致融合失败
// - 当前 wisefido-card-aggregator 的卡片创建功能还未实现，需要优先实现
//
// 处理流程：
// 1. 解析设备数据消息
// 2. 根据 device_id 查询 cards 表（依赖卡片管理层）
// 3. 融合该卡片的所有设备数据
// 4. 更新 Redis 缓存
func (c *StreamConsumer) processMessage(ctx context.Context, msg rediscommon.StreamMessage) error {
	startTime := time.Now()
	
	// 解析消息数据
	var dataStr string
	if val, ok := msg.Values["data"]; ok {
		if str, ok := val.(string); ok {
			dataStr = str
		} else {
			c.metrics.IncrementFailed("parse")
			return fmt.Errorf("invalid data format in message")
		}
	} else {
		c.metrics.IncrementFailed("parse")
		return fmt.Errorf("missing data field in message")
	}
	
	// 解析 JSON
	var iotData models.IoTDataMessage
	if err := json.Unmarshal([]byte(dataStr), &iotData); err != nil {
		c.metrics.IncrementFailed("parse")
		c.logger.Error("Failed to parse message data",
			zap.String("stream_id", msg.ID),
			zap.String("device_id", iotData.DeviceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal message data: %w", err)
	}
	
	c.logger.Debug("Processing IoT data",
		zap.String("device_id", iotData.DeviceID),
		zap.String("device_type", iotData.DeviceType),
		zap.String("tenant_id", iotData.TenantID),
	)
	
	// 1. 根据 device_id 和 tenant_id 查询关联的卡片
	cardInfo, err := c.cardRepo.GetCardByDeviceID(iotData.TenantID, iotData.DeviceID)
	if err != nil {
		c.metrics.IncrementSkipped()
		c.logger.Warn("Card not found for device",
			zap.String("device_id", iotData.DeviceID),
			zap.String("tenant_id", iotData.TenantID),
			zap.Error(err),
		)
		return nil // 设备可能未绑定到卡片，忽略
	}
	
	// 2. 融合卡片的所有设备数据（传递卡片类型）
	realtimeData, err := c.fusion.FuseCardData(cardInfo.TenantID, cardInfo.CardID, cardInfo.CardType)
	if err != nil {
		c.metrics.IncrementFailed("fusion_failed")
		c.logger.Error("Failed to fuse card data",
			zap.String("card_id", cardInfo.CardID),
			zap.String("device_id", iotData.DeviceID),
			zap.String("tenant_id", iotData.TenantID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to fuse card data: %w", err)
	}
	
	// 3. 更新 Redis 缓存
	if err := c.cache.UpdateRealtimeData(cardInfo.CardID, realtimeData); err != nil {
		c.metrics.IncrementFailed("cache_failed")
		c.logger.Error("Failed to update cache",
			zap.String("card_id", cardInfo.CardID),
			zap.String("device_id", iotData.DeviceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update cache: %w", err)
	}
	
	processingDuration := time.Since(startTime)
	c.metrics.IncrementSucceeded(processingDuration)
	
	c.logger.Info("Fused and cached card data",
		zap.String("card_id", cardInfo.CardID),
		zap.String("device_id", iotData.DeviceID),
		zap.String("tenant_id", iotData.TenantID),
		zap.Duration("processing_time", processingDuration),
	)
	
	return nil
}

// reportMetrics 定期报告指标（每60秒）
func (c *StreamConsumer) reportMetrics(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := c.metrics.GetSnapshot()
			uptime := time.Since(snapshot.StartTime)
			
			var avgProcessingTime time.Duration
			if snapshot.MessagesSucceeded > 0 {
				avgProcessingTime = snapshot.TotalProcessingTime / time.Duration(snapshot.MessagesSucceeded)
			}
			
			successRate := float64(0)
			if snapshot.MessagesProcessed > 0 {
				successRate = float64(snapshot.MessagesSucceeded) / float64(snapshot.MessagesProcessed) * 100
			}
			
			c.logger.Info("Metrics report",
				zap.Int64("messages_processed", snapshot.MessagesProcessed),
				zap.Int64("messages_succeeded", snapshot.MessagesSucceeded),
				zap.Int64("messages_failed", snapshot.MessagesFailed),
				zap.Int64("messages_skipped", snapshot.MessagesSkipped),
				zap.Float64("success_rate", successRate),
				zap.Int64("errors_parse", snapshot.ErrorsParse),
				zap.Int64("errors_card_not_found", snapshot.ErrorsCardNotFound),
				zap.Int64("errors_fusion_failed", snapshot.ErrorsFusionFailed),
				zap.Int64("errors_cache_failed", snapshot.ErrorsCacheFailed),
				zap.Duration("avg_processing_time", avgProcessingTime),
				zap.Duration("uptime", uptime),
			)
		}
	}
}

