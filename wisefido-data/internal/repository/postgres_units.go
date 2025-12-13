package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type PostgresUnitsRepo struct {
	db *sql.DB
}

func NewPostgresUnitsRepo(db *sql.DB) *PostgresUnitsRepo {
	return &PostgresUnitsRepo{db: db}
}

// ListBuildings: owlFront 需要 buildings 列表，但 owlRD 暂无 buildings 表
// 这里用 units 表做“虚拟 buildings”：按 (branch_tag, building) 分组，floors 用该组出现的最大楼层号（解析 '1F' -> 1）兜底
func (r *PostgresUnitsRepo) ListBuildings(ctx context.Context, tenantID string, branchTag string) ([]map[string]any, error) {
	if tenantID == "" {
		return []map[string]any{}, nil
	}
	where := "tenant_id = $1"
	args := []any{tenantID}
	if branchTag != "" {
		where += " AND branch_tag = $2"
		args = append(args, branchTag)
	}
	q := `
		SELECT
			COALESCE(branch_tag,'-') as branch_tag,
			COALESCE(building,'-') as building,
			MAX((NULLIF(REGEXP_REPLACE(floor, '[^0-9]', '', 'g'), '')::int)) as max_floor
		FROM units
		WHERE ` + where + `
		GROUP BY COALESCE(branch_tag,'-'), COALESCE(building,'-')
		ORDER BY COALESCE(branch_tag,'-'), COALESCE(building,'-')
	`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var tag string
		var b string
		var maxFloor sql.NullInt64
		if err := rows.Scan(&tag, &b, &maxFloor); err != nil {
			return nil, err
		}
		floors := 1
		if maxFloor.Valid && maxFloor.Int64 > 0 {
			floors = int(maxFloor.Int64)
		}
		// building_id：前端只是用作 key，这里用可读的稳定值
		buildingID := fmt.Sprintf("%s-%s", tag, b)
		out = append(out, map[string]any{
			"building_id":   buildingID,
			"building_name": b,
			"floors":        floors,
			"tenant_id":     tenantID,
			"branch_tag":    tag,
		})
	}
	return out, rows.Err()
}

func (r *PostgresUnitsRepo) ListUnits(ctx context.Context, tenantID string, filters map[string]string, page, size int) ([]Unit, int, error) {
	if tenantID == "" {
		return []Unit{}, 0, nil
	}

	where := []string{"u.tenant_id = $1"}
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

	addEq("u.branch_tag", filters["branch_tag"])
	addEq("u.building", filters["building"])
	addEq("u.floor", filters["floor"])
	addEq("u.area_tag", filters["area_tag"])
	addEq("u.unit_number", filters["unit_number"])
	addEq("u.unit_name", filters["unit_name"])
	addEq("u.unit_type", filters["unit_type"])

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
			u.timezone
		FROM units u
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY u.unit_name
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Unit, 0)
	for rows.Next() {
		var u Unit
		var branchTag sql.NullString
		var building sql.NullString
		var floor sql.NullString
		if err := rows.Scan(
			&u.UnitID,
			&u.TenantID,
			&branchTag,
			&u.UnitName,
			&building,
			&floor,
			&u.AreaTag,
			&u.UnitNumber,
			&u.LayoutConfig,
			&u.UnitType,
			&u.IsPublicSpace,
			&u.IsMultiPersonRoom,
			&u.Timezone,
		); err != nil {
			return nil, 0, err
		}
		if branchTag.Valid {
			u.BranchTag = branchTag.String
		}
		if building.Valid {
			u.Building = building.String
		}
		if floor.Valid {
			u.Floor = floor.String
		}
		items = append(items, u)
	}
	return items, total, rows.Err()
}

