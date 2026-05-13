// Package repository — Resident V2 (Forward Design)
//
// 不依赖 v1 ResidentsRepository / 任何 v1 字段。
// 所有 SQL 用 v2 schema：hoa INET PK；空间归属通过位掩码 + 关联 resident_unit。
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/phi"

	"go.uber.org/zap"
)

type PostgresResidentsRepository struct {
	db      *sql.DB
	logger  *zap.Logger
	cryptor *phi.PHICryptor // PHI/contacts 加解密；nil 时跳过加密路径并返回 error
	// onResidentUnitChange — 写 resident_unit 后触发的回调（main.go 装配 → cardSync.ReconcileCards）
	// 同步调，commit 后才调。nil 时跳过（测试场景 / 卡同步未装配时）。
	// 设计：调用方（CreateResident/UpdateResident/SoftDelete/HardDelete）算出受影响的 unit scope（/80 INET）
	// 传给本回调，避免 caller 重复写 "service 层记得调 ReconcileCards" 的纪律代码。
	onResidentUnitChange func(ctx context.Context, unitScope string)
}

func NewPostgresResidentsRepository(db *sql.DB) *PostgresResidentsRepository {
	return &PostgresResidentsRepository{db: db, logger: zap.NewNop()}
}

// SetOnResidentUnitChange — main.go 装配 hook：写 resident_unit 后自动调 cardSync.ReconcileCards
func (r *PostgresResidentsRepository) SetOnResidentUnitChange(cb func(ctx context.Context, unitScope string)) {
	r.onResidentUnitChange = cb
}

// fireResidentUnitChange — 内部 helper，nil-check + log warn 失败
func (r *PostgresResidentsRepository) fireResidentUnitChange(ctx context.Context, unitScope string) {
	if r.onResidentUnitChange == nil || strings.TrimSpace(unitScope) == "" {
		return
	}
	r.onResidentUnitChange(ctx, unitScope)
}

// DB — 暴露 *sql.DB 让 service 层做 scope verify 等共用查询（不直接借出更解耦但调用面太广，权衡选这个）
func (r *PostgresResidentsRepository) DB() *sql.DB { return r.db }

func (r *PostgresResidentsRepository) SetLogger(logger *zap.Logger) {
	if logger != nil {
		r.logger = logger
	}
}

// ============================================================================
// Read
// ============================================================================

// ListResidents — tenant 范围内列表（含分页 + nickname/account ILIKE 搜索）
func (r *PostgresResidentsRepository) ListResidents(
	ctx context.Context, tenantPrefix string, f domain.ResidentListFilter,
) ([]*domain.Resident, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}

	args := []any{tenantPrefix}
	where := []string{"network(set_masklen(r.resident_id, 48)) = $1::INET"}
	argN := 2
	if !f.IncludeDelete {
		where = append(where, "r.status <> 'deleted'")
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argN))
		args = append(args, f.Status)
		argN++
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		where = append(where, fmt.Sprintf("(LOWER(r.nickname) LIKE $%d OR LOWER(COALESCE(r.resident_account,'')) LIKE $%d)", argN, argN))
		args = append(args, "%"+strings.ToLower(s)+"%")
		argN++
	}
	// Family scope：只返回与该 family user_id 关联的 resident
	if fid := strings.TrimSpace(f.FamilyUserID); fid != "" {
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM resident_caregivers rc WHERE rc.resident_id = r.resident_id AND rc.family_id = $%d::UUID AND rc.valid_to IS NULL)",
			argN,
		))
		args = append(args, fid)
		argN++
	}
	// Current Branch scope：resident 当前 active resident_unit 所在 branch (/56) 必须 == 该 branch
	if bp := strings.TrimSpace(f.BranchPrefix); bp != "" {
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM resident_unit ru WHERE ru.resident_id = r.resident_id AND ru.valid_to IS NULL AND ru.spatial_prefix <<= $%d::INET)",
			argN,
		))
		args = append(args, bp)
		argN++
	}
	whereClause := strings.Join(where, " AND ")

	// total
	var total int
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM residents r WHERE "+whereClause, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count residents: %w", err)
	}
	if total == 0 {
		return []*domain.Resident{}, 0, nil
	}

	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	// LATERAL active 子查询取 resident 当前有效 unit 分配；列名直接引 active.spatial_prefix
	// （不能在子查询的 FROM 里再 "active a" 当 sub-table — LATERAL 别名是 row-set 不是 table；
	// 历史 GetResident 用 WITH CTE，那里 CTE 名可入 FROM；这里 list 跨 resident 必须 LATERAL）。
	q := `
		SELECT host(r.resident_id),
		       host(network(set_masklen(r.resident_id, 48))) || '/48' AS tenant_id,
		       host(network(set_masklen(r.resident_id, 56))) || '/56' AS branch_prefix,
		       r.resident_slot,
		       COALESCE(r.resident_account, ''),
		       r.nickname,
		       r.status,
		       r.service_level,
		       r.admission_date,
		       r.discharge_date,
		       r.note,
		       r.family_access,
		       (SELECT u.unit_name FROM units u WHERE u.unit_id = network(set_masklen(active.spatial_prefix, 80))),
		       CASE WHEN active.m >= 88 THEN (SELECT rm.room_name FROM rooms rm WHERE rm.room_id = network(set_masklen(active.spatial_prefix, 88))) END,
		       CASE WHEN active.m >= 96 THEN (SELECT bd.bed_name FROM beds bd WHERE bd.bed_id = network(set_masklen(active.spatial_prefix, 96))) END,
		       (SELECT s.site_name FROM sites s WHERE s.site_id = network(set_masklen(active.spatial_prefix, 64))),
		       (SELECT b.branch_name FROM branches b WHERE b.branch_id = network(set_masklen(active.spatial_prefix, 56))),
		       (SELECT u.unit_type FROM units u WHERE u.unit_id = network(set_masklen(active.spatial_prefix, 80))),
		       (SELECT t.kind FROM tenants t WHERE t.tenant_id = network(set_masklen(r.resident_id, 48))) AS kind
		  FROM residents r
		  LEFT JOIN LATERAL (
		    SELECT spatial_prefix, masklen(spatial_prefix) AS m
		      FROM resident_unit
		     WHERE resident_id = r.resident_id AND valid_to IS NULL
		     ORDER BY valid_from DESC LIMIT 1
		  ) active ON TRUE
		 WHERE ` + whereClause + `
		 ORDER BY COALESCE(r.resident_account, ''), r.resident_slot
		 LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list residents: %w", err)
	}
	defer rows.Close()

	out := []*domain.Resident{}
	for rows.Next() {
		v, err := scanResident(rows)
		if err != nil {
			continue
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

// GetCurrentBranchID — 查 user 的 Current Branch（user_branches.is_primary=TRUE）
// 复用 user_service.getCurrentBranchID 同语义；ResidentService 用以做 Manager/Nurse scope 过滤
// 返回 "" 表示 user 没挂任何 branch（admin/tenant-wide），调用方不加 branch 过滤
func (r *PostgresResidentsRepository) GetCurrentBranchID(ctx context.Context, userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", nil
	}
	var branchID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT host(branch_prefix) || '/56'
		  FROM user_branches
		 WHERE user_id = $1::UUID
		   AND is_primary = TRUE
		   AND valid_to IS NULL
		 LIMIT 1`, userID).Scan(&branchID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("query current branch: %w", err)
	}
	if !branchID.Valid {
		return "", nil
	}
	return branchID.String, nil
}

