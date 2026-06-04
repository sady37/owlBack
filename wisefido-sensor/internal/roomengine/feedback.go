// Package roomengine
//
// feedback.go — 把 alarm_events 人类反馈喂回 cell history（PR-4 + PR-6）。
//
// 反馈源（cardagg 前端 admin 标记 → alarm_events.notes）：
//
//   operation='false_alarm' (PR-6 新格式):
//     False Alarm Reason:
//     ☑ NoPerson / Electric / AC Interference  → cell.GhostCount++
//     ☑ Sit on Chair / behind Chair             → cell.RestZoneConfirmed++
//     ☑ Lying/Sit in Sofa or Lounge chair       → cell.RestZoneConfirmed++
//     ☑ Sit in Wheelchair                        → cell.RestZoneConfirmed++ (mobile decay 自滤)
//     ☑ Other / 无勾选                            → cell.FakeAlarmCount++ (PR-4 兜底)
//
//   operation='verified' (Observed Conditions):
//     ☑ Fall / Sitting on the Ground             → cell.RealFallCount++
//     ☑ Awake/Verbal/Unresponsive/Bleeding       → 严重程度 metadata（不直接喂 cell）
//
// 流程：
//  1. 从 engine_alarm_feedback_cursor 读 last_hand_time
//  2. SELECT alarm_events WHERE operation IN (false_alarm, verified) AND event_type IN (Fall, ...)
//     AND hand_time > cursor ORDER BY hand_time
//  3. 每条：解析 notes 提取 ☑ checkbox → 路由到对应 cell counter
//  4. 90s lookback iot_timeseries 找触发位置 → RadarToCanvas → cell mark
//  5. UPDATE cursor

package roomengine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"owl-common/alarm"
	"owl-common/radarutils"
)

// AlarmFeedbackIngester 从 alarm_events 增量拉 false_alarm 反馈到 cell。
type AlarmFeedbackIngester struct {
	db     *sql.DB
	engine *Engine
	logger *zap.Logger

	// 90s 回看窗口，覆盖 silent fall 60s windows 略有余量
	lookbackMs int64

	// 仅这些 event_type 的 false_alarm 喂入（fall 类才与 cell 位置相关）
	eventTypes []string
}

// NewAlarmFeedbackIngester 默认 90s lookback，仅 Fall/SittingOnGround/Stay 类。
func NewAlarmFeedbackIngester(db *sql.DB, engine *Engine, logger *zap.Logger) *AlarmFeedbackIngester {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AlarmFeedbackIngester{
		db:         db,
		engine:     engine,
		logger:     logger,
		lookbackMs: 90_000,
		eventTypes: []string{alarm.Fall, alarm.SittingOnGround, alarm.Stay},
	}
}

// ParsedConditions 从 notes 解析的 checkbox 状态。
type ParsedConditions struct {
	// False Alarm Reason: 5 项
	FAGhostInterference bool // ☑ NoPerson / Electric / AC Interference
	FAChair             bool // ☑ Sit on Chair / behind Chair
	FASofa              bool // ☑ Lying/Sit in Sofa or Lounge chair
	FAWheelchair        bool // ☑ Sit in Wheelchair
	FAOther             bool // ☑ Other

	// Observed Conditions: 5 项 (verified)
	OCFall         bool // ☑ Fall / Sitting on the Ground
	OCAwake        bool // ☑ Awake and Responsive
	OCVerbal       bool // ☑ Responding Verbally
	OCUnresponsive bool // ☑ Unresponsive
	OCBleeding     bool // ☑ Visible Bleeding

	// True 表示 notes 里有 "False Alarm Reason:" 或 "Observed Conditions:" 块
	HasFalseAlarmBlock bool
	HasObservedBlock   bool
}

// AnyFASelected 是否有任一 false_alarm checkbox 勾选。
func (p ParsedConditions) AnyFASelected() bool {
	return p.FAGhostInterference || p.FAChair || p.FASofa || p.FAWheelchair || p.FAOther
}

