# P6A 反馈日志（项目组 A 侧）— Xsensorv1 联合占用滤波 + cd2b 床边真摔 FN

> **QA 三方分离（2026-06-14，因 git 同步写冲突）**：P6 评审采用竞争式三卷，各写各的文件——
> - **本文件 `feedback-p6A.md` = 项目组 A 侧**：方案陈述 / 根因 / 前置实验 / 对评审的回应。
> - `feedback-p6B.md` = 评审组 B 侧（B 的独立评审）。
> - `feedback-p6C.md` = 评审组 C 侧（C 的独立评审，与 B 竞争）。
>
> 三方不写同一文件 → push 不再非-fast-forward 撞车。倒序，最新在上。
> **共同基线**：`doc/DBN-Zone-Room.md`（联合占用模型 ground truth）、`doc/DBN-cd2b.md`（cd2b case）、`CLAUDE.md`。

---

## 2026-06-16（其十五）— silent-fall FN-safe 契约：default-fall + 正向压制 + 总时长兜底（给 C 审 FN 守门）

**纠其十四的二次错误**：其十四把 still-box 摆回观测层（对），但顺手提的"排除法涌现"（Fallen=残差，排干净才出）**把 FN-safe 默认方向反了**——二义时变成默认不报=低召回，违背"保不漏"长期路线图（漏报≫误报）。架构师拍回：**除非有正向数据，否则偏 Fallen。默认有罪（偏摔），正向数据无罪释放**——不是默认无罪、要凑齐证据才定罪。

### 原则（FN-safe 默认）
still-box 久静 → **默认偏 Fallen**（疑似摔）；由**正向证据**压开：z=坐/站高、接触 InBed、area 印证坐/床/卫浴。**无正向证据 → fall 倾向保留。** 老的 still→SFallen 直投**方向本就对**（久静默认疑似摔），错的是短阈 + 没有效正向压制，**不是"偏 Fallen"本身**。

### 四层归属（reason from model.go，正交分量各回各层）
| 分量 | 层 | 文件 | 作用 |
|---|---|---|---|
| still-box（每帧二值） | ③ emission | emission.go | **偏 Fallen**（精度主路径） |
| area_type（每帧读活的 cell） | ③ emission | emission.go | **正向压制/redirect**（bed/sit/toilet→抬 Bed/Sit/Bath 压 Fallen） |
| Z（每帧） | ③ emission | emission.go | 单向正抬、低端中性（z<30 含0=中性） |
| 接触（sleepad） | ③ coupling Ψ | coupling.go | LeftBed→B vac 抽空 SBed（cd2b 样板，不动） |
| 时长**总量** | **floor 兜底** | （新增） | 总时长≥T_floor→强制 suspect（召回，绕过滤波误压） |
| risk-time / N_r | ④ decide | decide.go | C_FN（漏掉多糟，不碰 P(Fallen)） |

注：时长**总量**喂 floor（召回）；每帧二值喂 emission（精度），总时长在前向滤波里以"累积帧数"涌现、不被直接读。**per-area 是 floor 的阈（按 area 取），不是 emission 的项。**

### 三个 FN 守门（涌现/压制被误压的三条路）
1. **area 误学压 Fallen**：真摔落在误学成 Sit/Bed/Toilet 的 cell → area 抬错态压 Fallen。守：③ area 权重**有上限**（不能单凭 label 锁死，要能被 still + floor 翻过来）。
2. **z 假阳压 Fallen**：z 误报坐/站高给倒地的人。守：② **单向正抬、z<30（含0）中性** → 倒地 z≈0 占多数 → z 大多不参与，天然挡。
3. **接触假阳（陈旧 InBed）压 Fallen**：sleepad 没翻 LeftBed → Ψ 锁 SBed。守：床融合 LeftBed 否决 + 指数衰减（[[bed_stale_leftbed_vetoes_radar_inbed]]）。

### 一个兜底（last-resort floor）
```
still-box 总时长 ≥ T_floor(按 area) 且 无正向休息证据 → fall-suspect
豁免: z=坐/站高(真直立) ∨ 接触 InBed(真在床) ∨ AreaDeny(15天高bar 静态反射) → 不 floor-fire
```
- 接住 emission 被 area/z/接触误压的真摔。**豁免挂可观测证据，不挂 label**（否则 area 误学的 FN 从兜底又漏）。
- 与被否的 60s 区别：60s 是**主路径+短阈**（海量 FP）；floor 是**兜底+保守长阈**（正常活动不触，只接被误压真摔）。
- 诚实 tradeoff：兜底豁免读 z/接触（防 constipation FP），而 z/接触正是守门2/3 会失效的——故兜底不能 100% 堵"z假阳+接触假阳同时发生"的真摔，但需两 guard 同时失效，概率低，署名接受残余有界 FN。

### 三个回归闸（改动不能破已验证路）
1. **cd2b**（EG1 0.9992）：接触 LeftBed→Ψ 抽空 SBed→Fallen。机制保（Ψ 不动），但 **area 项 + still 项改动会动 S 轴竞争 → P(Fallen) 数值必重验**，最可能被这次动。
2. **d5f7-0524**（silent，present-track）：偏 Fallen 无正向证据 → 报。
3. **lost-fall**（blind 路）：消失→logPhi=nil→仅 Predict→Blind→Fallen 0.5 种子+0.99 自持。**全程无 emission/still-box → 本次改动碰不到 → 应天然完好**，跑一个 lost case 证。

### 范围 / 分轴
- 本轮：emission（still 保偏 Fallen + area 每帧压制）+ floor（总时长兜底）**同批不分家**（分家=有精度无召回）。
- **d523 静态伪迹 → realness/AreaDeny 单独立项**，不让 emission/floor 背伪迹的锅。
- transition / decide / coupling / blind 路 **不动**。

### 给 C
请审 FN 守门是否够（尤其守门1 area 误学 + 兜底的可观测证据豁免是否堵住 area-mislearn 的 FN）。其十三/其十四的"per-zone ramp"和"排除涌现"**均已撤回作废**，以本契约为准。

---

## 2026-06-16（其十四）— 【撤回其十三】still-box 是观测、被摆错位置；框架订正

**结论：其十三的移植（survival.go per-zone ramp）框架错，已撤回（删 survival.go + 还原 emission/observation/adapter/main/bootstrap 到 b4b04ce 前）。** 架构师停 replay 拍定。

**错在哪**：把 still-box 这条**观测**直接接成了**状态裁决**（`still-box → fallLRFromDwell → 抬 SFallen`）。
- "dwell" 是伪概念：`base.StillSec = base.StillBoxSec`（track_manager.go:841），它就是 still-box（30s 滚动 50×50cm 方框抗质心抖动算出的"目标连续没移出秒数"），源自 firmware 位置读数。production survival.go 把它包装成「P4 dwell HSMM」，我移植时连这套伪命名 + gate 规则一起搬了——违反 #1.1（一个 canonical 名）。
- **still-box 只观测一件事：静止 vs 移动**。静止是二义的（睡床/坐马桶/摔地都静止），它**分不出**是哪种。让"没动"直接给"摔"投票 = observation 摆到了 decision 层。

**replay 实证错框架（已停）**：d523 静物伪迹（站立 z=0 静止）P^F=0.945 vs d5f7 真摔（坐姿 z=0 静止）P^F=0.979——**几乎一样**。因为 still-box 对这俩物理上无法区分，tuning scale 改不了二义。中途为治 d5f7 FN 删了 dwell→SBed 又让伪迹 SFallen 爬到 0.945（差 90s T_hold 没 fire）——按下葫芦起了瓢，证明此路不通。

**订正框架（架构师拍定，A 认）**：
1. **still-box = 观测层**，只作"静止 vs 移动"证据：对等抬所有静止态（Bed/Sit/Fallen/BlindRest）、压移动态（OpenFloor-活动），**断开 → SFallen 直连**。
2. **"久静多久算异常" = 时长先验**，进 **transition/duration 层**（按 area），不进 emission。survival.go 的 per-area scale（toilet20/sit90）若保留，归这层，不归发射。
3. **摔 = belief 排除法涌现**：无接触（排 Bed）+ 低 z + 非坐区（排 Sit）→ 残下 Fallen。不是 still-box 喊出来的。
4. **静态伪迹**（站立-z=0 固定位置）交 **realness/AreaDeny**（d523 取证 is_refl=False·p_real=1.0 = realness 现没抓到它 = d523 真 gap），不归 still-box。

**下一步**：按订正框架重画 still-box 进 belief 的接法（emission 改"静止占用对等证据"+ duration 层放时长阈），design 先行再码，不再 replay 直到框架对。**roadmap §127「FallRulesParam.Still DBN 永久消费」仍是误标**（per-area 阈从没接进 DBN），订正照旧。

> 以下其十三原文（错框架，留档备查，已撤回不再有效）：

### 〔已撤回〕原其十三 — silent-fall per-zone 久静阈移植（survival.go ramp 替 stillTau=60）

**根因（C 全程参与核实，揪出 roadmap §127 一行误标 + 修正 A 一个事实错误）**：

生产误报 `case-d523-0611-22262238` 在 Xsensorv1 复现误火。逐层核出根因 = **belief emission 的久静判摔退回 `stillTau=60` 全局硬阈**：
- 伪迹帧实据：`pose=4(Standing) z=0 无 sleepad`，静止 ~2.7min。pose≠6 无 lying boost、z=0 无 lZ 直立信用 → **dwell 单独把 SFallen 抬过阈**。
- 老架构（Tsensor/wisefido-sensor）**早治过同类**：`belief/survival.go`（commit `272a48e` 06-09 生、`ea459c9` 06-11 per-zone 尾表、06-13 三连补丁调优）用 **Weibull 平滑生存 ramp** `fallLR=1+(d/scale)²` 替「still≥硬阈即报」悬崖，按 cell areaType 分尾尺度（**toilet/shower 20min（constipation-safe）· learned sit 90min · 默认 12min · bed/deny 不报**）。注释实锤：硬编码 60s 让 DBN 在正常久坐误火（101/Kitchen/Hunzi/Ton 海量 FP）。
- **Xsensorv1 belief 06-14 从零重写时没移植 survival.go**，emission 临时塞 `stillTau=60` 占位 = 在老架构杀掉该 bug 的次日又重造一遍。

**roadmap §127 误标（A 卷，A 自订正）**：§127 把 `FallRulesParam.Still`（per-area 久静阈）标成「DBN realness 消费，永久留对」——**实际断的**。grep 实证：belief/realness/census/adapter 无一消费 `EffectiveStillTimeoutSec`，馈送层 track_manager 算了但 fire 已删。per-area 阈从没接进 DBN。

**C 修正 A 一个事实错误（已核实认账）**：A 一度说"seam 不传 areaType、要打通 seam"——**错**。`TrackStatusBase.CellAreaType`（track_manager.go:789）字段在、馈送层填值（:848 `base.CellAreaType = c.Belief[0].Type`）、经 `OnRoomFrame` seam 透传。真正断点 = `dbnRouter.onRoomFrame` 构造 TrackObs 时**丢了 CellAreaType** + belief 无 ramp。**所以 seam 零改动**，修复边界缩到 adapter+emission。

