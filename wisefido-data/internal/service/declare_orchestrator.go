package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"owl-common/observation"
	"owl-common/radarutils"

	"go.uber.org/zap"
)

// declare_orchestrator: canvas → 每雷达的 declare set（area_id 粘性 + typeName 级下发码 + 雷达系顶点 cm）。
// 移植自 FE radarUtils（toRadarCoordinate / getObjectVertices / getObjectVerticesInRadar），
// 几何必须与 FE 逐字对齐；类型下发策略按 pin_declare_downlink_plan §2.1（per-typeName，Chair 停发）。
// T1：仅 shadow log，不下发。

type cvPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type cvBoundary struct {
	LeftH  float64 `json:"leftH"`
	RightH float64 `json:"rightH"`
	FrontV float64 `json:"frontV"`
	RearV  float64 `json:"rearV"`
}

type cvRadar struct {
	InstallModel string     `json:"installModel"`
	Boundary     cvBoundary `json:"boundary"`
}

type cvGeometry struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type cvObject struct {
	ID         string  `json:"id"`
	TypeName   string  `json:"typeName"`
	Angle      float64 `json:"angle"`
	Source     string  `json:"source"`
	DeviceAddr string  `json:"device_addr"`
	Device     struct {
		Category string `json:"category"`
		IOT      struct {
			Radar *cvRadar `json:"radar"`
		} `json:"iot"`
	} `json:"device"`
	Geometry cvGeometry `json:"geometry"`
}

type cvDoc struct {
	Objects []cvObject `json:"objects"`
}

// radarFrame 一个雷达对象的下发上下文。
type radarFrame struct {
	id           string
	deviceAddr   string
	center       cvPoint
	angle        float64
	installModel string
	boundary     cvBoundary
}

// radarPt 雷达系顶点（cm，整十）。
type radarPt struct {
	H float64
	V float64
	Z float64
}

// declareArea 单个区域的下发产物。
type declareArea struct {
	AreaID   int
	FwType   int
	TypeName string
	ObjectID string
	Vertices [4]radarPt
}

// declareSkip 落 layout/cell 但不下发固件的对象（策略不下发，如 Chair/BlindArea/Wall）。
type declareSkip struct {
	ObjectID   string
	TypeName   string
	Reason     string
	DeviceAddr string // 所属雷达（供 verify 把 not-sent 对象并到对应雷达）
	Vertices   [4]radarPt
}

// engineAreaType typeName → 引擎 AreaType(0-9)（not-sent 对象在 verify 显引擎类型，如 Chair=7 Sit）。
func engineAreaType(typeName string) int {
	switch typeName {
	case "Bed":
		return 2
	case "MonitorBed":
		return 5
	case "Enter":
		return 4
	case "Mirror", "MetalCan":
		return 3
	case "Curtain", "WheelChair", "Interfere":
		return 6
	case "Chair":
		return 7
	case "Recliner", "LongSofa":
		return 8
	case "BlindArea":
		return 9
	case "Furniture", "Deny":
		return 1
	default: // Wall 等
		return 0
	}
}

// declareDrop 本该下发但超 16 槽被优先级挤掉的区域（→ 回客户端截断提示）。
type declareDrop struct {
	ObjectID string
	TypeName string
	FwType   int
	Vertices [4]radarPt
}

// maxDeclareAreasData 设备最多 16 区（厂家 3.4.3）；超出按保留优先级丢尾部，单源同 qinglan。
const maxDeclareAreasData = 16

// declareKeepPriority 保留优先级（小=优先占槽）：删除 > 床(2/5) > 门(4) > 噪(3) > 自定义(1) > 其余。
func declareKeepPriority(fwType int) int {
	switch fwType {
	case 0:
		return 0
	case 2, 5:
		return 1
	case 4:
		return 2
	case 3:
		return 3
	case 1:
		return 4
	default:
		return 5
	}
}

// firmwarePolicy 按 typeName 决定下发码与是否下发（§2.1 权威）。
// isMonitorBed=Bed∩Radar 派生。Chair(Sit)/BlindArea/Wall 不下发；其余按表。
func firmwarePolicy(typeName string, isMonitorBed bool) (code int, send bool) {
	switch typeName {
	case "Bed":
		if isMonitorBed {
			return 5, true
		}
		return 2, true
	case "MonitorBed":
		return 5, true
	case "Enter":
		return 4, true
	case "MetalCan", "Mirror", "Curtain", "WheelChair", "Interfere":
		return 3, true
	case "Recliner", "Furniture":
		return 1, true
	default: // Chair / BlindArea / Wall / 未识别
		return 0, false
	}
}

// radarMount 把 declare 的 radarFrame 折成 owl-common/radarutils.RadarMount（corn 内旋 -45 交由 radarutils）。
// center/angle 取整 ≤1cm 误差，60G 雷达 ±十几cm + cell 容错下可接受（坐标转换全仓单源 int）。
func (r radarFrame) radarMount() radarutils.RadarMount {
	im := radarutils.InstallCeiling
	switch r.installModel {
	case "corn":
		im = radarutils.InstallCorn
	case "wall":
		im = radarutils.InstallWall
	}
	return radarutils.RadarMount{
		Center:       radarutils.Point{X: roundInt(r.center.X), Y: roundInt(r.center.Y)},
		Rotation:     roundInt(r.angle),
		InstallModel: im,
	}
}