// parseConditions 从 alarm_events.notes 提取 ☑ checkbox 状态。
// 用 prefix-based substring 匹配；标签文字与前端 cardagg UI 严格对齐。
func parseConditions(notes string) ParsedConditions {
	out := ParsedConditions{}
	if notes == "" {
		return out
	}
	out.HasFalseAlarmBlock = strings.Contains(notes, "False Alarm Reason:")
	out.HasObservedBlock = strings.Contains(notes, "Observed Conditions:")

	// 用 prefix-based 检查；同标签的别名 / 标点变体在前端可能不一致，列若干兼容写法
	hasMark := func(label string) bool {
		return strings.Contains(notes, "☑ "+label)
	}

	// ----- False Alarm Reason 5 项 -----
	if hasMark("NoPerson / Electric / AC Interference") ||
		hasMark("No Person / Electric / AC Interference") ||
		hasMark("NoPerson") ||
		hasMark("No Person") {
		out.FAGhostInterference = true
	}
	if hasMark("Sit on Chair / behind Chair") || hasMark("Sit on Chair") {
		out.FAChair = true
	}
	if hasMark("Lying/Sit in Sofa or Lounge chair") || hasMark("Lying/Sit in Sofa") {
		out.FASofa = true
	}
	if hasMark("Sit in Wheelchair") {
		out.FAWheelchair = true
	}
	// Other 在 FA 段和 Observed 段都没有别的同名项；只在 FA 段有 "Other" 选项 → 直接 prefix 匹配
	if out.HasFalseAlarmBlock && hasMark("Other") {
		out.FAOther = true
	}

	// ----- Observed Conditions 5 项 -----
	if hasMark("Fall / Sitting on the Ground") {
		out.OCFall = true
	}
	if hasMark("Awake and Responsive") {
		out.OCAwake = true
	}
	if hasMark("Responding Verbally") {
		out.OCVerbal = true
	}
	if hasMark("Unresponsive") {
		out.OCUnresponsive = true
	}
	if hasMark("Visible Bleeding") {
		out.OCBleeding = true
	}
	return out
}

// IngestOnce 跑一次增量同步。
// 返回 (处理的反馈条数, 错误)。错误时游标不更新；下次会重试同一批。
func (i *AlarmFeedbackIngester) IngestOnce(ctx context.Context) (int, error) {
	if i.db == nil || i.engine == nil {
		return 0, fmt.Errorf("ingester not configured (nil db or engine)")
	}

	// 1. 读游标 (v2 schema: cursor_key VARCHAR PK + last_processed_at；id/processed_total 退役)
	var cursorTime time.Time
	err := i.db.QueryRowContext(ctx,
		`SELECT last_processed_at FROM engine_alarm_feedback_cursor WHERE cursor_key = 'engine_alarm_feedback'`,
	).Scan(&cursorTime)
	if err != nil {
		if err == sql.ErrNoRows {
			// 还没初始化，按 epoch 当成首次
			cursorTime = time.Unix(0, 0)
		} else {
			return 0, fmt.Errorf("read cursor: %w", err)
		}
	}

	// 2. 拉新事件 (v2 alarm_events: trigger_data → payload；notes → handler_notes)
	placeholders := make([]string, len(i.eventTypes))
	args := make([]interface{}, 0, len(i.eventTypes)+1)
	args = append(args, cursorTime)
	for idx, et := range i.eventTypes {
		placeholders[idx] = fmt.Sprintf("$%d", idx+2)
		args = append(args, et)
	}
	q := fmt.Sprintf(`
		SELECT event_id::text, host(device_addr)::text, event_type, operation, triggered_at, hand_time,
		       payload, COALESCE(handler_notes, '')
		FROM alarm_events
		WHERE operation IN ('false_alarm', 'verified')
		  AND event_type IN (%s)
		  AND hand_time IS NOT NULL
		  AND hand_time > $1
		ORDER BY hand_time
		LIMIT 500
	`, strings.Join(placeholders, ","))

	rows, err := i.db.QueryContext(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("scan alarm_events: %w", err)
	}
	defer rows.Close()

	processed := 0
	var maxHandTime time.Time = cursorTime
	var lastEventID string

	for rows.Next() {
		var eventID, deviceAddr, eventType, operation, notes string
		var triggeredAt, handTime time.Time
		var triggerData []byte
		if err := rows.Scan(&eventID, &deviceAddr, &eventType, &operation, &triggeredAt, &handTime, &triggerData, &notes); err != nil {
			i.logger.Warn("alarm_feedback: scan row", zap.Error(err))
			continue
		}

		// 反查位置 → 按 conditions 分流到不同 cell counter
		ok := i.processOne(ctx, eventID, deviceAddr, eventType, operation, triggeredAt, handTime, triggerData, notes)
		if ok {
			processed++
		}
		// 不论成功与否，都向前推游标（避免错误事件反复重试）
		if handTime.After(maxHandTime) {
			maxHandTime = handTime
			lastEventID = eventID
		}
	}
	if err := rows.Err(); err != nil {
		return processed, fmt.Errorf("iterate alarm_events: %w", err)
	}

	// 3. 推进游标（v2: UPSERT cursor_key 主键；processed_total 退役，notes 字段记最近一批 N）
	if maxHandTime.After(cursorTime) {
		_, err := i.db.ExecContext(ctx, `
			INSERT INTO engine_alarm_feedback_cursor (cursor_key, last_processed_id, last_processed_at, notes)
			VALUES ('engine_alarm_feedback', $1::uuid, $2, $3)
			ON CONFLICT (cursor_key) DO UPDATE
			SET last_processed_id = EXCLUDED.last_processed_id,
			    last_processed_at = EXCLUDED.last_processed_at,
			    notes             = EXCLUDED.notes
		`, lastEventID, maxHandTime, fmt.Sprintf("processed=%d", processed))
		if err != nil {
			return processed, fmt.Errorf("update cursor: %w", err)
		}
	}
	return processed, nil
}

