# P6A 反馈日志（项目组 A 侧）— Xsensorv1 联合占用滤波 + cd2b 床边真摔 FN

> **QA 三方分离（2026-06-14，因 git 同步写冲突）**：P6 评审采用竞争式三卷，各写各的文件——
> - **本文件 `feedback-p6A.md` = 项目组 A 侧**：方案陈述 / 根因 / 前置实验 / 对评审的回应。
> - `feedback-p6B.md` = 评审组 B 侧（B 的独立评审）。
> - `feedback-p6C.md` = 评审组 C 侧（C 的独立评审，与 B 竞争）。
>
> 三方不写同一文件 → push 不再非-fast-forward 撞车。倒序，最新在上。
> **共同基线**：`doc/DBN-Zone-Room.md`（联合占用模型 ground truth）、`doc/DBN-cd2b.md`（cd2b case）、`CLAUDE.md`。

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
