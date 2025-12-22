package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"wisefido-data/internal/domain"
)

// PostgresDevicesRepository 设备Repository实现（强类型）
// 遵循"bottom-up"设计原则，替代已删除的数据库触发器
type PostgresDevicesRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresDevicesRepository 创建设备Repository
func NewPostgresDevicesRepository(db *sql.DB) *PostgresDevicesRepository {
	return &PostgresDevicesRepository{db: db}
}

// SetLogger 设置日志记录器（可选，用于记录设备连接事件）
func (r *PostgresDevicesRepository) SetLogger(logger *zap.Logger) {
	r.logger = logger
}

// ListDevices 查询设备列表
// 功能：支持多种过滤条件和分页，自动过滤status='disabled'的设备
func (r *PostgresDevicesRepository) ListDevices(ctx context.Context, tenantID string, filters DeviceFilters, page, size int) ([]*domain.Device, int, error) {
	if tenantID == "" {
		return []*domain.Device{}, 0, nil
	}

	where := []string{"d.tenant_id = $1", "d.status <> 'disabled'"}
	args := []any{tenantID}
	argN := 2

	// status IN (...)
	if len(filters.Status) > 0 {
		where = append(where, fmt.Sprintf("d.status = ANY($%d)", argN))
		args = append(args, pq.Array(filters.Status))
		argN++
	}
	if filters.BusinessAccess != "" {
		where = append(where, fmt.Sprintf("d.business_access = $%d", argN))
		args = append(args, filters.BusinessAccess)
		argN++
	}
	if filters.DeviceType != "" {
		where = append(where, fmt.Sprintf("ds.device_type = $%d", argN))
		args = append(args, filters.DeviceType)
		argN++
	}
	if filters.SearchType != "" && filters.SearchKeyword != "" {
		col := "d.device_name"
		switch filters.SearchType {
		case "device_name":
			col = "d.device_name"
		case "serial_number":
			col = "d.serial_number"
		case "uid":
			col = "d.uid"
		}
		where = append(where, fmt.Sprintf("%s ILIKE $%d", col, argN))
		args = append(args, "%"+filters.SearchKeyword+"%")
		argN++
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
			d.device_store_id,
			d.device_name,
			d.serial_number,
			d.uid,
			d.bound_room_id,
			d.bound_bed_id,
			d.status,
			d.business_access,
			d.monitoring_enabled,
			d.metadata
		FROM devices d
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY d.device_name
		LIMIT $` + fmt.Sprintf("%d", limitPos) + ` OFFSET $` + fmt.Sprintf("%d", offsetPos)

	rows, err := r.db.QueryContext(ctx, q, argsList...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []*domain.Device{}
	for rows.Next() {
		var d domain.Device
		if err := rows.Scan(
			&d.DeviceID,
			&d.TenantID,
			&d.DeviceStoreID,
			&d.DeviceName,
			&d.SerialNumber,
			&d.UID,
			&d.BoundRoomID,
			&d.BoundBedID,
			&d.Status,
			&d.BusinessAccess,
			&d.MonitoringEnabled,
			&d.Metadata,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, &d)
	}
	return out, total, rows.Err()
}

// GetDevice 查询单个设备
func (r *PostgresDevicesRepository) GetDevice(ctx context.Context, tenantID, deviceID string) (*domain.Device, error) {
	q := `
		SELECT
			d.device_id::text,
			d.tenant_id::text,
			d.device_store_id,
			d.device_name,
			d.serial_number,
			d.uid,
			d.bound_room_id,
			d.bound_bed_id,
			d.status,
			d.business_access,
			d.monitoring_enabled,
			d.metadata
		FROM devices d
		WHERE d.tenant_id = $1 AND d.device_id = $2
	`
	var d domain.Device
	if err := r.db.QueryRowContext(ctx, q, tenantID, deviceID).Scan(
		&d.DeviceID,
		&d.TenantID,
		&d.DeviceStoreID,
		&d.DeviceName,
		&d.SerialNumber,
		&d.UID,
		&d.BoundRoomID,
		&d.BoundBedID,
		&d.Status,
		&d.BusinessAccess,
		&d.MonitoringEnabled,
		&d.Metadata,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: tenant_id=%s, device_id=%s", tenantID, deviceID)
		}
		return nil, err
	}
	return &d, nil
}

// CreateDevice 手动创建设备与位置的绑定关系（出库操作）
// 替代触发器：trigger_validate_device_bed_tenant, trigger_validate_device_store_tenant
// 功能：系统管理员从device_store出库，创建设备与位置的绑定关系
func (r *PostgresDevicesRepository) CreateDevice(ctx context.Context, tenantID string, device *domain.Device) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if device == nil {
		return "", fmt.Errorf("device is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// 1. 验证device_store_id
	if !device.DeviceStoreID.Valid || device.DeviceStoreID.String == "" {
		return "", fmt.Errorf("device_store_id is required")
	}
	deviceStoreID := device.DeviceStoreID.String

	// 2. 查询device_store，验证是否存在且已分配给该tenant
	var dsTenantID sql.NullString
	var dsSerialNumber, dsUID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT tenant_id, serial_number, uid
		FROM device_store
		WHERE device_store_id = $1
	`, deviceStoreID).Scan(&dsTenantID, &dsSerialNumber, &dsUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("device_store not found: device_store_id=%s", deviceStoreID)
		}
		return "", fmt.Errorf("failed to query device_store: %w", err)
	}

	// 验证租户一致性
	unallocatedTenantID := "00000000-0000-0000-0000-000000000000"
	if !dsTenantID.Valid || dsTenantID.String == unallocatedTenantID {
		return "", fmt.Errorf("device_store not allocated to tenant: device_store_id=%s (device must be allocated before checkout)", deviceStoreID)
	}
	if dsTenantID.String != tenantID {
		return "", fmt.Errorf("device_store belongs to different tenant: device_store_id=%s (expected %s, got %s)", deviceStoreID, tenantID, dsTenantID.String)
	}

	// 3. 验证serial_number和uid至少填一个（从device_store获取）
	if !dsSerialNumber.Valid && !dsUID.Valid {
		return "", fmt.Errorf("device_store has no serial_number or uid: device_store_id=%s", deviceStoreID)
	}

	// 4. 验证位置绑定（如果提供）
	boundRoomID := ""
	if device.BoundRoomID.Valid {
		boundRoomID = device.BoundRoomID.String
	}
	boundBedID := ""
	if device.BoundBedID.Valid {
		boundBedID = device.BoundBedID.String
	}

	// 验证bound_room_id（如果提供）
	if boundRoomID != "" {
		var unitTenantID string
		err := tx.QueryRowContext(ctx, `
			SELECT u.tenant_id::text
			FROM rooms r
			JOIN units u ON r.unit_id = u.unit_id
			WHERE r.room_id = $1
		`, boundRoomID).Scan(&unitTenantID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("room not found: room_id=%s (room must belong to an existing unit)", boundRoomID)
			}
			return "", fmt.Errorf("failed to validate room: %w", err)
		}
		if unitTenantID != tenantID {
			return "", fmt.Errorf("room belongs to different tenant: room_id=%s (expected %s, got %s)", boundRoomID, tenantID, unitTenantID)
		}
	}

	// 验证bound_bed_id（如果提供）
	if boundBedID != "" {
		var unitTenantID string
		err := tx.QueryRowContext(ctx, `
			SELECT u.tenant_id::text
			FROM beds b
			JOIN rooms r ON b.room_id = r.room_id
			JOIN units u ON r.unit_id = u.unit_id
			WHERE b.bed_id = $1
		`, boundBedID).Scan(&unitTenantID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("bed not found: bed_id=%s (bed must belong to an existing room)", boundBedID)
			}
			return "", fmt.Errorf("failed to validate bed: %w", err)
		}
		if unitTenantID != tenantID {
			return "", fmt.Errorf("bed belongs to different tenant: bed_id=%s (expected %s, got %s)", boundBedID, tenantID, unitTenantID)
		}
	}

	// 验证不能同时绑定到room和bed
	if boundRoomID != "" && boundBedID != "" {
		return "", fmt.Errorf("cannot bind to both room and bed: bound_room_id=%s, bound_bed_id=%s (mutually exclusive)", boundRoomID, boundBedID)
	}

	// 5. 生成device_name（如果未提供，从device_store生成）
	deviceName := device.DeviceName
	if deviceName == "" {
		if dsSerialNumber.Valid && dsSerialNumber.String != "" {
			deviceName = dsSerialNumber.String
		} else if dsUID.Valid && dsUID.String != "" {
			deviceName = dsUID.String
		} else {
			deviceName = "Device-" + deviceStoreID[:8]
		}
	}

	// 6. 插入devices记录
	insertQuery := `
		INSERT INTO devices (
			tenant_id, device_store_id, device_name,
			serial_number, uid,
			bound_room_id, bound_bed_id,
			status, business_access, monitoring_enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING device_id::text
	`

	var deviceID string
	var boundRoomIDVal, boundBedIDVal interface{}
	if boundRoomID != "" {
		boundRoomIDVal = boundRoomID
	} else {
		boundRoomIDVal = nil
	}
	if boundBedID != "" {
		boundBedIDVal = boundBedID
	} else {
		boundBedIDVal = nil
	}

	status := device.Status
	if status == "" {
		status = "offline"
	}
	businessAccess := device.BusinessAccess
	if businessAccess == "" {
		businessAccess = "pending"
	}

	err = tx.QueryRowContext(ctx, insertQuery,
		tenantID,
		deviceStoreID,
		deviceName,
		dsSerialNumber,
		dsUID,
		boundRoomIDVal,
		boundBedIDVal,
		status,
		businessAccess,
		device.MonitoringEnabled,
	).Scan(&deviceID)
	if err != nil {
		return "", fmt.Errorf("failed to create device: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return deviceID, nil
}

// UpdateDevice 更新设备信息
// 替代触发器: trigger_validate_device_bed_tenant, trigger_validate_device_store_tenant
func (r *PostgresDevicesRepository) UpdateDevice(ctx context.Context, tenantID, deviceID string, device *domain.Device) error {
	if tenantID == "" || deviceID == "" {
		return fmt.Errorf("tenant_id and device_id are required")
	}
	if device == nil {
		return fmt.Errorf("device is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. 验证bound_room_id（如果更新）
	if device.BoundRoomID.Valid && device.BoundRoomID.String != "" {
		var unitTenantID string
		if err := tx.QueryRowContext(ctx,
			"SELECT u.tenant_id::text FROM rooms r JOIN units u ON r.unit_id = u.unit_id WHERE r.room_id = $1",
			device.BoundRoomID.String,
		).Scan(&unitTenantID); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("room not found: room_id=%s", device.BoundRoomID.String)
			}
			return fmt.Errorf("failed to validate bound_room_id: %w", err)
		}
		if unitTenantID != tenantID {
			return fmt.Errorf("room not found: room_id=%s (room's unit belongs to tenant %s, expected %s)", device.BoundRoomID.String, unitTenantID, tenantID)
		}
	}

	// 2. 验证bound_bed_id（如果更新）
	if device.BoundBedID.Valid && device.BoundBedID.String != "" {
		var unitTenantID string
		if err := tx.QueryRowContext(ctx,
			"SELECT u.tenant_id::text FROM beds b JOIN rooms r ON b.room_id = r.room_id JOIN units u ON r.unit_id = u.unit_id WHERE b.bed_id = $1",
			device.BoundBedID.String,
		).Scan(&unitTenantID); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("bed not found: bed_id=%s", device.BoundBedID.String)
			}
			return fmt.Errorf("failed to validate bound_bed_id: %w", err)
		}
		if unitTenantID != tenantID {
			return fmt.Errorf("bed not found: bed_id=%s (bed's room's unit belongs to tenant %s, expected %s)", device.BoundBedID.String, unitTenantID, tenantID)
		}
	}

	// 3. 验证device_store_id的租户一致性（如果更新）
	if device.DeviceStoreID.Valid && device.DeviceStoreID.String != "" {
		var dsTenantID sql.NullString
		if err := tx.QueryRowContext(ctx,
			"SELECT tenant_id FROM device_store WHERE device_store_id = $1",
			device.DeviceStoreID.String,
		).Scan(&dsTenantID); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("device_store not found: device_store_id=%s", device.DeviceStoreID.String)
			}
			return fmt.Errorf("failed to validate device_store_id: %w", err)
		}
		unallocatedTenantID := "00000000-0000-0000-0000-000000000000"
		if dsTenantID.Valid && dsTenantID.String != unallocatedTenantID && dsTenantID.String != tenantID {
			return fmt.Errorf("device_store_id %s is assigned to a different tenant (expected %s, got %s)", device.DeviceStoreID.String, tenantID, dsTenantID.String)
		}
	}

	// 4. 执行UPDATE（动态构建）
	set := []string{}
	args := []any{tenantID, deviceID}
	argN := 3
	add := func(col string, v any) {
		set = append(set, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, v)
		argN++
	}

	if device.DeviceName != "" {
		add("device_name", device.DeviceName)
	}
	if device.BusinessAccess != "" {
		add("business_access", device.BusinessAccess)
	}
	if device.Status != "" {
		add("status", device.Status)
	}
	add("monitoring_enabled", device.MonitoringEnabled)
	if device.BoundRoomID.Valid {
		if device.BoundRoomID.String != "" {
			add("bound_room_id", device.BoundRoomID.String)
		} else {
			set = append(set, "bound_room_id = NULL")
		}
	}
	if device.BoundBedID.Valid {
		if device.BoundBedID.String != "" {
			add("bound_bed_id", device.BoundBedID.String)
		} else {
			set = append(set, "bound_bed_id = NULL")
		}
	}
	if device.DeviceStoreID.Valid {
		if device.DeviceStoreID.String != "" {
			add("device_store_id", device.DeviceStoreID.String)
		} else {
			set = append(set, "device_store_id = NULL")
		}
	}
	if device.Metadata.Valid {
		set = append(set, fmt.Sprintf("metadata = $%d::jsonb", argN))
		args = append(args, device.Metadata.String)
		argN++
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

// DeleteDevice 删除设备与位置的绑定关系（设备退回）
// 验证：检查设备是否已使用（is_device_used()），如果已使用，只能软删除（DisableDevice）
func (r *PostgresDevicesRepository) DeleteDevice(ctx context.Context, tenantID, deviceID string) error {
	// 1. 检查设备是否存在
	var deviceExists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM devices WHERE tenant_id = $1 AND device_id = $2)
	`, tenantID, deviceID).Scan(&deviceExists)
	if err != nil {
		return fmt.Errorf("failed to check device: %w", err)
	}
	if !deviceExists {
		return fmt.Errorf("device not found: tenant_id=%s, device_id=%s", tenantID, deviceID)
	}

	// 2. 检查设备是否已使用（调用is_device_used()函数）
	var isUsed bool
	err = r.db.QueryRowContext(ctx, `SELECT is_device_used($1)`, deviceID).Scan(&isUsed)
	if err != nil {
		return fmt.Errorf("failed to check if device is used: %w", err)
	}

	if isUsed {
		return fmt.Errorf("cannot delete device: device has reported data (use DisableDevice for soft delete): device_id=%s", deviceID)
	}

	// 3. 物理删除设备记录
	_, err = r.db.ExecContext(ctx, `
		DELETE FROM devices
		WHERE tenant_id = $1 AND device_id = $2
	`, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	return nil
}

// DisableDevice 软删除设备
// 功能：设置status='disabled', business_access='rejected', monitoring_enabled=FALSE
func (r *PostgresDevicesRepository) DisableDevice(ctx context.Context, tenantID, deviceID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE devices
		SET status='disabled', business_access='rejected', monitoring_enabled=FALSE
		WHERE tenant_id=$1 AND device_id=$2
	`, tenantID, deviceID)
	return err
}

