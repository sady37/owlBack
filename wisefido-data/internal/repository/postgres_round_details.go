package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresRoundDetailsRepository 巡房详细记录Repository实现（强类型版本）
type PostgresRoundDetailsRepository struct {
	db *sql.DB
}

// NewPostgresRoundDetailsRepository 创建巡房详细记录Repository
func NewPostgresRoundDetailsRepository(db *sql.DB) *PostgresRoundDetailsRepository {
	return &PostgresRoundDetailsRepository{db: db}
}

// 确保实现了接口
var _ RoundDetailsRepository = (*PostgresRoundDetailsRepository)(nil)

// GetRoundDetail 获取巡房详细记录
func (r *PostgresRoundDetailsRepository) GetRoundDetail(ctx context.Context, tenantID, detailID string) (*domain.RoundDetail, error) {
	if tenantID == "" || detailID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			detail_id::text,
			tenant_id::text,
			round_id::text,
			resident_id::text,
			bed_id::text,
			unit_id::text,
			bed_status,
			sleep_state_snomed_code,
			sleep_state_display,
			heart_rate,
			respiratory_rate,
			posture_snomed_code,
			posture_display,
			data_timestamp,
			notes
		FROM round_details
		WHERE tenant_id = $1 AND detail_id = $2
	`

	var detail domain.RoundDetail
	var bedID, unitID, bedStatus sql.NullString
	var sleepStateSNOMEDCode, sleepStateDisplay sql.NullString
	var heartRate, respiratoryRate sql.NullInt64
	var postureSNOMEDCode, postureDisplay sql.NullString
	var dataTimestamp sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, detailID).Scan(
		&detail.DetailID,
		&detail.TenantID,
		&detail.RoundID,
		&detail.ResidentID,
		&bedID,
		&unitID,
		&bedStatus,
		&sleepStateSNOMEDCode,
		&sleepStateDisplay,
		&heartRate,
		&respiratoryRate,
		&postureSNOMEDCode,
		&postureDisplay,
		&dataTimestamp,
		&notes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("round detail not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get round detail: %w", err)
	}

	if bedID.Valid {
		detail.BedID = bedID.String
	}
	if unitID.Valid {
		detail.UnitID = unitID.String
	}
	if bedStatus.Valid {
		detail.BedStatus = bedStatus.String
	}
	if sleepStateSNOMEDCode.Valid {
		detail.SleepStateSNOMEDCode = sleepStateSNOMEDCode.String
	}
	if sleepStateDisplay.Valid {
		detail.SleepStateDisplay = sleepStateDisplay.String
	}
	if heartRate.Valid {
		hr := int(heartRate.Int64)
		detail.HeartRate = &hr
	}
	if respiratoryRate.Valid {
		rr := int(respiratoryRate.Int64)
		detail.RespiratoryRate = &rr
	}
	if postureSNOMEDCode.Valid {
		detail.PostureSNOMEDCode = postureSNOMEDCode.String
	}
	if postureDisplay.Valid {
		detail.PostureDisplay = postureDisplay.String
	}
	if dataTimestamp.Valid {
		detail.DataTimestamp = &dataTimestamp.Time
	}
	if notes.Valid {
		detail.Notes = notes.String
	}

	return &detail, nil
}

// GetRoundDetailsByRound 获取某个巡房记录的所有详细记录
func (r *PostgresRoundDetailsRepository) GetRoundDetailsByRound(ctx context.Context, tenantID, roundID string) ([]*domain.RoundDetail, error) {
	if tenantID == "" || roundID == "" {
		return []*domain.RoundDetail{}, nil
	}

	query := `
		SELECT 
			detail_id::text,
			tenant_id::text,
			round_id::text,
			resident_id::text,
			bed_id::text,
			unit_id::text,
			bed_status,
			sleep_state_snomed_code,
			sleep_state_display,
			heart_rate,
			respiratory_rate,
			posture_snomed_code,
			posture_display,
			data_timestamp,
			notes
		FROM round_details
		WHERE tenant_id = $1 AND round_id = $2
		ORDER BY resident_id
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, roundID)
	if err != nil {
		return nil, fmt.Errorf("failed to get round details by round: %w", err)
	}
	defer rows.Close()

	var details []*domain.RoundDetail
	for rows.Next() {
		var detail domain.RoundDetail
		var bedID, unitID, bedStatus sql.NullString
		var sleepStateSNOMEDCode, sleepStateDisplay sql.NullString
		var heartRate, respiratoryRate sql.NullInt64
		var postureSNOMEDCode, postureDisplay sql.NullString
		var dataTimestamp sql.NullTime
		var notes sql.NullString

		if err := rows.Scan(
			&detail.DetailID,
			&detail.TenantID,
			&detail.RoundID,
			&detail.ResidentID,
			&bedID,
			&unitID,
			&bedStatus,
			&sleepStateSNOMEDCode,
			&sleepStateDisplay,
			&heartRate,
			&respiratoryRate,
			&postureSNOMEDCode,
			&postureDisplay,
			&dataTimestamp,
			&notes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan round detail: %w", err)
		}

		if bedID.Valid {
			detail.BedID = bedID.String
		}
		if unitID.Valid {
			detail.UnitID = unitID.String
		}
		if bedStatus.Valid {
			detail.BedStatus = bedStatus.String
		}
		if sleepStateSNOMEDCode.Valid {
			detail.SleepStateSNOMEDCode = sleepStateSNOMEDCode.String
		}
		if sleepStateDisplay.Valid {
			detail.SleepStateDisplay = sleepStateDisplay.String
		}
		if heartRate.Valid {
			hr := int(heartRate.Int64)
			detail.HeartRate = &hr
		}
		if respiratoryRate.Valid {
			rr := int(respiratoryRate.Int64)
			detail.RespiratoryRate = &rr
		}
		if postureSNOMEDCode.Valid {
			detail.PostureSNOMEDCode = postureSNOMEDCode.String
		}
		if postureDisplay.Valid {
			detail.PostureDisplay = postureDisplay.String
		}
		if dataTimestamp.Valid {
			detail.DataTimestamp = &dataTimestamp.Time
		}
		if notes.Valid {
			detail.Notes = notes.String
		}

		details = append(details, &detail)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate round details: %w", err)
	}

	return details, nil
}

