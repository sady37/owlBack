# UserService Handler 重构分析

## 📋 第一步：当前 Handler 业务功能点分析

### 1.1 Handler 基本信息

```
Handler 名称：AdminUsers
文件路径：internal/http/admin_users_handlers.go
当前行数：582 行
业务领域：用户管理
```

### 1.2 业务功能点列表

| 功能点 | HTTP 方法 | 路径 | 功能描述 | 复杂度 | 当前实现行数 |
|--------|----------|------|----------|--------|-------------|
| 查询用户列表 | GET | `/admin/api/v1/users` | 支持搜索、分页、权限过滤（assigned_only, branch_only） | 高 | ~150 |
| 查询用户详情 | GET | `/admin/api/v1/users/:id` | 获取单个用户信息（包含标签） | 中 | ~50 |
| 创建用户 | POST | `/admin/api/v1/users` | 创建新用户，包含密码、角色验证、标签同步 | 高 | ~100 |
| 更新用户 | PUT | `/admin/api/v1/users/:id` | 更新用户信息，包含角色、状态验证、标签同步 | 高 | ~120 |
| 删除用户 | DELETE | `/admin/api/v1/users/:id` | 删除用户，检查依赖、清理标签 | 中 | ~40 |
| 重置密码 | POST | `/admin/api/v1/users/:id/reset-password` | 重置用户密码，权限检查 | 中 | ~60 |
| 重置 PIN | POST | `/admin/api/v1/users/:id/reset-pin` | 重置用户 PIN，权限检查 | 中 | ~60 |

**总计**：7 个功能点，582 行代码

### 1.3 业务规则分析

#### 权限检查

| 功能点 | 权限要求 | 特殊规则 |
|--------|---------|---------|
| 查询用户列表 | R 权限 | 支持 assigned_only（仅分配的用户）和 branch_only（仅同分支）过滤 |
| 查询用户详情 | R 权限 | 支持 assigned_only 和 branch_only 过滤 |
| 创建用户 | C 权限 | 需要 SystemAdmin 或 Admin 角色 |
| 更新用户 | U 权限 | 需要 SystemAdmin 或 Admin 角色，不能修改自己的角色 |
| 删除用户 | D 权限 | 需要 SystemAdmin 角色，不能删除自己 |
| 重置密码 | U 权限 | 需要 SystemAdmin 或 Admin 角色 |
| 重置 PIN | U 权限 | 需要 SystemAdmin 或 Admin 角色 |

#### 业务规则验证

1. **角色层级验证**
   - SystemAdmin > Admin > Manager > Caregiver > IT > Nurse > Resident > Family
   - 不能创建比自己层级更高的角色
   - 不能修改比自己层级更高的用户

2. **密码强度验证**
   - 最小长度：8 字符
   - 必须包含字母和数字
   - 创建时必填，更新时可选

3. **用户账号唯一性验证**
   - 同一租户内账号唯一
   - 使用 `user_account_hash` 进行哈希匹配

4. **租户一致性验证**
   - 所有操作必须在同一租户内
   - 不能跨租户操作

5. **依赖检查**
   - 删除前检查是否有关联数据（residents, caregivers 等）
   - 如果有关联数据，不允许删除

#### 数据转换

1. **前端格式 ↔ 领域模型**
   - `User` 领域模型 ↔ 前端 `User` 格式
   - 标签数组 ↔ `tags` JSONB 字段
   - 密码哈希处理

2. **密码哈希**
   - 使用 `HashPassword` 函数
   - 存储为 `password_hash` 字节数组

3. **角色代码转换**
   - 前端角色名称 ↔ 数据库角色代码

#### 业务编排

1. **标签同步**
   - 创建用户时：同步标签到 `tags_catalog`（调用 `upsert_tag_to_catalog`）
   - 更新用户时：同步标签到 `tags_catalog`
   - 删除用户时：清理标签关联（调用 `drop_object_from_all_tags`）

2. **依赖检查**
   - 删除前检查 `residents` 表中的 `caregiver_id`
   - 删除前检查其他关联数据

---

## 📐 第二步：Service 方法拆解

### 2.1 Service 接口设计

```go
type UserService interface {
    // 查询
    ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error)
    GetUser(ctx context.Context, req GetUserRequest) (*UserItem, error)
    
    // 创建
    CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)
    
    // 更新
    UpdateUser(ctx context.Context, req UpdateUserRequest) error
    
    // 删除
    DeleteUser(ctx context.Context, req DeleteUserRequest) error
    
    // 密码管理
    ResetPassword(ctx context.Context, req ResetPasswordRequest) error
    ResetPIN(ctx context.Context, req ResetPINRequest) error
}
```

