# 字段更新统一设计规则（Field Update Design）

## 一、字段分类

### 1. 普通字段（String/Number/Boolean）
- **示例**：`nickname`, `role`, `status`, `alarm_scope`
- **特点**：单一值，无 hash，无占位符

### 2. 数组字段（Array）
- **示例**：`alarm_levels`, `alarm_channels`, `tags`, `branch_ids`
- **特点**：多个值，无 hash，无占位符

### 3. 有 Hash 的字段（String + Hash）
- **示例**：`email/email_hash`, `phone/phone_hash`, `account/account_hash`
- **特点**：
  - 有对应的 hash 字段，可能有占位符
  - **绝对不返回 email_hash 和 phone_hash 给前端（安全考虑）**

### 4. 密码字段（Password）
- **示例**：`password` → `password_hash`
- **特点**：
  - 不存储明文，只存储 hash
  - **绝对不返回 password 或 password_hash 给前端（安全考虑）**
  - 通过主表单提交（不在表单中显示，但可以通过主表单提交）

---

## 二、Service 返回给前端的规则

### 规则 1：普通字段
```
字段不存在/空/null → 不返回字段（omitempty）
字段有值 → 返回值
```

**示例**：
- `nickname = ""` → 不返回 `nickname` 字段
- `nickname = "John"` → 返回 `"nickname": "John"`

### 规则 2：数组字段
```
数组为空/null → 不返回字段（omitempty）
数组有值 → 返回数组
```

**示例**：
- `tags = []` → 不返回 `tags` 字段
- `tags = ["A", "B"]` → 返回 `"tags": ["A", "B"]`

### 规则 3：有 Hash 的字段（Email/Phone/Account）
```
情况 1：字段有值且不为 "" → 直接返回值
情况 2：字段无值或 ""
  - 当对应的 hash 有值且不为空 → 返回占位符
    - email → "xxx@xxx.xxx"
    - phone → "xxx-xxx-xxxx"
    - account → "xxx@xxx.xxx"（如果适用）
  - 当对应的 hash null 或 "" → 返回 null

EmailHash/PhoneHash：
  - **绝对不返回给前端（安全考虑）**
  - Service 层：不返回 email_hash 和 phone_hash
  - Handler 层：不返回 email_hash 和 phone_hash
  - 前端：不期望接收 email_hash 和 phone_hash
  - 前端：通过主表单提交时，前端计算并发送 email_hash 和 phone_hash
```

**示例**：
- `email = "user@example.com"` → 返回 `"email": "user@example.com"`
- `email = ""` 且 `email_hash != null` → 返回 `"email": "xxx@xxx.xxx"`
- `email = ""` 且 `email_hash = null` → 返回 `"email": null`
- **绝对不返回 `email_hash` 和 `phone_hash` 字段**

### 规则 4：密码字段（Password）
```
**绝对不返回 password 或 password_hash 字段（安全考虑）**
- Service 层：不返回任何密码相关信息
- 前端：不显示 password 输入框（安全考虑）
- 前端：不期望接收 password_hash（安全考虑）
- 前端：通过独立的输入或 Modal 让用户输入新密码，前端计算 password_hash
- 前端：通过主表单提交 password_hash（如果用户需要修改密码）
```

---

## 三、前端返回给 Service 的规则

### 规则 1：普通字段
```
情况 1：值改变且不为空 → 返回新值
情况 2：值 = "" 或 null → 返回 null（清空）
情况 3：值未变 → 返回 nil（不包含字段，不更新）
```

**实现逻辑**：
```typescript
if (currentValue !== originalValue) {
  if (currentValue.trim() === '') {
    params.fieldName = null  // 清空
  } else {
    params.fieldName = currentValue  // 更新
  }
}
// 值未变，不包含在 params 中
```

### 规则 2：数组字段
```
情况 1：值改变且不为空 → 返回新数组
情况 2：值 = [] 或 null → 返回 null（清空）
情况 3：值未变 → 返回 nil（不包含字段，不更新）
```

**实现逻辑**：
```typescript
if (!arraysEqual(currentValue, originalValue)) {
  if (currentValue.length === 0) {
    params.fieldName = null  // 清空
  } else {
    params.fieldName = currentValue  // 更新
  }
}
// 值未变，不包含在 params 中
```

### 规则 3：有 Hash 的字段（Email/Phone/Account）
```
情况 1：值改变且不为空 → 返回新值 + 计算 hash
情况 2：值 = "" 或 null → 返回 null + hash = null（清空）
情况 3：值未变 → 返回 nil（不包含字段，不更新）
情况 4：占位符未变 → 返回 nil（不包含字段，不更新）
```

