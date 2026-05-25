package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ReportService handles downloading and persisting sleep reports.
type ReportService struct {
	db     *sql.DB
	api    *SleepaceAPI
	card   *CardMappingService
	logger *zap.Logger
}

func NewReportService(db *sql.DB, api *SleepaceAPI, card *CardMappingService, logger *zap.Logger) *ReportService {
	return &ReportService{db: db, api: api, card: card, logger: logger}
}

// DownloadAndSave 从 sleepace-service 拉报告并 upsert sleepace_report。
//
// 入参 deviceUIDIn = device_factory_meta.device_uid (logMAC)；拉厂家接口时该值作为 data.userId。
func (s *ReportService) DownloadAndSave(ctx context.Context, deviceUIDIn string, startTime, endTime int64) error {
	tenantID, deviceUID, deviceCode, deviceAddr, residentID, err := s.loadReportWriteContext(ctx, deviceUIDIn)
	if err != nil {
		return err
	}

	s.logger.Debug("report write context",
		zap.String("device_uid", deviceUID),
		zap.String("device_code", deviceCode),
		zap.String("device_addr", deviceAddr),
		zap.String("resident_id", residentID),
		zap.String("tenant", tenantID),
	)

	reports, err := s.api.Get24HourDailyWithMaxReport(DailyMaxReportQuery{
		UserID:    deviceUID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		return fmt.Errorf("fetch report: %w", err)
	}

	for i := len(reports) - 1; i >= 0; i-- {
		parsed := struct {
			Summary struct {
				RecordCount int   `json:"recordCount"`
				StartTime   int64 `json:"startTime"`
				StopMode    int   `json:"stopMode"`
				TimeStep    int   `json:"timeStep"`
				Timezone    int   `json:"timezone"`
			} `json:"summary"`
			Analysis struct {
				SleepStateStr json.RawMessage `json:"sleepStateStr"`
			} `json:"analysis"`
		}{}
		if err := json.Unmarshal(reports[i], &parsed); err != nil {
			s.logger.Error("parse report", zap.Error(err))
			continue
		}

		date := timeToDate(parsed.Summary.StartTime)
		startTimeMs := parsed.Summary.StartTime * 1000
		endTimeMs := startTimeMs + int64(parsed.Summary.TimeStep)*int64(parsed.Summary.RecordCount)*1000
		sleepState := string(parsed.Analysis.SleepStateStr)
		reportJSON := "[" + string(reports[i]) + "]"

		metadata := s.buildMetadata(ctx, deviceUID)

		if err := s.upsert(ctx, deviceAddr, deviceUID, residentID,
			parsed.Summary.RecordCount, startTimeMs, endTimeMs,
			date, parsed.Summary.StopMode, parsed.Summary.TimeStep, parsed.Summary.Timezone,
			sleepState, reportJSON, metadata); err != nil {
			s.logger.Error("upsert report", zap.String("device_uid", deviceUID), zap.Int("date", date), zap.Error(err))
		}
	}
	return nil
}

// loadReportWriteContext 按 device_uid 解析 Sleepad 写库所需上下文。
//
// TODO Phase 2.1: sleepace SDK userId format verify — caller 现在传 device_uid (logMAC)，
// 不再传 dfm.device_id UUID；sleepace 厂家 SDK 接受性需 verify。
//
// 返回:
//   tenantID         tenant /48 host repr (e.g. "fd00:0:3::")
//   deviceUID        device_factory_meta.device_uid (logMAC，identity 不变量)
//   deviceCode       device_factory_meta.device_code (厂家平台 deviceId, 可空)
//   deviceAddr       devices.device_addr host repr (e.g. "fd00:0:3:411:3:201:c827:e11b") — v2 sleepace_report.device_addr 主键
//   residentID       LPM resident_unit 反查 (host repr 含 /128) 或空 (无入住)
func (s *ReportService) loadReportWriteContext(ctx context.Context, deviceUIDIn string) (
	tenantID, deviceUID, deviceCode, deviceAddr, residentID string, err error,
) {
	if s.db == nil {
		return "", "", "", "", "", fmt.Errorf("database not available")
	}
	var residentNull sql.NullString
	row := s.db.QueryRowContext(ctx, `
		SELECT host(network(set_masklen(d.device_addr, 48))) AS tenant_host,
		       dfm.device_uid,
		       COALESCE(NULLIF(TRIM(dfm.device_code), ''), ''),
		       host(d.device_addr) AS device_addr_host,
		       (SELECT host(ru.resident_id)
		          FROM resident_unit ru
		         WHERE ru.valid_to IS NULL
		           AND d.device_addr <<= ru.spatial_prefix
		         ORDER BY masklen(ru.spatial_prefix) DESC
		         LIMIT 1) AS resident_host
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE dfm.device_uid = $1
	`, deviceUIDIn)
	if scanErr := row.Scan(&tenantID, &deviceUID, &deviceCode, &deviceAddr, &residentNull); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return "", "", "", "", "", fmt.Errorf("device not found: %s", deviceUIDIn)
		}
		return "", "", "", "", "", fmt.Errorf("load report context: %w", scanErr)
	}
	if tenantID == "" || deviceUID == "" || deviceAddr == "" {
		return "", "", "", "", "", fmt.Errorf("incomplete device row for %s", deviceUIDIn)
	}
	if residentNull.Valid {
		residentID = residentNull.String
	}
	return tenantID, deviceUID, deviceCode, deviceAddr, residentID, nil
}

