# Resident Account 和 PHI 处理修复总结

## 1. resident_account 生成规则 ✅

### 修复前
- 如果未提供 `resident_account`，自动从 `nickname` 生成（格式: `lowercase(nickname.replace(' ', '-'))`）

### 修复后
- **`resident_account` 必须由前端提供**（必填字段）
- 原因：各家机构已有自己的编码模式，不应统一生成规则
- 验证：如果未提供，返回错误 "resident_account is required (each institution has its own encoding pattern)"

### 代码变更
- `admin_residents_handlers.go` - POST /admin/api/v1/residents
- 移除了自动生成逻辑
- 添加了必填验证

---

## 2. email/phone 保存逻辑 ✅

### 修复前
- email/phone 参数被忽略（因为 residents 表不存储这些字段）

### 修复后
- **phone_hash 和 email_hash**：始终保存到 `residents` 表（用于 login）
  - 如果提供了 `phone`，计算 `phone_hash` 并保存
  - 如果提供了 `email`，计算 `email_hash` 并保存
  - 这些 hash 用于 login，不存储明文（HIPAA 合规）

- **resident_phone 和 resident_email**（明文）：只在用户明确选择保存时才保存到 `resident_phi` 表
  - 需要 `save_phone=true` 且提供了 `resident_phone`
  - 需要 `save_email=true` 且提供了 `resident_email`
  - 默认不保存明文（HIPAA 合规）

### 代码变更
- `admin_residents_handlers.go` - POST /admin/api/v1/residents
- 添加了 phone_hash 和 email_hash 的处理
- 添加了 save_phone/save_email 标志的检查
- 只在用户明确选择保存时才保存明文到 resident_phi 表

---

## 3. Login 逻辑验证 ✅

### 当前实现
- 后端同时比对 3 组 hash：
  ```sql
  WHERE (r.resident_account_hash = $1 OR r.phone_hash = $1 OR r.email_hash = $1)
    AND r.password_hash = $2
  ```

### 说明
- 前端发送一个 `accountHash`（可能是 account、email 或 phone 的 hash）
- 后端使用 OR 条件同时比对 3 种可能性：
  - `resident_account_hash = accountHash`
  - `phone_hash = accountHash`
  - `email_hash = accountHash`
- 如果 account 是 email 格式，它可能同时匹配 `resident_account_hash` 和 `email_hash`
- 如果 account 是 number 格式，它可能同时匹配 `resident_account_hash` 和 `phone_hash`
- 这确保了无论用户输入的是 account、email 还是 phone，都能正确登录

### 文件
- `auth_handlers.go` - /auth/api/v1/login
- 当前实现已满足需求，无需修改

---

## 4. 前端需要调整

### CreateResidentParams
- `resident_account`: 从可选改为必填
- `save_phone`: 新增标志，表示是否保存明文 phone 到 resident_phi
- `save_email`: 新增标志，表示是否保存明文 email 到 resident_phi
- `resident_phone`: 如果 `save_phone=true`，保存到 resident_phi
- `resident_email`: 如果 `save_email=true`，保存到 resident_phi

### 建议
1. 前端创建 resident 表单中，`resident_account` 应该标记为必填
2. 如果提供了 phone/email，前端应该：
   - 自动计算 hash 并发送（用于 login）
   - 提供选项让用户选择是否保存明文（save_phone/save_email）
   - 如果用户选择保存，才发送 `resident_phone`/`resident_email` 和对应的标志

---

## 测试建议

1. **resident_account 必填验证**：
   - 不提供 `resident_account`，应该返回错误

2. **phone/email hash 保存**：
   - 提供 `phone`，验证 `residents.phone_hash` 已保存
   - 提供 `email`，验证 `residents.email_hash` 已保存

3. **明文 PHI 保存**：
   - 提供 `phone` 但不设置 `save_phone=true`，验证 `resident_phi.resident_phone` 为空
   - 提供 `phone` 且设置 `save_phone=true`，验证 `resident_phi.resident_phone` 已保存
   - 同样测试 `email` 和 `save_email`

4. **Login 测试**：
   - 使用 account 登录
   - 使用 email 登录（如果 account 是 email 格式）
   - 使用 phone 登录（如果 account 是 number 格式）

