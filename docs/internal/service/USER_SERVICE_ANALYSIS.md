# User Service 深度分析文档

## 阶段 1：深度分析旧 Handler

### 文件信息
- **文件路径**: `internal/http/admin_users_handlers.go`
- **代码行数**: 1257 行
- **实现方式**: `StubHandler.AdminUsers` 方法

---

## 一、功能清单

### 1. ListUsers - 查询用户列表
**路由**: `GET /admin/api/v1/users`

**业务逻辑**:
1. 获取 tenant_id（从 Query 或 Header）
2. 获取当前用户信息（userID, role, branch_tag）
3. 权限检查：
   - 调用 `GetResourcePermission` 检查 `users` 资源的 `R` 权限
   - `AssignedOnly=true`: 只能查看自己（Caregiver/Nurse）
   - `BranchOnly=true`: 只能查看同 branch 的用户（Manager）
   - 否则：可以查看所有用户（Admin/IT）
4. 构建 SQL 查询：
   - 基础查询：`SELECT user_id, tenant_id, user_account, nickname, email, phone, role, status, alarm_levels, alarm_channels, alarm_scope, branch_tag, last_login_at, tags, preferences FROM users WHERE tenant_id = $1`
   - 权限过滤：
     - `AssignedOnly`: `AND user_id = $X`
     - `BranchOnly`: `AND (branch_tag IS NULL OR branch_tag = '-')` 或 `AND branch_tag = $X`
   - 搜索过滤：`AND (user_account ILIKE $X OR nickname ILIKE $X OR email ILIKE $X OR phone ILIKE $X)`
   - 排序：`ORDER BY user_account ASC`
5. 数据转换：
   - `alarm_levels`, `alarm_channels`: PostgreSQL Array → Go []string
   - `tags`: JSONB → []string
   - `preferences`: JSONB → any
   - `last_login_at`: sql.NullTime → RFC3339 string
   - 可选字段：nickname, email, phone, alarm_scope, branch_tag 使用 sql.NullString

**响应格式**:
```json
{
  "code": 2000,
  "data": {
    "items": [
      {
        "user_id": "...",
        "tenant_id": "...",
        "user_account": "...",
        "nickname": "...",
        "email": "...",
        "phone": "...",
        "role": "...",
        "status": "...",
        "alarm_levels": ["..."],
        "alarm_channels": ["..."],
        "alarm_scope": "...",
        "branch_tag": "...",
        "last_login_at": "2024-01-01T00:00:00Z",
        "tags": ["..."],
        "preferences": {...}
      }
    ],
    "total": 10
  }
}
```

---

### 2. CreateUser - 创建用户
**路由**: `POST /admin/api/v1/users`

**业务逻辑**:
1. 参数验证：
   - `user_account`（必填）
   - `role`（必填）
   - `password`（必填）
2. 权限检查：
   - 获取当前用户角色（从数据库，不信任 Header）
   - 系统角色检查：
     - `SystemAdmin`/`SystemOperator` 只能由 SystemAdmin 在 System tenant 中创建
   - 角色层级检查：
     - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以创建目标角色
     - 规则：只能创建同级或下级角色
3. 数据准备：
   - `user_account`: 转小写并去空格
   - `account_hash`: `SHA256(lower(user_account))`
   - `password_hash`: `SHA256(password)`（独立于 account）
   - `email_hash`: `SHA256(lower(email))`（如果 email 提供）
   - `phone_hash`: `SHA256(lower(phone))`（如果 phone 提供）
   - `status`: 默认 "active"
   - `alarm_levels`, `alarm_channels`: []any → pq.StringArray
   - `alarm_scope`: 根据角色设置默认值
     - `Caregiver`/`Nurse`: "ASSIGNED_ONLY"
     - `Manager`: "BRANCH"
     - 其他：NULL
   - `tags`: []any → JSONB ([]string)
4. 唯一性检查：
   - `checkEmailUniqueness`: 检查 email 是否已存在
   - `checkPhoneUniqueness`: 检查 phone 是否已存在
5. 插入数据库：
   ```sql
   INSERT INTO users (tenant_id, user_account, user_account_hash, password_hash, nickname, email, phone, email_hash, phone_hash, role, status, alarm_levels, alarm_channels, alarm_scope, tags)
   VALUES ($1, $2, $3, $4, NULLIF($5,''), NULLIF($6,''), NULLIF($7,''), $8, $9, $10, $11, $12, $13, $14, $15)
   RETURNING user_id::text
   ```
6. 同步 AuthStore（可选，用于开发环境）

**响应格式**:
```json
{
  "code": 2000,
  "data": {
    "user_id": "..."
  }
}
```

---

### 3. GetUser - 查询用户详情
**路由**: `GET /admin/api/v1/users/:id`

**业务逻辑**:
1. 权限检查：
   - 获取当前用户角色（从数据库）
   - `isViewingSelf`: 检查是否查看自己
   - 如果不是查看自己：
     - 获取目标用户角色
     - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以查看
