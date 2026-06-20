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

**模型(C feedback_a5 定稿:双腿不可合并,根在 γ 矛盾)**:**空间分不开哪张床时靠时间共现破平局**——雷达只知 `pos+是床`,推不出 BedA/BedB(几何平局 `g^xy_A≈g^xy_B`);但 sleepad_A InBed 事件与 radar track 入床区**同刻**(±15s)→ 绑同人(靠时间不靠位置)。故 κ 两条腿,各管各 γ:

| κ 输入 | 管 | 节拍 | γ |
|---|---|---|---|
| **强化**:同向共跳(15s deferred 窗) | 建立关联(时间共现绑同人,破空间平局) | 事件触发(窗满一次) | 强 `gammaEvt=0.2` |
| **维持**:持续共态(sIn∧rIn) | 维持 κ 抗衰减(熟睡托住) | 每帧 | 弱 `gammaHold≈0.01` |

**一个 γ 做不到两者**(建立要大步快锁、维持要小步稳不抖)→ 两腿不可合并。两腿同更 `c.kappa` 叠加:事件快跳、持续慢托。

### ①.1 `belief/coupling.go`:γ 拆两档 + 维持方法

```go
type couplingParams struct {
	gammaEvt  float64 // 强化:同向共跳 EMA(强,稀疏触发)
	gammaHold float64 // 维持:持续共态 per-frame EMA(弱,抗衰减)
	cEmpty    float64
	epsArt    float64
}
func defaultCouplingParams() couplingParams {
	return couplingParams{gammaEvt: 0.2, gammaHold: 0.01, cEmpty: 0.2, epsArt: 1e-2}
}
func (c *Coupling) emaKappa(target, live []bool, gamma float64) {
	for j := range c.kappa {
		if j < len(live) && live[j] {
			m := 0.0
			if j < len(target) && target[j] {
				m = 1.0
			}
			c.kappa[j] = (1-gamma)*c.kappa[j] + gamma*m
		}
	}
}
func (c *Coupling) UpdateKappa(matched, live []bool)   { c.emaKappa(matched, live, c.p.gammaEvt) }  // 强化
func (c *Coupling) MaintainKappa(agree, live []bool)   { c.emaKappa(agree, live, c.p.gammaHold) }  // 维持
```

(原单 `gamma` 字段删,换 `gammaEvt`/`gammaHold`;`emaKappa` 共用。)

### ①.2 `engine/engine.go`:两腿——维持(per-frame)+ 强化(15s 窗)

`Room` 加 `kappaLedger []bedKappaLedger`;`NewRoom`: `kappaLedger: make([]bedKappaLedger, nb)`。类型 + 窗常量 + 共用 helper(**钉① lost 边界**:radarInBed 不 gate Online):

```go
type bedKappaLedger struct {
	sleepadPrev belief.BedReading
	radarPrev   bool
	pend        bool
	pendMs      int64
	pendSDir    int // +1→InBed / -1→LeftBed
	pendRDir    int // +1 enter / -1 leave
}
const kappaWindowMs = 15000 // 强化窗 K(case1/2 的 15s)

// radarInBed κ 的 radar 半:FwAreaID==床。**不 gate Online**(钉① lost 边界)——lost-coasting track
// 的 FwAreaID 是 M×N 冻结末值;radar 半冻结、sleepad 半实时把关:人真离床→sleepad LeftBed→matched/agree=F
// →κ 衰减,冻结 N 不会错误维持。evicted track 不在 fi.Tracks,自然不计。
func radarInBed(fi adapter.FrameInput, j int) bool {
	if j >= len(fi.BedAreaIDs) || fi.BedAreaIDs[j] == 0 {
		return false
	}
	for _, t := range fi.Tracks {
		if t.FwAreaID == fi.BedAreaIDs[j] {
			return true
		}
	}
	return false
}
```

每帧在 per-track loop **之前**(census.Update 之后)**无条件**调两腿:`r.maintainKappa(fi)` + `r.eventKappa(fi)`。**🔴 反-orphan**:两腿末尾各**必调** `MaintainKappa`/`UpdateKappa`(全 false 也调)。落地 grep 两调用点 + probe 验 κ 真在动。

**维持腿**(per-frame,弱 γ):

```go
func (r *Room) maintainKappa(fi adapter.FrameInput) {
	live := make([]bool, r.nb)
	agree := make([]bool, r.nb)
	radarOnline := adapter.Online(fi)
	for j := 0; j < r.nb; j++ {
		sOn := j < len(fi.Sleepads) && fi.Sleepads[j].Present
		sIn := sOn && fi.Sleepads[j].Reading == belief.BedInBed
		rIn := radarInBed(fi, j) // 冻结末值(钉① lost 边界)
		live[j] = radarOnline && sOn && (sIn || rIn) // both-vacant→冻结
		agree[j] = sIn && rIn                        // 持续共占→维持抬;真离床 sIn=F→agree=F→弱衰减
	}
	r.cp.MaintainKappa(agree, live)
}
```

**强化腿**(15s deferred 窗,强 γ,**含钉② 矛盾判据修**):

