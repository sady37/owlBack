package publisher

import (
	"context"
	"fmt"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigPublisher 配置变更消息发布器
// 统一管理所有 config:* 消息的发送
type ConfigPublisher struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewConfigPublisher 创建配置变更消息发布器
func NewConfigPublisher(redisClient *redis.Client, logger *zap.Logger) *ConfigPublisher {
	return &ConfigPublisher{
		redisClient: redisClient,
		logger:      logger,
	}
}

// PublishAlarmProcessMessage 发送报警处理消息到 config:alarm.process:stream
// 供 cardagg 消费用于更新告警显示
func (p *ConfigPublisher) PublishAlarmProcessMessage(
	ctx context.Context,
	tenantID, cardID, deviceID, alarmLevel, alarmType, processType string,
	alarmTimestamp int64,
) error {
	// 构建报警处理消息
	alarmProcessMsg := rediscommon.BuildAlarmProcessMessage(
		"wisefido-data",
		tenantID,
		cardID,
		deviceID,
		alarmLevel,
		alarmType,
		processType, // 如 "ack"
		alarmTimestamp,
	)

	// 发布到 config:alarm.process:stream
	streamName := rediscommon.StreamConfigAlarmProcess.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigAlarmProcess, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		alarmProcessMsg,
		maxLen,
		retentionSeconds,
	)

	if err != nil {
		p.logger.Error("Failed to publish alarm process message",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.String("device_id", deviceID),
			zap.String("alarm_level", alarmLevel),
			zap.String("process_type", processType),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish alarm process message: %w", err)
	}

	p.logger.Info("Published alarm process message",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.String("device_id", deviceID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType),
		zap.String("process_type", processType),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// PublishDeviceAlarmSettingMessage 发送设备告警配置变更消息到 config:device.alarm.setting:stream
// 供 qinglan 消费用于更新设备告警配置
func (p *ConfigPublisher) PublishDeviceAlarmSettingMessage(
	ctx context.Context,
	tenantID, deviceID, deviceUID, settingType string,
	settingData map[string]interface{},
) error {
	// 构建设备告警配置变更消息
	settingMsg := rediscommon.BuildDeviceAlarmSettingMessage(
		"wisefido-data",
		tenantID,
		deviceID,
		deviceUID,
		settingType,
		settingData,
	)

	// 发布到 config:device.alarm.setting:stream
	streamName := rediscommon.StreamConfigDeviceAlarmSetting.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigDeviceAlarmSetting, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		settingMsg,
		maxLen,
		retentionSeconds,
	)

	if err != nil {
		p.logger.Error("Failed to publish device alarm setting message",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
			zap.String("setting_type", settingType),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish device alarm setting message: %w", err)
	}

	p.logger.Info("Published device alarm setting message",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.String("setting_type", settingType),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}
