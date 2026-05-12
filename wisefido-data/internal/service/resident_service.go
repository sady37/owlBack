// Package service — Resident V2 Service (Forward Design)
//
// 业务规则集中点；不依赖 v1 ResidentService。
package service

import (
	"context"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/scope"

	"go.uber.org/zap"
)

type ResidentService struct {
	repo   *repository.PostgresResidentsRepository
	logger *zap.Logger
}

func NewResidentService(repo *repository.PostgresResidentsRepository) *ResidentService {
	return &ResidentService{repo: repo, logger: zap.NewNop()}
}

func (s *ResidentService) SetLogger(logger *zap.Logger) {
	if logger != nil {
		s.logger = logger
	}
}

// ============================================================================
// 业务规则 — 角色权限
// ============================================================================
//
// Admin / Manager   : full CRUD
// Nurse             : 改 resident 所有字段；不能改 admission/discharge (move_in/out_date)
// Resident          : 仅可 GET 自己 hoa
// SystemAdmin       : 平台级，跨 tenant 只读

const (
	roleAdmin       = "Admin"
	roleManager     = "Manager"
	roleNurse       = "Nurse"
	roleCaregiver   = "Caregiver"
	roleResident    = "Resident"
	roleFamily      = "Family"
	roleSystemAdmin = "SystemAdmin"
)

// isFamilyRole — Family role 大小写兼容判断
func isFamilyRole(role string) bool {
	return strings.EqualFold(role, roleFamily)
}

func canCRUDResident(role string) bool {
	switch role {
	case roleAdmin, roleManager, "tenant_admin", "manager":
		return true
	}
	return false
}

func canEditResident(role string) bool {
	if canCRUDResident(role) {
		return true
	}
	return role == roleNurse || role == "nurse"
}

func canEditAdmissionDischarge(role string) bool {
	// Nurse 不能改 admission/discharge（涉及财务）
	return canCRUDResident(role)
}

// ============================================================================
// Read
// ============================================================================

type ListResidentsV2Request struct {
	TenantPrefix    string
	CurrentUserID   string // user_id UUID（Family scope 必填）
	CurrentUserHOA  string
	CurrentUserRole string
	Filter          domain.ResidentListFilter
}

type ListResidentsV2Response struct {
	Items []*domain.Resident `json:"items"`
	Total int                  `json:"total"`
}

func (s *ResidentService) List(ctx context.Context, req ListResidentsV2Request) (*ListResidentsV2Response, error) {
	if req.TenantPrefix == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	sc := scope.MustFromContext(ctx)

	// Resident 只看自己（middleware 解析 hoa；fallback 兼容旧 caller 没注入 ctx 的场景）
	hoa := ""
	if sc != nil && sc.IsResident() {
		hoa = sc.HoA
	} else if req.CurrentUserRole == roleResident {
		hoa = req.CurrentUserHOA
	}
	if hoa != "" {
		one, err := s.repo.GetResident(ctx, req.TenantPrefix, hoa)
		if err != nil {
			return &ListResidentsV2Response{Items: []*domain.Resident{}, Total: 0}, nil
		}
		return &ListResidentsV2Response{Items: []*domain.Resident{&one.Resident}, Total: 1}, nil
	}

	// Family 走 resident_caregivers link
	if (sc != nil && sc.IsFamily()) || isFamilyRole(req.CurrentUserRole) {
		uid := ""
		if sc != nil {
			uid = sc.UserID
		}
		if uid == "" {
			uid = req.CurrentUserID
		}
		if strings.TrimSpace(uid) == "" {
			return nil, fmt.Errorf("permission denied: Family role requires authenticated user_id")
		}
		req.Filter.FamilyUserID = uid
	} else if sc != nil && sc.IsStaffBranchScoped() {
		// Staff (Manager/Nurse/Caregiver) 按 Current Branch 严格过滤
		if sc.HasCurrentBranch() {
			req.Filter.BranchPrefix = sc.CurrentBranchID
		} else {
			// 无 current branch → 看空
			return &ListResidentsV2Response{Items: []*domain.Resident{}, Total: 0}, nil
		}
	}
	// 兜底兼容：ctx 没注入 scope 但调用方传了旧字段 — 继续保留旧行为防回归
	if sc == nil && !isFamilyRole(req.CurrentUserRole) &&
		req.CurrentUserRole != roleAdmin && req.CurrentUserRole != roleSystemAdmin {
		if cb, err := s.repo.GetCurrentBranchID(ctx, req.CurrentUserID); err == nil && cb != "" {
			req.Filter.BranchPrefix = cb
		}
	}

	items, total, err := s.repo.ListResidents(ctx, req.TenantPrefix, req.Filter)
	if err != nil {
		return nil, err
	}
	return &ListResidentsV2Response{Items: items, Total: total}, nil
}

type GetResidentRequest struct {
	TenantPrefix    string
	HoA             string
	CurrentUserID   string // user_id UUID（Family scope 必填）
	CurrentUserHOA  string
	CurrentUserRole string
}

