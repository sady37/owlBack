package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"owl-common/alarm"
	"wisefido-qinglan/internal/domain"
)

// PostgresDeviceRepository 设备Repository的PostgreSQL实现
type PostgresDeviceRepository struct {
	db              *sql.DB
	enablementCache *sync.Map // 报警使能配置缓存 (key: "tenant_id:device_uid", value: []alarm.AlarmEnablementItem)
}

// NewPostgresDeviceRepository 创建PostgreSQL设备Repository
func NewPostgresDeviceRepository(db *sql.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{
		db:              db,
		enablementCache: &sync.Map{},
	}
}

// 确保实现了接口
var _ DeviceRepository = (*PostgresDeviceRepository)(nil)

// GetDeviceByUID 根据设备UID获取设备信息
func (r *PostgresDeviceRepository) GetDeviceByUID(ctx context.Context, uid string) (*domain.Device, error) {
	query := `
		SELECT 
			d.device_id, d.device_uid, d.tenant_id, d.device_name,
			d.bound_room_id, d.bound_bed_id, d.status, d.business_access,
			d.monitoring_enabled, d.metadata,
			ds.device_type, ds.device_model, ds.imei, ds.comm_mode,
			ds.mcu_model, ds.firmware_version
		FROM devices d
		LEFT JOIN device_store ds ON d.device_id = ds.device_id
		WHERE d.device_uid = $1 AND d.status != 'disabled'
	`

	row := r.db.QueryRowContext(ctx, query, uid)

	var device domain.Device
	var boundRoomID, boundBedID sql.NullString
	var metadata sql.NullString
	var deviceType, deviceModel, imei, commMode, mcuModel, firmwareVersion sql.NullString

	err := row.Scan(
		&device.DeviceID,
		&device.DeviceUID,
		&device.TenantID,
		&device.DeviceName,
		&boundRoomID,
		&boundBedID,
		&device.Status,
		&device.BusinessAccess,
		&device.MonitoringEnabled,
		&metadata,
		&deviceType,
		&deviceModel,
		&imei,
		&commMode,
		&mcuModel,
		&firmwareVersion,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", uid)
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	if boundRoomID.Valid {
		device.BoundRoomID = boundRoomID
	}
	if boundBedID.Valid {
		device.BoundBedID = boundBedID
	}
	if metadata.Valid {
		device.Metadata = metadata
	}
	if deviceType.Valid {
		device.DeviceType = deviceType
	}
	if deviceModel.Valid {
		device.DeviceModel = deviceModel
	}
	if imei.Valid {
		device.IMEI = imei
	}
	if commMode.Valid {
		device.CommMode = commMode
	}
	if mcuModel.Valid {
		device.MCUModel = mcuModel
	}
	if firmwareVersion.Valid {
		device.FirmwareVersion = firmwareVersion
	}

	return &device, nil
}

// UpdateDeviceMonitoring 更新设备监控状态
func (r *PostgresDeviceRepository) UpdateDeviceMonitoring(ctx context.Context, uid string, enabled bool) error {
	query := `
		UPDATE devices 
		SET monitoring_enabled = $1
		WHERE device_uid = $2 AND status != 'disabled'
	`

	result, err := r.db.ExecContext(ctx, query, enabled, uid)
	if err != nil {
		return fmt.Errorf("failed to update device monitoring: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device not found or is disabled: %s", uid)
	}

	return nil
}

// GetDeviceProperties 获取设备属性（从metadata字段解析）
func (r *PostgresDeviceRepository) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	query := `
		SELECT metadata
		FROM devices
		WHERE device_uid = $1 AND status != 'disabled'
	`

	row := r.db.QueryRowContext(ctx, query, uid)

	var metadata sql.NullString
	err := row.Scan(&metadata)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", uid)
		}
		return nil, fmt.Errorf("failed to query device metadata: %w", err)
	}

	if !metadata.Valid {
		return make(map[string]interface{}), nil
	}

	// 解析JSON metadata
	var metadataMap map[string]interface{}
	if err := json.Unmarshal([]byte(metadata.String), &metadataMap); err != nil {
		return nil, fmt.Errorf("failed to parse device metadata: %w", err)
	}

	// 如果指定了keys，只返回指定的属性
	if len(keys) > 0 {
		result := make(map[string]interface{})
		for _, key := range keys {
			if value, ok := metadataMap[key]; ok {
				result[key] = value
			}
		}
		return result, nil
	}

	return metadataMap, nil
}