// IsResidentLinkedToFamily — 检查 resident_caregivers 是否有 active 的 (resident_id, family_id) link
// GetActiveSpatialPrefix — 返回 resident 当前有效 unit 分配的 spatial_prefix（INET CIDR 字符串），
// 没有 active 分配 / 查不到 / 出错时返回空串 + nil error（service 层视为 "无 prefix" 处理）。
//
// 用于 Phase F cards 物化路径：service 在 Update/Delete 前后两次调用以检测 transfer/discharge。
func (r *PostgresResidentsRepository) GetActiveSpatialPrefix(ctx context.Context, hoa string) (string, error) {
	if strings.TrimSpace(hoa) == "" {
		return "", nil
	}
	var prefix sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT spatial_prefix::text
		  FROM resident_unit
		 WHERE resident_id = $1::INET AND valid_to IS NULL
		 LIMIT 1
	`, hoa).Scan(&prefix)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get active spatial prefix: %w", err)
	}
	if !prefix.Valid {
		return "", nil
	}
	return prefix.String, nil
}

// 用于 Family role 的 GetResident scope 强制：family 用户只能看绑定到自己的 resident
func (r *PostgresResidentsRepository) IsResidentLinkedToFamily(
	ctx context.Context, hoa, familyUserID string,
) (bool, error) {
	if strings.TrimSpace(hoa) == "" || strings.TrimSpace(familyUserID) == "" {
		return false, nil
	}
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM resident_caregivers
			 WHERE resident_id = $1::INET
			   AND family_id   = $2::UUID
			   AND valid_to IS NULL
		)`, hoa, familyUserID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check family link: %w", err)
	}
	return exists, nil
}

// GetResident — 单个 resident 含完整反推 + caregivers + teams
func (r *PostgresResidentsRepository) GetResident(
	ctx context.Context, tenantPrefix, hoa string,
) (*domain.ResidentDetail, error) {
	d := &domain.ResidentDetail{}
	var (
		serviceTier, notes                 sql.NullString
		moveIn, moveOut                    sql.NullString
		familyAccess                       sql.NullBool
		unitPrefix, roomPrefix, bedPrefix  sql.NullString
		unitType                            sql.NullInt32
		branchName, buildingName           sql.NullString
		floor                               sql.NullInt32
		unitName, roomName, bedName        sql.NullString
	)

	// 关键：hoa byte 6=0xFF (subject namespace)，**branch 不能从 hoa 反推**
	// branch / building / floor 全部从 active resident_unit.spatial_prefix 反推
	err := r.db.QueryRowContext(ctx, `
		WITH active AS (
		  SELECT spatial_prefix, masklen(spatial_prefix) AS m
		    FROM resident_unit
		   WHERE resident_id = $1::INET AND valid_to IS NULL
		   ORDER BY valid_from DESC LIMIT 1
		)
		SELECT host(r.resident_id),
		       host(network(set_masklen(r.resident_id, 48))) || '/48',
		       COALESCE(
		         (SELECT host(network(set_masklen(spatial_prefix, 56))) || '/56' FROM active),
		         ''
		       ) AS branch_prefix,
		       r.resident_slot,
		       COALESCE(r.resident_account, ''),
		       r.nickname,
		       r.status,
		       r.service_level,
		       r.admission_date::text,
		       r.discharge_date::text,
		       r.note,
		       r.family_access,

		       (SELECT host(network(set_masklen(spatial_prefix, 80))) || '/80' FROM active),
		       (SELECT CASE WHEN m >= 88 THEN host(network(set_masklen(spatial_prefix, 88))) || '/88' END FROM active),
		       (SELECT CASE WHEN m >= 96 THEN host(network(set_masklen(spatial_prefix, 96))) || '/96' END FROM active),
		       (SELECT u.unit_type FROM units u, active a WHERE u.unit_id = network(set_masklen(a.spatial_prefix, 80))),
		       (SELECT b.branch_name FROM branches b, active a WHERE b.branch_id = network(set_masklen(a.spatial_prefix, 56))),
		       (SELECT s.site_name FROM sites s, active a WHERE s.site_id = network(set_masklen(a.spatial_prefix, 64))),
		       (SELECT s.floor FROM sites s, active a WHERE s.site_id = network(set_masklen(a.spatial_prefix, 64))),
		       (SELECT u.unit_name FROM units u, active a WHERE u.unit_id = network(set_masklen(a.spatial_prefix, 80))),
		       (SELECT rm.room_name FROM rooms rm, active a WHERE a.m >= 88 AND rm.room_id = network(set_masklen(a.spatial_prefix, 88))),
		       (SELECT bd.bed_name FROM beds bd, active a WHERE a.m >= 96 AND bd.bed_id = network(set_masklen(a.spatial_prefix, 96)))
		  FROM residents r
		 WHERE r.resident_id = $1::INET
		   AND network(set_masklen(r.resident_id, 48)) = $2::INET
		`, hoa, tenantPrefix,
	).Scan(&d.ResidentID, &d.TenantID, &d.BranchID, &d.ResidentSlot,
		&d.ResidentAccount, &d.Nickname, &d.Status,
		&serviceTier, &moveIn, &moveOut, &notes, &familyAccess,
		&unitPrefix, &roomPrefix, &bedPrefix, &unitType,
		&branchName, &buildingName, &floor, &unitName, &roomName, &bedName)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("resident not found: %s", hoa)
	}
	if err != nil {
		return nil, fmt.Errorf("get resident: %w", err)
	}

	if serviceTier.Valid {
		d.ServiceLevel = &serviceTier.String
	}
	if moveIn.Valid {
		d.AdmissionDate = &moveIn.String
	}
	if moveOut.Valid {
		d.DischargeDate = &moveOut.String
	}
	if notes.Valid {
		d.Note = &notes.String
	}
	if familyAccess.Valid {
		v := familyAccess.Bool
		d.FamilyAccess = &v
	}
	if unitPrefix.Valid {
		d.UnitID = &unitPrefix.String
	}
	if roomPrefix.Valid {
		d.RoomID = &roomPrefix.String
	}
	if bedPrefix.Valid {
		d.BedID = &bedPrefix.String
	}
	if unitType.Valid {
		v := int(unitType.Int32)
		d.UnitType = &v
	}
	if branchName.Valid {
		d.BranchName = &branchName.String
	}
	if buildingName.Valid {
		// 直接用 sites.site_name (如 "A"、"B" 任意字符串)
		d.BuildingName = &buildingName.String
	}
	if floor.Valid {
		v := int(floor.Int32)
		d.Floor = &v
	}
	if unitName.Valid {
		d.UnitName = &unitName.String
	}
	if roomName.Valid {
		d.RoomName = &roomName.String
	}
	if bedName.Valid {
		d.BedName = &bedName.String
	}

	// caregivers / teams / family / contacts / phi
	d.Caregivers = r.loadCaregivers(ctx, hoa)
	d.Teams = r.loadTeams(ctx, hoa)
	d.Family = r.loadFamily(ctx, hoa)
	d.Contacts = r.loadContacts(ctx, tenantPrefix, hoa)
	d.PHI = r.loadPHI(ctx, tenantPrefix, hoa)
	r.logger.Info("[GetResident] loaded",
		zap.String("hoa", hoa),
		zap.Int("caregiver_count", len(d.Caregivers)),
		zap.Int("team_count", len(d.Teams)),
		zap.Int("family_count", len(d.Family)),
		zap.Int("contact_count", len(d.Contacts)))

	return d, nil
}

func (r *PostgresResidentsRepository) loadCaregivers(ctx context.Context, hoa string) []domain.ResidentCaregiverV2 {
	// v2: 业务标识 = user_id；不读 hoa（admin/family 等 user 无 hoa，旧 SELECT 因 NULL→Scan 失败丢行）
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.user_id::text,
		       u.user_account,
		       COALESCE(u.nickname,''),
		       COALESCE(u.role,'')
		  FROM resident_caregivers rc
		  JOIN users u ON u.user_id = rc.caregiver_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.caregiver_id IS NOT NULL
		 ORDER BY u.user_account`, hoa)
	if err != nil {
		r.logger.Error("[loadCaregivers] query failed", zap.String("hoa", hoa), zap.Error(err))
		return nil
	}
	defer rows.Close()
	out := []domain.ResidentCaregiverV2{}
	for rows.Next() {
		var c domain.ResidentCaregiverV2
		if err := rows.Scan(&c.UserID, &c.UserAccount, &c.Nickname, &c.Role); err != nil {
			r.logger.Error("[loadCaregivers] scan failed (row skipped)", zap.String("hoa", hoa), zap.Error(err))
			continue
		}
		out = append(out, c)
	}
	return out
}

func (r *PostgresResidentsRepository) loadTeams(ctx context.Context, hoa string) []domain.ResidentTeamV2 {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.team_id::text, t.team_name, t.team_kind
		  FROM resident_caregivers rc
		  JOIN care_teams t ON t.team_id = rc.care_team_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.care_team_id IS NOT NULL
		 ORDER BY t.team_name`, hoa)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []domain.ResidentTeamV2{}
	for rows.Next() {
		var t domain.ResidentTeamV2
		if err := rows.Scan(&t.TeamID, &t.TeamName, &t.TeamKind); err == nil {
			out = append(out, t)
		}
	}
	return out
}

