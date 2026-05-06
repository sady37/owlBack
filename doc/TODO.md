# Backlog — 单独开会话解决

记录已经达成共识但未实施的改造。每条含前因/后果/设计/验收，可独立开会话推进。

---

## 0. 协议层架构演进路线图（北极星，所有后续工作向其靠近）

> **不是改造 task，是目标范式**。每条具体改动（包括下方 1-5）评估时都问："离这个范式更近还是更远？" 远了就退回重想。

### 北极星：IPv6 + Cognitive Agent + TDP datagram

> 项目最早设计的视角。三个独立维度合在一起：
>
> - **IPv6 视角**：每个 agent 是 first-class 网络实体（端到端寻址，无 NAT-like 中间代理）。当前 Redis Streams 中心 broker 是工程简化版，远期向端到端 datagram（CoAP/TDP）演进。
> - **TDP datagram**：[`owlRD/TDPv2-1122.md`](../../owlRD/TDPv2-1122.md) 是数据 schema 基础——producer_id / sequence_number / subject_entity / tags 等字段。**TDP 只管数据结构，不决定 agent 行为**。
> - **Cognitive Agent**：真正的 AI agent 是 LLM 驱动的认知主体（reasoning / reflection / 跨领域综合判断）——**当前系统里还不存在**，是远期接入位。

### 分层（把 "AI" 标签让给 cognitive 层）

```
Layer 0: Firmware (sensor)
         producer="device:E598A2ACD523"

Layer 1: Sensor Agent / Rules Engine          ← wisefido-ai 当前定位（不是 AI）
         producer="sensor.caregiver01"
         产出：track_verdict / 规则触发的 Fall / Ghost 标注

Layer 2: Cognitive Agent (LLM-driven)         ← 当前系统不存在，未来 Claude/GPT 类接入
         producer="cognitive.caregiver01"
         订阅 Layer 0 + Layer 1
         产出：综合判断（"近 30min Fall + vital 异常 + 行为偏离 → 高优先级"）

Layer 3: Action (notification / call / dispatch)
         producer="action.notifier"
```

每层 agent 在 IPv6 视角下是 first-class 寻址；TDP envelope 的 `producer_id` 字段承载分层信息（system=layer, code=instance）。

### 命名误导（最大的概念污染，当前急需修正）

`wisefido-ai` 模块名 **claim 了它不是的东西**：
- 它解析 firmware frame / 跑 Kalman / 维护 cell history / 规则触发 alarm
- 它是 **firmware 的延伸 + 规则引擎**，**没有 reasoning，没有跨领域综合判断**
- 真正的 cognitive AI（LLM 驱动）将来会接入**消费它的 verdict**，不是替代它

`producer="ai.caregiver01"` 同样错——把 sensor 规则引擎标榜成 AI。未来 cognitive Caregiver 接入时命名冲突不可避免。

**正确做法**：
- 当前 `wisefido-ai` 应改名为 `wisefido-sensor` 或 `wisefido-rules`（远期重构）
- producer 命名约定 `sensor.caregiver01` / `cognitive.caregiver01`，layer 1 不占 "ai." 前缀
- "AI" 标签留给 layer 2（cognitive）

### 核心约束

- Agent 单一职责（safety / clinical / behavioral / ...）
- Agent 互信（下游信上游 verdict，不重复同质判断）
- Agent 横向扩展（同类多实例靠 producer_id 区分）
- Agent 互相解耦（订阅/发布通过 stream，无直接调用）
- **Layer 不混**：sensor 不冒认 cognitive；cognitive 不重做 sensor 的工作

### 设计依据：TDPv2-1122 envelope

[`/home/wisefido/owl/owlRD/TDPv2-1122.md`](../../owlRD/TDPv2-1122.md) 这套 protobuf schema 是项目最初的设计，**就是为 agent pipeline 量身做的**：

| TDPv2 字段 | 语义 | 当前实现 | 偏离 |
|---|---|---|---|
| `producer_id` (CodeableConcept) | "谁发的" 一等公民 | DeviceUID + DeviceID 混用，AI 标识藏 dataValue.source | ✗ |
| `sequence_number` (uint64) | producer+seq 全局事件 ID | 仅 Timestamp（重复/错序风险）| ✗ |
| `danger_level` (Syslog 0-7) | envelope 一等公民 | InsertAlarm 时算 | ✗ |
| `subject_entity` | 事件主体（住户）— 与 producer 解耦 | CardID 凑合（混淆 subject 与设备绑定）| △ |
| `semantic_location` | 房间语义位置 | RoomID（不在 envelope）| △ |
| `tags` ([]Tag) | FHIR Category:code 列表，复合事件多标签 | TopicType + Category 二元 split | △ |
| `event_mode` (LITE/FULL) | 高频/低频分流 | 无 | ✗ |

