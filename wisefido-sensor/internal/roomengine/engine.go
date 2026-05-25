package roomengine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	"owl-common/radarutils"
	rediscommon "owl-common/redis"
	"owl-common/spatial"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ========================================================================
// RoomConfig：从 layout_config JSONB 解析得到的房间配置
// ========================================================================

// RoomConfig 房间配置（全 int 化 + 对齐 radarutils 类型）
type RoomConfig struct {
	RoomID   string
	RoomName string // rooms.room_name；保留兼容；sensor_v2 决定 16 后只用 RoomType 做分支判定，不再用 RoomName 字符串匹配
	RoomType int    // owl-common/card.RoomType*: 0=Default / 1=Bathroom / 2=Kitchen
	RoomW    int    // 画布宽（cm）
	RoomH    int    // 画布深（cm）
	OriginX  int    // grid[0][0] 左上角的画布坐标 X（让 grid 覆盖物体 bbox）
	OriginY  int    // grid[0][0] 左上角的画布坐标 Y

	// Wall 围出的房间多边形（闭合），用于 StampRoomPolygon
	WallPolygon []radarutils.Point

	// 人工标注的矩形先验
	Enters     []radarutils.Rect // AreaEnter
	Beds       []radarutils.Rect // AreaBed
	Toilets    []radarutils.Rect // AreaToilet
	Showers    []radarutils.Rect // AreaShower
	Chairs     []radarutils.Rect // AreaSit（粗标沙发/椅子，Conf=80）
	Furnitures []radarutils.Rect // AreaDeny（家具/桌子）
	Interferes []radarutils.Rect // AreaDeny（镜子/金属反射区/吊灯）

	// 物体顶部高度（cm，地面为 0），与上面同名切片一一对应（Heights[i] ↔ Beds[i]）。
	// 来源：layout JSON 里每个对象的 height 字段，由前端 Toolbar 录入；
	// 缺失时 ParseLayoutConfig 用 FurnitureType 默认值（与前端 FURNITURE_CONFIGS.defaultHeight 对齐）。
	// 当前 RoomEngine 不使用，仅持久化保留——未来用于：
	//   - 空中物体（吊灯 height>200）不阻挡通行 → InFOV 不刷 Deny
	//   - 床/桌面高度参与 Z 轴范围判定 → fall 检测加先验
	EnterHeights     []int
	BedHeights       []int
	ToiletHeights    []int
	ShowerHeights    []int
	ChairHeights     []int
	FurnitureHeights []int
	InterfereHeights []int

	// sensor_v2 决定 15: EnterTargets 与 Enters 同长度平行 slice。
	// 取值：
	//   ""         → inside_enter (默认；不强制标识)
	//   "outside"  → 通向 unit 外（必须人工标）
	//   "bathroom" → 通向 bathroom（必须人工标；v2 单 bathroom 无需 id）
	// vue 编辑器 dropdown 三选项，写入 layout JSON 每个 enter 对象的 enter_target 字段。
	// 缺失或为 "" → 视作 inside_enter。
	EnterTargets []string

	// 雷达安装
	Radar radarutils.RadarMount

	// Sleepad 位置（point 几何，可能多个）— 用于事件路由 + 可视化
	Sleepads []radarutils.Point

	// Timezone IANA 时区字符串（如 "America/Denver"），来自该 room 所属 unit 的 units.timezone。
	// 用于 IsNightTime 判定 risk-time（默认 23:30-07:30 本地时间）。
	// 空值 → engine 退化为 UTC（错位风险，必须设置）。
	Timezone string

	// sensor_v2 PR-3 wiring：
	// SuiteID    = census bucket key；取值规则按 IsPublicBathroom 分两路（决定 24 / PR-7）：
	//                public bathroom (IsPublicBathroom=true)  → bathroom_spatial_prefix /128（自身）
	//                suite bathroom / bedroom / 其他           → unit_spatial_prefix /80
	//              bootstrap 责任：按下面规则计算 SuiteID 传入 RegisterRoom。
	// ResidentID = bedroom 关联的 resident.id；bathroom 不必填；PR-2 SuiteCensusManager 升格用。
	// 空值 = bootstrap 未配置 → TrackStatus.PersonID/PersonRole 一律空。
	//
	// **当前是单 resident 字段，multi-resident bedroom 暂未支持**：
	//   PR-3 范围内 caller 责任传入"该 room 的预期 resident"（单 resident unit 直接传 residents[0]）。
	//   决定 19 multi-resident 场景（dual sleepad bedroom）的正解 = sleepad device_uid → resident.id
	//   静态映射，未来扩展为 []ResidentBinding 或 ResidentLookup func；
	//   PR-5/6 BathroomGate + 实测 multi-resident fixture 时再决（详 sensor_v2.md 决定 19）。
	SuiteID    string
	ResidentID string

	// sensor_v2 决定 24 / PR-7 Public Bathroom Standalone Mode：
	// 仅当 RoomType==card.RoomTypeBathroom 时有意义：
	//   true  = public bathroom（楼层共用，多人不识别身份）→ RegisterRoom 触发 census.MarkPublicBathroom，
	//           BathroomGate.Process / BathroomGhostAdjudicator 规则 2/3 因 IsPublicBathroom 自动降级
	//   false = suite bathroom（附属某 bedroom，elder 私用）→ 走完整 §4.A 判定 + BathroomGate 入口流量
	// 来源：rooms.is_public_bathroom 列（migration 2026-05-17_rooms_add_is_public_bathroom.sql）
	// 显式配置不推断（详 sensor_v2.md §4.A.6 触发条件 + 拒绝拓扑推断的理由）
	IsPublicBathroom bool
}

// ========================================================================
// ParamSet[3] 与 winner 选择
// ========================================================================

// DefaultParamSets 三组并行参数（保守/中庸/激进）
var DefaultParamSets = [3]ParamSet{
	{Alpha: 0.01, Beta: 0.2, FlipTh: 10, Name: "conservative"},
	{Alpha: 0.02, Beta: 0.5, FlipTh: 20, Name: "balanced"},
	{Alpha: 0.05, Beta: 1.0, FlipTh: 30, Name: "aggressive"},
}

// AccuracyTracker 单组参数的准确率累计（供 winner 选择使用）
// TP / FP / TN / FN 由 feedback_loop.go 灌入
type AccuracyTracker struct {
	TruePositive  int
	FalsePositive int
	TrueNegative  int
	FalseNegative int
	LastEvalAt    time.Time
}

// Accuracy 准确率 [0,1]；样本不足返回 -1
func (a *AccuracyTracker) Accuracy() float64 {
	total := a.TruePositive + a.FalsePositive + a.TrueNegative + a.FalseNegative
	if total < 5 {
		return -1
	}
	return float64(a.TruePositive+a.TrueNegative) / float64(total)
}

// ========================================================================
// Engine
// ========================================================================

type Engine struct {
	mu         sync.RWMutex
	rooms      map[string]*TrackManager         // roomID → TrackManager
	grids      map[string]*RoomGrid             // roomID → Grid
	mounts     map[string]radarutils.RadarMount // roomID → Radar 安装参数（坐标转换用）
	deviceRoom map[string]string                // deviceAddr → roomID (sensor 唯一物理寻址)
	// PR-8: AI publish 用 — 源 radar UUID 反查 deviceUID + room 反查 tenant/card
	deviceIDToUID   map[string]string // deviceID(UUID) → device_uid（IoTStreamMessage.DeviceUID）
	deviceIDToType  map[string]string // deviceID(UUID) → 源 sensor 类型（"Radar"/"Sleepad"），AI publish 加 ".AI<node>" 后缀
	roomTenants     map[string]string // roomID → tenant_id（alarm_events 必填）

	// 北极星 reasoning trace 链：每 device 最近 inbound msg 的 envelope.SequenceNumber，
	// AI verdict publish 时 evidence["trigger_seq_num"] = lastSrcSeq[deviceAddr]。
	// 多 producer (qinglan/sleepace) 各自有独立 seq counter；本 map 按 deviceAddr 区分。
	srcSeqMu  sync.RWMutex
	lastSrcSeq map[string]uint64
	// 不再维护 roomCards：AI 是 sensor 层 producer，只发 device 标识；card_id (subject_entity)
	// 由 cardagg IotPreparedHandler 反查 device→cards 路由（多卡共享设备时自然 fan-out）。
	// 协议层北极星 layer 1 / 2 解耦原则。详 doc/TODO.md 第 0 项。

	// Sensor agent 节点完整身份字符串（如 "sensor.caregiver01"），同时作 envelope.Producer
	// + wire fields["source"] + ai_emit log 审计字段。多实例横向扩展时各进程 source 不同 →
	// wire 自动区分实例。来源见 cfg.AIPublish.Source。Phase B：命名规范 sensor.<role><实例号>。
	aiSource string

	// publish 模式："log" | "log&publish"。
	// "log" 模式仅写 sensor.log，不发 redis stream；任何模式都不影响 alarm 触发路径。
	aiPublishMode string

	// layout 几何 hash，RegisterRoom 时计算；snapshot save/load 比对用
	layoutHashes map[string]string

	// 自适应参数
	paramSets [3]ParamSet
	accuracy  [3]AccuracyTracker
	winner    int // 当前 winner 组（0/1/2），-1=无 winner 用 baseline

	// 运行时参数（由 Configure 注入；Decay/Learn 字段语义见 cell.go / cell_learning.go）
	decayParams     DecayParams
	learnParams     LearnParams
	bedsideFallCfg  BedsideFallConfig // R4 床边晕倒参数；RegisterRoom 时下发到 TrackManager

	// 定时器
	decayInterval      time.Duration // 默认 1 小时（Decay 计算一次）
	beliefScanInterval time.Duration // PR-11: 默认 10 分钟（原 5min；降低 CPU 功率）
	winnerEvalInterval time.Duration // 默认 24 小时（winner 重评）
	snapshotInterval   time.Duration // 默认 5 分钟（持久化全量 dump）；0 = 关闭

	// 持久化（nil = 不持久化）
	persister Persister

	// 历史归档（nil = 不归档）。每天 dailySnapshotHour:dailySnapshotMinute 触发，保留 historyRetainDays 天
	historyPersister    HistoryPersister
	dailySnapshotHour   int // 0-23 local；-1=禁用 daily snapshot
	dailySnapshotMinute int // 0-59 local
	historyRetainDays   int // <=0 表示不清理

	// alarm-feedback ingestion（cell.IncrFakeAlarm 反馈链）
	feedbackDB        *sql.DB
	feedbackInterval  time.Duration // 默认 5 分钟；0 = 关闭
	feedbackIngester  *AlarmFeedbackIngester

	// PR-15: 每日定时 layout reload。dailyReloadHour 0-23（local time）；-1=禁用
	// 目的：管理员下班后重读 rooms.layout_config，hash 变 → 重置该 room grid，从 0 重学
	dailyReloadHour int
	dailyReloadDB   *sql.DB // 用于 SELECT layout_config；nil 时跳过

	// 路由失败的 device 频率限制告警（避免每条 frame 都 warn 一次）。
	// 同一 device key 60s 内只 warn 一次，确保 deviceRoom 缺失能被发现而不淹日志。
	unroutedMu sync.Mutex
	unrouted   map[string]int64 // device key → last warn epoch ms

	redisClient *redis.Client
	logger      *zap.Logger

	// weakBioSource: fall verifier "WeakBio≥80 force real" 短路；engine 持有 + RegisterRoom 时
	// 转发给每个 TrackManager（默认 nil，verifier 走原三档评分）。详 fall_verify.go 注释。
	weakBioSource WeakBioSource

	onOutput func(roomID string, outputs []TrackOutput)

	// sensor v2 PR-3 wiring：
	//   - suiteCensus 进程级共享 census manager（nil = PR-3 未启用，TrackStatus.PersonID 空）
	//   - roomSuiteID    roomID → SuiteID（unit spatial_prefix INET 字符串）；bootstrap 注入
	//   - roomResidentID roomID → resident.id（PR-2 升格用；bathroom room 不必填）
	//   - roomType       roomID → card.RoomType* (0=Default / 1=Bathroom / 2=Kitchen)
	// 单 reader（handleMessage 调 SnapshotTrackStatuses 后 publish）；上述 map 在 RegisterRoom
	// + SetSuiteCensus 时写入，运行时只读，无并发风险。
	suiteCensus    *SuiteCensusManager
	roomSuiteID    map[string]string
	roomResidentID map[string]string
	roomType       map[string]int

	// trackLastSeen 是 publishTrackStatuses 的失锁判定状态：roomID → trackID → 上次出现的 nowMs。
	// 用途：firmware track_id 复用场景下，SuiteCensus.AnchorTrackID 必须在 track 真正失锁后清空，
	// 否则新 track 复用旧 trackID 会被当成"同一人活动"持续 update LastActiveMs。
	// 判据：当前 snapshot 未含 trackID 且 (nowMs - lastSeenMs) ≥ 60s → 调 ClearAnchorOnLostTrack。
	// 60s 偏保守（远大于 MaxMissCount=10 帧的 coast 期），避免 firmware coast 期间误清。
	// 写者：publishTrackStatuses（per-room 串行，与 handleMessage 同 goroutine）；无锁需要。
	trackLastSeen map[string]map[int]int64

	// sensor_v2 PR-4 ghost adjudicator 分支选择（决定 16 / §10.3.1）：
	//   bathroomGhostAdj → room_type == card.RoomTypeBathroom
	//   generalGhostAdj  → 默认（含 bedroom / living / kitchen / ...）
	// nil = 用 NoopGhostAdjudicator（行为 == v1，PR-4 默认）；bootstrap 调 SetGhostAdjudicators 注入真实实现。
	// 读取走 e.mu.RLock；写入仅在 SetGhostAdjudicators 一次（无 hot swap 需求）。
	bathroomGhostAdj GhostAdjudicator
	generalGhostAdj  GhostAdjudicator

	// sensor_v2 PR-5 BathroomGate 入口流量子模块（§4.A.2）：
	//   每 bathroom room 一个 gate（key = roomID），lazy 创建在 publishTrackStatuses。
	//   census + suiteID 为空时不创建（fallback noop）。
	//   单 goroutine 读写（publishTrackStatuses per-room 串行），无锁需要。
	bathroomGates map[string]*BathroomGate

	// sensor_v2 PR-10 BathroomFallRules（§6.A）：
	//   实例级单例（所有 bathroom rooms 共用一个 rules，内部 per-roomID 状态分桶）；
	//   bootstrap 调 SetBathroomFallRules 注入；nil = 跳过 bathroom fall 判定（PR-9 后 v1 路径已删，
	//   不注入意味着 dev 阶段 fall 检测降级）。
	bathroomFall *BathroomFallRules

	// sensor_v2 PR-11 BedroomFallRules（§6.B 11b + 11c）：
	//   仅 non-bathroom room（bedroom 默认 + 其他 room.kind）入；
	//   silent_fall (11a) 仍由 TrackManager.scanSilentFallLeftBed 处理（PR-9 后已对齐 §6.B.1）。
	bedroomFall *BedroomFallRules
}

