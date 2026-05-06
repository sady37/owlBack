---
name: Device Health Offline Detection Bug
description: cardagg/qinglan deviceHealth 未正确检测设备离线，WiFi断开后 offline 仍为 0
type: project
originSessionId: 765631c1-fe23-4b51-87d0-5feb18148e27
---
设备 WiFi 断开后，cardagg SSE 推送的 device_status.offline 仍为 0（在线），导致前端/app toolbar 显示 Online 不更新。

**设备**: 9923003AB17F (device_id: a4ea0c4a-0024-4298-a1b3-6583c1c9d6e4)

**预期行为**: 超过 90 秒无数据或心跳，deviceHealth 应自动标记 offline: 1

**实际表现**: WiFi 物理断开后，device_status.offline 持续为 0

**影响范围**:
- Web radar-trajectory toolbar 的 Online 状态不更新
- App RadarConfigToolbarView 的 Online 状态不更新
- Overview 页面卡片的设备离线图标不显示

**根因已查明 (2026-04-20):**

cardagg Bug（已修，eb1c1f5）：runDeriveLoop 只对有 snapshot 的 card 调 DeriveDeviceOnlineOnly。设备停数据后无 snapshot，b.online 标记了 offline 但永远不写 Redis。Fix: 无 snapshot 的活跃 card 也调 DeriveDeviceOnlineOnly + 补上 PruneStaleDevices 每 90s 调用。

qinglan Bug（待修）：health_check.go:122-127 GetDeviceProperties 失败时只 warn+return，不发 FieldOffline=1。这导致 cardagg 重启后已离线设备的状态无法更新（cardagg 被动聚合，从未见过的设备无法判断离线）。

**How to apply:** cardagg 已修已验证已部署。qinglan health_check 待修（health check 失败时应 publishDeviceAlarm FieldOffline=1）。