### 当前偏离的代价

- **wisefido-ai 单体**：safety + ghost + cell-learning + verdict 揉在一个 engine。改任何一块要重启整个 engine。
- **AI 派生 alarm 装作 "radar device" 发出**：cardagg 收到无法区分"是雷达直发"还是"AI Caregiver 派生"，分流靠嗅探 dataValue.source（藏起来的标识）。
- **多 AI agent 无法接入**：ai-health phase 0 即将复刻同样反模式——如果不及时拉回到"独立 Doctor producer"，技术债 ×2。
- **Caregiver 和 Doctor 不能互信**：未来 Doctor 看到 vital 异常无法引用"30min 前发生过 Fall"加权——producer/subject 解耦的能力丢失。

### 实施策略：开发测试阶段一刀切（不留兼容包袱）

> 当前在开发测试阶段，**不需要向后兼容**。一味兼容会让妥协代码永久留存，永远到不了北极星。
> 每个 Phase 内部"一次到位"，旧字段直接删除/重命名，所有相关服务同时重启。
>
> 用户原话：「不向后兼容，当前仍在开发测试阶段，如果一味向后兼容，永远到不了终点」。

### 四个 Phase（依赖序，每个 Phase 内部一次到位）

```
Phase A（协议层 v2 大重构 — server 内部 envelope）
   ├──> Phase B（命名规范化）
   ├──> Phase C（alarm 三件套 + dedup）
   └──> Phase D（device 端 LITE 直发 + 隐私强化 — 远期）
```

#### Phase A: 协议层 v2 — envelope 重构 + device_status 拆出 + 妥协清理

**envelope schema 重构**：

```go
type IoTStreamMessage struct {
    Producer        string  // 强制必填："device:<UUID>" / "sensor.caregiver01"
                            // ← 用 device_id (server UUID)，不用 device_uid (MAC/SN)
                            //   避免 plain wire 时 leak 硬件序列号导致物理追踪
    SubjectEntity   string  // 可空（未绑卡时直接空，不再用 device_id 占位）
    SemanticLocation string // room_id / unit_id
    SequenceNumber  uint64  // producer 内单调

    Timestamp       int64
    TopicType       string
    Category        string  // Phase C 时改为 Tags []Tag
    DataValue       []interface{}

    TenantID        string
    DeviceID        string  // server UUID（业务路由 / alarm_events 关联）
    DeviceUID       string  // 物理 MAC/SN——保留但 wire 上**逐步淡出**（详见 Phase D）
}
// 删除：CardID 字段（subject 已迁移到 SubjectEntity）
```

**Producer 字段安全约束**：
- Phase A 内：`Producer = "device:" + device_id (UUID)`，**禁止**用 `device_uid` (MAC/SN)
- 理由：开发期内网仍有 MQTT 1883 plain 路径 + 边缘转发节点不一定加密；attacker 拿到流量可以基于 MAC 关联到具体设备/位置/住户
- Phase D（远期）引入 session token 进一步替换 server UUID

**SubjectEntity 字段填写责任分层**（保留"降低 cardagg 压力 + card_id 变化极少"的原设计）：

| Publisher 角色 | SubjectEntity 填写责任 | 为什么 |
|---|---|---|
| **device-gateway**（qinglan / sleepace） | **必填**（已绑卡=card_id，未绑卡=device_id 占位）| 已有 device→card cache（`cardMappingService`），card_id 极少变 → cache 寿命长，cardagg 直接 trust envelope 零反查 |
| **AI sensor agent**（wisefido-ai / 未来 cognitive） | **不填**（envelope 留空）| Layer 1/2 解耦原则：sensor 不染卡概念；subject 由 cardagg 反查 |
| **未来 cognitive agent**（LLM 类） | 不填，复用 sensor 的 verdict 时保留原 SubjectEntity | 同上 |

**cardagg 反查优化**（避免每条 SQL）：

`DeviceMetaCache` 加内存反向索引：

```go
deviceUIDIndex map[string][]string  // device_uid → []cardID
deviceIDIndex  map[string][]string  // device_id  → []cardID
```

