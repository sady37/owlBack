package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// DeviceLocationInfo 设备位置信息
// 包含完整的地址层级信息和设备规格信息
// 注意：所有字段都是指针类型，表示可能为空
// 不包含 name 字段（branch_name, building_name, unit_name, room_name, bed_name, device_name）
type DeviceLocationInfo struct {
	// 设备基本信息
	DeviceID          string  `json:"device_id"`
	TenantID          string  `json:"tenant_id"`
	DeviceUID         *string `json:"device_uid,omitempty"`
	Status            *string `json:"status,omitempty"`
	BusinessAccess    *string `json:"business_access,omitempty"`
	MonitoringEnabled *bool   `json:"monitoring_enabled,omitempty"`

	// 位置信息（只包含ID字段，不包含name字段）
	BranchID   *string `json:"branch_id,omitempty"`
	BuildingID *string `json:"building_id,omitempty"`
	UnitID     *string `json:"unit_id,omitempty"`
	RoomID     *string `json:"room_id,omitempty"`
	BedID      *string `json:"bed_id,omitempty"`

	// 设备规格信息（来自 device_store）
	DeviceType      *string `json:"device_type,omitempty"`
	DeviceModel     *string `json:"device_model,omitempty"`
	IMEI            *string `json:"imei,omitempty"`
	MAC             *string `json:"mac,omitempty"`
	CommMode        *string `json:"comm_mode,omitempty"`
	MCUModel        *string `json:"mcu_model,omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty"`
}

// RadarAlarmTypes Radar 设备的所有报警类型列表
var RadarAlarmTypes = []string{
	"OfflineAlarm",
	"LowBattery",
	"DeviceFailure",
	"Radar_ApneaHypopnea",
	"Radar_AbnormalHeartRate",
	"Radar_AbnormalRespiratoryRate",
	"SuspectedFall",
	"Fall",
	"VitalsWeak",
	"Radar_LeftBed",
	"Stay",
	"NoActivity24h",
	"WarningArea",
	"PoorReception",
	"AngleException",
	"SittingOnGround",
}

// DeviceRepository 设备仓库
type DeviceRepository struct {
	db                   *sql.DB
	logger               *zap.Logger
	alarmEnablementCache map[string]AlarmEnablement
	cacheMutex           sync.RWMutex
}

// NewDeviceRepository 创建设备仓库
func NewDeviceRepository(db *sql.DB, logger *zap.Logger) *DeviceRepository {
	return &DeviceRepository{
		db:                   db,
		logger:               logger,
		alarmEnablementCache: make(map[string]AlarmEnablement),
	}
}

// Device 设备模型
type Device struct {
	DeviceID       string
	TenantID       string
	DeviceUID      string
	Status         string
	BusinessAccess string
	BoundBedID     *string
	BoundRoomID    *string
}

// AlarmEnablement 报警使能配置
type AlarmEnablement map[string]bool

// GetDeviceBySerialNumber 根据序列号获取设备（已废弃，使用GetDeviceByUID）
func (r *DeviceRepository) GetDeviceBySerialNumber(serialNumber string) (*Device, error) {
	return r.GetDeviceByUID(serialNumber)
}

// GetDeviceByUID 根据 device_uid 获取设备
func (r *DeviceRepository) GetDeviceByUID(uid string) (*Device, error) {
	query := `
		SELECT 
			d.device_id,
			d.tenant_id,
			d.device_uid,
			d.status,
			d.business_access,
			d.bound_bed_id,
			d.bound_room_id
		FROM devices d
		WHERE d.device_uid = $1
		LIMIT 1
	`

	device := &Device{}
	err := r.db.QueryRow(query, uid).Scan(
		&device.DeviceID,
		&device.TenantID,
		&device.DeviceUID,
		&device.Status,
		&device.BusinessAccess,
		&device.BoundBedID,
		&device.BoundRoomID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", uid)
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	return device, nil
}

// GetOrCreateDeviceFromStore 从 device_store 获取或创建设备
func (r *DeviceRepository) GetOrCreateDeviceFromStore(ctx context.Context, uid, topic string) (*Device, error) {
	// 1. 查询 device_store
	queryStore := `
		SELECT 
			ds.tenant_id,
			ds.device_type,
			ds.allow_access
		FROM device_store ds
		WHERE ds.device_uid = $1
		LIMIT 1
	`

	var tenantID sql.NullString
	var deviceType sql.NullString
	var allowAccess sql.NullBool

	err := r.db.QueryRowContext(ctx, queryStore, uid).Scan(&tenantID, &deviceType, &allowAccess)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found in device_store: %s", uid)
		}
		return nil, fmt.Errorf("failed to query device_store: %w", err)
	}

	if !tenantID.Valid || tenantID.String == "" {
		return nil, fmt.Errorf("device not allocated to tenant: %s", uid)
	}

	if !allowAccess.Bool {
		return nil, fmt.Errorf("device access not allowed: %s", uid)
	}

	// 2. 尝试从 devices 表查询
	device, err := r.GetDeviceByUID(uid)
	if err == nil {
		return device, nil
	}

	// 3. 设备不存在，创建新设备
	insertQuery := `
		INSERT INTO devices (
			tenant_id,
			device_uid,
			device_name,
			status,
			business_access,
			monitoring_enabled
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING device_id, tenant_id, device_uid, status, business_access, bound_bed_id, bound_room_id
	`

	deviceName := fmt.Sprintf("Radar-%s", uid)
	newDevice := &Device{}
	err = r.db.QueryRowContext(ctx, insertQuery,
		tenantID.String,
		uid,
		deviceName,
		"offline",
		"pending",
		false,
	).Scan(
		&newDevice.DeviceID,
		&newDevice.TenantID,
		&newDevice.DeviceUID,
		&newDevice.Status,
		&newDevice.BusinessAccess,
		&newDevice.BoundBedID,
		&newDevice.BoundRoomID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return newDevice, nil
}

