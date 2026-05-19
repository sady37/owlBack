package zoneengine

// adapter_radar.go — radar firmware → zoneengine SignalEvidence 翻译层。
//
// 订阅 iot:event:stream（独立 consumer group，与现有 EventConsumer/AlarmConsumer 完全解耦），
// 把 firmware 发的 EnterRoom / ExitRoom / NumberPeople / InBed / LeftBed 翻成 SignalEvidence
// 喂 Engine.Apply。SignalEvidence.Source 固定 "radar"。
//
// Bathroom 是 Room 的子集 / 特例（同 /88 prefix）：BathroomLookup.IsBathroom(roomPref)
// 命中时 EnterRoom/ExitRoom/NumberPeople 走 ZoneTypeBathroom（FSM/risk 阈值族不同），
// 输出仍走 card.RoomState 带 Kind=bathroom。bed 父级 zone 始终是 Room（床不在 bathroom）。
//
// 设计约束（来自上游 memory）：
//   - 未绑卡 device（subject_entity 空）不入 zone engine（zone 状态以 cardID 为主键）。
//   - 仅处理 device_type=="radar"；sleepace 的 InBed/LeftBed 走 adapter_sleepace 从 iot:alarm。
//   - 6s inbound age 过滤，与 sensor monitor consumer 同。

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// BathroomLookup 询问某 /88 RoomPref 是否属 bathroom 类。
//
// 由 wiring 层注入实现（典型实现读 rooms.room_name + 缓存）。空实现 / nil 时退化为
// 全部按 Room 处理 —— RoomState.Kind 不会标 bathroom，risk 阈值族退化为通用。
type BathroomLookup interface {
	IsBathroom(roomPref string) bool
}

// RadarAdapter 订阅 iot:event:stream 并把 radar 事件翻译为 SignalEvidence 喂 engine。
type RadarAdapter struct {
	client   *redislib.Client
	engine   *Engine
	bathroom BathroomLookup
	logger   *zap.Logger
}

const (
	radarConsumerGroup   = "wisefido-zoneengine-radar"
	radarConsumerName    = "consumer-zoneengine-radar"
	radarReadCount       = 16
	radarReadBlock       = 5 * time.Second
	radarMaxInboundAgeMs = 6000
)

// NewRadarAdapter 构造。bathroom 可为 nil（退化为全部按 Room 处理，仅日志路径用）。
func NewRadarAdapter(client *redislib.Client, engine *Engine, bathroom BathroomLookup, logger *zap.Logger) *RadarAdapter {
	return &RadarAdapter{client: client, engine: engine, bathroom: bathroom, logger: logger}
}

// Start 起独立 goroutine 跑读流循环，consumer group 与 cardagg/sensor 现有消费者隔离。
func (a *RadarAdapter) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, a.client, rediscommon.StreamEvent.Name, radarConsumerGroup); err != nil {
		a.logger.Warn("zoneengine radar adapter: create consumer group", zap.Error(err))
	}
	go a.runLoop(ctx)
	a.logger.Info("zoneengine radar adapter started",
		zap.String("stream", rediscommon.StreamEvent.Name),
		zap.String("group", radarConsumerGroup),
	)
}

func (a *RadarAdapter) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, a.client,
			rediscommon.StreamEvent.Name, radarConsumerGroup, radarConsumerName,
			radarReadCount, radarReadBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			a.logger.Warn("zoneengine radar adapter: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			a.handleRaw(m.Values)
			a.client.XAck(ctx, rediscommon.StreamEvent.Name, radarConsumerGroup, m.ID)
		}
	}
}

// handleRaw 单条 raw map → SignalEvidence。提出函数以便测试不依赖 redis client。
func (a *RadarAdapter) handleRaw(raw map[string]interface{}) {
	msg, err := rediscommon.FromStreamMap(raw)
	if err != nil {
		return
	}
	a.handleMsg(msg)
}

