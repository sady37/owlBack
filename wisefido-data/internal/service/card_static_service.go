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
			COUNT(*) OVER() AS total_count
		FROM cards c
		-- 用 network(set_masklen(...)) 真正截断到对应粒度的 network address
		-- (set_masklen 单用只换 mask 不清 host bits，会 join 不上)
		LEFT JOIN units    u  ON u.unit_id   = network(set_masklen(c.spatial_prefix, 80))
		LEFT JOIN branches br ON br.branch_id = network(set_masklen(c.spatial_prefix, 56))
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

	query += ` ORDER BY c.card_name ASC NULLS LAST`

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
			unitID, unitName, timezone, branchName, roomName, bedName           string
		)
		if err := rows.Scan(
			&cardID, &spatialPrefix, &cardType, &cardName, &dnsShortName, &residentID,
			&unitID, &unitName, &timezone, &branchName, &roomName, &bedName,
			&totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan card row: %w", err)
		}

		unit := &commoncard.UnitInfo{
			UnitID:     unitID,
			UnitName:   unitName,
			BranchName: branchName,
			Timezone:   timezone,
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

// composeCardAddress 按最长 mask 拼 "Unit X / Room Y / Bed Z" 友好串；空段跳过。
// /80 → "Unit 101"；/88 → "Unit 101 / Room A"；/96 → "Unit 101 / Room A / Bed 1"
func composeCardAddress(unit, room, bed string) string {
	parts := []string{}
	if unit != "" {
		parts = append(parts, "Unit "+unit)
	}
	if room != "" {
		parts = append(parts, "Room "+room)
	}
	if bed != "" {
		parts = append(parts, "Bed "+bed)
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
	// === 2. 批量 fill devices via LPM ===
	prefixes := make([]string, 0, len(cards))
	for _, c := range cards {
		prefixes = append(prefixes, c.SpatialPrefix)
	}
	deviceMap := make(map[string][]commoncard.DeviceInfo) // spatial_prefix → []DeviceInfo
	if len(prefixes) > 0 {
		drows, err := s.db.QueryContext(ctx, `
			SELECT c.spatial_prefix::text,
			       dfm.device_id::text,
			       dfm.device_uid,
			       COALESCE(dfm.device_code, ''),
			       COALESCE(dfm.device_uid, '') AS device_name,
			       COALESCE(dfm.device_type::text, ''),
			       COALESCE(dfm.device_model, '')
			  FROM cards c
			  JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
			  JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
			 WHERE c.spatial_prefix = ANY($1::INET[])
		`, pq.Array(prefixes))
		if err == nil {
			defer drows.Close()
			for drows.Next() {
				var cardPrefix string
				var di commoncard.DeviceInfo
				if drows.Scan(&cardPrefix, &di.DeviceID, &di.DeviceUID, &di.DeviceCode,
					&di.DeviceName, &di.DeviceType, &di.DeviceModel) == nil {
					if di.DeviceCode == "" {
						di.DeviceCode = di.DeviceUID
					}
					deviceMap[cardPrefix] = append(deviceMap[cardPrefix], di)
				}
			}
		}
	}
	// === 3. apply back ===
	for i := range cards {
		for j := range cards[i].Residents {
			rid := cards[i].Residents[j].ResidentID
			if nick, ok := residentMap[rid]; ok {
				cards[i].Residents[j].Nickname = nick
			}
		}
		if devs, ok := deviceMap[cards[i].SpatialPrefix]; ok {
			cards[i].Devices = devs
		}
	}
	return nil
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
