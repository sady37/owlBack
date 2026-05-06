---
name: 未绑卡 device 用 device_id 充当 card_id（临时妥协，待 device_status 独立后清理）
description: device-gateway / cardagg 统一规则；本质是 device_status 嵌在 card:state 下的 workaround；TODO 第 2 项做完后可清理
type: project
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---
**这是 workaround 不是设计目标**。根因：[doc/TODO 第 2 项](../../../owl/owlBack/doc/TODO.md) `device_status` 当前嵌在 `card:state:{cardID}.DeviceStatus[deviceID]` 子字段 → 没 card 就没地方挂 device 状态 → 用 device_id 充 card_id 自建虚拟卡 key。

第 2 项做完后（`device:status:{deviceID}` 独立 Hash），本规则可清理：
- envelope card_id 未绑卡时可保持空，不再 device_id 占位
- `StreamHeadCardID` / `EffectiveCardID` 标 deprecated 过渡后下线

## owl-common 当前规则

[`card.StreamHeadCardID(cardID, deviceID)`](../../../owl/owlBack/owl-common/card/device_baseline.go#L60-L66)[`card.StreamHeadCardID(cardID, deviceID)`](../../../owl/owlBack/owl-common/card/device_baseline.go#L60-L66)

```go
// 已绑卡用 card_id，未绑卡用 device_id（UUID），与 cardagg 未绑卡占位一致。
func StreamHeadCardID(cardID, deviceID string) string {
    if s := strings.TrimSpace(cardID); s != "" {
        return s
    }
    return strings.TrimSpace(deviceID)
}
```

## 业务场景：什么时候 device 未绑卡

按 [Card Creation Rules](../../../owl/owlRD/docs/20_Card_Creation_Rules_Final.md) 规则 2，cards 在以下情况**不会创建**：
- 新装设备尚未配置（unit/bed 未绑定）
- `monitoring_enabled = FALSE`（监护未激活）
- 公共空间但未起 UnitCard（场景 C 条件 `COUNT(未绑床设备) = 0` 时不创建）
- 卡片刚被 unit_id+device_id 去重删除而新卡未生成的瞬间

这些情况设备已上线但发出的 alarm/event 不应丢失（巡检 / 设备健康 / debugging 都需要）。

## 实施一致性检查

| 模块 | 未绑卡时 envelope card_id | 状态 |
|---|---|---|
| owl-common StreamHeadCardID | device_id (UUID) | ✓ 标准定义 |
| wisefido-qinglan resolveDeviceIdentity | "" (空，未用 StreamHeadCardID) | △ 兜底由 cardagg 处理 |
| wisefido-sleepace | TODO 待审 | ? |
| wisefido-ai publishAIMessage | "" (空，按 sensor 层不染卡设计) | ✓ |
| **cardagg IotPreparedHandler** | **0 卡 → 用 device_id 兜底**（不再 drop）| ✓ 2026-05-03 修 |

## cardagg 的 IsUUID 跳过查库守卫

[device_meta.go IsUUID 注释](../../../owl/owlBack/wisefido-cardagg/internal/service/device_meta.go#L378)：
> "cards.card_id 为 UUID，未绑卡时会用 deviceKey 充 card_id，不应查库"

实际：`GetOrLoad(cardID)` 找不到对应 cards 行 → 返回 nil → 跳过 `EnsureCardStatePrepared`。这是**自然的兜底**，不需要预先判别 card 还是 device。

## alarm_events 落库兼容

[InsertAlarmAndUpdateCard:124-134](../../../owl/owlBack/owl-common/card/alarm_db.go#L124-L134) 当 cardID 为空时反查 device_id；当 cardID 是 device_id UUID 时（cards 表查无），需要 metadata.card_id 兜底或 fallback。需要进一步审视 `findCardIDByDevice` 失败时的兜底是否覆盖此 case。

## 红线

不要再回退到 "0 卡 → drop"。未绑卡设备消息丢失会导致：
- 新设备装机调试时看不到信号
- 公共区设备配置过渡期数据丢失
- 巡检/health-check 设备状态丢失
