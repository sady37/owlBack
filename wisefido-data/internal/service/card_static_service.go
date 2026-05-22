package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	commoncard "owl-common/card"
	"wisefido-data/internal/models"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// CardStaticService 静态卡片服务
// login 时查一次 DB，返回 CardStatic 列表，并推 cardList 给 realtime
type CardStaticService struct {
	db              *sql.DB
	allowedProvider AllowedCardIDsProvider
	realtime        *CardRealtimeService
	logger          *zap.Logger
}

// NewCardStaticService 创建静态卡片服务
func NewCardStaticService(
	db *sql.DB,
	allowedProvider AllowedCardIDsProvider,
	realtime *CardRealtimeService,
	logger *zap.Logger,
) *CardStaticService {
	return &CardStaticService{
		db:              db,
		allowedProvider: allowedProvider,
		realtime:        realtime,
		logger:          logger,
	}
}

// ListCardStatic 后端专用：从 database 获取静态卡片列表，仅按 tenant/branch/unit 过滤，不做用户权限校验
func (s *CardStaticService) GetCardList(ctx context.Context, tenantID, userID, userRole string, branchIDs []string, page, pageSize int) ([]commoncard.CardStatic, *models.BackendPagination, error) {
	// 1. 从 context 取 userType
	_, _, userType, _, _ := GetSessionFromContext(ctx)
	if userType == "" {
		userType = "staff"
	}

	// 2. 获取允许的 CardList（按 branch 分组）
	cardList, err := s.allowedProvider.GetCardList(ctx, tenantID, userID, userType)
	if err != nil {
		return nil, nil, fmt.Errorf("get allowed card list: %w", err)
	}
	allCardIDs := cardList.AllCardIDs()
	if len(allCardIDs) == 0 {
		return nil, &models.BackendPagination{Page: page, Size: pageSize, Count: 0}, nil
	}

	// 3. 推 cardList 给 realtime（更新允许清单）
	if s.realtime != nil {
		s.realtime.UpdateCardList(cardList, userType)
	}

	// 4. SQL JOIN 查询，带 branchIDs 过滤 + 分页
	cards, total, err := s.queryCardsByIDs(ctx, allCardIDs, branchIDs, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("query cards: %w", err)
	}

	pagination := &models.BackendPagination{
		Page:  page,
		Size:  pageSize,
		Count: total,
	}

	s.logger.Info("GetCardList",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("allowed", len(allCardIDs)),
		zap.Int("total", total),
		zap.Int("returned", len(cards)),
	)

	return cards, pagination, nil
}

// GetCardInfo 根据 cardID 查询单个卡片信息（带安全检查）
// cardID 接受两种格式：
//  1. spatial_prefix INET CIDR (含 ":" 或 "/")，如 "fd00:0:3:111:3:301::/96"
//  2. dns_short_name 6 位 base36 (无 ":"/"/")，如 "poqfqu" — URL 友好，不暴露 IPv6 拓扑
//
// 短码先查 cards.card_dns 解析为 spatial_prefix，再走原权限/查询路径。
func (s *CardStaticService) GetCardInfo(ctx context.Context, tenantID, userID, cardID string) (*commoncard.CardStatic, error) {
	resolvedID, err := s.resolveCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	_, _, userType, _, _ := GetSessionFromContext(ctx)
	if userType == "" {
		userType = "staff"
	}
	cardList, err := s.allowedProvider.GetCardList(ctx, tenantID, userID, userType)
	if err != nil {
		return nil, fmt.Errorf("get allowed card list: %w", err)
	}
	allowed := make(map[string]bool)
	for _, id := range cardList.AllCardIDs() {
		allowed[id] = true
	}
	if !allowed[resolvedID] {
		return nil, fmt.Errorf("card not found or no permission")
	}
	cards, _, err := s.queryCardsByIDs(ctx, []string{resolvedID}, nil, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, fmt.Errorf("card not found")
	}
	return &cards[0], nil
}

