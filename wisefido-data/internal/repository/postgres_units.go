package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

type PostgresUnitsRepository struct {
	db *sql.DB
}

func NewPostgresUnitsRepository(db *sql.DB) *PostgresUnitsRepository {
	return &PostgresUnitsRepository{db: db}
}

// ============================================
// Building 操作
// ============================================

// ListBuildings: 从 buildings 表查询（Building 已改为实体，不再从 units 表虚拟获取）
// 替代触发器：无（仅查询）
// 参数说明：
//   - branchID: 如果 Valid=true，直接使用 branch_id 过滤；如果 Valid=false，忽略此参数
//   - branchName: 如果 branchID 无效且 branchName 不为空，通过 JOIN branches 表使用 branch_name 过滤（向后兼容）；如果为空，查询所有
func (r *PostgresUnitsRepository) ListBuildings(ctx context.Context, tenantID string, branchID sql.NullString, branchName string) ([]*domain.Building, error) {
	if tenantID == "" {
		return []*domain.Building{}, nil
	}

	where := "b.tenant_id = $1"
	args := []any{tenantID}
	argIdx := 2

	// 优先使用 branch_id 过滤（如果提供）
	if branchID.Valid && branchID.String != "" {
		where += " AND b.branch_id = $" + fmt.Sprintf("%d", argIdx)
		args = append(args, branchID.String)
		argIdx++
	} else if branchName != "" {
		// 向后兼容：如果提供了 branchName，通过 JOIN branches 表查找
		where += " AND COALESCE(br.branch_name, 'default') = $" + fmt.Sprintf("%d", argIdx)
		args = append(args, branchName)
		argIdx++
	}
	// 如果都没有提供，查询整个 tenant 的所有 buildings

	q := `
		SELECT
			b.building_id::text,
			b.tenant_id::text,
			b.branch_id::text,
			b.building_name,
			br.branch_name,
			b.created_at,
			b.updated_at
		FROM buildings b
		LEFT JOIN branches br ON br.branch_id = b.branch_id
		WHERE ` + where + `
		  AND NOT (b.branch_id IS NULL AND b.building_name = '-')
		ORDER BY COALESCE(br.branch_name, 'default'), b.building_name
	`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.Building{}
	for rows.Next() {
		var b domain.Building
		var branchID, branchName sql.NullString
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&b.BuildingID, &b.TenantID, &branchID, &b.BuildingName, &branchName, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		b.BranchID = branchID
		b.BranchName = branchName
		b.CreatedAt = createdAt
		b.UpdatedAt = updatedAt
		// Additional check: filter out buildings where both branch_id is NULL and building_name is '-'
		if !branchID.Valid && b.BuildingName == "-" {
			continue
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}

// GetBuilding: 从 buildings 表获取 building 信息（不包含 created_at/updated_at，返回 branch_name）
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) GetBuilding(ctx context.Context, tenantID, buildingID string) (*domain.Building, error) {
	if tenantID == "" || buildingID == "" {
		return nil, fmt.Errorf("tenant_id and building_id are required")
	}

	q := `
		SELECT
			b.building_id::text,
			b.tenant_id::text,
			b.branch_id::text,
			b.building_name,
			br.branch_name
		FROM buildings b
		LEFT JOIN branches br ON br.branch_id = b.branch_id
		WHERE b.tenant_id = $1 AND b.building_id = $2
	`
	var b domain.Building
	var branchID, branchName sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, buildingID).Scan(
		&b.BuildingID,
		&b.TenantID,
		&branchID,
		&b.BuildingName,
		&branchName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("building not found: building_id=%s", buildingID)
		}
		return nil, err
	}
	b.BranchID = branchID
	b.BranchName = branchName
	return &b, nil
}

// GetBuildingUnits: 获取指定 building 下的所有 units（只返回 unit_id, unit_name, floor）
func (r *PostgresUnitsRepository) GetBuildingUnits(ctx context.Context, tenantID, buildingID string) ([]BuildingUnitInfo, error) {
	if tenantID == "" || buildingID == "" {
		return nil, fmt.Errorf("tenant_id and building_id are required")
	}

	q := `
		SELECT
			u.unit_id::text,
			u.unit_name,
			u.floor
		FROM units u
		WHERE u.tenant_id = $1 AND u.building_id = $2
		ORDER BY u.floor, u.unit_name
	`
	rows, err := r.db.QueryContext(ctx, q, tenantID, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query building units: %w", err)
	}
	defer rows.Close()

	var units []BuildingUnitInfo
	for rows.Next() {
		var unit BuildingUnitInfo
		if err := rows.Scan(&unit.UnitID, &unit.UnitName, &unit.Floor); err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate units: %w", err)
	}
	return units, nil
}

// FindBuildingByBranchAndName: 通过 branch_id + building_name（小写比较）查找 building
func (r *PostgresUnitsRepository) FindBuildingByBranchAndName(ctx context.Context, tenantID, branchID, buildingName string) (string, error) {
	if tenantID == "" || branchID == "" || buildingName == "" {
		return "", fmt.Errorf("tenant_id, branch_id, and building_name are required")
	}

	// 使用 LOWER 函数进行大小写不敏感的比较
	q := `
		SELECT building_id::text
		FROM buildings
		WHERE tenant_id = $1 
		  AND branch_id = $2 
		  AND LOWER(building_name) = LOWER($3)
		LIMIT 1
	`
	var buildingID string
	err := r.db.QueryRowContext(ctx, q, tenantID, branchID, buildingName).Scan(&buildingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("building not found")
		}
		return "", fmt.Errorf("failed to find building: %w", err)
	}
	return buildingID, nil
}

