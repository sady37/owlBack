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
	"time"

	"owl-common/card"
	owlredis "owl-common/redis"

	"wisefido-sensor/internal/consumer"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

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
		client:    client,
		logger:    logger,
		timeout:   2 * time.Second,
		tickEvery: 60 * time.Second,
	}
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
			if p.aggregator == nil {
				continue
			}
			p.tickPullAndPublish(ctx)
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

// OnZoneEvent satisfy ZoneEventListener。
// 输出消息 SubjectEntity = 物理实体地址（/88 room CIDR / /96 bed CIDR；e.ZoneID）。
// sensor 不读 card:state，不查 card 视图——cardagg projector 收到后按字段级 merge
// 保留非 sensor owner 字段（SleepStage / TrackNumber 等）。
func (p *StreamPublisher) OnZoneEvent(e ZoneEvent) {
	if e.ZoneID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	switch e.ZoneType {
	case ZoneTypeBed:
		bs := TranslateBedState(e)
		_ = p.PublishBedState(ctx, e.ZoneID, bs)
	case ZoneTypeRoom:
		rs := TranslateRoomState(e, card.RoomTypeDefault)
		_ = p.PublishRoomState(ctx, e.ZoneID, rs)
	case ZoneTypeBathroom:
		rs := TranslateRoomState(e, card.RoomTypeBathroom)
		_ = p.PublishRoomState(ctx, e.ZoneID, rs)
	}
}

// PublishBedState 发完整 card.BedState 到 sensor:derived:stream，category=bed.state。
func (p *StreamPublisher) PublishBedState(ctx context.Context, cardID string, bs *card.BedState) error {
	return p.publish(ctx, cardID, "bed.state", bs)
}

// PublishRoomState category=room.state（bathroom kind 也走这条，由 RoomState.Kind 区分）。
func (p *StreamPublisher) PublishRoomState(ctx context.Context, cardID string, rs *card.RoomState) error {
	return p.publish(ctx, cardID, "room.state", rs)
}

// PublishTargetState category=target.state。
// 接口先到位；sensor 内部 Target 派生（Pose/Visitor/WeakBio/LogicID）实现后调。
func (p *StreamPublisher) PublishTargetState(ctx context.Context, cardID string, ts *card.TargetState) error {
	return p.publish(ctx, cardID, "target.state", ts)
}

func (p *StreamPublisher) publish(ctx context.Context, cardID, category string, payload interface{}) error {
	if cardID == "" || payload == nil {
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
	addr := parseCardAddr(cardID)
	msg := owlredis.NewSingleItemMessage(addr, cardID, "sensor", time.Now().UnixMilli(), "derived", category, data)
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

// parseCardAddr cardID 现为 INET CIDR 文本（v2）；解析失败返 zero netip.Addr。
func parseCardAddr(cardID string) netip.Addr {
	if pfx, err := netip.ParsePrefix(cardID); err == nil {
		return pfx.Addr()
	}
	if a, err := netip.ParseAddr(cardID); err == nil {
		return a
	}
	return netip.Addr{}
}
