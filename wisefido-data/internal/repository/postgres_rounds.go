package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresRoundsRepository 巡房记录 Repository — v2 schema 重写 (2026-05-14)
//
// v1→v2 列名映射：
//   tenant_id (UUID)   → 派生 host(set_masklen(caregiver_id, 48))||'/48'；scope 过滤用 caregiver_id <<= $tenant::INET
//   executor_id (UUID) → caregiver_id INET；FE 入参仍传 user_id UUID，repo 内部查 users.hoa 转 INET 后写入
//                        SELECT 时反向 LEFT JOIN users 还原 user_id 给 FE
//   unit_id (UUID)     → unit_prefix INET；FE 现在传 INET CIDR（unit pilot 后），直接落
//   round_time         → 用 ended_at 落（完成时刻）；FE 仍读 round_time 字段（SQL alias）
//   started_at         → 保留 v2 同名列
//   items_checked      → 拆到 rounds_report_snapshot.snapshot.items_checked JSONB（v2 新表）
//   notes / round_type / status → 保留
type PostgresRoundsRepository struct {
	db *sql.DB
}

// NewPostgresRoundsRepository 创建巡房记录 Repository
func NewPostgresRoundsRepository(db *sql.DB) *PostgresRoundsRepository {
	return &PostgresRoundsRepository{db: db}
}

var _ RoundsRepository = (*PostgresRoundsRepository)(nil)

// 公共 SELECT 子句：v2 列 → v1-style alias，FE 不感知 schema 变化
//
// caregiver_id INET → executor_id UUID via LEFT JOIN users（user_id::text）
// unit_prefix INET → unit_id::text
// ended_at TIMESTAMPTZ → round_time（完成时刻；fallback started_at 给 in_progress 行）
// snapshot JSONB → items_checked（从 rounds_report_snapshot LEFT JOIN）
const roundSelectV2 = `
	SELECT
	    r.round_id::text                                       AS round_id,
	    host(network(set_masklen(r.caregiver_id, 48)))||'/48'  AS tenant_id,
	    r.round_type                                           AS round_type,
	    COALESCE(host(r.unit_prefix)||'/'||masklen(r.unit_prefix), '') AS unit_id,
	    COALESCE(u.user_id::text, '')                          AS executor_id,
	    COALESCE(r.ended_at, r.started_at)                     AS round_time,
	    r.started_at                                           AS started_at,
	    COALESCE(rrs.snapshot->>'items_checked', '')           AS items_checked,
	    COALESCE(r.notes, '')                                  AS notes,
	    r.status                                               AS status
	FROM rounds r
	LEFT JOIN users u ON u.hoa = r.caregiver_id
	LEFT JOIN rounds_report_snapshot rrs ON rrs.round_id = r.round_id
`

// scanRoundRow 公共 row.Scan helper（GetRound 单行 + ListRounds 批量复用）
func scanRoundRow(row interface{ Scan(...any) error }) (*domain.Round, error) {
	var round domain.Round
	var unitID, executorID, items, notes sql.NullString
	var startedAt sql.NullTime
	var roundTime sql.NullTime

	if err := row.Scan(
		&round.RoundID,
		&round.TenantID,
		&round.RoundType,
		&unitID,
		&executorID,
		&roundTime,
		&startedAt,
		&items,
		&notes,
		&round.Status,
	); err != nil {
		return nil, err
	}
	round.UnitID = unitID.String
	round.ExecutorID = executorID.String
	round.Notes = notes.String
	if roundTime.Valid {
		round.RoundTime = roundTime.Time
	}
	if startedAt.Valid {
		round.StartedAt = &startedAt.Time
	}
	if items.Valid && items.String != "" {
		round.ItemsChecked = []byte(items.String)
	}
	return &round, nil
}

// GetRound 获取巡房记录
func (r *PostgresRoundsRepository) GetRound(ctx context.Context, tenantID, roundID string) (*domain.Round, error) {
	if tenantID == "" || roundID == "" {
		return nil, sql.ErrNoRows
	}
	query := roundSelectV2 + ` WHERE r.round_id = $1::UUID AND r.caregiver_id <<= $2::INET`
	row := r.db.QueryRowContext(ctx, query, roundID, tenantID)
	rd, err := scanRoundRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("round not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get round: %w", err)
	}
	return rd, nil
}

