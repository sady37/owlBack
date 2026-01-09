package alarm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-card-aggregator/internal/models"
	"wisefido-card-aggregator/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AlarmHandler 报警处理器
type AlarmHandler struct {
	alarmEventsRepo *repository.AlarmEventsRepository
	alarmDeviceRepo *repository.AlarmDeviceRepository
	cardRepo        *repository.CardRepository
	logger          *zap.Logger
}

// NewAlarmHandler 创建报警处理器
func NewAlarmHandler(
	alarmEventsRepo *repository.AlarmEventsRepository,
	alarmDeviceRepo *repository.AlarmDeviceRepository,
	cardRepo *repository.CardRepository,
	logger *zap.Logger,
) *AlarmHandler {
	return &AlarmHandler{
		alarmEventsRepo: alarmEventsRepo,
		alarmDeviceRepo: alarmDeviceRepo,
		cardRepo:        cardRepo,
		logger:          logger,
	}
}

// IsDeviceDirectAlarm 判断是否是设备直接报警类型
// 设备直接报警：设备直接上报的报警事件（如 Fall, OfflineAlarm 等）
// 云端事件报警：由云端规则评估产生的报警（如事件1-4），不在本函数中处理
func IsDeviceDirectAlarm(eventType string) bool {
	// 设备直接报警类型列表
	deviceDirectAlarms := map[string]bool{
		// 安全报警（Radar 设备直接上报）
		"Fall":          true,
		"SuspectedFall": true,
		"Stay":          true,
		"NoActivity24h": true,

		// 设备状态报警（所有设备）
		"OfflineAlarm":   true,
		"LowBattery":     true,
		"DeviceFailure":  true,
		"AngleException": true,

		// 行为报警（设备直接上报）
		"LeftBed": true,
		"SitUp":   true,

		// 生理报警（设备直接上报，注意：云端计算的报警不在列表中）
		// 例如：Radar_AbnormalHeartRate 是云端计算的，不在列表中
		// 但如果有设备直接上报的心率异常，可以在这里添加
	}

	return deviceDirectAlarms[eventType]
}

// GetAlarmCategory 获取报警分类（FHIR Category）
func GetAlarmCategory(eventType string) string {
	// 安全报警
	safetyAlarms := map[string]bool{
		"Fall":          true,
		"SuspectedFall": true,
		"Stay":          true,
		"NoActivity24h": true,
	}

	// 临床报警
	clinicalAlarms := map[string]bool{
		"AbnormalHeartRate":        true,
		"Radar_AbnormalHeartRate": true,
		"AbnormalRespiratoryRate":  true,
		"Radar_AbnormalRespiratoryRate": true,
		"ApneaHypopnea":            true,
		"Radar_ApneaHypopnea":      true,
	}

	// 行为报警
	behavioralAlarms := map[string]bool{
		"LeftBed":           true,
		"SleepPad_LeftBed":  true,
		"Radar_LeftBed":     true,
		"SitUp":             true,
		"SleepPad_SitUp":    true,
		"NoTurning.2H":      true,
		"NoBodyMovement.2H": true,
	}

	// 设备报警
	deviceAlarms := map[string]bool{
		"OfflineAlarm":  true,
		"LowBattery":    true,
		"DeviceFailure": true,
		"AngleException": true,
	}

	if safetyAlarms[eventType] {
		return "safety"
	}
	if clinicalAlarms[eventType] {
		return "clinical"
	}
	if behavioralAlarms[eventType] {
		return "behavioral"
	}
	if deviceAlarms[eventType] {
		return "device"
	}

	// 默认返回 safety
	return "safety"
}

// GetAlarmLevel 获取报警级别
func GetAlarmLevel(eventType string) string {
	// EMERGENCY 级别（最高）
	emergencyAlarms := map[string]bool{
		"Fall":          true,
		"NoActivity24h":  true,
		"DeviceFailure": true,
	}

	// WARNING 级别
	warningAlarms := map[string]bool{
		"SuspectedFall": true,
		"Stay":          true,
		"OfflineAlarm":  true,
		"LowBattery":    true,
		"AngleException": true,
		"LeftBed":       true,
		"SitUp":         true,
	}

	if emergencyAlarms[eventType] {
		return "EMERGENCY"
	}
	if warningAlarms[eventType] {
		return "WARNING"
	}

	// 默认返回 WARNING
	return "WARNING"
}

