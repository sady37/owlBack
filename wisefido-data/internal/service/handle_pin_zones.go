package service

// handle_pin_zones.go — 护士 handle+PIN 的全部业务编排（单源在 data，sensor 被动）。
//
// 护士在 handle 弹窗勾"钉为坐/躺/反射/干扰/盲/门区"或"永不在此抑制(sticky)"，FE 把勾选写成
// remarks 里的 marker 行。data handle 完调本函数：解析 marker → fire 点(payload raw)转 canvas →
// 生成"以 fire 为心、rotate=0 正方形" → ① AppendFeedbackObject(layout+firmware+config:card)
// → 写回 alarm_events.metadata → ③ POST sensor /cell/stamp 即时刷 grid（桥接 reload 窗口）。
// Clear/Never 否决 → POST sensor /cell/veto：Clear 仅清 cell（sticky=false，可逆）；
// Never 清 + 永久封自动学习（sticky=true，Never ⊇ Clear）。

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/netip"
	"strings"

	"go.uber.org/zap"
)

// marker 字符串契约真相源 = owlFront/src/utils/alarm.ts（PIN_MARKER + FALL_RADAR_REASONS labels）。改须同步。
const (
	pinFalseAlarmBlockMark = "False Alarm Reason:"
	pinToken               = "☑ PIN"
	pinClearLearnMark      = "Clear learned fall suppression here"
	pinStickyVetoMark      = "Sticky veto: never auto-learn fall suppression here"
)

// pinHalfCm 永久加区方块半边长（40×40cm 方块，fire 点为心；与 sensor 旧 feedbackObjectHalfCm 一致）。
const pinHalfCm = 20

// pinSpec 一类永久加区的 layout 对象先验（typeName 决定 reload 后 engine AreaType）。
// area = observation.AreaType（单源，与 declare_orchestrator.engineAreaType 对齐）；conf = live-stamp 置信。
type pinSpec struct {
	idPrefix string
	name     string
	typeName string
	category string
	color    string
	height   int
	area     int
	conf     int
	logKey   string
}

var (
	pinSpecSit       = pinSpec{"feedback_sit_", "Chair (pin)", "Chair", "furniture", "#e9b48f", 90, 7, 80, "pin_sit"}
	pinSpecLounge    = pinSpec{"feedback_lounge_", "Recliner (pin)", "Recliner", "furniture", "#c19a6b", 40, 8, 99, "pin_lounge"}
	pinSpecReflector = pinSpec{"feedback_reflector_", "Reflector (pin)", "MetalCan", "interference", "#F5F5F5", 80, 3, 99, "pin_reflector"}
	pinSpecInterfere = pinSpec{"feedback_interfere_", "Interfer (pin)", "Interfere", "interference", "#82BBEB", 120, 6, 99, "pin_interfere"}
	pinSpecBlind     = pinSpec{"feedback_blind_", "Blind (pin)", "BlindArea", "furniture", "#5b3a1a", 80, 9, 99, "pin_blind"}
	pinSpecEnter     = pinSpec{"feedback_enter_", "Exit (pin)", "Enter", "structure", "#A9EAA9", 0, 4, 99, "pin_enter"}
)

// reason label → pinSpec（reason↔区 1:1；label 与 FE FALL_RADAR_REASONS 逐字一致）。
// remark 里行尾带 pinToken（"☑ PIN"）的 reason 行 → 钉对应区类型。
var pinReasonSpecs = []struct {
	label string
	spec  pinSpec
}{
	{"Lying on Sofa / Recliner", pinSpecLounge},
	{"Sit on Chair / Short Sofa", pinSpecSit},
	{"Metal / Mirror (reflection)", pinSpecReflector},
	{"Curtain / Fan / Plants", pinSpecInterfere},
	{"BlindArea / No Fall", pinSpecBlind},
	{"Enter / Door", pinSpecEnter},
}

