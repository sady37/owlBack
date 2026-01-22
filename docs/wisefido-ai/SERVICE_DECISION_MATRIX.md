# Service 层决策矩阵表

## 📊 决策矩阵

### 决策维度

1. **使用场景**：
   - HTTP API（有 Handler 层）
   - 后台服务（没有 Handler 层）

2. **业务复杂度**：
   - 复杂领域（完整的 CRUD、状态管理、业务规则验证）
   - 简单领域（只读操作、业务逻辑简单）

3. **功能特点**：
   - 需要权限检查
   - 需要状态管理
   - 需要业务规则验证
   - 需要数据转换

---

## 🎯 完整决策矩阵

| Repository              | 使用场景   | API 端点 | 业务复杂度 | CRUD      | 状态管理   | 权限检查   | 业务规则   | 复杂查询 | **是否需要 Service** | **原因**                                   |
|------------------------|-----------|---------|-----------|-----------|-----------|-----------|-----------|---------|---------------------|-------------------------------------------|
| **AlarmEventsRepository** | HTTP API  | GET /admin/api/v1/alarm-events | 复杂      | ❌ 只读    | ❌ 不需要   | ✅ 需要    | ❌ 不需要   | ✅ **复杂** | ✅ **需要**          | 需要权限过滤、复杂查询（多表JOIN）、数据转换 |
| **AlarmEventsRepository** | HTTP API  | PUT /admin/api/v1/alarm-events/:id/handle | 复杂      | ✅ 更新    | ✅ 需要    | ✅ 需要    | ✅ 需要    | ✅ **复杂** | ✅ **需要**          | 需要权限检查（Facility vs Home）、业务规则验证、状态管理、跨表查询 |
| **AlarmEventsRepository** | 后台服务   | -       | 复杂      | ✅ 完整    | ✅ 需要    | ❌ 不需要   | ✅ 需要    | ❌ 简单    | ❌ **不需要**        | 后台服务，直接使用 Repository              |
| **AlarmCloudRepository** | HTTP API  | GET /admin/api/v1/alarm-cloud | 中等      | ❌ 只读    | ❌ 不需要   | ✅ 需要    | ❌ 不需要   | ❌ 简单    | ✅ **需要**          | 需要权限检查、数据转换（JSONB）             |
| **AlarmCloudRepository** | HTTP API  | PUT /admin/api/v1/alarm-cloud | 高        | ✅ 更新    | ❌ 不需要   | ✅ 需要    | ✅ 需要    | ❌ 简单    | ✅ **需要**          | 需要权限检查、业务规则验证、数据转换（JSONB） |
| **AlarmCloudRepository** | 后台服务   | -       | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ 简单    | ❌ **不需要**        | 后台服务，直接使用 Repository              |
| AlarmDeviceRepository  | HTTP API  | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 只读操作，简单领域，可以跳过 Service        |
| AlarmDeviceRepository  | 后台服务   | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 后台服务，直接使用 Repository              |
| CardRepository         | HTTP API  | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 只读操作，简单领域，可以跳过 Service        |
| CardRepository         | 后台服务   | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 后台服务，直接使用 Repository              |
| DeviceRepository       | HTTP API  | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 只读操作，简单领域，可以跳过 Service        |
| DeviceRepository       | 后台服务   | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 后台服务，直接使用 Repository              |
| RoomRepository         | HTTP API  | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 只读操作，简单领域，可以跳过 Service        |
| RoomRepository         | 后台服务   | 简单      | ❌ 只读    | ❌ 不需要   | ❌ 不需要   | ❌ 不需要   | ❌ **不需要**        | 后台服务，直接使用 Repository              |

---

## 📋 简化决策矩阵（按 Repository）

### 当前场景（后台服务）

| Repository              | 是否需要 Service | 架构                    | 原因                        |
|------------------------|----------------|------------------------|---------------------------|
| AlarmEventsRepository  | ❌ **不需要**   | Evaluator → Repository | 后台服务，直接使用 Repository |
| AlarmCloudRepository   | ❌ **不需要**   | Evaluator → Repository | 后台服务，只读操作            |
| AlarmDeviceRepository  | ❌ **不需要**   | Evaluator → Repository | 后台服务，只读操作            |
| CardRepository         | ❌ **不需要**   | Evaluator → Repository | 后台服务，只读操作            |
| DeviceRepository       | ❌ **不需要**   | Evaluator → Repository | 后台服务，只读操作            |
| RoomRepository         | ❌ **不需要**   | Evaluator → Repository | 后台服务，只读操作            |

**结论**：当前场景下（后台服务），**所有 Repository 都不需要 Service 层**。

---

### HTTP API 场景（基于实际需求）

