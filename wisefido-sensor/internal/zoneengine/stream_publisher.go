// stream_publisher.go — sensor 派生 per-card 状态发到 sensor:derived:stream。
//
// 替代旧 adapter_redis.go 直写 card:status hash 的方式（CLAUDE.md 规则 #1.3 单 writer：
// cardagg sensor_state_projector 是 card:status 唯一 writer，sensor 通过流推消息让 cardagg 写）。
//
// 4 个发送接口：
//   PublishBedState / PublishRoomState / PublishBathRoomState  — zoneengine.OnZoneEvent 自动调
//   PublishTargetState                                          — 留接口，sensor 内部 Target 派生
//                                                                 实现后调（当前为空）
//
// 设计：sensor 读 card:status 取 prev 字段保留（TrackNumber/SleepStage/AreaPeople/StayFSMPhase
// 等非 engine 字段），merge engine 翻译后的字段后发完整 JSON；cardagg projector blindly 写 hash。

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
func (p *StreamPublisher) OnZoneEvent(e ZoneEvent) {
	if e.CardID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	cur, err := p.reader.ReadCardStatus(ctx, e.CardID)
	if err != nil {
		p.logger.Warn("read card status", zap.String("cid", e.CardID), zap.Error(err))
		cur = &card.CardStatus{CardID: e.CardID}
	}

	switch e.ZoneType {
	case ZoneTypeBed:
		bs := TranslateBedState(e, cur.BedState)
		_ = p.PublishBedState(ctx, e.CardID, bs)
	case ZoneTypeRoom:
		rs := TranslateRoomState(e, cur.RoomState)
		_ = p.PublishRoomState(ctx, e.CardID, rs)
	case ZoneTypeBathroom:
		br := TranslateBathRoomState(e, cur.BathRoomState)
		_ = p.PublishBathRoomState(ctx, e.CardID, br)
	}
}

// PublishBedState 发完整 card.BedState 到 sensor:derived:stream，category=bed.state。
func (p *StreamPublisher) PublishBedState(ctx context.Context, cardID string, bs *card.BedState) error {
	return p.publish(ctx, cardID, "bed.state", bs)
}

// PublishRoomState category=room.state。
func (p *StreamPublisher) PublishRoomState(ctx context.Context, cardID string, rs *card.RoomState) error {
	return p.publish(ctx, cardID, "room.state", rs)
}

// PublishBathRoomState category=bathroom.state。
func (p *StreamPublisher) PublishBathRoomState(ctx context.Context, cardID string, br *card.BathRoomState) error {
	return p.publish(ctx, cardID, "bathroom.state", br)
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