// toRadarCoordinate 画布(x,y)→雷达系(h,v)。公式单源到 owl-common/radarutils。
func toRadarCoordinate(x, y float64, r radarFrame) radarPt {
	rp := radarutils.CanvasToRadar(radarutils.Point{X: roundInt(x), Y: roundInt(y)}, r.radarMount())
	return radarPt{H: float64(rp.H), V: float64(rp.V), Z: 0}
}

// toCanvasCoordinate 雷达系(h,v,z)→画布(x,y)。公式单源到 owl-common/radarutils（Z 透传）。
func toCanvasCoordinate(p radarPt, r radarFrame) cvPoint {
	cv := radarutils.RadarToCanvas(radarutils.RadarPoint{H: roundInt(p.H), V: roundInt(p.V)}, r.radarMount())
	return cvPoint{X: float64(cv.X), Y: float64(cv.Y), Z: p.Z}
}

func roundInt(x float64) int { return int(math.Round(x)) }

// objectVertices 对象在画布系的顶点（含对象自转）。furniture+structure 都算（feedback 门=structure）；
// 非矩形/circle 几何（measure_dots/point/雷达）由 switch default 返回 nil 自然排除。
func objectVertices(o cvObject) []cvPoint {
	switch o.Geometry.Type {
	case "rectangle":
		var d struct {
			Vertices []cvPoint `json:"vertices"`
		}
		if json.Unmarshal(o.Geometry.Data, &d) != nil || len(d.Vertices) < 4 {
			return nil
		}
		v := d.Vertices
		ordered := []cvPoint{v[0], v[1], v[3], v[2]}
		if o.Angle == 0 {
			return ordered
		}
		cx := (v[0].X + v[3].X) / 2
		cy := (v[0].Y + v[3].Y) / 2
		a := o.Angle * math.Pi / 180
		cosA, sinA := math.Cos(a), math.Sin(a)
		out := make([]cvPoint, 4)
		for i, p := range ordered {
			dx := p.X - cx
			dy := p.Y - cy
			out[i] = cvPoint{X: cx + dx*cosA + dy*sinA, Y: cy - dx*sinA + dy*cosA, Z: p.Z}
		}
		return out
	case "circle":
		var d struct {
			Center cvPoint `json:"center"`
			Radius float64 `json:"radius"`
		}
		if json.Unmarshal(o.Geometry.Data, &d) != nil {
			return nil
		}
		c, rad := d.Center, d.Radius
		return []cvPoint{
			{X: c.X + rad, Y: c.Y - rad, Z: c.Z},
			{X: c.X + rad, Y: c.Y + rad, Z: c.Z},
			{X: c.X - rad, Y: c.Y + rad, Z: c.Z},
			{X: c.X - rad, Y: c.Y - rad, Z: c.Z},
		}
	case "polygon":
		var d struct {
			Vertices []cvPoint `json:"vertices"`
		}
		if json.Unmarshal(o.Geometry.Data, &d) != nil {
			return nil
		}
		return d.Vertices
	default:
		return nil
	}
}

// radarBoundaryVertices 雷达边界在画布系的顶点（已含旋转）。移植 radarUtils.ts:134。
func radarBoundaryVertices(r radarFrame) []cvPoint {
	b := r.boundary
	switch r.installModel {
	case "corn":
		L := b.LeftH
		R := b.RightH
		a := r.angle * math.Pi / 180
		cosA, sinA := math.Cos(a), math.Sin(a)
		s := math.Sqrt2
		diamond := [][2]float64{
			{R / s, R / s},
			{0, 0},
			{(R - L) / s, (R + L) / s},
			{-L / s, L / s},
		}
		out := make([]cvPoint, 4)
		for i, d := range diamond {
			dx, dy := d[0], d[1]
			out[i] = cvPoint{X: r.center.X + dx*cosA + dy*sinA, Y: r.center.Y - dx*sinA + dy*cosA}
		}
		return out
	case "wall":
		rv := []radarPt{
			{H: -b.RightH, V: 0}, {H: b.LeftH, V: 0},
			{H: -b.RightH, V: b.FrontV}, {H: b.LeftH, V: b.FrontV},
		}
		return mapToCanvas(rv, r)
	default: // ceiling
		rv := []radarPt{
			{H: -b.RightH, V: -b.RearV}, {H: b.LeftH, V: -b.RearV},
			{H: -b.RightH, V: b.FrontV}, {H: b.LeftH, V: b.FrontV},
		}
		return mapToCanvas(rv, r)
	}
}

func mapToCanvas(rv []radarPt, r radarFrame) []cvPoint {
	out := make([]cvPoint, len(rv))
	for i, p := range rv {
		out[i] = toCanvasCoordinate(p, r)
	}
	return out
}

// isPointInPolygon 射线法。移植 radarUtils.ts:302。
func isPointInPolygon(p cvPoint, verts []cvPoint) bool {
	if len(verts) < 3 {
		return false
	}
	inside := false
	for i, j := 0, len(verts)-1; i < len(verts); j, i = i, i+1 {
		vi, vj := verts[i], verts[j]
		if (vi.Y > p.Y) != (vj.Y > p.Y) {
			intersectX := (vj.X-vi.X)*(p.Y-vi.Y)/(vj.Y-vi.Y) + vi.X
			if p.X < intersectX {
				inside = !inside
			}
		}
	}
	return inside
}

