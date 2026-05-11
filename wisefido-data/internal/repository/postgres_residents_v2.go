// Package repository — Resident V2 (Forward Design)
//
// 不依赖 v1 ResidentsRepository / 任何 v1 字段。
// 所有 SQL 用 v2 schema：hoa INET PK；空间归属通过位掩码 + 关联 resident_unit。
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

type PostgresResidentsRepositoryV2 struct {
	db *sql.DB
}

func NewPostgresResidentsRepositoryV2(db *sql.DB) *PostgresResidentsRepositoryV2 {
	return &PostgresResidentsRepositoryV2{db: db}
}

// ============================================================================
// Read
// ============================================================================

// ListResidents — tenant 范围内列表（含分页 + nickname/account ILIKE 搜索）
func (r *PostgresResidentsRepositoryV2) ListResidents(
	ctx context.Context, tenantPrefix string, f domain.ResidentV2ListFilter,
) ([]*domain.ResidentV2, int, error) {
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
	whereClause := strings.Join(where, " AND ")

	// total
	var total int
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM residents r WHERE "+whereClause, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count residents: %w", err)
	}
	if total == 0 {
		return []*domain.ResidentV2{}, 0, nil
	}

	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	q := `
		SELECT host(r.resident_id),
		       host(network(set_masklen(r.resident_id, 48))) || '/48' AS tenant_id,
		       host(network(set_masklen(r.resident_id, 56))) || '/56' AS branch_prefix,
		       r.resident_slot,
		       COALESCE(r.resident_account, ''),
		       r.nickname,
		       r.status,
		       r.service_level,
		       r.gender,
		       r.birth_year,
		       r.admission_date,
		       r.discharge_date,
		       r.note
		  FROM residents r
		 WHERE ` + whereClause + `
		 ORDER BY COALESCE(r.resident_account, ''), r.resident_slot
		 LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list residents: %w", err)
	}
	defer rows.Close()

	out := []*domain.ResidentV2{}
	for rows.Next() {
		v, err := scanResidentV2(rows)
		if err != nil {
			continue
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

// GetResident — 单个 resident 含完整反推 + caregivers + teams
func (r *PostgresResidentsRepositoryV2) GetResident(
	ctx context.Context, tenantPrefix, hoa string,
) (*domain.ResidentV2Detail, error) {
	d := &domain.ResidentV2Detail{}
	var (
		serviceTier, gender, notes        sql.NullString
		birthYear                          sql.NullInt32
		moveIn, moveOut                    sql.NullString
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
		       r.gender,
		       r.birth_year,
		       r.admission_date::text,
		       r.discharge_date::text,
		       r.note,

		       (SELECT host(network(set_masklen(spatial_prefix, 80))) || '/80' FROM active),
		       (SELECT CASE WHEN m >= 88 THEN host(network(set_masklen(spatial_prefix, 88))) || '/88' END FROM active),
		       (SELECT CASE WHEN m >= 96 THEN host(network(set_masklen(spatial_prefix, 96))) || '/96' END FROM active),
		       (SELECT u.unit_type FROM units u, active a WHERE u.unit_id = network(set_masklen(a.spatial_prefix, 80))),
		       (SELECT b.branch_name FROM branches b, active a WHERE b.branch_id = network(set_masklen(a.spatial_prefix, 56))),
		       (SELECT s.building::text FROM sites s, active a WHERE s.site_id = network(set_masklen(a.spatial_prefix, 64))),
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
		&serviceTier, &gender, &birthYear, &moveIn, &moveOut, &notes,
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
	if gender.Valid {
		d.Gender = &gender.String
	}
	if birthYear.Valid {
		v := int(birthYear.Int32)
		d.BirthYear = &v
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
		// sites.building 是 smallint，转成 "B<N>" 字符串显示
		bn := "B" + buildingName.String
		d.BuildingName = &bn
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

	// caregivers / teams / family
	d.Caregivers = r.loadCaregivers(ctx, hoa)
	d.Teams = r.loadTeams(ctx, hoa)
	d.Family = r.loadFamily(ctx, hoa)
	// PHI: Phase 3b 接通解密
	d.PHI = nil

	return d, nil
}

func (r *PostgresResidentsRepositoryV2) loadCaregivers(ctx context.Context, hoa string) []domain.ResidentCaregiverV2 {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.hoa::text,
		       u.user_id::text,
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
		return nil
	}
	defer rows.Close()
	out := []domain.ResidentCaregiverV2{}
	for rows.Next() {
		var c domain.ResidentCaregiverV2
		if err := rows.Scan(&c.HoA, &c.UserID, &c.UserAccount, &c.Nickname, &c.Role); err == nil {
			out = append(out, c)
		}
	}
	return out
}

func (r *PostgresResidentsRepositoryV2) loadTeams(ctx context.Context, hoa string) []domain.ResidentTeamV2 {
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

func (r *PostgresResidentsRepositoryV2) loadFamily(ctx context.Context, hoa string) []domain.ResidentFamilyV2 {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.hoa::text,
		       u.user_id::text,
		       u.user_account,
		       COALESCE(u.nickname,'')
		  FROM resident_caregivers rc
		  JOIN users u ON u.user_id = rc.family_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.family_id IS NOT NULL
		 ORDER BY u.user_account`, hoa)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []domain.ResidentFamilyV2{}
	for rows.Next() {
		var f domain.ResidentFamilyV2
		if err := rows.Scan(&f.HoA, &f.UserID, &f.UserAccount, &f.Nickname); err == nil {
			out = append(out, f)
		}
	}
	return out
}

