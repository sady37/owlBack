package httpapi

import (
	"context"
	"database/sql"
	"fmt"
)

// PermissionCheck 权限检查结果
// 包含 assigned_only 和 branch_only 两个标志
type PermissionCheck struct {
	AssignedOnly bool // 是否仅限分配的资源
	BranchOnly   bool // 是否仅限同一 Branch 的资源
}

// GetResourcePermission 查询资源权限配置
// 从 role_permissions 表中查询指定角色对指定资源的权限配置
// 返回 assigned_only 和 branch_only 标志
//
// 参数:
//   - db: 数据库连接
//   - ctx: 上下文
//   - roleCode: 角色代码（如 "Manager", "Admin"）
//   - resourceType: 资源类型（如 "residents", "users"）
//   - permissionType: 权限类型（"R", "C", "U", "D"）
//
// 返回:
//   - *PermissionCheck: 权限检查结果，如果查询失败或记录不存在，返回最严格的权限（assigned_only=true, branch_only=true）
//   - error: 查询错误（如果记录不存在，返回 nil，使用默认值）
func GetResourcePermission(db *sql.DB, ctx context.Context,
	roleCode, resourceType, permissionType string) (*PermissionCheck, error) {

	var assignedOnly, branchOnly bool
	err := db.QueryRowContext(ctx,
		`SELECT 
			COALESCE(assigned_only, FALSE) as assigned_only,
			COALESCE(branch_only, FALSE) as branch_only
		 FROM role_permissions
		 WHERE tenant_id = $1 
		   AND role_code = $2 
		   AND resource_type = $3 
		   AND permission_type = $4
		 LIMIT 1`,
		SystemTenantID(), roleCode, resourceType, permissionType,
	).Scan(&assignedOnly, &branchOnly)

	if err == sql.ErrNoRows {
		// 记录不存在：返回最严格的权限（安全默认值）
		return &PermissionCheck{AssignedOnly: true, BranchOnly: true}, nil
	}
	if err != nil {
		return nil, err
	}

	return &PermissionCheck{AssignedOnly: assignedOnly, BranchOnly: branchOnly}, nil
}

// ApplyBranchFilter 应用 branch 过滤条件到 SQL 查询
// 实现空值匹配逻辑：
//   - 当 userBranchTag IS NULL 时，匹配 units.branch_tag IS NULL OR units.branch_tag = '-'
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
//   - userBranchTag = NULL: WHERE (u.branch_tag IS NULL OR u.branch_tag = '-')
//   - userBranchTag = "BranchA": WHERE u.branch_tag = $1
func ApplyBranchFilter(query *string, args *[]any, userBranchTag sql.NullString,
	tableAlias string, isFirstCondition bool) {

	if !userBranchTag.Valid || userBranchTag.String == "" {
		// 用户 branch_tag 为 NULL：只能管理 branch_tag 为 NULL 或 '-' 的资源
		condition := fmt.Sprintf(`(%s.branch_tag IS NULL OR %s.branch_tag = '-')`, tableAlias, tableAlias)
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