// objectInBoundaryWithTolerance 家具全顶点落在雷达边界 AABB±tol 内。移植 radarUtils.ts:547（tol=30）。
func objectInBoundaryWithTolerance(o cvObject, r radarFrame, tol float64) bool {
	ov := objectVertices(o)
	if len(ov) == 0 {
		return false
	}
	bv := radarBoundaryVertices(r)
	minX, maxX := bv[0].X, bv[0].X
	minY, maxY := bv[0].Y, bv[0].Y
	for _, v := range bv {
		minX = math.Min(minX, v.X)
		maxX = math.Max(maxX, v.X)
		minY = math.Min(minY, v.Y)
		maxY = math.Max(maxY, v.Y)
	}
	minX, maxX, minY, maxY = minX-tol, maxX+tol, minY-tol, maxY+tol
	for _, v := range ov {
		if v.X < minX || v.X > maxX || v.Y < minY || v.Y > maxY {
			return false
		}
	}
	return true
}

func roundToTen(n float64) float64 { return math.Round(n/10) * 10 }

// objectVerticesInRadar 家具顶点→雷达系→整十→按 v 升序(同 v 按 h 升序)。移植 radarUtils.ts:591。
func objectVerticesInRadar(o cvObject, r radarFrame) []radarPt {
	cv := objectVertices(o)
	out := make([]radarPt, len(cv))
	for i, v := range cv {
		p := toRadarCoordinate(v.X, v.Y, r)
		out[i] = radarPt{H: roundToTen(p.H), V: roundToTen(p.V), Z: roundToTen(v.Z)}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if math.Abs(out[i].V-out[j].V) < 1 {
			return out[i].H < out[j].H
		}
		return out[i].V < out[j].V
	})
	return out
}

func radarFrameOf(o cvObject) (radarFrame, bool) {
	if o.TypeName != "Radar" || o.Geometry.Type != "point" || o.Device.IOT.Radar == nil {
		return radarFrame{}, false
	}
	var c cvPoint
	if json.Unmarshal(o.Geometry.Data, &c) != nil {
		return radarFrame{}, false
	}
	rc := o.Device.IOT.Radar
	im := rc.InstallModel
	if im == "" {
		im = "ceiling"
	}
	return radarFrame{
		id: o.ID, deviceAddr: o.DeviceAddr, center: c, angle: o.Angle,
		installModel: im, boundary: rc.Boundary,
	}, true
}

func bedContainsAnyRadar(bed cvObject, radars []radarFrame) bool {
	bv := objectVertices(bed)
	if len(bv) < 3 {
		return false
	}
	for _, r := range radars {
		if isPointInPolygon(r.center, bv) {
			return true
		}
	}
	return false
}

// computeDeclareSet 每雷达 → declare set + 跳过清单。
func computeDeclareSet(canvas []byte) (map[string][]declareArea, []declareSkip, []declareDrop, error) {
	var doc cvDoc
	if err := json.Unmarshal(canvas, &doc); err != nil {
		return nil, nil, nil, err
	}
	var radars []radarFrame
	for _, o := range doc.Objects {
		if rf, ok := radarFrameOf(o); ok {
			radars = append(radars, rf)
		}
	}
	perRadar := make(map[string][]declareArea, len(radars))
	var skips []declareSkip
	var drops []declareDrop
	for _, r := range radars {
		type candidate struct {
			o     cvObject
			code  int
			verts [4]radarPt
		}
		var sendable []candidate
		for _, o := range doc.Objects {
			if o.Device.Category == "iot" { // 跳雷达/sleepad/sensor；furniture+structure(feedback门) 都处理
				continue
			}
			// pin = 权威下发：Human(手画) + Feedback(护工 pin) 都进 declare；Learned(自动学) 仅引擎提示，不下发。
			if o.Source == "Learned" {
				continue
			}
			if !objectInBoundaryWithTolerance(o, r, 30) {
				continue
			}
			rv := objectVerticesInRadar(o, r)
			var verts [4]radarPt
			for k := 0; k < 4 && k < len(rv); k++ {
				verts[k] = rv[k]
			}
			isMB := o.TypeName == "Bed" && bedContainsAnyRadar(o, radars)
			code, send := firmwarePolicy(o.TypeName, isMB)
			if !send {
				skips = append(skips, declareSkip{ObjectID: o.ID, TypeName: o.TypeName, Reason: "policy:not-sent", DeviceAddr: r.deviceAddr, Vertices: verts})
				continue
			}
			sendable = append(sendable, candidate{o: o, code: code, verts: verts})
		}
		// 超 16 槽：按保留优先级稳定排序，留前 16，其余 → 截断提示。
		if len(sendable) > maxDeclareAreasData {
			sort.SliceStable(sendable, func(i, j int) bool {
				return declareKeepPriority(sendable[i].code) < declareKeepPriority(sendable[j].code)
			})
			for _, c := range sendable[maxDeclareAreasData:] {
				drops = append(drops, declareDrop{ObjectID: c.o.ID, TypeName: c.o.TypeName, FwType: c.code, Vertices: c.verts})
			}
			sendable = sendable[:maxDeclareAreasData]
		}
		// 顺序占槽 0..N-1（确定性）：旧/新均如此编号，删旧=DELETE 尾部 [N..M-1]，无需追踪设备态。
		areas := make([]declareArea, 0, len(sendable))
		for i, c := range sendable {
			areas = append(areas, declareArea{AreaID: i, FwType: c.code, TypeName: c.o.TypeName, ObjectID: c.o.ID, Vertices: c.verts})
		}
		perRadar[r.deviceAddr] = areas
	}
	return perRadar, skips, drops, nil
}