// SetDeviceProperties 设置设备属性（更新metadata字段）
func (r *PostgresDeviceRepository) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
	// 获取当前metadata
	currentMetadata, err := r.GetDeviceProperties(ctx, uid, []string{})
	if err != nil {
		return err
	}

	// 合并属性
	for key, value := range properties {
		currentMetadata[key] = value
	}

	// 序列化为JSON
	metadataJSON, err := json.Marshal(currentMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE devices 
		SET metadata = $1
		WHERE device_uid = $2 AND status != 'disabled'
	`

	result, err := r.db.ExecContext(ctx, query, string(metadataJSON), uid)
	if err != nil {
		return fmt.Errorf("failed to set device properties: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device not found or is disabled: %s", uid)
	}

	return nil
}

// GetAllDeviceStoreInfo 获取所有device_store记录（用于启动时初始化设备缓存）
func (r *PostgresDeviceRepository) GetAllDeviceStoreInfo(ctx context.Context) ([]*DeviceStoreInfo, error) {
	query := `
		SELECT device_id, device_uid, device_code, device_type, device_model, mac, imei, comm_mode, mcu_model, firmware_version, tenant_id, allow_access
		FROM device_store
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query device_store: %w", err)
	}
	defer rows.Close()

	var results []*DeviceStoreInfo
	for rows.Next() {
		var d DeviceStoreInfo
		var deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
		var deviceCode sql.NullString
		var allowAccess bool

		if err := rows.Scan(&d.DeviceID, &d.DeviceUID, &deviceCode, &d.DeviceType, &deviceModel, &mac, &imei, &commMode, &mcuModel, &firmwareVersion, &d.TenantID, &allowAccess); err != nil {
			return nil, fmt.Errorf("failed to scan device_store row: %w", err)
		}

		d.DeviceCode = deviceCode
		d.DeviceModel = deviceModel
		d.MAC = mac
		d.IMEI = imei
		d.CommMode = commMode
		d.MCUModel = mcuModel
		d.FirmwareVersion = firmwareVersion
		d.AllowAccess = allowAccess

		results = append(results, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("device_store rows iteration error: %w", err)
	}

	return results, nil
}

