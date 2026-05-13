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

// ReconcileCards — Card 表唯一写入入口（v2 收敛）。
//
// 业务事件（device on/off / resident create/update/delete / startup）都调它，幂等。
//
// 规则：
//
//	1) Card EXISTS 的条件：scope 内有 device 绑到 unit/room/bed 且 monitoring_enabled=TRUE
//	2) Card.resident_id ⟺ LPM(resident_unit active, masklen ≥ 80) — /48/56 哨兵不算占用
//	3) Card.card_name ⟺ resident.nickname / "NoOne"
//
// Private unit 规则（单向阀：merge → split 不可逆，split 后增减都保持 split）：
//
//	初始 (cards 中无 /96):
//	  bedCount == 0 + 有 non-bed device → 1 张 /80
//	  bedCount == 1 → merge: 1 张 /80（bed 与 non-bed device 都在这一张）
//	  bedCount >= 2 → 触发 split: 每 bed 一张 /96 + 1 张 /80
//
//	已 split (cards 中 /96 数 >= 1):
//	  bedCount >= 1 → 持续 split: 每 bed /96 + /80（即使减到 1 bed 也不合并回 merge）
//	  bedCount == 0 → 只剩 /80（如有 non-bed device），/96 全删
//
// 单向阀语义来源：split 后再合并回 merge 会让原本独立的 bed card 消失 → 下游 DDNS/cardagg
// 缓存失效。卡稳定 > 拓扑精确。判定依赖 cards 表当前状态（非纯函数；接受路径依赖）。
//
// Share/Public unit 不应用这套规则：device anchor 自然层级 (/96 /88 /80) → 自然 card 集合。
//
// ACL **不**由本函数维护 — per-request 现场算（scope.go）。
//
// 事务：独立 tx；reconcile 失败仅 log warn，调用方业务 tx 不回滚（下次 reconcile 自愈）。
// scope: INET CIDR 任意级。
func (s *CardSyncService) ReconcileCards(ctx context.Context, scopePrefix string) error {
	if s.db == nil {
		s.logger.Debug("ReconcileCards skipped (db not wired)")
		return nil
	}
	scopePrefix = strings.TrimSpace(scopePrefix)
	if scopePrefix == "" {
		return fmt.Errorf("scopePrefix required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// === Step 1: 算 expected ===
	// 1a) 每个 monitor-on device 反推 (unit_prefix, anchor_prefix)
	//     anchor 优先级: bed > room > unit；同时记下 unit 用于 Layer 2 聚合
	expRows, err := tx.QueryContext(ctx, `
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
	`, scopePrefix)
	if err != nil {
		return fmt.Errorf("query expected: %w", err)
	}
	// 按 unit 聚合：每 unit 收集所有 bed/non-bed anchors
	type unitInfo struct {
		bedAnchors    map[string]bool // /96 anchors
		nonBedAnchors map[string]bool // /80 /88 anchors (room or unit-level device)
	}
	units := map[string]*unitInfo{}
	for expRows.Next() {
		var unitPrefix, anchor sql.NullString
		if err := expRows.Scan(&unitPrefix, &anchor); err != nil {
			_ = expRows.Close()
			return fmt.Errorf("scan expected: %w", err)
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
	_ = expRows.Close()
	if err := expRows.Err(); err != nil {
		return fmt.Errorf("iter expected: %w", err)
	}

	// 1b) Layer 2 — 分层 split 规则（按 unit / room 两级递归 activeBed > 1 判定）
	//
	// **决策规则**（已与用户对齐 2026-05-13）：
	//
	//   Layer-1 Unit：
	//     unit 内 activeBed > 1 → split:
	//       - 进 Layer-2 计算 room 归宿
	//       - 若 unit 内有 "Room 之外的 device"（即 device 在 activeBed=0 room 或 unit 层），出 1 张 unitCard
	//     unit 内 activeBed ≤ 1 → merge:
	//       - 仅 1 张 /80 unit card，装所有 device（bed + room + unit 层）
	//
	//   Layer-2 Room（仅在 Unit split 时进入）:
	//     room 内 activeBed > 1 → split:
	//       - N 张 /96 bed card（每 active bed 一张）
	//       - 若 room 内有 room-level device（byte11=0，无 bed 绑定），出 1 张 /88 roomCard
	//     room 内 activeBed = 1 → /96 bed card 吸收所有 room 内 device（room radar 也归入）
	//     room 内 activeBed = 0 → 不出 roomCard，device 上推到 Layer-1 unitCard
	//
	// **单向阀**: split 仅由 activeBed > 1 触发；activeBed ≤ 1 时永远 merge（取消旧的 alreadySplit latching）。
	//             已存在的 split 同级卡：新设备磁吸进同级卡（不重建结构）。
	expected := map[string]bool{}
	for unitPrefix, info := range units {
		bedCount := len(info.bedAnchors)
		hasNonBed := len(info.nonBedAnchors) > 0
		if bedCount == 0 && !hasNonBed {
			continue
		}

		// Merge mode：unit 内 activeBed ≤ 1 → 1 张 /80 unit card
		if bedCount <= 1 {
			expected[unitPrefix] = true
			continue
		}

		// Split mode (activeBed > 1) — 每 active bed 必出 /96
		for a := range info.bedAnchors {
			expected[a] = true
		}

		// 算每 room 内 active_bed_count
		roomBedCount := map[string]int{}
		for bedPrefix := range info.bedAnchors {
			roomPrefix := narrowPrefixToRoom(bedPrefix)
			if roomPrefix != "" {
				roomBedCount[roomPrefix]++
			}
		}

		// L2/L1: 决定每个 non-bed anchor 归宿
		needUnitCard := false
		for anchor := range info.nonBedAnchors {
			switch {
			case strings.HasSuffix(anchor, "/88"):
				// room-level device
				switch {
				case roomBedCount[anchor] >= 2:
					// activeBed > 1 → room 自己 split + 出 /88 roomCard 装 room-level device
					expected[anchor] = true
				case roomBedCount[anchor] == 1:
					// 唯一 bed 吸收 — 不创建 /88 room card；device 归该 bed card
				default:
					// activeBed = 0 → 不出 /88，device 上推到 unitCard
					needUnitCard = true
				}
			default:
				// /80 直挂 device → /80 unit card 出现
				needUnitCard = true
			}
		}
		if needUnitCard {
			expected[unitPrefix] = true
		}
	}

	// === Step 2: current — scope 内现有 cards (prefix + resident_id + card_name) ===
	// 多查 resident_id 和 card_name 用于 diff 判定 admission/discharge/transfer/name_changed
	curRows, err := tx.QueryContext(ctx, `
		SELECT spatial_prefix::text, COALESCE(host(resident_id), ''), COALESCE(card_name, '')
		  FROM cards
		 WHERE spatial_prefix <<= $1::INET
	`, scopePrefix)
	if err != nil {
		return fmt.Errorf("query current: %w", err)
	}
	type currentCard struct {
		residentID string
		cardName   string
	}
	current := map[string]currentCard{}
	for curRows.Next() {
		var p, rid, name string
		if err := curRows.Scan(&p, &rid, &name); err != nil {
			_ = curRows.Close()
			return fmt.Errorf("scan current: %w", err)
		}
		current[p] = currentCard{residentID: rid, cardName: name}
	}
	_ = curRows.Close()
	if err := curRows.Err(); err != nil {
		return fmt.Errorf("iter current: %w", err)
	}

	// === Step 3: diff & DELETE 多余 ===
	// 同时记录 diff，commit 后 emit CloudEvent 给下游（device_gateway / wisefido-sensor / cardagg）
	diffs := []cardDiff{}

	toDelete := []string{}
	for p, cur := range current {
		if !expected[p] {
			toDelete = append(toDelete, p)
			diffs = append(diffs, cardDiff{prefix: p, op: "delete", prevHoA: cur.residentID})
		}
	}
	for _, p := range toDelete {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM cards WHERE spatial_prefix = $1::INET`, p); err != nil {
			return fmt.Errorf("delete card %s: %w", p, err)
		}
	}

	// === Step 4: INSERT 缺失 + Step 5: overlay resident_id/card_name（合并在一个 UPSERT）===
	// 对 expected 集合每个 prefix：
	//   - 查 LPM resident_unit active @ masklen>=80 → resident_id（可空）
	//   - 查 residents.nickname → card_name（resident_id 为空时 "NoOne"）
	//   - INSERT ... ON CONFLICT DO UPDATE
	//   - 同时记录 diff（create / admission / discharge / transfer / name_changed）
	for p := range expected {
		var residentID sql.NullString
		// 两阶段 overlay：
		//   1) LPM 双向 overlap：card prefix 与 ru.spatial_prefix 任一包含对方
		//      处理 Private unit /80 找子级 resident，或 bed card 找父级 unit resident
		//   2) Unit fallback：若阶段 1 找不到，看该 card 所在 /80 unit 内是否唯一 active resident
		//      命中此 fallback 的常见场景：Private unit 1 个 resident 但绑在 /96 specific bed，
		//      同 unit 的其他 room/bathroom cards (/88) 与 bed prefix 不在同一 prefix 树 → LPM 单点找不到
		//      此 fallback 保证 "Private unit 整 unit 归一人" 语义（不需 unit_type 分支即可成立）；
		//      Share unit 多 resident 时 fallback 不命中（cnt != 1），保持 NoOne
		_ = tx.QueryRowContext(ctx, `
			SELECT COALESCE(
			  -- 阶段 1: LPM 双向 overlap
			  (SELECT host(ru.resident_id)
			     FROM resident_unit ru
			    WHERE ru.valid_to IS NULL
			      AND ($1::INET <<= ru.spatial_prefix OR ru.spatial_prefix <<= $1::INET)
			      AND masklen(ru.spatial_prefix) >= 80
			    ORDER BY masklen(ru.spatial_prefix) DESC
			    LIMIT 1),
			  -- 阶段 2: unit-level fallback — 该 unit 内唯一 active resident 时命中
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
		`, p).Scan(&residentID)

		cardName := CardNameNoResident
		var ridArg interface{} = nil
		newHoA := ""
		if residentID.Valid && residentID.String != "" {
			var nick sql.NullString
			_ = tx.QueryRowContext(ctx,
				`SELECT COALESCE(nickname, '') FROM residents WHERE host(resident_id) = $1`,
				residentID.String).Scan(&nick)
			if nick.Valid && nick.String != "" {
				cardName = nick.String
			}
			ridArg = residentID.String
			newHoA = residentID.String
		}

		cardType, err := cardTypeForPrefix(p)
		if err != nil {
			s.logger.Warn("ReconcileCards: skip prefix (unknown card_type)",
				zap.String("prefix", p), zap.Error(err))
			continue
		}
		// dns_short_name: br01-s11-u0003[-r03[-b01]] 助记词；按 prefix masklen 自然展开
		shortName := ""
		if parsed, perr := netip.ParsePrefix(p); perr == nil {
			if sn, snerr := ddns.CardShortName(parsed); snerr == nil {
				shortName = sn
			}
		}

		// 判 diff：本次是 INSERT 还是 UPDATE，UPDATE 又分 admission/discharge/transfer/name_changed
		cur, existed := current[p]
		if !existed {
			diffs = append(diffs, cardDiff{prefix: p, op: "create", newHoA: newHoA})
			if newHoA != "" {
				diffs = append(diffs, cardDiff{prefix: p, op: "admission", newHoA: newHoA})
			}
		} else if cur.residentID != newHoA {
			switch {
			case cur.residentID == "":
				diffs = append(diffs, cardDiff{prefix: p, op: "admission", prevHoA: "", newHoA: newHoA})
			case newHoA == "":
				diffs = append(diffs, cardDiff{prefix: p, op: "discharge", prevHoA: cur.residentID, newHoA: ""})
			default:
				diffs = append(diffs, cardDiff{prefix: p, op: "transfer", prevHoA: cur.residentID, newHoA: newHoA})
			}
		} else if cur.cardName != cardName {
			// resident_id 不变但 nickname 变了
			diffs = append(diffs, cardDiff{prefix: p, op: "name_changed", prevHoA: newHoA, newHoA: newHoA})
		}

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
			return fmt.Errorf("upsert card %s: %w", p, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	s.logger.Info("ReconcileCards done",
		zap.String("scope", scopePrefix),
		zap.Int("expected", len(expected)),
		zap.Int("current_before", len(current)),
		zap.Int("deleted", len(toDelete)),
		zap.Int("diffs", len(diffs)))

	// === Step 6: 测试 observer + emit CloudEvent (commit 后才发，事务回滚则一次都不发) ===
	// 下游订阅方:
	//   - device_gateway: 接 create/delete 更新设备→prefix 路由
	//   - wisefido-sensor: 接 create/delete 订阅/退订监测流
	//   - cardagg: 接 admission/discharge/transfer 失效 resident 视图缓存
	if s.reconcileObserver != nil {
		s.reconcileObserver(scopePrefix, diffs)
	}
	s.emitDiffs(ctx, scopePrefix, diffs)

	// === Step 7: DDNS ensure register for ALL expected prefixes (force-sync) ===
	// 不仅 create diff 调 register，每次 reconcile 都对当前 expected 集合 ensure DNS 记录
	// nsupdate "add AAAA" 同名同值是 noop / 同名不同值是替换 → 天然 idempotent
	// 这样保证：DB cards 列与 bind9 zone 双向同步，即使中间 bind9 重启或网络抖动也能自愈
	s.ensureDDNSForExpected(ctx, scopePrefix, expected)
	return nil
}

// ensureDDNSForExpected — 对 expected 集合的每个 prefix 做 DDNS register（idempotent）
// 与 emitDiffs 的区别：emitDiffs 只对 create/delete 触发；本函数对所有 expected 全量 ensure。
// 失败仅 log warn，不影响 cards 表已 commit 的状态。
func (s *CardSyncService) ensureDDNSForExpected(ctx context.Context, scopePrefix string, expected map[string]bool) {
	if s.ddns == nil || len(expected) == 0 {
		return
	}
	tenantSlot := tenantSlotOf(scopePrefix)
	if tenantSlot == 0 {
		return
	}
	zone := ddns.ZoneForTenant(tenantSlot, s.owlDomain)
	for p := range expected {
		parsed, perr := netip.ParsePrefix(p)
		if perr != nil {
			continue
		}
		shortName := card.ShortCodeOf(p)
		if err := s.ddns.RegisterCardName(ctx, parsed, shortName, zone); err != nil {
			s.logger.Warn("ensureDDNSForExpected: register failed",
				zap.String("fqdn", shortName+"."+zone),
				zap.String("prefix", p),
				zap.Error(err))
		}
	}
}

// cardDiff — ReconcileCards 算出的单卡变化记录，commit 后用于 emit CloudEvent
type cardDiff struct {
	prefix  string // spatial_prefix
	op      string // "create" / "delete" / "admission" / "discharge" / "transfer" / "name_changed"
	prevHoA string // 旧 resident_id（host 格式，无 /128 后缀）
	newHoA  string // 新 resident_id
}

// SetReconcileObserver — 测试 hook，每次 ReconcileCards commit 后调一次。生产留 nil。
// 比拦截 ConfigPublisher 简单：直接看 diffs slice 内容，与 publisher 实现解耦。
func (s *CardSyncService) SetReconcileObserver(fn func(scope string, diffs []cardDiff)) {
	s.reconcileObserver = fn
}

// emitDiffs — commit 之后做两件事：
//   1. DDNS register/deregister 同步 bind9（仅 create/delete）
//   2. Redis CloudEvent 通知下游 cardagg / sensor / device_gateway
// 失败仅 log warn（cards 表已 commit，bind9 / Redis 失败下次 reconcile 自然兜底）
func (s *CardSyncService) emitDiffs(ctx context.Context, scopePrefix string, diffs []cardDiff) {
	if len(diffs) == 0 {
		return
	}
	tenantPrefix := tenantPrefixOf(scopePrefix)
	tenantSlot := tenantSlotOf(scopePrefix) // 推 zone name 用
	for _, d := range diffs {
		// === 1. DDNS sync ===
		if s.ddns != nil && tenantSlot > 0 {
			parsed, perr := netip.ParsePrefix(d.prefix)
			if perr == nil {
				shortName := card.ShortCodeOf(d.prefix) // 与 cards.dns_short_name 同源
				zone := ddns.ZoneForTenant(tenantSlot, s.owlDomain)
				switch d.op {
				case "create":
					if err := s.ddns.RegisterCardName(ctx, parsed, shortName, zone); err != nil {
						s.logger.Warn("emitDiffs: DDNS register failed",
							zap.String("fqdn", shortName+"."+zone), zap.Error(err))
					}
				case "delete":
					if err := s.ddns.UnregisterCardName(ctx, parsed, shortName, zone); err != nil {
						s.logger.Warn("emitDiffs: DDNS unregister failed",
							zap.String("fqdn", shortName+"."+zone), zap.Error(err))
					}
				}
			}
		}

		// === 2. Redis CloudEvent ===
		if s.publisher == nil {
			continue
		}
		var err error
		switch d.op {
		case "create", "delete":
			err = s.publisher.PublishCardChanged(ctx, tenantPrefix, d.prefix, d.op, d.prefix, "")
		case "admission", "discharge", "transfer", "name_changed":
			err = s.publisher.PublishCardResidentChanged(ctx, tenantPrefix, d.prefix, d.op, d.prevHoA, d.newHoA, d.prefix)
		}
		if err != nil {
			s.logger.Warn("emitDiffs: publish failed",
				zap.String("prefix", d.prefix),
				zap.String("op", d.op),
				zap.Error(err))
		}
	}
}

// tenantSlotOf — 从 INET CIDR 推 tenant slot (第 3 段 hex 转 uint16)
//   "fd00:0:3:..." → 3；用于 ddns.ZoneForTenant
func tenantSlotOf(prefix string) uint16 {
	idx := strings.LastIndex(prefix, "/")
	if idx <= 0 {
		return 0
	}
	parts := strings.Split(strings.Split(prefix[:idx], "::")[0], ":")
	if len(parts) < 3 {
		return 0
	}
	var v uint64
	for _, ch := range parts[2] {
		var d uint64
		switch {
		case ch >= '0' && ch <= '9':
			d = uint64(ch - '0')
		case ch >= 'a' && ch <= 'f':
			d = uint64(ch-'a') + 10
		case ch >= 'A' && ch <= 'F':
			d = uint64(ch-'A') + 10
		default:
			return 0
		}
		v = v*16 + d
	}
	return uint16(v)
}

// tenantPrefixOf — 把任意 INET CIDR 截到 /48 tenant prefix
//   "fd00:0:3:112:3::/80" → "fd00:0:3::/48"
//   "fd00:0:3::/48"       → "fd00:0:3::/48"
func tenantPrefixOf(prefix string) string {
	idx := strings.LastIndex(prefix, "/")
	if idx <= 0 {
		return prefix
	}
	addr := prefix[:idx]
	// 取前 3 段（fd00:0:T）= 48 bit
	parts := strings.Split(strings.Split(addr, "::")[0], ":")
	if len(parts) < 3 {
		return prefix
	}
	return strings.Join(parts[:3], ":") + "::/48"
}

// narrowPrefixToRoom — 把 /96 bed prefix 截到 /88 room；其他原样返回
// IPv6 第 6 段（uint16）= room_slot(8 bit) << 8 | bed_slot(8 bit)
// /96 lock 整 16 bit；/88 lock 高 8 bit（room），低 8 bit（bed）清零
// 例: "fd00:0:3:111:3:101::/96" → "fd00:0:3:111:3:100::/88"
//     (segment 0x0101 → 高 byte 0x01 → 新 segment 0x0100，IPv6 显示 "100")
func narrowPrefixToRoom(prefix string) string {
	idx := strings.LastIndex(prefix, "/")
	if idx <= 0 || !strings.HasSuffix(prefix, "/96") {
		return ""
	}
	addr := prefix[:idx]
	parts := strings.Split(strings.Split(addr, "::")[0], ":")
	if len(parts) < 6 {
		return ""
	}
	// 解析第 6 段为 uint16（容忍 leading-zero 缩写：fmt.Sprintf("%x", ...) 后无前导 0）
	var seg uint64
	for _, ch := range parts[5] {
		var d uint64
		switch {
		case ch >= '0' && ch <= '9':
			d = uint64(ch - '0')
		case ch >= 'a' && ch <= 'f':
			d = uint64(ch-'a') + 10
		case ch >= 'A' && ch <= 'F':
			d = uint64(ch-'A') + 10
		default:
			return ""
		}
		seg = seg*16 + d
	}
	roomByte := seg >> 8           // 高 8 bit = room slot
	newSeg := roomByte << 8        // 低 8 bit 清零
	// IPv6 segment 缩写格式（无前导 0）
	parts[5] = fmt.Sprintf("%x", newSeg)
	return strings.Join(parts[:6], ":") + "::/88"
}

// cardTypeForPrefix — 按 masklen 决定 card_type
//   /80 → unit；/88 → room；/96 → active_bed
//   其他 masklen（如 /56 /128）不应作为 card prefix
func cardTypeForPrefix(prefixStr string) (string, error) {
	// 简单按 "/N" 后缀判定，避开引入 netip 解析依赖
	switch {
	case strings.HasSuffix(prefixStr, "/96"):
		return "active_bed", nil
	case strings.HasSuffix(prefixStr, "/88"):
		return "room", nil
	case strings.HasSuffix(prefixStr, "/80"):
		return "unit", nil
	default:
		return "", fmt.Errorf("unsupported masklen for card_type: %s", prefixStr)
	}
}
