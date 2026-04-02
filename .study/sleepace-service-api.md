# sleepace-service API 接口文档

> Sleepace 厂商协议层服务 HTTP API 完整参考手册
>
> 基于 wisefido-sleepace 源码反向分析生成，版本：2026-04-02

## 服务信息

- **默认端口**: 8090
- **Base URL**: `http://localhost:8090`
- **技术栈**: Java + Tomcat
- **数据库**: MySQL 5.7
- **角色**: 厂商协议层，直接与 Sleepace 硬件通信
- **通信协议**: MQTT + 厂商私有协议

## 架构定位

```
wisefido-sleepace (:8083) --HTTP--> sleepace-service (:8090) --MQTT/厂商协议--> Sleepace 硬件
                                           |
                                           +--> MySQL (设备绑定、配置、报告)
```

**说明：**
- sleepace-service 不对外直接暴露，仅供 wisefido-sleepace 内部调用
- 所有 API 均需要 token 认证（appId + secureKey）
- 数据通过 MySQL 持久化，通过 MQTT 与设备实时通信

---

## 通用请求/响应格式

### 请求格式

所有接口均为 **POST** 方法，请求体统一为：

```json
{
  "token": {
    "appId": "your-app-id",
    "secureKey": "your-secure-key"
  },
  "data": {
    // 接口特定参数
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | object | 是 | 认证令牌 |
| token.appId | string | 是 | 应用 ID |
| token.secureKey | string | 是 | 安全密钥 |
| data | object | 是 | 接口特定的请求参数 |

### 响应格式

所有接口响应统一为：

```json
{
  "status": 0,
  "msg": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| status | number | 状态码：`0` = 成功，非 0 = 失败 |
| msg | string | 状态描述信息 |
| data | any | 响应数据，可能是对象、数组或 null |

---

## API 端点列表

| 端点 | 功能 |
|------|------|
| `/sleepace/system/pushType/set` | 设置系统推送类型 |
| `/sleepace/bind` | 绑定设备到用户 |
| `/sleepace/unbind` | 解绑设备 |
| `/sleepace/deviceInfo/plaintextId` | 通过硬件 ID 查询设备信息 |
| `/sleepace/deviceInfo/deviceId` | 通过平台 ID 查询设备信息 |
| `/sleepace/heartModeSet` | 设置心率监测模式 |
| `/sleepace/device/updateconfig` | 更新实时数据上报间隔 |
| `/sleepace/device/updateAlgMode` | 更新离床检测灵敏度 |
| `/sleepace/device/getAlgMode` | 获取离床检测灵敏度 |
| `/sleepace/updateSetting` | 更新床垫厚度和材质参数 |
| `/sleepace/getalarmnotifyconfig` | 获取告警通知配置 |
| `/sleepace/updatealarmnotifyconfig` | 更新告警通知配置 |
| `/sleepace/reportUploadType/set` | 设置睡眠报告上传类型 |
| `/sleepace/setReportUploadTime` | 设置睡眠报告上传时间 |
| `/sleepace/get24HourDailyWithMaxReport` | 获取 24 小时睡眠报告 |

---

## 1. 系统配置

### POST /sleepace/system/pushType/set

设置系统数据推送类型（初始化时调用）。

**data 参数**:
```json
{
  "pushType": "MQTT",
  "alarmDataType": "array"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pushType | string | 是 | 推送类型：`"MQTT"` 或 `"HTTP"` |
| alarmDataType | string | 是 | 告警数据格式：`"array"` 或 `"object"` |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

## 2. 设备管理

### POST /sleepace/bind

绑定 Sleepace 设备到用户。

**data 参数**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "leftRight": 0,
  "userId": "user123",
  "gender": 1,
  "age": 50,
  "timezone": 28800
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceId | string | 是 | Sleepace 平台设备 ID（device_code） |
| leftRight | number | 是 | 左右床位：`0` = 左侧（单人床）, `1` = 右侧 |
| userId | string | 是 | 用户 ID |
| gender | number | 是 | 性别：`1` = 男，`2` = 女 |
| age | number | 是 | 年龄 |
| timezone | number | 是 | 时区偏移（秒），如东八区 = 28800 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

### POST /sleepace/unbind

解绑设备。

**data 参数**:
```json
{
  "deviceId": "1ua3erivl9pv1"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceId | string | 是 | Sleepace 平台设备 ID |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

### POST /sleepace/deviceInfo/plaintextId

通过硬件标签（BM87...）查询设备信息，获取平台 deviceId。

**data 参数**:
```json
{
  "plaintextId": "BM87224601903"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| plaintextId | string | 是 | 设备硬件标签（device_uid） |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": {
    "deviceId": "1ua3erivl9pv1",
    "plaintextId": "BM87224601903",
    "deviceType": 1,
    "deviceVersion": "1.2.3"
  }
}
```

**data 字段说明**:
| 字段 | 类型 | 说明 |
|------|------|------|
| deviceId | string | Sleepace 平台设备 ID（MQTT 消息中使用） |
| plaintextId | string | 设备硬件标签 |
| deviceType | number | 设备类型 |
| deviceVersion | string | 固件版本 |

---

### POST /sleepace/deviceInfo/deviceId

通过平台 deviceId 反查硬件标签。

**data 参数**:
```json
{
  "deviceId": "1ua3erivl9pv1"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceId | string | 是 | Sleepace 平台设备 ID |

**响应示例**：同上。

---

## 3. 实时监测配置

### POST /sleepace/heartModeSet

设置心率监测模式。

**data 参数**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "mode": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceId | string | 是 | Sleepace 平台设备 ID |
| mode | number | 是 | `0` = 关闭心率监测，`1` = 开启心率监测 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

### POST /sleepace/device/updateconfig

更新实时数据上报间隔。

**data 参数**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "userId": "user123",
  "leftRight": 0,
  "interval": 30
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceId | string | 是 | Sleepace 平台设备 ID |
| userId | string | 是 | 用户 ID |
| leftRight | number | 是 | 左右床位：`0` = 左，`1` = 右 |
| interval | number | 是 | 上报间隔（秒），范围 10~60 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

### POST /sleepace/device/updateAlgMode

更新离床检测灵敏度。

**data 参数**:
```json
{
  "deviceId": "1ua3erivl9pv1",
  "userId": "user123",
  "leftRight": 0,
  "mode": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceId | string | 是 | Sleepace 平台设备 ID |
| userId | string | 是 | 用户 ID |
| leftRight | number | 是 | 左右床位 |
| mode | number | 是 | `0` = 低灵敏度，`1` = 中等（推荐），`2` = 高灵敏度 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

### POST /sleepace/device/getAlgMode

获取当前离床检测灵敏度。

**data 参数**:
```json
{
  "userId": "user123",
  "deviceId": "1ua3erivl9pv1",
  "leftRight": 0
}
```

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": {
    "aloMode": 1
  }
}
```

**data.aloMode** 取值同 updateAlgMode 的 mode。

---

### POST /sleepace/updateSetting

更新床垫厚度和材质参数，用于算法校准。

**data 参数**:
```json
{
  "userId": "user123",
  "deviceId": "1ua3erivl9pv1",
  "leftRight": 0,
  "thickness": 20,
  "material": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | string | 是 | 用户 ID |
| deviceId | string | 是 | Sleepace 平台设备 ID |
| leftRight | number | 是 | 左右床位 |
| thickness | number | 是 | 床垫厚度（cm），范围 5~50 |
| material | number | 是 | 床垫材质：`0` = 弹簧，`1` = 记忆棉，`2` = 乳胶 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

## 4. 告警配置

### POST /sleepace/getalarmnotifyconfig

获取告警通知配置。

**data 参数**:
```json
{
  "userId": "user123",
  "deviceId": "1ua3erivl9pv1"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | string | 是 | 用户 ID |
| deviceId | string | 是 | Sleepace 平台设备 ID |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
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
```

**告警类型说明**:
| alarmType | 含义 | threshold 是否必需 |
|-----------|------|-------------------|
| HEARTRATE_HIGH | 心率过高 | 是 |
| HEARTRATE_LOW | 心率过低 | 是 |
| BREATHRATE_HIGH | 呼吸率过高 | 是 |
| BREATHRATE_LOW | 呼吸率过低 | 是 |
| LEAVE_BED | 离床 | 否 |
| IN_BED | 上床 | 否 |

---

### POST /sleepace/updatealarmnotifyconfig

更新告警通知配置。

**data 参数**:
```json
{
  "userId": "user123",
  "deviceId": "1ua3erivl9pv1",
  "alarmNotifySettings": [
    {
      "alarmType": "HEARTRATE_HIGH",
      "isEnabled": 1,
      "threshold": 130
    },
    {
      "alarmType": "LEAVE_BED",
      "isEnabled": 1
    }
  ]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | string | 是 | 用户 ID |
| deviceId | string | 是 | Sleepace 平台设备 ID |
| alarmNotifySettings | array | 是 | 告警配置数组 |
| alarmNotifySettings[].alarmType | string | 是 | 告警类型 |
| alarmNotifySettings[].isEnabled | number | 是 | `0` = 禁用，`1` = 启用 |
| alarmNotifySettings[].threshold | number | 条件 | 阈值类告警必填 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

## 5. 睡眠报告

### POST /sleepace/reportUploadType/set

设置睡眠报告上传类型。

**data 参数**:
```json
{
  "userId": "user123",
  "deviceId": "1ua3erivl9pv1",
  "reportUploadType": 0
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | string | 是 | 用户 ID |
| deviceId | string | 是 | Sleepace 平台设备 ID |
| reportUploadType | number | 是 | `0` = 定时自动上传，`1` = 手动触发上传 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

### POST /sleepace/setReportUploadTime

设置自动上传睡眠报告的时间（仅 reportUploadType=0 时生效）。

**data 参数**:
```json
{
  "userId": "user123",
  "deviceId": "1ua3erivl9pv1",
  "leftRight": 0,
  "reportUploadTime": 10
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | string | 是 | 用户 ID |
| deviceId | string | 是 | Sleepace 平台设备 ID |
| leftRight | number | 是 | 左右床位 |
| reportUploadTime | number | 是 | 上传时间（小时），范围 0~23，如 `10` 表示上午 10 点 |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": null
}
```

---

### POST /sleepace/get24HourDailyWithMaxReport

获取指定时间范围的 24 小时睡眠报告。

**data 参数**:
```json
{
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "startTime": 1711920000000,
  "endTime": 1712006400000
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | string | 是 | **注意：此处 userId 必须是 devices.device_id（UUID），不是 device_uid** |
| startTime | number | 是 | 开始时间（Unix 时间戳，毫秒） |
| endTime | number | 是 | 结束时间（Unix 时间戳，毫秒） |

**响应示例**:
```json
{
  "status": 0,
  "msg": "success",
  "data": [
    {
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
  ]
}
```

**data 字段说明**:
| 字段 | 类型 | 说明 |
|------|------|------|
| reportId | string | 报告唯一标识 |
| deviceId | string | Sleepace 平台设备 ID |
| reportDate | string | 报告日期（YYYYMMDD） |
| sleepScore | number | 睡眠评分（0~100） |
| totalSleepTime | number | 总睡眠时长（分钟） |
| deepSleepTime | number | 深睡时长（分钟） |
| lightSleepTime | number | 浅睡时长（分钟） |
| remSleepTime | number | REM 睡眠时长（分钟） |
| awakeTime | number | 清醒时长（分钟） |
| avgHeartRate | number | 平均心率（次/分钟） |
| avgBreathRate | number | 平均呼吸率（次/分钟） |
| turnOverTimes | number | 翻身次数 |
| leaveBedTimes | number | 离床次数 |

---

## 附录 A：字段映射关系

| sleepace-service 字段 | 对应的 wisefido 字段 | 说明 |
|----------------------|---------------------|------|
| deviceId | device_code | Sleepace 平台设备 ID（MQTT 中使用） |
| plaintextId | device_uid | 设备硬件标签（BM87...） |
| userId (bind) | device_id | 平台业务 UUID（devices 表） |
| userId (get report) | device_id | **特别注意：报告接口的 userId 是 UUID** |
| leftRight | — | 床位：0=左/单人，1=右 |

---

## 附录 B：初始化流程

wisefido-sleepace 在设备初始化时按顺序调用以下接口：

```
1. POST /sleepace/bind                    绑定设备
2. POST /sleepace/heartModeSet            开启心率监测
3. POST /sleepace/device/updateconfig     设置实时间隔（默认 30s）
4. POST /sleepace/device/updateAlgMode    设置离床灵敏度（默认中等）
5. POST /sleepace/reportUploadType/set    设置报告上传类型（默认自动）
6. POST /sleepace/setReportUploadTime     设置上传时间（默认上午 10 点）
```

任一步骤失败不影响后续步骤执行。

---

## 附录 C：调用示例（curl）

### 绑定设备

```bash
curl -X POST "http://localhost:8090/sleepace/bind" \
  -H "Content-Type: application/json" \
  -d '{
    "token": {
      "appId": "your-app-id",
      "secureKey": "your-secure-key"
    },
    "data": {
      "deviceId": "1ua3erivl9pv1",
      "leftRight": 0,
      "userId": "user123",
      "gender": 1,
      "age": 50,
      "timezone": 28800
    }
  }'
```

### 更新告警配置

```bash
curl -X POST "http://localhost:8090/sleepace/updatealarmnotifyconfig" \
  -H "Content-Type: application/json" \
  -d '{
    "token": {
      "appId": "your-app-id",
      "secureKey": "your-secure-key"
    },
    "data": {
      "userId": "user123",
      "deviceId": "1ua3erivl9pv1",
      "alarmNotifySettings": [
        {
          "alarmType": "HEARTRATE_HIGH",
          "isEnabled": 1,
          "threshold": 120
        }
      ]
    }
  }'
```

### 获取睡眠报告

```bash
curl -X POST "http://localhost:8090/sleepace/get24HourDailyWithMaxReport" \
  -H "Content-Type: application/json" \
  -d '{
    "token": {
      "appId": "your-app-id",
      "secureKey": "your-secure-key"
    },
    "data": {
      "userId": "550e8400-e29b-41d4-a716-446655440000",
      "startTime": 1711920000000,
      "endTime": 1712006400000
    }
  }'
```

---

## 附录 D：错误处理

当 `status != 0` 时表示请求失败，`msg` 字段包含错误描述。

**常见错误**:
| status | msg | 原因 |
|--------|-----|------|
| 1 | Invalid token | token 认证失败 |
| 2 | Device not found | 设备不存在或未绑定 |
| 3 | User not found | 用户不存在 |
| 4 | Parameter error | 参数错误或缺失 |
| 5 | Device offline | 设备离线，无法下发配置 |

---

## 附录 E：与 wisefido-sleepace 的关系

**调用链路：**

```
wisefido-data (:8080)
    ↓ HTTP
wisefido-sleepace (:8083)  ← 代理层，添加 token
    ↓ HTTP (POST)
sleepace-service (:8090)   ← 本文档涵盖的服务
    ↓ MQTT / 厂商协议
Sleepace 硬件
```

**wisefido-sleepace 提供的封装：**
- 自动注入 token（从配置文件读取）
- 重试机制（3 次，间隔 1~5 秒）
- 超时控制（10 秒）
- 错误统一处理

**直接调用 vs 通过 wisefido-sleepace：**
- **直接调用**：需自行管理 token、重试、超时
- **通过 wisefido-sleepace**：使用 `/api/v1/proxy/*` 端点，自动注入 token，更简洁

推荐通过 wisefido-sleepace 间接调用，而非直接调用 sleepace-service。
