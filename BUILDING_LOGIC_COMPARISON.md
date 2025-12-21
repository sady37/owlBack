# Building 逻辑对比分析

## 问题：为什么 unit 中没有数据？

### 1. Mock 实现逻辑

**`mockCreateBuilding`** (`owlFront/test/admin/unit/mock.ts`):
```typescript
export function mockCreateBuilding(params: CreateBuildingParams & { tag_name?: string }): Building {
  // ...
  const newBuilding: Building = {
    building_id: `building-${buildingIdCounter++}`,
    building_name: finalBuildingName,
    floors: params.floors,
    tenant_id: 'tenant-1',
    branch_tag: finalBranchTag,
  }
  buildings.push(newBuilding)  // ✅ 直接添加到独立的 buildings 数组
  return newBuilding
}
```

**`mockGetBuildings`**:
```typescript
export function mockGetBuildings(): Building[] {
  return [...buildings]  // ✅ 直接返回独立的 buildings 数组
}
```

**结论**：
- ✅ **Buildings 是独立存储的**，不依赖 units
- ✅ 创建 building 时，**不需要 units 存在**
- ✅ 查询 buildings 时，**直接从 buildings 数组返回**

---

### 2. MemoryUnitsRepo 实现逻辑

**`ListBuildings`** (`owlBack/wisefido-data/internal/repository/memory_units.go`):
```go
func (r *MemoryUnitsRepo) ListBuildings(_ context.Context, tenantID string, branchTag string) ([]map[string]any, error) {
    // ...
    for _, b := range r.buildings[tenantID] {  // ✅ 从独立的 buildings map 返回
        // ...
    }
    return out, nil
}
```

**`CreateBuilding`**:
```go
func (r *MemoryUnitsRepo) CreateBuilding(_ context.Context, tenantID string, payload map[string]any) (map[string]any, error) {
    // ...
    r.buildings[tenantID][id] = b  // ✅ 直接存储到独立的 buildings map
    return b, nil
}
```

**结论**：
- ✅ **Buildings 是独立存储的**，存储在 `r.buildings[tenantID]` map 中
- ✅ 创建 building 时，**不需要 units 存在**
- ✅ 查询 buildings 时，**直接从 buildings map 返回**

---

### 3. PostgresUnitsRepo 实现逻辑（原来的）

**`ListBuildings`** (`owlBack/wisefido-data/internal/repository/postgres_units.go`):
```go
func (r *PostgresUnitsRepo) ListBuildings(ctx context.Context, tenantID string, branchTag string) ([]map[string]any, error) {
    q := `
        SELECT
            COALESCE(branch_tag,'-') as branch_tag,
            COALESCE(building,'-') as building,
            MAX((NULLIF(REGEXP_REPLACE(floor, '[^0-9]', '', 'g'), '')::int)) as max_floor
        FROM units  // ❌ 从 units 表查询
        WHERE ` + where + `
        GROUP BY COALESCE(branch_tag,'-'), COALESCE(building,'-')
    `
    // ...
}
```

**原来的 `CreateBuilding`**：
- ❌ **不存在**！PostgresUnitsRepo 原来没有 `CreateBuilding` 方法

**结论**：
- ❌ **Buildings 是虚拟的**，从 units 表分组得到
- ❌ 如果没有 units，**就查不到 buildings**
- ❌ 这是设计上的问题：**Buildings 应该独立存储，不应该依赖 units**

---

### 4. 我的修改（错误的）

为了解决 Postgres 实现的问题，我创建了占位 unit：

```go
func (r *PostgresUnitsRepo) CreateBuilding(ctx context.Context, tenantID string, payload map[string]any) (map[string]any, error) {
    // ...
    // 创建一个占位 unit 来代表 building
    placeholderUnitName := fmt.Sprintf("__BUILDING__%s__%s", buildingName, branchTag)
    
    q := `
        INSERT INTO units (tenant_id, branch_tag, unit_name, building, floor, unit_number, unit_type, timezone)
        VALUES ($1, $2, $3, $4, '1F', $3, 'Facility', 'America/Denver')
    `
    // ...
}
```

**问题**：
- ❌ 这是**临时解决方案**，不是正确的设计
- ❌ 创建了不应该存在的占位 unit
- ❌ 污染了 units 表

---

## 正确的解决方案

### 方案 1: 创建独立的 buildings 表（推荐）

```sql
CREATE TABLE buildings (
    building_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    branch_tag VARCHAR(255),
    building_name VARCHAR(50) NOT NULL,
    floors INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, branch_tag, building_name)
);
```

然后：
- `CreateBuilding`: 插入到 `buildings` 表
- `ListBuildings`: 从 `buildings` 表查询
- 不再依赖 units 表

### 方案 2: 保持虚拟 buildings，但允许空 building

修改 `ListBuildings` 逻辑：
- 如果用户创建了 building 但没有 units，仍然可以查询到
- 但这需要额外的存储机制（如独立的 buildings 表或配置表）

---

## 当前状态

### 数据库中的实际情况

```
Buildings (从 units 分组):
- branch_tag: '-', building: 'A', unit_count: 1

Units:
- 1 个占位 unit: __BUILDING__A__-
- 0 个真实 units
```

**问题**：
- 用户创建了 Building (Branch: null, Building: A)
- 我创建了占位 unit 来让 `ListBuildings` 能查询到这个 building
- 但这是**错误的设计**，不应该依赖占位 unit

---

## 建议

1. **短期**：保持当前实现（占位 unit），但需要：
   - 确保占位 unit 不会在 UI 中显示（已实现）
   - 确保占位 unit 不会影响业务逻辑

2. **长期**：创建独立的 `buildings` 表，实现正确的设计

3. **立即修复**：如果用户要求，可以：
   - 删除占位 unit 的实现
   - 创建独立的 buildings 表
   - 修改 `CreateBuilding` 和 `ListBuildings` 使用新表

