# User Service 接口设计文档

## 一、Service 接口定义

```go
type UserService interface {
    // 查询
    ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error)
    GetUser(ctx context.Context, req GetUserRequest) (*GetUserResponse, error)

    // 创建
    CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)

    // 更新
    UpdateUser(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error)

    // 删除
    DeleteUser(ctx context.Context, req DeleteUserRequest) (*DeleteUserResponse, error)

    // 密码和 PIN 管理
    ResetPassword(ctx context.Context, req ResetPasswordRequest) (*ResetPasswordResponse, error)
    ResetPIN(ctx context.Context, req ResetPINRequest) (*ResetPINResponse, error)
}
```

---

## 二、请求/响应 DTO 详细说明

### 1. ListUsers - 查询用户列表

#### ListUsersRequest
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `TenantID` | string | ✅ | 租户 ID |
| `CurrentUserID` | string | ✅ | 当前用户 ID（用于权限过滤） |
| `Search` | string | ❌ | 搜索关键词（模糊匹配 user_account, nickname, email, phone） |
| `Page` | int | ❌ | 页码，默认 1 |
| `Size` | int | ❌ | 每页数量，默认 20 |

**业务逻辑**:
- 获取当前用户角色和 branch_tag（从数据库查询）
- 调用 `GetResourcePermission` 检查 `users` 资源的 `R` 权限
- 根据权限过滤：
  - `AssignedOnly=true`: 只能查看自己（`user_id = currentUserID`）
  - `BranchOnly=true`: 只能查看同 branch 的用户（`branch_tag` 匹配）
  - 否则：可以查看所有用户
- 如果 `Search` 提供，添加搜索条件
- 排序：`ORDER BY user_account ASC`

#### ListUsersResponse
| 字段 | 类型 | 说明 |
|------|------|------|
| `Items` | []*UserDTO | 用户列表 |
| `Total` | int | 总数量 |

---

### 2. GetUser - 查询用户详情

#### GetUserRequest
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `TenantID` | string | ✅ | 租户 ID |
| `UserID` | string | ✅ | 用户 ID |
| `CurrentUserID` | string | ✅ | 当前用户 ID（用于权限检查） |

**业务逻辑**:
- 获取当前用户角色（从数据库查询，不信任 Header）
- 检查是否查看自己（`currentUserID == userID`）
- 如果不是查看自己：
  - 获取目标用户角色
  - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以查看
- 查询用户详情

#### GetUserResponse
| 字段 | 类型 | 说明 |
|------|------|------|
| `User` | *UserDTO | 用户信息 |

---

### 3. CreateUser - 创建用户

#### CreateUserRequest
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `TenantID` | string | ✅ | 租户 ID |
| `CurrentUserID` | string | ✅ | 当前用户 ID（用于权限检查） |
| `UserAccount` | string | ✅ | 用户账号（会转小写并去空格） |
| `Password` | string | ✅ | 密码（明文，Service 层会哈希） |
| `Role` | string | ✅ | 角色 |
| `Nickname` | string | ❌ | 昵称 |
| `Email` | string | ❌ | 邮箱 |
| `Phone` | string | ❌ | 手机号 |
| `Status` | string | ❌ | 状态，默认 "active" |
| `AlarmLevels` | []string | ❌ | 告警级别列表 |
| `AlarmChannels` | []string | ❌ | 告警渠道列表 |
| `AlarmScope` | string | ❌ | 告警范围（根据角色设置默认值） |
| `Tags` | []string | ❌ | 标签列表 |
| `BranchTag` | string | ❌ | 分支标签 |

**业务逻辑**:
1. **参数验证**:
   - `UserAccount`, `Role`, `Password` 必填
2. **权限检查**:
   - 获取当前用户角色（从数据库查询）
   - 系统角色检查：
     - `SystemAdmin`/`SystemOperator` 只能由 SystemAdmin 在 System tenant 中创建
   - 角色层级检查：
     - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以创建
