package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	commoncard "owl-common/card"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/store"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// CardList 统一返回结构，按 branch 分组
type CardList struct {
	UserID        string              `json:"user_id"`
	TenantID      string              `json:"tenant_id"`
	CardsByBranch map[string][]string `json:"cards_by_branch"` // branchID → card_id[]（无 branch 用 "_"）
}

// AllCardIDs 展平所有 branch 下的 card_id；nil 安全
func (cl *CardList) AllCardIDs() []string {
	if cl == nil {
		return nil
	}
	var ids []string
	for _, cids := range cl.CardsByBranch {
		ids = append(ids, cids...)
	}
	return ids
}

// BranchIDs 返回去重的 branch_id 列表（不含 "_"）；nil 安全
func (cl *CardList) BranchIDs() []string {
	if cl == nil {
		return nil
	}
	var out []string
	for bid := range cl.CardsByBranch {
		if bid != "_" {
			out = append(out, bid)
		}
	}
	return out
}

// AllowedCardIDsProviderImpl 实现 AllowedCardIDsProvider
type AllowedCardIDsProviderImpl struct {
	kv            store.KV
	usersRepo     repository.UsersRepository
	residentsRepo *repository.PostgresResidentsRepository
	db            *sql.DB
	cardRepo      *repository.PostgresCardRepository
	logger        *zap.Logger
}

// AllowedCardIDsProvider 对外接口
type AllowedCardIDsProvider interface {
	GetCardList(ctx context.Context, tenantID, userID, userType string) (*CardList, error)
}

// NewAllowedCardIDsProvider 创建实现
func NewAllowedCardIDsProvider(
	kv store.KV,
	usersRepo repository.UsersRepository,
	residentsRepo *repository.PostgresResidentsRepository,
	db *sql.DB,
	cardRepo *repository.PostgresCardRepository,
	logger *zap.Logger,
) *AllowedCardIDsProviderImpl {
	return &AllowedCardIDsProviderImpl{
		kv:            kv,
		usersRepo:     usersRepo,
		residentsRepo: residentsRepo,
		db:            db,
		cardRepo:      cardRepo,
		logger:        logger,
	}
}

// GetAllowedCardIDsByBranches 返回按 branch 分组的卡片 ID
func (p *AllowedCardIDsProviderImpl) GetAllowedCardIDsByBranches(ctx context.Context, tenantID, userID, userType string) (map[string][]string, error) {
	cl, err := p.GetCardList(ctx, tenantID, userID, userType)
	if err != nil {
		return nil, err
	}
	if cl == nil {
		return make(map[string][]string), nil
	}
	return cl.CardsByBranch, nil
}

// GetCardList 统一入口，返回 *CardList
// relegation 和 role 从 DB (usersRepo.GetUser) 获取，不需外部传入
func (p *AllowedCardIDsProviderImpl) GetCardList(ctx context.Context, tenantID, userID, userType string) (*CardList, error) {
	if userType == "resident" {
		return p.filterCardsForResident(ctx, userID, tenantID)
	}
	return p.filterCardsForStaff(ctx, userID, tenantID)
}


// filterCardsForResident 按 resident→unit→card 过滤，返回 *CardList
func (p *AllowedCardIDsProviderImpl) filterCardsForResident(ctx context.Context, residentID, tenantID string) (*CardList, error) {
	resident, err := p.residentsRepo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		return nil, err
	}
	if resident.UnitID == nil || *resident.UnitID == "" {
		return nil, nil
	}

	unitInfo, err := p.cardRepo.GetUnitInfo(tenantID, *resident.UnitID)
	if err != nil {
		p.logger.Warn("filterCardsForResident GetUnitInfo failed", zap.String("unit_id", *resident.UnitID), zap.Error(err))
		return nil, nil
	}

	// Home unit → 该 unit 下全部 card
	if unitInfo.UnitProperty == commoncard.UnitPropertyHome {
		return p.cardIDsByUnit(ctx, tenantID, *resident.UnitID, residentID)
	}

	// Facility + public → nil
	if unitInfo.IsPublic {
		return nil, nil
	}

	// Facility + not public + not shared → 该 unit 下全部 card
	if !unitInfo.IsSharedUnit {
		return p.cardIDsByUnit(ctx, tenantID, *resident.UnitID, residentID)
	}

	// facility + not public + shared → ActiveBedCard only
	return p.ActiveBedcardIDsByUnitShared(ctx, tenantID, *resident.UnitID, residentID)
}

