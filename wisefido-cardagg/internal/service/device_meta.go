package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"sync"

	"go.uber.org/zap"
	"owl-common/roomutil"
)

// DeviceMeta holds static metadata + runtime status for one device within a card.
// Static fields populated lazily from cards.devices JSONB on first access.
// RuntimeStatus updated by event stream (e.g. pressureSenSor init status).
type DeviceMeta struct {
	DeviceID      string // DB device_id (UUID)，仅 DB/告警使能等需要时用
	DeviceUID     string // 设备序列号，仅 log 等用；业务主 key 为 DeviceID
	DeviceType    string // "Radar", "SleepPad", etc.
	BoundBedID    string // may be empty
	BoundRoomID   string // may be empty
	BoundRoomName string // rooms.room_name, resolved via bound_room_id or bound_bed_id→beds→rooms
	// EffectiveRoomID BoundRoomID，若空则为 BoundBedID 解析出的 room_id；与 ClassifyRoomType(BoundRoomName) 一致使用。
	EffectiveRoomID string
	// BoundRoomHasBed 设备解析到的 room_id 在 tenant 内 beds 表中是否有记录；Stay 链仅对无床 room 开放（卫生间名除外）。
	BoundRoomHasBed bool
	UnitID          string
	RuntimeStatus map[string]string // event-driven device status (e.g. left_init_status)
}

// CardMeta holds all device metadata for one card.
type CardMeta struct {
	CardID   string
	TenantID string                 // from cards.tenant_id，供使能表等查询
	CardType string                 // "ActiveBedCard" | "UnitCard"
	BedID    string                 // card-level bed binding (from cards.bed_id)
	Devices  map[string]*DeviceMeta // device_id →（业务主 key，与 DB 一致）；仅 log 用 device_uid
	dbLoaded bool
}

func (m *CardMeta) IsActiveBedCard() bool {
	return m != nil && m.CardType == "ActiveBedCard"
}

func (m *CardMeta) IsUnitCard() bool {
	return m != nil && m.CardType == "UnitCard"
}

// BedBoundDeviceIDs 返回绑定到该卡床位的设备 device_id 列表（BoundBedID == CardMeta.BedID）。
func (m *CardMeta) BedBoundDeviceIDs() []string {
	if m == nil || m.BedID == "" || m.Devices == nil {
		return nil
	}
	var out []string
	for deviceID, dm := range m.Devices {
		if dm != nil && dm.BoundBedID == m.BedID {
			out = append(out, deviceID)
		}
	}
	return out
}

// DeviceMetaCache is a lazy-loading, card_change-invalidated in-memory cache.
//
// 反向索引（Phase A 启动时全量 load + cardChange 增量重建）：
//
//	deviceIDIndex  : device_id  → []cardID
//	deviceUIDIndex : device_uid → []cardID
//
// 设计动机：sensor agent (wisefido-sensor 等) 派生消息空 SubjectEntity 时，cardagg
// IotPreparedHandler 需 O(1) 反查 device→card；避免每条消息都 SQL 命中 cards.devices JSONB。
type DeviceMetaCache struct {
	mu     sync.RWMutex
	cards  map[string]*CardMeta // cardID →
	db     *sql.DB
	logger *zap.Logger

	idxMu          sync.RWMutex
	deviceIDIndex  map[string][]string // device_id  → []cardID
	deviceUIDIndex map[string][]string // device_uid → []cardID
}

func NewDeviceMetaCache(db *sql.DB, logger *zap.Logger) *DeviceMetaCache {
	return &DeviceMetaCache{
		cards:          make(map[string]*CardMeta),
		db:             db,
		logger:         logger,
		deviceIDIndex:  make(map[string][]string),
		deviceUIDIndex: make(map[string][]string),
	}
}

// Get returns cached CardMeta. Returns nil if not loaded yet.
func (c *DeviceMetaCache) Get(cardID string) *CardMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cards[cardID]
}

