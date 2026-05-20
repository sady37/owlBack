// card_reconcile.go — Card 表唯一权威写入入口（v2 收敛）。
//
// 调用方（业务事件触发）：
//   - device on/off（新增/绑定/解绑）
//   - resident create/update/delete/transfer
//   - 启动时全量 reconcile（一 tenant 一次）
//
// 责任：让 cards 表与 (devices, residents) 当前真相对齐 + Redis CloudEvent 通知下游
// （cardagg / sensor / device-gateway 各订配置流），失败重试由下次 reconcile 兜底。
//
// Split 规则（user 拍板 2026-05-17）：
//
//	Unit 层：
//	  bed > 1   → split: 进 Room 层 + (有 device 上推 unit 时) 1 张 /80
//	  bed ≤ 1   → merge: 仅 1 张 /80 unit card (room + bed device 全装这一张)
//
//	Room 层（仅 Unit split 时进入）— 按 (bedN, hasRoomDevice) 二维：
//	  bed > 1 + hasRoomDev → /88 room + N /96 bed
//	  bed > 1, no roomDev  → N /96 bed only       (room 内无设备 LPM 到 /88)
//	  bed = 1 + hasRoomDev → /88 room (absorb bed)
//	  bed = 1, no roomDev  → /96 bed (absorb room)
//	  bed = 0 + hasRoomDev → 不建，device 上推 unit
//	  bed = 0, no roomDev  → 跳过
//
//	Unit card 存在条件（split 模式下）：
//	  有 device 上推 unit (bed=0+hasRoomDev room 或 unit-level /80 anchor) → 出 unit
//	  否则 → 不建 unit card (empty unit card 必删，Step 3 DELETE 兜底)
//
// IPv6 LPM 设计要点：/88 mask 仅检高字节，:300: 与 :301: 同高字节 0x03 同落 /88；
// 故 bed-level :301: 设备能 LPM 命中 /88 room card。
// 若 /96 bed card 也存在（bedN>1），longer prefix wins，bed 设备精确落 /96。

package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"
	"strings"

	"owl-common/card"
	"owl-common/ddns"

	"go.uber.org/zap"
)

