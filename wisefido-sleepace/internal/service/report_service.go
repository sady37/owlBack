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

// DownloadAndSave fetches the report from sleepace-service and upserts into sleepace_report.
// mqttDeviceID is the MQTT deviceId (= device_store.device_code). userID is Sleepace API userId (= devices.device_id UUID).
func (s *ReportService) DownloadAndSave(ctx context.Context, mqttDeviceID, userID string, startTime, endTime int64) error {
	reports, err := s.api.GetDailyReport(userID, startTime, endTime)
	if err != nil {
		return fmt.Errorf("fetch report: %w", err)
	}

	cardID, tenantID, deviceID := "", "", ""
	if info, err := s.card.GetCardInfo(ctx, mqttDeviceID); err == nil {
		cardID = info.CardID
		tenantID = info.TenantID
		deviceID = info.DeviceID
	} else {
		s.logger.Warn("report skip: card lookup failed",
			zap.String("mqtt_device_id", mqttDeviceID),
			zap.String("api_user_id", userID),
			zap.Error(err),
		)
		return nil
	}
	if tenantID == "" || deviceID == "" {
		s.logger.Warn("report skip: empty tenant_id or device_id",
			zap.String("mqtt_device_id", mqttDeviceID),
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
		)
		return nil
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

		metadata := s.buildMetadata(ctx, deviceID, tenantID)

		if err := s.upsert(ctx, tenantID, deviceID, mqttDeviceID, cardID,
			parsed.Summary.RecordCount, parsed.Summary.StartTime, calculatedEndTime,
			date, parsed.Summary.StopMode, parsed.Summary.TimeStep, parsed.Summary.Timezone,
			sleepState, reportJSON, metadata); err != nil {
			s.logger.Error("upsert report", zap.String("mqtt_device_id", mqttDeviceID), zap.Int("date", date), zap.Error(err))
		}
	}
	return nil
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
	tenantID, deviceID, deviceCode, cardID string,
	recordCount int, startTime, endTime int64, date, stopMode, timeStep, timezone int,
	sleepState, report, metadata string,
) error {
	query := `
		INSERT INTO sleepace_report (
			tenant_id, device_id, device_code, card_id,
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
		tenantID, deviceID, deviceCode, cardID,
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