// processOne 处理单条 alarm（false_alarm 或 verified）：
// 1) 解析 notes 提取 ☑ checkbox
// 2) 反查 90s 内位置 → canvas
// 3) 按 conditions 分流到不同 cell counter
//
// 返回是否成功（false 时记 debug 日志，不阻塞批次）。
func (i *AlarmFeedbackIngester) processOne(ctx context.Context,
	eventID, deviceAddr, eventType, operation string,
	triggeredAt, handTime time.Time,
	triggerData []byte, notes string) bool {

	// 1. 路由 device → room
	roomID := i.engine.RoomForDevice(deviceAddr)
	if roomID == "" {
		i.logger.Debug("alarm_feedback: device not routed",
			zap.String("event_id", eventID), zap.String("device_addr", deviceAddr))
		return false
	}
	mount, ok := i.engine.MountForDevice(deviceAddr)
	if !ok {
		i.logger.Debug("alarm_feedback: device not mounted",
			zap.String("event_id", eventID), zap.String("device_addr", deviceAddr))
		return false
	}

	// 2. 抽 trigger_data 里的 track_id（可选过滤）
	wantTrackID, hasWant := extractTriggerTrackID(triggerData)

	// 3. 90s lookback iot_timeseries 找最近的 monitor 帧
	triggerMs := triggeredAt.UnixMilli()
	loX, loY, loZ, foundLocal := i.findRadarPositionAt(ctx, deviceAddr, triggerMs, wantTrackID, hasWant)
	if !foundLocal {
		i.logger.Debug("alarm_feedback: no matching monitor frame in lookback",
			zap.String("event_id", eventID),
			zap.String("device_addr", deviceAddr),
			zap.Int64("trigger_ms", triggerMs))
		return false
	}

	// 4. RadarToCanvas
	canvas := radarutils.RadarToCanvas(radarutils.RadarPoint{H: loX, V: loY, Z: loZ}, mount)

	// 5. 解析 notes conditions
	pc := parseConditions(notes)

	// 6. 按 conditions 路由到 cell counter
	handMs := handTime.UnixMilli()
	routes := i.routeFeedback(roomID, canvas.X, canvas.Y, handMs, operation, pc)
	if len(routes) == 0 {
		i.logger.Debug("alarm_feedback: no route applied (cell out of grid?)",
			zap.String("event_id", eventID),
			zap.String("room_id", roomID),
			zap.Int("canvas_x", canvas.X), zap.Int("canvas_y", canvas.Y))
		return false
	}

	i.logger.Info("alarm_feedback_marked",
		zap.String("event_id", eventID),
		zap.String("device_addr", deviceAddr),
		zap.String("event_type", eventType),
		zap.String("operation", operation),
		zap.String("room_id", roomID),
		zap.Time("triggered_at", triggeredAt),
		zap.Time("hand_time", handTime),
		zap.Strings("routes", routes),
		zap.Int("canvas_x", canvas.X), zap.Int("canvas_y", canvas.Y),
		zap.Int("local_h", loX), zap.Int("local_v", loY), zap.Int("local_z", loZ),
	)
	return true
}