// cardIDsByUnit 查该 unit 下所有 card，返回 *CardList
// v2 unified: card_id ≡ spatial_prefix；tenantID 是 /48 INET，unitID 是 /80 INET
func (p *AllowedCardIDsProviderImpl) cardIDsByUnit(ctx context.Context, tenantID, unitID, userID string) (*CardList, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT c.card_id::text AS card_id,
		        host(set_masklen(c.card_id, 56)) || '/56' AS branch_id
		   FROM cards c
		  WHERE c.card_id <<= $1::INET
		    AND c.card_id <<= $2::INET`,
		tenantID, unitID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byBranch := make(map[string][]string)
	for rows.Next() {
		var cardID, branchID string
		if rows.Scan(&cardID, &branchID) == nil {
			if branchID == "" {
				branchID = "_"
			}
			byBranch[branchID] = append(byBranch[branchID], cardID)
		}
	}
	return &CardList{UserID: userID, TenantID: tenantID, CardsByBranch: byBranch}, nil
}

// ActiveBedcardIDsByUnitShared shared unit：只返回 active_bed 且 resident_id 匹配的卡。
// v2: cards.resident_id 是 INET HoA pointer；不再走 residents JSONB；unitID 是 INET /80 CIDR。
func (p *AllowedCardIDsProviderImpl) ActiveBedcardIDsByUnitShared(ctx context.Context, tenantID, unitID, residentID string) (*CardList, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT c.card_id::text AS card_id,
		        host(set_masklen(c.card_id, 56)) || '/56' AS branch_id
		   FROM cards c
		  WHERE c.card_id <<= $1::INET
		    AND c.card_id <<= $2::INET
		    AND c.has_bed = TRUE
		    AND c.resident_id = $3::INET`,
		tenantID, unitID, residentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byBranch := make(map[string][]string)
	for rows.Next() {
		var cardID, branchID string
		if rows.Scan(&cardID, &branchID) == nil {
			if branchID == "" {
				branchID = "_"
			}
			byBranch[branchID] = append(byBranch[branchID], cardID)
		}
	}
	return &CardList{UserID: residentID, TenantID: tenantID, CardsByBranch: byBranch}, nil
}

// filterCardsForStaff 按 users.relegation 直接查 DB，返回 *CardList
// ALL / Admin = 该 tenant 全部卡片
// ASSIGNED_ONLY = 仅分配的住户关联卡片
// BRANCH = 仅所在院区的卡片
func (p *AllowedCardIDsProviderImpl) filterCardsForStaff(ctx context.Context, userID, tenantID string) (*CardList, error) {
	if p.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	user, err := p.usersRepo.GetUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}

	// scope 大小写宽松匹配（DB 里 demo Manager 存 'branch' 小写，老种子可能存 'BRANCH'/'Branch'）
	scope := ""
	if user.Relegation.Valid && user.Relegation.String != "" {
		scope = strings.ToUpper(strings.TrimSpace(user.Relegation.String))
	}
	if strings.EqualFold(user.Role, "Admin") || scope == "ALL" {
		return p.filterTenantCards(ctx, tenantID, userID)
	}
	// Family 角色 — 不论 relegation 都强制走 ASSIGNED（业务定义：family 只看自己亲属相关 cards）
	// 这避免 admin 误设 relegation=branch 让 family 看到整 branch 卡
	if strings.EqualFold(user.Role, "Family") || strings.EqualFold(user.Role, "family") {
		return p.filterByAssignedOnly(ctx, tenantID, userID)
	}
	switch scope {
	case "ASSIGNED_ONLY", "ASSIGNED":
		return p.filterByAssignedOnly(ctx, tenantID, userID)
	case "BRANCH":
		return p.filterByBranchOnly(ctx, tenantID, userID)
	default:
		p.logger.Warn("filterCardsForStaff: unknown scope, returning empty",
			zap.String("user_id", userID), zap.String("role", user.Role),
			zap.String("scope", scope))
		return &CardList{UserID: userID, TenantID: tenantID, CardsByBranch: make(map[string][]string)}, nil
	}
}

