# Platform Agent 寻址 + alarm_events Producer 契约

记录日期：2026-05-15  
拍板：本次会话讨论（见 conversation context）；权威覆盖 [cardagg_sensor_split.md](cardagg_sensor_split.md) 中关于 producer 字符串 / cardagg-fire 的旧约定。

---

## 1. 背景与目标

owl 全栈是 IPv6-native（[`agent_pipeline_north_star`](../../.claude/projects/-home-wisefido-owl/memory/agent_pipeline_north_star.md) / [`envelope_protocol_evolution`](../../.claude/projects/-home-wisefido-owl/memory/envelope_protocol_evolution.md)），但 **platform agent**（sensor / cardagg / cognitive / action / ai-health / iot / data / device-gateway 各家）此前没有 first-class 网络身份——producer 仅是个 VARCHAR 字符串。后果：

- `alarm_events.producer` 字符串没规范（`cardagg:pending-scanner` / `wisefido-sensor` / `device:qinglan` 混着）→ 不可 JOIN / 不可 prefix-match
- `parent_span` / `trace_id` 没有可信源（sensor envelope `sequence_number` 硬编码 "0"）→ 无法跨服务 trace
- cardagg pending-scanner 在 fire 时把 sensor 的原 producer / parent_span 全丢失 → 数据失真

终态：让 platform agent 跟 device / resident 在同一套 IPv6 地址体系下 first-class 存在，producer 是 INET 类型，trace 链可追。

---

## 2. IPv6 Layout 扩展

### 2.1 现状

| 范围 | 用途 |
|---|---|
| `fd00::/32` | owl B2B 根 |
| `fd00:0:<tenant>::/48` | 业务 tenant；当前实际占用 1-9（7 个 tenant） |

`alarm_events` 既有约束：`ae_addr_in_owl_b2b CHECK (device_addr <<= 'fd00:0000::/32')`

### 2.2 Platform Slot 分配（本次新增）

平台 agent 占 tenant 16-bit 字段的**前后保留段**，避免与业务 tenant 冲突：

| Tenant 段 | 状态 | 用途 |
|---|---|---|
| `fd00:0:0::/48` | 已保留 | all-zero 边界 |
| `fd00:0:1::/48` ~ `fd00:0:f::/48` | 保留（15 slot） | 未来扩展 / 测试 / 迁移；当前未启用 |
| `fd00:0:10::/48` ~ `fd00:0:ffef::/48` | 业务 tenant | 真实客户 |
| `fd00:0:fff0::/48` | 保留 | platform 边界 |
| **`fd00:0:fff1::/48`** | **wisefido-sensor** | sensor 集群 |
| **`fd00:0:fff2::/48`** | **wisefido-cardagg** | cardagg 集群（不当 producer，但需 IP 给 audit/monitoring） |
| **`fd00:0:fff3::/48`** | **wisefido-data** | data 集群 |
| **`fd00:0:fff4::/48`** | **wisefido-iot** | iot 集群 |
| **`fd00:0:fff5::/48`** | **wisefido-qinglan** | qinglan gateway 集群 |
| **`fd00:0:fff6::/48`** | **wisefido-sleepace** | sleepace gateway 集群 |
| **`fd00:0:fff7::/48`** | **wisefido-ai-health** | ai-health 集群 |
| `fd00:0:fff8::/48` | 保留 | wisefido-cognitive（未来） |
| `fd00:0:fff9::/48` | 保留 | wisefido-action（未来） |
| `fd00:0:fffa::/48` ~ `fd00:0:fffe::/48` | 保留 | 5 slot 备用 |
| `fd00:0:ffff::/48` | 已保留 | all-one 边界 |

### 2.3 Agent 内 IP 分配

每个 platform agent /48 内部，多 node 部署用 host 段 1, 2, 3... 区分：

```
wisefido-sensor node-1: fd00:0:fff1::1/128
wisefido-sensor node-2: fd00:0:fff1::2/128
...
```

单节点部署直接用 `::1`。

注：业务 tenant 占用 `fd00:0:1::/48` ~ `fd00:0:9::/48`（既有），与 reserved 段不冲突；reserved 段 IPAM 不分配给新 tenant。

