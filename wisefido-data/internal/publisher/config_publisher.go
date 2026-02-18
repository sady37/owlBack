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

// PublishAlarmProcessMessage 发送报警处理消息到 config:alarmProcess:stream
// 供 cardagg 消费用于更新告警显示
func (p *ConfigPublisher) PublishAlarmProcessMessage(
	ctx context.Context,
	tenantID, cardID, deviceID, alarmLevel, alarmType, processType, eventID string,
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
		eventID,
		alarmTimestamp,
	)

	// 发布到 config:alarmProcess:stream
	streamName := rediscommon.StreamConfigAlarmProcess.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigAlarmProcess, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		alarmProcessMsg,
		int64(maxLen),
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

// PublishAlarmDeviceMessage 发送设备告警配置变更消息到 config:alarmDevice:stream
// 供 qinglan 消费用于更新设备告警配置
func (p *ConfigPublisher) PublishAlarmDeviceMessage(
	ctx context.Context,
	tenantID, deviceID, deviceUID, settingType string,
	settingData map[string]interface{},
) error {
	// 构建设备告警配置变更消息
	settingMsg := rediscommon.BuildAlarmDeviceMessage(
		"wisefido-data",
		tenantID,
		deviceID,
		deviceUID,
		settingType,
		settingData,
	)

	// 发布到 config:alarmDevice:stream
	streamName := rediscommon.StreamConfigAlarmDevice.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigAlarmDevice, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		settingMsg,
		int64(maxLen),
		retentionSeconds,
	)

	if err != nil {
		p.logger.Error("Failed to publish alarm device message",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID),
			zap.String("setting_type", settingType),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish alarm device message: %w", err)
	}

	p.logger.Info("Published alarm device message",
		zap.String("tenant_id", tenantID),
		zap.String("device_id", deviceID),
		zap.String("device_uid", deviceUID),
		zap.String("setting_type", settingType),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// PublishCardChangeMessage 发送卡片变更消息到 config:card:stream
// 供 qinglan 和其他服务消费用于更新卡片相关配置
func (p *ConfigPublisher) PublishCardChangeMessage(
	ctx context.Context,
	tenantID, cardID, unitID, branchID string,
) error {
	return p.PublishCardChangeMessageWithExtra(ctx, tenantID, cardID, unitID, branchID, nil)
}

// PublishCardChangeMessageWithExtra 发送卡片变更消息到 config:card:stream（支持额外字段）
// 用于处理 card 变化导致的卡片变更（消息类型为 config.card）
// extraData 可以包含额外字段
func (p *ConfigPublisher) PublishCardChangeMessageWithExtra(
	ctx context.Context,
	tenantID, cardID, unitID, branchID string,
	extraData map[string]interface{},
) error {
	return p.PublishCardChangeMessageWithExtraAndType(ctx, tenantID, cardID, unitID, branchID, rediscommon.ConfigCardChanged, extraData)
}

// PublishCardChangeMessageWithExtraAndType 发送卡片变更消息到 config:card:stream（支持额外字段和自定义消息类型）
// 用于处理 device_store 变化导致的卡片变更或其他配置变化
// messageType: 消息类型，如 ConfigCardChanged（"config.card"）或 ConfigCardDeviceStoreChanged（"config.card.device_store"）
// 当处理 device_store 变化时，cardID 应为空，extraData 可包含：
//   - device_id: 受影响的设备ID
//   - change_type: 变化类型 (device_updated 或 device_deleted)
func (p *ConfigPublisher) PublishCardChangeMessageWithExtraAndType(
	ctx context.Context,
	tenantID, cardID, unitID, branchID, messageType string,
	extraData map[string]interface{},
) error {
	// 构建卡片变更消息（使用 BuildCardChangeMessageWithExtraAndType）
	cardMsg := rediscommon.BuildCardChangeMessageWithExtraAndType(
		"wisefido-data",
		tenantID,
		cardID,
		unitID,
		branchID,
		messageType,
		extraData,
	)

	// 发布到 config:card:stream
	streamName := rediscommon.StreamConfigCard.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigCard, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		cardMsg,
		int64(maxLen),
		retentionSeconds,
	)

	if err != nil {
		p.logger.Error("Failed to publish card change message",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.String("unit_id", unitID),
			zap.String("branch_id", branchID),
			zap.String("message_type", messageType),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish card change message: %w", err)
	}

	p.logger.Info("Published card change message",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.String("unit_id", unitID),
		zap.String("branch_id", branchID),
		zap.String("message_type", messageType),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}