// RuntimeConfig 与 wisefido-sensor/internal/config::RoomEngineConfig 一一对应；
// engine 包不依赖 config 包，wiring 在 cmd/wisefido-sensor/main.go 中完成转换。
type RuntimeConfig struct {
	Decay              DecayParams
	Learn              LearnParams
	ParamSets          [3]ParamSet
	DecayInterval      time.Duration
	BeliefScanInterval time.Duration
	WinnerEvalInterval time.Duration
	SnapshotInterval   time.Duration // 0 = 关闭持久化定时器；Persister 仍可在退出时 dump
	Persister          Persister     // nil = 不持久化

	// HistoryPersister + DailySnapshotHour/Minute + HistoryRetainDays：每日归档
	HistoryPersister    HistoryPersister
	DailySnapshotHour   int // -1=禁用；默认 11（11:50 local）
	DailySnapshotMinute int // 0-59；默认 50
	HistoryRetainDays   int // 默认 365；<=0 不清理

	// 风险时段（夜间）；通过 SetRiskTimeConfig 注入到包级 nightCfg。
	// 全 0 视为未设置，保留 IsNightTime 默认（23:30 - 07:30）。
	RiskTime RiskTimeConfig

	// R4 床边晕倒参数；通过 TrackManager.SetBedsideFallConfig 注入到每个房间。
	// 任一字段 0 保留 defaultBedsideFallCfg 默认值。
	BedsideFall BedsideFallConfig

	// FeedbackDB / FeedbackInterval：alarm_events false_alarm 反馈链。
	// 二者都设置才启用；缺任一退化为不启动 feedbackLoop。
	FeedbackDB       *sql.DB
	FeedbackInterval time.Duration // 默认 5 分钟
}

// NewEngine 创建 Room Engine（默认参数）。生产环境调用 Configure 注入 yaml 配置。
func NewEngine(redisClient *redis.Client, logger *zap.Logger) *Engine {
	return &Engine{
		rooms:              make(map[string]*TrackManager),
		grids:              make(map[string]*RoomGrid),
		mounts:             make(map[string]radarutils.RadarMount),
		deviceRoom:         make(map[string]string),
		deviceIDToUID:      make(map[string]string),
		deviceIDToType:     make(map[string]string),
		roomTenants:        make(map[string]string),
		lastSrcSeq:         make(map[string]uint64),
		aiSource:           "sensor.caregiver01",
		aiPublishMode:      "log&publish",
		layoutHashes:       make(map[string]string),
		paramSets:          DefaultParamSets,
		winner:             1, // 默认 balanced
		decayParams:        DefaultDecayParams(),
		learnParams:        DefaultLearnParams(),
		decayInterval:        1 * time.Hour,
		beliefScanInterval:   10 * time.Minute, // PR-11
		winnerEvalInterval:   24 * time.Hour,
		snapshotInterval:     5 * time.Minute,
		dailyReloadHour:      22, // PR-15：22:00 local 重读 layout
		dailySnapshotHour:    11, // 每天 11:50 local 归档 daily history
		dailySnapshotMinute:  50,
		historyRetainDays:    365, // 一年滚动清理
		unrouted:             make(map[string]int64),
		redisClient:          redisClient,
		logger:               logger,
		roomSuiteID:          make(map[string]string),
		roomResidentID:       make(map[string]string),
		roomType:             make(map[string]int),
		trackLastSeen:        make(map[string]map[int]int64),
		bathroomGates:        make(map[string]*BathroomGate),
	}
}

// warnUnrouted 路由失败的 device 频率限制告警（同一 key 60s 内一次）。
// stream：消息来自哪个 stream（"monitor"/"event"/"alarm"）便于排查。
func (e *Engine) warnUnrouted(stream, cardID, deviceAddr, deviceType string) {
	key := deviceAddr
	if key == "" {
		key = cardID
	}
	if key == "" {
		return
	}
	nowMs := time.Now().UnixMilli()
	e.unroutedMu.Lock()
	last := e.unrouted[key]
	if nowMs-last < 60_000 {
		e.unroutedMu.Unlock()
		return
	}
	e.unrouted[key] = nowMs
	e.unroutedMu.Unlock()
	e.logger.Warn("dropped_unrouted_message",
		zap.String("stream", stream),
		zap.String("device_addr", deviceAddr),
		zap.String("card_id", cardID),
		zap.String("device_type", deviceType),
		zap.String("hint", "device not in deviceRoom; config:card:stream subscriber should heal on next event"),
	)
}

// SetDailyLayoutReload 注入 daily layout reload 配置。
// hourLocal：0-23 表示 local time 时刻；-1 禁用。
// db：用于 SELECT rooms.layout_config；nil 时禁用。
// PR-15：管理员下班后定时重读 layout，hash 变化 → 重置该 room grid 从 0 重学。
func (e *Engine) SetDailyLayoutReload(hourLocal int, db *sql.DB) {
	if hourLocal < -1 || hourLocal > 23 {
		hourLocal = -1
	}
	e.dailyReloadHour = hourLocal
	e.dailyReloadDB = db
}

// Configure 注入 yaml 加载的运行时参数。零值字段保留 New 时的默认值。
// Persister 字段单独判断：显式传 nil 即关闭持久化（覆盖默认）。
func (e *Engine) Configure(cfg RuntimeConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if cfg.Decay.ImmediateSec > 0 {
		e.decayParams = cfg.Decay
	}
	if cfg.Learn.SitActiveX10 > 0 {
		e.learnParams = cfg.Learn
	}
	if cfg.ParamSets[0].Name != "" {
		e.paramSets = cfg.ParamSets
	}
	if cfg.DecayInterval > 0 {
		e.decayInterval = cfg.DecayInterval
	}
	if cfg.BeliefScanInterval > 0 {
		e.beliefScanInterval = cfg.BeliefScanInterval
	}
	if cfg.WinnerEvalInterval > 0 {
		e.winnerEvalInterval = cfg.WinnerEvalInterval
	}
	if cfg.SnapshotInterval > 0 {
		e.snapshotInterval = cfg.SnapshotInterval
	}
	// Persister 直接赋值（nil 也接受，表示禁用）
	e.persister = cfg.Persister

	// HistoryPersister 与 daily snapshot 时刻；时刻字段 0 视为未覆盖（保留默认 11:50/365 天）
	e.historyPersister = cfg.HistoryPersister
	if cfg.DailySnapshotHour != 0 || cfg.DailySnapshotMinute != 0 {
		e.dailySnapshotHour = cfg.DailySnapshotHour
		e.dailySnapshotMinute = cfg.DailySnapshotMinute
	}
	if cfg.HistoryRetainDays > 0 {
		e.historyRetainDays = cfg.HistoryRetainDays
	}

	// 风险时段（IsNightTime 用）—— 包级 var，所有房间共享
	SetRiskTimeConfig(cfg.RiskTime)

	// R4 参数：保存到 engine 级 + 下发到所有已注册房间的 TrackManager。
	// （RegisterRoom 在 Configure 之前调用的房间也要补设。）
	e.bedsideFallCfg = cfg.BedsideFall
	for _, tm := range e.rooms {
		tm.SetBedsideFallConfig(cfg.BedsideFall)
	}

	// alarm feedback：默认 5min；缺 DB 不启用
	e.feedbackDB = cfg.FeedbackDB
	if cfg.FeedbackInterval > 0 {
		e.feedbackInterval = cfg.FeedbackInterval
	} else if e.feedbackInterval == 0 {
		e.feedbackInterval = 5 * time.Minute
	}
	if e.feedbackDB != nil {
		e.feedbackIngester = NewAlarmFeedbackIngester(e.feedbackDB, e, e.logger)
	} else {
		e.feedbackIngester = nil
	}
}