func (r *PostgresResidentsRepository) loadFamily(ctx context.Context, hoa string) []domain.ResidentFamilyV2 {
	// v2: 业务标识 = user_id；不读 hoa
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.user_id::text,
		       u.user_account,
		       COALESCE(u.nickname,'')
		  FROM resident_caregivers rc
		  JOIN users u ON u.user_id = rc.family_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.family_id IS NOT NULL
		 ORDER BY u.user_account`, hoa)
	if err != nil {
		r.logger.Error("[loadFamily] query failed", zap.String("hoa", hoa), zap.Error(err))
		return nil
	}
	defer rows.Close()
	out := []domain.ResidentFamilyV2{}
	for rows.Next() {
		var f domain.ResidentFamilyV2
		if err := rows.Scan(&f.UserID, &f.UserAccount, &f.Nickname); err != nil {
			r.logger.Error("[loadFamily] scan failed (row skipped)", zap.String("hoa", hoa), zap.Error(err))
			continue
		}
		out = append(out, f)
	}
	return out
}

// ============================================================================
// Write
// ============================================================================

// recordAudit — 写一行 audit_log
//
// 关系类（resident_caregivers）写入和 resident 主表 CUD 都通过这里记录。
// nil tx → 用 r.db；非 nil → 走当前事务（与业务写入同步原子）。
// 失败仅 warn 不 propagate（不让审计失败阻断业务）。
func (r *PostgresResidentsRepository) recordAudit(
	ctx context.Context, tx *sql.Tx,
	actorUserID, actorRole, action, targetKind, targetID string,
	payload any,
) {
	var (
		actorArg any = nil
		payArg   any = nil
	)
	if s := strings.TrimSpace(actorUserID); s != "" {
		actorArg = s
	}
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			payArg = b
		}
	}
	q := `INSERT INTO audit_log (actor_user_id, actor_role, action, target_kind, target_id, payload, success)
	      VALUES ($1::UUID, $2, $3, $4, $5, $6, TRUE)`
	args := []any{actorArg, actorRole, action, targetKind, targetID, payArg}
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, q, args...)
	} else {
		_, err = r.db.ExecContext(ctx, q, args...)
	}
	if err != nil {
		r.logger.Warn("[recordAudit] insert failed",
			zap.String("action", action),
			zap.String("target", targetKind+":"+targetID),
			zap.Error(err))
	}
}

// CreateResident — IPAM 分配 slot + INSERT residents + 可选关联
func (r *PostgresResidentsRepository) CreateResident(
	ctx context.Context, tenantPrefix string, in *domain.ResidentCreateInput,
	actorUserID, actorRole string,
) (string, error) {
	if strings.TrimSpace(in.Nickname) == "" {
		return "", fmt.Errorf("nickname is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// 分 slot：per-tenant MAX+1，避开 0/65535
	var slot int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(resident_slot), 0) + 1
		  FROM residents
		 WHERE network(set_masklen(resident_id, 48)) = $1::INET`, tenantPrefix,
	).Scan(&slot); err != nil {
		return "", fmt.Errorf("alloc slot: %w", err)
	}
	if slot < 1 || slot >= 65535 {
		return "", fmt.Errorf("slot out of range: %d", slot)
	}

	// 构造 hoa：tenant /48 host 部分 + ":ff01:<slot hex>::"
	tenantHost := strings.SplitN(strings.TrimSuffix(tenantPrefix, "/48"), "::", 2)[0]
	hoa := fmt.Sprintf("%s:ff01:%x::", tenantHost, slot)

	// resident_account 默认 'R0001'..
	account := ""
	if in.ResidentAccount != nil {
		account = strings.TrimSpace(*in.ResidentAccount)
	}
	if account == "" {
		account = fmt.Sprintf("R%04d", slot)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO residents (resident_id, resident_slot, nickname, resident_account,
		                       service_level, admission_date, note,
		                       status)
		VALUES ($1::INET, $2, $3, $4,
		        $5, $6, $7,
		        'active')`,
		hoa, slot, in.Nickname, account,
		nullStr(in.ServiceLevel),
		nullStr(in.AdmissionDate), nullStr(in.Note),
	); err != nil {
		return "", fmt.Errorf("insert resident: %w", err)
	}

	// 空间分配（取 deepest non-empty prefix；Bed>Room>Unit>Branch）
	// 没填 unit/room/bed 时 fallback 到 branch_id (/56 招待状态)，
	// 否则 staff branch scope check (scope.go VerifyResident) 查不到 resident_unit 行 → 拒访。
	sp := pickDeepestPrefix(in.BedID, in.RoomID, in.UnitID, in.BranchID)
	if sp != "" {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO resident_unit (resident_id, spatial_prefix, move_reason) VALUES ($1::INET, $2::INET, 'initial')`,
			hoa, sp,
		); err != nil {
			return "", fmt.Errorf("insert resident_unit: %w", err)
		}
	}

	// caregivers
	for _, uid := range in.CaregiverUserIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, caregiver_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, caregiver_id) WHERE caregiver_id IS NOT NULL DO NOTHING`,
			hoa, uid,
		); err != nil {
			return "", fmt.Errorf("insert caregiver: %w", err)
		}
	}
	for _, tid := range in.CareTeamIDs {
		tid = strings.TrimSpace(tid)
		if tid == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, care_team_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, care_team_id) WHERE care_team_id IS NOT NULL DO NOTHING`,
			hoa, tid,
		); err != nil {
			return "", fmt.Errorf("insert team: %w", err)
		}
	}

	// PHI（可选）
	if in.PHI != nil {
		if err := r.upsertPHI(ctx, tx, tenantPrefix, hoa, in.PHI); err != nil {
			return "", err
		}
	}

	// contacts（可选）
	if len(in.Contacts) > 0 {
		if err := r.replaceContacts(ctx, tx, tenantPrefix, hoa, in.Contacts); err != nil {
			return "", err
		}
	}

	// audit_log（事务内，业务回滚时审计也回滚）
	r.recordAudit(ctx, tx, actorUserID, actorRole, "resident.create", "resident", hoa, map[string]any{
		"caregiver_count": len(in.CaregiverUserIDs),
		"team_count":      len(in.CareTeamIDs),
		"family_count":    len(in.FamilyUserIDs),
		"phi":             in.PHI != nil,
		"contacts":        len(in.Contacts),
	})

	if err := tx.Commit(); err != nil {
		return "", err
	}

	// Hook: 通知 cards 表 reconcile（commit 后才 fire，回滚不触发）
	// scope = 新分配的 resident_unit spatial_prefix 所在 unit /80
	if sp := pickDeepestPrefix(in.BedID, in.RoomID, in.UnitID, in.BranchID); sp != "" {
		r.fireResidentUnitChange(ctx, narrowPrefixToUnit(sp))
	}
	return hoa, nil
}