// CreateBuilding: 直接在 buildings 表中创建记录
// 替代触发器：无（仅插入）
func (r *PostgresUnitsRepository) CreateBuilding(ctx context.Context, tenantID string, building *domain.Building) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if building == nil {
		return "", fmt.Errorf("building is required")
	}

	// 处理 branch_id：必须提供，不能为 NULL（业务规则：一家机构必然有一个分支或总部）
	// 注意：building_name 的空值处理应该在 Service 层完成，Repository 层不做业务逻辑处理
	if !building.BranchID.Valid || building.BranchID.String == "" {
		return "", fmt.Errorf("branch_id is required and cannot be NULL")
	}
	branchIDArg := building.BranchID.String

	// branch_id 现在为 NOT NULL，使用唯一约束 (tenant_id, branch_id, building_name)
	var buildingID string
	var insertedBranchID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO buildings (tenant_id, branch_id, building_name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (tenant_id, branch_id, building_name)
		 DO UPDATE SET building_name = EXCLUDED.building_name, updated_at = CURRENT_TIMESTAMP
		 RETURNING building_id::text, branch_id::text`,
		tenantID, branchIDArg, building.BuildingName,
	).Scan(&buildingID, &insertedBranchID)
	if err != nil {
		return "", fmt.Errorf("failed to create building: %w", err)
	}

	// 注意：branch_tag 不应该在这里创建
	// branch_tag 应该由前端在 TagList 页面创建（tag_name = "Branch"）
	// unit 的 branch_name 只是数据，不需要同步到 tags_catalog

	return buildingID, nil
}

// UpdateBuilding: 直接更新 buildings 表的记录
// 替代触发器：trigger_sync_branch_tag（同步branch_tag到tags_catalog）
func (r *PostgresUnitsRepository) UpdateBuilding(ctx context.Context, tenantID, buildingID string, building *domain.Building) error {
	if tenantID == "" || buildingID == "" {
		return fmt.Errorf("tenant_id and building_id are required")
	}
	if building == nil {
		return fmt.Errorf("building is required")
	}

	// 先获取现有的 building 记录
	var oldBranchID sql.NullString
	var oldBuildingName string
	err := r.db.QueryRowContext(ctx,
		`SELECT branch_id::text, building_name 
		 FROM buildings 
		 WHERE tenant_id = $1 AND building_id = $2`,
		tenantID, buildingID,
	).Scan(&oldBranchID, &oldBuildingName)

	if err == sql.ErrNoRows {
		return fmt.Errorf("building not found")
	}
	if err != nil {
		return fmt.Errorf("failed to find building: %w", err)
	}

	// 获取新的值
	newBuildingName := building.BuildingName
	if newBuildingName == "" {
		newBuildingName = oldBuildingName
	}

	// 设置默认值
	if newBuildingName == "" {
		newBuildingName = "-"
	}

	// 处理 branch_id：必须提供，不能为 NULL（业务规则：一家机构必然有一个分支或总部）
	if !building.BranchID.Valid || building.BranchID.String == "" {
		return fmt.Errorf("branch_id is required and cannot be NULL")
	}
	branchIDArg := building.BranchID.String

	// 更新 buildings 表（branch_id 现在为 NOT NULL）
	var updatedBranchID sql.NullString
	err = r.db.QueryRowContext(ctx,
		`UPDATE buildings 
		 SET branch_id = $1, building_name = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE tenant_id = $3 AND building_id = $4
		 RETURNING branch_id::text`,
		branchIDArg, newBuildingName, tenantID, buildingID,
	).Scan(&updatedBranchID)

	if err != nil {
		return fmt.Errorf("failed to update building: %w", err)
	}

	return nil
}


// DeleteBuilding: 直接删除 buildings 表的记录
// 替代触发器：无（仅删除）
func (r *PostgresUnitsRepository) DeleteBuilding(ctx context.Context, tenantID, buildingID string) error {
	if tenantID == "" || buildingID == "" {
		return fmt.Errorf("tenant_id and building_id are required")
	}

	_, err := r.db.ExecContext(ctx,
		`DELETE FROM buildings 
		 WHERE tenant_id = $1 AND building_id = $2`,
		tenantID, buildingID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete building: %w", err)
	}

	return nil
}

// ============================================
// Unit 操作
// ============================================

// ListUnits: 查询 units 列表
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) ListUnits(ctx context.Context, tenantID string, filters UnitFilters, page, size int) ([]*domain.Unit, int, error) {
	if tenantID == "" {
		return []*domain.Unit{}, 0, nil
	}

	where := []string{"u.tenant_id = $1", "u.unit_name NOT LIKE '__BUILDING__%'"}
	args := []any{tenantID}
	argN := 2

	addEq := func(col, val string) {
		if val == "" {
			return
		}
		where = append(where, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
		argN++
	}

	// Handle branch_id/branch_name and building_id/building_name together:
	// 优先使用 branch_id 和 building_id，如果没有提供则使用 branch_name 和 building_name（向后兼容）
	// - 当 building_id 不为空时：直接使用 u.building_id = X
	// - 当 building_id 为空但 building 不为空时：通过 JOIN buildings 表匹配 building_name = Y
	// - 当 building_id 和 building 都为空时：不添加 building 过滤条件（查询所有 units，包括 building_id IS NULL 的情况）

	// 处理 branch 过滤
	if len(filters.BranchIDs) > 0 {
		// 分支 1：branch_ids 不为空 → 使用 IN 查询
		placeholders := make([]string, len(filters.BranchIDs))
		for i := range filters.BranchIDs {
			placeholders[i] = fmt.Sprintf("$%d", argN)
			args = append(args, filters.BranchIDs[i])
			argN++
		}
		where = append(where, fmt.Sprintf("u.branch_id IN (%s)", strings.Join(placeholders, ", ")))
	} else if filters.BranchID != "" {
		// 分支 2：branch_id 不为空 → 直接使用 u.branch_id
		where = append(where, fmt.Sprintf("u.branch_id = $%d", argN))
		args = append(args, filters.BranchID)
		argN++
	} else if filters.BranchName == "" {
		// 分支 3：branch_id 和 branch_name 都为空 → 匹配 branch_id IS NULL
		where = append(where, "u.branch_id IS NULL")
	} else {
		// 分支 4：branch_id 为空，branch_name 不为空 → 通过 JOIN branches 表匹配
		where = append(where, fmt.Sprintf("b.branch_name = $%d", argN))
		args = append(args, filters.BranchName)
		argN++
	}

	// 处理 building 过滤：优先使用 building_id，如果没有则使用 building_name（向后兼容）
	if filters.BuildingID != "" {
		// 优先使用 building_id（UUID 类型）
		where = append(where, fmt.Sprintf("u.building_id = $%d", argN))
		args = append(args, filters.BuildingID)
		argN++
	} else if filters.Building != "" {
		// 向后兼容：通过 JOIN buildings 表匹配 building_name
		where = append(where, fmt.Sprintf("bd.building_name = $%d", argN))
		args = append(args, filters.Building)
		argN++
	}
	// 如果 building_id 和 building 都为空：不添加 building 过滤条件（查询所有 units，包括 building_id IS NULL 的情况）
	addEq("u.floor", filters.Floor)
	addEq("u.unit_name", filters.UnitName)
	addEq("u.unit_type", filters.UnitType)

	// Search filter: 模糊搜索 unit_name
	if filters.Search != "" {
		where = append(where, fmt.Sprintf("u.unit_name ILIKE $%d", argN))
		args = append(args, "%"+filters.Search+"%")
		argN++
	}

	// 住户绑定时：VIP 单元仅返回未被其他住户占用的（unit/room/bed 任一被占即视为不可选）
	if filters.ResidentID != nil {
		rid := *filters.ResidentID
		where = append(where, fmt.Sprintf(`(u.is_public = true OR u.is_shared_unit = true OR NOT EXISTS (
			SELECT 1 FROM residents res
			WHERE res.tenant_id = u.tenant_id AND res.resident_id IS DISTINCT FROM $%d
			AND (res.unit_id = u.unit_id
				OR res.room_id IN (SELECT room_id FROM rooms WHERE tenant_id = u.tenant_id AND unit_id = u.unit_id)
				OR res.bed_id IN (SELECT bed_id FROM beds b2 JOIN rooms rm ON rm.room_id = b2.room_id AND rm.tenant_id = u.tenant_id WHERE rm.unit_id = u.unit_id)
			)
		))`, argN))
		args = append(args, rid)
		argN++
	}

	// 构建 FROM 子句：总是 JOIN branches 和 buildings 表以获取 branch_name 和 building_name（domain模型要求）
	fromClause := "FROM units u"
	fromClause += " LEFT JOIN branches b ON b.branch_id = u.branch_id"
	fromClause += " LEFT JOIN buildings bd ON bd.building_id = u.building_id"

	// COUNT查询也需要JOIN branches 和 buildings 表（因为WHERE子句可能使用b.branch_name 或 bd.building_name）
	queryCount := "SELECT COUNT(*) " + fromClause + " WHERE " + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	offset := (page - 1) * size

	argsList := append(args, size, offset)

	query := `
		SELECT 
			u.unit_id::text,
			u.tenant_id::text,
			u.branch_id::text,
			COALESCE(b.branch_name, NULL) as branch_name,
			u.unit_name,
			u.building_id::text,
			COALESCE(bd.building_name, NULL) as building_name,
			u.floor,
			CASE WHEN u.layout_config IS NULL THEN NULL ELSE u.layout_config::text END as layout_config,
			u.unit_type,
			u.is_public,
			u.is_shared_unit,
			u.timezone
		` + fromClause + `
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY 
			-- First sort by floor (extract number from "1F", "2F", etc.)
			COALESCE((NULLIF(REGEXP_REPLACE(u.floor, '[^0-9]', '', 'g'), '')::int), 0),
			-- Then sort by unit_name (try numeric, fallback to string)
			CASE 
				WHEN u.unit_name ~ '^[0-9]+$' THEN u.unit_name::int
				ELSE 999999
			END,
			u.unit_name
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*domain.Unit, 0)
	for rows.Next() {
		var u domain.Unit
		var branchID, branchName, buildingID, buildingName, floor, layoutConfig sql.NullString
		if err := rows.Scan(
			&u.UnitID,
			&u.TenantID,
			&branchID,
			&branchName,
			&u.UnitName,
			&buildingID,
			&buildingName,
			&floor,
			&layoutConfig,
			&u.UnitType,
			&u.IsPublic,
			&u.IsSharedUnit,
			&u.Timezone,
		); err != nil {
			return nil, 0, err
		}
		u.BranchID = branchID
		u.BranchName = branchName
		u.BuildingID = buildingID
		u.BuildingName = buildingName
		u.Floor = floor
		u.LayoutConfig = layoutConfig
		items = append(items, &u)
	}
	return items, total, rows.Err()
}