// ListRounds 批量查询巡房记录（支持过滤和分页）
func (r *PostgresRoundsRepository) ListRounds(ctx context.Context, tenantID string, filters *RoundFilters, page, size int) ([]*domain.Round, int, error) {
	if tenantID == "" {
		return []*domain.Round{}, 0, nil
	}
	// scope：tenant /48 via caregiver_id <<= tenant；同时 unit_prefix <<= tenant 兜底 (caregiver_id 在 ff01: subject ns 也在同 tenant /48 内)
	where := []string{"r.caregiver_id <<= $1::INET"}
	args := []any{tenantID}
	argN := 2

	if filters != nil {
		if filters.ExecutorID != "" {
			// executor_id 是 user_id UUID；转 hoa INET 后过滤 r.caregiver_id
			where = append(where, fmt.Sprintf("r.caregiver_id = (SELECT hoa FROM users WHERE user_id = $%d::UUID)", argN))
			args = append(args, filters.ExecutorID)
			argN++
		}
		if filters.UnitID != "" {
			where = append(where, fmt.Sprintf("r.unit_prefix = $%d::INET", argN))
			args = append(args, filters.UnitID)
			argN++
		}
		if filters.BranchPrefix != "" {
			// tenant-wide rounds (unit_prefix=NULL) 对所有 branch 可见
			where = append(where, fmt.Sprintf("(r.unit_prefix IS NULL OR r.unit_prefix <<= $%d::INET)", argN))
			args = append(args, filters.BranchPrefix)
			argN++
		}
		if filters.RoundType != "" {
			where = append(where, fmt.Sprintf("r.round_type = $%d", argN))
			args = append(args, filters.RoundType)
			argN++
		}
		if filters.Status != "" {
			where = append(where, fmt.Sprintf("r.status = $%d", argN))
			args = append(args, filters.Status)
			argN++
		}
		if filters.StartTime != nil {
			where = append(where, fmt.Sprintf("COALESCE(r.ended_at, r.started_at) >= $%d", argN))
			args = append(args, *filters.StartTime)
			argN++
		}
		if filters.EndTime != nil {
			where = append(where, fmt.Sprintf("COALESCE(r.ended_at, r.started_at) <= $%d", argN))
			args = append(args, *filters.EndTime)
			argN++
		}
	}

	// COUNT
	countQ := `SELECT COUNT(*) FROM rounds r WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count rounds: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	listArgs := append(args, size, offset)
	listQ := roundSelectV2 + `
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY COALESCE(r.ended_at, r.started_at) DESC
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)
	rows, err := r.db.QueryContext(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list rounds: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Round, 0)
	for rows.Next() {
		rd, err := scanRoundRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan round: %w", err)
		}
		out = append(out, rd)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate rounds: %w", err)
	}
	return out, total, nil
}

