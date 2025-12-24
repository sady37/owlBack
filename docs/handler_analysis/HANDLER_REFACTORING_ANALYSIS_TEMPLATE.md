# Handler 重构分析模板

## 📋 使用说明

在重构每个 Handler 之前，必须先完成以下分析：
1. 列出当前 Handler 的所有业务功能点
2. 分析每个功能点的复杂度
3. 拆解为 Service 方法
4. 拆解为 Handler 方法
5. 确认职责边界

---

## 📐 分析模板

### 第一步：当前 Handler 业务功能点分析

#### 1.1 Handler 基本信息

```
Handler 名称：AdminUsers
文件路径：internal/http/admin_users_handlers.go
当前行数：582 行
业务领域：用户管理
```

#### 1.2 业务功能点列表

| 功能点 | HTTP 方法 | 路径 | 功能描述 | 复杂度 | 当前实现行数 |
|--------|----------|------|----------|--------|-------------|
| 查询用户列表 | GET | `/admin/api/v1/users` | 支持搜索、分页、权限过滤 | 高 | ~150 |
| 查询用户详情 | GET | `/admin/api/v1/users/:id` | 获取单个用户信息 | 中 | ~50 |
| 创建用户 | POST | `/admin/api/v1/users` | 创建新用户，包含密码、角色验证 | 高 | ~100 |
| 更新用户 | PUT | `/admin/api/v1/users/:id` | 更新用户信息，包含角色、状态验证 | 高 | ~120 |
| 删除用户 | DELETE | `/admin/api/v1/users/:id` | 删除用户，检查依赖 | 中 | ~40 |
| 重置密码 | POST | `/admin/api/v1/users/:id/reset-password` | 重置用户密码 | 中 | ~60 |
| 重置 PIN | POST | `/admin/api/v1/users/:id/reset-pin` | 重置用户 PIN | 中 | ~60 |

**总计**：7 个功能点，582 行代码

#### 1.3 业务规则分析

**权限检查**：
- ✅ 查询用户列表：需要 R 权限，支持 assigned_only 和 branch_only 过滤
- ✅ 创建用户：需要 C 权限，需要 SystemAdmin 或 Admin 角色
- ✅ 更新用户：需要 U 权限，需要 SystemAdmin 或 Admin 角色
- ✅ 删除用户：需要 D 权限，需要 SystemAdmin 角色
- ✅ 重置密码：需要 U 权限，需要 SystemAdmin 或 Admin 角色

**业务规则验证**：
- ✅ 角色层级验证（SystemAdmin > Admin > Manager > ...）
- ✅ 密码强度验证
- ✅ 用户账号唯一性验证
- ✅ 租户一致性验证
- ✅ 依赖检查（删除前检查是否有关联数据）

**数据转换**：
- ✅ 前端格式 ↔ 领域模型（User）
- ✅ 密码哈希处理
- ✅ 角色代码转换

**业务编排**：
- ✅ 创建用户时同步标签到 tags_catalog
- ✅ 更新用户时同步标签到 tags_catalog
- ✅ 删除用户时清理标签关联

---

### 第二步：Service 方法拆解

#### 2.1 Service 接口设计

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

#### 2.2 Service 方法详细设计

| Service 方法 | 对应 Handler 功能点 | 职责 | 复杂度 |
|-------------|-------------------|------|--------|
| `ListUsers` | 查询用户列表 | 权限检查、业务规则验证、数据转换、调用 Repository | 高 |
| `GetUser` | 查询用户详情 | 权限检查、调用 Repository | 中 |
| `CreateUser` | 创建用户 | 权限检查、业务规则验证、数据转换、业务编排、调用 Repository | 高 |
| `UpdateUser` | 更新用户 | 权限检查、业务规则验证、数据转换、业务编排、调用 Repository | 高 |
| `DeleteUser` | 删除用户 | 权限检查、依赖检查、调用 Repository | 中 |
| `ResetPassword` | 重置密码 | 权限检查、密码验证、调用 Repository | 中 |
| `ResetPIN` | 重置 PIN | 权限检查、PIN 验证、调用 Repository | 中 |

#### 2.3 Service 请求/响应结构

```go
// ListUsersRequest 查询用户列表请求
type ListUsersRequest struct {
    TenantID    string
    UserID      string  // 当前用户ID（用于权限检查）
    UserRole    string  // 当前用户角色（用于权限检查）
    UserBranchTag *string  // 当前用户分支标签（用于权限检查）
    Search      string  // 搜索关键词
    Role        string  // 角色过滤
    Status      string  // 状态过滤
    Page        int
    Size        int
}

// ListUsersResponse 查询用户列表响应
type ListUsersResponse struct {
    Items []UserItem `json:"items"`
    Total int        `json:"total"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    TenantID    string
    UserID      string  // 当前用户ID（用于权限检查）
    UserRole    string  // 当前用户角色（用于权限检查）
    UserBranchTag *string  // 当前用户分支标签（用于权限检查）
    UserAccount string
    Password    string
    Nickname    string
    Role        string
    BranchTag   *string
    Phone       *string
    Email       *string
    Tags        []string
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
    UserID string `json:"user_id"`
}
```

---

### 第三步：Handler 方法拆解

#### 3.1 Handler 结构设计

```go
type UsersHandler struct {
    userService *service.UserService
    logger      *zap.Logger
}

