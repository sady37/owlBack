# Unit Service 业务逻辑分析文档

## 阶段 1：深度分析旧 Handler

### 1.1 当前 Handler 实现位置
- **文件**: `internal/http/admin_units_devices_impl.go`
- **Handler 方法**: `AdminAPI` 的方法

### 1.2 Unit 相关端点清单

#### 1.2.1 Buildings（楼栋）
- `GET /admin/api/v1/buildings` - `getBuildings()`
- `POST /admin/api/v1/buildings` - `createBuilding()` (通过 buildingWriter 接口)
- `PUT /admin/api/v1/buildings/:id` - `updateBuilding()` (通过 buildingWriter 接口)
- `DELETE /admin/api/v1/buildings/:id` - `deleteBuilding()` (通过 buildingWriter 接口)

#### 1.2.2 Units（单元）
- `GET /admin/api/v1/units` - `getUnits()`
- `GET /admin/api/v1/units/:id` - `getUnitDetail()`
- `POST /admin/api/v1/units` - `createUnit()`
- `PUT /admin/api/v1/units/:id` - `updateUnit()`
- `DELETE /admin/api/v1/units/:id` - `deleteUnit()`

#### 1.2.3 Rooms（房间）
- `GET /admin/api/v1/rooms` - `getRoomsWithBeds()` (需要 unit_id 参数)
- `POST /admin/api/v1/rooms` - `createRoom()`
- `PUT /admin/api/v1/rooms/:id` - `updateRoom()`
- `DELETE /admin/api/v1/rooms/:id` - `deleteRoom()`

#### 1.2.4 Beds（床位）
- `GET /admin/api/v1/beds` - `getBeds()` (需要 room_id 参数)
- `POST /admin/api/v1/beds` - `createBed()`
- `PUT /admin/api/v1/beds/:id` - `updateBed()`
- `DELETE /admin/api/v1/beds/:id` - `deleteBed()`

### 1.3 业务逻辑提取

#### 1.3.1 getUnits() 业务逻辑
```go
// 位置: admin_units_devices_impl.go:54-99
业务逻辑：
1. 从请求中提取 tenant_id（通过 tenantIDFromReq）
2. 构建 filters map[string]string：
   - branch_tag: 如果 query 参数中存在（即使为空也添加，用于匹配 NULL）
   - building, floor, area_tag, unit_number, unit_name, unit_type: 如果非空则添加
3. 分页参数：page (默认1), size (默认100)
4. 调用 Repository: ListUnits(ctx, tenantID, filters, page, size)
5. 转换结果：每个 unit 调用 ToJSON()
6. 返回格式：{items: [...], total: number}

问题：
- filters 类型是 map[string]string，但 Repository 需要 UnitFilters 类型
- 缺少搜索参数处理（Repository 支持 Search 字段）
```

#### 1.3.2 getUnitDetail() 业务逻辑
```go
// 位置: admin_units_devices_impl.go:101-116
业务逻辑：
1. 从请求中提取 tenant_id
2. 调用 Repository: GetUnit(ctx, tenantID, unitID)
3. 错误处理：
   - sql.ErrNoRows -> "unit not found"
   - 其他错误 -> "failed to get unit"
4. 返回格式：unit.ToJSON()
```

#### 1.3.3 createUnit() 业务逻辑
```go
// 位置: admin_units_devices_impl.go:118-139
业务逻辑：
1. 从请求中提取 tenant_id
2. 读取请求体 JSON -> payload map[string]any
3. 调用 Repository: CreateUnit(ctx, tenantID, payload)
4. 错误处理：
   - 唯一约束冲突 -> checkUnitUniqueConstraintError() 返回友好错误消息
   - 其他错误 -> "failed to create unit: " + err.Error()
5. 返回格式：unit.ToJSON()

问题：
- Repository 接口需要 *domain.Unit，但传入的是 map[string]any
- 缺少数据验证和转换逻辑
```

#### 1.3.4 updateUnit() 业务逻辑
```go
// 位置: admin_units_devices_impl.go:141-162
业务逻辑：
1. 从请求中提取 tenant_id
2. 读取请求体 JSON -> payload map[string]any
3. 调用 Repository: UpdateUnit(ctx, tenantID, unitID, payload)
4. 错误处理：
   - 唯一约束冲突 -> checkUnitUniqueConstraintError()
   - 其他错误 -> "failed to update unit: " + err.Error()
5. 返回格式：unit.ToJSON()

问题：
- Repository 接口需要 *domain.Unit，但传入的是 map[string]any
- Repository 返回类型不匹配（返回1个值，但代码期望2个值）
```

#### 1.3.5 deleteUnit() 业务逻辑
```go
// 位置: admin_units_devices_impl.go:164-174
业务逻辑：
1. 从请求中提取 tenant_id
2. 调用 Repository: DeleteUnit(ctx, tenantID, unitID)
3. 错误处理：统一返回 "failed to delete unit"
4. 返回格式：Ok(nil)
```

