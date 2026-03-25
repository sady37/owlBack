package service

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"owl-common/card"
)

// DeviceCardResolver 根据 device_uid 或 device_id (UUID) 解析设备基准；缓存与 config:card 失效配套（InvalidateAll / InvalidateKeys）。
type DeviceCardResolver struct {
	mu     sync.RWMutex
	cache  map[string]baselineCacheEntry
	cardDB *card.CardDB
}

type baselineCacheEntry struct {
	b  card.DeviceBaseline
	ok bool
}

// NewDeviceCardResolver 创建解析器。db 可为 nil（不查库）。
func NewDeviceCardResolver(db *sql.DB) *DeviceCardResolver {
	var cardDB *card.CardDB
	if db != nil {
		cardDB = card.NewCardDB(db)
	}
	return &DeviceCardResolver{
		cache:  make(map[string]baselineCacheEntry),
		cardDB: cardDB,
	}
}

// InvalidateAll 清空缓存，收到 card 配置变更时调用。
func (r *DeviceCardResolver) InvalidateAll() {
	r.mu.Lock()
	r.cache = make(map[string]baselineCacheEntry)
	r.mu.Unlock()
}

// InvalidateKeys 移除给定 device_id / device_uid 作为 key 的缓存项。
func (r *DeviceCardResolver) InvalidateKeys(deviceIDs, deviceUIDs []string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range deviceIDs {
		if id == "" {
			continue
		}
		delete(r.cache, id)
	}
	for _, id := range deviceUIDs {
		if id == "" {
			continue
		}
		delete(r.cache, id)
	}
}

// ResolveBaseline 先缓存后 card.CardDB.ResolveDeviceBaseline（构建唯一入口）。
func (r *DeviceCardResolver) ResolveBaseline(ctx context.Context, deviceKey string) (card.DeviceBaseline, bool) {
	if r == nil || deviceKey == "" {
		return card.DeviceBaseline{}, false
	}
	r.mu.RLock()
	if e, has := r.cache[deviceKey]; has {
		r.mu.RUnlock()
		if !e.ok {
			return card.DeviceBaseline{}, false
		}
		return e.b, true
	}
	r.mu.RUnlock()
	if r.cardDB == nil {
		return card.DeviceBaseline{}, false
	}
	b, ok := r.cardDB.ResolveDeviceBaseline(ctx, deviceKey)
	r.mu.Lock()
	// 只缓存成功结果：首次解析时卡尚未入库会 ok=false，若写入缓存则永不重试，导致 card_id 一直错用 device_id、device_status 永不上线
	if ok {
		r.cache[deviceKey] = baselineCacheEntry{b: b, ok: true}
	}
	r.mu.Unlock()
	return b, ok
}

// Resolve 用 deviceKey 解析 card_id 与 device_uid（兼容旧调用；语义来自 DeviceBaseline）。
func (r *DeviceCardResolver) Resolve(ctx context.Context, deviceKey string) (cardID, deviceUID string, ok bool) {
	b, ok := r.ResolveBaseline(ctx, deviceKey)
	if !ok {
		return "", "", false
	}
	cardID = strings.TrimSpace(b.CardID)
	if cardID == "" {
		cardID = strings.TrimSpace(b.DeviceID)
	}
	deviceUID = strings.TrimSpace(b.DeviceUID)
	if deviceUID == "" {
		deviceUID = deviceKey
	}
	return cardID, deviceUID, true
}
