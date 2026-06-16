package roomengine

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	"owl-common/radarutils"
	"owl-common/roomutil"
)

// intPtr observation.Track 的 PositionX/Y/Z 是 *int（区分"没值"vs"=0"）。
func intPtr(v int) *int { return &v }

// TrackOutput Room Engine 对外输出的单条 track 评估结果
type TrackOutput struct {
	TrackID    int
	DeviceAddr string
	RoomID     string
	Verdict    TrackVerdict
	Score      int // [0,100]
	Risk       int // [0,100]
	Anomaly    Anomaly
	X, Y, Z    int // 当前估计位置（画布坐标 cm）
	VX, VY     int // cm/s
	StillSec   int
	Source     string // 当前仅 "radar_direct"（firmware 直发 track）；推断型 fall 已迁出 gate-list → DBN belief shadow
}

// TrackFrame 一帧输入。X/Y/Z 是 engine 层 RadarToCanvas 转换后的画布坐标，供
// 内部 grid/cell/fall 算法使用；RawH/RawV/RawZ 是 firmware 直发的雷达本地坐标，
// **不参与算法**，仅在 alarm/event publish 时作为"parent track" 原样上抛
// （契约：对外 position_x/y/z 永远 == firmware 原始值，跟 monitor_stream 同语义）。
type TrackFrame struct {
	TrackID          int
	DeviceAddr       string
	X, Y, Z          int // 画布坐标 cm（内部算法用）
	RawH, RawV, RawZ int // firmware raw 雷达本地坐标（对外契约用，保持不变）
	Pose             int
	AreaType         int // 雷达给的 area_id（保留兼容字段，engine 不信其判定，用 cell.AreaType）
	TMs              int64

	// 用于 StillBox（静止无移动）检测的辅助字段（每个值是 firmware 给的原始信号）
	// firmware 失锁后这些字段会保持 byte-equal 5+ 分钟，是判 StillBox 静止的强证据
	TrackConfidence int // 0-100
	RemainingTime   int
}

// TrackManager 管理一个房间内所有 track 的生命周期
type TrackManager struct {
	mu      sync.Mutex
	roomID  string
	grid    *RoomGrid
	tracks  map[int]*TrackState
	outputs map[int]*TrackOutput

	// lastRealTrackByDevice：每雷达设备最近一次本房观测到"真 track"的 ms（key=源 device_addr）。
	// 同房多雷达占用对账：lost-fall fire 前若**另一台**雷达近期仍见真人 → 人还在房里被别台看着 → 抑制。
	// 按"批次源设备"(frames[0].DeviceAddr) 记，绕开 tracks[trackID] 在多雷达房的 trackID 碰撞
	// （tracks 按 trackID 索引、ts.DeviceAddr 出生即定不随帧刷新——多雷达同 trackID 会污染 TrackState）。
	lastRealTrackByDevice map[string]int64

	// bedCount：本房结构床数（RegisterRoom 注入 = len(cfg.Beds)）。同房多雷达占用对账**仅在单床房(==1)
	// 启用**：双床房可能两位老人，A 摔(lost)+B 在场被另一台看着会被误压成漏报，故双床房不启用（宁报勿漏）。
	// 固件确认跌倒(pose5)走独立路径、不受此守卫影响。
	bedCount int

	// bedSessions：sleepad 设备维度的"在床会话"状态机；新版 silent fall 触发源。
	// key = sleepad device_uid。详见 BedSession 结构体。
	bedSessions map[string]*BedSession

	// sleepadStates：同房间 sleepad 设备的最新观测（device_uid → obs）
	// 由 ProcessSleepadObservation 写入。
	sleepadStates map[string]*SleepadObservation

	// PR-9: v1 4 fall 计数器全部删除（bathroomRealCount / bedPersonCount / lastLeftBedAt / lastNumberPeopleZeroMs）。
	// Phase C (PR-10/11) 用 SuiteCensus.BathroomCount + SuitePerson.AnchorRoomType + BedSession 重写。

	// PR-Bootstrap: v1 stillFallReportCount 已删除（v1 fire path 同时删除）。
	// PR-10 BathroomFallRules 的统计走 ai.log audit 路径。

	sleepadInBedCount int

	// moveSpeedCms：Kalman 速度阈值（cm/s）。> 此速度的帧即使 pose 不是 Walking 也算 Move。
	// 设计动机：雷达对老人慢走常报 Standing → ActiveType[Move]/TraverseCount 永远不涨。
	// 默认 20（≈2 cells/s）；由 engine.Configure / playback.Run 从 yaml 注入。
	moveSpeedCms int

	// PR-11: 双方一致 InBed 时间窗判定 — sleepad 与 radar 各自最近 InBed 事件 ts。
	// 仅当 |sleepadInBed - radarInBed| ≤ 15s 时，sleepad HR/RR 视作可信，触发 AreaBed cell refresh。
	lastSleepadInBedMs int64
	lastRadarInBedMs   int64
	lastRadarLeftBedMs int64 // 审查㊾:radar LeftBed ts(any-source-OR bed 占用 veto)

	// lastNumberPeople：固件最近一次上报的房内人数(单一 np latch,审查51 #1.3 单源真相)。
	// np=0 ≡ count==0(原 lastNumberPeopleZeroMs subsume 进此,不留两个独立 latch 防 drift)。
	// 喂 belief ObsNumberPeople(弱 corroboration:np=0 弱 Empty 非 substitution / np≥1 压 Empty,R5 不入 SFallen)。
	lastNumberPeople   int   // 最近人数;仅当 lastNumberPeopleTs>0 才有效(0=有效计数,非"未上报")
	lastNumberPeopleTs int64 // 最近 number_people 事件 ts(0=从未上报)
	// noTargetSinceMs：固件最近一次明示"无目标"（track_id=88 心跳 / 全零帧）起始 ts；收到真 track 帧即清 0。
	// 供 88-加速驱逐：持续无目标 ≥ heartbeat88EvictMs 时陈旧 track 不必等满 MaxMissCount
	// （稀疏心跳下要 >120s，见 Case2 142s）即可快速进 lost_fall/vanish。镜像 cardagg ClearDeviceTracks。
	noTargetSinceMs int64
	// 占用账（enter/exit 权威的人数守恒）：房账"空" = lastExitMs > lastEnterMs（只信 ExitRoom，见 roomLedgerEmpty）。
	// lost_fall 入池前查：房账为空 → 失锁 track 是"人走后残影"（冻住的反射）→ 抑制。不靠 ghost，靠 enter/exit。
	lastEnterMs int64
	lastExitMs  int64

	// bedsideFallCfg：R4（床边晕倒）参数；PR-9 v1 R4 触发已删，字段保留供 PR-10/11 BathroomBedsideFall 复用。
	// 全 0 = 用默认（180s / 100cm / 900s）。
	bedsideFallCfg BedsideFallConfig

	// recentRadarAlarms / recentRadarEvents：来自 iot:alarm:stream / iot:event:stream 的 radar 来源记录。
	// 仅"落账"，当前无消费方；未来段 7 (radar fall verify) 会读取做 narrative。
	// 保留窗口 = recentBufferMs（默认 5 min），由 RecordXxx 顺手 evict。
	recentRadarAlarms map[int64]*RadarFallAlarm  // key = TMs
	recentRadarEvents map[int64]*RadarTrackEvent // key = TMs
	recentBufferMs    int64                      // 默认 5 min

	// logger：用于 ai.log 输出 ghost / fall 结构化事件。
	// 默认 zap.NewNop()，engine.Run 会调 SetLogger 注入真 logger。
	logger *zap.Logger

	// timezone：本房间 unit 的 IANA 时区（如 America/Denver），由 engine.RegisterRoom 注入。
	// IsNightTime 调用时传入；nil 时退化为 UTC（错位风险，bootstrap 必须设置）。
	timezone *time.Location

	// roomName：rooms.room_name，由 engine.RegisterRoom 注入。
	// Still fall 触发时与 cell.Belief[0].Type 取并集判 bathroom 语义（见 owl-common/roomutil.ClassifyRoomType）。
	roomName string

	// roomType：rooms.room_type（card.RoomType* 0=Default/1=Bathroom/2=Kitchen），由 engine.RegisterRoom 注入。
	// 权威 bathroom 标志（FE 由 room_name 同义词 ∪ 勾选 bathroom 复选框写库）。
	// lost-fall 等待时长：room_type==Bathroom 即整房按 bathroom 长档（不依赖是否画了 Toilet/Shower 子区）。
	roomType int

	// PR-Bootstrap: v1 stayAlarmEnabled 字段已删除（loadStayAlarmEnablement DB 路径同时删除）。
	// PR-10 BathroomStillFall 用 room.kind=="bathroom" 分支替代"运维显式启用 Stay alarm"语义。

	// interferes：本房间镜面/反射区矩形（cfg.Interferes：mirror、glass-tv、metal、curtain）。
	// 用于因子 7（镜面对称 ghost 检测）：对当前 track 求关于 interfere 长轴的镜像位置，
	// 若另一 Real track 距镜像点 < 50cm → 当前 track 是其镜像 ghost，+60 penalty 直接判 Ghost。
	// 由 engine.RegisterRoom 调 SetInterferes 注入；nil 时因子 7 不参与。
	interferes []radarutils.Rect

	// L1 mirror pair 检测：跨 track 几何对称（无需 layout 镜面坐标先验）。
	// 详 mirror_detect.go。SetRadarMount 由 engine.RegisterRoom 注入。
	radarMount       radarutils.RadarMount
	mirrorBuffer     map[mirrorPairKey]*mirrorPairBuffer
	mirrorCooldownMs int64 // 单 pair 命中后 60s 内不重复 paint + 加 penalty（防同一 pair 持续命中刷分）

	// 静止金属反射体自学习（static_reflector.go）：wallPolygon 判近墙；staticReflectorLastMark 去抖。
	wallPolygon             []radarutils.Point
	staticReflectorLastMark map[int]int64

	// startupMs：TrackManager 创建时间。用于"service 启动 5min grace"反 ghost 兜底
	// （grace 内 first-seen 的 track 视为已存在，birth filter 不打 ghost）。
	startupMs int64

	// uidHexFn 反查 device addr → hex MAC（logicID 出生锚定 + log 双格式用）。engine 注入；
	// nil 时退化空前缀（logicID 仍含 trackID+时戳保唯一）。
	uidHexFn func(deviceAddr string) string
}

// devUIDHex 反查 device addr → hex MAC；nil resolver 返回空字符串。
func (tm *TrackManager) devUIDHex(deviceAddr string) string {
	if tm.uidHexFn == nil {
		return ""
	}
	return tm.uidHexFn(deviceAddr)
}

// BedsideFallConfig R4 床边晕倒参数。
// 物理含义：风险时段（IsNightTime）内 LeftBed 之后 WindowSec 秒内，
// 若 track 距 AreaBed cell ≤ BedsideMarginCm 且静止 > StillTimeoutSec → AnomalyBedsideFall。
type BedsideFallConfig struct {
	WindowSec       int // LeftBed 后多少秒内是开窗期
	BedsideMarginCm int // 距 AreaBed 此值内视为"床边"
	StillTimeoutSec int // 静止此秒数触发
}

// 默认值（与 config.go 中 setRoomEngineDefaults 一致；零值兜底用）
var defaultBedsideFallCfg = BedsideFallConfig{
	WindowSec:       180,
	BedsideMarginCm: 100,
	StillTimeoutSec: 900,
}

// NewTrackManager 创建 track 管理器
func NewTrackManager(roomID string, grid *RoomGrid) *TrackManager {
	return &TrackManager{
		roomID:                  roomID,
		grid:                    grid,
		tracks:                  make(map[int]*TrackState),
		outputs:                 make(map[int]*TrackOutput),
		lastRealTrackByDevice:   make(map[string]int64),
		bedSessions:             make(map[string]*BedSession),
		sleepadStates:           make(map[string]*SleepadObservation),
		moveSpeedCms:            20, // 默认值（与 DefaultLearnParams.MoveSpeedCms 一致）
		bedsideFallCfg:          defaultBedsideFallCfg,
		recentRadarAlarms:       make(map[int64]*RadarFallAlarm),
		recentRadarEvents:       make(map[int64]*RadarTrackEvent),
		recentBufferMs:          5 * 60 * 1000, // 5 min
		logger:                  zap.NewNop(),
		startupMs:               time.Now().UnixMilli(),
		mirrorBuffer:            make(map[mirrorPairKey]*mirrorPairBuffer),
		mirrorCooldownMs:        60_000,
		staticReflectorLastMark: make(map[int]int64),
	}
}

// SetWallPolygon 注入本房间墙多边形（cfg.WallPolygon），静止反射体检测判"近墙"用。
func (tm *TrackManager) SetWallPolygon(poly []radarutils.Point) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.wallPolygon = append([]radarutils.Point(nil), poly...)
}

// SetRadarMount 注入本房间雷达安装坐标（cfg.Radar），用于 L1 mirror pair tiebreaker
// （距 radar 远者 = ghost）+ bounce point 计算。由 engine.RegisterRoom 调用。
func (tm *TrackManager) SetRadarMount(m radarutils.RadarMount) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.radarMount = m
}

