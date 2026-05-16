package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"

	"go.uber.org/zap"
)

// StateService cardagg 卡片状态读写服务。
//
// **2026-05-14 zone engine cutover**：BedStatus / BedEvent / UpdatedAt / StartTime /
// DurationSec 已迁出 — 由 wisefido-sensor zone engine RedisAdapter 统一权威。
// 本服务保留写入：TrackNumber + BedConfidence + SleepStage / SleepConfidence + Target
// + 卡初始化 + AreaPeople（PublishRoomStateFromEvent）+ Activity 派生。
//
// 已删函数（下游显式信号 + zone engine sustain 替代）：
//   - DeriveBedStateFromRealtime（vital 累积 N 轮推导上下床）
//   - ReconcileRoomStateFromBedState（zone engine subset_invariant 接管）
//   - NoteLeftBedCooldown / vitalDeriveRound* / leftBedCooldownUntil 等内部状态
type StateService struct {
	writer *card.Writer
	reader *card.Reader
	logger *zap.Logger

	preparedMu    sync.Mutex
	preparedCards map[string]struct{} // cardID 已执行过 EnsureCardStatePrepared
}

func NewStateService(writer *card.Writer, reader *card.Reader, logger *zap.Logger) *StateService {
	return &StateService{
		writer:        writer,
		reader:        reader,
		logger:        logger,
		preparedCards: make(map[string]struct{}),
	}
}

// BedConfidence 采用 100 分制：Sleepad 基准 90，Radar 基准 60。
const (
	BedConfidenceSleepadBase   = 90
	BedConfidenceRadarBase     = 60
	SleepConfidenceSleepadBase = 90
	SleepConfidenceRadarBase   = 60
)

// bedConfidenceFromDeviceType 进出床事件置信度基准（100 分制）：Sleepad=90，Radar=60，其它=0。
func bedConfidenceFromDeviceType(deviceType string) int {
	switch strings.ToLower(deviceType) {
	case "sleepad", "sleeppad":
		return BedConfidenceSleepadBase
	case "radar":
		return BedConfidenceRadarBase
	default:
		return 0
	}
}

const bedStateWindowMs = 5000 // 5s 内视为同一时间窗口，做权重比较

// ReadCardStatus 读取当前卡片状态（Targets/BedState 等），供 event 层防静电等逻辑使用。
// Phase A：DeviceStatus 已迁出 card:state，调用 reader.ReadDeviceStatus 单独读。
func (s *StateService) ReadCardStatus(ctx context.Context, cardID string) (*card.CardStatus, error) {
	if s.reader == nil {
		return nil, nil
	}
	return s.reader.ReadCardStatus(ctx, cardID)
}

// UpdateTargetPose 按事件更新 CardStatus.Target（单 Target）的 Pose、PoseSince 等，读当前后合并写回。
func (s *StateService) UpdateTargetPose(ctx context.Context, cardID string, trackID, pose int, poseSinceMs int64) error {
	if s.writer == nil || s.reader == nil || cardID == "" || trackID <= 0 || trackID == 88 {
		return nil
	}
	curr, err := s.reader.ReadCardStatus(ctx, cardID)
	if err != nil || curr == nil {
		return nil
	}
	ts := curr.Target
	if ts == nil {
		ts = &card.TargetState{TrackID: trackID, UpdatedAt: poseSinceMs}
	}
	ts.UpdatedAt = time.Now().UnixMilli()
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{Target: ts})
}

// BedEvent 常量：与 card.BedState.BedEvent 及事件语义一致
const (
	BedEventNone    = 8 // 无事件或保持不变
	BedEventInBed   = 0 // InBed
	BedEventLeftBed = 1 // LeftBed
)

// BedStatusUnchanged 无上下床事件时的占位，与 card.BedState 注释 8=未改变 一致
const BedStatusUnchanged = 8

// bedStatusFromTrackNumber 由 TrackNumber 推导 BedStatus：0=在床, 1=离床
func bedStatusFromTrackNumber(trackNumber int) int {
	if trackNumber > 0 {
		return 0 // 在床
	}
	return 1 // 离床
}

// PublishBedStateFromEvent 在 event_handler 收到 InBed/LeftBed 时写 BedState。
//
// **2026-05-14 zone engine cutover**：本函数职责收窄到只写 TrackNumber + BedConfidence
// （+ LeftBed 时清 SleepStage）。BedStatus / BedEvent / UpdatedAt / StartTime /
// DurationSec 由 zone engine RedisAdapter 独立写入（read-modify-write 保留 cardagg
// 写的字段）。
//
// 仍保留 timestampMs 入参防签名变化波及 caller，本函数不再直接用它落库（zone engine 写
// UpdatedAt）；trackNumberOverride 仍生效（决定 TrackNumber 算法）。
func (s *StateService) PublishBedStateFromEvent(ctx context.Context, cardID, eventName, deviceType string, timestampMs int64, _ int, trackNumberOverride *int) (written bool, err error) {
	_ = timestampMs // 仅为兼容 caller 签名；engine 写 UpdatedAt
	switch eventName {
	case alarm.InBed:
		return s.publishBedStateInBed(ctx, cardID, deviceType, trackNumberOverride)
	case alarm.LeftBed:
		return s.publishBedStateLeftBed(ctx, cardID, deviceType, trackNumberOverride)
	default:
		return false, nil
	}
}

