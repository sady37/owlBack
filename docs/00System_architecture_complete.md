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
│                   数据存储层 (Data Storage) ✅ 已实现                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────┐               │
│  │         wisefido-iot-timeseries                      │               │
│  │                                                       │               │
│  │  - 消费 iot:monitor:stream                           │               │
│  │  - 消费 iot:stat:stream                              │               │
│  │  - 消费 iot:event:stream                             │               │
│  │  - 消费 iot:alarm:stream                             │               │
│  │  - 存储到 PostgreSQL (iot_timeseries)                │               │
│  │  - 发布到 iot:data:stream (触发下游服务)             │               │
│  └───────────────────────┬───────────────────────────────┘               │
│                          │                                               │
│                          ▼                                               │
│  ┌─────────────────────────────────────────────────────┐               │
│  │  PostgreSQL - iot_timeseries 表                      │               │
│  │  (HIPAA/FDA 合规存储，只保存转换后的标准值)           │               │
│  └─────────────────────────────────────────────────────┘               │
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
│  │  - Redis Streams: iot:*:stream (设备数据事件)     │               │
│  │  - PostgreSQL: cards 表（基础信息）                   │               │
│  │  - PostgreSQL: iot_timeseries 表（设备原始数据）      │               │
│  │  - PostgreSQL: alarm_events 表（报警数据）            │               │
│  │                                                       │               │
│  │  处理流程（2秒聚合一次）：                            │               │
│  │  1. 直接消费 iot:data:stream（事件驱动）              │               │
│  │  2. 检测设备直接报警并同步更新 alarm_events          │               │
│  │  3. 从 iot_timeseries 读取设备原始数据                │               │
│  │  4. 按源分存与 display：realtime 存 radar/sleepad，  │               │
│  │     mergeRealtimeData 按 Sleepad>Radar 算 display：  │               │
│  │     Heart/Breath/HeartSource/BreathSource、          │               │
│  │     SleepStage、BedStatus；姿态来自 Radar（含 ActiveBed │               │
│  │     场景 A 下并入的未绑床 Radar，如 bathroom）        │               │
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
│  │  - 读取 full 中的实时 display（由 mergeRealtimeData 计算）            │               │
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
    ├─→ wisefido-radar (使用 owl-common/encode.RadarEncode 转换) → Redis Streams (iot:monitor/stat/event/alarm:stream)
    └─→ wisefido-sleepace (使用 owl-common/encode.SleepaceEncode 转换) → Redis Streams (iot:monitor/event/alarm:stream)
    │
    ▼
wisefido-iot-timeseries (数据存储服务)
    │
    ├─ 消费 iot:monitor:stream (实时数据)
    ├─ 消费 iot:stat:stream (统计数据)
    ├─ 消费 iot:event:stream (事件数据)
    ├─ 消费 iot:alarm:stream (告警数据)
    ├─→ PostgreSQL (iot_timeseries) - 持久化存储 (HIPAA/FDA 合规)
    └─→ Redis Streams (iot:data:stream) - 触发下游
    │
    ▼
wisefido-card-aggregator (卡片数据聚合服务，2秒聚合一次)
    │
    ├─ 直接消费 iot:monitor/stat/event/alarm:stream（事件驱动）
    ├─ 检测设备直接报警 (Fall, SuspectedFall, OfflineAlarm 等)
    │   └─→ PostgreSQL (alarm_events - 设备报警)
    ├─ 根据 device_id 查询 cards 表
    │   └─ 找到该设备所属的卡片 (card_id)
    ├─ 查询该卡片的所有设备 (cards.devices JSONB)
    ├─ 从 iot_timeseries / 缓存 获取这些设备的最新数据
    ├─ 按源分存（不聚合），便于 card 主动判断与 vue-radar 手工选择：
    │   - Radar:   HR, RR, sleepStatus（含 bed_status）
    │   - Sleepad: HR, RR, sleepStatus
    │   - 姿态数据: 仍用所有 Radar 数据（仅 Radar 有）
    │   - display（HRdisplay, RRdisplay, sleepStatusDisplay）: 不在此计算，由 card/vue 在展示时按 HR_sleep>HR_radar、RR_sleep>RR_radar 等规则主动判断
    ├─ 读取卡片基础信息（从 cards 表）
    ├─ 读取报警数据（从 alarm_events 或 Redis）
    ├─ 聚合完整的 VitalFocusCard 对象
    └─→ Redis (vital-focus:card:{card_id}:full, TTL: 10秒)
    │
    ▼