// SetOutputCallback 设置 track 输出回调（发 alarm 等下游）
func (e *Engine) SetOutputCallback(fn func(roomID string, outputs []TrackOutput)) {
	e.onOutput = fn
}

// RoomForDevice 查 deviceKey（device_id 或 device_uid）对应的 room_id。
// 用于 alarm feedback 反查。空字符串 = 未路由（设备未绑定房间或未注册）。
func (e *Engine) RoomForDevice(deviceKey string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.deviceRoom[deviceKey]
}

// MountForRoom 取指定房间的雷达安装参数（用于 RadarToCanvas 转换）。
// found=false 表示 room 未注册或未 stamp radar。
func (e *Engine) MountForRoom(roomID string) (radarutils.RadarMount, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	mount, ok := e.mounts[roomID]
	return mount, ok
}

// MarkFakeAlarmAt 在指定房间的 cell at (x, y) 累计 FakeAlarmCount + 1。
// 用于 alarm_events.operation='false_alarm' 反馈链：
// 人工标 fake → 反查 trigger 位置 → 命中 cell → 影响未来 still fall 阈值。
// 房间不存在 / cell 越界 → false（调用方可记日志或忽略）。
func (e *Engine) MarkFakeAlarmAt(roomID string, x, y int, nowMs int64) bool {
	return e.ApplyToCell(roomID, x, y, nowMs, func(c *Cell) { c.IncrFakeAlarm() })
}

// ApplyToCell 通用 cell 修改接口。返回是否成功（cell 在 grid 内）。
// 用于 PR-6 alarm-feedback 按 conditions 分流到不同 counter。
// 调用方传 closure 直接操作 cell；engine 负责 grid 路由 + LastUpdateMs 更新。
func (e *Engine) ApplyToCell(roomID string, x, y int, nowMs int64, fn func(*Cell)) bool {
	e.mu.RLock()
	g := e.grids[roomID]
	e.mu.RUnlock()
	if g == nil {
		return false
	}
	cell := g.CellAt(x, y)
	if cell == nil {
		return false
	}
	fn(cell)
	if nowMs > cell.LastUpdateMs {
		cell.LastUpdateMs = nowMs
	}
	return true
}

func (e *Engine) MapDeviceToRoom(deviceKey, roomID string) {
	e.mu.Lock()
	e.deviceRoom[deviceKey] = roomID
	e.mu.Unlock()
}

// MapDeviceIDToUID 注册 device UUID → device_uid（PR-8 AI publish 反查源 radar UID 用）。
func (e *Engine) MapDeviceIDToUID(deviceID, deviceUID string) {
	if deviceID == "" || deviceUID == "" {
		return
	}
	e.mu.Lock()
	e.deviceIDToUID[deviceID] = deviceUID
	e.mu.Unlock()
}

// MapDeviceIDToType 注册 device UUID → 源 sensor 类型（"Radar"/"Sleepad"）。
// AI publish 时直接作 message.DeviceType（不再拼 AI 后缀）。缺失时兜底 "Radar"。
func (e *Engine) MapDeviceIDToType(deviceID, deviceType string) {
	if deviceID == "" || deviceType == "" {
		return
	}
	e.mu.Lock()
	e.deviceIDToType[deviceID] = deviceType
	e.mu.Unlock()
}

// DeviceUIDHex 反查 device addr (IPv6 字符串) → device_uid hex MAC（人眼可读）。
// 用于 log 双格式（IPv6 字符串机器友好 / hex MAC 人眼对照设备贴纸）。
// 缺失返回空字符串（caller 用 zap.String 仍可写空，不会 panic）。
func (e *Engine) DeviceUIDHex(deviceAddr string) string {
	if deviceAddr == "" {
		return ""
	}
	e.mu.RLock()
	hex := e.deviceIDToUID[deviceAddr]
	e.mu.RUnlock()
	return hex
}

// SetRoomTenant 注入 roomID → tenant_id（alarm_events.tenant_id 必填，从 rooms 表查）。
func (e *Engine) SetRoomTenant(roomID, tenantID string) {
	if roomID == "" {
		return
	}
	e.mu.Lock()
	e.roomTenants[roomID] = tenantID
	e.mu.Unlock()
}

// recordLastSrcSeq 记录最近一条来自 deviceAddr 的消息 envelope.SequenceNumber。
// AI verdict publish 时反查作 evidence.trigger_seq_num（reasoning trace 链锚定）。
func (e *Engine) recordLastSrcSeq(deviceAddr string, seq uint64) {
	if deviceAddr == "" || seq == 0 {
		return
	}
	e.srcSeqMu.Lock()
	e.lastSrcSeq[deviceAddr] = seq
	e.srcSeqMu.Unlock()
}

// readLastSrcSeq 读取最近 source seq；0 表示无记录或源 producer 未填 SequenceNumber。
func (e *Engine) readLastSrcSeq(deviceAddr string) uint64 {
	if deviceAddr == "" {
		return 0
	}
	e.srcSeqMu.RLock()
	defer e.srcSeqMu.RUnlock()
	return e.lastSrcSeq[deviceAddr]
}

// nextAgentSeq sensor 作为 platform agent producer 的单调 sequence number。
// 跨重启不重置（Redis INCR），让 trace_id = "<producer>.<seqN>" 全局唯一可追溯。
// Redis 不可用时 degrade 返 0（不阻塞 alarm 链；trace_id 会暂时为空，下次 publish 自愈）。
func (e *Engine) nextAgentSeq(ctx context.Context, producer string) int64 {
	if e.redisClient == nil || producer == "" {
		return 0
	}
	key := "wisefido-sensor:seq:" + producer
	v, err := e.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0
	}
	return v
}

// SetAIPublishConfig 注入 AI publish 单点配置：mode + node_id（来自 yaml/env）。
// 替换旧的 SetAIDeviceType + SetAIPublishEnabled，单一入口避免漂移。
//
// mode："log" 仅写 ai.log；"log&publish" 还会推 redis stream。任何模式下 alarm
// 的 fire 路径都不变（"宁可误报不可漏报"原则——mode 仅控制下游 stream，不 gate
// 告警生成）。
//
// source：完整节点身份字符串（"AI.Caregiver01" / "AI.Doctor01" 等），同时作
// wire fields["source"] 默认值 + ai_emit log 审计字段。直接由 config 注入，
// 不在此拼接——多实例横向扩展时各进程 source 不同即可。
// device_type 保持源 sensor 类型，AI 派生身份在 dataValue.source 表达。
func (e *Engine) SetAIPublishConfig(mode, source string) {
	if mode == "" {
		mode = "log&publish"
	}
	if source == "" {
		source = "sensor.caregiver01"
	}
	e.mu.Lock()
	e.aiPublishMode = mode
	e.aiSource = source
	e.mu.Unlock()
}

// publishEnabled 当前是否会真发到 redis（mode == "log&publish"）。
func (e *Engine) publishEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.aiPublishMode == "log&publish"
}

// SetSuiteCensus 注入进程级共享 SuiteCensusManager（sensor_v2 PR-2 数据结构 + PR-3 publish 关联）。
// nil = 禁用 PR-3 PersonID 关联（TrackStatus.PersonID 一律空）；
// bootstrap 调用方负责生命周期（含 SaveToRedis 定时任务）。
// SetWeakBioSource 注入 fall verifier 的 WeakBio 提级 source（A 风险放大消费者）。
// 已注册的 TrackManager 同步转发；新 RegisterRoom 创建的 TrackManager 启动时同样注入。
// nil 允许（短路 disable）。详 fall_verify.go §"WeakBioForceRealThreshold"。
func (e *Engine) SetWeakBioSource(s WeakBioSource) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.weakBioSource = s
	for _, tm := range e.rooms {
		tm.SetWeakBioSource(s)
	}
}

// OnDeviceUnfit 实现 DeviceFitnessTracker UnfitCallback —— device offline / SensorDetached
// 等触发时清该 device 在 engine 内的所有 in-memory state（"device offline = 内存重启"原则，
// 用户 2026-05-19 拍板）。
//
// 通过 deviceRoom 反查 device 所在 room → 调 tm.ClearDevice。device 跨多 room 不存在（v2
// device→room LPM 命中唯一），单 room dispatch 足够。
func (e *Engine) OnDeviceUnfit(deviceAddr string) {
	if e == nil || deviceAddr == "" {
		return
	}
	e.mu.RLock()
	roomID := e.deviceRoom[deviceAddr]
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return
	}
	tm.ClearDevice(deviceAddr)
	e.logger.Info("engine cleared device on unfit",
		zap.String("device_addr", deviceAddr),
		zap.String("room_id", roomID),
	)
}

func (e *Engine) SetSuiteCensus(m *SuiteCensusManager) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.suiteCensus = m
}

// SetBathroomFallRules 注入 PR-10 BathroomFallRules（§6.A 4 类 fall）。
// nil → 跳过 bathroom fall 判定（PR-9 后 v1 路径已删，相当于 fall 检测降级）。
// 实例级单例：所有 bathroom rooms 共用，内部 per-roomID 分桶状态。
func (e *Engine) SetBathroomFallRules(r *BathroomFallRules) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.bathroomFall = r
}

// SetBedroomFallRules 注入 PR-11 BedroomFallRules（§6.B 11b + 11c）。
// 仅 non-bathroom room 入；silent_fall 11a 仍走 TrackManager.scanSilentFallLeftBed。
func (e *Engine) SetBedroomFallRules(r *BedroomFallRules) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.bedroomFall = r
}

// BathroomFallRulesWired bootstrap invariant 检查（sensor_v2_known_limitations.md L4）。
// main.go 启动后调用；false → log.Fatal 拒启动。
func (e *Engine) BathroomFallRulesWired() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.bathroomFall != nil
}

// BedroomFallRulesWired bootstrap invariant 检查（同上）。
func (e *Engine) BedroomFallRulesWired() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.bedroomFall != nil
}

