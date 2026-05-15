package service

import (
	"context"
	"database/sql"
	"net/netip"
	"strings"
	"sync"

	"go.uber.org/zap"
	"owl-common/roomutil"
)

// device_ipv6 单程票（doc/device_ipv6_migration_checklist.md）落地后：
//
//   - device 业务身份唯一为 device_addr (IPv6 /128)
//   - cardID = cards.spatial_prefix (INET CIDR text，e.g. "fd00:0:3:111:3:101::/96")
//   - 反向索引由原 deviceIDIndex/deviceUIDIndex 双索引合并为 deviceAddrIndex 单索引
//   - DeviceMeta 内部不再保留 DeviceID/DeviceUID 字段；空间归属字段 (UnitID/RoomID/BedID)
//     从 DeviceAddr.Prefix(80/88/96) 派生，accessor 提供
//   - IsUUID / IsCardID / NormalizeUIDKey / GetDeviceMetaByUID / ResolveDeviceID 全部删除（R-001）

// DeviceMeta holds static metadata + runtime status for one device within a card.
//
// 字段语义在 device_ipv6 单程票后（保留字段名最小化下游 ripple；值全部 IPv6 派生）：
//
//	DeviceID/DeviceUID  = device_addr canonical IPv6 文本（不再 UUID/MAC）
//	BoundBedID          = device_addr.Prefix(96).String() (CIDR text)
//	BoundRoomID         = device_addr.Prefix(88).String()
//	UnitID              = device_addr.Prefix(80).String()
//	EffectiveRoomID     = BoundRoomID（v2 不再有"派生 room"概念）
type DeviceMeta struct {
	DeviceAddr      netip.Addr        // 设备唯一身份 /128（新字段，主 key）
	DeviceID        string            // 兼容老 caller：== DeviceAddr.String()
	DeviceUID       string            // 兼容老 caller：== DeviceAddr.String()
	DeviceType      string            // "Radar", "SleepPad", etc.
	BoundBedID      string            // /96 CIDR text，派生自 DeviceAddr
	BoundRoomID     string            // /88 CIDR text
	UnitID          string            // /80 CIDR text
	BoundRoomName   string            // rooms.room_name，DB 反查
	EffectiveRoomID string            // == BoundRoomID
	BoundRoomHasBed bool              // 该 room 是否有床（Stay 链门控）
	RuntimeStatus   map[string]string // event-driven device status (e.g. left_init_status)
}

// fillFromAddr 从 DeviceAddr 派生所有字符串字段（构造期 + loadFromDB 调）。
func (d *DeviceMeta) fillFromAddr() {
	if d == nil || !d.DeviceAddr.IsValid() {
		return
	}
	d.DeviceID = d.DeviceAddr.String()
	d.DeviceUID = d.DeviceID
	d.BoundBedID = netip.PrefixFrom(d.DeviceAddr, 96).Masked().String()
	d.BoundRoomID = netip.PrefixFrom(d.DeviceAddr, 88).Masked().String()
	d.UnitID = netip.PrefixFrom(d.DeviceAddr, 80).Masked().String()
	d.EffectiveRoomID = d.BoundRoomID
}

// UnitPref / RoomPref / BedPref 派生空间前缀（INET CIDR text）— 新代码可直接用 BoundBedID 等同效字段。
func (d *DeviceMeta) UnitPref() string { return d.UnitID }
func (d *DeviceMeta) RoomPref() string { return d.BoundRoomID }
func (d *DeviceMeta) BedPref() string  { return d.BoundBedID }

// AddrStr 返回 canonical IPv6 文本（map key 用）。
func (d *DeviceMeta) AddrStr() string  { return d.DeviceID }

// CardMeta holds all device metadata for one card.
//
// CardID = cards.spatial_prefix INET CIDR；TenantID/TenantPref = /48 prefix CIDR；BedID/BedPref = /96 (active_bed) 或同 CardID。
//
// TenantID/BedID 兼容字段保留，值同 TenantPref/BedPref；老 caller 不必 rename。
type CardMeta struct {
	CardID     string                 // INET CIDR e.g. "fd00:0:3:111:3:101::/96"
	TenantID   string                 // 兼容字段：== TenantPref
	TenantPref string                 // /48 CIDR e.g. "fd00:0:3::/48"
	CardType   string                 // "active_bed" | "unit" | "room" | "public" | ...
	BedID      string                 // 兼容字段：== BedPref
	BedPref    string                 // 当 CardID 是 /96 时 == CardID；否则空
	Devices    map[string]*DeviceMeta // key = device_addr canonical IPv6 string
	dbLoaded   bool
}

