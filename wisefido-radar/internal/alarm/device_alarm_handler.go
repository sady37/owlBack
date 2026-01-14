package alarm

import (
	"context"
	"wisefido-radar/internal/repository"

	"go.uber.org/zap"
)

// DeviceAlarmHandler 设备报警处理器
// 负责判断事件是否应该发布为报警，以及检查报警是否启用
type DeviceAlarmHandler struct {
	deviceRepo *repository.DeviceRepository
	logger     *zap.Logger
}

// NewDeviceAlarmHandler 创建设备报警处理器
func NewDeviceAlarmHandler(
	deviceRepo *repository.DeviceRepository,
	logger *zap.Logger,
) *DeviceAlarmHandler {
	return &DeviceAlarmHandler{
		deviceRepo: deviceRepo,
		logger:     logger,
	}
}

// ShouldPublishAsAlarm 判断是否应该将消息发布为报警
// 参数：
//   - ctx: 上下文
//   - tenantID: 租户ID
//   - deviceID: 设备ID
//   - topicType: 消息类型（"event" 或 "statistics"）
//   - dataValue: 数据值（从 RadarDecoder 返回）
// 返回：
//   - shouldPublish: 是否应该发布为报警
//   - possibleAlarmTypes: 可能触发的报警类型列表（用于日志）
func (h *DeviceAlarmHandler) ShouldPublishAsAlarm(
	ctx context.Context,
	tenantID, deviceID, topicType string,
	dataValue interface{},
) (shouldPublish bool, possibleAlarmTypes []string, err error) {
	// 1. 根据 topicType 和 dataValue 确定可能触发的报警类型
	switch topicType {
	case "event":
		possibleAlarmTypes = repository.GetPossibleAlarmTypesFromEvent(dataValue)
	case "statistics":
		possibleAlarmTypes = repository.GetPossibleAlarmTypesFromStat(dataValue)
	default:
		// monitor 或其他类型，不发布为报警
		return false, nil, nil
	}

	// 2. 如果没有可能的报警类型，不发布为报警
	if len(possibleAlarmTypes) == 0 {
		return false, nil, nil
	}

	// 3. 获取设备的报警使能配置
	enablement, err := h.deviceRepo.GetAlarmEnablement(ctx, tenantID, deviceID)
	if err != nil {
		h.logger.Warn("Failed to get alarm enablement",
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		// 查询失败，不发布为报警（避免误报）
		return false, possibleAlarmTypes, err
	}

	// 4. 检查是否有任何可能的报警类型已启用
	for _, alarmType := range possibleAlarmTypes {
		if enablement[alarmType] {
			// 至少有一个报警类型已启用，应该发布为报警
			h.logger.Debug("Alarm type is enabled, should publish as alarm",
				zap.String("device_id", deviceID),
				zap.String("alarm_type", alarmType),
				zap.String("topic_type", topicType),
			)
			return true, possibleAlarmTypes, nil
		}
	}

	// 5. 所有可能的报警类型都未启用，不发布为报警
	h.logger.Debug("All possible alarm types are disabled, publish as normal event/stat",
		zap.String("device_id", deviceID),
		zap.Strings("possible_alarm_types", possibleAlarmTypes),
		zap.String("topic_type", topicType),
	)
	return false, possibleAlarmTypes, nil
}
