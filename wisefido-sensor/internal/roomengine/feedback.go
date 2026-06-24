// Package roomengine
//
// feedback.go — 把 alarm_events 人类反馈喂回 cell history（PR-4 + PR-6）。
//
// 反馈源（cardagg 前端 admin 标记 → alarm_events.notes）：
//
//   operation='false_alarm' (False Alarm Reason；label 镜像 owlFront/src/utils/alarm.ts):
//     ☑ Sit on Chair / Short Sofa · Sit in Wheelchair
//                                          → RestZoneConfirmed++ + MarkRestZoneByFeedback(AreaSit)
//         ↳ Sit zone pin                   → 追加 Chair object（reload→AreaSit 神圣不衰减）
//     ☑ Lying Lounge Chair / Long Sofa     → ↳ Lounge placement（与 Sit pin 一致，只永久无 2H）:
//         "Permanent (update layout)"      → pinFeedbackZone：追加 LongSofa object + StampPriorRect(AreaLying,SourceHuman)
//         (无 ↳ 行)                         → 不动 layout
//     ☑ BlindArea / No Fall (家具后遮挡盲区) ↳ permanent → 追加 Furniture object（reload→AreaDeny；保留 5min tFloor 兜底真摔）
//     ☑ Enter / Door (未画出口)             ↳ 直接 → 追加 Enter object（reload→AreaEnter）
//     ☑ Electric / AC Interference         → GhostCount++（ghost 学习，非 lying）
//     ☑ Error Pose Detection / Out of Detection Range → 不进任何 counter（传感误差，非空间属性）
//     ☑ Unknown / 无勾选                    → FakeAlarmCount++（兜底容忍）
//
//   operation='verified' (人确认真摔):       → RealFallCount++ + ClearNonHumanLearnedZone（擦非 Human 抑制/deny）
//     ↳ Sticky veto: ...                   → 额外 MarkLearnBlocked（永久禁该 cell 自动/反馈学抑制）
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

// False Alarm Reason / Lounge placement / Sticky veto 标记字符串——契约真相源 =
// owlFront/src/utils/alarm.ts（FALL_RADAR_REASONS / LOUNGE_PLACEMENT_MARKER / STICKY_VETO_MARKER）。
// FE 把勾选写成 remarks 里的 "☑ <label>" 行 + "↳ <marker>" 子行，这里按 substring 反解。改这些常量必须与 alarm.ts 同步。
const (
	faSitChairShortSofa = "Sit on Chair / Short Sofa"
	faLoungeLongSofa    = "Lying Lounge Chair / Long Sofa"
	faBlindAreaNoFall   = "BlindArea / No Fall"
	faEnterDoor         = "Enter / Door"
	faWheelchair        = "Sit in Wheelchair"
	faMetalMirror       = "Metal / Mirror (reflection)"
	faCurtainFanPlants  = "Curtain / Fan / Plants"
	faPet               = "Pet"
	faErrorPose         = "Error Pose Detection"
	faUnknown           = "Unknown"

	loungePlacementPermanent = "Lounge placement: Permanent (update layout)"
	stickyVetoMark           = "Sticky veto: never auto-learn fall suppression here"

	// 永久加区 marker（owlFront alarm.ts SIT_PIN_MARKER / DENY_ZONE_MARKER）。
	sitPinMark    = "Sit zone: pin permanently (update layout)"
	denyZoneMark  = "Interference zone: mark permanently (deny, update layout)"
	blindAreaMark = "Blind area: mark permanently (no-fall zone, update layout)"
	enterDoorMark = "Exit zone: add permanently (update layout)"
)

