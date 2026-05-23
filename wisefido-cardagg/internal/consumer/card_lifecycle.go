// card_lifecycle.go — config:card:stream 消费者。
//
// 收到 config.changed → 按 op + affected 范围失效 cache + 刷新 AlarmState：
//   - op=reset  → 全 cache rebuild
//   - op=delete → 对每个 cards[i] 清 Redis card:state hash + drop merger snapshot
//   - op=update → 对每个 cards[i] metaCache.Remove + RefreshDeviceIndexForCard
//   通用：deviceAddrs[] → enablement.InvalidateDevices；cards[] → UnitPicker.InvalidateUnit (derive /80)
package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/netip"

	"owl-common/card"

	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type configChangedData struct {
	Op          string   `json:"op"`
	Cards       []string `json:"cards"`
	DeviceAddrs []string `json:"device_addrs"`
	DeviceUIDs  []string `json:"device_uids"`
}

type CardLifecycle struct {
	db          *sql.DB
	writer      *card.Writer
	reader      *card.Reader
	redisClient *redislib.Client
	metaCache   *service.DeviceMetaCache
	enablement  *service.AlarmEnablementCache
	picker      *UnitPicker
	merger      *service.TargetMerger
	tracker     *DeviceStatusTracker
	logger      *zap.Logger
}

func NewCardLifecycle(db *sql.DB, writer *card.Writer, reader *card.Reader, redisClient *redislib.Client, meta *service.DeviceMetaCache, enable *service.AlarmEnablementCache, picker *UnitPicker, logger *zap.Logger) *CardLifecycle {
	return &CardLifecycle{
		db:          db,
		writer:      writer,
		reader:      reader,
		redisClient: redisClient,
		metaCache:   meta,
		enablement:  enable,
		picker:      picker,
		logger:      logger,
	}
}

func (h *CardLifecycle) SetTargetMerger(m *service.TargetMerger) {
	if h == nil {
		return
	}
	h.merger = m
}

// SetDeviceStatusTracker rebind/unbind 时由 CardLifecycle 调用 tracker.MigrateDevice，
// 清掉旧 addr 的内存 entry + Redis device:status hash，防 watchdog 把旧 addr 误标 offline。
func (h *CardLifecycle) SetDeviceStatusTracker(t *DeviceStatusTracker) {
	if h == nil {
		return
	}
	h.tracker = t
}

func (h *CardLifecycle) Handle(ctx context.Context, raw map[string]interface{}) error {
	dataStr, _ := raw["data"].(string)
	if dataStr == "" {
		return nil
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		h.logger.Warn("config.changed parse envelope", zap.Error(err))
		return nil
	}
	var d configChangedData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		h.logger.Warn("config.changed parse data", zap.Error(err))
		return nil
	}

	switch d.Op {
	case "reset":
		// data 重启了，in-memory lookup cache 全失效；从 DB 重 build 所有卡 display 兜消息丢失边界。
		// data 端不再清 Redis card:state（保留），所以这里是平滑 overwrite，不是 fill-empty。
		defer RebuildAllFromDB(ctx, h.db, h.reader, h.writer, h.picker, h.metaCache, h.logger)

		h.metaCache.InvalidateAll()
		if err := h.metaCache.BuildDeviceIndex(ctx); err != nil {
			h.logger.Warn("rebuild device index", zap.Error(err))
		}
		h.enablement.InvalidateAll()
		if h.picker != nil {
			h.picker.InvalidateAll()
		}
		if h.merger != nil {
			h.merger.ResetAllVisitor()
		}
		return nil

	case "delete":
		for _, cardID := range d.Cards {
			h.metaCache.Remove(cardID)
			h.metaCache.RefreshDeviceIndexForCard(ctx, cardID)
			if err := h.writer.DeleteCardState(ctx, cardID); err != nil {
				h.logger.Warn("delete card state", zap.String("cid", cardID), zap.Error(err))
			}
			if h.merger != nil {
				h.merger.ForgetCard(cardID)
			}
			if h.picker != nil {
				if unitPref, ok := parentUnitPrefix(cardID); ok {
					h.picker.InvalidateUnit(unitPref)
				}
			}
		}
		if len(d.DeviceAddrs) > 0 {
			h.enablement.InvalidateDevices(d.DeviceAddrs)
		}
		return nil

	default: // update
		for _, cardID := range d.Cards {
			h.metaCache.Remove(cardID)
			h.metaCache.RefreshDeviceIndexForCard(ctx, cardID)
			if h.picker != nil {
				if unitPref, ok := parentUnitPrefix(cardID); ok {
					h.picker.InvalidateUnit(unitPref)
				}
			}
			cas, err := card.QueryCardAlarmState(ctx, h.db, cardID)
			if err != nil {
				h.logger.Warn("query alarm state", zap.String("cid", cardID), zap.Error(err))
			} else if cas != nil {
				if err := h.writer.WriteCardStatus(ctx, &card.CardStatus{
					CardID:     cardID,
					AlarmState: cas.ToAlarmState(),
				}); err != nil {
					h.logger.Warn("write alarm state", zap.String("cid", cardID), zap.Error(err))
				}
			}
			// has_bed_changed / has_bathroom_changed 等 flag 翻转 → display 必须重 build；
			// 否则 sensor 没下条事件来重 derive，display 一直停在旧 has_bed 时的 card_priority。
			h.rebuildDisplay(ctx, cardID)
		}
		if len(d.DeviceAddrs) > 0 {
			h.enablement.InvalidateDevices(d.DeviceAddrs)
		}
		// device_status rebind-migration：device_addrs[] 是 publish 时刻的 addr（可能是 bind 前 / 不变 /
		// 不再存在）；查 DB 拿 device_uid 对应的 currentAddr，若与 oldAddr 不同 → 清旧 Redis hash + 内存
		// entry，防 watchdog 把旧 addr 错标 offline。
		// 配对长度按短者，未配对的 addr 跳过（无 uid 不知道是否 rebind）。
		if h.tracker != nil && len(d.DeviceUIDs) > 0 {
			n := len(d.DeviceUIDs)
			if len(d.DeviceAddrs) < n {
				n = len(d.DeviceAddrs)
			}
			for i := 0; i < n; i++ {
				uid := d.DeviceUIDs[i]
				oldAddr := d.DeviceAddrs[i]
				if uid == "" {
					continue
				}
				var currentAddr string
				err := h.db.QueryRowContext(ctx,
					`SELECT COALESCE(host(d.device_addr), '') FROM devices d WHERE d.device_uid = $1`, uid,
				).Scan(&currentAddr)
				if err != nil && err != sql.ErrNoRows {
					h.logger.Warn("lookup current device_addr", zap.String("uid", uid), zap.Error(err))
					continue
				}
				// err == sql.ErrNoRows → device 行被删，currentAddr 保持 ""，MigrateDevice 会清旧 addr 并删 uidIndex
				h.tracker.MigrateDevice(ctx, uid, oldAddr, currentAddr)
			}
		}
		return nil
	}
}

