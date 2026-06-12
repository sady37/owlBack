package belief

import (
	"testing"

	"owl-common/observation"
)

// 3-case 回归 oracle（room_belief_state_machine.md §7）：
// 复现现状判对的必须仍对（真跌倒）、判错的必须修对（CABB / John.Y）。
// v1 scope = 单实体 + §5.5.2 弱耦合。

func ob(ts int64, kind ObsKind, val, conf float64, area int) Observation {
	return Observation{Kind: kind, Value: val, Conf: conf, Ts: ts, Fresh: true, AreaType: area}
}

// Decider 确认窗：维持 ≥confirmMs 才确认 Fall；中途跌回 θ 下（如人返回）则永不确认。
// 治本 cd2b 类"fire 后恢复"误报，同时真跌倒（持续维持）仍报。
func TestDeciderConfirmWindow(t *testing.T) {
	high := Vector{SFallen: 0.8, SStandWalk: 0.2}
	low := Vector{SStandWalk: 0.7, SEmpty: 0.3}

	var sustained Decider
	now, fired := int64(0), false
	for i := 0; i < 200; i++ {
		now += 1000
		if sustained.Update(high, now) == DecisionFall {
			fired = true
			break
		}
	}
	if !fired {
		t.Fatal("持续 Fallen 应在 confirmMs 后确认 Fall")
	}

	var recovered Decider
	now = 0
	for i := 0; i < 40; i++ { // 40s < 90s confirm
		now += 1000
		if recovered.Update(high, now) == DecisionFall {
			t.Fatal("40s < confirmMs 不该提前确认")
		}
	}
	for i := 0; i < 200; i++ { // 人返回，P 崩回 low
		now += 1000
		if recovered.Update(low, now) == DecisionFall {
			t.Fatal("恢复（跌回 θ 下）后永不该确认 Fall")
		}
	}
}

// 真跌倒（cabb-fall 同型）：开阔地板行走 → z 骤降 + firmware 确认 + pose=fallen。
// 期望 P(Fallen)>θ_fire → DecisionFall。今天判对，必须仍对。
func TestGenuineFall(t *testing.T) {
	be := New(DefaultModel())
	ts := int64(1000)
	// 进房 + 几帧行走
	be.Step(ts, []Observation{ob(ts, ObsEnterExit, +1, 0.9, areaEnter)})
	for i := 0; i < 5; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseWalking, 0.8, areaActive)})
	}
	// 跌倒帧 + 持续躺地。断言**不变量**:持续 pose-fallen 在窗内必确认 Fall —— 不锁具体帧数。
	// 删 Δz(P2.1)/ firmware 降权(P2.4)后,单帧冲击退场,genuine-fall 由**正向 pose 多帧累积**确认;
	// 真摔本持续躺地(cabb-0605 躺 52s)。latency 是 shadow-moot 量(belief Decide 不接 alarm;
	// production fall 走 firmware Device_ALARM 独立路径);单帧响应议题挂 P3 选项D(可信 XY-jerk),不回退 pose/firmware 提权。
	ts += 1000
	be.Step(ts, []Observation{
		ob(ts, ObsPose, observation.PoseFallen, 0.8, areaActive),
	})
	fired := be.Decide() == DecisionFall
	for i := 0; i < 6 && !fired; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseFallen, 0.8, areaActive)})
		fired = be.Decide() == DecisionFall
	}
	if !fired {
		t.Fatalf("genuine fall 持续躺地窗内未确认: b=%v", be.Vector())
	}
}

// CABB lost_track（今天误报）：浴室内行走至门口 track 丢失，firmware 未发 ExitRoom，
// 全程无任何跌倒签名。期望 track 丢后信念漂向 Left/Empty，P(Fallen) 低 → 不报 Fall。
func TestCabbLostTrackNoFall(t *testing.T) {
	be := New(DefaultModel())
	ts := int64(1000)
	be.Step(ts, []Observation{ob(ts, ObsEnterExit, +1, 0.9, areaEnter)})
	// 浴室内 Walking（pose=1×多帧，z 直立，从无 pose5/2）走到门口
	for i := 0; i < 6; i++ {
		ts += 1000
		geom := areaToilet
		if i >= 4 {
			geom = areaEnter // 走到门口区
		}
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseWalking, 0.8, geom)})
	}
	// track 丢失：此后无观测（stale）。仅时间推进（5min ≈ lost_fall 等待窗）。
	for i := 0; i < 30; i++ {
		ts += 10_000
		be.Step(ts, nil)
	}
	// 修对 = 不再误报 Fall。firmware 没发 ExitRoom，belief 无法观测"已离开"，故对
	// in-room/left 摊平、输出 DecisionUncertain（§8 诚实边界：不创造可观测性，给恰当不确定度
	// 让决策升级人工而非瞎猜）——但 P(Fallen) 必须被压到极低，Fallen 绝不能是主假设。
	if d := be.Decide(); d == DecisionFall {
		t.Fatalf("CABB lost_track falsely fired Fall: b=%v", be.Vector())
	}
	if be.Vector()[SFallen] > 0.05 {
		t.Fatalf("CABB P(Fallen)=%.3f 应被压到极低", be.Vector()[SFallen])
	}
	if arg, _ := be.Vector().Max(); arg == SFallen {
		t.Fatalf("CABB Fallen 不该是主假设: b=%v", be.Vector())
	}
}

