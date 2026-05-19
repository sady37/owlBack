// suite_census.go — SuiteBedroomCensus + SuitePerson 数据结构与 anchor 升格逻辑
//
// sensor_v2 PR-2（详 owlBack/doc/sensor_v2.md §4.A.1 + §10.1.2）
//
// 设计原则（决定 11+12+13+15）：
//   - v2 elder care suite 拓扑 = 1 bedroom（含 sleepad + radar）+ 1 bathroom（仅 radar）
//   - **只在 bedroom 内识别 SuitePerson**（bathroom 不可靠）
//   - SuiteID = unit spatial_prefix INET 字符串（如 "fd00:0:3:111::/80"）
//     **不**用 bedroom RoomID（review 反馈：v3 unit-layout 多 room 时无需重构 SuiteID 语义）
//   - Resident 升格：sleepad 双源锚定 OR radar 5min anchor + traverse ≥ 10 cells
//   - Visitor 升格：radar 2min anchor + traverse ≥ 5 cells（决定 13 时序假设：护工进 suite 不会马上冲 bathroom）
//   - Bathroom 跨入口流量更新留 PR-5 BathroomGate（本文件仅 bedroom 内升格）
//
// 持久化语义（review Doc-2 明确）：
//   Redis 持久化**仅用于快重启恢复 + 调试**（< 5min 重启窗口期不裸奔重做识别）。
//   **不是 HA 兜底**（决定明确 v2 不做 HA pending）。
//   TTL 24h 用于覆盖长维护窗口；如未来明确不需要超过 1-2h，可缩短。
//
// 多 resident 场景（review Doc-3）：
//   PR-2 UpdatePersonFromTrack 的 residentID 参数由 **caller 责任决定映射**：
//     - 床位绑定 resident（layout 配置）→ caller 知道 sleepad device → resident.id
//     - 单 resident unit → 直接传 unit.residents[0]
//     - PR-2 范围内多 resident bedroom 不实现自动消歧（v3 加 per-bed sleepad 映射）

package roomengine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"owl-common/card"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// SuitePerson 单个 suite 内识别出的人（resident 或 visitor）。
type SuitePerson struct {
	PersonID            string `json:"person_id"`             // resident.id (UUID/INET) 或 "visitor_<unix_ms>"
	Role                string `json:"role"`                  // "resident" / "visitor"
	AnchorTrackID       int    `json:"anchor_track_id"`       // 当前 anchor 的 radar track id（bedroom 内）
	AnchorRoomType      int    `json:"anchor_room_type"`      // owl-common/card.RoomType* — 当前 anchor 所在 room 的类型
	AnchorSinceMs       int64  `json:"anchor_since_ms"`       // anchor 起点 unix ms
	LastActiveMs        int64  `json:"last_active_ms"`        // 最近"非静止"帧 ms；决定 8 lost 判据
	LastBathroomEntryMs int64  `json:"last_bathroom_entry_ms,omitempty"`
	SleepadAnchored     bool   `json:"sleepad_anchored"`      // 床压双源锚定（仅 resident 可能为 true）
	TraverseCount       int    `json:"traverse_count"`        // 累计 traverse cells（升格判据）
}

// SuitePersonRole 取值
const (
	SuitePersonResident = "resident"
	SuitePersonVisitor  = "visitor"
)

// SuiteBedroomCensus 单 suite 的 census 状态。
//
// 一个 SuiteBedroomCensus 实例对应一个 unit（即一个 elder care suite）。
// SuiteID = unit spatial_prefix INET 字符串（如 "fd00:0:3:111::/80"）。
//
// IsPublicBathroom == true 时（决定 24 / §4.A.6）：
//   - Persons 不识别身份（public bathroom 没有附属 bedroom 做 census）
//   - BathroomCount 来自 bathroom 内 disambiguate 真人 track 数（不来自入口流量）
//   - SuiteID = bathroom 自己的 spatial_prefix（无附属 bedroom）
type SuiteBedroomCensus struct {
	SuiteID          string                  `json:"suite_id"`
	IsPublicBathroom bool                    `json:"is_public_bathroom,omitempty"`
	Persons          map[string]*SuitePerson `json:"persons"`
	BathroomCount    int                     `json:"bathroom_count"` // bathroom 内当前人数（入口流量累积；public mode 时为内部 track 数）
	UpdatedAtMs      int64                   `json:"updated_at_ms"`
}

