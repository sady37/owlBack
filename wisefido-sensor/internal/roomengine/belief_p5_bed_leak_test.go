package roomengine

import (
	"testing"

	"owl-common/observation"
	"wisefido-sensor/internal/roomengine/belief"
)

// P5(审查㊴ R5 裁定 P5c)bed O_b 迟滞 + R5-clean 合取门控 治 α LeftBed-co-fire。
// 放行前置(委员会㊴):① 8 CD2B α 案机理 → shadow 压制;② 滚下床真摔对抗例 → 不压仍 escalate(R5 不漏报);
// ③ R5 专项断言:压制路径不含任何 z高/pose 作压制因子(防 P5a 残留)。
// 全 belief shadow(只 log 不 fire),不碰生产。

func p5ob(ts int64, kind belief.ObsKind, val, conf float64, g belief.Geom) belief.Observation {
	return belief.Observation{Kind: kind, Value: val, Conf: conf, Ts: ts, Fresh: true, Geom: g}
}

// TestP5BedLeakBlipSticky — brief LeftBed 翻拍(InBed 很快返回)→ authority 粘滞高(治 α 翻身伪迹)。
func TestP5BedLeakBlipSticky(t *testing.T) {
	var b bedLeakState
	ts := int64(1000)
	if got := b.update(true, ts); got != 1 { // InBed → 回满
		t.Fatalf("InBed authority 应=1,得 %v", got)
	}
	ts += 2000 // 2s LeftBed blip
	if got := b.update(false, ts); got < 0.9 {
		t.Fatalf("2s LeftBed blip 后 authority 应仍粘滞 >0.9(治翻身伪迹),得 %v", got)
	}
	ts += 1000 // InBed 返回
	if got := b.update(true, ts); got != 1 {
		t.Fatalf("InBed 返回 → authority 回满,得 %v", got)
	}
}

// TestP5BedLeakSustainedDecays — sustained LeftBed no-return → authority 衰减 → 床权威掉(让位真离床/滚下床)。
func TestP5BedLeakSustainedDecays(t *testing.T) {
	var b bedLeakState
	ts := int64(1000)
	b.update(true, ts) // InBed → 1
	ts += 120_000      // 持续 120s LeftBed 不返回
	got := b.update(false, ts)
	if got >= 0.5 {
		t.Fatalf("120s sustained LeftBed 后 authority 应显著衰减(<0.5,让位真离床),得 %v", got)
	}
}

// TestP5GateConjunction — R5-clean 合取门控:bedVal 透出仅当 占用(authority)∧ 位置(radarOnBed)。
func TestP5GateConjunction(t *testing.T) {
	ts := int64(1000)
	if v := bedAuthorityObs(0.9, true, ts).Value; v != 0.9 {
		t.Fatalf("占用∧on-bed → bedVal=authority(0.9),得 %v", v)
	}
	if v := bedAuthorityObs(0.9, false, ts).Value; v != 0 {
		t.Fatalf("radar 离床(displaced)→ bedVal=0 不压(默认 escalate),得 %v", v)
	}
	if v := bedAuthorityObs(0, true, ts).Value; v != 0 {
		t.Fatalf("authority=0(sustained 衰减殆尽)→ bedVal=0,得 %v", v)
	}
}

// TestP5AlphaSuppressed — α LeftBed-co-fire(床上翻身/坐起 + sleepad blip + firmware Fall,radar 仍 on-bed)
// → 床权威压制 SFallen → P(Fallen) 远低于无压制基线(治 ≥5/8 CD2B 卧室 FP)。
func TestP5AlphaSuppressed(t *testing.T) {
	ts := int64(1000)
	// 共同前置:数帧 on-bed 躺(radar geom InBed)。
	build := func(withBed bool) *belief.Belief {
		be := belief.New(belief.DefaultModel())
		tt := ts
		var leak bedLeakState
		for i := 0; i < 5; i++ {
			tt += 1000
			leak.update(true, tt) // sleepad InBed
			o := []belief.Observation{p5ob(tt, belief.ObsPose, observation.PoseLying, 0.8, belief.GeomInBed)}
			if withBed {
				o = append(o, bedAuthorityObs(leak.update(true, tt), true, tt))
			}
			be.Step(tt, o)
		}
		// co-fire tick:sleepad 瞬 LeftBed(blip)+ radar 把床上翻身误读 pose=fallen@InBed(真生产抬升源,
		// WF-b:shadow 独立判,无 firmware Fall),radar 仍 on-bed(geom InBed)。
		tt += 1000
		auth := leak.update(false, tt) // LeftBed blip → 迟滞仍高
		o := []belief.Observation{
			p5ob(tt, belief.ObsPose, observation.PoseFallen, 0.8, belief.GeomInBed),
		}
		if withBed {
			o = append(o, bedAuthorityObs(auth, true, tt)) // radarOnBed=true → 合取成立 → 压制
		}
		be.Step(tt, o)
		return be
	}
	suppressed := build(true)
	baseline := build(false)
	pSup := suppressed.Vector().P(belief.SFallen)
	pBase := baseline.Vector().P(belief.SFallen)
	if pSup >= pBase {
		t.Fatalf("P5 床权威应压制 α:有床 P(Fallen)=%.4f 应 < 无床基线 %.4f", pSup, pBase)
	}
	if suppressed.Decide() == belief.DecisionFall {
		t.Fatalf("α(床上翻身伪迹)被压制后不应确认 Fall;P(Fallen)=%.4f", pSup)
	}
}

