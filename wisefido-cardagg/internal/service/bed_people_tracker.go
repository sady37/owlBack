// bed_people_tracker.go — per bed-bound radar device 的 PeopleCount in-memory tracker。
//
// 数据流（doc/card_display.md §4.4 修订版）：
//
//   radar firmware type=3 number_people event
//        ↓ iot:event:stream（不是 monitor 流：qinglan radar_decoder.go:578 type=3 走 event）
//   cardagg EventHandler filter category=number_people
//        ↓ Update(deviceAddr, count, ts)
//   BedPeopleTracker per-device snapshot
//        ↓ CardPeopleCount(cardID, nowMs)
//   VisitorDeriver bed-level path
//        ↓ ≥2 累加 segment / <2 reset
//   write /96 bed card target.visitor_*
//
// 过滤：仅 offline 过滤（DeviceStatusTracker.IsOnline 注入）。
// 不做时间窗 staleness——number_people 是 event-driven（firmware 变化时才发），
// 静态 2 人不变就不会再报，加 staleness 窗口反而会误 reset 真实持续 visitor。
package service

import (
	"context"
	"net/netip"
	"strings"
	"sync"
)

type BedPeopleTracker struct {
	metaCache *DeviceMetaCache
	online    DeviceOnlineChecker

	mu        sync.RWMutex
	perDevice map[string]bedPeopleDeviceState // key = device addr canonical IPv6 string
}

type bedPeopleDeviceState struct {
	peopleCount int
	updatedAt   int64 // ms
}

func NewBedPeopleTracker(metaCache *DeviceMetaCache) *BedPeopleTracker {
	return &BedPeopleTracker{
		metaCache: metaCache,
		perDevice: make(map[string]bedPeopleDeviceState),
	}
}

// SetOnlineChecker main wiring 注入 device 在线判定 callback；未注入时不做 offline 过滤。
func (t *BedPeopleTracker) SetOnlineChecker(fn DeviceOnlineChecker) {
	if t == nil {
		return
	}
	t.online = fn
}

// Update 收到 NumberPeople event 时写入 device snapshot。
// out-of-order（tsMs < 已存 updatedAt）丢弃，避免乱序回放污染当前态。
func (t *BedPeopleTracker) Update(deviceAddr string, peopleCount int, tsMs int64) {
	if deviceAddr == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	cur, ok := t.perDevice[deviceAddr]
	if ok && tsMs < cur.updatedAt {
		return
	}
	t.perDevice[deviceAddr] = bedPeopleDeviceState{peopleCount: peopleCount, updatedAt: tsMs}
}

// ForgetDevice 解绑 / Cleanup 调用。
func (t *BedPeopleTracker) ForgetDevice(deviceAddr string) {
	if deviceAddr == "" {
		return
	}
	t.mu.Lock()
	delete(t.perDevice, deviceAddr)
	t.mu.Unlock()
}

// CardPeopleCount 返回 cardID（/96 bed card）下所有 bed-bound radar device 的 max people count。
// offline device 不参与；从未上报 number_people 的 device 也不参与（视作 0）。
//
// max 而非 sum：share 双床场景下，两 radar boundary 不重叠——任一 radar 见 2 人即视作"该 bed 卡 2 人"。
// 单床多 radar 是冗余安装，max 也是正确合并。
func (t *BedPeopleTracker) CardPeopleCount(ctx context.Context, cardID string) int {
	meta := t.metaCache.GetOrLoad(ctx, cardID)
	if meta == nil || !meta.HasBedBoundRadar() || meta.BedPref == "" {
		return 0
	}
	bedNet, err := netip.ParsePrefix(meta.BedPref)
	if err != nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	max := 0
	for _, dm := range meta.Devices {
		if dm == nil || !dm.DeviceAddr.IsValid() {
			continue
		}
		if !strings.Contains(strings.ToLower(dm.DeviceType), "radar") {
			continue
		}
		if !bedNet.Contains(dm.DeviceAddr) {
			continue
		}
		addrStr := dm.AddrStr()
		st, ok := t.perDevice[addrStr]
		if !ok {
			continue
		}
		if t.online != nil && !t.online(addrStr) {
			continue
		}
		if st.peopleCount > max {
			max = st.peopleCount
		}
	}
	return max
}
