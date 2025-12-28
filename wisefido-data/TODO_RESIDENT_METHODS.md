# Resident Service 方法修复清单

## 1. ✅ 已完成
- ✅ `UpdateResident` - 已使用 `UpdateResidentFields` repository 方法
- ✅ `DeleteResident` - 已使用 `UpdateResidentFields` repository 方法
- ✅ `ResetResidentPassword` - 已使用 `UpdateResidentFields` repository 方法
- ✅ `UpdateResidentContact` - 已使用 `UpdateResidentContactFields` repository 方法

## 2. ⚠️ 需要修复的方法

### 2.1 `CreateResident` 方法
**问题**：
- Line 1460: 使用了 `s.residentsRepo.UpsertResidentPHI`（旧的 Deprecated 方法）
- 应该改为：`s.residentsRepo.UpsertResidentPHIFields`

**修复内容**：
- 需要将 PHI 数据转换为 `domain.ResidentPHIUpdate` 类型
- 使用 `UpsertResidentPHIFields` 方法

### 2.2 `UpdateResidentAccountSettings` 方法
**问题**：
- Line 2659: 使用了 `contact_id` 作为查询条件，但 `resident_contacts` 表的主键是 `(resident_id, slot)`，不是 `contact_id`
- Line 2510: `GetResidentAccountSettings` 方法也使用了 `contact_id`
- Line 2622, 2636, 2651: 尝试更新 `resident_contacts` 表中不存在的字段 `password_hash`, `email_hash`, `phone_hash`（这些字段已从表中删除）
- 直接使用 SQL 更新，没有使用 repository 方法

**修复内容**：
- 需要知道 `slot` 才能更新 contact
- 删除对 `password_hash`, `email_hash`, `phone_hash` 的更新（联系人不登录系统）
- 使用 `UpdateResidentContactFields` repository 方法

### 2.3 `GetResidentAccountSettings` 方法
**问题**：
- Line 2510: 使用了 `contact_id` 查询，但表的主键是 `(resident_id, slot)`
- 需要知道 `slot` 才能查询 contact

**修复内容**：
- 需要传入 `slot` 参数
- 修改查询条件为 `(tenant_id, resident_id, slot)`

### 2.4 DTO 中的废弃字段
**问题**：
- Line 153, 200: `ResidentListItemDTO.FamilyTag` - 字段已从数据库删除
- Line 258: `ResidentContactDTO.ContactFamilyTag` - 字段已从数据库删除
- Line 258, 259: `ResidentContactDTO.IsEmergencyContact` - 字段已重命名为 `IsEnabled`
- Line 281: `CreateResidentRequest.FamilyTag` - 字段已从数据库删除

**修复内容**：
- 删除所有 `FamilyTag` 相关字段
- 将 `IsEmergencyContact` 改为 `IsEnabled`

### 2.5 `ResetContactPassword` 方法
**问题**：
- 联系人不登录系统，不需要密码，这个方法应该被删除或标记为 Deprecated

**修复内容**：
- 评估是否需要此方法
- 如果不需要，删除或标记为 Deprecated

## 3. 优先级建议

1. **高优先级**：
   - `CreateResident` - 使用旧方法，影响创建功能
   - `UpdateResidentAccountSettings` - 使用了错误的字段和查询条件
   - `GetResidentAccountSettings` - 使用了错误的查询条件

2. **中优先级**：
   - DTO 字段清理 - 不影响功能，但会产生混淆

3. **低优先级**：
   - `ResetContactPassword` - 需要业务确认是否还需要