// John.Y 9h person_silent（今天误报）：D523 无床雷达把 track 静止冻结在开阔地板，人实际走进
// 床区（09E7+sleepad，D523 盲区）。命门：长时间静止(StillBox) → ObsPose stale(Conf=0) 不更新；
// 弱耦合 sleepad InBed 作 ObsNeighbor 压低 P(Fallen)。期望 9h 内 P(Fallen) 永不起来。
func TestJohnY9hNoSilentFall(t *testing.T) {
	be := New(DefaultModel())
	ts := int64(1000)
	be.Step(ts, []Observation{ob(ts, ObsEnterExit, +1, 0.9, areaEnter)})
	// 走向床区边界（D523 视野里最后几帧 walking）
	for i := 0; i < 4; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseWalking, 0.8, areaActive)})
	}
	// track 长时间静止：D523 这路停止更新（命门——stale 观测 effConf=0）。
	// 同时 sleepad 在 09E7 床区报 InBed → 弱耦合 ObsNeighbor（邻居占用）。
	for i := 0; i < 60; i++ { // 60×~9min step 覆盖 9h 量级
		ts += 540_000
		stale := ob(ts, ObsPose, observation.PoseStanding, 0.8, areaActive)
		stale.Fresh = false // 静止超时=非新鲜，effConf=0，不更新（治本命门）
		neighbor := ob(ts, ObsNeighbor, 0.9, 0.85, areaUnknown)
		be.Step(ts, []Observation{stale, neighbor})
	}
	if d := be.Decide(); d == DecisionFall {
		t.Fatalf("John.Y 9h falsely fired person_silent Fall: b=%v", be.Vector())
	}
	if be.Vector()[SFallen] > thFire {
		t.Fatalf("John.Y P(Fallen)=%.3f over θ_fire", be.Vector()[SFallen])
	}
	resolved := be.Vector()[SEmpty] + be.Vector()[SLeft]
	if resolved < 0.5 {
		t.Fatalf("John.Y should resolve to person-left/elsewhere, got %.3f (b=%v)", resolved, be.Vector())
	}
}

// 命门对照：同一静止 pose，Fresh=true（=今天 census 当活观测的 bug）会把信念钉死在原态，
// 不再随时间漂向 Left；Fresh=false（治本）才让 A+邻居证据接管。
func TestStaleEvidenceIsLinchpin(t *testing.T) {
	run := func(fresh bool) Vector {
		be := New(DefaultModel())
		ts := int64(1000)
		for i := 0; i < 4; i++ {
			ts += 1000
			be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseStanding, 0.8, areaActive)})
		}
		for i := 0; i < 40; i++ {
			ts += 10_000
			o := ob(ts, ObsPose, observation.PoseStanding, 0.8, areaActive)
			o.Fresh = fresh
			be.Step(ts, []Observation{o})
		}
		return be.Vector()
	}
	pinned := run(true)   // bug：静止当活观测
	drifted := run(false) // fix：静止超时=缺证据
	if drifted[SStandWalk] >= pinned[SStandWalk] {
		t.Fatalf("freshness 命门未体现：fix 应让 StandWalk 信念衰减 (fix=%.3f bug=%.3f)",
			drifted[SStandWalk], pinned[SStandWalk])
	}
}