// ReconcileCards 入口；scope=INET CIDR (/48 tenant / /80 unit 任意级)。
func (s *CardSyncService) ReconcileCards(ctx context.Context, scope string) error {
	if s.db == nil {
		s.logger.Debug("ReconcileCards skipped (db not wired)")
		return nil
	}
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return fmt.Errorf("scope required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	expected, err := buildExpected(ctx, tx, scope)
	if err != nil {
		return err
	}
	current, err := loadCurrent(ctx, tx, scope)
	if err != nil {
		return err
	}
	diffs, err := s.applyDiffs(ctx, tx, expected, current)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	s.logger.Info("ReconcileCards done",
		zap.String("scope", scope),
		zap.Int("expected", len(expected)),
		zap.Int("current_before", len(current)),
		zap.Int("diffs", len(diffs)))

	if s.reconcileObserver != nil {
		s.reconcileObserver(scope, diffs)
	}
	s.emitDiffs(ctx, scope, diffs)
	s.ensureDDNSForExpected(ctx, scope, expected)
	return nil
}

// ============================================================================
// Step 1 — 算 expected
// ============================================================================

type unitInfo struct {
	bedAnchors    map[string]bool // /96 anchors
	nonBedAnchors map[string]bool // /80 or /88 anchors
}

// buildExpected 查 device 反推 unit anchors，per unit 应用 split 规则。
func buildExpected(ctx context.Context, tx *sql.Tx, scope string) (map[string]bool, error) {
	units, err := queryUnitAnchors(ctx, tx, scope)
	if err != nil {
		return nil, err
	}
	expected := map[string]bool{}
	for unitPrefix, info := range units {
		applyUnitSplitRule(unitPrefix, info, expected)
	}
	return expected, nil
}

// queryUnitAnchors scope 内每 monitor-on device 取 (unit_prefix, anchor_prefix)；
// anchor 优先 bed > room > unit。
func queryUnitAnchors(ctx context.Context, tx *sql.Tx, scope string) (map[string]*unitInfo, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT DISTINCT
		  (SELECT u.unit_id::text FROM units u WHERE d.device_ipv6 <<= u.unit_id LIMIT 1) AS unit_prefix,
		  COALESCE(
		    (SELECT b.bed_id::text  FROM beds  b  WHERE d.device_ipv6 <<= b.bed_id  LIMIT 1),
		    (SELECT rm.room_id::text FROM rooms rm WHERE d.device_ipv6 <<= rm.room_id LIMIT 1),
		    (SELECT u.unit_id::text FROM units u  WHERE d.device_ipv6 <<= u.unit_id  LIMIT 1)
		  ) AS anchor_prefix
		  FROM devices d
		 WHERE d.monitoring_enabled = TRUE
		   AND d.device_ipv6 <<= $1::INET
	`, scope)
	if err != nil {
		return nil, fmt.Errorf("query device anchors: %w", err)
	}
	defer rows.Close()

	units := map[string]*unitInfo{}
	for rows.Next() {
		var unitPrefix, anchor sql.NullString
		if err := rows.Scan(&unitPrefix, &anchor); err != nil {
			return nil, fmt.Errorf("scan device anchors: %w", err)
		}
		if !unitPrefix.Valid || unitPrefix.String == "" || !anchor.Valid || anchor.String == "" {
			continue
		}
		u, ok := units[unitPrefix.String]
		if !ok {
			u = &unitInfo{bedAnchors: map[string]bool{}, nonBedAnchors: map[string]bool{}}
			units[unitPrefix.String] = u
		}
		if strings.HasSuffix(anchor.String, "/96") {
			u.bedAnchors[anchor.String] = true
		} else {
			u.nonBedAnchors[anchor.String] = true
		}
	}
	return units, rows.Err()
}

// applyUnitSplitRule 把一个 unit 的 split 决策结果写入 expected。
func applyUnitSplitRule(unitPrefix string, info *unitInfo, expected map[string]bool) {
	bedCount := len(info.bedAnchors)
	hasNonBed := len(info.nonBedAnchors) > 0
	if bedCount == 0 && !hasNonBed {
		return
	}

	// Merge mode
	if bedCount <= 1 {
		expected[unitPrefix] = true
		return
	}

	// Split mode — 算 per-room bed count + per-room 是否有 room-level device
	roomBeds := map[string]map[string]bool{} // room /88 → set of bed /96 prefixes
	for bed := range info.bedAnchors {
		if room := narrowPrefixToRoom(bed); room != "" {
			if roomBeds[room] == nil {
				roomBeds[room] = map[string]bool{}
			}
			roomBeds[room][bed] = true
		}
	}
	roomHasDev := map[string]bool{}
	for anchor := range info.nonBedAnchors {
		if strings.HasSuffix(anchor, "/88") {
			roomHasDev[anchor] = true
		}
	}

	// 集合所有需决策的 room
	rooms := map[string]bool{}
	for r := range roomBeds {
		rooms[r] = true
	}
	for r := range roomHasDev {
		rooms[r] = true
	}

	needUnit := false
	for room := range rooms {
		beds := roomBeds[room]
		bedN := len(beds)
		hasDev := roomHasDev[room]

		switch {
		case bedN > 1 && hasDev:
			// /88 room + 每 bed /96
			expected[room] = true
			for b := range beds {
				expected[b] = true
			}
		case bedN > 1 && !hasDev:
			// 仅 N /96
			for b := range beds {
				expected[b] = true
			}
		case bedN == 1 && hasDev:
			// /88 room (absorb bed)
			expected[room] = true
		case bedN == 1 && !hasDev:
			// /96 bed (absorb room)
			for b := range beds {
				expected[b] = true
			}
		case bedN == 0 && hasDev:
			// 独立 /88 room 卡（split 模式下，0-bed-with-device room 不上推 unit，
			// 保持卡片粒度与 split 模式一致 —— 否则 /80 与 /96 混杂语义混乱）
			expected[room] = true
		}
	}

	// /80 unit-level device → unit card 必须
	for anchor := range info.nonBedAnchors {
		if !strings.HasSuffix(anchor, "/88") {
			needUnit = true
		}
	}

	if needUnit {
		expected[unitPrefix] = true
	}
	// else: split 模式下无 device 上推 → 不建 unit card；applyDiffs DELETE 兜底清旧 empty unit。
}

// ============================================================================
// Step 2 — 加载 current
// ============================================================================

type currentCard struct {
	residentID string
	cardName   string
}

func loadCurrent(ctx context.Context, tx *sql.Tx, scope string) (map[string]currentCard, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT spatial_prefix::text, COALESCE(host(resident_id), ''), COALESCE(card_name, '')
		  FROM cards
		 WHERE spatial_prefix <<= $1::INET
	`, scope)
	if err != nil {
		return nil, fmt.Errorf("query current cards: %w", err)
	}
	defer rows.Close()
	out := map[string]currentCard{}
	for rows.Next() {
		var p, rid, name string
		if err := rows.Scan(&p, &rid, &name); err != nil {
			return nil, fmt.Errorf("scan current card: %w", err)
		}
		out[p] = currentCard{residentID: rid, cardName: name}
	}
	return out, rows.Err()
}

