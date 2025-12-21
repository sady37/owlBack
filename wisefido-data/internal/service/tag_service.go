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

// TagService 标签服务
type TagService struct {
	tagRepo repository.TagsRepository
	db      *sql.DB // 用于复杂查询（JOIN、从源表查询标签）
	logger  *zap.Logger
}

// NewTagService 创建标签服务
func NewTagService(tagRepo repository.TagsRepository, db *sql.DB, logger *zap.Logger) *TagService {
	return &TagService{
		tagRepo: tagRepo,
		db:      db,
		logger:  logger,
	}
}

// ListTagsRequest 查询标签列表请求
type ListTagsRequest struct {
	TenantID          string
	UserRole          string
	TagType           string
	IncludeSystemTags bool
	Page              int
	Size              int
}

// ListTagsResponse 查询标签列表响应
type ListTagsResponse struct {
	Items                    []TagItem `json:"items"`
	Total                    int       `json:"total"`
	AvailableTagTypes        []string  `json:"available_tag_types"`
	SystemPredefinedTagTypes []string  `json:"system_predefined_tag_types"`
}

// TagItem 标签项（前端格式）
type TagItem struct {
	TagID          string  `json:"tag_id"`
	TenantID       string  `json:"tenant_id,omitempty"` // 可选，GetTagsForObject 响应中不包含
	TagType        string  `json:"tag_type"`
	TagName        string  `json:"tag_name"`
	ObjectNameInTag *string `json:"object_name_in_tag,omitempty"` // 对象在 tag 中的名称（GetTagsForObject 使用）
}