// SetGhostAdjudicators 注入 ghost 判定器二件套（sensor_v2 PR-4，决定 16 + §10.3.1）。
// general nil → fallback NoopGhostAdjudicator（PR-7 §4.B 落地前 = v1 行为）；
// bathroom nil → fallback NoopGhostAdjudicator（未注入真实 BathroomGhostAdjudicator 时不做 §4.A 判定）。
// 任一调用都是幂等替换，没有 hot swap 边界 —— PR-X 想换实现请重启 engine。
func (e *Engine) SetGhostAdjudicators(general, bathroom GhostAdjudicator) {
	if general == nil {
		general = NoopGhostAdjudicator{}
	}
	if bathroom == nil {
		bathroom = NoopGhostAdjudicator{}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.generalGhostAdj = general
	e.bathroomGhostAdj = bathroom
}

// GridForRoom 暴露 grid 给 BathroomGhostAdjudicator 用（lookup callback）。
// 走 RLock 读 e.grids — 与 RegisterRoom 加锁路径互斥但快速通过。
// 返回 nil = room 未注册（caller 据此 noop）。
func (e *Engine) GridForRoom(roomID string) *RoomGrid {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.grids[roomID]
}

// SuiteIDForRoom 暴露 roomID → suiteID 给 BathroomGhostAdjudicator 用。
func (e *Engine) SuiteIDForRoom(roomID string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.roomSuiteID[roomID]
}

// pickAdjudicator 按 room_type binary 分类挑选 adjudicator。读锁内调用。
// nil-safe：未调 SetGhostAdjudicators 时退化 Noop（保 PR-4 默认行为 == v1）。
func (e *Engine) pickAdjudicator(roomType int) GhostAdjudicator {
	if roomType == card.RoomTypeBathroom {
		if e.bathroomGhostAdj != nil {
			return e.bathroomGhostAdj
		}
	} else {
		if e.generalGhostAdj != nil {
			return e.generalGhostAdj
		}
	}
	return NoopGhostAdjudicator{}
}

// applyVerdictDeltas 应用 GhostAdjudicator 输出到 TrackStatus 副本（**唯一 verdict 写点**）。
//
// 不变量守卫（sensor_v2 决定 21 / Q4 review）：
//   Anchored → Ghost 转换被拒绝；verdict 字段保持 Anchored，penalty 仍累加供 PR-6 BathroomGhost
//   "持续疑似"审计观察。未来若需真正 unanchor，必须走显式重置路径（重启 / FeedbackEvent 清
//   LongSurvival/StartupGrace），不能在单帧 delta 里悄悄做。
//
// 边界：
//   - PenaltyDelta clamp 到 [0, 100]
//   - delta 引用不存在的 TrackID → 静默丢弃（adjudicator 看到的 base 是当前帧 snapshot，
//     长生命周期 stale delta 应当不会出现；warn log 留给单测验证）
//   - delta.Reason 为空时仍应用（PR-4 默认 Noop 不返 reason；PR-6 之后 adjudicator 必填）
func (e *Engine) applyVerdictDeltas(statuses []*TrackStatus, deltas []VerdictDelta) {
	if len(deltas) == 0 {
		return
	}
	byID := make(map[int]*TrackStatus, len(statuses))
	for _, s := range statuses {
		byID[s.TrackID] = s
	}
	for _, d := range deltas {
		s, ok := byID[d.TrackID]
		if !ok {
			continue
		}
		if d.NewVerdict != nil {
			newV := *d.NewVerdict
			if s.Verdict == VerdictAnchored && newV == VerdictGhost {
				e.logger.Warn("adjudicator_anchored_to_ghost_rejected",
					zap.Int("track_id", d.TrackID),
					zap.String("device_id", s.DeviceID),
					zap.String("room_id", s.RoomID),
					zap.String("reason", d.Reason),
				)
			} else {
				s.Verdict = newV
			}
		}
		if d.PenaltyDelta != 0 {
			next := s.GhostPenalty + d.PenaltyDelta
			if next < 0 {
				next = 0
			}
			if next > 100 {
				next = 100
			}
			s.GhostPenalty = next
		}
	}
}

// trackLostAnchorMs 失锁判定阈值：snapshot 未含某 trackID 持续 ≥ 60s → 视为真失锁，
// 调 SuiteCensusManager.ClearAnchorOnLostTrack。偏保守的取值（MaxMissCount=10 帧 coast
// 期之外再加足够 buffer），避免 firmware track coast 期间误清。
// 详见 suite_census.go ClearAnchorOnLostTrack 注释 "PR-3 wire 失锁判定建议"。
const trackLostAnchorMs int64 = 60_000

// publishTrackStatuses 把 TrackManager 的 Layer 1 投影 enrich 成 TrackStatus 列表，
// 跑 GhostAdjudicator → apply deltas → 推流。调用时机：handleMessage 在 tm.ProcessFrame 之后。
//
// 流水线（PR-4）：
//   1) 失锁清理：上一帧出现但本帧未出现且 ≥ 60s 的 trackID → census.ClearAnchorOnLostTrack（PR-3）
//   2) Build：bases → []*TrackStatus 副本（含 PR-3 PersonID 关联 / zone 占位推断）
//   3) Adjudicate：按 room.kind 挑 GhostAdjudicator（决定 16），调 Adjudicate(bases, nowMs)
//   4) Apply：applyVerdictDeltas 执行 Anchored 守卫 + penalty clamp（决定 21 / Q4）
//   5) Publish：每条 TrackStatus 推 sensor:track:status:stream
//
// PersonID 写入规则（sensor_v2 PR-3 review）：
//   if person, upgraded := suiteCensus.UpdatePersonFromTrack(...); upgraded && person != nil {
//       status.PersonID, status.PersonRole = person.PersonID, person.Role
//   } // else 一律空
//
// Bathroom room 不调 UpdatePersonFromTrack —— bathroom 内 person 关联由 PR-5 BathroomGate
// 入口流量 + suiteCensus.MarkPersonExitToBathroom/Return 维护，bathroom 帧只读 AnchorRoomType。
func (e *Engine) publishTrackStatuses(ctx context.Context, roomID string, bases []TrackStatusBase, nowMs int64) {
	if e.redisClient == nil {
		return
	}

	e.mu.RLock()
	suiteID := e.roomSuiteID[roomID]
	residentID := e.roomResidentID[roomID]
	roomType := e.roomType[roomID]
	census := e.suiteCensus
	adjudicator := e.pickAdjudicator(roomType)
	e.mu.RUnlock()

	isBathroom := roomType == card.RoomTypeBathroom

	// 步骤 1：失锁清理（trackLastSeen 仅在本 goroutine 读写，无锁需要）。
	seen := e.trackLastSeen[roomID]
	if seen == nil {
		seen = make(map[int]int64)
		e.trackLastSeen[roomID] = seen
	}
	curIDs := make(map[int]struct{}, len(bases))
	for i := range bases {
		curIDs[bases[i].TrackID] = struct{}{}
	}
	for tid, lastMs := range seen {
		if _, alive := curIDs[tid]; alive {
			continue
		}
		if nowMs-lastMs < trackLostAnchorMs {
			continue
		}
		if census != nil && suiteID != "" {
			census.ClearAnchorOnLostTrack(suiteID, tid)
		}
		delete(seen, tid)
	}

	// 步骤 1.5：PR-5 BathroomGate（仅 bathroom room + 已挂 census + suiteID 三条件齐备）。
	// gate 自己处理空 bases（依然要扫成员表做 timeout exit），所以放在 Build 之前调。
	if isBathroom && census != nil && suiteID != "" {
		gate, ok := e.bathroomGates[roomID]
		if !ok {
			gate = NewBathroomGate(roomID, suiteID, census, e.logger)
			e.bathroomGates[roomID] = gate
		}
		gate.Process(bases, nowMs)
	}

	// 步骤 1.6：PR-10 BathroomFallRules（gate 之后 fall 之前，保 BathroomCount 已最新）。
	// 仅 bathroom room 入；空 bases 也调（10c 最强档需要扫上帧 anchor）。
	if isBathroom && e.bathroomFall != nil {
		e.bathroomFall.Evaluate(roomID, bases, nowMs)
	}

	// 步骤 1.7：PR-11 BedroomFallRules（non-bathroom room 入；含 bedroom 默认 + 未来 living/kitchen）。
	// 11b BedsideFall 依赖 BedSession.LeftBedAtMs latch；从 TrackManager 取 snapshot 后传入。
	if !isBathroom && e.bedroomFall != nil {
		var beds []BedSessionLatch
		e.mu.RLock()
		tm := e.rooms[roomID]
		e.mu.RUnlock()
		if tm != nil {
			beds = tm.SnapshotBedSessions()
		}
		e.bedroomFall.Evaluate(roomID, bases, beds, nowMs)
	}

	// 步骤 2：Build TrackStatus 副本 + PR-3 PersonID 关联。
	statuses := make([]*TrackStatus, 0, len(bases))
	for i := range bases {
		b := &bases[i]
		seen[b.TrackID] = nowMs
		status := &TrackStatus{
			TrackID:      b.TrackID,
			DeviceID:     b.DeviceID,
			RoomID:       b.RoomID,
			Verdict:      b.Verdict,
			GhostPenalty: b.GhostPenalty,
			X:            b.X,
			Y:            b.Y,
			Z:            b.Z,
			Pose:         b.Pose,
			StillSec:     b.StillSec,
			CellAreaType: b.CellAreaType,
			EnterTarget:  b.EnterTarget,
			InRoomZoneID: roomID,
			UpdatedAtMs:  nowMs,
		}
		switch b.CellAreaType {
		case AreaBed:
			status.InBedZoneID = roomID
		case AreaToilet, AreaShower:
			status.InBathroomZoneID = roomID
		}
		if isBathroom {
			status.InBathroomZoneID = roomID
		}

		if census != nil && suiteID != "" && !isBathroom {
			person, upgraded := census.UpdatePersonFromTrack(
				suiteID, residentID,
				b.TrackID, b.SleepadInBed,
				b.TraverseDelta, b.MoveActive,
				nowMs,
			)
			if upgraded && person != nil {
				status.PersonID = person.PersonID
				status.PersonRole = person.Role
			}
		}
		statuses = append(statuses, status)
	}

	// 步骤 3-4：Adjudicator 跑 + apply delta（Anchored 守卫 + penalty clamp）。
	// PR-4 默认 adjudicator 是 Noop → deltas 空 → applyVerdictDeltas 直接 return。
	deltas := adjudicator.Adjudicate(bases, nowMs)
	e.applyVerdictDeltas(statuses, deltas)

	// 步骤 5：推流。
	for _, status := range statuses {
		PublishTrackStatus(ctx, e.redisClient, e.logger, status)
	}
}

// PublishAIEvent 发布 AI 派生 event。
//
// 路由（PR1 A10 后）：
//   - category=="track_verdict" → 走专用流 ai:track:verdict:stream（短 TTL=30s）
//     cardagg 端用 ai_verdict_handler 单独消费，喂入 aiOverrides cache
//   - 其他 category → 仍走 iot:event:stream（兼容历史路径）
//
// 旧路径：所有 category 都走 iot:event:stream，cardagg event_handler 统一消费 —
// 已切走以便后续 cardagg 整体停订阅 iot:event:stream（B 组迁移前置）。
// 旧代码可能仍传 "ghost" 兼容（按 track_verdict 等价处理）。
//
// 消息字段（Phase A v2 envelope）：
//   Producer:        e.aiSource（如 "sensor.caregiver01"），sensor agent layer 1 标识
//   SubjectEntity:   留空（AI 不染卡概念，cardagg IotPreparedHandler 反查 device→subject）
//   DeviceUID:       源 sensor 的 device_uid
//   DeviceID:        源 sensor 的 UUID
//   DeviceType:      源 sensor 类型（"Radar" / "Sleepad"）
//   TenantID:        roomTenants[roomID]
//   DataValue:       [{ track_id, ts, position_x/y/z, area_type, pose, track_confidence, source, ... }]
//
// 推送失败仅 warn 日志，不阻塞调用方。任何模式都打 ai_emit 结构化日志，作演示
// 与审计追溯依据；mode=log 时 published=false 仅 log，mode=log&publish 时尝试推流。
func (e *Engine) PublishAIEvent(ctx context.Context, p AIPayload, category string, nowMs int64) {
	streamDef := rediscommon.StreamEvent
	if category == "track_verdict" || category == "ghost" {
		streamDef = rediscommon.StreamAITrackVerdict
	}
	e.publishAIMessage(ctx, p, category, "event",
		streamDef.Name, streamDef.MaxLen, streamDef.RetentionSeconds, nowMs)
}

// PublishAIAlarm 发布 AI 派生 alarm 到 iot:alarm:stream。
// category ∈ {"silent_fall", "still_fall", "lost_fall", "silent_leftbed_fall", "bedside_fall"}
//
// alarm 路径不受 mode 影响 fire 决策（"宁可误报不可漏报"），mode 仅控制是否推到 stream。
func (e *Engine) PublishAIAlarm(ctx context.Context, p AIPayload, category string, nowMs int64) {
	e.publishAIMessage(ctx, p, category, "alarm",
		rediscommon.StreamAlarm.Name,
		rediscommon.StreamAlarm.MaxLen, rediscommon.StreamAlarm.RetentionSeconds, nowMs)
}

func (e *Engine) publishAIMessage(ctx context.Context, p AIPayload,
	category, topicType, streamName string, maxLen int64, retentionSec int, nowMs int64) {
	// device_ipv6 单程票：p.DeviceID 现为 canonical IPv6 字符串（上游已切；engine 内部 Map 同步切）
	addr, _ := netip.ParseAddr(p.DeviceID)
	e.mu.RLock()
	baseType := e.deviceIDToType[p.DeviceID]
	defaultSource := e.aiSource
	mode := e.aiPublishMode
	g := e.grids[p.RoomID]
	e.mu.RUnlock()
	// SubjectEntity 留空：sensor 不做 device→card 反查（非其职责）；
	// cardagg alarm_router 在 SubjectEntity 空时调 metaCache.LookupCardByDeviceAddr LPM 兜底。

	if baseType == "" {
		baseType = "Radar" // 兜底：路由表缺失时默认按 radar 派生
	}
	// device_type 保持源 sensor 类型（不再拼 ".AI<NodeID>" 后缀）。
	// AI 派生身份由 fields["source"] 一等公民字段表达。
	deviceType := baseType

	// alarm/event 流统一走 EventItem 契约（与 qinglan/sleepace publisher 一致）：
	// EventItem 提供生命周期 + first-class 业务字段（TrackID/Pose/HeartRate/RespiratoryRate）；
	// 其余 sensor-specific 字段（position / track_confidence / area_type / source / reason / evidence）
	// 作为 dataValue map 同级平铺补充。
	//
	// EventStatus：默认 "instant"；payload 显式指定（如 RecordRadarAlarm forward Initialization→end）
	// 时覆盖，让下游 cardagg AlarmRouter 按 EndPolicy=AutoResolve 关 alarm。
	eventStatus := p.EventStatus
	if eventStatus == "" {
		eventStatus = "instant"
	}
	item := observation.NewEventItem(nowMs, eventStatus)
	item.TrackID = p.Track.TrackID
	if p.Track.Pose != 0 {
		item.Pose = p.Track.Pose
	}
	if p.Track.HeartRate != 0 {
		item.HeartRate = p.Track.HeartRate
	}
	if p.Track.RespiratoryRate != 0 {
		item.RespiratoryRate = p.Track.RespiratoryRate
	}
	fields, _ := observation.EventItemToDataMap(&item)
	if fields == nil {
		fields = make(map[string]interface{})
	}
	// sensor-specific 业务扩展字段平铺
	if p.Track.PositionX != nil {
		fields[observation.FieldPositionX] = *p.Track.PositionX
	}
	if p.Track.PositionY != nil {
		fields[observation.FieldPositionY] = *p.Track.PositionY
	}
	if p.Track.PositionZ != nil {
		fields[observation.FieldPositionZ] = *p.Track.PositionZ
	}
	if p.Track.TrackConfidence != 0 {
		fields[observation.FieldTrackConfidence] = p.Track.TrackConfidence
	}
	// AI 派生 track_verdict 与床状态无关；仅 sleepad_radar_conflict 显式传 BedStatus 才保留。
	if p.Track.BedStatus != nil {
		fields[observation.FieldBedStatus] = *p.Track.BedStatus
	}
	// area_type engine 自己算（observation.Track 的 AreaType 是字符串，engine 这边类型不同）
	if g != nil {
		px, py := 0, 0
		if p.Track.PositionX != nil {
			px = *p.Track.PositionX
		}
		if p.Track.PositionY != nil {
			py = *p.Track.PositionY
		}
		if cell := g.CellAt(px, py); cell != nil {
			fields["area_type"] = areaTypeProtocolStr(cell.Belief[0].Type)
		}
	}
	// PR5b: Source（一等公民）+ Reason / Evidence（审计元数据）
	// Source 默认 = e.aiSource（来自 cfg.AIPublish.Source，如 "AI.Caregiver01"）。
	// p.Source 非空时尊重 caller override（未来多角色场景，如健康风险模块发 verdict）
	source := p.Source
	if source == "" {
		source = defaultSource
	}
	if source != "" {
		fields["source"] = source
	}
	if p.Reason != "" {
		fields["reason"] = p.Reason
	}
	if len(p.Evidence) > 0 {
		fields["evidence"] = p.Evidence
	}
	// 北极星 reasoning trace：verdict 必带触发它的 source envelope.seq —
	// 下游审计可一句 grep 把 AI verdict 反向链回 producer 的具体 envelope。
	if srcSeq := e.readLastSrcSeq(p.DeviceID); srcSeq != 0 {
		fields["trigger_seq_num"] = srcSeq
	}

	willPublish := e.redisClient != nil && mode == "log&publish"

	// 预先取 producer + seq，让 ai_emit log 带 trace_id（"<producer>.<seqN>"），
	// 跨服务 grep trace_id 即可 join sensor→cardagg→data 整条链。
	producer := source
	if producer == "" {
		producer = defaultSource
	}
	seq := e.nextAgentSeq(ctx, producer)
	traceID := fmt.Sprintf("%s.%d", producer, seq)

	// 任何模式都打 ai_emit 审计日志：sandbox 演示靠这条 log 看 AI 在思考
	e.logger.Info("ai_emit",
		zap.String("trace_id", traceID),
		zap.String("source", source),
		zap.String("mode", mode),
		zap.String("device_type", deviceType),
		zap.String("device_addr", p.DeviceID),
		zap.String("device_uid_hex", e.DeviceUIDHex(p.DeviceID)),
		zap.String("category", category),
		zap.String("topic_type", topicType),
		zap.String("would_publish_to", streamName),
		zap.Bool("published", willPublish),
		zap.Int("track_id", p.Track.TrackID),
		zap.Int("track_confidence", p.Track.TrackConfidence),
		zap.Int64("ts_ms", nowMs),
	)

	if !willPublish {
		return
	}

	// device_ipv6 单程票：Producer = sensor agent /128 INET；
	// SubjectEntity 留空（cardagg LPM 反查兜底，见上注释）；
	// DeviceAddr = p.DeviceID parse 后 /128。
	msg := rediscommon.IoTStreamMessage{
		Producer:       producer,
		SequenceNumber: uint64(seq),
		SubjectEntity:  "",
		DeviceAddr:     addr,
		DeviceType:     deviceType,
		Timestamp:      nowMs,
		TopicType:      topicType,
		Category:       category,
		DataValue:      []interface{}{fields},
	}
	if _, err := rediscommon.PublishToStream(ctx, e.redisClient, streamName, msg.ToStreamMap(), maxLen, retentionSec); err != nil {
		e.logger.Warn("ai_publish_failed",
			zap.String("source", source),
			zap.String("stream", streamName),
			zap.String("category", category),
			zap.String("device_addr", p.DeviceID),
			zap.Error(err),
		)
	}
}

// areaTypeProtocolStr 把 engine 的 AreaType 映射成 observation.EnumAreaType 协议字符串。
// 协议：0=none / 1=custom / 2=bed / 3=interfer / 4=door / 5=monitor_bed / 6=sensing
// engine 内部枚举跟协议不完全对齐；做一个 best-effort 映射，cardagg 端按 device_type=Radar 处理。
func areaTypeProtocolStr(t AreaType) string {
	switch t {
	case AreaBed:
		return "bed"
	case AreaSit:
		return "custom" // 沙发/椅子归 custom（协议没单独 sit 类）
	case AreaToilet, AreaShower:
		return "monitor_bed" // 卫浴归 monitor_bed（最接近的 high-risk 区）
	case AreaEnter:
		return "door"
	case AreaDeny:
		return "interfer"
	case AreaActive:
		return "sensing"
	}
	return "none"
}

// GetRoomOutputs 查询某房间最新 track 输出
func (e *Engine) GetRoomOutputs(roomID string) []TrackOutput {
	e.mu.RLock()
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		return nil
	}
	return tm.GetOutputs()
}

