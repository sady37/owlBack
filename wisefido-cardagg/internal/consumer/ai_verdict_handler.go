// ai_verdict_handler.go — cardagg 侧 ai:track:verdict:stream 消费者（S5a port from v1）。
//
// sensor RoomEngine.PublishAIEvent(category="track_verdict") 切到专用流
// `ai:track:verdict:stream`（[[sensor_stream_subscriptions]] §"专用流"；CLAUDE.md 规则 #2.1
// AI 派生判定不入业务库）。本 handler 把流上消息喂回 cardagg aiOverrides cache，
// 让 monitor_handler.Apply 合并到 realtime publish。

package consumer

import (
	"context"
	"time"

	rediscommon "owl-common/redis"

	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	aiVerdictGroup    = "$cons-cardagg-ai-verdict"
	aiVerdictConsumer = "consumer-ai-verdict"
	aiVerdictRead     = 10
	aiVerdictBlock    = 5 * time.Second
)

type AIVerdictHandler struct {
	client      *redislib.Client
	aiOverrides *service.AIOverrideCache
	logger      *zap.Logger
}

func NewAIVerdictHandler(client *redislib.Client, aiOverrides *service.AIOverrideCache, logger *zap.Logger) *AIVerdictHandler {
	return &AIVerdictHandler{client: client, aiOverrides: aiOverrides, logger: logger}
}

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
	if !msg.DeviceAddr.IsValid() {
		return
	}
	// 同流上还有 sensor_decision（lost-fall 决策审计，iot 落 sensor_decision_log）——
	// 那类无 realtime override 语义、且常无 track_confidence，会把 cache 置 0。按 category 跳过。
	if msg.Category != "track_verdict" {
		return
	}
	deviceAddr := msg.DeviceAddr.String()
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
