# 雷达和睡眠板服务分析（v1.0 vs v1.5）

## 📋 概述

本文档分析 v1.0（wisefido-backend）和 v1.5（owlBack）中雷达和睡眠板提供的服务，以及是否需要 Service 层。

---

## 🔍 v1.0 架构（wisefido-backend）

### 服务列表

1. **wisefido-radar** - 雷达服务（HTTP API）
   - 提供 RESTful API 接口
   - 处理雷达设备数据
   - 直接写入 MySQL 数据库

2. **wisefido-sleepace** - 睡眠垫服务（HTTP API）
   - 提供 RESTful API 接口
   - 订阅 Sleepace 厂家 MQTT
   - 处理睡眠垫数据
   - 直接写入 MySQL 数据库

### 主要功能

#### 1. 实时轨迹（Radar Realtime Trajectory）

**API 端点**（推测）：
- GET `/radar/api/v1/realtime/:deviceId` - 获取实时轨迹数据
- GET `/radar/api/v1/trajectory/:deviceId` - 获取轨迹历史数据

**数据流**：
```
Radar 设备 → MQTT Broker
    ↓
wisefido-radar 服务（HTTP API）
    ├─ MQTT 订阅
    ├─ 数据处理
    └─→ MySQL 数据库（radar_realtime_record, radar_trajectory 等表）
```

**架构**：
- **Handler** → **Repository** → **MySQL**
- ❓ **是否有 Service 层**：未知（需要查看 v1.0 代码）

#### 2. 睡眠报告（Sleepace Reports）

**API 端点**（v1.5 中已定义）：
- GET `/sleepace/api/v1/sleepace/reports/:id` - 获取睡眠报告列表
- GET `/sleepace/api/v1/sleepace/reports/:id/detail` - 获取睡眠报告详情
- GET `/sleepace/api/v1/sleepace/reports/:id/dates` - 获取有数据的日期列表

**数据流**：
```
Sleepace 设备 → Sleepace 厂家服务（第三方，独立 DB + HTTP API + MQTT）
    ↓
Sleepace 厂家 MQTT Broker
    ↓
wisefido-sleepace 服务（HTTP API）
    ├─ MQTT 订阅（厂家 MQTT）
    ├─ 数据处理
    └─→ MySQL 数据库（sleepace_report, sleepace_realtime_record 等表）
```

**架构**：
- **Handler** → **Repository** → **MySQL**
- ❓ **是否有 Service 层**：未知（需要查看 v1.0 代码）

#### 3. 设备监控配置（Settings）

**API 端点**（v1.5 中已定义）：
- GET `/settings/api/v1/monitor/sleepace/:deviceId` - 获取 Sleepace 监控配置
- PUT `/settings/api/v1/monitor/sleepace/:deviceId` - 更新 Sleepace 监控配置
- GET `/settings/api/v1/monitor/radar/:deviceId` - 获取 Radar 监控配置
- PUT `/settings/api/v1/monitor/radar/:deviceId` - 更新 Radar 监控配置

**数据流**：
```
前端 → wisefido-radar/wisefido-sleepace 服务（HTTP API）
    ↓
MySQL 数据库（device_config, alarm_config 等表）
```

**架构**：
- **Handler** → **Repository** → **MySQL**
- ❓ **是否有 Service 层**：未知（需要查看 v1.0 代码）

---

## 🚀 v1.5 架构（owlBack）

### 服务列表

1. **wisefido-radar** - 雷达服务（后台服务）
   - MQTT 订阅（雷达设备数据）
   - 发布到 Redis Streams（`radar:data:stream`）
   - **不是 HTTP API 服务**

2. **wisefido-sleepace** - 睡眠垫服务（后台服务）
   - MQTT 订阅（Sleepace 厂家 MQTT）
   - 发布到 Redis Streams（`sleepace:data:stream`）
   - **不是 HTTP API 服务**

3. **wisefido-data** - API 服务（HTTP API）
   - 提供所有 RESTful API 接口
   - 包括雷达和睡眠板的 API

### 主要功能

#### 1. 实时轨迹（Radar Realtime Trajectory）

**数据流**：
```
Radar 设备 → MQTT Broker
    ↓
wisefido-radar 服务（后台服务）
    ├─ MQTT 订阅
    └─→ Redis Streams (radar:data:stream)
        ↓
wisefido-data-transformer 服务
    ├─ 消费 radar:data:stream
    ├─ 数据标准化（SNOMED CT映射）
    └─→ PostgreSQL TimescaleDB (iot_timeseries)
        ↓
wisefido-sensor-fusion 服务
    ├─ 消费标准化数据
    ├─ 多传感器融合
    └─→ Redis (vital-focus:card:{card_id}:realtime)
```

**API 端点**（当前状态）：
- ❌ **未实现** - 需要从 `iot_timeseries` 表查询实时轨迹数据
- 可能需要：
  - GET `/radar/api/v1/realtime/:deviceId` - 获取实时轨迹数据
  - GET `/radar/api/v1/trajectory/:deviceId` - 获取轨迹历史数据

**架构**：
- **Handler** → **Repository** → **PostgreSQL (iot_timeseries)**
- ⚠️ **当前状态**：未实现（StubHandler 占位）

#### 2. 睡眠报告（Sleepace Reports）

**数据流**：
```
Sleepace 设备 → Sleepace 厂家服务（第三方，独立 DB + HTTP API + MQTT）
    ↓
Sleepace 厂家 MQTT Broker
    ↓
wisefido-sleepace 服务（后台服务）
    ├─ MQTT 订阅（厂家 MQTT）
    └─→ Redis Streams (sleepace:data:stream)
        ↓
wisefido-data-transformer 服务
    ├─ 消费 sleepace:data:stream
    ├─ 数据标准化（SNOMED CT映射）
    └─→ PostgreSQL TimescaleDB (iot_timeseries)
```

