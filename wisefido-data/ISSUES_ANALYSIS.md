# 问题分析报告

## 问题1: test1 密码第一列显示密码

### 现象
jack(role=admin) 查看 test1 时，密码第一列显示密码

### 分析
1. **表格列定义检查**：
   - `ResidentList.vue` 中的 `columns` 定义没有密码列
   - 只有 "Passwd" 按钮用于重置密码
   - 表格列包括：Nickname, Branch, Building, area_tag, unit_name, Status, Service Level, Admission Date, account, Family_tag, Allow Access, Operation

2. **可能原因**：
   - API 返回的数据中可能包含了密码字段
   - 某个列可能误显示了密码数据
   - 需要检查 API 返回的数据结构

### 建议
- 检查 `/admin/api/v1/residents` API 返回的数据结构
- 确认是否有字段意外返回了密码信息
- 检查前端表格的数据绑定

---

## 问题2: smith 的 PHI email/phone 全是空的

### 现象
查看 smith 的 PHI，email/phone 全是空的

### 分析
1. **数据库检查**：
   - smith (r1) 在 `residents` 表中有记录
   - 需要检查 `resident_phi` 表中是否有对应的记录

2. **代码逻辑**：
   ```go
   // admin_residents_handlers.go:477-500
   savePhone, _ := payload["save_phone"].(bool)
   saveEmail, _ := payload["save_email"].(bool)
   residentPhone, _ := payload["resident_phone"]
   residentEmail, _ := payload["resident_email"]
   if savePhone {
       // 只有 savePhone=true 时才保存 resident_phone
   }
   if saveEmail {
       // 只有 saveEmail=true 时才保存 resident_email
   }
   ```
   - **问题**：如果创建 resident 时没有勾选 "Save" 复选框，`resident_phone` 和 `resident_email` 不会保存到 `resident_phi` 表
   - 即使提供了 `phone_hash` 和 `email_hash`（保存到 `residents` 表用于登录），明文也不会保存

3. **更新逻辑**：
   - 更新 PHI 时，如果 `save_phone=false` 或 `save_email=false`，会设置为 NULL
   - 这可能导致之前保存的数据被清除

### 建议
- 检查 `resident_phi` 表中 smith 的记录
- 确认创建时是否勾选了 "Save" 复选框
- 考虑是否需要修改逻辑：如果提供了 email/phone 且用于登录（有 hash），是否应该自动保存明文

---

## 问题3: email_hash/phone_hash 唯一性检查

### 现象
应用层在插入DB时，有没有保证 resident/user/resident_contacts 3张表，同一租户下，email_hash、phone_hash 不重复

### 分析

#### 1. residents 表
**当前状态：❌ 没有唯一性检查**

- **创建 resident** (`admin_residents_handlers.go:429-436`)：
  ```go
  INSERT INTO residents (..., phone_hash, email_hash)
  ```
  - 没有调用 `checkHashUniqueness` 检查
  - 可能导致同一租户下多个 resident 使用相同的 email/phone

- **更新 resident PHI** (`admin_residents_handlers.go:1173-1196`)：
  ```go
  UPDATE residents SET phone_hash = ..., email_hash = ...
  ```
  - 没有检查唯一性
  - 可能导致更新后与其他 resident 冲突

#### 2. resident_contacts 表
**当前状态：✅ 有唯一性检查（但不够严格）**

- **创建 contact** (`admin_residents_handlers.go:552`)：
  ```go
  if err := checkHashUniqueness(s.DB, r, tenantID, "resident_contacts", phoneHashBytes, emailHashBytes, "", ""); err != nil {
      // 只记录错误，不阻止创建
      continue
  }
  ```
  - ✅ 有检查
  - ❌ 但错误时只记录日志，不阻止创建（`continue`）

- **更新 contact** (`admin_residents_handlers.go:1362`)：
  ```go
  if err := checkHashUniqueness(s.DB, r, tenantID, "resident_contacts", phoneHashBytes, emailHashBytes, existingContactID, "contact_id"); err != nil {
      writeJSON(w, http.StatusOK, Fail(err.Error()))
      return
  }
  ```
  - ✅ 有检查
  - ✅ 错误时正确返回

#### 3. users 表
**当前状态：⚠️ 检查明文，不是 hash**

- **创建 user** (`admin_users_handlers.go:242-249`)：
  ```go
  if err := checkEmailUniqueness(s.DB, r, tenantID, email, ""); err != nil {
      // 检查明文 email
  }
  if err := checkPhoneUniqueness(s.DB, r, tenantID, phone, ""); err != nil {
      // 检查明文 phone
  }
  ```
  - ✅ 有唯一性检查
  - ⚠️ 但检查的是明文，不是 hash
  - ⚠️ `users` 表可能没有 `email_hash` 和 `phone_hash` 字段（需要确认）

### 问题总结

1. **residents 表**：
   - ❌ 创建时没有唯一性检查
   - ❌ 更新时没有唯一性检查
   - **风险**：同一租户下可能创建多个使用相同 email/phone 的 resident

2. **resident_contacts 表**：
   - ✅ 创建时有检查，但错误时不阻止
   - ✅ 更新时有检查且正确返回错误
   - **风险**：创建时检查失败仍会继续创建

3. **users 表**：
   - ✅ 有唯一性检查（明文）
   - ⚠️ 需要确认是否使用 hash

### 建议修复

1. **residents 表**：
   - 在创建 resident 前检查 `phone_hash` 和 `email_hash` 的唯一性
   - 在更新 resident PHI 时检查唯一性（排除当前 resident）

2. **resident_contacts 表**：
   - 修复创建时的逻辑：检查失败应该返回错误，而不是继续创建

3. **统一检查逻辑**：
   - 确保三张表都使用相同的唯一性检查逻辑
   - 考虑在数据库层面添加唯一约束（如果业务允许）

---

## 修复优先级

1. **高优先级**：residents 表的唯一性检查（可能导致登录冲突）
2. **中优先级**：resident_contacts 创建时的错误处理
3. **低优先级**：确认 users 表是否使用 hash 存储