// ClearDevice 清空 device 在本 tm 内的所有 in-memory state（"device offline = 内存重启"原则）。
//
// 清：
//   - bedSessions[deviceAddr] — sleepad InBed/LeftBed session（G2 误抑制根因：offline 后
//     stale BedSession 永久挡 BedroomLostFall；清掉后 fall-through 让 lost-fall 正常评估）
//   - sleepadStates[deviceAddr] — sleepad 最新观测（fall verifier sleepadInBed() 读它，
//     stale 让 fall 评分错偏）
//
// 不清：
//   - tracks[trackID] — 按 trackID 索引非 deviceAddr，且 firmware 复用 trackID 常见，
//     自然 evict 走 MaxMissCount timeout 更安全；offline 后 radar 不再喂新帧，stale tracks
//     最多 coast 后自动 evict
//
// 已 fire 入 DB 的 alarm 不丢（DB 单写者）；只清没 fire 的 pending / cache。
func (tm *TrackManager) ClearDevice(deviceAddr string) {
	if tm == nil || deviceAddr == "" {
		return
	}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.bedSessions, deviceAddr)
	delete(tm.sleepadStates, deviceAddr)
}

// SetLogger 注入 zap logger（engine.Run 启动时调用）。
// nil 输入会被替换为 NopLogger（防止后续 nil deref）。
func (tm *TrackManager) SetLogger(l *zap.Logger) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if l == nil {
		tm.logger = zap.NewNop()
	} else {
		tm.logger = l
	}
}

// SetTimezone 注入本房间所在 unit 的 IANA 时区。
// 由 engine.RegisterRoom 调用；nil 表示未配置（IsNightTime 会退化为 UTC）。
func (tm *TrackManager) SetTimezone(loc *time.Location) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.timezone = loc
}

// SetRoomName 注入 rooms.room_name；用于 still fall 触发时按房间名判 bathroom 语义。
// 由 engine.RegisterRoom 调用。
func (tm *TrackManager) SetRoomName(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.roomName = name
}

// SetRoomType 注入 rooms.room_type（card.RoomType*）；lost-fall 等待按 bathroom 长档放宽。
// 由 engine.RegisterRoom 调用。
func (tm *TrackManager) SetRoomType(t int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.roomType = t
}

// SetInterferes 注入本房间镜面/反射区矩形（cfg.Interferes）。
// 由 engine.RegisterRoom 启动时调用。用于因子 7 镜面对称 ghost 检测。
// 内部 deep-copy 避免共享 slice 被外部修改。
func (tm *TrackManager) SetInterferes(rects []radarutils.Rect) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if len(rects) == 0 {
		tm.interferes = nil
		return
	}
	tm.interferes = append([]radarutils.Rect(nil), rects...)
}

// GetInterferes 线程安全返回反射面矩形快照(P1-final 增量2:shadow mirror 对称用,beliefShadowTick 不持 tm.mu)。
func (tm *TrackManager) GetInterferes() []radarutils.Rect {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return append([]radarutils.Rect(nil), tm.interferes...)
}

// SetStartupMs 覆盖默认的 startup 时间戳。playback / 离线测试用：把 startupMs
// 对齐到回放窗口起点（默认是 time.Now()，离线时无意义）。
func (tm *TrackManager) SetStartupMs(ms int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if ms > 0 {
		tm.startupMs = ms
	}
}

// SetUIDHexResolver 注入 device addr → hex MAC 反查（logicID 出生锚定用）。nil = 退化空前缀。
func (tm *TrackManager) SetUIDHexResolver(fn func(deviceAddr string) string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.uidHexFn = fn
}

// makeLogicID 出生时锚定的稳定逻辑身份 = uidlast4 + track_id + mmssms（分秒毫秒，UTC）。
// uidHex 缺失（nil publisher）时退化为空前缀，仍含 track_id+时戳保唯一。
func makeLogicID(uidHex string, trackID int, birthMs int64) string {
	last4 := uidHex
	if len(last4) > 4 {
		last4 = last4[len(last4)-4:]
	}
	t := time.UnixMilli(birthMs).UTC()
	return fmt.Sprintf("%s%d%02d%02d%03d", last4, trackID, t.Minute(), t.Second(), birthMs%1000)
}

// nearestAliveTrack 在当前存活 track 中找离 (x,y)（画布坐标）最近、且已有 logic_id 的一条。
// 用于"无 enter 事件的新 track"继承同一逻辑身份（firmware track_id 重用/跳变/分裂的数据关联）。
// 无候选返回 nil。调用时新 track 尚未加入 tm.tracks，故不会自指。
func (tm *TrackManager) nearestAliveTrack(x, y int) *TrackState {
	var best *TrackState
	bestD := 1 << 30
	for _, ts := range tm.tracks {
		if ts.LogicID == "" || ts.Kalman == nil {
			continue
		}
		px, py := ts.Kalman.Position()
		if d := distInt(x, y, int(math.Round(px)), int(math.Round(py))); d < bestD {
			bestD = d
			best = ts
		}
	}
	return best
}

// hasOtherLiveTrackWithLogicID 是否有"另一条存活 track 持同一 logic_id"。
// 用于 lost_fall 的换ID守恒闸：某 track_id 失锁但其 logic_id 仍活在别的 track_id 上 =
// 同一逻辑目标只是被 firmware 换了 ID（数量守恒，无人真消失）→ 不入 lost-fall pending。
// 不依赖 ghost verdict（铁律：ghost 不进 Fall 决策路径），纯靠身份连续性。
func (tm *TrackManager) hasOtherLiveTrackWithLogicID(logicID string, exceptTrackID int) bool {
	if logicID == "" {
		return false
	}
	for id, ts := range tm.tracks {
		if id != exceptTrackID && ts.LogicID == logicID {
			return true
		}
	}
	return false
}

// HasOtherLiveTrackWithLogicID 自锁版（beliefShadowTick 不持 tm.mu，跨 track range 须锁防并发 map 读写）。
func (tm *TrackManager) HasOtherLiveTrackWithLogicID(logicID string, exceptTrackID int) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.hasOtherLiveTrackWithLogicID(logicID, exceptTrackID)
}

// roomLedgerEmpty 占用账是否为空：最近 ExitRoom 晚于最近 EnterRoom。
// **只信 ExitRoom（过门的空间证据=确实离开），不信 np=0**——铁律：count=0≠离开
// （人摔倒后雷达丢锁也会 np=0，用 np=0 抑制会漏真摔）。默认（无事件）=非空，保守不抑制。
func (tm *TrackManager) roomLedgerEmpty() bool {
	return tm.lastExitMs > tm.lastEnterMs
}

// RoomLedgerEmpty 自锁版（beliefShadowTick 不持 tm.mu；读 lastExit/lastEnterMs 与 RecordRadarEvent 写并发须锁）。
func (tm *TrackManager) RoomLedgerEmpty() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.roomLedgerEmpty()
}

// NeighborRoomEnterMs 跨房 hand-off：最近一次 EnterRoom 的 ts + 当前是否仍占用（未被 ExitRoom 翻掉）。
// 自锁（与 RecordRadarEvent 写 lastEnterMs 持的 tm.mu 一致；仅 beliefShadowTick 跨房读，无重入）。
func (tm *TrackManager) NeighborRoomEnterMs() (enterMs int64, occupied bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.lastEnterMs, tm.lastEnterMs > tm.lastExitMs
}

// NeighborBedHandoff 跨房 hand-off：bed InBed 翻转 ts + 当前是否在床（接触式，BedConfidence>0 才有效）。
func (tm *TrackManager) NeighborBedHandoff(nowMs int64) (statusTs int64, inBed bool, conf float64) {
	bs := tm.BedOccupancyState(nowMs)
	return bs.BedStatusTs, bs.BedStatus == 0 && bs.BedConfidence > 0, float64(bs.BedConfidence) / 100
}

// IsBathroomByRoomName 用 owl-common/roomutil.ClassifyRoomType 判定本房间是否 bathroom。
// 与 cell.Belief[0].Type ∈ {AreaToilet, AreaShower} 取并集驱动 still fall。
func (tm *TrackManager) IsBathroomByRoomName() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return roomutil.IsBathroom(tm.roomName)
}

// ProcessSleepadBedEvent 接收 sleepad InBed/LeftBed 事件。
// 维护 BedSession（InBed/LeftBed 时刻,喂 BedOccupancyState→cardagg bed_state）。
// 由 Engine 路由 iot:event:stream 中 device_type=Sleepad 时调用。
func (tm *TrackManager) ProcessSleepadBedEvent(evt SleepadBedEvent) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if evt.IsInBed {
		// PR-11: 记录 sleepad InBed ts 用于 radar/sleepad 一致期判定
		if evt.TMs > tm.lastSleepadInBedMs {
			tm.lastSleepadInBedMs = evt.TMs
		}
		// BedSession：首次 InBed 启动会话
		s := tm.bedSessions[evt.DeviceUID]
		if s == nil || s.InBedSinceMs == 0 {
			s = &BedSession{DeviceUID: evt.DeviceUID, InBedSinceMs: evt.TMs}
			tm.bedSessions[evt.DeviceUID] = s
		}
		s.LeftBedAtMs = 0 // 任意 InBed → 清掉离床态（视为新一轮上床）
		return
	}

	// LeftBed：如实落 LeftBedAtMs（BedOccupancyState→cardagg bed_state）。
	// 设备开机即在床 / sensor 重启 / 数据裁切等情形常缺前置 InBed 事件,但 LeftBed 本身是 firmware
	// 观测信号且床边离床=高风险 → 从宽认定:无 in-bed 历史也建会话记 LeftBedAtMs(不丢),
	// 由 BedOccupancyState(latestInBed==0 仍认离床)消费。
	s := tm.bedSessions[evt.DeviceUID]
	if s == nil {
		s = &BedSession{DeviceUID: evt.DeviceUID}
		tm.bedSessions[evt.DeviceUID] = s
	}
	s.LeftBedAtMs = evt.TMs
}

// ProcessSleepadObservation 接收 sleepad 一帧观测，按设备 UID 保留最新状态。
// 由 Engine.handleMessage 路由 device_type=Sleepad 时调用。
func (tm *TrackManager) ProcessSleepadObservation(obs SleepadObservation) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cur, ok := tm.sleepadStates[obs.DeviceUID]
	if !ok || obs.TMs > cur.TMs {
		copyObs := obs
		tm.sleepadStates[obs.DeviceUID] = &copyObs
	}
	// 双方一致 InBed ±15s → sleepad HR/RR 可信，把 active radar track 当前 cell 学为 AreaBed
	// （sleepad 不知坐标,但 radar 同报 InBed 且时间一致 → radar 当前 track 位置即床位）。
	if obs.InBed && obs.HasVitalSign() {
		if tm.consistentBedInBed(obs.TMs) {
			for _, ts := range tm.tracks {
				if ts.Verdict != VerdictReal {
					continue
				}
				pxF, pyF := ts.Kalman.Position()
				px, py := int(math.Round(pxF)), int(math.Round(pyF))
				cell := tm.grid.CellAt(px, py)
				if cell == nil {
					continue
				}
				if cell.MarkRestZoneByFeedback(AreaBed) {
					tm.grid.boostNeighborSameType(px, py, AreaBed, obs.TMs)
				}
			}
		}
	}
}

// bedInBedConsistencyMs PR-11/PR-14：sleepad 与 radar InBed 事件视作"双源一致"的最大时差。
// 同时用于 PR-11 HR/RR 即时锁 AreaBed 与 PR-14 BedSession 入场门控。
const bedInBedConsistencyMs = int64(15_000)

// 床态融合参数（2026-06-04，Layer-1）：sleepad 接触式为权威，radar 印证叠加，EMI 兜底地板。
const (
	bedConfSleepad = 90 // 接触式上床基础置信（BedOccupancyState→cardagg bed_state）
	bedConfRadar   = 70 // radar-only InBed 降档幅度（BedOccupancyState）
)

// track 驱逐时间阈（2026-06-04，治 88 稀疏心跳下陈旧 track 拖 142s，Case2）：
//
//	trackEvictMaxMs：纯帧计数 MaxMissCount 之外的时间兜底——上次真观测超此值即驱逐（不受帧稀疏影响）。
//	heartbeat88EvictMs：固件明示无目标(88 心跳)持续 ≥ 此值的加速驱逐窗（镜像 cardagg ClearDeviceTracks 6s 平滑窗）。
const (
	trackEvictMaxMs    = int64(12_000)
	heartbeat88EvictMs = int64(6_000)
)


// ghostJudgable ghost≥2 闸：ghost 是真 track 的镜像/反射，需 ≥2 条 track（至少 1 条作母体）才可判 ghost。
// 仅 1 条 track 时不判 ghost——单 track 判 ghost 会压制从盲区返回的真人/真摔。调用方持锁。
func (tm *TrackManager) ghostJudgable() bool {
	return len(tm.tracks) >= 2
}

// markRadarInBedCell PR-14：radar 报 InBed 事件时，把对应 track 当前位置 cell 锁为 AreaBed。
// 设计：radar 事件本身就是空间证据，不再依赖 sleepad+HR/RR 一致期；
// 双源一致只是 BedSession 触发条件，与"标床区"独立。
//
// 调用方持锁。
func (tm *TrackManager) markRadarInBedCell(e RadarTrackEvent) {
	ts, ok := tm.tracks[e.TrackID]
	if !ok || ts.Verdict == VerdictGhost {
		return
	}
	pxF, pyF := ts.Kalman.Position()
	px, py := int(math.Round(pxF)), int(math.Round(pyF))
	cell := tm.grid.CellAt(px, py)
	if cell == nil {
		return
	}
	if cell.MarkRestZoneByFeedback(AreaBed) {
		tm.grid.boostNeighborSameType(px, py, AreaBed, e.TMs)
	}
}

