# AI 跌倒检测与房间认知（Room Engine 设计整合）

## 0. 背景与问题

当前 radar 的跌倒检测存在两个核心痛点：

1. **大量 ghost track 造成假警报**——雷达的 track 会凭空出现、分裂、跳变，导致跌倒误报。
2. **真实 fall 会被漏报**——人跌倒后贴地，雷达可能直接丢失目标，看起来像"track 消失"，不被报成 Fall。

单纯靠雷达原始 pose 判定（pose=5 Fall）既多虚报也有漏报。需要一个**房间级空间认知模块**，把每个房间当作一张自学习的"底片"，结合 radar 原始信号、业务事件（床/门）、家属反馈，给出更准确的异常判定。

模块代号：**Room Engine**，位于 `owlBack/wisefido-ai/internal/roomengine/`。

---

## 1. 核心设计观点

贯穿整个设计的 10 条"硬"观点，后续所有模块都服从：

1. **反演式设计**——Engine 不信雷达单帧；用房间长期记忆反推真相。Radar 说什么是 evidence，不是 verdict。

2. **物理必然性**——一切观测必须有物理原因：出生必从入口、消失必在边缘、运动必须平滑、Pose 必须与运动学一致、Z 必在信号可达域。没原因 = 假。

3. **连续性是统一打分尺度**——Kalman 残差同时捕获空间跳跃、速度不合理、ghost 漂移。是 track 层判 Real/Ghost 的唯一量化量。

4. **Cell 是工程容器（不是拟人化的"房间意识"）**——10cm×10cm 格子存在的两个目的：
   - **统一坐标**：把雷达本地 (h,v,z) / 画布 (x,y) / Layout 矩形等异构几何落到同一套 (col, row) 索引
   - **降低精度**：10cm 量化把浮点几何问题变成整数索引查询
   - 不拟人、不存概率、不做软边界。

5. **Cell 字段布尔/计数，不存概率**——单 cell 简单；模糊/渐变由"多 track × 时间累积"涌现。每次观测都是一次独立样本，长期分布自然表达空间的真实形状。

6. **没有点云**——雷达每帧给一个 track 一个 (x,y,z)。每帧单点 mark 单 cell。"人体占多个 cell"这个事实靠**时间+多样本**自然覆盖，不做人工扩散。

7. **每帧双维度喂**——每个轨迹点同时喂两条流：
   - **即时流**（track.History 近 5 秒）→ 运动学判断
   - **历史流**（cell 累积字段）→ 区域属性学习
   - 历史流**不过滤**（不看 Verdict），按本帧质量 `q` 分桶到 Real/Ghost。

8. **时间双分层**——即时流答"现在在干什么"（5 秒窗口），历史流答"这个位置长期是什么"（24h/7d 半衰期）；两者独立演化，在 engine 层交叉使用。

9. **人类稀疏先验 + AI 稠密学习**——人标 Enter/Bed/Toilet + 粗 Wall，高 Confidence 但非硬锁（床可搬）；Sofa/Walkway/Metal 等全由 AI 从观测反推。

10. **在线学习闭环**——3 组参数并行；Silent Fall 必须报警（拿家属反馈）；ground truth 从 alarm_events.resolution 回流；AI 必须超雷达 baseline 20% 才启用。

---

## 2. 系统数据流概览

```
Layout (DB)  ──┐
               ├→ RoomConfig → RoomGrid (静态 rasterize：InRoom/InFOV/EdgeDist/MaxZ/MinZ)
FOV 几何 ──────┘                       ↑
                                       │
iot:monitor:stream   (radar 每帧 track) │
                        │               │
                        ├── 即时流：push History → Kalman → 运动学判定
                        │                               ↓
                        │                      Verdict + Anomaly + alarm
                        │
                        └── 历史流：每帧一个 (x,y,z) 单点 mark 一个 cell
                              (不过滤，按本帧 q 分桶到 Real/Ghost 累积)
                                                        ↓
                                  cell 多维字段：Real/Ghost/ZVar/Dwell/Events/Pose...
                                                        ↓
                                  定时 UpdateBelief（每 5 min）+ 事件触发点更新
                                                        ↓
                                          Belief[3] 三组参数并行演化
                                                        ↑
alarm_events.resolution (扫表) ──→ AccuracyTracker → winner 自选
  家属反馈：真 fall / fake
```

---

## 3. 几何模型

### 3.1 房间范围 ≠ 雷达视场

两套独立几何，Cell 同时带两个布尔位：

| 字段 | 含义 |
|---|---|
| `InRoom` | 在物理墙多边形内 |
| `InFOV`  | 在雷达物理信号可达域内（由 hfov/vfov/afov/efov/installHeight 算出，非 Boundary）|

| InRoom | InFOV | 含义 | Engine 处理 |
|---|---|---|---|
| T | T | 真正的学习区 | 参与 UpdateBelief |
| T | F | 房间内的雷达盲区（柜子背面等）| 不学习，track 在此消失合法 |
| F | T | 雷达看到墙外（反射/门外）| Ghost 强证据来源 |
| F | F | 忽略 | 不分配 |

