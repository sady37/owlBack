# Unit Service 实现对比文档

## 阶段 3：实现 Service - 业务逻辑对比

### 实现状态
✅ **所有方法已实现并编译通过**

### 方法实现清单

#### Building 管理
- ✅ `ListBuildings` - 查询楼栋列表
- ✅ `GetBuilding` - 获取单个楼栋详情
- ✅ `CreateBuilding` - 创建楼栋
- ✅ `UpdateBuilding` - 更新楼栋
- ✅ `DeleteBuilding` - 删除楼栋

#### Unit 管理
- ✅ `ListUnits` - 查询单元列表
- ✅ `GetUnit` - 获取单个单元详情
- ✅ `CreateUnit` - 创建单元
- ✅ `UpdateUnit` - 更新单元
- ✅ `DeleteUnit` - 删除单元

#### Room 管理
- ✅ `ListRooms` - 查询房间列表
- ✅ `ListRoomsWithBeds` - 查询房间及其床位列表
- ✅ `GetRoom` - 获取单个房间详情
- ✅ `CreateRoom` - 创建房间
- ✅ `UpdateRoom` - 更新房间
- ✅ `DeleteRoom` - 删除房间

#### Bed 管理
- ✅ `ListBeds` - 查询床位列表
- ✅ `GetBed` - 获取单个床位详情
- ✅ `CreateBed` - 创建床位
- ✅ `UpdateBed` - 更新床位
- ✅ `DeleteBed` - 删除床位

### 业务逻辑对比

#### 1. ListUnits

**旧 Handler 逻辑** (`admin_units_devices_impl.go:54-99`):
```go
1. 从请求中提取 tenant_id
2. 构建 filters map[string]string：
   - branch_tag: 如果 query 参数中存在（即使为空也添加，用于匹配 NULL）
   - building, floor, area_tag, unit_number, unit_name, unit_type: 如果非空则添加
3. 分页参数：page (默认1), size (默认100)
4. 调用 Repository: ListUnits(ctx, tenantID, filters, page, size)
5. 转换结果：每个 unit 调用 ToJSON()
6. 返回格式：{items: [...], total: number}
```

**新 Service 逻辑** (`unit_service.go:ListUnits`):
```go
1. 参数验证：tenant_id 必填
2. 构建 UnitFilters 结构体（与旧逻辑对齐）：
   - 使用 strings.TrimSpace 处理所有字段
   - branch_tag 保持原值（空字符串表示匹配 NULL）
3. 分页参数：page (默认1), size (默认100) - 与旧逻辑一致
4. 调用 Repository: ListUnits(ctx, tenantID, filters, page, size)
5. 返回 ListUnitsResponse{Items: items, Total: total}
```

**差异点**：
- ✅ Service 层增加了参数验证
- ✅ Service 层使用结构化的 UnitFilters 而不是 map
- ✅ Service 层增加了日志记录
- ✅ 返回类型从 map[string]any 改为结构化响应

#### 2. GetUnit

**旧 Handler 逻辑** (`admin_units_devices_impl.go:101-116`):
```go
1. 从请求中提取 tenant_id
2. 调用 Repository: GetUnit(ctx, tenantID, unitID)
3. 错误处理：
   - sql.ErrNoRows -> "unit not found"
   - 其他错误 -> "failed to get unit"
4. 返回格式：unit.ToJSON()
```

**新 Service 逻辑** (`unit_service.go:GetUnit`):
```go
1. 参数验证：tenant_id 和 unit_id 必填
2. 调用 Repository: GetUnit(ctx, tenantID, unitID)
3. 错误处理：
   - sql.ErrNoRows -> "unit not found" (与旧逻辑一致)
   - 其他错误 -> "failed to get unit: %w" (增加了错误包装)
4. 返回 GetUnitResponse{Unit: unit}
5. 增加了日志记录
```

**差异点**：
- ✅ Service 层增加了参数验证
- ✅ Service 层增加了详细的错误日志
- ✅ 返回类型从 map 改为结构化响应

#### 3. CreateUnit

**旧 Handler 逻辑** (`admin_units_devices_impl.go:118-139`):
```go
1. 从请求中提取 tenant_id
2. 读取请求体 JSON -> payload map[string]any
3. 调用 Repository: CreateUnit(ctx, tenantID, payload)
4. 错误处理：
   - 唯一约束冲突 -> checkUnitUniqueConstraintError() 返回友好错误消息
   - 其他错误 -> "failed to create unit: " + err.Error()
5. 返回格式：unit.ToJSON()
```

**新 Service 逻辑** (`unit_service.go:CreateUnit`):
```go
1. 参数验证：tenant_id, unit_name, unit_number, unit_type, timezone 必填
2. 构建 domain.Unit：
   - 使用 normalizeBranchTag, normalizeAreaTag, normalizeLayoutConfig 规范化数据
   - 设置默认值：building = "-", floor = "1F"
3. 业务规则验证：branch_tag 和 building 不能同时为空
4. 调用 Repository: CreateUnit(ctx, tenantID, unit)
5. 错误处理：统一返回 "failed to create unit: %w"
6. 返回 CreateUnitResponse{UnitID: unitID}
7. 增加了日志记录
```