// ListTags 查询标签列表
func (s *TagService) ListTags(ctx context.Context, req ListTagsRequest) (*ListTagsResponse, error) {
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

	// 构建过滤器
	filter := repository.TagsFilter{
		TagType:           strings.TrimSpace(req.TagType),
		IncludeSystemTags: req.IncludeSystemTags,
	}

	// 查询标签列表
	tags, total, err := s.tagRepo.ListTags(ctx, req.TenantID, filter, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	// 转换为前端格式
	items := make([]TagItem, 0, len(tags))
	for _, tag := range tags {
		items = append(items, TagItem{
			TagID:    tag.TagID,
			TenantID: tag.TenantID,
			TagType:  tag.TagType,
			TagName:  tag.TagName,
		})
	}

	return &ListTagsResponse{
		Items:                    items,
		Total:                    total,
		AvailableTagTypes:        []string{"branch_tag", "family_tag", "area_tag", "user_tag"},
		SystemPredefinedTagTypes: []string{"branch_tag", "family_tag", "area_tag"},
	}, nil
}

// GetTagRequest 查询标签详情请求
type GetTagRequest struct {
	TenantID string
	TagID    string
}

// GetTag 查询标签详情
func (s *TagService) GetTag(ctx context.Context, req GetTagRequest) (*TagItem, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.TagID == "" {
		return nil, fmt.Errorf("tag_id is required")
	}

	tag, err := s.tagRepo.GetTag(ctx, req.TenantID, req.TagID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return &TagItem{
		TagID:    tag.TagID,
		TenantID: tag.TenantID,
		TagType:  tag.TagType,
		TagName:  tag.TagName,
	}, nil
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	TenantID string
	UserRole string
	TagName  string
	TagType  string // 可选，默认为 "user_tag"
}

// CreateTagResponse 创建标签响应
type CreateTagResponse struct {
	TagID string `json:"tag_id"`
}

// CreateTag 创建标签
func (s *TagService) CreateTag(ctx context.Context, req CreateTagRequest) (*CreateTagResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if strings.TrimSpace(req.TagName) == "" {
		return nil, fmt.Errorf("tag_name is required")
	}

	// 标签类型验证
	tagType := strings.TrimSpace(req.TagType)
	if tagType == "" {
		tagType = "user_tag" // 默认类型
	}

	// 验证标签类型
	allowedTypes := map[string]bool{
		"branch_tag": true,
		"family_tag": true,
		"area_tag":   true,
		"user_tag":   true,
	}
	if !allowedTypes[tagType] {
		return nil, fmt.Errorf("invalid tag_type: %s", tagType)
	}

	// 创建标签
	tag := &domain.Tag{
		TagName: strings.TrimSpace(req.TagName),
		TagType: tagType,
	}

	tagID, err := s.tagRepo.CreateTag(ctx, req.TenantID, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return &CreateTagResponse{
		TagID: tagID,
	}, nil
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	TenantID string
	UserRole string
	TagID    string
	TagName  string
}

// UpdateTag 更新标签名称
func (s *TagService) UpdateTag(ctx context.Context, req UpdateTagRequest) error {
	// 参数验证
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if req.TagID == "" {
		return fmt.Errorf("tag_id is required")
	}
	if strings.TrimSpace(req.TagName) == "" {
		return fmt.Errorf("tag_name is required")
	}

	// 查询现有标签（用于验证）
	_, err := s.tagRepo.GetTag(ctx, req.TenantID, req.TagID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag not found")
		}
		return fmt.Errorf("failed to get tag: %w", err)
	}

	// 更新标签名称
	// 注意：tag_id 在创建时基于 tag_name 确定性生成（UUID v5），但生成后就不变了
	// 即使 tag_name 修改，tag_id 也不会变化（因为 tag_id 是主键，不会自动重新计算）
	// 所以可以直接更新 tag_name，tag_id 保持不变
	err = s.tagRepo.UpdateTagName(ctx, req.TenantID, req.TagID, strings.TrimSpace(req.TagName))
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	return nil
}

// DeleteTagRequest 删除标签请求
type DeleteTagRequest struct {
	TenantID string
	UserRole string
	TagName  string // 使用 tag_name（全局唯一）
}

// DeleteTag 删除标签（调用 drop_tag 函数）
func (s *TagService) DeleteTag(ctx context.Context, req DeleteTagRequest) error {
	// 参数验证
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if strings.TrimSpace(req.TagName) == "" {
		return fmt.Errorf("tag_name is required")
	}

	// 业务规则验证：查询 tag 信息
	tag, err := s.tagRepo.GetTagByName(ctx, req.TenantID, strings.TrimSpace(req.TagName))
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag not found: %s", req.TagName)
		}
		return fmt.Errorf("failed to get tag: %w", err)
	}

	// 系统预定义类型不能删除
	tagType := domain.TagType(tag.TagType)
	if tagType.IsSystemTagType() {
		return fmt.Errorf("cannot delete system predefined tag type: %s", tag.TagType)
	}

	// 调用 Repository（Repository 调用数据库函数 drop_tag）
	// 数据库函数会自动清理所有使用该 tag 的地方
	err = s.tagRepo.DeleteTag(ctx, req.TenantID, strings.TrimSpace(req.TagName))
	if err != nil {
		// 数据库函数会检查是否还在使用，如果还在使用会返回错误
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// DeleteTagTypeRequest 删除标签类型请求
type DeleteTagTypeRequest struct {
	TenantID string
	UserRole string
	TagType  string
}

// DeleteTagType 删除标签类型（删除所有指定类型的标签）
func (s *TagService) DeleteTagType(ctx context.Context, req DeleteTagTypeRequest) error {
	// 参数验证
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if strings.TrimSpace(req.TagType) == "" {
		return fmt.Errorf("tag_type is required")
	}

	// 权限检查：只有 SystemAdmin 可以删除标签类型
	if !strings.EqualFold(req.UserRole, "SystemAdmin") {
		return fmt.Errorf("permission denied: only SystemAdmin can delete tag type")
	}

	// 业务规则验证：系统预定义类型不能删除
	tagType := domain.TagType(strings.TrimSpace(req.TagType))
	if tagType.IsSystemTagType() {
		return fmt.Errorf("cannot delete system predefined tag type: %s", req.TagType)
	}

	// 调用 Repository 删除标签类型
	updatedCount, err := s.tagRepo.DeleteTagType(ctx, req.TenantID, strings.TrimSpace(req.TagType))
	if err != nil {
		return fmt.Errorf("failed to delete tag type: %w", err)
	}

	s.logger.Info("Deleted tag type", zap.String("tag_type", req.TagType), zap.Int("updated_count", updatedCount))
	return nil
}

// TagObject 标签对象
type TagObject struct {
	ObjectID   string `json:"object_id"`
	ObjectName string `json:"object_name"`
}

// AddTagObjectsRequest 添加标签对象请求
type AddTagObjectsRequest struct {
	TenantID   string
	UserRole   string
	TagID      string
	ObjectType string // "user", "resident", "unit"
	Objects    []TagObject
}

// AddTagObjects 添加标签对象（成员）
func (s *TagService) AddTagObjects(ctx context.Context, req AddTagObjectsRequest) error {
	// 参数验证
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if req.TagID == "" {
		return fmt.Errorf("tag_id is required")
	}
	if req.ObjectType == "" {
		return fmt.Errorf("object_type is required")
	}
	if len(req.Objects) == 0 {
		return fmt.Errorf("objects are required")
	}

	// 查询标签信息
	tag, err := s.tagRepo.GetTag(ctx, req.TenantID, req.TagID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag not found")
		}
		return fmt.Errorf("failed to get tag: %w", err)
	}

	// 添加每个对象
	for _, obj := range req.Objects {
		if obj.ObjectID == "" || obj.ObjectName == "" {
			continue
		}

		// 调用 Repository 添加标签对象
		err := s.tagRepo.AddTagObject(ctx, req.TagID, req.ObjectType, obj.ObjectID, obj.ObjectName)
		if err != nil {
			// 如果 update_tag_objects 函数不存在，记录警告但继续
			if strings.Contains(err.Error(), "not available") {
				s.logger.Warn("update_tag_objects function not available", zap.Error(err))
			} else {
				return fmt.Errorf("failed to add tag object: %w", err)
			}
		}

		// 如果是 user_tag 类型，同步更新 users.tags
		if req.ObjectType == "user" && tag.TagType == "user_tag" {
			err = s.tagRepo.SyncUserTag(ctx, tag.TagName, obj.ObjectID, true)
			if err != nil {
				s.logger.Warn("Failed to sync tag to user's tags", zap.Error(err))
				// 不失败整个操作，只记录警告
			}
		}
	}

	return nil
}

// RemoveTagObjectsRequest 删除标签对象请求
type RemoveTagObjectsRequest struct {
	TenantID   string
	UserRole   string
	TagID      string
	ObjectType string
	ObjectIDs  []string  // 支持 object_ids 格式
	Objects    []TagObject  // 支持 objects 格式
}

// RemoveTagObjects 删除标签对象（成员）
func (s *TagService) RemoveTagObjects(ctx context.Context, req RemoveTagObjectsRequest) error {
	// 参数验证
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if req.TagID == "" {
		return fmt.Errorf("tag_id is required")
	}
	if req.ObjectType == "" {
		return fmt.Errorf("object_type is required")
	}

	// 查询标签信息
	tag, err := s.tagRepo.GetTag(ctx, req.TenantID, req.TagID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag not found")
		}
		return fmt.Errorf("failed to get tag: %w", err)
	}

	// 处理 object_ids 格式
	if len(req.ObjectIDs) > 0 {
		for _, objectID := range req.ObjectIDs {
			if objectID == "" {
				continue
			}

			// 调用 Repository 删除标签对象
			err := s.tagRepo.RemoveTagObject(ctx, req.TagID, req.ObjectType, objectID)
			if err != nil {
				// 如果 update_tag_objects 函数不存在，记录警告但继续
				if strings.Contains(err.Error(), "not available") {
					s.logger.Warn("update_tag_objects function not available", zap.Error(err))
				} else {
					return fmt.Errorf("failed to remove tag object: %w", err)
				}
			}

			// 如果是 user_tag 类型，同步更新 users.tags
			if req.ObjectType == "user" && tag.TagType == "user_tag" {
				err = s.tagRepo.SyncUserTag(ctx, tag.TagName, objectID, false)
				if err != nil {
					s.logger.Warn("Failed to remove tag from user's tags", zap.Error(err))
				}
			}

			// 如果是 family_tag 类型，同步清除 residents.family_tag
			if req.ObjectType == "resident" && tag.TagType == "family_tag" {
				err = s.tagRepo.SyncResidentFamilyTag(ctx, tag.TagName, objectID, true)
				if err != nil {
					s.logger.Warn("Failed to clear family_tag from resident", zap.Error(err))
				}
			}
		}
	}

	// 处理 objects 格式
	if len(req.Objects) > 0 {
		for _, obj := range req.Objects {
			if obj.ObjectID == "" {
				continue
			}

			// 调用 Repository 删除标签对象
			err := s.tagRepo.RemoveTagObject(ctx, req.TagID, req.ObjectType, obj.ObjectID)
			if err != nil {
				// 如果 update_tag_objects 函数不存在，记录警告但继续
				if strings.Contains(err.Error(), "not available") {
					s.logger.Warn("update_tag_objects function not available", zap.Error(err))
				} else {
					return fmt.Errorf("failed to remove tag object: %w", err)
				}
			}

			// 如果是 user_tag 类型，同步更新 users.tags
			if req.ObjectType == "user" && tag.TagType == "user_tag" {
				err = s.tagRepo.SyncUserTag(ctx, tag.TagName, obj.ObjectID, false)
				if err != nil {
					s.logger.Warn("Failed to remove tag from user's tags", zap.Error(err))
				}
			}

			// 如果是 family_tag 类型，同步清除 residents.family_tag
			if req.ObjectType == "resident" && tag.TagType == "family_tag" {
				err = s.tagRepo.SyncResidentFamilyTag(ctx, tag.TagName, obj.ObjectID, true)
				if err != nil {
					s.logger.Warn("Failed to clear family_tag from resident", zap.Error(err))
				}
			}
		}
	}

	if len(req.ObjectIDs) == 0 && len(req.Objects) == 0 {
		return fmt.Errorf("object_ids or objects is required")
	}

	return nil
}

