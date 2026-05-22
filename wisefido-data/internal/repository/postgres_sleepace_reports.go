package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresSleepaceReportsRepository Sleepace 报告 Repository 实现（v2 schema）。
//
// v2 sleepace_report 表与 v1 的差异：
//   - tenant_id 列移除 — 派生自 device_addr 前 /48 prefix
//   - device_uid / device_code 列移除 — 通过 JOIN device_factory_meta 获取
//   - start_time / end_time 列重命名为 start_time_ms / end_time_ms (BIGINT, 毫秒)
//   - report 列重命名为 report_data
//   - 新增 device_addr INET (canonical /128 IPv6) / resident_id INET (resident hoa) / metadata JSONB
//
// 业务接口保持 v1 形（domain.SleepaceReport），repo 内部做 v1⇄v2 字段映射；
// 上层 caller 仍传 tenantID（INET /48 CIDR 字符串）+ deviceID UUID + date 等。
type PostgresSleepaceReportsRepository struct {
	db *sql.DB
}

// NewPostgresSleepaceReportsRepository 创建 Sleepace 报告 Repository
func NewPostgresSleepaceReportsRepository(db *sql.DB) *PostgresSleepaceReportsRepository {
	return &PostgresSleepaceReportsRepository{db: db}
}

// 确保实现了接口
var _ SleepaceReportsRepository = (*PostgresSleepaceReportsRepository)(nil)

// sleepaceReportSelectColumns Phase 2 一刀切：sleepace_report.device_uid + device_addr。
const sleepaceReportSelectColumns = `
	sr.report_id::text,
	host(network(set_masklen(d.device_addr, 48))) || '/48' AS tenant_id,
	host(sr.device_addr) AS device_addr,
	COALESCE(dfm.device_code, '') AS device_code,
	dfm.device_uid,
	sr.record_count,
	sr.start_time_ms AS start_time,
	sr.end_time_ms   AS end_time,
	sr.date,
	sr.stop_mode,
	sr.time_step,
	sr.timezone,
	COALESCE(sr.sleep_state, '') AS sleep_state,
	COALESCE(sr.report_data, '') AS report,
	EXTRACT(EPOCH FROM sr.created_at)::bigint AS created_at,
	EXTRACT(EPOCH FROM sr.updated_at)::bigint AS updated_at
`

const sleepaceReportJoinClause = `
	FROM sleepace_report sr
	JOIN devices d                  ON d.device_uid = sr.device_uid
	JOIN device_factory_meta dfm    ON dfm.device_uid = sr.device_uid
`

func scanSleepaceReport(scan func(...any) error, report *domain.SleepaceReport) error {
	return scan(
		&report.ReportID,
		&report.TenantID,
		&report.DeviceAddr,
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
}

// GetReport Phase 2 一刀切：deviceID 入参承载 device_uid (sleepace.userId == owl device_uid logMAC)。
// tenantID 校验：device_addr 必须落在 tenantID prefix 内（防越权）。
func (r *PostgresSleepaceReportsRepository) GetReport(ctx context.Context, tenantID, deviceID string, date int) (*domain.SleepaceReport, error) {
	if tenantID == "" || deviceID == "" || date == 0 {
		return nil, fmt.Errorf("tenant_id, device_uid and date are required")
	}

	query := `SELECT ` + sleepaceReportSelectColumns + sleepaceReportJoinClause + `
		WHERE d.device_addr <<= $1::INET
		  AND sr.device_uid = $2
		  AND sr.date       = $3
	`

	var report domain.SleepaceReport
	err := scanSleepaceReport(r.db.QueryRowContext(ctx, query, tenantID, deviceID, date).Scan, &report)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sleepace report: %w", err)
	}

	return &report, nil
}

