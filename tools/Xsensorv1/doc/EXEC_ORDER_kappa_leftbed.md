# 执行批令:① κ 变活(wire UpdateKappa)+ ③ LeftBed→SOpenFloor

> 新会话照此执行。方向已定(权威 `doc/DBN-Zone-Room.md` §11),**不要重开设计**,直接实现 + build 验证。
> 工作目录:`/home/wisefido/owl/owlBack/tools/Xsensorv1`
> 前置:M×N(`EXEC_ORDER_MN_bed.md`,commit c7e8ebe)已落地;本批令是 cd2b 真解,只新建这两件。

## 目标(一句话)

- **① κ 变活**:现 `UpdateKappa` 是 orphan(零调用),κ 死在几何冷启、只 `g^xy` 半边活。接上"15s 内雷达↔sleepad 同向床事件共发 → 拉高 κ"的事件驱动 EMA。κ = **纯归属权重**,不碰 SBed 维持。
- **③ LeftBed→SOpenFloor**:sleepad 报 LeftBed 时,在 **S 轴主动满幅**抬 SOpenFloor(present)/SBlindRest(lost),按 `a_j` 归属。这是 cd2b 主路径。

## 为什么(只读,别重证)

见 DBN §11 + §4 承重不变量。要点:κ 升降 FN-safe(SBed 由"证据抬 vs 0.80 自衰减"维持,真睡者总有 κ-free 直接源 M×N/HR-RR/pose=Lying 撑,故 κ 衰减饿不死、无需 gate)。③「满幅」= 不被 Con 二次折扣(`a_j` 已分概率),靠 SOpenFloor 单位量级 ≫ SBed 做"进床慢离床快"。多住户假摔被 `g^xy` 几何门控(邻床 LeftBed 对在别床 track `g^xy≈0→a_j≈0` 不串)。

## 已确认事实(直接用,不用再查)

- `Coupling` **每房一份**(`engine.go:84 r.cp`),κ 跨房内所有 track 共享;`g^xy` 才 per-track。→ `UpdateKappa` 每帧每房调**一次**,在 per-track loop 之前。
- `UpdateKappa(matched, live []bool)`(`coupling.go:43`):仅 `live[j]=true` 才 `κ=(1-γ)κ+γ·matched`(无 max,可升可降)。
- `attachment(gxy)`(`coupling.go:58`)→ `a_j=κ·g^xy/Σ`。`Kappa(j)`(`coupling.go:37`)取当前 κ。
- 帧循环 `Room.Tick`(`engine/engine.go:132`):per-track loop 在 `:157`;present 分支算 `logPsi=r.cp.LogPsi(...Gxy...)`+`logPhi=r.em.LogPhi`;**lost 分支不算 logPsi**(只 logPhi)。
- `LogPhi = contactLogB(B轴,含 LeftBed→B vac)+ radarLogS(S轴)`(`emission.go:106`)——**present/lost 两分支都含 sleepad B 轴**。
- `js.AddLogToS(v, s, delta)`(`joint.go:122`)= 给 S==s 的所有 cell 加 delta(③ 注 S 轴用它,范本=`engine.go:196` exitL 注 SLeft)。
- `obs.Sleepad []BedReading`(`BedInBed`/`BedLeftBed`/`BedNoReport`)present/lost 都填(`BuildObservation` 吃 `fi.Sleepads`)。
- `fi.Tracks []TrackObs`,每条 `RadarTrack.FwAreaID`+`.Online`;`fi.Sleepads[j].{Present,Reading}`;`fi.BedAreaIDs[j]`。

---

## 落点 ③(先做,可独立测——只骑 g^xy + 静态 κ,单床 cd2b 不依赖 ①)

### ③.1 `belief/coupling.go`:加满幅参 + 新方法

- `couplingParams` 加 `lLeftOpen float64`(LeftBed→SOpenFloor 满幅似然)。`defaultCouplingParams()` 给 `lLeftOpen ≥ emission lArea`(标定见下,先锚一个 ≥ 床 boost 量级的值,使 SOpenFloor 单位量级 ≫ SBed)。
- 新方法:

```go
// LeftBedOpenLogS §11③：sleepad LeftBed → 按 a_j 归属抬 SOpenFloor(present)/SBlindRest(lost)。
// 满幅 lLeftOpen 不被 Con 二次折扣(a_j 已分概率)；g^xy 几何门控多住户假摔(在别床 track a_j≈0)。
func (c *Coupling) LeftBedOpenLogS(o Observation, gxy []float64, radarOnline bool) [numStates]float64 {
	var out [numStates]float64
	a, _ := c.attachment(gxy)
	target := SOpenFloor
	if !radarOnline {
		target = SBlindRest // lost：人离床但雷达丢轨 → 歇态进盲(open-then-lost 由 T 行 SOpenFloor→SBlindOpen 自然流)
	}
	for j := range c.kappa {
		if j < len(o.Sleepad) && o.Sleepad[j] == BedLeftBed {
			out[target] += a[j] * math.Log(c.p.lLeftOpen)
		}
	}
	return out
}
```

