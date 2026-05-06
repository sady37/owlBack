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

// vitalDeriveRound 单轮：HR/RR 置信度 + Sleepad InBedStatus 推断（inBedConf，正=在床负=离床）。
type vitalDeriveRound struct {
	confidence int // vital 置信度
	inBedConf  int // Sleepad：inBedStatus=0 → +TrackConfidence，=1 → -TrackConfidence
}

// vitalDeriveRounds 当 bed_status!=0/1 时累计 N 轮 track 的 HR/RR 用于推导在床/离床，N 见 weights.VitalDeriveRoundsNeeded。
type vitalDeriveRounds struct {
	rounds []vitalDeriveRound
}

type StateService struct {
	writer *card.Writer
	reader *card.Reader
	logger *zap.Logger

	vitalDeriveMu     sync.Mutex
	vitalDeriveRounds map[string]*vitalDeriveRounds // cardID -> N 轮累计（见 weights.VitalDeriveRoundsNeeded）N=5
	// recoverInBedRounds BedStatus=1 时由 Sleepad 实时强信号累计，满阈纠回在床（与 vitalDeriveRounds 隔离）
	recoverInBedRounds map[string]*vitalDeriveRounds
	leftBedCooldownUntil map[string]int64 // cardID -> 冷却结束时间 ms，期内禁止 recover 纠回在床
	preparedMu        sync.Mutex
	preparedCards     map[string]struct{} // cardID 已执行过 EnsureCardStatePrepared
}

func NewStateService(writer *card.Writer, reader *card.Reader, logger *zap.Logger) *StateService {
	return &StateService{
		writer:               writer,
		reader:               reader,
		logger:               logger,
		vitalDeriveRounds:    make(map[string]*vitalDeriveRounds),
		recoverInBedRounds:   make(map[string]*vitalDeriveRounds),
		leftBedCooldownUntil: make(map[string]int64),
		preparedCards:        make(map[string]struct{}),
	}
}

