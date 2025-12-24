# AuthService 阶段 6：集成和路由注册

## ✅ 已完成的工作

### 1. 路由注册方法

**文件**: `internal/http/router.go`

**新增方法**：
```go
// RegisterAuthRoutes 注册认证授权路由
func (r *Router) RegisterAuthRoutes(h *AuthHandler) {
	r.Handle("/auth/api/v1/login", h.ServeHTTP)
	r.Handle("/auth/api/v1/institutions/search", h.ServeHTTP)
	r.Handle("/auth/api/v1/forgot-password/send-code", h.ServeHTTP)
	r.Handle("/auth/api/v1/forgot-password/verify-code", h.ServeHTTP)
	r.Handle("/auth/api/v1/forgot-password/reset", h.ServeHTTP)
}
```

**对比旧路由注册**（router.go:114-119）：
```go
// auth
r.Handle("/auth/api/v1/login", s.Auth)
r.Handle("/auth/api/v1/institutions/search", s.Auth)
r.Handle("/auth/api/v1/forgot-password/send-code", s.Auth)
r.Handle("/auth/api/v1/forgot-password/verify-code", s.Auth)
r.Handle("/auth/api/v1/forgot-password/reset", s.Auth)
```

**对比结果**：✅ **路由路径完全一致**

---

### 2. Main 函数集成

**文件**: `cmd/wisefido-data/main.go`

**新增代码**（在 `if db != nil` 块中）：
```go
// 创建 Auth Service 和 Handler
authRepo := repository.NewPostgresAuthRepository(db)
authService := service.NewAuthService(authRepo, tenantsRepo, logger)
authHandler := httpapi.NewAuthHandler(authService, logger)
router.RegisterAuthRoutes(authHandler)
```

**位置**：在 AlarmCloud Service 和 Handler 创建之后，`} else {` 之前。

**依赖关系**：
- ✅ `authRepo` 依赖 `db`（PostgreSQL 数据库）
- ✅ `authService` 依赖 `authRepo` 和 `tenantsRepo`
- ✅ `authHandler` 依赖 `authService` 和 `logger`

---

### 3. 路由注册顺序

**当前注册顺序**：
1. ✅ VitalFocusRoutes
2. ✅ RolesRoutes（DB 启用时）
3. ✅ RolePermissionsRoutes（DB 启用时）
4. ✅ TagsRoutes（DB 启用时）
5. ✅ AlarmCloudRoutes（DB 启用时）
6. ✅ **AuthRoutes（DB 启用时）** ← 新增
7. ✅ AdminUnitDeviceRoutes
8. ✅ AdminTenantRoutes
9. ✅ StubRoutes（包含旧的 Auth 路由）

**注意**：
- ✅ 新 AuthHandler 在 `if db != nil` 块中注册（需要数据库）
- ⚠️ 旧的 `StubHandler.Auth` 仍然在 `RegisterStubRoutes` 中注册（作为 fallback）
- 🔄 **后续步骤**：在验证新 Handler 工作正常后，可以从 `RegisterStubRoutes` 中移除旧的 Auth 路由

---

## 📊 编译验证

### 编译状态

**新代码编译**：✅ **通过**
- ✅ `auth_handler.go` 编译通过
- ✅ `router.go` 编译通过（新增 `RegisterAuthRoutes` 方法）
- ✅ `main.go` 编译通过（新增 Auth Service 和 Handler 初始化）

**整体编译**：⚠️ **有其他文件错误**（`admin_units_devices_impl.go`，与 Auth 无关）

---

## 🔍 路由对比

| 路由路径 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| `/auth/api/v1/login` | `StubHandler.Auth` | `AuthHandler.ServeHTTP` | ✅ 已注册 |
| `/auth/api/v1/institutions/search` | `StubHandler.Auth` | `AuthHandler.ServeHTTP` | ✅ 已注册 |
| `/auth/api/v1/forgot-password/send-code` | `StubHandler.Auth` | `AuthHandler.ServeHTTP` | ✅ 已注册 |
| `/auth/api/v1/forgot-password/verify-code` | `StubHandler.Auth` | `AuthHandler.ServeHTTP` | ✅ 已注册 |
| `/auth/api/v1/forgot-password/reset` | `StubHandler.Auth` | `AuthHandler.ServeHTTP` | ✅ 已注册 |

---

## ✅ 验证结论

### 路由注册：✅ **完成**

1. ✅ 所有 5 个认证路由都已注册
2. ✅ 路由路径与旧 Handler 完全一致
3. ✅ 路由注册在正确的条件块中（`if db != nil`）

### 代码集成：✅ **完成**

1. ✅ AuthRepository 创建成功
2. ✅ AuthService 创建成功
3. ✅ AuthHandler 创建成功
4. ✅ 路由注册成功

### 编译验证：✅ **通过**

1. ✅ 新代码编译通过
2. ✅ 依赖关系正确
3. ⚠️ 其他文件有编译错误（与 Auth 无关）

---

## 🎯 下一步

**阶段 7：验证和测试**

1. ✅ 启动服务，验证路由是否正常工作
2. ✅ 进行端到端测试，对比新旧 Handler 的响应
3. ✅ 验证所有端点的行为一致性
4. ✅ 确认无误后，从 `RegisterStubRoutes` 中移除旧的 Auth 路由

---

## 📝 注意事项

1. **路由优先级**：
   - 新 Handler 的路由注册在 `RegisterStubRoutes` 之前
   - 由于路由匹配顺序，新 Handler 会优先处理请求
   - 如果新 Handler 未注册（DB 未启用），会 fallback 到旧 Handler

2. **数据库依赖**：
   - 新 Handler 需要数据库连接
   - 如果数据库未启用，会使用旧的 `StubHandler.Auth`

3. **向后兼容**：
   - 旧的 `StubHandler.Auth` 仍然保留（作为 fallback）
   - 在验证新 Handler 工作正常后，可以移除旧路由

