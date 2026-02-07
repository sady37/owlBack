package repository

import "context"

// CardDeviceInfo 完整的卡片-设备映射信息
// 对应 Redis 存储：field=deviceUID, value=deviceID:cardID:branchID:tenantID
type CardDeviceInfo struct {
	DeviceID  string // 设备 ID
	DeviceUID string // 设备 UID（MQTT 标准字段，用作 Redis field）
	CardID    string // 卡片 ID
	BranchID  string // 分支 ID
	TenantID  string // 租户 ID
}

// CardRepository 卡片数据仓库接口
type CardRepository interface {
	// GetDeviceCardMappings 获取指定租户的所有设备-卡片映射
	// 返回完整对象列表：[]{DeviceID, DeviceUID, CardID, BranchID, TenantID}
	// 用于构建 Redis Hash：field=deviceUID, value=deviceID:cardID:branchID:tenantID
	GetDeviceCardMappings(ctx context.Context, tenantID string) ([]CardDeviceInfo, error)

	// GetDeviceCardMappingsByBranch 获取指定租户指定分支的设备-卡片映射（租户隔离）
	// tenantID 和 branchID 都是必填的
	// 返回完整对象列表：[]{DeviceID, DeviceUID, CardID, BranchID, TenantID}
	GetDeviceCardMappingsByBranch(ctx context.Context, tenantID string, branchID string) ([]CardDeviceInfo, error)
}
