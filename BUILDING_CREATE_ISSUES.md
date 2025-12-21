# Building 创建问题分析

## 问题描述

1. **验证问题**：后端需要添加验证，要求 `branch_tag` 或 `building_name` 必须有一个不为空
2. **显示问题**：创建 Building (Branch:null, Building:A) 成功，但页面没有出现 `-A`

## 问题分析

### 1. PostgresUnitsRepo 缺少 CreateBuilding 方法

**位置**: `owlBack/wisefido-data/internal/repository/postgres_units.go`

**问题**:
- `PostgresUnitsRepo` 没有实现 `CreateBuilding` 方法
- 只有 `MemoryUnitsRepo` 实现了 `CreateBuilding`
- 当使用 Postgres 数据库时，`buildingWriter` interface 检查会失败，导致创建失败

**代码位置**: `owlBack/wisefido-data/internal/http/admin_units_devices_handlers.go:81-97`
```go
type buildingWriter interface {
    CreateBuilding(ctx context.Context, tenantID string, payload map[string]any) (map[string]any, error)
    UpdateBuilding(ctx context.Context, tenantID, buildingID string, payload map[string]any) (map[string]any, error)
    DeleteBuilding(ctx context.Context, tenantID, buildingID string) error
}

// ...
bw, ok := a.Units.(buildingWriter)
if !ok {
    a.Stub.AdminUnits(w, r)  // 如果 PostgresUnitsRepo 没有实现，会走 stub
    return
}
```

### 2. Building 是虚拟概念

**关键点**:
- Building 不是数据库中的实体表
- `ListBuildings` 从 `units` 表查询，按 `(branch_tag, building)` 分组
- **只有当有 units 时，building 才会出现在列表中**

**代码位置**: `owlBack/wisefido-data/internal/repository/postgres_units.go:18-67`
```go
// ListBuildings: owlFront 需要 buildings 列表，但 owlRD 暂无 buildings 表
// 这里用 units 表做"虚拟 buildings"：按 (branch_tag, building) 分组
func (r *PostgresUnitsRepo) ListBuildings(ctx context.Context, tenantID string, branchTag string) ([]map[string]any, error) {
    q := `
        SELECT
            COALESCE(branch_tag,'-') as branch_tag,
            COALESCE(building,'-') as building,
            MAX((NULLIF(REGEXP_REPLACE(floor, '[^0-9]', '', 'g'), '')::int)) as max_floor
        FROM units
        WHERE ` + where + `
        GROUP BY COALESCE(branch_tag,'-'), COALESCE(building,'-')
    `
}
```

### 3. 后端验证缺失

**位置**: `owlBack/wisefido-data/internal/repository/memory_units.go:42-79`

**问题**:
- `CreateBuilding` 方法没有验证 `branch_tag` 或 `building_name` 必须有一个不为空
- 当前逻辑：如果为空，自动设置为 `'-'`

### 4. 显示问题分析

**前端显示逻辑**: `owlFront/src/views/units/UnitList.vue:1317-1360`

```typescript
const buildingsWithDisplayName = computed(() => {
  const buildingList = buildings.value.map((building) => {
    const tagName = building.branch_tag || '-'  // null 会变成 '-'
    const buildingName = building.building_name || '-'
    const displayName = `${tagName}-${buildingName}`  // 应该是 '-A'
    return { ...building, displayName }
  })
})
```

**可能的问题**:
- 如果 `branch_tag` 是 `null`（不是空字符串），`building.branch_tag || '-'` 应该能正确处理
- 但如果后端返回的 `branch_tag` 是空字符串 `''`，也会显示为 `-A`
- 如果创建成功但没有显示，可能是因为 `ListBuildings` 查询不到（因为没有对应的 units）

## 解决方案

### 方案 1: 实现 PostgresUnitsRepo.CreateBuilding（推荐）

创建一个占位 unit 来代表 building，这样 `ListBuildings` 就能查询到：

```go
func (r *PostgresUnitsRepo) CreateBuilding(ctx context.Context, tenantID string, payload map[string]any) (map[string]any, error) {
    // 验证：branch_tag 或 building_name 必须有一个不为空
    branchTag, _ := payload["branch_tag"].(string)
    buildingName, _ := payload["building_name"].(string)
    
    if (branchTag == "" || branchTag == "-") && (buildingName == "" || buildingName == "-") {
        return nil, fmt.Errorf("branch_tag or building_name must be provided (at least one must not be empty)")
    }
    
    // 设置默认值
    if branchTag == "" {
        branchTag = "-"
    }
    if buildingName == "" {
        buildingName = "-"
    }
    
    floors := 1
    if v, ok := payload["floors"].(float64); ok && int(v) > 0 {
        floors = int(v)
    }
    if v, ok := payload["floors"].(int); ok && v > 0 {
        floors = v
    }
    
    // 创建一个占位 unit 来代表 building
    // unit_name 使用特殊格式：__BUILDING__<building_name>
    // 这样 ListBuildings 就能查询到这个 building
    placeholderUnitName := fmt.Sprintf("__BUILDING__%s", buildingName)
    
    q := `
        INSERT INTO units (tenant_id, branch_tag, unit_name, building, floor, unit_number, unit_type, timezone)
        VALUES ($1, $2, $3, $4, '1F', $3, 'Facility', 'America/Denver')
        ON CONFLICT (tenant_id, branch_tag, unit_name) DO NOTHING
        RETURNING unit_id::text
    `
    var unitID string
    err := r.db.QueryRowContext(ctx, q, tenantID, branchTag, placeholderUnitName, buildingName).Scan(&unitID)
    if err != nil {
        if err == sql.ErrNoRows {
            // 已存在，查询现有的
            err = r.db.QueryRowContext(ctx, 
                `SELECT unit_id::text FROM units WHERE tenant_id = $1 AND branch_tag = $2 AND unit_name = $3`,
                tenantID, branchTag, placeholderUnitName,
            ).Scan(&unitID)
            if err != nil {
                return nil, err
            }
        } else {
            return nil, err
        }
    }
    
    // 返回 building 信息（与 ListBuildings 格式一致）
    buildingID := fmt.Sprintf("%s-%s", branchTag, buildingName)
    return map[string]any{
        "building_id":   buildingID,
        "building_name": buildingName,
        "floors":        floors,
        "tenant_id":     tenantID,
        "branch_tag":    branchTag,
    }, nil
}
```

### 方案 2: 添加验证到 MemoryUnitsRepo

在 `MemoryUnitsRepo.CreateBuilding` 中添加验证：

```go
func (r *MemoryUnitsRepo) CreateBuilding(_ context.Context, tenantID string, payload map[string]any) (map[string]any, error) {
    // 验证：branch_tag 或 building_name 必须有一个不为空
    branchTag, _ := payload["branch_tag"].(string)
    buildingName, _ := payload["building_name"].(string)
    
    if (branchTag == "" || branchTag == "-") && (buildingName == "" || buildingName == "-") {
        return nil, fmt.Errorf("branch_tag or building_name must be provided (at least one must not be empty)")
    }
    
    // ... 其余逻辑
}
```

## 修复步骤

1. ✅ 在 `MemoryUnitsRepo.CreateBuilding` 中添加验证
2. ✅ 在 `PostgresUnitsRepo` 中实现 `CreateBuilding` 方法
3. ✅ 确保 `UpdateBuilding` 和 `DeleteBuilding` 也实现（如果需要）

