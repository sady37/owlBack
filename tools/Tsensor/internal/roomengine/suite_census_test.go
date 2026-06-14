package roomengine

import (
	"testing"

	"owl-common/card"
)

const (
	testSuiteID   = "fd00:0:3:111::/80"
	testResident  = "11111111-1111-1111-1111-111111111111"
	testRadarTrk1 = 1
	testRadarTrk2 = 2
)

func newTestManager() *SuiteCensusManager {
	// nil redisClient + nil logger = 纯内存模式 + no-op log
	return NewSuiteCensusManager(nil, DefaultSuiteCensusConfig(), nil)
}

// TestResidentSleepadStrongAnchor sleepad 双源锚定直接升 resident（不等 5min）
func TestResidentSleepadStrongAnchor(t *testing.T) {
	m := newTestManager()
	nowMs := int64(1_000_000_000)

	p, ok := m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true /*sleepadInBed*/, 0 /*traverseDelta*/, true /*moveActive*/, nowMs)
	if !ok {
		t.Fatal("expected immediate resident upgrade via sleepad anchor")
	}
	if p.Role != SuitePersonResident {
		t.Errorf("role want resident, got %s", p.Role)
	}
	if p.PersonID != testResident {
		t.Errorf("personID want %s, got %s", testResident, p.PersonID)
	}
	if !p.SleepadAnchored {
		t.Error("SleepadAnchored should be true")
	}
}

// TestResidentRadarAnchor 无 sleepad 时，5min + 10 cells 升 resident
func TestResidentRadarAnchor(t *testing.T) {
	m := newTestManager()
	startMs := int64(1_000_000_000)

	// t=0：首帧
	p, ok := m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		false, 1, true, startMs)
	if ok {
		t.Fatal("should not upgrade on first frame")
	}
	if p != nil {
		t.Error("first frame should return nil person")
	}

	// 4min30s + 9 cells → 不够（age 不满 5min）
	midMs := startMs + 4*60*1000 + 30*1000
	p, ok = m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		false, 8, true, midMs)
	if ok {
		t.Errorf("should not upgrade at 4min30s (age=%dms < 5min)", midMs-startMs)
	}

	// 5min + 1s + 11 cells → 升格
	upMs := startMs + 5*60*1000 + 1000
	p, ok = m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		false, 2, true, upMs)
	if !ok {
		t.Fatalf("should upgrade at 5min+1s with traverse=%d", p.TraverseCount)
	}
	if p.Role != SuitePersonResident {
		t.Errorf("role want resident, got %s", p.Role)
	}
	if p.PersonID != testResident {
		t.Errorf("personID want %s, got %s", testResident, p.PersonID)
	}
}

// TestVisitorAnchor 2min + 5 cells 升 visitor（决定 13 时序假设）
func TestVisitorAnchor(t *testing.T) {
	m := newTestManager()
	startMs := int64(2_000_000_000)

	// 首帧：residentID="" 表示未知 person → 走 visitor 路径
	_, ok := m.UpdatePersonFromTrack(testSuiteID, "", testRadarTrk2,
		false, 1, true, startMs)
	if ok {
		t.Fatal("should not upgrade on first frame")
	}

	// 1min30s + 4 cells → 不够
	_, ok = m.UpdatePersonFromTrack(testSuiteID, "", testRadarTrk2,
		false, 3, true, startMs+90*1000)
	if ok {
		t.Error("should not upgrade at 1m30s")
	}

	// 2min + 1s + 6 cells → 升 visitor
	p, ok := m.UpdatePersonFromTrack(testSuiteID, "", testRadarTrk2,
		false, 2, true, startMs+2*60*1000+1000)
	if !ok {
		t.Fatal("should upgrade visitor at 2min+1s + 6 cells")
	}
	if p.Role != SuitePersonVisitor {
		t.Errorf("role want visitor, got %s", p.Role)
	}
	if !IsCandidate(p.PersonID) && p.PersonID[:8] != "visitor_" {
		t.Errorf("visitor PersonID should start with 'visitor_', got %s", p.PersonID)
	}
}

// TestPublicBathroomNoPerson public mode 下不识别 person（决定 24 / §4.A.6）
func TestPublicBathroomNoPerson(t *testing.T) {
	m := newTestManager()
	m.MarkPublicBathroom(testSuiteID)

	p, ok := m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true, 100, true, 3_000_000_000)
	if ok {
		t.Error("public bathroom should not identify person, even with sleepad anchor")
	}
	if p != nil {
		t.Error("public bathroom should return nil person")
	}

	c := m.Get(testSuiteID)
	if !c.IsPublicBathroom {
		t.Error("census should be flagged public bathroom")
	}
	if len(c.Persons) != 0 {
		t.Errorf("public bathroom Persons should be empty, got %d", len(c.Persons))
	}
}

