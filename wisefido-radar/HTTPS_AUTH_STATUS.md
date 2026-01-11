# HTTPS 认证功能实现状态

## ✅ 功能已完整实现

### 1. 认证端点 ✅

**标准路径**（协议文档要求）:
```
POST https://服务器地址:8443/prod-api/thirdmqtt/v2/auth/device
```

**兼容路径**:
```
POST https://服务器地址:8443/radar/api/v1/auth
POST https://服务器地址:8443/
```

### 2. 实现文件 ✅

| 文件 | 功能 | 状态 |
|------|------|------|
| `internal/http/server.go` | HTTPS 服务器实现 | ✅ |
| `internal/http/auth_handler.go` | 认证请求处理器 | ✅ |
| `internal/http/auth_service.go` | 认证业务逻辑 | ✅ |
| `internal/http/router.go` | 路由注册 | ✅ |
| `internal/models/auth_request.go` | 认证请求模型 | ✅ |
| `internal/models/auth_response.go` | 认证响应模型 | ✅ |
| `internal/service/radar.go` | 服务启动集成 | ✅ |

### 3. 认证流程 ✅

```
设备启动 8 秒后
    ↓
发送 HTTPS POST 请求
    ↓
/prod-api/thirdmqtt/v2/auth/device
    ↓
AuthHandler.ServeHTTP()
    ↓
AuthService.AuthenticateDevice()
    ↓
验证设备合法性（device_store 表）
    ↓
生成 MQTT 配置
    ↓
返回认证响应（包含 MQTT 账号信息）
```

### 4. 认证请求格式 ✅

```json
{
  "uid": "设备序列号",
  "type": 1,
  "auth": "认证随机数（可忽略）",
  "salt": "认证随机数（可忽略）",
  "mcu": {
    "hw": "HC2-2.0",
    "sw": "20240101",
    "mac": "AA:BB:CC:DD:EE:FF",
    "iccid": "4G设备ICCID"
  },
  "radar": {
    "hw": "1.0",
    "sw": "20240101",
    "cap": "T"
  }
}
```

### 5. 认证响应格式 ✅

```json
{
  "msg": "操作成功",
  "code": 200,
  "data": {
    "uid": "设备序列号",
    "mqtt": {
      "clientId": "radar-设备序列号",
      "timeout": 30,
      "keepalive": 60,
      "server": "47.77.194.143",
      "port": 8883,
      "account": "wfiot",
      "pwd": "MQTT密码",
      "protocol": "2",
      "prefix": "",
      "productId": "88"
    }
  }
}
```

### 6. 配置项 ✅

**环境变量**:
- `RADAR_HTTPS_PORT`: HTTPS 端口（默认 443，启动脚本设置为 8443）
- `RADAR_HTTPS_CERT_FILE`: 证书文件路径（默认 `server.crt`）
- `RADAR_HTTPS_KEY_FILE`: 私钥文件路径（默认 `server.key`）

**MQTT 配置**（返回给设备）:
- `RADAR_MQTT_SERVER`: MQTT 服务器地址（默认 `47.77.194.143`）
- `RADAR_MQTT_PORT`: MQTT 端口（默认 `8883`）
- `RADAR_MQTT_ACCOUNT`: MQTT 账号（默认 `wfiot`）
- `RADAR_MQTT_PASSWORD`: MQTT 密码
- `RADAR_MQTT_PROTOCOL`: 协议类型（默认 `2`=加密）

### 7. 设备验证逻辑 ✅

1. 从 `device_store` 表查询设备（通过 `uid` 或 `serial_number`）
2. 检查设备是否存在
3. 检查设备是否允许访问（`allow_access = true`）
4. 检查设备是否已分配给租户（`tenant_id != '00000000-0000-0000-0000-000000000000'`）
5. 验证通过后生成 MQTT 配置并返回

## ⚠️ 待改进功能

### 1. 存储设备认证信息 ⏳

**当前状态**: 认证时接收了设备版本/MAC等信息，但未存储到数据库

**需要实现**:
- 在认证成功时，将 `mcu.hw`, `mcu.sw`, `mcu.mac`, `mcu.iccid`, `radar.hw`, `radar.sw`, `radar.cap` 存储到数据库
- 建议存储在 `devices` 表的 `auth_info` JSONB 字段中

### 2. GetProperties 过滤不可变属性 ⏳

**当前状态**: GetProperties 返回所有属性，包括设备版本/MAC等

**需要实现**:
- 在 `GetDeviceProperties` 返回结果时，过滤掉已在认证时存储的不可变属性
- 从数据库读取已存储的认证信息，从设备返回的属性中移除这些字段

**需要过滤的属性**:
- `mcu.hw`, `mcu.sw`, `mcu.mac`, `mcu.iccid`
- `radar.hw`, `radar.sw`, `radar.cap`

## 📋 测试方法

### 1. 测试认证端点

```bash
# 使用 curl 测试认证
curl -k -X POST https://47.77.194.143:8443/prod-api/thirdmqtt/v2/auth/device \
  -H "Content-Type: application/json" \
  -d '{
    "uid": "test-device-001",
    "type": 1,
    "mcu": {
      "hw": "HC2-2.0",
      "sw": "20240101",
      "mac": "AA:BB:CC:DD:EE:FF"
    },
    "radar": {
      "hw": "1.0",
      "sw": "20240101",
      "cap": "T"
    }
  }'
```

### 2. 验证数据库

```sql
-- 检查设备是否在 device_store 中
SELECT * FROM device_store WHERE uid = 'test-device-001';

-- 检查认证信息是否存储（待实现）
SELECT device_id, uid, auth_info FROM devices WHERE uid = 'test-device-001';
```

### 3. 检查服务日志

```bash
# 查看认证日志
tail -f /tmp/owl_radar_startup.log | grep -i "auth\|认证"
```

## ✅ 总结

**HTTPS 认证功能已完整实现**，包括：
- ✅ HTTPS 服务器
- ✅ 认证端点（标准路径 `/prod-api/thirdmqtt/v2/auth/device`）
- ✅ 设备验证逻辑
- ✅ MQTT 配置生成
- ✅ 认证响应格式

**待实现功能**：
- ⏳ 存储设备认证信息（版本/MAC等）
- ⏳ GetProperties 过滤不可变属性

**设备连接方式**:
```
https://47.77.194.143:8443/prod-api/thirdmqtt/v2/auth/device
```
