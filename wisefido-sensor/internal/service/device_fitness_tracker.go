// device_fitness_tracker.go — per-device 健康状态跟踪（S6 sensor 收 4 类设备 alarm 影响 engine）。
//
// 用户 2026-05-19 拍板：device offline / sensor_detached / angle_exception / signal_poor
// 触发后该 device 的信号入 zone engine 前必须 gate（unfit → skip），避免 stale/假数据污染 FSM。
//
// 4 类 unfit reason（owl-common/alarm 常量）：
//   - Offline           ↔ OfflineRecover           — 离线（最严重；常见）
//   - SensorDetached    ↔ SensorDetachedRecover    — 物理脱床（垃圾数据 P0 风险）
//   - AngleException    ↔ AngleExceptionRecover    — 物理偏移（pose 数据错乱）
//   - SignalPoor        ↔ SignalPoorRecover        — 信号弱（data quality 警告）
//
// 设计：任一 reason 标 unfit → device unfit；对应 Recover 单独清；全清才 fit。
// adapter_sleepace / adapter_radar handleMsg 入口调 IsFit(deviceAddr)；false → skip
// 不喂 engine.Apply（防 stale/假数据污染 FSM 与派生 alarm）。
//
// **不影响 cardagg side**：cardagg DeviceStatusTracker 是 cardagg 自家 device fitness
// （独立维度，影响 target_merger LastActive/Standing UI 灰化）；sensor 这套是 sensor
// 自家影响 FSM 决策，互不耦合。

package service

import (
	"sync"

	"go.uber.org/zap"
)

// 4 类 unfit reason（位标志；OR 入 device flags）。
// 与 consumer 包同步的 fitnessReason* 镜像常量；caller 跨包传 uint8 不需 type cast。
const (
	FitnessReasonOffline        uint8 = 1 << 0
	FitnessReasonSensorDetached uint8 = 1 << 1
	FitnessReasonAngleException uint8 = 1 << 2
	FitnessReasonSignalPoor     uint8 = 1 << 3
)

// UnfitCallback device 从 fit → unfit 边沿触发；各持 per-device in-memory state 的组件
// 注册自己的 clear 函数（"device offline = 内存 state 重启"原则，用户 2026-05-19 拍板）。
//
// 已 fire 入 DB 的 alarm 不丢（DB 单写者）；只清没 fire 的 pending / cache / accumulator。
//
// 注册方：
//   - TrackManager.ClearDevice (via Engine.OnDeviceUnfit) — bedSessions / sleepadStates
//   - TargetStateAggregator.ForgetDevice — weakBio score / standing / lastActive accumulator
//   - SleepStageConsumer.OnDeviceUnfit — ladder state + emit bed.sleepstage(0,0)
//   - BedEventDedup.Forget — 10s 同类 dedup map
type UnfitCallback func(deviceAddr string)

// DeviceFitnessTracker per-device 健康状态。线程安全。
//
// state map: deviceAddr (canonical IPv6 string) → fitnessFlags。0 = fit；非 0 = 至少 1 类 unfit。
type DeviceFitnessTracker struct {
	mu               sync.RWMutex
	flags            map[string]uint8
	callbacks        []UnfitCallback
	offlineCallbacks []UnfitCallback
	logger           *zap.Logger
}

func NewDeviceFitnessTracker(logger *zap.Logger) *DeviceFitnessTracker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DeviceFitnessTracker{
		flags:  make(map[string]uint8),
		logger: logger,
	}
}

// RegisterUnfitCallback 注册 device fit→unfit 边沿回调（仅首次 unfit 触发，避免重复 mark spam）。
// main wiring 调；callback 应 idempotent（被重复调不破）。
func (t *DeviceFitnessTracker) RegisterUnfitCallback(fn UnfitCallback) {
	if t == nil || fn == nil {
		return
	}
	t.mu.Lock()
	t.callbacks = append(t.callbacks, fn)
	t.mu.Unlock()
}

