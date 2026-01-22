# Wisefido-AI 项目文件结构与依赖关系分析

## 项目概览

**项目名称**: wisefido-ai  
**项目类型**: Go 后端服务 (报警处理/AI 评估服务)  
**文件总数**: 33 个 Go 文件  
**主要依赖**: owl-common (内部模块), Redis, PostgreSQL, zap (日志)

---

## 目录结构

```
wisefido-ai/
├── cmd/                          # 入口点
│   └── wisefido-ai/
│       └── main.go              # 应用主入口
├── internal/                      # 内部模块 (按功能分层)
│   ├── config/                   # 配置管理层 (2 文件)
│   │   ├── config.go            # 配置结构定义
│   │   └── config_test.go       # 单元测试
│   │
│   ├── models/                   # 数据模型层 (4 文件)
│   │   ├── alarm_config.go      # 报警配置模型
│   │   ├── alarm_event.go       # 报警事件模型
│   │   ├── iot_data_message.go  # IoT 数据消息模型
│   │   └── realtime_data.go     # 实时数据模型
│   │
│   ├── repository/               # 数据访问层 (10 文件)
│   │   ├── alarm_cloud.go       # 云端报警数据访问
│   │   ├── alarm_device.go      # 设备报警数据访问
│   │   ├── alarm_events.go      # 报警事件数据访问 (含测试)
│   │   ├── card.go              # 卡片数据访问 (含测试)
│   │   ├── config_version.go    # 配置版本数据访问
│   │   ├── device.go            # 设备数据访问
│   │   ├── iot_timeseries.go    # IoT 时间序列数据访问
│   │   └── room.go              # 房间数据访问
│   │
│   ├── consumer/                 # 消费层 (事件驱动) (5 文件)
│   │   ├── cache_consumer.go    # Redis 缓存消费器
│   │   ├── cache_manager.go     # 缓存管理器 (含测试)
│   │   ├── event_consumer.go    # 事件消费器 (Stream/Queue)
│   │   └── state_manager.go     # 状态管理器
│   │
│   ├── evaluator/                # 评估层 (报警规则引擎) (9 文件)
│   │   ├── evaluator.go         # 核心评估引擎
│   │   ├── alarm_event_builder.go  # 报警事件构建器 (含测试)
│   │   ├── event1_bed_fall.go   # 事件1: 床上跌倒检测
│   │   ├── event1_helpers.go    # 事件1 辅助函数
│   │   ├── event2_sleepad_reliability.go  # 事件2: Sleepace 可靠性检测
│   │   ├── event3_bathroom_fall.go  # 事件3: 浴室跌倒检测
│   │   ├── event3_helpers.go    # 事件3 辅助函数
│   │   └── event4_sudden_disappear.go  # 事件4: 突然消失检测
│   │
│   └── service/                  # 业务逻辑层 (2 文件)
│       ├── alarm.go             # 报警服务 (核心编排)
│       └── alarm_event_service.go  # 报警事件服务
│
├── go.mod                        # Go 模块定义
├── go.sum                        # Go 依赖版本锁定
└── [文档文件]                    # 各类分析/设计文档

```

---

## 分层架构

### 1️⃣ **入口层** (cmd/)

| 文件 | 职责 | 依赖 |
|------|------|------|
| `cmd/wisefido-ai/main.go` | 应用初始化、启动各服务 | config, service, logger |

---

### 2️⃣ **业务层** (service/)

| 文件 | 职责 | 关键依赖 |
|------|------|---------|
| `service/alarm.go` | **核心编排** - 协调各组件完成报警处理 | config, consumer, evaluator, repository |
| `service/alarm_event_service.go` | 报警事件管理服务 | models, repository |

**处理流程**:
```
alarm.go (Service)
    ↓
    ├→ consumer (获取实时数据)
    ├→ evaluator (评估报警规则)
    └→ repository (保存结果)
```

---

### 3️⃣ **评估层** (evaluator/)

| 文件 | 职责 |
|------|------|
| `evaluator/evaluator.go` | **规则引擎** - 协调评估、调用各事件处理器 |
| `alarm_event_builder.go` | 构建报警事件对象 |
| `event1_bed_fall.go` + `event1_helpers.go` | **事件1**: 床上跌倒检测规则 |
| `event2_sleepad_reliability.go` | **事件2**: Sleepace 可靠性检测规则 |
| `event3_bathroom_fall.go` + `event3_helpers.go` | **事件3**: 浴室跌倒检测规则 |
| `event4_sudden_disappear.go` | **事件4**: 突然消失检测规则 |

