package repository

import (
	"database/sql"
	"fmt"
	"wisefido-card-aggregator/internal/models"

	"github.com/lib/pq"
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
	// tenant_id 存储在 data_values JSONB 字段中
	var query string
	var args []interface{}
	
	if tenantID == "*" {
		// 查询所有租户的数据
		query = `
			SELECT 
				its.id,
				its.data_values->>'tenant_id' as tenant_id,
				its.device_id,
				its.timestamp,
				its.data_values->'data_value'->>'heart_rate' as heart_rate,
				its.data_values->'data_value'->>'heart_rate_code' as heart_rate_code,
				its.data_values->'data_value'->>'heart_rate_display' as heart_rate_display,
				its.data_values->'data_value'->>'respiratory_rate' as respiratory_rate,
				its.data_values->'data_value'->>'respiratory_rate_code' as respiratory_rate_code,
				its.data_values->'data_value'->>'respiratory_rate_display' as respiratory_rate_display,
				its.data_values->'data_value'->>'pose_snomed_code' as posture_snomed_code,
				its.data_values->'data_value'->>'pose_snomed_display' as posture_display,
				its.data_values->'data_value'->>'target_id' as tracking_id,
				its.data_values->'data_value'->>'bed_status_snomed_code' as bed_status_snomed_code,
				its.data_values->'data_value'->>'bed_status_display' as bed_status_display,
				its.data_values->'data_value'->>'sleep_state_snomed_code' as sleep_state_snomed_code,
				its.data_values->'data_value'->>'sleep_state_display' as sleep_state_display,
				its.data_values->'data_value'->>'position_x' as radar_pos_x,
				its.data_values->'data_value'->>'position_y' as radar_pos_y,
				its.data_values->'data_value'->>'position_z' as radar_pos_z,
				its.data_values->'data_value'->>'area_id' as area_id,
				COALESCE(ds.device_type, '') as device_type
			FROM iot_timeseries its
			LEFT JOIN devices d ON its.device_id = d.device_id
			LEFT JOIN device_store ds ON d.device_id = ds.device_id
			WHERE its.device_id = $1
			ORDER BY its.timestamp DESC
			LIMIT $2
		`
		args = []interface{}{deviceID, limit}
	} else {
		// 查询指定租户的数据
		query = `
			SELECT 
				its.id,
				its.data_values->>'tenant_id' as tenant_id,
				its.device_id,
				its.timestamp,
				its.data_values->'data_value'->>'heart_rate' as heart_rate,
				its.data_values->'data_value'->>'heart_rate_code' as heart_rate_code,
				its.data_values->'data_value'->>'heart_rate_display' as heart_rate_display,
				its.data_values->'data_value'->>'respiratory_rate' as respiratory_rate,
				its.data_values->'data_value'->>'respiratory_rate_code' as respiratory_rate_code,
				its.data_values->'data_value'->>'respiratory_rate_display' as respiratory_rate_display,
				its.data_values->'data_value'->>'pose_snomed_code' as posture_snomed_code,
				its.data_values->'data_value'->>'pose_snomed_display' as posture_display,
				its.data_values->'data_value'->>'target_id' as tracking_id,
				its.data_values->'data_value'->>'bed_status_snomed_code' as bed_status_snomed_code,
				its.data_values->'data_value'->>'bed_status_display' as bed_status_display,
				its.data_values->'data_value'->>'sleep_state_snomed_code' as sleep_state_snomed_code,
				its.data_values->'data_value'->>'sleep_state_display' as sleep_state_display,
				its.data_values->'data_value'->>'position_x' as radar_pos_x,
				its.data_values->'data_value'->>'position_y' as radar_pos_y,
				its.data_values->'data_value'->>'position_z' as radar_pos_z,
				its.data_values->'data_value'->>'area_id' as area_id,
				COALESCE(ds.device_type, '') as device_type
			FROM iot_timeseries its
			LEFT JOIN devices d ON its.device_id = d.device_id
			LEFT JOIN device_store ds ON d.device_id = ds.device_id
			WHERE its.device_id = $1 AND its.data_values->>'tenant_id' = $2
			ORDER BY its.timestamp DESC
			LIMIT $3
		`
		args = []interface{}{deviceID, tenantID, limit}
	}

	rows, err := r.db.Query(query, args...)
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
		var posX, posY, posZ, areaID sql.NullInt64
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
			&posX,
			&posY,
			&posZ,
			&areaID,
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

		// 设置位置和高度数据
		if posX.Valid {
			x := int(posX.Int64)
			item.PositionX = &x
		}
		if posY.Valid {
			y := int(posY.Int64)
			item.PositionY = &y
		}
		if posZ.Valid {
			z := int(posZ.Int64)
			item.PositionZ = &z
		}
		if areaID.Valid {
			a := int(areaID.Int64)
			item.AreaID = &a
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
//   - tenantID: 租户 ID（用于数据隔离）。如果为 "*"，则查询所有租户的数据
//   - deviceIDs: 设备 ID 列表
//   - limit: 每个设备返回的记录数限制
//
// 查询逻辑：
//   1. 如果 tenantID 为 "*"：直接用 deviceID 在 iot_timeseries 中查找（device_id 已经唯一标识设备）
//   2. 如果 tenantID 有值：先验证 deviceID 是否属于该 tenant_id，然后查询
func (r *IoTTimeSeriesRepository) GetLatestByDeviceIDs(tenantID string, deviceIDs []string, limit int) (map[string][]*models.IoTTimeSeries, error) {
	if len(deviceIDs) == 0 {
		return make(map[string][]*models.IoTTimeSeries), nil
	}

	// 如果 tenantID 不为 "*"，先验证设备是否属于该租户
	if tenantID != "*" {
		// 查询这些设备是否属于指定租户
		checkQuery := `
			SELECT COUNT(*) 
			FROM devices 
			WHERE device_id = ANY($1::uuid[]) AND tenant_id = $2
		`
		var count int
		err := r.db.QueryRow(checkQuery, pq.Array(deviceIDs), tenantID).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to verify device tenant: %w", err)
		}
		if count != len(deviceIDs) {
			return nil, fmt.Errorf("some devices do not belong to tenant: %s", tenantID)
		}
	}

	// 直接用 deviceID 查询，不再需要 tenant_id 过滤（device_id 已经唯一标识设备）
	query := `
		SELECT 
			its.id,
			its.data_values->>'tenant_id' as tenant_id,
			its.device_id,
			its.timestamp,
			its.data_values->'data_value'->>'heart_rate' as heart_rate,
			its.data_values->'data_value'->>'heart_rate_code' as heart_rate_code,
			its.data_values->'data_value'->>'heart_rate_display' as heart_rate_display,
			its.data_values->'data_value'->>'respiratory_rate' as respiratory_rate,
			its.data_values->'data_value'->>'respiratory_rate_code' as respiratory_rate_code,
			its.data_values->'data_value'->>'respiratory_rate_display' as respiratory_rate_display,
			its.data_values->'data_value'->>'pose_snomed_code' as posture_snomed_code,
			its.data_values->'data_value'->>'pose_snomed_display' as posture_display,
			its.data_values->'data_value'->>'target_id' as tracking_id,
			its.data_values->'data_value'->>'bed_status_snomed_code' as bed_status_snomed_code,
			its.data_values->'data_value'->>'bed_status_display' as bed_status_display,
			its.data_values->'data_value'->>'sleep_state_snomed_code' as sleep_state_snomed_code,
			its.data_values->'data_value'->>'sleep_state_display' as sleep_state_display,
			its.data_values->'data_value'->>'position_x' as radar_pos_x,
			its.data_values->'data_value'->>'position_y' as radar_pos_y,
			its.data_values->'data_value'->>'position_z' as radar_pos_z,
			its.data_values->'data_value'->>'area_id' as area_id,
			COALESCE(ds.device_type, '') as device_type,
			ROW_NUMBER() OVER (PARTITION BY its.device_id ORDER BY its.timestamp DESC) as rn
		FROM iot_timeseries its
		LEFT JOIN devices d ON its.device_id = d.device_id
		LEFT JOIN device_store ds ON d.device_id = ds.device_id
		WHERE its.device_id = ANY($1::uuid[])
		`
	args := []interface{}{pq.Array(deviceIDs)}

	rows, err := r.db.Query(query, args...)

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
		var posX, posY, posZ, areaID sql.NullInt64
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
			&posX,
			&posY,
			&posZ,
			&areaID,
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

		// 设置位置和高度数据
		if posX.Valid {
			x := int(posX.Int64)
			item.PositionX = &x
		}
		if posY.Valid {
			y := int(posY.Int64)
			item.PositionY = &y
		}
		if posZ.Valid {
			z := int(posZ.Int64)
			item.PositionZ = &z
		}
		if areaID.Valid {
			a := int(areaID.Int64)
			item.AreaID = &a
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
		INNER JOIN device_store ds ON d.device_id = ds.device_id
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
