# Vue-Radar 与 Qinglan 对接流程文档

## 概述

本文档描述 vue-radar 前端组件如何通过 `radar_install_service.go` 和 `qinglan_client.go` 实现安装模式/高度边界设置，并接收实时数据的完整流程。

**重要说明**：
- 原始的 vue-radar 是一个独立的 Vue 应用（参考 `vue_radar/ARCHITECTURE_DIAGRAM.md`）
- 已整合进 owlFront 项目中，通过 `RadarAppWrapper` 组件包装
- 本文档描述整合后的对接流程

## 架构流程

### 整合后的架构

```
owlFront 项目
    ├─ RadarTrajectory.vue (页面)
    │   └─ RadarAppWrapper.vue (包装组件)
    │       ├─ RadarCanvas.vue (原始组件)
    │       ├─ WaveMonitor.vue (原始组件)
    │       └─ Toolbar.vue (原始组件)
    │           └─ 通过 provide('externalCallbacks') 提供回调
    │               ├─ sendCommand() → emit('sendCommand')
    │               └─ queryDevice() → emit('queryDevice')
    │
    └─ RadarTrajectory.vue 处理回调
        ├─ handleSendCommand() → HTTP API
        └─ handleQueryDevice() → HTTP API
            ↓
RadarHandler (wisefido-data)
    ↓
RadarInstall Service
    ↓
QinglanClient
    ↓ HTTP API
wisefido-qinglan Service
    ↓ MQTT
雷达设备
```

### 原始 vue-radar 架构（参考）

原始 vue-radar 是一个独立应用，包含：
- `App.vue` - 主应用组件
- `RadarCanvas.vue` - 雷达画布（2647行）
- `WaveMonitor.vue` - 波形监测（1400+行）
- `Toolbar.vue` - 工具栏（3132行）
- Pinia Stores: `canvas.ts`, `objects.ts`, `radarData.ts`, `waveform.ts`

整合进 owlFront 后，通过 `RadarAppWrapper` 包装这些组件，并通过 provide/inject 机制提供外部回调。

## 1. 安装模式/高度边界设置流程

### 1.1 前端调用

**整合后的调用链**:

1. **Toolbar.vue** (`owlFront/src/components/Radar/Toolbar.vue`)
   - 用户操作触发 `WriteRadar()` 函数
   - 通过 `inject('externalCallbacks')` 获取回调函数
   - 调用 `externalCallbacks.sendCommand(deviceId, commandData)`

2. **RadarAppWrapper.vue** (`owlFront/src/views/monitoring/radar-trajectory/components/RadarAppWrapper.vue`)
   - 通过 `provide('externalCallbacks')` 提供回调
   - `sendCommand` 回调 emit `'sendCommand'` 事件

3. **RadarTrajectory.vue** (`owlFront/src/views/monitoring/radar-trajectory/RadarTrajectory.vue`)
   - 接收 `@send-command` 事件
   - 调用 `handleSendCommand()` 函数
   - **TODO**: 需要实现实际的 API 调用

**配置格式**（v1.0 API 格式）:
```typescript
const config = {
  install_model: 'wall' | 'ceiling' | 'corn',  // 安装模式
  height: 170,                                  // 高度（cm，前端单位）
  boundary_left: 300,                          // 左边界（cm）
  boundary_right: 300,                          // 右边界（cm）
  boundary_front: 400,                          // 前边界（cm）
  boundary_rear: 0                              // 后边界（cm）
};

// 需要调用的 API
PUT /radar-device/api/v1/radar-device/device/:id/config
```

**当前状态**:
- ✅ 前端组件已整合
- ✅ 回调机制已建立
- ⚠️ `handleSendCommand()` 需要实现实际的 API 调用（当前是 TODO）

### 1.2 API 路由

**位置**: `owlBack/wisefido-data/internal/http/radar_handler.go`

```go
// UpdateConfig 更新设备配置
// PUT /radar-device/api/v1/radar-device/device/:id/config
func (h *RadarHandler) UpdateConfig(w http.ResponseWriter, r *http.Request)
```

**处理流程**:
1. 从 URL 路径提取 `device_id`
2. 从请求头获取 `tenant_id` (X-Tenant-Id)
3. 解析请求体中的配置数据
4. 调用 `radarInstall.UpdateConfig()`

### 1.3 RadarInstall Service

**位置**: `owlBack/wisefido-data/internal/service/radar_install_service.go`