// publishBedStateInBed read-modify-write 仅修改 TrackNumber + BedConfidence；engine
// owns 字段（BedStatus / BedEvent / UpdatedAt / StartTime / DurationSec）原样保留。
//
// trackNumberOverride 非 nil（Sleepad 路径）→ 直接覆盖 [0, trackNumberMax]；
// nil（Radar 路径）→ prev.TrackNumber + 1，cap trackNumberMax。
func (s *StateService) publishBedStateInBed(ctx context.Context, cardID, deviceType string, trackNumberOverride *int) (written bool, err error) {
	prev := s.readBedState(ctx, cardID)
	out := &card.BedState{}
	if prev != nil {
		*out = *prev // 完整保留 engine owns + SleepStage 等
	}

	if trackNumberOverride != nil {
		out.BedConfidence = BedConfidenceSleepadBase
		out.TrackNumber = clampTrackNumber(*trackNumberOverride)
	} else {
		out.BedConfidence = BedConfidenceRadarBase
		if prev == nil || prev.BedStatus != 0 {
			out.TrackNumber = 1
		} else {
			out.TrackNumber = clampTrackNumber(prev.TrackNumber + 1)
		}
	}

	err = PublishCardStatus(ctx, s.writer, cardID, PublishFields{BedState: out})
	return err == nil, err
}

// publishBedStateLeftBed read-modify-write 仅修改 TrackNumber + BedConfidence + 清
// SleepStage（离床即清，避免"Sleepad 关了仍显示 Awake"）。engine owns 字段保留。
func (s *StateService) publishBedStateLeftBed(ctx context.Context, cardID, deviceType string, trackNumberOverride *int) (written bool, err error) {
	prev := s.readBedState(ctx, cardID)
	out := &card.BedState{}
	if prev != nil {
		*out = *prev
	}
	out.SleepStage = 0
	out.SleepConfidence = 0

	if trackNumberOverride != nil {
		out.BedConfidence = BedConfidenceSleepadBase
		out.TrackNumber = clampTrackNumber(*trackNumberOverride)
	} else {
		out.BedConfidence = BedConfidenceRadarBase
		next := 0
		if prev != nil && prev.TrackNumber > 0 {
			next = prev.TrackNumber - 1
		}
		out.TrackNumber = clampTrackNumber(next)
	}

	err = PublishCardStatus(ctx, s.writer, cardID, PublishFields{BedState: out})
	return err == nil, err
}

// readBedState 读取卡当前 BedState；nil-safe。
func (s *StateService) readBedState(ctx context.Context, cardID string) *card.BedState {
	if s.reader == nil {
		return nil
	}
	curr, err := s.reader.ReadCardStatus(ctx, cardID)
	if err != nil || curr == nil {
		return nil
	}
	return curr.BedState
}

func clampTrackNumber(n int) int {
	if n < 0 {
		return 0
	}
	if n > trackNumberMax {
		return trackNumberMax
	}
	return n
}

const trackNumberMax = 2

// RoomStateEventKind 进出房/人数事件类型，用于 HasMulti/HasRisk 规则。
type RoomStateEventKind int

const (
	RoomStateEventEnter RoomStateEventKind = iota
	RoomStateEventExit
	RoomStateEventNumberPeople
)

// RoomStateTotalPeopleNoRadar 无房间雷达时 RoomState.TotalPeople 置此值，表示不维护房间人数；前端仅显示 bed state。
const RoomStateTotalPeopleNoRadar = -1