// resolveCardID 把 dns_short_name (6 位 base36) 解析回 spatial_prefix CIDR；
// IPv6 形式（含 ":"/"/"）原样返回。短码查不到 → error，不静默 fallback。
func (s *CardStaticService) resolveCardID(ctx context.Context, cardID string) (string, error) {
	if cardID == "" {
		return "", fmt.Errorf("card_id required")
	}
	// 含 ":" 或 "/" 必是 IPv6 CIDR；纯 6 位字母数字才走短码解析
	if strings.ContainsAny(cardID, ":/") {
		return cardID, nil
	}
	if !isShortCodeFormat(cardID) {
		// 既不像 IPv6 也不像 6 位短码，留给后续 allow check 报 "not found"
		return cardID, nil
	}
	var spatialPrefix string
	err := s.db.QueryRowContext(ctx,
		`SELECT host(card_id) || '/' || masklen(card_id) FROM cards WHERE card_dns = $1`,
		cardID,
	).Scan(&spatialPrefix)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("card not found (short_name=%s)", cardID)
	}
	if err != nil {
		return "", fmt.Errorf("resolve dns_short_name: %w", err)
	}
	return spatialPrefix, nil
}

// isShortCodeFormat 6 位 base36 (a-z + 0-9)；与 owl-common/card.ShortCodeOf 输出格式一致
func isShortCodeFormat(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

// GetCardsByCardIDs 根据 cardID 列表直接查询 CardStatic（无分页）
func (s *CardStaticService) GetCardsByCardIDs(ctx context.Context, tenantID, userID string, cardIDs []string) ([]commoncard.CardStatic, error) {
	if len(cardIDs) == 0 {
		return nil, nil
	}

	// 权限校验：从 realtime 取已存储的 CardList
	stored := s.realtime.getCardList(tenantID, userID)
	if stored == nil {
		return nil, fmt.Errorf("no stored card list, please call GetCardList first")
	}
	allowedSet := make(map[string]bool)
	for _, id := range stored.AllCardIDs() {
		allowedSet[id] = true
	}

	var validIDs []string
	for _, id := range cardIDs {
		if allowedSet[id] {
			validIDs = append(validIDs, id)
		}
	}
	if len(validIDs) == 0 {
		return nil, nil
	}

	cards, _, err := s.queryCardsByIDs(ctx, validIDs, nil, 1, len(validIDs))
	if err != nil {
		return nil, err
	}
	return cards, nil
}

// queryCardsByIDs 用 card_id[] 做联合查询，直接组装 CardStatic
// v2 unified: card_id ≡ spatial_prefix；unit/room/bed 通过 set_masklen 派生 + JOIN 拿名字。
// 返回时 card_address 是 "Unit X / Room Y / Bed Z" 友好串（按最长 mask 拼）；
// residents/devices 由 enrichResidentsAndDevices 后续批量填充。
func (s *CardStaticService) queryCardsByIDs(ctx context.Context, cardIDs []string, branchIDs []string, page, pageSize int) ([]commoncard.CardStatic, int, error) {
	// v2.5: card_id ≡ spatial_prefix (原 doc 锁定方案)；card_type 派生自 masklen + card_name
	query := `
		SELECT
			text(c.card_id) AS card_id,
			text(c.card_id) AS spatial_prefix,
			CASE masklen(c.card_id)
				WHEN  48 THEN 'tenant'
				WHEN  56 THEN 'branch'
				WHEN  64 THEN 'site'
				WHEN  80 THEN CASE WHEN c.card_name = 'public' THEN 'public' ELSE 'unit' END
				WHEN  88 THEN 'room'
				WHEN  96 THEN 'bed'
				WHEN 128 THEN 'device'
			END AS card_type,
			COALESCE(c.card_name, ''),
			COALESCE(c.card_dns, ''),
			COALESCE(text(c.resident_id), ''),
			COALESCE(u.unit_id::text, ''),
			COALESCE(u.unit_name, ''),
			COALESCE(u.unit_type, 0)     AS unit_type,
			COALESCE(u.unit_property, 0) AS unit_property,
			COALESCE(u.timezone, ''),
			COALESCE(br.branch_id::text, '') AS branch_id,
			COALESCE(br.branch_name, ''),
			COALESCE(s.site_id::text, '')    AS building_id,
			COALESCE(s.site_name, '')        AS building_name,
			COALESCE(rm.room_id::text, '')   AS room_id,
			COALESCE(rm.room_name, '')       AS room_name,
			COALESCE(rm.room_type, 0)        AS room_type,
			COALESCE(b.bed_id::text, '')     AS bed_id,
			COALESCE(b.bed_name, '')         AS bed_name,
			COUNT(*) OVER() AS total_count
		FROM cards c
		LEFT JOIN units    u  ON u.unit_id    = c.unit_id
		LEFT JOIN branches br ON br.branch_id = network(set_masklen(c.unit_id, 56))
		LEFT JOIN sites    s  ON s.site_id    = network(set_masklen(c.unit_id, 64))
		-- room/bed lookup 按 card_id mask 严格分级（rule.md C9 终版三种卡）：
		--   /80 unit card: room=NULL bed=NULL  (FE 显示到 unit 层即止)
		--   /88 room card: room=自己 bed=NULL  (FE 显示到 room 层)
		--   /96 bed  card: room=parent /88 bed=自己 (FE 显示到 bed 层)
		LEFT JOIN LATERAL (
			SELECT room_id, room_name, room_type
			  FROM rooms
			 WHERE (masklen(c.card_id) = 88 AND room_id = c.card_id)
			    OR (masklen(c.card_id) = 96 AND room_id >>= c.card_id)
			 LIMIT 1
		) rm ON TRUE
		LEFT JOIN LATERAL (
			SELECT bed_id, bed_name
			  FROM beds
			 WHERE masklen(c.card_id) = 96 AND bed_id = c.card_id
			 LIMIT 1
		) b ON TRUE
		WHERE c.card_id = ANY($1::inet[])
	`
	args := []any{pq.Array(cardIDs)}
	argIdx := 2

	if len(branchIDs) > 0 {
		query += fmt.Sprintf(` AND c.unit_id <<= ANY($%d::inet[])`, argIdx)
		args = append(args, pq.Array(branchIDs))
		argIdx++
	}

	// 排序: Unit > Nickname > Room > Bed
	//   同 unit 卡聚合在一起；unit 内按 nickname (resident 名，NoOne 排在该层末尾)；
	//   再按 room / bed 字典序，保证 FE 分页时同 unit 卡不被切散
	query += ` ORDER BY u.unit_name ASC NULLS LAST,
	                  c.card_name  ASC NULLS LAST,
	                  rm.room_name ASC NULLS LAST,
	                  b.bed_name   ASC NULLS LAST`

	offset := (page - 1) * pageSize
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query cards: %w", err)
	}
	defer rows.Close()

	var cards []commoncard.CardStatic
	var totalCount int

	for rows.Next() {
		var (
			cardID, spatialPrefix, cardType, cardName, dnsShortName, residentID string
			unitID, unitName, timezone                                          string
			unitType, unitProperty, roomType                                    int
			branchID, branchName, buildingID, buildingName                      string
			roomID, roomName, bedID, bedName                                    string
		)
		if err := rows.Scan(
			&cardID, &spatialPrefix, &cardType, &cardName, &dnsShortName, &residentID,
			&unitID, &unitName, &unitType, &unitProperty, &timezone,
			&branchID, &branchName, &buildingID, &buildingName,
			&roomID, &roomName, &roomType,
			&bedID, &bedName,
			&totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan card row: %w", err)
		}

		unit := &commoncard.UnitInfo{
			UnitID:       unitID,
			UnitName:     unitName,
			UnitType:     unitType,
			UnitProperty: unitProperty,
			BranchID:     branchID,
			BranchName:   branchName,
			BuildingID:   buildingID,
			BuildingName: buildingName,
			Timezone:     timezone,
			IsPublic:     unitType == commoncard.UnitTypePublic,
			IsSharedUnit: unitType == commoncard.UnitTypeShare,
		}
		c := commoncard.CardStatic{
			CardID:       cardID, // v2.5: cardID == spatial_prefix
			CardType:     cardType,
			CardName:     cardName,
			DNSShortName: dnsShortName,
			Unit:         unit,
		}
		_ = spatialPrefix // v2.5: spatial_prefix 字段移除（== cardID）；保留 local var 防破坏其他 caller
		if roomID != "" || roomName != "" {
			c.Room = &commoncard.RoomInfo{
				RoomID:   roomID,
				RoomName: roomName,
				RoomType: roomType,
			}
		}
		if bedID != "" || bedName != "" {
			c.Bed = &commoncard.BedInfo{
				BedID:   bedID,
				BedName: bedName,
			}
		}
		if residentID != "" {
			c.Residents = []commoncard.ResidentInfo{{ResidentID: residentID}}
		}
		cards = append(cards, c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rows: %w", err)
	}

	// 批量 enrich：residents.nickname + devices LPM
	if err := s.enrichResidentsAndDevices(ctx, cards); err != nil {
		s.logger.Warn("enrichResidentsAndDevices failed (returning partial cards)",
			zap.Error(err))
	}

	return cards, totalCount, nil
}

// enrichResidentsAndDevices 批量补 residents.nickname / devices LPM 列表。
// 避免 N+1 query：先收集所有 resident_id 和 spatial_prefix，再两条 query 拿回，最后 in-memory join。
func (s *CardStaticService) enrichResidentsAndDevices(ctx context.Context, cards []commoncard.CardStatic) error {
	if len(cards) == 0 {
		return nil
	}
	// === 1. 批量 fill residents.nickname ===
	residentIDs := make([]string, 0, len(cards))
	for _, c := range cards {
		for _, r := range c.Residents {
			if r.ResidentID != "" {
				residentIDs = append(residentIDs, r.ResidentID)
			}
		}
	}
	type residentMeta struct {
		nickname     string
		dnsShortName string
	}
	residentMap := make(map[string]residentMeta) // resident_id → {nickname, dns_short_name}
	if len(residentIDs) > 0 {
		rows, err := s.db.QueryContext(ctx,
			`SELECT resident_id::text, COALESCE(nickname, ''), COALESCE(dns_short_name, '') FROM residents WHERE resident_id = ANY($1::INET[])`,
			pq.Array(residentIDs),
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var rid, nick, short string
				if rows.Scan(&rid, &nick, &short) == nil {
					residentMap[rid] = residentMeta{nickname: nick, dnsShortName: short}
				}
			}
		}
	}
	// === 2. 批量 fill devices via v3 套娃归属规则 ===
	// 规则:
	//   bed-anchor (/96) device → 该 bed card
	//   room-anchor (/88) device:
	//     ├ room 内 active_bed_count=1 → 归该唯一 bed card (吸收)
	//     ├ 否则 cards 表有 /88 room card → 归该 /88
	//     └ 否则 → 归 /80 unit card (兜底)
	//   unit-anchor (/80) device → /80 unit card
	//
	// 实现：一次 SQL 查 device + 三 anchor，Go in-memory join cards 表算 owning_card
	if err := s.fillDevicesV3(ctx, cards); err != nil {
		s.logger.Warn("fillDevicesV3 failed", zap.Error(err))
	}

	// === 3. apply residents nickname + dns_short_name ===
	for i := range cards {
		for j := range cards[i].Residents {
			rid := cards[i].Residents[j].ResidentID
			if meta, ok := residentMap[rid]; ok {
				cards[i].Residents[j].Nickname = meta.nickname
				cards[i].Residents[j].DNSShortName = meta.dnsShortName
			}
		}
	}
	return nil
}