**实施（6 文件，seam 不动，commit 见下）**：
1. `belief/survival.go`（**新建，移植 Tsensor 蓝本**）：`dwellTailFor(roomType,areaType)` + `fallLRFromDwell(...)` + dwell 常量。areaType 数值与 Xsensorv1 `AreaType` 枚举一致（Bed=2/Sit=3/Deny=5/Shower=6/Toilet=7），roomType=card.RoomType（Bathroom=1）。
2. `belief/emission.go`：dwell 块拆两半——**SFallen 走 per-zone ramp**（替 `StillSec≥stillTau` 那条对 SFallen 的 boost）；**保留 SBed/SOpenFloor/SSit 的占用塑形**（still≥τ 抬 SBed·罚活动态），减小回归面（fall 单走 ramp，占用形态不变）。
3. `belief/observation.go` + `adapter`：`Observation`/`RadarTrack` 加 `AreaType/RoomType`，`BuildObservation` 透传（零签名改动）。
4. `cmd/xsensor/main.go` + `bootstrap.go`：TrackObs 填 `int(b.CellAreaType)` + `g.roomType`；合成 bed-track 设 AreaType=2（bed 不报）；roomGeom 加 roomType。

**本轮范围 = ① per-zone scale 载重最小集**。tolerance 反向（`toleranceMult`，破近吸收棘轮）+ 夜间短尾 + 雷达边缘暂传中性（1.0/false/0），②③ 下轮叠（逐 case 单变量归因）。

**FN-safe（守的红线）**：per-zone scale 只放松「纯久静」敏感度（toilet 60s→20min），ramp 渐进 + 摔证据（pose lying / z 直立缺失走 dwell）不受 scale 拖延——真长躺 >scale 照样 ramp 过阈。真摔不靠久静也能浮出（pose=Fallen/lying 经 PoseLying boost 直抬 SFallen）。

**验收闸**：真 case 重放（铁律 [[validate_real_case_no_unit_tests]]）——d523-0611 应熄火 / d5f7-0524 bathroom 真摔仍报 / 久坐马桶不火。结果见下「验证」追加。

**给 C**：① 移植蓝本取 survival.go 那套（toilet20/sit90/默认12），**非** Xsensorv1 现有 `fall_rules_param.go` 的 stale 15/5/8min（退役 gate 路径）。② SBed dwell boost 保留（fall 单走 ramp）这个取舍要在 cd2b 在床 case 复核睡床人不衰减成 SEmpty。③ roadmap §127「FallRulesParam.Still DBN 永久消费」那行 A 会订正为「per-area 阈本轮才经 CellAreaType 接进 emission ramp」。

---

## 2026-06-16（其十二）— W3.2 单房 roomengine 主循环 + cd2b 零回归（gate②，待 C 复审）

C §43 过 gate①。起 **W3.2**（commit edbc56b）：新建 `engine` 包 `Room{filter+coupling+emission+decider}` + `Tick(fi, rhoXroom, pFallReal)` = 四轴主循环单源；replay.Run 重构复用 engine.Room（删内联循环 + dup Frame 类型，主循环单源 #1.3）。

**gate② 零回归达成**：
- **EG1**（engine 包，无 fixture）：合成"在床 InBed → 在线 LeftBed + 雷达仍躺静止近床"经 `Room.Tick` 端到端 → **P(SFallen)=0.9992 fire=true**——belief 单元那个 Ψ 相容涌现值在真 engine 主循环复现。
- **replay cd2b**（TestHR2，真 fixture 经 engine.Room）：fire=true@+531s 不变。全 roomengine 包绿。

**⚠ 一个要 C 拍的点（copy 范围现实，floor-strip 教训相关）**：

摸了 copy 候选规模 + 重新认识 replay：
1. **replay 已经吃原始 radar/sleepad/layout JSON → FrameInput**（track/pose/XY、bed_status、vertices）。所以**engine 主循环 + 原始解析已存在**，cd2b 0.9992 复现**没用到** cell/grid/track copy。
2. **cell/grid/track 的 copy 真正是给 realness（W3.3）的出生档案输入**（Displaced/ConfinedNearWall/CrossedStillPeriod/CoexistRho）用的——gate② 只测 S/B（realness/neighbor 中性），不需要。
3. **track_manager.go = 2442 行、缠 alarm/card/gate-list**（整搬 = floor-strip 陷阱反面）。其余：kalman 194/grid 676/cell 816/mirror 408/reflector 133。

**A 建议**：cell/grid/track 的 copy **下沉到 W3.3**（realness 真正消费处），且**lean-extract 不 bulk-copy**——只抽 realness/几何需要的语义（kalman 平滑 / still-box（track_manager 拆出连续指标层）/ Rect·entrance·wall / AreaDeny / mirror·reflector 几何），不搬 2442 行 production 缠绕。**请 C 裁**：(a) 按字面 W3.2 先 bulk-copy，还是 (b) lean-extract 下沉 W3.3（A 推荐，避免提前导入 production tangle）？

过 gate② + 裁了 copy 方式 → 起 W3.3（realness 接通：lean 几何/轨迹层 + adapter 译 RealnessObs）。

---

## 2026-06-16（其十一）— W3.1 四轴融合接 filter.Step 完成（gate①，待 C 复审）

C §42 三 gap 全补进图（commit 7e1ef8f）+ 三裁定全接，据图起 **W3.1**（纯 belief 包，commit 68eb4ff）。

**两轴内化（C §42 裁定的方式）**：
- **neighbor → Predict**：`Predict(online, rhoXroom)`。ρ>0 时仅 Blind from-行（SBlindRest/SBlindOpen）走 `GateBlindRow` 整流 F→L（转移先验，行和守恒）；ρ≤0 用静态 logA（零回归）。**非 gate**——ρ=0 行不变 = lost-fall 安全默认。
- **realness → logPhi**：`foldRealness(logPhi, pFallReal)`，SFallen 发射 ×P(real) 折进 logPhi（C 裁①：走同一 Correct 路径 = 真内化，独立步=软 gate 被否）。pFallReal=1 中性原样返回。
- **Step 签名** +rhoXroom +pFallReal；中性 (0,1) → 逐 tick 等价 S/B-only。无兼容 shim，改全 7 处调用点。

**三测全绿（gate① 验收）**：
- **WF1 零回归 oracle**：6 tick 中性融合（0,1）逐态等价 `Predict(·,0)+Correct`（融合包装透明，realness+neighbor 中性 = S/B-only）。
- **WF2 neighbor 整流**：Blind 出发 P(Fallen) 0.005→0.001、P(Left) 0.131→0.136（rho=0.8 vs 0）。
- **WF3 realness 调制**：P(Fallen) 真人(1.0)=0.407 > ghost(0.1)=0.064；真人=1 等价基线（不抑制真摔 = 共生律）。
- **零回归确认**：全 roomengine 包绿（NV1-8/RV/cd2b/multibed 不变）。

**请 C 复审 gate①**（融合方式 / 零回归 oracle）。过 gate① → 起 **W3.2**（roomengine 单房骨架 + copy 最小包 + cd2b 单房零回归闸 = gate②）。

---

## 2026-06-16（其十）— build order ③ 集成路线图草拟（方案乙，doc-only，待 C 审 gap 再动码）

C §40 放行 ③（四轴全内化、避免 gate 的根本目的在 belief 单元兑现）。③ = 把四轴 wire 进真 roomengine，是**大活**——floor-strip 血泪教训（"在打补丁不是建框架"）正是大活无图的后果，故**先出图再动码**（不抢跑）。

**为何 A 出图而非 C 出图**：路线图深度耦合实现（belief schema/包结构/adapter），A 有上下文写得准；C 价值 = adversarial 审图逼 gap（同五规则/ghost 法）。A建C审 proven pattern，C 出图 A 建 = 职责倒置。

**路线图 = `doc/DBN-wire-roadmap-p6.md`**（commit 待推）。survey 坐实现状 gap：belief 四轴单元验完，但 **pipeline 只接 S/B**（adapter→filter.Step），realness/neighbor 未进 per-tick 回路，无真 roomengine（replay 单房）。

核心架构决定（§2/§3）：四轴**内化进 filter 不加 gate**——neighbor→Predict（GateBlindRow 改 T_S Blind 行=转移先验）、realness→Correct（PRoomHasReal 调 fall 发射）。**正确性 oracle**=realness+neighbor 中性应逐 tick 等价现 S/B-only（零回归闸）。**唯一设计自由度**=realness 进 Correct 折进 logPhi（A 倾向）vs 独立步——请 C 裁。copy 清单（§4）/adapter 译入契约（§5）/wire 顺序（§6 五步每步可验）/验收点（§7）全列。

**请 C 审 §8 三点**（融合契约 / copy 边界 / wire gate 设哪步），审完 A 据图分步 wire（W3.1 起，每步零回归 oracle）。

---

## 2026-06-16（其九）— neighbor 据五领域规则 + D=10min 重做完成（§38/§39，待 C 复审）

C §38 复审命中（A 认）：其八 neighbor 骨架对（有向/安全默认/sole-resident/吃去-ghost），但对照用户五领域规则有 **1 核形态错 + 3 缺口**。本次据 §38/§39 全补（commit `9769e74`，方程定稿后做，非抢跑）。

**规则③ 核形态纠错（连文档一起改）**：旧 `w^dir = e^(−Δ/τh)` 在 Δ=0 取峰是**错的**——同 tick 两房 = 人没法瞬移过去 = 非 hand-off。改 **band-pass** `g(x)=x·e^(1−x)`：Δ≈0 压低但非零、峰在 1-5s（老人正常步速）、之后衰减。NV6 验：w(0)=0.21 < w(2.5s)=1.0 > w(55s)≈0；ρ(0)=0.17 < ρ(2.5s)=0.81。§A.1 同步改（曲线留 oracle，"先升后降峰1-5s"形状定）。

**规则② track 守恒（缺口补）**：hand-off 权威信号 = 本房 −1 ↔ 兄弟房 +1 track（doc P6.5「人在 X 丢必在别处冒」），非裸"兄弟房有人"。`PRealPresent`→`GainedReal`（兄弟房新增 +1 real track 守恒重现后验），载体复用 SuiteCensus/P_id 跨区账。

**规则④⑤ unit 自适应窗（缺口补，新增 §A.4）**：
- W（hand-off 检测窗）随**公共度收小**（`HandoffWindowFor`，防陌生人偶合误判 hand-off）。NV7：私有 60s→公共 18s。
- D（延迟裁决窗）随**覆盖差放长**（`DelayWindowFor`），**锚静止门限+余量**。

**§39 D=10min 定稿（同一物理时标）**：静止消失门限和 neighbor D 是同一类二义（track 消失=走了 vs 真摔被降功率滤掉）、同一时钟（功率自适应静止过滤）→ D = 静止门限 8min + 余量 2min = **10min**。层层留余量不卡边界：5min 物理 → 8min 门限(+3) → 10min 延迟窗(+2)。D=8min 是边界重合（无观察缓冲，最脆）。NV8：覆盖好 8min → 覆盖差 10min 定值。

**NV1-8 全绿** + build + 全 belief 包绿。请 C 据五规则重审（尤其规则③ band-pass 形 / 规则② 守恒 / §39 D 锚点）。build order ② neighbor 据五规则补全完，下一步 ③ 新建 roomengine wire 四轴 + copy 非DBN包（方案乙）。

---