- 启动时：扫 `cards.devices` JSONB 一次构建
- 增量：监听 `config.card` cardChange 事件 → 重建对应 device 条目
- 反查 `LookupCardsByDevice` 改成内存 lookup（O(1)）
- 与 device-gateway `cardMappingService` cache 机制对称（同 cardChange 事件源）

**压力分布**：

| 路径 | 流量量级 | SubjectEntity 来源 | 反查成本 |
|---|---|---|---|
| device-gateway → cardagg | 绝大多数（firmware 原始数据）| publisher cache（已有）| 0（trust envelope）|
| AI sensor → cardagg | 少量（派生 alarm / verdict）| cardagg 反向索引 cache | O(1) memory |

cardagg 不会被反查压垮——AI 派生流量远小于 firmware 原始流量。

**存储重构**：

```
device:status:{deviceID}    ← 新增独立 Hash（device-level 真相）
  ├─ online / signal_poor / angle_abnormal / sensor_detached / last_seen_ms

card:state:{cardID}          ← 直接删除 DeviceStatus 子字段（不双写）
```

**全部清理同步做**（不留 deprecated）：

- 删 `StreamHeadCardID` / `EffectiveCardID` 函数
- 删 `IotPreparedHandler` device_id 占位逻辑
- 删 `mapRoomsToCards` / `SetRoomCard` 残留
- 删散布的 `IsUUID(cardID)` 守卫
- 删 `card:state.DeviceStatus` 所有读写点

**Publisher 改造分层**（按上面"SubjectEntity 填写责任分层"）：
- `qinglan / sleepace`（device-gateway）：填 `Producer="device:UUID"` + `SubjectEntity=cardMapping(deviceUID)`（已绑卡用 card_id；未绑卡用 device_id 占位）
- `wisefido-ai`（sensor agent）：填 `Producer="sensor.caregiver01"` + `SubjectEntity=""`（不染卡）

旧 `CardID` 字段全删除。

**部署约束**：所有 publisher + cardagg 同时重启。Redis 中老 `card:state.DeviceStatus` 数据弃用（不 migrate，开发测试阶段重新累积即可）。

**工程量**：~780 行（schema 50 + publisher 200 + consumer 150 + storage 150 + cleanup 80 + tests 150）

**收益**：消除所有累积妥协（device_id 占位 / IsUUID 守卫 / 单体 wisefido-ai 染卡概念 / device_status 多卡冗余），envelope 一次性对齐 TDPv2 三维度。

#### Phase B: 命名层修正

- `wisefido-ai` 模块 → `wisefido-sensor`（service / systemd unit / 二进制全改）
- producer 命名规范：`sensor.caregiver01` / `cognitive.doctor01` / `device:UID`
- "AI" 标签留给 layer 2 cognitive

独立 PR，Phase A 之后做。

#### Phase C: alarm 三件套 + dedup + Tags

- `PersistAlarmAndPublish` 改用 `(Producer, SubjectEntity, Tags)` 三件套
- `AlarmDef.DedupWhileActive` 字段（device-class 一律 true）
- alarm_events schema 加 `producer` + `sequence_number` 列
- `Tags []Tag` 替代 `Category` 字段（FHIR Category:code 复合）

独立 PR，Phase A 之后做。吸收原 TODO 第 1 项 dedup（自然合并）。

#### Phase D: device 端 LITE 直发 + 隐私强化（远期）

> 终极目标：firmware 直接发 TDP datagram 进入流，绕过 qinglan/sleepace envelope 包装。
> 这是 IPv6 + first-class agent 视角的实现层兑现。

**LITE/FULL 二级模式**（TDPv2 已设计，参 `LiteEventDatagram` / `EventDatagram`）：

- LITE：device 高频小数据（心率/呼吸/track），ESP32 protobuf nano ~50 bytes envelope，仅 LiteEventHeader（sequence_number + event_time + producer_id + danger_level + tags）
- FULL：低频危险事件（fall/alarm），含 ExtendEventHeader（subject_entity / semantic_location / sleep_period）由 server-side sensor agent 富化补完

**隐私强化（session token 化）**：

- `producer_id` 在 device 端用 firmware-provisioned **session token**，不用 device_id UUID 也不用 MAC
- `system: "xai:session", code: "<rotating session token>"`：device 配对时 server 颁发，周期 rotate
- session ↔ device_id mapping 仅在 server 内部，attacker 抓包无法关联硬件 / 位置 / 住户
- DTLS-PSK / mTLS 是底层传输强制（Phase D 前提）

**清理**：
- envelope `DeviceUID` 字段彻底从 wire 移除（Phase A 已淡出，D 完成移除）
- `DeviceID` 也可考虑替换为 session-scoped ID（但仍保留 server-side 真名做业务路由）

