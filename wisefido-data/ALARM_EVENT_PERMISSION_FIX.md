# AlarmEventService 权限检查修复

## 📋 修复内容

### 问题描述

`AlarmEventService` 的 `checkHandlePermission` 方法权限检查不完整：
- ✅ 只检查了 Facility/Home 类型（Facility 只有 Nurse/Caregiver 可以处理）
- ❌ 缺少 `assigned_only` 权限检查（Caregiver/Nurse 只能处理分配的住户）
- ❌ 缺少 `branch_only` 权限检查（Manager 只能处理同分支的住户）

---

## ✅ 修复方案

### 1. 修改 `checkHandlePermission` 方法

**文件**：`internal/service/alarm_event_service.go`

**修改内容**：
1. 添加 `userID` 和 `userType` 参数
2. 添加 `getResidentByDeviceID` 方法（通过 device_id 获取住户信息）
3. 添加 `isResidentAssignedToUser` 方法（检查住户是否分配给用户）
4. 添加 `getResourcePermission` 方法（查询权限配置）
5. 实现完整的权限检查逻辑：
   - Facility 类型：只有 Nurse/Caregiver 可以处理
   - Home 类型：
     - Caregiver/Nurse：检查 `assigned_only` 权限
     - Manager：检查 `branch_only` 权限

---

### 2. 添加权限检查辅助方法

#### `getResourcePermission`

查询 `role_permissions` 表，获取角色的权限配置（`assigned_only`、`branch_only`）。

**实现**：
```go
func (s *alarmEventService) getResourcePermission(ctx context.Context, roleCode, resourceType, permissionType string) (*PermissionCheck, error)
```

#### `getResidentByDeviceID`

通过 `device_id` 获取关联的住户信息（包括 `resident_id`、`branch_tag`、`unit_id`）。

**查询路径**：
- `devices → beds → residents`
- `devices → rooms → units → residents`

**实现**：
```go
func (s *alarmEventService) getResidentByDeviceID(ctx context.Context, tenantID, deviceID string) (*residentInfo, error)
```

#### `isResidentAssignedToUser`

检查住户是否分配给该用户（查询 `resident_caregivers` 表的 `userList` JSONB 字段）。

**实现**：
```go
func (s *alarmEventService) isResidentAssignedToUser(ctx context.Context, tenantID, residentID, userID string) bool
```

---

### 3. 修改 `HandleAlarmEventRequest` 结构体

**文件**：`internal/service/alarm_event_service.go`

**添加字段**：
```go
CurrentUserType string // 当前用户类型：'resident' | 'family' | 'staff'（用于权限检查）
```

---

### 4. 修改 Handler 层

**文件**：`internal/http/alarm_event_handler.go`

**修改内容**：
- 从 HTTP Header 中获取 `X-User-Type`
- 如果为空，默认为 `"staff"`
- 传递给 Service 层的 `HandleAlarmEventRequest`

---

## 🔐 权限规则

### Facility 类型卡片

- **规则**：只有 Nurse 或 Caregiver 可以处理
- **实现**：直接检查 `userRole`

---

### Home 类型卡片

#### 1. Caregiver/Nurse

- **规则**：如果配置了 `assigned_only=true`，只能处理分配的住户
- **实现**：
  1. 查询 `role_permissions` 表，获取 `assigned_only` 配置
  2. 如果 `assigned_only=true`，检查住户是否分配给该用户
  3. 通过 `resident_caregivers` 表的 `userList` 字段检查

#### 2. Manager

- **规则**：如果配置了 `branch_only=true`，只能处理同分支的住户
- **实现**：
  1. 查询 `role_permissions` 表，获取 `branch_only` 配置
  2. 如果 `branch_only=true`，检查用户的 `branch_tag` 和住户的 `branch_tag` 是否匹配
  3. 特殊处理：
     - 用户 `branch_tag` 为 NULL：只能访问 `branch_tag` 为 NULL 或 '-' 的住户
     - 用户 `branch_tag` 有值：只能访问匹配的 `branch_tag` 的住户

#### 3. 其他角色

- **规则**：默认允许（SystemAdmin 等）

---

## 📝 代码变更

### 修改的文件

1. **`internal/service/alarm_event_service.go`**
   - 修改 `checkHandlePermission` 方法签名和实现
   - 添加 `getResourcePermission` 方法
   - 添加 `getResidentByDeviceID` 方法
   - 添加 `isResidentAssignedToUser` 方法
   - 添加 `PermissionCheck` 和 `residentInfo` 结构体
   - 修改 `HandleAlarmEventRequest` 结构体（添加 `CurrentUserType` 字段）

2. **`internal/http/alarm_event_handler.go`**
   - 从 HTTP Header 获取 `X-User-Type`
   - 传递给 Service 层

---

## ✅ 验证

### 编译检查

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go build ./internal/service/alarm_event_service.go
go build ./internal/http/alarm_event_handler.go
```

**结果**：✅ 编译通过

---

## 🔗 参考实现

参考了 `SleepaceReportHandler` 的权限检查实现：
- `checkReportPermission` 方法
- `getResidentByDeviceID` 方法
- `isResidentAssignedToUser` 方法
- `GetResourcePermission` 函数（在 `permission_utils.go` 中）

---

## 📊 权限检查流程图

```
HandleAlarmEvent
    ↓
checkHandlePermission
    ↓
1. 查询卡片信息（通过 device_id）
    ↓
2. 查询卡片的 unit_type
    ↓
3. Facility 类型？
    ├─ 是 → 检查 userRole（Nurse/Caregiver）
    └─ 否 → 继续
    ↓
4. Home 类型？
    ├─ 是 → 获取住户信息（通过 device_id）
    │   ↓
    │   5. Staff 角色？
    │   ├─ Caregiver/Nurse → 检查 assigned_only
    │   │   ├─ assigned_only=true → 检查住户分配
    │   │   └─ assigned_only=false → 允许
    │   │
    │   ├─ Manager → 检查 branch_only
    │   │   ├─ branch_only=true → 检查 branch_tag 匹配
    │   │   └─ branch_only=false → 允许
    │   │
    │   └─ 其他角色 → 允许
    │
    └─ 否 → 允许（fallback）
```

---

## 🎯 完成状态

- ✅ 修改 `checkHandlePermission` 方法
- ✅ 添加 `assigned_only` 权限检查
- ✅ 添加 `branch_only` 权限检查
- ✅ 添加辅助方法（`getResourcePermission`、`getResidentByDeviceID`、`isResidentAssignedToUser`）
- ✅ 修改 `HandleAlarmEventRequest` 结构体
- ✅ 修改 Handler 层
- ✅ 编译通过
- ✅ 与 `SleepaceReportHandler` 权限检查逻辑保持一致

---

## 📝 后续建议

1. **测试**：添加单元测试和集成测试，验证权限检查逻辑
2. **文档**：更新 API 文档，说明权限规则
3. **日志**：添加权限检查失败的日志记录（用于审计）

