# AuthService 接口设计文档

## 📋 阶段 2：设计 Service 接口

### 设计原则

1. ✅ **职责边界清晰**：
   - Service 层：业务逻辑、规则验证、数据转换、业务编排
   - Repository 层：数据访问、SQL 查询
   - Handler 层：HTTP 请求/响应处理、参数解析

2. ✅ **强类型设计**：
   - 使用强类型 Request/Response 结构
   - 不使用 `map[string]any`

3. ✅ **错误处理**：
   - 返回明确的错误信息
   - 记录详细的日志

---

## 🔧 AuthService 接口定义

### 接口概览

```go
package service

import (
    "context"
    "wisefido-data/internal/domain"
)

// AuthService 认证授权服务接口
type AuthService interface {
    // 登录功能
    Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
    
    // 搜索机构功能
    SearchInstitutions(ctx context.Context, req SearchInstitutionsRequest) (*SearchInstitutionsResponse, error)
    
    // 密码重置功能（待实现）
    SendVerificationCode(ctx context.Context, req SendVerificationCodeRequest) (*SendVerificationCodeResponse, error)
    VerifyCode(ctx context.Context, req VerifyCodeRequest) (*VerifyCodeResponse, error)
    ResetPassword(ctx context.Context, req ResetPasswordRequest) (*ResetPasswordResponse, error)
}
```

---

## 📝 详细接口设计

### 1. Login - 用户登录

#### LoginRequest

```go
type LoginRequest struct {
    TenantID     string // 可选，如果为空则自动解析
    UserType     string // "staff" | "resident"，默认为 "staff"
    AccountHash  string // SHA256(account) 的 hex 编码，必填
    PasswordHash string // SHA256(password) 的 hex 编码，必填
    IPAddress    string // 客户端 IP（用于日志）
    UserAgent    string // 客户端 User-Agent（用于日志）
}
```

#### LoginResponse

```go
type LoginResponse struct {
    AccessToken  string  `json:"accessToken"`  // 访问令牌（占位符）
    RefreshToken string  `json:"refreshToken"` // 刷新令牌（占位符）
    UserID       string  `json:"userId"`       // 用户 ID
    UserAccount  string  `json:"user_account"` // 用户账号
    UserType     string  `json:"userType"`     // 用户类型
    Role         string  `json:"role"`         // 角色
    NickName     string  `json:"nickName"`     // 昵称
    TenantID     string  `json:"tenant_id"`    // 租户 ID
    TenantName   string  `json:"tenant_name"`  // 租户名称
    Domain       string  `json:"domain"`        // 域名
    HomePath     string  `json:"homePath"`      // 首页路径
    BranchTag    *string `json:"branchTag,omitempty"` // 分支标签（可选）
}
```

#### 业务逻辑职责

1. ✅ **参数验证**：
   - `AccountHash` 和 `PasswordHash` 必填
   - `UserType` 规范化（转换为小写，默认为 "staff"）

2. ✅ **Hash 解码和验证**：
   - 将 hex 字符串解码为 `[]byte`
   - 验证解码是否成功

3. ✅ **Tenant ID 自动解析**（如果为空）：
   - 根据 `UserType` 查询匹配的机构
   - 处理多机构匹配情况（0/1/>1）

4. ✅ **用户验证**：
   - 根据 `UserType` 查询不同的表
   - 优先级处理（email_hash > phone_hash > account_hash）
   - 状态验证（is_enabled, status='active', can_view_status）

5. ✅ **登录后处理**：
   - 更新 `last_login_at`（仅 staff）
   - 记录成功登录日志

6. ✅ **响应构建**：
   - 构建完整的登录响应

---

### 2. SearchInstitutions - 搜索机构

#### SearchInstitutionsRequest

```go
type SearchInstitutionsRequest struct {
    AccountHash  string // SHA256(account) 的 hex 编码，必填
    PasswordHash string // SHA256(password) 的 hex 编码，必填
    UserType     string // "staff" | "resident"，默认为 "staff"
}
```

#### SearchInstitutionsResponse

```go
type Institution struct {
    ID          string `json:"id"`          // 机构 ID
    Name        string `json:"name"`        // 机构名称
    Domain      string `json:"domain,omitempty"` // 机构域名（可选）
    AccountType string `json:"accountType"` // 账号类型（email/phone/account）
}

type SearchInstitutionsResponse struct {
    Institutions []Institution `json:"institutions"`
}
```

#### 业务逻辑职责

1. ✅ **参数验证**：
   - `AccountHash` 和 `PasswordHash` 必填
   - `UserType` 规范化

2. ✅ **Hash 解码和验证**：
   - 将 hex 字符串解码为 `[]byte`
   - 如果解码失败，返回空数组

3. ✅ **查询匹配的机构**：
   - 根据 `UserType` 查询不同的表
   - 优先级处理
   - 状态验证

4. ✅ **机构信息补充**：
   - 查询机构详细信息（tenant_name, domain）
   - 特殊处理 System tenant

5. ✅ **安全机制**：
   - 只返回匹配账号和密码的机构
   - 如果无匹配，返回空数组

---

### 3. SendVerificationCode - 发送验证码（待实现）

#### SendVerificationCodeRequest

