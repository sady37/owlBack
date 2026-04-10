package consumer

import (
	"context"
	"database/sql"
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
	bedCoord     *service.BedEventCoordinator
	db           *sql.DB
	logger       *zap.Logger
}

func NewCardChangeHandler(alarms *service.AlarmService, stateService *service.StateService, metaCache *service.DeviceMetaCache, enablement *service.AlarmEnablementCache, resolver *service.DeviceCardResolver, bedCoord *service.BedEventCoordinator, db *sql.DB, logger *zap.Logger) *CardChangeHandler {
	return &CardChangeHandler{alarms: alarms, stateService: stateService, metaCache: metaCache, enablement: enablement, resolver: resolver, bedCoord: bedCoord, db: db, logger: logger}
}

type cardChangeData struct {
	CardID   string `json:"card_id"`
	Op       string `json:"op"`
	TenantID string `json:"tenant_id"`
	UnitID   string `json:"unit_id"`
	BranchID string `json:"branch_id"`
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
		if h.bedCoord != nil {
			h.bedCoord.ClearCard(d.CardID)
		}
		h.metaCache.Remove(d.CardID)
	} else {
		h.metaCache.Invalidate(d.CardID)
	}

	// 同 unit 内多卡联动：失效该 unit 下所有卡 meta + 使能缓存；resolver 在删卡时全量失效（已删卡上的 device key 无法再从 DB 枚举）
	if d.TenantID != "" && d.UnitID != "" && h.db != nil {
		h.metaCache.InvalidateCardsInTenantUnit(ctx, d.TenantID, d.UnitID)
		dids, uids := service.DeviceKeysInTenantUnit(ctx, h.db, d.TenantID, d.UnitID)
		if h.enablement != nil {
			h.enablement.InvalidateDevices(dids)
		}
		if h.resolver != nil {
			if d.Op == "deleted" || d.Op == "delete" {
				h.resolver.InvalidateAll()
			} else {
				h.resolver.InvalidateKeys(dids, uids)
			}
		}
	} else if h.resolver != nil {
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
