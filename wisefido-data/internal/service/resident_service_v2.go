package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
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
		// 自动识别搜索类型
		searchType := ClassifySearch(req.Search)
		searchLower := strings.ToLower(strings.TrimSpace(req.Search))

		switch searchType {
		case SearchTypeEmail, SearchTypePhone:
			// Email/Phone 搜索：需要查询 resident_phi 表并解密（Phase 3b）
			phiResults, err := s.searchResidentsByPHI(ctx, tenantPrefix, searchLower, searchType, req.PermissionCheck, req.CurrentUserID, pageSize)
			if err != nil {
				s.logger.Warn("PHI search failed",
					zap.String("search_type", GetSearchTypeDescription(searchType)),
					zap.Error(err))
				// 降级：返回空结果而不是失败
				return &ListResidentsResponse{Items: []*ResidentListItemDTO{}, Total: 0}, true, nil
			}
			// 应用分页
			start := (page - 1) * pageSize
			end := start + pageSize
			if start >= len(phiResults) {
				return &ListResidentsResponse{Items: []*ResidentListItemDTO{}, Total: len(phiResults)}, true, nil
			}
			if end > len(phiResults) {
				end = len(phiResults)
			}
			return &ListResidentsResponse{Items: phiResults[start:end], Total: len(phiResults)}, true, nil

		case SearchTypeAccount:
			// Account 搜索：residents.resident_account (case-insensitive)
			where = append(where, fmt.Sprintf("LOWER(COALESCE(r.resident_account,'')) LIKE $%d", argN))
			args = append(args, "%"+searchLower+"%")
			argN++

		case SearchTypeNickname:
			// Nickname 搜索：residents.nickname (case-insensitive)
			where = append(where, fmt.Sprintf("LOWER(r.nickname) LIKE $%d", argN))
			args = append(args, "%"+searchLower+"%")
			argN++

		default:
			// 未知类型：同时搜索 nickname 和 account
			where = append(where, fmt.Sprintf("(LOWER(r.nickname) LIKE $%d OR LOWER(COALESCE(r.resident_account,'')) LIKE $%d)", argN, argN))
			args = append(args, "%"+searchLower+"%")
			argN++
		}
	}

	// 权限过滤：assigned_only
	if req.PermissionCheck != nil && req.PermissionCheck.AssignedOnly && req.CurrentUserID != "" {
		args = append(args, req.CurrentUserID)
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM resident_caregivers rc
			WHERE rc.resident_id = r.resident_id
			  AND rc.caregiver_id = $%d::UUID
		)`, argN))
		argN++
	}

	// 权限过滤：branch_only（通过 resident_unit 的 spatial_prefix /56 掩码）
	if req.PermissionCheck != nil && req.PermissionCheck.BranchOnly && req.CurrentUserID != "" {
		userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, req.TenantID, req.CurrentUserID)
		if err != nil {
			return nil, true, fmt.Errorf("failed to get user branch IDs: %w", err)
		}

		if !hasBranches {
			// 用户无关联院区：仅能查看 /56 为 NULL（跨院区过渡态）的住户
			where = append(where, `EXISTS (
				SELECT 1 FROM resident_unit ru
				WHERE ru.resident_id = r.resident_id
				  AND ru.valid_to IS NULL
				  AND masklen(ru.spatial_prefix) <= 56
			)`)
		} else if len(userBranchIDs) == 1 {
			// 用户属于 1 个院区：查看该院区的住户
			args = append(args, userBranchIDs[0])
			where = append(where, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM resident_unit ru
				WHERE ru.resident_id = r.resident_id
				  AND ru.valid_to IS NULL
				  AND host(network(set_masklen(ru.spatial_prefix, 56))) = $%d
			)`, argN))
			argN++
		} else {
			// 用户属于多个院区：查看所有关联院区的住户
			placeholders := make([]string, len(userBranchIDs))
			for i, branchID := range userBranchIDs {
				args = append(args, branchID)
				placeholders[i] = fmt.Sprintf("$%d", len(args))
			}
			where = append(where, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM resident_unit ru
				WHERE ru.resident_id = r.resident_id
				  AND ru.valid_to IS NULL
				  AND host(network(set_masklen(ru.spatial_prefix, 56))) IN (%s)
			)`, strings.Join(placeholders, ", ")))
		}
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
			FamilyAccess: true,
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
		familyAccess                                   bool
		note                                           sql.NullString
		gender                                         sql.NullString
		birthYear                                      sql.NullInt32
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
		       r.family_access,
		       r.note,
		       r.gender,
		       r.birth_year,
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
		&moveIn, &moveOut, &familyAccess, &note, &gender, &birthYear,
		&unitID, &roomID, &bedID, &branchPrefix,
		&branchName, &unitName, &roomName, &building, &floorN)
	if err == sql.ErrNoRows {
		return nil, true, fmt.Errorf("resident not found")
	}
	if err != nil {
		return nil, true, fmt.Errorf("get resident: %w", err)
	}

	resident := &ResidentDetailDTO{
		ResidentID:   hoa, // v2: hoa 即业务 ID
		TenantID:     tenantPrefix,
		Nickname:     nickname,
		Status:       status,
		FamilyAccess: familyAccess,
	}
	if note.Valid {
		nv := note.String
		resident.Note = &nv
	}
	if gender.Valid {
		gv := gender.String
		resident.Gender = &gv
	}
	if birthYear.Valid {
		by := int(birthYear.Int32)
		resident.BirthYear = &by
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

	// PHI / Contacts 解密（v2 PHICryptor）
	var phiDTO *ResidentPHIDTO
	var contactsDTO []*ResidentContactDTO
	if s.phiCryptor != nil {
		if phiData, perr := s.loadResidentPHIv2(ctx, hoa, tenantPrefix); perr == nil && phiData != nil {
			phiDTO = mapPHIDomainToDTO(phiData)
		}
		if list, cerr := s.loadResidentContactsV2(ctx, hoa, tenantPrefix); cerr == nil && len(list) > 0 {
			contactsDTO = mapContactsDomainToDTO(list)
		}
	}

	return &GetResidentResponse{
		Resident:   resident,
		PHI:        phiDTO,
		Contacts:   contactsDTO,
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
		  JOIN users u ON u.user_id = rc.caregiver_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.caregiver_id IS NOT NULL
		 ORDER BY u.user_account`, residentHOA)
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
		 ORDER BY t.team_name`, residentHOA)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var t ResidentTeamDTO
			if err := rows2.Scan(&t.TeamID, &t.TeamName, &t.TeamKind); err == nil {
				out.TeamList = append(out.TeamList, t)
			}
		}
	}

	// 3. 家属绑定 (family_id 非空, users.role=Family)
	rows3, err := db.QueryContext(ctx, `
		SELECT u.user_id::text, u.user_account, COALESCE(u.nickname,'') AS nickname,
		       COALESCE(u.role,'') AS role, COALESCE(u.email,'') AS email,
		       host(u.tenant_id) || '/48' AS tenant_id,
		       'active' AS status
		  FROM resident_caregivers rc
		  JOIN users u ON u.user_id = rc.family_id
		 WHERE rc.resident_id = $1::INET
		   AND rc.valid_to IS NULL
		   AND rc.family_id IS NOT NULL
		 ORDER BY u.user_account`, residentHOA)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var u UserDTO
			if err := rows3.Scan(&u.UserID, &u.UserAccount, &u.Nickname, &u.Role, &u.Email, &u.TenantID, &u.Status); err == nil {
				out.FamilyList = append(out.FamilyList, u)
			}
		}
	}

	if len(out.UserList) == 0 && len(out.TeamList) == 0 && len(out.FamilyList) == 0 {
		// Always return object (never nil) to allow frontend to safely check fields
		return out
	}
	return out
}

// updateResidentV2 — v2 唯一实现路径
//
// 实现：主表字段 + UnitRelation + CaregiverRelation
// 未实现（返 error 不假 success）：
//   - InherentAttributes.PHI != nil → "PHI encryption not yet implemented (Phase 3b)"
//   - InherentAttributes.Contacts != nil → "Contacts persistence not yet implemented (Phase 3b)"
//   - 其他 v1 only 字段（FamilyAccess / Metadata 等）默默忽略
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

	if req.InherentAttributes == nil && req.UnitRelation == nil && req.CaregiverRelation == nil {
		// 无字段更新；视为 no-op success
		return &UpdateResidentResponse{Success: true}, true, nil
	}

	// PHI / Contacts: 加密落库（v2 schema 字段名 1:1 对应）
	if req.InherentAttributes != nil {
		if req.InherentAttributes.PHI != nil {
			phiInput := mapPHIRequestToDomain(req.InherentAttributes.PHI)
			if err := s.writeResidentPHIv2(ctx, req.ResidentID, tenantPrefix, phiInput); err != nil {
				return nil, true, fmt.Errorf("write resident_phi: %w", err)
			}
		}
		if len(req.InherentAttributes.Contacts) > 0 {
			contactsInput := mapContactsRequestToDomain(req.InherentAttributes.Contacts)
			if err := s.writeResidentContactsV2(ctx, req.ResidentID, tenantPrefix, contactsInput); err != nil {
				return nil, true, fmt.Errorf("write resident_contacts: %w", err)
			}
		}
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

	addBool := func(col string, f *domain.UpdateBool) {
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

	addStr("nickname", nil)
	addStr("resident_account", nil)
	addStr("status", nil)
	addStr("service_level", nil)
	addTime("admission_date", nil)
	addTime("discharge_date", nil)
	addStr("note", nil)
	addBool("family_access", nil)

	// 只在 InherentAttributes 存在时才处理主表字段
	if req.InherentAttributes != nil {
		addStr("nickname", req.InherentAttributes.Nickname)
		addStr("resident_account", req.InherentAttributes.ResidentAccount)
		addStr("status", req.InherentAttributes.Status)
		addStr("service_level", req.InherentAttributes.ServiceLevel)
		addTime("admission_date", req.InherentAttributes.AdmissionDate)
		addTime("discharge_date", req.InherentAttributes.DischargeDate)
		addStr("note", req.InherentAttributes.Note)
		addBool("family_access", req.InherentAttributes.FamilyAccess)
	}

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

	// 1. 主表字段 update — req.ResidentID 即 resident_id (hoa) 字符串
	if len(sets) > 0 {
		q := "UPDATE residents SET " + strings.Join(sets, ", ") + ", updated_at = NOW() " +
			"WHERE resident_id = $1::INET AND network(set_masklen(resident_id, 48)) = $2::INET"
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
		fmt.Printf("DEBUG: CaregiverRelation received: UserList=%v, GroupList=%v, FamilyList=%v\n",
			req.CaregiverRelation.UserList != nil, req.CaregiverRelation.GroupList != nil, req.CaregiverRelation.FamilyList != nil)
		// UserList: caregiver_id 直绑 (FE 提交 user_id 字符串数组 JSON)
		if req.CaregiverRelation.UserList != nil {
			switch req.CaregiverRelation.UserList.Action {
			case domain.UpdateActionDelete:
				fmt.Printf("DEBUG: deleting caregivers\n")
				_, _ = s.db.ExecContext(ctx,
					`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND caregiver_id IS NOT NULL`,
					residentHOA)
			case domain.UpdateActionUpdate:
				fmt.Printf("DEBUG: updating caregivers, Value=%s\n", string(req.CaregiverRelation.UserList.Value))
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
		// FamilyList: family_id 行 (FE 提交 user_id 字符串数组 JSON，必须是 role=Family)
		if req.CaregiverRelation.FamilyList != nil {
			switch req.CaregiverRelation.FamilyList.Action {
			case domain.UpdateActionDelete:
				_, _ = s.db.ExecContext(ctx,
					`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND family_id IS NOT NULL`,
					residentHOA)
			case domain.UpdateActionUpdate:
				if err := writeResidentFamily(ctx, s.db, residentHOA, req.CaregiverRelation.FamilyList.Value); err != nil {
					return nil, true, fmt.Errorf("write family: %w", err)
				}
			}
		}
	} else {
		fmt.Printf("DEBUG: CaregiverRelation is nil\n")
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
	familyAccess := req.InherentAttributes.FamilyAccess // Create input: bool 直接来自 handler，默认值由 handler 决定

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO residents (resident_id, resident_slot, nickname, resident_account, status, service_level, admission_date, discharge_date, family_access)
		VALUES ($1::INET, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), $7, $8, $9)`,
		hoa, nextSlot, req.InherentAttributes.Nickname, account, status,
		req.InherentAttributes.ServiceLevel, unixToTimePtr(req.InherentAttributes.AdmissionDate),
		unixToTimePtr(req.InherentAttributes.DischargeDate), familyAccess,
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

	// 5. CaregiverRelation (caregiver/team) + Family
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
		if len(req.CaregiverRelation.FamilyList) > 0 {
			if raw, err := json.Marshal(req.CaregiverRelation.FamilyList); err == nil {
				_ = writeResidentFamily(ctx, s.db, hoa, raw)
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
		if err := json.Unmarshal(raw, &userIDs); err != nil {
			fmt.Printf("DEBUG: failed to unmarshal caregivers JSON: %v, raw=%s\n", err, string(raw))
			return nil // 容错：解析失败 → 空数组 → 仅清空
		}
	}
	fmt.Printf("DEBUG: writeResidentCaregivers called: residentHOA=%s, userIDs=%v\n", residentHOA, userIDs)

	// 1. 删旧
	if _, err := db.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND caregiver_id IS NOT NULL`,
		residentHOA); err != nil {
		fmt.Printf("DEBUG: delete failed: %v\n", err)
		return err
	}
	if len(userIDs) == 0 {
		fmt.Printf("DEBUG: userIDs is empty, returning\n")
		return nil
	}
	// 2. 按 user_id 直接插入（不再依赖 hoa，改为生成 namespace IPv6 地址）
	for _, uid := range userIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		// v2: caregiver_id 存的是 user_id (UUID)，直接参考 users.user_id
		fmt.Printf("DEBUG: inserting caregiver: residentHOA=%s, uid=%s\n", residentHOA, uid)
		if result, err := db.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, caregiver_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, caregiver_id) WHERE caregiver_id IS NOT NULL DO NOTHING`,
			residentHOA, uid); err != nil {
			fmt.Printf("DEBUG: insert failed: %v\n", err)
			return fmt.Errorf("insert caregiver %s: %w", uid, err)
		} else {
			n, _ := result.RowsAffected()
			fmt.Printf("DEBUG: insert success: rows affected=%d\n", n)
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
			INSERT INTO resident_caregivers (resident_id, care_team_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, care_team_id) WHERE care_team_id IS NOT NULL DO NOTHING`,
			residentHOA, tid); err != nil {
			return err
		}
	}
	return nil
}

// writeResidentFamily — 重置 resident_caregivers 中 family_id 类型行
//   raw 是 JSON `["user_id1","user_id2",...]`（必须是 role=Family 的用户）
func writeResidentFamily(ctx context.Context, db interface {
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}, residentHOA string, raw json.RawMessage) error {
	var userIDs []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &userIDs)
	}
	if _, err := db.ExecContext(ctx,
		`DELETE FROM resident_caregivers WHERE resident_id = $1::INET AND family_id IS NOT NULL`,
		residentHOA); err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return nil
	}
	for _, uid := range userIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO resident_caregivers (resident_id, family_id)
			VALUES ($1::INET, $2::UUID)
			ON CONFLICT (resident_id, family_id) WHERE family_id IS NOT NULL DO NOTHING`,
			residentHOA, uid); err != nil {
			return fmt.Errorf("insert family %s: %w", uid, err)
		}
	}
	return nil
}