// PublishRoomStateFromEvent 房间（非卫生间）事件更新 RoomState。AreaPeople 存各 device_uid→NumberPeople，TotalPeople=sum(AreaPeople)；Enter/Exit 仅更新时间，用 NumberPeople 校正。
// Enter: LastEnterTime=ts；HasMulti/HasRisk 按当前 TotalPeople（sum(AreaPeople)）算
// Exit: LastExitTime=ts；同上
// NumberPeople: AreaPeople[device_uid]=totalPeople，TotalPeople=sum(AreaPeople)；该卡下仅1人即 HasRisk。
func (s *StateService) PublishRoomStateFromEvent(ctx context.Context, cardID, deviceUID string, kind RoomStateEventKind, totalPeople int, ts int64) error {
	if cardID == "" || deviceUID == "" {
		return nil
	}
	var cur *card.RoomState
	if s.reader != nil {
		if curr, err := s.reader.ReadCardStatus(ctx, cardID); err == nil && curr != nil && curr.RoomState != nil {
			cur = &card.RoomState{}
			*cur = *curr.RoomState
		}
	}
	if cur == nil {
		cur = &card.RoomState{UpdatedAt: ts, TotalPeople: 0, HasMulti: false, HasRisk: false}
	}
	if cur.AreaPeople == nil {
		cur.AreaPeople = make(map[string]int)
	}
	if kind == RoomStateEventNumberPeople {
		// 先更新 AreaPeople[device_uid]，再算 TotalPeople
		if totalPeople < 0 {
			totalPeople = 0
		}
		cur.AreaPeople[deviceUID] = totalPeople
		cur.TotalPeople = sumAreaPeople(cur.AreaPeople)
		cur.UpdatedAt = ts
	} else if kind == RoomStateEventEnter {
		cur.LastEnterTime = ts
		cur.UpdatedAt = ts
		cur.TotalPeople = sumAreaPeople(cur.AreaPeople)
	} else if kind == RoomStateEventExit {
		cur.LastExitTime = ts
		cur.UpdatedAt = ts
		cur.TotalPeople = sumAreaPeople(cur.AreaPeople)
	}
	if cur.TotalPeople == 0 {
		cur.HasMulti = false
		cur.HasRisk = false
	} else if cur.TotalPeople == 1 {
		cur.HasMulti = false
		cur.HasRisk = true
	} else {
		cur.HasMulti = true
		cur.HasRisk = false
	}
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{RoomState: cur})
}

func sumAreaPeople(m map[string]int) int {
	var n int
	for _, v := range m {
		n += v
	}
	return n
}

// PublishBathRoomStateFromEvent 卫生间设备 Enter/Exit/NumberPeople 事件更新 BathRoomState。匹配以 deviceID 为主，无则用 deviceUID（一卡一卫生间雷达）。
// roomID/roomName 非空时写入或覆盖 BathRoomState（与 device 绑定房间一致）。
// Exit 仅写 LastExitTime；HasMulti/HasRisk 与 TotalPeople 仅随 NumberPeople 变化，与 StayFSM 一致。
func (s *StateService) PublishBathRoomStateFromEvent(ctx context.Context, cardID, deviceID, deviceUID string, kind RoomStateEventKind, totalPeople int, ts int64, roomID, roomName string) error {
	if s.writer == nil || cardID == "" || (deviceID == "" && deviceUID == "") {
		return nil
	}
	var cur *card.BathRoomState
	if s.reader != nil {
		if curr, err := s.reader.ReadCardStatus(ctx, cardID); err == nil && curr != nil && curr.BathRoomState != nil {
			c := curr.BathRoomState
			match := (c.DeviceID != "" && c.DeviceID == deviceID) || (c.DeviceUID != "" && c.DeviceUID == deviceUID)
			if !match {
				return nil
			}
			cur = &card.BathRoomState{}
			*cur = *c
		}
	}
	if cur == nil {
		cur = &card.BathRoomState{DeviceID: deviceID, DeviceUID: deviceUID, UpdatedAt: ts, TotalPeople: 0, HasMulti: false, HasRisk: false}
	} else {
		cur.UpdatedAt = ts
	}
	if roomID != "" {
		cur.RoomID = roomID
	}
	if roomName != "" {
		cur.RoomName = roomName
	}
	if kind == RoomStateEventEnter {
		cur.LastEnterTime = ts
		if cur.TotalPeople >= 1 {
			cur.HasMulti = true
			cur.HasRisk = false
		} else {
			cur.HasMulti = false
			cur.HasRisk = true
		}
	} else if kind == RoomStateEventExit {
		cur.LastExitTime = ts
		// TotalPeople 仅由 NumberPeople 更新；Exit 不改写 HasMulti/HasRisk，避免与 Stay 解除窗（仍可能有人）不同步。
	} else {
		if totalPeople >= 0 {
			cur.TotalPeople = totalPeople
		}
		if cur.TotalPeople == 0 {
			cur.HasMulti = false
			cur.HasRisk = false
		} else if cur.TotalPeople == 1 {
			cur.HasMulti = false
			cur.HasRisk = true
		} else {
			cur.HasMulti = true
			cur.HasRisk = false
		}
	}
	if cur.TotalPeople == 0 {
		cur.StaySec = 0
	}
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{BathRoomState: cur})
}