3. **数据准备**:
   - `user_account`: 转小写并去空格
   - `account_hash`: `SHA256(lower(user_account))`
   - `password_hash`: `SHA256(password)`（独立于 account）
   - `email_hash`: `SHA256(lower(email))`（如果 email 提供）
   - `phone_hash`: `SHA256(lower(phone))`（如果 phone 提供）
   - `alarm_scope`: 根据角色设置默认值
     - `Caregiver`/`Nurse`: "ASSIGNED_ONLY"
     - `Manager`: "BRANCH"
     - 其他：NULL
   - `tags`: []string → JSONB
4. **唯一性检查**:
   - `checkEmailUniqueness(tenantID, email, "")`
   - `checkPhoneUniqueness(tenantID, phone, "")`
5. **创建用户**:
   - 调用 `usersRepo.CreateUser`
   - 同步标签到目录（`SyncUserTagsToCatalog`）

#### CreateUserResponse
| 字段 | 类型 | 说明 |
|------|------|------|
| `UserID` | string | 新创建的用户 ID |

---

### 4. UpdateUser - 更新用户

#### UpdateUserRequest
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `TenantID` | string | ✅ | 租户 ID |
| `UserID` | string | ✅ | 用户 ID |
| `CurrentUserID` | string | ✅ | 当前用户 ID（用于权限检查） |
| `Nickname` | *string | ❌ | 昵称（nil=不更新，空字符串=清空） |
| `Email` | *string | ❌ | 邮箱（nil=不更新，null=删除） |
| `EmailHash` | *string | ❌ | 邮箱哈希（前端计算的 hash） |
| `Phone` | *string | ❌ | 手机号（nil=不更新，null=删除） |
| `PhoneHash` | *string | ❌ | 手机号哈希（前端计算的 hash） |
| `Role` | *string | ❌ | 角色 |
| `Status` | *string | ❌ | 状态（必须是 active/disabled/left） |
| `AlarmLevels` | []string | ❌ | 告警级别列表（nil=不更新，空数组=清空） |
| `AlarmChannels` | []string | ❌ | 告警渠道列表（nil=不更新，空数组=清空） |
| `AlarmScope` | *string | ❌ | 告警范围 |
| `Tags` | []string | ❌ | 标签列表（nil=不更新，空数组=清空） |
| `BranchTag` | *string | ❌ | 分支标签（空字符串=NULL） |

**业务逻辑**:
1. **权限检查**:
   - 获取当前用户角色（从数据库查询）
   - 检查是否更新自己（`currentUserID == userID`）
   - 确定更新字段：
     - `updatingRole`: role 不为空
     - `updatingStatus`: status 不为空
     - `updatingOtherFields`: 其他字段
   - 权限规则：
     - 如果更新自己且只更新 password/email/phone：无限制
     - 如果更新其他用户或更新 role/status/otherFields：需要权限检查
   - 角色更新检查：
     - 系统角色：只能由 SystemAdmin 在 System tenant 中分配
     - 其他角色：调用 `canCreateRole` 检查
   - 管理权限检查：
     - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以管理
2. **字段处理**:
   - `email`/`email_hash`: 复杂逻辑
     - 如果 `email_hash` 提供：
       - `email` 为 null：删除 email 但保留 hash
       - `email` 有值：同时更新 email 和 hash
     - 如果只有 `email` 提供（legacy）：
       - `email` 为 null：删除 email 和 hash
       - `email` 有值：计算 hash 并更新
   - `phone`/`phone_hash`: 同 email 逻辑
   - `status`: 验证值（active/disabled/left）
   - `tags`: []string → JSONB（允许清空）
3. **唯一性检查**:
   - 如果更新 email 且不为空：`checkEmailUniqueness(tenantID, email, userID)`
   - 如果更新 phone 且不为空：`checkPhoneUniqueness(tenantID, phone, userID)`
4. **更新用户**:
   - 调用 `usersRepo.UpdateUser`（只更新提供的字段）
   - 如果 tags 更新，同步标签到目录