func (h *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 路由分发
}
```

#### 3.2 Handler 方法详细设计

| Handler 方法 | 对应 Service 方法 | 职责 | 复杂度 |
|------------|------------------|------|--------|
| `ListUsers` | `UserService.ListUsers` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `GetUser` | `UserService.GetUser` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `CreateUser` | `UserService.CreateUser` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `UpdateUser` | `UserService.UpdateUser` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `DeleteUser` | `UserService.DeleteUser` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `ResetPassword` | `UserService.ResetPassword` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `ResetPIN` | `UserService.ResetPIN` | HTTP 参数解析、调用 Service、返回响应 | 低 |

#### 3.3 Handler 方法实现模板

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
        TenantID:     tenantID,
        UserID:       userID,
        UserRole:     userRole,
        UserBranchTag: userBranchTag,
        Search:       search,
        Role:         role,
        Status:       status,
        Page:         page,
        Size:         size,
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
```

---

### 第四步：职责边界确认

#### 4.1 Handler 职责

**只负责**：
- ✅ HTTP 请求/响应处理
- ✅ 参数解析和验证（HTTP 层面：类型、格式）
- ✅ 调用 Service
- ✅ 错误处理和日志记录

**不应该**：
- ❌ 直接操作数据库
- ❌ 业务规则验证（应该在 Service 层）
- ❌ 权限检查（应该在 Service 层）
- ❌ 数据转换（应该在 Service 层）
- ❌ 复杂业务逻辑（应该在 Service 层）

#### 4.2 Service 职责

**负责**：
- ✅ 权限检查（基于 role_permissions 表）
- ✅ 业务规则验证（角色层级、密码强度、唯一性等）
- ✅ 数据转换（前端格式 ↔ 领域模型）
- ✅ 业务编排（标签同步、依赖检查等）
- ✅ 调用 Repository

**不应该**：
- ❌ 直接操作数据库（应该通过 Repository）
- ❌ HTTP 请求/响应处理（应该在 Handler 层）

#### 4.3 Repository 职责

**负责**：
- ✅ 数据访问（CRUD 操作）
- ✅ 数据完整性验证（外键、唯一性约束等）
- ✅ SQL 查询优化

**不应该**：
- ❌ 业务规则验证（应该在 Service 层）
- ❌ 权限检查（应该在 Service 层）
- ❌ 数据转换（应该在 Service 层）

---

### 第五步：重构计划

#### 5.1 实施步骤

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
   - [ ] 编写 Handler + Service 集成测试
   - [ ] 运行所有测试

4. **路由注册**
   - [ ] 在 `router.go` 中添加注册方法
   - [ ] 在 `main.go` 中集成 Service 和 Handler

5. **验证和清理**
   - [ ] 手动测试 API 端点
   - [ ] 前端功能验证
   - [ ] 清理旧代码（可选）

#### 5.2 预估工作量

| 任务 | 预估时间 | 优先级 |
|------|---------|--------|
| Service 实现 | 4-6 小时 | 高 |
| Handler 实现 | 2-3 小时 | 高 |
| 测试编写 | 3-4 小时 | 高 |
| 集成和验证 | 2-3 小时 | 中 |
| **总计** | **11-16 小时** | |

---

## 📋 检查清单

### 分析阶段

- [ ] 列出所有业务功能点
- [ ] 分析每个功能点的复杂度
- [ ] 识别业务规则和权限检查
- [ ] 拆解为 Service 方法
- [ ] 拆解为 Handler 方法
- [ ] 确认职责边界
- [ ] 设计请求/响应结构
- [ ] 制定重构计划

### 实施阶段

- [ ] Service 接口定义
- [ ] Service 实现
- [ ] Service 测试
- [ ] Handler 实现
- [ ] Handler 测试
- [ ] 集成测试
- [ ] 路由注册
- [ ] 功能验证

---

## 📚 参考示例

### 已完成的分析（RoleService）

参考：`ROLE_SERVICE_HANDLER_IMPLEMENTATION.md`

**业务功能点**：
1. 查询角色列表
2. 创建角色
3. 更新角色
4. 更新角色状态
5. 删除角色

**Service 方法**：
- `ListRoles`
- `CreateRole`
- `UpdateRole`

**Handler 方法**：
- `ListRoles`
- `CreateRole`
- `UpdateRole`
- `UpdateRoleStatus`
- `DeleteRole`

---

## 🎯 使用流程

1. **选择要重构的 Handler**
2. **填写分析模板**（使用本文档）
3. **与团队确认**拆解方案
4. **实施重构**（按计划执行）
5. **验证和测试**
6. **更新文档**

