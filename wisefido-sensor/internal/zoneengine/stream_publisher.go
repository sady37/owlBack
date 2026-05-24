// stream_publisher.go — sensor 派生物理实体状态发到 sensor:derived:stream。
//
// 单向流原则（CLAUDE.md 规则 #1.3）：
//   - sensor 不读 card:state hash（那是 cardagg 维护的视图）
//   - sensor 只输出自己 owner 的 engine 字段，按物理实体地址（/88 room / /96 bed / /128 device）寻址
//   - 跨字段合并（保留 SleepStage 等非 sensor owner 字段）由 cardagg projector 字段级 merge 完成
//   - 跨实体合并（多 device target → 单 card target）由 cardagg TargetMerger 完成
//
// 3 个发送接口：
//   PublishBedState / PublishRoomState  — zoneengine.OnZoneEvent 自动调
//   PublishTargetState                  — sensor 内部 Target 累加器实现后调

package zoneengine

import (
	"context"
	"encoding/json"
	"net/netip"
	"sync"
	"time"

	"owl-common/card"
	owlredis "owl-common/redis"

	"wisefido-sensor/internal/consumer"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RoomCountReader publisher 读 raw Z（service.RadarRoomCountCache 实现）。
// ok=false 时退化到 e.NewState.Count（无 cache 或 stale）。
type RoomCountReader interface {
	GetZ(roomCIDR string) (int, bool)
}

// BedDedupQuery publisher 查 room 范围内 sleepad InBed && !radar InBed 的 bed 数。
// service.BedPresenceFusion 实现。
type BedDedupQuery interface {
	ExtraPeopleInRoom(roomCIDR string) int
}

// StreamPublisher 实现 ZoneEventListener 把 ZoneEvent 翻译并发 sensor:derived:stream。
//
// 现在两条触发路径（都走 publishCardState 单一出口，单 writer 不破）：
//
//	1) OnZoneEvent       — zoneengine state machine 触发 bed.state / room.state
//	2) Run 60s ticker    — 主动 pull TargetStateAggregator 拿 LastActiveTs / StandingMin /
//	                       WeakBioScore，per-entity 发 target.state（dirty 检查跳过无变化）
//
// 详 [[target_state_aggregator]] design doc：aggregator 是纯 state holder，
// publisher 是唯一发往 sensor:derived:stream 的 writer。
//
// envelope.Producer = sensor agent /128 IPv6（slot fd00:0:fff1::/48，per
// [[platform_agent_addressing]]）；通过 SetIdentity 注入。未注入时退化为 entity addr
// （兼容老 wiring，但 cardagg side 无法按 Producer "属 sensor" 路由）。
type StreamPublisher struct {
	client     *redislib.Client
	logger     *zap.Logger
	timeout    time.Duration
	aggregator AggregatorPuller // 可空：未注入时 ticker 路径退化为 no-op
	tickEvery  time.Duration
	producer   string // sensor agent IPv6 canonical 文本；SetIdentity 注入

	// RoomState dedup（[[bed_presence_fusion]]）：
	//   roomCount = service.RadarRoomCountCache（raw Z）
	//   bedDedup  = service.BedPresenceFusion（sleepad InBed && !radar InBed 计数）
	// 两者都 nil 时退化为 e.NewState.Count（兼容未 wire）。
	roomCount RoomCountReader
	bedDedup  BedDedupQuery

	// lastRoomState / lastBedState 缓存最近一次 publish 的 state（按 spatial CIDR）。
	// 用途：
	//   - bed event → republishRoomAfterBed 复用 RoomState 其它字段 + 更 TotalPeople
	//   - 60s tick → tickRepublishStates 兜底 re-publish（覆盖 missed-message / cardagg 重启场景）
	mu            sync.RWMutex
	lastRoomState map[string]*card.RoomState
	lastBedState  map[string]*card.BedState
}

// AggregatorPuller StreamPublisher 60s tick 时 pull aggregator 数据用的接口。
// 由 service.TargetStateAggregator 实现（multi-return 避免 service 反向导入 zoneengine）。
//
// 接口按"物理实体地址"（spatial_prefix INET CIDR）寻址，跟 sensor 内部表达保持一致；
// 同 v2 实现里 cardID 也是 spatial_prefix 字符串，cardagg projector 自然能写到 card:state Hash。
//
// 返回值:
//   target      Target.LastActiveTs / WeakBiometricSignal / Visitor* — 写到 card.target Hash
//   standingMin RoomState.StandingContinuousMin —— 写到 card.room_state Hash
//   dirty       自上次 MarkPublished 是否变化（false → publisher skip 该实体 publish）
//   ok          实体是否有 accumulator entry（false → 跳过）
type AggregatorPuller interface {
	ActiveSpatialPrefixes() []string
	GetSnapshot(spatialPrefix string) (target *card.TargetState, standingMin int, dirty bool, ok bool)
	MarkPublished(spatialPrefix string, tsMs int64)
}

func NewStreamPublisher(client *redislib.Client, logger *zap.Logger) *StreamPublisher {
	return &StreamPublisher{
		client:        client,
		logger:        logger,
		timeout:       2 * time.Second,
		tickEvery:     60 * time.Second,
		lastRoomState: make(map[string]*card.RoomState),
		lastBedState:  make(map[string]*card.BedState),
	}
}

// SetRoomDedup main wiring 注入 RoomState people count dedup 双源。两个都 nil 时退化为
// 直接用 engine NewState.Count（无 dedup；兼容未 wire 场景）。
func (p *StreamPublisher) SetRoomDedup(z RoomCountReader, dedup BedDedupQuery) {
	if p == nil {
		return
	}
	p.roomCount = z
	p.bedDedup = dedup
}

// SetAggregator 注入 aggregator（main wiring 调）。可选；不调时 Run ticker 跳过 pull。
func (p *StreamPublisher) SetAggregator(a AggregatorPuller) {
	p.aggregator = a
}

// SetIdentity 注入 sensor agent /128 IPv6（main wiring 调）。空 / 无效时 publish 时
// 退化为 entity addr — 不推荐生产用（cardagg 端 Producer 防 loop 会失效）。
func (p *StreamPublisher) SetIdentity(id consumer.AgentIdentity) {
	if id.IPv6.IsValid() {
		p.producer = id.IPv6.String()
	}
}

// Run 60s ticker 主循环：遍历 aggregator.ActiveCardIDs，dirty 卡走 publishMergedFromAggregator。
// ctx done 时退出；P2 scaffold 不发实际 publish（仅日志 stub），P3/P4 接上业务后启用。
func (p *StreamPublisher) Run(ctx context.Context) {
	if p.tickEvery <= 0 {
		p.tickEvery = 60 * time.Second
	}
	t := time.NewTicker(p.tickEvery)
	defer t.Stop()
	p.logger.Info("stream_publisher: tick loop started",
		zap.Duration("interval", p.tickEvery))
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("stream_publisher: tick loop stopped")
			return
		case <-t.C:
			p.tickRepublishStates(ctx)
			if p.aggregator != nil {
				p.tickPullAndPublish(ctx)
			}
		}
	}
}

