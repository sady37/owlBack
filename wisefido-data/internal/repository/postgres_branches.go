package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresBranchesRepository 院区 Repository — owl_v2 schema 版本。
//
// v2 schema 与 v1 关键差异（不向后兼容）：
//   - 主键从 branch_id UUID 改为 branch_id INET (/56)
//   - 父租户关系不再有显式 tenant_id 外键列；靠 INET prefix-match (`branches.branch_id <<= tenants.branch_id`)
//   - 列改名：description → address
//   - 列新增：branch_slot SMALLINT (uint8 0..254) / timezone / created_at / updated_at
//   - users/residents/units 不再用 user_branches 表关联；rooms 等下游也是 prefix-match
//
// 兼容策略：domain.Branch.BranchID/TenantID 在 v2 装 CIDR 字符串；Description 字段映射到 address。
type PostgresBranchesRepository struct {
	db *sql.DB
}

func NewPostgresBranchesRepository(db *sql.DB) *PostgresBranchesRepository {
	return &PostgresBranchesRepository{db: db}
}

var _ BranchesRepository = (*PostgresBranchesRepository)(nil)

// network(set_masklen(...)) — 必须先 set_masklen 再 network，否则 host bits（byte 6+）不会被 mask 掉。
// status 列：HIPAA + 业务约束 — branches 也走软删（active/suspended/deleted），不向 v1 domain.Branch 暴露
// （domain.Branch 没有 Status 字段；此值在 SQL 层 default 过滤掉 deleted，外层不见）。
const branchesSelectCols = `
		host(branch_id) || '/56' AS branch_id,
		network(set_masklen(branch_id, 48))::text AS tenant_id,
		branch_name,
		address AS description,
		COALESCE(timezone, '') AS timezone,
		created_at,
		updated_at
`

// GetBranch 按 branch CIDR 查；tenantID 可选，给了就限定父 prefix。
func (r *PostgresBranchesRepository) GetBranch(ctx context.Context, tenantID, branchID string) (*domain.Branch, error) {
	if branchID == "" {
		return nil, sql.ErrNoRows
	}
	if !looksLikeINETPrefix(branchID) {
		return nil, fmt.Errorf("branch not found: %q is not a v2 INET prefix", branchID)
	}

	q := `SELECT ` + branchesSelectCols + ` FROM branches WHERE branch_id = $1::INET`
	args := []any{branchID}
	if tenantID != "" && looksLikeINETPrefix(tenantID) {
		q += ` AND branch_id <<= $2::INET`
		args = append(args, tenantID)
	}
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("branch not found: branch_id '%s'", branchID)
	}
	return scanBranch(rows)
}

// GetBranchByName 按 (tenant prefix, branch_name) 查。
func (r *PostgresBranchesRepository) GetBranchByName(ctx context.Context, tenantID, branchName string) (*domain.Branch, error) {
	if tenantID == "" || branchName == "" {
		return nil, sql.ErrNoRows
	}
	if !looksLikeINETPrefix(tenantID) {
		return nil, fmt.Errorf("tenant not found: %q is not a v2 INET prefix", tenantID)
	}
	q := `SELECT ` + branchesSelectCols + ` FROM branches WHERE branch_id <<= $1::INET AND branch_name = $2 LIMIT 1`
	rows, err := r.db.QueryContext(ctx, q, tenantID, branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch by name: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("branch not found: tenant_id '%s', branch_name '%s'", tenantID, branchName)
	}
	return scanBranch(rows)
}