// radarDownlinkResult 单雷达下发结果（回客户端显示）。
type radarDownlinkResult struct {
	DeviceAddr string `json:"device_addr"`
	DeviceUID  string `json:"device_uid,omitempty"`
	AreaCount  int    `json:"area_count"`
	OK         bool   `json:"ok"`
	DeviceCode int    `json:"device_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

// SaveRoomLayoutResult layout 保存 + declare 下发的结果（warnings 回客户端统一显示，3 端不各算）。
type SaveRoomLayoutResult struct {
	Warnings []string              `json:"warnings,omitempty"`
	Downlink []radarDownlinkResult `json:"downlink,omitempty"`
	Verify   []RadarZonesVerify    `json:"verify,omitempty"` // 下发后 db↔iot 态，FE save 完直接显示（Q2a）
}

// ZoneInfo 单个权威声明区（供 FE status 显示，单源 = data 对 canvas 的编排结果）。
type ZoneInfo struct {
	AreaID   int       `json:"area_id"`
	Type     int       `json:"type"`
	TypeName string    `json:"type_name"`
	ObjectID string    `json:"object_id"`
	Vertices [4][2]int `json:"vertices"`
}

type RadarZones struct {
	DeviceAddr string     `json:"device_addr"`
	Zones      []ZoneInfo `json:"zones"`
}

type RoomZonesResult struct {
	Radars []RadarZones `json:"radars"`
}

// GetRoomZones 读 room_visual_layout → computeDeclareSet → 返回每雷达权威区集（与 data 实际下发一致）。
func (s *RadarInstall) GetRoomZones(ctx context.Context, tenantID, roomID string) (*RoomZonesResult, error) {
	prefix, err := normalizeLayoutPrefix(roomID)
	if err != nil {
		return nil, fmt.Errorf("room_id invalid: %w", err)
	}
	result := &RoomZonesResult{Radars: []RadarZones{}}
	canvas := s.queryRoomVisualLayoutCanvas(ctx, prefix)
	if len(canvas) == 0 {
		return result, nil
	}
	perRadar, _, _, err := computeDeclareSet(canvas)
	if err != nil {
		return nil, err
	}
	for addr, areas := range perRadar {
		rz := RadarZones{DeviceAddr: addr, Zones: []ZoneInfo{}}
		for _, a := range areas {
			v := a.Vertices
			rz.Zones = append(rz.Zones, ZoneInfo{
				AreaID: a.AreaID, Type: a.FwType, TypeName: a.TypeName, ObjectID: a.ObjectID,
				Vertices: [4][2]int{
					{int(v[0].H), int(v[0].V)}, {int(v[1].H), int(v[1].V)},
					{int(v[2].H), int(v[2].V)}, {int(v[3].H), int(v[3].V)},
				},
			})
		}
		result.Radars = append(result.Radars, rz)
	}
	return result, nil
}

// ── block 2：读设备实际 declare + 对意图态 diff（verify，✓/X）──

type deviceZone struct {
	Type int
	V    [4][2]int // dm
}

// parseDeviceDeclareArea 解析设备 declare_area 串 → map[areaID]deviceZone（dm）。
// ⚠️ qinglan READ 已把设备 dm ×10 成 cm（radar_decoder declareAreaDMToCM），故此处坐标 /10 转回 dm，
// 与 intentToDM（cm/10）同量纲，diff/synced/显示才对（写腿仍发 dm，读写不对称在此抹平）。
func parseDeviceDeclareArea(s string) map[int]deviceZone {
	out := map[int]deviceZone{}
	for _, part := range strings.Split(s, "}") {
		part = strings.Trim(part, " ,{\t\r\n")
		if part == "" {
			continue
		}
		fields := strings.Split(part, ",")
		if len(fields) != 10 {
			continue
		}
		nums := [10]int{}
		bad := false
		for i, f := range fields {
			n, err := strconv.Atoi(strings.TrimSpace(f))
			if err != nil {
				bad = true
				break
			}
			nums[i] = n
		}
		if bad {
			continue
		}
		out[nums[0]] = deviceZone{Type: nums[1], V: [4][2]int{ // cm→dm /10（qinglan read 已 ×10）
			{nums[2] / 10, nums[3] / 10}, {nums[4] / 10, nums[5] / 10}, {nums[6] / 10, nums[7] / 10}, {nums[8] / 10, nums[9] / 10},
		}}
	}
	return out
}

func intentToDM(a declareArea) (int, [4][2]int) {
	dm := func(cm float64) int { return int(math.Round(cm / 10)) }
	v := a.Vertices
	return a.FwType, [4][2]int{
		{dm(v[0].H), dm(v[0].V)}, {dm(v[1].H), dm(v[1].V)},
		{dm(v[2].H), dm(v[2].V)}, {dm(v[3].H), dm(v[3].V)},
	}
}

// ZoneVerify 单槽位意图↔设备 diff（synced=两侧都有且类型/坐标一致）。
// db 侧 = 意图(IntentType/Name)；iot 侧 = 设备 RAW dm 原值(DeviceType/DeviceVertices，不转换，查啥显啥)。
type ZoneVerify struct {
	AreaID         int       `json:"area_id"`
	OnIntent       bool      `json:"on_intent"`
	OnDevice       bool      `json:"on_device"`
	Synced         bool      `json:"synced"`
	IntentType     int       `json:"intent_type"`
	IntentTypeName string    `json:"intent_type_name"`
	ObjectID       string    `json:"object_id,omitempty"`
	DeviceType     int       `json:"device_type"`
	DeviceVertices [4][2]int `json:"device_vertices"` // dm RAW（设备原值，不 ×10）
	NotNeed        bool      `json:"not_need,omitempty"` // 引擎用、不下发固件（Chair 等）→ FE 显 "-"
}

type RadarZonesVerify struct {
	DeviceAddr string       `json:"device_addr"`
	DeviceUID  string       `json:"device_uid,omitempty"`
	Reachable  bool         `json:"reachable"`
	Error      string       `json:"error,omitempty"`
	Zones      []ZoneVerify `json:"zones"`
}

type RoomZonesVerifyResult struct {
	Radars []RadarZonesVerify `json:"radars"`
}

// GetRoomZonesVerify 意图态(canvas 编排) ∪ 设备实际态(qinglan 读) 并集逐槽 diff，供 FE 渲染 ✓/X。
func (s *RadarInstall) GetRoomZonesVerify(ctx context.Context, tenantID, roomID string) (*RoomZonesVerifyResult, error) {
	prefix, err := normalizeLayoutPrefix(roomID)
	if err != nil {
		return nil, fmt.Errorf("room_id invalid: %w", err)
	}
	result := &RoomZonesVerifyResult{Radars: []RadarZonesVerify{}}
	canvas := s.queryRoomVisualLayoutCanvas(ctx, prefix)
	if len(canvas) == 0 {
		return result, nil
	}
	intentPerRadar, skips, _, err := computeDeclareSet(canvas)
	if err != nil {
		return nil, err
	}
	for addr, areas := range intentPerRadar {
		rv := RadarZonesVerify{DeviceAddr: addr, Zones: []ZoneVerify{}}
		intentByID := make(map[int]declareArea, len(areas))
		for _, a := range areas {
			intentByID[a.AreaID] = a
		}
		deviceByID := map[int]deviceZone{}
		switch {
		case addr == "":
			rv.Error = "radar object has no device_addr binding"
		case s.qinglanClient == nil:
			rv.Error = "qinglan client not available"
		default:
			uid, uerr := s.GetDeviceUID(ctx, tenantID, addr)
			if uerr != nil {
				rv.Error = fmt.Sprintf("resolve uid: %v", uerr)
				break
			}
			rv.DeviceUID = uid
			props, perr := s.qinglanClient.GetDeviceProperties(ctx, uid, []string{"declare_area"})
			if perr != nil {
				rv.Error = perr.Error()
				break
			}
			rv.Reachable = true
			if da, ok := props["declare_area"].(string); ok {
				deviceByID = parseDeviceDeclareArea(da)
			}
		}
		ids := map[int]bool{}
		for id := range intentByID {
			ids[id] = true
		}
		for id := range deviceByID {
			ids[id] = true
		}
		sortedIDs := make([]int, 0, len(ids))
		for id := range ids {
			sortedIDs = append(sortedIDs, id)
		}
		sort.Ints(sortedIDs)
		for _, id := range sortedIDs {
			ia, onIntent := intentByID[id]
			dz, onDevice := deviceByID[id]
			zv := ZoneVerify{AreaID: id, OnIntent: onIntent, OnDevice: onDevice}
			if onIntent {
				zv.IntentType, zv.IntentTypeName, zv.ObjectID = ia.FwType, ia.TypeName, ia.ObjectID
			}
			if onDevice {
				zv.DeviceType = dz.Type
				zv.DeviceVertices = dz.V // 设备 RAW dm，原值不转换
			}
			if onIntent && onDevice && rv.Reachable {
				it, iv := intentToDM(ia)
				zv.Synced = it == dz.Type && iv == dz.V
			}
			rv.Zones = append(rv.Zones, zv)
		}
		result.Radars = append(result.Radars, rv)
	}
	// B：not-sent 对象（Chair 等引擎用）并进对应雷达，标 not_need → FE 显 "-"。
	for _, sk := range skips {
		zv := ZoneVerify{OnIntent: true, NotNeed: true, IntentType: engineAreaType(sk.TypeName), IntentTypeName: sk.TypeName, ObjectID: sk.ObjectID}
		for i := range result.Radars {
			if result.Radars[i].DeviceAddr == sk.DeviceAddr {
				result.Radars[i].Zones = append(result.Radars[i].Zones, zv)
				break
			}
		}
	}
	return result, nil
}

// AppendFeedbackObject 原子追加 source=Feedback object 进 canvas（jsonb ||，幂等）→ 真追加则触发下发 + config:card。
// sensor feedback pin 单写入口（取代 sensor 直写 room_visual_layout）；pin 即经 data 下发设备。
func (s *RadarInstall) AppendFeedbackObject(ctx context.Context, spatialPrefix, objID, objJSON string) (*SaveRoomLayoutResult, error) {
	prefix, err := normalizeLayoutPrefix(spatialPrefix)
	if err != nil {
		return nil, fmt.Errorf("prefix invalid: %w", err)
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE room_visual_layout
		SET canvas      = jsonb_set(canvas, '{objects}',
		                    COALESCE(canvas->'objects', '[]'::jsonb) || $2::jsonb),
		    canvas_hash = NULL,
		    version     = version + 1,
		    updated_at  = NOW()
		WHERE spatial_prefix = $1::inet
		  AND NOT (COALESCE(canvas->'objects', '[]'::jsonb) @> $3::jsonb)
	`, prefix, objJSON, `[{"id":"`+objID+`"}]`)
	if err != nil {
		return nil, fmt.Errorf("append feedback object: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return &SaveRoomLayoutResult{}, nil // 已存在/无行：幂等，不下发
	}
	canvas := s.queryRoomVisualLayoutCanvas(ctx, prefix)
	if len(canvas) == 0 {
		return &SaveRoomLayoutResult{}, nil
	}
	result := s.applyDeclare(ctx, "", prefix, canvas)
	if s.configPublisher != nil {
		if perr := s.configPublisher.PublishConfigChanged(ctx, "update", []string{prefix}, nil, nil); perr != nil {
			s.logger.Warn("AppendFeedbackObject: publish config:card failed", zap.String("prefix", prefix), zap.Error(perr))
		}
	}
	return result, nil
}

