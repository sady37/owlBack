# AlarmEvent Service 接口设计文档

## 阶段 2：设计 Service 接口

### 2.1 Service 接口定义

```go
// AlarmEventService 报警事件服务接口
type AlarmEventService interface {
    // 查询报警事件列表（支持复杂过滤、跨表 JOIN、权限过滤）
    ListAlarmEvents(ctx context.Context, req ListAlarmEventsRequest) (*ListAlarmEventsResponse, error)
    
    // 处理报警事件（确认或解决）
    HandleAlarmEvent(ctx context.Context, req HandleAlarmEventRequest) (*HandleAlarmEventResponse, error)
}
```

---

### 2.2 请求/响应 DTO 设计

#### 2.2.1 ListAlarmEventsRequest - 查询报警事件列表请求

```go
type ListAlarmEventsRequest struct {
    // 必填字段
    TenantID      string // 租户ID
    CurrentUserID string // 当前用户ID（用于权限过滤）
    CurrentUserRole string // 当前用户角色（用于权限过滤）
    
    // 状态过滤
    Status string // 'active' | 'resolved' - 报警状态过滤
    
    // 时间范围过滤
    AlarmTimeStart *int64 // timestamp (开始时间)
    AlarmTimeEnd   *int64 // timestamp (结束时间)
    
    // 搜索参数（模糊匹配）
    Resident   string // 住户搜索（姓名或账号）
    BranchTag  string // 位置标签搜索
    UnitName   string // 单元名称搜索
    DeviceName string // 设备名称搜索
    
    // 过滤参数（多选）
    EventTypes  []string // 事件类型过滤
    Categories  []string // 类别过滤
    AlarmLevels []string // 报警级别过滤
    
    // 关联过滤
    CardID    string   // 按卡片ID过滤（后端会转换为 device_ids）
    DeviceIDs []string // 按设备ID过滤
    
    // 分页
    Page     int // 页码，默认 1
    PageSize int // 每页数量，默认 20，最大 100
}
```

**字段说明**:
- `TenantID`: 从请求头 `X-Tenant-Id` 或查询参数获取
- `CurrentUserID`: 从请求头 `X-User-Id` 获取
- `CurrentUserRole`: 从请求头 `X-User-Role` 获取
- `Status`: 前端传递 'active' 或 'resolved'
- `AlarmTimeStart/End`: 前端传递 timestamp (number)，后端转换为 time.Time
- 搜索参数：支持模糊匹配（ILIKE）
- 过滤参数：支持多选（IN 查询）
- `CardID`: 如果提供，需要查询卡片关联的设备列表，转换为 `DeviceIDs`

#### 2.2.2 ListAlarmEventsResponse - 查询报警事件列表响应

```go
type ListAlarmEventsResponse struct {
    Items []*AlarmEventDTO // 报警事件列表（包含关联数据）
    Pagination PaginationDTO // 分页信息
}

type PaginationDTO struct {
    Size  int // 每页数量
    Page  int // 当前页码
    Count int // 当前页数量
    Total int // 总数量
}
```

#### 2.2.3 AlarmEventDTO - 报警事件 DTO（包含关联数据）

```go
type AlarmEventDTO struct {
    // 基础字段（来自 alarm_events 表）
    EventID     string // UUID
    TenantID    string // 租户ID
    DeviceID    string // 设备ID
    EventType   string // 事件类型
    Category    string // 类别（safety, clinical, behavioral, device）
    AlarmLevel  string // 报警级别（'0'/'EMERG', '1'/'ALERT', '2'/'CRIT', '3'/'ERR', '4'/'WARNING'）
    AlarmStatus string // 报警状态（'active', 'acknowledged', 'resolved'）
    TriggeredAt int64  // timestamp（触发时间）
    
    // 处理信息
    HandlingState   *string // 'verified' | 'false_alarm' | 'test'（从 operation 映射）
    HandlingDetails *string // 备注（从 notes 获取）
    HandlerID       *string // 处理人ID（从 handler 获取）
    HandlerName     *string // 处理人名称（通过 JOIN users 获取）
    HandledAt       *int64  // timestamp（处理时间，从 hand_time 获取）
    
    // 关联数据（通过 JOIN 查询）
    CardID     *string // 卡片ID（通过 device_id JOIN cards 获取）
    DeviceName *string // 设备名称（通过 device_id JOIN devices 获取）
    
    // 住户信息（通过 device → bed → resident 获取）
    ResidentID     *string // 住户ID
    ResidentName   *string // 住户名称（nickname 或 first_name + last_name）
    ResidentGender *string // 住户性别（从 residents.phi 获取）
    ResidentAge    *int    // 住户年龄（从 residents.phi.birth_date 计算）
    ResidentNetwork *string // 住户网络（从 residents 表获取）
    
    // 地址信息（通过 device → unit/room/bed → locations 获取）
    BranchTag     *string // 分支标签
    Building      *string // 建筑名称
    Floor         *string // 楼层（如 "2F"）
    AreaTag       *string // 区域标签（从 units 表获取）
    UnitName      *string // 单元名称
    RoomName      *string // 房间名称（从 rooms 表获取）
    BedName       *string // 床位名称（从 beds 表获取）
    AddressDisplay *string // 格式化地址显示（"branch_tag-Building-UnitName"）
    
    // JSONB 字段（解析后返回）
    TriggerData    map[string]interface{} // 触发数据快照
    NotifiedUsers  []interface{}          // 通知接收者列表
    Metadata       map[string]interface{} // 元数据
}
```