func (m *CardMeta) IsActiveBedCard() bool { return m != nil && m.CardType == "active_bed" }
func (m *CardMeta) IsUnitCard() bool      { return m != nil && m.CardType == "unit" }

// BedBoundDeviceAddrs 返回绑定到该卡床位（/96）下的设备 addr 列表。
func (m *CardMeta) BedBoundDeviceAddrs() []string {
	if m == nil || m.BedPref == "" || m.Devices == nil {
		return nil
	}
	bedNet, err := netip.ParsePrefix(m.BedPref)
	if err != nil {
		return nil
	}
	var out []string
	for addrStr, dm := range m.Devices {
		if dm == nil || !dm.DeviceAddr.IsValid() {
			continue
		}
		if bedNet.Contains(dm.DeviceAddr) {
			out = append(out, addrStr)
		}
	}
	return out
}

// BedBoundDeviceIDs 兼容 alias — 返回 device_addr 字符串列表（== BedBoundDeviceAddrs）。
func (m *CardMeta) BedBoundDeviceIDs() []string { return m.BedBoundDeviceAddrs() }

// DeviceMetaCache is a lazy-loading, card_change-invalidated in-memory cache.
//
// 反向索引（启动 BuildDeviceIndex 全量 load + cardChange 事件触发增量重建）：
//
//	deviceAddrIndex: device_addr (canonical IPv6) → cardID
//
// 因 cards.spatial_prefix LPM 命中唯一，per device 至多 1 cardID（不再 fan-out）。
type DeviceMetaCache struct {
	mu     sync.RWMutex
	cards  map[string]*CardMeta // cardID →
	db     *sql.DB
	logger *zap.Logger

	idxMu           sync.RWMutex
	deviceAddrIndex map[string]string // device_addr → cardID
}

func NewDeviceMetaCache(db *sql.DB, logger *zap.Logger) *DeviceMetaCache {
	return &DeviceMetaCache{
		cards:           make(map[string]*CardMeta),
		db:              db,
		logger:          logger,
		deviceAddrIndex: make(map[string]string),
	}
}

// Get returns cached CardMeta. Returns nil if not loaded yet.
func (c *DeviceMetaCache) Get(cardID string) *CardMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cards[cardID]
}

// GetOrLoad returns cached CardMeta, loading from DB on first access.
func (c *DeviceMetaCache) GetOrLoad(ctx context.Context, cardID string) *CardMeta {
	c.mu.RLock()
	if m, ok := c.cards[cardID]; ok && m.dbLoaded {
		c.mu.RUnlock()
		return m
	}
	c.mu.RUnlock()

	dbMeta := c.loadFromDB(ctx, cardID)

	c.mu.Lock()
	defer c.mu.Unlock()

	existing := c.cards[cardID]
	if dbMeta == nil {
		return existing
	}

	if existing != nil && !existing.dbLoaded {
		// Merge: DB static fields into existing shell, preserve RuntimeStatus
		existing.TenantPref = dbMeta.TenantPref
		existing.TenantID = dbMeta.TenantPref
		existing.BedPref = dbMeta.BedPref
		existing.BedID = dbMeta.BedPref
		existing.CardType = dbMeta.CardType
		existing.dbLoaded = true
		for addr, dbDev := range dbMeta.Devices {
			if exDev, ok := existing.Devices[addr]; ok {
				exDev.DeviceType = dbDev.DeviceType
				exDev.BoundRoomName = dbDev.BoundRoomName
				exDev.BoundRoomHasBed = dbDev.BoundRoomHasBed
			} else {
				existing.Devices[addr] = dbDev
			}
		}
		return existing
	}

	c.cards[cardID] = dbMeta
	return dbMeta
}

// Invalidate marks a card for DB reload on next GetOrLoad,
// but preserves RuntimeStatus.
func (c *DeviceMetaCache) Invalidate(cardID string) {
	c.mu.Lock()
	if cm, ok := c.cards[cardID]; ok {
		cm.dbLoaded = false
	}
	c.mu.Unlock()
}

func (c *DeviceMetaCache) Remove(cardID string) {
	c.mu.Lock()
	delete(c.cards, cardID)
	c.mu.Unlock()
}

func (c *DeviceMetaCache) InvalidateAll() {
	c.mu.Lock()
	c.cards = make(map[string]*CardMeta)
	c.mu.Unlock()
}