// ============================================================================
// Step 3 — DELETE 多余 + UPSERT 缺失/变更
// ============================================================================

func (s *CardSyncService) applyDiffs(ctx context.Context, tx *sql.Tx, expected map[string]bool, current map[string]currentCard) ([]cardDiff, error) {
	diffs := []cardDiff{}

	// DELETE: current 里 expected 没有的卡
	for p, cur := range current {
		if expected[p] {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM cards WHERE spatial_prefix = $1::INET`, p); err != nil {
			return nil, fmt.Errorf("delete card %s: %w", p, err)
		}
		diffs = append(diffs, cardDiff{prefix: p, op: "delete", prevHoA: cur.residentID})
	}

	// UPSERT: expected 每个 prefix → 算 resident_id + card_name + dns_short → INSERT/UPDATE
	for p := range expected {
		diff, err := s.upsertCard(ctx, tx, p, current)
		if err != nil {
			return nil, err
		}
		if diff != nil {
			diffs = append(diffs, *diff)
		}
	}
	return diffs, nil
}

// upsertCard 单 card 的 INSERT/UPDATE，返回 diff 用于 emit CloudEvent。
//
// resident_id 解析两阶段 LPM：
//
//	阶段 1: LPM 双向 overlap（card prefix ↔ resident_unit.spatial_prefix 任一包含另一）
//	阶段 2: unit fallback（card 所在 /80 unit 内唯一 active resident，handle Private unit 整 unit 归一人）
//
// card_name 来源：
//
//	public unit (unit_type=3) → 固定 "public" 标签（占位 resident nickname 太长无意义）
//	其他       → residents.nickname / CardNameNoResident
func (s *CardSyncService) upsertCard(ctx context.Context, tx *sql.Tx, p string, current map[string]currentCard) (*cardDiff, error) {
	cardType, err := cardTypeForPrefix(p)
	if err != nil {
		s.logger.Warn("upsertCard: skip prefix (unknown card_type)",
			zap.String("prefix", p), zap.Error(err))
		return nil, nil
	}

	residentID, err := resolveResidentForCard(ctx, tx, p)
	if err != nil {
		return nil, err
	}

	cardName := CardNameNoResident
	newHoA := ""
	var ridArg interface{} = nil

	isPublic, err := isPublicUnit(ctx, tx, p)
	if err != nil {
		return nil, err
	}

	if isPublic {
		cardName = CardNamePublic
		if residentID != "" {
			ridArg = residentID
			newHoA = residentID
		}
	} else if residentID != "" {
		nick, err := lookupResidentNick(ctx, tx, residentID)
		if err != nil {
			return nil, err
		}
		if nick != "" {
			cardName = nick
		}
		ridArg = residentID
		newHoA = residentID
	}

	shortName := ""
	if parsed, perr := netip.ParsePrefix(p); perr == nil {
		if sn, snerr := ddns.CardShortName(parsed); snerr == nil {
			shortName = sn
		}
	}

	// 算 diff（在 INSERT 前算，否则 ON CONFLICT 后看不到差异）
	diff := computeCardDiff(p, current[p], newHoA, cardName)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO cards (spatial_prefix, card_type, card_name, dns_short_name, resident_id, is_active, enabled_at)
		VALUES ($1::INET, $2, $3, NULLIF($4,''), $5::INET, TRUE, NOW())
		ON CONFLICT (spatial_prefix) DO UPDATE SET
		    card_name      = EXCLUDED.card_name,
		    dns_short_name = COALESCE(EXCLUDED.dns_short_name, cards.dns_short_name),
		    resident_id    = EXCLUDED.resident_id,
		    is_active      = TRUE,
		    updated_at     = NOW()
		WHERE cards.card_name      IS DISTINCT FROM EXCLUDED.card_name
		   OR cards.dns_short_name IS DISTINCT FROM EXCLUDED.dns_short_name
		   OR cards.resident_id    IS DISTINCT FROM EXCLUDED.resident_id
		   OR cards.is_active      IS DISTINCT FROM TRUE
	`, p, cardType, cardName, shortName, ridArg); err != nil {
		return nil, fmt.Errorf("upsert card %s: %w", p, err)
	}

	return diff, nil
}

// resolveResidentForCard 两阶段 LPM 找 resident_id（host 格式 / "" 表无 resident）
func resolveResidentForCard(ctx context.Context, tx *sql.Tx, prefix string) (string, error) {
	var rid sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(
		  -- 阶段 1: LPM 双向 overlap
		  (SELECT host(ru.resident_id)
		     FROM resident_unit ru
		    WHERE ru.valid_to IS NULL
		      AND ($1::INET <<= ru.spatial_prefix OR ru.spatial_prefix <<= $1::INET)
		      AND masklen(ru.spatial_prefix) >= 80
		    ORDER BY masklen(ru.spatial_prefix) DESC
		    LIMIT 1),
		  -- 阶段 2: unit fallback — 该 /80 unit 内唯一 active resident
		  (SELECT host(ru.resident_id)
		     FROM resident_unit ru
		    WHERE ru.valid_to IS NULL
		      AND ru.spatial_prefix <<= network(set_masklen($1::INET, 80))
		      AND masklen(ru.spatial_prefix) >= 80
		      AND (SELECT COUNT(DISTINCT ru2.resident_id)
		             FROM resident_unit ru2
		            WHERE ru2.valid_to IS NULL
		              AND ru2.spatial_prefix <<= network(set_masklen($1::INET, 80))
		              AND masklen(ru2.spatial_prefix) >= 80) = 1
		    LIMIT 1)
		)
	`, prefix).Scan(&rid)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("lookup resident for %s: %w", prefix, err)
	}
	if rid.Valid {
		return rid.String, nil
	}
	return "", nil
}

// isPublicUnit 判 /80 unit 是否 unit_type=3（公共区）— 仅 /80 prefix 适用。
func isPublicUnit(ctx context.Context, tx *sql.Tx, prefix string) (bool, error) {
	if !strings.HasSuffix(prefix, "/80") {
		return false, nil
	}
	var ut sql.NullInt32
	err := tx.QueryRowContext(ctx,
		`SELECT unit_type FROM units WHERE unit_id = $1::INET`, prefix).Scan(&ut)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return ut.Valid && ut.Int32 == 3, nil
}

func lookupResidentNick(ctx context.Context, tx *sql.Tx, residentHostID string) (string, error) {
	var n sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(nickname, '') FROM residents WHERE host(resident_id) = $1`,
		residentHostID).Scan(&n)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if n.Valid {
		return n.String, nil
	}
	return "", nil
}