wisefido-ai (AI智能推理)
    │
    ├─ 读取 vital-focus:card:{card_id}:full
    ├─ 读取云端报警规则 (alarm_cloud)
    ├─ 评估云端推理事件 (事件1, 3, 4)
    ├─ 访客识别和智能分析
    ├─ 巡房优化和模式识别
    ├─→ PostgreSQL (alarm_events - 云端报警)
    └─→ Redis (vital-focus:card:{card_id}:alarms)
    │
    ▼
wisefido-data (API)
    │
    ├─ 仅订阅 config:device_status:stream（设备在线）；不订阅 iot:monitor/stat 等（由 card-aggregator 的 iot_stream_consumer 消费）
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
- **触发**：设备数据到达 `iot:data:stream` / `iot:monitor/event/alarm:stream`
- **查询**：根据 `device_id` 查询该设备所属的卡片
- **按源分存**：该卡片上 Radar、Sleepad 的 HR/RR/sleepStatus **分源存储**，不做单一融合；display（HRdisplay、RRdisplay、sleepStatusDisplay）由 **card 或 vue 在展示时主动判断**（HR_sleep>HR_radar 等），vue-radar 可手工选择看 Radar / Sleepad / Auto
- **输出**：以卡片为单位输出实时数据（含 Radar 与 Sleepad 分源字段）

### 3. 服务职责分离
- **wisefido-radar / wisefido-sleepace**：设备数据采集和转换（使用 `owl-common/encode` 公共库）
- **wisefido-iot-timeseries**：数据存储服务（消费 Redis Streams，存储到 PostgreSQL）
- **wisefido-card-manage**：卡片创建和维护（低频操作）
- **wisefido-card-aggregator**：卡片数据聚合（高频操作，2秒一次）
  - 消费 `iot:data:stream` / `iot:monitor/event/alarm:stream`（事件驱动）
  - 检测设备直接报警并更新 `alarm_events`
  - 按源分存 Radar / Sleepad 的 HR、RR、sleepStatus，**不聚合**；姿态用 Radar；display 由 card/vue 展示时按 HR_sleep>HR_radar 等规则计算，vue-radar 可手工选源
  - 聚合完整卡片数据（含分源 vital 与姿态）
- **wisefido-ai**：AI智能推理（云端报警评估、访客识别、巡房优化）
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
1. ✅ 数据采集层（wisefido-radar, wisefido-sleepace）← **已实现，使用 owl-common/encode 进行数据转换**
2. ✅ 数据存储层（wisefido-iot-timeseries）← **已实现**
3. ✅ **卡片管理层（wisefido-card-aggregator - 卡片创建）** ← **已实现**
4. ✅ 传感器融合层（wisefido-sensor-fusion）← **功能已整合到 wisefido-card-aggregator**
5. ✅ 报警评估层（wisefido-alarm）← **已实现**
6. ✅ 卡片聚合层（wisefido-card-aggregator - 数据聚合）← **已实现**
7. ✅ API 服务层（wisefido-data）← **已实现**

### 问题：wisefido-sensor-fusion 的流程

**当前流程**：
1. 消费 `iot:data:stream` / `iot:monitor/event/alarm:stream`（单条设备数据）
2. 根据 `device_id` 查询卡片 ← **如果 cards 表为空，这里会失败**
3. 按源分存 Radar / Sleepad 的 HR、RR、sleepStatus（不聚合）；姿态用 Radar
4. 更新 Redis 缓存（realtime 含分源字段，无 display 融合）

**潜在问题**：
- ✅ 流程基本正确：设备数据 → 找到卡片 → 融合卡片数据
- ⚠️ **但前提是 cards 表必须有数据**（需要先实现卡片创建功能）

## 📝 总结

系统架构是**卡片中心**的：
1. **卡片预先存在**（由 card-aggregator 创建和维护）✅ **已实现**
2. **数据流以卡片为单位**（设备数据触发，但处理卡片数据）
3. **服务职责清晰**（采集 → **存储** → **卡片创建** → **数据聚合** → 报警 → API）

### ✅ 当前实现状态

**已实现** ✅：
- 数据采集层（wisefido-radar, wisefido-sleepace）
  - 使用 `owl-common/encode` 公共库进行数据转换 ✅
