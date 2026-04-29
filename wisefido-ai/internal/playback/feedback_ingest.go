// Package playback
//
// feedback_ingest.go (PR-9.1) — playback 离线场景：把数据库里历史 alarm_events 反馈
// 灌进 grid.cells。让 playback HTML 的 SVG overlay 可以展示真实的人类反馈层
// （GhostCount / RestZoneConfirmed / RealFallCount / FakeAlarmCount）。
//
// 运行时（cmd/wisefido-ai daemon）有 PR-4 持续 ingester；playback 是离线快照工具，
// 不持有 redis client / 不调 engine.PublishAIEvent，所以单独走一条灌库逻辑。
//
// 与 internal/roomengine/feedback.go 的 IngestOnce 同样：
//   1. 拉指定 device 的 alarm_events 历史（false_alarm + verified）
//   2. parseConditions(notes) 解析 ☑ checkbox
//   3. 90s lookback iot_timeseries 拿 monitor 帧（雷达本地 h/v/z）
//   4. RadarToCanvas → 落到 cell
//   5. 按 conditions 调 grid.Mark{Ghost,RestZone,RealFall,FakeAlarm}Feedback

package playback

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"owl-common/radarutils"
	"wisefido-ai/internal/roomengine"
)

// IngestHistoricalFeedback 把指定 device_uid 在 [start, end] 内的 alarm_events 反馈
// 解析后落到 grid。返回 (success_count, total_attempted, error)。
func IngestHistoricalFeedback(ctx context.Context, db *sql.DB, deviceUID string,
	start, end time.Time, grid *roomengine.RoomGrid, mount radarutils.RadarMount,
	logger *zap.Logger) (int, int, error) {

	if logger == nil {
		logger = zap.NewNop()
	}

	q := `
		SELECT
			ae.event_id::text, ae.device_id::text, ae.event_type, ae.operation,
			extract(epoch from ae.triggered_at)*1000 AS trigger_ms,
			extract(epoch from COALESCE(ae.hand_time, ae.triggered_at))*1000 AS hand_ms,
			COALESCE(ae.notes, '') AS notes,
			COALESCE(ae.trigger_data->>'event_payload', '') AS payload
		FROM alarm_events ae
		JOIN devices d ON d.device_id = ae.device_id
		WHERE d.device_uid = $1
		  AND ae.event_type IN ('Fall', 'SittingOnGround', 'Stay')
		  AND ae.operation IN ('false_alarm', 'verified')
		  AND ae.triggered_at BETWEEN to_timestamp($2/1000.0) AND to_timestamp($3/1000.0)
		ORDER BY ae.triggered_at
	`
	rows, err := db.QueryContext(ctx, q, deviceUID, start.UnixMilli(), end.UnixMilli())
	if err != nil {
		return 0, 0, fmt.Errorf("query alarm_events: %w", err)
	}
	defer rows.Close()

	success := 0
	total := 0
	for rows.Next() {
		total++
		var eventID, deviceID, eventType, operation, notes, payload string
		var triggerMs, handMs float64
		if err := rows.Scan(&eventID, &deviceID, &eventType, &operation,
			&triggerMs, &handMs, &notes, &payload); err != nil {
			continue
		}

		// 1. 解析 conditions
		pc := parseConditionsLite(notes, operation)

		// 2. 90s lookback 找 trigger 位置（radar local h/v/z）
		h, v, z, ok := lookbackPositionFromIoT(ctx, db, deviceID, int64(triggerMs), payload)
		if !ok {
			continue
		}

		// 3. RadarToCanvas
		canvas := radarutils.RadarToCanvas(radarutils.RadarPoint{H: h, V: v, Z: z}, mount)

		// 4. 路由到 grid 反馈
		applied := false
		switch operation {
		case "false_alarm":
			if pc.FAGhostInterference {
				grid.MarkGhostFeedback(canvas.X, canvas.Y, int64(handMs))
				applied = true
			}
			// PR-9.2: 拆分 target — Chair/Wheelchair → AreaSit；Sofa → AreaBed (lying)
			if pc.FAChair || pc.FAWheelchair {
				grid.MarkRestZoneFeedback(canvas.X, canvas.Y, roomengine.AreaSit, int64(handMs))
				applied = true
			}
			if pc.FASofa {
				grid.MarkRestZoneFeedback(canvas.X, canvas.Y, roomengine.AreaBed, int64(handMs))
				applied = true
			}
			if pc.FAOther || (!pc.FAGhostInterference && !pc.FAChair && !pc.FASofa && !pc.FAWheelchair) {
				grid.MarkFakeAlarm(canvas.X, canvas.Y, int64(handMs))
				applied = true
			}
		case "verified":
			if pc.OCFall || pc.OCAwake || pc.OCVerbal || pc.OCUnresponsive || pc.OCBleeding {
				grid.MarkRealFallFeedback(canvas.X, canvas.Y, int64(handMs))
				applied = true
			}
		}
		if applied {
			success++
		}
	}
	return success, total, rows.Err()
}

