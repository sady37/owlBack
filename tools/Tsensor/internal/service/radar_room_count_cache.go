package service

// radar_room_count_cache.go — per-/88 room 的 radar NumberPeople 缓存。
//
// 用途：BedPresenceFusion dedup 公式中的 Z（radar room 总数）。BedZone subset_invariant
// 会把 room.Count 从 0 抬到 1（sleepad InBed 但 radar Z=0 的瞬时态），让 e.NewState.Count
// 失去"radar 原值"语义；dedup 必须用 raw Z 否则双计数。
//
// 与 [[bed_presence_fusion]] 配套：adapter_radar applyCount 写入 cache；
// stream_publisher OnZoneEvent ZoneTypeRoom 读 cache 得 Z，再 max OccupiedBedsInRoom 即 publish 值。

import (
	"sync"
	"time"
)

// RadarRoomCountCache 跟踪 per-/88 room 最近一次 radar NumberPeople。
type RadarRoomCountCache struct {
	mu      sync.RWMutex
	entries map[string]*radarRoomEntry
	now     func() int64
}

type radarRoomEntry struct {
	count     int
	updatedAt int64
}

// radarRoomFreshMs radar NumberPeople 心跳容忍：超 5min 未上报视作 stale，GetZ 返 ok=false。
// 5min == radar 1Hz monitor 心跳 N 次 + 容忍 firmware 上行抖动；超期视作设备失联，让上层
// 走 fallback（用 e.NewState.Count）。
const radarRoomFreshMs = 5 * 60 * 1000

func NewRadarRoomCountCache() *RadarRoomCountCache {
	return &RadarRoomCountCache{
		entries: make(map[string]*radarRoomEntry),
		now:     func() int64 { return time.Now().UnixMilli() },
	}
}

// SetZ adapter_radar applyCount 调；ts=0 时取 now。count 即 radar 上报的 number_people。
func (c *RadarRoomCountCache) SetZ(roomCIDR string, count int, ts int64) {
	if roomCIDR == "" {
		return
	}
	if ts == 0 {
		ts = c.now()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.entries[roomCIDR]
	if e == nil {
		e = &radarRoomEntry{}
		c.entries[roomCIDR] = e
	}
	e.count = count
	e.updatedAt = ts
}

// GetZ 返回 (count, ok)。
//   - ok=true: cache 内有 fresh 值（≤5min）
//   - ok=false: 无 entry 或 stale；caller 应 fallback 到 e.NewState.Count
func (c *RadarRoomCountCache) GetZ(roomCIDR string) (int, bool) {
	if roomCIDR == "" {
		return 0, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[roomCIDR]
	if !ok {
		return 0, false
	}
	if c.now()-e.updatedAt > radarRoomFreshMs {
		return 0, false
	}
	return e.count, true
}