// InvalidateCardsInTenantUnit 把指定 unit (/80) 下所有 cards 的 meta 标为需重载。
func (c *DeviceMetaCache) InvalidateCardsInTenantUnit(ctx context.Context, unitPref string) {
	if c.db == nil || unitPref == "" {
		return
	}
	prefix := unitPref
	if !strings.Contains(prefix, "/") && strings.Contains(prefix, ":") {
		prefix = prefix + "/80"
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT c.spatial_prefix::text FROM cards c WHERE c.spatial_prefix <<= $1::INET`,
		prefix)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("InvalidateCardsInTenantUnit query failed",
				zap.String("unit_pref", unitPref), zap.Error(err))
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil || cid == "" {
			continue
		}
		c.Invalidate(cid)
	}
}

// DeviceAddrsInUnit 返回 unit (/80) 下所有 device_addr 字符串（去重）。
func DeviceAddrsInUnit(ctx context.Context, db *sql.DB, unitPref string) []string {
	if db == nil || unitPref == "" {
		return nil
	}
	prefix := unitPref
	if !strings.Contains(prefix, "/") && strings.Contains(prefix, ":") {
		prefix = prefix + "/80"
	}
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT host(d.device_ipv6)::text
		FROM devices d
		WHERE d.device_ipv6 <<= $1::INET`,
		prefix)
	if err != nil {
		return nil
	}
	defer rows.Close()
	seen := make(map[string]struct{})
	out := []string{}
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil || addr == "" {
			continue
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	return out
}

// GetDeviceMeta returns metadata for a specific device within a card (key = device_addr canonical IPv6 string).
func (c *DeviceMetaCache) GetDeviceMeta(ctx context.Context, cardID, deviceAddr string) *DeviceMeta {
	cm := c.GetOrLoad(ctx, cardID)
	if cm == nil {
		return nil
	}
	return cm.Devices[deviceAddr]
}

// GetDeviceMetaByUID 兼容老 caller — 等价 GetDeviceMeta（单程票后 DeviceID == DeviceUID == addr canonical IPv6）。
func (c *DeviceMetaCache) GetDeviceMetaByUID(ctx context.Context, cardID, deviceUID string) *DeviceMeta {
	return c.GetDeviceMeta(ctx, cardID, deviceUID)
}

// ResolveDeviceID 兼容老 caller — 单程票后 deviceID/deviceUID 都是 addr canonical IPv6；返回第一个非空。
// 不再做 UUID/UID 双路径反查（R-001）；caller 应直接传 addr.String()。
func (c *DeviceMetaCache) ResolveDeviceID(ctx context.Context, cardID, deviceID, deviceUID string) string {
	if deviceID != "" {
		return deviceID
	}
	return deviceUID
}

// LookupCardByDeviceAddr 反查 device_addr → cardID（O(1) 内存索引；空时回退 SQL LPM）。
//
// 命中返回单一 cardID（cards.spatial_prefix INET CIDR text）；
// 未命中（unbound device，R-009）返回空字符串，caller 走"无卡 device 状态"分支。
//
// 与原 LookupCardsByDevice (返回 []string) 不同：device_ipv6 LPM 命中唯一，per device
// 至多 1 cardID（业务模型 1:1 绑定 + IPv6 LPM 自然唯一性）。
func (c *DeviceMetaCache) LookupCardByDeviceAddr(ctx context.Context, addr netip.Addr) string {
	if !addr.IsValid() {
		return ""
	}
	addrStr := addr.String()
	c.idxMu.RLock()
	cardID := c.deviceAddrIndex[addrStr]
	c.idxMu.RUnlock()
	if cardID != "" {
		return cardID
	}
	// 索引未命中 → SQL LPM 兜底
	if c.db == nil {
		return ""
	}
	var cid sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT c.spatial_prefix::text
		  FROM cards c
		 WHERE c.spatial_prefix >>= $1::INET
		 ORDER BY masklen(c.spatial_prefix) DESC
		 LIMIT 1
	`, addrStr).Scan(&cid)
	if err != nil {
		if err != sql.ErrNoRows && c.logger != nil {
			c.logger.Warn("LookupCardByDeviceAddr LPM failed",
				zap.String("device_addr", addrStr), zap.Error(err))
		}
		return ""
	}
	if !cid.Valid {
		return ""
	}
	return cid.String
}

// BuildDeviceIndex 全量扫 devices + cards 派生 device_addr → cardID 反向索引（启动期）。
//
// device_ipv6 LPM 命中唯一 → per addr 单 cardID。
// 多张父级 card 同时 contain 该 addr 时（罕见，例如 /96 + /80 共存），取最深 prefix。
func (c *DeviceMetaCache) BuildDeviceIndex(ctx context.Context) error {
	if c.db == nil {
		return nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT host(d.device_ipv6)::text AS device_addr,
		       c.spatial_prefix::text     AS card_id
		FROM devices d
		JOIN LATERAL (
		  SELECT spatial_prefix
		    FROM cards
		   WHERE spatial_prefix >>= d.device_ipv6
		   ORDER BY masklen(spatial_prefix) DESC
		   LIMIT 1
		) c ON TRUE
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	idx := make(map[string]string)
	for rows.Next() {
		var addr, cid string
		if err := rows.Scan(&addr, &cid); err != nil {
			continue
		}
		if addr == "" || cid == "" {
			continue
		}
		idx[addr] = cid
	}
	c.idxMu.Lock()
	c.deviceAddrIndex = idx
	c.idxMu.Unlock()
	if c.logger != nil {
		c.logger.Info("device addr index built", zap.Int("count", len(idx)))
	}
	return nil
}

// RefreshDeviceIndexForCard 重建单张 card 的索引条目（cardChange 事件触发）。
// 先剔除该 card 在索引中所有 addr 条目，再按当前 DB 行重新加入（card 已删时仅剔除）。
func (c *DeviceMetaCache) RefreshDeviceIndexForCard(ctx context.Context, cardID string) {
	if cardID == "" {
		return
	}
	c.idxMu.Lock()
	for addr, cid := range c.deviceAddrIndex {
		if cid == cardID {
			delete(c.deviceAddrIndex, addr)
		}
	}
	c.idxMu.Unlock()

	if c.db == nil {
		return
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT host(d.device_ipv6)::text
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		WHERE c.spatial_prefix = $1::INET
	`, cardID)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("refresh device index query failed", zap.String("card_id", cardID), zap.Error(err))
		}
		return
	}
	defer rows.Close()
	c.idxMu.Lock()
	defer c.idxMu.Unlock()
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil || addr == "" {
			continue
		}
		c.deviceAddrIndex[addr] = cardID
	}
}

