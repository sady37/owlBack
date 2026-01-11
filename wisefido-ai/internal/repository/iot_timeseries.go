package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// IoTTimeSeriesRepository IoT 时序数据仓库（用于报警评估）
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

// LyingBaselineData Lying基线数据
type LyingBaselineData struct {
	Height   float64  // 高度值（radar_pos_z，单位：厘米）
	Position struct { // 位置（radar_pos_x, radar_pos_y）
		X int
		Y int
	}
	Timestamp time.Time // 时间戳
}

// GetLyingBaselineData 获取最近N天夜间2-5点的lying基线数据
// 用于事件1的基线建立
func (r *IoTTimeSeriesRepository) GetLyingBaselineData(
	ctx context.Context,
	tenantID, deviceID string,
	bedAreaID int,
	days int,
) ([]LyingBaselineData, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	if days <= 0 {
		days = 5 // 默认5天
	}

	// 计算时间范围：最近N天
	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	query := `
		SELECT 
			its.radar_pos_z,
			its.radar_pos_x,
			its.radar_pos_y,
			its.timestamp
		FROM iot_timeseries its
		WHERE its.tenant_id = $1
		  AND its.device_id = $2
		  AND its.data_type = 'observation'
		  AND its.area_id = $3
		  AND its.radar_pos_z IS NOT NULL
		  AND its.sleep_state_snomed_code IN ('248232005', '248233000')
		  AND its.timestamp >= $4
		  AND EXTRACT(HOUR FROM its.timestamp) >= 2
		  AND EXTRACT(HOUR FROM its.timestamp) < 5
		ORDER BY its.timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, deviceID, bedAreaID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query lying baseline data: %w", err)
	}
	defer rows.Close()

	var results []LyingBaselineData
	for rows.Next() {
		var data LyingBaselineData
		var height sql.NullInt64
		var posX, posY sql.NullInt64

		err := rows.Scan(
			&height,
			&posX,
			&posY,
			&data.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lying baseline data: %w", err)
		}

		if height.Valid {
			data.Height = float64(height.Int64)
		} else {
			continue
		}

		if posX.Valid {
			data.Position.X = int(posX.Int64)
		}
		if posY.Valid {
			data.Position.Y = int(posY.Int64)
		}

		results = append(results, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate lying baseline data: %w", err)
	}

	return results, nil
}

// CalculateLyingBaseline 计算lying基线（最近5天的平均值）
func (r *IoTTimeSeriesRepository) CalculateLyingBaseline(
	ctx context.Context,
	tenantID, deviceID string,
	bedAreaID int,
) (float64, error) {
	data, err := r.GetLyingBaselineData(ctx, tenantID, deviceID, bedAreaID, 5)
	if err != nil {
		return 0, fmt.Errorf("failed to get lying baseline data: %w", err)
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("no lying baseline data found for device %s, bed area %d", deviceID, bedAreaID)
	}

	var sum float64
	for _, d := range data {
		sum += d.Height
	}

	baseline := sum / float64(len(data))

	r.logger.Info("Calculated lying baseline",
		zap.String("device_id", deviceID),
		zap.Int("bed_area_id", bedAreaID),
		zap.Int("data_count", len(data)),
		zap.Float64("baseline_height", baseline),
	)

	return baseline, nil
}

// GetBedAreaID 获取床的区域ID（area_id）
func (r *IoTTimeSeriesRepository) GetBedAreaID(
	ctx context.Context,
	tenantID, radarDeviceID string,
	bedID string,
) (int, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}
	if radarDeviceID == "" {
		return 0, fmt.Errorf("radar_device_id is required")
	}
	if bedID == "" {
		return 0, fmt.Errorf("bed_id is required")
	}

	query := `
		SELECT DISTINCT its.area_id
		FROM iot_timeseries its
		JOIN devices d ON its.device_id = d.device_id
		JOIN device_store ds ON d.device_store_id = ds.device_store_id
		WHERE its.tenant_id = $1
		  AND its.device_id = $2
		  AND its.data_type = 'observation'
		  AND its.area_id IS NOT NULL
		  AND its.tracking_id IS NOT NULL
		  AND d.bound_bed_id = $3
		  AND ds.device_type = 'Radar'
		  AND its.timestamp >= NOW() - INTERVAL '7 days'
		ORDER BY its.timestamp DESC
		LIMIT 1
	`

	var areaID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, tenantID, radarDeviceID, bedID).Scan(&areaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("bed area_id not found for radar device %s, bed %s", radarDeviceID, bedID)
		}
		return 0, fmt.Errorf("failed to query bed area_id: %w", err)
	}

	if !areaID.Valid {
		return 0, fmt.Errorf("bed area_id is null for radar device %s, bed %s", radarDeviceID, bedID)
	}

	return int(areaID.Int64), nil
}

// GetLatestHeightByDeviceAndTrack 获取设备指定track_id的最新高度数据
func (r *IoTTimeSeriesRepository) GetLatestHeightByDeviceAndTrack(
	ctx context.Context,
	tenantID, deviceID, trackID string,
) (float64, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return 0, fmt.Errorf("device_id is required")
	}
	if trackID == "" {
		return 0, fmt.Errorf("track_id is required")
	}

	query := `
		SELECT its.radar_pos_z
		FROM iot_timeseries its
		WHERE its.tenant_id = $1
		  AND its.device_id = $2
		  AND its.tracking_id::text = $3
		  AND its.radar_pos_z IS NOT NULL
		  AND its.data_type = 'observation'
		ORDER BY its.timestamp DESC
		LIMIT 1
	`

	var height sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceID, trackID).Scan(&height)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no height data found for device %s, track %s", deviceID, trackID)
		}
		return 0, fmt.Errorf("failed to query latest height: %w", err)
	}

	if !height.Valid {
		return 0, fmt.Errorf("height data is null for device %s, track %s", deviceID, trackID)
	}

	return float64(height.Int64), nil
}

// GetLatestPositionByDeviceAndTrack 获取设备指定track_id的最新位置数据
// 用于位置变化检测
func (r *IoTTimeSeriesRepository) GetLatestPositionByDeviceAndTrack(
	ctx context.Context,
	tenantID, deviceID, trackID string,
) (int, int, error) {
	if tenantID == "" {
		return 0, 0, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return 0, 0, fmt.Errorf("device_id is required")
	}
	if trackID == "" {
		return 0, 0, fmt.Errorf("track_id is required")
	}

	query := `
		SELECT its.radar_pos_x, its.radar_pos_y
		FROM iot_timeseries its
		WHERE its.tenant_id = $1
		  AND its.device_id = $2
		  AND its.tracking_id::text = $3
		  AND its.radar_pos_x IS NOT NULL
		  AND its.radar_pos_y IS NOT NULL
		  AND its.data_type = 'observation'
		ORDER BY its.timestamp DESC
		LIMIT 1
	`

	var posX, posY sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceID, trackID).Scan(&posX, &posY)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, fmt.Errorf("no position data found for device %s, track %s", deviceID, trackID)
		}
		return 0, 0, fmt.Errorf("failed to query latest position: %w", err)
	}

	if !posX.Valid || !posY.Valid {
		return 0, 0, fmt.Errorf("position data is null for device %s, track %s", deviceID, trackID)
	}

	return int(posX.Int64), int(posY.Int64), nil
}

// CheckNewMovableTrack 检查是否有新的可移动track出现
// 用于T0+10秒检测：如果track消失，检查30秒内是否有新的可移动track
func (r *IoTTimeSeriesRepository) CheckNewMovableTrack(
	ctx context.Context,
	tenantID, deviceID, originalTrackID string,
	sinceTime time.Time,
) (bool, error) {
	if tenantID == "" {
		return false, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return false, fmt.Errorf("device_id is required")
	}

	query := `
		SELECT COUNT(DISTINCT its.tracking_id)
		FROM iot_timeseries its
		WHERE its.tenant_id = $1
		  AND its.device_id = $2
		  AND its.timestamp >= $3
		  AND its.tracking_id IS NOT NULL
		  AND its.tracking_id::text != $4
		  AND its.radar_pos_x IS NOT NULL
		  AND its.radar_pos_y IS NOT NULL
		  AND its.data_type = 'observation'
	`

	var count sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceID, sinceTime, originalTrackID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check new movable track: %w", err)
	}

	return count.Valid && count.Int64 > 0, nil
}
