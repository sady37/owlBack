package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// BranchService 院区服务
type BranchService struct {
	branchesRepo repository.BranchesRepository
	db           *sql.DB // 用于查询 user_branches 表（Manager 权限过滤）
	logger       *zap.Logger
}

// NewBranchService 创建院区服务
func NewBranchService(branchesRepo repository.BranchesRepository, db *sql.DB, logger *zap.Logger) *BranchService {
	return &BranchService{
		branchesRepo: branchesRepo,
		db:           db,
		logger:       logger,
	}
}

// ListBranchesRequest 查询院区列表请求
type ListBranchesRequest struct {
	TenantID        string // 必填
	CurrentUserID   string // 当前用户 ID（用于 Manager 权限过滤）
	CurrentUserRole string // 当前用户角色（用于权限过滤）
	BranchOnly      bool   // 是否只显示本 Branch（Manager 权限）
	Search          string
	Page            int
	Size            int
}

// ListBranchesResponse 查询院区列表响应
type ListBranchesResponse struct {
	Items []BranchItem `json:"items"`
	Total int          `json:"total"`
}

// BranchUserItem 院区用户项（id + account + nickname）
type BranchUserItem struct {
	UserID      string `json:"user_id"`
	UserAccount string `json:"user_account"`
	Nickname    string `json:"nickname,omitempty"`
}

// BranchUnitItem 院区单元项（id + name）
type BranchUnitItem struct {
	UnitID   string `json:"unit_id"`
	UnitName string `json:"unit_name"`
}

// BranchResidentItem 院区住户项（id + nickname）
type BranchResidentItem struct {
	ResidentID string `json:"resident_id"`
	Nickname   string `json:"nickname"`
}

// BranchItem 院区项（前端格式）
type BranchItem struct {
	BranchID    string               `json:"branch_id"`
	TenantID    string               `json:"tenant_id"`
	BranchName  string               `json:"branch_name"`
	Users       []BranchUserItem     `json:"users,omitempty"`     // 该 branch_id 下的所有 user（id + account）
	Units       []BranchUnitItem     `json:"units,omitempty"`     // 该 branch_id 下的所有 unit（id + name）
	Residents   []BranchResidentItem `json:"residents,omitempty"` // 该 branch_id 下的所有 resident（id + nickname）
	Description string               `json:"description,omitempty"`
	CreatedAt   string               `json:"created_at,omitempty"`
	UpdatedAt   string               `json:"updated_at,omitempty"`
}