**触发条件**：
- ESP32/firmware 端 protobuf nano + DTLS 协议栈成熟
- session token 颁发/rotate 机制设计落地（PSK 升级路径）
- 新硬件批次上线（旧设备只能走当前 qinglan 中转）

**工程量**：大；跨 firmware + qinglan + cardagg + KMS（session token 颁发）。**当前不安排时间**，记录作为远期目标。

### 关键工作约束（每件 PR 自检）

进入这个 backlog 之前，每个新工作都应当问自己：

- 我加的字段 / 字段语义是否与 TDPv2 envelope 一致？
- 我加的功能是属于一个 agent 的职责，还是跨多个 agent？跨多个意味着错了
- 下游 agent 想信任我的 verdict，能不能从 envelope 直接看到？还是要解 dataValue？
- 如果未来要拆出独立 agent，我现在写的代码会成为绊脚石吗？
- **是否在为"向后兼容"付代码税？开发测试阶段拒绝兼容包袱，一刀切**

### Reference

- TDPv2 protobuf schema：[`owlRD/TDPv2-1122.md`](../../owlRD/TDPv2-1122.md)
- 当前 wire schema：[`docs/AI_Iot_stream.md`](AI_Iot_stream.md)
- Card Creation Rules：[`owlRD/docs/20_Card_Creation_Rules_Final.md`](../../owlRD/docs/20_Card_Creation_Rules_Final.md)
- AI publish 现状：[`wisefido-sensor/internal/roomengine/engine.go publishAIMessage`](../wisefido-sensor/internal/roomengine/engine.go)

---

## 1. cardagg 端 onset dedup（PersistAlarmAndPublish 翻转判断）

> **已合并到第 0 项 Phase C**。本节作为详细 spec 保留供参考；具体实施在 Phase C 内。

### 前因

2026-05-03 修 AngleException 走 alarm 流时发现 [alarm_service.go::PersistAlarmAndPublish](../wisefido-cardagg/internal/service/alarm_service.go) → `InsertAlarmAndUpdateCard` 完全不去重：每次调用都直接 INSERT 一条 alarm_events、`updateAlarmCount(+1)`、可能刷 popAlarm + push 通知。