// PublishBathRoomStayFSM 仅合并 Stay 状态机展示字段到 BathRoomState（不改变人数与进出时间）。
func (s *StateService) PublishBathRoomStayFSM(ctx context.Context, cardID, deviceID, deviceUID, phase string, armEnterAt, resolveExitAt int64) error {
	if s.writer == nil || cardID == "" || (deviceID == "" && deviceUID == "") {
		return nil
	}
	var cur *card.BathRoomState
	if s.reader != nil {
		if curr, err := s.reader.ReadCardStatus(ctx, cardID); err == nil && curr != nil && curr.BathRoomState != nil {
			c := curr.BathRoomState
			match := (c.DeviceID != "" && c.DeviceID == deviceID) || (c.DeviceUID != "" && c.DeviceUID == deviceUID)
			if !match {
				return nil
			}
			cur = &card.BathRoomState{}
			*cur = *c
		}
	}
	if cur == nil {
		cur = &card.BathRoomState{DeviceID: deviceID, DeviceUID: deviceUID, UpdatedAt: time.Now().UnixMilli()}
	}
	cur.StayFSMPhase = phase
	cur.StayArmEnterAt = armEnterAt
	cur.StayResolveExitAt = resolveExitAt
	cur.UpdatedAt = time.Now().UnixMilli()
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{BathRoomState: cur})
}

// ResetRoomState 设备上线或 card 变更时，将该卡 RoomState 重置为空（卫生间卡用）。
func (s *StateService) ResetRoomState(ctx context.Context, cardID string) error {
	rs := &card.RoomState{UpdatedAt: time.Now().UnixMilli(), TotalPeople: 0, HasMulti: false, HasRisk: false}
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{RoomState: rs})
}

// DeriveAndWriteState 在 derive 定时点执行：写 Target。
// device:status 不再由 derive 路径维护——已切换为 monitor/event/alarm 流事件驱动 + 看门狗 fail-safe。
func (s *StateService) DeriveAndWriteState(
	ctx context.Context,
	snap CardSnapshot,
	meta *CardMeta,
	prevTarget *card.TargetState,
	buf *MonitorBuffer,
) (*card.CardStatus, error) {
	_ = meta
	_ = buf

	status := &card.CardStatus{
		CardID: snap.CardID,
		Target: prevTarget,
	}

	if err := PublishCardStatus(ctx, s.writer, snap.CardID, PublishFields{
		Target: status.Target,
	}); err != nil {
		return nil, err
	}

	s.logger.Debug("card:status derive",
		zap.String("cid", snap.CardID),
		zap.Bool("has_target", prevTarget != nil))

	return status, nil
}

// hasBedCapableDevice — 卡是否含 bed-capable 设备：
//   - active_bed 卡按定义有床
//   - 其它卡（unit/room）：Devices 内含 Sleepad → 视为有床（v3 absorb 模式 UnitCard 也可能直管 sleepad）
//
// 非 bed-capable 卡（如纯雷达 UnitCard，覆盖 Bathroom/Kitchen 等无床房间）不应初始化 BedState；
// 否则 Redis 长期停留 bed_status=8/start_time=init 占位值，FE 计算 `now - start_time` 当离床时长
// 会显示假的 OOB. 15h+ 之类数据（详见 FE Overview.vue:2131 修复）。
func hasBedCapableDevice(meta *CardMeta) bool {
	if meta == nil {
		return false
	}
	if meta.IsActiveBedCard() {
		return true
	}
	for _, dm := range meta.Devices {
		if dm == nil {
			continue
		}
		if strings.EqualFold(dm.DeviceType, "Sleepad") {
			return true
		}
	}
	return false
}

// InitCardRoomAndBathroomState Card 初始化时：默认创建 RoomState；存在 isBathroomRadar 时创建 BathRoomState；
// 仅 bed-capable 卡（active_bed 或含 Sleepad）写占位 BedState(bed_status=8)。
// 无房间雷达时 RoomState.TotalPeople = -1，前端仅显示 bed state。
func (s *StateService) InitCardRoomAndBathroomState(ctx context.Context, cardID string, meta *CardMeta, enablement *AlarmEnablementCache) error {
	if s.writer == nil || meta == nil || len(meta.Devices) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	totalPeople := 0
	if !CardHasRoomRadar(ctx, meta, meta.TenantID, enablement) {
		totalPeople = RoomStateTotalPeopleNoRadar
	}
	roomState := &card.RoomState{UpdatedAt: now, TotalPeople: totalPeople, HasMulti: false, HasRisk: false}
	fields := PublishFields{
		RoomState: roomState,
	}
	if hasBedCapableDevice(meta) {
		fields.BedState = &card.BedState{UpdatedAt: now, BedStatus: BedStatusUnchanged, TrackNumber: 0}
	}
	bathroomDeviceID := PickBathroomDeviceIDForInit(ctx, meta, meta.TenantID, enablement)
	if bathroomDeviceID != "" {
		dm := meta.Devices[bathroomDeviceID]
		br := &card.BathRoomState{
			DeviceUID:   "",
			DeviceID:    bathroomDeviceID,
			UpdatedAt:   now,
			TotalPeople: 0,
			HasMulti:    false,
			HasRisk:     false,
		}
		if dm != nil {
			br.DeviceUID = dm.DeviceUID
			br.RoomID = dm.EffectiveRoomID
			br.RoomName = dm.BoundRoomName
		}
		fields.BathRoomState = br
	}
	return PublishCardStatus(ctx, s.writer, cardID, fields)
}

