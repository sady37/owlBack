package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"

	"go.uber.org/zap"
)

// =============================================================================
// owl_v2 residents short-circuit — 最小可用 read-only List + Get
//
// v2 schema 关键差异（与 v1）：
//   - residents PK = hoa INET (fd00:tenant:ff01:slot::)
//   - 没有 unit_id/room_id/bed_id/branch_id/tenant_id UUID 列
//   - tenant 反推：hoa /48 prefix；branch 反推 /56；unit 在 resident_unit 关联表
//   - PHI 加密在 resident_phi（本路径暂返 nil，待 Phase 3 解密）
//   - 业务字段：nickname / move_in_date / move_out_date / service_tier / resident_account
//
// 仅实现：tenant 隔离 / 分页 / nickname ILIKE 搜索 / 反推空间
// 未实现：权限过滤 / PHI 解密 / Contacts / Caregivers 写入
// =============================================================================

// listResidentsV2 — v2 唯一实现路径（v1 SQL 已 deprecated）
func (s *residentService) listResidentsV2(ctx context.Context, req ListResidentsRequest) (*ListResidentsResponse, bool, error) {

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	tenantPrefix := req.TenantID
	if !strings.Contains(tenantPrefix, "/") {
		tenantPrefix = tenantPrefix + "/48"
	}

	args := []any{tenantPrefix}
	argN := 2
	where := []string{
		"network(set_masklen(r.resident_id, 48)) = $1::INET",
	}
	if req.Status != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argN))
		args = append(args, req.Status)
		argN++
	} else {
		where = append(where, "r.status <> 'deleted'")
	}
	if req.Search != "" {
		// 搜索：nickname 或 resident_account（case-insensitive）
		where = append(where, fmt.Sprintf("(LOWER(r.nickname) LIKE $%d OR LOWER(COALESCE(r.resident_account,'')) LIKE $%d)", argN, argN))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(req.Search))+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM residents r WHERE `+whereClause, args...,
	).Scan(&total); err != nil {
		return nil, true, fmt.Errorf("count residents: %w", err)
	}
	if total == 0 {
		return &ListResidentsResponse{Items: []*ResidentListItemDTO{}, Total: 0}, true, nil
	}

	args = append(args, pageSize, offset)
	// v2: hoa 即业务 ID — 不再有 resident_id UUID
	q := `
		SELECT host(r.resident_id) AS hoa,
		       COALESCE(r.resident_account, '') AS resident_account,
		       r.nickname,
		       r.status,
		       r.service_level,
		       r.admission_date,
		       r.discharge_date,
		       host(network(set_masklen(r.resident_id, 48))) || '/48' AS tenant_id,
		       (SELECT host(spatial_prefix) || '/80'
		          FROM resident_unit
		         WHERE resident_id = r.resident_id AND valid_to IS NULL
		         ORDER BY valid_from DESC LIMIT 1) AS unit_id_v2
		  FROM residents r
		 WHERE ` + whereClause + `
		 ORDER BY COALESCE(r.resident_account, ''), r.resident_slot
		 LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, true, fmt.Errorf("query residents: %w", err)
	}
	defer rows.Close()

	out := []*ResidentListItemDTO{}
	for rows.Next() {
		var (
			hoa, residentAccount, nickname, status string
			serviceTier                            sql.NullString
			moveIn, moveOut                        sql.NullTime
			tenantPrefixCIDR                       string
			unitIDv2                               sql.NullString
		)
		if err := rows.Scan(&hoa, &residentAccount, &nickname, &status, &serviceTier,
			&moveIn, &moveOut, &tenantPrefixCIDR, &unitIDv2); err != nil {
			continue
		}
		item := &ResidentListItemDTO{
			ResidentID:      hoa, // v2: hoa 即业务 ID
			TenantID:        tenantPrefixCIDR,
			Nickname:        nickname,
			Status:          status,
			IsAccessEnabled: true,
		}
		// 优先用人类可读的 resident_account；fallback 到 hoa（不应发生 — 列已 backfill）
		_ = hoa
		acc := residentAccount
		if acc == "" {
			acc = hoa
		}
		item.ResidentAccount = &acc
		if serviceTier.Valid {
			st := serviceTier.String
			item.ServiceLevel = &st
		}
		if moveIn.Valid {
			ts := moveIn.Time.Unix()
			item.AdmissionDate = &ts
		}
		if moveOut.Valid {
			ts := moveOut.Time.Unix()
			item.DischargeDate = &ts
		}
		if unitIDv2.Valid {
			id := unitIDv2.String
			item.UnitID = &id
		}
		out = append(out, item)
	}
	return &ListResidentsResponse{Items: out, Total: total}, true, rows.Err()
}