// ========================================================================
// RegisterRoom：构建 grid + rasterize 物理/几何/先验
// ========================================================================

func (e *Engine) RegisterRoom(cfg RoomConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if cfg.RoomW <= 0 || cfg.RoomW > radarutils.MaxRoomWidth {
		e.logger.Warn("room width out of range, using default",
			zap.Int("requested", cfg.RoomW), zap.Int("max", radarutils.MaxRoomWidth))
		cfg.RoomW = radarutils.MaxRoomWidth
	}
	if cfg.RoomH <= 0 || cfg.RoomH > radarutils.MaxRoomHeight {
		cfg.RoomH = radarutils.MaxRoomHeight
	}

	// 优化 grid 范围：bbox(WallPolygon ∪ FOV) + 4 cells，cap 600cm
	// 对 cell 索引语义有影响 → 旧 snapshot 的 layoutHash 也会随之变（如预期）
	rawW, rawH := cfg.RoomW, cfg.RoomH
	ApplyOptimizedExtent(&cfg)

	// 幂等 short-circuit：room 已注册 + layout 未变 → 仅刷 soft 字段，**不**重建 TrackManager/grid。
	// 否则 config:card.changed 触发的 ReloadRooms 会把所有 17 个 room 的 bedSessions / fall pending /
	// trackLastSeen / sleepadStates 等 in-memory 状态归零，导致正在床上 / 跌倒判定窗口丢失。
	// layout 真的改了（用户编辑了 wall/bed/radar 等）→ 通过 hash 比对触发完整重建（旧行为）。
	newHash := LayoutHash(cfg)
	if tm, exists := e.rooms[cfg.RoomID]; exists && e.layoutHashes[cfg.RoomID] == newHash {
		// soft-only：room name / timezone / 静态归属（resident_id / suite_id / room_type）+ public bathroom 标记
		if cfg.RoomName != "" {
			tm.SetRoomName(cfg.RoomName)
		}
		if cfg.Timezone != "" {
			if loc, err := time.LoadLocation(cfg.Timezone); err == nil {
				tm.SetTimezone(loc)
			}
		}
		e.roomSuiteID[cfg.RoomID] = cfg.SuiteID
		e.roomResidentID[cfg.RoomID] = cfg.ResidentID
		e.roomType[cfg.RoomID] = cfg.RoomType
		if cfg.RoomType == card.RoomTypeBathroom && cfg.IsPublicBathroom && e.suiteCensus != nil && cfg.SuiteID != "" {
			e.suiteCensus.MarkPublicBathroom(cfg.SuiteID) // 幂等
		}
		return
	}

	// 1. 创建空 grid，覆盖优化后的 RoomW × RoomH
	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	grid.OriginX = cfg.OriginX
	grid.OriginY = cfg.OriginY

	// 2. Stamp Wall 多边形 → InRoom
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}

	// 3. Stamp Radar 物理 FOV → InFOV / EdgeDist / MaxZ / MinZ
	grid.StampRadar(cfg.Radar)

	// 4. Stamp Enters → 记 Enters 列表 + 覆写矩形内 InRoom=true（门洞可穿）
	grid.StampEnters(cfg.Enters, cfg.EnterTargets)

	// 5. SetPrior 人标矩形（AreaType + Confidence + Source）
	for _, r := range cfg.Enters {
		grid.SetPrior(r, AreaEnter, 99, SourceHuman)
	}
	for _, r := range cfg.Beds {
		grid.SetPrior(r, AreaBed, 99, SourceHuman)
	}
	for _, r := range cfg.Toilets {
		grid.SetPrior(r, AreaToilet, 99, SourceHuman)
	}
	for _, r := range cfg.Showers {
		grid.SetPrior(r, AreaShower, 99, SourceHuman)
	}
	for _, r := range cfg.Chairs {
		grid.SetPrior(r, AreaSit, 80, SourceHuman) // 粗标 Conf=80
	}
	for _, r := range cfg.Furnitures {
		grid.SetPrior(r, AreaDeny, 99, SourceHuman)
	}
	for _, r := range cfg.Interferes {
		grid.SetPrior(r, AreaDeny, 99, SourceHuman)
	}

	e.grids[cfg.RoomID] = grid
	e.mounts[cfg.RoomID] = cfg.Radar
	tm := NewTrackManager(cfg.RoomID, grid)
	tm.SetMoveSpeedCms(e.learnParams.MoveSpeedCms)
	tm.SetBedsideFallConfig(e.bedsideFallCfg)
	tm.SetLogger(e.logger)
	tm.SetRoomName(cfg.RoomName)
	tm.SetInterferes(cfg.Interferes)
	tm.SetRadarMount(cfg.Radar) // L1 mirror pair 检测用 radar 中心做 ghost tiebreaker
	// PR-8: 注入 AI 派生事件 / 告警发布器（engine 实现 AIPublisher 接口）
	tm.SetAIPublisher(e)
	// A 风险放大消费者: WeakBio≥80 force real 短路 source（engine 在 SetWeakBioSource 时
	// 同步转发给已注册 tm；新 RegisterRoom 在这里补设。nil 允许 — verifier 走原三档评分）
	if e.weakBioSource != nil {
		tm.SetWeakBioSource(e.weakBioSource)
	}
	// 注入 IANA 时区（IsNightTime 用）；空串保持 nil → IsNightTime 退化 UTC
	if cfg.Timezone != "" {
		if loc, err := time.LoadLocation(cfg.Timezone); err == nil {
			tm.SetTimezone(loc)
		} else {
			e.logger.Warn("invalid timezone, falling back to UTC",
				zap.String("room_id", cfg.RoomID),
				zap.String("timezone", cfg.Timezone),
				zap.Error(err))
		}
	} else {
		e.logger.Warn("room registered without timezone, IsNightTime will use UTC (likely wrong)",
			zap.String("room_id", cfg.RoomID))
	}
	e.rooms[cfg.RoomID] = tm

	// sensor_v2 PR-3 wiring：记录 room → (suite_id, resident_id, room_type)。
	e.roomSuiteID[cfg.RoomID] = cfg.SuiteID
	e.roomResidentID[cfg.RoomID] = cfg.ResidentID
	e.roomType[cfg.RoomID] = cfg.RoomType

	// sensor_v2 PR-7：public bathroom 显式标记 census bucket。
	// 仅当 RoomType==Bathroom 且 IsPublicBathroom=true 才触发，其它 room 即使 IsPublicBathroom=true 也忽略。
	// 幂等：MarkPublicBathroom 多次调同 SuiteID 等价（PR-2 实现）。
	if cfg.RoomType == card.RoomTypeBathroom && cfg.IsPublicBathroom && e.suiteCensus != nil && cfg.SuiteID != "" {
		e.suiteCensus.MarkPublicBathroom(cfg.SuiteID)
		e.logger.Info("room_registered_as_public_bathroom",
			zap.String("room_id", cfg.RoomID),
			zap.String("suite_id", cfg.SuiteID),
		)
	}

	// 保存 layout hash（snapshot save/load 都按此 hash 比对）— 上面 short-circuit 已算过
	e.layoutHashes[cfg.RoomID] = newHash
	hash := newHash

	e.logger.Info("room registered",
		zap.String("room_id", cfg.RoomID),
		zap.String("room_name", cfg.RoomName),
		zap.Int("raw_w", rawW),
		zap.Int("raw_h", rawH),
		zap.Int("grid_w", cfg.RoomW),
		zap.Int("grid_h", cfg.RoomH),
		zap.Int("enters", len(cfg.Enters)),
		zap.Int("beds", len(cfg.Beds)),
		zap.Int("toilets", len(cfg.Toilets)),
		zap.Int("showers", len(cfg.Showers)),
		zap.Int("furnitures", len(cfg.Furnitures)),
		zap.Int("cells", grid.Width*grid.Height),
		zap.String("layout_hash", hash[:12]),
	)

	// 尝试加载已有 snapshot（hydrate Belief + counters）
	// persister 可能未配置（dev/test 模式），用 background ctx 容忍 DB 慢
	if e.persister != nil {
		e.hydrateRoom(cfg.RoomID, grid, hash)
	}
}

