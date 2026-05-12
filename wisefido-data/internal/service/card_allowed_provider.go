package service

import (
	"context"
	"database/sql"
	"fmt"

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

// AllCardIDs 展平所有 branch 下的 card_id
func (cl *CardList) AllCardIDs() []string {
	var ids []string
	for _, cids := range cl.CardsByBranch {
		ids = append(ids, cids...)
	}
	return ids
}

// BranchIDs 返回去重的 branch_id 列表（不含 "_"）
func (cl *CardList) BranchIDs() []string {
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

	// unit_type=home → 该 unit 下全部 card
	if unitInfo.UnitType == "home" {
		return p.cardIDsByUnit(ctx, tenantID, *resident.UnitID, residentID)
	}

	// unit_type=facility + is_public → nil
	if unitInfo.IsPublic {
		return nil, nil
	}

	// facility + not public + not shared → 该 unit 下全部 card
	if !unitInfo.IsSharedUnit {
		return p.cardIDsByUnit(ctx, tenantID, *resident.UnitID, residentID)
	}

	// facility + not public + shared → ActiveBedCard only
	return p.ActiveBedcardIDsByUnitShared(ctx, tenantID, *resident.UnitID, residentID)
}

// cardIDsByUnit 查该 unit 下所有 card，返回 *CardList
func (p *AllowedCardIDsProviderImpl) cardIDsByUnit(ctx context.Context, tenantID, unitID, userID string) (*CardList, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT card_id::text, COALESCE(branch_id::text, '_') FROM cards WHERE tenant_id = $1 AND unit_id = $2`,
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

// ActiveBedcardIDsByUnitShared shared unit：只返回 ActiveBedCard 且 residents JSONB 含 residentID，返回 *CardList
func (p *AllowedCardIDsProviderImpl) ActiveBedcardIDsByUnitShared(ctx context.Context, tenantID, unitID, residentID string) (*CardList, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT card_id::text, COALESCE(branch_id::text, '_') FROM cards
		 WHERE tenant_id = $1 AND unit_id = $2
		   AND card_type = 'ActiveBedCard'
		   AND EXISTS (
		     SELECT 1 FROM jsonb_array_elements(residents) elem
		     WHERE elem->>'resident_id' = $3
		   )`,
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

	scope := ""
	if user.Relegation.Valid && user.Relegation.String != "" {
		scope = user.Relegation.String
	}
	if user.Role == "Admin" || scope == "ALL" {
		return p.filterTenantCards(ctx, tenantID, userID)
	}
	switch scope {
	case "ASSIGNED_ONLY":
		return p.filterByAssignedOnly(ctx, tenantID, userID)
	case "BRANCH":
		return p.filterByBranchOnly(ctx, tenantID, userID)
	default:
		return nil, nil
	}
}

// filterTenantCards 查该 tenant 全部卡片，按 branch 分组返回 *CardList
func (p *AllowedCardIDsProviderImpl) filterTenantCards(ctx context.Context, tenantID, userID string) (*CardList, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT card_id::text, COALESCE(branch_id::text, '_') FROM cards WHERE tenant_id = $1 AND card_type <> 'DeviceCard'`,
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
		`SELECT c.card_id::text,
		        host(network(set_masklen(c.spatial_prefix, 56))) || '/56' AS branch_id
		   FROM cards c
		  WHERE c.spatial_prefix <<= $1::INET
		    AND c.card_type <> 'DeviceCard'
		    AND EXISTS (
		      SELECT 1 FROM user_branches ub
		       WHERE ub.user_id = $2::UUID
		         AND ub.is_primary = TRUE
		         AND ub.valid_to IS NULL
		         AND c.spatial_prefix <<= ub.branch_prefix
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
// 1. 查 resident_caregivers 得到该 staff 负责的 resident_id[]
// 2. 查 cards 中 resident_id = ANY(assigned) 的 ActiveBedCard
//    + cards 中 residents JSONB 包含 assigned resident 的 UnitCard
// 直接返回 *CardList
func (p *AllowedCardIDsProviderImpl) filterByAssignedOnly(ctx context.Context, tenantID, userID string) (*CardList, error) {
	// Step1: 该 staff 负责的 resident_id[]（user_list JSONB 数组包含该 userID）
	rows, err := p.db.QueryContext(ctx,
		`SELECT resident_id::text FROM resident_caregivers
		 WHERE tenant_id = $1
		   AND user_list IS NOT NULL
		   AND EXISTS (
		     SELECT 1 FROM jsonb_array_elements_text(user_list) elem WHERE elem = $2
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

	// Step2: ActiveBedCard（resident_id 直接匹配）+ UnitCard（residents JSONB 包含）
	rows2, err := p.db.QueryContext(ctx,
		`SELECT card_id::text, COALESCE(branch_id::text, '_')
		 FROM cards
		 WHERE tenant_id = $1
		   AND (
		     (card_type = 'ActiveBedCard' AND resident_id = ANY($2::uuid[]))
		     OR
		     (card_type = 'UnitCard' AND EXISTS (
		       SELECT 1 FROM jsonb_array_elements(residents) r
		       WHERE r->>'resident_id' = ANY($3::text[])
		     ))
		   )`,
		tenantID, pq.Array(rids), pq.Array(rids),
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