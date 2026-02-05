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

// PublishCardRealtime publishes aggregated card data to iot:card:stream
func (p *CardStreamPublisher) PublishCardRealtime(ctx context.Context, tenantID, cardID string, realtimeData *card.RealtimeData) {
	if realtimeData == nil {
		p.logger.Warn("cannot publish nil realtimeData", zap.String("card_id", cardID))
		return
	}

	message := p.buildCardStreamMessage(realtimeData)
	streamKey := fmt.Sprintf("iot:%s:stream", tenantID)

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamKey, message, 1000, 0)
	if err != nil {
		p.logger.Error("failed to publish card data to stream",
			zap.String("stream_key", streamKey),
			zap.String("card_id", cardID),
			zap.Error(err))
		return
	}

	p.logger.Debug("published card data to stream",
		zap.String("stream_key", streamKey),
		zap.String("card_id", cardID),
		zap.String("stream_id", streamID))
}

// buildCardStreamMessage constructs the message for iot:card:stream
func (p *CardStreamPublisher) buildCardStreamMessage(realtimeData *card.RealtimeData) map[string]interface{} {
	message := map[string]interface{}{
		"topic_type": "card",
		"card_id":    realtimeData.CardID,
		"timestamp":  realtimeData.Timestamp,
		"data": map[string]interface{}{
			"bed_state":  realtimeData.BedState,
			"room_state": realtimeData.RoomState,
			"vital":      realtimeData.Vital,
			"postures":   realtimeData.Postures,
		},
	}

	// Add alarm data if available
	if realtimeData.ActiveAlarms != nil {
		message["alarms"] = map[string]interface{}{
			"EMERG":     realtimeData.ActiveAlarms.EMERG,
			"ALERT":     realtimeData.ActiveAlarms.ALERT,
			"CRIT":      realtimeData.ActiveAlarms.CRIT,
			"ERR":       realtimeData.ActiveAlarms.ERR,
			"WARNING":   realtimeData.ActiveAlarms.WARNING,
			"NOTICE":    realtimeData.ActiveAlarms.NOTICE,
			"now_alarm": realtimeData.ActiveAlarms.NowAlarm,
		}
	}

	return message
}