// EnsureCardStatePrepared 收到任意 iot 流时调用：若该卡 Redis 中无 BedState/RoomState 则写占位，保证下游有完整状态。每卡每进程只执行一次。
func (s *StateService) EnsureCardStatePrepared(ctx context.Context, cardID string, meta *CardMeta) error {
	if s.writer == nil || cardID == "" || meta == nil || len(meta.Devices) == 0 {
		return nil
	}
	s.preparedMu.Lock()
	_, done := s.preparedCards[cardID]
	s.preparedMu.Unlock()
	curr, _ := s.reader.ReadCardStatus(ctx, cardID)
	if done {
		if curr != nil && curr.BedState != nil {
			bs := curr.BedState
			if bs.BedStatus != 0 || bs.TrackNumber != 0 || bs.BedEvent != BedEventNone {
				return nil
			}
		}
	}
	needRoom := curr == nil || curr.RoomState == nil
	// bed-capable 才补 BedState；非 bed-capable 卡（纯雷达 UnitCard 覆盖 Bathroom 等）跳过，
	// 避免长期 bed_status=8 占位被 FE 误判 OOB。
	bedCapable := hasBedCapableDevice(meta)
	needBed := bedCapable && (curr == nil || curr.BedState == nil)
	if bedCapable && !needBed && curr != nil && curr.BedState != nil {
		bs := curr.BedState
		if bs.BedStatus == 0 && bs.TrackNumber == 0 && bs.BedEvent == BedEventNone {
			needBed = true
		}
	}
	if !needRoom && !needBed {
		s.preparedMu.Lock()
		s.preparedCards[cardID] = struct{}{}
		s.preparedMu.Unlock()
		return nil
	}
	now := time.Now().UnixMilli()
	fields := PublishFields{}
	if needRoom {
		totalPeople := 0
		if !CardHasRoomRadar(ctx, meta, meta.TenantID, nil) {
			totalPeople = RoomStateTotalPeopleNoRadar
		}
		fields.RoomState = &card.RoomState{UpdatedAt: now, TotalPeople: totalPeople, HasMulti: false, HasRisk: false}
	}
	if needBed {
		bed := &card.BedState{UpdatedAt: now, BedStatus: BedStatusUnchanged, TrackNumber: 0, StartTime: now, DurationSec: card.BedStateDurationNotSet}
		if curr != nil && curr.BedState != nil {
			bed.UpdatedAt = curr.BedState.UpdatedAt
			bed.SleepStage = curr.BedState.SleepStage
			bed.SleepConfidence = curr.BedState.SleepConfidence
			bed.StartTime = curr.BedState.StartTime
			bed.DurationSec = curr.BedState.DurationSec
		}
		if bed.UpdatedAt == 0 {
			bed.UpdatedAt = now
		}
		if bed.StartTime == 0 {
			bed.StartTime = now
		}
		fields.BedState = bed
	}
	if err := PublishCardStatus(ctx, s.writer, cardID, fields); err != nil {
		return err
	}
	s.preparedMu.Lock()
	s.preparedCards[cardID] = struct{}{}
	s.preparedMu.Unlock()
	return nil
}

// ClearBedStateSleepStage 清掉 SleepStage（事件式：LeftBed/ExitRoom/EnterRoom/device clear 触发）。
// SleepStage 是事件驱动状态——Sleepad 仅在状态变化或重连时上报，Radar 不发独立 sleep 事件。
// 不能用 TTL（两类设备 sample 周期不同），必须按生命周期事件主动清。
// 同时把 SleepConfidence 重置为 0，让下一个事件无论从哪台设备来都能写入。
func (s *StateService) ClearBedStateSleepStage(ctx context.Context, cardID string) error {
	if s.writer == nil || cardID == "" {
		return nil
	}
	now := time.Now().UnixMilli()
	var cur *card.BedState
	if s.reader != nil {
		if curr, err := s.reader.ReadCardStatus(ctx, cardID); err == nil && curr != nil && curr.BedState != nil {
			cur = &card.BedState{}
			*cur = *curr.BedState
		}
	}
	if cur == nil {
		return nil // 没 BedState 不需要清
	}
	if cur.SleepStage == 0 && cur.SleepConfidence == 0 {
		return nil // 已清，幂等
	}
	cur.UpdatedAt = now
	cur.SleepStage = 0
	cur.SleepConfidence = 0
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{BedState: cur})
}

