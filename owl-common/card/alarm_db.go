package card

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"owl-common/alarm"

	"github.com/lib/pq"
)

// AlarmInsertParams INSERT alarm_events 所需参数
type AlarmInsertParams struct {
	TenantID    string
	DeviceID    string
	EventType   string // "Fall", "AbnormalHeartRate" etc
	Category    string // "safety", "clinical", "behavioral", "device"
	AlarmLevel  string // "EMERG", "ALERT", "CRIT", "ERR", "WARNING"
	TriggeredAt time.Time
	TriggerData json.RawMessage // JSONB
	Metadata    json.RawMessage // JSONB
}

// AlarmInsertResult INSERT 后返回的结果
type AlarmInsertResult struct {
	EventID     string
	TriggeredAt time.Time
}

// AlarmUpdateParams UPDATE alarm_events 所需参数
type AlarmUpdateParams struct {
	AlarmStatus   string          // required: "acked", "resolved", "auto_resolved", "expired"
	Handler       string          // required: handler user/device ID
	Operation     string          // optional: "verified", "false_alarm", "test", "auto_resolved"
	Notes         *string         // optional
	NotifiedUsers json.RawMessage // optional (nil = don't update)
	Metadata      json.RawMessage // optional (nil = don't update)
}

// CardAlarmState 从 cards 表读取的报警状态（用于构建 AlarmState）
type CardAlarmState struct {
	UnhandledAlarm0 int
	UnhandledAlarm1 int
	UnhandledAlarm2 int
	UnhandledAlarm3 int
	UnhandledAlarm4 int
	PopAlarmLevel   string
	PopAlarmType    string
	PopAlarmEventId string    // pop alarm 对应的 alarm_events.event_id
	HandTime        time.Time // UpdateAlarmAndUpdateCard 回填的实际 hand_time
}

// ToAlarmState 转换为 AlarmState（用于 Redis/SSE）
func (c *CardAlarmState) ToAlarmState() *AlarmState {
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
	}
	return s
}

// InsertAlarmAndUpdateCard 在一个事务中完成：
//  1. INSERT alarm_events
//  2. updateAlarmCount(+1)
//  3. setPopAlarmIfHigher: 如果新报警优先级 >= 当前 popAlarm，更新 popAlarm
//
// cardID 由调用方提供；为空时通过 DeviceID 自动查找，查不到返回 error。
func InsertAlarmAndUpdateCard(ctx context.Context, db *sql.DB, cardID string, params AlarmInsertParams) (*AlarmInsertResult, *CardAlarmState, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 默认值
	if len(params.TriggerData) == 0 {
		params.TriggerData = json.RawMessage("{}")
	}
	if len(params.Metadata) == 0 {
		params.Metadata = json.RawMessage("{}")
	}

	// 快照当前 device→unit/bed→resident 绑定关系到 metadata
	params.Metadata = snapshotBindingToMetadata(ctx, tx, params.TenantID, params.DeviceID, params.Metadata)

	// 1. INSERT alarm_events
	var eventID string
	insertSQL := `
		INSERT INTO alarm_events (
			tenant_id, device_id, event_type, category, alarm_level,
			alarm_status, triggered_at, trigger_data, metadata,
			notified_users, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, 'active', $6, $7, $8, '[]'::jsonb,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING event_id
	`
	err = tx.QueryRowContext(ctx, insertSQL,
		params.TenantID, params.DeviceID, params.EventType, params.Category, params.AlarmLevel,
		params.TriggeredAt, params.TriggerData, params.Metadata,
	).Scan(&eventID)
	if err != nil {
		return nil, nil, fmt.Errorf("insert alarm_events: %w", err)
	}

	// cardID 为空时通过 DeviceID 查找；若仍找不到（如 AI 虚拟设备），fallback 到 metadata.card_id
	if cardID == "" {
		cardID, err = findCardIDByDevice(ctx, tx, params.DeviceID)
		if err != nil {
			if cid := extractMetadataCardID(params.Metadata); cid != "" {
				cardID = cid
			} else {
				return nil, nil, fmt.Errorf("find card by device %s: %w", params.DeviceID, err)
			}
		}
	}

	// 2. updateAlarmCount(+1)
	counts, err := updateAlarmCount(ctx, tx, cardID, params.AlarmLevel, +1)
	if err != nil {
		return nil, nil, fmt.Errorf("update alarm count: %w", err)
	}

	// 3. setPopAlarmIfHigher: 新报警优先级 >= 当前 pop → 更新
	popLevel, popType, popEventId, err := setPopAlarmIfHigher(ctx, tx, cardID, params.AlarmLevel, params.EventType, eventID)
	if err != nil {
		return nil, nil, fmt.Errorf("set pop_alarm: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit: %w", err)
	}

	state := &CardAlarmState{
		UnhandledAlarm0: counts[0],
		UnhandledAlarm1: counts[1],
		UnhandledAlarm2: counts[2],
		UnhandledAlarm3: counts[3],
		UnhandledAlarm4: counts[4],
		PopAlarmLevel:   popLevel,
		PopAlarmType:    popType,
		PopAlarmEventId: popEventId,
	}

	return &AlarmInsertResult{EventID: eventID, TriggeredAt: params.TriggeredAt}, state, nil
}