// IsCardID 判断是否为 v2 card_id 格式（INET CIDR with mask）。
func IsCardID(s string) bool {
	return len(s) >= 9 && strings.Contains(s, "/") && strings.Contains(s, ":")
}

func (c *DeviceMetaCache) loadFromDB(ctx context.Context, cardID string) *CardMeta {
	if c.db == nil {
		return nil
	}
	if !IsCardID(cardID) {
		return nil
	}

	// v2 unified: card_id ≡ spatial_prefix；从 cards 取 cardType，spatial_prefix 派生 tenant /48 + bed /96
	var spatialPrefix, cardType sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT spatial_prefix::text, card_type
		  FROM cards
		 WHERE spatial_prefix = $1::INET
	`, cardID).Scan(&spatialPrefix, &cardType)
	if err != nil {
		if err != sql.ErrNoRows && c.logger != nil {
			c.logger.Warn("load device meta", zap.String("cid", cardID), zap.Error(err))
		}
		return nil
	}

	meta := &CardMeta{
		CardID:   cardID,
		Devices:  make(map[string]*DeviceMeta),
		dbLoaded: true,
	}
	if spatialPrefix.Valid {
		// 派生 tenant /48 prefix；bed_pref 当 mask=96 时 == cardID
		if pref, perr := netip.ParsePrefix(spatialPrefix.String); perr == nil {
			meta.TenantPref = netip.PrefixFrom(pref.Addr(), 48).Masked().String()
			meta.TenantID = meta.TenantPref
			if pref.Bits() == 96 {
				meta.BedPref = spatialPrefix.String
				meta.BedID = spatialPrefix.String
			}
		}
	}
	if cardType.Valid {
		meta.CardType = cardType.String
	}

	// 加载本卡 LPM-拥有的 device → DeviceMeta（addr 为 key）。
	// v3 owning rule: 设备归 LPM 最长的覆盖卡；UnitCard (/80) 不含被 /96 bed card 拥有的 sleepad。
	// 之前 SQL "d.device_ipv6 <<= c.spatial_prefix" 让 UnitCard.Devices 含所有 /80 下设备
	// 包括下层 bed card 拥有的 sleepad → hasBedCapableDevice 误判 → bed_state 反复 init。
	rows, err := c.db.QueryContext(ctx, `
		SELECT host(d.device_ipv6)::text,
		       COALESCE(dfm.device_type::text, '')
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE (
		  SELECT spatial_prefix FROM cards
		  WHERE spatial_prefix >>= d.device_ipv6
		  ORDER BY masklen(spatial_prefix) DESC
		  LIMIT 1
		) = $1::INET
	`, cardID)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("load devices for card", zap.String("cid", cardID), zap.Error(err))
		}
		return meta
	}
	defer rows.Close()

	for rows.Next() {
		var addrStr, devType string
		if err := rows.Scan(&addrStr, &devType); err != nil {
			continue
		}
		addr, perr := netip.ParseAddr(addrStr)
		if perr != nil {
			continue
		}
		dm := &DeviceMeta{
			DeviceAddr:    addr,
			DeviceType:    devType,
			RuntimeStatus: make(map[string]string),
		}
		dm.fillFromAddr()
		meta.Devices[addrStr] = dm
	}

	// 批量解析 BoundRoomName + BoundRoomHasBed（按 device → /88 room prefix）
	c.fillRoomMetadata(ctx, meta)

	if c.logger != nil {
		c.logger.Debug("loaded device meta",
			zap.String("cid", cardID),
			zap.Int("devices", len(meta.Devices)))
	}
	return meta
}

// fillRoomMetadata 按每 device 的 /88 room prefix 反查 rooms.room_name + 该 room 是否有床。
// v2 schema: rooms.room_id INET (/88 CIDR)；beds.room_id 同 INET。
func (c *DeviceMetaCache) fillRoomMetadata(ctx context.Context, meta *CardMeta) {
	if c.db == nil || meta == nil || len(meta.Devices) == 0 {
		return
	}
	roomSet := make(map[string]struct{})
	for _, dm := range meta.Devices {
		if dm == nil {
			continue
		}
		if rp := dm.RoomPref(); rp != "" {
			roomSet[rp] = struct{}{}
		}
	}
	if len(roomSet) == 0 {
		return
	}
	roomList := make([]string, 0, len(roomSet))
	for rp := range roomSet {
		roomList = append(roomList, rp)
	}

	roomNames := make(map[string]string)
	roomHasBed := make(map[string]bool)

	// rooms.room_name
	for _, rp := range roomList {
		var name sql.NullString
		_ = c.db.QueryRowContext(ctx,
			`SELECT room_name FROM rooms WHERE room_id = $1::INET LIMIT 1`,
			rp).Scan(&name)
		if name.Valid {
			roomNames[rp] = name.String
		}
	}
	// 至少一张床
	for _, rp := range roomList {
		var cnt int
		_ = c.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM beds WHERE bed_id <<= $1::INET`,
			rp).Scan(&cnt)
		roomHasBed[rp] = cnt > 0
	}

	for _, dm := range meta.Devices {
		if dm == nil {
			continue
		}
		rp := dm.RoomPref()
		dm.BoundRoomName = roomNames[rp]
		dm.BoundRoomHasBed = roomHasBed[rp]
	}
}