// hydrateRoom 从 persister 取回 snapshot 并灌入 grid。
// 失败/不匹配/无记录都不致命——退化为 fresh start（baseline 已由 SetPrior 烤好）。
// 调用方必须持有 e.mu.Lock。
func (e *Engine) hydrateRoom(roomID string, grid *RoomGrid, expectedHash string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	storedHash, payload, found, err := e.persister.Load(ctx, roomID)
	if err != nil {
		e.logger.Warn("snapshot load failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	if !found {
		e.logger.Info("no snapshot found, fresh start", zap.String("room_id", roomID))
		return
	}
	if storedHash != expectedHash {
		e.logger.Warn("snapshot layout_hash mismatch, discarding (layout edited?)",
			zap.String("room_id", roomID),
			zap.String("stored", storedHash[:12]),
			zap.String("expected", expectedHash[:12]),
		)
		return
	}

	snap, err := UnmarshalSnapshot(payload)
	if err != nil {
		e.logger.Warn("snapshot unmarshal failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	if err := DecodeSnapshot(snap, grid); err != nil {
		e.logger.Warn("snapshot decode failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	e.logger.Info("snapshot hydrated",
		zap.String("room_id", roomID),
		zap.Int("cells", len(snap.Cells)),
	)
}

// ========================================================================
// Run：订阅 iot:monitor:stream 主循环 + 后台定时任务
// ========================================================================

func (e *Engine) Run(ctx context.Context) error {
	const (
		monitorStream = "iot:monitor:stream"
		eventStream   = "iot:event:stream"
		alarmStream   = "iot:alarm:stream"
		group         = "roomengine"
	)

	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, monitorStream, group); err != nil {
		e.logger.Warn("create consumer group (monitor)", zap.Error(err))
	}
	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, eventStream, group); err != nil {
		e.logger.Warn("create consumer group (event)", zap.Error(err))
	}
	if err := rediscommon.CreateConsumerGroup(ctx, e.redisClient, alarmStream, group); err != nil {
		e.logger.Warn("create consumer group (alarm)", zap.Error(err))
	}

	// 后台定时任务
	go e.decayLoop(ctx)
	go e.beliefScanLoop(ctx)
	go e.winnerEvalLoop(ctx)
	if e.persister != nil && e.snapshotInterval > 0 {
		go e.snapshotLoop(ctx)
	}
	if e.feedbackIngester != nil && e.feedbackInterval > 0 {
		go e.alarmFeedbackLoop(ctx)
	}
	// PR-15: daily layout reload — 管理员下班后重读 layout_config，hash 变 → reset grid
	if e.dailyReloadHour >= 0 && e.dailyReloadDB != nil {
		go e.dailyLayoutReloadLoop(ctx)
	}
	// Daily history snapshot — 每天指定时刻归档 grid 状态，保留 365 天
	if e.historyPersister != nil && e.dailySnapshotHour >= 0 {
		go e.dailySnapshotLoop(ctx)
	}
	// 单独 goroutine 消费 event 流（sleepad + radar 的 InBed/LeftBed/EnterRoom/ExitRoom）
	go e.runEventLoop(ctx, eventStream, group)
	// 单独 goroutine 消费 alarm 流（radar Fall 等）
	go e.runAlarmLoop(ctx, alarmStream, group)
	// 路由表 reload 由 config:card:stream subscriber 事件驱动（cmd/wisefido-sensor main.go configCardConsumer）。

	e.logger.Info("room engine started",
		zap.String("monitor_stream", monitorStream),
		zap.String("event_stream", eventStream),
		zap.String("alarm_stream", alarmStream),
		zap.String("winner", e.paramSets[e.winner].Name),
		zap.Bool("persist", e.persister != nil),
	)

	for {
		select {
		case <-ctx.Done():
			// 优雅退出：dump 一次再走（避免最近 5 min 学习丢失）
			if e.persister != nil {
				e.saveAllRooms(context.Background())
			}
			e.logger.Info("room engine stopped")
			return nil
		default:
		}

		messages, err := rediscommon.ReadFromStream(ctx, e.redisClient, monitorStream, group, "roomengine-monitor-1", 50)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, msg := range messages {
			e.handleMessage(ctx, msg)
			// XAck 每条消息（即便路由失败也 ACK——路由层已 warnUnrouted 暴露问题，
			// 不 ACK 只会让 pending 列表无限增长直到 redis 内存爆）。
			e.redisClient.XAck(ctx, monitorStream, group, msg.ID)
		}
	}
}

// runEventLoop 消费 iot:event:stream，路由 sleepad InBed/LeftBed 事件到 TrackManager
func (e *Engine) runEventLoop(ctx context.Context, stream, group string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		messages, err := rediscommon.ReadFromStream(ctx, e.redisClient, stream, group, "roomengine-event-1", 50)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, msg := range messages {
			e.handleEventMessage(msg)
			e.redisClient.XAck(ctx, stream, group, msg.ID)
		}
	}
}

