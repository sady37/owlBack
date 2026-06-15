# feedback-p6C — 评审组 C 对 P6（联合占用 DBN）方案的评审

> **30 秒导读（委员会用）**

**总判**：结构正确、纪律到位，**支持建 Xsensorv1**。$S$ 轴保留、$B$ 轴联合化、删硬 $O_b$、并存对照——方向无误。评审价值在**纠正优先级叙事** + **补一根 A 遗漏的轴** + **指出一条贯穿全案的病**。

**C 的框架：cd2b 二义性 = 四轴，终裁在期望损失**

| 轴 | 证据 | 解的二义性 | A 状态 | C 节 |
|---|---|---|---|---|
| 空间 δ（emission $g^{xy}$） | 雷达 XY | 床边摔 vs 垫上躺 | ✅ 内化 | §5 |
| 时间（dwell 符号） | 久静 × cell 容忍 | 卧地 vs 久坐 | ✅ 内化 | §6/§7 |
| 裁决（$C_{FN}$ 期望损失） | 风险因子（独居/夜/人数） | 低 $P^F$ 报不报 | 🟡 形态待定 | §7/§8 |
| **跨房 neighbor（ρ_xroom）** | **兄弟房 hand-off** | **挪去邻房 vs 本房真摔** | **🔴 A 遗漏** | §9 |

**一条主线（贯穿三处的同一个病）**：风险/归因被做成**离散 gate**，而 DBN 终极（发现风险）要**连续代价加权**——
① §7 dwell 符号（容忍 cell 二分）② §8 risk_evaluator（离散三档 RiskLevel）③ §9 neighbor sole-resident 门（rc≠1 硬 OFF）。三处都应统一到 §8 的期望损失 $P^F C_{FN} > (1-P^F) C_{FP}$。

**C 的两处自我修正（诚实记录）**：① 撤回"cd2b 必走风险兜底"（trim 版数据缺失误读成物理不可判；完整版 δ≫0 可判）② 更正裁决层定位——风险偏好裁决是**主框架（恒在）**，非"退路"；emission/dwell/位置似然全是供 $P^F$ 的证据层。

**C vs B 的分工**：B 审"方程写对没有"（设计层，4 个参数阻塞项）；**C 审"方案押对没有 + 目的对不对"**（实证层：fixture 实测 + 代码核查 + 风险目的）。C 独占的三手 B 结构上做不到：
- §5 用 fixture 把 B 的开问题落值（B1 μ、B2 λ、B3 在床 HR/RR 100% 缺失→设计否决）。
- §6/§7 dwell 符号是框架不是标定（cell 容忍可靠性升格为框架级前提）。
- **§9 neighbor 是 A 遗漏的第四轴——它根本不在 `DBN-Zone-Room.md` 里，只在代码里，B 不读代码看不到。**

**C 列出的三个前提（Xsensorv1 落地必答）**：δ 跨 case 稳定性（标定/脆弱）、cell 容忍判定可靠性（框架）、$C_{FN}$ 代价函数（框架形态 + 标定取值）。

---

# P6C 反馈日志（评审组 C 侧）— Xsensorv1 联合占用滤波方案评审

> **QA 三方分离（2026-06-14）**：P6 评审采用竞争式三卷，各写各的文件防 git push 撞车——
> - `feedback-p6A.md` = 项目组 A 侧（方案陈述 / 根因 / 前置实验）。
> - `feedback-p6B.md` = 评审组 B 侧（与 C 竞争）。
> - **本文件 `feedback-p6C.md` = 评审组 C 侧**：C 对 A 方案的独立评审。
>
> 三方不写同一文件 → push 不再非-fast-forward 撞车。倒序，最新在上。
> **审查基准**：`doc/DBN-Zone-Room.md`、`doc/DBN-cd2b.md`、`CLAUDE.md`、`doc/feedback-p6A.md`（被审方案）。

---

# feedback-p6C — 评审组 C 对 P6（联合占用 DBN）方案的评审

状态：评审意见（竞争评审，与评审组 B 并列）。
评审对象：项目组 A 的 P6 提交 = [[DBN-Zone-Room]]（联合占用模型方程）+ Xsensorv1 重写方案（$S$ 轴保留、$B$ 轴联合化、删硬 $O_b$、并存对照）。
评审基准：fixture 实证（case-cd2b-0604）+ 现有代码（Tsensor belief 层）+ memory 铁律。
关联：[[DBN-Zone-Room]]、[[belief_dbn_completeness]]、[[belief_gate_to_matrix]]、[[radar_hr_rr_bed_enter_gated]]、[[fall_detection_risk_stratified_design]]、[[partial_monitoring_fall_suppression_law]]。

---

## 0. 总判

**结构正确、纪律到位，支持建 Xsensorv1。** $S$ 轴保留、$B$ 轴联合化、删硬 $O_b$、并存对照——方向无误。联合滤波 $\alpha\propto\Psi\cdot\Phi\cdot\bar\alpha$ 是正确结构，不是改进方向而是 ground truth。

评审的价值集中在**纠正 A 方案叙事里两处被 fixture 数据推翻的优先级判断**，以及守住三条工程纪律。下分实质问题与纪律点。

---

## 1. 实质问题（按严重度）

### 🔴 P-1 δ 是标定参数，非架构分支点（已与 A 达成一致）

A 初版方案把 $\delta_{\text{pad/floor}}$ 定为"决定 emission/decide 形态的单一前提"，阶段 0 阻塞一切。

**评审意见：错。** 联合滤波骨架 $\alpha\propto\Psi\cdot\Phi\cdot\bar\alpha$ 对所有 $\delta$ 不变：

- $T_S$ 9×9 + $T_B$（$\varepsilon\ll\lambda$ 门控）
- $\Phi$ 按 attachment 分轴（接触→$B$，雷达→$S$）
- $\Psi$ 物理相容表（$\psi(F,\text{occ})=\varepsilon_{\text{art}}$）
- $\Lambda_t$ + 风险不对称裁决

$\delta$ 只是 $\Phi$ 里 `g^{xy}` 项的乘子：$\delta$ 大→尖峰，$\delta$ 小→平坦。`decide.go` 不可判路径也**不是"δ 小才需要"**——即使 $\delta$ 大也有边缘 case（贴床躺/翻身使垫上分布变宽 → $\delta$ 缩小），那条路径是安全网，恒在。

**结论：** $\delta$ 进标定阶段再调，**阶段 1 骨架与 $\delta$ 零依赖，立即可起**。$\delta$ 不"二选一定形态"，只给 cd2b 这一个 case 贴"落哪条路"的验证标签。**A 已接受。**

### 🔴 P-2 HR/RR 闸门救不了 cd2b——A 误置为 emission 最高优先级（fixture 实证）

A §5 把"HR/RR nearBed + 非对称似然"钦定为 emission 阶段最高优先级离线闸门，并暗示其为 cd2b 离线态的解。

**评审意见：fixture 实证推翻。** 跑 case-cd2b-0604：

- radar HR/RR：42 条带 heart_rate 字段，**无一 >0**——印证 [[radar_hr_rr_bed_enter_gated]]：人摔床边（离床）radar 不返 HR/RR。
- sleepad：t+602s（完整版）停。

cd2b 摔倒段 HR/RR **两源皆空**，HR/RR 闸门对它**结构性零作用**。

**结论：** HR/RR 闸门有价值，但服务"**床附近仍有 vital**"的别的场景，不是 cd2b 的解。**别把"做完 HR/RR"误当成"解决了 cd2b"。** 优先级**后置**。

### 🔴 P-3 cd2b 的真正解是 emission 几何位置似然，不是 HR/RR、不是风险兜底（双向修正）

C 初评曾据 trim 版数据判"cd2b 近乎四柱皆断 → 必走 decide 风险兜底"。**此判错**——trim 版把 sleepad 裁空、中段 z 失效，是**数据缺失**被误读成**物理不可判**（正是假性不可判）。

**用完整版 0604 重测：$\delta\gg0$（~1 nat，边际可分离），雷达几何这柱没断。** 故：

- **cd2b 离线摔的解 = emission 位置似然**，靠雷达几何直接救回，**不需要 HR/RR、不需要风险兜底**。
- A 原方案"$\delta$ 决定 emission 形态、emission 位置似然是 cd2b 主解"——**这部分 A 是对的，C 初评说轻了，撤回**。

**但 $\delta$ 脆弱**：押在"垫上躺 $y$ 分布窄"，人翻身/坐起/大范围活动→垫上分布变宽→$\delta$ 缩小；且单 case 单摔点只证"这次摔的位置可分"，不证"所有床沿摔可分"。**铁律：定形态、不定参数。**

**净含义（落到方案重心）：**

```
δ≫0(本case) → emission 位置似然 = cd2b 主解（雷达几何救回）
δ 脆弱       → decide 不可判兜底必须存在（δ 被活动/贴床压小时的退路）
HR/RR 闸门   → 服务"床附近仍有 vital"的别场景，优先级后置
```

emission 几何位置似然要**做扎实**（cd2b 主解）；decide 不可判兜底**保留为退路**；HR/RR 闸门**后置**。

### 🟠 P-4 δ 实验设计：z 不进 $\delta_{\text{pad/floor}}$（双向修正）

C 初评建议 $\delta$ 扩成 $(XY,z)$，理由是"判别力大头在 z 时间轨迹"。

**A 反驳，C 接受撤回：** 对**稳态躺**，两个躺簇 $z$ 都 $\approx 0$（躺着本来就低），$z$ 在此**零判别力**。$z$ 的判别力在"**站→倒**的转换瞬间"（直立 $z$ 高→骤降），那是进 **$S$ 轴的 dwell/z 通道**，不属于 $\delta_{\text{pad/floor}}$ 这个**稳态位置量**。

**结论：** $\delta_{\text{pad/floor}}$ 只测 $XY$ 稳态可分性；$z$ 轨迹归 $S$ 轴 pose/dwell 发射。两个量不混。

### 🟠 P-5 状态爆炸 $9\cdot2^{|B|}$ + κ 的 $O(|B|^2)$（纪律）

$J$ 基数随床数指数、$\kappa$ 床间耦合平方。养老院房 ≤3–4 床可控，但必须：

- `joint.go` 把 $|\mathcal B|$ **硬 bound** 为房间床数 + 写死上界（`maxBeds=3`）+ **超界拒绝**（panic/error，不静默膨胀）。

---

## 2. 纪律点

### 🟡 D-1 对照 baseline 必须选止血补丁前的 commit

Tsensor 已被 `7ffec9c`（cd2b 两道闸止血）污染。干净 diff 的 baseline = **`bd70194`**（`7ffec9c` 的父，止血前）。否则比的是"打补丁 Tsensor vs Xsensorv1"，不纯，会低估 Xsensorv1 改进。

### 🟡 D-2 "$\varepsilon\ll\lambda$ 替代所有 staleness/TTL" 是待验不变量

不是已知事实。阶段 1/2 须专门验证 $B$ 自持衰减能**复现原 30s staleness 行为**，别假设。建议作为 `joint_test.go` 的显式用例。

---

## 3. 评审给出的执行序（修正版）

```
阶段0  测 δ —— 只测 XY 稳态可分；定位"验证标签"非阻塞；用完整版 0604（非 trim）
阶段1  骨架(joint/bed_axis/filter) —— 立即起，不等 δ；含 ε≪λ 复现 30s staleness 验证
阶段2  emission 几何位置似然(g^{xy}) —— cd2b 主解，做扎实
阶段3  decide 不可判+风险兜底 —— 与 emission 并重，δ 脆弱时的退路（恒在）
阶段-  HR/RR nearBed 闸门 —— 后置，服务"床附近仍有 vital"场景
阶段4  probe + cd2b 三态对照 —— baseline = bd70194
```

**重心：emission 几何位置似然（cd2b 主解）+ decide 不可判兜底（退路）。HR/RR 后置。**

---

## 4. 一句话评审

A 的方案**结构是对的**，但"阶段 0 决定一切 + HR/RR 闸门优先"这条叙事，被 cd2b 自己的数据双向修正：δ 是标定参数非分支点（解除阶段 0 阻塞）；cd2b 离线摔靠 emission 雷达几何救回（$\delta\gg0$），HR/RR 因两源皆空对它零作用、应后置；decide 不可判兜底作为 δ 脆弱时的恒在退路。**不削弱方案，使重心落到真正承重处。**

支持建 Xsensorv1，阶段 1 骨架可即起。

---

## 附：δ 实测摘要（case-cd2b-0604，完整版）

- 雷达几何稳态可分性 $\delta(XY)\approx 1$ nat（边际可分离，$\gg0$）。
- HR/RR 两源（radar 42 条 / sleepad t+602s 后）摔倒段皆空——实证铁律。
- 结论：cd2b 落**可判侧**，主解 emission 位置似然。δ 严谨化（多摔点 / pose=6 且床区严格筛 / 置信区间）并行补，服务 `g^{xy}` 初始权重标定。

---

# §5 增补（2026-06-14）：C 对评审组 B 阻塞项的实证回应

> B 的评审是**纯设计层**（自我限定"不碰 fixture"），逐方程挑出 4 个参数未约束的阻塞项（B1–B4）+ 3 澄清（C1–C3）+ 1 标定依赖（S1）。这些洞**真实**。但 B 只能"建议 A 定义"——B 没有数据去定值。
> C 的差异化价值：**用完整版 cd2b fixture（case-cd2b-0604-16141631）实测，把 B 停在开问题的地方，推进到初始值/设计结论。** 下逐条回应 B 的阻塞项。

## 对 B1（$\mu$ 未约束）的实证回应 — 给出初值 + 默认对称的依据

B 问：$K^{obs}$ 的 vac→occ 参数 $\mu$ 从未约束，$\mu$ vs $\varepsilon$ 关系决定 $B$ 链漂移。

**实测（完整版 0604 sleepad）**：413s 跨度内床态翻转 **仅 1 次**（InBed→LeftBed @413s）。翻转率 ≈ 1/400 /帧。

**C 给出的初值与依据**：
- $\varepsilon$（occ→occ 自持的补）对应"每 ~400s 才翻一次" → $\varepsilon \sim \mathcal O(10^{-2})$/帧（1Hz 下），在线自持极强。
- **$\mu=\varepsilon$（对称）是 fixture 支持的默认**：数据中上床/离床各仅观测到一侧，无证据支持"进床比离床容易"的非对称漂移。无证据则不引入偏置——对称是最大熵默认。
- B 担心的"$\mu\gg\varepsilon$ → 向 occ 漂移"在 fixture 上无依据；若未来 case 显示上床事件系统性多于离床，再放开对称约束。

→ **B1 从"未约束"推进到"$\mu=\varepsilon\sim10^{-2}$，对称默认有数据支撑"。**

## 对 B2（$K^{unobs}_\lambda$ 弛豫目标 + 速率未指定）的实证回应 — 给出 λ 硬锚点

B 问：弛豫目标是均匀（定义 A）还是泄漏（定义 B），速率 $\lambda$ 未定。

**实测**：sleepad 正常帧间隔中位 **10.0s**（最大 11s）→ 离线判定阈值 ~3× = **30s**。sleepad 末帧 @558s（报 LeftBed），其后 radar 独跑 509s。

