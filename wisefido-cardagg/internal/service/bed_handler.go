package service

import (
	"context"
	"strings"
	"sync"
	"time"
)

// 离床跌倒阈值见 weights.go：LeftBedFallMaxActivityCount=5、LeftBedFallSumStandThreshold=260sec、LeftBedFallSumMoveUpdateSec=10sec、LeftBedFallDanceUpdateMin=1metre

// pendingLeftBedFall 离床后前 5 个 Activity 的累计；deviceID 供最终上报用。
type pendingLeftBedFall struct {
	mu           sync.Mutex
	cardID       string
	deviceID     string
	count        int
	sumStand     int
	sumMove      int
	sumDance     int
	exitTrackGT1 bool
}

var leftBedFallPending sync.Map // cardID -> *pendingLeftBedFall

// CardHasRadar 该卡是否绑有雷达（仅雷达能提供 Activity 与 track_count）。供调用方判断是否调用 StartLeftBedFall。
func CardHasRadar(meta *CardMeta) bool {
	if meta == nil || meta.Devices == nil {
		return false
	}
	for _, dm := range meta.Devices {
		if dm == nil {
			continue
		}
		t := strings.ToLower(dm.DeviceType)
		if strings.Contains(t, "radar") {
			return true
		}
	}
	return false
}

// RadarDeviceIDBoundToBed 返回绑定在该卡床位上的第一个雷达的 device_id（BoundBedID == meta.BedID 且 DeviceType 含 radar）。
func RadarDeviceIDBoundToBed(meta *CardMeta) string {
	if meta == nil || meta.BedID == "" || meta.Devices == nil {
		return ""
	}
	for id, dm := range meta.Devices {
		if dm == nil || dm.BoundBedID != meta.BedID {
			continue
		}
		if strings.Contains(strings.ToLower(dm.DeviceType), "radar") {
			return id
		}
	}
	return ""
}

const bedEventLeftBed = 1

// StartLeftBedFall 离床时由调用方在确认有 radar 后调用，仅创建 pending（cardID + deviceID）。不读 meta、不判断 radar。
func StartLeftBedFall(cardID, deviceID string) {
	if cardID == "" || deviceID == "" {
		return
	}
	leftBedFallPending.Store(cardID, &pendingLeftBedFall{cardID: cardID, deviceID: deviceID})
}

// NeedBedFallCheck 是否有该卡的离床跌倒检测 pending，供调用方在 Activity 时先查再调 LeftBedFallActivity，避免每条 Activity 都走完整逻辑。
func NeedBedFallCheck(cardID string) bool {
	_, ok := leftBedFallPending.Load(cardID)
	return ok
}

// LeftBedFallActivity 收到该卡一次 Activity 时调用。累加；sum_move/sum_dance 达阈值则更新 target lastActive；满 5 次做最终判定并删 pending，返回 (done, suspectedFall, reportDeviceID)。是否落库由调用方根据返回值处理。
func LeftBedFallActivity(
	ctx context.Context,
	cardID, deviceID string,
	standSec, moveSec, danceSec, trackNum int,
	state *StateService,
) (done, suspectedFall bool, reportDeviceID string) {
	v, ok := leftBedFallPending.Load(cardID)
	if !ok || v == nil {
		return false, false, ""
	}
	p := v.(*pendingLeftBedFall)
	p.mu.Lock()
	if p.count >= LeftBedFallMaxActivityCount {
		p.mu.Unlock()
		return false, false, ""
	}
	p.count++
	p.sumStand += standSec
	p.sumMove += moveSec
	p.sumDance += danceSec
	if trackNum > 1 {
		p.exitTrackGT1 = true
	}
	if state != nil && (p.sumMove >= LeftBedFallSumMoveUpdateSec || p.sumMove >= 1) {
		danceMin := p.sumDance / 60
		_ = state.UpdateTargetLastActive(context.Background(), cardID, time.Now().UnixMilli(), p.sumMove, LeftBedFallSumMoveUpdateSec, danceMin, LeftBedFallDanceUpdateMin)
	}
	count := p.count
	sumStand := p.sumStand
	exitTrackGT1 := p.exitTrackGT1
	reportDeviceID = p.deviceID
	p.mu.Unlock()

	if count < LeftBedFallMaxActivityCount {
		return false, false, ""
	}

	leftBedFallPending.Delete(cardID)
	if exitTrackGT1 {
		return true, false, ""
	}
	if sumStand < LeftBedFallSumStandThreshold {
		return true, false, ""
	}

	status, err := state.ReadCardStatus(ctx, cardID)
	if err != nil || status == nil || status.BedState == nil {
		return true, false, ""
	}
	bs := status.BedState
	if bs.BedStatus != 1 || bs.BedEvent != bedEventLeftBed {
		return true, false, ""
	}
	if reportDeviceID == "" {
		return true, false, ""
	}
	return true, true, reportDeviceID
}
