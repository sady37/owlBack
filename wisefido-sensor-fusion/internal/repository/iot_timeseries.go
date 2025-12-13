package repository

import (
	"database/sql"
	"fmt"
	"wisefido-sensor-fusion/internal/models"
	
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

// GetLatestByDeviceID 获取设备最新的时序数据
// 
// 参数:
//   - tenantID: 租户 ID（用于数据隔离）
//   - deviceID: 设备 ID
//   - limit: 返回记录数限制
func (r *IoTTimeSeriesRepository) GetLatestByDeviceID(tenantID, deviceID string, limit int) ([]*models.IoTTimeSeries, error) {
	query := `
		SELECT 
			its.id,
			its.tenant_id,
			its.device_id,
			its.timestamp,
			its.heart_rate,
			its.heart_rate_code,
			its.heart_rate_display,
			its.respiratory_rate,
			its.respiratory_rate_code,
			its.respiratory_rate_display,
			its.posture_snomed_code,
			its.posture_display,
			its.tracking_id,
			its.bed_status_snomed_code,
			its.bed_status_display,
			its.sleep_state_snomed_code,
			its.sleep_state_display,
			COALESCE(ds.device_type, '') as device_type
		FROM iot_timeseries its
		LEFT JOIN devices d ON its.device_id = d.device_id
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE its.device_id = $1 AND its.tenant_id = $2
		ORDER BY its.timestamp DESC
		LIMIT $3
	`
	
	rows, err := r.db.Query(query, deviceID, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query iot_timeseries: %w", err)
	}
	defer rows.Close()
	
	var results []*models.IoTTimeSeries
	for rows.Next() {
		item := &models.IoTTimeSeries{}
		var heartRate, respiratoryRate sql.NullInt64
		var heartRateCode, heartRateDisplay sql.NullString
		var respiratoryRateCode, respiratoryRateDisplay sql.NullString
		var postureCode, postureDisplay sql.NullString
		var trackingID sql.NullString
		var bedStatusCode, bedStatusDisplay sql.NullString
		var sleepStateCode, sleepStateDisplay sql.NullString
		var deviceType sql.NullString
		
		err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.DeviceID,
			&item.Timestamp,
			&heartRate,
			&heartRateCode,
			&heartRateDisplay,
			&respiratoryRate,
			&respiratoryRateCode,
			&respiratoryRateDisplay,
			&postureCode,
			&postureDisplay,
			&trackingID,
			&bedStatusCode,
			&bedStatusDisplay,
			&sleepStateCode,
			&sleepStateDisplay,
			&deviceType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		if heartRate.Valid {
			hr := int(heartRate.Int64)
			item.HeartRate = &hr
		}
		if heartRateCode.Valid {
			item.HeartRateCode = &heartRateCode.String
		}
		if heartRateDisplay.Valid {
			item.HeartRateDisplay = &heartRateDisplay.String
		}
		
		if respiratoryRate.Valid {
			rr := int(respiratoryRate.Int64)
			item.RespiratoryRate = &rr
		}
		if respiratoryRateCode.Valid {
			item.RespiratoryRateCode = &respiratoryRateCode.String
		}
		if respiratoryRateDisplay.Valid {
			item.RespiratoryRateDisplay = &respiratoryRateDisplay.String
		}
		
		if postureCode.Valid {
			item.PostureSNOMEDCode = &postureCode.String
		}
		if postureDisplay.Valid {
			item.PostureDisplay = &postureDisplay.String
		}
		if trackingID.Valid {
			item.TrackingID = &trackingID.String
		}
		
		if bedStatusCode.Valid {
			item.BedStatusSNOMEDCode = &bedStatusCode.String
		}
		if bedStatusDisplay.Valid {
			item.BedStatusDisplay = &bedStatusDisplay.String
		}
		
		if sleepStateCode.Valid {
			item.SleepStateSNOMEDCode = &sleepStateCode.String
		}
		if sleepStateDisplay.Valid {
			item.SleepStateDisplay = &sleepStateDisplay.String
		}
		
		// 设置设备类型（从 JOIN 查询获取，避免额外查询）
		if deviceType.Valid {
			item.DeviceType = deviceType.String
		}
		
		results = append(results, item)
	}
	
	return results, nil
}

// GetLatestByDeviceIDs 批量获取多个设备的最新时序数据（优化 N+1 查询）
// 
// 参数:
//   - tenantID: 租户 ID（用于数据隔离）
//   - deviceIDs: 设备 ID 列表
//   - limit: 每个设备返回的记录数限制
func (r *IoTTimeSeriesRepository) GetLatestByDeviceIDs(tenantID string, deviceIDs []string, limit int) (map[string][]*models.IoTTimeSeries, error) {
	if len(deviceIDs) == 0 {
		return make(map[string][]*models.IoTTimeSeries), nil
	}
	
	// 构建 IN 子句
	query := `
		SELECT 
			its.id,
			its.tenant_id,
			its.device_id,
			its.timestamp,
			its.heart_rate,
			its.heart_rate_code,
			its.heart_rate_display,
			its.respiratory_rate,
			its.respiratory_rate_code,
			its.respiratory_rate_display,
			its.posture_snomed_code,
			its.posture_display,
			its.tracking_id,
			its.bed_status_snomed_code,
			its.bed_status_display,
			its.sleep_state_snomed_code,
			its.sleep_state_display,
			COALESCE(ds.device_type, '') as device_type,
			ROW_NUMBER() OVER (PARTITION BY its.device_id ORDER BY its.timestamp DESC) as rn
		FROM iot_timeseries its
		LEFT JOIN devices d ON its.device_id = d.device_id
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE its.device_id = ANY($1) AND its.tenant_id = $2
	`
	
	rows, err := r.db.Query(query, deviceIDs, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query iot_timeseries: %w", err)
	}
	defer rows.Close()
	
	result := make(map[string][]*models.IoTTimeSeries)
	for rows.Next() {
		item := &models.IoTTimeSeries{}
		var heartRate, respiratoryRate sql.NullInt64
		var heartRateCode, heartRateDisplay sql.NullString
		var respiratoryRateCode, respiratoryRateDisplay sql.NullString
		var postureCode, postureDisplay sql.NullString
		var trackingID sql.NullString
		var bedStatusCode, bedStatusDisplay sql.NullString
		var sleepStateCode, sleepStateDisplay sql.NullString
		var deviceType sql.NullString
		var rowNum int64
		
		err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.DeviceID,
			&item.Timestamp,
			&heartRate,
			&heartRateCode,
			&heartRateDisplay,
			&respiratoryRate,
			&respiratoryRateCode,
			&respiratoryRateDisplay,
			&postureCode,
			&postureDisplay,
			&trackingID,
			&bedStatusCode,
			&bedStatusDisplay,
			&sleepStateCode,
			&sleepStateDisplay,
			&deviceType,
			&rowNum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		// 只取每个设备的前 limit 条记录
		if rowNum > int64(limit) {
			continue
		}
		
		if heartRate.Valid {
			hr := int(heartRate.Int64)
			item.HeartRate = &hr
		}
		if heartRateCode.Valid {
			item.HeartRateCode = &heartRateCode.String
		}
		if heartRateDisplay.Valid {
			item.HeartRateDisplay = &heartRateDisplay.String
		}
		
		if respiratoryRate.Valid {
			rr := int(respiratoryRate.Int64)
			item.RespiratoryRate = &rr
		}
		if respiratoryRateCode.Valid {
			item.RespiratoryRateCode = &respiratoryRateCode.String
		}
		if respiratoryRateDisplay.Valid {
			item.RespiratoryRateDisplay = &respiratoryRateDisplay.String
		}
		
		if postureCode.Valid {
			item.PostureSNOMEDCode = &postureCode.String
		}
		if postureDisplay.Valid {
			item.PostureDisplay = &postureDisplay.String
		}
		if trackingID.Valid {
			item.TrackingID = &trackingID.String
		}
		
		if bedStatusCode.Valid {
			item.BedStatusSNOMEDCode = &bedStatusCode.String
		}
		if bedStatusDisplay.Valid {
			item.BedStatusDisplay = &bedStatusDisplay.String
		}
		
		if sleepStateCode.Valid {
			item.SleepStateSNOMEDCode = &sleepStateCode.String
		}
		if sleepStateDisplay.Valid {
			item.SleepStateDisplay = &sleepStateDisplay.String
		}
		
		if deviceType.Valid {
			item.DeviceType = deviceType.String
		}
		
		result[item.DeviceID] = append(result[item.DeviceID], item)
	}
	
	return result, nil
}

// GetDeviceType 获取设备类型
// 
// 参数:
//   - tenantID: 租户 ID（用于数据隔离）
//   - deviceID: 设备 ID
func (r *IoTTimeSeriesRepository) GetDeviceType(tenantID, deviceID string) (string, error) {
	query := `
		SELECT ds.device_type
		FROM devices d
		INNER JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE d.device_id = $1 AND d.tenant_id = $2
	`
	
	var deviceType string
	err := r.db.QueryRow(query, deviceID, tenantID).Scan(&deviceType)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("device not found: %s", deviceID)
		}
		return "", fmt.Errorf("failed to query device type: %w", err)
	}
	
	return deviceType, nil
}

