---
name: target-state-per-device
description: 2026-05-18 拍板 v2：TargetState 按 radar device (/128) 寻址；sensor per-device 发布；cardagg 合并多 device 写 single Target；StandingContinuousMin 从 RoomState 挪到 TargetState；v3 引入 logicID 真融合
metadata: 
  node_type: memory
  type: project
  originSessionId: 3803fdc4-b793-4d26-9019-59b8c61adfa9
---

## 核心架构（v2 拍板 2026-05-18）

| 层 | TargetState 寻址 | 责任 |
|---|---|---|
| Sensor | **per radar device (/128)** | 每个 radar 维护自己的 TargetState；不跨 radar 合并 |
| Cardagg | per card | 接收多 device target → 合并 → 写单一 card:state.target Hash |
| FE | per card | dumb render 合并后的单 Target（无感知 device 维度）|

v3 未来引入 logicID 实现 cross-radar 真融合；v2 暂以 `max(across devices in card)` 近似。

## 字段位置（schema 改动）

| 字段 | 旧位置 | 新位置（v2 拍板）| 维度 |
|---|---|---|---|
| `LastActiveTs` | TargetState | TargetState | per device |
| `StandingContinuousMin` | **RoomState** ⭐ | **TargetState** ⭐ | per device |
| `WeakBiometricSignal` | TargetState | TargetState | per device |
| `TrackID` / `LogicID` | TargetState | TargetState | per device |

⭐ `StandingContinuousMin` 从 RoomState 迁到 TargetState 是 P3 实施项。

## Sensor 输出格式

- per radar device 独立发一条消息（不 batch）
- key = device_addr (/128 INET CIDR text)
- category = `target.state`
- 同 unit 多 radar → 多条独立消息

理由：更新时间不同 → per-device 独立发更新流畅；cardagg 自己 batch。

## Cardagg 合并规则

**Active 示例**（同卡 R1 + R2 两 radar）：
```
T=0   R1 active → cardLastActive = max(T=0, 0)            = T=0,  active now
T=3m  R2 active → cardLastActive = max(T=0+续, T=3m)      = T=3m, active now
T=5m  R2 停      → cardLastActive = max(T=5m+R1续, T=3m) = T=5m, active now
T=10m R1 停      → cardLastActive = max(T=10m, T=3m)     = T=10m, active 0s ago
T=12m            → cardLastActive = T=10m                  → active 2min ago
```

**写入逻辑**：
- card.target.last_active_ts        = max(deviceTarget.LastActiveTs across devices in card)
- card.target.standing_continuous_min = max(deviceTarget.StandingContinuousMin)
- card.target.weak_biometric_signal = max(deviceTarget.WeakBiometricSignal)
- card.target.track_id / logic_id   = 待定（v3 logicID 上线再定；v2 可能存最近活跃 radar 的）

## Visitor 例外（仍归 cardagg）

详 [[visitor_belongs_to_cardagg]]：visitor 三字段不走"sensor per device 发 → cardagg 合并"路径，cardagg `VisitorDeriver` 60s tick **直接写** /88 room card target。

## 不影响的部分（v2 保持现状）

- RoomState 仍 per /88 room 发（不变）
- BedState 仍 per /96 bed 发（不变）
- Bathroom 在 publish 层已合并 RoomState (kind=Bathroom)；state machine 层仍 ZoneTypeBathroom（不动）

## 实施 checkpoint

- 2026-05-18 P2-fix-2: sensor TargetStateAggregator 退化 3 累加器（lastActive / standing / weakBio）；visitor 已挪
- **P3 (待做)**：sensor `spatialAccumulator` → `deviceAccumulator` (key=/128)，per-device 累加；
  card_types.go StandingContinuousMin 从 RoomState 挪到 TargetState
- **P4 (待做)**：cardagg SensorStateProjector 加 device map + max 合并；
  cardagg 新建 VisitorDeriver
- v3：引入 logicID cross-radar 真融合（替代 max 近似）

## 关联

- [[visitor_belongs_to_cardagg]] — visitor 跨 room 合并归 cardagg
- [[target_state_weak_bio_signal_design]] — weakBio 累加细节（含 80+ escalation）
- [[card_display_projector_handoff]] — Task 2A/2B 拆分
