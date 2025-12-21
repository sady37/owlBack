package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// UsersRepository 用户Repository接口
// 使用强类型领域模型，不使用map[string]any
type UsersRepository interface {
	// 用户基本信息
	GetUser(ctx context.Context, tenantID, userID string) (*domain.User, error)
	GetUserByAccount(ctx context.Context, tenantID, account string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, tenantID string, emailHash []byte) (*domain.User, error)
	GetUserByPhone(ctx context.Context, tenantID string, phoneHash []byte) (*domain.User, error)
	ListUsers(ctx context.Context, tenantID string, filters UserFilters, page, size int) ([]*domain.User, int, error)
	CreateUser(ctx context.Context, tenantID string, user *domain.User) (string, error)
	UpdateUser(ctx context.Context, tenantID, userID string, user *domain.User) error
	DeleteUser(ctx context.Context, tenantID, userID string) error

	// 用户标签管理（维护tags_catalog目录）
	// 注意：只需要调用upsert_tag_to_catalog维护目录，不需要反向索引
	SyncUserTagsToCatalog(ctx context.Context, tenantID, userID string, tags []string) error
}

// UserFilters 用户查询过滤器
type UserFilters struct {
	Role      string
	Status    string
	BranchTag string
	Tag       string // 查询包含指定tag的用户
	Search    string // 模糊搜索：支持user_account, nickname, email, phone
}

