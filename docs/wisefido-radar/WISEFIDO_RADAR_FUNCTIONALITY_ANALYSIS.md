# wisefido-radar 功能分析

## 一、wisefido-radar 的核心职责

`wisefido-radar` 是 Radar 设备与平台之间的**通信桥梁**，负责：

1. **设备认证**（HTTPS）
2. **数据接收**（MQTT 订阅）
3. **数据编码**（调用 `owl-common/encode.RadarEncode`）
4. **数据分类发布**（发布到 Redis Streams）
5. **命令下发**（接收 HTTP 请求，通过 MQTT 发送命令）

## 二、详细功能分析

### 2.1 HTTPS 认证服务

**功能**：设备启动时向平台请求认证，获取 MQTT 连接信息

**实现位置**：
- `internal/http/auth_handler.go` - HTTP 处理器
- `internal/http/auth_service.go` - 认证业务逻辑

**接口**：
- `POST /prod-api/thirdmqtt/v2/auth/device` - 标准 Radar 认证接口

**流程**：
1. 设备发送认证请求（包含 UID、硬件版本、固件版本等）
2. 服务验证设备合法性（查询 `device_store` 表）
3. 返回 MQTT 连接信息（clientId, server, port, account, pwd, protocol, prefix, productId）

**状态**：✅ 已实现

### 2.2 MQTT 消息订阅

**功能**：订阅设备发布的 6 类主题

**实现位置**：
- `internal/consumer/mqtt_consumer.go` - MQTT 消费者

**订阅的主题**：
1. `/prefix/prop/productId/UID/post` - 属性响应（配置/读取属性结果）
2. `/prefix/monitor/productId/UID/post` - 实时数据（轨迹、呼吸心率）
3. `/prefix/func/productId/UID/post` - 功能响应（重启、OTA 等结果）
4. `/prefix/stat/productId/UID/post` - 统计数据（轨迹统计、睡眠统计）
5. `/prefix/event/productId/UID/post` - 事件/日志（进出事件、姿态变化）
6. `/prefix/alarm/productId/UID/post` - 告警（跌倒、异常等）

**状态**：✅ 已实现

### 2.3 MQTT 消息处理

**功能**：处理接收到的 MQTT 消息

**实现位置**：
- `internal/consumer/mqtt_consumer.go` - `handleMessage()` 函数

**处理流程**：
1. **解析主题**：提取 `prefix`, `type`, `productId`, `UID`
2. **查询设备**：根据 `UID` 查询设备信息（如果不存在，从 `device_store` 自动创建）
3. **处理命令响应**（仅 `prop` 和 `func` 类型）：
   - 如果有 `requestId`，存储到 Redis（`radar:response:{requestId}`）
   - **不发布到 Streams**（命令响应是同步的 request-response 模式）
4. **构建数据**：
   - 添加元数据（`device_id`, `tenant_id`, `serial_number`, `uid`, `device_type`, `topic_type`, `timestamp`）
   - 直接展开原始数据到顶层（不保存在 `raw_data` 中）
5. **数据编码**：
   - 调用 `encode.RadarEncode(data, topicType)` 进行编码
   - 单位转换（dm→cm, m→cm, 10s→s）
   - Base64 解码（`bh`, `sleep`, `track` 字段）
   - SNOMED 映射（姿态、事件、睡眠状态等）
6. **发布到 Redis Streams**（仅数据类主题）：
   - 根据 `topic_type` 分类到不同的 Stream：
     - `monitor` → `iot:monitor:stream`
     - `stat` → `iot:stat:stream`
     - `event` → `iot:event:stream`
     - `alarm` → `iot:alarm:stream`

**状态**：✅ 已实现（但需要检查 `prop` 和 `func` 是否正确跳过 Streams 发布）

### 2.4 命令下发服务

**功能**：接收 HTTP 请求，通过 MQTT 向设备发送命令

**实现位置**：
- `internal/http/command_handler.go` - HTTP 处理器
- `internal/http/command_service.go` - 命令业务逻辑
- `internal/publisher/mqtt_publisher.go` - MQTT 发布器

