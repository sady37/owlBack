// card_lifecycle.go — config:card:stream 消费者。
//
// card 增/改/删/重置触发：metaCache + enablement cache 失效 + card:status hash 刷新。

package consumer

import (
	"context"
	"database/sql"
	"encoding/json"

	"owl-common/card"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type cardChangeData struct {
	CardID   string `json:"card_id"`
	Op       string `json:"op"`
	TenantID string `json:"tenant_id"`
	UnitID   string `json:"unit_id"`
	BranchID string `json:"branch_id"`
}

type CardLifecycle struct {
	db         *sql.DB
	writer     *card.Writer
	metaCache  *service.DeviceMetaCache
	enablement *service.AlarmEnablementCache
	picker     *UnitPicker
	logger     *zap.Logger
}

func NewCardLifecycle(db *sql.DB, writer *card.Writer, meta *service.DeviceMetaCache, enable *service.AlarmEnablementCache, picker *UnitPicker, logger *zap.Logger) *CardLifecycle {
	return &CardLifecycle{
		db:         db,
		writer:     writer,
		metaCache:  meta,
		enablement: enable,
		picker:     picker,
		logger:     logger,
	}
}

func (h *CardLifecycle) Handle(ctx context.Context, raw map[string]interface{}) error {
	dataStr, _ := raw["data"].(string)
	if dataStr == "" {
		return nil
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		h.logger.Warn("card_change parse envelope", zap.Error(err))
		return nil
	}
	var d cardChangeData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		h.logger.Warn("card_change parse data", zap.Error(err))
		return nil
	}
	if d.CardID == "" && d.Op != "reset" {
		return nil
	}

	switch d.Op {
	case "reset":
		h.metaCache.InvalidateAll()
		if err := h.metaCache.BuildDeviceIndex(ctx); err != nil {
			h.logger.Warn("rebuild device index", zap.Error(err))
		}
		h.enablement.InvalidateAll()
		if h.picker != nil {
			h.picker.InvalidateAll()
		}
	case "deleted", "delete":
		h.metaCache.Remove(d.CardID)
		h.metaCache.RefreshDeviceIndexForCard(ctx, d.CardID)
		if err := h.writer.DeleteCardState(ctx, d.CardID); err != nil {
			h.logger.Warn("delete card state", zap.String("cid", d.CardID), zap.Error(err))
		}
	default:
		h.metaCache.Remove(d.CardID)
		h.metaCache.RefreshDeviceIndexForCard(ctx, d.CardID)
	}

	if d.UnitID != "" {
		h.metaCache.InvalidateCardsInTenantUnit(ctx, d.UnitID)
		addrs := service.DeviceAddrsInUnit(ctx, h.db, d.UnitID)
		h.enablement.InvalidateDevices(addrs)
		if h.picker != nil {
			h.picker.InvalidateUnit(d.UnitID)
		}
	}

	// card 改/增 → 刷新 AlarmState（实时聚合 alarm_events 写 card:status）
	if d.Op != "deleted" && d.Op != "delete" && d.CardID != "" {
		cas, err := card.QueryCardAlarmState(ctx, h.db, d.CardID)
		if err != nil {
			h.logger.Warn("query alarm state", zap.String("cid", d.CardID), zap.Error(err))
			return nil
		}
		if cas == nil {
			return nil
		}
		return h.writer.WriteCardStatus(ctx, &card.CardStatus{
			CardID:     d.CardID,
			AlarmState: cas.ToAlarmState(),
		})
	}
	return nil
}
