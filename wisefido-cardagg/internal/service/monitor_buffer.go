package service

import (
	"sync"
)

type FieldValue struct {
	Value any
	Ts    int64 // ms, field last reported time
}

// TrackBuffer 单条 track 的字段缓存。
// track_id 约定见 owl-common/observation: 0-8=人, 9=未知人, 10=空间, 11=设备; 88=无效（Handle 层丢弃不入库）。
type TrackBuffer struct {
	Fields map[string]*FieldValue
	LastTs int64 // ms, 该 track 最后 Write 时间
}

type DeviceBuffer struct {
	Tracks map[string]*TrackBuffer // track_id 字符串, e.g. "0"~"11"（88 不入库）
	LastTs int64                   // ms, 该 device 下任意 track 最大 LastTs，用于 device 级在线判定
}

type CardBuffer struct {
	Devices map[string]*DeviceBuffer
}

// MonitorBuffer 接收 iot:monitor 写入，按 card/device 聚合，供 Flush 派生 card-level 业务（人数/床/区域）。
// device 在线状态不再由本 buffer 推导——已切换为 monitor/event/alarm 流事件驱动 + 看门狗 fail-safe；
// 本结构仅保留 card-level 字段聚合 + GC，PruneStaleDevices/RemoveDevice 仅作内存清理。
type MonitorBuffer struct {
	mu    sync.RWMutex
	cards map[string]*CardBuffer // cardID →
}

func NewMonitorBuffer() *MonitorBuffer {
	return &MonitorBuffer{
		cards: make(map[string]*CardBuffer),
	}
}

// DefaultTrackID 无 track_id 时使用的默认 key（如 Sleepad 等单轨设备）
const DefaultTrackID = "0"

// SnapshotByDevice 返回某 device 下所有 track 的当前字段快照，按 track_id 分组。
// 返回结构：{ "tracks": [ {track_id: "0", last_ts_ms: N, fields: {heart_rate: 65, ...}}, ... ] }
// 用途：alarm 触发时附在 alarm_events.evidence —— 取报警设备最近 monitor 缓存值，
// 不受设备网关限制（sleepace HR/RR、radar position/pose 同一渠道统一处理）。
// 无数据时返回 nil（caller 自行省略 evidence.monitor 字段）。
func (b *MonitorBuffer) SnapshotByDevice(cardID, deviceAddr string) map[string]any {
	if cardID == "" || deviceAddr == "" {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()

	cb := b.cards[cardID]
	if cb == nil {
		return nil
	}
	db := cb.Devices[deviceAddr]
	if db == nil || len(db.Tracks) == 0 {
		return nil
	}

	tracks := make([]map[string]any, 0, len(db.Tracks))
	for tid, tb := range db.Tracks {
		fields := make(map[string]any, len(tb.Fields))
		for k, fv := range tb.Fields {
			fields[k] = fv.Value
		}
		tracks = append(tracks, map[string]any{
			"track_id":   tid,
			"last_ts_ms": tb.LastTs,
			"fields":     fields,
		})
	}
	return map[string]any{
		"device_addr": deviceAddr,
		"tracks":    tracks,
	}
}

// Write updates fields for a card:device:track. Only fields present in the
// incoming map are touched; each field gets the message timestamp.
func (b *MonitorBuffer) Write(cardID, deviceAddr, trackID string, fields map[string]any, ts int64) {
	if cardID == "" || deviceAddr == "" || len(fields) == 0 {
		return
	}
	if trackID == "" {
		trackID = DefaultTrackID
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	cb := b.cards[cardID]
	if cb == nil {
		cb = &CardBuffer{Devices: make(map[string]*DeviceBuffer)}
		b.cards[cardID] = cb
	}

	db := cb.Devices[deviceAddr]
	if db == nil {
		db = &DeviceBuffer{Tracks: make(map[string]*TrackBuffer)}
		cb.Devices[deviceAddr] = db
	}

	tb := db.Tracks[trackID]
	if tb == nil {
		tb = &TrackBuffer{Fields: make(map[string]*FieldValue)}
		db.Tracks[trackID] = tb
	}

	for k, v := range fields {
		tb.Fields[k] = &FieldValue{Value: v, Ts: ts}
	}
	if ts > tb.LastTs {
		tb.LastTs = ts
	}
	if ts > db.LastTs {
		db.LastTs = ts
	}
}

// DeviceSnapshot is the flush output for one device: per-track fields (track_id -> fields+ts).
type DeviceSnapshot struct {
	DeviceAddr string                    // device_addr canonical IPv6 text
	Tracks   map[string]map[string]any // track_id -> { ...fields..., "ts": maxTs }
}

// CardSnapshot is the flush output for one card.
type CardSnapshot struct {
	CardID  string
	Devices []DeviceSnapshot
}

// PruneFields 删除超时的 field：(nowMs - fv.Ts) > fieldTTLMs。建议仅在 秒数%4==0 时调用，与 4s TTL 对齐。
func (b *MonitorBuffer) PruneFields(nowMs, fieldTTLMs int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, cb := range b.cards {
		for _, db := range cb.Devices {
			for _, tb := range db.Tracks {
				for k, fv := range tb.Fields {
					if nowMs-fv.Ts > fieldTTLMs {
						delete(tb.Fields, k)
					}
				}
			}
		}
	}
}

// Flush 仅根据当前 buffer 构建 snapshot，不删除 field。用 RLock，与 Write（Lock）不互斥时写可优先插入。
func (b *MonitorBuffer) Flush(nowMs int64) []CardSnapshot {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.flushLocked(nowMs)
}

func (b *MonitorBuffer) flushLocked(nowMs int64) []CardSnapshot {
	var result []CardSnapshot
	for cardID, cb := range b.cards {
		var devSnaps []DeviceSnapshot
		for devAddr, db := range cb.Devices {
			if len(db.Tracks) == 0 {
				continue
			}
			tracks := make(map[string]map[string]any)
			for trackID, tb := range db.Tracks {
				if len(tb.Fields) == 0 {
					continue
				}
				m := make(map[string]any, len(tb.Fields)+1)
				for k, fv := range tb.Fields {
					m[k] = fv.Value
				}
				m["ts"] = tb.LastTs
				tracks[trackID] = m
			}
			if len(tracks) == 0 {
				continue
			}
			devSnaps = append(devSnaps, DeviceSnapshot{DeviceAddr: devAddr, Tracks: tracks})
		}
		if len(devSnaps) > 0 {
			result = append(result, CardSnapshot{CardID: cardID, Devices: devSnaps})
		}
	}
	return result
}