// ============================================================================
// PHI / Contacts encrypt+upsert helpers (v2)
//
// FE/BE/DB 字段名 1:1 对应，所有 PHI 字段 AES-256-GCM 加密为三件套
// (*_enc / *_iv / *_tag)。值类型先 sprintf 成字符串再加密。
// ============================================================================

// encryptPHIField — 把任意类型值加密到 (enc, iv, tag) 三件套
func (s *residentService) encryptPHIField(tenantID string, v interface{}) ([]byte, []byte, []byte, error) {
	if s.phiCryptor == nil {
		return nil, nil, nil, fmt.Errorf("phi cryptor not initialized")
	}
	var plaintext string
	switch x := v.(type) {
	case nil:
		return nil, nil, nil, nil
	case *string:
		if x == nil {
			return nil, nil, nil, nil
		}
		plaintext = *x
	case *int:
		if x == nil {
			return nil, nil, nil, nil
		}
		plaintext = fmt.Sprintf("%d", *x)
	case *float64:
		if x == nil {
			return nil, nil, nil, nil
		}
		plaintext = strconv.FormatFloat(*x, 'f', -1, 64)
	case *bool:
		if x == nil {
			return nil, nil, nil, nil
		}
		if *x {
			plaintext = "true"
		} else {
			plaintext = "false"
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported phi field type %T", v)
	}
	if plaintext == "" {
		return nil, nil, nil, nil
	}
	full, err := s.phiCryptor.EncryptString(tenantID, plaintext)
	if err != nil {
		return nil, nil, nil, err
	}
	// PHICryptor.EncryptString returns concatenated bytes; for storage we split
	// (kmscrypto.EncryptWithDataKey returns iv|ciphertext|tag — but PHICryptor API
	// returns single byte slice). We store the full blob in _enc and leave _iv/_tag NULL.
	// Decryption uses the same single-blob format.
	return full, nil, nil, nil
}

// writeResidentPHIv2 — UPSERT resident_phi 加密所有 PHI 字段
func (s *residentService) writeResidentPHIv2(ctx context.Context, residentHOA, tenantPrefix string, p *domain.ResidentPHIv2) error {
	if p == nil {
		return nil
	}
	if s.phiCryptor == nil {
		return fmt.Errorf("phi cryptor not initialized — set KMS_SOCKET and MASTER_PIN")
	}

	type fieldSpec struct {
		col string
		val interface{}
	}
	specs := []fieldSpec{
		{"first_name", p.FirstName},
		{"last_name", p.LastName},
		{"date_of_birth", p.DateOfBirth},
		{"resident_phone", p.ResidentPhone},
		{"resident_email", p.ResidentEmail},
		{"weight_lb", p.WeightLb},
		{"height_ft", p.HeightFt},
		{"height_in", p.HeightIn},
		{"mobility_level", p.MobilityLevel},
		{"tremor_status", p.TremorStatus},
		{"mobility_aid", p.MobilityAid},
		{"adl_assistance", p.ADLAssistance},
		{"comm_status", p.CommStatus},
		{"has_hypertension", p.HasHypertension},
		{"has_hyperlipaemia", p.HasHyperlipaemia},
		{"has_hyperglycaemia", p.HasHyperglycaemia},
		{"has_stroke_history", p.HasStrokeHistory},
		{"has_paralysis", p.HasParalysis},
		{"has_alzheimer", p.HasAlzheimer},
		{"medical_history", p.MedicalHistory},
		{"home_address_street", p.HomeAddressStreet},
		{"home_address_city", p.HomeAddressCity},
		{"home_address_state", p.HomeAddressState},
		{"home_address_postal_code", p.HomeAddressPostalCode},
	}

	cols := []string{"resident_id"}
	placeholders := []string{"$1::INET"}
	args := []interface{}{residentHOA}
	updates := []string{}
	idx := 2

	for _, f := range specs {
		enc, iv, tag, err := s.encryptPHIField(tenantPrefix, f.val)
		if err != nil {
			return fmt.Errorf("encrypt %s: %w", f.col, err)
		}
		if enc == nil {
			continue
		}
		cols = append(cols, f.col+"_enc", f.col+"_iv", f.col+"_tag")
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx), fmt.Sprintf("$%d", idx+1), fmt.Sprintf("$%d", idx+2))
		args = append(args, enc, iv, tag)
		updates = append(updates,
			fmt.Sprintf("%s_enc = EXCLUDED.%s_enc", f.col, f.col),
			fmt.Sprintf("%s_iv = EXCLUDED.%s_iv", f.col, f.col),
			fmt.Sprintf("%s_tag = EXCLUDED.%s_tag", f.col, f.col),
		)
		idx += 3
	}

	// plus_code 明文列
	if p.PlusCode != nil {
		cols = append(cols, "plus_code")
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		args = append(args, *p.PlusCode)
		updates = append(updates, "plus_code = EXCLUDED.plus_code")
		idx++
	}

	// OTP 子对象（明文）— 仅当传入时写
	if p.OTP != nil {
		if p.OTP.Code != nil {
			cols = append(cols, "otp_code")
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, *p.OTP.Code)
			updates = append(updates, "otp_code = EXCLUDED.otp_code")
			idx++
		}
		if p.OTP.Purpose != nil {
			cols = append(cols, "otp_purpose")
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, *p.OTP.Purpose)
			updates = append(updates, "otp_purpose = EXCLUDED.otp_purpose")
			idx++
		}
		if p.OTP.IssuedAt != nil {
			cols = append(cols, "otp_issued_at")
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, *p.OTP.IssuedAt)
			updates = append(updates, "otp_issued_at = EXCLUDED.otp_issued_at")
			idx++
		}
		if p.OTP.ExpiresAt != nil {
			cols = append(cols, "otp_expires_at")
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, *p.OTP.ExpiresAt)
			updates = append(updates, "otp_expires_at = EXCLUDED.otp_expires_at")
			idx++
		}
	}

	if len(cols) == 1 {
		// 只有 resident_id 一列 → 无字段可写
		return nil
	}

	updates = append(updates, "encrypted_at = NOW()")
	q := fmt.Sprintf(
		"INSERT INTO resident_phi (%s) VALUES (%s) ON CONFLICT (resident_id) DO UPDATE SET %s",
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)
	if _, err := s.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("upsert resident_phi: %w", err)
	}
	return nil
}

