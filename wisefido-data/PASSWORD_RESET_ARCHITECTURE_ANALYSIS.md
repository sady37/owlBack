# 密码重置功能架构分析

## 当前实现情况

### 1. AuthService.ResetPassword（忘记密码流程）
**用途**：用户主动发起的"忘记密码"流程
- **触发场景**：用户忘记密码，通过验证码验证后重置
- **流程**：发送验证码 → 验证验证码 → 获取令牌 → 重置密码
- **权限**：无需权限检查（用户自己操作）
- **支持类型**：staff 和 resident（包括 resident_contacts）
- **特点**：需要验证码和令牌验证

### 2. UserService.ResetPassword（管理员重置用户密码）
**用途**：管理员重置其他用户的密码
- **触发场景**：管理员在用户管理界面重置用户密码
- **流程**：管理员操作 → 权限检查 → 重置密码
- **权限**：需要角色层级检查（不能重置同级或更高级角色的密码）
- **支持类型**：staff（users 表）
- **特点**：管理员操作，需要权限验证

### 3. ResidentService.ResetResidentPassword（重置住户密码）
**用途**：重置住户密码（管理员或住户自己）
- **触发场景**：
  - 管理员在住户管理界面重置住户密码
  - 住户自己重置密码
- **流程**：权限检查 → 重置密码
- **权限**：
  - Resident/Family：只能重置自己的密码
  - Staff：需要 AssignedOnly/BranchOnly 权限检查
- **支持类型**：resident（residents 表）
- **特点**：支持住户自己操作和管理员操作

### 4. ResidentService.ResetContactPassword（重置联系人密码）
**用途**：重置住户联系人密码（管理员或住户自己）
- **触发场景**：
  - 管理员在住户管理界面重置联系人密码
  - 住户自己重置联系人密码
- **流程**：权限检查 → 重置密码
- **权限**：
  - Resident/Family：只能重置自己的联系人密码
  - Staff：需要 AssignedOnly/BranchOnly 权限检查
- **支持类型**：resident_contacts 表
- **特点**：支持住户自己操作和管理员操作

---

## 架构方案对比

### 方案 A：统一到 AuthService（集中式）

**优点**：
- ✅ 密码相关逻辑集中管理，便于维护
- ✅ 统一的密码 hash 逻辑
- ✅ 减少代码重复
- ✅ 符合"单一职责"原则（AuthService 负责所有认证相关功能）

**缺点**：
- ❌ AuthService 需要了解 User/Resident 的业务逻辑（权限检查、角色层级）
- ❌ 耦合度高：AuthService 需要依赖 UserService/ResidentService 的权限检查逻辑
- ❌ 违反"关注点分离"：AuthService 应该只负责认证，不应该处理业务权限

**实现方式**：
```go
// AuthService 需要调用 UserService/ResidentService 进行权限检查
func (s *authService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
    // 1. 权限检查（需要调用 UserService/ResidentService）
    // 2. 重置密码
}
```

---

### 方案 B：保持在各 Service 内部（分布式）✅ **推荐**

**优点**：
- ✅ **关注点分离**：每个 Service 负责自己的业务逻辑
  - AuthService：负责认证流程（登录、忘记密码）
  - UserService：负责用户管理（包括密码重置）
  - ResidentService：负责住户管理（包括密码重置）
- ✅ **权限检查内聚**：权限检查逻辑和业务逻辑在一起，更容易理解和维护
- ✅ **低耦合**：各 Service 独立，不相互依赖
- ✅ **符合领域驱动设计**：每个 Service 代表一个业务领域

**缺点**：
- ❌ 密码 hash 逻辑可能重复（但可以通过工具函数解决）
- ❌ 需要维护多个密码重置方法

**实现方式**：
```go
// UserService 内部处理
func (s *userService) ResetPassword(ctx context.Context, req UserResetPasswordRequest) error {
    // 1. 权限检查（UserService 自己的业务逻辑）
    // 2. 重置密码
}

// ResidentService 内部处理
func (s *residentService) ResetResidentPassword(ctx context.Context, req ResetResidentPasswordRequest) error {
    // 1. 权限检查（ResidentService 自己的业务逻辑）
    // 2. 重置密码
}
```

---

## 推荐方案：方案 B（保持在各 Service 内部）

### 理由

1. **职责清晰**
   - **AuthService**：负责认证流程（登录、忘记密码、验证码）
   - **UserService**：负责用户管理（CRUD + 密码重置）
   - **ResidentService**：负责住户管理（CRUD + 密码重置）

2. **权限检查内聚**
   - UserService 的密码重置需要角色层级检查（业务逻辑）
   - ResidentService 的密码重置需要 AssignedOnly/BranchOnly 检查（业务逻辑）
   - 这些权限检查逻辑属于各自的业务领域，不应该放在 AuthService

3. **低耦合**
   - AuthService 不需要了解 User/Resident 的业务规则
   - 各 Service 可以独立演进

4. **符合当前实现**
   - 当前代码已经按照这种方式实现
   - 只需要统一密码 hash 工具函数即可

---

## 优化建议

### 1. 统一密码 Hash 工具函数

**当前问题**：
- `user_service.go` 中有 `HashPassword` 和 `sha256Hex`
- `resident_service.go` 中也有 `HashPassword`
- `auth_service.go` 中有 `sha256HexAuth`

**解决方案**：
创建统一的工具包 `internal/util/password.go`：

```go
package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashPassword 计算密码的 SHA256 hash（hex 编码）
// 密码 hash 独立于账号，仅 hash 密码本身
func HashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

// HashAccount 计算账号的 SHA256 hash（hex 编码）
// 账号 hash 使用小写和规范化
func HashAccount(account string) string {
	normalized := strings.ToLower(strings.TrimSpace(account))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
```

### 2. 保持当前架构

**AuthService.ResetPassword**：
- 仅用于"忘记密码"流程（通过验证码验证）
- 不需要权限检查（用户自己操作）

**UserService.ResetPassword**：
- 用于管理员重置用户密码
- 需要角色层级权限检查

**ResidentService.ResetResidentPassword / ResetContactPassword**：
- 用于重置住户/联系人密码
- 需要 AssignedOnly/BranchOnly 权限检查

---

## 总结

**推荐方案**：**保持在各 Service 内部（方案 B）**

**理由**：
1. ✅ 职责清晰，符合单一职责原则
2. ✅ 权限检查内聚，业务逻辑清晰
3. ✅ 低耦合，易于维护和扩展
4. ✅ 符合领域驱动设计

**需要做的优化**：
1. ✅ 统一密码 hash 工具函数（创建 `internal/util/password.go`）
2. ✅ 保持当前架构不变

