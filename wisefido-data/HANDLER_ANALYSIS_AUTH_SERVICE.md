# Auth Handler 深度分析文档

## 📋 阶段 1：深度分析旧 Handler

### 文件信息

- **文件路径**: `wisefido-data/internal/http/auth_handlers.go`
- **函数名**: `StubHandler.Auth`
- **代码行数**: 887 行
- **复杂度**: ⭐⭐⭐⭐⭐ (极高)

---

## 🔍 端点清单

### 1. POST /auth/api/v1/login

**功能**: 用户登录

**代码位置**: 15-600

**业务逻辑复杂度**: ⭐⭐⭐⭐⭐

---

### 2. GET /auth/api/v1/institutions/search

**功能**: 搜索机构（根据账号和密码查找匹配的机构列表）

**代码位置**: 601-862

**业务逻辑复杂度**: ⭐⭐⭐⭐

---

### 3. POST /auth/api/v1/forgot-password/send-code

**功能**: 发送验证码（待实现）

**代码位置**: 863-869

**业务逻辑复杂度**: ⭐ (仅返回错误)

---

### 4. POST /auth/api/v1/forgot-password/verify-code

**功能**: 验证验证码（待实现）

**代码位置**: 870-876

**业务逻辑复杂度**: ⭐ (仅返回错误)

---

### 5. POST /auth/api/v1/forgot-password/reset

**功能**: 重置密码（待实现）

**代码位置**: 877-883

**业务逻辑复杂度**: ⭐ (仅返回错误)

---

## 📝 详细业务逻辑清单

### 1. POST /auth/api/v1/login

#### 1.1 参数解析（15-87行）

**业务逻辑**：
1. ✅ **支持多种参数格式**：
   - 从 JSON Body 获取参数
   - 支持 `{params: {...}}` 包装格式
   - 从 Query 参数获取（fallback）
   - 参数优先级：Body > Query

2. ✅ **参数列表**：
   - `tenant_id` (string, 可选)
   - `userType` (string, 可选，默认为 "staff")
   - `accountHash` (string, 必填) - SHA256(account) 的 hex 编码
   - `passwordHash` (string, 必填) - SHA256(password) 的 hex 编码

3. ✅ **参数验证**：
   - `accountHash` 不能为空
   - `passwordHash` 不能为空
   - 如果为空，记录警告日志并返回 "missing credentials"

4. ✅ **参数规范化**：
   - `userType` 转换为小写并 trim
   - 如果为空，默认为 "staff"
   - `accountHash` 和 `passwordHash` 都 trim

#### 1.2 Hash 解码和验证（94-126行）

**业务逻辑**：
1. ✅ **Hex 解码**：
   - `accountHash` 从 hex 字符串解码为 `[]byte`
   - `passwordHash` 从 hex 字符串解码为 `[]byte`
   - 如果解码失败或长度为 0，记录警告日志并返回 "invalid credentials"

2. ✅ **安全说明**：
   - `account_hash` 和 `password_hash` 是独立的
   - `account_hash = SHA256(account)`
   - `password_hash = SHA256(password)`（不包含 account）

#### 1.3 Tenant ID 自动解析（128-299行）

**业务逻辑**：
1. ✅ **触发条件**：
   - 如果 `tenant_id` 为空且数据库可用

2. ✅ **查询逻辑（根据 userType）**：

   **resident 类型**：
   - **Step 1**: 查询 `resident_contacts` 表
     - 匹配条件：`password_hash = $2` AND (`email_hash = $1` OR `phone_hash = $1`)
     - 优先级：`email_hash` > `phone_hash`
     - 状态检查：`is_enabled = true` AND `can_view_status = true`
   - **Step 2**: 如果 Step 1 无匹配，查询 `residents` 表
     - 匹配条件：`password_hash = $2` AND (`email_hash = $1` OR `phone_hash = $1` OR `resident_account_hash = $1`)
     - 优先级：`email_hash` > `phone_hash` > `resident_account_hash`
     - 状态检查：`status = 'active'` AND `can_view_status = true`

   **staff 类型**：
   - 查询 `users` 表
     - 匹配条件：`password_hash = $2` AND (`email_hash = $1` OR `phone_hash = $1` OR `user_account_hash = $1`)
     - 优先级：`email_hash` > `phone_hash` > `user_account_hash`
     - 状态检查：`status = 'active'`

3. ✅ **结果处理**：
   - 0 个匹配：返回 "invalid credentials"
   - 1 个匹配：自动设置 `tenant_id`
   - >1 个匹配：返回 "Multiple institutions found, please select one"

#### 1.4 用户验证和登录（301-536行）

**业务逻辑**：
1. ✅ **tenant_id 验证**：
   - 如果 `tenant_id` 为空，返回 "tenant_id is required"

