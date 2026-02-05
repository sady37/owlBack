package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	commoncard "owl-common/card"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/store"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// AllowedCardIDsProviderImpl 实现 AllowedCardIDsProvider
// 借鉴 vital_focus_service 的 filterCardsForResident / filterCardsForStaff
type AllowedCardIDsProviderImpl struct {
	kv            store.KV
	usersRepo     repository.UsersRepository
	residentsRepo repository.ResidentsRepository
	db            *sql.DB
	logger        *zap.Logger
}

// NewAllowedCardIDsProvider 创建实现
func NewAllowedCardIDsProvider(
	kv store.KV,
	usersRepo repository.UsersRepository,
	residentsRepo repository.ResidentsRepository,
	db *sql.DB,
	logger *zap.Logger,
) *AllowedCardIDsProviderImpl {
	return &AllowedCardIDsProviderImpl{
		kv:            kv,
		usersRepo:     usersRepo,
		residentsRepo: residentsRepo,
		db:            db,
		logger:        logger,
	}
}

// GetAllowedCardIDsByBranches 返回按 branch 分组的卡片 ID，用于 Redis 缓存 {tenantID}:{branchID}:{userID}
// branchID 为空时用 "_" 表示
func (p *AllowedCardIDsProviderImpl) GetAllowedCardIDsByBranches(ctx context.Context, tenantID, userID, userType, userRole string) (map[string][]string, error) {
	filtered, err := p.getFilteredCards(ctx, tenantID, userID, userType, userRole)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string)
	for _, c := range filtered {
		if c.CardID == "" {
			continue
		}
		branchID := c.BranchID
		if branchID == "" {
			branchID = "_"
		}
		out[branchID] = append(out[branchID], c.CardID)
	}
	return out, nil
}

func (p *AllowedCardIDsProviderImpl) getFilteredCards(ctx context.Context, tenantID, userID, userType, userRole string) ([]commoncard.VitalFocusCardInfo, error) {
	keys, err := p.kv.ScanKeys(ctx, staticKeyPrefix+"*"+staticKeySuffix)
	if err != nil {
		return nil, err
	}
	cards := make([]commoncard.VitalFocusCardInfo, 0, len(keys))
	for _, key := range keys {
		raw, err := p.kv.Get(ctx, key)
		if err != nil {
			if errors.Is(err, store.ErrMiss) {
				continue
			}
			continue
		}
		var card commoncard.VitalFocusCardInfo
		if json.Unmarshal([]byte(raw), &card) != nil {
			continue
		}
		if tenantID != "" && tenantID != "*" && card.TenantID != tenantID {
			continue
		}
		cards = append(cards, card)
	}
	var filtered []commoncard.VitalFocusCardInfo
	if userType == "resident" {
		filtered, err = p.filterCardsForResident(ctx, userID, tenantID, cards)
	} else {
		filtered, err = p.filterCardsForStaff(ctx, userID, tenantID, cards)
	}
	if err != nil {
		return nil, err
	}
	return filtered, nil
}

// GetAllowedCardIDs 返回用户有权查看的卡片 ID 列表（扁平）
func (p *AllowedCardIDsProviderImpl) GetAllowedCardIDs(ctx context.Context, tenantID, userID, userType, userRole string) ([]string, error) {
	filtered, err := p.getFilteredCards(ctx, tenantID, userID, userType, userRole)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(filtered))
	for _, c := range filtered {
		if c.CardID != "" {
			ids = append(ids, c.CardID)
		}
	}
	return ids, nil
}

// filterCardsForResident Resident 本身在 card 内，按 card 内容直接过滤
func (p *AllowedCardIDsProviderImpl) filterCardsForResident(ctx context.Context, residentID, tenantID string, cards []commoncard.VitalFocusCardInfo) ([]commoncard.VitalFocusCardInfo, error) {
	resident, err := p.residentsRepo.GetResident(ctx, tenantID, residentID)
	if err != nil {
		return nil, err
	}
	filtered := make([]commoncard.VitalFocusCardInfo, 0)
	for _, card := range cards {
		if card.CardType == "ActiveBed" {
			if card.BedID != nil && *card.BedID == resident.BedID {
				if card.PrimaryResidentID != nil && *card.PrimaryResidentID == residentID {
					filtered = append(filtered, card)
					continue
				}
			}
		}
		if card.CardType == "Location" {
			if card.UnitID != "" && card.UnitID == resident.UnitID {
				for _, r := range card.Residents {
					if r.ResidentID == residentID {
						filtered = append(filtered, card)
						break
					}
				}
			}
		}
	}
	return filtered, nil
}

