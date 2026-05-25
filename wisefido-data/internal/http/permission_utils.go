package httpapi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/scope"
)

// PermissionCheck 权限检查结果
// 包含 assigned_only 和 branch_only 两个标志（v1 历史语义保留）
//
// v2 (owl_v2) 改造说明：
//   v1 表 role_permissions(tenant_id, role_code, resource_type, permission_type, permission_scope)
//   v2 表 role_permissions(role_id FK→roles, permission, resource_scope INET)
//   v2 角色：role_code 为 snake_case（platform_admin / tenant_admin / manager / nurse / caregiver / family / viewer）
//   v2 通配：platform_admin = '*' (全平台); tenant_admin = 'tenant.*' (单租户全权)
//
// 现在的实现策略：把 v1 (resource, permType) 翻译成 v2 permission 字符串，
// 用 EXISTS 查 role 是否拥有 (resource.action / resource.* / tenant.* / *) 任一匹配。
// scope (assigned_only / branch_only) 在 v2 由 IPv6 prefix 自带层级表达，业务侧用 utils/spatial 派生；
// 这里返回 (false,false) = 允许，让上层不做额外 branch 过滤；不允许时仍返回 strictest 兜底。
type PermissionCheck struct {
	AssignedOnly bool // 是否仅限分配的资源
	BranchOnly   bool // 是否仅限同一 Branch 的资源
}

// mapRoleToV2 把 v1 PascalCase role code（前端/session 注入用）映射回 v2 snake_case role_code。
// 与 normalizeRole 在 auth_v2_handler 里的方向相反。
func mapRoleToV2(role string) string {
	switch role {
	case "SystemAdmin":
		return "platform_admin"
	case "SystemOperator":
		// owl_v2 seed 暂无独立 operator 角色；按最高权限处理（与 v1 行为一致）
		return "platform_admin"
	case "Admin":
		return "tenant_admin"
	case "Manager":
		return "manager"
	case "Nurse":
		return "nurse"
	case "Caregiver":
		return "caregiver"
	case "Family":
		return "family"
	case "Viewer":
		return "viewer"
	}
	return strings.ToLower(role)
}

// permWord 把 v1 单字母权限码转 v2 动词。
func permWord(p string) string {
	switch strings.ToUpper(p) {
	case "R":
		return "read"
	case "C":
		return "create"
	case "U":
		return "update"
	case "D":
		return "delete"
	}
	return strings.ToLower(p)
}

