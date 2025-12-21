# Units 表唯一性约束分析

## 当前设计

### 当前唯一性约束
- `(tenant_id, branch_tag, unit_name)` - 当 branch_tag IS NOT NULL
- `(tenant_id, unit_name)` - 当 branch_tag IS NULL

### 当前注释说明
- unit_name 是"人类易记的名称"，如 "E203"、"201"、"Home-001"
- unit_name 可能已编码或区分了 building + floor + unit_number 信息
- 但唯一性只依赖 branch_tag + unit_name

## 问题场景分析

### 场景 1：同一 branch_tag，不同 building，相同 unit_name
- branch_tag="A", building="Building A", unit_name="201"
- branch_tag="A", building="Building B", unit_name="201"
- **当前约束**：不允许（违反 branch_tag + unit_name 唯一性）
- **业务需求**：应该允许吗？

### 场景 2：同一 building，不同 floor，相同 unit_name
- building="Building A", floor="1F", unit_name="201"
- building="Building A", floor="2F", unit_name="201"
- **当前约束**：不允许（违反 branch_tag + unit_name 唯一性）
- **业务需求**：应该允许吗？（通常允许，因为不同楼层可以有相同房间号）

### 场景 3：同一 building，同一 floor，相同 unit_name
- building="Building A", floor="1F", unit_name="201"
- building="Building A", floor="1F", unit_name="201"
- **业务需求**：不应该允许（同一位置不能有重复）

## 方案对比

### 方案 1：保持现状 (branch_tag + unit_name)
**优点**：
- 简单，unit_name 在同一个 branch_tag 下唯一
- 如果 unit_name 已经包含了 building/floor 信息（如 "BuildingA-1F-201"），则足够

**缺点**：
- 如果 unit_name 只是简单的房间号（如 "201"），则：
  - 同一 building 的不同楼层不能有相同的 unit_name
  - 同一 branch_tag 的不同 building 不能有相同的 unit_name

**适用场景**：
- unit_name 必须包含足够的信息来区分位置（如 "1F-201"、"BuildingA-201"）

### 方案 2：改为 (branch_tag + building + unit_name)
**优点**：
- 允许同一 building 的不同楼层有相同的 unit_name（如 1F-201 和 2F-201）
- 允许同一 branch_tag 的不同 building 有相同的 unit_name

**缺点**：
- 同一 building 的同一楼层仍然可以有相同的 unit_name（需要 floor 来区分）

**适用场景**：
- unit_name 是简单的房间号（如 "201"），不包含 floor 信息
- 不同楼层可以有相同的房间号

### 方案 3：改为 (branch_tag + building + floor + unit_name)
**优点**：
- 最严格，完全避免重复
- 允许同一 building 的不同楼层有相同的 unit_name
- 允许同一 branch_tag 的不同 building 有相同的 unit_name

**缺点**：
- 如果 unit_name 已经包含了 floor 信息（如 "1F-201"），会冗余
- 约束更复杂

**适用场景**：
- unit_name 是简单的房间号（如 "201"），不包含 floor 信息
- 需要最严格的唯一性保证

### 方案 4：改为 (branch_tag + building + floor + unit_number)
**优点**：
- 使用 unit_number（准确的房间号）而不是 unit_name（易记名称）
- 最符合业务逻辑：同一 building 的同一楼层，unit_number 应该唯一

**缺点**：
- 需要修改约束，使用 unit_number 而不是 unit_name
- unit_name 可以重复（但 unit_number 不能）

**适用场景**：
- unit_number 是准确的房间号（如 "201"、"E203"）
- unit_name 是易记名称，可以重复

## 推荐方案

### 推荐：方案 3 - (branch_tag + building + floor + unit_name)

**理由**：
1. **业务逻辑**：在实际场景中，不同楼层通常可以有相同的房间号（如 1F-201 和 2F-201）
2. **灵活性**：允许 unit_name 是简单的房间号（如 "201"），不需要编码 building/floor 信息
3. **唯一性保证**：同一 building 的同一楼层，unit_name 必须唯一
4. **与 Building 表的关系**：Building 现在是独立实体，unit 应该通过 building + floor 来定位

**注意事项**：
- 如果 unit_name 已经包含了 floor 信息（如 "1F-201"），仍然可以工作（只是 floor 字段会冗余）
- 如果 unit_name 是简单的房间号（如 "201"），则 floor 字段是必需的

### 备选：方案 2 - (branch_tag + building + unit_name)

**如果业务规则是**：
- 同一 building 的不同楼层**不能**有相同的 unit_name
- 但同一 branch_tag 的不同 building **可以**有相同的 unit_name

## 实施建议

1. **先确认业务规则**：
   - 同一 building 的不同楼层，unit_name 可以相同吗？
   - 同一 branch_tag 的不同 building，unit_name 可以相同吗？

2. **如果采用方案 3**：
   - 修改唯一性约束为 `(tenant_id, branch_tag, building, floor, unit_name)`
   - 更新相关注释和文档
   - 考虑数据迁移（如果有现有数据）

3. **如果保持方案 1**：
   - 需要确保 unit_name 包含足够信息（如 "1F-201"、"BuildingA-201"）
   - 或者限制：同一 branch_tag 下，unit_name 必须唯一

