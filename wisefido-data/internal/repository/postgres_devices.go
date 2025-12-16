package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

type PostgresDevicesRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresDevicesRepo(db *sql.DB) *PostgresDevicesRepo {
	return &PostgresDevicesRepo{db: db}
}

// SetLogger sets the logger for this repository (optional, for logging device connection events)
func (r *PostgresDevicesRepo) SetLogger(logger *zap.Logger) {
	r.logger = logger
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

// GetOrCreateDeviceFromStore attempts to get device from devices table, and if not found,
// checks device_store table. If device_store exists and is allocated, creates devices record.
// This is used for automatic device creation on first MQTT connection.
// Returns the device and error. Errors are logged for security auditing.
func (r *PostgresDevicesRepo) GetOrCreateDeviceFromStore(ctx context.Context, identifier string, mqttTopic string) (*Device, error) {
	// Helper function to log messages (if logger is available)
	logInfo := func(msg string, fields ...zap.Field) {
		if r.logger != nil {
			r.logger.Info(msg, fields...)
		}
	}
	logWarn := func(msg string, fields ...zap.Field) {
		if r.logger != nil {
			r.logger.Warn(msg, fields...)
		}
	}

	// 1. First, try to get device from devices table by serial_number or uid
	deviceQuery := `
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
			NULL as unit_id,
			CASE WHEN d.bound_room_id IS NULL THEN NULL ELSE d.bound_room_id::text END as bound_room_id,
			CASE WHEN d.bound_bed_id  IS NULL THEN NULL ELSE d.bound_bed_id::text  END as bound_bed_id,
			CASE WHEN d.metadata IS NULL THEN NULL ELSE d.metadata::text END as metadata
		FROM devices d
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE (d.serial_number = $1 OR d.uid = $1)
		LIMIT 1
	`

	var d Device
	err := r.db.QueryRowContext(ctx, deviceQuery, identifier).Scan(
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
	)

	if err == nil {
		// Device found in devices table, return it
		return &d, nil
	}

	if err != sql.ErrNoRows {
		// Unexpected database error
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	// 2. Device not found in devices table, check device_store table
	unallocatedTenantID := "00000000-0000-0000-0000-000000000000"
	deviceStoreQuery := `
		SELECT
			device_store_id::text,
			device_type,
			device_model,
			serial_number,
			uid,
			imei,
			comm_mode,
			mcu_model,
			firmware_version,
			tenant_id::text,
			allow_access
		FROM device_store
		WHERE (serial_number = $1 OR uid = $1)
		LIMIT 1
	`

	var dsDeviceStoreID, dsDeviceType, dsTenantID string
	var dsDeviceModel, dsSerialNumber, dsUID, dsIMEI, dsCommMode, dsMCUModel, dsFirmwareVersion sql.NullString
	var dsAllowAccess bool

	err = r.db.QueryRowContext(ctx, deviceStoreQuery, identifier).Scan(
		&dsDeviceStoreID,
		&dsDeviceType,
		&dsDeviceModel,
		&dsSerialNumber,
		&dsUID,
		&dsIMEI,
		&dsCommMode,
		&dsMCUModel,
		&dsFirmwareVersion,
		&dsTenantID,
		&dsAllowAccess,
	)

	if err == sql.ErrNoRows {
		// Case 3: Device not registered in device_store
		logWarn("Unauthorized device connection attempt",
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("reason", "device_not_registered"),
			zap.String("action", "connection_rejected"),
			zap.String("security_level", "warning"),
		)
		return nil, fmt.Errorf("unauthorized device: not registered in device_store")
	}

	if err != nil {
		// Unexpected database error
		return nil, fmt.Errorf("failed to query device_store: %w", err)
	}

	// 3. Check if device is allocated to a tenant
	if dsTenantID == unallocatedTenantID {
		// Case 2: Device registered but not allocated
		serialNum := ""
		if dsSerialNumber.Valid {
			serialNum = dsSerialNumber.String
		}
		uid := ""
		if dsUID.Valid {
			uid = dsUID.String
		}
		logWarn("Device connection rejected: not allocated",
			zap.String("device_store_id", dsDeviceStoreID),
			zap.String("serial_number", serialNum),
			zap.String("uid", uid),
			zap.String("reason", "device_not_allocated"),
			zap.String("action", "connection_rejected"),
		)
		return nil, fmt.Errorf("device not allocated to tenant")
	}

	// Case 1: Device is registered and allocated, create devices record
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Generate device name from device_type and serial_number/uid
	deviceName := dsDeviceType
	if dsSerialNumber.Valid && dsSerialNumber.String != "" {
		deviceName = dsDeviceType + "-" + dsSerialNumber.String
	} else if dsUID.Valid && dsUID.String != "" {
		deviceName = dsDeviceType + "-" + dsUID.String
	}

	// Insert device record
	insertQuery := `
		INSERT INTO devices (
			tenant_id,
			device_store_id,
			device_name,
			serial_number,
			uid,
			status,
			business_access,
			monitoring_enabled
		) VALUES ($1, $2, $3, $4, $5, 'online', 'pending', FALSE)
		RETURNING device_id::text
	`

	var newDeviceID string
	err = tx.QueryRowContext(ctx, insertQuery,
		dsTenantID,
		dsDeviceStoreID,
		deviceName,
		dsSerialNumber,
		dsUID,
	).Scan(&newDeviceID)

	if err != nil {
		return nil, fmt.Errorf("failed to create device record: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log successful auto-creation
	serialNum := ""
	if dsSerialNumber.Valid {
		serialNum = dsSerialNumber.String
	}
	uid := ""
	if dsUID.Valid {
		uid = dsUID.String
	}
	logInfo("Device auto-created from device_store",
		zap.String("device_store_id", dsDeviceStoreID),
		zap.String("device_id", newDeviceID),
		zap.String("tenant_id", dsTenantID),
		zap.String("serial_number", serialNum),
		zap.String("uid", uid),
		zap.String("device_type", dsDeviceType),
		zap.String("source", "mqtt_first_connection"),
		zap.String("mqtt_topic", mqttTopic),
	)

	// Query and return the newly created device
	return r.GetDevice(ctx, dsTenantID, newDeviceID)
}


