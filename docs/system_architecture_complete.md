# 完整系统架构图

## 🏗️ 系统整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           IoT 设备层                                      │
├─────────────────────────────────────────────────────────────────────────┤
│  Radar 设备          Sleepace 设备 (SleepPad)                             │
│     │                      │                                              │
│     └──────────┬───────────┘                                              │
│                │                                                          │
│                ▼                                                          │
│         ┌──────────────┐                                                 │
│         │ MQTT Broker  │                                                 │
│         └──────────────┘                                                 │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      数据采集层 (Device Collection)                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────┐      ┌─────────────────────┐                   │
│  │ wisefido-radar      │      │ wisefido-sleepace    │                   │
│  │                     │      │                     │                   │
│  │ - MQTT 订阅         │      │ - MQTT 订阅         │                   │
│  │ - 数据解析          │      │ - 数据解析          │                   │
│  │ - 发布到 Streams    │      │ - 发布到 Streams    │                   │
│  └──────────┬──────────┘      └──────────┬──────────┘                   │
│             │                            │                               │
│             └────────────┬───────────────┘                               │
│                          ▼                                               │
│              ┌───────────────────────┐                                    │
│              │   Redis Streams      │                                    │
│              │  - radar:data:stream │                                    │
│              │  - sleepace:data:... │                                    │
│              └───────────────────────┘                                    │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   数据转换层 (Data Transformation)                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         wisefido-data-transformer                    │               │
│  │                                                       │               │
│  │  - 消费 radar:data:stream                            │               │
│  │  - 消费 sleepace:data:stream                         │               │
│  │  - SNOMED CT 映射                                     │               │
│  │  - FHIR Category 分类                                 │               │
│  │  - 数据标准化                                         │               │
│  └───────────────┬───────────────────────┬───────────────┘               │
│                  │                       │                               │
│                  ▼                       ▼                               │
│  ┌──────────────────────┐   ┌──────────────────────┐                    │
│  │  PostgreSQL          │   │  Redis Streams       │                    │
│  │  TimescaleDB         │   │  iot:data:stream     │                    │
│  │  iot_timeseries      │   │  (标准化数据事件)     │                    │
│  └──────────────────────┘   └──────────────────────┘                    │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   卡片管理层 (Card Management) ✅ 已实现                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         wisefido-card-manage                        │               │
│  │  (卡片创建和维护服务) ✅ 已实现                       │               │
│  │                                                       │               │
│  │  触发条件：                                           │               │
│  │  - API 触发：wisefido-data 的 service 层直接调用     │               │
│  │    * 设备绑定/解绑 API → cardCreator.CreateCardsForUnit │
│  │    * 住户绑定/解绑 API → cardCreator.CreateCardsForUnit │
│  │    * 单元信息更新 API → cardCreator.CreateCardsForUnit │
│  │  - 定时兜底：wisefido-card-manage 定时全量更新      │               │
│  │                                                       │               │
│  │  处理逻辑：                                           │               │
│  │  - 根据规则创建/更新 cards 表                        │               │
│  │  - 计算卡片绑定的设备列表 (devices JSONB)            │               │
│  │  - 计算卡片关联的住户列表 (residents JSONB)         │               │
│  │  - 更新卡片基础信息 (card_name, card_address)        │               │
│  │  - 更新报警路由配置 (routing_alarm_user_ids, tags)   │               │
│  └───────────────┬───────────────────────────────────────┘               │
│                  │                                                       │
│                  ▼                                                       │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         PostgreSQL - cards 表                        │               │
│  │                                                       │               │
│  │  - card_id, tenant_id, card_type                     │               │
│  │  - bed_id (ActiveBed) / unit_id (Location)          │               │
│  │  - card_name, card_address                          │               │
│  │  - devices (JSONB) - 预计算的设备列表               │               │
│  │  - residents (JSONB) - 预计算的住户列表             │               │
│  │  - routing_alarm_user_ids, routing_alarm_tags       │               │
│  └─────────────────────────────────────────────────────┘               │
│                                                                           │
│  ✅ 卡片更新机制：                                         │               │
│     - wisefido-data 的 service 层直接调用 wisefido-card-manage API      │
│     - 设备/住户/单元变化时，同步更新 cards 表（实时响应）                │
│     - wisefido-card-manage 的轮询模式仅作为保底机制（定时全量更新）    │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              │ ✅ 卡片已创建
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   卡片数据聚合层 (Card Data Aggregation)                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         wisefido-card-aggregator                     │               │
│  │  (卡片数据聚合服务) ✅ 已实现                         │               │
│  │                                                       │               │
│  │  输入：                                               │               │
│  │  - Redis Streams: iot:data:stream (设备数据事件)     │               │
│  │  - PostgreSQL: cards 表（基础信息）                   │               │
│  │  - PostgreSQL: iot_timeseries 表（设备原始数据）      │               │
│  │  - PostgreSQL: alarm_events 表（报警数据）            │               │
│  │                                                       │               │
│  │  处理流程（2秒聚合一次）：                            │               │
│  │  1. 直接消费 iot:data:stream（事件驱动）              │               │
│  │  2. 检测设备直接报警并同步更新 alarm_events          │               │
│  │  3. 从 iot_timeseries 读取设备原始数据                │               │
│  │  4. 融合数据（真实事件的展示）：                      │               │
│  │     - HR/RR: 优先 Sleepace，无则 Radar               │               │
│  │     - 床状态/睡眠状态: 优先 Sleepace                  │               │
│  │     - 姿态数据: 使用所有 Radar 数据                   │               │
│  │  5. 读取卡片基础信息（从 cards 表）                    │               │
│  │  6. 读取报警数据（从 alarm_events 或 Redis）          │               │
│  │  7. 聚合完整的 VitalFocusCard 对象                   │               │
│  └───────────────┬───────────────────────────────────────┘               │
│                  │                                                       │
│                  ▼                                                       │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         Redis Cache                                  │               │
│  │         vital-focus:card:{card_id}:full              │               │
│  │         (完整的 VitalFocusCard, TTL: 10秒)           │               │
│  └─────────────────────────────────────────────────────┘               │
│                                                                           │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         wisefido-alarm                               │               │
│  │  (云端高级推理报警) ✅ 已实现                         │               │
│  │                                                       │               │
│  │  输入：                                               │               │
│  │  - Redis: vital-focus:card:{card_id}:full            │               │
│  │  - PostgreSQL: alarm_cloud (云端报警规则)             │               │
│  │                                                       │               │
│  │  处理：                                               │               │
│  │  - 读取融合后的实时数据                               │               │
│  │  - 应用云端推理规则 (事件1, 3, 4)                    │               │
│  │    * 事件1：床上跌落检测                              │               │
│  │    * 事件3：Bathroom可疑跌倒检测                      │               │
│  │    * 事件4：雷达检测到人突然消失                      │               │
│  │  - 生成报警事件                                       │               │
│  └───────────────┬───────────────────────┬───────────────┘               │
│                  │                       │                               │
│                  ▼                       ▼                               │
│  ┌──────────────────────┐   ┌──────────────────────┐                    │
│  │  PostgreSQL          │   │  Redis Cache         │                    │
│  │  alarm_events        │   │  vital-focus:card:    │                    │
│  │  (云端报警事件)       │   │  {card_id}:alarms    │                    │
│  │                      │   │  (TTL: 30秒)         │                    │
│  └──────────────────────┘   └──────────────────────┘                    │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       API 服务层 (API Service)                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         wisefido-data (HTTP API)                    │               │
│  │                                                       │               │
│  │  API 功能：                                           │               │
│  │  - 接收 HTTP 请求 (JWT Token)                        │               │
│  │  - 权限过滤 (tenant_id, role, caregiver_id)          │               │
│  │  - Focus 过滤 (selectedCardIds)                      │               │
│  │  - 从 Redis 读取 vital-focus:card:{card_id}:full     │               │
│  │  - 返回 VitalFocusCard[]                             │               │
│  │                                                       │               │
│  │  ✅ 卡片更新机制：                                     │               │
│  │  - 设备绑定/解绑 API → 直接调用 cardCreator.CreateCardsForUnit │
│  │  - 住户绑定/解绑 API → 直接调用 cardCreator.CreateCardsForUnit │
│  │  - 单元信息更新 API → 直接调用 cardCreator.CreateCardsForUnit │
│  │  - 同步更新 cards 表（cards 表是最终结果）            │               │
│  └───────────────┬───────────────────────────────────────┘               │
│                  │                                                       │
│                  ▼                                                       │
│  ┌─────────────────────────────────────────────────────┐               │
│  │                   前端 (Vue.js)                       │               │
│  └─────────────────────────────────────────────────────┘               │
└─────────────────────────────────────────────────────────────────────────┘
```

## 🔄 关键数据流

### 1. 卡片创建流程（设备/住户/床位变化时）

```
设备/住户/床位绑定关系变化
    │
    ▼
