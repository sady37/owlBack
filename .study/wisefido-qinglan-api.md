# wisefido-qinglan API 接口文档

> Radar 设备网关服务 HTTP API 完整参考手册
>
> 基于源码分析生成，版本：2026-04-02

## 服务信息

- **默认端口**: 8081
- **Base URL**: `http://127.0.0.1:8081`
- **协议**: HTTP（无认证中间件，认证在独立 HTTPS Server 处理）
- **路由库**: gorilla/mux

## 通用响应格式

除 `/health` 端点外，所有接口均使用统一响应格式：

```json
{
  "success": true,
  "data": {},
  "error": "",
  "count": 0
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 操作是否成功 |
| data | object/array | 成功时返回的数据 |
| error | string | 失败时返回的错误信息 |
| count | number | 列表接口返回的记录数 |

---

## API 端点列表

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/radar/devices/status` | 批量查询设备在线状态 |
| GET | `/api/v1/radar/devices/{uid}/properties` | 读取设备属性 |
| PUT | `/api/v1/radar/devices/{uid}/properties` | 写入设备属性 |
| POST | `/api/v1/radar/devices/{uid}/subscribe` | 启动实时数据订阅 |
| POST | `/api/v1/radar/devices/{uid}/function` | 调用设备功能（重启/清数据） |
| GET | `/api/v1/radar/devices/{uid}/info` | 查询设备元数据 |
| GET | `/api/v1/radar/devices/{uid}/status` | 查询单设备在线状态 |
| GET | `/api/v1/tenants/{tenantId}/devices` | 查询租户下所有设备 |

---

## 1. 健康检查

### GET /health

系统健康检查端点，不使用标准响应格式。

**请求参数**: 无

**响应示例**:
```json
{
  "status": "healthy",
  "service": "wisefido-qinglan",
  "time": "2026-04-02T10:30:00+08:00"
}
```

---

## 2. 读取设备属性

### GET /api/v1/radar/devices/{uid}/properties

通过 MQTT 读取设备的配置属性。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | 设备物理标识（device_uid） |

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keys | string | 否 | 逗号分隔的属性名，不传则读取全部属性 |

**行为说明**:
- 每个 key 独立发起一次 MQTT 请求-响应，单次超时 10s
- 多个 key 之间间隔 50ms
- 上下文总超时 30s

**请求示例**:
```bash
GET /api/v1/radar/devices/BM87XXXX/properties?keys=install_height,heart_breath_switch
```

**成功响应**:
```json
{
  "success": true,
  "data": {
    "install_height": "280",
    "install_model": "0",
    "heart_breath_switch": "1",
    "rectangle": "{\"x1\":10,\"y1\":20,\"x2\":500,\"y2\":600}"
  }
}
```

**错误响应** (500):
```json
{
  "success": false,
  "error": "Failed to get device properties: device timeout"
}
```

---

## 3. 写入设备属性

### PUT /api/v1/radar/devices/{uid}/properties

通过 MQTT 写入设备的配置属性，支持批量写入。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | 设备物理标识 |

**请求体**:
```json
{
  "properties": {
    "install_height": 280,
    "install_model": 0,
    "rectangle": "{\"x1\":10,\"y1\":20,\"x2\":500,\"y2\":600}",
    "declare_area": [
      {
        "id": 1,
        "type": "0",
        "x1": 10, "y1": 20,
        "x2": 30, "y2": 40,
        "x3": 50, "y3": 60,
        "x4": 70, "y4": 80
      }
    ]
  }
}
```

**属性分组发送规则**:

服务会自动将属性分组，按顺序发送（每组间隔 100ms）：

| 分组 | 属性 | 说明 |
|------|------|------|
| 1 | `radar_func_ctrl`, `radar_install_style`, `radar_install_height` | 工作模式组（一起发送） |
| 2 | `heart_breath_param` | 心率呼吸参数（单独） |
| 3 | `fall_param` | 跌倒参数（单独） |
| 4 | `rectangle` | 检测矩形（单独） |
| 5 | `declare_area` | 区域声明（多区域自动拆分，每个间隔 300ms） |
| 6 | `_alarm_items_json` | 特殊字段，解析为 AlarmItem 数组后编码 |
| 其他 | 其他未知属性 | 每个单独发送 |

