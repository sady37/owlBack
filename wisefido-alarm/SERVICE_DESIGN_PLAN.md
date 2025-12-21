# Service 层设计计划

基于已有的 Repository 层，设计对应的 Service 层。

## 📊 Repository 分析

### 1. AlarmEventsRepository（报警事件仓库）

**功能**：
- ✅ 完整的 CRUD 操作（Create, Read, Update, Delete）
- ✅ 复杂查询（ListAlarmEvents 支持多条件过滤和分页）
- ✅ 状态管理（AcknowledgeAlarmEvent, UpdateAlarmEventOperation）
- ✅ 统计查询（CountAlarmEvents）
- ✅ 便捷查询方法（GetActiveAlarmEvents, GetAlarmEventsByDevice 等）

**特点**：
- 业务逻辑复杂
- 涉及状态转换（active → acknowledged）
- 涉及业务规则验证（只能确认 active 状态的报警）
- 需要权限检查（如需要）
- 需要数据转换（如需要）

**结论**：**需要 Service 层**（应用 Service 模式）

---

### 2. AlarmCloudRepository（报警策略仓库）

**功能**：
- ✅ 只读操作（GetAlarmCloudConfig, GetDeviceTypeAlarmConfig）
- ✅ 简单的查询逻辑（租户配置 → 系统默认配置）

**特点**：
- 业务逻辑简单
- 主要是数据查询
- 不需要状态管理
- 不需要权限检查（配置读取）

**结论**：**不需要 Service 层**（直接使用 Repository）

---

### 3. AlarmDeviceRepository（设备报警配置仓库）

**功能**：
- ✅ 只读操作（GetAlarmDeviceConfig, GetDeviceMonitorConfig, GetDeviceDefaultMonitorConfig）
- ✅ 简单的查询逻辑

**特点**：
- 业务逻辑简单
- 主要是数据查询
- 不需要状态管理
- 不需要权限检查（配置读取）

**结论**：**不需要 Service 层**（直接使用 Repository）

---

### 4. CardRepository（卡片仓库）

**功能**：
- ✅ 只读操作（GetCardByID, GetCardDevices, GetAllCards）
- ✅ 简单的查询逻辑

**特点**：
- 业务逻辑简单
- 主要是数据查询
- 用于报警评估（内部使用）

**结论**：**不需要 Service 层**（直接使用 Repository，内部使用）

---

### 5. DeviceRepository（设备仓库）

**功能**：
- ✅ 只读操作（GetDeviceBindingInfo, GetDevicesByRoom, GetDevicesByBed）
- ✅ 简单的查询逻辑

**特点**：
- 业务逻辑简单
- 主要是数据查询
- 用于报警评估（内部使用）

**结论**：**不需要 Service 层**（直接使用 Repository，内部使用）

---

### 6. RoomRepository（房间仓库）

**功能**：
- ✅ 只读操作（GetRoomInfo, IsBathroom, GetRoomByBedID）
- ✅ 简单的查询逻辑

**特点**：
- 业务逻辑简单
- 主要是数据查询
- 用于报警评估（内部使用）

**结论**：**不需要 Service 层**（直接使用 Repository，内部使用）

---

## 🎯 Service 层设计决策

### 关键区分：使用场景

**重要**：Service 层的设计取决于**使用场景**：

1. **HTTP API 场景**（有 Handler 层）：
   - 复杂领域：Handler → **Service** → Repository（必须有 Service）
   - 简单领域：Handler → Repository（可以跳过 Service）

2. **后台服务场景**（没有 Handler 层）：
   - 可以直接使用 Repository
   - Service 层可选（除非有复杂业务逻辑需要封装）

### 决策原则

根据 `ARCHITECTURE_DESIGN.md` 和 `SERVICE_DESIGN_PATTERNS.md`：

1. **HTTP API 场景**：
   - **复杂领域需要 Service**（应用 Service 模式）
     - 完整的 CRUD 操作
     - 复杂的业务逻辑
     - 需要权限检查
     - 需要状态管理
   - **简单领域可以不设 Service**（直接使用 Repository）
     - 只读操作
     - 业务逻辑简单
     - 不需要权限检查
     - 不需要状态管理

2. **后台服务场景**：
   - 可以直接使用 Repository
   - Service 层可选（除非有复杂业务逻辑需要封装）
   - 需要数据转换

### 设计决策表

#### 当前场景（后台服务，没有 HTTP API）

