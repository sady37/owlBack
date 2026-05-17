package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/service"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigSubscriber 配置变更订阅器
type ConfigSubscriber struct {
	redisClient    *redis.Client
	config         *config.Config
	logger         *zap.Logger
	cardMappingSvc *service.CardMappingService
	consumerGroup  string
	consumerName   string
}

// NewConfigSubscriber 创建配置变更订阅器
func NewConfigSubscriber(
	redisClient *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
	cardMappingSvc *service.CardMappingService,
) *ConfigSubscriber {
	return &ConfigSubscriber{
		redisClient:    redisClient,
		config:         cfg,
		logger:         logger,
		cardMappingSvc: cardMappingSvc,
		consumerGroup:  "qinglan-config-consumer",
		consumerName:   "qinglan-config-consumer-1",
	}
}

// Start 仅日志：消费协程由 main.go runConfigCardStreamReader 启动 + XReadGroup MKSTREAM 隐式建组。
func (s *ConfigSubscriber) Start(ctx context.Context) error {
	s.logger.Info("Starting config change subscriber",
		zap.String("stream", rediscommon.StreamConfigCard.Name),
		zap.String("consumer_group", s.consumerGroup),
		zap.String("consumer_name", s.consumerName),
	)
	return nil
}

// HandleConfigChangeMessage 处理配置变更消息（支持 CloudEvents 格式和旧格式）
func (s *ConfigSubscriber) HandleConfigChangeMessage(ctx context.Context, stream string, message redis.XMessage) error {
	s.logger.Debug("Received config change message",
		zap.String("stream", stream),
		zap.String("message_id", message.ID),
	)

	// 包装格式（wisefido-data PublishJSONToStream）：从 "data" 字段解析整个 JSON envelope
	dataStr, ok := message.Values["data"].(string)
	if !ok || dataStr == "" {
		return fmt.Errorf("config change message missing 'data' field")
	}
	var configMsg rediscommon.ConfigChangeMessage
	if err := json.Unmarshal([]byte(dataStr), &configMsg); err != nil {
		return fmt.Errorf("parse config change envelope: %w", err)
	}

	// 根据事件类型处理配置变更
	switch configMsg.Type {
	case rediscommon.ConfigCardChanged:
		s.handleCardChange(configMsg, message.ID)
	default:
		s.logger.Debug("Skipping config change event (not subscribed)",
			zap.String("event_type", configMsg.Type),
		)
	}

	return nil
}

// Stop 停止配置变更订阅器。subscriber goroutine 由 main 通过 ctx 取消，本方法仅日志。
func (s *ConfigSubscriber) Stop() error {
	s.logger.Info("Config change subscriber stopped")
	return nil
}

// handleCardChange 处理卡片配置变更 (BuildCardChangeMessage)
func (s *ConfigSubscriber) handleCardChange(configMsg rediscommon.ConfigChangeMessage, messageID string) {
	ctx := context.Background()

	// 检查消息时间戳，防止过时消息覆盖新配置
	// 允许最大时间差 5 秒（包括网络延迟和时钟偏差）
	var messageTimestampMs int64
	if ts, ok := configMsg.Data["timestamp_ms"].(float64); ok {
		messageTimestampMs = int64(ts)
	} else if ts, ok := configMsg.Data["timestamp_ms"].(int64); ok {
		messageTimestampMs = ts
	}

	if messageTimestampMs > 0 {
		currentTimeMs := time.Now().UnixMilli()
		timeDiffMs := currentTimeMs - messageTimestampMs
		const maxAllowedDelayMs = 5000 // 5 秒

		if timeDiffMs > maxAllowedDelayMs {
			s.logger.Warn("Discarding stale card config message (too old)",
				zap.String("message_id", messageID),
				zap.String("type", configMsg.Type),
				zap.Int64("time_diff_ms", timeDiffMs),
				zap.Int64("max_allowed_ms", maxAllowedDelayMs))
			return
		}

		if timeDiffMs < -5000 { // 消息时间戳在未来
			s.logger.Warn("Discarding card config message with future timestamp",
				zap.String("message_id", messageID),
				zap.String("type", configMsg.Type),
				zap.Int64("time_diff_ms", timeDiffMs))
			return
		}
	}

	s.logger.Info("Handling card config change",
		zap.String("message_id", messageID),
		zap.String("type", configMsg.Type),
	)

	// 消息数据已经是 map[string]interface{}
	cardData := configMsg.Data

	// 提取关键字段
	tenantID, _ := cardData["tenant_id"].(string)
	op, _ := cardData["op"].(string)

	if op == "reset" {
		s.cardMappingSvc.InvalidateCache()
		s.logger.Info("Card change processed (reset)",
			zap.String("message_id", messageID))
		return
	}

	if tenantID == "" {
		s.logger.Warn("Card change message missing tenant_id",
			zap.String("message_id", messageID))
		return
	}

	s.handleNormalCardChange(ctx, cardData, messageID)
}

// handleNormalCardChange 处理正常的卡片配置变更 (ConfigCardChanged)
// 优先使用 affected_device_uids 刷新；回退到 card_id 失效；否则全量失效。
func (s *ConfigSubscriber) handleNormalCardChange(ctx context.Context, cardData map[string]interface{}, messageID string) {
	cardID, _ := cardData["card_id"].(string)
	tenantID, _ := cardData["tenant_id"].(string)
	unitID, _ := cardData["unit_id"].(string)

	uids := affectedDeviceUIDsFromCardData(cardData)
	if len(uids) > 0 {
		for _, uid := range uids {
			s.cardMappingSvc.InvalidateByDeviceUID(uid)
			s.cardMappingSvc.RefreshBaseline(ctx, uid)
		}
	} else if cardID != "" {
		s.cardMappingSvc.InvalidateByCardID(cardID)
	} else {
		s.cardMappingSvc.InvalidateCache()
	}

	s.logger.Info("Card change processed",
		zap.String("message_id", messageID),
		zap.String("card_id", cardID),
		zap.String("tenant_id", tenantID),
		zap.String("unit_id", unitID),
		zap.Int("affected_uids", len(uids)))

	// data 中带 device_id 时按设备刷新 baseline（device_store / devices 变更）。
	if deviceID, _ := cardData["device_id"].(string); deviceID != "" {
		s.handleDeviceStoreChange(ctx, cardData, messageID)
	}
}

// handleDeviceStoreChange data.device_id 非空时：按 device_uid 失效并重刷 baseline
func (s *ConfigSubscriber) handleDeviceStoreChange(ctx context.Context, data map[string]interface{}, messageID string) {
	deviceUID, _ := data["device_uid"].(string)
	if deviceUID == "" {
		return
	}
	s.cardMappingSvc.InvalidateByDeviceUID(deviceUID)
	s.cardMappingSvc.RefreshBaseline(ctx, deviceUID)

	s.logger.Info("Device store change processed",
		zap.String("message_id", messageID),
		zap.String("device_uid", deviceUID))
}

func affectedDeviceUIDsFromCardData(data map[string]interface{}) []string {
	if data == nil {
		return nil
	}
	raw, ok := data["affected_device_uids"]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, x := range v {
			if s, ok := x.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

