// spatial_cache.go — sensor 端 per-unit spatial 缓存。
//
// 设计：lazy load by unit /80。任意 query prefix → 派生 /80 → ensureLoaded(unit) →
// 一次 SQL 返该 unit 下全部 elements（unit_property + rooms + beds + devices）。
//
// 替代旧 4 个独立 lookup (bathroom/bed_size/bed_device/unit_property) — 每个原本各自
// 一次 per-key SQL，N+1 严重。新设计每 unit 一次 query，cache 60s TTL。
//
// 字段全部存 DB raw（每层自描述）；分类/桶映射在公共方法 query 时现算。

package wiring

import (
	"context"
	"database/sql"
	"net/netip"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"wisefido-sensor/internal/zoneengine"

	"go.uber.org/zap"
)

// UnitData 一个 /80 unit 下的全部 spatial 元素（一次 SQL 装载）。
type UnitData struct {
	UnitID       netip.Prefix
	UnitProperty int // raw units.unit_property: 0=Home, 1=Facility
	Rooms        map[netip.Prefix]*RoomMeta
	Beds         map[netip.Prefix]*BedMeta
	Devices      map[netip.Addr]*DeviceMeta
	loadedAt     time.Time
}

// RoomMeta /88 room raw 字段。
type RoomMeta struct {
	Prefix       netip.Prefix
	Name         string // raw rooms.room_name
	RoomTypeText string // raw rooms.room_type ("bathroom"/"restroom"/"kitchen"/"")
}

// BedMeta /96 bed raw 字段。
type BedMeta struct {
	Prefix       netip.Prefix
	Name         string // raw beds.bed_name
	SizeKindText string // raw beds.size_kind
}

// DeviceMeta /128 device raw 字段。
type DeviceMeta struct {
	Addr       netip.Addr
	DeviceType string // raw device_factory_meta.device_type ("Radar"/"Sleepad")
	Access     bool
	Monitoring bool
}

// UnitProperty 常量（与 units.unit_property 列对齐）。
const (
	UnitPropertyHome     = 0
	UnitPropertyFacility = 1
)

// SpatialCache per-unit lazy spatial 缓存。
type SpatialCache struct {
	db     *sql.DB
	logger *zap.Logger

	mu    sync.RWMutex
	units map[netip.Prefix]*UnitData

	queryTimeout    time.Duration
	refreshInterval time.Duration // unit TTL；超过即下次 query 时 reload
}

func NewSpatialCache(db *sql.DB, logger *zap.Logger) *SpatialCache {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SpatialCache{
		db:              db,
		logger:          logger,
		units:           make(map[netip.Prefix]*UnitData),
		queryTimeout:    3 * time.Second,
		refreshInterval: 60 * time.Second,
	}
}

// LoadAll boot-time 可选预热：扫 units 表逐 unit ensureLoaded。
// 失败仅 warn；运行时缺数据 lazy 路径自然兜底。
func (c *SpatialCache) LoadAll(ctx context.Context) error {
	if c.db == nil {
		return nil
	}
	qctx, cancel := context.WithTimeout(ctx, c.queryTimeout)
	defer cancel()
	rows, err := c.db.QueryContext(qctx, `SELECT unit_id::text FROM units`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var unitIDs []netip.Prefix
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			continue
		}
		if p, err := netip.ParsePrefix(s); err == nil {
			unitIDs = append(unitIDs, p)
		}
	}
	for _, u := range unitIDs {
		if err := c.loadUnit(ctx, u); err != nil {
			c.logger.Warn("spatial cache loadUnit failed", zap.String("unit", u.String()), zap.Error(err))
		}
	}
	c.logger.Info("spatial cache warmed", zap.Int("units", len(unitIDs)))
	return nil
}

// InvalidateAll 清空所有 unit cache，下次 query 时 lazy reload。
func (c *SpatialCache) InvalidateAll() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.units = make(map[netip.Prefix]*UnitData)
	c.mu.Unlock()
}

// Run 60s tick — 重 load 所有当前 cache 中的 unit（与 InvalidateAll 等效但 mu 临界区更小）。
func (c *SpatialCache) Run(ctx context.Context) {
	t := time.NewTicker(c.refreshInterval)
	defer t.Stop()
	c.logger.Info("spatial cache refresher started", zap.Duration("interval", c.refreshInterval))
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.refreshCachedUnits(ctx)
		}
	}
}