qinglan health-check 每 10min 周期 publish 设备类 onset/recover（[health_check.go:167-169](../wisefido-qinglan/internal/subscriber/health_check.go#L167-L169) 无脑全发），如果设备角度持续异常，每 10min 会写一条新 AngleException 记录。

恢复端已有翻转 dedup 等价物——`HandleRecoveryWithTypes` 对无 active alarm 是 noop（幂等）。**onset 端缺对称设计**。

### 后果（不做的代价）

- alarm_events 表周期性灌行（每 10min/设备/类型 一条）
- cards.unhandled_alarm_X 计数失真
- pop_alarm 频繁刷新，前端"最新告警"显示混乱
- push 通知重复轰炸

### 设计

[owl-common/alarm/alarm.go](../owl-common/alarm/alarm.go) `AlarmDef` 加字段：

```go
type AlarmDef struct {
    ...
    DedupWhileActive bool   // 同 device + 同 type 已有 active 时跳过 INSERT
}
```

注册侧（device-class 全部置 true）：
- `Offline / SignalPoor / AngleException / SensorDetached`

事件型（Fall / SuspectedFall / NightAbsence / Stay / InBed / LeftBed）保持 false（每次独立记录）。

[alarm_service.go::PersistAlarmAndPublish](../wisefido-cardagg/internal/service/alarm_service.go) 在 `InsertAlarmAndUpdateCard` 前加守卫：

```go
if def := alarm.LookupAlarm(eventName); def != nil && def.DedupWhileActive {
    var exists bool
    _ = s.db.QueryRowContext(ctx,
        `SELECT EXISTS(SELECT 1 FROM alarm_events
                       WHERE device_id=$1 AND event_type=$2 AND alarm_status='active')`,
        msg.DeviceID, eventName,
    ).Scan(&exists)
    if exists {
        s.logger.Debug("alarm onset deduped, active exists",
            zap.String("device_id", msg.DeviceID), zap.String("event", eventName))
        return nil
    }
}
```

`alarm_events` 已有索引 `(device_id, alarm_status)` 或类似（按需补索引）。

### 验收

- D523 持续 angle_abnormal=1 时，alarm_events 仅 1 条 active AngleException（不再每 10min 灌）
- AngleAbnormal=0 一次后再变 1 → alarm_events 多一条新 active（翻转后允许新报）
- Fall 类不受影响（每次单独入库验证）
- pop_alarm 不被周期 onset 反复刷新

---

## 2. device_status 独立维护（拆离 card:state）

> **已合并到第 0 项 Phase A**。本节作为详细 spec 保留供参考；具体实施在 Phase A 内（与 envelope 重构 + 妥协清理同步做）。

### 前因

当前 device 在线/SignalPoor/AngleAbnormal/SensorDetached 等状态嵌在 [state_service.go::DeriveAndWriteState](../wisefido-cardagg/internal/service/state_service.go) 写入的 `card:state:{cardID}` Hash 的 `DeviceStatus map[device_id]*card.DeviceStatus` 子字段里。

设备共享多卡场景（一个雷达服务两个 resident，或 sleepad 跨房间）时，每张卡的 `card:state:*` 都要存一份这个设备的状态副本。

### 后果（不做的代价）

- **数据冗余**：同设备 N 卡 → N 份状态拷贝
- **不一致风险**：N 个写入路径，任一路径漏更新 → 不同 card 看到设备状态不同
- **语义错位**：angle/signal/online 是设备本身属性，与"哪张卡"无关，硬塞进 card:state 模型是为方便聚合查询而非真实归属
- **dedup 路径绕弯**：检查"是否有 active alarm"应直接读 device-level 真相（`alarm_events.device_id`），而不是经 card:state 转一道
- **未绑卡 device 状态无处可挂** ← 本项未做带出的连锁工程妥协：

  设备未绑卡（新装/未配置/monitoring 关闭/卡片刚去重重建瞬间）时也会上线、发 health-check、有信号差/角度异常等状态需要追踪。但当前 device_status 嵌在 card:state 下面 → 没 card 就没地方挂这些状态。

  **临时妥协**（owl-common `StreamHeadCardID` + cardagg `IotPreparedHandler`）：未绑卡时用 device_id 充当 card_id，相当于自建一个"虚拟卡 key"挂 device_status。

  这是 workaround，不是设计目标。本项做完后这条妥协可清理：
  - `device:status:{deviceID}` 直接写，不需要假 card_id
  - `IotPreparedHandler` 未绑卡分支可以简化（envelope card_id 可以保持空，不再需要 device_id 占位）
  - `StreamHeadCardID` / `EffectiveCardID` 可标 deprecated（保留过渡期，最终下线）

### 设计

#### Redis schema 改造

```
device:status:{deviceID}   ← 新增 Hash（device-level 单点真相）
  ├─ online: 0|1
  ├─ signal_poor: 0|1
  ├─ angle_abnormal: 0|1
  ├─ sensor_detached: 0|1
  ├─ last_seen_ms: bigint
  └─ updated_at_ms: bigint

card:state:{cardID}        ← 不再嵌 DeviceStatus
  └─ 其它字段照旧
```

#### 读路径改造

- 前端/聚合查询：`card:state:*` 拿 card.devices 列表 → `MGET device:status:*` 拼装
- AlarmService dedup 查询：直接 `SELECT FROM alarm_events WHERE device_id=$1 AND alarm_status='active'`（已是 device-level）

#### 写路径改造

[state_service.go::SetDeviceOnline](../wisefido-cardagg/internal/service/state_service.go)、[BuildDeviceStatus](../wisefido-cardagg/internal/service/monitor_buffer.go)、[event_handler.go UpdateStatus](../wisefido-cardagg/internal/consumer/event_handler.go) 等所有写入点：从写 `card:state:{cardID}.DeviceStatus[deviceID]` 改为写 `device:status:{deviceID}` 直接。

#### 迁移策略

1. **双写期**：新写入路径同时更新 `device:status:*`（新）和 `card:state.DeviceStatus`（旧），读路径仍读 card:state（兼容前端）
2. **切换读**：前端/聚合查询切到 device:status:* + 拼装；老 reader 保留兼容
3. **下线旧字段**：确认无读者后清理 card:state.DeviceStatus 写入

### 验收

- 一个雷达绑两张 card 时，雷达离线一次 → 两张 card 都看到该设备 offline=1
- `device:status:{deviceID}` 是该设备状态的唯一真相，无重复存储
- alarm_events dedup 查询直接走 device_id 索引，无需 card 中介
- 老聚合接口前端零感知（数据形态不变，来源换了）

---

## 共同约束

- 不影响现有 alarm 流量（修改集中在 cardagg 内部）
- 兼容 sleepace 已有的非对称 dedup（[wisefido-sleepace/internal/subscriber/health_check.go:170-192](../wisefido-sleepace/internal/subscriber/health_check.go#L170-L192) onset 翻转发，recover 永远发 KeepAlive）
- OfflineRecover 周期发**保留**（用于 monitor_buffer KeepAlive 软心跳，见 [alarm_handler.go:271-277](../wisefido-cardagg/internal/consumer/alarm_handler.go#L271-L277)）

## 相关 case

- 2026-05-03 D523 fall 测试 [doc/cases/d523-fall-13-52/raw.tsv](cases/d523-fall-13-52/raw.tsv) — 暴露 AngleException 周期触发但实测倾角 -63.89 全天不变，正是要被 dedup 的典型场景
- D523 配置 install_model=2（墙角 45-60°），实测稳定 ~64°，可能需要顺手放宽阈值到 -65 ≤ Y ≤ -45（独立讨论）

---

## 3. lost-fall "走着突然摔进盲区" 漏报场景

### 前因

2026-05-03 frame-level frozen 判据由 byte-equal 改为 box（StillBoxCm=30，最近 30s 位移 box ≤ 30cm 视为 still）。
PR-C 流式 cancel pending 用 `FrozenStartMs > 0` 作守卫（保留 D523/CD2B 现有保护）。

但 box 判据**仍无法识别**一种边界场景：人持续走（位移 > 30cm），走到房间深处突然摔进盲区（视角被家具遮挡）。
- 失锁前 30s box > StillBoxCm（一直在走）→ FrozenRunStart = 0
- firmware 发 number_people=0（误判屋空）

→ pending 入池后 number_people=0 流式 cancel（守卫 FrozenStartMs=0 不通过保留条件）→ **漏报**

### 后果

elder care 中真摔倒在盲区的 case 不能被 lost-fall 路径捕获。但属于罕见场景：
- 多数摔倒前有几秒减速 / 停顿（box 内会标记 still）
- 完全"持续走"突然摔多见于年轻人滑倒，elder 步态较慢

### 设计候选

- **位置语义**：消失点 NearestEntryDist > 100cm（远离门）+ 在已知盲区附近 cell → 不取消
- **firmware confidence 衰减信号**：失锁前几帧 track_confidence 走低 → 区分"主动离场"与"被动失锁"
- **多帧速度趋势**：失锁前几帧 Kalman 速度从 normal 突然降到 0 → 可能摔倒
- **房间已知盲区标定**：layout 中标记 ObstaclePeak 区，临近这些区的 lost 不被 number_people=0 取消

### 验收

- 模拟"持续走 → 突然失锁 + np=0"序列：现状 cancel 不报；改造后保留 pending 报警
- D523/CD2B 现有 case 行为不退化


---

## 4. LeftBed 事件双触发 + inBedSince 第二条丢上下文

### 前因

2026-05-03 Phase A 部署后实测 BM87224700978 上下床：
- 21:47:15 InBed → 21:47:39 LeftBed（间隔 24s）
- cardagg event_handler 21:47:39 连续输出 2 条 `LeftBed.pending.check`，相隔 2ms
  - 第 1 条：`inBedMs=24007`（正确）
  - 第 2 条：`inBedMs=1777870059890`（= `m.Timestamp` 自身，inBedSince 已被第 1 条 `delete`）

### 怀疑

sleepace 把同一 LeftBed 同时 publish 到 `iot:event:stream` + `iot:alarm:stream` 两条流，
cardagg `event_handler` + `alarm_handler` 各处理一次。
第 1 条 LeftBed 在 event_handler 里把 `inBedSince[deviceID]` 取出 + delete（用完清除）；
第 2 条进来时 inBedSince 命中 0，导致 `inBedMs = m.Timestamp - 0 = m.Timestamp` 异常。

### 后果

- `stableInBed = inBedTs > 0 && inBedMs >= leftBedMinInBedMs`：第 2 条永远 stableInBed=false（短期影响很小，因为 inBedTs=0 已经不通过守卫）
- 但 LeftBed.pending.check 日志双打 + 后续 PersistAlarmAndPublish/RemovePendingAlarm 走两次，需确认幂等
- 与 Phase A envelope 重构无关（Phase A 前已存在）

### 设计候选

- **dedup 维度**：alarm_handler / event_handler 共用一个 dedupKey（device_id + event_name + ts 桶 100ms），任一路径处理过即跳过另一条
- **职责拆分**：event_handler 不再处理 LeftBed alarm 路径（让 alarm_handler 独占），只处理 BedState 写回 → 单源
- **publisher 端去重**：sleepace 不要同时往 event + alarm 两条流发 LeftBed（最干净，但要确认其它消费者不依赖旧路径）

### 验收

- 同一 LeftBed 触发只产生一条 LeftBed.pending.check
- inBedMs 始终 = LeftBed.ts - InBed.ts，不会走到 m.Timestamp 自身
- 现有 InBed_cancel / pending.add / pending.removed 行为不退化

### Reference

- 实测日志: log/wisefido-cardagg.log 21:47:39（device_uid=BM87224700978）

---

## 5. alarm_state.pop_alarm ack 后没有 fallback 选下一条 active（wontfix）

> **2026-05-03 决定不做**：生产期 caregiver 不会让 alarm 堆积超过几分钟，wisefido-data 重启 / 前端 SSE 重连时 pushStatusSnapshot 会用当前 alarm_state（含 active_X 计数）重新渲染，等价于把次高顶上来。"次高 fallback 缺失"的实际触发频率近 0。
>
> 多 Admin 协作场景的灾难（最新 verify 后老 active 隐形）只发生在数分钟级 alarm 堆积，elder care 实际部署不允许此情况。下方 spec 保留作为参考。

### 前因（保留供参考）

2026-05-03 lost-fall 实测：cid 9d460580 一卡多 active CRITICAL Fall 累积。
- 22:16:06 新 Fall 入库，pop_alarm = "CRITICAL.Fall" + event_id = 20e8197c-... + active_crit = 5
- 22:22:00 Admin 在前端按 "Verified Alarm" → wisefido-data 调 HandleAlarmEvent → publish alarmProcess:ack → cardagg 处理后写回 card:state.alarm_state：
  - active_crit 5 → 4（正确）
  - **pop_alarm = ""，event_id = ""**（错误）
- 22:23:02 / 22:23:04 / 22:23:05 / 22:23:07 又连 verify 4 条老 Fall，每次 pop 都被清空

### 后果

- 前端弹窗渲染靠 `pop_alarm`：ack 一条后弹窗消失，**剩余 4 条 active CRITICAL 不重新弹**
- "看似无报警"假象——active_crit > 0 但页面静默
- 多人协作场景灾难（团队里只要一个 Admin verify 了最新的，其他人就再也看不到剩余 active）

### 设计

**ACK / Resolve / FalseAlarm 任何一类操作走 alarm_process_handler 时：**
1. 计算"剩余 active 集合"（DB query：cards 表 alarm 字段 + alarm_events.alarm_status='active'）
2. 按优先级选下一条 pop：
   - 严重度优先：EMERGENCY > CRITICAL > ALERT > ERROR > WARNING
   - 同级别取 triggered_at 最新的
3. 写 alarm_state.pop_alarm + event_id（无剩余则真空）

### 验收

- 多 Fall 累积场景：连续 verify 5 条，前 4 次 pop 一直有值（每次顶上去），最后一次才空
- 前端弹窗持续可见直到所有 active CRITICAL 处理完
- alarm_state.active_crit 与 pop_alarm 同步：count=0 → pop=""；count>0 → pop 必非空

### Reference

- 实测：log/wisefido-cardagg.log + log/wisefido-data.log 22:16:06 ~ 22:23:07
- Redis 取证：
  ```
  XRANGE card:status:stream - +  # 查 cid 9d460580 的 alarm_state 演变
  HGET card:state:9d460580-... alarm_state
  ```
- 引发场景：跨班次/跨多 Admin 协作，老报警累积未及时处理，新报警 ack 不可见

---

## 6. cardagg AlarmProcessHandler 缺 Info 日志

### 前因

实测 22:22:00 ack 命令到达 cardagg 后，wisefido-cardagg.log 完全没有任何痕迹。
排查 alarm 状态变化时只能从 Redis stream + DB 反推。

### 后果

- 运维难复原 ack 时间线
- 无法区分 ack 来自哪个 user / device_id / event_id
- 与 alarm_handler.go (`offline_recover.done` 频发 Info)、event_handler.go (`pending.removed` Info) 不一致

### 设计

`AlarmProcessHandler.Handle` 末尾加一条 Info：

```go
h.logger.Info("alarm_process.applied",
    zap.String("cid", d.CardID),
    zap.String("event_id", d.EventID),
    zap.String("alarm_type", d.AlarmType),
    zap.String("alarm_level", d.AlarmLevel),
    zap.String("process_type", d.ProcessType),
    zap.String("tenant_id", d.TenantID),
)
```

### 验收

- ack/verify/false_alarm 命令到达 cardagg 后立即出现一条 Info
- 字段足够支撑"谁 ack 了哪条"反查
- 不引入热路径性能影响（alarm 流量极低）

### Reference

- 文件: wisefido-cardagg/internal/consumer/alarm_process_handler.go:81
- 配套: 与 #5 一起改，便于实测验收 pop_alarm fallback 流程

---

## 7. ~~owlCare app 升级后，把 offline 字段拆回 raw ds.Offline~~（已撤销）

> **2026-05-04 撤销**：聚合方案过早 over-engineer。最终决定字段保持单一职责（offline = raw ds.Offline，sensor_detached 独立），各端按场景渲染。owlCare 升级时直接在 card 上加 SensorDetached 图标，不再回头改 wire format。下方 spec 仅作历史记录。

### 前因（已不适用）

2026-05-04 把 device:status hash 4 标志位（offline/signal_poor/angle_abnormal/sensor_detached）通过 `wisefido-data` API 边界完整下发，让 owlFront 能区分 SensorDetached（红）/ WeakSignal（黄）/ Tilted（黄）/ Offline（红）/ Online（绿）。

但为兼容**未升级的 owlCare iOS（Swift）**——它的 `CardDeviceStatus` 只声明 4 字段（device_id/device_type/offline/updated_at），没有 sensor_detached——把 `offline` 字段在 API 边界**聚合**为 `(Offline OR SensorDetached)`。这样老 Swift 看到 `offline=1` 即正确显示传感器脱落设备为红灯，不漏报。

### 不变量（当前必须守住）

- `enrichDeviceStatus` / `FillDeviceOnlineStatusFromCardagg` 写出的 JSON `offline` 字段语义 = `(ds.Offline OR ds.SensorDetached)`，**不能**退化成 raw `ds.Offline`
- 否则老 Swift 看到 `offline=0` 时 SensorDetached 设备误显示为在线 → REGRESSION（设备物理不可用却挂绿灯）

### 触发条件（什么时候做这次清理）

owlCare iOS Swift 升级版本完整上线，所有终端用户都升级到新版（CardDeviceStatus 含 sensor_detached/signal_poor/angle_abnormal/last_seen_ms 字段）。

可通过 App Store 强制更新策略 + 后端版本检测 + 一段灰度观察期确认。

### 拆解（升级后做）

1. `wisefido-data/internal/service/card_realtime_service.go` `enrichDeviceStatus`：
   - `entry["offline"] = ds.Offline`（不再聚合 SensorDetached）
2. `wisefido-data/internal/service/device_service.go` `FillDeviceOnlineStatusFromCardagg`：
   - 判 online 改回 `ds.Offline == 0`（不再 AND `SensorDetached == 0`）
3. `owlFront/src/api/monitors/model/monitorModel.ts` DeviceStatus type 注释更新：
   - 删除「offline 是 (Offline OR SensorDetached) 的聚合值」段落
   - 改为「offline 仅反映网络/心跳；sensor_detached 独立；前端 OR 合成显示语义"不可用"」
4. `owlFront/src/views/monitoring/detail/Detail.vue` `deviceStatusDetail`：
   - 先判 sensor_detached 再判 offline 的优先级保持不变
   - 但触发分支语义更纯：sensor-detached 不再依赖 offline=1 兜底
5. 任何后续新建的客户端不再受兼容包袱

### 验收

- 设备 sensor 脱落但网络正常：API 返回 `{offline:0, sensor_detached:1}`；新 owlCare/owlFront 显示 "Sensor Detached" 红
- 设备纯网络断开：API 返回 `{offline:1}`；客户端显示 "Offline" 红
- 同时 sensor 脱落 + 网络断开：API 返回 `{offline:1, sensor_detached:1}`；客户端按优先级显示 "Sensor Detached"
- 老 Swift 已不存在生产环境（升级率 100%），不再出现"sensor 脱落 → offline=0 → 显示在线"误判

### Reference

- 当前实现：[wisefido-data/internal/service/card_realtime_service.go enrichDeviceStatus](../wisefido-data/internal/service/card_realtime_service.go)
- 当前实现：[wisefido-data/internal/service/device_service.go FillDeviceStatusFromCardagg](../wisefido-data/internal/service/device_service.go)
- owlCare 待升级：[owlCare/OwlCare/Network/DataModels.swift CardDeviceStatus](../../../owlCare/OwlCare/Network/DataModels.swift) (line 434-446)