func computeCardDiff(prefix string, prev currentCard, newHoA, newName string) *cardDiff {
	if prev.residentID == "" && prev.cardName == "" {
		// 不在 current 集合 = 新卡
		if newHoA != "" {
			return &cardDiff{prefix: prefix, op: "admission", newHoA: newHoA}
		}
		return &cardDiff{prefix: prefix, op: "create", newHoA: newHoA}
	}
	if prev.residentID != newHoA {
		switch {
		case prev.residentID == "":
			return &cardDiff{prefix: prefix, op: "admission", newHoA: newHoA}
		case newHoA == "":
			return &cardDiff{prefix: prefix, op: "discharge", prevHoA: prev.residentID}
		default:
			return &cardDiff{prefix: prefix, op: "transfer", prevHoA: prev.residentID, newHoA: newHoA}
		}
	}
	if prev.cardName != newName {
		return &cardDiff{prefix: prefix, op: "name_changed", prevHoA: newHoA, newHoA: newHoA}
	}
	return nil
}

// ============================================================================
// CloudEvent emit + DDNS sync
// ============================================================================

// cardDiff — ReconcileCards 算出的单卡变化记录，commit 后用于 emit CloudEvent。
type cardDiff struct {
	prefix  string
	op      string // create / delete / admission / discharge / transfer / name_changed
	prevHoA string
	newHoA  string
}

