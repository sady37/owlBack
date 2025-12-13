package repository

import (
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