// HandlePinZones 护士 handle+PIN 编排（best-effort，失败仅 warn，不阻塞 handle）。
// deviceHost = event.DeviceAddr 去 mask；triggerData = alarm_events.payload；notes = handler remarks。
func (s *RadarInstall) HandlePinZones(ctx context.Context, eventID, tenantID, deviceHost string, triggerData json.RawMessage, operation, notes string) {
	if deviceHost == "" || notes == "" {
		return
	}

	// 真摔否决该处自动抑制（需 fire 点 canvas 坐标）：Never ⊇ Clear。
	// block=Never（清 + MarkLearnBlocked 永久封）；clear=擦当前学习（Never 已含，故 ||）。
	block := strings.Contains(notes, pinStickyVetoMark)
	clear := block || strings.Contains(notes, pinClearLearnMark)

	var pins []pinSpec
	if operation == "false_alarm" || strings.Contains(notes, pinFalseAlarmBlockMark) {
		// 逐行解析：行尾带 pinToken 的 reason 行 → 钉该 reason 对应区（区类型单源 = 勾的 reason label）。
		for _, line := range strings.Split(notes, "\n") {
			if !strings.Contains(line, pinToken) {
				continue
			}
			for _, rs := range pinReasonSpecs {
				if strings.Contains(line, rs.label) {
					pins = append(pins, rs.spec)
					break
				}
			}
		}
	}

	if len(pins) == 0 && !clear {
		return
	}

	// fire 点 raw(h,v,z) → canvas（与 sensor RadarToCanvas 同公式：data toCanvasCoordinate 移植 FE）。
	px, py, pz, hasPos := positionFromPayload(triggerData)
	if !hasPos {
		s.logger.Debug("HandlePinZones: payload 无 position（producer 未落点）", zap.String("event_id", eventID))
		return
	}
	rf, ok := s.radarFrameForDevice(ctx, deviceHost)
	if !ok {
		s.logger.Warn("HandlePinZones: 无 radar frame（房 layout 缺该雷达对象，跳过）",
			zap.String("event_id", eventID), zap.String("device", deviceHost))
		return
	}
	cv := toCanvasCoordinate(radarPt{H: float64(px), V: float64(py), Z: float64(pz)}, rf)
	cx, cy := int(math.Round(cv.X)), int(math.Round(cv.Y))

	if clear {
		// 覆盖 fire 点 40×40 足迹（与 pin 方块同 pinHalfCm），非单 10cm cell——火点 cm 级抖动 + 危险点足迹 > 单格。
		s.notifySensorVeto(ctx, deviceHost, cx-pinHalfCm, cy-pinHalfCm, cx+pinHalfCm, cy+pinHalfCm, block, "veto_"+eventID)
	}

	for _, spec := range pins {
		s.applyPinZone(ctx, eventID, deviceHost, cx, cy, spec)
	}
}

// applyPinZone 单类永久加区：① AppendFeedbackObject(layout+firmware+config:card) → 写回
// alarm_events.metadata（Q2 坐标+名字）→ ③ POST sensor /cell/stamp 即时刷 grid。
func (s *RadarInstall) applyPinZone(ctx context.Context, eventID, deviceHost string, cx, cy int, spec pinSpec) {
	objID := spec.idPrefix + eventID
	h := pinHalfCm
	objJSON, err := json.Marshal(map[string]interface{}{
		"id":        objID,
		"name":      spec.name,
		"typeName":  spec.typeName,
		"typeValue": spec.area, // = observation.AreaType（单源，对齐 FE FURNITURE_CONFIGS）
		"source":    "Feedback",
		"geometry": map[string]interface{}{
			"type": "rectangle",
			"data": map[string]interface{}{
				// 顺序对齐 FE Rectangle：[左上, 右上, 左下, 右下]
				"vertices": []map[string]int{
					{"x": cx - h, "y": cy - h},
					{"x": cx + h, "y": cy - h},
					{"x": cx - h, "y": cy + h},
					{"x": cx + h, "y": cy + h},
				},
			},
		},
		"visual":      map[string]interface{}{"color": spec.color, "transparent": false, "reflectivity": 30},
		"interactive": map[string]interface{}{"selected": false, "locked": false},
		"device":      map[string]interface{}{"category": spec.category, "type": spec.typeName},
		"height":      spec.height,
	})
	if err != nil {
		s.logger.Warn("applyPinZone: marshal object", zap.String("type", spec.typeName), zap.Error(err))
		return
	}

	if _, err := s.AppendFeedbackObject(ctx, deviceHost, objID, string(objJSON)); err != nil {
		s.logger.Warn("applyPinZone: AppendFeedbackObject failed",
			zap.String("event_id", eventID), zap.String("type", spec.typeName), zap.Error(err))
		return
	}

	s.appendPinNote(ctx, eventID, spec, cx, cy)
	s.notifySensorStamp(ctx, deviceHost, cx-pinHalfCm, cy-pinHalfCm, cx+pinHalfCm, cy+pinHalfCm, spec.area, spec.conf)

	s.logger.Info("pin_zone_applied", zap.String("event_id", eventID), zap.String("type", spec.typeName),
		zap.String("log_key", spec.logKey), zap.Int("area", spec.area), zap.Int("cx", cx), zap.Int("cy", cy))
}

