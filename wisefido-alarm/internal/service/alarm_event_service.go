package service

import (
	"context"
	"fmt"
	"wisefido-alarm/internal/models"
	"wisefido-alarm/internal/repository"

	"go.uber.org/zap"
)

// AlarmEventService 报警事件服务层
// 职责：
// 1. 业务规则验证
// 2. 数据转换（JSON ↔ 领域模型）
// 3. 业务编排（协调多个 Repository）
// 4. 事务管理（跨 Repository 的事务）
// 5. 权限检查（如需要）
type AlarmEventService struct {
	alarmEventsRepo *repository.AlarmEventsRepository
	logger          *zap.Logger
}

// NewAlarmEventService 创建报警事件服务
func NewAlarmEventService(
	alarmEventsRepo *repository.AlarmEventsRepository,
	logger *zap.Logger,
) *AlarmEventService {
	return &AlarmEventService{
		alarmEventsRepo: alarmEventsRepo,
		logger:          logger,
	}
}

// ============================================
// 查询相关方法
// ============================================

// ListAlarmEvents 查询报警事件列表（支持多条件过滤和分页）
// 业务规则：
// - tenant_id 必填
// - page 和 size 必须 > 0
// - 自动过滤软删除的记录
func (s *AlarmEventService) ListAlarmEvents(
	ctx context.Context,
	tenantID string,
	filters repository.AlarmEventFilters,
	page, size int,
) ([]*models.AlarmEvent, int, error) {
	// 业务规则验证
	if tenantID == "" {
		return nil, 0, fmt.Errorf("tenant_id is required")
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20 // 默认每页 20 条
	}
	if size > 100 {
		size = 100 // 最大每页 100 条
	}

	// 调用 Repository
	events, total, err := s.alarmEventsRepo.ListAlarmEvents(ctx, tenantID, filters, page, size)
	if err != nil {
		s.logger.Error("Failed to list alarm events",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to list alarm events: %w", err)
	}

	return events, total, nil
}

// GetAlarmEvent 获取单个报警事件
// 业务规则：
// - tenant_id 和 event_id 必填
// - 自动过滤软删除的记录
func (s *AlarmEventService) GetAlarmEvent(
	ctx context.Context,
	tenantID, eventID string,
) (*models.AlarmEvent, error) {
	// 业务规则验证
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return nil, fmt.Errorf("event_id is required")
	}

	// 调用 Repository
	event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
	if err != nil {
		s.logger.Error("Failed to get alarm event",
			zap.String("tenant_id", tenantID),
			zap.String("event_id", eventID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get alarm event: %w", err)
	}

	return event, nil
}

// CountAlarmEvents 统计报警事件数量
// 业务规则：
// - tenant_id 必填
func (s *AlarmEventService) CountAlarmEvents(
	ctx context.Context,
	tenantID string,
	filters repository.AlarmEventFilters,
) (int, error) {
	// 业务规则验证
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}

	// 调用 Repository
	count, err := s.alarmEventsRepo.CountAlarmEvents(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to count alarm events",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to count alarm events: %w", err)
	}

	return count, nil
}

// ============================================
// 状态管理方法
// ============================================

