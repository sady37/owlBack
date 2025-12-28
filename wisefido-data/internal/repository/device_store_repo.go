package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// DeviceStoreRepository 设备库存Repository接口
// 使用强类型领域模型，不使用map[string]any
type DeviceStoreRepository interface {
	// 查询
	ListDeviceStores(ctx context.Context, filters DeviceStoreFilters, page, size int) ([]*domain.DeviceStore, int, error)
	GetDeviceStore(ctx context.Context, deviceStoreID string) (*domain.DeviceStore, error)

	// 创建（单个设备入库）
	CreateDeviceStore(ctx context.Context, deviceStore *domain.DeviceStore) (string, error)

	// 更新
	// Deprecated: 使用 BatchUpdateDeviceStoresFields 替代，支持区分"不更新"、"更新"、"删除"三种状态
	BatchUpdateDeviceStores(ctx context.Context, updates []*domain.DeviceStore) error

	// BatchUpdateDeviceStoresFields 批量更新设备库存（使用更新模型）
	// 支持区分"不更新"、"更新"、"删除"三种状态
	BatchUpdateDeviceStoresFields(ctx context.Context, updates []*DeviceStoreUpdateItem) error

	// 删除
	DeleteDeviceStore(ctx context.Context, deviceStoreID string) error

	// 批量导入
	ImportDeviceStores(ctx context.Context, items []*domain.DeviceStore) (int, []*domain.DeviceStore, []*domain.DeviceStore, error)
}

// DeviceStoreFilters 设备库存查询过滤器
type DeviceStoreFilters struct {
	TenantID   string // 租户ID过滤
	DeviceType string // 设备类型过滤
	Search     string // 搜索（serial_number, uid, imei）
}

// DeviceStoreUpdateItem 设备库存更新项（包含设备ID和更新内容）
type DeviceStoreUpdateItem struct {
	DeviceStoreID string                  // 设备库存ID
	Update        *domain.DeviceStoreUpdate // 更新内容
}


