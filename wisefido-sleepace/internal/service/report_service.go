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

// DownloadAndSave 从 sleepace-service 拉报告并 upsert sleepace_report（v2）。
//
// deviceID = device_factory_meta.device_id（UUID）；拉厂家接口时该值作为 data.userId。
// 租户 / 设备 IPv6 / 入住住户均由此 device_id 查 v2 表派生。
func (s *ReportService) DownloadAndSave(ctx context.Context, deviceID string, startTime, endTime int64) error {
	tenantID, resolvedDeviceID, deviceUID, deviceCode, deviceAddr, residentID, err := s.loadReportWriteContext(ctx, deviceID)
	if err != nil {
		return err
	}

	s.logger.Debug("report write context",
		zap.String("device_id", resolvedDeviceID),
		zap.String("device_uid", deviceUID),
		zap.String("device_code", deviceCode),
		zap.String("device_addr", deviceAddr),
		zap.String("resident_id", residentID),
		zap.String("tenant", tenantID),
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
		// sleepace 用秒级 startTime；v2 schema 用 ms — 乘 1000
		startTimeMs := parsed.Summary.StartTime * 1000
		endTimeMs := startTimeMs + int64(parsed.Summary.TimeStep)*int64(parsed.Summary.RecordCount)*1000
		sleepState := string(parsed.Analysis.SleepStateStr)
		reportJSON := "[" + string(reports[i]) + "]"

		metadata := s.buildMetadata(ctx, resolvedDeviceID)

		if err := s.upsert(ctx, deviceAddr, resolvedDeviceID, residentID,
			parsed.Summary.RecordCount, startTimeMs, endTimeMs,
			date, parsed.Summary.StopMode, parsed.Summary.TimeStep, parsed.Summary.Timezone,
			sleepState, reportJSON, metadata); err != nil {
			s.logger.Error("upsert report", zap.String("device_id", resolvedDeviceID), zap.Int("date", date), zap.Error(err))
		}
	}
	return nil
}

// loadReportWriteContext 按 devices.device_id（UUID）解析 Sleepad 写库所需上下文（v2）。
//
// 返回:
//   tenantID         tenant /48 host repr (e.g. "fd00:0:3::"), 兼容签名；v2 sleepace_report 不存这列
//   resolvedDeviceID device_factory_meta.device_id (UUID 字符串)
//   deviceUID        device_factory_meta.device_uid (logMAC)
//   deviceCode       device_factory_meta.device_code (厂家平台 deviceId, 可空)
//   deviceAddr       devices.device_ipv6 host repr (e.g. "fd00:0:3:411:3:201:c827:e11b") — v2 sleepace_report.device_addr 主键
//   residentID       LPM resident_unit 反查 (host repr 含 /128) 或空 (无入住)
func (s *ReportService) loadReportWriteContext(ctx context.Context, deviceID string) (
	tenantID, resolvedDeviceID, deviceUID, deviceCode, deviceAddr, residentID string, err error,
) {
	if s.db == nil {
		return "", "", "", "", "", "", fmt.Errorf("database not available")
	}
	var residentNull sql.NullString
	row := s.db.QueryRowContext(ctx, `
		SELECT host(network(set_masklen(d.device_ipv6, 48))) AS tenant_host,
		       dfm.device_id::text,
		       dfm.device_uid,
		       COALESCE(NULLIF(TRIM(dfm.device_code), ''), ''),
		       host(d.device_ipv6) AS device_addr_host,
		       (SELECT host(ru.resident_id)
		          FROM resident_unit ru
		         WHERE ru.valid_to IS NULL
		           AND d.device_ipv6 <<= ru.spatial_prefix
		         ORDER BY masklen(ru.spatial_prefix) DESC
		         LIMIT 1) AS resident_host
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE dfm.device_id = $1::uuid
	`, deviceID)
	if scanErr := row.Scan(&tenantID, &resolvedDeviceID, &deviceUID, &deviceCode, &deviceAddr, &residentNull); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return "", "", "", "", "", "", fmt.Errorf("device not found: %s", deviceID)
		}
		return "", "", "", "", "", "", fmt.Errorf("load report context: %w", scanErr)
	}
	if tenantID == "" || resolvedDeviceID == "" || deviceUID == "" || deviceAddr == "" {
		return "", "", "", "", "", "", fmt.Errorf("incomplete device row for %s", deviceID)
	}
	if residentNull.Valid {
		residentID = residentNull.String
	}
	return tenantID, resolvedDeviceID, deviceUID, deviceCode, deviceAddr, residentID, nil
}

