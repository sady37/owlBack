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
//         "Permanent (update layout)"      → pinFeedbackZone：追加 Recliner object + StampPriorRect(AreaLying,SourceHuman)
//         (无 ↳ 行)                         → 不动 layout
//     ☑ BlindArea / No Fall (家具后遮挡盲区) ↳ permanent → 追加 BlindArea object（reload→AreaActive；12min tFloor 快速兜底盲区真摔，非 deny）
//     ☑ Enter / Door (未画出口)             ↳ 直接 → 追加 Enter object（reload→AreaEnter）
//     ☑ Metal / Mirror (reflection)        ↳ reflector zone → 追加 MetalCan object（reload→AreaReflector 豁免；ghost 走 realness OR）
//     ☑ Curtain / Fan / Plants             ↳ interference zone → 追加 Interfere object（reload→AreaInterfer，floor 豁免；帘动→瞬态track停后必误火）
//     ☑ Error Pose Detection / Out of Detection Range → 不进任何 counter（传感误差，非空间属性）
//     ☑ Unknown / 无勾选                    → FakeAlarmCount++（兜底容忍）
//
//   operation='verified' (人确认真摔):       → RealFallCount++ + ClearNonHumanLearnedZone（擦非 Human 抑制/deny）
//     ↳ Sticky veto: ...                   → 额外 MarkLearnBlocked（永久禁该 cell 自动/反馈学抑制）
//
// 流程（事件触发：wisefido-data handle 完直调 sensor /roomengine/feedback/ingest {event_id}）：
//  1. event_id 内存去重（cell counter 非幂等，重复调禁双增）
//  2. SELECT alarm_events WHERE event_id=$1 AND operation IN (false_alarm, verified) AND event_type IN (Fall, ...)
//  3. 解析 notes 提取 ☑ checkbox → 路由到对应 cell counter
//  4. 读 payload 顶层 position_x/y/z（producer 落点单源）→ RadarToCanvas → cell mark

package roomengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"owl-common/radarutils"
)

// FeedbackEvent data handle 完 POST 过来的反馈事件（payload 由 data 推，sensor 不再回读 alarm_events）。
type FeedbackEvent struct {
	EventID       string          `json:"event_id"`
	DeviceAddr    string          `json:"device_addr"` // host text（无 mask）
	EventType     string          `json:"event_type"`
	Operation     string          `json:"operation"`
	TriggeredAtMs int64           `json:"triggered_at_ms"`
	HandTimeMs    int64           `json:"hand_time_ms"`
	Payload       json.RawMessage `json:"payload"`
	HandlerNotes  string          `json:"handler_notes"`
}

// AlarmFeedbackIngester 把 data 推来的 false_alarm/verified 反馈喂回 cell（不连库，payload 随事件推送）。
type AlarmFeedbackIngester struct {
	engine *Engine
	logger *zap.Logger

	// event_id 去重集：事件触发是乱序单条，processOne 的 cell counter 非幂等，重复调禁双增。
	mu   sync.Mutex
	seen map[string]struct{}
}

func NewAlarmFeedbackIngester(engine *Engine, logger *zap.Logger) *AlarmFeedbackIngester {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AlarmFeedbackIngester{
		engine: engine,
		logger: logger,
		seen:   map[string]struct{}{},
	}
}

// False Alarm Reason / Lounge placement / Sticky veto 标记字符串——契约真相源 =
// owlFront/src/utils/alarm.ts（FALL_RADAR_REASONS / LOUNGE_PLACEMENT_MARKER / STICKY_VETO_MARKER）。
// FE 把勾选写成 remarks 里的 "☑ <label>" 行 + "↳ <marker>" 子行，这里按 substring 反解。改这些常量必须与 alarm.ts 同步。
const (
	faSitChairShortSofa = "Sit on Chair / Short Sofa"
	faLoungeLongSofa    = "Lying on Sofa / Recliner"
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

	// 永久加区 marker（owlFront alarm.ts 单源）。reflector(metal/mirror) 与 interfere(curtain/fan) 拆开：
	// 前者 AreaReflector 豁免（无真人静态反射，ghost 走 realness OR），后者 AreaInterfer 同样 floor 豁免（帘动→生瞬态 track 停后 still-timeout 必误火，故不给 floor 网；firmware type3 masking 源头滤）。
	sitPinMark        = "Sit zone: pin permanently (update layout)"
	reflectorZoneMark = "Reflector zone: mark permanently (mirror/metal, update layout)"
	interfereZoneMark = "Interference zone: mark permanently (curtain/fan, update layout)"
	blindAreaMark     = "Blind area: mark permanently (no-fall zone, update layout)"
	enterDoorMark     = "Exit zone: add permanently (update layout)"
)