// GetDeviceLocationInfoByIdentifier 通过设备标识符获取位置信息
func (r *DeviceRepository) GetDeviceLocationInfoByIdentifier(ctx context.Context, identifier string) (*DeviceLocationInfo, error) {
	if identifier == "" {
		return nil, fmt.Errorf("identifier is required")
	}

	query := `
		SELECT 
			d.device_id::text,
			d.tenant_id::text,
			d.device_uid,
			d.status,
			d.business_access,
			d.monitoring_enabled,
			
			u.branch_id,
			u.building_id,
			u.unit_id,
			r.room_id,
			b.bed_id,
			
			ds.device_type
		FROM devices d
		LEFT JOIN device_store ds ON d.device_uid = ds.device_uid
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON (
			(d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id) OR
			(b.room_id = r.room_id AND b.tenant_id = r.tenant_id)
		)
		LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE d.device_uid = $1
			AND d.status != 'disabled'
		LIMIT 1
	`

	var (
		deviceIDStr, tenantIDStr string
		deviceUID                sql.NullString
		status, businessAccess   sql.NullString
		monitoringEnabled        sql.NullBool
		branchID                 sql.NullString
		buildingID               sql.NullString
		unitID                   sql.NullString
		roomID                   sql.NullString
		bedID                    sql.NullString
		deviceType               sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, identifier).Scan(
		&deviceIDStr,
		&tenantIDStr,
		&deviceUID,
		&status,
		&businessAccess,
		&monitoringEnabled,
		&branchID,
		&buildingID,
		&unitID,
		&roomID,
		&bedID,
		&deviceType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: identifier=%s", identifier)
		}
		return nil, fmt.Errorf("failed to query device location: %w", err)
	}

	locationInfo := &DeviceLocationInfo{
		DeviceID: deviceIDStr,
		TenantID: tenantIDStr,
	}

	if deviceUID.Valid && deviceUID.String != "" {
		uidStr := deviceUID.String
		locationInfo.DeviceUID = &uidStr
	}
	if status.Valid && status.String != "" {
		statusStr := status.String
		locationInfo.Status = &statusStr
	}
	if businessAccess.Valid && businessAccess.String != "" {
		businessAccessStr := businessAccess.String
		locationInfo.BusinessAccess = &businessAccessStr
	}
	if monitoringEnabled.Valid {
		locationInfo.MonitoringEnabled = &monitoringEnabled.Bool
	}

	// 位置信息字段（按结构体顺序：BranchID → BuildingID → UnitID → RoomID → BedID）
	if branchID.Valid {
		branchIDStr := branchID.String
		locationInfo.BranchID = &branchIDStr
	}
	if buildingID.Valid {
		buildingIDStr := buildingID.String
		locationInfo.BuildingID = &buildingIDStr
	}
	if unitID.Valid {
		unitIDStr := unitID.String
		locationInfo.UnitID = &unitIDStr
	}
	if roomID.Valid {
		roomIDStr := roomID.String
		locationInfo.RoomID = &roomIDStr
	}
	if bedID.Valid {
		bedIDStr := bedID.String
		locationInfo.BedID = &bedIDStr
	}

	if deviceType.Valid && deviceType.String != "" {
		deviceTypeStr := deviceType.String
		locationInfo.DeviceType = &deviceTypeStr
	}

	return locationInfo, nil
}