// consistentBedInBed PR-11：检查 sleepad 与 radar 的 InBed 事件是否在 ±15s 内一致。
// 一致性窗口语义：双方独立确认"床上有人"，sleepad HR/RR 此刻可信，可作为 AreaBed cell ground truth。
// 调用方持锁。
//
// PR-9: 删除 lastLeftBedAt 新鲜度守卫 — delta ≤ bedInBedConsistencyMs(15s) 已隐式保证
// "两个 InBed ts 都是最近的"（陈旧 InBed 与新 InBed 不可能 delta < 15s）。
func (tm *TrackManager) consistentBedInBed(nowMs int64) bool {
	if tm.lastSleepadInBedMs == 0 || tm.lastRadarInBedMs == 0 {
		return false
	}
	delta := tm.lastSleepadInBedMs - tm.lastRadarInBedMs
	if delta < 0 {
		delta = -delta
	}
	return delta <= bedInBedConsistencyMs
}

// sleepadInBed 检查同房间任一 sleepad 在 30s 内报告 InBed（不要求 HR/RR）。
// 设计动机：用户明确"确认在床 = 雷达坐标在床 + sleepad InBed，不要求 HR/RR"
//
//	原因：人坐床上 sleepad HR/RR 信号可能弱；只要 InBed 就是床压传感器侧的存在性证据。
//
// 30s 阈值：sleepad 数据偶有延迟，太老（>30s）就不可信。
func (tm *TrackManager) sleepadInBed(nowMs int64) bool {
	const maxStaleMs = 30_000
	for _, s := range tm.sleepadStates {
		if nowMs-s.TMs > maxStaleMs {
			continue
		}
		if s.InBed {
			return true
		}
	}
	return false
}

// sleepadOffBed 检查同房间任一 sleepad 在 30s 内**有数据**且**显示离床**。
// 与 sleepadInBed 不同：sleepadInBed 看到 InBed=true 即返回；
// sleepadOffBed 要求至少有一条新鲜数据，且没有任何在床信号。
// 用于 bed-fall（雷达坐标在床 + sleepad 离床）检测。
func (tm *TrackManager) sleepadOffBed(nowMs int64) bool {
	const maxStaleMs = 30_000
	hasFresh := false
	for _, s := range tm.sleepadStates {
		if nowMs-s.TMs > maxStaleMs {
			continue
		}
		hasFresh = true
		if s.InBed {
			return false // 任一 sleepad 在床即否决"全离床"
		}
	}
	return hasFresh
}

// BedOccupancyState P5(审查㊾):既有 bed 贝叶斯 Markov(BedSession)+ **any-source-OR LeftBed** → card.BedState。
// 占用 = 最近一次床事件(任一源 sleepad∨radar)是 InBed(BedStatus=0);任一源 LeftBed 更晚 → NotInBed(=1,释放)。
// BedConfidence=90(sleepad 接触式)/30(radar-only 降档)/0(无床数据→bedAdapter Fresh=false 不喂)。
// 供 belief shadow P5:bedAdapter→ObsBedOccupied 占用概率压 SFallen(无 radar-on-bed 要求,R5-clean 接触占用非 pose/z)。
// 自锁(beliefShadowTick 不持 tm.mu)。
func (tm *TrackManager) BedOccupancyState(nowMs int64) card.BedState {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	var latestInBed, latestLeftBed int64
	inBedFromSleepad := false
	for _, s := range tm.bedSessions {
		if s.InBedSinceMs > latestInBed {
			latestInBed, inBedFromSleepad = s.InBedSinceMs, true
		}
		if s.LeftBedAtMs > latestLeftBed {
			latestLeftBed = s.LeftBedAtMs
		}
	}
	if tm.lastRadarInBedMs > latestInBed {
		latestInBed, inBedFromSleepad = tm.lastRadarInBedMs, false
	}
	if tm.lastRadarLeftBedMs > latestLeftBed {
		latestLeftBed = tm.lastRadarLeftBedMs
	}
	if latestInBed == 0 && latestLeftBed == 0 {
		return card.BedState{} // 真无床数据(InBed/LeftBed 都没)→ BedConfidence=0,不喂 shadow
	}
	if latestLeftBed >= latestInBed {
		// 任一源 LeftBed ≥ 最近 InBed(any-source-OR veto)→ 离床 → 不压 → Fall 浮出(漏报-safe,审查㊾)。
		// 含"无在先 InBed"(latestInBed==0):LeftBed 是观测信号 + 床边真摔高风险 → 从宽认定离床,
		// 不依赖在先 InBed(否则 sensor 重启/数据裁切丢了 InBed 就吞掉离床=漏报)。
		return card.BedState{BedStatus: 1, BedStatusTs: latestLeftBed, BedConfidence: bedConfSleepad}
	}
	conf := bedConfSleepad // sleepad 接触式权威=90
	if !inBedFromSleepad {
		conf = bedConfSleepad - bedConfRadar // radar-only InBed 降档
	}
	return card.BedState{BedStatus: 0, BedStatusTs: latestInBed, BedConfidence: conf}
}

// SetMoveSpeedCms 注入"在动"速度阈值。<=0 保留默认。
func (tm *TrackManager) SetMoveSpeedCms(v int) {
	if v <= 0 {
		return
	}
	tm.mu.Lock()
	tm.moveSpeedCms = v
	tm.mu.Unlock()
}

// SetBedsideFallConfig 注入 R4 床边晕倒参数；任一字段 0 保留默认值。
func (tm *TrackManager) SetBedsideFallConfig(c BedsideFallConfig) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if c.WindowSec > 0 {
		tm.bedsideFallCfg.WindowSec = c.WindowSec
	}
	if c.BedsideMarginCm > 0 {
		tm.bedsideFallCfg.BedsideMarginCm = c.BedsideMarginCm
	}
	if c.StillTimeoutSec > 0 {
		tm.bedsideFallCfg.StillTimeoutSec = c.StillTimeoutSec
	}
}

// RecordRadarAlarm 落账 radar firmware Fall/SittingOnGround 进 recentRadarAlarms，
// 供下游 DBN 经 SnapshotTrackStatuses bases 读取 firmware fall 作观测。调用方应紧跟 tm.Tick(a.TMs)。
func (tm *TrackManager) RecordRadarAlarm(a RadarFallAlarm) {
	tm.mu.Lock()
	cp := a
	tm.recentRadarAlarms[a.TMs] = &cp
	tm.evictOldRadarAlarms(a.TMs)
	tm.mu.Unlock()
}

// RecordRadarEvent 落账 radar 来源的事件（EnterRoom/ExitRoom/InBed/LeftBed）。
// 仅落账。PR-14: radar InBed 触发两个副作用：
//  1. 当前 track 位置 cell 锁为 AreaBed（"床区自学"——radar 有事件即给出空间证据）
//  2. 若同房 sleepad InBed 已在 ±15s 内到达，标记 BedSession.RadarInBedConfirmedMs（双源一致门控）
//
// PR-9: 删除 LeftBed → lastLeftBedAt 更新路径（字段已删）。
//   - silent_fall 的 LeftBed latch 仍由 ProcessSleepadBedEvent 维护 BedSession.LeftBedAtMs
//   - PR-14 双源门控的"radar InBed > lastLeftBedAt"新鲜度守卫 → 仅靠 delta ≤ 15s 隐式保证
func (tm *TrackManager) RecordRadarEvent(e RadarTrackEvent) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cp := e
	tm.recentRadarEvents[e.TMs] = &cp
	tm.evictOldRadarEvents(e.TMs)
	if e.EventName == alarm.InBed && e.TMs > tm.lastRadarInBedMs {
		tm.lastRadarInBedMs = e.TMs
		tm.markRadarInBedCell(e) // radar InBed → 当前 track 位置 cell 学为 AreaBed（空间证据）
	}
	// number_people=0：固件「屋内空」断言。只落 ts，不直接驱动任何取消——
	// count=0 不等于人离开（可能盲区/水气丢信号），须与门区空间证据合取才采信（见 bathroom_fall）。
	if e.EventName == EventNameNumberPeople && e.TMs > tm.lastNumberPeopleTs {
		tm.lastNumberPeople = e.NumberPeople // 单 latch:任一 count(含 0)都 latch;np=0 ≡ count==0
		tm.lastNumberPeopleTs = e.TMs
	}
	// radar LeftBed 落 ts(审查㊾ any-source-OR:bed 占用 veto 须 OR 所有源,非只 sleepad → radar 先报 LeftBed 立即释放压制)。
	if e.EventName == alarm.LeftBed && e.TMs > tm.lastRadarLeftBedMs {
		tm.lastRadarLeftBedMs = e.TMs
	}
	// 占用账：EnterRoom→占用，ExitRoom→空（np=0 另在 lastNumberPeopleZeroMs）。np≥1 不计占用（镜面虚增）。
	if e.EventName == alarm.EnterRoom && e.TMs > tm.lastEnterMs {
		tm.lastEnterMs = e.TMs
	}
	if e.EventName == alarm.ExitRoom && e.TMs > tm.lastExitMs {
		tm.lastExitMs = e.TMs
	}
}

// LastNumberPeopleZeroMs 固件最近一次 number_people=0 的 ts（0 = 从未上报/最近 count≠0）。
// 审查51 #1.3:派生视图(非独立 latch)——仅当单 latch 最近 count==0 才返回其 ts。
// 门区 exit 推断的「确认证据」分量；调用方须再校验最后帧门区位置，不可单凭此判离。
func (tm *TrackManager) LastNumberPeopleZeroMs() int64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.lastNumberPeopleTs > 0 && tm.lastNumberPeople == 0 {
		return tm.lastNumberPeopleTs
	}
	return 0
}

// numberPeopleTTLMs number_people event 新鲜窗（firmware 分钟级 push，>1min stale 不再当新鲜）。
const numberPeopleTTLMs = 70_000

// CurrentNumberPeople 单 np latch:最近房内人数 + 是否新鲜(TTL 内)。count=-1/fresh=false=从未上报。
func (tm *TrackManager) CurrentNumberPeople(nowMs int64) (count int, fresh bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.lastNumberPeopleTs == 0 || nowMs-tm.lastNumberPeopleTs > numberPeopleTTLMs {
		return -1, false
	}
	return tm.lastNumberPeople, true
}

// evictOldRadarAlarms / evictOldRadarEvents：删除超出 recentBufferMs 的旧记录。
// 调用方持锁。
func (tm *TrackManager) evictOldRadarAlarms(nowMs int64) {
	cutoff := nowMs - tm.recentBufferMs
	for k := range tm.recentRadarAlarms {
		if k < cutoff {
			delete(tm.recentRadarAlarms, k)
		}
	}
}

func (tm *TrackManager) evictOldRadarEvents(nowMs int64) {
	cutoff := nowMs - tm.recentBufferMs
	for k := range tm.recentRadarEvents {
		if k < cutoff {
			delete(tm.recentRadarEvents, k)
		}
	}
}

// Tick 不带新 frame 的扫描入口，用 ts 作为 nowMs 推进段 2-6。
// 用途：alarm/event 到达时立即触发，让 silent fall pending 检查 / bed-fall 矛盾检测 / R4
// 都用 alarm 的时间戳重新评估（比等下一帧 monitor 更及时，最差能省 0-1s）。
//
// 返回当前所有 track 的输出快照（同 ProcessFrame 语义）。
func (tm *TrackManager) Tick(ts int64) []TrackOutput {
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	return tm.processFrameAt(nil, ts)
}

// NoTargetTick 固件无目标心跳（track_id=88 / 全零帧）专用 tick：标记"无目标起始"后推进。
// 与普通 Tick 区分——alarm/event 到达触发的 Tick 不代表无目标，不应触发 88-加速驱逐。
func (tm *TrackManager) NoTargetTick(ts int64) []TrackOutput {
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	tm.mu.Lock()
	if tm.noTargetSinceMs == 0 {
		tm.noTargetSinceMs = ts
	}
	tm.mu.Unlock()
	return tm.processFrameAt(nil, ts)
}

// SetSleepadInBedCount 外部更新 Sleepad 在床人数
func (tm *TrackManager) SetSleepadInBedCount(count int) {
	tm.mu.Lock()
	tm.sleepadInBedCount = count
	tm.mu.Unlock()
}

// TrackStatusBase 单 track 的 Layer 1 原始投影（不含 PersonID / Zone enrichment）。
// SnapshotTrackStatuses 返回 base 列表，engine 层进一步 enrich 成 TrackStatus 再 publish。
type TrackStatusBase struct {
	TrackID          int
	DeviceAddr       string
	RoomID           string
	Verdict          TrackVerdict
	GhostPenalty     int
	X, Y, Z          int // 画布坐标（grid/cell 算法用；Kalman 输出）
	RawH, RawV, RawZ int // firmware raw 雷达本地坐标 — alarm publish 用，对外契约不变
	Pose             int
	StillSec         int // 逐帧 |Δpos|<15cm 累计；任一帧 ≥15cm 瞬清（脆，质心抖动易断）
	StillBoxSec      int // still-box 时长：30s 滚动 50×50 方框内连续静止的秒数（抗质心抖动，fall 判据用）
	CellAreaType     AreaType
	EnterTarget      string // 当前位置 cell.EnterTarget；非 AreaEnter 时为 ""
	MoveActive       bool   // 本次快照是否"非静止"（StillBoxRunStart==0 OR LastObservedMs == nowMs）
	TraverseDelta    int    // 自上次 SnapshotTrackStatuses 累计的 traverse cells（用于 SuiteCensus 升格判定）
	SleepadInBed     bool   // 同房间最近一帧任一 sleepad InBed 视作 true（resident 强升格判据）
}