// AcknowledgeAlarmEvent 确认报警事件
// 业务规则：
// - tenant_id 和 event_id 必填
// - handler_id 必填（确认人）
// - 只能确认状态为 'active' 的报警
// - 自动设置 hand_time 为当前时间
func (s *AlarmEventService) AcknowledgeAlarmEvent(
	ctx context.Context,
	tenantID, eventID, handlerID string,
) error {
	// 业务规则验证
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if handlerID == "" {
		return fmt.Errorf("handler_id is required")
	}

	// 先获取报警事件，检查状态
	event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
	if err != nil {
		return fmt.Errorf("failed to get alarm event: %w", err)
	}

	// 业务规则：只能确认状态为 'active' 的报警
	if event.AlarmStatus != "active" {
		return fmt.Errorf("can only acknowledge active alarms, current status: %s", event.AlarmStatus)
	}

	// 调用 Repository
	if err := s.alarmEventsRepo.AcknowledgeAlarmEvent(ctx, tenantID, eventID, handlerID); err != nil {
		s.logger.Error("Failed to acknowledge alarm event",
			zap.String("tenant_id", tenantID),
			zap.String("event_id", eventID),
			zap.String("handler_id", handlerID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to acknowledge alarm event: %w", err)
	}

	s.logger.Info("Alarm event acknowledged",
		zap.String("tenant_id", tenantID),
		zap.String("event_id", eventID),
		zap.String("handler_id", handlerID),
	)

	return nil
}

// UpdateAlarmEventOperation 更新报警事件操作结果
// 业务规则：
// - tenant_id 和 event_id 必填
// - operation 必填（如 'verified_and_processed', 'false_alarm'）
// - handler_id 必填（操作人）
// - 只能更新状态为 'active' 或 'acknowledged' 的报警
// - 自动设置 hand_time 为当前时间
func (s *AlarmEventService) UpdateAlarmEventOperation(
	ctx context.Context,
	tenantID, eventID, operation, handlerID string,
	notes *string,
) error {
	// 业务规则验证
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if operation == "" {
		return fmt.Errorf("operation is required")
	}
	if handlerID == "" {
		return fmt.Errorf("handler_id is required")
	}

	// 验证 operation 值
	validOperations := []string{
		"verified_and_processed",
		"false_alarm",
		"resolved",
		"escalated",
		"cancelled",
	}
	isValid := false
	for _, validOp := range validOperations {
		if operation == validOp {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid operation: %s, valid values: %v", operation, validOperations)
	}

	// 先获取报警事件，检查状态
	event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
	if err != nil {
		return fmt.Errorf("failed to get alarm event: %w", err)
	}

	// 业务规则：只能更新状态为 'active' 或 'acknowledged' 的报警
	if event.AlarmStatus != "active" && event.AlarmStatus != "acknowledged" {
		return fmt.Errorf("can only update operation for active or acknowledged alarms, current status: %s", event.AlarmStatus)
	}

	// 调用 Repository
	if err := s.alarmEventsRepo.UpdateAlarmEventOperation(ctx, tenantID, eventID, operation, handlerID, notes); err != nil {
		s.logger.Error("Failed to update alarm event operation",
			zap.String("tenant_id", tenantID),
			zap.String("event_id", eventID),
			zap.String("operation", operation),
			zap.String("handler_id", handlerID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update alarm event operation: %w", err)
	}

	s.logger.Info("Alarm event operation updated",
		zap.String("tenant_id", tenantID),
		zap.String("event_id", eventID),
		zap.String("operation", operation),
		zap.String("handler_id", handlerID),
	)

	return nil
}

// ============================================
// CRUD 方法
// ============================================

// CreateAlarmEvent 创建报警事件
// 业务规则：
// - tenant_id 必填
// - event 必填且 tenant_id 必须匹配
// - event_id 必须为空（自动生成）
// - triggered_at 必须设置
// - alarm_status 默认为 'active'
func (s *AlarmEventService) CreateAlarmEvent(
	ctx context.Context,
	tenantID string,
	event *models.AlarmEvent,
) error {
	// 业务规则验证
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if event == nil {
		return fmt.Errorf("event is required")
	}
	if event.TenantID != tenantID {
		return fmt.Errorf("event tenant_id (%s) does not match provided tenant_id (%s)", event.TenantID, tenantID)
	}
	if event.EventID == "" {
		return fmt.Errorf("event_id is required (should be generated by builder)")
	}
	if event.TriggeredAt.IsZero() {
		return fmt.Errorf("triggered_at is required")
	}
	if event.AlarmStatus == "" {
		event.AlarmStatus = "active" // 默认状态
	}

	// 调用 Repository
	if err := s.alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, event); err != nil {
		s.logger.Error("Failed to create alarm event",
			zap.String("tenant_id", tenantID),
			zap.String("event_id", event.EventID),
			zap.String("event_type", event.EventType),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create alarm event: %w", err)
	}

	s.logger.Info("Alarm event created",
		zap.String("tenant_id", tenantID),
		zap.String("event_id", event.EventID),
		zap.String("event_type", event.EventType),
		zap.String("alarm_level", event.AlarmLevel),
	)

	return nil
}

// UpdateAlarmEvent 更新报警事件（部分更新）
// 业务规则：
// - tenant_id 和 event_id 必填
// - updates 不能为空
// - 只能更新允许的字段
// - 不能更新 event_id, tenant_id, device_id, created_at
func (s *AlarmEventService) UpdateAlarmEvent(
	ctx context.Context,
	tenantID, eventID string,
	updates map[string]interface{},
) error {
	// 业务规则验证
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if len(updates) == 0 {
		return fmt.Errorf("updates cannot be empty")
	}

	// 定义允许更新的字段
	allowedFields := map[string]bool{
		"notes": true,
		// 注意：alarm_status, handler, operation, hand_time 应该通过专门的方法更新
		// 这里只允许更新 notes
	}

	// 验证字段
	for field := range updates {
		if !allowedFields[field] {
			return fmt.Errorf("field '%s' is not allowed to be updated directly, use specific methods instead", field)
		}
	}

	// 调用 Repository
	if err := s.alarmEventsRepo.UpdateAlarmEvent(ctx, tenantID, eventID, updates); err != nil {
		s.logger.Error("Failed to update alarm event",
			zap.String("tenant_id", tenantID),
			zap.String("event_id", eventID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update alarm event: %w", err)
	}

	s.logger.Info("Alarm event updated",
		zap.String("tenant_id", tenantID),
		zap.String("event_id", eventID),
	)

	return nil
}

// DeleteAlarmEvent 删除报警事件（软删除）
// 业务规则：
// - tenant_id 和 event_id 必填
// - 软删除（设置 metadata->>'deleted_at'）
func (s *AlarmEventService) DeleteAlarmEvent(
	ctx context.Context,
	tenantID, eventID string,
) error {
	// 业务规则验证
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}

	// 调用 Repository
	if err := s.alarmEventsRepo.DeleteAlarmEvent(ctx, tenantID, eventID); err != nil {
		s.logger.Error("Failed to delete alarm event",
			zap.String("tenant_id", tenantID),
			zap.String("event_id", eventID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete alarm event: %w", err)
	}

	s.logger.Info("Alarm event deleted",
		zap.String("tenant_id", tenantID),
		zap.String("event_id", eventID),
	)

	return nil
}

// ============================================
// 便捷查询方法
// ============================================

// GetActiveAlarmEvents 获取活跃的报警事件
func (s *AlarmEventService) GetActiveAlarmEvents(
	ctx context.Context,
	tenantID string,
	filters repository.AlarmEventFilters,
	page, size int,
) ([]*models.AlarmEvent, int, error) {
	status := "active"
	filters.AlarmStatus = &status
	return s.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

// GetAlarmEventsByDevice 根据设备ID获取报警事件
func (s *AlarmEventService) GetAlarmEventsByDevice(
	ctx context.Context,
	tenantID, deviceID string,
	filters repository.AlarmEventFilters,
	page, size int,
) ([]*models.AlarmEvent, int, error) {
	if deviceID == "" {
		return nil, 0, fmt.Errorf("device_id is required")
	}
	filters.DeviceID = &deviceID
	return s.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

// GetAlarmEventsByCategory 根据分类获取报警事件
func (s *AlarmEventService) GetAlarmEventsByCategory(
	ctx context.Context,
	tenantID, category string,
	filters repository.AlarmEventFilters,
	page, size int,
) ([]*models.AlarmEvent, int, error) {
	if category == "" {
		return nil, 0, fmt.Errorf("category is required")
	}
	filters.Category = &category
	return s.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

// GetAlarmEventsByLevel 根据报警级别获取报警事件
func (s *AlarmEventService) GetAlarmEventsByLevel(
	ctx context.Context,
	tenantID, alarmLevel string,
	filters repository.AlarmEventFilters,
	page, size int,
) ([]*models.AlarmEvent, int, error) {
	if alarmLevel == "" {
		return nil, 0, fmt.Errorf("alarm_level is required")
	}
	filters.AlarmLevel = &alarmLevel
	return s.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

