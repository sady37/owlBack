# 雷达订阅自动续订机制实现说明

## 概述

基于 1.0 版本（wisefido-backend）的实现经验，在 2.0 版本（owlBack/wisefido-radar）中实现了类似的自动订阅和自动续订机制。

## 实现功能

### 1. 设备上线时自动订阅 ✅

**实现位置**: `internal/consumer/mqtt_consumer.go`

- 当设备首次连接时（设备不存在或没有订阅记录），自动订阅实时数据
- 使用 goroutine 异步执行，不阻塞消息处理
- 可配置是否启用自动订阅（`RADAR_SUBSCRIPTION_AUTO`）

### 2. 自动续订机制 ✅

**实现位置**: `internal/service/subscription_manager.go`

- 后台 goroutine 定期检查订阅状态
- 默认每 50 分钟检查一次（可配置 `RADAR_SUBSCRIPTION_RENEWAL_INTERVAL`）
- 在订阅过期前 10 分钟自动续订（可配置 `RADAR_SUBSCRIPTION_RENEWAL_ADVANCE`）
- 续订失败会记录日志，下次检查时重试

### 3. 订阅状态管理 ✅

**实现位置**: Redis

- 订阅信息存储在 Redis 中，key 格式：`radar:subscription:{uid}`
- 包含订阅时间、过期时间、订阅内容、订阅时长等信息
- TTL 设置为订阅时长 + 10 分钟（作为缓冲）
- 支持服务重启后恢复订阅状态

### 4. 配置项

**环境变量**:

```bash
# 是否在设备上线时自动订阅（默认 true）
RADAR_SUBSCRIPTION_AUTO=true

# 默认订阅时长（秒，默认 3600，即1小时）
RADAR_SUBSCRIPTION_DURATION=3600

# 默认订阅内容：0-同时订阅，1-轨迹，2-呼吸心率（默认 0）
RADAR_SUBSCRIPTION_CONTENT=0

# 续订检查间隔（分钟，默认 50）
RADAR_SUBSCRIPTION_RENEWAL_INTERVAL=50

# 提前续订时间（分钟，默认 10）
RADAR_SUBSCRIPTION_RENEWAL_ADVANCE=10
```

## 架构设计

### 文件结构

```
wisefido-radar/
├── internal/
│   ├── service/
│   │   ├── subscription_manager.go  # 订阅管理器（新增）
│   │   └── radar.go                 # 雷达服务（修改：启动订阅管理器）
│   ├── consumer/
│   │   └── mqtt_consumer.go        # MQTT 消费者（修改：设备首次连接时自动订阅）
│   ├── http/
│   │   └── command_service.go      # 命令服务（修改：订阅时记录状态）
│   └── config/
│       └── config.go               # 配置（修改：添加订阅配置项）
```

### 数据流

```
设备首次连接
  ↓
MQTT Consumer 接收消息
  ↓
检测到设备首次连接
  ↓
调用 SubscriptionManager.AutoSubscribe()
  ↓
CommandService.SubscribeRealtimeData()
  ↓
MQTTPublisher.PublishMonitorSubscriptionCommand()
  ↓
发送订阅命令到设备
  ↓
保存订阅状态到 Redis
```

### 续订流程

```
SubscriptionManager.Start()
  ↓
每 50 分钟检查一次
  ↓
从 Redis 获取所有活跃订阅
  ↓
检查每个订阅的过期时间
  ↓
在过期前 10 分钟续订
  ↓
调用 CommandService.SubscribeRealtimeData()
  ↓
更新 Redis 中的订阅状态
```

## 与 1.0 版本的对比

| 特性 | 1.0 版本 | 2.0 版本 |
|------|---------|---------|
| **自动订阅** | ✅ Socket 连接时 | ✅ MQTT 消息时 |
| **自动续订** | ✅ 每 50 分钟 | ✅ 每 50 分钟（可配置） |
| **状态管理** | 连接对象中 | Redis 中（持久化） |
| **订阅时长** | 固定 3600 秒 | 可配置（默认 3600） |
| **订阅内容** | 固定（同时订阅） | 可配置（0/1/2） |
| **失败重试** | ✅ 每 5 分钟重试 | ⚠️ 下次检查时重试 |
| **协议** | Socket (TCP) | MQTT |

## 使用说明

### 启动服务

服务启动后，订阅管理器会自动启动：

```bash
./wisefido-radar
```

日志输出：
```
Subscription manager started
  renewal_interval_minutes=50
  subscription_duration=3600
  auto_subscribe=true
```

### 设备首次连接

当设备首次发送 MQTT 消息时，会自动订阅：

```
Device auto-created from device_store on MQTT connection
  device_id=xxx
  uid=xxx
  
Auto-subscribed on device first connection
  uid=xxx
```

### 自动续订

每 50 分钟自动检查并续订：

```
Renewing subscription
  uid=xxx
  expires_at=2026-01-09T01:00:00Z

Successfully renewed subscription
  uid=xxx

Renewed subscriptions
  count=5
```

## 注意事项

1. **Redis 依赖**: 订阅状态存储在 Redis 中，确保 Redis 服务正常运行
2. **配置调整**: 根据实际需求调整续订间隔和提前时间
3. **监控**: 建议监控订阅续订日志，确保续订成功
4. **服务重启**: 服务重启后，订阅管理器会从 Redis 恢复订阅状态并继续续订

## 后续优化建议

1. **失败重试机制**: 实现类似 1.0 版本的失败重试（每 5 分钟重试）
2. **订阅状态查询 API**: 提供 HTTP API 查询设备订阅状态
3. **订阅统计**: 记录订阅成功/失败次数，用于监控
4. **告警机制**: 订阅续订失败时发送告警
