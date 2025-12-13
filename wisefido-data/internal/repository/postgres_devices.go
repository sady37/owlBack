package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type PostgresDevicesRepo struct {
	db *sql.DB
}

func NewPostgresDevicesRepo(db *sql.DB) *PostgresDevicesRepo {
	return &PostgresDevicesRepo{db: db}
}

func (r *PostgresDevicesRepo) ListDevices(ctx context.Context, tenantID string, filters map[string]any) ([]Device, int, error) {
	if tenantID == "" {
		return []Device{}, 0, nil
	}

	where := []string{"d.tenant_id = $1", "d.status <> 'disabled'"}
	args := []any{tenantID}
	argN := 2

	// status IN (...)
	if v, ok := filters["status"]; ok {
		if arr, ok := v.([]string); ok && len(arr) > 0 {
			where = append(where, fmt.Sprintf("d.status = ANY($%d)", argN))
			args = append(args, pq.Array(arr))
			argN++
		}
	}
	if v, ok := filters["business_access"].(string); ok && v != "" {
		where = append(where, fmt.Sprintf("d.business_access = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v, ok := filters["device_type"].(string); ok && v != "" {
		where = append(where, fmt.Sprintf("ds.device_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v, ok := filters["search_type"].(string); ok && v != "" {
		kw, _ := filters["search_keyword"].(string)
		if kw != "" {
			col := "d.device_name"
			switch v {
			case "device_name":
				col = "d.device_name"
			case "serial_number":
				col = "d.serial_number"
			case "uid":
				col = "d.uid"
			}
			where = append(where, fmt.Sprintf("%s ILIKE $%d", col, argN))
			args = append(args, "%"+kw+"%")
			argN++
		}
	}

	queryCount := `
		SELECT COUNT(*)
		FROM devices d
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page, _ := filters["page"].(int)
	size, _ := filters["size"].(int)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	argsList := append(args, size, offset)
	limitPos := argN
	offsetPos := argN + 1

	q := `
		SELECT
			d.device_id::text,
			d.tenant_id::text,
			CASE WHEN d.device_store_id IS NULL THEN NULL ELSE d.device_store_id::text END as device_store_id,
			d.device_name,
			ds.device_model,
			ds.device_type,
			d.serial_number,
			d.uid,
			ds.imei,
			ds.comm_mode,
			ds.firmware_version,
			ds.mcu_model,
			d.status,
			d.business_access,
			d.monitoring_enabled,
			COALESCE(r1.unit_id::text, r2.unit_id::text) as unit_id,
			CASE WHEN d.bound_room_id IS NULL THEN NULL ELSE d.bound_room_id::text END as bound_room_id,
			CASE WHEN d.bound_bed_id  IS NULL THEN NULL ELSE d.bound_bed_id::text  END as bound_bed_id,
			CASE WHEN d.metadata IS NULL THEN NULL ELSE d.metadata::text END as metadata
		FROM devices d
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		LEFT JOIN rooms r1 ON d.bound_room_id = r1.room_id
		LEFT JOIN beds  b  ON d.bound_bed_id  = b.bed_id
		LEFT JOIN rooms r2 ON b.room_id = r2.room_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY d.device_name
		LIMIT $` + fmt.Sprintf("%d", limitPos) + ` OFFSET $` + fmt.Sprintf("%d", offsetPos)

	rows, err := r.db.QueryContext(ctx, q, argsList...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []Device{}
	for rows.Next() {
		var d Device
		if err := rows.Scan(
			&d.DeviceID,
			&d.TenantID,
			&d.DeviceStoreID,
			&d.DeviceName,
			&d.DeviceModel,
			&d.DeviceType,
			&d.SerialNumber,
			&d.UID,
			&d.IMEI,
			&d.CommMode,
			&d.FirmwareVersion,
			&d.MCUModel,
			&d.Status,
			&d.BusinessAccess,
			&d.MonitoringEnabled,
			&d.UnitID,
			&d.BoundRoomID,
			&d.BoundBedID,
			&d.Metadata,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (r *PostgresDevicesRepo) GetDevice(ctx context.Context, tenantID, deviceID string) (*Device, error) {
	q := `
		SELECT
			d.device_id::text,
			d.tenant_id::text,
			CASE WHEN d.device_store_id IS NULL THEN NULL ELSE d.device_store_id::text END as device_store_id,
			d.device_name,
			ds.device_model,
			ds.device_type,
			d.serial_number,
			d.uid,
			ds.imei,
			ds.comm_mode,
			ds.firmware_version,
			ds.mcu_model,
			d.status,
			d.business_access,
			d.monitoring_enabled,
			COALESCE(r1.unit_id::text, r2.unit_id::text) as unit_id,
			CASE WHEN d.bound_room_id IS NULL THEN NULL ELSE d.bound_room_id::text END as bound_room_id,
			CASE WHEN d.bound_bed_id  IS NULL THEN NULL ELSE d.bound_bed_id::text  END as bound_bed_id,
			CASE WHEN d.metadata IS NULL THEN NULL ELSE d.metadata::text END as metadata
		FROM devices d
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		LEFT JOIN rooms r1 ON d.bound_room_id = r1.room_id
		LEFT JOIN beds  b  ON d.bound_bed_id  = b.bed_id
		LEFT JOIN rooms r2 ON b.room_id = r2.room_id
		WHERE d.tenant_id = $1 AND d.device_id = $2
	`
	var d Device
	if err := r.db.QueryRowContext(ctx, q, tenantID, deviceID).Scan(
		&d.DeviceID,
		&d.TenantID,
		&d.DeviceStoreID,
		&d.DeviceName,
		&d.DeviceModel,
		&d.DeviceType,
		&d.SerialNumber,
		&d.UID,
		&d.IMEI,
		&d.CommMode,
		&d.FirmwareVersion,
		&d.MCUModel,
		&d.Status,
		&d.BusinessAccess,
		&d.MonitoringEnabled,
		&d.UnitID,
		&d.BoundRoomID,
		&d.BoundBedID,
		&d.Metadata,
	); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *PostgresDevicesRepo) UpdateDevice(ctx context.Context, tenantID, deviceID string, payload map[string]any) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	set := []string{}
	args := []any{tenantID, deviceID}
	argN := 3
	add := func(col string, v any) {
		set = append(set, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, v)
		argN++
	}
	if v, ok := payload["device_name"]; ok {
		add("device_name", v)
	}
	if v, ok := payload["business_access"]; ok {
		add("business_access", v)
	}
	if v, ok := payload["status"]; ok {
		add("status", v)
	}
	if v, ok := payload["monitoring_enabled"]; ok {
		add("monitoring_enabled", v)
	}
	if v, ok := payload["bound_room_id"]; ok {
		add("bound_room_id", v)
	}
	if v, ok := payload["bound_bed_id"]; ok {
		add("bound_bed_id", v)
	}

	if len(set) == 0 {
		return nil
	}
	q := "UPDATE devices SET " + strings.Join(set, ", ") + " WHERE tenant_id = $1 AND device_id = $2"
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresDevicesRepo) DisableDevice(ctx context.Context, tenantID, deviceID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE devices
		SET status='disabled', business_access='rejected', monitoring_enabled=FALSE
		WHERE tenant_id=$1 AND device_id=$2
	`, tenantID, deviceID)
	return err
}


