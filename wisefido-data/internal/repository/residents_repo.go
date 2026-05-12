package repository

import (
	"context"
	"wisefido-data/internal/domain"
)

// ResidentsRepository 住户Repository接口
// 使用强类型领域模型，不使用map[string]any
// 设计原则：从底层（数据库）向上设计，Repository层只负责数据访问
type ResidentsRepository interface {
	// ========== Residents 表操作 ==========
	// 查询接口
	GetResident(ctx context.Context, tenantID, residentID string) (*domain.ResidentDetail, error)
	GetResidentByAccount(ctx context.Context, tenantID string, accountHash []byte) (*domain.Resident, error)
	GetResidentByEmail(ctx context.Context, tenantID string, emailHash []byte) (*domain.Resident, error)
	GetResidentByPhone(ctx context.Context, tenantID string, phoneHash []byte) (*domain.Resident, error)
	ListResidents(ctx context.Context, tenantID string, filters ResidentFilters, page, size int) ([]*domain.Resident, int, error)

	// 创建接口（替代触发器：trigger_sync_family_tag）
	// 支持 v1: *domain.Resident 或 v2: *domain.ResidentCreateInput
	CreateResident(ctx context.Context, tenantID string, resident interface{}) (string, error)

	// Deprecated: Use UpdateResidentFields instead.
	UpdateResident(ctx context.Context, tenantID, residentID string, resident *domain.Resident) error
	// UpdateResidentFields 更新住户信息（使用更新模型）
	// 支持区分"不更新"、"更新"、"删除"三种状态
	UpdateResidentFields(ctx context.Context, tenantID, residentID string, update *domain.ResidentUpdate) error

	// 删除接口
	DeleteResident(ctx context.Context, tenantID, residentID string) error

	// 位置绑定接口
	BindResidentToLocation(ctx context.Context, tenantID, residentID string, unitID, roomID, bedID *string) error

	// ========== ResidentPHI 表操作 ==========
	GetResidentPHI(ctx context.Context, tenantID, residentID string) (*domain.ResidentPHI, error)
	// Deprecated: 使用 UpsertResidentPHIFields 替代，支持区分"不更新"、"更新"、"删除"三种状态
	UpsertResidentPHI(ctx context.Context, tenantID, residentID string, phi *domain.ResidentPHI) error
	// UpsertResidentPHIFields 创建或更新住户PHI信息（使用更新模型）
	// 支持区分"不更新"、"更新"、"删除"三种状态
	// 注意：使用 UPSERT 语义（UNIQUE(tenant_id, resident_id)）
	UpsertResidentPHIFields(ctx context.Context, tenantID, residentID string, update *domain.ResidentPHIUpdate) error

	// ========== ResidentContacts 表操作 ==========
	GetResidentContacts(ctx context.Context, tenantID, residentID string) ([]*domain.ResidentContact, error)
	CreateResidentContact(ctx context.Context, tenantID, residentID string, contact *domain.ResidentContact) (string, error)
	// Deprecated: 使用 UpdateResidentContactFields 替代，支持区分"不更新"、"更新"、"删除"三种状态
	// 注意：主键是 (resident_id, slot)，不是 contact_id
	UpdateResidentContact(ctx context.Context, tenantID, residentID, slot string, contact *domain.ResidentContact) error
	// UpdateResidentContactFields 更新联系人（使用更新模型）
	// 支持区分"不更新"、"更新"、"删除"三种状态
	// 注意：主键是 (resident_id, slot)
	UpdateResidentContactFields(ctx context.Context, tenantID, residentID, slot string, update *domain.ResidentContactUpdate) error
	// Deprecated: 主键是 (resident_id, slot)，不是 contact_id
	DeleteResidentContact(ctx context.Context, tenantID, contactID string) error
	// DeleteResidentContactBySlot 删除联系人（使用主键 resident_id 和 slot）
	DeleteResidentContactBySlot(ctx context.Context, tenantID, residentID, slot string) error

	// ========== ResidentCaregivers 表操作 ==========
	GetResidentCaregivers(ctx context.Context, tenantID, residentID string) ([]*domain.ResidentCaregiver, error)
	// Deprecated: 使用 UpsertResidentCaregiverFields 替代，支持区分"不更新"、"更新"、"删除"三种状态
	UpsertResidentCaregiver(ctx context.Context, tenantID, residentID string, caregiver *domain.ResidentCaregiver) error
	// UpsertResidentCaregiverFields 创建或更新护理人员关联（使用更新模型）
	// 支持区分"不更新"、"更新"、"删除"三种状态
	// 注意：使用 UPSERT 语义（UNIQUE(tenant_id, resident_id)）
	UpsertResidentCaregiverFields(ctx context.Context, tenantID, residentID string, update *domain.ResidentCaregiverUpdate) error
}

// ResidentFilters 住户查询过滤器
type ResidentFilters struct {
	// 基本过滤
	Status       string // 按status过滤
	ServiceLevel string // 按service_level过滤
	UnitID       string // 按unit_id过滤
	RoomID       string // 按room_id过滤
	BedID        string // 按bed_id过滤
	BranchID     string // 按branch_id过滤

	// 搜索（支持account, email_hash, phone_hash, nickname, unit_name, first_name）
	Search string // 模糊搜索：支持resident_account, nickname, first_name (在resident_phi表中)

	// 权限过滤
	AssignedUserID string // 仅查询分配给该用户的住户
	BranchTag      string // 仅查询该分支的住户
}
