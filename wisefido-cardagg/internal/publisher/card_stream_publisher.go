package publisher

import (
	"context"

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
// 直接序列化 CardRealTime struct，不额外包装 data 层
func (p *CardStreamPublisher) PublishCardRealTime(ctx context.Context, tenantID, cardID string, rtData *card.CardRealTime) {
	if rtData == nil {
		p.logger.Warn("cannot publish nil CardRealTime", zap.String("card_id", cardID))
		return
	}

	streamKey := "card:realtime:stream"

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamKey, rtData, 5000, 6)
	if err != nil {
		p.logger.Error("failed to publish card realtime data",
			zap.String("stream_key", streamKey),
			zap.String("card_id", cardID),
			zap.Error(err))
		return
	}

	p.logger.Info("[PUBLISH] card realtime data",
		zap.String("stream_key", streamKey),
		zap.String("card_id", cardID),
		zap.String("stream_id", streamID))
}

// PublishCardStatus publishes card status data (BedState/RoomState/DeviceStatus/ActiveAlarms) to card:status:stream
// TTL: 12 hours (low frequency updates, long retention)
// 直接序列化 CardStatus struct，不额外包装 data 层
func (p *CardStreamPublisher) PublishCardStatus(ctx context.Context, tenantID, cardID string, statusData *card.CardStatus) {
	if statusData == nil {
		p.logger.Warn("cannot publish nil CardStatus", zap.String("card_id", cardID))
		return
	}

	streamKey := "card:status:stream"

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamKey, statusData, 2000, 43200)
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

// buildCardRealTimeMessage 和 buildCardStatusMessage 已移除
// PublishCardRealTime / PublishCardStatus 直接传 struct 给 PublishJSONToStream
// 序列化后格式与 card.CardRealTime / card.CardStatus JSON tag 一致
