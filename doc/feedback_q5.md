# Q5 反馈日志(项目组侧)— belief 全空间区域占用重写 + cd2b 床边真摔 FN

> **QA 分离(2026-06-12,因 git 同步写冲突)**:P5 主题拆成两文件各写各的——
> **本文件 `feedback_q5.md` = 项目组侧**(提案/问题/根因/复算请求);委员会裁决/审查/工单写 `feedback_a5.md`。
> 两侧不写同一文件 → push 不再非-fast-forward 撞车。倒序,最新在上。
>
> **协作协议**:项目组提案 → 委员会裁 → 裁后建;裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 0 FAIL。

---

## 工单2 + 工单3-前半段 落码完成(2026-06-13,先提交待委员会审)

放行 bar 达成:`build ./...` ✓ / `vet ./internal/roomengine/...` ✓ / **belief 包 test ✓ + roomengine 0 FAIL**。

**已落(11 文件,全在 wisefido-sensor/internal/roomengine)**:
- `state.go` 新 9 态(`SEmpty/SBed/SSit/SOpenFloor/SBath/SFallen/SBlindRest/SBlindOpen/SLeft`,Fallen 仍 index 5);`model.go` 空间转移 A(SBed→Fallen 0.05、Blind→Fallen 0.5 种子、BlindRest 无 Left 逃生阀)+ Prior;`likelihood.go` 发射重映射(ObsNoDetect 压可见区+抬盲、poseLikelihood 卫浴分支、ExitRoom 压盲、删 ObsTrackPresent 整条);`calibration.go` 清死常量+新常量;`observation.go` 删 ObsTrackPresent;`belief_adapter.go` 删 ghost 发射。
- 测试:删 `belief_test.go`/`r5_calibration_lock_test.go`(旧 9 态 oracle 全废)→ 重建 `belief_test.go` 骨架 5 测(状态空间形状 / 床躺→Bed / 床离床→Fallen=cd2b belief 层出口 / 丢轨→blind ramp / ExitRoom 反证);修 `belief_adapter_test.go`/`belief_b_replay_test.go` 引用。

**落码中测试逮到并修的真问题**:① ExitRoom 似然没压盲区(补 `lrExitBlind`,且与"BlindRest 在 A 无 Left 逃生阀、离场靠观测拉"自洽);② Blind→Fallen 初设 12 是 drift 引擎(纯 Predict 凭空造跌倒)→ 改 0.5 种子,守住"★→Fallen 极小、靠观测驱动"原则。

**deferred 到工单3 后半段(oracle 重基线,已 t.Skip 带 reason 引工单3)**:3 个 FP 回归在新拓扑下需重标——
- `TestMoMLostTrackVanishNoFire`:喂纯 nil 是旧模型假设;新模型 MoM='走出 exit'应喂 ExitRoom→Left,**test 语义待更新**。
- `TestP4OpenFloorDwellToleranceGate`:dwell scale/tolerance 按旧拓扑标,新拓扑竞争 Fallen 的态变了,**dwell 常量待重标**(高/低 tol 当前不分离)。
- `TestRecallManifestAll`(case-5 hunzi FP peak 0.994):真数据 TP/FP recall/precision oracle **待按新拓扑重基线**。
- ⚠️ 暴露的标定张力:Blind→Fallen 种子大小 vs 纯 Predict 漂 / dwell tolerance 在新拓扑的分离力,是工单3 后半段重标的核心,且会牵动 lost-fall TP——须用真 fixture 标,别压死真摔。

**工单状态**:工单2 ✅(belief 重写)/ 工单3 前半段 ✅(骨架)/ 工单3 后半段 ⏳(oracle 重基线,上列 3 项)/ 工单1(bed 融合)、工单4(cd2b fire)未动。

---

## 当前状态(2026-06-12)

**定论(用户/架构师)**:DBN 自第一天(`b3a45ca`,2026-06-01)状态空间建错。重写方向 = 把 Room 层从「姿态 9 态」(位置不进状态,只作观测条件)换成「**全空间区域占用 × {直立, 倒地}**,含盲区一等态」;转移=空间运动;观测=区域有无 track;fall 从「占用 + 不可见或静止 + 未离场」涌现。`belief_shadow.go` 的 lost-fall sweep(MovingPrecondition 译反)删除变涌现。

