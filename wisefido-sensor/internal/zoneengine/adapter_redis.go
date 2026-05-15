package zoneengine

// adapter_redis.go — ZoneEvent → card:status hash 投影。
//
// 实现 ZoneEventListener，按 ZoneType 三向分流写 card:status:{cardID} 的
// bed_state / room_state / bathroom_state 字段，复用 owl-common/card.Writer。
//
// 关键设计：
//
//  1. 读-改-写：先 ReadCardStatus 保留 engine 不管的字段（SleepStage / StayFSMPhase /
//     AreaPeople / TrackNumber / DeviceID 等），再覆写 engine 拥有的字段。
//     —— 避免 HSET 整 JSON 导致下游 satellite supervisor 写入的字段被抹掉。
//
//  2. Engine-owned 字段集（types.go 注释明确）：
//     · BedState:      BedStatus / BedEvent / UpdatedAt / StartTime / DurationSec
//     · RoomState:     TotalPeople / LastEnterTime / LastExitTime / UpdatedAt / HasMulti
//     · BathRoomState: TotalPeople / LastEnterTime / LastExitTime / UpdatedAt / HasMulti
//     其余字段一律保留（StayFSMPhase / SleepStage / AreaPeople / RoomName 等）。
//
//  3. 过渡期警告：cardagg B+C 组（bed_handler / room_state / state_service 派生半边）仍在写
//     同 hash；race 窗口内偶尔被 cardagg 覆盖回旧值，下次 ZoneEvent 触发即修正。
//     B+C 组迁完即闭合该 race（详 [zoneengine_phase1_done] memory）。

import (
	"context"
	"time"

	"owl-common/card"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisAdapter 把 ZoneEvent 投影到 card:status hash。
type RedisAdapter struct {
	reader  *card.Reader
	writer  *card.Writer
	logger  *zap.Logger
	timeout time.Duration
}

// NewRedisAdapter writer/reader 由 caller 注入（共享同一个 redis client）。
// timeout 单次 Redis 操作上限，0 用 2s 默认。
func NewRedisAdapter(client *redislib.Client, statusMaxLen, realtimeMaxLen int64, timeout time.Duration, logger *zap.Logger) *RedisAdapter {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &RedisAdapter{
		reader:  card.NewReader(client),
		writer:  card.NewWriter(client, statusMaxLen, realtimeMaxLen),
		logger:  logger,
		timeout: timeout,
	}
}

// OnZoneEvent satisfy ZoneEventListener。
//
// 每条 ZoneEvent 单独触发一次 read+write；不批量。理由：ZoneEvent 已经是翻转沿（频率低），
// 每秒每 card 至多个位数；hot path 走 sustain 时不发 ZoneEvent，不会暴 IO。
func (a *RedisAdapter) OnZoneEvent(e ZoneEvent) {
	if e.CardID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	cur, err := a.reader.ReadCardStatus(ctx, e.CardID)
	if err != nil {
		if a.logger != nil {
			a.logger.Warn("zoneengine redis adapter: read card status failed",
				zap.String("card_id", e.CardID),
				zap.Error(err),
			)
		}
		// read 失败仍然继续 write —— 等价丢失保留字段一次，下条 event 修正。
		cur = &card.CardStatus{CardID: e.CardID}
	}

	fields := a.translate(e, cur)
	if isPublishFieldsEmpty(fields) {
		return
	}

	status := &card.CardStatus{CardID: e.CardID}
	if fields.bedState != nil {
		status.BedState = fields.bedState
	}
	if fields.roomState != nil {
		status.RoomState = fields.roomState
	}
	if fields.bathRoomState != nil {
		status.BathRoomState = fields.bathRoomState
	}
	if err := a.writer.WriteCardStatus(ctx, status); err != nil && a.logger != nil {
		a.logger.Warn("zoneengine redis adapter: write card status failed",
			zap.String("card_id", e.CardID),
			zap.Error(err),
		)
	}
}

// publishFields engine 实际要写的三块（仅指针非 nil 的会被 publish）。
type publishFields struct {
	bedState      *card.BedState
	roomState     *card.RoomState
	bathRoomState *card.BathRoomState
}

func isPublishFieldsEmpty(f publishFields) bool {
	return f.bedState == nil && f.roomState == nil && f.bathRoomState == nil
}

// translate ZoneEvent + 现有状态 → 仅含被改字段的 publishFields。
func (a *RedisAdapter) translate(e ZoneEvent, cur *card.CardStatus) publishFields {
	switch e.ZoneType {
	case ZoneTypeBed:
		return publishFields{bedState: a.deriveBedState(e, cur.BedState)}
	case ZoneTypeRoom:
		return publishFields{roomState: a.deriveRoomState(e, cur.RoomState)}
	case ZoneTypeBathroom:
		return publishFields{bathRoomState: a.deriveBathroomState(e, cur.BathRoomState)}
	}
	return publishFields{}
}

// deriveBedState 引擎管辖：BedStatus / BedEvent / UpdatedAt / StartTime / DurationSec。
// 其余字段（TrackNumber / BedConfidence / SleepStage / SleepConfidence）从 prev 保留。
func (a *RedisAdapter) deriveBedState(e ZoneEvent, prev *card.BedState) *card.BedState {
	out := &card.BedState{}
	if prev != nil {
		*out = *prev // copy 保留全部
	}
	out.UpdatedAt = e.NewState.UpdatedAt

	// BedStatus（与 cardagg 旧定义一致 0=在床, 1=离床）：IsPresent → 0，Vacant → 1。
	if e.NewState.IsPresent() {
		out.BedStatus = 0
	} else {
		out.BedStatus = 1
	}

	// BedEvent（与 cardagg 常量对齐 0=InBed / 1=LeftBed / 8=None）。
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.BedEvent = 0
	case TransitionVacant:
		out.BedEvent = 1
	case TransitionLeaving, TransitionCountChange:
		out.BedEvent = 8
	}

	// StartTime / DurationSec：占用区间窗口；leaving 期间保持窗口（仍 IsPresent）。
	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.StartTime = e.NewState.LastEnterTs
		out.DurationSec = 0
	case TransitionVacant:
		// 真正离床；duration 用 LastExit-LastEnter（毫秒 → 秒）
		if out.StartTime > 0 && e.NewState.LastExitTs > out.StartTime {
			out.DurationSec = int((e.NewState.LastExitTs - out.StartTime) / 1000)
		}
	case TransitionLeaving:
		// 维持 StartTime / DurationSec 推进
		if out.StartTime > 0 {
			out.DurationSec = int((e.NewState.UpdatedAt - out.StartTime) / 1000)
		}
	}
	return out
}