---

## 3. Platform Agent Identity

### 3.1 三要素

每个 platform agent 进程启动时必须确定三样：

- **IPv6**：自身 /128 spatial 身份（如 `fd00:0:fff1::1`）
- **UID**：UUID v5，由 IPv6 派生（确定性）
- **Sequence Counter**：Redis INCR，跨重启单调

### 3.2 UID 派生规则

```
namespace = uuid_v5(NAMESPACE_DNS, "owl.platform")  // 固定常量
         = a3b8c5d2-1234-5678-9abc-def012345678     // 占位，实际跑一次得出
uid       = uuid_v5(namespace, agent_name + ":" + ipv6_canonical)
```

实现见 [`owl-common/spatial/platform_uid.go`](../owl-common/spatial/platform_uid.go)。

**关键约束**：
- `agent_name` 取小写 service 名（`wisefido-sensor` / `wisefido-cardagg` 等）
- `ipv6_canonical` 用 Go `netip.Addr.String()` 标准化（小写、压缩零段）
- UID 派生一次后**写入 `.env`**，启动只读不重算（避免改 IP 时 UID 跟历史 trace 对不上）

### 3.3 .env 字段

每个 platform agent 的 `.env` 必须有：

```bash
# Platform agent identity (owl B2B reserved slot)
SENSOR_IPV6=fd00:0:fff1::1
SENSOR_UID=<uuid_v5 derived once and pinned here>

# UID derivation rule (for manual regen only — DO NOT recompute at startup):
#   namespace = a3b8c5d2-1234-5678-9abc-def012345678
#   uid = uuid_v5(namespace, "wisefido-sensor:" + SENSOR_IPV6)
# Tool: `wisefido-sensor uid-gen --ipv6 fd00:0:fff1::1`
```

每个 agent 用各自的 `<AGENT>_IPV6` / `<AGENT>_UID` 命名。

### 3.4 Sequence Counter

Redis key：`<agent_name>:seq:<ipv6>`，例如 `wisefido-sensor:seq:fd00:0:fff1::1`

- 启动**不重置**
- 每次 envelope publish 前 `INCR` 取值，填入 envelope.sequence_number
- 跨重启单调（重启后 INCR 从上次 +1 继续）

---

## 4. Envelope 字段填充契约

### 4.1 sensor / cognitive / action / ai-health 等"内部 agent"

```
producer        = <self.ipv6>          // INET, e.g. "fd00:0:fff1::1"
subject_entity  = <card_id or other>   // 与上游一致
sequence_number = <INCR value>          // Redis 取值（不再硬编码 0）
device_addr     = <trigger device /128> // 必须是 devices 表里有的真 /128
                                        // sensor 派生时：通过 BedDeviceLookup
                                        // 选定主 trigger 设备（LeftBed→bed 主 radar；
                                        // Fall→roomengine track 关联 radar）
device_type     = "radar" | "sleepad" | "" (依 trigger 设备)
timestamp       = <event 实际发生时刻 ms>  // sensor 派生：p.Trigger.Ts (bed→vacant 那一刻)
                                        // 非 publish 时刻
topic_type      = "alarm"
category        = <eventName, e.g. "LeftBed">
DataValue[0]    = {
    event_name:   <eventName>,
    event_status: "start",  // 不再有 pending_arm / pending_cancel
    alarm_level:  <WARN/...>,
    trace: [<SignalEvidence...>],   // 上游 trace evidence（雷达 track / sleepad alarm seq）
    evidence: {
        engine_fire_ms: <publish 时刻 ms>,
        armed_at_ms:    <ArmedAt>,
        due_at_ms:      <DueAt>,
        // 各 rule 特有元数据，如 fall_score / silent_duration_s
    }
}
```

### 4.2 device-gateway（qinglan / sleepace 等）publish 设备直发 alarm