wisefido-data (API 服务) ✅ 已实现
    │
    ├─ 更新数据库 (devices/residents/beds/units 表)
    └─ 直接调用 wisefido-card-manage API  ← 同步更新
       │
       ▼
wisefido-card-manage (卡片管理服务) ✅ 已实现
       │
       ├─ 查询 devices, beds, residents, units 表
       ├─ 根据规则计算需要创建的卡片
       └─ 创建/更新 cards 表（cards 表是最终结果）
          ├─ card_type (ActiveBed / Location)
          ├─ devices (JSONB) - 该卡片绑定的所有设备
          ├─ residents (JSONB) - 该卡片关联的所有住户
          └─ routing_alarm_user_ids, routing_alarm_tags

定时兜底（wisefido-card-manage 定时任务）：
    │
    ▼
wisefido-card-manage (定时全量更新)
    │
    └─ 全量重新创建所有卡片（确保数据最终一致性，保底机制）
```

### 2. 实时数据流（设备数据 → 卡片数据）

```
IoT 设备发送数据
    │
    ▼
MQTT Broker
    │
    ├─→ wisefido-radar → Redis Streams (radar:data:stream)
    └─→ wisefido-sleepace → Redis Streams (sleepace:data:stream)
    │
    ▼
wisefido-data-transformer
    │
    ├─ 消费 radar:data:stream
    ├─ 消费 sleepace:data:stream
    ├─ 数据标准化 (SNOMED CT, FHIR Category)
    ├─→ PostgreSQL (iot_timeseries) - 持久化存储
    └─→ Redis Streams (iot:data:stream) - 触发下游
    │
    ▼