### 3.2 Layout 输入（人类画）

人类只标三类**矩形**（Enter / Bed / Toilet）+ 若干 **Wall 矩形**。其中：

- **Enter** 矩形：门/入口。算法用"点到矩形最近距离"算 `dEnter`（矩形内=0）。
- **Bed / Toilet** 矩形：床、马桶（合法静止区）。
- **Wall 矩形**：人类能画出的墙（偏粗），用来定义房间边界 + 减少 Open 区 cell 数量。
- **Sofa/Chair** 允许人类可选标注（尽力粗标），初始 Confidence 中等（见 §7）。

### 3.3 雷达视场（物理 FOV）

**Boundary 仅作 UI 参考**。真实信号可达域由物理参数算出：

```go
type RadarMount struct {
    Center       Point        // 画布坐标（Z = installHeight，cm）
    Rotation     int          // 画布旋转角（度，0-359）
    InstallModel InstallModel // Ceiling / Wall / Corn

    HFOV         int // 物理水平视场角（度）
    VFOV         int // 物理垂直视场角（度）
    AzimuthFOV   int // 安装后方位角有效视场（度）
    ElevationFOV int // 安装后俯仰角有效视场（度）

    Boundary     Boundary // 用户设置的 XY 有效边界（cm，UI 参考）
}
```

- **物理层**（硬件规格）：HFOV / VFOV
- **安装层**（姿态变换后有效视角）：AzimuthFOV / ElevationFOV
- **画布层**（layout 编辑）：Rotation
- **用户层**（UI 设置）：Boundary（仅参考，不作判定依据）

### 3.4 Cell 网格

- 网格边长：10 cm
- 房间最大：600×600 cm（MaxRoomWidth / MaxRoomHeight）
- Wall 矩形内部 cell 不参与学习/衰减（性能收益 ×N）

### 3.5 坐标系约定

**画布坐标系 = 房间坐标系**（前端 layout_config 所有对象共用）：

- 原点：画布顶部中心
- X 轴：左为负(-)，右为正(+)
- Y 轴：下为正(+)，上为负(-)
- 范围：x ∈ [-W/2, W/2]，y ∈ [0, H]
- 单位：cm

**雷达本地坐标系**（ceiling 模式）：

- 原点：雷达中心 (0,0)
- H 轴：左为正(+H = -X)
- V 轴：下为正(+V = +Y)

**坐标转换**：`owl-common/radarutils/coord.go` 的 `RadarToCanvas` / `CanvasToRadar`，与前端 `radarUtils.ts` 对齐。雷达帧的 position_x/y/z 都是雷达本地 cm，engine 必须自己做本地→画布转换。

**RoomGrid Origin 抽象**：`grid[0][0]` 左上角对应的画布坐标偏移（`OriginX, OriginY`）。默认 `(-RoomW/2, 0)`。对外 API 不暴露，调用方用画布坐标调 `CellAt(x,y)` 即可。

### 3.6 int 化约定与精度单位

**系统 10cm 粒度量化，小数无物理意义**——所有持久化字段、公共 API 签名统一用 `int`，四舍五入取整。

理由：
1. 雷达数据源头就是 cm 整数（decoder 里 dm×10→cm）
2. Cell 粗粒度 10cm，小数点后位数不产生可区分的判定差异
3. int 存储紧凑、序列化清晰、debug 可读
4. 整除自然形成"旧数据遗忘"效果（衰减到 0 即丢弃）

**精度单位表**：

| 字段族 | 单位 | 说明 |
|---|---|---|
| 坐标 / 距离 / 边界 | cm（int）| 画布/雷达本地/Boundary 全部 cm |
| 角度 | 度（int）| Rotation / HFOV / VFOV / AzimuthFOV / ElevationFOV 全用度 |
| 时间戳 | ms（int64）| `LastUpdateMs` 等 |
| 计时累计（Dwell / Still / PoseGroupSec） | 秒（int）| |
| 速度（FlowX/Y）| cm/s（int）| |
| 百分比（Confidence / Score / RiskScore）| 0-100（int）| |
| 次数累计（Real/GhostDecay / Events / PoseMismatch）| 次（int）| 带衰减，整除到 0 即遗忘 |
| Z 方差 ZVar | cm²（int）| 粗量化 |

**允许浮点的三个场景**：

1. **Kalman 内部矩阵运算** — 协方差 P / 增益 K / 残差逆 S⁻¹ 必须 float64，否则梯度失真。API 边界（`Update/Position/Velocity`）转 int。
2. **似然计算中间量** — UpdateBelief 的 `computeLikelihoods` 用 float 算比例（`clamp01`），结果 × 50 后 `math.Round` 到 int Confidence。
3. **衰减系数** — `math.Pow(0.5, dt/halfLife)` 内部 float，乘到 int 字段后 `math.Round` 取整（见 `scaleInt`）。

