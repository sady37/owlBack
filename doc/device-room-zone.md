# device-room-zone.md — z 高度档（ObsZBand）设计决策

> 2026-06-15 记。来源：d5f7-0524 bathroom 真摔 case 排查中发现新 DBN 丢了 z，
> 架构师定 z 作 posture 正向证据接回。本文记 z 在 DBN 里的语义 + 校准 + 实现口径。

## 1. 问题：新 Xsensorv1 DBN 把 z 丢了

- **生产 wisefido-sensor**（`belief_adapter.go:124/172`）：`z → ObsZBand`（高度档 posture **正向证据**，非 fall）；
  注明 "fall 压制只走 realness/spatial/sleepad，不走 z；z↓ 连正向都不算"（kinematics Δz 已删 P2.1）。
- **新 DBN**（`adapter.BuildObservation` / `belief.Observation`）：**无 z 字段，ObsZBand 整个丢了**。
  占用模型重写时把姿态/z 层删了 → S 轴摔判定里 z 出现 0 次（grep 坐实）。

## 2. 实证：d5f7-0524 真摔（z 出卖 firmware 错 pose）

- case：bathroom 真摔，**firmware pose 误判成 sit(3)** 没报（漏报）；人躺地 ~21min 后起身 leftRoom。
- pose=sit(3) 段 z 分布（1247 帧）：

| z 桶 | 占比 | 真实姿态 |
|---|---|---|
| **0（贴地）** | **92.7%** | 摔/躺地板 |
| 21-40 | 1.3% | 马桶坐（标准款/过渡）|
| 41-60 | 5.1% | 马桶坐（comfort height）|
| >60 | 0.9% | 站姿 |

- 每分钟 z 均值全程 ~2-6cm（21min 无一分钟接近坐高）→ **全程躺地 = 真摔**，z 干净映射姿态。
- 新 DBN 丢了 ObsZBand → 看不到 z 差异 → 久坐马桶 / 真摔 都只剩 dwell → 都判 Fallen（久坐误火 + 真摔机理错）。

## 3. 校准：美国马桶座高

- 标准款：14-15″ ≈ **36-38cm**；Comfort/ADA（适老化常用，本系统老人监护场景）：17-19″ ≈ **43-48cm**。
- 坐马桶 → 下半身/座面 z ≈ **38-48cm**。摔（z≈0）与坐（z≈40-48）间有 **~20cm+ 清晰间隔**。

## 4. 设计决策（架构师定）：z 单向正向证据，绝不负向

- **z 只做正向证据，绝不做负向证据**（不否决、不压制任何态）。
- **z≥30 → 增加 "NO fall" 置信度**（有高度=直立/坐着=没摔），**时间积分**（持续 z≥30 → not-fall 置信越积越强；靠前向滤波逐帧累积天然实现）。
- **z<30（贴地）→ 中性**：**不是 fall 的证据**（z=0 不触发摔、不否决任何东西）。摔的检测仍走 dwell/pose/realness，z 不参与。

**z 的职责 = "排除摔"（高 z=直立），不是"检测摔"。** 修的是假阳性：
- 正常马桶久坐 z≈40（≥30）→ 累积 not-fall → **压住久静 dwell 想推的 Fallen → 不误火** ✓
- 真摔 z≈0（<30）→ z 中性不帮忙 → dwell 照常推 Fallen → 照报 ✓（d5f7 不受影响）

## 5. z 档阈

**采用生产原阈**（wisefido-sensor `calibration.go`：`zUprightCm=80 / zSitMinCm=30`，稳；我的校准阈 35-55/>60 作精调留 oracle，架构师定）：

| 档 | z (cm) | 语义 |
|---|---|---|
| ZNone | <30 | 假低/贴地：无信息中性（z=0 不否决，fall 走 dwell）|
| ZSit | 30-80 | 坐（美国马桶座 38-48 在内）→ 抬 Sit |
| ZStand | >80 | 直立活动 → 抬 OpenFloor |

