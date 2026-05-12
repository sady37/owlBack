package card

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"

	"owl-common/alarm"
)

// ====================================================================
// alarm_db.go — v2 缩减版（cards-v2 phase B1c）
//
// 背景：v2 alarm_events schema 与 v1 不同：
//   - device_addr INET（非 device_id only）
//   - alarm_kind / severity 0-7（非 event_type / alarm_level VARCHAR）
//   - process_status enum: pending/acked/verified/false_positive/dismissed
//   - payload/evidence（非 trigger_data/metadata）
//   - tenant/branch/unit/room/bed_name 等 snapshot 列已 inline，不再走 metadata
//
// 与此同时，cards 表 v2 schema 已删除 unhandled_alarm_0..4 / pop_alarm_* 列，
// counter 与 pop 由 caller 实时聚合 alarm_events 得出（详 view 设计）。
//
// 本文件保留 API 签名供 caller 编译通过；返回值用空 CardAlarmState（zero counts / 空 pop）。
// 真正的 v2 alarm 改造（详 doc/cards_v2_migration_checklist.md § Phase G）单独立项。
// ====================================================================

// AlarmInsertParams INSERT alarm_events 所需参数（API 兼容；map 到 v2 列由各 caller 自行实现）
type AlarmInsertParams struct {
	TenantID    string
	DeviceID    string
	EventType   string          // 映射 v2 alarm_kind
	Category    string
	AlarmLevel  string          // 映射 v2 severity（caller normalize）
	TriggeredAt time.Time
	TriggerData json.RawMessage // 映射 v2 payload
	Metadata    json.RawMessage // 映射 v2 evidence
	RoomID      string
	UnitID      string
}

// AlarmInsertResult INSERT 后返回的结果
type AlarmInsertResult struct {
	EventID       string
	TriggeredAt   time.Time
	Deduped       bool
	SkippedNotify bool
}

// AlarmUpdateParams UPDATE alarm_events 所需参数
type AlarmUpdateParams struct {
	AlarmStatus     string          // 映射 v2 process_status
	Handler         string          // 映射 v2 processed_by (UUID) — caller 解析
	Operation       string          // 映射 v2 handle_type
	Notes           *string         // 映射 v2 handler_notes
	NotifiedUsers   json.RawMessage
	Metadata        json.RawMessage
	ResolveSnapshot json.RawMessage
}

// CardAlarmState v2 cards 表无 alarm 计数列；本 struct 仅保留以兼容签名，
// 各 counter/pop 字段由 caller 通过 view 实时聚合填充。本文件返回零值。
type CardAlarmState struct {
	UnhandledAlarm0  int
	UnhandledAlarm1  int
	UnhandledAlarm2  int
	UnhandledAlarm3  int
	UnhandledAlarm4  int
	PopAlarmLevel    string
	PopAlarmType     string
	PopAlarmEventId  string
	PopTriggeredAtMs int64
	HandTime         time.Time
}

// ToAlarmState 转换为 AlarmState
func (c *CardAlarmState) ToAlarmState() *AlarmState {
	if c == nil {
		return &AlarmState{UpdatedAt: time.Now().UnixMilli()}
	}
	s := &AlarmState{
		UpdatedAt:     time.Now().UnixMilli(),
		ActiveEmerg:   c.UnhandledAlarm0,
		ActiveAlert:   c.UnhandledAlarm1,
		ActiveCrit:    c.UnhandledAlarm2,
		ActiveErr:     c.UnhandledAlarm3,
		ActiveWarning: c.UnhandledAlarm4,
	}
	if c.PopAlarmLevel != "" {
		s.PopAlarm = c.PopAlarmLevel + "." + c.PopAlarmType
		s.EventID = c.PopAlarmEventId
		if c.PopTriggeredAtMs > 0 {
			s.TriggeredAt = c.PopTriggeredAtMs
		}
	}
	return s
}

// InsertAlarmAndUpdateCard v2 stub：
//   - alarm_events v2 schema 改动太大，本函数暂返回 EventID="" 让 caller 失败优雅
//   - cards 表不再有 counter/pop 列；返回空 CardAlarmState
//
// TODO Phase G — v2 alarm 改造：device_addr INET / alarm_kind / severity / payload。
func InsertAlarmAndUpdateCard(ctx context.Context, db *sql.DB, cardID string, params AlarmInsertParams) (*AlarmInsertResult, *CardAlarmState, error) {
	if def := alarm.LookupAlarm(params.EventType); def != nil && def.SkipUnhandledCount {
		// device-class 告警在 v2 走独立 device:status hash 通道，不入 alarm_events
		return &AlarmInsertResult{EventID: "", TriggeredAt: params.TriggeredAt, SkippedNotify: true}, &CardAlarmState{}, nil
	}
	return nil, nil, fmt.Errorf("alarm v2 not yet implemented; see Phase G in doc/cards_v2_migration_checklist.md")
}

// QueryCardAlarmState v2 stub：实时聚合 alarm_events 实现 pending — 返回零状态。
func QueryCardAlarmState(ctx context.Context, db *sql.DB, cardID string) (*CardAlarmState, error) {
	if cardID == "" {
		return nil, fmt.Errorf("cardID is required")
	}
	return &CardAlarmState{}, nil
}

// UpdateAlarmAndUpdateCard v2 stub：alarm_events 状态更新 + cards counter 同步 pending。
func UpdateAlarmAndUpdateCard(ctx context.Context, db *sql.DB, cardID, tenantID, eventID string, params AlarmUpdateParams) (*CardAlarmState, error) {
	return nil, fmt.Errorf("alarm v2 not yet implemented; see Phase G in doc/cards_v2_migration_checklist.md")
}