// GetOrLoad returns cached CardMeta, loading from DB on first access.
// If UpdateStatus created an empty shell before DB load, this merges DB data
// into the existing entry (preserving RuntimeStatus).
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
		existing.TenantID = dbMeta.TenantID
		existing.BedID = dbMeta.BedID
		existing.CardType = dbMeta.CardType
		existing.dbLoaded = true
		for devID, dbDev := range dbMeta.Devices {
			if exDev, ok := existing.Devices[devID]; ok {
				// Device exists from UpdateStatus — fill in static fields, keep RuntimeStatus
				exDev.DeviceType = dbDev.DeviceType
				exDev.BoundBedID = dbDev.BoundBedID
				exDev.BoundRoomID = dbDev.BoundRoomID
				exDev.BoundRoomName = dbDev.BoundRoomName
				exDev.BoundRoomHasBed = dbDev.BoundRoomHasBed
				exDev.UnitID = dbDev.UnitID
			} else {
				existing.Devices[devID] = dbDev
			}
		}
		return existing
	}

	c.cards[cardID] = dbMeta
	return dbMeta
}

// Invalidate marks a card for DB reload on next GetOrLoad,
// but preserves RuntimeStatus (event-driven state survives re-bind).
func (c *DeviceMetaCache) Invalidate(cardID string) {
	c.mu.Lock()
	if cm, ok := c.cards[cardID]; ok {
		cm.dbLoaded = false
	}
	c.mu.Unlock()
}

// Remove deletes a card entirely from cache (used when card is deleted).
func (c *DeviceMetaCache) Remove(cardID string) {
	c.mu.Lock()
	delete(c.cards, cardID)
	c.mu.Unlock()
}

// InvalidateAll clears the entire cache (used on reset).
func (c *DeviceMetaCache) InvalidateAll() {
	c.mu.Lock()
	c.cards = make(map[string]*CardMeta)
	c.mu.Unlock()
}

// InvalidateCardsInTenantUnit v2：把指定 unit (/80) 下所有 cards 的 meta 标为需重载。
//
// v2 unified: card_id ≡ spatial_prefix；unit_id 是 /80 INET CIDR 字符串（含 "/80"）。
// 用 prefix-match：cards.spatial_prefix <<= unit_prefix (/80)。
func (c *DeviceMetaCache) InvalidateCardsInTenantUnit(ctx context.Context, tenantID, unitID string) {
	if c.db == nil || unitID == "" {
		return
	}
	_ = tenantID // v2 不再用 tenant_id 列过滤；unitID prefix 已隐含 tenant /48
	prefix := unitID
	if !strings.Contains(prefix, "/") && strings.Contains(prefix, ":") {
		prefix = prefix + "/80"
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT c.spatial_prefix::text FROM cards c WHERE c.spatial_prefix <<= $1::INET`,
		prefix)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("InvalidateCardsInTenantUnit query failed",
				zap.String("unit_id", unitID), zap.Error(err))
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

// DeviceKeysInTenantUnit v2：返回 unit (/80) 下所有 device_id/device_uid（去重）。
//
// v2: cards.devices JSONB 已删；改用 devices + device_factory_meta JOIN by unit /80 prefix-match。
func DeviceKeysInTenantUnit(ctx context.Context, db *sql.DB, tenantID, unitID string) (deviceIDs []string, deviceUIDs []string) {
	if db == nil || unitID == "" {
		return nil, nil
	}
	_ = tenantID // v2 隐含在 unit prefix /48 内
	prefix := unitID
	if !strings.Contains(prefix, "/") && strings.Contains(prefix, ":") {
		prefix = prefix + "/80"
	}
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT dfm.device_id::text, dfm.device_uid
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE d.device_ipv6 <<= $1::INET`,
		prefix)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	seenD := make(map[string]struct{})
	seenU := make(map[string]struct{})
	for rows.Next() {
		var did, uid sql.NullString
		if err := rows.Scan(&did, &uid); err != nil {
			continue
		}
		if did.Valid && did.String != "" {
			if _, ok := seenD[did.String]; !ok {
				seenD[did.String] = struct{}{}
				deviceIDs = append(deviceIDs, did.String)
			}
		}
		if uid.Valid && uid.String != "" {
			if _, ok := seenU[uid.String]; !ok {
				seenU[uid.String] = struct{}{}
				deviceUIDs = append(deviceUIDs, uid.String)
			}
		}
	}
	return deviceIDs, deviceUIDs
}