### 2.2 Service 方法详细设计

| Service 方法 | 对应 Handler 功能点 | 职责 | 复杂度 |
|-------------|-------------------|------|--------|
| `ListUsers` | 查询用户列表 | 权限检查（R 权限，assigned_only, branch_only）、数据转换、调用 Repository | 高 |
| `GetUser` | 查询用户详情 | 权限检查（R 权限，assigned_only, branch_only）、数据转换、调用 Repository | 中 |
| `CreateUser` | 创建用户 | 权限检查（C 权限，角色层级）、业务规则验证（密码强度、账号唯一性）、数据转换、业务编排（标签同步）、调用 Repository | 高 |
| `UpdateUser` | 更新用户 | 权限检查（U 权限，角色层级）、业务规则验证（不能修改自己角色）、数据转换、业务编排（标签同步）、调用 Repository | 高 |
| `DeleteUser` | 删除用户 | 权限检查（D 权限，不能删除自己）、依赖检查、业务编排（清理标签）、调用 Repository | 中 |
| `ResetPassword` | 重置密码 | 权限检查（U 权限）、密码验证、调用 Repository | 中 |
| `ResetPIN` | 重置 PIN | 权限检查（U 权限）、PIN 验证、调用 Repository | 中 |

### 2.3 Service 请求/响应结构

```go
// ListUsersRequest 查询用户列表请求
type ListUsersRequest struct {
    TenantID      string
    UserID        string   // 当前用户ID（用于权限检查）
    UserRole      string   // 当前用户角色（用于权限检查）
    UserBranchTag *string  // 当前用户分支标签（用于权限检查）
    Search        string   // 搜索关键词（账号、昵称、手机、邮箱）
    Role          string   // 角色过滤
    Status        string   // 状态过滤（active/inactive）
    Page          int
    Size          int
}

// ListUsersResponse 查询用户列表响应
type ListUsersResponse struct {
    Items []UserItem `json:"items"`
    Total int        `json:"total"`
}

// UserItem 用户项（前端格式）
type UserItem struct {
    UserID      string   `json:"user_id"`
    TenantID    string   `json:"tenant_id"`
    UserAccount string   `json:"user_account"`
    Nickname    string   `json:"nickname"`
    Role        string   `json:"role"`
    BranchTag   *string  `json:"branch_tag"`
    Phone       *string  `json:"phone"`
    Email       *string  `json:"email"`
    Status      string   `json:"status"`
    Tags        []string `json:"tags"`
}

// GetUserRequest 查询用户详情请求
type GetUserRequest struct {
    TenantID      string
    UserID        string   // 要查询的用户ID
    CurrentUserID string   // 当前用户ID（用于权限检查）
    CurrentUserRole string // 当前用户角色（用于权限检查）
    CurrentUserBranchTag *string // 当前用户分支标签（用于权限检查）
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    TenantID      string
    UserID        string   // 当前用户ID（用于权限检查）
    UserRole      string   // 当前用户角色（用于权限检查）
    UserBranchTag *string  // 当前用户分支标签（用于权限检查）
    UserAccount   string
    Password      string
    Nickname      string
    Role          string
    BranchTag     *string
    Phone         *string
    Email         *string
    Tags          []string
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
    UserID string `json:"user_id"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
    TenantID      string
    UserID        string   // 要更新的用户ID
    CurrentUserID string   // 当前用户ID（用于权限检查）
    CurrentUserRole string // 当前用户角色（用于权限检查）
    CurrentUserBranchTag *string // 当前用户分支标签（用于权限检查）
    Nickname      *string
    Role          *string
    BranchTag     *string
    Phone         *string
    Email         *string
    Status        *string
    Tags          *[]string
    Password      *string  // 可选，更新密码
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
    TenantID      string
    UserID        string   // 要删除的用户ID
    CurrentUserID string   // 当前用户ID（用于权限检查）
    CurrentUserRole string // 当前用户角色（用于权限检查）
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
    TenantID      string
    UserID        string   // 要重置密码的用户ID
    CurrentUserID string   // 当前用户ID（用于权限检查）
    CurrentUserRole string // 当前用户角色（用于权限检查）
    NewPassword   string
}

