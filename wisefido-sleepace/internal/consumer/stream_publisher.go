package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"owl-common/card"
	rediscommon "owl-common/redis"

	"wisefido-sleepace/internal/config"
	"wisefido-sleepace/internal/service"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var errEmptyCardID = fmt.Errorf("card_id is empty, message dropped")

func isSleepaceVerboseLog() bool {
	return os.Getenv("SLEEPACE_VERBOSE_LOG") == "true"
}

type StreamPublisher struct {
	redisClient    *redis.Client
	config         *config.Config
	cardMappingSvc *service.CardMappingService
	logger         *zap.Logger
	// seqCounter: TDPv2 envelope sequence_number — publisher 内单调；下游 verdict.evidence.trigger_seq_num 链
	seqCounter atomic.Uint64
}

func NewStreamPublisher(redisClient *redis.Client, cfg *config.Config, logger *zap.Logger) *StreamPublisher {
	return &StreamPublisher{
		redisClient: redisClient,
		config:      cfg,
		logger:      logger,
	}
}

func (p *StreamPublisher) SetCardMappingService(svc *service.CardMappingService) {
	p.cardMappingSvc = svc
}

// ResolveToDeviceUID 将 device_code 或 device_id 转为 device_uid，供网关内部统一使用。
func (p *StreamPublisher) ResolveToDeviceUID(ctx context.Context, id string) string {
	if p.cardMappingSvc == nil || id == "" {
		return id
	}
	return p.cardMappingSvc.ResolveToDeviceUID(ctx, id)
}

// Resolve 把 sleepace 厂家 MQTT 报文里的 deviceId（可能是 device_uid 或 device_code）翻译成 owl 内部
// 设备身份 DeviceBaseline（tenant/branch/unit + device_uid/code/type + access + device_addr）。
//
// 始终返回非 nil（未命中时 stub: DeviceUID=deviceKey, 其它字段 0 值；caller 用 access==false 过滤）。
// 不再返回 v1 deviceID(UUID)——identity 一刀切走 device_uid (logMAC)。
// envelope SubjectEntity 由 caller 直接传 info.DeviceUID（CloudEvents subject = 观测设备）。
// cardagg 需要 cardID 走自己 device_addr → card LPM（cards 表归 cardagg 自家域，不污染上游）。
func (p *StreamPublisher) Resolve(ctx context.Context, deviceKey string) *card.DeviceBaseline {
	if p.cardMappingSvc == nil {
		return &card.DeviceBaseline{DeviceUID: deviceKey}
	}
	info, err := p.cardMappingSvc.GetBaseline(ctx, deviceKey)
	if err != nil || info == nil {
		p.logger.Debug("card lookup miss", zap.String("device_key", deviceKey), zap.Error(err))
		return &card.DeviceBaseline{DeviceUID: deviceKey}
	}
	return info
}

// PublishMonitor sends an IoTStreamMessage to iot:monitor:stream.
func (p *StreamPublisher) PublishMonitor(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publish(ctx, rediscommon.StreamMonitor, msg)
}

// PublishEvent sends an IoTStreamMessage to iot:event:stream.
func (p *StreamPublisher) PublishEvent(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publish(ctx, rediscommon.StreamEvent, msg)
}

// PublishAlarm sends an IoTStreamMessage to iot:alarm:stream.
func (p *StreamPublisher) PublishAlarm(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publish(ctx, rediscommon.StreamAlarm, msg)
}

func (p *StreamPublisher) publish(ctx context.Context, stream rediscommon.StreamDefinition, msg *rediscommon.IoTStreamMessage) error {
	if msg.Producer == "" {
		msg.Producer = rediscommon.BuildDeviceProducer(msg.DeviceAddr)
	}
	// SequenceNumber 缺省自增（北极星 reasoning trace 链；详 qinglan stream_publisher 同段注释）
	if msg.SequenceNumber == 0 {
		msg.SequenceNumber = p.seqCounter.Add(1)
	}
	// device_ipv6 单程票 R-009：subject_entity 可空（unbound device，cardagg IotPreparedHandler
	// 用 LPM 反查 cards 兜底）；只要 device_addr 有效就发。
	if !msg.DeviceAddr.IsValid() {
		p.logger.Error("device_addr is invalid, message dropped",
			zap.String("stream", stream.Name),
			zap.String("subject_entity", msg.SubjectEntity))
		return fmt.Errorf("device_addr invalid")
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	if p.logger != nil && isSleepaceVerboseLog() {
		payload, _ := json.Marshal(msg.DataValue)
		p.logger.Debug("publish to redis",
			zap.String("stream", stream.Name),
			zap.String("subject_entity", msg.SubjectEntity),
			zap.String("device_addr", msg.DeviceAddr.String()),
			zap.Int64("ts", msg.Timestamp),
			zap.ByteString("event", payload))
	}
	data := msg.ToStreamMap()
	maxLen, retention := p.config.GetStreamConfig(stream.Name)
	_, err := rediscommon.PublishToStream(ctx, p.redisClient, stream.Name, data, maxLen, retention)
	if err != nil {
		p.logger.Error("publish failed",
			zap.String("stream", stream.Name),
			zap.String("device_addr", msg.DeviceAddr.String()),
			zap.Error(err))
	}
	return err
}