// GetAlarmEnablement 获取设备的报警使能配置
func (r *DeviceRepository) GetAlarmEnablement(ctx context.Context, tenantID, deviceUID string) (AlarmEnablement, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceUID == "" {
		return nil, fmt.Errorf("device_uid is required")
	}

	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceUID)
	r.cacheMutex.RLock()
	if cached, ok := r.alarmEnablementCache[cacheKey]; ok {
		r.cacheMutex.RUnlock()
		return cached, nil
	}
	r.cacheMutex.RUnlock()

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
			// 没有配置，返回所有报警类型都未启用
			enablement := make(AlarmEnablement)
			for _, alarmType := range RadarAlarmTypes {
				enablement[alarmType] = false
			}
			return enablement, nil
		}
		return nil, fmt.Errorf("failed to query alarm enablement: %w", err)
	}

	enablement := make(AlarmEnablement)
	for _, alarmType := range RadarAlarmTypes {
		enablement[alarmType] = false
	}

	if !monitorConfigJSON.Valid || monitorConfigJSON.String == "" {
		r.cacheMutex.Lock()
		r.alarmEnablementCache[cacheKey] = enablement
		r.cacheMutex.Unlock()
		return enablement, nil
	}

	var monitorConfig map[string]interface{}
	if err := json.Unmarshal([]byte(monitorConfigJSON.String), &monitorConfig); err != nil {
		return nil, fmt.Errorf("failed to parse monitor_config: %w", err)
	}

	alarms, ok := monitorConfig["alarms"].(map[string]interface{})
	if !ok {
		r.cacheMutex.Lock()
		r.alarmEnablementCache[cacheKey] = enablement
		r.cacheMutex.Unlock()
		return enablement, nil
	}

	for _, alarmType := range RadarAlarmTypes {
		alarmConfig, ok := alarms[alarmType].(map[string]interface{})
		if !ok {
			enablement[alarmType] = false
			continue
		}

		if level, ok := alarmConfig["level"].(string); ok {
			levelUpper := strings.ToUpper(level)
			enablement[alarmType] = (levelUpper != "DISABLE" && levelUpper != "DISABLED")
			continue
		}

		if enabled, ok := alarmConfig["enabled"].(bool); ok {
			enablement[alarmType] = enabled
		}
	}

	r.cacheMutex.Lock()
	r.alarmEnablementCache[cacheKey] = enablement
	r.cacheMutex.Unlock()

	return enablement, nil
}

// ClearAlarmEnablementCache 清除报警使能缓存
func (r *DeviceRepository) ClearAlarmEnablementCache(tenantID, deviceUID string) {
	if tenantID == "" || deviceUID == "" {
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceUID)
	r.cacheMutex.Lock()
	delete(r.alarmEnablementCache, cacheKey)
	r.cacheMutex.Unlock()
}

// IsAlarmTypeEnabled 检查报警类型是否启用
func (r *DeviceRepository) IsAlarmTypeEnabled(ctx context.Context, tenantID, deviceUID, alarmType string) (bool, error) {
	enablement, err := r.GetAlarmEnablement(ctx, tenantID, deviceUID)
	if err != nil {
		return false, err
	}
	return enablement[alarmType], nil
}

// RefreshAlarmCacheForTenant 刷新租户的报警缓存
func (r *DeviceRepository) RefreshAlarmCacheForTenant(tenantID string) {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	prefix := tenantID + ":"
	for key := range r.alarmEnablementCache {
		if strings.HasPrefix(key, prefix) {
			delete(r.alarmEnablementCache, key)
		}
	}
}

// GetPossibleAlarmTypesFromEvent 从事件数据中获取可能的报警类型
func GetPossibleAlarmTypesFromEvent(dataValue interface{}) []string {
	var alarmTypes []string

	dataArray, ok := dataValue.([]interface{})
	if !ok {
		return alarmTypes
	}

	for _, item := range dataArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		category, _ := itemMap["category"].(string)

		switch category {
		case "enter2out":
			event, _ := itemMap["event"].(string)
			if event == "enter" || event == "exit" {
				alarmTypes = append(alarmTypes, "WarningArea")
			}
		case "pose":
			pose, _ := itemMap["pose"].(string)
			if pose == "Fall" {
				alarmTypes = append(alarmTypes, "Fall")
			} else if pose == "SuspectedFall" {
				alarmTypes = append(alarmTypes, "SuspectedFall")
			} else if pose == "SittingOnGround" {
				alarmTypes = append(alarmTypes, "SittingOnGround")
			}
		case "isOnline":
			deviceStatus, _ := itemMap["device_status"].(string)
			if deviceStatus == "offline" {
				alarmTypes = append(alarmTypes, "OfflineAlarm")
			}
		case "signal_poor":
			alarmTypes = append(alarmTypes, "PoorReception")
		case "angle_abnormal":
			alarmTypes = append(alarmTypes, "AngleException")
		case "other":
			alarmType, _ := itemMap["alarmType"].(string)
			switch alarmType {
			case "1":
				alarmTypes = append(alarmTypes, "Radar_LeftBed")
			case "2":
				alarmTypes = append(alarmTypes, "Stay")
			case "3":
				alarmTypes = append(alarmTypes, "NoActivity24h")
			}
		}
	}

	return alarmTypes
}