// CreateRound 创建巡房记录
// v2: executor_id UUID → 查 users.hoa 后落 caregiver_id INET；
//     items_checked 拆到 rounds_report_snapshot.snapshot JSONB。
func (r *PostgresRoundsRepository) CreateRound(ctx context.Context, tenantID string, round *domain.Round) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if round.ExecutorID == "" {
		return "", fmt.Errorf("executor_id is required")
	}
	if round.RoundType == "" {
		round.RoundType = "manual"
	}
	if round.Status == "" {
		round.Status = "completed"
	}
	if round.RoundTime.IsZero() {
		round.RoundTime = time.Now()
	}

	// 1. 查 user_id UUID → hoa INET（caregiver_id）
	var caregiverID sql.NullString
	if err := r.db.QueryRowContext(ctx,
		`SELECT host(hoa) FROM users WHERE user_id = $1::UUID AND hoa IS NOT NULL`,
		round.ExecutorID).Scan(&caregiverID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("executor_id %s has no hoa (only resident-tied users can create rounds in v2 schema)", round.ExecutorID)
		}
		return "", fmt.Errorf("lookup caregiver hoa: %w", err)
	}

	// 2. 事务：rounds INSERT + 可选 rounds_report_snapshot
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 准备 nullable 字段
	var unitPrefix interface{}
	if round.UnitID != "" {
		unitPrefix = round.UnitID
	}
	var notes interface{}
	if round.Notes != "" {
		notes = round.Notes
	}
	// started_at fallback: 若 caller 没给，用 round_time（完成时刻）
	startedAt := round.RoundTime
	if round.StartedAt != nil {
		startedAt = *round.StartedAt
	}
	// ended_at: 仅 status=completed/cancelled 时填
	var endedAt interface{}
	if round.Status == "completed" || round.Status == "cancelled" {
		endedAt = round.RoundTime
	}

	var roundID string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO rounds (caregiver_id, unit_prefix, started_at, ended_at, round_type, status, notes)
		VALUES ($1::INET, $2::INET, $3, $4, $5, $6, $7)
		RETURNING round_id::text
	`, caregiverID.String, unitPrefix, startedAt, endedAt, round.RoundType, round.Status, notes).Scan(&roundID); err != nil {
		return "", fmt.Errorf("failed to create round: %w", err)
	}

	// items_checked → rounds_report_snapshot
	if len(round.ItemsChecked) > 0 {
		// snapshot JSONB 用 {"items_checked": [...]} 形态包一层
		snapshot := fmt.Sprintf(`{"items_checked": %s}`, string(round.ItemsChecked))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO rounds_report_snapshot (round_id, snapshot)
			VALUES ($1::UUID, $2::JSONB)
		`, roundID, snapshot); err != nil {
			return "", fmt.Errorf("failed to insert rounds_report_snapshot: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return roundID, nil
}

// UpdateRound 更新巡房记录（v2: 仅支持 unit/started_at/notes/status 更新；不动 caregiver_id/round_type）
func (r *PostgresRoundsRepository) UpdateRound(ctx context.Context, tenantID, roundID string, round *domain.Round) error {
	if tenantID == "" || roundID == "" {
		return fmt.Errorf("tenant_id and round_id are required")
	}
	var unitPrefix interface{}
	if round.UnitID != "" {
		unitPrefix = round.UnitID
	}
	var notes interface{}
	if round.Notes != "" {
		notes = round.Notes
	}
	startedAt := round.RoundTime
	if round.StartedAt != nil {
		startedAt = *round.StartedAt
	}
	var endedAt interface{}
	if round.Status == "completed" || round.Status == "cancelled" {
		endedAt = round.RoundTime
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE rounds
		SET unit_prefix = $3::INET,
		    started_at = $4,
		    ended_at = $5,
		    status = $6,
		    notes = $7
		WHERE round_id = $1::UUID AND caregiver_id <<= $2::INET
	`, roundID, tenantID, unitPrefix, startedAt, endedAt, round.Status, notes)
	if err != nil {
		return fmt.Errorf("failed to update round: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("round not found")
	}
	return nil
}

// DeleteRound 删除巡房记录（rounds_report_snapshot ON DELETE CASCADE 自动跟删）
func (r *PostgresRoundsRepository) DeleteRound(ctx context.Context, tenantID, roundID string) error {
	if tenantID == "" || roundID == "" {
		return fmt.Errorf("tenant_id and round_id are required")
	}
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM rounds
		WHERE round_id = $1::UUID AND caregiver_id <<= $2::INET
	`, roundID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete round: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("round not found")
	}
	return nil
}

// SetRoundStatus 更新巡房记录状态（in_progress/completed/cancelled — v2 status check constraint 实际允许值）
func (r *PostgresRoundsRepository) SetRoundStatus(ctx context.Context, tenantID, roundID, status string) error {
	if tenantID == "" || roundID == "" {
		return fmt.Errorf("tenant_id and round_id are required")
	}
	switch status {
	case "in_progress", "completed", "cancelled", "draft":
		// 接受 draft 作 v1 兼容（FE 可能仍传 draft）
	default:
		return fmt.Errorf("invalid status: %s", status)
	}
	// completed/cancelled 时同步设 ended_at = now()
	var endedAt interface{}
	if status == "completed" || status == "cancelled" {
		endedAt = time.Now()
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE rounds
		SET status = $3, ended_at = COALESCE($4, ended_at)
		WHERE round_id = $1::UUID AND caregiver_id <<= $2::INET
	`, roundID, tenantID, status, endedAt)
	if err != nil {
		return fmt.Errorf("failed to set round status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("round not found")
	}
	return nil
}