// GetUnit: 获取单个 unit
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) GetUnit(ctx context.Context, tenantID, unitID string) (*domain.Unit, error) {
	if tenantID == "" || unitID == "" {
		return nil, sql.ErrNoRows
	}
	q := `
		SELECT 
			u.unit_id::text,
			u.tenant_id::text,
			u.branch_id::text,
			COALESCE(b.branch_name, NULL) as branch_name,
			u.unit_name,
			u.building_id::text,
			COALESCE(bd.building_name, NULL) as building_name,
			u.floor,
			CASE WHEN u.layout_config IS NULL THEN NULL ELSE u.layout_config::text END as layout_config,
			u.unit_type,
			u.is_public,
			u.is_shared_unit,
			u.timezone
		FROM units u
		LEFT JOIN branches b ON b.branch_id = u.branch_id
		LEFT JOIN buildings bd ON bd.building_id = u.building_id
		WHERE u.tenant_id = $1 AND u.unit_id = $2
	`
	var u domain.Unit
	var branchID, branchName, buildingID, buildingName, floor, layoutConfig sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, unitID).Scan(
		&u.UnitID,
		&u.TenantID,
		&branchID,
		&branchName,
		&u.UnitName,
		&buildingID,
		&buildingName,
		&floor,
		&layoutConfig,
		&u.UnitType,
		&u.IsPublic,
		&u.IsSharedUnit,
		&u.Timezone,
	)
	if err != nil {
		return nil, err
	}
	u.BranchID = branchID
	u.BranchName = branchName
	u.BuildingID = buildingID
	u.BuildingName = buildingName
	u.Floor = floor
	u.LayoutConfig = layoutConfig
	return &u, nil
}

// CreateUnit: 创建 unit
// 替代触发器：trigger_sync_branch_tag, trigger_sync_area_tag
func (r *PostgresUnitsRepository) CreateUnit(ctx context.Context, tenantID string, unit *domain.Unit) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if unit == nil {
		return "", fmt.Errorf("unit is required")
	}

	// 验证必填字段
	if unit.UnitName == "" {
		return "", fmt.Errorf("unit_name is required")
	}
	if unit.UnitType == "" {
		return "", fmt.Errorf("unit_type is required")
	}
	if unit.Timezone == "" {
		return "", fmt.Errorf("timezone is required")
	}

	// 处理 branch_id：如果提供了 BranchName，需要通过 branches 表查找 branch_id
	var branchIDSQL sql.NullString
	if unit.BranchName.Valid && unit.BranchName.String != "" && unit.BranchName.String != "default" {
		// 通过 branch_name 查找 branch_id
		var branchID string
		err := r.db.QueryRowContext(ctx,
			`SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`,
			tenantID, unit.BranchName.String,
		).Scan(&branchID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("branch not found: branch_name=%s", unit.BranchName.String)
			}
			return "", fmt.Errorf("failed to find branch: %w", err)
		}
		branchIDSQL = sql.NullString{String: branchID, Valid: true}
	} else if unit.BranchID.Valid && unit.BranchID.String != "" {
		// 如果直接提供了 BranchID，使用它
		branchIDSQL = unit.BranchID
	}

	// 处理 building_id：如果提供了 BuildingID，使用它；如果提供了 BuildingName，通过 buildings 表查找 building_id
	var buildingIDSQL sql.NullString
	if unit.BuildingID.Valid && unit.BuildingID.String != "" {
		// 如果直接提供了 BuildingID，使用它
		buildingIDSQL = unit.BuildingID
		// 验证 building_id 是否存在
		var buildingBranchID sql.NullString
		err := r.db.QueryRowContext(ctx,
			`SELECT branch_id::text FROM buildings 
			 WHERE tenant_id = $1 AND building_id = $2`,
			tenantID, unit.BuildingID.String,
		).Scan(&buildingBranchID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("building not found: building_id=%s", unit.BuildingID.String)
			}
			return "", fmt.Errorf("failed to validate building: %w", err)
		}
		// 如果 unit 也指定了 branch_id，验证它们是否一致
		if branchIDSQL.Valid && buildingBranchID.Valid {
			if branchIDSQL.String != buildingBranchID.String {
				return "", fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=%s", branchIDSQL.String, buildingBranchID.String)
			}
		} else if branchIDSQL.Valid {
			// unit 指定了 branch_id，但 building 的 branch_id 为 NULL，不一致
			return "", fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=NULL", branchIDSQL.String)
		} else if buildingBranchID.Valid {
			// building 有 branch_id，但 unit 没有指定，使用 building 的 branch_id
			branchIDSQL = buildingBranchID
		}
	} else if unit.BuildingName.Valid && unit.BuildingName.String != "" && unit.BuildingName.String != "-" {
		// 如果提供了 BuildingName（向后兼容），通过 buildings 表查找 building_id
		var buildingID string
		var buildingBranchID sql.NullString
		err := r.db.QueryRowContext(ctx,
			`SELECT building_id::text, branch_id::text FROM buildings 
			 WHERE tenant_id = $1 AND building_name = $2`,
			tenantID, unit.BuildingName.String,
		).Scan(&buildingID, &buildingBranchID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("building not found: building_name=%s", unit.BuildingName.String)
			}
			return "", fmt.Errorf("failed to validate building: %w", err)
		}
		buildingIDSQL = sql.NullString{String: buildingID, Valid: true}
		// 如果 unit 也指定了 branch_id，验证它们是否一致
		if branchIDSQL.Valid && buildingBranchID.Valid {
			if branchIDSQL.String != buildingBranchID.String {
				return "", fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=%s", branchIDSQL.String, buildingBranchID.String)
			}
		} else if branchIDSQL.Valid {
			// unit 指定了 branch_id，但 building 的 branch_id 为 NULL，不一致
			return "", fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=NULL", branchIDSQL.String)
		} else if buildingBranchID.Valid {
			// building 有 branch_id，但 unit 没有指定，使用 building 的 branch_id
			branchIDSQL = buildingBranchID
		}
	}

	// 验证：branch_id 必须提供，不能为 NULL（业务规则：一家机构必然有一个分支或总部）
	if !branchIDSQL.Valid {
		return "", fmt.Errorf("branch_id is required and cannot be NULL")
	}
	// building_id 可选，可以为 NULL（支持 home care 等无 building 的场景）

	// 设置 floor 默认值（如果为 NULL 或空，设置为 "1F"）
	var floorSQL sql.NullString
	if unit.Floor.Valid && unit.Floor.String != "" {
		floorSQL = unit.Floor
	} else {
		floorSQL = sql.NullString{String: "1F", Valid: true}
	}

	var layoutConfigSQL sql.NullString
	if unit.LayoutConfig.Valid && unit.LayoutConfig.String != "" {
		layoutConfigSQL = sql.NullString{String: unit.LayoutConfig.String, Valid: true}
	}

	// 检查是否已存在相同的 unit（避免唯一约束冲突）
	// 唯一性约束：如果 branch_id 不为 NULL，使用 (tenant_id, branch_id, building_id, floor, unit_name)
	// 如果 branch_id 为 NULL，使用 (tenant_id, building_id, floor, unit_name)
	var existingUnitID string
	var checkQuery string
	floorValue := ""
	if floorSQL.Valid {
		floorValue = floorSQL.String
	}
	buildingIDValue := ""
	if buildingIDSQL.Valid {
		buildingIDValue = buildingIDSQL.String
	}

	if branchIDSQL.Valid {
		// branch_id 不为 NULL：检查 (tenant_id, branch_id, building_id, floor, unit_name)
		checkQuery = `
			SELECT unit_id::text
			FROM units
			WHERE tenant_id = $1
			  AND branch_id = $2
			  AND COALESCE(building_id::text, '') = COALESCE($3, '')
			  AND COALESCE(floor, '') = COALESCE($4, '')
			  AND unit_name = $5
			LIMIT 1
		`
		err := r.db.QueryRowContext(ctx, checkQuery,
			tenantID,
			branchIDSQL.String,
			buildingIDValue,
			floorValue,
			unit.UnitName,
		).Scan(&existingUnitID)
		if err == nil {
			return "", fmt.Errorf("unit already exists: unit_name=%s, building_id=%s, floor=%s, branch_id=%s (unit_id=%s)",
				unit.UnitName,
				buildingIDValue,
				floorValue,
				branchIDSQL.String,
				existingUnitID)
		} else if err != sql.ErrNoRows {
			return "", fmt.Errorf("failed to check duplicate unit: %w", err)
		}
	} else {
		// branch_id 为 NULL：检查 (tenant_id, building_id, floor, unit_name)
		checkQuery = `
			SELECT unit_id::text
			FROM units
			WHERE tenant_id = $1
			  AND branch_id IS NULL
			  AND COALESCE(building_id::text, '') = COALESCE($2, '')
			  AND COALESCE(floor, '') = COALESCE($3, '')
			  AND unit_name = $4
			LIMIT 1
		`
		err := r.db.QueryRowContext(ctx, checkQuery,
			tenantID,
			buildingIDValue,
			floorValue,
			unit.UnitName,
		).Scan(&existingUnitID)
		if err == nil {
			return "", fmt.Errorf("unit already exists: unit_name=%s, building_id=%s, floor=%s, branch_id=NULL (unit_id=%s)",
				unit.UnitName,
				buildingIDValue,
				floorValue,
				existingUnitID)
		} else if err != sql.ErrNoRows {
			return "", fmt.Errorf("failed to check duplicate unit: %w", err)
		}
	}

	// 构建 INSERT 语句：使用 building_id 而不是 building_name
	q := `
		INSERT INTO units (tenant_id, branch_id, unit_name, building_id, floor, layout_config, unit_type, is_public, is_shared_unit, timezone)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, COALESCE($8,false), COALESCE($9,false), $10)
		RETURNING unit_id::text
	`

	var unitID string
	var layoutConfigArg any
	if layoutConfigSQL.Valid {
		layoutConfigArg = layoutConfigSQL.String
	} else {
		layoutConfigArg = nil
	}
	var buildingIDArg any
	if buildingIDSQL.Valid {
		buildingIDArg = buildingIDSQL.String
	} else {
		buildingIDArg = nil
	}
	if err := r.db.QueryRowContext(ctx, q,
		tenantID,
		branchIDSQL,
		unit.UnitName,
		buildingIDArg,
		floorSQL,
		layoutConfigArg,
		unit.UnitType,
		unit.IsPublic,
		unit.IsSharedUnit,
		unit.Timezone,
	).Scan(&unitID); err != nil {
		// 如果仍然出现唯一约束冲突，提供更友好的错误信息
		floorDisplay := ""
		if floorSQL.Valid {
			floorDisplay = floorSQL.String
		}
		buildingDisplay := ""
		if buildingIDSQL.Valid {
			buildingDisplay = buildingIDSQL.String
		} else {
			buildingDisplay = "NULL"
		}
		branchDisplay := ""
		if branchIDSQL.Valid {
			branchDisplay = branchIDSQL.String
		} else {
			branchDisplay = "NULL"
		}
		if strings.Contains(err.Error(), "idx_units_unique_without_branch") {
			return "", fmt.Errorf("unit already exists: unit_name=%s, building_id=%s, floor=%s, branch_id=NULL (unique constraint violation)",
				unit.UnitName,
				buildingDisplay,
				floorDisplay)
		}
		if strings.Contains(err.Error(), "idx_units_unique_with_branch") {
			return "", fmt.Errorf("unit already exists: unit_name=%s, building_id=%s, floor=%s, branch_id=%s (unique constraint violation)",
				unit.UnitName,
				buildingDisplay,
				floorDisplay,
				branchDisplay)
		}
		return "", err
	}

	// 注意：branch_tag 和 area_tag 不应该在这里创建
	// unit 的 branch_name 和 area_name 只是数据

	return unitID, nil
}

