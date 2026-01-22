package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"wisefido-data/internal/models"

	"go.uber.org/zap"
)

// enrichCardTimeFields 为卡片添加时间字段（bed_status_timestamp 和 status_duration）
// 仅在 ActiveBed 卡片上处理
func (s *cardService) enrichCardTimeFields(ctx context.Context, card *models.VitalFocusCard) error {
	if card.CardType != "ActiveBed" {
		return nil // Location 卡片不需要时间字段
	}

	// 从 Redis 读取 realtime data 获取 BedStatusTimestamp
	realtimeKey := fmt.Sprintf("vital-focus:card:%s:realtime", card.CardID)
	realtimeDataRaw, err := s.kv.Get(ctx, realtimeKey)
	if err != nil {
		// realtime data 不存在，时间字段保持为空
		return nil
	}

	var realtimeData struct {
		BedStatusTimestamp *int64 `json:"bed_status_timestamp,omitempty"`
	}
	if err := json.Unmarshal([]byte(realtimeDataRaw), &realtimeData); err != nil {
		// 解析失败，时间字段保持为空
		return nil
	}

	if realtimeData.BedStatusTimestamp == nil || *realtimeData.BedStatusTimestamp == 0 {
		// 时间戳无效，时间字段保持为空
		return nil
	}

	// 获取 unit 的时区
	unitID, err := s.getUnitIDFromCard(ctx, card)
	if err != nil {
		// unit_id 获取失败，使用 UTC
		unitID = ""
	}

	timezone, err := s.getUnitTimezone(ctx, unitID)
	if err != nil {
		// 时区获取失败，使用 UTC
		timezone = "UTC"
	}

	// 格式化 bed_status_timestamp
	card.BedStatusTimestamp = s.formatBedStatusTimestamp(*realtimeData.BedStatusTimestamp, timezone)

	// 计算 status_duration（已检查设备连接状态）
	card.StatusDuration = s.calculateStatusDuration(ctx, card, *realtimeData.BedStatusTimestamp)

	return nil
}

// getUnitIDFromCard 从卡片数据获取 unit_id
// 对于 ActiveBed 卡片，通过 bed_id 查询 unit_id
// 对于 Location 卡片，LocationID 就是 unit_id
func (s *cardService) getUnitIDFromCard(ctx context.Context, card *models.VitalFocusCard) (string, error) {
	if card.CardType == "Location" {
		// Location 卡片：LocationID 就是 unit_id
		if card.LocationID != "" {
			return card.LocationID, nil
		}
		return "", fmt.Errorf("location_id is empty")
	}

	// ActiveBed 卡片：通过 bed_id 查询 unit_id
	if card.BedID == "" {
		return "", fmt.Errorf("bed_id is empty")
	}

	query := `
		SELECT u.unit_id::text
		FROM beds b
		JOIN rooms r ON b.room_id = r.room_id
		JOIN units u ON r.unit_id = u.unit_id
		WHERE b.bed_id = $1 AND b.tenant_id = $2
	`
	var unitID string
	err := s.db.QueryRowContext(ctx, query, card.BedID, card.TenantID).Scan(&unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("unit not found for bed_id: %s", card.BedID)
		}
		return "", fmt.Errorf("failed to query unit_id: %w", err)
	}

	return unitID, nil
}

// getUnitTimezone 获取 unit 的时区
func (s *cardService) getUnitTimezone(ctx context.Context, unitID string) (string, error) {
	if unitID == "" {
		return "UTC", nil
	}

	query := `SELECT COALESCE(timezone, 'UTC') FROM units WHERE unit_id = $1`
	var timezone string
	err := s.db.QueryRowContext(ctx, query, unitID).Scan(&timezone)
	if err != nil {
		if err == sql.ErrNoRows {
			return "UTC", nil
		}
		return "", err
	}
	if timezone == "" {
		return "UTC", nil
	}
	return timezone, nil
}

// formatBedStatusTimestamp 格式化床状态时间戳为 "HH:mm:ss"
func (s *cardService) formatBedStatusTimestamp(timestamp int64, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
		s.logger.Warn("Failed to load timezone, using UTC",
			zap.String("timezone", timezone),
			zap.Error(err),
		)
	}

	t := time.Unix(timestamp, 0).In(loc)
	return t.Format("15:04:05") // HH:mm:ss
}