**特殊字段**:
- `_alarm_items_json` (string): 传入 AlarmItem 数组的 JSON，内部自动编码为 `fall_param` 和 `heart_breath_param` 的 base64 值
- `declare_area`: 如果包含多个区域（检测到 `},{`），会自动拆分为多条指令，每条间隔 300ms（设备固件限制）

**单位转换**:
- 前端传入的长度单位为 cm，内部自动转换为设备单位 dm（10cm）
- 如 `install_height: 280` (cm) → 设备收到 `radar_install_height: "28"` (dm)

**成功响应**:
```json
{
  "success": true,
  "message": "Device properties set successfully",
  "device_code": 200
}
```

**设备返回非 200 时**（仍为 HTTP 200）:
```json
{
  "success": false,
  "error": "device returned code 777: device offline",
  "device_code": 777
}
```

**设备响应码**:
| Code | 含义 |
|------|------|
| 200 | 成功 |
| 500 | 设备侧处理失败 |
| 777 | 设备离线 |
| 778 | 设备不支持该模式 |

**错误响应** (400):
```json
{
  "success": false,
  "error": "Properties cannot be empty"
}
```

---

## 4. 启动实时数据订阅

### POST /api/v1/radar/devices/{uid}/subscribe

向设备发送 monitor 订阅指令，设备会在指定时长内持续推送实时数据。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | 设备物理标识 |