// filterTenantCards 查该 tenant 全部卡片，按 branch 分组返回 *CardList。
// v2: tenant_id INET CIDR /48；branch /56 由 spatial_prefix 派生。
func (p *AllowedCardIDsProviderImpl) filterTenantCards(ctx context.Context, tenantID, userID string) (*CardList, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT c.card_id::text AS card_id,
		        host(set_masklen(c.card_id, 56)) || '/56' AS branch_id
		   FROM cards c
		  WHERE c.card_id <<= $1::INET`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("filterTenantCards: %w", err)
	}
	defer rows.Close()

	byBranch := make(map[string][]string)
	for rows.Next() {
		var cardID, branchID string
		if err := rows.Scan(&cardID, &branchID); err != nil {
			return nil, fmt.Errorf("filterTenantCards scan: %w", err)
		}
		if branchID == "" {
			branchID = "_"
		}
		byBranch[branchID] = append(byBranch[branchID], cardID)
	}
	return &CardList{UserID: userID, TenantID: tenantID, CardsByBranch: byBranch}, nil
}


// filterByBranchOnly BRANCH scope —— Phase 3：只返 user.is_primary Current Branch 内的 card
// v2 schema：cards 没冗余 branch_id 列；用 spatial_prefix /56 反推 + user_branches.is_primary INET 过滤
func (p *AllowedCardIDsProviderImpl) filterByBranchOnly(ctx context.Context, tenantID, userID string) (*CardList, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT c.card_id::text AS card_id,
		        host(network(set_masklen(c.card_id, 56))) || '/56' AS branch_id
		   FROM cards c
		  WHERE c.card_id <<= $1::INET
		    AND masklen(c.card_id) <> 128
		    AND EXISTS (
		      SELECT 1 FROM user_branches ub
		       WHERE ub.user_id = $2::UUID
		         AND ub.is_primary = TRUE
		         AND ub.valid_to IS NULL
		         AND c.card_id <<= ub.branch_prefix
		    )`,
		tenantID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("filterByBranchOnly: %w", err)
	}
	defer rows.Close()

	byBranch := make(map[string][]string)
	for rows.Next() {
		var cardID, branchID string
		if err := rows.Scan(&cardID, &branchID); err != nil {
			return nil, fmt.Errorf("filterByBranchOnly scan: %w", err)
		}
		if branchID == "" {
			branchID = "_"
		}
		byBranch[branchID] = append(byBranch[branchID], cardID)
	}
	return &CardList{UserID: userID, TenantID: tenantID, CardsByBranch: byBranch}, nil
}

// filterByAssignedOnly ASSIGNED_ONLY scope：
// 1. 查 resident_caregivers (v2 列式 schema：caregiver_id / family_id / care_team_id) 得到该 user 关联的 resident_id[]
//    覆盖三条路径：直接 caregiver、family 关系、care_team 成员
// 2. 返 cards 中 LPM 双向 overlap 到 assigned resident 的所有 cards（不限 active_bed，含 fallback 覆盖的 unit/room cards）
func (p *AllowedCardIDsProviderImpl) filterByAssignedOnly(ctx context.Context, tenantID, userID string) (*CardList, error) {
	// Step1: v2 列式 — 该 user 关联的 resident_id[]
	rows, err := p.db.QueryContext(ctx,
		`SELECT DISTINCT rc.resident_id::text
		   FROM resident_caregivers rc
		  WHERE rc.valid_to IS NULL
		    AND rc.resident_id <<= $1::INET
		    AND (
		      rc.caregiver_id = $2::UUID
		      OR rc.family_id   = $2::UUID
		      OR rc.care_team_id IN (
		        SELECT ctm.team_id FROM care_team_members ctm
		         WHERE ctm.user_id = $2::UUID AND ctm.valid_to IS NULL
		      )
		    )`,
		tenantID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("filterByAssignedOnly caregivers: %w", err)
	}
	defer rows.Close()
	var rids []string
	for rows.Next() {
		var rid string
		if rows.Scan(&rid) == nil && rid != "" {
			rids = append(rids, rid)
		}
	}
	if len(rids) == 0 {
		return &CardList{UserID: userID, TenantID: tenantID, CardsByBranch: make(map[string][]string)}, nil
	}

	// Step2: 返该 resident 涉及的所有 cards
	//   - 直接：cards.resident_id ∈ assigned (active_bed/room/unit overlay 后命中)
	//   - 间接：unit-level fallback 命中的 cards（resident 在 /96 但 LPM 跨 prefix 树时由 fallback 关联）
	//     —— 这部分通过 resident_unit ru ∈ assigned + cards 在 ru 所在 unit /80 范围内联出
	//   两者用 UNION 合并去重
	rows2, err := p.db.QueryContext(ctx,
		`(SELECT c.card_id::text AS card_id,
		         host(network(set_masklen(c.card_id, 56))) || '/56' AS branch_id
		    FROM cards c
		   WHERE c.card_id <<= $1::INET
		     AND c.resident_id = ANY($2::INET[]))
		 UNION
		 (SELECT c.card_id::text AS card_id,
		         host(network(set_masklen(c.card_id, 56))) || '/56' AS branch_id
		    FROM cards c
		    JOIN resident_unit ru ON ru.resident_id = ANY($2::INET[])
		                         AND ru.valid_to IS NULL
		   WHERE c.card_id <<= $1::INET
		     AND c.card_id <<= network(set_masklen(ru.spatial_prefix, 80)))`,
		tenantID, pq.Array(rids),
	)
	if err != nil {
		return nil, fmt.Errorf("filterByAssignedOnly cards: %w", err)
	}
	defer rows2.Close()

	byBranch := make(map[string][]string)
	for rows2.Next() {
		var cardID, branchID string
		if err := rows2.Scan(&cardID, &branchID); err != nil {
			return nil, fmt.Errorf("filterByAssignedOnly scan: %w", err)
		}
		if branchID == "" {
			branchID = "_"
		}
		byBranch[branchID] = append(byBranch[branchID], cardID)
	}
	return &CardList{UserID: userID, TenantID: tenantID, CardsByBranch: byBranch}, nil
}