**ParamSet.Alpha/Beta** 保持 float64（运行时配置，非持久化）；**ParamSet.FlipTh** 是 Confidence 阈值，int。

---

## 4. Pose 编码与两阶段状态机

### 4.1 编码（对齐 `owl-common/observation/fields.go`）

```
0  Initialization    1  Walking           2  SuspectedFall (疑似跌倒)
3  Sitting (蹲坐)    4  Standing          5  Fall (确认跌倒)
6  Lying             7  SuspectedFloor (疑似坐地)
8  SittingOnGround (确认坐地)
9  BedSitUp (普通床上坐起)  10 SuspectedBedUp  11 ConfirmedBedUp
12 Running
```

注意：雷达的 `Sitting(3)` 是**蹲坐**，不是沙发坐姿。人坐沙发时雷达看到的是 `Sitting(3)` 和 `Standing(4)` 的反复跳变。

### 4.2 三对"疑似-确认-回撤"事件

| 事件对 | Suspected | Confirmed |
|---|---|---|
| Fall  | 2  | 5  |
| Floor | 7  | 8  |
| BedUp | 10 | 11 |

除了累计 Suspected/Confirmed，**"回撤"（Retract）是第三个关键信号**：

```
走完  : Walking → Suspected → Confirmed       → Confirmed++
回撤  : Walking → Suspected → Walking(未升级) → Retract++
```

**回撤的价值**：某位置雷达频繁疑似跌倒但都纠回 → 极可能是沙发（坐下 Z 降触发疑似，稳定后纠回）。这把"Z 忽大忽小但没真跌倒 → 沙发"的洞察量化到了雷达状态机层面。

### 4.3 Pose 分组（减维）

用于累计时长：

| Group | 包含 pose | 用途 |
|---|---|---|
| Move | 1, 12 | Flow 证据 |
| StandSit | 3, 4 | 沙发签名 |
| Lie | 6 | 床 |
| Floor | 8 | 地面坐姿 |
| BedSit | 9, 10, 11 | 床事件 |

---

## 5. 时空双分层（核心架构）

### 5.1 两条独立的数据流

```
即时流（近 5 秒）: 运动学判断
  track.History 滚动窗口（5-10 帧）
  Kalman Predict/Update
  残差、速度、Z 突降、空间跳跃
  → 答"这个 track 现在在干什么"
  数据只活在 track object 里，track 消失就清

历史流（24h / 7d 半衰期）: 区域属性学习
  cell.RealDecay / GhostDecay / ZVar / DwellEMA / PoseHist[] / Events[].*
  cell.Belief[3] 后验
  → 答"这个位置长期是什么"
  数据落在 cell 上，衰减但不清除
```

两条流**不互相污染**，在 engine 层交叉使用。

### 5.2 没有点云

**雷达每帧给一个 track 一个 (x, y, z)，不是点云**。

- 站立时反射中心在胸部附近
- 翻身/摔倒时反射中心跳到躺下身体某处（投影 XY 可能跳 80cm+）
- 呼吸/小动作时反射中心小幅抖动 ±30cm

**一个点没法表达人体**。Engine 靠 **每帧单点 mark + 时间累积 + 多 track 样本** 三位一体，让"人体占多个 cell"的事实自然涌现——**不做点云聚合、不做扩散半径**。

### 5.3 每帧双维度喂

每个轨迹点 `f = (x, y, z, pose, t)` 到达时，**同时**喂两个维度：

```
f
│
├── 维度 A: 即时流（track 层）
│   ├ push 进 track.History
│   ├ Kalman Predict(dt) → Update(x,y) → residual
│   ├ 算本帧质量 q（残差/速度/一致性）
│   └ 更新 Score / Verdict / Anomaly
│
└── 维度 B: 历史流（cell 层）
    ├ 按本帧 q 分桶：
    │   q > 阈值 → cell.RealDecay++  + FlowX/Y
    │   q < 阈值 → cell.GhostDecay++
    ├ 无条件累积（不看 q）：
    │   ZMean / ZVar (Z 观测统计)
    │   PoseHist[group] (pose 时长)
    │   Events[].* (两阶段事件)
    │   PoseMismatch（本帧若检测到矛盾则累加）
    └ 事件触发时额外调用 MarkEvent
```

**关键：历史流不过滤**。
- 不用 Verdict 作闸门
- Ghost 点也累积（进 GhostDecay 桶，有正面价值：识别反射热点）
- 试用期的点也累积
- 每一帧都是历史证据

时间衰减和后验 UpdateBelief 会自动淘汰不一致的信号。

### 5.4 单帧单 cell mark

每帧雷达给一个点 → 找到对应的**单个 cell** → 累积：

```go
c := grid.CellAt(f.X, f.Y)
c.RealDecay or GhostDecay += 1
c.ZMean = EMA(c.ZMean, f.Z)
c.ZVar  = EMA_residual_sq(c.ZVar, f.Z - c.ZMean)
c.PoseHist[poseGroup(f.Pose)] += dtSec
...
```

