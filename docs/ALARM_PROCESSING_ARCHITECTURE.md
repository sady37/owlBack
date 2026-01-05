# 报警处理架构说明

## 📋 概述

系统中有两类报警需要处理：
1. **设备直接报警**：设备直接上报的报警事件（如 Fall, OfflineAlarm 等）
2. **云端事件报警**：由云端规则评估产生的报警（事件1-4）

## 🏗️ 架构设计

### 职责划分

```
设备数据流：
设备 → MQTT → wisefido-radar/sleepace
    ↓
wisefido-data-transformer
    ├─ 转换数据 → iot_timeseries (包含 event_type)
    └─ 发布事件 → iot:data:stream (包含 event_type)
    ↓
wisefido-sensor-fusion
    ├─ 融合数据 → Redis (realtime)
    └─ 【设备直接报警处理】检测 event_type → alarm_events
    ↓
wisefido-alarm
    └─ 【云端事件报警评估】评估规则（事件1-4）→ alarm_events
```

### 1. 设备直接报警处理（wisefido-sensor-fusion）

**处理位置**：`wisefido-sensor-fusion/internal/consumer/stream_consumer.go`

**触发时机**：
- 消费 `iot:data:stream` 消息时
- 检测到 `event_type` 字段存在
- 判断是否是设备直接报警类型

**处理流程**：
1. 检测 `iotData.EventType` 是否存在
2. 调用 `IsDeviceDirectAlarm` 判断是否是设备直接报警
3. 调用 `IsAlarmEnabled` 检查设备配置中是否启用
4. 构建报警事件（`BuildDeviceAlarmEvent`）
5. 创建 `alarm_events` 记录

**设备直接报警类型**：
- `Fall` - 跌倒
- `SuspectedFall` - 可疑跌倒
- `OfflineAlarm` - 设备离线
- `LowBattery` - 低电量
- `DeviceFailure` - 设备故障
- `AngleException` - 角度异常
- `LeftBed` - 离床
- `SitUp` - 坐起
- 等

**优点**：
- ✅ 事件驱动，实时性好（0延迟）
- ✅ 在数据流中处理，不需要额外查询
- ✅ 逻辑简单，职责清晰

### 2. 云端事件报警处理（wisefido-alarm）

**处理位置**：`wisefido-alarm/internal/consumer/cache_consumer.go`

**触发时机**：
- 轮询模式：每 10 秒轮询一次
- 读取融合后的实时数据（`vital-focus:card:{card_id}:realtime`）
- 评估云端规则（事件1-4）

**处理流程**：
1. 读取所有卡片的实时数据
2. 调用 `Evaluator.Evaluate` 评估云端事件
3. 事件1：床上跌落检测（sleepad离床事件 + radar存在）
4. 事件3：Bathroom可疑跌倒检测（基于卡片数据更新触发）
5. 事件4：雷达检测到人突然消失（基于卡片数据更新触发）
6. 创建 `alarm_events` 记录

**云端事件类型**：
- **事件1**：床上跌落检测（`AI-fall` / `SuspectedFall`）
- **事件3**：Bathroom可疑跌倒检测（`AI-suspected fall`）
- **事件4**：雷达检测到人突然消失（`AI-suspected fall`）

**注意**：
- 事件2（Sleepad可靠性判断）已移除
- 事件1 基于 sleepad 离床事件触发（事件驱动）
- 事件3/4 基于卡片数据更新触发（2秒1次）

**优点**：
- ✅ 规则评估逻辑集中
- ✅ 支持复杂的状态管理和定时器
- ✅ 可以访问融合后的实时数据

## 🔄 数据流

### 设备直接报警流程

```
设备上报 Fall → MQTT
    ↓
wisefido-radar → Redis Streams
    ↓
wisefido-data-transformer
    ├─ 转换 event_type = "Fall"
    ├─ 写入 iot_timeseries (event_type = "Fall")
    └─ 发布 iot:data:stream (包含 event_type)
    ↓
wisefido-sensor-fusion
    ├─ 检测 event_type = "Fall"
    ├─ 判断是设备直接报警
    ├─ 检查设备配置（是否启用）
    └─ 创建 alarm_events 记录
    ↓
PostgreSQL (alarm_events)
    ↓
前端显示报警
```

### 云端事件报警流程

```
设备数据 → wisefido-sensor-fusion → Redis (realtime)
    ↓
wisefido-alarm (每10秒轮询)
    ├─ 读取实时数据
    ├─ 评估事件1（床上跌落检测）
    ├─ 评估事件3（Bathroom可疑跌倒）
    ├─ 评估事件4（人突然消失）
    └─ 创建 alarm_events 记录
    ↓
PostgreSQL (alarm_events)
    ↓
前端显示报警
```

## 📊 对比

| 项目 | 设备直接报警 | 云端事件报警 |
|------|------------|------------|
| **处理服务** | `wisefido-sensor-fusion` | `wisefido-alarm` |
| **触发方式** | 事件驱动（立即） | 轮询（10秒）或事件驱动（事件1） |
| **数据来源** | `iot:data:stream` 消息中的 `event_type` | Redis 实时数据 + 规则评估 |
| **报警类型** | Fall, OfflineAlarm 等 | 事件1, 3, 4 |
| **延迟** | 0延迟 | 0-10秒（轮询）或立即（事件1） |
| **配置检查** | 检查 `alarm_device` 表 | 检查 `alarm_cloud` 和 `alarm_device` 表 |

## ✅ 实施状态

### 已完成 ✅

- [x] 在 `wisefido-sensor-fusion` 中添加报警处理依赖
- [x] 实现设备直接报警检测和处理逻辑
- [x] 创建报警工具函数
- [x] 更新 `wisefido-data-transformer` 发布 `event_type`
- [x] 更新文档，明确职责划分

### 待测试 ⏳

- [ ] 测试设备直接报警（Fall）是否能正确创建 `alarm_events`
- [ ] 验证云端事件报警不受影响（事件1-4）
- [ ] 验证两个报警处理流程互不干扰

## 🎯 关键点

1. **职责清晰**：
   - `wisefido-sensor-fusion`：处理设备直接报警
   - `wisefido-alarm`：处理云端事件报警

2. **互不干扰**：
   - 两个服务独立处理各自的报警类型
   - 不会重复创建相同的报警事件

3. **实时性**：
   - 设备直接报警：事件驱动，0延迟
   - 云端事件报警：轮询或事件驱动，可能有延迟

4. **配置检查**：
   - 两个服务都会检查设备配置，只有启用的报警才创建记录

