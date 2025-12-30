package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
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
			phone,
			email,
			phone_hash,
			email_hash,
			password_hash,
			is_access_enabled,
			branch_id::text,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND resident_id = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, phone, email, branchID, unitID, roomID, bedID sql.NullString
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
		&phone,
		&email,
		&phoneHash,
		&emailHash,
		&passwordHash,
		&resident.IsAccessEnabled,
		&branchID,
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
	if phone.Valid {
		resident.Phone = phone.String
	}
	if email.Valid {
		resident.Email = email.String
	}
	if branchID.Valid {
		resident.BranchID = branchID.String
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
			phone,
			email,
			phone_hash,
			email_hash,
			password_hash,
			is_access_enabled,
			branch_id::text,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND resident_account_hash = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, phone, email, branchID, unitID, roomID, bedID sql.NullString
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
		&phone,
		&email,
		&phoneHash,
		&emailHash,
		&passwordHash,
		&resident.IsAccessEnabled,
		&branchID,
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
	if phone.Valid {
		resident.Phone = phone.String
	}
	if email.Valid {
		resident.Email = email.String
	}
	if branchID.Valid {
		resident.BranchID = branchID.String
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
			phone,
			email,
			phone_hash,
			email_hash,
			password_hash,
			is_access_enabled,
			branch_id::text,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND email_hash = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, phone, email, branchID, unitID, roomID, bedID sql.NullString
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
		&phone,
		&email,
		&phoneHashVal,
		&emailHashVal,
		&passwordHash,
		&resident.IsAccessEnabled,
		&branchID,
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
	if phone.Valid {
		resident.Phone = phone.String
	}
	if email.Valid {
		resident.Email = email.String
	}
	if branchID.Valid {
		resident.BranchID = branchID.String
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
			phone,
			email,
			phone_hash,
			email_hash,
			password_hash,
			is_access_enabled,
			branch_id::text,
			unit_id::text,
			room_id::text,
			bed_id::text
		FROM residents
		WHERE tenant_id = $1 AND phone_hash = $2
	`

	var resident domain.Resident
	var admissionDate, dischargeDate sql.NullTime
	var serviceLevel, note, phone, email, branchID, unitID, roomID, bedID sql.NullString
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
		&phone,
		&email,
		&phoneHashVal,
		&emailHashVal,
		&passwordHash,
		&resident.IsAccessEnabled,
		&branchID,
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
	if phone.Valid {
		resident.Phone = phone.String
	}
	if email.Valid {
		resident.Email = email.String
	}
	if branchID.Valid {
		resident.BranchID = branchID.String
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
	// family_tag 字段已删除，不再支持此过滤
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
			r.phone,
			r.email,
			r.phone_hash,
			r.email_hash,
			r.password_hash,
			r.is_access_enabled,
			r.branch_id::text,
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
		var serviceLevel, note, phone, email, branchID, unitID, roomID, bedID sql.NullString
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
			&phone,
			&email,
			&phoneHash,
			&emailHash,
			&passwordHash,
			&resident.IsAccessEnabled,
			&branchID,
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
		if phone.Valid {
			resident.Phone = phone.String
		}
		if email.Valid {
			resident.Email = email.String
		}
		if branchID.Valid {
			resident.BranchID = branchID.String
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
	var branchIDArg any = nil
	if resident.BranchID != "" {
		branchIDArg = resident.BranchID
	}
	var phoneArg any = nil
	if resident.Phone != "" {
		phoneArg = resident.Phone
	}
	var emailArg any = nil
	if resident.Email != "" {
		emailArg = resident.Email
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
			metadata, note, phone, email, phone_hash, email_hash, password_hash,
			is_access_enabled, branch_id, unit_id, room_id, bed_id
		) VALUES ($1, LOWER($2), $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING resident_id::text`,
		tenantID, resident.ResidentAccount, resident.ResidentAccountHash, resident.Nickname,
		admissionDate, dischargeDateArg, serviceLevelArg, status, role,
		metadataArg, noteArg, phoneArg, emailArg, phoneHashArg, emailHashArg, passwordHashArg,
		resident.IsAccessEnabled, branchIDArg, unitIDArg, roomIDArg, bedIDArg,
	).Scan(&residentID)
	if err != nil {
		return "", fmt.Errorf("failed to create resident: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return residentID, nil
}

// UpdateResident 更新住户信息
// Deprecated: Use UpdateResidentFields instead.
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
	// 处理 phone_hash：如果提供了（包括 nil），则更新
	// 使用指针判断是否提供了值，nil 表示设置为 NULL，非 nil 且长度 > 0 表示设置值
	if resident.PhoneHash != nil {
		if len(resident.PhoneHash) > 0 {
			updates = append(updates, fmt.Sprintf("phone_hash = $%d", argIdx))
			args = append(args, resident.PhoneHash)
			argIdx++
		} else {
			// 空 slice 表示设置为 NULL
			updates = append(updates, "phone_hash = NULL")
		}
	}
	// 处理 email_hash：如果提供了（包括 nil），则更新
	if resident.EmailHash != nil {
		if len(resident.EmailHash) > 0 {
			updates = append(updates, fmt.Sprintf("email_hash = $%d", argIdx))
			args = append(args, resident.EmailHash)
			argIdx++
		} else {
			// 空 slice 表示设置为 NULL
			updates = append(updates, "email_hash = NULL")
		}
	}
	if len(resident.PasswordHash) > 0 {
		updates = append(updates, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, resident.PasswordHash)
		argIdx++
	}
	if resident.Phone != "" {
		updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, resident.Phone)
		argIdx++
	} else {
		// 允许设置为NULL
		updates = append(updates, "phone = NULL")
	}
	if resident.Email != "" {
		updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, resident.Email)
		argIdx++
	} else {
		// 允许设置为NULL
		updates = append(updates, "email = NULL")
	}
	updates = append(updates, fmt.Sprintf("is_access_enabled = $%d", argIdx))
	args = append(args, resident.IsAccessEnabled)
	argIdx++
	if resident.BranchID != "" {
		updates = append(updates, fmt.Sprintf("branch_id = $%d", argIdx))
		args = append(args, resident.BranchID)
		argIdx++
	} else {
		// 允许设置为NULL
		updates = append(updates, "branch_id = NULL")
	}
	if resident.UnitID != "" {
		updates = append(updates, fmt.Sprintf("unit_id = $%d", argIdx))
		args = append(args, resident.UnitID)
		argIdx++
	} else {
		// 允许设置为NULL
		updates = append(updates, "unit_id = NULL")
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

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateResidentFields 更新住户信息（使用更新模型）
func (r *PostgresResidentsRepository) UpdateResidentFields(ctx context.Context, tenantID, residentID string, update *domain.ResidentUpdate) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}
	if update == nil {
		return fmt.Errorf("update is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	updates := []string{}
	args := []any{tenantID, residentID}
	argIdx := 3

	// Helper function to add update part
	addUpdate := func(col string, val any) {
		updates = append(updates, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	// Handle ResidentAccount (NOT NULL)
	if update.ResidentAccount != nil {
		switch update.ResidentAccount.Action {
		case domain.UpdateActionUpdate:
			if update.ResidentAccount.Value == "" {
				return fmt.Errorf("resident_account cannot be empty (NOT NULL constraint)")
			}
			updates = append(updates, fmt.Sprintf("resident_account = LOWER($%d)", argIdx))
			args = append(args, update.ResidentAccount.Value)
			argIdx++
		case domain.UpdateActionDelete:
			return fmt.Errorf("resident_account cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle ResidentAccountHash (NOT NULL)
	if update.ResidentAccountHash != nil {
		switch update.ResidentAccountHash.Action {
		case domain.UpdateActionUpdate:
			if len(update.ResidentAccountHash.Value) == 0 {
				return fmt.Errorf("resident_account_hash cannot be empty (NOT NULL constraint)")
			}
			addUpdate("resident_account_hash", update.ResidentAccountHash.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("resident_account_hash cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle Nickname (NOT NULL)
	if update.Nickname != nil {
		switch update.Nickname.Action {
		case domain.UpdateActionUpdate:
			if update.Nickname.Value == "" {
				return fmt.Errorf("nickname cannot be empty (NOT NULL constraint)")
			}
			addUpdate("nickname", update.Nickname.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("nickname cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle AdmissionDate (NOT NULL)
	if update.AdmissionDate != nil {
		switch update.AdmissionDate.Action {
		case domain.UpdateActionUpdate:
			if update.AdmissionDate.Value == nil {
				return fmt.Errorf("admission_date cannot be empty (NOT NULL constraint)")
			}
			addUpdate("admission_date", *update.AdmissionDate.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("admission_date cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle DischargeDate (nullable)
	if update.DischargeDate != nil {
		switch update.DischargeDate.Action {
		case domain.UpdateActionUpdate:
			if update.DischargeDate.Value != nil {
				addUpdate("discharge_date", *update.DischargeDate.Value)
			} else {
				updates = append(updates, "discharge_date = NULL")
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "discharge_date = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle ServiceLevel (nullable)
	if update.ServiceLevel != nil {
		switch update.ServiceLevel.Action {
		case domain.UpdateActionUpdate:
			addUpdate("service_level", update.ServiceLevel.Value)
		case domain.UpdateActionDelete:
			updates = append(updates, "service_level = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle Status (NOT NULL)
	if update.Status != nil {
		switch update.Status.Action {
		case domain.UpdateActionUpdate:
			if update.Status.Value == "" {
				return fmt.Errorf("status cannot be empty (NOT NULL constraint)")
			}
			addUpdate("status", update.Status.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("status cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle Role (NOT NULL)
	if update.Role != nil {
		switch update.Role.Action {
		case domain.UpdateActionUpdate:
			if update.Role.Value == "" {
				return fmt.Errorf("role cannot be empty (NOT NULL constraint)")
			}
			addUpdate("role", update.Role.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("role cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle Metadata (nullable JSONB)
	if update.Metadata != nil {
		switch update.Metadata.Action {
		case domain.UpdateActionUpdate:
			if len(update.Metadata.Value) > 0 {
				updates = append(updates, fmt.Sprintf("metadata = $%d::jsonb", argIdx))
				args = append(args, string(update.Metadata.Value))
				argIdx++
			} else {
				updates = append(updates, "metadata = NULL")
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "metadata = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle Note (nullable)
	if update.Note != nil {
		switch update.Note.Action {
		case domain.UpdateActionUpdate:
			addUpdate("note", update.Note.Value)
		case domain.UpdateActionDelete:
			updates = append(updates, "note = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle Phone (nullable)
	if update.Phone != nil {
		switch update.Phone.Action {
		case domain.UpdateActionUpdate:
			addUpdate("phone", update.Phone.Value)
		case domain.UpdateActionDelete:
			updates = append(updates, "phone = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle Email (nullable)
	if update.Email != nil {
		switch update.Email.Action {
		case domain.UpdateActionUpdate:
			addUpdate("email", update.Email.Value)
		case domain.UpdateActionDelete:
			updates = append(updates, "email = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle PhoneHash (nullable)
	if update.PhoneHash != nil {
		switch update.PhoneHash.Action {
		case domain.UpdateActionUpdate:
			if len(update.PhoneHash.Value) > 0 {
				addUpdate("phone_hash", update.PhoneHash.Value)
			} else {
				updates = append(updates, "phone_hash = NULL")
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "phone_hash = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle EmailHash (nullable)
	if update.EmailHash != nil {
		switch update.EmailHash.Action {
		case domain.UpdateActionUpdate:
			if len(update.EmailHash.Value) > 0 {
				addUpdate("email_hash", update.EmailHash.Value)
			} else {
				updates = append(updates, "email_hash = NULL")
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "email_hash = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle PasswordHash (NOT NULL)
	if update.PasswordHash != nil {
		switch update.PasswordHash.Action {
		case domain.UpdateActionUpdate:
			if len(update.PasswordHash.Value) == 0 {
				return fmt.Errorf("password_hash cannot be empty (NOT NULL constraint)")
			}
			addUpdate("password_hash", update.PasswordHash.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("password_hash cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle IsAccessEnabled (NOT NULL)
	if update.IsAccessEnabled != nil {
		switch update.IsAccessEnabled.Action {
		case domain.UpdateActionUpdate:
			addUpdate("is_access_enabled", update.IsAccessEnabled.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("is_access_enabled cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle BranchID (nullable)
	if update.BranchID != nil {
		switch update.BranchID.Action {
		case domain.UpdateActionUpdate:
			if update.BranchID.Value == "" {
				updates = append(updates, "branch_id = NULL")
			} else {
				addUpdate("branch_id", update.BranchID.Value)
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "branch_id = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle UnitID (nullable)
	if update.UnitID != nil {
		switch update.UnitID.Action {
		case domain.UpdateActionUpdate:
			if update.UnitID.Value == "" {
				updates = append(updates, "unit_id = NULL")
			} else {
				addUpdate("unit_id", update.UnitID.Value)
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "unit_id = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle RoomID (nullable)
	if update.RoomID != nil {
		switch update.RoomID.Action {
		case domain.UpdateActionUpdate:
			if update.RoomID.Value == "" {
				updates = append(updates, "room_id = NULL")
			} else {
				addUpdate("room_id", update.RoomID.Value)
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "room_id = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle BedID (nullable)
	if update.BedID != nil {
		switch update.BedID.Action {
		case domain.UpdateActionUpdate:
			if update.BedID.Value == "" {
				updates = append(updates, "bed_id = NULL")
			} else {
				addUpdate("bed_id", update.BedID.Value)
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "bed_id = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
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
		return fmt.Errorf("resident not found")
	}

	return tx.Commit()
}

// DeleteResident 删除住户
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

// UpsertResidentPHIFields 创建或更新住户PHI信息（使用更新模型）
// 注意：使用 UPSERT 语义（UNIQUE(tenant_id, resident_id)）
func (r *PostgresResidentsRepository) UpsertResidentPHIFields(ctx context.Context, tenantID, residentID string, update *domain.ResidentPHIUpdate) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}
	if update == nil {
		return fmt.Errorf("update is required")
	}

	// 获取当前记录（如果存在）
	currentPHI, err := r.GetResidentPHI(ctx, tenantID, residentID)
	if err != nil && err.Error() != "resident PHI not found: sql: no rows in result set" {
		// 如果不是"未找到"错误，返回错误
		if !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to get current resident PHI: %w", err)
		}
		// 如果是"未找到"，currentPHI 为 nil，继续处理
		currentPHI = nil
	}

	// Helper functions to get final values
	getFinalStringValue := func(updateField *domain.UpdateString, currentValue string) interface{} {
		if updateField == nil {
			return currentValue
		}
		switch updateField.Action {
		case domain.UpdateActionUpdate:
			return updateField.Value
		case domain.UpdateActionDelete:
			return nil
		case domain.UpdateActionKeep:
			return currentValue
		default:
			return currentValue
		}
	}

	getFinalTimeValue := func(updateField *domain.UpdateTime, currentValue *time.Time) interface{} {
		if updateField == nil {
			return currentValue
		}
		switch updateField.Action {
		case domain.UpdateActionUpdate:
			return updateField.Value
		case domain.UpdateActionDelete:
			return nil
		case domain.UpdateActionKeep:
			return currentValue
		default:
			return currentValue
		}
	}

	getFinalFloat64Value := func(updateField *domain.UpdateFloat64, currentValue *float64) interface{} {
		if updateField == nil {
			return currentValue
		}
		switch updateField.Action {
		case domain.UpdateActionUpdate:
			return updateField.Value
		case domain.UpdateActionDelete:
			return nil
		case domain.UpdateActionKeep:
			return currentValue
		default:
			return currentValue
		}
	}

	getFinalIntValue := func(updateField *domain.UpdateInt, currentValue *int) interface{} {
		if updateField == nil {
			return currentValue
		}
		switch updateField.Action {
		case domain.UpdateActionUpdate:
			return updateField.Value
		case domain.UpdateActionDelete:
			return nil
		case domain.UpdateActionKeep:
			return currentValue
		default:
			return currentValue
		}
	}

	getFinalBoolValue := func(updateField *domain.UpdateBool, currentValue bool) interface{} {
		if updateField == nil {
			return currentValue
		}
		switch updateField.Action {
		case domain.UpdateActionUpdate:
			return updateField.Value
		case domain.UpdateActionDelete:
			return nil
		case domain.UpdateActionKeep:
			return currentValue
		default:
			return currentValue
		}
	}

	// 确定最终值
	var currentFirstName, currentLastName, currentGender, currentResidentPhone, currentResidentEmail string
	var currentDateOfBirth *time.Time
	var currentWeightLb, currentHeightFt, currentHeightIn *float64
	var currentMobilityLevel *int
	var currentTremorStatus, currentMobilityAid, currentADLAssistance, currentCommStatus, currentMedicalHistory string
	var currentHasHypertension, currentHasHyperlipaemia, currentHasHyperglycaemia, currentHasStrokeHistory, currentHasParalysis, currentHasAlzheimer bool
	var currentHomeAddressStreet, currentHomeAddressCity, currentHomeAddressState, currentHomeAddressPostalCode, currentPlusCode string

	if currentPHI != nil {
		currentFirstName = currentPHI.FirstName
		currentLastName = currentPHI.LastName
		currentGender = currentPHI.Gender
		currentDateOfBirth = currentPHI.DateOfBirth
		currentResidentPhone = currentPHI.ResidentPhone
		currentResidentEmail = currentPHI.ResidentEmail
		currentWeightLb = currentPHI.WeightLb
		currentHeightFt = currentPHI.HeightFt
		currentHeightIn = currentPHI.HeightIn
		currentMobilityLevel = currentPHI.MobilityLevel
		currentTremorStatus = currentPHI.TremorStatus
		currentMobilityAid = currentPHI.MobilityAid
		currentADLAssistance = currentPHI.ADLAssistance
		currentCommStatus = currentPHI.CommStatus
		currentHasHypertension = currentPHI.HasHypertension
		currentHasHyperlipaemia = currentPHI.HasHyperlipaemia
		currentHasHyperglycaemia = currentPHI.HasHyperglycaemia
		currentHasStrokeHistory = currentPHI.HasStrokeHistory
		currentHasParalysis = currentPHI.HasParalysis
		currentHasAlzheimer = currentPHI.HasAlzheimer
		currentMedicalHistory = currentPHI.MedicalHistory
		currentHomeAddressStreet = currentPHI.HomeAddressStreet
		currentHomeAddressCity = currentPHI.HomeAddressCity
		currentHomeAddressState = currentPHI.HomeAddressState
		currentHomeAddressPostalCode = currentPHI.HomeAddressPostalCode
		currentPlusCode = currentPHI.PlusCode
	}

	finalFirstName := getFinalStringValue(update.FirstName, currentFirstName)
	finalLastName := getFinalStringValue(update.LastName, currentLastName)
	finalGender := getFinalStringValue(update.Gender, currentGender)
	finalDateOfBirth := getFinalTimeValue(update.DateOfBirth, currentDateOfBirth)
	finalResidentPhone := getFinalStringValue(update.ResidentPhone, currentResidentPhone)
	finalResidentEmail := getFinalStringValue(update.ResidentEmail, currentResidentEmail)
	finalWeightLb := getFinalFloat64Value(update.WeightLb, currentWeightLb)
	finalHeightFt := getFinalFloat64Value(update.HeightFt, currentHeightFt)
	finalHeightIn := getFinalFloat64Value(update.HeightIn, currentHeightIn)
	finalMobilityLevel := getFinalIntValue(update.MobilityLevel, currentMobilityLevel)
	finalTremorStatus := getFinalStringValue(update.TremorStatus, currentTremorStatus)
	finalMobilityAid := getFinalStringValue(update.MobilityAid, currentMobilityAid)
	finalADLAssistance := getFinalStringValue(update.ADLAssistance, currentADLAssistance)
	finalCommStatus := getFinalStringValue(update.CommStatus, currentCommStatus)
	finalHasHypertension := getFinalBoolValue(update.HasHypertension, currentHasHypertension)
	finalHasHyperlipaemia := getFinalBoolValue(update.HasHyperlipaemia, currentHasHyperlipaemia)
	finalHasHyperglycaemia := getFinalBoolValue(update.HasHyperglycaemia, currentHasHyperglycaemia)
	finalHasStrokeHistory := getFinalBoolValue(update.HasStrokeHistory, currentHasStrokeHistory)
	finalHasParalysis := getFinalBoolValue(update.HasParalysis, currentHasParalysis)
	finalHasAlzheimer := getFinalBoolValue(update.HasAlzheimer, currentHasAlzheimer)
	finalMedicalHistory := getFinalStringValue(update.MedicalHistory, currentMedicalHistory)
	finalHomeAddressStreet := getFinalStringValue(update.HomeAddressStreet, currentHomeAddressStreet)
	finalHomeAddressCity := getFinalStringValue(update.HomeAddressCity, currentHomeAddressCity)
	finalHomeAddressState := getFinalStringValue(update.HomeAddressState, currentHomeAddressState)
	finalHomeAddressPostalCode := getFinalStringValue(update.HomeAddressPostalCode, currentHomeAddressPostalCode)
	finalPlusCode := getFinalStringValue(update.PlusCode, currentPlusCode)

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

	_, err = r.db.ExecContext(ctx, query,
		tenantID, residentID,
		finalFirstName, finalLastName, finalGender, finalDateOfBirth,
		finalResidentPhone, finalResidentEmail,
		finalWeightLb, finalHeightFt, finalHeightIn,
		finalMobilityLevel,
		finalTremorStatus, finalMobilityAid, finalADLAssistance, finalCommStatus,
		finalHasHypertension, finalHasHyperlipaemia, finalHasHyperglycaemia,
		finalHasStrokeHistory, finalHasParalysis, finalHasAlzheimer,
		finalMedicalHistory,
		finalHomeAddressStreet, finalHomeAddressCity, finalHomeAddressState, finalHomeAddressPostalCode, finalPlusCode,
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
			tenant_id::text,
			resident_id::text,
			slot,
			relationship,
			is_enabled,
			COALESCE(alert_time_window, '{}'::jsonb)::text as alert_time_window,
			contact_first_name,
			contact_last_name,
			contact_phone,
			contact_email,
			receive_sms,
			receive_email
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

		err := rows.Scan(
			&contact.TenantID,
			&contact.ResidentID,
			&contact.Slot,
			&relationship,
			&contact.IsEnabled,
			&alertTimeWindow,
			&contactFirstName,
			&contactLastName,
			&contactPhone,
			&contactEmail,
			&contact.ReceiveSMS,
			&contact.ReceiveEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}

		// 处理可空字段
		contact.Relationship = relationship
		contact.ContactFirstName = contactFirstName
		contact.ContactLastName = contactLastName
		contact.ContactPhone = contactPhone
		contact.ContactEmail = contactEmail
		if alertTimeWindow.Valid && alertTimeWindow.String != "" && alertTimeWindow.String != "null" {
			contact.AlertTimeWindow = json.RawMessage(alertTimeWindow.String)
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

	// 处理可空字段
	var relationshipArg any = nil
	if contact.Relationship.Valid {
		relationshipArg = contact.Relationship.String
	}
	var contactFirstNameArg any = nil
	if contact.ContactFirstName.Valid {
		contactFirstNameArg = contact.ContactFirstName.String
	}
	var contactLastNameArg any = nil
	if contact.ContactLastName.Valid {
		contactLastNameArg = contact.ContactLastName.String
	}
	var contactPhoneArg any = nil
	if contact.ContactPhone.Valid {
		contactPhoneArg = contact.ContactPhone.String
	}
	var contactEmailArg any = nil
	if contact.ContactEmail.Valid {
		contactEmailArg = contact.ContactEmail.String
	}
	var alertTimeWindowArg any = nil
	if len(contact.AlertTimeWindow) > 0 {
		alertTimeWindowArg = string(contact.AlertTimeWindow)
	}

	// 注意：contact_phone_hash 和 contact_email_hash 列可能不存在于旧版本的数据库中
	// 如果数据库表没有这些列，则不插入它们
	// 检查列是否存在（通过尝试查询 information_schema）
	var hasPhoneHashColumn, hasEmailHashColumn bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'resident_contacts' 
			AND column_name = 'contact_phone_hash'
			AND table_schema = current_schema()
		)`,
	).Scan(&hasPhoneHashColumn)
	if err != nil {
		hasPhoneHashColumn = false // 如果查询失败，假设列不存在
	}

	err = r.db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'resident_contacts' 
			AND column_name = 'contact_email_hash'
			AND table_schema = current_schema()
		)`,
	).Scan(&hasEmailHashColumn)
	if err != nil {
		hasEmailHashColumn = false // 如果查询失败，假设列不存在
	}

	// 处理 hash 字段：如果为空或 nil，则传入 NULL（仅在列存在时）
	// 注意：数据库字段是 VARCHAR(64)，需要存储 hex 字符串，而不是二进制数据
	var contactPhoneHashArg any = nil
	if hasPhoneHashColumn && len(contact.ContactPhoneHash) > 0 {
		// 将 []byte 转换为 hex 字符串（数据库字段是 VARCHAR(64)）
		contactPhoneHashArg = hex.EncodeToString(contact.ContactPhoneHash)
	}
	var contactEmailHashArg any = nil
	if hasEmailHashColumn && len(contact.ContactEmailHash) > 0 {
		// 将 []byte 转换为 hex 字符串（数据库字段是 VARCHAR(64)）
		contactEmailHashArg = hex.EncodeToString(contact.ContactEmailHash)
	}

	// 根据列是否存在构建 SQL
	var insertSQL string
	var args []any

	if hasPhoneHashColumn && hasEmailHashColumn {
		// 两个 hash 列都存在
		insertSQL = `INSERT INTO resident_contacts (
			tenant_id, resident_id, slot, relationship,
			is_enabled, alert_time_window,
			contact_first_name, contact_last_name, contact_phone, contact_email,
			contact_phone_hash, contact_email_hash,
			receive_sms, receive_email
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (tenant_id, resident_id, slot) DO UPDATE SET
			relationship = EXCLUDED.relationship,
			is_enabled = EXCLUDED.is_enabled,
			alert_time_window = EXCLUDED.alert_time_window,
			contact_first_name = EXCLUDED.contact_first_name,
			contact_last_name = EXCLUDED.contact_last_name,
			contact_phone = EXCLUDED.contact_phone,
			contact_email = EXCLUDED.contact_email,
			contact_phone_hash = EXCLUDED.contact_phone_hash,
			contact_email_hash = EXCLUDED.contact_email_hash,
			receive_sms = EXCLUDED.receive_sms,
			receive_email = EXCLUDED.receive_email`
		args = []any{
			tenantID, residentID, contact.Slot, relationshipArg,
			contact.IsEnabled, alertTimeWindowArg,
			contactFirstNameArg, contactLastNameArg, contactPhoneArg, contactEmailArg,
			contactPhoneHashArg, contactEmailHashArg,
			contact.ReceiveSMS, contact.ReceiveEmail,
		}
	} else {
		// 至少一个 hash 列不存在，不插入 hash 列
		insertSQL = `INSERT INTO resident_contacts (
			tenant_id, resident_id, slot, relationship,
			is_enabled, alert_time_window,
			contact_first_name, contact_last_name, contact_phone, contact_email,
			receive_sms, receive_email
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (tenant_id, resident_id, slot) DO UPDATE SET
			relationship = EXCLUDED.relationship,
			is_enabled = EXCLUDED.is_enabled,
			alert_time_window = EXCLUDED.alert_time_window,
			contact_first_name = EXCLUDED.contact_first_name,
			contact_last_name = EXCLUDED.contact_last_name,
			contact_phone = EXCLUDED.contact_phone,
			contact_email = EXCLUDED.contact_email,
			receive_sms = EXCLUDED.receive_sms,
			receive_email = EXCLUDED.receive_email`
		args = []any{
			tenantID, residentID, contact.Slot, relationshipArg,
			contact.IsEnabled, alertTimeWindowArg,
			contactFirstNameArg, contactLastNameArg, contactPhoneArg, contactEmailArg,
			contact.ReceiveSMS, contact.ReceiveEmail,
		}
	}

	_, err = r.db.ExecContext(ctx, insertSQL, args...)
	if err != nil {
		return "", fmt.Errorf("failed to create resident contact: %w", err)
	}

	// 返回组合键标识（用于兼容性，实际使用 (resident_id, slot)）
	return fmt.Sprintf("%s:%s", residentID, contact.Slot), nil
}

// UpdateResidentContact 更新联系人信息
// 注意：主键是 (resident_id, slot)，不是 contact_id
func (r *PostgresResidentsRepository) UpdateResidentContact(ctx context.Context, tenantID, residentID, slot string, contact *domain.ResidentContact) error {
	if tenantID == "" || residentID == "" || slot == "" {
		return fmt.Errorf("tenant_id, resident_id and slot are required")
	}
	if contact == nil {
		return fmt.Errorf("contact is required")
	}

	// 构建UPDATE语句
	updates := []string{}
	args := []any{tenantID, residentID, slot}
	argIdx := 4

	// 处理可空字段
	if contact.Relationship.Valid {
		updates = append(updates, fmt.Sprintf("relationship = $%d", argIdx))
		args = append(args, contact.Relationship.String)
		argIdx++
	} else {
		updates = append(updates, "relationship = NULL")
	}

	updates = append(updates, fmt.Sprintf("is_enabled = $%d", argIdx))
	args = append(args, contact.IsEnabled)
	argIdx++

	if len(contact.AlertTimeWindow) > 0 {
		updates = append(updates, fmt.Sprintf("alert_time_window = $%d::jsonb", argIdx))
		args = append(args, string(contact.AlertTimeWindow))
		argIdx++
	} else {
		updates = append(updates, "alert_time_window = NULL")
	}

	if contact.ContactFirstName.Valid {
		updates = append(updates, fmt.Sprintf("contact_first_name = $%d", argIdx))
		args = append(args, contact.ContactFirstName.String)
		argIdx++
	} else {
		updates = append(updates, "contact_first_name = NULL")
	}

	if contact.ContactLastName.Valid {
		updates = append(updates, fmt.Sprintf("contact_last_name = $%d", argIdx))
		args = append(args, contact.ContactLastName.String)
		argIdx++
	} else {
		updates = append(updates, "contact_last_name = NULL")
	}

	if contact.ContactPhone.Valid {
		updates = append(updates, fmt.Sprintf("contact_phone = $%d", argIdx))
		args = append(args, contact.ContactPhone.String)
		argIdx++
	} else {
		updates = append(updates, "contact_phone = NULL")
	}

	if contact.ContactEmail.Valid {
		updates = append(updates, fmt.Sprintf("contact_email = $%d", argIdx))
		args = append(args, contact.ContactEmail.String)
		argIdx++
	} else {
		updates = append(updates, "contact_email = NULL")
	}

	if len(contact.ContactPhoneHash) > 0 {
		// 将 []byte 转换为 hex 字符串（数据库字段是 VARCHAR(64)）
		hexString := hex.EncodeToString(contact.ContactPhoneHash)
		updates = append(updates, fmt.Sprintf("contact_phone_hash = $%d", argIdx))
		args = append(args, hexString)
		argIdx++
	} else {
		updates = append(updates, "contact_phone_hash = NULL")
	}

	if len(contact.ContactEmailHash) > 0 {
		// 将 []byte 转换为 hex 字符串（数据库字段是 VARCHAR(64)）
		hexString := hex.EncodeToString(contact.ContactEmailHash)
		updates = append(updates, fmt.Sprintf("contact_email_hash = $%d", argIdx))
		args = append(args, hexString)
		argIdx++
	} else {
		updates = append(updates, "contact_email_hash = NULL")
	}

	updates = append(updates, fmt.Sprintf("receive_sms = $%d", argIdx))
	args = append(args, contact.ReceiveSMS)
	argIdx++

	updates = append(updates, fmt.Sprintf("receive_email = $%d", argIdx))
	args = append(args, contact.ReceiveEmail)
	argIdx++

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE resident_contacts
		SET %s
		WHERE tenant_id = $1 AND resident_id = $2 AND slot = $3
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
		return fmt.Errorf("resident contact not found: tenant_id '%s', resident_id '%s', slot '%s'", tenantID, residentID, slot)
	}

	return nil
}

// UpdateResidentContactFields 更新联系人（使用更新模型）
func (r *PostgresResidentsRepository) UpdateResidentContactFields(ctx context.Context, tenantID, residentID, slot string, update *domain.ResidentContactUpdate) error {
	if tenantID == "" || residentID == "" || slot == "" {
		return fmt.Errorf("tenant_id, resident_id and slot are required")
	}
	if update == nil {
		return fmt.Errorf("update is required")
	}

	updates := []string{}
	args := []any{tenantID, residentID, slot}
	argIdx := 4

	// Helper function to add update part
	addUpdate := func(col string, val any) {
		updates = append(updates, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	// Handle Relationship (nullable)
	if update.Relationship != nil {
		switch update.Relationship.Action {
		case domain.UpdateActionUpdate:
			addUpdate("relationship", update.Relationship.Value)
		case domain.UpdateActionDelete:
			updates = append(updates, "relationship = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle IsEnabled (NOT NULL)
	if update.IsEnabled != nil {
		switch update.IsEnabled.Action {
		case domain.UpdateActionUpdate:
			addUpdate("is_enabled", update.IsEnabled.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("is_enabled cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle AlertTimeWindow (nullable JSONB)
	if update.AlertTimeWindow != nil {
		switch update.AlertTimeWindow.Action {
		case domain.UpdateActionUpdate:
			if len(update.AlertTimeWindow.Value) > 0 {
				updates = append(updates, fmt.Sprintf("alert_time_window = $%d::jsonb", argIdx))
				args = append(args, string(update.AlertTimeWindow.Value))
				argIdx++
			} else {
				updates = append(updates, "alert_time_window = NULL")
			}
		case domain.UpdateActionDelete:
			updates = append(updates, "alert_time_window = NULL")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle nullable string fields
	nullableStringFields := map[string]*domain.UpdateString{
		"contact_first_name": update.ContactFirstName,
		"contact_last_name":  update.ContactLastName,
		"contact_phone":      update.ContactPhone,
		"contact_email":      update.ContactEmail,
	}

	for col, updateField := range nullableStringFields {
		if updateField != nil {
			switch updateField.Action {
			case domain.UpdateActionUpdate:
				addUpdate(col, updateField.Value)
			case domain.UpdateActionDelete:
				updates = append(updates, fmt.Sprintf("%s = NULL", col))
			case domain.UpdateActionKeep:
				// Do nothing
			}
		}
	}

	// Handle nullable bytes fields (hash fields)
	// 注意：数据库字段是 VARCHAR(64)，需要存储 hex 字符串，而不是二进制数据
	nullableBytesFields := map[string]*domain.UpdateBytes{
		"contact_phone_hash": update.ContactPhoneHash,
		"contact_email_hash": update.ContactEmailHash,
	}

	for col, updateField := range nullableBytesFields {
		if updateField != nil {
			switch updateField.Action {
			case domain.UpdateActionUpdate:
				if len(updateField.Value) > 0 {
					// 将 []byte 转换为 hex 字符串（数据库字段是 VARCHAR(64)）
					hexString := hex.EncodeToString(updateField.Value)
					addUpdate(col, hexString)
				} else {
					updates = append(updates, fmt.Sprintf("%s = NULL", col))
				}
			case domain.UpdateActionDelete:
				updates = append(updates, fmt.Sprintf("%s = NULL", col))
			case domain.UpdateActionKeep:
				// Do nothing
			}
		}
	}

	// Handle ReceiveSMS (NOT NULL)
	if update.ReceiveSMS != nil {
		switch update.ReceiveSMS.Action {
		case domain.UpdateActionUpdate:
			addUpdate("receive_sms", update.ReceiveSMS.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("receive_sms cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	// Handle ReceiveEmail (NOT NULL)
	if update.ReceiveEmail != nil {
		switch update.ReceiveEmail.Action {
		case domain.UpdateActionUpdate:
			addUpdate("receive_email", update.ReceiveEmail.Value)
		case domain.UpdateActionDelete:
			return fmt.Errorf("receive_email cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// Do nothing
		}
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE resident_contacts
		SET %s
		WHERE tenant_id = $1 AND resident_id = $2 AND slot = $3
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
		return fmt.Errorf("resident contact not found: tenant_id '%s', resident_id '%s', slot '%s'", tenantID, residentID, slot)
	}

	return nil
}

// DeleteResidentContact 删除联系人
// Deprecated: 主键是 (resident_id, slot)，不是 contact_id。使用 DeleteResidentContactBySlot 替代
func (r *PostgresResidentsRepository) DeleteResidentContact(ctx context.Context, tenantID, contactID string) error {
	// 为了向后兼容，尝试解析 contactID 为 "residentID:slot" 格式
	// 如果解析失败，返回错误
	parts := strings.Split(contactID, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid contact_id format, expected 'residentID:slot', got '%s'", contactID)
	}
	return r.DeleteResidentContactBySlot(ctx, tenantID, parts[0], parts[1])
}

// DeleteResidentContactBySlot 删除联系人（使用主键 resident_id 和 slot）
func (r *PostgresResidentsRepository) DeleteResidentContactBySlot(ctx context.Context, tenantID, residentID, slot string) error {
	if tenantID == "" || residentID == "" || slot == "" {
		return fmt.Errorf("tenant_id, resident_id and slot are required")
	}

	_, err := r.db.ExecContext(ctx,
		`DELETE FROM resident_contacts WHERE tenant_id = $1 AND resident_id = $2 AND slot = $3`,
		tenantID, residentID, slot,
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
//  1. 首先：通过所绑定的unit，unit指定的caregiver/caregiver_group（从units表获取）
//  2. 其次：通过直接绑定的caregiver/caregiver_group（从resident_caregivers表获取）
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

	// 2. 获取unit级别的caregiver配置（已移除，units表不再有groupList和userList字段）

	// 3. 获取resident级别的caregiver配置（从resident_caregivers表）
	var caregiverID sql.NullString
	var residentGroupList, residentUserList sql.NullString
	err = r.db.QueryRowContext(ctx,
		`SELECT 
			caregiver_id::text,
			CASE WHEN group_list IS NULL THEN NULL ELSE group_list::text END as group_list,
			CASE WHEN user_list IS NULL THEN NULL ELSE user_list::text END as user_list
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
			tenant_id, resident_id, group_list, user_list
		) VALUES ($1, $2, $3::jsonb, $4::jsonb)
		ON CONFLICT (tenant_id, resident_id)
		DO UPDATE SET
			group_list = EXCLUDED.group_list,
			user_list = EXCLUDED.user_list
	`

	_, err := r.db.ExecContext(ctx, query, tenantID, residentID, groupListArg, userListArg)
	if err != nil {
		return fmt.Errorf("failed to upsert resident caregiver: %w", err)
	}

	return nil
}

// UpsertResidentCaregiverFields 创建或更新护理人员关联（使用更新模型）
// 注意：使用 UPSERT 语义（UNIQUE(tenant_id, residentID)）
func (r *PostgresResidentsRepository) UpsertResidentCaregiverFields(ctx context.Context, tenantID, residentID string, update *domain.ResidentCaregiverUpdate) error {
	if tenantID == "" || residentID == "" {
		return fmt.Errorf("tenant_id and resident_id are required")
	}
	if update == nil {
		return fmt.Errorf("update is required")
	}

	// 获取当前记录（如果存在）
	currentCaregiver, err := r.GetResidentCaregivers(ctx, tenantID, residentID)
	if err != nil {
		return fmt.Errorf("failed to get current resident caregiver: %w", err)
	}

	// 确定最终值
	var finalGroupList interface{}
	var finalUserList interface{}

	// 处理 GroupList
	if update.GroupList != nil {
		switch update.GroupList.Action {
		case domain.UpdateActionUpdate:
			if len(update.GroupList.Value) > 0 {
				finalGroupList = string(update.GroupList.Value)
			} else {
				finalGroupList = nil
			}
		case domain.UpdateActionDelete:
			finalGroupList = nil
		case domain.UpdateActionKeep:
			// 保持现有值
			if len(currentCaregiver) > 0 && len(currentCaregiver[0].GroupList) > 0 {
				finalGroupList = string(currentCaregiver[0].GroupList)
			} else {
				finalGroupList = nil
			}
		}
	} else {
		// 如果没有指定，保持现有值或使用 nil
		if len(currentCaregiver) > 0 && len(currentCaregiver[0].GroupList) > 0 {
			finalGroupList = string(currentCaregiver[0].GroupList)
		} else {
			finalGroupList = nil
		}
	}

	// 处理 UserList
	if update.UserList != nil {
		switch update.UserList.Action {
		case domain.UpdateActionUpdate:
			if len(update.UserList.Value) > 0 {
				finalUserList = string(update.UserList.Value)
			} else {
				finalUserList = nil
			}
		case domain.UpdateActionDelete:
			finalUserList = nil
		case domain.UpdateActionKeep:
			// 保持现有值
			if len(currentCaregiver) > 0 && len(currentCaregiver[0].UserList) > 0 {
				finalUserList = string(currentCaregiver[0].UserList)
			} else {
				finalUserList = nil
			}
		}
	} else {
		// 如果没有指定，保持现有值或使用 nil
		if len(currentCaregiver) > 0 && len(currentCaregiver[0].UserList) > 0 {
			finalUserList = string(currentCaregiver[0].UserList)
		} else {
			finalUserList = nil
		}
	}

	query := `
		INSERT INTO resident_caregivers (
			tenant_id, resident_id, group_list, user_list
		) VALUES ($1, $2, $3::jsonb, $4::jsonb)
		ON CONFLICT (tenant_id, resident_id)
		DO UPDATE SET
			group_list = EXCLUDED.group_list,
			user_list = EXCLUDED.user_list
	`

	_, err = r.db.ExecContext(ctx, query, tenantID, residentID, finalGroupList, finalUserList)
	if err != nil {
		return fmt.Errorf("failed to upsert resident caregiver: %w", err)
	}

	return nil
}