2. ✅ **根据 userType 查询用户**：

   **resident 类型**（317-459行）：
   - **Step 1**: 查询 `resident_contacts` 表
     - 匹配条件：`tenant_id = $1` AND `password_hash = $3` AND (`email_hash = $2` OR `phone_hash = $2`)
     - 优先级：`email_hash` > `phone_hash`
     - 状态检查：`is_enabled = true` AND `can_view_status = true`
     - 返回字段：`contact_id`, `resident_id`, `slot`, `contact_first_name`, `contact_last_name`, `role`, `is_enabled`, `tenant_name`, `domain`, `branch_tag`, `account_type`
   - **Step 2**: 如果 Step 1 无匹配，查询 `residents` 表
     - 匹配条件：`tenant_id = $1` AND `password_hash = $3` AND (`email_hash = $2` OR `phone_hash = $2` OR `resident_account_hash = $2`)
     - 优先级：`email_hash` > `phone_hash` > `resident_account_hash`
     - 状态检查：`status = 'active'` AND `can_view_status = true`
     - 返回字段：`resident_id`, `resident_account`, `nickname`, `role`, `status`, `tenant_name`, `domain`, `branch_tag`, `account_type`

   **staff 类型**（460-528行）：
   - 查询 `users` 表
     - 匹配条件：`tenant_id = $1` AND `password_hash = $3` AND (`email_hash = $2` OR `phone_hash = $2` OR `user_account_hash = $2`)
     - 优先级：`email_hash` > `phone_hash` > `user_account_hash`
     - 状态检查：`status = 'active'`
     - 返回字段：`user_id`, `user_account`, `nickname`, `role`, `status`, `tenant_name`, `domain`, `branch_tag`, `account_type`

3. ✅ **状态验证**：
   - resident_contact: 检查 `is_enabled = true`
   - resident: 检查 `status = 'active'`
   - staff: 检查 `status = 'active'`
   - 如果状态不符合，记录警告日志并返回 "user is not active"

4. ✅ **用户信息处理**：
   - resident_contact: `userAccount = contact_id`, `nickName = first + last` 或 `role`
   - resident: `userAccount = resident_account`, `nickName = nickname`
   - staff: `userAccount = user_account`, `nickName = nickname`

5. ✅ **Fallback 处理**：
   - 如果数据库不可用，检查 `AuthStore`（已禁用）
   - 如果都不可用，返回 "db auth not configured"

#### 1.5 登录后处理（549-599行）

**业务逻辑**：
1. ✅ **nickName 默认值**：
   - 如果 `nickName` 为空，使用 `role` 或 `userAccount`

2. ✅ **更新 last_login_at**：
   - 仅对 staff 用户更新 `users.last_login_at = NOW()`

3. ✅ **日志记录**：
   - 记录成功登录日志（包含 user_id, user_account, user_type, tenant_id, tenant_name, role, ip_address, user_agent, login_time）

4. ✅ **响应构建**：
   - `accessToken`: "stub-access-token"（占位符）
   - `refreshToken`: "stub-refresh-token"（占位符）
   - `userId`: 用户 ID
   - `user_account`: 用户账号
   - `userType`: 用户类型
   - `role`: 角色
   - `nickName`: 昵称
   - `tenant_id`: 租户 ID
   - `tenant_name`: 租户名称
   - `domain`: 域名
   - `homePath`: "/monitoring/overview"
   - `branchTag`: 分支标签（如果存在）

---

### 2. GET /auth/api/v1/institutions/search

#### 2.1 参数解析（608-616行）

**业务逻辑**：
1. ✅ **参数列表**：
   - `accountHash` (string, 必填) - 从 Query 参数获取
   - `passwordHash` (string, 必填) - 从 Query 参数获取
   - `userType` (string, 可选，默认为 "staff") - 从 Query 参数获取

2. ✅ **参数规范化**：
   - `userType` 转换为小写并 trim
   - 如果为空，默认为 "staff"
   - `accountHash` 和 `passwordHash` 都 trim

#### 2.2 Hash 解码和验证（620-630行）

**业务逻辑**：
1. ✅ **Hex 解码**：
   - `accountHash` 从 hex 字符串解码为 `[]byte`
   - `passwordHash` 从 hex 字符串解码为 `[]byte`
   - 如果解码失败或长度为 0，返回空数组 `[]`

#### 2.3 查询匹配的机构（631-813行）

**业务逻辑**：
1. ✅ **查询逻辑（根据 userType）**：

   **resident 类型**（634-726行）：
   - **Step 1**: 查询 `resident_contacts` 表
     - 匹配条件：`password_hash = $2` AND (`email_hash = $1` OR `phone_hash = $1`)
     - 优先级：`email_hash` > `phone_hash`
     - 状态检查：`is_enabled = true` AND `can_view_status = true`
   - **Step 2**: 如果 Step 1 无匹配，查询 `residents` 表
     - 匹配条件：`password_hash = $2` AND (`email_hash = $1` OR `phone_hash = $1` OR `resident_account_hash = $1`)
     - 优先级：`email_hash` > `phone_hash` > `resident_account_hash`
     - 状态检查：`status = 'active'` AND `can_view_status = true`

   **staff 类型**（727-756行）：
   - 查询 `users` 表
     - 匹配条件：`password_hash = $2` AND (`email_hash = $1` OR `phone_hash = $1` OR `user_account_hash = $1`)
     - 优先级：`email_hash` > `phone_hash` > `user_account_hash`
     - 状态检查：`status = 'active'`