// deriveRoomState 引擎管辖：TotalPeople / LastEnterTime / LastExitTime / UpdatedAt / HasMulti。
// 其余字段（AreaPeople / StandingContinuousMin / HasRisk）从 prev 保留。
func (a *RedisAdapter) deriveRoomState(e ZoneEvent, prev *card.RoomState) *card.RoomState {
	out := &card.RoomState{}
	if prev != nil {
		*out = *prev
		// AreaPeople 是 map，shallow copy 已 ok（不会被本函数改）
	}
	out.UpdatedAt = e.NewState.UpdatedAt
	out.TotalPeople = e.NewState.Count

	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.LastEnterTime = e.NewState.LastEnterTs
	case TransitionVacant:
		out.LastExitTime = e.NewState.LastExitTs
	}
	out.HasMulti = out.TotalPeople > 1
	return out
}

// deriveBathroomState 引擎管辖：TotalPeople / LastEnterTime / LastExitTime / UpdatedAt / HasMulti。
// 其余字段（StayFSMPhase / StayArmEnterAt / StayResolveExitAt / StaySec / StandingContinuousMin /
// HasRisk / DeviceID / DeviceUID / RoomID / RoomName）从 prev 保留 —— stay_fsm satellite 独立维护。
func (a *RedisAdapter) deriveBathroomState(e ZoneEvent, prev *card.BathRoomState) *card.BathRoomState {
	out := &card.BathRoomState{}
	if prev != nil {
		*out = *prev
	}
	out.UpdatedAt = e.NewState.UpdatedAt
	out.TotalPeople = e.NewState.Count

	switch e.Transition {
	case TransitionOccupied, TransitionReturned:
		out.LastEnterTime = e.NewState.LastEnterTs
	case TransitionVacant:
		out.LastExitTime = e.NewState.LastExitTs
	}
	out.HasMulti = out.TotalPeople > 1
	return out
}
