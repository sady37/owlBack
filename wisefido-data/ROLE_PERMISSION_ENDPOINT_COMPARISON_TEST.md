# RolePermission 端点对比测试

## 📋 测试目标

确保新 Handler 的响应格式与旧 Handler **完全一致**。

---

## 🔍 GET /admin/api/v1/role-permissions 对比

### 测试场景 1：查询所有权限

**请求**：
```http
GET /admin/api/v1/role-permissions
```

**旧 Handler 响应格式**（admin_role_permissions_handlers.go:84）：
```json
{
  "status": "ok",
  "data": {
    "items": [
      {
        "permission_id": "...",
        "tenant_id": null,
        "role_code": "...",
        "resource_type": "...",
        "permission_type": "read",
        "scope": "all",
        "branch_only": false,
        "is_active": true
      }
    ],
    "total": 10
  }
}
```

**新 Handler 响应格式**（admin_role_permissions_handler.go:85）：
```json
{
  "status": "ok",
  "data": {
    "items": [
      {
        "permission_id": "...",
        "tenant_id": null,
        "role_code": "...",
        "resource_type": "...",
        "permission_type": "read",
        "scope": "all",
        "branch_only": false,
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
- ✅ `permission_type`: 前端格式一致（read/create/update/delete）
- ✅ `scope`: 前端格式一致（all/assigned_only）
- ✅ `tenant_id`: NULL 值处理一致

---

### 测试场景 2：按 role_code 过滤

**请求**：
```http
GET /admin/api/v1/role-permissions?role_code=SystemAdmin
```

**对比结果**：
- ✅ 过滤逻辑一致
- ✅ 响应格式一致

---

### 测试场景 3：分页查询

**请求**：
```http
GET /admin/api/v1/role-permissions?page=1&size=20
```

**旧 Handler 逻辑**：
- 不支持分页（返回所有结果）

**新 Handler 逻辑**：
- 支持分页（page, size 参数）

**对比结果**：
- ✅ 响应格式一致
- ✅ 新增：分页支持（改进）

---

## 🔍 POST /admin/api/v1/role-permissions 对比

### 测试场景 1：创建权限

**请求**：
```http
POST /admin/api/v1/role-permissions
Content-Type: application/json

