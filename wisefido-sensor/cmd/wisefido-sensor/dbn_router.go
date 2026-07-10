// dbn_router.go — DBN 顶层裁决路由（S0.c-4b）：把 Engine.OnRoomFrame 的 per-room bases
// 构造成 adapter.FrameInput → engine.Unit.Tick（DBN 四轴），返回 (fired, dropped, confidence)。
// 纯裁决无 publish（A·R15.1：publish 在 engine 内 routeRoomFrame 消费返回值）。
// port 自 tools/Xsensorv1/cmd/xsensor（冻结只读参考），生产差异：confidence 第三腿产出（A·R13/R15.3-2）。
package main

import (
	"sync"
	"time"

	"wisefido-sensor/internal/roomengine"
	"wisefido-sensor/internal/roomengine/adapter"
	"wisefido-sensor/internal/roomengine/belief"
	"wisefido-sensor/internal/roomengine/engine"

	"owl-common/card"
	"owl-common/radarutils"

	"go.uber.org/zap"
)

// poseLying observation.PoseLying = 6（sleepad-only 房合成 bed-track 的姿态）。
const poseLying = 6

// roomGeom 一房的 config-static 几何（RegisterRoom 时从 RoomConfig 派生，onRoomFrame 构造 FrameInput 用）。
type roomGeom struct {
	beds           []adapter.Rect
	bedAreaIDs     []int
	walls          []adapter.Rect
	entrances      []adapter.Rect
	radarPos       adapter.Point
	sleepadPresent bool
	radarLess      bool
	nb             int
}

// dbnRouter 每房一份 engine.Room（懒建）；按 suiteID 分组 engine.Unit（跨房 hand-off）。
type dbnRouter struct {
	mu       sync.Mutex
	geom     map[string]*roomGeom
	rooms    map[string]*engine.Room
	units    map[string]*engine.Unit
	roomUnit map[string]string
	roomType map[string]int
	roomTZ   map[string]*time.Location
	mm       map[string]*roomengine.RoomMM
	eng      *roomengine.Engine
	logger   *zap.Logger
}

func newDBNRouter(logger *zap.Logger) *dbnRouter {
	return &dbnRouter{
		geom:     map[string]*roomGeom{},
		rooms:    map[string]*engine.Room{},
		units:    map[string]*engine.Unit{},
		roomUnit: map[string]string{},
		roomType: map[string]int{},
		roomTZ:   map[string]*time.Location{},
		mm:       map[string]*roomengine.RoomMM{},
		logger:   logger,
	}
}

