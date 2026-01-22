# wisefido-radar 功能需求分析

## 一、核心职责

`wisefido-radar` 是 Radar 设备与平台之间的**通信桥梁**，负责处理**双向通信**：

1. **设备 → 平台**（数据上报）：接收设备 MQTT 消息，编码后发布到 Redis Streams
2. **平台 → 设备**（命令下发）：接收 HTTP 请求，通过 MQTT 向设备发送命令

## 二、功能模块

### 2.1 HTTPS 认证服务 ✅

**功能**：设备启动时向平台请求认证，获取 MQTT 连接信息

**接口**：
- `POST /prod-api/thirdmqtt/v2/auth/device` - 标准 Radar 认证接口

**流程**：
1. 设备发送认证请求（包含 UID、硬件版本、固件版本等）
2. 服务验证设备合法性（查询 `device_store` 表）
3. 返回 MQTT 连接信息（clientId, server, port, account, pwd, protocol, prefix, productId）

**状态**：✅ 已实现

### 2.2 MQTT 消息订阅 ✅

**功能**：订阅设备发布的 6 类主题

**订阅的主题**：
1. `/prefix/prop/productId/UID/post` - 属性响应（配置/读取属性结果）
2. `/prefix/monitor/productId/UID/post` - 实时数据（轨迹、呼吸心率）
3. `/prefix/func/productId/UID/post` - 功能响应（重启、OTA 等结果）
4. `/prefix/stat/productId/UID/post` - 统计数据（轨迹统计、睡眠统计）
5. `/prefix/event/productId/UID/post` - 事件/日志（进出事件、姿态变化）
6. `/prefix/alarm/productId/UID/post` - 告警（跌倒、异常等）

**状态**：✅ 已实现

### 2.3 MQTT 消息处理 ⚠️ 需优化

**功能**：处理接收到的 MQTT 消息

**处理流程**：
1. **解析主题**：提取 `prefix`, `type`, `productId`, `UID`
2. **查询设备**：根据 `UID` 查询设备信息（如果不存在，从 `device_store` 自动创建）
3. **区分消息类型**：
   - **命令响应**（`prop`, `func`）：存储到 Redis，**不发布到 Streams**
   - **数据上报**（`monitor`, `stat`, `event`, `alarm`）：编码后发布到 Streams
4. **数据编码**（仅数据类消息）：
   - 调用 `encode.RadarEncode(data, topicType)` 进行编码
   - 单位转换（dm→cm, m→cm, 10s→s）
   - Base64 解码（`bh`, `sleep`, `track` 字段）
   - SNOMED 映射（姿态、事件、睡眠状态等）
5. **发布到 Redis Streams**（仅数据类消息）：
   - `monitor` → `iot:monitor:stream`
   - `stat` → `iot:stat:stream`
   - `event` → `iot:event:stream`
   - `alarm` → `iot:alarm:stream`

**问题**：⚠️ 当前实现中，`prop` 和 `func` 消息在存储到 Redis 后，仍然继续执行了编码和发布步骤，这是不对的。

**状态**：✅ 已实现，但需要修正逻辑

### 2.4 命令下发服务 ✅

**功能**：接收 HTTP 请求，通过 MQTT 向设备发送命令

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

### 2.5 实时数据订阅管理 ❓

**功能**：订阅设备的实时数据（轨迹、呼吸心率）

**说明**：根据协议文档 3.7 节，设备默认不上报实时数据，需要服务器主动订阅。

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

**主题**：`/prefix/monitor/productId/UID/get`

**HTTP API 需求**：
- `POST /api/v1/radar/devices/{uid}/subscribe` - 订阅实时数据
- 参数：`content`（0/1/2）, `duration`（秒）

**状态**：❓ 需要检查是否已实现

## 三、数据流分析

### 3.1 数据上报流（设备 → 平台）

```
Radar 设备
  ↓ MQTT 发布
wisefido-radar (MQTT Consumer)
  ↓ 解析主题、查询设备
  ↓ 判断消息类型
  ├─ prop/func (命令响应)
  │   └─ 存储到 Redis (radar:response:{requestId})
  │   └─ 结束（不发布到 Streams）
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

### 3.2 命令下发流（平台 → 设备）

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
  ↓ 结束（不发布到 Streams）
CommandService
  ↓ 从 Redis 读取响应
  ↓ 返回结果
wisefido-data (HTTP API 响应)
```

