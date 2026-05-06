package subscriber

import (
	"context"
	"fmt"
	"strings"
	"time"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ProbeSubscriber 订阅 iot:probe:device:stream，对 device_type=Sleepad 的请求
// 立刻调用 HealthCheck.ProbeAfterCardChange 触发一次单设备 probe，
// 缩短前端 refresh 到状态更新的延迟（不再等 80s 周期 ticker）。
type ProbeSubscriber struct {
	healthCheck *HealthCheck
	logger      *zap.Logger
}

func NewProbeSubscriber(healthCheck *HealthCheck, logger *zap.Logger) *ProbeSubscriber {
	return &ProbeSubscriber{healthCheck: healthCheck, logger: logger}
}

func (s *ProbeSubscriber) HandleProbeMessage(ctx context.Context, msg redis.XMessage) error {
	deviceType := strings.TrimSpace(stringField(msg.Values, "device_type"))
	if !strings.EqualFold(deviceType, "Sleepad") && !strings.EqualFold(deviceType, "SleepPad") {
		return nil
	}
	deviceUID := strings.TrimSpace(stringField(msg.Values, "device_uid"))
	deviceID := strings.TrimSpace(stringField(msg.Values, "device_id"))
	if deviceUID == "" || deviceID == "" {
		return nil
	}
	tenantID := stringField(msg.Values, "tenant_id")
	cardID := stringField(msg.Values, "card_id")
	source := stringField(msg.Values, "trigger_source")

	s.logger.Info("probe sleepad requested",
		zap.String("device_uid", deviceUID),
		zap.String("device_id", deviceID),
		zap.String("source", source),
	)

	// 复用现有 ProbeAfterCardChange：传入 affectedUIDs=[deviceUID] 走单设备 probe 路径，
	// 不会触发 scanAll 风暴。HealthCheck 内置 probeMinInterval 去重防止短时间多次抖动。
	go s.healthCheck.ProbeAfterCardChange(ctx, tenantID, "", cardID, deviceID, []string{deviceUID})
	return nil
}

// SubscribeProbeLoop 阻塞消费 iot:probe:device:stream。
func SubscribeProbeLoop(ctx context.Context, logger *zap.Logger, redisClient *redis.Client,
	groupName, consumerName string, sub *ProbeSubscriber) {

	stream := rediscommon.StreamProbeDevice.Name
	logger.Info("starting probe device stream subscriber",
		zap.String("stream", stream),
		zap.String("group", groupName))

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, redisClient, stream, groupName, consumerName, 10, 5*time.Second)
		if err != nil {
			logger.Debug("read probe stream", zap.String("stream", stream), zap.Error(err))
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
			if err := sub.HandleProbeMessage(ctx, xMsg); err != nil {
				logger.Error("handle probe message", zap.String("stream", stream), zap.Error(err))
			}
			redisClient.XAck(ctx, stream, groupName, msg.ID)
		}
	}
}

func stringField(values map[string]interface{}, key string) string {
	if v, ok := values[key]; ok && v != nil {
		switch s := v.(type) {
		case string:
			return s
		case []byte:
			return string(s)
		default:
			return fmt.Sprintf("%v", s)
		}
	}
	return ""
}
