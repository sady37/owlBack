package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/domain"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
)

// StreamPublisher Redis Stream发布器
type StreamPublisher struct {
	redisClient *redis.Client
	config      *config.Config
}

// NewStreamPublisher 创建Stream发布器
func NewStreamPublisher(redisClient *redis.Client, cfg *config.Config) *StreamPublisher {
	return &StreamPublisher{
		redisClient: redisClient,
		config:      cfg,
	}
}

// PublishToStream 发布数据到Redis Stream
// 使用 owl-common/redis.PublishToStream，直接展开字段到 Redis Stream
func (p *StreamPublisher) PublishToStream(ctx context.Context, streamName string, data map[string]interface{}) (string, error) {
	// 获取Stream配置（兼容旧方式，从配置中获取）
	maxLen, retentionSeconds := p.config.GetStreamConfig(streamName)

	// 使用 owl-common/redis.PublishToStream，直接展开字段
	streamID, err := rediscommon.PublishToStream(ctx, p.redisClient, streamName, data, maxLen, retentionSeconds)
	if err != nil {
		return "", fmt.Errorf("failed to publish to stream %s: %w", streamName, err)
	}

	return streamID, nil
}

// BuildEncodedData 构建包含完整位置信息的输出数据
// 使用 owl-common/redis 中的 IoTStreamMessage 类型定义
// 注意：device_id 始终来自 device_store，即使设备还没有在 devices 表中创建记录
func (p *StreamPublisher) BuildEncodedData(
	device *domain.Device,
	locationInfo *domain.DeviceLocationInfo,
	topicType, category string,
	data map[string]interface{},
) map[string]interface{} {
	// 优先使用 locationInfo 中的 device_id（来自 device_store，权威来源）
	// 如果 locationInfo 为 nil，使用 device 中的 device_id
	var deviceID, deviceUID, tenantID string
	var deviceType string = "Radar"

	if locationInfo != nil {
		deviceID = locationInfo.DeviceID
		deviceUID = locationInfo.DeviceUID
		tenantID = locationInfo.TenantID
		if locationInfo.DeviceType.Valid {
			deviceType = locationInfo.DeviceType.String
		}
	} else if device != nil {
		deviceID = device.DeviceID
		deviceUID = device.DeviceUID
		tenantID = device.TenantID
		if device.DeviceType.Valid {
			deviceType = device.DeviceType.String
		}
	} else {
		// 如果 device 和 locationInfo 都为 nil，无法构建数据
		// 这不应该发生，但为了安全起见，返回空 map
		return map[string]interface{}{
			"error": "device and locationInfo are both nil",
		}
	}

	// 转换位置信息（只在 event/alarm 中包含）
	var locInfo *rediscommon.LocationInfo
	if (topicType == "event" || topicType == "alarm") && locationInfo != nil {
		locInfo = &rediscommon.LocationInfo{}
		if locationInfo.BranchID.Valid && locationInfo.BranchID.String != "" {
			locInfo.BranchID = &locationInfo.BranchID.String
		}
		if locationInfo.BuildingID.Valid && locationInfo.BuildingID.String != "" {
			locInfo.BuildingID = &locationInfo.BuildingID.String
		}
		if locationInfo.UnitID.Valid && locationInfo.UnitID.String != "" {
			locInfo.UnitID = &locationInfo.UnitID.String
		}
		if locationInfo.RoomID.Valid && locationInfo.RoomID.String != "" {
			locInfo.RoomID = &locationInfo.RoomID.String
		}
		if locationInfo.BedID.Valid && locationInfo.BedID.String != "" {
			locInfo.BedID = &locationInfo.BedID.String
		}
	}

	// 使用 owl-common/redis 中的统一构建函数
	msg := rediscommon.BuildIoTStreamMessage(
		deviceID,
		deviceUID,
		deviceType,
		tenantID,
		time.Now().Unix(),
		topicType,
		category,
		data,
		locInfo,
	)

	// 转换为 map（用于发布到 Redis Stream）
	return iotStreamMessageToMap(msg)
}

// iotStreamMessageToMap 将 IoTStreamMessage 转换为 map（用于发布到 Redis Stream）
// 字段顺序与 IoTStreamMessage 定义一致：device_id, device_uid, device_type, tenant_id, timestamp, topic_type, category, data_value, LocationInfo
func iotStreamMessageToMap(msg rediscommon.IoTStreamMessage) map[string]interface{} {
	// 序列化 data_value 为 JSON 字符串（PublishToStream 需要字符串值）
	dataValueJSON, _ := json.Marshal(msg.DataValue)

	result := make(map[string]interface{})

	// 按定义顺序添加字段
	if msg.DeviceID != "" {
		result["device_id"] = msg.DeviceID
	}
	result["device_uid"] = msg.DeviceUID
	result["device_type"] = msg.DeviceType
	result["tenant_id"] = msg.TenantID
	result["timestamp"] = fmt.Sprintf("%d", msg.Timestamp)
	result["topic_type"] = msg.TopicType
	result["category"] = msg.Category
	result["data_value"] = string(dataValueJSON)

	// 位置信息字段（可选，只在 event/alarm 中包含）
	// 只有当指针不为 nil 且值不为空字符串时才添加（与 omitempty 行为一致）
	if msg.BranchID != nil && *msg.BranchID != "" {
		result["branch_id"] = *msg.BranchID
	}
	if msg.BuildingID != nil && *msg.BuildingID != "" {
		result["building_id"] = *msg.BuildingID
	}
	if msg.UnitID != nil && *msg.UnitID != "" {
		result["unit_id"] = *msg.UnitID
	}
	if msg.RoomID != nil && *msg.RoomID != "" {
		result["room_id"] = *msg.RoomID
	}
	if msg.BedID != nil && *msg.BedID != "" {
		result["bed_id"] = *msg.BedID
	}

	return result
}

// GetOutputStreamName 获取输出流名称
func (p *StreamPublisher) GetOutputStreamName(topicType string) string {
	switch topicType {
	case "monitor":
		return rediscommon.StreamMonitor.Name
	case "stat":
		return rediscommon.StreamStat.Name
	case "event":
		return rediscommon.StreamEvent.Name
	case "alarm":
		return rediscommon.StreamAlarm.Name
	default:
		return rediscommon.StreamOther.Name
	}
}

// GetOutputStreamConfig 获取输出流配置
func (p *StreamPublisher) GetOutputStreamConfig(topicType string) (maxLen int64, retentionSeconds int) {
	var stream rediscommon.StreamDefinition
	switch topicType {
	case "monitor":
		stream = rediscommon.StreamMonitor
	case "stat":
		stream = rediscommon.StreamStat
	case "event":
		stream = rediscommon.StreamEvent
	case "alarm":
		stream = rediscommon.StreamAlarm
	default:
		stream = rediscommon.StreamOther
	}
	return rediscommon.GetStreamConfig(stream, &p.config.Streams)
}

// StoreCommandResponse 存储命令响应
func (p *StreamPublisher) StoreCommandResponse(ctx context.Context, requestID string, response interface{}) error {
	key := fmt.Sprintf("cmd:response:%s", requestID)
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return p.redisClient.Set(ctx, key, jsonData, 5*time.Minute).Err()
}

// GetCommandResponse 获取命令响应
func (p *StreamPublisher) GetCommandResponse(ctx context.Context, requestID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("cmd:response:%s", requestID)
	jsonData, err := p.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		return nil, err
	}

	return response, nil
}
