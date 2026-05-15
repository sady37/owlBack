package consumer

// PR1 (A10): cardagg 侧 ai:track:verdict:stream 消费者。
//
// sensor 端 RoomEngine.PublishAIEvent(category="track_verdict") 已切到 ai:track:verdict:stream
// （PR1 同期改动，见 wisefido-sensor/internal/roomengine/engine.go）。本 handler 把流上消息
// 喂回 cardagg 现有 aiOverrides cache，使 monitor_handler.Apply 合并逻辑完全不变 → 前端契约不变。
//
// 切走原因：cardagg 准备退订 iot:event:stream（B 组迁移），同时保留 aiOverrides cache
// 作为 sensor verdict 与 cardagg realtime publish 之间的合并点。

import (
	"context"
	"time"

	rediscommon "owl-common/redis"
	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	aiVerdictGroup    = "$cons-cardagg"
	aiVerdictConsumer = "consumer-ai-verdict"
	aiVerdictRead     = 10
	aiVerdictBlock    = 5 * time.Second
)

// AIVerdictHandler 订阅 ai:track:verdict:stream 喂 aiOverrides。
type AIVerdictHandler struct {
	client      *redislib.Client
	aiOverrides *service.AIOverrideCache
	logger      *zap.Logger
}

func NewAIVerdictHandler(client *redislib.Client, aiOverrides *service.AIOverrideCache, logger *zap.Logger) *AIVerdictHandler {
	return &AIVerdictHandler{client: client, aiOverrides: aiOverrides, logger: logger}
}

// Start goroutine 跑读流循环。group 与 cardagg 其他订阅一致（$cons-cardagg），
// stream 不同故 offset 互不冲突。
func (h *AIVerdictHandler) Start(ctx context.Context) {
	if h == nil || h.client == nil || h.aiOverrides == nil {
		return
	}
	if err := rediscommon.CreateConsumerGroup(ctx, h.client, rediscommon.StreamAITrackVerdict.Name, aiVerdictGroup); err != nil {
		h.logger.Warn("ai verdict: create consumer group", zap.Error(err))
	}
	go h.runLoop(ctx)
	h.logger.Info("ai verdict consumer started",
		zap.String("stream", rediscommon.StreamAITrackVerdict.Name),
		zap.String("group", aiVerdictGroup))
}

func (h *AIVerdictHandler) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msgs, err := rediscommon.ReadFromStreamWithBlock(ctx, h.client,
			rediscommon.StreamAITrackVerdict.Name, aiVerdictGroup, aiVerdictConsumer,
			aiVerdictRead, aiVerdictBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			h.logger.Warn("ai verdict: read stream", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			h.handleRaw(m.Values)
			h.client.XAck(ctx, rediscommon.StreamAITrackVerdict.Name, aiVerdictGroup, m.ID)
		}
	}
}

func (h *AIVerdictHandler) handleRaw(raw map[string]interface{}) {
	msg, err := rediscommon.FromStreamMap(raw)
	if err != nil {
		h.logger.Warn("ai verdict: parse", zap.Error(err))
		return
	}
	deviceAddr := ""
	if msg.DeviceAddr.IsValid() {
		deviceAddr = msg.DeviceAddr.String()
	}
	if deviceAddr == "" {
		return
	}
	data := rediscommon.FirstDataValue(msg.DataValue)
	if data == nil {
		return
	}
	tid := intFromAny(data["track_id"])
	if tid <= 0 {
		return
	}
	conf := intFromAny(data["track_confidence"])
	source, _ := data["source"].(string)
	reason, _ := data["reason"].(string)
	h.aiOverrides.Set(deviceAddr, tid, service.AIVerdict{
		Confidence: conf,
		Source:     source,
		Reason:     reason,
		UpdatedMs:  msg.Timestamp,
	})
}
