# AlarmEvent Service 深度分析文档

## 阶段 1：深度分析旧 Handler

### 1.1 当前实现状态

#### 旧 Handler 位置
- **文件**: `internal/http/admin_alarm_handlers.go`
- **方法**: `StubHandler.AdminAlarm`
- **路由**: 
  - `GET /admin/api/v1/alarm-events` - 查询报警事件列表
  - `PUT /admin/api/v1/alarm-events/:id/handle` - 处理报警事件

#### 当前实现（Stub）
```go
// GET /admin/api/v1/alarm-events
case r.URL.Path == "/admin/api/v1/alarm-events":
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    // 对齐 GetAlarmEventsResult
    writeJSON(w, http.StatusOK, Ok(map[string]any{
        "items": []any{},
        "pagination": map[string]any{
            "size":  10,
            "page":  1,
            "count": 0,
            "total": 0,
        },
    }))
    return

// PUT /admin/api/v1/alarm-events/:id/handle
case strings.HasPrefix(r.URL.Path, "/admin/api/v1/alarm-events/") && strings.HasSuffix(r.URL.Path, "/handle"):
    if r.Method != http.MethodPut {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
    return
```

**当前状态**: 只是 Stub 实现，返回空数据，没有实际业务逻辑。

---

### 1.2 前端 API 需求分析

#### 1.2.1 查询报警事件列表 API

**端点**: `GET /admin/api/v1/alarm-events`

**请求参数** (`GetAlarmEventsParams`):
```typescript
{
  status: 'active' | 'resolved'  // 报警状态过滤
  // 搜索参数
  alarm_time_start?: number      // timestamp (开始时间)
  alarm_time_end?: number        // timestamp (结束时间)
  resident?: string              // 住户搜索（姓名或账号）
  branch_tag?: string            // 位置标签搜索
  unit_name?: string             // 单元名称搜索
  device_name?: string           // 设备名称搜索
  // 过滤参数
  event_types?: string[]         // 事件类型过滤（多选）
  categories?: string[]          // 类别过滤（多选）
  alarm_levels?: string[]        // 报警级别过滤（多选）
  card_id?: string              // 按卡片ID过滤（后端会转换为 device_ids）
  device_ids?: string[]          // 按设备ID过滤
  // 分页
  page?: number
  page_size?: number
  // 注意：userId 和 role 通过请求头传递（X-User-Id, X-User-Role）
}
```

**响应格式** (`GetAlarmEventsResult`):
```typescript
{
  items: AlarmEvent[]           // 报警事件列表
  pagination: {
    size: number                // 每页数量
    page: number                // 当前页码
    count: number               // 当前页数量
    total: number               // 总数量
  }
}
```

**AlarmEvent 数据结构**:
```typescript
interface AlarmEvent {
  event_id: string              // UUID
  event_type: string            // 事件类型（如 'Fall', 'Radar_AbnormalHeartRate'）
  category: 'safety' | 'clinical' | 'behavioral' | 'device'
  alarm_level: string | number  // 报警级别（'0'/'EMERG', '1'/'ALERT', '2'/'CRIT', '3'/'ERR', '4'/'WARNING'）
  alarm_status: 'active' | 'acknowledged' | 'resolved'
  triggered_at: number          // timestamp
  
  // 处理信息
  handling_state?: 'verified' | 'false_alarm' | 'test'
  handling_details?: string
  handler_id?: string
  handler_name?: string
  handled_at?: number
  
  // 关联数据（通过 JOIN 查询）
  card_id?: string
  device_id?: string
  device_name?: string
  
  resident_id?: string
  resident_name?: string
  resident_gender?: string
  resident_age?: number
  resident_network?: string
  
  // 地址信息（通过 device → unit/room/bed → locations）
  branch_tag?: string
  building?: string
  floor?: string
  area_tag?: string
  unit_name?: string
  room_name?: string
  bed_name?: string
  address_display?: string      // 格式化地址显示
}
```

#### 1.2.2 处理报警事件 API

**端点**: `PUT /admin/api/v1/alarm-events/:id/handle`

**请求参数** (`HandleAlarmEventParams`):
```typescript
{
  alarm_status: 'acknowledged' | 'resolved'  // 目标状态
  handle_type: 'verified' | 'false_alarm' | 'test'  // 处理类型
  remarks?: string                            // 备注
}
```

**响应格式**:
```typescript
{
  success: boolean
}
```

---

