// stream_consumer.go — wisefido-iot v2 唯一持久化边界。
//
// 路由 (CLAUDE.md 规则 #2)：
//   iot:monitor:stream → monitor_stream (90d)
//   iot:event:stream   → event_log      (90d)
//   iot:alarm:stream   → skip (cardagg alarm_router 单 writer 写 alarm_events)
//   iot:stat:stream    → skip (sleepace 自己 DB；当前 owl_v2 无统一 stat 表)
//   iot:auth:stream    → skip (auth 走 device-gateway 自己处理)
//
// envelope 解析走 owl-common/redis.FromStreamMap 单一入口 (规则 #1.3)。

package consumer

import (
	"context"
	"fmt"
	"time"

	"wisefido-iot/internal/config"
	"wisefido-iot/internal/repository"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// persistSustainedThreshold 连续持久化失败到此条数后，逐条 error 降噪为周期采样的 sustained 告警。
const persistSustainedThreshold = 50

type StreamConsumer struct {
	config      *config.Config
	redisClient *redis.Client
	streamRepo  *repository.StreamRepo
	logger      *zap.Logger
	failStreak  map[string]int
}

func NewStreamConsumer(
	cfg *config.Config,
	redisClient *redis.Client,
	streamRepo *repository.StreamRepo,
	logger *zap.Logger,
) *StreamConsumer {
	return &StreamConsumer{
		config:      cfg,
		redisClient: redisClient,
		streamRepo:  streamRepo,
		logger:      logger,
		failStreak:  map[string]int{},
	}
}

// notePersistFailure 跟踪每流连续持久化失败。逐条 error 在持续故障下会刷屏淹没告警；
// 跨阈值后改为周期采样并打 sustained marker，让外部告警规则可抓「持续丢数据」这一态。
func (c *StreamConsumer) notePersistFailure(stream, msgID string, err error) {
	c.failStreak[stream]++
	n := c.failStreak[stream]
	if n < persistSustainedThreshold {
		c.logger.Error("process message failed",
			zap.String("stream", stream),
			zap.String("message_id", msgID),
			zap.Error(err),
		)
		return
	}
	if n == persistSustainedThreshold || n%500 == 0 {
		c.logger.Error("event_persist_sustained_failure",
			zap.String("stream", stream),
			zap.Int("consecutive", n),
			zap.String("message_id", msgID),
			zap.Error(err),
		)
	}
}

func (c *StreamConsumer) notePersistSuccess(stream string) {
	if c.failStreak[stream] > 0 {
		c.logger.Info("event_persist_recovered",
			zap.String("stream", stream),
			zap.Int("after_failures", c.failStreak[stream]),
		)
		c.failStreak[stream] = 0
	}
}

func (c *StreamConsumer) Start(ctx context.Context) error {
	streams := []string{
		c.config.Streams.Monitor,
		c.config.Streams.Event,
	}

	for _, stream := range streams {
		if err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, c.config.ConsumerGroup); err != nil {
			c.logger.Warn("create consumer group failed (will retry on read)",
				zap.String("stream", stream),
				zap.Error(err),
			)
		}
	}

	c.logger.Info("stream consumer started",
		zap.String("consumer_group", c.config.ConsumerGroup),
		zap.String("consumer_name", c.config.ConsumerName),
		zap.Strings("streams", streams),
	)

	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			monitorErr := c.consumeStream(ctx, c.config.Streams.Monitor)
			eventErr := c.consumeStream(ctx, c.config.Streams.Event)

			if monitorErr != nil && eventErr != nil {
				c.logger.Error("all streams failed", zap.Duration("backoff", backoff))
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(backoff):
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
				}
				continue
			}
			backoff = time.Second
			if monitorErr != nil {
				c.logger.Error("monitor stream consume failed", zap.Error(monitorErr))
			}
			if eventErr != nil {
				c.logger.Error("event stream consume failed", zap.Error(eventErr))
			}
		}
	}
}

func (c *StreamConsumer) consumeStream(ctx context.Context, streamName string) error {
	messages, err := rediscommon.ReadFromStreamWithBlock(
		ctx,
		c.redisClient,
		streamName,
		c.config.ConsumerGroup,
		c.config.ConsumerName,
		c.config.BatchSize,
		500*time.Millisecond,
	)
	if err != nil {
		return fmt.Errorf("read stream %s: %w", streamName, err)
	}

	for _, raw := range messages {
		if err := c.processMessage(ctx, streamName, raw); err != nil {
			c.notePersistFailure(streamName, raw.ID, err)
		} else {
			c.notePersistSuccess(streamName)
		}
		if err := c.redisClient.XAck(ctx, streamName, c.config.ConsumerGroup, raw.ID).Err(); err != nil {
			c.logger.Warn("XAck failed",
				zap.String("stream", streamName),
				zap.String("message_id", raw.ID),
				zap.Error(err),
			)
		}
	}
	return nil
}

func (c *StreamConsumer) processMessage(ctx context.Context, streamName string, raw rediscommon.StreamMessage) error {
	msg, err := rediscommon.FromStreamMap(raw.Values)
	if err != nil {
		return fmt.Errorf("parse envelope: %w", err)
	}

	switch streamName {
	case c.config.Streams.Monitor:
		return c.streamRepo.InsertMonitor(ctx, msg)
	case c.config.Streams.Event:
		return c.streamRepo.InsertEvent(ctx, msg)
	default:
		return nil
	}
}