// PublishBedStateSleepStage 仅更新 BedState.SleepStage（Sleepad/Radar 事件写入）。SleepConfidence 100 分制：60=Radar, 90=Sleepad, 100=双设备；仅更高或同级覆盖。
func (s *StateService) PublishBedStateSleepStage(ctx context.Context, cardID string, sleepStage int, confidence int) error {
	if s.writer == nil || cardID == "" {
		return nil
	}
	now := time.Now().UnixMilli()
	var cur *card.BedState
	if s.reader != nil {
		if curr, err := s.reader.ReadCardStatus(ctx, cardID); err == nil && curr != nil && curr.BedState != nil {
			cur = &card.BedState{}
			*cur = *curr.BedState
		}
	}
	if cur == nil {
		cur = &card.BedState{UpdatedAt: now, StartTime: now, DurationSec: card.BedStateDurationNotSet, BedEvent: BedEventNone, BedStatus: BedStatusUnchanged, TrackNumber: 0}
	}
	if confidence < cur.SleepConfidence {
		return nil
	}
	cur.UpdatedAt = now
	cur.SleepStage = sleepStage
	cur.SleepConfidence = confidence
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{BedState: cur})
}

// **2026-05-14 zone engine cutover**：DeriveBedStateFromRealtime 已删 — vital 累积
// N 轮推导上下床的逻辑由 zone engine 的 vital adapter（HR>0+RR>0 sustain，scorer.go
// case "sustain"）替代：sustain 不直接翻转状态而是维持 score 防 enter decay。失去
// "事件丢失时 vital 兜底纠偏"能力，靠 sleepace/radar 显式 InBed/LeftBed event 兜底。
//
// 同期删的 dead helpers（仅 DeriveBedStateFromRealtime 内部使用）：
//   - deriveStartTime / deriveConfidenceFromSnapshot / deviceVitalConfidenceInSnapshot
//   - deviceTrackConfidenceInSnapshot / sleepadDeviceFullyInitializing
//   - sleepadTrackConfidenceForInBedDerivation / sleepadInBedConfFromSnapshot
//   - deviceBedStatusInSnapshot / deviceSignalQualityInSnapshot / deviceHasHROrRRInSnapshot
//
// 保留的 helpers（仍被 SleepadTrackCountFromSnapshot 等外部调用使用）：
//   sleepadTrackInitializing / SleepadTrackCountFromSnapshot / intFromAny

func sleepadTrackInitializing(fields map[string]any) bool {
	if fields == nil {
		return false
	}
	raw, ok := fields[observation.FieldInitStatus]
	if !ok {
		return false
	}
	return intFromAny(raw) == 1
}


// SleepadTrackCountFromSnapshot 从 snapshot 统计床绑 Sleepad 设备中 bed_status=0（在床）的 track 数，leftright 最多 2。无 Sleepad 或无数据返回 0。
func SleepadTrackCountFromSnapshot(snap *CardSnapshot, meta *CardMeta, bedDevs []string) int {
	if snap == nil || meta == nil || len(bedDevs) == 0 {
		return 0
	}
	var count int
	for _, devID := range bedDevs {
		dm := meta.Devices[devID]
		if dm == nil {
			continue
		}
		t := strings.ToLower(dm.DeviceType)
		if !strings.Contains(t, "sleepad") && !strings.Contains(t, "sleeppad") {
			continue
		}
		for _, d := range snap.Devices {
			if d.DeviceID != devID && d.DeviceUID != devID {
				continue
			}
			for _, fields := range d.Tracks {
				if sleepadTrackInitializing(fields) {
					continue
				}
				raw, ok := fields[observation.FieldBedStatus]
				if !ok {
					continue
				}
				bs := intFromAny(raw)
				if bs == 0 {
					count++
					if count >= 2 {
						return 2
					}
				}
			}
			break
		}
	}
	return count
}