// QueryCardAlarmState 直接查 DB 获取 card 的报警状态（供 cardagg HandleAlarmProcessMessage 使用）
func QueryCardAlarmState(ctx context.Context, db *sql.DB, cardID string) (*CardAlarmState, error) {
	if cardID == "" {
		return nil, fmt.Errorf("cardID is required")
	}
	var state CardAlarmState
	var popEventId sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT unhandled_alarm_0, unhandled_alarm_1, unhandled_alarm_2, unhandled_alarm_3, unhandled_alarm_4,
		       pop_alarm_level, pop_alarm_type, pop_alarm_event_id
		FROM cards WHERE card_id = $1
	`, cardID).Scan(
		&state.UnhandledAlarm0, &state.UnhandledAlarm1, &state.UnhandledAlarm2,
		&state.UnhandledAlarm3, &state.UnhandledAlarm4,
		&state.PopAlarmLevel, &state.PopAlarmType, &popEventId,
	)
	if err != nil {
		return nil, err
	}
	if popEventId.Valid {
		state.PopAlarmEventId = popEventId.String
	}
	return &state, nil
}

// UpdateAlarmAndUpdateCard 更新报警状态 + 更新 cards（单事务）：
//  1. SELECT FOR UPDATE 获取旧状态 + device_id + alarm_level
//  2. UPDATE alarm_events SET alarm_status, handler, hand_time, updated_at + optional fields
//  3. 旧状态为 active 时 updateAlarmCount(-1)（acked 状态已在 ack 时减过）
//  4. clearPopAlarmIfMatch: 如果处理的 eventID == card.popAlarmEventId，清空 popAlarm
//
// 适用于 ack / resolve / auto_resolve / expired 所有状态转换。
// cardID 由调用方提供；为空时通过 device_id 自动查找 card。
func UpdateAlarmAndUpdateCard(ctx context.Context, db *sql.DB, cardID, tenantID, eventID string, params AlarmUpdateParams) (*CardAlarmState, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. SELECT FOR UPDATE 获取旧状态 + metadata 中的 card_id（AI 报警 fallback）
	var oldStatus, deviceID, alarmLevel string
	var metaCardID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT alarm_status, device_id, alarm_level, metadata->>'card_id'
		FROM alarm_events
		WHERE event_id = $1 AND tenant_id = $2
		FOR UPDATE
	`, eventID, tenantID).Scan(&oldStatus, &deviceID, &alarmLevel, &metaCardID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alarm event %s not found", eventID)
		}
		return nil, fmt.Errorf("read alarm event: %w", err)
	}
	if oldStatus != "active" && oldStatus != "acked" {
		return nil, fmt.Errorf("alarm event %s not in active/acked state (current: %s)", eventID, oldStatus)
	}

	// 2. UPDATE alarm_events — always: alarm_status, handler, hand_time, updated_at
	q := `UPDATE alarm_events SET alarm_status = $1, handler = $2, hand_time = NOW(), updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{params.AlarmStatus, params.Handler}
	argIdx := 3
	if params.Operation != "" {
		q += fmt.Sprintf(", operation = $%d", argIdx)
		args = append(args, params.Operation)
		argIdx++
	}
	if params.Notes != nil {
		q += fmt.Sprintf(", notes = $%d", argIdx)
		args = append(args, *params.Notes)
		argIdx++
	}
	if params.NotifiedUsers != nil {
		q += fmt.Sprintf(", notified_users = $%d", argIdx)
		args = append(args, params.NotifiedUsers)
		argIdx++
	}
	if params.Metadata != nil {
		q += fmt.Sprintf(", metadata = $%d", argIdx)
		args = append(args, params.Metadata)
		argIdx++
	}
	q += fmt.Sprintf(" WHERE event_id = $%d AND tenant_id = $%d RETURNING hand_time", argIdx, argIdx+1)
	args = append(args, eventID, tenantID)
	var handTime time.Time
	err = tx.QueryRowContext(ctx, q, args...).Scan(&handTime)
	if err != nil {
		return nil, fmt.Errorf("update alarm_events: %w", err)
	}

	// cardID 为空 或 不存在 时通过 deviceID 查找；再 fallback 到 metadata.card_id
	if cardID == "" || !cardExistsInDB(ctx, tx, cardID) {
		found, err2 := findCardIDByDevice(ctx, tx, deviceID)
		if err2 == nil {
			cardID = found
		} else if metaCardID.Valid && metaCardID.String != "" {
			cardID = metaCardID.String
		} else if cardID == "" {
			return nil, fmt.Errorf("find card by device %s: %w", deviceID, err2)
		}
		// cardID 仍可能指向不存在的 card（全部 fallback 失败），下面 cardExists 再兜底
	}

	// 3. 旧状态为 active 时才减计数（acked 状态已在 ack 时减过）
	// 如果 card 不存在（已删除），跳过计数更新，alarm_events 状态仍正常更新
	delta := 0
	if oldStatus == "active" {
		delta = -1
	}
	cardExists := cardExistsInDB(ctx, tx, cardID)
	var counts [5]int
	if cardExists {
		counts, err = updateAlarmCount(ctx, tx, cardID, alarmLevel, delta)
		if err != nil {
			return nil, fmt.Errorf("update alarm count: %w", err)
		}
	}

	// 4. clearPopAlarmIfMatch: 处理的 eventID == pop → 清空
	var popLevel, popType, popEventId string
	if cardExists {
		popLevel, popType, popEventId, err = clearPopAlarmIfMatch(ctx, tx, cardID, eventID)
		if err != nil {
			return nil, fmt.Errorf("clear pop_alarm: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &CardAlarmState{
		UnhandledAlarm0: counts[0],
		UnhandledAlarm1: counts[1],
		UnhandledAlarm2: counts[2],
		UnhandledAlarm3: counts[3],
		UnhandledAlarm4: counts[4],
		PopAlarmLevel:   popLevel,
		PopAlarmType:    popType,
		PopAlarmEventId: popEventId,
		HandTime:        handTime,
	}, nil
}

// RecalcCardAlarmState 仅重新统计 + 更新 cards（不改 alarm_events）
// 供 cardagg HandleAlarmProcessMessage、外部恢复等场景使用
//
// cardID 必须由调用方提供，为空返回 error。
func RecalcCardAlarmState(ctx context.Context, db *sql.DB, cardID, tenantID string) (*CardAlarmState, error) {
	if cardID == "" {
		return nil, fmt.Errorf("cardID is required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := recalcAndUpdateCards(ctx, tx, cardID, tenantID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return state, nil
}

// ========== 内部辅助函数 ==========

// recalcAndUpdateCards 内部复用：通过 cards.devices 获取设备列表，统计 + 更新 cards
// 注意：此函数会丢失 device_id 不在 cards.devices 中的报警（如 AI 报警），慎用
func recalcAndUpdateCards(ctx context.Context, tx *sql.Tx, cardID, tenantID string) (*CardAlarmState, error) {
	deviceIDs, err := getCardDeviceIDs(ctx, tx, cardID)
	if err != nil {
		return nil, fmt.Errorf("get card devices: %w", err)
	}

	counts, err := countActiveAlarms(ctx, tx, tenantID, deviceIDs)
	if err != nil {
		return nil, fmt.Errorf("count active alarms: %w", err)
	}

	newPopLevel, newPopType, newPopEventId, err := findTopActiveAlarm(ctx, tx, tenantID, deviceIDs)
	if err != nil {
		return nil, fmt.Errorf("find top active alarm: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE cards SET
			unhandled_alarm_0 = $2, unhandled_alarm_1 = $3, unhandled_alarm_2 = $4,
			unhandled_alarm_3 = $5, unhandled_alarm_4 = $6,
			pop_alarm_level = $7, pop_alarm_type = $8, pop_alarm_event_id = $9
		WHERE card_id = $1
	`, cardID, counts[0], counts[1], counts[2], counts[3], counts[4],
		newPopLevel, newPopType, toNullString(newPopEventId))
	if err != nil {
		return nil, fmt.Errorf("update cards: %w", err)
	}

	return &CardAlarmState{
		UnhandledAlarm0: counts[0],
		UnhandledAlarm1: counts[1],
		UnhandledAlarm2: counts[2],
		UnhandledAlarm3: counts[3],
		UnhandledAlarm4: counts[4],
		PopAlarmLevel:   newPopLevel,
		PopAlarmType:    newPopType,
		PopAlarmEventId: newPopEventId,
	}, nil
}