#### 1.3.6 getRoomsWithBeds() 业务逻辑
```go
// 位置: admin_units_devices_impl.go:176-188
业务逻辑：
1. 从 query 参数获取 unit_id（必需）
2. 调用 Repository: ListRoomsWithBeds(ctx, unitID)
3. 错误处理：统一返回 "failed to list rooms"
4. 返回格式：Ok(out) - 直接返回 Repository 结果

问题：
- 缺少 tenant_id 验证
- Repository 接口需要 (ctx, tenantID, unitID)，但只传了 unitID
```

#### 1.3.7 createRoom() 业务逻辑
```go
// 位置: admin_units_devices_impl.go:190-207
业务逻辑：
1. 读取请求体 JSON -> payload map[string]any
2. 从 payload 提取 unit_id（必需）
3. 调用 Repository: CreateRoom(ctx, unitID, payload)
4. 错误处理：统一返回 "failed to create room"
5. 返回格式：room.ToJSON()

问题：
- 缺少 tenant_id 验证
- Repository 接口需要 (ctx, tenantID, unitID, *domain.Room)，但传入的是 map[string]any
```

#### 1.3.8 updateRoom() / deleteRoom() / getBeds() / createBed() / updateBed() / deleteBed()
类似问题：缺少 tenant_id 验证，类型不匹配

### 1.4 发现的问题总结

#### 1.4.1 类型不匹配问题
1. **filters 类型**: Handler 使用 `map[string]string`，但 Repository 需要 `UnitFilters` 结构体
2. **CreateUnit/UpdateUnit**: Handler 传入 `map[string]any`，但 Repository 需要 `*domain.Unit`
3. **CreateRoom/UpdateRoom**: Handler 传入 `map[string]any`，但 Repository 需要 `*domain.Room`
4. **UpdateUnit 返回值**: Repository 返回1个值，但代码期望2个值

#### 1.4.2 业务逻辑缺失
1. **tenant_id 验证**: Rooms 和 Beds 操作缺少 tenant_id 验证
2. **数据转换**: 缺少 map[string]any -> domain 模型的转换逻辑
3. **参数验证**: 缺少业务规则验证（如唯一性、依赖关系等）
4. **搜索功能**: getUnits 缺少 search 参数处理

#### 1.4.3 错误处理
1. **错误消息**: 部分错误消息过于简单，缺少上下文
2. **错误类型**: 缺少对不同错误类型的区分处理

### 1.5 Repository 接口分析

#### 1.5.1 UnitsRepository 接口（units_repo.go）
```go
// Building 操作
ListBuildings(ctx, tenantID, branchTag) ([]*domain.Building, error)
GetBuilding(ctx, tenantID, buildingID) (*domain.Building, error)
CreateBuilding(ctx, tenantID, building *domain.Building) (string, error)
UpdateBuilding(ctx, tenantID, buildingID, building *domain.Building) error
DeleteBuilding(ctx, tenantID, buildingID) error

// Unit 操作
ListUnits(ctx, tenantID, filters UnitFilters, page, size) ([]*domain.Unit, int, error)
GetUnit(ctx, tenantID, unitID) (*domain.Unit, error)
CreateUnit(ctx, tenantID, unit *domain.Unit) (string, error)
UpdateUnit(ctx, tenantID, unitID, unit *domain.Unit) error
DeleteUnit(ctx, tenantID, unitID) error

// Room 操作
ListRooms(ctx, tenantID, unitID) ([]*domain.Room, error)
ListRoomsWithBeds(ctx, tenantID, unitID) ([]RoomWithBeds, error)
GetRoom(ctx, tenantID, roomID) (*domain.Room, error)
CreateRoom(ctx, tenantID, unitID, room *domain.Room) (string, error)
UpdateRoom(ctx, tenantID, roomID, room *domain.Room) error
DeleteRoom(ctx, tenantID, roomID) error

// Bed 操作
ListBeds(ctx, tenantID, roomID) ([]*domain.Bed, error)
GetBed(ctx, tenantID, bedID) (*domain.Bed, error)
CreateBed(ctx, tenantID, roomID, bed *domain.Bed) (string, error)
UpdateBed(ctx, tenantID, bedID, bed *domain.Bed) error
DeleteBed(ctx, tenantID, bedID) error
```

### 1.6 需要 Service 层处理的业务逻辑

1. **参数验证和转换**
   - Query 参数 -> UnitFilters 结构体
   - map[string]any -> domain 模型转换
   - tenant_id 验证（所有操作都需要）

2. **业务规则验证**
   - Unit 唯一性约束（tenant_id + branch_tag + building + floor + unit_name）
   - 依赖关系验证（Room 需要 Unit 存在，Bed 需要 Room 存在）
   - 数据完整性验证

3. **错误处理增强**
   - 区分不同类型的错误
   - 提供更友好的错误消息
   - 记录详细的错误日志

4. **数据转换**
   - HTTP 层数据 -> Service 层数据
   - Service 层数据 -> Repository 层数据

### 1.7 下一步：设计 Service 接口

基于以上分析，需要设计 UnitService 接口，包含：
- Building 管理方法
- Unit 管理方法
- Room 管理方法
- Bed 管理方法

每个方法需要：
- 清晰的请求/响应结构
- 完整的参数验证
- 业务规则验证
- 错误处理

