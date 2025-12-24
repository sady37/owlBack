# User AccountSettings API 设计文档

## 1. 设计原则

- **统一接口**：与 Resident/Contact 的 AccountSettings API 保持一致的结构
- **权限控制**：用户只能查看和更新自己的账户设置
- **Staff 特性**：Staff 的 save_email 和 save_phone 总是 true（不可改）

## 2. Service 层设计

### 2.1 GetAccountSettings

#### 接口定义
```go
GetAccountSettings(ctx context.Context, req GetAccountSettingsRequest) (*GetAccountSettingsResponse, error)
```

#### 请求结构
```go
type GetAccountSettingsRequest struct {
    TenantID      string // 必填
    UserID        string // 必填
    CurrentUserID string // 当前用户 ID（用于权限检查）
}
```

#### 响应结构
```go
type GetAccountSettingsResponse struct {
    ID        string  // UUID: user_id（前端需要）
    Account   string  // user_account
    Nickname  string  // 昵称
    Email     *string // 邮箱（可选，nil 表示不存在）
    Phone     *string // 电话（可选，nil 表示不存在）
    Role      string  // 角色代码（前端需要，用于判断使用哪种表）
    SaveEmail bool    // 是否保存 email 明文（Staff 总是 true）
    SavePhone bool    // 是否保存 phone 明文（Staff 总是 true）
}
```

#### 处理逻辑
1. **权限检查**：`CurrentUserID == UserID`（只能查看自己的账户设置）
2. **获取用户信息**：从 `users` 表获取用户信息
3. **构建响应**：
   - `ID = user.UserID`
   - `Account = user.UserAccount`
   - `Nickname = user.Nickname`（如果存在）
   - `Email = user.Email`（如果存在，否则 nil）
   - `Phone = user.Phone`（如果存在，否则 nil）
   - `Role = user.Role`
   - `SaveEmail = true`（Staff 总是保存）
   - `SavePhone = true`（Staff 总是保存）

#### save_email 和 save_phone 生成逻辑
- **如果 email/phone 存在**：`save_email/save_phone = true`（已保存明文）
- **如果 email/phone 不存在**：`save_email/save_phone = true`（Staff 总是保存，即使当前不存在，将来添加时也会保存）

### 2.2 UpdateAccountSettings

#### 接口定义
```go
UpdateAccountSettings(ctx context.Context, req UpdateAccountSettingsRequest) (*UpdateAccountSettingsResponse, error)
```

#### 请求结构
```go
type UpdateAccountSettingsRequest struct {
    TenantID      string  // 必填
    UserID        string  // 必填
    CurrentUserID string  // 当前用户 ID（用于权限检查）
    PasswordHash  *string // 可选：密码 hash（nil 表示不更新）
    Email         *string // 可选：邮箱（nil 表示不更新，空字符串或 null 表示删除明文但保留 hash）
    EmailHash     *string // 可选：邮箱 hash（nil 表示不更新，空字符串或 null 表示删除 hash）
    Phone         *string // 可选：电话（nil 表示不更新，空字符串或 null 表示删除明文但保留 hash）
    PhoneHash     *string // 可选：电话 hash（nil 表示不更新，空字符串或 null 表示删除 hash）
    // 注意：Staff 不需要 save_email 和 save_phone，因为总是 true
}
```

#### 响应结构
```go
type UpdateAccountSettingsResponse struct {
    Success bool   // 是否成功
    Message string // 消息（可选，用于错误详情）
}
```

#### 处理逻辑
1. **参数验证**：`TenantID`, `UserID`, `CurrentUserID` 必填
2. **权限检查**：`CurrentUserID == UserID`（只能更新自己的账户设置）
3. **获取目标用户信息**
4. **构建更新对象**（只更新提供的字段）：
   - **密码更新**：如果 `PasswordHash != nil && *PasswordHash != ""`，更新密码
   - **Email 更新**：
     - 如果 `Email == nil`：不更新 email
     - 如果 `Email != nil && *Email == ""`：删除 email 明文，但保留 hash（如果 `EmailHash == nil`）
     - 如果 `Email != nil && *Email != ""`：更新 email 明文
     - 如果 `EmailHash != nil && *EmailHash != ""`：更新 email_hash
     - 如果 `EmailHash != nil && *EmailHash == ""`：删除 email_hash（同时删除 email 明文）
   - **Phone 更新**：逻辑与 Email 相同
