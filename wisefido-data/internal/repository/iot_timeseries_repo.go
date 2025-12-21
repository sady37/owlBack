package repository

import (
	"context"
	"time"

	"wisefido-data/internal/domain"
)

// IoTTimeSeriesFilters 时序数据查询过滤器
type IoTTimeSeriesFilters struct {
	DeviceID   string    // 设备ID
	DataType   string    // 'observation'/'alarm'
	Category   string    // FHIR Category
	EventType  string    // 事件类型
	UnitID     string    // 单元ID
	RoomID     string    // 房间ID
	StartTime  *time.Time // 开始时间
	EndTime    *time.Time // 结束时间
	IncludeAlarmEvent bool // 是否包含告警事件信息（需要 JOIN alarm_events）
}

// IoTTimeSeriesRepository IoT时序数据Repository接口
// 注意：此Repository只提供查询方法，数据写入由 wisefido-data-transformer 服务负责
type IoTTimeSeriesRepository interface {
	// GetTimeSeriesData 获取时序数据（按ID）
	GetTimeSeriesData(ctx context.Context, id int64) (*domain.IoTTimeSeries, error)

	// GetLatestData 获取最新数据（按设备）
	GetLatestData(ctx context.Context, tenantID, deviceID string, limit int) ([]*domain.IoTTimeSeries, error)

	// GetDataByDevice 按设备查询（支持过滤）
	GetDataByDevice(ctx context.Context, tenantID, deviceID string, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error)

	// GetDataByResident 按住户查询（通过device关联）
	GetDataByResident(ctx context.Context, tenantID, residentID string, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error)

	// GetDataByTimeRange 时间范围查询
	GetDataByTimeRange(ctx context.Context, tenantID string, startTime, endTime time.Time, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error)

	// GetDataByLocation 按位置查询（unit_id/room_id）
	GetDataByLocation(ctx context.Context, tenantID string, unitID, roomID *string, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error)
}

