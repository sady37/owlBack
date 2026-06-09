package roomengine

import (
	"context"
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
	Source     string // "radar_direct" / "engine_silent_leftbed" / "engine_lost" / "engine_still_fall" / ...
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

	// pendingLostFalls：lost-fall 规则的挂起池。
	// 时长按消失点 cell areaType（5min walkway / 60min bed / ...），含 StillBox 静止 credit；
	// 取消条件：新 track 出生 / ExitRoom 事件 / room.NumberPeople ≥ 2。
	pendingLostFalls map[int]*PendingLostFall

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

	// Lost fall 统计
	lostFallPendingCreated   int
	lostFallPendingCancelled int // 含 birth-recovery / ExitRoom 两类取消（PR-9 删除 NumberPeople 路径）
	lostFallReported         int

	// Silent fall（LeftBed 矛盾路径）统计
	silentFallLeftbedReported  int
	silentFallLeftbedCancelled int // wait 满但 radar 已离开 Bed 邻域 → 取消（人正常起床）

	// PR-Bootstrap: v1 stillFallReportCount 已删除（v1 fire path 同时删除）。
	// PR-10 BathroomFallRules 的统计走 ai.log audit 路径。

	// Firmware Fall verifier 累计（PR-5；仅打分，不否决 alarm）
	fallVerifyGhostCount   int
	fallVerifySuspectCount int
	fallVerifyRealCount    int

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

	// weakBioSource: fall verifier "WeakBio≥80 force real" 短路用（A 风险放大消费者）。
	// 默认 nil（verifier 走原三档评分）；engine.SetWeakBioSource 注入。
	weakBioSource WeakBioSource

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

	// PR-8: AI 派生事件 / 告警发布回调。engine 注入；nil 时不发布（playback / 测试场景）。
	aiPublisher AIPublisher
}

// AIPayload 发布 AI 事件/告警的轻量载荷。
// 不耦合 TrackState（lost fall 在 track 删除后才报，没有 ts 在手）。
//
// 设计原则（PR5b）：
//  1. 观测信号走 observation.Track —— 与 firmware/engine 上游同一 schema，
//     AI 只填它要表达的字段（如 ghost verdict 只填 TrackConfidence + 位置；
//     未来 vital 修正只填 HeartRate + VitalConfidence）。零值 = "AI 无意见"。
//  2. Source 是决策路径标识（一等公民）—— 不是 AI/非AI 二元 tag，而是具体的
//     "哪个算法路径产生的"，用于审计、调参、误报追溯、商业演示讲故事。
//     cardagg 落 alarm_events.metadata.source；前端可读、可 group by。
//  3. Reason / Evidence 是审计元数据 —— 解释"为什么 AI 这么判"，不参与下游
//     分支判定（分支判定由 Source 或 Track 数值阈值驱动）。
type AIPayload struct {
	DeviceAddr string // 源 sensor UUID（FK to devices.device_addr）
	RoomID     string

	// AI 写的观测（与上游 firmware/engine 同一 schema）。
	// AI 只填它要表达的字段，其它留零值 = "AI 对此无意见"。
	Track observation.Track

	// Source 数据生产者节点身份（WHO）。当前 TrackManager 全部派生自护工角色，
	// engine.publishAIMessage 默认填 cfg.AIPublish.Source（如 "AI.Caregiver01"）。
	// 本字段保留作未来多角色 override 钩子（例：同一 TrackManager 内派生健康风险
	// verdict 时可显式设为 "AI.Doctor01"）。空 = 用 engine 默认。
	Source string

	// Reason AI 派生决策的理由路径（WHY）。填本地 Reason* 路径常量；空 = 非 AI 派生。
	Reason string

	// Evidence 证据 KV map（score / penalty / context 等审计字段，下游不解析）。
	Evidence map[string]interface{}

	// Event 决策事件类型（审计用，写 sensor_decision_log.event）：verdict_change /
	// lostfall_pending / lostfall_cancel / lostfall_fire / lostfall_suppress。空 = 不是决策审计事件。
	Event string

	// EventStatus 事件生命周期阶段（"start" / "end" / "instant"）。空 = "instant"（默认）。
	// 用于 firmware 撤销链路：qinglan 收到 Initialization (last_pose=2/7) → forward end →
	// sensor 透传 → cardagg AlarmRouter 按 Registry[Fall/SittingOnGround].EndPolicy=AutoResolve 关 alarm。
	EventStatus string

	// IncidentMs 实际发生时刻 ms（推断类 fall 才有值）。
	// firmware Fall = 0（incident == alerted == nowMs）；silent/lost/still fall = last_active/still_box_start/leftbed/empty_since 等真实 incident 时刻。
	// 写到 fields["incident_ts_ms"]，cardagg AlarmRouter 当 alarm_events.triggered_at；nowMs 当 alerted_at。
	// 0 = 不写字段，下游回退 triggered_at = alerted_at = nowMs。
	IncidentMs int64
}

// CategoryTrackVerdict 是 track verdict 的 category 路由键（事件 TYPE，不是 verdict label）。
const CategoryTrackVerdict = "track_verdict"

// CategorySensorDecision 是 lost-fall 决策审计事件的 category（与 verdict 同流 ai:track:verdict:stream，
// iot 落 sensor_decision_log；cardagg override 只认 track_verdict，按 category 跳过本类）。
const CategorySensorDecision = "sensor_decision"

// Reason 常量本地定义——目前 wisefido-sensor 是唯一 producer，下游 cardagg /
// wisefido-data 透传字符串不解析。未来若出现第二个 AI producer（如健康风险
// 模块），再上移到 owl-common 共享。
//
// Source（"AI.Caregiver01" 等）由 engine 在 publishAIMessage 默认填入
// （取自 cfg.AIPublish.Source），TrackManager 不直接设置——避免节点身份
// 散落到业务代码。
const (
	// Reason: ghost 判定（track_verdict category）
	ReasonGhostPostReal = "ghost_post_real" // Real → Ghost 翻转（已确认 real 后 penalty 累积超阈）
	ReasonGhostPenalty  = "ghost_penalty"   // 主路径：累积 penalty ≥ 阈值（含 motion_symmetry / no_enter_pair）
	ReasonGhostLowScore = "ghost_low_score" // probation 期满 score 低于阈值

	// Reason: fall 判定（alarm category，4 个 fall 派生子类型）
	ReasonLostTrack            = "lost_track"             // track 异常消失（可能是真摔倒后失锁）
	ReasonStillInBathroom      = "still_in_bathroom"      // 浴室长时间静止 still-fall
	ReasonBedsideSilent        = "bedside_silent"         // LeftBed 后床边静止过久 R4
	ReasonSleepadRadarConflict = "sleepad_radar_conflict" // sleepad LeftBed + radar 仍在床

	// sensor_v2 PR-10 BathroomFallRules（§6.A）— SuitePerson 主语 fall：
	ReasonBathroomStill              = "bathroom_still"                        // §6.A.1 cell Toilet/Shower + Stand + 10/12min
	ReasonBathroomLongStatic         = "bathroom_long_static"                  // §6.A.2 90s grace + 任意位置 8min（非 Toilet/Shower）
	ReasonSuitePersonCompletelyLost  = "suite_person_completely_lost_no_ghost" // §6.A.3 最强档 bathroom 内活 track==0 ≥ 30s
	ReasonSuitePersonSilentWithGhost = "suite_person_silent_with_ghost_proxy"  // §6.A.3 次强档 SuitePerson static ≥ 7min

	// sensor_v2 PR-11 BedroomFallRules（§6.B）— SuitePerson 主语 bedroom fall：
	ReasonBedroomBedsideStatic = "bedroom_bedside_static" // §6.B.3 BedState→Vacant + 床边 ≤100cm 静止 ≥15min
	ReasonBedroomPersonSilent  = "bedroom_person_silent"  // §6.B.2 SuitePerson AnchorRoomType=bedroom + LastActiveMs > threshold
)

// AIPublisher PR-8 解耦：TrackManager 不直接持有 redis client，由 engine 实现接口注入。
type AIPublisher interface {
	PublishAIEvent(ctx context.Context, p AIPayload, category string, nowMs int64)
	PublishAIAlarm(ctx context.Context, p AIPayload, category string, nowMs int64)
	// DeviceUIDHex 反查 device addr → hex MAC（log 双格式人眼可读）；缺失返回空字符串
	DeviceUIDHex(deviceAddr string) string
}

// devUIDHex helper：track_manager 内部调用 aiPublisher 反查；nil-safe（nil publisher 返回空）
func (tm *TrackManager) devUIDHex(deviceAddr string) string {
	if tm.aiPublisher == nil {
		return ""
	}
	return tm.aiPublisher.DeviceUIDHex(deviceAddr)
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
		pendingLostFalls:        make(map[int]*PendingLostFall),
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

// SetWeakBioSource 注入 fall verifier 的 WeakBio 提级 source（engine.SetWeakBioSource 转发）。
// nil 允许（verifier 短路 disable，走原三档评分）。
func (tm *TrackManager) SetWeakBioSource(s WeakBioSource) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.weakBioSource = s
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

// SetStartupMs 覆盖默认的 startup 时间戳。playback / 离线测试用：把 startupMs
// 对齐到回放窗口起点（默认是 time.Now()，离线时无意义）。
func (tm *TrackManager) SetStartupMs(ms int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if ms > 0 {
		tm.startupMs = ms
	}
}

// SetAIPublisher PR-8：注入 AI 派生事件/告警的发布器。
// engine 在 RegisterRoom 后调用，传入 engine 自身（实现 AIPublisher 接口）。
// nil = 不发布（playback / 测试场景）。
func (tm *TrackManager) SetAIPublisher(p AIPublisher) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.aiPublisher = p
}