// GetPossibleAlarmTypesFromStat 从统计数据中获取可能的报警类型
func GetPossibleAlarmTypesFromStat(dataValue interface{}) []string {
	var alarmTypes []string

	var dataArray []interface{}
	switch v := dataValue.(type) {
	case []interface{}:
		dataArray = v
	case map[string]interface{}:
		dataArray = []interface{}{v}
	default:
		return alarmTypes
	}

	for _, item := range dataArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		category, _ := itemMap["category"].(string)

		if category == "sleep" {
			breathState, _ := itemMap["breath_state"].(string)
			if breathState == "Apnea" {
				alarmTypes = append(alarmTypes, "Radar_ApneaHypopnea")
			} else if strings.Contains(breathState, "Breath rate") {
				alarmTypes = append(alarmTypes, "Radar_AbnormalRespiratoryRate")
			}

			heartState, _ := itemMap["heart_state"].(string)
			if strings.Contains(heartState, "Heart rate") {
				alarmTypes = append(alarmTypes, "Radar_AbnormalHeartRate")
			}

			vitalSignsState, _ := itemMap["vital_signs_state"].(string)
			if vitalSignsState == "Vital signs weak" {
				alarmTypes = append(alarmTypes, "VitalsWeak")
			}
		}
	}

	return alarmTypes
}

// GetDevicesByBranchID 根据 branch_id 获取设备ID列表
func (r *DeviceRepository) GetDevicesByBranchID(ctx context.Context, tenantID, branchID string) ([]string, error) {
	query := `
		SELECT DISTINCT d.device_id::text
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON (
			(d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id) OR
			(b.room_id = r.room_id AND b.tenant_id = r.tenant_id)
		)
		LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE d.tenant_id = $1
			AND u.branch_id = $2
			AND d.status != 'disabled'
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices by branch_id: %w", err)
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("failed to scan device_id: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs, rows.Err()
}

// GetDevicesByBuildingID 根据 building_id 获取设备ID列表
func (r *DeviceRepository) GetDevicesByBuildingID(ctx context.Context, tenantID, buildingID string) ([]string, error) {
	query := `
		SELECT DISTINCT d.device_id::text
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON (
			(d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id) OR
			(b.room_id = r.room_id AND b.tenant_id = r.tenant_id)
		)
		LEFT JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE d.tenant_id = $1
			AND u.building_id = $2
			AND d.status != 'disabled'
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices by building_id: %w", err)
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("failed to scan device_id: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs, rows.Err()
}

// GetDevicesByUnitID 根据 unit_id 获取设备ID列表
func (r *DeviceRepository) GetDevicesByUnitID(ctx context.Context, tenantID, unitID string) ([]string, error) {
	query := `
		SELECT DISTINCT d.device_id::text
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		LEFT JOIN rooms r ON (
			(d.bound_room_id = r.room_id AND d.tenant_id = r.tenant_id) OR
			(b.room_id = r.room_id AND b.tenant_id = r.tenant_id)
		)
		WHERE d.tenant_id = $1
			AND r.unit_id = $2
			AND d.status != 'disabled'
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices by unit_id: %w", err)
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("failed to scan device_id: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs, rows.Err()
}

// GetUnitIDForRoomOrBed 根据 room_id 或 bed_id 获取 unit_id
func (r *DeviceRepository) GetUnitIDForRoomOrBed(ctx context.Context, tenantID, addressID, addressType string) (string, error) {
	var query string
	if addressType == "room" {
		query = `
			SELECT r.unit_id::text
			FROM rooms r
			WHERE r.tenant_id = $1 AND r.room_id = $2
			LIMIT 1
		`
	} else if addressType == "bed" {
		query = `
			SELECT r.unit_id::text
			FROM beds b
			INNER JOIN rooms r ON b.room_id = r.room_id AND b.tenant_id = r.tenant_id
			WHERE b.tenant_id = $1 AND b.bed_id = $2
			LIMIT 1
		`
	} else {
		return "", fmt.Errorf("invalid address_type: %s", addressType)
	}

	var unitID sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, addressID).Scan(&unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("unit not found for %s: %s", addressType, addressID)
		}
		return "", fmt.Errorf("failed to query unit_id: %w", err)
	}

	if !unitID.Valid {
		return "", fmt.Errorf("unit_id is null for %s: %s", addressType, addressID)
	}

	return unitID.String, nil
}