## 2026-06-16（其八）— neighbor 隐轴落地（§A ρ_xroom 有向门控，build order ②，待 C 复审）

C §37 放行 ② 后落地（commit `6bd0069`，方程 C §16 已审，非抢跑）：`belief/neighbor.go` + 合成涌现测试。

**§A.1–A.3 实现**：
- **ρ_xroom**（§A.1）= η(rc)·max_r'[ w^dir(Δ)·c_attr·P_realPresent ]。
- **w^dir 有向核**：Δ≥0 先走后到 exp(−Δ/τ_h) / −J≤Δ<0 jitter exp(Δ/τ_j) / 窗外·真反向 = 0。**对 sign(Δ) 不对称——区别 ghost 对称核**（A 校正 C「同构」的关键）。
- **GateBlindRow**（§A.2）：Blind 行 →Fallen 按 ρ 整流入 →Left（F↔L 转移行和守恒）；ρ=0 → 行不变 → 保 lost-fall。
- **η sole-resident 连续衰减**（替离散 OFF）：多住户弱不归零。
- **§10 接口**：q 吃兄弟房**去 ghost** 占用后验（ghost 轴喂 neighbor）；neighbor 是房间 T_S 转移耦合，**不加本房 J 隐维**（状态空间不爆）。

**NV1-5 全过**：NV1 fresh hand-off ρ=0.58→F 0.5→0.21、L 0.1→0.39（人挪去邻房）/ NV2 无 hand-off ρ=0→保 lost-fall / NV3 stale 超窗 ρ=0 不抑制 / **NV4 有向：正向 ρ=0.58 vs 反向 0（区别 ghost 对称）** / NV5 sole-resident rc=1→0.58 rc=3→0.14（弱不归零）。

**待 wiring（roomengine 新建阶段）**：SiblingHandoff ← 跨房 belief 读出（去 ghost 占用）+ census；GateBlindRow 接 filter 的 Blind 行转移；η 与 §8 C_FN 同 census 双消费。axis 数学已在 belief 单元独立验。

**build order 进度：① ghost/realness ✅ 通过（C §37）；② neighbor 隐轴 ✅ 落地待 C 复审。** 四轴（S/B/ghost/neighbor）belief 单元全内化，下一步 ③ 新建 roomengine wire 四轴 + copy 非DBN包（方案乙）。

请 C 据 §A.1-A.3 + NV1-5（尤其 NV4 有向性 / NV2-3 lost-fall 安全默认）复审。

---

## 2026-06-16（其七）— realness 隐轴补全 §35 三缺口（ghost.go→realness.go，待 C 复审）

机制收敛后重做（commit `a852721`，**非抢跑**）：删 `ghost.go`（镜像消歧子集）→ `realness.go` 三类 `{Real, Mirror, Static}` 连续 P(real)。统一原理（用户领域）：真人=入口出生+会动；非人违反其一。

**补全 C §35 三缺口**：
1. **静止金属反射体**（RCStatic）：困 BirthPos + 久未移走 + 近墙（static_reflector 三签名）→ RV2 P(static)=0.978。
2. **功率自适应 + 时间过滤**：`CrossedStillPeriod` survival 慢证据 → Real（ghost/金属活不过静止降功率期，真人活得过）。
3. **real 信心闸门（最危险，连续非硬阈）**：**自主走动闩** `movedFromBirth`（Displaced ∧ 非同步反射）→ 一旦确认 real 即**锁 Real + 不 leak**；后续静止/消失 = 摔倒（走 fall 轴），**绝不被静止重判 ghost** = cd2b 病另一面防护。闸门是连续 P(real) 的结构性自带（Static 签名要求"从未走开"），非硬阈（守 A 其六）。

**共生律（决定14）**：`PRoomHasReal` = 1−Π(1−[P(real)+P(mirror)])（金属不蕴含真人）。丢真人 track 但镜像存活 → PRoomHasReal 仍高 → fall 不被"无 track"抑制 → RV5 丢真人只剩镜像 PRoomHasReal=0.99。

**RV1-5 全过**：RV1 真人 P(real)=1.0 / RV2 金属 P(static)=0.98 / RV3 镜像 P(mirror)=0.92 / **RV4 闸门：确认 real→静止 60 帧 P(real)=1.0、P(static)=0** / RV5 共生律 0.99。

**P(real) 连续量喂 decide 的设计（待 wiring）**：fall×P(real)；track 消失=ObsNoDetect，P(real) 高→S 落 Blind→Fallen ramp（dbn_cutover ③）。realness 连续量是 §26（高度不可判默认不报）与 partial_monitoring（消失不抑制 fall）的裁决者。

**adapter wiring（roomengine 新建阶段）**：RealnessObs 由 track 出生档案（BirthScore/BirthReason/MaxImpliedSpeedFromBirth）+ cell（AreaEnter/AreaDeny）译入；axis 数学已在 belief 单元独立验涌现。

请 C 据 §35 三缺口 + RV1-5（尤其 RV4 闸门 / RV5 共生律）复审。复审过 → 步骤② neighbor 隐轴。

---

## 2026-06-16（其六）— A 接受抢跑批评 + realness 轴补全方向（连续 P(real) 非硬阈）

**接受 C §35 批评：我抢跑了。** ghost.go 在 ghost 物理机制还没收敛时就写了，只覆盖最早的「镜像 co-existence 消歧」子集，漏了用户后几轮纠正的主线（功率自适应导致 track 消失、时间过滤建 real 信心、real 信心护已确认真人摔倒）。**ghost 隐轴未做完，C 判对。** 教训记下：涉底层雷达物理的设计，先吃透机制（功率自适应是代码/文档没写全、靠用户领域知识给的）再写码。

**认同补全三缺口**（金属单 track 静止 / 时间过滤 survival / real 信心护真人摔倒），复用 BirthScore/MaxImpliedSpeedFromBirth/AreaEnter/AreaDeny。

**一处精化（缺口3 real 信心闸门）**：C §35 写「real 信心**达阈**后翻转」——我建议**做连续 P(real) 后验，不设硬阈**：
- 硬阈 = 把刚拆掉的 gate 从后门请回；且安全攸关阈无法从人为数据标定（铁律 [[fall_data_is_artificial_test]]）。
- 正确：**P(real) 连续 × 消失观测(ObsNoDetect) → 连续推 P(Fallen) → §26 55%三分裁决**。现 ghost.go 的 `PReal` 即此原语（输出对），待补：① 源补全（出生地/运动/survival 三类发射，非仅 mirror co-existence）② 消失耦合（P(real)→S 落 Blind→Fallen ramp，dbn_cutover ③ fall×P(Real)）。**连续 PReal 本身就是闸门，不需另设硬阈。**
- 自洽点：realness 连续量恰是 **§26（高度不可判默认不报）与 partial_monitoring（消失不抑制 fall）的裁决者**——P(real) 高的消失 = 确信真人 → fire；P(real) 低的消失 = 高度不可判/ghost → 不报。

**ghost.go 去向**：mirror co-existence 留作「运动-同步」发射分量（mirror-ghost 类），不再当主判据；realness 轴整体重做成 3 类 + 两层（快出生/运动/区域 + 慢 survival）+ 连续 PReal × 消失耦合 Blind/Fallen + 共生律抬占用后验。

**不再抢跑**：待用户/C 确认补全方向（尤其「连续非硬阈」这条），我再动手。

---

## 2026-06-15（其五）— ghost realness 隐轴落地（§34 track=2，零补丁涌现，待 C 复审）

build-order 步骤① 完成（commit `c7d3e30`）：`belief/ghost.go` realness 隐轴 + 合成涌现测试。

**结构**（§34 锁定）：成对 realness `(T^A,T^B)` 4 态（RR/RG/GR/GG）+ co-existence ρ **pairwise 耦合**（与 §3 κ 同源，**不做 (S×T)² 全联合**）。S^(i) 仍在各自 S-filter，realness 是正交隐轴。

**机制**：P(ghost) **纯从 co-existence ρ 涌现**（共动 + 镜面几何）；mirror 几何（biasA）破 RG/GR 对称定哪个是反射；ρ=0（独立）→ 仅 RR 存活（皆真，§10 孤立安全）。**废 Track 层 Conf×P(Real) 硬外挂**——belief 出 P(ghost)，不吃 ghost_adjudicator 算好的标量。`PReal` 喂 decide（fall×P(Real)，ghost 的「摔」喂不动 SFallen，§10/dbn_cutover ③）。

**病根规避**（[[fall_detection_risk_stratified_design]]）：realness 轴**无单 track 静止输入**，ghost 只能由 co-existence ρ 涌现——结构上不可能「单 track 久静→判 ghost」。

**GH1-4 全过**（C 验收点兑现）：
- GH1 一真(A)一镜(B)+镜面指 B → **P(GhostB)=0.9747 涌现、P(GhostA)=0.0120 低**（零硬外挂）。
- GH2 独立两 track（ρ=0）→ 皆 Real（P≈0.0001）。
- GH3 病根规避：ρ=0 久持 200 帧 → P(恰一 ghost)=**0.0000**（绝不靠静止判 ghost）。
- GH4 对称（镜面未知）→ 恰一 ghost（PExactlyOne=0.9759，PGG≈0），不确定哪个（PGhostA≈PGhostB=0.49）。

请 C 据验收点（镜涌现≥阈/真低/零硬外挂/病根规避）复审。复审过 → 步骤② neighbor 隐轴（§A ρ_xroom 有向门控）。

---

## 2026-06-15（其四）— A 答 build-order + 认同方案乙（四轴全内化新建 roomengine）

**认同 §33 方案乙**（新建 roomengine + copy 非DBN包）：四轴全内化 = roomengine 整体重写，"注入躯干"前提消失；copy 非DBN包是机械改 import 非搬运。DBN 边界 = 四隐轴全内化（S/B/ghost-T^(i)/neighbor）零 gate 零硬外挂。

### build-order：A 推荐**先在现 Xsensorv1 belief 建好 ghost/neighbor 两隐轴 + 各自验涌现，再新建完整 roomengine wire**（非新建过程中一起建）。

四条理由：
1. **延续奏效的模式**：S/B 是在 belief 单元里验涌现（cd2b 验 B、MB 验多床）才扎实的；ghost/neighbor 同法——各自像 cd2b 一样在 belief 单元验「零补丁涌现」。
2. **守归因边界（刚证明其价值）**：本会话 floor-strip/译层调试之痛 = 补丁掩盖真 bug、归因纠缠；拆补丁让 belief 黑盒边界（HR-5/AC-1）重生效才揪出两 bug。**若边建轴边新建 roomengine，axis 数学错 vs wiring 错又纠缠**——丢掉刚救命的归因边界。先在 belief 单元独立验四轴，再 wire，axis 正确性与 wiring 正确性分离。
3. **依赖序**：belief 是 DBN 核心，roomengine 是脚手架。先把核心（四轴全内化涌现）验完，新建只剩 wire 已验证的核心 + copy 非DBN包。
4. **可行性**：ghost/neighbor 的 axis 数学（涌现）可用**合成多 track / 多房输入**在 belief 单元验（像 MB 多床合成、cd2b 合成），**不需 roomengine**。真实多 track（track_manager）/多房编排是 wiring，留新建时接。