```go
// UpdateConfig 更新设备配置（v1.0 API 格式）
func (s *RadarInstall) UpdateConfig(ctx context.Context, tenantID, deviceID string, config map[string]interface{}) error
```

**处理流程**:
1. 通过 `GetDeviceUID()` 将 `device_id` 转换为 `device_uid`
2. 调用 `V1ConfigToRadarDeviceProps()` 将 v1.0 格式转换为设备属性格式
3. 调用 `SetDeviceProperties()` 设置设备属性

**格式转换** (`radar_device_props.go`):
- `install_model` (wall/ceiling) → `radar_install_style` ("0"/"1")
- `height` (cm) → `radar_install_height` (dm，除以10)
- `boundary_*` (cm) → `rectangle` (格式: `{x1, y1; x2, y2; x3, y3; x4, y4}`)

### 1.4 QinglanClient

**位置**: `owlBack/wisefido-data/internal/service/qinglan_client.go`

```go
// SetDeviceProperties 设置设备属性
// PUT /api/v1/radar/devices/{uid}/properties
func (c *QinglanClient) SetDeviceProperties(ctx context.Context, deviceUID string, properties map[string]interface{}) error
```

**关键特性**:
- **自动分组**: 设备要求必须分组设置，不能一次性设置所有属性
- **分组规则**:
  1. 工作模式组: `radar_func_ctrl`, `radar_install_style`, `radar_install_height` (3个key一组)
  2. 呼吸心率参数: `heart_breath_param` (单独1组)
  3. 跌倒参数: `fall_param` (单独1组)
  4. 边界: `rectangle` (单独1组)
  5. 区域: `declare_area` (一次只能设置一个区域)

- **按组依次发送**: 每组之间延迟 100ms，避免设备处理不过来

### 1.5 wisefido-qinglan Service

**位置**: `owlBack/wisefido-qinglan/internal/http/api.go`

```go
// SetDeviceProperties 设置设备属性
// PUT /api/v1/radar/devices/{uid}/properties
func (h *APIHandler) SetDeviceProperties(w http.ResponseWriter, r *http.Request)
```

**处理流程**:
1. 从 URL 提取 `device_uid`
2. 解析请求体中的属性
3. 调用 `radarService.SetDeviceProperties()` 通过 MQTT 下发到设备

## 2. 实时数据接收流程

### 2.1 订阅实时数据

**API**: `POST /api/v1/radar/devices/{uid}/subscribe`

**请求格式**:
```json
{
  "content": 0,    // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
  "duration": 3600 // 订阅时长（秒），最大3600
}
```

**调用链**:
1. `RadarInstall.SubscribeRealtimeData()` 
2. → `QinglanClient.SubscribeRealtimeData()`
3. → `wisefido-qinglan API: POST /api/v1/radar/devices/{uid}/subscribe`
4. → `RadarService.SubscribeRealtimeData()`
5. → `MQTTPublisher.SubscribeRealtimeData()` (通过 MQTT 发送订阅命令)

**MQTT 命令格式**:
```json
{
  "cmd": "subscription",
  "data": {
    "content": "0",    // 注意：必须使用字符串格式，不是数字
    "duration": 3600
  }
}
```

**MQTT 主题**: `monitor/{productId}/{deviceUID}/get`

**订阅管理**:
- 设备连接时自动订阅（通过 `DeviceSubscriptionManager`）
- 订阅到期前自动续订（默认每 50 分钟续订一次）
- 订阅信息存储在数据库中，支持持久化

### 2.2 数据流

```
雷达设备
    ↓ MQTT (发布实时数据到 monitor 主题)
wisefido-qinglan (MQTT Consumer)
    ↓ 解析并发布
Redis Streams (iot:monitor:stream, iot:stat:stream, iot:event:stream, iot:alarm:stream)
    ↓ 消费
wisefido-card-aggregator / wisefido-iot-timeseries
    ↓ HTTP API / WebSocket
前端 (Vue-Radar)
```

**数据流详细说明**:

1. **设备发布**: 雷达设备通过 MQTT 发布实时数据到 `monitor` 主题
2. **MQTT Consumer**: `wisefido-qinglan/internal/consumer/mqtt_consumer.go` 接收并解析数据
3. **发布到 Redis Streams**: 
   - `iot:monitor:stream` - 轨迹数据
   - `iot:stat:stream` - 统计数据
   - `iot:event:stream` - 事件数据
   - `iot:alarm:stream` - 告警数据
