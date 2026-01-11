# 雷达订阅自动续订机制移植总结

## ✅ 移植完成状态

### 1. **订阅管理器** ✅
- **文件**: `internal/service/subscription_manager.go`
- **状态**: 已从 `exports/` 目录移植到 `internal/service/`
- **功能**:
  - 自动订阅设备实时数据
  - 自动续订即将过期的订阅（每50分钟检查一次）
  - 订阅状态管理（存储在 Redis 中）

### 2. **配置项** ✅
- **文件**: `internal/config/config.go`
- **状态**: 已添加订阅相关配置
- **配置项**:
  - `RADAR_SUBSCRIPTION_AUTO` (默认: true) - 是否自动订阅
  - `RADAR_SUBSCRIPTION_DURATION` (默认: 3600) - 订阅时长（秒）
  - `RADAR_SUBSCRIPTION_CONTENT` (默认: 0) - 订阅内容（0=同时订阅，1=轨迹，2=呼吸心率）
  - `RADAR_SUBSCRIPTION_RENEWAL_INTERVAL` (默认: 50) - 续订检查间隔（分钟）
  - `RADAR_SUBSCRIPTION_RENEWAL_ADVANCE` (默认: 10) - 提前续订时间（分钟）

### 3. **MQTT 消费者** ✅
- **文件**: `internal/consumer/mqtt_consumer.go`
- **状态**: 已添加自动订阅逻辑
- **功能**:
  - 设备首次连接时自动订阅
  - 通过检查 Redis 订阅状态判断是否首次连接
  - 使用 goroutine 异步执行，不阻塞消息处理

### 4. **服务启动** ✅
- **文件**: `internal/service/radar.go`
- **状态**: 已添加订阅管理器启动逻辑
- **功能**:
  - 创建订阅管理器
  - 设置订阅管理器到 MQTT 消费者
  - 启动订阅管理器（后台自动续订）
  - 停止订阅管理器（优雅关闭）

### 5. **命令服务** ✅
- **文件**: `internal/http/command_service.go`
- **状态**: 已添加订阅状态保存逻辑
- **功能**:
  - 手动订阅时也保存订阅状态到 Redis
  - 使用与订阅管理器兼容的格式

## 📊 实现对比

| 特性 | 1.0 版本 (wisefido-backend) | 2.0 版本 (owlBack) | 状态 |
|------|---------------------------|-------------------|------|
| **自动订阅** | ✅ Socket 连接时 | ✅ MQTT 消息时 | ✅ 已实现 |
| **自动续订** | ✅ 每 50 分钟 | ✅ 每 50 分钟（可配置） | ✅ 已实现 |
| **状态管理** | 连接对象中 | Redis 中（持久化） | ✅ 已实现 |
| **订阅时长** | 固定 3600 秒 | 可配置（默认 3600） | ✅ 已实现 |
| **订阅内容** | 固定（同时订阅） | 可配置（0/1/2） | ✅ 已实现 |
| **失败重试** | ✅ 每 5 分钟重试 | ⚠️ 下次检查时重试 | ⚠️ 待优化 |
| **协议** | Socket (TCP) | MQTT | ✅ 已适配 |

## 🔧 使用说明

### 环境变量配置

```bash
# 是否在设备上线时自动订阅（默认 true）
export RADAR_SUBSCRIPTION_AUTO=true

# 默认订阅时长（秒，默认 3600，即1小时）
export RADAR_SUBSCRIPTION_DURATION=3600

# 默认订阅内容：0-同时订阅，1-轨迹，2-呼吸心率（默认 0）
export RADAR_SUBSCRIPTION_CONTENT=0

# 续订检查间隔（分钟，默认 50）
export RADAR_SUBSCRIPTION_RENEWAL_INTERVAL=50

# 提前续订时间（分钟，默认 10）
export RADAR_SUBSCRIPTION_RENEWAL_ADVANCE=10
```

### 启动服务

服务启动后，订阅管理器会自动启动：

```bash
cd /home/wisefido/owl-project/owlBack/wisefido-radar
./start-radar.sh
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

## ⚠️ 注意事项

1. **Redis 依赖**: 订阅状态存储在 Redis 中，确保 Redis 服务正常运行
2. **配置调整**: 根据实际需求调整续订间隔和提前时间
3. **监控**: 建议监控订阅续订日志，确保续订成功
4. **服务重启**: 服务重启后，订阅管理器会从 Redis 恢复订阅状态并继续续订

## 🔄 数据流

### 自动订阅流程

```
设备首次连接
  ↓
MQTT Consumer 接收消息
  ↓
检测到设备首次连接（设备不存在或没有订阅记录）
  ↓
调用 SubscriptionManager.AutoSubscribe()
  ↓
CommandService.SubscribeRealtimeData()
  ↓
MQTTPublisher.PublishMonitorSubscriptionCommand()
  ↓
发送订阅命令到设备
  ↓
保存订阅状态到 Redis (radar:subscription:{uid})
```

### 自动续订流程

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

## 📝 后续优化建议

1. **失败重试机制**: 实现类似 1.0 版本的失败重试（每 5 分钟重试）
2. **订阅状态查询 API**: 提供 HTTP API 查询设备订阅状态
3. **订阅统计**: 记录订阅成功/失败次数，用于监控
4. **告警机制**: 订阅续订失败时发送告警
5. **格式统一**: 考虑将 SubscriptionInfo 移到 models 包，统一格式

## ✅ 验证结果

- ✅ 代码编译成功
- ✅ 所有文件已正确移植
- ✅ 配置项已添加
- ✅ 自动订阅逻辑已实现
- ✅ 自动续订逻辑已实现
- ✅ 状态管理已实现

## 🎯 总结

**移植工作已完成！** 所有功能已从 1.0 版本成功移植到 2.0 版本，并适配了 MQTT 协议。系统现在支持：

- ✅ 设备上线时自动订阅
- ✅ 订阅自动续订（每50分钟检查一次）
- ✅ 订阅状态持久化（Redis）
- ✅ 可配置的订阅参数

**可以开始测试和验证！** 🚀