// appendPinNote 把 fire 点 + 生成的 pin rect（canvas 坐标，人读审计）追加进 alarm_events.notes（Q2，单源不双写 metadata）。
// 格式：fall(cx,cy) add <Type>(x1,y1;x2,y2;x3,y3;x4,y4)，4 角同 layout object 顶点序 [左上,右上,左下,右下]。
func (s *RadarInstall) appendPinNote(ctx context.Context, eventID string, spec pinSpec, cx, cy int) {
	h := pinHalfCm
	line := fmt.Sprintf("\nfall(%d,%d) add %s(%d,%d;%d,%d;%d,%d;%d,%d)",
		cx, cy, spec.typeName,
		cx-h, cy-h, cx+h, cy-h, cx-h, cy+h, cx+h, cy+h)
	if _, err := s.db.ExecContext(ctx,
		`UPDATE alarm_events SET handler_notes = COALESCE(handler_notes, '') || $2 WHERE event_id = $1`,
		eventID, line); err != nil {
		s.logger.Warn("appendPinNote: update notes failed",
			zap.String("event_id", eventID), zap.Error(err))
	}
}

// radarFrameForDevice 在 /128→/88→/80 layout canvas 里找该 device 的 Radar 对象 → radarFrame。
func (s *RadarInstall) radarFrameForDevice(ctx context.Context, deviceHost string) (radarFrame, bool) {
	addr, err := netip.ParseAddr(deviceHost)
	if err != nil {
		return radarFrame{}, false
	}
	for _, m := range []int{128, 88, 80} {
		cidr := netip.PrefixFrom(addr, m).Masked().String()
		canvas := s.queryRoomVisualLayoutCanvas(ctx, cidr)
		if len(canvas) == 0 {
			continue
		}
		var doc cvDoc
		if json.Unmarshal(canvas, &doc) != nil {
			continue
		}
		for _, o := range doc.Objects {
			rf, ok := radarFrameOf(o)
			if ok && sameHost(rf.deviceAddr, deviceHost) {
				return rf, true
			}
		}
	}
	return radarFrame{}, false
}

// sameHost 比较两个 IPv6（去 mask）是否同 host。
func sameHost(a, b string) bool {
	pa, ea := netip.ParseAddr(strings.SplitN(a, "/", 2)[0])
	pb, eb := netip.ParseAddr(strings.SplitN(b, "/", 2)[0])
	return ea == nil && eb == nil && pa == pb
}

// stampCanvasCells 「stamp 跟随 layout 写入」：每次 layout 写入后，按 canvas 各 cell-area 对象的完整
// 几何即时刷 sensor grid（pin 初始方块 / FE resize 后完整矩形）。桥接 config:card reload 窗口 —— 修
// E598「resize 后 chair 区不生效要等重启」。sensor SetPrior 幂等 + 优先级 gate，重复刷安全。
func (s *RadarInstall) stampCanvasCells(ctx context.Context, canvas []byte) {
	var doc cvDoc
	if len(canvas) == 0 || json.Unmarshal(canvas, &doc) != nil {
		return
	}
	var deviceAddr string
	for _, o := range doc.Objects {
		if rf, ok := radarFrameOf(o); ok {
			deviceAddr = strings.SplitN(rf.deviceAddr, "/", 2)[0]
			break
		}
	}
	if deviceAddr == "" {
		return // 该 canvas 无雷达对象，无法路由到 room grid（如 /128 纯 feedback 行，靠 applyPinZone 显式刷）
	}
	for _, o := range doc.Objects {
		area := engineAreaType(o.TypeName)
		if area == 0 {
			continue // Wall / Radar / 未识别：无 cell
		}
		verts := objectVertices(o)
		if len(verts) < 3 {
			continue
		}
		x1, y1, x2, y2 := boundsOf(verts)
		conf := 99
		if o.TypeName == "Chair" {
			conf = 80
		}
		s.notifySensorStamp(ctx, deviceAddr, x1, y1, x2, y2, area, conf)
	}
}

// boundsOf canvas 顶点集的轴对齐包围盒（cm，四舍五入）。
func boundsOf(verts []cvPoint) (x1, y1, x2, y2 int) {
	minX, minY := verts[0].X, verts[0].Y
	maxX, maxY := verts[0].X, verts[0].Y
	for _, p := range verts[1:] {
		minX = math.Min(minX, p.X)
		minY = math.Min(minY, p.Y)
		maxX = math.Max(maxX, p.X)
		maxY = math.Max(maxY, p.Y)
	}
	return int(math.Round(minX)), int(math.Round(minY)), int(math.Round(maxX)), int(math.Round(maxY))
}

// positionFromPayload 从 alarm payload 顶层取 fire 点 raw 坐标（producer 落点单源）。移植 sensor。
func positionFromPayload(triggerData json.RawMessage) (x, y, z int, ok bool) {
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
	return jsonNum(px), jsonNum(py), jsonNum(pz), true
}

// jsonNum interface{}(JSON number / string) → int。
func jsonNum(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		f, _ := n.Float64()
		return int(f)
	case string:
		var f float64
		_, _ = fmt.Sscanf(n, "%f", &f)
		return int(f)
	}
	return 0
}
