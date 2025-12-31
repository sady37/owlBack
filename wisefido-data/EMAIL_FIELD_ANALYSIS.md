# Email 字段处理逻辑分析

## 一、场景分析

### 场景 1：Email 为占位符 "xxx@xxx.xxx"（原始值也是占位符）

**代码位置**：`UserDetail.vue` 第 1060-1063 行

```typescript
if (isPlaceholder && wasPlaceholder) {
  // 仍是占位符，返回 nil（不更新）
  return
}
```

**行为**：
- ✅ **不发送任何字段**（`email` 和 `email_hash` 都不包含在 `params` 中）
- ✅ 相当于字段不存在，Handler 层会识别为 `req.Email = nil`（不更新）

**触发条件**：
- `currentValue = "xxx@xxx.xxx"`（当前值）
- `originalValue = "xxx@xxx.xxx"`（原始值）

---

### 场景 2：Email 从占位符改为实际值

**代码位置**：`UserDetail.vue` 第 1065-1076 行

```typescript
if (current !== original) {
  // 值改变
  if (current === '') {
    // 值 = "" 或 null → 返回 null + hash = null（清空）
    params[fieldName] = null
    params[hashFieldName] = null
  } else {
    // 值改变且不为空 → 返回新值 + 计算 hash
    params[fieldName] = current
    params[hashFieldName] = await hashAccount(current)
  }
}
```

**行为**：
- ✅ 发送 `"email": "new@example.com"`（新值）
- ✅ 发送 `"email_hash": "hash..."`（计算的 hash）

**触发条件**：
- `currentValue = "new@example.com"`（新值）
- `originalValue = "xxx@xxx.xxx"`（原始占位符）

---

### 场景 3：Email 从占位符改为空字符串

**代码位置**：`UserDetail.vue` 第 1067-1070 行

```typescript
if (current === '') {
  // 值 = "" 或 null → 返回 null + hash = null（清空）
  params[fieldName] = null
  params[hashFieldName] = null
}
```

**行为**：
- ✅ 发送 `"email": null`（清空）
- ✅ 发送 `"email_hash": null`（清空 hash）

**触发条件**：
- `currentValue = ""`（空字符串）
- `originalValue = "xxx@xxx.xxx"`（原始占位符）

---

### 场景 4：Email 从实际值改为占位符

**注意**：这个场景理论上不应该发生，因为：
- 用户不能直接输入占位符 "xxx@xxx.xxx"
- 占位符是后端返回的，表示有 hash 但没有实际值

**如果发生**（用户手动输入 "xxx@xxx.xxx"）：
- 会触发 `current !== original` 条件
- 会发送 `"email": "xxx@xxx.xxx"` + `"email_hash": "hash..."`（将占位符当作实际值处理）

---

### 场景 5：Email 从实际值改为另一个实际值

**行为**：
- ✅ 发送 `"email": "new@example.com"`（新值）
- ✅ 发送 `"email_hash": "hash..."`（计算的 hash）

**触发条件**：
- `currentValue = "new@example.com"`（新值）
- `originalValue = "old@example.com"`（旧值）

---

### 场景 6：Email 从实际值改为空字符串

**行为**：
- ✅ 发送 `"email": null`（清空）
- ✅ 发送 `"email_hash": null`（清空 hash）

**触发条件**：
- `currentValue = ""`（空字符串）
- `originalValue = "old@example.com"`（旧值）

---

### 场景 7：Email 值未改变

**代码位置**：`UserDetail.vue` 第 1077 行

```typescript
// 值未变 → 返回 nil（不包含在 params 中）
```

**行为**：
- ✅ **不发送任何字段**（`email` 和 `email_hash` 都不包含在 `params` 中）
- ✅ 相当于字段不存在，Handler 层会识别为 `req.Email = nil`（不更新）

**触发条件**：
- `currentValue === originalValue`（值相同）

---

## 二、Save 按钮何时可用

### Save 按钮的启用条件

**代码位置**：`UserDetail.vue` 第 7-9 行

```vue
<a-button type="primary" @click="handleSave" :loading="saving">
  Save
</a-button>
```

**启用条件**：
- ✅ 按钮始终可用（没有 `:disabled` 属性）
- ✅ 点击后会触发 `handleSave` 函数
- ✅ 在保存过程中显示 `loading` 状态

