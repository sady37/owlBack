package consumer

import (
	"context"
	"fmt"
	"net/netip"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// IotPreparedHandler 在 monitor/event/alarm 前统一处理 envelope subject_entity
// （device_ipv6 单程票 / doc/device_ipv6_migration_checklist.md Phase C）：
//
//   - 已填 subject_entity（device-gateway 必填）：直接走原路径
//   - 空 subject_entity（典型 sensor agent layer 1+ 派生消息）：按 device_addr LPM 反查 cards
//     · 命中：填入 cardID 走原路径
//     · 未命中：unbound device（R-009）— subject 留空，consumer 走"无卡 device 状态"分支，
//             不再用 device_id UUID 充 card_id placeholder
//
// 设计动机：sensor agent (wisefido-sensor 等) 派生消息只带 device_addr，subject_entity (card)
// 由 cardagg 反查路由。AI 不染卡概念；cards 表变更立即生效（不依赖 AI 端路由表 reload）。
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
	producer := streamMapStr(raw["producer"])
	subject := streamMapStr(raw["subject_entity"])
	deviceAddrStr := streamMapStr(raw["device_addr"])

	if subject == "" {
		if deviceAddrStr == "" {
			h.logger.Error("iot_prepared: empty subject_entity and device_addr, dropping",
				zap.String("producer", producer),
				zap.String("topic_type", streamMapStr(raw["topic_type"])),
			)
			return nil
		}
		addr, perr := netip.ParseAddr(deviceAddrStr)
		if perr != nil {
			h.logger.Warn("iot_prepared: invalid device_addr, dropping",
				zap.String("device_addr", deviceAddrStr),
				zap.String("producer", producer),
				zap.Error(perr),
			)
			return nil
		}
		cardID := h.metaCache.LookupCardByDeviceAddr(ctx, addr)
		if cardID == "" {
			// unbound device (R-009)：subject 留空，consumer 走"无卡 device 状态"分支。
			// 不再用 device_id 充 subject_entity placeholder（device_ipv6 单程票红线）。
			h.logger.Debug("iot_prepared: unbound device, subject_entity empty",
				zap.String("device_addr", deviceAddrStr),
				zap.String("producer", producer),
			)
		} else {
			subject = cardID
			raw["subject_entity"] = subject
		}
	}

	// GetOrLoad 仅当 subject 是 cardID（INET CIDR）；unbound device subject=="" 不查 cards 表。
	if subject != "" {
		if meta := h.metaCache.GetOrLoad(ctx, subject); meta != nil {
			_ = h.stateSvc.EnsureCardStatePrepared(ctx, subject, meta)
		}
	}
	return h.inner.Handle(ctx, raw)
}
