package repository

import (
	"database/sql"
	"fmt"
	"wisefido-data-transformer/internal/models"
	
	"go.uber.org/zap"
)

// IoTTimeSeriesRepository IoT 时序数据仓库
type IoTTimeSeriesRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewIoTTimeSeriesRepository 创建 IoT 时序数据仓库
func NewIoTTimeSeriesRepository(db *sql.DB, logger *zap.Logger) *IoTTimeSeriesRepository {
	return &IoTTimeSeriesRepository{
		db:     db,
		logger: logger,
	}
}

// Insert 插入标准化数据到 iot_timeseries 表
func (r *IoTTimeSeriesRepository) Insert(data *models.StandardizedData) (int64, error) {
	query := `
		INSERT INTO iot_timeseries (
			tenant_id,
			device_id,
			timestamp,
			data_type,
			category,
			tracking_id,
			radar_pos_x,
			radar_pos_y,
			radar_pos_z,
			posture_snomed_code,
			posture_display,
			event_type,
			event_snomed_code,
			event_display,
			area_id,
			heart_rate_code,
			heart_rate_display,
			heart_rate,
			respiratory_rate_code,
			respiratory_rate_display,
			respiratory_rate,
			sleep_state_snomed_code,
			sleep_state_display,
			bed_status_snomed_code,
			bed_status_display,
			raw_original,
			unit_id,
			room_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
		)
		RETURNING id
	`
	
	var id int64
	err := r.db.QueryRow(
		query,
		data.TenantID,
		data.DeviceID,
		data.Timestamp,
		data.DataType,
		data.Category,
		data.TrackingID,
		data.RadarPosX,
		data.RadarPosY,
		data.RadarPosZ,
		data.PostureSNOMEDCode,
		data.PostureDisplay,
		data.EventType,
		data.EventSNOMEDCode,
		data.EventDisplay,
		data.AreaID,
		data.HeartRateCode,
		data.HeartRateDisplay,
		data.HeartRate,
		data.RespiratoryRateCode,
		data.RespiratoryRateDisplay,
		data.RespiratoryRate,
		data.SleepStateSNOMEDCode,
		data.SleepStateDisplay,
		data.BedStatusSNOMEDCode,
		data.BedStatusDisplay,
		data.RawOriginal,
		nil, // unit_id - 从 device 表 JOIN 获取
		nil, // room_id - 从 device 表 JOIN 获取
	).Scan(&id)
	
	if err != nil {
		return 0, fmt.Errorf("failed to insert iot_timeseries: %w", err)
	}
	
	return id, nil
}

// GetDeviceLocation 获取设备位置信息（unit_id, room_id）
func (r *IoTTimeSeriesRepository) GetDeviceLocation(deviceID string) (unitID *string, roomID *string, err error) {
	query := `
		SELECT 
			u.unit_id,
			r.room_id
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
		LEFT JOIN rooms r ON COALESCE(b.room_id, d.bound_room_id) = r.room_id
		LEFT JOIN units u ON r.unit_id = u.unit_id
		WHERE d.device_id = $1
		LIMIT 1
	`
	
	var uID, rID sql.NullString
	err = r.db.QueryRow(query, deviceID).Scan(&uID, &rID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("device not found: %s", deviceID)
		}
		return nil, nil, fmt.Errorf("failed to query device location: %w", err)
	}
	
	if uID.Valid {
		unitID = &uID.String
	}
	if rID.Valid {
		roomID = &rID.String
	}
	
	return unitID, roomID, nil
}

// UpdateLocation 更新记录的 unit_id 和 room_id
func (r *IoTTimeSeriesRepository) UpdateLocation(id int64, unitID *string, roomID *string) error {
	query := `
		UPDATE iot_timeseries
		SET unit_id = $1, room_id = $2
		WHERE id = $3
	`
	
	_, err := r.db.Exec(query, unitID, roomID, id)
	if err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}
	
	return nil
}

