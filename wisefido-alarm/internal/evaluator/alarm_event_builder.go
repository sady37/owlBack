package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-alarm/internal/models"

	"github.com/google/uuid"
)

// AlarmEventBuilder 报警事件构建器
type AlarmEventBuilder struct {
	tenantID string
	deviceID string
}

// NewAlarmEventBuilder 创建报警事件构建器
func NewAlarmEventBuilder(tenantID, deviceID string) *AlarmEventBuilder {
	return &AlarmEventBuilder{
		tenantID: tenantID,
		deviceID: deviceID,
	}
}

// BuildAlarmEvent 构建报警事件
func (b *AlarmEventBuilder) BuildAlarmEvent(
	eventType string,
	category string,
	alarmLevel string,
	triggerData *models.TriggerData,
	metadata map[string]interface{},
) (*models.AlarmEvent, error) {
	now := time.Now()

	// 序列化 trigger_data
	triggerDataJSON, err := json.Marshal(triggerData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trigger data: %w", err)
	}

	// 序列化 metadata
	metadataJSON := "{}"
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	// 序列化 notified_users（默认空数组）
	notifiedUsersJSON := "[]"

	event := &models.AlarmEvent{
		EventID:         uuid.New().String(),
		TenantID:        b.tenantID,
		DeviceID:        b.deviceID,
		EventType:       eventType,
		Category:        category,
		AlarmLevel:      alarmLevel,
		AlarmStatus:     "active",
		TriggeredAt:     now,
		TriggerData:     string(triggerDataJSON),
		NotifiedUsers:   notifiedUsersJSON,
		Metadata:        metadataJSON,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return event, nil
}

// BuildTriggerData 构建触发数据
func BuildTriggerData(
	eventType string,
	source string,
	heartRate *int,
	respiratoryRate *int,
	posture *string,
	postureDisplay *string,
	snomedCode *string,
	snomedDisplay *string,
	confidence *int,
	durationSec *int,
) *models.TriggerData {
	return &models.TriggerData{
		EventType:        eventType,
		Source:           source,
		HeartRate:        heartRate,
		RespiratoryRate:  respiratoryRate,
		Posture:          posture,
		PostureDisplay:   postureDisplay,
		SNOMEDCode:       snomedCode,
		SNOMEDDisplay:    snomedDisplay,
		Confidence:       confidence,
		DurationSec:      durationSec,
	}
}

// CheckDuplicate 检查是否重复报警（在 Evaluator 中使用）
func (e *Evaluator) CheckDuplicate(
	ctx context.Context,
	tenantID, deviceID, eventType string,
	withinMinutes int,
) (bool, error) {
	recentEvent, err := e.alarmEventsRepo.GetRecentAlarmEvent(ctx, tenantID, deviceID, eventType, withinMinutes)
	if err != nil {
		return false, err
	}

	return recentEvent != nil, nil
}