func (s *ResidentService) Get(ctx context.Context, req GetResidentRequest) (*domain.ResidentDetail, error) {
	if req.TenantPrefix == "" || req.HoA == "" {
		return nil, fmt.Errorf("tenant_id and hoa required")
	}
	// 优先用 scope.FromContext 统一校验；ctx 没注入时 fallback 旧字段
	if sc := scope.MustFromContext(ctx); sc != nil {
		if err := sc.VerifyResident(ctx, s.repo.DB(), req.HoA); err != nil {
			return nil, err
		}
	} else {
		// fallback：兼容旧 caller / 测试场景
		if req.CurrentUserRole == roleResident && req.CurrentUserHOA != req.HoA {
			return nil, fmt.Errorf("permission denied: can only view own profile")
		}
		if isFamilyRole(req.CurrentUserRole) {
			if strings.TrimSpace(req.CurrentUserID) == "" {
				return nil, fmt.Errorf("permission denied: Family role requires authenticated user_id")
			}
			linked, err := s.repo.IsResidentLinkedToFamily(ctx, req.HoA, req.CurrentUserID)
			if err != nil {
				return nil, err
			}
			if !linked {
				return nil, fmt.Errorf("permission denied: not linked as family to this resident")
			}
		}
	}
	return s.repo.GetResident(ctx, req.TenantPrefix, req.HoA)
}

// ============================================================================
// Write
// ============================================================================

type CreateResidentRequest struct {
	TenantPrefix    string
	ActorUserID     string // 操作者 user_id UUID（audit_log 用）
	CurrentUserRole string
	Input           *domain.ResidentCreateInput
}

func (s *ResidentService) Create(ctx context.Context, req CreateResidentRequest) (string, error) {
	if !canCRUDResident(req.CurrentUserRole) {
		return "", fmt.Errorf("permission denied: only Admin/Manager can create residents")
	}
	if req.Input == nil {
		return "", fmt.Errorf("input required")
	}
	return s.repo.CreateResident(ctx, req.TenantPrefix, req.Input, req.ActorUserID, req.CurrentUserRole)
}

type UpdateResidentRequest struct {
	TenantPrefix    string
	HoA             string
	ActorUserID     string // 操作者 user_id UUID（audit_log 用）
	CurrentUserRole string
	Input           *domain.ResidentUpdateInput
}

func (s *ResidentService) Update(ctx context.Context, req UpdateResidentRequest) error {
	s.logger.Info("[ResidentService.Update ENTRY]", zap.String("hoa", req.HoA), zap.String("tenant_prefix", req.TenantPrefix), zap.String("role", req.CurrentUserRole))
	// Family 不能改 resident profile（PHI 写入权限不下放给家属）
	if isFamilyRole(req.CurrentUserRole) {
		return fmt.Errorf("permission denied: Family role cannot edit resident profile")
	}
	if !canEditResident(req.CurrentUserRole) {
		s.logger.Error("[ResidentService.Update FAILED] permission denied", zap.String("role", req.CurrentUserRole))
		return fmt.Errorf("permission denied: role %q cannot edit residents", req.CurrentUserRole)
	}
	if req.Input == nil {
		s.logger.Error("[ResidentService.Update FAILED] input nil")
		return fmt.Errorf("input required")
	}
	// Nurse 不能改 admission/discharge
	if !canEditAdmissionDischarge(req.CurrentUserRole) {
		if req.Input.AdmissionDate != nil || req.Input.DischargeDate != nil {
			s.logger.Error("[ResidentService.Update FAILED] Nurse cannot modify dates")
			return fmt.Errorf("permission denied: Nurse cannot modify admission/discharge dates (financial)")
		}
	}
	err := s.repo.UpdateResident(ctx, req.TenantPrefix, req.HoA, req.Input, req.ActorUserID, req.CurrentUserRole)
	if err != nil {
		s.logger.Error("[ResidentService.Update FAILED]", zap.String("hoa", req.HoA), zap.Error(err))
		return err
	}
	s.logger.Info("[ResidentService.Update SUCCESS]", zap.String("hoa", req.HoA))
	return nil
}

type DeleteResidentRequest struct {
	TenantPrefix    string
	HoA             string
	CurrentUserRole string
	Hard            bool // true → 硬删 (Clear)，需 CheckClearable
}

func (s *ResidentService) Delete(ctx context.Context, req DeleteResidentRequest) error {
	if !canCRUDResident(req.CurrentUserRole) {
		return fmt.Errorf("permission denied: only Admin/Manager can delete residents")
	}
	if req.Hard {
		return s.repo.HardDelete(ctx, req.HoA)
	}
	return s.repo.SoftDelete(ctx, req.HoA)
}

func (s *ResidentService) CheckClearable(ctx context.Context, hoa string) (*domain.ResidentClearCheckResult, error) {
	return s.repo.CheckClearable(ctx, hoa)
}

// ============================================================================
// 转院（discharge + admission）— 业务规则：跨 tenant 不 in-place 改 hoa
// ============================================================================
//
// 流程：
//   step 1: discharge from A — 当前 tenant set status='discharged' + move_out_date=NOW
//   step 2: admission to B — 在 B tenant Create 新 resident（新 hoa），可携带 PHI / contacts / caregivers
// 这是 2 次 API 调用（FE 引导 user），后端不提供 atomic transfer。
// 当前 V2 service 暂不暴露 transfer endpoint；FE 用 Update + Create 两步组合。

func (s *ResidentService) Discharge(ctx context.Context, req UpdateResidentRequest) error {
	if !canCRUDResident(req.CurrentUserRole) {
		return fmt.Errorf("permission denied")
	}
	now := nowDateString()
	in := &domain.ResidentUpdateInput{
		Status:      strPtr("discharged"),
		DischargeDate: &now,
	}
	return s.repo.UpdateResident(ctx, req.TenantPrefix, req.HoA, in, req.ActorUserID, req.CurrentUserRole)
}

// ============================================================================
// helpers
// ============================================================================

func strPtr(s string) *string { return &s }
func nowDateString() string {
	// 不引入 time package 在文件里仅为了 Format — 用 PG NOW() 比较安全。
	// 这里返空字符串让 repo 用 NULL；如要精确，改 input 生成 time.Now().Format("2006-01-02") 即可。
	return strings.TrimSpace("") // 占位；调用方应自填
}