// fillDevicesV3 — 按套娃规则把 device 归到 owning card；同时算 coverage_label（unit card 用）
// in-memory join 比 SQL CASE WHEN 嵌套清晰：每 device 算 owning_card，按 owning 分组装 cards[].Devices
func (s *CardStaticService) fillDevicesV3(ctx context.Context, cards []commoncard.CardStatic) error {
	if len(cards) == 0 {
		return nil
	}
	// 索引 cards by prefix（用于 owning_card 反查 是否真实存在）
	// 注意：cardByPrefix 标记"该 prefix 是本次请求要返回的卡"（用于 owning 写回）。
	// 真实"卡存在性"另用 siblingCardExists 查 DB（避免 single-card 视图误判 merge mode）。
	cardByPrefix := map[string]int{} // prefix → index in cards
	for i, c := range cards {
		cardByPrefix[c.CardID] = i
	}

	// 查 scope 内所有 monitor-on device + 三个 anchor + device meta
	prefixes := make([]string, 0, len(cards))
	for _, c := range cards {
		prefixes = append(prefixes, c.CardID)
	}

	// siblingCardExists：DB 内所有跟本批 cards 同 /80 unit 范围的卡 prefix 集合。
	// 用于 single-card 视图（cards 切片只有 1 张）时仍能识别同辈 bed/room card 存在 → 不走 merge。
	siblingCardExists := map[string]bool{}
	if len(prefixes) > 0 {
		// 同 /80 unit 范围内的兄弟卡集合（single-card 视图识别用）
		sibRows, err := s.db.QueryContext(ctx, `
			SELECT text(card_id)
			  FROM cards
			 WHERE EXISTS (
			   SELECT 1 FROM unnest($1::INET[]) p
			    WHERE cards.card_id <<= network(set_masklen(p, 80))
			       OR cards.card_id = network(set_masklen(p, 80))
			 )
		`, pq.Array(prefixes))
		if err == nil {
			defer sibRows.Close()
			for sibRows.Next() {
				var p string
				if sibRows.Scan(&p) == nil {
					siblingCardExists[p] = true
				}
			}
		}
	}
	// 注意 WHERE: 用 device 的 /80 unit 覆盖（不只是 cards prefix 范围）
	// 否则 room-anchor device 当对应 /88 room card 不存在（被同 room bed 吸收）时被过滤掉
	// 例: guest radar 在 /88 guest, 但 guest 只有 1 bed → 不出 /88 guest card → cards 集合无此 prefix
	//     若 WHERE 限 cards prefix 内，guest radar 落不到任何 cards prefix → 被滤掉 → bed 301 收不到
	// 修法: 用 set_masklen 把 cards prefix 截到 /80 unit，device 在任一 unit 内即纳入
	// device_id = UUID (外部对接) / device_ipv6 = INET host 文本（owlcare 内部 lookup 用，
	// 与 card:realtime:stream.devices map key + device:status:{IPv6} 一致）
	// [[feedback_api_ids_ipv6_only]]
	rows, err := s.db.QueryContext(ctx, `
		SELECT host(d.device_addr) AS device_addr,
		       dfm.device_uid,
		       COALESCE(dfm.device_code, ''),
		       COALESCE(dfm.device_uid, ''),
		       COALESCE(dfm.device_type::text, ''),
		       COALESCE(dfm.device_model, ''),
		       d.monitoring_enabled,
		       'offline'::text AS status,
		       COALESCE((SELECT b.bed_id::text  FROM beds  b  WHERE d.device_addr <<= b.bed_id  LIMIT 1), '') AS bed_anchor,
		       COALESCE((SELECT rm.room_id::text FROM rooms rm WHERE d.device_addr <<= rm.room_id LIMIT 1), '') AS room_anchor,
		       COALESCE((SELECT u.unit_id::text FROM units u  WHERE d.device_addr <<= u.unit_id  LIMIT 1), '') AS unit_anchor
		  FROM devices d
		  JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		 WHERE d.monitoring_enabled = TRUE
		   AND EXISTS (
		     SELECT 1 FROM unnest($1::INET[]) p
		      WHERE d.device_addr <<= network(set_masklen(p, 80))
		   )
	`, pq.Array(prefixes))
	if err != nil {
		return fmt.Errorf("query devices: %w", err)
	}
	defer rows.Close()

	type devRow struct {
		info        commoncard.DeviceInfo
		bedAnchor   string
		roomAnchor  string
		unitAnchor  string
	}
	devs := []devRow{}
	for rows.Next() {
		var dr devRow
		if err := rows.Scan(&dr.info.DeviceAddr, &dr.info.DeviceUID, &dr.info.DeviceCode,
			&dr.info.DeviceName, &dr.info.DeviceType, &dr.info.DeviceModel,
			&dr.info.MonitoringEnabled, &dr.info.Status,
			&dr.bedAnchor, &dr.roomAnchor, &dr.unitAnchor); err != nil {
			return fmt.Errorf("scan device: %w", err)
		}
		if dr.info.DeviceCode == "" {
			dr.info.DeviceCode = dr.info.DeviceUID
		}
		devs = append(devs, dr)
	}

	// 用 Redis device:status:{addr} 覆盖 SQL 占位 'offline' status（cardagg 是真相源）
	if s.realtime != nil {
		if reader := s.realtime.StateReader(); reader != nil {
			addrs := make([]string, 0, len(devs))
			for _, dr := range devs {
				if dr.info.DeviceAddr != "" {
					addrs = append(addrs, dr.info.DeviceAddr)
				}
			}
			online := FillDeviceOnlineStatusFromCardagg(ctx, reader, addrs, s.logger)
			for i := range devs {
				if st, ok := online[devs[i].info.DeviceAddr]; ok {
					devs[i].info.Status = st
				}
			}
		}
	}

	// 算每 room 内 bed card 数 + 该 room 唯一 bed prefix（用于吸收）
	// 用 siblingCardExists（DB 实际同辈 bed card），不只看本批 cards 切片
	roomBedCount := map[string]int{}    // room prefix → bed card 数
	roomSoleBed := map[string]string{}  // room prefix → 唯一 bed card prefix (仅 count=1 时有意义)
	for p := range siblingCardExists {
		if !strings.HasSuffix(p, "/96") {
			continue
		}
		roomPrefix := narrowPrefixToRoom(p)
		if roomPrefix == "" {
			continue
		}
		roomBedCount[roomPrefix]++
		roomSoleBed[roomPrefix] = p
	}

	// 索引 unit card / room card 是否真实存在
	unitCardOf := map[string]string{} // unit prefix → cards 表里那张 /80 (if exists)
	for _, c := range cards {
		if strings.HasSuffix(c.CardID, "/80") {
			unitCardOf[c.CardID] = c.CardID
		}
	}

	// 算 device owning card — 与 card_reconcile.go 2026-05-17 新规则一致：
	//
	//   bed-anchor device:
	//     DB 有 /96 bed card  → 归 /96
	//     else DB 有 /88 room card → 归 /88   ⭐ (room absorb bed 新规则)
	//     else                → 归 /80 unit
	//
	//   room-anchor device:
	//     DB 有 /88 room card → 归 /88   (room=1+hasRoomDev: /88 absorb bed)
	//     else 有 sole /96 bed → 归 /96   (room=1, no roomDev: /96 absorb room)
	//     else                → 归 /80 unit  (bed=0 case 上推)
	//
	//   unit-anchor device:
	//     归 /80 unit
	//
	// 最终 owning 必须命中 cardByPrefix 才写入（保留"只填本批次卡的 Devices"）。
	deviceMap := map[string][]commoncard.DeviceInfo{} // card prefix → []DeviceInfo
	deviceRoomSet := map[string]map[string]bool{}     // card prefix → distinct room set (用于 unit-card coverage_label)
	for _, dr := range devs {
		owning := ""
		switch {
		case dr.bedAnchor != "":
			roomPrefix := narrowPrefixToRoom(dr.bedAnchor)
			switch {
			case siblingCardExists[dr.bedAnchor]:
				owning = dr.bedAnchor
			case roomPrefix != "" && siblingCardExists[roomPrefix]:
				owning = roomPrefix
			default:
				owning = dr.unitAnchor
			}
		case dr.roomAnchor != "":
			switch {
			case siblingCardExists[dr.roomAnchor]:
				owning = dr.roomAnchor
			case roomBedCount[dr.roomAnchor] == 1:
				owning = roomSoleBed[dr.roomAnchor]
			default:
				owning = dr.unitAnchor
			}
		case dr.unitAnchor != "":
			owning = dr.unitAnchor
		}
		if owning == "" {
			continue
		}
		if _, ok := cardByPrefix[owning]; !ok {
			continue // owning 不在本次请求的 cards 切片 → 不写回（避免单卡查询把别人的设备塞进来）
		}
		deviceMap[owning] = append(deviceMap[owning], dr.info)
		// 记 distinct room set for coverage_label（按 device 实际来源 room）
		var deviceRoom string
		if dr.roomAnchor != "" {
			deviceRoom = dr.roomAnchor
		} else if dr.bedAnchor != "" {
			deviceRoom = narrowPrefixToRoom(dr.bedAnchor)
		}
		if deviceRoom != "" {
			if deviceRoomSet[owning] == nil {
				deviceRoomSet[owning] = map[string]bool{}
			}
			deviceRoomSet[owning][deviceRoom] = true
		}
	}

	for i := range cards {
		if devs, ok := deviceMap[cards[i].CardID]; ok {
			cards[i].Devices = devs
		}
	}
	return nil
}