**C 给出的 λ 锚点**：
- $\lambda$ 半衰期**必须 < 30s**——使陈旧 occ 在 fall confirm 窗（90s）**之前**蒸发完。否则 cd2b 漏报重演（陈旧 InBed 撑到 confirm）。
- 这个 30s **正是原 staleness/TTL 窗**——故 **B2 的 $\lambda$ 标定与 C 的 D-2（$\varepsilon\ll\lambda$ 复现 30s staleness）是同一个锚点**：$\lambda$ 取"半衰期 ≈ 10–15s"即可在 30s 内蒸发 occ，且 $\varepsilon(\sim10^{-2}) \ll \lambda$ 自然成立。
- 弛豫目标 A/B 之争：cd2b 只依赖 occ→vac 方向（B 自己也指出两定义在此等价）。C 主张**定义 B（泄漏，vacant 吸收）**——理由是养老院场景"床垫离线期间无人上床"远比"离线期间有人上床且需模型自发学到 occ"常见；vacant 吸收避免空房离线被弛豫成 P(occ)=0.5 的伪占用。床垫离线时真有人上床 → 靠 radar pose/位置经 $\Psi$ 涌现，不需 $B$ 链自发升 occ。

→ **B2 从"目标/速率未定"推进到"定义 B + λ 半衰期 10–15s，与 D-2 共锚 30s"。**

## 对 B3（HR/RR absent 否决 AtBed 的 FP 风险）的实证回应 — 风险量级 = 100%，升级为设计否决

B 标注：HR/RR absent 否决 AtBed 时，人在床但 HR/RR 因遮挡缺失 → 误推向 $F$ = fall 假阳性。B 定性为"固有取舍，建议 §5 加风险标注"。

**实测（完整版 0604，人确在床的 558s 段）**：radar track 帧 **573 帧，返回 HR/RR>0 的 = 0 帧。在床段 radar HR/RR 缺失率 = 100%。**

**C 的结论（比 B 的"标注风险"更进一步——这是设计否决，不是标注）**：
- radar 在床位被 firmware enter-gate，**结构性不返 vital**（铁律 [radar_hr_rr_bed_enter_gated]）。故 B3 的 FP 不是"偶发遮挡风险"，是 **100% 必然**：只要 radar 是房内唯一 vital 源，`ℓ_hrrr(absent|AtBed)=1/L_hr` 会在**每一个在床帧**否决 AtBed。
- 由此锁定设计约束：**HR/RR 的 absent 分支必须 gate 在"房内有独立在线 vital 源（sleepad）"之下。radar 自身 HR/RR 的 absent 不得作为否决 AtBed 的证据**——radar 在床不返 vital 是结构性零信息，不是"人不在床"信号。把它当否决证据 = 100% 在床误判。
- 这与 P-2 互补：P-2 说"HR/RR 救不了 cd2b（两源皆空）"；B3+实测说"HR/RR 若不 gate 还会害别的场景（在床 100% 误推 F）"。**两面夹击 → HR/RR absent 分支不仅后置，还必须带 sleepad-online gate 才能启用。**

→ **B3 从"建议标注风险"推进到"风险量级 100% + 必须 sleepad-online gate 的设计否决"。**

## 对 C3（$\varepsilon_{art}$ 量级未定）的补充

B 正确指出 $\varepsilon_{art}$ 需在 log 域与 $L_{in}$ 联合定数量级。C 补一个约束来源：$\varepsilon_{art}$ 要小到使 $\psi(F,occ)$ 压得住"床上翻身 pose 误读为 Lying"，但大到不在 log 域被 $\Phi$ 正向似然淹没。结合 B1 的翻转率，床上 pose 噪声事件频率可作 $\varepsilon_{art}$ 上界的经验锚——建议标定阶段用"在床段 pose=Lying 帧占比"反推，而非凭空取 $10^{-3}$。

## 净结论（C vs B 的分工与 C 的增量）

| B 提出（设计层开问题） | C 实证推进（给值/给设计结论） |
|---|---|
| B1 $\mu$ 未约束 | $\mu=\varepsilon\sim10^{-2}$，对称默认有数据支撑（413s 仅 1 翻转） |
| B2 弛豫目标/λ 未定 | 定义 B（泄漏）+ λ 半衰期 10–15s，与 D-2 共锚 30s（帧间隔实测） |
| B3 HR/RR FP 风险（建议标注） | 量级 = 100%（在床 573 帧 0 vital）→ 升级为"absent 须 sleepad-online gate"的设计否决 |
| C3 $\varepsilon_{art}$ 量级 | 用在床段 pose=Lying 占比反推，非凭空取值 |

**B 审"方程写对没有"，C 审"方案押对没有"且用数据给 B 的开问题落值。两卷互补；C 的增量 = 把 B 的 4 个开问题里的 3 个从'待 A 定义'直接推进到'有 fixture 依据的初值/否决'，这是 B 自限设计层做不到的一手。**


---

# §6 增补（2026-06-14）：步速上限前提 + "慢=高危" — C 发现的 dwell 轴反向冲突

> **前提（A 提请补充，C 据代码权威值落实）**：老人室内步行速度有生理上限；且**越慢风险越高**——弱→易摔且难恢复→慢速/静止不是"低信号"而是"高危信号"。
> 此前提在 C 的 §5 增补与 A 方案中均未显式纳入。纳入后**修正一条 B 未发现、跨 A 方程层的冲突**。

## 前提的代码权威值（引用，非自造）

第一层 **速度上限**已实现（`belief_adapter.go` / `kalman.go`）：老人步速标定 0.3–1.0 m/s；封顶生理上限 150cm/s（挡超人噪声）、下限 30cm/s、全局兜底 60cm/s；per-device EWMA 学习。**用途仅 ghost-filter**（剔 track-swap 假跳），不做"走速判走动"（radar 量化重，per-frame 走速不可靠）。

第二层 **慢=高危**已实现于 reachable-exit（`belief_adapter.go` 安全偏置原文）：*"逼近速度慢速封顶=老人保守（越弱越易摔且难恢复→宁少抑制少漏报）；独居老人步态慢→封顶自然低→绝不给『他本可走出去』虚高信用→不漏报。"* 即：走得越慢，越不敢判"已走出"，保 fall 信念。**方向正确。**

## C 发现的冲突：`survival.go` 的"久静=正常"与"慢=高危"在非容忍 cell 上反向

`survival.go` 另有一条相反逻辑（为破 SFallen 近吸收棘轮设）：*"此处久站越久=越正常=反 fall 证据……随 dwell **单调下压** SFallen。"*

两条在"慢=高危"前提下冲突，且冲突点正落 cd2b：

| 机制 | 对慢速/久静的处理 | 在 cd2b 床沿静止 37s 上 |
|---|---|---|
| reachable-exit 安全偏置 | 慢→不敢判走出→**保 fall 信念** | 正确 |
| survival dwell-tolerance | 久静→当正常久站→**压 SFallen** | **错**（把摔后卧地压成正常） |

cd2b 实证：床沿簇 t+561~598s 静止 37s（1Hz 稳定，conf=80，已核实真实静止非 artifact）。survival 的 dwell-tolerance 会把这段往"正常久站、压 SFallen"拉——**与"慢=高危"前提要求的方向相反**。

## C 给出的修正：dwell 轴按 cell 容忍属性分流，不是全局单调下压

"久静→压 fall"只在**容忍 cell**（椅/沙发，人常坐久驻）成立；在**非容忍 cell**（床沿地面、开阔地）下，"慢=高危"前提要求久静止**单调上抬** SFallen，而非下压。区分键 = **cell 容忍属性**（`calibration.go` 已有 `tolWeight`/`ToleranceFactor` 这个量，可直接驱动分流）：

```
dwell 久静止 × cell 容忍属性：
  容忍 cell(椅/沙发)   → 久静 = 正常久驻 → 单调下压 SFallen（survival 现行为，保留）
  非容忍 cell(床沿/开阔) → 久静 = 高危卧地 → 单调上抬 SFallen（"慢=高危"，现缺）
```

这与 §5 对"静止 37s"的处理衔接：该判据不只是"这次确实摔了"的正面证据，更应在 **dwell 轴单调利用**（HSMM 驻留尾）——非容忍 cell 静止越久，$P(F)$ 越往上爬，不是到阈值才触发、更不是被 tolerance 下压。

## 为什么这是 C 独占的一手（vs A / B）

- A 方案 §5 dwell 核写 "$\ell_{dwell}(\text{still}\ge\tau\mid F)=\ell_{dwell}(\text{still}\ge\tau\mid\text{AtBed})=D>1$，dwell 不分 F/AtBed"——只处理"静止占用 vs 活动"，**未处理非容忍 cell 静止的 F-单调上抬**，也未接入"慢=高危"。
- B 在设计层审了 dwell 核（C3 提 $\varepsilon_{art}$、未及 dwell 方向），但 **B 没碰 survival.go 的"久站压 fall"，更未发现它与"慢=高危"在非容忍 cell 反向**——因 B 自限不碰 fixture，缺 cd2b"床沿静止 37s 在非容忍 cell"的实测语境。
- C 凭"前提（慢=高危）× 代码（survival 现行下压）× fixture（cd2b 床沿非容忍 cell 静止 37s）"三者交叉，定位这条冲突并给出 cell-容忍分流的修法。

→ **新增 C 的设计结论：dwell 轴对 SFallen 的单调方向必须由 cell 容忍属性 gate；非容忍 cell 久静止单调上抬 SFallen（"慢=高危"），仅容忍 cell 保留 survival 现行的久静下压。这是 cd2b 床沿摔在 dwell 轴上不被压死的必要条件，与 §5 的 emission 位置似然（空间轴）互为正交保险。**


---

# §7 增补（2026-06-14）：站立/静坐时长体系 + 两个框架级澄清（C 自我修正 §6）

> A 提请：老人静立站立一般 5min，故设静止站立 8min/12min 阈值；静坐 90min。这些已成体系，C §6 称"非容忍 cell 久静上抬 SFallen 现缺"**是错的——机制已存在，C 撤回该'独占发现'**。
> 借此回答 A 的两个问题：(1) 这些影响框架还是标定？(2) DBN 终极目标是发现风险，二义性从风险偏好考虑。

## 一、已存在的时长体系（C 此前漏读，更正记录）

`survival.go` dwell 生存尾（Weibull，`-ln S_vol`）按 zone 分档，**且 tolerance 翻转方向已实现**：

| zone | scale 尾尺度 | 方向 |
|---|---|---|
| 浴室 toilet/shower | 20min（便秘安全/医学） | 正向 ramp |
| 浴室其它 | 12min | 正向 |
| 学习久坐区 | 90min（静坐） | 正向 |
| 床/休息区 | 不报 | — |
| 未知/开阔 | 20min | 正向 |

`fallLRFromDwell`：**非容忍 cell(mult=1)正向 ramp `1+(d/scale)^shape` 久静上抬 SFallen；容忍 cell(mult>1)`(1-tolWeight)<0` 随 dwell 单调下压**。站立自学习：非浴室 stand-static ≥12min（物理：人很难纯站立超 12min）→判站位/坐位；RestZone 8min 强化。夜间短尾（久静更可疑）、雷达远边缘 ×1.5。

→ **C §6 提议的"按 cell 容忍属性分流上抬/下压"已在代码中（委员会选的 A 方案）。C 撤回 §6 独占性主张，保留其与 cd2b 床沿语境的衔接价值（确认该机制对 cd2b 非容忍床沿生效），但不再声称是新发现。**

## 二、问题(1)：影响框架还是标定？——**数值是标定，方向/符号是框架**

| 量 | 类别 | 理由 |
|---|---|---|
| 20min/12min/90min/8min 尾尺度 | **标定** | Weibull `scale` 参数，改它只左右平移 ramp 曲线；框架（dwell 进 $\Phi$ 的 `ObsDwellStill` 发射）不动。注释自承"P9.6 待 oracle 收紧" |
| 5min/8min/12min 站立物理常数 | **标定** | 生理先验锚，定 scale 初值，可随 oracle 调 |
| 夜间短尾、边缘 ×1.5 | **标定** | scale 乘子 |
| **dwell 证据方向（上抬 vs 下压）由 cell 容忍属性决定** | **框架** | 决定同一"静止 600s"在似然层是 LR>1（增 fall）还是 LR<1（减 fall）——**符号翻转是结构不是数值** |

**关键框架命题**：dwell 对 SFallen 的**符号**挂在 `toleranceMult`（cell 容忍属性）上。故 **"cell 容忍属性的权威来源"是框架必须钉死的**——若床沿被误学成容忍 cell，符号翻错，真摔被下压（呼应 C 此前标的"cell 容忍标定可靠性"脆弱点，现定位为**框架级**而非标定级风险）。

→ **结论：时长阈值本身（数值）只影响标定，可放心随 oracle 调；但 dwell 方向由 cell 容忍属性 gate 这件事是框架，且 cell 容忍属性的判定可靠性是框架级前提，须有权威源（FE 画 > feedback > 自学），不能让自学误翻符号。**

## 三、问题(2)：DBN 终极是发现风险，二义性从风险偏好裁决——**C 自我修正裁决层定位**

A 的命题切中 C 此前一个方向性错误。**C 此前（§3/执行序）把 decide 的风险兜底称"退路"、与 emission"并重"——这是降格，错。** 更正：

**裁决不是 `argmax P(S)`（比哪个值高），是期望损失最小化（[[DBN-Zone-Room]] §8 已 ground truth，line 173）：**

$$\text{fire} \iff P^F \cdot C_{FN}(\text{risk}) > (1-P^F)\cdot C_{FP}$$

**含义重述（这是 C 对 A 命题的接受 + 自我更正）：**

1. **二义性不靠"把某个值算得更高"消解，靠"代价不对称"裁决。** cd2b 摔倒段证据稀薄 → $P^F$ 可能就卡 0.4、与 SBed 0.45 纠缠。按"哪个值高"→ SBed 赢 → 漏。按期望损失 → 独居老人 $C_{FN}$ 巨大 → 0.4 的 $P^F$ 已压过 0.45 SBed → fire。**不需要 $P^F$ 赢，只需期望代价翻转。**

2. **survival 的"慢=高危→久静上抬 fall"是同一精神的局部体现**：它不在说"久静更可能是摔"（概率），是"久静的漏报代价更高所以更该报"（风险）。dwell 方向由风险偏好驱动，与 §8 裁决同源。

3. **裁决层定位更正——风险偏好裁决是主框架，不是退路：**
   - ❌ C 旧定位："decide 不可判兜底 = emission 失败时的退路，与 emission 并重"。
   - ✅ 更正："**期望损失裁决是整个裁决层的主框架，恒在；emission/dwell 是给它供 $P^F$ 的证据层。** 即使 $P^F$ 因证据稀薄算不高，风险裁决仍在高危独居老人身上 fire。cd2b 不是'$P^F$ 够高所以报'，是'$P^F$ × 巨大 $C_{FN}$ > 误报代价所以报'。"
   - emission 位置似然（§5）、dwell 方向（§6/本节）都**降为风险裁决的输入**，不是与裁决并列的解。**cd2b 的终解在 decide 的期望损失，emission/dwell 只是供料。**

## 四、对 A 两问的一句话答复