- 数据存储层（wisefido-iot-timeseries）✅ **已实现**
  - 消费 Redis Streams (iot:monitor/stat/event/alarm:stream) ✅
  - 存储到 PostgreSQL (iot_timeseries) ✅
  - 发布到 iot:data:stream (触发下游) ✅
- **卡片管理层（wisefido-card-manage - 卡片创建）** ✅ **已完成**
- **卡片数据聚合层（wisefido-card-aggregator）** ✅ **已实现**
  - 消费 `iot:data:stream`（事件驱动）✅
  - 检测设备直接报警并更新 `alarm_events` ✅
  - 从缓存按源聚合（radar/sleepad），mergeRealtimeData 算 display（HR/RR、床状态、姿态等）✅
  - 聚合完整卡片数据 ✅
- **AI智能推理层（wisefido-ai）** ✅ **已实现**
  - 云端高级推理报警（事件1, 3, 4）✅
  - 访客识别和智能分析 ✅
  - 巡房优化和模式识别 ✅
- **API 服务层（wisefido-data - HTTP API 功能）** ✅ **已实现**

**已移除** ❌：
- **wisefido-sensor-fusion**：功能已整合到 `wisefido-card-aggregator`
- **wisefido-data-transformer**：功能已迁移到 `owl-common/encode` 公共库和 `wisefido-iot-timeseries`

### 🔧 下一步行动

**已完成** ✅：
- 移除 `wisefido-data-transformer` 服务（功能已迁移到 `owl-common/encode` 公共库）
- 实现 `wisefido-iot-timeseries` 数据存储服务
- 验证 `wisefido-radar` 和 `wisefido-sleepace` 使用 `owl-common/encode` 进行数据转换

**优先级 1**：验证完整数据流
- 验证数据转换是否正确（`owl-common/encode` 公共库）
- 验证数据存储是否正常（`wisefido-iot-timeseries` → PostgreSQL）
- 验证数据流是否正常：设备 → 采集服务（带转换） → Redis Streams → 存储服务 → 下游服务

**优先级 2**：优化和验证
- 验证卡片创建逻辑是否正确（设备/住户/单元变化时）
- 验证卡片数据聚合是否正常（实时数据、报警数据）
- 性能优化（如有需要）                         



                                           vue前端-----创建关系/处理报警
                                           |
                                           |
radar--6topic-->mqtt网关-----查询/设置----〉data--------config:unit/user/device/alarm:strem
                                  |                          | 
                                  |                          |-----card_create订阅unit/user/device: 创建/更新卡片                                    
                                  |                          | 
                                  |                          |-----radar&sleep订阅alarm: 更新报警使能表                                      
                                  |                           
                                  |
                                  |
                 monintor/statistisc/even/alarm  解码标准化
                                  |
                       使用 device_alarm_enabled filter
                                  |
                                  |
                  iot:auth/monintor/statistisc/even/alarm:stream
                  |
                  |----------IoT_timeseries subcribe all---> pgsql_timeDB
                  |
                  |
                  |----------Card_Aggagator subcribe monitor/event/alarm----|----monitor:stream:按源分存 Radar/Sleepad(HR,RR,sleepStatus)，卡片/vue展示时主动判断 display(HR_sleep>HR_radar 等)，vue-radar 可手工选源---vital-focus:card:*--vue前端
                  |                                                         | 
                  |                                                         |----event:strem:left/enterBed 展示
                  |                                                         |                  
                  |                                                         |____alarm:strem: 根据alarm级别在卡片展，写入alarm_evnet库     
                  |
                  |
                  |----------AI: subcribe all----|-----alarm--|---自动计时并跟踪，自动recover一些自恢复报警
                                                 |            
                                                 |-----event/statistisc：高级报警推理--老人在bathroom >15分钟站立即异常（人不能长站，可能跌停但雷达干扰显示长站不动）           
                                                 | 
                                                 |
                                                 |----monitor-|-任何超过alarm_cloud 阀值----filter同unit所有设备，跟踪30秒--如异常，发送 iot:alarm:stream 
                                                 |
                                                 |——读取IoT_DB:建立行为基线                


#### monitor 传 vue-radar、Card_Aggregator 与 data、HR/RR 优先级

**1. monitor 到 vue-radar 的路径**