// emitAIEvent / emitAIAlarm 内部 helper：仅当 aiPublisher 非 nil 时调用。
// payloadFromTrack 从 TrackState 构造 AIPayload。

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

func (tm *TrackManager) payloadFromTrack(ts *TrackState) AIPayload {
	conf := 100 - ts.GhostPenalty
	if conf < 0 {
		conf = 0
	}
	if conf > 100 {
		conf = 100
	}
	return AIPayload{
		DeviceAddr: ts.DeviceAddr,
		RoomID:     ts.RoomID,
		Track: observation.Track{
			BedStatus:       observation.BedStatusUnchanged,
			TrackID:         ts.TrackID,
			LogicID:         ts.LogicID,
			PositionX:       intPtr(ts.LastRawH),
			PositionY:       intPtr(ts.LastRawV),
			PositionZ:       intPtr(ts.LastRawZ),
			Pose:            ts.LastPose,
			TrackConfidence: conf,
		},
	}
}

func (tm *TrackManager) emitAIEvent(p AIPayload, category string, nowMs int64) {
	if tm.aiPublisher == nil {
		return
	}
	tm.aiPublisher.PublishAIEvent(context.Background(), p, category, nowMs)
}

func (tm *TrackManager) emitAIAlarm(p AIPayload, category string, nowMs int64) {
	if tm.aiPublisher == nil {
		return
	}
	tm.aiPublisher.PublishAIAlarm(context.Background(), p, category, nowMs)
}

// emitGhostVerdict 发布 track_verdict（事后裁决）到 iot:event:stream。
//
// reason 传 ReasonGhostXxx 路径常量；context 传自然语言细节（如 BirthReason
// / "low_score"），进 Evidence["context"] 作审计。Reason 是机器可分类的子路
// 径，下游可按值 group by；context 是自然语言审计文本，下游不解析。
// Source 由 engine.publishAIMessage 默认填入（cfg.AIPublish.Source），不在此设置。
// TrackConfidence 由 payloadFromTrack 给（= 100 - GhostPenalty，AI 实时评估值，
// ghost 路径自然落入低分区间），下游按数值阈值渲染饱满度（如 ≤30 → 低饱和）。
//
// 不参与 alarm 触发路径——firmware/AI alarm 仍按"宁可误报不可漏报"原则照常 fire。
func (tm *TrackManager) emitGhostVerdict(ts *TrackState, reason, context string, nowMs int64) {
	p := tm.payloadFromTrack(ts)
	p.Reason = reason
	p.Evidence = map[string]interface{}{
		"score":         ts.Score,
		"birth_score":   ts.BirthScore,
		"ghost_penalty": ts.GhostPenalty,
		"frame_count":   ts.FrameCount,
	}
	if context != "" {
		p.Evidence["context"] = context
	}
	p.Event = "verdict_change"
	p.Evidence["verdict"] = int(ts.Verdict)
	tm.emitAIEvent(p, CategoryTrackVerdict, nowMs)
}

// emitDecision 发布一条 lost-fall 决策审计事件到 ai:track:verdict:stream（category=sensor_decision）。
// iot 落 sensor_decision_log；cardagg 按 category 跳过（不污染 override cache）。旁路审计，不在热路径。
// event=lostfall_pending/lostfall_cancel/lostfall_fire/lostfall_suppress；reason 机器可分类；ev 是全量特征。
func (tm *TrackManager) emitDecision(deviceAddr, logicID string, trackID, rawH, rawV, rawZ int, event, reason string, ev map[string]interface{}, nowMs int64) {
	if tm.aiPublisher == nil {
		return
	}
	tm.emitAIEvent(AIPayload{
		DeviceAddr: deviceAddr,
		RoomID:     tm.roomID,
		Track: observation.Track{
			BedStatus: observation.BedStatusUnchanged,
			TrackID:   trackID,
			LogicID:   logicID,
			PositionX: intPtr(rawH),
			PositionY: intPtr(rawV),
			PositionZ: intPtr(rawZ),
		},
		Event:    event,
		Reason:   reason,
		Evidence: ev,
	}, CategorySensorDecision, nowMs)
}

// IsBathroomByRoomName 用 owl-common/roomutil.ClassifyRoomType 判定本房间是否 bathroom。
// 与 cell.Belief[0].Type ∈ {AreaToilet, AreaShower} 取并集驱动 still fall。
func (tm *TrackManager) IsBathroomByRoomName() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return roomutil.IsBathroom(tm.roomName)
}

// ProcessSleepadBedEvent 接收 sleepad InBed/LeftBed 事件。
// 维护 BedSession（silent fall 状态机入口）+ PR-14 双源入场门控。
//
// PR-9: v1 bedPersonCount + lastLeftBedAt 维护已删除。
//   - 多人床场景由 SuiteCensus (PR-2) 在 person 维度处理，不再 per-bed-device counter
//   - LeftBedMaxPeople latch 保留为字段但 PR-9 阶段无源（值固定 0）；
//     PR-11 silent_fall 重写时改用 SuiteCensus.BathroomCount 或其它人维度信号
//
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
		// 任意 InBed → 清掉之前的等待状态（视为新一轮上床）
		s.LeftBedAtMs = 0
		s.SilentFallAlerted = false
		// PR-14：入场门控——若 radar InBed 在 ±15s 内已到，标记双源一致确认。
		// PR-9: lastLeftBedAt 新鲜度守卫已删；delta ≤ 15s 隐式保证 radar InBed 是最近的。
		if tm.lastRadarInBedMs > 0 {
			delta := evt.TMs - tm.lastRadarInBedMs
			if delta < 0 {
				delta = -delta
			}
			if delta <= bedInBedConsistencyMs {
				s.RadarInBedConfirmedMs = evt.TMs
			}
		}
		return
	}

	// LeftBed：BedSession 进入「等待矛盾」状态，要求满足 5min precondition。
	// PR-9: 删除多人床 c>0 guard（依赖 bedPersonCount）；任意 LeftBed event 都视作 session 终结。
	s := tm.bedSessions[evt.DeviceUID]
	if s == nil || s.InBedSinceMs == 0 {
		return // 没有有效 in-bed 历史
	}
	if evt.TMs-s.InBedSinceMs < int64(FallRulesParam.Silent.MinInBedSec)*1000 {
		// 在床时间不足 5min，不进入等待；直接结束 session
		s.InBedSinceMs = 0
		s.HasHRRR = false
		return
	}
	s.LeftBedAtMs = evt.TMs
	s.LeftBedHadHRRR = s.HasHRRR
	// LeftBedMaxPeople 在 PR-9 阶段无 source（bedPersonCount 已删），保留字段值 0；
	// PR-11 silent_fall 重写时改用 SuiteCensus 派生
	s.LeftBedMaxPeople = 0
	// Layer-1：LeftBed 时 latch 上床置信（radar 印证度），供 vanish-fire 门控
	s.InBedConfidence = bedInBedConfidence(s)
}

// ProcessSleepadObservation 接收 sleepad 一帧观测，按设备 UID 保留最新状态。
// 由 Engine.handleMessage 路由 device_type=Sleepad 时调用；
// silent fall 报警前会查询此状态做 short-circuit（"sleepad 确认在床有 vital 即不报"）。
//
// BedSession 钩子：在 in-bed 期间任意时刻观测到 HR/RR > 0 → HasHRRR=true（用于 LeftBed 时刻 latch）。
func (tm *TrackManager) ProcessSleepadObservation(obs SleepadObservation) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cur, ok := tm.sleepadStates[obs.DeviceUID]
	if !ok || obs.TMs > cur.TMs {
		// PR-9: 状态转换 InBed=true → InBed=false 不再更新 lastLeftBedAt（字段已删）。
		// BedSession.LeftBedAtMs latch 仍由 ProcessSleepadBedEvent event 路径维护；
		// 仅 monitor 路径未带 LeftBed event 的 firmware 会 miss session 终结，
		// PR-11 silent_fall 重写时如有需要再走 monitor→event 归一化。
		copyObs := obs
		tm.sleepadStates[obs.DeviceUID] = &copyObs
	}
	// BedSession：在 in-bed 期间见到 HR/RR > 0 → 打 vital flag
	if obs.InBed && obs.HasVitalSign() {
		if s := tm.bedSessions[obs.DeviceUID]; s != nil && s.InBedSinceMs > 0 {
			s.HasHRRR = true
		}
		// PR-11: 双方一致 InBed ±15s → sleepad HR/RR 可信，refresh active radar track 当前 cell 为 AreaBed
		// 设计：sleepad 不知坐标，但 radar 也报 InBed 且时间一致 → radar 当前 track 位置就是床位
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
	bedConfSleepad   = 90 // 接触式上床基础置信
	bedConfRadar     = 70 // radar InBed 印证增量 / radar 全程无 track 时的降权幅度
	bedConfEMIFloor  = 30 // radar 全程无 track 时的置信地板（防震动/EMI 假上床；真人静卧雷达也可能无回波，故不归零）
	bedVanishMinConf = 50 // vanish-fire 所需最低床态置信（低于此=可能假上床，不发跌倒）
)

