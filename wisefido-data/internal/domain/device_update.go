package domain

// DeviceUpdate 设备更新模型（对应 devices 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（device_id, tenant_id）和必填字段（device_name, status, business_access, monitoring_enabled）不在更新模型中
type DeviceUpdate struct {
	// 关联 device_store
	DeviceStoreID *UpdateString // UUID, nullable

	// 标识/资产
	DeviceName   *UpdateString // NOT NULL（但更新时可以为空，表示不更新）
	SerialNumber *UpdateString // nullable
	UID          *UpdateString // nullable

	// 位置绑定（互斥）
	BoundRoomID *UpdateString // UUID, nullable
	BoundBedID  *UpdateString // UUID, nullable

	// 状态/维护
	Status            *UpdateString // NOT NULL（但更新时可以为空，表示不更新）
	BusinessAccess    *UpdateString // NOT NULL（但更新时可以为空，表示不更新）
	MonitoringEnabled *UpdateBool   // NOT NULL（但更新时可以为空，表示不更新）

	// 元数据
	Metadata *UpdateJSON // JSONB, nullable
}

