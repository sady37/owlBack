// device_status_tracker.go — device:status:{addr} Hash 单 writer 守门。
//
// 输入端：
//   - alarm_router.DeviceSignal 接口（设备类 alarm/recover/connectivity）
//   - monitor_buffer 入口（每条 monitor sample 触刷 last_seen）
//
// 输出：device:status:{deviceAddr} Hash 增量字段 + 180s 看门狗 fail-safe。
//
// 性能优化（vs v1）：
//   - monitor 每条 sample 走 TouchLastSeen，纯内存 map RMW（不写 Redis）
//   - 仅在 transition（offline→online / online→offline）时写 Redis，省 ~99% IO
//   - 看门狗扫内存 map 找 stale device，必要时写 Redis offline=1

package consumer

import (
	"context"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/card"

	"go.uber.org/zap"
)

// DefaultStaleAfter / DefaultWatchdogInterval 看门狗参数。
// staleAfter = 180s ≈ 2 × sleepace OfflineRecover 80s 周期 + 余量；
// qinglan radar 1Hz monitor 流远在阈值之内。
const (
	DefaultStaleAfter        = 180 * time.Second
	DefaultWatchdogInterval  = 30 * time.Second
)

// alarmType → device:status 字段名（除 Offline 外的设备健康类）
var deviceFlagMap = map[string]string{
	alarm.SignalPoor:     "signal_poor",
	alarm.AngleException: "angle_abnormal",
	alarm.SensorDetached: "sensor_detached",
}

// recoverType → 对应 flag（清零）
var deviceFlagRecoverMap = map[string]string{
	alarm.SingalPoorRecover:     "signal_poor",
	alarm.AngleExceptionRecover: "angle_abnormal",
	alarm.SensorDetachedRecover: "sensor_detached",
}

// deviceLiveness 内存中的设备活跃记录。
type deviceLiveness struct {
	lastSeenMs int64
	deviceType string
	online     bool
}

type DeviceStatusTracker struct {
	writer          *card.Writer
	mu              sync.Mutex
	state           map[string]*deviceLiveness // deviceAddr → liveness
	staleAfter      time.Duration
	scanInterval    time.Duration
	logger          *zap.Logger
}

func NewDeviceStatusTracker(writer *card.Writer, logger *zap.Logger) *DeviceStatusTracker {
	return &DeviceStatusTracker{
		writer:       writer,
		state:        make(map[string]*deviceLiveness),
		staleAfter:   DefaultStaleAfter,
		scanInterval: DefaultWatchdogInterval,
		logger:       logger,
	}
}

// OnDeviceAlarm 设备类 alarm onset（SignalPoor / AngleException / SensorDetached）。
// Offline 走 OnDeviceConnectivity，不在本路径。
func (t *DeviceStatusTracker) OnDeviceAlarm(ctx context.Context, deviceAddr, alarmType string, _ map[string]interface{}) {
	field, ok := deviceFlagMap[alarmType]
	if !ok {
		return
	}
	if err := t.writer.PatchDeviceStatus(ctx, deviceAddr, map[string]interface{}{field: "1"}); err != nil {
		t.logger.Warn("device_status patch", zap.String("addr", deviceAddr), zap.String("field", field), zap.Error(err))
	}
}

// OnDeviceRecover 设备类 recover（SignalPoorRecover 等）→ 清 flag。
func (t *DeviceStatusTracker) OnDeviceRecover(ctx context.Context, deviceAddr, recoverType string) {
	field, ok := deviceFlagRecoverMap[recoverType]
	if !ok {
		return
	}
	if err := t.writer.PatchDeviceStatus(ctx, deviceAddr, map[string]interface{}{field: "0"}); err != nil {
		t.logger.Warn("device_status recover", zap.String("addr", deviceAddr), zap.String("field", field), zap.Error(err))
	}
}

// OnDeviceConnectivity Offline / OfflineRecover / DeviceRecover 流来的连通性。
// alarm-driven 路径必写 Redis（不与内存 dedup —— alarm 比 monitor 权威）。
func (t *DeviceStatusTracker) OnDeviceConnectivity(ctx context.Context, deviceAddr, deviceType string, online bool) {
	if err := t.writer.SetDeviceOnline(ctx, deviceAddr, deviceAddr, deviceType, online); err != nil {
		t.logger.Warn("device_status connectivity", zap.String("addr", deviceAddr), zap.Bool("online", online), zap.Error(err))
		return
	}
	t.mu.Lock()
	dl := t.state[deviceAddr]
	if dl == nil {
		dl = &deviceLiveness{}
		t.state[deviceAddr] = dl
	}
	dl.online = online
	dl.deviceType = deviceType
	if online {
		dl.lastSeenMs = time.Now().UnixMilli()
	}
	t.mu.Unlock()
}

// IsOnline 查询某 device 当前内存判定的在线状态。
// 未见过的 device 返回 false（保守：不曾出现的 device 视作离线，避免它的 stale snapshot
// 参与下游卡级 merge）。
func (t *DeviceStatusTracker) IsOnline(deviceAddr string) bool {
	if t == nil || deviceAddr == "" {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	dl := t.state[deviceAddr]
	return dl != nil && dl.online
}

// TouchLastSeen monitor 流每条 sample 调；纯内存 RMW，仅 offline→online transition 写 Redis。
func (t *DeviceStatusTracker) TouchLastSeen(ctx context.Context, deviceAddr, deviceType string) {
	now := time.Now().UnixMilli()
	t.mu.Lock()
	dl := t.state[deviceAddr]
	wasOffline := dl == nil || !dl.online
	if dl == nil {
		dl = &deviceLiveness{}
		t.state[deviceAddr] = dl
	}
	dl.lastSeenMs = now
	dl.deviceType = deviceType
	dl.online = true
	t.mu.Unlock()

	if !wasOffline {
		return
	}
	// transition: offline → online（首次见 或 watchdog 之前判 offline）
	if err := t.writer.SetDeviceOnline(ctx, deviceAddr, deviceAddr, deviceType, true); err != nil {
		t.logger.Warn("device_status transition online", zap.String("addr", deviceAddr), zap.Error(err))
	}
}

// Run 看门狗：扫内存 state，对 stale device patch offline=1。
func (t *DeviceStatusTracker) Run(ctx context.Context) {
	ticker := time.NewTicker(t.scanInterval)
	defer ticker.Stop()
	t.logger.Info("device_status_tracker started",
		zap.Duration("stale_after", t.staleAfter),
		zap.Duration("interval", t.scanInterval),
	)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.scanStale(ctx)
		}
	}
}

func (t *DeviceStatusTracker) scanStale(ctx context.Context) {
	threshold := time.Now().UnixMilli() - t.staleAfter.Milliseconds()
	var stale []struct {
		addr       string
		deviceType string
	}
	t.mu.Lock()
	for addr, dl := range t.state {
		if dl.online && dl.lastSeenMs < threshold {
			stale = append(stale, struct {
				addr       string
				deviceType string
			}{addr, dl.deviceType})
			dl.online = false
		}
	}
	t.mu.Unlock()
	for _, s := range stale {
		if err := t.writer.SetDeviceOnline(ctx, s.addr, s.addr, s.deviceType, false); err != nil {
			t.logger.Warn("watchdog set offline", zap.String("addr", s.addr), zap.Error(err))
			continue
		}
		t.logger.Info("watchdog marked offline", zap.String("addr", s.addr))
	}
}
