# 代码检查报告 - 字段更新统一设计规则

## 检查时间
2025-01-XX

## 检查范围
按照 `FIELD_UPDATE_DESIGN.md` 中的规则，检查各层代码实现。

---

## 一、Service 层检查（user_service.go）

### ✅ 1. domainUserToDTO - 返回逻辑

**规则 1：普通字段**
- ✅ `Nickname`: 正确实现 `omitempty`，只在有值时返回
- ✅ `AlarmScope`: 正确实现 `omitempty`，只在有值时返回

**规则 2：数组字段**
- ✅ `AlarmLevels`: 正确实现 `omitempty`，只在 `len > 0` 时返回
- ✅ `AlarmChannels`: 正确实现 `omitempty`，只在 `len > 0` 时返回
- ✅ `Tags`: 正确实现，通过 JSON 解析

**规则 3：有 Hash 的字段**
- ✅ `Email`: 
  - 有值 → 返回实际值 ✅
  - 无值但 hash 有值 → 返回占位符 `"xxx@xxx.xxx"` ✅
  - 无值且 hash 无值 → 返回空字符串（Handler 层会转换为 null）✅
- ✅ `Phone`: 同 Email 逻辑 ✅
- ✅ **不返回 `email_hash` 和 `phone_hash`** ✅

**规则 4：密码字段**
- ✅ **不返回 `password` 和 `password_hash`** ✅
- ✅ `UserDTO` 结构中没有 `password_hash` 字段 ✅

### ✅ 2. UpdateUser - 更新逻辑（带值比较）

**规则 1：普通字段**
- ✅ `Nickname`: 使用 `updateStringField`，带值比较 ✅
- ✅ `Role`: 使用 `updateStringField`，带值比较 ✅
- ✅ `Status`: 使用 `updateStringField`，带值比较 ✅
- ✅ `AlarmScope`: 使用 `updateStringField`，带值比较 ✅

**规则 2：数组字段**
- ✅ `AlarmLevels`: 使用 `updateStringArrayField`，带值比较 ✅
- ✅ `AlarmChannels`: 使用 `updateStringArrayField`，带值比较 ✅
- ✅ `Tags`: 使用 `tagsEqual` 比较，带值比较 ✅

**规则 3：有 Hash 的字段**
- ✅ `Email/EmailHash`: 正确处理，支持单独更新 hash ✅
- ✅ `Phone/PhoneHash`: 正确处理，支持单独更新 hash ✅

**规则 4：密码字段**
- ✅ `PasswordHash`: 正确处理，只在提供时更新 ✅

---

## 二、Handler 层检查（user_handler.go）

### ✅ 1. GetUser/ListUsers - 返回逻辑

**规则 1：普通字段**
- ✅ `Nickname`: 只在非空时返回 ✅

**规则 2：数组字段**
- ✅ `AlarmLevels`: 只在 `len > 0` 时返回 ✅
- ✅ `AlarmChannels`: 只在 `len > 0` 时返回 ✅

**规则 3：有 Hash 的字段**
- ✅ `Email`: 
  - 空字符串 → 返回 `null` ✅
  - 占位符 → 返回占位符 ✅
  - 有值 → 返回值 ✅
- ✅ `Phone`: 同 Email 逻辑 ✅
- ✅ **不返回 `email_hash` 和 `phone_hash`** ✅

**规则 4：密码字段**
- ✅ **不返回 `password` 和 `password_hash`** ✅

### ✅ 2. UpdateUser - 解析逻辑

**规则 1：普通字段**
- ✅ 使用 `parseStringField` 统一解析 ✅
- ✅ 字段不存在 → `nil`（不更新）✅
- ✅ 字段为 `null` → 空字符串指针（清空）✅
- ✅ 字段有值 → 值指针（更新）✅

**规则 2：数组字段**
- ✅ 使用 `parseStringArrayField` 统一解析 ✅
- ✅ 字段不存在 → `nil`（不更新）✅
- ✅ 字段为 `null` → 空数组（清空）✅
- ✅ 字段有值 → 数组（更新）✅

**规则 3：有 Hash 的字段**
- ✅ `Email/Phone`: 使用 `parseStringField` 解析 ✅
- ✅ `EmailHash/PhoneHash`: 使用 `parseStringField` 解析 ✅
- ✅ **不返回 `email_hash` 和 `phone_hash`** ✅

**规则 4：密码字段**
- ✅ `PasswordHash`: 正确处理 ✅
- ✅ 字段不存在 → `nil`（不更新）✅
- ✅ 字段为 `null` → `nil`（不更新，密码不允许清空）✅
- ✅ 字段有值 → 值指针（更新）✅