**数据转换说明**:
- `TriggeredAt`: `time.Time` → `int64` (timestamp)
- `HandledAt`: `*time.Time` → `*int64` (timestamp)
- `HandlingState`: `operation` 字段映射
  - `'verified_and_processed'` → `'verified'`
  - `'false_alarm'` → `'false_alarm'`
  - `'test'` → `'test'`
- `TriggerData`, `NotifiedUsers`, `Metadata`: JSONB → Go map/array

#### 2.2.4 HandleAlarmEventRequest - 处理报警事件请求

```go
type HandleAlarmEventRequest struct {
    // 必填字段
    TenantID      string // 租户ID
    EventID       string // 报警事件ID
    CurrentUserID string // 当前用户ID（处理人）
    CurrentUserRole string // 当前用户角色（用于权限检查）
    
    // 处理参数
    AlarmStatus string // 'acknowledged' | 'resolved' - 目标状态
    HandleType  string // 'verified' | 'false_alarm' | 'test' - 处理类型（仅 resolved 时需要）
    Remarks     string // 备注（可选）
}
```

**字段说明**:
- `TenantID`: 从请求头 `X-Tenant-Id` 获取
- `EventID`: 从 URL 路径参数获取（`:id`）
- `CurrentUserID`: 从请求头 `X-User-Id` 获取
- `CurrentUserRole`: 从请求头 `X-User-Role` 获取
- `AlarmStatus`: 前端传递 'acknowledged' 或 'resolved'
- `HandleType`: 前端传递 'verified', 'false_alarm', 或 'test'（仅当 `AlarmStatus` 为 'resolved' 时需要）
- `Remarks`: 可选，前端传递备注信息

#### 2.2.5 HandleAlarmEventResponse - 处理报警事件响应

```go
type HandleAlarmEventResponse struct {
    Success bool // 是否成功
}
```

---

### 2.3 业务逻辑设计

#### 2.3.1 ListAlarmEvents 业务逻辑

**流程**:
1. **参数验证**:
   - `TenantID` 必填
   - `CurrentUserID` 和 `CurrentUserRole` 必填（用于权限过滤）
   - `Page` 和 `PageSize` 验证（默认值、最大值）

2. **权限过滤**:
   - 根据用户角色过滤可查看的报警事件：
     - `Resident`: 只能查看与自己相关的报警（通过 card/resident 关联）
     - `Family`: 只能查看与家庭成员相关的报警
     - `Staff` (Nurse, Caregiver): 根据卡片权限过滤
     - `Admin/Manager/IT`: 可以查看租户内所有报警

3. **构建查询过滤器**:
   - 转换前端参数为 Repository 层的 `AlarmEventFilters`
   - 处理 `CardID` → `DeviceIDs` 转换（查询卡片关联的设备列表）
   - 处理时间戳转换（`int64` → `time.Time`）
   - 处理搜索参数（模糊匹配）
   - 处理多选过滤（数组 → IN 查询）

4. **执行查询**:
   - 调用 Repository 层的 `ListAlarmEvents`（支持复杂 JOIN）
   - Repository 层返回 `[]*models.AlarmEvent` 和总数