// NoteLeftBedCooldown 离床落库后调用：在 LeftBedDeriveCooldownMs 内禁止 Derive 将 BedStatus=1 纠回在床。
func (s *StateService) NoteLeftBedCooldown(cardID string) {
	if s == nil || cardID == "" {
		return
	}
	until := time.Now().UnixMilli() + LeftBedDeriveCooldownMs
	s.vitalDeriveMu.Lock()
	if s.leftBedCooldownUntil == nil {
		s.leftBedCooldownUntil = make(map[string]int64)
	}
	s.leftBedCooldownUntil[cardID] = until
	delete(s.recoverInBedRounds, cardID)
	s.vitalDeriveMu.Unlock()
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

// UpdateTargetLastActive 当 moveSec >= moveSecThreshold 或 danceMin >= danceMinThreshold 时更新 Target.LastActiveTs。阈值由调用方传入（如 10/20 秒，1/2 分钟）。
func (s *StateService) UpdateTargetLastActive(ctx context.Context, cardID string, tsMs int64, moveSec, moveSecThreshold int, danceMin, danceMinThreshold int) error {
	if s.writer == nil || s.reader == nil || cardID == "" {
		return nil
	}
	if moveSecThreshold <= 0 && danceMinThreshold <= 0 {
		return nil
	}
	if moveSec < moveSecThreshold && (danceMinThreshold <= 0 || danceMin < danceMinThreshold) {
		return nil
	}
	curr, err := s.reader.ReadCardStatus(ctx, cardID)
	if err != nil || curr == nil {
		return nil
	}
	target := curr.Target
	if target == nil {
		// 推断出的活动，无具体 track，用 9(未知人) 见 observation.TrackUnknownPerson
		target = &card.TargetState{TrackID: observation.TrackUnknownPerson, UpdatedAt: time.Now().UnixMilli()}
	} else {
		t := *target
		target = &t
	}
	target.LastActiveTs = tsMs
	target.UpdatedAt = time.Now().UnixMilli()
	return PublishCardStatus(ctx, s.writer, cardID, PublishFields{Target: target})
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

// PublishBedStateFromEvent 在 event_handler 收到 InBed/LeftBed 时写 BedState。上床/下床彻底分开：有 Sleepad 用 buffer 在床轨数，无则事件 ±1 限制 [0,2]。UpdatedAt=事件时间，BedEvent：0=InBed，1=LeftBed。
func (s *StateService) PublishBedStateFromEvent(ctx context.Context, cardID, eventName, deviceType string, timestampMs int64, _ int, trackNumberOverride *int) (written bool, err error) {
	switch eventName {
	case alarm.InBed:
		return s.publishBedStateInBed(ctx, cardID, deviceType, timestampMs, trackNumberOverride)
	case alarm.LeftBed:
		return s.publishBedStateLeftBed(ctx, cardID, deviceType, timestampMs, trackNumberOverride)
	default:
		return false, nil
	}
}

// publishBedStateInBed 上床简化版：有 Sleepad 用 buffer 在床轨数+Sleepad 置信度；无则 prev+1 cap 2+Radar 置信度，StartTime 若 prev 在床沿用否则 event.ts。
func (s *StateService) publishBedStateInBed(ctx context.Context, cardID, deviceType string, timestampMs int64, trackNumberOverride *int) (written bool, err error) {
	var prev *card.BedState
	if s.reader != nil {
		if curr, readErr := s.reader.ReadCardStatus(ctx, cardID); readErr == nil && curr != nil && curr.BedState != nil {
			prev = curr.BedState
		}
	}
	bedFromPrev := func(b *card.BedState) {
		if prev != nil {
			b.SleepStage = prev.SleepStage
			b.SleepConfidence = prev.SleepConfidence
		}
	}

	var trackNumber int
	var startTime int64
	var confidence int
	if trackNumberOverride != nil {
		confidence = BedConfidenceSleepadBase
		trackNumber = *trackNumberOverride
		if trackNumber < 0 {
			trackNumber = 0
		}
		if trackNumber > trackNumberMax {
			trackNumber = trackNumberMax
		}
		if prev != nil && prev.BedStatus == 0 {
			startTime = prev.StartTime
			if startTime == 0 && prev.UpdatedAt > 0 {
				startTime = prev.UpdatedAt
			}
		} else {
			startTime = timestampMs
		}
	} else {
		confidence = BedConfidenceRadarBase
		if prev == nil || prev.BedStatus != 0 {
			trackNumber = 1
			startTime = timestampMs
		} else {
			trackNumber = prev.TrackNumber + 1
			if trackNumber > trackNumberMax {
				trackNumber = trackNumberMax
			}
			startTime = prev.StartTime
			if startTime == 0 && prev.UpdatedAt > 0 {
				startTime = prev.UpdatedAt
			}
		}
	}
	if startTime == 0 {
		startTime = timestampMs
	}
	status := bedStatusFromTrackNumber(trackNumber)
	durationSec := int((timestampMs - startTime) / 1000)
	if durationSec < 0 {
		durationSec = 0
	}
	bs := &card.BedState{
		UpdatedAt:     timestampMs,
		BedStatus:     status,
		TrackNumber:   trackNumber,
		StartTime:     startTime,
		DurationSec:   durationSec,
		BedConfidence: confidence,
		BedEvent:      BedEventInBed,
	}
	bedFromPrev(bs)
	err = PublishCardStatus(ctx, s.writer, cardID, PublishFields{BedState: bs})
	return err == nil, err
}

// publishBedStateLeftBed 离床简化版：有 Sleepad 用 override+Sleepad 置信度；无则 prev-1 floor 0+Radar 置信度；TotalPeople==0 时 StartTime=event.ts。

func (s *StateService) publishBedStateLeftBed(ctx context.Context, cardID, deviceType string, timestampMs int64, trackNumberOverride *int) (written bool, err error) {
	var prev *card.BedState
	if s.reader != nil {
		if curr, readErr := s.reader.ReadCardStatus(ctx, cardID); readErr == nil && curr != nil && curr.BedState != nil {
			prev = curr.BedState
		}
	}
	// LeftBed 不再继承 SleepStage——离床即清，下一次 InBed 后等 Sleepad/Radar 再上报
	// （SleepStage 是事件，不是连续量；持有 stale 值会导致"Sleepad 关了仍显示 Awake"）
	bedFromPrev := func(b *card.BedState) {
		_ = prev // 占位避免未使用警告
		b.SleepStage = 0
		b.SleepConfidence = 0
	}

	var trackNumber int
	var startTime int64
	var confidence int
	if trackNumberOverride != nil {
		confidence = BedConfidenceSleepadBase
		trackNumber = *trackNumberOverride
		if trackNumber < 0 {
			trackNumber = 0
		}
		if trackNumber > trackNumberMax {
			trackNumber = trackNumberMax
		}
		if trackNumber == 0 {
			startTime = timestampMs
		} else if prev != nil {
			startTime = prev.StartTime
			if startTime == 0 && prev.UpdatedAt > 0 {
				startTime = prev.UpdatedAt
			}
		} else {
			startTime = timestampMs
		}
	} else {
		confidence = BedConfidenceRadarBase
		if prev != nil && prev.TrackNumber > 0 {
			trackNumber = prev.TrackNumber - 1
		}
		if trackNumber < 0 {
			trackNumber = 0
		}
		if trackNumber == 0 {
			startTime = timestampMs
		} else if prev != nil {
			startTime = prev.StartTime
			if startTime == 0 && prev.UpdatedAt > 0 {
				startTime = prev.UpdatedAt
			}
		} else {
			startTime = timestampMs
		}
	}
	if startTime == 0 {
		startTime = timestampMs
	}
	status := bedStatusFromTrackNumber(trackNumber)
	durationSec := int((timestampMs - startTime) / 1000)
	if durationSec < 0 {
		durationSec = 0
	}
	bs := &card.BedState{
		UpdatedAt:     timestampMs,
		BedStatus:     status,
		TrackNumber:   trackNumber,
		StartTime:     startTime,
		DurationSec:   durationSec,
		BedConfidence: confidence,
		BedEvent:      BedEventLeftBed,
	}
	bedFromPrev(bs)
	err = PublishCardStatus(ctx, s.writer, cardID, PublishFields{BedState: bs})
	written = err == nil
	if written {
		// Sleepad 与 monitor 强一致：override==0 表示缓冲里已无在床轨，立即离床 → 不启 Derive 纠回在床冷却。
		// 不一致经 pending 超时或非 Sleepad 路径（override nil）仍设冷却，抑制第二层误纠偏。
		strongSleepadAligned := trackNumberOverride != nil && *trackNumberOverride == 0
		if !strongSleepadAligned {
			s.NoteLeftBedCooldown(cardID)
		}
	}
	return written, err
}

const trackNumberMax = 2

// computeBedStateUpdate 返回 (startTime, trackNumber)。TrackNumber 限制在 [0, trackNumberMax]。
func computeBedStateUpdate(eventName string, prev *card.BedState, timestampMs int64) (startTime int64, trackNumber int) {
	switch eventName {
	case alarm.InBed:
		if prev == nil || prev.BedStatus != 0 {
			return timestampMs, 1 //inbed, 且原来不在床1或未知8，tracknumber=1
		}
		trackNumber = prev.TrackNumber + 1
		if trackNumber > trackNumberMax {
			trackNumber = trackNumberMax
		}
		startTime = prev.StartTime
		return startTime, trackNumber
	case alarm.LeftBed:
		if prev != nil && prev.TrackNumber > 0 {
			trackNumber = prev.TrackNumber - 1
		}
		if trackNumber < 0 {
			trackNumber = 0
		}
		if trackNumber == 0 {
			startTime = timestampMs
		} else if prev != nil {
			startTime = prev.StartTime
		} else {
			startTime = timestampMs
		}
		return startTime, trackNumber
	default:
		return timestampMs, 0
	}
}

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

// ReconcileRoomStateFromBedState 若 sum(AreaPeople) < BedState.TrackNumber，则用床上人数抬升 TotalPeople 并写回；AreaPeople 保留。
func (s *StateService) ReconcileRoomStateFromBedState(ctx context.Context, cardID string) error {
	if s.reader == nil || s.writer == nil {
		return nil
	}
	curr, err := s.reader.ReadCardStatus(ctx, cardID)
	if err != nil || curr == nil || curr.BedState == nil || curr.RoomState == nil {
		return nil
	}
	rs := curr.RoomState
	bs := curr.BedState
	if rs.TotalPeople < 0 {
		return nil
	}
	sum := sumAreaPeople(rs.AreaPeople)
	if sum >= bs.TrackNumber {
		return nil
	}
	cur := &card.RoomState{
		UpdatedAt:     bs.UpdatedAt,
		TotalPeople:   bs.TrackNumber,
		LastEnterTime: rs.LastEnterTime,
		LastExitTime:  rs.LastExitTime,
	}
	if rs.AreaPeople != nil {
		cur.AreaPeople = rs.AreaPeople
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

// DeriveAndWriteState 在 derive 定时点执行：写 Target；并把每台设备的实时在线状态写到 device:status:{deviceID} 独立 Hash。
// Phase A：DeviceStatus 已迁出 card:state（每设备一个 Hash）；不再合并跨调用，仅写本轮 derive 命中的设备。
func (s *StateService) DeriveAndWriteState(
	ctx context.Context,
	snap CardSnapshot,
	meta *CardMeta,
	prevTarget *card.TargetState,
	buf *MonitorBuffer,
) (*card.CardStatus, error) {
	var devices map[string]*DeviceMeta
	if meta != nil {
		devices = meta.Devices
	}
	fromTracker := buf.BuildDeviceStatus(snap.CardID, devices, time.Now().UnixMilli())
	for _, ds := range fromTracker {
		if err := s.writer.WriteDeviceStatus(ctx, ds); err != nil {
			s.logger.Warn("write device status",
				zap.String("device_id", ds.DeviceID),
				zap.Error(err))
		}
	}

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
		zap.Bool("has_target", prevTarget != nil),
		zap.Int("devices", len(fromTracker)))

	return status, nil
}

// InitCardRoomAndBathroomState Card 初始化时：默认创建 RoomState；存在 isBathroomRadar 时创建 BathRoomState；无上下床事件时写入占位 BedState(bed_status=8)。无房间雷达时 RoomState.TotalPeople = -1，前端仅显示 bed state。
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
		BedState:  &card.BedState{UpdatedAt: now, BedStatus: BedStatusUnchanged, TrackNumber: 0},
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
	needBed := curr == nil || curr.BedState == nil
	if !needBed && curr != nil && curr.BedState != nil {
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

// DeriveBedStateFromRealtime 当 bed_status！=0/1 时在 derive 定时点调用：每轮累加置信度，累加和>350 即发布在床（BedConfidence=sum/n）并 Reconcile；满 5 轮仍<350 则发布离床（BedStatus=1，BedConfidence=100-sum/5）并 Reconcile。
// deriveStartTime 返回 StartTime：状态未变时继承 prev，状态变化时用 now。
func deriveStartTime(now int64, newBedStatus int, prev *card.BedState) int64 {
	if prev != nil && prev.BedStatus == newBedStatus && prev.StartTime > 0 {
		return prev.StartTime
	}
	return now
}

func (s *StateService) DeriveBedStateFromRealtime(ctx context.Context, snap CardSnapshot, meta *CardMeta) {
	if meta == nil || meta.BedID == "" {
		return
	}
	bedDevs := meta.BedBoundDeviceIDs()
	if len(bedDevs) == 0 {
		return
	}
	curr, err := s.reader.ReadCardStatus(ctx, snap.CardID)
	if err != nil {
		s.vitalDeriveMu.Lock()
		delete(s.vitalDeriveRounds, snap.CardID)
		delete(s.recoverInBedRounds, snap.CardID)
		s.vitalDeriveMu.Unlock()
		return
	}

	hasSleepadOnBed := false
	for _, devID := range bedDevs {
		if dm := meta.Devices[devID]; dm != nil {
			t := strings.ToLower(dm.DeviceType)
			if strings.Contains(t, "sleepad") || strings.Contains(t, "sleeppad") {
				hasSleepadOnBed = true
				break
			}
		}
	}

	// 在床(0)：清空累计
	if curr != nil && curr.BedState != nil && curr.BedState.BedStatus == 0 {
		s.vitalDeriveMu.Lock()
		delete(s.vitalDeriveRounds, snap.CardID)
		delete(s.recoverInBedRounds, snap.CardID)
		s.vitalDeriveMu.Unlock()
		return
	}

	// 离床(1)：Sleepad 床且非 LeftBed 冷却 → 用实时累计尝试纠回在床（与未定态推导隔离）
	if curr != nil && curr.BedState != nil && curr.BedState.BedStatus == 1 {
		nowMs := time.Now().UnixMilli()
		s.vitalDeriveMu.Lock()
		coolUntil := int64(0)
		if s.leftBedCooldownUntil != nil {
			coolUntil = s.leftBedCooldownUntil[snap.CardID]
		}
		inCooldown := nowMs < coolUntil
		if !hasSleepadOnBed || inCooldown {
			delete(s.vitalDeriveRounds, snap.CardID)
			delete(s.recoverInBedRounds, snap.CardID)
			s.vitalDeriveMu.Unlock()
			return
		}
		inBedConf := sleepadInBedConfFromSnapshot(&snap, meta, bedDevs)
		rv := s.recoverInBedRounds[snap.CardID]
		if rv == nil {
			rv = &vitalDeriveRounds{rounds: make([]vitalDeriveRound, 0, VitalDeriveRoundsNeeded)}
			s.recoverInBedRounds[snap.CardID] = rv
		}
		rv.rounds = append(rv.rounds, vitalDeriveRound{inBedConf: inBedConf})
		var sumRecover int
		for _, r := range rv.rounds {
			sumRecover += r.inBedConf
		}
		nr := len(rv.rounds)
		if nr >= VitalDeriveRoundsNeeded {
			if sumRecover >= DeriveBedStateThreshold {
				delete(s.recoverInBedRounds, snap.CardID)
				delete(s.vitalDeriveRounds, snap.CardID)
				s.vitalDeriveMu.Unlock()
				now := time.Now().UnixMilli()
				bedConf := sumRecover / VitalDeriveRoundsNeeded
				bs := &card.BedState{
					UpdatedAt:     now,
					BedStatus:     0,
					TrackNumber:   1,
					StartTime:     deriveStartTime(now, 0, curr.BedState),
					DurationSec:   0,
					BedEvent:      BedEventNone,
					BedConfidence: bedConf,
					SleepStage:    curr.BedState.SleepStage,
					SleepConfidence: curr.BedState.SleepConfidence,
				}
				if err := PublishCardStatus(ctx, s.writer, snap.CardID, PublishFields{BedState: bs}); err != nil {
					s.logger.Warn("recover in-bed from derive", zap.String("cid", snap.CardID), zap.Error(err))
					return
				}
				_ = s.ReconcileRoomStateFromBedState(ctx, snap.CardID)
				s.logger.Debug("derive recover bed_state=0 (Sleepad sumInBedConf)", zap.String("cid", snap.CardID), zap.Int("sum", sumRecover))
				return
			}
			if sumRecover <= -DeriveBedStateThreshold {
				delete(s.recoverInBedRounds, snap.CardID)
				s.vitalDeriveMu.Unlock()
				return
			}
			delete(s.recoverInBedRounds, snap.CardID)
			s.vitalDeriveMu.Unlock()
			return
		}
		s.vitalDeriveMu.Unlock()
		return
	}

	s.vitalDeriveMu.Lock()
	delete(s.recoverInBedRounds, snap.CardID)
	s.vitalDeriveMu.Unlock()

	// nil / 8 / 其它：继续 N=5 轮 vital 推导
	roundConf := deriveConfidenceFromSnapshot(&snap, meta, bedDevs)
	s.vitalDeriveMu.Lock()
	v := s.vitalDeriveRounds[snap.CardID]
	if v == nil {
		v = &vitalDeriveRounds{rounds: make([]vitalDeriveRound, 0, VitalDeriveRoundsNeeded)}
		s.vitalDeriveRounds[snap.CardID] = v
	}
	inBedConf := sleepadInBedConfFromSnapshot(&snap, meta, bedDevs)
	v.rounds = append(v.rounds, vitalDeriveRound{confidence: roundConf, inBedConf: inBedConf})
	var sumConf, sumInBedConf int
	for _, r := range v.rounds {
		sumConf += r.confidence
		sumInBedConf += r.inBedConf
	}
	n := len(v.rounds)

	if hasSleepadOnBed && n >= VitalDeriveRoundsNeeded {
		if sumInBedConf >= DeriveBedStateThreshold {
			delete(s.vitalDeriveRounds, snap.CardID)
			s.vitalDeriveMu.Unlock()
			now := time.Now().UnixMilli()
			bedConf := sumInBedConf / VitalDeriveRoundsNeeded
			var ss int
			var sc int
			if curr != nil && curr.BedState != nil {
				ss = curr.BedState.SleepStage
				sc = curr.BedState.SleepConfidence
			}
			if err := PublishCardStatus(ctx, s.writer, snap.CardID, PublishFields{
				BedState: &card.BedState{
					UpdatedAt:       now,
					BedStatus:       0,
					TrackNumber:     1,
					StartTime:       deriveStartTime(now, 0, curr.BedState),
					DurationSec:     0,
					BedEvent:        BedEventNone,
					BedConfidence:   bedConf,
					SleepStage:      ss,
					SleepConfidence: sc,
				},
			}); err != nil {
				s.logger.Warn("derive bed state from realtime (Sleepad InBed)", zap.String("cid", snap.CardID), zap.Error(err))
				return
			}
			_ = s.ReconcileRoomStateFromBedState(ctx, snap.CardID)
			s.logger.Debug("derive bed_state=0 from realtime (Sleepad sumInBedConf>=350)", zap.String("cid", snap.CardID), zap.Int("sumInBedConf", sumInBedConf))
			return
		}
		if sumInBedConf <= -DeriveBedStateThreshold {
			delete(s.vitalDeriveRounds, snap.CardID)
			s.vitalDeriveMu.Unlock()
			now := time.Now().UnixMilli()
			bedConf := sumInBedConf / VitalDeriveRoundsNeeded
			if bedConf < 0 {
				bedConf = -bedConf
			}
			var ss int
			var sc int
			if curr != nil && curr.BedState != nil {
				ss = curr.BedState.SleepStage
				sc = curr.BedState.SleepConfidence
			}
			if err := PublishCardStatus(ctx, s.writer, snap.CardID, PublishFields{
				BedState: &card.BedState{
					UpdatedAt:       now,
					BedStatus:       1,
					TrackNumber:     0,
					StartTime:       deriveStartTime(now, 1, curr.BedState),
					DurationSec:     0,
					BedEvent:        BedEventNone,
					BedConfidence:   bedConf,
					SleepStage:      ss,
					SleepConfidence: sc,
				},
			}); err != nil {
				s.logger.Warn("derive bed state from realtime (Sleepad InBed)", zap.String("cid", snap.CardID), zap.Error(err))
				return
			}
			_ = s.ReconcileRoomStateFromBedState(ctx, snap.CardID)
			s.logger.Debug("derive bed_state=1 from realtime (Sleepad sumInBedConf<=-350)", zap.String("cid", snap.CardID), zap.Int("sumInBedConf", sumInBedConf))
			return
		}
		delete(s.vitalDeriveRounds, snap.CardID)
		s.vitalDeriveMu.Unlock()
		return
	}

	if sumConf >= DeriveBedStateThreshold {
		delete(s.vitalDeriveRounds, snap.CardID)
		s.vitalDeriveMu.Unlock()
		now := time.Now().UnixMilli()
		bedConf := sumConf / n
		var ss int
		var sc int
		if curr != nil && curr.BedState != nil {
			ss = curr.BedState.SleepStage
			sc = curr.BedState.SleepConfidence
		}
		if err := PublishCardStatus(ctx, s.writer, snap.CardID, PublishFields{
			BedState: &card.BedState{
				UpdatedAt:       now,
				BedStatus:       0,
				TrackNumber:     1,
				StartTime:       deriveStartTime(now, 0, curr.BedState),
				DurationSec:     0,
				BedEvent:        BedEventNone,
				BedConfidence:   bedConf,
				SleepStage:      ss,
				SleepConfidence: sc,
			},
		}); err != nil {
			s.logger.Warn("derive bed state from realtime", zap.String("cid", snap.CardID), zap.Error(err))
			return
		}
		_ = s.ReconcileRoomStateFromBedState(ctx, snap.CardID)
		s.logger.Debug("derive bed_state=0 from realtime (sum>350)", zap.String("cid", snap.CardID), zap.Int("sumConf", sumConf), zap.Int("n", n))
		return
	}
	if n >= VitalDeriveRoundsNeeded {
		delete(s.vitalDeriveRounds, snap.CardID)
		s.vitalDeriveMu.Unlock()
		now := time.Now().UnixMilli()
		bedConf := max(100-sumConf/VitalDeriveRoundsNeeded, 50) //50<radar 60, 防止radar 不能更新
		if bedConf < 0 {
			bedConf = 0
		}
		var ss int
		var sc int
		if curr != nil && curr.BedState != nil {
			ss = curr.BedState.SleepStage
			sc = curr.BedState.SleepConfidence
		}
		if err := PublishCardStatus(ctx, s.writer, snap.CardID, PublishFields{
			BedState: &card.BedState{
				UpdatedAt:       now,
				BedStatus:       1,
				TrackNumber:     0,
				StartTime:       deriveStartTime(now, 1, curr.BedState),
				DurationSec:     0,
				BedEvent:        BedEventNone,
				BedConfidence:   bedConf,
				SleepStage:      ss,
				SleepConfidence: sc,
			},
		}); err != nil {
			s.logger.Warn("derive bed state from realtime", zap.String("cid", snap.CardID), zap.Error(err))
			return
		}
		_ = s.ReconcileRoomStateFromBedState(ctx, snap.CardID)
		s.logger.Debug("derive bed_state=1 from realtime (5 rounds sum<=350)", zap.String("cid", snap.CardID), zap.Int("sumConf", sumConf))
		return
	}
	s.vitalDeriveMu.Unlock()
}

// deriveConfidenceFromSnapshot 按 snapshot 判断有 HR/RR 的设备并取 track.VitalConfidence；无 vital 则置信度 0。同类型取最大，两种都有 100。
func deriveConfidenceFromSnapshot(snap *CardSnapshot, meta *CardMeta, bedDevs []string) int {
	if meta == nil || len(bedDevs) == 0 {
		return 0
	}
	var maxRadar, maxSleepad int
	for _, devID := range bedDevs {
		if !deviceHasHROrRRInSnapshot(snap, devID) {
			continue
		}
		dm := meta.Devices[devID]
		if dm == nil {
			continue
		}
		c := deviceVitalConfidenceInSnapshot(snap, devID)
		if c == 0 {
			continue
		}
		t := strings.ToLower(dm.DeviceType)
		if strings.Contains(t, "radar") {
			if c > maxRadar {
				maxRadar = c
			}
		} else if strings.Contains(t, "sleepad") || strings.Contains(t, "sleeppad") {
			if c > maxSleepad {
				maxSleepad = c
			}
		}
	}
	if maxRadar > 0 && maxSleepad > 0 {
		return 100
	}
	if maxRadar > 0 {
		return maxRadar - 10 //hr/rr有干扰，防止超过Sleepad Inbed 90的置信度
	}
	if maxSleepad > 0 {
		return maxSleepad
	}
	return 0
}

// deviceVitalConfidenceInSnapshot 该设备在 snapshot 的 track 中 VitalConfidence 最大值（网关已下发）。
func deviceVitalConfidenceInSnapshot(snap *CardSnapshot, deviceKey string) int {
	if snap == nil {
		return 0
	}
	var max int
	for _, d := range snap.Devices {
		if d.DeviceID != deviceKey && d.DeviceUID != deviceKey {
			continue
		}
		for _, fields := range d.Tracks {
			if v := intFromAny(fields[observation.FieldVitalConfidence]); v > max {
				max = v
			}
		}
		break
	}
	return max
}

// deviceTrackConfidenceInSnapshot 该设备在 snapshot 的 track 中 TrackConfidence 最大值；无则 0。
func deviceTrackConfidenceInSnapshot(snap *CardSnapshot, deviceKey string) int {
	if snap == nil {
		return 0
	}
	var max int
	for _, d := range snap.Devices {
		if d.DeviceID != deviceKey && d.DeviceUID != deviceKey {
			continue
		}
		for _, fields := range d.Tracks {
			if v := intFromAny(fields[observation.FieldTrackConfidence]); v > max {
				max = v
			}
		}
		break
	}
	return max
}

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

// sleepadDeviceFullyInitializing 至少有一条 track 且每条均为 init_status==1（厂家约定此时 bed/生命体征无效）。
func sleepadDeviceFullyInitializing(snap *CardSnapshot, deviceKey string) bool {
	if snap == nil {
		return false
	}
	foundAny := false
	for _, d := range snap.Devices {
		if d.DeviceID != deviceKey && d.DeviceUID != deviceKey {
			continue
		}
		for _, fields := range d.Tracks {
			foundAny = true
			if !sleepadTrackInitializing(fields) {
				return false
			}
		}
		break
	}
	return foundAny
}

// sleepadTrackConfidenceForInBedDerivation 仅非初始化 track：全无 track_confidence 键 → 90；有键且最大≤0 → 30；否则取已上报最大值。
func sleepadTrackConfidenceForInBedDerivation(snap *CardSnapshot, deviceKey string) int {
	if snap == nil {
		return 90
	}
	anyPresent := false
	maxConf := -1
	for _, d := range snap.Devices {
		if d.DeviceID != deviceKey && d.DeviceUID != deviceKey {
			continue
		}
		for _, fields := range d.Tracks {
			if sleepadTrackInitializing(fields) {
				continue
			}
			raw, ok := fields[observation.FieldTrackConfidence]
			if !ok {
				continue
			}
			anyPresent = true
			v := intFromAny(raw)
			if v > maxConf {
				maxConf = v
			}
		}
		break
	}
	if !anyPresent {
		return 90
	}
	if maxConf <= 0 {
		return 30
	}
	return maxConf
}

// sleepadInBedConfFromSnapshot 仅当床绑设备含 Sleepad 时：从该设备 track 读 BedStatus（0=在床 1=离床）与 TrackConfidence；inBedStatus=0 → +TrackConfidence，=1 → -TrackConfidence，无数据或非 Sleepad 返回 0。
func sleepadInBedConfFromSnapshot(snap *CardSnapshot, meta *CardMeta, bedDevs []string) int {
	if meta == nil || snap == nil || len(bedDevs) == 0 {
		return 0
	}
	for _, devID := range bedDevs {
		dm := meta.Devices[devID]
		if dm == nil {
			continue
		}
		t := strings.ToLower(dm.DeviceType)
		if !strings.Contains(t, "sleepad") && !strings.Contains(t, "sleeppad") {
			continue
		}
		if sleepadDeviceFullyInitializing(snap, devID) {
			return 0
		}
		trackConf := sleepadTrackConfidenceForInBedDerivation(snap, devID)
		bedStatus := deviceBedStatusInSnapshot(snap, devID)
		if bedStatus == 0 {
			return trackConf
		}
		if bedStatus == 1 {
			return -trackConf
		}
		return 0
	}
	return 0
}

// deviceBedStatusInSnapshot 该设备 snapshot 中 BedStatus（0/1）；跳过 init_status==1；bed_status 键缺失不当作 0。无则 -1。
func deviceBedStatusInSnapshot(snap *CardSnapshot, deviceKey string) int {
	if snap == nil {
		return -1
	}
	for _, d := range snap.Devices {
		if d.DeviceID != deviceKey && d.DeviceUID != deviceKey {
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
			v := intFromAny(raw)
			if v == 0 || v == 1 {
				return v
			}
		}
		break
	}
	return -1
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

// deviceSignalQualityInSnapshot 该设备在 snapshot 的 track 中 signal_quality 最大值（0~90）。若为 1~5 则乘 18 与 Sleepace 一致。
func deviceSignalQualityInSnapshot(snap *CardSnapshot, deviceKey string) int {
	if snap == nil {
		return 0
	}
	var max int
	for _, d := range snap.Devices {
		if d.DeviceID != deviceKey && d.DeviceUID != deviceKey {
			continue
		}
		for _, fields := range d.Tracks {
			n := intFromAny(fields[observation.FieldSignalQuality])
			if n >= 1 && n <= 5 {
				n = n * 18
			}
			if n > max {
				max = n
			}
		}
		break
	}
	if max > 90 {
		max = 90
	}
	return max
}

// deviceHasHROrRRInSnapshot 该设备在 snapshot 的任一条 track 上存在 heart_rate 或 respiratory_rate 且 >0。
// deviceKey 为 device_id（业务主 key）。
func deviceHasHROrRRInSnapshot(snap *CardSnapshot, deviceKey string) bool {
	if snap == nil {
		return false
	}
	for _, d := range snap.Devices {
		if d.DeviceID != deviceKey && d.DeviceUID != deviceKey {
			continue
		}
		for _, fields := range d.Tracks {
			if intFromAny(fields[observation.FieldHeartRate]) > 0 || intFromAny(fields[observation.FieldRespiratoryRate]) > 0 {
				return true
			}
		}
		return false
	}
	return false
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

// DeriveDeviceOnlineOnly Phase A：仅写每台设备的 device:status:{deviceID} Hash。
// derive 定时点：buffer 中有活跃 device 但无 field-level snapshot 时也要刷新心跳，避免被健康检测误判 offline。
func (s *StateService) DeriveDeviceOnlineOnly(ctx context.Context, cardID string, meta *CardMeta, buf *MonitorBuffer) error {
	var devices map[string]*DeviceMeta
	if meta != nil {
		devices = meta.Devices
	}
	fromTracker := buf.BuildDeviceStatus(cardID, devices, time.Now().UnixMilli())
	if len(fromTracker) == 0 {
		return nil
	}
	for _, ds := range fromTracker {
		if err := s.writer.WriteDeviceStatus(ctx, ds); err != nil {
			s.logger.Warn("write device status",
				zap.String("device_id", ds.DeviceID),
				zap.Error(err))
		}
	}
	return nil
}

// SetDeviceOnline Phase A：写 device:status:{deviceID} 独立 Hash。cardID 仍接受但仅用于日志。
func (s *StateService) SetDeviceOnline(ctx context.Context, cardID, deviceID, deviceUID, deviceType string, online bool) error {
	_ = cardID
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

// SetCardOffline Phase A：把该卡所有设备的 device:status:{deviceID} 标 offline=1。
func (s *StateService) SetCardOffline(ctx context.Context, cardID string, meta *CardMeta) {
	if meta == nil {
		return
	}
	for devID, dm := range meta.Devices {
		if devID == "" {
			continue
		}
		if err := s.writer.SetDeviceOnline(ctx, devID, dm.DeviceUID, dm.DeviceType, false); err != nil {
			s.logger.Warn("set offline",
				zap.String("cid", cardID),
				zap.String("device_id", devID),
				zap.Error(err))
		}
	}
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