// ListReports 查询报告列表（支持分页）。deviceID 入参承载 device_uid。
func (r *PostgresSleepaceReportsRepository) ListReports(ctx context.Context, tenantID, deviceID string, startDate, endDate int, page, size int) ([]*domain.SleepaceReport, int, error) {
	if tenantID == "" || deviceID == "" {
		return nil, 0, fmt.Errorf("tenant_id and device_uid are required")
	}

	countQuery := `
		SELECT COUNT(*)
		FROM sleepace_report sr
		JOIN devices d ON d.device_uid = sr.device_uid
		WHERE d.device_addr <<= $1::INET
		  AND sr.device_uid = $2
		  AND sr.date >= $3
		  AND sr.date <= $4
	`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, tenantID, deviceID, startDate, endDate).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sleepace reports: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	query := `SELECT ` + sleepaceReportSelectColumns + sleepaceReportJoinClause + `
		WHERE d.device_addr <<= $1::INET
		  AND sr.device_uid = $2
		  AND sr.date >= $3
		  AND sr.date <= $4
		ORDER BY sr.date DESC
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
		if err := scanSleepaceReport(rows.Scan, &report); err != nil {
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

	// Phase 2: deviceID 入参承载 device_uid
	query := `
		SELECT sr.date
		FROM sleepace_report sr
		JOIN devices d ON d.device_uid = sr.device_uid
		WHERE d.device_addr <<= $1::INET
		  AND sr.device_uid = $2
		ORDER BY sr.date DESC
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
	query := `SELECT ` + sleepaceReportSelectColumns + sleepaceReportJoinClause + `
		WHERE d.device_addr <<= $1::INET
		  AND sr.device_uid = $2
		  AND sr.date >= $3
		  AND sr.date <= $4
		ORDER BY sr.date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, deviceID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list sleepace reports in range: %w", err)
	}
	defer rows.Close()
	out := make([]*domain.SleepaceReport, 0)
	for rows.Next() {
		var report domain.SleepaceReport
		if err := scanSleepaceReport(rows.Scan, &report); err != nil {
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
		SELECT sr.date
		FROM sleepace_report sr
		JOIN devices d ON d.device_uid = sr.device_uid
		WHERE d.device_addr <<= $1::INET
		  AND sr.device_uid = $2
		  AND sr.date >= $3
		  AND sr.date <= $4
		ORDER BY sr.date ASC
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

// GetDeviceUIDByDeviceUID Phase 2 一刀切：identity = device_uid；本函数保留 idempotent 校验入参 device_uid 存在性。
func (r *PostgresSleepaceReportsRepository) GetDeviceIDByDeviceUID(ctx context.Context, tenantID, deviceUID string) (string, error) {
	if tenantID == "" || deviceUID == "" {
		return "", fmt.Errorf("tenant_id and device_uid are required")
	}

	query := `
		SELECT dfm.device_uid
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_uid = dfm.device_uid
		WHERE d.device_addr <<= $1::INET
		  AND dfm.device_uid = $2
		LIMIT 1
	`

	var uid string
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceUID).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("device not found: device_uid=%s", deviceUID)
		}
		return "", fmt.Errorf("failed to verify device_uid: %w", err)
	}

	return uid, nil
}

// SaveReport Phase 2 一刀切：UPSERT 用 (device_uid, date) UNIQUE 约束。
// device_addr 从 devices 派生写入。
// TODO Phase 2.1: sleepace SDK userId format verify — report.DeviceUID 现在 == sleepace.userId
func (r *PostgresSleepaceReportsRepository) SaveReport(ctx context.Context, tenantID string, report *domain.SleepaceReport) error {
	if tenantID == "" || report == nil {
		return fmt.Errorf("tenant_id and report are required")
	}

	deviceUID := report.DeviceUID
	if deviceUID == "" {
		return fmt.Errorf("device_uid is required")
	}
	if _, err := r.GetDeviceIDByDeviceUID(ctx, tenantID, deviceUID); err != nil {
		return err
	}
	report.DeviceUID = deviceUID

	now := time.Now()

	upsert := `
		INSERT INTO sleepace_report (
			device_addr,
			device_uid,
			record_count,
			start_time_ms,
			end_time_ms,
			date,
			stop_mode,
			time_step,
			timezone,
			sleep_state,
			report_data,
			created_at,
			updated_at
		)
		SELECT
			d.device_addr,
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12
		FROM devices d
		WHERE d.device_uid = $1
		ON CONFLICT (device_uid, date) DO UPDATE SET
			record_count  = EXCLUDED.record_count,
			start_time_ms = EXCLUDED.start_time_ms,
			end_time_ms   = EXCLUDED.end_time_ms,
			stop_mode     = EXCLUDED.stop_mode,
			time_step     = EXCLUDED.time_step,
			timezone      = EXCLUDED.timezone,
			sleep_state   = EXCLUDED.sleep_state,
			report_data   = EXCLUDED.report_data,
			updated_at    = EXCLUDED.updated_at
		RETURNING report_id::text, EXTRACT(EPOCH FROM created_at)::bigint, EXTRACT(EPOCH FROM updated_at)::bigint
	`
	err := r.db.QueryRowContext(ctx, upsert,
		deviceUID,
		report.RecordCount,
		report.StartTime,
		report.EndTime,
		report.Date,
		report.StopMode,
		report.TimeStep,
		report.Timezone,
		report.SleepState,
		report.Report,
		now,
		now,
	).Scan(&report.ReportID, &report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert sleepace report: %w", err)
	}

	return nil
}
