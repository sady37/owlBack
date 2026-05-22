package card

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/lib/pq"

	"owl-common/alarm"
)

// ====================================================================
// alarm_db.go — v2 Phase G 实施版（2026-05-14）
//
// 设计要点：
//   - device_addr INET /128 是核心定位 key（device_ipv6 单程票）
//   - alarm_events v2 列：event_type / category / alarm_level SMALLINT (syslog 0-7) /
//     alarm_status (active/acked/resolved/auto_resolved/expired) / payload / evidence
//   - snapshot 列（tenant_name/branch_name/unit_name/room_name/bed_name/resident_nickname/
//     device_uid + resident_id INET / device_id UUID）— trigger 时刻一次 SELECT 锁住
//   - 北极星：producer / parent_span / trace_id 三列承载 datagram 因果链
//   - cards 表 v2 无 alarm counter 列；CardAlarmState 由 alarm_events 实时聚合（GROUP BY alarm_level）
//   - card_id INET 列（v2 schema 已是 INET，等同 cards.spatial_prefix CIDR）：INSERT 时填入 cardID
//     （由 cardagg IotPreparedHandler LPM 反查 device_addr → cards.spatial_prefix；空时仍写 NULL）
//
// 详 doc/alarm_v2_phase_g.md（待补）
// ====================================================================

// AlarmInsertParams INSERT alarm_events 所需参数
type AlarmInsertParams struct {
	TenantID    string
	DeviceAddr  string          // IPv6 /128 canonical host text（device_ipv6 单程票）
	EventType   string          // 映射 alarm_events.event_type
	Category    string
	AlarmLevel  string          // 字符串名（EMERG/ALERT/...）；写库前 normalize → SMALLINT
	TriggeredAt time.Time
	TriggerData json.RawMessage // 映射 payload
	Metadata    json.RawMessage // 映射 evidence
	RoomID      string
	UnitID      string
	// 北极星 envelope（详 alarm_events.producer / parent_span / trace_id 列注释）
	Producer   string
	ParentSpan string
	TraceID    string
}

// AlarmInsertResult INSERT 后返回的结果
type AlarmInsertResult struct {
	EventID       string
	TriggeredAt   time.Time
	Deduped       bool // 命中 active 同类 → 不重复插入
	SkippedNotify bool // device-class（SkipUnhandledCount=true）— 已落库审计但跳过 pop/notify
}

// AlarmUpdateParams UPDATE alarm_events 所需参数
type AlarmUpdateParams struct {
	AlarmStatus     string          // active/acked/resolved/auto_resolved/expired
	Handler         string          // user_id UUID
	Operation       string          // verified/false_alarm/test/auto_resolved
	Notes           *string         // handler_notes
	NotifiedUsers   json.RawMessage // 写入 evidence.notified_users
	Metadata        json.RawMessage // 合并到 evidence
	ResolveSnapshot json.RawMessage // 写入 evidence.resolve_snapshot
}

// CardAlarmState v2 实时聚合：按 sensor_v2.md §6.7 三级告警状态机（决定 17）—
//   Critical 级别（0/1/2）计入 status ∈ {active, acked, auto_resolved}（resolved 终态不计）；
//   Error/Warning（3/4）仅计入 status=active；
//   PopAlarm 仅从 status=active 行挑选（acked/auto_resolved 不参与 popAlarm 显示，但仍维持 Bell）。
//
// 字段名沿用 v1 但语义升级：v1 "UnhandledAlarm" 含义模糊（"未处理"），v2 明确按 §6.7 规则分级；
// 改字段名跨服务影响大，注释明示语义即可。
type CardAlarmState struct {
	UnhandledAlarm0  int    // EMERG  (lvl 0) — Critical：count(status ∈ {active, acked, auto_resolved})
	UnhandledAlarm1  int    // ALERT  (lvl 1) — Critical：同上
	UnhandledAlarm2  int    // CRIT   (lvl 2) — Critical：同上
	UnhandledAlarm3  int    // ERROR  (lvl 3) — Warning 语义：count(status = active)
	UnhandledAlarm4  int    // WARN   (lvl 4) — Warning 语义：同上
	PopAlarmLevel    string // 最高优先级 (lvl 数字最小) + 最新 active 行的字符串级别
	PopAlarmType     string // 同行 event_type
	PopAlarmEventId  string
	PopTriggeredAtMs int64
	HandTime         time.Time
}