// applyDeclare canvas → 编排 declare set → 下发设备（删旧尾部 + 加新）→ 收集 warnings/下发结果。
// 编排、16-cap、删旧、下发、失败/截断提示全在此一处（单源，3 端只画 canvas + 显示结果）。
func (s *RadarInstall) applyDeclare(ctx context.Context, tenantID, prefix string, newCanvas []byte) *SaveRoomLayoutResult {
	result := &SaveRoomLayoutResult{}
	newPerRadar, skips, drops, err := computeDeclareSet(newCanvas)
	if err != nil {
		s.logger.Warn("declare: canvas parse failed", zap.String("prefix", prefix), zap.Error(err))
		return result
	}
	for addr, areas := range newPerRadar {
		fields := []zap.Field{zap.String("prefix", prefix), zap.String("addr", addr), zap.Int("count", len(areas))}
		for _, a := range areas {
			fields = append(fields, zap.String("area", declareAreaLine(a)))
		}
		s.logger.Info("declare: radar areas", fields...)
	}
	for _, sk := range skips {
		s.logger.Info("declare: skip", zap.String("prefix", prefix), zap.String("typeName", sk.TypeName), zap.String("reason", sk.Reason))
	}
	if len(drops) > 0 {
		for _, dp := range drops {
			s.logger.Warn("declare: DROPPED over-16", zap.String("prefix", prefix), zap.String("typeName", dp.TypeName), zap.Int("type", dp.FwType))
		}
		result.Warnings = append(result.Warnings, fmt.Sprintf("Zone limit reached: %d lower-priority zone(s) were dropped to fit the device's 16-zone limit. Beds, doors, and interfer zones were kept.", len(drops)))
	}
	result.Downlink, result.Verify = s.downlinkDeclares(ctx, tenantID, newPerRadar, computeInstallProps(newCanvas))
	// B：not-sent 对象（Chair 等引擎用）并进对应雷达 verify，标 not_need → FE 显 "-"。
	for _, sk := range skips {
		zv := ZoneVerify{OnIntent: true, NotNeed: true, IntentType: engineAreaType(sk.TypeName), IntentTypeName: sk.TypeName, ObjectID: sk.ObjectID}
		for i := range result.Verify {
			if result.Verify[i].DeviceAddr == sk.DeviceAddr {
				result.Verify[i].Zones = append(result.Verify[i].Zones, zv)
				break
			}
		}
	}
	s.storeBaseline(ctx, prefix, result.Verify) // save 时把设备 baseline(db↔iot 态)存进 layout
	for _, d := range result.Downlink {
		if !d.OK {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Couldn't write zones to device %s: %s", d.DeviceUID, d.Error))
		}
	}
	s.stampCanvasCells(ctx, newCanvas) // ③ stamp 跟随 layout 写入：即时刷 cell（含 FE resize 完整 rect，修 E598）
	return result
}

