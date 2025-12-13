package repository

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// DeviceRepository 设备仓库（用于报警评估）
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

// DeviceBindingInfo 设备绑定信息
type DeviceBindingInfo struct {
	DeviceID     string
	DeviceType   string
	BoundBedID   *string
	BoundRoomID  *string
	UnitID       string
	RoomID       *string // 通过 bound_room_id 或 bound_bed_id 查询得到
}

// GetDeviceBindingInfo 获取设备的绑定信息
func (r *DeviceRepository) GetDeviceBindingInfo(tenantID, deviceID string) (*DeviceBindingInfo, error) {
	query := `
		SELECT 
			d.device_id,
			ds.device_type,
			d.bound_bed_id,
			d.bound_room_id,
			d.unit_id,
			COALESCE(
				d.bound_room_id,
				(SELECT r.room_id FROM rooms r WHERE r.bed_id = d.bound_bed_id AND r.tenant_id = d.tenant_id LIMIT 1),
				NULL
			) as room_id
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE d.device_id = $1 AND d.tenant_id = $2
	`
	
	var info DeviceBindingInfo
	var roomID sql.NullString
	
	err := r.db.QueryRow(query, deviceID, tenantID).Scan(
		&info.DeviceID,
		&info.DeviceType,
		&info.BoundBedID,
		&info.BoundRoomID,
		&info.UnitID,
		&roomID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", deviceID)
		}
		return nil, fmt.Errorf("failed to query device binding: %w", err)
	}
	
	if roomID.Valid {
		info.RoomID = &roomID.String
	}
	
	return &info, nil
}

// GetDevicesByRoom 获取房间内的所有设备
func (r *DeviceRepository) GetDevicesByRoom(tenantID, roomID string) ([]DeviceBindingInfo, error) {
	query := `
		SELECT 
			d.device_id,
			ds.device_type,
			d.bound_bed_id,
			d.bound_room_id,
			d.unit_id,
			$2 as room_id
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE d.tenant_id = $1
		  AND (
			d.bound_room_id = $2
			OR d.bound_bed_id IN (
				SELECT bed_id FROM beds WHERE room_id = $2 AND tenant_id = $1
			)
		  )
		  AND d.monitoring_enabled = TRUE
	`
	
	rows, err := r.db.Query(query, tenantID, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices by room: %w", err)
	}
	defer rows.Close()
	
	var devices []DeviceBindingInfo
	for rows.Next() {
		var device DeviceBindingInfo
		var roomID sql.NullString
		
		if err := rows.Scan(
			&device.DeviceID,
			&device.DeviceType,
			&device.BoundBedID,
			&device.BoundRoomID,
			&device.UnitID,
			&roomID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		
		if roomID.Valid {
			device.RoomID = &roomID.String
		}
		
		devices = append(devices, device)
	}
	
	return devices, nil
}

// GetDevicesByBed 获取床上的所有设备
func (r *DeviceRepository) GetDevicesByBed(tenantID, bedID string) ([]DeviceBindingInfo, error) {
	query := `
		SELECT 
			d.device_id,
			ds.device_type,
			d.bound_bed_id,
			d.bound_room_id,
			d.unit_id,
			(SELECT r.room_id FROM rooms r WHERE r.bed_id = $2 AND r.tenant_id = $1 LIMIT 1) as room_id
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE d.tenant_id = $1
		  AND d.bound_bed_id = $2
		  AND d.monitoring_enabled = TRUE
	`
	
	rows, err := r.db.Query(query, tenantID, bedID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices by bed: %w", err)
	}
	defer rows.Close()
	
	var devices []DeviceBindingInfo
	for rows.Next() {
		var device DeviceBindingInfo
		var roomID sql.NullString
		
		if err := rows.Scan(
			&device.DeviceID,
			&device.DeviceType,
			&device.BoundBedID,
			&device.BoundRoomID,
			&device.UnitID,
			&roomID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		
		if roomID.Valid {
			device.RoomID = &roomID.String
		}
		
		devices = append(devices, device)
	}
	
	return devices, nil
}