```go
func (r *Room) eventKappa(fi adapter.FrameInput) {
	matched := make([]bool, r.nb)
	live := make([]bool, r.nb)
	for j := 0; j < r.nb; j++ {
		L := &r.kappaLedger[j]
		cur, sOn := belief.BedNoReport, false
		if j < len(fi.Sleepads) {
			cur, sOn = fi.Sleepads[j].Reading, fi.Sleepads[j].Present
		}
		sDir := 0
		if cur != L.sleepadPrev {
			if cur == belief.BedInBed {
				sDir = +1
			} else if cur == belief.BedLeftBed {
				sDir = -1
			}
		}
		L.sleepadPrev = cur
		rNow := radarInBed(fi, j)
		rDir := 0
		if rNow != L.radarPrev {
			if rNow {
				rDir = +1
			} else {
				rDir = -1
			}
		}
		L.radarPrev = rNow
		if sDir != 0 || rDir != 0 { // 开窗 / 记方向
			if !L.pend {
				L.pend, L.pendMs = true, fi.NowMs
			}
			if sDir != 0 {
				L.pendSDir = sDir
			}
			if rDir != 0 {
				L.pendRDir = rDir
			}
		}
		if L.pend && fi.NowMs-L.pendMs >= kappaWindowMs { // 窗满解析,一床一次
			coTrans := L.pendSDir != 0 && L.pendRDir != 0 && L.pendSDir == L.pendRDir // 同向共跳=强证据同人→强抬
			sInNow := sOn && cur == belief.BedInBed
			contradiction := sInNow && !rNow // 钉②:仅"sleepad 在床 ∧ radar 持续别床"=真矛盾→强降
			matched[j] = coTrans
			live[j] = coTrans || contradiction // 钉②:"sleepad 跳入 + radar 持续在床(agree)"→两者皆 F→不强更新,交维持腿托
			*L = bedKappaLedger{sleepadPrev: cur, radarPrev: rNow}
		}
	}
	r.cp.UpdateKappa(matched, live)
}
```

**逐情形(两腿合成)**:

| 情形 | 强化腿(事件窗) | 维持腿(per-frame) | κ 净 |
|---|---|---|---|
| 同步上床(15s 内双方进床) | coTrans=T **强抬** | agree=T 弱抬 | **快建立** ✓(时间绑同人,破空间平局) |
| 熟睡持续在床 | 无跳变 no-op | agree=T 弱抬 | **持续保持高** ✓ |
| **sleepad 跳入 + radar 持续在床**(钉②) | coTrans=F、contradiction=F→**live=F no-op** | agree=T 弱抬 | **升不降** ✓(不误判矛盾) |
| 矛盾(sleepad 在床 + radar 持续别床) | contradiction=T→**强降** | agree=F 弱降 | 降(多床消歧) |
| 空床(both vacant) | 无事件 | live=F 冻结 | 冻结 |
| lost 后人真离床 | sleepad LeftBed→sInNow=F | agree=F(sIn=F,冻结 rIn 不救)→弱降 | **降** ✓(钉①,冻结 N 不误维持) |
| 一方离线 | live=F | live=F | 冻结 |

**γ 标定**:`gammaEvt=0.2`(强化稀疏事件大步);`gammaHold≈0.01`(维持 per-frame 小步 τ≈100s)。留 oracle form-anchor。

### ①.3 `belief/probe.go`:κ 进 FrameProbe(验证用)

`FrameProbe` 加 `Kappa []float64`;`Snapshot` 填 `cp.Kappa(j)` 各床(已有 `cp *Coupling` 参)。验证两腿(replay 验收项见下)。

---

## 细节约束

- **守 DBN §4 不变量**:① 只动权重,**绝不**在 κ 路径上碰 SBed 维持/衰减。
- **③ 满幅与归属分离**:`lLeftOpen` 是量级(满幅),`a_j` 是归属——别把 `a_j` 再乘进 magnitude(那是 Con 二次折扣,弃)。
- **lost gxy 取冻结末位**:从 `ts.Obs.RadarTrack`(保留 XY)算,不从被剥 XY 的 lost `obs`。
- **① κ 双腿不可合并(根=γ 矛盾)**:强化腿(同向共跳 15s 窗,强 γ_evt,时间共现绑同人破空间平局)+ 维持腿(持续共态 per-frame,弱 γ_hold,抗衰减熟睡托)。钉①=radarInBed 不 gate Online(lost 冻结、sleepad 实时把关);钉②=强化腿只在"sleepad 在床∧radar 持续别床"才强降,"sleepad 跳入+radar 持续在床"交维持腿托不强降。单床 κ 冷启≈1,① 主要给多床消歧。
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
- **① κ probe**(多床 fixture,两腿验收):同步上床 κ **强抬建立**(强化腿)、熟睡持续在床 **保持高不掉**(维持腿)、**sleepad 跳入+radar 持续在床 κ 升不降**(钉②)、矛盾(sleepad 在床+radar 持续别床)**降**、**lost 后人离床 κ 降**(钉①,冻结 N 不误维持)、空床 **冻结**。
- **反-orphan**:`grep -E 'maintainKappa|eventKappa' engine.go` 两调用点都在;probe 的 κ co-entry **跳升**、熟睡段**不掉**(死函数 = κ 永等冷启值不动)。
- **标定项**(留 oracle / form-anchor,铁律 [[fall_data_is_artificial_test]] 无真实多床/真摔数据):`lLeftOpen` 量级、`gammaEvt=0.2`/`gammaHold≈0.01`、`kappaWindowMs=15s`。先锚方向/符号,曲线留实测。

## ⚠️ 注

- sleepad-only 房:无真 radar track,合成 bed-track(`main.go`)带 pose=Lying+firmware 床;LeftBed→不合成→缺席→blind。③ 对合成 track 同样按 `o.Sleepad` 走(LeftBed 时合成 track 已缺席,无 S 轴载体,交 D 窗)。
- ④ latch / ⑤ floor:**不做**(B 轴 kObs 0.99 自持已是牙齿;floor.go 是 belief-独立天花板)。详 DBN §11。
