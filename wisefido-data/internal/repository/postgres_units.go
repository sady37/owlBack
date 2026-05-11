package repository

// PostgresUnitsRepository — owl_v2 schema 版本（覆盖 buildings→sites / units / rooms / beds）。
//
// v2 schema 不向后兼容（IPv6 prefix 派生取代 UUID 多 FK）：
//   - buildings 表已删 → 并入 sites（site_slot = building<<4 | floor，IPv6 /64）
//   - units 表用 spatial_prefix /80 PK；父级 sites /64 通过 prefix-match + trigger 校验
//   - rooms 表用 spatial_prefix /88 PK；父级 unit /80
//   - beds 表用 spatial_prefix /96 PK；父级 room /88
//   - 删除 trigger 不级联，删 site/unit 之前必须先删下游
//
// 兼容策略：UnitsRepository interface 不变；domain.Building/Unit/Room/Bed 字段不动；
// 各 ID 字段（BuildingID / UnitID / RoomID / BedID）装对应层级的 IPv6 CIDR 字符串。
//
// v1 building 1:1 映射到 v2 site（每个 (branch, building_int, floor) 组合一个 site）：
//   - v1 用户输入 "Building A" 字符串 → v2 建一个 site（site_name='Building A'），
//     在该 branch 内分配 building 整数（顺序：第一个出现的 building_name 拿 0；上限 0..14）
//   - 同 building_name 不同 floor → v2 多个 site 共享 building 整数；UI 按 site_name 分组显示
//   - v1 floor 字符串（"1" "2" "1F" "G"）→ v2 floor INT（parseFloorToInt 解析）
//
// CreateBuilding 在 v2 = lookup-or-create site：相同 (branch, building_name, floor) 直接返回；
// 相同 (branch, building_name) 已用 building_int 时 floor 不同则新建一行复用该整数。
//
// 其它 v1→v2 简化点：
//   - v1 unit.layout_config jsonb / bed.mattress_*：v2 schema 删了，对外伪装空 NullString
//   - GetUnitAvailability / ListBedsWithResident：v2 residents 关系尚未 v2 化，保守值

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"

	"wisefido-data/internal/domain"
)

type PostgresUnitsRepository struct {
	db *sql.DB
}

func NewPostgresUnitsRepository(db *sql.DB) *PostgresUnitsRepository {
	return &PostgresUnitsRepository{db: db}
}

var _ UnitsRepository = (*PostgresUnitsRepository)(nil)

// =============================================================================
// helpers — IPv6 prefix 派生 / 解析
// =============================================================================

// parseFloorToInt 把 v1 floor 字符串转 1..14 INT（v2: floor 0 = unbound 哨兵不分配；越界/失败返回 1 兜底）。
func parseFloorToInt(s string) int {
	t := strings.TrimSpace(strings.ToUpper(s))
	if t == "" || t == "G" || t == "GROUND" {
		return 1 // G/Ground 也归 1 楼（v2 不再用 floor=0）
	}
	t = strings.TrimSuffix(t, "F")
	t = strings.TrimSuffix(t, "层")
	if n, err := strconv.Atoi(t); err == nil {
		if n < 1 || n > 14 {
			return 1
		}
		return n
	}
	return 1
}

// formatFloor INT 1..14 反向格式化成 v1 用的字符串（"<n>F"）。
func formatFloor(n int) string {
	if n < 1 {
		n = 1
	}
	return strconv.Itoa(n) + "F"
}

func deriveSiteCIDR(branchPrefixCIDR string, building, floor byte) (string, error) {
	out, err := zeroSuffix(branchPrefixCIDR, 7)
	if err != nil {
		return "", err
	}
	out[7] = (building&0x0F)<<4 | (floor & 0x0F)
	return out.String() + "/64", nil
}

func deriveUnitCIDR(sitePrefixCIDR string, unitSlot uint16) (string, error) {
	out, err := zeroSuffix(sitePrefixCIDR, 8)
	if err != nil {
		return "", err
	}
	out[8] = byte(unitSlot >> 8)
	out[9] = byte(unitSlot & 0xff)
	return out.String() + "/80", nil
}

func deriveRoomCIDR(unitPrefixCIDR string, roomSlot byte) (string, error) {
	out, err := zeroSuffix(unitPrefixCIDR, 10)
	if err != nil {
		return "", err
	}
	out[10] = roomSlot
	return out.String() + "/88", nil
}

func deriveBedCIDR(roomPrefixCIDR string, bedSlot byte) (string, error) {
	out, err := zeroSuffix(roomPrefixCIDR, 11)
	if err != nil {
		return "", err
	}
	out[11] = bedSlot
	return out.String() + "/96", nil
}

// nullStringToAny 把 sql.NullString 转 INSERT/UPDATE 用的 any（NULL 或字符串值）。
// 保留供其它 repo 使用（postgres_device_store 等历史依赖）。
func nullStringToAny(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}

// zeroSuffix 解析 CIDR 字符串，clear 从 byteIdx 起所有字节，返回 16 字节 net.IP。
func zeroSuffix(prefixCIDR string, byteIdx int) (net.IP, error) {
	addr := prefixCIDR
	if i := strings.IndexByte(addr, '/'); i >= 0 {
		addr = addr[:i]
	}
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("invalid prefix: %q", prefixCIDR)
	}
	v6 := ip.To16()
	if v6 == nil {
		return nil, fmt.Errorf("not IPv6: %q", prefixCIDR)
	}
	out := make(net.IP, 16)
	copy(out, v6)
	for i := byteIdx; i < 16; i++ {
		out[i] = 0
	}
	return out, nil
}

// =============================================================================
// Buildings (v2 sites 表) — v1 building 字符串名 → v2 site_name + 自动分配 building 整数
// =============================================================================

const buildingsSelectCols = `
		host(s.site_id) || '/64' AS building_id,
		network(set_masklen(s.site_id, 48))::text AS tenant_id,
		network(set_masklen(s.site_id, 56))::text AS branch_id,
		COALESCE(b.branch_name, '') AS branch_name,
		COALESCE(s.site_name, '') AS building_name,
		s.created_at, s.updated_at
`

