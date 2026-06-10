---
name: 业务模型 — 设备与卡 1:1 绑定
description: 每个 device 只属一个 resident 或一张 public 卡；公共区域用 public 卡，不分发到住户卡；不要为"一设备多卡"做 fan-out 设计
type: project
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---
**用户 2026-05-03 校准**：当前业务模型上每个 device 严格 1:1 绑一张卡。

## 三种 cards 类型

| 卡类型 | 绑定语义 | 设备归属 |
|---|---|---|
| 住户卡（resident） | 1 卡 = 1 老人 | 老人房间内的雷达/床垫，1:1 绑该卡 |
| **public 卡** | **1 卡 = 1 公共区域**（客厅/食堂/走廊）| 该区域内设备 1:1 绑 public 卡，不属于任何具体住户 |
| 临时/换户 | 短暂存在 | 同上 1:1 |

## 重要业务推论

- 公共区域设备触发的事件（如客厅雷达 fall）**归 public 卡**，不分发到附近住户的卡
- 即使夫妻共住共床——业务上也是单卡（couple 卡或主卡），不双卡
- **永远不应该 fan-out 一条 alarm 到多张卡**

## 工程含义（IotPreparedHandler 反查策略）

[wisefido-cardagg/internal/consumer/iot_prepared.go](../../../owl/owlBack/wisefido-cardagg/internal/consumer/iot_prepared.go) `LookupCardsByDevice` 反查路径：

```
zero cards   → drop（device 未绑定任何卡，异常）
single card  → 填 raw 走原路径（业务常态）
multi cards  → warn + 取首张兜底（数据异常，cards.devices 重复绑定）
```

**不做 fan-out**。多卡命中说明 cards 表数据异常需运维排查，不是设计场景。

## 何时这条不再适用

如果未来业务模型变更：
- 出现"多 resident 共享同一设备"且需要每位独立 alarm 路由
- 出现"事件按 subject_entity 拆分到不同卡"的需求

那时再讨论 fan-out 是否合理（可能要先实现 subject_entity 解耦——见 doc/TODO 第 0 项 layer 1/2）。
