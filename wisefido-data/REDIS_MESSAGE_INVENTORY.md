# wisefido-data Redis 消息发送清单

## 概览
本清单记录 wisefido-data 服务发出的所有 Redis Stream 消息。共计 4 个消息发送位置，涉及 2 个主要 stream。

---

## 详细清单

### 1. 设备配置更新通知
**文件路径**: [internal/notifier/config_notifier.go](internal/notifier/config_notifier.go)  
**函数名**: `NotifyAlarmDeviceUpdated`  
**Stream 名称**: `config:device_status:stream`  
**消息类型**: `com.wisefido.alarm.device.updated` (CloudEvents 格式)  
**代码行号**: [61](internal/notifier/config_notifier.go#L61)

**说明**: 当设备报警配置更新时，发送设备信息到配置流。消息包含设备 ID、设备代码、设备类型和可选的位置信息。

**调用链**: 设备配置更新 → NotifyAlarmDeviceUpdated → rediscommon.PublishToStream

---

### 2. 卡片创建通知
**文件路径**: [internal/notifier/config_notifier.go](internal/notifier/config_notifier.go)  
**函数名**: `NotifyCardCreated` (通过 `publishConfigChange`)  
**Stream 名称**: `config:device_status:stream`  
**消息类型**: `config.card.created` (CloudEvents 格式)  
**代码行号**: [107](internal/notifier/config_notifier.go#L107)

**说明**: 当新卡片创建时发送消息。消息包含卡片 ID、单位 ID、分支 ID 和关联的设备列表。

**调用链**: CardSyncService.emitCardChange → NotifyCardCreated → publishConfigChange → rediscommon.PublishToStream

**上层调用文件**: [internal/service/card_sync_service.go](internal/service/card_sync_service.go#L199)

---

### 3. 卡片更新通知
**文件路径**: [internal/notifier/config_notifier.go](internal/notifier/config_notifier.go)  
**函数名**: `NotifyCardUpdated` (通过 `publishConfigChange`)  
**Stream 名称**: `config:device_status:stream`  
**消息类型**: `config.card.updated` (CloudEvents 格式)  
**代码行号**: [107](internal/notifier/config_notifier.go#L107)

**说明**: 当卡片更新时发送消息。消息包含卡片 ID、单位 ID、分支 ID 和关联的设备列表。

**调用链**: CardSyncService.emitCardChange → NotifyCardUpdated → publishConfigChange → rediscommon.PublishToStream

**上层调用文件**: [internal/service/card_sync_service.go](internal/service/card_sync_service.go#L203)

---

### 4. 卡片删除通知
**文件路径**: [internal/notifier/config_notifier.go](internal/notifier/config_notifier.go)  
**函数名**: `NotifyCardDeleted` (通过 `publishConfigChange`)  
**Stream 名称**: `config:device_status:stream`  
**消息类型**: `config.card.deleted` (CloudEvents 格式)  
**代码行号**: [107](internal/notifier/config_notifier.go#L107)

**说明**: 当卡片删除时发送消息。消息包含卡片 ID、单位 ID 和分支 ID。

**调用链**: CardSyncService.emitCardChange → NotifyCardDeleted → publishConfigChange → rediscommon.PublishToStream

**上层调用文件**: [internal/service/card_sync_service.go](internal/service/card_sync_service.go#L207)

---

### 5. 报警处理消息
**文件路径**: [internal/service/alarm_event_service.go](internal/service/alarm_event_service.go)  
**函数名**: `HandleAlarmEvent`  
**Stream 名称**: `config:alarm.process:stream`  
**消息类型**: `config.alarm.process.ack` (CloudEvents 格式)  
**代码行号**: [807](internal/service/alarm_event_service.go#L807)

**说明**: 当用户处理（确认或解决）报警事件时，异步发送报警处理消息给 cardagg 消费。消息包含卡片 ID、设备 ID、报警级别、事件类型和时间戳。采用异步发送方式（goroutine），不阻止主流程。

**调用链**: HandleAlarmEvent (异步 goroutine) → rediscommon.BuildAlarmProcessMessage → rediscommon.PublishJSONToStream

---

## Stream 统计

| Stream 名称 | 消息数量 | 消息类型 |
|-----------|--------|--------|
| `config:device_status:stream` | 4 | device.updated, card.created, card.updated, card.deleted |
| `config:alarm.process:stream` | 1 | alarm.process.ack |

## 消息格式

所有消息均采用 **CloudEvents 标准格式**，包含以下字段：
- `specversion`: CloudEvents 规范版本
- `id`: 消息唯一标识符
- `source`: 消息来源（均为 "wisefido-data"）
- `type`: 消息类型（见上表）
- `time`: 事件发生时间（ISO 8601 格式）
- `data`: 消息数据体（JSON 字符串）

## 消费者

- **cardagg 服务**: 消费 `config:device_status:stream` 中的卡片配置变更消息
- **cardagg 服务**: 消費 `config:alarm.process:stream` 中的报警处理消息

## 发送方式

- **同步发送**: NotifyAlarmDeviceUpdated, NotifyCardCreated, NotifyCardUpdated, NotifyCardDeleted
  - 使用 `rediscommon.PublishToStream()`
  - 发送失败时记录错误日志但继续执行

- **异步发送**: HandleAlarmEvent 中的报警处理消息
  - 在独立 goroutine 中发送
  - 使用 `rediscommon.PublishJSONToStream()`
  - 发送失败时记录 WARN 级别日志

## 相关配置

所有 streams 的最大长度均为 **1000** 条消息，无时间限制保留。配置定义在 [owl-common/redis/stream_names.go](../owl-common/redis/stream_names.go)。
