package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
)

// DeviceRepository 设备仓库
type DeviceRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDeviceRepository 创建设备仓库
func NewDeviceRepository(db *sql.DB, logger *zap.Logger) *DeviceRepository {
	return &DeviceRepository{
		db:     db,
		logger: logger,
	}
}

// GetDeviceByCode 根据设备代码获取设备（Sleepace 使用 device_code）
func (r *DeviceRepository) GetDeviceByCode(deviceCode string) (*Device, error) {
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
		WHERE d.serial_number = $1 OR d.uid = $1
		LIMIT 1
	`
	
	device := &Device{}
	err := r.db.QueryRow(query, deviceCode).Scan(
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
			return nil, fmt.Errorf("device not found: %s", deviceCode)
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
			device_model,
			serial_number,
			uid,
			tenant_id::text,
			allow_access
		FROM device_store
		WHERE (serial_number = $1 OR uid = $1)
		LIMIT 1
	`

	var dsDeviceStoreID, dsDeviceType, dsTenantID string
	var dsDeviceModel, dsSerialNumber, dsUID sql.NullString
	var dsAllowAccess bool

	err = r.db.QueryRowContext(ctx, deviceStoreQuery, identifier).Scan(
		&dsDeviceStoreID,
		&dsDeviceType,
		&dsDeviceModel,
		&dsSerialNumber,
		&dsUID,
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
		device, err := r.GetDeviceByCode(serialNum)
		if err == nil {
			return device, nil
		}
	}
	if uid != "" {
		device, err := r.GetDeviceByCode(uid)
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
	device = &Device{}
	err = r.db.QueryRowContext(ctx, query, newDeviceID).Scan(
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
		return nil, fmt.Errorf("failed to query newly created device: %w", err)
	}
	return device, nil
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

