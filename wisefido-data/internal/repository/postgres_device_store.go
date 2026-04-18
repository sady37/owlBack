package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"wisefido-data/internal/domain"

	"github.com/lib/pq"
)

// PostgresDeviceStoreRepository 设备库存Repository实现（强类型）
type PostgresDeviceStoreRepository struct {
	db *sql.DB
}

// NewPostgresDeviceStoreRepository 创建设备库存Repository
func NewPostgresDeviceStoreRepository(db *sql.DB) *PostgresDeviceStoreRepository {
	return &PostgresDeviceStoreRepository{db: db}
}

const (
	trashTenantID       = "00000000-0000-0000-0000-000000000000"
	systemTenantID      = "00000000-0000-0000-0000-000000000001"
	unallocatedTenantID = "00000000-0000-0000-0000-000000000002"
)

func isDeviceStorePivotTenant(tenantID string) bool {
	return tenantID == trashTenantID || tenantID == systemTenantID
}

// orderByClauseDeviceStore 白名单排序字段，防止 SQL 注入
func orderByClauseDeviceStore(sort, direction string) string {
	dir := "ASC"
	if strings.TrimSpace(strings.ToUpper(direction)) == "DESC" {
		dir = "DESC"
	}
	col := strings.TrimSpace(strings.ToLower(sort))
	switch col {
	case "device_uid", "device_code":
		return "ds." + col + " " + dir
	case "device_type", "device_model", "device_name", "mac", "imei", "comm_mode", "mcu_model", "firmware_version":
		return "ds." + col + " " + dir
	case "ota_target_firmware_version", "ota_target_mcu_model",
		"ota_permit", "ota_way", "ota_status", "ota_progress", "ota_updated_at":
		return "ds." + col + " " + dir
	case "tenant_id", "allow_access", "import_date", "allocate_time":
		return "ds." + col + " " + dir
	case "tenant_name":
		return "t.tenant_name " + dir
	case "device_id":
		return "ds.device_id " + dir
	default:
		return "ds.import_date DESC, ds.device_type, ds.device_uid"
	}
}

