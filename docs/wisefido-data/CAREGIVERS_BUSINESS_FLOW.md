# Caregivers 和 Caregiver Group 业务流程分析

## 一、前端实现（ResidentProfileContent.vue）

### 1. Available Caregivers 计算

**代码位置**：`ResidentProfileContent.vue` 第 1109-1128 行

```typescript
const fetchCaregivers = async () => {
  try {
    const tenantId = userStore.getUserInfo?.tenant_id
    if (!tenantId) return
    
    const result = await getUsersApi()
    // Filter: role='Nurse' or 'Caregiver' and status='active'
    availableCaregivers.value = (result.items || []).filter(
      (user: User) => 
        (user.role === 'Nurse' || user.role === 'Caregiver') && 
        user.status === 'active'
    )
    console.log('Caregivers loaded:', availableCaregivers.value)
    // Update display after loading caregivers
    updateCaregiversDisplay()
  } catch (error: any) {
    console.error('Failed to fetch caregivers:', error)
    availableCaregivers.value = []
  }
}
```

**计算逻辑**：
- ✅ 调用 `getUsersApi()` 获取所有用户
- ✅ **前端过滤**：`role='Nurse' or 'Caregiver'` 且 `status='active'`
- ✅ 存储在 `availableCaregivers.value` 中

**调用时机**：
- ✅ 在 `onMounted` 时调用（第 1872 行）
- ✅ 与 `fetchServiceLevels()`, `fetchBranches()`, `fetchCaregiverTags()` 并行调用

**返回数据**：
- ✅ 返回过滤后的用户列表（`User[]`）
- ✅ 包含 `user_id`, `nickname`, `email`, `phone`, `role` 等字段

---

### 2. Available Caregiver Groups 计算

**代码位置**：`ResidentProfileContent.vue` 第 1146-1150 行

```typescript
const fetchCaregiverTags = async () => {
  // TODO: Re-implement using users.user_tags directly
  // user_tags are stored as JSONB array in users table
  availableCaregiverTags.value = []
}
```

**当前状态**：
- ⚠️ **功能被禁用**：返回空数组 `[]`
- ⚠️ **TODO 注释**：需要重新实现，使用 `users.user_tags` 直接查询

**调用时机**：
- ✅ 在 `onMounted` 时调用（第 1873 行）
- ✅ 与 `fetchServiceLevels()`, `fetchBranches()`, `fetchCaregivers()` 并行调用

**返回数据**：
- ⚠️ 当前返回空数组 `[]`
- ⚠️ 应该返回从 `users.user_tags` 提取的标签列表

---

### 3. 数据返回（getCaregiversData）

**代码位置**：`ResidentProfileContent.vue` 第 1675-1680 行

```typescript
const getCaregiversData = () => {
  return {
    userList: selectedCaregiverIds.value,
    groupList: selectedCaregiverTagIds.value,
  }
}
```

**返回内容**：
- ✅ `userList`: 选中的 caregiver 用户 ID 列表（`string[]`）
- ✅ `groupList`: 选中的 caregiver group 标签列表（`string[]`）

**暴露方式**：
- ✅ 通过 `defineExpose` 暴露给父组件（第 1682-1687 行）
- ✅ 父组件在保存时调用 `getCaregiversData()` 获取数据

---

## 二、后端实现

### 1. 数据结构

**代码位置**：`resident_service.go` 第 264-268 行

```go
// ResidentCaregiverDTO 住户护理人员分配 DTO
type ResidentCaregiverDTO struct {
	UserList  []string `json:"user_list,omitempty"`  // JSONB array -> []string
	GroupList []string `json:"group_list,omitempty"` // JSONB array -> []string
}
```

**存储位置**：
- ✅ 存储在 `resident_caregivers` 表中
- ✅ `user_list`: JSONB array，存储用户 ID 列表
- ✅ `group_list`: JSONB array，存储用户组标签列表（用于匹配 `users.user_tags`）

---

### 2. 数据返回（GetResident）

**代码位置**：`resident_service.go` 第 1329-1350 行

```go
// 7. 查询 Caregivers 数据（必须查询，Profile Tab 需要）
var caregivers *ResidentCaregiverDTO
var userListRaw, groupListRaw sql.NullString

err = tx.QueryRowContext(ctx,
	`SELECT user_list, group_list
	 FROM resident_caregivers
	 WHERE tenant_id = $1 AND resident_id = $2`,
	tenantID, residentID,
).Scan(&userListRaw, &groupListRaw)

if err == nil {
	caregivers = &ResidentCaregiverDTO{}
	
	if userListRaw.Valid && userListRaw.String != "" && userListRaw.String != "null" {
		var userList []string
		if err := json.Unmarshal([]byte(userListRaw.String), &userList); err == nil {
			caregivers.UserList = userList
		}
	}
	
	if groupListRaw.Valid && groupListRaw.String != "" && groupListRaw.String != "null" {
		var groupList []string
		if err := json.Unmarshal([]byte(groupListRaw.String), &groupList); err == nil {
			caregivers.GroupList = groupList
		}
	}
}
```

