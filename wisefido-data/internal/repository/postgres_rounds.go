package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresRoundsRepository 巡房记录 Repository
//
// rounds.user_id UUID FK users(user_id) — 任意 user 都可记录
// rounds.tenant_id INET — create 时从 users.tenant_id /48 snapshot
// snapshot JSONB blob (rounds_report_snapshot.snapshot) 含 items_checked / rows / schema_version 等
type PostgresRoundsRepository struct {
	db *sql.DB
}

func NewPostgresRoundsRepository(db *sql.DB) *PostgresRoundsRepository {
	return &PostgresRoundsRepository{db: db}
}

var _ RoundsRepository = (*PostgresRoundsRepository)(nil)

const roundSelectV2 = `
	SELECT
	    r.round_id::text                                       AS round_id,
	    host(r.tenant_id)||'/'||masklen(r.tenant_id)           AS tenant_id,
	    r.user_id::text                                        AS user_id,
	    r.round_type                                           AS round_type,
	    COALESCE(host(r.unit_prefix)||'/'||masklen(r.unit_prefix), '') AS unit_id,
	    r.started_at                                           AS started_at,
	    r.ended_at                                             AS ended_at,
	    COALESCE(r.notes, '')                                  AS notes,
	    r.status                                               AS status,
	    COALESCE(rrs.snapshot::text, '')                       AS snapshot,
	    COALESCE(NULLIF(u.nickname, ''), u.full_name, '')      AS executor_display,
	    COALESCE(t.tenant_name, '')                            AS facility_name,
	    COALESCE(un.unit_name, '')                             AS unit_name
	FROM rounds r
	LEFT JOIN rounds_report_snapshot rrs ON rrs.round_id = r.round_id
	LEFT JOIN users u   ON u.user_id = r.user_id
	LEFT JOIN tenants t ON t.tenant_id = r.tenant_id
	LEFT JOIN units un  ON un.unit_id = r.unit_prefix
`

func scanRoundRow(row interface{ Scan(...any) error }) (*domain.Round, error) {
	var round domain.Round
	var unitID, snapshot, notes, executorDisplay, facilityName, unitName sql.NullString
	var startedAt, endedAt sql.NullTime

	if err := row.Scan(
		&round.RoundID,
		&round.TenantID,
		&round.UserID,
		&round.RoundType,
		&unitID,
		&startedAt,
		&endedAt,
		&notes,
		&round.Status,
		&snapshot,
		&executorDisplay,
		&facilityName,
		&unitName,
	); err != nil {
		return nil, err
	}
	round.UnitID = unitID.String
	round.Notes = notes.String
	round.ExecutorDisplay = executorDisplay.String
	round.FacilityName = facilityName.String
	round.UnitName = unitName.String
	if startedAt.Valid {
		round.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		round.EndedAt = &endedAt.Time
	}
	if snapshot.Valid && snapshot.String != "" {
		round.Snapshot = []byte(snapshot.String)
	}
	return &round, nil
}

func (r *PostgresRoundsRepository) GetRound(ctx context.Context, tenantID, roundID string) (*domain.Round, error) {
	if tenantID == "" || roundID == "" {
		return nil, sql.ErrNoRows
	}
	query := roundSelectV2 + ` WHERE r.round_id = $1::UUID AND r.tenant_id <<= $2::INET`
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

func (r *PostgresRoundsRepository) ListRounds(ctx context.Context, tenantID string, filters *RoundFilters, page, size int) ([]*domain.Round, int, error) {
	if tenantID == "" {
		return []*domain.Round{}, 0, nil
	}
	where := []string{"r.tenant_id <<= $1::INET"}
	args := []any{tenantID}
	argN := 2

	if filters != nil {
		if filters.UserID != "" {
			where = append(where, fmt.Sprintf("r.user_id = $%d::UUID", argN))
			args = append(args, filters.UserID)
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

// CreateRound 创建巡房记录。userID = session 的 users.user_id (UUID)。
// tenant_id 从 users.tenant_id snapshot — user 调岗不影响历史归属。
// round.Snapshot 是已 merge 好的 JSONB blob，由 service 层组装。
func (r *PostgresRoundsRepository) CreateRound(ctx context.Context, userID string, round *domain.Round) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("userID is required")
	}
	if round.RoundType == "" {
		round.RoundType = "manual"
	}
	if round.Status == "" {
		round.Status = "completed"
	}

	var tenantPrefix sql.NullString
	if err := r.db.QueryRowContext(ctx,
		`SELECT host(tenant_id)||'/'||masklen(tenant_id) FROM users WHERE user_id = $1::UUID AND tenant_id IS NOT NULL`,
		userID).Scan(&tenantPrefix); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user %s has no tenant_id (platform-level users cannot create tenant-scoped rounds)", userID)
		}
		return "", fmt.Errorf("lookup user tenant: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var unitPrefix interface{}
	if round.UnitID != "" {
		unitPrefix = round.UnitID
	}
	var notes interface{}
	if round.Notes != "" {
		notes = round.Notes
	}
	startedAt := time.Now()
	if round.StartedAt != nil {
		startedAt = *round.StartedAt
	}
	var endedAt interface{}
	if round.Status == "completed" || round.Status == "aborted" {
		if round.EndedAt != nil {
			endedAt = *round.EndedAt
		} else {
			endedAt = time.Now()
		}
	}

	var roundID string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO rounds (tenant_id, user_id, unit_prefix, started_at, ended_at, round_type, status, notes)
		VALUES ($1::INET, $2::UUID, $3::INET, $4, $5, $6, $7, $8)
		RETURNING round_id::text
	`, tenantPrefix.String, userID, unitPrefix, startedAt, endedAt, round.RoundType, round.Status, notes).Scan(&roundID); err != nil {
		return "", fmt.Errorf("failed to create round: %w", err)
	}

	if len(round.Snapshot) > 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO rounds_report_snapshot (round_id, snapshot)
			VALUES ($1::UUID, $2::JSONB)
		`, roundID, string(round.Snapshot)); err != nil {
			return "", fmt.Errorf("failed to insert rounds_report_snapshot: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return roundID, nil
}