// updateAlarmCount 增量更新 card 的某个级别的 alarm count
// delta: +1 (new active) or -1 (ack/resolve)
// 返回更新后的 5 个 counts
func updateAlarmCount(ctx context.Context, tx *sql.Tx, cardID, alarmLevel string, delta int) ([5]int, error) {
	var counts [5]int
	idx, ok := alarm.AlarmLevelPriority[alarmLevel]
	if !ok || idx < 0 || idx > 4 {
		return counts, fmt.Errorf("invalid alarm level for count update: %s", alarmLevel)
	}

	col := fmt.Sprintf("unhandled_alarm_%d", idx)
	query := fmt.Sprintf(`
		UPDATE cards SET %s = GREATEST(%s + $2, 0)
		WHERE card_id = $1
		RETURNING unhandled_alarm_0, unhandled_alarm_1, unhandled_alarm_2, unhandled_alarm_3, unhandled_alarm_4
	`, col, col)

	err := tx.QueryRowContext(ctx, query, cardID, delta).Scan(
		&counts[0], &counts[1], &counts[2], &counts[3], &counts[4],
	)
	if err != nil {
		return counts, fmt.Errorf("updateAlarmCount card=%s level=%s delta=%d: %w", cardID, alarmLevel, delta, err)
	}
	return counts, nil
}