// ListBranches 按 tenant_id 列出 branches，支持 branch_name + address 模糊搜索。
//
// v1 search 还覆盖了 nested users/units/residents 文本（user_branches JOIN）；
// v2 schema 删了 user_branches，这部分降级（前端 BranchList nested arrays 暂用空数组占位）。
func (r *PostgresBranchesRepository) ListBranches(ctx context.Context, tenantID string, search string, page, size int) ([]*domain.Branch, int, error) {
	if tenantID == "" {
		return []*domain.Branch{}, 0, nil
	}
	if !looksLikeINETPrefix(tenantID) {
		return []*domain.Branch{}, 0, nil
	}

	where := []string{
		"branch_id <<= $1::INET",
		"COALESCE(status, 'active') <> 'deleted'", // 默认隐藏软删的
	}
	args := []any{tenantID}
	argIdx := 2

	if search != "" {
		pat := "%" + strings.ToLower(search) + "%"
		where = append(where, fmt.Sprintf("(LOWER(branch_name) LIKE $%d OR LOWER(COALESCE(address, '')) LIKE $%d)", argIdx, argIdx))
		args = append(args, pat)
		argIdx++
	}
	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM branches `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count branches: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	listQ := `SELECT ` + branchesSelectCols + ` FROM branches ` + whereClause +
		fmt.Sprintf(` ORDER BY branch_slot ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list branches: %w", err)
	}
	defer rows.Close()

	branches := []*domain.Branch{}
	for rows.Next() {
		b, err := scanBranch(rows)
		if err != nil {
			return nil, 0, err
		}
		branches = append(branches, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate branches: %w", err)
	}
	return branches, total, nil
}

// CreateBranch 在 tenant 下分配下一个 branch_slot 派生 /56 prefix。
//
// 并发安全：advisory lock 串行化同 tenant 内的 slot 分配。
func (r *PostgresBranchesRepository) CreateBranch(ctx context.Context, tenantID string, branch *domain.Branch) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if !looksLikeINETPrefix(tenantID) {
		return "", fmt.Errorf("tenant_id %q is not a v2 INET prefix", tenantID)
	}
	if branch == nil || branch.BranchName == "" {
		return "", fmt.Errorf("branch_name is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// advisory lock 按 tenant prefix 哈希
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "owl_v2.branches.alloc:"+tenantID); err != nil {
		return "", fmt.Errorf("advisory lock: %w", err)
	}

	// 唯一性 (tenant, branch_name)
	var dup string
	err = tx.QueryRowContext(ctx,
		`SELECT host(branch_id) || '/56' FROM branches
		 WHERE branch_id <<= $1::INET AND branch_name = $2 LIMIT 1`,
		tenantID, branch.BranchName,
	).Scan(&dup)
	if err == nil {
		return "", fmt.Errorf("branch_name already exists: tenant_id '%s', branch_name '%s'", tenantID, branch.BranchName)
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("check branch_name: %w", err)
	}

	// 下一 slot — 从 1 起（slot 0 保留为 "unbound 哨兵"，0xFF wildcard 保留）
	var nextSlot int
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(branch_slot), 0) + 1 FROM branches WHERE branch_id <<= $1::INET`, tenantID,
	).Scan(&nextSlot); err != nil {
		return "", fmt.Errorf("compute next branch_slot: %w", err)
	}
	if nextSlot < 1 || nextSlot > 254 {
		return "", fmt.Errorf("branch_slot exhausted: %d (valid 1..254)", nextSlot)
	}

	branchPrefix, err := deriveBranchCIDR(tenantID, byte(nextSlot))
	if err != nil {
		return "", fmt.Errorf("derive branch prefix: %w", err)
	}

	var addrArg any
	if branch.Description.Valid && branch.Description.String != "" {
		addrArg = branch.Description.String
	}

	var tzArg any
	if tz := strings.TrimSpace(branch.Timezone); tz != "" {
		tzArg = tz
	}

	var newPrefix string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO branches (branch_id, branch_slot, branch_name, address, timezone)
		VALUES ($1::INET, $2, $3, $4, $5)
		RETURNING host(branch_id) || '/56'
	`, branchPrefix, nextSlot, branch.BranchName, addrArg, tzArg).Scan(&newPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	return newPrefix, nil
}

