---
name: fall_data_is_artificial_test
description: "radar/monitor_stream 的跌倒数据全是人为测试,倒地时长不真实,不能据此标定覆盖/阈"
metadata: 
  node_type: memory
  type: project
  originSessionId: f86f523b-5bda-4c73-ac49-9b2cf61a4b1c
---

2026-06-09 用户拍正:owl_v2 里 radar Fall / monitor_stream 的跌倒数据**全是人为测试**——测试员故意摔、**躺 ~1min 就站起来**,不会真在地上躺 ≥5min(真长躺受害者那种)。**除了几个 bedside 专门测试**相对可信。

**Why**: 我曾全量测 287 个"有恢复的 Fall"的"摔后倒地时长"分布,得出"52% 自救真摔 / 35% 纯误火"的 mix,并据此说"5min 砍 90% 过乐观、真实上限 ~35%"。**这全错**——倒地时长由**测试行为**决定(测试员演 ~1min),不是真实跌倒分布;那 52% 大部分是测试员演的不是真自救。

**How to apply**:
- **绝不**用现有 radar 测试数据标定 recovery-veto 的**覆盖/价值/阈定量**(15s vs 60s 之争失去数据基础)。
- 跌倒"倒地多久"的真实分布**我们没有数据**——recovery-veto 覆盖/「5min 砍 90%」既不能证实也不能证伪,只能推迟到**真实跌倒数据**(临床/养老院实采)。
- **仍能做**:验**安全性/精度**(真摔不被误否)——测试的"真摔"案(#9/#2)够用;ghost/lost/bed 否决路径不依赖时长仍有效;recovery-veto 的 safety 原则(真长躺>5min→零 recovery 证据→永不被否,track-lost 螺丝+正向 up only)是**机制保证**不依赖数据分布。
- 阈取**保守偏高**(宁可少否多放行=安全)。
- 关联 [[silent_leftbed_fall_recovery_window_gap]](自救跌倒不可静默抹)、[[partial_monitoring_fall_suppression_law]]。