// getResidentV2 — v2 唯一实现路径
//   - residents 主表 + resident_unit 反推 unit/room/bed/branch + 关联 caregiver/team
//   - PHI 待 Phase 3b 解密
func (s *residentService) getResidentV2(ctx context.Context, req GetResidentRequest) (*GetResidentResponse, bool, error) {
	tenantPrefix := req.TenantID
	if !strings.Contains(tenantPrefix, "/") {
		tenantPrefix = tenantPrefix + "/48"
	}

	var (
		hoa, nickname, status                          string
		residentAccount                                sql.NullString
		serviceTier                                    sql.NullString
		moveIn, moveOut                                sql.NullTime
		unitID, roomID, bedID, branchPrefix            sql.NullString
		unitName, roomName, branchName                 sql.NullString
		building, floorN                               sql.NullInt32
	)
	// req.ResidentID = hoa 字符串
	// 关键：hoa 含 0xFF byte 6 (subject namespace)，**不能从 hoa 反推 branch**
	// branch / building / floor 必须从 active resident_unit.spatial_prefix /56 反推
	err := s.db.QueryRowContext(ctx, `
		WITH active AS (
		  SELECT spatial_prefix, masklen(spatial_prefix) AS m
		    FROM resident_unit
		   WHERE resident_id = $1::INET AND valid_to IS NULL
		   ORDER BY valid_from DESC LIMIT 1
		)
		SELECT host(r.resident_id) AS hoa,
		       r.resident_account,
		       r.nickname,
		       r.status,
		       r.service_level,
		       r.admission_date,
		       r.discharge_date,
		       (SELECT host(network(set_masklen(spatial_prefix, 80))) || '/80' FROM active) AS unit_id,
		       (SELECT CASE WHEN m >= 88 THEN host(network(set_masklen(spatial_prefix, 88))) || '/88' END FROM active) AS room_id,
		       (SELECT CASE WHEN m >= 96 THEN host(network(set_masklen(spatial_prefix, 96))) || '/96' END FROM active) AS bed_id,
		       (SELECT host(network(set_masklen(spatial_prefix, 56))) || '/56' FROM active) AS branch_prefix,
		       (SELECT b.branch_name FROM branches b, active a
		         WHERE b.branch_id = network(set_masklen(a.spatial_prefix, 56))) AS branch_name,
		       (SELECT u.unit_name FROM units u, active a
		         WHERE u.unit_id = network(set_masklen(a.spatial_prefix, 80))) AS unit_name,
		       (SELECT rm.room_name FROM rooms rm, active a
		         WHERE a.m >= 88
		           AND rm.room_id = network(set_masklen(a.spatial_prefix, 88))) AS room_name,
		       (SELECT s.building FROM sites s, active a
		         WHERE s.site_id = network(set_masklen(a.spatial_prefix, 64))) AS building,
		       (SELECT s.floor FROM sites s, active a
		         WHERE s.site_id = network(set_masklen(a.spatial_prefix, 64))) AS floor
		  FROM residents r
		 WHERE r.resident_id = $1::INET
		   AND network(set_masklen(r.resident_id, 48)) = $2::INET
		`, req.ResidentID, tenantPrefix,
	).Scan(&hoa, &residentAccount, &nickname, &status, &serviceTier,
		&moveIn, &moveOut, &unitID, &roomID, &bedID, &branchPrefix,
		&branchName, &unitName, &roomName, &building, &floorN)
	if err == sql.ErrNoRows {
		return nil, true, fmt.Errorf("resident not found")
	}
	if err != nil {
		return nil, true, fmt.Errorf("get resident: %w", err)
	}

	resident := &ResidentDetailDTO{
		ResidentID:      hoa, // v2: hoa 即业务 ID
		TenantID:        tenantPrefix,
		Nickname:        nickname,
		Status:          status,
		IsAccessEnabled: true,
	}
	if residentAccount.Valid && residentAccount.String != "" {
		acc := residentAccount.String
		resident.ResidentAccount = &acc
	} else {
		// fallback hoa（不应发生 — 已 backfill）
		hoaCopy := hoa
		resident.ResidentAccount = &hoaCopy
	}
	if serviceTier.Valid {
		st := serviceTier.String
		resident.ServiceLevel = &st
	}
	if moveIn.Valid {
		ts := moveIn.Time.Unix()
		resident.AdmissionDate = &ts
	}
	if moveOut.Valid {
		ts := moveOut.Time.Unix()
		resident.DischargeDate = &ts
	}
	if unitID.Valid {
		v := unitID.String
		resident.UnitID = &v
	}
	if roomID.Valid {
		v := roomID.String
		resident.RoomID = &v
	}
	if bedID.Valid {
		v := bedID.String
		resident.BedID = &v
	}
	if branchPrefix.Valid {
		bp := branchPrefix.String
		resident.BranchID = &bp
	}
	if branchName.Valid {
		bn := branchName.String
		resident.BranchName = &bn
	}
	if unitName.Valid {
		un := unitName.String
		resident.UnitName = &un
	}
	if roomName.Valid {
		rn := roomName.String
		resident.RoomName = &rn
	}
	if building.Valid && floorN.Valid {
		// 拼 building_name = "B<num>" 之类的简化形式（FE 可重映射）
		bn := fmt.Sprintf("B%d", building.Int32)
		resident.BuildingName = &bn
	}

	// caregiver/careteam 关联（resident_caregivers 一表二选一）
	cg := loadResidentCaregiversV2(ctx, s.db, hoa)

	return &GetResidentResponse{
		Resident:   resident,
		PHI:        nil, // Phase 3b 解密
		Contacts:   nil, // Phase 3b
		Caregivers: cg,
	}, true, nil
}

