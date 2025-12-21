# Handler 重构策略和规范

## 📋 重构策略

### 策略选择：**按业务领域边界，增量式重构**

**推荐方案**：**先完成一个业务领域的 Service，然后立即重构对应的 Handler**

**原因**：
1. ✅ **快速验证**：每个领域完成后可以立即测试，确保 Service 和 Handler 集成正确
2. ✅ **降低风险**：小步快跑，避免大规模重构带来的风险
3. ✅ **持续交付**：每个领域完成后都可以交付使用
4. ✅ **易于回滚**：如果某个领域有问题，可以单独回滚

**不推荐**：等所有 Service 完成后再统一重构 Handler
- ❌ 风险集中：所有问题会在最后暴露
- ❌ 难以测试：无法逐步验证
- ❌ 回滚困难：改动范围太大

---

## 🎯 业务领域优先级

### Phase 1: 用户权限层（高优先级）
1. ✅ **RoleService** - 已完成
2. ✅ **RolePermissionService** - 已完成
3. ⏳ **UserService** - 待实现
4. ⏳ **AuthService** - 待实现

### Phase 2: 业务层（中优先级）
5. ⏳ **TagService** - 待实现
6. ⏳ **ResidentService** - 待实现
7. ⏳ **UnitService** - 待实现
8. ⏳ **DeviceService** - 待实现

### Phase 3: 其他（低优先级）
9. ⏳ **AlarmCloudService** - 待实现
10. ⏳ **AlarmEventService** - 待实现
11. ⏳ **VitalFocusService** - 待实现

---

## 📐 Handler 规范

### 1. Handler 结构规范

#### 1.1 独立 Handler 类型（推荐）

**适用场景**：业务领域有多个端点，需要独立管理

```go
// admin_roles_handler.go
package httpapi

import (
    "context"
    "net/http"
    "strings"
    "wisefido-data/internal/service"
    "go.uber.org/zap"
)

// RolesHandler 角色管理 Handler
type RolesHandler struct {
    roleService *service.RoleService
    logger      *zap.Logger
}

// NewRolesHandler 创建角色管理 Handler
func NewRolesHandler(roleService *service.RoleService, logger *zap.Logger) *RolesHandler {
    return &RolesHandler{
        roleService: roleService,
        logger:      logger,
    }
}

// ServeHTTP 实现 http.Handler 接口
func (h *RolesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 路由分发
    switch {
    case r.URL.Path == "/admin/api/v1/roles" && r.Method == http.MethodGet:
        h.ListRoles(w, r)
    case r.URL.Path == "/admin/api/v1/roles" && r.Method == http.MethodPost:
        h.CreateRole(w, r)
    case strings.HasPrefix(r.URL.Path, "/admin/api/v1/roles/") && r.Method == http.MethodPut:
        h.UpdateRole(w, r)
    case strings.HasPrefix(r.URL.Path, "/admin/api/v1/roles/") && r.Method == http.MethodDelete:
        h.DeleteRole(w, r)
    default:
        w.WriteHeader(http.StatusNotFound)
    }
}

// ListRoles 查询角色列表
func (h *RolesHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 参数解析和验证
    tenantID, ok := h.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    
    search := strings.TrimSpace(r.URL.Query().Get("search"))
    page := parseInt(r.URL.Query().Get("page"), 1)
    size := parseInt(r.URL.Query().Get("size"), 20)
    
    // 2. 调用 Service
    req := service.ListRolesRequest{
        TenantID: &tenantID,
        Search:   search,
        Page:     page,
        Size:     size,
    }
    
    resp, err := h.roleService.ListRoles(ctx, req)
    if err != nil {
        h.logger.Error("ListRoles failed", zap.Error(err))
        writeJSON(w, http.StatusOK, Fail(err.Error()))
        return
    }
    
    // 3. 返回响应
    writeJSON(w, http.StatusOK, Ok(resp))
}

// CreateRole 创建角色
func (h *RolesHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 参数解析和验证
    tenantID, ok := h.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    
    var payload struct {
        RoleCode    string `json:"role_code"`
        DisplayName string `json:"display_name"`
        Description string `json:"description"`
    }
    if err := readBodyJSON(r, 1<<20, &payload); err != nil {
        writeJSON(w, http.StatusOK, Fail("invalid body"))
        return
    }
    
    // 2. 调用 Service
    req := service.CreateRoleRequest{
        TenantID:    tenantID,
        RoleCode:    payload.RoleCode,
        DisplayName: payload.DisplayName,
        Description: payload.Description,
    }
    
    resp, err := h.roleService.CreateRole(ctx, req)
    if err != nil {
        h.logger.Error("CreateRole failed", zap.Error(err))
        writeJSON(w, http.StatusOK, Fail(err.Error()))
        return
    }
    
    // 3. 返回响应
    writeJSON(w, http.StatusOK, Ok(resp))
}

// UpdateRole 更新角色
func (h *RolesHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 参数解析
    roleID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/roles/")
    if roleID == "" || strings.Contains(roleID, "/") {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    userRole := r.Header.Get("X-User-Role")
    
    var payload map[string]any
    if err := readBodyJSON(r, 1<<20, &payload); err != nil {
        writeJSON(w, http.StatusOK, Fail("invalid body"))
        return
    }
    
    // 2. 构建请求
    req := service.UpdateRoleRequest{
        RoleID:   roleID,
        UserRole: userRole,
    }
    
    // 处理 is_active
    if v, ok := payload["is_active"].(bool); ok {
        req.IsActive = &v
    }
    
    // 处理 _delete
    if v, ok := payload["_delete"].(bool); ok && v {
       	deleteFlag := true
        req.Delete = &deleteFlag
    }
    
    // 处理 display_name 和 description
    if v, ok := payload["display_name"].(string); ok {
        req.DisplayName = &v
    }
    if v, ok := payload["description"].(string); ok {
        req.Description = &v
    }
    
    // 3. 调用 Service
    err := h.roleService.UpdateRole(ctx, req)
    if err != nil {
        h.logger.Error("UpdateRole failed", zap.Error(err))
        writeJSON(w, http.StatusOK, Fail(err.Error()))
        return
    }
    
    // 4. 返回响应
    writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// DeleteRole 删除角色
func (h *RolesHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 参数解析
    roleID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/roles/")
    if roleID == "" || strings.Contains(roleID, "/") {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    // 2. 调用 Service
    req := service.UpdateRoleRequest{
        RoleID:  roleID,
        Delete:  func() *bool { b := true; return &b }(),
    }
    
    err := h.roleService.UpdateRole(ctx, req)
    if err != nil {
        h.logger.Error("DeleteRole failed", zap.Error(err))
        writeJSON(w, http.StatusOK, Fail(err.Error()))
        return
    }
    
    // 3. 返回响应
    writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// tenantIDFromReq 从请求中获取 tenant_id（复用 StubHandler 的逻辑）
func (h *RolesHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
    // 复用 StubHandler 的逻辑
    // 或者提取为公共函数
    return "", false // TODO: 实现
}
```