// ParsedConditions 从 notes 解析的勾选/标记状态（label 见上方常量）。
type ParsedConditions struct {
	// 坐类（→ AreaSit 学习）
	FASitChairShortSofa bool
	FAWheelchair        bool
	// 盲区/出口（左列良性，→ AreaDeny / AreaEnter，不进 sit 学习）
	FABlindAreaNoFall bool
	FAEnterDoor       bool
	// 躺类（→ 二次问询 Lounge placement）
	FALoungeLongSofa bool
	// 伪迹组（无真人，不进 sit/lying 学习）
	FAMetalMirror      bool // 静止反射 → ghost
	FACurtainFanPlants bool // 运动杂波（死物）→ fake
	FAPet              bool // 活体目标 → fake
	FAErrorPose        bool // 姿态误判 → 不进 counter
	FAUnknown          bool // 兜底 → fake

	// Lounge placement（仅 FALoungeLongSofa 勾选时有意义；与 Sit pin 一致只永久，无 2H）
	LoungePermanent bool // ↳ permanent → pinFeedbackZone(AreaLying)：追加 LongSofa object + StampPriorRect

	// 永久加区二次动作（追加 layout 对象）
	SitPin        bool // ↳ sit zone pin → 追加 Chair object（reload→AreaSit 神圣）
	DenyZone      bool // ↳ interference zone → 追加 Interfere object（reload→AreaDeny）
	BlindArea     bool // ↳ blind area permanent → 追加 Furniture object（reload→AreaDeny）
	EnterDoorZone bool // ↳ enter/door（勾 reason 即永久）→ 追加 Enter object（reload→AreaEnter）

	// verified
	StickyVeto bool // ↳ sticky veto → MarkLearnBlocked

	// True 表示 notes 里有 "False Alarm Reason:" 或 "Observed Conditions:" 块
	HasFalseAlarmBlock bool
	HasObservedBlock   bool
}

// anyFASelected 是否有任一 false_alarm 勾选。
func (p ParsedConditions) anyFASelected() bool {
	return p.FASitChairShortSofa || p.FAWheelchair || p.FABlindAreaNoFall || p.FAEnterDoor ||
		p.FALoungeLongSofa || p.FAMetalMirror || p.FACurtainFanPlants || p.FAPet ||
		p.FAErrorPose || p.FAUnknown
}

// anySitClass 坐类（Chair/ShortSofa/Wheelchair → AreaSit）。Behind 已退役为 BlindArea(→AreaDeny)，不再进 sit。
func (p ParsedConditions) anySitClass() bool {
	return p.FASitChairShortSofa || p.FAWheelchair
}

// parseConditions 从 alarm_events.notes 提取勾选/标记状态；label 严格对齐 owlFront alarm.ts。
func parseConditions(notes string) ParsedConditions {
	out := ParsedConditions{}
	if notes == "" {
		return out
	}
	out.HasFalseAlarmBlock = strings.Contains(notes, "False Alarm Reason:")
	out.HasObservedBlock = strings.Contains(notes, "Observed Conditions:")

	hasMark := func(label string) bool { return strings.Contains(notes, "☑ "+label) }

	out.FASitChairShortSofa = hasMark(faSitChairShortSofa)
	out.FALoungeLongSofa = hasMark(faLoungeLongSofa)
	out.FABlindAreaNoFall = hasMark(faBlindAreaNoFall)
	out.FAEnterDoor = hasMark(faEnterDoor)
	out.FAWheelchair = hasMark(faWheelchair)
	out.FAMetalMirror = hasMark(faMetalMirror)
	out.FACurtainFanPlants = hasMark(faCurtainFanPlants)
	out.FAPet = hasMark(faPet)
	out.FAErrorPose = hasMark(faErrorPose)
	out.FAUnknown = hasMark(faUnknown)

	// 二次问询 / veto 是 "↳ <marker>" 子行（非 ☑），直接 substring
	out.LoungePermanent = strings.Contains(notes, loungePlacementPermanent)
	out.SitPin = strings.Contains(notes, sitPinMark)
	out.DenyZone = strings.Contains(notes, denyZoneMark)
	out.BlindArea = strings.Contains(notes, blindAreaMark)
	out.EnterDoorZone = strings.Contains(notes, enterDoorMark)
	out.StickyVeto = strings.Contains(notes, stickyVetoMark)

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

	// 永久加区：护士勾"钉为坐/躺/干扰/盲/门区" → 追加 source=Feedback layout 对象 + 即时 live 刷 grid。
	//   live-stamp 桥接 daily reload(22:00)/重启窗口：否则区要到下次 reload 才落 grid，期间 track 仍读 none→12min 误报。
	if operation == "false_alarm" {
		if pc.FALoungeLongSofa && pc.LoungePermanent {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackLoungeObject)
		}
		if pc.SitPin {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackSitObject)
		}
		if pc.DenyZone {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackInterfereObject)
		}
		if pc.BlindArea {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackBlindObject)
		}
		if pc.EnterDoorZone {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackEnterObject)
		}
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