// ParsedConditions 从 notes 解析的勾选/标记状态（label 见上方常量）。
type ParsedConditions struct {
	// 坐类（→ AreaSit 学习）
	FASitChairShortSofa bool
	FAWheelchair        bool
	// 盲区/出口（左列良性，→ AreaActive / AreaEnter，不进 sit 学习）
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
	LoungePermanent bool // ↳ permanent → pinFeedbackZone(AreaLying)：追加 Recliner object + StampPriorRect

	// 永久加区二次动作（追加 layout 对象）
	SitPin        bool // ↳ sit zone pin → 追加 Chair object（reload→AreaSit 神圣）
	ReflectorZone bool // ↳ reflector zone(metal/mirror) → 追加 MetalCan object（reload→AreaReflector 豁免）
	InterfereZone bool // ↳ interference zone(curtain/fan) → 追加 Interfere object（reload→AreaInterfer，floor 豁免；帘动→瞬态track停后必误火）
	BlindArea     bool // ↳ blind area permanent → 追加 BlindArea object（reload→AreaActive 12min）
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

// anySitClass 坐类（Chair/ShortSofa/Wheelchair → AreaSit）。Behind 已退役为 BlindArea(→AreaActive)，不再进 sit。
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
	out.ReflectorZone = strings.Contains(notes, reflectorZoneMark)
	out.InterfereZone = strings.Contains(notes, interfereZoneMark)
	out.BlindArea = strings.Contains(notes, blindAreaMark)
	out.EnterDoorZone = strings.Contains(notes, enterDoorMark)
	out.StickyVeto = strings.Contains(notes, stickyVetoMark)

	return out
}

// IngestOne 处理单条 alarm 反馈（事件触发：wisefido-data handle 完 POST 推送，payload 随带）。
// event_id 去重：同一 event 重复调只跑一次 processOne（cell counter 非幂等，禁双增）。
// payload/notes 由 data 推（不再回读 alarm_events，铁律写硬读软 [[sensor_asks_data_sync_not_db]] §9.1）。
func (i *AlarmFeedbackIngester) IngestOne(ctx context.Context, fe FeedbackEvent) (bool, error) {
	if i.engine == nil {
		return false, fmt.Errorf("ingester not configured (nil engine)")
	}
	if fe.EventID == "" {
		return false, fmt.Errorf("empty event_id")
	}

	i.mu.Lock()
	_, dup := i.seen[fe.EventID]
	i.mu.Unlock()
	if dup {
		i.logger.Debug("alarm_feedback: event already processed", zap.String("event_id", fe.EventID))
		return false, nil
	}

	// 裁决看 handler_notes 的块头（False Alarm Reason: / Observed Conditions:），与 alarm 生命周期正交。
	ok := i.processOne(ctx, fe.EventID, fe.DeviceAddr, fe.EventType, fe.Operation,
		time.UnixMilli(fe.TriggeredAtMs), time.UnixMilli(fe.HandTimeMs), fe.Payload, fe.HandlerNotes)

	i.mu.Lock()
	i.seen[fe.EventID] = struct{}{}
	i.mu.Unlock()

	return ok, nil
}

