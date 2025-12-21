# AlarmCloud Handler 重构分析

## 📋 第一步：当前 Handler 业务功能点分析

### 1.1 Handler 基本信息

```
Handler 名称：AdminAlarm (alarm-cloud 部分)
文件路径：internal/http/admin_alarm_handlers.go
当前行数：~110 行（alarm-cloud 部分）
业务领域：告警配置管理
```

### 1.2 业务功能点列表

| 功能点 | HTTP 方法 | 路径 | 功能描述 | 复杂度 | 当前实现行数 |
|--------|----------|------|----------|--------|-------------|
| 查询告警配置 | GET | `/admin/api/v1/alarm-cloud` | 获取租户的告警策略配置，支持系统默认配置回退 | 中 | ~60 |
| 更新告警配置 | PUT | `/admin/api/v1/alarm-cloud` | 创建或更新租户的告警策略配置（UPSERT） | 中 | ~50 |

**总计**：2 个功能点，~110 行代码

### 1.3 业务规则分析

**权限检查**：
- ✅ 查询告警配置：需要 R 权限（查看配置）
- ✅ 更新告警配置：需要 U 权限（编辑配置）
- ⚠️ 当前实现：没有明确的权限检查（在 StubHandler 中）

**业务规则验证**：
- ✅ 租户ID验证（不能为空）
- ✅ 配置数据格式验证（JSONB 字段）
- ✅ 系统默认配置回退（如果租户没有配置，使用系统默认）
- ✅ UPSERT 语义（INSERT ... ON CONFLICT DO UPDATE）

**数据转换**：
- ✅ 前端格式 ↔ 领域模型（AlarmCloud）
- ✅ JSONB 字段处理（device_alarms, conditions, notification_rules, metadata）
- ✅ NULL 值处理（OfflineAlarm, LowBattery, DeviceFailure）

**业务编排**：
- ✅ 查询时优先使用租户配置，如果没有则回退到系统默认配置
- ✅ 更新时使用 UPSERT 语义

---

## 📋 第二步：Service 方法拆解

### 2.1 Service 接口设计

```go
type AlarmCloudService interface {
    // 查询
    GetAlarmCloudConfig(ctx context.Context, req GetAlarmCloudConfigRequest) (*AlarmCloudConfigResponse, error)
    
    // 更新
    UpdateAlarmCloudConfig(ctx context.Context, req UpdateAlarmCloudConfigRequest) (*AlarmCloudConfigResponse, error)
}
```

### 2.2 Service 方法详细设计

| Service 方法 | 对应 Handler 功能点 | 职责 | 复杂度 |
|-------------|-------------------|------|--------|
| `GetAlarmCloudConfig` | 查询告警配置 | 权限检查、业务规则验证、数据转换、调用 Repository（支持系统默认回退） | 中 |
| `UpdateAlarmCloudConfig` | 更新告警配置 | 权限检查、业务规则验证、数据转换、调用 Repository（UPSERT） | 中 |

### 2.3 Service 请求/响应结构

```go
// GetAlarmCloudConfigRequest 查询告警配置请求
type GetAlarmCloudConfigRequest struct {
    TenantID string
    UserID   string  // 当前用户ID（用于权限检查）
    UserRole string  // 当前用户角色（用于权限检查）
}

// UpdateAlarmCloudConfigRequest 更新告警配置请求
type UpdateAlarmCloudConfigRequest struct {
    TenantID          string
    UserID            string  // 当前用户ID（用于权限检查）
    UserRole          string  // 当前用户角色（用于权限检查）
    OfflineAlarm      *string // 可选
    LowBattery        *string // 可选
    DeviceFailure      *string // 可选
    DeviceAlarms      json.RawMessage // 可选
    Conditions         json.RawMessage // 可选
    NotificationRules  json.RawMessage // 可选
    Metadata           json.RawMessage // 可选
}

// AlarmCloudConfigResponse 告警配置响应
type AlarmCloudConfigResponse struct {
    TenantID          string          `json:"tenant_id"`
    OfflineAlarm      *string         `json:"OfflineAlarm,omitempty"`
    LowBattery        *string         `json:"LowBattery,omitempty"`
    DeviceFailure      *string         `json:"DeviceFailure,omitempty"`
    DeviceAlarms      json.RawMessage `json:"device_alarms"`
    Conditions         json.RawMessage `json:"conditions,omitempty"`
    NotificationRules  json.RawMessage `json:"notification_rules,omitempty"`
    Metadata           json.RawMessage `json:"metadata,omitempty"`
}
```

