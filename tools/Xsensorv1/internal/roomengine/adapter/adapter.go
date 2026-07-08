package adapter

import (
	"math"

	"owl-common/observation"
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
	TrackID  int     // firmware track_id（透传供 logicID↔track_id 反查：ExitRoom 事件只带 track_id 无坐标，须按号反查丢轨人）
	Online   bool    // 本 tick 在 radar TTL 内有上报
	Pose     int     // observation pose 枚举（6=Lying）
	X, Y, Z  int     // canvas cm
	HR, RR   int     // 0=无信号
	StillSec float64 // 连续静止秒（still-box 总时长）
	FrameMoveCm float64 // 帧间绝对位移 cm（在动/史料不足=999）→ emission 躺姿二义 SBed/SFallen 分配（仅 AreaBed 内）
	AreaType int     // track 当前 cell.Belief[0].Type（CellAreaType 透传）→ emission 正向压制 + floor per-area 阈
	// Chair 区久坐兜底（实时读 px,py 的 cell）→ floor 连续 tFloor 单源（仅 chair 区）
	InChair     bool
	ChairMu     float64 // 14 日久坐均值 AV
	ChairSigma  float64 // 14 日久坐标准差
	ChairMaxSit float64 // false_alarm 反馈棘轮（人工确认安全久坐下限）
	RoomType   int // 房型(card.RoomType: 1=Bathroom)透传 → emission still CDF (μ,σ) room×cell 保守合并
	FwAreaID   int // firmware area_id（地面真值，不随 canvas drift）→ 命中床 areaId = N（在床）→ 驱动 emission SBed
}

// SleepadFrame 单床 sleepad raw 量（§32 二态：设备在线 OR 没有，不建模中途掉线）。
type SleepadFrame struct {
	Present      bool              // 房间有此 sleepad（config-static）→ ρ=1（K^obs）。false=radar-only 床（ρ=0）。
	Reading      belief.BedReading // InBed / LeftBed / NoReport(unknown，首报前；§30 unknown=后验不确定)
	VitalPresent bool              // sleepad 接触 vital(InBed + HR/RR fresh within TTL)→ 活体在垫,抬 SBed
}

// Census 风险因子（risk_evaluator 同源）。
// PeopleCount 与 AloneContinuousMin 均不在此——二者皆 = 真人占用单源派生（§G 排 ghost + S∉{E,L} 含
// blind），由 engine filter 后算出经 BuildRiskContext 注入（规则 #1.3：不留可外部设值的入口层挂钟代理）。
type Census struct {
	Night    bool
	Disabled bool
}

