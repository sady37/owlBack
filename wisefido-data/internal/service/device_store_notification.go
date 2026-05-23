package service

import (
	"context"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/publisher"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// NotifyDeviceStoreBatchAfterUpdate 在 BatchUpdateDeviceStores 成功后，把受影响的 device_uid + device_addr 一次性 publish。
func NotifyDeviceStoreBatchAfterUpdate(ctx context.Context, repo repository.DeviceStoreRepository, pub *publisher.ConfigPublisher, updates []*domain.DeviceStore, log *zap.Logger) {
	if pub == nil || repo == nil || len(updates) == 0 {
		return
	}
	seenUIDs := make(map[string]struct{})
	seenAddrs := make(map[string]struct{})
	var uids, addrs []string
	for _, u := range updates {
		if u == nil || u.DeviceUID == "" {
			continue
		}
		row, err := repo.GetDeviceStore(ctx, u.DeviceUID)
		if err != nil || row == nil || row.DeviceUID == "" {
			continue
		}
		if _, ok := seenUIDs[row.DeviceUID]; !ok {
			seenUIDs[row.DeviceUID] = struct{}{}
			uids = append(uids, row.DeviceUID)
		}
		if row.DeviceAddr != "" {
			if _, ok := seenAddrs[row.DeviceAddr]; !ok {
				seenAddrs[row.DeviceAddr] = struct{}{}
				addrs = append(addrs, row.DeviceAddr)
			}
		}
	}
	if len(uids) == 0 && len(addrs) == 0 {
		return
	}
	if err := pub.PublishConfigChanged(ctx, "update", nil, addrs, uids); err != nil && log != nil {
		log.Warn("PublishConfigChanged failed",
			zap.Int("uids", len(uids)),
			zap.Int("addrs", len(addrs)),
			zap.Error(err))
	}
}

// NotifyDeviceStoreFromStores 按已落库的 device_store 行发通知（如导入成功后的 inserted）。
func NotifyDeviceStoreFromStores(ctx context.Context, pub *publisher.ConfigPublisher, stores []*domain.DeviceStore, changeType string, log *zap.Logger) {
	_ = changeType // 历史保留入参；新 schema 统一 op=update
	if pub == nil || len(stores) == 0 {
		return
	}
	seenUIDs := make(map[string]struct{})
	seenAddrs := make(map[string]struct{})
	var uids, addrs []string
	for _, u := range stores {
		if u == nil || u.DeviceUID == "" {
			continue
		}
		if _, ok := seenUIDs[u.DeviceUID]; !ok {
			seenUIDs[u.DeviceUID] = struct{}{}
			uids = append(uids, u.DeviceUID)
		}
		if u.DeviceAddr != "" {
			if _, ok := seenAddrs[u.DeviceAddr]; !ok {
				seenAddrs[u.DeviceAddr] = struct{}{}
				addrs = append(addrs, u.DeviceAddr)
			}
		}
	}
	if len(uids) == 0 && len(addrs) == 0 {
		return
	}
	if err := pub.PublishConfigChanged(ctx, "update", nil, addrs, uids); err != nil && log != nil {
		log.Warn("PublishConfigChanged failed",
			zap.Int("uids", len(uids)),
			zap.Int("addrs", len(addrs)),
			zap.Error(err))
	}
}