// SnapshotTrackStatuses 返回当前所有 live track 的 Layer 1 原始投影。
// 调用时机：engine.handleMessage 调 tm.ProcessFrame 之后；nowMs 用同一帧时间戳。
//
// 锁语义：内部加 tm.mu；返回值是值拷贝，调用方可在锁外读。
// TraverseDelta 通过 track 内部 lastTraverseSnapshotCount 字段差分（首次快照视为 0）。
func (tm *TrackManager) SnapshotTrackStatuses(nowMs int64) []TrackStatusBase {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 同房间 sleepad InBed 状态（任一 sleepad inbed 即 true；多 sleepad 取或）
	sleepadInBed := false
	for _, obs := range tm.sleepadStates {
		if obs != nil && obs.InBed {
			sleepadInBed = true
			break
		}
	}

	out := make([]TrackStatusBase, 0, len(tm.tracks))
	for _, ts := range tm.tracks {
		pxF, pyF := ts.Kalman.Position()
		px := int(math.Round(pxF))
		py := int(math.Round(pyF))

		base := TrackStatusBase{
			TrackID:      ts.TrackID,
			DeviceAddr:   ts.DeviceAddr,
			RoomID:       ts.RoomID,
			Verdict:      ts.Verdict,
			GhostPenalty: ts.GhostPenalty,
			X:            px,
			Y:            py,
			Z:            ts.LastZ,
			RawH:         ts.LastRawH,
			RawV:         ts.LastRawV,
			RawZ:         ts.LastRawZ,
			Pose:         ts.LastPose,
			MoveActive:   ts.StillBoxRunStart == 0 || ts.LastObservedMs == nowMs,
			SleepadInBed: sleepadInBed,
		}
		// StillSec/StillBoxSec：still-box 单源（StillBoxRunStart==0 非静止；否则 box run 秒，30s 滚动 50×50 抗抖动）。
		if ts.StillBoxRunStart > 0 && nowMs > ts.StillBoxRunStart {
			base.StillBoxSec = int((nowMs - ts.StillBoxRunStart) / 1000)
			base.StillSec = base.StillBoxSec
		}
		// LongSurvival / StartupGrace 锚定 → 升格 Anchored verdict（v2 §10.1.1）
		if base.Verdict == VerdictReal && (ts.LongSurvivalAnchored || ts.StartupGrace) {
			base.Verdict = VerdictAnchored
		}
		if c := tm.grid.CellAt(px, py); c != nil {
			base.CellAreaType = c.Belief[0].Type
			if c.Belief[0].Type == AreaEnter {
				base.EnterTarget = c.EnterTarget
			}
		}
		// TraverseDelta：差分本 track 的 LifetimeTraverse 累计（无字段时退化为 0）。
		// PR-3 暂用 FrameCount 近似单调累计；PR-5 (BathroomGate) 接入后换实际 grid.MarkTraverse 计数。
		base.TraverseDelta = 0 // 留待 PR-5 wire 精确 delta；当前阶段 SuiteCensus traverse 升格走 sleepad 强锚 + 时长门
		out = append(out, base)
	}
	return out
}

// ========================================================================
// ProcessFrame：每帧双维度喂（即时流 + 历史流）
// ========================================================================

func (tm *TrackManager) ProcessFrame(frames []TrackFrame) []TrackOutput {
	nowMs := time.Now().UnixMilli()
	if len(frames) > 0 {
		nowMs = frames[0].TMs
	}
	return tm.processFrameAt(frames, nowMs)
}

// processFrameAt 内部入口，nowMs 显式传入，**测试可直接控制时间推进**。
// 无 frames 也可调用（推进 pending 超时检查 + 段 5 bed-fall + 段 6 输出）。
func (tm *TrackManager) processFrameAt(frames []TrackFrame, nowMs int64) []TrackOutput {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if len(frames) > 0 {
		tm.noTargetSinceMs = 0 // 收到真 track 帧 → 清"无目标"标记
	}

	activeIDs := make(map[int]bool)

	// PR-9: v1 bathroomRealCount 入口盘点已删除（caregiver 例外抑制 dropped）。
	// PR-10 BathroomStillFall 将通过 SuiteCensus.BathroomCount + AnchorRoomType 重新引入。

	// ========== 段 1: 观测到的 track ==========
	for _, f := range frames {
		activeIDs[f.TrackID] = true
		ts, exists := tm.tracks[f.TrackID]

		var quality, vx, vy, dtSec int

		if !exists {
			ts = NewTrackState(f.TrackID, f.DeviceAddr, tm.roomID, f.X, f.Y, f.Z, f.TMs)
			// logic_id：有 EnterRoom 配对 = 真新人进门 → 全新身份；无 enter = firmware
			// 重用/跳变/分裂 → 继承最近存活 track 的 logic_id（跨 track_id 数据关联，
			// 让"漂走/重编"的同一逻辑目标保持身份连续，供 ghost/lost-fall 按 logic_id 聚合）。
			if !tm.hasRecentEnterRoom(f.TMs) {
				if parent := tm.nearestAliveTrack(f.X, f.Y); parent != nil {
					ts.LogicID = parent.LogicID
					tm.logger.Info("logic_id_inherited_no_enter",
						zap.String("device_uid", f.DeviceAddr),
						zap.Int("track_id", f.TrackID),
						zap.String("logic_id", parent.LogicID),
						zap.Int("birth_x", f.X), zap.Int("birth_y", f.Y),
					)
				}
			}
			if ts.LogicID == "" {
				ts.LogicID = makeLogicID(tm.devUIDHex(f.DeviceAddr), f.TrackID, f.TMs)
			}

			// PR-5.3 反 ghost: service startup 5min grace 内 first-seen → 默认 Real
			isStartupGrace := tm.startupMs > 0 && (f.TMs-tm.startupMs) < 5*60*1000
			if isStartupGrace {
				ts.StartupGrace = true
				ts.Verdict = VerdictReal
				ts.Score = ScoreConfirmTh
				ts.BirthScore = ScoreConfirmTh
				ts.BirthReason = "startup_grace"
				ts.GhostPenalty = 0
				ts.ConfirmedAtMs = f.TMs
			} else {
				b := tm.birthScore(f.X, f.Y, f.TMs)
				ts.BirthScore = b.score
				ts.BirthReason = b.reason
				ts.Score = ts.BirthScore
				ts.GhostPenalty = b.penalty
			}
			tm.tracks[f.TrackID] = ts
			ts.PrevCore = RadarPoseToCore(f.Pose)
			ts.LastPose = f.Pose
			ts.LastZ = f.Z
			ts.LastRawH = f.RawH
			ts.LastRawV = f.RawV
			ts.LastRawZ = f.RawZ
			quality = ts.Score
			dtSec = 1
		} else {
			// 已有 track
			dt := float64(f.TMs-ts.LastUpdateMs) / 1000.0
			if dt <= 0 {
				dt = 1
			}
			dtSec = int(math.Round(dt))
			if dtSec < 1 {
				dtSec = 1
			}

			ts.Kalman.Predict(dt)
			residualF := ts.Kalman.Update(float64(f.X), float64(f.Y))
			ts.PushPoint(f.X, f.Y, f.Z, f.TMs)
			residual := int(math.Round(residualF))

			// 连续指标（StillBox 静止 + Kalman birth-coherence），在 Kalman update 之后维护
			tm.updateContinuousIndicators(ts, f, nowMs, residualF)

			// 维度 A: 即时流
			tm.scoreResidual(ts, residual)
			tm.scoreMovement(ts, f.X, f.Y, nowMs, f.Pose)
			tm.detectZNoise(ts, f.Z)
			tm.detectPoseMismatch(ts, f.Pose)
			tm.updateLieStateMachine(ts, f.Pose, f.X, f.Y, nowMs)

			quality = ts.Score
			vxF, vyF := ts.Kalman.Velocity()
			vx = int(math.Round(vxF))
			vy = int(math.Round(vyF))

			ts.LastPose = f.Pose
			ts.LastZ = f.Z
			ts.LastRawH = f.RawH
			ts.LastRawV = f.RawV
			ts.LastRawZ = f.RawZ
		}

		// 维度 B: 历史流（每帧无条件）
		tm.grid.MarkOccupancy(f.X, f.Y, quality, vx, vy, nowMs)
		core := RadarPoseToCore(f.Pose)
		// Speed 兜底：雷达对慢走老人常报 Standing → ActiveType[Move] 永远不涨。
		// 仅对 Stand / Unknown 做升格（这两个本来就是"静态站立或不确定"）；
		// Sit / Lie 不动 —— 坐着上半身晃动 / 床上翻身也可能让 Kalman 速度 > 阈值，强升 Move 会污染 Sit/Bed 学习。
		// 阈值由 SetMoveSpeedCms 注入，默认 20 cm/s。
		if core == CorePoseStand || core == CorePoseUnknown {
			speed := math.Hypot(float64(vx), float64(vy))
			if speed > float64(tm.moveSpeedCms) {
				core = CorePoseMove
			}
		}
		tm.grid.MarkPoseTime(f.X, f.Y, core, dtSec, nowMs)

		// Walk 区学习：core==Move 且进入新 cell 时 ++ TraverseCount
		curCol, curRow := tm.grid.ToIndex(f.X, f.Y)
		if core == CorePoseMove && (curCol != ts.LastCellCol || curRow != ts.LastCellRow) {
			tm.grid.MarkTraverse(f.X, f.Y, nowMs)
		}
		ts.LastCellCol = curCol
		ts.LastCellRow = curRow
	}

	// ========== 段 2: 未观测到的 track ==========
	for id, ts := range tm.tracks {
		if activeIDs[id] {
			continue
		}
		dt := float64(nowMs-ts.LastUpdateMs) / 1000.0
		if dt <= 0 {
			dt = 1
		}
		ts.Kalman.PredictOnly(dt)
		ts.LastUpdateMs = nowMs

		// 驱逐 = 帧计数(MaxMissCount) ∪ 时间兜底(trackEvictMaxMs) ∪ 88-加速(固件持续无目标≥6s)。
		// 治"88 稀疏心跳下 MissCount 凑满 10 要 142s"（Case2）：LastObservedMs 是上次真观测，按真实时间判，
		// 不受帧稀疏影响。88-加速：固件明示无目标且持续 ≥ heartbeat88EvictMs 时，6s 即驱逐进 lost_fall/vanish。
		unseenMs := nowMs - ts.LastObservedMs
		noTargetSustained := tm.noTargetSinceMs > 0 && nowMs-tm.noTargetSinceMs >= heartbeat88EvictMs
		if ts.Kalman.MissCount > MaxMissCount ||
			unseenMs > trackEvictMaxMs ||
			(noTargetSustained && unseenMs >= heartbeat88EvictMs) {
			// track 消失：lost-fall 判定已迁出 gate-list（DBN belief shadow lost-track 路径）。
			// 此处仅保留 PathBreak——非门区 Real track 突然消失的非 fall 异常标记。
			if ts.Verdict == VerdictReal {
				pxF, pyF := ts.Kalman.Position()
				px := int(math.Round(pxF))
				py := int(math.Round(pyF))
				if !tm.grid.IsEdge(px, py, 50) {
					ts.CurrentAnomaly = AnomalyPathBreak
				}
			}
			delete(tm.tracks, id)
			continue
		}
	}

	// ========== 段 3: 试用期判定 ==========
	for _, ts := range tm.tracks {
		// PR-5.3 反 ghost：track 存活 ≥ 5min 锚定 Real（不论之前 verdict）
		if !ts.LongSurvivalAnchored && ts.AgeSec() >= 300 {
			ts.LongSurvivalAnchored = true
			if ts.Verdict != VerdictReal {
				ts.Verdict = VerdictReal
				ts.ConfirmedAtMs = nowMs
				ts.BirthReason = "long_survival_anchor"
			}
			continue
		}

		if ts.Verdict != VerdictPending {
			// PR-5.x 持续期 ghost factor：每帧重算 lifetime 因子（30s 静止等）
			tm.applyLifetimeGhostFactors(ts, nowMs)
			// 即使已是 Real verdict，penalty ≥ 80 也翻 Ghost（除非 LongSurvival 锚定）。
			// ghost≥2 闸（2026-06-05）：ghost = 真 track 的镜像/反射，无第二条 track 作母体不判 ghost
			// （单 track 判 ghost 会压制从盲区返回的真人/真摔，见 [[longsurvival_anchor_ghost_gap]]）。
			if ts.GhostPenalty >= GhostPenaltyThreshold && !ts.LongSurvivalAnchored && !ts.StartupGrace && tm.ghostJudgable() {
				if ts.Verdict != VerdictGhost {
					ts.Verdict = VerdictGhost
				}
			}
			continue
		}
		// Birth grace recompute（方案 A）：deadline 到了再扫扩展窗，给 EnterRoom event-stream 缓冲
		tm.tryGraceUpgrade(ts, nowMs)

		// 持续期 ghost 因子（30s 静止等）
		tm.applyLifetimeGhostFactors(ts, nowMs)

		// PR-5.x 主路径：累积 GhostPenalty ≥ 80 → Ghost（独立于 ProbationFrames）
		// ghost≥2 闸：单 track 不判 ghost（无母体可镜像；防压制盲区返回的真人/真摔）。
		if ts.GhostPenalty >= GhostPenaltyThreshold && tm.ghostJudgable() {
			ts.Verdict = VerdictGhost
			reason := ts.BirthReason
			if reason == "" {
				reason = "ghost_penalty_accumulated"
			}
			if !ts.LoggedGhost {
				pxF, pyF := ts.Kalman.Position()
				tm.logger.Info("ghost_veto",
					zap.String("reason", "penalty_accumulated"),
					zap.String("birth_reason", reason),
					zap.String("device_uid", ts.DeviceAddr),
					zap.Int("track_id", ts.TrackID),
					zap.String("verdict", "ghost"),
					zap.Int("score", ts.Score),
					zap.Int("birth_score", ts.BirthScore),
					zap.Int("ghost_penalty", ts.GhostPenalty),
					zap.Int("x", int(math.Round(pxF))),
					zap.Int("y", int(math.Round(pyF))),
					zap.Int64("ts_ms", nowMs),
				)
				ts.LoggedGhost = true
			}
			continue
		}

		if ts.FrameCount >= ProbationFrames {
			if ts.Score >= ScoreConfirmTh {
				ts.Verdict = VerdictReal
				ts.ConfirmedAtMs = nowMs
			} else if ts.Score < ScoreGhostTh {
				ts.Verdict = VerdictGhost
				reason := ts.BirthReason
				if reason == "" {
					reason = "low_score"
				}
				if !ts.LoggedGhost {
					pxF, pyF := ts.Kalman.Position()
					tm.logger.Info("ghost_veto",
						zap.String("reason", "low_score"),
						zap.String("birth_reason", reason),
						zap.String("device_uid", ts.DeviceAddr),
						zap.Int("track_id", ts.TrackID),
						zap.String("verdict", "ghost"),
						zap.Int("score", ts.Score),
						zap.Int("birth_score", ts.BirthScore),
						zap.Int("x", int(math.Round(pxF))),
						zap.Int("y", int(math.Round(pyF))),
						zap.Int64("ts_ms", nowMs),
					)
					ts.LoggedGhost = true
				}
			}
		}
	}

	// 推断型 fall（lost / silent-leftbed / z-drop）已迁出 gate-list → DBN belief shadow。
	results := make([]TrackOutput, 0, len(tm.tracks))

	// 同房多雷达占用对账：记录本批次源设备本帧是否见到真 track（按批次设备记，绕 trackID 碰撞）。
	if len(frames) > 0 {
		batchDev := frames[0].DeviceAddr
		for id := range activeIDs {
			if t := tm.tracks[id]; t != nil && t.Verdict == VerdictReal {
				tm.lastRealTrackByDevice[batchDev] = nowMs
				break
			}
		}
	}

	// ========== 段 4d: L1 mirror pair 检测 + 自学习 ==========
	// 跨 track 几何对称检测（无 layout 镜面坐标先验），命中 → ghost 端 GhostPenalty +50
	// + 5 帧 bounce 点 grid.MarkMirrorBounce（2×2 微块累 MBC，≥3 升 AreaDeny+SourceLearned）。
	tm.scanMirrorGhostPairs(nowMs)

	// ========== 段 4e: 静止金属反射体自学习（near-wall + 长期静止 + 游走真人共存）==========
	// Phase A：仅累计 StaticReflectorCount + log static_reflector_candidate，不改 verdict。详 static_reflector.go。
	tm.scanStaticReflectors(nowMs)

	// PR-9: v1 段 5 Bed-Fall 物理矛盾检测整段删除（依赖 totalBedPeople / bedPersonCount）。
	// PR-11 silent_fall 重写时用 BedSession + SuiteCensus 重新表达"床上方矛盾"语义，
	// 不再走 per-bed-device 人数累计路径（多人床的"陪同"由 SuitePerson AnchorRoomType 表达）。

	// ========== 段 6: 构建输出 ==========
	for _, ts := range tm.tracks {
		pxF, pyF := ts.Kalman.Position()
		vxF, vyF := ts.Kalman.Velocity()
		px := int(math.Round(pxF))
		py := int(math.Round(pyF))
		vx := int(math.Round(vxF))
		vy := int(math.Round(vyF))

		stillSec := 0 // still-box 单源（投影显示 TrackOutput.StillSec；computeRisk 不再吃此死参数）
		if ts.StillBoxRunStart > 0 {
			stillSec = int((nowMs - ts.StillBoxRunStart) / 1000)
		}

		source := "radar_direct"

		out := TrackOutput{
			TrackID:    ts.TrackID,
			DeviceAddr: ts.DeviceAddr,
			RoomID:     ts.RoomID,
			Verdict:    ts.Verdict,
			Score:      ts.Score,
			Risk:       tm.computeRisk(ts, nowMs),
			Anomaly:    ts.CurrentAnomaly,
			X:          px,
			Y:          py,
			Z:          ts.LastZ,
			VX:         vx,
			VY:         vy,
			StillSec:   stillSec,
			Source:     source,
		}
		results = append(results, out)
		tm.outputs[ts.TrackID] = &out
	}

	return results
}

