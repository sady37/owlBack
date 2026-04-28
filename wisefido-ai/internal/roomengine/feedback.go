// Package roomengine
//
// feedback.go — 把 alarm_events.operation='false_alarm' 反馈到 cell.FakeAlarmCount
// （cell history integral 第一类反馈：人工标 fake 学习）。
//
// 流程：
//  1. 从 engine_alarm_feedback_cursor 读 last_hand_time
//  2. SELECT alarm_events WHERE operation='false_alarm' AND event_type IN (Fall, SittingOnGround, Stay)
//     AND hand_time > last_hand_time ORDER BY hand_time
//  3. 每条反馈：
//     a. device_id → engine.RoomForDevice → room_id
//     b. room_id → engine.MountForRoom → radar mount
//     c. SELECT iot_timeseries WHERE device_id AND timestamp BETWEEN [trigger-90s, trigger+1s]
//        AND topic_type='monitor' ORDER BY timestamp DESC LIMIT 50
//     d. 取最接近 triggered_at 的 monitor 帧（优先匹配 trigger_data.event_payload.track_id）
//     e. RadarToCanvas(local h/v/z, mount) → canvas (x, y)
//     f. engine.MarkFakeAlarmAt(roomID, x, y, hand_time_ms)
//  4. UPDATE engine_alarm_feedback_cursor 到本批最大 hand_time
//
// 调用方式：engine.Run 启动 feedbackLoop（默认 5min 周期）。

package roomengine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
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
		eventTypes: []string{"Fall", "SittingOnGround", "Stay"},
	}
}

// IngestOnce 跑一次增量同步。
// 返回 (处理的反馈条数, 错误)。错误时游标不更新；下次会重试同一批。
func (i *AlarmFeedbackIngester) IngestOnce(ctx context.Context) (int, error) {
	if i.db == nil || i.engine == nil {
		return 0, fmt.Errorf("ingester not configured (nil db or engine)")
	}

	// 1. 读游标
	var cursorTime time.Time
	err := i.db.QueryRowContext(ctx,
		`SELECT last_hand_time FROM engine_alarm_feedback_cursor WHERE id = 1`,
	).Scan(&cursorTime)
	if err != nil {
		if err == sql.ErrNoRows {
			// 还没初始化，按 epoch 当成首次
			cursorTime = time.Unix(0, 0)
		} else {
			return 0, fmt.Errorf("read cursor: %w", err)
		}
	}

	// 2. 拉新事件
	placeholders := make([]string, len(i.eventTypes))
	args := make([]interface{}, 0, len(i.eventTypes)+1)
	args = append(args, cursorTime)
	for idx, et := range i.eventTypes {
		placeholders[idx] = fmt.Sprintf("$%d", idx+2)
		args = append(args, et)
	}
	q := fmt.Sprintf(`
		SELECT event_id::text, device_id::text, event_type, triggered_at, hand_time, trigger_data
		FROM alarm_events
		WHERE operation = 'false_alarm'
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
		var eventID, deviceID, eventType string
		var triggeredAt, handTime time.Time
		var triggerData []byte
		if err := rows.Scan(&eventID, &deviceID, &eventType, &triggeredAt, &handTime, &triggerData); err != nil {
			i.logger.Warn("alarm_feedback: scan row", zap.Error(err))
			continue
		}

		// 反查位置 → 落点 cell
		ok := i.processOne(ctx, eventID, deviceID, eventType, triggeredAt, handTime, triggerData)
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

	// 3. 推进游标（即使本批为 0，maxHandTime == cursorTime 时也无副作用）
	if maxHandTime.After(cursorTime) {
		_, err := i.db.ExecContext(ctx, `
			UPDATE engine_alarm_feedback_cursor
			SET last_hand_time  = $1,
			    last_event_id   = $2::uuid,
			    processed_total = processed_total + $3,
			    updated_at      = NOW()
			WHERE id = 1
		`, maxHandTime, lastEventID, processed)
		if err != nil {
			return processed, fmt.Errorf("update cursor: %w", err)
		}
	}
	return processed, nil
}

// processOne 处理单条 false_alarm：查位置 → 标 cell。
// 返回是否成功（false 时记 warn 日志，不阻塞批次）。
func (i *AlarmFeedbackIngester) processOne(ctx context.Context,
	eventID, deviceID, eventType string,
	triggeredAt, handTime time.Time,
	triggerData []byte) bool {

	// 1. 路由 device → room
	roomID := i.engine.RoomForDevice(deviceID)
	if roomID == "" {
		i.logger.Debug("alarm_feedback: device not routed",
			zap.String("event_id", eventID), zap.String("device_id", deviceID))
		return false
	}
	mount, ok := i.engine.MountForRoom(roomID)
	if !ok {
		i.logger.Debug("alarm_feedback: room not mounted",
			zap.String("event_id", eventID), zap.String("room_id", roomID))
		return false
	}

	// 2. 抽 trigger_data 里的 track_id（可选过滤）
	wantTrackID, hasWant := extractTriggerTrackID(triggerData)

	// 3. 90s lookback iot_timeseries 找最近的 monitor 帧
	triggerMs := triggeredAt.UnixMilli()
	loX, loY, loZ, foundLocal := i.findRadarPositionAt(ctx, deviceID, triggerMs, wantTrackID, hasWant)
	if !foundLocal {
		i.logger.Debug("alarm_feedback: no matching monitor frame in lookback",
			zap.String("event_id", eventID),
			zap.String("device_id", deviceID),
			zap.Int64("trigger_ms", triggerMs))
		return false
	}

	// 4. RadarToCanvas
	canvas := radarutils.RadarToCanvas(radarutils.RadarPoint{H: loX, V: loY, Z: loZ}, mount)

	// 5. 标 cell
	if !i.engine.MarkFakeAlarmAt(roomID, canvas.X, canvas.Y, handTime.UnixMilli()) {
		i.logger.Debug("alarm_feedback: cell out of grid",
			zap.String("event_id", eventID),
			zap.String("room_id", roomID),
			zap.Int("canvas_x", canvas.X), zap.Int("canvas_y", canvas.Y))
		return false
	}

	i.logger.Info("alarm_feedback_marked",
		zap.String("event_id", eventID),
		zap.String("device_id", deviceID),
		zap.String("event_type", eventType),
		zap.String("room_id", roomID),
		zap.Time("triggered_at", triggeredAt),
		zap.Time("hand_time", handTime),
		zap.Int("canvas_x", canvas.X), zap.Int("canvas_y", canvas.Y),
		zap.Int("local_h", loX), zap.Int("local_v", loY), zap.Int("local_z", loZ),
	)
	return true
}

// findRadarPositionAt 在 [triggerMs - lookbackMs, triggerMs + 1000] 区间查 monitor 帧。
// 优先匹配指定 track_id；若无则取最接近 triggerMs 的任一 track 帧。
// 返回雷达本地坐标 (h, v, z) 与是否找到。
func (i *AlarmFeedbackIngester) findRadarPositionAt(ctx context.Context,
	deviceID string, triggerMs int64, wantTrackID int, hasWant bool) (int, int, int, bool) {

	q := `
		SELECT timestamp, data_value
		FROM iot_timeseries
		WHERE device_id = $1::uuid
		  AND topic_type = 'monitor'
		  AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
		LIMIT 50
	`
	rows, err := i.db.QueryContext(ctx, q,
		deviceID, triggerMs-i.lookbackMs, triggerMs+1000)
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
			eventName, _ := m["event_name"].(string)
			if eventName != "track" {
				continue
			}
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