// RegisterOfflineCallback 注册 device-offline 专用回调：仅 Offline reason 新置位时触发（顺序无关，
// 即便设备先因 SignalPoor 已 unfit、后转 Offline 也照触发）。与 fit→unfit 首边沿的 UnfitCallback 分离，
// 因清 radar track 只该在整机离线时做，不该被 SignalPoor/AngleException 等抖动触发。callback 须 idempotent。
func (t *DeviceFitnessTracker) RegisterOfflineCallback(fn UnfitCallback) {
	if t == nil || fn == nil {
		return
	}
	t.mu.Lock()
	t.offlineCallbacks = append(t.offlineCallbacks, fn)
	t.mu.Unlock()
}

// MarkUnfit 标 device 为 unfit（OR 入 reason flag）+ 在 fit→unfit 边沿广播 clear callback。
//
// 边沿检测：prev==0 → non-zero 时为首次 unfit，broadcast 给所有 callbacks 清 device-related
// state（"device offline = 内存重启"原则）。后续重复 Mark（不同 reason 累加）不再 broadcast，
// state 已清没必要再清。
//
// 注：broadcast 在锁外执行 — callback 可能反向调 IsFit / 持自己锁，必须避免死锁。
func (t *DeviceFitnessTracker) MarkUnfit(deviceAddr string, reason uint8) {
	if t == nil || deviceAddr == "" || reason == 0 {
		return
	}
	t.mu.Lock()
	prev := t.flags[deviceAddr]
	t.flags[deviceAddr] = prev | reason
	newlySet := reason &^ prev
	var firstEdge, offlineEdge []UnfitCallback
	if prev == 0 {
		firstEdge = make([]UnfitCallback, len(t.callbacks))
		copy(firstEdge, t.callbacks)
	}
	if newlySet&FitnessReasonOffline != 0 {
		offlineEdge = make([]UnfitCallback, len(t.offlineCallbacks))
		copy(offlineEdge, t.offlineCallbacks)
	}
	t.mu.Unlock()

	for _, fn := range offlineEdge {
		fn(deviceAddr)
	}

	if prev == 0 {
		t.logger.Info("device fitness: marked unfit (first edge → broadcast clear)",
			zap.String("device_addr", deviceAddr),
			zap.Uint8("reason", reason),
		)
		for _, fn := range firstEdge {
			fn(deviceAddr)
		}
	} else if prev&reason == 0 {
		t.logger.Info("device fitness: marked unfit (additional reason, state already cleared)",
			zap.String("device_addr", deviceAddr),
			zap.Uint8("reason", reason),
			zap.Uint8("flags_after", prev|reason),
		)
	}
}

// ClearReason 清除指定 reason flag（Recover alarm 触发）；其他 reason flag 保留。
func (t *DeviceFitnessTracker) ClearReason(deviceAddr string, reason uint8) {
	if t == nil || deviceAddr == "" || reason == 0 {
		return
	}
	t.mu.Lock()
	prev, ok := t.flags[deviceAddr]
	if !ok {
		t.mu.Unlock()
		return
	}
	next := prev &^ reason
	if next == 0 {
		delete(t.flags, deviceAddr)
	} else {
		t.flags[deviceAddr] = next
	}
	t.mu.Unlock()
	if prev&reason != 0 {
		t.logger.Info("device fitness: cleared reason",
			zap.String("device_addr", deviceAddr),
			zap.Uint8("reason", reason),
			zap.Uint8("flags_after", next),
		)
	}
}

// IsFit device 是否全 fit（无 unfit reason）。空 deviceAddr 视作 fit（兜底兼容）。
func (t *DeviceFitnessTracker) IsFit(deviceAddr string) bool {
	if t == nil || deviceAddr == "" {
		return true
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.flags[deviceAddr] == 0
}

// Reasons 返回 device 当前 unfit reason mask（调试用）。0 = fit。
func (t *DeviceFitnessTracker) Reasons(deviceAddr string) uint8 {
	if t == nil || deviceAddr == "" {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.flags[deviceAddr]
}

// Forget 清理 device 全部 entries（device 解绑 / 永久下线时调）。
func (t *DeviceFitnessTracker) Forget(deviceAddr string) {
	if t == nil || deviceAddr == "" {
		return
	}
	t.mu.Lock()
	delete(t.flags, deviceAddr)
	t.mu.Unlock()
}
