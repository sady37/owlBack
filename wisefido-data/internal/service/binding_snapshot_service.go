package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// TakeBindingSnapshot 采集受影响 unit 下所有关联表的完整行，写入 binding_snapshots。
//   - triggerType: "device_binding" | "resident_move" | "caregiver_change"
//   - triggerEntityID: 触发变更的 device_id / resident_id
//   - triggerSummary: 人可读描述，如 "Radar_D523: Room101→Room201"
//   - unitIDs: 受影响的 unit_id 列表（新旧都传）
//   - changedBy: 操作人 user_id（可空）
func TakeBindingSnapshot(
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
	tenantID string,
	triggerType string,
	triggerEntityID string,
	triggerSummary string,
	unitIDs []string,
	changedBy string,
) error {
	if db == nil || tenantID == "" || len(unitIDs) == 0 {
		return nil
	}

	snapshot, err := collectSnapshot(ctx, db, tenantID, unitIDs)
	if err != nil {
		if logger != nil {
			logger.Warn("collectSnapshot failed", zap.String("tenant_id", tenantID), zap.Strings("unit_ids", unitIDs), zap.Error(err))
		}
		return err
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	var changedByPtr *string
	if changedBy != "" {
		changedByPtr = &changedBy
	}
	var triggerEntityPtr *string
	if triggerEntityID != "" {
		triggerEntityPtr = &triggerEntityID
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO binding_snapshots (tenant_id, trigger_type, trigger_entity_id, trigger_summary, affected_unit_ids, snapshot, changed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, tenantID, triggerType, triggerEntityPtr, triggerSummary, pq.Array(unitIDs), snapshotJSON, changedByPtr)
	if err != nil {
		return fmt.Errorf("insert binding_snapshot: %w", err)
	}

	if logger != nil {
		logger.Info("binding snapshot saved",
			zap.String("tenant_id", tenantID),
			zap.String("trigger_type", triggerType),
			zap.String("trigger_summary", triggerSummary),
			zap.Strings("unit_ids", unitIDs),
		)
	}
	return nil
}

// collectSnapshot 收集 unitIDs 下所有关联表的完整行
func collectSnapshot(ctx context.Context, db *sql.DB, tenantID string, unitIDs []string) (map[string]interface{}, error) {
	snap := map[string]interface{}{}

	// units → branch_ids
	units, branchIDs, err := queryRows(ctx, db,
		`SELECT row_to_json(u.*) FROM units u WHERE u.tenant_id = $1 AND u.unit_id = ANY($2)`,
		tenantID, pq.Array(unitIDs))
	if err != nil {
		return nil, fmt.Errorf("units: %w", err)
	}
	snap["units"] = units

	// branches
	if len(branchIDs) > 0 {
		branches, _, err := queryRows(ctx, db,
			`SELECT row_to_json(b.*) FROM branches b WHERE b.tenant_id = $1 AND b.branch_id = ANY($2)`,
			tenantID, pq.Array(branchIDs))
		if err != nil {
			return nil, fmt.Errorf("branches: %w", err)
		}
		snap["branches"] = branches
	}

	// rooms（layout_config 通过 config_versions + room_id + 时间点查）
	rooms, _, err := queryRows(ctx, db,
		`SELECT row_to_json(r.*) FROM rooms r WHERE r.tenant_id = $1 AND r.unit_id = ANY($2)`,
		tenantID, pq.Array(unitIDs))
	if err != nil {
		return nil, fmt.Errorf("rooms: %w", err)
	}
	snap["rooms"] = rooms

	// beds（通过 room）
	beds, _, err := queryRows(ctx, db,
		`SELECT row_to_json(bd.*) FROM beds bd
		 JOIN rooms r ON bd.room_id = r.room_id
		 WHERE bd.tenant_id = $1 AND r.unit_id = ANY($2)`,
		tenantID, pq.Array(unitIDs))
	if err != nil {
		return nil, fmt.Errorf("beds: %w", err)
	}
	snap["beds"] = beds

	// devices（bound_room 或 bound_bed 在这些 unit 下的）
	devices, _, err := queryRows(ctx, db,
		`SELECT row_to_json(d.*) FROM devices d
		 WHERE d.tenant_id = $1
		   AND (d.bound_room_id IN (SELECT room_id FROM rooms WHERE tenant_id = $1 AND unit_id = ANY($2))
		     OR d.bound_bed_id IN (SELECT bed_id FROM beds WHERE tenant_id = $1 AND room_id IN
		         (SELECT room_id FROM rooms WHERE tenant_id = $1 AND unit_id = ANY($2))))`,
		tenantID, pq.Array(unitIDs))
	if err != nil {
		return nil, fmt.Errorf("devices: %w", err)
	}
	snap["devices"] = devices

	// alarm_device / layout_config 不冗余存储，通过 device_id/room_id + snapshot_at 去 config_versions 查

	// residents（unit_id 在这些 unit 下的）
	residents, _, err := queryRows(ctx, db,
		`SELECT row_to_json(res.*) FROM residents res WHERE res.tenant_id = $1 AND res.unit_id = ANY($2)`,
		tenantID, pq.Array(unitIDs))
	if err != nil {
		return nil, fmt.Errorf("residents: %w", err)
	}
	snap["residents"] = residents

	// resident_caregivers
	resCaregivers, _, err := queryRows(ctx, db,
		`SELECT row_to_json(rc.*) FROM resident_caregivers rc
		 WHERE rc.tenant_id = $1
		   AND rc.resident_id IN (SELECT resident_id FROM residents WHERE tenant_id = $1 AND unit_id = ANY($2))`,
		tenantID, pq.Array(unitIDs))
	if err != nil {
		return nil, fmt.Errorf("resident_caregivers: %w", err)
	}
	snap["resident_caregivers"] = resCaregivers

	// users（护理人员：从 resident_caregivers.user_list 提取的 user_id）
	caregiverUsers, _, err := queryRows(ctx, db,
		`SELECT row_to_json(u.*) FROM users u
		 WHERE u.tenant_id = $1
		   AND u.user_id::text IN (
		     SELECT jsonb_array_elements_text(rc.user_list)
		     FROM resident_caregivers rc
		     WHERE rc.tenant_id = $1
		       AND rc.resident_id IN (SELECT resident_id FROM residents WHERE tenant_id = $1 AND unit_id = ANY($2))
		       AND jsonb_typeof(rc.user_list) = 'array'
		   )`,
		tenantID, pq.Array(unitIDs))
	if err != nil {
		// user_list 可能为空或格式不符，不阻塞
		caregiverUsers = []json.RawMessage{}
	}
	snap["users"] = caregiverUsers

	return snap, nil
}

// queryRows 执行 row_to_json 查询，返回 JSON 行数组 + 提取 branch_id 列表（用于级联查 branches）
func queryRows(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]json.RawMessage, []string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var result []json.RawMessage
	var branchIDs []string
	seen := map[string]bool{}

	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return nil, nil, err
		}
		result = append(result, raw)

		// 提取 branch_id（用于 units → branches 级联）
		var obj map[string]interface{}
		if json.Unmarshal(raw, &obj) == nil {
			if bid, ok := obj["branch_id"].(string); ok && bid != "" && !seen[bid] {
				branchIDs = append(branchIDs, bid)
				seen[bid] = true
			}
		}
	}
	if result == nil {
		result = []json.RawMessage{}
	}
	return result, branchIDs, rows.Err()
}

