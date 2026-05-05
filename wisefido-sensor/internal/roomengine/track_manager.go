package roomengine

import (
	"context"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
	"owl-common/alarm"
	"owl-common/observation"
	"owl-common/radarutils"
	"owl-common/roomutil"
)

// intPtr observation.Track 的 PositionX/Y/Z 是 *int（区分"没值"vs"=0"）。
func intPtr(v int) *int { return &v }

// TrackOutput Room Engine 对外输出的单条 track 评估结果
type TrackOutput struct {
	TrackID  int
	DeviceID string
	RoomID   string
	Verdict  TrackVerdict
	Score    int // [0,100]
	Risk     int // [0,100]
	Anomaly  Anomaly
	X, Y, Z  int // 当前估计位置（画布坐标 cm）
	VX, VY   int // cm/s
	StillSec int
	Source   string // "radar_direct" / "engine_silent_leftbed" / "engine_lost" / "engine_still_fall" / ...
}

// TrackFrame 一帧输入（已由 engine 层做完 RadarToCanvas 转换）
type TrackFrame struct {
	TrackID  int
	DeviceID string
	X, Y, Z  int // 画布坐标 cm
	Pose     int
	AreaType int // 雷达给的 area_id（保留兼容字段，engine 不信其判定，用 cell.AreaType）
	TMs      int64

	// 用于 frozen-frame 检测的辅助字段（每个值是 firmware 给的原始信号）
	// firmware 失锁后这些字段会保持 byte-equal 5+ 分钟，是判 frozen 的强证据
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
	// 时长按消失点 cell areaType（5min walkway / 60min bed / ...），含 frozen credit；
	// 取消条件：新 track 出生 / ExitRoom 事件 / room.NumberPeople ≥ 2。
	pendingLostFalls map[int]*PendingLostFall

	// bedSessions：sleepad 设备维度的"在床会话"状态机；新版 silent fall 触发源。
	// key = sleepad device_uid。详见 BedSession 结构体。
	bedSessions map[string]*BedSession

	// bathroomRealCount：当前帧 ProcessFrame 起点时，所在 cell 为 AreaToilet/AreaShower 的 Real track 数。
	// 用途：≥2 时视为护工陪同，scoreMovement 跳过 long-still 15min 超时报警。
	// 由 ProcessFrame 入口刷新，scoreMovement 读取（同步在锁内，无竞态）。
	bathroomRealCount int

	// sleepadStates：同房间 sleepad 设备的最新观测（device_uid → obs）
	// 由 ProcessSleepadObservation 写入。
	sleepadStates map[string]*SleepadObservation

	// bedPersonCount：每张床（device_uid → 人数）当前在床人数计数。
	// 由 ProcessSleepadBedEvent 增减：InBed +1，LeftBed -1（floor 0）。
	// 用途：bed-fall 触发前判断"仅 1 人"避免家属/护工陪同时误报。
	bedPersonCount map[string]int

	// Lost fall 统计
	lostFallPendingCreated   int
	lostFallPendingCancelled int // 含 birth-recovery / ExitRoom / NumberPeople 三类取消
	lostFallReported         int

	// Silent fall（LeftBed 矛盾路径）统计
	silentFallLeftbedReported  int
	silentFallLeftbedCancelled int // wait 满但 radar 已离开 Bed 邻域 → 取消（人正常起床）

	// Still fall 统计（bathroom + pose=Stand + 15/18min 持续静止）
	stillFallReportCount int

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

	// lastLeftBedAt：任意来源（sleepad event / sleepad bed_status 转换 / 未来 radar event）
	// 最近一次 LeftBed 事件的时间戳。R4（AnomalyBedsideFall）开窗判定用。
	// 0 = 从未有 LeftBed 事件（或上次 InBed 后又被 InBed 抹掉）。
	lastLeftBedAt int64

	// lastNumberPeopleZeroMs：最近一次 firmware number_people=0 事件时间戳。
	// 部分 firmware（如 D523）在 FOV 边角离场不发 ExitRoom，但会发 number_people=0
	// （实测早 track_id=88 心跳 36-44ms）。当 track 终止入 pendingLostFall 池前，
	// 检查 ±NumberPeopleZeroFallbackMs 窗口内有无 number_people=0 → 视作 ExitRoom 兜底，
	// 跳过 pending 创建（避免人正常离场误报 lost_fall）。同时也用于取消已存在的 pending。
	lastNumberPeopleZeroMs int64

	// bedsideFallCfg：R4（床边晕倒）参数；由 SetBedsideFallConfig 注入。
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

	// stayAlarmEnabled：本房间任一设备是否启用 Stay alarm（alarm_device.monitor_config Stay.is_enabled=1）。
	// 启用 = 运维明确表达「需要长时间静止检测」 → still fall 视为 bathroom 类型。
	// 与 cell.AreaToilet/Shower、room.name 三者取并集；任一命中即触发 still fall。
	// 由 engine.RegisterRoom 启动时注入（查 DB），运行时变更需重启或加 Redis 通道（暂未做）。
	stayAlarmEnabled bool

	// interferes：本房间镜面/反射区矩形（cfg.Interferes：mirror、glass-tv、metal、curtain）。
	// 用于因子 7（镜面对称 ghost 检测）：对当前 track 求关于 interfere 长轴的镜像位置，
	// 若另一 Real track 距镜像点 < 50cm → 当前 track 是其镜像 ghost，+60 penalty 直接判 Ghost。
	// 由 engine.RegisterRoom 调 SetInterferes 注入；nil 时因子 7 不参与。
	interferes []radarutils.Rect

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
//   1. 观测信号走 observation.Track —— 与 firmware/engine 上游同一 schema，
//      AI 只填它要表达的字段（如 ghost verdict 只填 TrackConfidence + 位置；
//      未来 vital 修正只填 HeartRate + VitalConfidence）。零值 = "AI 无意见"。
//   2. Source 是决策路径标识（一等公民）—— 不是 AI/非AI 二元 tag，而是具体的
//      "哪个算法路径产生的"，用于审计、调参、误报追溯、商业演示讲故事。
//      cardagg 落 alarm_events.metadata.source；前端可读、可 group by。
//   3. Reason / Evidence 是审计元数据 —— 解释"为什么 AI 这么判"，不参与下游
//      分支判定（分支判定由 Source 或 Track 数值阈值驱动）。
type AIPayload struct {
	DeviceID string // 源 sensor UUID（FK to devices.device_id）
	RoomID   string

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
}

// CategoryTrackVerdict 是 track verdict 的 category 路由键（事件 TYPE，不是 verdict label）。
const CategoryTrackVerdict = "track_verdict"

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
)

