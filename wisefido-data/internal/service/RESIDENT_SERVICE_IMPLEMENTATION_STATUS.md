# ResidentService 实现状态

## 当前状态

✅ **接口定义已完成**（阶段 2）
- 所有接口方法已定义
- 所有 Request/Response DTO 已定义
- 代码可以编译通过

⏳ **方法实现待完成**（阶段 3）
- 所有方法已创建框架，但实现待后续完善
- 每个方法都有详细的 TODO 注释说明需要实现的功能

---

## 实现复杂度分析

ResidentService 是项目中**最复杂的服务**，需要实现：

### 1. 权限过滤逻辑
- **Resident/Family 登录**：只能查看/修改自己
- **Staff 登录**：根据 `role_permissions` 表检查权限
  - `AssignedOnly=true`：只返回分配给该用户的住户（需要查询 `resident_caregivers` 表）
  - `BranchOnly=true`：只返回该分支的住户（需要 JOIN `units` 表检查 `branch_tag`）

### 2. JOIN 查询
- `ListResidents` 需要 JOIN `units`, `rooms`, `beds` 表获取完整信息
- `GetResident` 需要 JOIN 相同表，并可选 JOIN `resident_phi`, `resident_contacts` 表

### 3. 业务规则验证
- `resident_account` 必填（每家机构有自己的编码模式）
- `discharge_date` 仅在 `status='discharged'` 或 `'transferred'` 时可以有值
- `unit_id` 验证（包括 branch_tag 匹配）
- Hash 唯一性检查（phone_hash, email_hash）

### 4. 多表操作
- **CreateResident**：
  - 创建 `residents` 记录
  - 可选创建 `resident_phi` 记录
  - 可选创建 `resident_contacts` 记录
- **UpdateResident**：
  - 更新 `residents` 记录
  - 可选更新 `resident_phi` 记录
  - 可选更新 `resident_caregivers` 记录

### 5. 数据转换
- `map[string]any` → `domain.Resident` → `ResidentDTO`
- 时间格式转换：`time.Time` ↔ Unix timestamp (int64)
- Hash 格式转换：hex string ↔ `[]byte`

---

## 待实现方法清单

### 1. ListResidents
**复杂度**：⭐⭐⭐⭐⭐
- [ ] 权限过滤逻辑（Resident/Family vs Staff）
- [ ] JOIN units, rooms, beds 表
- [ ] 搜索功能（nickname, unit_name）
- [ ] 过滤功能（status, service_level）
- [ ] 分页处理
- [ ] 数据转换（domain → DTO）

**参考**：`internal/http/admin_residents_handlers.go:213-430`

### 2. GetResident
**复杂度**：⭐⭐⭐⭐
- [ ] 支持通过 resident_id 或 contact_id 查询
- [ ] 权限检查（Resident/Family vs Staff）
- [ ] JOIN units, rooms, beds 表
- [ ] 可选查询 PHI 数据
- [ ] 可选查询联系人数据
- [ ] 数据转换（domain → DTO）

**参考**：`internal/http/admin_residents_handlers.go:2071-2600`

### 3. CreateResident
**复杂度**：⭐⭐⭐⭐⭐
- [ ] 业务规则验证（resident_account, discharge_date, unit_id）
- [ ] 权限检查（C 权限，BranchOnly 检查）
- [ ] Hash 计算（account_hash, password_hash, phone_hash, email_hash）
- [ ] 创建 Resident 记录
- [ ] 可选创建 PHI 记录
- [ ] 可选创建联系人记录
- [ ] Hash 唯一性检查

**参考**：`internal/http/admin_residents_handlers.go:432-846`

### 4. UpdateResident
**复杂度**：⭐⭐⭐⭐⭐
- [ ] 权限检查（Resident vs Staff）
- [ ] 业务规则验证（discharge_date）
- [ ] 部分更新逻辑（只更新提供的字段）
- [ ] 更新 Resident 记录
- [ ] 可选更新 PHI 记录
- [ ] 可选更新 Caregivers 记录

**参考**：`internal/http/admin_residents_handlers.go:2602-2890`

### 5. DeleteResident
**复杂度**：⭐⭐⭐
- [ ] 权限检查（D 权限，Resident/Family 不能删除）
- [ ] 软删除（将 status 设置为 'discharged'）

**参考**：`internal/http/admin_residents_handlers.go:2894-3024`

### 6. ResetResidentPassword
**复杂度**：⭐⭐⭐
- [ ] 权限检查（Resident 只能重置自己的密码，Staff 需要 U 权限）
- [ ] 生成新密码（如果未提供）
- [ ] 计算 password_hash
- [ ] 更新 residents 表

**参考**：`internal/http/admin_residents_handlers.go:856-1050`

### 7. ResetContactPassword
**复杂度**：⭐⭐⭐
- [ ] 权限检查（Contact vs Resident vs Staff）
- [ ] 生成新密码（如果未提供）
- [ ] 计算 password_hash
- [ ] 更新 resident_contacts 表

**参考**：`internal/http/admin_residents_handlers.go:14-212`

---

## 实现建议

### 优先级
1. **高优先级**：`ListResidents`, `GetResident`, `CreateResident`（核心 CRUD 操作）
2. **中优先级**：`UpdateResident`, `DeleteResident`
3. **低优先级**：`ResetResidentPassword`, `ResetContactPassword`

### 实现步骤
1. **第一步**：实现 `ListResidents`（包含权限过滤和 JOIN 查询）
2. **第二步**：实现 `GetResident`（复用权限检查逻辑）
3. **第三步**：实现 `CreateResident`（包含业务规则验证和多表操作）
4. **第四步**：实现 `UpdateResident`（复用业务规则验证逻辑）
5. **第五步**：实现 `DeleteResident`（相对简单）
6. **第六步**：实现密码重置方法（相对独立）

### 注意事项
- 权限检查逻辑复杂，建议提取为辅助函数
- JOIN 查询需要在 Repository 层扩展或使用原生 SQL
- 业务规则验证建议提取为独立函数
- 数据转换逻辑建议提取为辅助函数
- 多表操作需要使用事务确保数据一致性

---

## 相关文件

- **接口定义**：`internal/service/resident_service.go`
- **Repository 接口**：`internal/repository/residents_repo.go`
- **Repository 实现**：`internal/repository/postgres_residents.go`
- **旧 Handler 实现**：`internal/http/admin_residents_handlers.go` (3032 行)
- **设计文档**：`internal/service/RESIDENT_SERVICE_DESIGN.md`

---

## 更新日期

- 2024-XX-XX：接口定义完成
- 待更新：方法实现完成日期