### 1.3 业务逻辑需求

#### 1.3.1 查询报警事件列表

**核心业务逻辑**:

1. **权限过滤**:
   - 根据用户角色和权限过滤可查看的报警事件
   - Resident: 只能查看与自己相关的报警（通过 card/resident 关联）
   - Family: 只能查看与家庭成员相关的报警
   - Staff (Nurse, Caregiver): 根据卡片权限过滤
   - Admin/Manager/IT: 可以查看租户内所有报警

2. **复杂查询**:
   - 支持多条件过滤（时间范围、住户、位置、设备、事件类型、类别、级别）
   - 支持模糊搜索（住户名称、位置标签、单元名称、设备名称）
   - 支持多选过滤（event_types, categories, alarm_levels）
   - 支持按 card_id 过滤（需要转换为 device_ids）

3. **跨表 JOIN**:
   - `alarm_events` → `devices` (通过 device_id)
   - `devices` → `cards` (通过 device_id，一个设备只能属于一个卡片)
   - `devices` → `beds` / `rooms` (通过 bound_bed_id / bound_room_id)
   - `beds` → `residents` (通过 resident_id)
   - `beds` / `rooms` → `units` (通过 unit_id)
   - `units` → `locations` (通过 location_id，获取 branch_tag, building, floor)

4. **数据转换**:
   - 时间戳转换（triggered_at, hand_time → timestamp）
   - JSONB 字段解析（trigger_data, notified_users, metadata）
   - 地址格式化（branch_tag-Building-UnitName）
   - 住户信息聚合（nickname 或 first_name + last_name）

5. **分页处理**:
   - 支持 page 和 page_size
   - 返回总数和当前页数量

#### 1.3.2 处理报警事件

**核心业务逻辑**:

1. **权限检查**（重要）:
   - **Facility 类型卡片** (`unit_type = 'Facility'`):
     - 只有 `Nurse` 或 `Caregiver` 角色可以处理报警
     - 其他角色（如 `Admin`, `Manager`, `IT` 等）不能处理
     - **后端必须验证此权限**，防止前端绕过检查
   - **Home 类型卡片** (`unit_type = 'Home'`):
     - 所有角色都可以处理报警
     - 包括：`SystemAdmin`, `Admin`, `Manager`, `IT`, `Nurse`, `Caregiver` 等

2. **权限验证流程**:
   ```
   1. 根据 event_id 查询报警事件
   2. 通过 device_id 关联到设备
   3. 通过设备关联到卡片（一个设备只能属于一个卡片）
   4. 查询卡片的 unit_type
   5. 验证用户角色是否符合权限要求
   ```

3. **状态转换**:
   - `active` → `acknowledged`: 确认报警
   - `active` / `acknowledged` → `resolved`: 解决报警（需要设置 operation）
   - 只能处理状态为 `active` 或 `acknowledged` 的报警

4. **操作结果设置**:
   - `handle_type` 映射到 `operation`:
     - `'verified'` → `'verified_and_processed'`
     - `'false_alarm'` → `'false_alarm'`
     - `'test'` → `'test'`
   - 设置 `handler` 为当前用户ID
   - 设置 `hand_time` 为当前时间
   - 设置 `notes` 为 remarks（如果提供）

5. **数据验证**:
   - event_id 必须存在
   - tenant_id 必须匹配
   - alarm_status 必须是 'active' 或 'acknowledged'
   - handle_type 必须是有效值

---

### 1.4 数据库结构

#### 1.4.1 alarm_events 表

**主要字段**:
- `event_id` (UUID, PRIMARY KEY)
- `tenant_id` (UUID, NOT NULL)
- `device_id` (UUID, NOT NULL)
- `event_type` (VARCHAR(50), NOT NULL)
- `category` (VARCHAR(50), CHECK IN ('safety', 'clinical', 'behavioral', 'device'))
- `alarm_level` (VARCHAR(20), NOT NULL)
- `alarm_status` (VARCHAR(20), DEFAULT 'active', CHECK IN ('active', 'acknowledged'))
- `triggered_at` (TIMESTAMPTZ, NOT NULL)
- `hand_time` (TIMESTAMPTZ, nullable)
- `iot_timeseries_id` (BIGINT, nullable)
- `trigger_data` (JSONB)
- `handler` (UUID, nullable, REFERENCES users(user_id))
- `operation` (VARCHAR(30), nullable, CHECK IN ('verified_and_processed', 'false_alarm', 'test', 'auto_relieved'))
- `notes` (TEXT, nullable)
- `notified_users` (JSONB, DEFAULT '[]'::JSONB)
- `metadata` (JSONB, DEFAULT '{}'::JSONB)
- `created_at` (TIMESTAMPTZ, NOT NULL)
- `updated_at` (TIMESTAMPTZ, NOT NULL)

