# Building 和 Unit 唯一性约束改进 - 下一步计划

## ✅ 已完成的工作

### 1. 数据库设计
- ✅ 创建了 `buildings` 表 (`owlRD/db/04.5_buildings.sql`)
- ✅ 修改了 `units` 表的唯一性约束：从 `(branch_tag + unit_name)` 改为 `(branch_tag + building + floor + unit_name)`
- ✅ 创建了迁移脚本 (`owlRD/db/migration_update_units_uniqueness.sql`)

### 2. 后端实现
- ✅ `CreateBuilding`: 直接插入到 `buildings` 表（不再使用占位 unit）
- ✅ `UpdateBuilding`: 直接更新 `buildings` 表记录
- ✅ `DeleteBuilding`: 直接删除 `buildings` 表记录
- ✅ `ListBuildings`: 优先从 `buildings` 表查询
- ✅ `GetBuilding`: 优先从 `buildings` 表获取
- ✅ `validateUnitFloor`: 验证 unit.floor 是否在 building.floors 范围内

### 3. 前端实现
- ✅ Create Unit 表单：自动使用 `selectedBuilding` 和 `selectedFloor`（不再需要手动输入）
- ✅ Floor 下拉选择：从 `selectedBuilding.floors` 生成选项

### 4. 文档更新
- ✅ 更新了 `owlRD/db/05_units.sql` 的注释
- ✅ 更新了 `owlRD/db/22_tags_catalog.sql` 的注释
- ✅ 更新了验证检查清单和报告

## 📋 下一步需要完成的工作

### 1. **执行数据库迁移**（重要！）

#### 步骤 1：检查现有数据是否有重复
在数据库中执行以下查询，检查是否有违反新唯一性约束的数据：

```sql
-- 检查 branch_tag IS NOT NULL 的情况
SELECT tenant_id, branch_tag, building, floor, unit_name, COUNT(*) as cnt
FROM units
WHERE branch_tag IS NOT NULL
  AND unit_name NOT LIKE '__BUILDING__%'  -- 排除占位 unit
GROUP BY tenant_id, branch_tag, building, floor, unit_name
HAVING COUNT(*) > 1;

-- 检查 branch_tag IS NULL 的情况
SELECT tenant_id, building, floor, unit_name, COUNT(*) as cnt
FROM units
WHERE branch_tag IS NULL
  AND unit_name NOT LIKE '__BUILDING__%'  -- 排除占位 unit
GROUP BY tenant_id, building, floor, unit_name
HAVING COUNT(*) > 1;
```

**如果有重复数据**：
- 需要先清理重复数据（删除或重命名）
- 或者修改数据使其符合新的唯一性约束

#### 步骤 2：执行迁移脚本
```bash
# 在数据库中执行
psql -d your_database -f owlRD/db/migration_update_units_uniqueness.sql
```

或者手动执行：
```sql
-- 1. 删除旧的唯一性索引
DROP INDEX IF EXISTS idx_units_unique_with_tag;
DROP INDEX IF EXISTS idx_units_unique_without_tag;

-- 2. 创建新的唯一性索引
CREATE UNIQUE INDEX idx_units_unique_with_tag 
    ON units(tenant_id, branch_tag, building, floor, unit_name) 
    WHERE branch_tag IS NOT NULL;

CREATE UNIQUE INDEX idx_units_unique_without_tag 
    ON units(tenant_id, building, floor, unit_name) 
    WHERE branch_tag IS NULL;
```

### 2. **清理占位 unit 数据**（可选）

如果之前有使用占位 unit 创建的 building，现在可以清理这些占位 unit：

```sql
-- 删除所有占位 unit（unit_name 以 __BUILDING__ 开头）
DELETE FROM units 
WHERE unit_name LIKE '__BUILDING__%';
```

**注意**：删除前请确认这些占位 unit 没有关联的数据（如 rooms、beds、devices 等）

### 3. **测试验证**

#### 测试场景 1：创建 Building
- [ ] 创建 Building（Branch, Building, Floors）
- [ ] 验证 buildings 表中有记录
- [ ] 验证 ListBuildings 能正确显示

#### 测试场景 2：创建 Unit（新唯一性约束）
- [ ] 同一 building，不同 floor，相同 unit_name（应该允许）
  - Building A, 1F, unit_name="201" ✅
  - Building A, 2F, unit_name="201" ✅
- [ ] 同一 building，同一 floor，相同 unit_name（应该不允许）
  - Building A, 1F, unit_name="201" ✅
  - Building A, 1F, unit_name="201" ❌（应该报错）
- [ ] 不同 building，相同 unit_name（应该允许）
  - Building A, 1F, unit_name="201" ✅
  - Building B, 1F, unit_name="201" ✅

#### 测试场景 3：Floor 验证
- [ ] 创建 unit 时，floor 超出 building.floors 范围（应该报错）
  - Building A (floors=3), floor="4F" ❌（应该报错）

#### 测试场景 4：Edit Building
- [ ] 编辑 Building 的 building_name、branch_tag、floors
- [ ] 验证 buildings 表记录已更新

#### 测试场景 5：Delete Building
- [ ] 删除 Building
- [ ] 验证 buildings 表记录已删除
- [ ] 验证相关 units 仍然存在（只是不再被 building 分组）

### 4. **前端测试**

#### 测试场景 1：Create Unit 流程
- [ ] 选择 Building 和 Floor
- [ ] 打开 Create Unit 表单
- [ ] 验证 Branch、Building、Floor 自动填充（只读）
- [ ] 填写 Unit Number、Unit Name
- [ ] 提交创建
- [ ] 验证创建成功

#### 测试场景 2：Floor 下拉选择
- [ ] 选择 Building（floors=3）
- [ ] 打开 Create Unit 表单
- [ ] 验证 Floor 下拉显示 1F, 2F, 3F
- [ ] 选择 Floor
- [ ] 提交创建

### 5. **数据迁移检查清单**

- [ ] 检查现有数据是否有重复（执行验证查询）
- [ ] 如果有重复，清理或修改数据
- [ ] 执行迁移脚本
- [ ] 验证新索引创建成功
- [ ] 测试创建 unit（验证唯一性约束工作正常）
- [ ] 清理占位 unit（如果存在）

## 🎯 优先级

1. **高优先级**：执行数据库迁移（步骤 1-2）
2. **中优先级**：测试验证（步骤 3-4）
3. **低优先级**：清理占位 unit（步骤 2，可选）

## 📝 注意事项

1. **数据备份**：执行迁移前，建议备份数据库
2. **重复数据**：如果发现重复数据，需要先处理再执行迁移
3. **占位 unit**：清理占位 unit 前，确认没有关联数据
4. **向后兼容**：ListBuildings 仍然支持从 units 表虚拟获取（向后兼容）