**返回逻辑**：
- ✅ 从 `resident_caregivers` 表查询 `user_list` 和 `group_list`
- ✅ 解析 JSONB 数组为 `[]string`
- ✅ 包含在 `GetResidentResponse` 的 `Caregivers` 字段中

---

### 3. 数据保存（CreateResident/UpdateResident）

**代码位置**：`resident_service.go` 第 386-394 行

```go
// CreateResidentCaregiverRelation Resident 与 Caregiver 的关系创建结构体
type CreateResidentCaregiverRelation struct {
	UserList  []string // 用户ID列表（可选，JSONB array）
	GroupList []string // 用户组标签列表（可选，JSONB array，用于匹配 users.user_tags）
	// 说明：
	//   - 每个租户+住户最多一条记录（UNIQUE(tenant_id, resident_id)）
	//   - 如果 user_list 和 group_list 都为空，使用默认告警路由规则（由应用层处理）
}
```

**保存逻辑**：
- ✅ 接收前端的 `userList` 和 `groupList`
- ✅ 存储到 `resident_caregivers` 表
- ✅ 如果两者都为空，使用默认告警路由规则

---

## 三、业务流程总结

### 1. Available Caregivers 计算流程

```
前端 onMounted
  ↓
调用 fetchCaregivers()
  ↓
调用 getUsersApi() 获取所有用户
  ↓
前端过滤：role='Nurse' or 'Caregiver' and status='active'
  ↓
存储到 availableCaregivers.value
  ↓
在 Modal 中显示供用户选择
```

**问题**：
- ⚠️ **前端过滤**：所有计算在前端完成，后端不参与
- ⚠️ **性能问题**：需要获取所有用户，然后前端过滤
- ⚠️ **权限问题**：前端可以看到所有用户，没有权限控制

**建议**：
- ✅ 后端应该提供专门的 API：`GET /admin/api/v1/users/caregivers`
- ✅ 后端过滤：`role='Nurse' or 'Caregiver'` 且 `status='active'`
- ✅ 后端权限控制：根据当前用户的权限过滤可见的 caregivers

---

### 2. Available Caregiver Groups 计算流程

```
前端 onMounted
  ↓
调用 fetchCaregiverTags()
  ↓
当前实现：返回空数组 []
  ↓
TODO: 需要重新实现
```

**当前状态**：
- ⚠️ **功能被禁用**：`fetchCaregiverTags` 返回空数组
- ⚠️ **TODO 注释**：需要重新实现，使用 `users.user_tags` 直接查询

**应该的实现**：
- ✅ 后端应该提供专门的 API：`GET /admin/api/v1/users/caregiver-groups`
- ✅ 后端查询：从 `users.user_tags` JSONB 数组中提取所有唯一的标签
- ✅ 后端过滤：只返回 `role='Nurse' or 'Caregiver'` 且 `status='active'` 的用户的标签
- ✅ 后端权限控制：根据当前用户的权限过滤可见的标签

---

### 3. 数据保存流程

```
用户选择 Caregivers 和 Caregiver Groups
  ↓
点击 Save 按钮
  ↓
父组件调用 getCaregiversData()
  ↓
返回 { userList: string[], groupList: string[] }
  ↓
发送到后端 CreateResident/UpdateResident API
  ↓
后端存储到 resident_caregivers 表
```

**数据流**：
- ✅ 前端：`selectedCaregiverIds.value` → `userList`
- ✅ 前端：`selectedCaregiverTagIds.value` → `groupList`
- ✅ 后端：接收 `userList` 和 `groupList`，存储到 `resident_caregivers` 表

---

### 4. 数据返回流程

```
前端调用 GetResident API
  ↓
后端查询 resident_caregivers 表
  ↓
解析 JSONB 数组为 []string
  ↓
返回 { caregivers: { user_list: string[], group_list: string[] } }
  ↓
前端通过 watch 初始化 selectedCaregiverIds 和 selectedCaregiverTagIds
```

**数据流**：
- ✅ 后端：`resident_caregivers.user_list` (JSONB) → `caregivers.user_list` ([]string)
- ✅ 后端：`resident_caregivers.group_list` (JSONB) → `caregivers.group_list` ([]string)
- ✅ 前端：`caregivers.user_list` → `selectedCaregiverIds.value`
- ✅ 前端：`caregivers.group_list` → `selectedCaregiverTagIds.value`

