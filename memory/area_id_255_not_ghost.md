---
name: area_id=255 不等于 ghost（firmware 语义）
description: D523 ghost 分析时混淆 area_id=255 与 ghost 的教训：area=255 仅表示"declared area 之外"，不是 ghost 判据
type: project
originSessionId: b1d43ed2-c37b-4fc4-9663-3580c50fb0c9
---
2026-05-06 D523 现场分析：firmware emit 4 个 track（0/1/2/3），track 0 在 area_id=7（declared area），track 1/2/3 都在 area_id=255。

**误判**：把 area=255 当作 ghost 判据，结论"漏标 track 2"。

**真相**（用户 09:51 现场确认）：实际 2 真人——track 0 + track 2 都是真人；track 1+3 是 ghost。track 2 在 area=255 是因为真人走到了 declared area 之外（走廊/墙边等），不是 ghost。

## 正确语义

- `area_id=7`：firmware 自定义编号的 declared monitoring area 之内
- `area_id=255`：declared area 之外（走廊 / 边缘 / 客户没画 area 的地方）
- ghost 判定**不依赖 area_id**——由 roomengine 用位置一致性 / 速度 / 镜面对称 / cell history 综合判断

## roomengine ghost 判定路径

- `birth_grace_skip_far_from_enter`（[track_manager.go:2065](../../../owl/owlBack/wisefido-sensor/internal/roomengine/track_manager.go#L2065)）：怀疑标记，不是 ghost 确认
- grace period 给观察时间：表现像真人 → **不 emit verdict（默认放行）**；表现像 ghost → emit verdict=20

**"无 verdict" = 真人/通过**，不是漏报。

## 类似教训

- [dining_table_walk_intentional.md](dining_table_walk_intentional.md)：餐桌区学不成 AreaDeny 是预期，declared area 之外的 track 不应自动否定
- 任何 firmware 自定义枚举值都要先确认语义，不能靠字段名/数值范围猜