**API 端点**（当前状态）：
- ✅ **已定义路由**：`/sleepace/api/v1/sleepace/reports/:id`
- ⚠️ **当前实现**：StubHandler（占位，返回空数据）
- 需要实现：
  - 从 `iot_timeseries` 表查询睡眠报告数据
  - 或从 Sleepace 厂家服务获取（如果厂家服务提供 HTTP API）

**架构**：
- **Handler** → **Repository** → **PostgreSQL (iot_timeseries)** 或 **Sleepace 厂家服务**
- ⚠️ **当前状态**：StubHandler（占位）

#### 3. 设备监控配置（Settings）

**API 端点**（当前状态）：
- ✅ **已定义路由**：
  - `/settings/api/v1/monitor/sleepace/:deviceId`
  - `/settings/api/v1/monitor/radar/:deviceId`
- ⚠️ **当前实现**：StubHandler（占位，返回默认配置）

**数据流**：
```
前端 → wisefido-data 服务（HTTP API）
    ↓
PostgreSQL 数据库（alarm_device 表）
```

**架构**：
- **Handler** → **Repository** → **PostgreSQL (alarm_device)**
- ⚠️ **当前状态**：StubHandler（占位）

---

## 📊 Service 层需求分析

### 1. 实时轨迹（Radar Realtime Trajectory）

**复杂度分析**：
- ✅ **权限检查**：中等（device_id 验证、tenant_id 过滤）
- ❌ **业务规则验证**：简单（无复杂业务规则）
- ✅ **数据转换**：中等（时间序列数据格式化、轨迹点聚合）
- ❌ **业务编排**：简单（无跨 Repository 操作）
- ✅ **Handler 复杂度**：中等（需要时间序列查询、数据聚合）

**结论**：⚠️ **可能需要 Service** - 取决于数据转换和聚合的复杂度

**建议**：
- 如果只是简单的数据库查询和格式化，可以不需要 Service
- 如果需要复杂的数据聚合（如轨迹点聚合、时间窗口计算），则需要 Service

### 2. 睡眠报告（Sleepace Reports）

**复杂度分析**：
- ✅ **权限检查**：中等（device_id 验证、tenant_id 过滤）
- ❌ **业务规则验证**：简单（无复杂业务规则）
- ✅ **数据转换**：复杂（从 `iot_timeseries` 表聚合生成睡眠报告，或调用 Sleepace 厂家服务）
- ✅ **业务编排**：复杂（可能需要调用外部服务、数据聚合）
- ✅ **Handler 复杂度**：复杂（需要时间序列数据聚合、报告生成）

**结论**：✅ **需要 Service** - SleepaceReportService

**原因**：
- 数据转换复杂（从时间序列数据聚合生成睡眠报告）
- 可能需要调用外部服务（Sleepace 厂家服务）
- 业务编排复杂（数据聚合、报告生成）

### 3. 设备监控配置（Settings）

**复杂度分析**：
- ✅ **权限检查**：中等（device_id 验证、tenant_id 过滤）
- ✅ **业务规则验证**：中等（配置参数验证、范围检查）
- ✅ **数据转换**：中等（前端格式 ↔ 数据库格式）
- ❌ **业务编排**：简单（无跨 Repository 操作）
- ✅ **Handler 复杂度**：中等（配置参数验证、数据转换）

**结论**：✅ **需要 Service** - DeviceMonitorSettingsService

**原因**：
- 业务规则验证（配置参数验证、范围检查）
- 数据转换（前端格式 ↔ 数据库格式）
- 可能需要同步更新到设备（如果设备支持远程配置）

---

## 🎯 最终结论

### v1.0 架构

**推测**：
- **Handler** → **Repository** → **MySQL**
- ❓ **是否有 Service 层**：未知（需要查看 v1.0 代码）
- 如果 Handler 逻辑简单，可能没有 Service 层
- 如果 Handler 逻辑复杂，可能有 Service 层

### v1.5 架构

**当前状态**：
- **Handler**（StubHandler 占位）→ **Repository**（未实现）→ **PostgreSQL**
- ❌ **没有 Service 层**（当前都是 StubHandler）

**建议实现**：

1. ✅ **SleepaceReportService** - 睡眠报告服务
   - 从 `iot_timeseries` 表聚合生成睡眠报告
   - 或调用 Sleepace 厂家服务获取报告
   - 数据转换和格式化

2. ✅ **DeviceMonitorSettingsService** - 设备监控配置服务
   - 配置参数验证
   - 数据转换（前端格式 ↔ 数据库格式）
   - 可能需要同步更新到设备

3. ⚠️ **RadarRealtimeService** - 雷达实时轨迹服务（可选）
   - 如果只是简单的数据库查询，可以不需要 Service
   - 如果需要复杂的数据聚合，则需要 Service

---

## 📋 实现优先级

### Phase 1: 高优先级
1. ✅ **DeviceMonitorSettingsService** - 设备监控配置（前端已使用）
2. ✅ **SleepaceReportService** - 睡眠报告（前端已使用）

### Phase 2: 中优先级
3. ⚠️ **RadarRealtimeService** - 雷达实时轨迹（如果前端需要）

---

## 📚 参考文档

- `docs/09_Sleepace_v1.0_Architecture_Analysis.md` - Sleepace v1.0 架构分析
- `docs/API_FRONTEND_BACKEND_MATRIX.md` - API 前后端一致性总表
- `wisefido-data/internal/http/admin_other_handlers.go` - StubHandler 实现

