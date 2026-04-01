package service

import (
	"context"
	"strings"
	"sync"
	"time"
)

// 离床跌倒阈值见 weights.go：LeftBedFallMaxActivityCount=5、LeftBedFallSumStandThreshold=260sec、LeftBedFallSumMoveUpdateSec=10sec、LeftBedFallDanceUpdateMin=1metre
//
// 本文件「pending」= 进程内存中的离床跌倒检测窗口（leftBedFallPending），与告警模块里 Redis 的 Stay 等 pending 告警不是同一概念；进程重启会丢失未判完的窗口。

// pendingLeftBedFall 离床后至多 LeftBedFallMaxActivityCount 条 Activity 的滚动累计；满次数后终判并删除 map 项。deviceID 为 StartLeftBedFall 写入的上报设备。
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

// leftBedFallPending：有 entry 表示该 card 处于「已离床、正用接下来若干条 Activity 做跌倒累计」；无 entry 则 NeedBedFallCheck 为 false，LeftBedFallActivity 直接返回。
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

// StartLeftBedFall 离床时由调用方在确认有 radar 后调用：向 leftBedFallPending 放入新窗口；若已存在同 card 会覆盖。不读 meta、不判 radar。
func StartLeftBedFall(cardID, deviceID string) {
	if cardID == "" || deviceID == "" {
		return
	}
	leftBedFallPending.Store(cardID, &pendingLeftBedFall{cardID: cardID, deviceID: deviceID})
}

// NeedBedFallCheck 该卡是否在 leftBedFallPending 中；true 时 routeActivityEvent 才调用 LeftBedFallActivity，避免无窗口时每条 Activity 都进完整逻辑。
func NeedBedFallCheck(cardID string) bool {
	_, ok := leftBedFallPending.Load(cardID)
	return ok
}

// LeftBedFallActivity 离床跌倒：NeedBedFallCheck 为 true（已 StartLeftBedFall）时，每条雷达 Activity 调一次。
//
// 业务背景：床侧有 radar（提供 activity / track_count）时，离床后可 StartLeftBedFall；窗口内累计离床后前 LeftBedFallMaxActivityCount（5）条 Activity，不是「固定 5 分钟」，实际时长取决于上报间隔。
//
// 流程摘要：
//   - 无 pending 或本批已计满次数：返回 (false,false,"")。
//   - 每条：count+1，累加 stand/move/dance 秒；任一次 trackNum>1 → exitTrackGT1（多轨迹视为非单人跌倒场景，终判不报疑似）。
//   - sum_move 达 LeftBedFallSumMoveUpdateSec 等时仅 UpdateTargetLastActive（活动可见性），不参与「排除跌倒」终判。
//   - 未满 5 条：返回 (false,false,"")。
//   - 满 5 条：删 pending 终判——exitTrackGT1 → 不疑似；sum_stand < LeftBedFallSumStandThreshold（260s）→ 不疑似（站立累计不足，不按跌倒报）；
//     再读 BedState：须仍为离床(bed_status==1 且 bed_event==离床)，否则不疑似；否则 (true,true,deviceID) 由调用方落库 SuspectedFall。
//   - 与口语「移动多=有活动」不同：实现上终判主要看 5 条内累计站立秒 sum_stand 是否 ≥260（长时间少动），见 weights.go。

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
