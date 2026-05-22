package service

import (
	"context"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// NotifyDeviceStoreBatchAfterUpdate 在 BatchUpdateDeviceStores 成功后，按受影响的 device_uid 各发一条（去重）。
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
		if err != nil || row == nil || row.DeviceUID == "" {
			continue
		}
		if _, ok := seen[row.DeviceUID]; ok {
			continue
		}
		seen[row.DeviceUID] = struct{}{}
		if err := pub.PublishCardChangeForDevice(ctx, row.TenantID, row.DeviceAddr, "device_store_updated", row.DeviceUID); err != nil && log != nil {
			log.Warn("PublishCardChangeForDevice failed",
				zap.String("tenant_id", row.TenantID),
				zap.String("device_addr", row.DeviceAddr),
				zap.String("device_uid", row.DeviceUID),
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
		if u == nil || u.DeviceUID == "" {
			continue
		}
		if _, ok := seen[u.DeviceUID]; ok {
			continue
		}
		seen[u.DeviceUID] = struct{}{}
		if err := pub.PublishCardChangeForDevice(ctx, u.TenantID, u.DeviceAddr, changeType, u.DeviceUID); err != nil && log != nil {
			log.Warn("PublishCardChangeForDevice failed",
				zap.String("tenant_id", u.TenantID),
				zap.String("device_addr", u.DeviceAddr),
				zap.String("device_uid", u.DeviceUID),
				zap.String("change_type", changeType),
				zap.Error(err))
		}
	}
}