4. **数据消费**: 
   - `wisefido-card-aggregator` 消费并聚合数据
   - `wisefido-iot-timeseries` 消费并存储到 PostgreSQL
5. **前端获取**: 
   - HTTP 轮询: `GET /radar-device/api/v1/radar-device/device/:id/realtime`
   - 或通过 WebSocket 实时推送

**当前实现状态**:
- ✅ MQTT Consumer 已实现，接收设备数据并发布到 Redis Streams
- ✅ Redis Streams 数据流已建立
- ⚠️ `GetRealtimeData` API 当前返回空数据，需要实现从 Redis Streams 或 PostgreSQL 读取

### 2.3 数据格式

**轨迹数据** (positions):
```json
{
  "positions": [
    {
      "x": 100,
      "y": 200,
      "timestamp": 1234567890
    }
  ]
}
```

**生命体征数据** (vital):
```json
{
  "vital": {
    "heart_rate": 72,
    "breath_rate": 18,
    "timestamp": 1234567890
  }
}
```

## 3. 配置读取流程

### 3.1 获取设备原始属性

**API**: `GET /radar-device/api/v1/radar-device/device/:id/original-properties`

**调用链**:
1. `RadarHandler.GetOriginalProperties()`
2. → `RadarInstall.GetOriginalProperties()`
3. → `GetDeviceUID()` (device_id → device_uid)
4. → `GetDeviceProperties()` (读取所有属性)
5. → `QinglanClient.GetDeviceProperties()` 
6. → `wisefido-qinglan API: GET /api/v1/radar/devices/{uid}/properties`

**返回格式**: JSON 字符串，包含雷达所有配置参数

## 4. 关键代码位置

### 后端

| 功能 | 文件路径 | 关键函数 |
|------|---------|---------|
| API Handler | `wisefido-data/internal/http/radar_handler.go` | `UpdateConfig()`, `GetOriginalProperties()` |
| 安装服务 | `wisefido-data/internal/service/radar_install_service.go` | `UpdateConfig()`, `GetOriginalProperties()` |
| Qinglan 客户端 | `wisefido-data/internal/service/qinglan_client.go` | `SetDeviceProperties()`, `GetDeviceProperties()`, `SubscribeRealtimeData()` |
| 格式转换 | `wisefido-data/internal/service/radar_device_props.go` | `V1ConfigToRadarDeviceProps()` |
| Qinglan API | `wisefido-qinglan/internal/http/api.go` | `SetDeviceProperties()`, `SubscribeRealtimeData()` |

### 前端

| 功能 | 文件路径 | 关键函数 |
|------|---------|---------|
| 配置更新 | `owlFront/src/components/Radar/Toolbar.vue` | `WriteRadar()`, `QueryRadar()` |
| 配置转换 | `owlFront/src/config/radarMqttConfig.ts` | `convertInstallModel()`, `convertHeight()`, `convertBoundary()` |

## 5. 数据格式转换

### 5.1 安装模式

| 前端值 | 设备属性值 | 说明 |
|--------|-----------|------|
| `'ceiling'` | `"0"` | 顶装 |
| `'wall'` | `"1"` | 侧装 |
| `'corn'` | `"2"` | 角落安装 |

### 5.2 单位转换

| 前端单位 | 设备单位 | 转换规则 |
|---------|---------|---------|
| cm (厘米) | dm (分米) | `dm = cm / 10` (向下取整) |

**示例**:
- 前端: `height: 170` (cm)
- 设备: `radar_install_height: "17"` (dm)

### 5.3 边界格式

**前端格式**:
```typescript
{
  boundary_left: 300,   // cm
  boundary_right: 300,  // cm
  boundary_front: 400,  // cm
  boundary_rear: 0      // cm
}
```

**设备格式** (`rectangle`):
```
{x1, y1; x2, y2; x3, y3; x4, y4}
```

**转换规则** (顶装):
```
x1 = -boundary_left,  y1 = -boundary_front
x2 = boundary_right,  y2 = -boundary_front
x3 = -boundary_left,  y3 = boundary_rear
x4 = boundary_right,  y4 = boundary_rear
```

**转换规则** (侧装):
```
x1 = -boundary_left,  y1 = 0
x2 = boundary_right,  y2 = 0
x3 = -boundary_left,  y3 = boundary_rear
x4 = boundary_right,  y4 = boundary_rear
```

## 6. 注意事项

### 6.1 属性分组

