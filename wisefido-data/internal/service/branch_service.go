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

// BranchItem 院区项（前端格式）
type BranchItem struct {
	BranchID    string   `json:"branch_id"`
	TenantID    string   `json:"tenant_id"`
	BranchName  string   `json:"branch_name"`
	Users       []string `json:"users,omitempty"`     // 该 branch_id 下的所有 user（user_account 列表）
	Units       []string `json:"units,omitempty"`     // 该 branch_id 下的所有 unit（unit_name 列表）
	Residents   []string `json:"residents,omitempty"` // 该 branch_id 下的所有 resident（nickname 列表）
	Description string   `json:"description,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
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

	// 查询院区列表
	branches, total, err := s.branchesRepo.ListBranches(ctx, req.TenantID, req.Page, req.Size)
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

		// 如果提供了搜索关键词，进行过滤
		if req.Search != "" {
			searchLower := strings.ToLower(req.Search)
			branchNameLower := strings.ToLower(item.BranchName)
			descriptionLower := strings.ToLower(item.Description)
			if !strings.Contains(branchNameLower, searchLower) && !strings.Contains(descriptionLower, searchLower) {
				continue
			}
		}

		items = append(items, item)
	}

	// 如果进行了搜索过滤或权限过滤，更新总数
	if req.Search != "" || req.BranchOnly {
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
		Users:      []string{},
		Units:      []string{},
		Residents:  []string{},
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

// getBranchUsers 获取该 branch_id 下的所有 user（user_account 列表）
func (s *BranchService) getBranchUsers(ctx context.Context, tenantID, branchID string) ([]string, error) {
	if s.db == nil {
		return []string{}, nil
	}

	query := `
		SELECT DISTINCT u.user_account
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

	users := []string{}
	for rows.Next() {
		var userAccount string
		if err := rows.Scan(&userAccount); err != nil {
			return nil, fmt.Errorf("failed to scan user account: %w", err)
		}
		users = append(users, userAccount)
	}

	return users, rows.Err()
}

// getBranchUnits 获取该 branch_id 下的所有 unit（unit_name 列表）
func (s *BranchService) getBranchUnits(ctx context.Context, tenantID, branchID string) ([]string, error) {
	if s.db == nil {
		return []string{}, nil
	}

	query := `
		SELECT DISTINCT u.unit_name
		FROM units u
		WHERE u.tenant_id = $1 AND u.branch_id::text = $2
		ORDER BY u.unit_name ASC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query branch units: %w", err)
	}
	defer rows.Close()

	units := []string{}
	for rows.Next() {
		var unitName string
		if err := rows.Scan(&unitName); err != nil {
			return nil, fmt.Errorf("failed to scan unit name: %w", err)
		}
		units = append(units, unitName)
	}

	return units, rows.Err()
}

// getBranchResidents 获取该 branch_id 下的所有 resident（nickname 列表）
// 通过 residents.unit_id -> units.branch_id 关联
func (s *BranchService) getBranchResidents(ctx context.Context, tenantID, branchID string) ([]string, error) {
	if s.db == nil {
		return []string{}, nil
	}

	query := `
		SELECT DISTINCT r.nickname
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

	residents := []string{}
	for rows.Next() {
		var nickname string
		if err := rows.Scan(&nickname); err != nil {
			return nil, fmt.Errorf("failed to scan resident nickname: %w", err)
		}
		residents = append(residents, nickname)
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
	if branchName == "" {
		// 如果 branch_name 为空，自动设置为 '-'
		branchName = "-"
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

	// 处理 branch_name
	if req.BranchName != nil {
		branchName := strings.TrimSpace(*req.BranchName)
		if branchName == "" {
			branchName = "-"
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