---

## 📋 第三步：Handler 方法拆解

### 3.1 Handler 结构设计

```go
type AlarmCloudHandler struct {
    alarmCloudService *service.AlarmCloudService
    logger            *zap.Logger
}

func (h *AlarmCloudHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 路由分发
}
```

### 3.2 Handler 方法详细设计

| Handler 方法 | 对应 Service 方法 | 职责 | 复杂度 |
|------------|------------------|------|--------|
| `GetAlarmCloudConfig` | `AlarmCloudService.GetAlarmCloudConfig` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `UpdateAlarmCloudConfig` | `AlarmCloudService.UpdateAlarmCloudConfig` | HTTP 参数解析、调用 Service、返回响应 | 低 |

---

## 📋 第四步：职责边界确认

### 4.1 Handler 职责

**只负责**：
- ✅ HTTP 请求/响应处理
- ✅ 参数解析和验证（HTTP 层面：类型、格式）
- ✅ 调用 Service
- ✅ 错误处理和日志记录

### 4.2 Service 职责

**负责**：
- ✅ 权限检查（基于 role_permissions 表）
- ✅ 业务规则验证（租户ID验证、配置数据格式验证）
- ✅ 数据转换（前端格式 ↔ 领域模型）
- ✅ 业务编排（系统默认配置回退）
- ✅ 调用 Repository

### 4.3 Repository 职责

**负责**：
- ✅ 数据访问（CRUD 操作）
- ✅ 数据完整性验证（外键、唯一性约束等）
- ✅ SQL 查询优化

---

## 📋 第五步：重构计划

### 5.1 实施步骤

1. **创建 Service 接口和实现**
   - [ ] 定义 Service 接口
   - [ ] 实现所有 Service 方法
   - [ ] 编写 Service 单元测试

2. **创建 Handler**
   - [ ] 定义 Handler 结构
   - [ ] 实现所有 Handler 方法
   - [ ] 编写 Handler 单元测试

3. **集成测试**
   - [ ] 编写 Service + Repository 集成测试
   - [ ] 运行所有测试

4. **路由注册**
   - [ ] 在 `router.go` 中添加注册方法
   - [ ] 在 `main.go` 中集成 Service 和 Handler

5. **验证和清理**
   - [ ] 手动测试 API 端点
   - [ ] 前端功能验证

### 5.2 预估工作量

| 任务 | 预估时间 | 优先级 |
|------|---------|--------|
| Service 实现 | 2-3 小时 | 高 |
| Handler 实现 | 1-2 小时 | 高 |
| 测试编写 | 2-3 小时 | 高 |
| 集成和验证 | 1-2 小时 | 中 |
| **总计** | **6-10 小时** | |

---

## 📋 检查清单

### 分析阶段

- [x] 列出所有业务功能点
- [x] 分析每个功能点的复杂度
- [x] 识别业务规则和权限检查
- [x] 拆解为 Service 方法
- [x] 拆解为 Handler 方法
- [x] 确认职责边界
- [x] 设计请求/响应结构
- [x] 制定重构计划

### 实施阶段

- [ ] Service 接口定义
- [ ] Service 实现
- [ ] Service 测试
- [ ] Handler 实现
- [ ] Handler 测试
- [ ] 集成测试
- [ ] 路由注册
- [ ] 功能验证