// GetOutputs 返回当前所有 track 的最新输出
func (tm *TrackManager) GetOutputs() []TrackOutput {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	out := make([]TrackOutput, 0, len(tm.outputs))
	for _, o := range tm.outputs {
		out = append(out, *o)
	}
	return out
}

// PR-Bootstrap: StillFallStats + stillFallReportCount 已删除（v1 fire path 删除后无 source）。
// PR-10 BathroomStillFall 的统计走 ai.log audit 路径，不再 per-TrackManager 计数。


// ========================================================================
// 出生打分（PR-5.1+5.2+5.3 重写：累积 GhostPenalty 模式）
// ========================================================================
//
// 设计：用户 2026-04-29 对齐
//   主路径：累积 GhostPenalty，> 80 判 Ghost；阈值高、必须多重证据合击。
//   保留 ts.Score 双轨（baseline 50 + 加分）作向后兼容；不再 hard cut 0/100。
//
// 因子（出生瞬时）：
//   1. dEntry / EnterRoom 速度反推    最多 -60
//      - 配对窗 ±3s（实证 92% coverage）
//      - 无配对 EnterRoom: -60
//      - D=0（室外但 InFOV）: 0
//      - speed >= 100 cm/s: -60；50-100 线性；< 50: 0
//   2. 出生 cell 类型                  -10
//      - AreaDeny / AreaSit / AreaBed / AreaToilet / AreaShower: -10
//   4. 已有 ≥1 个 verdict=Real track：    -10  (后出现的 ID)
//   6. cell.GhostRatio > 0.3：              -10  (历史 ghost 多发区)
//
// 因子 3 (30s 静止) / 因子 5 (运动对称性) 在 lifetime 期间累积。
//
// 反 ghost 兜底（在 verdict 转换段处理）：
//   - track 存活 ≥ 5min 锚定 Real
//   - service startup 5min grace 内 first-seen 默认 Real

const (
	enterPairWindowMs     = 3_000 // EnterRoom 与 birth 的最大时间差（ms）
	birthMaxRealisticCm   = 150   // 距 Enter 此值内才可能是 1 秒走入的真人
	birthEnterPairBonus   = 20    // 有 EnterRoom 配对加分
	GhostPenaltyThreshold = 80    // 累积 ghost penalty 阈值
)

// birthScoreResult birthScore 计算结果。
// score: baseline 50 + 加分（用于 ts.Score 向后兼容，影响 ProbationFrames 升降）
// penalty: ghost 累积扣分（写入 ts.GhostPenalty），≥ 80 在 verdict 转换段判 Ghost
// reason: 主导信号（用于 ai.log）
type birthScoreResult struct {
	score   int
	penalty int
	reason  string
}

func (tm *TrackManager) birthScore(x, y int, tMs int64) birthScoreResult {
	cell := tm.grid.CellAt(x, y)
	score := 50
	penalty := 0
	reason := ""

	// ===== 因子 1: dEntry / EnterRoom 速度反推（最多 -60）=====
	dEntry := tm.grid.NearestEntryDist(x, y)
	isOutdoorButInFOV := cell == nil || !cell.InRoom
	switch {
	case isOutdoorButInFOV:
		// D=0 边界规则：在室外但仍 InFOV 的 birth 不扣
	default:
		enterMs, enterFound := tm.nearestEnterRoomMs(tMs, enterPairWindowMs)
		if !enterFound {
			penalty += 60
			reason = "no_enter_pair"
		} else {
			T := math.Abs(float64(tMs-enterMs)) / 1000.0
			if T < 1.0 {
				T = 1.0 // 防除零；T=0 视为 1s
			}
			speed := float64(dEntry) / T
			switch {
			case speed >= 100:
				penalty += 60
				reason = "birth_speed_impossible"
			case speed >= 50:
				p := int(math.Round(60 * (speed - 50) / 50))
				penalty += p
				if reason == "" {
					reason = "birth_speed_high"
				}
			}
		}
	}

	// ===== 因子 2: 出生 cell 类型（-10）=====
	if cell != nil {
		switch cell.Belief[0].Type {
		case AreaDeny:
			penalty += 10
			if reason == "" {
				reason = "born_in_deny"
			}
		case AreaSit, AreaBed, AreaToilet, AreaShower:
			penalty += 10
			if reason == "" {
				reason = "born_in_rest_or_shower"
			}
		}
	}

	// ===== 因子 4: 后出现的 ID（已有 verdict=Real 的 track）-10 =====
	nReal := 0
	for _, t := range tm.tracks {
		if t.Verdict == VerdictReal {
			nReal++
		}
	}
	if nReal >= 1 {
		penalty += 10
		if reason == "" {
			reason = "later_born_with_real_track"
		}
	}

	// ===== 因子 6: 历史 ghost 多发区（cell.GhostRatio > 0.3）-10 =====
	if cell != nil && cell.GhostRatio() > 0.3 {
		penalty += 10
		if reason == "" {
			reason = "ghost_history_zone"
		}
	}

	// ===== Score baseline 加分（保持 verdict 旧路径可走 Real）=====
	if dEntry < 50 {
		score += 20
	} else if dEntry <= birthMaxRealisticCm && tm.hasRecentEnterRoom(tMs) {
		score += birthEnterPairBonus
	}
	if cell != nil && cell.IsEntry() {
		score += 15
	}

	return birthScoreResult{
		score:   clampInt(score, 0, 100),
		penalty: penalty,
		reason:  reason,
	}
}

// applyLifetimeGhostFactors 在 track 生命周期内增量累积 ghost penalty。
//
// 因子覆盖：
//   - 因子 3（30s 内位移很少 → 凭空出现又不动）：每 track 一次性
//   - 因子 5/7（对称性 ghost：紧贴运动 OR 关于镜面几何对称）：并列 peer，
//     仅在 GhostPenalty ∈ [70, 80) 边缘时启用；任一命中 +10 跨阈值判 Ghost。
//
// 调用方持锁；因子 3 幂等（LifetimeFactorsApplied bitmask 防重复扣）；
// 因子 5/7 每帧检查一次，命中即跨阈值。
func (tm *TrackManager) applyLifetimeGhostFactors(ts *TrackState, nowMs int64) {
	const (
		factor3Bit          = 1 << 0
		factor3CheckAgeMs   = 30_000 // 出生后 30s 检查一次
		factor3DispThreshCm = 50     // 位移 < 50cm 视为"凭空又不动"
		factor3Penalty      = 10
	)
	// 因子 3: 30s 内位移很少
	if ts.LifetimeFactorsApplied&factor3Bit == 0 &&
		ts.LastUpdateMs-ts.BirthPos.TMs >= factor3CheckAgeMs {
		px, py := ts.Kalman.Position()
		dx := int(math.Round(px)) - ts.BirthPos.X
		dy := int(math.Round(py)) - ts.BirthPos.Y
		disp := int(math.Sqrt(float64(dx*dx + dy*dy)))
		if disp < factor3DispThreshCm {
			ts.GhostPenalty += factor3Penalty
			if ts.BirthReason == "" || ts.BirthReason == "low_score" {
				ts.BirthReason = "static_after_birth"
			}
		}
		ts.LifetimeFactorsApplied |= factor3Bit
	}

	// 因子 5/7: 对称性 ghost 检测（peer 因子，并列结构）
	// 仅当 GhostPenalty 已在 [70, 80) 边缘时启用 — 此时任一命中 +10 = 80 跨阈值。
	// 之外不计算（避免每帧扫所有 track 对 / interferes，开销不必要）。
	//   - 因子 5: 双 track 紧贴（<100cm）+ 同向移动 → 多径反射 ghost
	//   - 因子 7: 当前 track 关于 cfg.Interferes 长轴镜像 ≤50cm 处有 Real → 镜面 ghost
	if ts.GhostPenalty >= 70 && ts.GhostPenalty < GhostPenaltyThreshold {
		if tm.checkMotionSymmetry(ts, nowMs) {
			ts.GhostPenalty += 10
			ts.BirthReason = "motion_symmetric_with_real_track"
			tm.logger.Info("ghost_veto",
				zap.String("reason", "motion_symmetry"),
				zap.String("device_uid", ts.DeviceAddr),
				zap.Int("track_id", ts.TrackID),
				zap.Int("ghost_penalty", ts.GhostPenalty),
				zap.Int64("ts_ms", nowMs),
			)
		} else if partner, mx, my, rx, ry, ok := tm.checkMirrorSymmetry(ts); ok {
			ts.GhostPenalty += 10
			ts.BirthReason = "mirror_image_of_real_track"
			bxF, byF := ts.Kalman.Position()
			tm.logger.Info("ghost_veto",
				zap.String("reason", "mirror_symmetry"),
				zap.String("device_uid", ts.DeviceAddr),
				zap.Int("track_id", ts.TrackID),
				zap.Int("partner_track_id", partner),
				zap.Int("ghost_penalty", ts.GhostPenalty),
				zap.Int("track_x", int(math.Round(bxF))), zap.Int("track_y", int(math.Round(byF))),
				zap.Int("reflected_x", rx), zap.Int("reflected_y", ry),
				zap.Int("partner_x", mx), zap.Int("partner_y", my),
				zap.Int64("ts_ms", nowMs),
			)
		}
	}
}