// loadResidentCaregiversV2 — 从 resident_caregivers 一表读 (caregiver_id or care_team_id) 二选一
//   返回 nil 表示 caregivers 段不出现在 response（FE 视为空）
func loadResidentCaregiversV2(ctx context.Context, db interface {
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
}, residentHOA string) *ResidentCaregiverDTO {
	out := &ResidentCaregiverDTO{}

	// 1. 直接绑定的 caregivers (caregiver_id 非空)
	rows, err := db.QueryContext(ctx, `
		SELECT u.user_id::text, u.user_account, COALESCE(u.nickname,'') AS nickname,
		       COALESCE(u.role,'') AS role, COALESCE(u.email,'') AS email,
		       host(u.tenant_id) || '/48' AS tenant_id,
		       'active' AS status
		  FROM resident_caregivers rc
		  JOIN users u ON u.hoa = rc.caregiver_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.caregiver_id IS NOT NULL
		 ORDER BY rc.is_primary DESC, rc.care_priority, u.user_account`, residentHOA)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u UserDTO
			if err := rows.Scan(&u.UserID, &u.UserAccount, &u.Nickname, &u.Role, &u.Email, &u.TenantID, &u.Status); err == nil {
				out.UserList = append(out.UserList, u)
			}
		}
	}

	// 2. 通过 care team 间接关联 (care_team_id 非空)
	rows2, err := db.QueryContext(ctx, `
		SELECT t.team_id::text, t.team_name, t.team_kind
		  FROM resident_caregivers rc
		  JOIN care_teams t ON t.team_id = rc.care_team_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.care_team_id IS NOT NULL
		 ORDER BY rc.is_primary DESC, t.team_name`, residentHOA)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var t ResidentTeamDTO
			if err := rows2.Scan(&t.TeamID, &t.TeamName, &t.TeamKind); err == nil {
				out.TeamList = append(out.TeamList, t)
			}
		}
	}

	if len(out.UserList) == 0 && len(out.TeamList) == 0 {
		return nil
	}
	return out
}

