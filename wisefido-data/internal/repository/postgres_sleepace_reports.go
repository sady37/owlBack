package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresSleepaceReportsRepository Sleepace 报告 Repository 实现
type PostgresSleepaceReportsRepository struct {
	db *sql.DB
}

// NewPostgresSleepaceReportsRepository 创建 Sleepace 报告 Repository
func NewPostgresSleepaceReportsRepository(db *sql.DB) *PostgresSleepaceReportsRepository {
	return &PostgresSleepaceReportsRepository{db: db}
}

// 确保实现了接口
var _ SleepaceReportsRepository = (*PostgresSleepaceReportsRepository)(nil)

// GetReport 根据 device_id 和 date 获取报告详情
func (r *PostgresSleepaceReportsRepository) GetReport(ctx context.Context, tenantID, deviceID string, date int) (*domain.SleepaceReport, error) {
	if tenantID == "" || deviceID == "" || date == 0 {
		return nil, fmt.Errorf("tenant_id, device_id and date are required")
	}

	query := `
		SELECT 
			report_id::text,
			tenant_id::text,
			device_id::text,
			COALESCE(device_code, '') AS device_code,
			device_uid,
			record_count,
			start_time,
			end_time,
			date,
			stop_mode,
			time_step,
			timezone,
			COALESCE(sleep_state, '') as sleep_state,
			COALESCE(report, '') as report,
			EXTRACT(EPOCH FROM created_at)::bigint as created_at,
			EXTRACT(EPOCH FROM updated_at)::bigint as updated_at
		FROM sleepace_report
		WHERE tenant_id = $1::uuid
		  AND device_id = $2::uuid
		  AND date = $3
	`

	var report domain.SleepaceReport
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceID, date).Scan(
		&report.ReportID,
		&report.TenantID,
		&report.DeviceID,
		&report.DeviceCode,
		&report.DeviceUID,
		&report.RecordCount,
		&report.StartTime,
		&report.EndTime,
		&report.Date,
		&report.StopMode,
		&report.TimeStep,
		&report.Timezone,
		&report.SleepState,
		&report.Report,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 报告不存在，返回 nil
		}
		return nil, fmt.Errorf("failed to get sleepace report: %w", err)
	}

	return &report, nil
}

// ListReports 查询报告列表（支持分页）
func (r *PostgresSleepaceReportsRepository) ListReports(ctx context.Context, tenantID, deviceID string, startDate, endDate int, page, size int) ([]*domain.SleepaceReport, int, error) {
	if tenantID == "" || deviceID == "" {
		return nil, 0, fmt.Errorf("tenant_id and device_id are required")
	}

	// 计算总数
	countQuery := `
		SELECT COUNT(*)
		FROM sleepace_report
		WHERE tenant_id = $1::uuid
		  AND device_id = $2::uuid
		  AND date >= $3
		  AND date <= $4
	`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, tenantID, deviceID, startDate, endDate).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sleepace reports: %w", err)
	}

	// 分页参数
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	// 查询数据
	query := `
		SELECT 
			report_id::text,
			tenant_id::text,
			device_id::text,
			COALESCE(device_code, '') AS device_code,
			device_uid,
			record_count,
			start_time,
			end_time,
			date,
			stop_mode,
			time_step,
			timezone,
			COALESCE(sleep_state, '') as sleep_state,
			COALESCE(report, '') as report,
			EXTRACT(EPOCH FROM created_at)::bigint as created_at,
			EXTRACT(EPOCH FROM updated_at)::bigint as updated_at
		FROM sleepace_report
		WHERE tenant_id = $1::uuid
		  AND device_id = $2::uuid
		  AND date >= $3
		  AND date <= $4
		ORDER BY date DESC
		LIMIT $5 OFFSET $6
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, deviceID, startDate, endDate, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sleepace reports: %w", err)
	}
	defer rows.Close()

	reports := make([]*domain.SleepaceReport, 0)
	for rows.Next() {
		var report domain.SleepaceReport
		err := rows.Scan(
			&report.ReportID,
			&report.TenantID,
			&report.DeviceID,
			&report.DeviceCode,
			&report.DeviceUID,
			&report.RecordCount,
			&report.StartTime,
			&report.EndTime,
			&report.Date,
			&report.StopMode,
			&report.TimeStep,
			&report.Timezone,
			&report.SleepState,
			&report.Report,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan sleepace report: %w", err)
		}
		reports = append(reports, &report)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate sleepace reports: %w", err)
	}

	return reports, total, nil
}

// GetValidDates 获取设备的所有有效日期列表
func (r *PostgresSleepaceReportsRepository) GetValidDates(ctx context.Context, tenantID, deviceID string) ([]int, error) {
	if tenantID == "" || deviceID == "" {
		return nil, fmt.Errorf("tenant_id and device_id are required")
	}

	query := `
		SELECT date
		FROM sleepace_report
		WHERE tenant_id = $1::uuid
		  AND device_id = $2::uuid
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid dates: %w", err)
	}
	defer rows.Close()

	dates := make([]int, 0)
	for rows.Next() {
		var date int
		if err := rows.Scan(&date); err != nil {
			return nil, fmt.Errorf("failed to scan date: %w", err)
		}
		dates = append(dates, date)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dates: %w", err)
	}

	return dates, nil
}