// feedbackObjectSpec 描述一类"护士永久加区"追加的 layout 对象（typeName 决定 reload 后 engine 先验）。
type feedbackObjectSpec struct {
	idPrefix string // object id 前缀 + eventID（幂等键）
	name     string
	typeName string // 对齐 FE BaseObject + sensor layout_parser 的 case
	category string
	color    string
	height   int
	logKey   string
	// liveArea/liveConf = 该 typeName 经 layout_parser→RegisterRoom 的 reload 等价 (AreaType, conf)。
	// live-stamp 照此刷 grid 桥接 reload 窗口；必须与 layout_parser 映射逐项对齐（改 parser 须同步此处）。
	// Lounge 用 AreaLying（非 AreaBed）：floor 纯按 AreaType 豁免，bed 型会关掉 90min 兜底 → 漏报。
	liveArea AreaType
	liveConf int
}

// feedbackObjectHalfCm = 永久加区追加 object 的半边长（40×40cm 方块，告警点为心）；
// 同源供 live grid 即时刷 rect 复用（与 reload parser 读同一几何）。
const feedbackObjectHalfCm = 20

// reload 后 engine 先验（liveArea，与 layout_parser typeName 映射一致）：
//   LongSofa→AreaLying(90min,非床躺) / Chair→AreaSit / Interfere→AreaInterfer / Furniture→AreaDeny / Enter→AreaEnter。
// 均 source=Feedback：drawObjects 渲染虚线、updateRadarAreas 过滤不下发雷达（人确认转 Human 后才 declare）。
var (
	feedbackLoungeObject    = feedbackObjectSpec{"feedback_lounge_", "Lounge (learned)", "LongSofa", "furniture", "#c19a6b", 40, "feedback_lounge_object_appended", AreaLying, 99}
	feedbackSitObject       = feedbackObjectSpec{"feedback_sit_", "Sit zone (pinned)", "Chair", "furniture", "#ffff00", 90, "feedback_sit_object_appended", AreaSit, chairPriorConf}
	feedbackInterfereObject = feedbackObjectSpec{"feedback_deny_", "Interference (marked)", "Interfere", "interference", "#4a4a4a", 120, "feedback_interfere_object_appended", AreaInterfer, 99}
	feedbackBlindObject     = feedbackObjectSpec{"feedback_blind_", "Blind area (no-fall)", "Furniture", "furniture", "#d3d3d3", 80, "feedback_blind_object_appended", AreaDeny, 99}
	feedbackEnterObject     = feedbackObjectSpec{"feedback_enter_", "Exit (learned)", "Enter", "structure", "#a9eaa9", 0, "feedback_enter_object_appended", AreaEnter, 99}
)

// pinFeedbackZone 永久加区单一入口：append source=Feedback layout 对象 + 即时 live 刷同一 40×40 rect。
// live-stamp 用 spec.liveArea/liveConf(=reload 等价) + SourceHuman，与 daily reload 产出逐 cell 一致(零漂移)，
// 且 SourceHuman 扛 verified 擦。已知小边界：reload 有 Chairs/Furnitures/Interferes 覆盖次序，live 单对象刷不复制，
// 坐/躺点压在别类区上时暂以本类盖过，下次 reload 自愈(罕见)。
func (i *AlarmFeedbackIngester) pinFeedbackZone(ctx context.Context, deviceAddr, roomID string, x, y int, eventID string, spec feedbackObjectSpec) {
	i.appendFeedbackObject(ctx, deviceAddr, x, y, eventID, spec)
	h := feedbackObjectHalfCm
	rect := radarutils.Rect{X1: x - h, Y1: y - h, X2: x + h, Y2: y + h}
	if i.engine.StampPriorRect(roomID, rect, spec.liveArea, spec.liveConf) {
		i.logger.Info("feedback_zone_live_stamped",
			zap.String("event_id", eventID), zap.String("room_id", roomID),
			zap.String("type", spec.typeName), zap.Int("area_type", int(spec.liveArea)),
			zap.Int("canvas_x", x), zap.Int("canvas_y", y))
	}
}

