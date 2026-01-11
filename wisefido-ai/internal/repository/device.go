package repository

import (
	"context"
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

// GetDeviceBindingInfo 获取设备的绑定信息（需验证 tenant_id）
func (r *DeviceRepository) GetDeviceBindingInfo(ctx context.Context, tenantID, deviceID string) (*DeviceBindingInfo, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	query := `
		SELECT 
			d.device_id,
			ds.device_type,
			d.bound_bed_id,
			d.bound_room_id,
			COALESCE(
				(SELECT u.unit_id FROM units u WHERE u.unit_id = (
					SELECT r.unit_id FROM rooms r WHERE r.room_id = d.bound_room_id AND r.tenant_id = d.tenant_id LIMIT 1
				) AND u.tenant_id = d.tenant_id LIMIT 1),
				(SELECT u.unit_id FROM units u WHERE u.unit_id = (
					SELECT r.unit_id FROM rooms r JOIN beds b ON r.room_id = b.room_id WHERE b.bed_id = d.bound_bed_id AND r.tenant_id = d.tenant_id LIMIT 1
				) AND u.tenant_id = d.tenant_id LIMIT 1),
				NULL
			) as unit_id,
			COALESCE(
				d.bound_room_id,
				(SELECT r.room_id FROM rooms r JOIN beds b ON r.room_id = b.room_id WHERE b.bed_id = d.bound_bed_id AND r.tenant_id = d.tenant_id LIMIT 1),
				NULL
			) as room_id
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE d.device_id = $1 AND d.tenant_id = $2
	`
	
	var info DeviceBindingInfo
	var unitID, roomID sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, deviceID, tenantID).Scan(
		&info.DeviceID,
		&info.DeviceType,
		&info.BoundBedID,
		&info.BoundRoomID,
		&unitID,
		&roomID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: device_id=%s, tenant_id=%s", deviceID, tenantID)
		}
		return nil, fmt.Errorf("failed to query device binding: %w", err)
	}
	
	if unitID.Valid {
		info.UnitID = unitID.String
	} else {
		return nil, fmt.Errorf("device unit_id not found: device_id=%s, tenant_id=%s", deviceID, tenantID)
	}
	
	if roomID.Valid {
		info.RoomID = &roomID.String
	}
	
	return &info, nil
}

// GetDevicesByRoom 获取房间内的所有设备（需验证 tenant_id）
func (r *DeviceRepository) GetDevicesByRoom(ctx context.Context, tenantID, roomID string) ([]DeviceBindingInfo, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if roomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	query := `
		SELECT 
			d.device_id,
			ds.device_type,
			d.bound_bed_id,
			d.bound_room_id,
			r.unit_id,
			$2 as room_id
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		JOIN rooms r ON (
			d.bound_room_id = r.room_id
			OR d.bound_bed_id IN (
				SELECT bed_id FROM beds WHERE room_id = r.room_id AND tenant_id = $1
			)
		) AND r.tenant_id = $1
		WHERE d.tenant_id = $1
		  AND r.room_id = $2
		  AND d.monitoring_enabled = TRUE
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, roomID)
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
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate devices: %w", err)
	}
	
	return devices, nil
}

// GetDevicesByBed 获取床上的所有设备（需验证 tenant_id）
func (r *DeviceRepository) GetDevicesByBed(ctx context.Context, tenantID, bedID string) ([]DeviceBindingInfo, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if bedID == "" {
		return nil, fmt.Errorf("bed_id is required")
	}

	query := `
		SELECT 
			d.device_id,
			ds.device_type,
			d.bound_bed_id,
			d.bound_room_id,
			r.unit_id,
			r.room_id
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		JOIN beds b ON d.bound_bed_id = b.bed_id AND d.tenant_id = b.tenant_id
		JOIN rooms r ON b.room_id = r.room_id AND b.tenant_id = r.tenant_id
		WHERE d.tenant_id = $1
		  AND d.bound_bed_id = $2
		  AND d.monitoring_enabled = TRUE
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, bedID)
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
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate devices: %w", err)
	}
	
	return devices, nil
}