func (c *SpatialCache) refreshCachedUnits(ctx context.Context) {
	c.mu.RLock()
	unitIDs := make([]netip.Prefix, 0, len(c.units))
	for u := range c.units {
		unitIDs = append(unitIDs, u)
	}
	c.mu.RUnlock()
	for _, u := range unitIDs {
		_ = c.loadUnit(ctx, u)
	}
}

// IsBathroom satisfy zoneengine.BathroomLookup — 现算 raw text 分类。
func (c *SpatialCache) IsBathroom(roomZoneID string) bool {
	p, err := netip.ParsePrefix(roomZoneID)
	if err != nil {
		return false
	}
	ud := c.ensureUnit(maskToUnit(p))
	if ud == nil {
		return false
	}
	rm := ud.Rooms[p]
	if rm == nil {
		return false
	}
	return classifyRoomType(rm.RoomTypeText, rm.Name) == card.RoomTypeBathroom
}

// BedSizeBucket satisfy zoneengine.BedSizeLookup — 现算 raw size_kind → bucket。
func (c *SpatialCache) BedSizeBucket(bedZoneID string) string {
	p, err := netip.ParsePrefix(bedZoneID)
	if err != nil {
		return "small"
	}
	ud := c.ensureUnit(maskToUnit(p))
	if ud == nil {
		return "small"
	}
	if bm := ud.Beds[p]; bm != nil {
		return zoneengine.BedSizeBucket(bm.SizeKindText)
	}
	return "small"
}

// BedMode satisfy zoneengine.BedModeLookup。
func (c *SpatialCache) BedMode(bedZoneID string) zoneengine.BedMode {
	p, err := netip.ParsePrefix(bedZoneID)
	if err != nil || p.Bits() < 80 {
		return zoneengine.BedModeFacility
	}
	ud := c.ensureUnit(maskToUnit(p))
	if ud == nil {
		return zoneengine.BedModeFacility
	}
	if ud.UnitProperty == UnitPropertyHome {
		return zoneengine.BedModeHome
	}
	return zoneengine.BedModeFacility
}

// FindPrimaryDevice 替代旧 BedDeviceLookup — 按 alarm 类型选 preferred device_type，
// 从 zone 前缀范围内取第一台 access+monitoring 全开的 /128。
func (c *SpatialCache) FindPrimaryDevice(zoneID, alarmType string) netip.Addr {
	preferred := preferredDeviceTypeForAlarm(alarmType)
	if preferred == "" || zoneID == "" {
		return netip.Addr{}
	}
	p, err := netip.ParsePrefix(zoneID)
	if err != nil {
		return netip.Addr{}
	}
	ud := c.ensureUnit(maskToUnit(p))
	if ud == nil {
		return netip.Addr{}
	}
	candidates := []netip.Prefix{p}
	if p.Bits() > 88 {
		candidates = append(candidates, netip.PrefixFrom(p.Addr(), 88).Masked())
	}
	for _, cidr := range candidates {
		for addr, dm := range ud.Devices {
			if !dm.Access || !dm.Monitoring {
				continue
			}
			if !strings.EqualFold(dm.DeviceType, preferred) {
				continue
			}
			if cidr.Contains(addr) {
				return addr
			}
		}
	}
	return netip.Addr{}
}