### 具体序
1. **belief 单元建 ghost 隐轴**（§10 换轴：S^(i) 多份 + 跨 track ρ 耦合 → P(ghost) 涌现；废 Track 层 Conf×P(Real) 硬外挂）+ 合成多 track 涌现测试（co-existence→ghost 涌现 / 单 track 不误判 ghost）。注意 P-5 状态爆炸：基数随 track 数涨，须 bound。
2. **belief 单元建 neighbor 隐轴**（§A ρ_xroom：有向 hand-off 门控 T_S 跨房改向）+ 合成多房涌现测试（fresh 有向 hand-off→lost-fall 整流入 Left / 无 hand-off 不抑制，铁律 [[partial_monitoring_fall_suppression_law]]）。
3. **新建完整 roomengine**（方案乙）：wire 已验证四轴 belief + copy 非DBN包（改 import）+ cell/track/mirror 几何当输入 + zoneengine 旁挂产品面。
4. 全回归 + cd2b/多 case replay。

**一个 caveat**：ghost 内化是 belief 状态空间「换轴」（§10），基数增长 + P-5 bound 是真设计活（像 B 轴当初），属 belief 核心，正该在 belief 单元做。

请 C 据此定序（或纠）；定了我先起 belief 单元 ghost 隐轴。

---

## 2026-06-15（其三）— §32 拆 floor-strip 补丁完成 + 拆中揪出两处框架/译层不忠（AC-拆1~4 全过）

C §32 放行四步拆补丁，A 执行完毕（commit `2a8341d`）。**cd2b 现靠框架 (S,B) 相容零补丁涌现，不靠任何几何 δ 或 cd2b 专用规则。**

### 四步拆补丁
1. **删 δ FloorStripXY**：`emission.go`（δ 块 + lDelta）/`observation.go`（FloorStripXY 字段）/`adapter.go`（floorStrip + isDown + DownPoses）全删。
2. **删 harness TTL 离线判定**（`replay.go` sleepadTTLMs）——它建模了用户明令不建模的「中途掉线」。
3. **E5/AD4 补丁测试 → 框架涌现测试**：belief `TestFrameworkEmergenceCd2b`（LeftBed→B vac→Ψ→SFallen，零 floor-strip）；adapter `TestAdapterCd2bFrameworkEmergence`（接触轴）；replay `TestHR2Cd2bContactAxis`（真摔点 fire、在床不误火）。
4. **K^unobs 论证改 config-static**（`bed_axis.go` 注释；机制不变，C 裁过的 §C 箭头记法保留）。

### 拆中 AC-1 揪出两处**框架/译层不忠**（非 cd2b 补丁，是 translation 修正）
拆掉 floor-strip 后 cd2b replay 仍在**在床段误火**（+143s，非真摔）。AC-1 归因边界逐层定位（belief 正确）：
- **① ρ 非 config-static**：harness 把「sleepad 还没首报」误判离线（ρ=0）→ K^unobs 抽 B→vac → 造假 SFallen（near-absorbing 黏住）。**修**：`SleepadFrame{InBed,Fresh}`→`{Present,Reading}`，sleepad **存在即 ρ=1**（含首报前，§32 二态），Reading 走 InBed/LeftBed/NoReport(unknown，§30)。
- **② §D HR/RR absent 在合法在床期误否决 AtBed**：adapter 原 `HRRRObserved=nearBed` → 雷达近床但未返 vital 被当「观测到 absent」→ §D 否决 AtBed → SBed 塌、SFallen 漏。**修**：`HRRRObserved=HR>0||RR>0`（铁律 [[radar_hr_rr_bed_enter_gated]]：雷达 enter-gate，近床未测=**零信息**非 absent）。
- **③ harness 窗到双传感器期**（sleepad 首报起）：fixture sleepad 数据始于人已在床（+144s），更早 radar 缺 sleepad 上下文；真实部署 sleepad 全程在线无此缺口，窗到双传感器期=诚实复现。

### AC-拆1~4 验收（全过）
- **AC-拆1** 框架纯路径涌现：`P(SFallen)=0.9992`（零 floor-strip）。
- **AC-拆2** 全测试绿（belief/adapter/replay 三包）。
- **AC-拆3** harness 去 TTL 后 cd2b replay：`fire@+531s` 在 **LeftBed(+413s) 之后**真摔点，**在床段不误火**。
- **AC-拆4** 歧路记录 §23-§31 保留不删。

请 C 据 AC-拆1~4 复审清理结果（两处框架/译层修正一并审：ρ config-static + HRRRObserved 零信息）。

---

## 2026-06-15（其二）— decide 重写到 §26 55% 三分判据（用户裁定，A 执笔）

C §26 / 用户裁定落地：`decide.go` 从「§8 全局期望损失 P^F·C_FN>(1-P^F)·C_FP」改为 **55% 三分**：

| P^F | 裁决 | C_FN |
|---|---|---|
| ≥55% | 报 | 不需要（证据自足）|
| ≤45% | 不报 | 不需要 |
| 45–55% 两可 **且可判** | C_FN 风险偏好打破平衡（`cFN>cFP`）| **唯一作用窗口** |
| 高度不可判（Λ≤lambdaInformative）| **默认不报** | 不介入 |

- **A 立场① 推翻记录（诚实）**：原「Λ 不 gate、不可判走同一不等式兜底 fire」是资源充足逻辑；§26 资源稀缺前提下 **Λ 现作 gate**（高度不可判→默认不报），C_FN 作用域收窄至两可窗。decide.go 头注明确记此推翻。
- **知情设计决定署名**（用户 2026-06-15）：高度不可判默认不报 = 知情接受设备不足时漏掉部分**高度不可判**真摔，换告警可信度（防 alarm fatigue 烧穿稀缺护理注意力）。仅高度不可判侧；可判侧不受影响。
- DEC1-6 全过：≥55报+持续 / 两可 C_FN 打破（独处报·多人不报）/ **高度不可判默认不报（即使独处高风险，§26 核心反转）** / ≤45不报（C_FN 不救低 P^F）/ ≥55多人也报（证据自足）。阈 55/45/lambdaInformative 全 form-anchor 留 oracle。
- AD4/E5 cd2b 离线 P(Fallen)=0.9998（Λ≫1 可判）→ report，不受影响；HR-2 仍 skip 在 FloorStripXY open（下条）。

**多床绑定提醒（用户 2026-06-15，记入 on-pad 参考设计前提）**：layout 床矩形**无 bed_id**；bed↔sleepad 绑定 = **时间窗→sleepad→bed_id**。故 on-pad 参考须**按 sleepad 为键**学习（哪 sleepad 报 InBed 学那段雷达 XY），不按 layout 索引硬配。adapter 现 `Sleepads[j]↔Beds[j]` 索引对齐仅单床简化，多床/方案甲移植须改为 sleepad-keyed。

---

## 2026-06-15 — 集成步骤2 replay harness 落地 + HR-5 揪出 FloorStripXY 真缺陷（fork 待裁）

按 C §21 规格起 replay harness（`internal/roomengine/replay/`），复用 adapter+belief+probe+fixture，不碰 DB、不克隆引擎。**harness 工作正常，HR-5 归因边界兑现了它的全部价值——把缺陷精确定位到 adapter，证明 belief 正确。**

### 结果
- **HR-1 ✓ 解析忠实**：cd2b window.json → 644 帧 FrameInput，NowMs 单调，床矩形 {-70,90,150,240}。
- **HR-5 归因边界兑现**：cd2b 经 harness（raw XY 派生 floor-strip）**误火**——但不是 belief 错，是 **adapter FloorStripXY 派生**错。两层 bug，belief 全程正确：
  1. **【已修】pose-agnostic floorStrip**：+100s 一个**走动**(pose=1)的人路过床缘 → rect 几何判 floor-strip → δ 推 SFallen → 误火。修：**down-pose 门控**（δ 只对 lying/fallen 有意义，δ 实验本就 pose=6 only）。AC-2 译层修，非数据拟合。
  2. **【OPEN，真 δ 难题】rect 派生 vs 真实雷达 XY 的 layout drift**：实测 on-pad 卧姿雷达 XY（x≈-80~-130, y≈210-240）落在**画的床矩形外**（x≥-70）→ rect+margin 误判 on-pad 为 floor-strip → 正常睡觉 P^F=0.996 假信号（仅因段 <90s T_hold 侥幸没 fire）；而真摔 floor 簇（x=-170，>margin）反被判 false。**rect 几何 ≠ δ 簇边界**——A δ 实验本就是离线 Mahalanobis 簇分析，不是矩形包含。
- **HR-2 端到端正确 BLOCKED**：在 FloorStripXY 运行时实现 fork 上（下）。HR-3 三态对照同样待此。AC-3（alone<0 守卫落 adapter 边界）已做。

### FloorStripXY 运行时实现 fork（A 提请委员会裁）
| 方案 | 评 |
|---|---|
| **a. on-pad 参考（学自 sleepad-InBed 段）** | **A 推荐**。sleepad 确认 InBed 时雷达 XY 即 on-pad（定义上），累积成 on-pad 簇；floor-strip = XY 偏离 on-pad 参考（主要 y 下移）。**这正是 δ 簇边界的运行时化**（匹配离线实验方法 + κ/MM「双设备共观」哲学），per-installation 自适应、**非单 case 拟合**（守铁律 [[fall_data_is_artificial_test]]）。代价：adapter 需 on-pad 参考状态（或预跑一遍 InBed 段）。 |
| b. 扩床矩形到雷达观测足迹 | **否**。把矩形按这个 case 的 drift 扩 = 单 case 拟合，违铁律。 |
| c. 弃 δ FloorStripXY，cd2b 离线纯靠 offline-λ + 风险兜底 | 失去 A δ 实验的「可判」优势（δ≫0），cd2b 离线退回不可判→C_FN 兜底。保底但弱。 |

**A 待委员会定 a/b/c 再实现**（adapter 设计经 B/C 审过，改派生机制须对齐）。harness 是干净的验证台，方案定后此处直接验 HR-2/HR-3。

---

## 2026-06-14（交接清单）— belief 包收口，集成/neighbor 暂停待排期

> 用户裁定：belief 包在此**清晰交接、暂停**。belief 数学层阶段 1–4 全 B/C 放行、24 测绿、cd2b 治本兑现。下面是重启点。

### 状态：DONE（B/C 全放行）
- **正本**：`tools/Xsensorv1/internal/roomengine/belief/`（joint / bed_axis / filter / model / state / mm / coupling / emission / decide / probe + 各 _test）。
- **测试**：24 全绿 = T1-5 骨架 / E1-5 发射 / C1-5 耦合 / DEC1-5+Λ 裁决 / MB1-4+probe 多床。`go build ./... && go vet ./internal/roomengine/belief/` 干净。
- **关键实证**：E5 cd2b 离线 **P(Fallen)=0.9998 > P(AtBed)≈0**（δ 几何独立判定，不靠 sleepad/HRRR）；MB1 §E mixture |B|=3 存活 69×product。
- **方程**：`doc/DBN-Zone-Room.md` §1-§10 + §A-§F + §A.1-§A.3 neighbor。评审卷 feedback-p6A/B（至 round3）/C（至 §19）。

