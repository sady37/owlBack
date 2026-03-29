package subscriber

import (
	"context"
	"encoding/json"
	"time"

	rediscommon "owl-common/redis"
	"wisefido-sleepace/internal/service"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigSubscriber 订阅 config:card:stream：失效 cardMapping 并触发 HealthCheck 探测。
type ConfigSubscriber struct {
	redisClient *redis.Client
	cardMapping *service.CardMappingService
	healthCheck *HealthCheck
	logger      *zap.Logger
}

func NewConfigSubscriber(redisClient *redis.Client, cardMapping *service.CardMappingService, healthCheck *HealthCheck, logger *zap.Logger) *ConfigSubscriber {
	return &ConfigSubscriber{
		redisClient: redisClient,
		cardMapping: cardMapping,
		healthCheck: healthCheck,
		logger:      logger,
	}
}

func (s *ConfigSubscriber) HandleConfigMessage(ctx context.Context, stream string, msg redis.XMessage) error {
	rawData, ok := msg.Values["data"]
	if !ok {
		return nil
	}
	dataStr, ok := rawData.(string)
	if !ok {
		return nil
	}

	var configMsg rediscommon.ConfigChangeMessage
	if err := json.Unmarshal([]byte(dataStr), &configMsg); err != nil {
		s.logger.Warn("unmarshal config CloudEvent", zap.Error(err))
		return nil
	}

	switch configMsg.Type {
	case rediscommon.ConfigCardChanged:
		s.handleCardLike(ctx, configMsg)
	default:
		s.logger.Debug("skip config event type", zap.String("type", configMsg.Type))
	}
	return nil
}

func (s *ConfigSubscriber) handleCardLike(ctx context.Context, configMsg rediscommon.ConfigChangeMessage) {
	data := configMsg.Data
	if data == nil {
		return
	}
	tenantID, _ := data["tenant_id"].(string)
	unitID, _ := data["unit_id"].(string)
	cardID, _ := data["card_id"].(string)
	deviceID, _ := data["device_id"].(string)

	if tenantID != "" && unitID != "" {
		s.cardMapping.InvalidateByTenantUnit(ctx, tenantID, unitID)
	} else if cardID != "" {
		s.cardMapping.InvalidateByCardID(cardID)
	} else if deviceID != "" {
		uid := s.cardMapping.ResolveToDeviceUID(ctx, deviceID)
		if uid != "" {
			s.cardMapping.InvalidateByDeviceUID(uid)
		} else {
			s.cardMapping.InvalidateCache(ctx)
		}
	} else {
		s.cardMapping.InvalidateCache(ctx)
	}

	if s.healthCheck != nil {
		s.healthCheck.ProbeAfterCardChange(ctx, tenantID, unitID, cardID, deviceID)
	}
}

// SubscribeLoop 阻塞消费 config:card:stream。
func SubscribeLoop(ctx context.Context, logger *zap.Logger, redisClient *redis.Client,
	stream, groupName, consumerName string, sub *ConfigSubscriber) {

	logger.Info("starting config stream subscriber",
		zap.String("stream", stream),
		zap.String("group", groupName))

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			logger.Info("config subscriber stopped", zap.String("stream", stream))
			return
		default:
		}

		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, redisClient, stream, groupName, consumerName, 10, 5*time.Second)
		if err != nil {
			logger.Debug("read config stream", zap.String("stream", stream), zap.Error(err))
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		if len(msgs) > 0 {
			backoff = time.Second
		}

		for _, msg := range msgs {
			xMsg := redis.XMessage{ID: msg.ID, Values: msg.Values}
			if err := sub.HandleConfigMessage(ctx, stream, xMsg); err != nil {
				logger.Error("handle config message", zap.String("stream", stream), zap.Error(err))
			}
			redisClient.XAck(ctx, stream, groupName, msg.ID)
		}
	}
}