| Repository              | API 端点 | 是否需要 Service | 架构                              | 原因                                   |
|------------------------|---------|----------------|-----------------------------------|--------------------------------------|
| **AlarmEventsRepository** | GET /admin/api/v1/alarm-events | ✅ **需要**     | Handler → **Service** → Repository | 需要权限过滤、复杂查询（多表JOIN）、数据转换 |
| **AlarmEventsRepository** | PUT /admin/api/v1/alarm-events/:id/handle | ✅ **需要**     | Handler → **Service** → Repository | 需要权限检查（Facility vs Home）、业务规则验证、状态管理 |
| **AlarmCloudRepository** | GET /admin/api/v1/alarm-cloud | ✅ **需要**     | Handler → **Service** → Repository | 需要权限检查、数据转换（JSONB）             |
| **AlarmCloudRepository** | PUT /admin/api/v1/alarm-cloud | ✅ **需要**     | Handler → **Service** → Repository | 需要权限检查、业务规则验证、数据转换（JSONB） |
| AlarmDeviceRepository  | -       | ❌ **不需要**   | 后台服务直接使用                   | 后台服务使用，无 HTTP API                |
| CardRepository         | -       | ❌ **不需要**   | 后台服务直接使用                   | 后台服务使用，无 HTTP API                |
| DeviceRepository       | -       | ❌ **不需要**   | 后台服务直接使用                   | 后台服务使用，无 HTTP API                |
| RoomRepository         | -       | ❌ **不需要**   | 后台服务直接使用                   | 后台服务使用，无 HTTP API                |

**结论**：HTTP API 场景下，**AlarmEventsRepository 和 AlarmCloudRepository 都需要 Service 层**。

---

## 🔍 详细分析

### AlarmEventsRepository（报警事件仓库）

#### 功能特点
- ✅ 完整的 CRUD 操作（Create, Read, Update, Delete）
- ✅ 状态管理（active → acknowledged）
- ✅ 业务规则验证（只能确认 active 状态的报警）
- ✅ 复杂查询（多条件过滤、分页）
- ✅ 统计查询（CountAlarmEvents）

#### 决策

**当前（后台服务）**：
- ❌ **不需要 Service**
- 原因：后台服务，直接使用 Repository
- 架构：Evaluator → AlarmEventsRepository

**HTTP API 场景**：
- ✅ **需要 Service**
- 原因：需要权限检查、权限过滤、复杂查询、状态管理、业务规则验证
- 架构：Handler → AlarmEventService → AlarmEventsRepository
- API 端点：
  - GET /admin/api/v1/alarm-events - 需要权限过滤、复杂查询（多表JOIN）、数据转换
  - PUT /admin/api/v1/alarm-events/:id/handle - 需要权限检查（Facility vs Home）、业务规则验证、状态管理

---

### AlarmCloudRepository（报警策略仓库）

#### 功能特点
- ✅ GET /admin/api/v1/alarm-cloud：获取配置（需要权限检查）
- ✅ PUT /admin/api/v1/alarm-cloud：更新配置（需要权限检查、业务规则验证）
- ✅ 数据转换（JSONB 字段 ↔ 领域模型）
- ✅ 业务逻辑（租户配置 → 系统默认配置）

#### 决策

**当前（后台服务）**：
- ❌ **不需要 Service**
- 原因：后台服务，直接使用 Repository
- 架构：Evaluator → AlarmCloudRepository

**HTTP API 场景**：
- ✅ **需要 Service**
- 原因：需要权限检查（canEdit）、业务规则验证、数据转换
- 架构：Handler → AlarmCloudService → AlarmCloudRepository
- API 端点：
  - GET /admin/api/v1/alarm-cloud - 需要权限检查
  - PUT /admin/api/v1/alarm-cloud - 需要权限检查、业务规则验证

---

### AlarmDeviceRepository（设备报警配置仓库）

#### 功能特点
- ❌ 只读操作（GetAlarmDeviceConfig, GetDeviceMonitorConfig）
- ❌ 不需要状态管理
- ❌ 不需要权限检查（配置读取）
- ❌ 业务逻辑简单

#### 决策

**当前（后台服务）**：
- ❌ **不需要 Service**
- 原因：后台服务，只读操作
- 架构：Evaluator → AlarmDeviceRepository

**未来（HTTP API）**：
- ❌ **不需要 Service**
- 原因：简单领域，只读操作，可以跳过 Service
- 架构：Handler → AlarmDeviceRepository（跳过 Service）

---

### CardRepository（卡片仓库）

#### 功能特点
- ❌ 只读操作（GetCardByID, GetCardDevices, GetAllCards）
- ❌ 不需要状态管理
- ❌ 不需要权限检查（内部使用）
- ❌ 业务逻辑简单

