package service

// bed_presence_fusion.go — per-/96 bed 的 sleepad/radar InBed 双源状态跟踪。
//
// 用途：RoomState.TotalPeople dedup。radar firmware NumberPeople 偶尔漏数床上人
// （静卧低 RCS / 床边遮挡 / 极慢动 microdoppler），但 sleepad 压感对在床稳定。
// 反之 radar 看到的 bed-area track 在 NumberPeople 已计入。
//
// 推导公式（2026-06-05 修：旧的 Z + Σ(X && !Y) 加法在两个相位翻车——占用期 radar room
// count 掉 0 漏数静卧人、离床期 radar 还 track + 床垫未清双计 → 改 max）：
//
//	radar_np = radar NumberPeople (room 总数, GetZ)
//	bed_np   = 房内"有人在床"的床数 (sleepad ∪ radar InBed)  ← OccupiedBedsInRoom
//	publishTotalPeople = max(radar_np, bed_np)
//
// max 两头都对：占用期 max(0,1)=1（radar 漏静卧人取 bed_np）、离床期 max(1,0)=1（走出
// 的人只 radar 见取 radar_np，不与床垫双计）、正常 max(1,1)=1（床上人两源都看到取 1 不加 2）。
//
// 不引入 zoneengine 内：BedZone FSM 把 sleepad/radar 两源融合成单一 enter/leave，
// 源信息丢失。这里旁路记账，供 stream_publisher 在 publish 前查询。
//
// 不进 alarm 决策路径（[[feedback_no_dynamic_threshold_modulation]]）：dedup 是 presence
// 计数 fusion，不影响 verdict/severity/routing/dedup/timing。

import (
	"net/netip"
	"sync"
	"time"
)

// BedPresenceFusion 跟踪 per-/96 bed 的 sleepad/radar InBed 状态 + 时间戳。
type BedPresenceFusion struct {
	mu      sync.RWMutex
	entries map[string]*bedPresenceEntry // bedCIDR (/96) → state
	now     func() int64                 // 测试用注入
}

type bedPresenceEntry struct {
	sleepadInBed bool
	radarInBed   bool
	sleepadAt    int64 // 最近 sleepad 更新 ms
	radarAt      int64 // 最近 radar 更新 ms
}

// bedPresenceFreshMs 10min 内的 sleepad/radar 信号视作有效；超时跳过 dedup。
// sleepad 心跳 sample 周期 2-10s（依 sleepace adapter scheduler），10min 足以覆盖任何
// vendor backoff 场景；radar InBed alarm 是 edge 事件无心跳，10min stale 后视同未触发。
const bedPresenceFreshMs = 10 * 60 * 1000

func NewBedPresenceFusion() *BedPresenceFusion {
	return &BedPresenceFusion{
		entries: make(map[string]*bedPresenceEntry),
		now:     func() int64 { return time.Now().UnixMilli() },
	}
}

func (f *BedPresenceFusion) getOrCreate(bedCIDR string) *bedPresenceEntry {
	e := f.entries[bedCIDR]
	if e == nil {
		e = &bedPresenceEntry{}
		f.entries[bedCIDR] = e
	}
	return e
}

// SetSleepad sleepad InBed/LeftBed 翻转时调；ts=0 时取 now。
func (f *BedPresenceFusion) SetSleepad(bedCIDR string, inBed bool, ts int64) {
	if bedCIDR == "" {
		return
	}
	if ts == 0 {
		ts = f.now()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	e := f.getOrCreate(bedCIDR)
	e.sleepadInBed = inBed
	e.sleepadAt = ts
}

// SetRadar radar InBed/LeftBed 翻转时调；ts=0 时取 now。
func (f *BedPresenceFusion) SetRadar(bedCIDR string, inBed bool, ts int64) {
	if bedCIDR == "" {
		return
	}
	if ts == 0 {
		ts = f.now()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	e := f.getOrCreate(bedCIDR)
	e.radarInBed = inBed
	e.radarAt = ts
}

// OccupiedBedsInRoom 在 /88 roomCIDR 下数"有人在床"的 /96 bed 数（sleepad ∪ radar InBed，
// 任一 fresh 即算占用）。max(radar_np, bed_np) dedup 的 bed_np。
// stale entries（信号超 10min 未更新）跳过，避免久未上报的源长期挂"InBed"导致 phantom 占用。
func (f *BedPresenceFusion) OccupiedBedsInRoom(roomCIDR string) int {
	if roomCIDR == "" {
		return 0
	}
	roomPfx, err := netip.ParsePrefix(roomCIDR)
	if err != nil {
		return 0
	}
	now := f.now()
	f.mu.RLock()
	defer f.mu.RUnlock()
	n := 0
	for bedCIDR, e := range f.entries {
		bedPfx, err := netip.ParsePrefix(bedCIDR)
		if err != nil {
			continue
		}
		if !roomPfx.Contains(bedPfx.Addr()) {
			continue
		}
		sleepadOcc := e.sleepadInBed && now-e.sleepadAt <= bedPresenceFreshMs
		radarOcc := e.radarInBed && e.radarAt > 0 && now-e.radarAt <= bedPresenceFreshMs
		if sleepadOcc || radarOcc {
			n++
		}
	}
	return n
}
