package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
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
func (r *PostgresUnitsRepository) ListBuildings(ctx context.Context, tenantID string, branchTag string) ([]*domain.Building, error) {
	if tenantID == "" {
		return []*domain.Building{}, nil
	}

	where := "tenant_id = $1"
	args := []any{tenantID}
	argIdx := 2
	if branchTag != "" {
		where += " AND COALESCE(branch_tag, '-') = $" + fmt.Sprintf("%d", argIdx)
		args = append(args, branchTag)
		argIdx++
	}

	q := `
		SELECT
			building_id::text,
			tenant_id::text,
			branch_tag,
			building_name,
			created_at,
			updated_at
		FROM buildings
		WHERE ` + where + `
		  AND NOT (COALESCE(branch_tag, '-') = '-' AND building_name = '-')
		ORDER BY COALESCE(branch_tag, '-'), building_name
	`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.Building{}
	for rows.Next() {
		var b domain.Building
		var branchTag sql.NullString
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&b.BuildingID, &b.TenantID, &branchTag, &b.BuildingName, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		b.BranchTag = branchTag
		b.CreatedAt = createdAt
		b.UpdatedAt = updatedAt
		// Additional check: filter out buildings where both branch_tag and building_name are '-'
		if branchTag.Valid && branchTag.String == "-" && b.BuildingName == "-" {
			continue
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}

// GetBuilding: 从 buildings 表获取 building 信息
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) GetBuilding(ctx context.Context, tenantID, buildingID string) (*domain.Building, error) {
	if tenantID == "" || buildingID == "" {
		return nil, fmt.Errorf("tenant_id and building_id are required")
	}

	q := `
		SELECT
			building_id::text,
			tenant_id::text,
			branch_tag,
			building_name,
			created_at,
			updated_at
		FROM buildings
		WHERE tenant_id = $1 AND building_id = $2
	`
	var b domain.Building
	var branchTag sql.NullString
	var createdAt, updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, q, tenantID, buildingID).Scan(
		&b.BuildingID,
		&b.TenantID,
		&branchTag,
		&b.BuildingName,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("building not found: building_id=%s", buildingID)
		}
		return nil, err
	}
	b.BranchTag = branchTag
	b.CreatedAt = createdAt
	b.UpdatedAt = updatedAt
	return &b, nil
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

	// 验证：branch_tag 或 building_name 必须有一个不为空
	branchTagValue := ""
	if building.BranchTag.Valid {
		branchTagValue = building.BranchTag.String
	}
	if (branchTagValue == "" || branchTagValue == "-") && (building.BuildingName == "" || building.BuildingName == "-") {
		return "", fmt.Errorf("branch_tag or building_name must be provided (at least one must not be empty)")
	}

	// 设置默认值
	if building.BuildingName == "" {
		building.BuildingName = "-"
	}

	var buildingID string
	var insertedBranchTag sql.NullString
	if branchTagValue == "" || branchTagValue == "-" {
		// branch_tag 为 "-" 或空时，插入 NULL
		err := r.db.QueryRowContext(ctx,
			`INSERT INTO buildings (tenant_id, building_name)
			 VALUES ($1, $2)
			 ON CONFLICT (tenant_id, building_name) WHERE branch_tag IS NULL
			 DO UPDATE SET building_name = EXCLUDED.building_name, updated_at = CURRENT_TIMESTAMP
			 RETURNING building_id::text, branch_tag`,
			tenantID, building.BuildingName,
		).Scan(&buildingID, &insertedBranchTag)
		if err != nil {
			return "", fmt.Errorf("failed to create building: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx,
			`INSERT INTO buildings (tenant_id, branch_tag, building_name)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (tenant_id, branch_tag, building_name) WHERE branch_tag IS NOT NULL
			 DO UPDATE SET building_name = EXCLUDED.building_name, updated_at = CURRENT_TIMESTAMP
			 RETURNING building_id::text, branch_tag`,
			tenantID, branchTagValue, building.BuildingName,
		).Scan(&buildingID, &insertedBranchTag)
		if err != nil {
			return "", fmt.Errorf("failed to create building: %w", err)
		}
	}

	// 同步branch_tag到tags_catalog目录（替代trigger_sync_branch_tag）
	if branchTagValue != "" && branchTagValue != "-" {
		_, err := r.db.ExecContext(ctx,
			`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
			tenantID, branchTagValue, "branch_tag",
		)
		if err != nil {
			fmt.Printf("[PostgresUnitsRepository] Failed to sync branch_tag to catalog: %v\n", err)
		}
	}

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
	var oldBranchTag sql.NullString
	var oldBuildingName string
	err := r.db.QueryRowContext(ctx,
		`SELECT branch_tag, building_name 
		 FROM buildings 
		 WHERE tenant_id = $1 AND building_id = $2`,
		tenantID, buildingID,
	).Scan(&oldBranchTag, &oldBuildingName)

	if err == sql.ErrNoRows {
		return fmt.Errorf("building not found")
	}
	if err != nil {
		return fmt.Errorf("failed to find building: %w", err)
	}

	// 获取新的值
	newBranchTagValue := ""
	if building.BranchTag.Valid {
		newBranchTagValue = building.BranchTag.String
	}
	newBuildingName := building.BuildingName
	if newBuildingName == "" {
		newBuildingName = oldBuildingName
	}

	// 验证：branch_tag 或 building_name 必须有一个不为空
	if (newBranchTagValue == "" || newBranchTagValue == "-") && (newBuildingName == "" || newBuildingName == "-") {
		return fmt.Errorf("branch_tag or building_name must be provided (at least one must not be empty)")
	}

	// 设置默认值
	if newBuildingName == "" {
		newBuildingName = "-"
	}

	// 更新 buildings 表
	var updatedBranchTag sql.NullString
	if newBranchTagValue == "" || newBranchTagValue == "-" {
		// branch_tag 为 "-" 或空时，更新为 NULL
		err = r.db.QueryRowContext(ctx,
			`UPDATE buildings 
			 SET building_name = $1, branch_tag = NULL, updated_at = CURRENT_TIMESTAMP
			 WHERE tenant_id = $2 AND building_id = $3
			 RETURNING branch_tag`,
			newBuildingName, tenantID, buildingID,
		).Scan(&updatedBranchTag)
	} else {
		err = r.db.QueryRowContext(ctx,
			`UPDATE buildings 
			 SET branch_tag = $1, building_name = $2, updated_at = CURRENT_TIMESTAMP
			 WHERE tenant_id = $3 AND building_id = $4
			 RETURNING branch_tag`,
			newBranchTagValue, newBuildingName, tenantID, buildingID,
		).Scan(&updatedBranchTag)
	}

	if err != nil {
		return fmt.Errorf("failed to update building: %w", err)
	}

	// 同步branch_tag变化到tags_catalog目录（替代trigger_sync_branch_tag）
	oldBranchTagValue := ""
	if oldBranchTag.Valid {
		oldBranchTagValue = oldBranchTag.String
	}
	if oldBranchTagValue == "-" {
		oldBranchTagValue = ""
	}
	if newBranchTagValue == "-" {
		newBranchTagValue = ""
	}
	if newBranchTagValue != oldBranchTagValue {
		// 如果新branch_tag不为空，添加到目录
		if newBranchTagValue != "" {
			_, err := r.db.ExecContext(ctx,
				`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
				tenantID, newBranchTagValue, "branch_tag",
			)
			if err != nil {
				fmt.Printf("[PostgresUnitsRepository] Failed to sync branch_tag to catalog: %v\n", err)
			}
		}
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

	// Handle branch_tag: if empty string, match NULL in database
	if filters.BranchTag != "" {
		if filters.BranchTag == "" {
			where = append(where, "u.branch_tag IS NULL")
		} else {
			addEq("u.branch_tag", filters.BranchTag)
		}
	}
	addEq("u.building", filters.Building)
	addEq("u.floor", filters.Floor)
	addEq("u.area_tag", filters.AreaTag)
	addEq("u.unit_number", filters.UnitNumber)
	addEq("u.unit_name", filters.UnitName)
	addEq("u.unit_type", filters.UnitType)

	// Search filter: 模糊搜索 unit_name, unit_number
	if filters.Search != "" {
		where = append(where, fmt.Sprintf("(u.unit_name ILIKE $%d OR u.unit_number ILIKE $%d)", argN, argN))
		args = append(args, "%"+filters.Search+"%")
		argN++
	}

	queryCount := "SELECT COUNT(*) FROM units u WHERE " + strings.Join(where, " AND ")
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
			u.branch_tag,
			u.unit_name,
			u.building,
			u.floor,
			u.area_tag,
			u.unit_number,
			CASE WHEN u.layout_config IS NULL THEN NULL ELSE u.layout_config::text END as layout_config,
			u.unit_type,
			u.is_public_space,
			u.is_multi_person_room,
			u.timezone,
			CASE WHEN u.groupList IS NULL THEN NULL ELSE u.groupList::text END as groupList,
			CASE WHEN u.userList IS NULL THEN NULL ELSE u.userList::text END as userList
		FROM units u
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY 
			-- First sort by floor (extract number from "1F", "2F", etc.)
			COALESCE((NULLIF(REGEXP_REPLACE(u.floor, '[^0-9]', '', 'g'), '')::int), 0),
			-- Then sort by unit_number (try numeric, fallback to string)
			CASE 
				WHEN u.unit_number ~ '^[0-9]+$' THEN u.unit_number::int
				ELSE 999999
			END,
			u.unit_number
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*domain.Unit, 0)
	for rows.Next() {
		var u domain.Unit
		var branchTag, areaTag, layoutConfig, groupList, userList sql.NullString
		if err := rows.Scan(
			&u.UnitID,
			&u.TenantID,
			&branchTag,
			&u.UnitName,
			&u.Building,
			&u.Floor,
			&areaTag,
			&u.UnitNumber,
			&layoutConfig,
			&u.UnitType,
			&u.IsPublicSpace,
			&u.IsMultiPersonRoom,
			&u.Timezone,
			&groupList,
			&userList,
		); err != nil {
			return nil, 0, err
		}
		u.BranchTag = branchTag
		u.AreaTag = areaTag
		u.LayoutConfig = layoutConfig
		u.GroupList = groupList
		u.UserList = userList
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
			u.branch_tag,
			u.unit_name,
			u.building,
			u.floor,
			u.area_tag,
			u.unit_number,
			CASE WHEN u.layout_config IS NULL THEN NULL ELSE u.layout_config::text END as layout_config,
			u.unit_type,
			u.is_public_space,
			u.is_multi_person_room,
			u.timezone,
			CASE WHEN u.groupList IS NULL THEN NULL ELSE u.groupList::text END as groupList,
			CASE WHEN u.userList IS NULL THEN NULL ELSE u.userList::text END as userList
		FROM units u
		WHERE u.tenant_id = $1 AND u.unit_id = $2
	`
	var u domain.Unit
	var branchTag, areaTag, layoutConfig, groupList, userList sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, unitID).Scan(
		&u.UnitID,
		&u.TenantID,
		&branchTag,
		&u.UnitName,
		&u.Building,
		&u.Floor,
		&areaTag,
		&u.UnitNumber,
		&layoutConfig,
		&u.UnitType,
		&u.IsPublicSpace,
		&u.IsMultiPersonRoom,
		&u.Timezone,
		&groupList,
		&userList,
	)
	if err != nil {
		return nil, err
	}
	u.BranchTag = branchTag
	u.AreaTag = areaTag
	u.LayoutConfig = layoutConfig
	u.GroupList = groupList
	u.UserList = userList
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
	if unit.UnitNumber == "" {
		return "", fmt.Errorf("unit_number is required")
	}
	if unit.UnitType == "" {
		return "", fmt.Errorf("unit_type is required")
	}
	if unit.Timezone == "" {
		return "", fmt.Errorf("timezone is required")
	}

	// 验证：branch_tag 和 building 不能同时为空
	branchTagValue := ""
	if unit.BranchTag.Valid {
		branchTagValue = unit.BranchTag.String
	}
	if (branchTagValue == "" || branchTagValue == "-") && (unit.Building == "" || unit.Building == "-") {
		return "", fmt.Errorf("branch_tag and building cannot both be empty (at least one must be provided)")
	}

	// 统一处理：空字符串''或"-"视为NULL，与display逻辑统一
	var branchTagValueSQL sql.NullString
	if branchTagValue != "" && branchTagValue != "-" {
		branchTagValueSQL = sql.NullString{String: branchTagValue, Valid: true}
	}

	// 验证 building 是否存在（如果提供）
	if unit.Building != "" && unit.Building != "-" {
		var exists bool
		err := r.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM buildings 
				WHERE tenant_id = $1 
				  AND building_name = $2 
				  AND COALESCE(branch_tag, '-') = COALESCE($3, '-')
			)`,
			tenantID, unit.Building, branchTagValueSQL,
		).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("failed to validate building: %w", err)
		}
		if !exists {
			return "", fmt.Errorf("building not found: branch_tag=%s, building_name=%s (unit must belong to an existing building)", branchTagValue, unit.Building)
		}
	}

	// 设置默认值
	if unit.Building == "" {
		unit.Building = "-"
	}
	if unit.Floor == "" {
		unit.Floor = "1F"
	}

	var areaTagSQL sql.NullString
	if unit.AreaTag.Valid && unit.AreaTag.String != "" {
		areaTagSQL = sql.NullString{String: unit.AreaTag.String, Valid: true}
	}
	var layoutConfigSQL sql.NullString
	if unit.LayoutConfig.Valid && unit.LayoutConfig.String != "" {
		layoutConfigSQL = sql.NullString{String: unit.LayoutConfig.String, Valid: true}
	}

	q := `
		INSERT INTO units (tenant_id, branch_tag, unit_name, building, floor, area_tag, unit_number, layout_config, unit_type, is_public_space, is_multi_person_room, timezone)
		VALUES ($1, $2, $3, COALESCE($4,'-'), COALESCE($5,'1F'), $6, $7, $8::jsonb, $9, COALESCE($10,false), COALESCE($11,false), $12)
		RETURNING unit_id::text
	`

	var unitID string
	if err := r.db.QueryRowContext(ctx, q,
		tenantID,
		branchTagValueSQL,
		unit.UnitName,
		unit.Building,
		unit.Floor,
		areaTagSQL,
		unit.UnitNumber,
		nullStringToAny(layoutConfigSQL),
		unit.UnitType,
		unit.IsPublicSpace,
		unit.IsMultiPersonRoom,
		unit.Timezone,
	).Scan(&unitID); err != nil {
		return "", err
	}

	// 同步branch_tag到tags_catalog目录（替代trigger_sync_branch_tag）
	if branchTagValueSQL.Valid {
		_, err := r.db.ExecContext(ctx,
			`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
			tenantID, branchTagValueSQL.String, "branch_tag",
		)
		if err != nil {
			fmt.Printf("[PostgresUnitsRepository] Failed to sync branch_tag to catalog: %v\n", err)
		}
	}

	// 同步area_tag到tags_catalog目录（替代trigger_sync_area_tag）
	if areaTagSQL.Valid {
		_, err := r.db.ExecContext(ctx,
			`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
			tenantID, areaTagSQL.String, "area_tag",
		)
		if err != nil {
			fmt.Printf("[PostgresUnitsRepository] Failed to sync area_tag to catalog: %v\n", err)
		}
	}

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

	// 先获取当前 unit 的信息（用于比较 branch_tag、area_tag、groupList）
	currentUnit, err := r.GetUnit(ctx, tenantID, unitID)
	if err != nil {
		return err
	}

	oldBranchTagValue := ""
	if currentUnit.BranchTag.Valid {
		oldBranchTagValue = currentUnit.BranchTag.String
	}
	oldAreaTagValue := ""
	if currentUnit.AreaTag.Valid {
		oldAreaTagValue = currentUnit.AreaTag.String
	}

	// 验证：branch_tag 和 building 不能同时为空
	branchTagValue := ""
	if unit.BranchTag.Valid {
		branchTagValue = unit.BranchTag.String
	}
	if (branchTagValue == "" || branchTagValue == "-") && (unit.Building == "" || unit.Building == "-") {
		return fmt.Errorf("branch_tag and building cannot both be empty (at least one must be provided)")
	}

	// 验证 building 是否存在（如果提供且改变）
	if unit.Building != "" && unit.Building != "-" && unit.Building != currentUnit.Building {
		var exists bool
		var branchTagValueSQL sql.NullString
		if branchTagValue != "" && branchTagValue != "-" {
			branchTagValueSQL = sql.NullString{String: branchTagValue, Valid: true}
		}
		err := r.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM buildings 
				WHERE tenant_id = $1 
				  AND building_name = $2 
				  AND COALESCE(branch_tag, '-') = COALESCE($3, '-')
			)`,
			tenantID, unit.Building, branchTagValueSQL,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to validate building: %w", err)
		}
		if !exists {
			return fmt.Errorf("building not found: branch_tag=%s, building_name=%s (unit must belong to an existing building)", branchTagValue, unit.Building)
		}
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

	// 统一处理 branch_tag：空字符串''或"-"视为NULL
	if unit.BranchTag.Valid {
		if unit.BranchTag.String == "" || unit.BranchTag.String == "-" {
			set = append(set, fmt.Sprintf("branch_tag = NULL"))
		} else {
			add("branch_tag", unit.BranchTag.String)
		}
	}
	if unit.UnitName != "" {
		add("unit_name", unit.UnitName)
	}
	if unit.Building != "" {
		add("building", unit.Building)
	}
	if unit.Floor != "" {
		add("floor", unit.Floor)
	}
	if unit.AreaTag.Valid {
		if unit.AreaTag.String == "" {
			set = append(set, fmt.Sprintf("area_tag = NULL"))
		} else {
			add("area_tag", unit.AreaTag.String)
		}
	}
	if unit.UnitNumber != "" {
		add("unit_number", unit.UnitNumber)
	}
	if unit.LayoutConfig.Valid && unit.LayoutConfig.String != "" {
		set = append(set, fmt.Sprintf("layout_config = $%d::jsonb", argN))
		args = append(args, unit.LayoutConfig.String)
		argN++
	}
	if unit.UnitType != "" {
		add("unit_type", unit.UnitType)
	}
	set = append(set, fmt.Sprintf("is_public_space = $%d", argN))
	args = append(args, unit.IsPublicSpace)
	argN++
	set = append(set, fmt.Sprintf("is_multi_person_room = $%d", argN))
	args = append(args, unit.IsMultiPersonRoom)
	argN++
	if unit.Timezone != "" {
		add("timezone", unit.Timezone)
	}
	if unit.GroupList.Valid {
		set = append(set, fmt.Sprintf("groupList = $%d::jsonb", argN))
		args = append(args, unit.GroupList.String)
		argN++
	}
	if unit.UserList.Valid {
		set = append(set, fmt.Sprintf("userList = $%d::jsonb", argN))
		args = append(args, unit.UserList.String)
		argN++
	}

	if len(set) == 0 {
		return nil
	}

	q := fmt.Sprintf("UPDATE units SET %s WHERE tenant_id = $1 AND unit_id = $2", strings.Join(set, ", "))
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return err
	}

	// 检查 groupList 是否变化（用于替代 trigger_sync_units_groupList_to_cards）
	if unit.GroupList.Valid {
		var oldGroupList sql.NullString
		err := r.db.QueryRowContext(ctx,
			`SELECT groupList FROM units WHERE tenant_id = $1 AND unit_id = $2`,
			tenantID, unitID,
		).Scan(&oldGroupList)
		if err == nil {
			var changed bool
			err = r.db.QueryRowContext(ctx,
				`SELECT COALESCE($1::jsonb, '[]'::jsonb) IS DISTINCT FROM COALESCE($2::jsonb, '[]'::jsonb)`,
				oldGroupList.String, unit.GroupList.String,
			).Scan(&changed)
			if err == nil && changed {
				err = r.syncUnitGroupListToCards(ctx, tenantID, unitID, unit.GroupList.String)
				if err != nil {
					fmt.Printf("[PostgresUnitsRepository] Failed to sync groupList to cards: %v\n", err)
				}
			}
		}
	}

	// 同步branch_tag变化到tags_catalog目录（替代trigger_sync_branch_tag）
	newBranchTagValue := branchTagValue
	if newBranchTagValue == "-" {
		newBranchTagValue = ""
	}
	if oldBranchTagValue == "-" {
		oldBranchTagValue = ""
	}
	if newBranchTagValue != oldBranchTagValue {
		if newBranchTagValue != "" {
			_, err := r.db.ExecContext(ctx,
				`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
				tenantID, newBranchTagValue, "branch_tag",
			)
			if err != nil {
				fmt.Printf("[PostgresUnitsRepository] Failed to sync branch_tag to catalog: %v\n", err)
			}
		}
	}

	// 同步area_tag变化到tags_catalog目录（替代trigger_sync_area_tag）
	newAreaTagValue := ""
	if unit.AreaTag.Valid {
		newAreaTagValue = unit.AreaTag.String
	}
	if newAreaTagValue != oldAreaTagValue {
		if newAreaTagValue != "" {
			_, err := r.db.ExecContext(ctx,
				`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
				tenantID, newAreaTagValue, "area_tag",
			)
			if err != nil {
				fmt.Printf("[PostgresUnitsRepository] Failed to sync area_tag to catalog: %v\n", err)
			}
		}
	}

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
func (r *PostgresUnitsRepository) ListRooms(ctx context.Context, tenantID, unitID string) ([]*domain.Room, error) {
	if tenantID == "" || unitID == "" {
		return []*domain.Room{}, nil
	}

	q := `
		SELECT 
			r.room_id::text,
			r.tenant_id::text,
			r.unit_id::text,
			r.room_name,
			CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r
		WHERE r.tenant_id = $1 AND r.unit_id = $2
		ORDER BY r.room_name
	`
	rows, err := r.db.QueryContext(ctx, q, tenantID, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]*domain.Room, 0)
	for rows.Next() {
		var room domain.Room
		var layoutConfig sql.NullString
		if err := rows.Scan(&room.RoomID, &room.TenantID, &room.UnitID, &room.RoomName, &layoutConfig); err != nil {
			return nil, err
		}
		room.LayoutConfig = layoutConfig
		rooms = append(rooms, &room)
	}
	return rooms, rows.Err()
}

// ListRoomsWithBeds: 查询 rooms 及其 beds
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) ListRoomsWithBeds(ctx context.Context, tenantID, unitID string) ([]*RoomWithBeds, error) {
	if tenantID == "" || unitID == "" {
		return []*RoomWithBeds{}, nil
	}

	// 查询 rooms
	rooms, err := r.ListRooms(ctx, tenantID, unitID)
	if err != nil {
		return nil, err
	}

	if len(rooms) == 0 {
		return []*RoomWithBeds{}, nil
	}

	// 查询 beds
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
			b.bed_name,
			b.bed_type,
			b.mattress_material,
			b.mattress_thickness
		FROM beds b
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
		var mattressMaterial, mattressThickness sql.NullString
		if err := brows.Scan(
			&bed.BedID,
			&bed.TenantID,
			&bed.RoomID,
			&bed.BedName,
			&bed.BedType,
			&mattressMaterial,
			&mattressThickness,
		); err != nil {
			return nil, err
		}
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
			r.room_name,
			CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r
		WHERE r.tenant_id = $1 AND r.room_id = $2
	`
	var room domain.Room
	var layoutConfig sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, roomID).Scan(
		&room.RoomID,
		&room.TenantID,
		&room.UnitID,
		&room.RoomName,
		&layoutConfig,
	)
	if err != nil {
		return nil, err
	}
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

// ListBeds: 查询 beds 列表
// 替代触发器：无（仅查询）
func (r *PostgresUnitsRepository) ListBeds(ctx context.Context, tenantID, roomID string) ([]*domain.Bed, error) {
	if tenantID == "" || roomID == "" {
		return []*domain.Bed{}, nil
	}

	q := `
		SELECT 
			b.bed_id::text,
			b.tenant_id::text,
			b.room_id::text,
			b.bed_name,
			b.bed_type,
			b.mattress_material,
			b.mattress_thickness
		FROM beds b
		WHERE b.tenant_id = $1 AND b.room_id = $2
		ORDER BY b.bed_name
	`
	rows, err := r.db.QueryContext(ctx, q, tenantID, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	beds := make([]*domain.Bed, 0)
	for rows.Next() {
		var bed domain.Bed
		var mattressMaterial, mattressThickness sql.NullString
		if err := rows.Scan(
			&bed.BedID,
			&bed.TenantID,
			&bed.RoomID,
			&bed.BedName,
			&bed.BedType,
			&mattressMaterial,
			&mattressThickness,
		); err != nil {
			return nil, err
		}
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
			b.bed_name,
			b.bed_type,
			b.mattress_material,
			b.mattress_thickness
		FROM beds b
		WHERE b.tenant_id = $1 AND b.bed_id = $2
	`
	var bed domain.Bed
	var mattressMaterial, mattressThickness sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, bedID).Scan(
		&bed.BedID,
		&bed.TenantID,
		&bed.RoomID,
		&bed.BedName,
		&bed.BedType,
		&mattressMaterial,
		&mattressThickness,
	)
	if err != nil {
		return nil, err
	}
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

	// 默认 bed_type = NonActiveBed
	if bed.BedType == "" {
		bed.BedType = "NonActiveBed"
	}

	var mattressMaterialSQL, mattressThicknessSQL sql.NullString
	if bed.MattressMaterial.Valid && bed.MattressMaterial.String != "" {
		mattressMaterialSQL = sql.NullString{String: bed.MattressMaterial.String, Valid: true}
	}
	if bed.MattressThickness.Valid && bed.MattressThickness.String != "" {
		mattressThicknessSQL = sql.NullString{String: bed.MattressThickness.String, Valid: true}
	}

	var bedID string
	q := `
		INSERT INTO beds (tenant_id, room_id, bed_name, bed_type, mattress_material, mattress_thickness)
		SELECT tenant_id, $1, $2, $3, $4, $5
		FROM rooms WHERE room_id = $1
		RETURNING bed_id::text
	`
	if err := r.db.QueryRowContext(ctx, q, roomID, bed.BedName, bed.BedType, mattressMaterialSQL, mattressThicknessSQL).Scan(&bedID); err != nil {
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
	if bed.BedType != "" {
		set = append(set, fmt.Sprintf("bed_type = $%d", argN))
		args = append(args, bed.BedType)
		argN++
	}
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
			argN++
		}
	}

	if len(set) == 0 {
		return nil
	}

	q := "UPDATE beds SET " + strings.Join(set, ", ") + " WHERE tenant_id = $1 AND bed_id = $2"
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return err
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

// syncUnitGroupListToCards 同步 units.groupList 到 cards.routing_alarm_tags（替代 trigger_sync_units_groupList_to_cards）
func (r *PostgresUnitsRepository) syncUnitGroupListToCards(ctx context.Context, tenantID, unitID string, groupList string) error {
	// 将 groupList (JSONB string) 转换为 VARCHAR[] 数组
	var groupListArray []string

	if groupList != "" {
		query := `
			SELECT ARRAY(
				SELECT jsonb_array_elements_text($1::jsonb)
			)::VARCHAR[]
		`
		err := r.db.QueryRowContext(ctx, query, groupList).Scan(&groupListArray)
		if err != nil {
			// 如果转换失败，设置为空数组
			groupListArray = []string{}
		}
	} else {
		groupListArray = []string{}
	}

	// 更新所有关联到该 unit 的 cards.routing_alarm_tags
	_, err := r.db.ExecContext(ctx,
		`UPDATE cards
		 SET routing_alarm_tags = $1
		 WHERE unit_id = $2
		   AND tenant_id = $3
		   AND (routing_alarm_tags IS DISTINCT FROM $1)`,
		pq.Array(groupListArray), unitID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update cards routing_alarm_tags: %w", err)
	}

	return nil
}