// alarmLevelToInt 将 alarm level 字符串转为优先级数字（数字越小优先级越高）
// 未知 level 返回 99（最低优先级）
func alarmLevelToInt(level string) int {
	if p, ok := alarm.AlarmLevelPriority[level]; ok {
		return p
	}
	switch level {
	case "0":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	}
	return 99
}

// readCurrentPop 读取 card 当前的 pop_alarm 字段
func readCurrentPop(ctx context.Context, tx *sql.Tx, cardID string) (level, typ, eventId string, err error) {
	var eid sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT pop_alarm_level, pop_alarm_type, pop_alarm_event_id
		FROM cards WHERE card_id = $1
	`, cardID).Scan(&level, &typ, &eid)
	if err != nil {
		return
	}
	if eid.Valid {
		eventId = eid.String
	}
	return
}

// writePopAlarm 写入 card 的 pop_alarm 字段（level/type 为空字符串 + event_id 为 NULL 表示清空）
func writePopAlarm(ctx context.Context, tx *sql.Tx, cardID, level, typ, eventId string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE cards SET pop_alarm_level = $2, pop_alarm_type = $3, pop_alarm_event_id = $4
		WHERE card_id = $1
	`, cardID, level, typ, toNullString(eventId))
	return err
}

// setPopAlarmIfHigher Insert 专用：如果新报警优先级 >= 当前 popAlarm，更新 popAlarm
// 返回更新后的 pop_alarm 值
func setPopAlarmIfHigher(ctx context.Context, tx *sql.Tx, cardID, newLevel, newType, newEventID string) (string, string, string, error) {
	curLevel, curType, curEid, err := readCurrentPop(ctx, tx, cardID)
	if err != nil {
		return "", "", "", err
	}

	// 无当前 pop 或新报警优先级 >= 当前 pop → 更新
	if curLevel == "" || alarmLevelToInt(newLevel) <= alarmLevelToInt(curLevel) {
		if err := writePopAlarm(ctx, tx, cardID, newLevel, newType, newEventID); err != nil {
			return "", "", "", err
		}
		return newLevel, newType, newEventID, nil
	}

	// 当前 pop 优先级更高，保持不变
	return curLevel, curType, curEid, nil
}

// clearPopAlarmIfMatch Update 专用：如果处理的 eventID == card 当前 popAlarm 的 event_id，清空 popAlarm
// 返回更新后的 pop_alarm 值
func clearPopAlarmIfMatch(ctx context.Context, tx *sql.Tx, cardID, eventID string) (string, string, string, error) {
	curLevel, curType, curEid, err := readCurrentPop(ctx, tx, cardID)
	if err != nil {
		return "", "", "", err
	}

	if curEid != "" && curEid == eventID {
		if err := writePopAlarm(ctx, tx, cardID, "", "", ""); err != nil {
			return "", "", "", err
		}
		return "", "", "", nil
	}

	return curLevel, curType, curEid, nil
}

// clearPopAlarmIfResolved AutoResolve 专用：如果 popAlarm 的 event 已不是 active，清空 popAlarm
// 返回更新后的 pop_alarm 值
func clearPopAlarmIfResolved(ctx context.Context, tx *sql.Tx, cardID string) (string, string, string, error) {
	curLevel, curType, curEid, err := readCurrentPop(ctx, tx, cardID)
	if err != nil {
		return "", "", "", err
	}

	if curEid == "" {
		return curLevel, curType, "", nil
	}

	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT alarm_status FROM alarm_events WHERE event_id = $1
	`, curEid).Scan(&status)
	if err != nil || status != "active" {
		// event 不存在或已不是 active → 清空 pop
		if err := writePopAlarm(ctx, tx, cardID, "", "", ""); err != nil {
			return "", "", "", err
		}
		return "", "", "", nil
	}

	return curLevel, curType, curEid, nil
}

// cardExistsInDB 检查 card 是否存在
func cardExistsInDB(ctx context.Context, tx *sql.Tx, cardID string) bool {
	if cardID == "" {
		return false
	}
	var exists bool
	_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cards WHERE card_id = $1)`, cardID).Scan(&exists)
	return exists
}

