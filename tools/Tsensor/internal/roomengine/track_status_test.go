// track_status_test.go — TrackStatus 序列化 + PersonID 写入规则单测
//
// sensor_v2 PR-3 验收：
//   - VerdictName 4 档字符串契约稳定
//   - ToStreamMap 字段名 / 零值省略 / wire schema
//   - 候选阶段 (upgraded=false) 不写 PersonID（review 抓的语义漏洞）

package roomengine

import (
	"strings"
	"testing"
)

func TestVerdictNameStable(t *testing.T) {
	cases := map[TrackVerdict]string{
		VerdictPending:  "pending",
		VerdictReal:     "real",
		VerdictGhost:    "ghost",
		VerdictAnchored: "anchored",
	}
	for v, want := range cases {
		if got := VerdictName(v); got != want {
			t.Fatalf("VerdictName(%d) = %q, want %q", v, got, want)
		}
	}
	// 未来新增 verdict 默认 fallback 到 pending（防 wire 漏报字符串）
	if got := VerdictName(TrackVerdict(99)); got != "pending" {
		t.Fatalf("unknown verdict fallback should be pending, got %q", got)
	}
}

func TestTrackStatusToStreamMap_RequiredFields(t *testing.T) {
	ts := &TrackStatus{
		TrackID:      7,
		DeviceAddr:   "fd00:0:3:111:80::cafe",
		RoomID:       "fd00:0:3:111:80::/128",
		Verdict:      VerdictReal,
		GhostPenalty: 12,
		X:            120, Y: 240, Z: 90,
		Pose:         4,
		StillSec:     33,
		CellAreaType: AreaBed,
		EnterTarget:  "",
		UpdatedAtMs:  1700000000123,
	}
	m := ts.ToStreamMap()
	// 必填字段：verdict / track_id / device_addr / room_id / x / y / z / pose / still_sec / cell_area_type / updated_at_ms
	requireKey(t, m, "verdict", "real")
	requireKey(t, m, "track_id", 7)
	requireKey(t, m, "device_addr", "fd00:0:3:111:80::cafe")
	requireKey(t, m, "room_id", "fd00:0:3:111:80::/128")
	requireKey(t, m, "cell_area_type", "bed")
	requireKey(t, m, "ghost_penalty", 12)
	requireKey(t, m, "still_sec", 33)
	// 零值 ZoneID / PersonID 必须省略（防 wire payload 膨胀）
	if _, ok := m["in_bed_zone_id"]; ok {
		t.Errorf("empty in_bed_zone_id should be omitted")
	}
	if _, ok := m["person_id"]; ok {
		t.Errorf("empty person_id should be omitted")
	}
}

func TestTrackStatusToStreamMap_OptionalFields(t *testing.T) {
	ts := &TrackStatus{
		Verdict:      VerdictAnchored,
		InBedZoneID:  "room_a",
		InRoomZoneID: "room_a",
		PersonID:     "resident_uuid",
		PersonRole:   "resident",
		CellAreaType: AreaBed,
	}
	m := ts.ToStreamMap()
	requireKey(t, m, "verdict", "anchored")
	requireKey(t, m, "in_bed_zone_id", "room_a")
	requireKey(t, m, "in_room_zone_id", "room_a")
	requireKey(t, m, "person_id", "resident_uuid")
	requireKey(t, m, "person_role", "resident")
}

// 验证 PR-3 review 抓的语义漏洞：candidate 阶段（upgraded=false）不应写 PersonID。
// 直接用 SuiteCensusManager 模拟 — 第一次喂帧返回 upgraded=false（candidate），
// 此时下游 caller 必须 NOT 把临时 candidate key 写进 TrackStatus.PersonID。
func TestPersonIDNotPopulatedWhileCandidate(t *testing.T) {
	mgr := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	const suiteID = "fd00:0:3:111::/80"
	const resID = "resident-uuid"
	const trackID = 9

	// 第 1 次喂帧：未到升格阈值 → upgraded=false
	person, upgraded := mgr.UpdatePersonFromTrack(suiteID, resID, trackID, false /*sleepad*/, 1, true, 1_000)
	if upgraded || person != nil {
		t.Fatalf("first frame should be candidate (upgraded=false), got upgraded=%v person=%+v", upgraded, person)
	}

	// 模拟 publishTrackStatuses 的写入规则
	status := &TrackStatus{TrackID: trackID}
	if upgraded && person != nil { // 故意保留原始判定结构，确保不会进 if
		status.PersonID = person.PersonID
		status.PersonRole = person.Role
	}
	if status.PersonID != "" {
		t.Fatalf("candidate 阶段 PersonID 必须保持空，实际 %q", status.PersonID)
	}

	// 验证 candidate key 不会 leak 到 wire schema
	m := status.ToStreamMap()
	if v, ok := m["person_id"]; ok {
		t.Fatalf("candidate 阶段 person_id 不应进入 wire payload, got %v", v)
	}
}

