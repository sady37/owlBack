package consumer

// S1 (FU3): sensor-side alarm consumer.
//
// 订阅 iot:alarm:stream 独立 consumer group "wisefido-sensor-alarm"（与 roomengine 的
// "roomengine" group 隔离，互不影响 XAck/pending）。当前消费范围（[[sensor_stream_subscriptions]]）：
//
//   - WeakBio 4 类（WeakBiometricSignal / HeartRateAlert / RespRateAlert / ApneaHypopnea
//     含 High/Low 变体）→ aggregator WeakBio 30min 滑窗累加
//
// 暂不消费（后续 PR）：
//   - sleepace InBed/LeftBed（engine.runAlarmLoop 当前消费 radar Fall；sleepace bed event 旁路读后续）
//   - 设备类 Offline/SignalPoor（device fitness 旁路读后续）
//
// 防 loop：sensor 自家派生 alarm（platform agent slot fd00:0:fff1::/48 范围 IPv6 Producer）
// 跳过，不再进 aggregator 累加。同时兼容 roomengine engine.go 旧 hardcoded "wisefido-sensor"
// / "sensor.<name>" 命名。

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"owl-common/alarm"
	rediscommon "owl-common/redis"
	"owl-common/spatial"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// AlarmSink 解耦 service 包反向 import：service.TargetStateAggregator.PushAlarmFields 隐式满足。
// 用 primitive 形参避免 caller 需引 service.AlarmEventSnapshot 类型。
type AlarmSink interface {
	PushAlarmFields(spatialPrefix, alarmType, producer string, tsMs int64, rawValue int)
}

const (
	sensorAlarmGroup    = "wisefido-sensor-alarm"
	sensorAlarmConsumer = "consumer-sensor-alarm"
)

type AlarmConsumer struct {
	client     *redislib.Client
	sink       AlarmSink
	sensorSlot netip.Prefix
	logger     *zap.Logger
}

func NewAlarmConsumer(client *redislib.Client, sink AlarmSink, logger *zap.Logger) *AlarmConsumer {
	slot, _ := netip.ParsePrefix(spatial.SlotSensor)
	return &AlarmConsumer{
		client:     client,
		sink:       sink,
		sensorSlot: slot,
		logger:     logger,
	}
}

func (c *AlarmConsumer) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, c.client, rediscommon.StreamAlarm.Name, sensorAlarmGroup); err != nil {
		c.logger.Warn("sensor alarm: create consumer group", zap.Error(err))
	}
	go c.runLoop(ctx)
	c.logger.Info("sensor alarm consumer started",
		zap.String("stream", rediscommon.StreamAlarm.Name),
		zap.String("group", sensorAlarmGroup))
}

func (c *AlarmConsumer) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, c.client,
			rediscommon.StreamAlarm.Name, sensorAlarmGroup, sensorAlarmConsumer,
			sensorReadCount, sensorReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Warn("sensor alarm: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			c.handleRaw(m.Values)
			c.client.XAck(ctx, rediscommon.StreamAlarm.Name, sensorAlarmGroup, m.ID)
		}
	}
}

func (c *AlarmConsumer) handleRaw(raw map[string]interface{}) {
	msg, err := rediscommon.FromStreamMap(raw)
	if err != nil {
		c.logger.Warn("sensor alarm: parse", zap.Error(err))
		return
	}
	if c.isSelfProducer(msg.Producer) {
		return
	}
	if msg.SubjectEntity == "" {
		return
	}
	base, ok := normalizeWeakBioCategory(msg.Category)
	if !ok {
		return
	}
	rawValue := 0
	if base == alarm.WeakBiometricSignal {
		fields := rediscommon.FirstDataValue(msg.DataValue)
		if fields != nil {
			// firmware state×20 (0/20/40/60)；非 WeakBio 变体 rawValue 不使用
			rawValue = intFromAlarmField(fields["state"]) * 20
		}
	}
	c.sink.PushAlarmFields(msg.SubjectEntity, base, msg.Producer, msg.Timestamp, rawValue)
}

// isSelfProducer 判定 alarm 是否本平台 agent 派生：
//   - Producer ∈ fd00:0:fff1::/48 slot（owl-common/spatial.SlotSensor）
//   - 旧版 hardcoded "wisefido-sensor" / "sensor.<name>"（兼容 engine.go:1568）
func (c *AlarmConsumer) isSelfProducer(producer string) bool {
	if producer == "" {
		return false
	}
	if producer == "wisefido-sensor" || strings.HasPrefix(producer, "sensor.") {
		return true
	}
	addr, err := netip.ParseAddr(producer)
	if err != nil {
		return false
	}
	return c.sensorSlot.Contains(addr)
}

// normalizeWeakBioCategory 把 HR/RR/WeakBio/ApneaH 各 .High/.Low 变体规整到 4 类 base name，
// aggregator handleAlarmEvent 内仅识别 base name（详 [[target_state_weak_bio_signal_design]]）。
func normalizeWeakBioCategory(cat string) (string, bool) {
	switch cat {
	case alarm.WeakBiometricSignal:
		return alarm.WeakBiometricSignal, true
	case alarm.HeartRateAlert, alarm.HeartRateAlertHigh, alarm.HeartRateAlertLow:
		return alarm.HeartRateAlert, true
	case alarm.RespRateAlert, alarm.RespRateAlertHigh, alarm.RespRateAlertLow:
		return alarm.RespRateAlert, true
	case alarm.ApneaHypopnea:
		return alarm.ApneaHypopnea, true
	}
	return "", false
}

func intFromAlarmField(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}
