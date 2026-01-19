package repository

import (
	"database/sql"
	"encoding/json"
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

// Insert 插入数据到 iot_timeseries 表（窄表结构）
// data: 从 Redis Stream 读取的 map[string]interface{} 数据（已通过 encode 函数转换）
// 新表结构：id, device_id, timestamp, data_type, data_values (JSONB)
func (r *IoTTimeSeriesRepository) Insert(data map[string]interface{}) (int64, error) {
	// 1. 提取基本字段
	deviceID, _ := data["device_id"].(string)
	if deviceID == "" {
		return 0, fmt.Errorf("missing required field: device_id")
	}

	// 3. 提取 timestamp
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

	// 4. 确定 data_type（从 topic_type 映射：monitor, statistics, event, alarm）
	dataType := "monitor" // 默认值
	if topicType, ok := data["topic_type"].(string); ok && topicType != "" {
		// 映射 topic_type 到 data_type
		switch topicType {
		case "monitor":
			dataType = "monitor"
		case "stat":
		dataType = "stat" // 统一使用 "stat"
		case "event":
			dataType = "event"
		case "alarm":
			dataType = "alarm"
		default:
			// 如果 topic_type 不在预期范围内，尝试使用原值
			dataType = topicType
		}
	} else if dataKey, ok := data["data_key"].(string); ok {
		// Sleepace 数据根据 data_key 判断
		if dataKey == "alarmNotify" {
			dataType = "alarm"
		} else if dataKey == "realtime" {
			dataType = "monitor"
		} else if dataKey == "sleepStage" {
			dataType = "statistics"
		}
	}

	// 5. 将所有 encode 后的数据存储在 data_values JSONB 中
	// 直接使用传入的 data map，它已经包含了所有 encode 后的标准字段
	// 注意：data 的字段顺序为：device_id → device_type → tenant_id → timestamp → topic_type → data_value → 位置信息
	// category 字段保留在 data_value 内部，不提取到顶层，避免冗余
	dataValuesJSON, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal data_values: %w", err)
	}

	// 6. 构建 INSERT 语句（窄表结构）
	query := `
		INSERT INTO iot_timeseries (
			device_id,
			timestamp,
			data_type,
			data_values
		) VALUES (
			$1, $2, $3, $4
		)
		RETURNING id
	`

	var id int64
	err = r.db.QueryRow(
		query,
		deviceID,          // $1 (设备 ID)
		timestamp,         // $2 (数据时间戳)
		dataType,          // $3 (数据类型：monitor, statistics, event, alarm)
		dataValuesJSON,    // $4 (所有数据值存储在 JSONB 中)
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert iot_timeseries: %w", err)
	}

	r.logger.Debug("IoT timeseries data inserted",
		zap.Int64("id", id),
		zap.String("device_id", deviceID),
		zap.String("data_type", dataType),
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
			ds.device_uid
		FROM devices d
		JOIN device_store ds ON d.device_id = ds.device_id
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