func (r *PostgresUnitsRepo) GetUnit(ctx context.Context, tenantID, unitID string) (*Unit, error) {
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
			u.timezone
		FROM units u
		WHERE u.tenant_id = $1 AND u.unit_id = $2
	`
	var u Unit
	var branchTag sql.NullString
	var building sql.NullString
	var floor sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, unitID).Scan(
		&u.UnitID,
		&u.TenantID,
		&branchTag,
		&u.UnitName,
		&building,
		&floor,
		&u.AreaTag,
		&u.UnitNumber,
		&u.LayoutConfig,
		&u.UnitType,
		&u.IsPublicSpace,
		&u.IsMultiPersonRoom,
		&u.Timezone,
	)
	if err != nil {
		return nil, err
	}
	if branchTag.Valid {
		u.BranchTag = branchTag.String
	}
	if building.Valid {
		u.Building = building.String
	}
	if floor.Valid {
		u.Floor = floor.String
	}
	return &u, nil
}

func (r *PostgresUnitsRepo) CreateUnit(ctx context.Context, tenantID string, payload map[string]any) (*Unit, error) {
	// 最小实现：插入 units；缺失字段用默认值
	q := `
		INSERT INTO units (tenant_id, branch_tag, unit_name, building, floor, area_tag, unit_number, layout_config, unit_type, is_public_space, is_multi_person_room, timezone)
		VALUES ($1, $2, $3, COALESCE($4,'-'), COALESCE($5,'1F'), $6, $7, $8::jsonb, $9, COALESCE($10,false), COALESCE($11,false), $12)
		RETURNING unit_id::text
	`
	branchTag, _ := payload["branch_tag"].(string)
	unitName, _ := payload["unit_name"].(string)
	building, _ := payload["building"].(string)
	floor, _ := payload["floor"].(string)
	areaTag, _ := payload["area_tag"].(string)
	unitNumber, _ := payload["unit_number"].(string)
	layoutJSON, _ := payload["layout_config"].(string) // allow json string
	unitType, _ := payload["unit_type"].(string)
	isPublic, _ := payload["is_public_space"].(bool)
	isMulti, _ := payload["is_multi_person_room"].(bool)
	timezone, _ := payload["timezone"].(string)

	var area sql.NullString
	if areaTag != "" {
		area = sql.NullString{String: areaTag, Valid: true}
	}
	var layout sql.NullString
	if layoutJSON != "" {
		layout = sql.NullString{String: layoutJSON, Valid: true}
	}

	var unitID string
	if err := r.db.QueryRowContext(ctx, q, tenantID, branchTag, unitName, building, floor, area, unitNumber, nullStringToAny(layout), unitType, isPublic, isMulti, timezone).Scan(&unitID); err != nil {
		return nil, err
	}
	return r.GetUnit(ctx, tenantID, unitID)
}

func (r *PostgresUnitsRepo) UpdateUnit(ctx context.Context, tenantID, unitID string, payload map[string]any) (*Unit, error) {
	// 简化：只更新允许字段（不做动态 SQL 的全覆盖）
	set := []string{}
	args := []any{tenantID, unitID}
	argN := 3
	add := func(col string, v any) {
		set = append(set, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, v)
		argN++
	}
	if v, ok := payload["branch_tag"]; ok {
		add("branch_tag", v)
	}
	if v, ok := payload["unit_name"]; ok {
		add("unit_name", v)
	}
	if v, ok := payload["building"]; ok {
		add("building", v)
	}
	if v, ok := payload["floor"]; ok {
		add("floor", v)
	}
	if v, ok := payload["area_tag"]; ok {
		add("area_tag", v)
	}
	if v, ok := payload["unit_number"]; ok {
		add("unit_number", v)
	}
	if v, ok := payload["layout_config"]; ok {
		// expect json string
		add("layout_config", fmt.Sprintf("%v", v))
		set[len(set)-1] = fmt.Sprintf("layout_config = $%d::jsonb", argN-1)
	}
	if v, ok := payload["is_public_space"]; ok {
		add("is_public_space", v)
	}
	if v, ok := payload["is_multi_person_room"]; ok {
		add("is_multi_person_room", v)
	}
	if v, ok := payload["timezone"]; ok {
		add("timezone", v)
	}

	if len(set) == 0 {
		return r.GetUnit(ctx, tenantID, unitID)
	}
	q := fmt.Sprintf("UPDATE units SET %s WHERE tenant_id = $1 AND unit_id = $2", strings.Join(set, ", "))
	if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	return r.GetUnit(ctx, tenantID, unitID)
}

func (r *PostgresUnitsRepo) DeleteUnit(ctx context.Context, tenantID, unitID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM units WHERE tenant_id = $1 AND unit_id = $2", tenantID, unitID)
	return err
}

func (r *PostgresUnitsRepo) ListRoomsWithBeds(ctx context.Context, unitID string) ([]map[string]any, error) {
	// unit_id -> tenant_id + unit_name
	var tenantID string
	var unitName string
	if err := r.db.QueryRowContext(ctx, "SELECT tenant_id::text, unit_name FROM units WHERE unit_id = $1", unitID).Scan(&tenantID, &unitName); err != nil {
		return nil, err
	}

	qRooms := `
		SELECT 
			r.room_id::text,
			r.tenant_id::text,
			r.unit_id::text,
			r.room_name,
			(r.room_name = u.unit_name) as is_default,
			CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r
		JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE r.tenant_id = $1 AND r.unit_id = $2
		ORDER BY r.room_name
	`
	rows, err := r.db.QueryContext(ctx, qRooms, tenantID, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type roomRow struct {
		room Room
	}
	rooms := make([]Room, 0)
	roomIDs := make([]string, 0)
	for rows.Next() {
		var rr Room
		var layout sql.NullString
		var tenant sql.NullString
		if err := rows.Scan(&rr.RoomID, &tenant, &rr.UnitID, &rr.RoomName, &rr.IsDefault, &layout); err != nil {
			return nil, err
		}
		rr.TenantID = tenant
		rr.LayoutConfig = layout
		rooms = append(rooms, rr)
		roomIDs = append(roomIDs, rr.RoomID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// beds
	bedsByRoom := map[string][]map[string]any{}
	if len(roomIDs) > 0 {
		// build IN clause
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
				b.mattress_thickness,
				b.bound_device_count
			FROM beds b
			WHERE b.tenant_id = $1 AND b.room_id IN (` + strings.Join(in, ",") + `)
			ORDER BY b.bed_name
		`
		brows, err := r.db.QueryContext(ctx, qBeds, args...)
		if err != nil {
			return nil, err
		}
		defer brows.Close()
		for brows.Next() {
			var b Bed
			var tenant sql.NullString
			if err := brows.Scan(
				&b.BedID,
				&tenant,
				&b.RoomID,
				&b.BedName,
				&b.BedType,
				&b.MattressMaterial,
				&b.MattressThickness,
				&b.BoundDeviceCount,
			); err != nil {
				return nil, err
			}
			b.TenantID = tenant
			bedsByRoom[b.RoomID] = append(bedsByRoom[b.RoomID], b.ToJSON())
		}
		if err := brows.Err(); err != nil {
			return nil, err
		}
	}

	out := make([]map[string]any, 0, len(rooms))
	for _, rr := range rooms {
		m := rr.ToJSON()
		m["beds"] = bedsByRoom[rr.RoomID]
		if m["beds"] == nil {
			m["beds"] = []any{}
		}
		out = append(out, m)
	}
	_ = unitName // used for is_default already
	return out, nil
}