不扩散到邻居。不聚合成椭圆。**时间和样本做空间聚合**：
- 人坐沙发 1 小时：(x,y) 抖动 ±10cm → 自然落到 1~3 个 cell 累积
- 多人多次坐同沙发：36 个 cell 逐渐都有 DwellEMA
- 6 个月后：整片沙发区域涌现出 Sofa 判定

### 5.5 交叉使用的场景

| 判定 | 即时流 | 历史流 | 组合 |
|---|---|---|---|
| Ghost 出生 | 5 因子打分、残差 | 出生 cell 的 GhostRatio 历史 | 两者都扣分则判 Ghost |
| Silent Fall | 消失前 3 帧 Z<20、空间跳跃 | 消失 cell 的 EdgeDist、FallEventCount | 位置可疑 + 轨迹可疑 = 报警 |
| Fake Fall（沙发）| Fall 事件触发 | 该 cell 的 Retract 累积、StandSit 占比 | 历史提示"沙发签名" → 降级报警 |
| 静止超时 | StillSince 满 15min | 该 cell 的 Belief.Type | 历史判 Bed 则不报 |

---

## 6. Track 层：Kalman 运动模型与连续性打分

Cell 是"静态底片"的最小单位；Track 是"运动的点"。底片决定判定的先验，Track 即时判定出本帧质量 `q`，`q` 决定历史累积的分桶。

### 6.1 为什么用 Kalman：统一的连续性度量

人的运动是平滑的（加速度有界），ghost 和目标跳变不服从。Kalman 残差（innovation）同时捕获：

- 空间跳跃（残差极大）
- 物理不可能的速度（K 更新后速度超过人类上限）
- Ghost 出生漂移（残差方向与已知 flow 相反）

### 6.2 状态与观测模型

**状态向量**（4 维）：

```
x = [px, py, vx, vy]ᵀ     单位：px,py cm ；vx,vy cm/s
```

**匀速运动模型**（老人步速 0.3–1.0 m/s，1 Hz 雷达帧）：

```
F(dt) = [1  0  dt 0]        H = [1 0 0 0]
        [0  1  0  dt]            [0 1 0 0]
        [0  0  1  0]
        [0  0  0  1]
```

### 6.3 噪声参数（物理标定）

| 参数 | 值 | 物理含义 |
|---|---|---|
| `QAccel` | **30 cm/s²** | 老人加速度上限（过程噪声）|
| `RStd`   | **20 cm**    | 雷达定位标准差（观测噪声）|

`Q(dt) = QAccel² × 4×4 连续白噪声加速度矩阵`；`R = RStd² × I₂`。

### 6.4 初始协方差 P₀

| 元素 | 值 | 物理含义 |
|---|---|---|
| `P[0][0] = P[1][1]` | **400** | 位置 σ = 20 cm |
| `P[2][2] = P[3][3]` | **2500** | 速度 σ = 50 cm/s（开局未知）|

### 6.5 Predict / Update 方程

```
x̂⁻ = F(dt) · x̂                 // 状态预测
P⁻  = F(dt) · P · Fᵀ + Q(dt)    // 协方差预测

y = z - H·x̂⁻                    // 残差（innovation）—— 核心连续性量
S = H·P⁻·Hᵀ + R
K = P⁻·Hᵀ·S⁻¹
x̂ = x̂⁻ + K·y
P = (I - K·H)·P⁻
```

### 6.6 本帧质量分 q → 历史流分桶

残差 `|y|` 作为每帧质量分 `q`：

| `|y|` 范围 | q 含义 | track Score Δ | 历史流归桶 |
|---|---|---|---|
| < 30 cm | 运动平滑 | **+2** | RealDecay + FlowX/Y |
| 30–80 cm | 中等偏差 | 0 | RealDecay + FlowX/Y |
| 80–200 cm | 较大偏差 | **-5** | RealDecay（弱）或 GhostDecay |
| > 200 cm | 空间跳跃 | **-20** | GhostDecay |

**注意：历史流分桶不等 Verdict**——每帧就按本帧 q 直接累积。Verdict 只管 track 本身要不要报 alarm。

### 6.7 其他关键常量

| 常量 | 值 | 含义 |
|---|---|---|
| `HistoryLen` | **30** | 滚动窗口帧数（30 秒，用于回溯）|
| `MotionWindowSec` | **5** | 运动学判定窗口（近 5 秒）|
| `ProbationFrames` | **5** | 试用期帧数 |
| `StillThreshCm` | **15** | 帧间位移 < 15 cm 视为静止 |
| `ScoreConfirmTh` | **50** | Score ≥ 50 → Verdict = Real |
| `ScoreGhostTh` | **20** | Score < 20 → Verdict = Ghost |
| `MissCount` 上限 | **10** | 连续 10 帧无观测 → 消失 |

### 6.8 5 因子出生评分（track 首帧一次性打）

在初始 50 基础上：

