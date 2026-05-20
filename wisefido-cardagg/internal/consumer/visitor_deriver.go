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
//   - 通过 TargetMerger.ApplyVisitor 注入 visitor 字段，最终由 SensorStateProjector 写 hash
//   - 本地午夜（按 unit timezone，当前用 UTC）reset visitor 字段
//
// **Episode 持久化（2026-05-19 加）**：跨 5min 阈值时 INSERT visitor_history (started_at, ended_at=NULL)；
// 下降沿 ≥2→<2 时 UPDATE 关 episode；午夜 reset 时关进行中 episode（ended_at=midnight）。
// 物理上跨日的连续访客在 DB 里被切两条 record（每日数据封闭，FE query 当天即可）。
//
// 物理实体 vs card 视图：visitor 是跨 device/room 的"组合人数"派生 → 属 card 层语义，
// sensor 不参与（[[visitor_belongs_to_cardagg]] §"Visitor 职责分工"）。
//
// **5min 阈值的物理意义**：进房 5min 才算 visitor，过滤"快闪进出"假阳；段断（人离开）
// 立刻清 segment，同一天再来重新累加；HasVisitorToday 保留至午夜。
//
// 时区：当前用 UTC 算午夜（unit timezone 配置后续补；详 doc/card_display.md §4.4 与
// memory [[server_internal_utc_only]]：UTC 内部 + TZ 仅 API 边界）。
package consumer