设备要求必须分组设置属性，`QinglanClient.SetDeviceProperties()` 会自动分组：
- 工作模式组（`radar_func_ctrl`, `radar_install_style`, `radar_install_height`）必须一起发送
- 每组之间延迟 100ms，避免设备处理不过来

### 6.2 单位转换

- 前端使用 **cm (厘米)**
- 设备使用 **dm (分米)**
- 转换时注意取整方向（前端除以10后向下取整）

### 6.3 实时数据

- 当前 `GetRealtimeData` 返回空数据，需要实现从 Redis Streams 或 PostgreSQL 读取
- 建议使用 WebSocket 替代 HTTP 轮询，提高实时性

### 6.4 错误处理

- 所有 API 调用都应包含错误处理
- 设备属性设置失败时，应同步实际值到前端（通过 `RadarDevicePropsToAlarmItems` 反向解析）

## 7. 测试建议

### 7.1 配置更新测试

1. 测试安装模式切换（ceiling/wall/corn）
2. 测试高度设置（15-33 dm 范围）
3. 测试边界设置（4个边界值）
4. 验证单位转换正确性

### 7.2 实时数据测试

1. 订阅实时数据
2. 验证数据格式正确
3. 测试数据延迟
4. 测试订阅续期（50分钟自动续订）

### 7.3 错误场景测试

1. 设备离线时的处理
2. 网络错误时的重试机制
3. 属性设置失败时的回滚

## 8. 整合进 owlFront 后的集成工作

### 8.1 当前状态

**已完成**:
- ✅ vue-radar 组件已整合进 owlFront（组件文件在 `src/components/Radar/`）
- ✅ `RadarAppWrapper` 组件已创建，包装原始组件
- ✅ `RadarTrajectory.vue` 页面已创建，路由已配置
- ✅ provide/inject 回调机制已建立
- ✅ 后端 API 已实现（`radar_handler.go`, `radar_install_service.go`, `qinglan_client.go`）
- ✅ 雷达监控设置 API 已实现（`/settings/api/v1/monitor/radar/:deviceId`，用于工作模式、跌倒、呼吸心率等）

**待完成**:
- ⚠️ **雷达安装配置 API 未实现**：`RadarTrajectory.vue` 中的 `handleSendCommand()` 和 `handleQueryDevice()` 目前是 TODO 状态
- ⚠️ 需要创建前端 API 服务文件（如 `src/api/radar/radarDeviceApi.ts`）用于安装配置
- ⚠️ 需要实现实时数据接收（当前 `GetRealtimeData` 返回空数据）

**注意**：
- 雷达监控设置（工作模式、跌倒、呼吸心率）已有 API：`src/api/settings/settings.ts`
- 雷达安装配置（安装模式、高度、边界）的 API 需要新增：`src/api/radar/radarDeviceApi.ts`

### 8.2 需要实现的 API 调用

#### 8.2.1 创建前端 API 服务

**需要创建新文件**: `owlFront/src/api/radar/radarDeviceApi.ts`

**注意**：与现有的 `src/api/settings/settings.ts`（监控设置）不同，这是用于安装配置的 API。

```typescript
/**
 * Radar Device Installation Config API
 * For radar device installation configuration (install_model, height, boundary)
 * 
 * Note: This is different from radar monitor settings API (work mode, fall, vital)
 * Monitor settings: /settings/api/v1/monitor/radar/:deviceId
 * Installation config: /radar-device/api/v1/radar-device/device/:id/config
 */

import { defHttp } from '@/utils/http/axios'
import type { ErrorMessageMode } from '/#/axios'

enum Api {
  UpdateRadarConfig = '/radar-device/api/v1/radar-device/device/:id/config',
  GetRadarOriginalProperties = '/radar-device/api/v1/radar-device/device/:id/original-properties',
  GetRadarRealtimeData = '/radar-device/api/v1/radar-device/device/:id/realtime',
}

// 雷达安装配置接口
export interface RadarInstallConfig {
  install_model?: 'wall' | 'ceiling' | 'corn'
  height?: number  // cm
  boundary_left?: number  // cm
  boundary_right?: number  // cm
  boundary_front?: number  // cm
  boundary_rear?: number  // cm
}

// 更新雷达安装配置
export function updateRadarConfigApi(
  deviceId: string,
  config: RadarInstallConfig,
  mode: ErrorMessageMode = 'modal'
) {
  return defHttp.put<string[]>({
    url: Api.UpdateRadarConfig.replace(':id', deviceId),
    data: config,
  }, { errorMessageMode: mode })
}

// 获取雷达原始属性（返回 JSON 字符串）
export function getRadarOriginalPropertiesApi(
  deviceId: string,
  mode: ErrorMessageMode = 'modal'
) {
  return defHttp.get<string>({
    url: Api.GetRadarOriginalProperties.replace(':id', deviceId),
  }, { errorMessageMode: mode })
}

// 获取实时数据
export function getRadarRealtimeDataApi(
  deviceId: string,
  mode: ErrorMessageMode = 'modal'
) {
  return defHttp.get<{
    positions: Array<{ x: number; y: number; timestamp: number }>
    vital: { heart_rate?: number; breath_rate?: number; timestamp: number } | null
  }>({
    url: Api.GetRadarRealtimeData.replace(':id', deviceId),
  }, { errorMessageMode: mode })
}
```