// checkMirrorSymmetry 因子 7：镜面对称 ghost 检测（peer to motion_symmetry）。
//
// 触发判定：
//  1. 房间存在另一 verdict=Real 的 track（partner）
//  2. 对每个 cfg.Interferes 矩形 m，求当前 track 关于 m 长轴中线的镜像 (rx, ry)
//  3. 任一 partner 距 (rx, ry) ≤ 50cm → 当前 track 是镜像 ghost
//
// 物理意义：镜子/玻璃/金属反射面把人毫米波信号镜像到镜面对侧，雷达把这个二次反射
// 当成"另一个人"在镜后等距处。motion_symmetry 抓"两 track 紧贴同向"的多径反射；
// 镜面对称抓"两 track 关于固定物理几何镜像"的反射，互为补充。
//
// 返回 (partnerTrackID, partnerX, partnerY, reflectedX, reflectedY, hit)。调用方持锁。
func (tm *TrackManager) checkMirrorSymmetry(ts *TrackState) (int, int, int, int, int, bool) {
	const mirrorDistCm = 50
	if len(tm.interferes) == 0 {
		return 0, 0, 0, 0, 0, false
	}
	bxF, byF := ts.Kalman.Position()
	bx, by := int(math.Round(bxF)), int(math.Round(byF))
	for _, m := range tm.interferes {
		rx, ry := reflectAcrossMirror(bx, by, m)
		for tid, t := range tm.tracks {
			if tid == ts.TrackID || t.Verdict != VerdictReal {
				continue
			}
			axF, ayF := t.Kalman.Position()
			ax, ay := int(math.Round(axF)), int(math.Round(ayF))
			if distInt(ax, ay, rx, ry) <= mirrorDistCm {
				return tid, ax, ay, rx, ry, true
			}
		}
	}
	return 0, 0, 0, 0, 0, false
}

// reflectAcrossMirror 把点 (px, py) 关于矩形 m 的"长轴中线"做镜像。
//
//	长轴 = AABB 较长那一对边（width >= height 时为水平镜，否则垂直镜）。
//	物理意义：mirror 通常薄长，AABB 长边方向 ≈ 镜面方向；中线 ≈ 镜面物理表面。
//	薄镜（depth ~5-10cm）下中线与表面差几 cm，对 50cm 距离阈值无影响。
//
// 数学：
//
//	水平镜 axis Y = (Y1+Y2)/2 → 反射 (x, 2*cy - y)
//	垂直镜 axis X = (X1+X2)/2 → 反射 (2*cx - x, y)
//
// 注意：当前只处理 AABB（layout 解析后的镜面均轴对齐）；旋转镜面将来 PR 再加。
func reflectAcrossMirror(px, py int, m radarutils.Rect) (int, int) {
	w := m.X2 - m.X1
	h := m.Y2 - m.Y1
	if w >= h {
		cy := (m.Y1 + m.Y2) / 2
		return px, 2*cy - py
	}
	cx := (m.X1 + m.X2) / 2
	return 2*cx - px, py
}

// checkMotionSymmetry PR-5.4 因子 5：双 track 运动对称性检测。
//
// 触发判定：
//  1. 房间存在另一 verdict=Real 的 track（partner）
//  2. 当前位置与 partner 距离 < 100cm（紧贴 → 反射 ghost 经典模式）
//  3. 取 2s 前两 track 的位置 → 计算位移向量 (dx, dy)
//  4. 两向量长度都 ≥ 10cm（避免静止 noise 触发）
//  5. 夹角 < 30°（cos > 0.866） → 同步移动 → 强 ghost 信号
//
// 返回 true 表示命中（应扣 -10 跨阈值）。调用方持锁。
func (tm *TrackManager) checkMotionSymmetry(ts *TrackState, nowMs int64) bool {
	const (
		coexistDistMaxCm = 100          // 双 track 距离上限
		windowMs         = int64(2_000) // 2s 滚动窗
		minDispCm        = 10           // 单 track 最小位移（< 10 视为静止 noise）
		cosThreshold     = 0.866        // cos(30°)
	)
	// 找另一个共存 Real track
	var partner *TrackState
	for tid, t := range tm.tracks {
		if tid == ts.TrackID {
			continue
		}
		if t.Verdict != VerdictReal {
			continue
		}
		partner = t
		break
	}
	if partner == nil {
		return false
	}
	// 当前位置 + 2s 前位置
	pxA, pyA := ts.Kalman.Position()
	pxB, pyB := partner.Kalman.Position()
	axNow, ayNow := int(math.Round(pxA)), int(math.Round(pyA))
	bxNow, byNow := int(math.Round(pxB)), int(math.Round(pyB))

	// 距离 < 100cm
	if distInt(axNow, ayNow, bxNow, byNow) > coexistDistMaxCm {
		return false
	}

	// 2s 前位置（从 History 取最接近的早期点）
	pa0, okA := positionAtMsAgo(ts, nowMs, windowMs)
	pb0, okB := positionAtMsAgo(partner, nowMs, windowMs)
	if !okA || !okB {
		return false
	}

	// 位移向量
	dxA, dyA := axNow-pa0.X, ayNow-pa0.Y
	dxB, dyB := bxNow-pb0.X, byNow-pb0.Y
	magA := math.Sqrt(float64(dxA*dxA + dyA*dyA))
	magB := math.Sqrt(float64(dxB*dxB + dyB*dyB))
	if magA < float64(minDispCm) || magB < float64(minDispCm) {
		return false // 至少有一方没动 → 不算同步运动
	}

	// 夹角余弦
	dot := float64(dxA*dxB + dyA*dyB)
	cosTheta := dot / (magA * magB)
	return cosTheta > cosThreshold
}

// positionAtMsAgo 从 ts.History 取最接近 (nowMs - windowMs) 的位置。
// 返回 (TimedPoint, found)；ok=false 表示 history 为空或全比目标新。
func positionAtMsAgo(ts *TrackState, nowMs int64, windowMs int64) (TimedPoint, bool) {
	if len(ts.History) == 0 {
		return TimedPoint{}, false
	}
	target := nowMs - windowMs
	// 倒序找第一个 TMs <= target 的点
	for i := len(ts.History) - 1; i >= 0; i-- {
		if ts.History[i].TMs <= target {
			return ts.History[i], true
		}
	}
	// 都比 target 新 → 用最早的（不严谨但作为 fallback）
	return ts.History[0], true
}

// updateRegionStatic PR-13：region static 累积器 + A/B 路径触发判定。
//
// 区域定义：连续帧间 |dx|≤15 AND |dy|≤15（D523 实测 ~96% 帧满足；single-axis 跳变 ≤5%）。
// 容忍机制：累积期内允许 ≤10% 帧打断（≥90% 静止）；解决 PR-7.2 严格累积被噪声打断 (4e-12 概率) 的问题。
// 触发条件 OR：
//
//	A 路径（z 突变 = 坐→站）：region static ≥2min AND |z - RegionStartZ| ≥10cm AND 双方 z>0
//	B 路径（持续累积）：region static ≥ threshold （RestZone cell 8min；其它 12min）AND ratio ≥0.90
//
// 触发后双 cell 加分：prev cell + cur cell 同时 MarkRestZoneFeedback(AreaSit)。
// per-track 一次性（AreaSitLearnedRegion flag）。
//
// 调用方持锁；nowMs/cell 已由 scoreMovement 计算。
func (tm *TrackManager) updateRegionStatic(ts *TrackState, prev TimedPoint, x, y int, nowMs int64, cell *Cell) {
	const (
		regionDxDyMaxCm  = 15 // |dx|≤15 AND |dy|≤15 = 同区域
		regionResetCm    = 50 // |dx|>50 或 |dy|>50 = 大跨步，立即 reset
		ratioMin         = 0.90
		ratioResetMin    = 0.85 // ratio 跌破此值 → region 失效 reset
		zJumpMinCm       = 10
		zJumpMinElapse   = 2 * 60 * 1000  // 2 min
		thresholdRest    = 8 * 60 * 1000  // RestZone cell: 8min
		thresholdNonRest = 12 * 60 * 1000 // 其它: 12min
	)

	dx := x - prev.X
	if dx < 0 {
		dx = -dx
	}
	dy := y - prev.Y
	if dy < 0 {
		dy = -dy
	}

	// 大跨步：立即 reset region
	if dx > regionResetCm || dy > regionResetCm {
		ts.RegionStaticStartedMs = 0
		ts.RegionStaticFrames = 0
		ts.RegionTotalFrames = 0
		return
	}

	// 维护 region 帧计数
	staticFrame := dx <= regionDxDyMaxCm && dy <= regionDxDyMaxCm
	if ts.RegionStaticStartedMs == 0 {
		// 启动 region；要求当前是 static 帧
		if !staticFrame {
			return
		}
		ts.RegionStaticStartedMs = nowMs
		ts.RegionStartZ = prev.Z
		ts.RegionStaticFrames = 1
		ts.RegionTotalFrames = 1
		return
	}

	ts.RegionTotalFrames++
	if staticFrame {
		ts.RegionStaticFrames++
	}

	// ratio 跌破 0.85 → 区域失效，reset
	if ts.RegionTotalFrames >= 10 {
		ratio := float64(ts.RegionStaticFrames) / float64(ts.RegionTotalFrames)
		if ratio < ratioResetMin {
			ts.RegionStaticStartedMs = 0
			ts.RegionStaticFrames = 0
			ts.RegionTotalFrames = 0
			return
		}
	}

	// 已学习过 → 不重复触发
	if ts.AreaSitLearnedRegion {
		return
	}

	// PR-13: 与 PR-7.2 一致 — still-fall 触发场景（toilet/shower/bathroom-room/Stay-alarm）
	// 不学 AreaSit；让 still-fall 处理。否则浴室长时间静止会被误学成坐位。
	if cell != nil && tm.stillFallTimeoutSec(cell, false) > 0 {
		return
	}

	elapsed := nowMs - ts.RegionStaticStartedMs

	// A 路径：region static ≥2min + z 突变 ≥10cm（双方 z 有效）
	if elapsed >= int64(zJumpMinElapse) && ts.RegionStartZ > 0 && prev.Z > 0 {
		dz := prev.Z - ts.RegionStartZ
		if dz < 0 {
			dz = -dz
		}
		if dz >= zJumpMinCm {
			tm.markBothCellsAreaSit(ts, prev, x, y, nowMs, "zjump")
			ts.AreaSitLearnedRegion = true
			return
		}
	}

	// B 路径：累积时长 + ratio ≥0.90
	threshold := int64(thresholdNonRest)
	if cell != nil && cell.IsRestZone() {
		threshold = int64(thresholdRest)
	}
	if elapsed >= threshold {
		ratio := float64(ts.RegionStaticFrames) / float64(ts.RegionTotalFrames)
		if ratio >= ratioMin {
			tm.markBothCellsAreaSit(ts, prev, x, y, nowMs, "static_accum")
			ts.AreaSitLearnedRegion = true
		}
	}
}

// markBothCellsAreaSit PR-13：双 cell 加分。prev cell + cur cell 同时 MarkRestZoneFeedback(AreaSit)。
// 内部各自调 boostNeighborSameType（PR-10）；cell1↔cell2 邻居时互相 boost cap 100。
func (tm *TrackManager) markBothCellsAreaSit(ts *TrackState, prev TimedPoint, x, y int, nowMs int64, trigger string) {
	tm.grid.MarkRestZoneFeedback(prev.X, prev.Y, AreaSit, nowMs)
	tm.grid.MarkRestZoneFeedback(x, y, AreaSit, nowMs)
	tm.logger.Info("area_sit_auto_learned_region",
		zap.String("device_uid", ts.DeviceAddr),
		zap.Int("track_id", ts.TrackID),
		zap.String("trigger", trigger),
		zap.Int("region_total_frames", ts.RegionTotalFrames),
		zap.Int("region_static_frames", ts.RegionStaticFrames),
		zap.Int64("region_elapsed_ms", nowMs-ts.RegionStaticStartedMs),
		zap.Int("prev_x", prev.X), zap.Int("prev_y", prev.Y), zap.Int("prev_z", prev.Z),
		zap.Int("cur_x", x), zap.Int("cur_y", y),
		zap.Int("region_start_z", ts.RegionStartZ),
		zap.Int64("ts_ms", nowMs),
	)
}