> ⚠️ **档阈仅作"增加 belief 置信度"用 + 时间积分，不是单帧硬判**（架构师强调）。
> boost 因子 `lZ=8`（我）vs 生产 `lr=2.0`：我的模型 dwell 更强（dwellHi=3/dwellLo=0.5），需 lZ≳6 才抵消久坐 dwell；form-anchor，精标见 feedback-p6C。

## 6. 实现口径

- **adapter**：`z → Observation` 加 z 档信号（z 值或档）。
- **emission**：**z≥30 → not-fall 态（Sit/OpenFloor/直立）ℓ>1（正向抬）；Fallen 态 ℓ=1（不压，绝不 <1）**；z<30 → 全态 ℓ=1 中性。与 pose/dwell 并列融合、不加 gate。
- **时间积分**：前向滤波逐帧累积 z≥30 的 not-fall 似然 = 时间积分（持续直立越积越确定没摔）。
- **z=0 的 Fallen/AtBed 二义**：靠 B 轴分（B vac→Fallen / B occ→AtBed），同 PoseLying 二义处理（z 不分）。

## 7. 实现（2026-06-15 已接，架构师拍细分两档）

- **§7 决定**：细分两档，**用生产原阈**（z **30-80 → ZSit** 抬 Sit / **>80 → ZStand** 抬 OpenFloor；<30 → ZNone 中性）。
  我的校准阈 35-55/>60 作精调留 oracle（架构师定：生产原阈稳）。
- **实现**：`belief.ZBand` 枚举（observation.go）；`adapter.zBandOf(z)`（>60 站 / ≥30 坐 / 否则中性）；
  `emission.radarLogS` z-band 项：ZSit→`SSit×lZ`、ZStand→`SOpenFloor×lZ`，`lZ=8`（须 ≳ dwellHi/dwellLo
  抵消久坐 dwell，form-anchor 标定见 feedback-p6C）；**正向 only，Fallen 永不 ℓ<1**。
- **验证**（`belief/emission_test.go`）：
  - `TestEmissionZBand`：ZSit 抬 SSit log(lZ)、ZStand 抬 SOpenFloor、z 不压 Fallen（正向 only）✓
  - `TestZBandSuppressToiletSit`：久坐马桶(ZSit) **P(Fallen)=0.036** vs 真摔(ZNone) **P(Fallen)=0.997** ✓
    → **久坐误火 FP 治本，真摔不漏**；全 belief 测试套无回归。

## 8. 待办

- **cell learning 线**（C 说的"双线"另一条）：cell 用 z 学习 sit/toilet 区——本次只接 emission 线
  （直接 FP 修复）；cell-learning 线后置（慢学习、二级增强）。
- 端到端：d5f7 重放确认久坐段 P(Fallen) 实际被压（emission 单元已证，端到端待跑）。

---

## 9. realness 完整模型（正向证据 / FN-safe）— 2026-06-16 架构师定，Xsensorv1 待重设计

> d5f7/0616 排查发现 Xsensorv1 realness 用错模型（aScore 超速单 track + AND mirror），把静卧真人 B 误判 ghost
> （一次 42cm 体动 + 22ms 极小 dt 假超速 → aScore +3.63 → 永久 ghost）。架构师定下完整模型。应同步 DBN-Zone-Room §G。

### 9.0 总则：信息不完整 → 只能用正向证据 → 默认 real
雷达信息天然不完整（截断重放无 enter 事件、enter 区画不准、盲区）。**很多时候只能用正向证据**（证明 real），
**默认 real，只在正向证据指向 ghost 才判 ghost**。这是 FN-safe 的接受成本：宁可放过疑似 ghost，绝不压真人的摔。

### 9.1 身份 + A/B 时序（全程 logic_ID，track==2 gate）
- **身份用 logic_ID**（最小作功距离关联，[[track_identity_logicid_ghost_track2_scope]]）**不用 track_id**——
  firmware track_id swap 会造**假出生**（真人被换号看着像凭空新生）；logic_ID 能关联到既有身份 = 延续 = real。
- **A = 先到 logic_ID = real 锚（无条件）**：不要求 enter 区出生（enter 区常画不准，进门无 InRoom event 也正常），
  截断重放首 track（cd2b 一开始就在床上、无 InRoom）照样 real。**B = 后到 logic_ID = ghost 候选**。