func (r *PostgresUnitsRepo) CreateRoom(ctx context.Context, unitID string, payload map[string]any) (*Room, error) {
	// infer tenant from unit
	var tenantID string
	if err := r.db.QueryRowContext(ctx, "SELECT tenant_id::text FROM units WHERE unit_id = $1", unitID).Scan(&tenantID); err != nil {
		return nil, err
	}
	roomName, _ := payload["room_name"].(string)
	layout, _ := payload["layout_config"].(string)
	var roomID string
	q := `
		INSERT INTO rooms (tenant_id, unit_id, room_name, layout_config)
		VALUES ($1, $2, $3, $4::jsonb)
		RETURNING room_id::text
	`
	if _, ok := payload["layout_config"]; !ok || layout == "" {
		q = `
			INSERT INTO rooms (tenant_id, unit_id, room_name)
			VALUES ($1, $2, $3)
			RETURNING room_id::text
		`
		if err := r.db.QueryRowContext(ctx, q, tenantID, unitID, roomName).Scan(&roomID); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.QueryRowContext(ctx, q, tenantID, unitID, roomName, layout).Scan(&roomID); err != nil {
			return nil, err
		}
	}

	// get row
	qGet := `
		SELECT r.room_id::text, r.tenant_id::text, r.unit_id::text, r.room_name, (r.room_name = u.unit_name) as is_default,
		       CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r JOIN units u ON r.unit_id=u.unit_id AND r.tenant_id=u.tenant_id
		WHERE r.room_id = $1
	`
	var rr Room
	var tenant sql.NullString
	if err := r.db.QueryRowContext(ctx, qGet, roomID).Scan(&rr.RoomID, &tenant, &rr.UnitID, &rr.RoomName, &rr.IsDefault, &rr.LayoutConfig); err != nil {
		return nil, err
	}
	rr.TenantID = tenant
	return &rr, nil
}