// SuiteCensusConfig 参数（每 suite 配置；可由 spatial_config / cardagg 传入）。
type SuiteCensusConfig struct {
	// Resident 升格阈值
	ResidentRadarAnchorMs   int64 // radar anchor 时长，默认 5min
	ResidentRadarTraverseMin int   // radar traverse cells 最小值，默认 10

	// Visitor 升格阈值（决定 13 时序假设：≥2min）
	VisitorAnchorMs   int64 // 默认 2min
	VisitorTraverseMin int   // 默认 5

	// Visitor 衰退（决定 24 注解：离开 suite ≥ 10min 后移出 census）
	VisitorIdleRemoveMs int64 // 默认 10min

	// Resident 衰退（决定 8：全 unit 无活动 ≥ 60min → stale；≥ 6h → 移出）
	ResidentStaleMs   int64
	ResidentRemoveMs  int64
}

// DefaultSuiteCensusConfig 默认参数（sensor_v2 文档 §4.A.1 + 决定 13）。
func DefaultSuiteCensusConfig() SuiteCensusConfig {
	return SuiteCensusConfig{
		ResidentRadarAnchorMs:    5 * 60 * 1000,
		ResidentRadarTraverseMin: 10,
		VisitorAnchorMs:          2 * 60 * 1000,
		VisitorTraverseMin:       5,
		VisitorIdleRemoveMs:      10 * 60 * 1000,
		ResidentStaleMs:          60 * 60 * 1000,
		ResidentRemoveMs:         6 * 60 * 60 * 1000,
	}
}

// SuiteCensusManager 进程内 census 管理器，按 unit_spatial_prefix 分桶。
//
// 线程安全：内部 mu 保护 buckets；redisClient 自身 thread-safe。
type SuiteCensusManager struct {
	mu          sync.RWMutex
	buckets     map[string]*SuiteBedroomCensus // key = SuiteID
	config      SuiteCensusConfig
	redisClient *redis.Client
	redisKeyFmt string // 默认 "sensor:suite_census:%s"
	logger      *zap.Logger
}

// NewSuiteCensusManager 构造。
//
// 参数：
//   - redisClient: 可为 nil（纯内存模式，单测用）；prod 传 owl-common/redis.Client
//   - cfg: 升格 / 衰退 阈值；零值会被替换为 DefaultSuiteCensusConfig()
//   - logger: 可为 nil（无 log）；prod 传 owl-common/logger.NewLogger 输出 zap structured JSON
func NewSuiteCensusManager(redisClient *redis.Client, cfg SuiteCensusConfig, logger *zap.Logger) *SuiteCensusManager {
	if cfg.ResidentRadarAnchorMs == 0 {
		cfg = DefaultSuiteCensusConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SuiteCensusManager{
		buckets:     make(map[string]*SuiteBedroomCensus),
		config:      cfg,
		redisClient: redisClient,
		redisKeyFmt: "sensor:suite_census:%s",
		logger:      logger,
	}
}

// GetOrCreate 取或创建一个 suite 的 census。
func (m *SuiteCensusManager) GetOrCreate(suiteID string) *SuiteBedroomCensus {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.buckets[suiteID]
	if !ok {
		c = &SuiteBedroomCensus{
			SuiteID: suiteID,
			Persons: make(map[string]*SuitePerson),
		}
		m.buckets[suiteID] = c
	}
	return c
}

// Get 仅查询不创建。
func (m *SuiteCensusManager) Get(suiteID string) *SuiteBedroomCensus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.buckets[suiteID]
}

