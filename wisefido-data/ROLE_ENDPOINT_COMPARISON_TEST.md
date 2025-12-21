# Role 端点对比测试

## 📋 测试目标

确保新 Handler 的响应格式与旧 Handler **完全一致**。

---

## 🔍 GET /admin/api/v1/roles 对比

### 测试场景 1：查询所有角色

**请求**：
```http
GET /admin/api/v1/roles
```

**旧 Handler 响应格式**（admin_roles_handlers.go:64）：
```json
{
  "status": "ok",
  "data": {
    "items": [
      {
        "role_id": "...",
        "tenant_id": null,
        "role_code": "...",
        "display_name": "...",
        "description": "...",
        "is_system": true,
        "is_active": true
      }
    ],
    "total": 10
  }
}
```

**新 Handler 响应格式**（admin_roles_handler.go:77）：
```json
{
  "status": "ok",
  "data": {
    "items": [
      {
        "role_id": "...",
        "tenant_id": null,
        "role_code": "...",
        "display_name": "...",
        "description": "...",
        "is_system": true,
        "is_active": true
      }
    ],
    "total": 10
  }
}
```

**对比结果**：
- ✅ `items`: 数组格式一致
- ✅ `total`: 整数格式一致
- ✅ `display_name`: 从 description 第一行提取，逻辑一致
- ✅ `tenant_id`: NULL 值处理一致

---

### 测试场景 2：搜索角色

**请求**：
```http
GET /admin/api/v1/roles?search=Admin
```

**旧 Handler 逻辑**：
- 使用 ILIKE 模糊匹配 role_code 或 description

**新 Handler 逻辑**：
- 通过 Repository 的 RolesFilter.Search 实现

**对比结果**：
- ✅ 搜索逻辑一致
- ✅ 响应格式一致

---

### 测试场景 3：分页查询

**请求**：
```http
GET /admin/api/v1/roles?page=1&size=20
```

**旧 Handler 逻辑**：
- 不支持分页（返回所有结果）

**新 Handler 逻辑**：
- 支持分页（page, size 参数）

**对比结果**：
- ✅ 响应格式一致
- ✅ 新增：分页支持（改进）

---

## 🔍 POST /admin/api/v1/roles 对比

### 测试场景 1：创建角色

**请求**：
```http
POST /admin/api/v1/roles
Content-Type: application/json

{
  "role_code": "TestRole",
  "display_name": "Test Role",
  "description": "Test Description"
}
```

**旧 Handler 响应格式**（admin_roles_handlers.go:108）：
```json
{
  "status": "ok",
  "data": {
    "role_id": "..."
  }
}
```

**新 Handler 响应格式**（admin_roles_handler.go:116）：
```json
{
  "status": "ok",
  "data": {
    "role_id": "..."
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致

---

## 🔍 PUT /admin/api/v1/roles/:id 对比

### 测试场景 1：更新角色描述

**请求**：
```http
PUT /admin/api/v1/roles/:id
Content-Type: application/json

{
  "display_name": "Updated Name",
  "description": "Updated Description"
}
```

**旧 Handler 响应格式**（admin_roles_handlers.go:227）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**新 Handler 响应格式**（admin_roles_handler.go:172）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致

---

### 测试场景 2：更新角色状态（通过 PUT）

**请求**：
```http
PUT /admin/api/v1/roles/:id
Content-Type: application/json

{
  "is_active": false
}
```

**旧 Handler 逻辑**（admin_roles_handlers.go:183-201）：
- 检查受保护角色
- 更新 is_active

**新 Handler 逻辑**（admin_roles_handler.go:144-147）：
- 通过 Service 层检查受保护角色
- 更新 is_active

**对比结果**：
- ✅ 响应格式一致
- ✅ 业务逻辑一致

---

### 测试场景 3：删除角色（通过 PUT _delete）

**请求**：
```http
PUT /admin/api/v1/roles/:id
Content-Type: application/json

