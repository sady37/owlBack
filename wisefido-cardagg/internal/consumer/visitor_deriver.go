// visitor_deriver.go — cardagg 端 visitor 累加器（双路径：bed-bound radar / Private room）。
//
// 设计参照 doc/card_display.md §4.4 + memory [[visitor_belongs_to_cardagg]]：
//
//   - 60s tick，不依赖事件流
//   - Path A（优先）：bed cards 且其 /96 prefix 下绑有 bed-bound radar；读 BedPeopleTracker
//     拿 per-card people count；firmware boundary 已物理裁剪视野到该床区域，count≥2 ≡ 床主+访客
//   - Path B（兜底）：父 unit_type=Private (1) 的 /88 room cards；读 card:state.room_state.total_people
//   - 优先级冲突：若某 /88 room card 下任一子 bed card 命中 Path A，本轮跳过该 room（避免父子双显）
//   - per-card segment 累加（≥2 推进；<2 重置）；segment 跨 5min 阈值即触发 visitor
//   - 通过 TargetMerger.ApplyVisitor 注入 visitor 三字段，最终由 SensorStateProjector 写 hash
//   - 本地午夜（按 unit timezone，当前用 UTC）reset today 三字段
//
// 物理实体 vs card 视图：visitor 是跨 device/room 的"组合人数"派生 → 属 card 层语义，
// sensor 不参与（[[visitor_belongs_to_cardagg]] §"Visitor 职责分工"）。
//
// **5min 阈值的物理意义**：进房 5min 才算 visitor，过滤"快闪进出"假阳；段断（人离开）
// 立刻清 segment，同一天再来重新累加；today 三字段保留至午夜。
//
// 时区：当前用 UTC 算午夜（unit timezone 配置后续补；详 doc/card_display.md §4.4 与
// memory [[server_internal_utc_only.md]]：UTC 内部 + TZ 仅 API 边界）。
package consumer

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"owl-common/card"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

const (
	visitorThresholdMin     = 5  // segment ≥ 5min 触发 visitor
	visitorDefaultTickEvery = 60 // tick interval seconds
)

// VisitorDeriver 60s tick 任务：bed level + room level 双路径计算 visitor。
type VisitorDeriver struct {
	metaCache  *service.DeviceMetaCache
	reader     *card.Reader
	merger     *service.TargetMerger
	bedPeople  *service.BedPeopleTracker
	logger     *zap.Logger

	interval time.Duration

	mu       sync.Mutex
	segments map[string]*visitorSegment // key = cardID (/88 room 或 /96 bed)
}

// visitorSegment per-card 段累积态。
type visitorSegment struct {
	cardID          string
	segmentStartTs  int64  // ms; 0 = 当前无 ongoing visitor segment
	segDurationMin  int    // 当前段累积分钟（≥2 持续命中累加）
	todayMaxMin     int    // 当日跨段最大
	hasToday        bool   // 当日是否曾达 5min 阈值
	visitorStartTs  int64  // 最近一次达阈值时的 segment_start_ts（写到 Target.VisitorStartTs）
	lastTickDateUTC string // YYYY-MM-DD（UTC）；跨日 reset
}

// NewVisitorDeriver 构造。interval=0 走默认 60s。
func NewVisitorDeriver(
	metaCache *service.DeviceMetaCache,
	reader *card.Reader,
	merger *service.TargetMerger,
	bedPeople *service.BedPeopleTracker,
	interval time.Duration,
	logger *zap.Logger,
) *VisitorDeriver {
	if interval <= 0 {
		interval = time.Duration(visitorDefaultTickEvery) * time.Second
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &VisitorDeriver{
		metaCache: metaCache,
		reader:    reader,
		merger:    merger,
		bedPeople: bedPeople,
		logger:    logger,
		interval:  interval,
		segments:  make(map[string]*visitorSegment),
	}
}

// Run 主循环；阻塞，直到 ctx done。
func (v *VisitorDeriver) Run(ctx context.Context) {
	if v == nil {
		return
	}
	t := time.NewTicker(v.interval)
	defer t.Stop()
	v.logger.Info("visitor_deriver: started", zap.Duration("interval", v.interval))
	defer v.logger.Info("visitor_deriver: stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			v.tick(ctx, now)
		}
	}
}