**实现逻辑**：
```typescript
const placeholder = 'xxx@xxx.xxx'  // 或 'xxx-xxx-xxxx'
const isPlaceholder = currentValue === placeholder
const wasPlaceholder = originalValue === placeholder

if (isPlaceholder && wasPlaceholder) {
  // 仍是占位符，不更新
  return
}

if (currentValue !== originalValue) {
  if (currentValue.trim() === '') {
    params.email = null
    params.email_hash = null  // 清空
  } else {
    params.email = currentValue
    params.email_hash = await hashAccount(currentValue)  // 更新
  }
}
// 值未变，不包含在 params 中
```

### 规则 4：密码字段（Password）
```
**绝对不返回 password 或 password_hash 字段（安全考虑）**
- Service 层：不返回任何密码相关信息（不返回 password_hash）
- 前端：不期望接收 password_hash（安全考虑）
- 前端：通过独立的输入或 Modal 让用户输入新密码，前端计算 password_hash
- 前端：通过主表单提交 password_hash（如果用户需要修改密码）
- 处理逻辑：
  - 用户输入新密码（通过独立的输入或 Modal）
  - 前端计算 password_hash
  - 在主表单提交时，如果 password_hash 有值，发送 password_hash（不发送 password）
  - 规则与其他字段相同：
    - 值改变且不为空 → 返回 password_hash
    - 值未变 → 返回 nil（不更新）
```

---

## 四、Handler 层解析规则

### 规则 1：普通字段
```
字段不存在 → req.Field = nil（不更新）
字段为 null → req.Field = &""（清空）
字段有值 → req.Field = &value（更新）
```

**实现逻辑**：
```go
if val, ok := payload["fieldName"]; ok {
    if val == nil {
        empty := ""
        req.Field = &empty  // null → 清空
    } else if s, ok := val.(string); ok {
        req.Field = &s  // 有值 → 更新
    }
}
// 字段不存在，req.Field = nil（不更新）
```

### 规则 2：数组字段
```
字段不存在 → req.Field = nil（不更新）
字段为 null → req.Field = []string{}（清空）
字段有值 → req.Field = []string{...}（更新）
```

**实现逻辑**：
```go
if val, ok := payload["fieldName"]; ok {
    if val == nil {
        req.Field = []string{}  // null → 清空
    } else if arr, ok := val.([]any); ok {
        result := make([]string, 0, len(arr))
        for _, v := range arr {
            if s, ok := v.(string); ok && s != "" {
                result = append(result, s)
            }
        }
        req.Field = result  // 有值 → 更新
    }
}
// 字段不存在，req.Field = nil（不更新）
```

### 规则 3：有 Hash 的字段
```
Email/Phone：
  - 字段不存在 → req.Email = nil（不更新）
  - 字段为 null → req.Email = &""（清空）
  - 字段有值 → req.Email = &value（更新）

EmailHash/PhoneHash：
  - **绝对不返回给前端（安全考虑）**
  - Service 层：不返回 email_hash 和 phone_hash
  - Handler 层：只接收 email_hash 和 phone_hash（不返回）
    - 字段不存在 → req.EmailHash = nil（不更新）
    - 字段为 null → req.EmailHash = &""（清空 hash）
    - 字段有值 → req.EmailHash = &hash（更新 hash）
  - 前端：通过主表单提交时，前端计算并发送 email_hash 和 phone_hash
```

### 规则 4：密码字段
```
**绝对不返回 password 或 password_hash 字段（安全考虑）**
- Service 层：不返回任何密码相关信息（不返回 password_hash）
- Handler 层：只接收 password_hash 字段（不接收 password 字段）
  - 字段不存在 → req.PasswordHash = nil（不更新）
  - 字段为 null → req.PasswordHash = nil（不更新，密码字段不允许清空）
  - 字段有值 → req.PasswordHash = &hash（更新）
- 通过主表单提交，与其他字段一起处理
```

---

## 五、Service 层更新规则

### 规则 1：普通字段（带值比较）
```
req.Field == nil → 不更新
req.Field != nil && *req.Field == currentValue → 不更新（值未变）
req.Field != nil && *req.Field != currentValue → 更新
  - *req.Field == "" → 设置为 NULL
  - *req.Field != "" → 设置新值
```

**实现逻辑**：
```go
if req.Field != nil {
    newVal := strings.TrimSpace(*req.Field)
    if newVal != currentValue {
        if newVal == "" {
            updateUser.Field = sql.NullString{Valid: false}  // 清空
        } else {
            updateUser.Field = sql.NullString{String: newVal, Valid: true}  // 更新
        }
    }
}
// req.Field == nil，不更新
```