2. ✅ **结果处理**：
   - 查询返回 `tenant_id` 和 `account_type`
   - 如果查询失败，返回空数组 `[]`

3. ✅ **机构信息补充**（777-813行）：
   - 如果 `Tenants` 服务可用，查询机构详细信息
   - 返回格式：`{id, name, accountType}` 或 `{id, name, domain, accountType}`
   - 特殊处理：System tenant 直接返回 `{id: SystemTenantID(), name: "System", accountType: ...}`
   - 如果 `Tenants` 服务不可用，只返回 `{id, accountType}`

4. ✅ **安全机制**：
   - 只返回匹配账号和密码的机构（防止信息泄露）
   - 如果无匹配，返回空数组（不返回所有机构）

---

### 3. POST /auth/api/v1/forgot-password/send-code

**业务逻辑**：
- ⚠️ **待实现**：目前只返回 "database not available"

---

### 4. POST /auth/api/v1/forgot-password/verify-code

**业务逻辑**：
- ⚠️ **待实现**：目前只返回 "database not available"

---

### 5. POST /auth/api/v1/forgot-password/reset

**业务逻辑**：
- ⚠️ **待实现**：目前只返回 "database not available"

---

## 📊 业务逻辑复杂度分析

### 登录功能（POST /auth/api/v1/login）

**复杂度维度**：

1. ✅ **权限检查复杂度**: ⭐⭐⭐
   - 需要检查用户状态（is_enabled, status='active', can_view_status）
   - 需要检查用户类型（staff/resident）

2. ✅ **业务规则验证复杂度**: ⭐⭐⭐⭐⭐
   - 多种参数格式支持
   - Hash 解码和验证
   - Tenant ID 自动解析（多机构匹配处理）
   - 根据 userType 查询不同表
   - resident 类型需要两步查询（resident_contacts → residents）
   - 优先级处理（email_hash > phone_hash > account_hash）
   - 状态验证（多个状态字段）

3. ✅ **数据转换复杂度**: ⭐⭐⭐
   - Hex 字符串 ↔ []byte 转换
   - 用户信息字段映射（不同 userType 字段不同）
   - 响应格式构建

4. ✅ **业务编排复杂度**: ⭐⭐⭐⭐
   - 多步骤查询流程
   - 条件分支处理（userType, tenant_id 是否存在）
   - 错误处理和日志记录
   - 登录后处理（更新 last_login_at）

5. ✅ **Handler 代码行数**: ⭐⭐⭐⭐⭐
   - 587 行代码（登录功能）

**总计**: 5/5 维度为"复杂" → **需要 Service**

---

### 搜索机构功能（GET /auth/api/v1/institutions/search）

**复杂度维度**：

1. ✅ **权限检查复杂度**: ⭐⭐
   - 需要检查用户状态（is_enabled, status='active', can_view_status）

2. ✅ **业务规则验证复杂度**: ⭐⭐⭐⭐
   - Hash 解码和验证
   - 根据 userType 查询不同表
   - resident 类型需要两步查询
   - 优先级处理
   - 状态验证

3. ✅ **数据转换复杂度**: ⭐⭐
   - Hex 字符串 ↔ []byte 转换
   - 机构信息补充和格式化

4. ✅ **业务编排复杂度**: ⭐⭐⭐
   - 多步骤查询流程
   - 条件分支处理
   - 安全机制（防止信息泄露）

5. ✅ **Handler 代码行数**: ⭐⭐⭐
   - 262 行代码

**总计**: 5/5 维度为"复杂" → **需要 Service**

---

## 🎯 需要 Service 层的原因

1. ✅ **复杂的业务规则**：
   - 多种用户类型（staff/resident）
   - 多种账号类型（email/phone/account）
   - 优先级处理
   - 状态验证

2. ✅ **复杂的数据查询**：
   - 多表查询（users, residents, resident_contacts）
   - 条件分支查询
   - Tenant ID 自动解析

3. ✅ **安全机制**：
   - Hash 验证
   - 状态检查
   - 信息泄露防护

4. ✅ **业务编排**：
   - 多步骤流程
   - 错误处理
   - 日志记录

---

## 📝 待实现功能

1. ⚠️ **发送验证码**（POST /auth/api/v1/forgot-password/send-code）
2. ⚠️ **验证验证码**（POST /auth/api/v1/forgot-password/verify-code）
3. ⚠️ **重置密码**（POST /auth/api/v1/forgot-password/reset）

---

## ✅ 阶段 1 完成

**业务逻辑清单已创建，可以进入阶段 2：设计 Service 接口**