- MQTT → 网关解码 → `iot:auth/monitor/statistics/event/alarm:stream`（文档中的 `iot:auth`，实现里多为 `iot:monitor:stream` 等）  
- **Card_Aggregator** 订阅 monitor/event/alarm，**按源分存**（不聚合）写入 Redis：
  - `vital-focus:card:{card_id}:realtime`：Radar 的 HR、RR、sleepStatus；Sleepad 的 HR、RR、sleepStatus；姿态（仅 Radar）。**不计算** display（HRdisplay/RRdisplay/sleepStatusDisplay），由 card 或 vue 展示时按规则算
  - `vital-focus:card:{card_id}:device:{device_id}:data`：该卡下某设备的**最新** monitor（track/vital 等），TTL 6s（2s×3）
  - `vital-focus:card:{card_id}:full`：完整卡片（含 realtime + alarms），供卡片列表/详情
- **vue-radar**（单雷达页 `/radar/:deviceId`）通过 **wisefido-data**：
  - `GET /radar-device/api/v1/radar-device/device/:id/realtime`（当前返回空，待实现）  
- **GetRealtimeData 建议实现**（data 不订阅 iot，只读 Redis）：
  1. `device_id` + `tenant_id` → `GetCardIDByDeviceID` 得 `card_id`
  2. **positions**：读 `vital-focus:card:{card_id}:device:{device_id}:data`，从 `PositionX/Y`、`Timestamp` 转成 `{x,y,timestamp}[]`（仅最新一条时为单点；若需短轨迹可日后由 Card_Aggregator 写 `vital-focus:device:{device_id}:track` 等）
  3. **vital**：读 `vital-focus:card:{card_id}:realtime` 的 **分源** 字段（`radar: {HR,RR,sleepStatus}`、`sleepad: {HR,RR,sleepStatus}`）；返回时可按 `vitalSource` 参数：`auto`（HRdisplay=HR_sleep>HR_radar，RRdisplay=RR_sleep>RR_radar）、`radar`、`sleepad`，以支持 vue-radar **手工选择看哪一个**

**2. Card_Aggregator 独立 vs 放进 data**

- **结论：Card_Aggregator 保持独立。**
- 理由：
  - 消费 `iot:monitor/event/alarm:stream` 是持续高吞吐，与 data 的请求-响应模式不同，独立扩缩容更合适。
  - 架构图里 Card_Aggregator 与 IoT_timeseries、AI 并列消费 iot，适合作为独立消费者。
  - **data 无需订阅 Redis iot stream**：只读 Card_Aggregator 写好的 `vital-focus:card:*` 即可；GetRealtimeData、卡片 API 均从 Redis KV 读。
- 若 data 直接订阅 `iot:monitor:stream`：仅在需要**按 device 的原始 monitor 流**且 Card_Aggregator 不写 device 级缓存时才有用；当前 Card_Aggregator 已写 `vital-focus:card:{card_id}:device:{device_id}:data`，不必要。

**3. redis:card:stream 与 实现 的对应**

- 文档中「monitor:stream 简单融合，卡片展示 → redis:card:stream → vue」在实现中对应：
  - **不是** 新的 `redis:card:stream`，而是 **Redis KV**：`vital-focus:card:{card_id}:full`、`vital-focus:card:{card_id}:realtime`。
  - 卡片列表/详情的 Vue 通过 data 的 HTTP 读 `vital-focus:card:{id}:full`。

**4. HR/RR/sleepStatus：按源分存，display 由 card 主动判断与 vue-radar 手工选择**

- **不聚合、按源存**：  
  - **Radar**：HR, RR, sleepStatus（及 bed_status）  
  - **Sleepad**：HR, RR, sleepStatus  
  - **display（HRdisplay, RRdisplay, sleepStatusDisplay）**：**不在 Card_Aggregator 计算**，由 **card**（卡片展示、API 返回）或 **vue**（含 vue-radar）在**展示时主动判断**，规则：`HRdisplay = HR_sleep 有则取 else HR_radar`，`RRdisplay = RR_sleep 有则取 else RR_radar`，sleepStatus 同理；card 实现该规则即可，也很容易。
- **vue-radar**：支持 **手工选择** 看哪一个：`Radar` | `Sleepad` | `Auto`（Auto 即按上规则）。GetRealtimeData 可加 `vitalSource=auto|radar|sleepad`，返回对应源或按 Auto 算出的 display。
- **RealtimeData / realtime 结构建议**：分源字段如 `radar: { heart, breath, sleep_status [, bed_status ] }`、`sleepad: { heart, breath, sleep_status [, bed_status ] }`；不再使用单一的 `Heart`/`Breath`/`HeartSource`/`BreathSource` 融合结果；display 仅在读端按规则或用户选择计算。