### 规则 2：数组字段（带值比较）
```
req.Field == nil → 不更新
req.Field != nil && arraysEqual(req.Field, currentValue) → 不更新（值未变）
req.Field != nil && !arraysEqual(req.Field, currentValue) → 更新
  - len(req.Field) == 0 → 设置为 NULL（空数组）
  - len(req.Field) > 0 → 设置新数组
```

**实现逻辑**：
```go
if req.Field != nil {
    if !stringSlicesEqual(currentValue, req.Field) {
        if len(req.Field) == 0 {
            updateUser.Field = []string{}  // Repository 层会转换为 NULL
        } else {
            updateUser.Field = req.Field  // 更新
        }
    }
}
// req.Field == nil，不更新
```

### 规则 3：有 Hash 的字段
```
Email/Phone：
  - req.Email == nil → 不更新 email
  - req.Email != nil && *req.Email == "" → 清空 email 和 hash
  - req.Email != nil && *req.Email != "" → 更新 email，计算或使用提供的 hash

EmailHash/PhoneHash：
  - req.EmailHash == nil → 不更新 hash
  - req.EmailHash != nil && *req.EmailHash == "" → 清空 hash（需要配合 Email 使用）
  - req.EmailHash != nil && *req.EmailHash != "" → 更新 hash
```

**实现逻辑**：
```go
if req.Email != nil {
    if *req.Email == "" {
        // 清空 email 和 hash
        updateUser.Email = sql.NullString{Valid: false}
        updateUser.EmailHash = nil
    } else {
        // 更新 email
        updateUser.Email = sql.NullString{String: *req.Email, Valid: true}
        // 计算或使用提供的 hash
        if req.EmailHash != nil && *req.EmailHash != "" {
            emailHashBytes, _ := hex.DecodeString(*req.EmailHash)
            updateUser.EmailHash = emailHashBytes
        } else {
            emailHash, _ := hex.DecodeString(HashAccount(*req.Email))
            updateUser.EmailHash = emailHash
        }
    }
} else if req.EmailHash != nil {
    // 只更新 hash，不更新 email
    if *req.EmailHash != "" {
        emailHashBytes, _ := hex.DecodeString(*req.EmailHash)
        updateUser.EmailHash = emailHashBytes
    }
}
```

### 规则 4：密码字段
```
req.PasswordHash == nil → 不更新
req.PasswordHash != nil && *req.PasswordHash != "" → 更新 password_hash
```

---

## 六、Repository 层更新规则

### 规则 1：普通字段
```
updateUser.Field.Valid == false → 设置为 NULL
updateUser.Field.Valid == true → 设置新值
```

### 规则 2：数组字段
```
len(updateUser.Field) == 0 → 设置为 NULL
len(updateUser.Field) > 0 → 设置新数组
```

### 规则 3：有 Hash 的字段
```
Email/Phone：
  - updateUser.Email.Valid == false → 设置为 NULL
  - updateUser.Email.Valid == true → 设置新值

EmailHash/PhoneHash：
  - updateUser.EmailHash == nil → 不更新
  - len(updateUser.EmailHash) == 0 → 设置为 NULL
  - len(updateUser.EmailHash) > 0 → 设置新 hash
```

---

## 七、完整数据流示例

### 示例 1：普通字段（nickname）

**场景**：用户修改 nickname 从 "John" 到 "Jane"

1. **前端**：
   - `currentValue = "Jane"`, `originalValue = "John"`
   - `currentValue !== originalValue` → 发送 `"nickname": "Jane"`

2. **Handler**：
   - `payload["nickname"] = "Jane"` → `req.Nickname = &"Jane"`

3. **Service**：
   - `req.Nickname != nil` → 检查值是否改变
   - `*req.Nickname = "Jane"`, `currentValue = "John"` → 值改变
   - `updateUser.Nickname = sql.NullString{String: "Jane", Valid: true}`

4. **Repository**：
   - `UPDATE users SET nickname = 'Jane' WHERE ...`

### 示例 2：数组字段（tags）

**场景**：用户修改 tags 从 ["A", "B"] 到 ["A", "C"]

1. **前端**：
   - `currentValue = ["A", "C"]`, `originalValue = ["A", "B"]`
   - `!arraysEqual(currentValue, originalValue)` → 发送 `"tags": ["A", "C"]`

2. **Handler**：
   - `payload["tags"] = ["A", "C"]` → `req.Tags = ["A", "C"]`

3. **Service**：
   - `req.Tags != nil` → 检查值是否改变
   - `!stringSlicesEqual(currentValue, req.Tags)` → 值改变
   - `updateUser.Tags = sql.NullString{String: '["A","C"]', Valid: true}`

4. **Repository**：
   - `UPDATE users SET user_tags = '["A","C"]'::jsonb WHERE ...`

