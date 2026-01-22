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
- **wisefido-radar / wisefido-sleepace**：设备数据采集和转换（使用 `owl-common/encode` 公共库）
- **wisefido-iot-timeseries**：数据存储服务（消费 Redis Streams，存储到 PostgreSQL）
- **wisefido-card-manage**：卡片创建和维护（低频操作）
- **wisefido-card-aggregator**：卡片数据聚合（高频操作，2秒一次）
  - 消费 `iot:data:stream`（事件驱动）
  - 检测设备直接报警并更新 `alarm_events`
  - 从 `iot_timeseries` 读取设备数据并融合（HR/RR、床状态、姿态）
  - 聚合完整卡片数据
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
  - 从 `iot_timeseries` 读取设备数据并融合（HR/RR、床状态、姿态）✅
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
                  |----------Card_Aggagator subcribe monitor/event/alarm----|----monitor:strem:简单融合，卡片展示-------redis:card:stream--vue前端
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