#### UpdateUserResponse
| 字段 | 类型 | 说明 |
|------|------|------|
| `Success` | bool | 是否成功 |

---

### 5. DeleteUser - 删除用户（软删除）

#### DeleteUserRequest
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `TenantID` | string | ✅ | 租户 ID |
| `UserID` | string | ✅ | 用户 ID |
| `CurrentUserID` | string | ✅ | 当前用户 ID（用于权限检查） |

**业务逻辑**:
1. **权限检查**:
   - 获取当前用户角色（从数据库查询）
   - 获取目标用户角色
   - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以删除
2. **软删除**:
   - 调用 `usersRepo.UpdateUser` 设置 `status = 'left'`

#### DeleteUserResponse
| 字段 | 类型 | 说明 |
|------|------|------|
| `Success` | bool | 是否成功 |

---

### 6. ResetPassword - 重置密码

#### ResetPasswordRequest
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `TenantID` | string | ✅ | 租户 ID |
| `UserID` | string | ✅ | 用户 ID |
| `CurrentUserID` | string | ✅ | 当前用户 ID（用于权限检查） |
| `NewPassword` | string | ✅ | 新密码（明文，Service 层会哈希） |

**业务逻辑**:
1. **权限检查**:
   - 获取当前用户角色（从数据库查询）
   - 检查是否重置自己（`currentUserID == userID`）
   - 如果不是重置自己：
     - 获取目标用户角色
     - 调用 `canCreateRole(currentRole, targetRole)` 检查是否可以重置
2. **密码哈希**:
   - `password_hash`: `SHA256(newPassword)`（独立于 account）
3. **更新密码**:
   - 调用 `usersRepo.UpdateUser` 更新 `password_hash`

#### ResetPasswordResponse
| 字段 | 类型 | 说明 |
|------|------|------|
| `Success` | bool | 是否成功 |
| `Message` | string | 消息（可选，默认 "ok"） |

---

### 7. ResetPIN - 重置 PIN

#### ResetPINRequest
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `TenantID` | string | ✅ | 租户 ID |
| `UserID` | string | ✅ | 用户 ID |
| `CurrentUserID` | string | ✅ | 当前用户 ID（用于权限检查） |
| `NewPIN` | string | ✅ | 新 PIN（必须是 4 位数字） |

**业务逻辑**:
1. **参数验证**:
   - `NewPIN` 必须是 4 位数字
2. **权限检查**: 同 ResetPassword
3. **PIN 哈希**:
   - `pin_hash`: `SHA256(newPIN)`（独立于 account）
4. **更新 PIN**:
   - 调用 `usersRepo.UpdateUser` 更新 `pin_hash`

#### ResetPINResponse
| 字段 | 类型 | 说明 |
|------|------|------|
| `Success` | bool | 是否成功 |

---

## 三、UserDTO 数据传输对象

| 字段 | 类型 | 说明 |
|------|------|------|
| `UserID` | string | 用户 ID |
| `TenantID` | string | 租户 ID |
| `UserAccount` | string | 用户账号 |
| `Nickname` | string | 昵称（可选） |
| `Email` | string | 邮箱（可选） |
| `Phone` | string | 手机号（可选） |
| `Role` | string | 角色 |
| `Status` | string | 状态 |
| `AlarmLevels` | []string | 告警级别列表（可选） |
| `AlarmChannels` | []string | 告警渠道列表（可选） |
| `AlarmScope` | string | 告警范围（可选） |
| `BranchTag` | string | 分支标签（可选） |
| `LastLoginAt` | string | 最后登录时间（RFC3339 格式，可选） |
| `Tags` | []string | 标签列表（可选） |
| `Preferences` | map[string]interface{} | 偏好设置（可选） |

**数据转换规则**:
- `domain.User` → `UserDTO`:
  - `sql.NullString` → `string`（Valid 时）
  - `pq.StringArray` → `[]string`
  - `sql.NullTime` → `string`（RFC3339 格式）
  - `sql.NullString` (JSONB) → `[]string` 或 `map[string]interface{}`

---

## 四、关键业务逻辑函数

