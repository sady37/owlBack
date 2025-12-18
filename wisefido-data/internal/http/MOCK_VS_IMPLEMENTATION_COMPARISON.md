# Mock vs 实际实现对比

## 1. GET /admin/api/v1/residents (获取列表)

### Mock 实现
- 搜索字段：`nickname`, `unit_name`, `location_tag`
- 过滤：`status`, `service_level`
- 返回字段：包含所有 Resident 字段（包括 email, phone, location_tag）

### 实际实现
- 搜索字段：`nickname`, `unit_name` ✅ (缺少 `location_tag`，但 mock 中 `location_tag` 已改为 `branch_tag`)
- 过滤：`status`, `service_level` ✅
- 返回字段：
  - ✅ 包含：resident_id, tenant_id, resident_account, nickname, status, service_level, admission_date, discharge_date, family_tag, unit_id, unit_name, branch_tag, area_tag, unit_number, is_multi_person_room, room_id, room_name, bed_id, bed_name
  - ❌ **缺少**：`email`, `phone` (这些字段不在 residents 表中，可能在 resident_phi 表中)
  - ❌ **缺少**：`location_tag` (已改为 `branch_tag`，但 mock 中仍使用 `location_tag`)

### 差异
1. **email/phone**: Mock 返回这些字段，但实际实现不返回（因为 residents 表不存储这些字段）
2. **location_tag vs branch_tag**: Mock 使用 `location_tag`，实际实现使用 `branch_tag`

---

## 2. POST /admin/api/v1/residents (创建)

### Mock 实现
- 参数：`resident_account`, `nickname`, `email`, `phone`, `status`, `service_level`, `admission_date`, `unit_id`, `note`
- 自动生成：`resident_account` (格式: `R${id}`)
- 返回：`{ resident_id: string }`

### 实际实现
- 参数：支持所有 CreateResidentParams 字段
- 自动生成：`resident_account` (从 nickname 生成，格式: `lowercase(nickname.replace(' ', '-'))`)
- **额外处理**：
  - ✅ 生成 `resident_account_hash` 和 `password_hash`
  - ✅ 如果提供了 `first_name`/`last_name`，自动创建 PHI 记录
  - ✅ 支持 `is_access_enabled` (映射到 `can_view_status`)
  - ✅ 支持 `family_tag`
- 返回：`{ resident_id: string }` ✅

### 差异
1. **resident_account 生成规则不同**:
   - Mock: `R${id}` (如 `R001`, `R002`)
   - 实际: `lowercase(nickname.replace(' ', '-'))` (如 `john-doe`)
2. **email/phone**: Mock 接受这些参数，但实际实现不处理（因为 residents 表不存储）
3. **PHI 自动创建**: 实际实现支持，Mock 不支持

---

## 3. GET /admin/api/v1/residents/:id (获取详情)

### Mock 实现
- 参数：`include_phi`, `include_contacts`
- 返回：完整的 Resident 对象（包括 email, phone, location_tag）

### 实际实现
- 参数：`include_phi`, `include_contacts` ✅
- 返回：
  - ✅ 基本字段：resident_id, tenant_id, resident_account, nickname, status, service_level, admission_date, discharge_date, family_tag, unit_id, unit_name, branch_tag, area_tag, unit_number, is_multi_person_room, room_id, room_name, bed_id, bed_name, note, is_access_enabled
  - ✅ 如果 `include_phi=true`: 返回 PHI 数据（first_name, last_name 等）
  - ✅ 如果 `include_contacts=true`: 返回 contacts 数组
  - ❌ **缺少**：`email`, `phone` (这些字段不在 residents 表中)

### 差异
1. **email/phone**: Mock 返回，实际实现不返回（因为 residents 表不存储）
2. **location_tag vs branch_tag**: Mock 使用 `location_tag`，实际实现使用 `branch_tag`

---

## 4. PUT /admin/api/v1/residents/:id (更新)

### Mock 实现
- 参数：所有 UpdateResidentParams 字段
- 行为：直接更新所有提供的字段

### 实际实现
- 参数：支持所有 UpdateResidentParams 字段 ✅
- 行为：
  - ✅ 只更新提供的字段（部分更新）
  - ✅ 支持 NULL 值（如 `service_level: ""` 会设置为 NULL）
  - ✅ 日期格式转换（`admission_date`, `discharge_date`）

### 差异
1. **部分更新**: 实际实现支持部分更新，Mock 也支持（但实现方式不同）
2. **NULL 处理**: 实际实现明确支持设置 NULL，Mock 不明确

---

## 5. DELETE /admin/api/v1/residents/:id (删除)

### Mock 实现
- 行为：从数组中删除（硬删除）

### 实际实现
- 行为：软删除（设置 `status = 'discharged'`）✅

### 差异
1. **删除方式**: Mock 是硬删除，实际实现是软删除（符合业务需求）

---

## 6. PUT /admin/api/v1/residents/:id/phi (更新 PHI)

### Mock 实现
- 参数：所有 UpdateResidentPHIParams 字段
- 行为：更新 resident.phi 对象

### 实际实现
- 参数：支持部分 PHI 字段（目前只支持 first_name, last_name）
- 行为：使用 `INSERT ... ON CONFLICT DO UPDATE`

### 差异
1. **字段支持**: 实际实现目前只支持 first_name 和 last_name，需要扩展支持其他 PHI 字段

---

## 7. PUT /admin/api/v1/residents/:id/contacts (更新联系人)

### Mock 实现
- 参数：UpdateResidentContactParams
- 行为：如果提供了 `contact_id`，更新现有联系人；否则添加新联系人

### 实际实现
- ❌ **未实现**：目前只是 stub，返回 `{ success: true }`

### 差异
1. **功能缺失**: 实际实现需要完整实现联系人更新逻辑

---

## 总结

### ✅ 已对齐
1. GET 列表：搜索和过滤功能
2. POST 创建：基本创建功能
3. GET 详情：基本详情获取
4. PUT 更新：基本更新功能
5. DELETE 删除：软删除（比 Mock 更合理）

### ⚠️ 需要调整
1. **email/phone 字段**:
   - Mock 返回这些字段，但实际实现不返回
   - **建议**: 如果前端需要这些字段，应该从 `resident_phi` 表读取（`resident_phone`, `resident_email`）

2. **location_tag vs branch_tag**:
   - Mock 使用 `location_tag`，实际实现使用 `branch_tag`
   - **建议**: 更新 Mock 数据，使用 `branch_tag` 替代 `location_tag`

3. **resident_account 生成规则**:
   - Mock: `R${id}`
   - 实际: `lowercase(nickname.replace(' ', '-'))`
   - **建议**: 统一生成规则，或者让前端明确指定

4. **PHI 更新**:
   - 实际实现只支持 first_name 和 last_name
   - **建议**: 扩展支持所有 PHI 字段

5. **联系人更新**:
   - 实际实现未完成
   - **建议**: 实现完整的联系人 CRUD 逻辑