**验证方法论**:用进程内 `belief_replay` 单测(干净可复现),不再 live 重放(bedSessions 跨 run 残留)。fixture = `case-cd2b-0604-16141631`,完成判据 = 该 fixture 必须 fire。

---

### 架构定性(项目组,待委员会印证)

两层各建不同的东西,正交:

- **Track 层**(`belief/track.go`,**保留**)= 观测质量滤波器,答「这条 blip 是真人还是反射」。状态 `T={None,Real,Ghost,JustLeft,Lost}`,5 态 HMM。数学核心 = co-existence 耦合 ρ(孤立 track 不可能 ghost——ghost=真人反射,必有共存 partner)。唯一对外输出 `P(Real)∈[0,1]`。
- **Room 层**(新空间模型,**重写**)= 状态估计器,答「人在哪个区、倒没倒」。状态 = Zone × {直立, 倒地}(或等效:空间 zone + `SFallen` 吸收态)。数学核心 = 空间转移阵 + 盲区一等态 + dwell 生存函数。
- **耦合(已落地)**:`belief_shadow.go:354` `o.Conf *= pReal`(`pReal=P(T=Real)`)。ghost→`P(Real)→0`→Conf→0→似然经 `temper` 退火成单位阵→不驱动空间推断。等价原始提案 §4.3 边缘化 `P(fall|y)=P(fall|real,y)·P(R=real|y)`。
- **工程经济**:保持 `numStates=9` → `Vector`/`normalize`/`Max`/`temper`/`Decider`/`belief.go` 全不动,重写收敛进 `state.go`(语义)+`model.go`(A)+`likelihood.go`(L)三文件(`SFallen` 仍须是 9 态之一,Decider 引用)。

**项目组提出的三处需拧紧(委员会待裁)**:
1. cd2b 的 fire **不来自 Track 层判 ghost**——co-existence 对孤立被子 ρ=0 判它 Real,反而保护它。fire 杠杆在 Room 层(sleepad 权威释放床占用 + 真人 absence + 未离场 → blind dwell ramp Fallen)。
2. absence 须定义在 realness 门控后的 track 上(「任何可见区无 `P(Real)>阈` 的 track」=有效缺失),而非「无 track」。
3. FP 安全性反转:旧 `A` 靠「→Fallen 极小,沉默不造跌倒」;新模型 blind-zone 久缺**必须**从沉默 ramp Fallen → 守门挪到 blind 进出纪律 + dwell 时序 + recapture/Exit cancel,且守[偏分监控铁律](只有极近两事件有向 hand-off 能排除 lost-fall)。

**盲区无几何 — 对上面架构的实质修正(2026-06-12,用户提出)**:layout(`room_visual_layout`/`radar.areas`)只编码雷达**覆盖到**的 area 对象;盲区=其补集,**任何数据源都不画它**;firmware 只报 track,不报「看不见」。所以**盲区不能是有几何、可定位的状态**。
- 重述:**Blind = 「有真人在场(近期证据)+ 没有任何 real track 解析出他 + 没有 Exit 事件」**——用否定定义,可观测性来自「track 从有到无 + 无 ExitRoom」,不来自几何。
- Near/Far/Bath 子态**不能**来自盲区几何(未知),只能来自**消失前最后一帧位置**(last-seen area_type + 门距,已知)→ 改成**入口条件化**:`Blind-from-Door`(高 exit 先验)/`Blind-from-OpenFloor`(高 fall 先验)/`Blind-from-Bath`。状态索引是「怎么进的盲」(可观测),不是「现在在盲哪」(不可观测)。
- cd2b 把此逼到极致:被子 track 把 Bed→Blind 转移**遮住**(连「track 没了」都观测不到)→ 该转移只能靠 **sleepad LeftBed 与 radar 床 track 矛盾**推出,对上原始提案 §7 诚实下限。

---

### cd2b 床边真摔 replay FN — 根因实证(供 CodeX 独立复算)