// ensureUnit 拿 unit cache；miss/expired → loadUnit。
func (c *SpatialCache) ensureUnit(unitID netip.Prefix) *UnitData {
	if !unitID.IsValid() {
		return nil
	}
	c.mu.RLock()
	ud, ok := c.units[unitID]
	c.mu.RUnlock()
	if ok && time.Since(ud.loadedAt) < c.refreshInterval {
		return ud
	}
	if err := c.loadUnit(context.Background(), unitID); err != nil {
		c.logger.Debug("spatial cache loadUnit", zap.String("unit", unitID.String()), zap.Error(err))
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.units[unitID]
}

// loadUnit 一次 SQL 装载指定 unit 的全部元素（unit_property + rooms + beds + devices）。
//
// SQL 用 UNION ALL flat rows + kind 列标记类型；devices JOIN device_factory_meta 拿 device_type。
// unit_property 单独 1 行（kind='unit'）。
func (c *SpatialCache) loadUnit(ctx context.Context, unitID netip.Prefix) error {
	if c.db == nil {
		return nil
	}
	qctx, cancel := context.WithTimeout(ctx, c.queryTimeout)
	defer cancel()

	// 6 列定值：kind/prefix/name/extra1/extra2/extra3
	// rooms:   kind='room',   prefix=room_id::text, name=room_name, extra1=room_type
	// beds:    kind='bed',    prefix=bed_id::text,  name=bed_name,  extra1=size_kind
	// devices: kind='device', prefix=host(addr),    name='',        extra1=device_type, extra2=access, extra3=monitoring_enabled
	// unit:    kind='unit',   prefix=unit_id::text, name='',        extra1=unit_property
	const q = `
		SELECT 'room'::text AS kind, room_id::text AS prefix, COALESCE(room_name,'') AS name,
		       COALESCE(room_type,'') AS extra1, ''::text AS extra2, ''::text AS extra3
		FROM rooms WHERE room_id <<= $1::inet
		UNION ALL
		SELECT 'bed', bed_id::text, COALESCE(bed_name,''),
		       COALESCE(size_kind,''), '', ''
		FROM beds WHERE bed_id <<= $1::inet
		UNION ALL
		SELECT 'device', host(d.device_addr), '',
		       COALESCE(dfm.device_type::text,''),
		       CASE WHEN COALESCE(d.access,false) THEN '1' ELSE '0' END,
		       CASE WHEN COALESCE(d.monitoring_enabled,false) THEN '1' ELSE '0' END
		FROM devices d JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE d.device_addr <<= $1::inet
		UNION ALL
		SELECT 'unit', unit_id::text, '',
		       COALESCE(unit_property::text,'1'), '', ''
		FROM units WHERE unit_id = $1::inet
	`
	rows, err := c.db.QueryContext(qctx, q, unitID.String())
	if err != nil {
		return err
	}
	defer rows.Close()

	ud := &UnitData{
		UnitID:       unitID,
		UnitProperty: UnitPropertyFacility, // DB 默认
		Rooms:        make(map[netip.Prefix]*RoomMeta),
		Beds:         make(map[netip.Prefix]*BedMeta),
		Devices:      make(map[netip.Addr]*DeviceMeta),
		loadedAt:     time.Now(),
	}
	for rows.Next() {
		var kind, prefixStr, name, e1, e2, e3 string
		if err := rows.Scan(&kind, &prefixStr, &name, &e1, &e2, &e3); err != nil {
			continue
		}
		switch kind {
		case "room":
			if p, err := netip.ParsePrefix(prefixStr); err == nil {
				ud.Rooms[p] = &RoomMeta{Prefix: p, Name: name, RoomTypeText: e1}
			}
		case "bed":
			if p, err := netip.ParsePrefix(prefixStr); err == nil {
				ud.Beds[p] = &BedMeta{Prefix: p, Name: name, SizeKindText: e1}
			}
		case "device":
			if addr, err := netip.ParseAddr(prefixStr); err == nil {
				ud.Devices[addr] = &DeviceMeta{
					Addr: addr, DeviceType: e1, Access: e2 == "1", Monitoring: e3 == "1",
				}
			}
		case "unit":
			if e1 == "0" {
				ud.UnitProperty = UnitPropertyHome
			}
		}
	}

	c.mu.Lock()
	c.units[unitID] = ud
	c.mu.Unlock()
	return nil
}

// maskToUnit 把任意 entry prefix 截到 /80 unit。
func maskToUnit(p netip.Prefix) netip.Prefix {
	if p.Bits() < 80 {
		return netip.Prefix{}
	}
	return netip.PrefixFrom(p.Addr(), 80).Masked()
}

// classifyRoomType room_type text → card.RoomType* int。
// 兼容老数据 room_type=NULL 时按 name 关键词兜底 bathroom。
func classifyRoomType(roomType, roomName string) int {
	t := strings.ToLower(strings.TrimSpace(roomType))
	switch t {
	case "bathroom", "restroom":
		return card.RoomTypeBathroom
	case "kitchen":
		return card.RoomTypeKitchen
	}
	if t == "" {
		n := strings.ToLower(roomName)
		for _, kw := range []string{"bathroom", "restroom", "toilet", "washroom"} {
			if strings.Contains(n, kw) {
				return card.RoomTypeBathroom
			}
		}
	}
	return card.RoomTypeDefault
}

// preferredDeviceTypeForAlarm — 与旧 bed_device_lookup 同语义。
// 床位 alarm 主源 radar；fall 类必走 radar；Stay/NightAbsence 不归因 bed device。
func preferredDeviceTypeForAlarm(alarmType string) string {
	switch alarmType {
	case alarm.LeftBed, alarm.InBed:
		return "Radar"
	case alarm.Fall, alarm.SuspectedFall, alarm.SittingOnGround, alarm.SuspectedSittingOnGround:
		return "Radar"
	}
	return ""
}