**验证逻辑**：
- ✅ 在 `handleSave` 中会先调用 `formRef.value?.validate()` 进行表单验证
- ✅ 如果验证失败，不会发送请求
- ✅ 如果验证通过，才会构建 `params` 并发送请求

---

## 三、返回值的总结

### 当 Email 为占位符 "xxx@xxx.xxx" 时

| 场景 | 原始值 | 当前值 | 发送的字段 | 说明 |
|------|--------|--------|-----------|------|
| 1 | `"xxx@xxx.xxx"` | `"xxx@xxx.xxx"` | **不发送** | 仍是占位符，不更新 |
| 2 | `"xxx@xxx.xxx"` | `"new@example.com"` | `email: "new@example.com"`<br>`email_hash: "hash..."` | 从占位符改为实际值 |
| 3 | `"xxx@xxx.xxx"` | `""` | `email: null`<br>`email_hash: null` | 从占位符清空 |
| 4 | `"old@example.com"` | `"xxx@xxx.xxx"` | `email: "xxx@xxx.xxx"`<br>`email_hash: "hash..."` | 从实际值改为占位符（不应该发生） |
| 5 | `"old@example.com"` | `"new@example.com"` | `email: "new@example.com"`<br>`email_hash: "hash..."` | 从实际值改为另一个实际值 |
| 6 | `"old@example.com"` | `""` | `email: null`<br>`email_hash: null` | 从实际值清空 |
| 7 | `"old@example.com"` | `"old@example.com"` | **不发送** | 值未改变，不更新 |

---

## 四、关键代码片段

### 1. handleHashField 函数

```typescript
const handleHashField = async (
  fieldName: string,
  hashFieldName: string,
  currentValue: string | undefined,
  originalValue: string | undefined,
  placeholder: string
) => {
  const current = currentValue?.trim() || ''
  const original = originalValue?.trim() || ''
  
  // 检查是否是占位符
  const isPlaceholder = current === placeholder
  const wasPlaceholder = original === placeholder
  
  if (isPlaceholder && wasPlaceholder) {
    // 仍是占位符，返回 nil（不更新）
    return
  }
  
  if (current !== original) {
    // 值改变
    if (current === '') {
      // 值 = "" 或 null → 返回 null + hash = null（清空）
      params[fieldName] = null
      params[hashFieldName] = null
    } else {
      // 值改变且不为空 → 返回新值 + 计算 hash
      params[fieldName] = current
      params[hashFieldName] = await hashAccount(current)
    }
  }
  // 值未变 → 返回 nil（不包含在 params 中）
}
```

### 2. 调用 handleHashField

```typescript
await handleHashField(
  'email',
  'email_hash',
  userData.value.email,
  originalUserData.value.email,
  'xxx@xxx.xxx'
)
```

---

## 五、与设计规则的对应关系

### 设计规则（FIELD_UPDATE_DESIGN.md 第 116-122 行）

```
规则 3：有 Hash 的字段（Email/Phone/Account）
情况 1：值改变且不为空 → 返回新值 + 计算 hash
情况 2：值 = "" 或 null → 返回 null + hash = null（清空）
情况 3：值未变 → 返回 nil（不包含字段，不更新）
情况 4：占位符未变 → 返回 nil（不包含字段，不更新）
```

### 实现验证

- ✅ **情况 1**：场景 2、5 正确实现
- ✅ **情况 2**：场景 3、6 正确实现
- ✅ **情况 3**：场景 7 正确实现
- ✅ **情况 4**：场景 1 正确实现

---

## 六、结论

1. **当 email 为占位符 "xxx@xxx.xxx" 且原始值也是占位符时**：
   - ✅ **不发送任何字段**（返回 nil，不更新）

2. **Save 按钮何时可用**：
   - ✅ 按钮始终可用（没有禁用条件）
   - ✅ 点击后会先进行表单验证
   - ✅ 验证通过后才会构建 `params` 并发送请求

3. **当选中 Save 时，返回什么**：
   - ✅ 根据值是否改变和是否为占位符，决定是否发送字段
   - ✅ 如果发送，会包含 `email` 和 `email_hash` 字段
   - ✅ 如果不发送，这两个字段都不包含在 `params` 中

