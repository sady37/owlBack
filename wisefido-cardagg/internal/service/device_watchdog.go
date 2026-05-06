package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"owl-common/card"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// DeviceWatchdog 周期扫描 device:status:* hash，对 last_seen_ms 长期未刷新的设备 patch offline=1。
//
// 在线状态由 monitor/event 流（正向：offline=0+last_seen_ms=now）和 alarm Offline（负向：offline=1）
// 事件驱动；本看门狗仅作 fail-safe：当 gateway 自身故障导致三路流全断时，主动把 stale 设备置离线。
//
// 阈值默认 180s = 2 × sleepace OfflineRecover 80s 周期 + 余量；qinglan radar 1Hz monitor 流远在阈值之内。
type DeviceWatchdog struct {
	client     *redislib.Client
	writer     *card.Writer
	staleAfter time.Duration
	interval   time.Duration
	logger     *zap.Logger
}

func NewDeviceWatchdog(client *redislib.Client, writer *card.Writer, staleAfter, interval time.Duration, logger *zap.Logger) *DeviceWatchdog {
	if staleAfter <= 0 {
		staleAfter = 180 * time.Second
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &DeviceWatchdog{
		client:     client,
		writer:     writer,
		staleAfter: staleAfter,
		interval:   interval,
		logger:     logger,
	}
}

func (w *DeviceWatchdog) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.logger.Info("device_watchdog started",
		zap.Duration("interval", w.interval),
		zap.Duration("stale_after", w.staleAfter))
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.scanOnce(ctx)
		}
	}
}

func (w *DeviceWatchdog) scanOnce(ctx context.Context) {
	var cursor uint64
	nowMs := time.Now().UnixMilli()
	thresholdMs := w.staleAfter.Milliseconds()
	flagged := 0
	scanned := 0
	for {
		keys, next, err := w.client.Scan(ctx, cursor, "device:status:*", 200).Result()
		if err != nil {
			w.logger.Warn("device_watchdog scan failed", zap.Error(err))
			return
		}
		for _, key := range keys {
			scanned++
			deviceID := strings.TrimPrefix(key, "device:status:")
			if deviceID == "" {
				continue
			}
			vals, err := w.client.HMGet(ctx, key, "offline", "last_seen_ms").Result()
			if err != nil || len(vals) < 2 {
				continue
			}
			offlineStr := stringFromHMGet(vals[0])
			lastSeenStr := stringFromHMGet(vals[1])
			// 已经 offline=1 → 不重复 patch
			if offlineStr == "1" {
				continue
			}
			lastSeen, _ := strconv.ParseInt(lastSeenStr, 10, 64)
			if lastSeen == 0 {
				// 没有 last_seen 证据（key 仅由 alarm flag patch 出来，从未收到正向数据）
				// → 不主动判离线，留给后续真实事件填充。
				continue
			}
			idle := nowMs - lastSeen
			if idle > thresholdMs {
				if err := w.writer.PatchDeviceStatus(ctx, deviceID, map[string]interface{}{
					"offline": "1",
				}); err != nil {
					w.logger.Warn("device_watchdog patch failed",
						zap.String("device_id", deviceID),
						zap.Error(err))
					continue
				}
				flagged++
				w.logger.Info("device_watchdog mark offline",
					zap.String("device_id", deviceID),
					zap.Int64("idle_ms", idle),
					zap.Int64("threshold_ms", thresholdMs))
			}
		}
		if next == 0 {
			break
		}
		cursor = next
	}
	if flagged > 0 {
		w.logger.Info("device_watchdog cycle done",
			zap.Int("scanned", scanned),
			zap.Int("flagged", flagged))
	}
}

func stringFromHMGet(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	default:
		return fmt.Sprintf("%v", x)
	}
}
