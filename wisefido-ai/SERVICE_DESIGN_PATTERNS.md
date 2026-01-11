# Service 层设计规范和模式

## 📋 目录

1. [Service 层的设计模式](#service-层的设计模式)
2. [设计规范](#设计规范)
3. [确认流程](#确认流程)
4. [最佳实践](#最佳实践)
5. [示例对比](#示例对比)

---

## Service 层的设计模式

### 模式 1: 简单 Service（Simple Service）

**适用场景**：
- 业务逻辑简单，主要是 CRUD 操作
- 不需要跨 Repository 协调
- 不需要复杂的事务管理

**特点**：
- Service 直接调用单个 Repository
- 主要做参数验证和错误处理
- 不涉及复杂的业务编排

**示例**：
```go
type AlarmEventService struct {
    alarmEventsRepo *repository.AlarmEventsRepository
    logger          *zap.Logger
}

func (s *AlarmEventService) GetAlarmEvent(ctx context.Context, tenantID, eventID string) (*models.AlarmEvent, error) {
    // 1. 参数验证
    if tenantID == "" {
        return nil, fmt.Errorf("tenant_id is required")
    }
    
    // 2. 调用 Repository
    return s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
}
```

### 模式 2: 编排 Service（Orchestration Service）

**适用场景**：
- 需要协调多个 Repository
- 涉及多个实体的操作
- 需要保证数据一致性

**特点**：
- Service 协调多个 Repository
- 可能涉及事务管理
- 包含复杂的业务逻辑

**示例**：
```go
type ResidentService struct {
    residentsRepo *repository.ResidentsRepository
    phiRepo       *repository.ResidentPHIRepository
    contactsRepo  *repository.ResidentContactsRepository
    db            *sql.DB
    logger        *zap.Logger
}

func (s *ResidentService) CreateResidentWithPHI(
    ctx context.Context,
    tenantID string,
    resident *domain.Resident,
    phi *domain.ResidentPHI,
) (string, error) {
    // 1. 参数验证
    if tenantID == "" {
        return "", fmt.Errorf("tenant_id is required")
    }
    
    // 2. 开始事务
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return "", fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback()
    
    // 3. 创建 Resident
    residentID, err := s.residentsRepo.CreateResident(ctx, tx, tenantID, resident)
    if err != nil {
        return "", fmt.Errorf("failed to create resident: %w", err)
    }
    
    // 4. 创建 PHI（如果提供）
    if phi != nil {
        phi.ResidentID = residentID
        if err := s.phiRepo.CreatePHI(ctx, tx, tenantID, phi); err != nil {
            return "", fmt.Errorf("failed to create PHI: %w", err)
        }
    }
    
    // 5. 提交事务
    if err := tx.Commit(); err != nil {
        return "", fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return residentID, nil
}
```

### 模式 3: 门面 Service（Facade Service）

**适用场景**：
- 需要为复杂的子系统提供简单接口
- 隐藏多个 Repository 的复杂性
- 提供统一的业务接口

**特点**：
- 封装多个 Repository 的调用
- 提供高级业务接口
- 简化 Handler 层的调用

**示例**：
```go
type AlarmManagementService struct {
    alarmEventsRepo  *repository.AlarmEventsRepository
    alarmCloudRepo   *repository.AlarmCloudRepository
    alarmDeviceRepo  *repository.AlarmDeviceRepository
    deviceRepo       *repository.DeviceRepository
    logger           *zap.Logger
}

// GetAlarmEventWithConfig 获取报警事件及其配置
func (s *AlarmManagementService) GetAlarmEventWithConfig(
    ctx context.Context,
    tenantID, eventID string,
) (*AlarmEventWithConfig, error) {
    // 1. 获取报警事件
    event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
    if err != nil {
        return nil, err
    }
    
    // 2. 获取设备配置
    deviceConfig, err := s.alarmDeviceRepo.GetAlarmDeviceConfig(ctx, tenantID, event.DeviceID)
    if err != nil {
        return nil, err
    }
    
    // 3. 获取报警策略
    cloudConfig, err := s.alarmCloudRepo.GetAlarmCloudConfig(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    
    // 4. 组合返回
    return &AlarmEventWithConfig{
        Event:       event,
        DeviceConfig: deviceConfig,
        CloudConfig: cloudConfig,
    }, nil
}
```

### 模式 4: 应用 Service（Application Service）

**适用场景**：
- 实现完整的业务流程
- 包含权限检查
- 包含业务规则验证
- 包含数据转换

**特点**：
- 完整的业务逻辑封装
- 权限检查集成
- 数据转换（JSON ↔ 领域模型）
- 错误处理和日志记录

**示例**：
```go
type ResidentService struct {
    residentsRepo     *repository.ResidentsRepository
    permissionChecker *PermissionChecker
    logger           *zap.Logger
}

func (s *ResidentService) CreateResident(
    ctx context.Context,
    tenantID string,
    userID, userRole string,
    payload map[string]interface{},
) (string, error) {
    // 1. 权限检查
    if !s.permissionChecker.CanCreateResident(ctx, tenantID, userID, userRole) {
        return "", ErrPermissionDenied
    }
    
    // 2. 业务规则验证
    if err := s.validateResidentPayload(payload); err != nil {
        return "", err
    }
    
    // 3. 数据转换
    resident, err := s.convertPayloadToResident(payload)
    if err != nil {
        return "", err
    }
    
    // 4. 调用 Repository
    residentID, err := s.residentsRepo.CreateResident(ctx, tenantID, resident)
    if err != nil {
        s.logger.Error("Failed to create resident",
            zap.String("tenant_id", tenantID),
            zap.Error(err),
        )
        return "", fmt.Errorf("failed to create resident: %w", err)
    }
    
    return residentID, nil
}
```

---

## 设计规范

### 规范 1: 接口设计

#### 1.1 方法命名规范

**查询方法**：
- `Get{Entity}` - 获取单个实体
- `List{Entities}` - 获取列表
- `Count{Entities}` - 统计数量
- `Get{Entity}By{Field}` - 根据字段查询

**操作方法**：
- `Create{Entity}` - 创建
- `Update{Entity}` - 更新
- `Delete{Entity}` - 删除
- `{Action}{Entity}` - 特定操作（如 `AcknowledgeAlarmEvent`）

#### 1.2 参数顺序规范

**标准顺序**：
1. `ctx context.Context` - 上下文（第一个参数）
2. `tenantID string` - 租户ID（第二个参数）
3. 其他业务参数
4. 可选参数（如 `filters`, `page`, `size`）

**示例**：
```go
// ✅ 正确
func (s *Service) ListAlarmEvents(
    ctx context.Context,
    tenantID string,
    filters AlarmEventFilters,
    page, size int,
) ([]*AlarmEvent, int, error)

// ❌ 错误
func (s *Service) ListAlarmEvents(
    tenantID string,
    ctx context.Context,  // context 应该在第一位
    filters AlarmEventFilters,
) ([]*AlarmEvent, error)
```

#### 1.3 返回值规范

**标准返回值**：
- 查询单个：`(*Entity, error)`
- 查询列表：`([]*Entity, int, error)` - 第二个返回值是总数
- 操作：`(result, error)` 或 `error`

**示例**：
```go
// ✅ 正确
func (s *Service) GetAlarmEvent(...) (*AlarmEvent, error)
func (s *Service) ListAlarmEvents(...) ([]*AlarmEvent, int, error)
func (s *Service) CreateAlarmEvent(...) error

// ❌ 错误
func (s *Service) ListAlarmEvents(...) ([]*AlarmEvent, error)  // 缺少总数
```

### 规范 2: 参数验证

#### 2.1 必填参数验证

**规则**：
- 所有必填参数必须在方法开始处验证
- 验证失败立即返回明确的错误信息
- 错误信息应该包含参数名

**示例**：
```go
func (s *Service) CreateAlarmEvent(
    ctx context.Context,
    tenantID string,
    event *models.AlarmEvent,
) error {
    // ✅ 参数验证
    if tenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    if event == nil {
        return fmt.Errorf("event is required")
    }
    if event.TenantID != tenantID {
        return fmt.Errorf("event tenant_id (%s) does not match provided tenant_id (%s)", 
            event.TenantID, tenantID)
    }
    
    // 继续业务逻辑...
}
```

#### 2.2 业务规则验证

**规则**：
- 业务规则验证应该在调用 Repository 之前进行
- 验证失败返回明确的业务错误
- 复杂的验证逻辑可以提取为独立方法

**示例**：
```go
func (s *Service) AcknowledgeAlarmEvent(
    ctx context.Context,
    tenantID, eventID, handlerID string,
) error {
    // 1. 参数验证
    if tenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    // ...
    
    // 2. 业务规则验证：先获取事件，检查状态
    event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
    if err != nil {
        return fmt.Errorf("failed to get alarm event: %w", err)
    }
    
    // 业务规则：只能确认状态为 'active' 的报警
    if event.AlarmStatus != "active" {
        return fmt.Errorf("can only acknowledge active alarms, current status: %s", 
            event.AlarmStatus)
    }
    
    // 3. 调用 Repository
    return s.alarmEventsRepo.AcknowledgeAlarmEvent(ctx, tenantID, eventID, handlerID)
}
```

### 规范 3: 错误处理

#### 3.1 错误信息规范

**规则**：
- 错误信息应该明确、可读
- 包含关键参数信息（如 tenant_id, event_id）
- 区分参数错误、业务规则错误、系统错误

**示例**：
```go
// ✅ 好的错误信息
return fmt.Errorf("tenant_id is required")
return fmt.Errorf("can only acknowledge active alarms, current status: %s", status)
return fmt.Errorf("failed to create alarm event: %w", err)

// ❌ 不好的错误信息
return fmt.Errorf("error")
return fmt.Errorf("invalid")
```

#### 3.2 日志记录规范

**规则**：
- 所有错误都应该记录日志
- 成功的重要操作也应该记录日志
- 日志应该包含关键参数（tenant_id, event_id 等）

**示例**：
```go
func (s *Service) CreateAlarmEvent(...) error {
    // ...
    if err := s.alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, event); err != nil {
        // ✅ 记录错误日志
        s.logger.Error("Failed to create alarm event",
            zap.String("tenant_id", tenantID),
            zap.String("event_id", event.EventID),
            zap.String("event_type", event.EventType),
            zap.Error(err),
        )
        return fmt.Errorf("failed to create alarm event: %w", err)
    }
    
    // ✅ 记录成功日志
    s.logger.Info("Alarm event created",
        zap.String("tenant_id", tenantID),
        zap.String("event_id", event.EventID),
        zap.String("event_type", event.EventType),
    )
    
    return nil
}
```

### 规范 4: 依赖注入

#### 4.1 构造函数规范

**规则**：
- 使用构造函数注入依赖
- 依赖应该通过接口（如需要）或具体类型注入
- 构造函数应该验证依赖不为 nil

**示例**：
```go
type AlarmEventService struct {
    alarmEventsRepo *repository.AlarmEventsRepository
    logger          *zap.Logger
}

func NewAlarmEventService(
    alarmEventsRepo *repository.AlarmEventsRepository,
    logger *zap.Logger,
) *AlarmEventService {
    if alarmEventsRepo == nil {
        panic("alarmEventsRepo is required")
    }
    if logger == nil {
        panic("logger is required")
    }
    
    return &AlarmEventService{
        alarmEventsRepo: alarmEventsRepo,
        logger:          logger,
    }
}
```

---

## 确认流程

### 流程 1: 需求分析

**步骤**：
1. **识别业务需求**
   - 需要哪些业务操作？
   - 涉及哪些实体？
   - 需要哪些业务规则？

2. **确定 Service 模式**
   - 简单 Service？
   - 编排 Service？
   - 门面 Service？
   - 应用 Service？

3. **识别依赖**
   - 需要哪些 Repository？
   - 需要哪些外部服务？
   - 需要权限检查吗？

**输出**：
- Service 接口定义（方法签名）
- 依赖列表
- 业务规则列表

**确认点**：
- [ ] 业务需求是否清晰？
- [ ] Service 模式是否合适？
- [ ] 依赖是否完整？

### 流程 2: 接口设计

**步骤**：
1. **定义方法签名**
   - 方法名是否符合规范？
   - 参数顺序是否正确？
   - 返回值是否合适？

2. **定义业务规则**
   - 参数验证规则
   - 业务逻辑规则
   - 状态转换规则

3. **定义错误处理**
   - 错误类型
   - 错误信息格式
   - 日志记录点

**输出**：
- 完整的方法签名
- 业务规则文档
- 错误处理规范

**确认点**：
- [ ] 方法签名是否符合规范？
- [ ] 业务规则是否完整？
- [ ] 错误处理是否清晰？

### 流程 3: 实现设计

**步骤**：
1. **实现参数验证**
   - 所有必填参数验证
   - 参数格式验证
   - 参数范围验证

2. **实现业务规则验证**
   - 业务状态检查
   - 业务规则验证
   - 权限检查（如需要）

3. **实现业务逻辑**
   - 调用 Repository
   - 事务管理（如需要）
   - 数据转换（如需要）

4. **实现错误处理**
   - 错误日志记录
   - 错误信息返回
   - 成功日志记录

**输出**：
- 完整的实现代码
- 单元测试（如需要）

**确认点**：
- [ ] 参数验证是否完整？
- [ ] 业务规则是否实现？
- [ ] 错误处理是否完善？
- [ ] 代码是否符合规范？

### 流程 4: 代码审查

**审查清单**：

#### 4.1 接口设计审查
- [ ] 方法命名是否符合规范？
- [ ] 参数顺序是否正确？
- [ ] 返回值是否合适？
- [ ] 方法职责是否单一？

#### 4.2 参数验证审查
- [ ] 所有必填参数是否验证？
- [ ] 参数格式是否验证？
- [ ] 参数范围是否验证？
- [ ] 错误信息是否明确？

#### 4.3 业务规则审查
- [ ] 业务规则是否完整？
- [ ] 业务规则验证是否在正确的位置？
- [ ] 业务规则错误信息是否明确？

#### 4.4 错误处理审查
- [ ] 所有错误是否记录日志？
- [ ] 错误信息是否明确？
- [ ] 成功操作是否记录日志？

#### 4.5 代码质量审查
- [ ] 代码是否可读？
- [ ] 代码是否可维护？
- [ ] 是否有重复代码？
- [ ] 是否有未使用的代码？

### 流程 5: 测试验证

**测试类型**：
1. **单元测试**
   - 参数验证测试
   - 业务规则测试
   - 错误处理测试

2. **集成测试**
   - Repository 集成测试
   - 事务测试（如需要）

3. **端到端测试**
   - Handler → Service → Repository 完整流程

**确认点**：
- [ ] 单元测试是否覆盖？
- [ ] 集成测试是否通过？
- [ ] 端到端测试是否通过？

---

## 最佳实践

### 实践 1: 保持 Service 层薄

**原则**：
- Service 层应该尽可能薄
- 复杂的业务逻辑应该放在领域模型或 Repository
- Service 层主要负责协调和验证

**示例**：
```go
// ✅ 好的做法：Service 层薄，业务逻辑在 Repository
func (s *Service) CreateAlarmEvent(...) error {
    // 1. 参数验证
    if tenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    
    // 2. 调用 Repository（业务逻辑在 Repository）
    return s.alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, event)
}

// ❌ 不好的做法：业务逻辑在 Service
func (s *Service) CreateAlarmEvent(...) error {
    // 业务逻辑不应该在 Service
    if event.AlarmLevel == "CRIT" {
        // 复杂的业务逻辑...
    }
}
```

### 实践 2: 使用领域模型

**原则**：
- Service 层应该使用领域模型，而不是 map[string]interface{}
- 领域模型应该在 models 或 domain 包中定义
- 避免在 Service 层进行复杂的数据转换

**示例**：
```go
// ✅ 好的做法：使用领域模型
func (s *Service) CreateAlarmEvent(
    ctx context.Context,
    tenantID string,
    event *models.AlarmEvent,
) error {
    // ...
}

// ❌ 不好的做法：使用 map
func (s *Service) CreateAlarmEvent(
    ctx context.Context,
    tenantID string,
    payload map[string]interface{},
) error {
    // 需要复杂的数据转换...
}
```

### 实践 3: 明确的错误处理

**原则**：
- 所有错误都应该明确处理
- 错误信息应该包含足够的上下文
- 区分可恢复错误和不可恢复错误

**示例**：
```go
// ✅ 好的做法：明确的错误处理
func (s *Service) GetAlarmEvent(...) (*AlarmEvent, error) {
    event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
    if err != nil {
        if err == repository.ErrNotFound {
            return nil, fmt.Errorf("alarm event not found: event_id=%s", eventID)
        }
        s.logger.Error("Failed to get alarm event",
            zap.String("tenant_id", tenantID),
            zap.String("event_id", eventID),
            zap.Error(err),
        )
        return nil, fmt.Errorf("failed to get alarm event: %w", err)
    }
    return event, nil
}
```

---

## 示例对比

### 示例 1: 简单 Service vs 编排 Service

**简单 Service**：
```go
type AlarmEventService struct {
    alarmEventsRepo *repository.AlarmEventsRepository
}

func (s *AlarmEventService) GetAlarmEvent(...) (*AlarmEvent, error) {
    return s.alarmEventsRepo.GetAlarmEvent(...)
}
```

**编排 Service**：
```go
type ResidentService struct {
    residentsRepo *repository.ResidentsRepository
    phiRepo       *repository.ResidentPHIRepository
    db            *sql.DB
}

func (s *ResidentService) CreateResidentWithPHI(...) error {
    tx, _ := s.db.BeginTx(...)
    defer tx.Rollback()
    
    // 协调多个 Repository
    s.residentsRepo.CreateResident(ctx, tx, ...)
    s.phiRepo.CreatePHI(ctx, tx, ...)
    
    tx.Commit()
}
```

### 示例 2: 参数验证对比

**好的做法**：
```go
func (s *Service) CreateAlarmEvent(
    ctx context.Context,
    tenantID string,
    event *models.AlarmEvent,
) error {
    // ✅ 明确的参数验证
    if tenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    if event == nil {
        return fmt.Errorf("event is required")
    }
    if event.TenantID != tenantID {
        return fmt.Errorf("event tenant_id (%s) does not match provided tenant_id (%s)", 
            event.TenantID, tenantID)
    }
    
    return s.alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, event)
}
```

**不好的做法**：
```go
func (s *Service) CreateAlarmEvent(
    ctx context.Context,
    tenantID string,
    event *models.AlarmEvent,
) error {
    // ❌ 缺少参数验证
    return s.alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, event)
}
```

---

## 总结

### Service 层设计检查清单

在实现 Service 层之前，确认以下内容：

1. **需求分析**
   - [ ] 业务需求是否清晰？
   - [ ] Service 模式是否确定？
   - [ ] 依赖是否完整？

2. **接口设计**
   - [ ] 方法命名是否符合规范？
   - [ ] 参数顺序是否正确？
   - [ ] 返回值是否合适？

3. **实现设计**
   - [ ] 参数验证是否完整？
   - [ ] 业务规则是否实现？
   - [ ] 错误处理是否完善？

4. **代码质量**
   - [ ] 代码是否可读？
   - [ ] 代码是否可维护？
   - [ ] 是否有重复代码？

5. **测试**
   - [ ] 单元测试是否覆盖？
   - [ ] 集成测试是否通过？

---

## 参考文档

- `ARCHITECTURE_DESIGN.md` - 架构设计文档（wisefido-data）
- `SERVICE_LAYER_DESIGN.md` - Service 层设计文档（wisefido-ai）
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结

