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
// deviceID = devices.device_id（UUID）；拉厂家接口时该值作为 data.userId。租户/卡片/hardware uid 均由此 device_id 查库。
func (s *ReportService) DownloadAndSave(ctx context.Context, deviceID string, startTime, endTime int64) error {
	tenantID, resolvedDeviceID, deviceUID, deviceCode, cardID, err := s.loadReportWriteContext(ctx, deviceID)
	if err != nil {
		return err
	}

	s.logger.Debug("report write context",
		zap.String("device_id", resolvedDeviceID),
		zap.String("device_uid", deviceUID),
		zap.String("device_code", deviceCode),
		zap.String("card_id", cardID),
	)

	// 厂家 data.userId ← devices.device_id（UUID）
	reports, err := s.api.Get24HourDailyWithMaxReport(DailyMaxReportQuery{
		UserID:    resolvedDeviceID,
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
		calculatedEndTime := parsed.Summary.StartTime + int64(parsed.Summary.TimeStep)*int64(parsed.Summary.RecordCount)
		sleepState := string(parsed.Analysis.SleepStateStr)
		reportJSON := "[" + string(reports[i]) + "]"

		metadata := s.buildMetadata(ctx, resolvedDeviceID, tenantID)

		if err := s.upsert(ctx, tenantID, resolvedDeviceID, deviceUID, cardID,
			parsed.Summary.RecordCount, parsed.Summary.StartTime, calculatedEndTime,
			date, parsed.Summary.StopMode, parsed.Summary.TimeStep, parsed.Summary.Timezone,
			sleepState, reportJSON, metadata); err != nil {
			s.logger.Error("upsert report", zap.String("device_id", resolvedDeviceID), zap.Int("date", date), zap.Error(err))
		}
	}
	return nil
}

// loadReportWriteContext 按 devices.device_id（UUID）解析 Sleepad 写库所需上下文。
// 返回：tenant_id、d.device_id、devices.device_uid、device_store.device_code（厂家平台 deviceId，可空）、card_id。
func (s *ReportService) loadReportWriteContext(ctx context.Context, deviceID string) (
	tenantID, resolvedDeviceID, deviceUID, deviceCode, cardID string, err error,
) {
	if s.db == nil {
		return "", "", "", "", "", fmt.Errorf("database not available")
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT d.tenant_id::text,
			d.device_id::text,
			d.device_uid,
			COALESCE(NULLIF(TRIM(ds.device_code), ''), ''),
			COALESCE((
				SELECT c.card_id::text FROM cards c
				WHERE c.tenant_id = d.tenant_id
				  AND c.devices @> jsonb_build_array(jsonb_build_object('device_id', d.device_id::text))
				LIMIT 1
			), '')
		FROM devices d
		LEFT JOIN device_store ds ON ds.device_id = d.device_id
		WHERE d.device_id = $1::uuid AND d.status <> 'disabled'
	`, deviceID)
	if scanErr := row.Scan(&tenantID, &resolvedDeviceID, &deviceUID, &deviceCode, &cardID); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return "", "", "", "", "", fmt.Errorf("device not found: %s", deviceID)
		}
		return "", "", "", "", "", fmt.Errorf("load report context: %w", scanErr)
	}
	if tenantID == "" || resolvedDeviceID == "" || deviceUID == "" {
		return "", "", "", "", "", fmt.Errorf("incomplete device row for %s", deviceID)
	}
	return tenantID, resolvedDeviceID, deviceUID, deviceCode, cardID, nil
}

func (s *ReportService) buildMetadata(ctx context.Context, deviceID, tenantID string) string {
	if deviceID == "" || s.db == nil {
		return "{}"
	}
	// Anchor on devices.device_id only; location from bound_room_id (room-bound devices).
	row := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(
			jsonb_build_object(
				'device_name', d.device_name,
				'bed_name', '',
				'room_name', COALESCE(r.room_name, ''),
				'unit_name', COALESCE(u.unit_name, ''),
				'branch_name', COALESCE(br.branch_name, ''),
				'building_name', COALESCE(bld.building_name, ''),
				'resident_nickname', ''
			)::text,
			'{}'
		)
		FROM devices d
		LEFT JOIN rooms r ON r.room_id = d.bound_room_id AND r.tenant_id = d.tenant_id
		LEFT JOIN units u ON u.unit_id = r.unit_id AND u.tenant_id = d.tenant_id
		LEFT JOIN branches br ON br.branch_id = u.branch_id AND br.tenant_id = d.tenant_id
		LEFT JOIN buildings bld ON bld.building_id = u.building_id AND bld.tenant_id = d.tenant_id
		WHERE d.device_id = $1::uuid AND d.tenant_id = $2::uuid
		LIMIT 1
	`, deviceID, tenantID)
	var meta string
	if err := row.Scan(&meta); err != nil {
		s.logger.Debug("snapshot metadata query failed", zap.Error(err))
		return "{}"
	}
	return meta
}

func (s *ReportService) upsert(ctx context.Context,
	tenantID, deviceID, deviceUID, cardID string,
	recordCount int, startTime, endTime int64, date, stopMode, timeStep, timezone int,
	sleepState, report, metadata string,
) error {
	query := `
		INSERT INTO sleepace_report (
			tenant_id, device_id, device_uid, card_id,
			record_count, start_time, end_time, date,
			stop_mode, time_step, timezone,
			sleep_state, report, metadata, updated_at
		) VALUES (
			NULLIF($1,'')::uuid, NULLIF($2,'')::uuid, $3, NULLIF($4,'')::uuid,
			$5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14::jsonb, NOW()
		)
		ON CONFLICT ON CONSTRAINT sleepace_report_unique
		DO UPDATE SET
			record_count = EXCLUDED.record_count,
			start_time   = EXCLUDED.start_time,
			end_time     = EXCLUDED.end_time,
			stop_mode    = EXCLUDED.stop_mode,
			time_step    = EXCLUDED.time_step,
			timezone     = EXCLUDED.timezone,
			sleep_state  = EXCLUDED.sleep_state,
			report       = EXCLUDED.report,
			metadata     = EXCLUDED.metadata,
			card_id      = EXCLUDED.card_id,
			updated_at   = NOW()
	`
	_, err := s.db.ExecContext(ctx, query,
		tenantID, deviceID, deviceUID, cardID,
		recordCount, startTime, endTime, date,
		stopMode, timeStep, timezone,
		sleepState, report, metadata,
	)
	return err
}

func timeToDate(t int64) int {
	tm := time.Unix(t, 0).UTC()
	return tm.Year()*10000 + int(tm.Month())*100 + tm.Day()
}