// findCardIDByDevice 通过 device_id 查找所属的 card_id
// 从 cards.devices JSONB 中查找包含该 device_id 的卡片
func findCardIDByDevice(ctx context.Context, tx *sql.Tx, deviceID string) (string, error) {
	var cardID string
	err := tx.QueryRowContext(ctx, `
		SELECT card_id FROM cards
		WHERE devices @> $1::jsonb
		LIMIT 1
	`, fmt.Sprintf(`[{"device_id":"%s"}]`, deviceID)).Scan(&cardID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no card found for device %s", deviceID)
		}
		return "", err
	}
	return cardID, nil
}

// getCardDeviceIDs 从 cards.devices JSONB 中提取 device_id 列表
func getCardDeviceIDs(ctx context.Context, tx *sql.Tx, cardID string) ([]string, error) {
	var devicesJSON sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT devices FROM cards WHERE card_id = $1`, cardID).Scan(&devicesJSON)
	if err != nil {
		return nil, err
	}
	if !devicesJSON.Valid || devicesJSON.String == "" || devicesJSON.String == "[]" {
		return nil, nil
	}
	var devices []map[string]interface{}
	if err := json.Unmarshal([]byte(devicesJSON.String), &devices); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(devices))
	for _, d := range devices {
		if did, ok := d["device_id"].(string); ok && did != "" {
			ids = append(ids, did)
		}
	}
	return ids, nil
}

// countActiveAlarms 按 device_id 列表 + alarm_level 分组统计 active alarm 数量，返回 [5]int (0~4)
func countActiveAlarms(ctx context.Context, tx *sql.Tx, tenantID string, deviceIDs []string) ([5]int, error) {
	var counts [5]int
	if len(deviceIDs) == 0 {
		return counts, nil
	}
	query := `
		SELECT
			COUNT(*) FILTER (WHERE alarm_level IN ('0', 'EMERG'))   AS c0,
			COUNT(*) FILTER (WHERE alarm_level IN ('1', 'ALERT'))   AS c1,
			COUNT(*) FILTER (WHERE alarm_level IN ('2', 'CRIT'))    AS c2,
			COUNT(*) FILTER (WHERE alarm_level IN ('3', 'ERR'))     AS c3,
			COUNT(*) FILTER (WHERE alarm_level IN ('4', 'WARNING')) AS c4
		FROM alarm_events
		WHERE tenant_id = $1
		  AND device_id = ANY($2::uuid[])
		  AND alarm_status = 'active'
		  AND (metadata->>'deleted_at' IS NULL)
	`
	err := tx.QueryRowContext(ctx, query, tenantID, pq.Array(deviceIDs)).Scan(
		&counts[0], &counts[1], &counts[2], &counts[3], &counts[4],
	)
	return counts, err
}

// findTopActiveAlarm 按 device_id 列表查找最高级别的最新 active alarm
// 返回 (alarm_level, event_type, event_id)；无 active alarm 时返回空值
func findTopActiveAlarm(ctx context.Context, tx *sql.Tx, tenantID string, deviceIDs []string) (string, string, string, error) {
	var level, eventType, eventId string

	if len(deviceIDs) == 0 {
		return "", "", "", nil
	}

	query := `
		SELECT alarm_level, event_type, event_id
		FROM alarm_events
		WHERE tenant_id = $1
		  AND device_id = ANY($2::uuid[])
		  AND alarm_status = 'active'
		  AND alarm_level IN ('0', '1', '2', '3', '4', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING')
		  AND (metadata->>'deleted_at' IS NULL)
		ORDER BY
			CASE alarm_level
				WHEN '0' THEN 0 WHEN 'EMERG' THEN 0
				WHEN '1' THEN 1 WHEN 'ALERT' THEN 1
				WHEN '2' THEN 2 WHEN 'CRIT' THEN 2
				WHEN '3' THEN 3 WHEN 'ERR' THEN 3
				WHEN '4' THEN 4 WHEN 'WARNING' THEN 4
			END ASC,
			triggered_at DESC
		LIMIT 1
	`
	err := tx.QueryRowContext(ctx, query, tenantID, pq.Array(deviceIDs)).Scan(&level, &eventType, &eventId)
	if err == sql.ErrNoRows {
		return "", "", "", nil
	}
	if err != nil {
		return "", "", "", err
	}
	return level, eventType, eventId, nil
}

// toNullString 将空字符串转为 sql.NullString（用于 nullable UUID 列）
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// extractMetadataCardID 从 metadata JSONB 中提取 card_id（AI 报警 fallback）
func extractMetadataCardID(metadata json.RawMessage) string {
	if len(metadata) == 0 {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(metadata, &m); err != nil {
		return ""
	}
	if cid, ok := m["card_id"].(string); ok {
		return cid
	}
	return ""
}

// DeviceSelfRecoveryAlarmTypes 设备恢复时需自动解除的报警类型
var DeviceSelfRecoveryAlarmTypes = []string{
	"OfflineAlarm", "DeviceFailure", "AngleException", "SignalPoor", "SensorDetached",
}

// AutoResolveResult 设备自恢复的返回结果
type AutoResolveResult struct {
	ResolvedCount int    // 被 resolve 的报警总数
	TopEventId    string // triggered_at 最新那条的 event_id
	TopAlarmLevel string // 最新那条的 alarm_level
}

// AutoResolveDeviceAlarms 设备恢复时，自动解除该设备的 active 报警
// alarmTypes == nil → resolve 所有 DeviceSelfRecoveryAlarmTypes
// alarmTypes != nil → 只 resolve 指定的类型（如 []string{"SignalPoor"}）
//
// 流程：
//  1. 先查出即将被 resolve 的 alarm 按 level 分组计数 + 最新 event_id
//  2. batch UPDATE alarm_status='auto_resolved'
//  3. 对每个 level 调 updateAlarmCount(cardID, level, -count)
//  4. clearPopAlarmIfResolved: 如果 popAlarm 的 event 已不是 active，清空 popAlarm
//
// 不用 recalcAndUpdateCards，避免丢失 AI 报警。
func AutoResolveDeviceAlarms(ctx context.Context, db *sql.DB, cardID, tenantID, deviceID string, alarmTypes []string) (*CardAlarmState, *AutoResolveResult, error) {
	if len(alarmTypes) == 0 {
		alarmTypes = DeviceSelfRecoveryAlarmTypes
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. 查出将被 resolve 的 alarm 按 level 分组计数 + 最新 event_id
	type levelCount struct {
		Level string
		Count int
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT alarm_level, COUNT(*)
		FROM alarm_events
		WHERE device_id = $1
		  AND tenant_id = $2
		  AND alarm_status = 'active'
		  AND event_type = ANY($3::text[])
		  AND (metadata->>'deleted_at' IS NULL)
		GROUP BY alarm_level
	`, deviceID, tenantID, pq.Array(alarmTypes))
	if err != nil {
		return nil, nil, fmt.Errorf("count alarms to resolve: %w", err)
	}
	var levelCounts []levelCount
	for rows.Next() {
		var lc levelCount
		if err := rows.Scan(&lc.Level, &lc.Count); err != nil {
			rows.Close()
			return nil, nil, err
		}
		levelCounts = append(levelCounts, lc)
	}
	rows.Close()

	totalResolved := 0
	for _, lc := range levelCounts {
		totalResolved += lc.Count
	}

	if totalResolved == 0 {
		return nil, &AutoResolveResult{}, nil
	}

	// 查最新那条的 event_id + alarm_level
	var topEventId, topAlarmLevel string
	err = tx.QueryRowContext(ctx, `
		SELECT event_id, alarm_level
		FROM alarm_events
		WHERE device_id = $1
		  AND tenant_id = $2
		  AND alarm_status = 'active'
		  AND event_type = ANY($3::text[])
		  AND (metadata->>'deleted_at' IS NULL)
		ORDER BY triggered_at DESC
		LIMIT 1
	`, deviceID, tenantID, pq.Array(alarmTypes)).Scan(&topEventId, &topAlarmLevel)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, fmt.Errorf("find top resolved alarm: %w", err)
	}

	// 2. batch UPDATE: active → auto_resolved
	_, err = tx.ExecContext(ctx, `
		UPDATE alarm_events
		SET alarm_status = 'auto_resolved',
		    handler = 'system:device_recovery',
		    hand_time = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP,
		    operation = 'auto_resolved'
		WHERE device_id = $1
		  AND tenant_id = $2
		  AND alarm_status = 'active'
		  AND event_type = ANY($3::text[])
		  AND (metadata->>'deleted_at' IS NULL)
	`, deviceID, tenantID, pq.Array(alarmTypes))
	if err != nil {
		return nil, nil, fmt.Errorf("auto resolve device alarms: %w", err)
	}

	// 3. cardID
	if cardID == "" {
		cardID, err = findCardIDByDevice(ctx, tx, deviceID)
		if err != nil {
			return nil, nil, fmt.Errorf("find card by device %s: %w", deviceID, err)
		}
	}

	// 4. 对每个 level 调 updateAlarmCount(cardID, level, -count)
	var counts [5]int
	for _, lc := range levelCounts {
		counts, err = updateAlarmCount(ctx, tx, cardID, lc.Level, -lc.Count)
		if err != nil {
			return nil, nil, fmt.Errorf("update alarm count for level %s: %w", lc.Level, err)
		}
	}
	// 如果只有一个 level，counts 已经是最终值；多个 level 时最后一次 RETURNING 是最终值
	if len(levelCounts) == 0 {
		// 不应该走到这里（totalResolved > 0），但防御
		return nil, &AutoResolveResult{}, nil
	}

	// 5. clearPopAlarmIfResolved: popAlarm 的 event 不再 active → 清空
	popLevel, popType, popEventId, err := clearPopAlarmIfResolved(ctx, tx, cardID)
	if err != nil {
		return nil, nil, fmt.Errorf("clear pop_alarm: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit: %w", err)
	}

	state := &CardAlarmState{
		UnhandledAlarm0: counts[0],
		UnhandledAlarm1: counts[1],
		UnhandledAlarm2: counts[2],
		UnhandledAlarm3: counts[3],
		UnhandledAlarm4: counts[4],
		PopAlarmLevel:   popLevel,
		PopAlarmType:    popType,
		PopAlarmEventId: popEventId,
	}

	result := &AutoResolveResult{
		ResolvedCount: totalResolved,
		TopEventId:    topEventId,
		TopAlarmLevel: topAlarmLevel,
	}

	return state, result, nil
}