### ③.2 `engine/engine.go`:两分支都注 logPhi

present 分支(`:178`)把 `Gxy` 提成变量复用;lost 分支用 track 冻结 XY 的 `Gxy`。两分支算完 `logPhi` 后:

```go
gxy := adapter.Gxy(ts.Obs.RadarTrack, fi.Beds, r.p) // lost 用冻结末位
lb := r.cp.LeftBedOpenLogS(obs, gxy, ts.Present)
for s, v := range lb {
	if v != 0 {
		r.js.AddLogToS(logPhi, belief.State(s), v)
	}
}
```

- present 分支:`logPsi` 那行的 `adapter.Gxy(...)` 改用上面 hoist 的 `gxy`(同一个,别算两次)。
- lost 分支:`ts.Obs.RadarTrack` 保留冻结 XY → `Gxy` 可分性按末位算(`obs` 在 lost 被剥了 XY,故 gxy 从 `ts.Obs.RadarTrack` 取,不是从 `obs`)。

---

## 落点 ①(后做,κ 动态;单床 κ 冷启≈1 本就够,① 主要给多床消歧)

### ①.1 `engine/engine.go`:每帧按"持续状态一致性"更新 κ(无 Room 状态)

**模型订正(C 2026-06-20)**:κ = sleepad_j 与"雷达床 j 占用"的**持续一致性**,**非跳变事件**。跳变只是一致性开始的瞬间;熟睡=两者持续都说"占用"=持续 `matched=true`,κ 被**持续建/保持高位**(不冻结不衰减)。这是 per-frame 的**连续条件**——**不是**"把一个离散事件数 15 遍"(那是重数事件),而是测"当前是否一致",每帧合法更新。故旧 deferred 窗/跳变检测/`bedKappaLedger` **全删**,**Room 不加任何状态**。

> 这也彻底解决早先"衰减 gate"之争:不靠 event-gate `live` 防熟睡衰减,而是让 `matched` 测持续一致——熟睡天然 `matched=true`,`live` 用 presence-gate 也安全。

每帧在 per-track loop **之前**(census.Update 之后)**无条件**调 `r.updateKappaFromState(fi)`。**🔴 反-orphan**:末尾**必调** `r.cp.UpdateKappa(matched, live)`(全 false 也调)——别把"修一个 orphan(UpdateKappa)又造一个(结算器没人调)"。落地后 grep 调用点 + probe 看 κ 真在动(见验证)。

```go
func (r *Room) updateKappaFromState(fi adapter.FrameInput) {
	matched := make([]bool, r.nb)
	live := make([]bool, r.nb)
	radarOnline := adapter.Online(fi)
	for j := 0; j < r.nb; j++ {
		sleepadOnline := j < len(fi.Sleepads) && fi.Sleepads[j].Present
		sleepadInBed := sleepadOnline && fi.Sleepads[j].Reading == belief.BedInBed
		radarInBedJ := false
		if j < len(fi.BedAreaIDs) && fi.BedAreaIDs[j] != 0 {
			for _, t := range fi.Tracks {
				if t.Online && t.FwAreaID == fi.BedAreaIDs[j] { // 同 M×N 的 RadarBedHitMask 信号
					radarInBedJ = true
					break
				}
			}
		}
		// live=双在线 ∧ 至少一方说"占用"(有待关联的人;both-vacant 空床无信息→冻结,不从互相"都说空"建同人相关)
		live[j] = radarOnline && sleepadOnline && (sleepadInBed || radarInBedJ)
		// matched=两方都说"占用"=同一人(持续在床→持续建 κ);一方占用另一方否=矛盾(live∧¬matched→衰减,多床消歧)
		matched[j] = sleepadInBed && radarInBedJ
	}
	r.cp.UpdateKappa(matched, live)
}
```

**逐情形(验收口径)**:

| 情形 | live | matched | κ |
|---|---|---|---|
| 熟睡(sleepad InBed + radar 在床 j,持续) | T | T | **持续建高/保持**(C 要的"持续证据抬 κ 对抗衰减")✓ |
| 空床(both vacant) | F | — | **冻结**(空床不建同人相关) |
| 矛盾(sleepad InBed 但 radar 明确不在床 j) | T | F | **衰减**(多床消歧;真矛盾才降,非"无事件") |
| 一方离线 | F | — | 冻结 |
| 雷达熟睡闪断 1 帧 | T | 瞬时 F | 慢 γ 的 EMA 几乎不动(旧窗版会误开窗误衰减,此版天然抗抖) |