**评估流程**:
```
evaluator.go
    ├→ event1_*.go (床上跌倒)
    ├→ event2_*.go (Sleepace 可靠性)
    ├→ event3_*.go (浴室跌倒)
    └→ event4_*.go (突然消失)
```

---

### 4️⃣ **消费层** (consumer/)

| 文件 | 职责 | 数据源 |
|------|------|--------|
| `event_consumer.go` | 消费 Redis Streams 事件 | Redis Streams |
| `cache_consumer.go` | 消费 Redis 缓存数据 | Redis Hash/String |
| `cache_manager.go` | 缓存数据的管理和更新 | Redis, 本地状态 |
| `state_manager.go` | 报警状态管理 | Redis |

**消费源**:
```
Redis Streams (event_consumer)
    ├→ radar:monitor:stream
    ├→ radar:stat:stream
    ├→ radar:event:stream
    ├→ sleepace:monitor:stream
    └→ sleepace:event:stream

Redis Hash/String (cache_consumer)
    └→ cache_manager (统一管理)
```

---

### 5️⃣ **数据访问层** (repository/)

| 文件 | 职责 | 操作对象 |
|------|------|---------|
| `alarm_cloud.go` | 云端报警数据访问 | 云平台 API |
| `alarm_device.go` | 设备报警数据访问 | 数据库 |
| `alarm_events.go` | 报警事件历史记录 | 数据库 + 测试 |
| `card.go` | 卡片信息访问 | 数据库 + 测试 |
| `device.go` | 设备信息访问 | 数据库 |
| `room.go` | 房间信息访问 | 数据库 |
| `config_version.go` | 配置版本控制 | 数据库 |
| `iot_timeseries.go` | 时间序列数据访问 | 数据库/时序库 |

---

### 6️⃣ **模型层** (models/)

| 文件 | 数据结构 |
|------|---------|
| `alarm_config.go` | 报警配置 |
| `alarm_event.go` | 报警事件 |
| `iot_data_message.go` | IoT 消息 |
| `realtime_data.go` | 实时数据 |

---

### 7️⃣ **配置层** (config/)

| 文件 | 职责 |
|------|------|
| `config.go` | 配置结构定义 + 加载 |
| `config_test.go` | 配置单元测试 |

**配置项**:
```go
Database       // 数据库配置 (PostgreSQL)
Redis          // Redis 配置
Alarm.Cache    // 缓存配置 (键前缀、TTL 等)
Alarm.Evaluation  // 评估配置 (批量大小)
Alarm.IoTStream   // 事件流配置 (enabled、consumer group)
```

---

## 文件依赖关系图

### **调用链路** (从上到下)

```
main.go
  │
  └─→ service/alarm.go (核心编排)
       │
       ├─→ config/config.go (读取配置)
       │
       ├─→ consumer/ (数据消费)
       │   ├─→ event_consumer.go
       │   └─→ cache_consumer.go
       │        └─→ cache_manager.go
       │
       ├─→ evaluator/ (规则评估)
       │   ├─→ evaluator.go
       │   ├─→ event1_*.go
       │   ├─→ event2_*.go
       │   ├─→ event3_*.go
       │   └─→ event4_*.go
       │
       └─→ repository/ (数据访问)
           ├─→ alarm_events.go (保存报警)
           ├─→ alarm_device.go
           ├─→ card.go (卡片信息)
           ├─→ device.go (设备信息)
           └─→ ...其他数据访问

models/
  └─→ 被所有层共享使用
```

---

## 核心依赖关系表

### **直接内部依赖**