// TestBathroomGateTransition PR-5 BathroomGate 调用 MarkPersonExitToBathroom 后状态切换
func TestBathroomGateTransition(t *testing.T) {
	m := newTestManager()
	nowMs := int64(4_000_000_000)

	// 先在 bedroom 升 resident（sleepad 直接锚定）
	p, ok := m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true, 0, true, nowMs)
	if !ok {
		t.Fatal("setup: resident upgrade failed")
	}
	if p.AnchorRoomType != card.RoomTypeDefault {
		t.Errorf("initial AnchorRoomType want bedroom, got %d", p.AnchorRoomType)
	}

	// resident 经 BathroomGate 进 bathroom
	if !m.MarkPersonExitToBathroom(testSuiteID, testResident, nowMs+1000) {
		t.Fatal("MarkPersonExitToBathroom returned false")
	}
	c := m.Get(testSuiteID)
	if c.BathroomCount != 1 {
		t.Errorf("BathroomCount want 1, got %d", c.BathroomCount)
	}
	if c.Persons[testResident].AnchorRoomType != card.RoomTypeBathroom {
		t.Errorf("AnchorRoomType want bathroom after exit, got %d", c.Persons[testResident].AnchorRoomType)
	}

	// 返回 bedroom
	if !m.MarkPersonReturnToBedroom(testSuiteID, testResident, nowMs+5000) {
		t.Fatal("MarkPersonReturnToBedroom returned false")
	}
	c = m.Get(testSuiteID)
	if c.BathroomCount != 0 {
		t.Errorf("BathroomCount want 0 after return, got %d", c.BathroomCount)
	}
	if c.Persons[testResident].AnchorRoomType != card.RoomTypeDefault {
		t.Errorf("AnchorRoomType want bedroom after return, got %d", c.Persons[testResident].AnchorRoomType)
	}
}

// TestDecayVisitor visitor 离开 ≥ 10min → 移出 census
func TestDecayVisitor(t *testing.T) {
	m := newTestManager()
	startMs := int64(5_000_000_000)

	// 升 visitor
	for delta := int64(0); delta <= 3*60*1000; delta += 30 * 1000 {
		m.UpdatePersonFromTrack(testSuiteID, "", testRadarTrk2,
			false, 1, true, startMs+delta)
	}
	c := m.Get(testSuiteID)
	if len(c.Persons) == 0 {
		t.Fatal("setup: visitor should be upgraded")
	}

	// 等 11min（visitor 已 idle）
	m.DecayInactive(startMs + 3*60*1000 + 11*60*1000)
	c = m.Get(testSuiteID)
	if len(c.Persons) != 0 {
		t.Errorf("visitor should be decayed after 11min idle, still has %d persons", len(c.Persons))
	}
}

// TestDecayCandidate 未升格 candidate 长期静止 → 清理（防 ghost 累积）
func TestDecayCandidate(t *testing.T) {
	m := newTestManager()
	startMs := int64(6_000_000_000)

	// 出现一次 track，但只来过 1 帧（不够升格）
	_, _ = m.UpdatePersonFromTrack(testSuiteID, "", testRadarTrk2,
		false, 1, true, startMs)
	c := m.Get(testSuiteID)
	if len(c.Persons) != 1 {
		t.Fatalf("setup: should have 1 candidate, got %d", len(c.Persons))
	}

	// 等 candidate 超过 2 × VisitorAnchorMs = 4min → 清理
	m.DecayInactive(startMs + 5*60*1000)
	c = m.Get(testSuiteID)
	if len(c.Persons) != 0 {
		t.Errorf("idle candidate should be decayed, still has %d persons", len(c.Persons))
	}
}

// TestSleepadAnchoredSync resident 在床 → 离床 → 重新在床，SleepadAnchored 实时同步（review Fix-1）
func TestSleepadAnchoredSync(t *testing.T) {
	m := newTestManager()
	startMs := int64(8_000_000_000)

	// t=0 sleepad InBed → 升 resident + SleepadAnchored=true
	p, ok := m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true, 0, true, startMs)
	if !ok || !p.SleepadAnchored {
		t.Fatalf("setup: should be sleepad anchored, ok=%v anchored=%v", ok, p.SleepadAnchored)
	}

	// t=1s 老人下床（sleepadInBed=false）→ SleepadAnchored 同步翻 false
	p, ok = m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		false, 1, true, startMs+1000)
	if !ok {
		t.Fatal("update existing resident should return ok=true")
	}
	if p.SleepadAnchored {
		t.Error("SleepadAnchored should be false after leaving bed (sleepadInBed=false)")
	}
	if p.Role != SuitePersonResident {
		t.Errorf("role should remain resident, got %s", p.Role)
	}

	// t=2s 老人重新上床 → SleepadAnchored 回 true
	p, _ = m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true, 0, false, startMs+2000)
	if !p.SleepadAnchored {
		t.Error("SleepadAnchored should be true after re-entering bed")
	}
}

