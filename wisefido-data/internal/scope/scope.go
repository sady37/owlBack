// Package scope — request-level security scope context.
//
// 业务背景：v2 schema 的 scope 维度有 3 层
//
//	(1) tenant   /48 INET — 多租户隔离
//	(2) branch   /56 INET — 每 user 在 user_branches 表挂多 branch (multi-branch 执业)，
//	    is_primary=TRUE 那行是 Current Branch (Sidebar 切换器选定)，所有 non tenant-wide
//	    list/query 都按此过滤
//	(3) role 分流
//	    - Admin / SystemAdmin / SystemOperator  → tenant-wide，不限 branch
//	    - Family / Resident                      → 按 resident_caregivers.family_id link
//	    - Manager / Nurse / Caregiver            → 按 Current Branch
//
// 此 package 把这 3 层抽象成一个 ScopeContext，由 middleware 在请求入口解析一次注入 ctx，
// 下游 handler / service 直接 `scope.FromContext(ctx)` 拿到，再用 BranchClause / VerifyXxx 走
// 标准化的 scope 检查路径。不再让每个 handler 自己拼 SQL 查 user_branches。
package scope

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// ScopeContext 是一个 request 一份的安全 scope，包含 user/tenant/branch/role 4 维信息。
//
// 字段都是 read-only 快照；切换 Current Branch 需要 user 显式调 /users/me/current-branch endpoint
// 并由 FE reload，新请求会拿到新快照。
type ScopeContext struct {
	UserID          string // UUID (空 = 未登录/内部请求)
	Role            string // normalized role: "Admin" / "SystemAdmin" / "Manager" / "Nurse" / "Caregiver" / "Family" / "Resident" / ...
	TenantPrefix    string // /48 INET CIDR (空 = 未限定 tenant，仅 SystemAdmin 跨 tenant 时)
	CurrentBranchID string // /56 INET CIDR (空 = tenant-wide user，无 branch 限制)
	HoA             string // 仅 Role=Resident 有意义 — 自己的 resident_id (INET /128)
}

// IsTenantWide — Admin / SystemAdmin / SystemOperator 跨 branch 不限制
func (sc *ScopeContext) IsTenantWide() bool {
	if sc == nil {
		return false
	}
	switch strings.ToLower(sc.Role) {
	case "admin", "systemadmin", "systemoperator":
		return true
	}
	return false
}

// IsFamily — Family 走 resident_caregivers.family_id link
func (sc *ScopeContext) IsFamily() bool {
	return sc != nil && strings.EqualFold(sc.Role, "Family")
}

// IsResident — Resident 仅看自己 hoa
func (sc *ScopeContext) IsResident() bool {
	return sc != nil && strings.EqualFold(sc.Role, "Resident")
}

// IsStaffBranchScoped — Manager / Nurse / Caregiver 等按 Current Branch 过滤
func (sc *ScopeContext) IsStaffBranchScoped() bool {
	if sc == nil || sc.IsTenantWide() || sc.IsFamily() || sc.IsResident() {
		return false
	}
	return true
}

// HasCurrentBranch — staff user 必须有挂载 branch 才能看数据；返 false 时 staff 应当看到空列表
func (sc *ScopeContext) HasCurrentBranch() bool {
	return sc != nil && strings.TrimSpace(sc.CurrentBranchID) != ""
}

// =============================================================================
// Resolver — 从 session/header 解析 ScopeContext，DB 查一次 user_branches.is_primary
// =============================================================================