- **ghost 仅 track==2**（1=永发，3+=不处理）。反射必成对（反射源是真人）→ 孤 track 永远 real（接受成本）。

### 9.2 墙外 ghost = 易、绝对（即时几何，非积分）
- track 超 **radar border** → firmware/radar **直接滤掉**，根本到不了我们。
- 到得了的墙外（border 内、wall 外）→ `reflectsAcrossWall` 几何**绝对**判 ghost（连线穿 wall、最近交点 ≥30cm）。
- **即时判定，无需积分**（我之前 §9 写的"墙外稳定 3-5min 负向积分"作废）。

### 9.3 墙内 ghost = 正向证据级联（难；按序 D → sync → 5min）
墙内镜像可能落屋里（几何抓不到），按**证据级联**，任一级判定即止：
1. **出生地距入口距离 D**：评分 **D/120**。**D<120cm（近门）→ real**（正向：从门进来的）。enter 区不准不要紧，
   用**距离软评分**不用硬门；老人走得慢、出生检测就在门边 D 小，120 足够罩住慢速真人入场。
2. **D≥120（远离门、可疑）→ A/B 镜像同步移动**（轴对移/平移）：**出生后 3-5 秒**内即可判，**同步 → ghost**。
   健康人走得快首检在屋里漏过第 1 级，但他在**活动**（不与人同步）→ 此级判非 ghost，误判风险小。
3. **仍不定 → 5min 后仍活动 → 认 real**（正向积分兜底）。

### 9.4 单 tick 瞬态忽略
- 墙内 track 的**瞬间移动（jump/超速）= mirror 或帧间隔过近(小 dt)假超速造成，只在那一 tick 出现，下一 tick 即回正常**。
  实证（0616 B）：5 个超速帧每个下帧位移即回 9-52cm；+72s 的 1916cm/s 实为 42cm 体动 / 22ms 小 dt 假超速。
- **单 tick 瞬态绝不判 ghost**（同步移动须**稳定持续**才算，不是一帧）。

### 9.5 latch 铁律（cd2b 病的反面）
- 确认 **real → 永久锁 real，绝不回判 ghost**（已确认真人摔倒静止/消失 = 仍 real，blind 续存照报）。现 `movedFromBirth` 闩保留。
- 确认 **ghost → 锁 ghost**。

### 9.6 aScore 超速单 track = 病根，整条删
现 `aScore += ArtifactQuantum`（realness.go:56）：每 track 每帧 `speed=位移/dt`，超 `SpeedCeilCm=100` 的部分
×0.002 单调累进，`PReal=exp(-(mScore+aScore))`。**单 track 即判 ghost**。删它**三重正当**：
1. **单 track 判 ghost = 违 track==2 铁律**（§9.1）。
2. **单调不退** = 一次 dt 噪声（22ms 帧太近 → 42cm 体动算成 1916cm/s）永久污染 → 真人 B 被永久 ghost。
3. 它想抓的 **firmware 换号已由 logic_ID 接管**：真换号瞬移 >`AssocCm=250cm` 自动成新 logic_ID（新生非超速），
   <250cm 吸收为延续。aScore 抓的全是 dt 噪声假阳 → 删了零损失。
- 我之前的 aScore 诊断 + EMA 衰减是**错方向**（EMA 只给单调累积加衰减，没换成出生地+同步移动模型）。

### 9.7 code mapping（重设计）
| 现状 | 改 |
|---|---|
| `aScore += ArtifactQuantum` 超速（realness.go:56） | **整条删** |
| `reflectsAcrossWall`（census.go:205） | **保留**作 §9.2 墙外绝对路径 |
| `CoexistRho>0 && IsReflection` AND（realness.go:53） | 改 §9.3 级联：**D/120 主 → sync(3-5s) → 5min-active** |
| census 无 birth→entrance 距离 | **新增** D（需把 enter 区 polygon 喂进 census）|
| `ReflSettleMs=3000`(3s) | sync 窗 3-5s（≈对），新增 **5min-active → real** latch |