// MarkPublicBathroom 把 census 标记为 public bathroom standalone mode（决定 24 / §4.A.6）。
// public mode 下不识别 person 身份，BathroomCount 来自内部 track 数。
func (m *SuiteCensusManager) MarkPublicBathroom(suiteID string) {
	c := m.GetOrCreate(suiteID)
	m.mu.Lock()
	c.IsPublicBathroom = true
	m.mu.Unlock()
}

// UpdatePersonFromTrack 在 bedroom 内观测到 track 后调用。
//
// 业务规则（决定 11+13）：
//   - 只在 bedroom 调（caller = roomengine for bedroom rooms only）
//   - sleepadInBed 为 true 表示该 track 与床压传感器**当前**双源锚定（resident 强证据）
//     语义：实时状态，不是历史标记。resident 离床时传 false → SleepadAnchored 同步翻 false
//   - traverseDelta 是本次 update 累计的 traverse cells 增量
//   - moveActive 表示本帧是否"非静止"（用于 LastActiveMs 更新）
//
// 返回值语义（review Doc-4 完整文档化）：
//   - (person, true):
//       本次调用导致 candidate → resident/visitor 升格；
//       或 sleepad 强升格命中已存在 person；
//       或更新了已升格 person 的 traverse / LastActiveMs / SleepadAnchored。
//       caller 可消费 person.PersonID 写入下游（TrackStatus.PersonID 等）。
//
//   - (nil, false):
//       仍是 candidate 未升格 / public mode 不识别 / suiteID 不存在；
//       **caller 不应**使用任何 PersonID 关联下游 — candidate 阶段 PersonID 是临时 key
//       （格式 "__cand_track_<id>"），写入下游会污染 TrackStatus / alarm_events。
//
// caller 检查惯例：
//   if person, ok := mgr.UpdatePersonFromTrack(...); ok {
//       // 安全使用 person.PersonID / Role / ...
//   }
//   // ok=false 时跳过 PersonID 关联，仅保留 raw track frames
func (m *SuiteCensusManager) UpdatePersonFromTrack(
	suiteID, residentID string,
	trackID int,
	sleepadInBed bool,
	traverseDelta int,
	moveActive bool,
	nowMs int64,
) (*SuitePerson, bool) {
	c := m.GetOrCreate(suiteID)
	m.mu.Lock()
	defer m.mu.Unlock()

	if c.IsPublicBathroom {
		return nil, false
	}

	// 1) Sleepad 双源锚定 → 立即 resident 强升格（最高优先级）
	if sleepadInBed && residentID != "" {
		// 清掉同 trackID 的 candidate（如果有）
		delete(c.Persons, candidateKeyForTrack(trackID))
		if existing, ok := c.Persons[residentID]; ok {
			existing.AnchorTrackID = trackID
			existing.SleepadAnchored = true
			existing.AnchorRoomType = card.RoomTypeDefault
			existing.LastActiveMs = nowMs
			existing.TraverseCount += traverseDelta
			c.UpdatedAtMs = nowMs
			return existing, true
		}
		p := &SuitePerson{
			PersonID:        residentID,
			Role:            SuitePersonResident,
			AnchorTrackID:   trackID,
			AnchorRoomType:  card.RoomTypeDefault,
			AnchorSinceMs:   nowMs,
			LastActiveMs:    nowMs,
			SleepadAnchored: true,
			TraverseCount:   traverseDelta,
		}
		c.Persons[residentID] = p
		c.UpdatedAtMs = nowMs
		return p, true
	}

	// 2) 查找已升格 person 绑定该 trackID
	for _, p := range c.Persons {
		if p.Role != "" && p.AnchorTrackID == trackID && p.AnchorRoomType == card.RoomTypeDefault {
			p.TraverseCount += traverseDelta
			if moveActive {
				p.LastActiveMs = nowMs
			}
			// SleepadAnchored 实时同步语义（review 反馈）：
			// 表示"床压当前是否锚定中"，不是"曾经被锚定过"。
			// resident 离床时 sleepadInBed=false → SleepadAnchored=false；
			// 重新上床时 sleepadInBed=true → SleepadAnchored=true。
			// visitor 永远不可能 sleepadInBed=true（caller 责任：visitor 不传 sleepadInBed=true）。
			if p.Role == SuitePersonResident {
				p.SleepadAnchored = sleepadInBed
			}
			c.UpdatedAtMs = nowMs
			return p, true
		}
	}

	// 3) Candidate 处理（未升格）
	candidateKey := candidateKeyForTrack(trackID)
	cand, ok := c.Persons[candidateKey]
	if !ok {
		cand = &SuitePerson{
			PersonID:       candidateKey,
			Role:           "", // 待升格
			AnchorTrackID:  trackID,
			AnchorRoomType: card.RoomTypeDefault,
			AnchorSinceMs:  nowMs,
			LastActiveMs:   nowMs,
			TraverseCount:  traverseDelta,
		}
		c.Persons[candidateKey] = cand
		c.UpdatedAtMs = nowMs
		return nil, false
	}
	cand.TraverseCount += traverseDelta
	if moveActive {
		cand.LastActiveMs = nowMs
	}
	c.UpdatedAtMs = nowMs

	// 检查升格
	ageMs := nowMs - cand.AnchorSinceMs

	// Resident radar 升格
	if residentID != "" && ageMs >= m.config.ResidentRadarAnchorMs && cand.TraverseCount >= m.config.ResidentRadarTraverseMin {
		delete(c.Persons, candidateKey)
		cand.PersonID = residentID
		cand.Role = SuitePersonResident
		c.Persons[residentID] = cand
		return cand, true
	}

	// Visitor 升格（residentID == "" 才走，避免提前误升预期 resident）
	if residentID == "" && ageMs >= m.config.VisitorAnchorMs && cand.TraverseCount >= m.config.VisitorTraverseMin {
		delete(c.Persons, candidateKey)
		visitorID := fmt.Sprintf("visitor_%d", cand.AnchorSinceMs)
		cand.PersonID = visitorID
		cand.Role = SuitePersonVisitor
		c.Persons[visitorID] = cand
		return cand, true
	}

	return nil, false
}