5. **数据转换**:
   - 将 `models.AlarmEvent` 转换为 `AlarmEventDTO`
   - 执行跨表 JOIN 查询获取关联数据：
     - 设备信息（device_name）
     - 卡片信息（card_id）
     - 住户信息（resident_id, resident_name, resident_gender, resident_age）
     - 地址信息（branch_tag, building, floor, area_tag, unit_name, room_name, bed_name）
     - 处理人信息（handler_name）
   - 格式化地址显示（"branch_tag-Building-UnitName"）
   - 转换时间戳（`time.Time` → `int64`）
   - 解析 JSONB 字段（`json.RawMessage` → `map[string]interface{}`）

6. **返回响应**:
   - 返回 `ListAlarmEventsResponse`（包含 items 和 pagination）

**关键点**:
- 权限过滤在 Service 层实现（根据用户角色和权限）
- 复杂查询在 Repository 层实现（跨表 JOIN）
- 数据转换在 Service 层实现（DTO 转换、关联数据查询）

#### 2.3.2 HandleAlarmEvent 业务逻辑

**流程**:
1. **参数验证**:
   - `TenantID`, `EventID`, `CurrentUserID`, `CurrentUserRole` 必填
   - `AlarmStatus` 必须是 'acknowledged' 或 'resolved'
   - 如果 `AlarmStatus` 为 'resolved'，`HandleType` 必填

2. **查询报警事件**:
   - 调用 Repository 层的 `GetAlarmEvent` 获取报警事件
   - 验证报警事件存在性和 `TenantID` 匹配

3. **权限检查**（重要）:
   - 查询设备信息（通过 `device_id`）
   - 查询卡片信息（通过 `device_id`，一个设备只能属于一个卡片）
   - 查询卡片的 `unit_type`
   - 验证用户角色是否符合权限要求：
     - **Facility 类型卡片** (`unit_type = 'Facility'`):
       - 只有 `Nurse` 或 `Caregiver` 角色可以处理报警
       - 其他角色（如 `Admin`, `Manager`, `IT` 等）不能处理
     - **Home 类型卡片** (`unit_type = 'Home'`):
       - 所有角色都可以处理报警

4. **状态转换验证**:
   - 确认报警：只能从 `active` 状态转换到 `acknowledged`
   - 解决报警：可以从 `active` 或 `acknowledged` 状态转换到 `resolved`
   - 验证当前状态是否符合转换规则

5. **映射处理类型**:
   - `HandleType` → `Operation`:
     - `'verified'` → `'verified_and_processed'`
     - `'false_alarm'` → `'false_alarm'`
     - `'test'` → `'test'`

6. **更新报警事件**:
   - 如果 `AlarmStatus` 为 'acknowledged':
     - 调用 Repository 层的 `AcknowledgeAlarmEvent`
     - 设置 `alarm_status = 'acknowledged'`
     - 设置 `hand_time = CURRENT_TIMESTAMP`
     - 设置 `handler = CurrentUserID`
   - 如果 `AlarmStatus` 为 'resolved':
     - 调用 Repository 层的 `UpdateAlarmEventOperation`
     - 设置 `alarm_status = 'resolved'`（通过 `UpdateAlarmEvent`）
     - 设置 `operation = HandleType映射值`
     - 设置 `hand_time = CURRENT_TIMESTAMP`
     - 设置 `handler = CurrentUserID`
     - 设置 `notes = Remarks`（如果提供）

7. **返回响应**:
   - 返回 `HandleAlarmEventResponse{Success: true}`

**关键点**:
- 权限检查必须在 Service 层实现（防止前端绕过）
- Facility 类型卡片的权限限制必须严格执行
- 状态转换规则必须验证

---

### 2.4 Repository 接口设计

#### 2.4.1 AlarmEventsRepository 接口

```go
type AlarmEventsRepository interface {
    // 查询报警事件列表（支持复杂过滤和跨表 JOIN）
    ListAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters, page, size int) ([]*models.AlarmEvent, int, error)
    
    // 获取单个报警事件
    GetAlarmEvent(ctx context.Context, tenantID, eventID string) (*models.AlarmEvent, error)
    
    // 确认报警（更新状态为 acknowledged）
    AcknowledgeAlarmEvent(ctx context.Context, tenantID, eventID, handlerID string) error
    
    // 更新操作结果（verified_and_processed, false_alarm, test）
    UpdateAlarmEventOperation(ctx context.Context, tenantID, eventID, operation, handlerID string, notes *string) error
    
    // 更新报警事件（部分更新）
    UpdateAlarmEvent(ctx context.Context, tenantID, eventID string, updates map[string]interface{}) error
}
```