```
producer        = <device.ipv6 /128>  // 与 device_addr 相同 — 设备直发标志
subject_entity  = <card_id>
sequence_number = <firmware alarm seq, if available; else gateway INCR>
device_addr     = <device.ipv6 /128>
device_type     = "radar" | "sleepad"
timestamp       = <firmware 上报时刻 ms>
topic_type      = "alarm"
category        = <eventName>
DataValue[0]    = <firmware 原始 payload>
```

**判断准则**：消费端看 `producer == device_addr` 即可识别「设备直发」，无需依赖字符串前缀。

### 4.3 cardagg

**不当 producer**。cardagg 永远是消费 envelope → 翻译 → 写库的 thin adapter；它的 IPv6 (`fd00:0:fff2::/48`) 仅用于 audit / monitoring / 进程间通信，**不出现在 `alarm_events.producer` 列**。

---

## 5. alarm_events.producer 列契约

### 5.1 Schema 变更

```sql
-- migration: producer VARCHAR(100) → INET

ALTER TABLE alarm_events DROP CONSTRAINT ae_producer_format;
DELETE FROM alarm_events WHERE producer IS NULL OR producer NOT LIKE 'fd%';  -- 清旧字符串数据
ALTER TABLE alarm_events
    ALTER COLUMN producer TYPE INET USING NULL;
ALTER TABLE alarm_events ADD CONSTRAINT ae_producer_in_owl
    CHECK (producer IS NULL OR producer <<= 'fd00::/32'::INET);
CREATE INDEX idx_ae_producer_gist
    ON alarm_events USING GIST (producer inet_ops)
    WHERE producer IS NOT NULL;
```

### 5.2 取值规则

| Producer 类型 | producer 列 | 与 device_addr 关系 |
|---|---|---|
| Sensor 派生 | `fd00:0:fff1::1`（sensor /128） | producer != device_addr（sensor 是 agent，device_addr 是 trigger 设备） |
| Device 直发 | 与 device_addr 相同 | **producer == device_addr** |
| AI Health / Cognitive / Action（未来） | 各自 /128 | producer != device_addr |

消费端可由 `producer == device_addr` 区分「设备直发」与「agent 派生」。

### 5.3 parent_span / trace_id

- `parent_span` = `<producer_ipv6>.<sequence_number>`  
  e.g. `fd00:0:fff1::1.42`
- `trace_id` = 顶端 envelope 的 `parent_span`（多跳时由首跳定，后续 hop 继承）

---

## 6. cardagg 永不 fire（red line）

