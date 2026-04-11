package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"sync"

	"go.uber.org/zap"
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
type DeviceMetaCache struct {
	mu     sync.RWMutex
	cards  map[string]*CardMeta // cardID →
	db     *sql.DB
	logger *zap.Logger
}

func NewDeviceMetaCache(db *sql.DB, logger *zap.Logger) *DeviceMetaCache {
	return &DeviceMetaCache{
		cards:  make(map[string]*CardMeta),
		db:     db,
		logger: logger,
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

// InvalidateCardsInTenantUnit 将 tenant+unit 下所有卡片的 meta 标为需重载（config:card 同 unit 内多卡联动）。
func (c *DeviceMetaCache) InvalidateCardsInTenantUnit(ctx context.Context, tenantID, unitID string) {
	if c.db == nil || tenantID == "" || unitID == "" {
		return
	}
	rows, err := c.db.QueryContext(ctx,
		`SELECT card_id::text FROM cards WHERE tenant_id = $1::uuid AND unit_id = $2::uuid`,
		tenantID, unitID)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("InvalidateCardsInTenantUnit query failed",
				zap.String("tenant_id", tenantID), zap.String("unit_id", unitID), zap.Error(err))
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

// DeviceKeysInTenantUnit 返回该 unit 下所有卡 devices JSON 中的 device_id / device_uid（去重），供告警使能与 resolver 键失效。
func DeviceKeysInTenantUnit(ctx context.Context, db *sql.DB, tenantID, unitID string) (deviceIDs []string, deviceUIDs []string) {
	if db == nil || tenantID == "" || unitID == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT j->>'device_id', j->>'device_uid'
		FROM cards c, jsonb_array_elements(COALESCE(c.devices, '[]'::jsonb)) AS j
		WHERE c.tenant_id = $1::uuid AND c.unit_id = $2::uuid
		  AND COALESCE(j->>'device_id', '') <> ''`,
		tenantID, unitID)
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

// IsUUID 判断是否为 UUID 格式（cards.card_id 为 UUID，未绑卡时会用 deviceKey 充 card_id，不应查库）
func IsUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	return s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
}

func (c *DeviceMetaCache) loadFromDB(ctx context.Context, cardID string) *CardMeta {
	if c.db == nil {
		return nil
	}
	if !IsUUID(cardID) {
		return nil
	}

	var devicesJSON sql.NullString
	var bedID sql.NullString
	var cardType sql.NullString
	var tenantID sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT devices, bed_id, card_type, tenant_id FROM cards WHERE card_id = $1
	`, cardID).Scan(&devicesJSON, &bedID, &cardType, &tenantID)
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
	if tenantID.Valid {
		meta.TenantID = tenantID.String
	}
	if bedID.Valid {
		meta.BedID = bedID.String
	}
	if cardType.Valid {
		meta.CardType = cardType.String
	}

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

// ClassifyRoomType 根据 room_name 推断 AreaType。
//
//	wc/bathroom/restroom/toilet → "bathroom"
//	bedroom/bed room            → "bedroom"
//	kitchen                     → "kitchen"
//	其余                         → "other"
func ClassifyRoomType(roomName string) string {
	lower := strings.ToLower(roomName)
	switch {
	case strings.Contains(lower, "wc"),
		strings.Contains(lower, "bathroom"),
		strings.Contains(lower, "restroom"),
		strings.Contains(lower, "toilet"):
		return "bathroom"
	case strings.Contains(lower, "bedroom"),
		strings.Contains(lower, "bed room"):
		return "bedroom"
	case strings.Contains(lower, "kitchen"):
		return "kitchen"
	default:
		return "other"
	}
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