// ResetPINRequest 重置 PIN 请求
type ResetPINRequest struct {
    TenantID      string
    UserID        string   // 要重置 PIN 的用户ID
    CurrentUserID string   // 当前用户ID（用于权限检查）
    CurrentUserRole string // 当前用户角色（用于权限检查）
    NewPIN        string
}
```

---

## 🔧 第三步：Handler 方法拆解

### 3.1 Handler 结构设计

```go
type UsersHandler struct {
    userService *service.UserService
    logger      *zap.Logger
}

func (h *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 路由分发
    switch {
    case r.URL.Path == "/admin/api/v1/users" && r.Method == http.MethodGet:
        h.ListUsers(w, r)
    case strings.HasPrefix(r.URL.Path, "/admin/api/v1/users/") && r.Method == http.MethodGet:
        h.GetUser(w, r)
    case r.URL.Path == "/admin/api/v1/users" && r.Method == http.MethodPost:
        h.CreateUser(w, r)
    case strings.HasPrefix(r.URL.Path, "/admin/api/v1/users/") && r.Method == http.MethodPut:
        h.UpdateUser(w, r)
    case strings.HasPrefix(r.URL.Path, "/admin/api/v1/users/") && r.Method == http.MethodDelete:
        h.DeleteUser(w, r)
    case strings.HasSuffix(r.URL.Path, "/reset-password") && r.Method == http.MethodPost:
        h.ResetPassword(w, r)
    case strings.HasSuffix(r.URL.Path, "/reset-pin") && r.Method == http.MethodPost:
        h.ResetPIN(w, r)
    default:
        w.WriteHeader(http.StatusNotFound)
    }
}
```

### 3.2 Handler 方法详细设计

| Handler 方法 | 对应 Service 方法 | 职责 | 复杂度 |
|------------|------------------|------|--------|
| `ListUsers` | `UserService.ListUsers` | HTTP 参数解析、获取当前用户信息、调用 Service、返回响应 | 低 |
| `GetUser` | `UserService.GetUser` | HTTP 参数解析、获取当前用户信息、调用 Service、返回响应 | 低 |
| `CreateUser` | `UserService.CreateUser` | HTTP 参数解析、获取当前用户信息、调用 Service、返回响应 | 低 |
| `UpdateUser` | `UserService.UpdateUser` | HTTP 参数解析、获取当前用户信息、调用 Service、返回响应 | 低 |
| `DeleteUser` | `UserService.DeleteUser` | HTTP 参数解析、获取当前用户信息、调用 Service、返回响应 | 低 |
| `ResetPassword` | `UserService.ResetPassword` | HTTP 参数解析、获取当前用户信息、调用 Service、返回响应 | 低 |
| `ResetPIN` | `UserService.ResetPIN` | HTTP 参数解析、获取当前用户信息、调用 Service、返回响应 | 低 |

### 3.3 Handler 方法实现示例

```go
// ListUsers 查询用户列表
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 参数解析和验证
    tenantID, ok := h.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    
    userID := r.Header.Get("X-User-Id")
    userRole := r.Header.Get("X-User-Role")
    userBranchTag := h.getUserBranchTag(ctx, tenantID, userID) // 从数据库查询
    
    search := strings.TrimSpace(r.URL.Query().Get("search"))
    role := strings.TrimSpace(r.URL.Query().Get("role"))
    status := strings.TrimSpace(r.URL.Query().Get("status"))
    page := parseInt(r.URL.Query().Get("page"), 1)
    size := parseInt(r.URL.Query().Get("size"), 20)
    
    // 2. 调用 Service
    req := service.ListUsersRequest{
        TenantID:       tenantID,
        UserID:         userID,
        UserRole:       userRole,
        UserBranchTag:  userBranchTag,
        Search:         search,
        Role:           role,
        Status:         status,
        Page:           page,
        Size:           size,
    }
    
    resp, err := h.userService.ListUsers(ctx, req)
    if err != nil {
        h.logger.Error("ListUsers failed", zap.Error(err))
        writeJSON(w, http.StatusOK, Fail(err.Error()))
        return
    }
    
    // 3. 返回响应
    writeJSON(w, http.StatusOK, Ok(resp))
}