| Repository | 是否需要 Service | 原因 | 使用方式 |
|-----------|----------------|------|---------|
| **AlarmEventsRepository** | ❌ **不需要** | 后台服务，直接使用 Repository | 直接使用 Repository |
| AlarmCloudRepository | ❌ 不需要 | 后台服务，只读操作 | 直接使用 Repository |
| AlarmDeviceRepository | ❌ 不需要 | 后台服务，只读操作 | 直接使用 Repository |
| CardRepository | ❌ 不需要 | 后台服务，只读操作 | 直接使用 Repository |
| DeviceRepository | ❌ 不需要 | 后台服务，只读操作 | 直接使用 Repository |
| RoomRepository | ❌ 不需要 | 后台服务，只读操作 | 直接使用 Repository |

#### 未来场景（如果添加 HTTP API）

| Repository | 是否需要 Service | 原因 | Service 模式 |
|-----------|----------------|------|-------------|
| **AlarmEventsRepository** | ✅ **需要** | 完整的 CRUD、状态管理、业务规则验证 | **应用 Service** |
| AlarmCloudRepository | ❌ 不需要 | 只读操作，业务逻辑简单（简单领域） | Handler → Repository（跳过 Service） |
| AlarmDeviceRepository | ❌ 不需要 | 只读操作，业务逻辑简单（简单领域） | Handler → Repository（跳过 Service） |
| CardRepository | ❌ 不需要 | 只读操作，内部使用 | 后台服务直接使用 |
| DeviceRepository | ❌ 不需要 | 只读操作，内部使用 | 后台服务直接使用 |
| RoomRepository | ❌ 不需要 | 只读操作，内部使用 | 后台服务直接使用 |

---

## 🏗️ Service 层架构设计

### 1. AlarmEventService（报警事件服务）

**模式**：应用 Service（Application Service）

**职责**：
1. **业务规则验证**
   - 参数验证（tenant_id, event_id 必填）
   - 状态验证（只能确认 active 状态的报警）
   - 操作值验证（operation 必须是有效值）

2. **权限检查**（如需要）
   - 确认报警权限
   - 更新报警权限
   - 删除报警权限

3. **数据转换**（如需要）
   - JSON ↔ 领域模型（如需要）

4. **错误处理和日志记录**
   - 统一的错误处理
   - 详细的日志记录

**依赖**：
- `repository.AlarmEventsRepository`

**接口设计**：

```go
type AlarmEventService struct {
    alarmEventsRepo *repository.AlarmEventsRepository
    logger          *zap.Logger
}

// 查询相关
func (s *AlarmEventService) GetAlarmEvent(ctx, tenantID, eventID) (*models.AlarmEvent, error)
func (s *AlarmEventService) ListAlarmEvents(ctx, tenantID, filters, page, size) ([]*models.AlarmEvent, int, error)
func (s *AlarmEventService) CountAlarmEvents(ctx, tenantID, filters) (int, error)

// 状态管理
func (s *AlarmEventService) AcknowledgeAlarmEvent(ctx, tenantID, eventID, handlerID) error
func (s *AlarmEventService) UpdateAlarmEventOperation(ctx, tenantID, eventID, operation, handlerID, notes) error

// CRUD 操作
func (s *AlarmEventService) CreateAlarmEvent(ctx, tenantID, event) error
func (s *AlarmEventService) UpdateAlarmEvent(ctx, tenantID, eventID, updates) error
func (s *AlarmEventService) DeleteAlarmEvent(ctx, tenantID, eventID) error

// 便捷查询
func (s *AlarmEventService) GetActiveAlarmEvents(ctx, tenantID, filters, page, size) ([]*models.AlarmEvent, int, error)
func (s *AlarmEventService) GetAlarmEventsByDevice(ctx, tenantID, deviceID, filters, page, size) ([]*models.AlarmEvent, int, error)
func (s *AlarmEventService) GetAlarmEventsByCategory(ctx, tenantID, category, filters, page, size) ([]*models.AlarmEvent, int, error)
func (s *AlarmEventService) GetAlarmEventsByLevel(ctx, tenantID, alarmLevel, filters, page, size) ([]*models.AlarmEvent, int, error)
```

---

### 2. 其他 Repository 的使用方式

#### 2.1 后台服务中使用（Evaluator）

**方式**：直接使用 Repository

```go
// evaluator.go
type Evaluator struct {
    cardRepo        *repository.CardRepository
    deviceRepo      *repository.DeviceRepository
    roomRepo        *repository.RoomRepository
    alarmCloudRepo  *repository.AlarmCloudRepository
    alarmDeviceRepo *repository.AlarmDeviceRepository
    alarmEventsRepo *repository.AlarmEventsRepository
}

// 直接调用 Repository
card, err := e.cardRepo.GetCardByID(tenantID, cardID)
config, err := e.alarmCloudRepo.GetAlarmCloudConfig(ctx, tenantID)
```