(1) **时长阈值（数值）= 标定，随 oracle 调无虞；dwell 证据的方向（符号）由 cell 容忍属性决定 = 框架，且 cell 容忍判定可靠性是框架级前提。**
(2) **DBN 裁决是期望损失最小化（$P^F C_{FN} > (1-P^F)C_{FP}$）非 argmax——二义性靠代价不对称裁决，不靠把值算高。C 据此自我更正：风险偏好裁决是裁决层主框架（恒在），emission/dwell/位置似然全是供 $P^F$ 的证据层，cd2b 终解在 decide 期望损失。** [[DBN-Zone-Room]] §8 已 ground truth，C 此前的"退路/并重"定位降格，撤回。


---

# §8 增补（2026-06-14）：风险标定依赖项 — $C_{FN}(\text{risk})$ 代价函数（裁决主框架下唯一悬空量）

> §7 把 cd2b 终解定到 decide 的期望损失 $P^F C_{FN}(\text{risk}) > (1-P^F)C_{FP}$。$C_{FN}(\text{risk})$ 是这个主框架下**唯一悬空的量**——[[DBN-Zone-Room]] §8 只给定性"独居→$C_{FN}$ 大"，无标定。本节按 C 的实证纪律，给出**复用源 + 落差 + 标定路径**（类比 B 的 S1 标定依赖项）。

## 一、关键发现：现有 risk_evaluator 是"分档"，§8 要的是"代价加权"——风险因子同源，消费方式不同

`risk_evaluator.go` 已有完整风险分层（`owl-common/card/risk_thresholds.go`）：

| room | Day standing | RiskTime standing | alone | multi |
|---|---|---|---|---|
| Bathroom | 8/15 | 5/8（夜更敏感） | day 30/45 / night 20/30 | 清零 |
| Default | 8/15 | 8/15 | — | 降一档 |
| Kitchen | 12/18 | 8/15 | — | 降一档 |

RiskTime 默认 21:00–08:00。产出 **离散三档** `RiskNormal/Attention/Risk`，机制是**阈值穿越**（`standingMin ≥ riskMin → RiskRisk`）。

**落差**：§8 的 $C_{FN}$ 是**乘进期望损失不等式的连续代价权重**；现有 risk_evaluator 产出**离散 RiskLevel**。两者**风险因子同源**（独居 min、RiskTime 昼夜、multi 人数），但：
- 现有：风险因子 → 选离散档（分类）。
- §8 要：风险因子 → 算连续 $C_{FN}$ 倍数（代价加权裁决）。

→ **$C_{FN}$ 标定不需新造风险因子，复用 risk_evaluator 现有三因子即可；缺的是"因子 → 连续代价倍数"的映射，这个映射悬空，且它直接决定 cd2b 这类边缘 case（$P^F$ 卡 0.4）报不报。**

## 二、$C_{FN}(\text{risk})$ 标定依赖项（C 列为待定，不凭空取值）

| 风险因子 | 复用源 | $C_{FN}$ 方向 | 标定悬空点 |
|---|---|---|---|
| 独居持续 | `AloneContinuousMin` | 独居↑→$C_{FN}$↑（无人代为发现，漏报代价指数升） | 独居 min → 代价倍数曲线（线性？阈后跳变？） |
| 昼夜 | `IsRiskTime`(21–08) | 夜↑→$C_{FN}$↑（夜间无人巡视，与 survival 夜间短尾同向） | 夜间倍数 vs 白天基准 |
| 人数 | `TotalPeople` | multi↓→$C_{FN}$↓（有人在场可代发现，现有 multi 降档同理） | multi 时 $C_{FN}$ 折扣系数 |
| 失能史 | 〔现无通道〕 | 失能↑→$C_{FN}$↑（难自救，[radar 走速封顶"越弱越易摔难恢复"]同源） | 是否引入、PHI 边界 |

**标定纪律（同 §5 对 B 的实证回应精神）**：
- $C_{FP}$ 锚护士响应成本（一次误报 = 一次白跑），相对稳定，可设为 1（归一基准）。
- $C_{FN}$ 锚"漏一次真摔的后果"（独居老人卧地数小时→失温/横纹肌溶解/死亡），量级远大于 $C_{FP}$——**这正是"低 $P^F$ 也发"的代价依据**。
- 具体倍数曲线**须 oracle 标定**，但**初始量级**可锚现有 risk_evaluator 的档间比 + 临床跌倒后果数据（"长躺时间 vs 死亡率"曲线），不凭空取。

## 三、为什么这是框架级前提而非纯标定（与 §7(1) 呼应）

$C_{FN}$ 的**取值**是标定；但"**裁决用期望损失（$C_{FN}$ 加权）而非 RiskLevel 分档**"是框架。现有 risk_evaluator 的离散三档**不能直接喂 §8 不等式**——RiskRisk 不是一个代价数。所以：

- **框架要求**：decide 层必须有一个 $C_{FN}(\text{risk})$ 连续代价函数，消费 risk_evaluator 的因子，输出代价倍数，进期望损失裁决。**这个"代价函数存在且连续"是框架，不是标定。**
- **标定**：该函数的具体曲线/倍数。

→ **C 列此为风险标定依赖项：$C_{FN}(\text{risk})$ 的"存在与形态（连续代价加权）"是框架级前提（须在 decide 落地，不能用 RiskLevel 离散档替代）；其"取值曲线"是标定（oracle + 临床数据锚，不凭空）。这是裁决主框架（§7）唯一悬空量，与 §5 的 δ（emission 输入）、§7 的 cell 容忍可靠性（dwell 符号）并列为 Xsensorv1 的三个标定/框架前提。**

## 四、C 的三个前提一览（收口）

| 前提 | 轴 | 类别 | 悬空点 |
|---|---|---|---|
| $\delta_{\text{pad/floor}}$ | 空间（emission $g^{xy}$） | 标定（脆弱） | 跨 case 稳定性，δ≈0 退化 |
| cell 容忍属性可靠性 | 时间（dwell 符号） | **框架** | 误学翻符号→真摔被压；须权威源 gate |
| $C_{FN}(\text{risk})$ 代价函数 | 裁决（期望损失） | **框架**（形态）+ 标定（取值） | 连续代价函数须存在；倍数曲线待 oracle+临床锚 |

**三者正交**：空间证据（δ）、时间证据（dwell 符号）供 $P^F$；裁决（$C_{FN}$）决定 $P^F$ 多高才 fire。cd2b 三者各有贡献，但**终裁在 $C_{FN}$ 加权的期望损失**——这是 §7 的结论在标定层的落地。


---

# §9 增补（2026-06-14）：跨房 neighbor 耦合（第四轴）— A 方案的结构性遗漏

> A 提请：当前代码已有一个二义性消解机制——**雷达信号缺失、无明显概率偏好时，用整个 unit 的相邻（兄弟房）信号决断**（neighbor hand-off，K=耦合强度）。
> C 核查：**这根轴 `DBN-Zone-Room.md` 全文零提及**（搜 neighbor/跨房/hand-off 无命中）。A 的联合滤波只内化了 $O_b$（床）、$N_r$（**同房**多 track 人数）、ghost（realness），**跨房 neighbor 是被遗漏的第四轴**。

## 一、当前代码的 neighbor 机制（A 方案要消灭的"外挂"，却漏在清单外）

`belief_neighbor.go` + `belief_adapter.go`：本房丢轨（lostSeenMs，二义 lost-fall）后，查同 unit 兄弟房有无**新鲜有向** hand-off（邻房 EnterRoom 仍占用 ∨ InBed 翻转）→ 人挪去邻房 → 压本房 phantom fall。

| 参数 | 值 | 含义 |
|---|---|---|
| HandoffWindowMs | 60s | 邻房事件相对本房丢轨最大滞后（先走后到） |
| JitterMs | 5s | 反向余量 |
| **dampNbrFallen (K)** | **0.7** | 邻房占用 n→ SFallen ×(1−0.7n) 压 |
| gainNbrEmpty/Left | 3.0 / 2.0 | 邻房占用→本房 Empty/Left 抬 |
| 邻房 bed 可信度 | sleepad 0.9 / radar-only 0.2 | 接触式 vs 雷达式 |

**两个安全门**：
- **N-3 sole-resident 门**：仅单住户允许用邻房压本房 fall；多人/无人→邻房占用可能是另一人→归因不安全→**不压（漏报-safe）**。
- **有向极近窗**：邻房事件须**晚于**本房丢轨；stale"上次在哪"不算 hit（证不了此刻在哪，人可能穿盲区真摔）。

**机制形态 = 外挂 gate**：`neighborHandoff` 在 belief **外**算出 neighborResult → `neighborToObs` 包成 `ObsNeighbor` → 喂进似然 `1−dampNbrFallen·n` 压 SFallen。

## 二、问题：这是 A "删硬外挂"清单漏掉的同类病

A 方案核心论证（[[DBN-Zone-Room]] §2-§3）：$O_b$ 做成外部观测 → **第一步不确定性在第二步丢失，DBN 无法质疑硬结论** → 故内化为隐变量。

**neighbor 犯的是同一个病**：`neighborHandoff` 在 belief 外算好 neighborResult（含 sole-resident 门的硬判定 + max 合并的 conf），belief 只收到一个 `ObsNeighbor.Value/Conf`，**无法质疑邻房归因是否可信**（人到底挪去邻房了，还是邻房那个占用是另一人/陈旧事件）。A 把 $O_b$ 内化了，却把**结构同型的 neighbor 留作外挂**——A 的删除清单（§6 删硬 $O_b$/止血/zone 硬交接）**漏了它**。

## 三、neighbor 的 sole-resident 门与 §8 期望损失裁决未对齐（同 risk_evaluator 的病）

N-3 sole-resident 门是**离散 gate**：`rc != 1`（非单住户）→ 邻房耦合整个 OFF。但 §8 的裁决是**连续期望损失**。两者都在处理"多人时归因不安全"，消费方式却不同：

- neighbor N-3：多人 → **硬 OFF**（二值）。
- §8：多人 → $C_{FN}$ 折扣（连续，多人时漏报代价降但不归零，因仍可能没人注意到摔倒）。

**这与 §8 发现的 risk_evaluator"离散档 vs 连续代价"是同一个病的第三处复现**（§7 dwell 符号、§8 risk_evaluator、本节 neighbor 门）。三处都把"风险/归因"做成离散 gate，而 DBN 终极（[[fall_detection_risk_stratified_design]]）要连续代价加权。

## 四、C 给 Xsensorv1 的处理建议：neighbor 内化为"跨房拓扑轴"，与 §10 ghost 轴正交同构

A 的 §10 已给出"换轴"范式：$N_r$/ghost 用"跨 track $\rho$ 耦合，同一滤波形式"。**neighbor 应同构地内化为第四轴**：

```
床轴   B^j          接触证据 → 床占用（A 已内化）
人数轴 S^(i)         同房多 track → N_r（A 已内化）
ghost  realness T^(i) 跨 track ρ → P(ghost)（A 已内化）
跨房   neighbor      兄弟房 hand-off ρ_xroom → 本房 S 的 {Left, Empty} vs {Fallen} 路由（★A 遗漏，C 补）
```

- **内化方式**：邻房占用不再外挂成 `ObsNeighbor` 标量，而是作为本房 $S$ 转移 $T_S$ 的**跨房门控发射**——`Real-present(本房丢轨) → {Left, 邻房 Real-present}` 的转移概率由 ρ_xroom（hand-off 新鲜度×有向性）驱动，与 §10 的 realness ρ 同源（都是"跨实体共现"耦合）。belief 据此**联合推断**"人挪去邻房 vs 本房真摔"，而非吃一个算好的 ObsNeighbor。
- **sole-resident 门改连续**：不再 `rc!=1` 硬 OFF，而是邻房归因的 ρ_xroom **按 resident 数衰减**（单住户 ρ 强、多住户 ρ 弱但不归零）+ 进 §8 的 $C_{FN}$（多住户时邻房占用压 fall 的"信用"降，但漏报代价仍在）。
- **K（dampNbrFallen=0.7）去向**：内化后不再是似然层的固定 damp 系数，而是 ρ_xroom 驱动的转移概率——K 从"标定常数"变成"由 hand-off 证据涌现的耦合强度"，与 §3 的 $\kappa$（同床）、§10 的 ρ（ghost）同为"几何/事件共现耦合"家族。

## 五、收口：C 的轴从三补到四

| 轴 | 证据 | 解决的二义性 | A 状态 |
|---|---|---|---|
| 空间 δ（emission $g^{xy}$） | 雷达 XY | 床边摔 vs 垫上躺 | ✅ 内化 |
| 时间（dwell 符号） | 久静 × cell 容忍 | 卧地 vs 久坐 | ✅ 内化（survival） |
| 裁决（$C_{FN}$） | 风险因子 | 低 $P^F$ 报不报 | 🟡 形态待定（§8） |
| **跨房 neighbor（ρ_xroom）** | **兄弟房 hand-off** | **人挪去邻房 vs 本房真摔** | **🔴 A 遗漏（本节）** |

**neighbor 是 cd2b 类"本房雷达缺失 + 无概率偏好"二义性的现役解（代码已用），但 A 的联合滤波方案未把它内化、且其 sole-resident 门与 §8 风险裁决未对齐。C 建议：内化为第四轴（跨房拓扑），ρ_xroom 与 §10 ghost ρ 同构；sole-resident 门从离散 OFF 改连续衰减 + 进 $C_{FN}$。这是 B（只审 DBN-Zone-Room 方程）结构上看不到的——因为 neighbor 根本不在那份文档里，只在代码里。**


---

# §10 增补（2026-06-14）：C 对评审组 B 回击的回应 — 认账 / 辩正 / 补 B4

> B 对 C §5–§9 回击，提出"C 三处过度宣称 + 漏 B4"。C 逐条回应：**认三条、辩一条（有分寸）、补 B4**。诚实优先于护盘。

## 一、C 认账（B 说得对，C 更正）

### B-反1：§6 dwell"独占发现"不成立 —— **认，且 C §7 已撤回，此处再钉死**
`fallLRFromDwell` 按 `toleranceMult` 分流上抬/下压**是现有代码已实现的机制**，A 方案继承它。C §6 措辞为"C 独占发现"**错**；C §7 已撤回，但撤回不够干净。**最终定性：cell 容忍属性 gate dwell 方向 = 框架命题（成立，§7(1) 已立），但此机制非 C 发现，是代码已有、C 确认其对 cd2b 非容忍床沿生效。删除"独占"宣称。**

### B-反3：§8 不是"发现连续优于离散" —— **认，措辞夸大**
`DBN-Zone-Room.md §8` 的 fire 条件**本就写了** $P^F C_{FN} > (1-P^F)C_{FP}$ 连续形式——**设计文档先于代码写了连续裁决**。C 的真实贡献是**识别 `risk_evaluator.go` 离散三档喂不进这个已有的连续不等式**（代码↔文档落差），不是"发现连续裁决"。**更正：C §7/§8 措辞"DBN 终极要连续代价非 argmax"应表述为"识别现有 risk_evaluator 离散档与 §8 已有连续裁决的落差"，贡献是落差识别非概念发现。**

### B-漏：C 漏了 B4 —— **认，且诊断 C 自己的方法论盲区**
C §5 逐条回应 B1/B2/B3/C3，**独漏 B4**（$\Psi$ mixture-vs-product）。B 的诊断准：**C 漏 B4 因 fixture 是单床（cd2b 一张床），$|\mathcal B|=1$ 时 mixture 与 product 退化等价，未触发差异**。这暴露 **C 实证路线的结构性盲区：单 case fixture 测不到多床，凡需 $|\mathcal B|\ge2$ 才显现的架构选择，C 的"fixture 实证"方法天然漏。** 补回应见第三节。