// tick 一轮：先跑 Path A（bed-bound radar bed cards），记录被覆盖的 parent room；
// 再跑 Path B（Private /88 rooms），跳过已被 Path A 覆盖的。
func (v *VisitorDeriver) tick(ctx context.Context, now time.Time) {
	nowMs := now.UnixMilli()
	dateUTC := now.UTC().Format("2006-01-02")

	// Path A — bed-bound radar bed cards (优先)
	bedCardIDs := v.metaCache.ListBedCardsWithBedBoundRadar(ctx)
	skipParents := make(map[string]struct{})
	for _, bedCardID := range bedCardIDs {
		peopleCount := 0
		if v.bedPeople != nil {
			peopleCount = v.bedPeople.CardPeopleCount(ctx, bedCardID)
		}
		v.tickCard(ctx, bedCardID, peopleCount, nowMs, dateUTC)
		if pr := parentRoomCardID(bedCardID); pr != "" {
			skipParents[pr] = struct{}{}
		}
	}

	// Path B — Private /88 room cards (兜底)
	roomCardIDs := v.metaCache.ListPrivateRoomCardIDs(ctx)
	for _, roomCardID := range roomCardIDs {
		if _, skip := skipParents[roomCardID]; skip {
			continue
		}
		peopleCount := v.readTotalPeople(ctx, roomCardID)
		v.tickCard(ctx, roomCardID, peopleCount, nowMs, dateUTC)
	}
}

// tickCard 处理单 card 的累加 + 阈值 + midnight reset。
func (v *VisitorDeriver) tickCard(ctx context.Context, cardID string, peopleCount int, nowMs int64, dateUTC string) {
	v.mu.Lock()
	seg, ok := v.segments[cardID]
	if !ok {
		seg = &visitorSegment{cardID: cardID, lastTickDateUTC: dateUTC}
		v.segments[cardID] = seg
	}

	// midnight reset (UTC)：跨日重置 today 三字段，segment 也清（新一天重新观察）
	if seg.lastTickDateUTC != "" && seg.lastTickDateUTC != dateUTC {
		seg.hasToday = false
		seg.todayMaxMin = 0
		seg.visitorStartTs = 0
		seg.segmentStartTs = 0
		seg.segDurationMin = 0
	}
	seg.lastTickDateUTC = dateUTC

	if peopleCount >= 2 {
		if seg.segmentStartTs == 0 {
			seg.segmentStartTs = nowMs
			seg.segDurationMin = 1
		} else {
			seg.segDurationMin++
		}
		// 跨 5min 阈值：触发 visitor 显示
		if seg.segDurationMin >= visitorThresholdMin {
			seg.visitorStartTs = seg.segmentStartTs
			seg.hasToday = true
			if seg.segDurationMin > seg.todayMaxMin {
				seg.todayMaxMin = seg.segDurationMin
			}
		}
	} else {
		// segment 断：reset segment but 保留 today 状态
		seg.segmentStartTs = 0
		seg.segDurationMin = 0
	}

	visitorStartTs := seg.visitorStartTs
	todayMax := seg.todayMaxMin
	hasToday := seg.hasToday
	v.mu.Unlock()

	// 注入 TargetMerger（visitor 字段下次 target.state 触发时合到 hash）
	if v.merger != nil {
		v.merger.ApplyVisitor(ctx, cardID, service.MakeVisitorFields(visitorStartTs, todayMax, hasToday))
	}
}

// readTotalPeople 从 card:state hash 读 room_state.total_people（值不存在或解析失败返 0）。
func (v *VisitorDeriver) readTotalPeople(ctx context.Context, cardID string) int {
	if v.reader == nil {
		return 0
	}
	status, err := v.reader.ReadCardStatus(ctx, cardID)
	if err != nil || status == nil || status.RoomState == nil {
		return 0
	}
	return status.RoomState.TotalPeople
}

// parentRoomCardID 由 /96 bed cardID 派生 /88 父 room cardID。非 /96 输入返空。
func parentRoomCardID(bedCardID string) string {
	pref, err := netip.ParsePrefix(bedCardID)
	if err != nil || pref.Bits() != 96 {
		return ""
	}
	return netip.PrefixFrom(pref.Addr(), 88).Masked().String()
}
