# Current Branch 设计文档

## 背景

`is_primary` 字段没有实际业务意义，因为即使用户有多个院区，系统仍然会显示所有有权限的院区信息。

用户需要一个 `current_branch` 字段，用于：
- 当用户有多个院区时，可以选择当前工作的院区
- 在 account_setting 里选择当前的 branch
- 用这个值去过滤：ALL（所有有权限的）、DV、SP
- 在 Tenant 后面展示当前选择的 branch

## 设计方案

### 1. 数据存储

**方案：使用 `users.preferences` JSONB 字段存储 `current_branch_id`**

```json
{
  "current_branch_id": "6b7bd97a-42e4-4086-8232-ecf7ae545eec",  // branch_id 或 "ALL"
  "vitalFocus": {
    "selectedCardIds": ["card-id-1", "card-id-2"]
  }
}
```

**值说明**：
- `branch_id`（UUID 字符串）：选择特定院区（如 "DV"）
- `"ALL"`（字符串）：显示所有有权限的院区
- `null` 或不存在：默认行为（显示所有有权限的院区）

### 2. 后端 API 扩展

#### 2.1 扩展 `UpdateAccountSettingsRequest`

```go
type UpdateAccountSettingsRequest struct {
    TenantID        string  // 必填
    UserID          string  // 必填
    CurrentUserID   string  // 当前用户 ID（用于权限检查）
    PasswordHash    *string // 可选：密码 hash
    PinHash         *string // 可选：PIN hash
    Email           *string // 可选：邮箱
    EmailHash       *string // 可选：邮箱 hash
    Phone           *string // 可选：电话
    PhoneHash       *string // 可选：电话 hash
    CurrentBranchID *string // 新增：当前选择的 branch_id（"ALL" 或 branch_id）
}
```

#### 2.2 扩展 `GetAccountSettingsResponse`

```go
type GetAccountSettingsResponse struct {
    // ... 现有字段
    CurrentBranchID string `json:"current_branch_id"` // 当前选择的 branch_id（"ALL" 或 branch_id）
    CurrentBranchName string `json:"current_branch_name"` // 当前选择的 branch_name（用于显示）
    AvailableBranches []BranchDTO `json:"available_branches"` // 用户有权限的所有 branch 列表
}
```

#### 2.3 实现逻辑

1. **更新 `current_branch_id`**：
   - 验证 `CurrentBranchID` 必须是 "ALL" 或用户有权限的 `branch_id`
   - 更新 `users.preferences` JSONB 字段中的 `current_branch_id`

2. **获取 `current_branch_id`**：
   - 从 `users.preferences` 中读取 `current_branch_id`
   - 如果不存在或为 `null`，返回 "ALL"（默认值）
   - 查询对应的 `branch_name`（如果不是 "ALL"）

### 3. 前端实现

#### 3.1 Sidebar 显示

在 `Sidebar.vue` 的 Tenant 后面显示当前选择的 branch：

```vue
<div class="sidebar-actions" :class="{ collapsed: collapsed }">
  <div class="tenant-name" v-if="!collapsed">
    {{ getFirstWord(userInfo?.tenant_name) || '' }}
    <span v-if="currentBranchName" class="current-branch">
      - {{ currentBranchName }}
    </span>
  </div>
  <!-- ... -->
</div>
```

#### 3.2 Account Settings 选择器

在 `Sidebar.vue` 的 Account Settings Modal 中添加 branch 选择器：

```vue
<a-form-item label="Current Branch" v-if="availableBranches.length > 1">
  <a-select v-model:value="accountSettingsForm.current_branch_id">
    <a-select-option value="ALL">All (所有有权限的院区)</a-select-option>
    <a-select-option 
      v-for="branch in availableBranches" 
      :key="branch.branch_id" 
      :value="branch.branch_id"
    >
      {{ branch.branch_name }}
    </a-select-option>
  </a-select>
</a-form-item>
```

**注意**：只有当用户有多个 branch 时才显示选择器（`availableBranches.length > 1`）。

#### 3.3 数据过滤

使用 `current_branch_id` 来过滤数据：

```typescript
// 在 API 调用时，根据 current_branch_id 过滤
const getFilterBranchId = (): string | undefined => {
  const currentBranchId = userInfo.value?.preferences?.current_branch_id
  if (currentBranchId === 'ALL' || !currentBranchId) {
    return undefined // 显示所有有权限的院区
  }
  return currentBranchId // 只显示选中的院区
}
```

### 4. 数据流

```
用户登录
  ↓
获取用户信息（包含 preferences.current_branch_id）
  ↓
如果 current_branch_id 不存在，默认设置为 "ALL"
  ↓
在 Sidebar 显示：Tenant - Current Branch
  ↓
用户点击 Account Settings
  ↓
显示 branch 选择器（如果用户有多个 branch）
  ↓
用户选择 branch（ALL 或特定 branch_id）
  ↓
调用 UpdateAccountSettings API
  ↓
更新 users.preferences.current_branch_id
  ↓
刷新用户信息，更新 Sidebar 显示
  ↓
使用 current_branch_id 过滤数据（Overview、Resident、Unit 等）
```

### 5. 权限验证

- 验证 `CurrentBranchID` 必须是 "ALL" 或用户有权限的 `branch_id`
- 从 `user_branches` 表查询用户有权限的所有 `branch_id`
- 如果 `CurrentBranchID` 不是 "ALL" 且不在用户有权限的 `branch_id` 列表中，返回错误

### 6. 向后兼容

- `is_primary` 字段保留（用于向后兼容），但不再作为业务逻辑的关键字段
- 如果 `current_branch_id` 不存在，默认行为是显示所有有权限的院区（等同于 "ALL"）
- 如果用户只有一个 branch，`current_branch_id` 可以设置为该 branch_id 或 "ALL"（效果相同）

## 实施步骤

1. **后端**：
   - [ ] 扩展 `UpdateAccountSettingsRequest` 添加 `CurrentBranchID` 字段
   - [ ] 扩展 `GetAccountSettingsResponse` 添加 `CurrentBranchID` 和 `CurrentBranchName` 字段
   - [ ] 实现 `UpdateAccountSettings` 中更新 `preferences.current_branch_id` 的逻辑
   - [ ] 实现 `GetAccountSettings` 中读取 `preferences.current_branch_id` 的逻辑
   - [ ] 添加权限验证（验证 `CurrentBranchID` 必须是 "ALL" 或用户有权限的 `branch_id`）

2. **前端**：
   - [ ] 在 `Sidebar.vue` 的 Tenant 后面显示当前选择的 branch
   - [ ] 在 Account Settings Modal 中添加 branch 选择器
   - [ ] 更新 `UserInfo` 接口添加 `current_branch_id` 和 `current_branch_name` 字段
   - [ ] 在数据过滤逻辑中使用 `current_branch_id`（Overview、Resident、Unit 等）

3. **测试**：
   - [ ] 测试用户有多个 branch 时的选择功能
   - [ ] 测试用户只有一个 branch 时的默认行为
   - [ ] 测试 "ALL" 选项的过滤逻辑
   - [ ] 测试权限验证（不能选择没有权限的 branch）

## 注意事项

1. **默认值**：如果 `current_branch_id` 不存在，默认行为是显示所有有权限的院区（等同于 "ALL"）
2. **权限验证**：必须验证 `CurrentBranchID` 是用户有权限的 `branch_id` 或 "ALL"
3. **UI 显示**：只有当用户有多个 branch 时才显示 branch 选择器
4. **数据过滤**：使用 `current_branch_id` 来过滤数据，而不是 `is_primary`
