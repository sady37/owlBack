package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// TagsRepository 标签Repository接口
// 使用强类型领域模型，不使用map[string]any
// 设计原则：从底层（数据库）向上设计，Repository层只负责数据访问
type TagsRepository interface {
	// ========== 查询（单个）==========
	// GetTag 根据tag_id获取tag
	GetTag(ctx context.Context, tenantID, tagID string) (*domain.Tag, error)
	
	// GetTagByName 根据tag_name获取tag（tag_name在tenant内全局唯一）
	GetTagByName(ctx context.Context, tenantID, tagName string) (*domain.Tag, error)

	// ========== 查询（列表）==========
	// ListTags 查询tags列表（支持分页、过滤、搜索）
	// 注意：统一使用此方法，不再使用GetTagsForTenant
	ListTags(ctx context.Context, tenantID string, filter TagsFilter, page, size int) ([]*domain.Tag, int, error)

	// ========== 创建/更新 ==========
	// CreateTag 创建tag（调用upsert_tag_to_catalog函数）
	// 注意：
	//   - 使用upsert语义：如果tag_name已存在，更新tag_type
	//   - 只允许创建family_tag和user_tag，branch_tag和area_tag是系统自动维护的
	//   - tag_id基于tag_name确定性生成，即使tag_name改名，tag_id也不变
	CreateTag(ctx context.Context, tenantID string, tag *domain.Tag) (string, error)
	
	// UpdateTag 更新tag（调用upsert_tag_to_catalog函数）
	// 注意：
	//   - 只能更新tag_type（tag_name不能修改，因为tag_id基于tag_name生成）
	//   - 权限控制由Service层处理（SystemAdmin可以修改，其他用户不能修改）
	UpdateTag(ctx context.Context, tenantID, tagName string, tag *domain.Tag) error

	// ========== 删除 ==========
	// DeleteTag 删除tag（调用drop_tag函数）
	// 注意：
	//   - 使用tag_name（因为tag_name在tenant内全局唯一）
	//   - 只允许删除family_tag和user_tag，branch_tag和area_tag不能删除
	//   - 删除前会检查tag是否还在源表中使用（residents.family_tag, units.branch_tag, units.area_tag, users.tags）
	DeleteTag(ctx context.Context, tenantID, tagName string) error

	// DeleteTagType 删除标签类型（调用drop_tag_type函数）
	// 注意：
	//   - 系统预定义类型不能删除
	//   - 会将该类型下的所有tag_name的tag_type设置为NULL
	DeleteTagType(ctx context.Context, tenantID, tagType string) (int, error)

	// UpdateTagName 更新标签名称
	// 注意：tag_id基于tag_name生成，修改tag_name不会改变tag_id
	UpdateTagName(ctx context.Context, tenantID, tagID, newTagName string) error

	// AddTagObject 添加标签对象（调用update_tag_objects函数，如果存在）
	// 注意：update_tag_objects函数可能已删除，需要检查
	AddTagObject(ctx context.Context, tagID, objectType, objectID, objectName string) error

	// RemoveTagObject 删除标签对象（调用update_tag_objects函数，如果存在）
	RemoveTagObject(ctx context.Context, tagID, objectType, objectID string) error

	// SyncUserTag 同步用户标签到users.tags JSONB
	SyncUserTag(ctx context.Context, tagName, userID string, add bool) error

	// SyncResidentFamilyTag 同步住户家庭标签
	SyncResidentFamilyTag(ctx context.Context, tagName, residentID string, clear bool) error
}

// TagsFilter 标签查询过滤器
type TagsFilter struct {
	TagType           string // 可选，按tag_type过滤
	TagName           string // 可选，按tag_name搜索（模糊匹配）
	IncludeSystemTags bool   // 是否包含系统预定义tag类型（branch_tag, family_tag, area_tag）
}