// ListDeviceStores 查询设备库存列表；sort/direction 为全量排序后分页
func (r *PostgresDeviceStoreRepository) ListDeviceStores(ctx context.Context, filters DeviceStoreFilters, page, size int, sort, direction string) ([]*domain.DeviceStore, int, error) {
	where := []string{}
	args := []any{}
	argN := 1

	// Search filter
	if filters.Search != "" {
		where = append(where, fmt.Sprintf("(ds.device_code ILIKE $%d OR ds.device_uid ILIKE $%d OR ds.mac ILIKE $%d OR ds.imei ILIKE $%d OR ds.device_id::text ILIKE $%d)", argN, argN, argN, argN, argN))
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

	// Column filters
	if filters.DeviceUID != "" {
		where = append(where, fmt.Sprintf("ds.device_uid ILIKE $%d", argN))
		args = append(args, "%"+filters.DeviceUID+"%")
		argN++
	}
	if filters.DeviceCode != "" {
		where = append(where, fmt.Sprintf("ds.device_code ILIKE $%d", argN))
		args = append(args, "%"+filters.DeviceCode+"%")
		argN++
	}
	if filters.DeviceName != "" {
		where = append(where, fmt.Sprintf("ds.device_name ILIKE $%d", argN))
		args = append(args, "%"+filters.DeviceName+"%")
		argN++
	}
	if filters.FirmwareVersion != "" {
		where = append(where, fmt.Sprintf("ds.firmware_version ILIKE $%d", argN))
		args = append(args, "%"+filters.FirmwareVersion+"%")
		argN++
	}
	if filters.AllowAccess != "" {
		if strings.EqualFold(filters.AllowAccess, "true") {
			where = append(where, "ds.allow_access = TRUE")
		} else {
			where = append(where, "ds.allow_access = FALSE")
		}
	}
	if filters.OTAStatus != "" {
		where = append(where, fmt.Sprintf("COALESCE(ds.ota_status, 'idle') = $%d", argN))
		args = append(args, filters.OTAStatus)
		argN++
	}
	if filters.OTAPermit != "" {
		where = append(where, fmt.Sprintf("ds.ota_permit = $%d", argN))
		args = append(args, filters.OTAPermit)
		argN++
	}
	if filters.OTAWay != "" {
		where = append(where, fmt.Sprintf("ds.ota_way = $%d", argN))
		args = append(args, filters.OTAWay)
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
			ds.device_name,
			ds.mac,
			ds.imei,
			ds.comm_mode,
			ds.mcu_model,
			ds.firmware_version,
			ds.ota_target_firmware_version,
			ds.ota_target_mcu_model,
			ds.ota_permit, ds.ota_way, ds.ota_schedule, ds.ota_status, ds.ota_progress,
			ds.ota_error, ds.ota_updated_at, ds.ota_firmware_url, ds.ota_firmware_sha256,
			ds.ota_firmware_size, COALESCE(ds.ota_tenant_approved, FALSE) as ota_tenant_approved,
			ds.tenant_id::text,
			COALESCE(t.tenant_name, '') as tenant_name,
			ds.allow_access,
			ds.import_date,
			ds.allocate_time
		FROM device_store ds
		LEFT JOIN tenants t ON ds.tenant_id = t.tenant_id
		` + whereClause + `
		ORDER BY ` + orderByClauseDeviceStore(sort, direction) + `
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
			&d.DeviceName,
			&d.MAC,
			&d.IMEI,
			&d.CommMode,
			&d.MCUModel,
			&d.FirmwareVersion,
			&d.OTATargetFirmwareVersion,
			&d.OTATargetMCUModel,
			&d.OTAPermit, &d.OTAWay, &d.OTASchedule, &d.OTAStatus, &d.OTAProgress,
			&d.OTAError, &d.OTAUpdatedAt, &d.OTAFirmwareURL, &d.OTAFirmwareSHA,
			&d.OTAFirmwareSize, &d.OTATenantApproved,
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
			ds.device_name,
			ds.mac,
			ds.imei,
			ds.comm_mode,
			ds.mcu_model,
			ds.firmware_version,
			ds.ota_target_firmware_version,
			ds.ota_target_mcu_model,
			ds.ota_permit, ds.ota_way, ds.ota_schedule, ds.ota_status, ds.ota_progress,
			ds.ota_error, ds.ota_updated_at, ds.ota_firmware_url, ds.ota_firmware_sha256,
			ds.ota_firmware_size, COALESCE(ds.ota_tenant_approved, FALSE) as ota_tenant_approved,
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
		&d.DeviceName,
		&d.MAC,
		&d.IMEI,
		&d.CommMode,
		&d.MCUModel,
		&d.FirmwareVersion,
		&d.OTATargetFirmwareVersion,
		&d.OTATargetMCUModel,
		&d.OTAPermit, &d.OTAWay, &d.OTASchedule, &d.OTAStatus, &d.OTAProgress,
		&d.OTAError, &d.OTAUpdatedAt, &d.OTAFirmwareURL, &d.OTAFirmwareSHA,
		&d.OTAFirmwareSize, &d.OTATenantApproved,
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

// GetDeviceStoreByDeviceID 按 device_id (UUID) 查询单条设备库存
func (r *PostgresDeviceStoreRepository) GetDeviceStoreByDeviceID(ctx context.Context, deviceID string) (*domain.DeviceStore, error) {
	query := `
		SELECT
			ds.device_id::text,
			ds.device_uid,
			ds.device_code,
			ds.device_type,
			ds.device_model,
			ds.device_name,
			ds.mac,
			ds.imei,
			ds.comm_mode,
			ds.mcu_model,
			ds.firmware_version,
			ds.ota_target_firmware_version,
			ds.ota_target_mcu_model,
			ds.ota_permit, ds.ota_way, ds.ota_schedule, ds.ota_status, ds.ota_progress,
			ds.ota_error, ds.ota_updated_at, ds.ota_firmware_url, ds.ota_firmware_sha256,
			ds.ota_firmware_size, COALESCE(ds.ota_tenant_approved, FALSE) as ota_tenant_approved,
			ds.tenant_id::text,
			COALESCE(t.tenant_name, '') as tenant_name,
			ds.allow_access,
			ds.import_date,
			ds.allocate_time
		FROM device_store ds
		LEFT JOIN tenants t ON ds.tenant_id = t.tenant_id
		WHERE ds.device_id = $1::uuid
	`
	var d domain.DeviceStore
	var deviceCode sql.NullString
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&d.DeviceID,
		&d.DeviceUID,
		&deviceCode,
		&d.DeviceType,
		&d.DeviceModel,
		&d.DeviceName,
		&d.MAC,
		&d.IMEI,
		&d.CommMode,
		&d.MCUModel,
		&d.FirmwareVersion,
		&d.OTATargetFirmwareVersion,
		&d.OTATargetMCUModel,
		&d.OTAPermit, &d.OTAWay, &d.OTASchedule, &d.OTAStatus, &d.OTAProgress,
		&d.OTAError, &d.OTAUpdatedAt, &d.OTAFirmwareURL, &d.OTAFirmwareSHA,
		&d.OTAFirmwareSize, &d.OTATenantApproved,
		&d.TenantID,
		&d.TenantName,
		&d.AllowAccess,
		&d.ImportDate,
		&d.AllocateTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
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

	// 3. 处理 tenant_id（未提供则 System 001）
	tenantID := deviceStore.TenantID
	if tenantID == "" {
		tenantID = systemTenantID
	}

	// 默认 allow_access = true
	if !deviceStore.AllowAccessSet {
		deviceStore.AllowAccess = true
	}

	// 4. 插入新设备
	// 注意：device_id 是主键，由数据库自动生成（gen_random_uuid()），不需要在 INSERT 中指定
	insertQuery := `
		INSERT INTO device_store (
			device_uid, device_code, device_type, device_model, device_name, mac, imei,
			comm_mode, mcu_model, firmware_version,
			tenant_id, allow_access
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING device_id::text
	`

	args := []any{
		deviceStore.DeviceUID,
		nullStringToAny(deviceStore.DeviceCode),
		deviceStore.DeviceType,
		nullStringToAny(deviceStore.DeviceModel),
		nullStringToAny(deviceStore.DeviceName),
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
// 迁移规则：须经由 pivot（system / trash / Unallocated）中转，禁止租户 A 直迁租户 B
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

		// 如果更新 tenant_id，需要验证迁移规则（仅校验，不写库）
		var tenantChanged bool
		var migrateFromPivot bool
		if update.TenantID != "" {
			var currentTenantID string
			err := tx.QueryRowContext(ctx, `SELECT tenant_id::text FROM device_store WHERE device_uid = $1`, deviceUID).Scan(&currentTenantID)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("device_store not found: device_uid=%s", deviceUID)
				}
				return fmt.Errorf("failed to query current tenant_id: %w", err)
			}
			if currentTenantID != update.TenantID {
				tenantChanged = true
				currentIs := isDeviceStorePivotTenant(currentTenantID)
				newIs := isDeviceStorePivotTenant(update.TenantID)
				if !currentIs && !newIs {
					name1, name2 := tenantNamesOrIDs(ctx, tx, currentTenantID, update.TenantID)
					return fmt.Errorf("cannot migrate device directly from %s to %s: move to System first", name1, name2)
				}
				migrateFromPivot = currentIs && !newIs
			}
		}

		setParts := []string{}
		args := []any{}
		argN := 1
		if update.DeviceType != "" {
			setParts = append(setParts, fmt.Sprintf("device_type = $%d", argN))
			args = append(args, update.DeviceType)
			argN++
		}
		if update.DeviceModel.Valid {
			setParts = append(setParts, fmt.Sprintf("device_model = $%d", argN))
			args = append(args, nullStringToAny(update.DeviceModel))
			argN++
		}
		// device_name 不由 device_store 更新，由 devices 表同步
		if update.TenantID != "" {
			setParts = append(setParts, fmt.Sprintf("tenant_id = $%d", argN))
			args = append(args, update.TenantID)
			argN++
		}
		if update.DeviceCodeSet {
			setParts = append(setParts, fmt.Sprintf("device_code = $%d", argN))
			args = append(args, nullStringToAny(update.DeviceCode))
			argN++
		}
		if update.OTATargetFWSet {
			setParts = append(setParts, fmt.Sprintf("ota_target_firmware_version = $%d", argN))
			args = append(args, nullStringToAny(update.OTATargetFirmwareVersion))
			argN++
		}
		if update.OTATargetMCUSet {
			setParts = append(setParts, fmt.Sprintf("ota_target_mcu_model = $%d", argN))
			args = append(args, nullStringToAny(update.OTATargetMCUModel))
			argN++
		}
		if update.AllowAccessSet {
			setParts = append(setParts, fmt.Sprintf("allow_access = $%d", argN))
			args = append(args, update.AllowAccess)
			argN++
		}
		if update.OTAPermitSet {
			setParts = append(setParts, fmt.Sprintf("ota_permit = $%d", argN))
			args = append(args, nullStringToAny(update.OTAPermit))
			argN++
		}
		if update.OTAWaySet {
			setParts = append(setParts, fmt.Sprintf("ota_way = $%d", argN))
			args = append(args, nullStringToAny(update.OTAWay))
			argN++
		}
		if update.OTAScheduleSet {
			setParts = append(setParts, fmt.Sprintf("ota_schedule = $%d", argN))
			args = append(args, nullStringToAny(update.OTASchedule))
			argN++
		}
		if update.OTAStatusSet {
			setParts = append(setParts, fmt.Sprintf("ota_status = $%d", argN))
			args = append(args, nullStringToAny(update.OTAStatus))
			argN++
			setParts = append(setParts, "ota_updated_at = CURRENT_TIMESTAMP")
		}
		// OTA auto-reset: when OTA plan fields change and OTAStatusSet is false, reset status
		otaPlanChanged := update.OTATargetFirmwareVersion.Valid || update.OTATargetMCUModel.Valid ||
			update.OTAPermitSet || update.OTAWaySet || update.OTAScheduleSet
		if otaPlanChanged && !update.OTAStatusSet {
			setParts = append(setParts, "ota_status = 'idle'", "ota_progress = NULL", "ota_error = NULL", "ota_tenant_approved = FALSE")
		}
		if update.TenantID != "" && update.TenantID != trashTenantID && update.TenantID != unallocatedTenantID {
			setParts = append(setParts, "allocate_time = CASE WHEN allocate_time IS NULL THEN CURRENT_TIMESTAMP ELSE allocate_time END")
		}
		if len(setParts) == 0 {
			continue
		}

		// 先更新 device_store，满足触发器 validate_device_store_tenant（devices.tenant_id 须与 device_store.tenant_id 一致）
		query := fmt.Sprintf(`UPDATE device_store SET %s WHERE device_uid = $%d`, strings.Join(setParts, ", "), argN)
		argsDs := append([]any{}, args...)
		argsDs = append(argsDs, deviceUID)
		if _, err := tx.ExecContext(ctx, query, argsDs...); err != nil {
			return err
		}

		// 再同步 devices：tenant_id + business_access='approved' + monitoring_enabled=TRUE + clear bindings
		if tenantChanged {
			_, err := tx.ExecContext(ctx, `
				UPDATE devices
				SET tenant_id = $1, business_access = 'approved', monitoring_enabled = TRUE,
				    bound_room_id = NULL, bound_bed_id = NULL
				WHERE device_id IN (SELECT device_id FROM device_store WHERE device_uid = $2)
			`, update.TenantID, deviceUID)
			if err != nil {
				return fmt.Errorf("failed to update devices table: %w", err)
			}
			if migrateFromPivot {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO devices (device_id, device_uid, tenant_id, device_name, status, business_access, monitoring_enabled)
					SELECT ds.device_id, ds.device_uid, $1,
						COALESCE(NULLIF(TRIM(ds.device_type), ''), 'Device') || '_' || RIGHT(ds.device_uid, 4),
						'offline', 'approved', TRUE
					FROM device_store ds
					WHERE ds.device_uid = $2
					  AND NOT EXISTS (SELECT 1 FROM devices d WHERE d.device_id = ds.device_id)
				`, update.TenantID, deviceUID)
				if err != nil {
					return fmt.Errorf("failed to create devices row on migrate: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// tenantNamesOrIDs 查询 tenants 表取 tenant_name，若不存在则返回 tenant_id
func tenantNamesOrIDs(ctx context.Context, tx *sql.Tx, id1, id2 string) (string, string) {
	rows, err := tx.QueryContext(ctx, `SELECT tenant_id::text, tenant_name FROM tenants WHERE tenant_id IN ($1, $2)`, id1, id2)
	if err != nil {
		return id1, id2
	}
	defer rows.Close()
	m := make(map[string]string)
	for rows.Next() {
		var tid, tname string
		if err := rows.Scan(&tid, &tname); err != nil {
			continue
		}
		m[tid] = tname
	}
	n1, n2 := id1, id2
	if s, ok := m[id1]; ok && s != "" {
		n1 = s
	}
	if s, ok := m[id2]; ok && s != "" {
		n2 = s
	}
	return n1, n2
}

// DeleteDeviceStore 删除设备库存。仅允许 tenant_id = trash(000)；需先迁回 trash 再删。
func (r *PostgresDeviceStoreRepository) DeleteDeviceStore(ctx context.Context, deviceUID string) error {
	var tenantID string
	err := r.db.QueryRowContext(ctx, `SELECT tenant_id::text FROM device_store WHERE device_uid = $1`, deviceUID).Scan(&tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("device_store not found: device_uid=%s", deviceUID)
		}
		return fmt.Errorf("failed to query device_store: %w", err)
	}
	if tenantID != trashTenantID {
		return fmt.Errorf("cannot delete: device must be in trash tenant first (current tenant_id=%s)", tenantID)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM devices WHERE device_uid = $1`, deviceUID)
	if err != nil {
		return fmt.Errorf("failed to delete devices: %w", err)
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM device_store WHERE device_uid = $1`, deviceUID)
	if err != nil {
		return fmt.Errorf("failed to delete device_store: %w", err)
	}
	return tx.Commit()
}

func deviceStoreNormUID(uid string) string {
	return strings.TrimSpace(uid)
}

func deviceStoreNormCode(c sql.NullString) string {
	if !c.Valid {
		return ""
	}
	return strings.TrimSpace(c.String)
}

// ImportDeviceStores 批量导入设备库存；返回成功数、新插入或更新行(含 device_id)、跳过、失败。
// 同一导入内相同 device_uid 多行时后者覆盖前者（cover）；库中已有该 uid 则 UPDATE，否则 INSERT。
func (r *PostgresDeviceStoreRepository) ImportDeviceStores(ctx context.Context, items []*domain.DeviceStore) (successCount int, inserted []*domain.DeviceStore, skipped []*domain.DeviceStore, errors []*domain.DeviceStore, err error) {
	if len(items) == 0 {
		return 0, nil, nil, nil, nil
	}

	var errs []*domain.DeviceStore
	for _, item := range items {
		if item == nil {
			continue
		}
		if item.DeviceType == "" || deviceStoreNormUID(item.DeviceUID) == "" {
			errs = append(errs, item)
		}
	}

	// 后者覆盖前者：按 device_uid 只保留最后一行
	lastByUID := make(map[string]*domain.DeviceStore)
	for _, item := range items {
		if item == nil || item.DeviceType == "" {
			continue
		}
		normUID := deviceStoreNormUID(item.DeviceUID)
		if normUID == "" {
			continue
		}
		lastByUID[normUID] = item
	}
	keys := make([]string, 0, len(lastByUID))
	for k := range lastByUID {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sk []*domain.DeviceStore
	var ins []*domain.DeviceStore

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	defer tx.Rollback()

	for _, k := range keys {
		item := lastByUID[k]
		normUID := k
		normCode := deviceStoreNormCode(item.DeviceCode)
		tenantID := item.TenantID
		if tenantID == "" {
			tenantID = unallocatedTenantID
		}
		codeForInsert := sql.NullString{String: normCode, Valid: normCode != ""}

		args := []any{
			normUID,
			nullStringToAny(codeForInsert),
			item.DeviceType,
			nullStringToAny(item.DeviceModel),
			nullStringToAny(item.DeviceName),
			nullStringToAny(item.MAC),
			nullStringToAny(item.IMEI),
			nullStringToAny(item.CommMode),
			nullStringToAny(item.MCUModel),
			nullStringToAny(item.FirmwareVersion),
			tenantID,
			item.AllowAccess,
		}

		var deviceID string
		err := tx.QueryRowContext(ctx, `SELECT device_id::text FROM device_store WHERE device_uid = $1`, normUID).Scan(&deviceID)
		if err == nil {
			updateQuery := `
				UPDATE device_store SET
					device_code = $2,
					device_type = $3,
					device_model = $4,
					device_name = $5,
					mac = $6,
					imei = $7,
					comm_mode = $8,
					mcu_model = $9,
					firmware_version = $10,
					tenant_id = $11,
					allow_access = $12
				WHERE device_uid = $1
			`
			_, err = tx.ExecContext(ctx, updateQuery, args...)
			if err != nil {
				errs = append(errs, item)
				continue
			}
			successCount++
			ins = append(ins, &domain.DeviceStore{
				DeviceID:   deviceID,
				DeviceUID:  normUID,
				DeviceType: item.DeviceType,
				DeviceCode: codeForInsert,
				TenantID:   tenantID,
			})
			continue
		}
		if err != sql.ErrNoRows {
			errs = append(errs, item)
			continue
		}

		insertQuery := `
			INSERT INTO device_store (
				device_uid, device_code, device_type, device_model, device_name, mac, imei,
				comm_mode, mcu_model, firmware_version,
				tenant_id, allow_access
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING device_id::text
		`
		err = tx.QueryRowContext(ctx, insertQuery, args...).Scan(&deviceID)
		if err != nil {
			errs = append(errs, item)
			continue
		}

		successCount++
		ins = append(ins, &domain.DeviceStore{
			DeviceID:   deviceID,
			DeviceUID:  normUID,
			DeviceType: item.DeviceType,
			DeviceCode: codeForInsert,
			TenantID:   tenantID,
		})
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, nil, nil, err
	}

	return successCount, ins, sk, errs, nil
}

// UpdateDeviceCode 更新 device_code。device_code 仅来自厂家导入文件；绑定/initialize 不提供可写回的值。
func (r *PostgresDeviceStoreRepository) UpdateDeviceCode(ctx context.Context, deviceID, deviceCode string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE device_store SET device_code = $1 WHERE device_id = $2`, deviceCode, deviceID)
	return err
}

// UpdateFirmwareVersion 仅更新 firmware_version（InitialAll 调 bindInfo 后按返回的 deviceVersion 写回）。
func (r *PostgresDeviceStoreRepository) UpdateFirmwareVersion(ctx context.Context, deviceID, firmwareVersion string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE device_store SET firmware_version = $1 WHERE device_id = $2`, firmwareVersion, deviceID)
	return err
}

// Helper function to convert sql.NullString to any (already defined in postgres_units.go)
// Using the same function from postgres_units.go