// TestP5RollOffBedNotSuppressed — 滚下床真摔对抗例(R5 不漏报):sustained LeftBed + radar 位移离床(OpenFloor)
// + z低躺地 + firmware Fall + PoseFallen@Open → 闸打开(bedVal=0)→ 不压 → P(Fallen) 高 → 确认 Fall。
func TestP5RollOffBedNotSuppressed(t *testing.T) {
	be := belief.New(belief.DefaultModel())
	tt := int64(1000)
	var leak bedLeakState
	// on-bed 躺数帧。
	for i := 0; i < 5; i++ {
		tt += 1000
		be.Step(tt, []belief.Observation{
			p5ob(tt, belief.ObsPose, observation.PoseLying, 0.8, belief.GeomInBed),
			bedAuthorityObs(leak.update(true, tt), true, tt),
		})
	}
	// 滚下床:radar 位移到 OpenFloor(radarOnBed=false),sleepad sustained LeftBed,z低,firmware Fall + 倒地 pose。
	fired := false
	for i := 0; i < 8 && !fired; i++ {
		tt += 1000
		auth := leak.update(false, tt) // 持续 LeftBed
		be.Step(tt, []belief.Observation{
			p5ob(tt, belief.ObsPose, observation.PoseFallen, 0.8, belief.GeomOpenFloor), // 位移离床倒地=真抬升源
			p5ob(tt, belief.ObsZBand, 10, 0.7, belief.GeomOpenFloor),                    // z低躺地(只正向,不压)
			bedAuthorityObs(auth, false, tt),                                            // radarOnBed=false → bedVal=0 不压
		})
		fired = be.Decide() == belief.DecisionFall
	}
	if !fired {
		t.Fatalf("滚下床真摔(位移离床)绝不可被 P5 压制 → 必确认 Fall(R5 红线);P(Fallen)=%.4f", be.Vector().P(belief.SFallen))
	}
}

// TestP5SuppressorIgnoresZ — R5 专项断言(防 P5a 残留):压制由 占用∧位置 驱动,与 z **无关**。
// z低(15,P5a 会判"真摔不压")时,只要 radar on-bed ∧ authority → 仍压制(P5c 不看 z)。
// 对照 z高(205):两者压制等同 → 证 z 不是压制因子。
func TestP5SuppressorIgnoresZ(t *testing.T) {
	runAlpha := func(zVal float64) *belief.Belief {
		be := belief.New(belief.DefaultModel())
		tt := int64(1000)
		var leak bedLeakState
		for i := 0; i < 5; i++ {
			tt += 1000
			be.Step(tt, []belief.Observation{
				p5ob(tt, belief.ObsPose, observation.PoseLying, 0.8, belief.GeomInBed),
				p5ob(tt, belief.ObsZBand, zVal, 0.7, belief.GeomInBed),
				bedAuthorityObs(leak.update(true, tt), true, tt),
			})
		}
		tt += 1000
		auth := leak.update(false, tt) // LeftBed blip
		be.Step(tt, []belief.Observation{
			p5ob(tt, belief.ObsPose, observation.PoseFallen, 0.8, belief.GeomInBed), // 床上翻身误读=真抬升源
			p5ob(tt, belief.ObsZBand, zVal, 0.7, belief.GeomInBed),
			bedAuthorityObs(auth, true, tt), // radarOnBed=true,z 不参与门控
		})
		return be
	}
	zLow := runAlpha(15)  // P5a 会当"位移躺地真摔"→ 不压;P5c 仍压
	zHigh := runAlpha(205) // 留床面
	// R5 专项断言:z低时仍被压制(不确认 Fall)——若有 z 压制因子残留,z低应放行。
	if zLow.Decide() == belief.DecisionFall {
		t.Fatalf("R5 残留:z低 α 不应确认 Fall(压制必须 z-无关);P(Fallen)=%.4f", zLow.Vector().P(belief.SFallen))
	}
	// 两者 SFallen 压制方向一致(均被床权威压低),z 不改变压制存在性。
	pLow := zLow.Vector().P(belief.SFallen)
	pHigh := zHigh.Vector().P(belief.SFallen)
	if pLow > 0.5 || pHigh > 0.5 {
		t.Fatalf("R5:占用∧位置 压制应 z-无关使两者 P(Fallen) 均低;zLow=%.4f zHigh=%.4f", pLow, pHigh)
	}
}