// processOne 处理单条 alarm（false_alarm 或 verified）：
// 1) 解析 notes 提取 ☑ checkbox
// 2) 读 payload 顶层 position → canvas
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

	// 2. fire 坐标直接读 payload 顶层（落点统一：sensor dbn 自带 incident 坐标 / 固件由 cardagg
	//    AlarmRouter 按 track_id 从 MonitorBuffer 补）。不再 lookback monitor_stream。
	loX, loY, loZ, hasPos := positionFromPayload(triggerData)
	if !hasPos {
		i.logger.Debug("alarm_feedback: payload 无 position（producer 未落点）",
			zap.String("event_id", eventID),
			zap.String("device_addr", deviceAddr))
		return false
	}

	// 3. RadarToCanvas
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
	//   裁决看 notes 块头（operation 兜底），与 auto_resolved 等生命周期正交。
	if pc.HasFalseAlarmBlock || operation == "false_alarm" {
		if pc.FALoungeLongSofa && pc.LoungePermanent {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackLoungeObject)
		}
		if pc.SitPin {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackSitObject)
		}
		if pc.ReflectorZone {
			i.pinFeedbackZone(ctx, deviceAddr, roomID, canvas.X, canvas.Y, eventID, feedbackReflectorObject)
		}
		if pc.InterfereZone {
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
//   Recliner→AreaLying(90min) / Chair→AreaSit / MetalCan→AreaReflector(豁免) / Interfere→AreaInterfer(豁免) /
//   BlindArea→AreaActive(12min) / Enter→AreaEnter。
// 均 source=Feedback：drawObjects 渲染虚线、updateRadarAreas 过滤不下发雷达（人确认转 Human 后才 declare）。
var (
	feedbackLoungeObject    = feedbackObjectSpec{"feedback_lounge_", "Recliner-PIN", "Recliner", "furniture", "#c19a6b", 40, "feedback_lounge_object_appended", AreaLying, 99}
	feedbackSitObject       = feedbackObjectSpec{"feedback_sit_", "Chair-PIN", "Chair", "furniture", "#ffff00", 90, "feedback_sit_object_appended", AreaSit, chairPriorConf}
	feedbackReflectorObject = feedbackObjectSpec{"feedback_reflector_", "Reflector-PIN", "MetalCan", "interference", "#F5F5F5", 80, "feedback_reflector_object_appended", AreaReflector, 99}
	feedbackInterfereObject = feedbackObjectSpec{"feedback_interfere_", "Interfer-PIN", "Interfere", "interference", "#4a4a4a", 120, "feedback_interfere_object_appended", AreaInterfer, 99}
	feedbackBlindObject     = feedbackObjectSpec{"feedback_blind_", "Blind-PIN", "BlindArea", "furniture", "#5b3a1a", 80, "feedback_blind_object_appended", AreaActive, 99}
	feedbackEnterObject     = feedbackObjectSpec{"feedback_enter_", "Exit-PIN", "Enter", "structure", "#a9eaa9", 0, "feedback_enter_object_appended", AreaEnter, 99}
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
var feedbackHTTPClient = &http.Client{Timeout: 5 * time.Second}

func feedbackDataURL() string {
	if v := os.Getenv("SENSOR_DATA_URL"); v != "" {
		return v
	}
	return "http://127.0.0.1:8080"
}

func (i *AlarmFeedbackIngester) appendFeedbackObject(ctx context.Context, deviceAddr string, x, y int, eventID string, spec feedbackObjectSpec) {
	if deviceAddr == "" {
		return
	}
	half := feedbackObjectHalfCm
	objID := spec.idPrefix + eventID
	obj := map[string]interface{}{
		"id":        objID,
		"name":      spec.name,
		"typeName":  spec.typeName,
		"typeValue": int(spec.liveArea), // = observation.AreaType（单源，对齐 FE FURNITURE_CONFIGS）
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
	// 经 data 单写入口（取代直写 room_visual_layout）：data 原子 append + 真追加则即时下发设备 + config:card。
	body, err := json.Marshal(map[string]interface{}{
		"spatial_prefix": deviceAddr,
		"object_id":      objID,
		"object":         json.RawMessage(objJSON),
	})
	if err != nil {
		i.logger.Warn("appendFeedbackObject: marshal body", zap.Error(err))
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, feedbackDataURL()+"/internal/radar/feedback-object", bytes.NewReader(body))
	if err != nil {
		i.logger.Warn("appendFeedbackObject: new request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := feedbackHTTPClient.Do(req)
	if err != nil {
		i.logger.Warn("appendFeedbackObject: post data", zap.String("device_addr", deviceAddr), zap.String("type", spec.typeName), zap.Error(err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		i.logger.Warn("appendFeedbackObject: data non-200", zap.String("device_addr", deviceAddr), zap.Int("status", resp.StatusCode))
		return
	}
	i.logger.Info(spec.logKey,
		zap.String("device_addr", deviceAddr), zap.String("object_id", objID),
		zap.Int("canvas_x", x), zap.Int("canvas_y", y))
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

	// 裁决看 handle 写的 notes 块头（False Alarm Reason / Observed Conditions），operation 仅兜底。
	// 与 alarm 生命周期正交：auto_resolved 把 operation 覆盖掉也不影响（块头仍在）。
	isFalseAlarm := pc.HasFalseAlarmBlock || operation == "false_alarm"
	isVerified := pc.HasObservedBlock || operation == "verified"
	switch {
	case isFalseAlarm:
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
		// 盲区/出口（左列良性）：cell 由 pinFeedbackZone 追加 layout 对象处理（非 cell counter）；
		// 此处仅记 route 让 processOne 不在 len(routes)==0 提前 return → pin 块可达（否则纯 Enter/Blind 反馈被丢）。
		if pc.FABlindAreaNoFall {
			routes = append(routes, "blind_area")
		}
		if pc.FAEnterDoor {
			routes = append(routes, "enter_door")
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
	case isVerified:
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

// positionFromPayload 从 alarm_events.payload 顶层读 fire 坐标（raw firmware h/v/z, cm）。
// 落点统一单源：sensor dbn 腿 payload 自带 incident 坐标；固件直发由 cardagg AlarmRouter 按 track_id
// 从 MonitorBuffer 补进 payload（见 alarm_router.go persist）。三轴任一缺 → ok=false（不落点）。
func positionFromPayload(triggerData []byte) (x, y, z int, ok bool) {
	if len(triggerData) == 0 {
		return 0, 0, 0, false
	}
	var p map[string]interface{}
	if err := json.Unmarshal(triggerData, &p); err != nil {
		return 0, 0, 0, false
	}
	px, okx := p["position_x"]
	py, oky := p["position_y"]
	pz, okz := p["position_z"]
	if !okx || !oky || !okz {
		return 0, 0, 0, false
	}
	return jsonInt(px), jsonInt(py), jsonInt(pz), true
}
