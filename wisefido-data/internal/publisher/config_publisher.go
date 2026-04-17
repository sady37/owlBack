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

// PublishCardChangeForDevice 发送 config.card，data 含 device_id、change_type；可选 deviceUIDs 写入 affected_device_uids 供网关精确失效。
func (p *ConfigPublisher) PublishCardChangeForDevice(ctx context.Context, tenantID, deviceID, changeType string, deviceUIDs ...string) error {
	if p == nil || deviceID == "" {
		return nil
	}
	extra := map[string]interface{}{
		"device_id":   deviceID,
		"change_type": changeType,
	}
	if u := compactDeviceUIDs(deviceUIDs); len(u) > 0 {
		extra["affected_device_uids"] = u
	}
	return p.PublishCardChangeMessageWithExtraAndType(ctx, tenantID, "", "", "", rediscommon.ConfigCardChanged, extra)
}

func compactDeviceUIDs(uids []string) []string {
	if len(uids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(uids))
	out := make([]string, 0, len(uids))
	for _, s := range uids {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// PublishConfigCardReset 发送 reset 通知到 config:card:stream。
func (p *ConfigPublisher) PublishConfigCardReset(ctx context.Context) error {
	resetMsg := rediscommon.BuildCardChangeMessageWithExtraAndType(
		"wisefido-data", "", "", "", "", rediscommon.ConfigCardChanged,
		map[string]interface{}{"op": "reset"},
	)
	streamName := rediscommon.StreamConfigCard.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigCard, nil)
	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamName, resetMsg, int64(maxLen), retentionSeconds)
	if err != nil {
		p.logger.Error("Failed to publish configCard reset", zap.String("stream", streamName), zap.Error(err))
		return fmt.Errorf("failed to publish configCard reset: %w", err)
	}
	p.logger.Info("Published configCard reset", zap.String("stream", streamName), zap.String("stream_id", streamID))
	return nil
}

// PublishCardChangeMessageWithExtraAndType 发送卡片变更消息到 config:card:stream（支持额外字段和自定义 type 字段）
// messageType 一般为 ConfigCardChanged；extraData 可含：
//   - device_id / change_type
//   - affected_device_uids: []string，网关按 UID 逐项失效 baseline
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