// filterCardsForStaff 使用 users.alarm_scope
// 入参 cards 已在 GetAllowedCardIDs 中按 tenantID 过滤，仅含本租户卡片
// ALL=本 tenant 内全部, ASSIGNED_ONLY=仅分配的住户/单元, BRANCH=仅所在院区
func (p *AllowedCardIDsProviderImpl) filterCardsForStaff(ctx context.Context, userID, tenantID string, cards []commoncard.VitalFocusCardInfo) ([]commoncard.VitalFocusCardInfo, error) {
	if p.db == nil {
		return cards, nil
	}
	user, err := p.usersRepo.GetUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if user.Role == "Admin" {
		return cards, nil
	}
	scope := ""
	if user.AlarmScope.Valid && user.AlarmScope.String != "" {
		scope = user.AlarmScope.String
	}
	switch scope {
	case "ALL":
		return cards, nil // cards 已限于本 tenant
	case "ASSIGNED_ONLY":
		return p.filterByAssignedOnly(ctx, tenantID, userID, cards)
	case "BRANCH":
		return p.filterByBranchOnly(ctx, tenantID, userID, cards)
	default:
		return nil, nil
	}
}

func (p *AllowedCardIDsProviderImpl) filterByAssignedOnly(ctx context.Context, tenantID, userID string, cards []commoncard.VitalFocusCardInfo) ([]commoncard.VitalFocusCardInfo, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT resident_id FROM resident_caregivers WHERE tenant_id = $1 AND caregiver_id = $2`,
		tenantID, userID,
	)
	if err != nil {
		return nil, err
	}
	var assignedResidentIDs []string
	for rows.Next() {
		var rid string
		if rows.Scan(&rid) == nil {
			assignedResidentIDs = append(assignedResidentIDs, rid)
		}
	}
	rows.Close()
	if len(assignedResidentIDs) == 0 {
		return nil, nil
	}
	rows, err = p.db.QueryContext(ctx,
		`SELECT DISTINCT r.unit_id FROM residents r WHERE r.tenant_id = $1 AND r.resident_id = ANY($2) AND r.unit_id IS NOT NULL`,
		tenantID, pq.Array(assignedResidentIDs),
	)
	if err != nil {
		return nil, err
	}
	ridMap := make(map[string]bool)
	for _, id := range assignedResidentIDs {
		ridMap[id] = true
	}
	locMap := make(map[string]bool)
	for rows.Next() {
		var loc string
		if rows.Scan(&loc) == nil && loc != "" {
			locMap[loc] = true
		}
	}
	rows.Close()
	filtered := make([]commoncard.VitalFocusCardInfo, 0)
	for _, card := range cards {
		if card.CardType == "ActiveBed" && card.PrimaryResidentID != nil && ridMap[*card.PrimaryResidentID] {
			filtered = append(filtered, card)
			continue
		}
		if card.CardType == "Location" && card.UnitID != "" && locMap[card.UnitID] {
			filtered = append(filtered, card)
		}
	}
	return filtered, nil
}

func (p *AllowedCardIDsProviderImpl) filterByBranchOnly(ctx context.Context, tenantID, userID string, cards []commoncard.VitalFocusCardInfo) ([]commoncard.VitalFocusCardInfo, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT branch_id::text FROM user_branches WHERE tenant_id = $1 AND user_id::text = $2`,
		tenantID, userID,
	)
	if err != nil {
		return nil, err
	}
	var branchIDs []string
	for rows.Next() {
		var bid sql.NullString
		if rows.Scan(&bid) == nil && bid.Valid && bid.String != "" {
			branchIDs = append(branchIDs, bid.String)
		}
	}
	rows.Close()
	if len(branchIDs) == 0 {
		return nil, nil
	}
	rows, err = p.db.QueryContext(ctx,
		`SELECT unit_id::text FROM units WHERE tenant_id = $1 AND branch_id = ANY($2::uuid[])`,
		tenantID, pq.Array(branchIDs),
	)
	if err != nil {
		return nil, err
	}
	unitMap := make(map[string]bool)
	for rows.Next() {
		var uid sql.NullString
		if rows.Scan(&uid) == nil && uid.Valid && uid.String != "" {
			unitMap[uid.String] = true
		}
	}
	rows.Close()
	filtered := make([]commoncard.VitalFocusCardInfo, 0)
	for _, card := range cards {
		if card.UnitID != "" && unitMap[card.UnitID] {
			filtered = append(filtered, card)
		}
	}
	return filtered, nil
}