// UpdateResident — partial update（只动 input 中提供的字段）
func (r *PostgresResidentsRepository) UpdateResident(
	ctx context.Context, tenantPrefix, hoa string, in *domain.ResidentUpdateInput,
	actorUserID, actorRole string,
) error {
	r.logger.Info("[UpdateResident ENTRY]", zap.String("hoa", hoa), zap.String("tenant_prefix", tenantPrefix))
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("[UpdateResident] BeginTx failed", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	// 主表
	sets := []string{}
	args := []any{hoa, tenantPrefix}
	argN := 3
	addS := func(col string, v *string) {
		if v == nil {
			return
		}
		s := strings.TrimSpace(*v)
		if s == "" {
			sets = append(sets, col+" = NULL")
		} else {
			sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
			args = append(args, s)
			argN++
		}
	}
	addB := func(col string, v *bool) {
		if v == nil {
			return
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, *v)
		argN++
	}
	addS("nickname", in.Nickname)
	addS("resident_account", in.ResidentAccount)
	addS("status", in.Status)
	addS("service_level", in.ServiceLevel)
	addS("admission_date", in.AdmissionDate)
	addS("discharge_date", in.DischargeDate)
	addS("note", in.Note)
	addB("family_access", in.FamilyAccess)

	if len(sets) > 0 {
		q := "UPDATE residents SET " + strings.Join(sets, ", ") + ", updated_at = NOW() " +
			"WHERE resident_id = $1::INET AND network(set_masklen(resident_id, 48)) = $2::INET"
		r.logger.Info("[UpdateResident] executing main table update", zap.Strings("sets", sets), zap.Int("arg_count", len(args)))
		result, err := tx.ExecContext(ctx, q, args...)
		if err != nil {
			r.logger.Error("[UpdateResident] main table update failed", zap.Error(err))
			return fmt.Errorf("update residents: %w", err)
		}
		affected, _ := result.RowsAffected()
		r.logger.Info("[UpdateResident] main table updated", zap.Int64("rows_affected", affected))
	} else {
		r.logger.Info("[UpdateResident] no main table updates needed")
	}

	// 空间分配（任一字段提供即视为重新分配）
	// 记录 spatialChanged + oldPrefix/newPrefix 供 commit 后 hook 触发 ReconcileCards
	var spatialOldPrefix, spatialNewPrefix string
	if in.UnitID != nil || in.RoomID != nil || in.BedID != nil {
		r.logger.Info("[UpdateResident] updating spatial assignment", zap.Any("unit_id", in.UnitID), zap.Any("room_id", in.RoomID), zap.Any("bed_id", in.BedID))
		// 先抓 oldPrefix（hook 用 + branch fallback 用）
		var oldSpatial sql.NullString
		_ = tx.QueryRowContext(ctx, `
			SELECT spatial_prefix::text
			  FROM resident_unit
			 WHERE resident_id = $1::INET AND valid_to IS NULL
			 LIMIT 1`, hoa).Scan(&oldSpatial)
		if oldSpatial.Valid {
			spatialOldPrefix = oldSpatial.String
		}

		newPrefix := pickDeepestPrefix(in.BedID, in.RoomID, in.UnitID)
		// 解绑 unit 不等于"离开 branch" — 拿老行 /56 fallback，避免 staff scope 突然看不到
		// （否则 VerifyResident 查不到 active resident_unit 行 → permission denied，
		//  自己刚操作的 resident 立刻消失，体验和一致性双输）
		if newPrefix == "" && spatialOldPrefix != "" {
			// 从 oldPrefix 推 /56 branch
			var branch56 sql.NullString
			_ = tx.QueryRowContext(ctx, `SELECT host(network(set_masklen($1::INET, 56))) || '/56'`,
				spatialOldPrefix).Scan(&branch56)
			if branch56.Valid && branch56.String != "" {
				newPrefix = branch56.String
				r.logger.Info("[UpdateResident] unbind unit fallback to branch /56", zap.String("branch", newPrefix))
			}
		}
		spatialNewPrefix = newPrefix
		// 终止 active 行
		if _, err := tx.ExecContext(ctx,
			`UPDATE resident_unit SET valid_to = NOW() WHERE resident_id = $1::INET AND valid_to IS NULL`,
			hoa); err != nil {
			r.logger.Error("[UpdateResident] end resident_unit failed", zap.Error(err))
			return fmt.Errorf("end resident_unit: %w", err)
		}
		if newPrefix != "" {
			moveReason := "transfer"
			if strings.HasSuffix(newPrefix, "/56") {
				moveReason = "unbind_to_branch"
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO resident_unit (resident_id, spatial_prefix, move_reason) VALUES ($1::INET, $2::INET, $3)`,
				hoa, newPrefix, moveReason); err != nil {
				r.logger.Error("[UpdateResident] insert resident_unit failed", zap.Error(err))
				return fmt.Errorf("insert resident_unit: %w", err)
			}
		}
		r.logger.Info("[UpdateResident] spatial assignment updated")
	}

	// caregivers / teams / family（提供即重置；nil = 不动）
	if in.CaregiverUserIDs != nil {
		r.logger.Info("[UpdateResident] replacing caregivers", zap.Int("count", len(*in.CaregiverUserIDs)))
		if err := r.replaceCaregivers(ctx, tx, hoa, *in.CaregiverUserIDs); err != nil {
			return err
		}
	}
	if in.CareTeamIDs != nil {
		r.logger.Info("[UpdateResident] replacing teams", zap.Int("count", len(*in.CareTeamIDs)))
		if err := r.replaceTeams(ctx, tx, hoa, *in.CareTeamIDs); err != nil {
			return err
		}
	}
	if in.FamilyUserIDs != nil {
		r.logger.Info("[UpdateResident] replacing family", zap.Int("count", len(*in.FamilyUserIDs)))
		if err := r.replaceFamily(ctx, tx, hoa, *in.FamilyUserIDs); err != nil {
			return err
		}
	}

	// contacts —— 全量替换：delete-all-then-insert（含 PHI 字段加密）
	if in.Contacts != nil {
		r.logger.Info("[UpdateResident] replacing contacts", zap.Int("count", len(*in.Contacts)))
		if err := r.replaceContacts(ctx, tx, tenantPrefix, hoa, *in.Contacts); err != nil {
			r.logger.Error("[UpdateResident] replaceContacts failed", zap.Error(err))
			return err
		}
	}

	// PHI（提供即整体替换）
	if in.PHI != nil {
		r.logger.Info("[UpdateResident] upserting PHI")
		if err := r.upsertPHI(ctx, tx, tenantPrefix, hoa, in.PHI); err != nil {
			r.logger.Error("[UpdateResident] upsertPHI failed", zap.Error(err))
			return err
		}
	}

	// audit_log：哪些段被改了（不记录具体值，避免 PHI 落审计明文）
	auditPayload := map[string]any{
		"profile":    len(sets) > 0,
		"caregivers": in.CaregiverUserIDs != nil,
		"teams":      in.CareTeamIDs != nil,
		"family":     in.FamilyUserIDs != nil,
		"contacts":   in.Contacts != nil,
		"phi":        in.PHI != nil,
	}
	if in.CaregiverUserIDs != nil {
		auditPayload["caregiver_count"] = len(*in.CaregiverUserIDs)
	}
	if in.FamilyUserIDs != nil {
		auditPayload["family_count"] = len(*in.FamilyUserIDs)
	}
	if in.CareTeamIDs != nil {
		auditPayload["team_count"] = len(*in.CareTeamIDs)
	}
	r.recordAudit(ctx, tx, actorUserID, actorRole, "resident.update", "resident", hoa, auditPayload)

	if err := tx.Commit(); err != nil {
		r.logger.Error("[UpdateResident] tx.Commit failed", zap.Error(err))
		return err
	}
	r.logger.Info("[UpdateResident SUCCESS]", zap.String("hoa", hoa))

	// Hook: 通知 cards 表 reconcile（spatial 变化时 old + new 两次 fire）
	if spatialOldPrefix != "" {
		r.fireResidentUnitChange(ctx, narrowPrefixToUnit(spatialOldPrefix))
	}
	if spatialNewPrefix != "" && spatialNewPrefix != spatialOldPrefix {
		r.fireResidentUnitChange(ctx, narrowPrefixToUnit(spatialNewPrefix))
	}
	// nickname 改了也要触发 reconcile（card_name 同步）— 此时 spatial 不一定变
	if spatialOldPrefix == "" && spatialNewPrefix == "" && in.Nickname != nil {
		// 用 LPM 查当前 active 行算 scope
		var cur sql.NullString
		_ = r.db.QueryRowContext(ctx,
			`SELECT spatial_prefix::text FROM resident_unit WHERE resident_id=$1::INET AND valid_to IS NULL LIMIT 1`,
			hoa).Scan(&cur)
		if cur.Valid && cur.String != "" {
			r.fireResidentUnitChange(ctx, narrowPrefixToUnit(cur.String))
		}
	}
	return nil
}

// resolveResidentBranch — 从 active resident_unit 反推 resident 当前 branch /56
// 返回 "" 表示 resident 尚未分配空间（无 branch 约束）
func (r *PostgresResidentsRepository) resolveResidentBranch(ctx context.Context, tx *sql.Tx, hoa string) (string, error) {
	var branchStr sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT host(network(set_masklen(spatial_prefix, 56))) || '/56'
		  FROM resident_unit
		 WHERE resident_id = $1::INET AND valid_to IS NULL
		 ORDER BY valid_from DESC LIMIT 1
	`, hoa).Scan(&branchStr)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !branchStr.Valid {
		return "", nil
	}
	return branchStr.String, nil
}

func (r *PostgresResidentsRepository) replaceCaregivers(ctx context.Context, tx *sql.Tx, hoa string, userIDs []string) error {
	r.logger.Info("[replaceCaregivers ENTRY]", zap.String("hoa", hoa), zap.Int("caregiver_count", len(userIDs)), zap.Strings("caregiver_ids", userIDs))

	residentBranch, err := r.resolveResidentBranch(ctx, tx, hoa)
	if err != nil {
		return fmt.Errorf("resolve resident branch: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND caregiver_id IS NOT NULL`,
		hoa)
	if err != nil {
		r.logger.Error("[replaceCaregivers] DELETE failed", zap.Error(err))
		return err
	}
	deleted, _ := result.RowsAffected()
	r.logger.Info("[replaceCaregivers] old caregivers deleted", zap.Int64("rows_deleted", deleted))

	for i, uid := range userIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			r.logger.Warn("[replaceCaregivers] empty caregiver_id at index", zap.Int("index", i))
			continue
		}
		// same-branch 校验：resident 已分配 branch 时，caregiver 必须挂在同 branch
		// caregiver 没 user_branches 行（admin/manager 跨 branch 角色）也允许
		if residentBranch != "" {
			var allowed bool
			if err := tx.QueryRowContext(ctx, `
				SELECT NOT EXISTS (SELECT 1 FROM user_branches WHERE user_id = $1::UUID AND valid_to IS NULL)
				    OR EXISTS (SELECT 1 FROM user_branches WHERE user_id = $1::UUID AND branch_prefix = $2::INET AND valid_to IS NULL)
			`, uid, residentBranch).Scan(&allowed); err != nil {
				return fmt.Errorf("check caregiver branch: %w", err)
			}
			if !allowed {
				return fmt.Errorf("caregiver %s not assigned to resident's branch %s", uid, residentBranch)
			}
		}
		r.logger.Info("[replaceCaregivers] inserting caregiver", zap.Int("index", i), zap.String("uid", uid))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, caregiver_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, caregiver_id) WHERE caregiver_id IS NOT NULL DO NOTHING`,
			hoa, uid,
		); err != nil {
			r.logger.Error("[replaceCaregivers] INSERT failed", zap.String("uid", uid), zap.Error(err))
			return err
		}
	}
	r.logger.Info("[replaceCaregivers SUCCESS]", zap.String("hoa", hoa))
	return nil
}

func (r *PostgresResidentsRepository) replaceTeams(ctx context.Context, tx *sql.Tx, hoa string, teamIDs []string) error {
	r.logger.Info("[replaceTeams ENTRY]", zap.String("hoa", hoa), zap.Int("team_count", len(teamIDs)), zap.Strings("team_ids", teamIDs))

	residentBranch, err := r.resolveResidentBranch(ctx, tx, hoa)
	if err != nil {
		return fmt.Errorf("resolve resident branch: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND care_team_id IS NOT NULL`,
		hoa)
	if err != nil {
		r.logger.Error("[replaceTeams] DELETE failed", zap.Error(err))
		return err
	}
	deleted, _ := result.RowsAffected()
	r.logger.Info("[replaceTeams] old teams deleted", zap.Int64("rows_deleted", deleted))

	for i, tid := range teamIDs {
		tid = strings.TrimSpace(tid)
		if tid == "" {
			r.logger.Warn("[replaceTeams] empty team_id at index", zap.Int("index", i))
			continue
		}
		// same-branch 校验：care_team.branch_id 必须 == resident.branch
		if residentBranch != "" {
			var teamBranch string
			if err := tx.QueryRowContext(ctx,
				`SELECT host(branch_id) || '/56' FROM care_teams WHERE team_id = $1::UUID`, tid,
			).Scan(&teamBranch); err != nil {
				return fmt.Errorf("lookup care_team branch: %w", err)
			}
			if teamBranch != residentBranch {
				return fmt.Errorf("care_team %s belongs to branch %s; cannot bind to resident in branch %s",
					tid, teamBranch, residentBranch)
			}
		}
		r.logger.Info("[replaceTeams] inserting team", zap.Int("index", i), zap.String("tid", tid))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, care_team_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, care_team_id) WHERE care_team_id IS NOT NULL DO NOTHING`,
			hoa, tid,
		); err != nil {
			r.logger.Error("[replaceTeams] INSERT failed", zap.String("tid", tid), zap.Error(err))
			return err
		}
	}
	r.logger.Info("[replaceTeams SUCCESS]", zap.String("hoa", hoa))
	return nil
}

func (r *PostgresResidentsRepository) replaceFamily(ctx context.Context, tx *sql.Tx, hoa string, familyUserIDs []string) error {
	r.logger.Info("[replaceFamily ENTRY]", zap.String("hoa", hoa), zap.Int("family_count", len(familyUserIDs)), zap.Strings("family_ids", familyUserIDs))
	result, err := tx.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND family_id IS NOT NULL`,
		hoa)
	if err != nil {
		r.logger.Error("[replaceFamily] DELETE failed", zap.Error(err))
		return err
	}
	deleted, _ := result.RowsAffected()
	r.logger.Info("[replaceFamily] old family deleted", zap.Int64("rows_deleted", deleted))

	for i, fid := range familyUserIDs {
		fid = strings.TrimSpace(fid)
		if fid == "" {
			r.logger.Warn("[replaceFamily] empty family_id at index", zap.Int("index", i))
			continue
		}
		r.logger.Info("[replaceFamily] inserting family", zap.Int("index", i), zap.String("fid", fid))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, family_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, family_id) WHERE family_id IS NOT NULL DO NOTHING`,
			hoa, fid,
		); err != nil {
			r.logger.Error("[replaceFamily] INSERT failed", zap.String("fid", fid), zap.Error(err))
			return err
		}
	}
	r.logger.Info("[replaceFamily SUCCESS]", zap.String("hoa", hoa))
	return nil
}

// loadContacts — 读取并解密 resident_contacts；cryptor 未注入时静默返回空（避免 GET 整体失败）
// optional string helpers — 空指针 / 解密失败 → nil
func (r *PostgresResidentsRepository) loadContacts(ctx context.Context, tenantPrefix, hoa string) []domain.ResidentContactV2 {
	if r.cryptor == nil {
		r.logger.Warn("[loadContacts] PHI cryptor not configured; skip contacts load", zap.String("hoa", hoa))
		return nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT contact_id::text,
		       COALESCE(linked_user_id::text, ''),
		       relationship,
		       contact_first_name_enc, contact_last_name_enc,
		       contact_phone_enc, contact_email_enc,
		       receive_sms, receive_email
		  FROM resident_contacts
		 WHERE resident_id = $1::INET
		 ORDER BY created_at`, hoa)
	if err != nil {
		r.logger.Error("[loadContacts] query failed", zap.String("hoa", hoa), zap.Error(err))
		return nil
	}
	defer rows.Close()

	out := []domain.ResidentContactV2{}
	for rows.Next() {
		var (
			contactID, linked, relationship          string
			firstEnc, lastEnc, phoneEnc, emailEnc    []byte
			receiveSMS, receiveEmail                 bool
		)
		if err := rows.Scan(&contactID, &linked, &relationship, &firstEnc, &lastEnc, &phoneEnc, &emailEnc, &receiveSMS, &receiveEmail); err != nil {
			r.logger.Error("[loadContacts] scan failed", zap.Error(err))
			continue
		}
		c := domain.ResidentContactV2{
			ContactID:    contactID,
			Relationship: relationship,
			ReceiveSMS:   receiveSMS,
			ReceiveEmail: receiveEmail,
		}
		if linked != "" {
			c.LinkedUserID = &linked
		}
		decStr := func(blob []byte) *string {
			if len(blob) == 0 {
				return nil
			}
			s, err := r.cryptor.DecryptString(tenantPrefix, blob)
			if err != nil {
				r.logger.Warn("[loadContacts] decrypt field failed", zap.String("contact_id", contactID), zap.Error(err))
				return nil
			}
			if s == "" {
				return nil
			}
			return &s
		}
		c.ContactFirstName = decStr(firstEnc)
		c.ContactLastName = decStr(lastEnc)
		c.ContactPhone = decStr(phoneEnc)
		c.ContactEmail = decStr(emailEnc)
		out = append(out, c)
	}
	return out
}

