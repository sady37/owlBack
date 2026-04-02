# wisefido-sleepace API 接口文档

> Sleepace 睡眠板设备网关服务 HTTP API 完整参考手册
>
> 基于源码分析生成，版本：2026-04-02

## 服务信息

- **默认端口**: 8083
- **Base URL**: `http://127.0.0.1:8083`
- **协议**: HTTP
- **路由库**: gorilla/mux
- **角色**: 网关/代理服务，连接 wisefido 平台与 sleepace-service (Java) 厂商服务

## 架构概览

```
wisefido-data (:8080)
    ↓ HTTP
wisefido-sleepace (:8083) ← 本文档覆盖的服务
    ↓ HTTP (代理)
sleepace-service (:8090) - Java 厂商协议层
    ↓ MQTT / 厂商协议
Sleepace 硬件设备
```

**数据上行**:
```
Sleepace 硬件 → sleepace-service (MySQL) → MQTT → wisefido-sleepace → iot:*:stream
```

**指令下行**:
```
wisefido-data → wisefido-sleepace → sleepace-service → MQTT → Sleepace 硬件
```

## 通用响应格式

所有接口均使用统一响应格式：

```json
{
  "success": true,
  "data": {},
  "error": "",
  "message": ""
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 操作是否成功 |
| data | object/array | 成功时返回的数据 |
| error | string | 失败时返回的错误信息 |
| message | string | 附加说明信息 |

---

## API 端点列表

| 方法 | 路径 | 功能分类 |
|------|------|---------|
| GET | `/health` | 健康检查 |
| POST | `/api/v1/proxy/*` | 透明代理到 sleepace-service |
| POST | `/api/v1/sleepace/device/initialize` | 设备初始化/绑定 |
| DELETE | `/api/v1/sleepace/device/{code}` | 设备解绑 |
| GET | `/api/v1/sleepace/devices/status` | 批量查询设备状态 |
| GET | `/api/v1/sleepace/devices/{uid}/status` | 查询单设备状态 |

---

## 1. 健康检查

### GET /health

系统健康检查端点。

**请求参数**: 无

**响应示例**:
```json
{
  "status": "healthy",
  "service": "wisefido-sleepace",
  "time": "2026-04-02T10:30:00+08:00"
}
```

---

## 2. 透明代理端点

### POST /api/v1/proxy/*

将请求透明代理到 sleepace-service (Java 服务)，自动注入认证 Token。

**路径参数**:
| 参数 | 说明 |
|------|------|
| * | 任意路径，会转发到 `http://localhost:8090/{匹配的路径}` |

**请求头**:
| 头部 | 必填 | 说明 |
|------|------|------|
| X-Sleepace-Token | 否 | 可选自定义 token，不传则使用配置的默认 token |

**行为说明**:
- 自动在请求体中注入 `"token": "{token}"`
- 转发到配置的 sleepace-service 地址（默认 `http://localhost:8090`）
- 超时 30 秒

**常用代理端点**:

| 代理路径 | 实际调用 | 功能 |
|---------|---------|------|
| `POST /api/v1/proxy/sleepace/updatealarmnotifyconfig` | `POST http://localhost:8090/sleepace/updatealarmnotifyconfig` | 更新告警配置 |
| `POST /api/v1/proxy/sleepace/getalarmnotifyconfig` | `POST http://localhost:8090/sleepace/getalarmnotifyconfig` | 获取告警配置 |
| `POST /api/v1/proxy/sleepace/device/updateconfig` | `POST http://localhost:8090/sleepace/device/updateconfig` | 更新实时数据间隔 |
| `POST /api/v1/proxy/sleepace/heartModeSet` | `POST http://localhost:8090/sleepace/heartModeSet` | 设置心率模式 |
| `POST /api/v1/proxy/sleepace/device/updateAlgMode` | `POST http://localhost:8090/sleepace/device/updateAlgMode` | 更新离床灵敏度 |
| `POST /api/v1/proxy/sleepace/updateSetting` | `POST http://localhost:8090/sleepace/updateSetting` | 更新床垫参数 |
| `POST /api/v1/proxy/sleepace/get24HourDailyWithMaxReport` | `POST http://localhost:8090/sleepace/get24HourDailyWithMaxReport` | 下载睡眠报告 |

**请求示例** (更新告警配置):
```bash
POST /api/v1/proxy/sleepace/updatealarmnotifyconfig
Content-Type: application/json

{
  "deviceId": "1ua3erivl9pv1",
  "alarmNotifySettings": [
    {
      "alarmType": "HEARTRATE_HIGH",
      "isEnabled": 1,
      "threshold": 120
    }
  ]
}
```

**成功响应**:
```json
{
  "success": true,
  "data": {
    "code": 0,
    "msg": "success"
  }
}
```

---

## 3. 设备初始化/绑定

### POST /api/v1/sleepace/device/initialize

初始化并绑定 Sleepace 设备到用户。

**请求体**:
```json
{
  "device_code": "1ua3erivl9pv1",
  "user_code": "user123",
  "device_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_code | string | 是 | Sleepace 设备编码（MQTT deviceId） |
| user_code | string | 是 | 用户编码 |
| device_id | string | 是 | 平台设备 UUID（devices 表主键） |

**行为说明**:
1. 调用 sleepace-service 的 `/sleepace/bind` 接口绑定设备
2. 设置默认配置：
   - 心率模式：开启
   - 实时数据间隔：30 秒
   - 离床灵敏度：中等
   - 报告上传类型：自动
   - 报告上传时间：上午 10 点

**成功响应**:
```json
{
  "success": true,
  "message": "Device initialized successfully",
  "data": {
    "device_code": "1ua3erivl9pv1",
    "bind_result": {...}
  }
}
```

**错误响应** (400):
```json
{
  "success": false,
  "error": "device_code is required"
}
```

---

## 4. 设备解绑

### DELETE /api/v1/sleepace/device/{code}

解绑 Sleepace 设备。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | Sleepace 设备编码 |

**成功响应**:
```json
{
  "success": true,
  "message": "Device unbound successfully"
}
```

**错误响应** (500):
```json
{
  "success": false,
  "error": "Failed to unbind device: connection timeout"
}
```

---

## 5. 批量查询设备状态

### GET /api/v1/sleepace/devices/status

批量查询多个 Sleepace 设备的在线状态。

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
      "device_uid": "1ua3erivl9pv1",
      "device_id": "550e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "tenant-uuid",
      "status": "online"
    },
    {
      "device_uid": "2vb4fsjwm0qw2",
      "device_id": "660e8400-e29b-41d4-a716-446655440001",
      "tenant_id": "tenant-uuid",
      "status": "offline"
    }
  ],
  "count": 2
}
```

**状态值说明**:
| status | 含义 |
|--------|------|
| online | 设备在线 |
| offline | 设备离线 |

---

## 6. 查询单设备状态

### GET /api/v1/sleepace/devices/{uid}/status

查询单个 Sleepace 设备的在线状态。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | string | 是 | Sleepace 设备编码（device_uid） |

**成功响应**:
```json
{
  "success": true,
  "data": {
    "device_uid": "1ua3erivl9pv1",
    "status": "online"
  }
}
```

---

## 附录 A：sleepace-service 常用端点详解

以下是通过代理端点可访问的 sleepace-service (Java) 核心接口。

### A.1 告警配置

#### 获取告警配置

**代理路径**: `POST /api/v1/proxy/sleepace/getalarmnotifyconfig`

**请求体**:
```json
{
  "deviceId": "1ua3erivl9pv1"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "code": 0,
    "data": {
      "alarmNotifySettings": [
        {
          "alarmType": "HEARTRATE_HIGH",
          "isEnabled": 1,
          "threshold": 120
        },
        {
          "alarmType": "HEARTRATE_LOW",
          "isEnabled": 1,
          "threshold": 40
        },
        {
          "alarmType": "BREATHRATE_HIGH",
          "isEnabled": 1,
          "threshold": 25
        },
        {
          "alarmType": "BREATHRATE_LOW",
          "isEnabled": 1,
          "threshold": 8
        },
        {
          "alarmType": "LEAVE_BED",
          "isEnabled": 1
        },
        {
          "alarmType": "IN_BED",
          "isEnabled": 0
        }
      ]
    }
  }
}
```

**告警类型说明**:
| alarmType | 含义 | 是否需要阈值 |
|-----------|------|-------------|
| HEARTRATE_HIGH | 心率过高 | 是 |
| HEARTRATE_LOW | 心率过低 | 是 |
| BREATHRATE_HIGH | 呼吸率过高 | 是 |
| BREATHRATE_LOW | 呼吸率过低 | 是 |
| LEAVE_BED | 离床 | 否 |
| IN_BED | 上床 | 否 |

#### 更新告警配置

**代理路径**: `POST /api/v1/proxy/sleepace/updatealarmnotifyconfig`

**请求体**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "alarmNotifySettings": [
    {
      "alarmType": "HEARTRATE_HIGH",
      "isEnabled": 1,
      "threshold": 120
    }
  ]
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "code": 0,
    "msg": "success"
  }
}
```

### A.2 设备配置

#### 更新实时数据间隔

**代理路径**: `POST /api/v1/proxy/sleepace/device/updateconfig`

**请求体**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "realtimeInterval": 30
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| deviceId | string | 设备编码 |
| realtimeInterval | number | 实时数据上报间隔（秒），范围 10~60 |

**响应**:
```json
{
  "success": true,
  "data": {
    "code": 0,
    "msg": "success"
  }
}
```

#### 设置心率模式

**代理路径**: `POST /api/v1/proxy/sleepace/heartModeSet`

**请求体**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "heartMode": 1
}
```

| heartMode | 含义 |
|-----------|------|
| 0 | 关闭心率监测 |
| 1 | 开启心率监测 |

#### 更新离床灵敏度

**代理路径**: `POST /api/v1/proxy/sleepace/device/updateAlgMode`

**请求体**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "algMode": 1
}
```

| algMode | 含义 |
|---------|------|
| 0 | 低灵敏度 |
| 1 | 中等灵敏度（推荐） |
| 2 | 高灵敏度 |

#### 更新床垫参数

**代理路径**: `POST /api/v1/proxy/sleepace/updateSetting`

**请求体**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "mattressThickness": 20,
  "mattressMaterial": "memory_foam"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| mattressThickness | number | 床垫厚度（cm），范围 5~50 |
| mattressMaterial | string | 床垫材质：`spring`（弹簧）/ `memory_foam`（记忆棉）/ `latex`（乳胶） |

### A.3 报告管理

#### 下载睡眠报告

**代理路径**: `POST /api/v1/proxy/sleepace/get24HourDailyWithMaxReport`

**请求体**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "reportDate": "20260402"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| deviceId | string | 设备编码 |
| reportDate | string | 报告日期，格式 YYYYMMDD |

**响应**:
```json
{
  "success": true,
  "data": {
    "code": 0,
    "data": {
      "reportId": "report-uuid",
      "deviceId": "1ua3erivl9pv1",
      "reportDate": "20260402",
      "sleepScore": 85,
      "totalSleepTime": 420,
      "deepSleepTime": 120,
      "lightSleepTime": 250,
      "remSleepTime": 50,
      "awakeTime": 30,
      "avgHeartRate": 65,
      "avgBreathRate": 15,
      "turnOverTimes": 8,
      "leaveBedTimes": 2
    }
  }
}
```

---

## 附录 B：设备绑定流程

完整的设备绑定流程如下：

```
1. wisefido-data 调用初始化接口
   POST /api/v1/sleepace/device/initialize
   body: {device_code, user_code, device_id}
       ↓
2. wisefido-sleepace 调用 sleepace-service 绑定
   POST http://localhost:8090/sleepace/bind
   body: {deviceId, userCode, token}
       ↓
3. wisefido-sleepace 设置默认配置
   - 心率模式：开启
   - 实时间隔：30s
   - 离床灵敏度：中等
   - 报告上传：自动
       ↓
4. sleepace-service 将绑定信息存入 MySQL
       ↓
5. sleepace-service 通过 MQTT 通知设备配置生效
       ↓
6. 设备开始按配置上报数据到 MQTT
       ↓
7. wisefido-sleepace 消费 MQTT 消息，写入 iot:*:stream
```

---

## 附录 C：调用示例（curl）

### 初始化设备

```bash
curl -X POST "http://127.0.0.1:8083/api/v1/sleepace/device/initialize" \
  -H "Content-Type: application/json" \
  -d '{
    "device_code": "1ua3erivl9pv1",
    "user_code": "user123",
    "device_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

### 更新告警配置（通过代理）

```bash
curl -X POST "http://127.0.0.1:8083/api/v1/proxy/sleepace/updatealarmnotifyconfig" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "1ua3erivl9pv1",
    "alarmNotifySettings": [
      {
        "alarmType": "HEARTRATE_HIGH",
        "isEnabled": 1,
        "threshold": 120
      }
    ]
  }'
```

### 查询设备状态

```bash
curl -X GET "http://127.0.0.1:8083/api/v1/sleepace/devices/1ua3erivl9pv1/status"
```

### 解绑设备

```bash
curl -X DELETE "http://127.0.0.1:8083/api/v1/sleepace/device/1ua3erivl9pv1"
```

---

## 附录 D：Python 调用示例

```python
import requests

SLEEPACE_URL = "http://127.0.0.1:8083"

# 初始化设备
def initialize_device(device_code, user_code, device_id):
    url = f"{SLEEPACE_URL}/api/v1/sleepace/device/initialize"
    payload = {
        "device_code": device_code,
        "user_code": user_code,
        "device_id": device_id
    }
    resp = requests.post(url, json=payload, timeout=30)
    return resp.json()

# 更新告警配置（通过代理）
def update_alarm_config(device_id, alarm_settings):
    url = f"{SLEEPACE_URL}/api/v1/proxy/sleepace/updatealarmnotifyconfig"
    payload = {
        "deviceId": device_id,
        "alarmNotifySettings": alarm_settings
    }
    resp = requests.post(url, json=payload, timeout=30)
    return resp.json()

# 设置心率模式
def set_heart_mode(device_id, enabled=True):
    url = f"{SLEEPACE_URL}/api/v1/proxy/sleepace/heartModeSet"
    payload = {
        "deviceId": device_id,
        "heartMode": 1 if enabled else 0
    }
    resp = requests.post(url, json=payload, timeout=30)
    return resp.json()

# 更新实时数据间隔
def set_realtime_interval(device_id, interval=30):
    url = f"{SLEEPACE_URL}/api/v1/proxy/sleepace/device/updateconfig"
    payload = {
        "deviceId": device_id,
        "realtimeInterval": interval
    }
    resp = requests.post(url, json=payload, timeout=30)
    return resp.json()

# 查询设备状态
def get_device_status(device_uid):
    url = f"{SLEEPACE_URL}/api/v1/sleepace/devices/{device_uid}/status"
    resp = requests.get(url, timeout=10)
    return resp.json()

# 下载睡眠报告
def get_sleep_report(device_id, report_date):
    url = f"{SLEEPACE_URL}/api/v1/proxy/sleepace/get24HourDailyWithMaxReport"
    payload = {
        "deviceId": device_id,
        "reportDate": report_date  # "20260402"
    }
    resp = requests.post(url, json=payload, timeout=30)
    return resp.json()

# 使用示例
if __name__ == "__main__":
    device_code = "1ua3erivl9pv1"
    device_id = "550e8400-e29b-41d4-a716-446655440000"

    # 初始化设备
    result = initialize_device(device_code, "user123", device_id)
    print(result)

    # 更新告警配置
    alarms = [
        {"alarmType": "HEARTRATE_HIGH", "isEnabled": 1, "threshold": 120},
        {"alarmType": "LEAVE_BED", "isEnabled": 1}
    ]
    result = update_alarm_config(device_code, alarms)
    print(result)

    # 查询状态
    result = get_device_status(device_code)
    print(result)
```

---

## 附录 E：与 wisefido-qinglan 的对比

| 维度 | wisefido-qinglan (Radar) | wisefido-sleepace (Sleep Pad) |
|------|-------------------------|------------------------------|
| 端口 | 8081 | 8083 |
| 厂商服务 | 无（直连设备） | sleepace-service (Java, :8090) |
| 参数标识 | device_uid (硬件 ID) | device_code (Sleepace deviceId) |
| 协议复杂度 | 高（TDPv2 二进制+JSON） | 低（代理到厂商服务） |
| 设备控制 | 直接 MQTT | 通过 sleepace-service |
| 主要功能 | 直接控制硬件 | 代理 + 设备生命周期管理 |
| 响应等待 | 是（轮询 Redis） | 否（sleepace-service 处理） |

---

## 附录 F：配置文件位置

- **wisefido-sleepace 配置**: `/home/wenhe/Study/owl-sady/owlBack/wisefido-sleepace/sleepace-dev.yaml`
- **sleepace-service 配置**: `/home/wenhe/Study/owl-sady/trans/sleepace/sleepace-service/classes/`
- **Docker Compose**: `/home/wenhe/Study/owl-sady/owlBack/docker-compose.yml` (MySQL 容器配置)
