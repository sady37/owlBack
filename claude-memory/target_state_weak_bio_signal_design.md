---
name: target-state-weak-bio-signal-design
description: 2026-05-19 终态 — WeakBio 仅 FE 横条展示，既不独立触发 alarm，也不作放大系数影响其他 alarm；新权重 HR=5/RR=5/ApneaH=15
metadata: 
  node_type: memory
  type: project
  originSessionId: 3803fdc4-b793-4d26-9019-59b8c61adfa9
---

## 字段本意（关键概念区分）

**WeakBiometricSignal ≠ device signal_quality**：

| 字段 | 含义 |
|---|---|
| `signal_quality` / `SignalPoor` | 设备硬件信号质量（WiFi/RF/接触） |
| `weak_biometric_signal` | **老人体征本身弱**（HR/RR amplitude 微弱） |

是医疗风险信号，不是设备质量问题。confound：设备接触不良也会让"测到的 vital 信号弱"，所以单源不能下结论，要叠加 HR/RR alarm 多源印证。

## 设计修订：风险描述符（2026-05-19 拍板）

**WeakBio score 不再独立触发 alarm**——它是**长期状态/风险放大系数**，不是事件源。

理由：
- firmware 已经直发 HR/RR/ApneaH/WeakBio raw alarm 进 alarm pipeline（DefaultLevel=Warn）
- 二次造一个 Critical 级 escalation 是 alarm 冗余
- "vital 长期弱"是状况描述符，伪装成事件 → alarm_events 表里多一种永远要 ack 但不对应具体瞬时的"alarm"，反人性
- 类比体温计：38℃ 不会触发"高温警报"，温度是**上下文**让其他判断更精准

## 现有架构（不变）

雷达 firmware → qinglan MQTT → `publishAlarm(WeakBiometricSignal, value=state*20)` → `iot:alarm:stream` → alarm_events 持久化

- alarm 常量 [owl-common/alarm/alarm.go:87](owlBack/owl-common/alarm/alarm.go#L87) WeakBiometricSignal
- DefaultLevel = AlarmLevelWarn / FHIRCategoryClinical / ProcessTypeImmediate
- user 可配 sensitivity + duration_min
- 同组：HeartRateAlert / RespRateAlert / ApneaHypopnea

## TargetState.WeakBiometricSignal 字段 = 摘要分（非事件）

per-spatial 30min lazy 滑窗累加，给 FE / 风险评估器读。**不 publish alarm**。

## 计算公式（修订 2026-05-19）

```
score = max(weakBio_raw_in_window)          # 0..60 (raw = state×20, 0/1/2/3)
      + count(HeartRateAlert)   × 5         # 旧 15 → 新 5
      + count(RespRateAlert)    × 5         # 旧 15 → 新 5
      + count(ApneaHypopnea)    × 15        # 旧 25 → 新 15
final = min(100, score)
```

**权重调整理由**：
- HR/RR 单次异常常见（睡眠 HR 50-60 边缘漂移、翻身导致 RR 噪声）→ 降权 5，让 1-2 条噪声不达 30
- ApneaH 单条临床上有意义但不应单条触红 → 15，让 2 条以上才显示
- WeakBio raw 保持 max（firmware 已分严重等级，不累加）

**临床校准**（AASM AHI 标准 5/15/30 = 轻/中/重）：
- 30min 内 6 次 ApneaH = AHI≈12 = 轻度 → score 90 → 红 ✓
- 30min 内 4 次 ApneaH = AHI≈8 = 临界轻度 → score 60 → 黄 ✓
- 12 次 HR alert (持续真异常) → score 60 → 黄 ✓

**规则细节**：
- **多设备不去重**：同 cardID 多 alarm 源（雷达+sleepad 同床）每条 alarm 都算
- **窗口过期 = drop**：30min 外事件直接丢，不 linear decay（lazy 实现：下次事件到来时扫旧 expire，不维护 timer）
- **空闲老人卡 30min 后自然回 0**（但需 sensor 端补 push，否则 cardagg 卡值——见 follow-up）

## 阈值表（修订 2026-05-19）

| score | UI 横条 | 业务动作 |
|---|---|---|
| 0-29 | **不显示** | 无 |
| 30-59 | **灰** Attention | 无 alarm，FE 横条提示 |
| 60-79 | **黄** Watch | nurse app 卡片关注，FE 横条 |
| **80-100** | **红** Alert | FE 横条；**不影响任何 alarm 决策** |

## 已废止：跨 80 escalation alarm 设计

~~score 跨 80（rising edge）时生成新 alarm event, alarm_type=`WeakBiometricSignal`, level=`AlarmLevelCrit`~~

2026-05-19 砍：
- AggregatorPublisher 接口删
- last80 字段删
- EscalationProducerTag 常量删
- NewTargetStateAggregator 入参去 publisher

## 写到哪个 cardID（不变）

**resident 所在最深卡**：
- 该 unit 有 bed binding → 写 /96 bed card 的 Target
- 否则 → 写 /88 room card 的 Target
- /80 unit 卡的 Target.WeakBiometricSignal **不直写**（UnitPicker 从 winner 子卡投影）

## FE 显示（2026-05-19 新增）

**Card Section3.down.right 横条**（待 Step2 实施）：
- `CardDisplay.VitalTrendLevel int` 字段（0=hide / 1=gray / 2=yellow / 3=red）
- cardagg `card_display_builder` 派生：score → level 阈值映射
- 跟 Section3.down.left 的 Visitor / Bed timing 并排

## 风险放大消费者模式 — 2026-05-19 推翻

之前规划过的 3 个放大消费者（cardagg AlarmRouter / sensor fall_verify / sensor zonealarm Supervisor）**全部不做**。

**Why**：既然 WeakBio 已退化为"风险描述符 + FE 横条"，再让它影响其他 alarm 决策又把它拉回事件源语义，自相矛盾；FE 横条本身就是给 nurse 看的风险上下文，由人决策即可。

后续影响：
- sensor `roomengine.fall_verify` commit `15ba836` (WeakBio≥80 force real) 需回退
- sensor `zonealarm.Supervisor` LeftBed/Stay 缩短 — 不做
- cardagg `AlarmRouter` Warn→Critical — 不做（实施后已回退）

## 不读

- alarm pipeline（已废 escalation，没人再读）
- risk_evaluator (room/bed 维度跟 vital 正交)

## 实施位置

[[target_state_aggregator]] 模块内（[wisefido-sensor/internal/service/target_state_aggregator.go](owlBack/wisefido-sensor/internal/service/target_state_aggregator.go)），跟 LastActiveTs / StandingContinuousMin 累加器同模块。

**当前承担 3 个累加器**：lastActive / standing / weakBio（visitor 已挪到 cardagg VisitorDeriver）。

## 关联

- [[target_state_per_device]] — per-device 维度 / cardagg max-merge
- [[visitor_belongs_to_cardagg]] — visitor 不在此模块
- [[card_display_projector_handoff]] — Section3 布局
