package adapter

import (
	"math"

	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// adapter.go — raw 帧 → belief 输入翻译层（集成里程碑步骤 1）。
// belief 包是已验证黑盒（阶段1-4 B/C 放行）；本层把 raw 雷达/sleepad/床几何 译成
// belief.Observation / BedGeom / RiskContext / BedOnline，scaffold wire 时填 FrameInput。
//
// gxy（g^xy 归属可分性，§4 attachment）：XY 对各床矩形的归属集中度——内→尖峰 / 近(margin 内)→部分 /
//   远→0；近两床时两者都得部分值 = §4「床间均匀」自然涌现。（精度需求中、mixture 容错，C §27 待办下轮改）。
//
// 注（§32 拆补丁）：原 floorStrip（δ floor-strip 运行时派生）已删——cd2b 靠 sleepad LeftBed→B vac 经
// §4 Ψ 相容涌现 SFallen（零补丁，C 独立测试 0.995 验证），不靠雷达 XY 精确空间。floor-strip 是补丁。

// Rect 矩形（canvas cm）：床来自 RoomConfig.Beds[] / engine.deviceBeds[]；墙来自 RoomConfig.Walls[]（桶二镜像几何）。
type Rect struct{ X1, Y1, X2, Y2 int }

// Point 画布坐标点（canvas cm）。雷达自身位置（桶二镜像几何：radar→ghost 连线求交）。
type Point struct{ X, Y int }

// RadarTrack 单 track raw 量（observation.Track 投影）。
type RadarTrack struct {
	Online   bool    // 本 tick 在 radar TTL 内有上报
	Pose     int     // observation pose 枚举（6=Lying）
	X, Y, Z  int     // canvas cm
	HR, RR   int     // 0=无信号
	StillSec float64 // 连续静止秒（still-box 总时长）
	AreaType int     // track 当前 cell.Belief[0].Type（CellAreaType 透传）→ emission 正向压制 + floor per-area 阈
}

// SleepadFrame 单床 sleepad raw 量（§32 二态：设备在线 OR 没有，不建模中途掉线）。
type SleepadFrame struct {
	Present bool              // 房间有此 sleepad（config-static）→ ρ=1（K^obs）。false=radar-only 床（ρ=0）。
	Reading belief.BedReading // InBed / LeftBed / NoReport(unknown，首报前；§30 unknown=后验不确定)
}

// Census 风险因子（risk_evaluator 同源）。
// PeopleCount 不在此——人数 N_r 单源 = TrackCensus.Nr()（§G 排 ghost），由 engine 算后经
// BuildRiskContext 注入（规则 #1.3 单源真相：不留可外部设值的第二计数）。
type Census struct {
	AloneContinuousMin float64
	Night              bool
	Disabled           bool
}

// FrameInput 一帧全部 raw 输入（decoupled；scaffold wire 时由 engine/track_manager 填）。
// Sleepads/Beds/Covers/Onbed/Overlap 长度均 = numBeds。
// Tracks = 本 tick 全部 raw track（§57 步2 归一：每条全量 per-track，既喂 census 数 N_r/排 ghost，
//   又各自驱动一份 belief S/B/realness 滤波——§A.3② 隐维复制并排，不进 J 基数）。
type FrameInput struct {
	NowMs    int64
	Tracks   []TrackObs
	Sleepads []SleepadFrame
	Beds     []Rect
	Covers   []float64
	Onbed    []float64
	Overlap  []float64
	Census   Census
	// 桶二镜像几何（§69/§70）：墙矩形 + 雷达自身位置（房 config-static；无 wall→IsReflection 恒 false=零回归）。
	Walls    []Rect
	RadarPos Point
	// §9.3① enter 区（门，areaType=4）矩形：出生地距门 D 软发射（无门→D=-1 跳过近门似然）。
	Entrances []Rect
}

// Params 派生层参数（form-anchor，标定留 oracle）。
type Params struct {
	PoseLying       int     // observation.PoseLying = 6
	FloorMarginCm   int     // Gxy「近邻」距离阈（XY 在床缘此带内算部分归属）
	NearBedMarginCm int     // HR/RR nearBed 邻域
	PeakInsideGxy   float64 // XY 在床内的 g^xy（尖峰）
	NearGxy         float64 // XY 在 FloorMargin 内的 g^xy（部分；近两床→均匀）
}

// DefaultParams 形态默认（铁律 [[fall_data_is_artificial_test]]：非权威值，留 oracle）。
func DefaultParams() Params {
	return Params{PoseLying: 6, FloorMarginCm: 60, NearBedMarginCm: 100, PeakInsideGxy: 1.0, NearGxy: 0.5}
}

// distCm 点到矩形的平面距离（内部=0）。
func distCm(x, y int, r Rect) float64 {
	dx := maxi(maxi(r.X1-x, x-r.X2), 0)
	dy := maxi(maxi(r.Y1-y, y-r.Y2), 0)
	return math.Hypot(float64(dx), float64(dy))
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Gxy g^xy 归属可分性派生：每床 XY 集中度。内→Peak / margin 内→Near / 远→0。
// 近两床时两者都 Near = §4「床间均匀」。雷达离线 → 全 0（看不见）。per-track（§57 步2）。
func Gxy(t RadarTrack, beds []Rect, p Params) []float64 {
	g := make([]float64, len(beds))
	if !t.Online {
		return g
	}
	for j, r := range beds {
		d := distCm(t.X, t.Y, r)
		switch {
		case d == 0:
			g[j] = p.PeakInsideGxy
		case d <= float64(p.FloorMarginCm):
			g[j] = p.NearGxy
		}
	}
	return g
}

// nearBed HR/RR 空间门控：XY 在某床 NearBedMargin 内。per-track（§57 步2）。
func nearBed(t RadarTrack, beds []Rect, p Params) bool {
	if !t.Online {
		return false
	}
	for _, r := range beds {
		if distCm(t.X, t.Y, r) <= float64(p.NearBedMarginCm) {
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

// Online 每床 ρ = sleepad **存在**（config-static，§32：在线 OR 没有，不建模中途掉线）→ belief.BedOnline。
// 关键：sleepad 存在即 ρ=1（含首报前），不因「还没上报」误判离线——否则 K^unobs 把 B 抽向 vac 造假 SFallen。
func Online(fi FrameInput) belief.BedOnline {
	o := make(belief.BedOnline, len(fi.Sleepads))
	for j, s := range fi.Sleepads {
		o[j] = s.Present
	}
	return o
}

// BuildObservation raw 帧 → belief.Observation（§5 分轴输入）。per-track（§57 步2）：t 为本 track 的雷达量，
// sleepads/beds 为房共享床证据（每 track 滤波各自吃自己的雷达 + 共享床）。
func BuildObservation(t RadarTrack, sleepads []SleepadFrame, beds []Rect, p Params) belief.Observation {
	sl := make([]belief.BedReading, len(sleepads))
	vitalSrc := false
	for j, s := range sleepads {
		sl[j] = s.Reading
		if s.Present && s.Reading != belief.BedNoReport {
			vitalSrc = true // 独立在线 vital 源（§D：radar absent 须 gate 在此下）= 存在且有实读数
		}
	}
	nb := nearBed(t, beds, p)
	return belief.Observation{
		Sleepad:     sl,
		RadarOnline: t.Online,
		PoseLying:   t.Online && t.Pose == p.PoseLying,
		StillSec:    t.StillSec,
		NearBed:     nb,
		ZBand:       zBandOf(t.Z),
		// HRRRObserved 仅当雷达**真返** HR/RR（铁律 [[radar_hr_rr_bed_enter_gated]]：radar enter-gate，
		// 近床但无 vital = 结构性未测 = 零信息，**非「观测到 absent」**；否则 §D 会在合法在床期误否决 AtBed）。
		HRRRObserved:      t.HR > 0 || t.RR > 0,
		HRRRPresent:       t.HR > 0 || t.RR > 0,
		VitalSourceOnline: vitalSrc,
		AreaType:          t.AreaType,
	}
}

// zBandOf 高度 z(cm) → ObsZBand 高度档。**用生产原阈**（wisefido-sensor calibration.go：
//   zUprightCm=80 / zSitMinCm=30）：>80 站 / 30-80 坐 / <30 假低无信息（z=0 贴地→ZNone 中性，
//   z 不是 fall 证据、不否决）。美国马桶座 38-48cm 落在坐档。
func zBandOf(z int) belief.ZBand {
	switch {
	case z > 80:
		return belief.ZStand
	case z >= 30:
		return belief.ZSit
	default:
		return belief.ZNone
	}
}

// BuildRiskContext Census + N_r（人数单源 = TrackCensus.Nr()，§56 候选①）→ belief.RiskContext。
// nr 由 engine 经 TrackCensus.Nr() 算出（已排 mirror/伪迹 ghost，§G 主职）；decide 仅在 45–55%
// 两可窗用它折扣 C_FN（拍法 A，§26/§56），≥55% 证据自足不折扣 → N_r=2 真摔照报（pillar D）。
// AC-3：alone<0（adapter 时钟回拨）守卫落**边界**（规则 #1.4），cFN 内部保持纯形态（B round3 + B/C 共识）。
func BuildRiskContext(fi FrameInput, nr int) belief.RiskContext {
	alone := fi.Census.AloneContinuousMin
	if alone < 0 {
		alone = 0
	}
	return belief.RiskContext{
		AloneContinuousMin: alone,
		Night:              fi.Census.Night,
		PeopleCount:        nr,
		Disabled:           fi.Census.Disabled,
	}
}

func at(s []float64, j int) float64 {
	if j < len(s) {
		return s[j]
	}
	return 0
}