// ListReportsAllInRange 返回 [startDate,endDate] 内全部报告，按 date ASC
func (r *PostgresSleepaceReportsRepository) ListReportsAllInRange(ctx context.Context, tenantID, deviceID string, startDate, endDate int) ([]*domain.SleepaceReport, error) {
	if tenantID == "" || deviceID == "" || startDate == 0 || endDate == 0 {
		return nil, fmt.Errorf("tenant_id, device_id, start_date and end_date are required")
	}
	if startDate > endDate {
		return nil, fmt.Errorf("start_date must be <= end_date")
	}
	query := `
		SELECT 
			report_id::text,
			tenant_id::text,
			device_id::text,
			COALESCE(device_code, '') AS device_code,
			device_uid,
			record_count,
			start_time,
			end_time,
			date,
			stop_mode,
			time_step,
			timezone,
			COALESCE(sleep_state, '') as sleep_state,
			COALESCE(report, '') as report,
			EXTRACT(EPOCH FROM created_at)::bigint as created_at,
			EXTRACT(EPOCH FROM updated_at)::bigint as updated_at
		FROM sleepace_report
		WHERE tenant_id = $1::uuid
		  AND device_id = $2::uuid
		  AND date >= $3
		  AND date <= $4
		ORDER BY date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, deviceID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list sleepace reports in range: %w", err)
	}
	defer rows.Close()
	out := make([]*domain.SleepaceReport, 0)
	for rows.Next() {
		var report domain.SleepaceReport
		if err := rows.Scan(
			&report.ReportID,
			&report.TenantID,
			&report.DeviceID,
			&report.DeviceCode,
			&report.DeviceUID,
			&report.RecordCount,
			&report.StartTime,
			&report.EndTime,
			&report.Date,
			&report.StopMode,
			&report.TimeStep,
			&report.Timezone,
			&report.SleepState,
			&report.Report,
			&report.CreatedAt,
			&report.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sleepace report: %w", err)
		}
		out = append(out, &report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sleepace reports: %w", err)
	}
	return out, nil
}

// GetValidDatesInRange 获取 [startDate,endDate] 内已有报告的日期
func (r *PostgresSleepaceReportsRepository) GetValidDatesInRange(ctx context.Context, tenantID, deviceID string, startDate, endDate int) ([]int, error) {
	if tenantID == "" || deviceID == "" || startDate == 0 || endDate == 0 {
		return nil, fmt.Errorf("tenant_id, device_id, start_date and end_date are required")
	}
	if startDate > endDate {
		return nil, fmt.Errorf("start_date must be <= end_date")
	}
	query := `
		SELECT date
		FROM sleepace_report
		WHERE tenant_id = $1::uuid
		  AND device_id = $2::uuid
		  AND date >= $3
		  AND date <= $4
		ORDER BY date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, deviceID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid dates in range: %w", err)
	}
	defer rows.Close()
	dates := make([]int, 0)
	for rows.Next() {
		var d int
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("failed to scan date: %w", err)
		}
		dates = append(dates, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dates: %w", err)
	}
	return dates, nil
}