// SetReconcileObserver — 测试 hook，每次 ReconcileCards commit 后调一次。生产留 nil。
func (s *CardSyncService) SetReconcileObserver(fn func(scope string, diffs []cardDiff)) {
	s.reconcileObserver = fn
}

// emitDiffs commit 后 1) bind9 DDNS sync (create/delete) 2) Redis CloudEvent 通知下游。
// 失败仅 log warn（cards 表已 commit，bind9/Redis 失败下次 reconcile 兜底）。
func (s *CardSyncService) emitDiffs(ctx context.Context, scope string, diffs []cardDiff) {
	if len(diffs) == 0 {
		return
	}
	tenantPrefix := tenantPrefixOf(scope)
	tenantSlot := tenantSlotOf(scope)
	for _, d := range diffs {
		if s.ddns != nil && tenantSlot > 0 {
			s.syncDDNSForDiff(ctx, d, tenantSlot)
		}
		if s.publisher != nil {
			s.publishDiff(ctx, tenantPrefix, d)
		}
	}
}

func (s *CardSyncService) syncDDNSForDiff(ctx context.Context, d cardDiff, tenantSlot uint16) {
	parsed, err := netip.ParsePrefix(d.prefix)
	if err != nil {
		return
	}
	shortName := card.ShortCodeOf(d.prefix)
	zone := ddns.ZoneForTenant(tenantSlot, s.owlDomain)
	switch d.op {
	case "create":
		if err := s.ddns.RegisterCardName(ctx, parsed, shortName, zone); err != nil {
			s.logger.Warn("DDNS register",
				zap.String("fqdn", shortName+"."+zone), zap.Error(err))
		}
	case "delete":
		if err := s.ddns.UnregisterCardName(ctx, parsed, shortName, zone); err != nil {
			s.logger.Warn("DDNS unregister",
				zap.String("fqdn", shortName+"."+zone), zap.Error(err))
		}
	}
}

func (s *CardSyncService) publishDiff(ctx context.Context, tenantPrefix string, d cardDiff) {
	var err error
	switch d.op {
	case "create", "delete":
		err = s.publisher.PublishCardChanged(ctx, tenantPrefix, d.prefix, d.op, d.prefix, "")
	case "admission", "discharge", "transfer", "name_changed":
		err = s.publisher.PublishCardResidentChanged(ctx, tenantPrefix, d.prefix, d.op, d.prevHoA, d.newHoA, d.prefix)
	}
	if err != nil {
		s.logger.Warn("publish card diff",
			zap.String("prefix", d.prefix), zap.String("op", d.op), zap.Error(err))
	}
}

