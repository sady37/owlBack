package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// RadarAlarmTypes Radar 设备的所有报警类型列表
// 这些是可能出现在 alarm_device.monitor_config.alarms 中的报警类型
// 注意：Radar_AbnormalHeartRate 和 Radar_AbnormalRespiratoryRate 是从 stat 数据中判断的
var RadarAlarmTypes = []string{
	// 通用报警（所有设备类型都支持）
	"OfflineAlarm",
	"LowBattery",
	"DeviceFailure",
	// Radar 特定报警
	"Radar_ApneaHypopnea",           // 从 stat 数据中判断（breath_state="Apnea"）
	"Radar_AbnormalHeartRate",       // 从 stat 数据中判断（heart_state="Heart rate low/high"）
	"Radar_AbnormalRespiratoryRate", // 从 stat 数据中判断（breath_state="Breath rate low/high"）
	"SuspectedFall",                 // 从 event type=2 (pose) 中判断
	"Fall",                          // 从 event type=2 (pose) 中判断
	"VitalsWeak",                    // 从 stat 数据中判断（vital_signs_state="Vital signs weak"）
	"Radar_LeftBed",                 // 从 event type=9 (alarmType="1") 中判断
	"Stay",                          // 从 event type=9 (alarmType="2") 中判断
	"NoActivity24h",                 // 从 event type=9 (alarmType="3") 中判断
	"WarningArea",                   // 从 event type=1 (enter2out) 中判断
	"PoorReception",                 // 从 event type=7 (signal_poor) 中判断
	"AngleException",                // 从 event type=8 (angle_abnormal) 中判断
	"SittingOnGround",               // 从 event type=2 (pose) 中判断，虽然不在 AlarmCloud.vue 的默认列表中，但在 device_monitor_settings_service.go 中存在
}

