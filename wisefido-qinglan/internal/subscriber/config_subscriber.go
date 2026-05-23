package subscriber

import (
	"context"
	"encoding/json"
	"fmt"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/service"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigSubscriber 订阅 config:card:stream — 按 op + 受影响范围失效 baseline cache。
type ConfigSubscriber struct {
	redisClient    *redis.Client
	config         *config.Config
	logger         *zap.Logger
	cardMappingSvc *service.CardMappingService
	consumerGroup  string
	consumerName   string
}

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

func (s *ConfigSubscriber) Start(ctx context.Context) error {
	s.logger.Info("Starting config change subscriber",
		zap.String("stream", rediscommon.StreamConfigCard.Name),
		zap.String("consumer_group", s.consumerGroup),
		zap.String("consumer_name", s.consumerName),
	)
	return nil
}

func (s *ConfigSubscriber) Stop() error {
	s.logger.Info("Config change subscriber stopped")
	return nil
}

// HandleConfigChangeMessage 解析 config.changed CloudEvent 并按 op 失效 baseline cache。
func (s *ConfigSubscriber) HandleConfigChangeMessage(ctx context.Context, stream string, message redis.XMessage) error {
	dataStr, ok := message.Values["data"].(string)
	if !ok || dataStr == "" {
		return fmt.Errorf("config change message missing 'data' field")
	}
	var configMsg rediscommon.ConfigChangeMessage
	if err := json.Unmarshal([]byte(dataStr), &configMsg); err != nil {
		return fmt.Errorf("parse config change envelope: %w", err)
	}
	if configMsg.Type != rediscommon.ConfigCardChanged {
		return nil
	}

	op, _ := configMsg.Data["op"].(string)
	uids := stringArrayFromData(configMsg.Data, "device_uids")

	if op == "reset" {
		s.cardMappingSvc.InvalidateCache()
		s.logger.Info("config.changed processed",
			zap.String("message_id", message.ID), zap.String("op", op))
		return nil
	}

	for _, uid := range uids {
		s.cardMappingSvc.InvalidateByDeviceUID(uid)
		s.cardMappingSvc.RefreshBaseline(ctx, uid)
	}
	s.logger.Info("config.changed processed",
		zap.String("message_id", message.ID),
		zap.String("op", op),
		zap.Int("device_uids", len(uids)))
	return nil
}

func stringArrayFromData(data map[string]interface{}, key string) []string {
	if data == nil {
		return nil
	}
	raw, ok := data[key]
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
	}
	return nil
}
