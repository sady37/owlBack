---
name: 协议层架构北极星 — IPv6 + Cognitive Agent + TDP datagram（演进式）
description: 每件具体工作（bug 修复 / 新功能）顺手向 TDPv2 靠近一步；不大重构，但坚决不偏离方向；命名分层（sensor vs cognitive）必须看清
type: feedback
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---
**北极星视角（用户 2026-05-03 校准）**：

1. **IPv6**：每个 agent first-class 端到端寻址（远期；当前 Redis Streams 是简化版）
2. **TDP datagram**：[`owlRD/TDPv2-1122.md`](../../../owl/owlRD/TDPv2-1122.md) 是数据 schema 基础（producer_id / sequence_number / subject_entity / tags）
3. **Cognitive Agent**：真正的 AI 是 LLM 驱动的认知主体（reasoning / reflection / 综合判断），**当前系统里不存在**，是未来接入位

## 分层（关键命名约束）

```
Layer 0: Firmware              producer="device:UID"
Layer 1: Sensor / Rules        producer="sensor.xxx"  ← wisefido-ai 当前定位
Layer 2: Cognitive (LLM)       producer="cognitive.xxx"  ← 未来接入
Layer 3: Action                producer="action.xxx"
```

**"AI" 标签留给 Layer 2**（cognitive）。当前 `wisefido-ai` / `producer="ai.caregiver"` 都错——它是 sensor/rules 层。远期改名 `wisefido-sensor`。

## 演进策略（不大重构，每件工作顺手靠近）

用户原话："在当前基础上向下演进，一边修改bug, 一边演进。"

每次具体工作（bug 修复 / 新功能 / refactor）**自检 4 问**：
1. 加的字段 / 语义是否与 TDPv2 envelope 一致？或至少不冲突？
2. 加的功能是单 agent 职责，还是跨多 agent？跨多个意味着错了
3. 下游 agent 想信任我的 verdict，能从 envelope 直接看到吗？还是要解 dataValue？
4. 未来拆独立 agent 时，我现在写的代码会成为绊脚石吗？

**每件 PR 至少做一件靠近动作**：
- 即使主线不动 wire schema，注释里加 TODO 指向北极星
- 命名上避免 "ai." 前缀给 sensor 层
- 新加字段优先用 TDPv2 同名同语义（如 `Producer` 代替 `Source`，`SequenceNumber` 代替隐式 `Timestamp` 排序）

## TDPv2 envelope 字段（应当对齐）

| 字段 | 语义 | 当前实现退化 |
|---|---|---|
| producer_id (CodeableConcept) | "谁发的"一等公民 | DeviceUID 混 AI 标识；AI 标识藏 dataValue.source |
| sequence_number (uint64) | producer+seq 全局事件 ID | 仅 Timestamp |
| danger_level (Syslog 0-7) | envelope 一等公民 | 落库时算，路由层不可见 |
| subject_entity | 与 producer 解耦的事件主体 | CardID 凑合 |
| tags ([]Tag) | FHIR Category:code 复合标签 | TopicType+Category 二元 split |

## 红线（最优先）

`ai-health` phase 0/1 **不能复刻 wisefido-ai 单体反模式**——它应当是独立 producer（远期 cognitive 层），不是另一个揉所有逻辑的 single engine。

每次 ai-health 进展时优先确认：是否设计成独立 agent + 独立 producer + 端到端 datagram 思路？

## 完整路径

doc/TODO.md 第 0 项已重写为 **4 个 Phase（A/B/C/D），每个 Phase 内部一次到位（不留兼容包袱）**。开发测试阶段拒绝向后兼容是用户明确指示。

```
Phase A：协议层 v2 大重构（envelope 重构 + device_status 拆出 + 妥协清理同步）
Phase B：命名规范化（wisefido-ai → wisefido-sensor，layer 命名约定）
Phase C：alarm 三件套 + dedup + Tags
Phase D：device 端 LITE 直发 + session token 隐私强化（远期）
```

## 关键安全约束（Phase A 阶段就确立）

`Producer` 字段：必须用 server-side `device_id` (UUID)，**禁止用 `device_uid` (MAC/SN)**。

- 内网 MQTT 1883 plain 路径仍存在
- 边缘 LTE/WiFi 转发节点不一定加密
- attacker 拿到 MAC = 物理设备追踪 + 位置关联 + 住户身份
- UUID 是 server 内部分配，泄露低危

Phase D 远期再用 session token 替换 UUID（DTLS-PSK 颁发 / 周期 rotate / 仅 server 内部知映射）。

## SubjectEntity 字段填写分层（不要破坏原设计）

device-gateway (qinglan/sleepace) **必填** SubjectEntity——已有 device→card cache（cardMappingService），card_id 变化极少，cache 寿命长。这是降低 cardagg 压力 + 利用 card_id 稳定性的原设计。

AI sensor agent (wisefido-ai 等) **不填** SubjectEntity——layer 1 不染卡，subject 由 cardagg 反查。

cardagg 端：trust envelope SubjectEntity；空时反查（device→card 反向索引内存 cache，启动 load + cardChange 增量更新，O(1) lookup）。

**禁止误把"AI 不染卡"原则套在 device-gateway 上**——device-gateway 早就实现了 publisher 端 card cache，应保留。
