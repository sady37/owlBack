package consumer

// S1 (FU3): sensor-side alarm consumer.
//
// 订阅 iot:alarm:stream 独立 consumer group "wisefido-sensor-alarm"（与 roomengine 的
// "roomengine" group 隔离，互不影响 XAck/pending）。当前 fan-out（[[sensor_stream_subscriptions]]）：
//
//   1. WeakBio 4 类 → aggregator WeakBio 30min 滑窗累加
//      WeakBiometricSignal / HeartRateAlert / RespRateAlert / ApneaHypopnea (含 High/Low 变体)
//
//   2. 设备类 4 类 → DeviceFitnessTracker.MarkUnfit；对应 Recover → ClearReason
//      Offline / SensorDetached / AngleException / SignalPoor + Recover 配对
//      adapter_sleepace / adapter_radar handleMsg 入口查 IsFit → unfit skip 防垃圾数据污染 FSM
//
// 不消费（设计上忽略）：
//   - sleepace InBed/LeftBed（zoneengine adapter_sleepace 自己直接订阅 iot:alarm 走另一路）
//   - radar Fall（roomengine.runAlarmLoop / runEventLoop 自己消费走 fall verifier）
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

// FitnessSink 解耦：service.DeviceFitnessTracker 隐式满足。
// reason 用 uint8（避免 caller 反向 import service.FitnessReason 类型）；
// 调用方用本文件内 FitnessReason* 常量传入，service 端语义一致。
type FitnessSink interface {
	MarkUnfit(deviceAddr string, reason uint8)
	ClearReason(deviceAddr string, reason uint8)
}

// 与 service.FitnessReason* 同步（位标志）；本 file 持镜像常量供 fan-out 用。
const (
	fitnessReasonOffline         uint8 = 1 << 0
	fitnessReasonSensorDetached  uint8 = 1 << 1
	fitnessReasonAngleException  uint8 = 1 << 2
	fitnessReasonSignalPoor      uint8 = 1 << 3
)

const (
	sensorAlarmGroup    = "wisefido-sensor-alarm"
	sensorAlarmConsumer = "consumer-sensor-alarm"
)

type AlarmConsumer struct {
	client      *redislib.Client
	sink        AlarmSink
	fitnessSink FitnessSink // 可空：未注入时设备类 alarm 不影响 fitness
	sensorSlot  netip.Prefix
	logger      *zap.Logger
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

// SetFitnessSink main wiring 注入设备健康 sink；不注入时 4 类设备 alarm 不影响 sensor engine。
func (c *AlarmConsumer) SetFitnessSink(s FitnessSink) {
	if c == nil {
		return
	}
	c.fitnessSink = s
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

	// Fan-out 1: 设备类 alarm → DeviceFitnessTracker（防垃圾数据污染 sensor zone engine）
	if reason, isUnfit, isRecover := classifyDeviceFitnessAlarm(msg.Category); reason != 0 && c.fitnessSink != nil && msg.DeviceAddr.IsValid() {
		deviceAddr := msg.DeviceAddr.String()
		switch {
		case isUnfit:
			c.fitnessSink.MarkUnfit(deviceAddr, reason)
		case isRecover:
			c.fitnessSink.ClearReason(deviceAddr, reason)
		}
		return // 设备类不入 WeakBio 路径
	}

	// Fan-out 2: WeakBio 4 类 → aggregator score 累加
	if !msg.DeviceAddr.IsValid() {
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
	// WeakBio 是 per-sleepad 信号 → aggregator key 用 /96 bed spatial prefix
	// （raw envelope SubjectEntity = device_uid，sensor 内部按物理 spatial 索引）。
	bedPrefix := netip.PrefixFrom(msg.DeviceAddr, 96).Masked().String()
	c.sink.PushAlarmFields(bedPrefix, base, msg.Producer, msg.Timestamp, rawValue)
}

// classifyDeviceFitnessAlarm 把 alarm category 映射到 fitness reason flag。
// 返回 (reason, isUnfit, isRecover)；reason=0 表示非设备类 alarm。
func classifyDeviceFitnessAlarm(category string) (reason uint8, isUnfit bool, isRecover bool) {
	switch category {
	case alarm.AlarmTypeOffline:
		return fitnessReasonOffline, true, false
	case alarm.AlarmTypeOfflineRecover:
		return fitnessReasonOffline, false, true
	case alarm.SensorDetached:
		return fitnessReasonSensorDetached, true, false
	case alarm.SensorDetachedRecover:
		return fitnessReasonSensorDetached, false, true
	case alarm.AngleException:
		return fitnessReasonAngleException, true, false
	case alarm.AngleExceptionRecover:
		return fitnessReasonAngleException, false, true
	case alarm.SignalPoor:
		return fitnessReasonSignalPoor, true, false
	case alarm.SingalPoorRecover: // owl-common 拼写历史保留
		return fitnessReasonSignalPoor, false, true
	}
	return 0, false, false
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