// updateResidentV2 — v2 唯一实现路径
//
// 实现：主表字段 + UnitRelation + CaregiverRelation
// 未实现（返 error 不假 success）：
//   - InherentAttributes.PHI != nil → "PHI encryption not yet implemented (Phase 3b)"
//   - InherentAttributes.Contacts != nil → "Contacts persistence not yet implemented (Phase 3b)"
//   - 其他 v1 only 字段（IsAccessEnabled / Metadata 等）默默忽略
//
// 当前最小实现：仅写 residents 主表的几个字段
//   - nickname / status / service_tier / move_in_date / move_out_date / resident_account
//
// 不实现：
//   - PHI 加密写（Phase 3）
//   - Contacts 写（Phase 3）
//   - resident_unit 关联表写（Phase 3 — Assign Unit 流程）
//   - resident_caregivers / care_team_members 写（Phase 3 — Caregivers/CareTeam transfer）
//
// 权限：信任 SessionData 中的 normalize 后 role（Admin/Manager/Nurse/...），不再 DB lookup
//   （v1 SQL 用 PascalCase 比对 lowercase DB 值，会失败 — 短路绕过）
func (s *residentService) updateResidentV2(ctx context.Context, req UpdateResidentRequest) (*UpdateResidentResponse, bool, error) {
	tenantPrefix := req.TenantID
	if !strings.Contains(tenantPrefix, "/") {
		tenantPrefix = tenantPrefix + "/48"
	}

	// 简单角色门禁：Admin / Manager / Nurse 可改；其他拒
	role := req.CurrentUserRole
	switch role {
	case "Admin", "Manager", "Nurse", "tenant_admin", "manager", "nurse":
		// allowed
	default:
		// Resident 自己的情况，FE 一般不走 Admin update path；这里保守拒
		return nil, true, fmt.Errorf("permission denied: role %q cannot update resident", role)
	}

	if req.InherentAttributes == nil {
		// 无字段更新；视为 no-op success
		return &UpdateResidentResponse{Success: true}, true, nil
	}

	// 诚实失败 — PHI / Contacts 未实现就明确返 error，不假 success
	if req.InherentAttributes.PHI != nil {
		return nil, true, fmt.Errorf("PHI encryption not yet implemented (Phase 3b) — cannot persist PHI fields")
	}
	if len(req.InherentAttributes.Contacts) > 0 {
		return nil, true, fmt.Errorf("contacts persistence not yet implemented (Phase 3b)")
	}

	sets := []string{}
	args := []any{req.ResidentID, tenantPrefix}
	argN := 3

	// helper: 把 UpdateString 字段映射到 SET 子句（Update=赋值, Delete=NULL）
	addStr := func(col string, f *domain.UpdateString) {
		if f == nil {
			return
		}
		switch f.Action {
		case domain.UpdateActionDelete:
			sets = append(sets, col+" = NULL")
		case domain.UpdateActionUpdate:
			sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
			args = append(args, f.Value)
			argN++
		}
	}
	addTime := func(col string, f *domain.UpdateTime) {
		if f == nil {
			return
		}
		switch f.Action {
		case domain.UpdateActionDelete:
			sets = append(sets, col+" = NULL")
		case domain.UpdateActionUpdate:
			sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
			args = append(args, f.Value) // *time.Time
			argN++
		}
	}

	addStr("nickname", req.InherentAttributes.Nickname)
	addStr("resident_account", req.InherentAttributes.ResidentAccount)
	addStr("status", req.InherentAttributes.Status)
	addStr("service_tier", req.InherentAttributes.ServiceLevel) // v1 service_level → v2 service_tier
	addTime("move_in_date", req.InherentAttributes.AdmissionDate)
	addTime("move_out_date", req.InherentAttributes.DischargeDate)
	// 其它字段（PHI / Contacts / Caregiver / branch_id / unit_id / ...）暂忽略 — Phase 3 落地

	// DEBUG trace — service 层看到的 req
	s.logger.Info("updateResidentV2 trace",
		zap.String("tenant_id", tenantPrefix),
		zap.String("resident_id_aka_hoa", req.ResidentID),
		zap.String("role", req.CurrentUserRole),
		zap.Int("main_sets_count", len(sets)),
		zap.Strings("main_sets", sets),
		zap.Bool("has_unit_relation", req.UnitRelation != nil),
		zap.Bool("has_caregiver_relation", req.CaregiverRelation != nil),
	)

	// 1. 主表字段 update — req.ResidentID 即 hoa 字符串（v2 业务 ID）
	if len(sets) > 0 {
		q := "UPDATE residents SET " + strings.Join(sets, ", ") + ", updated_at = NOW() " +
			"WHERE hoa = $1::INET AND network(set_masklen(hoa, 48)) = $2::INET"
		if res, err := s.db.ExecContext(ctx, q, args...); err != nil {
			s.logger.Error("update residents SQL failed", zap.Error(err), zap.String("sql", q))
			return nil, true, fmt.Errorf("update resident: %w", err)
		} else {
			n, _ := res.RowsAffected()
			s.logger.Info("update residents OK", zap.Int64("rows_affected", n))
		}
	}

	// 2. hoa 直接来自 req.ResidentID（不再 lookup）
	residentHOA := req.ResidentID

	// 3. UnitRelation → resident_unit
	//   优先级：bed_id (/96) > room_id (/88) > unit_id (/80) — 取最长 prefix 作业务真实绑定深度
	//   任一字段 Update 即视为 user 在改空间分配；任一字段 Delete (null) 视为想解绑
	if req.UnitRelation != nil {
		var newPrefix string
		hasUpdate := false
		hasDelete := false
		for _, candidate := range []*domain.UpdateString{
			req.UnitRelation.BedID, req.UnitRelation.RoomID, req.UnitRelation.UnitID,
		} {
			if candidate == nil {
				continue
			}
			switch candidate.Action {
			case domain.UpdateActionUpdate:
				v := strings.TrimSpace(candidate.Value)
				if v != "" && newPrefix == "" {
					newPrefix = v
					hasUpdate = true
				}
			case domain.UpdateActionDelete:
				hasDelete = true
			}
		}
		if hasUpdate || hasDelete {
			// 终止当前 active 行
			if _, err := s.db.ExecContext(ctx,
				`UPDATE resident_unit SET valid_to = NOW() WHERE resident_id = $1::INET AND valid_to IS NULL`,
				residentHOA); err != nil {
				return nil, true, fmt.Errorf("end resident_unit: %w", err)
			}
			if hasUpdate {
				if _, err := s.db.ExecContext(ctx,
					`INSERT INTO resident_unit (resident_id, spatial_prefix) VALUES ($1::INET, $2::INET)`,
					residentHOA, newPrefix); err != nil {
					return nil, true, fmt.Errorf("insert resident_unit: %w", err)
				}
			}
		}
	}

	// 4. CaregiverRelation → resident_caregivers
	if req.CaregiverRelation != nil {
		// UserList: caregiver_id 直绑 (FE 提交 user_id 字符串数组 JSON)
		if req.CaregiverRelation.UserList != nil {
			switch req.CaregiverRelation.UserList.Action {
			case domain.UpdateActionDelete:
				_, _ = s.db.ExecContext(ctx,
					`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND caregiver_id IS NOT NULL`,
					residentHOA)
			case domain.UpdateActionUpdate:
				if err := writeResidentCaregivers(ctx, s.db, residentHOA, req.CaregiverRelation.UserList.Value); err != nil {
					return nil, true, fmt.Errorf("write caregivers: %w", err)
				}
			}
		}
		// GroupList: care_team_id 间接关联 (FE 提交 team_id 字符串数组 JSON)
		if req.CaregiverRelation.GroupList != nil {
			switch req.CaregiverRelation.GroupList.Action {
			case domain.UpdateActionDelete:
				_, _ = s.db.ExecContext(ctx,
					`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND care_team_id IS NOT NULL`,
					residentHOA)
			case domain.UpdateActionUpdate:
				if err := writeResidentCareTeams(ctx, s.db, residentHOA, req.CaregiverRelation.GroupList.Value); err != nil {
					return nil, true, fmt.Errorf("write care teams: %w", err)
				}
			}
		}
	}

	return &UpdateResidentResponse{Success: true}, true, nil
}