2. 查询数据库：
   ```sql
   SELECT user_id::text, tenant_id::text, user_account, COALESCE(nickname,''), COALESCE(email,''), COALESCE(phone,''), role, COALESCE(status,'active'), COALESCE(alarm_levels, ARRAY[]::varchar[]), COALESCE(alarm_channels, ARRAY[]::varchar[]), alarm_scope, branch_tag, last_login_at, COALESCE(tags, '[]'::jsonb), COALESCE(preferences, '{}'::jsonb)
   FROM users
   WHERE tenant_id = $1 AND user_id::text = $2
   ```
3. 数据转换：同 ListUsers

**响应格式**:
```json
{
  "code": 2000,
  "data": {
    "user_id": "...",
    "tenant_id": "...",
    "user_account": "...",
    "nickname": "...",
    "email": "...",
    "phone": "...",
    "role": "...",
    "status": "...",
    "alarm_levels": ["..."],
    "alarm_channels": ["..."],
    "alarm_scope": "...",
    "branch_tag": "...",
    "last_login_at": "2024-01-01T00:00:00Z",
    "tags": ["..."],
    "preferences": {...}
  }
}
```

---

### 4. UpdateUser - 更新用户
**路由**: `PUT /admin/api/v1/users/:id`

**业务逻辑**:
1. 权限检查：
   - 获取当前用户角色（从数据库）
   - `isUpdatingSelf`: 检查是否更新自己
   - 确定更新字段：
     - `updatingRole`: role 不为空
     - `updatingStatus`: status 不为空
     - `updatingOtherFields`: nickname, email, phone, alarm_levels, alarm_channels, alarm_scope, tags, branch_tag
   - 权限规则：
     - 如果更新自己且只更新 password/email/phone：无限制
     - 如果更新其他用户或更新 role/status/otherFields：需要权限检查
   - 角色更新检查：
     - 系统角色：只能由 SystemAdmin 在 System tenant 中分配
     - 其他角色：调用 `canCreateRole` 检查
   - 管理权限检查：
     - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以管理目标用户
2. 字段处理：
   - `nickname`: 字符串，直接更新
   - `email`/`email_hash`: 复杂逻辑
     - 如果 `email_hash` 提供：
       - `email` 为 null：删除 email 但保留 hash（save 未勾选）
       - `email` 有值：同时更新 email 和 hash（save 已勾选）
     - 如果只有 `email` 提供（legacy）：
       - `email` 为 null：删除 email 和 hash
       - `email` 有值：计算 hash 并更新
   - `phone`/`phone_hash`: 同 email 逻辑
   - `role`: 字符串，直接更新
   - `status`: 验证值（active/disabled/left）
   - `alarm_levels`, `alarm_channels`: []any → pq.StringArray（仅当提供时更新）
   - `alarm_scope`: string → sql.NullString（仅当提供时更新）
   - `tags`: []any → JSONB（仅当提供时更新，允许清空）
   - `branch_tag`: string（空字符串表示 NULL）
3. 唯一性检查：
   - 如果更新 email 且不为空：`checkEmailUniqueness(tenantID, email, userID)`
   - 如果更新 phone 且不为空：`checkPhoneUniqueness(tenantID, phone, userID)`
4. 构建动态 UPDATE 查询：
   - 只更新提供的字段
   - 使用参数化查询防止 SQL 注入
5. 执行更新

**响应格式**:
```json
{
  "code": 2000,
  "data": {
    "success": true
  }
}
```

---

### 5. DeleteUser - 删除用户（软删除）
**路由**: `DELETE /admin/api/v1/users/:id` 或 `PUT /admin/api/v1/users/:id` with `{"_delete": true}`

**业务逻辑**:
1. 权限检查：
   - 获取当前用户角色（从数据库）
   - 获取目标用户角色
   - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以删除
2. 软删除：
   ```sql
   UPDATE users SET status = 'left' WHERE tenant_id = $1 AND user_id::text = $2
   ```

**响应格式**:
```json
{
  "code": 2000,
  "data": {
    "success": true
  }
}
```

---

### 6. ResetPassword - 重置密码
**路由**: `POST /admin/api/v1/users/:id/reset-password`

**业务逻辑**:
1. 权限检查：
   - 获取当前用户角色（从数据库）
   - `isResettingSelf`: 检查是否重置自己的密码
   - 如果不是重置自己：
     - 获取目标用户角色
     - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以重置
2. 参数验证：
   - `new_password`（必填）
3. 查询用户信息：
   - 获取 `user_account` 和 `role`（用于同步 AuthStore）
4. 密码哈希：
   - `password_hash`: `SHA256(new_password)`（独立于 account）
5. 更新数据库：
   ```sql
   UPDATE users SET password_hash = $3 WHERE tenant_id = $1 AND user_id::text = $2
   ```
6. 同步 AuthStore（可选）

**响应格式**:
```json
{
  "code": 2000,
  "data": {
    "success": true,
    "message": "ok"
  }
}
```

---

### 7. ResetPIN - 重置 PIN
**路由**: `POST /admin/api/v1/users/:id/reset-pin`

**业务逻辑**:
1. 权限检查：同 ResetPassword
2. 参数验证：
   - `new_pin`（必填）
   - 验证：必须是 4 位数字
