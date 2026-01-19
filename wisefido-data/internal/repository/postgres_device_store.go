package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"wisefido-data/internal/domain"
)

// PostgresDeviceStoreRepository 设备库存Repository实现（强类型）
type PostgresDeviceStoreRepository struct {
	db *sql.DB
}

// NewPostgresDeviceStoreRepository 创建设备库存Repository
func NewPostgresDeviceStoreRepository(db *sql.DB) *PostgresDeviceStoreRepository {
	return &PostgresDeviceStoreRepository{db: db}
}

// ListDeviceStores 查询设备库存列表
func (r *PostgresDeviceStoreRepository) ListDeviceStores(ctx context.Context, filters DeviceStoreFilters, page, size int) ([]*domain.DeviceStore, int, error) {
	where := []string{}
	args := []any{}
	argN := 1

	// Search filter
	if filters.Search != "" {
		where = append(where, fmt.Sprintf("(ds.device_uid ILIKE $%d OR ds.mac ILIKE $%d OR ds.imei ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+filters.Search+"%")
		argN++
	}

	// Tenant filter
	if filters.TenantID != "" {
		where = append(where, fmt.Sprintf("ds.tenant_id = $%d", argN))
		args = append(args, filters.TenantID)
		argN++
	}

	// Device type filter
	if filters.DeviceType != "" {
		where = append(where, fmt.Sprintf("ds.device_type = $%d", argN))
		args = append(args, filters.DeviceType)
		argN++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count total
	queryCount := `
		SELECT COUNT(*)
		FROM device_store ds
		LEFT JOIN tenants t ON ds.tenant_id = t.tenant_id
		` + whereClause

	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Pagination
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	offset := (page - 1) * size

	argsList := append(args, size, offset)
	limitPos := argN
	offsetPos := argN + 1

	// Query data
	query := `
		SELECT
			ds.device_id::text,
			ds.device_uid,
			ds.device_code,
			ds.device_type,
			ds.device_model,
			ds.mac,
			ds.imei,
			ds.comm_mode,
			ds.mcu_model,
			ds.firmware_version,
			ds.ota_target_firmware_version,
			ds.ota_target_mcu_model,
			ds.tenant_id::text,
			COALESCE(t.tenant_name, '') as tenant_name,
			ds.allow_access,
			ds.import_date,
			ds.allocate_time
		FROM device_store ds
		LEFT JOIN tenants t ON ds.tenant_id = t.tenant_id
		` + whereClause + `
		ORDER BY ds.import_date DESC, ds.device_type, ds.device_uid
		LIMIT $` + fmt.Sprintf("%d", limitPos) + ` OFFSET $` + fmt.Sprintf("%d", offsetPos)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []*domain.DeviceStore{}
	for rows.Next() {
		var d domain.DeviceStore
		var deviceCode sql.NullString
		if err := rows.Scan(
			&d.DeviceID,
			&d.DeviceUID,
			&deviceCode,
			&d.DeviceType,
			&d.DeviceModel,
			&d.MAC,
			&d.IMEI,
			&d.CommMode,
			&d.MCUModel,
			&d.FirmwareVersion,
			&d.OTATargetFirmwareVersion,
			&d.OTATargetMCUModel,
			&d.TenantID,
			&d.TenantName,
			&d.AllowAccess,
			&d.ImportDate,
			&d.AllocateTime,
		); err != nil {
			return nil, 0, err
		}
		if deviceCode.Valid {
			d.DeviceCode = deviceCode
		}
		out = append(out, &d)
	}
	return out, total, rows.Err()
}

// GetDeviceStore 查询单个设备库存
func (r *PostgresDeviceStoreRepository) GetDeviceStore(ctx context.Context, deviceUID string) (*domain.DeviceStore, error) {
	query := `
		SELECT
			ds.device_id::text,
			ds.device_uid,
			ds.device_code,
			ds.device_type,
			ds.device_model,
			ds.mac,
			ds.imei,
			ds.comm_mode,
			ds.mcu_model,
			ds.firmware_version,
			ds.ota_target_firmware_version,
			ds.ota_target_mcu_model,
			ds.tenant_id::text,
			COALESCE(t.tenant_name, '') as tenant_name,
			ds.allow_access,
			ds.import_date,
			ds.allocate_time
		FROM device_store ds
		LEFT JOIN tenants t ON ds.tenant_id = t.tenant_id
		WHERE ds.device_uid = $1
	`

	var d domain.DeviceStore
	var deviceCode sql.NullString
	err := r.db.QueryRowContext(ctx, query, deviceUID).Scan(
		&d.DeviceID,
		&d.DeviceUID,
		&deviceCode,
		&d.DeviceType,
		&d.DeviceModel,
		&d.MAC,
		&d.IMEI,
		&d.CommMode,
		&d.MCUModel,
		&d.FirmwareVersion,
		&d.OTATargetFirmwareVersion,
		&d.OTATargetMCUModel,
		&d.TenantID,
		&d.TenantName,
		&d.AllowAccess,
		&d.ImportDate,
		&d.AllocateTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device_store not found: device_uid=%s", deviceUID)
		}
		return nil, err
	}
	if deviceCode.Valid {
		d.DeviceCode = deviceCode
	}
	return &d, nil
}

// CreateDeviceStore 单个创建设备库存（入库操作）
func (r *PostgresDeviceStoreRepository) CreateDeviceStore(ctx context.Context, deviceStore *domain.DeviceStore) (string, error) {
	if deviceStore == nil {
		return "", fmt.Errorf("device_store is required")
	}

	// 1. 验证必填字段
	if deviceStore.DeviceType == "" {
		return "", fmt.Errorf("device_type is required")
	}
	if deviceStore.DeviceUID == "" {
		return "", fmt.Errorf("device_uid is required")
	}

	// 2. 检查是否已存在
	var existingUID string
	checkQuery := `
		SELECT device_uid
		FROM device_store
		WHERE device_uid = $1
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, checkQuery, deviceStore.DeviceUID).Scan(&existingUID)
	if err == nil {
		return "", fmt.Errorf("device already exists: device_uid=%s", existingUID)
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check existing device: %w", err)
	}

	// 3. 处理tenant_id（如果未提供，使用默认值）
	tenantID := deviceStore.TenantID
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000000"
	}

	// 4. 插入新设备
	// 注意：device_id 是主键，由数据库自动生成（gen_random_uuid()），不需要在 INSERT 中指定
	insertQuery := `
		INSERT INTO device_store (
			device_uid, device_code, device_type, device_model, mac, imei,
			comm_mode, mcu_model, firmware_version,
			tenant_id, allow_access
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING device_id::text
	`

	args := []any{
		deviceStore.DeviceUID,
		nullStringToAny(deviceStore.DeviceCode),
		deviceStore.DeviceType,
		nullStringToAny(deviceStore.DeviceModel),
		nullStringToAny(deviceStore.MAC),
		nullStringToAny(deviceStore.IMEI),
		nullStringToAny(deviceStore.CommMode),
		nullStringToAny(deviceStore.MCUModel),
		nullStringToAny(deviceStore.FirmwareVersion),
		tenantID,
		deviceStore.AllowAccess,
	}

	var deviceID string
	err = r.db.QueryRowContext(ctx, insertQuery, args...).Scan(&deviceID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", fmt.Errorf("device already exists (duplicate device_uid)")
		}
		return "", fmt.Errorf("failed to create device_store: %w", err)
	}

	return deviceID, nil
}

// BatchUpdateDeviceStores 批量更新设备库存
func (r *PostgresDeviceStoreRepository) BatchUpdateDeviceStores(ctx context.Context, updates []*domain.DeviceStore) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, update := range updates {
		if update == nil || update.DeviceUID == "" {
			continue
		}
		deviceUID := update.DeviceUID

		setParts := []string{}
		args := []any{}
		argN := 1

		// tenant_id
		if update.TenantID != "" {
			setParts = append(setParts, fmt.Sprintf("tenant_id = $%d", argN))
			args = append(args, update.TenantID)
			argN++
		}

		// device_code
		if update.DeviceCode.Valid {
			setParts = append(setParts, fmt.Sprintf("device_code = $%d", argN))
			args = append(args, update.DeviceCode.String)
			argN++
		}

		// ota_target_firmware_version
		if update.OTATargetFirmwareVersion.Valid {
			setParts = append(setParts, fmt.Sprintf("ota_target_firmware_version = $%d", argN))
			args = append(args, update.OTATargetFirmwareVersion.String)
			argN++
		}

		// ota_target_mcu_model
		if update.OTATargetMCUModel.Valid {
			setParts = append(setParts, fmt.Sprintf("ota_target_mcu_model = $%d", argN))
			args = append(args, update.OTATargetMCUModel.String)
			argN++
		}

		// allow_access
		setParts = append(setParts, fmt.Sprintf("allow_access = $%d", argN))
		args = append(args, update.AllowAccess)
		argN++

		// Update allocate_time when tenant_id is set
		if update.TenantID != "" && update.TenantID != "00000000-0000-0000-0000-000000000000" {
			setParts = append(setParts, "allocate_time = CASE WHEN allocate_time IS NULL THEN CURRENT_TIMESTAMP ELSE allocate_time END")
		}

		if len(setParts) == 0 {
			continue
		}

		query := fmt.Sprintf(`
			UPDATE device_store
			SET %s
			WHERE device_uid = $%d
		`, strings.Join(setParts, ", "), argN)
		args = append(args, deviceUID)

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteDeviceStore 删除设备库存
func (r *PostgresDeviceStoreRepository) DeleteDeviceStore(ctx context.Context, deviceUID string) error {
	// 1. 检查设备是否已分配给租户
	var tenantID string
	err := r.db.QueryRowContext(ctx, `
		SELECT tenant_id::text
		FROM device_store
		WHERE device_uid = $1
	`, deviceUID).Scan(&tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("device_store not found: device_uid=%s", deviceUID)
		}
		return fmt.Errorf("failed to query device_store: %w", err)
	}

	unallocatedTenantID := "00000000-0000-0000-0000-000000000000"
	if tenantID != unallocatedTenantID {
		return fmt.Errorf("cannot delete device_store: device is allocated to tenant %s (must unallocate first)", tenantID)
	}

	// 2. 检查设备是否已出库（devices表中有记录）
	var deviceCount int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM devices
		WHERE device_uid = $1
	`, deviceUID).Scan(&deviceCount)
	if err != nil {
		return fmt.Errorf("failed to check devices: %w", err)
	}
	if deviceCount > 0 {
		return fmt.Errorf("cannot delete device_store: device has been checked out (devices table has %d records)", deviceCount)
	}

	// 3. 删除设备库存记录
	_, err = r.db.ExecContext(ctx, `
		DELETE FROM device_store
		WHERE device_uid = $1
	`, deviceUID)
	if err != nil {
		return fmt.Errorf("failed to delete device_store: %w", err)
	}

	return nil
}

// ImportDeviceStores 批量导入设备库存
func (r *PostgresDeviceStoreRepository) ImportDeviceStores(ctx context.Context, items []*domain.DeviceStore) (int, []*domain.DeviceStore, []*domain.DeviceStore, error) {
	if len(items) == 0 {
		return 0, nil, nil, nil
	}

	var successCount int
	var errors []*domain.DeviceStore
	var skipped []*domain.DeviceStore

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	defer tx.Rollback()

	for _, item := range items {
		if item == nil {
			continue
		}

		// Validate required fields
		if item.DeviceType == "" {
			errors = append(errors, item)
			continue
		}
		if item.DeviceUID == "" {
			errors = append(errors, item)
			continue
		}

		// Check if device already exists
		var existingUID string
		checkQuery := `
			SELECT device_uid
			FROM device_store
			WHERE device_uid = $1
			LIMIT 1
		`
		err := tx.QueryRowContext(ctx, checkQuery, item.DeviceUID).Scan(&existingUID)
		if err == nil {
			skipped = append(skipped, item)
			continue
		} else if err != sql.ErrNoRows {
			errors = append(errors, item)
			continue
		}

		// Get tenant_id (use default if not provided)
		tenantID := item.TenantID
		if tenantID == "" {
			tenantID = "00000000-0000-0000-0000-000000000000"
		}

		// Insert new device
		// 注意：device_id 是主键，由数据库自动生成（gen_random_uuid()），不需要在 INSERT 中指定
		insertQuery := `
			INSERT INTO device_store (
				device_uid, device_code, device_type, device_model, mac, imei,
				comm_mode, mcu_model, firmware_version,
				tenant_id, allow_access
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		args := []any{
			item.DeviceUID,
			nullStringToAny(item.DeviceCode),
			item.DeviceType,
			nullStringToAny(item.DeviceModel),
			nullStringToAny(item.MAC),
			nullStringToAny(item.IMEI),
			nullStringToAny(item.CommMode),
			nullStringToAny(item.MCUModel),
			nullStringToAny(item.FirmwareVersion),
			tenantID,
			item.AllowAccess,
		}

		_, err = tx.ExecContext(ctx, insertQuery, args...)
		if err != nil {
			errors = append(errors, item)
			continue
		}

		successCount++
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, nil, err
	}

	return successCount, skipped, errors, nil
}

// Helper function to convert sql.NullString to any (already defined in postgres_units.go)
// Using the same function from postgres_units.go