// UpdateBranch 简单更新 branch_name / address（v2: description 列已改名为 address）。
func (r *PostgresBranchesRepository) UpdateBranch(ctx context.Context, tenantID, branchID string, branch *domain.Branch) error {
	if branchID == "" {
		return fmt.Errorf("branch_id is required")
	}
	if !looksLikeINETPrefix(branchID) {
		return fmt.Errorf("branch_id %q is not a v2 INET prefix", branchID)
	}
	if branch == nil {
		return fmt.Errorf("branch is required")
	}

	updates := []string{}
	args := []any{branchID}
	argIdx := 2

	if branch.BranchName != "" {
		updates = append(updates, fmt.Sprintf("branch_name = $%d", argIdx))
		args = append(args, branch.BranchName)
		argIdx++
	}
	if branch.Description.Valid {
		if branch.Description.String != "" {
			updates = append(updates, fmt.Sprintf("address = $%d", argIdx))
			args = append(args, branch.Description.String)
			argIdx++
		} else {
			updates = append(updates, "address = NULL")
		}
	}
	if tz := strings.TrimSpace(branch.Timezone); tz != "" {
		updates = append(updates, fmt.Sprintf("timezone = $%d", argIdx))
		args = append(args, tz)
		argIdx++
	}
	updates = append(updates, "updated_at = NOW()")

	if len(updates) == 1 {
		return nil // only updated_at, nothing meaningful
	}

	q := fmt.Sprintf(`UPDATE branches SET %s WHERE branch_id = $1::INET`, strings.Join(updates, ", "))
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("branch_name already exists: branch_id '%s', branch_name '%s'", branchID, branch.BranchName)
		}
		return fmt.Errorf("failed to update branch: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("branch not found: branch_id '%s'", branchID)
	}
	return nil
}

// UpdateBranchFields v2 schema 不再有 description 列；映射到 address。其它字段 keep/delete 语义保留。
func (r *PostgresBranchesRepository) UpdateBranchFields(ctx context.Context, tenantID, branchID string, update *domain.BranchUpdate) error {
	if branchID == "" {
		return fmt.Errorf("branch_id is required")
	}
	if !looksLikeINETPrefix(branchID) {
		return fmt.Errorf("branch_id %q is not a v2 INET prefix", branchID)
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
	args := []any{branchID}
	argIdx := 2

	if update.BranchName != nil {
		switch update.BranchName.Action {
		case domain.UpdateActionUpdate:
			if update.BranchName.Value == "" {
				return fmt.Errorf("branch_name cannot be empty (NOT NULL constraint)")
			}
			// 唯一性（同租户内，排除自己）
			var exists bool
			err = tx.QueryRowContext(ctx, `
				SELECT EXISTS (
				  SELECT 1 FROM branches
				   WHERE branch_id <<= set_masklen($1::INET, 48)
				     AND branch_name = $2
				     AND branch_id <> $1::INET
				)`, branchID, update.BranchName.Value).Scan(&exists)
			if err != nil {
				return fmt.Errorf("failed to check branch_name uniqueness: %w", err)
			}
			if exists {
				return fmt.Errorf("branch_name already exists: branch_name '%s'", update.BranchName.Value)
			}
			updates = append(updates, fmt.Sprintf("branch_name = $%d", argIdx))
			args = append(args, update.BranchName.Value)
			argIdx++
		case domain.UpdateActionDelete:
			return fmt.Errorf("branch_name cannot be deleted (NOT NULL constraint)")
		case domain.UpdateActionKeep:
			// noop
		}
	}

	if update.Description != nil {
		switch update.Description.Action {
		case domain.UpdateActionUpdate:
			updates = append(updates, fmt.Sprintf("address = $%d", argIdx))
			args = append(args, update.Description.Value)
			argIdx++
		case domain.UpdateActionDelete:
			updates = append(updates, "address = NULL")
		case domain.UpdateActionKeep:
			// noop
		}
	}

	if update.Timezone != nil {
		switch update.Timezone.Action {
		case domain.UpdateActionUpdate:
			updates = append(updates, fmt.Sprintf("timezone = $%d", argIdx))
			args = append(args, update.Timezone.Value)
			argIdx++
		case domain.UpdateActionDelete:
			updates = append(updates, "timezone = NULL")
		case domain.UpdateActionKeep:
			// noop
		}
	}

	updates = append(updates, "updated_at = NOW()")

	q := fmt.Sprintf(`UPDATE branches SET %s WHERE branch_id = $1::INET`, strings.Join(updates, ", "))
	res, err := tx.ExecContext(ctx, q, args...)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("branch_name already exists")
		}
		return fmt.Errorf("failed to update branch: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("branch not found: branch_id '%s'", branchID)
	}
	return tx.Commit()
}

