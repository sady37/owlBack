package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresResidentsRepository 住户Repository实现（强类型版本）
// 实现ResidentsRepository接口，使用domain领域模型
type PostgresResidentsRepository struct {
	db *sql.DB
}

// NewPostgresResidentsRepository 创建住户Repository
func NewPostgresResidentsRepository(db *sql.DB) *PostgresResidentsRepository {
	return &PostgresResidentsRepository{db: db}
}

// 确保实现了接口
var _ ResidentsRepository = (*PostgresResidentsRepository)(nil)

// ============================================
// Residents 表操作
// ============================================

// GetResident 根据resident_id获取住户信息
func (r *PostgresResidentsRepository) GetResident(ctx context.Context, tenantID, residentID string) (*domain.Resident, error) {
	if tenantID == "" || residentID == "" {
		return nil, fmt.Errorf("tenant_id and resident_id are required")
	}

	query := `
		SELECT 
			resident_id::text,
			tenant_id::text,
			resident_account,
			resident_account_hash,
			nickname,
			admission_date,
			discharge_date,
			service_level,
			status,
			role,
			COALESCE(metadata, '{}'::jsonb)::text as metadata,
			COALESCE(note, '') as note,
			phone_hash,
			email_hash,
			password_hash,
			family_tag,
			can_view_status,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND resident_id = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, familyTag, unitID, roomID, bedID sql.NullString
	var metadataRaw sql.NullString
	var phoneHash, emailHash, passwordHash sql.Null[[]byte]

	err := r.db.QueryRowContext(ctx, query, tenantID, residentID).Scan(
		&resident.ResidentID,
		&resident.TenantID,
		&resident.ResidentAccount,
		&resident.ResidentAccountHash,
		&resident.Nickname,
		&admissionDate,
		&dischargeDate,
		&serviceLevel,
		&resident.Status,
		&resident.Role,
		&metadataRaw,
		&note,
		&phoneHash,
		&emailHash,
		&passwordHash,
		&familyTag,
		&resident.CanViewStatus,
		&unitID,
		&roomID,
		&bedID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}

	// 处理可空字段
	if admissionDate.Valid {
		resident.AdmissionDate = &admissionDate.Time
	}
	if dischargeDate.Valid {
		resident.DischargeDate = &dischargeDate.Time
	}
	if serviceLevel.Valid {
		resident.ServiceLevel = serviceLevel.String
	}
	if note.Valid {
		resident.Note = note.String
	}
	if familyTag.Valid {
		resident.FamilyTag = familyTag.String
	}
	if unitID.Valid {
		resident.UnitID = unitID.String
	}
	if roomID.Valid {
		resident.RoomID = roomID.String
	}
	if bedID.Valid {
		resident.BedID = bedID.String
	}
	if phoneHash.Valid {
		resident.PhoneHash = phoneHash.V
	}
	if emailHash.Valid {
		resident.EmailHash = emailHash.V
	}
	if passwordHash.Valid {
		resident.PasswordHash = passwordHash.V
	}
	if metadataRaw.Valid && metadataRaw.String != "" {
		resident.Metadata = json.RawMessage(metadataRaw.String)
	}

	return &resident, nil
}

// GetResidentByAccount 根据account_hash获取住户（用于登录）
func (r *PostgresResidentsRepository) GetResidentByAccount(ctx context.Context, tenantID string, accountHash []byte) (*domain.Resident, error) {
	if tenantID == "" || len(accountHash) == 0 {
		return nil, fmt.Errorf("tenant_id and account_hash are required")
	}

	query := `
		SELECT 
			resident_id::text,
			tenant_id::text,
			resident_account,
			resident_account_hash,
			nickname,
			admission_date,
			discharge_date,
			service_level,
			status,
			role,
			COALESCE(metadata, '{}'::jsonb)::text as metadata,
			COALESCE(note, '') as note,
			phone_hash,
			email_hash,
			password_hash,
			family_tag,
			can_view_status,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND resident_account_hash = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, familyTag, unitID, roomID, bedID sql.NullString
	var metadataRaw sql.NullString
	var phoneHash, emailHash, passwordHash sql.Null[[]byte]

	err := r.db.QueryRowContext(ctx, query, tenantID, accountHash).Scan(
		&resident.ResidentID,
		&resident.TenantID,
		&resident.ResidentAccount,
		&resident.ResidentAccountHash,
		&resident.Nickname,
		&admissionDate,
		&dischargeDate,
		&serviceLevel,
		&resident.Status,
		&resident.Role,
		&metadataRaw,
		&note,
		&phoneHash,
		&emailHash,
		&passwordHash,
		&familyTag,
		&resident.CanViewStatus,
		&unitID,
		&roomID,
		&bedID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get resident by account: %w", err)
	}

	// 处理可空字段（与GetResident相同）
	if admissionDate.Valid {
		resident.AdmissionDate = &admissionDate.Time
	}
	if dischargeDate.Valid {
		resident.DischargeDate = &dischargeDate.Time
	}
	if serviceLevel.Valid {
		resident.ServiceLevel = serviceLevel.String
	}
	if note.Valid {
		resident.Note = note.String
	}
	if familyTag.Valid {
		resident.FamilyTag = familyTag.String
	}
	if unitID.Valid {
		resident.UnitID = unitID.String
	}
	if roomID.Valid {
		resident.RoomID = roomID.String
	}
	if bedID.Valid {
		resident.BedID = bedID.String
	}
	if phoneHash.Valid {
		resident.PhoneHash = phoneHash.V
	}
	if emailHash.Valid {
		resident.EmailHash = emailHash.V
	}
	if passwordHash.Valid {
		resident.PasswordHash = passwordHash.V
	}
	if metadataRaw.Valid && metadataRaw.String != "" {
		resident.Metadata = json.RawMessage(metadataRaw.String)
	}

	return &resident, nil
}

