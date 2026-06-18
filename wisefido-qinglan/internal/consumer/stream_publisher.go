package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"
	"sync/atomic"
	"time"

	"owl-common/alarm"
	rediscommon "owl-common/redis"

	"wisefido-qinglan/internal/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)


// StreamPublisher Redis Stream 发布器。
//
// 后所有 publish 路径只发 device_addr / category；subject_entity 永远空，
// 由 cardagg IotPreparedHandler 按 device_addr LPM 反查解析（R-009 单源真相）。
//
// seqCounter: 协议层北极星 (TDPv2 envelope sequence_number) — publisher 内单调 monotonic uint64，
// 每条消息自增；下游 sensor 派生 verdict 时把"触发的 source msg seq" 写入 evidence
// （reasoning trace 链 = qinglan envelope.seq → sensor verdict.evidence.trigger_seq_num）。
// EnablementGate qinglan 端 alarm 使能 gate；service.AlarmEnablementCache 包装 closure 实现。
// 未启用返回 false → PublishAlarm 静默 drop。nil 时退化 always-allow（boot 期兼容）。
type EnablementGate func(ctx context.Context, deviceAddr, alarmType string) bool

type StreamPublisher struct {
	redisClient *redis.Client
	config      *config.Config
	logger      *zap.Logger
	seqCounter  atomic.Uint64
	alarmGate   EnablementGate
}

// NewStreamPublisher 创建 Stream 发布器。
func NewStreamPublisher(redisClient *redis.Client, cfg *config.Config) *StreamPublisher {
	return &StreamPublisher{
		redisClient: redisClient,
		config:      cfg,
	}
}

// SetAlarmGate main wiring 注入 enablement gate；nil 退化 always-allow。
// 把 alarm enablement 检查收到 producer 端：未启用直接 drop，cardagg 端不重复 gate
// （符合 [[feedback_producer_first]] + [[platform_agent_addressing]] trust 模式）。
func (p *StreamPublisher) SetAlarmGate(gate EnablementGate) {
	p.alarmGate = gate
}

// PublishMonitor sends a redis.StreamMessage to iot:monitor:stream.
func (p *StreamPublisher) PublishMonitor(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publishObservation(ctx, rediscommon.StreamMonitor, msg)
}

// PublishEvent sends a redis.StreamMessage to iot:event:stream.
func (p *StreamPublisher) PublishEvent(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publishObservation(ctx, rediscommon.StreamEvent, msg)
}

// alarmGateSkip 设备类 alarm（HIPAA 强制审计）跳过 enablement gate — 客户禁用了也强制存档。
// 与 cardagg alarm_router.go 的 deviceClassOnset + recoveryMap 一致；Fall 走 event 流不在此处。
var alarmGateSkip = map[string]struct{}{
	alarm.AlarmTypeOffline:        {},
	alarm.AlarmTypeOfflineRecover: {},
	alarm.AlarmTypeDeviceFailure:  {},
	alarm.AlarmTypeDeviceRecover:  {},
	alarm.SensorDetached:          {},
	alarm.SensorDetachedRecover:   {},
	alarm.SignalPoor:              {},
	alarm.SingalPoorRecover:       {}, // sic: 'Singal' 是 owl-common/alarm 常量原拼写
	alarm.AngleException:          {},
	alarm.AngleExceptionRecover:   {},
}

// PublishAlarm sends a redis.StreamMessage to iot:alarm:stream.
//
// 源头 enablement gate（仅对 vital 类生效）：
//   - vital (HR/RR/Apnea/Weak)：查 device_config alarm.device_config，未启用 drop
//   - device-class (Offline/SensorDetached/SignalPoor/AngleException + Recovers)：跳 gate，HIPAA 强制审计
//
// LeftBed / Stay / Fall / SittingOnGround 不走 qinglan PublishAlarm — 那些由 sensor 派生（zonealarm 或
// roomengine 路径，从 event 流转化），各自源头自己 gate。
func (p *StreamPublisher) PublishAlarm(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	if p.alarmGate != nil && msg != nil && msg.DeviceAddr.IsValid() && msg.Category != "" {
		if _, skip := alarmGateSkip[msg.Category]; !skip {
			if !p.alarmGate(ctx, msg.DeviceAddr.String(), msg.Category) {
				if p.logger != nil {
					p.logger.Debug("publish alarm gated (disabled in device_config)",
						zap.String("device_addr", msg.DeviceAddr.String()),
						zap.String("category", msg.Category))
				}
				return nil
			}
		}
	}
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

// skipQinglanIotHeadPublish — subject_entity + device_addr 任一空都丢弃。
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
	// SequenceNumber 缺省自增（北极星：reasoning trace 链 = envelope.seq → 下游 verdict.evidence.trigger_seq_num）
	if msg.SequenceNumber == 0 {
		msg.SequenceNumber = p.seqCounter.Add(1)
	}
	if skipQinglanIotHeadPublish(msg.SubjectEntity, msg.DeviceAddr) {
		if p.logger != nil {
			p.logger.Debug("skip iot publish: invalid device_addr",
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
	traceID := fmt.Sprintf("%s.%d", msg.Producer, msg.SequenceNumber)
	if p.logger != nil {
		streamLabel := streamLabelFrom(stream.Name, msg)
		payload, _ := json.Marshal(msg.DataValue)
		p.logger.Debug("publish to redis",
			zap.String("trace_id", traceID),
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
			zap.String("trace_id", traceID),
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
// addr 是 device 路由层主键；subjectEntity 已绑卡填 cardID，未绑卡留空。
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

// PublishDeviceStatus 发布设备状态到 iot:event:stream（device 属于 event）。
//
// deviceStatus 是设备自身的 raw 事件，按 envelope 约定 SubjectEntity = device_uid（事件主体 =
// 观测设备本身）。cardagg 路由仍走 device_addr → cardID LPM，不依赖 SubjectEntity；填 uid
// 让 event_log.device_uid 落稳定身份锚（device_addr 随 rebind 变，反查不回历史设备）。
func (p *StreamPublisher) PublishDeviceStatus(
	ctx context.Context,
	addr netip.Addr,
	deviceUID, deviceType string,
	statuses map[string]int,
) error {
	msg := rediscommon.BuildDeviceStatusMessage(
		addr,
		deviceUID,
		deviceType,
		time.Now().UnixMilli(),
		statuses,
	)

	if skipQinglanIotHeadPublish(msg.SubjectEntity, msg.DeviceAddr) {
		if p.logger != nil {
			p.logger.Debug("skip PublishDeviceStatus: invalid device_addr",
				zap.String("device_addr", addr.String()))
		}
		return nil
	}

	eventData := iotStreamMessageToMap(msg)
	_, err := p.PublishToStream(ctx, rediscommon.StreamEvent.Name, eventData)
	return err
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
