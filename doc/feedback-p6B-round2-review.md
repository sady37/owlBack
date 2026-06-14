# 评审 B：第 2 轮审查 — 阶段 2 emission/coupling + neighbor 方程 + 🟡 修复

审查对象：commits `6018f11`–`c5abc2d`（9 文件，665+ 行）
对照基准：`doc/DBN-Zone-Room.md` §2–§5/§D/§E/§A + `doc/feedback-p6C-acceptance.md`
审查阶段：阶段 2（实现审查）— 对照设计文档方程，查代码忠实度

---

## 总判：通过。§5 Φ 分轴正确、§4 Ψ mixture 正确、§D HR/RR 门控正确、§A neighbor 方程完整。E5 cd2b 离线态实证兑现（P(F)=0.9998，不靠 sleepad/HRRR）。一项需要裁定。

---

## 一、测试实测

| 组 | 测试 | 结果 |
|---|---|---|
| T1–T5 | 阶段 1 骨架 | ✅ 全过（回归无退化）|
| E1–E5 | emission | ✅ 全过 |
| C1–C5 | coupling | ✅ 全过 |

**E5 cd2b 离线态**：sleepad 离线 + HR/RR absent 无独立源 + pose lying + floor-strip XY，90 帧后 P(Fallen)=0.9998 >> P(AtBed)=0.0000。δ 几何主解实证兑现。

**C4 mixture FN-safe**：双床 a≈0.5、一床陈旧 occ、人摔 → mixture Ψ(F)=0.55 vs product Ψ(F)=0.01（差 55×）。§E FN-safe 实证兑现。

---

## 二、方程→代码逐条对照

### §5 Φ 发射（emission.go）

设计：$$\Phi_t = \prod_j \ell_{s_j}(o^{s_j}|B^j)^{w_{s_j}} \cdot \prod_c \ell_c(o^c|S)^{w_c}$$

代码 `LogPhi`：
- `contactLogB(o)` → B 轴，`w=Onbed` ✅
- `radarLogS(o)` → S 轴，`w=covers`（max over beds）✅
- 两轴在 log 域分轴加（`v = radarS[s] + Σ_j contactB[j][bmask_j]`）✅

接触核：InBed → BOcc 抬 `w·log(L_in)`；LeftBed → BVac 抬 `w·log(L_left)`；NoReport → 0 ✅
pose 核：lying 二义（AtBed=F 同 boost `lPose`）✅
dwell：still≥τ → F/AtBed `dwellHi`，O/Sit `dwellLo`；不分 F/AtBed（§5 含义）✅

### §D HR/RR absent 门控（emission.go:108）

设计：absent 否决 AtBed 须 gate 在独立在线 vital 源下；radar 自身 absent=零信息

代码：`if o.HRRRPresent { ... } else if o.VitalSourceOnline { ... }`
- `VitalSourceOnline=false` → absent 分支跳过（零信息，不否决 AtBed）✅
- `VitalSourceOnline=true` → `av[SBed] = 1/L_hr`（否决 AtBed）✅
- E3 测试：无源→logPhi(SBed)=0；有源→logPhi(SBed)=log(1/L_hr) ✅

### δ 位置似然（emission.go:118）

设计：cd2b 主解，floor-strip XY → boost F、suppress AtBed

代码：`if o.FloorStripXY { ... SFallen: lDelta, SBed: 1/lDelta }` ✅
E4 测试：logPhi(F)>0, logPhi(AtBed)<0, F>AtBed ✅

### §4 Ψ 相容势（coupling.go）

设计：$$\Psi_t = \sum_j a_j \tilde\psi_j + a_\varnothing$$（§E mixture）

代码 `LogPsi`：
- `a_j = κ_j·g^xy_j / (Σ κ_j'·g^xy_j' + c_∅)` ✅
- `ψ̃_j = κ_j·ψ_phys + (1-κ_j)`（κ→0 退火中性）✅
- `ψ_phys(AtBed,occ)=1, ψ_phys(AtBed,vac)=1-o_j, ψ_phys(F,occ)=ε_art, ψ_phys(F,vac)=1` ✅
- mixture 加权和（非 product）✅

### §3 κ EMA（coupling.go:43）

设计：`κ ← (1-γ)κ + γ·1[match]`，互活门控，无 max

代码：
- `matched` 和 `live` 逐床传入 ✅
- 仅 `live[j]=true` 时更新 EMA ✅
- 无 max（可升可降）✅
- C1 测试：持续失配 20 步 κ→0 ✅

### §2 MM 原语（mm.go）

`BedGeom{Covers, Onbed, Overlap}` — 三标量，由 adapter 经 cell 枚举填 ✅

---

## 三、🟡 B/C 共识项修复确认

| 项 | 状态 |
|---|---|
| bedOnline 长度契约 | ✅ `len(online) != nb → panic`（filter.go:64-66）|
| buildLogTBCol 提循环外 | ✅ `buildLogTBCol` 在 Predict 入口预存 bmaskN×bmaskN 表（filter.go:68）|
| 两套 belief 包 | ✅ 已删（上轮修）|

---

## 四、需要裁定的一项

### emission.go `geom0Covers()` 取 max 而非 per-state 加权

`geom0Covers()` 对各床 covers 取 max，所有 S 态用同一权重。多床房下，若床 A covers=0.3、床 B covers=1.0，雷达轴对所有人态（含 AtBed）都用 w=1.0 全权——包括人实际在床 A 时。此时雷达对床 A 的 coverage 仅 0.3，pose/dwell/HRRR 观测应弱化到 0.3 权重，而非满权。

当前逻辑等价于「只要有一张床被雷达全覆盖，雷达轴就全权」。这是多床房的**雷达 coverage 高估**——被全覆盖的床 B「借」了 coverage 给被弱覆盖的床 A。风险：床 A 的 dwell/pose 弱信号被满权放大，可能推高床 A 的 F/AtBed 信念。

**但这不是错误——是设计选择。** 设计文档 §5 的 `w_pose = covers(r,·)` 的点参数本身就未指定（C2 悬空项），A 取 max 是一种可行解释。替代方案（per-state 加权）需要将 `covers` 按 S 态分解（AtBed 用对应床的 covers，F 用 max covers 等），这会打破 Φ 的 S/B 分轴清洁性。

**建议**：保留当前 max 实现，在 DBN-Zone-Room.md 的 C2 悬空项处标注 A 的选择（w_pose = max_j covers(r,j)）及理由（保持 Φ 分轴清洁）。若后续多床 case 出现 coverage 高估导致 FP，再引入 per-state covers。

---

## 五、放行判据

- E1–E5 全过 ✅
- C1–C5 全过 ✅
- T1–T5 回归全过 ✅
- 🟡 共识修复确认 ✅
- covers max 裁定：保留，标注设计选择 ✅

**B 净判：阶段 2 emission/coupling 方程实现忠实，cd2b 离线态 E5 实证兑现，放行。neighbor 方程看 §A。**