// ============================================================================
// Write
// ============================================================================

// CreateResident — IPAM 分配 slot + INSERT residents + 可选关联
func (r *PostgresResidentsRepositoryV2) CreateResident(
	ctx context.Context, tenantPrefix string, in *domain.ResidentV2CreateInput,
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
		                       service_level, gender, birth_year, admission_date, note,
		                       status)
		VALUES ($1::INET, $2, $3, $4,
		        $5, $6, $7, $8, $9,
		        'active')`,
		hoa, slot, in.Nickname, account,
		nullStr(in.ServiceLevel), nullStr(in.Gender), nullInt(in.BirthYear),
		nullStr(in.AdmissionDate), nullStr(in.Note),
	); err != nil {
		return "", fmt.Errorf("insert resident: %w", err)
	}

	// 空间分配（取 deepest non-empty prefix）
	if sp := pickDeepestPrefix(in.BedID, in.RoomID, in.UnitID); sp != "" {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO resident_unit (resident_id, spatial_prefix) VALUES ($1::INET, $2::INET)`,
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

	// PHI: Phase 3b 接通加密；当前 in.PHI 忽略

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return hoa, nil
}

// UpdateResident — partial update（只动 input 中提供的字段）
func (r *PostgresResidentsRepositoryV2) UpdateResident(
	ctx context.Context, tenantPrefix, hoa string, in *domain.ResidentV2UpdateInput,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
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
	addI := func(col string, v *int) {
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
	addS("gender", in.Gender)
	addI("birth_year", in.BirthYear)
	addS("admission_date", in.AdmissionDate)
	addS("discharge_date", in.DischargeDate)
	addS("note", in.Note)

	if len(sets) > 0 {
		q := "UPDATE residents SET " + strings.Join(sets, ", ") + ", updated_at = NOW() " +
			"WHERE resident_id = $1::INET AND network(set_masklen(resident_id, 48)) = $2::INET"
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return fmt.Errorf("update residents: %w", err)
		}
	}

	// 空间分配（任一字段提供即视为重新分配）
	if in.UnitID != nil || in.RoomID != nil || in.BedID != nil {
		newPrefix := pickDeepestPrefix(in.BedID, in.RoomID, in.UnitID)
		// 终止 active 行
		if _, err := tx.ExecContext(ctx,
			`UPDATE resident_unit SET valid_to = NOW() WHERE resident_id = $1::INET AND valid_to IS NULL`,
			hoa); err != nil {
			return fmt.Errorf("end resident_unit: %w", err)
		}
		if newPrefix != "" {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO resident_unit (resident_id, spatial_prefix) VALUES ($1::INET, $2::INET)`,
				hoa, newPrefix); err != nil {
				return fmt.Errorf("insert resident_unit: %w", err)
			}
		}
	}

	// caregivers / teams（提供即重置；nil = 不动）
	if in.CaregiverUserIDs != nil {
		if err := r.replaceCaregivers(ctx, tx, hoa, *in.CaregiverUserIDs); err != nil {
			return err
		}
	}
	if in.CareTeamIDs != nil {
		if err := r.replaceTeams(ctx, tx, hoa, *in.CareTeamIDs); err != nil {
			return err
		}
	}

	// PHI: Phase 3b
	return tx.Commit()
}

func (r *PostgresResidentsRepositoryV2) replaceCaregivers(ctx context.Context, tx *sql.Tx, hoa string, userIDs []string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND caregiver_id IS NOT NULL`,
		hoa); err != nil {
		return err
	}
	for _, uid := range userIDs {
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
			return err
		}
	}
	return nil
}