// onRoomFrame = Engine.OnRoomFrame 回调（纯裁决）。返回三腿：
//
//	fired→PublishAIAlarm（engine 内门控发布）/ dropped→emitGhostVerdict / confidence→TrackConfidence 写回。
func (d *dbnRouter) onRoomFrame(roomID string, bases []roomengine.TrackStatusBase, bed card.BedState, nowMs int64, exitLogOdds, ghostLeftLogOdds func(logicID string, atMs int64) float64, hardExited func(logicID string, atMs int64) bool) (fired, dropped []string, confidence map[string]int, firedBands map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	g := d.geom[roomID]
	if g == nil {
		return nil, nil, nil, nil
	}

	// B 轴读数 = room 级权威 bed 状态（sleepad+radar 床事件融合）。
	reading := belief.BedNoReport
	if g.sleepadPresent {
		switch {
		case bed.BedConfidence == 0:
			reading = belief.BedNoReport
		case bed.BedStatus == 1:
			reading = belief.BedLeftBed
			if bed.BedConfidence < 50 {
				reading = belief.BedLeftBedRadar
			}
		default:
			reading = belief.BedInBed
		}
	}

	tracks := make([]adapter.TrackObs, 0, len(bases)+1)
	for _, b := range bases {
		tracks = append(tracks, adapter.TrackObs{LogicID: b.LogicID, SplitConvicted: b.SplitConvicted, ForceReal: b.RealProven || b.SplitProvisionalReal, RadarTrack: adapter.RadarTrack{
			TrackID: b.TrackID,
			Online:  b.Present, Pose: b.Pose, X: b.X, Y: b.Y, Z: b.Z,
			StillSec: float64(b.StillBoxSec),
			FrameMoveCm: float64(b.FrameMoveCm),
			AreaAt: b.AreaAt,
			InChair:  b.InChair, ChairMu: b.ChairMu, ChairSigma: b.ChairSigma, ChairMaxSit: b.ChairMaxSit,
			BathMu:   b.BathMu, BathSigma: b.BathSigma, BathMaxSit: b.BathMaxSit,
			RoomType: d.roomType[roomID],
			FwAreaID: b.FwAreaID,
		}})
	}
	// sleepad-only 房：InBed 合成一条 bed-track 作 B 轴载体（engine.Room track-centric）。
	if g.radarLess && reading == belief.BedInBed {
		fwArea := 0
		if len(g.bedAreaIDs) > 0 {
			fwArea = g.bedAreaIDs[0]
		}
		bedX, bedY := 0, 0
		if len(g.beds) > 0 {
			bedX = (g.beds[0].X1 + g.beds[0].X2) / 2 // 合成载体放床中心 → deviceRadarInBed 几何 InRect 命中
			bedY = (g.beds[0].Y1 + g.beds[0].Y2) / 2
		}
		tracks = append(tracks, adapter.TrackObs{LogicID: "sleepad-bed", RadarTrack: adapter.RadarTrack{
			Online: true, Pose: poseLying, X: bedX, Y: bedY, Z: 0,
			AreaAt:   func(int, int) int { return int(roomengine.AreaBed) }, // 合成 bed 载体在床中心 → 区型=床
			FwAreaID: fwArea,
		}})
	}

	vitalPresent := false
	for _, b := range bases {
		if b.SleepadVitalPresent {
			vitalPresent = true
			break
		}
	}
	sleepads := make([]adapter.SleepadFrame, g.nb)
	for j := 0; j < g.nb; j++ {
		sleepads[j] = adapter.SleepadFrame{Present: g.sleepadPresent, Reading: reading, VitalPresent: vitalPresent}
	}

	isRiskTime := roomengine.IsNightTime(nowMs, d.roomTZ[roomID])
	fi := adapter.FrameInput{
		NowMs:       nowMs,
		Tracks:      tracks,
		Sleepads:    sleepads,
		Beds:        g.beds,
		BedAreaIDs:  g.bedAreaIDs,
		RadarLess:   g.radarLess,
		Entrances:   g.entrances,
		ExitLogOdds:      exitLogOdds,
		GhostLeftLogOdds: ghostLeftLogOdds,
		HardExited:       hardExited,
		Census:           adapter.Census{Night: isRiskTime},
	}

	u := d.units[d.roomUnit[roomID]]
	if u == nil {
		return nil, nil, nil, nil
	}
	fr := u.Tick(roomID, fi)

	// P1 占用人数回注（census 折叠 + uncovered-sleepad 吸纳）→ RealPeopleInRoom（服务 zoneengine total_people）。
	if d.eng != nil {
		pads := d.eng.SnapshotSleepads(roomID, nowMs)
		radars := roomengine.RadarBedStates(bases)
		uncovered, padAbs := roomengine.AbsorbSleepads(pads, radars, d.mm[roomID], roomengine.SamebedAbsorbThresh)
		d.eng.SetRoomRadarPeople(roomID, fr.Decision.PeopleCount+uncovered)
		for _, p := range padAbs {
			if p.RadarLeftBed {
				d.logger.Warn("de_absorption: 主 radar 离床+sleepad 仍 InBed（V5/V6 仲裁=启发式，no-silent-caps）",
					zap.String("room", roomID), zap.String("pad_lid", p.LogicID),
					zap.Bool("fresh", p.Fresh), zap.Bool("stale", p.Stale), zap.Bool("counted_uncovered", p.Uncovered))
			}
		}
	}

	// 第三腿 confidence（A·R13/R15.3-2）：per-track DBN realness PReal → 0-100，写回 cardagg。
	confidence = make(map[string]int, len(fr.Tracks))
	for _, t := range fr.Tracks {
		confidence[t.LogicID] = int(t.PReal*100 + 0.5)
	}

	d.logger.Info("dbn_xray",
		zap.String("room", roomID), zap.Int64("ts", nowMs),
		zap.Int("n_r", fr.Decision.PeopleCount), zap.Int("present_cnt", fr.PresentCount),
		zap.Float64("p_fallen", fr.Probe.PFallen), zap.String("band", fr.Decision.Band),
		zap.Bool("fire", fr.Decision.Fire),
		zap.Int("fired", len(fr.FiredLogicIDs)), zap.Int("dropped", len(fr.DroppedLogicIDs)),
		zap.String("bed_reading", bedReadingName(reading)), zap.Bool("risktime", isRiskTime))

	return fr.FiredLogicIDs, fr.DroppedLogicIDs, confidence, fr.FiredBands
}

// bedReadingName B 轴读数名（log 用）。
func bedReadingName(r belief.BedReading) string {
	switch r {
	case belief.BedInBed:
		return "InBed"
	case belief.BedLeftBed, belief.BedLeftBedRadar:
		return "LeftBed"
	default:
		return "NoReport"
	}
}

// wallsFromPolygon 闭合多边形顶点 → 墙段矩形（adapter.IsReflection 用）。
func wallsFromPolygon(poly []radarutils.Point) []adapter.Rect {
	if len(poly) < 2 {
		return nil
	}
	out := make([]adapter.Rect, 0, len(poly))
	for i := 0; i+1 < len(poly); i++ {
		out = append(out, adapter.Rect{X1: poly[i].X, Y1: poly[i].Y, X2: poly[i+1].X, Y2: poly[i+1].Y})
	}
	if last := poly[len(poly)-1]; last != poly[0] {
		out = append(out, adapter.Rect{X1: last.X, Y1: last.Y, X2: poly[0].X, Y2: poly[0].Y})
	}
	return out
}

// ones 长度 n 全 1（NewRoom 种子 BedGeom Covers/Onbed/Overlap）。
func ones(n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = 1
	}
	return s
}

func rectsFrom(rs []radarutils.Rect) []adapter.Rect {
	out := make([]adapter.Rect, len(rs))
	for i, r := range rs {
		out[i] = adapter.Rect{X1: r.X1, Y1: r.Y1, X2: r.X2, Y2: r.Y2}
	}
	return out
}
