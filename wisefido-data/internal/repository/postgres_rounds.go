package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresRoundsRepository 巡房记录Repository实现（强类型版本）
type PostgresRoundsRepository struct {
	db *sql.DB
}

// NewPostgresRoundsRepository 创建巡房记录Repository
func NewPostgresRoundsRepository(db *sql.DB) *PostgresRoundsRepository {
	return &PostgresRoundsRepository{db: db}
}

// 确保实现了接口
var _ RoundsRepository = (*PostgresRoundsRepository)(nil)

// GetRound 获取巡房记录
func (r *PostgresRoundsRepository) GetRound(ctx context.Context, tenantID, roundID string) (*domain.Round, error) {
	if tenantID == "" || roundID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			round_id::text,
			tenant_id::text,
			round_type,
			unit_id::text,
			executor_id::text,
			round_time,
			notes,
			status
		FROM rounds
		WHERE tenant_id = $1 AND round_id = $2
	`

	var round domain.Round
	var unitID, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, roundID).Scan(
		&round.RoundID,
		&round.TenantID,
		&round.RoundType,
		&unitID,
		&round.ExecutorID,
		&round.RoundTime,
		&notes,
		&round.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("round not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get round: %w", err)
	}

	if unitID.Valid {
		round.UnitID = unitID.String
	}
	if notes.Valid {
		round.Notes = notes.String
	}

	return &round, nil
}

// ListRounds 批量查询巡房记录（支持过滤和分页）
func (r *PostgresRoundsRepository) ListRounds(ctx context.Context, tenantID string, filters *RoundFilters, page, size int) ([]*domain.Round, int, error) {
	if tenantID == "" {
		return []*domain.Round{}, 0, nil
	}

	where := []string{"tenant_id = $1"}
	args := []any{tenantID}
	argN := 2

	if filters != nil {
		if filters.ExecutorID != "" {
			where = append(where, fmt.Sprintf("executor_id = $%d", argN))
			args = append(args, filters.ExecutorID)
			argN++
		}
		if filters.UnitID != "" {
			where = append(where, fmt.Sprintf("unit_id = $%d", argN))
			args = append(args, filters.UnitID)
			argN++
		}
		if filters.RoundType != "" {
			where = append(where, fmt.Sprintf("round_type = $%d", argN))
			args = append(args, filters.RoundType)
			argN++
		}
		if filters.Status != "" {
			where = append(where, fmt.Sprintf("status = $%d", argN))
			args = append(args, filters.Status)
			argN++
		}
		if filters.StartTime != nil {
			where = append(where, fmt.Sprintf("round_time >= $%d", argN))
			args = append(args, *filters.StartTime)
			argN++
		}
		if filters.EndTime != nil {
			where = append(where, fmt.Sprintf("round_time <= $%d", argN))
			args = append(args, *filters.EndTime)
			argN++
		}
	}

	// 查询总数
	queryCount := `
		SELECT COUNT(*)
		FROM rounds
		WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count rounds: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// 查询列表
	argsList := append(args, size, offset)
	query := `
		SELECT 
			round_id::text,
			tenant_id::text,
			round_type,
			unit_id::text,
			executor_id::text,
			round_time,
			notes,
			status
		FROM rounds
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY round_time DESC
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list rounds: %w", err)
	}
	defer rows.Close()

	var rounds []*domain.Round
	for rows.Next() {
		var round domain.Round
		var unitID, notes sql.NullString

		if err := rows.Scan(
			&round.RoundID,
			&round.TenantID,
			&round.RoundType,
			&unitID,
			&round.ExecutorID,
			&round.RoundTime,
			&notes,
			&round.Status,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan round: %w", err)
		}

		if unitID.Valid {
			round.UnitID = unitID.String
		}
		if notes.Valid {
			round.Notes = notes.String
		}

		rounds = append(rounds, &round)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate rounds: %w", err)
	}

	return rounds, total, nil
}

// CreateRound 创建巡房记录
func (r *PostgresRoundsRepository) CreateRound(ctx context.Context, tenantID string, round *domain.Round) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if round.ExecutorID == "" {
		return "", fmt.Errorf("executor_id is required")
	}
	if round.RoundType == "" {
		round.RoundType = "location"
	}
	if round.Status == "" {
		round.Status = "completed"
	}
	if round.RoundTime.IsZero() {
		round.RoundTime = time.Now()
	}

	query := `
		INSERT INTO rounds (
			tenant_id,
			round_type,
			unit_id,
			executor_id,
			round_time,
			notes,
			status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING round_id::text
	`

	var unitID, notes interface{}
	if round.UnitID != "" {
		unitID = round.UnitID
	}
	if round.Notes != "" {
		notes = round.Notes
	}

	var roundID string
	err := r.db.QueryRowContext(ctx, query, tenantID, round.RoundType, unitID, round.ExecutorID,
		round.RoundTime, notes, round.Status).Scan(&roundID)
	if err != nil {
		return "", fmt.Errorf("failed to create round: %w", err)
	}

	return roundID, nil
}

// UpdateRound 更新巡房记录
func (r *PostgresRoundsRepository) UpdateRound(ctx context.Context, tenantID, roundID string, round *domain.Round) error {
	if tenantID == "" || roundID == "" {
		return fmt.Errorf("tenant_id and round_id are required")
	}

	query := `
		UPDATE rounds
		SET
			round_type = $3,
			unit_id = $4,
			executor_id = $5,
			round_time = $6,
			notes = $7,
			status = $8
		WHERE tenant_id = $1 AND round_id = $2
	`

	var unitID, notes interface{}
	if round.UnitID != "" {
		unitID = round.UnitID
	}
	if round.Notes != "" {
		notes = round.Notes
	}

	result, err := r.db.ExecContext(ctx, query, tenantID, roundID, round.RoundType, unitID,
		round.ExecutorID, round.RoundTime, notes, round.Status)
	if err != nil {
		return fmt.Errorf("failed to update round: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("round not found")
	}

	return nil
}

// DeleteRound 删除巡房记录
func (r *PostgresRoundsRepository) DeleteRound(ctx context.Context, tenantID, roundID string) error {
	if tenantID == "" || roundID == "" {
		return fmt.Errorf("tenant_id and round_id are required")
	}

	query := `
		DELETE FROM rounds
		WHERE tenant_id = $1 AND round_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, roundID)
	if err != nil {
		return fmt.Errorf("failed to delete round: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("round not found")
	}

	return nil
}

// SetRoundStatus 更新巡房记录状态
func (r *PostgresRoundsRepository) SetRoundStatus(ctx context.Context, tenantID, roundID, status string) error {
	if tenantID == "" || roundID == "" {
		return fmt.Errorf("tenant_id and round_id are required")
	}
	if status != "draft" && status != "completed" && status != "cancelled" {
		return fmt.Errorf("invalid status: %s", status)
	}

	query := `
		UPDATE rounds
		SET status = $3
		WHERE tenant_id = $1 AND round_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, roundID, status)
	if err != nil {
		return fmt.Errorf("failed to set round status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("round not found")
	}

	return nil
}

