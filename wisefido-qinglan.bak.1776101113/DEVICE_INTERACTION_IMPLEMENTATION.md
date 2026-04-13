# 设备交互式通讯实现说明

## 概述

已完成与雷达设备之间的交互式通讯功能，支持通过 MQTT 协议主动向设备发送命令并接收响应。

## 已实现的功能

### 1. 设备属性操作（/prefix/prop/productId/UID/get）

#### 读取设备属性
- **API**: `GET /api/v1/radar/devices/{uid}/properties?keys=key1,key2`
- **实现**: `RadarService.GetDeviceProperties()`
- **命令格式**:
  ```json
  {
    "cmd": "read",
    "requestId": "prop_{uid}_{timestamp}",
    "data": {
      "key": ["key1", "key2"]
    }
  }
  ```
- **响应处理**: 设备在 `/prefix/prop/productId/UID/post` 主题返回响应，`handlePropertyMessage()` 提取 `requestId` 并存储到 Redis

#### 设置设备属性
- **API**: `PUT /api/v1/radar/devices/{uid}/properties`
- **实现**: `RadarService.SetDeviceProperties()`
- **命令格式**:
  ```json
  {
    "cmd": "update",
    "requestId": "prop_set_{uid}_{timestamp}",
    "data": {
      "property1": "value1",
      "property2": "value2"
    }
  }
  ```

### 2. 订阅实时数据（/prefix/monitor/productId/UID/get）

- **API**: `POST /api/v1/radar/devices/{uid}/subscribe`
- **实现**: `RadarService.SubscribeRealtimeData()`
- **命令格式**:
  ```json
  {
    "cmd": "subscription",
    "data": {
      "content": "0",  // 字符串格式：0-同时订阅，1-轨迹，2-呼吸心率
      "duration": 3600  // 订阅时长（秒），最大3600
    }
  }
  ```
- **注意**: 
  - monitor 订阅不需要 `requestId`，因为设备会持续推送数据
  - 设备会在 `/prefix/monitor/productId/UID/post` 主题持续推送实时数据
  - 数据由 `handleMonitorMessage()` 处理并发布到 Redis Stream

### 3. 调用设备功能（/prefix/func/productId/UID/get）

- **API**: `POST /api/v1/radar/devices/{uid}/function`
- **实现**: `RadarService.CallDeviceFunction()`
- **命令格式**:
  ```json
  {
    "cmd": "control",
    "requestId": "func_{uid}_{timestamp}",
    "data": {
      "dev": 0  // 0-重启雷达和主控，1-只重启雷达，2-只重启主控，100-清除设备数据，101-清除雷达数据，102-清除主控数据
    }
  }
  ```
- **响应处理**: 设备在 `/prefix/func/productId/UID/post` 主题返回响应，`handleFunctionMessage()` 提取 `requestId` 并存储到 Redis

## 响应处理机制

### 响应存储流程

1. **设备发送响应** → `/prefix/{type}/productId/UID/post` 主题
2. **MQTT Consumer 接收** → `handleMessage()` 路由到对应的处理器
3. **提取 requestId** → `handlePropertyMessage()` 或 `handleFunctionMessage()` 提取响应中的 `requestId`
4. **存储到 Redis** → `StreamPublisher.StoreCommandResponse()` 存储响应（TTL: 5分钟）
5. **等待响应** → `RadarService.waitForResponse()` 轮询 Redis 获取响应（100ms 间隔）
6. **返回结果** → 提取响应数据并返回给 API 调用者

### 响应格式

设备返回的响应格式：
```json
{
  "requestId": "prop_xxx_1234567890",
  "code": 200,
  "msg": "success",
  "data": {
    // 响应数据
  }
}
```

## 已订阅的设备发布主题

以下6个主题已在 `mqtt_consumer.go` 中全部订阅：

