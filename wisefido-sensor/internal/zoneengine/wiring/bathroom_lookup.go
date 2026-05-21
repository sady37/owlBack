package wiring

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BathroomLookup 实现 zoneengine.BathroomLookup —— 查 rooms.room_type='bathroom' 推导。
//
// 缓存策略：60s TTL（admin 改 room_type 异步生效，最坏 60s 收敛）。
//
// 判定优先级：
//  1. rooms.room_type 命中 'bathroom' / 'restroom' → bathroom
//  2. （回退）rooms.room_name 含关键词 → bathroom（兼容老数据 room_type 为 NULL）
//  3. 都不命中 → 非 bathroom
//
// 14_rooms.sql 约定 room_type='bathroom' 是首选；keyword 仅作老数据兜底。
type BathroomLookup struct {
	db     *sql.DB
	logger *zap.Logger

	mu    sync.RWMutex
	cache map[string]bathroomCacheEntry // roomZoneID(/88 CIDR) → entry

	queryTimeout time.Duration
	ttl          time.Duration
}

type bathroomCacheEntry struct {
	isBathroom bool
	expireAt   time.Time
}

func NewBathroomLookup(db *sql.DB, logger *zap.Logger) *BathroomLookup {
	return &BathroomLookup{
		db:           db,
		logger:       logger,
		cache:        make(map[string]bathroomCacheEntry),
		queryTimeout: 2 * time.Second,
		ttl:          60 * time.Second,
	}
}

// IsBathroom satisfy zoneengine.BathroomLookup。
func (l *BathroomLookup) IsBathroom(roomZoneID string) bool {
	if roomZoneID == "" {
		return false
	}
	now := time.Now()
	l.mu.RLock()
	if v, ok := l.cache[roomZoneID]; ok && now.Before(v.expireAt) {
		l.mu.RUnlock()
		return v.isBathroom
	}
	l.mu.RUnlock()

	is := l.queryIsBathroom(roomZoneID)

	l.mu.Lock()
	l.cache[roomZoneID] = bathroomCacheEntry{isBathroom: is, expireAt: now.Add(l.ttl)}
	l.mu.Unlock()
	return is
}

func (l *BathroomLookup) queryIsBathroom(roomZoneID string) bool {
	if l.db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), l.queryTimeout)
	defer cancel()

	var roomType sql.NullString
	var roomName sql.NullString
	err := l.db.QueryRowContext(ctx,
		`SELECT room_type, room_name FROM rooms WHERE room_id = $1::INET LIMIT 1`,
		roomZoneID,
	).Scan(&roomType, &roomName)
	if err != nil {
		if err != sql.ErrNoRows && l.logger != nil {
			l.logger.Debug("zoneengine BathroomLookup query failed",
				zap.String("room_id", roomZoneID),
				zap.Error(err),
			)
		}
		return false
	}
	if roomType.Valid {
		t := strings.ToLower(roomType.String)
		if t == "bathroom" || t == "restroom" {
			return true
		}
	}
	// 兜底：老数据 room_type=NULL 时按 name 关键词
	if roomName.Valid {
		n := strings.ToLower(roomName.String)
		for _, kw := range []string{"bathroom", "restroom", "toilet", "washroom"} {
			if strings.Contains(n, kw) {
				return true
			}
		}
	}
	return false
}