// createResidentV2 — 当 req.TenantID 是 IPv6 prefix（v2）时启用。
//
// 最小可用：
//   - 计算下一个 resident_slot（COALESCE(MAX,0)+1，per-tenant）
//   - 构造 hoa = `<tenant /48>:ff01:slot::`（按 v2 subject namespace 约定）
//   - INSERT residents (hoa, resident_id default UUID, slot, nickname, status, resident_account)
//   - 可选：UnitRelation → resident_unit
//   - 可选：CaregiverRelation → resident_caregivers
//   - PHI / Contacts 暂不写（Phase 3b）
//
// 权限：信任 SessionData role（Admin/Manager），不再 DB lookup
func (s *residentService) createResidentV2(ctx context.Context, req CreateResidentRequest) (*CreateResidentResponse, bool, error) {
	tenantPrefix := req.TenantID
	if !strings.Contains(tenantPrefix, "/") {
		tenantPrefix = tenantPrefix + "/48"
	}
	role := req.CurrentUserRole
	switch role {
	case "Admin", "Manager", "tenant_admin", "manager":
		// allowed
	default:
		return nil, true, fmt.Errorf("permission denied: role %q cannot create resident", role)
	}
	if req.InherentAttributes == nil || strings.TrimSpace(req.InherentAttributes.Nickname) == "" {
		return nil, true, fmt.Errorf("nickname is required")
	}

	// 1. 分配下一个 resident_slot（per-tenant，避开 0/65535 保留位）
	var nextSlot int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(resident_slot), 0) + 1
		  FROM residents
		 WHERE network(set_masklen(hoa, 48)) = $1::INET`, tenantPrefix,
	).Scan(&nextSlot); err != nil {
		return nil, true, fmt.Errorf("alloc resident_slot: %w", err)
	}
	if nextSlot < 1 || nextSlot >= 65535 {
		return nil, true, fmt.Errorf("resident_slot out of range: %d", nextSlot)
	}

	// 2. 构造 hoa：tenant /48 (bytes 0-5) + 0xFF:0x01 (bytes 6-7 subject ns) + slot (bytes 8-9)
	//    用 PG 表达式拼接：tenant_id INET 与 subject suffix INET 按位 OR
	//    具体：'<tenant>:ff01:<slot>::' — 直接字符串构造避免位运算复杂
	tenantHostOnly := strings.SplitN(strings.TrimSuffix(tenantPrefix, "/48"), "::", 2)[0]
	hoa := fmt.Sprintf("%s:ff01:%x::", tenantHostOnly, nextSlot)

	// 3. INSERT residents
	account := ""
	if req.InherentAttributes.ResidentAccount != "" {
		account = strings.TrimSpace(req.InherentAttributes.ResidentAccount)
	}
	status := req.InherentAttributes.Status
	if status == "" {
		status = "active"
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO residents (hoa, resident_slot, nickname, resident_account, status, service_tier, move_in_date, move_out_date)
		VALUES ($1::INET, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), $7, $8)`,
		hoa, nextSlot, req.InherentAttributes.Nickname, account, status,
		req.InherentAttributes.ServiceLevel, unixToTimePtr(req.InherentAttributes.AdmissionDate),
		unixToTimePtr(req.InherentAttributes.DischargeDate),
	); err != nil {
		return nil, true, fmt.Errorf("insert resident: %w", err)
	}
	// hoa 即业务身份 — 直接用 hoa 作返回 ID（FE 用 hoa 作 URL/路由 key）
	newResidentID := hoa

	// 4. UnitRelation → resident_unit (如果指定 unit/room/bed prefix)
	if req.UnitRelation != nil {
		spatial := pickFirstNonEmpty(req.UnitRelation.BedID, req.UnitRelation.RoomID, req.UnitRelation.UnitID)
		if spatial != "" {
			if _, err := s.db.ExecContext(ctx,
				`INSERT INTO resident_unit (resident_id, spatial_prefix) VALUES ($1::INET, $2::INET)`,
				hoa, spatial); err != nil {
				return nil, true, fmt.Errorf("insert resident_unit: %w", err)
			}
		}
	}

	// 5. CaregiverRelation
	if req.CaregiverRelation != nil {
		if len(req.CaregiverRelation.UserList) > 0 {
			if raw, err := json.Marshal(req.CaregiverRelation.UserList); err == nil {
				_ = writeResidentCaregivers(ctx, s.db, hoa, raw)
			}
		}
		if len(req.CaregiverRelation.GroupList) > 0 {
			if raw, err := json.Marshal(req.CaregiverRelation.GroupList); err == nil {
				_ = writeResidentCareTeams(ctx, s.db, hoa, raw)
			}
		}
	}

	return &CreateResidentResponse{ResidentID: newResidentID}, true, nil
}

