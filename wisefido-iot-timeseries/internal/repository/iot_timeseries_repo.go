package repository

import (
	"database/sql"
	"fmt"
	"time"

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

// Insert 插入数据到 iot_timeseries 表
// data: 从 Redis Stream 读取的 map[string]interface{} 数据（已通过 encode 函数转换）
// 根据 HIPAA/FDA 要求，只保存转换后的标准值，不保存原始数据
func (r *IoTTimeSeriesRepository) Insert(data map[string]interface{}) (int64, error) {
	// 1. 提取基本字段
	tenantID, _ := data["tenant_id"].(string)
	deviceID, _ := data["device_id"].(string)
	
	if tenantID == "" || deviceID == "" {
		return 0, fmt.Errorf("missing required fields: tenant_id=%s, device_id=%s", tenantID, deviceID)
	}

	// 提取 timestamp
	var timestamp time.Time
	if ts, ok := data["timestamp"]; ok {
		switch v := ts.(type) {
		case int64:
			timestamp = time.Unix(v, 0)
		case float64:
			timestamp = time.Unix(int64(v), 0)
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				timestamp = t
			} else {
				timestamp = time.Now()
			}
		default:
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	// 2. 确定 data_type（默认为 observation）
	dataType := "observation"
	if dt, ok := data["data_type"].(string); ok && dt != "" {
		dataType = dt
	} else if topicType, ok := data["topic_type"].(string); ok {
		if topicType == "alarm" {
			dataType = "alarm"
		}
	} else if dataKey, ok := data["data_key"].(string); ok {
		if dataKey == "alarmNotify" {
			dataType = "alarm"
		}
	}

	// 3. 获取设备硬件信息（冗余存储，避免 JOIN 查询，提高查询性能）
	deviceType, deviceModel, serialNumber, uid, err := r.GetDeviceHardwareInfo(deviceID)
	if err != nil {
		r.logger.Warn("Failed to get device hardware info",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		// 硬件信息获取失败不影响数据插入，继续处理（deviceType 等为 nil）
	}

	// 4. 获取设备位置信息（unit_id, room_id）
	// 注意：位置信息在 INSERT 时直接插入，避免后续 UPDATE 操作
	// 设备位置信息变化频率低（设备安装、迁移、卸载），但数据插入频率高（每秒或每分钟多次）
	// 后续优化：使用缓存机制进一步减少查询开销
	unitID, roomID, err := r.GetDeviceLocation(deviceID)
	if err != nil {
		r.logger.Warn("Failed to get device location",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		// 位置信息获取失败不影响数据插入，继续处理（unitID, roomID 为 nil）
	}

	// 5. 提取 category（可选）
	var category sql.NullString
	if cat, ok := data["category"].(string); ok && cat != "" {
		category = sql.NullString{String: cat, Valid: true}
	}

	// 6. 提取转换后的标准值字段
	trackingID, posX, posY, posZ := extractTrackingFields(data)
	postureCode, postureDisplay := extractPostureFields(data)
	eventType, eventCode, eventDisplay, areaID := extractEventFields(data)
	hrCode, hrDisplay, hr, rrCode, rrDisplay, rr := extractVitalSignsFields(data)
	sleepCode, sleepDisplay := extractSleepStateFields(data)

	// 7. 提取其他字段
	var confidence, remainingTime sql.NullInt64
	if conf, ok := data["confidence"]; ok {
		if c, err := parseInt(conf); err == nil {
			confidence = sql.NullInt64{Int64: int64(c), Valid: true}
		}
	}
	if rt, ok := data["remaining_time"]; ok {
		if r, err := parseInt(rt); err == nil {
			remainingTime = sql.NullInt64{Int64: int64(r), Valid: true}
		}
	}

	// 8. 构建 metadata JSONB（保存统计数据和扩展信息）
	metadataJSON, err := buildMetadata(data)
	if err != nil {
		r.logger.Warn("Failed to build metadata",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		metadataJSON = []byte("{}")
	}

	// 9. 构建最小化审计数据（用于 raw_original）
	// 根据 HIPAA/FDA 要求和用户要求，不保存原始数据，只保存必要的审计追溯信息
	rawOriginal, err := buildMinimalAuditData(data)
	if err != nil {
		return 0, fmt.Errorf("failed to build minimal audit data: %w", err)
	}

	// 10. 构建完整的 INSERT 语句（包含位置信息）
	query := `
		INSERT INTO iot_timeseries (
			tenant_id,
			device_id,
			device_type,
			device_model,
			serial_number,
			uid,
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
			confidence,
			remaining_time,
			raw_original,
			raw_format,
			raw_compression,
			metadata,
			unit_id,
			room_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35
		)
		RETURNING id
	`

	var id int64
	err = r.db.QueryRow(
		query,
		tenantID,           // $1
		deviceID,           // $2
		deviceType,         // $3 (硬件信息：device_type)
		deviceModel,        // $4 (硬件信息：device_model)
		serialNumber,       // $5 (硬件信息：serial_number)
		uid,                // $6 (硬件信息：uid)
		timestamp,          // $7 (数据时间戳，设备采集时间 = 上传时间)
		dataType,           // $8
		category,           // $9
		trackingID,         // $10
		posX,               // $11
		posY,               // $12
		posZ,               // $13
		postureCode,        // $14
		postureDisplay,     // $15
		eventType,          // $16
		eventCode,          // $17
		eventDisplay,       // $18
		areaID,             // $19
		hrCode,             // $20
		hrDisplay,          // $21
		hr,                 // $22
		rrCode,             // $23
		rrDisplay,          // $24
		rr,                 // $25
		sleepCode,          // $26
		sleepDisplay,       // $27
		confidence,         // $28
		remainingTime,      // $29
		rawOriginal,        // $30 (最小化审计数据)
		"json",             // $31 (raw_format)
		nil,                // $32 (raw_compression = NULL)
		metadataJSON,       // $33 (metadata JSONB)
		unitID,             // $34 (位置信息：unit_id，在 INSERT 时直接插入，避免后续 UPDATE)
		roomID,             // $35 (位置信息：room_id，在 INSERT 时直接插入，避免后续 UPDATE)
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert iot_timeseries: %w", err)
	}

	r.logger.Debug("IoT timeseries data inserted",
		zap.Int64("id", id),
		zap.String("device_id", deviceID),
		zap.String("data_type", dataType),
		zap.String("category", category.String),
	)

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

// GetDeviceHardwareInfo 获取设备硬件信息（从 device_store 表）
// 通过 device_id → devices → device_store 关联获取硬件信息
// 返回：device_type, device_model, serial_number, uid
func (r *IoTTimeSeriesRepository) GetDeviceHardwareInfo(deviceID string) (deviceType, deviceModel, serialNumber, uid *string, err error) {
	query := `
		SELECT 
			ds.device_type,
			ds.device_model,
			ds.serial_number,
			ds.uid
		FROM devices d
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE d.device_id = $1
		LIMIT 1
	`

	var dt, dm, sn, u sql.NullString
	err = r.db.QueryRow(query, deviceID).Scan(&dt, &dm, &sn, &u)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil, nil, fmt.Errorf("device not found: %s", deviceID)
		}
		return nil, nil, nil, nil, fmt.Errorf("failed to query device hardware info: %w", err)
	}

	if dt.Valid {
		deviceType = &dt.String
	}
	if dm.Valid {
		deviceModel = &dm.String
	}
	if sn.Valid {
		serialNumber = &sn.String
	}
	if u.Valid {
		uid = &u.String
	}

	return deviceType, deviceModel, serialNumber, uid, nil
}

// UpdateLocation 更新记录的 unit_id 和 room_id
// 注意：此方法已不再使用，位置信息在 INSERT 时直接插入
// 保留此方法以保持向后兼容性
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

// InvalidateLocationCache 清除设备位置信息缓存
// 当设备绑定关系变化时调用此方法清除缓存
// 注意：如果还没有实现缓存机制，此方法为空实现（为后续缓存做准备）
func (r *IoTTimeSeriesRepository) InvalidateLocationCache(deviceID string) {
	// TODO: 实现缓存机制后，在此清除缓存
	// 例如：r.locationCache.Delete(deviceID)
	r.logger.Debug("Location cache invalidated (cache not implemented yet)",
		zap.String("device_id", deviceID),
	)
}