| 因子 | 判据 | Score Δ |
|---|---|---|
| **d_enter** | < 50 cm：紧挨入口 | **+20** |
|  | 50–150 cm | +10 |
|  | 150–300 cm | -10 |
|  | > 300 cm：凭空出现 | **-25** |
| **grid_ghost** | 出生格 GhostRatio > 0.7 | **-25** |
|  | 0.4–0.7 | -10 |
|  | 出生格 IsEntry = true | **+15** |
| **n_co** | 房间已有 0 track | +10 |
|  | ≥ 2（伴生嫌疑）| -10 |

### 6.9 ProcessFrame 五段（双维度版）

```
┌──────────────────────────────────────────────────────┐
│ 段 1: 本帧观测到的 track                             │
│   ├─ 新 track: NewTrackState + scoreBirth()          │
│   └─ 已存在: Kalman.Predict(dt) → Update(x,y)        │
│       → 算本帧 q (残差)                              │
│       → 维度 A 即时：Score/Anomaly/pose 状态机       │
│       → 维度 B 历史：grid.MarkOccupancy(x,y,q,...)   │
│                     grid.MarkObservation(x,y,z,pose) │
│                     grid.MarkEvent(若状态机触发)     │
├──────────────────────────────────────────────────────┤
│ 段 2: 本帧未观测到的 track                           │
│   ├─ Kalman.PredictOnly(dt), MissCount++             │
│   └─ MissCount > 10 → 消失                           │
│      └─ Real 且非 Door/EdgeDist 大 → Silent Fall     │
├──────────────────────────────────────────────────────┤
│ 段 3: 试用期判定                                     │
│   FrameCount ≥ 5 → Score ≥ 50 Real / < 20 Ghost      │
├──────────────────────────────────────────────────────┤
│ 段 4: 构建 TrackOutput                               │
├──────────────────────────────────────────────────────┤
│ 段 5: 风险分 / alarm 触发                            │
│   (旧版"Real 才反哺 grid" 已合并到段 1)              │
└──────────────────────────────────────────────────────┘
```

### 6.10 风险分计算（computeRisk）

```
base = 100 (Fall) / 80 (PathBreak) / 70 (PoseMismatch) / 60 (StillTooLong) / 0 (else)

timeFactor  = 2.0  (5:00-6:00，生理低谷)
            = 1.5  (22:00-5:00，夜间)
            = 1.3  (6:00-8:00，晨起)
            = 1.0  其他

occFactor   = 1.5  (独居)
            = 1.2  (两人一人在床)
            = 1.0  (多人)

risk = clamp(base × timeFactor × occFactor, 0, 100)
```

---

## 7. Cell 证据与信念字段（完整清单）

Cell 的字段分 **几何（静态）** / **物理（rasterize 烤入）** / **证据（共享，带衰减）** / **信念（3 份并行）** 四类。

| 分类 | 字段 | 类型 | 含义 | 衰减? |
|---|---|---|---|---|
| 几何 | `InRoom` | bool | 在墙多边形内 | — |
|  | `InFOV` | bool | 在雷达物理信号可达域内 | — |
| 物理（rasterize）| `EdgeDist` | int | 到信号域边缘的 XY 距离（cm）| — |
|  | `MaxZ` | int | 该位置雷达可探测目标的 Z 上限（cm）| — |
|  | `MinZ` | int | Z 下限（cm）| — |
| 证据：访问 | `RealDecay` | int | 本帧 q 高的累计 | ✓ |
|  | `GhostDecay` | int | 本帧 q 低的累计 | ✓ |
| 证据：运动静止 | `FlowX` / `FlowY` | int | 平均速度向量 EMA | ✓ |
|  | `StillCount` | int | 累计静止秒 | ✓ |
|  | `DwellEMA` | int | 单次停留 EMA（秒）| ✓ |
|  | `LongStillCount` | int | 静止 > 15min 次数 | ✓ |
| 证据：Z 抖动 | `ZMean` | int | Z 高度 EMA 均值 | ✗ |
|  | `ZVar` | int | Z 残差² EMA | ✓ |
| 证据：姿态 | `PoseGroupSec[5]` | int | 各分组累计秒 | ✓ |
|  | `PoseMismatch` | int | pose 与运动学矛盾次数 | ✓ |
| 证据：两阶段事件 | `Events[3].Suspected` | int | 疑似次数 | ✓ (长半衰期)|
|  | `Events[3].Confirmed` | int | 确认次数 | ✓ (长半衰期)|
|  | `Events[3].Retract` | int | 回撤次数 | ✓ (长半衰期)|
| 证据：业务事件 | `BedEventCount` | int | In/LeftBed | ✓ |
|  | `DoorEventCount` | int | Enter2out | ✓ |
| 信念 | `Belief[3].Type` | CellType | 当前判定类型 | — |
|  | `Belief[3].Confidence` | int | [0,100] | ✗ 数据驱动 |
|  | `Belief[3].RiskScore` | int | [0,100] | — |
|  | `Belief[3].Source` | Source | 人 / 学 / 几何 | — |
| 元 | `LastUpdateMs` | int64 | 最后更新时间 | — |

