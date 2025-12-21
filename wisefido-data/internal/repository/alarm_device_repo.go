package repository

import (
	"context"

	"wisefido-data/internal/domain"
)

// AlarmDeviceRepository 设备告警配置Repository接口
type AlarmDeviceRepository interface {
	// GetAlarmDevice 获取设备的告警配置
	GetAlarmDevice(ctx context.Context, tenantID, deviceID string) (*domain.AlarmDevice, error)

	// UpsertAlarmDevice 创建或更新设备的告警配置
	// 注意：UNIQUE(device_id)，使用UPSERT语义
	UpsertAlarmDevice(ctx context.Context, tenantID, deviceID string, alarmDevice *domain.AlarmDevice) error

	// DeleteAlarmDevice 删除设备的告警配置
	DeleteAlarmDevice(ctx context.Context, tenantID, deviceID string) error

	// ListAlarmDevices 批量查询设备的告警配置（支持分页）
	ListAlarmDevices(ctx context.Context, tenantID string, page, size int) ([]*domain.AlarmDevice, int, error)
}