// candidateKeyForTrack 临时 candidate ID（trackID 维度，避免与 resident.id / visitor_ID 冲突）。
func candidateKeyForTrack(trackID int) string {
	return fmt.Sprintf("__cand_track_%d", trackID)
}

// IsCandidate 判断 PersonID 是否是未升格的临时 candidate。
func IsCandidate(personID string) bool {
	return len(personID) >= 13 && personID[:13] == "__cand_track_"
}

// MarkActiveCrossZone 跨 zone LastActiveMs 更新（review 反馈：bathroom roomengine 也要喂）。
//
// 用途：resident 在 bathroom 静止淋浴 30 分钟时，bedroom 内无 track 调用 UpdatePersonFromTrack，
// 但 LastActiveMs 不能因此被冻结——否则 6h resident 衰退会在 bathroom 内异常触发。
//
// caller 责任（PR-5/7 wire）：
//   - bathroom roomengine 每帧检测到 person 对应 track 非静止时调用
//   - bathroom roomengine 不识别 person，所以传入的 personID 由 BathroomGate 入口流量记录的关联得来
//
// 仅更新 LastActiveMs，不动 anchor / sleepad / room kind 字段。
func (m *SuiteCensusManager) MarkActiveCrossZone(suiteID, personID string, nowMs int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.buckets[suiteID]
	if !ok {
		return false
	}
	p, ok := c.Persons[personID]
	if !ok {
		return false
	}
	p.LastActiveMs = nowMs
	c.UpdatedAtMs = nowMs
	return true
}