### 示例 3：有 Hash 的字段（email）

**场景 A**：用户修改 email 从 "old@example.com" 到 "new@example.com"

1. **Service 返回给前端**：
   - **绝对不返回 email_hash 和 phone_hash（安全考虑）**
   - 只返回 `email` 字段（实际值或占位符）

2. **前端**：
   - `currentValue = "new@example.com"`, `originalValue = "old@example.com"`
   - `currentValue !== originalValue` → 发送 `"email": "new@example.com"` + 计算 `email_hash`

3. **Handler**：
   - `payload["email"] = "new@example.com"` → `req.Email = &"new@example.com"`
   - `payload["email_hash"] = "hash..."` → `req.EmailHash = &"hash..."`

4. **Service**：
   - `req.Email != nil` → 检查值是否改变
   - `*req.Email = "new@example.com"`, `currentValue = "old@example.com"` → 值改变
   - `updateUser.Email = sql.NullString{String: "new@example.com", Valid: true}`
   - `updateUser.EmailHash = emailHashBytes`

5. **Repository**：
   - `UPDATE users SET email = 'new@example.com', email_hash = '...' WHERE ...`

**场景 B**：用户清空 email（从有值到空）

1. **前端**：
   - `currentValue = ""`, `originalValue = "old@example.com"`
   - `currentValue !== originalValue` → 发送 `"email": null`, `"email_hash": null`

2. **Handler**：
   - `payload["email"] = nil` → `req.Email = &""`
   - `payload["email_hash"] = nil` → `req.EmailHash = &""`

3. **Service**：
   - `req.Email != nil && *req.Email == ""` → 清空
   - `updateUser.Email = sql.NullString{Valid: false}`
   - `updateUser.EmailHash = nil`

4. **Repository**：
   - `UPDATE users SET email = NULL, email_hash = NULL WHERE ...`

**场景 C**：用户未修改 email（仍是占位符）

1. **前端**：
   - `currentValue = "xxx@xxx.xxx"`, `originalValue = "xxx@xxx.xxx"`
   - `isPlaceholder && wasPlaceholder` → 不发送字段

2. **Handler**：
   - `payload["email"]` 不存在 → `req.Email = nil`

3. **Service**：
   - `req.Email == nil` → 不更新

4. **Repository**：
   - 不执行 UPDATE

### 示例 4：密码字段（password）

**场景**：用户修改密码（通过主表单提交）

1. **Service 返回给前端**：
   - **绝对不返回 password 或 password_hash 字段（安全考虑）**
   - 前端不期望接收任何密码相关信息

2. **前端**：
   - 用户输入新密码（通过独立的输入或 Modal）
   - 计算 `password_hash = hashPassword(newPassword)`
   - 在主表单提交时，如果 password_hash 有值且改变，发送 `"password_hash": "hash..."`（不发送 `password`）
   - 如果 password_hash 未改变，不发送字段（不更新）

3. **Handler**：
   - `payload["password_hash"] = "hash..."` → `req.PasswordHash = &"hash..."`
   - 如果字段不存在，`req.PasswordHash = nil`（不更新）

4. **Service**：
   - `req.PasswordHash != nil` → 检查值是否改变
   - 如果值改变，`updateUser.PasswordHash = passwordHashBytes`
   - 如果值未变，不更新

5. **Repository**：
   - `UPDATE users SET password_hash = '...' WHERE ...`

---

## 八、实现检查清单

### 前端（Vue）
- [ ] 保存所有字段的原始值
- [ ] 实现值比较函数（字符串、数组）
- [ ] 实现统一的字段更新函数
- [ ] 处理占位符逻辑（有 hash 的字段）
- [ ] 处理密码字段（独立 Modal）

### Handler 层
- [ ] 实现统一的字段解析函数
- [ ] 处理 null 值（清空）
- [ ] 处理字段不存在（不更新）
- [ ] 处理有 hash 的字段（email/phone）

### Service 层
- [ ] 实现值比较函数（字符串、数组）
- [ ] 实现统一的字段更新函数
- [ ] 处理有 hash 的字段（计算或使用提供的 hash）
- [ ] 处理密码字段（只更新 hash）

### Repository 层
- [ ] 处理 NULL 值（sql.NullString{Valid: false}）
- [ ] 处理空数组（转换为 NULL）
- [ ] 处理 hash 字段（[]byte）

---

## 九、注意事项

1. **值比较**：必须忽略顺序（数组字段）
2. **占位符**：只在有 hash 的字段中使用
3. **密码字段**：永远不发送明文，只发送 hash
4. **空值处理**：空字符串和 null 都表示清空
5. **字段不存在**：表示不更新，保持原值