5. **唯一性检查**：如果更新了 email 或 phone，检查唯一性（排除自己）
6. **执行更新**：Repository 层在事务中处理

## 3. Handler 层设计

### 3.1 GetAccountSettings Handler

#### 路径
`GET /admin/api/v1/users/:id/account-settings`

#### 处理逻辑
1. 从 URL 解析 `userID`
2. 从 Header 获取 `X-User-Id`（作为 `CurrentUserID`）
3. 从请求获取 `tenantID`
4. 调用 Service 层 `GetAccountSettings`
5. 返回 JSON 响应：
```json
{
  "success": true,
  "data": {
    "id": "user_id",
    "account": "user_account",
    "nickname": "nickname",
    "email": "email@example.com",  // 可选
    "phone": "1234567890",          // 可选
    "role": "Admin",
    "save_email": true,
    "save_phone": true
  }
}
```

### 3.2 UpdateAccountSettings Handler

#### 路径
`PUT /admin/api/v1/users/:id/account-settings`

#### 处理逻辑
1. 从 URL 解析 `userID`
2. 从 Header 获取 `X-User-Id`（作为 `CurrentUserID`）
3. 从请求获取 `tenantID`
4. 解析请求 Body（JSON）：
   - `password_hash`（可选）
   - `email`（可选，可能是 `null` 或空字符串）
   - `email_hash`（可选）
   - `phone`（可选，可能是 `null` 或空字符串）
   - `phone_hash`（可选）
5. **处理 null 值**：
   - 如果 `email` 是 `null`，转换为空字符串（表示删除明文但保留 hash）
   - 如果 `phone` 是 `null`，转换为空字符串（表示删除明文但保留 hash）
6. 调用 Service 层 `UpdateAccountSettings`
7. 返回 JSON 响应：
```json
{
  "success": true,
  "message": "Account settings updated successfully"
}
```

## 4. 与前端接口对应关系

### 前端期望的 AccountSettings 接口
```typescript
interface AccountSettings {
  id: string                    // UUID: user_id
  account?: string              // user_account
  nickname: string              // 昵称
  email?: string                // 邮箱（可选）
  phone?: string                // 电话（可选）
  save_email?: boolean          // 是否保存 email 明文（Staff 总是 true）
  save_phone?: boolean          // 是否保存 phone 明文（Staff 总是 true）
  role: string                  // 角色代码
}
```

### 字段映射
- `id` ← `GetAccountSettingsResponse.ID`
- `account` ← `GetAccountSettingsResponse.Account`
- `nickname` ← `GetAccountSettingsResponse.Nickname`
- `email` ← `GetAccountSettingsResponse.Email`（如果存在）
- `phone` ← `GetAccountSettingsResponse.Phone`（如果存在）
- `save_email` ← `GetAccountSettingsResponse.SaveEmail`（总是 true）
- `save_phone` ← `GetAccountSettingsResponse.SavePhone`（总是 true）
- `role` ← `GetAccountSettingsResponse.Role`

## 5. 与 Resident/Contact 的对比

| 特性 | Staff (User) | Resident/Contact |
|------|-------------|------------------|
| save_email/save_phone | 总是 true（不可改） | 根据数据库状态（可能 true/false） |
| email/phone 存储 | users 表（明文） | residents/resident_phi 或 resident_contacts 表 |
| 占位符处理 | 不适用（总是保存） | 需要处理占位符（***@***, xxx-xxx-xxxx） |

## 6. 实施步骤

1. 修改 `user_service.go`：
   - 更新 `GetAccountSettingsResponse` 结构（添加 `ID`, `Role`, `SaveEmail`, `SavePhone`）
   - 更新 `GetAccountSettings` 实现（添加字段生成逻辑）
   - 更新 `UpdateAccountSettings` 实现（处理 null 值和只删除明文但保留 hash 的逻辑）

2. 修改 `user_handler.go`：
   - 更新 `GetAccountSettings` handler（返回所有字段）
   - 更新 `UpdateAccountSettings` handler（处理 null 值）

3. 测试验证：
   - 测试 GetAccountSettings 返回所有字段
   - 测试 UpdateAccountSettings 处理各种场景（更新密码、更新 email/phone、删除 email/phone 但保留 hash）