func (r *PostgresUnitsRepo) UpdateRoom(ctx context.Context, roomID string, payload map[string]any) (*Room, error) {
	set := []string{}
	args := []any{roomID}
	argN := 2
	if v, ok := payload["room_name"]; ok {
		set = append(set, fmt.Sprintf("room_name = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v, ok := payload["layout_config"]; ok {
		set = append(set, fmt.Sprintf("layout_config = $%d::jsonb", argN))
		args = append(args, fmt.Sprintf("%v", v))
		argN++
	}
	if len(set) > 0 {
		q := "UPDATE rooms SET " + strings.Join(set, ", ") + " WHERE room_id = $1"
		if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
			return nil, err
		}
	}
	qGet := `
		SELECT r.room_id::text, r.tenant_id::text, r.unit_id::text, r.room_name, (r.room_name = u.unit_name) as is_default,
		       CASE WHEN r.layout_config IS NULL THEN NULL ELSE r.layout_config::text END as layout_config
		FROM rooms r JOIN units u ON r.unit_id=u.unit_id AND r.tenant_id=u.tenant_id
		WHERE r.room_id = $1
	`
	var rr Room
	var tenant sql.NullString
	if err := r.db.QueryRowContext(ctx, qGet, roomID).Scan(&rr.RoomID, &tenant, &rr.UnitID, &rr.RoomName, &rr.IsDefault, &rr.LayoutConfig); err != nil {
		return nil, err
	}
	rr.TenantID = tenant
	return &rr, nil
}

func (r *PostgresUnitsRepo) DeleteRoom(ctx context.Context, roomID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM rooms WHERE room_id = $1", roomID)
	return err
}

func (r *PostgresUnitsRepo) ListBeds(ctx context.Context, roomID string) ([]Bed, error) {
	q := `
		SELECT bed_id::text, tenant_id::text, room_id::text, bed_name, bed_type, mattress_material, mattress_thickness, bound_device_count
		FROM beds WHERE room_id = $1 ORDER BY bed_name
	`
	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Bed{}
	for rows.Next() {
		var b Bed
		var tenant sql.NullString
		if err := rows.Scan(&b.BedID, &tenant, &b.RoomID, &b.BedName, &b.BedType, &b.MattressMaterial, &b.MattressThickness, &b.BoundDeviceCount); err != nil {
			return nil, err
		}
		b.TenantID = tenant
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *PostgresUnitsRepo) CreateBed(ctx context.Context, roomID string, payload map[string]any) (*Bed, error) {
	bedName, _ := payload["bed_name"].(string)
	material, _ := payload["mattress_material"].(string)
	thickness, _ := payload["mattress_thickness"].(string)

	// 默认 bed_type = NonActiveBed
	var bedID string
	q := `
		INSERT INTO beds (tenant_id, room_id, bed_name, bed_type, mattress_material, mattress_thickness)
		SELECT tenant_id, $1, $2, 'NonActiveBed', NULLIF($3,''), NULLIF($4,'')
		FROM rooms WHERE room_id = $1
		RETURNING bed_id::text
	`
	if err := r.db.QueryRowContext(ctx, q, roomID, bedName, material, thickness).Scan(&bedID); err != nil {
		return nil, err
	}
	// get
	qGet := `
		SELECT bed_id::text, tenant_id::text, room_id::text, bed_name, bed_type, mattress_material, mattress_thickness, bound_device_count
		FROM beds WHERE bed_id = $1
	`
	var b Bed
	var tenant sql.NullString
	if err := r.db.QueryRowContext(ctx, qGet, bedID).Scan(&b.BedID, &tenant, &b.RoomID, &b.BedName, &b.BedType, &b.MattressMaterial, &b.MattressThickness, &b.BoundDeviceCount); err != nil {
		return nil, err
	}
	b.TenantID = tenant
	return &b, nil
}

func (r *PostgresUnitsRepo) UpdateBed(ctx context.Context, bedID string, payload map[string]any) (*Bed, error) {
	set := []string{}
	args := []any{bedID}
	argN := 2
	if v, ok := payload["bed_name"]; ok {
		set = append(set, fmt.Sprintf("bed_name = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v, ok := payload["mattress_material"]; ok {
		set = append(set, fmt.Sprintf("mattress_material = NULLIF($%d,'')", argN))
		args = append(args, fmt.Sprintf("%v", v))
		argN++
	}
	if v, ok := payload["mattress_thickness"]; ok {
		set = append(set, fmt.Sprintf("mattress_thickness = NULLIF($%d,'')", argN))
		args = append(args, fmt.Sprintf("%v", v))
		argN++
	}
	// resident_id 不在 beds 表中（前端字段为可选），这里忽略
	if len(set) > 0 {
		q := "UPDATE beds SET " + strings.Join(set, ", ") + " WHERE bed_id = $1"
		if _, err := r.db.ExecContext(ctx, q, args...); err != nil {
			return nil, err
		}
	}
	qGet := `
		SELECT bed_id::text, tenant_id::text, room_id::text, bed_name, bed_type, mattress_material, mattress_thickness, bound_device_count
		FROM beds WHERE bed_id = $1
	`
	var b Bed
	var tenant sql.NullString
	if err := r.db.QueryRowContext(ctx, qGet, bedID).Scan(&b.BedID, &tenant, &b.RoomID, &b.BedName, &b.BedType, &b.MattressMaterial, &b.MattressThickness, &b.BoundDeviceCount); err != nil {
		return nil, err
	}
	b.TenantID = tenant
	return &b, nil
}

func (r *PostgresUnitsRepo) DeleteBed(ctx context.Context, bedID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM beds WHERE bed_id = $1", bedID)
	return err
}

func nullStringToAny(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}