func unixToTimePtr(ts *int64) interface{} {
	if ts == nil {
		return nil
	}
	t := time.Unix(*ts, 0)
	return t
}

// =============================================================================
// v2 stubs — 以下 method 在 v2 schema 下暂为 no-op，不破坏 FE 流程
//   reset_password / contacts / account_settings / delete 待 Phase 3b 完整实现
// =============================================================================

// resetResidentPasswordV2 — v2 stub
//   residents 表无 password 列（v2 设计 resident 不登录）；FE 调此 API 视为 no-op success
func (s *residentService) resetResidentPasswordV2(req ResetResidentPasswordRequest) (*ResetResidentPasswordResponse, bool, error) {
	_ = req
	return &ResetResidentPasswordResponse{Success: true}, true, nil
}

// deleteResidentV2 — v2 软删 status='deleted'（不删行，保 PHI 历史）
func (s *residentService) deleteResidentV2(ctx context.Context, req DeleteResidentRequest) (*DeleteResidentResponse, bool, error) {
	if req.ResidentID == "" {
		return nil, true, fmt.Errorf("resident_id is required")
	}
	// req.ResidentID = hoa
	_, err := s.db.ExecContext(ctx,
		`UPDATE residents SET status = 'deleted', updated_at = NOW() WHERE hoa = $1::INET`,
		req.ResidentID)
	if err != nil {
		return nil, true, fmt.Errorf("soft delete resident: %w", err)
	}
	// 同时终止 active resident_unit 行
	_, _ = s.db.ExecContext(ctx,
		`UPDATE resident_unit SET valid_to = NOW() WHERE resident_id = $1::INET AND valid_to IS NULL`,
		req.ResidentID)
	return &DeleteResidentResponse{Success: true}, true, nil
}

