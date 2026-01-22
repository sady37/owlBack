# 雷达订阅机制：1.0 vs 2.0 对比分析

## 1.0 版本实现（wisefido-backend）

### 核心文件
- `wisefido-backend/wisefido-radar/socket/connection.go`
- `wisefido-backend/wisefido-radar/socket/server.go`

### 实现机制

#### 1. 设备上线时自动订阅

**位置**: `server.go` 的 `RegisterDevice()` 函数（第244-251行）

```go
// Start real-time data subscription
go func() {
    if err := deviceConn.StartRealtimeDataSubscription(); err != nil {
        utils.Logger.Errorf("Failed to start real-time data subscription for device %s: %v", req.Uid, err)
    } else {
        utils.Logger.Infof("Started real-time data subscription for device %s", req.Uid)
    }
}()
```

**特点**:
- ✅ 设备注册后立即自动订阅
- ✅ 使用 goroutine 异步执行，不阻塞设备注册流程
- ✅ 订阅失败会记录日志，但不影响设备连接

#### 2. 自动续订机制

**位置**: `connection.go` 的 `autoRenewSubscription()` 函数（第326-382行）

```go
func (dc *DeviceConnection) autoRenewSubscription() {
    // Create a ticker that fires every 50 minutes (to renew before the 1-hour expiry)
    ticker := time.NewTicker(50 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Renew subscription
            dc.subscriptionMutex.Lock()
            if !dc.subscriptionActive {
                dc.subscriptionMutex.Unlock()
                return
            }
            dc.subscriptionMutex.Unlock()
            if err := dc.subscribeRealtimeData(); err != nil {
                utils.Logger.Errorf("Failed to renew real-time data subscription for device %s: %v", dc.UID, err)
                // 失败后每5分钟重试一次，直到成功
                for {
                    time.Sleep(5 * time.Minute)
                    // ... 重试逻辑
                }
            } else {
                utils.Logger.Infof("Successfully renewed real-time data subscription for device %s", dc.UID)
            }

        case <-dc.subscriptionCtx.Done():
            // Subscription has been canceled
            return
        }
    }
}
```

**特点**:
- ✅ **每50分钟自动续订一次**（订阅时长为3600秒，即1小时）
- ✅ 续订失败后，每5分钟重试一次，直到成功
- ✅ 使用 `subscriptionCtx` 控制 goroutine 生命周期
- ✅ 线程安全：使用 `subscriptionMutex` 保护状态

#### 3. 订阅实现

**位置**: `connection.go` 的 `subscribeRealtimeData()` 函数（第281-324行）

```go
func (dc *DeviceConnection) subscribeRealtimeData() error {
    // Create the subscription message
    subMsg := &pb.SetModeReq{
        Seq:     seq,
        Seconds: 3600, // Subscription duration (1 hour)
    }
    
    // Send the subscription message (type 26)
    if err := dc.WriteMessage(constant.MsgTypeSubscribe, subMsg); err != nil {
        return fmt.Errorf("failed to send subscription request: %v", err)
    }
    
    // Wait for response with timeout (30 seconds)
    // ...
}
```

**特点**:
- ✅ 订阅时长为**固定3600秒**（1小时）
- ✅ 使用 request-response 模式，等待设备确认
- ✅ 超时时间为30秒

#### 4. 订阅状态管理

**字段**:
```go
type DeviceConnection struct {
    subscriptionActive bool               // Whether subscription is active
    subscriptionCtx    context.Context    // Context for subscription
    subscriptionCancel context.CancelFunc // Cancel function for subscription
    subscriptionMutex  sync.Mutex         // For thread-safe subscription handling
}
```

**特点**:
- ✅ 使用 `subscriptionActive` 标志跟踪订阅状态
- ✅ 使用 `context` 控制订阅生命周期
- ✅ 线程安全的状态管理

### 1.0 版本总结

**优点**:
1. ✅ **设备上线时自动订阅** - 无需手动操作
2. ✅ **自动续订机制** - 每50分钟续订一次，确保数据不中断
3. ✅ **失败重试机制** - 续订失败后每5分钟重试
4. ✅ **状态管理完善** - 使用 context 和 mutex 管理订阅生命周期
5. ✅ **线程安全** - 使用 mutex 保护并发访问

**架构特点**:
- 基于 **Socket 连接**（TCP）
- 每个设备一个连接，订阅状态保存在连接对象中
- 使用 **protobuf** 协议通信

---

## 2.0 版本实现（owlBack/wisefido-radar）

### 核心文件
- `owlBack/wisefido-radar/internal/publisher/mqtt_publisher.go`
- `owlBack/wisefido-radar/internal/http/command_service.go`
- `owlBack/wisefido-radar/internal/http/command_handler.go`
- `owlBack/wisefido-radar/internal/consumer/mqtt_consumer.go`

### 实现机制

#### 1. 设备上线时自动订阅

**状态**: ❌ **未实现**

**位置**: `mqtt_consumer.go` 的 `handleMessage()` 函数
- 只接收消息，不主动订阅
- 设备首次连接时不会自动订阅

#### 2. 自动续订机制

**状态**: ❌ **未实现**

**问题**:
- 没有后台 goroutine 定期检查订阅状态
- 没有自动续订逻辑
- 订阅过期后数据会中断

