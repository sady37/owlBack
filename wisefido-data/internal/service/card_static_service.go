package service

import (
	"context"
	"database/sql"
	"encoding/json"
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
func (s *CardStaticService) GetCardInfo(ctx context.Context, tenantID, userID, cardID string) (*commoncard.CardStatic, error) {
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
	if !allowed[cardID] {
		return nil, fmt.Errorf("card not found or no permission")
	}
	cards, _, err := s.queryCardsByIDs(ctx, []string{cardID}, nil, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, fmt.Errorf("card not found")
	}
	return &cards[0], nil
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
	query := `
		SELECT
			text(c.spatial_prefix) AS card_id,
			text(c.spatial_prefix) AS spatial_prefix,
			c.card_type,
			COALESCE(c.card_name, ''),
			COALESCE(c.dns_short_name, ''),
			COALESCE(text(c.resident_id), ''),
			COALESCE(u.unit_id::text, ''),
			COALESCE(u.unit_name, ''),
			COALESCE(u.timezone, ''),
			COALESCE(br.branch_name, ''),
			COALESCE(rm.room_name, '') AS room_name,
			COALESCE(b.bed_name, '')   AS bed_name,
			COALESCE(s.site_name, '')  AS building_name,
			COUNT(*) OVER() AS total_count
		FROM cards c
		-- 用 network(set_masklen(...)) 真正截断到对应粒度的 network address
		-- (set_masklen 单用只换 mask 不清 host bits，会 join 不上)
		LEFT JOIN units    u  ON u.unit_id   = network(set_masklen(c.spatial_prefix, 80))
		LEFT JOIN branches br ON br.branch_id = network(set_masklen(c.spatial_prefix, 56))
		LEFT JOIN sites    s  ON s.site_id   = network(set_masklen(c.spatial_prefix, 64))
		LEFT JOIN rooms    rm ON masklen(c.spatial_prefix) >= 88 AND rm.room_id = network(set_masklen(c.spatial_prefix, 88))
		LEFT JOIN beds     b  ON masklen(c.spatial_prefix) = 96  AND b.bed_id   = c.spatial_prefix
		WHERE c.spatial_prefix = ANY($1::inet[])
		  AND c.card_type <> 'device'
	`
	args := []any{pq.Array(cardIDs)}
	argIdx := 2

	if len(branchIDs) > 0 {
		query += fmt.Sprintf(` AND c.spatial_prefix <<= ANY($%d::inet[])`, argIdx)
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
			cardID, spatialPrefix, cardType, cardName, dnsShortName, residentID  string
			unitID, unitName, timezone, branchName, roomName, bedName, buildName string
		)
		if err := rows.Scan(
			&cardID, &spatialPrefix, &cardType, &cardName, &dnsShortName, &residentID,
			&unitID, &unitName, &timezone, &branchName, &roomName, &bedName, &buildName,
			&totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan card row: %w", err)
		}

		unit := &commoncard.UnitInfo{
			UnitID:       unitID,
			UnitName:     unitName,
			BranchName:   branchName,
			BuildingName: buildName, // sites.site_name → Detail 页 Branch/Building/Unit 路径用
			RoomName:     roomName,  // /88 /96 card 才非空
			BedName:      bedName,   // /96 card 才非空
			Timezone:     timezone,
		}
		c := commoncard.CardStatic{
			CardID:        cardID,
			CardType:      cardType,
			CardName:      cardName,
			DNSShortName:  dnsShortName,
			SpatialPrefix: spatialPrefix,
			Unit:          unit,
			CardAddress:   composeCardAddress(unitName, roomName, bedName),
		}
		if residentID != "" {
			// active_bed 卡的 residents 占位（nickname 等由 enrichResidentsAndDevices 补）
			c.Residents = []commoncard.ResidentInfo{{ResidentID: residentID}}
		}
		if bedName != "" {
			bn := bedName
			c.BedName = &bn
		}
		if roomName != "" {
			c.Rooms = []commoncard.RoomIdentifier{{RoomName: roomName}}
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

// composeCardAddress 按最长 mask 拼 "<unit> / <room> / <bed>" 友好串；空段跳过。
// 去单位词前缀（"Unit/Room/Bed"）— 单位信息由 card_type 暴露；分隔符 " / " 隐含层级。
// /80 → "101"；/88 → "101 / Guest"；/96 → "101 / Guest / BedA"
// FE 紧凑视图（grid card）可用独立字段 unit.room_name / unit.bed_name 自由拼接。
func composeCardAddress(unit, room, bed string) string {
	parts := []string{}
	if unit != "" {
		parts = append(parts, unit)
	}
	if room != "" {
		parts = append(parts, room)
	}
	if bed != "" {
		parts = append(parts, bed)
	}
	return strings.Join(parts, " / ")
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
	residentMap := make(map[string]string) // resident_id → nickname
	if len(residentIDs) > 0 {
		rows, err := s.db.QueryContext(ctx,
			`SELECT resident_id::text, COALESCE(nickname, '') FROM residents WHERE resident_id = ANY($1::INET[])`,
			pq.Array(residentIDs),
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var rid, nick string
				if rows.Scan(&rid, &nick) == nil {
					residentMap[rid] = nick
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

	// === 3. apply residents nickname ===
	for i := range cards {
		for j := range cards[i].Residents {
			rid := cards[i].Residents[j].ResidentID
			if nick, ok := residentMap[rid]; ok {
				cards[i].Residents[j].Nickname = nick
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
	cardByPrefix := map[string]int{} // prefix → index in cards
	for i, c := range cards {
		cardByPrefix[c.SpatialPrefix] = i
	}

	// 查 scope 内所有 monitor-on device + 三个 anchor + device meta
	prefixes := make([]string, 0, len(cards))
	for _, c := range cards {
		prefixes = append(prefixes, c.SpatialPrefix)
	}
	// 注意 WHERE: 用 device 的 /80 unit 覆盖（不只是 cards prefix 范围）
	// 否则 room-anchor device 当对应 /88 room card 不存在（被同 room bed 吸收）时被过滤掉
	// 例: guest radar 在 /88 guest, 但 guest 只有 1 bed → 不出 /88 guest card → cards 集合无此 prefix
	//     若 WHERE 限 cards prefix 内，guest radar 落不到任何 cards prefix → 被滤掉 → bed 301 收不到
	// 修法: 用 set_masklen 把 cards prefix 截到 /80 unit，device 在任一 unit 内即纳入
	rows, err := s.db.QueryContext(ctx, `
		SELECT dfm.device_id::text,
		       dfm.device_uid,
		       COALESCE(dfm.device_code, ''),
		       COALESCE(dfm.device_uid, ''),
		       COALESCE(dfm.device_type::text, ''),
		       COALESCE(dfm.device_model, ''),
		       COALESCE((SELECT b.bed_id::text  FROM beds  b  WHERE d.device_ipv6 <<= b.bed_id  LIMIT 1), '') AS bed_anchor,
		       COALESCE((SELECT rm.room_id::text FROM rooms rm WHERE d.device_ipv6 <<= rm.room_id LIMIT 1), '') AS room_anchor,
		       COALESCE((SELECT u.unit_id::text FROM units u  WHERE d.device_ipv6 <<= u.unit_id  LIMIT 1), '') AS unit_anchor
		  FROM devices d
		  JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		 WHERE d.monitoring_enabled = TRUE
		   AND EXISTS (
		     SELECT 1 FROM unnest($1::INET[]) p
		      WHERE d.device_ipv6 <<= network(set_masklen(p, 80))
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
		if err := rows.Scan(&dr.info.DeviceID, &dr.info.DeviceUID, &dr.info.DeviceCode,
			&dr.info.DeviceName, &dr.info.DeviceType, &dr.info.DeviceModel,
			&dr.bedAnchor, &dr.roomAnchor, &dr.unitAnchor); err != nil {
			return fmt.Errorf("scan device: %w", err)
		}
		if dr.info.DeviceCode == "" {
			dr.info.DeviceCode = dr.info.DeviceUID
		}
		devs = append(devs, dr)
	}

	// 算每 room 内 bed card 数 + 该 room 唯一 bed prefix（用于吸收）
	roomBedCount := map[string]int{}      // room prefix → bed card 数
	roomSoleBed  := map[string]string{}   // room prefix → 唯一 bed card prefix (仅 count=1 时有意义)
	for _, c := range cards {
		if !strings.HasSuffix(c.SpatialPrefix, "/96") {
			continue
		}
		// 算 bed 所属 room
		roomPrefix := narrowPrefixToRoom(c.SpatialPrefix)
		if roomPrefix == "" {
			continue
		}
		roomBedCount[roomPrefix]++
		roomSoleBed[roomPrefix] = c.SpatialPrefix
	}

	// 索引 unit card / room card 是否真实存在
	unitCardOf := map[string]string{} // unit prefix → cards 表里那张 /80 (if exists)
	for _, c := range cards {
		if strings.HasSuffix(c.SpatialPrefix, "/80") {
			unitCardOf[c.SpatialPrefix] = c.SpatialPrefix
		}
	}

	// 算 device owning card
	deviceMap := map[string][]commoncard.DeviceInfo{} // card prefix → []DeviceInfo
	deviceRoomSet := map[string]map[string]bool{}     // card prefix → distinct room set (用于 unit-card coverage_label)
	for _, dr := range devs {
		owning := ""
		// 1. bed anchor
		if dr.bedAnchor != "" {
			owning = dr.bedAnchor
		} else if dr.roomAnchor != "" {
			// 2. room anchor — 看 room 内 active_bed_count
			if roomBedCount[dr.roomAnchor] == 1 {
				// 吸收到唯一 bed
				owning = roomSoleBed[dr.roomAnchor]
			} else if _, ok := cardByPrefix[dr.roomAnchor]; ok {
				// /88 room card 存在
				owning = dr.roomAnchor
			} else {
				// 兜底到 /80 unit
				owning = dr.unitAnchor
			}
		} else if dr.unitAnchor != "" {
			// 3. unit anchor
			owning = dr.unitAnchor
		}
		if owning == "" {
			continue
		}
		if _, ok := cardByPrefix[owning]; !ok {
			continue // 该 prefix 没 card → skip
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

	// 写回 cards[].Devices + cards[].CoverageLabel
	for i := range cards {
		if devs, ok := deviceMap[cards[i].SpatialPrefix]; ok {
			cards[i].Devices = devs
		}
		cards[i].CoverageLabel = s.computeCoverageLabel(ctx, &cards[i], deviceRoomSet[cards[i].SpatialPrefix])
	}
	return nil
}

// computeCoverageLabel — 卡行 2 标签
//   bed card  → ""（FE 用 unit.room_name + bed_name 拼）
//   room card → ""（FE 用 unit.room_name）
//   unit card → 看该卡装的 device 跨多少 distinct room：
//                 1 room → 该 room name；≥2 room → "Whole Unit"；0 room → ""
func (s *CardStaticService) computeCoverageLabel(ctx context.Context, c *commoncard.CardStatic, rooms map[string]bool) string {
	if c.CardType != "unit" {
		return ""
	}
	if len(rooms) == 0 {
		return ""
	}
	if len(rooms) >= 2 {
		return "Whole Unit"
	}
	// 1 room — 反查 room_name
	var roomPrefix string
	for r := range rooms {
		roomPrefix = r
		break
	}
	var roomName sql.NullString
	_ = s.db.QueryRowContext(ctx,
		`SELECT room_name FROM rooms WHERE room_id = $1::INET LIMIT 1`,
		roomPrefix).Scan(&roomName)
	if roomName.Valid && roomName.String != "" {
		return roomName.String
	}
	return ""
}

// GetCardCaregivers 查询单张卡片的护理人员（按 residents 聚合）
// Detail 页进入时调用，不在列表查询中批量加载
func (s *CardStaticService) GetCardCaregivers(ctx context.Context, residentIDs []string) (groups []string, caregivers []commoncard.CaregiverInfo, err error) {
	if len(residentIDs) == 0 {
		return nil, nil, nil
	}

	// 1. 查 resident_caregivers
	rows, err := s.db.QueryContext(ctx,
		`SELECT group_list, user_list
		 FROM resident_caregivers
		 WHERE resident_id = ANY($1::uuid[])`,
		pq.Array(residentIDs),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("query resident_caregivers: %w", err)
	}
	defer rows.Close()

	groupSet := map[string]bool{}
	userIDSet := map[string]bool{}
	for rows.Next() {
		var groupJSON, userJSON []byte
		if rows.Scan(&groupJSON, &userJSON) != nil {
			continue
		}
		var gl, ul []string
		if len(groupJSON) > 0 {
			json.Unmarshal(groupJSON, &gl)
		}
		if len(userJSON) > 0 {
			json.Unmarshal(userJSON, &ul)
		}
		for _, g := range gl {
			groupSet[g] = true
		}
		for _, uid := range ul {
			userIDSet[uid] = true
		}
	}

	// 2. groups 去重输出
	for g := range groupSet {
		groups = append(groups, g)
	}

	// 3. 查 users 详情
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