// GetDevicesByTenant 根据租户获取设备列表
func (r *PostgresDeviceRepository) GetDevicesByTenant(ctx context.Context, tenantID string) ([]*domain.Device, error) {
	query := `
		SELECT 
			d.device_id, d.device_uid, d.tenant_id, d.device_name,
			d.bound_room_id, d.bound_bed_id, d.status, d.business_access,
			d.monitoring_enabled, d.metadata,
			ds.device_type, ds.device_model, ds.imei, ds.comm_mode,
			ds.mcu_model, ds.firmware_version
		FROM devices d
		LEFT JOIN device_store ds ON d.device_id = ds.device_id
		WHERE d.tenant_id = $1 AND d.status != 'disabled'
		ORDER BY d.device_name
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		var device domain.Device
		var boundRoomID, boundBedID sql.NullString
		var metadata sql.NullString
		var deviceType, deviceModel, imei, commMode, mcuModel, firmwareVersion sql.NullString

		err := rows.Scan(
			&device.DeviceID,
			&device.DeviceUID,
			&device.TenantID,
			&device.DeviceName,
			&boundRoomID,
			&boundBedID,
			&device.Status,
			&device.BusinessAccess,
			&device.MonitoringEnabled,
			&metadata,
			&deviceType,
			&deviceModel,
			&imei,
			&commMode,
			&mcuModel,
			&firmwareVersion,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}

		if boundRoomID.Valid {
			device.BoundRoomID = boundRoomID
		}
		if boundBedID.Valid {
			device.BoundBedID = boundBedID
		}
		if metadata.Valid {
			device.Metadata = metadata
		}
		if deviceType.Valid {
			device.DeviceType = deviceType
		}
		if deviceModel.Valid {
			device.DeviceModel = deviceModel
		}
		if imei.Valid {
			device.IMEI = imei
		}
		if commMode.Valid {
			device.CommMode = commMode
		}
		if mcuModel.Valid {
			device.MCUModel = mcuModel
		}
		if firmwareVersion.Valid {
			device.FirmwareVersion = firmwareVersion
		}

		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	return devices, nil
}

// CreateDevice 创建设备
func (r *PostgresDeviceRepository) CreateDevice(ctx context.Context, device *domain.Device) error {
	// 检查device_store中是否存在该设备
	checkQuery := `
		SELECT device_id, tenant_id, allow_access
		FROM device_store
		WHERE device_uid = $1
	`

	var storeDeviceID, storeTenantID string
	var allowAccess bool
	err := r.db.QueryRowContext(ctx, checkQuery, device.DeviceUID).Scan(&storeDeviceID, &storeTenantID, &allowAccess)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("device not found in device_store: %s", device.DeviceUID)
		}
		return fmt.Errorf("failed to check device_store: %w", err)
	}

	// 验证租户匹配
	if storeTenantID != device.TenantID {
		return fmt.Errorf("device belongs to different tenant: %s", storeTenantID)
	}

	// 验证系统级权限
	if !allowAccess {
		return fmt.Errorf("device not allowed to access system")
	}

	// 插入设备记录
	// 初始化时：business_access='pending', status='offline'（因为此时仍是拒绝状态）
	query := `
		INSERT INTO devices (
			device_id, device_uid, tenant_id, device_name,
			bound_room_id, bound_bed_id, status, business_access,
			monitoring_enabled, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var boundRoomID, boundBedID interface{}
	if device.BoundRoomID.Valid {
		boundRoomID = device.BoundRoomID.String
	} else {
		boundRoomID = nil
	}

	if device.BoundBedID.Valid {
		boundBedID = device.BoundBedID.String
	} else {
		boundBedID = nil
	}

	var metadata interface{}
	if device.Metadata.Valid {
		metadata = device.Metadata.String
	} else {
		metadata = nil
	}

	// 初始化时设置默认值：status='offline', business_access='pending'
	status := device.Status
	if status == "" {
		status = "offline"
	}
	businessAccess := device.BusinessAccess
	if businessAccess == "" {
		businessAccess = "pending"
	}

	_, err = r.db.ExecContext(ctx, query,
		storeDeviceID, // device_id 使用 device_store 的 device_id
		device.DeviceUID,
		device.TenantID,
		device.DeviceName,
		boundRoomID,
		boundBedID,
		status,
		businessAccess,
		device.MonitoringEnabled,
		metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	// 更新 device 的 DeviceID
	device.DeviceID = storeDeviceID

	return nil
}

// UpdateDevice 更新设备信息
func (r *PostgresDeviceRepository) UpdateDevice(ctx context.Context, device *domain.Device) error {
	// 构建动态更新语句
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	// 检查哪些字段需要更新
	if device.DeviceName != "" {
		updates = append(updates, fmt.Sprintf("device_name = $%d", argIdx))
		args = append(args, device.DeviceName)
		argIdx++
	}

	if device.BoundRoomID.Valid {
		updates = append(updates, fmt.Sprintf("bound_room_id = $%d", argIdx))
		args = append(args, device.BoundRoomID.String)
		argIdx++
	} else {
		// 明确设置为 NULL
		updates = append(updates, "bound_room_id = NULL")
	}

	if device.BoundBedID.Valid {
		updates = append(updates, fmt.Sprintf("bound_bed_id = $%d", argIdx))
		args = append(args, device.BoundBedID.String)
		argIdx++
	} else {
		// 明确设置为 NULL
		updates = append(updates, "bound_bed_id = NULL")
	}

	if device.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, device.Status)
		argIdx++
	}

	if device.BusinessAccess != "" {
		updates = append(updates, fmt.Sprintf("business_access = $%d", argIdx))
		args = append(args, device.BusinessAccess)
		argIdx++
	}

	updates = append(updates, fmt.Sprintf("monitoring_enabled = $%d", argIdx))
	args = append(args, device.MonitoringEnabled)
	argIdx++

	if device.Metadata.Valid {
		updates = append(updates, fmt.Sprintf("metadata = $%d", argIdx))
		args = append(args, device.Metadata.String)
		argIdx++
	} else {
		// 明确设置为 NULL
		updates = append(updates, "metadata = NULL")
	}

	if len(updates) == 0 {
		return nil // 没有需要更新的字段
	}

	// 添加 WHERE 条件
	args = append(args, device.DeviceUID)
	whereClause := fmt.Sprintf("WHERE device_uid = $%d AND status != 'disabled'", argIdx)

	query := fmt.Sprintf(`
		UPDATE devices 
		SET %s
		%s
	`, strings.Join(updates, ", "), whereClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device not found or is disabled: %s", device.DeviceUID)
	}

	return nil
}