// GetResidentByEmail 根据email_hash获取住户（用于登录）
func (r *PostgresResidentsRepository) GetResidentByEmail(ctx context.Context, tenantID string, emailHash []byte) (*domain.Resident, error) {
	if tenantID == "" || len(emailHash) == 0 {
		return nil, fmt.Errorf("tenant_id and email_hash are required")
	}

	// 使用相同的查询结构，但WHERE条件改为email_hash
	query := `
		SELECT 
			resident_id::text,
			tenant_id::text,
			resident_account,
			resident_account_hash,
			nickname,
			admission_date,
			discharge_date,
			service_level,
			status,
			role,
			COALESCE(metadata, '{}'::jsonb)::text as metadata,
			COALESCE(note, '') as note,
			phone_hash,
			email_hash,
			password_hash,
			family_tag,
			can_view_status,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND email_hash = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, familyTag, unitID, roomID, bedID sql.NullString
	var metadataRaw sql.NullString
	var phoneHashVal, emailHashVal, passwordHash sql.Null[[]byte]

	err := r.db.QueryRowContext(ctx, query, tenantID, emailHash).Scan(
		&resident.ResidentID,
		&resident.TenantID,
		&resident.ResidentAccount,
		&resident.ResidentAccountHash,
		&resident.Nickname,
		&admissionDate,
		&dischargeDate,
		&serviceLevel,
		&resident.Status,
		&resident.Role,
		&metadataRaw,
		&note,
		&phoneHashVal,
		&emailHashVal,
		&passwordHash,
		&familyTag,
		&resident.CanViewStatus,
		&unitID,
		&roomID,
		&bedID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get resident by email: %w", err)
	}

	// 处理可空字段
	if admissionDate.Valid {
		resident.AdmissionDate = &admissionDate.Time
	}
	if dischargeDate.Valid {
		resident.DischargeDate = &dischargeDate.Time
	}
	if serviceLevel.Valid {
		resident.ServiceLevel = serviceLevel.String
	}
	if note.Valid {
		resident.Note = note.String
	}
	if familyTag.Valid {
		resident.FamilyTag = familyTag.String
	}
	if unitID.Valid {
		resident.UnitID = unitID.String
	}
	if roomID.Valid {
		resident.RoomID = roomID.String
	}
	if bedID.Valid {
		resident.BedID = bedID.String
	}
	if phoneHashVal.Valid {
		resident.PhoneHash = phoneHashVal.V
	}
	if emailHashVal.Valid {
		resident.EmailHash = emailHashVal.V
	}
	if passwordHash.Valid {
		resident.PasswordHash = passwordHash.V
	}
	if metadataRaw.Valid && metadataRaw.String != "" {
		resident.Metadata = json.RawMessage(metadataRaw.String)
	}

	return &resident, nil
}

// GetResidentByPhone 根据phone_hash获取住户（用于登录）
func (r *PostgresResidentsRepository) GetResidentByPhone(ctx context.Context, tenantID string, phoneHash []byte) (*domain.Resident, error) {
	if tenantID == "" || len(phoneHash) == 0 {
		return nil, fmt.Errorf("tenant_id and phone_hash are required")
	}

	// 使用相同的查询结构，但WHERE条件改为phone_hash
	query := `
		SELECT 
			resident_id::text,
			tenant_id::text,
			resident_account,
			resident_account_hash,
			nickname,
			admission_date,
			discharge_date,
			service_level,
			status,
			role,
			COALESCE(metadata, '{}'::jsonb)::text as metadata,
			COALESCE(note, '') as note,
			phone_hash,
			email_hash,
			password_hash,
			family_tag,
			can_view_status,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND phone_hash = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, familyTag, unitID, roomID, bedID sql.NullString
	var metadataRaw sql.NullString
	var phoneHashVal, emailHashVal, passwordHash sql.Null[[]byte]

	err := r.db.QueryRowContext(ctx, query, tenantID, phoneHash).Scan(
		&resident.ResidentID,
		&resident.TenantID,
		&resident.ResidentAccount,
		&resident.ResidentAccountHash,
		&resident.Nickname,
		&admissionDate,
		&dischargeDate,
		&serviceLevel,
		&resident.Status,
		&resident.Role,
		&metadataRaw,
		&note,
		&phoneHashVal,
		&emailHashVal,
		&passwordHash,
		&familyTag,
		&resident.CanViewStatus,
		&unitID,
		&roomID,
		&bedID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get resident by phone: %w", err)
	}

	// 处理可空字段
	if admissionDate.Valid {
		resident.AdmissionDate = &admissionDate.Time
	}
	if dischargeDate.Valid {
		resident.DischargeDate = &dischargeDate.Time
	}
	if serviceLevel.Valid {
		resident.ServiceLevel = serviceLevel.String
	}
	if note.Valid {
		resident.Note = note.String
	}
	if familyTag.Valid {
		resident.FamilyTag = familyTag.String
	}
	if unitID.Valid {
		resident.UnitID = unitID.String
	}
	if roomID.Valid {
		resident.RoomID = roomID.String
	}
	if bedID.Valid {
		resident.BedID = bedID.String
	}
	if phoneHashVal.Valid {
		resident.PhoneHash = phoneHashVal.V
	}
	if emailHashVal.Valid {
		resident.EmailHash = emailHashVal.V
	}
	if passwordHash.Valid {
		resident.PasswordHash = passwordHash.V
	}
	if metadataRaw.Valid && metadataRaw.String != "" {
		resident.Metadata = json.RawMessage(metadataRaw.String)
	}

	return &resident, nil
}

