package wiring

import (
	"context"
	"database/sql"
	"net/netip"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"

	"go.uber.org/zap"
)

// BedDeviceLookup —— 给定 zone 前缀（bed_id /96 或 room_id /88）+ alarm 类型，
// 返回应作为 trigger 设备归因的 /128 IPv6。
//
// 由 BackChannelAlarmFirer 在 publish envelope 时调用，替代旧 pickDeviceAddr
// fallback 给 ZoneID prefix（zero host）导致下游 snapshot JOIN 全 NULL 的 bug。
//
// 策略（v1）：
//   - LeftBed / BedNightAbsence / InBed / Fall / SuspectedFall /
//     SittingOnGround / SuspectedSittingOnGround → 优先 Radar
//     (用户拍板 leftBed 以 09E7 雷达为主、fall 以 D523 雷达为主，per memory
//      track_fusion_and_gate_cardid + d5f7_ghost_cases 等)
//   - 其他 (Stay/NightAbsence on room/bathroom) → return zero Addr（不归因到 bed device）
//   - 严格 bed_id /96 prefix 查；查不到再扩 room_id /88（同房 radar 兼容多床）；
//     都没有 return zero Addr，由 caller fallback。
//
// 缓存策略：进程内永久（设备绑定一般不动），新增/迁移 device 后 Invalidate。
type BedDeviceLookup struct {
	db     *sql.DB
	logger *zap.Logger

	mu    sync.RWMutex
	cache map[bedDeviceCacheKey]netip.Addr // zoneID+deviceType → /128 Addr

	queryTimeout time.Duration
}

type bedDeviceCacheKey struct {
	zoneID     string // bed_id /96 CIDR 文本
	deviceType string // "Radar" / "Sleepad" / ""
}

func NewBedDeviceLookup(db *sql.DB, logger *zap.Logger) *BedDeviceLookup {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BedDeviceLookup{
		db:           db,
		logger:       logger,
		cache:        make(map[bedDeviceCacheKey]netip.Addr),
		queryTimeout: 2 * time.Second,
	}
}

// FindPrimaryDevice 主接口：按 alarm 类型选 preferred device_type，从 zone 前缀范围内拿 /128。
// alarmType 未识别 / 不需要归因（如 Stay）时返回 zero Addr；caller 自行 fallback。
func (l *BedDeviceLookup) FindPrimaryDevice(zoneID, alarmType string) netip.Addr {
	preferred := preferredDeviceTypeForAlarm(alarmType)
	if preferred == "" {
		return netip.Addr{}
	}
	if zoneID == "" {
		return netip.Addr{}
	}

	// 1) bed_id /96 严格查
	if addr := l.lookupCached(zoneID, preferred); addr.IsValid() {
		return addr
	}
	// 2) 扩到 room_id /88 — 把 /96 截短为 /88（同房雷达兼容多床）
	if roomZone := truncatePrefix(zoneID, 88); roomZone != "" && roomZone != zoneID {
		if addr := l.lookupCached(roomZone, preferred); addr.IsValid() {
			return addr
		}
	}
	return netip.Addr{}
}

func (l *BedDeviceLookup) lookupCached(zoneID, deviceType string) netip.Addr {
	key := bedDeviceCacheKey{zoneID: zoneID, deviceType: deviceType}
	l.mu.RLock()
	if v, ok := l.cache[key]; ok {
		l.mu.RUnlock()
		return v
	}
	l.mu.RUnlock()

	addr := l.query(zoneID, deviceType)

	l.mu.Lock()
	l.cache[key] = addr
	l.mu.Unlock()
	return addr
}

func (l *BedDeviceLookup) query(zoneID, deviceType string) netip.Addr {
	if l.db == nil {
		return netip.Addr{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), l.queryTimeout)
	defer cancel()

	// 同 prefix 内取第一条 access+monitoring 全开的指定类型 device
	const q = `
		SELECT d.device_ipv6
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE d.device_ipv6 <<= $1::INET
		  AND dfm.device_type = $2::device_type_enum
		  AND d.access = true
		  AND d.monitoring_enabled = true
		ORDER BY d.device_ipv6
		LIMIT 1
	`
	var addrStr string
	err := l.db.QueryRowContext(ctx, q, zoneID, deviceType).Scan(&addrStr)
	if err != nil {
		if err != sql.ErrNoRows {
			l.logger.Debug("BedDeviceLookup query failed",
				zap.String("zone_id", zoneID),
				zap.String("device_type", deviceType),
				zap.Error(err))
		}
		return netip.Addr{}
	}
	// PG INET 返回带 /128 后缀，netip.ParseAddr 不接受 prefix
	if i := strings.IndexByte(addrStr, '/'); i >= 0 {
		addrStr = addrStr[:i]
	}
	addr, err := netip.ParseAddr(addrStr)
	if err != nil {
		return netip.Addr{}
	}
	return addr
}

// Invalidate / InvalidateAll —— 设备 rebind / 新设备入网后调用。
func (l *BedDeviceLookup) Invalidate(zoneID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k := range l.cache {
		if k.zoneID == zoneID {
			delete(l.cache, k)
		}
	}
}

func (l *BedDeviceLookup) InvalidateAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[bedDeviceCacheKey]netip.Addr)
}

// preferredDeviceTypeForAlarm 决定一个 alarm 类型应该归因到哪种 device。
//
// 注：alarm 名跟 owl-common/alarm 包对齐。新增 alarm 类型时同步扩展。
func preferredDeviceTypeForAlarm(alarmType string) string {
	switch alarmType {
	case alarm.LeftBed, alarm.InBed, alarm.BedNightAbsence:
		return "Radar" // 用户拍板：床位 alarm 主源 radar
	case alarm.Fall, alarm.SuspectedFall, alarm.SittingOnGround, alarm.SuspectedSittingOnGround:
		return "Radar" // fall 类必走 radar
	default:
		// Stay (bathroom 占用), NightAbsence (room 维度), 其他 —— 不归因到 bed device
		return ""
	}
}

// truncatePrefix 把 CIDR 文本 (如 "fd00:0:3:111:3:301::/96") 截短到 newMask。
// 失败返回空字符串。
func truncatePrefix(cidr string, newMask int) string {
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		return ""
	}
	if p.Bits() <= newMask {
		return ""
	}
	masked := netip.PrefixFrom(p.Addr(), newMask).Masked()
	return masked.String()
}