// tickPullAndPublish 60s 周期被调；遍历 dirty 物理实体，pull aggregator + publish target.state。
//
// v2 实施：subject_entity = spatial_prefix（CIDR，/96 bed 或 /88 room；cardID == spatial_prefix）。
// 字段 owner（[[target_state_per_device]]）：
//   - LastActiveTs / WeakBiometricSignal — aggregator GetSnapshot.target 已填
//   - StandingContinuousMin              — GetSnapshot.standingMin 单返；这里合进 target struct
//
// Visitor 三字段不在 sensor 范围（cardagg VisitorDeriver 拥有；merger.mergeForCard 合并）。
//
// 失败不 MarkPublished — 下一 tick 继续重试（dirty 仍 true）。
func (p *StreamPublisher) tickPullAndPublish(ctx context.Context) {
	prefixes := p.aggregator.ActiveSpatialPrefixes()
	if len(prefixes) == 0 {
		return
	}
	dirty := 0
	for _, sp := range prefixes {
		target, standingMin, isDirty, ok := p.aggregator.GetSnapshot(sp)
		if !ok || !isDirty || target == nil {
			continue
		}
		target.StandingContinuousMin = standingMin
		if err := p.PublishTargetState(ctx, sp, target); err != nil {
			p.logger.Warn("stream_publisher: publish target.state failed",
				zap.String("spatial_prefix", sp), zap.Error(err))
			continue
		}
		p.aggregator.MarkPublished(sp, time.Now().UnixMilli())
		dirty++
	}
	if dirty > 0 {
		p.logger.Debug("stream_publisher: tick published target.state",
			zap.Int("count", dirty))
	}
}