// GetDeviceRelations 获取设备关联关系（设备、地址、住户）
func (r *PostgresDevicesRepository) GetDeviceRelations(ctx context.Context, tenantID, deviceID string) (*DeviceRelations, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// 查询设备基本信息、关联的 unit 信息
	query := `
		SELECT 
			d.device_id,
			d.device_name,
			COALESCE(d.serial_number, '') as device_internal_code,
			CASE 
				WHEN ds.device_type = 'Sleepad' THEN 0
				WHEN ds.device_type = 'Radar' THEN 1
				ELSE 0
			END as device_type,
			COALESCE(u.unit_id::text, '') as address_id,
			COALESCE(u.unit_name, '') as address_name,
			CASE 
				WHEN u.unit_type = 'Facility' THEN 0
				WHEN u.unit_type = 'Home' THEN 1
				ELSE 0
			END as address_type
		FROM devices d
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON (d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id) 
			OR (b.room_id = r.room_id AND b.tenant_id = r.tenant_id)
		LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE d.device_id = $1 AND d.tenant_id = $2
	`

	var relations DeviceRelations
	var deviceType, addressType int
	err := r.db.QueryRowContext(ctx, query, deviceID, tenantID).Scan(
		&relations.DeviceID,
		&relations.DeviceName,
		&relations.DeviceInternalCode,
		&deviceType,
		&relations.AddressID,
		&relations.AddressName,
		&addressType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: tenant_id=%s, device_id=%s", tenantID, deviceID)
		}
		return nil, fmt.Errorf("failed to get device relations: %w", err)
	}
	relations.DeviceType = deviceType
	relations.AddressType = addressType

	// 查询关联的住户信息（通过 bed 或 room）
	residentsQuery := `
		SELECT 
			res.resident_id::text,
			res.nickname,
			COALESCE(res.metadata->>'gender', '') as gender,
			COALESCE(res.metadata->>'birthday', '') as birthday
		FROM residents res
		WHERE res.tenant_id = $1
		  AND (
			-- 通过 bed 关联
			(res.bed_id IN (
				SELECT b.bed_id 
				FROM devices d
				JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
				WHERE d.device_id = $2 AND d.tenant_id = $1
			))
			OR
			-- 通过 room 关联
			(res.room_id IN (
				SELECT r.room_id
				FROM devices d
				LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
				LEFT JOIN rooms r ON (d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id)
					OR (b.room_id = r.room_id AND b.tenant_id = r.tenant_id)
				WHERE d.device_id = $2 AND d.tenant_id = $1 AND r.room_id IS NOT NULL
			))
		  )
		ORDER BY res.nickname
	`

	rows, err := r.db.QueryContext(ctx, residentsQuery, tenantID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query residents: %w", err)
	}
	defer rows.Close()

	relations.Residents = []DeviceRelationResident{}
	for rows.Next() {
		var resident DeviceRelationResident
		if err := rows.Scan(
			&resident.ID,
			&resident.Name,
			&resident.Gender,
			&resident.Birthday,
		); err != nil {
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}
		relations.Residents = append(relations.Residents, resident)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate residents: %w", err)
	}

	return &relations, nil
}

