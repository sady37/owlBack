# P6A 反馈日志（项目组 A 侧）— Xsensorv1 联合占用滤波 + cd2b 床边真摔 FN

> **QA 三方分离（2026-06-14，因 git 同步写冲突）**：P6 评审采用竞争式三卷，各写各的文件——
> - **本文件 `feedback-p6A.md` = 项目组 A 侧**：方案陈述 / 根因 / 前置实验 / 对评审的回应。
> - `feedback-p6B.md` = 评审组 B 侧（B 的独立评审）。
> - `feedback-p6C.md` = 评审组 C 侧（C 的独立评审，与 B 竞争）。
>
> 三方不写同一文件 → push 不再非-fast-forward 撞车。倒序，最新在上。
> **共同基线**：`doc/DBN-Zone-Room.md`（联合占用模型 ground truth）、`doc/DBN-cd2b.md`（cd2b case）、`CLAUDE.md`。

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