// buildDeclareDiff Option B：意图(newAreas,id=0..N-1) 对比设备 baseline(deviceByID,dm) → 只发 diff。
// 变化/新增区 ADD/UPDATE；设备有、新集合没有的槽 DELETE(type0,清残留)；未变跳过。返回 declare 串 + 变更数。
func buildDeclareDiff(newAreas []declareArea, deviceByID map[int]deviceZone) (string, int) {
	parts := make([]string, 0, maxDeclareAreasData)
	newIDs := make(map[int]bool, len(newAreas))
	for _, a := range newAreas {
		newIDs[a.AreaID] = true
		it, iv := intentToDM(a)
		if dz, on := deviceByID[a.AreaID]; on && dz.Type == it && dz.V == iv {
			continue // 未变，不发
		}
		parts = append(parts, fmt.Sprintf("{%d,%d,%d,%d,%d,%d,%d,%d,%d,%d}",
			a.AreaID, it,
			iv[0][0], iv[0][1], iv[1][0], iv[1][1], iv[2][0], iv[2][1], iv[3][0], iv[3][1]))
	}
	for id := range deviceByID {
		if !newIDs[id] {
			parts = append(parts, fmt.Sprintf("{%d,0,0,0,0,0,0,0,0,0}", id)) // 残留/移除槽 → 删
		}
	}
	return strings.Join(parts, ","), len(parts)
}