**索引**:
- `idx_alarm_events_device` (tenant_id, device_id, triggered_at DESC)
- `idx_alarm_events_type_level` (tenant_id, event_type, alarm_level, triggered_at DESC)
- `idx_alarm_events_category` (tenant_id, category, triggered_at DESC)
- `idx_alarm_events_status` (tenant_id, alarm_status, triggered_at DESC)
- `idx_alarm_events_triggered_at` (tenant_id, triggered_at DESC)

#### 1.4.2 关联表

**设备关联**:
- `devices` 表：`device_id` → `device_name`, `bound_bed_id`, `bound_room_id`
- `device_store` 表：`device_id` → `serial_number`

**卡片关联**:
- `cards` 表：通过 `device_id` JOIN（一个设备只能属于一个卡片）
- `cards` 表：`unit_type` ('Facility' | 'Home')

**位置关联**:
- `beds` 表：`bound_bed_id` → `room_id` → `unit_id`
- `rooms` 表：`bound_room_id` → `unit_id`
- `units` 表：`unit_id` → `unit_name`, `location_id`, `area_tag`
- `locations` 表：`location_id` → `branch_tag`, `building`, `floor`

**住户关联**:
- `residents` 表：`bed.resident_id` → `resident_id` → `nickname`, `phi` (first_name, last_name, birth_date, gender)

---

### 1.5 关键业务规则

#### 1.5.1 权限规则

1. **查询权限**:
   - 所有有权限查看报警事件的用户都可以查询
   - 后端应根据用户角色和权限过滤可查看的报警事件
   - 参考：`/alarm/records` 页面的权限配置

2. **处理权限**（重要）:
   - **Facility 类型卡片**: 只有 `Nurse` 或 `Caregiver` 可以处理
   - **Home 类型卡片**: 所有角色都可以处理
   - **后端必须验证此权限**，防止前端绕过检查

#### 1.5.2 状态转换规则

1. **确认报警** (`acknowledged`):
   - 只能从 `active` 状态转换
   - 设置 `hand_time` 为当前时间
   - 设置 `handler` 为当前用户ID

2. **解决报警** (`resolved`):
   - 可以从 `active` 或 `acknowledged` 状态转换
   - 必须设置 `operation`（'verified_and_processed', 'false_alarm', 'test'）
   - 设置 `hand_time` 为当前时间
   - 设置 `handler` 为当前用户ID
   - 设置 `notes`（如果提供）

#### 1.5.3 数据转换规则

1. **时间戳转换**:
   - `triggered_at` (TIMESTAMPTZ) → `triggered_at` (timestamp number)
   - `hand_time` (TIMESTAMPTZ) → `handled_at` (timestamp number)

2. **JSONB 字段解析**:
   - `trigger_data` (JSONB) → 解析为对象
   - `notified_users` (JSONB) → 解析为数组
   - `metadata` (JSONB) → 解析为对象

3. **地址格式化**:
   - `branch_tag-Building-UnitName` (用于列表显示)

4. **住户信息聚合**:
   - `nickname` 或 `first_name + last_name`
   - `birth_date` → `age` (计算)
   - `gender` → `resident_gender`

---

### 1.6 依赖关系

#### 1.6.1 Repository 依赖

需要以下 Repository:
1. **AlarmEventsRepository** - 报警事件数据访问
   - `ListAlarmEvents(ctx, tenantID, filters, page, size)` - 查询列表（支持复杂过滤和 JOIN）
   - `GetAlarmEvent(ctx, tenantID, eventID)` - 获取单个事件
   - `AcknowledgeAlarmEvent(ctx, tenantID, eventID, handlerID)` - 确认报警
   - `UpdateAlarmEventOperation(ctx, tenantID, eventID, operation, handlerID, notes)` - 更新操作结果

2. **DevicesRepository** - 设备信息（用于权限检查）
   - `GetDevice(ctx, tenantID, deviceID)` - 获取设备信息

3. **CardsRepository** - 卡片信息（用于权限检查）
   - `GetCardByDeviceID(ctx, tenantID, deviceID)` - 通过设备ID获取卡片