func (r *PostgresSleepaceReportsRepository) GetDeviceIDByDeviceUID(ctx context.Context, tenantID, deviceUID string) (string, error) {
	if tenantID == "" || deviceUID == "" {
		return "", fmt.Errorf("tenant_id and device_uid are required")
	}

	query := `
		SELECT device_id::text
		FROM devices
		WHERE tenant_id = $1::uuid
		  AND device_uid = $2
		  AND status <> 'disabled'
		LIMIT 1
	`

	var deviceID string
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceUID).Scan(&deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("device not found: device_uid=%s", deviceUID)
		}
		return "", fmt.Errorf("failed to get device_id by device_uid: %w", err)
	}

	return deviceID, nil
}

func (r *PostgresSleepaceReportsRepository) SaveReport(ctx context.Context, tenantID string, report *domain.SleepaceReport) error {
	if tenantID == "" || report == nil {
		return fmt.Errorf("tenant_id and report are required")
	}

	deviceID := report.DeviceID
	if deviceID == "" && report.DeviceUID != "" {
		var err error
		deviceID, err = r.GetDeviceIDByDeviceUID(ctx, tenantID, report.DeviceUID)
		if err != nil {
			return fmt.Errorf("failed to get device_id from device_uid: %w", err)
		}
		report.DeviceID = deviceID
	}

	if deviceID == "" {
		return fmt.Errorf("device_id is required (either provide device_id or device_uid)")
	}

	// 检查是否已存在
	existsQuery := `
		SELECT EXISTS(
			SELECT 1
			FROM sleepace_report
			WHERE tenant_id = $1::uuid
			  AND device_id = $2::uuid
			  AND date = $3
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, existsQuery, tenantID, report.DeviceID, report.Date).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if report exists: %w", err)
	}

	now := time.Now().Unix()

	if exists {
		// 更新
		updateQuery := `
			UPDATE sleepace_report
			SET device_uid = $4,
				device_code = $5,
				record_count = $6,
				start_time = $7,
				end_time = $8,
				stop_mode = $9,
				time_step = $10,
				timezone = $11,
				sleep_state = $12,
				report = $13,
				updated_at = $14
			WHERE tenant_id = $1::uuid
			  AND device_id = $2::uuid
			  AND date = $3
		`
		_, err = r.db.ExecContext(ctx, updateQuery,
			tenantID,
			deviceID,
			report.Date,
			report.DeviceUID,
			report.DeviceCode,
			report.RecordCount,
			report.StartTime,
			report.EndTime,
			report.StopMode,
			report.TimeStep,
			report.Timezone,
			report.SleepState,
			report.Report,
			time.Unix(now, 0),
		)
		if err != nil {
			return fmt.Errorf("failed to update sleepace report: %w", err)
		}
	} else {
		// 插入
		insertQuery := `
			INSERT INTO sleepace_report (
				tenant_id,
				device_id,
				device_uid,
				device_code,
				record_count,
				start_time,
				end_time,
				date,
				stop_mode,
				time_step,
				timezone,
				sleep_state,
				report,
				created_at,
				updated_at
			) VALUES (
				$1::uuid,
				$2::uuid,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				$11,
				$12,
				$13,
				$14,
				$15
			)
			RETURNING report_id::text, EXTRACT(EPOCH FROM created_at)::bigint, EXTRACT(EPOCH FROM updated_at)::bigint
		`
		err = r.db.QueryRowContext(ctx, insertQuery,
			tenantID,
			deviceID,
			report.DeviceUID,
			report.DeviceCode,
			report.RecordCount,
			report.StartTime,
			report.EndTime,
			report.Date,
			report.StopMode,
			report.TimeStep,
			report.Timezone,
			report.SleepState,
			report.Report,
			time.Unix(now, 0),
			time.Unix(now, 0),
		).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert sleepace report: %w", err)
		}
	}

	return nil
}

