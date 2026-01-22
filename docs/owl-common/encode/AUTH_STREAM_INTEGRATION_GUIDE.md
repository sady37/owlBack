# Redis Auth Stream 集成指南

## 概述

本文档说明如何在 `wisefido-radar` 服务中集成 Redis Auth Stream 功能，按照 `RADAR_REDIS_STREAM_FORMAT_STANDARD.md` 中定义的标准格式生成和发布设备认证事件。

## 架构设计

### 数据流程

```
设备认证请求
    ↓
auth_handler 接收 HTTP 请求
    ↓
auth_service 处理认证逻辑
    ├→ 发布 auth_request 到 Redis Stream
    ├→ 验证设备
    └→ 发布 auth_response 到 Redis Stream
    ↓
设备接收 MQTT 配置
```

### 核心组件

1. **auth_encoder.go** (`owl-common/encode/`)
   - 提供认证事件编码函数
   - `EncodeAuthRequest()`: 编码认证请求
   - `EncodeAuthResponse()`: 编码认证响应
   - `BuildAuthRequestFromHTTPRequest()`: 从 HTTP 请求构建设备信息
   - `ValidateAuthStreamEvent()`: 验证事件有效性

2. **auth_service.go** (`wisefido-radar/internal/http/`)
   - 调用 encoder 生成事件对象
   - 发布到 Redis Stream
   - 处理认证逻辑

3. **auth_handler.go** (`wisefido-radar/internal/http/`)
   - 接收 HTTP 请求
   - 传递 remoteAddr 给 auth_service

## Redis Stream 格式

### 认证请求（auth_request）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "00000000-0000-0000-0000-000000000000",
  "timestamp": 1234567890,
  "topic_type": "auth",
  "data_value": "{\"category\": \"auth_request\", \"device_uid\": \"E598A2ACD523\", \"remote_addr\": \"10.0.0.187:57087\", \"log\": {\"device_type\": 1, \"mcu_hw\": \"2.0\", \"mcu_sw\": \"Dec 17 2025 10:22:19\", \"radar_hw\": \"2.3\", \"radar_sw\": \"Jun 25 2025 11:33:44\"}}",
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": null,
  "bed_id": null
}
```

### 认证响应成功（auth_response）

```json
{
  "device_id": "uuid-yyy",
  "device_type": "Radar",
  "tenant_id": "bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c",
  "timestamp": 1234567890,
  "topic_type": "auth",
  "data_value": "{\"category\": \"auth_response\", \"auth_status\": \"success\", \"device_uid\": \"E598A2ACD523\", \"mqtt_server\": \"10.0.0.100\", \"mqtt_port\": 8883, \"log\": \"Device authenticated successfully, MQTT server: 10.0.0.100:8883\"}",
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": null,
  "bed_id": null
}
```

### 认证响应失败（auth_response）

```json
{
  "device_id": "",
  "device_type": "Radar",
  "tenant_id": "00000000-0000-0000-0000-000000000000",
  "timestamp": 1234567890,
  "topic_type": "auth",
  "data_value": "{\"category\": \"auth_response\", \"auth_status\": \"failure\", \"device_uid\": \"E598A2ACD523\", \"log\": \"device not found in device_store\"}",
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": null,
  "bed_id": null
}
```

## 使用示例

### 方式 1：使用 encoder 包直接生成

```go
import "owl-common/encode"

// 构建设备信息
deviceInfo := encode.BuildAuthRequestFromHTTPRequest(
    "E598A2ACD523",           // uid
    1,                        // device_type
    "2.0",                    // mcu_hw
    "Dec 17 2025 10:22:19",   // mcu_sw
    "2.3",                    // radar_hw
    "Jun 25 2025 11:33:44",   // radar_sw
    "10.0.0.187:57087",       // remoteAddr
)

// 编码为 Redis Stream 事件
authRequest := encode.EncodeAuthRequest(
    deviceID,           // 系统内 UUID
    "E598A2ACD523",     // 设备序列号
    "Radar",            // 设备类型
    "10.0.0.187:57087", // 远程地址
    deviceInfo,         // 设备信息
)