// parentUnitPrefix derive /80 unit prefix from card_id CIDR (/80/88/96)。
// /80 returns itself; /88 /96 truncate；其他 returns ok=false。
func parentUnitPrefix(cardID string) (string, bool) {
	p, err := netip.ParsePrefix(cardID)
	if err != nil {
		return "", false
	}
	if p.Bits() == 80 {
		return cardID, true
	}
	if p.Bits() != 88 && p.Bits() != 96 {
		return "", false
	}
	unit := netip.PrefixFrom(p.Addr(), 80).Masked()
	return unit.String(), true
}

// rebuildAllFromDB 从 cards 表枚举所有卡，逐张 rebuildDisplay。
// op=reset 时调（data startup_clear 把 card:state:* 都 DEL 了，Redis SCAN 取不到，必须走 DB）。
func (h *CardLifecycle) rebuildAllFromDB(ctx context.Context) {
	rows, err := h.db.QueryContext(ctx,
		`SELECT host(card_id)||'/'||masklen(card_id) FROM cards`)
	if err != nil {
		h.logger.Warn("rebuildAllFromDB: query cards failed", zap.Error(err))
		return
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			continue
		}
		h.rebuildDisplay(ctx, cardID)
		count++
	}
	h.logger.Info("rebuildAllFromDB done", zap.Int("cards", count))
}

// rebuildDisplay 单卡重 build card:state.display —— has_bed/has_bathroom flag 翻转后，
// 用最新 CardMeta + 现有 Redis 里的 RoomState/BedState/Target 重派生 CardPriority 等字段。
//
// 不同 mask 路径：
//   - /80 → picker.RefreshSelf（用 children 算 unit display，含 /80 自家 alarm/visitor）
//   - /88 /96 → 直接调 BuildCardDisplay 单卡视角；再 RefreshParent 让 /80 跟着算
func (h *CardLifecycle) rebuildDisplay(ctx context.Context, cardID string) {
	pfx, err := netip.ParsePrefix(cardID)
	if err != nil {
		return
	}
	if pfx.Bits() == 80 {
		if h.picker != nil {
			h.picker.RefreshSelf(ctx, cardID)
		}
		return
	}
	if pfx.Bits() != 88 && pfx.Bits() != 96 {
		return
	}
	status, _ := h.reader.ReadCardStatus(ctx, cardID)
	if status == nil {
		status = &card.CardStatus{CardID: cardID}
	}
	hasBed, isBath := false, false
	if h.metaCache != nil {
		m := h.metaCache.GetOrLoad(ctx, cardID)
		hasBed = m.HasBed()
		isBath = m.IsBathroom()
	}
	display := BuildCardDisplay(status, hasBed, isBath)
	if display != nil {
		if err := h.writer.WriteCardStatus(ctx, &card.CardStatus{
			CardID:  cardID,
			Display: display,
		}); err != nil {
			h.logger.Warn("rebuildDisplay write", zap.String("cid", cardID), zap.Error(err))
		}
	}
	if h.picker != nil {
		h.picker.RefreshParent(ctx, cardID)
	}
}
