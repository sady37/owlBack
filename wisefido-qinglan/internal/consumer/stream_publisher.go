package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	"owl-common/card"
	rediscommon "owl-common/redis"
	"wisefido-qinglan/internal/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// QINGLAN_VERBOSE_LOG=true 时热路径用 Info / log.Printf；默认 Debug 或静默。
func isQinglanVerboseLog() bool {
	return os.Getenv("QINGLAN_VERBOSE_LOG") == "true"
}

// QinglanHotPathLog 供 subscriber 等包复用：默认 Debug，verbose 时 Info。
func QinglanHotPathLog(logger *zap.Logger, msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}
	if isQinglanVerboseLog() {
		logger.Info(msg, fields...)
		return
	}
	logger.Debug(msg, fields...)
}

// CardMappingService 定义卡片映射服务接口（避免导入循环）。
//
// device_ipv6 单程票：DeviceBaseline.DeviceAddr 是路由层主键。
type CardMappingService interface {
	GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*card.DeviceBaseline, error)
	BaselineFor(deviceUID string) (card.DeviceBaseline, bool)
}

// StreamPublisher Redis Stream 发布器。
//
// device_ipv6 单程票后所有 publish 路径只发 device_addr / subject_entity / category；
// 不再写 device_id/device_uid/tenant_id/semantic_location 等冗余字段。
type StreamPublisher struct {
	redisClient    *redis.Client
	config         *config.Config
	cardMappingSvc CardMappingService
	logger         *zap.Logger
}

// NewStreamPublisher 创建 Stream 发布器。
func NewStreamPublisher(redisClient *redis.Client, cfg *config.Config) *StreamPublisher {
	return &StreamPublisher{
		redisClient: redisClient,
		config:      cfg,
	}
}

// SetCardMappingService 设置卡片映射服务（用于查询 deviceUID → baseline）。
func (p *StreamPublisher) SetCardMappingService(cardMappingSvc CardMappingService) {
	p.cardMappingSvc = cardMappingSvc
}

// GetCardID 获取设备的 cardID（INET CIDR text）。
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

// skipQinglanIotHeadPublish — device_ipv6 单程票：subject_entity + device_addr 任一空都丢弃。
//
// device_addr 必填（路由层主键）；subject_entity 可空（unbound device 走 cardagg 反查 LPM）。
func skipQinglanIotHeadPublish(subjectEntity string, deviceAddr netip.Addr) bool {
	if !deviceAddr.IsValid() {
		return true
	}
	// subject_entity 可空（unbound device，cardagg 反查 LPM 后填）
	_ = subjectEntity
	return false
}

var errEmptySubjectEntity = fmt.Errorf("subject_entity is empty, skip publish")

func (p *StreamPublisher) publishObservation(ctx context.Context, stream rediscommon.StreamDefinition, msg *rediscommon.IoTStreamMessage) error {
	// Producer 缺省按 device-gateway 语义自动填充
	if msg.Producer == "" {
		msg.Producer = rediscommon.BuildDeviceProducer(msg.DeviceAddr)
	}
	if skipQinglanIotHeadPublish(msg.SubjectEntity, msg.DeviceAddr) {
		if p.logger != nil {
			QinglanHotPathLog(p.logger, "skip iot publish: invalid device_addr",
				zap.String("stream", stream.Name),
				zap.String("subject_entity", msg.SubjectEntity))
		}
		return nil
	}
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
	if p.logger != nil {
		streamLabel := streamLabelFrom(stream.Name, msg)
		payload, _ := json.Marshal(msg.DataValue)
		QinglanHotPathLog(p.logger, "publish to redis",
			zap.String("stream", streamLabel),
			zap.String("cid", msg.SubjectEntity),
			zap.String("device_addr", msg.DeviceAddr.String()),
			zap.Int64("ts", msg.Timestamp),
			zap.ByteString("event", payload))
	}
	maxLen, retention := p.config.GetStreamConfig(stream.Name)
	_, err := rediscommon.PublishToStream(ctx, p.redisClient, stream.Name, data, maxLen, retention)
	if err != nil && p.logger != nil {
		p.logger.Error("publish observation failed",
			zap.String("stream", stream.Name),
			zap.String("device_addr", msg.DeviceAddr.String()),
			zap.Error(err))
	}
	return err
}

// SetLogger 设置 logger
func (p *StreamPublisher) SetLogger(logger *zap.Logger) {
	p.logger = logger
}

// PublishToStream 发布数据到 Redis Stream
func (p *StreamPublisher) PublishToStream(ctx context.Context, streamName string, data map[string]interface{}) (string, error) {
	maxLen, retentionSeconds := p.config.GetStreamConfig(streamName)
	streamID, err := rediscommon.PublishToStream(ctx, p.redisClient, streamName, data, maxLen, retentionSeconds)
	if err != nil {
		return "", fmt.Errorf("failed to publish to stream %s: %w", streamName, err)
	}
	return streamID, nil
}

// BuildEncodedData 构建 iot:*:stream 输出数据（顶层无 addressInfo，键 dataValue 为数组）。
//
// device_ipv6 单程票：addr 是 device 路由层主键；subjectEntity 已绑卡填 cardID，未绑卡留空。
func (p *StreamPublisher) BuildEncodedData(
	addr netip.Addr,
	subjectEntity string,
	topicType, category string,
	dataValue []interface{},
) map[string]interface{} {
	msg := rediscommon.BuildIoTStreamMessage(
		addr,
		DeviceTypeRadar, // wisefido-qinglan 网关的所有设备都是 Radar
		subjectEntity,
		time.Now().Unix(),
		topicType,
		category,
		dataValue,
	)
	return iotStreamMessageToMap(msg)
}

// iotStreamMessageToMap 将 IoTStreamMessage 转换为 map（用于发布到 Redis Stream）。
func iotStreamMessageToMap(msg rediscommon.IoTStreamMessage) map[string]interface{} {
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

// PublishDeviceStatus 发布设备状态到 iot:event:stream（device 属于 event）。
//
// device_ipv6 单程票：addr 是路由主键；deviceUID 仅作 cardID 反查用。
// subject 已绑=card_id；未绑=空（cardagg IotPreparedHandler LPM 反查兜底）。
func (p *StreamPublisher) PublishDeviceStatus(
	ctx context.Context,
	addr netip.Addr,
	deviceUID, deviceType string,
	statuses map[string]int,
) error {
	cardID := ""
	if p.cardMappingSvc != nil && deviceUID != "" {
		if cdi, err := p.cardMappingSvc.GetCardIDByDeviceUID(ctx, deviceUID); err == nil && cdi != nil {
			cardID = cdi.CardID
		}
	}
	msg := rediscommon.BuildDeviceStatusMessage(
		addr,
		cardID, // unbound device → 空，cardagg 反查 (R-009 不再 deviceID 占位)
		deviceType,
		time.Now().UnixMilli(),
		statuses,
	)

	if skipQinglanIotHeadPublish(msg.SubjectEntity, msg.DeviceAddr) {
		if p.logger != nil {
			QinglanHotPathLog(p.logger, "skip PublishDeviceStatus: invalid device_addr",
				zap.String("device_addr", addr.String()))
		}
		return nil
	}

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