// RecalcCardAlarmState v2 stub。
func RecalcCardAlarmState(ctx context.Context, db *sql.DB, cardID, tenantID string) (*CardAlarmState, error) {
	if cardID == "" {
		return nil, fmt.Errorf("cardID is required")
	}
	return &CardAlarmState{}, nil
}

// LookupCardIDByDeviceID v2: 通过 device_id → device_ipv6 → cards LPM。
func LookupCardIDByDeviceID(ctx context.Context, db *sql.DB, deviceID string) (string, error) {
	if deviceID == "" {
		return "", fmt.Errorf("device_id is required")
	}
	var cardID sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT find_card_by_device_addr(d.device_ipv6)::text
		FROM devices d
		WHERE d.device_id = $1::uuid
		LIMIT 1
	`, deviceID).Scan(&cardID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no devices row for device %s", deviceID)
		}
		return "", err
	}
	if !cardID.Valid || cardID.String == "" {
		return "", fmt.Errorf("no card found for device %s", deviceID)
	}
	return cardID.String, nil
}

// DeviceSelfRecoveryAlarmTypes 设备恢复时需自动解除的报警类型。
// Elder care 策略：仅设备类可自动恢复；非设备类必须人工确认。
var DeviceSelfRecoveryAlarmTypes = []string{
	"Offline", "DeviceFailure", "AngleException", "SignalPoor", "SensorDetached",
}

// AutoResolveResult 设备自恢复返回结果
type AutoResolveResult struct {
	ResolvedCount int
	TopEventId    string
	TopAlarmLevel string
}

// AutoResolveDeviceAlarms v2 stub：device-class 告警在 v2 走独立 device:status 通道，不再写 alarm_events。
func AutoResolveDeviceAlarms(ctx context.Context, db *sql.DB, cardID, tenantID, deviceID string, alarmTypes []string) (*CardAlarmState, *AutoResolveResult, error) {
	return nil, &AutoResolveResult{}, nil
}

// ExpireAlarmsByDeviceIDs v2 stub：保留签名；实际过期由 caller 调 v2 process_status='dismissed' UPDATE 完成。
func ExpireAlarmsByDeviceIDs(ctx context.Context, db *sql.DB, tenantID string, deviceIDs []string, reason string) error {
	if len(deviceIDs) == 0 {
		return nil
	}
	// v2 alarm_events: device_id 列已存在；按 device_id 把 pending alarm 标记 dismissed。
	// 不依赖已删除的 metadata.deleted_at / alarm_status enum 老值。
	_, err := db.ExecContext(ctx, `
		UPDATE alarm_events
		SET process_status = 'dismissed',
		    processed_at = CURRENT_TIMESTAMP,
		    handle_type = $2
		WHERE device_id = ANY($1::uuid[])
		  AND process_status = 'pending'
	`, pq.Array(deviceIDs), reason)
	if err != nil {
		return fmt.Errorf("expire alarms by device IDs: %w", err)
	}
	return nil
}

// ActiveAlarmRow alarm_events 单条 active 行（API 兼容）
type ActiveAlarmRow struct {
	EventID     string
	EventType   string
	AlarmLevel  string
	DeviceID    string
	TriggeredAt int64
}

// AlarmWithCardInfo 含 card 归属的 alarm 行（API 兼容）
type AlarmWithCardInfo struct {
	CardID string
	ActiveAlarmRow
}

// GetActiveAlarmsByCardID v2: 走 cards.spatial_prefix LPM → device_addr 反查。
func GetActiveAlarmsByCardID(ctx context.Context, db *sql.DB, cardID string) ([]ActiveAlarmRow, error) {
	if cardID == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT ae.event_id, ae.alarm_kind, ae.severity::text, COALESCE(ae.device_id::text, ''),
		       EXTRACT(EPOCH FROM ae.triggered_at)::bigint
		FROM alarm_events ae
		JOIN cards c ON ae.device_addr <<= c.spatial_prefix
		WHERE c.card_id = $1::uuid
		  AND ae.process_status = 'pending'
		ORDER BY ae.triggered_at DESC
	`, cardID)
	if err != nil {
		return nil, fmt.Errorf("get active alarms for card %s: %w", cardID, err)
	}
	defer rows.Close()

	var result []ActiveAlarmRow
	for rows.Next() {
		var a ActiveAlarmRow
		if err := rows.Scan(&a.EventID, &a.EventType, &a.AlarmLevel, &a.DeviceID, &a.TriggeredAt); err != nil {
			return nil, fmt.Errorf("scan alarm row: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// ListAllActiveAlarms v2: 走 cards.spatial_prefix LPM。
func ListAllActiveAlarms(ctx context.Context, db *sql.DB, tenantID string) ([]AlarmWithCardInfo, error) {
	query := `
		SELECT c.card_id::text, ae.event_id, ae.alarm_kind, ae.severity::text, COALESCE(ae.device_id::text, ''),
		       EXTRACT(EPOCH FROM ae.triggered_at)::bigint
		FROM alarm_events ae
		JOIN cards c ON ae.device_addr <<= c.spatial_prefix
		WHERE ae.process_status = 'pending'
	`
	var args []interface{}
	if tenantID != "" {
		query += ` AND c.spatial_prefix <<= $1::inet`
		args = append(args, tenantID)
	}
	query += ` ORDER BY ae.triggered_at DESC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all active alarms: %w", err)
	}
	defer rows.Close()

	var result []AlarmWithCardInfo
	for rows.Next() {
		var a AlarmWithCardInfo
		if err := rows.Scan(&a.CardID, &a.EventID, &a.EventType, &a.AlarmLevel, &a.DeviceID, &a.TriggeredAt); err != nil {
			return nil, fmt.Errorf("scan alarm row: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}