// handleMsg 单条 envelope → SignalEvidence。
func (a *RadarAdapter) handleMsg(msg *rediscommon.IoTStreamMessage) {
	if msg == nil || !msg.DeviceAddr.IsValid() {
		return
	}
	// 未绑卡 device 不入 zone engine（zone 状态以 cardID 为主键）。
	if msg.SubjectEntity == "" {
		return
	}
	// 仅 radar；sleepace InBed/LeftBed 走 adapter_sleepace。
	if !strings.EqualFold(msg.DeviceType, "radar") {
		return
	}
	// 老消息丢弃（与 sensor monitor consumer 同 6s）
	nowMs := time.Now().UnixMilli()
	if nowMs-msg.Timestamp > radarMaxInboundAgeMs {
		return
	}
	fields := rediscommon.FirstDataValue(msg.DataValue)
	if fields == nil {
		fields = map[string]interface{}{}
	}

	roomPref := prefixOf(msg.DeviceAddr, 88)
	bedPref := prefixOf(msg.DeviceAddr, 96)

	switch msg.Category {
	case alarm.EnterRoom:
		a.applyRoomLike(msg.SubjectEntity, roomPref, "enter", msg.Timestamp, fields)
	case alarm.ExitRoom:
		a.applyRoomLike(msg.SubjectEntity, roomPref, "leave", msg.Timestamp, fields)
	case alarm.NumberPeople:
		count := intFromAny(fields[observation.FieldNumberPeople])
		a.applyCount(msg.SubjectEntity, roomPref, count, msg.Timestamp, fields)
	case alarm.InBed:
		a.applyBed(msg.SubjectEntity, bedPref, "enter", msg.Timestamp, fields)
	case alarm.LeftBed:
		a.applyBed(msg.SubjectEntity, bedPref, "leave", msg.Timestamp, fields)
	default:
		// 忽略其它 event_name
	}
}

// applyRoomLike — bathroom 命中即仅发 ZoneTypeBathroom（不再发 ZoneTypeRoom），由 stay alarm enable 侧消费。
func (a *RadarAdapter) applyRoomLike(cardID, roomPref, kind string, ts int64, fields map[string]interface{}) {
	if cardID == "" || roomPref == "" {
		return
	}
	zt := a.routeRoomZoneType(roomPref)
	a.engine.Apply(SignalEvidence{
		CardID:      cardID,
		ZoneType:    zt,
		ZoneID:      roomPref,
		Source:      "radar",
		Kind:        kind,
		Ts:          ts,
		TriggerData: fields,
	})
}

func (a *RadarAdapter) applyCount(cardID, roomPref string, count int, ts int64, fields map[string]interface{}) {
	if cardID == "" || roomPref == "" {
		return
	}
	zt := a.routeRoomZoneType(roomPref)
	a.engine.Apply(SignalEvidence{
		CardID:      cardID,
		ZoneType:    zt,
		ZoneID:      roomPref,
		Source:      "radar",
		Kind:        "count_change",
		Count:       count,
		Ts:          ts,
		TriggerData: fields,
	})
}

func (a *RadarAdapter) applyBed(cardID, bedPref, kind string, ts int64, fields map[string]interface{}) {
	if cardID == "" || bedPref == "" {
		return
	}
	a.engine.Apply(SignalEvidence{
		CardID:      cardID,
		ZoneType:    ZoneTypeBed,
		ZoneID:      bedPref,
		Source:      "radar",
		Kind:        kind,
		Ts:          ts,
		TriggerData: fields,
	})
}

func (a *RadarAdapter) routeRoomZoneType(roomPref string) ZoneType {
	if a.bathroom != nil && a.bathroom.IsBathroom(roomPref) {
		return ZoneTypeBathroom
	}
	return ZoneTypeRoom
}

// prefixOf addr 在指定 mask 下的 CIDR 文本（"fd00:0:3:111:3:101::/96"）。
// 无效 addr 返回 ""。
func prefixOf(addr netip.Addr, bits int) string {
	if !addr.IsValid() {
		return ""
	}
	return netip.PrefixFrom(addr, bits).Masked().String()
}

// intFromAny 与 cardagg consumer 同语义；Redis stream 反序列化常出 float64。
func intFromAny(v interface{}) int {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	default:
		return 0
	}
}