#### 按设备类型分类

| FHIR Category      | 报警项                              | 设备类型   | 说明                              | 默认 DangerLevel |
|:-------------------|:-----------------------------------|:----------|:----------------------------------|:-----------------|
| **`safety`**       | `Fall`                             | Radar     | 跌倒                              | EMERGENCY        |
|                    | `SuspectedFall`                    | Radar     | 可疑跌倒                           | WARNING          |
|                    | `Stay`                             | Radar     | 滞留（卫生间/浴室）                  | WARNING          |
| **`clinical`**     | `Radar_ApneaHypopnea`              | Radar     | 呼吸暂停                           | DISABLE        |
|                    | `Radar_AbnormalHeartRate`          | Radar     | 心率异常（心动过速/心动过缓）         | EMERGENCY        |
|                    | `Radar_AbnormalRespiratoryRate`    | Radar     | 呼吸频率异常（呼吸急促/呼吸缓慢）      | EMERGENCY        |
|                    | `VitalsWeak`                       | Radar     | 生命体征微弱                        | WARNING          |
|                    | `SleepPad_ApneaHypopnea`           | SleepPad  | 呼吸暂停                           | EMERGENCY        |
|                    | `SleepPad_AbnormalHeartRate`       | SleepPad  | 心率异常（心动过速/心动过缓）          | EMERGENCY        |
|                    | `SleepPad_AbnormalRespiratoryRate` | SleepPad  | 呼吸频率异常（呼吸急促/呼吸缓慢）      | EMERGENCY        |
|                    | `SleepPad_AbnormalBodyMovement`    | SleepPad  | 异常体动（超2H未翻身/2H无体动）       | WARNING          |
| **`behavioral`**   | `Radar_LeftBed`                    | Radar     | 离床                              | WARNING          |
|                    | `SleepPad_LeftBed`                 | SleepPad  | 离床                              | WARNING          |
|                    | `SleepPad_BedSitUp`                | SleepPad  | 床上坐起                           | WARNING          |
|                    | `SleepPad_InBed`                   | SleepPad  | 上床（取决于住户service_level）     | DISABLE          |
|                    | `NoActivity24h`                    | Radar     | 24小时无人                         | EMERGENCY        |//但本质是行为缺失。
| **`device`**       | `OfflineAlarm`                     | 所有设备   | 设备离线                           | WARNING          |
|                    | `LowBattery`                       | 所有设备   | 低电量                             | WARNING          |
|                    | `DeviceFailure`                    | 所有设备   | 设备故障                           | EMERGENCY        |
|                    | `AngleException`                   | Radar     | 角度异常                           | WARNING          |



iot:monitor:stream   Maxtime: 30秒
1.fall/SuspectedFall        alarm_enable:转为alarm    
2.other_pose                card_aggagator只展示，不报警
3.HR/BR 呼吸心率超过阀值       card_aggagator只展示，不报警
    

iot:stastisc:stream   Maxtime: 5min    仅Radar支持，1分钟/次
1.vital_weak       card_aggagator立即报警，写alarm_event表
2.HR/BR过高/低    AI处理，即读取实时，滑动窗口：前30秒
3.track:stand    AI处理，即读取实时，滑动窗口：15分钟(老人很难连续站立15分钟)


iot:even:stream   Maxtime: 24H   
if alarm_enable=T   转为iot:Alarm:stream 
1。HR/BR      
2. fall  
3. 进出床     Card_aggagator:  睡眠时段段：card显示，
             AI：跟踪
4. 进出bath   Card_aggagator:  card显示，易跌  
             AI：跟踪
5. device   Card_aggagator: 可自动恢复
6. 其它               


iot:alarm:stream  Maxtime:2H    事件触发
1.fall/SitonBed   card_aggagator立即报警，写alarm_event表，禁止AI 恢复，必须人工处理
2.HR/BR 呼吸      card_aggagator立即报警，写alarm_event表  禁止AI 恢复，必须人工处理
3.Device事件      card_aggagator立即报警，写alarm_event表，card_aggagator可自动恢复
4.其它人为事件     card_aggagator立即报警，写alarm_event表  可AI 恢复，必须人工处理      

