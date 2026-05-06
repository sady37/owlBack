---
name: Walking Speed Health Roadmap
description: 老人室内步速临床数据 + 4 个未来扩展点（健康指数/fall先验/二级阈值/噪声过滤），完整文档在 owlBack/doc/walking_speed_health.md
type: project
originSessionId: c27ea410-6e58-493c-bbdc-8e863b87c54e
---
老人步速作为"第五大生命体征"的临床数据 + 基于步速的 4 个未来扩展点，完整记在
[owlBack/doc/walking_speed_health.md](../../../owl/owlBack/doc/walking_speed_health.md)。

**当前状态**：只用 20 cm/s 做 Move/Stand 二值兜底（cell_learning.go::LearnParams.MoveSpeedCms），
未利用更精细的速度分级。

**关键临床阈值**（< 60 cm/s 跌倒高危，60-80 衰弱前兆，80-120 健康，> 150 噪声）。

**4 个未来扩展点**（按优先级）：
1. **步速健康指数**：per-track EMA 速度 + 30 天滑窗趋势 → FrailtyTrendAlarm 告警家属
2. **慢走 + 高风险区 → fall 先验**：进卫生间速度 < 40 cm/s 时，跌倒判定阈值放宽 2 倍
3. **Move 阈值二级化**：Cell 加 SlowMove/FastMove 计数，做 per-resident baseline
4. **> 150 cm/s 噪声过滤**：雷达多径，连续 N 帧才丢

**Why**：开发 RoomEngine speed-based Move 检测（解决 Walk 学不出来的 bug）时，用户提供了完整临床数据。
当时只用了二值阈值，但临床分级有更深用法值得记下来。

**How to apply**：未来做 fall 检测优化、健康趋势报警、frailty 监测时翻这份文档。
不要试图一次性全做——3.1/3.3 需要 30 天数据积累，3.2 需要 shadow mode 验证 P/R。
