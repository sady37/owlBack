package consumer

import (
	"context"
	"time"

	rediscommon "owl-common/redis"

	"wisefido-sleepace/internal/config"
	"wisefido-sleepace/internal/service"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type StreamPublisher struct {
	redisClient    *redis.Client
	config         *config.Config
	cardMappingSvc *service.CardMappingService
	logger         *zap.Logger
}

func NewStreamPublisher(redisClient *redis.Client, cfg *config.Config, logger *zap.Logger) *StreamPublisher {
	return &StreamPublisher{
		redisClient: redisClient,
		config:      cfg,
		logger:      logger,
	}
}

func (p *StreamPublisher) SetCardMappingService(svc *service.CardMappingService) {
	p.cardMappingSvc = svc
}

// ResolveToDeviceUID 将 device_code 或 device_id 转为 device_uid，供网关内部统一使用。
func (p *StreamPublisher) ResolveToDeviceUID(ctx context.Context, id string) string {
	if p.cardMappingSvc == nil || id == "" {
		return id
	}
	return p.cardMappingSvc.ResolveToDeviceUID(ctx, id)
}

// Resolve 用 device key 查 device_store+cards，返回完整身份（与 DeviceCardMapping 一致）。
// 入参可为 device_uid 或 device_code（MQTT 首次可能发 uid，后续可能发 code），GetCardInfo/LookupCard 内部统一解析；未命中时 deviceID/deviceCode/deviceType 为空。
func (p *StreamPublisher) Resolve(ctx context.Context, deviceKey string) (tenantID, branchID, unitID, cardID, deviceID, outUID, deviceCode, deviceType string) {
	if p.cardMappingSvc == nil {
		return "", "", "", "", "", deviceKey, "", ""
	}
	info, err := p.cardMappingSvc.GetCardInfo(ctx, deviceKey)
	if err != nil {
		p.logger.Debug("card lookup miss", zap.String("device_key", deviceKey), zap.Error(err))
		return "", "", "", "", "", deviceKey, "", ""
	}
	return info.TenantID, info.BranchID, info.UnitID, info.CardID, info.DeviceID, info.DeviceUID, info.DeviceCode, info.DeviceType
}

// PublishMonitor sends an IoTStreamMessage to iot:monitor:stream.
func (p *StreamPublisher) PublishMonitor(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publish(ctx, rediscommon.StreamMonitor, msg)
}

// PublishEvent sends an IoTStreamMessage to iot:event:stream.
func (p *StreamPublisher) PublishEvent(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publish(ctx, rediscommon.StreamEvent, msg)
}

// PublishAlarm sends an IoTStreamMessage to iot:alarm:stream.
func (p *StreamPublisher) PublishAlarm(ctx context.Context, msg *rediscommon.IoTStreamMessage) error {
	return p.publish(ctx, rediscommon.StreamAlarm, msg)
}

func (p *StreamPublisher) publish(ctx context.Context, stream rediscommon.StreamDefinition, msg *rediscommon.IoTStreamMessage) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	data := msg.ToStreamMap()
	maxLen, retention := p.config.GetStreamConfig(stream.Name)
	_, err := rediscommon.PublishToStream(ctx, p.redisClient, stream.Name, data, maxLen, retention)
	if err != nil {
		p.logger.Error("publish failed",
			zap.String("stream", stream.Name),
			zap.String("device_uid", msg.DeviceUID),
			zap.Error(err))
	}
	return err
}
