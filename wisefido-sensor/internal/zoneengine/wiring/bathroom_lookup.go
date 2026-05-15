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
// 缓存策略同 [[BedSizeLookup]]：进程内永久缓存，房间类型不会频繁变化。
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
	cache map[string]bool // roomZoneID(/88 CIDR) → isBathroom

	queryTimeout time.Duration
}

func NewBathroomLookup(db *sql.DB, logger *zap.Logger) *BathroomLookup {
	return &BathroomLookup{
		db:           db,
		logger:       logger,
		cache:        make(map[string]bool),
		queryTimeout: 2 * time.Second,
	}
}

// IsBathroom satisfy zoneengine.BathroomLookup。
func (l *BathroomLookup) IsBathroom(roomZoneID string) bool {
	if roomZoneID == "" {
		return false
	}
	l.mu.RLock()
	if v, ok := l.cache[roomZoneID]; ok {
		l.mu.RUnlock()
		return v
	}
	l.mu.RUnlock()

	is := l.queryIsBathroom(roomZoneID)

	l.mu.Lock()
	l.cache[roomZoneID] = is
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

// Invalidate / InvalidateAll 同 BedSizeLookup。
func (l *BathroomLookup) Invalidate(roomZoneID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.cache, roomZoneID)
}

func (l *BathroomLookup) InvalidateAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[string]bool)
}