**差异点**：
- ✅ Service 层增加了完整的参数验证
- ✅ Service 层增加了数据规范化逻辑
- ✅ Service 层增加了业务规则验证（提前验证，更友好）
- ✅ Service 层使用强类型的 domain.Unit 而不是 map[string]any
- ⚠️ 唯一约束错误处理：旧 Handler 有 checkUnitUniqueConstraintError()，新 Service 依赖 Repository 的错误消息（可以后续增强）

#### 4. UpdateUnit

**旧 Handler 逻辑** (`admin_units_devices_impl.go:141-162`):
```go
1. 从请求中提取 tenant_id
2. 读取请求体 JSON -> payload map[string]any
3. 调用 Repository: UpdateUnit(ctx, tenantID, unitID, payload)
4. 错误处理：
   - 唯一约束冲突 -> checkUnitUniqueConstraintError()
   - 其他错误 -> "failed to update unit: " + err.Error()
5. 返回格式：unit.ToJSON()
```

**新 Service 逻辑** (`unit_service.go:UpdateUnit`):
```go
1. 参数验证：tenant_id 和 unit_id 必填
2. 先获取当前 unit（用于部分更新）
3. 构建更新后的 unit（只更新提供的字段）：
   - 对于可选字段（branch_tag, area_tag, layout_config）：
     * 如果提供了非空值，更新
     * 如果提供了空字符串且当前值存在，清除（设置为 NULL）
     * 如果未提供，保持原值
4. 调用 Repository: UpdateUnit(ctx, tenantID, unitID, unit)
5. 错误处理：统一返回 "failed to update unit: %w"
6. 返回 UpdateUnitResponse{Success: true}
7. 增加了日志记录
```

**差异点**：
- ✅ Service 层增加了参数验证
- ✅ Service 层实现了部分更新逻辑（先获取当前值，只更新提供的字段）
- ✅ Service 层使用强类型的 domain.Unit 而不是 map[string]any
- ✅ Service 层正确处理了可选字段的清除逻辑
- ⚠️ 唯一约束错误处理：可以后续增强

#### 5. DeleteUnit

**旧 Handler 逻辑** (`admin_units_devices_impl.go:164-174`):
```go
1. 从请求中提取 tenant_id
2. 调用 Repository: DeleteUnit(ctx, tenantID, unitID)
3. 错误处理：统一返回 "failed to delete unit"
4. 返回格式：Ok(nil)
```

**新 Service 逻辑** (`unit_service.go:DeleteUnit`):
```go
1. 参数验证：tenant_id 和 unit_id 必填
2. 调用 Repository: DeleteUnit(ctx, tenantID, unitID)
3. 错误处理：返回 "failed to delete unit: %w" (增加了错误包装)
4. 返回 DeleteUnitResponse{Success: true}
5. 增加了日志记录
```

**差异点**：
- ✅ Service 层增加了参数验证
- ✅ Service 层增加了错误包装和日志记录

### 关键改进点

1. **类型安全**：
   - 旧 Handler 使用 `map[string]any`，新 Service 使用强类型的 `domain.Unit`, `domain.Room`, `domain.Bed`
   - 旧 Handler 使用 `map[string]string` 作为 filters，新 Service 使用 `UnitFilters` 结构体

2. **参数验证**：
   - 所有方法都增加了完整的参数验证
   - 提前验证，提供更友好的错误消息

3. **数据规范化**：
   - 使用辅助函数规范化数据（`normalizeBranchTag`, `normalizeAreaTag`, `normalizeLayoutConfig`）
   - 统一处理空字符串和 "-" 的转换逻辑

4. **错误处理**：
   - 增加了详细的错误日志
   - 使用 `fmt.Errorf` 包装错误，保留错误链

5. **业务逻辑**：
   - 提前验证业务规则（如 branch_tag 和 building 不能同时为空）
   - 实现了部分更新逻辑（UpdateUnit, UpdateRoom, UpdateBed）

6. **日志记录**：
   - 所有方法都增加了结构化日志记录（使用 zap）
   - 记录关键参数和错误信息

### 待优化点

1. **唯一约束错误处理**：
   - 可以考虑在 Service 层增加类似 `checkUnitUniqueConstraintError` 的逻辑，提供更友好的错误消息

2. **部分更新逻辑**：
   - 当前实现中，对于可选字段，如果请求中提供了空字符串，会清除该字段
   - 但无法区分"未提供"和"空字符串"，可以考虑使用指针类型

3. **事务支持**：
   - 当前实现没有事务支持，如果后续需要，可以考虑增加事务管理

### 下一步

- ✅ 阶段 3 完成：所有方法已实现并编译通过
- ⏭️ 阶段 4：编写 Service 测试
- ⏭️ 阶段 5：实现 Handler
- ⏭️ 阶段 6：集成和路由注册
- ⏭️ 阶段 7：验证和测试