---

## 三、Repository 层检查（postgres_users.go）

### ✅ UpdateUser - NULL 处理

**规则 1：普通字段**
- ✅ `Nickname`: 正确处理 `sql.NullString{Valid: false}` → NULL ✅
- ✅ `Email`: 正确处理 `sql.NullString{Valid: false}` → NULL ✅
- ✅ `Phone`: 正确处理 `sql.NullString{Valid: false}` → NULL ✅
- ✅ `AlarmScope`: 正确处理 `sql.NullString{Valid: false}` → NULL ✅

**规则 2：数组字段**
- ✅ `AlarmLevels`: 正确处理空数组 → NULL ✅
- ✅ `AlarmChannels`: 正确处理空数组 → NULL ✅
- ✅ `Tags`: 通过 JSONB 处理，空数组通过 `sql.NullString{Valid: false}` 处理 ✅

**规则 3：有 Hash 的字段**
- ✅ `EmailHash`: 正确处理 `len == 0` → NULL ✅
- ✅ `PhoneHash`: 正确处理 `len == 0` → NULL ✅

**规则 4：密码字段**
- ✅ `PasswordHash`: 正确处理 `len == 0` → NULL ✅

---

## 四、前端检查（UserDetail.vue）

### ✅ handleSave - 字段更新逻辑

**规则 1：普通字段**
- ✅ `handleStringField`: 正确实现值比较 ✅
- ✅ 值改变且不为空 → 返回新值 ✅
- ✅ 值 = "" 或 null → 返回 null（清空）✅
- ✅ 值未变 → 返回 nil（不包含字段）✅

**规则 2：数组字段**
- ✅ `handleArrayField`: 正确实现值比较 ✅
- ✅ 值改变且不为空 → 返回新数组 ✅
- ✅ 值 = [] 或 null → 返回 null（清空）✅
- ✅ 值未变 → 返回 nil（不包含字段）✅

**规则 3：有 Hash 的字段**
- ✅ `handleHashField`: 正确实现占位符逻辑 ✅
- ✅ 占位符未变 → 返回 nil（不更新）✅
- ✅ 值改变且不为空 → 返回新值 + 计算 hash ✅
- ✅ 值 = "" 或 null → 返回 null + hash = null（清空）✅

**规则 4：密码字段**
- ✅ `passwordHash`: 正确实现 ✅
- ✅ 值改变且不为空 → 返回 password_hash ✅
- ✅ 值未变 → 返回 nil（不更新）✅

### ✅ fetchUser - 原始值保存

- ✅ `originalUserData`: 正确保存所有字段的原始值 ✅
- ✅ `originalPasswordHash`: 正确保存密码 hash 的原始值 ✅

---

## 五、发现的问题

### ✅ 问题 1：Repository 层空数组处理（已修复）

**位置**：`postgres_users.go` 第 948-965 行

**修复内容**：
- ✅ `AlarmLevels`: 空数组 `[]string{}` 现在正确转换为 NULL
- ✅ `AlarmChannels`: 空数组 `[]string{}` 现在正确转换为 NULL

**修复后的代码**：
```go
// AlarmLevels 字段：空数组 → NULL
if user.AlarmLevels != nil {
    if len(user.AlarmLevels) == 0 {
        // 空数组 → NULL
        updates = append(updates, "alarm_levels = NULL")
    } else {
        updates = append(updates, fmt.Sprintf("alarm_levels = $%d", argIdx))
        args = append(args, pq.Array(user.AlarmLevels))
        argIdx++
    }
}
```

---

## 六、总结

### ✅ 符合规则的实现

1. **Service 层**：
   - ✅ 正确返回占位符
   - ✅ 不返回 hash 字段
   - ✅ 带值比较，避免不必要的更新

2. **Handler 层**：
   - ✅ 正确解析字段
   - ✅ 不返回 hash 字段
   - ✅ 正确处理 null

3. **前端**：
   - ✅ 正确实现值比较
   - ✅ 正确处理占位符
   - ✅ 正确计算并发送 hash

### ✅ 已修复的问题

1. **Repository 层**：
   - ✅ 空数组现在正确转换为 NULL（`AlarmLevels`, `AlarmChannels`）

---

## 七、建议

1. **立即修复**：Repository 层空数组处理
2. **测试验证**：确保空数组正确存储为 NULL
3. **文档更新**：如果修复了问题，更新实现检查清单