#### 1.2 StubHandler 方法（过渡方案）

**适用场景**：快速迁移，保持现有结构

```go
// admin_roles_handlers.go
func (s *StubHandler) AdminRoles(w http.ResponseWriter, r *http.Request) {
    // 如果 Service 可用，使用 Service
    if s.RoleService != nil {
        h := NewRolesHandler(s.RoleService, s.Logger)
        h.ServeHTTP(w, r)
        return
    }
    
    // 否则，使用旧的 DB 直接操作（向后兼容）
    // ... 原有逻辑
}
```

---

### 2. Handler 职责规范

#### 2.1 Handler 只负责

1. **HTTP 请求/响应处理**
   - 解析请求参数（Query、Body、Header）
   - 验证 HTTP 层面的参数（类型、格式）
   - 构建 HTTP 响应

2. **调用 Service**
   - 将 HTTP 参数转换为 Service 请求
   - 调用 Service 方法
   - 处理 Service 返回的错误

3. **错误处理**
   - 将 Service 错误转换为 HTTP 错误响应
   - 记录错误日志

#### 2.2 Handler 不应该

1. ❌ **直接操作数据库**（应该通过 Service）
2. ❌ **业务规则验证**（应该在 Service 层）
3. ❌ **权限检查**（应该在 Service 层）
4. ❌ **数据转换**（应该在 Service 层）
5. ❌ **复杂业务逻辑**（应该在 Service 层）

---

### 3. Handler 代码规范

#### 3.1 统一错误处理

```go
// 使用统一的错误响应格式
func (h *RolesHandler) handleError(w http.ResponseWriter, err error, operation string) {
    h.logger.Error(operation+" failed", zap.Error(err))
    writeJSON(w, http.StatusOK, Fail(err.Error()))
}
```

#### 3.2 统一参数解析

```go
// 提取公共的参数解析函数
func parsePaginationParams(r *http.Request) (page, size int) {
    page = parseInt(r.URL.Query().Get("page"), 1)
    size = parseInt(r.URL.Query().Get("size"), 20)
    if page <= 0 {
        page = 1
    }
    if size <= 0 {
        size = 20
    }
    return page, size
}
```

#### 3.3 统一响应格式

```go
// 使用统一的响应格式
writeJSON(w, http.StatusOK, Ok(resp))  // 成功
writeJSON(w, http.StatusOK, Fail(err.Error()))  // 失败
```