补充 [cardagg_sensor_split.md §7](cardagg_sensor_split.md#7) 不要做的事：

> ❌ **cardagg 任何路径里设 producer = `cardagg:*` 或自己的 IPv6 — cardagg 永远是 envelope 翻译器，不是 producer**

具体红线代码位置（迁移目标 = 全部砍除 / 改造）：

- `cardaggProducerPendingScanner` 常量及 `ScanPendingAlarms` 方法 → **删除**
- `cardaggProducerAlarmRouter` 常量及 `PersistAlarmFromTrack` 方法 → **审计 + 改造**（让上游设置真实 producer）
- `AddPendingAlarm` / `RemovePendingAlarm` / `TryAddLeftBedPendingAtTrigger` → **删除**（Hash 不再存在）

Timer-based alarm（LeftBed 30min / Stay 10min / NightAbsence / BedNightAbsence）的 fire 由 **sensor 自己拥有**——sensor.zonealarm.Supervisor 内部 timer 到期直接 `BackChannelAlarmFirer.Fire` → envelope.event_status="start" → cardagg 翻译落库。

**进程重启行为**：sensor 启动时**清空** in-memory pending map **+ 清空 Redis `alarm:pending` Hash**（不做 HA 恢复，跟 cardagg 历史一致——服务都重启了，做 HA 得不偿失）。

---

## 7. Redis `alarm:pending` Hash 退役

整个 Hash + 配套 publish 协议（`event_status=pending_arm` / `pending_cancel`）全部砍：

| 文件 | 改动 |
|---|---|
| `wisefido-sensor/internal/zoneengine/wiring/alarm_firer.go` | 删 `Arm()` / `Cancel()` 方法；`Fire()` 保留并改造 |
| `wisefido-sensor/internal/consumer/alarm_back_channel.go` | 删 `PublishPendingArm` / `PublishPendingCancel` / `EventStatusPendingArm` / `EventStatusPendingCancel` 常量 |
| `wisefido-cardagg/internal/consumer/alarm_handler.go` | 删 A9 分支（处理 `pending_arm` / `pending_cancel`） |
| `wisefido-cardagg/internal/service/alarm_service.go` | 删 `ScanPendingAlarms` / `AddPendingAlarm` / `RemovePendingAlarm` / `TryAddLeftBedPendingAtTrigger` / `cardaggProducer*` 常量 / `redisPendingAlarmKey` 常量 |
| `wisefido-cardagg/main.go` | 删调用 `ScanPendingAlarms` 的 cron / loop |
| `wisefido-sensor/internal/zonealarm/supervisor.go` | 启动时 `redisClient.Del(ctx, "alarm:pending")` 清空（防上次崩溃残留） |

---

## 8. 实施战役（顺序锁定）

### 战役 1 — 数据层
- 1.1 migration SQL（producer→INET，清旧数据）
- 1.2 `owl-common/spatial/platform_uid.go` helper
- 1.3 unit test + apply migration

### 战役 2 — sensor identity + envelope 完整化
- 2.1 BedDeviceLookup（bed_id → 主 trigger 设备 /128）
- 2.2 sensor `.env` + config 接 platform identity
- 2.3 Redis seq INCR + envelope publisher 改造
- 2.4 `alarm_firer.Fire` 用 `p.Trigger.Ts` + `buildTriggerData` 加 trace evidence
- 2.5 端到端实测一条 LeftBed，验证 envelope 完备

### 战役 3 — cardagg 退 fire + pending Hash 退役
- 3.1 删 `ScanPendingAlarms` + `cardaggProducer*` 常量
- 3.2 删 `alarm_firer.Arm/Cancel` + cardagg A9 分支
- 3.3 删 `AddPendingAlarm` / `RemovePendingAlarm` / `TryAddLeftBedPendingAtTrigger`
- 3.4 sensor.Supervisor 启动清空 Redis Hash
- 3.5 端到端实测一条 LeftBed，验证 producer / parent_span / trace_id 正确

### 战役 4 — device-gateway producer 改 INET
- qinglan / sleepace 等 publisher 把 producer 字符串改成 device 自己的 /128

### 战役 5 — 收尾
- audit `cardaggProducerAlarmRouter`（PersistAlarmFromTrack 唯一 caller 是谁，改造或迁移）
- 更新 `AI_fall_detect.md` §17 / `cardagg_sensor_split.md` 引用本文档
- memory 收尾

---

## 9. 与既有契约的关系

- [`cardagg_sensor_split.md`](cardagg_sensor_split.md) §7 "不要做的事" 加入 cardagg 不当 producer 红线
- [`datagram_envelope.md`](datagram_envelope.md) 中关于 producer 字段的描述需更新为 INET 类型
- [`agent_pipeline_north_star`](../../.claude/projects/-home-wisefido-owl/memory/agent_pipeline_north_star.md) 一致：sensor 是 Layer 1 agent，与 cognitive / action 平级
- [`envelope_protocol_evolution`](../../.claude/projects/-home-wisefido-owl/memory/envelope_protocol_evolution.md) 一致：device-as-host stateless 派生扩展到 agent-as-host

---

## 10. 反共识 / 不要做的事

- ❌ 不要给 platform agent 单独存 UUID 列（device_uid 已在表里管 trigger device 那一面；agent 的 UID 是 IPv6 派生的）
- ❌ 不要在 startup 重算 UID（一定从 `.env` 读，否则改 IP 后历史 trace 断链）
- ❌ 不要让 producer 字符串里塞 IPv6（要么 INET 列要么不要碰 producer）
- ❌ 不要让 cardagg 任何路径设 `producer = cardagg.*`（哪怕 audit 用途也不行——audit 用 `evidence` JSONB 字段）
- ❌ 不要做 sensor HA pending 恢复（重启清空，跟 cardagg 历史一致）