// UpdateStatus sets a runtime status key for a specific device (key = device_addr canonical IPv6 string).
// Creates the card/device entry if needed (device may report before DB load).
func (c *DeviceMetaCache) UpdateStatus(cardID, deviceAddr, key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cm := c.cards[cardID]
	if cm == nil {
		cm = &CardMeta{CardID: cardID, Devices: make(map[string]*DeviceMeta)}
		c.cards[cardID] = cm
	}
	dm := cm.Devices[deviceAddr]
	if dm == nil {
		addr, _ := netip.ParseAddr(deviceAddr)
		dm = &DeviceMeta{DeviceAddr: addr, RuntimeStatus: make(map[string]string)}
		dm.fillFromAddr()
		cm.Devices[deviceAddr] = dm
	}
	if dm.RuntimeStatus == nil {
		dm.RuntimeStatus = make(map[string]string)
	}
	dm.RuntimeStatus[key] = value
}

// ClassifyRoomType 根据 room_name 推断房间语义类型。
func ClassifyRoomType(roomName string) string {
	return roomutil.ClassifyRoomType(roomName)
}

// CardHasRadarAndSleepadOnBed 是否该卡片在指定 bed (/96 CIDR) 上同时绑定了雷达和睡眠垫。
func CardHasRadarAndSleepadOnBed(meta *CardMeta, bedPref string) bool {
	if meta == nil || meta.Devices == nil || bedPref == "" {
		return false
	}
	bedNet, err := netip.ParsePrefix(bedPref)
	if err != nil {
		return false
	}
	var hasRadar, hasSleepad bool
	for _, dm := range meta.Devices {
		if dm == nil || !dm.DeviceAddr.IsValid() || !bedNet.Contains(dm.DeviceAddr) {
			continue
		}
		t := strings.ToLower(dm.DeviceType)
		if strings.Contains(t, "radar") {
			hasRadar = true
		}
		if strings.Contains(t, "sleepad") || strings.Contains(t, "sleeppad") {
			hasSleepad = true
		}
		if hasRadar && hasSleepad {
			return true
		}
	}
	return false
}