// SearchDevices 搜索设备
func (r *PostgresDeviceRepository) SearchDevices(ctx context.Context, criteria map[string]interface{}) ([]*domain.Device, error) {
	// 构建基础查询
	query := `
		SELECT 
			d.device_id, d.device_uid, d.tenant_id, d.device_name,
			d.bound_room_id, d.bound_bed_id, d.status, d.business_access,
			d.monitoring_enabled, d.metadata,
			ds.device_type, ds.device_model, ds.imei, ds.comm_mode,
			ds.mcu_model, ds.firmware_version
		FROM devices d
		LEFT JOIN device_store ds ON d.device_id = ds.device_id
		WHERE d.status != 'disabled'
	`

	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	// 处理搜索条件
	if tenantID, ok := criteria["tenant_id"].(string); ok && tenantID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.tenant_id = $%d", argIdx))
		args = append(args, tenantID)
		argIdx++
	}

	if status, ok := criteria["status"].(string); ok && status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if businessAccess, ok := criteria["business_access"].(string); ok && businessAccess != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.business_access = $%d", argIdx))
		args = append(args, businessAccess)
		argIdx++
	}

	if monitoringEnabled, ok := criteria["monitoring_enabled"].(bool); ok {
		whereClauses = append(whereClauses, fmt.Sprintf("d.monitoring_enabled = $%d", argIdx))
		args = append(args, monitoringEnabled)
		argIdx++
	}

	if deviceType, ok := criteria["device_type"].(string); ok && deviceType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ds.device_type = $%d", argIdx))
		args = append(args, deviceType)
		argIdx++
	}

	if searchKeyword, ok := criteria["search_keyword"].(string); ok && searchKeyword != "" {
		searchType, _ := criteria["search_type"].(string)
		switch searchType {
		case "device_name":
			whereClauses = append(whereClauses, fmt.Sprintf("d.device_name ILIKE $%d", argIdx))
		case "device_uid":
			whereClauses = append(whereClauses, fmt.Sprintf("d.device_uid ILIKE $%d", argIdx))
		default:
			whereClauses = append(whereClauses, fmt.Sprintf("(d.device_name ILIKE $%d OR d.device_uid ILIKE $%d)", argIdx, argIdx))
		}
		args = append(args, "%"+searchKeyword+"%")
		argIdx++
	}

	// 添加 WHERE 子句
	if len(whereClauses) > 0 {
		query += " AND " + strings.Join(whereClauses, " AND ")
	}

	// 添加排序
	query += " ORDER BY d.device_name"

	// 执行查询
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search devices: %w", err)
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		var device domain.Device
		var boundRoomID, boundBedID sql.NullString
		var metadata sql.NullString
		var deviceType, deviceModel, imei, commMode, mcuModel, firmwareVersion sql.NullString

		err := rows.Scan(
			&device.DeviceID,
			&device.DeviceUID,
			&device.TenantID,
			&device.DeviceName,
			&boundRoomID,
			&boundBedID,
			&device.Status,
			&device.BusinessAccess,
			&device.MonitoringEnabled,
			&metadata,
			&deviceType,
			&deviceModel,
			&imei,
			&commMode,
			&mcuModel,
			&firmwareVersion,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}

		if boundRoomID.Valid {
			device.BoundRoomID = boundRoomID
		}
		if boundBedID.Valid {
			device.BoundBedID = boundBedID
		}
		if metadata.Valid {
			device.Metadata = metadata
		}
		if deviceType.Valid {
			device.DeviceType = deviceType
		}
		if deviceModel.Valid {
			device.DeviceModel = deviceModel
		}
		if imei.Valid {
			device.IMEI = imei
		}
		if commMode.Valid {
			device.CommMode = commMode
		}
		if mcuModel.Valid {
			device.MCUModel = mcuModel
		}
		if firmwareVersion.Valid {
			device.FirmwareVersion = firmwareVersion
		}

		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	return devices, nil
}

