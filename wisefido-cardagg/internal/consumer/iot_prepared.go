package consumer

import (
	"context"
	"fmt"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// IotPreparedHandler 在 monitor/event/alarm 前统一校验 card_id：空则丢弃；有效 UUID 则查库占位。
type IotPreparedHandler struct {
	stateSvc  *service.StateService
	metaCache *service.DeviceMetaCache
	inner     StreamHandler
	logger    *zap.Logger
}

func NewIotPreparedHandler(stateSvc *service.StateService, metaCache *service.DeviceMetaCache, inner StreamHandler, logger *zap.Logger) StreamHandler {
	return &IotPreparedHandler{stateSvc: stateSvc, metaCache: metaCache, inner: inner, logger: logger}
}

func streamMapStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", s)
	}
}

func (h *IotPreparedHandler) Handle(ctx context.Context, msg interface{}) error {
	raw, ok := msg.(map[string]interface{})
	if !ok {
		return h.inner.Handle(ctx, msg)
	}
	cardID := streamMapStr(raw["card_id"])
	if cardID == "" {
		h.logger.Error("iot_prepared: empty card_id, dropping message",
			zap.String("device_id", streamMapStr(raw["device_id"])),
			zap.String("device_uid", streamMapStr(raw["device_uid"])),
			zap.String("topic_type", streamMapStr(raw["topic_type"])),
		)
		return nil
	}
	if service.IsUUID(cardID) {
		if meta := h.metaCache.GetOrLoad(ctx, cardID); meta != nil {
			_ = h.stateSvc.EnsureCardStatePrepared(ctx, cardID, meta)
		}
	}
	return h.inner.Handle(ctx, raw)
}