### belief API（集成时直接 wire，无需改 belief）
每帧：`cp.LogPsi(js, gxy)` + `em.LogPhi(js, obs)` → `f.Step(now, online, logPsi, logPhi)` → `js.PFallen(α)` & `ComputeLambda(js,logPsi,logPhi)` → `dec.Step(now, pF, λ, riskCtx)` → `Snapshot(...)` 出 probe。构造：`NewFilter(model,numBeds)` / `NewCoupling(geom)` / `NewEmission(geom)` / `NewDecider()`。

### 剩余里程碑 1 — 集成（**顺序在前**，neighbor 依赖它）
1. **adapter**（raw 帧 → belief 输入）：radar pose/z/HR-RR/XY + sleepad InBed/LeftBed + cell geom(covers/onbed/overlap) + room census → `Observation` / `BedGeom` / `RiskContext`。这是新代码主体。
2. **继承 Tsensor 非-belief 脚手架**：cell / track_manager / grid / stream（Xsensorv1 现仅 belief 包，脚手架未继承）。
3. **cd2b vs Tsensor 逐帧 diff**：baseline **`bd70194`**（C D-1：止血补丁 `7ffec9c` 之父，避免污染对照）。
4. **C 验收点（预告）**：集成后 cd2b 端到端**仍 fire 无回归**（probe 逐帧对照 Tsensor）。
5. **待办（B round3 建议）**：`cFN()` 对 `alone<0`（adapter 时钟回拨）的守卫——按**规则 #1.4「错误处理只在边界」落在 adapter 侧，不进 cFN 内部**（B/C 共识）。

#### adapter 译层待定点（集成入口先核 Tsensor 现有基建，非孤立试水可解）
A 自评：adapter 大部分是「wire 到 Tsensor 现有量」（pose→PoseLying、sleepad 事件→BedReading、census→RiskContext、cell 枚举→BedGeom），但**两处运行时算法依赖床区数据源、现未定**——是 plumbing 不确定（数据从哪来）非 algorithm 不确定（式子不会写），故**只能在集成时带真脚手架 + 真数据一次解对，孤立 adapter 骨架试水信息增益低**：
- **`FloorStripXY`（δ 主解运行时化）**：δ 是离线簇分析（Mahalanobis 3.35）；运行时判「XY ∈ 床沿地条簇」需**床垫区 / 地条区的数据源**（cell？`radar.areas`？layout？）——集成入口第一问。
- **`gxy`（g^xy 归属可分性运行时化）**：§4 抽象似然（能分→尖峰 / 床间均匀 / 看不见→0）的具体运行式 + 同一床区数据源——集成入口第二问。
- 次要核对：BedReading 的 sleepad 事件→状态映射 + 新鲜度 TTL；BedGeom 的 cell 枚举（R(r,b)/P(s,b)）接 Tsensor cell 层。

### 剩余里程碑 2 — neighbor 跨房 wiring（集成后）
- §A.2 `T_S` 跨房门控 wire 进**多房 filter 编排**：兄弟房**去 ghost 占用** + census 喂 ρ_xroom。方程 C §16 已审过，待 wire。
- **C 验收点（预告）**：真实 hand-off case 下 lost-fall 正确整流入 Left（ρ_xroom→1 路径）；**无新鲜 hand-off 时不抑制**（安全默认，铁律 [[partial_monitoring_fall_suppression_law]]）。

### 标定（全部留 oracle，非里程碑标定目标）
$C_{FN}$ 曲线 / neighbor $\tau_h,\tau_j,\beta$ / $\varepsilon_{art}$ / $\delta$ —— 现值全是**保守 form-anchor 非权威值**（铁律 [[fall_data_is_artificial_test]]：跌倒数据全人为，不可标定）。

---

## 2026-06-14（其四）— 阶段3 放行确认 + 阶段4(belief 单元)交付 + 集成里程碑明确

**阶段3 decide 已 B/C 放行**（C §18 D1–D6 全忠实，无修改要求）。

**阶段4（belief 单元部分）已交付**（`c1bdc2a`）：
- `probe.go` §9 逐帧 forensic 快照（α 全分量 + 边缘 S/B + P^F + Λ + κ + 裁决），供日后 cd2b 三态对照逐帧 diff。
- 多床验收（补 C §10 自认单床盲区）：MB1 §E mixture **|B|=3** FN-safe（Ψ(F) 存活 69×product）/ MB2 占用路由正确床 / MB3 covers max C2 / MB4 probe 完整性。
- **全套 24 单元测试绿**（T1-5 骨架 / E1-5 发射 / C1-5 耦合 / DEC1-5+Λ 裁决 / MB1-4+probe 多床）。

**用户裁定（2026-06-14）：阶段4 到此为止**——belief 单元验收（E5 cd2b 离线 P(Fallen)=0.9998 + MB1-4 + 24 测）已足证 belief 数学正确；**cd2b 全 replay + Tsensor diff 作为独立集成里程碑后续排期**。

**诚实现状（提请委员会知悉）**：`tools/Xsensorv1/` **目前只有 belief 包**，尚无 Tsensor 非-belief 脚手架（track_manager/cell/grid/stream）、无 adapter（raw 帧→Observation/BedGeom/RiskContext）、无 replay 入口。原计划「继承 Tsensor 脚手架」**未做**。故 §9「Xsensorv1 vs Tsensor on cd2b，baseline `bd70194`」三态逐帧 diff 是**独立集成里程碑**，需先建 adapter + 脚手架继承。belief 数学层已完备、API 齐（NewFilter/NewCoupling/NewEmission/NewDecider/Step + Snapshot），集成时直接 wire。

**P6 belief 重写状态：阶段1-3 全 B/C 放行 + 阶段4 belief 单元交付。剩余 = 集成里程碑（adapter+脚手架+cd2b replay diff）+ neighbor 跨房 wiring（依赖多房编排）。**

---

## 2026-06-14（其三）— A 对 C 第3轮复审的回应 + 阶段3 实现立场

**总体：按 C 审核走，无实质异议。** C §15 独立复审独跑 15 测全过、逐条核 ground truth 实现忠实（非测试侥幸）、与 B 一致放行；§11 自我更正 FN-safe 因果方向诚实；多床盲区（C §10）诚实标注。**B/C 一致放行阶段2 = cd2b 治本（P(Fallen)=0.9998 靠 δ 几何独立判定）最终验收兑现。A 接受。**

**A 的三条工程立场（非异议，写进阶段3 decide）：**

1. **不可判兜底不写独立分支**：§8 Λ→1（全暗）时**不 special-case**；期望损失主框架（§B 恒在）统一处理——证据两可→$P^F$ 中等，高 $C_{FN}$（独处）翻转。$\Lambda_t$ 纯诊断/forensic（probe 暴露），**绝不作 gate**。这正是 §8「诚实的不确定 + 风险不对称，不靠假装确信的耦合」的优雅。

2. **$C_{FN}$ 只设保守 form-anchor，不用现数据标定**：铁律 [[fall_data_is_artificial_test]]——跌倒数据全人为测试，$C_{FN}$ 曲线取值无法标定，只设保守形态锚（连续、各风险因子单调、多人折扣有下限不归零）+ 显式标注非权威值，留 oracle。与 C §8「取值留 oracle」一致。

3. **cd2b 主解定位守住**：decide 的 $C_{FN}$ 兜底是 **δ≈0 不可判时的退路**；cd2b（δ≫0）已在 emission 解掉（E5=0.9998），阶段3 **不改变 cd2b 解**。避免重蹈 C 早期「cd2b 终解在 decide」误框（C §3/§7 已自我更正为「emission 位置似然是 cd2b 主解、decide 兜底是退路」）。

**一个待办（提请委员会）**：§A neighbor ρ_xroom 完整方程**已交付**（`6018f11`，§A.1–§A.3），但 C §15 仍写「等 A 的 ρ_xroom 方程」——方程已落，**请 C 对 §A.1–§A.3 单独审一轮**（有向核 $w^{\text{dir}}$ 对 $\operatorname{sign}\Delta$ 不对称是否成立、$T_S$ 跨房门控、§10 三接口）。

---

## 2026-06-14（其二）— §A neighbor 方程开口已补 + 阶段 2 发射/耦合落地

### 一、§A neighbor ρ_xroom 完整方程（C 提出遗漏，A 补，落 `DBN-Zone-Room.md` §A.1–§A.3）

C §9 把"邻房 hand-off"识别为 A 漏的第四轴并把完整方程踢回 A。A 补齐三式：

