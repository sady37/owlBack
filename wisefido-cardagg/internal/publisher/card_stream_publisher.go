package publisher

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"owl-common/card"
	rediscommon "owl-common/redis"
)

type CardStreamPublisher struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

func NewCardStreamPublisher(redisClient *redis.Client, logger *zap.Logger) *CardStreamPublisher {
	return &CardStreamPublisher{
		redisClient: redisClient,
		logger:      logger,
	}
}

// PublishCardRealTime publishes card realtime data (TrackData/VitalData) to card:realtime:stream
// TTL: 6 seconds (high frequency updates)
func (p *CardStreamPublisher) PublishCardRealTime(ctx context.Context, tenantID, cardID string, rtData *card.CardRealTime) {
	if rtData == nil {
		p.logger.Warn("cannot publish nil CardRealTime", zap.String("card_id", cardID))
		return
	}

	message := p.buildCardRealTimeMessage(rtData)
	streamKey := fmt.Sprintf("iot:%s:card:realtime:stream", tenantID)

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamKey, message, 5000, 6)
	if err != nil {
		p.logger.Error("failed to publish card realtime data",
			zap.String("stream_key", streamKey),
			zap.String("card_id", cardID),
			zap.Error(err))
		return
	}

	p.logger.Debug("published card realtime data",
		zap.String("stream_key", streamKey),
		zap.String("card_id", cardID),
		zap.String("stream_id", streamID))
}

// PublishCardStatus publishes card status data (BedState/RoomState/DeviceStatus/ActiveAlarms) to card:status:stream
// TTL: 12 hours (low frequency updates, long retention)
func (p *CardStreamPublisher) PublishCardStatus(ctx context.Context, tenantID, cardID string, statusData *card.CardStatus) {
	if statusData == nil {
		p.logger.Warn("cannot publish nil CardStatus", zap.String("card_id", cardID))
		return
	}

	message := p.buildCardStatusMessage(statusData)
	streamKey := fmt.Sprintf("iot:%s:card:status:stream", tenantID)

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamKey, message, 2000, 43200)
	if err != nil {
		p.logger.Error("failed to publish card status data",
			zap.String("stream_key", streamKey),
			zap.String("card_id", cardID),
			zap.Error(err))
		return
	}

	p.logger.Debug("published card status data",
		zap.String("stream_key", streamKey),
		zap.String("card_id", cardID),
		zap.String("stream_id", streamID))
}

// buildCardRealTimeMessage constructs the message for card:realtime:stream
func (p *CardStreamPublisher) buildCardRealTimeMessage(rtData *card.CardRealTime) map[string]interface{} {
	message := map[string]interface{}{
		"card_id":   rtData.CardID,
		"timestamp": rtData.Timestamp,
		"data": map[string]interface{}{
			"track_data": rtData.TrackData,
			"vital_data": rtData.VitalData,
		},
	}
	return message
}

// buildCardStatusMessage constructs the message for card:status:stream
func (p *CardStreamPublisher) buildCardStatusMessage(statusData *card.CardStatus) map[string]interface{} {
	message := map[string]interface{}{
		"card_id":   statusData.CardID,
		"timestamp": statusData.Timestamp,
		"data": map[string]interface{}{
			"bed_state":  statusData.BedState,
			"room_state": statusData.RoomState,
		},
	}

	// Add alarm data if available
	if statusData.ActiveAlarms != nil {
		message["alarms"] = map[string]interface{}{
			"ActiveEmerg":   statusData.ActiveAlarms.ActiveEmerg,
			"ActiveAlert":   statusData.ActiveAlarms.ActiveAlert,
			"ActiveCrit":    statusData.ActiveAlarms.ActiveCrit,
			"ActiveErr":     statusData.ActiveAlarms.ActiveErr,
			"ActiveWarning": statusData.ActiveAlarms.ActiveWarning,
			"NowAlarm":      statusData.ActiveAlarms.NowAlarm,
		}
	}

	// DeviceStatus is now handled by wisefido-data directly
	// (subscribe from iot:DeviceStatus stream and push to frontend)
	// Keep DeviceStatus empty in CardStatus to avoid duplication
	// if statusData.DeviceStatus != nil && len(statusData.DeviceStatus) > 0 {
	// 	// Marshal DeviceStatus map to JSON for proper serialization
	// 	if devStatusJSON, err := json.Marshal(statusData.DeviceStatus); err == nil {
	// 		var devStatusMap map[string]interface{}
	// 		if err := json.Unmarshal(devStatusJSON, &devStatusMap); err == nil {
	// 			message["device_status"] = devStatusMap
	// 		}
	// 	}
	// }

	return message
}