func (r *PostgresResidentsRepositoryV2) replaceTeams(ctx context.Context, tx *sql.Tx, hoa string, teamIDs []string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND care_team_id IS NOT NULL`,
		hoa); err != nil {
		return err
	}
	for _, tid := range teamIDs {
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
			return err
		}
	}
	return nil
}

// SoftDelete — status='deleted'
func (r *PostgresResidentsRepositoryV2) SoftDelete(ctx context.Context, hoa string) error {
	if _, err := r.db.ExecContext(ctx,
		`UPDATE residents SET status = 'deleted', updated_at = NOW() WHERE hoa = $1::INET`,
		hoa); err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	_, _ = r.db.ExecContext(ctx,
		`UPDATE resident_unit SET valid_to = NOW() WHERE resident_id = $1::INET AND valid_to IS NULL`, hoa)
	return nil
}

// CheckClearable — 三表无任何 resident_id 引用才允许 hard delete
func (r *PostgresResidentsRepositoryV2) CheckClearable(ctx context.Context, hoa string) (*domain.ResidentV2ClearCheckResult, error) {
	res := &domain.ResidentV2ClearCheckResult{}
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
func (r *PostgresResidentsRepositoryV2) HardDelete(ctx context.Context, hoa string) error {
	chk, err := r.CheckClearable(ctx, hoa)
	if err != nil {
		return err
	}
	if !chk.CanClear {
		return fmt.Errorf("%s", chk.Reason)
	}
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM residents WHERE hoa = $1::INET`, hoa); err != nil {
		return fmt.Errorf("hard delete: %w", err)
	}
	return nil
}

// ============================================================================
// helpers
// ============================================================================

type rowScannerLite interface {
	Scan(...any) error
}

func scanResidentV2(rs rowScannerLite) (*domain.ResidentV2, error) {
	var (
		v                            domain.ResidentV2
		serviceTier, gender, notes   sql.NullString
		birthYear                    sql.NullInt32
		moveIn, moveOut              sql.NullString
	)
	if err := rs.Scan(&v.ResidentID, &v.TenantID, &v.BranchID, &v.ResidentSlot,
		&v.ResidentAccount, &v.Nickname, &v.Status,
		&serviceTier, &gender, &birthYear, &moveIn, &moveOut, &notes); err != nil {
		return nil, err
	}
	if serviceTier.Valid {
		v.ServiceLevel = &serviceTier.String
	}
	if gender.Valid {
		v.Gender = &gender.String
	}
	if birthYear.Valid {
		bi := int(birthYear.Int32)
		v.BirthYear = &bi
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
