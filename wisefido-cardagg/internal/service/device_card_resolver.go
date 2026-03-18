package service

import (
	"context"
	"database/sql"
	"sync"

	"owl-common/card"
)

// DeviceCardResolver 根据 device_uid 或 device_id (UUID) 解析出 card_id 与 device_uid。
// 带缓存，收到 config:card:stream 时调用 InvalidateAll 清空，避免每次查 DB。
type DeviceCardResolver struct {
	mu     sync.RWMutex
	cache  map[string]struct{ cardID, deviceUID string; ok bool }
	cardDB *card.CardDB
	db     *sql.DB
}

// NewDeviceCardResolver 创建解析器。db 可为 nil（仅做 fallback 时不用 DB）。
func NewDeviceCardResolver(db *sql.DB) *DeviceCardResolver {
	var cardDB *card.CardDB
	if db != nil {
		cardDB = card.NewCardDB(db)
	}
	return &DeviceCardResolver{
		cache:  make(map[string]struct{ cardID, deviceUID string; ok bool }),
		cardDB: cardDB,
		db:     db,
	}
}

// InvalidateAll 清空缓存，收到 card 配置变更时调用，下次 Resolve 从 DB 重载。
func (r *DeviceCardResolver) InvalidateAll() {
	r.mu.Lock()
	r.cache = make(map[string]struct{ cardID, deviceUID string; ok bool })
	r.mu.Unlock()
}

// Resolve 用 deviceKey（device_uid 或 device_id UUID）查 card_id 与 device_uid。先查缓存，未命中再查 DB 并回填缓存。
func (r *DeviceCardResolver) Resolve(ctx context.Context, deviceKey string) (cardID, deviceUID string, ok bool) {
	if deviceKey == "" {
		return "", "", false
	}
	r.mu.RLock()
	if c, has := r.cache[deviceKey]; has {
		r.mu.RUnlock()
		return c.cardID, c.deviceUID, c.ok
	}
	r.mu.RUnlock()

	cid, uid, found := r.resolveFromDB(ctx, deviceKey)
	if !found {
		cid, uid = deviceKey, deviceKey
	}

	r.mu.Lock()
	r.cache[deviceKey] = struct{ cardID, deviceUID string; ok bool }{cid, uid, found}
	r.mu.Unlock()
	return cid, uid, found
}

func (r *DeviceCardResolver) resolveFromDB(ctx context.Context, deviceKey string) (cardID, deviceUID string, ok bool) {
	if r.cardDB != nil {
		info, err := r.cardDB.LookupCard(ctx, deviceKey)
		if err == nil && info != nil && info.CardID != "" {
			return info.CardID, info.DeviceUID, true
		}
	}
	if r.db != nil {
		var cid, uid string
		err := r.db.QueryRowContext(ctx, `
			SELECT c.card_id::text, COALESCE(j->>'device_uid', j->>'device_id')
			FROM cards c, jsonb_array_elements(COALESCE(c.devices, '[]'::jsonb)) AS j
			WHERE j->>'device_id' = $1 OR j->>'device_uid' = $1
			LIMIT 1
		`, deviceKey).Scan(&cid, &uid)
		if err == nil && cid != "" {
			if uid == "" {
				uid = deviceKey
			}
			return cid, uid, true
		}
	}
	// 未绑卡：用 device_id 充当 card_id，与 FillDeviceOnlineStatusFromCardagg 一致（无 card 时 key=device_id）
	if r.cardDB != nil {
		info, err := r.cardDB.LookupDeviceOnly(ctx, deviceKey)
		if err == nil && info != nil && info.DeviceID != "" {
			return info.CardID, info.DeviceUID, true
		}
	}
	// 仅在 device_store、未入 devices 时：用 device_store.device_id 充当 card_id
	if r.db != nil {
		var did, uid string
		err := r.db.QueryRowContext(ctx, `
			SELECT device_id::text, COALESCE(device_uid, '') FROM device_store WHERE device_uid = $1 OR device_id::text = $1 LIMIT 1
		`, deviceKey, deviceKey).Scan(&did, &uid)
		if err == nil && did != "" {
			if uid == "" {
				uid = deviceKey
			}
			return did, uid, true
		}
	}
	return "", "", false
}