**fixture**:`doc/cases/case-cd2b-0604-16141631`(201 卧室,radar `cd2b` + sleepad `1641`,UTC 22:14–22:31)。
**oracle**:22:23:20 sleepad **LeftBed**(人跌床边、雷达不可见);之后 radar 持续在床报 lying(被子=静止反射伪装成在床的人)。`alarm.json=[]`(生产漏报)。**正确结果 = ~22:25 应 fire 一条 Fall**(被子是反射,不该按「在床睡」豁免)。
**replay 结果**(in-process belief replay,cd2b fixture):`belief_shadow_fall` 空 → `SFallen` 从未 confirm → **DBN 也 FN**。

**FN 代码链(file:line 可核)**:

1. **孤立被子 track 不被判 ghost,被当人。** `belief_shadow.go:349-354`:fall 证据 `o.Conf *= pReal`,注释明写 `real lone faller pReal≈1(孤立 ρ=0 不可能 ghost)→ 无折扣`。即 co-existence 对孤立被子判 Real → Conf 不降 → 被子 pose=lying@bed 正常驱动 Room。**与 ghost 无关。**
2. **床区躺→倒地的唯一出口 = BedReleased 分支。** `likelihood.go:161` `case inBed && !c.BedReleased:` → **SBedLying**;`likelihood.go:165` `(inBed && c.BedReleased)` → **SFallen** 候选。由 `Observation.BedReleased` 二选一。
3. **BedReleased 生产闸门**:`belief_shadow.go:361` `if bedReleased { o.BedReleased = true }`;`belief_shadow.go:229` `bedReleased := bedSt.BedConfidence > 0 && bedSt.BedStatus != 0`(0=InBed);`bedSt := tm.BedOccupancyState(nowMs)`(:228)。
4. **闸门输入错**:回放期 `BedOccupancyState` 返回 `BedStatus=0(InBed)/conf90` → `bedReleased=(90>0 && 0!=0)=false` → BedReleased 永不置位 → 走 SBedLying → P(Fallen) 不 ramp → FN。

**根因定位**:**不在** belief 层(D3 止血 `belief_shadow.go:359-368` 接线正确),**在上游 bed 占用融合 `tm.BedOccupancyState`**:sleepad 22:23:20 的权威 LeftBed 没能盖过 radar 被子的 InBed 证据,bed 贝叶斯 scorer 仍输出 InBed/conf90。契约:sleepad 接触式权威 > radar(见 `bed_fusion_authority_model` / `bed_stale_leftbed_vetoes_radar_inbed`)。

**请 CodeX 独立复算/证伪**:
- **(a)** FN 根是否就是 `BedOccupancyState` 在 22:23:20–22:30 窗内仍返回 `BedStatus=0/conf90`?即 sleepad LeftBed 为何未释放 bed_state——radar 被子 InBed 证据权重 ≥ sleepad LeftBed?还是 LeftBed 根本没喂进 scorer?
- **(b)** 若把 bed 融合修对(sleepad LeftBed 胜出 → `BedStatus!=0` → `bedReleased=true`),现有 D3 止血是否**即可**让 cd2b 在 ~22:25 fire?还是 pose/dwell ramp 时序不够、仍需别的证据?
- **(c)** 计划中「全空间区域占用×姿态」重写是否真能消掉这个 band-aid——还是只是把同一个「sleepad vs radar 谁定床占用」的权威问题从 BedReleased 闸门挪到 blind-zone 进入条件上(即重写**不**自动解 (a),仍依赖 bed 融合权威修对)?

**项目组判断(待 CodeX 印证)**:(a) 成立;(c) 重写**不**自动解 (a)——区域模型让 blind-zone 成一等态、删 MovingPrecondition 译反,但「状态离开 Bed 区」仍取决于 sleepad LeftBed 能否压过 radar 被子的床占用。故 **bed 占用融合权威(sleepad>radar)是 cd2b fire 的必要前置,与状态空间重写正交,两者都要做。**

---

## 施工方案(项目组,待委员会裁)