// GetTagsForObjectRequest 查询对象标签请求
type GetTagsForObjectRequest struct {
	TenantID   string
	ObjectType string
	ObjectID   string
}

// GetTagsForObjectResponse 查询对象标签响应
type GetTagsForObjectResponse struct {
	Items []TagItem `json:"items"`
}

// GetTagsForObject 查询对象标签
func (s *TagService) GetTagsForObject(ctx context.Context, req GetTagsForObjectRequest) (*GetTagsForObjectResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.ObjectType == "" || req.ObjectID == "" {
		return nil, fmt.Errorf("object_type and object_id are required")
	}

	// 从源表查询标签（tag_objects 字段已删除）
	// 根据 object_type 从不同的源表查询：
	// - user: 从 users.tags JSONB 字段查询
	// - resident: 从 residents.family_tag 查询
	// - unit: 从 units.branch_tag 和 units.area_tag 查询
	items := make([]TagItem, 0)

	switch req.ObjectType {
	case "user":
		// 查询 users.tags JSONB 字段
		// 需要查询哪些 tag_name 在 users.tags 数组中
		query := `
			SELECT DISTINCT tc.tag_id::text, tc.tag_type, tc.tag_name, COALESCE(u.nickname, '') as object_name_in_tag
			FROM tags_catalog tc
			INNER JOIN users u ON u.tenant_id = tc.tenant_id AND u.user_id::text = $2
			WHERE tc.tenant_id = $1
			  AND u.tags IS NOT NULL
			  AND u.tags ? tc.tag_name
		`
		rows, err := s.db.QueryContext(ctx, query, req.TenantID, req.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to query user tags: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tagID, tagType, tagName, objectNameInTag sql.NullString
			if err := rows.Scan(&tagID, &tagType, &tagName, &objectNameInTag); err != nil {
				return nil, fmt.Errorf("failed to scan user tag: %w", err)
			}
			if tagID.Valid && tagType.Valid && tagName.Valid {
				var objectName *string
				if objectNameInTag.Valid && objectNameInTag.String != "" {
					objectName = &objectNameInTag.String
				}
				items = append(items, TagItem{
					TagID:          tagID.String,
					TagType:        tagType.String,
					TagName:        tagName.String,
					ObjectNameInTag: objectName,
				})
			}
		}

	case "resident":
		// 查询 residents.family_tag
		query := `
			SELECT DISTINCT tc.tag_id::text, tc.tag_type, tc.tag_name, COALESCE(r.nickname, '') as object_name_in_tag
			FROM tags_catalog tc
			INNER JOIN residents r ON r.tenant_id = tc.tenant_id AND r.resident_id::text = $2
			WHERE tc.tenant_id = $1
			  AND r.family_tag IS NOT NULL
			  AND r.family_tag = tc.tag_name
		`
		rows, err := s.db.QueryContext(ctx, query, req.TenantID, req.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to query resident tags: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tagID, tagType, tagName, objectNameInTag sql.NullString
			if err := rows.Scan(&tagID, &tagType, &tagName, &objectNameInTag); err != nil {
				return nil, fmt.Errorf("failed to scan resident tag: %w", err)
			}
			if tagID.Valid && tagType.Valid && tagName.Valid {
				var objectName *string
				if objectNameInTag.Valid && objectNameInTag.String != "" {
					objectName = &objectNameInTag.String
				}
				items = append(items, TagItem{
					TagID:          tagID.String,
					TagType:        tagType.String,
					TagName:        tagName.String,
					ObjectNameInTag: objectName,
				})
			}
		}

	case "unit":
		// 查询 units.branch_tag 和 units.area_tag
		query := `
			SELECT DISTINCT tc.tag_id::text, tc.tag_type, tc.tag_name, COALESCE(u.unit_name, '') as object_name_in_tag
			FROM tags_catalog tc
			INNER JOIN units u ON u.tenant_id = tc.tenant_id AND u.unit_id::text = $2
			WHERE tc.tenant_id = $1
			  AND (u.branch_tag = tc.tag_name OR u.area_tag = tc.tag_name)
		`
		rows, err := s.db.QueryContext(ctx, query, req.TenantID, req.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to query unit tags: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tagID, tagType, tagName, objectNameInTag sql.NullString
			if err := rows.Scan(&tagID, &tagType, &tagName, &objectNameInTag); err != nil {
				return nil, fmt.Errorf("failed to scan unit tag: %w", err)
			}
			if tagID.Valid && tagType.Valid && tagName.Valid {
				var objectName *string
				if objectNameInTag.Valid && objectNameInTag.String != "" {
					objectName = &objectNameInTag.String
				}
				items = append(items, TagItem{
					TagID:          tagID.String,
					TagType:        tagType.String,
					TagName:        tagName.String,
					ObjectNameInTag: objectName,
				})
			}
		}

	default:
		return nil, fmt.Errorf("unsupported object_type: %s. Supported types: user, resident, unit", req.ObjectType)
	}

	return &GetTagsForObjectResponse{
		Items: items,
	}, nil
}