// track 驱逐时间阈（2026-06-04，治 88 稀疏心跳下陈旧 track 拖 142s，Case2）：
//
//	trackEvictMaxMs：纯帧计数 MaxMissCount 之外的时间兜底——上次真观测超此值即驱逐（不受帧稀疏影响）。
//	heartbeat88EvictMs：固件明示无目标(88 心跳)持续 ≥ 此值的加速驱逐窗（镜像 cardagg ClearDeviceTracks 6s 平滑窗）。
const (
	trackEvictMaxMs    = int64(12_000)
	heartbeat88EvictMs = int64(6_000)
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// bedInBedConfidence Layer-1 三档融合（确认同一张床后，于 LeftBed 时算）：
//
//	radar 印证在床（±15s 同事件 或 单床房 session 内 InBed） → min(sleepad+radar,100) 双确认冲顶
//	radar 看到人但不在床（session 内有 track，无 InBed）       → sleepad（接触式权威）
//	radar 全程无 track                                       → max(sleepad-radar,30) 抗震动/EMI 降权不归零
func bedInBedConfidence(s *BedSession) int {
	switch {
	case s.RadarInBedConfirmedMs > 0 || s.RadarInBedInSessionMs > 0:
		return minInt(bedConfSleepad+bedConfRadar, 100)
	case s.RadarSawTrackMs > 0:
		return bedConfSleepad
	default:
		return maxInt(bedConfSleepad-bedConfRadar, bedConfEMIFloor)
	}
}

// ghostJudgable ghost≥2 闸：ghost 是真 track 的镜像/反射，需 ≥2 条 track（至少 1 条作母体）才可判 ghost。
// 仅 1 条 track 时不判 ghost——单 track 判 ghost 会压制从盲区返回的真人/真摔。调用方持锁。
func (tm *TrackManager) ghostJudgable() bool {
	return len(tm.tracks) >= 2
}

// anyActiveTrack 房内是否有任一 active（非 ghost）track。调用方持锁。
func (tm *TrackManager) anyActiveTrack() bool {
	for _, ts := range tm.tracks {
		if ts.Verdict != VerdictGhost {
			return true
		}
	}
	return false
}

// latchRadarTrackForBedSessions 每 Tick：对仍在床（未 LeftBed）的 session，
// 若 radar 当前有 active（非 ghost）track，latch RadarSawTrackMs + 该雷达 addr。
// 供 Layer-1 区分"看到人" vs "全程无 track"，并给 vanish-fire 提供报警归因。调用方持锁。
func (tm *TrackManager) latchRadarTrackForBedSessions(nowMs int64) {
	if len(tm.bedSessions) == 0 {
		return
	}
	addr := ""
	for _, ts := range tm.tracks {
		if ts.Verdict != VerdictGhost {
			addr = ts.DeviceAddr
			break
		}
	}
	if addr == "" {
		return
	}
	for _, s := range tm.bedSessions {
		if s.InBedSinceMs > 0 && s.LeftBedAtMs == 0 {
			s.RadarSawTrackMs = nowMs
			s.RadarDeviceAddr = addr
		}
	}
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
	if latestInBed == 0 {
		return card.BedState{} // 无床数据 → BedConfidence=0 → bedAdapter Fresh=false,不喂 shadow
	}
	if latestLeftBed >= latestInBed {
		// 任一源 LeftBed ≥ 最近 InBed(any-source-OR veto)→ 离床 → 不压 → Fall 浮出(漏报-safe,审查㊾)
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

// RecordRadarAlarm 落账 radar 来源的 alarm（当前阶段仅 Fall + SittingOnGround）+ 跑 verifier 评分
// + 默认转发回 iot:alarm:stream（producer="wisefido-sensor"，让 cardagg 落库）。
// 调用方（engine.handleEventMessage 的 radar Fall 分支 / handleAlarmMessage 兼容路径）
// 应当紧跟 tm.Tick(alarm.TMs) 触发段 4-6 立即跑一次。
//
// 2026-05-15 cardagg_sensor_split：firmware radar Fall 走 event stream → sensor 接管。
// 当前阶段 verifier 仅 log（fake/suspect/real），不 gate；status="start" 时无条件转发到 alarm stream。
// 转发后 cardagg.alarm_handler case alarm.Fall 自动落库；verifier verdict 进 Evidence.fall_verdict 供审计。
//
// 未来 PR-段7：verifier verdict="ghost" 时不转发（真正成为 fall gate）。
func (tm *TrackManager) RecordRadarAlarm(a RadarFallAlarm) {
	tm.mu.Lock()
	cp := a
	tm.recentRadarAlarms[a.TMs] = &cp
	tm.evictOldRadarAlarms(a.TMs)

	// start: 跑 fall verifier 评分；end: 跳过评分但仍转发（让 cardagg auto-resolve）
	verdict := ""
	score := 0
	isStart := a.Status == "start" || a.Status == ""
	if isStart {
		result := tm.verifyRadarFall(a, a.TMs)
		tm.logFallVerify(a, result)
		verdict = result.Verdict
		score = result.Score
		switch result.Verdict {
		case "ghost":
			tm.fallVerifyGhostCount++
		case "suspect":
			tm.fallVerifySuspectCount++
		case "real":
			tm.fallVerifyRealCount++
		}
	}
	// 解锁后 emit（emit 内部走 redis publish，不能持锁；tm.aiPublisher 自身 thread-safe）
	tm.mu.Unlock()

	if tm.aiPublisher == nil {
		return // playback / 测试场景
	}

	// emit status：start/instant 触发新 alarm；end 让 cardagg AlarmRouter 按 EndPolicy 关 alarm。
	// qinglan 收到 firmware Initialization (last_pose=2/7) 时 forward status="end"，必须透传到 alarm stream。
	emitStatus := a.Status
	if emitStatus == "" {
		emitStatus = "start"
	}

	// 构造 forward payload。track 信息从 firmware 抄；位置/HR/RR 现阶段不从 firmware 携带。
	payload := AIPayload{
		DeviceAddr:  a.DeviceUID, // = canonical IPv6 string
		RoomID:      tm.roomID,
		EventStatus: emitStatus,
		Track: observation.Track{
			BedStatus: observation.BedStatusUnchanged,
			TrackID:   a.TrackID,
			Pose:      a.Pose,
		},
		Reason: "firmware_radar_fall",
		Evidence: map[string]interface{}{
			"context":         "qinglan_publisher_event_stream_passthrough",
			"firmware_status": a.Status,
			"fall_verdict":    verdict, // end 时为空（无评分）
			"fall_score":      score,
		},
	}
	// envelope.Category：Fall / SittingOnGround（qinglan 已 collapse Suspected→Confirmed）
	emitCat := a.Category
	if emitCat == "" {
		emitCat = alarm.Fall // 兜底：旧路径无 Category 字段时
	}
	tm.aiPublisher.PublishAIAlarm(context.Background(), payload, emitCat, a.TMs)
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
		// PR-14 副作用 1：mark 当前 track 位置 cell 为 AreaBed
		tm.markRadarInBedCell(e)
		// PR-14 副作用 2 + Layer-0 床身份（2026-06-04）：
		//   ±15s 同事件     → RadarInBedConfirmedMs（最高置信档；多床房=床身份关联器）
		//   单床房 bedCount==1 → radar 与 sleepad 必同一张床，session 内 radar InBed 无视时差即记印证
		within15s := false
		if tm.lastSleepadInBedMs > 0 {
			delta := e.TMs - tm.lastSleepadInBedMs
			if delta < 0 {
				delta = -delta
			}
			within15s = delta <= bedInBedConsistencyMs
		}
		for _, s := range tm.bedSessions {
			if s.InBedSinceMs == 0 {
				continue
			}
			if within15s {
				if s.RadarInBedConfirmedMs == 0 {
					s.RadarInBedConfirmedMs = e.TMs
				}
				s.RadarInBedInSessionMs = e.TMs
			} else if tm.bedCount == 1 {
				s.RadarInBedInSessionMs = e.TMs
			}
		}
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
	// ExitRoom 事件 → 取消所有挂起的 lost-fall（人正常走出房间，不再悬念）
	// 注：silent fall 不取消（其语义是床上方遮挡，与 ExitRoom 无关）
	if e.EventName == alarm.ExitRoom && len(tm.pendingLostFalls) > 0 {
		for pid, p := range tm.pendingLostFalls {
			tm.lostFallPendingCancelled++
			tm.logger.Info("lost_fall_cancelled_by_exit_room",
				zap.String("device_uid", p.DeviceAddr),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("pending_age_ms", e.TMs-p.DisappearMs),
				zap.Int64("exit_room_ms", e.TMs),
			)
			tm.emitDecision(p.DeviceAddr, p.LogicID, p.OriginalTrackID, p.LastRawH, p.LastRawV, p.LastRawZ,
				"lostfall_cancel", "exit_room", map[string]interface{}{
					"pending_age_ms": e.TMs - p.DisappearMs, "exit_room_ms": e.TMs,
				}, e.TMs)
			delete(tm.pendingLostFalls, pid)
		}
	}
	// pendingLostFall 取消仅依赖 ExitRoom event / 多人入屋 / birth-recovery 三路径——
	// 裸 np=0 永不取消（盲区/水气丢信号会假报 np=0）。np=0 仅在 bathroom_fall 里
	// 与「最后帧在门区」合取后推断离开（见 evaluateLostFallStrong），不在此单独成立。
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

// CurrentNumberPeople 单 np latch:最近房内人数 + 是否新鲜(TTL 内)。count=-1/fresh=false=从未上报。
// 审查51 NP-1:belief shadow tick 读它喂 ObsNumberPeople(弱 corroboration,R5 不入 SFallen)。自锁。
func (tm *TrackManager) CurrentNumberPeople(nowMs int64) (count int, fresh bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.lastNumberPeopleTs == 0 || nowMs-tm.lastNumberPeopleTs > beliefNumberPeopleTTLMs {
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
	MoveActive       bool   // 本次快照是否"非静止"（StillSince==0 OR LastObservedMs == nowMs）
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
			MoveActive:   ts.StillSince == 0 || ts.LastObservedMs == nowMs,
			SleepadInBed: sleepadInBed,
		}
		// StillSec：StillSince==0 表示非静止；否则 (nowMs - StillSince) / 1000。
		if ts.StillSince > 0 && nowMs > ts.StillSince {
			base.StillSec = int((nowMs - ts.StillSince) / 1000)
		}
		// StillBoxSec：still-box run 时长（30s 滚动 50×50 方框，抗质心抖动）。fall 判据优先用此。
		if ts.StillBoxRunStart > 0 && nowMs > ts.StillBoxRunStart {
			base.StillBoxSec = int((nowMs - ts.StillBoxRunStart) / 1000)
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

// BedSessionLatch BedSession 状态对外快照（PR-11 BedroomFallRules 消费）。
// 仅含 fall 决策必要字段，不暴露 BedSession 全部 internal 状态。
type BedSessionLatch struct {
	DeviceUID             string
	InBedSinceMs          int64
	LeftBedAtMs           int64
	HasHRRR               bool
	LeftBedHadHRRR        bool
	RadarInBedConfirmedMs int64
	SilentFallAlerted     bool
}

// SnapshotBedSessions BedSession 状态值拷贝快照。
// 用途：PR-11 BedroomFallRules.Evaluate 读 LeftBedAtMs 作 bedside_fall 起点。
// 锁语义：内部加 tm.mu；返回值是值拷贝，调用方可在锁外读。
func (tm *TrackManager) SnapshotBedSessions() []BedSessionLatch {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	out := make([]BedSessionLatch, 0, len(tm.bedSessions))
	for _, s := range tm.bedSessions {
		out = append(out, BedSessionLatch{
			DeviceUID:             s.DeviceUID,
			InBedSinceMs:          s.InBedSinceMs,
			LeftBedAtMs:           s.LeftBedAtMs,
			HasHRRR:               s.HasHRRR,
			LeftBedHadHRRR:        s.LeftBedHadHRRR,
			RadarInBedConfirmedMs: s.RadarInBedConfirmedMs,
			SilentFallAlerted:     s.SilentFallAlerted,
		})
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
			// 新 track 出生 — 是否有 pending lost-fall 在等候 → 人从盲区返回；取消 + 学习盲区出口
			recoveredFromLost := tm.cancelPendingLostFallByBirth(f.X, f.Y, f.TMs)

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
			// 盲区返回路径：直接给 Real verdict，绕过 ghost 检查
			// （这是漏报场景下人从盲区返回的预期行为：track 看似凭空出现但实为真人）
			if recoveredFromLost {
				ts.Verdict = VerdictReal
				ts.Score = ScoreConfirmTh
				ts.BirthScore = ScoreConfirmTh
				ts.BirthReason = "recovered_from_lost_fall"
				ts.GhostPenalty = 0
				ts.ConfirmedAtMs = f.TMs
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
			tm.updateLieStateMachine(ts, f.Pose, f.X, f.Y, f.Z, nowMs)

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
			// 消失判定：lost fall（按 cell areaType 分时长，verdict 未定也算）
			// PR-14：旧 Path 1（track 消失 + 60s 复现窗口）已删除——
			// silent fall 仅由 BedSession LeftBed 矛盾路径触发（见 scanSilentFallLeftBed）。
			//
			// dedup：若该 track 已报过 bedside_fall（R4 床边晕倒，track 仍活时报），
			// 跳过 lost_fall pending 入池——避免"15min static→bedside_fall fire→track
			// 失锁→lost_fall 又 fire"的同事件双报。
			if ts.BedsideFallReported {
				tm.logger.Info("lost_fall_pending_skipped_after_bedside_fall",
					zap.String("device_uid", ts.DeviceAddr),
					zap.Int("track_id", ts.TrackID),
					zap.Int64("ts_ms", nowMs),
				)
			} else if tm.hasOtherLiveTrackWithLogicID(ts.LogicID, id) {
				// 换ID守恒闸：该 track_id 失锁，但其 logic_id 仍活在另一条 track 上 =
				// 同一逻辑目标被 firmware 换了 ID（数量守恒，无人真消失，且非进出事件）→ 不入 pending。
				// 铁律：不读 ghost verdict，纯靠身份连续性判"换ID游戏 vs 真消失"。
				tm.lostFallPendingCancelled++
				tm.logger.Info("lost_fall_skipped_id_swap",
					zap.String("device_uid", ts.DeviceAddr),
					zap.Int("track_id", ts.TrackID),
					zap.String("logic_id", ts.LogicID),
					zap.Int64("ts_ms", nowMs),
				)
				tm.emitDecision(ts.DeviceAddr, ts.LogicID, id, ts.LastRawH, ts.LastRawV, ts.LastRawZ,
					"lostfall_suppress", "id_swap_logicid_alive", nil, nowMs)
			} else if tm.roomLedgerEmpty() {
				// 空房账闸：enter/exit 守恒下房间已空（最近 ExitRoom/np=0 晚于 EnterRoom）→
				// 此刻失锁的 track 是"人走后残影"（如冻住的镜面反射）→ 抑制。治 D5F7 残影 ExitRoom 后才消失漏 cancel。
				tm.lostFallPendingCancelled++
				tm.logger.Info("lost_fall_skipped_room_empty",
					zap.String("device_uid", ts.DeviceAddr),
					zap.Int("track_id", ts.TrackID),
					zap.String("logic_id", ts.LogicID),
					zap.Int64("last_enter_ms", tm.lastEnterMs),
					zap.Int64("last_exit_ms", tm.lastExitMs),
					zap.Int("last_np", tm.lastNumberPeople), zap.Int64("last_np_ms", tm.lastNumberPeopleTs),
					zap.Int64("ts_ms", nowMs),
				)
				tm.emitDecision(ts.DeviceAddr, ts.LogicID, id, ts.LastRawH, ts.LastRawV, ts.LastRawZ,
					"lostfall_suppress", "room_ledger_empty", nil, nowMs)
			} else if (ts.Verdict == VerdictReal || ts.Verdict == VerdictPending) && tm.checkLostFall(ts) {
				// PR-9: v1 NumberPeople=0 ExitRoom 兜底（依赖 lastNumberPeopleZeroMs）已删除。
				// PR-10 BathroomLostFall + PR-11 bedroom lost_fall 将以 SuiteCensus.BathroomCount
				// + SuitePerson.AnchorRoomType 作离场判定主源；D523 firmware-specific 兜底（如仍需）
				// 走 cardagg ExitRoom event 路径（不再 sensor 内置）。
				pxF, pyF := ts.Kalman.Position()
				px := int(math.Round(pxF))
				py := int(math.Round(pyF))
				cell := tm.grid.CellAt(px, py)
				cellArea := AreaUnknown
				if cell != nil {
					cellArea = cell.Belief[0].Type
				}
				spatialJump := ts.MaxImpliedSpeedFromBirth > FallRulesParam.Lost.SuspectSpeedCm
				tm.pendingLostFalls[id] = &PendingLostFall{
					OriginalTrackID: id,
					DeviceAddr:      ts.DeviceAddr,
					RoomID:          ts.RoomID,
					LogicID:         ts.LogicID,
					LastX:           px,
					LastY:           py,
					LastZ:           ts.LastZ,
					LastRawH:        ts.LastRawH,
					LastRawV:        ts.LastRawV,
					LastRawZ:        ts.LastRawZ,
					LastScore:       ts.Score,
					LastVerdict:     ts.Verdict,
					LastCellArea:    cellArea,
					DisappearMs:     nowMs,
					StillBoxStartMs: ts.StillBoxRunStart,
					SpatialJump:     spatialJump,
				}
				tm.lostFallPendingCreated++
				tm.emitDecision(ts.DeviceAddr, ts.LogicID, id, ts.LastRawH, ts.LastRawV, ts.LastRawZ,
					"lostfall_pending", "track_lost", map[string]interface{}{
						"verdict": int(ts.Verdict), "score": ts.Score, "cell_area": int(cellArea),
						"spatial_jump": spatialJump, "still_box_start_ms": ts.StillBoxRunStart,
					}, nowMs)
			} else if ts.Verdict == VerdictReal {
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
					// PR5b: track_verdict（事后裁决）—— Real → Ghost 翻转
					reason := ts.BirthReason
					if reason == "" {
						reason = "ghost_penalty_accumulated_post_real"
					}
					tm.emitGhostVerdict(ts, ReasonGhostPostReal, reason, nowMs)
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
			// PR5b: track_verdict（事后裁决）—— 主路径 penalty 累积
			tm.emitGhostVerdict(ts, ReasonGhostPenalty, reason, nowMs)
			if !ts.LoggedGhost {
				pxF, pyF := ts.Kalman.Position()
				tm.logger.Info("track_verdict_ghost",
					zap.String("device_uid", ts.DeviceAddr),
					zap.Int("track_id", ts.TrackID),
					zap.String("verdict", "ghost"),
					zap.Int("score", ts.Score),
					zap.Int("birth_score", ts.BirthScore),
					zap.Int("ghost_penalty", ts.GhostPenalty),
					zap.String("reason", reason),
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
				// PR5b: track_verdict（事后裁决）—— score-based 路径
				tm.emitGhostVerdict(ts, ReasonGhostLowScore, reason, nowMs)
				if !ts.LoggedGhost {
					pxF, pyF := ts.Kalman.Position()
					tm.logger.Info("track_verdict_ghost",
						zap.String("device_uid", ts.DeviceAddr),
						zap.Int("track_id", ts.TrackID),
						zap.String("verdict", "ghost"),
						zap.Int("score", ts.Score),
						zap.Int("birth_score", ts.BirthScore),
						zap.String("reason", reason),
						zap.Int("x", int(math.Round(pxF))),
						zap.Int("y", int(math.Round(pyF))),
						zap.Int64("ts_ms", nowMs),
					)
					ts.LoggedGhost = true
				}
			}
		}
	}

	// PR-14: 旧 Path 1（pendingSilentFalls 60s 超时报警）已删除。
	// silent fall 现在仅通过 BedSession LeftBed 矛盾路径触发（scanSilentFallLeftBed）。
	results := make([]TrackOutput, 0, len(tm.tracks)+len(tm.pendingLostFalls))

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

	// ========== 段 4b: 扫挂起的 lost fall，按 cell-area-typed wait + StillBox 静止 credit 超时即报 ==========
	// 取消条件已在他处处理（cancelPendingByBirth / handleExitRoom / numberPeopleCancel）。
	// 此处仅扫超时触发：等待时间到达且未被取消 → 报 lost fall。
	isRiskTime := IsNightTime(nowMs, tm.timezone)
	for pid, p := range tm.pendingLostFalls {
		waitMs := tm.lostFallWaitMs(p, isRiskTime)
		if nowMs-p.DisappearMs < waitMs {
			continue
		}
		// 实时再检多人入屋（与 birth/event 触发的取消互补；保 segment 4b 兜底）
		if tm.realTrackCount() >= 2 {
			tm.lostFallPendingCancelled++
			tm.logger.Info("lost_fall_cancelled_by_multiple_real",
				zap.String("device_uid", p.DeviceAddr),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("nowMs", nowMs),
			)
			tm.emitDecision(p.DeviceAddr, p.LogicID, p.OriginalTrackID, p.LastRawH, p.LastRawV, p.LastRawZ,
				"lostfall_cancel", "multiple_real", nil, nowMs)
			delete(tm.pendingLostFalls, pid)
			continue
		}
		// 同房多雷达占用对账（**仅单床房 bedCount==1**）：另一台雷达近期仍见真人 → 人还在房
		// （被别台看着，本台只是丢/幻影）→ 抑制。治 D523 无床雷达持幻影、09E7 同房看到真人那类 lost_track FP。
		// 双床房不启用：可能两位老人，A 摔(lost)+B 在场会被误压成漏报（宁报勿漏；固件确认跌倒不受此守卫影响）。
		if tm.bedCount == 1 && tm.otherDeviceRealTrackRecent(p.DeviceAddr, nowMs) {
			tm.lostFallPendingCancelled++
			tm.logger.Info("lost_fall_cancelled_by_other_radar_in_room",
				zap.String("device_uid", p.DeviceAddr),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("nowMs", nowMs),
			)
			tm.emitDecision(p.DeviceAddr, p.LogicID, p.OriginalTrackID, p.LastRawH, p.LastRawV, p.LastRawZ,
				"lostfall_cancel", "other_radar", nil, nowMs)
			delete(tm.pendingLostFalls, pid)
			continue
		}
		// 距离闸：丢轨点距雷达 > d_fall（≈500cm）→ 抑制。贴地/躺姿目标回波弱（加特兰 4T4R），
		// 跌倒有效检测半径远小于人员检测半径；超出 d_fall 处即便真摔，firmware 也读不出地面态、
		// 发不出 pose=5，lost_track 凭"消失"反推退化为结构性盲猜（治 D523 边缘 ~5.8m 丢轨误报）。
		// 平面距 = √(LastRawH²+LastRawV²)（raw 雷达本地坐标，雷达在本地系原点）。依据 doc/AI_fall_detect.md §3.7。
		if gate := FallRulesParam.Lost.DistanceGateCm; gate > 0 {
			if distCm := distInt(p.LastRawH, p.LastRawV, 0, 0); distCm > gate {
				tm.lostFallPendingCancelled++
				tm.logger.Info("lost_fall_suppressed_by_distance_gate",
					zap.String("device_uid", p.DeviceAddr),
					zap.Int("track_id", p.OriginalTrackID),
					zap.String("room_id", p.RoomID),
					zap.Int("dist_cm", distCm),
					zap.Int("gate_cm", gate),
					zap.Int64("nowMs", nowMs),
				)
				tm.emitDecision(p.DeviceAddr, p.LogicID, p.OriginalTrackID, p.LastRawH, p.LastRawV, p.LastRawZ,
					"lostfall_suppress", "distance_gate", map[string]interface{}{
						"dist_cm": distCm, "gate_cm": gate,
					}, nowMs)
				delete(tm.pendingLostFalls, pid)
				continue
			}
		}
		// 超时 → MarkFallEvent + 写输出（kind=engine_lost_fall）
		tm.lostFallReported++
		tm.emitDecision(p.DeviceAddr, p.LogicID, p.OriginalTrackID, p.LastRawH, p.LastRawV, p.LastRawZ,
			"lostfall_fire", "lost_track", map[string]interface{}{
				"verdict": int(p.LastVerdict), "fall_score": p.LastScore, "cell_area": int(p.LastCellArea),
				"spatial_jump": p.SpatialJump, "wait_ms": waitMs, "still_box_start_ms": p.StillBoxStartMs,
			}, nowMs)
		tm.grid.MarkFallEvent(p.LastX, p.LastY, nowMs)
		// PR5c: 发布到 iot:alarm:stream，category=alarm.Fall（cardagg 现有 Fall handler 接管）。
		// fall alarm 已是确认态——不再发 track_confidence/score（确信值无需信号），
		// 子类型通过 Reason 区分（"lost_track"），fall 严重度进 Evidence.fall_score。
		//
		// engine_lost_fall 的 nowMs 是"engine 推断时刻"（= 实际跌倒后 wait_ms 才确认）。
		// alarm.triggered_at 必须用"实际发生时刻"才能让用户在列表按时间找到这条 fall，
		// 否则列表显示 08:35 / replay 中心 08:30 二者对不上。
		// anchor：still_box_start_ms 优先（track 静止 ≈ 跌倒邻近时刻），
		// 否则 nowMs - waitMs - 30s（最后活跃时刻 -30s 给 context）。
		// engine 推断时刻 nowMs 作 envelope.Timestamp → alarm.alerted_at；anchor 作 IncidentMs → triggered_at。
		replayAnchorMs := p.StillBoxStartMs
		if replayAnchorMs <= 0 {
			replayAnchorMs = nowMs - waitMs - 30_000
		}
		tm.emitAIAlarm(AIPayload{
			DeviceAddr: p.DeviceAddr,
			RoomID:     p.RoomID,
			IncidentMs: replayAnchorMs,
			Track: observation.Track{
				BedStatus: observation.BedStatusUnchanged,
				TrackID:   p.OriginalTrackID,
				PositionX: intPtr(p.LastRawH),
				PositionY: intPtr(p.LastRawV),
				PositionZ: intPtr(p.LastRawZ),
			},
			Reason: ReasonLostTrack,
			Evidence: map[string]interface{}{
				"context":            "track_lost_no_exit_room_no_recovery",
				"fall_score":         p.LastScore,
				"still_box_start_ms": p.StillBoxStartMs,
				"spatial_jump":       p.SpatialJump,
				"cell_area_type":     int(p.LastCellArea),
				"wait_ms":            waitMs,
				"last_verdict":       int(p.LastVerdict),
				"replay_anchor_ms":   replayAnchorMs,
				"engine_fire_ms":     nowMs, // 引擎推断时刻（审计）
			},
		}, alarm.Fall, nowMs) // envelope.Timestamp = 决策时刻 → alerted_at；IncidentMs = anchor → triggered_at
		tm.logger.Info("real_fall",
			zap.String("device_uid", p.DeviceAddr),
			zap.String("device_uid_hex", tm.devUIDHex(p.DeviceAddr)),
			zap.Int("track_id", p.OriginalTrackID),
			zap.String("kind", "engine_lost_fall"),
			zap.Int("score", p.LastScore),
			zap.Int("risk", 100),
			zap.String("reason", "track_lost_no_exit_room_no_recovery"),
			zap.Int("cell_area_type", int(p.LastCellArea)),
			zap.Int64("still_box_start_ms", p.StillBoxStartMs),
			zap.Bool("spatial_jump", p.SpatialJump),
			zap.Int64("wait_ms", waitMs),
			zap.Int("x", p.LastX), zap.Int("y", p.LastY), zap.Int("z", p.LastZ),
			zap.Int64("anchor_ms", replayAnchorMs), // = alarm.triggered_at
			zap.Int64("ts_ms", nowMs),              // engine 推断时刻
			zap.String("ts_human", time.UnixMilli(nowMs).Format("15:04:05.000")),
		)
		out := TrackOutput{
			TrackID:    p.OriginalTrackID,
			DeviceAddr: p.DeviceAddr,
			RoomID:     p.RoomID,
			Verdict:    p.LastVerdict,
			Score:      p.LastScore,
			Risk:       100,
			Anomaly:    AnomalyFall,
			X:          p.LastX,
			Y:          p.LastY,
			Z:          p.LastZ,
			Source:     "engine_lost",
		}
		results = append(results, out)
		tm.outputs[p.OriginalTrackID] = &out
		delete(tm.pendingLostFalls, pid)
	}

	// ========== 段 4c: 新版 Silent Fall（sleepad LeftBed + radar 仍在 Bed 邻域） ==========
	// 触发：bedSession.LeftBedAtMs > 0，等待 60s（vital + 单人）/ 120s（其它），
	//       超时仍有任一活 track 在 AreaBed ±BedNeighborhood cm 内 → 矛盾报警。
	// 否则取消（人正常起床，radar 也离床区）。
	tm.latchRadarTrackForBedSessions(nowMs) // Layer-1：在床期间持续记录 radar 是否看到人
	results = append(results, tm.scanSilentFallLeftBed(nowMs)...)

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

		stillSec := 0
		if ts.StillSince > 0 {
			stillSec = int((nowMs - ts.StillSince) / 1000)
		}

		source := "radar_direct"

		out := TrackOutput{
			TrackID:    ts.TrackID,
			DeviceAddr: ts.DeviceAddr,
			RoomID:     ts.RoomID,
			Verdict:    ts.Verdict,
			Score:      ts.Score,
			Risk:       tm.computeRisk(ts, stillSec, nowMs),
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

// hasPendingLostFallInRoom 当前房间是否有挂起的 lost-fall（与 RoomID 等价于 tm.roomID）。
// 用于 birth verdict bypass：人从盲区返回时新 track 不该被判 ghost。
// 调用方持锁。
func (tm *TrackManager) hasPendingLostFallInRoom() bool {
	return len(tm.pendingLostFalls) > 0
}

// cancelPendingLostFallByBirth 新 track 出生时尝试取消挂起的 lost-fall。
//
// 不限位置 + 不限 birth pose（人从盲区任何角度返回都算）。
// 命中时：① 取消 pending ② cell.IncrBlindSpotRecovery（学习盲区出口） ③ 累计 cancelled 统计。
//
// 返回是否命中（调用方据此决定是否给新 track verdict 直接 Real / 加分）。
func (tm *TrackManager) cancelPendingLostFallByBirth(birthX, birthY int, nowMs int64) bool {
	if len(tm.pendingLostFalls) == 0 {
		return false
	}
	// 单房间一般只有 1 个 pending；遍历全部，全部取消（人回来 = 所有 lost 候选都解除）
	hit := false
	for pid, p := range tm.pendingLostFalls {
		hit = true
		tm.lostFallPendingCancelled++
		tm.logger.Info("lost_fall_cancelled_by_recovery",
			zap.String("device_uid", p.DeviceAddr),
			zap.Int("track_id", p.OriginalTrackID),
			zap.String("room_id", p.RoomID),
			zap.Int64("pending_age_ms", nowMs-p.DisappearMs),
			zap.Int("recovered_at_x", birthX), zap.Int("recovered_at_y", birthY),
		)
		delete(tm.pendingLostFalls, pid)
	}
	if hit {
		// 学习盲区出口：人在新 track 出生位置 cell 累计一个 BlindSpotRecovery
		tm.grid.MarkBlindSpotRecovery(birthX, birthY, nowMs)
	}
	return hit
}

// scanSilentFallLeftBed 扫描所有 BedSession，对到达等待窗的 LeftBed 进行裁决：
//
//	radar 仍在 Bed 邻域 → silent fall（矛盾，疑似跌倒地面 / 床边）
//	radar 已离开 Bed 邻域 → 取消（人正常起床走开）
//
// Bed 邻域定义：任一 active track 距离最近 AreaBed cell ≤ FallRulesParam.Silent.BedNeighborhood cm。
// 调用方持锁。
func (tm *TrackManager) scanSilentFallLeftBed(nowMs int64) []TrackOutput {
	if len(tm.bedSessions) == 0 {
		return nil
	}
	param := FallRulesParam.Silent
	var out []TrackOutput
	for _, s := range tm.bedSessions {
		if s.LeftBedAtMs == 0 || s.SilentFallAlerted {
			continue
		}
		// 武装放宽（2026-06-04）：sleepad InBed 单独即可进入等待窗（接触式正向证据，
		// "sleepad 未报上床不否定上床"）。radar 印证度不再作硬闸，改由 InBedConfidence
		// 在 vanish-fire 分支门控（radar 全程无 track 的低置信=可能 EMI 假上床 → 不发）。
		waitSec := param.WaitNoVitalSec
		if s.LeftBedHadHRRR && s.LeftBedMaxPeople == 1 {
			waitSec = param.WaitVitalSec
		}
		if nowMs-s.LeftBedAtMs < int64(waitSec)*1000 {
			continue
		}
		// 等待窗满 — 检查 radar 是否仍在 Bed 邻域
		if tm.anyActiveTrackNearBed(param.BedNeighborhood) {
			// 矛盾 → 候选 silent fall；但需先过 AreaBed 人工标定躺区豁免（详 fall_exempt.go）
			x, y, z, rawH, rawV, rawZ, scoreVal, verdict, deviceAddr := tm.pickActiveTrackNearBed(param.BedNeighborhood)
			if isHumanBedAt(tm.grid, x, y) {
				// 人就在人工标定的床上 — 不报；按 cancel 路径收尾（sleepad miscalibration 假阳）
				tm.silentFallLeftbedCancelled++
				tm.logger.Info("silent_fall_leftbed_exempt_human_bed",
					zap.String("sleepad_uid", s.DeviceUID),
					zap.String("device_uid", deviceAddr),
					zap.Int64("leftbed_ms", s.LeftBedAtMs),
					zap.Int("x", x), zap.Int("y", y),
				)
				s.SilentFallAlerted = true
				s.InBedSinceMs = 0
				continue
			}
			// 矛盾 → 报 silent fall
			s.SilentFallAlerted = true
			tm.silentFallLeftbedReported++
			tm.grid.MarkFallEvent(x, y, nowMs)
			// PR5c: alarm.Fall + Reason 区分子类型；BedStatus=1 反映 sleepad LeftBed 触发。
			// fall 已确认，不发 track_confidence；fall 严重度进 Evidence.fall_score。
			tm.emitAIAlarm(AIPayload{
				DeviceAddr: deviceAddr,
				RoomID:     tm.roomID,
				IncidentMs: s.LeftBedAtMs,
				Track: observation.Track{
					PositionX: intPtr(rawH),
					PositionY: intPtr(rawV),
					PositionZ: intPtr(rawZ),
					BedStatus: observation.BedStatusLeftBed, // sleepad 报 LeftBed
				},
				Reason: ReasonSleepadRadarConflict,
				Evidence: map[string]interface{}{
					"context":          "sleepad_leftbed_radar_still_on_bed",
					"fall_score":       scoreVal,
					"radar_verdict":    int(verdict),
					"sleepad_uid":      s.DeviceUID,
					"had_hr_rr":        s.LeftBedHadHRRR,
					"max_people":       s.LeftBedMaxPeople,
					"wait_sec":         waitSec,
					"leftbed_ms":       s.LeftBedAtMs,
					"replay_anchor_ms": s.LeftBedAtMs, // sleepad LeftBed 时刻 = 跌倒矛盾起点
					"engine_fire_ms":   nowMs,         // 引擎确认时刻（审计）
				},
			}, alarm.Fall, nowMs) // envelope.Timestamp = 决策时刻 → alerted_at；IncidentMs = LeftBed → triggered_at
			tm.logger.Info("real_fall",
				zap.String("device_uid", deviceAddr),
				zap.String("device_uid_hex", tm.devUIDHex(deviceAddr)),
				zap.String("kind", "engine_silent_leftbed"),
				zap.String("sleepad_uid", s.DeviceUID),
				zap.Int("score", scoreVal),
				zap.Int("risk", 100),
				zap.String("reason", "sleepad_leftbed_radar_still_on_bed"),
				zap.Bool("had_hr_rr", s.LeftBedHadHRRR),
				zap.Int("max_people", s.LeftBedMaxPeople),
				zap.Int("wait_sec", waitSec),
				zap.Int64("leftbed_ms", s.LeftBedAtMs),
				zap.Int("x", x), zap.Int("y", y), zap.Int("z", z),
				zap.Int64("anchor_ms", s.LeftBedAtMs), // = alarm.triggered_at
				zap.Int64("ts_ms", nowMs),             // engine 推断时刻
				zap.String("ts_human", time.UnixMilli(nowMs).Format("15:04:05.000")),
			)
			out = append(out, TrackOutput{
				DeviceAddr: deviceAddr,
				RoomID:     tm.roomID,
				Verdict:    verdict,
				Score:      scoreVal,
				Risk:       100,
				Anomaly:    AnomalyFall,
				X:          x,
				Y:          y,
				Z:          z,
				Source:     "engine_silent_leftbed",
			})
		} else if tm.anyActiveTrack() {
			// 人起身在房内走动，radar 看得见 = 已交代 → 取消
			tm.silentFallLeftbedCancelled++
			tm.logger.Info("silent_fall_leftbed_cancelled",
				zap.String("reason", "active_track_in_room"),
				zap.String("sleepad_uid", s.DeviceUID),
				zap.Int64("leftbed_ms", s.LeftBedAtMs),
			)
			s.SilentFallAlerted = true
			s.InBedSinceMs = 0
		} else if tm.roomLedgerEmpty() {
			// 人过门走了（ExitRoom 空间证据）= 正常离开 → 取消
			tm.silentFallLeftbedCancelled++
			tm.logger.Info("silent_fall_leftbed_cancelled",
				zap.String("reason", "exited_room"),
				zap.String("sleepad_uid", s.DeviceUID),
				zap.Int64("leftbed_ms", s.LeftBedAtMs),
			)
			s.SilentFallAlerted = true
			s.InBedSinceMs = 0
		} else if s.InBedConfidence >= bedVanishMinConf && s.RadarDeviceAddr != "" {
			// vanish-fire：sleepad 证明离床 + 房内无 track + 未过门 + 床态置信足够
			//   = 人离床后在房内消失（既没走出门、雷达也找不到）→ 摔进遮挡死角（床后/床下）。
			// 治本 Case1 那类"干净摔进死角"：lost_fall 被静止前置守卫挡掉，唯有 sleepad
			// LeftBed 正向证据能锚定。低置信（radar 全程没看到人=可能 EMI 假上床）走 else 不发。
			s.SilentFallAlerted = true
			tm.silentFallLeftbedReported++
			tm.emitAIAlarm(AIPayload{
				DeviceAddr: s.RadarDeviceAddr,
				RoomID:     tm.roomID,
				IncidentMs: s.LeftBedAtMs,
				Track:      observation.Track{BedStatus: observation.BedStatusLeftBed},
				Reason:     ReasonSleepadRadarConflict,
				Evidence: map[string]interface{}{
					"context":           "sleepad_leftbed_vanished_no_exit",
					"in_bed_confidence": s.InBedConfidence,
					"sleepad_uid":       s.DeviceUID,
					"had_hr_rr":         s.LeftBedHadHRRR,
					"wait_sec":          waitSec,
					"leftbed_ms":        s.LeftBedAtMs,
					"replay_anchor_ms":  s.LeftBedAtMs,
					"engine_fire_ms":    nowMs,
				},
			}, alarm.Fall, nowMs)
			tm.logger.Info("real_fall",
				zap.String("device_uid", s.RadarDeviceAddr),
				zap.String("kind", "engine_silent_leftbed_vanished"),
				zap.String("sleepad_uid", s.DeviceUID),
				zap.Int("in_bed_confidence", s.InBedConfidence),
				zap.Int("risk", 100),
				zap.String("reason", "sleepad_leftbed_vanished_no_exit"),
				zap.Bool("had_hr_rr", s.LeftBedHadHRRR),
				zap.Int("wait_sec", waitSec),
				zap.Int64("leftbed_ms", s.LeftBedAtMs),
				zap.Int64("anchor_ms", s.LeftBedAtMs),
				zap.Int64("ts_ms", nowMs),
				zap.String("ts_human", time.UnixMilli(nowMs).Format("15:04:05.000")),
			)
			out = append(out, TrackOutput{
				DeviceAddr: s.RadarDeviceAddr,
				RoomID:     tm.roomID,
				Verdict:    VerdictReal,
				Risk:       100,
				Anomaly:    AnomalyFall,
				Source:     "engine_silent_leftbed_vanished",
			})
			s.InBedSinceMs = 0
		} else {
			// 床态置信不足（radar 全程没看到人=可能震动/EMI 假上床）或无雷达归因 → 不发，取消
			tm.silentFallLeftbedCancelled++
			tm.logger.Info("silent_fall_leftbed_cancelled",
				zap.String("reason", "low_confidence_or_no_radar"),
				zap.Int("in_bed_confidence", s.InBedConfidence),
				zap.String("sleepad_uid", s.DeviceUID),
				zap.Int64("leftbed_ms", s.LeftBedAtMs),
			)
			s.SilentFallAlerted = true
			s.InBedSinceMs = 0
		}
	}
	return out
}

// anyActiveTrackNearBed 是否有任一 active track 距 AreaBed cell ≤ marginCm。
// 调用方持锁。
func (tm *TrackManager) anyActiveTrackNearBed(marginCm int) bool {
	for _, t := range tm.tracks {
		pxF, pyF := t.Kalman.Position()
		px, py := int(math.Round(pxF)), int(math.Round(pyF))
		if tm.grid.IsNearPriorType(px, py, AreaBed, marginCm) {
			return true
		}
	}
	return false
}

// pickActiveTrackNearBed 取最近 Bed 的 active track 信息（用于告警坐标）。
// 找不到时返回 (0,0,0, 0, VerdictReal, "")（此时调用方已确认 anyActiveTrackNearBed=true，理论上不会走 fallback）。
// 返回 canvas 坐标（x/y 给 grid 用）+ raw radar 坐标（rawH/rawV/rawZ 给 alarm publish 用，对外契约不变）。
func (tm *TrackManager) pickActiveTrackNearBed(marginCm int) (x, y, z, rawH, rawV, rawZ, score int, verdict TrackVerdict, deviceAddr string) {
	for _, t := range tm.tracks {
		pxF, pyF := t.Kalman.Position()
		px, py := int(math.Round(pxF)), int(math.Round(pyF))
		if tm.grid.IsNearPriorType(px, py, AreaBed, marginCm) {
			return px, py, t.LastZ, t.LastRawH, t.LastRawV, t.LastRawZ, t.Score, t.Verdict, t.DeviceAddr
		}
	}
	return 0, 0, 0, 0, 0, 0, 0, VerdictReal, ""
}

// LostFallStats lost-fall 统计快照（含三类 cancel 来源汇总）
type LostFallStats struct {
	PendingCreated   int // 进入挂起的总数
	PendingCancelled int // 取消累计（含 birth-recovery / ExitRoom / 多人入屋）
	Reported         int // 超时真报的 lost fall 数量
	Outstanding      int // 当前仍在挂起池中
}

// LostFallStatsSnapshot 返回 lost-fall 累计统计
func (tm *TrackManager) LostFallStatsSnapshot() LostFallStats {
	return LostFallStats{
		PendingCreated:   tm.lostFallPendingCreated,
		PendingCancelled: tm.lostFallPendingCancelled,
		Reported:         tm.lostFallReported,
		Outstanding:      len(tm.pendingLostFalls),
	}
}

// SilentFallLeftBedStats 新版 silent fall（sleepad LeftBed + radar 仍在 bed）统计
type SilentFallLeftBedStats struct {
	Reported  int // 报警次数
	Cancelled int // 等待窗满但 radar 也离床 → 取消
}

func (tm *TrackManager) SilentFallLeftBedStatsSnapshot() SilentFallLeftBedStats {
	return SilentFallLeftBedStats{
		Reported:  tm.silentFallLeftbedReported,
		Cancelled: tm.silentFallLeftbedCancelled,
	}
}

// PR-Bootstrap: StillFallStats + stillFallReportCount 已删除（v1 fire path 删除后无 source）。
// PR-10 BathroomStillFall 的统计走 ai.log audit 路径，不再 per-TrackManager 计数。

// FallVerifyStats firmware Fall verifier 三档累计（PR-5）
type FallVerifyStats struct {
	Ghost   int
	Suspect int
	Real    int
}

func (tm *TrackManager) FallVerifyStatsSnapshot() FallVerifyStats {
	return FallVerifyStats{
		Ghost:   tm.fallVerifyGhostCount,
		Suspect: tm.fallVerifySuspectCount,
		Real:    tm.fallVerifyRealCount,
	}
}

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
			tm.logger.Info("ghost_motion_symmetry_hit",
				zap.String("device_uid", ts.DeviceAddr),
				zap.Int("track_id", ts.TrackID),
				zap.Int("ghost_penalty", ts.GhostPenalty),
				zap.Int64("ts_ms", nowMs),
			)
		} else if partner, mx, my, rx, ry, ok := tm.checkMirrorSymmetry(ts); ok {
			ts.GhostPenalty += 10
			ts.BirthReason = "mirror_image_of_real_track"
			bxF, byF := ts.Kalman.Position()
			tm.logger.Info("ghost_mirror_symmetry_hit",
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

	if d < StillThreshCm {
		// 静止
		if ts.StillSince == 0 {
			ts.StillSince = nowMs
			ts.StillX = x
			ts.StillY = y
		}
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
		// 静止超时（综合 cell history 的自适应阈值）
		// PR-9: 删除 bathroomRealCount caregiver 例外抑制 — Phase C 期 v1 抑制 dropped；
		// PR-10 BathroomStillFall 用 census.BathroomCount + AnchorRoomType 重新引入 caregiver 例外。
		if cell != nil {
			isRiskTime := IsNightTime(nowMs, tm.timezone)
			timeout := cell.EffectiveStillTimeoutSec(isRiskTime)
			if timeout > 0 {
				stillSec := int((nowMs - ts.StillSince) / 1000)
				if stillSec > timeout {
					ts.CurrentAnomaly = AnomalyStillTooLong
					if !ts.LongStillReported {
						tm.grid.MarkLongStill(x, y, nowMs)
						ts.LongStillReported = true
					}
				}
			}

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
		// 在动：刚从静止恢复，记 Dwell EMA
		if ts.StillSince > 0 {
			dwellSec := int((nowMs - ts.StillSince) / 1000)
			if dwellSec > 0 {
				tm.grid.MarkDwell(ts.StillX, ts.StillY, dwellSec, nowMs)
			}
		}
		// Cell history integral：之前曾被系统判为 long-still（LongStillReported=true）但 track 自己走了
		// → 系统判错（容忍证据），喂给该 cell 自动放宽未来阈值
		if ts.LongStillReported {
			tm.grid.MarkToleratedStill(ts.StillX, ts.StillY, nowMs)
		}
		ts.StillSince = 0
		ts.LongStillReported = false
		ts.StillFallReported = false
		ts.StandStaticSince = 0 // PR-7.2: track 移动 → reset stand-static 计时
		ts.LyingOnBedSinceMs = 0
		ts.SitOnToiletSinceMs = 0 // PR-11: track 移动 → reset 持续观测计时
		if ts.CurrentAnomaly == AnomalyStillTooLong {
			ts.CurrentAnomaly = AnomalyNone
		}
		// 速度合理性
		speed := int(math.Round(ts.Kalman.Speed()))
		if speed > 10 && speed < 150 {
			ts.AdjustScore(3)
		} else if speed >= 150 {
			ts.AdjustScore(-2)
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

func (tm *TrackManager) updateLieStateMachine(ts *TrackState, pose, x, y, z int, nowMs int64) {
	curCore := RadarPoseToCore(pose)
	prevCore := ts.PrevCore

	// 进入 Lie 态
	if curCore == CorePoseLie && prevCore != CorePoseLie {
		ts.LieEnteredAt = nowMs
		ts.LieEnteredX = x
		ts.LieEnteredY = y

		// 自推 Fall：前姿态是 Stand/Move 且 Z 骤降
		if (prevCore == CorePoseStand || prevCore == CorePoseMove) &&
			ts.LastZ > 50 && z < 20 {
			tm.grid.MarkFallEvent(x, y, nowMs)
			ts.CurrentAnomaly = AnomalyFall
			tm.logger.Info("real_fall",
				zap.String("device_uid", ts.DeviceAddr),
				zap.String("device_uid_hex", tm.devUIDHex(ts.DeviceAddr)),
				zap.Int("track_id", ts.TrackID),
				zap.String("kind", "engine_z_drop"),
				zap.Int("score", ts.Score),
				zap.Int("risk", 100),
				zap.String("reason", "stand_or_move_to_lie_with_z_drop"),
				zap.Int("x", x), zap.Int("y", y), zap.Int("z", z),
				zap.Int("prev_z", ts.LastZ),
				zap.Int64("ts_ms", nowMs),
			)
		}
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
// Lost Fall（cell-area-typed wait + ExitRoom + NumberPeople 兜底）
// ========================================================================

// checkLostFall 判断 track 消失是否符合 lost-fall 触发条件。
//
//	Real/Pending verdict + age≥5s + 离任一出口 > ExitDistMinCm（全屋通用）
//
// PR-14：旧 silent fall path（消失 + 60s 复现窗口）已删除；
// silent fall 现在仅通过 BedSession LeftBed 矛盾路径触发（scanSilentFallLeftBed）。
func (tm *TrackManager) checkLostFall(ts *TrackState) bool {
	if ts.AgeSec() < 5 {
		return false
	}
	// 走动前置（2026-06-01）：lost-fall 仅认"走动中突然消失"。消失前若已 settle 进长静止
	// （still-box 持续 ≥ MovingPreconditionMs）= 站洗手台/坐马桶/卧床不动 = 正常静止态，
	// 归 Still-fall 域（长阈值兜底），不进 lost-fall。box-based still-box 抗坐姿 jitter。
	// 治本：CABB/MoM 多条冻结站/坐 lost_track FP（无真跌倒）。
	if ts.StillBoxRunStart > 0 &&
		ts.LastObservedMs-ts.StillBoxRunStart >= int64(FallRulesParam.Lost.MovingPreconditionMs) {
		return false
	}
	pxF, pyF := ts.Kalman.Position()
	px := int(math.Round(pxF))
	py := int(math.Round(pyF))
	cell := tm.grid.CellAt(px, py)
	if cell == nil {
		return false
	}
	// 出门正常 — 在 Enter 区消失合法
	if cell.Belief[0].Type == AreaEnter {
		return false
	}
	// 离最近门 ≤ ExitDistMinCm（30cm）视为「贴在门口正常通过」；>30cm 即使无
	// ExitRoom 事件也进 lost-fall pending（elder care 宁可误报）。
	if tm.grid.NearestEntryDist(px, py) <= FallRulesParam.Lost.ExitDistMinCm {
		return false
	}
	return true
}

// lostFallWaitMs 计算 lost-fall pending 的等待时长（毫秒）。
//
// 基线按消失点 cell areaType：
//   - AreaBed / AreaSit：60min（睡觉/坐着雷达丢 track 常见）
//   - AreaToilet / AreaShower：与 still fall 同（risk-time / non-risk-time）
//   - AreaDeny / 其它：5min
//
// StillBox 静止 credit（box 判据）：失锁前 30s 内位移 box <= StillBoxCm 视为 still，
//
//	StillBoxStartMs 是该 still box run 起点，stillDur = DisappearMs - StillBoxStartMs，
//	credit = stillDur / 2 半计入等待。
//
// SpatialJump factor：track 表现过空间跳跃 → 等待时间 ×0.5（更敏感）
// 兜底：min EffectiveWaitFloorSec
func (tm *TrackManager) lostFallWaitMs(p *PendingLostFall, isRiskTime bool) int64 {
	pl := FallRulesParam.Lost
	var base int
	// bathroom 长档判据（与 bathroom still-fall 同时长）：
	//   ① 整房 room_type==Bathroom（权威标志，即便未画 Toilet/Shower 子区、cell_area=0 也放宽）；
	//   ② 否则按消失点 cell areaType ∈ {Toilet, Shower}。
	// 治本：CABB「叫 Bathroom 但无声明子区」lost_track FP——原 cell_area=0 落 Walkway 5min 误报。
	bathroomStd := tm.roomType == card.RoomTypeBathroom || p.LastCellArea == AreaToilet || p.LastCellArea == AreaShower
	switch {
	case bathroomStd:
		if isRiskTime {
			base = FallRulesParam.Still.ToiletShowerSec
		} else {
			base = int(float64(FallRulesParam.Still.ToiletShowerSec) * FallRulesParam.Still.NonRiskTimeFactor)
		}
	case p.LastCellArea == AreaBed || p.LastCellArea == AreaSit:
		base = pl.RestZoneWaitSec
	case p.LastCellArea == AreaDeny:
		base = pl.DenyZoneWaitSec
	default:
		base = pl.WalkwayWaitSec
	}

	// StillBox 静止 credit：half of still box duration counted toward wait
	if p.StillBoxStartMs > 0 {
		stillDurMs := p.DisappearMs - p.StillBoxStartMs
		if stillDurMs > 0 {
			base -= int(stillDurMs / 2 / 1000)
		}
	}

	// Spatial jump factor
	if p.SpatialJump {
		base = int(float64(base) * pl.SpatialJumpFactor)
	}

	if base < pl.EffectiveWaitFloorSec {
		base = pl.EffectiveWaitFloorSec
	}
	return int64(base) * 1000
}

// ========================================================================
// 综合风险分
// ========================================================================

func (tm *TrackManager) computeRisk(ts *TrackState, stillSec int, nowMs int64) int {
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
//     最近 30s History 滚动窗口内位移 box (max-min) ≤ StillBoxCm(30) → still。
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
	disp := ts.BoxRangeWithinMs(30_000, nowMs)
	if disp <= FallRulesParam.Lost.StillBoxCm && len(ts.History) >= 2 {
		if ts.StillBoxRunStart == 0 {
			// 起点回填到 History 最早帧（box 内最早可见点）
			ts.StillBoxRunStart = ts.History[0].TMs
		}
	} else {
		ts.StillBoxRunStart = 0
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