// ListResidents 查询住户列表（支持分页、过滤、搜索）
func (r *PostgresResidentsRepository) ListResidents(ctx context.Context, tenantID string, filters ResidentFilters, page, size int) ([]*domain.Resident, int, error) {
	if tenantID == "" {
		return []*domain.Resident{}, 0, nil
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	// 构建WHERE条件
	where := []string{"r.tenant_id = $1"}
	args := []any{tenantID}
	argIdx := 2

	// 基本过滤
	if filters.Status != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.ServiceLevel != "" {
		where = append(where, fmt.Sprintf("r.service_level = $%d", argIdx))
		args = append(args, filters.ServiceLevel)
		argIdx++
	}
	if filters.FamilyTag != "" {
		where = append(where, fmt.Sprintf("r.family_tag = $%d", argIdx))
		args = append(args, filters.FamilyTag)
		argIdx++
	}
	if filters.UnitID != "" {
		where = append(where, fmt.Sprintf("r.unit_id = $%d", argIdx))
		args = append(args, filters.UnitID)
		argIdx++
	}
	if filters.RoomID != "" {
		where = append(where, fmt.Sprintf("r.room_id = $%d", argIdx))
		args = append(args, filters.RoomID)
		argIdx++
	}
	if filters.BedID != "" {
		where = append(where, fmt.Sprintf("r.bed_id = $%d", argIdx))
		args = append(args, filters.BedID)
		argIdx++
	}

	// 搜索功能：支持resident_account, nickname, first_name (在resident_phi表中)
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		where = append(where, fmt.Sprintf("(r.resident_account ILIKE $%d OR r.nickname ILIKE $%d OR EXISTS (SELECT 1 FROM resident_phi rp WHERE rp.resident_id = r.resident_id AND rp.first_name ILIKE $%d))", argIdx, argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	// 权限过滤：assigned_user_id, branch_tag
	// 注意：这些过滤需要JOIN其他表，暂时不实现，由Service层处理

	whereClause := strings.Join(where, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM residents r WHERE %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count residents: %w", err)
	}

	// 查询列表（带分页）
	query := fmt.Sprintf(`
		SELECT 
			r.resident_id::text,
			r.tenant_id::text,
			r.resident_account,
			r.resident_account_hash,
			r.nickname,
			r.admission_date,
			r.discharge_date,
			r.service_level,
			r.status,
			r.role,
			COALESCE(r.metadata, '{}'::jsonb)::text as metadata,
			COALESCE(r.note, '') as note,
			r.phone_hash,
			r.email_hash,
			r.password_hash,
			r.family_tag,
			r.can_view_status,
			r.unit_id::text,
			r.room_id::text,
			r.bed_id::text
		FROM residents r
		WHERE %s
		ORDER BY r.nickname
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list residents: %w", err)
	}
	defer rows.Close()

	residents := []*domain.Resident{}
	for rows.Next() {
		var resident domain.Resident
		var admissionDate, dischargeDate sql.NullTime
		var serviceLevel, note, familyTag, unitID, roomID, bedID sql.NullString
		var metadataRaw sql.NullString
		var phoneHash, emailHash, passwordHash sql.Null[[]byte]

		err := rows.Scan(
			&resident.ResidentID,
			&resident.TenantID,
			&resident.ResidentAccount,
			&resident.ResidentAccountHash,
			&resident.Nickname,
			&admissionDate,
			&dischargeDate,
			&serviceLevel,
			&resident.Status,
			&resident.Role,
			&metadataRaw,
			&note,
			&phoneHash,
			&emailHash,
			&passwordHash,
			&familyTag,
			&resident.CanViewStatus,
			&unitID,
			&roomID,
			&bedID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan resident: %w", err)
		}

		// 处理可空字段
		if admissionDate.Valid {
			resident.AdmissionDate = &admissionDate.Time
		}
		if dischargeDate.Valid {
			resident.DischargeDate = &dischargeDate.Time
		}
		if serviceLevel.Valid {
			resident.ServiceLevel = serviceLevel.String
		}
		if note.Valid {
			resident.Note = note.String
		}
		if familyTag.Valid {
			resident.FamilyTag = familyTag.String
		}
		if unitID.Valid {
			resident.UnitID = unitID.String
		}
		if roomID.Valid {
			resident.RoomID = roomID.String
		}
		if bedID.Valid {
			resident.BedID = bedID.String
		}
		if phoneHash.Valid {
			resident.PhoneHash = phoneHash.V
		}
		if emailHash.Valid {
			resident.EmailHash = emailHash.V
		}
		if passwordHash.Valid {
			resident.PasswordHash = passwordHash.V
		}
		if metadataRaw.Valid && metadataRaw.String != "" {
			resident.Metadata = json.RawMessage(metadataRaw.String)
		}

		residents = append(residents, &resident)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate residents: %w", err)
	}

	return residents, total, nil
}


// CreateResident 创建新住户
// 触发器替代：同步family_tag到tags_catalog（调用upsert_tag_to_catalog）
func (r *PostgresResidentsRepository) CreateResident(ctx context.Context, tenantID string, resident *domain.Resident) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if resident == nil {
		return "", fmt.Errorf("resident is required")
	}
	if resident.ResidentAccount == "" {
		return "", fmt.Errorf("resident_account is required")
	}
	if resident.Nickname == "" {
		return "", fmt.Errorf("nickname is required")
	}
	if len(resident.ResidentAccountHash) == 0 {
		return "", fmt.Errorf("resident_account_hash is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 处理默认值
	status := resident.Status
	if status == "" {
		status = "active"
	}
	role := resident.Role
	if role == "" {
		role = "Resident"
	}

	// 处理admission_date
	admissionDate := time.Now()
	if resident.AdmissionDate != nil {
		admissionDate = *resident.AdmissionDate
	}

	// 处理可空字段
	var dischargeDateArg any = nil
	if resident.DischargeDate != nil {
		dischargeDateArg = *resident.DischargeDate
	}
	var serviceLevelArg any = nil
	if resident.ServiceLevel != "" {
		serviceLevelArg = resident.ServiceLevel
	}
	var unitIDArg any = nil
	if resident.UnitID != "" {
		unitIDArg = resident.UnitID
	}
	var roomIDArg any = nil
	if resident.RoomID != "" {
		roomIDArg = resident.RoomID
	}
	var bedIDArg any = nil
	if resident.BedID != "" {
		bedIDArg = resident.BedID
	}
	var familyTagArg any = nil
	if resident.FamilyTag != "" {
		familyTagArg = resident.FamilyTag
	}
	var noteArg any = nil
	if resident.Note != "" {
		noteArg = resident.Note
	}
	var phoneHashArg any = nil
	if len(resident.PhoneHash) > 0 {
		phoneHashArg = resident.PhoneHash
	}
	var emailHashArg any = nil
	if len(resident.EmailHash) > 0 {
		emailHashArg = resident.EmailHash
	}
	var passwordHashArg any = nil
	if len(resident.PasswordHash) > 0 {
		passwordHashArg = resident.PasswordHash
	}
	var metadataArg any = nil
	if len(resident.Metadata) > 0 {
		metadataArg = string(resident.Metadata)
	}

	// 插入住户
	var residentID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO residents (
			tenant_id, resident_account, resident_account_hash, nickname,
			admission_date, discharge_date, service_level, status, role,
			metadata, note, phone_hash, email_hash, password_hash,
			family_tag, can_view_status, unit_id, room_id, bed_id
		) VALUES ($1, LOWER($2), $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING resident_id::text`,
		tenantID, resident.ResidentAccount, resident.ResidentAccountHash, resident.Nickname,
		admissionDate, dischargeDateArg, serviceLevelArg, status, role,
		metadataArg, noteArg, phoneHashArg, emailHashArg, passwordHashArg,
		familyTagArg, resident.CanViewStatus, unitIDArg, roomIDArg, bedIDArg,
	).Scan(&residentID)
	if err != nil {
		return "", fmt.Errorf("failed to create resident: %w", err)
	}

	// 同步family_tag到tags_catalog（替代trigger_sync_family_tag）
	if resident.FamilyTag != "" {
		_, err = tx.ExecContext(ctx,
			`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
			tenantID, resident.FamilyTag, "family_tag",
		)
		if err != nil {
			return "", fmt.Errorf("failed to sync family_tag to catalog: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return residentID, nil
}

// UpdateResident 更新住户信息
// 触发器替代：同步family_tag到tags_catalog（调用upsert_tag_to_catalog）
func (r *PostgresResidentsRepository) UpdateResident(ctx context.Context, tenantID, residentID string, resident *domain.Resident) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}
	if resident == nil {
		return fmt.Errorf("resident is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取旧数据（用于比较family_tag变化）
	var oldFamilyTag sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT family_tag FROM residents WHERE tenant_id = $1 AND resident_id = $2`,
		tenantID, residentID,
	).Scan(&oldFamilyTag)
	if err != nil {
		return fmt.Errorf("failed to get old resident data: %w", err)
	}

	// 构建UPDATE语句
	updates := []string{}
	args := []any{tenantID, residentID}
	argIdx := 3

	if resident.ResidentAccount != "" {
		updates = append(updates, fmt.Sprintf("resident_account = LOWER($%d)", argIdx))
		args = append(args, resident.ResidentAccount)
		argIdx++
	}
	if len(resident.ResidentAccountHash) > 0 {
		updates = append(updates, fmt.Sprintf("resident_account_hash = $%d", argIdx))
		args = append(args, resident.ResidentAccountHash)
		argIdx++
	}
	if resident.Nickname != "" {
		updates = append(updates, fmt.Sprintf("nickname = $%d", argIdx))
		args = append(args, resident.Nickname)
		argIdx++
	}
	if resident.AdmissionDate != nil {
		updates = append(updates, fmt.Sprintf("admission_date = $%d", argIdx))
		args = append(args, *resident.AdmissionDate)
		argIdx++
	}
	if resident.DischargeDate != nil {
		updates = append(updates, fmt.Sprintf("discharge_date = $%d", argIdx))
		args = append(args, *resident.DischargeDate)
		argIdx++
	}
	if resident.ServiceLevel != "" {
		updates = append(updates, fmt.Sprintf("service_level = $%d", argIdx))
		args = append(args, resident.ServiceLevel)
		argIdx++
	} else if resident.ServiceLevel == "" {
		// 允许设置为空
		updates = append(updates, "service_level = NULL")
	}
	if resident.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, resident.Status)
		argIdx++
	}
	if resident.Role != "" {
		updates = append(updates, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, resident.Role)
		argIdx++
	}
	if resident.Note != "" {
		updates = append(updates, fmt.Sprintf("note = $%d", argIdx))
		args = append(args, resident.Note)
		argIdx++
	}
	if len(resident.PhoneHash) > 0 {
		updates = append(updates, fmt.Sprintf("phone_hash = $%d", argIdx))
		args = append(args, resident.PhoneHash)
		argIdx++
	}
	if len(resident.EmailHash) > 0 {
		updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
		args = append(args, resident.EmailHash)
		argIdx++
	}
	if len(resident.PasswordHash) > 0 {
		updates = append(updates, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, resident.PasswordHash)
		argIdx++
	}
	if resident.FamilyTag != "" {
		updates = append(updates, fmt.Sprintf("family_tag = $%d", argIdx))
		args = append(args, resident.FamilyTag)
		argIdx++
	}
	updates = append(updates, fmt.Sprintf("can_view_status = $%d", argIdx))
	args = append(args, resident.CanViewStatus)
	argIdx++
	if resident.UnitID != "" {
		updates = append(updates, fmt.Sprintf("unit_id = $%d", argIdx))
		args = append(args, resident.UnitID)
		argIdx++
	}
	if resident.RoomID != "" {
		updates = append(updates, fmt.Sprintf("room_id = $%d", argIdx))
		args = append(args, resident.RoomID)
		argIdx++
	} else {
		// 允许设置为NULL
		updates = append(updates, "room_id = NULL")
	}
	if resident.BedID != "" {
		updates = append(updates, fmt.Sprintf("bed_id = $%d", argIdx))
		args = append(args, resident.BedID)
		argIdx++
	} else {
		// 允许设置为NULL
		updates = append(updates, "bed_id = NULL")
	}
	if len(resident.Metadata) > 0 {
		updates = append(updates, fmt.Sprintf("metadata = $%d::jsonb", argIdx))
		args = append(args, string(resident.Metadata))
		argIdx++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE residents
		SET %s
		WHERE tenant_id = $1 AND resident_id = $2
	`, strings.Join(updates, ", "))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update resident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("resident not found: tenant_id '%s', resident_id '%s'", tenantID, residentID)
	}

	// 同步family_tag变化（替代trigger_sync_family_tag）
	oldFamilyTagValue := ""
	if oldFamilyTag.Valid {
		oldFamilyTagValue = oldFamilyTag.String
	}
	newFamilyTagValue := resident.FamilyTag

	if oldFamilyTagValue != newFamilyTagValue {
		// 如果新tag不为空，添加到目录
		if newFamilyTagValue != "" {
			_, err = tx.ExecContext(ctx,
				`SELECT upsert_tag_to_catalog($1::uuid, $2, $3)`,
				tenantID, newFamilyTagValue, "family_tag",
			)
			if err != nil {
				return fmt.Errorf("failed to sync new family_tag to catalog: %w", err)
			}
		}
		// 注意：不删除旧tag（用户确认：不删除family_tag）
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteResident 删除住户
// 注意：不删除family_tag（即使没有其他住户使用，也保留在tags_catalog中）
func (r *PostgresResidentsRepository) DeleteResident(ctx context.Context, tenantID, residentID string) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}

	_, err := r.db.ExecContext(ctx,
		`DELETE FROM residents WHERE tenant_id = $1 AND resident_id = $2`,
		tenantID, residentID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete resident: %w", err)
	}

	return nil
}

// BindResidentToLocation 绑定住户到位置（unit/room/bed）
// 支持解绑：可以传入空值（nil/空字符串）来解绑room_id或bed_id
func (r *PostgresResidentsRepository) BindResidentToLocation(ctx context.Context, tenantID, residentID string, unitID, roomID, bedID *string) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}
	if unitID == nil || *unitID == "" {
		return fmt.Errorf("unit_id is required (cannot be unbound)")
	}

	// 约束验证：如果指定bed_id，必须同时指定room_id
	if bedID != nil && *bedID != "" {
		if roomID == nil || *roomID == "" {
			return fmt.Errorf("room_id is required when bed_id is specified")
		}
	}

	// 构建UPDATE语句
	updates := []string{"unit_id = $3"}
	args := []any{tenantID, residentID, *unitID}
	argIdx := 4

	if roomID != nil && *roomID != "" {
		updates = append(updates, fmt.Sprintf("room_id = $%d", argIdx))
		args = append(args, *roomID)
		argIdx++
	} else {
		updates = append(updates, "room_id = NULL")
	}

	if bedID != nil && *bedID != "" {
		updates = append(updates, fmt.Sprintf("bed_id = $%d", argIdx))
		args = append(args, *bedID)
		argIdx++
	} else {
		updates = append(updates, "bed_id = NULL")
	}

	query := fmt.Sprintf(`
		UPDATE residents
		SET %s
		WHERE tenant_id = $1 AND resident_id = $2
	`, strings.Join(updates, ", "))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bind resident to location: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("resident not found: tenant_id '%s', resident_id '%s'", tenantID, residentID)
	}

	return nil
}


// ============================================
// ResidentPHI 表操作
// ============================================

// GetResidentPHI 获取住户PHI信息
func (r *PostgresResidentsRepository) GetResidentPHI(ctx context.Context, tenantID, residentID string) (*domain.ResidentPHI, error) {
	if tenantID == "" || residentID == "" {
		return nil, fmt.Errorf("tenant_id and resident_id are required")
	}

	query := `
		SELECT 
			phi_id::text,
			tenant_id::text,
			resident_id::text,
			first_name,
			last_name,
			gender,
			date_of_birth,
			resident_phone,
			resident_email,
			weight_lb,
			height_ft,
			height_in,
			mobility_level,
			tremor_status,
			mobility_aid,
			adl_assistance,
			comm_status,
			has_hypertension,
			has_hyperlipaemia,
			has_hyperglycaemia,
			has_stroke_history,
			has_paralysis,
			has_alzheimer,
			medical_history,
			home_address_street,
			home_address_city,
			home_address_state,
			home_address_postal_code,
			plus_code
		FROM resident_phi
		WHERE tenant_id = $1 AND resident_id = $2
	`

	var phi domain.ResidentPHI
	var firstName, lastName, gender, residentPhone, residentEmail sql.NullString
	var dateOfBirth sql.NullTime
	var weightLb, heightFt, heightIn sql.NullFloat64
	var mobilityLevel sql.NullInt64
	var tremorStatus, mobilityAid, adlAssistance, commStatus sql.NullString
	var hasHypertension, hasHyperlipaemia, hasHyperglycaemia, hasStrokeHistory, hasParalysis, hasAlzheimer sql.NullBool
	var medicalHistory sql.NullString
	var homeAddressStreet, homeAddressCity, homeAddressState, homeAddressPostalCode, plusCode sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, residentID).Scan(
		&phi.PhiID,
		&phi.TenantID,
		&phi.ResidentID,
		&firstName,
		&lastName,
		&gender,
		&dateOfBirth,
		&residentPhone,
		&residentEmail,
		&weightLb,
		&heightFt,
		&heightIn,
		&mobilityLevel,
		&tremorStatus,
		&mobilityAid,
		&adlAssistance,
		&commStatus,
		&hasHypertension,
		&hasHyperlipaemia,
		&hasHyperglycaemia,
		&hasStrokeHistory,
		&hasParalysis,
		&hasAlzheimer,
		&medicalHistory,
		&homeAddressStreet,
		&homeAddressCity,
		&homeAddressState,
		&homeAddressPostalCode,
		&plusCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident PHI not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get resident PHI: %w", err)
	}

	// 处理可空字段
	if firstName.Valid {
		phi.FirstName = firstName.String
	}
	if lastName.Valid {
		phi.LastName = lastName.String
	}
	if gender.Valid {
		phi.Gender = gender.String
	}
	if dateOfBirth.Valid {
		phi.DateOfBirth = &dateOfBirth.Time
	}
	if residentPhone.Valid {
		phi.ResidentPhone = residentPhone.String
	}
	if residentEmail.Valid {
		phi.ResidentEmail = residentEmail.String
	}
	if weightLb.Valid {
		phi.WeightLb = &[]float64{weightLb.Float64}[0]
	}
	if heightFt.Valid {
		phi.HeightFt = &[]float64{heightFt.Float64}[0]
	}
	if heightIn.Valid {
		phi.HeightIn = &[]float64{heightIn.Float64}[0]
	}
	if mobilityLevel.Valid {
		level := int(mobilityLevel.Int64)
		phi.MobilityLevel = &level
	}
	if tremorStatus.Valid {
		phi.TremorStatus = tremorStatus.String
	}
	if mobilityAid.Valid {
		phi.MobilityAid = mobilityAid.String
	}
	if adlAssistance.Valid {
		phi.ADLAssistance = adlAssistance.String
	}
	if commStatus.Valid {
		phi.CommStatus = commStatus.String
	}
	if hasHypertension.Valid {
		phi.HasHypertension = hasHypertension.Bool
	}
	if hasHyperlipaemia.Valid {
		phi.HasHyperlipaemia = hasHyperlipaemia.Bool
	}
	if hasHyperglycaemia.Valid {
		phi.HasHyperglycaemia = hasHyperglycaemia.Bool
	}
	if hasStrokeHistory.Valid {
		phi.HasStrokeHistory = hasStrokeHistory.Bool
	}
	if hasParalysis.Valid {
		phi.HasParalysis = hasParalysis.Bool
	}
	if hasAlzheimer.Valid {
		phi.HasAlzheimer = hasAlzheimer.Bool
	}
	if medicalHistory.Valid {
		phi.MedicalHistory = medicalHistory.String
	}
	if homeAddressStreet.Valid {
		phi.HomeAddressStreet = homeAddressStreet.String
	}
	if homeAddressCity.Valid {
		phi.HomeAddressCity = homeAddressCity.String
	}
	if homeAddressState.Valid {
		phi.HomeAddressState = homeAddressState.String
	}
	if homeAddressPostalCode.Valid {
		phi.HomeAddressPostalCode = homeAddressPostalCode.String
	}
	if plusCode.Valid {
		phi.PlusCode = plusCode.String
	}

	return &phi, nil
}

// UpsertResidentPHI 创建或更新住户PHI信息
// 注意：UNIQUE(tenant_id, resident_id)，使用UPSERT语义
func (r *PostgresResidentsRepository) UpsertResidentPHI(ctx context.Context, tenantID, residentID string, phi *domain.ResidentPHI) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}
	if phi == nil {
		return fmt.Errorf("phi is required")
	}

	query := `
		INSERT INTO resident_phi (
			tenant_id, resident_id,
			first_name, last_name, gender, date_of_birth,
			resident_phone, resident_email,
			weight_lb, height_ft, height_in,
			mobility_level,
			tremor_status, mobility_aid, adl_assistance, comm_status,
			has_hypertension, has_hyperlipaemia, has_hyperglycaemia,
		has_stroke_history, has_paralysis, has_alzheimer,
		medical_history,
		home_address_street, home_address_city, home_address_state, home_address_postal_code, plus_code
		) VALUES (
			$1, $2,
			NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6,
			NULLIF($7, ''), NULLIF($8, ''),
			$9, $10, $11,
			$12,
			NULLIF($13, ''), NULLIF($14, ''), NULLIF($15, ''), NULLIF($16, ''),
			$17, $18, $19,
			$20, $21, $22,
			NULLIF($23, ''),
			NULLIF($24, ''), NULLIF($25, ''), NULLIF($26, ''), NULLIF($27, ''), NULLIF($28, '')
		)
		ON CONFLICT (tenant_id, resident_id)
		DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			gender = EXCLUDED.gender,
			date_of_birth = EXCLUDED.date_of_birth,
			resident_phone = EXCLUDED.resident_phone,
			resident_email = EXCLUDED.resident_email,
			weight_lb = EXCLUDED.weight_lb,
			height_ft = EXCLUDED.height_ft,
			height_in = EXCLUDED.height_in,
			mobility_level = EXCLUDED.mobility_level,
			tremor_status = EXCLUDED.tremor_status,
			mobility_aid = EXCLUDED.mobility_aid,
			adl_assistance = EXCLUDED.adl_assistance,
			comm_status = EXCLUDED.comm_status,
			has_hypertension = EXCLUDED.has_hypertension,
			has_hyperlipaemia = EXCLUDED.has_hyperlipaemia,
			has_hyperglycaemia = EXCLUDED.has_hyperglycaemia,
			has_stroke_history = EXCLUDED.has_stroke_history,
			has_paralysis = EXCLUDED.has_paralysis,
			has_alzheimer = EXCLUDED.has_alzheimer,
			medical_history = EXCLUDED.medical_history,
			home_address_street = EXCLUDED.home_address_street,
			home_address_city = EXCLUDED.home_address_city,
			home_address_state = EXCLUDED.home_address_state,
			home_address_postal_code = EXCLUDED.home_address_postal_code,
			plus_code = EXCLUDED.plus_code
	`

	// 处理可空字段
	var firstName, lastName, gender, residentPhone, residentEmail any = nil, nil, nil, nil, nil
	if phi.FirstName != "" {
		firstName = phi.FirstName
	}
	if phi.LastName != "" {
		lastName = phi.LastName
	}
	if phi.Gender != "" {
		gender = phi.Gender
	}
	if phi.ResidentPhone != "" {
		residentPhone = phi.ResidentPhone
	}
	if phi.ResidentEmail != "" {
		residentEmail = phi.ResidentEmail
	}

	var dateOfBirth any = nil
	if phi.DateOfBirth != nil {
		dateOfBirth = *phi.DateOfBirth
	}

	var weightLb, heightFt, heightIn any = nil, nil, nil
	if phi.WeightLb != nil {
		weightLb = *phi.WeightLb
	}
	if phi.HeightFt != nil {
		heightFt = *phi.HeightFt
	}
	if phi.HeightIn != nil {
		heightIn = *phi.HeightIn
	}

	var mobilityLevel any = nil
	if phi.MobilityLevel != nil {
		mobilityLevel = *phi.MobilityLevel
	}

	var tremorStatus, mobilityAid, adlAssistance, commStatus any = nil, nil, nil, nil
	if phi.TremorStatus != "" {
		tremorStatus = phi.TremorStatus
	}
	if phi.MobilityAid != "" {
		mobilityAid = phi.MobilityAid
	}
	if phi.ADLAssistance != "" {
		adlAssistance = phi.ADLAssistance
	}
	if phi.CommStatus != "" {
		commStatus = phi.CommStatus
	}

	var medicalHistory any = nil
	if phi.MedicalHistory != "" {
		medicalHistory = phi.MedicalHistory
	}

	var homeAddressStreet, homeAddressCity, homeAddressState, homeAddressPostalCode, plusCode any = nil, nil, nil, nil, nil
	if phi.HomeAddressStreet != "" {
		homeAddressStreet = phi.HomeAddressStreet
	}
	if phi.HomeAddressCity != "" {
		homeAddressCity = phi.HomeAddressCity
	}
	if phi.HomeAddressState != "" {
		homeAddressState = phi.HomeAddressState
	}
	if phi.HomeAddressPostalCode != "" {
		homeAddressPostalCode = phi.HomeAddressPostalCode
	}
	if phi.PlusCode != "" {
		plusCode = phi.PlusCode
	}

	_, err := r.db.ExecContext(ctx, query,
		tenantID, residentID,
		firstName, lastName, gender, dateOfBirth,
		residentPhone, residentEmail,
		weightLb, heightFt, heightIn,
		mobilityLevel,
		tremorStatus, mobilityAid, adlAssistance, commStatus,
		phi.HasHypertension, phi.HasHyperlipaemia, phi.HasHyperglycaemia,
		phi.HasStrokeHistory, phi.HasParalysis, phi.HasAlzheimer,
		medicalHistory,
		homeAddressStreet, homeAddressCity, homeAddressState, homeAddressPostalCode, plusCode,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert resident PHI: %w", err)
	}

	return nil
}


// ============================================
// ResidentContacts 表操作
// ============================================

// GetResidentContacts 获取住户的所有联系人
func (r *PostgresResidentsRepository) GetResidentContacts(ctx context.Context, tenantID, residentID string) ([]*domain.ResidentContact, error) {
	if tenantID == "" || residentID == "" {
		return nil, fmt.Errorf("tenant_id and resident_id are required")
	}

	query := `
		SELECT 
			contact_id::text,
			tenant_id::text,
			resident_id::text,
			slot,
			is_enabled,
			relationship,
			role,
			is_emergency_contact,
			COALESCE(alert_time_window, '{}'::jsonb)::text as alert_time_window,
			contact_first_name,
			contact_last_name,
			contact_phone,
			contact_email,
			receive_sms,
			receive_email,
			phone_hash,
			email_hash,
			password_hash
		FROM resident_contacts
		WHERE tenant_id = $1 AND resident_id = $2
		ORDER BY slot
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, residentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resident contacts: %w", err)
	}
	defer rows.Close()

	contacts := []*domain.ResidentContact{}
	for rows.Next() {
		var contact domain.ResidentContact
		var relationship, contactFirstName, contactLastName, contactPhone, contactEmail sql.NullString
		var alertTimeWindow sql.NullString
		var phoneHash, emailHash, passwordHash sql.Null[[]byte]

		err := rows.Scan(
			&contact.ContactID,
			&contact.TenantID,
			&contact.ResidentID,
			&contact.Slot,
			&contact.IsEnabled,
			&relationship,
			&contact.Role,
			&contact.IsEmergencyContact,
			&alertTimeWindow,
			&contactFirstName,
			&contactLastName,
			&contactPhone,
			&contactEmail,
			&contact.ReceiveSMS,
			&contact.ReceiveEmail,
			&phoneHash,
			&emailHash,
			&passwordHash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}

		// 处理可空字段
		if relationship.Valid {
			contact.Relationship = relationship.String
		}
		if contactFirstName.Valid {
			contact.ContactFirstName = contactFirstName.String
		}
		if contactLastName.Valid {
			contact.ContactLastName = contactLastName.String
		}
		if contactPhone.Valid {
			contact.ContactPhone = contactPhone.String
		}
		if contactEmail.Valid {
			contact.ContactEmail = contactEmail.String
		}
		if alertTimeWindow.Valid && alertTimeWindow.String != "" {
			contact.AlertTimeWindow = json.RawMessage(alertTimeWindow.String)
		}
		if phoneHash.Valid {
			contact.PhoneHash = phoneHash.V
		}
		if emailHash.Valid {
			contact.EmailHash = emailHash.V
		}
		if passwordHash.Valid {
			contact.PasswordHash = passwordHash.V
		}

		contacts = append(contacts, &contact)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate contacts: %w", err)
	}

	return contacts, nil
}

// CreateResidentContact 创建联系人
// 注意：UNIQUE(tenant_id, resident_id, slot)
func (r *PostgresResidentsRepository) CreateResidentContact(ctx context.Context, tenantID, residentID string, contact *domain.ResidentContact) (string, error) {
	if tenantID == "" || residentID == "" {
		return "", fmt.Errorf("tenant_id and resident_id are required")
	}
	if contact == nil {
		return "", fmt.Errorf("contact is required")
	}
	if contact.Slot == "" {
		return "", fmt.Errorf("slot is required")
	}

	// 处理默认值
	role := contact.Role
	if role == "" {
		role = "Family"
	}

	// 处理可空字段
	var relationshipArg any = nil
	if contact.Relationship != "" {
		relationshipArg = contact.Relationship
	}
	var contactFirstNameArg any = nil
	if contact.ContactFirstName != "" {
		contactFirstNameArg = contact.ContactFirstName
	}
	var contactLastNameArg any = nil
	if contact.ContactLastName != "" {
		contactLastNameArg = contact.ContactLastName
	}
	var contactPhoneArg any = nil
	if contact.ContactPhone != "" {
		contactPhoneArg = contact.ContactPhone
	}
	var contactEmailArg any = nil
	if contact.ContactEmail != "" {
		contactEmailArg = contact.ContactEmail
	}
	var phoneHashArg any = nil
	if len(contact.PhoneHash) > 0 {
		phoneHashArg = contact.PhoneHash
	}
	var emailHashArg any = nil
	if len(contact.EmailHash) > 0 {
		emailHashArg = contact.EmailHash
	}
	var passwordHashArg any = nil
	if len(contact.PasswordHash) > 0 {
		passwordHashArg = contact.PasswordHash
	}
	var alertTimeWindowArg any = nil
	if len(contact.AlertTimeWindow) > 0 {
		alertTimeWindowArg = string(contact.AlertTimeWindow)
	}

	var contactID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO resident_contacts (
			tenant_id, resident_id, slot, is_enabled, relationship, role,
			is_emergency_contact, alert_time_window,
			contact_first_name, contact_last_name, contact_phone, contact_email,
			receive_sms, receive_email,
			phone_hash, email_hash, password_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING contact_id::text`,
		tenantID, residentID, contact.Slot, contact.IsEnabled, relationshipArg, role,
		contact.IsEmergencyContact, alertTimeWindowArg,
		contactFirstNameArg, contactLastNameArg, contactPhoneArg, contactEmailArg,
		contact.ReceiveSMS, contact.ReceiveEmail,
		phoneHashArg, emailHashArg, passwordHashArg,
	).Scan(&contactID)
	if err != nil {
		return "", fmt.Errorf("failed to create resident contact: %w", err)
	}

	return contactID, nil
}

// UpdateResidentContact 更新联系人信息
func (r *PostgresResidentsRepository) UpdateResidentContact(ctx context.Context, tenantID, contactID string, contact *domain.ResidentContact) error {
	if tenantID == "" || contactID == "" {
		return fmt.Errorf("tenant_id and contact_id are required")
	}
	if contact == nil {
		return fmt.Errorf("contact is required")
	}

	// 构建UPDATE语句
	updates := []string{}
	args := []any{tenantID, contactID}
	argIdx := 3

	if contact.Slot != "" {
		updates = append(updates, fmt.Sprintf("slot = $%d", argIdx))
		args = append(args, contact.Slot)
		argIdx++
	}
	updates = append(updates, fmt.Sprintf("is_enabled = $%d", argIdx))
	args = append(args, contact.IsEnabled)
	argIdx++
	if contact.Relationship != "" {
		updates = append(updates, fmt.Sprintf("relationship = $%d", argIdx))
		args = append(args, contact.Relationship)
		argIdx++
	} else {
		updates = append(updates, "relationship = NULL")
	}
	if contact.Role != "" {
		updates = append(updates, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, contact.Role)
		argIdx++
	}
	updates = append(updates, fmt.Sprintf("is_emergency_contact = $%d", argIdx))
	args = append(args, contact.IsEmergencyContact)
	argIdx++
	if len(contact.AlertTimeWindow) > 0 {
		updates = append(updates, fmt.Sprintf("alert_time_window = $%d::jsonb", argIdx))
		args = append(args, string(contact.AlertTimeWindow))
		argIdx++
	} else {
		updates = append(updates, "alert_time_window = NULL")
	}
	if contact.ContactFirstName != "" {
		updates = append(updates, fmt.Sprintf("contact_first_name = $%d", argIdx))
		args = append(args, contact.ContactFirstName)
		argIdx++
	} else {
		updates = append(updates, "contact_first_name = NULL")
	}
	if contact.ContactLastName != "" {
		updates = append(updates, fmt.Sprintf("contact_last_name = $%d", argIdx))
		args = append(args, contact.ContactLastName)
		argIdx++
	} else {
		updates = append(updates, "contact_last_name = NULL")
	}
	if contact.ContactPhone != "" {
		updates = append(updates, fmt.Sprintf("contact_phone = $%d", argIdx))
		args = append(args, contact.ContactPhone)
		argIdx++
	} else {
		updates = append(updates, "contact_phone = NULL")
	}
	if contact.ContactEmail != "" {
		updates = append(updates, fmt.Sprintf("contact_email = $%d", argIdx))
		args = append(args, contact.ContactEmail)
		argIdx++
	} else {
		updates = append(updates, "contact_email = NULL")
	}
	updates = append(updates, fmt.Sprintf("receive_sms = $%d", argIdx))
	args = append(args, contact.ReceiveSMS)
	argIdx++
	updates = append(updates, fmt.Sprintf("receive_email = $%d", argIdx))
	args = append(args, contact.ReceiveEmail)
	argIdx++
	if len(contact.PhoneHash) > 0 {
		updates = append(updates, fmt.Sprintf("phone_hash = $%d", argIdx))
		args = append(args, contact.PhoneHash)
		argIdx++
	} else {
		updates = append(updates, "phone_hash = NULL")
	}
	if len(contact.EmailHash) > 0 {
		updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
		args = append(args, contact.EmailHash)
		argIdx++
	} else {
		updates = append(updates, "email_hash = NULL")
	}
	if len(contact.PasswordHash) > 0 {
		updates = append(updates, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, contact.PasswordHash)
		argIdx++
	} else {
		updates = append(updates, "password_hash = NULL")
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE resident_contacts
		SET %s
		WHERE tenant_id = $1 AND contact_id = $2
	`, strings.Join(updates, ", "))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update resident contact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("resident contact not found: tenant_id '%s', contact_id '%s'", tenantID, contactID)
	}

	return nil
}

// DeleteResidentContact 删除联系人
func (r *PostgresResidentsRepository) DeleteResidentContact(ctx context.Context, tenantID, contactID string) error {
	if tenantID == "" || contactID == "" {
		return fmt.Errorf("tenant_id and contact_id are required")
	}

	_, err := r.db.ExecContext(ctx,
		`DELETE FROM resident_contacts WHERE tenant_id = $1 AND contact_id = $2`,
		tenantID, contactID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete resident contact: %w", err)
	}

	return nil
}

// ============================================
// ResidentCaregivers 表操作
// ============================================

// GetResidentCaregivers 获取住户的护理人员关联
// 返回数组，包含两类配置：
//   1. 首先：通过所绑定的unit，unit指定的caregiver/caregiver_group（从units表获取）
//   2. 其次：通过直接绑定的caregiver/caregiver_group（从resident_caregivers表获取）
func (r *PostgresResidentsRepository) GetResidentCaregivers(ctx context.Context, tenantID, residentID string) ([]*domain.ResidentCaregiver, error) {
	if tenantID == "" || residentID == "" {
		return nil, fmt.Errorf("tenant_id and resident_id are required")
	}

	// 1. 获取住户信息（用于获取unit_id）
	var unitID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT unit_id::text FROM residents WHERE tenant_id = $1 AND resident_id = $2`,
		tenantID, residentID,
	).Scan(&unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}

	caregivers := []*domain.ResidentCaregiver{}

	// 2. 获取unit级别的caregiver配置（如果unit_id存在）
	if unitID.Valid && unitID.String != "" {
		var unitGroupList, unitUserList sql.NullString
		err = r.db.QueryRowContext(ctx,
			`SELECT 
				CASE WHEN groupList IS NULL THEN NULL ELSE groupList::text END as groupList,
				CASE WHEN userList IS NULL THEN NULL ELSE userList::text END as userList
			FROM units
			WHERE tenant_id = $1 AND unit_id = $2`,
			tenantID, unitID.String,
		).Scan(&unitGroupList, &unitUserList)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get unit caregiver config: %w", err)
		}

		// 如果unit有caregiver配置，添加到结果
		if (unitGroupList.Valid && unitGroupList.String != "" && unitGroupList.String != "null") ||
			(unitUserList.Valid && unitUserList.String != "" && unitUserList.String != "null") {
			caregiver := &domain.ResidentCaregiver{
				TenantID:   tenantID,
				ResidentID: residentID,
				Source:     "unit",
			}
			if unitGroupList.Valid && unitGroupList.String != "" && unitGroupList.String != "null" {
				caregiver.GroupList = json.RawMessage(unitGroupList.String)
			}
			if unitUserList.Valid && unitUserList.String != "" && unitUserList.String != "null" {
				caregiver.UserList = json.RawMessage(unitUserList.String)
			}
			caregivers = append(caregivers, caregiver)
		}
	}

	// 3. 获取resident级别的caregiver配置（从resident_caregivers表）
	var caregiverID sql.NullString
	var residentGroupList, residentUserList sql.NullString
	err = r.db.QueryRowContext(ctx,
		`SELECT 
			caregiver_id::text,
			CASE WHEN groupList IS NULL THEN NULL ELSE groupList::text END as groupList,
			CASE WHEN userList IS NULL THEN NULL ELSE userList::text END as userList
		FROM resident_caregivers
		WHERE tenant_id = $1 AND resident_id = $2`,
		tenantID, residentID,
	).Scan(&caregiverID, &residentGroupList, &residentUserList)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get resident caregiver config: %w", err)
	}

	// 如果resident有caregiver配置，添加到结果
	if err != sql.ErrNoRows && caregiverID.Valid {
		caregiver := &domain.ResidentCaregiver{
			CaregiverID: caregiverID.String,
			TenantID:    tenantID,
			ResidentID:  residentID,
			Source:      "resident",
		}
		if residentGroupList.Valid && residentGroupList.String != "" && residentGroupList.String != "null" {
			caregiver.GroupList = json.RawMessage(residentGroupList.String)
		}
		if residentUserList.Valid && residentUserList.String != "" && residentUserList.String != "null" {
			caregiver.UserList = json.RawMessage(residentUserList.String)
		}
		caregivers = append(caregivers, caregiver)
	}

	return caregivers, nil
}

// UpsertResidentCaregiver 创建或更新护理人员关联
// 注意：UNIQUE(tenant_id, resident_id)，使用UPSERT语义
func (r *PostgresResidentsRepository) UpsertResidentCaregiver(ctx context.Context, tenantID, residentID string, caregiver *domain.ResidentCaregiver) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}
	if caregiver == nil {
		return fmt.Errorf("caregiver is required")
	}

	// 处理可空字段
	var groupListArg any = nil
	if len(caregiver.GroupList) > 0 {
		groupListArg = string(caregiver.GroupList)
	}
	var userListArg any = nil
	if len(caregiver.UserList) > 0 {
		userListArg = string(caregiver.UserList)
	}

	query := `
		INSERT INTO resident_caregivers (
			tenant_id, resident_id, groupList, userList
		) VALUES ($1, $2, $3::jsonb, $4::jsonb)
		ON CONFLICT (tenant_id, resident_id)
		DO UPDATE SET
			groupList = EXCLUDED.groupList,
			userList = EXCLUDED.userList
	`

	_, err := r.db.ExecContext(ctx, query, tenantID, residentID, groupListArg, userListArg)
	if err != nil {
		return fmt.Errorf("failed to upsert resident caregiver: %w", err)
	}

	return nil
}