**支持的命令**：
1. **属性配置**（`SetDeviceProperties`）：
   - 设置设备属性（如 `radar_install_height`, `rectangle`, `declare_area`, `fall_param` 等）
   - 主题：`/prefix/prop/productId/UID/get`
2. **属性读取**（`GetDeviceProperties`）：
   - 读取设备属性
   - 主题：`/prefix/prop/productId/UID/get`
3. **功能调用**（`CallDeviceFunction`）：
   - 重启设备、清除数据等
   - 主题：`/prefix/func/productId/UID/get`

**请求-响应机制**：
1. 生成 `requestId`
2. 通过 MQTT 发送命令（包含 `requestId`）
3. 等待设备响应（从 Redis 读取 `radar:response:{requestId}`）
4. 返回响应结果

**状态**：✅ 已实现

### 2.5 实时数据订阅管理

**功能**：订阅设备的实时数据（轨迹、呼吸心率）

**说明**：根据协议文档 3.7 节，设备默认不上报实时数据，需要服务器主动订阅。

**实现位置**：
- `internal/http/command_service.go` - 可以添加订阅管理功能

**订阅命令格式**：
```json
{
  "cmd": "subscription",
  "data": {
    "content": 0,  // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
    "duration": 30  // 订阅时长（秒），最大 3600
  }
}
```

**状态**：❓ 需要检查是否已实现

## 三、数据流分析

### 3.1 设备主动上报的数据（数据流）

```
Radar 设备
  ↓ MQTT 发布
wisefido-radar (MQTT Consumer)
  ↓ 解析、设备查询、编码
encode.RadarEncode()
  ↓ 单位转换、Base64 解码、SNOMED 映射
Redis Streams
  ├─ iot:monitor:stream  (实时数据)
  ├─ iot:stat:stream     (统计数据)
  ├─ iot:event:stream    (事件/日志)
  └─ iot:alarm:stream    (告警)
  ↓
下游服务（wisefido-iot-timeseries, wisefido-card-aggregator, wisefido-alarm）
```

### 3.2 平台主动下发的命令（命令流）

```
wisefido-data (HTTP API)
  ↓ HTTP 请求
wisefido-radar (HTTP Handler)
  ↓ 验证、生成 requestId
CommandService
  ↓ MQTT 发布
Radar 设备
  ↓ MQTT 发布响应
wisefido-radar (MQTT Consumer)
  ↓ 解析 requestId
Redis (radar:response:{requestId})
  ↓ 读取响应
CommandService
  ↓ 返回结果
wisefido-data (HTTP API 响应)
```

## 四、当前实现的问题分析

### 4.1 问题 1：prop 和 func 主题的消息处理

**问题**：在 `handleMessage()` 函数中，对于 `prop` 和 `func` 类型的消息：
1. 先存储到 Redis（用于 request-response）
2. 然后继续编码和发布到 Streams

**分析**：
- `prop` 和 `func` 是**命令响应**，属于同步的 request-response 模式
- 这些响应**不应该**发布到 Streams（它们不是数据流）
- 但是，代码中在存储到 Redis 后，仍然继续执行了编码和发布逻辑

**建议**：
- 在存储到 Redis 后，应该 `return nil`，跳过后续的编码和发布步骤
- 或者，将 `prop` 和 `func` 的处理逻辑分离出来

**状态**：⚠️ 需要修正

### 4.2 问题 2：实时数据订阅管理

**问题**：当前代码中是否有订阅实时数据的功能？

**分析**：
- 根据协议文档 3.7 节，设备默认不上报实时数据（`monitor` 主题）
- 需要服务器主动订阅，才能收到实时数据
- 订阅命令需要发送到 `/prefix/monitor/productId/UID/get` 主题

**建议**：
- 检查是否有订阅管理功能
- 如果没有，需要添加：
  - HTTP API：`POST /api/v1/radar/devices/{uid}/subscribe`
  - 参数：`content`（0/1/2）, `duration`（秒）

**状态**：❓ 需要检查

### 4.3 问题 3：monitor 主题的数据格式

