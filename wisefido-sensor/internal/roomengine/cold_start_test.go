package roomengine

import "testing"

// 工单5 cold-start per-unit 升档闸:锁 unitCap 时钟门 + effectiveMode=min(ceiling,cap) 语义。
func TestColdStartUnitGate(t *testing.T) {
	const hourMs = int64(60 * 60 * 1000)
	graduate := coldGraduateMs
	if coldFloorMs > graduate {
		graduate = coldFloorMs
	}

	e := &Engine{
		unitFirstTrackMs: map[string]int64{},
		roomSuiteID:      map[string]string{"room1": "unitA"},
	}

	// 1) markUnitTrack 单调:只落首次,不覆盖。
	e.markUnitTrack("unitA", 1_000)
	e.markUnitTrack("unitA", 9_000)
	if got := e.unitFirstTrackMs["unitA"]; got != 1_000 {
		t.Fatalf("markUnitTrack 应只落首次, got %d want 1000", got)
	}

	first := int64(1_000)
	// 2) unitCap 时钟门:从没见 track / cold 期内 → 1;满 graduate → 2。
	if cap := e.unitCap("unknown", first); cap != 1 {
		t.Errorf("未见 track 的 unit 应 cap=1, got %d", cap)
	}
	if cap := e.unitCap("unitA", first+graduate-1); cap != 1 {
		t.Errorf("cold 期内应 cap=1, got %d", cap)
	}
	if cap := e.unitCap("unitA", first+graduate); cap != 2 {
		t.Errorf("满 graduate 应 cap=2, got %d", cap)
	}

	// 3) T_floor 硬下限:即便只过了 24h 多一点但默认 72h 未满 → 仍 1。
	if cap := e.unitCap("unitA", first+coldFloorMs+hourMs); cap != 1 {
		t.Errorf("默认 72h 门未满(仅过 25h)应 cap=1, got %d", cap)
	}

	// 4) effectiveMode = min(全局 dbnMode, unitCap)。
	saved := dbnMode
	defer func() { dbnMode = saved }()

	coldNow := first + graduate - 1 // cold unit, cap=1
	matureNow := first + graduate   // mature, cap=2

	dbnMode = 2
	if !e.dbnSelfFireEnabledFor("unitA", coldNow) {
		t.Error("dbnMode=2 cold unit 应仍可自发(cap=1≥1)")
	}
	if e.dbnVetoFirmwareEnabledFor("room1", coldNow) {
		t.Error("dbnMode=2 cold unit 不应否决 firmware(effectiveMode=1)")
	}
	if !e.dbnVetoFirmwareEnabledFor("room1", matureNow) {
		t.Error("dbnMode=2 mature unit 应可否决 firmware(effectiveMode=2)")
	}

	dbnMode = 1
	if e.dbnVetoFirmwareEnabledFor("room1", matureNow) {
		t.Error("dbnMode=1 ceiling 下 mature unit 不应否决(min(1,2)=1)")
	}

	dbnMode = 0
	if e.dbnSelfFireEnabledFor("unitA", matureNow) {
		t.Error("dbnMode=0 运维全局静默,mature unit 也不自发(min(0,2)=0)")
	}
}