4. **UnitsRepository** - 单元信息（用于位置查询）
   - `GetUnit(ctx, tenantID, unitID)` - 获取单元信息

5. **ResidentsRepository** - 住户信息（用于住户查询）
   - `GetResident(ctx, tenantID, residentID)` - 获取住户信息

#### 1.6.2 Service 依赖

可能需要以下 Service:
1. **PermissionService** - 权限检查（如果已有）
2. **CardService** - 卡片服务（如果已有，用于权限检查）

---

### 1.7 待确认问题

1. **Service 层位置**:
   - ✅ `AlarmEventService` 已在 `wisefido-alarm` 项目中实现
   - ❓ 是否需要迁移到 `wisefido-data` 项目？
   - ❓ 还是直接在 `wisefido-data` 中创建新的实现？

2. **Repository 层位置**:
   - ✅ `AlarmEventsRepository` 已在 `wisefido-alarm` 项目中实现
   - ❓ 是否需要迁移到 `wisefido-data` 项目？
   - ❓ 还是直接在 `wisefido-data` 中创建新的实现？

3. **权限检查实现**:
   - ❓ 权限检查逻辑是在 Service 层还是 Handler 层？
   - ❓ 是否需要创建 `PermissionService` 或使用现有的权限检查工具？

4. **跨表 JOIN 实现**:
   - ❓ 复杂的跨表 JOIN 是在 Repository 层还是 Service 层实现？
   - ❓ 是否需要创建专门的查询方法？

5. **数据转换实现**:
   - ❓ 数据转换（时间戳、JSONB、地址格式化）是在 Service 层还是 Handler 层？
   - ❓ 是否需要创建专门的转换工具函数？

---

### 1.8 业务逻辑清单

#### 查询报警事件列表

1. ✅ 解析请求参数（status, 时间范围, 搜索文本, 过滤条件, 分页）
2. ✅ 获取 tenant_id（从请求头或查询参数）
3. ✅ 获取用户信息（user_id, role，从请求头）
4. ✅ 权限过滤（根据用户角色过滤可查看的报警事件）
5. ✅ 构建查询过滤器（AlarmEventFilters）
6. ✅ 执行复杂查询（支持多条件过滤和跨表 JOIN）
7. ✅ 数据转换（时间戳、JSONB、地址格式化、住户信息聚合）
8. ✅ 分页处理（计算总数和当前页数量）
9. ✅ 返回响应（items + pagination）

#### 处理报警事件

1. ✅ 解析请求参数（event_id, alarm_status, handle_type, remarks）
2. ✅ 获取 tenant_id（从请求头或查询参数）
3. ✅ 获取用户信息（user_id, role，从请求头）
4. ✅ 查询报警事件（通过 event_id）
5. ✅ 验证报警事件存在性和 tenant_id 匹配
6. ✅ 查询设备信息（通过 device_id）
7. ✅ 查询卡片信息（通过 device_id）
8. ✅ 权限检查（Facility 类型卡片：只有 Nurse/Caregiver 可以处理）
9. ✅ 验证状态转换（只能处理 active 或 acknowledged 状态的报警）
10. ✅ 映射 handle_type 到 operation
11. ✅ 更新报警事件（设置 alarm_status, operation, handler, hand_time, notes）
12. ✅ 返回响应（success: true）

---

### 1.9 关键注意事项

1. **权限检查必须在后端实现**:
   - 前端检查只是 UI 层面的限制
   - 后端必须实现权限验证，防止恶意用户绕过前端检查

2. **Facility 类型卡片的权限限制**:
   - 只有 `Nurse` 或 `Caregiver` 可以处理
   - 其他角色（如 `Admin`, `Manager`, `IT` 等）不能处理
   - 这是业务规则，必须严格执行

3. **复杂查询性能**:
   - 跨表 JOIN 可能影响性能
   - 需要合理使用索引
   - 考虑是否需要缓存

4. **数据一致性**:
   - 确保 tenant_id 匹配
   - 确保 device_id 存在
   - 确保卡片和设备关联正确

---

## 下一步

1. 确认 Service 和 Repository 的实现位置（wisefido-alarm vs wisefido-data）
2. 设计 Service 接口（确认方法签名和 DTO）
3. 实现 Service 层（迁移或新建）
4. 实现 Handler 层
5. 集成和路由注册
6. 验证和测试