```go
type SendVerificationCodeRequest struct {
    Account   string // 账号（email/phone/userAccount）
    UserType  string // "staff" | "resident"
    TenantID  string // 租户 ID（可选）
    TenantName string // 租户名称（可选）
}
```

#### SendVerificationCodeResponse

```go
type SendVerificationCodeResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}
```

#### 业务逻辑职责

⚠️ **待实现**：需要设计验证码生成、存储和发送逻辑

---

### 4. VerifyCode - 验证验证码（待实现）

#### VerifyCodeRequest

```go
type VerifyCodeRequest struct {
    Account    string // 账号
    Code       string // 验证码
    UserType   string // "staff" | "resident"
    TenantID   string // 租户 ID（可选）
    TenantName string // 租户名称（必填）
}
```

#### VerifyCodeResponse

```go
type VerifyCodeResponse struct {
    Success    bool   `json:"success"`
    Token      string `json:"token,omitempty"` // 验证令牌（用于重置密码）
    Message    string `json:"message,omitempty"`
}
```

#### 业务逻辑职责

⚠️ **待实现**：需要设计验证码验证逻辑

---

### 5. ResetPassword - 重置密码（待实现）

#### ResetPasswordRequest

```go
type ResetPasswordRequest struct {
    Token       string // 验证令牌（从 VerifyCode 获取）
    NewPassword string // 新密码（明文，后端会进行 hash）
    UserType    string // "staff" | "resident"
}
```

#### ResetPasswordResponse

```go
type ResetPasswordResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}
```

#### 业务逻辑职责

⚠️ **待实现**：需要设计密码重置逻辑

---

## 🔗 Repository 依赖

### 需要的 Repository 接口

#### 现有方法（可直接使用）

1. ✅ **UsersRepository**（已存在）
   - `GetUserByEmail` - 根据 email_hash 查询用户
   - `GetUserByPhone` - 根据 phone_hash 查询用户
   - `GetUserByAccount` - 根据 user_account 查询用户

2. ✅ **ResidentsRepository**（已存在）
   - `GetResidentByEmail` - 根据 email_hash 查询住户
   - `GetResidentByPhone` - 根据 phone_hash 查询住户
   - `GetResidentByAccount` - 根据 resident_account_hash 查询住户
   - `GetResidentContacts` - 获取住户联系人列表

3. ✅ **TenantsRepository**（已存在）
   - `GetTenant` - 获取租户信息
   - `ListTenants` - 列出租户列表

#### 需要新增的方法（用于登录查询）

由于登录查询逻辑复杂（需要优先级排序、JOIN 多表、状态检查），需要在 Repository 层创建专门的登录查询方法：

1. ⚠️ **UsersRepository**（需要新增）
   - `GetUserForLogin` - 根据 tenant_id, account_hash, password_hash 查询用户（支持优先级，返回完整信息包括 tenant_name, domain, branch_tag）
   - `SearchTenantsForUserLogin` - 根据 account_hash, password_hash 搜索匹配的机构（用于 tenant_id 自动解析）
   - `UpdateUserLastLogin` - 更新 last_login_at

2. ⚠️ **ResidentsRepository**（需要新增）
   - `GetResidentForLogin` - 根据 tenant_id, account_hash, password_hash 查询住户（支持优先级，返回完整信息）
   - `GetResidentContactForLogin` - 根据 tenant_id, account_hash, password_hash 查询联系人（支持优先级，返回完整信息）
   - `SearchTenantsForResidentLogin` - 根据 account_hash, password_hash 搜索匹配的机构（用于 tenant_id 自动解析，包含 resident_contacts 和 residents 两步查询）

**注意**：这些方法需要：
- 支持优先级排序（email_hash > phone_hash > account_hash）
- JOIN tenants 表获取 tenant_name, domain
- JOIN units 表获取 branch_tag（对于 resident）
- 状态检查（is_enabled, status='active', can_view_status）
- 返回完整的登录信息

---

## 📊 职责边界确认

### Service 层职责

1. ✅ **业务逻辑**：
   - 参数验证和规范化
   - Hash 解码和验证
   - Tenant ID 自动解析
   - 多机构匹配处理
   - 优先级处理（email_hash > phone_hash > account_hash）
   - 状态验证（is_enabled, status='active', can_view_status）
   - 登录后处理（更新 last_login_at）
   - 日志记录

2. ✅ **数据转换**：
   - Hex 字符串 ↔ []byte
   - 用户信息字段映射
   - 响应格式构建

3. ✅ **业务编排**：
   - 多步骤查询流程
   - 条件分支处理
   - 错误处理

### Repository 层职责

1. ✅ **数据访问**：
   - 执行 SQL 查询
   - 返回领域模型

2. ✅ **查询优化**：
   - 使用索引
   - 优化查询性能

### Handler 层职责

1. ✅ **HTTP 请求/响应处理**：
   - 参数解析（支持多种格式）
   - 响应格式构建
   - 错误响应处理

2. ✅ **日志记录**：
   - 记录请求信息（IP, User-Agent）
   - 记录错误信息

---

## ✅ 阶段 2 完成

**Service 接口已设计，职责边界已确认，可以进入阶段 3：实现 Service**