{
  "role_code": "SystemAdmin",
  "resource_type": "user",
  "permission_type": "read",
  "scope": "all",
  "branch_only": false
}
```

**旧 Handler 响应格式**（admin_role_permissions_handlers.go:137）：
```json
{
  "status": "ok",
  "data": {
    "permission_id": "..."
  }
}
```

**新 Handler 响应格式**（admin_role_permissions_handler.go:130）：
```json
{
  "status": "ok",
  "data": {
    "permission_id": "..."
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致

---

### 测试场景 2：创建 "manage" 类型权限

**请求**：
```http
POST /admin/api/v1/role-permissions
Content-Type: application/json

{
  "role_code": "SystemAdmin",
  "resource_type": "user",
  "permission_type": "manage",
  "scope": "all"
}
```

**旧 Handler 逻辑**：
- "manage" 类型不展开（只创建一条记录）

**新 Handler 逻辑**：
- "manage" 类型不展开（只创建一条记录）

**对比结果**：
- ✅ 逻辑一致（注意：批量操作中 "manage" 会展开）

---

## 🔍 POST /admin/api/v1/role-permissions/batch 对比

### 测试场景 1：批量创建权限

**请求**：
```http
POST /admin/api/v1/role-permissions/batch
Content-Type: application/json

{
  "role_code": "SystemAdmin",
  "permissions": [
    {
      "resource_type": "user",
      "permission_type": "manage",
      "scope": "all",
      "branch_only": false,
      "is_active": true
    }
  ]
}
```

**旧 Handler 响应格式**（admin_role_permissions_handlers.go:249-260）：
```json
{
  "status": "ok",
  "data": {
    "success_count": 4,
    "failed_count": 0
  }
}
```

**新 Handler 响应格式**（admin_role_permissions_handler.go:180）：
```json
{
  "status": "ok",
  "data": {
    "success_count": 4,
    "failed_count": 0
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致
- ✅ "manage" 类型展开逻辑一致（展开为 R, C, U, D）

---

## 🔍 GET /admin/api/v1/role-permissions/resource-types 对比

### 测试场景 1：获取资源类型列表

**请求**：
```http
GET /admin/api/v1/role-permissions/resource-types
```

**旧 Handler 响应格式**（admin_role_permissions_handlers.go:280）：
```json
{
  "status": "ok",
  "data": {
    "resource_types": ["user", "resident", "unit", ...]
  }
}
```

**新 Handler 响应格式**（admin_role_permissions_handler.go:200）：
```json
{
  "status": "ok",
  "data": {
    "resource_types": ["user", "resident", "unit", ...]
  }
}
```

**对比结果**：
- ✅ 响应格式完全一致

---

## 🔍 PUT /admin/api/v1/role-permissions/:id 对比

### 测试场景 1：更新权限

**请求**：
```http
PUT /admin/api/v1/role-permissions/:id
Content-Type: application/json

{
  "scope": "assigned_only",
  "branch_only": true
}
```

**旧 Handler 响应格式**（admin_role_permissions_handlers.go:320）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**新 Handler 响应格式**（admin_role_permissions_handler.go:240）：
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

## 🔍 PUT /admin/api/v1/role-permissions/:id/status 对比

### 测试场景 1：禁用权限（is_active=false）

**请求**：
```http
PUT /admin/api/v1/role-permissions/:id/status
Content-Type: application/json

{
  "is_active": false
}
```

**旧 Handler 逻辑**：
- 删除权限（记录存在表示激活）

**新 Handler 逻辑**：
- 删除权限（记录存在表示激活）

**对比结果**：
- ✅ 逻辑一致
- ✅ 响应格式一致

---

## 🔍 DELETE /admin/api/v1/role-permissions/:id 对比

### 测试场景 1：删除权限

**请求**：
```http
DELETE /admin/api/v1/role-permissions/:id
```

**旧 Handler 响应格式**（admin_role_permissions_handlers.go:376）：
```json
{
  "status": "ok",
  "data": {
    "success": true
  }
}
```

**新 Handler 响应格式**（admin_role_permissions_handler.go:310）：
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

## 📊 对比总结

### GET 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 查询所有权限 | ✅ | ✅ | ✅ 一致 |
| 按 role_code 过滤 | ✅ | ✅ | ✅ 一致 |
| 分页查询 | ❌ 不支持 | ✅ 支持 | ✅ 改进 |

### POST 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 创建权限 | ✅ | ✅ | ✅ 一致 |
| 创建 "manage" 类型 | ✅ 不展开 | ✅ 不展开 | ✅ 一致 |

### POST batch 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 批量创建权限 | ✅ | ✅ | ✅ 一致 |
| "manage" 类型展开 | ✅ 展开为 R,C,U,D | ✅ 展开为 R,C,U,D | ✅ 一致 |

### GET resource-types 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 获取资源类型列表 | ✅ | ✅ | ✅ 一致 |

### PUT 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 更新权限 | ✅ | ✅ | ✅ 一致 |

### PUT status 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 禁用权限 | ✅ 删除记录 | ✅ 删除记录 | ✅ 一致 |

### DELETE 方法

| 测试场景 | 旧 Handler | 新 Handler | 状态 |
|---------|-----------|-----------|------|
| 删除权限 | ✅ | ✅ | ✅ 一致 |

---

## ✅ 验证结论

### 响应格式一致性：✅ **完全一致**

1. ✅ **GET 方法**：响应格式完全一致
2. ✅ **POST 方法**：响应格式完全一致
3. ✅ **POST batch 方法**：响应格式完全一致
4. ✅ **GET resource-types 方法**：响应格式完全一致
5. ✅ **PUT 方法**：响应格式完全一致
6. ✅ **PUT status 方法**：响应格式完全一致
7. ✅ **DELETE 方法**：响应格式完全一致

### 业务逻辑一致性：✅ **完全一致**

1. ✅ 权限类型转换一致
2. ✅ Scope 转换一致
3. ✅ "manage" 类型展开逻辑一致（批量操作中）
4. ✅ 权限检查一致
5. ✅ 错误处理一致

### 改进点：✅ **分页支持**

- ✅ 新 Handler 增加了分页支持（GET 方法）
- ✅ 这是改进，不是问题

---

## 🎯 最终结论

**新 Handler 与旧 Handler 的响应格式完全一致，业务逻辑完全一致。**

**✅ 可以安全替换旧 Handler**