// replaceContacts — 全量替换 resident_contacts；transaction-level，含 PHI 字段加密
func (r *PostgresResidentsRepository) replaceContacts(
	ctx context.Context, tx *sql.Tx, tenantPrefix, hoa string, contacts []domain.ResidentContactV2,
) error {
	// 加密支持必须可用（contacts PHI 字段是 BYTEA，不能明文存）
	if r.cryptor == nil && len(contacts) > 0 {
		return fmt.Errorf("contacts write requires PHI cryptor (KMS_SOCKET + MASTER_PIN not configured)")
	}
	r.logger.Info("[replaceContacts ENTRY]", zap.String("hoa", hoa), zap.Int("count", len(contacts)))

	result, err := tx.ExecContext(ctx, `DELETE FROM resident_contacts WHERE resident_id = $1::INET`, hoa)
	if err != nil {
		r.logger.Error("[replaceContacts] DELETE failed", zap.Error(err))
		return err
	}
	deleted, _ := result.RowsAffected()
	r.logger.Info("[replaceContacts] old contacts deleted", zap.Int64("rows_deleted", deleted))

	encStr := func(s *string) ([]byte, error) {
		if s == nil || *s == "" {
			return nil, nil
		}
		return r.cryptor.EncryptString(tenantPrefix, *s)
	}

	for i, c := range contacts {
		firstEnc, err := encStr(c.ContactFirstName)
		if err != nil {
			return fmt.Errorf("encrypt first_name at index %d: %w", i, err)
		}
		lastEnc, err := encStr(c.ContactLastName)
		if err != nil {
			return fmt.Errorf("encrypt last_name at index %d: %w", i, err)
		}
		phoneEnc, err := encStr(c.ContactPhone)
		if err != nil {
			return fmt.Errorf("encrypt phone at index %d: %w", i, err)
		}
		emailEnc, err := encStr(c.ContactEmail)
		if err != nil {
			return fmt.Errorf("encrypt email at index %d: %w", i, err)
		}
		relationship := strings.TrimSpace(c.Relationship)
		if relationship == "" {
			relationship = "other"
		}
		var linkedUserID interface{}
		if c.LinkedUserID != nil && strings.TrimSpace(*c.LinkedUserID) != "" {
			linkedUserID = strings.TrimSpace(*c.LinkedUserID)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO resident_contacts
			    (resident_id, linked_user_id, relationship,
			     contact_first_name_enc, contact_last_name_enc,
			     contact_phone_enc, contact_email_enc,
			     receive_sms, receive_email)
			VALUES ($1::INET, $2, $3, $4, $5, $6, $7, $8, $9)`,
			hoa, linkedUserID, relationship,
			firstEnc, lastEnc, phoneEnc, emailEnc,
			c.ReceiveSMS, c.ReceiveEmail,
		); err != nil {
			r.logger.Error("[replaceContacts] INSERT failed", zap.Int("index", i), zap.Error(err))
			return err
		}
	}
	r.logger.Info("[replaceContacts SUCCESS]", zap.String("hoa", hoa))
	return nil
}

// loadPHI — 读取 resident_phi 行并全字段 AES-256-GCM 解密；cryptor 未注入或行不存在 → nil
//
// schema: resident_phi.resident_id = INET PK；_enc BYTEA 列存 owl-common envelope（含 iv+tag）。
// plus_code 与 otp_* 列明文存（短码 + 一次性授权码不视作 PHI 同等敏感）。
func (r *PostgresResidentsRepository) loadPHI(ctx context.Context, tenantPrefix, hoa string) *domain.ResidentPHIv2 {
	if r.cryptor == nil {
		r.logger.Warn("[loadPHI] PHI cryptor not configured; skip", zap.String("hoa", hoa))
		return nil
	}
	row := r.db.QueryRowContext(ctx, `
		SELECT first_name_enc, last_name_enc, gender_enc, date_of_birth_enc,
		       resident_phone_enc, resident_email_enc,
		       weight_lb_enc, height_ft_enc, height_in_enc,
		       mobility_level_enc, tremor_status_enc, mobility_aid_enc,
		       adl_assistance_enc, comm_status_enc,
		       has_hypertension_enc, has_hyperlipaemia_enc, has_hyperglycaemia_enc,
		       has_stroke_history_enc, has_paralysis_enc, has_alzheimer_enc,
		       medical_history_enc,
		       home_address_street_enc, home_address_city_enc,
		       home_address_state_enc, home_address_postal_code_enc,
		       plus_code,
		       otp_code, otp_purpose,
		       otp_issued_at::text, otp_expires_at::text,
		       otp_used_at::text, otp_used_by::text
		  FROM resident_phi
		 WHERE resident_id = $1::INET`, hoa)

	var (
		firstName, lastName, gender, dob, phone, email []byte
		weightLb, heightFt, heightIn                []byte
		mobilityLevel, tremorStatus, mobilityAid    []byte
		adlAssistance, commStatus                    []byte
		hHTN, hHLD, hHGM, hStroke, hPara, hAlz       []byte
		medHistory                                   []byte
		street, city, state, postal                  []byte
		plusCode                                     sql.NullString
		otpCode, otpPurpose                          sql.NullString
		otpIssuedAt, otpExpiresAt                    sql.NullString
		otpUsedAt, otpUsedBy                         sql.NullString
	)
	err := row.Scan(
		&firstName, &lastName, &gender, &dob, &phone, &email,
		&weightLb, &heightFt, &heightIn,
		&mobilityLevel, &tremorStatus, &mobilityAid,
		&adlAssistance, &commStatus,
		&hHTN, &hHLD, &hHGM, &hStroke, &hPara, &hAlz,
		&medHistory,
		&street, &city, &state, &postal,
		&plusCode,
		&otpCode, &otpPurpose,
		&otpIssuedAt, &otpExpiresAt,
		&otpUsedAt, &otpUsedBy,
	)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		r.logger.Error("[loadPHI] scan failed", zap.String("hoa", hoa), zap.Error(err))
		return nil
	}

	out := &domain.ResidentPHIv2{}
	c := r.cryptor
	decStr := func(blob []byte) *string {
		if len(blob) == 0 {
			return nil
		}
		s, err := c.DecryptString(tenantPrefix, blob)
		if err != nil {
			r.logger.Warn("[loadPHI] decrypt string failed", zap.String("hoa", hoa), zap.Error(err))
			return nil
		}
		if s == "" {
			return nil
		}
		return &s
	}
	decFloat := func(blob []byte) *float64 {
		if len(blob) == 0 {
			return nil
		}
		v, err := c.DecryptFloat(tenantPrefix, blob)
		if err != nil {
			r.logger.Warn("[loadPHI] decrypt float failed", zap.String("hoa", hoa), zap.Error(err))
			return nil
		}
		return v
	}
	decInt := func(blob []byte) *int {
		if len(blob) == 0 {
			return nil
		}
		v, err := c.DecryptInt(tenantPrefix, blob)
		if err != nil {
			r.logger.Warn("[loadPHI] decrypt int failed", zap.String("hoa", hoa), zap.Error(err))
			return nil
		}
		return v
	}
	decBool := func(blob []byte) *bool {
		if len(blob) == 0 {
			return nil
		}
		v, err := c.DecryptBool(tenantPrefix, blob)
		if err != nil {
			r.logger.Warn("[loadPHI] decrypt bool failed", zap.String("hoa", hoa), zap.Error(err))
			return nil
		}
		return &v
	}

	out.FirstName = decStr(firstName)
	out.LastName = decStr(lastName)
	out.Gender = decStr(gender)
	out.DateOfBirth = decStr(dob)
	out.ResidentPhone = decStr(phone)
	out.ResidentEmail = decStr(email)
	out.WeightLb = decFloat(weightLb)
	out.HeightFt = decFloat(heightFt)
	out.HeightIn = decFloat(heightIn)
	out.MobilityLevel = decInt(mobilityLevel)
	out.TremorStatus = decStr(tremorStatus)
	out.MobilityAid = decStr(mobilityAid)
	out.ADLAssistance = decStr(adlAssistance)
	out.CommStatus = decStr(commStatus)
	out.HasHypertension = decBool(hHTN)
	out.HasHyperlipaemia = decBool(hHLD)
	out.HasHyperglycaemia = decBool(hHGM)
	out.HasStrokeHistory = decBool(hStroke)
	out.HasParalysis = decBool(hPara)
	out.HasAlzheimer = decBool(hAlz)
	out.MedicalHistory = decStr(medHistory)
	out.HomeAddressStreet = decStr(street)
	out.HomeAddressCity = decStr(city)
	out.HomeAddressState = decStr(state)
	out.HomeAddressPostalCode = decStr(postal)
	if plusCode.Valid && plusCode.String != "" {
		s := plusCode.String
		out.PlusCode = &s
	}

	// OTP（明文列）—— 任一非空才返
	if otpCode.Valid || otpPurpose.Valid || otpIssuedAt.Valid || otpExpiresAt.Valid {
		o := &domain.ResidentOTPv2{}
		if otpCode.Valid && otpCode.String != "" {
			s := otpCode.String
			o.Code = &s
		}
		if otpPurpose.Valid && otpPurpose.String != "" {
			s := otpPurpose.String
			o.Purpose = &s
		}
		if otpIssuedAt.Valid && otpIssuedAt.String != "" {
			s := otpIssuedAt.String
			o.IssuedAt = &s
		}
		if otpExpiresAt.Valid && otpExpiresAt.String != "" {
			s := otpExpiresAt.String
			o.ExpiresAt = &s
		}
		if otpUsedAt.Valid && otpUsedAt.String != "" {
			s := otpUsedAt.String
			o.UsedAt = &s
		}
		if otpUsedBy.Valid && otpUsedBy.String != "" {
			s := otpUsedBy.String
			o.UsedBy = &s
		}
		out.OTP = o
	}

	return out
}

// upsertPHI — 加密全字段后 INSERT ... ON CONFLICT DO UPDATE。
//
// 语义：调用方传 in（非 nil）即视为整体写入 → 未提供的字段写 NULL（清空），
//      这与 contacts replaceContacts 全量替换语义一致；FE 在保存 PHI 时必须把当前完整状态送上来。
func (r *PostgresResidentsRepository) upsertPHI(
	ctx context.Context, tx *sql.Tx, tenantPrefix, hoa string, in *domain.ResidentPHIv2,
) error {
	if in == nil {
		return nil
	}
	if r.cryptor == nil {
		return fmt.Errorf("PHI write requires cryptor (KMS_SOCKET + MASTER_PIN not configured)")
	}
	c := r.cryptor

	encStr := func(v *string) ([]byte, error) {
		if v == nil || *v == "" {
			return nil, nil
		}
		return c.EncryptString(tenantPrefix, *v)
	}
	encFloat := func(v *float64) ([]byte, error) { return c.EncryptFloat(tenantPrefix, v) }
	encInt := func(v *int) ([]byte, error) { return c.EncryptInt(tenantPrefix, v) }
	encBool := func(v *bool) ([]byte, error) {
		if v == nil {
			return nil, nil
		}
		return c.EncryptBool(tenantPrefix, *v)
	}

	type field struct {
		name string
		blob []byte
	}
	encs := make([]field, 0, 24)
	encode := func(name string, blob []byte, err error) error {
		if err != nil {
			return fmt.Errorf("encrypt %s: %w", name, err)
		}
		encs = append(encs, field{name: name, blob: blob})
		return nil
	}

	{
		b, e := encStr(in.FirstName)
		if err := encode("first_name_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.LastName)
		if err := encode("last_name_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.Gender)
		if err := encode("gender_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.DateOfBirth)
		if err := encode("date_of_birth_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.ResidentPhone)
		if err := encode("resident_phone_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.ResidentEmail)
		if err := encode("resident_email_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encFloat(in.WeightLb)
		if err := encode("weight_lb_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encFloat(in.HeightFt)
		if err := encode("height_ft_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encFloat(in.HeightIn)
		if err := encode("height_in_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encInt(in.MobilityLevel)
		if err := encode("mobility_level_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.TremorStatus)
		if err := encode("tremor_status_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.MobilityAid)
		if err := encode("mobility_aid_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.ADLAssistance)
		if err := encode("adl_assistance_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.CommStatus)
		if err := encode("comm_status_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encBool(in.HasHypertension)
		if err := encode("has_hypertension_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encBool(in.HasHyperlipaemia)
		if err := encode("has_hyperlipaemia_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encBool(in.HasHyperglycaemia)
		if err := encode("has_hyperglycaemia_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encBool(in.HasStrokeHistory)
		if err := encode("has_stroke_history_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encBool(in.HasParalysis)
		if err := encode("has_paralysis_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encBool(in.HasAlzheimer)
		if err := encode("has_alzheimer_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.MedicalHistory)
		if err := encode("medical_history_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.HomeAddressStreet)
		if err := encode("home_address_street_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.HomeAddressCity)
		if err := encode("home_address_city_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.HomeAddressState)
		if err := encode("home_address_state_enc", b, e); err != nil {
			return err
		}
	}
	{
		b, e := encStr(in.HomeAddressPostalCode)
		if err := encode("home_address_postal_code_enc", b, e); err != nil {
			return err
		}
	}

	// 列表：24 个 _enc + plus_code + 6 OTP 列 = 31 个非主键列
	colNames := make([]string, 0, 31)
	args := []any{hoa}
	placeholders := make([]string, 0, 31)
	argN := 2
	for _, f := range encs {
		colNames = append(colNames, f.name)
		placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
		// nil-valued BYTEA blob → NULL（清空字段）
		if f.blob == nil {
			args = append(args, nil)
		} else {
			args = append(args, f.blob)
		}
		argN++
	}

	addPlain := func(col string, v any) {
		colNames = append(colNames, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
		args = append(args, v)
		argN++
	}
	if in.PlusCode != nil && *in.PlusCode != "" {
		addPlain("plus_code", *in.PlusCode)
	} else {
		addPlain("plus_code", nil)
	}

	// OTP 字段（全空时也写 NULL，达成清空语义）
	var oCode, oPurpose, oIssued, oExpires, oUsed, oUsedBy any
	if o := in.OTP; o != nil {
		if o.Code != nil && *o.Code != "" {
			oCode = *o.Code
		}
		if o.Purpose != nil && *o.Purpose != "" {
			oPurpose = *o.Purpose
		}
		if o.IssuedAt != nil && *o.IssuedAt != "" {
			oIssued = *o.IssuedAt
		}
		if o.ExpiresAt != nil && *o.ExpiresAt != "" {
			oExpires = *o.ExpiresAt
		}
		if o.UsedAt != nil && *o.UsedAt != "" {
			oUsed = *o.UsedAt
		}
		if o.UsedBy != nil && *o.UsedBy != "" {
			oUsedBy = *o.UsedBy
		}
	}
	addPlain("otp_code", oCode)
	addPlain("otp_purpose", oPurpose)
	addPlain("otp_issued_at", oIssued)
	addPlain("otp_expires_at", oExpires)
	addPlain("otp_used_at", oUsed)
	addPlain("otp_used_by", oUsedBy)

	// ON CONFLICT 更新所有列
	setClauses := make([]string, 0, len(colNames))
	for _, n := range colNames {
		setClauses = append(setClauses, n+" = EXCLUDED."+n)
	}

	q := "INSERT INTO resident_phi (resident_id, " + strings.Join(colNames, ", ") +
		") VALUES ($1::INET, " + strings.Join(placeholders, ", ") + ") " +
		"ON CONFLICT (resident_id) DO UPDATE SET " + strings.Join(setClauses, ", ") +
		", encrypted_at = NOW()"

	r.logger.Info("[upsertPHI] writing", zap.String("hoa", hoa), zap.Int("col_count", len(colNames)))
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		r.logger.Error("[upsertPHI] exec failed", zap.String("hoa", hoa), zap.Error(err))
		return fmt.Errorf("upsert resident_phi: %w", err)
	}
	r.logger.Info("[upsertPHI SUCCESS]", zap.String("hoa", hoa))
	return nil
}

func (r *PostgresResidentsRepository) SoftDelete(ctx context.Context, hoa string) error {
	// status='disabled' — 保留 resident_account 占用（audit 一致性）
	// 关闭 resident_unit active 行 → branch-scoped staff 列表不再显示
	// 先抓 oldPrefix 供 hook 用
	var old sql.NullString
	_ = r.db.QueryRowContext(ctx,
		`SELECT spatial_prefix::text FROM resident_unit WHERE resident_id=$1::INET AND valid_to IS NULL LIMIT 1`,
		hoa).Scan(&old)

	if _, err := r.db.ExecContext(ctx,
		`UPDATE residents SET status = 'disabled', updated_at = NOW() WHERE resident_id = $1::INET`,
		hoa); err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	_, _ = r.db.ExecContext(ctx,
		`UPDATE resident_unit SET valid_to = NOW() WHERE resident_id = $1::INET AND valid_to IS NULL`, hoa)

	// Hook: 通知 cards reconcile 旧 unit（resident 走了，cards 表 resident_id 要置空）
	if old.Valid && old.String != "" {
		r.fireResidentUnitChange(ctx, narrowPrefixToUnit(old.String))
	}
	return nil
}

// CheckClearable — 三表无任何 resident_id 引用才允许 hard delete
func (r *PostgresResidentsRepository) CheckClearable(ctx context.Context, hoa string) (*domain.ResidentClearCheckResult, error) {
	res := &domain.ResidentClearCheckResult{}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM alarm_events WHERE resident_id = $1::INET`, hoa,
	).Scan(&res.AlarmEventsCount); err != nil {
		// alarm_events 列可能不含 resident_id；忽略错误视为 0
		res.AlarmEventsCount = 0
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM event_log WHERE resident_id = $1::INET`, hoa,
	).Scan(&res.EventLogCount); err != nil {
		res.EventLogCount = 0
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM monitor_stream WHERE resident_id = $1::INET`, hoa,
	).Scan(&res.MonitorCount); err != nil {
		res.MonitorCount = 0
	}
	total := res.AlarmEventsCount + res.EventLogCount + res.MonitorCount
	if total > 0 {
		res.CanClear = false
		res.Reason = fmt.Sprintf("resident has historical records (alarms: %d, events: %d, monitor: %d). Use soft delete.",
			res.AlarmEventsCount, res.EventLogCount, res.MonitorCount)
	} else {
		res.CanClear = true
	}
	return res, nil
}

// HardDelete — 物理删除（FK CASCADE 自动清子表）
func (r *PostgresResidentsRepository) HardDelete(ctx context.Context, hoa string) error {
	chk, err := r.CheckClearable(ctx, hoa)
	if err != nil {
		return err
	}
	if !chk.CanClear {
		return fmt.Errorf("%s", chk.Reason)
	}
	// 先抓 oldPrefix（hook 用），DELETE 后 resident_unit 也被 cascade 删了，查不到
	var old sql.NullString
	_ = r.db.QueryRowContext(ctx,
		`SELECT spatial_prefix::text FROM resident_unit WHERE resident_id=$1::INET AND valid_to IS NULL LIMIT 1`,
		hoa).Scan(&old)

	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM residents WHERE resident_id = $1::INET`, hoa); err != nil {
		return fmt.Errorf("hard delete: %w", err)
	}
	// Hook: 通知 cards reconcile 旧 unit
	if old.Valid && old.String != "" {
		r.fireResidentUnitChange(ctx, narrowPrefixToUnit(old.String))
	}
	return nil
}

// ============================================================================
// helpers
// ============================================================================

type rowScannerLite interface {
	Scan(...any) error
}

func scanResident(rs rowScannerLite) (*domain.Resident, error) {
	var (
		v                                                domain.Resident
		serviceTier, notes                               sql.NullString
		moveIn, moveOut                                  sql.NullString
		familyAccess                                     sql.NullBool
		unitName, roomName, bedName, buildingNum         sql.NullString
		branchName                                       sql.NullString
		unitType                                         sql.NullInt32
		tenantKind                                       sql.NullString
	)
	if err := rs.Scan(&v.ResidentID, &v.TenantID, &v.BranchID, &v.ResidentSlot,
		&v.ResidentAccount, &v.Nickname, &v.Status,
		&serviceTier, &moveIn, &moveOut, &notes,
		&familyAccess,
		&unitName, &roomName, &bedName, &buildingNum,
		&branchName, &unitType, &tenantKind); err != nil {
		return nil, err
	}
	if serviceTier.Valid {
		v.ServiceLevel = &serviceTier.String
	}
	if moveIn.Valid {
		v.AdmissionDate = &moveIn.String
	}
	if moveOut.Valid {
		v.DischargeDate = &moveOut.String
	}
	if notes.Valid {
		v.Note = &notes.String
	}
	if familyAccess.Valid {
		b := familyAccess.Bool
		v.FamilyAccess = &b
	}
	if unitName.Valid {
		v.UnitName = &unitName.String
	}
	if roomName.Valid {
		v.RoomName = &roomName.String
	}
	if bedName.Valid {
		v.BedName = &bedName.String
	}
	if buildingNum.Valid {
		// 直接用 sites.site_name (如 "A"、"B")
		v.BuildingName = &buildingNum.String
	}
	if branchName.Valid {
		v.BranchName = &branchName.String
	}
	if unitType.Valid {
		ut := int(unitType.Int32)
		v.FacilityType = &ut
	}
	if tenantKind.Valid {
		var prop string
		switch strings.ToUpper(tenantKind.String) {
		case "B2C":
			prop = "home"
		default: // B2B 或异常都按 facility 显示
			prop = "facility"
		}
		v.Property = &prop
	}
	return &v, nil
}

func nullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return v
}
func nullInt(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
func pickDeepestPrefix(vals ...*string) string {
	for _, v := range vals {
		if v == nil {
			continue
		}
		s := strings.TrimSpace(*v)
		if s == "" {
			continue
		}
		// 简单校验：必须是 IPv6 prefix（含 ::）
		if !strings.Contains(s, "::") {
			continue
		}
		return s
	}
	return ""
}

// narrowPrefixToUnit — 把任意 INET CIDR 截到 unit /80
//   /48 /56 → "" (跳过，cards 表无这两级 — startup 全 reconcile 自愈)
//   /80 /88 /96 → 截到 /80（含整 unit 下所有 cards）
//   /128 → "" (host 地址不是 unit 锚点)
// Repository hook 用，避免 hook scope 太窄漏 reconcile /80 unit card。
func narrowPrefixToUnit(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return ""
	}
	idx := strings.LastIndex(prefix, "/")
	if idx <= 0 {
		return ""
	}
	switch prefix[idx:] {
	case "/96", "/88", "/80":
		// pass
	default:
		return "" // /48 /56 /128 都跳过
	}
	addr := prefix[:idx]
	// 取前 5 个 segment (fd00:0:T:UU:VV) = 80 bit，后置 "::/80"
	segs := strings.Split(addr, "::")[0]
	parts := strings.Split(segs, ":")
	if len(parts) < 5 {
		return ""
	}
	return strings.Join(parts[:5], ":") + "::/80"
}

// SetPHICryptor — 注入 PHI/contacts 加密器（main.go 启动时调用，从 K 服务派生 tenant key）
func (r *PostgresResidentsRepository) SetPHICryptor(c *phi.PHICryptor) {
	r.cryptor = c
}