// handleEventMessage 处理 iot:event:stream 一条消息（sleepad + radar 来源都消费）
func (e *Engine) handleEventMessage(msg rediscommon.StreamMessage) {
	m, err := rediscommon.FromStreamMap(msg.Values)
	if err != nil {
		return
	}
	dt := strings.ToLower(m.DeviceType)
	if dt != "sleepad" && dt != "sleeppad" && dt != "radar" {
		return
	}
	ts := m.Timestamp
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	// 路由到房间（device_ipv6 单程票：addr 是唯一 device 标识）
	addrStr := m.DeviceAddr.String()
	// 北极星 reasoning trace：记录这条 envelope.seq → 后续 AI verdict 引用为 trigger_seq_num
	e.recordLastSrcSeq(addrStr, m.SequenceNumber)
	e.mu.RLock()
	roomID := e.deviceRoom[addrStr]
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		e.warnUnrouted("event", m.SubjectEntity, addrStr, m.DeviceType)
		return
	}

	switch dt {
	case "sleepad", "sleeppad":
		for _, evt := range ParseSleepadBedEvents(m.DataValue, addrStr, m.Category, ts) {
			tm.ProcessSleepadBedEvent(evt)
		}
	case "radar":
		// 2026-05-15 起 radar firmware Fall/SittingOnGround 也走 event stream（gateway 分流）。
		// 见 doc/cardagg_sensor_split.md：cardagg 不再处理 radar producer Fall；sensor 接管 verifier。
		// Suspected* (pose=2/7) 是 firmware qualification 前的预警态，30~90s 后自动升级到 Fall/SittingOnGround。
		// Per Option C (2026-05-17): server 不双层 — Suspected* 只让 iot 自动写 event_log（审计链不丢），
		// 不进 verifier / 不 forward 到 iot:alarm:stream / 不入 alarm_events。
		// 见 memory firmware_fall_qualification + doc/cardagg_sensor_split.md
		if m.Category == alarm.SuspectedFall || m.Category == alarm.SuspectedSittingOnGround {
			e.logger.Debug("radar suspected (event_log only, no alarm forward)",
				zap.String("device_addr", addrStr),
				zap.String("category", m.Category),
				zap.Int64("ts_ms", ts),
			)
			return
		}
		if m.Category == alarm.Fall || m.Category == alarm.SittingOnGround {
			alarms := ParseRadarFallAlarm(m.DataValue, addrStr, m.Category, ts)
			for _, a := range alarms {
				tm.RecordRadarAlarm(a)
				e.logger.Info("radar_fall_received_via_event_stream",
					zap.String("device_addr", addrStr),
					zap.String("device_uid_hex", e.DeviceUIDHex(addrStr)),
					zap.String("category", m.Category),
					zap.Int("track_id", a.TrackID),
					zap.Int("pose", a.Pose),
					zap.String("status", a.Status),
					zap.String("room_id", roomID),
					zap.Int64("ts_ms", a.TMs),
					zap.String("ts_human", time.UnixMilli(a.TMs).Format("15:04:05.000")),
				)
			}
			tm.Tick(ts)
			return
		}
		// 落账 radar EnterRoom/ExitRoom/InBed/LeftBed；同时 InBed/LeftBed 走"事件触发器"
		// 路径：tm.RecordRadarEvent + tm.Tick(ts) → 段 4/5/6 立即跑一次。
		// 当前不消费 EnterRoom/ExitRoom 做行为推断，仅落账供未来段 7 使用。
		evts := ParseRadarTrackEvents(m.DataValue, addrStr, m.Category, ts)
		if len(evts) == 0 {
			return
		}
		for _, evt := range evts {
			tm.RecordRadarEvent(evt)
		}
		tm.Tick(ts)
	}
}

// runAlarmLoop 消费 iot:alarm:stream，路由 radar Fall 报警（仅落账 + Tick 触发，不做 verify）
func (e *Engine) runAlarmLoop(ctx context.Context, stream, group string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		messages, err := rediscommon.ReadFromStream(ctx, e.redisClient, stream, group, "roomengine-alarm-1", 50)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, msg := range messages {
			e.handleAlarmMessage(msg)
			e.redisClient.XAck(ctx, stream, group, msg.ID)
		}
	}
}

// handleAlarmMessage 处理 iot:alarm:stream 一条消息。
//
// 2026-05-15 gateway 分流后：alarm stream 上的 radar Fall 几乎全是 sensor 自己 emit 的回声
// （RecordRadarAlarm 转发回 alarm stream，producer="sensor.caregiver01"）。
// 必须 guard 防自循环：检测到 producer 是 sensor 自身 emit 时直接跳过，不再 RecordRadarAlarm + emit。
//
// 真正的 firmware radar Fall 现在走 iot:event:stream → handleEventMessage 接管。
func (e *Engine) handleAlarmMessage(msg rediscommon.StreamMessage) {
	m, err := rediscommon.FromStreamMap(msg.Values)
	if err != nil {
		return
	}
	// 自循环 guard：sensor 自身 emit 的 alarm 不重新处理（否则 RecordRadarAlarm → 再 emit → 再消费 → 无限循环）
	// 2026-05-25 修：sensor publish 用 platform-agent IP (fd00:0:fff1::1) 作 producer
	// （见 memory platform_agent_addressing），旧字符串 "sensor.caregiver01"/"wisefido-sensor"
	// 已不再使用 → 必须用 IsPlatformAgentAddr 识别，否则 self-loop guard 失效 (实测放大 43729×)。
	if spatial.IsPlatformAgentAddr(m.Producer) {
		return
	}
	if !strings.EqualFold(m.DeviceType, "radar") {
		return // 非 radar 报警暂不消费
	}
	ts := m.Timestamp
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	// 路由（device_ipv6 单程票：addr 是唯一 device 标识）
	addrStr := m.DeviceAddr.String()
	// 北极星 reasoning trace：记录这条 envelope.seq → 后续 AI verdict 引用
	e.recordLastSrcSeq(addrStr, m.SequenceNumber)
	e.mu.RLock()
	roomID := e.deviceRoom[addrStr]
	tm := e.rooms[roomID]
	e.mu.RUnlock()
	if tm == nil {
		e.warnUnrouted("alarm", m.SubjectEntity, addrStr, m.DeviceType)
		return
	}

	alarms := ParseRadarFallAlarm(m.DataValue, addrStr, m.Category, ts)
	if len(alarms) == 0 {
		return
	}
	for _, a := range alarms {
		tm.RecordRadarAlarm(a)
		// kind=radar_fall_received: radar firmware 发的 Fall 报警；engine 当前不验证（段 7 待做）。
		// 未来 verifier 上线后会进一步分流：fake_fall（否决）/ real_fall（确认）/ lost_fall（应报漏报）
		e.logger.Info("radar_fall_received",
			zap.String("device_addr", addrStr),
			zap.String("device_uid_hex", e.DeviceUIDHex(addrStr)),
			zap.Int("track_id", a.TrackID),
			zap.String("kind", "radar_firmware_fall"),
			zap.Int("pose", a.Pose),
			zap.String("status", a.Status),
			zap.String("room_id", roomID),
			zap.Int64("ts_ms", a.TMs),
			zap.String("ts_human", time.UnixMilli(a.TMs).Format("15:04:05.000")),
		)
	}
	tm.Tick(ts)
}

// ========================================================================
// 消息处理 + 坐标转换 + 无效帧过滤
// ========================================================================

func (e *Engine) handleMessage(_ context.Context, msg rediscommon.StreamMessage) {
	// 真实生产消息是 flat key（device_uid/dataValue/.../timestamp 都是 stream Values 顶层 key），
	// 不是包在 "data" 单 key 里的 JSON。用 FromStreamMap 与 cardagg/wisefido-data 一致解析。
	m, err := rediscommon.FromStreamMap(msg.Values)
	if err != nil {
		return
	}

	// 路由到房间（device_ipv6 单程票：addr 是唯一 device 标识）
	addrStr := m.DeviceAddr.String()
	// 北极星 reasoning trace：记录这条 envelope.seq → 后续 AI verdict 引用
	e.recordLastSrcSeq(addrStr, m.SequenceNumber)
	e.mu.RLock()
	roomID := e.deviceRoom[addrStr]
	tm := e.rooms[roomID]
	mount, hasMount := e.mounts[roomID]
	e.mu.RUnlock()

	if tm == nil {
		e.warnUnrouted("monitor", m.SubjectEntity, addrStr, m.DeviceType)
		return
	}

	ts := m.Timestamp
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	switch strings.ToLower(m.DeviceType) {
	case "radar":
		if !hasMount {
			return
		}
		frames := ParseRadarTracks(m.DataValue, addrStr, mount, ts)
		if len(frames) == 0 {
			// tid=88 heartbeat / 全零无效帧被 ParseRadarTracks 过滤后，仍要 tick 推进 MissCount，
			// 否则之前活着的 track 永不进入消失判定 → silent/lost fall pending 不会创建。
			tm.Tick(ts)
			return
		}
		outputs := tm.ProcessFrame(frames)
		if e.onOutput != nil && len(outputs) > 0 {
			e.onOutput(roomID, outputs)
		}
		// sensor_v2 PR-3：Layer 1 → Layer 2 TrackStatus 投影。
		// 只在 radar 帧后派生（sleepad 帧不带 track 几何）；
		// SnapshotTrackStatuses 自带 tm.mu，与 ProcessFrame 串行无竞态。
		bases := tm.SnapshotTrackStatuses(ts)
		e.publishTrackStatuses(context.Background(), roomID, bases, ts)

	case "sleepad", "sleeppad":
		// 多传感器融合：sleepad 帧喂入 TrackManager，silent fall 触发时做 short-circuit
		obs := ParseSleepadObservations(m.DataValue, addrStr, addrStr, ts)
		for _, o := range obs {
			tm.ProcessSleepadObservation(o)
		}

		// 兼容老接口：把"在床人数"也推给 TrackManager（occupancyFactor 用）
		inBedCount := 0
		for _, o := range obs {
			if o.InBed {
				inBedCount++
			}
		}
		tm.SetSleepadInBedCount(inBedCount)
	}
}

// ========================================================================
// 后台定时任务
// ========================================================================

// decayLoop 每 decayInterval 对所有 grid 做一次 DecayAll
func (e *Engine) decayLoop(ctx context.Context) {
	ticker := time.NewTicker(e.decayInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.mu.RLock()
			dtSec := e.decayInterval.Seconds()
			dp := e.decayParams
			for _, grid := range e.grids {
				grid.DecayAll(dtSec, dp)
			}
			e.mu.RUnlock()
			e.logger.Debug("decay all rooms done", zap.Float64("dt_sec", dtSec))
		}
	}
}

// beliefScanLoop 每 beliefScanInterval 对所有 cell 跑 UpdateBelief（3 组并行）
func (e *Engine) beliefScanLoop(ctx context.Context) {
	ticker := time.NewTicker(e.beliefScanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.scanBeliefAll()
		}
	}
}

// scanBeliefAll 全量扫每个 grid 的每 cell：
//  1. cell_learning（硬阈值规则）：Walk/Sit 升格 + 床外 Lie 异常累计
//  2. UpdateBelief（软概率规则）：3 组参数各 UpdateBelief 一次
//
// 两者顺序：硬规则先跑（确定性强），UpdateBelief 后跑做细粒度调整。
func (e *Engine) scanBeliefAll() {
	e.mu.RLock()
	defer e.mu.RUnlock()
	totalCells := 0
	totalLieAnomalies := 0
	lp := e.learnParams
	nowMs := time.Now().UnixMilli()
	for _, grid := range e.grids {
		grid.LearnCellAreas(lp, nowMs)
		totalLieAnomalies += grid.LearnLyingAnomalies(lp)
		for i := range grid.Cells {
			for g := 0; g < 3; g++ {
				grid.Cells[i].UpdateBelief(g, e.paramSets[g])
			}
		}
		totalCells += len(grid.Cells)
	}
	e.logger.Debug("belief scan done",
		zap.Int("total_cells", totalCells),
		zap.Int("lie_anomalies", totalLieAnomalies),
		zap.Int("winner", e.winner),
	)
}

// alarmFeedbackLoop 每 feedbackInterval 跑一次 alarm_events false_alarm 反馈同步。
// 启动时立即跑一次（首次会全量扫历史，把过往 false_alarm 灌入 cell.FakeAlarmCount）。
// 单次失败只警告，不退出循环。
func (e *Engine) alarmFeedbackLoop(ctx context.Context) {
	if e.feedbackIngester == nil {
		return
	}
	// 启动延迟 5s，避免与 RegisterRoom / hydrate 同时进 DB
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
	}
	if n, err := e.feedbackIngester.IngestOnce(ctx); err != nil {
		e.logger.Warn("alarm_feedback ingest at startup", zap.Error(err))
	} else {
		e.logger.Info("alarm_feedback startup ingest", zap.Int("processed", n))
	}
	ticker := time.NewTicker(e.feedbackInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := e.feedbackIngester.IngestOnce(ctx); err != nil {
				e.logger.Warn("alarm_feedback ingest tick", zap.Error(err))
			} else if n > 0 {
				e.logger.Info("alarm_feedback ingest tick", zap.Int("processed", n))
			}
		}
	}
}