// ClearAnchorOnLostTrack track 失锁时清空 SuitePerson.AnchorTrackID（review Doc-5）。
//
// 用途：firmware track_id 复用（Q-A1 方案 2 瞬跳检测之后还有这种 case）。
// 老 track 失锁后 AnchorTrackID 清 0；之后即使 firmware 用同 trackID 复用，新 frames 走
// candidate 路径重新升格判定。
//
// **PR-3 wire 失锁判定建议**（review 反馈：偏保守避免 firmware coast 阶段误清）：
//   - 触发源：roomengine processFrameAt 段 2（未观测到 track）
//   - 阈值：连续 **≥ 60s 未收到该 trackID 的 PushPoint** 才调本函数清 AnchorTrackID
//   - 不要 30s — firmware coast 期间 track 可能维持发送（位置不变），coast 结束后又活；
//     30s 过早清会导致 resident 反复降级 / 重新走 candidate 路径
//   - 60s 仍然远早于 6h resident decay，对真正 lost 场景延迟可接受
//
// 注意：本函数仅清 AnchorTrackID，不删 person 也不动 LastActiveMs。
// 后续 person 是否衰退由 DecayInactive 按 idle 时长决定（与 track 失锁解耦）。
func (m *SuiteCensusManager) ClearAnchorOnLostTrack(suiteID string, trackID int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.buckets[suiteID]
	if !ok {
		return false
	}
	for _, p := range c.Persons {
		if p.AnchorTrackID == trackID {
			p.AnchorTrackID = 0
			// 注意：不动 Role / PersonID / LastActiveMs / SleepadAnchored
			// — person 仍然"存在"，只是没绑定到任何 radar track；
			// 新 track 来时通过 candidate 路径重新关联。
			return true
		}
	}
	return false
}

// MarkPersonExitToBathroom 标记某 SuitePerson 跨 BathroomGate 进入 bathroom（PR-5 调用）。
func (m *SuiteCensusManager) MarkPersonExitToBathroom(suiteID, personID string, nowMs int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.buckets[suiteID]
	if !ok {
		return false
	}
	p, ok := c.Persons[personID]
	if !ok {
		return false
	}
	p.AnchorRoomType = card.RoomTypeBathroom
	p.LastBathroomEntryMs = nowMs
	c.BathroomCount++
	c.UpdatedAtMs = nowMs
	return true
}

// MarkPersonReturnToBedroom 标记某 SuitePerson 跨 BathroomGate 返回 bedroom（PR-5 调用）。
func (m *SuiteCensusManager) MarkPersonReturnToBedroom(suiteID, personID string, nowMs int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.buckets[suiteID]
	if !ok {
		return false
	}
	p, ok := c.Persons[personID]
	if !ok {
		return false
	}
	p.AnchorRoomType = card.RoomTypeDefault
	if c.BathroomCount > 0 {
		c.BathroomCount--
	}
	c.UpdatedAtMs = nowMs
	return true
}

// GetBathroomCount 读 census.BathroomCount（PR-6 BathroomGhost 规则 2/3 用）。
// 返回 -1 sentinel：suite 不存在或 IsPublicBathroom（caller 据此跳过规则 2/3）。
func (m *SuiteCensusManager) GetBathroomCount(suiteID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.buckets[suiteID]
	if !ok || c.IsPublicBathroom {
		return -1
	}
	return c.BathroomCount
}