**原因**：
- 后台服务不需要权限检查
- 业务逻辑简单（主要是数据查询）
- 不需要状态管理

#### 2.2 HTTP API 中使用（如需要）

**方式**：直接使用 Repository（只读操作）

```go
// handler.go
type AlarmConfigHandler struct {
    alarmCloudRepo  *repository.AlarmCloudRepository
    alarmDeviceRepo *repository.AlarmDeviceRepository
}

func (h *AlarmConfigHandler) GetAlarmConfig(w http.ResponseWriter, r *http.Request) {
    tenantID, _ := getTenantIDFromRequest(r)
    
    // 直接调用 Repository（只读操作，不需要 Service）
    config, err := h.alarmCloudRepo.GetAlarmCloudConfig(r.Context(), tenantID)
    // ...
}
```

**原因**：
- 只读操作，不需要业务规则验证
- 不需要状态管理
- 不需要权限检查（配置读取）

---

## 📋 Service 层实现计划

### 当前状态（后台服务）

**所有 Repository 都直接使用**，不需要 Service 层。

**原因**：
- 没有 HTTP API（没有 Handler 层）
- 后台服务可以直接使用 Repository
- 业务逻辑简单（主要是数据查询）

---

### 未来计划（如果添加 HTTP API）

#### Phase 1: AlarmEventService（必须实现）

**优先级**：最高

**原因**：
- 完整的 CRUD 操作
- 复杂的业务逻辑（状态管理、业务规则验证）
- 需要为 HTTP API 提供统一接口

**实现内容**：
1. ✅ 接口设计（已完成）
2. ✅ 实现查询相关方法（已完成）
3. ✅ 实现状态管理方法（已完成）
4. ✅ 实现 CRUD 方法（已完成）
5. ✅ 实现业务规则验证（已完成）
6. ⏳ 编写单元测试（待完成）

**文件**：
- `internal/service/alarm_event_service.go` ✅

**注意**：当前已实现，但**暂时不需要**（因为还没有 HTTP API）。如果将来添加 HTTP API，可以直接使用。

---

### Phase 2: 其他 Repository（不需要 Service）

**优先级**：低

**原因**：
- 只读操作，业务逻辑简单
- 主要用于后台服务（Evaluator）
- 不需要权限检查

**使用方式**：
- 后台服务：直接使用 Repository
- HTTP API（如需要）：直接使用 Repository

---

## 🎯 总结

### Service 层设计决策

#### 当前（后台服务）

1. **所有 Repository** ❌
   - **模式**：直接使用 Repository
   - **原因**：后台服务，没有 HTTP API，不需要 Service 层
   - **使用场景**：Evaluator 直接使用

#### 未来（如果添加 HTTP API）

1. **AlarmEventService** ✅
   - **模式**：应用 Service
   - **原因**：完整的 CRUD、状态管理、业务规则验证（复杂领域）
   - **状态**：已实现（待测试，但暂时不需要）
   - **架构**：Handler → AlarmEventService → AlarmEventsRepository

2. **其他 Repository** ❌
   - **模式**：直接使用 Repository（简单领域）
   - **原因**：只读操作，业务逻辑简单
   - **架构**：Handler → Repository（跳过 Service）

### 架构图

```
┌─────────────────────────────────────────────────────────┐
│ HTTP API（如需要）                                       │
│  - AlarmEventHandler → AlarmEventService                │
│  - AlarmConfigHandler → AlarmCloudRepository（直接）     │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ Service 层                                               │
│  - AlarmEventService（应用 Service）                     │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ Repository 层                                            │
│  - AlarmEventsRepository                                │
│  - AlarmCloudRepository（直接使用）                      │
│  - AlarmDeviceRepository（直接使用）                     │
│  - CardRepository（直接使用，内部）                      │
│  - DeviceRepository（直接使用，内部）                   │
│  - RoomRepository（直接使用，内部）                      │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ Database                                                 │
└─────────────────────────────────────────────────────────┘
```

---

## 📝 下一步

1. ✅ **AlarmEventService 实现**（已完成）
2. ⏳ **编写单元测试**（待完成）
3. ⏳ **集成到 HTTP Handler**（如需要）
4. ⏳ **添加权限检查**（如需要）

---

## 📚 参考文档

- `ARCHITECTURE_DESIGN.md` - 架构设计文档（wisefido-data）
- `SERVICE_DESIGN_PATTERNS.md` - Service 层设计规范和模式
- `SERVICE_LAYER_DESIGN.md` - Service 层设计文档
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结

