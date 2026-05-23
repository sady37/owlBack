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
	"sort"
	"strings"

	"github.com/lib/pq"

	"owl-common/card"
	"owl-common/ddns"

	"go.uber.org/zap"
)

// ReconcileCards 入口；scope=INET CIDR (/48 tenant / /80 unit 任意级)。
//
// 流程（原 v2 split 规则保留 — user 2026-05-21 确认未变）：
//   1. buildExpected: 用 device → anchor 反推 + per-unit split rule v3 算 expected anchor set
//   2. loadCurrent:    读 cards 表当前内容
//   3. applyDiffs:     DELETE 多余 / UPSERT 新增或变更
//
// v2.5 schema 适配：upsertCard / applyDiffs 内部已切到新 INSERT 字段（card_id /88 slot + card_slot
// + unit_id + has_bed snapshot）；anchor → card_id slot 的映射在 upsertCard 内分配并 store。
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

// buildExpected — 纯空间结构驱动算 expected card 集合（rule.md C9 / 2026-05-23 终版）。
//
// 规则：
//
//	N = bedCount_in_unit          (整 unit 床数 /96 候选)
//	M = noBedRoomCount_in_unit    (不含 bed 的 room 数，含 0-device 空 room)
//
//	Step 0: unit 无 room → 无卡
//	Step 1: Unit
//	  N ≤ 1 → /80 only (merge，吸所有)
//	  N > 1 → split：M > 0 时预创建 /80（装 noBed room device）
//	Step 2: 每个含 bed 的 room (split 模式)
//	  Room.N ≤ 1 → /88 卡 (吸 bed + room device)
//	  Room.N > 1 → N 张 /96 + room device 上推 /80（M=0 时 lazy create /80）
//
// device / alarm 路由：PG GiST `<<=` LPM 自动 — device <<= 现存 /96 落 /96；<<= 现存 /88 落 /88；否则 /80。
// alarm.producer 列保留发起设备身份。
func buildExpected(ctx context.Context, tx *sql.Tx, scope string) (map[string]bool, error) {
	expected := map[string]bool{}

	// 一次扫 scope 内所有 unit，LEFT JOIN rooms 拿到每 (unit, room) 配对 + 整 unit 的 room/bed count
	// + per-room bed count、beds 列表、是否有非 bed device（用于 split 模式 lazy /80 触发）
	rows, err := tx.QueryContext(ctx, `
		SELECT
		    host(u.unit_id)||'/'||masklen(u.unit_id) AS unit_prefix,
		    (SELECT COUNT(*) FROM beds  WHERE bed_id  <<= u.unit_id) AS unit_bed_count,
		    (SELECT COUNT(*) FROM rooms WHERE room_id <<= u.unit_id) AS unit_room_count,
		    CASE WHEN r.room_id IS NULL THEN NULL
		         ELSE host(r.room_id)||'/'||masklen(r.room_id)
		    END AS room_prefix,
		    COALESCE((SELECT COUNT(*) FROM beds WHERE bed_id <<= r.room_id), 0) AS room_bed_count,
		    COALESCE((
		        SELECT array_agg(host(b.bed_id)||'/'||masklen(b.bed_id))
		          FROM beds b WHERE b.bed_id <<= r.room_id
		    ), ARRAY[]::text[]) AS room_beds,
		    COALESCE((
		        SELECT EXISTS(
		            SELECT 1 FROM devices d
		             WHERE d.device_addr <<= r.room_id
		               AND NOT EXISTS(
		                 SELECT 1 FROM beds b
		                  WHERE b.bed_id <<= r.room_id
		                    AND d.device_addr <<= b.bed_id
		               )
		        )
		    ), FALSE) AS room_has_nonbed_device
		  FROM units u
		  LEFT JOIN rooms r ON r.room_id <<= u.unit_id
		 WHERE u.unit_id <<= $1::INET
		 ORDER BY u.unit_id
	`, scope)
	if err != nil {
		return nil, fmt.Errorf("query unit structure: %w", err)
	}
	defer rows.Close()

	type roomData struct {
		prefix       string
		bedCount     int
		beds         []string
		hasNonBedDev bool
	}
	type unitData struct {
		prefix    string
		bedCount  int
		roomCount int
		rooms     []roomData
	}
	units := map[string]*unitData{}

	for rows.Next() {
		var unitPrefix string
		var unitBedCount, unitRoomCount, roomBedCount int
		var roomPrefix sql.NullString
		var roomBeds pq.StringArray
		var roomHasNonBedDev bool
		if err := rows.Scan(&unitPrefix, &unitBedCount, &unitRoomCount, &roomPrefix, &roomBedCount, &roomBeds, &roomHasNonBedDev); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		u, ok := units[unitPrefix]
		if !ok {
			u = &unitData{prefix: unitPrefix, bedCount: unitBedCount, roomCount: unitRoomCount}
			units[unitPrefix] = u
		}
		if roomPrefix.Valid {
			u.rooms = append(u.rooms, roomData{
				prefix:       roomPrefix.String,
				bedCount:     roomBedCount,
				beds:         []string(roomBeds),
				hasNonBedDev: roomHasNonBedDev,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iter rows: %w", err)
	}

	for _, u := range units {
		// Step 0: unit 无 room → 无卡
		if u.roomCount == 0 {
			continue
		}
		// Step 1: Unit
		if u.bedCount <= 1 {
			// merge: /80 only
			expected[u.prefix] = true
			continue
		}
		// split mode (N > 1)
		// M = bed-less room count
		m := 0
		for _, r := range u.rooms {
			if r.bedCount == 0 {
				m++
			}
		}
		has80 := m > 0 // M>0 → 预创建 /80

		// Step 2: per room with bed
		for _, r := range u.rooms {
			if r.bedCount == 0 {
				continue // bed-less room device 经 LPM 自动落 /80（has80 已 by M>0 触发）
			}
			if r.bedCount <= 1 {
				// /88 leaf 吸 bed + room device
				expected[r.prefix] = true
			} else {
				// Room.N > 1 → split：N 张 /96 + room device → /80
				for _, bed := range r.beds {
					expected[bed] = true
				}
				if r.hasNonBedDev {
					has80 = true // lazy create /80（M=0 时也兜底）
				}
			}
		}
		if has80 {
			expected[u.prefix] = true
		}
	}
	return expected, nil
}

// ============================================================================
// Step 2 — 加载 current
// ============================================================================

type currentCard struct {
	residentID  string
	cardName    string
	hasBed      bool
	hasBathroom bool
	hasKitchen  bool
}

// loadCurrent map key = card_id CIDR text（== anchor，原 doc 模型）
func loadCurrent(ctx context.Context, tx *sql.Tx, scope string) (map[string]currentCard, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT host(card_id)||'/'||masklen(card_id),
		       COALESCE(host(resident_id), ''),
		       COALESCE(card_name, ''),
		       has_bed, has_bathroom, has_kitchen
		  FROM cards
		 WHERE card_id <<= $1::INET
	`, scope)
	if err != nil {
		return nil, fmt.Errorf("query current cards: %w", err)
	}
	defer rows.Close()
	out := map[string]currentCard{}
	for rows.Next() {
		var p, rid, name string
		var hasBed, hasBath, hasKit bool
		if err := rows.Scan(&p, &rid, &name, &hasBed, &hasBath, &hasKit); err != nil {
			return nil, fmt.Errorf("scan current card: %w", err)
		}
		out[p] = currentCard{
			residentID: rid, cardName: name,
			hasBed: hasBed, hasBathroom: hasBath, hasKitchen: hasKit,
		}
	}
	return out, rows.Err()
}

// ============================================================================
// Step 3 — DELETE 多余 + UPSERT 缺失/变更
// ============================================================================

// applyDiffs 原 doc 方案：card_id ≡ anchor prefix（无 slot 分配）
//   1. DELETE current 里不在 expected 的 card
//   2. UPSERT each anchor: card_id = anchor 直接，has_bed 由 anchor mask 派生
//   3. 子表 FK: 按 anchor mask DESC 最具体优先挂 rooms.card_id / devices.card_id
func (s *CardSyncService) applyDiffs(ctx context.Context, tx *sql.Tx, expected map[string]bool, current map[string]currentCard) ([]cardDiff, error) {
	diffs := []cardDiff{}

	// Step 1: DELETE current 里 expected 没有的卡（current map 的 key 是 card_id == anchor）
	// 按 alarm 锚定卡规则：card 删 → 同 card_id 上非终态 alarm 一并 expired（事务内原子）
	for cardID, cur := range current {
		if expected[cardID] {
			continue
		}
		if err := card.ExpireAlarmsByCardID(ctx, tx, cardID, "reconcile_delete"); err != nil {
			return nil, fmt.Errorf("expire alarms for card %s: %w", cardID, err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM cards WHERE card_id = $1::INET`, cardID); err != nil {
			return nil, fmt.Errorf("delete card %s: %w", cardID, err)
		}
		diffs = append(diffs, cardDiff{prefix: cardID, op: "delete", prevHoA: cur.residentID})
	}

	// Step 2: UPSERT each anchor (anchor 就是 card_id)
	type assignEntry struct{ anchor, unitPref string }
	var assigns []assignEntry
	for anchor := range expected {
		ap, perr := netip.ParsePrefix(anchor)
		if perr != nil {
			s.logger.Warn("applyDiffs: skip bad anchor", zap.String("anchor", anchor))
			continue
		}
		unitPref := ""
		if ap.Bits() >= 80 {
			unitPref = netip.PrefixFrom(ap.Addr(), 80).Masked().String()
		}
		assigns = append(assigns, assignEntry{anchor: anchor, unitPref: unitPref})
		diff, err := s.upsertCard(ctx, tx, anchor, unitPref, current)
		if err != nil {
			return nil, err
		}
		if diff != nil {
			diffs = append(diffs, *diff)
		}
	}

	// Step 3: 子表 FK 重新挂——按 anchor mask DESC 最具体优先
	sort.Slice(assigns, func(i, j int) bool {
		ai, _ := netip.ParsePrefix(assigns[i].anchor)
		aj, _ := netip.ParsePrefix(assigns[j].anchor)
		return ai.Bits() > aj.Bits()
	})
	// 清 scope 内 card_id（让 SET 重新生效）
	scopeForClear := map[string]bool{}
	for _, a := range assigns {
		if a.unitPref != "" {
			scopeForClear[a.unitPref] = true
		}
	}
	for unit := range scopeForClear {
		_, _ = tx.ExecContext(ctx, `UPDATE rooms SET card_id = NULL WHERE room_id <<= $1::INET`, unit)
		_, _ = tx.ExecContext(ctx, `UPDATE devices SET card_id = NULL WHERE device_addr <<= $1::INET`, unit)
	}
	// 按 mask 从细到粗填，最具体优先（已 sort）
	for _, a := range assigns {
		_, _ = tx.ExecContext(ctx,
			`UPDATE rooms SET card_id = $1::INET WHERE room_id <<= $2::INET AND card_id IS NULL`,
			a.anchor, a.anchor)
		_, _ = tx.ExecContext(ctx,
			`UPDATE devices SET card_id = $1::INET WHERE device_addr <<= $2::INET AND card_id IS NULL`,
			a.anchor, a.anchor)
	}

	// Step 4: recompute 三 flag per anchor — 必须在 Step 3 rooms/devices.card_id reassign 之后。
	//
	// has_bed: IPv6 bed_slot 字节（bit 88-95）非 0 ⟺ device 绑某床。用
	//   `host(network(/96)) != host(network(/88))` 判定（inet != 含 masklen，要 host() 去掉）。
	//   不 JOIN beds 表 — 地址编码本身即真相，beds 表只是命名索引。
	// has_bathroom / has_kitchen: 卡归属的 rooms 里有 room_type=1/2。
	//   按 rooms.card_id = anchor 判定（Step 3 LPM 结果），保证 /80 lazy 卡只算自己拿到的 room。
	//
	// 翻转 → emit "*_changed" diff 让 publishDiff 推 cards=[anchor] 供 cardagg metaCache 失效。
	// 否则 bind/room-rebind 不改 card 结构时 metaCache 持 stale flag，FE card_priority 错。
	diffByPrefix := map[string]bool{}
	for _, d := range diffs {
		diffByPrefix[d.prefix] = true
	}
	for _, a := range assigns {
		var newHasBed, newHasBath, newHasKitchen bool
		err := tx.QueryRowContext(ctx, `
			UPDATE cards SET
			    has_bed = EXISTS(
			        SELECT 1 FROM devices d
			         WHERE d.card_id = $1::INET
			           AND d.monitoring_enabled = TRUE
			           AND host(network(set_masklen(d.device_addr, 96))) != host(network(set_masklen(d.device_addr, 88)))
			    ),
			    has_bathroom = EXISTS(
			        SELECT 1 FROM rooms r
			         WHERE r.card_id = $1::INET AND r.room_type = 1
			    ),
			    has_kitchen = EXISTS(
			        SELECT 1 FROM rooms r
			         WHERE r.card_id = $1::INET AND r.room_type = 2
			    )
			WHERE card_id = $1::INET
			RETURNING has_bed, has_bathroom, has_kitchen`, a.anchor,
		).Scan(&newHasBed, &newHasBath, &newHasKitchen)
		if err != nil {
			return nil, fmt.Errorf("recompute flags %s: %w", a.anchor, err)
		}
		if diffByPrefix[a.anchor] {
			continue // 已有结构性 diff，flag 变化自然包含其中
		}
		prev := current[a.anchor]
		var op string
		switch {
		case prev.hasBed != newHasBed:
			op = "has_bed_changed"
		case prev.hasBathroom != newHasBath:
			op = "has_bathroom_changed"
		case prev.hasKitchen != newHasKitchen:
			op = "has_kitchen_changed"
		}
		if op != "" {
			diffs = append(diffs, cardDiff{prefix: a.anchor, op: op})
			diffByPrefix[a.anchor] = true
		}
	}

	return diffs, nil
}

// upsertCard 原 doc 方案：card_id ≡ anchor prefix
//   p:        anchor CIDR (/96 /88 /80 等) — 同时就是 card_id
//   unitPref: anchor 所在 unit /80（denormalize；/48..../80 卡为空）
func (s *CardSyncService) upsertCard(ctx context.Context, tx *sql.Tx,
	p, unitPref string,
	current map[string]currentCard) (*cardDiff, error) {

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

	cardDNS := card.ShortCodeOf(p)

	// has_bed / has_bathroom / has_kitchen 三 flag 全在 Step 4 统一 recompute
	//（Step 3 rooms/devices.card_id reassign 后按 card_id = anchor 精确归属判定）。
	// 此处占位 false，避免 /80 lazy 卡误吸 /88 子卡的床/卫/厨。
	hasBed, hasBath, hasKitchen := false, false, false

	diff := computeCardDiff(p, current[p], newHoA, cardName)

	var unitArg interface{} = nil
	if unitPref != "" {
		unitArg = unitPref
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO cards (card_id, unit_id, card_name, card_dns, resident_id,
		                   has_bed, has_bathroom, has_kitchen)
		VALUES ($1::INET, $2::INET, $3, $4, $5::INET, $6, $7, $8)
		ON CONFLICT (card_id) DO UPDATE SET
		    card_name    = EXCLUDED.card_name,
		    card_dns     = EXCLUDED.card_dns,
		    resident_id  = EXCLUDED.resident_id,
		    has_bed      = EXCLUDED.has_bed,
		    has_bathroom = EXCLUDED.has_bathroom,
		    has_kitchen  = EXCLUDED.has_kitchen,
		    updated_at   = NOW()
	`, p, unitArg, cardName, cardDNS, ridArg,
		hasBed, hasBath, hasKitchen)
	if err != nil {
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
	var op string
	switch d.op {
	case "delete":
		op = "delete"
	case "create", "admission", "discharge", "transfer", "name_changed",
		"has_bed_changed", "has_bathroom_changed", "has_kitchen_changed":
		op = "update"
	default:
		return
	}
	// 补齐 affected device 范围 — 让 gateway (qinglan/sleepace) 也能按 device_uid 精确失效 baseline cache，
	// cardagg 也能用 device_addrs 失效 enablement cache。否则 gateway 拿到 cards-only payload 啥也做不了。
	// delete op：用 publishDiff 之前的 prefix 范围查（DB row 还在；但 caller 已在 Step1 ExpireAlarms 后 DELETE
	// FROM cards，devices.card_id 也已被 Step3 reassign 到别处或 NULL）—— 此处先查 device_addr <<= prefix 而非
	// d.card_id = prefix，覆盖被搬走的 device。
	var devAddrs, devUIDs []string
	if s.db != nil {
		rows, err := s.db.QueryContext(ctx, `
			SELECT host(d.device_addr), dfm.device_uid
			  FROM devices d
			  JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
			 WHERE d.device_addr <<= $1::INET
		`, d.prefix)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var addr, uid string
				if err := rows.Scan(&addr, &uid); err == nil {
					if addr != "" {
						devAddrs = append(devAddrs, addr)
					}
					if uid != "" {
						devUIDs = append(devUIDs, uid)
					}
				}
			}
		}
	}
	if err := s.publisher.PublishConfigChanged(ctx, op, []string{d.prefix}, devAddrs, devUIDs); err != nil {
		s.logger.Warn("publish card diff",
			zap.String("prefix", d.prefix), zap.String("op", d.op),
			zap.String("tenant", tenantPrefix), zap.Error(err))
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
