package domain

// Tag 标签领域模型（对应 tags_catalog 表）
// 基于实际DB表结构：4个字段（tag_objects已删除）
type Tag struct {
	TagID    string `db:"tag_id"`
	TenantID string `db:"tenant_id"`
	TagType  string `db:"tag_type"` // branch_tag, family_tag, area_tag, user_tag
	TagName  string `db:"tag_name"`
}

// TagType 标签类型常量
type TagType string

const (
	TagTypeBranchTag TagType = "branch_tag" // 系统预定义
	TagTypeFamilyTag TagType = "family_tag" // 系统预定义
	TagTypeAreaTag   TagType = "area_tag"   // 系统预定义
	TagTypeUserTag   TagType = "user_tag"   // 系统定义（租户新建）
)

// IsSystemTagType 判断是否为系统预定义tag类型
func (t TagType) IsSystemTagType() bool {
	return t == TagTypeBranchTag || t == TagTypeFamilyTag || t == TagTypeAreaTag
}