// buildMetadata v2：从 device_ipv6 派生 site/branch/unit/room/bed name snapshot。
//
// v2 派生（prefix-match 替代 v1 多列 JOIN）：
//   - bed:    bedDB.bed_name where device_ipv6 <<= bed_id (/96)
//   - room:   rooms.room_name where device_ipv6 <<= room_id (/88)
//   - unit:   units.unit_name where device_ipv6 <<= unit_id (/80)
//   - site:   sites.site_name where device_ipv6 <<= site_id (/64) — 取代 v1 buildings 表
//   - branch: branches.branch_name where device_ipv6 <<= branch_id (/56)
//
// 不再用 d.device_name (v2 已删)；取 dfm.device_code fallback dfm.device_uid。
// 不再用 buildings 表（已并入 sites）。
func (s *ReportService) buildMetadata(ctx context.Context, deviceID string) string {
	if deviceID == "" || s.db == nil {
		return "{}"
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT jsonb_build_object(
			'device_name', COALESCE(dfm.device_code, dfm.device_uid),
			'bed_name',    COALESCE((SELECT bed_name  FROM beds     b  WHERE d.device_ipv6 <<= b.bed_id    LIMIT 1), ''),
			'room_name',   COALESCE((SELECT room_name FROM rooms    r  WHERE d.device_ipv6 <<= r.room_id   LIMIT 1), ''),
			'unit_name',   COALESCE((SELECT unit_name FROM units    u  WHERE d.device_ipv6 <<= u.unit_id   LIMIT 1), ''),
			'site_name',   COALESCE((SELECT site_name FROM sites    s  WHERE d.device_ipv6 <<= s.site_id   LIMIT 1), ''),
			'branch_name', COALESCE((SELECT branch_name FROM branches br WHERE d.device_ipv6 <<= br.branch_id LIMIT 1), ''),
			'resident_nickname', COALESCE((
				SELECT r.nickname
				FROM resident_unit ru
				JOIN residents r ON r.resident_id = ru.resident_id
				WHERE ru.valid_to IS NULL
				  AND d.device_ipv6 <<= ru.spatial_prefix
				ORDER BY masklen(ru.spatial_prefix) DESC
				LIMIT 1
			), '')
		)::text
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE dfm.device_id = $1::uuid
		LIMIT 1
	`, deviceID)
	var meta string
	if err := row.Scan(&meta); err != nil {
		s.logger.Debug("snapshot metadata query failed", zap.Error(err))
		return "{}"
	}
	return meta
}

// upsert v2: sleepace_report (device_addr INET, device_id UUID, resident_id INET, *_ms BIGINT, report_data, ...)
//
// 列变化 vs v1:
//   tenant_id          → 删（v2 不存；tenant 派生自 device_addr /48）
//   device_code/uid    → 删（v2 不存；查 dfm 派生）
//   card_id            → 删（v2 不存；查 cards LPM 派生）
//   start_time/end_time → start_time_ms/end_time_ms（v2 用 ms）
//   report             → report_data
//   + device_addr INET（新）
//   + resident_id INET (LPM 派生)
func (s *ReportService) upsert(ctx context.Context,
	deviceAddr, deviceID, residentID string,
	recordCount int, startTimeMs, endTimeMs int64, date, stopMode, timeStep, timezone int,
	sleepState, reportData, metadata string,
) error {
	query := `
		INSERT INTO sleepace_report (
			device_addr, device_id, resident_id,
			record_count, start_time_ms, end_time_ms, date,
			stop_mode, time_step, timezone,
			sleep_state, report_data, metadata, updated_at
		) VALUES (
			$1::INET, $2::uuid, NULLIF($3,'')::INET,
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
		deviceAddr, deviceID, residentID,
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