**问题**：`monitor` 主题的数据可能包含 `track` 和 `bh` 字段，这些字段是 base64 编码的。

**分析**：
- `track` 字段：base64 编码的 16 字节 * N 数组（N 为人数）
- `bh` 字段：base64 编码的 16 字节数组（呼吸心率数据）
- 这些字段需要在 `encode.RadarEncode()` 中解码

**建议**：
- 确保 `encode.RadarEncode()` 正确处理 `monitor` 主题的数据
- 检查 `encodeRadarMonitor()` 是否正确处理 base64 解码

**状态**：✅ 应该已实现（在 `radar_encoder.go` 中）

### 4.4 问题 4：stat 主题的数据格式

**问题**：`stat` 主题的数据包含 `sleep` 和 `track` 字段，这些字段是 base64 编码的。

**分析**：
- `sleep` 字段：base64 编码的 16 字节数组（睡眠统计）
- `track` 字段：base64 编码的 16 字节数组（轨迹统计）
- 这些字段需要在 `encode.RadarEncode()` 中解码

**建议**：
- 确保 `encode.RadarEncode()` 正确处理 `stat` 主题的数据
- 检查 `encodeRadarStat()` 是否正确处理 base64 解码和位字段提取

**状态**：✅ 应该已实现（在 `radar_encoder.go` 中）

## 五、需要实现的功能清单

### 5.1 必须实现（核心功能）

- [x] HTTPS 认证服务
- [x] MQTT 消息订阅（6 个主题）
- [x] MQTT 消息处理（解析、设备查询、编码）
- [x] 数据编码（调用 `encode.RadarEncode`）
- [x] 数据分类发布（发布到 Redis Streams）
- [x] 命令下发（属性配置、属性读取、功能调用）
- [x] 请求-响应机制（Redis 存储响应）

### 5.2 需要完善（优化功能）

- [ ] **修正 prop 和 func 消息处理**：存储到 Redis 后应该跳过 Streams 发布
- [ ] **实时数据订阅管理**：添加订阅实时数据的 HTTP API
- [ ] **错误处理优化**：增强错误处理和日志记录
- [ ] **性能优化**：批量处理、异步处理等

### 5.3 可选功能（扩展功能）

- [ ] **OTA 升级管理**：完整的 OTA 升级流程
- [ ] **设备状态监控**：设备在线状态、连接质量等
- [ ] **数据统计**：消息计数、错误统计等

## 六、关键代码位置

### 6.1 核心文件

| 文件 | 功能 | 状态 |
|------|------|------|
| `cmd/wisefido-radar/main.go` | 服务入口 | ✅ |
| `internal/service/radar.go` | 服务编排 | ✅ |
| `internal/consumer/mqtt_consumer.go` | MQTT 消息处理 | ✅ 需优化 |
| `internal/http/auth_handler.go` | HTTPS 认证 | ✅ |
| `internal/http/command_handler.go` | 命令下发 | ✅ |
| `internal/http/command_service.go` | 命令业务逻辑 | ✅ |
| `internal/publisher/mqtt_publisher.go` | MQTT 发布 | ✅ |

### 6.2 关键函数

| 函数 | 功能 | 状态 |
|------|------|------|
| `handleMessage()` | 处理 MQTT 消息 | ✅ 需优化 |
| `getOutputStreamName()` | 获取输出 Stream | ✅ |
| `encode.RadarEncode()` | 数据编码 | ✅ |

## 七、总结

`wisefido-radar` 的核心功能**已基本实现**，但需要以下优化：

1. **修正 prop 和 func 消息处理逻辑**：存储到 Redis 后应该跳过后续的编码和发布步骤
2. **添加实时数据订阅管理**：提供 HTTP API 来订阅设备的实时数据
3. **验证数据编码**：确保所有类型的数据都能正确编码（特别是 base64 字段）

**核心流程**：
- ✅ 设备认证 → HTTPS 服务
- ✅ 数据接收 → MQTT 订阅 → 编码 → 发布到 Streams
- ✅ 命令下发 → HTTP API → MQTT 发布 → 等待响应