// snapshotLocationLabel 返回 "UnitName/RoomName/BedName" 或 "(unbound)"
func snapshotLocationLabel(ctx context.Context, db *sql.DB, tenantID, roomID, bedID string) string {
	if roomID == "" && bedID == "" {
		return "(unbound)"
	}
	var unitName, roomName, bedName string
	if roomID != "" {
		_ = db.QueryRowContext(ctx,
			`SELECT COALESCE(u.unit_name,''), COALESCE(r.room_name,'')
			 FROM rooms r LEFT JOIN units u ON r.unit_id = u.unit_id
			 WHERE r.room_id = $1 AND r.tenant_id = $2`, roomID, tenantID,
		).Scan(&unitName, &roomName)
	}
	if bedID != "" {
		_ = db.QueryRowContext(ctx,
			`SELECT COALESCE(b.bed_name,'') FROM beds b WHERE b.bed_id = $1 AND b.tenant_id = $2`,
			bedID, tenantID,
		).Scan(&bedName)
	}
	label := unitName
	if roomName != "" {
		label += "/" + roomName
	}
	if bedName != "" {
		label += "/" + bedName
	}
	if label == "" {
		return "(unbound)"
	}
	return label
}

// snapshotUnitLabel 返回 "UnitName" 或 "(none)"
func snapshotUnitLabel(ctx context.Context, db *sql.DB, tenantID, unitID string) string {
	if unitID == "" {
		return "(none)"
	}
	var name string
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(unit_name,'') FROM units WHERE unit_id = $1 AND tenant_id = $2`,
		unitID, tenantID,
	).Scan(&name); err != nil || name == "" {
		return unitID
	}
	return name
}
