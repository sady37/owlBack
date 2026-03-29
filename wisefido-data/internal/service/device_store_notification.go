package service

import (
	"context"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

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
		if err := pub.PublishCardChangeForDevice(ctx, row.TenantID, row.DeviceID, "device_store_updated"); err != nil && log != nil {
			log.Warn("PublishCardChangeForDevice failed",
				zap.String("tenant_id", row.TenantID),
				zap.String("device_id", row.DeviceID),
				zap.Error(err))
		}
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
		if err := pub.PublishCardChangeForDevice(ctx, u.TenantID, u.DeviceID, changeType); err != nil && log != nil {
			log.Warn("PublishCardChangeForDevice failed",
				zap.String("tenant_id", u.TenantID),
				zap.String("device_id", u.DeviceID),
				zap.String("change_type", changeType),
				zap.Error(err))
		}
	}
}
