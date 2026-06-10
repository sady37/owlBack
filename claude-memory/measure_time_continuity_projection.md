---
name: measure-time-continuity-projection
description: 2026-05-28 user 提的 measure 拟合改进 idea —— 拟合线上各 dot 投影后按时间连续性判同一段，过滤跨圈巧合共线点；治本对策对应 bed_b17f 尾部 outlier 拉伸问题
metadata: 
  node_type: memory
  type: project
  originSessionId: 3838d8aa-8f98-4e95-adb4-7fbbebba1a19
---

User 2026-05-28 提的 measure furniture 拟合改进 idea，先存档不立即实现：

**机制**：dot 投影到拟合线后，沿主方向相邻投影点之间的**时间差**必须 ≤ 阈值（约 1-2s）才算"同一段连续覆盖"，否则后一个 dot 属于"另一圈"或"另一次走过"。

**用户原话举例**：
> A 5s, B 15s, C 6s 三 dot 投影到拟合线，按位置排序相邻
> AB time gap = 10s > 2，CB = 9s > 2，但 AC = 1s ≤ 2
> 推断：A-C 是同一圈连续走，B 是第二圈巧合落在同一直线上
> 应该让段 extent 只覆盖 [A, C]，不包括 B

**为什么需要这个**：
当前 LS 拟合 + axial snap 在 cleanup 充分时 work，但残留的 tail outlier 会被诚实地拉进 extent（bed_b17f 实例：3 个尾部 dot 把 y 拉到 330 而真实边在 250）。时间连续性能从根本上区分"主圈在 y=250 走的 dot" vs "尾部停太晚在 y=330 飘的 dot" —— 后者跟主圈在 time 上没有连续邻居。

**实现要点**：
- Pt 需要加 time 字段一路传到 `extendSegmentBy` / `mergeColinearSegments`（当前已剥成纯 {x,y}）
- 拟合时按主方向 sort 投影 + 检查相邻时间差
- 形成多段"time-continuous run"，每段独立成 segment 而非合一条
- extent = 各 run 的 extent 并集 OR 各 run 单独 segment

**先验证 prerequisites**：当前 [[radar_measure_flow_design]] 的 Dot 格式已有 `time` 字段；只需 Pt → 派生时保留即可。

**何时讨论**：上一轮 bed_b17f 退化讨论结束后，user 会主动来续。

---

## Idea 2 (2026-05-28 续): Real-time incremental 拟合（user 更倾向这条）

替代 batch 后处理，改成 dot 流入时增量维护"当前生长段"，**段闭合后位置/方向锁死**，后续 dot 无法回去 merge —— 跨圈共线天然解。

机制：
- 维护单一"当前段" state：start dot, last dot, accumulated direction, last dot time
- 新 dot 到：
  - 方向跟当前段差 ≤30° AND 时间差 ≤2s → extend
  - 方向差 >30° → 闭合当前段 + axial snap，开新段
  - 时间差 >2s（暂停/盲区）→ 也闭合
- 全 path 走完 → 段数 → 分类拓扑（rect/L/C/line/blob）

优势 vs Idea 1（batch + time）：
- 状态 bounded（O(1) 不是 O(N²)）
- 决策时机自然：走完每边就闭合，不需回溯
- 实时反馈：走完 2 条边就能预测 L；4 条边预测 rect
- RadarCanvas 本来就流 Dot.time，零数据 schema 改动

抗瞬时噪声设计点：
- 不能单 dot 抖动就闭合 → 滚动 N dot 平均方向再判
- 时间 gap 阈值不能太小（用户拐弯可能慢 1s）
- 段闭合后能否回滚 —— 简单：不能；复杂：维护小 lookback buffer

## 我们对工业界的信息优势（user 2026-05-28 强调）

工业界 SLAM / floor-plan / RANSAC + Manhattan 都基于**无时间戳点云**（LIDAR/TOF 一次扫描），算法被迫只能用纯几何。我们 1Hz dot 流**每点带 UTC 秒**：

| | 工业界数据 | 我们的数据 |
|---|---|---|
| 来源 | LIDAR/TOF 点云、户型扫描 | 1Hz radar dot 流 |
| 时间戳 | 无 / 一致 | 每 dot 带 UTC 秒 |
| 规模 | 上万点 | ~30-100 dot |
| 算法假设 | 纯几何 | 几何 + 时间序 + 走法语义 |

**含义**：我们没必要 follow 工业界 RANSAC/Manhattan 路线 —— 那是数据缺失下的折衷。我们能做工业界做不了的：
- 闭合不可逆段（解跨圈共线）
- 时间 gap 反向推断行为（用户停顿/进盲区/拐弯）
- 实时增量收敛，不必等"全数据收完"

**反向警告**：当前 `measureSketch` SECTION 2 把 Dot 投成 Pt，等于自废武功；Idea 2 是把时间资产捡回来用。

**何时实施**：等 user 主动开题。
