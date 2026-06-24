package roomengine

// sensor_fusion.go
//
// 多传感器融合的扩展点：roomengine 接收非雷达传感器观测（Sleepad / 未来 Geophone 等），
// 与雷达 track 数据交叉验证，提升异常判定准确度。
//
// 当前已实现：
//   - SleepadObservation：床压传感器（在床/离床 + 呼吸/心率）
//   - 用途：silent fall 触发时，若同房间 sleepad 确认"在床有生命体征"，
//     说明雷达消失只是失锁，不报警
//
// 未来扩展模板：
//   1. 定义 XxxObservation struct（包含 DeviceAddr + TMs + 关键状态字段）
//   2. TrackManager 加 xxxStates map[deviceAddr]*XxxObservation
//   3. 加 ProcessXxxObservation(obs) 方法（per-device 保留最新）
//   4. Engine.handleMessage 按 device_type 路由
//   5. 异常判定（silent fall / long still / 等）查询相关传感器状态做 short-circuit

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"

	"owl-common/alarm"
)

// SleepadObservation 一帧 sleepad 观测（与 radar TrackFrame 平行）
// 字段对应 iot_timeseries.data_value 中 sleepad 的 JSON
//
// 字段语义说明（实测数据 2026-04-25）：
//   - bed_status: **0 = 在床（未脱离），1 = 离床（detached）**
//     注意极性反直觉！通过 monitor 帧 bed_status==0 + heart_rate>0 验证。
//   - event_name=InBed → bed_status=0；event_name=LeftBed → bed_status=1
//   - 故 InBed = (bed_status == 0)
type SleepadObservation struct {
	DeviceAddr      string
	DeviceUID       string
	TMs             int64
	TrackID         int  // sleepad 板号(0/1)；2 板归 1 人，透传进 lid 作信息（[[validate_real_case_no_unit_tests]] 单人床）
	InBed           bool // bed_status == 0
	HeartRate       int  // 0 = no signal
	RespiratoryRate int  // 0 = no signal
	BodyMove        int  // 0/1
	TurnOver        int  // 0/1
}

// HasVitalSign 任一生命体征非零（可信存活信号）
func (o *SleepadObservation) HasVitalSign() bool {
	return o.HeartRate > 0 || o.RespiratoryRate > 0
}

// SleepadStatus 一台在场 sleepad 的占用身份快照（forensic + 吸纳用；与 radar TrackStatusBase 平行）。
type SleepadStatus struct {
	LogicID    string     // uid_last4 + "S" + bedHex2 + track_id
	DeviceUID  string     // hex MAC（uid_last4 源）
	DeviceAddr netip.Addr // 设备 /128（mm.SamebedConf 查 radar↔sleepad 同床 prior）
	InBed      bool
	Fresh      bool // 接触 vital fresh(InBed + HR/RR within TTL)；forensic
	Stale      bool // 观测数据陈旧(nowMs-TMs ≥ TTL)：垫哑了 → 解吸纳防幽灵不复活(§3.3 V6)
}

// RadarBedState 一台 radar 设备本帧"是否在床"（N = MN/FwAreaID 命中床区，§9.1 raw 固件权威，**非 belief 后验**：
// 防 belief 受 sleepad vital boost 又反喂吸纳=软反馈环）。LogicID = 该在床 radar track 的 lid（pad_absorbed_by 归属）。
type RadarBedState struct {
	Addr    netip.Addr
	LogicID string
	InBed   bool
}

// PadAbsorption 一张垫的吸纳裁决（forensic + uncovered 计数）。
type PadAbsorption struct {
	LogicID      string // pad lid
	InBed        bool
	Fresh        bool
	Stale        bool
	AbsorbedBy   string  // 吸纳它的 radar lid；""=未吸纳
	Samebed      float32 // 与吸纳 radar 的 samebed prior（未吸纳=与最强候选 radar 的值，forensic）
	Uncovered    bool    // 计入 P1 占用 +1
	RadarLeftBed bool    // 有主 radar(samebed≥阈) 但它本帧不在床 = de-absorption 场景(§3.3 V5/V6)，caller LOG
}

// SamebedAbsorbThresh 吸纳判"radar 与该垫同床"的 samebed prior 阈（form-anchor）。
//
//	> 0.5：单床 samebed=1≥阈→立即吸纳；多床候选 0.5<阈→不吸纳→uncovered（§9.2④ FN-safe，靠切片3 learned 收敛）。
//	取 0.8（同 unit_matrix.go bedResolveConf：单床确定吸、几何歧义不吸）。
const SamebedAbsorbThresh float32 = 0.8