// DeviceRepository 设备仓库
type DeviceRepository struct {
	db     *sql.DB
	logger *zap.Logger
	// 报警使能配置缓存：key = tenantID:deviceID, value = AlarmEnablement
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

// GetDeviceBySerialNumber 根据序列号获取设备
func (r *DeviceRepository) GetDeviceBySerialNumber(serialNumber string) (*Device, error) {
	query := `
		SELECT 
			d.device_id,
			d.tenant_id,
			d.serial_number,
			d.uid,
			d.device_name,
			d.status,
			d.business_access,
			d.bound_bed_id,
			d.bound_room_id
		FROM devices d
		WHERE d.serial_number = $1
		LIMIT 1
	`

	device := &Device{}
	err := r.db.QueryRow(query, serialNumber).Scan(
		&device.DeviceID,
		&device.TenantID,
		&device.SerialNumber,
		&device.UID,
		&device.DeviceName,
		&device.Status,
		&device.BusinessAccess,
		&device.BoundBedID,
		&device.BoundRoomID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", serialNumber)
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	return device, nil
}

// GetDeviceByUID 根据 UID 获取设备
func (r *DeviceRepository) GetDeviceByUID(uid string) (*Device, error) {
	query := `
		SELECT 
			d.device_id,
			d.tenant_id,
			d.serial_number,
			d.uid,
			d.device_name,
			d.status,
			d.business_access,
			d.bound_bed_id,
			d.bound_room_id
		FROM devices d
		WHERE d.uid = $1
		LIMIT 1
	`

	device := &Device{}
	err := r.db.QueryRow(query, uid).Scan(
		&device.DeviceID,
		&device.TenantID,
		&device.SerialNumber,
		&device.UID,
		&device.DeviceName,
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

// GetOrCreateDeviceFromStore attempts to get device from devices table, and if not found,
// checks device_store table. If device_store exists and is allocated, creates devices record.
// This is used for automatic device creation on first MQTT connection.
// Returns the device and error. Errors are logged for security auditing.
func (r *DeviceRepository) GetOrCreateDeviceFromStore(ctx context.Context, identifier string, mqttTopic string) (*Device, error) {
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
			d.device_id,
			d.tenant_id,
			d.serial_number,
			d.uid,
			d.device_name,
			d.status,
			d.business_access,
			d.bound_bed_id,
			d.bound_room_id
		FROM devices d
		WHERE (d.serial_number = $1 OR d.uid = $1)
		LIMIT 1
	`

	device := &Device{}
	err := r.db.QueryRowContext(ctx, deviceQuery, identifier).Scan(
		&device.DeviceID,
		&device.TenantID,
		&device.SerialNumber,
		&device.UID,
		&device.DeviceName,
		&device.Status,
		&device.BusinessAccess,
		&device.BoundBedID,
		&device.BoundRoomID,
	)

	if err == nil {
		// Device found in devices table, return it
		return device, nil
	}

	if err != sql.ErrNoRows {
		// Unexpected database error
		logWarn("Device connection failed: database error",
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("reason", "database_error"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	// 2. Device not found in devices table, check device_store table
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
		// Case 3: Device not registered in device_store
		var serialNum, uid string
		if dsSerialNumber.Valid {
			serialNum = dsSerialNumber.String
		}
		if dsUID.Valid {
			uid = dsUID.String
		}
		logWarn("Unauthorized device connection attempt",
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("serial_number", serialNum),
			zap.String("uid", uid),
			zap.String("reason", "device_not_registered"),
			zap.String("action", "connection_rejected"),
			zap.String("security_level", "warning"),
		)
		return nil, fmt.Errorf("unauthorized device: not registered in device_store")
	}

	if err != nil {
		// Unexpected database error
		logWarn("Device connection failed: database error",
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("reason", "database_error"),
			zap.Error(err),
		)
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
		logWarn("Device connection failed: transaction error",
			zap.String("device_store_id", dsDeviceStoreID),
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("reason", "transaction_error"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Generate device name from device_type and serial_number/uid
	deviceName := dsDeviceType
	serialNum := ""
	uid := ""
	if dsSerialNumber.Valid && dsSerialNumber.String != "" {
		serialNum = dsSerialNumber.String
		deviceName = dsDeviceType + "-" + serialNum
	} else if dsUID.Valid && dsUID.String != "" {
		uid = dsUID.String
		deviceName = dsDeviceType + "-" + uid
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
		RETURNING device_id
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
		logWarn("Device connection failed: failed to create device record",
			zap.String("device_store_id", dsDeviceStoreID),
			zap.String("tenant_id", dsTenantID),
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("reason", "device_creation_failed"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create device record: %w", err)
	}

	if err = tx.Commit(); err != nil {
		logWarn("Device connection failed: transaction commit error",
			zap.String("device_store_id", dsDeviceStoreID),
			zap.String("device_id", newDeviceID),
			zap.String("tenant_id", dsTenantID),
			zap.String("identifier", identifier),
			zap.String("mqtt_topic", mqttTopic),
			zap.String("reason", "transaction_commit_failed"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log successful auto-creation
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
	// Try serial_number first, then uid
	if serialNum != "" {
		device, err := r.GetDeviceBySerialNumber(serialNum)
		if err == nil {
			return device, nil
		}
	}
	if uid != "" {
		device, err := r.GetDeviceByUID(uid)
		if err == nil {
			return device, nil
		}
	}
	// Fallback: query by device_id
	query := `
		SELECT 
			d.device_id,
			d.tenant_id,
			d.serial_number,
			d.uid,
			d.device_name,
			d.status,
			d.business_access,
			d.bound_bed_id,
			d.bound_room_id
		FROM devices d
		WHERE d.device_id = $1
		LIMIT 1
	`
	var fallbackDevice Device
	err = r.db.QueryRowContext(ctx, query, newDeviceID).Scan(
		&fallbackDevice.DeviceID,
		&fallbackDevice.TenantID,
		&fallbackDevice.SerialNumber,
		&fallbackDevice.UID,
		&fallbackDevice.DeviceName,
		&fallbackDevice.Status,
		&fallbackDevice.BusinessAccess,
		&fallbackDevice.BoundBedID,
		&fallbackDevice.BoundRoomID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query newly created device: %w", err)
	}
	return &fallbackDevice, nil
}

// Device 设备模型
type Device struct {
	DeviceID       string
	TenantID       string
	SerialNumber   string
	UID            string
	DeviceName     string
	Status         string
	BusinessAccess string
	BoundBedID     *string
	BoundRoomID    *string
}

// AlarmEnablement 报警使能配置
// key: 报警类型（如 "Fall", "SuspectedFall", "OfflineAlarm" 等）
// value: 是否启用（true=启用，false=未启用或未配置）
type AlarmEnablement map[string]bool

// GetAlarmEnablement 获取设备的报警使能配置
// 从 alarm_device.monitor_config.alarms 中读取所有报警类型的 enabled 状态
// 如果设备没有配置或某个报警类型未配置，返回 false
// 使用内存缓存提高性能，配置变更时通过 ClearAlarmEnablementCache 清除缓存
func (r *DeviceRepository) GetAlarmEnablement(ctx context.Context, tenantID, deviceID string) (AlarmEnablement, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceID)
	r.cacheMutex.RLock()
	if cached, ok := r.alarmEnablementCache[cacheKey]; ok {
		r.cacheMutex.RUnlock()
		return cached, nil
	}
	r.cacheMutex.RUnlock()

	// 缓存未命中，从数据库查询
	// 查询 alarm_device 表的 monitor_config 字段
	query := `
		SELECT monitor_config
		FROM alarm_device
		WHERE device_id = $1 AND tenant_id = $2
	`

	var monitorConfigJSON json.RawMessage
	err := r.db.QueryRowContext(ctx, query, deviceID, tenantID).Scan(&monitorConfigJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// 设备没有配置，返回空的使能配置（所有报警类型都未启用）
			return make(AlarmEnablement), nil
		}
		return nil, fmt.Errorf("failed to query alarm_device: %w", err)
	}

	// 解析 monitor_config JSON
	var monitorConfig map[string]interface{}
	if err := json.Unmarshal(monitorConfigJSON, &monitorConfig); err != nil {
		if r.logger != nil {
			r.logger.Warn("Failed to parse monitor_config",
				zap.String("device_id", deviceID),
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
		}
		// 解析失败，返回空的使能配置
		return make(AlarmEnablement), nil
	}

	// 提取 alarms 对象
	alarmsObj, ok := monitorConfig["alarms"]
	if !ok {
		// 没有 alarms 字段，返回空的使能配置
		return make(AlarmEnablement), nil
	}

	// 将 alarms 转换为 map[string]interface{}
	alarmsMap, ok := alarmsObj.(map[string]interface{})
	if !ok {
		if r.logger != nil {
			r.logger.Warn("Invalid alarms format in monitor_config",
				zap.String("device_id", deviceID),
				zap.String("tenant_id", tenantID),
			)
		}
		return make(AlarmEnablement), nil
	}

	// 构建使能配置 map
	enablement := make(AlarmEnablement)
	for alarmType, alarmConfig := range alarmsMap {
		// alarmConfig 应该是 map[string]interface{}，包含 "enabled" 字段
		alarmConfigMap, ok := alarmConfig.(map[string]interface{})
		if !ok {
			// 格式不正确，跳过
			continue
		}

		// 检查 enabled 字段
		enabled, ok := alarmConfigMap["enabled"]
		if !ok {
			// 没有 enabled 字段，默认为 false
			enablement[alarmType] = false
			continue
		}

		// 将 enabled 转换为 bool
		enabledBool, ok := enabled.(bool)
		if !ok {
			// 类型转换失败，默认为 false
			enablement[alarmType] = false
			continue
		}

		enablement[alarmType] = enabledBool
	}

	// 写入缓存
	r.cacheMutex.Lock()
	r.alarmEnablementCache[cacheKey] = enablement
	r.cacheMutex.Unlock()

	return enablement, nil
}

// ClearAlarmEnablementCache 清除指定设备的报警使能配置缓存
// 当收到配置变更事件时调用此方法
func (r *DeviceRepository) ClearAlarmEnablementCache(tenantID, deviceID string) {
	if tenantID == "" || deviceID == "" {
		return
	}

	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceID)
	r.cacheMutex.Lock()
	delete(r.alarmEnablementCache, cacheKey)
	r.cacheMutex.Unlock()

	if r.logger != nil {
		r.logger.Debug("Cleared alarm enablement cache",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
		)
	}
}

// IsAlarmTypeEnabled 检查特定报警类型是否启用
// 如果设备没有配置或报警类型未配置，返回 false
func (r *DeviceRepository) IsAlarmTypeEnabled(ctx context.Context, tenantID, deviceID, alarmType string) (bool, error) {
	enablement, err := r.GetAlarmEnablement(ctx, tenantID, deviceID)
	if err != nil {
		return false, err
	}

	// 检查报警类型是否在使能配置中，且为 true
	return enablement[alarmType], nil
}

// GetPossibleAlarmTypesFromEvent 根据 event 数据确定可能触发的报警类型列表
// 注意：这只是根据 event type 和 category 判断可能触发的报警类型，不检查具体值
// 具体值检查由后端服务（wisefido-AI 或 wisefido-card-aggregator）处理
// 参考：wisefido-card-aggregator/internal/alarm/alarm_handler.go 中的 IsDeviceDirectAlarm 函数
func GetPossibleAlarmTypesFromEvent(dataValue interface{}) []string {
	possibleAlarms := []string{}

	// 处理 data_value 可能是数组或对象的情况
	var eventArray []interface{}
	switch v := dataValue.(type) {
	case []interface{}:
		eventArray = v
	case map[string]interface{}:
		eventArray = []interface{}{v}
	default:
		return possibleAlarms
	}

	for _, eventItem := range eventArray {
		eventMap, ok := eventItem.(map[string]interface{})
		if !ok {
			continue
		}

		category, ok := eventMap["category"].(string)
		if !ok {
			continue
		}

		switch category {
		case "pose":
			// type=2: 姿态变化事件
			// 可能触发：Fall, SuspectedFall, SittingOnGround
			possibleAlarms = append(possibleAlarms, "Fall", "SuspectedFall", "SittingOnGround")

		case "isOnline":
			// type=5: 设备在线状态
			// 可能触发：OfflineAlarm
			possibleAlarms = append(possibleAlarms, "OfflineAlarm")

		case "signal_poor":
			// type=7: 信号差事件
			// 可能触发：PoorReception
			possibleAlarms = append(possibleAlarms, "PoorReception")

		case "angle_abnormal":
			// type=8: 倾角异常事件
			// 可能触发：AngleException
			possibleAlarms = append(possibleAlarms, "AngleException")

		case "other":
			// type=9: 其他告警事件
			// 根据 alarmType 字段判断
			if alarmType, ok := eventMap["alarmType"].(string); ok {
				switch alarmType {
				case "1":
					// 离床未归
					possibleAlarms = append(possibleAlarms, "Radar_LeftBed")
				case "2":
					// 滞留
					possibleAlarms = append(possibleAlarms, "Stay")
				case "3":
					// 长时间无人活动
					possibleAlarms = append(possibleAlarms, "NoActivity24h")
				}
			}

		case "enter2out":
			// type=1: 进出事件
			// 可能触发：WarningArea（进入警告区域）
			possibleAlarms = append(possibleAlarms, "WarningArea")
		}
	}

	return possibleAlarms
}

// GetPossibleAlarmTypesFromStat 根据 stat 数据确定可能触发的报警类型列表
// 注意：这只是根据 category 和状态字段判断可能触发的报警类型，不检查具体值
// 具体值检查由后端服务（wisefido-AI 或 wisefido-card-aggregator）处理
func GetPossibleAlarmTypesFromStat(dataValue interface{}) []string {
	possibleAlarms := []string{}

	// 处理 data_value 可能是数组或对象的情况
	var statArray []interface{}
	switch v := dataValue.(type) {
	case []interface{}:
		statArray = v
	case map[string]interface{}:
		statArray = []interface{}{v}
	default:
		return possibleAlarms
	}

	for _, statItem := range statArray {
		statMap, ok := statItem.(map[string]interface{})
		if !ok {
			continue
		}

		category, ok := statMap["category"].(string)
		if !ok {
			continue
		}

		if category == "sleep" {
			// sleep category: 可能触发生命体征相关报警
			// 检查 heart_state
			if heartState, ok := statMap["heart_state"].(string); ok {
				if heartState == "Heart rate low" || heartState == "Heart rate high" {
					possibleAlarms = append(possibleAlarms, "Radar_AbnormalHeartRate")
				}
			}

			// 检查 breath_state
			if breathState, ok := statMap["breath_state"].(string); ok {
				if breathState == "Breath rate low" || breathState == "Breath rate high" {
					possibleAlarms = append(possibleAlarms, "Radar_AbnormalRespiratoryRate")
				} else if breathState == "Apnea" {
					possibleAlarms = append(possibleAlarms, "Radar_ApneaHypopnea")
				}
			}

			// 检查 vital_signs_state
			if vitalState, ok := statMap["vital_signs_state"].(string); ok {
				if vitalState == "Vital signs weak" {
					possibleAlarms = append(possibleAlarms, "VitalsWeak")
				}
			}
		}
	}

	return possibleAlarms
}