核心一句话:Room 层从「姿态分类器」变成「单住户的空间占用估计器」,fall 不再靠 pose 抬,而是从「该在的人解析不到 + 没离场 + 拖时间」涌现;Track 层原样保留当真伪前置。保持 `numStates=9`,重写只落在 `state.go`/`model.go`/`likelihood.go` 三文件,`belief.go`/`track.go` 不动。

### 1. 状态空间(state.go,9 态,Fallen 仍占 index 5 → Decider 不改)

| idx | 新态 | 含义 | 来源 |
|-----|------|------|------|
| 0 | Empty | 房内确认无人 | 同旧 SEmpty |
| 1 | Bed | 占用在床(睡/休息) | area=Bed + sleepad InBed |
| 2 | Sit | 占用在休息区(椅/沙发)——久坐正常,低 dwell-fall | area=Sit |
| 3 | OpenFloor | 占用在开阔活动区(站/走,摔在这里) | area=Active/Deny |
| 4 | Bath | 占用在卫浴(高风险,常盲) | area=Toilet/Shower |
| 5 | Fallen | 倒地(唯一 fire 目标,近吸收) | 涌现 |
| 6 | BlindRest | 占用但解析不到,从 rest 区(床/卫浴)进盲 = 床边/卫浴真摔(cd2b!) | 否定定义 |
| 7 | BlindOpen | 占用但解析不到,从开阔区进盲 = 走动中消失 | 否定定义 |
| 8 | Left | 已经门离场(→收敛 Empty) | ExitRoom/门区 hand-off |

**关键设计决策(待委员会裁)**:

- **删 SArtifact**:ghost/反射的活全归 Track 层(`Conf*=P(Real)` 在似然层就把它退火成中性),Room 层不该再有伪迹态——兑现"Track=观测滤波器/Room=状态估计器"两层分离。腾出 1 个 slot。
- **删 STransition/SBedRestless**:不确定性由 Blind 态 + normalize 承载,不需要专门枢纽态;床上翻动并入 Bed。再腾 2 个 slot。
- **"门"不设独立态**:延续现在的做法(NearDoor 作观测条件 + Left 态 + ObsEnterExit 事件),省 1 个 slot 给 Blind 一分为二。
- **Blind 拆 Rest/Open 是整个重写的命门**:两者转移语义不同才值得各占一态——BlindRest→Fallen 几乎无逃生阀(rest 区不是门,人不会从床里凭空离场),BlindOpen→{Fallen | Left} 由 reachable-exit 裁。盲区本身无几何(layout/sensor 都不画),所以态的索引是"怎么进的盲"(last-seen 区,可观测),不是"在盲哪"(不可观测)。

### 2. 转移阵(model.go A,空间运动)

- **可见区互联**:Bed↔Sit↔OpenFloor↔Bath 经 OpenFloor/门枢纽连通(粗粒度、房无关)。
- **进盲**:每个可见区 → 对应 Blind 态小倾向(钻家具背侧/床边死角)。Bed/Bath→BlindRest、OpenFloor→BlindOpen。
- **出盲**:BlindOpen→OpenFloor(重现)/→Left(reachable-exit 高)/→Fallen(久缺);BlindRest→Bed/Bath(重现)/→Fallen(久缺,无 Left 逃生阀)。
- **Fallen** 近吸收(自持~0.99),只有正向 recapture/exit 证据能压下。Left→Empty。
- **安全性反转(必须正视)**:旧 A 的安全靠"→Fallen 极小,沉默不造跌倒"。新模型反过来——Blind 里久缺就该从沉默 ramp Fallen(这才是 lost-fall 的本义)。FP 守门从"A 惰性"挪到三处:① BlindOpen 的 reachable-exit 逃生阀;② 进盲纪律(Bed→BlindRest 必须先有"人离床"证据,不能凭 radar 抖动空降);③ ramp 由 dwell 生存函数定速(对齐床边真摔 6–8min 物理下限,不是秒级)。ramp 机制复用现有 ObsNoDetect(每盲 tick 固定强度 × Fallen 近吸收 → 自然累积),不手调 gain。

### 3. 发射(likelihood.go L)

