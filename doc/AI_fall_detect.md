# AI 跌倒检测与房间认知（Room Engine 设计整合）

## 0. 背景与问题

当前 radar 的跌倒检测存在两个核心痛点：

1. **大量 ghost track 造成假警报**——雷达的 track 会凭空出现、分裂、跳变，导致跌倒误报。
2. **真实 fall 会被漏报**——人跌倒后贴地，雷达可能直接丢失目标，看起来像"track 消失"，不被报成 Fall。

单纯靠雷达原始 pose 判定（pose=5 Fall）既多虚报也有漏报。需要一个**房间级空间认知模块**，把每个房间当作一张自学习的"底片"，结合 radar 原始信号、业务事件（床/门）、家属反馈，给出更准确的异常判定。

模块代号：**Room Engine**，位于 `owlBack/wisefido-sensor/internal/roomengine/`。

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
                                  cell 多维字段：Real/Ghost/ActiveType[4]/Flow/Dwell
                                               /FallEvent/LieRetract/Sleepad/Door...
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

### 3.3 雷达视场（本期：用 Boundary 多边形；二期再精细化物理 FOV）

**本期简化策略**：直接用前端画的 Radar signal（Boundary 多边形）+ InRoom 交集作为可达域。
公式判定留给二期精细化。物理字段（HFOV/VFOV/AzimuthFOV/ElevationFOV）保留但不参与判定，作为未来扩展位。

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
| 计时累计（Dwell） | 秒（int）| 含 DwellEMA |
| 姿态累计（ActiveType[4]）| 饱和计数（uint8 0-255）| Move/Stand/Sit/Lie |
| 速度（FlowX/Y）| cm/s（int）| EMA |
| 百分比（Confidence / Score / RiskScore）| 0-100（int）| |
| 次数累计（Real/GhostDecay / LieRetract / FallEvent / Sleepad / Door / LongStill）| 次（int）| 带衰减，整除到 0 即遗忘 |

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

历史流（15min 短档 / 7d 长档 半衰期）: 区域属性学习
  cell.RealDecay / GhostDecay / FlowX/Y / DwellEMA / ActiveType[4]
  cell.FallEventCount / LieRetract / SleepadInBedCount / DoorEventCount 等
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
    ├ 按本帧 q 分桶（不看 Verdict）：
    │   grid.MarkOccupancy(x, y, q, vx, vy)
    │     q ≥ 50 → RealDecay++ + FlowEMA
    │     q < 50 → GhostDecay++
    ├ 无条件累积：
    │   grid.MarkPoseTime(x, y, core, dtSec)   // ActiveType[core] 饱和累加
    ├ 事件触发（Track 层检测后调）：
    │   grid.MarkFallEvent / MarkLieRetract / MarkDwell / MarkLongStill
    │   grid.MarkSleepadInBed / MarkSleepadLeftBed / MarkDoorEvent
    └ Z 抖动 / PoseMismatch 归 Track 层，不累计到 cell
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
// q 分桶
if q >= 50 {
    c.RealDecay++
    c.FlowX = (9*c.FlowX + vx) / 10
    c.FlowY = (9*c.FlowY + vy) / 10
} else {
    c.GhostDecay++
}
// ActiveType 饱和累加
idx := CorePoseToActiveIdx(RadarPoseToCore(f.Pose))
if idx >= 0 {
    c.ActiveType[idx] = satAddUint8(c.ActiveType[idx], dtSec)
}
// ... 事件条件触发则 MarkFallEvent / MarkLieRetract / MarkDwell / ...
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
| 证据：运动/停留 | `FlowX` / `FlowY` | int | 平均速度向量 EMA (cm/s) | ✓（短档）|
|  | `DwellEMA` | int | 单次停留 EMA（秒）| ✓（短档）|
|  | `LongStillCount` | int | 静止 > 15min 次数 | ✓（长档）|
| 核心姿态 | `ActiveType[4]` | [4]uint8 | 饱和累加 0-255 [Move/Stand/Sit/Lie] | ✓（短档）|
|  | `AreaType` | AreaType | 推断属性镜像 | — |
| 姿态转换 | `LieRetract` | int | 短暂 Lie 回撤（沙发签名）| ✓（长档）|
| 自推事件 | `FallEventCount` | int | Stand/Move → Lie + Z 骤降 | ✓（长档）|
| 外部传感器 | `SleepadInBedCount` | int | HR/RR/InBed | ✓（长档）|
|  | `SleepadLeftBedCount` | int | LeftBed | ✓（长档）|
|  | `DoorEventCount` | int | Enter/Exit | ✓（长档）|
| 信念 | `Belief[3].Type` | AreaType | 当前判定类型 | — |
|  | `Belief[3].Confidence` | int | [0,100] | ✗ 数据驱动 |
|  | `Belief[3].RiskScore` | int | [0,100] | — |
|  | `Belief[3].Source` | Source | 人 / 学 / 几何 | — |
| 元 | `LastUpdateMs` | int64 | 最后更新时间 | — |