// appendFeedbackObject 把"永久加区"动作作为 source='Feedback' object 并发安全 append 进
// 该 /128 device 的 room_visual_layout.canvas.objects[]（[[layout_authority_ai_correction_model]]）。
//   - 并发安全：DB 端 jsonb `||` 原子追加，不在 app 层读-改-写（FE 同时存 Human object 不冲突）。
//   - 幂等：同 event 的 object id 已存在则不重复 append。
//   - canvas_hash 置 NULL：让下次 FE save 必写；engine snapshot 用自身 geometry hash，不受影响。
//   - 几何 = 告警点周边 40×40cm 小方块（angle=0 轴对齐，人可 resize 到真实家具）。
func (i *AlarmFeedbackIngester) appendFeedbackObject(ctx context.Context, deviceAddr string, x, y int, eventID string, spec feedbackObjectSpec) {
	if i.db == nil || deviceAddr == "" {
		return
	}
	half := feedbackObjectHalfCm
	objID := spec.idPrefix + eventID
	obj := map[string]interface{}{
		"id":        objID,
		"name":      spec.name,
		"typeName":  spec.typeName,
		"typeValue": 0,
		"source":    "Feedback",
		"geometry": map[string]interface{}{
			"type": "rectangle",
			"data": map[string]interface{}{
				// 顺序对齐 FE Rectangle：[左上, 右上, 左下, 右下]
				"vertices": []map[string]int{
					{"x": x - half, "y": y - half},
					{"x": x + half, "y": y - half},
					{"x": x - half, "y": y + half},
					{"x": x + half, "y": y + half},
				},
			},
		},
		"visual":      map[string]interface{}{"color": spec.color, "transparent": false, "reflectivity": 30},
		"interactive": map[string]interface{}{"selected": false, "locked": false},
		"device":      map[string]interface{}{"category": spec.category, "type": spec.typeName},
		"height":      spec.height,
	}
	objJSON, err := json.Marshal(obj)
	if err != nil {
		i.logger.Warn("appendFeedbackObject: marshal", zap.String("type", spec.typeName), zap.Error(err))
		return
	}
	res, err := i.db.ExecContext(ctx, `
		UPDATE room_visual_layout
		SET canvas      = jsonb_set(canvas, '{objects}',
		                    COALESCE(canvas->'objects', '[]'::jsonb) || $2::jsonb),
		    canvas_hash = NULL,
		    version     = version + 1,
		    updated_at  = NOW()
		WHERE spatial_prefix = $1::inet
		  AND NOT (COALESCE(canvas->'objects', '[]'::jsonb) @> $3::jsonb)
	`, deviceAddr, string(objJSON), `[{"id":"`+objID+`"}]`)
	if err != nil {
		i.logger.Warn("appendFeedbackObject: update layout", zap.String("device_addr", deviceAddr), zap.String("type", spec.typeName), zap.Error(err))
		return
	}
	n, _ := res.RowsAffected()
	i.logger.Info(spec.logKey,
		zap.String("device_addr", deviceAddr), zap.String("object_id", objID),
		zap.Int("canvas_x", x), zap.Int("canvas_y", y), zap.Int64("rows", n))
}

