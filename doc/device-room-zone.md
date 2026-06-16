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

## 9. realness 正确模型（出生地 + 三时标 + 单 tick 瞬态忽略）— 2026-06-16 架构师定，Xsensorv1 待重设计

> d5f7/0616 排查发现 Xsensorv1 realness 用错模型（aScore 超速单调 + mScore mirror），把静卧真人 B 误判 ghost
> （一次 42cm 体动 + 22ms 极小 dt 假超速 → aScore +3.63 → 永久 ghost）。架构师给出正确模型。
> 应同步 DBN-Zone-Room §G。

### 9.1 三时标 + 谁做
| 机制 | 时标 | 谁做 |
|---|---|---|
| metal 反射 filter（无 micro-moving） | 3~5 **秒** | **firmware 主动滤**（到 DBN 已无 metal，realness 不管）|
| 墙外稳定 ghost（出生地判定；雷达自身过滤不掉） | 3~5 **分钟** | **realness 负向积分** |
| 墙内 track（含墙外 30cm）持续活动 → real | >5min（5 或 8min）| **realness 正向积分** |

### 9.2 两个积分（realness 重设计核心）
- **墙外稳定 → ghost（负向积分）**：track 位置**稳定**在墙外（>30cm 余量外）按时间累积 → 判 ghost。
  出生地定；一形成雷达自身过滤不掉。**不是一次偶发跑墙外**，是稳定持续（3-5min 时标）。
- **墙内活动 → real（正向积分）**：track 在墙内（含墙外 30cm）**持续活动**累积 >5min → 确认 real。
  （B 在床睡眠体动属持续活动 → 5min 后 real，不被误判 ghost。）

### 9.3 单 tick 瞬态忽略（关键，区别于"稳定"）
- 墙内 track 的**瞬间移动（jump/超速）= mirror 或帧间隔过近(小 dt)假超速造成，只在那一 tick 出现，
  下一 tick 即回正常**。实证（0616 B）：5 个超速帧每个下帧位移即回 9-52cm；+72s 的 1916cm/s 实为
  42cm 体动 / 22ms 小 dt 假超速。
- **单 tick 瞬态绝不判 ghost**（ghost 须 §9.2 的「稳定」墙外）。现 aScore 单调累积一次跳变即永久 ghost = 病根，违此。

### 9.4 Xsensorv1 现状 gap（待重设计）
- 现 realness：`mScore`（mirror reflectsAcrossWall）+ `aScore`（超速单调累积）—— **都不是 §9.1-9.3 模型**。
- 重设计：**删 aScore 超速单调**；实现 ① 墙外位置**负向积分**（稳定墙外→ghost）② 墙内活动**正向积分**（含 30cm 余量，>5min→real）③ 单 tick 瞬态忽略。metal 不做（firmware）。
- 我之前的 aScore 超速诊断 + EMA 衰减是**错方向**（EMA 只是给单调累积加衰减，没换成正确的墙位置+活动模型）。