// TestGeomProvenanceWeighting：AreaConf(Source 信任) 越低，rest-geom 对跌倒的抑制越弱（软先验替硬闸）。
func TestGeomProvenanceWeighting(t *testing.T) {
	mk := func(gc float64) Observation {
		return Observation{Kind: ObsPose, Value: float64(observation.PoseFallen), AreaType: areaBed, AreaConf: gc, Conf: 0.8, Fresh: true, Ts: 1000}
	}
	full := rawLikelihood(mk(1.0))      // FE 画 bed：全抑制（SFallen 最低）
	tentative := rawLikelihood(mk(0.4)) // 自学 bed：抑制打折
	unset := rawLikelihood(mk(0))       // 未设=全信=同 full（向后兼容）
	if tentative[SFallen] <= full[SFallen] {
		t.Fatalf("暂定 bed 应抑制更弱(SFallen 更高): full=%.2f tentative=%.2f", full[SFallen], tentative[SFallen])
	}
	if unset[SFallen] != full[SFallen] {
		t.Fatalf("AreaConf=0 应=全信(向后兼容): unset=%.2f full=%.2f", unset[SFallen], full[SFallen])
	}
}

// TestBedsideFallBedReleased — 床边 FN 红线回归（D3 止血换载体：silent_leftbed 判据进 DBN，由 bed_state 驱动）。
// 床区(area=Bed)躺：bed_state 占用=睡觉 → SBedLying 主导、SFallen 中性；bed_state 离床(BedReleased)=床边真摔 →
// SFallen 抬（倒地候选，等价开阔地躺）。锁住"lying@床区 不再被几何盖掉 bed_state"——床边真摔不再 FN。
func TestBedsideFallBedReleased(t *testing.T) {
	lying := func(released bool) Vector {
		return rawLikelihood(Observation{Kind: ObsPose, Value: float64(observation.PoseLying), AreaType: areaBed, BedReleased: released, Conf: 1, Fresh: true, Ts: 1000})
	}
	sleeping := lying(false) // 床态占用 = 睡觉豁免
	fallen := lying(true)    // 床态离床 = 床边真摔

	if sleeping[SFallen] > 1.0 {
		t.Fatalf("床区躺+床占用=睡觉,SFallen 不该抬: %.3f", sleeping[SFallen])
	}
	if sleeping[SBedLying] <= 1.0 {
		t.Fatalf("床区躺+床占用=睡觉,SBedLying 应主导: %.3f", sleeping[SBedLying])
	}
	if fallen[SFallen] <= 1.0 {
		t.Fatalf("床区躺+bed_state 离床=床边真摔,SFallen 应抬(倒地候选): %.3f", fallen[SFallen])
	}
	// 止血语义 = 翻 OpenFloor：床区离床躺的 SFallen 须等于开阔地躺。
	openFloor := rawLikelihood(Observation{Kind: ObsPose, Value: float64(observation.PoseLying), AreaType: areaActive, Conf: 1, Fresh: true, Ts: 1000})
	if fallen[SFallen] != openFloor[SFallen] {
		t.Fatalf("床区离床躺应=开阔地躺(止血翻 OpenFloor 语义): bedReleased=%.3f openFloor=%.3f", fallen[SFallen], openFloor[SFallen])
	}
	// 门优先：床区但近门(NearDoor)→ 视为门区走 default，不因 BedReleased 误判倒地(复刻 geomFromGrid 门优先)。
	nearDoor := rawLikelihood(Observation{Kind: ObsPose, Value: float64(observation.PoseLying), AreaType: areaBed, NearDoor: true, BedReleased: true, Conf: 1, Fresh: true, Ts: 1000})
	if nearDoor[SFallen] != lrPoseLyingDefFall {
		t.Fatalf("床区近门躺应走 default(门优先),SFallen=%.3f want %.3f", nearDoor[SFallen], lrPoseLyingDefFall)
	}
}

