# P6A 反馈日志（项目组 A 侧）— Xsensorv1 联合占用滤波 + cd2b 床边真摔 FN

> **QA 三方分离（2026-06-14，因 git 同步写冲突）**：P6 评审采用竞争式三卷，各写各的文件——
> - **本文件 `feedback-p6A.md` = 项目组 A 侧**：方案陈述 / 根因 / 前置实验 / 对评审的回应。
> - `feedback-p6B.md` = 评审组 B 侧（B 的独立评审）。
> - `feedback-p6C.md` = 评审组 C 侧（C 的独立评审，与 B 竞争）。
>
> 三方不写同一文件 → push 不再非-fast-forward 撞车。倒序，最新在上。
> **共同基线**：`doc/DBN-Zone-Room.md`（联合占用模型 ground truth）、`doc/DBN-cd2b.md`（cd2b case）、`CLAUDE.md`。

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
