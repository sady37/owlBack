package repository

import (
	"context"
	"time"

	"owl-common/card"
)

// CacheRepository 缓存仓库接口
type CacheRepository interface {
	// GetRealtimeData 获取卡片实时数据
	GetRealtimeData(ctx context.Context, cardID string) (*card.RealtimeData, error)
	// SetRealtimeData 设置卡片实时数据（带 TTL）
	SetRealtimeData(ctx context.Context, cardID string, data *card.RealtimeData, ttl time.Duration) error
	// GetDevicePosture 获取设备姿态数据
	GetDevicePosture(ctx context.Context, cardID string, deviceID string) (*card.DevicePosture, error)
	// SetDevicePosture 设置设备姿态数据
	SetDevicePosture(ctx context.Context, cardID string, posture *card.DevicePosture) error
	// GetVitalSimplified 获取设备生命体征
	GetVitalSimplified(ctx context.Context, cardID string, deviceID string) (*card.VitalSimplified, error)
	// SetVitalSimplified 设置设备生命体征
	SetVitalSimplified(ctx context.Context, cardID string, vital *card.VitalSimplified) error
	// DeleteVitalSimplified 删除设备生命体征缓存
	DeleteVitalSimplified(ctx context.Context, cardID string, deviceID string) error
	// DeleteDevicePosture 删除设备姿态缓存
	DeleteDevicePosture(ctx context.Context, cardID string, deviceID string) error
	// GetAllCardIds 获取所有卡片 ID
	GetAllCardIds(ctx context.Context) ([]string, error)
}