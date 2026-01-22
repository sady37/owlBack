# wisefido-radar 阶段 2 修改总结

## 一、修正的问题

### 1.1 prop 和 func 消息处理逻辑 ✅

**问题**：
- `prop` 和 `func` 类型的消息（命令响应）在存储到 Redis 后，仍然继续执行编码和发布到 Streams
- 这些消息是同步的 request-response 模式，不应该发布到 Streams

**修正**：
- 在 `handleMessage()` 函数中，对于 `prop` 和 `func` 类型的消息：
  - 存储到 Redis 后，直接 `return nil`，跳过后续的编码和发布步骤
  - 添加注释说明这些消息的处理逻辑

**代码位置**：`internal/consumer/mqtt_consumer.go` 第 186-221 行

**修正后的逻辑**：
```go
// 5. 处理命令响应（如果是属性响应或功能响应，需要存储到 Redis 供 CommandService 使用）
// 注意：prop 和 func 类型的消息是命令响应，属于同步的 request-response 模式
// 这些响应不应该发布到 Streams，只需要存储到 Redis 供 CommandService 读取即可
if topicInfo.Type == mqtt.TopicTypeProp {
    // 属性响应：存储到 Redis
    // ...
    return nil  // 直接返回，不发布到 Streams
} else if topicInfo.Type == mqtt.TopicTypeFunc {
    // 功能响应：存储到 Redis
    // ...
    return nil  // 直接返回，不发布到 Streams
}

// 6. 处理数据上报消息（monitor, stat, event, alarm）
// 这些消息需要编码后发布到 Redis Streams
// ...
```

**状态**：✅ 已修正

### 1.2 导入循环问题 ✅

**问题**：
- `internal/http/command_handler.go` 使用了 `service.CommandService`
- 但 `CommandService` 实际在 `http` 包中定义
- 导致导入循环：`service` → `http` → `service`

**修正**：
- 将 `command_handler.go` 中的 `service.CommandService` 改为 `CommandService`（同包）
- 将 `service/radar.go` 中的 `NewCommandService` 改为 `http.NewCommandService`

**代码位置**：
- `internal/http/command_handler.go` 第 14 行
- `internal/service/radar.go` 第 58 行

**状态**：✅ 已修正

### 1.3 重复函数声明 ✅

**问题**：
- `endsWith` 函数在 `router.go` 和 `command_handler.go` 中都有定义

**修正**：
- 删除 `command_handler.go` 中的 `endsWith` 函数（使用 `router.go` 中的定义）

**状态**：✅ 已修正

## 二、实时数据订阅管理功能检查

### 2.1 功能实现状态 ✅

**检查结果**：实时数据订阅管理功能**已完整实现**

**实现位置**：
1. **HTTP API 处理器**：`internal/http/command_handler.go`
   - `SubscribeRealtime()` - 处理 HTTP 请求
   - 路由：`POST /internal/api/v1/radar/devices/{uid}/realtime/subscribe`

2. **命令服务**：`internal/http/command_service.go`
   - `SubscribeRealtimeData()` - 业务逻辑
   - 验证设备存在性
   - 验证参数（duration: 1-3600 秒）
   - 调用发布器发送订阅命令

3. **MQTT 发布器**：`internal/publisher/mqtt_publisher.go`
   - `PublishMonitorSubscriptionCommand()` - 发送订阅命令
   - 主题：`/prefix/monitor/productId/UID/get`
   - 消息格式：`{"cmd":"subscription","data":{"content":0,"duration":30}}`

4. **路由注册**：`internal/http/router.go`
   - 已注册 `/internal/api/v1/radar/devices/` 路径处理

**状态**：✅ 已完整实现

### 2.2 API 接口

**接口**：`POST /internal/api/v1/radar/devices/{uid}/realtime/subscribe`

**请求体**：
```json
{
  "content": 0,   // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
  "duration": 30  // 订阅时长（秒），最大 3600
}
```