**衰减半衰期**：普通证据 **24 小时**；事件计数 **7 天**。

**关键说明：**
- Cell 字段都是**布尔/计数/EMA**，没有概率。
- 证据字段**所有 3 组参数共享**；Belief[3] 独立演化。
- `Confidence` 不随时间衰减——只被数据驱动升降。
- 物理字段（EdgeDist/MaxZ/MinZ）在 Grid 构建时**一次性 rasterize**，运行时只读。

---

## 8. CellType 单一类型系统

取消"静态 Type vs InferredType"的双轨，合并为一套：

```
CellOpen       未知开放空间
CellDoor       门/入口
CellBed        床
CellChair      椅子/沙发（人坐之处）
CellWall       墙
CellFOVEdge    雷达视场边缘
CellToilet     马桶
CellShower     淋浴
CellWalkway    通道（学习得到为主）
CellMetal      反射/干扰区（学习得到）
```

人标和 AI 学到的**共用同一套类型**，用 `Confidence + Source` 区分来源：

| Prior 来源 | Type | 初始 Confidence | Source |
|---|---|---|---|
| 人标 Bed / Door / Toilet / Wall | CellBed / CellDoor / … | **99** | Human |
| 人标 Sofa（允许粗标）| CellChair | **80** | Human |
| 雷达几何 → FOV 边 | CellFOVEdge | **99** | Geometry |
| 未指定 | CellOpen | **0** | Unset |
| UpdateBelief 翻转 | 推断出的 Type | `bestL × 50` | Learned |

**Confidence 不是硬锁**——99 而非 100，留 1 分让系统有机会翻转（床被搬走）。

---

## 9. 信念更新算法（UpdateBelief）

### 9.1 单组更新流程

```
for cell in all cells:
    1. 观测量不足 → return
    2. 算所有候选类型的似然 L(T), T ∈ CellType
    3. best = argmax L(T)
    4. 分三种情况：
        a) best == cur:    Conf += α * (100 - Conf) * L_cur
        b) best != cur 且 Conf < flipTh:
              Type   = best
              Conf   = L_best * 50
              Source = Learned
        c) best != cur 但 Conf 还高:
              Conf -= β * (L_best - L_cur) * 100
    5. 更新 RiskScore
```

`α` 小（~0.02）让一致证据缓慢增强；`β` 大（~0.5）让矛盾证据快速削弱。

### 9.2 各类型似然公式（都读历史累积字段）

```
L(CellOpen)     = 0.3                                    // 常数底
L(CellDoor)     = clamp(DoorEventCount / 5, 0, 1)
L(CellBed)      = clamp(DwellEMA/600, 0, 1) × LieRatio
                + clamp(BedEventCount/3, 0, 1) × 0.5
                + BedSitRatio × 0.3

L(CellChair)    = clamp(ZVar/50, 0, 1)
                × clamp(DwellEMA/300, 0, 1)
                × (StandSitRatio + 0.2)
                × (1 + FallRetract/(Suspected+1))       ← 核心沙发签名
                × clamp(1 - FallConfirmed/5, 0, 1)      ← 真 fall 扣分

L(CellWalkway)  = clamp(|Flow|/50, 0, 1)
                × clamp(1 - DwellEMA/10, 0, 1)
                × (MoveRatio + 0.1)

L(CellMetal)    = GhostRatio
                × clamp(PoseMismatch/20, 0, 1)
                × (1 + clamp(AllRetract/10, 0, 1))
```

**沙发签名的核心**：雷达频繁疑似 fall 但都回撤 + Z 抖动 + 长 Dwell + 站坐姿主导。不依赖 ground truth 就能判定。

---

## 10. 触发策略

### 10.1 分层

| 层 | 频率 | 目的 |
|---|---|---|
| 定时全图扫 | 每 5 分钟 | 基线：所有 cell 跑一遍 UpdateBelief |
| 事件点触发 | 即时 | 只更新事件相关 cell + 半径 30cm 邻居 |

### 10.2 触发事件清单

| 事件 | 条件 | 作用 | 行动 |
|---|---|---|---|
| pose = Fall(5) | 雷达直报 | 落点 cell | Events[Fall].Confirmed++; UpdateBelief |
| SuspectedFall→Fall | 状态转移 | 同上 | 已计在上条 |
| SuspectedFall→非 Fall 超时 | 状态回撤 | 落点 cell | **Events[Fall].Retract++** |
| pose = 8 (SittingOnGround) | 确认坐地 | 落点 cell | Events[Floor].Confirmed++ |
| BedSitUp 家族 | 事件 | 落点 cell | Events[BedUp].* |
| In/LeftBed | Sleepad 事件 | 坐标 cell | BedEventCount++ |
| Enter2out | 门事件 | 坐标 cell | DoorEventCount++ |
| 静止 > 15 min | track_manager | 静止 cell | LongStillCount++ |
| **Silent Fall** | 合成规则 | 消失 cell | Events[Fall].Confirmed++ + 输出 alarm |