// calculateStatusDuration 计算状态持续时间
// 注意：仅当设备 offline 或 error 时才会触发检查逻辑
func (s *cardService) calculateStatusDuration(
	ctx context.Context,
	card *models.VitalFocusCard,
	bedStatusTimestamp int64,
) string {
	// 校验1：timestamp 不能为 0 或异常旧的时间
	if bedStatusTimestamp <= 0 {
		s.logger.Warn("Invalid bed_status_timestamp: zero or negative",
			zap.String("card_id", card.CardID),
			zap.Int64("bed_status_timestamp", bedStatusTimestamp),
		)
		return ""
	}

	// 获取设备连接状态
	sOnline := card.SConnection != nil && *card.SConnection == 1
	rOnline := card.RConnection != nil && *card.RConnection == 1

	// 1. 正常情况：Sleepad 和 Radar 都 online，直接正常计算
	if sOnline && rOnline {
		return s.calculateDuration(bedStatusTimestamp)
	}

	// 2. 异常情况：有设备 offline/error，进入检查逻辑

	// 2.1 Sleepad offline/error
	if !sOnline {
		// Sleepad 离线，检查 Bed 上是否绑定了 Radar
		hasRadar := s.hasRadarInBed(ctx, card)
		if !hasRadar {
			// 没有 Radar，重置为 ""
			return ""
		}

		// 检查 Radar 是否在线
		if !rOnline {
			// Radar 也离线，重置为 ""
			return ""
		}

		// Radar 在线，比较 Radar 和 Sleepad 的最后一次 bed_event 状态
		sleepadBedStatus, radarBedStatus, err := s.getLastBedStatusFromDevices(ctx, card)
		if err != nil {
			// 获取失败，重置为 ""
			s.logger.Warn("Failed to get last bed status from devices",
				zap.String("card_id", card.CardID),
				zap.Error(err),
			)
			return ""
		}

		// 比较状态是否一致
		if sleepadBedStatus == radarBedStatus {
			// 状态一致，继续计算 status_duration（使用 Sleepad 的时间值）
			return s.calculateDuration(bedStatusTimestamp)
		} else {
			// 状态不一致，重置为 ""（可信度不高）
			return ""
		}
	}

	// 2.2 Sleepad online，但 Radar offline/error
	// 继续使用 Sleepad 的数据计算 status_duration（不重置）
	return s.calculateDuration(bedStatusTimestamp)
}

