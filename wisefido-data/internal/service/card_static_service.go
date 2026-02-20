package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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
func (s *CardStaticService) queryCardsByIDs(ctx context.Context, cardIDs []string, branchIDs []string, page, pageSize int) ([]commoncard.CardStatic, int, error) {
	// 构建查询
	query := `
		SELECT
			c.card_id::text, c.tenant_id::text, c.card_type, c.card_name, c.card_address,
			c.bed_id::text, c.unit_id::text, c.timezone,
			c.devices, c.residents,
			COALESCE(c.icon_alarm_level, 2), COALESCE(c.pop_alarm, 2),
			COALESCE(u.branch_id::text, '')       AS branch_id,
			COALESCE(b.branch_name, '')           AS branch_name,
			COALESCE(u.unit_name, '')             AS unit_name,
			COALESCE(bld.building_name, '')       AS building,
			COALESCE(u.is_public, false)          AS is_public,
			COALESCE(u.is_shared_unit, false)     AS is_shared_unit,
			COALESCE(u.unit_type, '')             AS unit_type,
			COALESCE(bed.bed_name, '')            AS bed_name,
			COALESCE(room.room_id::text, '')      AS room_id,
			COALESCE(room.room_name, '')          AS room_name,
			COUNT(*) OVER()                       AS total_count
		FROM cards c
		LEFT JOIN units u      ON c.unit_id = u.unit_id
		LEFT JOIN branches b   ON u.branch_id = b.branch_id
		LEFT JOIN buildings bld ON u.building_id = bld.building_id
		LEFT JOIN beds bed     ON c.bed_id = bed.bed_id
		LEFT JOIN rooms room   ON bed.room_id = room.room_id
		WHERE c.card_id = ANY($1::uuid[])
	`
	args := []any{pq.Array(cardIDs)}
	argIdx := 2

	if len(branchIDs) > 0 {
		query += fmt.Sprintf(` AND u.branch_id = ANY($%d::uuid[])`, argIdx)
		args = append(args, pq.Array(branchIDs))
		argIdx++
	}

	query += ` ORDER BY c.card_address ASC`

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
			cardID, tenantID, cardType, cardName, cardAddress string
			bedID, unitID, timezone                           sql.NullString
			devicesJSON, residentsJSON                        []byte
			iconAlarmLevel, popAlarm                          int
			branchID, branchName, unitName                    string
			building, unitType                                string
			isPublic, isSharedUnit                            bool
			bedName, roomID, roomName                         string
		)

		if err := rows.Scan(
			&cardID, &tenantID, &cardType, &cardName, &cardAddress,
			&bedID, &unitID, &timezone,
			&devicesJSON, &residentsJSON,
			&iconAlarmLevel, &popAlarm,
			&branchID, &branchName, &unitName,
			&building, &isPublic, &isSharedUnit, &unitType,
			&bedName, &roomID, &roomName,
			&totalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan card row: %w", err)
		}

		unit := &commoncard.UnitInfo{
			BranchID:     branchID,
			BranchName:   branchName,
			Building:     building,
			IsPublic:     isPublic,
			IsSharedUnit: isSharedUnit,
			UnitType:     unitType,
			Timezone:     timezone.String,
		}
		if unitID.Valid {
			unit.UnitID = unitID.String
			unit.UnitName = unitName
		}
		card := commoncard.CardStatic{
			CardID:      cardID,
			TenantID:    tenantID,
			CardType:    cardType,
			CardName:    cardName,
			CardAddress: cardAddress,
			Unit:        unit,
		}
		if bedID.Valid {
			card.BedID = &bedID.String
			if bedName != "" {
				card.BedName = &bedName
			}
		}
		if roomID != "" {
			card.Rooms = []commoncard.RoomIdentifier{{RoomID: roomID, RoomName: roomName}}
		}
		if iconAlarmLevel > 0 {
			card.IconAlarmLevel = &iconAlarmLevel
		}
		if popAlarm > 0 {
			card.PopAlarm = &popAlarm
		}

		// 展开 devices JSONB
		if len(devicesJSON) > 0 {
			var devices []commoncard.DeviceInfo
			if json.Unmarshal(devicesJSON, &devices) == nil {
				card.Devices = devices
			}
		}
		// 展开 residents JSONB
		if len(residentsJSON) > 0 {
			var residents []commoncard.ResidentInfo
			if json.Unmarshal(residentsJSON, &residents) == nil {
				card.Residents = residents
			}
		}

		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rows: %w", err)
	}

	return cards, totalCount, nil
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
