package repository

import (
	"context"
	"time"
	"wisefido-data/internal/domain"
)

// AlarmEventsRepository 报警事件Repository接口
// 写操作（Create/Ack/Resolve）已迁移至 owl-common/card，此处仅保留查询方法
type AlarmEventsRepository interface {
	// 查询报警事件列表（支持复杂过滤和跨表 JOIN）
	// 注意：此方法需要支持跨表 JOIN 查询关联数据（设备、卡片、住户、地址信息）
	ListAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters, page, size int) ([]*domain.AlarmEvent, int, error)

	// 获取单个报警事件
	GetAlarmEvent(ctx context.Context, tenantID, eventID string) (*domain.AlarmEvent, error)

	// 更新报警事件（部分更新）
	UpdateAlarmEvent(ctx context.Context, tenantID, eventID string, updates map[string]interface{}) error

	// 获取最近的报警事件（用于去重检查）
	GetRecentAlarmEvent(ctx context.Context, tenantID, deviceID, eventType string, withinMinutes int) (*domain.AlarmEvent, error)

	// 统计报警事件数量（按条件）
	CountAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters) (int, error)
}

// AlarmEventFilters 报警事件过滤条件
// 注意：与 wisefido-ai 的 AlarmEventFilters 保持一致
type AlarmEventFilters struct {
	// 时间段过滤
	StartTime *time.Time // 开始时间（triggered_at >= StartTime）
	EndTime   *time.Time // 结束时间（triggered_at <= EndTime）

	// 住户过滤
	ResidentID *string // 住户ID（通过 device_id JOIN devices → beds → residents 获取）

	// 位置过滤
	BranchTag *string // 分支标签（通过 device_id JOIN devices → beds/rooms → units → units.branch_tag 获取）
	UnitID    *string // 单元ID（通过 device_id JOIN devices → beds/rooms → units 获取）

	// 设备过滤
	DeviceID     *string   // 设备ID（直接过滤）
	DeviceName   *string   // 设备名称（通过 device_id JOIN devices.device_name 获取，支持模糊匹配）
	DeviceIDs    []string  // 设备ID列表（IN 查询）

	// 事件类型和级别过滤
	EventType   *string   // 事件类型（单选）
	EventTypes  []string  // 事件类型列表（IN 查询）
	Category    *string   // 分类（单选）
	Categories  []string  // 分类列表（IN 查询）
	AlarmLevel  *string   // 报警级别（单选）
	AlarmLevels []string  // 报警级别列表（IN 查询）

	// 状态过滤
	AlarmStatus *string   // 报警状态（active, acked, resolved, auto_resolved, expired）
	AlarmStatuses []string // 报警状态列表（IN 查询）

	// 操作结果过滤
	Operation *string   // 操作结果
	Operations []string // 操作结果列表（IN 查询）

	// 处理人过滤
	HandlerID *string // 处理人ID
}