// TestMarkActiveCrossZone bathroom 内 active 也更新 LastActiveMs（review Fix-2）
func TestMarkActiveCrossZone(t *testing.T) {
	m := newTestManager()
	startMs := int64(9_000_000_000)

	// 升 resident
	_, _ = m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true, 0, true, startMs)

	// 移到 bathroom
	if !m.MarkPersonExitToBathroom(testSuiteID, testResident, startMs+1000) {
		t.Fatal("setup: MarkPersonExitToBathroom failed")
	}

	// bathroom roomengine 喂 active（即使 bedroom 内无 frames，LastActiveMs 仍更新）
	if !m.MarkActiveCrossZone(testSuiteID, testResident, startMs+5*60*1000) {
		t.Fatal("MarkActiveCrossZone failed for known resident")
	}

	c := m.Get(testSuiteID)
	if c.Persons[testResident].LastActiveMs != startMs+5*60*1000 {
		t.Errorf("LastActiveMs not updated by MarkActiveCrossZone")
	}

	// 不存在的 person → return false
	if m.MarkActiveCrossZone(testSuiteID, "unknown-person", startMs+10*60*1000) {
		t.Error("MarkActiveCrossZone should return false for unknown person")
	}
}

// TestClearAnchorOnLostTrack track 失锁清空 AnchorTrackID（review Doc-5 / PR-3 wire）
func TestClearAnchorOnLostTrack(t *testing.T) {
	m := newTestManager()
	startMs := int64(10_000_000_000)

	// 升 resident，AnchorTrackID = testRadarTrk1
	_, _ = m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true, 0, true, startMs)

	// track 失锁
	if !m.ClearAnchorOnLostTrack(testSuiteID, testRadarTrk1) {
		t.Fatal("ClearAnchorOnLostTrack should succeed")
	}
	c := m.Get(testSuiteID)
	p := c.Persons[testResident]
	if p == nil {
		t.Fatal("resident should still exist after track lost")
	}
	if p.AnchorTrackID != 0 {
		t.Errorf("AnchorTrackID should be cleared to 0 after lost, got %d", p.AnchorTrackID)
	}
	// 其它字段不动
	if p.Role != SuitePersonResident {
		t.Error("Role should remain resident")
	}
	if !p.SleepadAnchored {
		t.Error("SleepadAnchored should remain (resident still in bed when track lost)")
	}

	// 重复调用 / 不存在 trackID → return false
	if m.ClearAnchorOnLostTrack(testSuiteID, testRadarTrk1) {
		t.Error("ClearAnchorOnLostTrack should return false on second call (already cleared)")
	}
}

// TestSleepadUpgradeExistingCandidate 已存在 candidate 时 sleepad 锚定 → 立即升 resident（替换 candidate）
func TestSleepadUpgradeExistingCandidate(t *testing.T) {
	m := newTestManager()
	startMs := int64(7_000_000_000)

	// 先 1 帧 candidate
	_, ok := m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		false, 1, true, startMs)
	if ok {
		t.Fatal("first frame should not upgrade")
	}

	// 第 2 帧带 sleepad → 立即升 resident
	p, ok := m.UpdatePersonFromTrack(testSuiteID, testResident, testRadarTrk1,
		true, 0, true, startMs+1000)
	if !ok {
		t.Fatal("sleepad anchor should upgrade existing candidate")
	}
	if p.Role != SuitePersonResident {
		t.Errorf("role want resident, got %s", p.Role)
	}
	if !p.SleepadAnchored {
		t.Error("SleepadAnchored should be true")
	}

	c := m.Get(testSuiteID)
	if len(c.Persons) != 1 {
		t.Errorf("census should have exactly 1 person (candidate replaced), got %d", len(c.Persons))
	}
	// 应只有 resident，不应有 candidate
	if _, hasCand := c.Persons[candidateKeyForTrack(testRadarTrk1)]; hasCand {
		t.Error("candidate should be removed after sleepad upgrade")
	}
}
