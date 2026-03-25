package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"owl-common/card"
	rediscommon "owl-common/redis"
	"wisefido-qinglan/internal/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CardMappingService 定义卡片映射服务接口（避免导入循环）
type CardMappingService interface {
	GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*card.DeviceBaseline, error)
	BaselineFor(deviceUID string) (card.DeviceBaseline, bool)
}

// StreamPublisher Redis Stream发布器
type StreamPublisher struct {
	redisClient    *redis.Client
	config         *config.Config
	cardMappingSvc CardMappingService
	logger         *zap.Logger
}

// NewStreamPublisher 创建Stream发布器
func NewStreamPublisher(redisClient *redis.Client, cfg *config.Config) *StreamPublisher {
	return &StreamPublisher{
		redisClient: redisClient,
		config:      cfg,
	}
}

// SetCardMappingService 设置卡片映射服务（用于查询 deviceUID → cardID）
func (p *StreamPublisher) SetCardMappingService(cardMappingSvc CardMappingService) {
	p.cardMappingSvc = cardMappingSvc
}

// GetCardID 获取设备的 cardID
func (p *StreamPublisher) GetCardID(ctx context.Context, deviceUID string) string {
	if p.cardMappingSvc != nil && deviceUID != "" {
		if cdi, err := p.cardMappingSvc.GetCardIDByDeviceUID(ctx, deviceUID); err == nil && cdi != nil {
			return cdi.CardID
		}
	}
	return ""
}

// PublishMonitor sends a redis.StreamMessage to iot:monitor:stream.
func (p *StreamPublisher) PublishMonitor(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publishObservation(ctx, rediscommon.StreamMonitor, msg)
}

// PublishStat sends a redis.StreamMessage to iot:stat:stream.
func (p *StreamPublisher) PublishStat(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publishObservation(ctx, rediscommon.StreamStat, msg)
}

// PublishEvent sends a redis.StreamMessage to iot:event:stream.
func (p *StreamPublisher) PublishEvent(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publishObservation(ctx, rediscommon.StreamEvent, msg)
}

// PublishAlarm sends a redis.StreamMessage to iot:alarm:stream.
func (p *StreamPublisher) PublishAlarm(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publishObservation(ctx, rediscommon.StreamAlarm, msg)
}

// streamLabelFrom 生成日志用 stream 标签 iot:xxx:yyy，xxx=event/alarm/monitor，yyy=category
func streamLabelFrom(streamName string, msg *rediscommon.IoTStreamMessage) string {
	parts := strings.SplitN(streamName, ":", 3)
	xxx := "stream"
	if len(parts) >= 2 {
		xxx = parts[1]
	}
	yyy := msg.Category
	if yyy == "" && len(msg.DataValue) > 0 {
		if m, ok := msg.DataValue[0].(map[string]interface{}); ok {
			if v, ok := m["dataCategory"].(string); ok && v != "" {
				yyy = v
			}
		}
	}
	if yyy == "" {
		yyy = "stream"
	}
	return "iot:" + xxx + ":" + yyy
}

func (p *StreamPublisher) publishObservation(ctx context.Context, stream rediscommon.StreamDefinition, msg *rediscommon.IoTStreamMessage) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	// 从 stream 名推导 topic_type（iot:monitor:stream -> monitor）
	if msg.TopicType == "" && len(stream.Name) > 0 {
		parts := strings.SplitN(stream.Name, ":", 3)
		if len(parts) >= 2 {
			msg.TopicType = parts[1]
		}
	}
	data := msg.ToStreamMap()
	if p.logger != nil && stream.Name != rediscommon.StreamMonitor.Name {
		// 日志格式 iot:xxx:yyy，xxx=event/alarm/monitor，yyy=category
		streamLabel := streamLabelFrom(stream.Name, msg)
		payload, _ := json.Marshal(msg.DataValue)
		p.logger.Info("publish to redis",
			zap.String("stream", streamLabel),
			zap.String("cid", msg.CardID),
			zap.String("device_uid", msg.DeviceUID),
			zap.Int64("ts", msg.Timestamp),
			zap.ByteString("event", payload))
	}
	maxLen, retention := p.config.GetStreamConfig(stream.Name)
	_, err := rediscommon.PublishToStream(ctx, p.redisClient, stream.Name, data, maxLen, retention)
	if err != nil && p.logger != nil {
		p.logger.Error("publish observation failed",
			zap.String("stream", stream.Name),
			zap.String("device_uid", msg.DeviceUID),
			zap.Error(err))
	}
	return err
}

// SetLogger 设置 logger
func (p *StreamPublisher) SetLogger(logger *zap.Logger) {
	p.logger = logger
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

// BuildEncodedData 构建 iot:*:stream 输出数据（顶层无 addressInfo，键 dataValue 为数组）
// cardID 未绑卡可传空；dataValue 为数组，每项含 category 及对应字段。
func (p *StreamPublisher) BuildEncodedData(
	cardID string,
	tenantID string,
	deviceID string,
	topicType, category string,
	dataValue []interface{},
) map[string]interface{} {
	msg := rediscommon.BuildIoTStreamMessage(
		deviceID,
		DeviceTypeRadar, // wisefido-qinglan网关的所有设备都是Radar（常量在mqtt_consumer.go中定义）
		cardID,
		tenantID,
		time.Now().Unix(),
		topicType,
		category,
		dataValue,
	)
	return iotStreamMessageToMap(msg)
}

// iotStreamMessageToMap 将 IoTStreamMessage 转换为 map（用于发布到 Redis Stream）
// 顶层顺序：device_id, device_type, card_id, tenant_id, timestamp, topic_type, category, dataValue（无 addressInfo）
func iotStreamMessageToMap(msg rediscommon.IoTStreamMessage) map[string]interface{} {
	// 统一复用 owl-common 的标准序列化，避免手写字段漏传（例如 device_uid）。
	m := msg
	return (&m).ToStreamMap()
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

// PublishDeviceStatus 发布设备状态到 iot:event:stream（device 属于 event）
// deviceUID: 设备UID（用于查询关联的 cardID）
// statuses: 设备状态 map[string]int，支持字段：
//   - online: 0=离线, 1=在线
//   - angle_abnormal: 0=正常, 1=异常
//   - signal_poor: 0=正常, 1=信号差
//   - detached: 0=正常, 1=脱落
//   - device_failure: 0=正常, 1=故障
func (p *StreamPublisher) PublishDeviceStatus(
	ctx context.Context,
	deviceID, deviceType, tenantID, deviceUID string,
	statuses map[string]int,
) error {
	// 查询 cardID（如果提供了 cardMappingSvc）
	cardID := ""
	if p.cardMappingSvc != nil && deviceUID != "" {
		if cdi, err := p.cardMappingSvc.GetCardIDByDeviceUID(ctx, deviceUID); err == nil && cdi != nil {
			cardID = cdi.CardID
			if p.logger != nil {
				p.logger.Debug("Found cardID for device",
					zap.String("device_uid", deviceUID),
					zap.String("card_id", cardID),
					zap.String("device_id", deviceID))
			}
		} else if p.logger != nil {
			p.logger.Debug("Unable to find cardID for device",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
				zap.Error(err))
		}
	}

	msg := rediscommon.BuildDeviceStatusMessage(
		deviceUID,
		deviceType,
		cardID, // 填充查询到的 cardID
		tenantID,
		time.Now().Unix(),
		statuses,
	)
	msg.DeviceID = deviceID

	eventData := iotStreamMessageToMap(msg)

	_, err := p.PublishToStream(ctx, rediscommon.StreamEvent.Name, eventData)
	return err
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