#### 3. 订阅实现

**位置**: `command_service.go` 的 `SubscribeRealtimeData()` 函数

```go
func (s *CommandService) SubscribeRealtimeData(ctx context.Context, uid string, content interface{}, duration int) error {
    // 验证参数
    if duration < 1 || duration > 3600 {
        return fmt.Errorf("invalid duration: %d (must be between 1 and 3600)", duration)
    }
    
    // 发送订阅命令
    if err := s.publisher.PublishMonitorSubscriptionCommand(ctx, uid, content, duration); err != nil {
        return fmt.Errorf("failed to publish monitor subscription command: %w", err)
    }
    
    return nil
}
```

**特点**:
- ✅ 支持自定义订阅时长（1-3600秒）
- ✅ 支持选择订阅内容（0-同时订阅，1-轨迹，2-呼吸心率）
- ❌ 没有等待设备响应确认
- ❌ 没有记录订阅状态

#### 4. 订阅状态管理

**状态**: ❌ **未实现**

**问题**:
- 没有记录订阅状态（订阅时间、过期时间）
- 无法查询哪些设备正在订阅中
- 服务重启后无法恢复订阅状态

### 2.0 版本总结

**优点**:
1. ✅ 支持自定义订阅时长和内容类型
2. ✅ 基于 MQTT 协议，更灵活
3. ✅ HTTP API 接口，便于调用

**缺点**:
1. ❌ **缺少设备上线时自动订阅**
2. ❌ **缺少自动续订机制**
3. ❌ **缺少订阅状态管理**
4. ❌ **订阅过期后数据会中断**

**架构特点**:
- 基于 **MQTT** 协议
- 使用 HTTP API 调用订阅接口
- 订阅状态没有持久化

---

## 对比总结

| 特性 | 1.0 版本 | 2.0 版本 |
|------|---------|---------|
| **设备上线自动订阅** | ✅ 实现 | ❌ 未实现 |
| **自动续订机制** | ✅ 每50分钟续订 | ❌ 未实现 |
| **失败重试机制** | ✅ 每5分钟重试 | ❌ 未实现 |
| **订阅状态管理** | ✅ 连接对象中管理 | ❌ 未实现 |
| **订阅时长** | 固定3600秒 | 可配置1-3600秒 |
| **订阅内容选择** | 固定（同时订阅） | 可配置（0/1/2） |
| **协议** | Socket (TCP) + Protobuf | MQTT + JSON |
| **通信方式** | 长连接 | 发布/订阅 |

---

## 建议的改进方案（基于1.0版本经验）

### 方案：在2.0版本中实现类似1.0的自动续订机制

#### 1. 设备上线时自动订阅

**实现位置**: `mqtt_consumer.go` 的 `handleMessage()` 函数

```go
// 在设备首次连接时自动订阅
if isFirstConnection {
    go func() {
        if err := subscriptionManager.AutoSubscribe(uid); err != nil {
            logger.Error("Failed to auto-subscribe", zap.String("uid", uid), zap.Error(err))
        }
    }()
}
```

#### 2. 创建订阅管理器

**新建文件**: `wisefido-radar/internal/service/subscription_manager.go`

```go
type SubscriptionManager struct {
    config       *config.Config
    publisher    *publisher.MQTTPublisher
    deviceRepo   *repository.DeviceRepository
    redisClient  *redis.Client
    logger       *zap.Logger
}

// Start 启动订阅管理器（后台定时续订）
func (m *SubscriptionManager) Start(ctx context.Context) {
    ticker := time.NewTicker(50 * time.Minute) // 每50分钟检查一次
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkAndRenewSubscriptions(ctx)
        }
    }
}

// checkAndRenewSubscriptions 检查并续订即将过期的订阅
func (m *SubscriptionManager) checkAndRenewSubscriptions(ctx context.Context) {
    // 从 Redis 获取所有活跃订阅
    // 检查过期时间，在过期前10%时间续订
    // 续订失败时重试
}
```

#### 3. 使用 Redis 管理订阅状态

**Redis Key 设计**:
```
radar:subscription:{uid} -> {
    "content": 0,
    "duration": 3600,
    "subscribed_at": "2026-01-09T00:00:00Z",
    "expires_at": "2026-01-09T01:00:00Z"
}
TTL: 3600 秒
```

**优点**:
- 状态持久化，服务重启后可恢复
- 可以跨服务实例共享订阅状态
- 利用 Redis TTL 自动清理过期订阅

#### 4. 在 main.go 中启动订阅管理器

```go
subscriptionManager := service.NewSubscriptionManager(cfg, publisher, deviceRepo, redisClient, logger)
go subscriptionManager.Start(ctx)
```

---

## 结论

**1.0 版本的实现非常完善**，特别是：
- ✅ 设备上线时自动订阅
- ✅ 自动续订机制（每50分钟）
- ✅ 失败重试机制（每5分钟）
- ✅ 完善的状态管理

**2.0 版本需要借鉴1.0的经验**，实现类似的自动续订机制，但可以：
- 使用 Redis 管理订阅状态（而不是连接对象）
- 支持可配置的订阅时长和内容类型
- 保持 MQTT 协议的灵活性