// tickRepublishStates 60s 周期 re-publish lastRoomState / lastBedState 兜底。
//
// **Why**：OnZoneEvent 是 transition-only（state machine flip）；卡长时间 stay Occupied
// 无新 transition → 无新 publish → cardagg state hash 静默 stale（cardagg 重启 / 消息
// 丢包后无法恢复）。本 tick 周期 re-emit 最近一次 publish 过的 state（state-change-anchored
// 语义下：值/ts 不变 → cardagg merge no-op，仅保活 hash + display rebuild loop）。
//
// 对 RoomState 还顺带 re-apply dedup（roomCount.GetZ + bedDedup.ExtraPeopleInRoom 可能
// 在两次 transition 之间漂移）；BedState 无 dedup 直接重发。
//
// 与 [[maintained_stream_producer_owns_maintenance]] 一致：sensor 是 sensor:derived:stream
// 的 producer，TTL / dedup / snapshot 由 producer 守门，consumer 只读。
func (p *StreamPublisher) tickRepublishStates(ctx context.Context) {
	p.mu.RLock()
	rooms := make(map[string]*card.RoomState, len(p.lastRoomState))
	for k, v := range p.lastRoomState {
		cpy := *v
		rooms[k] = &cpy
	}
	beds := make(map[string]*card.BedState, len(p.lastBedState))
	for k, v := range p.lastBedState {
		cpy := *v
		beds[k] = &cpy
	}
	p.mu.RUnlock()

	roomCount := 0
	for cidr, rs := range rooms {
		p.applyRoomDedupInPlace(cidr, rs)
		if err := p.PublishRoomState(ctx, cidr, rs); err != nil {
			p.logger.Warn("stream_publisher: tick re-publish room.state failed",
				zap.String("room", cidr), zap.Error(err))
			continue
		}
		roomCount++
	}
	bedCount := 0
	for cidr, bs := range beds {
		if err := p.PublishBedState(ctx, cidr, bs); err != nil {
			p.logger.Warn("stream_publisher: tick re-publish bed.state failed",
				zap.String("bed", cidr), zap.Error(err))
			continue
		}
		bedCount++
	}
	if roomCount+bedCount > 0 {
		p.logger.Debug("stream_publisher: tick re-published states",
			zap.Int("rooms", roomCount), zap.Int("beds", bedCount))
	}
}

// OnZoneEvent satisfy ZoneEventListener。
// 输出消息 SubjectEntity = 物理实体地址（/88 room CIDR / /96 bed CIDR；e.ZoneID）。
// sensor 不读 card:state，不查 card 视图——cardagg projector 收到后按字段级 merge
// 保留非 sensor owner 字段（SleepStage / TrackNumber 等）。
//
// Bed event 后追加 room.state re-publish（[[bed_presence_fusion]]）：sleepad InBed
// flip 时 radar Z 未变 BedZone 也未触发 room transition，需 publisher 主动 re-publish
// 让 dedup 后的 TotalPeople 反映新 X，FE 才能看到正确人数。
func (p *StreamPublisher) OnZoneEvent(e ZoneEvent) {
	if e.ZoneID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	switch e.ZoneType {
	case ZoneTypeBed:
		bs := TranslateBedState(e)
		p.cacheLastBed(e.ZoneID, bs)
		_ = p.PublishBedState(ctx, e.ZoneID, bs)
		p.republishRoomAfterBed(ctx, e)
	case ZoneTypeRoom:
		rs := TranslateRoomState(e, card.RoomTypeDefault)
		p.applyRoomDedupInPlace(e.ZoneID, rs)
		p.cacheLastRoom(e.ZoneID, rs)
		_ = p.PublishRoomState(ctx, e.ZoneID, rs)
	case ZoneTypeBathroom:
		rs := TranslateRoomState(e, card.RoomTypeBathroom)
		p.applyRoomDedupInPlace(e.ZoneID, rs)
		p.cacheLastRoom(e.ZoneID, rs)
		_ = p.PublishRoomState(ctx, e.ZoneID, rs)
	}
}