// ExpireAlarmsByDeviceIDs 将指定设备的 active/acked 报警标记为 expired
// reason: "card_update" | "card_delete" 等，写入 metadata
func ExpireAlarmsByDeviceIDs(ctx context.Context, db *sql.DB, tenantID string, deviceIDs []string, reason string) error {
	if tenantID == "" || len(deviceIDs) == 0 {
		return nil
	}

	query := `
		UPDATE alarm_events
		SET alarm_status = 'expired',
		    metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object(
		        'expired_reason', $3::text,
		        'expired_at', to_char(CURRENT_TIMESTAMP, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		    ),
		    updated_at = CURRENT_TIMESTAMP
		WHERE tenant_id = $1
		  AND device_id = ANY($2::uuid[])
		  AND alarm_status IN ('active', 'acked')
		  AND (metadata->>'deleted_at' IS NULL)
	`

	_, err := db.ExecContext(ctx, query, tenantID, pq.Array(deviceIDs), reason)
	if err != nil {
		return fmt.Errorf("expire alarms by device IDs: %w", err)
	}
	return nil
}

// snapshotBindingToMetadata 在报警触发时快照 device 当前绑定关系到 metadata.snapshot
//
// metadata.snapshot 结构：
//
//	{
//	  "device_name":    "Radar-001",
//	  "device_type":    "Radar",
//	  "card_id":        "...",
//	  "branch_id":      "...",
//	  "branch_name":    "Branch A",
//	  "building_name":  "Building 1",
//	  "floor":          "1F",
//	  "unit_id":        "...",
//	  "unit_name":      "Room 101",
//	  "unit_type":      "facility",
//	  "room_id":        "...",
//	  "room_name":      "Room 101",
//	  "bed_id":         "...",
//	  "bed_name":       "Bed A",
//	  "resident_id":    "...",
//	  "resident_name":  "Fox"
//	}
func snapshotBindingToMetadata(ctx context.Context, tx *sql.Tx, tenantID, deviceID string, metadata json.RawMessage) json.RawMessage {
	if tenantID == "" || deviceID == "" {
		return metadata
	}

	// 用 SAVEPOINT 保护：查询失败不会导致事务 abort
	var snapErrors []string
	safeQuery := func(label string, fn func() error) {
		if _, err := tx.ExecContext(ctx, "SAVEPOINT snap_query"); err != nil {
			snapErrors = append(snapErrors, fmt.Sprintf("%s: savepoint: %s", label, err.Error()))
			return
		}
		if err := fn(); err != nil && err != sql.ErrNoRows {
			snapErrors = append(snapErrors, fmt.Sprintf("%s: %s", label, err.Error()))
			tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT snap_query")
		} else {
			tx.ExecContext(ctx, "RELEASE SAVEPOINT snap_query")
		}
	}

	// 一次 JOIN 查出 device → bed/room → unit → branch/building 全链路
	var (
		deviceName, deviceType     sql.NullString
		boundBedID, bedName        sql.NullString
		roomID, roomName           sql.NullString
		unitID, unitName, unitType sql.NullString
		floor                      sql.NullString
		isPublic, isSharedUnit     sql.NullBool
		branchID, branchName       sql.NullString
		buildingName               sql.NullString
	)
	safeQuery("device_binding", func() error {
		return tx.QueryRowContext(ctx, `
			SELECT
				d.device_name,  ds.device_type,
				d.bound_bed_id::text,   b.bed_name,
				COALESCE(rb.room_id, rr.room_id)::text,
				COALESCE(rb.room_name, rr.room_name),
				u.unit_id::text,  u.unit_name,  u.unit_type,
				u.floor,          u.is_public,  u.is_shared_unit,
				br.branch_id::text,  br.branch_name,
				bld.building_name
			FROM devices d
			LEFT JOIN device_store ds ON d.device_id = ds.device_id
			LEFT JOIN beds    b   ON d.bound_bed_id  = b.bed_id
			LEFT JOIN rooms   rb  ON b.room_id       = rb.room_id
			LEFT JOIN rooms   rr  ON d.bound_room_id = rr.room_id
			LEFT JOIN units   u   ON COALESCE(rb.unit_id, rr.unit_id) = u.unit_id
			LEFT JOIN branches br ON u.branch_id     = br.branch_id
			LEFT JOIN buildings bld ON u.building_id  = bld.building_id
			WHERE d.device_id = $1 AND d.tenant_id = $2
		`, deviceID, tenantID).Scan(
			&deviceName, &deviceType,
			&boundBedID, &bedName,
			&roomID, &roomName,
			&unitID, &unitName, &unitType,
			&floor, &isPublic, &isSharedUnit,
			&branchID, &branchName,
			&buildingName,
		)
	})

	// 查 card_id
	var cardID sql.NullString
	safeQuery("card_id", func() error {
		return tx.QueryRowContext(ctx, `
			SELECT c.card_id::text
			FROM cards c
			WHERE c.tenant_id = $1
			  AND c.devices @> jsonb_build_array(jsonb_build_object('device_id', $2::text))
			LIMIT 1
		`, tenantID, deviceID).Scan(&cardID)
	})

	// 查 resident：bed → resident，fallback unit → resident
	var residentID, residentName sql.NullString
	if boundBedID.Valid && boundBedID.String != "" {
		safeQuery("resident_by_bed", func() error {
			return tx.QueryRowContext(ctx, `
				SELECT resident_id::text, nickname FROM residents
				WHERE tenant_id = $1 AND bed_id = $2
				ORDER BY created_at ASC LIMIT 1
			`, tenantID, boundBedID.String).Scan(&residentID, &residentName)
		})
	}
	if !residentID.Valid && unitID.Valid && unitID.String != "" {
		canResolve := false
		ut := unitType.String
		if strings.EqualFold(ut, "home") {
			canResolve = true
		} else if strings.EqualFold(ut, "facility") && !(isPublic.Valid && isPublic.Bool) && !(isSharedUnit.Valid && isSharedUnit.Bool) {
			canResolve = true
		}
		if canResolve {
			safeQuery("resident_by_unit", func() error {
				return tx.QueryRowContext(ctx, `
					SELECT resident_id::text, nickname FROM residents
					WHERE tenant_id = $1 AND unit_id = $2
					ORDER BY created_at ASC LIMIT 1
				`, tenantID, unitID.String).Scan(&residentID, &residentName)
			})
		}
	}

	// 构建 snapshot 对象（只写非空字段）
	snap := map[string]interface{}{}
	setStr := func(k string, v sql.NullString) {
		if v.Valid && v.String != "" {
			snap[k] = v.String
		}
	}
	setStr("device_name", deviceName)
	setStr("device_type", deviceType)
	setStr("card_id", cardID)
	setStr("branch_id", branchID)
	setStr("branch_name", branchName)
	setStr("building_name", buildingName)
	setStr("floor", floor)
	setStr("unit_id", unitID)
	setStr("unit_name", unitName)
	setStr("unit_type", unitType)
	setStr("room_id", roomID)
	setStr("room_name", roomName)
	setStr("bed_id", boundBedID)
	setStr("bed_name", bedName)
	setStr("resident_id", residentID)
	setStr("resident_name", residentName)
	if len(snapErrors) > 0 {
		snap["_errors"] = snapErrors
	}

	// 合并到 metadata
	var m map[string]interface{}
	if err := json.Unmarshal(metadata, &m); err != nil {
		m = map[string]interface{}{}
	}
	m["snapshot"] = snap

	out, err := json.Marshal(m)
	if err != nil {
		return metadata
	}
	return out
}
