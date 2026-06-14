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
// DeviceFitnessChecker adapter 入口 gate：unfit device 的事件不喂 engine 防垃圾数据污染 FSM。
// service.DeviceFitnessTracker 隐式满足；nil 注入时全部 device 视作 fit（兼容未 wire 场景）。
type DeviceFitnessChecker interface {
	IsFit(deviceAddr string) bool
}

// BedPresenceSink — adapter_radar / adapter_sleepace 在 bed enter/leave 翻转时旁路回调。
// service.BedPresenceFusion 实现；供 stream_publisher 做 RoomState dedup 时查 (X,Y)。
// BedZone FSM 把两源 fuse 成单一 enter/leave，源信息丢失——这条 sink 是旁路保留。
type BedPresenceSink interface {
	SetSleepad(bedCIDR string, inBed bool, ts int64)
	SetRadar(bedCIDR string, inBed bool, ts int64)
}

// RadarRoomCountSink — adapter_radar NumberPeople 写入时旁路记账。
// service.RadarRoomCountCache 实现；供 publisher 读 raw Z（不被 subset_invariant lift 污染）。
type RadarRoomCountSink interface {
	SetZ(roomCIDR string, count int, ts int64)
}

// BedResolver 按 radar 设备 + 所在 /88 room，用 MM covers 置信解析出该 radar 床事件该归的床 /96。
// 让 radar 床事件喂到床的 /96（与 sleepad 同 zone），激活 bed_bayesian 双源融合（LR_S+LR_R+γ）。
// 单床房 covers=1.0 → 解析；多床候选 0.5/0.5 几何分不清 → 返回 ""（退回 radar 自己 /96，待 DBN 把
// covers 推到 0.95 后自动解析）。nil 或返回 "" 时退回 radar 自己的 /96（保持原行为）。
type BedResolver interface {
	ResolveBed(radarAddr, roomPrefix string) (bed string, coversConf float64)
}

type RadarAdapter struct {
	client      *redislib.Client
	engine      *Engine
	bathroom    BathroomLookup
	bedDedup    *BedEventDedup       // S5b: InBed/LeftBed per-device 10s 同类 dedup
	fitness     DeviceFitnessChecker // S6: 设备类 alarm gate
	presence    BedPresenceSink      // RoomState dedup：bed-area Y 跟踪
	roomCount   RadarRoomCountSink   // RoomState dedup：raw Z 缓存
	bedResolver BedResolver          // MM covers 解析：radar 床事件按覆盖置信归到床 /96
	logger      *zap.Logger
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
	return &RadarAdapter{client: client, engine: engine, bathroom: bathroom, bedDedup: NewBedEventDedup(), logger: logger}
}

// SetFitnessChecker main wiring 注入；不调时退化为"所有 device 都 fit"。
func (a *RadarAdapter) SetFitnessChecker(c DeviceFitnessChecker) {
	if a == nil {
		return
	}
	a.fitness = c
}

// SetBedPresence main wiring 注入；不调时 dedup 旁路 disable（兼容未 wire）。
func (a *RadarAdapter) SetBedPresence(s BedPresenceSink) {
	if a == nil {
		return
	}
	a.presence = s
}

// SetRoomCountSink main wiring 注入；不调时 raw Z 不缓存（publisher 退化到 e.NewState.Count）。
func (a *RadarAdapter) SetRoomCountSink(s RadarRoomCountSink) {
	if a == nil {
		return
	}
	a.roomCount = s
}

// SetBedResolver main wiring 注入 MM covers 床解析；不调时 radar 床事件用自己的 /96（原行为）。
func (a *RadarAdapter) SetBedResolver(r BedResolver) {
	if a == nil {
		return
	}
	a.bedResolver = r
}

// resolveBedPref 用 MM covers 把 radar 床事件归到它覆盖的床 /96（与 sleepad 同 zone → bed_bayesian 双源融合）。
// 无 resolver / 多床候选未定 / 无解析 → 退回 radar 自己的 /96（fallback，保持原行为）。
// 返回 (床 /96, covers 结构权重)：解析到床带其 covers 置信（喂 bed_bayesian radar 加权）；
// 退回 radar 自己 /96 时权重 1.0（自家 zone，全权）。
func (a *RadarAdapter) resolveBedPref(radarAddr, roomPref, fallback string) (string, float64) {
	if a.bedResolver != nil {
		if b, conf := a.bedResolver.ResolveBed(radarAddr, roomPref); b != "" {
			return b, conf
		}
	}
	return fallback, 1.0
}