// applyRoomDedupInPlace 把 raw Z + ExtraPeopleInRoom 覆写到 rs.TotalPeople。
// roomCount/bedDedup 未注入时退化为 no-op（保留 TranslateRoomState 给出的 e.NewState.Count）。
//
// **Vacant trust 守卫**：rs.TotalPeople == 0 表示 engine 刚跑 Vacant transition 或 count_change to 0
// （applyTransitionToState 强 set Count=0），这是 ground truth；此时**禁止用 cache.Z 覆写**——
// 否则 radar EnterRoom/ExitRoom 路径没 NumberPeople=0 心跳时，cache.Z 留 1 stale，让 bathroom
// 离场后永远卡 InBath（实测 f61e8o 6min stale）。
//
// 取数顺序（仅 TotalPeople > 0 时）：
//  1. Z = roomCount.GetZ()，ok 时用之；否则用 rs.TotalPeople（fallback，含 subset_invariant lift）
//  2. extras = bedDedup.ExtraPeopleInRoom()
//  3. rs.TotalPeople = Z + extras
func (p *StreamPublisher) applyRoomDedupInPlace(roomCIDR string, rs *card.RoomState) {
	if rs == nil || rs.TotalPeople == 0 {
		return
	}
	z := rs.TotalPeople
	if p.roomCount != nil {
		if cached, ok := p.roomCount.GetZ(roomCIDR); ok {
			z = cached
		}
	}
	extras := 0
	if p.bedDedup != nil {
		extras = p.bedDedup.ExtraPeopleInRoom(roomCIDR)
	}
	rs.TotalPeople = z + extras
}

// cacheLastRoom 保存最近一次 publish 的 RoomState（深拷贝），bed event 后 re-publish 用。
func (p *StreamPublisher) cacheLastRoom(roomCIDR string, rs *card.RoomState) {
	if roomCIDR == "" || rs == nil {
		return
	}
	cpy := *rs
	p.mu.Lock()
	p.lastRoomState[roomCIDR] = &cpy
	p.mu.Unlock()
}

// cacheLastBed 保存最近一次 publish 的 BedState（深拷贝），60s tick re-publish 用。
func (p *StreamPublisher) cacheLastBed(bedCIDR string, bs *card.BedState) {
	if bedCIDR == "" || bs == nil {
		return
	}
	cpy := *bs
	p.mu.Lock()
	p.lastBedState[bedCIDR] = &cpy
	p.mu.Unlock()
}

// republishRoomAfterBed bed event 后用最新 dedup 重发 room.state。
//
// 触发场景：sleepad InBed 0→1 但 radar Z 未变 → BedZone transition 不带 room transition
// → 仅 ZoneTypeBed OnZoneEvent fire；若不主动 republish，FE 看到的 TotalPeople 落后
// 直到下次 radar count_change（可能数分钟）。
//
// 复用 lastRoomState 缓存（含 Kind / LastEnterTime / RiskLevel 等），仅更 TotalPeople +
// UpdatedAt。无 lastRoomState 时退出（rooms 未被 publish 过 = 还没人进过 = sleepad 单独
// 触发也无意义，let subset_invariant 路径走 ZoneTypeRoom 正常 flow）。
func (p *StreamPublisher) republishRoomAfterBed(ctx context.Context, e ZoneEvent) {
	if p.roomCount == nil && p.bedDedup == nil {
		return
	}
	roomCIDR := deriveRoomZoneIDFromBed(e.ZoneID)
	if roomCIDR == "" {
		return
	}
	p.mu.RLock()
	prev, ok := p.lastRoomState[roomCIDR]
	p.mu.RUnlock()
	if !ok || prev == nil {
		return
	}
	rs := *prev
	// Re-apply dedup with current cache + fusion state；TotalPeopleTs 在值变化时刷新
	prevCount := prev.TotalPeople
	rs.TotalPeople = 0
	if p.roomCount != nil {
		if z, ok := p.roomCount.GetZ(roomCIDR); ok {
			rs.TotalPeople = z
		} else {
			rs.TotalPeople = prev.TotalPeople
		}
	} else {
		rs.TotalPeople = prev.TotalPeople
	}
	if p.bedDedup != nil {
		rs.TotalPeople += p.bedDedup.ExtraPeopleInRoom(roomCIDR)
	}
	if rs.TotalPeople == prevCount {
		return // 无变化跳过 publish 防风暴
	}
	// TotalPeople 真的变了 → 刷 TotalPeopleTs（state-change-anchored）
	rs.TotalPeopleTs = e.NewState.UpdatedAt
	// AloneSinceTs 锚：变到 1 时刷新；离开 1 时清 0
	if rs.TotalPeople == 1 && prevCount != 1 {
		rs.AloneSinceTs = e.NewState.UpdatedAt
	} else if rs.TotalPeople != 1 {
		rs.AloneSinceTs = 0
	}
	p.cacheLastRoom(roomCIDR, &rs)
	_ = p.PublishRoomState(ctx, roomCIDR, &rs)
}

