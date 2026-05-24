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
	cache       *service.SpatialCache
	enablement  *service.AlarmEnablementCache
	rebuilder   *DisplayRebuilder
	merger      *service.TargetMerger
	tracker     *DeviceStatusTracker
	logger      *zap.Logger
}

func NewCardLifecycle(db *sql.DB, writer *card.Writer, reader *card.Reader, redisClient *redislib.Client, cache *service.SpatialCache, enable *service.AlarmEnablementCache, rebuilder *DisplayRebuilder, logger *zap.Logger) *CardLifecycle {
	return &CardLifecycle{
		db:          db,
		writer:      writer,
		reader:      reader,
		redisClient: redisClient,
		cache:       cache,
		enablement:  enable,
		rebuilder:   rebuilder,
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
		// schema 全失效：cache 全量 reload + display 全量重 build。
		h.cache.Invalidate(ctx)
		h.enablement.InvalidateAll()
		if h.merger != nil {
			h.merger.ResetAllVisitor()
		}
		if h.rebuilder != nil {
			h.rebuilder.RebuildAll(ctx)
		}
		return nil

	case "delete":
		h.cache.Invalidate(ctx)
		for _, cardID := range d.Cards {
			if err := h.writer.DeleteCardState(ctx, cardID); err != nil {
				h.logger.Warn("delete card state", zap.String("cid", cardID), zap.Error(err))
			}
			if h.merger != nil {
				h.merger.ForgetCard(cardID)
			}
		}
		if len(d.DeviceAddrs) > 0 {
			h.enablement.InvalidateDevices(d.DeviceAddrs)
		}
		return nil

	default: // update
		h.cache.Invalidate(ctx)
		for _, cardID := range d.Cards {
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
			// has_bed_changed / has_bathroom_changed 等 flag 翻转 → display 必须重 build。
			if h.rebuilder != nil {
				h.rebuilder.Rebuild(ctx, cardID)
			}
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