**响应**：
```json
{
  "code": 200,
  "msg": "操作成功",
  "data": null
}
```

### 2.3 功能流程

```
HTTP 请求
  ↓
CommandHandler.SubscribeRealtime()
  ↓ 解析 UID、content、duration
CommandService.SubscribeRealtimeData()
  ↓ 验证设备、参数
MQTTPublisher.PublishMonitorSubscriptionCommand()
  ↓ 构建命令消息
MQTT 发布到 /prefix/monitor/productId/UID/get
  ↓
Radar 设备接收订阅命令
  ↓ 开始上报实时数据
MQTT 发布到 /prefix/monitor/productId/UID/post
  ↓
wisefido-radar (MQTT Consumer)
  ↓ 编码、发布到 Streams
iot:monitor:stream
```

## 三、wisefido-radar 完整功能清单

### 3.1 数据上报（设备 → 平台）

| 功能 | MQTT Topic | 处理逻辑 | Redis Stream | 状态 |
|------|-----------|---------|-------------|------|
| 属性响应 | `/prefix/prop/productId/UID/post` | 存储到 Redis（request-response） | ❌ 不发布 | ✅ |
| 实时数据 | `/prefix/monitor/productId/UID/post` | 编码后发布 | `iot:monitor:stream` | ✅ |
| 功能响应 | `/prefix/func/productId/UID/post` | 存储到 Redis（request-response） | ❌ 不发布 | ✅ |
| 统计数据 | `/prefix/stat/productId/UID/post` | 编码后发布 | `iot:stat:stream` | ✅ |
| 事件数据 | `/prefix/event/productId/UID/post` | 编码后发布 | `iot:event:stream` | ✅ |
| 告警数据 | `/prefix/alarm/productId/UID/post` | 编码后发布 | `iot:alarm:stream` | ✅ |

### 3.2 命令下发（平台 → 设备）

| 功能 | HTTP API | MQTT Topic | 状态 |
|------|---------|-----------|------|
| 设备认证 | `POST /prod-api/thirdmqtt/v2/auth/device` | - | ✅ |
| 属性读取 | `POST /internal/api/v1/radar/devices/{uid}/properties/get` | `/prefix/prop/productId/UID/get` | ✅ |
| 属性设置 | `POST /internal/api/v1/radar/devices/{uid}/properties/set` | `/prefix/prop/productId/UID/get` | ✅ |
| 实时数据订阅 | `POST /internal/api/v1/radar/devices/{uid}/realtime/subscribe` | `/prefix/monitor/productId/UID/get` | ✅ |
| 功能调用 | `POST /internal/api/v1/radar/devices/{uid}/commands` | `/prefix/func/productId/UID/get` | ✅ |

### 3.3 数据编码

| 功能 | 实现位置 | 状态 |
|------|---------|------|
| 单位转换 | `owl-common/encode.RadarEncode()` | ✅ |
| Base64 解码 | `encodeRadarMonitor()`, `encodeRadarStat()` | ✅ |
| SNOMED 映射 | `applyRadarSNOMedMapping()` | ✅ |

## 四、数据流总结

### 4.1 数据上报流（设备 → 平台）

```
Radar 设备
  ↓ MQTT 发布
wisefido-radar (MQTT Consumer)
  ↓ 解析主题、查询设备
  ↓ 判断消息类型
  ├─ prop/func (命令响应)
  │   └─ 存储到 Redis (radar:response:{requestId})
  │   └─ return nil (不发布到 Streams) ✅ 已修正
  └─ monitor/stat/event/alarm (数据上报)
      ↓ 构建数据（添加元数据）
      ↓ 调用 encode.RadarEncode()
      │   ├─ 单位转换（dm→cm, m→cm, 10s→s）
      │   ├─ Base64 解码（bh, sleep, track）
      │   └─ SNOMED 映射
      ↓ 根据 topic_type 确定 Stream
      ↓ 发布到 Redis Streams
      ├─ iot:monitor:stream  (实时数据)
      ├─ iot:stat:stream     (统计数据)
      ├─ iot:event:stream    (事件/日志)
      └─ iot:alarm:stream    (告警)
```