// buildMetadata 按 device_uid 从 device_addr 派生 site/branch/unit/room/bed name snapshot。
func (s *ReportService) buildMetadata(ctx context.Context, deviceUID string) string {
	if deviceUID == "" || s.db == nil {
		return "{}"
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT jsonb_build_object(
			'device_name', COALESCE(dfm.device_code, dfm.device_uid),
			'bed_name',    COALESCE((SELECT bed_name  FROM beds     b  WHERE d.device_addr <<= b.bed_id    LIMIT 1), ''),
			'room_name',   COALESCE((SELECT room_name FROM rooms    r  WHERE d.device_addr <<= r.room_id   LIMIT 1), ''),
			'unit_name',   COALESCE((SELECT unit_name FROM units    u  WHERE d.device_addr <<= u.unit_id   LIMIT 1), ''),
			'site_name',   COALESCE((SELECT site_name FROM sites    s  WHERE d.device_addr <<= s.site_id   LIMIT 1), ''),
			'branch_name', COALESCE((SELECT branch_name FROM branches br WHERE d.device_addr <<= br.branch_id LIMIT 1), ''),
			'resident_nickname', COALESCE((
				SELECT r.nickname
				FROM resident_unit ru
				JOIN residents r ON r.resident_id = ru.resident_id
				WHERE ru.valid_to IS NULL
				  AND d.device_addr <<= ru.spatial_prefix
				ORDER BY masklen(ru.spatial_prefix) DESC
				LIMIT 1
			), '')
		)::text
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE dfm.device_uid = $1
		LIMIT 1
	`, deviceUID)
	var meta string
	if err := row.Scan(&meta); err != nil {
		s.logger.Debug("snapshot metadata query failed", zap.Error(err))
		return "{}"
	}
	return meta
}

// upsert : sleepace_report (device_addr INET, device_uid VARCHAR(50), resident_id INET, ...)
func (s *ReportService) upsert(ctx context.Context,
	deviceAddr, deviceUID, residentID string,
	recordCount int, startTimeMs, endTimeMs int64, date, stopMode, timeStep, timezone int,
	sleepState, reportData, metadata string,
) error {
	query := `
		INSERT INTO sleepace_report (
			device_addr, device_uid, resident_id,
			record_count, start_time_ms, end_time_ms, date,
			stop_mode, time_step, timezone,
			sleep_state, report_data, metadata, updated_at
		) VALUES (
			$1::INET, $2, NULLIF($3,'')::INET,
			$4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13::jsonb, NOW()
		)
		ON CONFLICT ON CONSTRAINT sleepace_report_unique
		DO UPDATE SET
			device_addr   = EXCLUDED.device_addr,
			resident_id   = EXCLUDED.resident_id,
			record_count  = EXCLUDED.record_count,
			start_time_ms = EXCLUDED.start_time_ms,
			end_time_ms   = EXCLUDED.end_time_ms,
			stop_mode     = EXCLUDED.stop_mode,
			time_step     = EXCLUDED.time_step,
			timezone      = EXCLUDED.timezone,
			sleep_state   = EXCLUDED.sleep_state,
			report_data   = EXCLUDED.report_data,
			metadata      = EXCLUDED.metadata,
			updated_at    = NOW()
	`
	_, err := s.db.ExecContext(ctx, query,
		deviceAddr, deviceUID, residentID,
		recordCount, startTimeMs, endTimeMs, date,
		stopMode, timeStep, timezone,
		sleepState, reportData, metadata,
	)
	return err
}

func timeToDate(t int64) int {
	tm := time.Unix(t, 0).UTC()
	return tm.Year()*10000 + int(tm.Month())*100 + tm.Day()
}