**请求体**:
```json
{
  "content": 0,
  "duration": 300
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | number | 是 | 订阅内容：`0`=轨迹+体征，`1`=仅轨迹，`2`=仅体征 |
| duration | number | 是 | 订阅时长（秒），范围 1~3600 |

**行为说明**:
- Fire-and-forget 模式，不等待设备响应
- 设备会在 duration 秒内持续向 MQTT `.../monitor/.../post` topic 推送数据
- MQTT 消费者会将数据写入 `iot:monitor:stream`

**成功响应**:
```json
{
  "success": true,
  "message": "Realtime data subscription started",
  "data": {
    "uid": "BM87XXXX",
    "duration": 300
  }
}
```

**错误响应** (400):
```json
{
  "success": false,
  "error": "Duration must be between 1 and 3600 seconds"
}
```

---

## 5. 调用设备功能（重启/清数据）

### POST /api/v1/radar/devices/{uid}/function

发送设备功能调用指令，支持重启和清除数据。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | 设备物理标识 |

**请求体**:
```json
{
  "dev": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| dev | number | 是 | 功能码，仅允许 0/1/2 |

**功能码定义**:
| dev | 功能 | 超时 |
|-----|------|------|
| 0 | 重启雷达 + MCU | 30s |
| 1 | 仅重启雷达 | 30s |
| 2 | 仅重启 MCU | 30s |

**注意**: HTTP 层校验仅允许 0/1/2，其他值（如 100/101/102 清数据）会被拒绝。

**成功响应**:
```json
{
  "success": true,
  "message": "Device function called successfully"
}
```

**错误响应** (400):
```json
{
  "success": false,
  "error": "Dev must be 0, 1, or 2"
}
```

---

## 6. 查询单设备在线状态

### GET /api/v1/radar/devices/{uid}/status

从内存的 DeviceStatusManager 读取设备在线状态（无 MQTT/DB 调用）。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | 设备物理标识 |

**成功响应**:
```json
{
  "success": true,
  "data": {
    "device_uid": "BM87XXXX",
    "status": "online"
  }
}
```

**状态值说明**:
| status | 含义 |
|--------|------|
| online | 设备在线，正常通信 |
| offline | 设备离线，超过 180s 无消息 |
| unsubscribed | 设备未订阅，需重新认证 |

**错误响应** (500):
```json
{
  "success": false,
  "error": "Device status manager not available"
}
```

---

## 7. 批量查询设备在线状态

### GET /api/v1/radar/devices/status

批量查询多个设备的在线状态，可按租户过滤。

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tenant_id | string | 否 | 租户 UUID，不传返回全部设备 |

**成功响应**:
```json
{
  "success": true,
  "data": [
    {
      "device_uid": "BM87XXXX",
      "device_id": "550e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "tenant-uuid",
      "status": "online"
    },
    {
      "device_uid": "DEF9XXXX",
      "device_id": "660e8400-e29b-41d4-a716-446655440001",
      "tenant_id": "tenant-uuid",
      "status": "offline"
    }
  ],
  "count": 2
}
```

**注意**: 此路由必须在 `/{uid}/status` 之前注册，避免 `"status"` 被误匹配为 `{uid}`。

---

## 8. 查询设备元数据

### GET /api/v1/radar/devices/{uid}/info

从 PostgreSQL 数据库查询设备的元数据信息。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | 设备物理标识 |

**成功响应**:
```json
{
  "success": true,
  "data": {
    "DeviceID": "550e8400-e29b-41d4-a716-446655440000",
    "DeviceUID": "BM87XXXX",
    "TenantID": "tenant-uuid",
    "DeviceName": "Room 301 Radar",
    "BoundRoomID": {"String": "room-uuid", "Valid": true},
    "BoundBedID": {"String": "", "Valid": false},
    "Status": "online",
    "BusinessAccess": "approved",
    "MonitoringEnabled": true,
    "DeviceType": {"String": "radar", "Valid": true},
    "DeviceModel": {"String": "QL-T200", "Valid": true},
    "IMEI": {"String": "860012345678901", "Valid": true},
    "CommMode": {"String": "4G", "Valid": true},
    "MCUModel": {"String": "1.0", "Valid": true},
    "FirmwareVersion": {"String": "2024-01-15", "Valid": true}
  }
}
```

**字段说明**:
- 响应直接序列化 Go 的 `domain.Device` 结构体
- 由于未定义 `json` tag，字段名为 PascalCase
- `sql.NullString` 类型字段序列化为 `{"String":"value", "Valid":true/false}` 对象

**错误响应** (404):
```json
{
  "success": false,
  "error": "Device not found: no rows in result set"
}
```

---

## 9. 查询租户下所有设备

### GET /api/v1/tenants/{tenantId}/devices

查询指定租户下的所有设备列表。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tenantId | string | 是 | 租户 UUID |

**成功响应**:
```json
{
  "success": true,
  "data": [
    {
      "DeviceID": "550e8400-e29b-41d4-a716-446655440000",
      "DeviceUID": "BM87XXXX",
      "TenantID": "tenant-uuid",
      "DeviceName": "Room 301 Radar",
      "Status": "online",
      "BusinessAccess": "approved",
      "MonitoringEnabled": true
    }
  ],
  "count": 1
}
```

**错误响应** (500):
```json
{
  "success": false,
  "error": "Failed to get devices: database connection error"
}
```

---

## 附录 A：常用属性键名

### 安装配置

| 属性名 | 类型 | 单位 | 说明 |
|--------|------|------|------|
| radar_install_height | string | dm | 安装高度（设备单位为 dm，前端传 cm 自动转换） |
| radar_install_style | string | - | 安装模式：0=水平，1=壁挂，2=吸顶 |
| rectangle | string | cm | 检测矩形 JSON：`{"x1":10,"y1":20,"x2":500,"y2":600}` |
| declare_area | array | cm | 区域声明数组，每个区域为四边形坐标 |

### 功能开关

| 属性名 | 类型 | 说明 |
|--------|------|------|
| radar_func_ctrl | string | 工作模式（0~7 位掩码） |
| heart_breath_switch | string | 心率呼吸监测开关：0=关，1=开 |

### 告警参数

| 属性名 | 类型 | 说明 |
|--------|------|------|
| fall_param | string | 跌倒检测参数（base64 编码 16 字节） |
| heart_breath_param | string | 心率呼吸告警阈值（base64 编码 16 字节） |

---

## 附录 B：请求-响应流程

### 属性读写流程

```
Client ──HTTP GET/PUT──> qinglan:8081
                            │
                            ├─> MQTT publish to /{prefix}/prop/{productID}/{uid}/get
                            │   (topic 举例: /prop/88/BM87XXXX/get)
                            │
                            ├─> Device 处理并响应 to /prop/88/BM87XXXX/post
                            │
                            ├─> mqtt_consumer 接收响应，提取 requestId
                            │
                            ├─> 存入 Redis key: cmd:response:{requestId} (TTL 5min)
                            │
                            ├─> RadarService.waitForResponse() 轮询 Redis
                            │   (100ms 间隔，最多 10s 或 30s 超时)
                            │
                            └─> HTTP response 返回给 Client
```

### 订阅流程

```
Client ──HTTP POST /subscribe──> qinglan:8081
                                    │
                                    ├─> MQTT publish to /monitor/88/{uid}/get
                                    │   payload: {"cmd":"subscription","data":{"content":"0","duration":3600}}
                                    │
                                    └─> HTTP 200 立即返回（不等设备响应）

Device ──持续推送 monitor 数据──> /monitor/88/{uid}/post
                                    │
                                    ├─> mqtt_consumer 接收
                                    │
                                    └─> XADD to iot:monitor:stream
```

---

## 附录 C：错误码汇总

### HTTP 错误码

| HTTP Status | 场景 |
|-------------|------|
| 200 | 成功，或设备返回错误但 HTTP 层通信成功 |
| 400 | 请求参数错误、JSON 解析失败、参数范围不合法 |
| 404 | 设备不存在（仅 GetDeviceInfo 端点） |
| 500 | MQTT 发布失败、设备响应超时、内部服务不可用 |

### 设备响应码（device_code）

| Code | 含义 | 出现场景 |
|------|------|----------|
| 200 | 成功 | 设备成功处理指令 |
| 500 | 设备处理失败 | 设备内部错误 |
| 777 | 设备离线 | 设备未连接 MQTT |
| 778 | 不支持 | 设备固件不支持该功能 |

---

## 附录 D：调用示例（curl）

### 读取所有属性

```bash
curl -X GET "http://127.0.0.1:8081/api/v1/radar/devices/BM87XXXX/properties"
```

### 读取指定属性

```bash
curl -X GET "http://127.0.0.1:8081/api/v1/radar/devices/BM87XXXX/properties?keys=install_height,heart_breath_switch"
```

### 写入安装高度

```bash
curl -X PUT "http://127.0.0.1:8081/api/v1/radar/devices/BM87XXXX/properties" \
  -H "Content-Type: application/json" \
  -d '{
    "properties": {
      "install_height": 280,
      "install_model": 0
    }
  }'
```

### 重启设备

```bash
curl -X POST "http://127.0.0.1:8081/api/v1/radar/devices/BM87XXXX/function" \
  -H "Content-Type: application/json" \
  -d '{"dev": 0}'
```

### 启动实时订阅

```bash
curl -X POST "http://127.0.0.1:8081/api/v1/radar/devices/BM87XXXX/subscribe" \
  -H "Content-Type: application/json" \
  -d '{
    "content": 0,
    "duration": 300
  }'
```

### 查询设备状态

```bash
curl -X GET "http://127.0.0.1:8081/api/v1/radar/devices/BM87XXXX/status"
```

---

## 附录 E：Python 调用示例

```python
import requests

QINGLAN_URL = "http://127.0.0.1:8081"

# 读取设备属性
def get_properties(device_uid, keys=None):
    url = f"{QINGLAN_URL}/api/v1/radar/devices/{device_uid}/properties"
    params = {"keys": keys} if keys else {}
    resp = requests.get(url, params=params, timeout=30)
    return resp.json()

# 写入设备属性
def set_properties(device_uid, properties):
    url = f"{QINGLAN_URL}/api/v1/radar/devices/{device_uid}/properties"
    resp = requests.put(url, json={"properties": properties}, timeout=30)
    return resp.json()

# 重启设备（0=全部, 1=雷达, 2=MCU）
def restart_device(device_uid, dev=0):
    url = f"{QINGLAN_URL}/api/v1/radar/devices/{device_uid}/function"
    resp = requests.post(url, json={"dev": dev}, timeout=30)
    return resp.json()

# 启动实时订阅
def subscribe_realtime(device_uid, content=0, duration=300):
    url = f"{QINGLAN_URL}/api/v1/radar/devices/{device_uid}/subscribe"
    resp = requests.post(url, json={"content": content, "duration": duration}, timeout=30)
    return resp.json()

# 查询设备状态
def get_status(device_uid):
    url = f"{QINGLAN_URL}/api/v1/radar/devices/{device_uid}/status"
    resp = requests.get(url, timeout=10)
    return resp.json()

# 使用示例
if __name__ == "__main__":
    uid = "BM87XXXX"

    # 读取安装高度
    result = get_properties(uid, "install_height")
    print(result)

    # 设置安装高度为 280cm
    result = set_properties(uid, {"install_height": 280})
    print(result)

    # 重启设备
    result = restart_device(uid, dev=0)
    print(result)
```