// storeBaseline 把设备 baseline（下发后 db↔iot 态）存进 layout 的 canvas->'device_baseline'（jsonb_set）。
// 不动 objects/canvas_hash/version：device_baseline 是 data 旁路记录设备态，不影响 dedup/编排。
func (s *RadarInstall) storeBaseline(ctx context.Context, prefix string, verify []RadarZonesVerify) {
	if s.db == nil || len(verify) == 0 {
		return
	}
	b, err := json.Marshal(verify)
	if err != nil {
		return
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE room_visual_layout
		SET canvas = jsonb_set(canvas, '{device_baseline}', $2::jsonb, true)
		WHERE spatial_prefix = $1::inet
	`, prefix, string(b)); err != nil {
		s.logger.Warn("storeBaseline failed", zap.String("prefix", prefix), zap.Error(err))
	}
}

func installModelCode(m string) int {
	switch m {
	case "ceiling":
		return 0
	case "corn":
		return 2
	default: // wall
		return 1
	}
}

// boundaryToRectangleDM 移植 FE boundaryToRectangle + cm→dm（/10），出参设备 rectangle 串 {h1,v1;..}。
// 顶点定义按 install mode（ceiling 四角 / wall rearV=0 / corn 顶点+固定左 30）。
func boundaryToRectangleDM(b cvBoundary, installModel string) string {
	dm := func(cm float64) int { return int(math.Round(cm / 10)) }
	var rv [4][2]float64
	switch installModel {
	case "ceiling":
		rv = [4][2]float64{{-b.RightH, -b.RearV}, {b.LeftH, -b.RearV}, {-b.RightH, b.FrontV}, {b.LeftH, b.FrontV}}
	case "corn":
		L, R := b.LeftH, b.RightH
		const leftHFixedCm = 300 // FE leftH_fixed=30(dm)；此处 cm，/10→30dm
		rv = [4][2]float64{{-R, 0}, {leftHFixedCm, 0}, {-R, L}, {leftHFixedCm, L}}
	default: // wall
		rv = [4][2]float64{{-b.RightH, 0}, {b.LeftH, 0}, {-b.RightH, b.FrontV}, {b.LeftH, b.FrontV}}
	}
	parts := make([]string, 4)
	for i := range rv {
		parts[i] = fmt.Sprintf("%d,%d", dm(rv[i][0]), dm(rv[i][1]))
	}
	return "{" + strings.Join(parts, ";") + "}"
}

// computeInstallProps 每雷达安装配置（rectangle/install_height/install_model），dm，随 save 下发。
func computeInstallProps(canvas []byte) map[string]map[string]interface{} {
	var doc cvDoc
	if json.Unmarshal(canvas, &doc) != nil {
		return nil
	}
	out := map[string]map[string]interface{}{}
	for _, o := range doc.Objects {
		r, ok := radarFrameOf(o)
		if !ok || r.deviceAddr == "" {
			continue
		}
		out[r.deviceAddr] = map[string]interface{}{
			"rectangle":            boundaryToRectangleDM(r.boundary, r.installModel),
			"radar_install_height": int(math.Round(r.center.Z / 10)), // cm→dm
			"install_model":        installModelCode(r.installModel),
		}
	}
	return out
}

// buildVerifyZones 下发后 db↔iot 态：成功→设备=意图(全 ✓)；失败→意图对 baseline(差异/残留显 X)。
func buildVerifyZones(areas []declareArea, deviceByID map[int]deviceZone, ok bool) []ZoneVerify {
	zones := make([]ZoneVerify, 0, len(areas))
	intentIDs := map[int]bool{}
	for _, a := range areas {
		intentIDs[a.AreaID] = true
		it, iv := intentToDM(a)
		zv := ZoneVerify{AreaID: a.AreaID, OnIntent: true, IntentType: a.FwType, IntentTypeName: a.TypeName, ObjectID: a.ObjectID}
		if ok {
			zv.OnDevice, zv.Synced, zv.DeviceType, zv.DeviceVertices = true, true, it, iv
		} else if dz, on := deviceByID[a.AreaID]; on {
			zv.OnDevice, zv.DeviceType, zv.DeviceVertices = true, dz.Type, dz.V
			zv.Synced = dz.Type == it && dz.V == iv
		}
		zones = append(zones, zv)
	}
	if !ok { // 失败：设备残留仍在 → X
		for id, dz := range deviceByID {
			if !intentIDs[id] {
				zones = append(zones, ZoneVerify{AreaID: id, OnDevice: true, DeviceType: dz.Type, DeviceVertices: dz.V})
			}
		}
	}
	return zones
}

// downlinkDeclares 每雷达 addr→UID，读 baseline→diff 只发变化+删残留+安装配置→qinglan；返回下发结果 + db↔iot verify。
func (s *RadarInstall) downlinkDeclares(ctx context.Context, tenantID string, newPerRadar map[string][]declareArea, installProps map[string]map[string]interface{}) ([]radarDownlinkResult, []RadarZonesVerify) {
	var results []radarDownlinkResult
	var verifyList []RadarZonesVerify
	if s.qinglanClient == nil {
		for addr, areas := range newPerRadar {
			results = append(results, radarDownlinkResult{DeviceAddr: addr, AreaCount: len(areas), Error: "qinglan client not available"})
		}
		return results, verifyList
	}
	for addr, areas := range newPerRadar {
		res := radarDownlinkResult{DeviceAddr: addr, AreaCount: len(areas)}
		rv := RadarZonesVerify{DeviceAddr: addr, Zones: []ZoneVerify{}}
		if addr == "" {
			res.Error = "radar object has no device_addr binding"
			rv.Error = res.Error
			results = append(results, res)
			verifyList = append(verifyList, rv)
			continue
		}
		uid, err := s.GetDeviceUID(ctx, tenantID, addr)
		if err != nil {
			res.Error = fmt.Sprintf("resolve uid: %v", err)
			rv.Error = res.Error
			results = append(results, res)
			verifyList = append(verifyList, rv)
			continue
		}
		res.DeviceUID = uid
		rv.DeviceUID = uid

		// Option B：读设备 baseline（declare + 安装配置）→ 只发变化；读失败兜底全发。
		deviceByID := map[int]deviceZone{}
		var deviceProps map[string]interface{}
		var declStr string
		var changed int
		if dp, perr := s.qinglanClient.GetDeviceProperties(ctx, uid, []string{"declare_area", "rectangle", "radar_install_height", "install_model"}); perr == nil {
			rv.Reachable = true
			deviceProps = dp
			if da, ok := dp["declare_area"].(string); ok {
				deviceByID = parseDeviceDeclareArea(da)
			}
			declStr, changed = buildDeclareDiff(areas, deviceByID)
		} else {
			declStr = buildDeclareAreaString(areas) // baseline 读失败 → 全发
			changed = len(areas)
		}

		props := map[string]interface{}{}
		if declStr != "" {
			props["declare_area"] = declStr
		}
		for k, v := range installProps[addr] { // rectangle / radar_install_height / install_model：只发变化的
			if deviceProps != nil && fmt.Sprintf("%v", deviceProps[k]) == fmt.Sprintf("%v", v) {
				continue // 与设备一致（如只改 objectName，安装配置没动）→ 不发
			}
			props[k] = v
		}
		if len(props) == 0 { // 区域无变化 + 无安装配置 → 不下发，设备已=意图
			res.OK = true
			rv.Zones = buildVerifyZones(areas, deviceByID, true)
			results = append(results, res)
			verifyList = append(verifyList, rv)
			continue
		}
		code, derr := s.qinglanClient.SetDeviceProperties(ctx, uid, props)
		res.DeviceCode = code
		res.OK = derr == nil
		if derr != nil {
			res.Error = derr.Error()
			s.logger.Warn("declare downlink failed", zap.String("uid", uid), zap.Int("code", code), zap.Error(derr))
		} else {
			s.logger.Info("declare downlink ok (diff)", zap.String("uid", uid), zap.Int("changed", changed), zap.Int("total", len(areas)))
		}
		rv.Zones = buildVerifyZones(areas, deviceByID, res.OK)
		results = append(results, res)
		verifyList = append(verifyList, rv)
	}
	return results, verifyList
}

// buildDeclareAreaString 拼 {id,type,h1,v1,..} 串（dm）：加新 0..N-1 + 删 [N..15]（type0 清所有未用槽）。
// 删 [N..15] 而非仅旧尾部：彻底清掉 data 接管前老 FE/历史写入的残留槽（否则幽灵区永不消失）。
// 顶点 cm→dm（/10）：declare_area 字符串路径 qinglan 不再除（仅 area_{i}_* 路径除），故 data 必须直接发 dm。
func buildDeclareAreaString(areas []declareArea) string {
	dm := func(cm float64) int { return int(math.Round(cm / 10)) }
	parts := make([]string, 0, maxDeclareAreasData)
	for _, a := range areas {
		v := a.Vertices
		parts = append(parts, fmt.Sprintf("{%d,%d,%d,%d,%d,%d,%d,%d,%d,%d}",
			a.AreaID, a.FwType,
			dm(v[0].H), dm(v[0].V), dm(v[1].H), dm(v[1].V),
			dm(v[2].H), dm(v[2].V), dm(v[3].H), dm(v[3].V)))
	}
	for j := len(areas); j < maxDeclareAreasData; j++ {
		parts = append(parts, fmt.Sprintf("{%d,0,0,0,0,0,0,0,0,0}", j))
	}
	return strings.Join(parts, ",")
}

func verticesLine(v [4]radarPt) string {
	b, _ := json.Marshal([4][3]int{
		{int(v[0].H), int(v[0].V), int(v[0].Z)},
		{int(v[1].H), int(v[1].V), int(v[1].Z)},
		{int(v[2].H), int(v[2].V), int(v[2].Z)},
		{int(v[3].H), int(v[3].V), int(v[3].Z)},
	})
	return string(b)
}

func declareAreaLine(a declareArea) string {
	at := observation.AreaType(a.FwType)
	b, _ := json.Marshal(struct {
		ID    int    `json:"id"`
		Type  int    `json:"type"`
		TName string `json:"typeName"`
		AName string `json:"areaName"`
		V     string `json:"v"`
	}{
		ID: a.AreaID, Type: a.FwType, TName: a.TypeName, AName: at.Name(), V: verticesLine(a.Vertices),
	})
	return string(b)
}