// UpdateUnit: 更新 unit
// 替代触发器：trigger_sync_branch_tag, trigger_sync_area_tag, trigger_sync_units_groupList_to_cards
func (r *PostgresUnitsRepository) UpdateUnit(ctx context.Context, tenantID, unitID string, unit *domain.Unit) error {
	if tenantID == "" || unitID == "" {
		return fmt.Errorf("tenant_id and unit_id are required")
	}
	if unit == nil {
		return fmt.Errorf("unit is required")
	}

	// 先获取当前 unit 的信息（用于验证）
	currentUnit, err := r.GetUnit(ctx, tenantID, unitID)
	if err != nil {
		return err
	}

	// 处理 branch_id：如果提供了 BranchName，需要通过 branches 表查找 branch_id
	var branchIDSQL sql.NullString
	if unit.BranchName.Valid && unit.BranchName.String != "" && unit.BranchName.String != "default" {
		// 通过 branch_name 查找 branch_id
		var branchID string
		err := r.db.QueryRowContext(ctx,
			`SELECT branch_id::text FROM branches WHERE tenant_id = $1 AND branch_name = $2`,
			tenantID, unit.BranchName.String,
		).Scan(&branchID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("branch not found: branch_name=%s", unit.BranchName.String)
			}
			return fmt.Errorf("failed to find branch: %w", err)
		}
		branchIDSQL = sql.NullString{String: branchID, Valid: true}
	} else if unit.BranchID.Valid && unit.BranchID.String != "" {
		// 如果直接提供了 BranchID，使用它
		branchIDSQL = unit.BranchID
	}

	// 处理 building_id：如果提供了 BuildingID，使用它；如果提供了 BuildingName（向后兼容），通过 buildings 表查找 building_id
	var buildingIDSQL sql.NullString
	currentBuildingID := currentUnit.BuildingID
	if unit.BuildingID.Valid && unit.BuildingID.String != "" {
		// 如果直接提供了 BuildingID，使用它
		buildingIDSQL = unit.BuildingID
		// 验证 building_id 是否存在（如果改变）
		if !currentBuildingID.Valid || currentBuildingID.String != unit.BuildingID.String {
			var buildingBranchID sql.NullString
			err := r.db.QueryRowContext(ctx,
				`SELECT branch_id::text FROM buildings 
				 WHERE tenant_id = $1 AND building_id = $2`,
				tenantID, unit.BuildingID.String,
			).Scan(&buildingBranchID)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("building not found: building_id=%s", unit.BuildingID.String)
				}
				return fmt.Errorf("failed to validate building: %w", err)
			}
			// 如果 unit 也指定了 branch_id，验证它们是否一致
			if branchIDSQL.Valid && buildingBranchID.Valid {
				if branchIDSQL.String != buildingBranchID.String {
					return fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=%s", branchIDSQL.String, buildingBranchID.String)
				}
			} else if branchIDSQL.Valid {
				// unit 指定了 branch_id，但 building 的 branch_id 为 NULL，不一致
				return fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=NULL", branchIDSQL.String)
			} else if buildingBranchID.Valid {
				// building 有 branch_id，但 unit 没有指定，使用 building 的 branch_id
				branchIDSQL = buildingBranchID
			}
		}
	} else if unit.BuildingName.Valid && unit.BuildingName.String != "" && unit.BuildingName.String != "-" {
		// 如果提供了 BuildingName（向后兼容），通过 buildings 表查找 building_id
		currentBuildingName := ""
		if currentUnit.BuildingName.Valid {
			currentBuildingName = currentUnit.BuildingName.String
		}
		// 只有当 building_name 改变时才需要查找
		if unit.BuildingName.String != currentBuildingName {
			var buildingID string
			var buildingBranchID sql.NullString
			err := r.db.QueryRowContext(ctx,
				`SELECT building_id::text, branch_id::text FROM buildings 
				 WHERE tenant_id = $1 AND building_name = $2`,
				tenantID, unit.BuildingName.String,
			).Scan(&buildingID, &buildingBranchID)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("building not found: building_name=%s", unit.BuildingName.String)
				}
				return fmt.Errorf("failed to validate building: %w", err)
			}
			buildingIDSQL = sql.NullString{String: buildingID, Valid: true}
			// 如果 unit 也指定了 branch_id，验证它们是否一致
			if branchIDSQL.Valid && buildingBranchID.Valid {
				if branchIDSQL.String != buildingBranchID.String {
					return fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=%s", branchIDSQL.String, buildingBranchID.String)
				}
			} else if branchIDSQL.Valid {
				// unit 指定了 branch_id，但 building 的 branch_id 为 NULL，不一致
				return fmt.Errorf("branch_id mismatch: unit.branch_id=%s, building.branch_id=NULL", branchIDSQL.String)
			} else if buildingBranchID.Valid {
				// building 有 branch_id，但 unit 没有指定，使用 building 的 branch_id
				branchIDSQL = buildingBranchID
			}
		} else {
			// building_name 没有改变，使用当前的 building_id
			buildingIDSQL = currentBuildingID
		}
	}

	// 验证：branch_id 必须提供，不能为 NULL（业务规则：一家机构必然有一个分支或总部）
	// 如果没有提供新的 branch_id，使用当前的 branch_id
	if !branchIDSQL.Valid {
		// 使用当前的 branch_id
		branchIDSQL = currentUnit.BranchID
		if !branchIDSQL.Valid {
			return fmt.Errorf("branch_id is required and cannot be NULL")
		}
	}
	// building_id 可选，可以为 NULL（支持 home care 等无 building 的场景）
	// 如果没有提供新的 building_id，保持当前值（可能是 NULL）
	if !buildingIDSQL.Valid {
		buildingIDSQL = currentBuildingID
	}

	// 构建动态 UPDATE 语句
	set := []string{}
	args := []any{tenantID, unitID}
	argN := 3

	add := func(col string, v any) {
		set = append(set, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, v)
		argN++
	}

	// 处理 branch_id：必须提供，不能为 NULL（业务规则要求）
	// branch_id 已经在验证阶段确保不为 NULL
	if branchIDSQL.Valid {
		add("branch_id", branchIDSQL.String)
	} else {
		// 这不应该发生，因为已经在验证阶段检查过
		return fmt.Errorf("branch_id is required and cannot be NULL")
	}
	if unit.UnitName != "" {
		add("unit_name", unit.UnitName)
	}
	// building_id：如果提供了 BuildingID 或 BuildingName，更新它
	if buildingIDSQL.Valid {
		add("building_id", buildingIDSQL.String)
	} else if unit.BuildingID.Valid && unit.BuildingID.String == "" {
		// 如果明确设置为空字符串，设置为 NULL
		set = append(set, "building_id = NULL")
	} else if unit.BuildingName.Valid && (unit.BuildingName.String == "" || unit.BuildingName.String == "-") {
		// 如果 BuildingName 明确设置为空字符串或 "-"，设置为 NULL（向后兼容）
		set = append(set, "building_id = NULL")
	}
	// 注意：如果没有提供 building_id，保持当前值（不更新）
	if unit.Floor.Valid && unit.Floor.String != "" {
		add("floor", unit.Floor.String)
	}
	if unit.LayoutConfig.Valid && unit.LayoutConfig.String != "" {
		set = append(set, fmt.Sprintf("layout_config = $%d::jsonb", argN))
		args = append(args, unit.LayoutConfig.String)
		argN++
	}
	if unit.UnitType != "" {
		add("unit_type", unit.UnitType)
	}
	set = append(set, fmt.Sprintf("is_public = $%d", argN))
	args = append(args, unit.IsPublic)
	argN++
	set = append(set, fmt.Sprintf("is_shared_unit = $%d", argN))
	args = append(args, unit.IsSharedUnit)
	argN++
	if unit.Timezone != "" {
		add("timezone", unit.Timezone)
	}

	if len(set) == 0 {
		return nil
	}

	q := fmt.Sprintf("UPDATE units SET %s WHERE tenant_id = $1 AND unit_id = $2", strings.Join(set, ", "))
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return err
	}

	// 注意：branch_tag 和 area_tag 不应该在这里创建
	// unit 的 branch_name 和 area_name 只是数据

	return nil
}