// FrameInput 一帧全部 raw 输入（decoupled；scaffold wire 时由 engine/track_manager 填）。
// Sleepads/Beds/Covers/Onbed/Overlap 长度均 = numBeds。
// Tracks = 本 tick 全部 raw track（§57 步2 归一：每条全量 per-track，既喂 census 数 N_r/排 ghost，
//
//	又各自驱动一份 belief S/B/realness 滤波——§A.3② 隐维复制并排，不进 J 基数）。
type FrameInput struct {
	NowMs    int64
	Tracks   []TrackObs
	Sleepads []SleepadFrame
	Beds     []Rect
	// BedAreaIDs 与 Beds 一一对齐：firmware radar.areas 该床的 area_id（0=无声明区）。
	// track.FwAreaID 命中 → 该床 N=1（在床），驱动 emission SBed boost（替代旧 cell area 判定）。
	BedAreaIDs []int
	Covers     []float64
	Onbed      []float64
	Overlap    []float64
	Census     Census
	// RadarLess 无雷达 layout（sleepad-only 房）：无姿态轴 → 物理上测不了 fall。合成 bed-track 只带 B 轴占用，
	// 其 pose=Lying 是占用载体非真姿态 → 硬闸 Decision.Fire=false，杜绝离床转 lost 时 SFallen 爬出的假 Fall。
	RadarLess bool
	// §9.3① enter 区（门，areaType=4）矩形：出生地距门 D 软发射（无门→D=-1 跳过近门似然）。
	Entrances []Rect
	// ExitLogOdds 按 LogicID 算这条丢轨人"离房"的 SLeft 对数几率（ExitRoom 硬 + trend+np 软；事件无坐标
	//   走不了 census 关联，丢轨后 base 也空，故按身份反查；tm 内用 设备+firmware track_id 还原匹配 ExitRoom）。
	//   喂 blind track 的 logPhi[SLeft]：够强 → 自然 absorbed-drop + 压低 pF 不 fire；≥ExitFlipLogOdds →
	//   Unit D/UD timer cancel。nil=无源→0（保守不抑制）。
	ExitLogOdds func(logicID string, atMs int64) float64
	// GhostLeftLogOdds 按 LogicID 算 present 静止占用轨「已离房」的 SLeft 对数几率（room-empty 信号
	//   + 门距 + 时间门；硬门 Δz/pose 否决真摔）。喂 present 分支 logPhi[SLeft]：镜面残留 ghost 在真人
	//   离房后被划 SLeft → 压 floor 不误火 + 丢轨即 absorbed-drop。nil=无源→0。
	GhostLeftLogOdds func(logicID string, atMs int64) float64
	// HardExited 该 LogicID 是否收到逐帧离房事件(byte-14 event==2)。true → engine.Room 即 hard-drop
	//   该节点（绕过 SLeft 阈，mirror absorbed-drop）+ 回传 EvictTrack。nil=无源→不 drop。
	HardExited func(logicID string, atMs int64) bool
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

// nearBed HR/RR 空间门控：XY 在某床 NearBedMargin 内。per-track（§57 步2）。门控 Online（离线无新鲜 XY）。
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

// nearBedMask 逐床几何邻近（XY 在该床 NearBedMargin 内）。不门控 Online：present 用当前 XY、lost 用冻结坐标
// ——供 FloorGuard 把"接触 InBed 豁免"收窄到本 track 所在床。长 numBeds，与 Sleepad 同索引。
func nearBedMask(t RadarTrack, beds []Rect, p Params) []bool {
	m := make([]bool, len(beds))
	for j, r := range beds {
		m[j] = distCm(t.X, t.Y, r) <= float64(p.NearBedMarginCm)
	}
	return m
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

// bedHitMask N 床判定：track.FwAreaID 命中某床声明的 areaId → 该床 true。
// 0/255（无区/声明区外）不命中。离线 = 全 false（无新鲜帧）。长 = len(bedAreaIDs)（= numBeds）。
func bedHitMask(t RadarTrack, bedAreaIDs []int) []bool {
	m := make([]bool, len(bedAreaIDs))
	if !t.Online || t.FwAreaID == 0 || t.FwAreaID == 255 {
		return m
	}
	for j, id := range bedAreaIDs {
		if id != 0 && id != 255 && t.FwAreaID == id {
			m[j] = true
		}
	}
	return m
}

// BuildObservation raw 帧 → belief.Observation（§5 分轴输入）。per-track（§57 步2）：t 为本 track 的雷达量，
// sleepads/beds 为房共享床证据（每 track 滤波各自吃自己的雷达 + 共享床）。
func BuildObservation(t RadarTrack, sleepads []SleepadFrame, beds []Rect, bedAreaIDs []int, p Params, isRiskTime bool) belief.Observation {
	sl := make([]belief.BedReading, len(sleepads))
	vitalSrc := false
	sleepadVital := false
	for j, s := range sleepads {
		sl[j] = s.Reading
		if s.Present && s.Reading != belief.BedNoReport && s.Reading != belief.BedLeftBedRadar {
			vitalSrc = true // 独立在线 vital 源（§D：radar absent 须 gate）= sleepad 接触有实读数；radar 几何 LeftBed 非接触源,不算
		}
		if s.VitalPresent {
			sleepadVital = true // 任一 sleepad 接触 vital(InBed+HR/RR fresh)→ 活体在垫
		}
	}
	nb := nearBed(t, beds, p)
	return belief.Observation{
		Sleepad:         sl,
		RadarOnline:     t.Online,
		PoseLying:       t.Online && t.Pose == p.PoseLying,
		PoseWalking:     t.Online && (t.Pose == observation.PoseWalking || t.Pose == observation.PoseRunning),
		PoseStanding:    t.Online && t.Pose == observation.PoseStanding,
		PoseSit:         t.Online && t.Pose == observation.PoseSitting,
		PoseFallen:      t.Online && t.Pose == observation.PoseFallen,
		PoseSuspectFall: t.Online && t.Pose == observation.PoseSuspectedFall,
		Z:               t.Z,
		StillSec:        t.StillSec,
		FrameMoveCm:     t.FrameMoveCm,
		NearBed:         nb,
		NearBedMask:     nearBedMask(t, beds, p),
		// HRRRObserved 仅当雷达**真返** HR/RR（铁律 [[radar_hr_rr_bed_enter_gated]]：radar enter-gate，
		// 近床但无 vital = 结构性未测 = 零信息，**非「观测到 absent」**；否则 §D 会在合法在床期误否决 AtBed）。
		HRRRObserved:        t.HR > 0 || t.RR > 0,
		HRRRPresent:         t.HR > 0 || t.RR > 0,
		VitalSourceOnline:   vitalSrc,
		SleepadVitalPresent: sleepadVital,
		AreaType:            t.AreaType,
		InChair:             t.InChair,
		ChairMu:             t.ChairMu,
		ChairSigma:          t.ChairSigma,
		ChairMaxSit:         t.ChairMaxSit,
		RoomType:            t.RoomType,
		IsRiskTime:          isRiskTime, // risktime 只缩短 floor tFloor(纯时间轴),不进 C_FN
		RadarBedHitMask:     bedHitMask(t, bedAreaIDs),
	}
}

// BuildRiskContext Census + N_r + 独居连续分钟 → belief.RiskContext。
// nr 由 engine 经 TrackCensus.Nr() 算出（已排 mirror/伪迹 ghost，§G 主职）；decide 仅在 45–55%
// 两可窗用它折扣 C_FN（拍法 A，§26/§56），≥55% 证据自足不折扣 → N_r=2 真摔照报（pillar D）。
// aloneMin 由 engine filter 后的真人占用 streak 算出（占用==1 连续时长，含 blind 续存的 faller，
// 不在入口层用挂钟凑——F1）。AC-3：alone<0（时钟回拨）守卫落**边界**（规则 #1.4），cFN 内部保持纯形态。
func BuildRiskContext(fi FrameInput, nr int, aloneMin float64) belief.RiskContext {
	if aloneMin < 0 {
		aloneMin = 0
	}
	return belief.RiskContext{
		AloneContinuousMin: aloneMin,
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