// TestFallGeomRouting #3：**可观测**跌倒的几何路由（设计意图回归锁，非真机验证——真机靠 shadow + 201 测试）。
// 仅覆盖"radar 看得到跌倒姿/运动学"的子情形；occluded fall-off-bed（雷达完全看不到，靠 LeftBed+不重捕推断）
// 不在此测，留给 201 真机 fixture 驱动 LeftBed-armed 推断设计。
// 设计意图：床=躺非摔 → Fallen@InBed 降权（多为躺姿误读）；下床摔在床边地板=OpenFloor 不压；桶区塌陷默认不压。
func TestFallGeomRouting(t *testing.T) {
	enter := func() (*Belief, int64) {
		be := New(DefaultModel())
		ts := int64(1000)
		be.Step(ts, []Observation{ob(ts, ObsEnterExit, +1, 0.9, areaEnter)})
		return be, ts
	}

	// (a) FP 向：床上某帧被误读 pose=Fallen → BedLying 解释胜出，不该判主假设 Fallen。
	be, ts := enter()
	for i := 0; i < 5; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseLying, 0.8, areaBed)})
	}
	ts += 1000
	be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseFallen, 0.8, areaBed)})
	if arg, _ := be.Vector().Max(); arg == SFallen {
		t.Fatalf("床上 pose=Fallen 不该判主假设 Fallen（床=躺非摔）: b=%v", be.Vector())
	}

	// (b) FN 向：下床摔在床边地板（OpenFloor）→ 必须能 fire（不因身处卧室就被压）。
	be, ts = enter()
	for i := 0; i < 3; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseWalking, 0.8, areaActive)})
	}
	ts += 1000
	be.Step(ts, []Observation{
		ob(ts, ObsPose, observation.PoseFallen, 0.8, areaActive),
	})
	fired := be.Decide() == DecisionFall
	for i := 0; i < 6 && !fired; i++ { // 持续躺地累积(不锁帧数,latency shadow-moot)
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseFallen, 0.8, areaActive)})
		fired = be.Decide() == DecisionFall
	}
	if !fired {
		t.Fatalf("床边地板真跌倒被几何路由漏报: b=%v", be.Vector())
	}

	// (c) FN 向：马桶区塌陷（InToilet）→ 默认不被特殊压制，应能 fire。
	be, ts = enter()
	for i := 0; i < 3; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseWalking, 0.8, areaToilet)})
	}
	ts += 1000
	be.Step(ts, []Observation{
		ob(ts, ObsPose, observation.PoseFallen, 0.8, areaToilet),
	})
	fired = be.Decide() == DecisionFall
	for i := 0; i < 6 && !fired; i++ { // 持续躺地累积(不锁帧数)
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsPose, observation.PoseFallen, 0.8, areaToilet)})
		fired = be.Decide() == DecisionFall
	}
	if !fired {
		t.Fatalf("马桶区塌陷被几何路由漏报: b=%v", be.Vector())
	}
}

// P2.3 z 三档:z 只喂 posture(高度档),**绝不进 fall**(R5:z 不确认不否决跌倒)。
func TestZBandPostureNotFall(t *testing.T) {
	mkEnter := func() (*Belief, int64) {
		be := New(DefaultModel())
		ts := int64(1000)
		be.Step(ts, []Observation{ob(ts, ObsEnterExit, +1, 0.9, areaEnter)})
		return be, ts
	}
	// (a) z>80 持续 → 偏 stand,绝不 fire fall。
	be, ts := mkEnter()
	for i := 0; i < 10; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsZBand, 160, 0.7, areaActive)})
	}
	if be.Decide() == DecisionFall {
		t.Fatalf("z>80 不该触发 fall(R5 z 不进 fall): b=%v", be.Vector())
	}
	if v := be.Vector(); v[SStandWalk] <= v[SFallen] {
		t.Fatalf("z>80 应偏 stand 而非 fallen: stand=%.3f fallen=%.3f", v[SStandWalk], v[SFallen])
	}
	// (b) z<30 假低持续 → 无信息,不自行 fire fall。
	be, ts = mkEnter()
	for i := 0; i < 10; i++ {
		ts += 1000
		be.Step(ts, []Observation{ob(ts, ObsZBand, 10, 0.7, areaActive)})
	}
	if be.Decide() == DecisionFall {
		t.Fatalf("z<30 假低不该触发 fall: b=%v", be.Vector())
	}
}

// TestP6NoDetectFallenGate — P6.1a(阻塞项#1):no-detect 抬 Fallen 受 R_i + door-exit 门控。
// 同样"走动者消失"序列,仅门控参数不同:真人非门区(Ri=1,door=0)抬 Fallen 最高;
// ghost 消失(Ri→0)或门区可达走出(door→1)→ 因子→1.0 中性 → 不裸 absence 抬 fall(治 dropout-FP)。
func TestP6NoDetectFallenGate(t *testing.T) {
	run := func(ri, dx float64) float64 {
		be := New(DefaultModel())
		now := int64(1_000)
		for i := 0; i < 4; i++ { // 先建立"走动者在场"(否则全在 Empty,no-detect 无对象)
			now += 1000
			be.Step(now, []Observation{ob(now, ObsPose, float64(observation.PoseWalking), 0.8, areaActive)})
		}
		for i := 0; i < 8; i++ { // 走动者消失:repeated no-detect,门控 (ri,dx)
			now += 1000
			be.Step(now, []Observation{{Kind: ObsNoDetect, Conf: 0.8, Ts: now, Fresh: true, AreaType: areaActive, RealnessP: ri, DoorExitP: dx}})
		}
		return be.Vector().P(SFallen)
	}
	pReal := run(1.0, 0.0)  // 真人非门区消失 → 抬 Fallen
	pGhost := run(0.0, 0.0) // ghost 消失 → 因子 1.0 不抬
	pDoor := run(1.0, 1.0)  // 门区可达走出 → 因子 1.0 不抬
	t.Logf("P6.1a no-detect gate: real-non-door=%.3f ghost=%.3f door=%.3f", pReal, pGhost, pDoor)
	if pReal <= pGhost {
		t.Fatalf("真人非门区消失应比 ghost 消失更抬 Fallen(ghost 不裸抬):real=%.3f ghost=%.3f", pReal, pGhost)
	}
	if pReal <= pDoor {
		t.Fatalf("真人非门区消失应比门区可达走出更抬 Fallen(门区不裸抬):real=%.3f door=%.3f", pReal, pDoor)
	}
}

