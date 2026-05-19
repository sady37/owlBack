// stream_publisher.go — sensor 派生 per-card 状态发到 sensor:derived:stream。
//
// 替代旧 adapter_redis.go 直写 card:status hash 的方式（CLAUDE.md 规则 #1.3 单 writer：
// cardagg sensor_state_projector 是 card:status 唯一 writer，sensor 通过流推消息让 cardagg 写）。
//
// 3 个发送接口：
//   PublishBedState / PublishRoomState  — zoneengine.OnZoneEvent 自动调
//   PublishTargetState                  — 留接口，sensor 内部 Target 派生实现后调
//
// 设计：bathroom 不再独立流，bathroom zone 翻译为 RoomState 带 Kind=bathroom。
// sensor 读 card:status 取 prev 字段保留（TrackNumber/SleepStage/AreaPeople 等非 engine 字段），
// merge engine 翻译后的字段后发完整 JSON；cardagg projector blindly 写 hash。

package zoneengine

import (
	"context"
	"encoding/json"
	"net/netip"
	"time"

	"owl-common/card"
	owlredis "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// StreamPublisher 实现 ZoneEventListener 把 ZoneEvent 翻译并发 sensor:derived:stream。
type StreamPublisher struct {
	reader  *card.Reader
	client  *redislib.Client
	logger  *zap.Logger
	timeout time.Duration
}

func NewStreamPublisher(client *redislib.Client, logger *zap.Logger) *StreamPublisher {
	return &StreamPublisher{
		reader:  card.NewReader(client),
		client:  client,
		logger:  logger,
		timeout: 2 * time.Second,
	}
}

// OnZoneEvent satisfy ZoneEventListener — 替代 RedisAdapter 同名方法。
// 输出 cardID = e.ZoneID（physical /88 room / /96 bed），与 radar 的 SubjectEntity 解耦：
// 多 radar 同绑 /80 unit card 时 bedroom + bathroom 数据落各自 /88 hash 不互相覆盖。
// /80 unit 没有自己的 state 块；FE unit 卡按 /80 prefix 找子 /88，挑最近 UpdatedAt 渲染。
func (p *StreamPublisher) OnZoneEvent(e ZoneEvent) {
	if e.ZoneID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	cur, err := p.reader.ReadCardStatus(ctx, e.ZoneID)
	if err != nil {
		p.logger.Warn("read card status", zap.String("zone", e.ZoneID), zap.Error(err))
		cur = &card.CardStatus{CardID: e.ZoneID}
	}

	switch e.ZoneType {
	case ZoneTypeBed:
		bs := TranslateBedState(e, cur.BedState)
		_ = p.PublishBedState(ctx, e.ZoneID, bs)
	case ZoneTypeRoom:
		rs := TranslateRoomState(e, cur.RoomState, card.RoomTypeDefault)
		_ = p.PublishRoomState(ctx, e.ZoneID, rs)
	case ZoneTypeBathroom:
		rs := TranslateRoomState(e, cur.RoomState, card.RoomTypeBathroom)
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
