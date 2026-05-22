package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// DeviceStoreRepository 设备库存Repository接口（Phase 2 一刀切：identity = device_uid VARCHAR(50)）。
type DeviceStoreRepository interface {
	// 查询
	ListDeviceStores(ctx context.Context, filters DeviceStoreFilters, page, size int, sort, direction string) ([]*domain.DeviceStore, int, error)
	GetDeviceStore(ctx context.Context, deviceUID string) (*domain.DeviceStore, error)
	GetDeviceStoreByDeviceAddr(ctx context.Context, deviceAddr string) (*domain.DeviceStore, error)

	// 创建（单个设备入库）
	CreateDeviceStore(ctx context.Context, deviceStore *domain.DeviceStore) (string, error)

	// 更新
	BatchUpdateDeviceStores(ctx context.Context, updates []*domain.DeviceStore) error

	// 删除
	DeleteDeviceStore(ctx context.Context, deviceUID string) error

	// 批量导入，返回成功数、新插入行(含 device_uid)、跳过、失败
	ImportDeviceStores(ctx context.Context, items []*domain.DeviceStore) (successCount int, inserted []*domain.DeviceStore, skipped []*domain.DeviceStore, errors []*domain.DeviceStore, err error)

	// 更新 device_code。device_code 仅来自厂家导入文件（导入时已写入）；绑定/initialize 不提供可查写的 device_code，此处供导入或其它有明确来源时更新用
	UpdateDeviceCode(ctx context.Context, deviceUID, deviceCode string) error
	// 仅更新 firmware_version（InitialAll 调 bindInfo 后按 deviceVersion 写回）
	UpdateFirmwareVersion(ctx context.Context, deviceUID, firmwareVersion string) error
}

// DeviceStoreFilters 设备库存查询过滤器
type DeviceStoreFilters struct {
	TenantID        string // 租户ID过滤
	DeviceType      string // 设备类型过滤
	Search          string // 搜索（device_code, device_uid, mac, imei）
	DeviceUID       string
	DeviceCode      string
	DeviceName      string
	FirmwareVersion string
	AllowAccess     string
	OTAStatus       string
	OTAPermit       string
	OTAWay          string
}


