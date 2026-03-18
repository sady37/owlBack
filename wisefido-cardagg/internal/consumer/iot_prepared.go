package consumer

import (
	"context"
	"fmt"

	"wisefido-cardagg/internal/service"
)

// IotPreparedHandler 在 monitor/event/alarm 前统一填充 card_id：gateway 不填时用 device_id 查绑定的 card_id；查不到则用 device_id 充当 card_id。仅当解析到真实 card（DB 有绑定）时才查库占位。
type IotPreparedHandler struct {
	resolver  *service.DeviceCardResolver
	stateSvc  *service.StateService
	metaCache *service.DeviceMetaCache
	inner     StreamHandler
}

func NewIotPreparedHandler(resolver *service.DeviceCardResolver, stateSvc *service.StateService, metaCache *service.DeviceMetaCache, inner StreamHandler) StreamHandler {
	return &IotPreparedHandler{resolver: resolver, stateSvc: stateSvc, metaCache: metaCache, inner: inner}
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
	deviceID := streamMapStr(raw["device_id"])   // 设备 UUID，业务主键
	deviceUID := streamMapStr(raw["device_uid"]) // 设备序列号，如雷达 MAC
	unboundFallback := false                     // 未绑卡时用 device_id 充当 card_id，不查库
	// 需要解析：card_id 为空或非 UUID 时，用 resolveKey 查 card_id。iotHead 同时有 device_id/device_uid 时优先用 device_id(UUID) 与 cards.devices 一致，便于稳定解析出 card_id。
	if cardID == "" || !service.IsUUID(cardID) || deviceUID == "" {
		resolveKey := deviceID
		if resolveKey == "" || !service.IsUUID(resolveKey) {
			resolveKey = deviceUID
		}
		if resolveKey == "" {
			resolveKey = deviceID
		}
		if resolveKey == "" {
			return h.inner.Handle(ctx, raw)
		}
		resolvedCardID, resolvedUID, found := h.resolver.Resolve(ctx, resolveKey)
		if found {
			cardID = resolvedCardID
			deviceUID = resolvedUID
			// 未绑卡时 resolver 返回 card_id=device_id(UUID)，回填 device_id 供下游与 alarm_db 等使用
			if deviceID == "" && service.IsUUID(cardID) {
				raw["device_id"] = cardID
			}
		} else {
			// 查询失败：用 device_id 充当 card_id（UUID），无则用 resolveKey（多为 device_uid）
			cardID = deviceID
			if cardID == "" {
				cardID = resolveKey
			}
			if deviceUID == "" {
				deviceUID = resolveKey
			}
			unboundFallback = true
		}
		raw["card_id"] = cardID
		raw["device_uid"] = deviceUID
	}
	// 仅当为真实 card（gateway 已填或 DB 解析到）时查库占位；未绑卡用 device_id 充当的不查库
	if !unboundFallback && service.IsUUID(cardID) {
		if meta := h.metaCache.GetOrLoad(ctx, cardID); meta != nil {
			_ = h.stateSvc.EnsureCardStatePrepared(ctx, cardID, meta)
		}
	}
	return h.inner.Handle(ctx, raw)
}