// CountDevicesByStatus 按状态统计设备数量
func (r *PostgresDeviceRepository) CountDevicesByStatus(ctx context.Context, tenantID string) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM devices
		WHERE tenant_id = $1 AND status != 'disabled'
		GROUP BY status
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count devices by status: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int

		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan count row: %w", err)
		}

		counts[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating count rows: %w", err)
	}

	return counts, nil
}

// GetDeviceStoreInfo 根据设备UID获取 device_store 信息（含 device_code）
func (r *PostgresDeviceRepository) GetDeviceStoreInfo(ctx context.Context, deviceUID string) (*DeviceStoreInfo, error) {
	query := `
		SELECT 
			device_id, device_uid, device_code, device_type, device_model,
			mac, imei, comm_mode, mcu_model, firmware_version,
			tenant_id, allow_access
		FROM device_store
		WHERE device_uid = $1
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, deviceUID)

	var ds DeviceStoreInfo
	var deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
	var deviceCode sql.NullString

	err := row.Scan(
		&ds.DeviceID, &ds.DeviceUID, &deviceCode, &ds.DeviceType, &deviceModel,
		&mac, &imei, &commMode, &mcuModel, &firmwareVersion,
		&ds.TenantID, &ds.AllowAccess,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found in device_store")
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	ds.DeviceCode = deviceCode
	ds.DeviceModel = deviceModel
	ds.MAC = mac
	ds.IMEI = imei
	ds.CommMode = commMode
	ds.MCUModel = mcuModel
	ds.FirmwareVersion = firmwareVersion

	return &ds, nil
}

// GetDeviceStoreByDeviceID 根据设备ID获取 device_store 信息（用于 device_store 变化信号处理）
func (r *PostgresDeviceRepository) GetDeviceStoreByDeviceID(ctx context.Context, deviceID string) (*DeviceStoreInfo, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	query := `
		SELECT 
			device_id, device_uid, device_code, device_type, device_model,
			mac, imei, comm_mode, mcu_model, firmware_version,
			tenant_id, allow_access
		FROM device_store
		WHERE device_id = $1
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, deviceID)

	var ds DeviceStoreInfo
	var deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
	var deviceCode sql.NullString

	err := row.Scan(
		&ds.DeviceID, &ds.DeviceUID, &deviceCode, &ds.DeviceType, &deviceModel,
		&mac, &imei, &commMode, &mcuModel, &firmwareVersion,
		&ds.TenantID, &ds.AllowAccess,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found in device_store")
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	ds.DeviceCode = deviceCode
	ds.DeviceModel = deviceModel
	ds.MAC = mac
	ds.IMEI = imei
	ds.CommMode = commMode
	ds.MCUModel = mcuModel
	ds.FirmwareVersion = firmwareVersion

	return &ds, nil
}

