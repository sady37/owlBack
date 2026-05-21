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
	Tracks     map[string]*TrackBuffer // track_id 字符串, e.g. "0"~"11"（88 不入库）
	LastTs     int64                   // ms, 该 device 下任意 track 最大 LastTs，用于 device 级在线判定
	DeviceType string                  // "Radar" / "Sleepad" / "SleepPad" — 用于下游过滤（sustain 只信 sleepad）
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

// Write updates fields for a card:device:track. Only fields present in the
// incoming map are touched; each field gets the message timestamp.
//
// deviceType 可空（兼容旧 caller / 测试），传入时记录到 DeviceBuffer 供下游 vital_source
// 等按 device_type 过滤（如 "sleepad" 才作 bed sustain，"radar" 不作）。
func (b *MonitorBuffer) Write(cardID, deviceID, deviceType, trackID string, fields map[string]any, ts int64) {
	if cardID == "" || deviceID == "" || len(fields) == 0 {
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

	db := cb.Devices[deviceID]
	if db == nil {
		db = &DeviceBuffer{Tracks: make(map[string]*TrackBuffer)}
		cb.Devices[deviceID] = db
	}
	if deviceType != "" {
		db.DeviceType = deviceType
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
	DeviceID   string                    // 业务主 key（与 buffer key 一致）
	DeviceUID  string                    // 仅 log 等用
	DeviceType string                    // "Radar" / "Sleepad" / "SleepPad" — vital_source 用于 sustain 过滤
	Tracks     map[string]map[string]any // track_id -> { ...fields..., "ts": maxTs }
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

// SnapshotCard 返回指定 card 当前 buffer 快照（只读，供告警日志等使用）。
func (b *MonitorBuffer) SnapshotCard(cardID string) *CardSnapshot {
	if cardID == "" {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	cb := b.cards[cardID]
	if cb == nil || len(cb.Devices) == 0 {
		return nil
	}
	var devSnaps []DeviceSnapshot
	for devID, db := range cb.Devices {
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
		devSnaps = append(devSnaps, DeviceSnapshot{DeviceID: devID, DeviceUID: devID, DeviceType: db.DeviceType, Tracks: tracks})
	}
	if len(devSnaps) == 0 {
		return nil
	}
	return &CardSnapshot{CardID: cardID, Devices: devSnaps}
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
		for devID, db := range cb.Devices {
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
			devSnaps = append(devSnaps, DeviceSnapshot{DeviceID: devID, DeviceUID: devID, DeviceType: db.DeviceType, Tracks: tracks})
		}
		if len(devSnaps) > 0 {
			result = append(result, CardSnapshot{CardID: cardID, Devices: devSnaps})
		}
	}
	return result
}

// FlushAndPrune 先 PruneFields 再 Flush。兼容旧用法；主循环建议用 秒数%4==0 时 PruneFields + 每秒 Flush。
func (b *MonitorBuffer) FlushAndPrune(nowMs, fieldTTLMs int64) []CardSnapshot {
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
	return b.flushLocked(nowMs)
}

// PruneStaleDevices removes devices (and empty cards) where (nowMs - db.LastTs) > deviceTTLMs.
// Intended to run periodically (e.g. every 90s); devices that go offline never hit Write again.
func (b *MonitorBuffer) PruneStaleDevices(nowMs, deviceTTLMs int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for cardID, cb := range b.cards {
		for devID, db := range cb.Devices {
			if nowMs-db.LastTs > deviceTTLMs {
				delete(cb.Devices, devID)
			}
		}
		if len(cb.Devices) == 0 {
			delete(b.cards, cardID)
		}
	}
}

// RemoveDevice removes all fields for a specific device from the buffer.
// Called when a message exceeds the hard stale timeout.
func (b *MonitorBuffer) RemoveDevice(cardID, deviceID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	cb := b.cards[cardID]
	if cb == nil {
		return
	}
	delete(cb.Devices, deviceID)
	if len(cb.Devices) == 0 {
		delete(b.cards, cardID)
	}
}

// ActiveCardIDs returns card IDs currently in the buffer.
func (b *MonitorBuffer) ActiveCardIDs() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ids := make([]string, 0, len(b.cards))
	for cid := range b.cards {
		ids = append(ids, cid)
	}
	return ids
}

// ActiveDevicesByCard returns cardID → deviceID set based on device-level
// entries in the buffer (survives 90s device TTL, independent of field TTL).
func (b *MonitorBuffer) ActiveDevicesByCard() map[string]map[string]bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]map[string]bool, len(b.cards))
	for cid, cb := range b.cards {
		if len(cb.Devices) == 0 {
			continue
		}
		devs := make(map[string]bool, len(cb.Devices))
		for did := range cb.Devices {
			devs[did] = true
		}
		result[cid] = devs
	}
	return result
}