// calculateDuration 计算持续时间（从 bed_status_timestamp 到当前时间）
func (s *cardService) calculateDuration(bedStatusTimestamp int64) string {
	now := time.Now().Unix()
	duration := now - bedStatusTimestamp

	// 校验1：duration 不能为负数
	if duration < 0 {
		s.logger.Warn("Invalid bed_status_timestamp: future timestamp",
			zap.Int64("bed_status_timestamp", bedStatusTimestamp),
			zap.Int64("now", now),
		)
		return ""
	}

	// 校验2：duration 不能超过 24 小时（合理范围）
	// 超过24小时可能是设备离线导致的数据过期
	const maxDurationHours = 24
	if duration > maxDurationHours*3600 {
		s.logger.Warn("Invalid bed_status_timestamp: duration too long (possibly device offline)",
			zap.Int64("bed_status_timestamp", bedStatusTimestamp),
			zap.Int64("duration_seconds", duration),
		)
		return "" // 数据过期（可能是设备离线），不显示 duration
	}

	// 格式化为 "HH:MM"
	hours := duration / 3600
	minutes := (duration % 3600) / 60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

// hasRadarInBed 检查 Bed 上是否绑定了 Radar
// 通过 devices 表查询 bed_id 绑定的 Radar 设备
func (s *cardService) hasRadarInBed(ctx context.Context, card *models.VitalFocusCard) bool {
	if card.BedID == "" {
		return false
	}

	query := `
		SELECT COUNT(*) > 0
		FROM devices d
		JOIN device_store ds ON d.device_uid = ds.device_uid
		WHERE d.bound_bed_id = $1
		  AND d.tenant_id = $2
		  AND ds.device_type = 'Radar'
		  AND d.monitoring_enabled = TRUE
	`
	var hasRadar bool
	err := s.db.QueryRowContext(ctx, query, card.BedID, card.TenantID).Scan(&hasRadar)
	if err != nil {
		s.logger.Warn("Failed to check if bed has Radar device",
			zap.String("card_id", card.CardID),
			zap.String("bed_id", card.BedID),
			zap.Error(err),
		)
		return false
	}
	return hasRadar
}

// getLastBedStatusFromDevices 获取 Sleepad 和 Radar 的最后一次 bed_event 状态
// 从 iot_timeseries 表查询最后一次 bed_status_snomed_code
func (s *cardService) getLastBedStatusFromDevices(
	ctx context.Context,
	card *models.VitalFocusCard,
) (sleepadBedStatus, radarBedStatus string, err error) {
	if card.BedID == "" {
		return "", "", fmt.Errorf("bed_id is empty")
	}

	// 查询 Bed 上绑定的 Sleepad 和 Radar 设备 ID
	// 注意：device_type 可能有多种命名（Sleepace, Sleepad, SleepPad），使用 ILIKE 支持所有变体
	deviceQuery := `
		SELECT d.device_id, ds.device_type
		FROM devices d
		JOIN device_store ds ON d.device_uid = ds.device_uid
		WHERE d.bound_bed_id = $1
		  AND d.tenant_id = $2
		  AND (ds.device_type ILIKE '%sleep%' OR ds.device_type = 'Radar')
		  AND d.monitoring_enabled = TRUE
	`
	rows, err := s.db.QueryContext(ctx, deviceQuery, card.BedID, card.TenantID)
	if err != nil {
		return "", "", fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var sleepadDeviceID, radarDeviceID string
	for rows.Next() {
		var deviceID, deviceType string
		if err := rows.Scan(&deviceID, &deviceType); err != nil {
			continue
		}
		// 支持多种命名：Sleepace, Sleepad, SleepPad, sleepad 等（统一转换为 Sleepad）
		if deviceType == "Radar" {
			radarDeviceID = deviceID
		} else {
			// 所有包含 'sleep' 的设备类型都视为 Sleepad
			sleepadDeviceID = deviceID
		}
	}
	if err := rows.Err(); err != nil {
		return "", "", fmt.Errorf("failed to scan devices: %w", err)
	}

	// 查询 Sleepad 的最后一次 bed_status
	var sleepadStatus sql.NullString
	if sleepadDeviceID != "" {
		sleepadQuery := `
			SELECT bed_status_snomed_code
			FROM iot_timeseries
			WHERE device_id = $1
			  AND tenant_id = $2
			  AND bed_status_snomed_code IS NOT NULL
			ORDER BY timestamp DESC
			LIMIT 1
		`
		err := s.db.QueryRowContext(ctx, sleepadQuery, sleepadDeviceID, card.TenantID).Scan(&sleepadStatus)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Warn("Failed to query Sleepad last bed status",
				zap.String("device_id", sleepadDeviceID),
				zap.Error(err),
			)
			// 继续执行，sleepadStatus 保持为 NULL
		}
	}

	// 查询 Radar 的最后一次 bed_status
	var radarStatus sql.NullString
	if radarDeviceID != "" {
		radarQuery := `
			SELECT bed_status_snomed_code
			FROM iot_timeseries
			WHERE device_id = $1
			  AND tenant_id = $2
			  AND bed_status_snomed_code IS NOT NULL
			ORDER BY timestamp DESC
			LIMIT 1
		`
		err := s.db.QueryRowContext(ctx, radarQuery, radarDeviceID, card.TenantID).Scan(&radarStatus)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Warn("Failed to query Radar last bed status",
				zap.String("device_id", radarDeviceID),
				zap.Error(err),
			)
			// 继续执行，radarStatus 保持为 NULL
		}
	}

	// 返回状态（如果查询失败或没有数据，返回空字符串）
	sleepadResult := ""
	if sleepadStatus.Valid {
		sleepadResult = sleepadStatus.String
	}

	radarResult := ""
	if radarStatus.Valid {
		radarResult = radarStatus.String
	}

	return sleepadResult, radarResult, nil
}