// fwAreaIsBed MN：FwAreaID 命中活体 declare_area 床区（type∈{2,5}）。0/255=区外。
func fwAreaIsBed(areaID int, bedAreaIDs []int) bool {
	if areaID == 0 || areaID == 255 {
		return false
	}
	for _, id := range bedAreaIDs {
		if id != 0 && id != 255 && areaID == id {
			return true
		}
	}
	return false
}

// RadarBedStates 从 per-radar bases 算每台 radar 设备本帧"是否在床"（N=MN/FwAreaID，§9.1 raw 固件权威，
// **非 belief 后验** ArgmaxIsBed——防 belief 受 sleepad vital boost 又反喂吸纳=软反馈环）。
// per-device 去重：任一在场 track 命中床区 → 该设备 InBed（带该 track lid 作吸纳归属 pad_absorbed_by）。
func RadarBedStates(bases []TrackStatusBase, bedAreaIDs []int) []RadarBedState {
	idx := map[netip.Addr]int{}
	out := make([]RadarBedState, 0, len(bases))
	for _, b := range bases {
		addr, err := netip.ParseAddr(b.DeviceAddr)
		if err != nil {
			continue
		}
		inBed := b.Present && fwAreaIsBed(b.FwAreaID, bedAreaIDs)
		if i, ok := idx[addr]; ok {
			if inBed && !out[i].InBed {
				out[i].InBed, out[i].LogicID = true, b.LogicID
			}
			continue
		}
		rs := RadarBedState{Addr: addr, InBed: inBed}
		if inBed {
			rs.LogicID = b.LogicID
		}
		idx[addr] = len(out)
		out = append(out, rs)
	}
	return out
}

// AbsorbSleepads 吸纳裁决（切片2 = prior-only，§9.2；30s learned overlay 延切片3）。
//
//	每张 InBed 垫：找在场 radar `samebed_prior(r,s) ≥ thresh` ∧ 该 r `N-in-bed`(MN/FwAreaID)→ 吸纳（pad_absorbed_by=r.lid，不计 uncovered）。
//	否则 uncovered +1（撑 P1 占用，解缺陷②，§8.5）。仅进 P1（占用/alone），**绝不进 fall 风险/C_FN**（§9.4）。
//	de-absorption（§3.3，FN 风险最高 §9.4）：主 radar 离床 → r.InBed=false → 不满足 → 垫落 uncovered；
//	  此时靠 Stale 仲裁——!Stale(接触数据新鲜)=真人还在(保人 +1)；Stale(垫哑滞后)=不复活幽灵(不计)。
//	  ⚠️ V5/V6 真值仲裁须真 case（radar 离床+垫仍 InBed）验，09e7 缺 → caller 据 RadarLeftBed LOG 未覆盖（no-silent-caps）。
//	out 含全部 pad（非 InBed 也带，供 forensic 可见）；uncovered 只数 InBed∧未吸纳∧!Stale。
func AbsorbSleepads(pads []SleepadStatus, radars []RadarBedState, mm *RoomMM, thresh float32) (uncovered int, out []PadAbsorption) {
	out = make([]PadAbsorption, 0, len(pads))
	for _, p := range pads {
		pa := PadAbsorption{LogicID: p.LogicID, InBed: p.InBed, Fresh: p.Fresh, Stale: p.Stale}
		if !p.InBed {
			out = append(out, pa) // 离床的垫不撑占用，仅 forensic
			continue
		}
		var bestSamebed float32
		relatedRadar := false // ∃ radar samebed≥阈（不论在不在床）= 该垫有"主"radar
		for _, r := range radars {
			sb := mm.SamebedConf(r.Addr, p.DeviceAddr)
			if sb > bestSamebed {
				bestSamebed = sb
			}
			if sb >= thresh {
				relatedRadar = true
				if r.InBed {
					pa.AbsorbedBy, pa.Samebed = r.LogicID, sb
					break
				}
			}
		}
		if pa.AbsorbedBy == "" {
			pa.Samebed = bestSamebed
			pa.RadarLeftBed = relatedRadar // 有主 radar 但它不在床 = de-absorption(§3.3 V5/V6)
			if !p.Stale {                  // 接触数据新鲜=真人还在(保人)；陈旧=不复活幽灵(§3.3 V6)
				pa.Uncovered = true
				uncovered++
			}
		}
		out = append(out, pa)
	}
	return uncovered, out
}

// SleepadLogicID sleepad 占用身份 lid = uid_last4 + "S" + bedHex2 + track_id。
//
//	uid_last4 取自终身稳定的 UID（DevKey，[[event_log_monitor_stream_device_uid]]）；bedHex2 取自会变的 addr=跟踪当前绑床；
//	全本地派生零 data 取数。详 doc/fusion-absorption-B-sleepad.md §2。
func SleepadLogicID(uidHex, deviceAddr string, trackID int) string {
	last4 := uidHex
	if len(uidHex) > 4 {
		last4 = uidHex[len(uidHex)-4:]
	}
	return fmt.Sprintf("%sS%s%d", last4, bedSlotHex(deviceAddr), trackID)
}

