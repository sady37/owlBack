// Package wiring 是 zone engine 与 sensor 主进程的组装层（DB / MonitorBuffer / 配置文件等
// 外部依赖在这里包装为 zone engine 接口）。
//
// 职责分层：
//   - internal/zoneengine 是 leaf 包，只依赖 owl-common；规则纯净
//   - wiring 包负责把 DB / MonitorBuffer / yaml 包装成 BedSizeLookup / BathroomLookup /
//     VitalSource 接口实现
//   - cmd/wisefido-sensor/main.go 调 wiring.Setup 一次性接好
package wiring

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"wisefido-sensor/internal/zoneengine"

	"go.uber.org/zap"
)

// BedSizeLookup 实现 zoneengine.BedSizeLookup —— 查 beds.size_kind 列推导 small/large bucket。
//
// 缓存策略：60s TTL（admin 改 size_kind 异步生效，最坏 60s 收敛）；
// 未命中 / 查询失败默认 "small"（与 zoneengine.BedSizeBucket() 兜底一致）。
type BedSizeLookup struct {
	db     *sql.DB
	logger *zap.Logger

	mu    sync.RWMutex
	cache map[string]bedSizeCacheEntry

	queryTimeout time.Duration
	ttl          time.Duration
}

type bedSizeCacheEntry struct {
	bucket   string // "small" / "large"
	expireAt time.Time
}

func NewBedSizeLookup(db *sql.DB, logger *zap.Logger) *BedSizeLookup {
	return &BedSizeLookup{
		db:           db,
		logger:       logger,
		cache:        make(map[string]bedSizeCacheEntry),
		queryTimeout: 2 * time.Second,
		ttl:          60 * time.Second,
	}
}

// BedSizeBucket satisfy zoneengine.BedSizeLookup。返回 "small" 或 "large"。
func (l *BedSizeLookup) BedSizeBucket(bedZoneID string) string {
	if bedZoneID == "" {
		return "small"
	}
	now := time.Now()
	l.mu.RLock()
	if v, ok := l.cache[bedZoneID]; ok && now.Before(v.expireAt) {
		l.mu.RUnlock()
		return v.bucket
	}
	l.mu.RUnlock()

	bucket := l.queryBucket(bedZoneID)

	l.mu.Lock()
	l.cache[bedZoneID] = bedSizeCacheEntry{bucket: bucket, expireAt: now.Add(l.ttl)}
	l.mu.Unlock()
	return bucket
}

// InvalidateAll 清空整个 cache（hot-reload / schema 变更后用）。
func (l *BedSizeLookup) InvalidateAll() {
	l.mu.Lock()
	l.cache = make(map[string]bedSizeCacheEntry)
	l.mu.Unlock()
}

func (l *BedSizeLookup) queryBucket(bedZoneID string) string {
	if l.db == nil {
		return "small"
	}
	ctx, cancel := context.WithTimeout(context.Background(), l.queryTimeout)
	defer cancel()

	var sizeKind string
	// bed_id 列是 INET PRIMARY KEY，存的是 /96；bedZoneID 也是 /96 CIDR text → 直接 ::INET 比较
	err := l.db.QueryRowContext(ctx,
		`SELECT size_kind FROM beds WHERE bed_id = $1::INET LIMIT 1`,
		bedZoneID,
	).Scan(&sizeKind)
	if err != nil {
		if err != sql.ErrNoRows && l.logger != nil {
			l.logger.Debug("zoneengine BedSizeLookup query failed",
				zap.String("bed_id", bedZoneID),
				zap.Error(err),
			)
		}
		return "small"
	}
	return zoneengine.BedSizeBucket(sizeKind)
}

