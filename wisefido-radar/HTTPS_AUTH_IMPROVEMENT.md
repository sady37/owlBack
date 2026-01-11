# HTTPS 认证改进方案

## 📋 当前状态

### 1. HTTPS 端口配置 ✅

**设备应连接的端口**: `8443` (HTTPS)

**配置位置**:
- 启动脚本: `start-radar.sh` 中设置 `RADAR_HTTPS_PORT=8443`
- 配置文件: `internal/config/config.go` 中默认值 `443`，但启动脚本覆盖为 `8443`

**认证端点**:
- 标准路径: `POST /prod-api/thirdmqtt/v2/auth/device`
- 兼容路径: `POST /radar/api/v1/auth`
- 根路径: `POST /`

**完整URL示例**:
```
https://47.77.194.143:8443/prod-api/thirdmqtt/v2/auth/device
```

### 2. 当前认证流程 ✅

1. 设备在启动 8 秒后发起 HTTPS 认证请求
2. 认证请求包含:
   - `uid`: 设备序列号
   - `type`: 设备类型
   - `mcu`: 主控信息（HW, SW, MAC, ICCID）
   - `radar`: 雷达信息（HW, SW, Cap）
3. 服务器验证设备是否在 `device_store` 中存在
4. 验证通过后返回 MQTT 连接配置

### 3. 当前问题 ⚠️

1. **未存储设备版本/MAC等信息**: 认证时虽然接收了这些信息，但没有存储到数据库
2. **GetProperties 未过滤不可变属性**: 后续获取属性时，仍然会返回设备版本/MAC等信息

## 🎯 改进方案

### 改进 1: 存储设备认证信息

**目标**: 在认证时存储设备的不可变属性（版本、MAC等）

**实现位置**: `internal/http/auth_service.go`

**需要存储的信息**:
- MCU 硬件版本 (`mcu.hw`)
- MCU 软件版本 (`mcu.sw`)
- MCU MAC 地址 (`mcu.mac`)
- MCU ICCID (`mcu.iccid`, 4G设备)
- Radar 硬件版本 (`radar.hw`)
- Radar 软件版本 (`radar.sw`)
- Radar 功能类别 (`radar.cap`)

**存储位置**: 
- 方案A: 存储在 `devices` 表的扩展字段（JSONB）
- 方案B: 创建新表 `device_auth_info` 存储认证信息
- 方案C: 存储在 `device_store` 表的扩展字段

**推荐方案**: 方案A，因为认证信息属于设备属性的一部分

### 改进 2: GetProperties 过滤不可变属性

**目标**: 在 `GetDeviceProperties` 返回结果时，过滤掉已在认证时存储的不可变属性

**实现位置**: `internal/http/command_service.go`

**需要过滤的属性**:
- `mcu.hw`, `mcu.sw`, `mcu.mac`, `mcu.iccid`
- `radar.hw`, `radar.sw`, `radar.cap`

**实现逻辑**:
1. 从数据库读取设备的认证信息
2. 从设备返回的属性中移除这些不可变属性
3. 返回过滤后的属性列表

## 📝 实现步骤

### 步骤 1: 数据库迁移

在 `devices` 表中添加字段存储认证信息：

```sql
ALTER TABLE devices ADD COLUMN IF NOT EXISTS auth_info JSONB;
```

### 步骤 2: 修改认证服务

在 `AuthenticateDevice` 方法中，认证成功后存储设备信息：

```go
// 存储设备认证信息
authInfo := map[string]interface{}{
    "mcu": map[string]interface{}{
        "hw":    req.MCU.HW,
        "sw":    req.MCU.SW,
        "mac":   req.MCU.MAC,
        "iccid": req.MCU.ICCID,
    },
    "radar": map[string]interface{}{
        "hw":  req.Radar.HW,
        "sw":  req.Radar.SW,
        "cap": req.Radar.Cap,
    },
    "auth_time": time.Now().Format(time.RFC3339),
}

// 更新或创建设备记录，存储认证信息
err = s.saveDeviceAuthInfo(ctx, uid, authInfo)
```

### 步骤 3: 修改 GetProperties

在 `GetDeviceProperties` 方法中，过滤不可变属性：

```go
// 从数据库读取已存储的认证信息
authInfo, err := s.getDeviceAuthInfo(ctx, uid)
if err == nil && authInfo != nil {
    // 过滤掉已存储的不可变属性
    filteredData := filterImmutableProperties(response.Data, authInfo)
    return filteredData, nil
}
```

## 🔍 验证方法

### 1. 认证测试

```bash
# 模拟设备认证请求
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

### 2. 验证数据库存储

```sql
SELECT device_id, uid, auth_info 
FROM devices 
WHERE uid = 'test-device-001';
```

### 3. 验证 GetProperties 过滤

```bash
# 调用 GetProperties
curl -X POST http://localhost:8080/internal/api/v1/radar/devices/test-device-001/properties/get \
  -H "Content-Type: application/json" \
  -d '{"keys": []}'

# 验证返回结果中不包含 mcu.hw, mcu.sw, mcu.mac, radar.hw, radar.sw, radar.cap
```

## 📌 注意事项

1. **认证时机**: 设备在启动 8 秒后进行认证，这是设备固件的行为，服务器端无需特殊处理
2. **认证频率**: 认证通过后，除非设备重启，否则不再进行第二次认证（设备端控制）
3. **向后兼容**: 如果设备没有进行认证，GetProperties 仍然返回所有属性（包括不可变属性）
4. **数据一致性**: 如果设备固件升级，版本信息会变化，需要在认证时更新存储的信息

## 🚀 下一步行动

1. ✅ 确认 HTTPS 端口为 `8443`
2. ⏳ 实现认证信息存储功能
3. ⏳ 实现 GetProperties 过滤功能
4. ⏳ 添加单元测试
5. ⏳ 更新文档