// ListBranches 查询院区列表
func (s *BranchService) ListBranches(ctx context.Context, req ListBranchesRequest) (*ListBranchesResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	// Manager 权限过滤：只显示本 Branch
	var allowedBranchIDs []string
	if req.BranchOnly && req.CurrentUserID != "" && s.db != nil {
		// 查询用户关联的所有 branch_id
		rows, err := s.db.QueryContext(ctx,
			`SELECT DISTINCT branch_id::text
			 FROM user_branches
			 WHERE tenant_id = $1 AND user_id::text = $2`,
			req.TenantID, req.CurrentUserID,
		)
		if err != nil {
			s.logger.Warn("Failed to query user branches", zap.Error(err))
		} else {
			defer rows.Close()
			for rows.Next() {
				var branchID string
				if err := rows.Scan(&branchID); err == nil {
					allowedBranchIDs = append(allowedBranchIDs, branchID)
				}
			}
		}
		// 如果用户没有关联任何 branch，返回空列表（Manager 只能查看本 Branch）
		if len(allowedBranchIDs) == 0 {
			return &ListBranchesResponse{
				Items: []BranchItem{},
				Total: 0,
			}, nil
		}
	}

	// 查询院区列表（搜索在 SQL 查询中进行）
	branches, total, err := s.branchesRepo.ListBranches(ctx, req.TenantID, req.Search, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	// 转换为前端格式
	items := make([]BranchItem, 0, len(branches))
	for _, branch := range branches {
		// Manager 权限过滤：只显示本 Branch
		if req.BranchOnly && len(allowedBranchIDs) > 0 {
			found := false
			for _, allowedID := range allowedBranchIDs {
				if branch.BranchID == allowedID {
					found = true
					break
				}
			}
			if !found {
				continue // 跳过不在允许列表中的 branch
			}
		}

		item := s.branchToItem(branch)

		// 获取该 branch 下的 users
		if s.db != nil {
			users, err := s.getBranchUsers(ctx, req.TenantID, branch.BranchID)
			if err != nil {
				s.logger.Warn("Failed to get branch users", zap.String("branch_id", branch.BranchID), zap.Error(err))
			} else {
				item.Users = users
			}

			// 获取该 branch 下的 units
			units, err := s.getBranchUnits(ctx, req.TenantID, branch.BranchID)
			if err != nil {
				s.logger.Warn("Failed to get branch units", zap.String("branch_id", branch.BranchID), zap.Error(err))
			} else {
				item.Units = units
			}

			// 获取该 branch 下的 residents
			residents, err := s.getBranchResidents(ctx, req.TenantID, branch.BranchID)
			if err != nil {
				s.logger.Warn("Failed to get branch residents", zap.String("branch_id", branch.BranchID), zap.Error(err))
			} else {
				item.Residents = residents
			}
		}

		items = append(items, item)
	}

	// 如果进行了权限过滤，更新总数（搜索过滤已经在 SQL 查询中完成）
	if req.BranchOnly {
		total = len(items)
	}

	return &ListBranchesResponse{
		Items: items,
		Total: total,
	}, nil
}

// branchToItem 将 domain.Branch 转换为 BranchItem
func (s *BranchService) branchToItem(branch *domain.Branch) BranchItem {
	item := BranchItem{
		BranchID:   branch.BranchID,
		TenantID:   branch.TenantID,
		BranchName: branch.BranchName,
		Users:      []BranchUserItem{},
		Units:      []BranchUnitItem{},
		Residents:  []BranchResidentItem{},
	}

	if branch.Description.Valid {
		item.Description = branch.Description.String
	}
	if branch.CreatedAt.Valid {
		item.CreatedAt = branch.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	if branch.UpdatedAt.Valid {
		item.UpdatedAt = branch.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	return item
}

// getBranchUsers 获取该 branch_id 下的所有 user（id + account + nickname）
func (s *BranchService) getBranchUsers(ctx context.Context, tenantID, branchID string) ([]BranchUserItem, error) {
	if s.db == nil {
		return []BranchUserItem{}, nil
	}

	query := `
		SELECT DISTINCT u.user_id::text, u.user_account, COALESCE(u.nickname, '') as nickname
		FROM users u
		INNER JOIN user_branches ub ON ub.user_id = u.user_id AND ub.tenant_id = u.tenant_id
		WHERE u.tenant_id = $1 AND ub.branch_id::text = $2
		ORDER BY u.user_account ASC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query branch users: %w", err)
	}
	defer rows.Close()

	users := []BranchUserItem{}
	for rows.Next() {
		var userItem BranchUserItem
		if err := rows.Scan(&userItem.UserID, &userItem.UserAccount, &userItem.Nickname); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, userItem)
	}

	return users, rows.Err()
}

// getBranchUnits 获取该 branch_id 下的所有 unit（id + name）
func (s *BranchService) getBranchUnits(ctx context.Context, tenantID, branchID string) ([]BranchUnitItem, error) {
	if s.db == nil {
		return []BranchUnitItem{}, nil
	}

	query := `
		SELECT DISTINCT u.unit_id::text, u.unit_name
		FROM units u
		WHERE u.tenant_id = $1 AND u.branch_id::text = $2
		ORDER BY u.unit_name ASC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query branch units: %w", err)
	}
	defer rows.Close()

	units := []BranchUnitItem{}
	for rows.Next() {
		var unitItem BranchUnitItem
		if err := rows.Scan(&unitItem.UnitID, &unitItem.UnitName); err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, unitItem)
	}

	return units, rows.Err()
}

// getBranchResidents 获取该 branch_id 下的所有 resident（id + nickname）
// 通过 residents.unit_id -> units.branch_id 关联
func (s *BranchService) getBranchResidents(ctx context.Context, tenantID, branchID string) ([]BranchResidentItem, error) {
	if s.db == nil {
		return []BranchResidentItem{}, nil
	}

	query := `
		SELECT DISTINCT r.resident_id::text, r.nickname
		FROM residents r
		INNER JOIN units u ON u.unit_id = r.unit_id AND u.tenant_id = r.tenant_id
		WHERE r.tenant_id = $1 AND u.branch_id::text = $2
		ORDER BY r.nickname ASC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query branch residents: %w", err)
	}
	defer rows.Close()

	residents := []BranchResidentItem{}
	for rows.Next() {
		var residentItem BranchResidentItem
		if err := rows.Scan(&residentItem.ResidentID, &residentItem.Nickname); err != nil {
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}
		residents = append(residents, residentItem)
	}

	return residents, rows.Err()
}

// CreateBranchRequest 创建院区请求
type CreateBranchRequest struct {
	TenantID    string
	BranchName  string
	Description string
}

// CreateBranchResponse 创建院区响应
type CreateBranchResponse struct {
	BranchID string `json:"branch_id"`
}

// CreateBranch 创建院区
func (s *BranchService) CreateBranch(ctx context.Context, req CreateBranchRequest) (*CreateBranchResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// Service 层：trim 空格但保留大小写
	branchName := strings.TrimSpace(req.BranchName)
	// 验证：禁止空字符串作为 BranchName
	if branchName == "" {
		return nil, fmt.Errorf("branch_name is required and cannot be empty")
	}

	// 创建院区领域模型
	branch := &domain.Branch{
		BranchName: branchName,
	}
	if req.Description != "" {
		branch.Description = sql.NullString{String: strings.TrimSpace(req.Description), Valid: true}
	}

	// 调用 Repository
	branchID, err := s.branchesRepo.CreateBranch(ctx, req.TenantID, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// 自动为所有 Admin 用户创建 branch 关联
	// 使用事务确保原子性
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Warn("Failed to begin transaction for auto-linking admin users", zap.Error(err))
		// 不阻塞 branch 创建，只记录警告
	} else {
		defer tx.Rollback()

		// 查询该 tenant 下所有 Admin 角色的用户
		rows, err := tx.QueryContext(ctx,
			`SELECT user_id FROM users 
			 WHERE tenant_id = $1 AND role = 'Admin' AND status = 'active'`,
			req.TenantID,
		)
		if err != nil {
			s.logger.Warn("Failed to query admin users", zap.Error(err))
			tx.Rollback()
		} else {
			defer rows.Close()

			// 为每个 Admin 用户创建 branch 关联
			// 注意：如果用户已经有其他 branch 关联，新 branch 不作为主院区（is_primary = FALSE）
			// 如果用户没有任何 branch 关联，新 branch 作为主院区（is_primary = TRUE）
			for rows.Next() {
				var userID string
				if err := rows.Scan(&userID); err != nil {
					s.logger.Warn("Failed to scan user_id", zap.Error(err))
					continue
				}

				// 检查用户是否已有 branch 关联
				var existingCount int
				err := tx.QueryRowContext(ctx,
					`SELECT COUNT(*) FROM user_branches WHERE tenant_id = $1 AND user_id = $2`,
					req.TenantID, userID,
				).Scan(&existingCount)
				if err != nil {
					s.logger.Warn("Failed to check existing branches for user", zap.String("user_id", userID), zap.Error(err))
					continue
				}

				// 如果用户没有任何 branch 关联，新 branch 作为主院区
				isPrimary := existingCount == 0

				// 插入 user_branches 记录
				_, err = tx.ExecContext(ctx,
					`INSERT INTO user_branches (tenant_id, user_id, branch_id, is_primary)
					 VALUES ($1, $2, $3, $4)
					 ON CONFLICT (tenant_id, user_id, branch_id) DO NOTHING`,
					req.TenantID, userID, branchID, isPrimary,
				)
				if err != nil {
					s.logger.Warn("Failed to create user_branch association",
						zap.String("user_id", userID),
						zap.String("branch_id", branchID),
						zap.Error(err))
					// 继续处理下一个用户
				}
			}

			// 提交事务
			if err := tx.Commit(); err != nil {
				s.logger.Warn("Failed to commit transaction for auto-linking admin users", zap.Error(err))
			} else {
				s.logger.Info("Auto-linked new branch to all admin users",
					zap.String("branch_id", branchID),
					zap.String("tenant_id", req.TenantID))
			}
		}
	}

	return &CreateBranchResponse{
		BranchID: branchID,
	}, nil
}

// UpdateBranchRequest 更新院区请求
type UpdateBranchRequest struct {
	BranchID    string
	BranchName  *string
	Description *string
	Delete      *bool
}

// UpdateBranch 更新院区
func (s *BranchService) UpdateBranch(ctx context.Context, req UpdateBranchRequest) error {
	// 参数验证
	if req.BranchID == "" {
		return fmt.Errorf("branch_id is required")
	}

	// 处理删除
	if req.Delete != nil && *req.Delete {
		// 获取 branch 以获取 tenant_id
		branch, err := s.branchesRepo.GetBranch(ctx, "", req.BranchID)
		if err != nil {
			return fmt.Errorf("branch not found: %w", err)
		}
		return s.branchesRepo.DeleteBranch(ctx, branch.TenantID, req.BranchID)
	}

	// 获取当前院区
	branch, err := s.branchesRepo.GetBranch(ctx, "", req.BranchID)
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	// 构建更新模型
	update := &domain.BranchUpdate{}

	// 处理 branch_name：禁止空字符串
	if req.BranchName != nil {
		branchName := strings.TrimSpace(*req.BranchName)
		// 验证：禁止空字符串作为 BranchName
		if branchName == "" {
			return fmt.Errorf("branch_name cannot be empty")
		}
		update.BranchName = &domain.UpdateString{
			Action: domain.UpdateActionUpdate,
			Value:  branchName,
		}
	}

	// 处理 description
	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		if desc == "" {
			update.Description = &domain.UpdateString{
				Action: domain.UpdateActionDelete,
				Value:  "",
			}
		} else {
			update.Description = &domain.UpdateString{
				Action: domain.UpdateActionUpdate,
				Value:  desc,
			}
		}
	}

	// 调用 Repository
	return s.branchesRepo.UpdateBranchFields(ctx, branch.TenantID, req.BranchID, update)
}
