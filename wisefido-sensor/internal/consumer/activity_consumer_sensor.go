package consumer

// S2 (FU4): sensor-side activity event consumer.
//
// 订阅 iot:event:stream 独立 consumer group "wisefido-sensor-activity"，filter
// category=activity（[[radar_activity_stats_on_event_stream]] §"radar 分钟级活动统计字段流向"）。
// 解析 walk_distance / walk_duration / stand_duration / multi_person_duration 字段
// 翻译成 service.EventFrame 后 push 给 aggregator。
//
// **不订阅其他 event category**：Fall/SitOnGround/EnterRoom/ExitRoom/InBed/LeftBed/NumberPeople/Pose
// 由 roomengine 自己的 event loop 消费（engine.go runEventLoop，组名 "roomengine"，与本 consumer
// group 隔离互不影响）。aggregator 只关心分钟级 activity 统计派生量。
//
// 心跳契约：firmware 每分钟一发 activity stat（既有事实），aggregator 自然每分钟 dirty；
// 下游 cardagg Standing 2min staleness 阈值天然成立——不另造心跳。
//
// 防 loop：与 alarm_consumer_sensor.go 同规则，自家 producer 跳过。

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	rediscommon "owl-common/redis"
	"owl-common/spatial"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	sensorActivityGroup    = "wisefido-sensor-activity"
	sensorActivityConsumer = "consumer-sensor-activity"
)

// ActivitySink 解耦 service 反向 import：service.TargetStateAggregator.PushEventFields 隐式满足。
type ActivitySink interface {
	PushEventFields(spatialPrefix, deviceAddr string, tsMs int64,
		walkDistanceMeters, walkDurationSec, standDurationSec, multiPersonDurationSec int)
}

type ActivityConsumer struct {
	client     *redislib.Client
	sink       ActivitySink
	sensorSlot netip.Prefix
	logger     *zap.Logger
}

func NewActivityConsumer(client *redislib.Client, sink ActivitySink, logger *zap.Logger) *ActivityConsumer {
	slot, _ := netip.ParsePrefix(spatial.SlotSensor)
	return &ActivityConsumer{
		client:     client,
		sink:       sink,
		sensorSlot: slot,
		logger:     logger,
	}
}

func (c *ActivityConsumer) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, c.client, rediscommon.StreamEvent.Name, sensorActivityGroup); err != nil {
		c.logger.Warn("sensor activity: create consumer group", zap.Error(err))
	}
	go c.runLoop(ctx)
	c.logger.Info("sensor activity consumer started",
		zap.String("stream", rediscommon.StreamEvent.Name),
		zap.String("group", sensorActivityGroup))
}

func (c *ActivityConsumer) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, c.client,
			rediscommon.StreamEvent.Name, sensorActivityGroup, sensorActivityConsumer,
			sensorReadCount, sensorReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Warn("sensor activity: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			c.handleRaw(m.Values)
			c.client.XAck(ctx, rediscommon.StreamEvent.Name, sensorActivityGroup, m.ID)
		}
	}
}

func (c *ActivityConsumer) handleRaw(raw map[string]interface{}) {
	msg, err := rediscommon.FromStreamMap(raw)
	if err != nil {
		c.logger.Warn("sensor activity: parse", zap.Error(err))
		return
	}
	if c.isSelfProducer(msg.Producer) {
		return
	}
	if msg.Category != alarm.Activity {
		return
	}
	if msg.SubjectEntity == "" {
		return
	}
	fields := rediscommon.FirstDataValue(msg.DataValue)
	if fields == nil {
		return
	}
	deviceAddr := msg.DeviceAddr.String()
	c.sink.PushEventFields(
		msg.SubjectEntity,
		deviceAddr,
		msg.Timestamp,
		intFromActivityField(fields[observation.FieldWalkDistance]),
		intFromActivityField(fields[observation.FieldWalkDuration]),
		intFromActivityField(fields[observation.FieldStandDuration]),
		intFromActivityField(fields[observation.FieldMultiPersonDuration]),
	)
}

// isSelfProducer 同 AlarmConsumer：自家平台 agent slot 内 IPv6 + hardcoded 名字跳过防 loop。
func (c *ActivityConsumer) isSelfProducer(producer string) bool {
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

func intFromActivityField(v interface{}) int {
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
