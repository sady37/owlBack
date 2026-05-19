// bed_event_dedup.go — radar/sleepace adapter 共享 InBed/LeftBed 10s 同类去重。
//
// S5b 新增防御（[[cardagg_v1_to_v2_migration_audit]] §S5b：v2 新增，v1 无）。
//
// 防 firmware/sleepad 上报 spam：同一 device 在 10s 内重复发同一 kind 的 bed event
// （e.g. 连发两条 InBed），第二条跳过不再 engine.Apply。能防 95% 误报，可能少量
// 漏报（边界 case 第二条 9.5s 后到的合法 event 会被吞）。
//
// per-device per-kind state（device IPv6 字符串 + "enter"/"leave"），单 device 多 kind
// 互不影响（InBed 9s 后来一条 LeftBed 不会被 enter 的 dedup 吞）。

package zoneengine

import "sync"

// BedEventDedupMs 同类事件 dedup 窗口；10s = 用户拍板（防 95% 误报）。
const BedEventDedupMs int64 = 10_000

// BedEventDedup per-device per-kind 时间窗 dedup 状态持有者。
//
// 用法：adapter 调 Admit(deviceAddr, kind, tsMs)；返回 true 才往 engine.Apply 推。
type BedEventDedup struct {
	mu       sync.Mutex
	lastSeen map[string]map[string]int64 // deviceAddr → kind → tsMs
}

func NewBedEventDedup() *BedEventDedup {
	return &BedEventDedup{lastSeen: make(map[string]map[string]int64)}
}

// Admit 检查 device 是否在 10s 窗口外，是 → 更新 lastSeen 并返回 true；否 → 返回 false 跳过。
// 空 deviceAddr 或 kind 直接放行（不能强阻 caller，保持兼容）。
func (d *BedEventDedup) Admit(deviceAddr, kind string, tsMs int64) bool {
	if d == nil || deviceAddr == "" || kind == "" {
		return true
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	m, ok := d.lastSeen[deviceAddr]
	if !ok {
		m = make(map[string]int64)
		d.lastSeen[deviceAddr] = m
	}
	if last, ok := m[kind]; ok && tsMs-last < BedEventDedupMs {
		return false
	}
	m[kind] = tsMs
	return true
}

// Forget 清理 device 全部 entries（设备下线 / 解绑时 caller 调；MVP 不强制调用，
// lastSeen map 增长上限 = 接入 device 数，量级可控）。
func (d *BedEventDedup) Forget(deviceAddr string) {
	if d == nil || deviceAddr == "" {
		return
	}
	d.mu.Lock()
	delete(d.lastSeen, deviceAddr)
	d.mu.Unlock()
}
