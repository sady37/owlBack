// spatial_cache.go — cardagg 的 spatial 真相 cache。
//
// 三层各管各的，无冗余：
//
//	entries[prefix] *Entry   ground truth — 全部 spatial entry（rooms+beds+devices）按 CIDR 索引
//	devices[addr]   *DeviceMeta  设备非空间属性（DeviceUID / DeviceType；IPv6 mask 算不出来）
//	cards[card_id]  *CardMeta    卡级 snapshot（has_bed / has_bathroom / unit_type 等旗标）
//
// 路由：sensor.subject_entity → entries[subject_entity].CardID → 写该 card hash。
// 不再用 LookupCardByPrefix LPM；用 entry-table 精确等值反查。

package service

import (
	"context"
	"database/sql"
	"net/netip"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Entry 是 spatial entity 的统一表示（rooms / beds / devices 一锅）。
// mask 自描述层级（/88 room / /96 bed / /128 device）；CardID 反挂归属容器。
type Entry struct {
	Prefix netip.Prefix
	Name   string       // room_name / bed_name / device_type（display 用）
	CardID netip.Prefix // 归属 card 容器
}

func (e *Entry) IsRoom() bool   { return e != nil && e.Prefix.Bits() == 88 }
func (e *Entry) IsBed() bool    { return e != nil && e.Prefix.Bits() == 96 }
func (e *Entry) IsDevice() bool { return e != nil && e.Prefix.Bits() == 128 }

// DeviceMeta 设备的非空间属性 — IPv6 mask 描述不了。
type DeviceMeta struct {
	DeviceAddr netip.Addr
	DeviceUID  string // devices.device_uid（OTA / 厂家通信 key）
	DeviceType string // "Radar" / "SleepPad" / etc.
}

// RoomMeta room 的非空间原始字段。
type RoomMeta struct {
	Prefix    netip.Prefix
	Name      string
	RoomType  int  // rooms.room_type: 0=Default, 1=Bathroom, 2=Kitchen
	IsPrimary bool // rooms.is_primary
}

// UnitMeta unit 的非空间原始字段。
type UnitMeta struct {
	Prefix       netip.Prefix
	UnitType     int    // units.unit_type: 0=Unknown, 1=Private, 2=Share, 3=Public
	UnitProperty int    // units.unit_property: 0=Home, 1=Facility
	Timezone     string // effective_tz (COALESCE unit→branch→tenant, IANA 名)
}

// CardMeta card 自家原始字段（不存任何派生 snapshot；has_bed / has_bathroom / unit_type
// 等派生值由 cardagg 按需从 entries + rooms + units 现算）。
type CardMeta struct {
	CardID     netip.Prefix
	TenantPref netip.Prefix // /48 派生
	CardName   string       // 'nickname' / 'NoOne' / 'public'
}

func (m *CardMeta) IsPublicCard() bool { return m != nil && m.CardName == "public" }

// SpatialCache 持各层原始数据，无派生 snapshot。统一生命周期。
type SpatialCache struct {
	mu      sync.RWMutex
	entries map[netip.Prefix]*Entry
	devices map[netip.Addr]*DeviceMeta
	rooms   map[netip.Prefix]*RoomMeta
	units   map[netip.Prefix]*UnitMeta
	cards   map[netip.Prefix]*CardMeta

	db     *sql.DB
	logger *zap.Logger
}

func NewSpatialCache(db *sql.DB, logger *zap.Logger) *SpatialCache {
	return &SpatialCache{
		entries: make(map[netip.Prefix]*Entry),
		devices: make(map[netip.Addr]*DeviceMeta),
		rooms:   make(map[netip.Prefix]*RoomMeta),
		units:   make(map[netip.Prefix]*UnitMeta),
		cards:   make(map[netip.Prefix]*CardMeta),
		db:      db,
		logger:  logger,
	}
}

// LoadAll 启动 + Invalidate 后调；各表全量 reload 填五个 map。
func (c *SpatialCache) LoadAll(ctx context.Context) error {
	entries, devices, err := c.queryEntries(ctx)
	if err != nil {
		return err
	}
	rooms, err := c.queryRooms(ctx)
	if err != nil {
		return err
	}
	units, err := c.queryUnits(ctx)
	if err != nil {
		return err
	}
	cards, err := c.queryCards(ctx)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.entries = entries
	c.devices = devices
	c.rooms = rooms
	c.units = units
	c.cards = cards
	c.mu.Unlock()
	c.logger.Info("spatial cache loaded",
		zap.Int("entries", len(entries)),
		zap.Int("devices", len(devices)),
		zap.Int("rooms", len(rooms)),
		zap.Int("units", len(units)),
		zap.Int("cards", len(cards)))
	return nil
}

// Invalidate 全量 reload。schema 改频率低，全量比增量简单可靠。
func (c *SpatialCache) Invalidate(ctx context.Context) {
	if err := c.LoadAll(ctx); err != nil {
		c.logger.Warn("spatial cache invalidate reload failed", zap.Error(err))
	}
}

// LookupEntry sensor 路由用：subject_entity → entry（含 CardID）。
func (c *SpatialCache) LookupEntry(p netip.Prefix) *Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries[p]
}

