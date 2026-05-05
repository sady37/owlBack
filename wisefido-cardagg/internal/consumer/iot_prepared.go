package consumer

import (
	"context"
	"fmt"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// IotPreparedHandler 在 monitor/event/alarm 前统一处理 envelope subject_entity（Phase A v2）：
//
//   - 已填 subject_entity（device-gateway 必填）：直接走原路径
//   - 空 subject_entity（典型 sensor agent layer 1+ 派生消息）：按 device_uid/device_id 反查 cards
//     · 单卡：填入 raw 走原路径（业务模型常态——住户卡或 public 卡 1:1 绑设备）
//     · 多卡：业务模型当前不允许（每个设备只属一个 resident 或 public 区域），
//             实际出现说明 cards 表数据异常 → 记 warn 取首张兜底
//     · 零卡：device 未绑卡（新装未配置 / monitoring_enabled=false / 公共区未起卡）
//             → 用 device_id 充当 subject_entity（与 device-gateway 端 SubjectEntityForBaseline 占位一致）
//
// 设计动机（协议层北极星 layer 1 / 2 解耦）：sensor agent (wisefido-sensor 等) 派生消息
// 只带 device 标识（producer-centric），subject_entity (card) 由 cardagg 反查路由。
// AI 不染卡概念；cards 表变更立即生效（不依赖 AI 端路由表 reload）。
// 未绑卡设备消息不丢，保留 device-id 占位走完链路——device 上线 / 入网 / 巡检场景需要。
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
	deviceUID := streamMapStr(raw["device_uid"])
	deviceID := streamMapStr(raw["device_id"])

	if subject == "" {
		// sensor agent / 未填的 publisher → cardagg 反查
		if deviceUID == "" && deviceID == "" {
			h.logger.Error("iot_prepared: empty subject_entity, no device identity, dropping",
				zap.String("producer", producer),
				zap.String("topic_type", streamMapStr(raw["topic_type"])),
			)
			return nil
		}
		cards := h.metaCache.LookupCardsByDevice(ctx, deviceUID, deviceID)
		if len(cards) == 0 {
			// device 未绑卡：用 device_id 占位（保留 unbound 链路）
			if deviceID == "" {
				h.logger.Warn("iot_prepared: unbound device with no device_id, dropping",
					zap.String("device_uid", deviceUID),
					zap.String("producer", producer),
				)
				return nil
			}
			subject = deviceID
			h.logger.Debug("iot_prepared: unbound device, using device_id as subject_entity",
				zap.String("device_id", deviceID),
				zap.String("device_uid", deviceUID),
				zap.String("producer", producer),
			)
		} else {
			if len(cards) > 1 {
				// 业务模型：每个设备只属一个 resident 或一张 public 卡。出现多卡说明
				// cards.devices JSONB 数据异常（重复绑定？）。取首张（DISTINCT 后稳定排序）兜底，
				// 同时 warn 出来供运维排查。
				h.logger.Warn("iot_prepared: device unexpectedly bound to multiple cards (data anomaly), using first",
					zap.Strings("cards", cards),
					zap.String("device_uid", deviceUID),
				)
			}
			subject = cards[0]
		}
		raw["subject_entity"] = subject
	}

	// GetOrLoad 仅当 subject 是 cardID（不是 device_id 占位）。
	// device_id 占位场景（未绑卡）不会有 cards 表行，跳过即可——下游 alarm 落库由 InsertAlarmAndUpdateCard
	// 经 metadata.card_id 兜底。
	if subject != "" && subject != deviceID {
		if meta := h.metaCache.GetOrLoad(ctx, subject); meta != nil {
			_ = h.stateSvc.EnsureCardStatePrepared(ctx, subject, meta)
		}
	}
	return h.inner.Handle(ctx, raw)
}