// DeleteBranch 智能删除：
//   - 空 branch（下游 sites/units/rooms/beds/residents/devices 全空）→ 真物理 DELETE（误创建救济）
//   - 非空 branch → 拒绝（要求先删除子节点）；HIPAA 数据保留靠子层数据本身，不靠 branch 软删
func (r *PostgresBranchesRepository) DeleteBranch(ctx context.Context, tenantID, branchID string) error {
	if branchID == "" {
		return fmt.Errorf("branch_id is required")
	}
	if !looksLikeINETPrefix(branchID) {
		return fmt.Errorf("branch_id %q is not a v2 INET prefix", branchID)
	}
	empty, err := r.isBranchEmpty(ctx, branchID)
	if err != nil {
		return fmt.Errorf("check branch empty: %w", err)
	}
	if !empty {
		return fmt.Errorf("branch has children: delete sites/units/rooms/beds/residents/devices first")
	}
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM branches WHERE branch_id = $1::INET`, branchID)
	if err != nil {
		return fmt.Errorf("hard delete empty branch: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("branch not found: branch_id '%s'", branchID)
	}
	return nil
}

// isBranchEmpty 判定 branch 下游是否有任何业务数据（active/suspended/deleted 都算）。
//
// 检查表（branch_id <<= /56 范围内）：
//   - sites / units / rooms / beds（spatial 子层）
//   - residents（hoa /128 在 /56 范围内）
//   - devices（device_ipv6 /128 在 /56 范围内）
//
// 注意：v2 不再用 user_branches 关联用户与 branch；users 表只关联 tenant_id（/48），
// 不在此 check 范围内（删 branch 不影响 users）。
func (r *PostgresBranchesRepository) isBranchEmpty(ctx context.Context, branchPrefix string) (bool, error) {
	var hasData bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM sites     WHERE site_id <<= $1::INET
		  UNION ALL
		  SELECT 1 FROM units     WHERE unit_id <<= $1::INET
		  UNION ALL
		  SELECT 1 FROM rooms     WHERE room_id <<= $1::INET
		  UNION ALL
		  SELECT 1 FROM beds      WHERE bed_id <<= $1::INET
		  UNION ALL
		  SELECT 1 FROM residents WHERE resident_id IS NOT NULL AND resident_id <<= $1::INET
		  UNION ALL
		  SELECT 1 FROM devices   WHERE device_ipv6 IS NOT NULL AND device_ipv6 <<= $1::INET
		)
	`, branchPrefix).Scan(&hasData)
	if err != nil {
		return false, err
	}
	return !hasData, nil
}

// =============================================================================
// helpers
// =============================================================================

func scanBranch(rs rowScanner) (*domain.Branch, error) {
	var branch domain.Branch
	var description sql.NullString
	var timezone string
	var createdAt, updatedAt sql.NullTime
	if err := rs.Scan(
		&branch.BranchID,
		&branch.TenantID,
		&branch.BranchName,
		&description,
		&timezone,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to scan branch: %w", err)
	}
	branch.Description = description
	branch.Timezone = timezone
	branch.CreatedAt = createdAt
	branch.UpdatedAt = updatedAt
	return &branch, nil
}

// deriveBranchCIDR 从 tenant /48 + branch_slot 派生 branch /56 CIDR 字符串。
//
// IPv6 layout (per owl_v2 doc §2.1)：byte 6 = branch_slot；byte 7 = site_slot（branch /56 时为 0）。
//
//	tenant 'fd00:0:6::/48' + slot 1 → 'fd00:0:6:0100::/56'
//	tenant 'fd00:0:3::/48' + slot 0xab → 'fd00:0:3:ab00::/56'
func deriveBranchCIDR(tenantPrefix string, branchSlot byte) (string, error) {
	addrPart := tenantPrefix
	if i := strings.IndexByte(addrPart, '/'); i >= 0 {
		addrPart = addrPart[:i]
	}
	ip := net.ParseIP(addrPart)
	if ip == nil {
		return "", fmt.Errorf("invalid tenant prefix: %q", tenantPrefix)
	}
	v6 := ip.To16()
	if v6 == nil {
		return "", fmt.Errorf("not an IPv6 address: %q", tenantPrefix)
	}
	out := make(net.IP, 16)
	copy(out, v6)
	// 清掉 mask 之外的位（保险），再写 byte 6
	for i := 6; i < 16; i++ {
		out[i] = 0
	}
	out[6] = branchSlot
	return out.String() + "/56", nil
}