// Resolve 从 session/header 已知的 user/tenant/role 出发，DB 查一次 Current Branch（is_primary），
// 拼出完整 ScopeContext。
//
// 调用方应在 handler 入口（或 auth middleware 之后）调用一次，结果用 WithContext 写入 ctx。
// 后续 handler/service 用 FromContext 拿。
func Resolve(ctx context.Context, db *sql.DB, userID, role, tenantPrefix, hoa string) (*ScopeContext, error) {
	sc := &ScopeContext{
		UserID:       strings.TrimSpace(userID),
		Role:         strings.TrimSpace(role),
		TenantPrefix: strings.TrimSpace(tenantPrefix),
		HoA:          strings.TrimSpace(hoa),
	}
	if sc.IsTenantWide() || sc.IsFamily() || sc.IsResident() {
		// Admin/Family/Resident 不依赖 user_branches —— 直接返
		return sc, nil
	}
	if sc.UserID == "" || db == nil {
		return sc, nil
	}
	var bid sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT host(branch_prefix) || '/56'
		  FROM user_branches
		 WHERE user_id = $1::UUID
		   AND is_primary = TRUE
		   AND valid_to IS NULL
		 LIMIT 1`, sc.UserID).Scan(&bid)
	if err == sql.ErrNoRows {
		return sc, nil
	}
	if err != nil {
		return sc, fmt.Errorf("resolve current branch: %w", err)
	}
	if bid.Valid {
		sc.CurrentBranchID = bid.String
	}
	return sc, nil
}

// =============================================================================
// Context plumbing
// =============================================================================

type ctxKey struct{}

// WithContext 把 ScopeContext 注入 ctx；middleware 用
func WithContext(ctx context.Context, sc *ScopeContext) context.Context {
	if sc == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, sc)
}

// FromContext 取出注入的 ScopeContext。若未注入返 nil 和 false。
func FromContext(ctx context.Context) (*ScopeContext, bool) {
	if ctx == nil {
		return nil, false
	}
	sc, ok := ctx.Value(ctxKey{}).(*ScopeContext)
	return sc, ok
}

// MustFromContext 取出 ScopeContext，未注入则返空指针（避免 panic；调用方应 nil-check）
func MustFromContext(ctx context.Context) *ScopeContext {
	sc, _ := FromContext(ctx)
	return sc
}

// =============================================================================
// SQL Filter clauses — 标准化各 list 端点的 branch / resident scope WHERE 片段
// =============================================================================

// FilterClause — SQL 片段 + 对应位置的参数；空 clause 表示 "不加过滤"
type FilterClause struct {
	SQL  string // 不含前导 AND/WHERE；为空表示放行
	Args []any
}

// IsEmpty — 表示是否要拼到主 SQL（admin tenant-wide 时这是 true）
func (fc FilterClause) IsEmpty() bool { return strings.TrimSpace(fc.SQL) == "" }

// BranchClause — staff scope: column <<= current_branch_prefix
//
// 用法：
//
//	fc := sc.BranchClause("d.device_ipv6", argIdx)
//	if !fc.IsEmpty() {
//	    where = append(where, fc.SQL)
//	    args = append(args, fc.Args...)
//	}
//
// admin 返空 clause；family/resident 不该走这个（应走 ResidentClause）
func (sc *ScopeContext) BranchClause(column string, argIdx int) FilterClause {
	if sc == nil || sc.IsTenantWide() {
		return FilterClause{}
	}
	if !sc.HasCurrentBranch() {
		// staff 没挂 branch → 应看到空列表（强制返一个永假 clause）
		return FilterClause{SQL: "FALSE"}
	}
	return FilterClause{
		SQL:  fmt.Sprintf("%s <<= $%d::INET", column, argIdx),
		Args: []any{sc.CurrentBranchID},
	}
}

// ResidentClause — Family scope: column IN (SELECT resident_id FROM resident_caregivers WHERE family_id = self)
// Resident scope: column = self.hoa
// Admin / Staff 不该用此 clause
func (sc *ScopeContext) ResidentClause(column string, argIdx int) FilterClause {
	if sc == nil || sc.IsTenantWide() {
		return FilterClause{}
	}
	if sc.IsResident() {
		if sc.HoA == "" {
			return FilterClause{SQL: "FALSE"}
		}
		return FilterClause{
			SQL:  fmt.Sprintf("%s = $%d::INET", column, argIdx),
			Args: []any{sc.HoA},
		}
	}
	if sc.IsFamily() {
		if sc.UserID == "" {
			return FilterClause{SQL: "FALSE"}
		}
		return FilterClause{
			SQL: fmt.Sprintf(
				"%s IN (SELECT resident_id FROM resident_caregivers WHERE family_id = $%d::UUID AND valid_to IS NULL)",
				column, argIdx,
			),
			Args: []any{sc.UserID},
		}
	}
	return FilterClause{} // staff 用 BranchClause 不用这个
}

// =============================================================================
// Single-resource Verify — 用于 GET /{resource}/{id} 类查询的点对点权限校验
// =============================================================================

// VerifyCard — 校验 user 能查看该 cardID。返 nil = 允许；error = 越权
func (sc *ScopeContext) VerifyCard(ctx context.Context, db *sql.DB, cardID string) error {
	if sc == nil || db == nil || strings.TrimSpace(cardID) == "" {
		return nil
	}
	if sc.IsTenantWide() {
		return nil
	}
	if sc.IsFamily() || sc.IsResident() {
		var ok bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
			  SELECT 1 FROM cards c
			  JOIN resident_caregivers rc
			    ON rc.resident_id = c.resident_id
			   AND rc.family_id = $1::UUID
			   AND rc.valid_to IS NULL
			 WHERE c.card_id = $2::INET
			)`, sc.UserID, cardID).Scan(&ok)
		if err != nil {
			return fmt.Errorf("verify card family link: %w", err)
		}
		if !ok {
			return fmt.Errorf("permission denied: card not linked to your residents")
		}
		return nil
	}
	// Staff: Current Branch
	if !sc.HasCurrentBranch() {
		return fmt.Errorf("permission denied: no current branch")
	}
	var ok bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM cards c
		 WHERE c.card_id = $1::INET
		   AND c.card_id <<= $2::INET
		)`, cardID, sc.CurrentBranchID).Scan(&ok)
	if err != nil {
		return fmt.Errorf("verify card branch: %w", err)
	}
	if !ok {
		return fmt.Errorf("permission denied: card not in your current branch")
	}
	return nil
}

// VerifyDevice — 校验 user 能查看该 deviceAddr (INET /128 canonical text)
// 业务键改为 device_addr，原 device_id UUID 退役。
func (sc *ScopeContext) VerifyDevice(ctx context.Context, db *sql.DB, deviceAddr string) error {
	if sc == nil || db == nil || strings.TrimSpace(deviceAddr) == "" {
		return nil
	}
	if sc.IsTenantWide() {
		return nil
	}
	if sc.IsFamily() || sc.IsResident() {
		var ok bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
			  SELECT 1 FROM devices d
			  JOIN resident_caregivers rc ON rc.family_id = $1::UUID AND rc.valid_to IS NULL
			  JOIN resident_unit ru
			    ON ru.resident_id = rc.resident_id
			   AND ru.valid_to IS NULL
			 WHERE d.device_addr = $2::INET
			   AND d.device_addr <<= ru.spatial_prefix
			)`, sc.UserID, deviceAddr).Scan(&ok)
		if err != nil {
			return fmt.Errorf("verify device family link: %w", err)
		}
		if !ok {
			return fmt.Errorf("permission denied: device not in your residents' space")
		}
		return nil
	}
	if !sc.HasCurrentBranch() {
		return fmt.Errorf("permission denied: no current branch")
	}
	var ok bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM devices d
		 WHERE d.device_addr = $1::INET
		   AND d.device_addr <<= $2::INET
		)`, deviceAddr, sc.CurrentBranchID).Scan(&ok)
	if err != nil {
		return fmt.Errorf("verify device branch: %w", err)
	}
	if !ok {
		return fmt.Errorf("permission denied: device not in your current branch")
	}
	return nil
}

// VerifyResident — 校验 user 能查看该 resident (hoa INET /128)
func (sc *ScopeContext) VerifyResident(ctx context.Context, db *sql.DB, hoa string) error {
	if sc == nil || db == nil || strings.TrimSpace(hoa) == "" {
		return nil
	}
	if sc.IsTenantWide() {
		return nil
	}
	if sc.IsResident() {
		if sc.HoA == hoa {
			return nil
		}
		return fmt.Errorf("permission denied: can only view own profile")
	}
	if sc.IsFamily() {
		var ok bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
			  SELECT 1 FROM resident_caregivers
			   WHERE resident_id = $1::INET
			     AND family_id   = $2::UUID
			     AND valid_to IS NULL
			)`, hoa, sc.UserID).Scan(&ok)
		if err != nil {
			return fmt.Errorf("verify resident family link: %w", err)
		}
		if !ok {
			return fmt.Errorf("permission denied: not linked as family to this resident")
		}
		return nil
	}
	// Staff: resident 当前住的 branch /56 必须 == user.CurrentBranchID
	if !sc.HasCurrentBranch() {
		return fmt.Errorf("permission denied: no current branch")
	}
	var ok bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM resident_unit ru
		   WHERE ru.resident_id = $1::INET
		     AND ru.valid_to IS NULL
		     AND ru.spatial_prefix <<= $2::INET
		)`, hoa, sc.CurrentBranchID).Scan(&ok)
	if err != nil {
		return fmt.Errorf("verify resident branch: %w", err)
	}
	if !ok {
		return fmt.Errorf("permission denied: resident not in your current branch")
	}
	return nil
}

// VerifyUnit — 校验 user 能查看 unit (INET /80)；按 prefix 反推 /56 比对 Current Branch
func (sc *ScopeContext) VerifyUnit(ctx context.Context, db *sql.DB, unitID string) error {
	if sc == nil || strings.TrimSpace(unitID) == "" {
		return nil
	}
	if sc.IsTenantWide() {
		return nil
	}
	// Family/Resident 通过 unit 间接看 resident — 不直接 verify unit；调用方需走 resident 链
	if sc.IsFamily() || sc.IsResident() {
		return fmt.Errorf("permission denied: family/resident cannot access unit directly")
	}
	if !sc.HasCurrentBranch() {
		return fmt.Errorf("permission denied: no current branch")
	}
	// 纯 INET 比较即可 — 不需要 DB（unit_id /80 <<= branch /56 是纯 IP 数学）
	// 但保险起见还是 DB 查 units 表确认 unit 真实存在
	if db == nil {
		return nil
	}
	var ok bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM units WHERE unit_id = $1::INET AND $1::INET <<= $2::INET
		)`, unitID, sc.CurrentBranchID).Scan(&ok)
	if err != nil {
		return fmt.Errorf("verify unit branch: %w", err)
	}
	if !ok {
		return fmt.Errorf("permission denied: unit not in your current branch")
	}
	return nil
}
