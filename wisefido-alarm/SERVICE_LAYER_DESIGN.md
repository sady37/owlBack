# Service 层设计文档

## 📋 设计原则

参考 `wisefido-data/ARCHITECTURE_DESIGN.md` 的分层原则，Service 层负责：

### 职责边界

#### Service 层职责
- ✅ **业务规则验证**（如 tenant_id 必填、operation 值验证）
- ✅ **数据转换**（JSON ↔ 领域模型，如需要）
- ✅ **业务编排**（协调多个 Repository，如需要）
- ✅ **事务管理**（跨 Repository 的事务，如需要）
- ✅ **权限检查**（如需要，调用 PermissionChecker）
- ✅ **错误处理和日志记录**

#### Service 层不负责
- ❌ HTTP 请求/响应处理（属于 Handler 层）
- ❌ 数据库 SQL 操作（属于 Repository 层）
- ❌ 数据一致性维护（属于 Repository 层）

### 依赖方向

```
Handler → Service → Repository → Database
```

**规则**：
- Service 只能调用 Repository
- Service 不能直接操作 Database
- **不允许反向依赖**

## 🏗️ 架构设计

### 当前 Service 层结构

```
internal/service/
├── alarm.go                    # 报警服务（整合各层，用于后台服务）
└── alarm_event_service.go      # 报警事件服务（业务逻辑层，用于 HTTP API）
```

### AlarmEventService 设计

**位置**：`internal/service/alarm_event_service.go`

**职责**：
- 提供报警事件的业务逻辑封装
- 业务规则验证
- 错误处理和日志记录
- 为 HTTP Handler 提供统一接口

**依赖**：
- `repository.AlarmEventsRepository` - 数据访问层

## 📊 功能接口

### 1. 查询相关方法

#### `ListAlarmEvents` - 查询报警事件列表
```go
func (s *AlarmEventService) ListAlarmEvents(
    ctx context.Context,
    tenantID string,
    filters repository.AlarmEventFilters,
    page, size int,
) ([]*models.AlarmEvent, int, error)
```

**业务规则**：
- `tenant_id` 必填
- `page` 和 `size` 必须 > 0
- `size` 最大为 100（防止过大查询）
- 默认 `size` 为 20

#### `GetAlarmEvent` - 获取单个报警事件
```go
func (s *AlarmEventService) GetAlarmEvent(
    ctx context.Context,
    tenantID, eventID string,
) (*models.AlarmEvent, error)
```

**业务规则**：
- `tenant_id` 和 `event_id` 必填
- 自动过滤软删除的记录

#### `CountAlarmEvents` - 统计报警事件数量
```go
func (s *AlarmEventService) CountAlarmEvents(
    ctx context.Context,
    tenantID string,
    filters repository.AlarmEventFilters,
) (int, error)
```

**业务规则**：
- `tenant_id` 必填

### 2. 状态管理方法

#### `AcknowledgeAlarmEvent` - 确认报警事件
```go
func (s *AlarmEventService) AcknowledgeAlarmEvent(
    ctx context.Context,
    tenantID, eventID, handlerID string,
) error
```

**业务规则**：
- `tenant_id`、`event_id`、`handler_id` 必填
- 只能确认状态为 `'active'` 的报警
- 自动设置 `hand_time` 为当前时间
- 自动更新 `alarm_status` 为 `'acknowledged'`

#### `UpdateAlarmEventOperation` - 更新报警事件操作结果
```go
func (s *AlarmEventService) UpdateAlarmEventOperation(
    ctx context.Context,
    tenantID, eventID, operation, handlerID string,
    notes *string,
) error
```

**业务规则**：
- `tenant_id`、`event_id`、`operation`、`handler_id` 必填
- `operation` 必须是有效值：`verified_and_processed`、`false_alarm`、`resolved`、`escalated`、`cancelled`
- 只能更新状态为 `'active'` 或 `'acknowledged'` 的报警
- 自动设置 `hand_time` 为当前时间

### 3. CRUD 方法

#### `CreateAlarmEvent` - 创建报警事件
```go
func (s *AlarmEventService) CreateAlarmEvent(
    ctx context.Context,
    tenantID string,
    event *models.AlarmEvent,
) error
```

**业务规则**：
- `tenant_id` 必填
- `event` 必填且 `tenant_id` 必须匹配
- `event_id` 必须已生成（由 Builder 生成）
- `triggered_at` 必须设置
- `alarm_status` 默认为 `'active'`

#### `UpdateAlarmEvent` - 更新报警事件（部分更新）
```go
func (s *AlarmEventService) UpdateAlarmEvent(
    ctx context.Context,
    tenantID, eventID string,
    updates map[string]interface{},
) error
```

**业务规则**：
- `tenant_id` 和 `event_id` 必填
- `updates` 不能为空
- 只能更新允许的字段（当前只允许 `notes`）
- 不能更新 `event_id`、`tenant_id`、`device_id`、`created_at`
- `alarm_status`、`handler`、`operation`、`hand_time` 应该通过专门的方法更新

#### `DeleteAlarmEvent` - 删除报警事件（软删除）
```go
func (s *AlarmEventService) DeleteAlarmEvent(
    ctx context.Context,
    tenantID, eventID string,
) error
```