---

### 4. Handler 测试规范

#### 4.1 单元测试（Mock Service）

```go
func TestRolesHandler_ListRoles(t *testing.T) {
    // Mock Service
    mockService := &MockRoleService{}
    handler := NewRolesHandler(mockService, zap.NewNop())
    
    // 测试请求
    req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/roles?page=1&size=20", nil)
    req.Header.Set("X-Tenant-Id", "test-tenant")
    w := httptest.NewRecorder()
    
    // 执行
    handler.ListRoles(w, req)
    
    // 验证
    assert.Equal(t, http.StatusOK, w.Code)
    // ... 验证响应内容
}
```

#### 4.2 集成测试（真实 Service + Repository）

```go
func TestRolesHandler_Integration(t *testing.T) {
    // 使用真实的数据库和 Service
    db := getTestDB(t)
    roleRepo := repository.NewPostgresRolesRepository(db)
    roleService := service.NewRoleService(roleRepo, zap.NewNop())
    handler := NewRolesHandler(roleService, zap.NewNop())
    
    // 测试请求
    // ... 执行测试
}
```

---

## 🔄 重构步骤

### 步骤 1: 创建 Service（已完成 RoleService 和 RolePermissionService）

### 步骤 2: 创建 Handler

1. 创建独立的 Handler 类型（如 `RolesHandler`）
2. 实现所有端点方法
3. 添加单元测试

### 步骤 3: 集成到路由

```go
// cmd/wisefido-data/main.go
// 创建 Service
roleService := service.NewRoleService(roleRepo, logger)
rolePermService := service.NewRolePermissionService(rolePermRepo, logger)

// 创建 Handler
rolesHandler := httpapi.NewRolesHandler(roleService, logger)
rolePermHandler := httpapi.NewRolePermissionsHandler(rolePermService, logger)

// 注册路由
router.RegisterRolesRoutes(rolesHandler)
router.RegisterRolePermissionsRoutes(rolePermHandler)
```

### 步骤 4: 更新 StubHandler（向后兼容）

```go
// admin_roles_handlers.go
func (s *StubHandler) AdminRoles(w http.ResponseWriter, r *http.Request) {
    // 如果新的 Handler 可用，使用新的
    if s.RolesHandler != nil {
        s.RolesHandler.ServeHTTP(w, r)
        return
    }
    
    // 否则，使用旧的逻辑（向后兼容）
    // ... 原有逻辑
}
```

### 步骤 5: 测试和验证

1. 运行单元测试
2. 运行集成测试
3. 手动测试 API 端点
4. 验证前端功能正常

### 步骤 6: 清理旧代码

1. 确认新 Handler 工作正常
2. 删除 StubHandler 中的旧逻辑
3. 更新文档

---

## 📋 重构检查清单

### 每个 Handler 重构时检查：

- [ ] Handler 结构清晰（独立类型或 StubHandler 方法）
- [ ] 所有端点都已实现
- [ ] 参数解析和验证正确
- [ ] 错误处理统一
- [ ] 日志记录完整
- [ ] 单元测试覆盖
- [ ] 集成测试通过
- [ ] 前端功能验证通过
- [ ] 向后兼容（如果使用 StubHandler 过渡）

---

## 🎯 下一步行动

### 立即执行（RoleService 和 RolePermissionService 已完成）

1. ✅ 创建 `RolesHandler`（使用 RoleService）
2. ✅ 创建 `RolePermissionsHandler`（使用 RolePermissionService）
3. ✅ 更新路由注册
4. ✅ 添加测试
5. ✅ 验证功能

### 后续执行（按优先级）

1. 实现 `UserService` → 重构 `AdminUsers` Handler
2. 实现 `AuthService` → 重构 `Auth` Handler
3. 实现 `TagService` → 重构 `AdminTags` Handler
4. 实现 `ResidentService` → 重构 `AdminResidents` Handler
5. ...

---

## 📚 参考示例

### 已实现的 Handler（参考）

1. **VitalFocusHandler** - 独立 Handler 类型
   - 文件：`internal/http/vital_focus_handlers.go`
   - 特点：直接操作 Redis，没有 Service 层（简单场景）

2. **TenantsHandler** - 独立 Handler 类型
   - 文件：`internal/http/admin_tenants_handlers.go`
   - 特点：使用 Repository，没有 Service 层（简单场景）

3. **AdminAPI** - 组合 Handler
   - 文件：`internal/http/admin_units_devices_handlers.go`
   - 特点：使用 Repository，没有 Service 层（简单场景）

### 新实现的 Handler（目标）

1. **RolesHandler** - 使用 RoleService
2. **RolePermissionsHandler** - 使用 RolePermissionService