// TestP6DoorFallVsExit — 审查㉓ door-exit 放行前置对抗例:门口真摔 vs 门口离场。
// 二者门距运动学相同(doorExit≈1),**判别靠 exit 事件非门距否决**:
//   门口真摔 = 朝门逼近→末帧栽倒→**无** ExitRoom/track守恒重现 → door-exit 留 floor 的残余 Fallen 经 lost 窗累积 → 仍越 τ 浮出(不漏报)。
//   门口离场 = 同逼近→**有** ExitRoom 事件(SLeft:8 强压)→ Fallen 不越 τ(不 FP)。
// 证 door-exit 不全否决(floor 起作用),且真离场靠事件区分。
func TestP6DoorFallVsExit(t *testing.T) {
	const doorExitHi = 1.0 // 门口:近门+定向可达 → reachableExitScore≈1
	walkIn := func(be *Belief, now *int64) {
		for i := 0; i < 4; i++ { // 朝门走动(建立走动者在场)
			*now += 1000
			be.Step(*now, []Observation{ob(*now, ObsPose, float64(observation.PoseWalking), 0.8, areaEnter)})
		}
	}
	// 门口真摔:逼近后栽倒消失,无 exit 事件 → 重复 no-detect(door-exit 留 floor) → 应浮出。
	beFall := New(DefaultModel())
	nowF := int64(1_000)
	walkIn(beFall, &nowF)
	firedFall := false
	for i := 0; i < 40 && !firedFall; i++ {
		nowF += 1000
		beFall.Step(nowF, []Observation{{Kind: ObsNoDetect, Conf: 0.8, Ts: nowF, Fresh: true, AreaType: areaEnter, RealnessP: 1.0, DoorExitP: doorExitHi}})
		firedFall = beFall.Decide() == DecisionFall
	}
	// 门口离场:同逼近,但真离场 = ExitRoom 事件 + **房空 np=0 持续**(走出后房里没人)→ SLeft/SEmpty 持续压 → 不应浮出。
	// (判别靠"离场事件+持续空房"强信号,非门距——door-exit floor 对两者相同,事件才是区分。)
	beExit := New(DefaultModel())
	nowE := int64(1_000)
	walkIn(beExit, &nowE)
	nowE += 1000
	beExit.Step(nowE, []Observation{{Kind: ObsEnterExit, Value: -1, Conf: 0.9, Ts: nowE, Fresh: true, AreaType: areaEnter}}) // ExitRoom
	firedExit := false
	for i := 0; i < 40 && !firedExit; i++ {
		nowE += 1000
		beExit.Step(nowE, []Observation{
			{Kind: ObsNoDetect, Conf: 0.8, Ts: nowE, Fresh: true, AreaType: areaEnter, RealnessP: 1.0, DoorExitP: doorExitHi},
			{Kind: ObsNumberPeople, Value: 0, Conf: 0.8, Ts: nowE, Fresh: true, AreaType: areaEnter}, // 房空(走出后)
		})
		firedExit = beExit.Decide() == DecisionFall
	}
	t.Logf("P6.1a door 对抗: 门口真摔 fired=%v P(Fallen)=%.3f / 门口离场 fired=%v P(Fallen)=%.3f",
		firedFall, beFall.Vector().P(SFallen), firedExit, beExit.Vector().P(SFallen))
	if !firedFall {
		t.Fatalf("门口真摔(无exit事件)应越τ浮出(door-exit 不全否决,floor 累积):P(Fallen)=%.3f", beFall.Vector().P(SFallen))
	}
	if firedExit {
		t.Fatalf("门口离场(有ExitRoom)不应浮出(SLeft 强压):P(Fallen)=%.3f", beExit.Vector().P(SFallen))
	}
}
