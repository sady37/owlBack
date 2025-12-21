package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresTagsRepository 标签Repository实现（强类型版本）
// 实现TagsRepository接口，使用domain.Tag领域模型
type PostgresTagsRepository struct {
	db *sql.DB
}

// NewPostgresTagsRepository 创建标签Repository
func NewPostgresTagsRepository(db *sql.DB) *PostgresTagsRepository {
	return &PostgresTagsRepository{db: db}
}

// 确保实现了接口
var _ TagsRepository = (*PostgresTagsRepository)(nil)

// GetTag 根据tag_id获取tag
func (r *PostgresTagsRepository) GetTag(ctx context.Context, tenantID, tagID string) (*domain.Tag, error) {
	if tenantID == "" || tagID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			tag_id::text,
			tenant_id::text,
			tag_type,
			tag_name
		FROM tags_catalog
		WHERE tenant_id = $1 AND tag_id = $2
	`

	var tag domain.Tag
	err := r.db.QueryRowContext(ctx, query, tenantID, tagID).Scan(
		&tag.TagID,
		&tag.TenantID,
		&tag.TagType,
		&tag.TagName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return &tag, nil
}

// GetTagByName 根据tag_name获取tag
func (r *PostgresTagsRepository) GetTagByName(ctx context.Context, tenantID, tagName string) (*domain.Tag, error) {
	if tenantID == "" || tagName == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			tag_id::text,
			tenant_id::text,
			tag_type,
			tag_name
		FROM tags_catalog
		WHERE tenant_id = $1 AND tag_name = $2
	`

	var tag domain.Tag
	err := r.db.QueryRowContext(ctx, query, tenantID, tagName).Scan(
		&tag.TagID,
		&tag.TenantID,
		&tag.TagType,
		&tag.TagName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return &tag, nil
}

// ListTags 查询tags列表
func (r *PostgresTagsRepository) ListTags(ctx context.Context, tenantID string, filter TagsFilter, page, size int) ([]*domain.Tag, int, error) {
	if tenantID == "" {
		return nil, 0, fmt.Errorf("tenant_id is required")
	}

	// 构建WHERE条件
	where := []string{"tenant_id = $1"}
	args := []any{tenantID}
	argIdx := 2

	if filter.TagType != "" {
		where = append(where, fmt.Sprintf("tag_type = $%d", argIdx))
		args = append(args, filter.TagType)
		argIdx++
	} else if !filter.IncludeSystemTags {
		// 排除系统预定义tag类型
		where = append(where, fmt.Sprintf("tag_type NOT IN ($%d, $%d, $%d)", argIdx, argIdx+1, argIdx+2))
		args = append(args, "branch_tag", "family_tag", "area_tag")
		argIdx += 3
	}

	if filter.TagName != "" {
		where = append(where, fmt.Sprintf("tag_name ILIKE $%d", argIdx))
		args = append(args, "%"+filter.TagName+"%")
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM tags_catalog WHERE %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tags: %w", err)
	}

	// 查询列表（带分页）
	query := fmt.Sprintf(`
		SELECT 
			tag_id::text,
			tenant_id::text,
			tag_type,
			tag_name
		FROM tags_catalog
		WHERE %s
		ORDER BY tag_type, tag_name
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, size, (page-1)*size)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	tags := []*domain.Tag{}
	for rows.Next() {
		var tag domain.Tag
		err := rows.Scan(
			&tag.TagID,
			&tag.TenantID,
			&tag.TagType,
			&tag.TagName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, total, nil
}

// GetTagsForTenant 已废弃：使用 ListTags 替代
// 保留此方法以保持向后兼容，但建议使用 ListTags
// Deprecated: Use ListTags instead
func (r *PostgresTagsRepository) GetTagsForTenant(ctx context.Context, tenantID string, tagType *string) ([]*domain.Tag, error) {
	filter := TagsFilter{}
	if tagType != nil && *tagType != "" {
		filter.TagType = *tagType
	}
	filter.IncludeSystemTags = true
	
	tags, _, err := r.ListTags(ctx, tenantID, filter, 1, 10000) // 使用大size获取所有
	return tags, err
}

// CreateTag 创建tag（调用upsert_tag_to_catalog函数）
// 注意：
//   - 只允许创建 family_tag 和 user_tag
//   - branch_tag 和 area_tag 是系统自动维护的（在创建unit时自动调用upsert_tag_to_catalog）
//   - tags_catalog只存储tag_name，不存储member（member存储在源表中）
func (r *PostgresTagsRepository) CreateTag(ctx context.Context, tenantID string, tag *domain.Tag) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if tag == nil {
		return "", fmt.Errorf("tag is required")
	}
	if tag.TagName == "" {
		return "", fmt.Errorf("tag_name is required")
	}

	// 如果tag_type为空，默认设置为user_tag
	tagType := tag.TagType
	if tagType == "" {
		tagType = "user_tag"
	}

	// 只允许创建 family_tag 和 user_tag
	// branch_tag 和 area_tag 是系统自动维护的，不应该通过CreateTag手动创建
	allowedCreateTypes := map[string]bool{
		"family_tag": true,
		"user_tag":   true,
	}
	if !allowedCreateTypes[tagType] {
		return "", fmt.Errorf("cannot create tag_type '%s' via CreateTag. Only family_tag and user_tag can be created manually. branch_tag and area_tag are automatically maintained by the system", tagType)
	}

	// 调用upsert_tag_to_catalog函数
	var tagID string
	err := r.db.QueryRowContext(ctx,
		`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)::text`,
		tenantID, tag.TagName, tagType,
	).Scan(&tagID)
	if err != nil {
		return "", fmt.Errorf("failed to create tag: %w", err)
	}

	return tagID, nil
}

// UpdateTag 更新tag（调用upsert_tag_to_catalog函数）
// 注意：
//   - 只能更新tag_type（tag_name不能修改，因为tag_id基于tag_name生成）
//   - 权限控制由Service层处理（SystemAdmin可以修改，其他用户不能修改）
func (r *PostgresTagsRepository) UpdateTag(ctx context.Context, tenantID, tagName string, tag *domain.Tag) error {
	if tenantID == "" || tagName == "" {
		return fmt.Errorf("tenant_id and tag_name are required")
	}
	if tag == nil {
		return fmt.Errorf("tag is required")
	}
	if tag.TagType == "" {
		return fmt.Errorf("tag_type is required for update")
	}

	// 直接使用UPDATE语句更新tag_type
	// 注意：tag_name不能修改（因为tag_id基于tag_name生成），只能更新tag_type
	result, err := r.db.ExecContext(ctx,
		`UPDATE tags_catalog 
		 SET tag_type = $3
		 WHERE tenant_id = $1 AND tag_name = $2`,
		tenantID, tagName, tag.TagType,
	)
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	// 检查是否更新了记录
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found: tag_name '%s' does not exist", tagName)
	}

	return nil
}

// DeleteTag 删除tag（调用drop_tag函数）
// 注意：只允许删除family_tag和user_tag，branch_tag和area_tag不能删除
func (r *PostgresTagsRepository) DeleteTag(ctx context.Context, tenantID, tagName string) error {
	if tenantID == "" || tagName == "" {
		return fmt.Errorf("tenant_id and tag_name are required")
	}

	// 先查询tag_type，只允许删除family_tag和user_tag
	var tagType string
	err := r.db.QueryRowContext(ctx,
		`SELECT tag_type FROM tags_catalog WHERE tenant_id = $1 AND tag_name = $2`,
		tenantID, tagName,
	).Scan(&tagType)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag not found: %w", err)
		}
		return fmt.Errorf("failed to get tag type: %w", err)
	}

	// 只允许删除family_tag和user_tag
	allowedDeleteTypes := map[string]bool{
		"family_tag": true,
		"user_tag":   true,
	}
	if !allowedDeleteTypes[tagType] {
		return fmt.Errorf("cannot delete tag_type '%s'. Only family_tag and user_tag can be deleted. branch_tag and area_tag are system-maintained and cannot be deleted", tagType)
	}

	// 调用drop_tag函数
	var result bool
	err = r.db.QueryRowContext(ctx,
		`SELECT drop_tag($1::uuid, $2)`,
		tenantID, tagName,
	).Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// UpdateTagName 更新标签名称
func (r *PostgresTagsRepository) UpdateTagName(ctx context.Context, tenantID, tagID, newTagName string) error {
	if tenantID == "" || tagID == "" || newTagName == "" {
		return fmt.Errorf("tenant_id, tag_id, and new_tag_name are required")
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE tags_catalog SET tag_name = $3 WHERE tenant_id = $1 AND tag_id = $2`,
		tenantID, tagID, newTagName)
	if err != nil {
		return fmt.Errorf("failed to update tag name: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

// DeleteTagType 删除标签类型（调用drop_tag_type函数）
func (r *PostgresTagsRepository) DeleteTagType(ctx context.Context, tenantID, tagType string) (int, error) {
	if tenantID == "" || tagType == "" {
		return 0, fmt.Errorf("tenant_id and tag_type are required")
	}

	var updatedCount int
	err := r.db.QueryRowContext(ctx,
		`SELECT drop_tag_type($1::uuid, $2)`,
		tenantID, tagType).Scan(&updatedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to delete tag type: %w", err)
	}

	return updatedCount, nil
}

// AddTagObject 添加标签对象
// 注意：update_tag_objects函数可能已删除，这里暂时保留接口，实际实现需要检查
func (r *PostgresTagsRepository) AddTagObject(ctx context.Context, tagID, objectType, objectID, objectName string) error {
	if tagID == "" || objectType == "" || objectID == "" {
		return fmt.Errorf("tag_id, object_type, and object_id are required")
	}

	// 注意：update_tag_objects函数可能已删除，这里暂时保留调用
	// 如果函数不存在，会返回错误
	_, err := r.db.ExecContext(ctx,
		`SELECT update_tag_objects($1::uuid, $2, $3::uuid, $4, 'add')`,
		tagID, objectType, objectID, objectName)
	if err != nil {
		// 如果函数不存在，返回更友好的错误信息
		if strings.Contains(err.Error(), "does not exist") {
			return fmt.Errorf("update_tag_objects function is not available. Tag objects management may need to be redesigned")
		}
		return fmt.Errorf("failed to add tag object: %w", err)
	}

	return nil
}

// RemoveTagObject 删除标签对象
func (r *PostgresTagsRepository) RemoveTagObject(ctx context.Context, tagID, objectType, objectID string) error {
	if tagID == "" || objectType == "" || objectID == "" {
		return fmt.Errorf("tag_id, object_type, and object_id are required")
	}

	// 注意：update_tag_objects函数可能已删除
	_, err := r.db.ExecContext(ctx,
		`SELECT update_tag_objects($1::uuid, $2, $3::uuid, '', 'remove')`,
		tagID, objectType, objectID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return fmt.Errorf("update_tag_objects function is not available. Tag objects management may need to be redesigned")
		}
		return fmt.Errorf("failed to remove tag object: %w", err)
	}

	return nil
}

// SyncUserTag 同步用户标签到users.tags JSONB
func (r *PostgresTagsRepository) SyncUserTag(ctx context.Context, tagName, userID string, add bool) error {
	if tagName == "" || userID == "" {
		return fmt.Errorf("tag_name and user_id are required")
	}

	if add {
		_, err := r.db.ExecContext(ctx,
			`UPDATE users 
			 SET tags = COALESCE(tags, '[]'::jsonb) || jsonb_build_array($1::text)
			 WHERE user_id = $2::uuid
			   AND (tags IS NULL OR NOT (tags ? $1))`,
			tagName, userID)
		if err != nil {
			return fmt.Errorf("failed to sync tag to user's tags: %w", err)
		}
	} else {
		_, err := r.db.ExecContext(ctx,
			`UPDATE users 
			 SET tags = tags - $1
			 WHERE user_id = $2::uuid
			   AND tags IS NOT NULL
			   AND tags ? $1`,
			tagName, userID)
		if err != nil {
			return fmt.Errorf("failed to remove tag from user's tags: %w", err)
		}
	}

	return nil
}

// SyncResidentFamilyTag 同步住户家庭标签
func (r *PostgresTagsRepository) SyncResidentFamilyTag(ctx context.Context, tagName, residentID string, clear bool) error {
	if residentID == "" {
		return fmt.Errorf("resident_id is required")
	}

	if clear {
		_, err := r.db.ExecContext(ctx,
			`UPDATE residents 
			 SET family_tag = NULL
			 WHERE resident_id = $1::uuid
			   AND family_tag = $2`,
			residentID, tagName)
		if err != nil {
			return fmt.Errorf("failed to clear family_tag from resident: %w", err)
		}
	}

	return nil
}