## 四、需要修正的问题

### 4.1 问题 1：prop 和 func 消息处理逻辑 ⚠️

**问题**：在 `handleMessage()` 函数中，对于 `prop` 和 `func` 类型的消息：
1. ✅ 先存储到 Redis（用于 request-response）
2. ❌ 然后继续执行编码和发布到 Streams（不应该）

**分析**：
- `prop` 和 `func` 是**命令响应**，属于同步的 request-response 模式
- 这些响应**不应该**发布到 Streams（它们不是数据流）
- 它们只需要存储到 Redis，供 `CommandService` 读取即可

**修正方案**：
- 在存储到 Redis 后，应该 `return nil`，跳过后续的编码和发布步骤

**代码位置**：`internal/consumer/mqtt_consumer.go` 第 186-221 行

**状态**：⚠️ 需要修正

### 4.2 问题 2：实时数据订阅管理 ❓

**问题**：是否有订阅实时数据的功能？

**分析**：
- 根据协议文档 3.7 节，设备默认不上报实时数据（`monitor` 主题）
- 需要服务器主动订阅，才能收到实时数据
- 订阅命令需要发送到 `/prefix/monitor/productId/UID/get` 主题

**建议**：
- 检查是否有订阅管理功能
- 如果没有，需要添加 HTTP API

**状态**：❓ 需要检查

## 五、需要实现的功能清单

### 5.1 必须修正（核心问题）

- [ ] **修正 prop 和 func 消息处理逻辑**：
  - 存储到 Redis 后应该 `return nil`，跳过后续的编码和发布步骤
  - 代码位置：`internal/consumer/mqtt_consumer.go`

### 5.2 需要检查（功能完整性）

- [ ] **实时数据订阅管理**：
  - 检查是否有订阅实时数据的 HTTP API
  - 如果没有，需要添加 `POST /api/v1/radar/devices/{uid}/subscribe`

### 5.3 需要验证（功能正确性）

- [ ] **数据编码验证**：
  - 验证 `monitor` 主题的 base64 解码是否正确（`track`, `bh` 字段）
  - 验证 `stat` 主题的 base64 解码是否正确（`sleep`, `track` 字段）
  - 验证 SNOMED 映射是否正确应用

- [ ] **Stream 分类验证**：
  - 验证数据是否正确分类到对应的 Stream（`iot:monitor:stream`, `iot:stat:stream`, `iot:event:stream`, `iot:alarm:stream`）

## 六、总结

### 6.1 已实现的功能 ✅

1. ✅ HTTPS 认证服务
2. ✅ MQTT 消息订阅（6 个主题）
3. ✅ MQTT 消息处理（解析、设备查询）
4. ✅ 数据编码（调用 `encode.RadarEncode`）
5. ✅ 数据分类发布（发布到 Redis Streams）
6. ✅ 命令下发（属性配置、属性读取、功能调用）
7. ✅ 请求-响应机制（Redis 存储响应）

### 6.2 需要修正的问题 ⚠️

1. ⚠️ **prop 和 func 消息处理逻辑**：存储到 Redis 后应该跳过 Streams 发布

### 6.3 需要检查的功能 ❓

1. ❓ **实时数据订阅管理**：检查是否有订阅实时数据的 HTTP API

### 6.4 核心流程

**数据上报**：
```
设备 MQTT 发布 → 解析主题 → 查询设备 → 判断类型
├─ prop/func: 存储到 Redis → 结束
└─ monitor/stat/event/alarm: 编码 → 发布到 Streams
```

**命令下发**：
```
HTTP 请求 → 生成 requestId → MQTT 发布命令 → 等待响应 → 返回结果
```

## 七、下一步行动

1. **修正 prop 和 func 消息处理逻辑**（优先级：高）
2. **检查实时数据订阅管理功能**（优先级：中）
3. **验证数据编码功能**（优先级：中）
4. **测试完整流程**（优先级：高）