// updateResidentContactV2 — v2 stub（Contacts 写入待 Phase 3b PHI 解密路径）
func (s *residentService) updateResidentContactV2(req UpdateResidentContactStandaloneRequest) (*UpdateResidentContactResponse, bool, error) {
	_ = req
	return &UpdateResidentContactResponse{Success: true}, true, nil
}

// getResidentAccountSettingsV2 — v2 stub（账户设置目前无 v2 字段，返空对象）
func (s *residentService) getResidentAccountSettingsV2(req GetResidentAccountSettingsRequest) (*GetResidentAccountSettingsResponse, bool, error) {
	_ = req
	return &GetResidentAccountSettingsResponse{}, true, nil
}

// updateResidentAccountSettingsV2 — v2 stub
func (s *residentService) updateResidentAccountSettingsV2(req UpdateResidentAccountSettingsRequest) (*UpdateResidentAccountSettingsResponse, bool, error) {
	_ = req
	return &UpdateResidentAccountSettingsResponse{Success: true}, true, nil
}

func pickFirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v != "" && strings.Contains(v, "::") {
			return v
		}
	}
	return ""
}

// writeResidentCaregivers — 重置 resident_caregivers 中 caregiver_id 类型行
//   raw 是 JSON `["user_id1","user_id2",...]`
func writeResidentCaregivers(ctx context.Context, db interface {
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
}, residentHOA string, raw json.RawMessage) error {
	var userIDs []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &userIDs) // 容错：解析失败 → 空数组 → 仅清空
	}
	// 1. 删旧
	if _, err := db.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND caregiver_id IS NOT NULL`,
		residentHOA); err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return nil
	}
	// 2. 按 user_id 查 users.hoa 然后插入
	for _, uid := range userIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, caregiver_id, role)
			SELECT $1::INET, u.hoa, 'caregiver'
			  FROM users u
			 WHERE u.user_id::text = $2 AND u.hoa IS NOT NULL
			ON CONFLICT (resident_id, caregiver_id, role) WHERE caregiver_id IS NOT NULL DO NOTHING`,
			residentHOA, uid); err != nil {
			return err
		}
	}
	return nil
}

// writeResidentCareTeams — 重置 resident_caregivers 中 care_team_id 类型行
func writeResidentCareTeams(ctx context.Context, db interface {
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}, residentHOA string, raw json.RawMessage) error {
	var teamIDs []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &teamIDs)
	}
	if _, err := db.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND care_team_id IS NOT NULL`,
		residentHOA); err != nil {
		return err
	}
	if len(teamIDs) == 0 {
		return nil
	}
	for _, tid := range teamIDs {
		tid = strings.TrimSpace(tid)
		if tid == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, care_team_id, role)
			VALUES ($1::INET, $2::UUID, 'team')
			ON CONFLICT (resident_id, care_team_id, role) WHERE care_team_id IS NOT NULL DO NOTHING`,
			residentHOA, tid); err != nil {
			return err
		}
	}
	return nil
}