**衰减半衰期**：短档 **15 分钟**（ActiveType/Real/Ghost/Flow/Dwell）；长档 **7 天**（所有事件类）。

**关键说明：**
- Cell 字段都是**布尔/uint8/int**，没有概率。
- 证据字段**所有 3 组参数共享**；Belief[3] 独立演化。
- `Confidence` 不随时间衰减——只被数据驱动升降。
- 物理字段（EdgeDist/MaxZ/MinZ）在 Grid 构建时**一次性 rasterize**，运行时只读。
- **Z 抖动 / PoseMismatch 归 Track 层**（track 内部 `ZNoiseCount` / `PoseMismatchCount`），不累计到 cell。

---

## 8. AreaType 类型系统（7 种，功能分类）

放弃按"家具名"分类，改按**engine 要问的问题**分类：

```
AreaUnknown (0)  未知
AreaEnter   (1)  进入区（门）
AreaBed     (2)  可躺区（床）
AreaSit     (3)  允许坐姿（沙发/椅子）
AreaActive  (4)  活动区（走廊、开放空间）
AreaDeny    (5)  禁止 track（墙/家具/Metal，track 出现 = Fake）
AreaShower  (6)  淋浴区（高风险：潮湿+跌倒）
AreaToilet  (7)  马桶（高风险：起身晕眩）
```

与功能对应：

| AreaType | StillTimeoutSec | IsRestZone | 风险特征 |
|---|---|---|---|
| Enter | 8 min | 否 | 正常进出 |
| Bed | 不限 | ✓ | 长躺姿 |
| Sit | 不限 | ✓ | 长坐姿 |
| Active | 8 min | 否 | 移动/短停 |
| Deny | 5 min | 否 | track 不该出现 |
| Shower | **15 min** | 否 | 高风险 |
| Toilet | **15 min** | ✓ | 高风险 |

人标和 AI 学到的**共用同一套类型**，用 `Confidence + Source` 区分来源：

| Prior 来源 | Type | 初始 Confidence | Source |
|---|---|---|---|
| 人标 Bed / Enter / Toilet / Shower / Wall | AreaBed / AreaEnter / … | **99** | Human |
| 人标 Sofa（允许粗标）| AreaSit | **80** | Human |
| 未指定 | AreaUnknown | **0** | Unset |
| UpdateBelief 翻转 | 推断出的 Type | `bestL × 50` | Learned |

**Confidence 不是硬锁**——99 而非 100，留 1 分让系统有机会翻转（床被搬走）。

**Shower / Toilet 默认不参与自动学习**（似然为 0），只认人标——它们形状单一、位置固定，硬锚性强。

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

### 9.2 各类型似然公式（基于 ActiveType 分布 + 事件累计）

设 `ratio(i) = ActiveType[i] / Σ ActiveType`，则 moveR/standR/sitR/lieR 为对应占比。

```
L(AreaUnknown) = 0.2                                    // 常数底

L(AreaEnter)   = clamp(DoorEventCount / 5, 0, 1)

L(AreaBed)     = lieR
               + (SleepadInBedCount > 0 ? 0.8 : 0)       // Sleepad 决定性证据
               + clamp(LongStillCount / 3, 0, 1) × 0.3

L(AreaSit)     = sitR × clamp(DwellEMA/300, 0, 1)
               × (1 + LieRetract/5)                      // ← 核心沙发签名：短暂 Lie 回撤多
               × clamp(1 - FallEventCount/5, 0, 1)        // ← 真 Fall 扣分

L(AreaActive)  = moveR
               + clamp(|Flow|/50, 0, 1) × 0.3
               + clamp(1 - DwellEMA/10, 0, 1) × 0.2
               + standR × 0.2

L(AreaDeny)    = GhostRatio
               + (若总 ActiveType 累积 < 5 且 RealDecay < 2: +0.3)

L(AreaShower)  = 0   // 不参与自动学习（形状固定，只认人标）
L(AreaToilet)  = 0   // 同上
```