---

## 四、发现的问题

### ⚠️ 问题 1：Available Caregivers 计算在前端

**问题**：
- 前端调用 `getUsersApi()` 获取所有用户，然后前端过滤
- 性能问题：需要获取所有用户
- 权限问题：前端可以看到所有用户

**建议**：
- 后端应该提供专门的 API：`GET /admin/api/v1/users/caregivers`
- 后端过滤：`role='Nurse' or 'Caregiver'` 且 `status='active'`
- 后端权限控制：根据当前用户的权限过滤

---

### ⚠️ 问题 2：Available Caregiver Groups 功能被禁用

**问题**：
- `fetchCaregiverTags` 返回空数组
- TODO 注释：需要重新实现，使用 `users.user_tags` 直接查询

**建议**：
- 后端应该提供专门的 API：`GET /admin/api/v1/users/caregiver-groups`
- 后端查询：从 `users.user_tags` JSONB 数组中提取所有唯一的标签
- 后端过滤：只返回 `role='Nurse' or 'Caregiver'` 且 `status='active'` 的用户的标签

---

### ⚠️ 问题 3：没有权限控制

**问题**：
- 前端可以看到所有用户（通过 `getUsersApi()`）
- 没有根据当前用户的权限过滤可见的 caregivers

**建议**：
- 后端应该根据当前用户的权限过滤可见的 caregivers
- 例如：Manager 只能看到自己 branch 的 caregivers

---

## 五、建议的改进方案

### 1. 后端 API：Get Available Caregivers

**API 路径**：`GET /admin/api/v1/users/caregivers`

**功能**：
- 查询 `role='Nurse' or 'Caregiver'` 且 `status='active'` 的用户
- 根据当前用户的权限过滤（例如：Manager 只能看到自己 branch 的 caregivers）
- 返回用户列表（`user_id`, `nickname`, `email`, `phone`, `role`）

**实现位置**：
- Handler: `user_handler.go`
- Service: `user_service.go` (新增 `GetAvailableCaregivers` 方法)

---

### 2. 后端 API：Get Available Caregiver Groups

**API 路径**：`GET /admin/api/v1/users/caregiver-groups`

**功能**：
- 从 `users.user_tags` JSONB 数组中提取所有唯一的标签
- 只返回 `role='Nurse' or 'Caregiver'` 且 `status='active'` 的用户的标签
- 根据当前用户的权限过滤（例如：Manager 只能看到自己 branch 的用户的标签）
- 返回标签列表（`tag_name`, `member_count` 等）

**实现位置**：
- Handler: `user_handler.go`
- Service: `user_service.go` (新增 `GetAvailableCaregiverGroups` 方法)

---

### 3. 前端修改

**修改 `fetchCaregivers`**：
```typescript
const fetchCaregivers = async () => {
  try {
    const result = await getAvailableCaregiversApi() // 新的 API
    availableCaregivers.value = result.items || []
    updateCaregiversDisplay()
  } catch (error: any) {
    console.error('Failed to fetch caregivers:', error)
    availableCaregivers.value = []
  }
}
```

**修改 `fetchCaregiverTags`**：
```typescript
const fetchCaregiverTags = async () => {
  try {
    const result = await getAvailableCaregiverGroupsApi() // 新的 API
    availableCaregiverTags.value = result.items || []
  } catch (error: any) {
    console.error('Failed to fetch caregiver groups:', error)
    availableCaregiverTags.value = []
  }
}
```

---

## 六、总结

### 当前实现

1. **Available Caregivers**：
   - ✅ 前端调用 `getUsersApi()` 获取所有用户
   - ✅ 前端过滤：`role='Nurse' or 'Caregiver'` 且 `status='active'`
   - ⚠️ 性能问题：需要获取所有用户
   - ⚠️ 权限问题：前端可以看到所有用户

2. **Available Caregiver Groups**：
   - ⚠️ 功能被禁用：返回空数组
   - ⚠️ TODO：需要重新实现，使用 `users.user_tags` 直接查询

3. **数据保存和返回**：
   - ✅ 正确实现：通过 `resident_caregivers` 表存储和返回

### 建议改进

1. **后端提供专门的 API**：
   - `GET /admin/api/v1/users/caregivers`
   - `GET /admin/api/v1/users/caregiver-groups`

2. **后端过滤和权限控制**：
   - 后端过滤：`role='Nurse' or 'Caregiver'` 且 `status='active'`
   - 后端权限控制：根据当前用户的权限过滤可见的 caregivers 和 groups

3. **前端简化**：
   - 前端直接调用新的 API，不需要自己过滤
   - 前端不需要处理权限逻辑