// Start 起独立 goroutine 跑读流循环，consumer group 与 cardagg/sensor 现有消费者隔离。
func (a *RadarAdapter) Start(ctx context.Context) {
	if err := rediscommon.CreateConsumerGroup(ctx, a.client, ("test:" + rediscommon.StreamEvent.Name), radarConsumerGroup); err != nil {
		a.logger.Warn("zoneengine radar adapter: create consumer group", zap.Error(err))
	}
	go a.runLoop(ctx)
	a.logger.Info("zoneengine radar adapter started",
		zap.String("stream", ("test:" + rediscommon.StreamEvent.Name)),
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
			("test:" + rediscommon.StreamEvent.Name), radarConsumerGroup, radarConsumerName,
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
			a.client.XAck(ctx, ("test:" + rediscommon.StreamEvent.Name), radarConsumerGroup, m.ID)
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
	// S6 fitness gate：unfit device（SensorDetached/AngleException/Offline/SignalPoor）发的数据
	// 可能是垃圾，不喂 engine 防 FSM 污染（详 service.DeviceFitnessTracker）。
	if a.fitness != nil && !a.fitness.IsFit(msg.DeviceAddr.String()) {
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
		a.applyRoomLike(roomPref, "enter", msg.Timestamp, fields)
	case alarm.ExitRoom:
		a.applyRoomLike(roomPref, "leave", msg.Timestamp, fields)
	case alarm.NumberPeople:
		count := intFromAny(fields[observation.FieldNumberPeople])
		a.applyCount(roomPref, count, msg.Timestamp, fields)
	case alarm.InBed:
		if a.bedDedup.Admit(msg.DeviceAddr.String(), "enter", msg.Timestamp) {
			bed, w := a.resolveBedPref(msg.DeviceAddr.String(), roomPref, bedPref)
			a.applyBed(bed, msg.DeviceAddr.String(), w, "enter", msg.Timestamp, fields)
		}
	case alarm.LeftBed:
		if a.bedDedup.Admit(msg.DeviceAddr.String(), "leave", msg.Timestamp) {
			bed, w := a.resolveBedPref(msg.DeviceAddr.String(), roomPref, bedPref)
			a.applyBed(bed, msg.DeviceAddr.String(), w, "leave", msg.Timestamp, fields)
		}
	default:
		// 忽略其它 event_name
	}
}

// applyRoomLike — bathroom 命中即仅发 ZoneTypeBathroom（不再发 ZoneTypeRoom），由 stay alarm enable 侧消费。
func (a *RadarAdapter) applyRoomLike(roomPref, kind string, ts int64, fields map[string]interface{}) {
	if roomPref == "" {
		return
	}
	zt := a.routeRoomZoneType(roomPref)
	a.engine.Apply(SignalEvidence{
		ZoneType:    zt,
		ZoneID:      roomPref,
		Source:      "radar",
		Kind:        kind,
		Ts:          ts,
		TriggerData: fields,
	})
}

func (a *RadarAdapter) applyCount(roomPref string, count int, ts int64, fields map[string]interface{}) {
	if roomPref == "" {
		return
	}
	zt := a.routeRoomZoneType(roomPref)
	a.engine.Apply(SignalEvidence{
		ZoneType:    zt,
		ZoneID:      roomPref,
		Source:      "radar",
		Kind:        "count_change",
		Count:       count,
		Ts:          ts,
		TriggerData: fields,
	})
	if a.roomCount != nil {
		a.roomCount.SetZ(roomPref, count, ts)
	}
}

func (a *RadarAdapter) applyBed(bedPref, radarAddr string, coversW float64, kind string, ts int64, fields map[string]interface{}) {
	if bedPref == "" {
		return
	}
	a.engine.Apply(SignalEvidence{
		ZoneType:          ZoneTypeBed,
		ZoneID:            bedPref,
		Source:            "radar",
		Kind:              kind,
		Ts:                ts,
		TriggerData:       fields,
		RadarCoversWeight: coversW,
		SourceDevice:      radarAddr,
	})
	if a.presence != nil {
		a.presence.SetRadar(bedPref, kind == "enter", ts)
	}
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
