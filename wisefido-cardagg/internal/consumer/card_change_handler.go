package consumer

import (
	"context"
	"encoding/json"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type CardChangeHandler struct {
	alarms       *service.AlarmService
	stateService *service.StateService
	metaCache    *service.DeviceMetaCache
	enablement   *service.AlarmEnablementCache
	resolver     *service.DeviceCardResolver
	logger       *zap.Logger
}

func NewCardChangeHandler(alarms *service.AlarmService, stateService *service.StateService, metaCache *service.DeviceMetaCache, enablement *service.AlarmEnablementCache, resolver *service.DeviceCardResolver, logger *zap.Logger) *CardChangeHandler {
	return &CardChangeHandler{alarms: alarms, stateService: stateService, metaCache: metaCache, enablement: enablement, resolver: resolver, logger: logger}
}

type cardChangeData struct {
	CardID string `json:"card_id"`
	Op     string `json:"op"`
}

func (h *CardChangeHandler) Handle(ctx context.Context, msg interface{}) error {
	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		return nil
	}

	dataStr, ok := streamMsg["data"].(string)
	if !ok {
		return nil
	}

	var env cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		h.logger.Warn("parse cloud events", zap.Error(err))
		return nil
	}

	var d cardChangeData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		h.logger.Warn("parse card change data", zap.Error(err))
		return nil
	}

	if d.CardID == "" {
		return nil
	}

	if d.Op == "deleted" || d.Op == "delete" {
		h.metaCache.Remove(d.CardID)
	} else {
		h.metaCache.Invalidate(d.CardID)
	}
	if h.resolver != nil {
		h.resolver.InvalidateAll()
	}

	// Card 初始化/更新：默认创建 RoomState，存在卫生间雷达时创建 BathRoomState
	if d.Op != "deleted" && d.Op != "delete" && h.stateService != nil {
		meta := h.metaCache.GetOrLoad(ctx, d.CardID)
		if meta != nil && len(meta.Devices) > 0 {
			_ = h.stateService.InitCardRoomAndBathroomState(ctx, d.CardID, meta, h.enablement)
		}
	}

	return h.alarms.HandleCardChange(ctx, d.CardID, d.Op)
}
