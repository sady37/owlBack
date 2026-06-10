---
name: radar_geometry_point_invariant
description: IoT 设备 geometry 必须是 point；旧 boundary-as-rectangle 编辑曾把雷达写成 rectangle 损坏数据，已加载自愈
metadata: 
  node_type: memory
  type: project
  originSessionId: 9565d152-c490-46f8-8f1f-6f693d65cf91
---

2026-06-01 排查 kitchen B17F「雷达图标看不到 + move 只动边界 + 模式调不动」根因：该雷达的 `geometry.type` 被存成了 `rectangle`（4 顶点小方块，丢了 point 的 x/y/z）。B197 同样损坏。

**铁律**：IoT 设备（Radar/Sleepad/Sensor）的 `geometry.type` 必须是 `point`（设备在空间是一个点）；boundary 属于 `device.iot.radar.boundary`，**绝不能写进 geometry**。

**损坏链**：旧版 boundary 编辑把雷达 geometry 当 rectangle 拖动 → 覆盖了 point。当前新代码已正确（边界拖动走 [RadarCanvas.vue](../../../owl/owlFront/src/components/Radar/RadarCanvas.vue) 的 `bdry-1/bdry-2` 分支只写 `iot.radar.boundary`，early-return 不碰 geometry），**但修不回**已损坏的：rectangle 雷达的边界拖动会落进 rectangle-resize 分支，越拖越坏并存回库（"永续陷阱"）。

**修复**（已落地）：
- 数据修复：B17F→`point{0,10,160}`wall / B197→`point{250,5,200}`corn（从 v1 备份 `backup1.0/dbv1-layouts-20260513/rooms/` 取原值）
- 加载自愈：[objects.ts](../../../owl/owlFront/src/stores/radar/objects.ts) `normalizeDeviceGeometry` 接进 loadCanvas + loadFromLayoutConfig，非 point 的设备几何用质心折回 point，打破永续陷阱

**同会话相关**：window 级 mouseup 兜底修了「拖动控制点出画布松手卡死」（rectangle/Wall 角点在画布边缘最易触发，表现为"要双击才释放"）。Wall resize 手柄按几何分：rectangle=对角 2 角 / line=2 端点 / polygon(trace)=每顶点(共轴邻居跟动)；locked=true 对象完全不可拖（Toolbar 🔒 复选框解锁）。