// CreateUser 创建用户
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 参数解析和验证
    tenantID, ok := h.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    
    userID := r.Header.Get("X-User-Id")
    userRole := r.Header.Get("X-User-Role")
    userBranchTag := h.getUserBranchTag(ctx, tenantID, userID)
    
    var payload struct {
        UserAccount string   `json:"user_account"`
        Password    string   `json:"password"`
        Nickname    string   `json:"nickname"`
        Role        string   `json:"role"`
        BranchTag   *string  `json:"branch_tag"`
        Phone       *string  `json:"phone"`
        Email       *string  `json:"email"`
        Tags        []string `json:"tags"`
    }
    if err := readBodyJSON(r, 1<<20, &payload); err != nil {
        writeJSON(w, http.StatusOK, Fail("invalid body"))
        return
    }
    
    // 2. 调用 Service
    req := service.CreateUserRequest{
        TenantID:      tenantID,
        UserID:        userID,
        UserRole:      userRole,
        UserBranchTag: userBranchTag,
        UserAccount:   payload.UserAccount,
        Password:      payload.Password,
        Nickname:      payload.Nickname,
        Role:          payload.Role,
        BranchTag:     payload.BranchTag,
        Phone:         payload.Phone,
        Email:         payload.Email,
        Tags:          payload.Tags,
    }
    
    resp, err := h.userService.CreateUser(ctx, req)
    if err != nil {
        h.logger.Error("CreateUser failed", zap.Error(err))
        writeJSON(w, http.StatusOK, Fail(err.Error()))
        return
    }
    
    // 3. 返回响应
    writeJSON(w, http.StatusOK, Ok(resp))
}
```

---

## ✅ 第四步：职责边界确认

### 4.1 Handler 职责

**只负责**：
- ✅ HTTP 请求/响应处理
- ✅ 参数解析和验证（HTTP 层面：类型、格式）
- ✅ 获取当前用户信息（从 Header 和数据库）
- ✅ 调用 Service
- ✅ 错误处理和日志记录

**不应该**：
- ❌ 直接操作数据库
- ❌ 业务规则验证（应该在 Service 层）
- ❌ 权限检查（应该在 Service 层）
- ❌ 数据转换（应该在 Service 层）
- ❌ 复杂业务逻辑（应该在 Service 层）

### 4.2 Service 职责

**负责**：
- ✅ 权限检查（基于 role_permissions 表，assigned_only, branch_only）
- ✅ 业务规则验证（角色层级、密码强度、账号唯一性、租户一致性）
- ✅ 数据转换（前端格式 ↔ 领域模型，密码哈希）
- ✅ 业务编排（标签同步、依赖检查）
- ✅ 调用 Repository

**不应该**：
- ❌ 直接操作数据库（应该通过 Repository）
- ❌ HTTP 请求/响应处理（应该在 Handler 层）

### 4.3 Repository 职责

**负责**：
- ✅ 数据访问（CRUD 操作）
- ✅ 数据完整性验证（外键、唯一性约束等）
- ✅ SQL 查询优化

**不应该**：
- ❌ 业务规则验证（应该在 Service 层）
- ❌ 权限检查（应该在 Service 层）
- ❌ 数据转换（应该在 Service 层）

---

## 📋 第五步：重构计划

### 5.1 实施步骤

1. **创建 Service 接口和实现**
   - [ ] 定义 Service 接口（`user_service.go`）
   - [ ] 实现所有 Service 方法
   - [ ] 编写 Service 单元测试

2. **创建 Handler**
   - [ ] 定义 Handler 结构（`admin_users_handler.go`）
   - [ ] 实现所有 Handler 方法
   - [ ] 编写 Handler 单元测试

3. **集成测试**
   - [ ] 编写 Service + Repository 集成测试
   - [ ] 编写 Handler + Service 集成测试
   - [ ] 运行所有测试

4. **路由注册**
   - [ ] 在 `router.go` 中添加注册方法
   - [ ] 在 `main.go` 中集成 Service 和 Handler

5. **验证和清理**
   - [ ] 手动测试 API 端点
   - [ ] 前端功能验证
   - [ ] 清理旧代码（可选）

### 5.2 预估工作量

| 任务 | 预估时间 | 优先级 |
|------|---------|--------|
| Service 实现 | 6-8 小时 | 高 |
| Handler 实现 | 3-4 小时 | 高 |
| 测试编写 | 4-5 小时 | 高 |
| 集成和验证 | 3-4 小时 | 中 |
| **总计** | **16-21 小时** | |

---

## 📚 参考

- `HANDLER_REFACTORING_ANALYSIS_TEMPLATE.md` - Handler 重构分析模板
- `ROLE_SERVICE_HANDLER_IMPLEMENTATION.md` - Role Service & Handler 实现示例

