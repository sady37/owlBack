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