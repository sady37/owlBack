# Discharge Date 字段实现总结

## 实现完成 ✅

### 1. 数据库 Schema 更新

**文件**: `owlRD/db/08_residents.sql`

**变更**:
- 添加 `discharge_date DATE` 字段
- 添加 CHECK 约束：`CHECK (discharge_date IS NULL OR status IN ('discharged', 'transferred'))`
- 添加索引：`idx_residents_discharge_date`（用于查询优化）
- 添加字段注释说明

**约束逻辑**:
- `discharge_date` 仅在 `status = 'discharged'` 或 `status = 'transferred'` 时可以有值
- 如果 `status = 'active'`，`discharge_date` 必须为 NULL

---

### 2. 后端 API 更新

**文件**: `owlBack/wisefido-data/internal/http/admin_residents_handlers.go`

#### GET /admin/api/v1/residents (列表)
- ✅ 查询中添加 `r.discharge_date`
- ✅ Scan 中添加 `dischargeDate` 变量
- ✅ 返回结果中包含 `discharge_date` 字段（如果有效）

#### GET /admin/api/v1/residents/:id (详情)
- ✅ 查询中添加 `r.discharge_date`
- ✅ Scan 中添加 `dischargeDate` 变量
- ✅ 返回结果中包含 `discharge_date` 字段（如果有效）

#### POST /admin/api/v1/residents (创建)
- ✅ 支持 `discharge_date` 参数
- ✅ 验证：仅在 `status = 'discharged'` 或 `'transferred'` 时设置 `discharge_date`
- ✅ INSERT 语句中包含 `discharge_date` 字段

#### PUT /admin/api/v1/residents/:id (更新)
- ✅ 支持 `discharge_date` 参数
- ✅ 验证：仅在 `status = 'discharged'` 或 `'transferred'` 时允许设置 `discharge_date`
- ✅ 如果 `status` 不是 `discharged`/`transferred`，不允许设置 `discharge_date`
- ✅ 支持清空 `discharge_date`（设置为 NULL）

---

### 3. 前端实现

**文件**: `owlFront/src/views/residents/components/ResidentProfileContent.vue`

**状态**:
- ✅ UI 已实现（日期选择器）
- ✅ 模型已定义（`residentModel.ts`）
- ✅ 无需修改（前端已支持）

---

## 业务逻辑

### 数据一致性规则
1. **创建时**：
   - 如果 `status = 'active'`，`discharge_date` 必须为 NULL
   - 如果 `status = 'discharged'` 或 `'transferred'`，可以设置 `discharge_date`

2. **更新时**：
   - 如果更新 `status` 为 `'active'`，应该清空 `discharge_date`
   - 如果更新 `status` 为 `'discharged'` 或 `'transferred'`，可以设置 `discharge_date`
   - 如果 `status` 不是 `discharged`/`transferred`，不允许设置 `discharge_date`

3. **数据库约束**：
   - CHECK 约束确保数据一致性
   - 如果违反约束，数据库会拒绝插入/更新

---

## 测试建议

1. **创建 resident**：
   - `status = 'active'`，不提供 `discharge_date` → 应该成功
   - `status = 'active'`，提供 `discharge_date` → 应该失败（数据库约束）
   - `status = 'discharged'`，提供 `discharge_date` → 应该成功

2. **更新 resident**：
   - 更新 `status` 为 `'discharged'`，同时设置 `discharge_date` → 应该成功
   - 更新 `status` 为 `'active'`，`discharge_date` 应该被清空（或保持不变，但数据库约束会阻止）

3. **查询**：
   - GET 列表应该返回 `discharge_date` 字段
   - GET 详情应该返回 `discharge_date` 字段

---

## 数据库迁移

**注意**：如果数据库已经存在，需要执行迁移：

```sql
ALTER TABLE residents ADD COLUMN discharge_date DATE;
ALTER TABLE residents ADD CONSTRAINT chk_residents_discharge_date 
    CHECK (discharge_date IS NULL OR status IN ('discharged', 'transferred'));
CREATE INDEX IF NOT EXISTS idx_residents_discharge_date 
    ON residents(tenant_id, discharge_date) WHERE discharge_date IS NOT NULL;
COMMENT ON COLUMN residents.discharge_date IS '出院日期：用于记录住户出院/转院的具体日期。仅在 status = ''discharged'' 或 ''transferred'' 时应该有值。用于历史记录查询、统计报表、业务逻辑判断（如设备管理）。非 PII 信息，可以放在 residents 表';
```

