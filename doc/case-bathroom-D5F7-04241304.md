# 案例：E598A2ACD5F7 浴室 — 吸顶模式 + 镜子反射 Ghost

- **设备 UID**：E598A2ACD5F7（Qinglan HC2）
- **安装模式**：**ceiling**（吸顶）
- **固件/MCU**：v2.3-Jun 2025 / v2.0-Nov 2025
- **房间名**：bathroom
- **录制窗口**：Denver 13:04 – 13:34（= UTC 2026-04-24 19:04 – 19:34）
- **帧数**：20 分钟 1370 帧（tid=0 真人 661 + tid=1 ghost 674 + tid=88 dummy 35）
- **场景价值**：**高风险 bathroom** + 小空间 + 镜子 Interfere 标注 + **长时间马桶坐姿**

---

## 1. 房间几何（从 layout 导出）

### Radar 安装

```
位置:   画布 (0, 240, 270)   吸顶 2.7m
模式:   ceiling
FOV:    leftH=190, rightH=120, frontV=70, rearV=90
```

### Wall（4 段 line，**开放下边界**）

```
Top    : (-180, 160) → (120, 160)
Left   : (-180, 160) → (-180, 310)
Right  : ( 120, 160) → ( 120, 310)
Bottom : (-180, 310) → ( 40,  310)    ← 只到 x=40，留出门洞
```

→ Wall 围出的主区域：**x ∈ [-180, 120], y ∈ [160, 310]**（300 × 150 cm = 4.5m²）

### Furniture / Interfere / Enter

| 对象 | 位置（画布）| 说明 |
|---|---|---|
| Enter（门）| (40,300)-(120,320) | 底部开口的门洞 |
| bath | (-150,180)-(-130,200) | 浴缸/花洒 |
| **Interfere mirror** | (0,160)-(120,170) | **已标注的镜子反射区** |
| Curtain line | (-100,160) → (-100,310) | 垂直浴帘 |

### 画布坐标系下的可达域

```
InRoom 真区  = x ∈ [-180, 120], y ∈ [160, 310]
InFOV 真区   = x ∈ [-120, 190], y ∈ [150, 310]
实际可达     = InRoom ∩ InFOV = x ∈ [-120, 120], y ∈ [160, 310]
                  ↑              ↑
            左半边 x ∈ [-180, -120] 是墙内但雷达 FOV 外（盲区）
            右半边 x ∈ [120, 190] 是 FOV 内但墙外
```

---

## 2. Track 汇总（20 分钟）

⚠️ **以下数据是雷达本地坐标** `(h, v)`，不是画布坐标。
转换公式（ceiling 模式, rotation=0）：`canvas_x = 0 - h`, `canvas_y = 240 + v`

### tid=0（p0 真人，坐马桶）

| pose | 帧数 | h 范围 | v 范围 | **canvas x** | **canvas y** | z_avg | z_σ |
|---|---|---|---|---|---|---|---|
| 1 Walking | 7 | -60~80 | 0~60 | -80~60 | 240~300 | 49.4 | 31.5 |
| **3 Sitting** | **637** | -10~110 | -50~50 | **-110~10** | **190~290** | 9.8 | 26.4 |
| 4 Standing | 17 | -70~20 | -30~70 | -20~70 | 210~310 | 54.7 | 47.4 |

**p0 Sitting 637 帧 / 占 96%**：
- canvas 位置 **x ∈ [-110, 10]**, **y ∈ [190, 290]**
- **都在 Wall + FOV 内**（x -110 刚好在 FOV 左边界 -120 内侧，y 全部在 160-310 内）
- 位置 σ 极小（x_σ=13.5, y_σ=17）→ **稳坐马桶**

### tid=1（p1 ghost）

| pose | 帧数 | h 范围 | v 范围 | **canvas x** | **canvas y** | z_avg | z_σ |
|---|---|---|---|---|---|---|---|
| 1 Walking | 6 | 150~190 | -30~10 | -190~-150 | 210~250 | 16.8 | 37.6 |
| **4 Standing** | **668** | **130~190** | **-70~40** | **-190~-130** | **170~280** | **1.0** | **9.6** |

**p1 Standing 668 帧 / 占 99%**：
- canvas 位置 **x ∈ [-190, -130]**, **y ∈ [170, 280]**
- **x=-190 超出 Wall 左边界 -180**（!InRoom）
- **x ∈ [-190, -130] 全部超出 FOV 左边界 -120**（**!InFOV**！）
- z 锁死 0（均值 1.0, σ 9.6）

**判据 0 直接命中**：**p1 所有帧都在雷达理论 FOV 之外** → **硬判 ghost**。

---

## 3. Ghost 成因推测

p1 位置与家具的关系：

```
canvas x 轴:   -190 ─── -180 ─── -150 ─── -130 ─── -120 ─── -100 ─── 0 ─── 120 ─── 190
                │       │        │        │        │        │       │    │
               p1       Wall    bath     bath    FOV边界  Curtain  镜子  Wall  FOV右
              左界     左界     左边     右边
```

**物理解读**：
- **bath**（浴缸）在 `x ∈ [-150, -130]`
- **Curtain**（浴帘）在 `x = -100`
- **mirror**（镜子）在 `x ∈ [0, 120]`
- 吸顶雷达发射信号 → 打到 p0（马桶坐姿，canvas x ~ -50）→ 经浴帘/浴缸/瓷砖反射 → 被错误定位到左侧
- ghost 投影位置（x=-190~-130）在**雷达物理 FOV 之外**
- 同时与浴缸/浴帘位置重叠——典型的"家具反射导致的镜像 ghost"