// GetOrCreateDeviceFromStore 首次连接时自动创建设备记录
// 替代触发器: trigger_validate_device_identifier, trigger_validate_device_store_tenant
func (r *PostgresDevicesRepository) GetOrCreateDeviceFromStore(ctx context.Context, identifier string, mqttTopic string) (*domain.Device, error) {
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

	// 1. 先尝试从devices表查询
	deviceQuery := `
		SELECT
			d.device_id::text,
			d.tenant_id::text,
			d.device_store_id,
			d.device_name,
			d.serial_number,
			d.uid,
			d.bound_room_id,
			d.bound_bed_id,
			d.status,
			d.business_access,
			d.monitoring_enabled,
			d.metadata
		FROM devices d
		WHERE (d.serial_number = $1 OR d.uid = $1)
		LIMIT 1
	`

	var d domain.Device
	err := r.db.QueryRowContext(ctx, deviceQuery, identifier).Scan(
		&d.DeviceID,
		&d.TenantID,
		&d.DeviceStoreID,
		&d.DeviceName,
		&d.SerialNumber,
		&d.UID,
		&d.BoundRoomID,
		&d.BoundBedID,
		&d.Status,
		&d.BusinessAccess,
		&d.MonitoringEnabled,
		&d.Metadata,
	)

	if err == nil {
		return &d, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	// 2. 从device_store表查询
	unallocatedTenantID := "00000000-0000-0000-0000-000000000000"
	deviceStoreQuery := `
		SELECT
			device_store_id::text,
			device_type,
			serial_number,
			uid,
			tenant_id::text,
			allow_access
		FROM device_store
		WHERE (serial_number = $1 OR uid = $1)
		LIMIT 1
	`

	var dsDeviceStoreID, dsDeviceType, dsTenantID string
	var dsSerialNumber, dsUID sql.NullString
	var dsAllowAccess bool

	err = r.db.QueryRowContext(ctx, deviceStoreQuery, identifier).Scan(
		&dsDeviceStoreID,
		&dsDeviceType,
		&dsSerialNumber,
		&dsUID,
		&dsTenantID,
		&dsAllowAccess,
	)

	if err == sql.ErrNoRows {
		logWarn("Unauthorized device connection attempt",
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("reason", "device_not_registered"),
		)
		return nil, fmt.Errorf("unauthorized device: not registered in device_store")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query device_store: %w", err)
	}

	// 3. 检查设备是否已分配给租户
	if dsTenantID == unallocatedTenantID {
		logWarn("Device connection rejected: not allocated",
			zap.String("device_store_id", dsDeviceStoreID),
			zap.String("reason", "device_not_allocated"),
		)
		return nil, fmt.Errorf("device not allocated to tenant")
	}

	// 4. 验证serial_number和uid至少填一个
	if (!dsSerialNumber.Valid || dsSerialNumber.String == "") && (!dsUID.Valid || dsUID.String == "") {
		logWarn("Device connection rejected: missing identifier",
			zap.String("device_store_id", dsDeviceStoreID),
			zap.String("identifier", identifier),
		)
		return nil, fmt.Errorf("device must have at least one identifier: serial_number or uid")
	}

	// 5. 创建设备记录
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	deviceName := dsDeviceType
	if dsSerialNumber.Valid && dsSerialNumber.String != "" {
		deviceName = dsDeviceType + "-" + dsSerialNumber.String
	} else if dsUID.Valid && dsUID.String != "" {
		deviceName = dsDeviceType + "-" + dsUID.String
	}

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

	logInfo("Device auto-created from device_store",
		zap.String("device_store_id", dsDeviceStoreID),
		zap.String("device_id", newDeviceID),
		zap.String("tenant_id", dsTenantID),
		zap.String("device_type", dsDeviceType),
	)

	return r.GetDevice(ctx, dsTenantID, newDeviceID)
}