func intFromAny(v any) int {
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

// SetDeviceOnline 写 device:status:{deviceID} Hash 的部分更新。
// online=true：offline=0 + last_seen_ms=now（正向维护——monitor/event/recover 流共用此入口）
// online=false：offline=1（不动 last_seen_ms——alarm Offline / 看门狗显式负向标记）
func (s *StateService) SetDeviceOnline(ctx context.Context, deviceID, deviceUID, deviceType string, online bool) error {
	return s.writer.SetDeviceOnline(ctx, deviceID, deviceUID, deviceType, online)
}

// PatchDeviceFlag Phase A 后续：把单个 device-class 标志位（signal_poor / angle_abnormal / sensor_detached）
// 同步到 device:status:{deviceID}，由 alarm_handler 在收到对应 alarm/recover 时联动调用。
// 这样前端轮询 device:status hash 能拿到当前真值；不再依赖 alarm_events 表反向推导。
func (s *StateService) PatchDeviceFlag(ctx context.Context, deviceID, fieldKey string, value int) error {
	if deviceID == "" || fieldKey == "" {
		return nil
	}
	return s.writer.PatchDeviceStatus(ctx, deviceID, map[string]interface{}{
		fieldKey: fmt.Sprintf("%d", value),
	})
}

// objectRoom 统一 Room/BathRoom 的更新字段（RoomState 无 Stay；BathRoomState.StaySec 在写回后按 LastEnterTime 墙钟补算）。
type objectRoom struct {
	StandingContinuousMin int
	HasMulti              bool
	HasRisk               bool
	TotalPeople           int
	LastEnterTime         int64
	LastExitTime          int64
	UpdatedAt             int64
}

// UpdateStateFromActivity Activity 事件更新 RoomState/BathRoomState（站立、多人、Stay）和 Target（活动时间、访客）。
// 行走/站立/多人/访客阈值见 weights.go（WalkDistanceMetersThreshold、WalkSecThresholdOR、StandingContinuousSec 等）。
// 使用 objectRoom 统一读写；仅当 ObjectRoomPush/ObjectTargetPush 为 true 时写回。返回 lastEnterTime, lastExitTime 供调用方做 Stay 计时告警。
func (s *StateService) UpdateStateFromActivity(
	ctx context.Context,
	cardID, deviceID, deviceUID string,
	isBathroom bool,
	walkDuration, walkDistance, standDuration, multiPersonDuration int,
	eventTs int64,
	roomID, roomName string,
) (lastEnterTime, lastExitTime int64, err error) {
	if s.writer == nil || s.reader == nil {
		return 0, 0, nil
	}
	now := time.Now().UnixMilli()
	curr, err := s.reader.ReadCardStatus(ctx, cardID)
	if err != nil {
		return 0, 0, err
	}

	var roomState *card.RoomState
	var bathRoomState *card.BathRoomState
	if isBathroom {
		if curr != nil && curr.BathRoomState != nil {
			br := curr.BathRoomState
			match := (br.DeviceID != "" && br.DeviceID == deviceID) || (br.DeviceUID != "" && br.DeviceUID == deviceUID)
			if !match {
				return 0, 0, nil
			}
			bathRoomState = &card.BathRoomState{}
			*bathRoomState = *br
		} else {
			bathRoomState = &card.BathRoomState{DeviceID: deviceID, DeviceUID: deviceUID, UpdatedAt: eventTs, TotalPeople: 0, HasMulti: false, HasRisk: false}
		}
		bathRoomState.UpdatedAt = eventTs
		if roomID != "" {
			bathRoomState.RoomID = roomID
		}
		if roomName != "" {
			bathRoomState.RoomName = roomName
		}
	} else {
		if curr != nil && curr.RoomState != nil {
			roomState = &card.RoomState{}
			*roomState = *curr.RoomState
		}
		if roomState == nil {
			roomState = &card.RoomState{UpdatedAt: eventTs, TotalPeople: 0, HasMulti: false, HasRisk: false}
		}
		roomState.UpdatedAt = eventTs
	}

	// objectRoom：统一取 StandingContinuousMin, HasMulti, HasRisk, TotalPeople, LastEnter/Exit, UpdatedAt（不含 Stay；Room 无 Stay）
	obj := objectRoom{}
	if isBathroom && bathRoomState != nil {
		obj.StandingContinuousMin = bathRoomState.StandingContinuousMin
		obj.HasMulti = bathRoomState.HasMulti
		obj.HasRisk = bathRoomState.HasRisk
		obj.TotalPeople = bathRoomState.TotalPeople
		obj.LastEnterTime = bathRoomState.LastEnterTime
		obj.LastExitTime = bathRoomState.LastExitTime
		obj.UpdatedAt = bathRoomState.UpdatedAt
	} else if roomState != nil {
		obj.StandingContinuousMin = roomState.StandingContinuousMin
		obj.HasMulti = roomState.HasMulti
		obj.HasRisk = roomState.HasRisk
		obj.TotalPeople = roomState.TotalPeople
		obj.LastEnterTime = roomState.LastEnterTime
		obj.LastExitTime = roomState.LastExitTime
		obj.UpdatedAt = roomState.UpdatedAt
	}

	var objectRoomPush, objectTargetPush bool

	var target *card.TargetState
	if curr != nil && curr.Target != nil {
		t := *curr.Target
		target = &t
		if target.VisitorStartTs == 0 {
			target.VisitorStartTs = -1
		}
	} else {
		// 新建 Target：推断出的活动/访客，无具体 track，用 9(未知人) 见 observation.TrackUnknownPerson，VisitorStartTs: -1表示无
		target = &card.TargetState{TrackID: observation.TrackUnknownPerson, UpdatedAt: now, VisitorStartTs: -1}
	}

	// 1. 行走达标 → 更新 Target（距离≥2m 或 时长≥WalkSecThresholdOR，见 weights）
	if walkDistance >= WalkDistanceMetersThreshold || walkDuration >= WalkSecThresholdOR {
		target.LastActiveTs = eventTs
		target.UpdatedAt = now
		objectTargetPush = true
	}

	// 2. 站立：阈值见 weights.StandingContinuousSec=55 / StandingContinuousMin
	if standDuration >= StandingContinuousSec {
		obj.StandingContinuousMin++
	} else {
		obj.StandingContinuousMin = 0
	}
	if obj.StandingContinuousMin >= StandingContinuousMin {
		obj.StandingContinuousMin = StandingContinuousMin
		objectRoomPush = true
	}

	// 3. 多人：阈值见 weights.MultiPersonDurationSec, multi_person_duration >= 30：TotalPeople<=1 → 2, HasMulti=true, HasRisk=false；后续可能再更新 HasMulti/HasRisk
	if multiPersonDuration >= MultiPersonDurationSec {
		if obj.TotalPeople <= 1 {
			obj.TotalPeople = 2
			obj.HasMulti = true
			obj.HasRisk = false
			objectRoomPush = true
		}
	}

	// 写回 objectRoom → RoomState/BathRoomState；StaySec：有人且已记录进入时间则为 (now-LastEnter)/1s
	if isBathroom && bathRoomState != nil {
		bathRoomState.StandingContinuousMin = obj.StandingContinuousMin
		bathRoomState.HasMulti = obj.HasMulti
		bathRoomState.HasRisk = obj.HasRisk
		bathRoomState.TotalPeople = obj.TotalPeople
		bathRoomState.LastEnterTime = obj.LastEnterTime
		bathRoomState.LastExitTime = obj.LastExitTime
		bathRoomState.UpdatedAt = obj.UpdatedAt
		if bathRoomState.TotalPeople > 0 && bathRoomState.LastEnterTime > 0 {
			sec := int((now - bathRoomState.LastEnterTime) / 1000)
			if sec < 0 {
				sec = 0
			}
			if bathRoomState.StaySec != sec {
				bathRoomState.StaySec = sec
				objectRoomPush = true
			}
		} else if bathRoomState.StaySec != 0 {
			bathRoomState.StaySec = 0
			objectRoomPush = true
		}
	} else if roomState != nil {
		roomState.StandingContinuousMin = obj.StandingContinuousMin
		roomState.HasMulti = obj.HasMulti
		roomState.HasRisk = obj.HasRisk
		roomState.TotalPeople = obj.TotalPeople
		roomState.LastEnterTime = obj.LastEnterTime
		roomState.LastExitTime = obj.LastExitTime
		roomState.UpdatedAt = obj.UpdatedAt
	}
	lastEnterTime = obj.LastEnterTime
	lastExitTime = obj.LastExitTime

	// 4. Target 访客逻辑：阈值见 weights.VisitorMinThreshold
	if multiPersonDuration >= MultiPersonDurationSec {
		if target.VisitorStartTs == 0 || target.VisitorStartTs == -1 {
			target.VisitorStartTs = eventTs
		} else {
			visitMin := int((now - target.VisitorStartTs) / (60 * 1000))
			if visitMin >= VisitorMinThreshold {
				target.HasVisitorToday = true
			}
		}
		objectTargetPush = true
	} else {
		if target.VisitorStartTs != 0 && target.VisitorStartTs != -1 {
			visitLong := now - target.VisitorStartTs
			visitMin := int(visitLong / (60 * 1000))
			if visitMin > target.TodayMaxVisitorMin {
				target.TodayMaxVisitorMin = visitMin
			}
			target.VisitorStartTs = -1
		}
	}
	target.UpdatedAt = now

	fields := PublishFields{}
	if objectTargetPush {
		fields.Target = target
	}
	if objectRoomPush {
		if isBathroom {
			fields.BathRoomState = bathRoomState
		} else {
			fields.RoomState = roomState
		}
	}
	if fields.Target != nil || fields.RoomState != nil || fields.BathRoomState != nil {
		return lastEnterTime, lastExitTime, PublishCardStatus(ctx, s.writer, cardID, fields)
	}
	return lastEnterTime, lastExitTime, nil
}