## 二、C 辩正（B 事实对，但评判标准错位）

### B-反2：§9 neighbor"只给方向声明、方程密度不够" —— **事实认，标准辩**
B 事实对：C 对 neighbor 内化只给方向（ρ_xroom 与 ghost ρ 同构），未给转移矩阵/ρ_xroom 计算/正交接口，对比 A 对 $O_b$ 给了 §2–§7 六节方程。

**但 B 的评判标准错位**（git 核实）：`DBN-Zone-Room.md` 的 $O_b$ 六节方程是**项目组 A（sady37, commit 26da3dd）写的，非评审组写的**。**C 是评审组，职责是"发现 A 漏了第四轴"，不是"替 A 把第四轴六节方程写完"**——要求评审给出项目组级方程密度，是要评审做项目组的活。

**分寸**：B 提醒的"方向声明 ≠ 可进实现"**对**。故 C 更正定性——**"neighbor 第四轴遗漏"是成立的评审发现（A 已接受）；"内化为联合滤波轴"是 C 指出的方向，但完整方程（ρ_xroom 计算、$T_S$ 跨房门控转移、与 §10 正交接口）是 A 待补的设计开口，C 不冒充已完成方案。** 球正确地踢回 A——这正是评审与项目组的职责边界。

## 三、C 补 B4：多床 $\Psi$ mixture vs product —— 用风险目的给 B 给不出的角度

B 在设计层指出 mixture（加权和）vs product（乘积）是架构选择、影响多床联合相容性。C 补一个 **B 的纯设计审查给不出的角度——从"发现风险"目的判选择**：

- **mixture（$\Psi=\sum_j a_j\tilde\psi_j$，A 现行）物理直觉**：人**只在一张床**上，按归属 $a_j$ 软分配。
- **product（$\Psi=\prod_j\tilde\psi_j$）物理直觉**：人对每张床**独立相容**。

**C 的风险目的判据**：养老院多床房，"人只在一张床"是物理事实（一人不能同时占两床）→ mixture 语义**物理正确**。但关键在 **Fallen 行的风险后果**：
- mixture 下，人摔在床 A 旁（$a_A$ 高），床 B 的 $\tilde\psi_B$ 以低权 $a_B$ 进 $\Psi$ —— 床 B 的"vac+F 相容=1"**不会稀释**床 A 的"occ+F=$\varepsilon_{art}$"压制 → **F 信念由归属床主导，正确**。
- product 下，床 B 的 $\tilde\psi_B(F,\text{vac})=1$ 会**乘进** → 若床 B 恰 vac，product 把 F 通道乘开，**可能在归属床 A occ（压 F）时被非归属床 B vac 错误放行 F** → 多床房 phantom fall 风险。

→ **C 对 B4 的结论：mixture 不仅物理正确（人只占一床），且在 Fallen 行的风险传播上更安全（非归属床不会错误放行 F）。product 在多床房有 phantom fall 风险。建议 A 在 §4 显式声明选 mixture 的理由 = 物理（单床占用）+ 风险（Fallen 行非归属床不污染）。** 这是 C 用"风险后果"补 B 的"语义选择"——B 指出选择存在，C 给出该选哪个及风险依据。

## 四、修正后的 C 自评（对 B 回击的净结算）

| C 原宣称 | B 回击 | C 最终定性 |
|---|---|---|
| §5 B1/B2/B3 fixture 落值 | B 认 | **成立**（C 比 B 多一手实证） |
| §6 dwell 冲突"独占发现" | 代码已有 | **撤回独占**；框架命题保留 |
| §7/§8 "发现连续裁决" | §8 已写连续 | **更正为"识别代码↔文档落差"** |
| §9 neighbor 第四轴遗漏 | 真发现但方程不全 | **遗漏发现成立；内化方程是 A 待补开口（评审边界）** |
| §9 三处"离散 gate"串联 | B 认（B 没串起来） | **成立**（C 跨 dwell/risk_eval/neighbor 识别同一模式） |
| （漏 B4） | C 漏 | **补**（mixture 物理+风险双正确，见第三节） |

**净结算**：C 的**实证落值**（B1/B2/B3）、**neighbor 遗漏发现**、**三处离散 gate 串联**三项是 B 结构上做不到的真增量；**§6 独占宣称、§8 措辞、漏 B4** 三处 C 认账更正。**C 与 B 的真分工确认：B 审"方程内部写对没有"（C 漏的 B4 是其强项），C 审"方案押对没有 + 漏了什么轴 + 目的对不对"（B 看不到的 neighbor/落差/风险目的是其强项）。两卷互补，各有对方结构上做不到的盲区——这正是竞争式三卷的价值。**


---

# §11 增补（2026-06-14）：C 自我更正 §10 的 B4 风险因果方向（A+B 推导推翻 C 原论证）

> A+B 共识用代数推导推翻了 C §10 第三节对 B4 的风险论证。**C 认错并更正——方向反了。**

## C §10 原论证（错）

C §10 第三节称："mixture 更安全，product 在多床房有 phantom fall（FP）风险——非归属空床的 $\tilde\psi(F,\text{vac})=1$ 乘进会错误放行 F。"

## A+B 的代数推翻（对）

人在两床间（$a_A\approx a_B\approx0.5$），床 A occ、床 B vac：

| | product | mixture |
|---|---|---|
| $\Psi$ | $\tilde\psi_A(F,\text{occ})\cdot\tilde\psi_B(F,\text{vac})\approx\varepsilon_{art}\cdot1=\varepsilon_{art}$ | $0.5\,\varepsilon_{art}+0.5\cdot1\approx0.5$ |
| 对 F | **压死** | **放行** |

**结论反转**：product 在任一床 occ 时**更激进压 F**（任何一个 $\varepsilon_{art}$ 乘进去 F 就没了）；mixture **反而更宽容**。故：

- C 原说"product → phantom fall（FP）"**方向错**。
- 真正的风险是 **product 在 ambiguous 床位把真摔也压死（FN）**。
- 选 mixture 的理由是 **FN-safe（真摔不被非归属 occ 床误压死）**，**不是**"防 phantom fall"。

## C 更正

**§10 第三节 + §10 第四节表格"补 B4"行的风险依据，全部更正为**：

> **mixture 选择的理由 = FN-safe**：product 下任一床 occ 即把 $\varepsilon_{art}$ 乘进 $\Psi$ 压死 F，ambiguous 床位（$a_A\approx a_B$）的真摔会被非归属 occ 床误压成 FN；mixture 按归属软分配，ambiguous 时不被单床压死，真摔仍能浮出。物理理由（人只占一床）不变；风险理由从"防 phantom fall（FP）"**更正为"防真摔被压死（FN-safe）"**。

## C 的方法论反省

此错暴露 **C 实证路线的第二个盲区（继 §10 的单床漏 B4 之后）**：C 凭定性直觉推 $\Psi$ 的乘/和差异，**未代数地算 $a_A=a_B=0.5$ 中间态**——纯代数推导是 B 设计层审查的强项，C 在此吃亏。**确认 B 在"方程内部代数正确性"上对 C 有不可替代的制衡。这正是竞争式三卷的价值：C 的 fixture 实证防 A 押错方向，B 的代数审查防 C 推错因果。**

→ **net：DBN-Zone-Room §E 应作 FN-safe（A 统一文档时改）；本 C 卷 §10 B4 论证作废，以本节 §11 为准。**


---

# §12 增补 — 阶段 1 骨架验收规格（原 acceptance 文件并入）


> **角色边界（C 自我更正）**：production code（含阶段 1 骨架）执笔归项目组 A；C 定验收标准 + 独立审。
> C 此前一稿曾代写 joint/bed_axis/filter——**作废为执笔产物**，仅可作 A 的实现参考；正式骨架由 A 执笔、B/C 审。
> 本文件 = C 交给 A 的**验收规格**（断言级，A 实现后 B/C 据此审真 code）。

---

## 验收前提（不验则骨架不可信）

阶段 1 骨架 = 纯联合滤波，Ψ/Φ 中性占位（=1），Correct 为恒等。验**数值正确性**非行为。与 δ/neighbor 零依赖。

---

## T1 · 归一化守恒 Σα=1

**断言**：任意 `numBeds ∈ {0,1,2,3}`，初始 α 及连续 ≥50 步 Predict 后，`Σ_i α[i]` 与 1 的偏差 < 1e-9。

**构造**：`NewFilter(DefaultModel(), nb)`；online 全 true；循环 Predict(1) ×50，每步后查 ΣΑ。

**通过判据**：`|ΣΑ − 1| < 1e-9` 恒成立（含初始态）。失败 = 因子化 Predict 或 normalize 有质量泄漏。

---

## T2 · 单床退化基数 + numBeds 显式持有（B1）

**断言**：联合空间基数 = `numStates · 2^numBeds`：

| numBeds | 期望 size |
|---|---|
| 0 | 9（回 S 轴单链） |
| 1 | 18 |
| 2 | 36 |
| 3 | 72 |

**附加断言（B1 闭合）**：`Filter.NumBeds()`（或等价显式字段）必须返回构造时的 numBeds——**床数显式持有，不靠隐式推断**。

**退化正确性**：`numBeds=0` 时 `MarginalS(α)` 必须逐分量等于原单实体 `Belief` 行为（初始 == model.Prior，偏差 < 1e-9）。

---

## T3 · P-5 maxBeds 超界硬断言

**断言**：`maxBeds=3`（养老院房 ≤3–4 床上界）。

- `numBeds ∈ {-1, 4, 99}` → 构造（NewJointSpace/NewFilter）必须 **panic**（不静默膨胀）。
- `numBeds == maxBeds`（=3）→ **不 panic**（边界合法）。

**通过判据**：超界 recover 到非 nil；边界值无 panic。

---

## T4 · D-2 核心：ε≪λ 复现 30s staleness（cd2b 治本不变量）

**这是阶段 1 最关键验收——"一个不等式 ε≪λ 替代所有 staleness/TTL"的实证，也是 cd2b 漏报根因（陈旧 occ 永不衰减）在新架构治本的直接证据。**

**喂数序列**：
1. 单床（numBeds=1）。构造初始 occ 主导：质量集中到 `(SBed, bmask=1)`（人在床睡，床 occ），归一化。
2. 验初始 `P(B^0=occ) = MarginalB(α,0) ≥ 0.99`。
3. **sleepad 离线**：`online = {false}`。逐步 Predict（1Hz，1 步 = 1s），共 120 步，每步记 `P(occ)`。

**期望衰减曲线**（§C 单向泄漏核 K^unobs，λ 半衰期 ≈ ln2/λ）：
- `P(occ)` 单调下降（occ→vac=λ 泄漏，vac→occ=0 吸收 → 单向）。
- **跌破 0.5 的步数 `crossStep` 必须 ∈ (0, 30]**——陈旧 occ 在原 30s staleness 窗内蒸发到 vac 主导。
- λ 形态约束：默认 λ 使半衰期 < 离线判定窗（~30s）；occ 主导跌破 0.5 约 1 个半衰期。**值是标定（feedback-p6C §5），但"< 30 步跌破"是形态验收，必过。**

**在线对照组（证明蒸发归因离线核，非数值漂移）**：
- 同初始 occ 主导，`online = {true}`，Predict ×30。
- 期望 `P(occ)` 仍 ≥ 0.5（K^obs 高自持，ε 小，不蒸发）。
- 失败 = ε 太大（在线核也蒸发）→ ε≪λ 不成立。

**通过判据**：离线 `crossStep ∈ (0,30]` **且** 在线 30 步后 `P(occ) ≥ 0.5`。两者缺一即 D-2 不通过。

---

## T5 · §C vac 吸收态（单向泄漏正确性）

**断言**：空房（`(SEmpty, bmask=0)`，床 vac）+ 离线（online=false）+ Predict ×120 后：
- `P(B^0=occ) ≤ 0.05`（vac→occ=0 吸收 → 空房离线**不被弛豫成 0.5 伪占用**）。

**为什么单列**：这是 A 坚持"弃矩阵、用箭头记法"要守的核心——若 K^unobs 方向写反（occ 吸收/vac 泄漏），此断言失败 = cd2b 漏报原样重演。**T5 是 §C 箭头记法方向正确性的守门测试。**

---

## 验收总判据

T1–T5 全过 = 阶段 1 骨架数值正确，可进阶段 2（emission/coupling/transition）。
任一不过 = 骨架有数值缺陷，阶段 2 不得动。

**审查归属**：A 执笔骨架 → B/C 据本规格独立审真 code（C 不审自写代码，独立性保住）。


---

# §13 增补 — C 对 A 阶段 1 骨架真 code 的独立审查（第 1 轮，原 skeleton-review 并入）


> 审查对象：commit `6baa31f`（A 执笔，`tools/Xsensorv1/`）。
> 审查方式：pull 真 code，对照 [[feedback-p6C-acceptance]] T1–T5 + 独立查规格未覆盖的隐患。
> 角色：C 审 A 写的 code（C 未参与执笔，独立性成立）。
> 工具链注：容器限 go 1.22（go.dev 不在白名单），临时降版编译验证，未改 A 的 go.mod（1.25.0）。

---

## 总判：**骨架通过，质量超出验收要求；一个结构问题须修，两个小点建议。**

T1–T5 全过（实测，非纸面）。A 的实现**优于 C 的参考稿**——C 参考是线性域，A 自己上了 log 域 + LogSumExp，满足 DBN-Zone-Room §7 数值稳定要求。核心不变量（ε≪λ 复现 staleness、§C vac 吸收）实测兑现。

---

## 一、验收结果（T1–T5 实测）

| 验收 | 结果 | 实测 |
|---|---|---|
| T1 Σα=1 守恒 | ✅ | log 域 LogNormalize，nb=0/1/2/3 多步后 Σexp(α)=1 |
| T2 单床退化 + numBeds 字段 | ✅ | size 9/18/36/72；`NumBeds()` 显式持有（B1 闭合） |
| T3 maxBeds panic | ✅ | -1/4/99 panic，maxBeds=3 合法 |
| T4 D-2 ε≪λ staleness | ✅ | **离线后陈旧 occ 14 步(≈14s)蒸发到 vac 主导，< 30s 窗**；在线对照不蒸发 |
| T5 §C vac 吸收 | ✅ | 空房离线 120s 后 P(occ)≈0，不伪占用 |

**D-2 是关键验收**：ε=1e-2、λ=0.05，半衰期≈14s，陈旧 occ 在 30s 窗内蒸发——"一个不等式 ε≪λ 替代所有 staleness/TTL"实证兑现，cd2b 漏报根因（陈旧 occ 永不衰减）在新架构治本。

## 二、C 认可 A 超出要求的两处（记功，非挑刺）

1. **log 域 + LogSumExp（A 自加，C 参考无）**：`filter.go` 全程 log 域，`LogSumExp` 防下溢（跳过 <-50 项防 subnormal）、`logP(p≤0)=-Inf` 边界干净。这是 §7 要求、C 验收规格**未强制**但 A 主动做对的——为阶段 2 的极小 $\varepsilon_{art}$（§E）log 域诊断预留了数值空间。**A 比 ground truth 要求更前一步。**
2. **logKernel 预存**：T_B 核 log 化预存（`makeLogKernel`），Predict 内层不重复 log，性能正确。