// ToAlarmState 转换为前端可消费的 AlarmState（写 redis card:state hash 用）
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

// alarmSnapshotRow snapshot 一次 SELECT 得到的 denormalization 字段
type alarmSnapshotRow struct {
	TenantName       sql.NullString
	BranchName       sql.NullString
	UnitName         sql.NullString
	RoomName         sql.NullString
	BedName          sql.NullString
	ResidentNickname sql.NullString
	DeviceUID        sql.NullString
	ResidentID       sql.NullString // INET 文本
}

// snapshotForAlarm trigger 时刻一次 SELECT 锁住 denormalization 字段。
// device_addr 必须 valid /128；返回值各字段允许 NULL（spatial 层级未对齐 / 设备未导入工厂表 / 未绑卡 / 未住人 都正常）。
func snapshotForAlarm(ctx context.Context, db *sql.DB, addr netip.Addr) (alarmSnapshotRow, error) {
	var snap alarmSnapshotRow
	if !addr.IsValid() {
		return snap, fmt.Errorf("invalid device_addr")
	}
	// LATERAL: cards LPM 取最长前缀的 active card 一行
	const q = `
		SELECT
		  t.tenant_name, b.branch_name, u.unit_name, r.room_name, bd.bed_name,
		  res.nickname, dfm.device_uid,
		  CASE WHEN res.resident_id IS NULL THEN NULL ELSE host(res.resident_id) END
		FROM (SELECT $1::INET AS addr) x
		LEFT JOIN tenants  t  ON t.tenant_id  = network(set_masklen(x.addr, 48))
		LEFT JOIN branches b  ON b.branch_id  = network(set_masklen(x.addr, 56))
		LEFT JOIN units    u  ON u.unit_id    = network(set_masklen(x.addr, 80))
		LEFT JOIN rooms    r  ON r.room_id    = network(set_masklen(x.addr, 88))
		LEFT JOIN beds     bd ON bd.bed_id    = network(set_masklen(x.addr, 96))
		LEFT JOIN devices  d  ON d.device_addr = x.addr
		LEFT JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		LEFT JOIN LATERAL (
		  SELECT resident_id FROM cards
		  WHERE x.addr <<= spatial_prefix
		  ORDER BY masklen(spatial_prefix) DESC
		  LIMIT 1
		) c ON true
		LEFT JOIN residents res ON res.resident_id = c.resident_id
	`
	err := db.QueryRowContext(ctx, q, addr.String()).Scan(
		&snap.TenantName, &snap.BranchName, &snap.UnitName, &snap.RoomName, &snap.BedName,
		&snap.ResidentNickname, &snap.DeviceUID, &snap.ResidentID,
	)
	if err != nil && err != sql.ErrNoRows {
		return snap, fmt.Errorf("snapshotForAlarm: %w", err)
	}
	return snap, nil
}