// GetDeviceMeta returns metadata for a specific device within a card (key = device_id).
func (c *DeviceMetaCache) GetDeviceMeta(ctx context.Context, cardID, deviceID string) *DeviceMeta {
	cm := c.GetOrLoad(ctx, cardID)
	if cm == nil {
		return nil
	}
	return cm.Devices[deviceID]
}

// normalizeUIDKey 去空格、冒号、连字符、点并大写，用于序列号/MAC 与库内 device_uid 对齐。
func normalizeUIDKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToUpper(s) {
		if r == ':' || r == '-' || r == ' ' || r == '.' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// GetDeviceMetaByUID 仅当仅有 device_uid 时查 meta（如 log）；业务主 key 为 device_id。
func (c *DeviceMetaCache) GetDeviceMetaByUID(ctx context.Context, cardID, deviceUID string) *DeviceMeta {
	cm := c.GetOrLoad(ctx, cardID)
	if cm == nil || cm.Devices == nil {
		return nil
	}
	want := normalizeUIDKey(deviceUID)
	if want == "" {
		return nil
	}
	for _, dm := range cm.Devices {
		if dm != nil && normalizeUIDKey(dm.DeviceUID) == want {
			return dm
		}
	}
	return nil
}

// LookupCardsByDevice 反查一个设备绑定的所有 card_id（O(1) 内存索引；空时回退 SQL）。
//
// 数据源：内存反向索引（启动 BuildDeviceIndex 全量 load；cardChange 事件触发 RefreshDeviceIndexForCard 增量更新）。
// 索引未命中时 fallback 一次 SQL 直查 cards.devices JSONB（防止 cold start / 并发竞态时返回错的"未绑卡"结论）。
//
// 返回：cards 列表（绝大多数 1 张；多人共享设备时多张）；nil 表示未绑定任何卡。
//
// 用途：iot:event/alarm:stream 入口处理 envelope subject_entity="" 的消息（典型场景：
// wisefido-sensor 等 sensor agent 派生消息只带 device 标识，subject_entity（card）
// 由 cardagg 反查 fan-out）。这是协议层北极星 layer 1 (sensor) 与 layer 2/3
// 解耦的关键——AI 不染卡概念，cardagg 反查 device→subject 路由。
func (c *DeviceMetaCache) LookupCardsByDevice(ctx context.Context, deviceUID, deviceID string) []string {
	if deviceUID == "" && deviceID == "" {
		return nil
	}
	c.idxMu.RLock()
	hit := mergeUniqueCards(c.deviceIDIndex[deviceID], c.deviceUIDIndex[deviceUID])
	c.idxMu.RUnlock()
	if len(hit) > 0 {
		return hit
	}
	// 索引未命中（cold start / 新绑卡尚未 cardChange / 数据异常）→ v2 SQL 兜底
	if c.db == nil {
		return nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT DISTINCT c.spatial_prefix::text
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE ($1 <> '' AND dfm.device_uid = $1)
		   OR ($2 <> '' AND dfm.device_id::text = $2)
	`, deviceUID, deviceID)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("LookupCardsByDevice query failed",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
				zap.Error(err))
		}
		return nil
	}
	defer rows.Close()
	var cards []string
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err == nil && cid != "" {
			cards = append(cards, cid)
		}
	}
	return cards
}

// mergeUniqueCards 合并两个 cardID 列表，保持顺序去重（少量元素，遍历即可）。
func mergeUniqueCards(a, b []string) []string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, cid := range a {
		if cid == "" {
			continue
		}
		if _, ok := seen[cid]; ok {
			continue
		}
		seen[cid] = struct{}{}
		out = append(out, cid)
	}
	for _, cid := range b {
		if cid == "" {
			continue
		}
		if _, ok := seen[cid]; ok {
			continue
		}
		seen[cid] = struct{}{}
		out = append(out, cid)
	}
	return out
}

// BuildDeviceIndex v2：全量扫 devices + cards 派生 device→cards 反向索引（Phase A 启动）。
//
// v2 unified: card_id ≡ spatial_prefix（字符串 host CIDR）；
//   device → cards: cards.spatial_prefix >>= devices.device_ipv6 — 一个 device 可能挂多张父级 card
//   (例：bed /96 + unit /80 同时存在时都 contain 该 device_ipv6)
//
// 注：未应用 fillDevicesV3 的"唯一 bed 吸收"等业务归属规则；这里返回所有 prefix-contain 的 cards，
// caller（iot_prepared）只用该列表做路由广播，多挂几张不影响 — 主路由按 alarm_event.device_addr 走。
func (c *DeviceMetaCache) BuildDeviceIndex(ctx context.Context) error {
	if c.db == nil {
		return nil
	}
	rows, err := c.db.QueryContext(ctx, `
		SELECT c.spatial_prefix::text AS card_id,
		       dfm.device_id::text     AS device_id,
		       dfm.device_uid          AS device_uid
		FROM cards c
		JOIN devices d  ON d.device_ipv6 <<= c.spatial_prefix
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	idxByID := make(map[string][]string)
	idxByUID := make(map[string][]string)
	for rows.Next() {
		var cid, did, uid string
		if err := rows.Scan(&cid, &did, &uid); err != nil {
			continue
		}
		if cid == "" {
			continue
		}
		if did != "" {
			idxByID[did] = appendUnique(idxByID[did], cid)
		}
		if uid != "" {
			idxByUID[uid] = appendUnique(idxByUID[uid], cid)
		}
	}
	c.idxMu.Lock()
	c.deviceIDIndex = idxByID
	c.deviceUIDIndex = idxByUID
	c.idxMu.Unlock()
	if c.logger != nil {
		c.logger.Info("device index built",
			zap.Int("by_device_id", len(idxByID)),
			zap.Int("by_device_uid", len(idxByUID)))
	}
	return nil
}

// RefreshDeviceIndexForCard 重建单张 card 的索引条目（cardChange 事件触发）。
// 先剔除旧条目，再按当前 DB 行重新加入。card 已删除时仅做剔除。
func (c *DeviceMetaCache) RefreshDeviceIndexForCard(ctx context.Context, cardID string) {
	if cardID == "" {
		return
	}
	// 先剔除该 card 在两张索引中的所有条目
	c.idxMu.Lock()
	for k, v := range c.deviceIDIndex {
		c.deviceIDIndex[k] = removeFromSlice(v, cardID)
		if len(c.deviceIDIndex[k]) == 0 {
			delete(c.deviceIDIndex, k)
		}
	}
	for k, v := range c.deviceUIDIndex {
		c.deviceUIDIndex[k] = removeFromSlice(v, cardID)
		if len(c.deviceUIDIndex[k]) == 0 {
			delete(c.deviceUIDIndex, k)
		}
	}
	c.idxMu.Unlock()

	if c.db == nil {
		return
	}
	// v2 unified: card_id ≡ spatial_prefix；device 走 LPM 反查
	rows, err := c.db.QueryContext(ctx, `
		SELECT dfm.device_id::text, dfm.device_uid
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
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
		var did, uid string
		if err := rows.Scan(&did, &uid); err != nil {
			continue
		}
		if did != "" {
			c.deviceIDIndex[did] = appendUnique(c.deviceIDIndex[did], cardID)
		}
		if uid != "" {
			c.deviceUIDIndex[uid] = appendUnique(c.deviceUIDIndex[uid], cardID)
		}
	}
}

func appendUnique(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}

func removeFromSlice(s []string, v string) []string {
	out := s[:0]
	for _, x := range s {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

// ResolveDeviceID 仅返回业务主键 device_id（UUID）或空字符串。禁止将 device_uid 当作返回值；使能、告警、meta 查表均只用本函数结果。
// 1）卡已加载且能按 device_uid 命中 meta → 返回该条 device_id。
// 2）否则使用流 / IotPreparedHandler 已填的 device_id（未绑卡时常为 card_id 同 UUID）。
// 3）无法得到 UUID 时返回 ""（勿用 UID 顶替）。
func (c *DeviceMetaCache) ResolveDeviceID(ctx context.Context, cardID, deviceID, deviceUID string) string {
	if cardID != "" && deviceUID != "" {
		if dm := c.GetDeviceMetaByUID(ctx, cardID, deviceUID); dm != nil && dm.DeviceID != "" {
			return dm.DeviceID
		}
	}
	if cardID != "" && deviceID != "" && !IsUUID(deviceID) {
		if dm := c.GetDeviceMetaByUID(ctx, cardID, deviceID); dm != nil && dm.DeviceID != "" {
			return dm.DeviceID
		}
	}
	if deviceID != "" {
		return deviceID
	}
	return ""
}

// UpdateStatus sets a runtime status key for a specific device (key = device_id).
// Creates the card/device entry if needed (device may report before DB load).
func (c *DeviceMetaCache) UpdateStatus(cardID, deviceID, key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cm := c.cards[cardID]
	if cm == nil {
		cm = &CardMeta{CardID: cardID, Devices: make(map[string]*DeviceMeta)}
		c.cards[cardID] = cm
	}
	dm := cm.Devices[deviceID]
	if dm == nil {
		dm = &DeviceMeta{DeviceID: deviceID, RuntimeStatus: make(map[string]string)}
		cm.Devices[deviceID] = dm
	}
	if dm.RuntimeStatus == nil {
		dm.RuntimeStatus = make(map[string]string)
	}
	dm.RuntimeStatus[key] = value
}

type devicesJSONItem struct {
	DeviceID    string  `json:"device_id"`
	DeviceUID   string  `json:"device_uid"`
	DeviceType  string  `json:"device_type"`
	BedID       *string `json:"bed_id"`
	RoomID      *string `json:"room_id"`
	BoundBedID  *string `json:"bound_bed_id"`
	BoundRoomID *string `json:"bound_room_id"`
	UnitID      string  `json:"unit_id"`
}

func coalesceNonEmptyPtr(a, b *string) *string {
	if a != nil && *a != "" {
		return a
	}
	if b != nil && *b != "" {
		return b
	}
	if a != nil {
		return a
	}
	return b
}

// IsCardID 判断是否为 v2 card_id 格式（INET CIDR with mask，e.g. "fd00:0:3:111:3:301::/96"）
// 未绑卡时 cardagg 会用 deviceKey 充 card_id（非 CIDR），不应查库。
func IsCardID(s string) bool {
	// v2 card_id ≡ spatial_prefix；最简单识别：包含 "::" 或 ":" + "/" mask 后缀
	return len(s) >= 9 && strings.Contains(s, "/") && strings.Contains(s, ":")
}

// IsUUID 历史 alias：现转发到 IsCardID。保留是为了平滑 caller 现有引用（不引入 rename 风暴）。
// Deprecated: 用 IsCardID。
func IsUUID(s string) bool { return IsCardID(s) }

func (c *DeviceMetaCache) loadFromDB(ctx context.Context, cardID string) *CardMeta {
	if c.db == nil {
		return nil
	}
	if !IsUUID(cardID) {
		return nil
	}

	// v2 unified: card_id ≡ spatial_prefix；tenant 由 prefix /48 派生；bed/cardType 同理
	var spatialPrefix, cardType sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT spatial_prefix::text, card_type
		  FROM cards
		 WHERE spatial_prefix = $1::INET
	`, cardID).Scan(&spatialPrefix, &cardType)
	if err != nil {
		if err != sql.ErrNoRows {
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
		// 派生 tenant prefix /48
		if i := strings.Index(spatialPrefix.String, "/"); i > 0 {
			meta.TenantID = spatialPrefix.String[:i] // 简化：完整 prefix 作 TenantID 占位
		}
		// 派生 bed_id = /96 spatial（masklen 96 时）
		meta.BedID = spatialPrefix.String
	}
	if cardType.Valid {
		meta.CardType = cardType.String
	}

	// v2: cards.devices JSONB 已删，devices 走 LPM 反查
	devicesJSON := sql.NullString{Valid: false}
	if !devicesJSON.Valid || devicesJSON.String == "" || devicesJSON.String == "[]" {
		return meta
	}

	var items []devicesJSONItem
	if err := json.Unmarshal([]byte(devicesJSON.String), &items); err != nil {
		c.logger.Warn("parse devices json", zap.String("cid", cardID), zap.Error(err))
		return meta
	}

	var roomIDs, bedIDs []string
	for _, item := range items {
		uid := item.DeviceUID
		if uid == "" {
			uid = item.DeviceID
		}
		key := item.DeviceID
		if key == "" {
			key = uid
		}
		dm := &DeviceMeta{
			DeviceID:      item.DeviceID,
			DeviceUID:     uid,
			DeviceType:    item.DeviceType,
			UnitID:        item.UnitID,
			RuntimeStatus: make(map[string]string),
		}
		bedPtr := coalesceNonEmptyPtr(item.BedID, item.BoundBedID)
		roomPtr := coalesceNonEmptyPtr(item.RoomID, item.BoundRoomID)
		if bedPtr != nil {
			dm.BoundBedID = *bedPtr
			bedIDs = append(bedIDs, *bedPtr)
		}
		if roomPtr != nil {
			dm.BoundRoomID = *roomPtr
			roomIDs = append(roomIDs, *roomPtr)
		}
		meta.Devices[key] = dm
	}

	// 批量解析 room_name：bound_room_id 直接查 rooms，bound_bed_id 走 beds→rooms
	roomNames := c.resolveRoomNames(ctx, roomIDs, bedIDs)
	bedToRoom := c.resolveBedToRoom(ctx, meta.TenantID, bedIDs)
	roomIDSet := make(map[string]struct{})
	for _, dm := range meta.Devices {
		if dm.BoundRoomID != "" {
			roomIDSet[dm.BoundRoomID] = struct{}{}
		}
		if dm.BoundBedID != "" {
			if rid := bedToRoom[dm.BoundBedID]; rid != "" {
				roomIDSet[rid] = struct{}{}
			}
		}
	}
	uniqRooms := make([]string, 0, len(roomIDSet))
	for rid := range roomIDSet {
		uniqRooms = append(uniqRooms, rid)
	}
	roomsWithBeds := c.resolveRoomsWithBeds(ctx, meta.TenantID, uniqRooms)
	for _, dm := range meta.Devices {
		if dm.BoundRoomID != "" {
			dm.BoundRoomName = roomNames[dm.BoundRoomID]
		} else if dm.BoundBedID != "" {
			dm.BoundRoomName = roomNames[dm.BoundBedID]
		}
		effectiveRoom := dm.BoundRoomID
		if effectiveRoom == "" {
			effectiveRoom = bedToRoom[dm.BoundBedID]
		}
		dm.EffectiveRoomID = effectiveRoom
		dm.BoundRoomHasBed = roomsWithBeds[effectiveRoom]
	}

	c.logger.Debug("loaded device meta",
		zap.String("cid", cardID),
		zap.Int("devices", len(meta.Devices)))

	return meta
}

// resolveRoomNames 批量查询 room_name。
// 返回 map: room_id→room_name 和 bed_id→room_name。
func (c *DeviceMetaCache) resolveRoomNames(ctx context.Context, roomIDs, bedIDs []string) map[string]string {
	result := make(map[string]string)
	if len(roomIDs) == 0 && len(bedIDs) == 0 {
		return result
	}

	// bound_room_id → rooms.room_name
	if len(roomIDs) > 0 {
		rows, err := c.db.QueryContext(ctx,
			`SELECT room_id::text, room_name FROM rooms WHERE room_id = ANY($1)`,
			pqUUIDs(roomIDs))
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, name string
				if rows.Scan(&id, &name) == nil {
					result[id] = name
				}
			}
		}
	}

	// bound_bed_id → beds.room_id → rooms.room_name
	if len(bedIDs) > 0 {
		rows, err := c.db.QueryContext(ctx,
			`SELECT b.bed_id::text, r.room_name
			 FROM beds b JOIN rooms r ON b.room_id = r.room_id
			 WHERE b.bed_id = ANY($1)`,
			pqUUIDs(bedIDs))
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, name string
				if rows.Scan(&id, &name) == nil {
					result[id] = name
				}
			}
		}
	}

	return result
}