// PublishBedState 发完整 card.BedState 到 sensor:derived:stream，category=bed.state。
func (p *StreamPublisher) PublishBedState(ctx context.Context, spatialPrefix string, bs *card.BedState) error {
	return p.publish(ctx, spatialPrefix, "bed.state", bs)
}

// PublishRoomState category=room.state（bathroom kind 也走这条，由 RoomState.Kind 区分）。
func (p *StreamPublisher) PublishRoomState(ctx context.Context, spatialPrefix string, rs *card.RoomState) error {
	return p.publish(ctx, spatialPrefix, "room.state", rs)
}

// PublishTargetState category=target.state。
// 接口先到位；sensor 内部 Target 派生（Pose/Visitor/WeakBio/LogicID）实现后调。
func (p *StreamPublisher) PublishTargetState(ctx context.Context, spatialPrefix string, ts *card.TargetState) error {
	return p.publish(ctx, spatialPrefix, "target.state", ts)
}

// PublishBedSleepStage category=bed.sleepstage — 仅写 SleepStage / SleepConfidence /
// UpdatedAt（projector 字段级 merge 不动其他 BedState owner 字段）。
//
// S4: SleepStage 由 sensor SleepStageConsumer confidence ladder 写出（Sleepad=90 / Radar=60）；
// 详 [[cardagg_v1_to_v2_migration_audit]]。
//
// **UpdatedAt 注入约定（2026-05-20 修订）**：sleepStage = SleepStageInitial(0) 表示"清空 prev"
// 占位调用（initial_publish 重启时用），不该把 UpdatedAt 当 fresh evidence 注入；否则
// cardagg `mergeBedStateSleepStage` max-merge 会把 bed_state.UpdatedAt 钉死在 startup 时刻 →
// FE active_anchor 全卡同步显示"Active Xm ago"假活跃。Initial 时 ts=0，让 cardagg
// 端 state-change merge 跳过（值不变 ts 保 prev）。
func (p *StreamPublisher) PublishBedSleepStage(ctx context.Context, spatialPrefix string, sleepStage, sleepConfidence int) error {
	if spatialPrefix == "" {
		return nil
	}
	bs := &card.BedState{
		SleepStage:      sleepStage,
		SleepConfidence: sleepConfidence,
	}
	// Initial(0) 不注入 ts — startup placeholder，避免假活跃 anchor
	if sleepStage != card.SleepStageInitial {
		now := time.Now().UnixMilli()
		bs.SleepStageTs = now
		bs.SleepConfidenceTs = now
	}
	return p.publish(ctx, spatialPrefix, "bed.sleepstage", bs)
}

func (p *StreamPublisher) publish(ctx context.Context, spatialPrefix, category string, payload interface{}) error {
	if spatialPrefix == "" || payload == nil {
		return nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	addr := parseSpatialAddr(spatialPrefix)
	msg := owlredis.NewSingleItemMessage(addr, spatialPrefix, "sensor", time.Now().UnixMilli(), "derived", category, data)
	// Producer 覆写为 sensor agent /128 IPv6（platform agent slot fd00:0:fff1::/48）。
	// NewSingleItemMessage 默认 producer=entity addr，cardagg side 看 Producer 防 loop / 路由会失效。
	if p.producer != "" {
		msg.Producer = p.producer
	}
	return p.client.XAdd(ctx, &redislib.XAddArgs{
		Stream: owlredis.StreamSensorDerived.Name,
		MaxLen: owlredis.StreamSensorDerived.MaxLen,
		Approx: true,
		Values: msg.ToStreamMap(),
	}).Err()
}

// parseSpatialAddr spatial prefix CIDR text → netip.Addr；解析失败返 zero netip.Addr。
func parseSpatialAddr(spatialPrefix string) netip.Addr {
	if pfx, err := netip.ParsePrefix(spatialPrefix); err == nil {
		return pfx.Addr()
	}
	if a, err := netip.ParseAddr(spatialPrefix); err == nil {
		return a
	}
	return netip.Addr{}
}