- **§A.1 ρ_xroom 计算式**：$\rho^{\text{xr}}_t=\eta(\text{rc})\cdot\max_{r'}[w^{\text{dir}}(\Delta_{r'})\cdot q_{r'}]$。
  - **有向新鲜度核 $w^{\text{dir}}(\Delta)$**：对 $\operatorname{sign}\Delta$ **不对称**（先走后到指数衰减 / 反向仅容 jitter / 窗外=0）——这是 A 校正 C「与 ghost ρ 同构」的关键点：**ghost ρ 对称共存、ρ_xroom 有向时序，不照搬对称核**。
  - **去 ghost 兄弟房占用 $q_{r'}$**：吃 §10 房内 ghost 后验 $P_{r'}(\text{real-present})(1-P(\text{ghost}))$——兄弟房一个 ghost 不算合法 hand-off 落点。
  - **sole-resident 连续衰减 $\eta(\text{rc})=e^{-\beta(\text{rc}-1)}$**：替离散 `rc≠1` 硬 OFF（C §三的「第三处离散 gate」病），单住户强、多住户弱不归零。
- **§A.2 $T_S$ 跨房门控**：仅 lost-track（Blind* 行）激活，把 →Fallen 倾向按 $\rho^{\text{xr}}$ 整流入 →Left；$\rho{=}0$ → 行不变 → Blind 照常 ramp Fallen（**lost-fall 本义=安全默认**，stale/多住户不抑制，铁律 [[partial_monitoring_fall_suppression_law]]）；旧 `dampNbrFallen=0.7` 从似然层固定常数变 ρ_xroom 上确界由证据涌现。
- **§A.3 与 §10 接口三条**：① ghost 轴喂 neighbor 轴；② **不加隐维**（neighbor 是房间 $T_S$ 转移耦合、非房内 $J$ 隐维复制，状态空间 $9\cdot2^{|\mathcal B|}$ 不爆）；③ $\eta(\text{rc})$ 与 §8 $C_{FN}$ 同 census 双消费（一致下拉，保 §B「期望损失主框架 + 证据层」分工）。

**边界守住**：本节只立**框架/方向/符号**，曲线参数（$\tau_h,\tau_j,\beta$）+ 初值（HandoffWindow 60s/Jitter 5s/K 上确界 0.7/源型可信度）留 [[feedback-p6C]] §9 标定，不被单 case 绑死。

### 二、阶段 2 发射/耦合落地（`tools/Xsensorv1/.../belief/`，T 全过）

| 文件 | 章节 | 验收 |
|---|---|---|
| `mm.go` | §2 | BedGeom 三标量 covers/onbed/overlap + κ 几何冷启 |
| `coupling.go` | §3/§4/§E | κ EMA **无 max** 互活门控（可升可降）；a_j 软归属 g^xy 门控；**Ψ mixture**（C1-C5 全过，§E mixture 存活 55×product，Fallen 行 ε_art 不被 κ 覆盖） |
| `emission.go`/`observation.go` | §5/§D | Φ 分轴；离线=中性；HR/RR nearBed+非对称+**§D absent 须 gate 在独立在线 vital 源下**；δ floor-strip→Fallen（E1-E5 全过） |

**E5 cd2b 离线态实证**：sleepad 离线 + HR/RR absent 无独立源 + pose lying + floor-strip XY，经 90 帧 → **P(Fallen)=0.9998 > P(AtBed)≈0**，**不靠 sleepad/HRRR，δ≫0 几何救回**（DBN-Zone-Room §9 row3 兑现，治本 cd2b 漏报）。

### 三、🟡 B/C 共识两项已在正本处理（不阻塞阶段 1）

`filter.go`：①`Predict` 入口断言 `len(online)==numBeds`（B1 契约，wiring 错 panic 规则 1.4）；②`buildLogTBCol` 因子化 log T_B 表提出 S 循环外（B 建议，O(numBeds·nBC²)→O(nBC²)）。T1-T5 不回归。

### 四、待续（新会话）

- **§A neighbor 落地**：ρ_xroom + $T_S$ 跨房门控是**跨房**耦合，需兄弟房 belief 读出（去 ghost 占用 + census），属 wiring 阶段（依赖多房 filter 编排），阶段 2 单房 emission/coupling 不含。
- 阶段 3 decide（$\Lambda_t$ + $C_{FN}$ 连续代价，§8/§B）；阶段 4 probe + cd2b §9 三态 Xsensorv1 vs Tsensor diff。

---

## 2026-06-14 — A 方案：Xsensorv1 联合占用滤波（与 Tsensor 并存做对照验证）

### 一、根基：三个已定事实（写代码前已从代码/文档坐实）

1. **`DBN-Zone-Room.md` 是 ground truth**，联合滤波器 $J=(S,\{B^j\})$ 是正确结构。
2. **$S$ 轴（9 态全空间）+ $T_S$（9×9 propensity）已在 `state.go`/`model.go` 实现且正确 → 保留**。要新建的是 $B$ 轴 + $\kappa/\Psi/\Phi/T_B$ 耦合发射，删硬 $O_b$ 交接。
3. **$\delta_{\text{pad/floor}}$ 是唯一前置实验**：决定 emission/decide 走"确定性可判"还是"不可判+风险兜底"形态。**已完成，见第三节。**

### 二、方案结构：新建 `tools/Xsensorv1/`（继承 Tsensor 脚手架，只重写 belief 层）

与 Tsensor 并存的意义 = **对照验证**：同一份 cd2b fixture，Tsensor（旧单轴）和 Xsensorv1（新联合）同时跑，逐场景 diff DBN-Zone-Room §9 那张表。不污染、可回滚、能 diff。继承 Tsensor 全部非 belief 脚手架（cell 加载、track_manager、grid、stream、三重隔离），**唯一重写区 = `internal/roomengine/belief/`**。

| 文件 | 章节 | 实现 |
|---|---|---|
| `state.go` / `model.go` | — | **保留**：S 9 态 + T_S（从 Tsensor 拷） |
| `joint.go` | §1,§7 | $J=(S,\{B^j\})$ 联合状态，基数 $9\cdot2^{|B|}$，log 域；$\alpha_t\propto\Psi\cdot\Phi\cdot\bar\alpha$ |
| `bed_axis.go` | §6 | $B^j$ 隐变量 + $T_{B^j}=\rho K^{obs}+(1-\rho)K^{unobs}_\lambda$，$\varepsilon\ll\lambda$ 观测门控自持 |
| `mm.go` | §2 | covers / onbed / overlap 三标量（cell 枚举一次产） |
| `coupling.go` | §3,§4 | $\kappa$ EMA 无 max 互活门控；$\Psi$ 相容势，Fallen 行不被覆盖 |
| `emission.go` | §5 | $\Phi$ 分轴（接触→B / 雷达 pose,z,dwell,HRRR→S）；HR/RR nearBed+非对称；**离线=中性 $\ell\equiv1$** |
| `decide.go` | §8 | $\Lambda_t$ 似然比；fire ⟺ $P^F C_{FN}>(1-P^F)C_{FP}$ 持续 $\ge T_{hold}$；**不可判→风险兜底** |
| `probe.go` | §9 | 逐帧吐 $\alpha_t(S,\{B^j\})$ 全分量 + $\Psi/\Phi/\bar\alpha/\Lambda/\kappa$ |

### 三、阶段 0 前置实验：$\delta_{\text{pad/floor}}$ 实测（已完成）

**方法**：cd2b 标杆 case `case-cd2b-0604-16141631`（"棉被旁摔"，sleepad 在线 InBed@103/147s→LeftBed@560s 提供分段标签）。**只取 pose=6（Lying）帧**，比"垫上躺"簇 vs"床沿地躺"簇的雷达 XY 可分性（Mahalanobis）。

**关键发现（pose 枚举 `5=Fallen / 6=Lying`）**：
- **整个 case 无一帧 pose=5**。摔倒全程以 **pose=6(Lying)** 出现，与"垫上睡"同一 pose 值 → **雷达姿态层分不出"垫上躺/床沿地躺"，判别 100% 落在位置(δ)上**。

**实测**：

| 簇（只 pose=6） | n | x | y | z |
|---|---|---|---|---|
| 垫上躺 | 393 | -114±47 | 212±**17** | 7±22 |
| 床沿地躺 | 39 | -170±**0** | 160±**0** | 0±0 |

- **Mahalanobis D = 3.35**（D≳2 可分）→ **δ≫0，倾向"可判"**。
- 床沿簇 std=0 经核实**是真实静止**（t+561~598s，1Hz 稳定，conf=80，人摔后躺地不动 37s），非 artifact。
- 可分性几乎全来自 **y 方向 ~3.1σ**（垫上 y 分布窄 ±17，床沿 y=160 比床垫 y=212 低一张床沿）。

**结论**：
1. **δ≫0 但脆弱** —— 押在"垫上躺 y 窄分布"上；单 case 单摔点，按铁律 [fall_data_is_artificial_test] **只定形态（可判侧），不当精确参数刻码**。
2. **A 自我修正两处**（方案演进，诚实记录）：
   - ❌ 撤回"cd2b 必走风险兜底"：那是漏算"位置"第三柱时的错判。δ≫0 → 雷达几何这柱没断 → **cd2b 可靠 emission 位置似然救回，不需 HR/RR、不需兜底**。
   - ❌ 撤回"δ 要加 z"：两躺簇 z 都≈0（躺着本就低），z 在稳态躺零判别力。z 的判别力在"站→倒"转换瞬间（进 S 轴 dwell/z），不属 $\delta_{\text{pad/floor}}$。

**净含义（落到 emission/decide 形态）**：
```
δ≫0(本case) → emission 位置似然 = cd2b 主解（雷达几何救回，不靠 sleepad/HRRR）
δ 脆弱       → decide 不可判兜底必须存在（δ 被床上活动/贴床压小时的退路）
HR/RR 闸门   → 服务"床附近仍有 vital"的别的场景，cd2b 拿不到 HR/RR，优先级后置
```

### 四、已识别风险（A 自检，提请 B/C 重点审）

1. **HR/RR 闸门救不了 cd2b（实证）**：fixture 全程 radar HR>0 为 0 条（铁律 [radar_hr_rr_bed_enter_gated]：摔床边离床不返 vital）+ 摔倒段 sleepad 已停 → HR/RR 在 cd2b **彻底缺失**。§5 把 HR/RR 定为阶段 2 最高优先级，但它不是 cd2b 的解。
2. **状态爆炸 $9\cdot2^{|B|}$ + $\kappa$ 的 $O(|B|^2)$**：必须把 $|B|$ 明确 bound 为"房间内床数"并写死上界 + 拒绝超界（`joint.go` 断言）。
3. **对照 baseline 被污染**：Tsensor 已被 `7ffec9c` 止血补丁改过两道闸；干净 diff 需选补丁前 commit 作 baseline。
4. **$\varepsilon\ll\lambda$ 替代所有 staleness/TTL 是待验不变量**，非已知事实；阶段 1/2 须验证自持衰减能复现原 30s staleness 行为。

### 五、执行序

- **阶段 0**：$\delta$ 测量 —— ✅ **已完成**（第三节，可判侧）。
- **阶段 1**：骨架 + 联合滤波数学（joint/bed_axis/mm/filter），$\Phi/\Psi$ 中性占位。验 log-sum-exp 数值稳定、单床退化回 18 态、$\sum\alpha=1$。**与 δ 零依赖，可立即起。**
- **阶段 2**：发射与耦合（emission/coupling）。emission 位置似然为 cd2b 主解先做；HR/RR 闸门后置。
- **阶段 3**：裁决（decide）+ **不可判兜底显式存在**（δ≈0 时不得伪装确定性）。
- **阶段 4**：probe + cd2b §9 三态对照（Xsensorv1 vs Tsensor diff）。

### 六、删除清单（不移植）

likelihood 硬 `BedReleased` 分支、`7ffec9c` 止血两道闸、zoneengine→belief 硬结论交接（`Probability()` 也不吃，原始证据直进 $\Phi$）、所有 staleness/TTL 床态补丁（被 $\varepsilon\ll\lambda$ 一个不等式替代）。

### 七、提请评审组 B/C 裁决的争议点

1. δ≫0 已够给 emission 定"可判侧"形态 → **A 主张直接起阶段 1 骨架**（与 δ 零依赖），δ 严谨化（多摔点/置信区间）并行补。**B/C 是否同意先起骨架，还是要求 δ 先做严谨？**
2. emission 位置似然为 cd2b 主解、decide 兜底为退路、HR/RR 后置 —— **此优先级 B/C 是否认可？**
3. $\varepsilon\ll\lambda$ 能否真正替代 staleness/TTL —— 请 B/C 给出验证判据。

---

## 2026-06-15 — A 实证：Stage B 三修复（C §74 揪出两 bug + bed-reading 治本）

> C §74 在 Stage B 框架（engine.Unit 接进管线）揪出两个独立 bug，A 修完。本节给**实证 + commit** 供 C 复审。commit 链：`85ea5b8`(unitKey初版)→`ed6df37`(SQL network())→`2bfabbd`(sleepad-only+bed-reading)。

### 修复1：unitKey bug（`set_masklen` 不 zero host）
- **根因**（A 诊断 / C 独立核实坐实）：`host(set_masklen(room_id,80))` 只改掩码长度、**保留主机位** → suiteID 每房唯一（`fd00:0:3:111:3:100::/80` 那个 `:100` 还在）→ 多房 unit 不 group（12 rooms→**12 units 全单房**）。同 bug 在 `tenant_pref`(/48)（C 补）。
- **修法**（采纳 C 建议）：`set_masklen` → `network(set_masklen(,80/48))` zero 主机位；suite_id + tenant_pref 两处；unitKey 用 `cfg.SuiteID`（public bathroom 自身/128 CASE 分支保留不动）。
- **实证**：units **12→8**；DB 验证 101 unit 三房（:111:3:100/200/300）现共享 suiteID `fd00:0:3:111:3::/80`。→ C 验收点①(units<rooms) ✅。

### 修复2：sleepad-only 房进 DBN（C §74 揪出 + 架构师定 bed_state≠layout）
- **根因**（C 揪出）：`if hasLayout` 把 sleepad-only 房（有 sleepad、无雷达 layout）跳过 → bed_state 丢 → 做不了 hand-off 源。DB 实锤全库 **5 个** sleepad-only 房（GuestRoom :111:3:200 + 4 个 Bedroom）。
- **架构师定调**：**bed_state 不依赖 layout**（layout 是 radar 专属、给 XY 几何）；sleepad 接触式直接给 bed_state。`hasLayout` 只该管"有没有雷达几何"，不该管"房存不存在/有没有 bed_state"。
- **修法**：guard `hasLayout||hasSleepad`；sleepad-only 房 `radarLess`（无 S 轴），InBed 合成一条 bed-track 作 B 轴载体（engine.Room track-centric、无 track 无载体），LeftBed→无合成→blind→S→Left→LostReal hand-off 源；sleepad 事件触发 routeRoomFrame（sleepad-only 房唯一驱动）；sleepadPresent 改用 DB Sleepad 设备（非 cfg.Sleepads 画 layout）→ 修 radar 房 sleepad 漏判。
- **实证**：rooms **12→17**（+5 sleepad-only），units 8→10，**101 unit 现 Bedroom+Bathroom+GuestRoom 3 房**，无 crash。

### 修复3：bed-reading 治本（与修复2 bundled；C 早点的待项）
- **C 早点**：cd2b 早期 InBed 前误判 Fallen（bed-reading 缺陷，不能长期靠"恰好不显现"）。
- **根因**：旧 bed 读数 = `bases[].SleepadInBed` bool，分不清 **NoReport**（首报前）vs LeftBed → 首报前误判床空 → 躺+床空=误 SFallen。
- **修法**：bed 读数改 `tm.BedOccupancyState(card.BedState)` 三态（InBed/LeftBed/**NoReport** 权威融合、带时戳）；OnRoomFrame 签名加 `bed card.BedState`。
- **实证**（cd2b 0604）：早期(<500s) Fallen 帧 **~18→0**（消失）；bed_reading 首报前 = NoReport（不再 LeftBed 误判床空）→ +103s InBed → S=Bed(pF=0)；真摔 Fallen 帧 104、pF 峰 0.995 **不回归**。→ **bed-reading latent 误火风险治本** ✅。

### 待 C 复审的开放点
1. **sleepad-only 房形态**（C escalate 给架构师，A 实现选了）：A 选 **option A**（合成 bed-track，复用 engine.Room/Unit 机器；sleepad InBed = 人在床的证据，语义可辩；realness 静止误判无害——sleepad 房无 fall 判定）。请 C 复审此选择 vs option B（直喂 Unit 不建 Room）。
2. **pose=5/8 固件 confirmed 摔**：DBN adapter `PoseLying` 只认 6，pose=5(Fall)/8(SittingOnGround) 在 track 帧里 adapter 认不认？固件摔若只走 pose=5 不走 Fall 事件就漏 —— A 标的单独 gap，待查。

### 待验证（需对应数据/case）
- sleepad-only 房端到端 tick（需 GuestRoom sleepad InBed/LeftBed 数据，手头 case 无）
- 多房 hand-off 真触发（:3:100 lost ↔ :3:300 gained，需人跨房 case）—— C 验收点③
- d5f7 久坐假阳性（新码重跑确认 DBN 不误报）
- 七柱

---

## 2026-06-15 — A 实证：z/ObsZBand 接回（久坐误火 FP blocker，Stage C 前置）

> 承 §74 后续。C 把 z 从"优化项"升级为"Stage C 前置 blocker"（久坐误火是 FP 灾难、删 Tsensor 不可逆），
> 架构师认同。本节给实证 + commit `980bebd`。

**根因**：新 DBN 占用重写丢了 z（生产 wisefido-sensor `belief_adapter.go:124/172` 有 `ObsZBand`=posture 正向证据）。
无 z → DBN 只剩 dwell → 人静坐马桶 >60s（老人如厕常态）→ 久静超限判 Fallen 误报（alarm fatigue）。
d5f7-0524 揭示：z 分布 92.7% 贴地（z=0=真摔），但**久坐 z~40 同样被 dwell 误判**（marking AreaToilet 也不解，cell.go:324 toilet 久留=风险）。

**设计**（doc/device-room-zone.md）：
- z **单向正向证据，绝不负向**；z≥30 → 抬直立态（30-60 ZSit 抬 Sit / >60 ZStand 抬 OpenFloor）；
- z<30（贴地）→ ZNone 中性（z=0 不否决任何东西，fall 仍走 dwell）；时间积分 = 前向滤波逐帧累积。
- 阈据美国马桶座高 38-48cm；`position_z` = firmware 已解算的目标离地高度（已扣雷达安装高）。

**实现**：`belief.ZBand` 枚举 + `Observation.ZBand`；`adapter.zBandOf(z)`；`emission.radarLogS` z-band 项（lZ=8，须 ≳ dwellHi/dwellLo 抵消久坐 dwell）。

**实证**（`belief/emission_test.go`）：
- `TestEmissionZBand`：ZSit 抬 SSit log(lZ)、ZStand 抬 SOpenFloor、**z 不压 Fallen（正向 only）** ✓
- `TestZBandSuppressToiletSit`：久坐马桶(ZSit) **P(Fallen)=0.036** vs 真摔(ZNone) **P(Fallen)=0.997** ✓
  → **久坐误火 FP 治本，真摔不漏**；全 belief 测试套无回归。

**待 C 知会/复审**：
1. **lZ=8** 是 form-anchor（须 ≳ dwellHi(3)/dwellLo(0.5) 抵消久坐 dwell），精标见 feedback-p6C。
2. **cell-learning 线**（C 提的"双线"另一条：cell 用 z 学习 sit/toilet 区）**后置**——本次只接 emission 线（直接 FP 修复），cell 线慢学习/二级增强后做。
3. 端到端：d5f7（z=0 真摔）仍 fire（z-band 对 z=0 中性）；FP 抑制对正常久坐(z~40) 起作用，手头无正常久坐 case，**emission 单元即验收**。

---

## 2026-06-16 — A: realness 重写为 2 态转移矩阵(零 gate) + 验证哲学转向真 case 解剖

### 1. realness 重写（device-room-zone.md §9 落地，commit `98ff1de`）

**背景**：旧 realness = aScore 超速单调 + AND mirror，把静卧真人 B 误判 ghost（一次 42cm 体动 + 22ms 极小 dt
假超速 → aScore +3.63 → 永久 ghost）。架构师给完整模型，并**两次拦截抢跑**：①「又用 gate」（我初版状态机
`rsProvisional/rsReal/rsGhost`+硬阈 = gate-list 老路）→ 改纯转移矩阵；②「先讨论别动手」逐步对齐判据。

**模型**（2 态前向滤波 `S^r∈{Real,Mirror}`，逐帧两态跳变过程）：
- 起 `[Real=1,Mirror=0]`；`mEv` 开 R→M 闸 / `rEv` 开 M→R 闸（率→概率 `1-e^{-rate·dt}`）。
- **latch = mEv=0 时 R→M=0 → Real 吸收恒 1.0**（孤轨/无 mirror 证据/真人摔静止消失 → 永不离 Real；cd2b 真人摔
  静止不当 ghost = 矩阵结构涌现，非 if）。
- **track==2 = mEv ∝ Coexist**（孤轨 Coexist=0 → mEv=0 → 永 Real；同 dbn_cutover「孤立 ρ=0→P(Ghost)=0」结构涌现，非 `len==2` 硬 gate）。
- 墙外 = `WallMargin` 几何自判别；同步 ρ 双 track 对称 → 只归**后到者 LaterBorn** 破对称（先到=real 锚，
  enter 区不可靠故不用 InRoom）；近门 = 出生地距门 D 软斜坡（D/120）= rEv real 率。
- 单 tick 瞬态由率 ×dt 积分自然忽略。
- 删 aScore 超速病根（三重正当：单 track 判 ghost 违 track==2 / dt 噪声永久污染 / firmware 换号已由
  track_manager logicID 接管）+ `PArtifact/PGhost/PRoomHasReal` 死码。
- **输出 PReal=bR 连续后验** → FE track confidence ×100（<30 隐 /20-79 半透 /≥80 全显），**不出二元 ghost/real 判定**。
- 身份订正（关键）：**logicID 的家在 track_manager**（`makeLogicID`+`nearestAliveTrack`，已处理 firmware
  track_id 跳变/分裂/重用），census 只**消费**身份做 realness/N_r，不另立。

### 2. cd2b "0.5203→0.1999" 查清 = 旁路 harness 假象，非 realness（无回归）

realness 对 cd2b **实测零影响**（pFallReal/PReal 全程 1.0，maxPresentCount=1）。值漂移是**预存**（去掉全部本轮
改动、HEAD 仍 0.1999）。根因 = 已删的 belief-only replay harness **绕过 track_manager**，把 **track_id=88
no-target 心跳帧**（62 帧、pose=null、全零定位）当 (0,0) 真 track 喂 census → 跳 >AssocCm 拆第二 logicID →
belief 分裂稀释。真路径（`tools/replay`→Redis→`cmd/xsensor`→track_manager）本就丢 88 心跳（`noTargetSinceMs`），
**真 Xsensor 跑 cd2b 无此问题**。88 不是垃圾——是 no-people 信号，DBN 据此（无 pending silent/moving lost 时）清 logicID。

### 3. 验证哲学转向（架构师拍铁律，commit `ce0b8cf`）

**禁止 unit test / 替身 harness**——用**真实 case 解剖真生产系统本身**（"NASA 地面放 100% 真航天飞机排障"）。
简化 rig 绕真路径、引入假象（如上 track_id=88 心跳）。
- 已删 `internal/roomengine/replay/` + `cmd/replay-d5f7/` + **全部 `*_test.go`**（含前述 emission_test.go 等）。
- **Xsensor 内不放 replay 代码**（replay=外部 `tools/replay` 发 test:* Redis；`cmd/xsensor` 只消费 + 配置知 speed）。
- 验证 = 跑真 `cmd/xsensor` on `tools/replay` 重放真 case，看 `xsensor_xray` 全景日志逐 tick 切片。

### 待 C 知会/复审
1. **realness 权重** = form-anchor 留 oracle（`rcWWall/rcWSync/rcWDoor`、`rcDoorScaleCm=120` 老人室内慢、
   `rcRealBase=0.02`）。跳变率稳态使「近门+强同步」ambiguous 态 PReal≈0.59（>0.5 判 real，FE 半透显示）——
   **接受这种 ambiguous 连续表达，还是要 door 更强主导（近门→mEv 直接压制）？**
2. **前几份 p6A/p6C 引用的 `*_test.go` 验收已随铁律删除**——z/ObsZBand、belief 各柱今后走真 case（重放看 xray），非单元。
3. 下一步真 case 验证序列：cd2b（单轨 PReal 恒 1、接触轴 fire）/ 0616 跨房（hand-off + B 不误判 ghost）/ 新导 2 个 fall case。

---

## 2026-06-16 — A: ghost 判定修改的**原因**（rationale，请 C 复核逻辑链）

为什么把 ghost 判定从「aScore 超速 + 规则」改成「出生地 + co-existence 转移矩阵」——架构师本轮逐步定的逻辑：

**R1. 信息不完整 → 只能用正向证据 → 默认 real（FN-safe 接受成本）**
雷达信息天然不完整（截断重放无 enter 事件、enter 区画不准、盲区）。突现 track（无 enter 区如 lobby）没有
反向证据 → 只能认 real。**这是系统必须接受的成本**：宁可放过疑似 ghost，绝不压真人的摔（漏报代价 ≫ 误报）。
故 realness = 正向证据累积、默认 real，只在正向证据指向 ghost 才判。

**R2. ghost 只能出生时判定（过期难判）**
ghost（反射/镜像）主要由**出生地**决定，一旦形成雷达自身无法过滤（metal 例外，firmware 3-5s 主动滤）。
过了出生窗位置判据失效，只剩「轴对移/同步平移」看两 track 是否锁步（co-existence）这种复杂退路。
→ 判据必须抓出生地为主，出生后退化为同步移动为辅。

**R3. 墙外易、墙内难（分两条路）**
- 墙外：超 radar border 的 firmware 直接滤掉（到不了）；border 内 wall 外的，radar→ghost 连线穿墙取最近
  交点 ≥30cm 的几何判定**简单绝对**。
- 墙内镜像（落屋里、几何抓不到）：**出生地距门 D**（D/120 软评分，老人室内走得慢、出生在门边 D 小→120
  够罩）→ 不足则**同步移动**（轴对移/平移，出生后 3-5s 即可判）→ 仍不定则 **5min 仍活动认 real** 兜底。

**R4. enter/InRoom 不可靠 → 先到无条件 real + logicID 防 swap**
enter 区常画不准/有偏差，真人进门也可能无 InRoom event → **不能**用「出生在 enter 区」判 real。
故 **A（先到）无条件 real 锚**（截断重放首 track 也 real）；身份必须用 **logicID**（track_manager 最小作功 +
`nearestAliveTrack` 已处理 firmware track_id swap/跳变/分裂），否则 swap 造**假出生**→假 ghost。

**R5. 删 aScore 超速单 track（病根）**
旧 aScore：单 track 超速累积判 ghost。三重错：①单 track 判 ghost **违「ghost 仅 track==2」铁律**（孤轨永发=
最高风险护到底）；②单调不退，一次 dt 噪声（22ms 帧太近 → 42cm 体动算成 1916cm/s 假超速）永久污染 → 真人 B
被误判 ghost；③它想抓的 firmware 换号**已由 track_manager logicID 接管**（>AssocCm 成新号/<内吸收）。删之零损失。

**R6. 转移矩阵不用 gate（涌现而非硬 if）**
FE 用连续 **track confidence** 显示（<30 隐 /20-79 半透 /≥80 全显），**不需要二元 ghost/real 判定**。且 gate-list
是病根（dbn_cutover 已把 ghost 检测重做成纯数学涌现）。故 **latch**（确认 real 永不回 ghost，cd2b 真人摔静止不
当 ghost）、**track==2**（孤轨永 Real）都应由**矩阵结构涌现**（Real 吸收转移 R→M=0 / mEv∝Coexist），而非硬 if。

---

## 2026-06-16 — A: realness **绝不否决 fall**（FN-safe 治本 C §79）+ 真 case 实证 + 镜像判定窗优化

> 承 C 的 §79 保留意见（消费层 FN 隐患）。用 d5f7-0616 真 case（metal+swap）跑真 cmd/xsensor 揪出并治本。

### 1. 核心修复：realness 对 fall 的否决权彻底拿掉（commit `e0dec8f`）

**原 §61 消费门控是错方向**：`present≥2 → pFallReal=PR` + `eligible=PR≥0.5`——在 ghost 谷底(PR<0.5)**两重否决**那条 track 的摔（排出 room OR + 压 SFallen 发射）。

**两条原理推翻它**（架构师定）：
- **① 风险不对称**：漏报(人躺地没人救)≫误报(白跑一趟)。
- **② 有 ghost 时永远凑不齐 95% 把握去否决**：present≥2 是镜像二义——那"第二条 track"可能就是真人自己的 metal/mirror，**真人摔→反射也摔→ghost fall 恰是真摔的证据**。要否决一个 fall 需~95% 确信"这不是真摔"，二义下拿不到。

**改**：`pFallReal≡1.0` + `eligible≡true`。realness **只经 N_r→C_FN 折扣**影响（排 ghost 防 1 真人+1 影子当 2 人 → 折扣误用；只帮 fire 从不压 fall）。回归架构 [[realness_axis_redefined_real_vs_mirror]] 的「fall 不压」。

### 2. d5f7-0616 真 case 实证（揪出 §79 + 验证修复）

- **场景**：metal 先到独存近 60s（真人 lost），真人后到被判 ghost——**无解**（人也分不出谁真）；真人 ghost confidence 谷底 PR=0.09。
- **原 bug 显形**：ghost 谷底 + present≥2 → 真人的摔会被否决（log 实证门控逻辑）。本 case 侥幸没撞上（谷底在前、摔在后已恢复 real），**但纯属时间错开不是机制保证**。
- **修复后实测**：`pFallReal 全程={1}`；真人误判 91%ghost(PR=0.09)谷底 `pFR 仍=1.0`；fall 照常 fire。**谷底否决窗根除。**

### 3. latch 尝试→否决（已撤）

试过 §9.3③「5min-active→real latch」想稳住 N_r。**实测 latch 了 metal（错的那条）**——根因：rho 噪声大（真镜像对 rho 也常=0），"自主活动"判据形同虚设；且 metal 在场更久先攒够。**结论：latch 时本就分不出真伪（能分就不用 latch），z 也不准（真伪都能变高）→ 撤掉。**

### 4. 镜像/sync 判定窗优化（commit 见下）

`reflectSep`（穿墙求交，最贵）+ sync **只在出生后 ≤5s 算**，过窗冻结判定省算力（ghost 只能出生时判）。**孤轨永 Real override（Coexist==0→ForceReal）实时常开**——防冻结成 Mirror 的轨变孤轨后 N_r 算 0。实测：窗后 sep=0（几何停算）、真人变孤轨时 override 纠回 PR=1.0、fall 照常 fire。

### 给 C
- **§79 已治本**：不是减轻折扣，是 realness 对 fall 的否决权被彻底移除。请复核两条原理（风险不对称 + 95%-不可达）是否成立。
- 已删 test → 验证走真 case 解剖（本节即 d5f7-0616 真 cmd/xsensor 实证）。

---

## 2026-06-17 — A: 两条架构原则（架构师拍）+ 二义 lost-fall 三锁实证 + 原则冲突待拍

> 起于正向审查链查到 decide.go：先删 filter.go pFallReal 死路径（`9940efb`）；Decision 5 forensic 字段一删一 revert（`9af58d5`→`0421533`，结论=**不删**，它们是 §8 fire 不等式 `P^F·C_FN>(1−P^F)·C_FP, 持续≥T_hold` 的因子分解，可删性暴露的是审计 sink 没接，非死代码）。审到 55% 阈语义，架构师拍两条原则。**C 独立逐行复核三锁全部确认、接受 A 两处纠正，但抛回一个原则冲突待架构师拍——本节记录到"待拍"为止，不预设修法。**

### 原则 1（leftRoom 仅 lost 触发反算）— 架构师拍
leftRoom 非每帧维护，而是 **lost 触发反向检查**：lost 那刻回看**过去 3s** 位置变化（是否朝 Enter）+ **lost 前坐标**（距 Enter）。判据=朝 Enter 的**移动趋势**（静止门口≠离房，门口摔也消失门口）。代码现状：lost 侧无此反算（dEntry 全在 birth 侧）=待新建。

### 原则 2（55% = 二义 lost 经 neighbor 解析后的出口阈，非通用 fire 门槛）— 架构师拍
55% 原义：**neighbor event 给出之后**对"lost = fall vs 单纯信号丢失"的**终判阈**（≥55 报）。伦理：连 neighbor 帮过忙还凑不到 55% → 机构设备覆盖不足 → 护理资源同样紧张 → 不该用证据不足的报警抢占稀缺护理。**绑定"资源不足该克制"，前提=neighbor 已介入 + 仅二义 lost。** DBN 到处拿 55% 当通用 per-tick 阈 = 失原意。

### 三锁实证（A 提，C 独立 grep 逐行复核，双人确认）
二义 lost（丢轨时 P^F≈50%）要开火须闯三关，当前一关都过不了：
- **锁① Fallen ramp 引擎不存在**：`ObsNoDetect`（model.go 注释说 ramp 靠它 ×1.6）全库零实现；blind `logPhi=nil`→Correct 中性；纯 Predict 下 SBlindOpen→Fallen 仅 0.55%/tick（→Left 13.1%/tick，快 24×）。
- **锁② blind 必然 indeterminate**：blind→emission 全暗→`ComputeLambda=1`≤3→`!identifiable`→band=indeterminate→默认不报。
- **锁③ floor present-only**：engine:176 `ts.Present && fg.Step`，blind 不走总时长兜底。
- **C 认栽两处**：(a)"SBlindRest Left=0 已守"错——仍撞①②不开火，只是"drop FN"换"挂 D 到点 FN"；(b)"Left=12 趋势门控"是三锁里最浅一道，非主缺口。

### 反向实证：失原意是语义层还是行为层
- **Q1 present 自足摔 → 语义层**：emission pose/dwell 对 SFallen/SBed **等量** boost（不分 F/Bed），F-vs-Bed 靠 B 轴+Ψ（床空→压 SBed）→ 自足摔 P^F≈0.99≫55%，**55% 不 binding**，仅二义时才 binding。present 未被误压。
- **Q2 blind lost → 行为层**：blind 不 ramp→P^F 卡 0.5→永不到 55%。confirmed-then-lost（丢轨前已≥55）经 D 窗隐式 neighbor 前置≈对；**ambiguous-at-loss 无向上 ramp→55% 出口阈永够不到。**

### 🔴 原则冲突待架构师拍（C 提出，A 纠其框架）
C 指出：A 的"造 ObsNoDetect ramp 把无离房证据的 blind 顶过 55% 开火"与原则 2"资源不足该克制不报"**方向相反**，问"三锁全闭永不开火"是 bug 还是 feature。
- **A 纠 C 的框架**：三锁全闭 **≠** 原则 2 的忠实落地。原则 2 是"**整合证据(含 neighbor)后** P^F 仍 <55 才克制"——是对**已整合 P^F** 的决策规则；当前是 P^F **从未整合**（blind 冻结，neighbor 只经 handoff 往**下**压、无向上整合）。这不是"证据不足后克制"，是"证据压根没算"。
- **且原则 2 与 [[partial_monitoring_fall_suppression_law]] 在此点正面冲突**：后者（架构师本人定）说"stale/盲区/无事件 lost → **保留告警**，唯一能排除=near+有向 neighbor handoff"。按此，**无 handoff → fall 未被排除 → 报**。
- **真冲突 = 架构师两条原则在"无 handoff 二义 lost"点相反**：partial-monitoring-law（报）vs 资源克制（可能不报）。非"A 的 ramp vs 现有 feature"。
- **可能的调和（A 提，待拍）**：按**实际 neighbor 覆盖**条件化（coverage 已是 DelayWindowFor 的入参）——有覆盖+无 handoff=证据存在→报；零覆盖（孤立卫浴等根本不可观测）=真不可判→克制。

**待拍**：见正文给架构师的三选项。拍定前不出规格、不改 decide。