// bedSlotHex device /128 addr 掩到 /96 的床槽 byte[11]（DeriveBedPrefix slot 字节，room/88→bed/96），渲 2 位 hex。
//
//	deviceAddr 是 firmware 上报的边界值；解析失败→"00"（forensic 路径不崩，异常自显形）。
func bedSlotHex(deviceAddr string) string {
	a, err := netip.ParseAddr(deviceAddr)
	if err != nil {
		return "00"
	}
	b := a.As16()
	return fmt.Sprintf("%02x", b[11])
}

// ParseSleepadObservations 把 sleepad iot:monitor:stream 的 data_value 解析为多帧观测
// data_value 是 array of object，每个 object 一条记录（同一设备一帧通常只 1 条 track_id=0）
func ParseSleepadObservations(dv interface{}, deviceAddr, deviceUID string, fallbackTs int64) []SleepadObservation {
	if dv == nil {
		return nil
	}
	// 兼容 data_value 直接是 array 或 string-encoded JSON
	var arr []map[string]interface{}
	switch v := dv.(type) {
	case []interface{}:
		for _, it := range v {
			if m, ok := it.(map[string]interface{}); ok {
				arr = append(arr, m)
			}
		}
	case []map[string]interface{}:
		arr = v
	case string:
		_ = json.Unmarshal([]byte(v), &arr)
	}
	if len(arr) == 0 {
		return nil
	}

	out := make([]SleepadObservation, 0, len(arr))
	for _, m := range arr {
		obs := SleepadObservation{
			DeviceAddr: deviceAddr,
			DeviceUID:  deviceUID,
			TMs:        fallbackTs,
		}
		if v, ok := m["track_id"]; ok {
			obs.TrackID = jsonInt(v)
		}
		if v, ok := m["bed_status"]; ok {
			// 注意：0 = 在床（未 detached），1 = 离床
			if jsonInt(v) == 0 {
				obs.InBed = true
			}
		}
		if v, ok := m["heart_rate"]; ok {
			obs.HeartRate = jsonInt(v)
		}
		if v, ok := m["respiratory_rate"]; ok {
			obs.RespiratoryRate = jsonInt(v)
		}
		if v, ok := m["body_move"]; ok {
			obs.BodyMove = jsonInt(v)
		}
		if v, ok := m["turn_over"]; ok {
			obs.TurnOver = jsonInt(v)
		}
		out = append(out, obs)
	}
	return out
}

// SleepadBedEvent sleepad 床压事件（来自 iot:event:stream），用于人数计数。
// 实测发现一次状态变化会发 2 条事件（status=instant + status=start），需要用 status 去重。
type SleepadBedEvent struct {
	DeviceUID string
	TMs       int64
	IsInBed   bool   // event_name="InBed" → true；"LeftBed" → false
	Status    string // "instant" / "start"（去重用，只取 instant）
}

// ParseSleepadBedEvents 把 sleepad event 流的 data_value 解析为床压事件。
// 仅返回 envelopeCat ∈ {InBed, LeftBed} 且 status=instant 的事件（避免重复计数）。
// envelopeCat 来自 IoTStreamMessage.Category（事件类型唯一权威）。
func ParseSleepadBedEvents(dv interface{}, deviceUID, envelopeCat string, fallbackTs int64) []SleepadBedEvent {
	if dv == nil {
		return nil
	}
	var isInBed bool
	switch envelopeCat {
	case alarm.InBed:
		isInBed = true
	case alarm.LeftBed:
		isInBed = false
	default:
		return nil
	}
	var arr []map[string]interface{}
	switch v := dv.(type) {
	case []interface{}:
		for _, it := range v {
			if m, ok := it.(map[string]interface{}); ok {
				arr = append(arr, m)
			}
		}
	case []map[string]interface{}:
		arr = v
	case string:
		_ = json.Unmarshal([]byte(v), &arr)
	}
	if len(arr) == 0 {
		return nil
	}
	out := make([]SleepadBedEvent, 0, len(arr))
	for _, m := range arr {
		st, _ := m["event_status"].(string)
		if st != "instant" {
			continue // 只用 instant 事件，避免 instant+start 重复
		}
		out = append(out, SleepadBedEvent{DeviceUID: deviceUID, TMs: fallbackTs, IsInBed: isInBed, Status: st})
	}
	return out
}

// jsonInt 把 JSON 数字（float64 / json.Number / int / string）安全转 int
func jsonInt(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}