// GetResourcePermission 查询 v2 RBAC 中 (role, resource, action) 是否被允许。
// 兼容 v1 调用方签名；内部走 v2 schema。
func GetResourcePermission(db *sql.DB, ctx context.Context,
	roleCode, resourceType, permissionType string) (*PermissionCheck, error) {

	v2Role := mapRoleToV2(roleCode)
	action := permWord(permissionType)

	// 候选 permission 字符串：精确 → 资源 .config（CRUD 合一）→ 资源通配 → 租户通配 → 全局通配
	// .config 语义：v2 seed 用 `<resource>.config` 表示"完全管理该 config"（含 R/C/U/D），
	// 比拆 4 行单独 grant 更紧凑（详见 owlrd/dbv2/42_role_permissions.sql 注释）。
	target := resourceType + "." + action
	resourceConfig := resourceType + ".config"
	resourceAll := resourceType + ".*"
	const tenantAll = "tenant.*"
	const platformAll = "*"

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			  FROM role_permissions rp
			  JOIN roles r ON r.role_id = rp.role_id
			 WHERE r.role_code = $1
			   AND rp.permission IN ($2, $3, $4, $5, $6)
		)
	`, v2Role, target, resourceConfig, resourceAll, tenantAll, platformAll).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		// 没匹配：返回 strictest（v1 行为兜底）
		return &PermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	// 命中通配/精确权限 → 完全允许（branch 边界后续业务侧用 utils/spatial 自己判）
	return &PermissionCheck{AssignedOnly: false, BranchOnly: false}, nil
}

// VerifyCardInScope / VerifyDeviceInScope — thin shim 调 internal/scope
//
// Step B 后真正实现在 internal/scope.ScopeContext.VerifyCard/VerifyDevice 里。
// 这里保留 wrapper 让旧 caller (radar / playback handlers) 不用改签名也能工作。
// 新代码建议直接：
//
//	sc := scope.MustFromContext(r.Context())
//	if err := sc.VerifyCard(ctx, db, cardID); err != nil { ... }

func VerifyCardInScope(db *sql.DB, ctx context.Context, userID, role, cardID string) error {
	// 优先用 ctx 里的 ScopeContext；fallback 现场构造一个最小版
	sc, ok := scope.FromContext(ctx)
	if !ok || sc == nil {
		sc = &scope.ScopeContext{UserID: userID, Role: role}
		// 注意：fallback 没填 CurrentBranchID — staff role 会被强制走 "no current branch" 拒绝路径。
		// 这是安全侧：未注入 ctx 时拒绝，提示 caller 接 middleware。
	}
	return sc.VerifyCard(ctx, db, cardID)
}

func VerifyDeviceInScope(db *sql.DB, ctx context.Context, userID, role, deviceAddr string) error {
	sc, ok := scope.FromContext(ctx)
	if !ok || sc == nil {
		sc = &scope.ScopeContext{UserID: userID, Role: role}
	}
	return sc.VerifyDevice(ctx, db, deviceAddr)
}

// 兼容旧 caller（不传 role 的早期版本）
func VerifyCardInCurrentBranch(db *sql.DB, ctx context.Context, userID, cardID string) error {
	return VerifyCardInScope(db, ctx, userID, "", cardID)
}
func VerifyDeviceInCurrentBranch(db *sql.DB, ctx context.Context, userID, deviceAddr string) error {
	return VerifyDeviceInScope(db, ctx, userID, "", deviceAddr)
}

// ApplyBranchFilter 应用 branch 过滤条件到 SQL 查询
// 实现空值匹配逻辑：
//   - 当 userBranchTag IS NULL 时，匹配 units.branch_tag IS NULL OR units.branch_tag = 'default'
//   - 当 userBranchTag 有值时，匹配 units.branch_tag = userBranchTag
//
// 参数:
//   - query: SQL 查询字符串（会被修改，追加 WHERE 或 AND 条件）
//   - args: SQL 参数数组（会被修改，追加参数）
//   - userBranchTag: 用户的 branch_tag（可能为 NULL）
//   - tableAlias: 表别名（如 "u" 表示 units 表）
//   - isFirstCondition: 是否是第一个 WHERE 条件（true 时使用 WHERE，false 时使用 AND）
//
// 示例:
//   - userBranchTag = NULL: WHERE (u.branch_tag IS NULL OR u.branch_tag = 'default')
//   - userBranchTag = "BranchA": WHERE u.branch_tag = $1
func ApplyBranchFilter(query *string, args *[]any, userBranchTag sql.NullString,
	tableAlias string, isFirstCondition bool) {

	if !userBranchTag.Valid || userBranchTag.String == "" {
		// 用户 branch_tag 为 NULL：只能管理 branch_tag 为 NULL 或 'default' 的资源
		condition := fmt.Sprintf(`(%s.branch_tag IS NULL OR %s.branch_tag = 'default')`, tableAlias, tableAlias)
		if isFirstCondition {
			*query += ` WHERE ` + condition
		} else {
			*query += ` AND ` + condition
		}
	} else {
		// 用户 branch_tag 有值：只能管理匹配的 branch
		*args = append(*args, userBranchTag.String)
		argIdx := len(*args)
		condition := fmt.Sprintf(`%s.branch_tag = $%d`, tableAlias, argIdx)
		if isFirstCondition {
			*query += ` WHERE ` + condition
		} else {
			*query += ` AND ` + condition
		}
	}
}