// routeFeedback 根据 operation + parsed conditions 把 cell counter 增量。
// 返回触发的 route 名称列表（用于日志可观测）。
//
// 路由规则：
//   operation=false_alarm:
//     ☑ NoPerson/Electric/AC Interference → cell.GhostCount++       route="ghost"
//     ☑ Sit on Chair / Sofa / Wheelchair  → cell.RestZoneConfirmed++ route="rest_zone"
//     ☑ Other 或无任何勾选                  → cell.FakeAlarmCount++   route="fake_generic"
//   operation=verified:
//     ☑ Fall / Sitting on the Ground       → cell.RealFallCount++    route="real_fall"
//     仅 5 项里勾了非 Fall 的（如 Bleeding 单独）→ 也算 real fall（毕竟 verified） route="real_fall_implicit"
//     完全无勾选 → 不进 cell（admin 没填 condition，作为 noise 跳过）
func (i *AlarmFeedbackIngester) routeFeedback(roomID string, x, y int, nowMs int64,
	operation string, pc ParsedConditions) []string {

	var routes []string

	// helper: 调 engine 改对应 cell 字段；返回是否成功（cell 在 grid 内）
	apply := func(fn func(*Cell)) bool {
		return i.engine.ApplyToCell(roomID, x, y, nowMs, fn)
	}

	switch operation {
	case "false_alarm":
		// 多选：每个勾选独立累计（同 cell 多个 counter 同时 +1）
		if pc.FAGhostInterference {
			if apply(func(c *Cell) { c.IncrGhostCount() }) {
				routes = append(routes, "ghost")
			}
		}
		// PR-9.2: 拆分 target — Chair/Wheelchair → AreaSit；Sofa → AreaBed (lying)
		if pc.FAChair || pc.FAWheelchair {
			locked := false
			if apply(func(c *Cell) {
				c.IncrRestZoneConfirmed()
				if c.MarkRestZoneByFeedback(AreaSit) {
					locked = true
				}
			}) {
				routes = append(routes, "rest_zone_chair")
				if locked {
					routes = append(routes, "rest_zone_locked_AreaSit")
				}
			}
		}
		if pc.FASofa {
			locked := false
			if apply(func(c *Cell) {
				c.IncrRestZoneConfirmed()
				if c.MarkRestZoneByFeedback(AreaBed) {
					locked = true
				}
			}) {
				routes = append(routes, "rest_zone_sofa")
				if locked {
					routes = append(routes, "rest_zone_locked_AreaBed")
				}
			}
		}
		// Other 或全无勾选 → 通用 fake alarm（PR-4 兜底）
		if pc.FAOther || !pc.AnyFASelected() {
			if apply(func(c *Cell) { c.IncrFakeAlarm() }) {
				routes = append(routes, "fake_generic")
			}
		}
	case "verified":
		// verified + ☑ Fall 是 ground truth；其它单 condition 也视为真 fall（admin 标了 verified 就是认 fall）
		if pc.OCFall || pc.OCAwake || pc.OCVerbal || pc.OCUnresponsive || pc.OCBleeding {
			cleared := false
			if apply(func(c *Cell) {
				c.IncrRealFallCount()
				// 真摔擦该处非 FE 画的 rest/deny 分类（feedback/自学的抑制被真摔证伪）；FE 画 bed 不擦（仅记录+人工复审）。
				if c.ClearNonHumanLearnedZone() {
					cleared = true
				}
			}) {
				if pc.OCFall {
					routes = append(routes, "real_fall")
				} else {
					routes = append(routes, "real_fall_implicit")
				}
				if cleared {
					routes = append(routes, "cleared_non_human_zone")
				}
			}
		}
		// 全无勾选 → 不进 cell（admin 没具体描述，跳过，避免噪声）
	}

	return routes
}