#### 决策

**当前（后台服务）**：
- ❌ **不需要 Service**
- 原因：后台服务，只读操作，内部使用
- 架构：Evaluator → CardRepository

**未来（HTTP API）**：
- ❌ **不需要 Service**
- 原因：简单领域，只读操作，可以跳过 Service
- 架构：Handler → CardRepository（跳过 Service）

---

### DeviceRepository（设备仓库）

#### 功能特点
- ❌ 只读操作（GetDeviceBindingInfo, GetDevicesByRoom, GetDevicesByBed）
- ❌ 不需要状态管理
- ❌ 不需要权限检查（内部使用）
- ❌ 业务逻辑简单

#### 决策

**当前（后台服务）**：
- ❌ **不需要 Service**
- 原因：后台服务，只读操作，内部使用
- 架构：Evaluator → DeviceRepository

**未来（HTTP API）**：
- ❌ **不需要 Service**
- 原因：简单领域，只读操作，可以跳过 Service
- 架构：Handler → DeviceRepository（跳过 Service）

---

### RoomRepository（房间仓库）

#### 功能特点
- ❌ 只读操作（GetRoomInfo, IsBathroom, GetRoomByBedID）
- ❌ 不需要状态管理
- ❌ 不需要权限检查（内部使用）
- ❌ 业务逻辑简单

#### 决策

**当前（后台服务）**：
- ❌ **不需要 Service**
- 原因：后台服务，只读操作，内部使用
- 架构：Evaluator → RoomRepository

**未来（HTTP API）**：
- ❌ **不需要 Service**
- 原因：简单领域，只读操作，可以跳过 Service
- 架构：Handler → RoomRepository（跳过 Service）

---

## 📊 总结表

### 快速参考

| Repository              | 当前（后台服务） | HTTP API 场景 |
|------------------------|----------------|--------------|
| **AlarmEventsRepository** | ❌ 不需要      | ✅ **需要**（GET /admin/api/v1/alarm-events, PUT /admin/api/v1/alarm-events/:id/handle） |
| **AlarmCloudRepository** | ❌ 不需要      | ✅ **需要**（GET /admin/api/v1/alarm-cloud, PUT /admin/api/v1/alarm-cloud） |
| AlarmDeviceRepository  | ❌ 不需要      | ❌ 不需要（无 HTTP API） |
| CardRepository         | ❌ 不需要      | ❌ 不需要（无 HTTP API） |
| DeviceRepository       | ❌ 不需要      | ❌ 不需要（无 HTTP API） |
| RoomRepository         | ❌ 不需要      | ❌ 不需要（无 HTTP API） |

---

## 🎯 决策规则

### 规则 1: 使用场景

- **后台服务**：所有 Repository 都不需要 Service（直接使用）
- **HTTP API**：根据业务复杂度决定

### 规则 2: 业务复杂度

- **复杂领域**（HTTP API 场景）：
  - ✅ 完整的 CRUD 操作
  - ✅ 状态管理
  - ✅ 权限检查
  - ✅ 业务规则验证
  - **→ 需要 Service**

- **简单领域**（HTTP API 场景）：
  - ❌ 只读操作
  - ❌ 不需要状态管理
  - ❌ 不需要权限检查
  - ❌ 业务逻辑简单
  - **→ 可以跳过 Service**

### 规则 3: 功能特点

如果满足以下**任意一项**，在 HTTP API 场景下需要 Service：
- ✅ 完整的 CRUD 操作
- ✅ 状态管理
- ✅ 权限检查
- ✅ 复杂的业务规则验证

---

## 📝 实施建议

### 当前实施（后台服务）

**所有 Repository 直接使用**，不需要 Service 层。

```go
// Evaluator 直接使用 Repository
type Evaluator struct {
    alarmEventsRepo *repository.AlarmEventsRepository
    alarmCloudRepo  *repository.AlarmCloudRepository
    // ...
}
```

### 未来实施（HTTP API）

**只有 AlarmEventsRepository 需要 Service**。

```go
// Handler 使用 Service（复杂领域）
type AlarmEventHandler struct {
    service *service.AlarmEventService
}

// Handler 直接使用 Repository（简单领域）
type AlarmConfigHandler struct {
    alarmCloudRepo *repository.AlarmCloudRepository
}
```

---

## 📚 参考文档

- `ARCHITECTURE_DESIGN.md` - 架构设计文档（wisefido-data）
- `SERVICE_DESIGN_PATTERNS.md` - Service 层设计规范和模式
- `SERVICE_DESIGN_CLARIFICATION.md` - Service 层设计澄清
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结

