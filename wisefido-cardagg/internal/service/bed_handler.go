package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
)

// 离床跌倒阈值见 weights.go：LeftBedFallMaxActivityCount=5、LeftBedFallSumStandThreshold=260sec、LeftBedFallSumMoveUpdateSec=10sec、LeftBedFallDanceUpdateMin=1（分钟）
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

// ImmediateLeftBedFallCheck 离床瞬间读 monitor buffer：床绑雷达 pose==PoseLying 时 (true, 床绑雷达 device_id)。
func ImmediateLeftBedFallCheck(cardID string, buf *MonitorBuffer, meta *CardMeta) (suspected bool, bedRadarID string) {
	return leftBedFallPoseLyingAndRadar(cardID, buf, meta)
}

// PersistSuspectedFallPoseLyingIfEnabled 快速轨：lying 时按使能落库 SuspectedFall（source 区分 event/alarm/coord/pending）。
func PersistSuspectedFallPoseLyingIfEnabled(ctx context.Context, al *AlarmService, cardID, tenantID string, buf *MonitorBuffer, meta *CardMeta, source string) {
	if al == nil || tenantID == "" || cardID == "" {
		return
	}
	ok, radarID := ImmediateLeftBedFallCheck(cardID, buf, meta)
	if !ok || radarID == "" {
		return
	}
	_, level, _, _, _, enabled := al.ResolveEnablementByDevice(ctx, tenantID, radarID, alarm.SuspectedFall)
	if !enabled || level == "" {
		level = alarm.AlarmLevelWarn
	}
	nowMs := time.Now().UnixMilli()
	_ = al.PersistAlarmWithTriggerData(ctx, cardID, tenantID, radarID, alarm.SuspectedFall, level, time.UnixMilli(nowMs), map[string]interface{}{
		"source": source,
		"ts":     nowMs,
	})
}

// LeftBedFallFromPoseLying monitor 流床绑雷达任一条 track 的 pose 为卧床（与 observation 枚举一致，协议 byte13=6）。用于离床跌倒 C&I 快速判据。
func LeftBedFallFromPoseLying(cardID string, buf *MonitorBuffer, meta *CardMeta) bool {
	ok, _ := ImmediateLeftBedFallCheck(cardID, buf, meta)
	return ok
}

func leftBedFallPoseLyingAndRadar(cardID string, buf *MonitorBuffer, meta *CardMeta) (ok bool, bedRadarID string) {
	if buf == nil || meta == nil || cardID == "" {
		return false, ""
	}
	radarID := RadarDeviceIDBoundToBed(meta)
	if radarID == "" {
		return false, ""
	}
	snap := buf.SnapshotCard(cardID)
	if snap == nil {
		return false, radarID
	}
	for _, d := range snap.Devices {
		if d.DeviceID != radarID && d.DeviceUID != radarID {
			continue
		}
		for _, fields := range d.Tracks {
			raw, present := fields[observation.FieldPose]
			if !present {
				continue
			}
			if leftBedFallPoseInt(raw) == observation.PoseLying {
				return true, radarID
			}
		}
		return false, radarID
	}
	return false, radarID
}

func leftBedFallPoseInt(v any) int {
	if v == nil {
		return -1
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return -1
		}
		return int(n)
	default:
		return -1
	}
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
//   - 无 pending 或本批已计满次数：返回 (false,false,"","")。
//   - 每条：count+1，累加 stand/move/dance 秒；任一次 trackNum>1 → exitTrackGT1（多轨迹视为非单人跌倒场景，终判不报疑似）。
//   - sum_move 达 LeftBedFallSumMoveUpdateSec 等时仅 UpdateTargetLastActive（活动可见性），不参与「排除跌倒」终判。
//   - 未满 5 条：返回 (false,false,"","")。
//   - 满 5 条：删 pending 终判——exitTrackGT1 → 不疑似；再读 BedState 须仍为离床。
//     满足其一即疑似：① 床绑雷达 buffer 中 pose==PoseLying（C&I：离床后仍判卧床姿态）；② sum_stand≥260s（久站兜底，高度不敏感时躺被报站）。
//     否则不疑似。(true,true,deviceID,path) 由调用方落库 SuspectedFall；path 为 "pose_lying"|"sum_stand"|""。

func LeftBedFallActivity(
	ctx context.Context,
	cardID, deviceID string,
	standSec, moveSec, danceSec, trackNum int,
	state *StateService,
	buf *MonitorBuffer,
	meta *CardMeta,
) (done, suspectedFall bool, reportDeviceID string, path string) {
	v, ok := leftBedFallPending.Load(cardID)
	if !ok || v == nil {
		return false, false, "", ""
	}
	p := v.(*pendingLeftBedFall)
	p.mu.Lock()
	if p.count >= LeftBedFallMaxActivityCount {
		p.mu.Unlock()
		return false, false, "", ""
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
		return false, false, "", ""
	}

	leftBedFallPending.Delete(cardID)
	if exitTrackGT1 {
		return true, false, "", ""
	}

	status, err := state.ReadCardStatus(ctx, cardID)
	if err != nil || status == nil || status.BedState == nil {
		return true, false, "", ""
	}
	bs := status.BedState
	if bs.BedStatus != 1 || bs.BedEvent != bedEventLeftBed {
		return true, false, "", ""
	}
	if reportDeviceID == "" {
		return true, false, "", ""
	}

	fromLying, bedRadar := leftBedFallPoseLyingAndRadar(cardID, buf, meta)
	fromStand := sumStand >= LeftBedFallSumStandThreshold
	if !fromLying && !fromStand {
		return true, false, "", ""
	}
	outID := reportDeviceID
	if fromLying && bedRadar != "" {
		outID = bedRadar
	}
	if fromLying {
		return true, true, outID, "pose_lying"
	}
	return true, true, outID, "sum_stand"
}