// AIPublisher PR-8 解耦：TrackManager 不直接持有 redis client，由 engine 实现接口注入。
type AIPublisher interface {
	PublishAIEvent(ctx context.Context, p AIPayload, category string, nowMs int64)
	PublishAIAlarm(ctx context.Context, p AIPayload, category string, nowMs int64)
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
		roomID:             roomID,
		grid:               grid,
		tracks:             make(map[int]*TrackState),
		outputs:            make(map[int]*TrackOutput),
		pendingLostFalls:   make(map[int]*PendingLostFall),
		bedSessions:        make(map[string]*BedSession),
		sleepadStates:      make(map[string]*SleepadObservation),
		bedPersonCount:     make(map[string]int),
		moveSpeedCms:       20, // 默认值（与 DefaultLearnParams.MoveSpeedCms 一致）
		bedsideFallCfg:     defaultBedsideFallCfg,
		recentRadarAlarms:  make(map[int64]*RadarFallAlarm),
		recentRadarEvents:  make(map[int64]*RadarTrackEvent),
		recentBufferMs:     5 * 60 * 1000, // 5 min
		logger:             zap.NewNop(),
		startupMs:          time.Now().UnixMilli(),
	}
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

// SetStayAlarmEnabled 注入本房间是否有任一设备启用 Stay alarm。
// 由 engine.RegisterRoom 启动时查 alarm_device 后调用。
// true 时 still fall 视该房间为 bathroom-like（与 cell ∪ room.name 取并集）。
func (tm *TrackManager) SetStayAlarmEnabled(enabled bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.stayAlarmEnabled = enabled
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

func (tm *TrackManager) payloadFromTrack(ts *TrackState) AIPayload {
	pxF, pyF := ts.Kalman.Position()
	conf := 100 - ts.GhostPenalty
	if conf < 0 {
		conf = 0
	}
	if conf > 100 {
		conf = 100
	}
	return AIPayload{
		DeviceID: ts.DeviceID,
		RoomID:   ts.RoomID,
		Track: observation.Track{
			TrackID:         ts.TrackID,
			PositionX:       intPtr(int(pxF + 0.5)),
			PositionY:       intPtr(int(pyF + 0.5)),
			PositionZ:       intPtr(ts.LastZ),
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
	tm.emitAIEvent(p, CategoryTrackVerdict, nowMs)
}

// IsBathroomByRoomName 用 owl-common/roomutil.ClassifyRoomType 判定本房间是否 bathroom。
// 与 cell.Belief[0].Type ∈ {AreaToilet, AreaShower} 取并集驱动 still fall。
func (tm *TrackManager) IsBathroomByRoomName() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return roomutil.IsBathroom(tm.roomName)
}

// ProcessSleepadBedEvent 接收 sleepad InBed/LeftBed 事件。
//   - 维护 bedPersonCount + lastLeftBedAt（旧逻辑）
//   - 维护 BedSession（新版 silent fall 状态机）
//
// 由 Engine 路由 iot:event:stream 中 device_type=Sleepad 时调用。
func (tm *TrackManager) ProcessSleepadBedEvent(evt SleepadBedEvent) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if evt.IsInBed {
		tm.bedPersonCount[evt.DeviceUID]++
		// PR-11: 记录 sleepad InBed ts 用于 radar/sleepad 一致期判定
		if evt.TMs > tm.lastSleepadInBedMs {
			tm.lastSleepadInBedMs = evt.TMs
		}
		// BedSession：首次 InBed 启动会话；重复 InBed 仅更新 MaxPeople
		s := tm.bedSessions[evt.DeviceUID]
		if s == nil || s.InBedSinceMs == 0 {
			s = &BedSession{DeviceUID: evt.DeviceUID, InBedSinceMs: evt.TMs}
			tm.bedSessions[evt.DeviceUID] = s
		}
		// 任意 InBed → 清掉之前的等待状态（视为新一轮上床）
		s.LeftBedAtMs = 0
		s.SilentFallAlerted = false
		if cnt := tm.bedPersonCount[evt.DeviceUID]; cnt > s.MaxPeople {
			s.MaxPeople = cnt
		}
		// PR-14：入场门控——若 radar InBed 在 ±15s 内已到，标记双源一致确认
		if tm.lastRadarInBedMs > 0 && tm.lastRadarInBedMs > tm.lastLeftBedAt {
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

	// LeftBed
	c := tm.bedPersonCount[evt.DeviceUID] - 1
	if c < 0 {
		c = 0
	}
	tm.bedPersonCount[evt.DeviceUID] = c
	if evt.TMs > tm.lastLeftBedAt {
		tm.lastLeftBedAt = evt.TMs
	}

	// BedSession：人数归 0 时进入「等待矛盾」状态，要求满足 5min precondition
	if c > 0 {
		return
	}
	s := tm.bedSessions[evt.DeviceUID]
	if s == nil || s.InBedSinceMs == 0 {
		return // 没有有效 in-bed 历史，可能 LeftBed 来得太早或重复事件
	}
	if evt.TMs-s.InBedSinceMs < int64(FallRulesParam.Silent.MinInBedSec)*1000 {
		// 在床时间不足 5min，不进入等待；直接结束 session
		s.InBedSinceMs = 0
		s.HasHRRR = false
		s.MaxPeople = 0
		return
	}
	s.LeftBedAtMs = evt.TMs
	s.LeftBedHadHRRR = s.HasHRRR
	s.LeftBedMaxPeople = s.MaxPeople
}

// totalBedPeople 同房间所有 sleepad 床的总人数（多床房间累加）
func (tm *TrackManager) totalBedPeople() int {
	n := 0
	for _, c := range tm.bedPersonCount {
		n += c
	}
	return n
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
		// 状态转换 InBed=true → InBed=false 视为 LeftBed（与 ProcessSleepadBedEvent 等价）
		// 单独的 event 流不一定到（部分固件只发 monitor），所以这里也要打开 R4 窗口。
		if ok && cur.InBed && !obs.InBed && obs.TMs > tm.lastLeftBedAt {
			tm.lastLeftBedAt = obs.TMs
		}
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
func (tm *TrackManager) consistentBedInBed(nowMs int64) bool {
	if tm.lastSleepadInBedMs == 0 || tm.lastRadarInBedMs == 0 {
		return false
	}
	// 必须都晚于 lastLeftBedAt（防止旧的 InBed 与新的 LeftBed 后又一致）
	if tm.lastSleepadInBedMs <= tm.lastLeftBedAt || tm.lastRadarInBedMs <= tm.lastLeftBedAt {
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
//   原因：人坐床上 sleepad HR/RR 信号可能弱；只要 InBed 就是床压传感器侧的存在性证据。
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

// RecordRadarAlarm 落账 radar 来源的 alarm（当前阶段仅 Fall）+ 跑 verifier 评分。
// 调用方（engine.handleAlarmMessage）应当紧跟 tm.Tick(alarm.TMs) 触发段 4-6 立即跑一次。
//
// PR-5：增加 verifier。仅 log 评分结果（fake/suspect/real），不否决固件 alarm 流程
// （alarm_events 表已由 wisefido-data 落账）。下游 cardagg 可订阅 ai.log 决定降级 risk。
func (tm *TrackManager) RecordRadarAlarm(a RadarFallAlarm) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cp := a
	tm.recentRadarAlarms[a.TMs] = &cp
	tm.evictOldRadarAlarms(a.TMs)

	// 仅 status="start" 的 fall 评分（end 是固件解除报警，不需要 verify）
	if a.Status == "start" || a.Status == "" {
		result := tm.verifyRadarFall(a, a.TMs)
		tm.logFallVerify(a, result)
		switch result.Verdict {
		case "ghost":
			tm.fallVerifyGhostCount++
		case "suspect":
			tm.fallVerifySuspectCount++
		case "real":
			tm.fallVerifyRealCount++
		}
	}
}

// RecordRadarEvent 落账 radar 来源的事件（EnterRoom/ExitRoom/InBed/LeftBed）。
// 仅落账。InBed/LeftBed 也"复用 sleepad 通道"更新 lastLeftBedAt（R4 开窗 + 床压计数）：
//   - LeftBed → 更新 tm.lastLeftBedAt（任意来源都开 R4 窗口）
//
// 注：radar 的 InBed/LeftBed 不更新 bedPersonCount——bedPersonCount 是 sleepad 床压传感器累计的"在床人数"，
// radar event 是空间检测，两者语义不同；混用会导致 bed-fall 段 5 的"仅 1 人"判定错乱。
//
// PR-14: radar InBed 触发两个副作用：
//   1) 当前 track 位置 cell 锁为 AreaBed（"床区自学"——radar 有事件即给出空间证据）
//   2) 若同房 sleepad InBed 已在 ±15s 内到达，标记 BedSession.RadarInBedConfirmedMs（双源一致门控）
func (tm *TrackManager) RecordRadarEvent(e RadarTrackEvent) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cp := e
	tm.recentRadarEvents[e.TMs] = &cp
	tm.evictOldRadarEvents(e.TMs)
	if e.EventName == "LeftBed" && e.TMs > tm.lastLeftBedAt {
		tm.lastLeftBedAt = e.TMs
	}
	if e.EventName == "InBed" && e.TMs > tm.lastRadarInBedMs {
		tm.lastRadarInBedMs = e.TMs
		// PR-14 副作用 1：mark 当前 track 位置 cell 为 AreaBed
		tm.markRadarInBedCell(e)
		// PR-14 副作用 2：若 sleepad InBed 在 ±15s 内已到，标记双源确认
		if tm.lastSleepadInBedMs > 0 && tm.lastSleepadInBedMs > tm.lastLeftBedAt {
			delta := e.TMs - tm.lastSleepadInBedMs
			if delta < 0 {
				delta = -delta
			}
			if delta <= bedInBedConsistencyMs {
				for _, s := range tm.bedSessions {
					if s.InBedSinceMs > 0 && s.RadarInBedConfirmedMs == 0 {
						s.RadarInBedConfirmedMs = e.TMs
					}
				}
			}
		}
	}
	// ExitRoom 事件 → 取消所有挂起的 lost-fall（人正常走出房间，不再悬念）
	// 注：silent fall 不取消（其语义是床上方遮挡，与 ExitRoom 无关）
	if e.EventName == "ExitRoom" && len(tm.pendingLostFalls) > 0 {
		for pid, p := range tm.pendingLostFalls {
			tm.lostFallPendingCancelled++
			tm.logger.Info("lost_fall_cancelled_by_exit_room",
				zap.String("device_uid", p.DeviceID),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("pending_age_ms", e.TMs-p.DisappearMs),
				zap.Int64("exit_room_ms", e.TMs),
			)
			delete(tm.pendingLostFalls, pid)
		}
	}
	// NumberPeople=0 → ExitRoom 兜底：① 记录时间戳供入池前查 ② 取消已存在的 non-frozen pending。
	// 实测 D523 firmware 在 FOV 边角离场不发 ExitRoom，但 number_people=0 始终发，
	// 且早 track_id=88 心跳 36-44ms，是更可靠的"屋空"信号。
	//
	// frozen 互斥：firmware 进入 frozen 残影状态时认为屋内有人（持续发同帧），
	// 绝不会发 number_people=0。若 pending.FrozenStartMs > 0（track 失锁前已 frozen），
	// 即使后来收到 number_people=0 也不取消——这是盲区返回 case（CD2B 类型），
	// 应保留 pending 等 birth-recovery 取消或正常超时报警。
	if e.EventName == "NumberPeople" && e.NumberPeople == 0 {
		if e.TMs > tm.lastNumberPeopleZeroMs {
			tm.lastNumberPeopleZeroMs = e.TMs
		}
		for pid, p := range tm.pendingLostFalls {
			if p.FrozenStartMs > 0 {
				tm.logger.Info("lost_fall_pending_kept_frozen_vs_number_people_zero",
					zap.String("device_uid", p.DeviceID),
					zap.Int("track_id", p.OriginalTrackID),
					zap.Int64("frozen_start_ms", p.FrozenStartMs),
					zap.Int64("number_people_zero_ms", e.TMs),
				)
				continue
			}
			tm.lostFallPendingCancelled++
			tm.logger.Info("lost_fall_cancelled_by_number_people_zero",
				zap.String("device_uid", p.DeviceID),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("pending_age_ms", e.TMs-p.DisappearMs),
				zap.Int64("number_people_zero_ms", e.TMs),
			)
			delete(tm.pendingLostFalls, pid)
		}
	}
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

// SetSleepadInBedCount 外部更新 Sleepad 在床人数
func (tm *TrackManager) SetSleepadInBedCount(count int) {
	tm.mu.Lock()
	tm.sleepadInBedCount = count
	tm.mu.Unlock()
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

	activeIDs := make(map[int]bool)

	// 入口先盘点 bathroom 内 real track 数（caregiver 例外用）
	tm.bathroomRealCount = 0
	for _, t := range tm.tracks {
		if t.Verdict != VerdictReal {
			continue
		}
		pxF, pyF := t.Kalman.Position()
		c := tm.grid.CellAt(int(math.Round(pxF)), int(math.Round(pyF)))
		if c == nil {
			continue
		}
		if c.Belief[0].Type == AreaToilet || c.Belief[0].Type == AreaShower {
			tm.bathroomRealCount++
		}
	}

	// ========== 段 1: 观测到的 track ==========
	for _, f := range frames {
		activeIDs[f.TrackID] = true
		ts, exists := tm.tracks[f.TrackID]

		var quality, vx, vy, dtSec int

		if !exists {
			// 新 track 出生 — 是否有 pending lost-fall 在等候 → 人从盲区返回；取消 + 学习盲区出口
			recoveredFromLost := tm.cancelPendingLostFallByBirth(f.X, f.Y, f.TMs)

			ts = NewTrackState(f.TrackID, f.DeviceID, tm.roomID, f.X, f.Y, f.Z, f.TMs)

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

			// 连续指标（frozen + Kalman birth-coherence），在 Kalman update 之后维护
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

		if ts.Kalman.MissCount > MaxMissCount {
			// 消失判定：lost fall（按 cell areaType 分时长，verdict 未定也算）
			// PR-14：旧 Path 1（track 消失 + 60s 复现窗口）已删除——
			// silent fall 仅由 BedSession LeftBed 矛盾路径触发（见 scanSilentFallLeftBed）。
			//
			// dedup：若该 track 已报过 bedside_fall（R4 床边晕倒，track 仍活时报），
			// 跳过 lost_fall pending 入池——避免"15min static→bedside_fall fire→track
			// 失锁→lost_fall 又 fire"的同事件双报。
			if ts.BedsideFallReported {
				tm.logger.Info("lost_fall_pending_skipped_after_bedside_fall",
					zap.String("device_uid", ts.DeviceID),
					zap.Int("track_id", ts.TrackID),
					zap.Int64("ts_ms", nowMs),
				)
			} else if (ts.Verdict == VerdictReal || ts.Verdict == VerdictPending) && tm.checkLostFall(ts) {
				// NumberPeople=0 ExitRoom 兜底：firmware 不发 ExitRoom 但发 number_people=0
				// （实测 D523 三个样本 100% 命中且早 track_id=88 36-44ms）→ 视作正常离场。
				//
				// 三道关卡：
				//   ① ts.FrozenRunStart == 0：track 失锁前未进入 frozen 残影状态。
				//      frozen ↔ number_people=0 互斥（firmware 状态机产物）：frozen 期间
				//      firmware 维持"屋内有人"判定，不发 number_people=0；若 track 经历过
				//      frozen 期才失锁，说明这是盲区残影 case（CD2B 类），应进 pending 池
				//      等 birth-recovery 取消或正常超时报警，不能被 number_people=0 抑制。
				//   ② lastNumberPeopleZeroMs > 0 && ts.LastObservedMs > 0：两个时间戳都有效。
				//   ③ |lastNumberPeopleZeroMs − LastObservedMs| ≤ NumberPeopleZeroFallbackMs(60s)：
				//      number_people=0 与 track 最后真实帧时间差在窗口内。
				//      比对基准是 LastObservedMs 不是 nowMs：后者因 MaxMissCount=10 会比
				//      number_people=0 晚约 10s，落不进窗口；前者紧贴 number_people=0
				//      （实测 ~1s 内）。
				if ts.FrozenRunStart == 0 && tm.lastNumberPeopleZeroMs > 0 && ts.LastObservedMs > 0 {
					gapMs := tm.lastNumberPeopleZeroMs - ts.LastObservedMs
					if gapMs < 0 {
						gapMs = -gapMs
					}
					if gapMs <= FallRulesParam.Lost.NumberPeopleZeroFallbackMs {
						tm.logger.Info("lost_fall_pending_skipped_by_number_people_zero",
							zap.String("device_uid", ts.DeviceID),
							zap.Int("track_id", id),
							zap.Int64("number_people_zero_ms", tm.lastNumberPeopleZeroMs),
							zap.Int64("last_observed_ms", ts.LastObservedMs),
							zap.Int64("gap_ms", gapMs),
							zap.Int64("ts_ms", nowMs),
						)
						tm.lostFallPendingCancelled++
						delete(tm.tracks, id)
						continue
					}
				}
				pxF, pyF := ts.Kalman.Position()
				px := int(math.Round(pxF))
				py := int(math.Round(pyF))
				cell := tm.grid.CellAt(px, py)
				cellArea := AreaUnknown
				if cell != nil {
					cellArea = cell.Belief[0].Type
				}
				tm.pendingLostFalls[id] = &PendingLostFall{
					OriginalTrackID: id,
					DeviceID:        ts.DeviceID,
					RoomID:          ts.RoomID,
					LastX:           px,
					LastY:           py,
					LastZ:           ts.LastZ,
					LastScore:       ts.Score,
					LastVerdict:     ts.Verdict,
					LastCellArea:    cellArea,
					DisappearMs:     nowMs,
					FrozenStartMs:   ts.FrozenRunStart,
					SpatialJump:     ts.MaxImpliedSpeedFromBirth > FallRulesParam.Lost.SuspectSpeedCm,
				}
				tm.lostFallPendingCreated++
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
			// 即使已是 Real verdict，penalty ≥ 80 也翻 Ghost（除非 LongSurvival 锚定）
			if ts.GhostPenalty >= GhostPenaltyThreshold && !ts.LongSurvivalAnchored && !ts.StartupGrace {
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
		if ts.GhostPenalty >= GhostPenaltyThreshold {
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
					zap.String("device_uid", ts.DeviceID),
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
						zap.String("device_uid", ts.DeviceID),
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

	// ========== 段 4b: 扫挂起的 lost fall，按 cell-area-typed wait + frozen credit 超时即报 ==========
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
				zap.String("device_uid", p.DeviceID),
				zap.Int("track_id", p.OriginalTrackID),
				zap.String("room_id", p.RoomID),
				zap.Int64("nowMs", nowMs),
			)
			delete(tm.pendingLostFalls, pid)
			continue
		}
		// 超时 → MarkFallEvent + 写输出（kind=engine_lost_fall）
		tm.lostFallReported++
		tm.grid.MarkFallEvent(p.LastX, p.LastY, nowMs)
		// PR5c: 发布到 iot:alarm:stream，category=alarm.Fall（cardagg 现有 Fall handler 接管）。
		// fall alarm 已是确认态——不再发 track_confidence/score（确信值无需信号），
		// 子类型通过 Reason 区分（"lost_track"），fall 严重度进 Evidence.fall_score。
		tm.emitAIAlarm(AIPayload{
			DeviceID: p.DeviceID,
			RoomID:   p.RoomID,
			Track: observation.Track{
				TrackID:   p.OriginalTrackID,
				PositionX: intPtr(p.LastX),
				PositionY: intPtr(p.LastY),
				PositionZ: intPtr(p.LastZ),
			},
			Reason: ReasonLostTrack,
			Evidence: map[string]interface{}{
				"context":         "track_lost_no_exit_room_no_recovery",
				"fall_score":      p.LastScore,
				"frozen_start_ms": p.FrozenStartMs,
				"spatial_jump":    p.SpatialJump,
				"cell_area_type":  int(p.LastCellArea),
				"wait_ms":         waitMs,
				"last_verdict":    int(p.LastVerdict),
			},
		}, alarm.Fall, nowMs)
		tm.logger.Info("real_fall",
			zap.String("device_uid", p.DeviceID),
			zap.Int("track_id", p.OriginalTrackID),
			zap.String("kind", "engine_lost_fall"),
			zap.Int("score", p.LastScore),
			zap.Int("risk", 100),
			zap.String("reason", "track_lost_no_exit_room_no_recovery"),
			zap.Int("cell_area_type", int(p.LastCellArea)),
			zap.Int64("frozen_start_ms", p.FrozenStartMs),
			zap.Bool("spatial_jump", p.SpatialJump),
			zap.Int64("wait_ms", waitMs),
			zap.Int("x", p.LastX), zap.Int("y", p.LastY), zap.Int("z", p.LastZ),
			zap.Int64("ts_ms", nowMs),
		)
		out := TrackOutput{
			TrackID:  p.OriginalTrackID,
			DeviceID: p.DeviceID,
			RoomID:   p.RoomID,
			Verdict:  p.LastVerdict,
			Score:    p.LastScore,
			Risk:     100,
			Anomaly:  AnomalyFall,
			X:        p.LastX,
			Y:        p.LastY,
			Z:        p.LastZ,
			Source:   "engine_lost",
		}
		results = append(results, out)
		tm.outputs[p.OriginalTrackID] = &out
		delete(tm.pendingLostFalls, pid)
	}

	// ========== 段 4c: 新版 Silent Fall（sleepad LeftBed + radar 仍在 Bed 邻域） ==========
	// 触发：bedSession.LeftBedAtMs > 0，等待 60s（vital + 单人）/ 120s（其它），
	//       超时仍有任一活 track 在 AreaBed ±BedNeighborhood cm 内 → 矛盾报警。
	// 否则取消（人正常起床，radar 也离床区）。
	results = append(results, tm.scanSilentFallLeftBed(nowMs)...)

	// ========== 段 5: Bed-Fall 物理矛盾检测 ==========
	// 物理意义：人从床上跌落到地面/床边，但雷达因坐标精度仍认为人在床
	// 触发条件：
	//   1. 雷达仍 track 着，坐标在 AreaBed cell
	//   2. sleepad 30s 内有数据且全部显示离床
	//   3. 房间总人数 == 1（避免家属/护工陪同时误报）
	// 房间总人数 = max(realTrackCount, totalBedPeople)
	//   - radar 多 track 即多人（雷达本来就追多人）
	//   - sleepad 床上有 ≥2 人即多人（连续 InBed 事件累计）
	if tm.sleepadOffBed(nowMs) {
		realCount := 0
		var soleReal *TrackState
		for _, ts := range tm.tracks {
			if ts.Verdict == VerdictReal {
				realCount++
				soleReal = ts
			}
		}
		bedPeople := tm.totalBedPeople()
		totalPeople := realCount
		if bedPeople > totalPeople {
			totalPeople = bedPeople
		}
		if totalPeople == 1 && realCount == 1 && soleReal != nil {
			pxF, pyF := soleReal.Kalman.Position()
			px, py := int(math.Round(pxF)), int(math.Round(pyF))
			cell := tm.grid.CellAt(px, py)
			if cell != nil && cell.Belief[0].Type == AreaBed {
				// 矛盾确认：雷达说在床 + sleepad 说离床 + 房间仅 1 人 → bed-fall
				if soleReal.CurrentAnomaly != AnomalyBedFall {
					soleReal.CurrentAnomaly = AnomalyBedFall
					tm.grid.MarkFallEvent(px, py, nowMs)
					tm.logger.Info("real_fall",
						zap.String("device_uid", soleReal.DeviceID),
						zap.Int("track_id", soleReal.TrackID),
						zap.String("kind", "engine_bed_fall"),
						zap.Int("score", soleReal.Score),
						zap.Int("risk", 100),
						zap.String("reason", "radar_in_bed_cell_but_sleepad_off_bed_solo"),
						zap.Int("x", px), zap.Int("y", py),
						zap.Int64("ts_ms", nowMs),
					)
				}
			}
		}
	}

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
			TrackID:  ts.TrackID,
			DeviceID: ts.DeviceID,
			RoomID:   ts.RoomID,
			Verdict:  ts.Verdict,
			Score:    ts.Score,
			Risk:     tm.computeRisk(ts, stillSec, nowMs),
			Anomaly:  ts.CurrentAnomaly,
			X:        px,
			Y:        py,
			Z:        ts.LastZ,
			VX:       vx,
			VY:       vy,
			StillSec: stillSec,
			Source:   source,
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
			zap.String("device_uid", p.DeviceID),
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
		// PR-14 入场门控：未双源（sleepad+radar InBed ±15s）确认 → 不报
		if s.RadarInBedConfirmedMs == 0 {
			continue
		}
		waitSec := param.WaitNoVitalSec
		if s.LeftBedHadHRRR && s.LeftBedMaxPeople == 1 {
			waitSec = param.WaitVitalSec
		}
		if nowMs-s.LeftBedAtMs < int64(waitSec)*1000 {
			continue
		}
		// 等待窗满 — 检查 radar 是否仍在 Bed 邻域
		if tm.anyActiveTrackNearBed(param.BedNeighborhood) {
			// 矛盾 → 报 silent fall
			s.SilentFallAlerted = true
			tm.silentFallLeftbedReported++
			// 选用「最近 Bed 的 active track」位置作为告警坐标
			x, y, z, scoreVal, verdict, deviceID := tm.pickActiveTrackNearBed(param.BedNeighborhood)
			tm.grid.MarkFallEvent(x, y, nowMs)
			// PR5c: alarm.Fall + Reason 区分子类型；BedStatus=1 反映 sleepad LeftBed 触发。
			// fall 已确认，不发 track_confidence；fall 严重度进 Evidence.fall_score。
			tm.emitAIAlarm(AIPayload{
				DeviceID: deviceID,
				RoomID:   tm.roomID,
				Track: observation.Track{
					PositionX: intPtr(x),
					PositionY: intPtr(y),
					PositionZ: intPtr(z),
					BedStatus: 1, // sleepad 报 LeftBed
				},
				Reason: ReasonSleepadRadarConflict,
				Evidence: map[string]interface{}{
					"context":      "sleepad_leftbed_radar_still_on_bed",
					"fall_score":   scoreVal,
					"radar_verdict": int(verdict),
					"sleepad_uid":  s.DeviceUID,
					"had_hr_rr":    s.LeftBedHadHRRR,
					"max_people":   s.LeftBedMaxPeople,
					"wait_sec":     waitSec,
					"leftbed_ms":   s.LeftBedAtMs,
				},
			}, alarm.Fall, nowMs)
			tm.logger.Info("real_fall",
				zap.String("device_uid", deviceID),
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
				zap.Int64("ts_ms", nowMs),
			)
			out = append(out, TrackOutput{
				DeviceID: deviceID,
				RoomID:   tm.roomID,
				Verdict:  verdict,
				Score:    scoreVal,
				Risk:     100,
				Anomaly:  AnomalyFall,
				X:        x,
				Y:        y,
				Z:        z,
				Source:   "engine_silent_leftbed",
			})
		} else {
			// 取消：radar 也离开了 Bed 邻域 → 人正常起床
			tm.silentFallLeftbedCancelled++
			tm.logger.Info("silent_fall_leftbed_cancelled",
				zap.String("sleepad_uid", s.DeviceUID),
				zap.Int64("leftbed_ms", s.LeftBedAtMs),
				zap.Int64("nowMs", nowMs),
				zap.Bool("had_hr_rr", s.LeftBedHadHRRR),
				zap.Int("max_people", s.LeftBedMaxPeople),
			)
			s.SilentFallAlerted = true // 防重复
			s.InBedSinceMs = 0         // session 结束，等待下次 InBed
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
func (tm *TrackManager) pickActiveTrackNearBed(marginCm int) (x, y, z, score int, verdict TrackVerdict, deviceID string) {
	for _, t := range tm.tracks {
		pxF, pyF := t.Kalman.Position()
		px, py := int(math.Round(pxF)), int(math.Round(pyF))
		if tm.grid.IsNearPriorType(px, py, AreaBed, marginCm) {
			return px, py, t.LastZ, t.Score, t.Verdict, t.DeviceID
		}
	}
	return 0, 0, 0, 0, VerdictReal, ""
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

// StillFallStats Still Fall（bathroom + Stand 静止 ≥ 15/18min）统计
type StillFallStats struct {
	Reported int
}

func (tm *TrackManager) StillFallStatsSnapshot() StillFallStats {
	return StillFallStats{Reported: tm.stillFallReportCount}
}

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
	enterPairWindowMs    = 3_000 // EnterRoom 与 birth 的最大时间差（ms）
	birthMaxRealisticCm  = 150   // 距 Enter 此值内才可能是 1 秒走入的真人
	birthEnterPairBonus  = 20    // 有 EnterRoom 配对加分
	GhostPenaltyThreshold = 80   // 累积 ghost penalty 阈值
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
				zap.String("device_uid", ts.DeviceID),
				zap.Int("track_id", ts.TrackID),
				zap.Int("ghost_penalty", ts.GhostPenalty),
				zap.Int64("ts_ms", nowMs),
			)
		} else if partner, mx, my, rx, ry, ok := tm.checkMirrorSymmetry(ts); ok {
			ts.GhostPenalty += 10
			ts.BirthReason = "mirror_image_of_real_track"
			bxF, byF := ts.Kalman.Position()
			tm.logger.Info("ghost_mirror_symmetry_hit",
				zap.String("device_uid", ts.DeviceID),
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
		coexistDistMaxCm = 100      // 双 track 距离上限
		windowMs         = int64(2_000) // 2s 滚动窗
		minDispCm        = 10       // 单 track 最小位移（< 10 视为静止 noise）
		cosThreshold     = 0.866    // cos(30°)
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
//   A 路径（z 突变 = 坐→站）：region static ≥2min AND |z - RegionStartZ| ≥10cm AND 双方 z>0
//   B 路径（持续累积）：region static ≥ threshold （RestZone cell 8min；其它 12min）AND ratio ≥0.90
// 触发后双 cell 加分：prev cell + cur cell 同时 MarkRestZoneFeedback(AreaSit)。
// per-track 一次性（AreaSitLearnedRegion flag）。
//
// 调用方持锁；nowMs/cell 已由 scoreMovement 计算。
func (tm *TrackManager) updateRegionStatic(ts *TrackState, prev TimedPoint, x, y int, nowMs int64, cell *Cell) {
	const (
		regionDxDyMaxCm = 15  // |dx|≤15 AND |dy|≤15 = 同区域
		regionResetCm   = 50  // |dx|>50 或 |dy|>50 = 大跨步，立即 reset
		ratioMin        = 0.90
		ratioResetMin   = 0.85 // ratio 跌破此值 → region 失效 reset
		zJumpMinCm      = 10
		zJumpMinElapse  = 2 * 60 * 1000      // 2 min
		thresholdRest   = 8 * 60 * 1000      // RestZone cell: 8min
		thresholdNonRest = 12 * 60 * 1000    // 其它: 12min
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
		zap.String("device_uid", ts.DeviceID),
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
		if e.EventName != "EnterRoom" {
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
		if e.EventName == "EnterRoom" {
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
			zap.String("device_uid", ts.DeviceID),
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
					zap.String("device_uid", ts.DeviceID),
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
					zap.String("device_uid", ts.DeviceID),
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
		if cell != nil {
			isRiskTime := IsNightTime(nowMs, tm.timezone)
			timeout := cell.EffectiveStillTimeoutSec(isRiskTime)
			if timeout > 0 {
				// Bathroom caregiver 例外：本 cell 在 toilet/shower + ≥2 real track 同在 bathroom
				// → 第二个 track 大概率是护工陪同，长时间静止合理（如老人坐马桶、护工旁边照看）
				inBathroom := cell.Belief[0].Type == AreaToilet || cell.Belief[0].Type == AreaShower
				skipTimeout := inBathroom && tm.bathroomRealCount >= 2
				if !skipTimeout {
					stillSec := int((nowMs - ts.StillSince) / 1000)
					if stillSec > timeout {
						ts.CurrentAnomaly = AnomalyStillTooLong
						if !ts.LongStillReported {
							tm.grid.MarkLongStill(x, y, nowMs)
							ts.LongStillReported = true
						}
					}
				}
			}

			// 升级 → Still Fall（bathroom + pose=Stand + 15/18min 持续静止）。
			// 触发位置语义：cell.Belief[0].Type ∈ {AreaToilet, AreaShower} 或 room.name 含 bathroom/restroom/toilet（取并集）。
			// 阈值：cell 是 toilet/shower 时用 cell.EffectiveStillTimeoutSec（含 cell history 自适应）；
			//       cell 未学到但 room name 是 bathroom 时用 FallRulesParam.Still.ToiletShowerSec（无 cell 容忍）。
			// pose=Stand：避免坐马桶（pose=Sit）误报。caregiver 例外同上。
			if !ts.StillFallReported && RadarPoseToCore(pose) == CorePoseStand {
				stillFallTimeoutSec := tm.stillFallTimeoutSec(cell, isRiskTime)
				if stillFallTimeoutSec > 0 && tm.bathroomRealCount < 2 {
					stillSec := int((nowMs - ts.StillSince) / 1000)
					if stillSec > stillFallTimeoutSec {
						ts.CurrentAnomaly = AnomalyFall
						ts.StillFallReported = true
						tm.stillFallReportCount++
						tm.grid.MarkFallEvent(x, y, nowMs)
						// PR5c: alarm.Fall + Reason=still_in_bathroom；时长进 Evidence。
						// fall 已确认，清零 TrackConfidence（payloadFromTrack 默认带 conf 值）。
						stillP := tm.payloadFromTrack(ts)
						stillP.Track.TrackConfidence = 0
						stillP.Reason = ReasonStillInBathroom
						stillP.Evidence = map[string]interface{}{
							"context":           "still_in_bathroom_over_threshold",
							"still_seconds":     stillSec,
							"still_timeout_sec": stillFallTimeoutSec,
							"cell_area_type":    int(cell.Belief[0].Type),
						}
						tm.emitAIAlarm(stillP, alarm.Fall, nowMs)
						tm.logger.Info("real_fall",
							zap.String("device_uid", ts.DeviceID),
							zap.Int("track_id", ts.TrackID),
							zap.String("kind", "engine_still_fall"),
							zap.Int("score", ts.Score),
							zap.Int("risk", 100),
							zap.String("reason", "bathroom_stand_static_too_long"),
							zap.Int("still_sec", stillSec),
							zap.Int("threshold_sec", stillFallTimeoutSec),
							zap.Bool("is_risk_time", isRiskTime),
							zap.Int("cell_area", int(cell.Belief[0].Type)),
							zap.String("room_name", tm.roomName),
							zap.Int("x", x), zap.Int("y", y),
							zap.Int64("ts_ms", nowMs),
						)
					}
				}
			}

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
							zap.String("device_uid", ts.DeviceID),
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

		// R4: 床边晕倒（升级 AnomalyStillTooLong → AnomalyBedsideFall）
		// 触发：风险时段 + 最近 WindowSec 内有 LeftBed + 当前位置距 AreaBed ≤ BedsideMarginCm
		//       + 静止 > StillTimeoutSec
		// 物理意义：人离床去卫生间途中在床边滑倒/晕倒，雷达失锁前最后位置仍在床边。
		// 与 R2(BedFall) 区别：R2 是"仍在床矛盾"，R4 是"离床后到不了远方"。
		if tm.lastLeftBedAt > 0 &&
			nowMs-tm.lastLeftBedAt < int64(tm.bedsideFallCfg.WindowSec)*1000 &&
			IsNightTime(nowMs, tm.timezone) &&
			tm.grid.IsNearPriorType(x, y, AreaBed, tm.bedsideFallCfg.BedsideMarginCm) {
			stillSec := int((nowMs - ts.StillSince) / 1000)
			if stillSec > tm.bedsideFallCfg.StillTimeoutSec && ts.CurrentAnomaly != AnomalyBedsideFall {
				ts.CurrentAnomaly = AnomalyBedsideFall
				tm.grid.MarkFallEvent(x, y, nowMs)
				ts.LongStillReported = true   // 复用 flag 防 LongStill 重复 mark
				ts.BedsideFallReported = true // 防双报：后续若 track 失锁，跳过 lost_fall pending 入池
				// PR5c: alarm.Fall + Reason=bedside_silent（R4 床边晕倒）。
				// fall 已确认，清零 TrackConfidence（payloadFromTrack 默认带 conf 值）。
				bedsideP := tm.payloadFromTrack(ts)
				bedsideP.Track.TrackConfidence = 0
				bedsideP.Reason = ReasonBedsideSilent
				bedsideP.Evidence = map[string]interface{}{
					"context":            "bedside_still_after_leftbed",
					"still_seconds":      stillSec,
					"window_sec":         tm.bedsideFallCfg.WindowSec,
					"still_timeout_sec":  tm.bedsideFallCfg.StillTimeoutSec,
					"bedside_margin_cm":  tm.bedsideFallCfg.BedsideMarginCm,
					"leftbed_at_ms":      tm.lastLeftBedAt,
				}
				tm.emitAIAlarm(bedsideP, alarm.Fall, nowMs)
				tm.logger.Info("real_fall",
					zap.String("device_uid", ts.DeviceID),
					zap.Int("track_id", ts.TrackID),
					zap.String("kind", "engine_bedside_fall_R4"),
					zap.Int("score", ts.Score),
					zap.Int("risk", 100),
					zap.String("reason", "night_left_bed_then_bedside_still_15min"),
					zap.Int("x", x), zap.Int("y", y),
					zap.Int("still_sec", stillSec),
					zap.Int64("ts_ms", nowMs),
				)
			}
		}
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
				zap.String("device_uid", ts.DeviceID),
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
// Frozen credit（box 判据）：失锁前 30s 内位移 box <= StillBoxCm 视为 still，
//   FrozenStartMs 是该 still box run 起点，stillDur = DisappearMs - FrozenStartMs，
//   credit = stillDur / 2 半计入等待。
//
// SpatialJump factor：track 表现过空间跳跃 → 等待时间 ×0.5（更敏感）
// 兜底：min EffectiveWaitFloorSec
func (tm *TrackManager) lostFallWaitMs(p *PendingLostFall, isRiskTime bool) int64 {
	pl := FallRulesParam.Lost
	var base int
	switch p.LastCellArea {
	case AreaBed, AreaSit:
		base = pl.RestZoneWaitSec
	case AreaToilet, AreaShower:
		// 与 still fall 同时长
		if isRiskTime {
			base = FallRulesParam.Still.ToiletShowerSec
		} else {
			base = int(float64(FallRulesParam.Still.ToiletShowerSec) * FallRulesParam.Still.NonRiskTimeFactor)
		}
	case AreaDeny:
		base = pl.DenyZoneWaitSec
	default:
		base = pl.WalkwayWaitSec
	}

	// Frozen credit：half of still box duration counted toward wait
	if p.FrozenStartMs > 0 {
		frozenDurMs := p.DisappearMs - p.FrozenStartMs
		if frozenDurMs > 0 {
			base -= int(frozenDurMs / 2 / 1000)
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

// stillFallTimeoutSec 当前 cell + room 上下文下 still-fall 的有效阈值（秒）。
//
//	cell.AreaToilet/AreaShower → cell.EffectiveStillTimeoutSec（含 cell history 自适应）
//	否则 (room.name 是 bathroom) ∪ (Stay alarm 启用) → FallRulesParam.Still.ToiletShowerSec
//	                                                  × NonRiskTimeFactor (non-risk-time 时)
//	都不匹配 → 0（不报 still fall）
//
// 三路并集：cell 学到 / room 命名 / 运维显式启用 Stay alarm。
// 任一命中即按 bathroom 处理；Stay alarm 启用语义="操作员明确需要长时间静止检测"。
//
// 调用方持锁。
func (tm *TrackManager) stillFallTimeoutSec(cell *Cell, isRiskTime bool) int {
	if cell == nil {
		return 0
	}
	cellType := cell.Belief[0].Type
	if cellType == AreaToilet || cellType == AreaShower {
		// cell 已学到 toilet/shower：完全用 cell.EffectiveStillTimeoutSec（已含 risk-time 与 tolerance）
		return cell.EffectiveStillTimeoutSec(isRiskTime)
	}
	// cell 未学到但 (room.name 是 bathroom) 或 (Stay alarm 启用) → 用基线时长（无 cell history）
	if roomutil.IsBathroom(tm.roomName) || tm.stayAlarmEnabled {
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

// updateContinuousIndicators 每帧维护 frozen-frame 检测 + Kalman birth-coherence 指标。
//
// 1. Frozen 检测（box 判据，2026-05-03 由 byte-equal 改为 box）：
//    最近 30s History 滚动窗口内位移 box (max-min) ≤ StillBoxCm(30) → still。
//    - 进入 still：起点 = History 最早帧 TMs（自然回填到 box 内最早一帧）
//    - 持续 still：FrozenRunStart 不变（即使 History 滚动丢早期帧）
//    - 跳出 box：FrozenRunStart 清零
//    用于 lost-fall pending credit（半计入 wait）+ PR-C 流式 cancel 守卫。
// 2. MaxKalmanResidual：track 生命周期峰值残差。
// 3. MaxImpliedSpeedFromBirth：max(dist(current, birth) / age) cm/s；
//    > ImpossibleSpeedCm 判硬 ghost；> SuspectSpeedCm + 无 EnterRoom 判软 ghost。
//
// 调用位置：processFrameAt 已有 track 分支，Kalman.Update 之后。
func (tm *TrackManager) updateContinuousIndicators(ts *TrackState, f TrackFrame, nowMs int64, residualF float64) {
	// ---- Frozen 检测（box 判据）----
	disp := ts.DisplacementWithinMs(30_000, nowMs)
	if disp <= FallRulesParam.Lost.StillBoxCm && len(ts.History) >= 2 {
		if ts.FrozenRunStart == 0 {
			// 起点回填到 History 最早帧（box 内最早可见点）
			ts.FrozenRunStart = ts.History[0].TMs
		}
	} else {
		ts.FrozenRunStart = 0
	}
	// NOTE（防御层备忘，未实施）：box 判据只看 max-min 范围。理论 edge case：
	// box 内反复抖动（30cm 范围内来回跨越）→ box 小但累计位移大 → 误判 still。
	// 实测 D523/CD2B/D5F7 未观察到（firmware 单帧抖动通常 ±5-10cm）。生产若遇
	// 此类"假 still 导致 lost-fall 误判"再加：30s 内逐帧累计 > 200cm 也清零
	// FrozenRunStart（200cm = box 周长 120cm × 1.6 倍，超过即反复跨越）。

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