**业务规则**：
- `tenant_id` 和 `event_id` 必填
- 软删除（设置 `metadata->>'deleted_at'`）

### 4. 便捷查询方法

#### `GetActiveAlarmEvents` - 获取活跃的报警事件
```go
func (s *AlarmEventService) GetActiveAlarmEvents(
    ctx context.Context,
    tenantID string,
    filters repository.AlarmEventFilters,
    page, size int,
) ([]*models.AlarmEvent, int, error)
```

#### `GetAlarmEventsByDevice` - 根据设备ID获取报警事件
```go
func (s *AlarmEventService) GetAlarmEventsByDevice(
    ctx context.Context,
    tenantID, deviceID string,
    filters repository.AlarmEventFilters,
    page, size int,
) ([]*models.AlarmEvent, int, error)
```

#### `GetAlarmEventsByCategory` - 根据分类获取报警事件
```go
func (s *AlarmEventService) GetAlarmEventsByCategory(
    ctx context.Context,
    tenantID, category string,
    filters repository.AlarmEventFilters,
    page, size int,
) ([]*models.AlarmEvent, int, error)
```

#### `GetAlarmEventsByLevel` - 根据报警级别获取报警事件
```go
func (s *AlarmEventService) GetAlarmEventsByLevel(
    ctx context.Context,
    tenantID, alarmLevel string,
    filters repository.AlarmEventFilters,
    page, size int,
) ([]*models.AlarmEvent, int, error)
```

## 🔍 使用示例

### 在 HTTP Handler 中使用

```go
// Handler 层
type AlarmEventHandler struct {
    service *service.AlarmEventService
}

func (h *AlarmEventHandler) ListAlarms(w http.ResponseWriter, r *http.Request) {
    // 1. 解析请求参数
    tenantID, _ := getTenantIDFromRequest(r)
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    size, _ := strconv.Atoi(r.URL.Query().Get("size"))
    
    // 2. 构建过滤条件
    filters := repository.AlarmEventFilters{}
    if deviceID := r.URL.Query().Get("device_id"); deviceID != "" {
        filters.DeviceID = &deviceID
    }
    
    // 3. 调用 Service
    events, total, err := h.service.ListAlarmEvents(
        r.Context(),
        tenantID,
        filters,
        page,
        size,
    )
    
    // 4. 处理错误
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, Fail(err.Error()))
        return
    }
    
    // 5. 返回响应
    writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
        "events": events,
        "total": total,
        "page": page,
        "size": size,
    }))
}

func (h *AlarmEventHandler) AcknowledgeAlarm(w http.ResponseWriter, r *http.Request) {
    // 1. 解析请求参数
    tenantID, _ := getTenantIDFromRequest(r)
    eventID := extractEventIDFromPath(r.URL.Path)
    
    var payload map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        writeJSON(w, http.StatusBadRequest, Fail("invalid body"))
        return
    }
    
    handlerID, _ := payload["handler_id"].(string)
    
    // 2. 调用 Service
    if err := h.service.AcknowledgeAlarmEvent(
        r.Context(),
        tenantID,
        eventID,
        handlerID,
    ); err != nil {
        writeJSON(w, http.StatusBadRequest, Fail(err.Error()))
        return
    }
    
    // 3. 返回响应
    writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
        "success": true,
    }))
}
```

### 在后台服务中使用

```go
// 后台服务中，通常直接使用 Repository
// 但如果需要业务规则验证，也可以使用 Service

// 方式1：直接使用 Repository（当前方式）
err := alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, event)

// 方式2：使用 Service（如果需要业务规则验证）
err := alarmEventService.CreateAlarmEvent(ctx, tenantID, event)
```

## 📝 业务规则总结

### 通用规则
1. **tenant_id 验证**：所有方法都验证 `tenant_id` 必填
2. **自动过滤软删除**：所有查询方法自动过滤软删除的记录
3. **错误处理**：所有方法都记录错误日志并返回明确的错误信息

### 状态管理规则
1. **确认报警**：只能确认状态为 `'active'` 的报警
2. **更新操作**：只能更新状态为 `'active'` 或 `'acknowledged'` 的报警
3. **操作值验证**：`operation` 必须是预定义的有效值

### 更新规则
1. **字段限制**：只能更新允许的字段
2. **保护字段**：不能更新 `event_id`、`tenant_id`、`device_id`、`created_at`
3. **状态字段**：`alarm_status`、`handler`、`operation`、`hand_time` 应该通过专门的方法更新

## 🚀 下一步

### 待实现功能
1. **权限检查**：如果需要，添加 `PermissionChecker` 集成
2. **事务管理**：如果需要跨 Repository 的事务，添加事务支持
3. **数据转换**：如果需要 JSON ↔ 领域模型转换，添加转换逻辑
4. **单元测试**：编写 Service 层的单元测试

### 集成到 HTTP Handler
1. 创建 `AlarmEventHandler`
2. 实现 HTTP 路由
3. 集成到主服务

## 📚 相关文档

- `ARCHITECTURE_DESIGN.md` - 架构设计文档（wisefido-data）
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结
- `ALARM_EVENT_WRITE.md` - 报警事件写入说明

