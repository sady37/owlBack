---
name: cardagg-vs-sensor
description: cardagg=即时+已确定+可信源 薄 adapter / sensor=不可信源(radar)+单物理实体内推导+同实体多源融合+时间窗 判定层；跨物理实体合并归 cardagg；权威文档在 owlBack/doc/cardagg_sensor_split.md
metadata: 
  node_type: memory
  type: project
  originSessionId: 93425a5a-d704-47ac-b637-b5fc082a6fff
---

2026-05-15 用户拍板（之前会话也讨论过但当时未落 doc，这次补落）；2026-05-24 修正"跨源融合"措辞为"单实体内"（见 [[visitor_belongs_to_cardagg]] 边界澄清）。

## 划分原则（一句话）

- **cardagg = 即时 + 已确定 + 可信源** 的薄 adapter（envelope→DB/Redis 翻译）；**跨物理实体的合并**也归 cardagg（visitor 跨 room、TargetState 跨 device max 合并）
- **sensor = 不可信源 (radar) + 单物理实体内推导 + 同实体内多源融合 (如 sleepad+radar 同床) + 时间窗** 的判定层；**永不跨物理实体合并**（一旦合并细节丢失）

## 可信源 vs 不可信源

- ✅ Sleepad firmware native alarm（设备端已 gate，时序由 firmware 完成）
- ✅ 设备健康类（Offline/SignalPoor/AngleException/SensorDetached/DeviceFailure，物理事实）
- ✅ Sensor 派生（producer="wisefido-sensor"，已 gate）
- ❌ **Radar firmware 整类**（含 Fall/SittingOnGround/Stay/InBed/LeftBed —— 全部不可信，必须 sensor verifier）

## 关键含义

- Fall 必须经 sensor — radar firmware 容易误报，cardagg 看到 radar producer 的 Fall 应 ignore
- cardagg 不做任何算法（不推导/不时间窗/不融合）
- sensor 派生回流后 cardagg 信任已 gate（理想态退化为 `if level != ""`）
- **同实体内**多源融合（如 sleepad LeftBed + radar 仍在**同床**区 → silent_fall）属 sensor 独有
- **跨实体合并**（如 同 card 多 radar 的 TargetState max 合并、跨 room 的 visitor 累加）归 cardagg；sensor 只 per-/128-device 发出原料

## Gateway 端分流（2026-05-15 拍板）

不靠 consumer 端 producer filter，**device-gateway 自己拍板发哪个 stream**：

- `iot:alarm:stream` = 可信、可直接落库 → 只 cardagg 订
- `iot:event:stream` = 待判定、需推断 → 只 sensor 订

qinglan 改动：Radar Fall/SittingOnGround 从 alarm stream 改发 event stream；设备健康类（Offline/SignalPoor/AngleException/SensorDetached）保持 alarm stream（硬件物理事实，可信）。

sleepace 完全不动，全留 alarm stream。

## 完整清单 + 数据链 + cleanup backlog

→ [doc/cardagg_sensor_split.md](../../../owl/owlBack/doc/cardagg_sensor_split.md)（权威文档）

## 与冲突的旧文档

- [alarm_back_channel.go](../../../owl/owlBack/wisefido-sensor/internal/consumer/alarm_back_channel.go) 注释"不绕过 cardagg 的 enablement gate" — 按新分工应改为"sensor 自 gate，cardagg trust"
- [AI_fall_detect.md](../../../owl/owlBack/doc/AI_fall_detect.md) §17 "cardagg 现有 case alarm.Fall handler 自动落 alarm_events" — 应限定为 producer="wisefido-sensor"

## 当前阻塞 (2026-05-15)

- cardagg.alarm_enablement.loadDevice IPv6 cast bug 阻塞 Fall 测试
- 终态下 cardagg 不再 gate radar Fall（radar 由 sensor 处理），但当前残留状态下还需先修这个 bug 才能让现有 firmware Fall 测试落库
- 否则跑 sensor verifier 接管的 PR 之前，整个 Fall 测试链都断

## Reference

- [agent_pipeline_north_star](agent_pipeline_north_star.md) — Layer 0/1/2 分层
- [zonealarm_done](zonealarm_done.md) — sensor 派生 alarm 4 件套已落
- [zoneengine_adapters_done](zoneengine_adapters_done.md) — sensor 写 card_status presence 三投影已落
- [v2_cutover_lessons](v2_cutover_lessons.md) — v1→v2 短路 vs 重写教训
