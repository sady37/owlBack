package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// BranchesRepository 院区Repository接口
// 使用强类型领域模型，不使用map[string]any
// 设计原则：从底层（数据库）向上设计，Repository层只负责数据访问
type BranchesRepository interface {
	// ========== 查询接口 ==========

	// GetBranch 获取院区信息（通过 branch_id）
	GetBranch(ctx context.Context, tenantID, branchID string) (*domain.Branch, error)

	// GetBranchByName 获取院区信息（通过 branch_name）
	// 注意：同一租户内 branch_name 唯一
	GetBranchByName(ctx context.Context, tenantID, branchName string) (*domain.Branch, error)

	// ListBranches 列出所有院区
	// 支持分页和搜索
	// search: 搜索关键词，模糊匹配 branch_name, description, user_nickname, unit_name, resident_nickname
	// 返回：院区列表和总数
	ListBranches(ctx context.Context, tenantID string, search string, page, size int) ([]*domain.Branch, int, error)

	// ========== 创建接口 ==========

	// CreateBranch 创建院区
	// 注意：
	//   - 如果 branch_name 为空，自动设置为 DefaultBranchName（见 domain.DefaultBranchName）
	//   - 唯一性约束：同一租户内 branch_name 唯一
	//   - created_at 和 updated_at 由数据库自动设置
	CreateBranch(ctx context.Context, tenantID string, branch *domain.Branch) (string, error)

	// ========== 更新接口 ==========

	// UpdateBranch 更新院区信息
	// 注意：
	//   - 如果更新 branch_name，需要检查唯一性约束
	//   - 更新时自动更新 updated_at（由数据库触发器或应用层处理）
	// Deprecated: 使用 UpdateBranchFields 替代，支持区分"不更新"、"更新"、"删除"三种状态
	UpdateBranch(ctx context.Context, tenantID, branchID string, branch *domain.Branch) error

	// UpdateBranchFields 更新院区信息（使用更新模型）
	// 注意：
	//   - 如果更新 branch_name，需要检查唯一性约束
	//   - 更新时自动更新 updated_at（由数据库触发器或应用层处理）
	//   - 支持区分"不更新"、"更新"、"删除"三种状态
	UpdateBranchFields(ctx context.Context, tenantID, branchID string, update *domain.BranchUpdate) error

	// ========== 删除接口 ==========

	// DeleteBranch 删除院区
	// 注意：
	//   - 删除时，关联的 buildings/units/users 的 branch_id 会被设置为 NULL（ON DELETE SET NULL）
	//   - 关联的 user_branches 记录会被删除（ON DELETE CASCADE）
	DeleteBranch(ctx context.Context, tenantID, branchID string) error
}