// GetCardCaregivers 查询单张卡片的护理人员（按 residents 聚合）
// Detail 页进入时调用，不在列表查询中批量加载。
//
// v2 schema:
//   - resident_id 是 INET (HoA /128)，不是 UUID
//   - resident_caregivers 拆成 caregiver_id / care_team_id / family_id 三列（一行只 set 一种），
//     不再有 v1 的 group_list / user_list jsonb 字段
//   - care_team_id → care_teams.team_id 取 team_name 当 group label
//   - caregiver_id / family_id → users.user_id 取详情
func (s *CardStaticService) GetCardCaregivers(ctx context.Context, residentIDs []string) (groups []string, caregivers []commoncard.CaregiverInfo, err error) {
	if len(residentIDs) == 0 {
		return nil, nil, nil
	}

	// 1. 查 resident_caregivers — v2: 三列分别 caregiver / team / family；只取 active (valid_to IS NULL)
	rows, err := s.db.QueryContext(ctx,
		`SELECT caregiver_id::text, care_team_id::text, family_id::text
		   FROM resident_caregivers
		  WHERE resident_id = ANY($1::inet[])
		    AND valid_to IS NULL`,
		pq.Array(residentIDs),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("query resident_caregivers: %w", err)
	}
	defer rows.Close()

	teamIDSet := map[string]bool{}
	userIDSet := map[string]bool{} // caregiver_id ∪ family_id → users
	for rows.Next() {
		var cgID, teamID, famID sql.NullString
		if rows.Scan(&cgID, &teamID, &famID) != nil {
			continue
		}
		if teamID.Valid && teamID.String != "" {
			teamIDSet[teamID.String] = true
		}
		if cgID.Valid && cgID.String != "" {
			userIDSet[cgID.String] = true
		}
		if famID.Valid && famID.String != "" {
			userIDSet[famID.String] = true
		}
	}

	// 2. care_teams.team_name 作为 group label 输出
	if len(teamIDSet) > 0 {
		tids := make([]string, 0, len(teamIDSet))
		for tid := range teamIDSet {
			tids = append(tids, tid)
		}
		tRows, tErr := s.db.QueryContext(ctx,
			`SELECT COALESCE(team_name, '') FROM care_teams WHERE team_id = ANY($1::uuid[])`,
			pq.Array(tids),
		)
		if tErr == nil {
			defer tRows.Close()
			for tRows.Next() {
				var name string
				if tRows.Scan(&name) == nil && name != "" {
					groups = append(groups, name)
				}
			}
		}
	}

	// 3. caregiver/family users 详情
	if len(userIDSet) > 0 {
		uids := make([]string, 0, len(userIDSet))
		for uid := range userIDSet {
			uids = append(uids, uid)
		}
		uRows, err := s.db.QueryContext(ctx,
			`SELECT user_id::text, COALESCE(nickname,''), COALESCE(user_account,''), COALESCE(role,'')
			 FROM users WHERE user_id = ANY($1::uuid[])`,
			pq.Array(uids),
		)
		if err == nil {
			defer uRows.Close()
			for uRows.Next() {
				var ci commoncard.CaregiverInfo
				if uRows.Scan(&ci.UserID, &ci.Nickname, &ci.UserAccount, &ci.Role) == nil {
					caregivers = append(caregivers, ci)
				}
			}
		}
	}

	return groups, caregivers, nil
}