// IsAlarmEnabled 判断报警是否在设备配置中启用
func (h *AlarmHandler) IsAlarmEnabled(ctx context.Context, tenantID, deviceID, eventType string) (bool, error) {
	// 获取设备报警配置
	config, err := h.alarmDeviceRepo.GetAlarmDeviceConfig(ctx, tenantID, deviceID)
	if err != nil {
		return false, fmt.Errorf("failed to get alarm device config: %w", err)
	}

	// 如果没有配置，默认启用（允许创建报警）
	if config == nil {
		h.logger.Debug("No alarm config found for device, defaulting to enabled",
			zap.String("device_id", deviceID),
			zap.String("event_type", eventType),
		)
		return true, nil
	}

	// 解析 monitor_config JSONB
	var monitorConfig map[string]interface{}
	if err := json.Unmarshal(config.MonitorConfig, &monitorConfig); err != nil {
		h.logger.Warn("Failed to parse monitor_config, defaulting to enabled",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		return true, nil // 解析失败，默认启用
	}

	// 检查 alarms 配置
	alarms, ok := monitorConfig["alarms"].(map[string]interface{})
	if !ok {
		// 没有 alarms 配置，默认启用
		return true, nil
	}

	// 检查该报警类型是否启用
	alarmConfig, ok := alarms[eventType].(map[string]interface{})
	if !ok {
		// 没有该报警类型的配置，默认启用
		return true, nil
	}

	// 检查 enabled 字段
	enabled, ok := alarmConfig["enabled"].(bool)
	if !ok {
		// enabled 字段不存在或类型错误，默认启用
		return true, nil
	}

	return enabled, nil
}

// BuildDeviceAlarmEvent 构建设备直接报警事件
func (h *AlarmHandler) BuildDeviceAlarmEvent(
	tenantID string,
	deviceID string,
	eventType string,
	iotTimeSeriesID *int64,
	triggerData *models.TriggerData,
) (*models.AlarmEvent, error) {
	now := time.Now()

	// 序列化 trigger_data
	var triggerDataJSON json.RawMessage = json.RawMessage("{}")
	if triggerData != nil {
		triggerDataBytes, err := json.Marshal(triggerData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal trigger data: %w", err)
		}
		triggerDataJSON = triggerDataBytes
	}

	// 构建 metadata
	metadata := map[string]interface{}{
		"trigger_source": "device", // 标记为设备直接报警
		"created_by":     "wisefido-card-aggregator",
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// 序列化 notified_users（默认空数组）
	notifiedUsersJSON := json.RawMessage("[]")

	event := &models.AlarmEvent{
		EventID:         uuid.New().String(),
		TenantID:        tenantID,
		DeviceID:        deviceID,
		EventType:       eventType,
		Category:        GetAlarmCategory(eventType),
		AlarmLevel:      GetAlarmLevel(eventType),
		AlarmStatus:     "active",
		TriggeredAt:     now,
		IoTTimeSeriesID: iotTimeSeriesID,
		TriggerData:     triggerDataJSON,
		NotifiedUsers:   notifiedUsersJSON,
		Metadata:        metadataJSON,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return event, nil
}

// CreateDeviceAlarm 创建设备直接报警事件
func (h *AlarmHandler) CreateDeviceAlarm(
	ctx context.Context,
	tenantID string,
	deviceID string,
	eventType string,
	iotTimeSeriesID *int64,
	triggerData *models.TriggerData,
) error {
	// 1. 检查是否是设备直接报警
	if !IsDeviceDirectAlarm(eventType) {
		h.logger.Debug("Event type is not a device direct alarm, skipping",
			zap.String("event_type", eventType),
		)
		return nil // 不是设备直接报警，不处理
	}

	// 2. 检查报警是否启用
	enabled, err := h.IsAlarmEnabled(ctx, tenantID, deviceID, eventType)
	if err != nil {
		h.logger.Warn("Failed to check if alarm is enabled, skipping",
			zap.String("device_id", deviceID),
			zap.String("event_type", eventType),
			zap.Error(err),
		)
		return nil // 检查失败，不创建报警（避免误报）
	}

	if !enabled {
		h.logger.Debug("Alarm is disabled in device config, skipping",
			zap.String("device_id", deviceID),
			zap.String("event_type", eventType),
		)
		return nil // 报警未启用，不创建
	}

	// 3. 构建报警事件
	alarmEvent, err := h.BuildDeviceAlarmEvent(tenantID, deviceID, eventType, iotTimeSeriesID, triggerData)
	if err != nil {
		return fmt.Errorf("failed to build alarm event: %w", err)
	}

	// 4. 创建报警事件
	if err := h.alarmEventsRepo.CreateAlarmEvent(ctx, tenantID, alarmEvent); err != nil {
		return fmt.Errorf("failed to create alarm event: %w", err)
	}

	h.logger.Info("Device direct alarm created",
		zap.String("event_id", alarmEvent.EventID),
		zap.String("event_type", eventType),
		zap.String("device_id", deviceID),
		zap.String("alarm_level", alarmEvent.AlarmLevel),
	)

	// 5. 更新相关卡片的报警计数
	// 通过 device_id 找到关联的卡片
	cardInfo, err := h.cardRepo.GetCardByDeviceID(tenantID, deviceID)
	if err != nil {
		// 如果找不到卡片，记录警告但不返回错误（设备可能未绑定到卡片）
		h.logger.Warn("Failed to get card for device, skipping alarm count update",
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return nil
	}

	// 更新卡片的报警计数
	if err := h.cardRepo.UpdateCardAlarmCounts(ctx, tenantID, cardInfo.CardID); err != nil {
		h.logger.Warn("Failed to update card alarm counts",
			zap.String("card_id", cardInfo.CardID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		// 不返回错误，报警已创建成功
	} else {
		h.logger.Debug("Updated card alarm counts",
			zap.String("card_id", cardInfo.CardID),
			zap.String("device_id", deviceID),
		)
	}

	return nil
}

