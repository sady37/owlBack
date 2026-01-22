package notifier

import (
	"context"
	"encoding/json"
	"fmt"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigNotifier 配置变更通知器
// 负责发布配置变更事件到 config:config:stream（非设备发出的 redis stream 格式：config:xxx:stream）
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

// ConfigChangeEvent 配置变更事件（兼容旧格式，用于解析）
type ConfigChangeEvent struct {
	EventType string `json:"event_type"` // "alarm_device_updated"（alarm_cloud_updated 已废弃）
	TenantID  string `json:"tenant_id"`
	DeviceID  string `json:"device_id,omitempty"`  // 仅当 event_type 为 "alarm_device_updated" 时存在
	DeviceUID string `json:"device_uid,omitempty"` // 仅当 event_type 为 "alarm_device_updated" 时存在（radar 服务需要）
	Timestamp int64  `json:"timestamp"`
}

// NotifyAlarmDeviceUpdated 通知设备报警配置已更新
// 使用统一的 config:device_status:stream（CloudEvents 格式）
// locationInfo 可选，可为空
func (n *ConfigNotifier) NotifyAlarmDeviceUpdated(ctx context.Context, tenantID, deviceID, deviceUID, deviceCode, deviceType string, locationInfo interface{}) error {
	// 转换 locationInfo 为 rediscommon.LocationInfo（可选，可为空）
	var locInfo *rediscommon.LocationInfo
	if locationInfo != nil {
		if loc, ok := locationInfo.(*rediscommon.LocationInfo); ok {
			locInfo = loc
		}
	}

	// 构建 CloudEvents 格式的消息
	message := rediscommon.BuildAlarmDeviceMessage("wisefido-data", tenantID, deviceID, deviceUID, deviceCode, deviceType, locInfo)

	// 转换为 map（用于发布到 Redis Stream，CloudEvents 标准格式）
	data := configChangeMessageToMap(message)

	// 发布到统一的 config:device_status:stream
	streamName := rediscommon.StreamConfigDeviceStatus.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigDeviceStatus, nil)
	
	// 使用 PublishToStream 发布（展开字段格式）
	streamID, err := rediscommon.PublishToStream(ctx, n.redisClient, streamName, data, maxLen, retentionSeconds)
	if err != nil {
		n.logger.Error("Failed to publish alarm_device_updated event",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish config change event: %w", err)
	}

	n.logger.Info("Published alarm_device_updated event",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
		zap.String("event_type", message.Type),
		zap.String("event_id", message.ID),
	)

	return nil
}

// configChangeMessageToMap 将 ConfigChangeMessage 转换为 map（用于发布到 Redis Stream）
// 使用 CloudEvents 标准格式
func configChangeMessageToMap(msg rediscommon.ConfigChangeMessage) map[string]interface{} {
	result := make(map[string]interface{})

	// CloudEvents 标准字段
	result["specversion"] = msg.SpecVersion
	result["id"] = msg.ID
	result["source"] = msg.Source
	result["type"] = msg.Type
	result["time"] = msg.Time

	// data 为 JSON 字符串
	dataJSON, _ := json.Marshal(msg.Data)
	result["data"] = string(dataJSON)

	return result
}

// ParseConfigChangeEvent 解析配置变更事件（用于消费者）
func ParseConfigChangeEvent(data []byte) (*ConfigChangeEvent, error) {
	var event ConfigChangeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config change event: %w", err)
	}
	return &event, nil
}
