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

	"wisefido-sensor-v1/internal/zoneengine"

	"go.uber.org/zap"
)

// BedSizeLookup 实现 zoneengine.BedSizeLookup —— 查 beds.size_kind 列推导 small/large bucket。
//
// 缓存策略：
//   - 命中后无 TTL，进程内永久缓存（床型迁移极罕见，新增床走默认 'standard'=small）
//   - 未命中 / 查询失败默认 "small"（与 zoneengine.BedSizeBucket() 兜底一致）
//   - bedZoneID 不存在的 db 行（罕见，可能 device 派生 /96 不对应真实床）→ 默认 small
type BedSizeLookup struct {
	db     *sql.DB
	logger *zap.Logger

	mu    sync.RWMutex
	cache map[string]string // bedZoneID(/96 CIDR) → bucket "small"/"large"

	queryTimeout time.Duration
}

func NewBedSizeLookup(db *sql.DB, logger *zap.Logger) *BedSizeLookup {
	return &BedSizeLookup{
		db:           db,
		logger:       logger,
		cache:        make(map[string]string),
		queryTimeout: 2 * time.Second,
	}
}

// BedSizeBucket satisfy zoneengine.BedSizeLookup。返回 "small" 或 "large"。
func (l *BedSizeLookup) BedSizeBucket(bedZoneID string) string {
	if bedZoneID == "" {
		return "small"
	}

	// fast path: cache hit
	l.mu.RLock()
	if v, ok := l.cache[bedZoneID]; ok {
		l.mu.RUnlock()
		return v
	}
	l.mu.RUnlock()

	// slow path: db query
	bucket := l.queryBucket(bedZoneID)

	l.mu.Lock()
	l.cache[bedZoneID] = bucket
	l.mu.Unlock()
	return bucket
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

// Invalidate 清除指定 bed 的缓存（床型 schema 变化时调用，目前几乎用不到）。
func (l *BedSizeLookup) Invalidate(bedZoneID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.cache, bedZoneID)
}

// InvalidateAll 清空整个缓存（hot reload 用）。
func (l *BedSizeLookup) InvalidateAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[string]string)
}
