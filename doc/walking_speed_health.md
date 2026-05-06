# 老人室内步速：临床参考 + 未来扩展点

记录于 2026-04-25。来源：用户在 RoomEngine speed-based Move 检测改造中提供的临床数据。
当前代码只用 20 cm/s 作为"是否在动"的二值阈值（cell_learning.go::LearnParams.MoveSpeedCms），
本文档存的是**未来可基于步速做的健康监测扩展**——RoomEngine / 跌倒检测 / 报警的下一阶段 roadmap。

---

## 1. 临床步速参考（75 岁老人室内）

被称为 **"第五大生命体征"** (5th Vital Sign)，是衡量健康衰退和跌倒风险的核心指标。

### 1.1 速度区间

| 速度 (cm/s) | 速度 (m/s) | 临床含义 | 我们的处理 |
|------------|-----------|---------|----------|
| 0-15       | 0-0.15    | 站立 jitter / 雷达噪声 | 不计为 Move |
| **20**     | 0.20      | 蹒跚挪步              | **当前 Move 阈值** |
| 40-60      | 0.40-0.60 | 衰弱老人慢走（pre-frail） | 升 Move ✓，可加 SlowMove tag |
| 60-80      | 0.60-0.80 | 衰弱前兆（Pre-frail）   | 健康预警分界 |
| 80-100     | 0.80-1.00 | 中等风险，可能平衡问题   | |
| 100-120    | 1.00-1.20 | 健康老龄化              | |
| > 150 (1s) | > 1.5     | 雷达噪声 / 多径干扰     | **应过滤但当前没做** |

**临床 Cut-off**：< 60 cm/s 跌倒高危，< 80 cm/s 衰弱前兆，> 100 cm/s 健康。

### 1.2 影响因子

- **夜间起夜**：速度 -20%~30%
- **地毯 vs 硬地板**：地毯更慢
- **认知负荷**（边说话边走）：明显变慢
- **环境复杂度**：转弯/避家具时速度低

### 1.3 步频参考

正常 75 岁室内步频 90-110 步/分钟。

---

## 2. 当前实现（基线）

[track_manager.go::ProcessFrame](../wisefido-sensor/internal/roomengine/track_manager.go) 中用 Kalman 速度做 **二值判定**：

```go
if core == CorePoseStand || core == CorePoseUnknown {
    speed := math.Hypot(float64(vx), float64(vy))
    if speed > float64(tm.moveSpeedCms) {  // 默认 20 cm/s
        core = CorePoseMove
    }
}
```

只解决"雷达 pose 把慢走老人误报 Standing → ActiveType[Move] 永远不涨"的问题。
**未利用任何更精细的速度分级**。

---

## 3. 未来扩展点（按优先级）

### 3.1 步速健康指数（Walking Speed Health Index）

**目标**：把步速作为长期健康趋势指标，从"已经出事再报"转向"提前几周预警衰退"。

**实现思路**：
- TrackManager 维护 per-track 的 `EMASpeed`（指数滑动平均，长档半衰期 7d）
- 每天定时计算 daily aggregate：`avg_walk_speed_cms`、`p50/p90 速度`、`Move 帧总时长`
- 落 PG（新表 `daily_health_metrics` 或 `iot_timeseries` 加 topic_type=`health_daily`）
- 趋势检测：30 天滑窗里 avg 从 X→Y 下降 > 30% → 触发 `FrailtyTrendAlarm`

**触发样例**：30 天均值从 90 cm/s → 60 cm/s → 推送家属"运动能力衰退"。

**依赖**：需要 cardagg / wisefido-sensor 加一个 daily aggregator goroutine。

### 3.2 慢走 + 高风险区域 → fall 预警权重提升

**目标**：用步速给 fall 检测加先验。当老人扶墙慢行进入卫生间，跌倒可能性显著高于正常情况。

**实现思路**（修改 fall_check.go / track_manager.go::checkSilentFall）：

```go
// 当前：消失 + 检查 cell 是否高风险区
// 改进：再叠加最近 30s 速度统计
recent := ts.HistoryWindow(30 * time.Second)
avgSpeed := computeAvgSpeed(recent)

riskMultiplier := 1.0
if cell.Belief[0].Type == AreaShower || cell.Belief[0].Type == AreaToilet {
    if avgSpeed < 40 {  // 低于衰弱阈值
        riskMultiplier = 2.0  // 跌倒判定阈值放宽 2 倍
    }
}
```

**关键**：不是直接报 fall，而是**降低 fall 判定的硬阈值** —— 慢走入厕的老人，z 下降幅度只要达到正常阈值的一半就能触发。

### 3.3 Move 阈值二级化

**目标**：区分"慢走"和"正常走"，长期统计 baseline 步速。

**Cell 字段扩展**：
```go
type Cell struct {
    ...
    ActiveType    [4]uint16  // [Move, Stand, Sit, Lie] 已有
    SlowMoveCount uint16     // 新增：< 60 cm/s 的 Move 帧数
    FastMoveCount uint16     // 新增：> 60 cm/s 的 Move 帧数
}
```

**用途**：
- 区域级分析："这条走廊老人平均走速从 80 → 50，可能装了助步器或骨折康复"
- per-resident baseline：累积 30 天得到该老人在该房间各 cell 的"正常步速"，偏离 > 2σ 触发关注
- snapshot 持久化里要加这两个字段（schema_v 升到 2，需要兼容读旧版）

### 3.4 噪声过滤（> 150 cm/s）

**目标**：当前没做，临床数据里明确指出 > 1.5m/s 室内基本不可能（除非跑），多半是雷达多径或目标跳变。

**实现思路**（在 ProcessFrame 速度计算后）：
```go
speed := math.Hypot(float64(vx), float64(vy))
if speed > 150 {
    // 不更新 Kalman / 不喂 ActiveType / 标 ts.Score 减分
    continue
}
```

注意：跑步的健康老人也可能瞬时 > 150 cm/s，要看持续时间。建议用"连续 N 帧 > 150"才丢弃，单帧只降权。

---

## 4. 取舍提醒

- 这些都是**长期分析**功能，需要积累数据才有意义。**新装机房间至少 30 天内别启用 3.1 / 3.3 的告警**，等 baseline 稳定。
- 步速跟身高、体重、用辅具相关，**不能跨人比较绝对值**，必须 per-resident baseline。
- 临床阈值（60/80/100）是**人群均值参考**，单个人偏差大；用 trend（自身相对变化）比绝对阈值更稳。
- fall_check 加先验（3.2）需要先验证不会显著提升误报率——建议先 shadow mode 跑 2 周，统计加先验后的 P/R 再上线。