1. `/prefix/prop/productId/UID/post` - 属性操作响应
2. `/prefix/monitor/productId/UID/post` - 实时数据推送
3. `/prefix/func/productId/UID/post` - 功能调用响应
4. `/prefix/stat/productId/UID/post` - 统计数据上报
5. `/prefix/event/productId/UID/post` - 事件上报
6. `/prefix/alarm/productId/UID/post` - 告警上报

## 关键实现细节

### 1. 请求ID生成
- 使用纳秒级时间戳确保唯一性：`fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())`
- 格式：`prop_{uid}_{timestamp}`, `prop_set_{uid}_{timestamp}`, `func_{uid}_{timestamp}`

### 2. 响应匹配
- 支持多种字段名：`requestId`, `request_id`, `requestID`
- 从设备响应中提取 `requestId` 并存储到 Redis key: `cmd:response:{requestId}`

### 3. 超时处理
- 属性操作：10秒超时
- 功能调用：10秒超时（重启操作30秒）
- 使用 `context.WithTimeout()` 实现超时控制

### 4. 错误处理
- 检查响应中的 `code` 字段，非200表示设备返回错误
- Redis key 不存在时继续等待（`redis.Nil`）
- 超时或错误时返回明确的错误信息

## 使用示例

### 读取设备属性
```bash
curl -X GET "http://localhost:8080/api/v1/radar/devices/BM87224700978/properties?keys=radarFunctionMode,radarInstallStyle"
```

### 设置设备属性
```bash
curl -X PUT "http://localhost:8080/api/v1/radar/devices/BM87224700978/properties" \
  -H "Content-Type: application/json" \
  -d '{
    "properties": {
      "radarFunctionMode": 1,
      "radarInstallStyle": 2
    }
  }'
```

### 订阅实时数据
```bash
curl -X POST "http://localhost:8080/api/v1/radar/devices/BM87224700978/subscribe" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "0",
    "duration": 3600
  }'
```

### 调用设备功能（重启）
```bash
curl -X POST "http://localhost:8080/api/v1/radar/devices/BM87224700978/function" \
  -H "Content-Type: application/json" \
  -d '{
    "dev": 0
  }'
```

## 文件修改清单

1. **`internal/consumer/mqtt_consumer.go`**
   - 完善 `handlePropertyMessage()`: 提取 `requestId` 并存储响应
   - 完善 `handleFunctionMessage()`: 提取 `requestId` 并存储响应

2. **`internal/service/radar_service.go`**
   - 修复 `GetDeviceProperties()`: 使用 `cmd: "read"` 格式
   - 修复 `SetDeviceProperties()`: 使用 `cmd: "update"` 格式
   - 修复 `SubscribeRealtimeData()`: 使用 `cmd: "subscription"` 格式，发送到 `monitor` 主题
   - 修复 `CallDeviceFunction()`: 使用 `cmd: "control"` 格式
   - 完善 `waitForResponse()`: 改进错误处理和响应验证

## 注意事项

1. **monitor 订阅**：不需要等待响应，设备会持续推送数据到 `/monitor/.../post` 主题
2. **content 字段**：必须使用字符串格式（`"0"` 而不是 `0`），兼容当前固件
3. **requestId 匹配**：设备返回的响应必须包含与请求相同的 `requestId`，否则无法匹配
4. **超时时间**：重启操作需要更长的超时时间（30秒），其他操作10秒
5. **Redis TTL**：响应存储在 Redis 中，TTL 为5分钟，超时后自动清理

## 测试建议

1. 测试属性读取：发送读取命令，验证响应是否正确返回
2. 测试属性设置：发送设置命令，验证设备是否成功应用设置
3. 测试实时数据订阅：发送订阅命令，验证设备是否持续推送数据
4. 测试功能调用：发送重启命令，验证设备是否执行并返回响应
5. 测试超时处理：断开设备连接，验证超时错误是否正确返回
6. 测试错误处理：发送无效命令，验证错误响应是否正确处理