// ensureDDNSForExpected 对当前 expected 集合每个 prefix ensure DNS 记录（idempotent）。
// 跟 emitDiffs 区别：emitDiffs 只 create/delete 触发；本函数对全 expected force-sync。
func (s *CardSyncService) ensureDDNSForExpected(ctx context.Context, scope string, expected map[string]bool) {
	if s.ddns == nil || len(expected) == 0 {
		return
	}
	tenantSlot := tenantSlotOf(scope)
	if tenantSlot == 0 {
		return
	}
	zone := ddns.ZoneForTenant(tenantSlot, s.owlDomain)
	for p := range expected {
		parsed, err := netip.ParsePrefix(p)
		if err != nil {
			continue
		}
		shortName := card.ShortCodeOf(p)
		if err := s.ddns.RegisterCardName(ctx, parsed, shortName, zone); err != nil {
			s.logger.Warn("ensureDDNS",
				zap.String("fqdn", shortName+"."+zone),
				zap.String("prefix", p), zap.Error(err))
		}
	}
}

// ============================================================================
// 小工具：prefix 解析
// ============================================================================

// tenantSlotOf 从 INET CIDR 推 tenant slot — "fd00:0:3:..." → 3
func tenantSlotOf(prefix string) uint16 {
	idx := strings.LastIndex(prefix, "/")
	if idx <= 0 {
		return 0
	}
	parts := strings.Split(strings.Split(prefix[:idx], "::")[0], ":")
	if len(parts) < 3 {
		return 0
	}
	v, err := parseHex(parts[2])
	if err != nil {
		return 0
	}
	return uint16(v)
}

// tenantPrefixOf 任意 INET CIDR 截到 /48 — "fd00:0:3:112:3::/80" → "fd00:0:3::/48"
func tenantPrefixOf(prefix string) string {
	idx := strings.LastIndex(prefix, "/")
	if idx <= 0 {
		return prefix
	}
	parts := strings.Split(strings.Split(prefix[:idx], "::")[0], ":")
	if len(parts) < 3 {
		return prefix
	}
	return strings.Join(parts[:3], ":") + "::/48"
}

// narrowPrefixToRoom /96 bed → /88 room；其他原样空串。
// IPv6 第 6 段 (uint16) = room_slot(高 8 bit) << 8 | bed_slot(低 8 bit)
// /96 lock 整 16 bit；/88 lock 高 8 bit (room)，低 8 bit 清零。
//
// 例: "fd00:0:3:111:3:101::/96" → "fd00:0:3:111:3:100::/88"
func narrowPrefixToRoom(prefix string) string {
	if !strings.HasSuffix(prefix, "/96") {
		return ""
	}
	idx := strings.LastIndex(prefix, "/")
	if idx <= 0 {
		return ""
	}
	parts := strings.Split(strings.Split(prefix[:idx], "::")[0], ":")
	if len(parts) < 6 {
		return ""
	}
	seg, err := parseHex(parts[5])
	if err != nil {
		return ""
	}
	roomSeg := (seg >> 8) << 8
	parts[5] = fmt.Sprintf("%x", roomSeg)
	return strings.Join(parts[:6], ":") + "::/88"
}

// cardTypeForPrefix 按 masklen 判 card_type；不支持的 masklen 返错。
func cardTypeForPrefix(prefix string) (string, error) {
	switch {
	case strings.HasSuffix(prefix, "/96"):
		return "bed", nil
	case strings.HasSuffix(prefix, "/88"):
		return "room", nil
	case strings.HasSuffix(prefix, "/80"):
		return "unit", nil
	}
	return "", fmt.Errorf("unsupported masklen for card_type: %s", prefix)
}

// parseHex 解析单段 hex，避开 strconv 引入。
func parseHex(s string) (uint64, error) {
	var v uint64
	for _, ch := range s {
		var d uint64
		switch {
		case ch >= '0' && ch <= '9':
			d = uint64(ch - '0')
		case ch >= 'a' && ch <= 'f':
			d = uint64(ch-'a') + 10
		case ch >= 'A' && ch <= 'F':
			d = uint64(ch-'A') + 10
		default:
			return 0, fmt.Errorf("invalid hex char: %c", ch)
		}
		v = v*16 + d
	}
	return v, nil
}
