---
name: Track.ToFieldMap 零值 bed_status 泄漏
description: 雷达心跳/纯雷达流莫名带 bed_status:0 的根因 + 已修一处的位置 + 未来要不要全局清理的判断
type: project
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---
`owl-common/observation/Track.ToFieldMap()` 对 `bed_status` 强制写零值（注释："0 为有效值，始终写入"，因 sleepace 0=在床有语义）。
雷达根本没床概念但复用同一个 Track 结构 → 雷达任何走 ToFieldMap 的路径都莫名带一个 `bed_status:0` 字段。

## Why
Track 是床+雷达双用合体 schema。床场景里 0=在床是合法状态，跳过零值会丢语义；
雷达场景里 0 是占位噪音，下游协议看了会困惑（实测：雷达 monitor/heart 心跳 payload 出现 bed_status:0 引发用户提问）。

## How to apply
- **已局部修复（2026-05-03）**：[wisefido-qinglan/internal/consumer/mqtt_consumer.go:767 publishRadarMonitorHeartbeat](owlBack/wisefido-qinglan/internal/consumer/mqtt_consumer.go#L767) 改为手写 `map[string]any{FieldTrackID, FieldTrackConfidence}`，不再走 ToFieldMap。
- **全局清理判断**：要不要给 ToFieldMap 加 `if HasBed` 守卫看后续是否在其他路径再被绊到。当前没有触发证据，先按"局部清理"原则不动通用 ToFieldMap，避免改动面过大。
- **判断规则**：未来再看到 "雷达数据为啥有 bed_status" 问题，先查那条路径是不是 ToFieldMap，复制本次手写 map 思路打补丁；累计 ≥3 处再考虑改 ToFieldMap 本身。