// parsedConditions 与 roomengine.ParsedConditions 同义副本（避免 import 循环）
type parsedConditions struct {
	FAGhostInterference bool
	FAChair             bool
	FASofa              bool
	FAWheelchair        bool
	FAOther             bool
	OCFall              bool
	OCAwake             bool
	OCVerbal            bool
	OCUnresponsive      bool
	OCBleeding          bool
}

func parseConditionsLite(notes, operation string) parsedConditions {
	out := parsedConditions{}
	if notes == "" {
		return out
	}
	hasMark := func(label string) bool { return strings.Contains(notes, "☑ "+label) }
	inFA := strings.Contains(notes, "False Alarm Reason:")
	if hasMark("NoPerson / Electric / AC Interference") || hasMark("No Person / Electric / AC Interference") ||
		hasMark("NoPerson") || hasMark("No Person") {
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
	if inFA && hasMark("Other") {
		out.FAOther = true
	}
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

// lookbackPositionFromIoT 90s lookback iot_timeseries 找 trigger 时刻最接近的 monitor 帧。
// 优先匹配 trigger_data 里的 track_id；找不到则取最接近 triggerMs 的任意 track 帧。
func lookbackPositionFromIoT(ctx context.Context, db *sql.DB, deviceID string,
	triggerMs int64, payload string) (h, v, z int, ok bool) {

	// 抽 track_id（可选过滤）
	wantTID := -1
	if payload != "" {
		var p map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &p); err == nil {
			if v, ok := p["track_id"].(float64); ok {
				wantTID = int(v)
			}
		}
	}

	q := `
		SELECT timestamp, data_value
		FROM iot_timeseries
		WHERE device_id = $1::uuid AND topic_type = 'monitor'
		  AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC LIMIT 50
	`
	rows, err := db.QueryContext(ctx, q, deviceID, triggerMs-90_000, triggerMs+1_000)
	if err != nil {
		return
	}
	defer rows.Close()

	type cand struct {
		tms     int64
		h, v, z int
		matched bool
	}
	var best *cand
	abs := func(x int64) int64 {
		if x < 0 {
			return -x
		}
		return x
	}
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
			ev, _ := m["event_name"].(string)
			if ev != "track" {
				continue
			}
			tid, _ := m["track_id"].(float64)
			hv, _ := m["position_x"].(float64)
			vv, _ := m["position_y"].(float64)
			zv, _ := m["position_z"].(float64)
			if hv == 0 && vv == 0 && zv == 0 {
				continue
			}
			matched := wantTID >= 0 && int(tid) == wantTID
			c := &cand{tms: tms, h: int(hv), v: int(vv), z: int(zv), matched: matched}
			if best == nil {
				best = c
				continue
			}
			if matched && !best.matched {
				best = c
				continue
			}
			if matched == best.matched && abs(c.tms-triggerMs) < abs(best.tms-triggerMs) {
				best = c
			}
		}
	}
	if best == nil {
		return
	}
	return best.h, best.v, best.z, true
}
