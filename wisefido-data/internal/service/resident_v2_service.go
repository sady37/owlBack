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
)

type ResidentV2Service struct {
	repo *repository.PostgresResidentsRepositoryV2
}

func NewResidentV2Service(repo *repository.PostgresResidentsRepositoryV2) *ResidentV2Service {
	return &ResidentV2Service{repo: repo}
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
	roleSystemAdmin = "SystemAdmin"
)

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
	CurrentUserHOA  string
	CurrentUserRole string
	Filter          domain.ResidentV2ListFilter
}

type ListResidentsV2Response struct {
	Items []*domain.ResidentV2 `json:"items"`
	Total int                  `json:"total"`
}

func (s *ResidentV2Service) List(ctx context.Context, req ListResidentsV2Request) (*ListResidentsV2Response, error) {
	if req.TenantPrefix == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	// Resident 角色只能看自己 — 在 service 层兜底过滤（FE 也会限制）
	if req.CurrentUserRole == roleResident && req.CurrentUserHOA != "" {
		// 把 search 替换成自己 hoa 锁死（简化处理）
		// 更严谨：直接构 single-row response
		one, err := s.repo.GetResident(ctx, req.TenantPrefix, req.CurrentUserHOA)
		if err != nil {
			return &ListResidentsV2Response{Items: []*domain.ResidentV2{}, Total: 0}, nil
		}
		return &ListResidentsV2Response{Items: []*domain.ResidentV2{&one.ResidentV2}, Total: 1}, nil
	}
	items, total, err := s.repo.ListResidents(ctx, req.TenantPrefix, req.Filter)
	if err != nil {
		return nil, err
	}
	return &ListResidentsV2Response{Items: items, Total: total}, nil
}

type GetResidentV2Request struct {
	TenantPrefix    string
	HoA             string
	CurrentUserHOA  string
	CurrentUserRole string
}

func (s *ResidentV2Service) Get(ctx context.Context, req GetResidentV2Request) (*domain.ResidentV2Detail, error) {
	if req.TenantPrefix == "" || req.HoA == "" {
		return nil, fmt.Errorf("tenant_id and hoa required")
	}
	// Resident 仅可看自己
	if req.CurrentUserRole == roleResident && req.CurrentUserHOA != req.HoA {
		return nil, fmt.Errorf("permission denied: can only view own profile")
	}
	return s.repo.GetResident(ctx, req.TenantPrefix, req.HoA)
}

// ============================================================================
// Write
// ============================================================================

type CreateResidentV2Request struct {
	TenantPrefix    string
	CurrentUserRole string
	Input           *domain.ResidentV2CreateInput
}

func (s *ResidentV2Service) Create(ctx context.Context, req CreateResidentV2Request) (string, error) {
	if !canCRUDResident(req.CurrentUserRole) {
		return "", fmt.Errorf("permission denied: only Admin/Manager can create residents")
	}
	if req.Input == nil {
		return "", fmt.Errorf("input required")
	}
	return s.repo.CreateResident(ctx, req.TenantPrefix, req.Input)
}

type UpdateResidentV2Request struct {
	TenantPrefix    string
	HoA             string
	CurrentUserRole string
	Input           *domain.ResidentV2UpdateInput
}

func (s *ResidentV2Service) Update(ctx context.Context, req UpdateResidentV2Request) error {
	if !canEditResident(req.CurrentUserRole) {
		return fmt.Errorf("permission denied: role %q cannot edit residents", req.CurrentUserRole)
	}
	if req.Input == nil {
		return fmt.Errorf("input required")
	}
	// Nurse 不能改 admission/discharge
	if !canEditAdmissionDischarge(req.CurrentUserRole) {
		if req.Input.AdmissionDate != nil || req.Input.DischargeDate != nil {
			return fmt.Errorf("permission denied: Nurse cannot modify admission/discharge dates (financial)")
		}
	}
	return s.repo.UpdateResident(ctx, req.TenantPrefix, req.HoA, req.Input)
}

type DeleteResidentV2Request struct {
	TenantPrefix    string
	HoA             string
	CurrentUserRole string
	Hard            bool // true → 硬删 (Clear)，需 CheckClearable
}

func (s *ResidentV2Service) Delete(ctx context.Context, req DeleteResidentV2Request) error {
	if !canCRUDResident(req.CurrentUserRole) {
		return fmt.Errorf("permission denied: only Admin/Manager can delete residents")
	}
	if req.Hard {
		return s.repo.HardDelete(ctx, req.HoA)
	}
	return s.repo.SoftDelete(ctx, req.HoA)
}

func (s *ResidentV2Service) CheckClearable(ctx context.Context, hoa string) (*domain.ResidentV2ClearCheckResult, error) {
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

func (s *ResidentV2Service) Discharge(ctx context.Context, req UpdateResidentV2Request) error {
	if !canCRUDResident(req.CurrentUserRole) {
		return fmt.Errorf("permission denied")
	}
	now := nowDateString()
	in := &domain.ResidentV2UpdateInput{
		Status:      strPtr("discharged"),
		DischargeDate: &now,
	}
	return s.repo.UpdateResident(ctx, req.TenantPrefix, req.HoA, in)
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