// AdjustBathroomCount 由 BathroomGate (PR-5) 调用的 raw count adjuster：
// 适配"匿名 track 进/出 bathroom"路径（track 无关联 SuitePerson 时）。
//
// 与 MarkPersonExitToBathroom/Return 的关系：
//   - 已升格 SuitePerson 进/出 bathroom：上层调 MarkPersonExitToBathroom/Return（PR-2 helper，含 count++ + anchor flip）
//   - 匿名 track（candidate / visitor 未升格 / anchor 已 clear）：上层调本函数仅调整 count
//   BathroomGate 当前对 bathroom 内 track 一律走"匿名"路径（v2 PR-3 决策：bathroom 不跑
//   UpdatePersonFromTrack），anchor flip 由 TryFlipSoleResidentRoomType 在 count 跨 0 边界时单独触发。
//
// 返回值 (oldCount, newCount)：caller 据此判断"是否跨 0 边界"决定是否触发 anchor flip。
// suiteID 不存在或 IsPublicBathroom 时返回 (-1, -1) 表示 noop（caller 自行判定不联动 anchor）。
// delta 任意整数（含 0 / 负）；newCount floor 在 0（不允许负计数）。
func (m *SuiteCensusManager) AdjustBathroomCount(suiteID string, delta int, nowMs int64) (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.buckets[suiteID]
	if !ok {
		return -1, -1
	}
	if c.IsPublicBathroom {
		// Public mode 不走入口流量路径（决定 24 / §4.A.6）；count 由 §4.A.3 在线 disambiguate 维护
		return -1, -1
	}
	old := c.BathroomCount
	newCount := old + delta
	if newCount < 0 {
		newCount = 0
	}
	c.BathroomCount = newCount
	c.UpdatedAtMs = nowMs
	return old, newCount
}

// TryFlipSoleResidentRoomType 单 resident 场景下的 anchor 翻转（PR-5 §4.A.2 + PR-3 multi-resident TODO）。
//
// 行为：
//   - 仅当 suite 内有**恰好 1 个** Role==resident 且 AnchorRoomType != targetRoomType 的 person 时翻转
//   - 0 个 / 多个 / 已在 targetRoomType → noop，返回 false
//
// 用途：BathroomGate 在 BathroomCount 跨 0 边界时调用——
//   - 0 → 1：targetRoomType = card.RoomTypeBathroom（首人进入）
//   - 1 → 0：targetRoomType = card.RoomTypeDefault（最后离开，回主居住区）
//
// 多 resident 场景（决定 19 留 PR-X）下本函数返回 false，count 仍正确累加，仅 anchor flip 跳过。
// 调用方应记 audit log 反馈"多 resident 模糊场景"，未来据此触发 sleepad → resident 静态映射改造。
func (m *SuiteCensusManager) TryFlipSoleResidentRoomType(suiteID string, targetRoomType int, nowMs int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.buckets[suiteID]
	if !ok || c.IsPublicBathroom {
		return false
	}
	var soleResident *SuitePerson
	residentCount := 0
	for _, p := range c.Persons {
		if p.Role != SuitePersonResident {
			continue
		}
		residentCount++
		soleResident = p
		if residentCount > 1 {
			return false
		}
	}
	if residentCount != 1 {
		return false
	}
	if soleResident.AnchorRoomType == targetRoomType {
		return false
	}
	soleResident.AnchorRoomType = targetRoomType
	if targetRoomType == card.RoomTypeBathroom {
		soleResident.LastBathroomEntryMs = nowMs
	}
	c.UpdatedAtMs = nowMs
	return true
}