// 验证
if err := encode.ValidateAuthStreamEvent(authRequest); err != nil {
    log.Fatal(err)
}
```

### 方式 2：通过 auth_service 自动发布（推荐）

```go
// auth_service 在处理认证时自动发布到 Redis Stream
authService.AuthenticateDevice(ctx, authRequest, remoteAddr)
// ↓ 自动发布 auth_request 和 auth_response 到 iot:auth:stream
```

## 集成检查清单

- [x] 创建 `auth_encoder.go` 实现编码逻辑
- [x] 在 `auth_service.go` 添加 Redis 客户端依赖
- [x] 在 `auth_service.go` 添加认证事件发布方法
- [x] 更新 `auth_handler.go` 传递 remoteAddr
- [x] 创建示例代码

## 待办事项

1. **依赖检查**: 确保 `owl-common/encode` 和 `owl-common/redis/common` 包可用
2. **配置检查**: 确保 Redis 客户端在 auth_service 初始化时正确传入
3. **单元测试**: 为 auth_encoder 和认证事件发布添加测试
4. **集成测试**: 测试认证流程的 Redis 发布功能
5. **错误处理**: 完善 Redis 连接失败等异常情况的处理

## 数据验证规则

所有 auth stream 事件必须满足：

1. **必需字段**:
   - `device_id`: 不能为空（失败时可为空字符串）
   - `device_type`: 必须为 "Radar"
   - `tenant_id`: 不能为空（请求时为默认UUID，响应时为实际租户ID）
   - `topic_type`: 必须为 "auth"
   - `data_value.category`: 必须为 "auth_request" 或 "auth_response"
   - `data_value.device_uid`: 不能为空

2. **条件字段**:
   - `mqtt_server` 和 `mqtt_port`: 仅在 `auth_status="success"` 时出现
   - `remote_addr`: 仅在 `auth_request` 中出现

## 监控和日志

auth_service 在以下时刻产生日志：

1. 认证请求接收 (INFO)
2. 认证请求发布 (INFO)
3. 设备验证失败 (WARN)
4. 认证失败响应发布 (INFO)
5. 设备认证成功 (INFO)
6. 认证成功响应发布 (INFO)
7. Redis 发布失败 (ERROR)

示例日志：

```
2025-01-14T10:00:00.000Z    info    Device authentication request    uid=E598A2ACD523    type=1    mcu_hw=2.0    radar_hw=2.3    remote_addr=10.0.0.187:57087
2025-01-14T10:00:00.001Z    info    Published auth request to Redis Stream    uid=E598A2ACD523    stream=iot:auth:stream    stream_id=1234567890000-0
2025-01-14T10:00:00.050Z    info    Device authenticated successfully    uid=E598A2ACD523    tenant_id=bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c    mqtt_server=10.0.0.100    mqtt_port=8883
2025-01-14T10:00:00.051Z    info    Published auth response (success) to Redis Stream    uid=E598A2ACD523    tenant_id=bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c    stream=iot:auth:stream    stream_id=1234567890001-0
```

## Redis CLI 命令参考

查看 auth stream 中的所有事件：

```bash
# 查看最新 100 条事件
XREVRANGE iot:auth:stream + - COUNT 100

# 查看特定时间范围的事件
XRANGE iot:auth:stream 1642000000000 1642000100000

# 订阅新事件
XREAD STREAMS iot:auth:stream $

# 获取 stream 信息
XINFO STREAM iot:auth:stream
```

## 性能考虑

- auth_service 在发布失败时仅记录日志，不阻塞认证流程
- Redis 发布是异步操作，不影响 HTTP 响应延迟
- 建议为 Redis Stream 配置合适的 MAXLEN 参数防止无限增长

## 参考文档

- [RADAR_REDIS_STREAM_FORMAT_STANDARD.md](./RADAR_REDIS_STREAM_FORMAT_STANDARD.md) - Redis Stream 标准格式定义
- [auth_request.go](../wisefido-radar/internal/models/auth_request.go) - 认证请求数据模型
- [auth_response.go](../wisefido-radar/internal/models/auth_response.go) - 认证响应数据模型
- [auth_encoder.go](./auth_encoder.go) - 认证编码器实现