## 三、🔴 必修：两套 belief 包重复提交（结构问题）

commit `6baa31f` 把 belief 包提交进**两个路径**，内容不同：
- `internal/belief/`（6 文件）
- `internal/roomengine/belief/`（6 文件，commit stat 列为正本）

**问题**：①包名都是 `belief`，并存造成 import 歧义、审查对象不明；②两套内容不同（疑早期试写 vs 定稿），后续维护会 drift；③C 实测的是 `roomengine/belief/`（与 Tsensor 路径一致），`internal/belief/` 来源/用途不明。

**C 要求**：确认正本（建议 `roomengine/belief/`，与 Tsensor `internal/roomengine/belief/` 路径一致、便于阶段 4 对照 diff），**删除另一套**。这是 D-1 baseline 对照纪律的延伸——审查对象必须唯一。

## 四、🟡 建议（不阻塞，阶段 2 前处理）

1. **`bedOnline` 长度与 numBeds 的契约未断言**：`Predict(online bedOnline)` 内 `if j < len(online)` 容错，但 `len(online) != numBeds` 时静默用 false（离线）。建议加断言或文档明确契约——否则阶段 2 adapter 传错长度会静默当离线，掩盖 bug。
2. **参数是形态占位，须在阶段 2 标定前显式标记**：`defaultBedAxisParams` 的 ε=1e-2/λ=0.05 是 feedback-p6C §5 的单 case 量级锚，**非标定值**。注释已写，但建议 A 在进 emission/coupling 前确认这些不被下游当权威值硬编码（铁律：定形态、不定参数）。

## 五、阶段 2 放行判据（C 立场）

- 🔴 三（两套包）修了 → 审查对象唯一 → 骨架正式通过。
- 阶段 2（emission Φ/§D gate、Ψ/§E mixture、transition/§A neighbor）**仍按硬序**：neighbor transition 等 A 的 ρ_xroom 方程；emission/coupling 的非 neighbor 部分可在骨架修后起。

**C 净判：A 骨架数值正确、质量超 C 参考；修掉两套包重复即正式通过。B 可独立复审。**


---

# §14 增补（第 2 轮审核）— A 修复后复审 + 与 B 骨架审查对照

> 第 2 轮：A 修了 C §13 的必修项；B 也出了独立骨架审查。C 复审 A 修复 + 对照 B。

## 一、C 必修项已闭合 ✅

C §13 🔴 必修「两套 belief 包重复」→ A commit `a092bb7 fix(xsensorv1): 删重复 belief 包（C 骨架审查发现）` 已修。

**C 复审实测**：仓库只剩一套正本 `internal/roomengine/belief/`（C 建议保留的、与 Tsensor 路径一致的那套），删包后 T1–T5 仍全过（重跑确认，删包未碰坏正本）。**审查对象唯一性恢复，骨架正式通过。**

## 二、与 B 骨架审查（`feedback-p6B-skeleton-review`）对照

**完全一致项**：T1–T5 全过、log 域+LogSumExp 满足 §7（双方都认这是 A 超 C 参考的优点）、两套包重复（B 明确归功「C 发现、A 已修」）、bedOnline 契约未断言（C 标、B 认同）。

**B 抓到 C 漏的一项（C 认）**：
- B §三：**Predict 内层 `Σ_j logK` 重复计算**——此和只依赖 (bFrom,bTo)、不依赖 S，却嵌在最内层 S 循环重复算。B 给重构方案（预存 bmaskN×bmaskN 的 logTB 表，提到 S 循环外，复杂度从 O(numBeds·nBC²) 降 O(nBC²)）。
- **C 认账**：C §13 盯方程正确性 + 不变量，漏了这个**计算结构冗余**——纯代码结构/复杂度是 B 的强项（呼应 §11：B 在代数/结构上对 C 有制衡）。不阻塞（|B|≤3 ~15k ops 可忽略），但 B 对：T_B 因子化是 Predict 第一步，不该嵌最内层。阶段 2 重构。

**C 抓到 B 没提的一项**：
- C §13 🟡②「参数形态占位须防下游硬编码成权威值」（铁律：定形态不定参数）——B 未提。这是目的/纪律层，B 盯方程实现没盯到参数标定纪律。

## 三、第 2 轮净结算

| 项 | B | C |
|---|---|---|
| T1–T5 + log 域 + §7 | ✅ | ✅（一致） |
| 两套包重复 | 认 C 发现 | C 发现、A 已修 ✅ |
| bedOnline 契约 | 认同 C | C 标 |
| Predict logB 重复计算 | **B 抓** | C 漏，认 |
| 参数形态占位纪律 | 未提 | **C 抓** |

**双方对骨架结论一致：通过，放行阶段 2。** 各补一项对方盲区（C 漏 B 的计算结构、B 漏 C 的参数纪律）——三卷互补再次兑现。

## 四、阶段 2 放行（C 立场，与 B 一致）

骨架正式通过（必修已闭合）。阶段 2 仍按硬序：
- emission Φ/§D HRRR gate、coupling Ψ/§E mixture：可起（ground truth 已定形态）。
- transition/§A neighbor 有向门控：**等 A 的 ρ_xroom 完整方程**（路二，A 按住中）。
- 阶段 2 前 A 处理两项：B 的 Predict 重构 + bedOnline 契约断言（双方共识）。


---

# §15 增补（第 3 轮审核）— A 阶段 2 emission/coupling/neighbor 独立审 + 与 B 对照

> 第 3 轮：A 交阶段 2（emission §5 Φ / coupling §2·§3·§4 / neighbor §A 方程 / 🟡 两项修复）。C 独立 pull、独立跑、独立审实现忠实性，再对照 B 第 2 轮 phase2 审查。

## 一、C 独立验证（不照搬 B 的数）

C 临时降版编译、独立跑全 15 测试：**T1–T5（骨架回归无退化）+ E1–E5（emission）+ C1–C5（coupling）全过。** 两个核心数 C 独立跑出、与 B 报告一致：

- **cd2b 离线态 P(Fallen)=0.9998 > P(AtBed)=0.0000**——不靠 sleepad/HRRR，δ 几何主解实证兑现（emission_test.go:112）。这是整条线 cd2b 治本的最终验收，C 独立确认属实。
- **§E mixture FN-safe：Ψ(F)=0.55 存活 vs product=0.01 压死（55×）**（coupling_test.go:82）。C §11 自我更正的 FN-safe 因果方向，A 实现里方向正确。

## 二、C 独立审实现忠实性（逐条核 ground truth，非看测试数）

测试是 A 自写，C 不止看 PASS，独立读实现判忠实：

| ground truth | A 实现 | C 独立核 |
|---|---|---|
| §E Ψ=Σ_j a_j ψ̃_j+a_∅（mixture 非 product） | coupling.go:100 加权和 | ✅ 非乘积；注释明写"product 任一床 occ 压死 F=漏报"，FN-safe 方向对 |
| §E Fallen 行不被 κ 覆盖 | psiPhys:80 物理常量 | ✅ κ 退火在 psiTilde，不碰 Fallen 行 |
| §D absent 须 sleepad-online gate | emission.go:108 `else if VitalSourceOnline` | ✅ radar 自身 absent 零信息不否决（C §13/B3 实证约束落地） |
| §5 Φ 分轴（接触→B/雷达→S） | contactLogB/radarLogS | ✅ 接触只依赖 bmask 位、雷达只依赖 S |
| δ 主解 floor-strip | emission.go:119 `SFallen:lDelta, SBed:1/lDelta` | ✅ 偏 F 压 AtBed，δ≫0 离线可判=cd2b 0.9998 来源 |
| §3 κ EMA 无 max | coupling.go:51 `(1-g)κ+g·m` | ✅ 可升可降，互活门控（仅 live 才更新） |

**C 净判：A 阶段 2 实现忠实于 ground truth，非仅测试通过——逐条方程映射正确。**

## 三、与 B 第 2 轮 phase2 审查对照

**本轮 B/C 完全一致，无分歧、无独占发现。** 双方独立审到同一结论：§5/§D/§4/§3/§2/§A 实现忠实、cd2b 0.9998、mixture 55×、🟡 两项（bedOnline panic + buildLogTBCol/Predict 重构）已修。

这与前几轮（C 有 fixture 独占、B 有代数独占）不同——**阶段 2 是直接的方程→代码映射，两卷独立审收敛到同一判断，正是"实现忠实"的最强信号：两个独立审查者都没找到偏差。**

## 四、covers max（C2）裁定确认

C2「`w_pose=covers(r,·)` 的点参数（多床取 max/最近/整体）」此前是 C 标的悬空项。A 选 `max_j covers`（B 第 2 轮裁定、ground truth §F C2 已记、edf5e61）。**C 立场：这是 C2 悬空项交 A 的正当选择，max 在单雷达多床下取"最能看清的那张床的覆盖"作 pose 权重，合理；C 标注接受，不反对。** 单 case fixture 是单床（max 退化为单值），多床 max 语义待阶段 4 多床 case 验证。

## 五、第 3 轮放行（C 立场）

阶段 2 通过。emission/coupling/neighbor 方程实现忠实、核心验收（cd2b 0.9998）兑现、🟡 修复确认。**与 B 一致放行。**

下一步（阶段 3/4）：
- 阶段 3 decide.go：§8 期望损失主框架（$P^F C_{FN} > (1-P^F)C_{FP}$）+ 不可判兜底。$C_{FN}(\text{risk})$ 代价函数形态须落地（C §8：框架级，连续非离散档），取值标定待 oracle。
- 阶段 4 probe + cd2b 三态对照（baseline `bd70194`，C D-1）+ 多床 case 验 §E mixture/covers max（补单 case 盲区，呼应 §10 C 实证路线多床短板）。


---

# §16 增补（neighbor 方程补审）— C 对 §A.1–§A.3 ρ_xroom 独立审

> C 欠审：§15 仍写「等 A 的 ρ_xroom 方程」，但 A 早在 `6018f11` 交付 §A.1–§A.3。C 独立读方程，逐点核 A 点名的三处。

## 一、§A.1 有向核 $w^{dir}$ 对 $\operatorname{sign}\Delta$ 不对称 — ✅ 成立，C 接受 A 对「同构」的校正

正/反向衰减常数不同（$\tau_j\ll\tau_h$：正向 60s 新鲜窗、反向仅 5s 容时钟噪声）→ $w^{dir}(\Delta)\ne w^{dir}(-\Delta)$，对 sign Δ 不对称。物理正确：先走后到（Δ>0）=合法 hand-off 给长窗；反向=抖动给极短余量。**C 接受 A 校正**：C 早先「与 ghost ρ 同构」过简；neighbor 带时序方向、不照搬 ghost 对称核。C §9「同构」措辞作废，以 §A.1 有向版为准。

## 二、§A.2 $T_S$ 跨房门控 — ✅ 成立，A 主动堵「耦合进 T vs §6」表面矛盾

Blind 行 $\to F$ 按 $\rho^{xr}$ 整流入 $\to L$；$\rho{=}0$ 行不变（lost-fall 本义保留=安全默认），$\rho{\to}1$ F 整流入 L（不造 phantom）。**关键审点**：耦合进 $T_S$ 是否违反 §6「耦合只在 Ψ」？A 论证：§6 针对床轴 B 与 S 同帧相容（防 B–S 双施）；neighbor 是 S 轴自身跨房转移先验、属 $T_S$ 本职、非 B–S 双施。**C 独立核：A 论证正确**——§6 防同一耦合在 T 与 Ψ 重复施加；ρ_xroom 只在 $T_S$ 施一次、不碰 B–S。A 提前堵得对。

## 三、§A.3 三接口 — ✅ 全成立

