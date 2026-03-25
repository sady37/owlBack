package service

import (
	"context"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

// PublishDeviceStoreConfigChange 发送 config.card.device_store，供 qinglan 等按 device_id 刷新 baseline。
func PublishDeviceStoreConfigChange(ctx context.Context, pub *publisher.ConfigPublisher, tenantID, deviceID, changeType string, log *zap.Logger) {
	if pub == nil || deviceID == "" {
		return
	}
	extra := map[string]interface{}{
		"device_id":    deviceID,
		"change_type":  changeType,
	}
	if err := pub.PublishCardChangeMessageWithExtraAndType(ctx, tenantID, "", "", "", rediscommon.ConfigCardDeviceStoreChanged, extra); err != nil {
		if log != nil {
			log.Warn("PublishDeviceStoreConfigChange failed",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", deviceID),
				zap.String("change_type", changeType),
				zap.Error(err))
		}
	}
}

// NotifyDeviceStoreBatchAfterUpdate 在 BatchUpdateDeviceStores 成功后，按受影响的 device_id 各发一条（去重）。
func NotifyDeviceStoreBatchAfterUpdate(ctx context.Context, repo repository.DeviceStoreRepository, pub *publisher.ConfigPublisher, updates []*domain.DeviceStore, log *zap.Logger) {
	if pub == nil || repo == nil || len(updates) == 0 {
		return
	}
	seen := make(map[string]struct{})
	for _, u := range updates {
		if u == nil || u.DeviceUID == "" {
			continue
		}
		row, err := repo.GetDeviceStore(ctx, u.DeviceUID)
		if err != nil || row == nil || row.DeviceID == "" {
			continue
		}
		if _, ok := seen[row.DeviceID]; ok {
			continue
		}
		seen[row.DeviceID] = struct{}{}
		PublishDeviceStoreConfigChange(ctx, pub, row.TenantID, row.DeviceID, "device_store_updated", log)
	}
}

// NotifyDeviceStoreFromStores 按已落库的 device_store 行发通知（如导入成功后的 inserted）。
func NotifyDeviceStoreFromStores(ctx context.Context, pub *publisher.ConfigPublisher, stores []*domain.DeviceStore, changeType string, log *zap.Logger) {
	if pub == nil || len(stores) == 0 {
		return
	}
	seen := make(map[string]struct{})
	for _, u := range stores {
		if u == nil || u.DeviceID == "" {
			continue
		}
		if _, ok := seen[u.DeviceID]; ok {
			continue
		}
		seen[u.DeviceID] = struct{}{}
		PublishDeviceStoreConfigChange(ctx, pub, u.TenantID, u.DeviceID, changeType, log)
	}
}
