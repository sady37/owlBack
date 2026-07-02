package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"owl-common/layout"
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

// cvObject 全用 owl-common/layout 权威 schema（BE 只读语义视图）；radar 私有配置从 Device.IOT.Radar(raw) 解析。
type cvObject = layout.CanvasObject

// rotationOf 平面旋转角（*float64 缺省 0）。
func rotationOf(o cvObject) float64 {
	if o.Rotation != nil {
		return *o.Rotation
	}
	return 0
}

type cvDoc struct {
	Objects []cvObject `json:"objects"`
}

// radarFrame 一个雷达对象的下发上下文。
type radarFrame struct {
	id           string
	deviceUID    string // 下发/分组 key：设备终身身份
	deviceAddr   string // display 用：设备实测 addr
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

// declareArea 单个区域的下发产物。AreaType=区域语义(判据/显示单源)，FwType=下发固件码(对账)。
type declareArea struct {
	AreaID     int
	FwType     int
	AreaType   observation.AreaType
	ObjectID   string
	ObjectName string
	Vertices   [4]radarPt
}

// declareSkip 落 layout/cell 但不下发固件的对象（策略不下发，如 Chair/BlindArea/Wall）。
type declareSkip struct {
	ObjectID   string
	ObjectName string
	AreaType   observation.AreaType
	Reason     string
	DeviceUID  string // 所属雷达 uid（供 verify 把 not-sent 对象并到对应雷达）
	Vertices   [4]radarPt
}

// declareDrop 本该下发但超 16 槽被优先级挤掉的区域（→ 回客户端截断提示）。
type declareDrop struct {
	ObjectID string
	AreaType observation.AreaType
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

// firmwarePolicy 按 AreaType 决定下发码与是否下发（§2.1 权威）。
// isMonitorBed=Bed∩Radar 仍判定；但固件 declare_area 不接受 MonitorBed(5)（收到会 ACK 却内部按 2 存），
// 故监护床与普通床下发码同为 2，监护床仅引擎语义、不下发 5。
func firmwarePolicy(at observation.AreaType, isMonitorBed bool) (code int, send bool) {
	switch at {
	case observation.AreaBed:
		if isMonitorBed {
			return 2, true
		}
		return 2, true
	case observation.AreaMonitorBed:
		return 2, true
	case observation.AreaEnter:
		return 4, true
	case observation.AreaReflector, observation.AreaInterfer: // 镜/金属/帘/轮椅 → masking(3)
		return 3, true
	case observation.AreaLying, observation.AreaDeny, observation.AreaSit, observation.AreaActive: // 沙发躺/家具/坐区(椅)/盲区遮挡 → custom(1)
		return 1, true
	default: // AreaUnknown：Wall(typeValue -1 归零)与未分类脏对象 —— 唯一非下发类
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
		if rotationOf(o) == 0 {
			return ordered
		}
		cx := (v[0].X + v[3].X) / 2
		cy := (v[0].Y + v[3].Y) / 2
		a := rotationOf(o) * math.Pi / 180
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
	// 雷达对象 = 带 Device.IOT.Radar 的 point；不靠字符串 typeName（对象种类识别单源=IOT.Radar）。
	if o.Device.IOT == nil || len(o.Device.IOT.Radar) == 0 || o.Geometry.Type != "point" {
		return radarFrame{}, false
	}
	var c cvPoint
	if json.Unmarshal(o.Geometry.Data, &c) != nil {
		return radarFrame{}, false
	}
	var rc cvRadar
	if json.Unmarshal(o.Device.IOT.Radar, &rc) != nil {
		return radarFrame{}, false
	}
	im := rc.InstallModel
	if im == "" {
		im = "ceiling"
	}
	return radarFrame{
		id: o.ID, deviceUID: o.Device.IOT.DeviceUID, deviceAddr: o.Device.IOT.DeviceAddr,
		center: c, angle: rotationOf(o), installModel: im, boundary: rc.Boundary,
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
// perRadar / uidAddr 均按 device_uid(设备终身身份)收口：下发认 uid，免疫 device_addr 漂移；uidAddr 仅供 display。
func computeDeclareSet(canvas []byte) (map[string][]declareArea, map[string]string, []declareSkip, []declareDrop, error) {
	var doc cvDoc
	if err := json.Unmarshal(canvas, &doc); err != nil {
		return nil, nil, nil, nil, err
	}
	var radars []radarFrame
	for _, o := range doc.Objects {
		if rf, ok := radarFrameOf(o); ok {
			radars = append(radars, rf)
		}
	}
	perRadar := make(map[string][]declareArea, len(radars))
	uidAddr := make(map[string]string, len(radars))
	var skips []declareSkip
	var drops []declareDrop
	for _, r := range radars {
		uidAddr[r.deviceUID] = r.deviceAddr
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
			at := observation.AreaTypeFrom(o.AreaType)
			isMB := at == observation.AreaBed && bedContainsAnyRadar(o, radars)
			code, send := firmwarePolicy(at, isMB)
			if !send {
				skips = append(skips, declareSkip{ObjectID: o.ID, ObjectName: o.Name, AreaType: at, Reason: "policy:not-sent", DeviceUID: r.deviceUID, Vertices: verts})
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
				drops = append(drops, declareDrop{ObjectID: c.o.ID, AreaType: observation.AreaTypeFrom(c.o.AreaType), FwType: c.code, Vertices: c.verts})
			}
			sendable = sendable[:maxDeclareAreasData]
		}
		// 顺序占槽 0..N-1（确定性）：旧/新均如此编号，删旧=DELETE 尾部 [N..M-1]，无需追踪设备态。
		areas := make([]declareArea, 0, len(sendable))
		for i, c := range sendable {
			areas = append(areas, declareArea{AreaID: i, FwType: c.code, AreaType: observation.AreaTypeFrom(c.o.AreaType), ObjectID: c.o.ID, ObjectName: c.o.Name, Vertices: c.verts})
		}
		perRadar[r.deviceUID] = areas
	}
	return perRadar, uidAddr, skips, drops, nil
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
}

// ZoneInfo 单个权威声明区（供 FE status 显示，单源 = data 对 canvas 的编排结果）。
type ZoneInfo struct {
	AreaID     int       `json:"area_id"`
	Type       int       `json:"type"`
	AreaName   string    `json:"area_name"`
	ObjectID   string    `json:"object_id"`
	ObjectName string    `json:"object_name,omitempty"`
	Vertices   [4][2]int `json:"vertices"`
}

type RadarZones struct {
	DeviceAddr string     `json:"device_addr"`
	DeviceUID  string     `json:"device_uid,omitempty"`
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
	perRadar, uidAddr, _, _, err := computeDeclareSet(canvas)
	if err != nil {
		return nil, err
	}
	for uid, areas := range perRadar {
		rz := RadarZones{DeviceAddr: uidAddr[uid], DeviceUID: uid, Zones: []ZoneInfo{}}
		for _, a := range areas {
			v := a.Vertices
			rz.Zones = append(rz.Zones, ZoneInfo{
				AreaID: a.AreaID, Type: a.FwType, AreaName: a.AreaType.Name(), ObjectID: a.ObjectID, ObjectName: a.ObjectName,
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
	IntentAreaName string    `json:"intent_area_name"`
	ObjectID       string    `json:"object_id,omitempty"`
	ObjectName     string    `json:"object_name,omitempty"`
	DeviceType     int       `json:"device_type"`
	DeviceVertices [4][2]int `json:"device_vertices"` // dm RAW（设备原值，不 ×10）
	NotNeed        bool      `json:"not_need,omitempty"` // 引擎用、不下发固件（Chair 等）→ FE 显 "-"
	Line           string    `json:"line,omitempty"`     // BE 拼好的成品显示行（单源）；FE 直吐不再重造
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
	intentPerRadar, uidAddr, skips, _, err := computeDeclareSet(canvas)
	if err != nil {
		return nil, err
	}
	for uid, areas := range intentPerRadar {
		rv := RadarZonesVerify{DeviceAddr: uidAddr[uid], DeviceUID: uid, Zones: []ZoneVerify{}}
		deviceByID := map[int]deviceZone{}
		switch {
		case uid == "":
			rv.Error = "radar object has no device_uid binding"
		case s.qinglanClient == nil:
			rv.Error = "qinglan client not available"
		default:
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
		rv.Zones = buildVerifyFromDevice(areas, deviceByID, rv.Reachable)
		result.Radars = append(result.Radars, rv)
	}
	// B：not-sent 对象（Chair 等引擎用）并进对应雷达，标 not_need → FE 显 "-"。
	for _, sk := range skips {
		zv := ZoneVerify{OnIntent: true, NotNeed: true, IntentType: int(sk.AreaType), IntentAreaName: sk.AreaType.Name(), ObjectID: sk.ObjectID, ObjectName: sk.ObjectName}
		for i := range result.Radars {
			if result.Radars[i].DeviceUID == sk.DeviceUID {
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
	newPerRadar, uidAddr, skips, drops, err := computeDeclareSet(newCanvas)
	if err != nil {
		s.logger.Warn("declare: canvas parse failed", zap.String("prefix", prefix), zap.Error(err))
		return result
	}
	for uid, areas := range newPerRadar {
		fields := []zap.Field{zap.String("prefix", prefix), zap.String("uid", uid), zap.String("addr", uidAddr[uid]), zap.Int("count", len(areas))}
		for _, a := range areas {
			fields = append(fields, zap.String("area", declareAreaLine(a)))
		}
		s.logger.Info("declare: radar areas", fields...)
	}
	for _, sk := range skips {
		s.logger.Info("declare: skip", zap.String("prefix", prefix), zap.String("areaName", sk.AreaType.Name()), zap.String("reason", sk.Reason))
	}
	if len(drops) > 0 {
		for _, dp := range drops {
			s.logger.Warn("declare: DROPPED over-16", zap.String("prefix", prefix), zap.String("areaName", dp.AreaType.Name()), zap.Int("type", dp.FwType))
		}
		result.Warnings = append(result.Warnings, fmt.Sprintf("Zone limit reached: %d lower-priority zone(s) were dropped to fit the device's 16-zone limit. Beds, doors, and interfer zones were kept.", len(drops)))
	}
	// 下发路径只管下发；verify（含 not-sent 汇总）完全解耦到 GetRoomZonesVerify（FE 主动 query）。
	result.Downlink = s.downlinkDeclares(ctx, newPerRadar, uidAddr, computeInstallProps(newCanvas))
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
		if !ok || r.deviceUID == "" {
			continue
		}
		out[r.deviceUID] = map[string]interface{}{
			"rectangle":            boundaryToRectangleDM(r.boundary, r.installModel),
			"radar_install_height": int(math.Round(r.center.Z / 10)), // cm→dm
			"install_model":        installModelCode(r.installModel),
		}
	}
	return out
}

// buildVerifyFromDevice 意图(areas) ∪ 设备实际态(deviceByID) 逐槽 diff → db↔iot verify（✓/X）。
// 设备真值来自回查（不再乐观假设"下发成功=设备=意图"）。downlinkDeclares 回查腿与 GetRoomZonesVerify 单源共用。
func buildVerifyFromDevice(areas []declareArea, deviceByID map[int]deviceZone, reachable bool) []ZoneVerify {
	intentByID := make(map[int]declareArea, len(areas))
	for _, a := range areas {
		intentByID[a.AreaID] = a
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
	zones := make([]ZoneVerify, 0, len(sortedIDs))
	for _, id := range sortedIDs {
		ia, onIntent := intentByID[id]
		dz, onDevice := deviceByID[id]
		zv := ZoneVerify{AreaID: id, OnIntent: onIntent, OnDevice: onDevice}
		if onIntent {
			zv.IntentType, zv.IntentAreaName, zv.ObjectID, zv.ObjectName = ia.FwType, ia.AreaType.Name(), ia.ObjectID, ia.ObjectName
		}
		if onDevice {
			zv.DeviceType = dz.Type
			zv.DeviceVertices = dz.V
		}
		if onIntent && onDevice && reachable {
			it, iv := intentToDM(ia)
			zv.Synced = it == dz.Type && iv == dz.V
		}
		zv.Line = formatVerifyLine(zv)
		zones = append(zones, zv)
	}
	return zones
}

// formatVerifyLine 成品显示行（单源，FE 直吐不再重造）：Area {id}, {AreaType名}, {对象名} {设备原始declare} iot: ✓/X。
// declare 用设备真值 dm→cm(×10)，与设备上报同量纲；无意图槽名显 "--"，无设备槽 declare 省略。
func formatVerifyLine(z ZoneVerify) string {
	areaName := "--"
	if z.OnIntent {
		areaName = z.IntentAreaName
	}
	line := fmt.Sprintf("Area %d, %s", z.AreaID, areaName)
	if z.ObjectName != "" {
		line += ", " + z.ObjectName
	}
	if z.OnDevice {
		var b strings.Builder
		fmt.Fprintf(&b, "{%d,%d", z.AreaID, z.DeviceType)
		for _, p := range z.DeviceVertices {
			fmt.Fprintf(&b, ",%d,%d", p[0]*10, p[1]*10)
		}
		b.WriteString("}")
		line += " " + b.String()
	}
	mark := "X"
	if z.Synced {
		mark = "✓"
	}
	return line + " iot: " + mark
}

// downlinkDeclares 每雷达 addr→UID，读 baseline→diff 只发变化+删残留+安装配置→qinglan；返回下发结果 + db↔iot verify。
func (s *RadarInstall) downlinkDeclares(ctx context.Context, newPerRadar map[string][]declareArea, uidAddr map[string]string, installProps map[string]map[string]interface{}) []radarDownlinkResult {
	var results []radarDownlinkResult
	if s.qinglanClient == nil {
		for uid, areas := range newPerRadar {
			results = append(results, radarDownlinkResult{DeviceAddr: uidAddr[uid], DeviceUID: uid, AreaCount: len(areas), Error: "qinglan client not available"})
		}
		return results
	}
	// 阶段1：逐雷达下发（读设备 baseline → 只发变化 diff → SetDeviceProperties）。
	// Downlink 结果 = success send to device（指令是否成功发到设备/ACK），不代表固件已生效。
	for uid, areas := range newPerRadar {
		res := radarDownlinkResult{DeviceAddr: uidAddr[uid], DeviceUID: uid, AreaCount: len(areas)}
		if uid == "" {
			res.Error = "radar object has no device_uid binding"
			results = append(results, res)
			continue
		}

		// Option B：读设备 baseline（declare + 安装配置）→ 只发变化；读失败兜底全发。
		deviceByID := map[int]deviceZone{}
		var deviceProps map[string]interface{}
		var declStr string
		var changed int
		if dp, perr := s.qinglanClient.GetDeviceProperties(ctx, uid, []string{"declare_area", "rectangle", "radar_install_height", "install_model"}); perr == nil {
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
		for k, v := range installProps[uid] { // rectangle / radar_install_height / install_model：只发变化的
			if deviceProps != nil && fmt.Sprintf("%v", deviceProps[k]) == fmt.Sprintf("%v", v) {
				continue // 与设备一致（如只改 objectName，安装配置没动）→ 不发
			}
			props[k] = v
		}
		if len(props) == 0 { // 区域无变化 + 无安装配置 → 不下发
			res.OK = true
			results = append(results, res)
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
		results = append(results, res)
	}

	// 下发即返回，不阻塞、不回查。verify 完全解耦到 GetRoomZonesVerify（FE save 后主动 query，隔几秒等固件生效）。
	return results
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
	b, _ := json.Marshal(struct {
		ID    int    `json:"id"`
		Type  int    `json:"type"` // 下发固件码(0-5)
		AName string `json:"areaName"`
		V     string `json:"v"`
	}{
		ID: a.AreaID, Type: a.FwType, AName: a.AreaType.Name(), V: verticesLine(a.Vertices),
	})
	return string(b)
}