import (
	"context"
	"database/sql"
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

// visitorHistoryStore 抽象 DB writer（测试可 mock）。
type visitorHistoryStore interface {
	insertEpisode(ctx context.Context, cardID string, startedAtMs int64) (string, error)
	closeEpisode(ctx context.Context, episodeID string, endedAtMs int64, durationSec int) error
}

// cardStatusWriter 抽象 hash 写（测试可 mock）。*card.Writer 隐式满足。
type cardStatusWriter interface {
	WriteCardStatus(ctx context.Context, status *card.CardStatus) error
}

// parentRefresher 抽象父 unit 卡刷新（测试可 mock）。*UnitPicker 隐式满足。
type parentRefresher interface {
	RefreshParent(ctx context.Context, cardID string)
}

// visitorMergerApplier 抽象 merger 接口（测试可 mock）。*service.TargetMerger 隐式满足。
type visitorMergerApplier interface {
	ApplyVisitor(ctx context.Context, cardID string, v service.VisitorFields) *card.TargetState
}

// VisitorDeriver 60s tick 任务：bed level + room level 双路径计算 visitor。
type VisitorDeriver struct {
	metaCache *service.DeviceMetaCache
	reader    *card.Reader
	merger    visitorMergerApplier
	bedPeople *service.BedPeopleTracker
	store     visitorHistoryStore // 可 nil（未 wire DB 时退化为纯内存累加）
	writer    cardStatusWriter    // 可 nil；用于午夜 reset 时主动落 hash 清残留
	picker    parentRefresher     // 可 nil；午夜 reset 写 hash 后同步刷父 unit 卡
	logger    *zap.Logger

	interval time.Duration

	mu       sync.Mutex
	segments map[string]*visitorSegment // key = cardID (/88 room 或 /96 bed)
}

// visitorSegment per-card 段累积态。
type visitorSegment struct {
	cardID          string
	segmentStartTs  int64  // ms; 0 = 当前无 ongoing segment
	segDurationMin  int    // 当前段累积分钟（≥2 持续命中累加）
	hasToday        bool   // 当日是否曾达 5min 阈值
	visitorStartTs  int64  // 最近一次达阈值时的 segment_start_ts（写到 Target.VisitorStartTs）
	currentEpisode  string // history 表行 id（segment 跨阈值后非空；下降沿/午夜关闭后清空）
	lastTickDateUTC string // YYYY-MM-DD（UTC）；跨日 reset
}

// NewVisitorDeriver 构造。interval=0 走默认 60s。store/writer/picker=nil 时退化（兼容测试）。
// 非午夜场景 hash 字段沿用 sensor target.state 自然带的路径；writer/picker 仅午夜跨日 reset 时主动写。
func NewVisitorDeriver(
	metaCache *service.DeviceMetaCache,
	reader *card.Reader,
	merger visitorMergerApplier,
	bedPeople *service.BedPeopleTracker,
	store visitorHistoryStore,
	writer cardStatusWriter,
	picker parentRefresher,
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
		store:     store,
		writer:    writer,
		picker:    picker,
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

// tickCard 处理单 card 的累加 + 阈值 + midnight reset + episode 持久化。
func (v *VisitorDeriver) tickCard(ctx context.Context, cardID string, peopleCount int, nowMs int64, dateUTC string) {
	v.mu.Lock()
	seg, ok := v.segments[cardID]
	if !ok {
		seg = &visitorSegment{cardID: cardID, lastTickDateUTC: dateUTC}
		v.segments[cardID] = seg
	}

	// midnight reset (UTC)：进行中 episode 关到 midnight，清今日字段
	var midnightCloseID string
	var midnightCloseEndMs int64
	var midnightCloseDur int
	midnightHappened := false
	if seg.lastTickDateUTC != "" && seg.lastTickDateUTC != dateUTC {
		midnightHappened = true
		if seg.currentEpisode != "" && seg.segmentStartTs > 0 {
			midnightCloseEndMs = midnightMsForDate(seg.lastTickDateUTC)
			midnightCloseID = seg.currentEpisode
			midnightCloseDur = int((midnightCloseEndMs - seg.segmentStartTs) / 1000)
		}
		seg.hasToday = false
		seg.visitorStartTs = 0
		seg.currentEpisode = ""
		seg.segmentStartTs = 0
		seg.segDurationMin = 0
	}
	seg.lastTickDateUTC = dateUTC

	// 状态机：≥2 推进；<2 重置 + 下降沿关 episode
	var openStartedAt int64 // >0 时本 tick 触发上升沿，需 INSERT
	var closeID string
	var closeDur int
	if peopleCount >= 2 {
		if seg.segmentStartTs == 0 {
			seg.segmentStartTs = nowMs
			seg.segDurationMin = 1
		} else {
			seg.segDurationMin++
		}
		// 跨 5min 阈值：上升沿（仅本 tick 首次跨；后续 ≥5 不重复 INSERT）
		if seg.segDurationMin == visitorThresholdMin && seg.currentEpisode == "" {
			openStartedAt = seg.segmentStartTs
		}
		if seg.segDurationMin >= visitorThresholdMin {
			seg.visitorStartTs = seg.segmentStartTs
			seg.hasToday = true
		}
	} else {
		// 下降沿：≥2→<2，若已开 episode 则关
		if seg.currentEpisode != "" && seg.segmentStartTs > 0 {
			closeID = seg.currentEpisode
			closeDur = int((nowMs - seg.segmentStartTs) / 1000)
			seg.currentEpisode = ""
		}
		seg.segmentStartTs = 0
		seg.segDurationMin = 0
	}

	visitorStartTs := seg.visitorStartTs
	hasToday := seg.hasToday
	v.mu.Unlock()

	// 锁外做 DB IO：先关旧 episode（午夜 / 下降沿），再开新 episode
	if midnightCloseID != "" && v.store != nil {
		if err := v.store.closeEpisode(ctx, midnightCloseID, midnightCloseEndMs, midnightCloseDur); err != nil {
			v.logger.Warn("visitor_history: midnight close failed",
				zap.String("cid", cardID), zap.String("eid", midnightCloseID), zap.Error(err))
		}
	}
	if closeID != "" && v.store != nil {
		if err := v.store.closeEpisode(ctx, closeID, nowMs, closeDur); err != nil {
			v.logger.Warn("visitor_history: close failed",
				zap.String("cid", cardID), zap.String("eid", closeID), zap.Error(err))
		}
	}
	if openStartedAt > 0 && v.store != nil {
		newID, err := v.store.insertEpisode(ctx, cardID, openStartedAt)
		if err != nil {
			v.logger.Warn("visitor_history: insert failed",
				zap.String("cid", cardID), zap.Int64("start_ms", openStartedAt), zap.Error(err))
		} else {
			v.mu.Lock()
			if s := v.segments[cardID]; s != nil && s.currentEpisode == "" {
				s.currentEpisode = newID
			}
			v.mu.Unlock()
		}
	}

	// 注入 TargetMerger（非午夜场景：visitor 字段沿用 sensor target.state 自然带的路径，
	// ~60s 内自然刷新到 hash；活动事件源源不断时延迟可忽略）。
	// 午夜跨日例外：凌晨可能无 activity event 推送，sensor target.state 不发，hash 会卡
	// 昨日 hasToday=true / visitorStartTs 数小时不变 —— 主动写一次 hash 清残留。
	if v.merger != nil {
		merged := v.merger.ApplyVisitor(ctx, cardID, service.MakeVisitorFields(visitorStartTs, hasToday))
		if midnightHappened && merged != nil && v.writer != nil {
			status := &card.CardStatus{CardID: cardID, Target: merged}
			if err := v.writer.WriteCardStatus(ctx, status); err != nil {
				v.logger.Warn("visitor_deriver: midnight hash write failed",
					zap.String("cid", cardID), zap.Error(err))
			} else if v.picker != nil {
				v.picker.RefreshParent(ctx, cardID)
			}
		}
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

// midnightMsForDate 取 dateUTC（YYYY-MM-DD）的次日 00:00:00 UTC ms（即该日的午夜终点）。
func midnightMsForDate(dateUTC string) int64 {
	t, err := time.Parse("2006-01-02", dateUTC)
	if err != nil {
		return 0
	}
	return t.UTC().Add(24 * time.Hour).UnixMilli()
}

// visitorHistorySQLStore 默认 DB writer 实现。
type visitorHistorySQLStore struct {
	db *sql.DB
}

// NewVisitorHistorySQLStore 构造默认 DB writer。db=nil 返 nil（caller 退化为内存模式）。
func NewVisitorHistorySQLStore(db *sql.DB) visitorHistoryStore {
	if db == nil {
		return nil
	}
	return &visitorHistorySQLStore{db: db}
}

func (s *visitorHistorySQLStore) insertEpisode(ctx context.Context, cardID string, startedAtMs int64) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO visitor_history (card_id, started_at)
		 VALUES ($1::inet, to_timestamp($2::bigint / 1000.0))
		 RETURNING history_id::text`,
		cardID, startedAtMs).Scan(&id)
	return id, err
}

func (s *visitorHistorySQLStore) closeEpisode(ctx context.Context, episodeID string, endedAtMs int64, durationSec int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE visitor_history
		   SET ended_at = to_timestamp($2::bigint / 1000.0),
		       duration_sec = $3
		 WHERE history_id = $1::uuid`,
		episodeID, endedAtMs, durationSec)
	return err
}