// findRadarPositionAt 在 [triggerMs - lookbackMs, triggerMs + 1000] 区间查 monitor 帧。
// 优先匹配指定 track_id；若无则取最接近 triggerMs 的任一 track 帧。
// 返回雷达本地坐标 (h, v, z) 与是否找到。
//
// v2 schema: iot_timeseries 退役 → monitor_stream (PK ts TIMESTAMPTZ, device_addr INET);
// stream_type 替代 topic_type+category 二元组合 ('radar.track')；payload 替代 data_value。
func (i *AlarmFeedbackIngester) findRadarPositionAt(ctx context.Context,
	deviceAddr string, triggerMs int64, wantTrackID int, hasWant bool) (int, int, int, bool) {

	q := `
		SELECT (extract(epoch from ts) * 1000)::bigint AS ts_ms, payload
		FROM monitor_stream
		WHERE device_addr = $1::INET
		  AND stream_type = 'radar.track'
		  AND ts BETWEEN to_timestamp($2 / 1000.0) AND to_timestamp($3 / 1000.0)
		ORDER BY ts DESC
		LIMIT 50
	`
	rows, err := i.db.QueryContext(ctx, q,
		deviceAddr, triggerMs-i.lookbackMs, triggerMs+1000)
	if err != nil {
		i.logger.Warn("alarm_feedback: lookup iot_timeseries", zap.Error(err))
		return 0, 0, 0, false
	}
	defer rows.Close()

	type cand struct {
		tms     int64
		h, v, z int
		matched bool // 与 wantTrackID 匹配
	}
	var best *cand

	for rows.Next() {
		var tms int64
		var dv []byte
		if err := rows.Scan(&tms, &dv); err != nil {
			continue
		}
		var arr []map[string]interface{}
		if err := json.Unmarshal(dv, &arr); err != nil {
			continue
		}
		for _, m := range arr {
			// Stage 1b：category='track' 已由 SQL 过滤；这里只解 payload 字段
			tid := jsonInt(m["track_id"])
			h := jsonInt(m["position_x"])
			v := jsonInt(m["position_y"])
			z := jsonInt(m["position_z"])
			// 全 0 帧（heartbeat）跳过
			if h == 0 && v == 0 && z == 0 {
				continue
			}
			matched := hasWant && tid == wantTrackID
			c := &cand{tms: tms, h: h, v: v, z: z, matched: matched}
			// 选优规则：优先 matched；其次最接近 triggerMs
			if best == nil {
				best = c
				continue
			}
			if matched && !best.matched {
				best = c
				continue
			}
			if matched == best.matched {
				if abs64(c.tms-triggerMs) < abs64(best.tms-triggerMs) {
					best = c
				}
			}
		}
	}
	if best == nil {
		return 0, 0, 0, false
	}
	return best.h, best.v, best.z, true
}

// extractTriggerTrackID 尝试从 trigger_data 里抽 event_payload.track_id。
// trigger_data 结构例：
//
//	{"event_name":"Fall", "event_payload": "{\"track_id\":0,...}"}
//
// event_payload 是字符串化 JSON。返回 (track_id, found)。
func extractTriggerTrackID(triggerData []byte) (int, bool) {
	if len(triggerData) == 0 {
		return 0, false
	}
	var top map[string]interface{}
	if err := json.Unmarshal(triggerData, &top); err != nil {
		return 0, false
	}
	payloadStr, _ := top["event_payload"].(string)
	if payloadStr == "" {
		// 可能直接是对象
		if pl, ok := top["event_payload"].(map[string]interface{}); ok {
			if v, ok := pl["track_id"]; ok {
				return jsonInt(v), true
			}
		}
		// 或顶层就有 track_id
		if v, ok := top["track_id"]; ok {
			return jsonInt(v), true
		}
		return 0, false
	}
	var pl map[string]interface{}
	if err := json.Unmarshal([]byte(payloadStr), &pl); err != nil {
		return 0, false
	}
	if v, ok := pl["track_id"]; ok {
		return jsonInt(v), true
	}
	return 0, false
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