①ghost→neighbor（$q_{r'}$ 吃去 ghost 占用，ghost 不算落点，belief 可质疑归因=内化非外挂）；②不加隐维（$\rho^{xr}$ 是房间 $T_S$ 转移耦合，不进本房 J 基数，解 P-5 爆炸）；③同 census 双消费（$\eta(\text{rc})$ 与 $C_{FN}$ resident 数同源，多住户两层一致下拉，守 §B 分工）。

## 四、C 净判
**§A.1–§A.3 通过，无修改要求。** 有向性成立、$T_S$ 门控与 §6 不冲突、三接口守状态空间不爆+§B 分工。曲线参数留标定。C §15「等方程」更正：方程已落且已审通过。

---

# §17 增补 — C 对 A 阶段 3 三条立场的回应（decide 实现前对齐）

## 立场 1 不可判兜底不写独立分支（$\Lambda$ 绝不作 gate）— ✅ C 接受，A 更优雅
special-case 一个不可判分支 = 假装知道何时不可判，反违背 §8「诚实的不确定」。**追加验收：禁止 $\Lambda$ 作 gate；$\Lambda$ 仅 probe forensic。**

## 立场 2 $C_{FN}$ 只设保守 form-anchor — ✅ 与 C §8 一致
形态框架（连续/各因子单调/多人折扣有下限不归零）+ 取值留 oracle（[[fall_data_is_artificial_test]]）。

## 立场 3 cd2b 主解定位守住（decide 不改 cd2b 解）— ✅ C 认，A 主动防 C 早期误框
cd2b（δ≫0，E5=0.9998）已在 emission 解掉；decide 只对 δ≈0 边缘兜底。C §3/§7 已更正「cd2b 终解在 decide」误框，A 守住此定位。

## C 给阶段 3 decide 的验收点（D1–D6，实现前规格）

| # | 验收 | 判据 |
|---|---|---|
| D1 | 期望损失裁决 | `fire ⟺ P^F·C_FN>(1-P^F)·C_FP 持续≥T_hold`；非 argmax |
| D2 | $\Lambda$ 不作 gate | decide 路径不读 $\Lambda$ 做分支 |
| D3 | $C_{FN}$ 连续形态 | 连续非离散档；各风险因子单调；多住户折扣有正下限不归零 |
| D4 | $C_{FP}$ 归一 | $C_{FP}=1$ 基准 |
| D5 | cd2b 不回归 | decide 接上后 cd2b 离线态仍 fire |
| D6 | 取值非权威标注 | $C_{FN}$ 曲线标 form-anchor、留 oracle |

**C 放行 A 起阶段 3。** neighbor 方程已审通过（§16）；decide 三立场对齐；D1–D6 待实现后审。

---

# §18 增补（第 4 轮审核）— A 阶段 3 decide.go 独立审（对照 §17 D1–D6）

> 第 4 轮：A 交 decide.go（`444dab7`）。C 独立 pull、独立跑全测试、读 decide.go 逐条核 §17 的 D1–D6 验收（非看测试名）。

## 一、C 独立验证

全 20 测试独立跑通（含 5 个 decide：Sustain/RiskStratified/CostFlipNotArgmax/UnidentifiableNoSpecialBranch/ComputeLambda），cd2b 0.9998 无回归。

## 二、D1–D6 逐条核（读实现，非看测试）

| # | 验收 | decide.go 实现 | C 核 |
|---|---|---|---|
| D1 | 期望损失非 argmax | `margin=pF·cFN-(1-pF)·cFP; inst=margin>0`；tHold=90s | ✅ 正是 $P^F C_{FN}>(1-P^F)C_{FP}$，非 argmax |
| **D2** | **Λ 绝不作 gate** | `inst/fired` 只依赖 `margin`，**完全不读 lambda**；`Lambda/Identifiable` 仅写 Decision 结构作 forensic，注明「不参与 fire 决策」 | ✅ **A 立场①硬约束落实**：无 special-case 分支，不可判由同一不等式吸收 |
| **D3** | **$C_{FN}$ 连续/单调/多人不归零** | `cFN()`：独居连续增益饱和、夜×1.5、失能×1.5 全单调；多人 `disc=1/N` 连续，`if disc<peopleFloor{disc=0.3}` 下限不归零 | ✅ 连续非离散档；各因子单调；多住户折扣有正下限（呼应 §A.3③） |
| D4 | $C_{FP}$ 归一 | `cFP=1.0` 基准 | ✅ |
| D5 | cd2b 不回归 | emission 0.9998 未动，decide 只在 $P^F$ 上裁决 | ✅ 独立跑确认 |
| D6 | 取值非权威标注 | `decideParams`/`cFN` 注「标定锚非权威，留 oracle」+ 引 [[fall_data_is_artificial_test]] | ✅ |

**D2/D3 是 §17 三立场的硬约束（C 重点查），A 落实正确。** 尤其 D2：decide 路径不读 Λ，Λ 仅 probe forensic——A 立场①「special-case 不可判 = 假装知道何时不可判，反不诚实」在 code 里兑现。

## 三、关键立场测试 C 独立确认

- **DEC3 代价翻转非 argmax**：$P^F=0.4$（argmax 下输给假想 AtBed 0.45）独处高 $C_{FN}$ 仍 fire——「不需 $P^F$ 赢，只需代价翻转」（C §7/§B）实证。
- **DEC4 不可判无独立分支**：$\Lambda\to1$（全暗）高风险独处仍 fire，且 Λ 不 gate——A 立场① 实证。

## 四、C 记 A 一处超规格（记功）

`ComputeLambda` 用 LogSumExp 在 log 域算 $\Lambda=\exp(\text{LSE}(F)-\text{LSE}(AtBed))$，数值稳定、与 §7 log 域一致。C 的 D 规格未要求 Λ 数值实现细节，A 自己做对——延续阶段 1 起 A 在数值稳定性上主动超 C 参考的一贯。

## 五、第 4 轮净判

**decide.go 通过，D1–D6 全忠实，无修改要求。** 边界 case 干净（margin>0 严格不等保守、PeopleCount≤1 当独处、fireSince 断开复位防噪声）。

**阶段 3 完成。期望损失主框架（§B）+ $C_{FN}$ 连续代价（§8/C §8）+ Λ 不作 gate（A 立场①）三者在 code 落地。** 与 B 待对照（B 若已审则比对，未审则 C 先行）。

下一步阶段 4：probe + cd2b 三态对照（baseline `bd70194`，C D-1）+ **多床 case 验 §E mixture/covers max/§A neighbor**——补 C 实证路线单床盲区（§10 自认短板）。$C_{FN}$ 曲线参数留 oracle（D6），非阶段 4 标定目标。


---

# §19 增补（第 5 轮审核）— A 阶段 4 belief 单元独立审（probe §9 + 多床 MB1–MB4）

> 第 5 轮：A 交阶段 4 belief 单元（`c1bdc2a`：probe.go + multibed_test.go）。C 独立 pull、独立跑全 24 测试、读多床构造与 probe 实现。**重点：多床 case 补 C §10 自认的实证路线单床盲区。**

## 一、C 独立验证

全 24 测试独立跑通（临时降版编译）。零回归（T1–T5 / E1–E5 / C1–C5 / D1–D6 全保持）。新增 4 项多床 + 1 项 probe 全过。

## 二、多床验收 — 补 C §10 单床盲区（本轮核心）

C §10/§11 反复自认：fixture 单床（cd2b 一张床），$|\mathcal B|=1$ 时 mixture/product 退化等价，**C 实证路线测不到多床**（当初漏 B4 即因此）。阶段 4 终于补上，C 独立核构造**真测到多床风险，非退化伪多床**：

| 单元 | 构造 | C 独立核 |
|---|---|---|
| **MB1 §E mixture FN-safe** | $|\mathcal B|=3$，bed0 陈旧 occ + bed1/2 vac + 人摔（SFallen, bmask=1） | ✅ **真多床 product 塌缩**：mixture Ψ(F)=0.69 存活 vs product=ε_art·1·1=0.01（69×）。单床触不到——bed1/2 的 vac 各贡献 1，product 乘起来仍被 bed0 的 ε_art 拖死=FN；mixture 按 a_j 软分配存活。**§E FN-safe 因果在多床实证（不再单床退化推测）** |
| **MB2 多床路由** | 雷达 g^xy=[1,0] 仅定位 bed0、仅 bed0 InBed，30 步 | ✅ P(B0)=1.0 >> P(B1)=0.113，S*=SBed——**多床证据不串床**，attachment a_j 按 g^xy 归属正确（单床无从验） |
| **MB3 covers max C2** | 异覆盖 [0.3,1.0] → geom0Covers 取 max=1.0 | ✅ ground truth §F C2 落地（A 选 max_j，C §15 接受） |
| **MB4 probe 快照** | Σα=1 + 边缘一致 | ✅ |

**C 立场更新**：§10「C 实证路线多床盲区」自认短板，阶段 4 MB1/MB2 已补——多床 mixture FN-safe（69×）+ 多床路由不串床均独立验证。**C 的单床局限在此闭合**（曲线参数仍留 oracle，但 §E/路由的"形态正确性"多床已证）。

## 三、probe §9 — 纯诊断确认（与 D2 Λ-不-gate 同精神）

`probe.go` C 独立核：`Snapshot` 只**读** filter/coupling（`f.Alpha()`、`copy(cp.kappa)`），**不写回、不被 decide/filter 决策路径引用**（grep 仅注释出现）。吐 α 全分量 + 边缘 S/B + P^F + Λ + κ + 裁决，供 Xsensorv1 vs Tsensor 逐帧 diff（baseline `bd70194`）。**纯 forensic，绝不参与推断**——与 §18 D2（Λ 不 gate）同精神：诊断暴露一切，不参与决策。

## 四、第 5 轮净判

**阶段 4 belief 单元通过，无修改要求。** 24 测试全绿（独立跑）、多床构造真实、probe 纯诊断。

**C 的实证路线单床盲区（§10 自认）在本轮闭合**——这是阶段 4 对 C 最大的意义：不是 A 又交一批测试，是 C 一直承认的"测不到多床"短板被 MB1/MB2 直接补掉，§E mixture FN-safe 从单床退化推测升为多床实证。

## 五、belief 包收口 + 下两个里程碑（C 立场）

belief 包（joint/bed_axis/filter/mm/coupling/emission/decide/probe + neighbor 方程）阶段 1–4 **B/C 全放行**。治本兑现：cd2b 离线 P(Fallen)=0.9998 δ 几何独立判定。

剩两个独立里程碑（belief 包外，C 同意 A 的裁定）：
- **集成**：建 adapter + 继承非 belief 脚手架（cell/track_manager/grid/stream）+ cd2b vs Tsensor 逐帧 diff（baseline `bd70194`，C D-1）。belief API 已齐待 wire。**C 验收点（预告）**：集成后 cd2b 端到端仍 fire（probe 逐帧对照 Tsensor，证 belief 替换未引入回归）。
- **neighbor 跨房 wiring**：依赖多房 filter 编排（§A 方程已审过，待 wire）。**C 验收点（预告）**：兄弟房 hand-off 真实 case 下 lost-fall 被正确整流入 Left（§A.2 ρ_xroom→1 路径），且无新鲜 hand-off 时不抑制（安全默认）。

$C_{FN}$ 曲线 / neighbor τ_h,τ_j,β 等参数留 oracle，非集成里程碑的标定目标（D6 纪律延续）。


---

# §20 增补（评审方交接确认）— belief 包收口，C 验收点钉死待重启

> 用户裁定：belief 包清晰交接、暂停，集成/neighbor 专门排期。A 已出 A 卷交接清单（`c8cb3ce`）。本节是 **C（评审方）的交接确认**：钉死 belief 验证状态 + 把两里程碑 C 验收点从「预告」转「待执行」，作为重启时 C 的依据。

## 一、C 对 belief 包的最终验证确认（交接前独立复跑）

C 交接前最后一次独立验证：`go test ./internal/roomengine/belief/` **24 测全绿**、`go vet` 干净。

| 维度 | C 确认状态 |
|---|---|
| 数值正确 | T1–T5：Σα=1、退化 9/18/36/72、maxBeds panic、ε≪λ staleness（14s<30s）、vac 吸收 |
| 发射 | E1–E5：Φ 分轴、§D HRRR gate、δ floor-strip、**cd2b 离线 P(Fallen)=0.9998**（治本兑现） |
| 耦合 | C1–C5：κ EMA 无 max、§E mixture FN-safe（单床 55×） |
| 裁决 | D1–D6：期望损失非 argmax、Λ 不 gate、C_FN 连续单调多人不归零、cd2b 不回归 |
| 多床 | MB1–MB4：**|B|=3 mixture 69×**、路由不串床、covers max、probe 快照——**C §10 单床盲区闭合** |
| 方程 | §A.1–§A.3 neighbor（有向核/T_S 门控/§10 接口）C §16 审过 |

**C 判：belief 数学层阶段 1–4 全放行，是干净的已验证黑盒。集成只需证「接入无回归」，不需重验 belief 内部。**

## 二、C 验收点钉死（重启时 C 据此审，从「预告」转「待执行」）

### 里程碑 1 集成 — C 验收点

- **AC-1（核心·无回归）**：集成后 cd2b 端到端**仍 fire**。判据：probe 逐帧对照 Tsensor（baseline `bd70194`，C D-1），P(Fallen) 轨迹在 cd2b 离线段达 fire 阈、与 belief 单元测一致（0.9998 量级）。**belief 是已验证黑盒，若端到端不 fire → 缺陷在 adapter/wire，不在 belief。** 这条归因边界是清晰交接的核心价值。
- **AC-2（adapter 译层忠实）**：adapter 译 raw→Observation/BedGeom/RiskContext 的字段语义正确。重点：BedReading 状态映射（InBed/LeftBed/NoReport）、g^xy 雷达定位算法、VitalSourceOnline 门（§D：radar 自身 absent≠否决，须独立 vital 源在线）、online[] 长度=numBeds 契约（C §13 标的）。
- **AC-3（边界守卫位置）**：`cFN()` 对 `alone<0`（adapter 时钟回拨）的守卫——**按规则「错误处理只在边界」落 adapter 侧、不进 cFN 内部**（B round3 建议 + B/C 共识）。C 验收：cFN 内部保持纯形态函数，守卫在 adapter。
- **AC-4（baseline 纯净）**：diff 对照基线必须是 `bd70194`（止血补丁 `7ffec9c` 之父），不可用打了补丁的版本——否则对照被旧 staleness 补丁污染（C D-1 纪律）。

### 里程碑 2 neighbor wiring — C 验收点

- **AC-5（路由正确）**：真实 hand-off case 下 lost-fall 正确整流入 Left（§A.2 ρ_xroom→1：fresh sleepad hand-off + 单住户）。
- **AC-6（安全默认·铁律）**：**无新鲜有向 hand-off 时不抑制 fall**（ρ_xroom=0 → Blind 行不变 → 照常 ramp Fallen）。stale「上次在哪」/多住户归因弱 → 不压。铁律 [[partial_monitoring_fall_suppression_law]]。**这条是 neighbor wiring 最危险处**：wire 错会让 neighbor 变成漏报源（早期外挂 ObsNeighbor 的病），C 重点审。
- **AC-7（去 ghost 落点）**：ρ_xroom 吃兄弟房**去 ghost 后**占用（§A.3 接口①），兄弟房一个 ghost 不算合法 hand-off 落点。
- **AC-8（不加隐维）**：neighbor wire 后本房 J 基数仍 9·2^|B|（§A.3 接口②，ρ_xroom 是房间 T_S 转移耦合、吃兄弟房 belief 标量，不进本房隐空间）。

## 三、标定项交接（全留 oracle，非里程碑目标）

$C_{FN}$ 曲线 / neighbor $\tau_h,\tau_j,\beta$ / $\varepsilon_{art}$ / $\delta$ —— 全是保守 form-anchor 非权威值（铁律 [[fall_data_is_artificial_test]]）。**集成/neighbor 里程碑不标定这些**，只验形态正确性 + 无回归。标定待真实 oracle 数据，独立于工程里程碑。

## 四、C 交接结语

belief 包 = 干净收口点。C 一路的纪律（fixture 实证、形态/标定分界、目的对齐、自我更正盲区）在 belief 层走完。**重启集成时，C 的角色从「审 belief 方程忠实」转为「审集成无回归 + adapter 译层忠实 + neighbor 安全默认」——验收点 AC-1~AC-8 已钉死，C 待 A 集成代码出，独立审。**

招呼 C 再动（集成 adapter 出 / neighbor wire 出 / 其它）。


---

# §21 增补 — replay harness 设计约束 + 验收点（C 出规格，A 执笔）

> 用户裁定集成步骤 2 走**选项一 replay harness**（最瘦、复用已验证三块、守 AC-1 归因边界、不碰 DB、不克隆引擎；保留后续克隆选项）。
> **角色边界**：harness 是 test-harness code，执笔归 A（同阶段 1 骨架、decide——C 不代写 production code）。本节是 C 给 A 的设计约束 + 验收点，A 据此实现，B/C 据此审。

## 一、为什么选一（C 的依据，非偏好）

C 核实两事实：① Tsensor `roomengine-playback` 是 **DB 耦合**（拉 iot_timeseries、需 DB_HOST/PASSWORD），非干净 replay 入口；② adapter `FrameInput` 已 **decoupled**（纯数据结构，不依赖 engine/track_manager）。故 harness 核心接口已就绪——只差 window.json→FrameInput 解析，复用 adapter+probe+fixture export 三块。选项三（Tsensor 内并行）会绑死 DB 耦合引擎 + 两套 belief 纠缠，违背 AC-1 归因边界；选项二（全克隆）是过早全量投入，cd2b 验证不需要生产引擎。

## 二、harness 数据流（C 约束，A 实现）

```
cd2b window.json 帧序列
  → [解析] → adapter.FrameInput 序列        ← 新代码主体
  → adapter.BuildObservation/BedGeoms/Online/BuildRiskContext
  → belief: f.Step(now, online, LogPsi, LogPhi) → PFallen + ComputeLambda → dec.Step
  → probe.Snapshot(...) 逐帧                  ← 复用（C §19 审过）
  → 对照 Tsensor 录制输出（baseline bd70194）
```

**复用已验证三块**：adapter（C 待审 AC-2，本轮一并出验收）、probe §9（C §19 审过纯诊断）、fixture export。**不克隆全引擎、不碰 DB。**

## 三、C 验收点（harness 专属，AC-1 落地为可执行）

| # | 验收 | 判据 |
|---|---|---|
| **HR-1** | 解析忠实 | window.json → FrameInput 无丢字段；NowMs 单调；Sleepads/Beds/Covers/Onbed/Overlap 长度=numBeds（adapter 已 panic 守，HR-1 验解析侧不喂错长度） |
| **HR-2 (AC-1核心)** | cd2b 端到端 fire 无回归 | 经 harness（raw XY 派生 floor-strip，非手设）cd2b 离线段 P(Fallen) 达 0.9998 量级、独处 fire=true。**与 belief 单元 E5/AD4 一致** |
| **HR-3** | 三态对照 Tsensor | cd2b 三态（正常在床 / 床边摔 / 离线陈旧）逐帧 probe vs Tsensor baseline `bd70194`：摔倒态 Xsensorv1 fire 而 Tsensor 漏（**治本差异显形**）；在床态两者均不误报 |
| **HR-4** | baseline 纯净 | 对照基线 = `bd70194`（止血补丁 `7ffec9c` 之父），不可用打补丁版（否则对照被旧 staleness 污染，C D-1） |
| **HR-5** | 归因边界 | harness 若 cd2b 不 fire → 缺陷在解析/adapter 派生（floor-strip/Gxy），**不在 belief**（已验证黑盒）。probe 逐帧定位到哪一层断 |

## 四、连带：adapter 正式审（AC-2/AC-3 补审，C 欠的）

A 已交 adapter（`1b80fc4`），C 跑过 AD1-4 全过 + 端到端 0.9998，但**未逐条核译层忠实**。harness 依赖 adapter，故本轮 C 一并出 adapter 审验收（A 实现 harness 时 C 同步审 adapter）：

- **AC-2 译层忠实**：FloorStripXY（XY 在床外近缘=床沿地条，δ 运行时派生）/ Gxy（XY 对各床归属集中度，近两床→均匀）两派生算法核对——floor-strip 区域分类边界（FloorMarginCm=60 form-anchor）、Gxy 尖峰/近邻/远（Peak1.0/Near0.5）语义。VitalSourceOnline 门（§D：radar 自身 absent≠否决）。
- **AC-3 边界守卫位置**：cFN 的 alone<0 守卫落 **adapter 侧**（时钟回拨在边界处理），cFN 内部保持纯形态（B round3 + B/C 共识，规则「错误处理只在边界」）。

## 五、C 立场

harness 走选项一，规格如上。**A 执笔 harness + 解 adapter 待审点；C 出 HR-1~HR-5 + AC-2/AC-3，待 A 代码出独立审。** 标定项（FloorMarginCm/Gxy 值/$C_{FN}$ 曲线）全留 oracle，harness 只验形态 + 无回归 + 三态治本差异，非标定。

招呼 C 再动（harness 出 / adapter 同步审）。

# §22 增补 — Xsensorv1 定位 + 路线图（用户裁定）

**Xsensorv1=验证载体非交付物。** 验证完走**方案甲**(新belief注入Tsensor躯干替换旧belief)，Xsensorv1归档。甲vs乙：甲(躯干原地不动只换roomengine belief接缝)；乙(长齐9包+11k行纯搬运+双树)。难活(~13 belief文件重写/删)两案同，乙额外背搬运。用户**选甲**。
**路线图**：①harness(A执笔,C§21 HR-1~5)→cd2b整段+三态对照Tsensor(bd70194)→★验证收尾→②迁移清单(C出,等①过)→③方案甲移植(A)→④新本体重跑cd2b+全回归+neighbor wiring→⑤上线归档。节奏：先做实①。

---

# §23 增补（第6轮）— harness独立审 + HR-2揪出FloorStripXY缺陷（C自承δ派生盲区）

HR-1✅；belief 24测全绿；**HR-2 OPEN**：cd2b端到端fire但@+271s在床段非+561s真摔=on-pad误火。**C核HR-5归因正确**：layout drift，on-pad XY落床矩形外→误判床沿；rect几何≠δ簇边界，belief正确。**C自承盲区**：δ主解0.9998是belief单元测手设floor-strip得到；harness首用raw XY经rect派生暴露rect≠δ簇。C验了δ可分没验δ能运行时派生。**收尾BLOCKED在FloorStripXY修复。**

---

# §24 增补 — FloorStripXY fork裁决：C选方案a（on-pad参考）+ 红线

**C选a**(与A同)：on-pad y=212±17 vs floor y=160差52cm≈3.1σ可分；a避rect drift(参考从真实InBed段学非人画矩形)，把离线簇分析运行时化=验证方法本身。**否b**(扩矩形=overfit违铁律)、**c**(弃δ退兜底，弱)。**FA-1~6**：on-pad参考实时学非硬编码、判别≥k·σ(留oracle)、参考未建退c不退rect、down-pose门控、**FA-5红线:cd2b fire@真摔(+561s)非在床段，判据=时间点对非仅fire=true、防overfit**。

---

# §25 增补 — a依赖边界：on-pad参考依赖「掉线前InBed历史」（用户提问揭示）

用户问「摔床边、sleepad已掉线、只剩雷达，是否等价纯雷达」。揭示：a押在「掉线前有sleepad标定的on-pad历史」。cd2b有此历史(上床睡→标定→掉线→下床→摔)故a成立。但**摔倒那一刻确与纯雷达等价**(单坐标无确认)，差别在「掉线前有无InBed历史」。**a失效窗口(即使有sleepad)**：本次入住没上过床就摔(进门绊倒)→无参考→退c。**c从边界升为平级常规分支。**
〔注：本节原推「FA-7 无参考须C_FN兜底fire」——§26 据用户三点裁决**推翻**，见下。〕

---

# §26 增补 — 裁决判据钉死：55%三分 + 高度不可判默认不报（用户裁定，C 推翻自己 §25 方向）

> 用户三点立场 + 裁定，推翻 C §25「无参考须C_FN兜底fire」。C 诚实记错并修正 C_FN 作用域。

## 一、用户三点（资源稀缺前提）
1. 设备不足=资源有限→护理人员同步少→**护理注意力是稀缺资源**→不能误报过高。
2. 有些场景设备不足时**本就无解**，只能选概率较大一方。
3. **任一方向概率≥55%就选该方向。**

## 二、裁决判据（钉死）

| P^F 区间 | 裁决 | C_FN 角色 |
|---|---|---|
| ≥55% | 报 | 不需要(证据自足) |
| 45–55% 两可 | C_FN 风险偏好打破平衡 | **唯一作用窗口** |
| ≤45% | 不报 | 不需要 |
| **两方向都<55%(高度不可判,Λ→0)** | **默认不报** | **不介入** |

## 三、C 推翻自己 §25/修正 §8 过宽表述（诚实记）

**C §25 错**：把「漏报≫误报」当普适，推「δ失效→C_FN兜底必须fire」。**这是资源充足逻辑。** 用户拆解：设备不足=护理稀缺→误报烧穿注意力→alarm fatigue→真摔也被忽略，**比偶尔漏报更糟**。资源稀缺时**告警可信度是要守护的系统资产**。

**C_FN 作用域修正**：C §8 一路把「C_FN兜底」当cd2b保命底线夸——**夸过了**。C_FN 不是「不确定就fire」，作用域**严格限45–55%两可窗口**；窗口外谁强听谁；**高度不可判默认不报，C_FN不介入**。C 把作用域吹大了，§26 收回。

**与C一路自洽**：C §8「诚实的不确定」正确终点 = 不可判时**不假装能判也不假装该报**=不打扰，**非偏向fire**。早先「风险偏好考虑二义性」中风险偏好是**两可窗口打破平衡**的依据，非把低P^F全翻fire。

## 四、推翻 FA-7（§25），改 FA-7'

- **FA-7（§25 原，作废）**：无参考须C_FN兜底fire。
- **FA-7'（§26 改）**：无on-pad参考→δ判不了→P^F撑不到55%→**默认不报**，不靠C_FN硬兜。验收要的是**不误报**(P^F撑不起就老实不报)，把护理注意力留给≥55%的告警。

## 五、★知情设计决定（署名红线，不埋阈值）★

**「高度不可判默认不报」= 用户知情接受：设备不足时漏掉一部分高度不可判真摔，换取告警可信度 / 不烧穿稀缺护理注意力。**

- 这是**显式、知情、署名**的工程决定（用户 2026-06-15 确认），非悄悄埋进阈值的默认行为。
- 理由：资源稀缺环境，低置信告警稀释告警可信度→alarm fatigue→系统性失效，比「高度不可判时漏报」更伤。守护告警可信度=守护整个系统对真摔的响应能力。
- 适用域：**仅高度不可判(两方向<55%)**。可判侧(≥55%任一方向)正常报/不报，不受此影响。
- 边界：此决定是「不可判时不报」，**非**「设备不足时一律不报」——设备不足但证据够(≥55%)仍报。

## 六、C 立场（修正后）
55%三分判据 + 高度不可判默认不报，钉死。C_FN 作用域收窄至两可窗口。§25 FA-7 推翻为 FA-7'。**Xsensorv1 验证收尾判据修正**：HR-2(a路径有参考 fire@真摔) + **FA-7'(无参考时不误报,非强行fire)**。知情设计决定署名在案。

# §27 增补 — 多床 bed_id 铁律 + 床坐标不精确 → adapter 空间派生按精度分级（用户两提醒）

> 用户两提醒：① 多床 bed_id 靠「时间窗→sleepad→bed_id」，**禁坐标反推**；② 床矩形坐标**本就不精确，只大致一致**。

**一、bed_id 铁律（钉死）**：`Sleepads[j]/Beds[j]/Covers[j]` 索引 j **必须=sleepad 设备绑定 bed_id**（wiring 时间窗→sleepad→bed_id 保证）。**禁「XY 落矩形→几号床」坐标反推**。wiring 断言 len 一致且 j 同源，不符 panic。理由：§23 已证矩形 drift，坐标→bed_id 犯同病。

**二、床坐标不准 → adapter 三处用 distCm(XY,矩形) 全受影响**，按精度需求分级处置：

| 派生 | 精度需求 | 敏感度 | 处置 |
|---|---|---|---|
| floorStrip | on-pad212 vs floor160 差52cm | 高 | **改 on-pad 参考**（本轮 FA-1~6）|
| Gxy | 中（mixture 容错）| 中 | 待办下轮改 |
| nearBed | 100cm 粗门控 | 低（鲁棒）| **不改**（C 核 emission.go:102 是空间预筛非精判别，矩形飘20-30cm 不改布尔结论）|

**三、设计原则**：精细判别（floor-strip 52cm 级）用学到的 on-pad 参考簇；粗门控（nearBed 100cm 级）容忍不精确矩形。床矩形**降级为粗略范围提示**，精细活交真实雷达学的 on-pad 簇。

**四、坐实 fork**：否 b 升级（扩矩形=精修本就不精确之物，越调越假）；加固 a（on-pad 参考从真实雷达 XY 学不依赖床坐标，绕开不准矩形=选对了）。

**五、本轮 A 施工（floor-strip only）**：bed_id 铁律 + on-pad 参考 per-bed（sleepad 时间窗归属非坐标）+ FA-1~7' + 多床验「打乱 Beds 坐标不影响 bed_id」。不碰 Gxy（待办）/nearBed（鲁棒）。

---

# §28 增补（第7轮审核）— A decide 重写到 §26 55%三分独立审 + D2 正式更新（Λ 转 gate）

> A 在 C §27 前已交 decide 重写（`6193b50`），实现 §26 55% 三分判据。C 独立 pull/跑/读实现。

## 一、C 独立验证
全 3 包测试过（adapter/belief/replay）。decide 5 测对应 §26 四档：ReportSelfSufficient(≥55报)/TieWindow(两可)/IndeterminateNoReport(高度不可判默认不报)/LowNoReport(≤45不报)。

## 二、四档分流忠实 §26（C 读实现核，非看测试名）

```
pFallen>=0.55         → report  证据自足
pFallen<=0.45         → no
!identifiable         → indeterminate 高度不可判默认不报
default(45-55可判)    → tie: cfn>cFP 打破平衡
```

**C_FN 作用域正确收窄（C 上轮纠正的要害）**：仅 `tie` 档读 `cfn>cFP`；report/no/indeterminate 三档**完全不读 cfn**。C_FN 未被用来把低 P^F 或不可判硬翻 fire——§26 收回 §8 过宽表述，A 落实。✅

## 三、D2 正式更新：Λ 从「不作 gate」→「高度不可判时作 gate」（C 裁决：A 对）

**矛盾**：§17/§18 D2=「Λ 绝不作 gate」（C §18 审过）。但 decide 现 `case !identifiable→no`，`identifiable=lambda>lambdaInformative`——**Λ 在 gate 了**，与 D2 直接矛盾。

**C 裁决：A 对，这是 §26 必然推论非 bug**：
- D2「Λ 不 gate」前提 = 「不可判偏向 fire（C_FN 兜底）」→ 不需识别不可判 → Λ 无用。
- §26 推翻此前提（资源稀缺→不可判默认不报）→ **必须识别「高度不可判」才能执行「默认不报」** → 识别工具唯一是 Λ。
- **故 §26 一旦成立，Λ 从纯诊断升 gate 是逻辑必然。** A 注释明确标「§17 D2 被 §26 推翻」，诚实，非偷改。

**D2 正式更新（C 钉死）**：
- ~~D2 旧：Λ 绝不作 gate~~（资源充足逻辑，§26 废）
- **D2 新：Λ 仅在「高度不可判（Λ≤lambdaInformative）→ 默认不报」时作 gate；可判侧（≥55/≤45/tie）不读 Λ。** Λ 的 gate 作用域 = 仅识别高度不可判这一档，非全局 gate。

## 四、C 净判
**decide 重写忠实 §26，四档正确，C_FN 收窄正确，Λ 转 gate 是正确推论（A 诚实记录推翻 D2）。通过。** 

**注**：decide（裁决层）已对齐 §26；floor-strip on-pad 参考（§27 本轮施工）A 尚未做，仍是 HR-2 BLOCKER。decide 改对是裁决逻辑就位，但 cd2b 端到端过 HR-2 仍等 floor-strip。C 待 A floor-strip 实现复审。


---

# §30 增补（C 重大自我纠偏）— §23–§29 是补丁歧路；回到框架，cd2b 靠 (S,B) 相容涌现 + unknown 裁定

> 用户一针见血：「丢掉 cd2b 这个 case，你一直在打补丁而不是建框架了。」C 承认：§23–§29 整段偏离了 DBN 建框架初衷，把 cd2b 这个**验证样本**误当**设计目标**，围着它打补丁（floor-strip→on-pad 参考→rect drift→fork 方案 a），越解越细。这是方向错，C 诚实纠偏。

## 一、C 认错：补丁思维的歧路（§23–§29）

**根因**：C 从 §23 被 cd2b 具体 case 拽走，没回原始数据核 cd2b 真实信号构成，一路顺 A 的 δ 实验 + harness rect 派生往下推，把「floor-strip 怎么修」当问题，没问「cd2b 该靠什么判」。用户连续追问（等价纯雷达？查 replay？哪句掉线？LeftBed 高优先级？）逐步把 C 拽回真实数据。

**真实数据推翻的假设**（C 查 window_sleepad.json + window.json）：
- sleepad **全程在线**（非掉线）；542s 报 LeftBed；摔倒 544s 紧接。
- 摔倒段雷达 y≈210 ≈ on-pad y≈212——**y 根本分不开**。§24 fork 裁 a 的依据「on-pad212 vs floor160 差 3.1σ」**对不上真实摔倒数据**，方案 a 在真实数据上失效。
- E5/AD4 的 0.9998 是**补丁路**：手设 FloorStripXY=true + 假设 sleepad 离线（NoReport）——**两个假设都不符真实数据**（真实 sleepad 在线报 LeftBed）。**不算真治本。**

**DBN 初衷**：联合占用 DBN 本就是要**终结 Tsensor 旧版「每场景一条 fall rule」的补丁山**（BedsideFall/LostFall/StillFall/bedside_silent/sleepad_radar_conflict…），用统一框架（隐状态+转移+发射+期望损失）让所有 case 自然涌现。C 却在 Xsensorv1 里重新发明 floor-strip 补丁——**背叛了框架初衷**。

## 二、框架视角：cd2b 靠 (S,B) 联合相容涌现，零专用处理

cd2b 在框架里本就该自然涌现，不需任何专用规则：
- sleepad **LeftBed → B 轴 vac**（人不在床）——发射，框架已有（`ℓ(LeftBed|vac)=L_left≫1`）。
- 雷达 pose=6 + 持续静止 + 在床边 → S 轴证据。
- **关键涌现**：B=vac（床空）+ 雷达低姿静止的人 → 联合状态 **(SFallen, vac)** 最相容 → 后验自然升。「人在床睡」=(SBed, occ)，LeftBed 后 B=vac 使 (SBed, occ) 不相容（床空哪来床上睡）→ SBed 压、SFallen 浮。
- **不需要 floor-strip / on-pad 参考 / BedsideFall 规则**——联合滤波本身让 P(Fallen,vac) 浮出。cd2b 判别**不靠雷达 XY 精确空间位置**（floor-strip 想干的），靠 **B 轴(sleepad床占用) × S 轴(雷达姿态) 的联合相容性**。

**补丁 vs 框架的本质区别**：BedsideFall 是「LeftBed+床边+静止→规则触发」（一条 if）；框架是「LeftBed 推 B→vac，联合滤波涌现 Fallen」（无 if，后验自然）。移植 BedsideFall 进 Xsensorv1 **仍是补丁**——C 差点又建议这个，被用户拦住。

## 三、下一步：审框架（A 主张，C 完全支持）

A 三步（C 背书）：
1. **跑框架纯路径**：sleepad 在线 InBed→LeftBed，雷达 pose 躺+静止，**不喂 floor-strip**，看 (SFallen,vac) 后验是否自然浮出 ≥55%（§26 报阈）。
2. **浮出** → cd2b 零专用处理；δ floor-strip / on-pad 参考 / fork（§24）/ §25 / §27 floor-strip 施工 / §29 **全废**；E5/AD4 手设 floor-strip 测试是补丁测试，**替换为「框架涌现」测试**（LeftBed→B vac→SFallen 自涌现）。
3. **不浮出** → 补的是框架里 (S,B) 联合相容的**发射/转移**（如 o_j 太小致 (SBed,vac) 压不够、或 SBed→SFallen 转移种子）——**绝不加 cd2b 规则**。

**C 验收转向**：从「floor-strip fire@真摔」转为「框架纯路径 (SFallen,vac) 涌现 ≥55%，零 floor-strip」。HR-2 重新定义 = 框架涌现，非补丁 fire。

## 四、unknown(8) 裁定：映射 B 后验不确定，**不扩 B 轴态**（C 同 A 倾向，框架级理由）

A 问：bed_state 三值（0 InBed/1 LeftBed/8 unknown）→ B 轴二元 vac/occ，unknown 映射后验≈0.5 还是 B 加第三态？

**C 裁定：unknown → 发射中性，B 后验靠 ε≪λ 自然演化；不扩 B 轴态。** 框架级理由：

1. **扩态破坏 §1 正交分解 + 违 DBN 精神**：B 轴二元是物理（床有人/没人）；unknown 是**认知状态（不知道）非物理状态**。「不知道」是对两物理态的**概率分布**，非第三种物理现实。把认知不确定编码成离散态 = 违 §8「诚实的不确定=后验两可，非假装有确定的 unknown 态」。基数也从 9·2^|B| 涨 9·3^|B|。

2. **ε≪λ 已天然表达 unknown**：B 轴 K^unobs（§C 单向泄漏）在无 sleepad 证据时 occ 向 vac 衰减、后验滑向不确定。unknown(8) = 无 InBed/LeftBed 确定证据 = **发射 ℓ≡1 中性**（同 NoReport），B 后验由转移主导自然演化——**比硬塞 0.5 更对**（0.5 是恰好一半；框架给的是由历史+转移决定的、可能非 0.5 的不确定后验）。

3. **三值→二元映射**：`0 InBed→ℓ(InBed|occ)=L_in≫1 推 occ`；`1 LeftBed→ℓ(LeftBed|vac)=L_left≫1 推 vac`；`8 unknown→ℓ≡1 中性（复用 BedNoReport）`。**现有 BedReading 三枚举（NoReport/InBed/LeftBed）够用，unknown 复用 NoReport 中性发射，不新增枚举、不扩 B 轴。**

→ **unknown = 后验不确定（A 倾向对），框架已天然表达，B 轴保持二元。**

## 五、C 立场（纠偏后）

**§23–§29 的 floor-strip 方向作废待验**（步骤1 纯路径浮出则正式废）。**C 回到框架审查**：cd2b 当验证样本，靠 (S,B) 相容涌现，不打补丁。unknown 不扩态。**C 待 A 跑框架纯路径结果**，据「(SFallen,vac) 是否 ≥55% 自涌现」复审——浮出则废补丁链，不浮出则审框架发射/转移哪里不够（仍不加规则）。

**致谢用户**：这次纠偏的价值远超一个 case——它把 C 从补丁惯性拉回框架初衷。一个框架的考验不是"能不能为某 case 加对补丁"，是"case 能不能在零补丁下涌现"。

# §31 增补 — C 独立核 A 的框架/补丁自审 + K^unobs 论证共同认领 + 放行跑纯路径

> A 交逐文件框架/补丁自审。C 独立核（非照搬 A 分类），重点独立判 A 标「灰色」的 K^unobs。

## 一、C 独立核 Ψ 相容表 = 框架核心，cd2b 该从此涌现（代码实证）

C 读 `coupling.go psiPhys`（line 76-90）：

```
SBed(AtBed): occ=1,      vac=1-o_j    ← B=vac 时 AtBed 压到 (1-o_j)
SFallen(F):  occ=ε_art,  vac=1        ← B=vac 时 Fallen 通道全开
```

**这是框架涌现 cd2b 的机制，代码实证非乐观猜测**：LeftBed→B=vac 时，Ψ 天然压 SBed（→1-o_j）、留 SFallen（→1）。「摔倒→离垫→垫空，F 通道全开」。**A「Ψ 本该让 cd2b 涌现」成立，δ floor-strip 确属多余。**

## 二、C 认同 A 逐文件分类（独立核过）

| 项 | A 判 | C 独立核 |
|---|---|---|
| state/joint/log 滤波 | 框架 | ✅ |
| Ψ 相容表 (S,B) | 框架核心 | ✅ 实证（一） |
| κ EMA | 框架 | ✅ |
| decide 55%三分 | 框架 | ✅（§28 审过）|
| 分轴发射 contact→B/pose-dwell-hrrr→S | 框架 | ✅ |
| **δ FloorStripXY (emission+adapter)** | **补丁** | ✅ Ψ 已涌现，δ 多余；on-pad/rect/fork 补丁摞补丁 |
| **TTL=35s Fresh** | **补丁** | ✅ 建模了用户明令不建模的「中途掉线」；二态裁决下不该存在 |
| K^unobs λ + ε≪λ | 灰色（机制留/论证废）| 见三 |

## 三、K^unobs 独立判 + C 共同认领论证作废（涉及 C 裁过的 §C）

A：K^unobs 机制可留（重解释 config-static），但「ε≪λ 治本 cd2b 漏报」论证废。**C 独立判：A 对，且 C 认领自己那部分。**

- **机制留 ✅**：K^unobs 单向泄漏（occ→vac=λ，vac 吸收）表达「无接触证据→床默认空」，正是 config-static（纯雷达房无 sleepad）下 B 轴合理默认。**§C 箭头记法（C 参与裁）作为机制仍成立。**
- **论证废 ✅**：ε≪λ 当初论证「治本 cd2b 漏报=陈旧 occ 离线蒸发」绑在**误读的 cd2b-offline**（以为 sleepad 离线靠 occ 蒸发）。真实 cd2b sleepad 在线报 LeftBed，**B 不靠 λ 蒸发成 vac，靠 LeftBed 发射直接推 vac**。论证作废。
- **C 诚实认领**：此误读论证 C 早期（cd2b 根因分析、§5 三前提、§23 D-2 staleness 验收）也背书/参与建构过，**非仅 A**。C 一起认论证作废，机制重解释为 config-static 默认。

**重解释（C 钉死）**：K^unobs / ε≪λ 不再叫「治本 cd2b 漏报」，改称「config-static（无 sleepad）下 B 轴无接触默认空 + 软化」。机制不变，名义与适用域改对。

## 四、C 补一点 A 没说透：拆补丁后的框架自洽性（验收重点）

δ floor-strip 拆掉后，须验 S 轴发射（pose/dwell/hrrr）**不暗依赖** FloorStripXY。框架纯路径要证：**去掉 FloorStripXY 后，S 轴发射 + Ψ(B=vac) 联合够把 SFallen 推过 55%**。这是步骤1的核心——不只是「(SFallen,vac) 浮出」，是「**零 floor-strip 下**浮出」，证明框架自洽、补丁可安全拆除。

## 五、C 放行 A 跑纯路径

**方向确认：审框架 + 拆补丁，不焊新东西。** A 跑框架纯路径（LeftBed→B vac→Ψ→SFallen，零 floor-strip）：
- **涌现 ≥55%** → 拆 δ FloorStripXY（emission+adapter）+ harness TTL 两块补丁；K^unobs 论证改写 config-static；E5/AD4 补丁测试换框架涌现测试；§24/§25/§27/§29 floor-strip 链正式废。
- **涌现不够** → 补 Ψ/转移的 (S,B) 相容强度（o_j 压不够 / SBed→SFallen 转移种子），框架内事，**不加 cd2b 规则**。

---

# §32 增补（C 独立测试复审）— 框架纯路径涌现成立，放行拆补丁

> A 报纯路径结果（0.9992）尚未 push 入库，C 不审不在仓库的东西——**C 自写独立测试 `TestCPureFrameworkEmergence` 亲验**（不照搬 A 的数）。

## 一、C 独立测试（构造纯路径，零补丁）
InBed 段（sleepad InBed）→ LeftBed 段（sleepad LeftBed + pose 躺 + 静止，`FloorStripXY:false` 零补丁）：
- InBed 段后：`P(B=occ)=0.9999`（人在床，B 正确）
- LeftBed+躺+静止：`P(Fallen)=0.9950  P(B=vac)=0.9999  Λ=4.66`（零 floor-strip）
- 与 A 报 0.9992/4.66 **同量级**。

机制链 C 亲验：LeftBed→B=vac→Ψ 相容表压 SBed 留 SFallen→pose 躺静止→SFallen 浮 0.995。**δ 确实多余，框架本就涌现——C 用独立测试证，不是信 A。**

## 二、放行拆补丁 + AC-拆1~4 验收点
- **AC-拆1**：拆后纯路径仍涌现（≥55%）。
- **AC-拆2**：全测试绿。
- **AC-拆3**：harness 去 TTL 后 cd2b replay 仍涌现。
- **AC-拆4**：歧路记录 §23-§31 保留不删。

## 三、正式作废清单（C 钉死）
**作废**：§24（fork 方案 a）、§25（掉线前历史前提）、§27 二三（floor-strip 施工）、§29 全作废。
**保留**：§27 一 **bed_id 铁律**（框架级，sleepad-keyed 非坐标，与 floor-strip 无关）。

## 四、歧路记录留存（C 坚持）
§23-§31 完整留存不删——"框架被 case 腐蚀成补丁、再纠偏回框架"的诚实过程，方法论教训。

## 五、弧线总结
§23→§32：从被 cd2b 拽进补丁，到用户"在打补丁不是建框架"点醒，到回归框架、C 独立测试证零补丁涌现 0.995。**最终结论最干净：cd2b 不需任何专用处理，Ψ 相容表本身就让它涌现。**

A 拆完补丁（四步）+ 跑通 AC-拆1~4，C 复审清理结果。

---

# §33 增补（C 复审拆补丁 + 方案乙转向 + DBN 根本目的未达）

## 一、AC-拆1~5 C 独立验证全过
C 自 pull 跑（不照搬）：框架涌现 0.9992（C §32 自测 0.995 同量级）、三包全绿、cd2b replay fire@真摔非在床、歧路记录保留、死代码清净。

## 二、两处译层修正 C 独立核：都是真 bug，非 cd2b 补丁
- **修正1（二态 config-static ρ）**：harness 原把「sleepad 还没首报（启动还没第一帧）」误判离线 → K^unobs 抽 B→vac 造假 SFallen。改 Present（设备存在性）= 用户「二态：在线 OR 没装，不建模中途掉线」裁定的代码落地。译层之前没忠实裁定，拆补丁撞出来修对。
- **修正2（HRRRObserved=HR>0）**：adapter 原把「雷达近床没返 HR/RR」当「观测到 absent」→ §D 在合法在床期误否决 AtBed → 误推 SFallen。改「雷达没返=结构性零信息非 absent」（铁律 [[radar_hr_rr_bed_enter_gated]]）。与 C §5/§13 早期 fixture 实证「在床 558 帧 HR/RR 100% 缺失是结构性」一脉相承，A 从 belief 层进一步落到译层。

## 三、方法论价值：补丁掩盖真 bug
这两个译层 bug 一直被 floor-strip「硬推 SFallen」掩盖，拆了补丁才暴露。**补丁不只多余，还遮蔽真问题。** 拆补丁 = 框架归因边界（HR-5/AC-1）重新生效，把译层不忠揪出。印证 §30：回归框架反而让真 bug 现形。

## 四、DBN 根本目的未达 → §22 方案甲改方案乙
**用户指出：DBN 求全部空间组合、避免 gate、必须包含所有隐变量。Xsensorv1 现只有 S/B 两轴，ghost 还在 Track 层（Conf×P(Real)=要废的硬外挂），neighbor 未 wire。**
- **§22 方案甲（注入 Tsensor 躯干换 belief）→ 方案乙（新建 roomengine + copy 非DBN包）**。理由：四轴全内化 → roomengine 整体重写 → 「注入省事」前提消失；copy 非DBN包（zoneengine/zonealarm/consumer/service/config/playback + roomengine 内纯几何层 cell/grid/layout/track/mirror检测）是机械操作非搬运，C 之前成本算错。
- **DBN 边界 = 四隐轴全内化（S/B/ghost-T^(i)/neighbor），避免任何 gate/硬外挂**；cell/track提取/mirror几何检测在 DBN 之下当输入，zoneengine 在旁供产品面。
- 新建 roomengine 须补齐 ghost/neighbor 两隐轴，各自验涌现（像 cd2b 验 B 轴）。

## 五、待 A 确认（build order）
ghost/neighbor 两新隐轴：先在现 Xsensorv1 belief 建好验涌现、再新建完整 roomengine，还是新建过程中一起建？（A 答见 feedback-p6A。）