#### 8.2.2 更新 RadarTrajectory.vue

**文件**: `owlFront/src/views/monitoring/radar-trajectory/RadarTrajectory.vue`

```typescript
import { 
  updateRadarConfigApi, 
  getRadarOriginalPropertiesApi 
} from '@/api/radar/radarDeviceApi'

// 发送命令回调（供 vue-radar 使用）
async function handleSendCommand(
  deviceId: string,
  commandData: Record<string, any>
): Promise<{
  success: boolean
  data?: Record<string, any>
  error?: string
}> {
  try {
    // 调用后端 API 发送雷达配置
    await updateRadarConfigApi(deviceId, commandData)
    
    console.log('✅ Radar config updated:', { deviceId, commandData })
    
    return {
      success: true,
      data: commandData,
    }
  } catch (error: any) {
    console.error('❌ Send command failed:', error)
    return {
      success: false,
      error: error.message || 'Failed to send command',
    }
  }
}

// 查询设备回调（供 vue-radar 使用）
async function handleQueryDevice(deviceId: string): Promise<{
  success: boolean
  data?: Record<string, any>
  error?: string
}> {
  try {
    // 调用后端 API 查询雷达配置
    const propertiesJSON = await getRadarOriginalPropertiesApi(deviceId)
    
    // 解析 JSON 字符串
    const properties = JSON.parse(propertiesJSON)
    
    // 转换为前端格式（需要根据实际返回格式调整）
    const config = {
      install_model: properties.radar_install_style === '0' ? 'ceiling' : 
                     properties.radar_install_style === '1' ? 'wall' : 'corn',
      height: parseInt(properties.radar_install_height || '17') * 10, // dm -> cm
      // TODO: 解析 rectangle 格式的边界
      boundary_left: 300,
      boundary_right: 300,
      boundary_front: 400,
      boundary_rear: 0,
    }
    
    console.log('✅ Radar config queried:', { deviceId, config })
    
    return {
      success: true,
      data: config,
    }
  } catch (error: any) {
    console.error('❌ Query device failed:', error)
    return {
      success: false,
      error: error.message || 'Failed to query device',
    }
  }
}
```

### 8.3 实时数据接收

**当前状态**: `GetRealtimeData` API 返回空数据，需要实现从 Redis Streams 读取。

**实现建议**:
1. 在 `radar_handler.go` 的 `GetRealtimeData` 中实现从 Redis Streams 读取
2. 或使用 WebSocket 推送实时数据（推荐）
3. 前端通过 WebSocket 或轮询获取数据

### 8.4 整合对比

| 项目 | 原始 vue-radar | 整合后 (owlFront) |
|------|---------------|-------------------|
| 应用结构 | 独立 Vue 应用 | 整合进 owlFront 项目 |
| 组件位置 | `vue_radar/src/components/` | `owlFront/src/components/Radar/` |
| 页面路由 | 独立路由 | `owlFront/src/views/monitoring/radar-trajectory/` |
| API 调用 | 直接调用（Mock） | 通过回调 → owlFront API → 后端 |
| 状态管理 | 独立 Pinia Stores | 整合进 owlFront Stores |
| 外部集成 | 通过 URL 参数 | 通过 provide/inject 回调 |

### 8.5 参考文档

- 原始 vue-radar 架构: `vue_radar/ARCHITECTURE_DIAGRAM.md`
- 原始 vue-radar API 集成: `vue_radar/API_INTEGRATION.md`
- 后端 API 实现: 本文档第 1-3 节