3. PIN 哈希：
   - `pin_hash`: `SHA256(new_pin)`（独立于 account）
4. 更新数据库：
   ```sql
   UPDATE users SET pin_hash = $3 WHERE tenant_id = $1 AND user_id::text = $2
   ```

**响应格式**:
```json
{
  "code": 2000,
  "data": {
    "success": true
  }
}
```

---

## 二、关键业务逻辑

### 1. 角色层级权限检查

**函数**: `canCreateRole(currentRole, targetRole)`

**规则**:
- Level 1: SystemAdmin, SystemOperator（系统级）
- Level 2: Admin（租户管理员）
- Level 3: Manager, IT（业务管理层）
- Level 4: Nurse, Caregiver（操作层）
- Level 5: Resident, Family（用户角色）

**逻辑**:
- SystemAdmin/SystemOperator 只能由 SystemAdmin 创建（单独检查）
- 其他角色：只能创建同级或下级角色（`targetLevel >= currentLevel`）

---

### 2. 权限过滤

**函数**: `GetResourcePermission(db, ctx, role, resource, permission)`

**返回**: `PermissionCheck{AssignedOnly, BranchOnly}`

**规则**:
- `AssignedOnly=true`: 只能查看/管理自己（Caregiver/Nurse）
- `BranchOnly=true`: 只能查看/管理同 branch 的用户（Manager）
- 否则：可以查看/管理所有用户（Admin/IT）

---

### 3. Email/Phone 唯一性检查

**函数**: `checkEmailUniqueness(db, r, tenantID, email, excludeUserID)`
**函数**: `checkPhoneUniqueness(db, r, tenantID, phone, excludeUserID)`

**逻辑**:
- 查询数据库中是否存在相同的 email/phone（在同一 tenant 内）
- 如果 `excludeUserID` 提供，排除该用户（用于更新场景）
- 如果存在冲突，返回错误

---

### 4. 密码/PIN 哈希处理

**规则**:
- `account_hash`: `SHA256(lower(user_account))`
- `password_hash`: `SHA256(password)`（独立于 account）
- `pin_hash`: `SHA256(pin)`（独立于 account）
- `email_hash`: `SHA256(lower(email))`
- `phone_hash`: `SHA256(lower(phone))`

**注意**: password_hash 和 pin_hash 只依赖于密码/PIN 本身，不依赖于 account/email/phone

---

### 5. Email/Phone 更新逻辑

**复杂场景**:
- 前端可能发送 `email_hash` 和 `email` 两个字段
- `email_hash` 总是提供（如果 email 有值）
- `email` 可能为 null（表示不保存 email，但保留 hash 用于登录）

**处理逻辑**:
1. 如果 `email_hash` 提供：
   - `email` 为 null：删除 email 但保留 hash
   - `email` 有值：同时更新 email 和 hash
2. 如果只有 `email` 提供（legacy）：
   - `email` 为 null：删除 email 和 hash
   - `email` 有值：计算 hash 并更新

---

## 三、数据转换规则

### 1. PostgreSQL Array → Go []string
- `alarm_levels`: `pq.Array(&alarmLevels)`
- `alarm_channels`: `pq.Array(&alarmChannels)`

### 2. JSONB → Go 类型
- `tags`: JSONB → []string
- `preferences`: JSONB → any

### 3. sql.NullString → string
- `nickname`, `email`, `phone`, `alarm_scope`, `branch_tag`: 使用 `COALESCE` 或条件判断

### 4. sql.NullTime → string
- `last_login_at`: `time.RFC3339` 格式

---

## 四、错误处理

### 1. 唯一约束错误
- 检查 PostgreSQL unique constraint violation
- 返回友好的错误消息

### 2. 权限错误
- 角色层级检查失败
- 权限过滤限制

### 3. 参数验证错误
- 必填字段缺失
- 格式验证失败（如 PIN 必须是 4 位数字）

---

## 五、依赖函数

### 1. 权限检查函数
**位置**: `internal/http/permission_utils.go`

- `GetResourcePermission(db, ctx, roleCode, resourceType, permissionType)`: 查询资源权限配置
- `PermissionCheck`: 权限检查结果结构体
  - `AssignedOnly`: 是否仅限分配的资源
  - `BranchOnly`: 是否仅限同一 Branch 的资源

### 2. 唯一性检查函数
**位置**: `internal/http/util.go`

- `checkEmailUniqueness(db, r, tenantID, email, excludeUserID)`: 检查 email 唯一性
- `checkPhoneUniqueness(db, r, tenantID, phone, excludeUserID)`: 检查 phone 唯一性

### 3. 哈希函数
**位置**: `internal/http/auth_store.go`

- `HashAccount(account)`: `SHA256(lower(account))`
- `HashPassword(password)`: `SHA256(password)`（独立于 account）

### 4. 其他工具函数
**位置**: `internal/http/util.go`

- `checkUniqueConstraintError(err, field)`: 检查唯一约束错误
- `SystemTenantID()`: 获取系统租户 ID

---

## 六、下一步

1. 查找并理解所有依赖函数
2. 设计 Service 接口
3. 实现 Service 层
4. 编写测试
5. 实现 Handler 层