// snapshotLoop 每 snapshotInterval 把所有 grid 状态 dump 到 persister。
// 单次 dump 失败只警告，不退出循环（DB 抖动不应拖垮 engine）。
func (e *Engine) snapshotLoop(ctx context.Context) {
	ticker := time.NewTicker(e.snapshotInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.saveAllRooms(ctx)
		}
	}
}

// saveAllRooms 遍历所有 grid，逐房间 UPSERT 到 persister。
// 调用方可来自 snapshotLoop（带 ctx）或 Run 退出（用 background ctx）。
func (e *Engine) saveAllRooms(ctx context.Context) {
	if e.persister == nil {
		return
	}
	e.mu.RLock()
	// 拷贝 (roomID, grid, hash) 三元组到本地切片，释放锁后再做 IO，避免锁内 DB 写
	type roomDump struct {
		id   string
		grid *RoomGrid
		hash string
	}
	dumps := make([]roomDump, 0, len(e.grids))
	for id, g := range e.grids {
		dumps = append(dumps, roomDump{id: id, grid: g, hash: e.layoutHashes[id]})
	}
	e.mu.RUnlock()

	saved, failed := 0, 0
	for _, d := range dumps {
		snap := EncodeSnapshot(d.grid)
		payload, cellCount, err := MarshalSnapshot(snap)
		if err != nil {
			e.logger.Warn("snapshot marshal failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		if err := e.persister.Save(ctx, d.id, d.hash, cellCount, payload); err != nil {
			e.logger.Warn("snapshot save failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		saved++
	}
	e.logger.Debug("snapshot batch done",
		zap.Int("saved", saved), zap.Int("failed", failed))
}

// dailySnapshotLoop 每天 dailySnapshotHour:dailySnapshotMinute (local) 归档 grid 状态到 history 表。
// 区别于 snapshotLoop（每 5min 写 live 表，仅最新一份）：history 表按日期保留 365 天用于 playback 历史起点。
func (e *Engine) dailySnapshotLoop(ctx context.Context) {
	for {
		next := nextDailyTriggerHM(time.Now(), e.dailySnapshotHour, e.dailySnapshotMinute)
		wait := time.Until(next)
		e.logger.Info("daily_snapshot_scheduled",
			zap.Time("next", next),
			zap.Duration("wait", wait),
			zap.Int("hour_local", e.dailySnapshotHour),
			zap.Int("minute_local", e.dailySnapshotMinute),
			zap.Int("retain_days", e.historyRetainDays),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		e.saveAllRoomsHistory(ctx, time.Now())
	}
}

// saveAllRoomsHistory 遍历所有 grid，逐房间 SaveDaily 到 history 表。
// snapshotDate 用 nowLocal 的日期部分；retainDays 写入完成后顺手清理。
func (e *Engine) saveAllRoomsHistory(ctx context.Context, nowLocal time.Time) {
	if e.historyPersister == nil {
		return
	}
	e.mu.RLock()
	type roomDump struct {
		id   string
		grid *RoomGrid
		hash string
	}
	dumps := make([]roomDump, 0, len(e.grids))
	for id, g := range e.grids {
		dumps = append(dumps, roomDump{id: id, grid: g, hash: e.layoutHashes[id]})
	}
	e.mu.RUnlock()

	saved, failed := 0, 0
	for _, d := range dumps {
		snap := EncodeSnapshot(d.grid)
		payload, cellCount, err := MarshalSnapshot(snap)
		if err != nil {
			e.logger.Warn("daily snapshot marshal failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		if err := e.historyPersister.SaveDaily(ctx, d.id, d.hash, nowLocal, cellCount, payload, e.historyRetainDays); err != nil {
			e.logger.Warn("daily snapshot save failed",
				zap.String("room_id", d.id), zap.Error(err))
			failed++
			continue
		}
		saved++
	}
	e.logger.Info("daily_snapshot_done",
		zap.Int("saved", saved),
		zap.Int("failed", failed),
		zap.String("snapshot_date", nowLocal.Format("2006-01-02")),
	)
}

// nextDailyTriggerHM 返回下一次 hour:minute 的 local 时间点；已过则取明天。
func nextDailyTriggerHM(now time.Time, hour, minute int) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// dailyLayoutReloadLoop PR-15：每天 dailyReloadHour:00 (local) 重读 rooms.layout_config。
//
// 用途：管理员通过前端 layout 编辑器修改房间布局后，无需重启 wisefido-sensor 即可生效。
// 每日触发时刻应避开管理员上班时间（默认 22:00 = 晚 10 点）。
//
// 流程：
//  1. 等到下次 dailyReloadHour:00:00 local
//  2. 扫所有已注册 room，对比 DB 中 layout_config 的 hash 与内存 layoutHashes
//  3. hash 变 → 重置该 room（替换 TrackManager + 清空 grid 学习状态）
//  4. 立即持久化（覆盖旧 snapshot）
//
// 不变 hash 的 room 跳过；DB 查不到（room 已删）的房间也跳过。
func (e *Engine) dailyLayoutReloadLoop(ctx context.Context) {
	for {
		next := nextDailyTrigger(time.Now(), e.dailyReloadHour)
		wait := time.Until(next)
		e.logger.Info("daily_layout_reload_scheduled",
			zap.Time("next", next),
			zap.Duration("wait", wait),
			zap.Int("hour_local", e.dailyReloadHour),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		e.runDailyLayoutReload(ctx)
	}
}

// nextDailyTrigger 返回下一次 hour:00:00 local 的时间点。
// 若当前已经过当天 hour 时刻，则取明天。
func nextDailyTrigger(now time.Time, hour int) time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// runDailyLayoutReload 扫 rooms 表 → hash diff → 重置变化 room。
// 调用方：dailyLayoutReloadLoop 触发时；不持锁进入。
func (e *Engine) runDailyLayoutReload(ctx context.Context) {
	if e.dailyReloadDB == nil {
		return
	}
	// v2 schema: layout 在 room_visual_layout 表（PK=spatial_prefix）；
	// rooms 表无 tenant_id/unit_id 列；tenant_id 由 room_id INET prefix /48 派生；
	// unit timezone 通过 unit /80 LPM contains room /88 取。
	rows, err := e.dailyReloadDB.QueryContext(ctx, `
		SELECT r.room_id::text,
		       r.room_name,
		       rvl.canvas::text AS layout_config,
		       COALESCE(u.timezone, '') AS timezone,
		       host(set_masklen(r.room_id, 48))::text || '/48' AS tenant_pref
		FROM rooms r
		JOIN room_visual_layout rvl ON rvl.spatial_prefix = r.room_id
		LEFT JOIN units u ON u.unit_id >>= r.room_id`)
	if err != nil {
		e.logger.Warn("daily_reload query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type pendingReload struct {
		cfg      RoomConfig
		tenantID string
		newHash  string
	}
	var pending []pendingReload
	for rows.Next() {
		var roomID, roomName, tenantID, timezone string
		var layoutStr sql.NullString
		if err := rows.Scan(&roomID, &roomName, &layoutStr, &timezone, &tenantID); err != nil {
			continue
		}
		if !layoutStr.Valid || layoutStr.String == "" {
			continue
		}
		cfg, err := ParseLayoutConfig(roomID, []byte(layoutStr.String))
		if err != nil {
			e.logger.Warn("daily_reload parse failed", zap.String("room_id", roomID), zap.Error(err))
			continue
		}
		cfg.RoomName = roomName
		cfg.Timezone = timezone
		ApplyOptimizedExtent(&cfg)
		newHash := LayoutHash(cfg)
		e.mu.RLock()
		oldHash := e.layoutHashes[roomID]
		e.mu.RUnlock()
		if oldHash == newHash {
			continue // 没变
		}
		pending = append(pending, pendingReload{cfg: cfg, tenantID: tenantID, newHash: newHash})
	}
	for _, pr := range pending {
		e.logger.Info("daily_reload room changed, resetting",
			zap.String("room_id", pr.cfg.RoomID),
			zap.String("new_hash", pr.newHash[:12]),
		)
		// RegisterRoom 替换 TrackManager + grid，hydrateRoom 见 hash 不同会丢弃旧 snapshot
		e.RegisterRoom(pr.cfg)
		e.SetRoomTenant(pr.cfg.RoomID, pr.tenantID)
	}
	if len(pending) > 0 && e.persister != nil {
		// 立刻持久化新状态，覆盖旧 snapshot
		e.saveAllRooms(ctx)
	}
}

// winnerEvalLoop 每 24 小时评估一次 winner（需要 feedback_loop 积累 accuracy 数据）
func (e *Engine) winnerEvalLoop(ctx context.Context) {
	ticker := time.NewTicker(e.winnerEvalInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.reevaluateWinner()
		}
	}
}

// reevaluateWinner 依据 accuracy[3] 选择新 winner。
// 规则：
//   - 若 3 组准确率都 < 样本阈值 → 不切换
//   - 若最高准确率 ≥ 雷达 baseline + 20% → 切到该组
//   - 否则维持当前 winner
//
// 实际 baseline（雷达直报准确率）需要 feedback_loop 单独统计，这里暂用 0.5 占位。
func (e *Engine) reevaluateWinner() {
	e.mu.Lock()
	defer e.mu.Unlock()

	const baselineAcc = 0.5 // 占位：二期从 alarm_events 统计雷达直报准确率

	best := -1
	bestAcc := -1.0
	for g := 0; g < 3; g++ {
		acc := e.accuracy[g].Accuracy()
		if acc < 0 {
			continue // 样本不足
		}
		if acc > bestAcc {
			best = g
			bestAcc = acc
		}
	}

	if best < 0 {
		e.logger.Debug("winner eval skipped: not enough samples")
		return
	}
	if bestAcc < baselineAcc+0.20 {
		e.logger.Info("winner eval: AI not beating baseline by 20%, keep current",
			zap.Float64("best_acc", bestAcc),
			zap.Float64("baseline", baselineAcc),
			zap.String("current_winner", e.paramSets[e.winner].Name),
		)
		return
	}
	if best == e.winner {
		return
	}
	e.logger.Info("winner switched",
		zap.String("from", e.paramSets[e.winner].Name),
		zap.String("to", e.paramSets[best].Name),
		zap.Float64("new_acc", bestAcc),
	)
	e.winner = best
}

// RecordGroundTruth 外部（feedback_loop.go）每收到一条家属反馈就调一次
// predicted: engine 在该 cell/track 上对三组的预测（如"是否 fall"），truthReal: 是否真 fall
func (e *Engine) RecordGroundTruth(predicted [3]bool, truthReal bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for g := 0; g < 3; g++ {
		switch {
		case predicted[g] && truthReal:
			e.accuracy[g].TruePositive++
		case predicted[g] && !truthReal:
			e.accuracy[g].FalsePositive++
		case !predicted[g] && truthReal:
			e.accuracy[g].FalseNegative++
		default:
			e.accuracy[g].TrueNegative++
		}
	}
}

// ========================================================================
// 类型转换 helper（保留原有）
// ========================================================================

func intFromAny(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}

func int64FromAny(v interface{}) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case json.Number:
		n, _ := x.Int64()
		return n
	}
	return 0
}