// nearestEnterRoomMs 在 [tMs - windowMs, tMs + windowMs] 内查最近 EnterRoom 事件。
// 返回最接近 tMs 的事件时间戳。调用方持锁。
func (tm *TrackManager) nearestEnterRoomMs(tMs int64, windowMs int64) (int64, bool) {
	var bestMs int64
	bestDelta := windowMs + 1
	found := false
	for k, e := range tm.recentRadarEvents {
		if e.EventName != alarm.EnterRoom {
			continue
		}
		delta := tMs - k
		if delta < 0 {
			delta = -delta
		}
		if delta <= windowMs && delta < bestDelta {
			bestMs = k
			bestDelta = delta
			found = true
		}
	}
	return bestMs, found
}

// hasRecentEnterRoom 检查 [tMs - enterPairWindowMs, tMs] 窗口内有无 EnterRoom 事件。
// 调用方持锁（segment 1 已持锁）。
//
// 注：仅 look-back（不 look-forward）—— 出生瞬时如 event-stream 还没到 → 判错 false ghost。
// 修正路径：BirthFinalDeadlineMs 给 grace 缓冲（FallRulesParam.Lost.BirthFinalGraceMs，默认 2s），
// 到点用 hasRecentEnterRoomBetween([T-3s, deadline]) 重检（见 tryGraceUpgrade）。
func (tm *TrackManager) hasRecentEnterRoom(tMs int64) bool {
	return tm.hasRecentEnterRoomBetween(tMs-enterPairWindowMs, tMs)
}

// hasRecentEnterRoomBetween 在显式时间窗 [fromMs, toMs] 内查 EnterRoom 事件。
func (tm *TrackManager) hasRecentEnterRoomBetween(fromMs, toMs int64) bool {
	for k, e := range tm.recentRadarEvents {
		if k < fromMs || k > toMs {
			continue
		}
		if e.EventName == alarm.EnterRoom {
			return true
		}
	}
	return false
}

// tryGraceUpgrade Birth verdict 多流时序兜底（方案 A）。
//
// 出生时 EnterRoom event-stream 可能落后于 monitor stream 1-3s，birth score 短路为 0。
// BirthFinalDeadlineMs 给 grace 缓冲；到点用扩展窗 [birth-3s, deadline] 再扫，
// 如有 EnterRoom 命中 + 出生位置物理可达（≤150cm to entry）→ 抬 birth_score。
//
// 一次性 recompute（不论结果如何，之后清空 deadline 不再重算）。
// 调用位置：segment 3 promote 之前；调用方持锁。
func (tm *TrackManager) tryGraceUpgrade(ts *TrackState, nowMs int64) {
	if ts.BirthFinalDeadlineMs == 0 {
		return // 已 finalize
	}
	if nowMs < ts.BirthFinalDeadlineMs {
		return // 还在 grace 期内
	}
	defer func() { ts.BirthFinalDeadlineMs = 0 }() // 不论结果，之后不再重算

	// 仅当 birth 是因「找不到 EnterRoom」而被短路时才补救
	if ts.BirthReason != "no_enter_pair" && ts.BirthReason != "far_from_enter" {
		tm.logger.Debug("birth_grace_skip_reason",
			zap.Int("track_id", ts.TrackID),
			zap.String("birth_reason", ts.BirthReason))
		return
	}
	// 距离判断：>150cm 物理仍不可能
	dEntry := tm.grid.NearestEntryDist(ts.BirthPos.X, ts.BirthPos.Y)
	if dEntry > birthMaxRealisticCm {
		tm.logger.Info("birth_grace_skip_far_from_enter",
			zap.Int("track_id", ts.TrackID),
			zap.Int("d_entry", dEntry),
			zap.Int("birth_x", ts.BirthPos.X), zap.Int("birth_y", ts.BirthPos.Y))
		return
	}
	// 扩展窗 [birth-3s, deadline] 重扫
	if !tm.hasRecentEnterRoomBetween(ts.BirthPos.TMs-enterPairWindowMs, ts.BirthFinalDeadlineMs) {
		tm.logger.Info("birth_grace_skip_no_enter_event_in_window",
			zap.Int("track_id", ts.TrackID),
			zap.Int64("birth_ms", ts.BirthPos.TMs),
			zap.Int64("deadline", ts.BirthFinalDeadlineMs),
			zap.Int("event_count", len(tm.recentRadarEvents)))
		return
	}
	// 命中：抬 birth score
	newScore := 50 + birthEnterPairBonus // base + entry bonus
	if dEntry < 50 {
		newScore = 50 + 20 // 与 birthScore 因子 1 dEntry<50 加分对齐
	}
	if newScore > ts.BirthScore {
		delta := newScore - ts.BirthScore
		ts.BirthScore = newScore
		ts.BirthReason = "grace_enter_pair_recovered"
		ts.Score = clampInt(ts.Score+delta, 0, 100)
		tm.logger.Info("birth_grace_upgraded",
			zap.String("device_uid", ts.DeviceAddr),
			zap.Int("track_id", ts.TrackID),
			zap.Int("new_birth_score", ts.BirthScore),
			zap.Int("d_entry_cm", dEntry),
			zap.Int64("birth_ms", ts.BirthPos.TMs),
			zap.Int64("nowMs", nowMs),
		)
	}
}

// ========================================================================
// 即时流打分（不累计到 cell，只在 track 内部）
// ========================================================================

// scoreResidual Kalman 残差打分
func (tm *TrackManager) scoreResidual(ts *TrackState, residual int) {
	switch {
	case residual < 30:
		ts.AdjustScore(2)
	case residual < 80:
		// 中等偏差
	case residual < 200:
		ts.AdjustScore(-5)
	default:
		ts.AdjustScore(-20) // 空间跳跃
	}
}

// scoreMovement 运动/静止评分 + 累计静止时长 + Dwell 记录。
// pose 参数为当前帧 raw pose（用于 still-fall 升级判定 stand 姿态）。
func (tm *TrackManager) scoreMovement(ts *TrackState, x, y int, nowMs int64, pose int) {
	if len(ts.History) < 2 {
		return
	}
	prev := ts.History[len(ts.History)-2]
	d := distInt(prev.X, prev.Y, x, y)

	cell := tm.grid.CellAt(x, y)
	isRest := cell != nil && cell.IsRestZone()

	// PR-13: region static 维护 — |dx|≤15 AND |dy|≤15 视为同区域；累积期 ≥90% 容忍单帧噪声
	tm.updateRegionStatic(ts, prev, x, y, nowMs, cell)

	// ── 即时态（d 帧间位移判据）：score / refresh / stand-static 计时（不变；不碰久静量）──
	if d < StillThreshCm {
		// PR-7.2: 跟踪 pose=Stand 静止专用计时（用于自学习升 AreaSit）
		if RadarPoseToCore(pose) == CorePoseStand {
			if ts.StandStaticSince == 0 {
				ts.StandStaticSince = nowMs
			}
		} else {
			ts.StandStaticSince = 0
		}
		// PR-11: pose=Lie on AreaBed 持续刷新 — 4h 累计 → 重锁 confidence=95
		if cell != nil && RadarPoseToCore(pose) == CorePoseLie && cell.Belief[0].Type == AreaBed {
			if ts.LyingOnBedSinceMs == 0 {
				ts.LyingOnBedSinceMs = nowMs
			}
			if !ts.AreaBedRefreshed && nowMs-ts.LyingOnBedSinceMs >= 4*3600*1000 {
				cell.MarkRestZoneByFeedback(AreaBed)
				tm.grid.boostNeighborSameType(x, y, AreaBed, nowMs)
				ts.AreaBedRefreshed = true
				tm.logger.Info("area_bed_refreshed_by_lying4h",
					zap.String("device_uid", ts.DeviceAddr),
					zap.Int("track_id", ts.TrackID),
					zap.Int("x", x), zap.Int("y", y),
					zap.Int64("ts_ms", nowMs),
				)
			}
		} else {
			ts.LyingOnBedSinceMs = 0
		}
		// PR-11: pose=Sit on AreaToilet 持续刷新 — 5min 累计 → 重锁
		if cell != nil && RadarPoseToCore(pose) == CorePoseSit && cell.Belief[0].Type == AreaToilet {
			if ts.SitOnToiletSinceMs == 0 {
				ts.SitOnToiletSinceMs = nowMs
			}
			if !ts.AreaToiletRefreshed && nowMs-ts.SitOnToiletSinceMs >= 5*60*1000 {
				cell.MarkRestZoneByFeedback(AreaToilet)
				tm.grid.boostNeighborSameType(x, y, AreaToilet, nowMs)
				ts.AreaToiletRefreshed = true
				tm.logger.Info("area_toilet_refreshed_by_sit5min",
					zap.String("device_uid", ts.DeviceAddr),
					zap.Int("track_id", ts.TrackID),
					zap.Int("x", x), zap.Int("y", y),
					zap.Int64("ts_ms", nowMs),
				)
			}
		} else {
			ts.SitOnToiletSinceMs = 0
		}
		if !isRest && ts.Verdict == VerdictPending {
			ts.AdjustScore(-3)
		}
		// 静止超时 LongStill（久静量）已迁到下方 box 判据块（单源 StillBoxRunStart，见函数末）。
		if cell != nil {
			isRiskTime := IsNightTime(nowMs, tm.timezone)

			// PR-Bootstrap: v1 TrackManager still-fall fire path 已删除，由 PR-10 BathroomStillFall
			// (§6.A.1) 在 bathroom room.kind 分支替代。stillFallTimeoutSec 谓词保留作"bathroom-like
			// 位置过滤器"（line 1971 + 2294 AreaSit 学习 gate），不再驱动 alarm 发射。

			// PR-7.2: stand-static 自学习 → AreaSit
			// 触发条件:
			//   pose=Stand 静止 ≥ 12min (cell 非 AreaSit) 或 ≥ 8min (已是 AreaSit 强化)
			//   且 cell 不在 still-fall 触发场景（toilet/shower/bathroom-room/Stay-alarm）
			// 物理意义: 人很难 stand-static 超过 12min 在非 bathroom 场景；如果 track 在此
			// 静止这么久且没消失/没报警 = 该位置实际是个站立工作/坐位（如水池前/化妆台/电脑桌）
			// 阈值 < 15min 避免与 still-fall 冲突。
			if ts.StandStaticSince > 0 && !ts.AreaSitAutoLearned && cell != nil {
				if tm.stillFallTimeoutSec(cell, isRiskTime) == 0 {
					// 非 still-fall 触发场景；可学习
					// PR-9.2: 任何已是 RestZone（AreaSit / AreaBed）都用 8min（强化）；其它 12min
					threshold := int64(12 * 60 * 1000)
					if cell.IsRestZone() {
						threshold = int64(8 * 60 * 1000)
					}
					if nowMs-ts.StandStaticSince >= threshold {
						locked := cell.MarkRestZoneByFeedback(AreaSit)
						cell.IncrToleratedStill()
						// PR-10: 自学习也触发邻居增强（与人工反馈一致）
						tm.grid.boostNeighborSameType(x, y, AreaSit, nowMs)
						ts.AreaSitAutoLearned = true
						tm.logger.Info("area_sit_auto_learned",
							zap.String("device_uid", ts.DeviceAddr),
							zap.Int("track_id", ts.TrackID),
							zap.Int("stand_static_sec", int((nowMs-ts.StandStaticSince)/1000)),
							zap.Int("threshold_sec", int(threshold/1000)),
							zap.Bool("area_sit_locked", locked),
							zap.Int("cell_area_before", int(cell.Belief[0].Type)),
							zap.Int("x", x), zap.Int("y", y),
							zap.Int64("ts_ms", nowMs),
						)
					}
				}
			}
		}

		// PR-9: v1 R4 bedside_fall 触发块整段删除（依赖 lastLeftBedAt）。
		// PR-10/11 BathroomBedsideFall + bedroom bedside_fall 将以 SuiteCensus + BedSession 重写，
		// 触发主体改为 SuitePerson（决定 8），用 90s grace + 任意位置静止 ≥ 8 min（决定 18 真多人不 fire）。
	} else {
		// 即时态移动：reset 即时态计时（StandStatic/Lying/Sit/StillFall）；速度合理性。
		// 久静量（Dwell/ToleratedStill/LongStill/Anomaly/LongStillReported）走下方 box 判据块。
		ts.StillFallReported = false
		ts.StandStaticSince = 0 // PR-7.2: track 移动 → reset stand-static 计时
		ts.LyingOnBedSinceMs = 0
		ts.SitOnToiletSinceMs = 0 // PR-11: track 移动 → reset 持续观测计时
		// 速度合理性
		speed := int(math.Round(ts.Kalman.Speed()))
		if speed > 10 && speed < 150 {
			ts.AdjustScore(3)
		} else if speed >= 150 {
			ts.AdjustScore(-2)
		}
	}

	// ── 久静量（box 判据，单源 StillBoxRunStart/StillBoxStartXY；updateContinuousIndicators 已同步算）──
	// 与 DBN 双读同一 box 字段，谁都不重算（§2.4 producer/maintainer；旧 StillSince 即时位移态已删）。
	if cell != nil && ts.StillBoxRunStart > 0 {
		// box 静止超时 → LongStill（位置用当前 x,y，与静止 track 位置一致）
		isRiskTime := IsNightTime(nowMs, tm.timezone)
		if timeout := cell.EffectiveStillTimeoutSec(isRiskTime); timeout > 0 {
			if stillSec := int((nowMs - ts.StillBoxRunStart) / 1000); stillSec > timeout {
				ts.CurrentAnomaly = AnomalyStillTooLong
				if !ts.LongStillReported {
					tm.grid.MarkLongStill(x, y, nowMs)
					ts.LongStillReported = true
				}
			}
		}
	} else if ts.StillBoxBreakDurMs > 0 {
		// box 刚 break = 静止结束 → MarkDwell + ToleratedStill（box 时长 + still 区覆盖的 10×10 格，
		// 与 DBN 查的 live 格对齐；bounds 不可得则回退单格 box 起点）。
		minX, minY, maxX, maxY, ok := ts.StillBoxBoundsWithinMs(30_000, nowMs)
		if dwellSec := int(ts.StillBoxBreakDurMs / 1000); dwellSec > 0 {
			if ok {
				tm.grid.MarkDwellRegion(minX, minY, maxX, maxY, dwellSec, nowMs)
			} else {
				tm.grid.MarkDwell(ts.StillBoxStartX, ts.StillBoxStartY, dwellSec, nowMs)
			}
		}
		if ts.LongStillReported {
			if ok {
				tm.grid.MarkToleratedStillRegion(minX, minY, maxX, maxY, nowMs)
			} else {
				tm.grid.MarkToleratedStill(ts.StillBoxStartX, ts.StillBoxStartY, nowMs)
			}
		}
		ts.LongStillReported = false
		if ts.CurrentAnomaly == AnomalyStillTooLong {
			ts.CurrentAnomaly = AnomalyNone
		}
	}

	// 因子 5: 平均速度
	age := ts.AgeSec()
	if age > 5 {
		avgSpeed := ts.TotalDisplacement() / age
		if avgSpeed < 2 && !isRest {
			ts.AdjustScore(-2)
		}
	}
}

