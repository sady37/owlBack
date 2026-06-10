---
name: Room Engine 设计与实现
description: wisefido-ai/internal/roomengine — 房间级空间认知模块，ghost track过滤+跌倒/异常检测，已完成基础框架，待完善时空因果关联
type: project
originSessionId: 27da4026-2c5f-47a5-b987-5c5965c65243
---
Room Engine 已在 wisefido-ai/internal/roomengine/ 下实现基础框架（6个文件，编译通过）。

**核心设计思想**（用户主导）：
1. "物理引擎"：track不能凭空出现，传感器需互证，没有"因"的"果"是假信号
2. 连续性：人的运动是平滑的，空间跳跃=fake；Kalman残差是统一的打分指标
3. 房间视为图片：10cm×10cm网格，每个格子带概率分布，自动从观测中学习（贝叶斯）
4. 5因子快筛（出生）+ Kalman持续跟踪（存活）
5. 时间风险因子：5-6点×2.0，夜间×1.5；在场因子：独居×1.5
6. 格子纠正pose：马桶旁stand→sit，金属干扰区pose降权以运动学为准
7. 静止超时：空地站立≤8min，卫生间≤15min，床不限

**已实现模块**：Cell, RoomGrid, KalmanFilter2D, TrackState, TrackManager, Engine
**集成方式**：Engine作为独立goroutine在AlarmService.Start中启动，订阅iot:monitor:stream

**待继续**：
- 用户正在构思"坐标出现与事件的时空距离"量化方法 — 因果链的数学表达
- 从DB加载layout_config解析为RoomConfig
- 与现有evaluator的alarm输出打通

**Why:** 当前radar ghost track导致大量假警报（跌倒误报），cardagg无过滤能力
**How to apply:** 后续在roomengine上扩展时空关联评分，不改cardagg