**沙发签名的核心**（与原设计一致，但现在用 LieRetract 替代 Events[Fall].Retract）：
短暂 Lie 回撤多（LieRetract 高）+ 长 Dwell + 坐姿主导 + 真 Fall 少。不依赖 ground truth 就能判定。

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
| Sleepad InBed | Sleepad 压床 | 事件坐标 | SleepadInBedCount++ |
| Sleepad LeftBed | Sleepad 离床 | 事件坐标 | SleepadLeftBedCount++ |
| Enter2out | 门事件 | 坐标 cell | DoorEventCount++ |
| 静止 > 15 min | track_manager | 静止 cell | LongStillCount++ |
| **Silent Fall** | 合成规则 | 消失 cell | Events[Fall].Confirmed++ + 输出 alarm |

### 10.3 Silent Fall 检测（5 要素）

雷达漏报的 fall，通过"违规消失"反推：

```
1. Verdict == VerdictReal                              // 真 track
2. AgeSec > 10                                         // 短暂闪现不算
3. 消失前 3 帧内 min(Z) < 20 cm                        // 贴地（短时流）
4. 消失 cell.EdgeDist > 30 cm 且 非 AreaEnter          // 物理上不是合法离场
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
| `go build ./...` | wisefido-sensor + owl-common 必须都过 |
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

---

## 16. 已知误报场景

### 16.1 Kitchen 长时间站立做饭 → lost-fall 误报

**场景**：人在厨房 counter / stove 前**长时间站立**做饭（通常 15-60min 不挪动），雷达可能报 track 短暂消失（pose=Stand 静止久 → 失锁），触发 lost-fall pending。5min 等待期内若无新 track 出生 + 无 ExitRoom event → 报 lost-fall。

**实测**：D523 bookroom 3 天回放 47 次 lost-fall（layout boundary 修复前）；Kitchen 3 天 11 次。多数集中在做饭时段（午餐/晚餐），位置在 counter 前。

**对 elder care 场景的影响**：**不适用**。
- 老人 elder care 场景下，厨房不是主要活动区（多由家属/护工代劳）
- 老人若进厨房，通常**短暂**（取水/取药/微波热菜），不会站 15min+
- 该误报模式仅在"老人独居 + 自己做饭 + 厨房有 radar"才显著触发

**部署建议**：
- 老人独居 + 厨房雷达：建议在厨房 layout 显式标 Counter / Stove 区为 `Furniture`（AreaDeny），让 radar 在该区消失视为正常（cell 是 Deny）
- 或考虑在 kitchen room 的 stay-alarm config 关闭 still-fall 触发
- 通用养老机构（多人厨房）：该房间不部署 radar，集中在卧室/卫生间/客厅监测

### 16.2 雷达近场 pose 误判 → AreaSit 误学（PR-15.3 前）

旧 `cell_learning` 简单规则 "ActiveType[Sit] >= 15s → AreaSit" 在雷达近场（≤1.5m）易误学（z 测距不准）。
**已修复**（PR-15）：移除该规则，AreaSit 仅由 PR-13 RegionStatic（2min Z-jump 或 12min 静止）+ PR-7.2 stand-static 学，要求"位姿切换硬证据"。

### 16.3 Auto-Deny island 核心识别需 ≥10 天数据（PR-15.5）

cell 距 walk path ≥2 cell 且持续 10 天无穿越才升 AreaDeny。新部署房间前 10 天 furniture 不会自动学到——需要：
- (a) Layout 显式标 Furniture 矩形（推荐）
- (b) 等 10 天数据自动积累

详见 PR-15.4 BFS 距离场设计。

---

## 17. AI Fall Rules 总结（PR1-5c 后）

四类 AI 派生 Fall，**category 统一 `alarm.Fall`**，通过 `Source` 字段区分子路径。
cardagg 现有 `case alarm.Fall` handler 自动落 alarm_events；子类型分析查
`trigger_data->>'source'`。

### 17.1 silent_fall — sleepad/radar 多模融合矛盾

**Source**: `ai_sleepad_radar_conflict`

**物理含义**：同房间 sleepad+radar 已确认人在床，sleepad 报 LeftBed 但 radar
仍看到人在床区 → 没真正离床（很可能从床边滑下，radar 仍捕捉到地面位置）。

**触发链（5 步）**：

1. **双源 InBed 入场门控**（PR-14 硬要求）：sleepad 报 InBed + radar 独立判
   定 InBed（area_id 落在 AreaBed cell），两事件 `|Δt| ≤ 15s` → 标记
   `RadarInBedConfirmedMs`。**未双源确认的 BedSession 永不触发** silent fall。
2. **多人计数**：同房 sleepad 各 +1 → `bedPersonCount`，BedSession 维护
   `MaxPeople`。
3. **Bed Area 锚定**：radar InBed 那帧的 `area_id` 作为该 Bed 几何锚。
4. **持续在床门槛**：`InBedSinceMs → LeftBedAtMs` ≥ `MinInBedSec = 5min`，
   过滤短暂坐床。
5. **LeftBed 后等待 + radar 矛盾判定**（两档等待时长）：
   - `LeftBedHadHRRR && LeftBedMaxPeople == 1` → **`WaitVitalSec = 60s`**
   - 其它 → **`WaitNoVitalSec = 120s`**
   - 等待窗满时任一 active radar track 距最近 AreaBed cell ≤
     **`BedNeighborhood = 100cm`** → 报 silent fall；已离开邻域 → 取消。

**参数（[fall_rules_param.go:108-113](../wisefido-sensor/internal/roomengine/fall_rules_param.go#L108-L113)）**：MinInBedSec=300s / WaitNoVitalSec=120s / WaitVitalSec=60s / BedNeighborhood=100cm。

---

### 17.2 bedside_fall — R4 床边晕倒

**Source**: `ai_bedside_silent`

**物理含义**：老人夜间 LeftBed 后**真离床走开了**（不像 silent_fall 还在床区），
但**没走远**就在床边停住，长时间静止——典型"起夜走两步晕倒"。

**与 silent_fall 区别（disjoint）**：

| 维度 | silent_fall | bedside_fall |
|---|---|---|
| LeftBed 后 radar 位置 | 仍在床 ≤100cm（**矛盾**）| 离床区但 ≤100cm（人真走开了，没走远）|
| 关键证据 | sleepad/radar 跨源不一致 | radar 自己的「长时间静止」 |
| 时间维度 | LeftBed 后 60-120s（短窗）| LeftBed 后 15min 静止（长窗）|
| 时段约束 | 全天 | **仅夜间** IsNightTime |

**触发链（4 个 AND）**：

1. **风险时段**：必须 `IsNightTime`（默认 23:30-07:30），白天不报。
2. **LeftBed 开窗**：最近 LeftBed 后 ≤ `WindowSec = 180s`（3 min）。
3. **位置约束**：track 距最近 AreaBed cell ≤ `BedsideMarginCm = 100cm`。
4. **静止时长**：> `StillTimeoutSec = 900s`（15 min）→ 报 bedside_fall。

**与 lost_fall 双报 corner case**：bedside_fall 触发时 track 仍活，后续若
firmware 丢 track 也会满足 lost_fall 条件 → 双报。**建议加 dedup**：
`ts.BedsideFallReported = true` 时 lost_fall pending 入池跳过。

---

### 17.3 still_fall — 长时间站立静止

**Source**: `ai_still_in_bathroom`

**物理含义**：老人在卫浴内 stand 姿态静止超过阈值——正常人不会站着不动 15
分钟以上。

**触发条件（5 个 AND）**：

1. **位置语义是浴室**——三条件**取并集**任一满足：
   - cell.Belief[0].Type ∈ {AreaToilet, AreaShower}
   - room.name 含 `bathroom` / `restroom` / `toilet`（不区分大小写）
   - alarm_device.monitor_config 显式启用 `Stay alarm`（运维标注的特殊房间）
2. **Pose=Stand**：避免坐马桶（pose=Sit）误报。
3. **位移静止**：`|Δpos| < StaticPosCm = 20cm`。
4. **静止时长超阈**：
   - 风险时段（夜）→ **15 min**（`ToiletShowerSec=900s`）
   - 非风险时段（白天）→ 15 × `NonRiskTimeFactor (1.2) = 18 min`
5. **Bathroom caregiver 例外**：cell ∈ {AreaToilet, AreaShower} 且同房间
   ≥ 2 个 real track → 大概率护工陪同，跳过。

**参数**：ToiletShowerSec=15min / DenyZoneSec=5min / DefaultSec=8min /
NonRiskTimeFactor=1.2 / StaticPosCm=20cm。

---

### 17.4 lost_fall — Real track 异常消失

**Source**: `ai_lost_track`

**物理含义**：track 在室内合法位置（非门口、非 Enter 区）被 firmware 丢失，
等待 cell-area-typed 时长后无"复现"（无新出生 + 无 ExitRoom + 无多人入屋）
→ 极可能跌倒倒地被低 z 信号失锁。

**触发条件（4 个 AND）**：

1. **track 状态**：消失前 `VerdictReal` 或 `VerdictPending`。
   - **Pending 也算**（用户对齐 2026-04-27）：还没确定是 ghost/real 时消失，
     按 real 计算（保守，宁可误报）。
   - `VerdictGhost` 不入池。
2. **位置约束**（消失瞬间最后位置）：
   - 不在 Enter 区（cell.Belief[0].Type != AreaEnter）。
   - **距最近 Enter 区 > `ExitDistMinCm = 30cm`**（**2026-04-30 从 100cm 收紧**）。
     - **改动动机**：原 1m 阈值下，门口 30cm-1m 区间真摔倒被当作"可能正常出门"
       漏掉。30cm 才是合理"贴在门口正常通过"的物理误差，>30cm 没 ExitRoom 都进
       fall pending。
3. **Track 年龄**：`AgeSec() ≥ 5s`，刚出生即消失不算（噪声）。
4. **等待窗满 + 无复现**：按消失点 cell.Belief[0].Type 分档：
   - AreaBed / AreaSit → **60 min**（睡觉/坐着丢 track 常见）
   - AreaToilet / AreaShower → **15 min**（与 still_fall 同步）
   - AreaDeny → **5 min**
   - 其它（Active / Unknown / Walk / Enter）→ **5 min**
   - **空间跳跃修正**：若消失前 `MaxImpliedSpeedFromBirth > SuspectSpeedCm = 100`
     → 等待时长 × `SpatialJumpFactor = 0.5`（更敏感，砍半）。
   - **取消条件**（任一）：新 track 出生附近（BlindSpotRecovery）/ ExitRoom
     事件 / 房间 number_people ≥ 2。

**参数（[fall_rules_param.go:121-130](../wisefido-sensor/internal/roomengine/fall_rules_param.go#L121-L130)）**：RestZoneWaitSec=60min / DenyZoneWaitSec=5min / WalkwayWaitSec=5min / **ExitDistMinCm=30cm** / SpatialJumpFactor=0.5。

**已知误报场景**（见 §16.1 Kitchen 长时间站立做饭）。30cm 收紧后 Kitchen 误报会**更频繁**，elder care 场景下"老人不下厨"可暂忽略，常规养老机构（多人厨房）建议 layout 显式标 Counter / Stove 为 Furniture (AreaDeny)。

---

### 17.5 一图对照

```
触发时刻 track 还在 tm.tracks ?
    │
    ├─ 是 → 看位置 + 静止特征
    │       │
    │       ├─ 浴室 + Stand + 静止 ≥15min → still_fall (ai_still_in_bathroom)
    │       │
    │       └─ 床边 ≤100cm + 静止 ≥15min + 夜间 + 在 LeftBed 窗内
    │          → bedside_fall (ai_bedside_silent)
    │
    └─ 否（firmware 已丢失 track）
            │
            ├─ sleepad LeftBed 矛盾路径（仅 BedSession 双源确认过）
            │   且 radar 仍在 Bed ≤100cm
            │   → silent_fall (ai_sleepad_radar_conflict)
            │
            └─ 一般消失 + 距门 >30cm + 等待窗满无复现
                → lost_fall (ai_lost_track)
```

### 17.6 wire 上的样子

发到 `iot:alarm:stream` 的统一形态（以 still_fall 为例）：

```json
{
    "device_uid": "9923003AB17F",
    "device_type": "Radar.AI01",
    "category": "Fall",
    "topic_type": "alarm",
    "data_value": [{
        "track_id": 0, "position_x": 50, "position_y": 320, "position_z": 80,
        "track_confidence": 60, "pose": 4,
        "dataCategory": "Fall", "event_name": "Fall",
        "source_device_uid": "9923003AB17F",
        "source": "ai_still_in_bathroom",
        "reason": "still_in_bathroom_over_threshold",
        "evidence": {
            "still_seconds": 920,
            "still_timeout_sec": 900,
            "cell_area_type": 6
        }
    }]
}
```

cardagg 端零改动接管（PR4 的 BaseDeviceType + 现有 case alarm.Fall），子类型
通过 `trigger_data->>'source'` 在 alarm_events 表里查询 / group by。