// detectZNoise Z 突变检测（本 track 内部统计，不累计到 cell）
func (tm *TrackManager) detectZNoise(ts *TrackState, z int) {
	if ts.LastZ == 0 {
		return
	}
	dz := ts.LastZ - z
	if dz < 0 {
		dz = -dz
	}
	if dz > 50 { // Z 单帧突变 > 50cm
		ts.ZNoiseCount++
	}
}

// detectPoseMismatch pose 与运动学矛盾（track 内部累计）
func (tm *TrackManager) detectPoseMismatch(ts *TrackState, pose int) {
	speed := int(math.Round(ts.Kalman.Speed()))
	// pose=Walking 但速度 ≈ 0
	if pose == observation.PoseWalking && speed < 5 && ts.FrameCount > 3 {
		ts.PoseMismatchCount++
		ts.CurrentAnomaly = AnomalyPoseMismatch
		ts.AdjustScore(-3)
	}
	// pose=Standing 但 Z < 50（在地面）
	if pose == observation.PoseStanding && ts.LastZ > 0 && ts.LastZ < 50 {
		ts.PoseMismatchCount++
	}
}

// ========================================================================
// 核心姿态状态机（替代原 Pose 两阶段状态机）
// 只做 Lie 进出检测：Stand/Move → Lie → Stand/Move < 3 秒 = LieRetract
//                    Stand/Move → Lie + Z 骤降 = FallEvent
// ========================================================================

func (tm *TrackManager) updateLieStateMachine(ts *TrackState, pose, x, y int, nowMs int64) {
	curCore := RadarPoseToCore(pose)
	prevCore := ts.PrevCore

	// 进入 Lie 态
	if curCore == CorePoseLie && prevCore != CorePoseLie {
		ts.LieEnteredAt = nowMs
		ts.LieEnteredX = x
		ts.LieEnteredY = y
	}

	// 退出 Lie 态
	if prevCore == CorePoseLie && curCore != CorePoseLie {
		if ts.LieEnteredAt > 0 {
			lieDuration := nowMs - ts.LieEnteredAt
			// 短暂 Lie 后回 Stand/Move → Retract
			if lieDuration < LieRetractMs &&
				(curCore == CorePoseStand || curCore == CorePoseMove) {
				tm.grid.MarkLieRetract(ts.LieEnteredX, ts.LieEnteredY, nowMs)
			}
			ts.LieEnteredAt = 0
		}
	}

	ts.PrevCore = curCore
}

// ========================================================================
// 综合风险分
// ========================================================================

func (tm *TrackManager) computeRisk(ts *TrackState, nowMs int64) int {
	if ts.Verdict == VerdictGhost {
		return 0
	}
	base := 0
	switch ts.CurrentAnomaly {
	case AnomalyFall:
		base = 100
	case AnomalyBedFall:
		base = 100 // 双源矛盾确认的 bed-fall，最高风险
	case AnomalyBedsideFall:
		base = 100 // R4 床边晕倒（夜间 + LeftBed 后床边静止超时），最高风险
	case AnomalyStillTooLong:
		base = 60
	case AnomalyPathBreak:
		base = 80
	case AnomalyPoseMismatch:
		base = 70
	}
	if base == 0 {
		return 0
	}
	tf := timeFactor(nowMs)
	of := tm.occupancyFactor()
	risk := float64(base) * tf * of
	return clampInt(int(math.Round(risk)), 0, 100)
}

func timeFactor(nowMs int64) float64 {
	hour := time.UnixMilli(nowMs).Hour()
	switch {
	case hour >= 5 && hour < 6:
		return 2.0
	case hour >= 22 || hour < 5:
		return 1.5
	case hour >= 6 && hour < 8:
		return 1.3
	default:
		return 1.0
	}
}

func (tm *TrackManager) occupancyFactor() float64 {
	realCount := 0
	for _, ts := range tm.tracks {
		if ts.Verdict == VerdictReal {
			realCount++
		}
	}
	bedOccupied := tm.sleepadInBedCount > 0
	if realCount <= 1 {
		return 1.5
	}
	if realCount == 2 && bedOccupied {
		return 1.2
	}
	return 1.0
}

// stillFallTimeoutSec "bathroom-like 位置过滤器"（不再驱动 alarm 发射，PR-Bootstrap 拆走 fire path）。
// 仅作 AreaSit 自学习 gate 用（line 1971 + 2294 处）：bathroom 内长时间 stand 不应被误学为坐位。
//
//	cell.AreaToilet/AreaShower → cell.EffectiveStillTimeoutSec
//	cell 未学到但 room.name 是 bathroom → FallRulesParam.Still.ToiletShowerSec × NonRiskTimeFactor
//	都不匹配 → 0
//
// PR-Bootstrap：删除 stayAlarmEnabled 分支（loadStayAlarmEnablement 已删，stayAlarmEnabled 永 false）。
// 调用方持锁。
func (tm *TrackManager) stillFallTimeoutSec(cell *Cell, isRiskTime bool) int {
	if cell == nil {
		return 0
	}
	cellType := cell.Belief[0].Type
	if cellType == AreaToilet || cellType == AreaShower {
		return cell.EffectiveStillTimeoutSec(isRiskTime)
	}
	if roomutil.IsBathroom(tm.roomName) {
		base := FallRulesParam.Still.ToiletShowerSec
		if !isRiskTime {
			base = int(float64(base) * FallRulesParam.Still.NonRiskTimeFactor)
		}
		return base
	}
	return 0
}

// realTrackCount 当前 VerdictReal track 数（用于 lost-fall NumberPeople≥2 取消判定）。
// 调用方持锁（segment 内部）。
func (tm *TrackManager) realTrackCount() int {
	n := 0
	for _, t := range tm.tracks {
		if t.Verdict == VerdictReal {
			n++
		}
	}
	return n
}

// NonGhostTrackCount 当前非 ghost track 数（排除 VerdictGhost，保留 Pending/Real/Anchored）。
// zoneengine room total_people 用作 radar_np（替 firmware number_people，后者含 ghost）。
// 保留 Pending：ghost 只在 ≥2 track 生效，单条新 track 不会被判 ghost，计入避免出生期少计。线程安全。
func (tm *TrackManager) NonGhostTrackCount() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	n := 0
	for _, t := range tm.tracks {
		if t.Verdict != VerdictGhost {
			n++
		}
	}
	return n
}

// otherDeviceRealTTLMs 同房另一雷达"近期见过真 track"的有效窗（>radar 1Hz + 抖动）。
const otherDeviceRealTTLMs = 8_000

// otherDeviceRealTrackRecent 同房是否有 excludeDevice 之外的雷达在 TTL 内见过真 track。
// 用于 lost-fall fire 前的同房占用对账：本台丢/幻影但别台仍见真人 → 人还在房 → 抑制。
// 调用方持锁（segment 内部）。
func (tm *TrackManager) otherDeviceRealTrackRecent(excludeDevice string, nowMs int64) bool {
	for dev, ts := range tm.lastRealTrackByDevice {
		if dev == excludeDevice {
			continue
		}
		if nowMs-ts <= otherDeviceRealTTLMs {
			return true
		}
	}
	return false
}

// updateContinuousIndicators 每帧维护 StillBox（静止无移动）检测 + Kalman birth-coherence 指标。
//
//  1. StillBox（静止无移动）检测（box 判据，2026-05-03 由 byte-equal 改为 box）：
//     最近 30s History 滚动窗口内位移 box per-axis (max-min) ≤ StillBoxCm(50) → still。
//     - 进入 still：起点 = History 最早帧 TMs（自然回填到 box 内最早一帧）
//     - 持续 still：StillBoxRunStart 不变（即使 History 滚动丢早期帧）
//     - 跳出 box：StillBoxRunStart 清零
//     用于 lost-fall pending credit（半计入 wait）+ PR-C 流式 cancel 守卫。
//  2. MaxKalmanResidual：track 生命周期峰值残差。
//  3. MaxImpliedSpeedFromBirth：max(dist(current, birth) / age) cm/s；
//     > ImpossibleSpeedCm 判硬 ghost；> SuspectSpeedCm + 无 EnterRoom 判软 ghost。
//
// 调用位置：processFrameAt 已有 track 分支，Kalman.Update 之后。
func (tm *TrackManager) updateContinuousIndicators(ts *TrackState, f TrackFrame, nowMs int64, residualF float64) {
	// ---- StillBox（静止无移动）检测（50×50 per-axis box 判据）----
	// 用 BoxRangeWithinMs（max(dx,dy)）而非 DisplacementWithinMs（对角线）：50×40 倒地框
	// 对角线 64 会被误判成"动"，per-axis 算 50（≤StillBoxCm=50）才正确判 still。
	ts.StillBoxBreakDurMs = 0 // 每帧清；仅本帧 break 时设（移动块 MarkDwell 消费）
	disp := ts.BoxRangeWithinMs(30_000, nowMs)
	if disp <= FallRulesParam.Lost.StillBoxCm && len(ts.History) >= 2 {
		if ts.StillBoxRunStart == 0 {
			// 起点回填到 History 最早帧（box 内最早可见点）；同步存起点位置（cell engine 久静量单源读）。
			ts.StillBoxRunStart = ts.History[0].TMs
			ts.StillBoxStartX = ts.History[0].X
			ts.StillBoxStartY = ts.History[0].Y
		}
	} else if ts.StillBoxRunStart > 0 {
		dur := nowMs - ts.StillBoxRunStart
		ts.StillBoxBreakDurMs = dur // 暂存刚结束的 dwell 时长，供 scoreMovement 移动块 MarkDwell
		ts.StillBoxRunStart = 0
		if tm.logger != nil {
			tm.logger.Info("still_box_break",
				zap.Int("track_id", f.TrackID),
				zap.Int64("duration_ms", dur),
				zap.Int("disp_cm", disp),
				zap.Int64("now_ms", nowMs))
		}
	}
	// NOTE（防御层备忘，未实施）：box 判据只看 max-min 范围。理论 edge case：
	// box 内反复抖动（30cm 范围内来回跨越）→ box 小但累计位移大 → 误判 still。
	// 实测 D523/CD2B/D5F7 未观察到（firmware 单帧抖动通常 ±5-10cm）。生产若遇
	// 此类"假 still 导致 lost-fall 误判"再加：30s 内逐帧累计 > 200cm 也清零
	// StillBoxRunStart（200cm = box 周长 120cm × 1.6 倍，超过即反复跨越）。

	// ---- MaxKalmanResidual ----
	if residualF > ts.MaxKalmanResidual {
		ts.MaxKalmanResidual = residualF
	}

	// ---- MaxImpliedSpeedFromBirth ----
	ageMs := nowMs - ts.BirthPos.TMs
	if ageMs >= 1000 {
		distFromBirth := distInt(f.X, f.Y, ts.BirthPos.X, ts.BirthPos.Y)
		// implied = dist / ageSec = dist * 1000 / ageMs (cm/s)
		implied := int(int64(distFromBirth) * 1000 / ageMs)
		if implied > ts.MaxImpliedSpeedFromBirth {
			ts.MaxImpliedSpeedFromBirth = implied
		}
	}
}