const buildingsFromClause = `
	FROM sites s
	LEFT JOIN branches b ON b.branch_id = network(set_masklen(s.site_id, 56))
`

func scanBuilding(rs rowScanner) (*domain.Building, error) {
	var b domain.Building
	var tenantID, branchID, branchName, buildingName string
	var createdAt, updatedAt sql.NullTime
	if err := rs.Scan(&b.BuildingID, &tenantID, &branchID, &branchName, &buildingName, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan building: %w", err)
	}
	b.TenantID = tenantID
	b.BranchID = sql.NullString{String: branchID, Valid: branchID != ""}
	b.BranchName = sql.NullString{String: branchName, Valid: branchName != ""}
	b.BuildingName = buildingName
	b.CreatedAt = createdAt
	b.UpdatedAt = updatedAt
	return &b, nil
}

func (r *PostgresUnitsRepository) ListBuildings(ctx context.Context, tenantID string, branchID sql.NullString, branchName string) ([]*domain.Building, error) {
	where := []string{}
	args := []any{}
	argIdx := 1
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		where = append(where, fmt.Sprintf("s.site_id <<= $%d::INET", argIdx))
		args = append(args, tenantID)
		argIdx++
	}
	if branchID.Valid && branchID.String != "" && looksLikeINETPrefix(branchID.String) {
		where = append(where, fmt.Sprintf("s.site_id <<= $%d::INET", argIdx))
		args = append(args, branchID.String)
		argIdx++
	} else if branchName != "" {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM branches b2 WHERE b2.branch_id = network(set_masklen(s.site_id, 56)) AND b2.branch_name = $%d)", argIdx))
		args = append(args, branchName)
		argIdx++
	}
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}
	// v1 业务：每个 building 一行（不展开 floor）；v2 sites 表里 (branch, building_int) 同组的多 floor sites
	// 用 DISTINCT ON 去重，取 floor 最小的那行作 building_id 占位（floor=0 的 placeholder 通常存在）。
	q := `SELECT DISTINCT ON (network(set_masklen(s.site_id, 56)), s.building) ` +
		buildingsSelectCols + buildingsFromClause + whereClause +
		` ORDER BY network(set_masklen(s.site_id, 56)), s.building, s.floor, s.site_id`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list buildings: %w", err)
	}
	defer rows.Close()
	out := []*domain.Building{}
	for rows.Next() {
		b, err := scanBuilding(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *PostgresUnitsRepository) GetBuilding(ctx context.Context, tenantID, buildingID string) (*domain.Building, error) {
	if buildingID == "" || !looksLikeINETPrefix(buildingID) {
		return nil, fmt.Errorf("invalid building_id")
	}
	q := `SELECT ` + buildingsSelectCols + buildingsFromClause + ` WHERE s.site_id = $1::INET`
	rows, err := r.db.QueryContext(ctx, q, buildingID)
	if err != nil {
		return nil, fmt.Errorf("get building: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("building not found: %s", buildingID)
	}
	return scanBuilding(rows)
}

func (r *PostgresUnitsRepository) GetBuildingUnits(ctx context.Context, tenantID, buildingID string) ([]BuildingUnitInfo, error) {
	if buildingID == "" || !looksLikeINETPrefix(buildingID) {
		return nil, nil
	}
	// 跨 floor 拿同 (branch, building_int) 下的所有 units
	rows, err := r.db.QueryContext(ctx, `
		SELECT host(u.unit_id) || '/80', u.unit_name, s.floor
		  FROM units u
		  JOIN sites s ON s.site_id = network(set_masklen(u.unit_id, 64))
		 WHERE u.unit_id <<= network(set_masklen($1::INET, 56))
		   AND s.building = (SELECT building FROM sites WHERE site_id = $1::INET)
		 ORDER BY s.floor, u.unit_slot`, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BuildingUnitInfo{}
	for rows.Next() {
		var unitID, unitName string
		var floor int
		if err := rows.Scan(&unitID, &unitName, &floor); err != nil {
			return nil, err
		}
		out = append(out, BuildingUnitInfo{UnitID: unitID, UnitName: unitName, Floor: sql.NullString{String: formatFloor(floor), Valid: true}})
	}
	return out, rows.Err()
}

// FindBuildingByBranchAndName 按 (branch, building_name) lookup 第一个匹配 site；空字符串/不存在返回 ""。
func (r *PostgresUnitsRepository) FindBuildingByBranchAndName(ctx context.Context, tenantID, branchID, buildingName string) (string, error) {
	if branchID == "" || !looksLikeINETPrefix(branchID) || buildingName == "" {
		return "", nil
	}
	var prefix string
	err := r.db.QueryRowContext(ctx, `
		SELECT host(site_id) || '/64' FROM sites
		 WHERE site_id <<= $1::INET AND LOWER(COALESCE(site_name,'')) = LOWER($2)
		 ORDER BY building, floor LIMIT 1`, branchID, buildingName).Scan(&prefix)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find building: %w", err)
	}
	return prefix, nil
}

// CreateBuilding lookup-or-create site：
//   - 相同 (branch, building_name) 已用 building 整数 → 复用；floor 默认 0
//   - 否则在 branch 内分配 building 整数（MAX+1，上限 14）
//   - 相同 (branch, building, floor) 已存在 → idempotent return
func (r *PostgresUnitsRepository) CreateBuilding(ctx context.Context, tenantID string, building *domain.Building) (string, error) {
	if building == nil || !building.BranchID.Valid || !looksLikeINETPrefix(building.BranchID.String) {
		return "", fmt.Errorf("branch_id required")
	}
	if building.BuildingName == "" {
		return "", fmt.Errorf("building_name required")
	}
	branchPrefix := building.BranchID.String

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "owl_v2.sites.alloc:"+branchPrefix); err != nil {
		return "", fmt.Errorf("lock: %w", err)
	}

	// 默认 floor=1（v1 业务 CreateBuilding 不带 floor 输入；后续 CreateUnit 时若需新 floor 会补建 site）
	// floor 1..14：0 = unbound 哨兵，0xF wildcard 保留
	floorInt := 1

	// 1. 已有同名 building 整数？
	var buildingInt int
	err = tx.QueryRowContext(ctx, `
		SELECT DISTINCT building FROM sites
		 WHERE site_id <<= $1::INET AND LOWER(COALESCE(site_name,'')) = LOWER($2)
		 LIMIT 1`, branchPrefix, building.BuildingName).Scan(&buildingInt)
	if err == sql.ErrNoRows {
		// 分配下一个 building 整数 — 从 1 起（0 = unbound 哨兵）
		if err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(building), 0) + 1 FROM sites WHERE site_id <<= $1::INET`, branchPrefix).Scan(&buildingInt); err != nil {
			return "", fmt.Errorf("allocate building int: %w", err)
		}
		if buildingInt < 1 || buildingInt > 14 {
			return "", fmt.Errorf("building slot exhausted: %d (valid 1..14)", buildingInt)
		}
	} else if err != nil {
		return "", fmt.Errorf("lookup building int: %w", err)
	}

	// 2. (branch, building, floor) idempotent check
	var existingPrefix string
	err = tx.QueryRowContext(ctx, `
		SELECT host(site_id) || '/64' FROM sites
		 WHERE site_id <<= $1::INET AND building = $2 AND floor = $3
		 LIMIT 1`, branchPrefix, buildingInt, floorInt).Scan(&existingPrefix)
	if err == nil {
		return existingPrefix, tx.Commit()
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("check existing site: %w", err)
	}

	// 3. 派生并 INSERT
	sitePrefix, err := deriveSiteCIDR(branchPrefix, byte(buildingInt), byte(floorInt))
	if err != nil {
		return "", err
	}
	siteSlot := (buildingInt << 4) | floorInt
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sites (site_id, site_slot, building, floor, site_name)
		VALUES ($1::INET, $2, $3, $4, $5)`,
		sitePrefix, siteSlot, buildingInt, floorInt, building.BuildingName); err != nil {
		return "", fmt.Errorf("insert site: %w", err)
	}
	return sitePrefix, tx.Commit()
}