{
  "_delete": true
}
```

**旧 Handler 响应格式**（admin_roles_handlers.go:179）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**新 Handler 响应格式**（admin_roles_handler.go:172）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致

---

## 🔍 PUT /admin/api/v1/roles/:id/status 对比

### 测试场景 1：更新角色状态

**请求**：
```http
PUT /admin/api/v1/roles/:id/status
Content-Type: application/json

{
  "is_active": false
}
```

**旧 Handler 响应格式**（admin_roles_handlers.go:142）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**新 Handler 响应格式**（admin_roles_handler.go:208）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致

---

### 测试场景 2：禁用受保护角色（应该失败）

**请求**：
```http
PUT /admin/api/v1/roles/:id/status
Content-Type: application/json

{
  "is_active": false
}
```

**前提条件**：role_code 为 "SystemAdmin"

**旧 Handler 错误响应**（admin_roles_handlers.go:190）：
```json
{
  "status": "fail",
  "message": "SystemAdmin is a critical system role and cannot be disabled"
}
```

**新 Handler 错误响应**（admin_roles_handler.go:203）：
```json
{
  "status": "fail",
  "message": "SystemAdmin is a critical system role and cannot be disabled"
}
```

**对比结果**：
- ✅ 错误响应格式一致
- ✅ 错误信息一致

---

## 🔍 DELETE /admin/api/v1/roles/:id 对比

### 测试场景 1：删除非系统角色

**请求**：
```http
DELETE /admin/api/v1/roles/:id
```

**旧 Handler 响应格式**（admin_roles_handlers.go:249）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**新 Handler 响应格式**（admin_roles_handler.go:236）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致

---

### 测试场景 2：删除系统角色（应该失败）

**请求**：
```http
DELETE /admin/api/v1/roles/:id
```

**前提条件**：is_system = true

**旧 Handler 错误响应**（admin_roles_handlers.go:241）：
```json
{
  "status": "fail",
  "message": "system roles cannot be deleted"
}
```

**新 Handler 错误响应**（admin_roles_handler.go:230）：
```json
{
  "status": "fail",
  "message": "system roles cannot be deleted"
}
```

**对比结果**：
- ✅ 错误响应格式一致
- ✅ 错误信息一致

---

## 📊 对比总结

### GET 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 查询所有角色 | ✅ | ✅ | ✅ 一致 |
| 搜索角色 | ✅ | ✅ | ✅ 一致 |
| 分页查询 | ❌ 不支持 | ✅ 支持 | ✅ 改进 |

### POST 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 创建角色 | ✅ | ✅ | ✅ 一致 |

### PUT 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 更新角色描述 | ✅ | ✅ | ✅ 一致 |
| 更新角色状态 | ✅ | ✅ | ✅ 一致 |
| 删除角色 | ✅ | ✅ | ✅ 一致 |

### PUT status 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 更新角色状态 | ✅ | ✅ | ✅ 一致 |
| 禁用受保护角色 | ✅ 拒绝 | ✅ 拒绝 | ✅ 一致 |

### DELETE 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 删除非系统角色 | ✅ | ✅ | ✅ 一致 |
| 删除系统角色 | ✅ 拒绝 | ✅ 拒绝 | ✅ 一致 |

---

## ✅ 验证结论

### 响应格式一致性：✅ **完全一致**

1. ✅ **GET 方法**：响应格式完全一致
2. ✅ **POST 方法**：响应格式完全一致
3. ✅ **PUT 方法**：响应格式完全一致
4. ✅ **PUT status 方法**：响应格式完全一致
5. ✅ **DELETE 方法**：响应格式完全一致

### 业务逻辑一致性：✅ **完全一致**

1. ✅ 搜索逻辑一致
2. ✅ 受保护角色检查一致
3. ✅ 系统角色限制一致
4. ✅ 错误处理一致

### 改进点：✅ **分页支持**

- ✅ 新 Handler 增加了分页支持（GET 方法）
- ✅ 这是改进，不是问题

---

## 🎯 最终结论

**新 Handler 与旧 Handler 的响应格式完全一致，业务逻辑完全一致。**

**✅ 可以安全替换旧 Handler**