// decryptPHIField — 反向：从 (enc, iv, tag) 解出明文字符串
func (s *residentService) decryptPHIField(tenantID string, enc []byte) (string, error) {
	if s.phiCryptor == nil || len(enc) == 0 {
		return "", nil
	}
	return s.phiCryptor.DecryptString(tenantID, enc)
}

// loadResidentPHIv2 — 解密读 resident_phi
func (s *residentService) loadResidentPHIv2(ctx context.Context, residentHOA, tenantPrefix string) (*domain.ResidentPHIv2, error) {
	if s.phiCryptor == nil {
		return nil, nil
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT
			first_name_enc, last_name_enc, date_of_birth_enc,
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
			otp_code, otp_purpose, otp_issued_at, otp_expires_at,
			otp_used_at, otp_used_by
		FROM resident_phi WHERE resident_id = $1::INET`, residentHOA)

	var (
		firstNameEnc, lastNameEnc, dobEnc           []byte
		phoneEnc, emailEnc                          []byte
		weightEnc, heightFtEnc, heightInEnc         []byte
		mobilityLvlEnc, tremorEnc, aidEnc           []byte
		adlEnc, commEnc                             []byte
		hypEnc, hyperlipEnc, hyperglycEnc           []byte
		strokeEnc, paraEnc, alzEnc                  []byte
		historyEnc                                  []byte
		streetEnc, cityEnc, stateEnc, postalEnc     []byte
		plusCode                                    sql.NullString
		otpCode, otpPurpose                         sql.NullString
		otpIssued, otpExpires, otpUsed              sql.NullTime
		otpUsedBy                                   sql.NullString
	)
	if err := row.Scan(
		&firstNameEnc, &lastNameEnc, &dobEnc,
		&phoneEnc, &emailEnc,
		&weightEnc, &heightFtEnc, &heightInEnc,
		&mobilityLvlEnc, &tremorEnc, &aidEnc,
		&adlEnc, &commEnc,
		&hypEnc, &hyperlipEnc, &hyperglycEnc,
		&strokeEnc, &paraEnc, &alzEnc,
		&historyEnc,
		&streetEnc, &cityEnc, &stateEnc, &postalEnc,
		&plusCode,
		&otpCode, &otpPurpose, &otpIssued, &otpExpires,
		&otpUsed, &otpUsedBy,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query resident_phi: %w", err)
	}

	dec := func(b []byte) *string {
		if len(b) == 0 {
			return nil
		}
		s, err := s.decryptPHIField(tenantPrefix, b)
		if err != nil || s == "" {
			return nil
		}
		return &s
	}
	decFloat := func(b []byte) *float64 {
		sp := dec(b)
		if sp == nil {
			return nil
		}
		f, err := strconv.ParseFloat(*sp, 64)
		if err != nil {
			return nil
		}
		return &f
	}
	decInt := func(b []byte) *int {
		sp := dec(b)
		if sp == nil {
			return nil
		}
		n, err := strconv.Atoi(*sp)
		if err != nil {
			return nil
		}
		return &n
	}
	decBool := func(b []byte) *bool {
		sp := dec(b)
		if sp == nil {
			return nil
		}
		v := *sp == "true"
		return &v
	}

	out := &domain.ResidentPHIv2{
		FirstName:             dec(firstNameEnc),
		LastName:              dec(lastNameEnc),
		DateOfBirth:           dec(dobEnc),
		ResidentPhone:         dec(phoneEnc),
		ResidentEmail:         dec(emailEnc),
		WeightLb:              decFloat(weightEnc),
		HeightFt:              decFloat(heightFtEnc),
		HeightIn:              decFloat(heightInEnc),
		MobilityLevel:         decInt(mobilityLvlEnc),
		TremorStatus:          dec(tremorEnc),
		MobilityAid:           dec(aidEnc),
		ADLAssistance:         dec(adlEnc),
		CommStatus:            dec(commEnc),
		HasHypertension:       decBool(hypEnc),
		HasHyperlipaemia:      decBool(hyperlipEnc),
		HasHyperglycaemia:     decBool(hyperglycEnc),
		HasStrokeHistory:      decBool(strokeEnc),
		HasParalysis:          decBool(paraEnc),
		HasAlzheimer:          decBool(alzEnc),
		MedicalHistory:        dec(historyEnc),
		HomeAddressStreet:     dec(streetEnc),
		HomeAddressCity:       dec(cityEnc),
		HomeAddressState:      dec(stateEnc),
		HomeAddressPostalCode: dec(postalEnc),
	}
	if plusCode.Valid && plusCode.String != "" {
		v := plusCode.String
		out.PlusCode = &v
	}
	if otpCode.Valid {
		otp := &domain.ResidentOTPv2{}
		v := otpCode.String
		otp.Code = &v
		if otpPurpose.Valid {
			p := otpPurpose.String
			otp.Purpose = &p
		}
		if otpIssued.Valid {
			t := otpIssued.Time.Format(time.RFC3339)
			otp.IssuedAt = &t
		}
		if otpExpires.Valid {
			t := otpExpires.Time.Format(time.RFC3339)
			otp.ExpiresAt = &t
		}
		if otpUsed.Valid {
			t := otpUsed.Time.Format(time.RFC3339)
			otp.UsedAt = &t
		}
		if otpUsedBy.Valid {
			id := otpUsedBy.String
			otp.UsedBy = &id
		}
		out.OTP = otp
	}
	return out, nil
}

// writeResidentContactsV2 — 全量替换 resident_contacts 的 contacts 列表
//   - 传入 nil → 不改
//   - 传入 [] → 清空
//   - 传入 N 项 → 替换全部
func (s *residentService) writeResidentContactsV2(ctx context.Context, residentHOA, tenantPrefix string, contacts []domain.ResidentContactV2) error {
	if s.phiCryptor == nil {
		return fmt.Errorf("phi cryptor not initialized")
	}
	// 全量替换：先删后插
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM resident_contacts WHERE resident_id = $1::INET`, residentHOA); err != nil {
		return fmt.Errorf("delete resident_contacts: %w", err)
	}
	for _, c := range contacts {
		fnEnc, _, _, err := s.encryptPHIField(tenantPrefix, c.ContactFirstName)
		if err != nil {
			return fmt.Errorf("encrypt first_name: %w", err)
		}
		lnEnc, _, _, err := s.encryptPHIField(tenantPrefix, c.ContactLastName)
		if err != nil {
			return fmt.Errorf("encrypt last_name: %w", err)
		}
		phEnc, _, _, err := s.encryptPHIField(tenantPrefix, c.ContactPhone)
		if err != nil {
			return fmt.Errorf("encrypt phone: %w", err)
		}
		emEnc, _, _, err := s.encryptPHIField(tenantPrefix, c.ContactEmail)
		if err != nil {
			return fmt.Errorf("encrypt email: %w", err)
		}
		relationship := strings.TrimSpace(c.Relationship)
		if relationship == "" {
			relationship = "other"
		}
		var linkedUserID interface{}
		if c.LinkedUserID != nil && *c.LinkedUserID != "" {
			linkedUserID = *c.LinkedUserID
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO resident_contacts
			  (resident_id, linked_user_id, relationship,
			   contact_first_name_enc, contact_last_name_enc, contact_phone_enc, contact_email_enc,
			   receive_sms, receive_email)
			VALUES ($1::INET, $2, $3, $4, $5, $6, $7, $8, $9)`,
			residentHOA, linkedUserID, relationship,
			fnEnc, lnEnc, phEnc, emEnc,
			c.ReceiveSMS, c.ReceiveEmail,
		); err != nil {
			return fmt.Errorf("insert resident_contact: %w", err)
		}
	}
	return nil
}

// loadResidentContactsV2 — 解密读 resident_contacts
func (s *residentService) loadResidentContactsV2(ctx context.Context, residentHOA, tenantPrefix string) ([]domain.ResidentContactV2, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT contact_id::text, COALESCE(linked_user_id::text, ''), relationship,
		       contact_first_name_enc, contact_last_name_enc,
		       contact_phone_enc, contact_email_enc,
		       receive_sms, receive_email
		  FROM resident_contacts
		 WHERE resident_id = $1::INET
		 ORDER BY created_at`, residentHOA)
	if err != nil {
		return nil, fmt.Errorf("query resident_contacts: %w", err)
	}
	defer rows.Close()
	out := []domain.ResidentContactV2{}
	for rows.Next() {
		var (
			contactID, linkedUserID, relationship string
			fnEnc, lnEnc, phEnc, emEnc            []byte
			recvSMS, recvEmail                    bool
		)
		if err := rows.Scan(&contactID, &linkedUserID, &relationship,
			&fnEnc, &lnEnc, &phEnc, &emEnc, &recvSMS, &recvEmail); err != nil {
			continue
		}
		c := domain.ResidentContactV2{
			ContactID:    contactID,
			Relationship: relationship,
			ReceiveSMS:   recvSMS,
			ReceiveEmail: recvEmail,
		}
		if linkedUserID != "" {
			c.LinkedUserID = &linkedUserID
		}
		dec := func(b []byte) *string {
			if len(b) == 0 {
				return nil
			}
			v, err := s.decryptPHIField(tenantPrefix, b)
			if err != nil || v == "" {
				return nil
			}
			return &v
		}
		c.ContactFirstName = dec(fnEnc)
		c.ContactLastName = dec(lnEnc)
		c.ContactPhone = dec(phEnc)
		c.ContactEmail = dec(emEnc)
		out = append(out, c)
	}
	return out, nil
}

// ============================================================================
// Service-layer request → domain.ResidentPHIv2 / domain.ResidentContactV2 mappers
//   v1 风格 UpdateXxx pointer sentinel → 直接 *Type pointer（v2 风格）
// ============================================================================

func mapPHIRequestToDomain(req *UpdateResidentPHIRequest) *domain.ResidentPHIv2 {
	if req == nil {
		return nil
	}
	out := &domain.ResidentPHIv2{}
	deStr := func(u *domain.UpdateString) *string {
		if u == nil || u.Action != domain.UpdateActionUpdate {
			return nil
		}
		v := u.Value
		return &v
	}
	deFloat := func(u *domain.UpdateFloat64) *float64 {
		if u == nil || u.Action != domain.UpdateActionUpdate {
			return nil
		}
		v := u.Value
		return &v
	}
	deInt := func(u *domain.UpdateInt) *int {
		if u == nil || u.Action != domain.UpdateActionUpdate {
			return nil
		}
		v := u.Value
		return &v
	}
	deBool := func(u *domain.UpdateBool) *bool {
		if u == nil || u.Action != domain.UpdateActionUpdate {
			return nil
		}
		v := u.Value
		return &v
	}
	deTime := func(u *domain.UpdateTime) *string {
		if u == nil || u.Action != domain.UpdateActionUpdate || u.Value == nil {
			return nil
		}
		s := u.Value.Format("2006-01-02")
		return &s
	}

	out.FirstName = deStr(req.FirstName)
	out.LastName = deStr(req.LastName)
	out.DateOfBirth = deTime(req.DateOfBirth)
	out.ResidentPhone = deStr(req.ResidentPhone)
	out.ResidentEmail = deStr(req.ResidentEmail)
	out.WeightLb = deFloat(req.WeightLb)
	out.HeightFt = deFloat(req.HeightFt)
	out.HeightIn = deFloat(req.HeightIn)
	out.MobilityLevel = deInt(req.MobilityLevel)
	out.TremorStatus = deStr(req.TremorStatus)
	out.MobilityAid = deStr(req.MobilityAid)
	out.ADLAssistance = deStr(req.ADLAssistance)
	out.CommStatus = deStr(req.CommStatus)
	out.HasHypertension = deBool(req.HasHypertension)
	out.HasHyperlipaemia = deBool(req.HasHyperlipaemia)
	out.HasHyperglycaemia = deBool(req.HasHyperglycaemia)
	out.HasStrokeHistory = deBool(req.HasStrokeHistory)
	out.HasParalysis = deBool(req.HasParalysis)
	out.HasAlzheimer = deBool(req.HasAlzheimer)
	out.MedicalHistory = deStr(req.MedicalHistory)
	out.HomeAddressStreet = deStr(req.HomeAddressStreet)
	out.HomeAddressCity = deStr(req.HomeAddressCity)
	out.HomeAddressState = deStr(req.HomeAddressState)
	out.HomeAddressPostalCode = deStr(req.HomeAddressPostalCode)
	out.PlusCode = deStr(req.PlusCode)
	return out
}

func mapContactsRequestToDomain(reqs []*UpdateResidentContactRequest) []domain.ResidentContactV2 {
	out := []domain.ResidentContactV2{}
	deStr := func(u *domain.UpdateString) *string {
		if u == nil || u.Action != domain.UpdateActionUpdate {
			return nil
		}
		v := u.Value
		return &v
	}
	deBool := func(u *domain.UpdateBool) bool {
		if u == nil || u.Action != domain.UpdateActionUpdate {
			return false
		}
		return u.Value
	}
	for _, r := range reqs {
		if r == nil {
			continue
		}
		// 仅当至少一个字段更新时视为有效 contact 行（避免空 slot 创建空白行）
		hasData := false
		c := domain.ResidentContactV2{}
		if v := deStr(r.Relationship); v != nil {
			c.Relationship = *v
			hasData = true
		}
		c.ContactFirstName = deStr(r.ContactFirstName)
		c.ContactLastName = deStr(r.ContactLastName)
		c.ContactPhone = deStr(r.ContactPhone)
		c.ContactEmail = deStr(r.ContactEmail)
		if c.ContactFirstName != nil || c.ContactLastName != nil || c.ContactPhone != nil || c.ContactEmail != nil {
			hasData = true
		}
		c.ReceiveSMS = deBool(r.ReceiveSMS)
		c.ReceiveEmail = deBool(r.ReceiveEmail)
		if !hasData {
			continue
		}
		if c.Relationship == "" {
			c.Relationship = "other"
		}
		out = append(out, c)
	}
	return out
}

// domain ResidentPHIv2 → DTO ResidentPHIDTO （response 用）
func mapPHIDomainToDTO(p *domain.ResidentPHIv2) *ResidentPHIDTO {
	if p == nil {
		return nil
	}
	out := &ResidentPHIDTO{
		FirstName:             p.FirstName,
		LastName:              p.LastName,
		ResidentPhone:         p.ResidentPhone,
		ResidentEmail:         p.ResidentEmail,
		WeightLb:              p.WeightLb,
		HeightFt:              p.HeightFt,
		HeightIn:              p.HeightIn,
		MobilityLevel:         p.MobilityLevel,
		TremorStatus:          p.TremorStatus,
		MobilityAid:           p.MobilityAid,
		ADLAssistance:         p.ADLAssistance,
		CommStatus:            p.CommStatus,
		HasHypertension:       p.HasHypertension,
		HasHyperlipaemia:      p.HasHyperlipaemia,
		HasHyperglycaemia:     p.HasHyperglycaemia,
		HasStrokeHistory:      p.HasStrokeHistory,
		HasParalysis:          p.HasParalysis,
		HasAlzheimer:          p.HasAlzheimer,
		MedicalHistory:        p.MedicalHistory,
		HomeAddressStreet:     p.HomeAddressStreet,
		HomeAddressCity:       p.HomeAddressCity,
		HomeAddressState:      p.HomeAddressState,
		HomeAddressPostalCode: p.HomeAddressPostalCode,
		PlusCode:              p.PlusCode,
	}
	// DateOfBirth: domain 是 ISO string, DTO 是 unix ts
	if p.DateOfBirth != nil {
		if t, err := time.Parse("2006-01-02", *p.DateOfBirth); err == nil {
			ts := t.Unix()
			out.DateOfBirth = &ts
		}
	}
	return out
}

func mapContactsDomainToDTO(in []domain.ResidentContactV2) []*ResidentContactDTO {
	out := []*ResidentContactDTO{}
	for _, c := range in {
		rel := c.Relationship
		dto := &ResidentContactDTO{
			ContactID:        c.ContactID,
			IsEnabled:        true, // v2: 行存在即视为 enabled
			Relationship:     &rel,
			ContactFirstName: c.ContactFirstName,
			ContactLastName:  c.ContactLastName,
			ContactPhone:     c.ContactPhone,
			ContactEmail:     c.ContactEmail,
			ReceiveSMS:       c.ReceiveSMS,
			ReceiveEmail:     c.ReceiveEmail,
		}
		out = append(out, dto)
	}
	return out
}

// searchResidentsByPHI — Phase 3b PHI 搜索：根据权限解密 email/phone 进行模糊查找
// 用于 email/phone 搜索，涉及 PHI 访问权限检查和审计记录
func (s *residentService) searchResidentsByPHI(
	ctx context.Context,
	tenantPrefix, searchTerm string,
	searchType SearchType,
	permCheck *PermissionCheckResult,
	currentUserID string,
	limit int,
) ([]*ResidentListItemDTO, error) {
	if s.phiCryptor == nil {
		// PHI 解密未启用，回退空结果
		return []*ResidentListItemDTO{}, nil
	}

	// 1. 获取权限允许的住户列表（base query）
	args := []any{tenantPrefix}
	where := []string{
		"network(set_masklen(r.resident_id, 48)) = $1::INET",
		"r.status <> 'deleted'",
	}

	// 应用 assigned_only/branch_only 权限过滤（同 listResidentsV2）
	argN := 2
	if permCheck != nil && permCheck.AssignedOnly && currentUserID != "" {
		args = append(args, currentUserID)
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM resident_caregivers rc
			WHERE rc.resident_id = r.resident_id
			  AND rc.caregiver_id = $%d::UUID
		)`, argN))
		argN++
	}

	if permCheck != nil && permCheck.BranchOnly && currentUserID != "" {
		userBranchIDs, hasBranches, err := s.getUserBranchIDs(ctx, strings.Split(tenantPrefix, "/")[0], currentUserID)
		if err != nil {
			return nil, fmt.Errorf("get user branch IDs: %w", err)
		}

		if !hasBranches {
			// 用户无关联院区：仅能查看 /56 为 NULL（跨院区过渡态）的住户
			where = append(where, `EXISTS (
				SELECT 1 FROM resident_unit ru
				WHERE ru.resident_id = r.resident_id
				  AND ru.valid_to IS NULL
				  AND masklen(ru.spatial_prefix) <= 56
			)`)
		} else if len(userBranchIDs) == 1 {
			// 用户属于 1 个院区：查看该院区的住户
			args = append(args, userBranchIDs[0])
			where = append(where, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM resident_unit ru
				WHERE ru.resident_id = r.resident_id
				  AND ru.valid_to IS NULL
				  AND host(network(set_masklen(ru.spatial_prefix, 56))) = $%d
			)`, argN))
			argN++
		} else {
			// 用户属于多个院区：查看所有关联院区的住户
			placeholders := make([]string, len(userBranchIDs))
			for i, branchID := range userBranchIDs {
				args = append(args, branchID)
				placeholders[i] = fmt.Sprintf("$%d", len(args))
			}
			where = append(where, fmt.Sprintf(`EXISTS (
				SELECT 1 FROM resident_unit ru
				WHERE ru.resident_id = r.resident_id
				  AND ru.valid_to IS NULL
				  AND host(network(set_masklen(ru.spatial_prefix, 56))) IN (%s)
			)`, strings.Join(placeholders, ", ")))
		}
	}

	whereClause := strings.Join(where, " AND ")

	// 2. 查询权限允许的住户及其 PHI 加密数据
	q := `
		SELECT host(r.resident_id) AS hoa,
		       COALESCE(r.resident_account, '') AS resident_account,
		       r.nickname,
		       rp.resident_email_enc,
		       rp.resident_phone_enc,
		       host(network(set_masklen(r.resident_id, 48))) || '/48' AS tenant_id
		  FROM residents r
		  LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id
		 WHERE ` + whereClause + `
		 LIMIT $` + fmt.Sprintf("%d", argN)

	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query residents for phi search: %w", err)
	}
	defer rows.Close()

	searchLower := strings.ToLower(strings.TrimSpace(searchTerm))
	out := []*ResidentListItemDTO{}

	for rows.Next() {
		var (
			hoa, account, nickname, tenantID string
			emailEnc, phoneEnc               sql.NullString
		)
		if err := rows.Scan(&hoa, &account, &nickname, &emailEnc, &phoneEnc, &tenantID); err != nil {
			continue
		}

		// 3. 根据搜索类型解密并匹配
		matched := false
		switch searchType {
		case SearchTypeEmail:
			if emailEnc.Valid && emailEnc.String != "" {
				if decrypted, err := s.decryptPHIField(tenantID, []byte(emailEnc.String)); err == nil {
					if strings.Contains(strings.ToLower(decrypted), searchLower) {
						matched = true
					}
				}
			}
		case SearchTypePhone:
			if phoneEnc.Valid && phoneEnc.String != "" {
				if decrypted, err := s.decryptPHIField(tenantID, []byte(phoneEnc.String)); err == nil {
					if strings.Contains(strings.ToLower(decrypted), searchLower) {
						matched = true
					}
				}
			}
		}

		if matched {
			item := &ResidentListItemDTO{
				ResidentID:      hoa,
				TenantID:        tenantID,
				Nickname:        nickname,
				Status:          "active",
				FamilyAccess:    true,
			}
			if account != "" {
				item.ResidentAccount = &account
			}
			out = append(out, item)

			// 4. 记录审计日志（PHI 访问）
			s.logger.Info("PHI access for search",
				zap.String("tenant_id", tenantID),
				zap.String("hoa", hoa),
				zap.String("search_type", GetSearchTypeDescription(searchType)),
				zap.String("user_id", currentUserID))
		}
	}

	return out, rows.Err()
}