// 验证 sleepad 强升格命中后 PersonID 立即可用（与 candidate 路径对称）。
func TestPersonIDPopulatedWhenUpgraded(t *testing.T) {
	mgr := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	const suiteID = "fd00:0:3:111::/80"
	const resID = "resident-uuid"
	const trackID = 9

	// sleepadInBed=true → 立即 resident 强升格
	person, upgraded := mgr.UpdatePersonFromTrack(suiteID, resID, trackID, true, 0, true, 1_000)
	if !upgraded || person == nil {
		t.Fatalf("sleepad 双源锚定应立即升格 (upgraded=true), got upgraded=%v person=%+v", upgraded, person)
	}

	status := &TrackStatus{TrackID: trackID}
	if upgraded && person != nil {
		status.PersonID = person.PersonID
		status.PersonRole = person.Role
	}
	if status.PersonID != resID {
		t.Fatalf("upgraded resident PersonID 应为 %q，实际 %q", resID, status.PersonID)
	}
	if status.PersonRole != SuitePersonResident {
		t.Fatalf("PersonRole 应为 %q，实际 %q", SuitePersonResident, status.PersonRole)
	}
	if !strings.HasPrefix(status.PersonID, "resident-") {
		t.Fatalf("PersonID 不应该是 __cand_track_ 临时 key: %q", status.PersonID)
	}
}

// 验证 PR-3 review 补的失锁清理：上一帧出现过 trackID=N，本帧没出现 ≥ 60s →
// 必须调 ClearAnchorOnLostTrack，防 firmware track_id 复用时 SuiteCensus 把新 track
// 误认作"同一人活动"持续 update LastActiveMs。
func TestSuiteCensusClearAnchorOnLostTrack(t *testing.T) {
	mgr := NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
	const suiteID = "fd00:0:3:111::/80"
	const resID = "resident-uuid"
	const trackID = 42

	// sleepad 强升格 resident 锚到 trackID=42
	person, upgraded := mgr.UpdatePersonFromTrack(suiteID, resID, trackID, true, 0, true, 1_000)
	if !upgraded || person == nil || person.AnchorTrackID != trackID {
		t.Fatalf("setup: expected resident anchored on trackID=%d, got %+v upgraded=%v", trackID, person, upgraded)
	}

	// 模拟 publishTrackStatuses 失锁清理：track 60s 没出现 → 调 ClearAnchorOnLostTrack
	cleared := mgr.ClearAnchorOnLostTrack(suiteID, trackID)
	if !cleared {
		t.Fatalf("ClearAnchorOnLostTrack should return true on existing anchor")
	}

	// AnchorTrackID 应被清空（但 Person 本身不删，等 DecayInactive 处理）
	c := mgr.Get(suiteID)
	if c == nil {
		t.Fatalf("census should still exist after clear")
	}
	p := c.Persons[resID]
	if p == nil {
		t.Fatalf("resident should still exist after clear")
	}
	if p.AnchorTrackID != 0 {
		t.Fatalf("AnchorTrackID should be 0 after clear, got %d", p.AnchorTrackID)
	}
}

func requireKey(t *testing.T, m map[string]interface{}, key string, want interface{}) {
	t.Helper()
	got, ok := m[key]
	if !ok {
		t.Fatalf("missing key %q in stream map", key)
	}
	if got != want {
		t.Fatalf("key %q = %v (%T), want %v (%T)", key, got, got, want, want)
	}
}