### 4.2 命令下发流（平台 → 设备）

```
wisefido-data (HTTP API)
  ↓ HTTP 请求
wisefido-radar (HTTP Handler)
  ↓ 验证、生成 requestId
CommandService
  ↓ MQTT 发布命令
Radar 设备
  ↓ MQTT 发布响应
wisefido-radar (MQTT Consumer)
  ↓ 解析 requestId
  ↓ 存储到 Redis (radar:response:{requestId})
  ↓ return nil (不发布到 Streams) ✅ 已修正
CommandService
  ↓ 从 Redis 读取响应
  ↓ 返回结果
wisefido-data (HTTP API 响应)
```

### 4.3 实时数据订阅流（平台 → 设备）

```
wisefido-data (HTTP API)
  ↓ POST /internal/api/v1/radar/devices/{uid}/realtime/subscribe
wisefido-radar (CommandHandler)
  ↓ 解析参数（content, duration）
CommandService.SubscribeRealtimeData()
  ↓ 验证设备、参数
MQTTPublisher.PublishMonitorSubscriptionCommand()
  ↓ MQTT 发布订阅命令
Radar 设备
  ↓ 接收订阅命令，开始上报实时数据
MQTT 发布到 /prefix/monitor/productId/UID/post
  ↓
wisefido-radar (MQTT Consumer)
  ↓ 编码、发布到 Streams
iot:monitor:stream
```

## 五、修正前后对比

### 5.1 prop 和 func 消息处理

**修正前**：
```
prop/func 消息
  ↓
存储到 Redis
  ↓ (继续执行)
编码
  ↓ (继续执行)
发布到 Streams ❌ (不应该)
```

**修正后**：
```
prop/func 消息
  ↓
存储到 Redis
  ↓
return nil ✅ (直接返回，不发布到 Streams)
```

### 5.2 代码质量改进

1. ✅ **修正逻辑错误**：prop 和 func 消息不再发布到 Streams
2. ✅ **修正导入循环**：统一使用 `http.CommandService`
3. ✅ **消除重复代码**：删除重复的 `endsWith` 函数
4. ✅ **完善注释**：添加详细注释说明消息处理逻辑

## 六、验证结果

### 6.1 编译检查 ✅

```bash
cd wisefido-radar && go build ./...
# 结果：编译成功，无错误
```

### 6.2 功能完整性 ✅

| 功能模块 | 状态 |
|---------|------|
| HTTPS 认证服务 | ✅ |
| MQTT 消息订阅 | ✅ |
| MQTT 消息处理 | ✅ 已修正 |
| 数据编码 | ✅ |
| 数据分类发布 | ✅ 已修正 |
| 命令下发 | ✅ |
| 实时数据订阅 | ✅ 已确认 |
| 请求-响应机制 | ✅ 已修正 |

## 七、总结

### 7.1 修正完成 ✅

1. ✅ **prop 和 func 消息处理逻辑**：存储到 Redis 后直接返回，不发布到 Streams
2. ✅ **导入循环问题**：修正 `CommandService` 的引用
3. ✅ **重复函数声明**：删除重复的 `endsWith` 函数

### 7.2 功能确认 ✅

1. ✅ **实时数据订阅管理功能**：已完整实现
   - HTTP API：`POST /internal/api/v1/radar/devices/{uid}/realtime/subscribe`
   - 命令服务：`SubscribeRealtimeData()`
   - MQTT 发布器：`PublishMonitorSubscriptionCommand()`
   - 路由注册：已注册

### 7.3 代码质量 ✅

- ✅ 编译通过
- ✅ 无 linter 错误
- ✅ 逻辑正确
- ✅ 注释完善

**wisefido-radar 阶段 2 修改完成！** 🎉