- **可见 track**(Conf*=P(Real) 已门控)落 zone Z → 抬 Z;OpenFloor + pose=lying/fallen → Fallen;Bed + bedReleased(sleepad 离床) → Fallen(cd2b 出口,即现有 BedReleased 分支,现在喂新 Fallen)。
- **ObsNoDetect**:对 realness 门控后的 track 做房级归约——"任何可见区无 P(Real)>阈 的 track"=有效缺失 → 抬 Blind(按 last-seen 分 Rest/Open)+ 略抬 Left;重复 → ramp Fallen。
- **ObsBedOccupied**:sleepad InBed→Bed / LeftBed→压 Bed(释放)。← cd2b 命门在这。
- **ObsReachableExit/ObsNeighbor/ObsEnterExit/ObsVitalPresent** 路由 Left/Empty,大体沿用。
- **删 ObsTrackPresent→SArtifact 整条**(Track 层接管)。

### 4. unit 内部联动(对 Blind 态承重)

unit 联动不是锦上添花——没有它,跨房走动在"被离开的房"里就是一条 lost-fall 误报。三类机制:

**① 同房多雷达 union**(cd2b + 333b 这种):一个 unit 里覆盖同一房的多台雷达,它们的 track 全喂同一个 room belief。命门:ObsNoDetect 的 absence 归约是对本房全部 unit 雷达的 real track 取并集——"任何一台雷达解析到 P(Real)>阈"就不算盲。一台雷达补另一台的盲区,blind 直接缩小。

**② 跨房 hand-off**(人从 A 房挪到 B 房):ObsNeighbor 弱耦合消息:A 房住户进 BlindOpen 时,消费"邻房 B 此刻有没有 fresh 有向出现"。有 → BlindOpen→Left(人挪走了,不是摔);无 → 留在 Blind 继续 ramp。兑现[偏分监控铁律](只有极近两事件有向 hand-off 能排除 lost-fall),保留现有 neighborHandoff 不删。

**③ 房内 sleepad↔radar 权威**:sleepad LeftBed 压 radar 被子本身就是 unit 内跨模态设备联动,cd2b 命门。

**结构底座**:`owl-common/spatial/relmatrix.go`(mm_relationship_matrix,已建+测,sensor 当 maintainer)。它给出"哪些雷达覆盖本房/本床(radar→room/bed covers)、哪些房相邻"。Room belief 的输入 scope 由 MM 矩阵界定(哪些设备 union 进本房 absence / 哪些房可 hand-off)。

### 5. 文件改动

| 文件 | 动作 |
|------|------|
| `belief/state.go` | 重写 9 态枚举 + label |
| `belief/model.go` | 重写 A(空间)+ Prior |
| `belief/likelihood.go` | 重写发射映射(zone-based + noDetect→blind + bed 权威 + 删 artifact) |
| `belief/observation.go` | 可能加 zone 字段/ObsTrackInZone;删 ObsTrackPresent 路径 |
| `belief_adapter.go` | 从 area_type 出 zone;noDetect 走 realness 门控后的 absence;sleepad LeftBed 喂 bed 权威 |
| `belief_shadow.go` | 删 lost-fall sweep(:430+ 的 MovingPrecondition 译反);引擎 wiring 迁生产 tick(DBN_FIRE);realness filter 保留喂 Track 层 |
| `belief.go` / `track.go` | **不动** |
| 上游 bed scorer / BedOccupancyState | 修 sleepad>radar 权威(belief 包外,前置) |

### 6. 验证与风险(诚实)

- **完成判据**:`belief_replay` 单测,cd2b 必须在 ~22:25 fire。
- **最大风险 = 回归网清零**:重写状态语义会让 `r5_calibration_lock_test.go`/`belief_test.go`/`fall_reason_test.go` + 已知 FP 回归(case-5 hunzi、cabb 静止站立)全部失效——它们是按旧 9 态写的安全网。必须先重建 oracle 测试套(已知 TP 必 fire / 已知 FP 必不 fire),否则等于在没有回归网的情况下重写核心。
- **施工顺序锁定**:bed 融合权威先修 → cd2b fire 验证 → 以此为 oracle 锚点建测试套 → 重写 belief → 跑测试套验证不回归。