// DeleteUnit: 删除 unit
// 替代触发器：无（仅删除）
func (r *PostgresUnitsRepository) DeleteUnit(ctx context.Context, tenantID, unitID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM units WHERE tenant_id = $1 AND unit_id = $2", tenantID, unitID)
	return err
}

// ============================================
// Room 操作
// ============================================

// ListRooms: 查询 rooms 列表
// 替代触发器：无（仅查询）
// 参数说明：
//   - search: 可选，按 room_name 模糊搜索（如果为空则不搜索）
func (r *PostgresUnitsRepository) ListRooms(ctx context.Context, tenantID, unitID string, search string) ([]*domain.Room, error) {
	if tenantID == "" || unitID == "" {
		return []*domain.Room{}, nil
	}

	where := []string{"r.tenant_id = $1", "r.unit_id = $2"}
	args := []any{tenantID, unitID}
	argN := 3

	// 添加搜索条件（如果提供）
	if search != "" {
		where = append(where, fmt.Sprintf("r.room_name ILIKE $%d", argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	q := fmt.Sprintf(`
		SELECT 
			r.room_id::text,
			r.tenant_id::text,
			r.unit_id::text,
			u.unit_name,
			u.floor,
			r.room_name,
			CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r
		LEFT JOIN units u ON u.unit_id = r.unit_id
		WHERE %s
		ORDER BY r.room_name
	`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]*domain.Room, 0)
	for rows.Next() {
		var room domain.Room
		var unitName, floor, layoutConfig sql.NullString
		if err := rows.Scan(&room.RoomID, &room.TenantID, &room.UnitID, &unitName, &floor, &room.RoomName, &layoutConfig); err != nil {
			return nil, err
		}
		room.UnitName = unitName
		room.Floor = floor
		room.LayoutConfig = layoutConfig
		rooms = append(rooms, &room)
	}
	return rooms, rows.Err()
}

// ListRoomsWithBeds: 查询 unit 下全部 rooms 及每 room 下全部 beds，不做住户相关过滤；bed 状态由前端/业务层统一展示
func (r *PostgresUnitsRepository) ListRoomsWithBeds(ctx context.Context, tenantID, unitID, search, residentID string) ([]*RoomWithBeds, error) {
	if tenantID == "" || unitID == "" {
		return []*RoomWithBeds{}, nil
	}
	_ = residentID // 不再按住户过滤，保留参数兼容调用方

	rooms, err := r.ListRooms(ctx, tenantID, unitID, search)
	if err != nil {
		return nil, err
	}
	if len(rooms) == 0 {
		return []*RoomWithBeds{}, nil
	}

	roomIDs := make([]string, len(rooms))
	for i, room := range rooms {
		roomIDs[i] = room.RoomID
	}
	in := make([]string, len(roomIDs))
	args := make([]any, 0, len(roomIDs)+1)
	args = append(args, tenantID)
	for i, id := range roomIDs {
		in[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	qBeds := `
		SELECT 
			b.bed_id::text,
			b.tenant_id::text,
			b.room_id::text,
			r.room_name,
			b.bed_name,
			b.mattress_material,
			b.mattress_thickness
		FROM beds b
		LEFT JOIN rooms r ON r.room_id = b.room_id
		WHERE b.tenant_id = $1 AND b.room_id IN (` + strings.Join(in, ",") + `)
		ORDER BY b.bed_name
		`
	brows, err := r.db.QueryContext(ctx, qBeds, args...)
	if err != nil {
		return nil, err
	}
	defer brows.Close()

	bedsByRoom := map[string][]*domain.Bed{}
	for brows.Next() {
		var bed domain.Bed
		var roomName, mattressMaterial, mattressThickness sql.NullString
		if err := brows.Scan(
			&bed.BedID,
			&bed.TenantID,
			&bed.RoomID,
			&roomName,
			&bed.BedName,
			&mattressMaterial,
			&mattressThickness,
		); err != nil {
			return nil, err
		}
		bed.RoomName = roomName
		bed.MattressMaterial = mattressMaterial
		bed.MattressThickness = mattressThickness
		bedsByRoom[bed.RoomID] = append(bedsByRoom[bed.RoomID], &bed)
	}
	if err := brows.Err(); err != nil {
		return nil, err
	}

	// 组合结果
	out := make([]*RoomWithBeds, 0, len(rooms))
	for _, room := range rooms {
		beds := bedsByRoom[room.RoomID]
		if beds == nil {
			beds = []*domain.Bed{}
		}
		out = append(out, &RoomWithBeds{
			Room: room,
			Beds: beds,
		})
	}

	return out, nil
}

// ListRoomsByBranch: 按 branch 列出所有 room，带 unit 信息与 is_full、is_bound（供前端 (full) 红字灰行、橙/绿）
func (r *PostgresUnitsRepository) ListRoomsByBranch(ctx context.Context, tenantID, branchID string) ([]*RoomWithAvailability, error) {
	if tenantID == "" || branchID == "" {
		return nil, nil
	}
	q := `
		SELECT 
			r.room_id::text,
			r.tenant_id::text,
			r.unit_id::text,
			COALESCE(u.unit_name, '') AS unit_name,
			COALESCE(b.building_name, '') AS building_name,
			COALESCE(u.floor, '') AS floor,
			r.room_name,
			COALESCE(u.unit_type, '') AS unit_type,
			CASE WHEN u.is_public THEN 'Public' WHEN u.is_shared_unit THEN 'Share' ELSE 'VIP' END AS facility_type,
			(CASE WHEN EXISTS (SELECT 1 FROM beds b WHERE b.tenant_id = r.tenant_id AND b.room_id = r.room_id) THEN
				NOT EXISTS (SELECT 1 FROM beds b WHERE b.tenant_id = r.tenant_id AND b.room_id = r.room_id
					AND NOT EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = b.tenant_id AND res.bed_id = b.bed_id))
			ELSE FALSE END) AS is_full,
			(EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = r.tenant_id AND res.room_id = r.room_id)
			 OR EXISTS (SELECT 1 FROM residents res INNER JOIN beds b ON b.tenant_id = res.tenant_id AND b.bed_id = res.bed_id AND b.room_id = r.room_id WHERE res.tenant_id = r.tenant_id)) AS is_bound
		FROM rooms r
		INNER JOIN units u ON u.unit_id = r.unit_id AND u.tenant_id = r.tenant_id
		LEFT JOIN buildings b ON b.building_id = u.building_id AND b.tenant_id = u.tenant_id
		WHERE r.tenant_id = $1 AND u.branch_id = $2
		ORDER BY COALESCE(b.building_name, ''), COALESCE(u.floor, ''), u.unit_name, r.room_name
	`
	rows, err := r.db.QueryContext(ctx, q, tenantID, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*RoomWithAvailability
	for rows.Next() {
		var item RoomWithAvailability
		if err := rows.Scan(&item.RoomID, &item.TenantID, &item.UnitID, &item.UnitName, &item.BuildingName, &item.Floor, &item.RoomName, &item.UnitType, &item.FacilityType, &item.IsFull, &item.IsBound); err != nil {
			return nil, err
		}
		out = append(out, &item)
	}
	return out, rows.Err()
}

// GetUnitAvailability: 批量查询 unit 的 has_available_bed、is_bound
func (r *PostgresUnitsRepository) GetUnitAvailability(ctx context.Context, tenantID string, unitIDs []string) (hasAvailableBed, isBound map[string]bool, err error) {
	hasAvailableBed = make(map[string]bool)
	isBound = make(map[string]bool)
	if tenantID == "" || len(unitIDs) == 0 {
		return hasAvailableBed, isBound, nil
	}
	ph := make([]string, len(unitIDs))
	args := []any{tenantID}
	for i := range unitIDs {
		ph[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, unitIDs[i])
	}
	inClause := strings.Join(ph, ",")
	// has_available_bed: unit 下存在至少一张未被占用的 bed
	qAvail := `
		SELECT DISTINCT u.unit_id::text FROM units u
		INNER JOIN rooms rm ON rm.unit_id = u.unit_id AND rm.tenant_id = u.tenant_id
		INNER JOIN beds b ON b.room_id = rm.room_id AND b.tenant_id = u.tenant_id
		WHERE u.tenant_id = $1 AND u.unit_id IN (` + inClause + `)
		AND NOT EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = b.tenant_id AND res.bed_id = b.bed_id)
	`
	rowsAvail, err := r.db.QueryContext(ctx, qAvail, args...)
	if err != nil {
		return nil, nil, err
	}
	for rowsAvail.Next() {
		var id string
		if _ = rowsAvail.Scan(&id); id != "" {
			hasAvailableBed[id] = true
		}
	}
	rowsAvail.Close()
	// is_bound: 有住户绑定了该 unit（unit_id / 该 unit 下 room / 该 unit 下 bed）
	qBound := `
		SELECT DISTINCT u.unit_id::text FROM units u
		INNER JOIN residents res ON res.tenant_id = u.tenant_id
		AND (res.unit_id = u.unit_id
		     OR res.room_id IN (SELECT room_id FROM rooms WHERE unit_id = u.unit_id AND tenant_id = u.tenant_id)
		     OR res.bed_id IN (SELECT b.bed_id FROM beds b INNER JOIN rooms r ON r.room_id = b.room_id AND r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id WHERE b.tenant_id = u.tenant_id))
		WHERE u.tenant_id = $1 AND u.unit_id IN (` + inClause + `)
	`
	rowsBound, err := r.db.QueryContext(ctx, qBound, args...)
	if err != nil {
		return nil, nil, err
	}
	for rowsBound.Next() {
		var id string
		if _ = rowsBound.Scan(&id); id != "" {
			isBound[id] = true
		}
	}
	rowsBound.Close()
	return hasAvailableBed, isBound, nil
}

// GetRoom: 获取单个 room
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) GetRoom(ctx context.Context, tenantID, roomID string) (*domain.Room, error) {
	if tenantID == "" || roomID == "" {
		return nil, sql.ErrNoRows
	}

	q := `
		SELECT 
			r.room_id::text,
			r.tenant_id::text,
			r.unit_id::text,
			u.unit_name,
			u.floor,
			r.room_name,
			CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r
		LEFT JOIN units u ON u.unit_id = r.unit_id AND u.tenant_id = r.tenant_id
		WHERE r.tenant_id = $1 AND r.room_id = $2
	`
	var room domain.Room
	var unitName, floor, layoutConfig sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, roomID).Scan(
		&room.RoomID,
		&room.TenantID,
		&room.UnitID,
		&unitName,
		&floor,
		&room.RoomName,
		&layoutConfig,
	)
	if err != nil {
		return nil, err
	}
	room.UnitName = unitName
	room.Floor = floor
	room.LayoutConfig = layoutConfig
	return &room, nil
}

// CreateRoom: 创建 room
// 替代触发器：无（仅插入，但需要验证 unit 存在）
func (r *PostgresUnitsRepository) CreateRoom(ctx context.Context, tenantID, unitID string, room *domain.Room) (string, error) {
	if tenantID == "" || unitID == "" {
		return "", fmt.Errorf("tenant_id and unit_id are required")
	}
	if room == nil {
		return "", fmt.Errorf("room is required")
	}
	if room.RoomName == "" {
		return "", fmt.Errorf("room_name is required")
	}

	// 验证 unit 是否存在
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM units WHERE tenant_id = $1 AND unit_id = $2)`,
		tenantID, unitID,
	).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("failed to validate unit: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("unit not found: unit_id=%s (room must belong to an existing unit)", unitID)
	}

	var layoutConfigSQL sql.NullString
	if room.LayoutConfig.Valid && room.LayoutConfig.String != "" {
		layoutConfigSQL = sql.NullString{String: room.LayoutConfig.String, Valid: true}
	}

	var roomID string
	if layoutConfigSQL.Valid {
		q := `
			INSERT INTO rooms (tenant_id, unit_id, room_name, layout_config)
			VALUES ($1, $2, $3, $4::jsonb)
			RETURNING room_id::text
		`
		if err := r.db.QueryRowContext(ctx, q, tenantID, unitID, room.RoomName, layoutConfigSQL.String).Scan(&roomID); err != nil {
			return "", err
		}
	} else {
		q := `
			INSERT INTO rooms (tenant_id, unit_id, room_name)
			VALUES ($1, $2, $3)
			RETURNING room_id::text
		`
		if err := r.db.QueryRowContext(ctx, q, tenantID, unitID, room.RoomName).Scan(&roomID); err != nil {
			return "", err
		}
	}

	return roomID, nil
}

// UpdateRoom: 更新 room
// 替代触发器：无（仅更新）
func (r *PostgresUnitsRepository) UpdateRoom(ctx context.Context, tenantID, roomID string, room *domain.Room) error {
	if tenantID == "" || roomID == "" {
		return fmt.Errorf("tenant_id and room_id are required")
	}
	if room == nil {
		return fmt.Errorf("room is required")
	}

	set := []string{}
	args := []any{tenantID, roomID}
	argN := 3

	if room.RoomName != "" {
		set = append(set, fmt.Sprintf("room_name = $%d", argN))
		args = append(args, room.RoomName)
		argN++
	}
	if room.LayoutConfig.Valid {
		if room.LayoutConfig.String == "" {
			set = append(set, "layout_config = NULL")
		} else {
			set = append(set, fmt.Sprintf("layout_config = $%d::jsonb", argN))
			args = append(args, room.LayoutConfig.String)
			argN++
		}
	}

	if len(set) == 0 {
		return nil
	}

	q := "UPDATE rooms SET " + strings.Join(set, ", ") + " WHERE tenant_id = $1 AND room_id = $2"
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return err
	}

	return nil
}

// DeleteRoom: 删除 room
// 替代触发器：无（仅删除，依赖 DB CASCADE）
func (r *PostgresUnitsRepository) DeleteRoom(ctx context.Context, tenantID, roomID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM rooms WHERE tenant_id = $1 AND room_id = $2", tenantID, roomID)
	return err
}

// ============================================
// Bed 操作
// ============================================

// ListBeds: 查询 beds 列表（全部，不做可用性过滤）
func (r *PostgresUnitsRepository) ListBeds(ctx context.Context, tenantID, roomID, search string) ([]*domain.Bed, error) {
	if tenantID == "" || roomID == "" {
		return []*domain.Bed{}, nil
	}

	where := []string{"b.tenant_id = $1", "b.room_id = $2"}
	args := []any{tenantID, roomID}
	argN := 3

	if search != "" {
		where = append(where, fmt.Sprintf("b.bed_name ILIKE $%d", argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	q := fmt.Sprintf(`
		SELECT 
			b.bed_id::text,
			b.tenant_id::text,
			b.room_id::text,
			r.room_name,
			b.bed_name,
			b.mattress_material,
			b.mattress_thickness
		FROM beds b
		LEFT JOIN rooms r ON r.room_id = b.room_id
		WHERE %s
		ORDER BY b.bed_name
	`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	beds := make([]*domain.Bed, 0)
	for rows.Next() {
		var bed domain.Bed
		var roomName, mattressMaterial, mattressThickness sql.NullString
		if err := rows.Scan(
			&bed.BedID,
			&bed.TenantID,
			&bed.RoomID,
			&roomName,
			&bed.BedName,
			&mattressMaterial,
			&mattressThickness,
		); err != nil {
			return nil, err
		}
		bed.RoomName = roomName
		bed.MattressMaterial = mattressMaterial
		bed.MattressThickness = mattressThickness
		beds = append(beds, &bed)
	}
	return beds, rows.Err()
}

// ListBedsWithResident 返回 room 下全部 bed 及 resident_id
func (r *PostgresUnitsRepository) ListBedsWithResident(ctx context.Context, tenantID, roomID, search string) ([]*BedWithResident, error) {
	if tenantID == "" || roomID == "" {
		return []*BedWithResident{}, nil
	}
	where := []string{"b.tenant_id = $1", "b.room_id = $2"}
	args := []any{tenantID, roomID}
	argN := 3
	if search != "" {
		where = append(where, fmt.Sprintf("b.bed_name ILIKE $%d", argN))
		args = append(args, "%"+search+"%")
		argN++
	}
	q := fmt.Sprintf(`
		SELECT b.bed_id::text, b.tenant_id::text, b.room_id::text, r.room_name, b.bed_name, b.mattress_material, b.mattress_thickness, res.resident_id::text
		FROM beds b
		LEFT JOIN rooms r ON r.room_id = b.room_id
		LEFT JOIN residents res ON res.bed_id = b.bed_id AND res.tenant_id = b.tenant_id
		WHERE %s
		ORDER BY b.bed_name
	`, strings.Join(where, " AND "))
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*BedWithResident
	for rows.Next() {
		var bed domain.Bed
		var roomName, mattressMaterial, mattressThickness sql.NullString
		var residentID sql.NullString
		if err := rows.Scan(
			&bed.BedID,
			&bed.TenantID,
			&bed.RoomID,
			&roomName,
			&bed.BedName,
			&mattressMaterial,
			&mattressThickness,
			&residentID,
		); err != nil {
			return nil, err
		}
		bed.RoomName = roomName
		bed.MattressMaterial = mattressMaterial
		bed.MattressThickness = mattressThickness
		item := &BedWithResident{Bed: &bed}
		if residentID.Valid && residentID.String != "" {
			item.ResidentID = &residentID.String
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListAvailableBeds: 仅返回可用床位（未被占用或仅被 residentID 占用）；VIP 规则：非 Share 时 room 被其他住户占用则整间不可选；Share 同房多床仅按 bed 占用判断
func (r *PostgresUnitsRepository) ListAvailableBeds(ctx context.Context, tenantID, roomID, search, residentID string) ([]*domain.Bed, error) {
	if tenantID == "" || roomID == "" {
		return []*domain.Bed{}, nil
	}

	var isSharedUnit bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(u.is_shared_unit, false) FROM rooms r
		INNER JOIN units u ON u.unit_id = r.unit_id AND u.tenant_id = r.tenant_id
		WHERE r.tenant_id = $1 AND r.room_id = $2`, tenantID, roomID,
	).Scan(&isSharedUnit); err != nil {
		if err == sql.ErrNoRows {
			return []*domain.Bed{}, nil
		}
		return nil, err
	}

	where := []string{"b.tenant_id = $1", "b.room_id = $2"}
	args := []any{tenantID, roomID}
	argN := 3

	// bed 未被占用或仅被当前住户占用
	if residentID == "" {
		where = append(where, `NOT EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = b.tenant_id AND res.bed_id = b.bed_id)`)
	} else {
		where = append(where, fmt.Sprintf(`(NOT EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = b.tenant_id AND res.bed_id = b.bed_id) OR EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = b.tenant_id AND res.bed_id = b.bed_id AND res.resident_id = $%d))`, argN))
		args = append(args, residentID)
		argN++
	}

	// VIP 规则：非 Share 时 room 被其他住户占用则不可选；Share 允许多住户同房不同床，不按 room 整间过滤
	if !isSharedUnit {
		if residentID == "" {
			where = append(where, `NOT EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = b.tenant_id AND res.room_id = b.room_id)`)
		} else {
			where = append(where, fmt.Sprintf(`NOT EXISTS (SELECT 1 FROM residents res WHERE res.tenant_id = b.tenant_id AND res.room_id = b.room_id AND res.resident_id IS NOT NULL AND res.resident_id != $%d)`, argN))
			args = append(args, residentID)
			argN++
		}
	}

	if search != "" {
		where = append(where, fmt.Sprintf("b.bed_name ILIKE $%d", argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	q := fmt.Sprintf(`
		SELECT b.bed_id::text, b.tenant_id::text, b.room_id::text, r.room_name, b.bed_name, b.mattress_material, b.mattress_thickness
		FROM beds b
		LEFT JOIN rooms r ON r.room_id = b.room_id
		WHERE %s
		ORDER BY b.bed_name
	`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	beds := make([]*domain.Bed, 0)
	for rows.Next() {
		var bed domain.Bed
		var roomName, mattressMaterial, mattressThickness sql.NullString
		if err := rows.Scan(
			&bed.BedID,
			&bed.TenantID,
			&bed.RoomID,
			&roomName,
			&bed.BedName,
			&mattressMaterial,
			&mattressThickness,
		); err != nil {
			return nil, err
		}
		bed.RoomName = roomName
		bed.MattressMaterial = mattressMaterial
		bed.MattressThickness = mattressThickness
		beds = append(beds, &bed)
	}
	return beds, rows.Err()
}

// GetBed: 获取单个 bed
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) GetBed(ctx context.Context, tenantID, bedID string) (*domain.Bed, error) {
	if tenantID == "" || bedID == "" {
		return nil, sql.ErrNoRows
	}

	q := `
		SELECT 
			b.bed_id::text,
			b.tenant_id::text,
			b.room_id::text,
			r.room_name,
			b.bed_name,
			b.mattress_material,
			b.mattress_thickness
		FROM beds b
		LEFT JOIN rooms r ON r.room_id = b.room_id
		WHERE b.tenant_id = $1 AND b.bed_id = $2
	`
	var bed domain.Bed
	var roomName, mattressMaterial, mattressThickness sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, bedID).Scan(
		&bed.BedID,
		&bed.TenantID,
		&bed.RoomID,
		&roomName,
		&bed.BedName,
		&mattressMaterial,
		&mattressThickness,
	)
	if err != nil {
		return nil, err
	}
	bed.RoomName = roomName
	bed.MattressMaterial = mattressMaterial
	bed.MattressThickness = mattressThickness
	return &bed, nil
}

// CreateBed: 创建 bed
// 替代触发器：无（仅插入，但需要验证 room 存在）
func (r *PostgresUnitsRepository) CreateBed(ctx context.Context, tenantID, roomID string, bed *domain.Bed) (string, error) {
	if tenantID == "" || roomID == "" {
		return "", fmt.Errorf("tenant_id and room_id are required")
	}
	if bed == nil {
		return "", fmt.Errorf("bed is required")
	}
	if bed.BedName == "" {
		return "", fmt.Errorf("bed_name is required")
	}

	// 验证 room 是否存在
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM rooms WHERE tenant_id = $1 AND room_id = $2)`,
		tenantID, roomID,
	).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("failed to validate room: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("room not found: room_id=%s (bed must belong to an existing room)", roomID)
	}

	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算

	var mattressMaterialSQL, mattressThicknessSQL sql.NullString
	if bed.MattressMaterial.Valid && bed.MattressMaterial.String != "" {
		mattressMaterialSQL = sql.NullString{String: bed.MattressMaterial.String, Valid: true}
	}
	if bed.MattressThickness.Valid && bed.MattressThickness.String != "" {
		mattressThicknessSQL = sql.NullString{String: bed.MattressThickness.String, Valid: true}
	}

	var bedID string
	q := `
		INSERT INTO beds (tenant_id, room_id, bed_name, mattress_material, mattress_thickness)
		SELECT tenant_id, $1, $2, $3, $4
		FROM rooms WHERE room_id = $1
		RETURNING bed_id::text
	`
	if err := r.db.QueryRowContext(ctx, q, roomID, bed.BedName, mattressMaterialSQL, mattressThicknessSQL).Scan(&bedID); err != nil {
		return "", err
	}

	return bedID, nil
}

// UpdateBed: 更新 bed
// 替代触发器：无（仅更新）
func (r *PostgresUnitsRepository) UpdateBed(ctx context.Context, tenantID, bedID string, bed *domain.Bed) error {
	if tenantID == "" || bedID == "" {
		return fmt.Errorf("tenant_id and bed_id are required")
	}
	if bed == nil {
		return fmt.Errorf("bed is required")
	}

	set := []string{}
	args := []any{tenantID, bedID}
	argN := 3

	if bed.BedName != "" {
		set = append(set, fmt.Sprintf("bed_name = $%d", argN))
		args = append(args, bed.BedName)
		argN++
	}
	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	if bed.MattressMaterial.Valid {
		if bed.MattressMaterial.String == "" {
			set = append(set, "mattress_material = NULL")
		} else {
			set = append(set, fmt.Sprintf("mattress_material = $%d", argN))
			args = append(args, bed.MattressMaterial.String)
			argN++
		}
	}
	if bed.MattressThickness.Valid {
		if bed.MattressThickness.String == "" {
			set = append(set, "mattress_thickness = NULL")
		} else {
			set = append(set, fmt.Sprintf("mattress_thickness = $%d", argN))
			args = append(args, bed.MattressThickness.String)
		}
	}

	if len(set) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE beds
		SET %s
		WHERE tenant_id = $1 AND bed_id = $2
	`, strings.Join(set, ", "))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update bed: %w", err)
	}

	return nil
}