// routeFeedback 根据 operation + parsed conditions 把 cell counter 增量。
// 返回触发的 route 名称列表（用于日志可观测）。nowMs = hand_time（人处理时刻），临时禁报窗以此为锚。
//
// 路由规则（设计 §3.2/§3.3/§1，label 见上方常量）：
//
//	operation=false_alarm:
//	  坐类(Chair/ShortSofa/Wheelchair)        → RestZoneConfirmed++ + MarkRestZoneByFeedback(AreaSit)
//	  躺类(Lounge/LongSofa) + ↳ Permanent     → 仅记 route，cell 走 pinFeedbackZone StampPriorRect(AreaLying)
//	  BlindArea/Enter（加区动作走 handler appendFeedbackObject，非 cell counter）
//	  Electric/AC                              → GhostCount++
//	  Error Pose / Out of Range                → 不进 counter（传感误差）
//	  Unknown 或全无勾选                        → FakeAlarmCount++
//	operation=verified:                         → RealFallCount++ + ClearNonHumanLearnedZone
//	  + ↳ Sticky veto                          → MarkLearnBlocked（永久封自动/反馈学抑制）
func (i *AlarmFeedbackIngester) routeFeedback(roomID string, x, y int, nowMs int64,
	operation string, pc ParsedConditions) []string {

	var routes []string

	// helper: 调 engine 改对应 cell 字段；返回是否成功（cell 在 grid 内）
	apply := func(fn func(*Cell)) bool {
		return i.engine.ApplyToCell(roomID, x, y, nowMs, fn)
	}

	switch operation {
	case "false_alarm":
		// 坐类 → AreaSit（wheelchair mobile decay 自滤）
		if pc.anySitClass() {
			locked := false
			if apply(func(c *Cell) {
				c.IncrRestZoneConfirmed()
				if c.MarkRestZoneByFeedback(AreaSit) {
					locked = true
				}
			}) {
				routes = append(routes, "rest_zone_sit")
				if locked {
					routes = append(routes, "locked_AreaSit")
				}
			}
		}
		// 躺类 permanent：cell 由 pinFeedbackZone StampPriorRect(AreaLying,SourceHuman) 刷（lying 高风险禁软自学，
		// 不在此走 MarkRestZoneByFeedback）；此处仅记 route 让 processOne 继续到 pin 块 / 无选不动
		if pc.FALoungeLongSofa {
			if pc.LoungePermanent {
				routes = append(routes, "lounge_permanent")
			} else {
				routes = append(routes, "lounge_no_action")
			}
		}
		// Metal/Mirror（静止反射）→ ghost 学习（继承旧 Electric/AC）
		if pc.FAMetalMirror {
			if apply(func(c *Cell) { c.IncrGhostCount() }) {
				routes = append(routes, "ghost")
			}
		}
		// Curtain/Fan/Plants（运动杂波）+ Pet（活体）：非人非反射 → 通用假报容忍
		if pc.FACurtainFanPlants || pc.FAPet {
			if apply(func(c *Cell) { c.IncrFakeAlarm() }) {
				routes = append(routes, "fake_generic")
			}
		}
		// Error Pose：姿态误判，不进任何 counter（§3.2，不污染 lying/sit/ghost）。
		// Unknown 或全无勾选 → 兜底 fake alarm 容忍。
		if pc.FAUnknown || !pc.anyFASelected() {
			if apply(func(c *Cell) { c.IncrFakeAlarm() }) {
				routes = append(routes, "fake_generic")
			}
		}
	case "verified":
		// 人确认真摔：RealFallCount++ + 擦非 Human 抑制/deny；勾 sticky veto 再永久封该 cell 自动学习。
		// verdict 本身即 ground truth，不再要求具体 Observed 勾选（FE Observed 标签已 per-type 化）。
		cleared := false
		blocked := false
		if apply(func(c *Cell) {
			c.IncrRealFallCount()
			if c.ClearNonHumanLearnedZone() {
				cleared = true
			}
			if pc.StickyVeto && c.MarkLearnBlocked() {
				blocked = true
			}
		}) {
			routes = append(routes, "real_fall")
			if cleared {
				routes = append(routes, "cleared_non_human_zone")
			}
			if blocked {
				routes = append(routes, "sticky_veto_learn_blocked")
			}
		}
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