**γ 标定**:`couplingParams.gamma` 现 0.2 是为**稀疏事件**设的;改 **per-frame** 后 0.2 太快(~5s 饱和、闪断也敏感),应降到时间常数 ~20-100s(γ≈0.01-0.05)。`defaultCouplingParams` 的 gamma 随之调,留 oracle form-anchor。

### ①.2 `belief/probe.go`:κ 进 FrameProbe(验证用)

`FrameProbe` 加 `Kappa []float64`;`Snapshot` 填 `cp.Kappa(j)` 各床(已有 `cp *Coupling` 参)。用于验证 κ 升/降/冻结。

---

## 细节约束

- **守 DBN §4 不变量**:① 只动权重,**绝不**在 κ 路径上碰 SBed 维持/衰减。
- **③ 满幅与归属分离**:`lLeftOpen` 是量级(满幅),`a_j` 是归属——别把 `a_j` 再乘进 magnitude(那是 Con 二次折扣,弃)。
- **lost gxy 取冻结末位**:从 `ts.Obs.RadarTrack`(保留 XY)算,不从被剥 XY 的 lost `obs`。
- **① 是持续状态一致性,非事件**:`matched`=两方持续都说"占用",熟睡持续 matched=true→持续建 κ(不衰减);`live`=双在线∧至少一方占用。慢 γ 抗抖,无需 deferred 窗。单床 κ 冷启≈1,① 主要给多床消歧。
- **radar-InBed→SBed 是 M×N 的事(c7e8ebe),κ-free,§4 承重第一腿,别动**:① 的 `radarInBedJ` 与 M×N 的 `RadarBedHitMask` 用**同一信号**(`FwAreaID==bedAreaID`)——κ 相关性与 SBed 抬升同源。回归须验:radar 在床仍 κ-free 抬 SBed(不因接 ① 而把 SBed 抬升误绑到 κ)。
- 守 CLAUDE.md:删即删不留兼容;不写 WHAT 注释;`go build ./... && go vet ./...` 全绿。

## 顺序

**①③ 互不依赖**(③ 骑 `g^xy`+静态 κ 即可,单床 cd2b 不需 ①)。建议:

1. **先 ③**(cd2b 主路径,可独立 cd2b 重测验证)。
2. **后 ①**(κ 动态,多床消歧泛化;单床无感)。

## 验证

- `go build ./... && go vet ./...` 全绿。
- **③ / cd2b**:用 M×N 重测的同步 layout fixture 回放 cd2b——LeftBed 后状态应从 SBed 翻 SOpenFloor/SBlindRest → SFallen 可达 → fire。**这条是 ③ 成败判据**(若床矩形内的摔仍被 M×N 每帧 κ-free 重抬 SBed 顶住=量级不够,调 `lLeftOpen`/或 Ψ overlap 压制,不改 floor)。
- **③ 回归**:非 LeftBed case 不受影响(无 LeftBed → `LeftBedOpenLogS` 全 0);现有 cd2b 0.5203 精确零回归基线(M×N 前)别破。
- **① κ probe**:多床 fixture 看 `FrameProbe.Kappa`——**持续在床(sleepad InBed + radar 在床)κ 建高并保持**(不衰减,这是 C 6-20 catch 的验收点)、空床 κ 冻结、矛盾(sleepad InBed 但 radar 别床持续)κ 降。
- **反-orphan**:`grep updateKappaFromState engine.go` 有调用点;probe 的 κ 在持续在床段确实**上升**(死函数 = κ 永远等于冷启值不动)。
- **标定项**(留 oracle / form-anchor,铁律 [[fall_data_is_artificial_test]] 无真实多床/真摔数据):`lLeftOpen` 量级、`kappaWindowMs`=15s、`γ`。先锚方向/符号,曲线留实测。

## ⚠️ 注

- sleepad-only 房:无真 radar track,合成 bed-track(`main.go`)带 pose=Lying+firmware 床;LeftBed→不合成→缺席→blind。③ 对合成 track 同样按 `o.Sleepad` 走(LeftBed 时合成 track 已缺席,无 S 轴载体,交 D 窗)。
- ④ latch / ⑤ floor:**不做**(B 轴 kObs 0.99 自持已是牙齿;floor.go 是 belief-独立天花板)。详 DBN §11。