// DeleteBed: 删除 bed
// 替代触发器：无（仅删除，依赖 DB CASCADE）
func (r *PostgresUnitsRepository) DeleteBed(ctx context.Context, tenantID, bedID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM beds WHERE tenant_id = $1 AND bed_id = $2", tenantID, bedID)
	return err
}

// ============================================
// 辅助函数
// ============================================

func nullStringToAny(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}

// getBuildingDisplay 获取 building 的显示值（用于错误信息）
func getBuildingDisplay(building sql.NullString) string {
	if !building.Valid {
		return "NULL"
	}
	if building.String == "" {
		return "''"
	}
	return building.String
}

// ============================================
// 批量查询（用于 ListUnitsWithFullHierarchy）
// ============================================

// ListRoomsByUnitIDs 批量查询多个 units 的 rooms
func (r *PostgresUnitsRepository) ListRoomsByUnitIDs(ctx context.Context, tenantID string, unitIDs []string) ([]*domain.Room, error) {
	if tenantID == "" || len(unitIDs) == 0 {
		return []*domain.Room{}, nil
	}

	// 构建 IN 子句
	in := make([]string, len(unitIDs))
	args := make([]any, 0, len(unitIDs)+1)
	args = append(args, tenantID)
	for i, id := range unitIDs {
		in[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	q := `
		SELECT 
			r.room_id::text,
			r.tenant_id::text,
			r.unit_id::text,
			u.unit_name,
			u.floor,
			r.room_name,
			CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r
		LEFT JOIN units u ON u.unit_id = r.unit_id
		WHERE r.tenant_id = $1 AND r.unit_id IN (` + strings.Join(in, ",") + `)
		ORDER BY r.unit_id, r.room_name
	`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]*domain.Room, 0)
	for rows.Next() {
		var room domain.Room
		var unitName, floor, layoutConfig sql.NullString
		if err := rows.Scan(&room.RoomID, &room.TenantID, &room.UnitID, &unitName, &floor, &room.RoomName, &layoutConfig); err != nil {
			return nil, err
		}
		room.UnitName = unitName
		room.Floor = floor
		room.LayoutConfig = layoutConfig
		rooms = append(rooms, &room)
	}
	return rooms, rows.Err()
}

