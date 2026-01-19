package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigNotifier 配置变更通知器
// 负责发布配置变更事件到 config:change:stream
type ConfigNotifier struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewConfigNotifier 创建配置通知器实例
func NewConfigNotifier(redisClient *redis.Client, logger *zap.Logger) *ConfigNotifier {
	return &ConfigNotifier{
		redisClient: redisClient,
		logger:      logger,
	}
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	EventType string `json:"event_type"` // "alarm_device_updated" 或 "alarm_cloud_updated"
	TenantID  string `json:"tenant_id"`
	DeviceID  string `json:"device_id,omitempty"` // 仅当 event_type 为 "alarm_device_updated" 时存在
	Timestamp int64  `json:"timestamp"`
}

// NotifyAlarmDeviceUpdated 通知设备报警配置已更新
func (n *ConfigNotifier) NotifyAlarmDeviceUpdated(ctx context.Context, tenantID, deviceID string) error {
	event := ConfigChangeEvent{
		EventType: "alarm_device_updated",
		TenantID:  tenantID,
		DeviceID:  deviceID,
		Timestamp: time.Now().Unix(),
	}

	streamName := "config:change:stream"
	// 使用默认配置：maxLen=1000, retentionSeconds=0（不限制保留时间）
	streamID, err := rediscommon.PublishJSONToStream(ctx, n.redisClient, streamName, event, 1000, 0)
	if err != nil {
		n.logger.Error("Failed to publish alarm_device_updated event",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish config change event: %w", err)
	}

	n.logger.Info("Published alarm_device_updated event",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// NotifyAlarmCloudUpdated 通知云端报警配置已更新
func (n *ConfigNotifier) NotifyAlarmCloudUpdated(ctx context.Context, tenantID string) error {
	event := ConfigChangeEvent{
		EventType: "alarm_cloud_updated",
		TenantID:  tenantID,
		Timestamp: time.Now().Unix(),
	}

	streamName := "config:change:stream"
	// 使用默认配置：maxLen=1000, retentionSeconds=0（不限制保留时间）
	streamID, err := rediscommon.PublishJSONToStream(ctx, n.redisClient, streamName, event, 1000, 0)
	if err != nil {
		n.logger.Error("Failed to publish alarm_cloud_updated event",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish config change event: %w", err)
	}

	n.logger.Info("Published alarm_cloud_updated event",
		zap.String("tenant_id", tenantID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// ParseConfigChangeEvent 解析配置变更事件（用于消费者）
func ParseConfigChangeEvent(data []byte) (*ConfigChangeEvent, error) {
	var event ConfigChangeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config change event: %w", err)
	}
	return &event, nil
}
