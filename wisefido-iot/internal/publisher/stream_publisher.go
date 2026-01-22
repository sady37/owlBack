package publisher

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	rediscommon "owl-common/redis"
	"go.uber.org/zap"
)

// StreamPublisher Redis Stream 发布器
type StreamPublisher struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewStreamPublisher 创建 Stream 发布器
func NewStreamPublisher(redisClient *redis.Client, logger *zap.Logger) *StreamPublisher {
	return &StreamPublisher{
		redisClient: redisClient,
		logger:      logger,
	}
}

// Publish 发布数据到 Redis Stream
func (p *StreamPublisher) Publish(ctx context.Context, streamName string, data map[string]interface{}) (string, error) {
	// 使用默认配置：maxLen=1000, retentionSeconds=0（不限制）
	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamName, data, 1000, 0)
	if err != nil {
		p.logger.Error("Failed to publish to stream",
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to publish to stream %s: %w", streamName, err)
	}

	p.logger.Debug("Published to stream",
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return streamID, nil
}