// ListBedsByRoomIDs 批量查询多个 rooms 的 beds
func (r *PostgresUnitsRepository) ListBedsByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) ([]*domain.Bed, error) {
	if tenantID == "" || len(roomIDs) == 0 {
		return []*domain.Bed{}, nil
	}

	// 构建 IN 子句
	in := make([]string, len(roomIDs))
	args := make([]any, 0, len(roomIDs)+1)
	args = append(args, tenantID)
	for i, id := range roomIDs {
		in[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	q := `
		SELECT 
			b.bed_id::text,
			b.tenant_id::text,
			b.room_id::text,
			r.room_name,
			b.bed_name,
			b.mattress_material,
			b.mattress_thickness
		FROM beds b
		LEFT JOIN rooms r ON r.room_id = b.room_id
		WHERE b.tenant_id = $1 AND b.room_id IN (` + strings.Join(in, ",") + `)
		ORDER BY b.room_id, b.bed_name
	`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	beds := make([]*domain.Bed, 0)
	for rows.Next() {
		var bed domain.Bed
		var roomName, mattressMaterial, mattressThickness sql.NullString
		if err := rows.Scan(
			&bed.BedID,
			&bed.TenantID,
			&bed.RoomID,
			&roomName,
			&bed.BedName,
			&mattressMaterial,
			&mattressThickness,
		); err != nil {
			return nil, err
		}
		bed.RoomName = roomName
		bed.MattressMaterial = mattressMaterial
		bed.MattressThickness = mattressThickness
		beds = append(beds, &bed)
	}
	return beds, rows.Err()
}
