# Sleepace Report 权限检查实现

## ✅ 已实现

### 权限规则

根据用户要求，实现了以下权限检查规则：

1. **住户及相关联系人**：
   - ✅ 可查看住户自己的睡眠报告
   - ✅ 通过 `resident_id == user_id` 检查

2. **Caregiver/Nurse**：
   - ✅ 可查看、处理 assign-only 住户的睡眠报告
   - ✅ 检查 `role_permissions` 表中的 `assigned_only` 标志
   - ✅ 如果 `assigned_only=true`，检查 `resident_caregivers.userList` 是否包含该用户

3. **Manager**：
   - ✅ 可查看、处理 branch 住户的睡眠报告
   - ✅ 检查 `role_permissions` 表中的 `branch_only` 标志
   - ✅ 如果 `branch_only=true`：
     - 用户 `branch_tag` 为 NULL：只能访问 `branch_tag` 为 NULL 或 '-' 的住户
     - 用户 `branch_tag` 有值：只能访问匹配的 `branch_tag` 的住户

---

## 📝 实现细节

### 1. 权限检查函数

**文件**：`internal/http/sleepace_report_handler.go`

**函数**：`checkReportPermission`

**参数**：
- `ctx`: 上下文
- `tenantID`: 租户ID
- `deviceID`: 设备ID
- `userID`: 用户ID
- `userType`: 用户类型（resident, family, staff）
- `userRole`: 用户角色（Caregiver, Nurse, Manager, etc.）
- `permissionType`: 权限类型（"read" 或 "manage"）

**逻辑流程**：
1. 通过 `device_id` 获取关联的住户信息（`getResidentByDeviceID`）
2. 如果设备没有关联住户，允许访问（fallback）
3. 根据用户类型和角色进行权限检查：
   - **Resident/Family**：检查是否是自己的住户
   - **Caregiver/Nurse**：检查 `assigned_only` 和 `resident_caregivers.userList`
   - **Manager**：检查 `branch_only` 和 `branch_tag` 匹配
   - **其他角色**：默认允许（SystemAdmin 等）

---

### 2. 辅助函数

#### `getResidentByDeviceID`

**功能**：通过 `device_id` 获取关联的住户信息

**查询路径**：
- `devices` → `beds` → `residents`
- `devices` → `rooms` → `units` → `residents`

**返回**：
- `resident_id`: 住户ID
- `branch_tag`: 住户所属分支标签
- `unit_id`: 单元ID

#### `isResidentAssignedToUser`

**功能**：检查住户是否分配给该用户

**实现**：
- 查询 `resident_caregivers` 表的 `userList` 字段（JSONB 数组）
- 解析 JSONB 数组，检查 `userID` 是否在列表中

---

### 3. 权限检查位置

权限检查已添加到以下 Handler 方法：

1. ✅ `GetSleepaceReports` - 查询报告列表（read 权限）
2. ✅ `GetSleepaceReportDetail` - 查询报告详情（read 权限）
3. ✅ `GetSleepaceReportDates` - 查询有效日期列表（read 权限）
4. ✅ `DownloadReport` - 下载报告（manage 权限）

---

## 🔍 权限检查流程

```
用户请求
    ↓
提取用户信息（userID, userType, userRole）
    ↓
通过 device_id 获取关联住户信息
    ↓
根据用户类型和角色进行权限检查
    ├─ Resident/Family → 检查是否是自己的住户
    ├─ Caregiver/Nurse → 检查 assigned_only 和 userList
    ├─ Manager → 检查 branch_only 和 branch_tag
    └─ 其他角色 → 默认允许
    ↓
权限通过 → 继续处理
权限拒绝 → 返回错误
```

---

## 📋 数据库查询

### 1. 获取住户信息

```sql
SELECT DISTINCT
    r.resident_id::text,
    u.branch_tag,
    u.unit_id::text
FROM devices d
LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
LEFT JOIN rooms rm ON (d.bound_room_id = rm.room_id OR b.room_id = rm.room_id)
LEFT JOIN units u ON rm.unit_id = u.unit_id
LEFT JOIN residents r ON (r.bed_id = b.bed_id OR r.room_id = rm.room_id OR r.unit_id = u.unit_id)
WHERE d.tenant_id = $1::uuid
  AND d.device_id = $2::uuid
  AND r.resident_id IS NOT NULL
LIMIT 1
```

### 2. 检查住户分配

```sql
SELECT userList
FROM resident_caregivers
WHERE tenant_id = $1::uuid
  AND resident_id = $2::uuid
LIMIT 1
```

然后解析 JSONB 数组，检查 `userID` 是否在列表中。

---

## ✅ 测试建议

### 1. 单元测试

- 测试 `checkReportPermission` 函数
- 测试 `getResidentByDeviceID` 函数
- 测试 `isResidentAssignedToUser` 函数

### 2. 集成测试

- 测试不同用户类型的权限检查
- 测试 `assigned_only` 权限
- 测试 `branch_only` 权限
- 测试设备没有关联住户的情况（fallback）

---

## 📝 注意事项

1. **设备没有关联住户**：
   - 如果设备没有关联住户，允许访问（fallback）
   - 这适用于设备未分配或临时设备的情况

2. **权限配置**：
   - 权限配置存储在 `role_permissions` 表中
   - 如果权限记录不存在，`GetResourcePermission` 返回最严格的权限（`assigned_only=true, branch_only=true`）

3. **JSONB 解析**：
   - `resident_caregivers.userList` 是 JSONB 数组
   - 需要解析 JSONB 并检查 `userID` 是否在数组中

4. **Branch Tag 匹配**：
   - 用户 `branch_tag` 为 NULL：只能访问 `branch_tag` 为 NULL 或 '-' 的住户
   - 用户 `branch_tag` 有值：只能访问匹配的 `branch_tag` 的住户

---

## ✅ 完成状态

- ✅ 权限检查函数已实现
- ✅ 所有 Handler 方法已添加权限检查
- ✅ 辅助函数已实现
- ✅ 代码已通过 lint 检查
- ⏳ 单元测试待添加
- ⏳ 集成测试待添加