// LookupCard 拿 card 自家属性（CardName / TenantPref）。
func (c *SpatialCache) LookupCard(cardID netip.Prefix) *CardMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cards[cardID]
}

// LookupRoom 拿 room 非空间属性（room_type / is_primary）；display 派生 priority 用。
func (c *SpatialCache) LookupRoom(roomID netip.Prefix) *RoomMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rooms[roomID]
}

// LookupUnit 拿 unit 非空间属性（unit_type / unit_property）。
func (c *SpatialCache) LookupUnit(unitID netip.Prefix) *UnitMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.units[unitID]
}

// LookupDevice 设备 UID / Type 用（OTA / alarm config）。
func (c *SpatialCache) LookupDevice(addr netip.Addr) *DeviceMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.devices[addr]
}

// EntryByCard 返回归属指定 card 的所有 entry（path 派生 display / debug 用）。
func (c *SpatialCache) EntryByCard(cardID netip.Prefix) []*Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []*Entry
	for _, e := range c.entries {
		if e.CardID == cardID {
			out = append(out, e)
		}
	}
	return out
}

// LookupCardByDevice — /128 device addr → 归属 card_id；未命中返零值 Prefix。
func (c *SpatialCache) LookupCardByDevice(addr netip.Addr) netip.Prefix {
	if !addr.IsValid() {
		return netip.Prefix{}
	}
	p := netip.PrefixFrom(addr, 128)
	c.mu.RLock()
	defer c.mu.RUnlock()
	if ent := c.entries[p]; ent != nil {
		return ent.CardID
	}
	return netip.Prefix{}
}

// DevicesByCard 返回归属指定 card 的所有 device meta（target_merger / bed_people_tracker 用）。
func (c *SpatialCache) DevicesByCard(cardID netip.Prefix) []*DeviceMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []*DeviceMeta
	for _, e := range c.entries {
		if e.CardID == cardID && e.IsDevice() {
			if dm := c.devices[e.Prefix.Addr()]; dm != nil {
				out = append(out, dm)
			}
		}
	}
	return out
}

// HasBed 卡 scope 下是否有 bed entry（按需现算，不存 snapshot）。
func (c *SpatialCache) HasBed(cardID netip.Prefix) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, e := range c.entries {
		if e.CardID == cardID && e.IsBed() {
			return true
		}
	}
	return false
}

// HasBedBoundRadar 卡 scope 下有 bed && 卡下任一 device 是 Radar（VisitorDeriver 激活条件）。
func (c *SpatialCache) HasBedBoundRadar(cardID netip.Prefix) bool {
	if !c.HasBed(cardID) {
		return false
	}
	for _, dm := range c.DevicesByCard(cardID) {
		if strings.Contains(strings.ToLower(dm.DeviceType), "radar") {
			return true
		}
	}
	return false
}

// ListBedCardsWithBedBoundRadar 返回卡 scope 下有 bed 且任一 device 是 Radar 的卡 ID。
// 全部按需现算 — 不存任何 snapshot。
func (c *SpatialCache) ListBedCardsWithBedBoundRadar(ctx context.Context) []string {
	_ = ctx
	c.mu.RLock()
	defer c.mu.RUnlock()
	// 先聚合每张卡是否有 bed / 有 radar
	hasBed := make(map[netip.Prefix]bool)
	hasRadar := make(map[netip.Prefix]bool)
	for _, e := range c.entries {
		switch {
		case e.IsBed():
			hasBed[e.CardID] = true
		case e.IsDevice():
			if dm := c.devices[e.Prefix.Addr()]; dm != nil &&
				strings.Contains(strings.ToLower(dm.DeviceType), "radar") {
				hasRadar[e.CardID] = true
			}
		}
	}
	var out []string
	for cardID := range c.cards {
		if hasBed[cardID] && hasRadar[cardID] {
			out = append(out, cardID.String())
		}
	}
	return out
}

// ListPrivateRoomCardIDs 返回所有父 unit_type=Private(1) 单元下的 /88 cards（visitor room-level 兜底）。
func (c *SpatialCache) ListPrivateRoomCardIDs(ctx context.Context) []string {
	_ = ctx
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []string
	for cardID := range c.cards {
		if cardID.Bits() != 88 {
			continue
		}
		unit := netip.PrefixFrom(cardID.Addr(), 80).Masked()
		if um := c.units[unit]; um != nil && um.UnitType == 1 {
			out = append(out, cardID.String())
		}
	}
	return out
}

// AllCardIDs cards 表枚举（display rebuilder 用）。
func (c *SpatialCache) AllCardIDs() []netip.Prefix {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]netip.Prefix, 0, len(c.cards))
	for id := range c.cards {
		out = append(out, id)
	}
	return out
}