---

## 4. Engine 判据验证

### 判据 0（★ 硬证据）：**InFOV=false → 直接判 Ghost**

```
p1 的 668 帧 Standing，canvas x ∈ [-190, -130]
FOV 左边界 = -120
→ 全部 x < -120 → 全部 !InFOV
→ engine 构建 grid 时每 cell.InFOV 已烤入，运行时 O(1) 查询即可
→ 每帧 p1 进来立刻判 ghost，不需要任何运动分析
```

### 判据 1：z 锁死

p1 Standing z 均值 1.0、σ 9.6。对比 p0：
- p0 Walking z σ=31.5（动态噪声自然）
- p0 Sitting z σ=26.4（静态但有呼吸等抖动）
- **p1 Standing z σ=9.6**（过度稳定 → ghost）

### 判据 2：pose 贫乏

p1 674 帧里 99% Standing，1% Walking，**无 Sit/Lie**。真人在浴室不会 20 分钟只站不坐。

### 判据 3：**AreaDeny 预判**（人工标注 Interfere）

Layout 里 **mirror** 已被标为 Interfere `(0, 160)-(120, 170)`——engine 构建 grid 时 `SetPrior(rect, AreaDeny, 99, Human)`。

但 **p1 不在 mirror 矩形内**（p1 canvas x<-130，mirror 在 x ∈ [0, 120]）——所以 mirror 标注对本 case 的 ghost 抑制**没直接作用**。真实反射源是**浴帘/浴缸**，用户可以考虑把这两个也标为 Interfere。

---

## 5. 本 Case 对 Engine 设计的 3 条反馈

### 5.1 **InFOV 判据 > z 锁死 > pose 贫乏 > 运动相关**

判据强度从高到低：

| 判据 | 本 case 的命中率 | 延迟 |
|---|---|---|
| **InFOV** | 100%（p1 每帧都出 FOV）| 即时（O(1)）|
| z 锁死 | ~95%（z σ < 10）| 需几帧累积 |
| pose 贫乏 | 明显（99% 仅 Standing）| 需几十帧累积 |
| 运动相关 | 可（跟随 p0）| 最慢 |

**Engine 构建 grid 时必须完整 rasterize InFOV**——这是最快最硬的 ghost 过滤器。

### 5.2 **Wall 画开放多边形的风险**

本 case 的 Wall 画法是 4 段 line，**bottom 只到 x=40**（给 Enter 留洞）。
如果 engine 用"Wall line 首尾相接围多边形"算 InRoom，**下边的开口（x ∈ [40, 120]）就漏出来了**——x 在 40-120、y > 310 的 cell 会错判为 InRoom=false。

**解决方案**（对齐你的判断 1）：
- 把 Enter 矩形 `(40,300)-(120,320)` 和 Wall 并集，补齐闭合边界
- 或者让 Wall 作者画成"完整矩形"然后用 Enter 矩形覆写该门洞 cell 允许穿越

### 5.3 **Z 的物理解读**

数据对比（都是坐姿、站立、走动）：

| 场景 | Sit z | Walk z | Stand z |
|---|---|---|---|
| D523 wall（客厅）| 23.0 | 69.9 | 23.6 |
| **D5F7 ceiling（浴室）**| **9.8** | **49.4** | **54.7** |
| B17F wall（客厅） | — | 73 | — |

**z 的绝对值因安装/场景变化大**（浴室 Sit z ≈ 10，客厅 Sit z ≈ 25），但**相对量有意义**：
- Walking z > Sitting z 普遍成立
- σ 差异稳定（Walking σ > 30，Standing σ < 20）

你的判断 2 强调"z 不是身高，是反射点"——实测支持这个观点。engine 应该**用 z 作为辅助相对信号**（同 cell 内分布的 σ、多 track 间的对比），不作绝对阈值。

---

## 6. Fixture 价值

这 20 分钟数据可直接作为 engine 测试输入：

```go
func TestBathroomGhostFiltering(t *testing.T) {
    // 加载 1370 帧
    frames := loadTSV("testdata/case-bathroom-D5F7-04241304.tsv")
    layout := loadJSON("testdata/layout-D5F7-bathroom.json")

    engine := NewEngine(layout)
    for _, f := range frames {
        engine.ProcessFrame(f)
    }

    // 断言
    cell_p0 := engine.grid.CellAt(-50, 240)  // 马桶区
    assert.Equal(t, AreaSit, cell_p0.Belief[0].Type)

    cell_p1 := engine.grid.CellAt(-170, 220) // ghost 区
    assert.False(t, cell_p1.InFOV)
    // 该 cell 不应该积累 RealDecay
    assert.Less(t, cell_p1.RealDecay, 5)
    assert.Greater(t, cell_p1.GhostDecay, 500)
}
```

---

## 7. 数据位置

| 文件 | 大小 | 说明 |
|---|---|---|
| [case-bathroom-D5F7-04241304.tsv](case-bathroom-D5F7-04241304.tsv) | ~60 KB | 1370 帧原始 track |
| [layout-D5F7-bathroom.json](layout-D5F7-bathroom.json) | ~12 KB | 浴室 layout |

---

## 8. 关联案例

- [case-Radar_D523-04231230.md](case-Radar_D523-04231230.md) — D523 wall 模式，墙面反射 ghost（需要坐标系修正）
- [case-Radar_B197-04241041.md](case-Radar_B197-04241041.md) — B197 坐姿误判 Fall 锁死
- [AI_fall_detect.md](AI_fall_detect.md) — Room Engine 完整设计
