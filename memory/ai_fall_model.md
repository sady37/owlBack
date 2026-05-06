---
name: AI-Fall Model Development
description: 另一个agent正在做的AI跌倒判定模型——用非跌倒态的z分布反推cell的空间语义（沙发/床/椅）
type: project
originSessionId: 346adb76-f827-44a1-924a-5f904ddd6f8d
---
另一个 agent 正在建立 AI-fall 模型，核心思路之一：利用**非跌倒状态下 z 值的正常变化特征**，反向给 roomengine 的 cell 打空间语义标签（沙发、床、椅子等可 lie 区域）。

**Why:** 单纯依赖 pose 序列和瞬时 z 变化判断跌倒误报率高（特别是人在沙发/床上躺下时，姿态和位置特征与地面跌倒相近）。引入"区域是否可 lie"这个空间先验，能大幅降低误报——如果 cell 在长期统计中已被识别为沙发，那么该 cell 的"跌倒"事件应优先判定为在沙发上活动。

**How to apply:**
- 涉及 roomengine 的 cell 语义、空间学习、z 分布统计时，应考虑与这个 AI-fall 模型对接
- 处理跌倒误报分析、姿态分类相关工单时，记住判据应结合 cell 的空间语义，而不仅是姿态序列
- 如果要修改 roomengine 输出给 cardagg/cnagg 的数据结构，需考虑 AI-fall 模型可能需要新增的 cell 语义字段（如 is_liedown_zone、typical_z_range）
- 与 radar_pose_interpretation.md 的判据配合使用
