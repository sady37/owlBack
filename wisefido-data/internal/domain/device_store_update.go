package domain

// DeviceStoreUpdate 设备库存更新模型（对应 device_store 表）
// 用于区分"不更新"、"更新"、"删除"三种状态
// 注意：主键字段（device_store_id）和必填字段（device_type, tenant_id, import_date, allow_access）不在更新模型中
type DeviceStoreUpdate struct {
	// 设备类型（NOT NULL，但更新时可以为空，表示不更新）
	DeviceType *UpdateString

	// 设备型号
	DeviceModel *UpdateString

	// 序列号/UID/IMEI
	SerialNumber *UpdateString
	UID          *UpdateString
	IMEI         *UpdateString

	// 物理属性
	CommMode *UpdateString
	MCUModel *UpdateString

	// 固件版本
	FirmwareVersion          *UpdateString
	OTATargetFirmwareVersion *UpdateString
	OTATargetMCUModel        *UpdateString

	// 租户分配
	TenantID *UpdateString // NOT NULL，但更新时可以为空，表示不更新

	// 时间戳
	AllocateTime *UpdateTime // nullable

	// 系统级访问权限
	AllowAccess *UpdateBool // NOT NULL，但更新时可以为空，表示不更新
}