// queryEntries 三表 union 拉所有 entry + 同步填 devices map（mask=128 项的非 spatial 属性）。
func (c *SpatialCache) queryEntries(ctx context.Context) (map[netip.Prefix]*Entry, map[netip.Addr]*DeviceMeta, error) {
	const q = `
		SELECT room_id::text AS prefix, COALESCE(room_name, '') AS name, card_id::text AS card_id, ''::text AS uid, ''::text AS dev_type
		  FROM rooms WHERE card_id IS NOT NULL
		UNION ALL
		SELECT b.bed_id::text,           COALESCE(b.bed_name, ''),         r.card_id::text,         '',           ''
		  FROM beds b JOIN rooms r ON b.bed_id <<= r.room_id
		 WHERE r.card_id IS NOT NULL
		UNION ALL
		SELECT d.device_addr::text,      '',                                d.card_id::text,         d.device_uid, COALESCE(dfm.device_type::text, '')
		  FROM devices d JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		 WHERE d.card_id IS NOT NULL
	`
	rows, err := c.db.QueryContext(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	entries := make(map[netip.Prefix]*Entry)
	devices := make(map[netip.Addr]*DeviceMeta)
	for rows.Next() {
		var prefixStr, name, cardIDStr, uid, devType string
		if err := rows.Scan(&prefixStr, &name, &cardIDStr, &uid, &devType); err != nil {
			continue
		}
		p, perr := netip.ParsePrefix(prefixStr)
		if perr != nil {
			continue
		}
		cardID, cerr := netip.ParsePrefix(cardIDStr)
		if cerr != nil {
			continue
		}
		ent := &Entry{Prefix: p, Name: name, CardID: cardID}
		entries[p] = ent

		if ent.IsDevice() {
			ent.Name = devType
			devices[p.Addr()] = &DeviceMeta{
				DeviceAddr: p.Addr(),
				DeviceUID:  uid,
				DeviceType: devType,
			}
		}
	}
	return entries, devices, nil
}

// queryCards 拉 cards 表自家字段（不再 JOIN units，不存 has_bed/has_bathroom snapshot）。
func (c *SpatialCache) queryCards(ctx context.Context) (map[netip.Prefix]*CardMeta, error) {
	rows, err := c.db.QueryContext(ctx, `SELECT card_id::text, COALESCE(card_name, '') FROM cards`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := make(map[netip.Prefix]*CardMeta)
	for rows.Next() {
		var cardIDStr, name string
		if err := rows.Scan(&cardIDStr, &name); err != nil {
			continue
		}
		cardID, cerr := netip.ParsePrefix(cardIDStr)
		if cerr != nil {
			continue
		}
		cards[cardID] = &CardMeta{
			CardID:     cardID,
			TenantPref: netip.PrefixFrom(cardID.Addr(), 48).Masked(),
			CardName:   name,
		}
	}
	return cards, nil
}

// queryRooms 拉 rooms 表 room 非空间原始字段（room_type / is_primary）。
func (c *SpatialCache) queryRooms(ctx context.Context) (map[netip.Prefix]*RoomMeta, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT room_id::text, COALESCE(room_name, ''), COALESCE(room_type, 0), COALESCE(is_primary, false)
		  FROM rooms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make(map[netip.Prefix]*RoomMeta)
	for rows.Next() {
		var pStr, name string
		var rt int
		var isPrimary bool
		if err := rows.Scan(&pStr, &name, &rt, &isPrimary); err != nil {
			continue
		}
		p, perr := netip.ParsePrefix(pStr)
		if perr != nil {
			continue
		}
		rooms[p] = &RoomMeta{Prefix: p, Name: name, RoomType: rt, IsPrimary: isPrimary}
	}
	return rooms, nil
}

// queryUnits 拉 units 表 unit 非空间原始字段 + spatial_effective_timezone view 拿 effective_tz。
func (c *SpatialCache) queryUnits(ctx context.Context) (map[netip.Prefix]*UnitMeta, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT u.unit_id::text,
		       COALESCE(u.unit_type, 0),
		       COALESCE(u.unit_property, 0),
		       COALESCE(et.effective_tz, '')
		  FROM units u
		  LEFT JOIN spatial_effective_timezone et ON et.unit_prefix = u.unit_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	units := make(map[netip.Prefix]*UnitMeta)
	for rows.Next() {
		var pStr, tz string
		var ut, up int
		if err := rows.Scan(&pStr, &ut, &up, &tz); err != nil {
			continue
		}
		p, perr := netip.ParsePrefix(pStr)
		if perr != nil {
			continue
		}
		units[p] = &UnitMeta{Prefix: p, UnitType: ut, UnitProperty: up, Timezone: tz}
	}
	return units, nil
}

// LookupTimezoneByCard cardID (any mask) → unit timezone *time.Location。
// 找不到或解析失败回退 UTC。
func (c *SpatialCache) LookupTimezoneByCard(cardID netip.Prefix) *time.Location {
	if !cardID.IsValid() {
		return time.UTC
	}
	unitPfx := cardID
	if cardID.Bits() > 80 {
		unitPfx = netip.PrefixFrom(cardID.Addr(), 80).Masked()
	}
	um := c.LookupUnit(unitPfx)
	if um == nil || um.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(um.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