// 从 alarm_device.monitor_config.alarms 中解析报警项，返回 []AlarmEnablementItem
// 先查缓存，缓存未命中再查数据库并存入缓存
func (r *PostgresDeviceRepository) GetAlarmEnablement(ctx context.Context, tenantID, deviceUID string) ([]alarm.AlarmEnablementItem, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceUID == "" {
		return nil, fmt.Errorf("device_uid is required")
	}

	// 先查缓存
	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceUID)
	if cached, ok := r.enablementCache.Load(cacheKey); ok {
		if enablementMap, ok := cached.([]alarm.AlarmEnablementItem); ok {
			return enablementMap, nil
		}
	}

	// 缓存未命中，查询数据库
	query := `
		SELECT ad.monitor_config
		FROM alarm_device ad
		INNER JOIN devices d ON ad.device_id = d.device_id
		WHERE d.tenant_id = $1 AND d.device_uid = $2
		LIMIT 1
	`

	var monitorConfigJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceUID).Scan(&monitorConfigJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有配置，返回空列表并缓存
			emptyEnablement := []alarm.AlarmEnablementItem{}
			r.enablementCache.Store(cacheKey, emptyEnablement)
			return emptyEnablement, nil
		}
		return nil, fmt.Errorf("failed to query alarm enablement: %w", err)
	}

	if !monitorConfigJSON.Valid || monitorConfigJSON.String == "" {
		// 空配置，返回空列表并缓存
		emptyEnablement := []alarm.AlarmEnablementItem{}
		r.enablementCache.Store(cacheKey, emptyEnablement)
		return emptyEnablement, nil
	}

	var monitorConfig map[string]interface{}
	if err := json.Unmarshal([]byte(monitorConfigJSON.String), &monitorConfig); err != nil {
		return nil, fmt.Errorf("failed to parse monitor_config: %w", err)
	}

	alarmsMap, ok := monitorConfig["alarms"].(map[string]interface{})
	if !ok {
		// 没有alarms配置，返回空列表并缓存
		emptyEnablement := []alarm.AlarmEnablementItem{}
		r.enablementCache.Store(cacheKey, emptyEnablement)
		return emptyEnablement, nil
	}

	// 转换为 AlarmItem 列表
	var alarmItems []alarm.AlarmItem
	for alarmType, alarmConfig := range alarmsMap {
		configMap, ok := alarmConfig.(map[string]interface{})
		if !ok {
			continue
		}

		item := alarm.AlarmItem{
			AlarmType:      alarmType,
			AlarmParams:    make(map[string]interface{}),
			DisplaySetting: 0,
		}

		// 解析 is_enabled 或 enabled（兼容两种字段名）
		var enabled *int
		if isEnabled, ok := configMap["is_enabled"].(float64); ok {
			enabledVal := int(isEnabled)
			enabled = &enabledVal
		} else if isEnabled, ok := configMap["is_enabled"].(int); ok {
			enabled = &isEnabled
		} else if enabledBool, ok := configMap["enabled"].(bool); ok {
			// 兼容 enabled 字段（布尔值）
			enabledVal := 0
			if enabledBool {
				enabledVal = 1
			}
			enabled = &enabledVal
		}
		if enabled != nil {
			item.IsEnabled = enabled
		}

		// 解析 alarm_level 或 level（兼容两种字段名）
		if level, ok := configMap["alarm_level"].(string); ok {
			item.AlarmLevel = &level
		} else if level, ok := configMap["level"].(string); ok {
			// 兼容 level 字段
			item.AlarmLevel = &level
		}

		// 解析 alarm_params
		if params, ok := configMap["alarm_params"].(map[string]interface{}); ok {
			item.AlarmParams = params
		}

		// 解析 display_setting
		if displaySetting, ok := configMap["display_setting"].(float64); ok {
			item.DisplaySetting = int(displaySetting)
		} else if displaySetting, ok := configMap["display_setting"].(int); ok {
			item.DisplaySetting = displaySetting
		}

		alarmItems = append(alarmItems, item)
	}

	// 使用 GetAlarmEnablementMap 转换为 AlarmEnablementItem 列表
	// 设备类型为 "radar"（qinglan 是雷达设备）
	enablementMap := alarm.GetAlarmEnablementMap("radar", alarmItems)

	// 存入缓存
	r.enablementCache.Store(cacheKey, enablementMap)

	return enablementMap, nil
}

// ClearAlarmEnablementCache 清除指定设备的报警使能配置缓存
func (r *PostgresDeviceRepository) ClearAlarmEnablementCache(tenantID, deviceUID string) {
	if tenantID == "" || deviceUID == "" {
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceUID)
	r.enablementCache.Delete(cacheKey)
}

// PreloadAlarmEnablement 预加载指定设备的报警使能配置到缓存
func (r *PostgresDeviceRepository) PreloadAlarmEnablement(ctx context.Context, tenantID, deviceUID string) error {
	_, err := r.GetAlarmEnablement(ctx, tenantID, deviceUID)
	return err
}

// GetAllAccessibleDevices 获取所有可访问的设备（用于启动时主动订阅）
// 条件：device_store.allow_access = TRUE 且 devices.business_access = 'approved' 或 'enable'
// 如果 devices 表中没有记录，也允许（d.business_access IS NULL）
func (r *PostgresDeviceRepository) GetAllAccessibleDevices(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT ds.device_uid
		FROM device_store ds
		LEFT JOIN devices d ON ds.device_id = d.device_id AND d.status != 'disabled'
		WHERE ds.allow_access = TRUE
		  AND (d.business_access = 'approved' OR d.business_access = 'enable' OR d.business_access IS NULL)
		ORDER BY ds.device_uid
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query accessible devices: %w", err)
	}
	defer rows.Close()

	var deviceUIDs []string
	for rows.Next() {
		var deviceUID string
		if err := rows.Scan(&deviceUID); err != nil {
			return nil, fmt.Errorf("failed to scan device_uid: %w", err)
		}
		deviceUIDs = append(deviceUIDs, deviceUID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	return deviceUIDs, nil
}