wisefido-card-aggregator (卡片数据聚合服务，2秒聚合一次)
    │
    ├─ 直接消费 iot:data:stream（事件驱动）
    ├─ 检测设备直接报警 (Fall, SuspectedFall, OfflineAlarm 等)
    │   └─→ PostgreSQL (alarm_events - 设备报警)
    ├─ 根据 device_id 查询 cards 表
    │   └─ 找到该设备所属的卡片 (card_id)
    ├─ 查询该卡片的所有设备 (cards.devices JSONB)
    ├─ 从 iot_timeseries 获取这些设备的最新数据
    ├─ 融合数据（真实事件的展示）：
    │   - HR/RR: 优先 Sleepace，无则 Radar
    │   - 床状态/睡眠状态: 优先 Sleepace
    │   - 姿态数据: 使用所有 Radar 数据
    ├─ 读取卡片基础信息（从 cards 表）
    ├─ 读取报警数据（从 alarm_events 或 Redis）
    ├─ 聚合完整的 VitalFocusCard 对象
    └─→ Redis (vital-focus:card:{card_id}:full, TTL: 10秒)
    │
    ▼
wisefido-alarm (云端高级推理报警)
    │
    ├─ 读取 vital-focus:card:{card_id}:full
    ├─ 读取云端报警规则 (alarm_cloud)
    ├─ 评估云端推理事件 (事件1, 3, 4)
    ├─→ PostgreSQL (alarm_events - 云端报警)
    └─→ Redis (vital-focus:card:{card_id}:alarms)
    │
    ▼
wisefido-data (API)
    │
    ├─ 接收 HTTP 请求
    ├─ 权限过滤
    ├─ 读取 vital-focus:card:{card_id}:full
    └─→ HTTP 响应 (VitalFocusCard[])
