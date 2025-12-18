# PUT /residents/:id/phi 权限讨论

## 问题 1: Resident/Family 是否应该能够更新 PHI？

### 当前代码实现
- 行 921-945：Resident/Family 可以通过自检更新 PHI
- Resident 可以更新自己的 PHI
- Family 可以更新关联住户的 PHI

### 用户反馈
- **Resident/Family 不可见，不能更新PHI**

### 数据库权限配置
- Resident/Family 在 `resident_phi` 表中**没有权限记录**
- 这意味着 Resident/Family 不应该有访问 PHI 的权限

### 前端代码
- `canViewPHI`: 只允许 Manager, Admin, Nurse, Caregiver（**不包括 Resident/Family**）
- `canEditPHI`: 只允许 Manager/Admin（**不包括 Nurse/Caregiver/Resident/Family**）

### 结论
✅ **Resident/Family 不应该能够更新 PHI**
- PHI 是敏感的健康信息，应该由医护人员管理
- 前端已经限制了 Resident/Family 的访问
- 后端应该拒绝 Resident/Family 的更新请求

---

## 问题 2: Nurse 是否应该能够更新分配的住户的 PHI？

### 当前数据库配置
- Nurse 在 `resident_phi` 表中只有 **R 权限**（查看），**没有 U 权限**（更新）
- `assigned_only=TRUE`，`branch_only=FALSE`

### 前端代码
- `canViewPHI`: ✅ Nurse 可以查看
- `canEditPHI`: ❌ Nurse **不能编辑**（只有 Manager/Admin）

### 业务场景讨论

#### 场景 A: Nurse 不能更新 PHI（当前设计）
**理由：**
- PHI 是敏感的健康信息，需要更高级别的权限
- 只有 Manager/Admin 可以更新，确保数据准确性
- Nurse 可以查看 PHI 以便提供护理服务，但更新需要 Manager 审核

**适用场景：**
- 严格的医疗数据管理
- 需要审计追踪所有 PHI 变更
- Manager 负责数据准确性

#### 场景 B: Nurse 可以更新分配的住户的 PHI
**理由：**
- Nurse 在日常护理中可能需要更新某些 PHI 信息（如体重、血压等）
- 如果每次都需要 Manager 更新，会影响工作效率
- Nurse 只对分配的住户有权限，范围受限

**适用场景：**
- 需要 Nurse 实时更新护理相关数据
- 提高工作效率
- 信任 Nurse 的专业判断

---

## 建议

### 方案 1: 保持当前设计（Nurse 不能更新 PHI）
- ✅ 符合当前数据库配置
- ✅ 符合前端代码逻辑
- ✅ 数据安全性更高
- ✅ 需要修改：移除 Resident/Family 的自检逻辑

### 方案 2: 允许 Nurse 更新分配的住户的 PHI
- ⚠️ 需要修改数据库：为 Nurse 添加 U 权限（`assigned_only=TRUE`）
- ⚠️ 需要修改前端：允许 Nurse 编辑 PHI
- ✅ 提高工作效率
- ⚠️ 需要评估安全风险

---

## 请确认

1. **Resident/Family 权限：**
   - [ ] 确认 Resident/Family 不能更新 PHI（移除自检逻辑）

2. **Nurse 权限：**
   - [ ] 方案 1：保持当前设计（Nurse 不能更新 PHI）
   - [ ] 方案 2：允许 Nurse 更新分配的住户的 PHI（需要修改数据库和前端）

请选择方案，我将据此修改代码。

