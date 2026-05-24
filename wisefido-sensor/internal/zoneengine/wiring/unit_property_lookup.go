package wiring

import (
	"context"
	"database/sql"
	"net/netip"
	"sync"
	"time"

	"wisefido-sensor/internal/zoneengine"

	"go.uber.org/zap"
)

// UnitPropertyLookup 实现按 /80 unit 前缀查 units.unit_property 列。
//
// 语义：
//   - unit_property = 0  → Home (B2C/家庭) — 大床，sleepad LeftBed 易假阳（翻身到床沿）
//   - unit_property = 1  → Facility (B2B/机构) — hospital bed，sleepad 全覆盖可信（默认）
//
// 缓存策略与 BedSizeLookup 同：60s TTL（admin 改 unit_property 异步生效，最坏 60s 收敛）；
// 未命中 / 查询失败默认 1 (Facility) —— 跟 DB 默认一致，保守不引入 home-mode patches。
type UnitPropertyLookup struct {
	db     *sql.DB
	logger *zap.Logger

	mu    sync.RWMutex
	cache map[string]unitPropEntry

	queryTimeout time.Duration
	ttl          time.Duration
}

type unitPropEntry struct {
	property int // 0=Home, 1=Facility
	expireAt time.Time
}

// UnitPropertyFacility / UnitPropertyHome 是 unit_property 列的语义常量。
const (
	UnitPropertyHome     = 0
	UnitPropertyFacility = 1
)

func NewUnitPropertyLookup(db *sql.DB, logger *zap.Logger) *UnitPropertyLookup {
	return &UnitPropertyLookup{
		db:           db,
		logger:       logger,
		cache:        make(map[string]unitPropEntry),
		queryTimeout: 2 * time.Second,
		ttl:          60 * time.Second,
	}
}

// IsHomeMode 返回该 /80 unit 是否 Home 模式 (unit_property=0)。
// 未命中 / 查询失败默认 false (Facility) —— 保守不启用 home-mode 协同。
func (l *UnitPropertyLookup) IsHomeMode(unitPrefix string) bool {
	return l.Property(unitPrefix) == UnitPropertyHome
}

// Property 返回 /80 unit 的 unit_property 原值（0/1）。
func (l *UnitPropertyLookup) Property(unitPrefix string) int {
	if unitPrefix == "" {
		return UnitPropertyFacility
	}
	now := time.Now()
	l.mu.RLock()
	if v, ok := l.cache[unitPrefix]; ok && now.Before(v.expireAt) {
		l.mu.RUnlock()
		return v.property
	}
	l.mu.RUnlock()

	prop := l.query(unitPrefix)

	l.mu.Lock()
	l.cache[unitPrefix] = unitPropEntry{property: prop, expireAt: now.Add(l.ttl)}
	l.mu.Unlock()
	return prop
}

// BedMode 实现 zoneengine.BedModeLookup —— 从 bed /96 CIDR 派生 /80 unit prefix 查 unit_property，
// 映射到 BedMode（Home=B2C 大床 / Facility=B2B hospital bed，默认 Facility）。
//
// bedZoneID 形如 "fd00:0:3:111:3:101::/96" → 截 /80 "fd00:0:3:111::/80"。
// 非法 / 不足 80 bit → Facility 默认。
func (l *UnitPropertyLookup) BedMode(bedZoneID string) zoneengine.BedMode {
	p, err := netip.ParsePrefix(bedZoneID)
	if err != nil || p.Bits() < 80 {
		return zoneengine.BedModeFacility
	}
	unitPrefix := netip.PrefixFrom(p.Addr(), 80).Masked().String()
	if l.Property(unitPrefix) == UnitPropertyHome {
		return zoneengine.BedModeHome
	}
	return zoneengine.BedModeFacility
}

// Invalidate 清掉单 unit cache entry（配合 config:card:stream 失效通知）。
func (l *UnitPropertyLookup) Invalidate(unitPrefix string) {
	if unitPrefix == "" {
		return
	}
	l.mu.Lock()
	delete(l.cache, unitPrefix)
	l.mu.Unlock()
}

func (l *UnitPropertyLookup) query(unitPrefix string) int {
	if l.db == nil {
		return UnitPropertyFacility
	}
	ctx, cancel := context.WithTimeout(context.Background(), l.queryTimeout)
	defer cancel()

	var prop int
	err := l.db.QueryRowContext(ctx,
		`SELECT unit_property FROM units WHERE unit_id = $1::INET LIMIT 1`,
		unitPrefix,
	).Scan(&prop)
	if err != nil {
		if err != sql.ErrNoRows && l.logger != nil {
			l.logger.Debug("zoneengine UnitPropertyLookup query failed",
				zap.String("unit_id", unitPrefix),
				zap.Error(err),
			)
		}
		return UnitPropertyFacility
	}
	return prop
}