### 10.3 Silent Fall 检测（5 要素）

雷达漏报的 fall，通过"违规消失"反推：

```
1. Verdict == VerdictReal                              // 真 track
2. AgeSec > 10                                         // 短暂闪现不算
3. 消失前 3 帧内 min(Z) < 20 cm                        // 贴地（短时流）
4. 消失 cell.EdgeDist > 30 cm 且 非 CellDoor           // 物理上不是合法离场
5. 消失前最后两帧 |Δpos| > 80 cm                       // 空间跳跃
   ↑ 80cm 跳跃 + Z 急降 = 真跌倒信号（反射中心从站立→躺下）
   ↑ 80cm 跳跃 + Z 不降 = 更可能是 Ghost
```

触发后：
- `Events[Fall].Confirmed++` 在消失 cell
- 生成 `AnomalyFall` 输出，带 `source=engine_silent` 标记
- App 提示 "AI 推测跌倒，请确认" 交家属反馈

---

## 11. 自适应参数（三组并行 + winner 选择）

### 11.1 三组候选

| 组 | Alpha | Beta | FlipTh | 策略 |
|---|---|---|---|---|
| 0 Conservative | 0.01 | 0.2 | 10 | 慢学慢忘，抗噪 |
| 1 Balanced | 0.02 | 0.5 | 20 | 中庸 |
| 2 Aggressive | 0.05 | 1.0 | 30 | 快速响应，可能抖 |

### 11.2 存储策略

证据字段**单一共享**（观测事实只有一个）；每个 cell 维护 `Belief[3]` 三份信念独立演化。

开销：10,000 cells × 3 × ~48 bytes ≈ 480 KB / 房间，可忽略。

### 11.3 Winner 选择规则

每 24 小时（或累计 100 条 ground truth）评估：

```
for g in 0..2:
    accuracy[g] = (TP + TN) / (TP + TN + FP + FN)

winner_new = argmax accuracy[g]

// AI 必须超雷达 baseline 20% 才启用
if accuracy[winner_new] > accuracy[radar_baseline] + 0.20:
    engine.winner = winner_new
else:
    engine.winner = fallback_to_radar

// 稳定后可以只跑 winner 组节省开销
if winner 连续 7 天 accuracy > 阈值:
    stable mode，只跑 winner
    // 准确率下降 > 20% 时重新激活 3 组
```

---

## 12. Ground Truth 回流（方案 B：扫表）

### 12.1 为什么需要

没有 ground truth 就无法知道 AI 准确率，也无法比较 3 组参数。Silent Fall 必须上报，就是为了拿家属反馈。

### 12.2 回流通路

```
Engine 输出 alarm → App
家属点"真跌倒 / 误报" → data 服务写 alarm_events.resolution
Engine 每 5 分钟扫表 → 读 resolution != null 的行
→ 按 cell 位置喂给 AccuracyTracker[3]
→ 评分驱动 winner 切换
```

**DB 改动**：

```sql
ALTER TABLE alarm_events ADD COLUMN resolution VARCHAR(16) DEFAULT NULL;
-- 取值: "real" / "fake" / NULL
ALTER TABLE alarm_events ADD COLUMN resolution_ts BIGINT DEFAULT NULL;
```

### 12.3 Ground truth 三个来源

| 来源 | 权威性 | 数据形式 |
|---|---|---|
| 家属/护工 App 反馈 | **最高** | alarm_events.resolution |
| Sleepad In/LeftBed 事件 | 高 | iot:alarm:stream |
| 运维手动补标 layout | 高 | layout_config 更新 |

雷达原始 Fall 判定 = **baseline B**，不是 truth。

### 12.4 未来升级（方案 A）

新建 `iot:feedback:stream` 做实时回流，engine 订阅即可。二期再做。

---

## 13. 实施路线图（6 Phase）

### Phase 0 — 文档同步 ✓（本文即是）

### Phase 1 — 基础数据层

| 文件 | 动作 |
|---|---|
| `radarutils/signal.go`（新）| 物理 FOV 信号域：`SignalReachable` / `SignalEdgeDist` |
| `radarutils/zrange.go`（已有）| 校准：去除对 Boundary 的依赖 |
| `roomengine/cell.go`（小改）| 加 `EdgeDist` / `MaxZ` / `MinZ` 字段 |

### Phase 2 — Grid 坐标 + Rasterize

| 动作 | 说明 |
|---|---|
| `RoomGrid` 加 `OriginX/OriginY` | 支持画布坐标（默认 `-RoomW/2, 0`）|
| `ToIndex/CellAt/ToRoom/IsEdge` | 用 Origin 偏移 |
| `Entries` 改为 `[]radarutils.Rect` | 矩形而非点 |
| **构建时 rasterize 物理域** | 对每 cell 算 InFOV/EdgeDist/MaxZ/MinZ |
| 拆分 Mark 函数 | `MarkOccupancy(x,y,q,...)` / `MarkObservation(x,y,z,pose,...)` / `MarkEvent(x,y,kind,phase,...)` |
| 删除旧 `MarkRealVisit / MarkGhostBirth / MarkStill / MarkPoseMismatch` | |