| 源文件 | 依赖目标 | 依赖原因 |
|--------|---------|---------|
| main.go | config | 读取应用配置 |
| main.go | service | 启动报警服务 |
| service/alarm.go | config | 读取报警配置 |
| service/alarm.go | consumer | 消费实时数据 |
| service/alarm.go | evaluator | 执行报警评估 |
| service/alarm.go | repository | 保存/查询数据 |
| evaluator.go | consumer | 获取消费的数据 |
| evaluator.go | models | 使用数据结构 |
| evaluator/*.go | consumer | 访问消费数据 |
| evaluator/*.go | repository | 查询历史数据 |
| event_consumer.go | config | 读取消费配置 |
| event_consumer.go | repository | 保存消费数据 |
| cache_manager.go | config | 读取缓存配置 |
| cache_consumer.go | repository | 访问数据库 |
| alarm_events.go | models | 序列化/反序列化 |

### **外部依赖**

| 包 | 来源 | 用途 |
|-----|------|------|
| `owl-common/config` | 内部模块 (../owl-common) | 公共配置 (DB, Redis) |
| `owl-common/redis` | 内部模块 | Redis 连接管理 |
| `owl-common/logger` | 内部模块 | 日志系统 (zap) |
| `github.com/lib/pq` | 外部 | PostgreSQL 驱动 |
| `github.com/go-redis/redis/v8` | 外部 | Redis 客户端 |
| `go.uber.org/zap` | 外部 | 日志库 |

---

## 测试文件覆盖

| 被测试模块 | 测试文件 |
|-----------|---------|
| config | `config_test.go` |
| consumer/cache_manager | `cache_manager_test.go` |
| repository/alarm_events | `alarm_events_test.go` |
| repository/card | `card_test.go` |
| evaluator/alarm_event_builder | `alarm_event_builder_test.go` |
| consumer/mqtt | `mqtt_consumer_test.go` (如果存在) |

---

## 数据流向

### **实时报警处理流程**

```
外部数据源 (Radar/Sleepace 设备)
    │
    ├→ Redis Streams (event_consumer 消费)
    └→ Redis Cache (cache_consumer 消费)
    
    ↓
    
consumer/ (数据消费和缓存)
    ├→ event_consumer: 处理 streams 事件
    ├→ cache_manager: 管理实时数据缓存
    └→ state_manager: 维护报警状态
    
    ↓
    
service/alarm.go (核心编排)
    ├→ 获取消费数据
    └→ 触发评估
    
    ↓
    
evaluator/ (报警规则引擎)
    ├→ evaluator.go: 协调评估
    ├→ event1-4: 各类事件检测
    └→ alarm_event_builder: 构建报警事件
    
    ↓
    
repository/ (数据持久化)
    ├→ alarm_events.go: 保存报警历史
    ├→ alarm_device.go: 更新设备报警状态
    ├→ card.go: 更新卡片关联数据
    └→ alarm_cloud.go: 同步到云端
```

---

## 关键交互点

### **1. 配置初始化**
- `config.go` 定义了所有配置项
- 包括数据库、Redis、报警规则、IoT 流配置
- 由 `main.go` 在启动时加载

### **2. 数据消费**
- `event_consumer.go` 订阅 Redis Streams
- `cache_consumer.go` 定期读取 Redis 缓存
- `cache_manager.go` 统一管理内存状态

### **3. 报警评估**
- `evaluator.go` 是决策中枢
- 4 个事件处理器实现具体规则
- 使用 `alarm_event_builder.go` 构建结果

### **4. 数据持久化**
- 10 个 repository 文件处理不同数据对象
- 共享 `models/` 中的数据结构
- 支持数据库和云端同步

---

## 文件大小分布 (参考)

```
大型文件 (300+ 行):
  - evaluator.go
  - alarm.go
  
中型文件 (100-300 行):
  - event1_bed_fall.go
  - event3_bathroom_fall.go
  - cache_manager.go
  
小型文件 (< 100 行):
  - models/ (数据结构定义)
  - config.go (配置定义)
```

---

## 扩展建议

### **新增功能**
1. **新报警事件**: 在 `evaluator/` 中添加 `event5_*.go`
2. **新数据源**: 在 `consumer/` 中添加消费器
3. **新存储目标**: 在 `repository/` 中添加访问器

### **重构建议**
1. **Consumer 优化**: 合并相似消费逻辑到基础类
2. **Evaluator 优化**: 提取共同算法到 `evaluator.go`
3. **Repository 优化**: 抽象通用 DB 操作

---

## 总结表

| 层级 | 模块 | 文件数 | 职责 | 关键接口 |
|------|------|--------|------|---------|
| 1 | cmd | 1 | 应用入口 | main() |
| 2 | service | 2 | 业务编排 | AlarmService |
| 3 | evaluator | 9 | 规则引擎 | Evaluator |
| 4 | consumer | 5 | 数据消费 | Consumer |
| 5 | repository | 10 | 数据访问 | Repository |
| 6 | models | 4 | 数据结构 | - |
| 7 | config | 2 | 配置管理 | Config |

