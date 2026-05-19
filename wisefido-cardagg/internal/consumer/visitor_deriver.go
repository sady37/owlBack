// visitor_deriver.go — cardagg 端 visitor 累加器（per /88 room card）。
//
// 设计参照 doc/card_display.md §4.4 + memory [[visitor_belongs_to_cardagg]]：
//
//   - 60s tick（不依赖任何事件流），遍历父 unit 是 Private (unit_type=1) 的 /88 room cards
//   - 读各卡 card:state.room_state.total_people
//   - per-card segment 累加（≥2 推进；<2 重置）；segment 跨 5min 阈值即触发 visitor
//   - 写入路径走 TargetMerger.ApplyVisitor（per-card visitor 字段 inject 进 max-merge result）
//   - 本地午夜（按 unit timezone）reset today 三字段；新一天重新计算
//
// 物理实体 vs card 视图：visitor 是跨 device/room 的"组合人数"派生 → 属 card 层语义，
// sensor 不参与（[[visitor_belongs_to_cardagg]] §"Visitor 职责分工"）。
//
// **5min 阈值的物理意义**：进房 5min 才算 visitor，过滤"快闪进出"假阳；段断（人离开）
// 也不立即清 today 字段——同一天再来还是同一 visitor 概念。
//
// 时区：当前用 UTC 算午夜（unit timezone 配置后续补；详 doc/card_display.md §4.4 与
// memory [[server_internal_utc_only.md]]：UTC 内部 + TZ 仅 API 边界）。

package consumer

import (
	"context"
	"sync"
	"time"

	"owl-common/card"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

const (
	visitorThresholdMin       = 5  // segment ≥ 5min 触发 visitor
	visitorStandingCapMin     = 0  // unused; standing 由 sensor 算
	visitorDefaultTickEvery   = 60 // tick interval seconds
	visitorPrivateUnitType    = 1  // matches card.UnitTypePrivate; 复制常量值避免 import 循环（card 包定义 card_types.go）
)

// VisitorDeriver 60s tick 任务，给每张 Private 父 unit 下 /88 room card 计算 visitor 状态。
type VisitorDeriver struct {
	metaCache *service.DeviceMetaCache
	reader    *card.Reader
	merger    *service.TargetMerger
	logger    *zap.Logger

	interval time.Duration

	mu       sync.Mutex
	segments map[string]*visitorSegment // key = /88 room cardID
}

// visitorSegment per-card 段累积态。
type visitorSegment struct {
	cardID           string
	segmentStartTs   int64 // ms; 0 = 当前无 ongoing visitor segment
	segDurationMin   int   // 当前段累积分钟（≥2 持续命中累加）
	todayMaxMin      int   // 当日跨段最大
	hasToday         bool  // 当日是否曾达 5min 阈值
	visitorStartTs   int64 // 最近一次达阈值时的 segment_start_ts（写到 Target.VisitorStartTs）
	lastTickDateUTC  string // YYYY-MM-DD（UTC）；跨日 reset
}

// PrivateRoomCardLister VisitorDeriver 从 metaCache 拉"父 unit_type=Private 的 /88 room cards"。
// 实现：service.DeviceMetaCache.ListPrivateRoomCardIDs。
type PrivateRoomCardLister interface {
	ListPrivateRoomCardIDs(ctx context.Context) []string
}

// NewVisitorDeriver 构造。interval=0 走默认 60s。
func NewVisitorDeriver(
	metaCache *service.DeviceMetaCache,
	reader *card.Reader,
	merger *service.TargetMerger,
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

// tick 一次扫描所有 Private 父下的 /88 room cards。
func (v *VisitorDeriver) tick(ctx context.Context, now time.Time) {
	cardIDs := v.metaCache.ListPrivateRoomCardIDs(ctx)
	if len(cardIDs) == 0 {
		return
	}
	nowMs := now.UnixMilli()
	dateUTC := now.UTC().Format("2006-01-02")

	for _, cardID := range cardIDs {
		v.tickCard(ctx, cardID, nowMs, dateUTC)
	}
}

// tickCard 处理单 /88 room card 的累加 + 阈值 + midnight reset。
func (v *VisitorDeriver) tickCard(ctx context.Context, cardID string, nowMs int64, dateUTC string) {
	totalPeople := v.readTotalPeople(ctx, cardID)

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

	if totalPeople >= 2 {
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
		// segment 断（人离开 / 房间空）：reset segment but 保留 today 状态
		seg.segmentStartTs = 0
		seg.segDurationMin = 0
	}

	visitorStartTs := seg.visitorStartTs
	todayMax := seg.todayMaxMin
	hasToday := seg.hasToday
	v.mu.Unlock()

	// 通过 TargetMerger 注入 visitor 字段（merger 合并 device + visitor 后写 hash）
	if v.merger != nil {
		merged := v.merger.ApplyVisitor(ctx, cardID, service.MakeVisitorFields(visitorStartTs, todayMax, hasToday))
		_ = merged // 写入由 merger 内部处理；当前我们只保证累积态注入了
		// 若需要立即触发 card:state hash 写入，可在此调 writer；当前依赖下次 target.state
		// 消息触发 SensorStateProjector 整段写。VisitorDeriver 不直接写 hash 避免 race。
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