### Phase 3 — Track 层改造

| 动作 | 说明 |
|---|---|
| ProcessFrame 每帧双维度喂 | 段 1 合并段 5（不再过滤 Verdict）|
| Pose 两阶段状态机 | Suspected/Confirmed/Retract 三路计数 |
| Silent Fall 检测 | 五要素 + `source=engine_silent` |
| LongStill 检测 | 静止 > 15min 上报 |

### Phase 4 — Engine 层

| 动作 | 说明 |
|---|---|
| `RoomConfig` 重构 | Enters/Beds/Toilets Rect + Walls + Radar RadarMount |
| Layout JSONB parser | 解析 `{objects, radar, sleepad, sensor}` |
| `ParamSet[3]` + `AccuracyTracker[3]` + winner | |
| 定时全图 UpdateBelief（每 5min goroutine）| |
| 事件点触发 UpdateBelief | 订阅 iot:alarm:stream |

### Phase 5 — 反馈闭环

| 动作 | 说明 |
|---|---|
| `feedback_loop.go`（新）| goroutine 每 5min 扫 alarm_events.resolution |
| Ground truth 喂 AccuracyTracker[3] | 按 cell 位置匹配历史预测 |
| SQL migration | 加 resolution / resolution_ts 列 |

### Phase 6 — 集成验证

| 动作 | 说明 |
|---|---|
| `go build ./...` | wisefido-ai + owl-common 必须都过 |
| Log 回放 | 拿 wisefido-qinglan.log 的 track 数据灌 engine |
| 关键单元测试 | Kalman 残差 / Silent Fall / UpdateBelief / 坐标转换 |

---

## 14. 关键设计决策与权衡（摘要）

| 决策 | 选择 | 理由 |
|---|---|---|
| Cell vs 矩形/圆形 | Cell（10cm 粗量化）| 统一坐标 + 降精度是工程本质，不是拟人 |
| Cell 字段类型 | **布尔/计数**，不存概率 | 模糊由时空累积涌现，不在单 cell 内表达 |
| 点云聚合 | **不做** | 雷达没给点云，每帧一个点直接 mark 一个 cell |
| 每帧双维度 | 即时流 + 历史流独立喂 | 运动判定和区域学习解耦 |
| 历史流过滤 | **不过滤**，按 q 分桶 | Ghost 累积也是信号（反射热点）|
| Radar Boundary | UI 参考，不作判定 | 物理 FOV（hfov/vfov/afov/efov/installHeight）才是真相 |
| 物理 FOV 计算 | Rasterize 到 cell 字段 | 公式一次算，O(1) 查询 |
| 坐标系 | 画布坐标（顶部中心原点）| 与前端 layout_config 零转换对齐 |
| 命名规范 | Rotation/HFOV/VFOV/AzimuthFOV/ElevationFOV | 与前端 types.ts + 标准雷达术语对齐 |
| 数值类型 | **全 int**（坐标/角度/计数/百分比）| 10cm 量化下小数无物理意义；仅 Kalman 内部矩阵 + 似然中间量 + 衰减系数保留 float |
| 人标 vs 学习 | 共用 CellType，用 Confidence 区分 | 床会被搬，不应有不可变类型 |
| Room vs FOV 几何 | 分开，两个 bool 位 | 物理边界和传感边界本来就不重合 |
| Confidence 衰减 | 不衰减 | 数据驱动更符合"有矛盾才动摇"直觉 |
| 运动模型 | 2D 匀速 Kalman | 老人低加速度，匀速 + Q 噪声够用 |
| α/β 参数 | 三组并行 + winner | 不拍脑袋，让数据选参数 |
| winner 切换门槛 | AI 超雷达 20% | 避免小幅胜出时切换 |
| Silent Fall 是否报警 | **报** | 不报就没反馈，没反馈就不知道准确率 |
| Ground truth 通路 | 方案 B 扫表 | 工作量小，能立刻跑；方案 A 二期升级 |

---

## 15. 待确认与后续扩展

1. 稳定期后 2 组 shadow 是否完全停跑省 CPU，还是保留低频跑以便劣化时快速切回——倾向后者，每 30 分钟跑一次 shadow。
2. RiskScore 的下游消费者（前端热力图 / 运维建议）接口待定。
3. 连通域聚合（把相邻高 Sofa 分 cell 合并成"整块沙发"）放 engine 的分析 pass，不进 Cell——二期。
4. 人标 Sofa "中心 80/边缘 40" 分级 Confidence——骨架内先统一 80，二期细化。
5. Kalman 从匀速升级到匀加速（CA 模型）或 IMM 多模型——二期。
6. 物理 FOV 的 r_max 信号最大作用距离：目前从 Boundary 最远边 fallback，等 device_meta 暴露硬件规格后精化。