// resolveBedToRoom bed_id → room_id（tenant 内）。
func (c *DeviceMetaCache) resolveBedToRoom(ctx context.Context, tenantID string, bedIDs []string) map[string]string {
	result := make(map[string]string)
	if c.db == nil || tenantID == "" || len(bedIDs) == 0 {
		return result
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT b.bed_id::text, b.room_id::text FROM beds b WHERE b.tenant_id = $1 AND b.bed_id = ANY($2)`,
		tenantID, pqUUIDs(bedIDs))
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("resolve bed to room", zap.Error(err))
		}
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var bid, rid string
		if rows.Scan(&bid, &rid) == nil && bid != "" && rid != "" {
			result[bid] = rid
		}
	}
	return result
}

// resolveRoomsWithBeds 返回 tenant 下该 room_id 是否至少存在一张床。
func (c *DeviceMetaCache) resolveRoomsWithBeds(ctx context.Context, tenantID string, roomIDs []string) map[string]bool {
	out := make(map[string]bool)
	if c.db == nil || tenantID == "" || len(roomIDs) == 0 {
		return out
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT DISTINCT b.room_id::text FROM beds b WHERE b.tenant_id = $1 AND b.room_id = ANY($2)`,
		tenantID, pqUUIDs(roomIDs))
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("resolve rooms with beds", zap.Error(err))
		}
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var rid string
		if rows.Scan(&rid) == nil && rid != "" {
			out[rid] = true
		}
	}
	return out
}

// pqUUIDs 将 string slice 转为 pq.Array 兼容的格式。
// PostgreSQL ANY($1) 需要 array 参数，使用 "{id1,id2,...}" 格式。
func pqUUIDs(ids []string) string {
	return "{" + strings.Join(ids, ",") + "}"
}

// ClassifyRoomType 根据 room_name 推断房间语义类型。
// 实现已迁移到 owl-common/roomutil 供多服务复用；保留本 wrapper 是为了不动本包 caller。
func ClassifyRoomType(roomName string) string {
	return roomutil.ClassifyRoomType(roomName)
}

// CardHasRadarAndSleepadOnBed 是否该卡片在指定 bed 上同时绑定了雷达和睡眠垫。
func CardHasRadarAndSleepadOnBed(meta *CardMeta, bedID string) bool {
	if meta == nil || meta.Devices == nil || bedID == "" {
		return false
	}
	var hasRadar, hasSleepad bool
	for _, dm := range meta.Devices {
		if dm == nil || dm.BoundBedID != bedID {
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