### 1. getRoleLevel(role string) int
返回角色的层级（数字越小，权限越高）：
- Level 1: SystemAdmin, SystemOperator
- Level 2: Admin
- Level 3: Manager, IT
- Level 4: Nurse, Caregiver
- Level 5: Resident, Family
- Default: 999（未知角色，最严格）

### 2. canCreateRole(currentRole, targetRole string) bool
检查当前用户是否可以创建指定角色：
- SystemAdmin/SystemOperator 只能由 SystemAdmin 创建（单独检查）
- 其他角色：只能创建同级或下级角色（`targetLevel >= currentLevel`）

### 3. HashAccount(account string) string
哈希账号：`SHA256(lower(account))`

### 4. HashPassword(password string) string
哈希密码：`SHA256(password)`（独立于 account）

### 5. domainUserToDTO(user *domain.User) *UserDTO
将 `domain.User` 转换为 `UserDTO`

---

## 五、依赖的外部函数

### 1. 权限检查（通过 Repository 层）
- `GetResourcePermission(ctx, roleCode, resourceType, permissionType)`: 查询资源权限配置
- 返回 `*PermissionCheck{AssignedOnly, BranchOnly}`
- **实现位置**: `repository.UsersRepository` 接口（需要添加）

### 2. 唯一性检查（通过 Repository 层）
- `CheckEmailUniqueness(ctx, tenantID, email, excludeUserID)`: 检查 email 唯一性
- `CheckPhoneUniqueness(ctx, tenantID, phone, excludeUserID)`: 检查 phone 唯一性
- **实现位置**: `repository.UsersRepository` 接口（需要添加）

### 3. 系统常量
- `SystemTenantID()`: 获取系统租户 ID
- **实现位置**: Service 层或使用 `service.SystemTenantID()` 常量

---

## 六、与旧 Handler 的对比

### 相同点
- ✅ 所有方法的功能和业务逻辑保持一致
- ✅ 权限检查规则完全一致
- ✅ 数据转换规则完全一致
- ✅ 响应格式完全一致

### 改进点
- ✅ 使用强类型 DTO，替代 `map[string]any`
- ✅ 清晰的接口定义，便于测试和维护
- ✅ 业务逻辑集中在 Service 层，Handler 层只负责 HTTP 处理
- ✅ 更好的错误处理和日志记录

---

## 七、设计决策（已确认）

1. **权限检查函数调用方式**: ✅ **方案 B - 通过 Repository 层封装权限检查**
   - 在 Repository 层添加 `GetResourcePermission` 方法
   - Service 层调用 Repository 方法，符合分层架构

2. **唯一性检查函数调用方式**: ✅ **方案 B - 在 Repository 层实现唯一性检查**
   - 在 Repository 层添加 `CheckEmailUniqueness` 和 `CheckPhoneUniqueness` 方法
   - Service 层调用 Repository 方法，符合分层架构

3. **AuthStore 同步**: ✅ **已采用新的 AuthService**
   - 检查 AuthService 实现，确认不再使用 AuthStore
   - Service 层不处理 AuthStore，由 Handler 层处理（如果需要）

4. **错误处理**: ✅ **Service 层返回业务错误，Handler 层转换为 HTTP 响应**
   - Service 层返回 `error`，包含业务错误信息
   - Handler 层将错误转换为 HTTP 响应格式

---

## 八、接口设计确认清单

- [ ] 所有 DTO 字段定义是否合理？
- [ ] 业务逻辑描述是否准确？
- [ ] 权限检查规则是否完整？
- [ ] 数据转换规则是否正确？
- [ ] 依赖的外部函数调用方式是否合适？
- [ ] 是否有遗漏的功能或字段？

---

## 九、下一步

确认接口设计后，将进行：
1. **阶段 3**: 实现所有 Service 方法
2. **阶段 4**: 编写 Service 测试
3. **阶段 5**: 实现 Handler 层
4. **阶段 6**: 集成和路由注册
5. **阶段 7**: 验证和测试