// GetRoundDetailsByResident 获取某个住户的所有巡房详细记录（支持分页）
func (r *PostgresRoundDetailsRepository) GetRoundDetailsByResident(ctx context.Context, tenantID, residentID string, filters *RoundDetailFilters, page, size int) ([]*domain.RoundDetail, int, error) {
	if tenantID == "" || residentID == "" {
		return []*domain.RoundDetail{}, 0, nil
	}

	where := []string{"tenant_id = $1", "resident_id = $2"}
	args := []any{tenantID, residentID}
	argN := 3

	if filters != nil {
		if filters.BedID != "" {
			where = append(where, fmt.Sprintf("bed_id = $%d", argN))
			args = append(args, filters.BedID)
			argN++
		}
		if filters.UnitID != "" {
			where = append(where, fmt.Sprintf("unit_id = $%d", argN))
			args = append(args, filters.UnitID)
			argN++
		}
		if filters.BedStatus != "" {
			where = append(where, fmt.Sprintf("bed_status = $%d", argN))
			args = append(args, filters.BedStatus)
			argN++
		}
		if filters.StartTime != nil {
			where = append(where, fmt.Sprintf("data_timestamp >= $%d", argN))
			args = append(args, *filters.StartTime)
			argN++
		}
		if filters.EndTime != nil {
			where = append(where, fmt.Sprintf("data_timestamp <= $%d", argN))
			args = append(args, *filters.EndTime)
			argN++
		}
	}

	// 查询总数
	queryCount := `
		SELECT COUNT(*)
		FROM round_details
		WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count round details: %w", err)
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
			detail_id::text,
			tenant_id::text,
			round_id::text,
			resident_id::text,
			bed_id::text,
			unit_id::text,
			bed_status,
			sleep_state_snomed_code,
			sleep_state_display,
			heart_rate,
			respiratory_rate,
			posture_snomed_code,
			posture_display,
			data_timestamp,
			notes
		FROM round_details
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY data_timestamp DESC
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list round details: %w", err)
	}
	defer rows.Close()

	var details []*domain.RoundDetail
	for rows.Next() {
		var detail domain.RoundDetail
		var bedID, unitID, bedStatus sql.NullString
		var sleepStateSNOMEDCode, sleepStateDisplay sql.NullString
		var heartRate, respiratoryRate sql.NullInt64
		var postureSNOMEDCode, postureDisplay sql.NullString
		var dataTimestamp sql.NullTime
		var notes sql.NullString

		if err := rows.Scan(
			&detail.DetailID,
			&detail.TenantID,
			&detail.RoundID,
			&detail.ResidentID,
			&bedID,
			&unitID,
			&bedStatus,
			&sleepStateSNOMEDCode,
			&sleepStateDisplay,
			&heartRate,
			&respiratoryRate,
			&postureSNOMEDCode,
			&postureDisplay,
			&dataTimestamp,
			&notes,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan round detail: %w", err)
		}

		if bedID.Valid {
			detail.BedID = bedID.String
		}
		if unitID.Valid {
			detail.UnitID = unitID.String
		}
		if bedStatus.Valid {
			detail.BedStatus = bedStatus.String
		}
		if sleepStateSNOMEDCode.Valid {
			detail.SleepStateSNOMEDCode = sleepStateSNOMEDCode.String
		}
		if sleepStateDisplay.Valid {
			detail.SleepStateDisplay = sleepStateDisplay.String
		}
		if heartRate.Valid {
			hr := int(heartRate.Int64)
			detail.HeartRate = &hr
		}
		if respiratoryRate.Valid {
			rr := int(respiratoryRate.Int64)
			detail.RespiratoryRate = &rr
		}
		if postureSNOMEDCode.Valid {
			detail.PostureSNOMEDCode = postureSNOMEDCode.String
		}
		if postureDisplay.Valid {
			detail.PostureDisplay = postureDisplay.String
		}
		if dataTimestamp.Valid {
			detail.DataTimestamp = &dataTimestamp.Time
		}
		if notes.Valid {
			detail.Notes = notes.String
		}

		details = append(details, &detail)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate round details: %w", err)
	}

	return details, total, nil
}

// UpsertRoundDetail 创建或更新巡房详细记录
func (r *PostgresRoundDetailsRepository) UpsertRoundDetail(ctx context.Context, tenantID, roundID string, roundDetail *domain.RoundDetail) (string, error) {
	if tenantID == "" || roundID == "" {
		return "", fmt.Errorf("tenant_id and round_id are required")
	}
	if roundDetail.ResidentID == "" {
		return "", fmt.Errorf("resident_id is required")
	}

	query := `
		INSERT INTO round_details (
			tenant_id,
			round_id,
			resident_id,
			bed_id,
			unit_id,
			bed_status,
			sleep_state_snomed_code,
			sleep_state_display,
			heart_rate,
			respiratory_rate,
			posture_snomed_code,
			posture_display,
			data_timestamp,
			notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (round_id, resident_id) DO UPDATE SET
			bed_id = EXCLUDED.bed_id,
			unit_id = EXCLUDED.unit_id,
			bed_status = EXCLUDED.bed_status,
			sleep_state_snomed_code = EXCLUDED.sleep_state_snomed_code,
			sleep_state_display = EXCLUDED.sleep_state_display,
			heart_rate = EXCLUDED.heart_rate,
			respiratory_rate = EXCLUDED.respiratory_rate,
			posture_snomed_code = EXCLUDED.posture_snomed_code,
			posture_display = EXCLUDED.posture_display,
			data_timestamp = EXCLUDED.data_timestamp,
			notes = EXCLUDED.notes
		RETURNING detail_id::text
	`

	var bedID, unitID, bedStatus interface{}
	if roundDetail.BedID != "" {
		bedID = roundDetail.BedID
	}
	if roundDetail.UnitID != "" {
		unitID = roundDetail.UnitID
	}
	if roundDetail.BedStatus != "" {
		bedStatus = roundDetail.BedStatus
	}

	var sleepStateSNOMEDCode, sleepStateDisplay interface{}
	if roundDetail.SleepStateSNOMEDCode != "" {
		sleepStateSNOMEDCode = roundDetail.SleepStateSNOMEDCode
	}
	if roundDetail.SleepStateDisplay != "" {
		sleepStateDisplay = roundDetail.SleepStateDisplay
	}

	var heartRate, respiratoryRate interface{}
	if roundDetail.HeartRate != nil {
		heartRate = *roundDetail.HeartRate
	}
	if roundDetail.RespiratoryRate != nil {
		respiratoryRate = *roundDetail.RespiratoryRate
	}

	var postureSNOMEDCode, postureDisplay interface{}
	if roundDetail.PostureSNOMEDCode != "" {
		postureSNOMEDCode = roundDetail.PostureSNOMEDCode
	}
	if roundDetail.PostureDisplay != "" {
		postureDisplay = roundDetail.PostureDisplay
	}

	var dataTimestamp interface{}
	if roundDetail.DataTimestamp != nil {
		dataTimestamp = *roundDetail.DataTimestamp
	}

	var notes interface{}
	if roundDetail.Notes != "" {
		notes = roundDetail.Notes
	}

	var detailID string
	err := r.db.QueryRowContext(ctx, query, tenantID, roundID, roundDetail.ResidentID,
		bedID, unitID, bedStatus, sleepStateSNOMEDCode, sleepStateDisplay,
		heartRate, respiratoryRate, postureSNOMEDCode, postureDisplay,
		dataTimestamp, notes).Scan(&detailID)
	if err != nil {
		return "", fmt.Errorf("failed to upsert round detail: %w", err)
	}

	return detailID, nil
}

// UpdateRoundDetail 更新巡房详细记录
func (r *PostgresRoundDetailsRepository) UpdateRoundDetail(ctx context.Context, tenantID, detailID string, roundDetail *domain.RoundDetail) error {
	if tenantID == "" || detailID == "" {
		return fmt.Errorf("tenant_id and detail_id are required")
	}

	query := `
		UPDATE round_details
		SET
			bed_id = $3,
			unit_id = $4,
			bed_status = $5,
			sleep_state_snomed_code = $6,
			sleep_state_display = $7,
			heart_rate = $8,
			respiratory_rate = $9,
			posture_snomed_code = $10,
			posture_display = $11,
			data_timestamp = $12,
			notes = $13
		WHERE tenant_id = $1 AND detail_id = $2
	`

	var bedID, unitID, bedStatus interface{}
	if roundDetail.BedID != "" {
		bedID = roundDetail.BedID
	}
	if roundDetail.UnitID != "" {
		unitID = roundDetail.UnitID
	}
	if roundDetail.BedStatus != "" {
		bedStatus = roundDetail.BedStatus
	}

	var sleepStateSNOMEDCode, sleepStateDisplay interface{}
	if roundDetail.SleepStateSNOMEDCode != "" {
		sleepStateSNOMEDCode = roundDetail.SleepStateSNOMEDCode
	}
	if roundDetail.SleepStateDisplay != "" {
		sleepStateDisplay = roundDetail.SleepStateDisplay
	}

	var heartRate, respiratoryRate interface{}
	if roundDetail.HeartRate != nil {
		heartRate = *roundDetail.HeartRate
	}
	if roundDetail.RespiratoryRate != nil {
		respiratoryRate = *roundDetail.RespiratoryRate
	}

	var postureSNOMEDCode, postureDisplay interface{}
	if roundDetail.PostureSNOMEDCode != "" {
		postureSNOMEDCode = roundDetail.PostureSNOMEDCode
	}
	if roundDetail.PostureDisplay != "" {
		postureDisplay = roundDetail.PostureDisplay
	}

	var dataTimestamp interface{}
	if roundDetail.DataTimestamp != nil {
		dataTimestamp = *roundDetail.DataTimestamp
	}

	var notes interface{}
	if roundDetail.Notes != "" {
		notes = roundDetail.Notes
	}

	result, err := r.db.ExecContext(ctx, query, tenantID, detailID, bedID, unitID, bedStatus,
		sleepStateSNOMEDCode, sleepStateDisplay, heartRate, respiratoryRate,
		postureSNOMEDCode, postureDisplay, dataTimestamp, notes)
	if err != nil {
		return fmt.Errorf("failed to update round detail: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("round detail not found")
	}

	return nil
}

// DeleteRoundDetail 删除巡房详细记录
func (r *PostgresRoundDetailsRepository) DeleteRoundDetail(ctx context.Context, tenantID, detailID string) error {
	if tenantID == "" || detailID == "" {
		return fmt.Errorf("tenant_id and detail_id are required")
	}

	query := `
		DELETE FROM round_details
		WHERE tenant_id = $1 AND detail_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, detailID)
	if err != nil {
		return fmt.Errorf("failed to delete round detail: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("round detail not found")
	}

	return nil
}