// nullStringOrNil 字符串空值 → NULL；非空 → sql.NullString
func nullStringOrNil(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// alarmLevelToInt 字符串级别 → SMALLINT；未知级别 default ERROR(3)
func alarmLevelToInt(s string) int16 {
	if v, ok := alarm.AlarmLevelPriority[alarm.NormalizeAlarmLevel(s)]; ok {
		return int16(v)
	}
	return int16(alarm.AlarmLevelIntErr)
}

// findActiveAlarmEventID dedup 检查：同 device_addr + event_type 已存在 active alarm → 返回其 event_id；否则 ""
func findActiveAlarmEventID(ctx context.Context, db *sql.DB, addr netip.Addr, eventType string) (string, error) {
	var id string
	err := db.QueryRowContext(ctx, `
		SELECT event_id::text
		FROM alarm_events
		WHERE device_addr = $1::INET
		  AND event_type = $2
		  AND alarm_status = 'active'
		ORDER BY triggered_at DESC
		LIMIT 1
	`, addr.String(), eventType).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// InsertAlarmAndUpdateCard v2 主路径：
//   1. SkipUnhandledCount 类型（device-class）短路：直接落库（审计）+ 标 SkippedNotify=true
//   2. DedupWhileActive 类型：先查 active 同类，命中→返回 Deduped=true 不重复插
//   3. 一次 SELECT 拿 snapshot 9 字段
//   4. INSERT alarm_events
//   5. cards 表无 counter 列；返回 QueryCardAlarmState 实时聚合
//
// cardID 仅作日志/snapshot 参考；不更新 cards 表（v2 无 counter 列）。
// 入参 params.DeviceAddr 必须是 IPv6 /128 canonical host text。
func InsertAlarmAndUpdateCard(ctx context.Context, db *sql.DB, cardID string, params AlarmInsertParams) (*AlarmInsertResult, *CardAlarmState, error) {
	if params.DeviceAddr == "" {
		return nil, nil, fmt.Errorf("device_addr is required")
	}
	addr, err := netip.ParseAddr(params.DeviceAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("device_addr %q is not a valid IPv6: %w", params.DeviceAddr, err)
	}
	if params.EventType == "" {
		return nil, nil, fmt.Errorf("event_type is required")
	}
	if params.TriggeredAt.IsZero() {
		params.TriggeredAt = time.Now()
	}

	def := alarm.LookupAlarm(params.EventType)
	skipNotify := def != nil && def.SkipUnhandledCount

	// dedup：仅对 DedupWhileActive 类型生效（避免 Fall 等突发类被误压制）
	if def != nil && def.DedupWhileActive {
		existing, err := findActiveAlarmEventID(ctx, db, addr, params.EventType)
		if err != nil {
			return nil, nil, fmt.Errorf("dedup lookup: %w", err)
		}
		if existing != "" {
			return &AlarmInsertResult{EventID: existing, TriggeredAt: params.TriggeredAt, Deduped: true, SkippedNotify: skipNotify}, &CardAlarmState{}, nil
		}
	}

	// snapshot 锁定（spatial 层级 + dfm + cards LPM + residents）
	snap, err := snapshotForAlarm(ctx, db, addr)
	if err != nil {
		return nil, nil, err
	}

	// payload / evidence 兜底 {}
	payload := params.TriggerData
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}
	evidence := params.Metadata
	if len(evidence) == 0 {
		evidence = json.RawMessage("{}")
	}

	// INSERT
	var eventID string
	var triggeredAt time.Time
	insertSQL := `
		INSERT INTO alarm_events (
		  device_addr, triggered_at, event_type, category, alarm_level,
		  tenant_name, branch_name, unit_name, room_name, bed_name,
		  resident_nickname, device_uid,
		  resident_id, card_id,
		  trace_id, parent_span, producer,
		  alarm_status, payload, evidence
		) VALUES (
		  $1::INET, $2, $3, $4, $5,
		  $6, $7, $8, $9, $10,
		  $11, $12,
		  NULLIF($13, '')::INET, NULLIF($14, '')::INET,
		  NULLIF($15, ''), NULLIF($16, ''), NULLIF($17, '')::INET,
		  'active', $18::JSONB, $19::JSONB
		)
		RETURNING event_id::text, triggered_at
	`
	residentIDStr := ""
	if snap.ResidentID.Valid {
		residentIDStr = snap.ResidentID.String
	}
	err = db.QueryRowContext(ctx, insertSQL,
		addr.String(), params.TriggeredAt, params.EventType, nullStringOrNil(params.Category), alarmLevelToInt(params.AlarmLevel),
		snap.TenantName, snap.BranchName, snap.UnitName, snap.RoomName, snap.BedName,
		snap.ResidentNickname, snap.DeviceUID,
		residentIDStr, cardID,
		params.TraceID, params.ParentSpan, params.Producer,
		string(payload), string(evidence),
	).Scan(&eventID, &triggeredAt)
	if err != nil {
		return nil, nil, fmt.Errorf("insert alarm_events: %w", err)
	}

	// 实时聚合 CardAlarmState（cards 表无 counter 列）
	cas, _ := QueryCardAlarmState(ctx, db, cardID)
	if cas == nil {
		cas = &CardAlarmState{}
	}

	return &AlarmInsertResult{
		EventID:       eventID,
		TriggeredAt:   triggeredAt,
		SkippedNotify: skipNotify,
	}, cas, nil
}

// QueryCardAlarmState 实时聚合：**精确归属本卡** 的 active alarms（不 fan-out 到父卡）。
//
// 2026-05-16 修：旧实现用 `device_addr <<= cardID` prefix 匹配，导致 unit /80 卡聚合所有子卡
// (room /88, bed /96) 设备的 alarms — 违反用户拍板「alarm 只在 card_id 精确匹配的卡上出现」。
// 改用 `ae.card_id = cardID::INET` 精确匹配，依赖 InsertAlarmAndUpdateCard 在写入时由
// cards GiST LPM 锁定的 ae.card_id 列值。
//
// cardID 为 cards.spatial_prefix CIDR 字符串；空 / 非法 INET 时返回零状态。
//
// SQL 计数策略（sensor_v2.md §6.7 决定 17）：
//
//	level 0/1/2 (Critical：Emerg/Alert/Crit)
//	  → count(alarm_status IN ('active', 'acked', 'auto_resolved'))
//	  → 'resolved' 是终态，不计入（已离开 Pending+AlarmBell 进 Resolved 历史）
//
//	level 3/4 (Error/Warning)
//	  → count(alarm_status = 'active')
//	  → 'auto_resolved' 自动归 Resolved（不强制人工 ack）；'acked' 此分支理论不出现（不强制）
//
// 一条 SQL 用 CASE 表达级别相关的 status 子集，比 5 个 UNION ALL 子查询更清晰：
func QueryCardAlarmState(ctx context.Context, db *sql.DB, cardID string) (*CardAlarmState, error) {
	cas := &CardAlarmState{}
	if cardID == "" {
		return cas, nil
	}
	// 计数：按 alarm_level GROUP BY；level 相关的 alarm_status 过滤用 CASE 在 WHERE 子句表达。
	// Critical (0/1/2) 含 acked/auto_resolved（决定 17 — Critical 必须人工 ack 才离开 Pending+Bell）；
	// Error/Warning (3/4) 仅 active（auto_resolved 自动归 Resolved）。
	rows, err := db.QueryContext(ctx, `
		SELECT alarm_level, COUNT(*)
		FROM alarm_events
		WHERE card_id = $1::INET
		  AND (
		    (alarm_level IN (0, 1, 2) AND alarm_status IN ('active', 'acked', 'auto_resolved'))
		    OR
		    (alarm_level IN (3, 4) AND alarm_status = 'active')
		  )
		GROUP BY alarm_level
	`, cardID)
	if err != nil {
		return nil, fmt.Errorf("query alarm counts for %s: %w", cardID, err)
	}
	defer rows.Close()
	for rows.Next() {
		var lvl int16
		var cnt int
		if err := rows.Scan(&lvl, &cnt); err != nil {
			return nil, fmt.Errorf("scan alarm count: %w", err)
		}
		switch lvl {
		case 0:
			cas.UnhandledAlarm0 = cnt
		case 1:
			cas.UnhandledAlarm1 = cnt
		case 2:
			cas.UnhandledAlarm2 = cnt
		case 3:
			cas.UnhandledAlarm3 = cnt
		case 4:
			cas.UnhandledAlarm4 = cnt
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// PopAlarm pick：仅从 alarm_status='active' 行挑选（acked/auto_resolved 不参与 popAlarm 显示，
	// 但仍维持 AlarmBell — 由上面 counter 表达）。
	// 排序：alarm_level ASC（高优先级先 — 0=EMERG 优先）, triggered_at DESC（新 cover 旧）。
	var popLvl int16
	var popType string
	var popEvent string
	var popTriggered time.Time
	err = db.QueryRowContext(ctx, `
		SELECT alarm_level, event_type, event_id::text, triggered_at
		FROM alarm_events
		WHERE card_id = $1::INET
		  AND alarm_status = 'active'
		ORDER BY alarm_level ASC, triggered_at DESC
		LIMIT 1
	`, cardID).Scan(&popLvl, &popType, &popEvent, &popTriggered)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query pop alarm for %s: %w", cardID, err)
	}
	if err == nil {
		cas.PopAlarmLevel = intToAlarmLevelStr(popLvl)
		cas.PopAlarmType = popType
		cas.PopAlarmEventId = popEvent
		cas.PopTriggeredAtMs = popTriggered.UnixMilli()
	}
	return cas, nil
}

// intToAlarmLevelStr SMALLINT → 字符串级别（与 ToAlarmState() PopAlarm 拼接保持一致）
func intToAlarmLevelStr(lvl int16) string {
	switch lvl {
	case 0:
		return alarm.AlarmLevelEmerg
	case 1:
		return alarm.AlarmLevelAlert
	case 2:
		return alarm.AlarmLevelCrit
	case 3:
		return alarm.AlarmLevelErr
	case 4:
		return alarm.AlarmLevelWarn
	case 5:
		return alarm.AlarmLevelNotice
	case 6:
		return alarm.AlarmLevelInfo
	case 7:
		return alarm.AlarmLevelDebug
	default:
		return ""
	}
}

// UpdateAlarmAndUpdateCard 处理告警（ack/resolve/auto_resolved 等）；
// notes 为 nil 时 SQL 用 COALESCE 保留原值；返回更新后的实时聚合 CardAlarmState。
func UpdateAlarmAndUpdateCard(ctx context.Context, db *sql.DB, cardID, tenantID, eventID string, params AlarmUpdateParams) (*CardAlarmState, error) {
	if eventID == "" {
		return nil, fmt.Errorf("eventID is required")
	}
	if params.AlarmStatus == "" {
		return nil, fmt.Errorf("alarm_status is required")
	}
	// notes：nil → 保留 NULL/原值；非 nil → 替换
	var notesArg interface{}
	if params.Notes != nil {
		notesArg = *params.Notes
	}
	_, err := db.ExecContext(ctx, `
		UPDATE alarm_events
		SET alarm_status = $1,
		    hand_time    = NOW(),
		    handler      = NULLIF($2, '')::UUID,
		    operation    = NULLIF($3, ''),
		    handler_notes = COALESCE($4, handler_notes)
		WHERE event_id = $5::UUID
	`, params.AlarmStatus, params.Handler, params.Operation, notesArg, eventID)
	if err != nil {
		return nil, fmt.Errorf("update alarm %s: %w", eventID, err)
	}
	cas, qErr := QueryCardAlarmState(ctx, db, cardID)
	if qErr != nil {
		return nil, qErr
	}
	return cas, nil
}

// RecalcCardAlarmState v2 = QueryCardAlarmState（实时聚合，无 cache）
func RecalcCardAlarmState(ctx context.Context, db *sql.DB, cardID, tenantID string) (*CardAlarmState, error) {
	if cardID == "" {
		return nil, fmt.Errorf("cardID is required")
	}
	return QueryCardAlarmState(ctx, db, cardID)
}

// LookupCardIDByDeviceID 已删除（device_ipv6 单程票 R-001）。
// 使用 LookupCardByDeviceAddr(addr netip.Addr) 替代。

// DeviceSelfRecoveryAlarmTypes 设备恢复时需自动解除的报警类型。
// Elder care 策略：仅设备类可自动恢复；非设备类必须人工确认。
var DeviceSelfRecoveryAlarmTypes = []string{
	alarm.AlarmTypeOffline, alarm.AlarmTypeDeviceFailure, alarm.AngleException, alarm.SignalPoor, alarm.SensorDetached,
}

// AutoResolveResult 设备自恢复返回结果
type AutoResolveResult struct {
	ResolvedCount int
	TopEventId    string
	TopAlarmLevel string
}

// AutoResolveDeviceAlarms 设备恢复时批量 auto_resolved：device_addr + event_type ∈ alarmTypes + active。
// 入参 deviceAddr 是 IPv6 /128 canonical host text。
func AutoResolveDeviceAlarms(ctx context.Context, db *sql.DB, cardID, tenantID, deviceAddr string, alarmTypes []string) (*CardAlarmState, *AutoResolveResult, error) {
	if deviceAddr == "" || len(alarmTypes) == 0 {
		return &CardAlarmState{}, &AutoResolveResult{}, nil
	}
	addr, err := netip.ParseAddr(deviceAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("AutoResolveDeviceAlarms: device_addr %q not IPv6: %w", deviceAddr, err)
	}
	// 先取 top（用于返回 TopEventId/TopAlarmLevel），再 UPDATE
	res := &AutoResolveResult{}
	row := db.QueryRowContext(ctx, `
		SELECT event_id::text, alarm_level
		FROM alarm_events
		WHERE device_addr = $1::INET
		  AND event_type = ANY($2::text[])
		  AND alarm_status = 'active'
		ORDER BY alarm_level ASC, triggered_at DESC
		LIMIT 1
	`, addr.String(), pq.Array(alarmTypes))
	var topLvl int16
	if err := row.Scan(&res.TopEventId, &topLvl); err != nil && err != sql.ErrNoRows {
		return nil, nil, fmt.Errorf("AutoResolveDeviceAlarms top: %w", err)
	}
	res.TopAlarmLevel = intToAlarmLevelStr(topLvl)
	// 批量 UPDATE
	result, err := db.ExecContext(ctx, `
		UPDATE alarm_events
		SET alarm_status = 'auto_resolved',
		    hand_time    = NOW(),
		    operation    = 'auto_resolved'
		WHERE device_addr = $1::INET
		  AND event_type = ANY($2::text[])
		  AND alarm_status = 'active'
	`, addr.String(), pq.Array(alarmTypes))
	if err != nil {
		return nil, nil, fmt.Errorf("AutoResolveDeviceAlarms update: %w", err)
	}
	if n, err := result.RowsAffected(); err == nil {
		res.ResolvedCount = int(n)
	}
	cas, qErr := QueryCardAlarmState(ctx, db, cardID)
	if qErr != nil {
		return nil, nil, qErr
	}
	return cas, res, nil
}

// ExpireAlarmsByDeviceAddrs 设备删除/解绑时把该 device 下所有 active alarm 标 expired。
// deviceAddrs 元素是 IPv6 /128 canonical host text。
func ExpireAlarmsByDeviceAddrs(ctx context.Context, db *sql.DB, tenantID string, deviceAddrs []string, reason string) error {
	if len(deviceAddrs) == 0 {
		return nil
	}
	addrs := make([]string, 0, len(deviceAddrs))
	for _, id := range deviceAddrs {
		if a, err := netip.ParseAddr(id); err == nil {
			addrs = append(addrs, a.String())
		}
	}
	if len(addrs) == 0 {
		return nil
	}
	_, err := db.ExecContext(ctx, `
		UPDATE alarm_events
		SET alarm_status  = 'expired',
		    hand_time     = NOW(),
		    operation     = 'auto_resolved',
		    handler_notes = COALESCE(NULLIF(handler_notes, ''), $2)
		WHERE device_addr = ANY($1::INET[])
		  AND alarm_status = 'active'
	`, pq.Array(addrs), reason)
	if err != nil {
		return fmt.Errorf("expire alarms by device addrs: %w", err)
	}
	return nil
}

// ActiveAlarmRow alarm_events 单条 active 行
type ActiveAlarmRow struct {
	EventID     string
	EventType   string
	AlarmLevel  string
	DeviceAddr  string // IPv6 /128 canonical host text
	TriggeredAt int64  // ms
}

// AlarmWithCardInfo 含 card 归属的 alarm 行（API 兼容）
type AlarmWithCardInfo struct {
	CardID string // cards.spatial_prefix CIDR 字符串
	ActiveAlarmRow
}

// GetActiveAlarmsByCardID v2: device_addr LPM 落在 cardID 范围内 + alarm_status='active'。
func GetActiveAlarmsByCardID(ctx context.Context, db *sql.DB, cardID string) ([]ActiveAlarmRow, error) {
	if cardID == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT event_id::text, event_type, alarm_level,
		       host(device_addr),
		       (EXTRACT(EPOCH FROM triggered_at) * 1000)::bigint
		FROM alarm_events
		WHERE device_addr <<= $1::INET
		  AND alarm_status = 'active'
		ORDER BY triggered_at DESC
	`, cardID)
	if err != nil {
		return nil, fmt.Errorf("get active alarms for card %s: %w", cardID, err)
	}
	defer rows.Close()
	var out []ActiveAlarmRow
	for rows.Next() {
		var a ActiveAlarmRow
		var lvl int16
		if err := rows.Scan(&a.EventID, &a.EventType, &lvl, &a.DeviceAddr, &a.TriggeredAt); err != nil {
			return nil, fmt.Errorf("scan alarm row: %w", err)
		}
		a.AlarmLevel = intToAlarmLevelStr(lvl)
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListAllActiveAlarms v2: 跨 tenant 查所有 active alarms，按 cards LPM 关联 card_id。
func ListAllActiveAlarms(ctx context.Context, db *sql.DB, tenantID string) ([]AlarmWithCardInfo, error) {
	q := `
		SELECT
		  COALESCE((
		    SELECT host(c.spatial_prefix) || '/' || masklen(c.spatial_prefix)
		    FROM cards c
		    WHERE ae.device_addr <<= c.spatial_prefix AND c.is_active = true
		    ORDER BY masklen(c.spatial_prefix) DESC
		    LIMIT 1
		  ), '') AS card_id,
		  ae.event_id::text, ae.event_type, ae.alarm_level,
		  host(ae.device_addr),
		  (EXTRACT(EPOCH FROM ae.triggered_at) * 1000)::bigint
		FROM alarm_events ae
		WHERE ae.alarm_status = 'active'
	`
	var args []interface{}
	if tenantID != "" {
		q += ` AND ae.device_addr <<= $1::INET`
		args = append(args, tenantID)
	}
	q += ` ORDER BY ae.triggered_at DESC`

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list all active alarms: %w", err)
	}
	defer rows.Close()
	var out []AlarmWithCardInfo
	for rows.Next() {
		var a AlarmWithCardInfo
		var lvl int16
		if err := rows.Scan(&a.CardID, &a.EventID, &a.EventType, &lvl, &a.DeviceAddr, &a.TriggeredAt); err != nil {
			return nil, fmt.Errorf("scan alarm row: %w", err)
		}
		a.AlarmLevel = intToAlarmLevelStr(lvl)
		out = append(out, a)
	}
	return out, rows.Err()
}