// UpdateBuilding 重命名 building：v2 一个逻辑 building = 同 (branch, building_int) 的全部 floor sites，
// 必须把组内每行的 site_name 都更新，否则其它 floor 的 site_name 仍是旧值，前端会看到不一致。
func (r *PostgresUnitsRepository) UpdateBuilding(ctx context.Context, tenantID, buildingID string, building *domain.Building) error {
	if buildingID == "" || !looksLikeINETPrefix(buildingID) {
		return fmt.Errorf("invalid building_id")
	}
	if building == nil || building.BuildingName == "" {
		return fmt.Errorf("building_name required")
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE sites SET site_name = $2, updated_at = NOW()
		 WHERE site_id <<= network(set_masklen($1::INET, 56))
		   AND building = (SELECT building FROM sites WHERE site_id = $1::INET)`,
		buildingID, building.BuildingName)
	if err != nil {
		return fmt.Errorf("update site: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("building not found: %s", buildingID)
	}
	return nil
}

// DeleteBuilding 删除整栋 building：v2 一个逻辑 building = 同 (branch, building_int) 的全部 floor sites。
//   - 若组内任意 site 有 units → 拒绝（与 branch 删除约定一致）；前端通常已 pre-check
//   - 否则 DELETE 该组所有 sites（不只是用户点击的某一 floor）
func (r *PostgresUnitsRepository) DeleteBuilding(ctx context.Context, tenantID, buildingID string) error {
	if buildingID == "" || !looksLikeINETPrefix(buildingID) {
		return fmt.Errorf("invalid building_id")
	}
	var hasUnits bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM units u
		   WHERE u.unit_id <<= network(set_masklen($1::INET, 56))
		     AND (SELECT s.building FROM sites s WHERE s.site_id = network(set_masklen(u.unit_id, 64)))
		         = (SELECT building FROM sites WHERE site_id = $1::INET)
		)`, buildingID).Scan(&hasUnits); err != nil {
		return fmt.Errorf("check building empty: %w", err)
	}
	if hasUnits {
		return fmt.Errorf("building has units: delete units first")
	}
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM sites
		 WHERE site_id <<= network(set_masklen($1::INET, 56))
		   AND building = (SELECT building FROM sites WHERE site_id = $1::INET)`, buildingID)
	if err != nil {
		return fmt.Errorf("delete sites: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("building not found: %s", buildingID)
	}
	return nil
}

// =============================================================================
// Units (v2 units 表)
// =============================================================================

// building_id 必须返回与 ListBuildings DISTINCT ON 同样的 winner（最小 floor 的 /64），
// 否则前端 unit.building_id ≠ building.building_id，filter 会把同 building 跨 floor 的 unit 全部隐藏。
const unitsSelectCols = `
		host(u.unit_id) || '/80' AS unit_id,
		network(set_masklen(u.unit_id, 48))::text AS tenant_id,
		network(set_masklen(u.unit_id, 56))::text AS branch_id,
		COALESCE(b.branch_name, '') AS branch_name,
		u.unit_name,
		(SELECT host(s2.site_id) || '/64'
		   FROM sites s2
		  WHERE s2.site_id <<= network(set_masklen(u.unit_id, 56))
		    AND s2.building = s.building
		  ORDER BY s2.floor LIMIT 1) AS building_id,
		COALESCE(s.site_name, '') AS building_name,
		s.floor AS floor_int,
		u.unit_property,
		u.unit_type,
		COALESCE(u.timezone, 'UTC') AS timezone
`

// JOIN 用 network(set_masklen(...)) 而非 set_masklen — 后者不清 host bits 会让 INET 比较失败。
const unitsFromClause = `
	FROM units u
	JOIN sites s    ON s.site_id = network(set_masklen(u.unit_id, 64))
	JOIN branches b ON b.branch_id = network(set_masklen(u.unit_id, 56))
`

func scanUnit(rs rowScanner) (*domain.Unit, error) {
	var u domain.Unit
	var branchID, branchName, buildingID, buildingName string
	var floorInt int
	if err := rs.Scan(&u.UnitID, &u.TenantID, &branchID, &branchName, &u.UnitName,
		&buildingID, &buildingName, &floorInt,
		&u.UnitProperty, &u.UnitType, &u.Timezone); err != nil {
		return nil, fmt.Errorf("scan unit: %w", err)
	}
	u.BranchID = sql.NullString{String: branchID, Valid: branchID != ""}
	u.BranchName = sql.NullString{String: branchName, Valid: branchName != ""}
	u.BuildingID = sql.NullString{String: buildingID, Valid: buildingID != ""}
	u.BuildingName = sql.NullString{String: buildingName, Valid: buildingName != ""}
	u.Floor = sql.NullString{String: formatFloor(floorInt), Valid: true}
	u.LayoutConfig = sql.NullString{} // v2 schema 无该列
	return &u, nil
}

func (r *PostgresUnitsRepository) ListUnits(ctx context.Context, tenantID string, filters UnitFilters, page, size int) ([]*domain.Unit, int, error) {
	where := []string{}
	args := []any{}
	argIdx := 1

	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		where = append(where, fmt.Sprintf("u.unit_id <<= $%d::INET", argIdx))
		args = append(args, tenantID)
		argIdx++
	}
	if filters.BranchID != "" && looksLikeINETPrefix(filters.BranchID) {
		where = append(where, fmt.Sprintf("u.unit_id <<= $%d::INET", argIdx))
		args = append(args, filters.BranchID)
		argIdx++
	} else if len(filters.BranchIDs) > 0 {
		ph := make([]string, 0, len(filters.BranchIDs))
		for _, bid := range filters.BranchIDs {
			if !looksLikeINETPrefix(bid) {
				continue
			}
			ph = append(ph, fmt.Sprintf("u.unit_id <<= $%d::INET", argIdx))
			args = append(args, bid)
			argIdx++
		}
		if len(ph) > 0 {
			where = append(where, "("+strings.Join(ph, " OR ")+")")
		}
	}
	if filters.BuildingID != "" && looksLikeINETPrefix(filters.BuildingID) {
		// v1 业务：按 building_id 过滤 = 该 building 的**全部楼层**的 units（不限单 floor）。
		// v2 building_id 是某 floor 的 site /64 占位 prefix；跨 floor 范围 = (branch, building_int) 组。
		where = append(where, fmt.Sprintf(
			"u.unit_id <<= network(set_masklen($%d::INET, 56)) AND s.building = (SELECT building FROM sites WHERE site_id = $%d::INET)",
			argIdx, argIdx))
		args = append(args, filters.BuildingID)
		argIdx++
	} else if filters.Building != "" {
		where = append(where, fmt.Sprintf("LOWER(COALESCE(s.site_name,'')) = LOWER($%d)", argIdx))
		args = append(args, filters.Building)
		argIdx++
	}
	if filters.Floor != "" {
		where = append(where, fmt.Sprintf("s.floor = $%d", argIdx))
		args = append(args, parseFloorToInt(filters.Floor))
		argIdx++
	}
	if filters.UnitName != "" {
		where = append(where, fmt.Sprintf("LOWER(u.unit_name) = LOWER($%d)", argIdx))
		args = append(args, filters.UnitName)
		argIdx++
	}
	if filters.UnitType != "" {
		where = append(where, fmt.Sprintf("COALESCE(u.unit_type,'') = $%d", argIdx))
		args = append(args, filters.UnitType)
		argIdx++
	}
	if filters.Search != "" {
		pat := "%" + strings.ToLower(filters.Search) + "%"
		where = append(where, fmt.Sprintf("(LOWER(u.unit_name) LIKE $%d OR LOWER(COALESCE(s.site_name,'')) LIKE $%d)", argIdx, argIdx))
		args = append(args, pat)
		argIdx++
	}
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) `+unitsFromClause+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count units: %w", err)
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	q := `SELECT ` + unitsSelectCols + unitsFromClause + whereClause +
		fmt.Sprintf(` ORDER BY s.building, s.floor, u.unit_slot LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list units: %w", err)
	}
	defer rows.Close()
	out := []*domain.Unit{}
	for rows.Next() {
		u, err := scanUnit(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (r *PostgresUnitsRepository) GetUnit(ctx context.Context, tenantID, unitID string) (*domain.Unit, error) {
	if unitID == "" || !looksLikeINETPrefix(unitID) {
		return nil, sql.ErrNoRows
	}
	q := `SELECT ` + unitsSelectCols + unitsFromClause + ` WHERE u.unit_id = $1::INET`
	rows, err := r.db.QueryContext(ctx, q, unitID)
	if err != nil {
		return nil, fmt.Errorf("get unit: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return scanUnit(rows)
}

// CreateUnit 在 site (BuildingID) + Floor 下分配 unit_slot 派生 /80。
//
// v1 业务流程：UnitList.vue 创建 unit 时携带 building_id（由前面 CreateBuilding 取得，
// 默认 floor=0 的那个 site）+ floor 字符串。
//
// 这里若 BuildingID 给定 site /64 但 floor 不一致：
//   - 解析出 (branch, building_int) → 在该 (branch, building_int, target_floor) lookup-or-create 真正的 site
//   - 然后在那个 site 下分配 unit_slot
//
// 若 BuildingID 没给但 BranchID 给了（v1 home care 场景）：自动 building_int=0, floor=parseFloor 建 site
func (r *PostgresUnitsRepository) CreateUnit(ctx context.Context, tenantID string, unit *domain.Unit) (string, error) {
	if unit == nil || unit.UnitName == "" {
		return "", fmt.Errorf("unit_name required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	floorInt := parseFloorToInt(unit.Floor.String)

	// 派生 sitePrefix：根据输入决定 (branch, building_int, floor) 三元组
	var sitePrefix string
	if unit.BuildingID.Valid && looksLikeINETPrefix(unit.BuildingID.String) {
		// 从已有 site /64 拿出 (branch, building_int)，target_floor 走 unit.Floor 输入
		var branchPrefix string
		var buildingInt int
		if err := tx.QueryRowContext(ctx, `
			SELECT network(set_masklen(site_id, 56))::text, building
			  FROM sites WHERE site_id = $1::INET`, unit.BuildingID.String).Scan(&branchPrefix, &buildingInt); err != nil {
			return "", fmt.Errorf("resolve building: %w", err)
		}
		// lookup-or-create (branch, building_int, floor)
		err := tx.QueryRowContext(ctx, `
			SELECT host(site_id) || '/64' FROM sites
			 WHERE site_id <<= $1::INET AND building = $2 AND floor = $3 LIMIT 1`,
			branchPrefix, buildingInt, floorInt).Scan(&sitePrefix)
		if err == sql.ErrNoRows {
			derived, derr := deriveSiteCIDR(branchPrefix, byte(buildingInt), byte(floorInt))
			if derr != nil {
				return "", derr
			}
			siteSlot := (buildingInt << 4) | floorInt
			// 拿原 building 的 site_name 用作新 floor site 的 name
			var siteName string
			_ = tx.QueryRowContext(ctx, `SELECT COALESCE(site_name,'') FROM sites WHERE site_id = $1::INET`, unit.BuildingID.String).Scan(&siteName)
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO sites (site_id, site_slot, building, floor, site_name)
				VALUES ($1::INET, $2, $3, $4, $5)`,
				derived, siteSlot, buildingInt, floorInt, siteName); err != nil {
				return "", fmt.Errorf("auto-create site for new floor: %w", err)
			}
			sitePrefix = derived
		} else if err != nil {
			return "", fmt.Errorf("lookup site: %w", err)
		}
	} else if unit.BranchID.Valid && looksLikeINETPrefix(unit.BranchID.String) {
		// 无 BuildingID：building=1 默认（home care 单楼宇场景；v2 不再用 building=0/floor=0）
		const defaultBuilding = 1
		err := tx.QueryRowContext(ctx, `
			SELECT host(site_id) || '/64' FROM sites
			 WHERE site_id <<= $1::INET AND building = $2 AND floor = $3 LIMIT 1`,
			unit.BranchID.String, defaultBuilding, floorInt).Scan(&sitePrefix)
		if err == sql.ErrNoRows {
			derived, derr := deriveSiteCIDR(unit.BranchID.String, byte(defaultBuilding), byte(floorInt))
			if derr != nil {
				return "", derr
			}
			siteSlot := (defaultBuilding << 4) | floorInt
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO sites (site_id, site_slot, building, floor, site_name)
				VALUES ($1::INET, $2, $3, $4, '')`, derived, siteSlot, defaultBuilding, floorInt); err != nil {
				return "", fmt.Errorf("auto-create default site: %w", err)
			}
			sitePrefix = derived
		} else if err != nil {
			return "", fmt.Errorf("lookup default site: %w", err)
		}
	} else {
		return "", fmt.Errorf("building_id or branch_id required")
	}

	// 同 site 内分配 unit_slot
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "owl_v2.units.alloc:"+sitePrefix); err != nil {
		return "", fmt.Errorf("lock: %w", err)
	}
	// unit_slot 从 1 起（slot 0 = unbound 哨兵；0xFFFF wildcard 保留）
	var nextSlot int
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(unit_slot), 0) + 1 FROM units WHERE unit_id <<= $1::INET`, sitePrefix).Scan(&nextSlot); err != nil {
		return "", fmt.Errorf("compute unit_slot: %w", err)
	}
	if nextSlot < 1 || nextSlot > 65534 {
		return "", fmt.Errorf("unit_slot exhausted: %d (valid 1..65534)", nextSlot)
	}
	unitPrefix, err := deriveUnitCIDR(sitePrefix, uint16(nextSlot))
	if err != nil {
		return "", err
	}
	// v2: unit_property + unit_type 双维度
	// 默认 Facility(1) + share(2)；Home → 强制 type=0
	unitProperty := unit.UnitProperty
	unitType := unit.UnitType
	if unitProperty == 0 {
		unitType = 0 // Home 强制 unknown
	} else {
		// Facility 时 type 必须 ∈ {1,2,3}，否则默认 share(2)
		if unitType < 1 || unitType > 3 {
			unitType = 2
		}
	}
	tz := unit.Timezone
	if tz == "" {
		tz = "UTC"
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO units (unit_id, unit_slot, unit_name, unit_property, unit_type, timezone)
		VALUES ($1::INET, $2, $3, $4, $5, $6)`,
		unitPrefix, nextSlot, unit.UnitName, unitProperty, unitType, tz); err != nil {
		return "", fmt.Errorf("insert unit: %w", err)
	}
	// 清空 floor 占位 site：同 (branch, building_int) 内任何 0-unit 的 site 都删，但保留至少 1 个（building placeholder）
	if err := cleanupEmptyFloorSites(ctx, tx, sitePrefix); err != nil {
		return "", fmt.Errorf("cleanup empty floor sites: %w", err)
	}
	return unitPrefix, tx.Commit()
}

// cleanupEmptyFloorSites 删除同 (branch, building_int) 内 unit 数=0 的 site rows，但**总保留至少 1 个**作 building placeholder。
// 用于 CreateUnit 后清掉 floor=0 占位、DeleteUnit 后清掉空了的 floor。
// referencePrefix 必须是该 building 的某个 site /64（用来反推 (branch, building_int)）。
func cleanupEmptyFloorSites(ctx context.Context, tx *sql.Tx, referencePrefix string) error {
	_, err := tx.ExecContext(ctx, `
		WITH grp AS (
		  SELECT network(set_masklen(site_id, 56)) AS branch_prefix, building
		    FROM sites WHERE site_id = $1::INET
		),
		all_sites AS (
		  SELECT s.site_id
		    FROM sites s, grp
		   WHERE s.site_id <<= grp.branch_prefix AND s.building = grp.building
		),
		empty_sites AS (
		  SELECT site_id FROM all_sites
		   WHERE NOT EXISTS (SELECT 1 FROM units u WHERE u.unit_id <<= all_sites.site_id)
		)
		DELETE FROM sites
		 WHERE site_id IN (SELECT site_id FROM empty_sites)
		   AND (SELECT COUNT(*) FROM all_sites) - (SELECT COUNT(*) FROM empty_sites) >= 1`,
		referencePrefix)
	return err
}

func (r *PostgresUnitsRepository) UpdateUnit(ctx context.Context, tenantID, unitID string, unit *domain.Unit) error {
	if unitID == "" || !looksLikeINETPrefix(unitID) {
		return fmt.Errorf("invalid unit_id")
	}
	if unit == nil {
		return fmt.Errorf("unit required")
	}
	updates := []string{}
	args := []any{unitID}
	argIdx := 2
	if unit.UnitName != "" {
		updates = append(updates, fmt.Sprintf("unit_name = $%d", argIdx))
		args = append(args, unit.UnitName)
		argIdx++
	}
	// v2: unit_property + unit_type 双维度，整体更新（service 层已校验配对合法）
	updates = append(updates, fmt.Sprintf("unit_property = $%d", argIdx))
	args = append(args, unit.UnitProperty)
	argIdx++
	updates = append(updates, fmt.Sprintf("unit_type = $%d", argIdx))
	args = append(args, unit.UnitType)
	argIdx++
	if unit.Timezone != "" {
		updates = append(updates, fmt.Sprintf("timezone = $%d", argIdx))
		args = append(args, unit.Timezone)
		argIdx++
	}
	updates = append(updates, "updated_at = NOW()")

	q := `UPDATE units SET ` + strings.Join(updates, ", ") + ` WHERE unit_id = $1::INET`
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("update unit: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("unit not found: %s", unitID)
	}
	return nil
}

func (r *PostgresUnitsRepository) DeleteUnit(ctx context.Context, tenantID, unitID string) error {
	if unitID == "" || !looksLikeINETPrefix(unitID) {
		return fmt.Errorf("invalid unit_id")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	// 先记下该 unit 所属 site /64（删完 unit 后用来反查同 building 组）
	var siteRef string
	if err := tx.QueryRowContext(ctx,
		`SELECT host(network(set_masklen($1::INET, 64))) || '/64'`, unitID).Scan(&siteRef); err != nil {
		return fmt.Errorf("derive site ref: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM units WHERE unit_id = $1::INET`, unitID)
	if err != nil {
		return fmt.Errorf("delete unit (likely has rooms): %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("unit not found: %s", unitID)
	}
	// 删完 unit 后：清除空的 floor sites（保留至少 1 个 placeholder）
	if err := cleanupEmptyFloorSites(ctx, tx, siteRef); err != nil {
		return fmt.Errorf("cleanup empty floor sites: %w", err)
	}
	return tx.Commit()
}

// GetUnitAvailability v2 简化：residents 表 v2 化未完成；保守返回 (true, false) 让 UI 显示可用绿色。
func (r *PostgresUnitsRepository) GetUnitAvailability(ctx context.Context, tenantID string, unitIDs []string) (map[string]bool, map[string]bool, error) {
	hasAvail := make(map[string]bool, len(unitIDs))
	isBound := make(map[string]bool, len(unitIDs))
	for _, id := range unitIDs {
		hasAvail[id] = true
		isBound[id] = false
	}
	return hasAvail, isBound, nil
}

// =============================================================================
// Rooms (v2 rooms 表)
// =============================================================================

const roomsSelectCols = `
		host(r.room_id) || '/88' AS room_id,
		network(set_masklen(r.room_id, 48))::text AS tenant_id,
		host(network(set_masklen(r.room_id, 80))) || '/80' AS unit_id,
		COALESCE(u.unit_name, '') AS unit_name,
		s.floor AS floor_int,
		r.room_name
`

const roomsFromClause = `
	FROM rooms r
	JOIN units u ON u.unit_id = network(set_masklen(r.room_id, 80))
	JOIN sites s ON s.site_id = network(set_masklen(r.room_id, 64))
`

func scanRoom(rs rowScanner) (*domain.Room, error) {
	var rm domain.Room
	var unitName string
	var floorInt int
	if err := rs.Scan(&rm.RoomID, &rm.TenantID, &rm.UnitID, &unitName, &floorInt, &rm.RoomName); err != nil {
		return nil, fmt.Errorf("scan room: %w", err)
	}
	rm.UnitName = sql.NullString{String: unitName, Valid: unitName != ""}
	rm.Floor = sql.NullString{String: formatFloor(floorInt), Valid: true}
	rm.LayoutConfig = sql.NullString{} // v2 无该列
	return &rm, nil
}

func (r *PostgresUnitsRepository) ListRooms(ctx context.Context, tenantID, unitID string, search string) ([]*domain.Room, error) {
	if unitID == "" || !looksLikeINETPrefix(unitID) {
		return nil, nil
	}
	where := []string{"r.room_id <<= $1::INET"}
	args := []any{unitID}
	argIdx := 2
	if search != "" {
		where = append(where, fmt.Sprintf("LOWER(r.room_name) LIKE $%d", argIdx))
		args = append(args, "%"+strings.ToLower(search)+"%")
		argIdx++
	}
	q := `SELECT ` + roomsSelectCols + roomsFromClause + ` WHERE ` + strings.Join(where, " AND ") + ` ORDER BY r.room_slot`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()
	out := []*domain.Room{}
	for rows.Next() {
		rm, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

func (r *PostgresUnitsRepository) GetRoom(ctx context.Context, tenantID, roomID string) (*domain.Room, error) {
	if roomID == "" || !looksLikeINETPrefix(roomID) {
		return nil, sql.ErrNoRows
	}
	q := `SELECT ` + roomsSelectCols + roomsFromClause + ` WHERE r.room_id = $1::INET`
	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return scanRoom(rows)
}

func (r *PostgresUnitsRepository) CreateRoom(ctx context.Context, tenantID, unitID string, room *domain.Room) (string, error) {
	if unitID == "" || !looksLikeINETPrefix(unitID) {
		return "", fmt.Errorf("invalid unit_id")
	}
	if room == nil || room.RoomName == "" {
		return "", fmt.Errorf("room_name required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "owl_v2.rooms.alloc:"+unitID); err != nil {
		return "", fmt.Errorf("lock: %w", err)
	}
	// room_slot 从 1 起（slot 0 = unbound 哨兵；0xFF wildcard 保留）
	var nextSlot int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(room_slot), 0) + 1 FROM rooms WHERE room_id <<= $1::INET`, unitID).Scan(&nextSlot); err != nil {
		return "", fmt.Errorf("compute room_slot: %w", err)
	}
	if nextSlot < 1 || nextSlot > 254 {
		return "", fmt.Errorf("room_slot exhausted: %d (valid 1..254)", nextSlot)
	}
	roomPrefix, err := deriveRoomCIDR(unitID, byte(nextSlot))
	if err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO rooms (room_id, room_slot, room_name)
		VALUES ($1::INET, $2, $3)`, roomPrefix, nextSlot, room.RoomName); err != nil {
		return "", fmt.Errorf("insert room: %w", err)
	}
	return roomPrefix, tx.Commit()
}

func (r *PostgresUnitsRepository) UpdateRoom(ctx context.Context, tenantID, roomID string, room *domain.Room) error {
	if roomID == "" || !looksLikeINETPrefix(roomID) {
		return fmt.Errorf("invalid room_id")
	}
	if room == nil || room.RoomName == "" {
		return fmt.Errorf("room_name required")
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE rooms SET room_name = $2, updated_at = NOW() WHERE room_id = $1::INET`, roomID, room.RoomName)
	if err != nil {
		return fmt.Errorf("update room: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("room not found: %s", roomID)
	}
	return nil
}

func (r *PostgresUnitsRepository) DeleteRoom(ctx context.Context, tenantID, roomID string) error {
	if roomID == "" || !looksLikeINETPrefix(roomID) {
		return fmt.Errorf("invalid room_id")
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM rooms WHERE room_id = $1::INET`, roomID)
	if err != nil {
		return fmt.Errorf("delete room (likely has beds): %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("room not found: %s", roomID)
	}
	return nil
}

func (r *PostgresUnitsRepository) ListRoomsByUnitIDs(ctx context.Context, tenantID string, unitIDs []string) ([]*domain.Room, error) {
	if len(unitIDs) == 0 {
		return nil, nil
	}
	args := make([]any, 0, len(unitIDs))
	conds := make([]string, 0, len(unitIDs))
	argIdx := 1
	for _, uid := range unitIDs {
		if !looksLikeINETPrefix(uid) {
			continue
		}
		conds = append(conds, fmt.Sprintf("r.room_id <<= $%d::INET", argIdx))
		args = append(args, uid)
		argIdx++
	}
	if len(conds) == 0 {
		return nil, nil
	}
	q := `SELECT ` + roomsSelectCols + roomsFromClause + ` WHERE (` + strings.Join(conds, " OR ") + `) ORDER BY r.room_id, r.room_slot`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list rooms by unit ids: %w", err)
	}
	defer rows.Close()
	out := []*domain.Room{}
	for rows.Next() {
		rm, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

// ListRoomsWithBeds 列 unit 下所有 room 携带 beds（v2 简化：bed 不带 resident 占用筛选）。
func (r *PostgresUnitsRepository) ListRoomsWithBeds(ctx context.Context, tenantID, unitID, search, residentID string) ([]*RoomWithBeds, error) {
	rooms, err := r.ListRooms(ctx, tenantID, unitID, search)
	if err != nil {
		return nil, err
	}
	out := make([]*RoomWithBeds, 0, len(rooms))
	for _, rm := range rooms {
		beds, err := r.ListBeds(ctx, tenantID, rm.RoomID, "")
		if err != nil {
			return nil, err
		}
		out = append(out, &RoomWithBeds{Room: rm, Beds: beds})
	}
	return out, nil
}

// ListRoomsByBranch 按 branch 列出 room 携带 unit + 占用信息（v2 简化：is_full/is_bound 暂返回 false）。
func (r *PostgresUnitsRepository) ListRoomsByBranch(ctx context.Context, tenantID, branchID string) ([]*RoomWithAvailability, error) {
	if branchID == "" || !looksLikeINETPrefix(branchID) {
		return nil, nil
	}
	// v2: unit_type 派生为 v1 字符串标签（Home / Facility），FacilityType 由 unit_type 子枚举派生
	q := `
		SELECT
		  host(r.room_id) || '/88' AS room_id,
		  network(set_masklen(r.room_id, 48))::text AS tenant_id,
		  host(network(set_masklen(r.room_id, 80))) || '/80' AS unit_id,
		  u.unit_name,
		  COALESCE(s.site_name, '') AS building_name,
		  s.floor AS floor_int,
		  r.room_name,
		  CASE u.unit_property WHEN 0 THEN 'Home' ELSE 'Facility' END AS unit_type,
		  CASE u.unit_type WHEN 1 THEN 'Private' WHEN 2 THEN 'Share' WHEN 3 THEN 'Public' ELSE '' END AS facility_type
		FROM rooms r
		JOIN units u ON u.unit_id = network(set_masklen(r.room_id, 80))
		JOIN sites s ON s.site_id = network(set_masklen(r.room_id, 64))
		WHERE r.room_id <<= $1::INET
		ORDER BY s.building, s.floor, u.unit_slot, r.room_slot`
	rows, err := r.db.QueryContext(ctx, q, branchID)
	if err != nil {
		return nil, fmt.Errorf("list rooms by branch: %w", err)
	}
	defer rows.Close()
	out := []*RoomWithAvailability{}
	for rows.Next() {
		var rwa RoomWithAvailability
		var floorInt int
		if err := rows.Scan(&rwa.RoomID, &rwa.TenantID, &rwa.UnitID, &rwa.UnitName,
			&rwa.BuildingName, &floorInt, &rwa.RoomName, &rwa.UnitType, &rwa.FacilityType); err != nil {
			return nil, err
		}
		rwa.Floor = formatFloor(floorInt)
		rwa.IsFull = false
		rwa.IsBound = false
		out = append(out, &rwa)
	}
	return out, rows.Err()
}

// =============================================================================
// Beds (v2 beds 表)
// =============================================================================

const bedsSelectCols = `
		host(b.bed_id) || '/96' AS bed_id,
		network(set_masklen(b.bed_id, 48))::text AS tenant_id,
		host(network(set_masklen(b.bed_id, 88))) || '/88' AS room_id,
		COALESCE(rm.room_name, '') AS room_name,
		b.bed_name
`

const bedsFromClause = `
	FROM beds b
	JOIN rooms rm ON rm.room_id = network(set_masklen(b.bed_id, 88))
`

func scanBed(rs rowScanner) (*domain.Bed, error) {
	var bd domain.Bed
	var roomName string
	if err := rs.Scan(&bd.BedID, &bd.TenantID, &bd.RoomID, &roomName, &bd.BedName); err != nil {
		return nil, fmt.Errorf("scan bed: %w", err)
	}
	bd.RoomName = sql.NullString{String: roomName, Valid: roomName != ""}
	bd.MattressMaterial = sql.NullString{} // v2 无该列
	bd.MattressThickness = sql.NullString{}
	return &bd, nil
}

func (r *PostgresUnitsRepository) ListBeds(ctx context.Context, tenantID, roomID, search string) ([]*domain.Bed, error) {
	if roomID == "" || !looksLikeINETPrefix(roomID) {
		return nil, nil
	}
	where := []string{"b.bed_id <<= $1::INET"}
	args := []any{roomID}
	argIdx := 2
	if search != "" {
		where = append(where, fmt.Sprintf("LOWER(b.bed_name) LIKE $%d", argIdx))
		args = append(args, "%"+strings.ToLower(search)+"%")
		argIdx++
	}
	q := `SELECT ` + bedsSelectCols + bedsFromClause + ` WHERE ` + strings.Join(where, " AND ") + ` ORDER BY b.bed_slot`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list beds: %w", err)
	}
	defer rows.Close()
	out := []*domain.Bed{}
	for rows.Next() {
		bd, err := scanBed(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, bd)
	}
	return out, rows.Err()
}

// ListAvailableBeds v2 简化：residents 关系未 v2 化；返回 ListBeds 全部。
func (r *PostgresUnitsRepository) ListAvailableBeds(ctx context.Context, tenantID, roomID, search, residentID string) ([]*domain.Bed, error) {
	return r.ListBeds(ctx, tenantID, roomID, search)
}

// ListBedsWithResident v2 简化：ResidentID 总返回 nil（未占用）。
func (r *PostgresUnitsRepository) ListBedsWithResident(ctx context.Context, tenantID, roomID, search string) ([]*BedWithResident, error) {
	beds, err := r.ListBeds(ctx, tenantID, roomID, search)
	if err != nil {
		return nil, err
	}
	out := make([]*BedWithResident, 0, len(beds))
	for _, bd := range beds {
		out = append(out, &BedWithResident{Bed: bd, ResidentID: nil})
	}
	return out, nil
}

func (r *PostgresUnitsRepository) GetBed(ctx context.Context, tenantID, bedID string) (*domain.Bed, error) {
	if bedID == "" || !looksLikeINETPrefix(bedID) {
		return nil, sql.ErrNoRows
	}
	q := `SELECT ` + bedsSelectCols + bedsFromClause + ` WHERE b.bed_id = $1::INET`
	rows, err := r.db.QueryContext(ctx, q, bedID)
	if err != nil {
		return nil, fmt.Errorf("get bed: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return scanBed(rows)
}

func (r *PostgresUnitsRepository) CreateBed(ctx context.Context, tenantID, roomID string, bed *domain.Bed) (string, error) {
	if roomID == "" || !looksLikeINETPrefix(roomID) {
		return "", fmt.Errorf("invalid room_id")
	}
	if bed == nil || bed.BedName == "" {
		return "", fmt.Errorf("bed_name required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "owl_v2.beds.alloc:"+roomID); err != nil {
		return "", fmt.Errorf("lock: %w", err)
	}
	// bed_slot 从 1 起（slot 0 = unbound 哨兵；0xFF wildcard 保留）
	var nextSlot int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(bed_slot), 0) + 1 FROM beds WHERE bed_id <<= $1::INET`, roomID).Scan(&nextSlot); err != nil {
		return "", fmt.Errorf("compute bed_slot: %w", err)
	}
	if nextSlot < 1 || nextSlot > 254 {
		return "", fmt.Errorf("bed_slot exhausted: %d (valid 1..254)", nextSlot)
	}
	bedPrefix, err := deriveBedCIDR(roomID, byte(nextSlot))
	if err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO beds (bed_id, bed_slot, bed_name)
		VALUES ($1::INET, $2, $3)`, bedPrefix, nextSlot, bed.BedName); err != nil {
		return "", fmt.Errorf("insert bed: %w", err)
	}
	return bedPrefix, tx.Commit()
}

func (r *PostgresUnitsRepository) UpdateBed(ctx context.Context, tenantID, bedID string, bed *domain.Bed) error {
	if bedID == "" || !looksLikeINETPrefix(bedID) {
		return fmt.Errorf("invalid bed_id")
	}
	if bed == nil || bed.BedName == "" {
		return fmt.Errorf("bed_name required")
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE beds SET bed_name = $2, updated_at = NOW() WHERE bed_id = $1::INET`, bedID, bed.BedName)
	if err != nil {
		return fmt.Errorf("update bed: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("bed not found: %s", bedID)
	}
	return nil
}

func (r *PostgresUnitsRepository) DeleteBed(ctx context.Context, tenantID, bedID string) error {
	if bedID == "" || !looksLikeINETPrefix(bedID) {
		return fmt.Errorf("invalid bed_id")
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM beds WHERE bed_id = $1::INET`, bedID)
	if err != nil {
		return fmt.Errorf("delete bed: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("bed not found: %s", bedID)
	}
	return nil
}

func (r *PostgresUnitsRepository) ListBedsByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) ([]*domain.Bed, error) {
	if len(roomIDs) == 0 {
		return nil, nil
	}
	args := make([]any, 0, len(roomIDs))
	conds := make([]string, 0, len(roomIDs))
	argIdx := 1
	for _, rid := range roomIDs {
		if !looksLikeINETPrefix(rid) {
			continue
		}
		conds = append(conds, fmt.Sprintf("b.bed_id <<= $%d::INET", argIdx))
		args = append(args, rid)
		argIdx++
	}
	if len(conds) == 0 {
		return nil, nil
	}
	q := `SELECT ` + bedsSelectCols + bedsFromClause + ` WHERE (` + strings.Join(conds, " OR ") + `) ORDER BY b.bed_id, b.bed_slot`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list beds by room ids: %w", err)
	}
	defer rows.Close()
	out := []*domain.Bed{}
	for rows.Next() {
		bd, err := scanBed(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, bd)
	}
	return out, rows.Err()
}