```

## 🎯 关键设计点

### 1. 卡片是预先存在的
- **卡片创建**：由 `wisefido-card-aggregator` 服务负责，当设备/住户/床位绑定关系变化时触发
- **卡片存储**：PostgreSQL `cards` 表，包含预计算的 `devices` 和 `residents` JSONB 字段
- **卡片类型**：`ActiveBed`（床位卡片）或 `Location`（门牌号卡片）

### 2. 数据流是卡片驱动的
- **触发**：设备数据到达 `iot:data:stream`
- **查询**：根据 `device_id` 查询该设备所属的卡片
- **融合**：融合该卡片上**所有设备**的数据（不是单个设备）
- **输出**：以卡片为单位输出融合后的实时数据

### 3. 服务职责分离
- **wisefido-card-manage**：卡片创建和维护（低频操作）
- **wisefido-card-aggregator**：卡片数据聚合（高频操作，2秒一次）
  - 消费 `iot:data:stream`（事件驱动）
  - 检测设备直接报警并更新 `alarm_events`
  - 从 `iot_timeseries` 读取设备数据并融合（HR/RR、床状态、姿态）
  - 聚合完整卡片数据
- **wisefido-alarm**：云端高级推理报警评估（事件1, 3, 4）
- **wisefido-data**：API 服务（HTTP 接口）

## ⚠️ 当前问题分析

### ⚠️ 关键问题：实现顺序错误

**问题**：
1. **wisefido-sensor-fusion** 已经实现 ✅
2. **wisefido-card-aggregator** 的卡片创建功能**未实现** ❌
3. **wisefido-sensor-fusion** 依赖 `cards` 表，但卡片还未创建

**影响**：
- 如果 `cards` 表为空，`wisefido-sensor-fusion` 无法工作
- `GetCardByDeviceID` 会返回错误，导致融合失败

**正确的实现顺序应该是**：
1. ✅ 数据采集层（wisefido-radar, wisefido-sleepace）
2. ✅ 数据转换层（wisefido-data-transformer）
3. ⚠️ **卡片管理层（wisefido-card-aggregator - 卡片创建）** ← **应该先实现这个**
4. ✅ 传感器融合层（wisefido-sensor-fusion）← **依赖卡片，应该后实现**
5. ⏳ 报警评估层（wisefido-alarm）
6. ⏳ 卡片聚合层（wisefido-card-aggregator - 数据聚合）
7. ⏳ API 服务层（wisefido-data）

### 问题：wisefido-sensor-fusion 的流程

**当前流程**：
1. 消费 `iot:data:stream`（单条设备数据）
2. 根据 `device_id` 查询卡片 ← **如果 cards 表为空，这里会失败**
3. 融合该卡片的所有设备数据
4. 更新 Redis 缓存

**潜在问题**：
- ✅ 流程基本正确：设备数据 → 找到卡片 → 融合卡片数据
- ⚠️ **但前提是 cards 表必须有数据**（需要先实现卡片创建功能）

## 📝 总结

系统架构是**卡片中心**的：
1. **卡片预先存在**（由 card-aggregator 创建和维护）✅ **已实现**
2. **数据流以卡片为单位**（设备数据触发，但处理卡片数据）
3. **服务职责清晰**（采集 → 转换 → **卡片创建** → **传感器融合** → 报警 → 聚合 → API）

### ✅ 当前实现状态

**已实现** ✅：
- 数据采集层（wisefido-radar, wisefido-sleepace）
- 数据转换层（wisefido-data-transformer）
- **卡片管理层（wisefido-card-manage - 卡片创建）** ✅ **已完成**
- **卡片数据聚合层（wisefido-card-aggregator）** ✅ **已实现**
  - 消费 `iot:data:stream`（事件驱动）✅
  - 检测设备直接报警并更新 `alarm_events` ✅
  - 从 `iot_timeseries` 读取设备数据并融合（HR/RR、床状态、姿态）✅
  - 聚合完整卡片数据 ✅
- **报警评估层（wisefido-alarm）** ✅ **已实现**
  - 云端高级推理报警（事件1, 3, 4）✅
- **API 服务层（wisefido-data - HTTP API 功能）** ✅ **已实现**

**已移除** ❌：
- **wisefido-sensor-fusion**：功能已整合到 `wisefido-card-aggregator`

### 🔧 下一步行动

**优先级 1**：架构调整和实现
- 拆分 `wisefido-card-aggregator` 为两个服务：
  - `wisefido-card-manage`：卡片创建和维护
  - `wisefido-card-aggregator`：卡片数据聚合
- 在 `wisefido-card-aggregator` 中实现：
  - 消费 `iot:data:stream`（事件驱动）
  - 检测设备直接报警并更新 `alarm_events`
  - 从 `iot_timeseries` 读取设备数据并融合（HR/RR、床状态、姿态）
- 移除 `wisefido-sensor-fusion` 服务
- 更新 `wisefido-data` 调用 `wisefido-card-manage` API

**优先级 2**：验证完整数据流
- 验证融合逻辑是否正确（HR/RR、床状态、姿态数据）
- 验证设备直接报警处理是否正常（Fall, SuspectedFall, OfflineAlarm 等）
- 验证云端推理报警评估是否正常（事件1, 3, 4）
- 验证 Redis 缓存更新是否正常（`vital-focus:card:{card_id}:full`, TTL: 10秒）
- 测试完整数据流：设备数据 → 数据融合 → 设备报警 → 云端推理报警 → 缓存更新

**优先级 2**：优化和验证
- 验证卡片创建逻辑是否正确（设备/住户/单元变化时）
- 验证卡片数据聚合是否正常（实时数据、报警数据）
- 性能优化（如有需要）

