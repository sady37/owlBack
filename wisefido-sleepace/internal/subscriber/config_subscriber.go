package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rediscommon "owl-common/redis"
	"wisefido-sleepace/internal/service"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigSubscriber subscribes to config:card:stream and refreshes local card mapping cache.
// Note: wisefido-sleepace only reads from qinglan's Redis cache, so the "refresh" here
// is just a log/notification — the actual cache is maintained by qinglan.
// If we need local in-memory caching, we can extend this.
type ConfigSubscriber struct {
	redisClient *redis.Client
	cardMapping *service.CardMappingService
	logger      *zap.Logger
}

func NewConfigSubscriber(redisClient *redis.Client, cardMapping *service.CardMappingService, logger *zap.Logger) *ConfigSubscriber {
	return &ConfigSubscriber{
		redisClient: redisClient,
		cardMapping: cardMapping,
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

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		s.logger.Warn("unmarshal config message", zap.Error(err))
		return nil
	}

	msgType, _ := msg.Values["type"].(string)
	if msgType == "" {
		if t, ok := data["type"]; ok {
			msgType, _ = t.(string)
		}
	}

	switch msgType {
	case rediscommon.ConfigCardChanged:
		cardID, _ := data["card_id"].(string)
		s.cardMapping.InvalidateCache(ctx)
		s.logger.Info("card config changed, cache invalidated",
			zap.String("card_id", cardID),
			zap.String("branch_id", fmt.Sprint(data["branch_id"])))
	default:
		s.logger.Debug("unhandled config message type", zap.String("type", msgType))
	}

	return nil
}

// SubscribeLoop runs the blocking subscription loop for a single config stream.
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