#### 2.4.2 AlarmEventFilters 过滤器

```go
type AlarmEventFilters struct {
    // 时间段过滤
    StartTime *time.Time
    EndTime   *time.Time
    
    // 住户过滤
    ResidentID *string
    
    // 位置过滤
    BranchTag *string
    UnitID    *string
    
    // 设备过滤
    DeviceID     *string
    DeviceName   *string // 模糊匹配
    DeviceIDs    []string // 设备ID列表（IN 查询）
    
    // 事件类型和级别过滤
    EventType  *string
    Category   *string
    AlarmLevel *string
    AlarmLevels []string // IN 查询
    
    // 状态过滤
    AlarmStatus *string
    AlarmStatuses []string // IN 查询
    
    // 操作结果过滤
    Operation *string
    Operations []string // IN 查询
    
    // 处理人过滤
    HandlerID *string
}
```

#### 2.4.3 其他 Repository 依赖

Service 层还需要以下 Repository 来获取关联数据：

```go
// DevicesRepository - 获取设备信息
type DevicesRepository interface {
    GetDevice(ctx context.Context, tenantID, deviceID string) (*domain.Device, error)
}

// CardsRepository - 获取卡片信息（用于权限检查）
type CardsRepository interface {
    GetCardByDeviceID(ctx context.Context, tenantID, deviceID string) (*domain.Card, error)
    GetCardDevices(ctx context.Context, tenantID, cardID string) ([]string, error) // 返回 device_ids
}

// UnitsRepository - 获取单元信息（用于地址查询）
type UnitsRepository interface {
    GetUnit(ctx context.Context, tenantID, unitID string) (*domain.Unit, error)
}

// ResidentsRepository - 获取住户信息（用于住户查询）
type ResidentsRepository interface {
    GetResident(ctx context.Context, tenantID, residentID string) (*domain.Resident, error)
}

// UsersRepository - 获取用户信息（用于处理人名称）
type UsersRepository interface {
    GetUser(ctx context.Context, tenantID, userID string) (*domain.User, error)
}
```

---

### 2.5 待确认问题

1. **Repository 层实现位置**:
   - ❓ 是否在 `wisefido-data` 中创建新的 `AlarmEventsRepository`？
   - ❓ 还是复用 `wisefido-alarm` 中的实现（需要跨项目依赖）？
   - ✅ **建议**: 在 `wisefido-data` 中创建新的实现，保持项目独立性

2. **关联数据查询策略**:
   - ❓ 是在 Repository 层的 `ListAlarmEvents` 中一次性 JOIN 查询所有关联数据？
   - ❓ 还是在 Service 层分别查询关联数据（N+1 查询问题）？
   - ✅ **建议**: 在 Repository 层实现复杂 JOIN 查询，一次性获取所有关联数据（性能更好）

3. **权限检查实现位置**:
   - ❓ 权限检查逻辑是在 Service 层还是 Handler 层？
   - ✅ **建议**: 在 Service 层实现权限检查，Handler 层只负责请求解析和响应格式化

4. **数据转换实现位置**:
   - ❓ 数据转换（时间戳、JSONB、地址格式化）是在 Service 层还是 Handler 层？
   - ✅ **建议**: 在 Service 层实现数据转换，Handler 层只负责 HTTP 层处理

5. **错误处理策略**:
   - ❓ Service 层返回业务错误，Handler 层转换为 HTTP 响应？
   - ✅ **建议**: 采用统一的错误处理策略（Service 层返回业务错误，Handler 层转换为 HTTP 响应）

---

### 2.6 接口设计总结

**Service 接口**:
- `ListAlarmEvents` - 查询报警事件列表（支持复杂过滤、跨表 JOIN、权限过滤）
- `HandleAlarmEvent` - 处理报警事件（权限检查、状态转换）

**关键特性**:
- ✅ 支持复杂查询（多条件过滤、跨表 JOIN）
- ✅ 支持权限过滤（根据用户角色）
- ✅ 支持关联数据查询（设备、卡片、住户、地址信息）
- ✅ 支持权限检查（Facility 类型卡片权限限制）
- ✅ 支持状态转换验证
- ✅ 支持数据转换（时间戳、JSONB、地址格式化）

**下一步**:
1. 确认接口设计（等待用户确认）
2. 实现 Repository 层（如果需要在 `wisefido-data` 中创建）
3. 实现 Service 层
4. 实现 Handler 层
5. 集成和测试