// DecayInactive 清理过期 person（visitor IdleRemoveMs / resident RemoveMs / candidate 防 ghost 累积）。
// 调用方周期触发（建议 60s 一次）。
//
// candidate 衰退率 vs 升格率是 visitor anchor 阈值校准的反向信号（review 反馈）：
// 衰退率高 → 阈值（2min/5cells）可能过严；升格率高但 prod 看到误升 → 阈值可能过宽。
//
// **告警路由策略**（review 反馈：暂不进 alarm stream）：
//   - visitor/candidate decay → 仅 ai.log Info/Debug 级（运营/调试用）
//   - **resident decay → 仅 ai.log Warn 级**，**暂不**进 iot:alarm:stream
//     理由：跨界判定 — resident 6h 不动几乎肯定是设备健康问题（sleepad 故障 / radar 失联），
//     不是 fall 告警语义。更适合走系统健康检查链路（§6.A.0 决定 9）。
//   - **prod 验证策略**：先跑 1 周看频率
//       - 若罕见（每月几次）→ 升级为系统健康告警（producer=sensor，category=device）
//       - 若频繁（每天多次）→ 6h 阈值可能过严或衰退判据本身有问题，调阈值不发告警
func (m *SuiteCensusManager) DecayInactive(nowMs int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for suiteID, c := range m.buckets {
		for id, p := range c.Persons {
			idleMs := nowMs - p.LastActiveMs
			if p.Role == SuitePersonVisitor && idleMs > m.config.VisitorIdleRemoveMs {
				m.logger.Info("suite_census_visitor_decayed",
					zap.String("suite_id", suiteID),
					zap.String("person_id", id),
					zap.Int("track_id", p.AnchorTrackID),
					zap.Int64("idle_ms", idleMs),
					zap.Int64("ts_ms", nowMs),
				)
				delete(c.Persons, id)
				c.UpdatedAtMs = nowMs
				continue
			}
			if p.Role == SuitePersonResident && idleMs > m.config.ResidentRemoveMs {
				m.logger.Warn("suite_census_resident_decayed",
					zap.String("suite_id", suiteID),
					zap.String("person_id", id),
					zap.Int("track_id", p.AnchorTrackID),
					zap.Int64("idle_ms", idleMs),
					zap.Bool("sleepad_anchored", p.SleepadAnchored),
					zap.Int64("ts_ms", nowMs),
				)
				delete(c.Persons, id)
				c.UpdatedAtMs = nowMs
				continue
			}
			// candidate 长期不活动也清理（不到升格门槛但又静止 → 大概率 ghost）
			if p.Role == "" && idleMs > m.config.VisitorAnchorMs*2 {
				m.logger.Debug("suite_census_candidate_decayed",
					zap.String("suite_id", suiteID),
					zap.Int("track_id", p.AnchorTrackID),
					zap.Int64("age_ms", nowMs-p.AnchorSinceMs),
					zap.Int("traverse_count", p.TraverseCount),
					zap.Int64("idle_ms", idleMs),
					zap.Int64("ts_ms", nowMs),
				)
				delete(c.Persons, id)
				c.UpdatedAtMs = nowMs
			}
		}
	}
}

// SaveToRedis 序列化所有 suite census 到 Redis（key = sensor:suite_census:<suite_id>）。
// 调用方周期触发（建议 30s 一次或 census 变化即触发）。
func (m *SuiteCensusManager) SaveToRedis(ctx context.Context) error {
	if m.redisClient == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for suiteID, c := range m.buckets {
		key := fmt.Sprintf(m.redisKeyFmt, suiteID)
		data, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("marshal census %s: %w", suiteID, err)
		}
		// 24h TTL（足以 cover 维护窗口；之后会被新数据覆盖）
		if err := m.redisClient.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("redis set %s: %w", key, err)
		}
	}
	return nil
}

// LoadFromRedis 启动时从 Redis 恢复 census（决定 9 系统健康检查依赖：census 在过去 1h 内被更新过）。
func (m *SuiteCensusManager) LoadFromRedis(ctx context.Context, suiteIDs []string) error {
	if m.redisClient == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, suiteID := range suiteIDs {
		key := fmt.Sprintf(m.redisKeyFmt, suiteID)
		data, err := m.redisClient.Get(ctx, key).Bytes()
		if err == redis.Nil {
			continue // 该 suite 历史无 census，正常
		}
		if err != nil {
			return fmt.Errorf("redis get %s: %w", key, err)
		}
		var c SuiteBedroomCensus
		if err := json.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("unmarshal census %s: %w", suiteID, err)
		}
		if c.Persons == nil {
			c.Persons = make(map[string]*SuitePerson)
		}
		m.buckets[suiteID] = &c
	}
	return nil
}
