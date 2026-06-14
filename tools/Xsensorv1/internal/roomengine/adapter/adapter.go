package adapter

import (
	"math"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// adapter.go — raw 帧 → belief 输入翻译层（集成里程碑步骤 1）。
// belief 包是已验证黑盒（阶段1-4 B/C 放行）；本层把 raw 雷达/sleepad/床几何 译成
// belief.Observation / BedGeom / RiskContext / BedOnline，scaffold wire 时填 FrameInput。
//
// 两处运行时派生（集成入口待定点已 retire，feedback-p6A 交接清单）——床区数据源 = 床矩形
// （RoomConfig.Beds[] / engine.deviceBeds[]，canvas cm），代码库**无现成**"床沿地条 / per-XY 可分性"：
//   FloorStripXY（δ 运行时化）：XY 在所有床矩形**外** 但在某床缘 FloorMargin 内 = 床沿地条。
//   gxy（g^xy 归属可分性）：XY 对各床矩形的归属集中度——内→尖峰 / 近(margin 内)→部分 / 远→0；
//     近两床时两者都得部分值 = §4「床间均匀」自然涌现。

// Rect 床矩形（canvas cm），来自 RoomConfig.Beds[] / engine.deviceBeds[]。
type Rect struct{ X1, Y1, X2, Y2 int }

// RadarTrack 单 track raw 量（observation.Track 投影）。
type RadarTrack struct {
	Online   bool    // 本 tick 在 radar TTL 内有上报
	Pose     int     // observation pose 枚举（6=Lying）
	X, Y, Z  int     // canvas cm
	HR, RR   int     // 0=无信号
	StillSec float64 // 连续静止秒（still-box）
}

// SleepadFrame 单床 sleepad raw 量。
type SleepadFrame struct {
	InBed bool // bed_status==InBed
	Fresh bool // 在 sleepad TTL 内（35s）；否则当离线
}

// Census 风险因子（risk_evaluator 同源）。
type Census struct {
	AloneContinuousMin float64
	Night              bool
	PeopleCount        int
	Disabled           bool
}

// FrameInput 一帧全部 raw 输入（decoupled；scaffold wire 时由 engine/track_manager 填）。
// Sleepads/Beds/Covers/Onbed/Overlap 长度均 = numBeds。
type FrameInput struct {
	NowMs    int64
	Track    RadarTrack
	Sleepads []SleepadFrame
	Beds     []Rect
	Covers   []float64
	Onbed    []float64
	Overlap  []float64
	Census   Census
}

// Params 派生层参数（form-anchor，标定留 oracle）。
type Params struct {
	PoseLying       int     // observation.PoseLying = 6
	DownPoses       []int   // 倒/卧姿（δ floor-strip 仅对 down 姿有意义；δ 实验本就 pose=6 only）
	FloorMarginCm   int     // 床缘外 floor-strip 带宽（δ 派生）
	NearBedMarginCm int     // HR/RR nearBed 邻域（≥ FloorMargin）
	PeakInsideGxy   float64 // XY 在床内的 g^xy（尖峰）
	NearGxy         float64 // XY 在 FloorMargin 内的 g^xy（部分；近两床→均匀）
}

// DefaultParams 形态默认（铁律 [[fall_data_is_artificial_test]]：非权威值，留 oracle）。
// DownPoses = Lying(6)/Fallen(5)/SuspectedFall(2)（observation 枚举）。
func DefaultParams() Params {
	return Params{PoseLying: 6, DownPoses: []int{2, 5, 6}, FloorMarginCm: 60, NearBedMarginCm: 100, PeakInsideGxy: 1.0, NearGxy: 0.5}
}

func isDown(pose int, p Params) bool {
	for _, d := range p.DownPoses {
		if pose == d {
			return true
		}
	}
	return false
}

// distCm 点到矩形的平面距离（内部=0）。
func distCm(x, y int, r Rect) float64 {
	dx := maxi(maxi(r.X1-x, x-r.X2), 0)
	dy := maxi(maxi(r.Y1-y, y-r.Y2), 0)
	return math.Hypot(float64(dx), float64(dy))
}

func insideRect(x, y int, r Rect) bool { return distCm(x, y, r) == 0 }

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Gxy g^xy 归属可分性派生：每床 XY 集中度。内→Peak / margin 内→Near / 远→0。
// 近两床时两者都 Near = §4「床间均匀」。雷达离线 → 全 0（看不见）。
func Gxy(fi FrameInput, p Params) []float64 {
	g := make([]float64, len(fi.Beds))
	if !fi.Track.Online {
		return g
	}
	for j, r := range fi.Beds {
		d := distCm(fi.Track.X, fi.Track.Y, r)
		switch {
		case d == 0:
			g[j] = p.PeakInsideGxy
		case d <= float64(p.FloorMarginCm):
			g[j] = p.NearGxy
		}
	}
	return g
}

// floorStrip δ 运行时化：down 姿 + XY 在所有床外 但在某床缘 FloorMargin 内 = 床沿地条。
// down-pose 门控（AC-2 修）：δ 是「卧 pad vs 卧床沿地」判别，δ 实验本就 pose=6 only；
// 走动/站立者在床缘不是「摔在床沿地条」——不门控则走过床边即误判 SFallen（HR-5 实测 +100s 误火根因）。
func floorStrip(fi FrameInput, p Params) bool {
	if !fi.Track.Online || !isDown(fi.Track.Pose, p) {
		return false
	}
	near := false
	for _, r := range fi.Beds {
		if insideRect(fi.Track.X, fi.Track.Y, r) {
			return false // 在床上 = on-pad，非床沿
		}
		if distCm(fi.Track.X, fi.Track.Y, r) <= float64(p.FloorMarginCm) {
			near = true
		}
	}
	return near
}

// nearBed HR/RR 空间门控：XY 在某床 NearBedMargin 内。
func nearBed(fi FrameInput, p Params) bool {
	if !fi.Track.Online {
		return false
	}
	for _, r := range fi.Beds {
		if distCm(fi.Track.X, fi.Track.Y, r) <= float64(p.NearBedMarginCm) {
			return true
		}
	}
	return false
}

// BedGeoms Covers/Onbed/Overlap → []belief.BedGeom。
func BedGeoms(fi FrameInput) []belief.BedGeom {
	n := len(fi.Beds)
	g := make([]belief.BedGeom, n)
	for j := 0; j < n; j++ {
		g[j] = belief.BedGeom{Covers: at(fi.Covers, j), Onbed: at(fi.Onbed, j), Overlap: at(fi.Overlap, j)}
	}
	return g
}

// Online 每床 sleepad 在线（Fresh）→ belief.BedOnline（长 = numBeds，B1 契约）。
func Online(fi FrameInput) belief.BedOnline {
	o := make(belief.BedOnline, len(fi.Sleepads))
	for j, s := range fi.Sleepads {
		o[j] = s.Fresh
	}
	return o
}

// Reading 单床 sleepad → BedReading：Fresh∧InBed=InBed / Fresh∧¬InBed=LeftBed / ¬Fresh=NoReport（离线中性）。
func reading(s SleepadFrame) belief.BedReading {
	if !s.Fresh {
		return belief.BedNoReport
	}
	if s.InBed {
		return belief.BedInBed
	}
	return belief.BedLeftBed
}

// BuildObservation raw 帧 → belief.Observation（§5 分轴输入）。
func BuildObservation(fi FrameInput, p Params) belief.Observation {
	sl := make([]belief.BedReading, len(fi.Sleepads))
	vitalSrc := false
	for j, s := range fi.Sleepads {
		sl[j] = reading(s)
		if s.Fresh {
			vitalSrc = true // 独立在线 vital 源（§D：radar absent 须 gate 在此下）
		}
	}
	nb := nearBed(fi, p)
	return belief.Observation{
		Sleepad:           sl,
		RadarOnline:       fi.Track.Online,
		PoseLying:         fi.Track.Online && fi.Track.Pose == p.PoseLying,
		StillSec:          fi.Track.StillSec,
		NearBed:           nb,
		HRRRObserved:      fi.Track.Online && nb, // radar firmware enter-gate：仅近床返 vital 通道
		HRRRPresent:       fi.Track.HR > 0 || fi.Track.RR > 0,
		VitalSourceOnline: vitalSrc,
		FloorStripXY:      floorStrip(fi, p),
	}
}

// BuildRiskContext Census → belief.RiskContext。
// AC-3：alone<0（adapter 时钟回拨）守卫落**边界**（规则 #1.4），cFN 内部保持纯形态（B round3 + B/C 共识）。
func BuildRiskContext(fi FrameInput) belief.RiskContext {
	alone := fi.Census.AloneContinuousMin
	if alone < 0 {
		alone = 0
	}
	return belief.RiskContext{
		AloneContinuousMin: alone,
		Night:              fi.Census.Night,
		PeopleCount:        fi.Census.PeopleCount,
		Disabled:           fi.Census.Disabled,
	}
}

func at(s []float64, j int) float64 {
	if j < len(s) {
		return s[j]
	}
	return 0
}
