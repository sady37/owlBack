package roomengine

// neighborResult 跨房 hand-off 评估结果。hit=fresh 有向命中(压本房 phantom fall);staleOcc=邻房有 durable
// 占用但相关度 stale(事件 ts 在窗外)= 留驻 deferred gap(plan-A 纯相关下不压,须 no-silent-caps LOG 量化 plan-B 价值)。
type neighborResult struct {
	hit      bool
	conf     float64
	src      string
	staleOcc bool
	staleGap int64 // lostSeenMs − 邻房事件ts（>0=占用早于丢轨=留驻；<0=晚于窗）
	staleSrc string
}

// neighborHandoff §5.5.2 跨房弱耦合 + 铁律（非全空间监控）：仅当本房丢轨（lostSeenMs，二义 lost-fall）后
// [−jitter, window] 内同 unit 兄弟房出现**新鲜有向** hand-off 事件（EnterRoom 仍占用 ∨ InBed 翻转）
// = 人这一刻挪去邻房 → hit（压本房 phantom fall）。
//
// 安全约束：
//   - 有向极近窗：邻房事件须晚于本房丢轨（先走后到），仅留 jitter 反向余量；stale/durable「上次在哪」不算 hit
//     （证不了此刻在哪；人可能穿盲区真摔）。窗固定绝对，不随房间距离伸缩。
//   - N-3 sole-resident 门（复用 SoleResidentRecaptureState 计数）：多/无 resident → 邻房占用可能是另一人 →
//     归因不安全 → 不压（漏报-safe）。
//   - N-6 单合并：至多一条 occ，occupancy=room-enter OR bed-InBed，confidence=max（非两条独立似然相乘）。
//   - N-4：多兄弟房任一命中即占用（sole-resident 前提"人只在一处"）。
//   - no-silent-caps（审查62）：无 fresh hit 但邻房有 durable 占用（相关度 stale）→ staleOcc，供 LOG 量化 deferred gap。
func (e *Engine) neighborHandoff(roomID string, lostSeenMs, nowMs int64) neighborResult {
	win := FallRulesParam.Neighbor.HandoffWindowMs
	jit := FallRulesParam.Neighbor.JitterMs

	e.mu.RLock()
	suiteID := e.roomSuiteID[roomID]
	var siblings []*TrackManager
	if suiteID != "" {
		for rid, tm := range e.rooms {
			if rid != roomID && e.roomSuiteID[rid] == suiteID {
				siblings = append(siblings, tm)
			}
		}
	}
	e.mu.RUnlock()

	if suiteID == "" || e.suiteCensus == nil || len(siblings) == 0 {
		return neighborResult{}
	}
	if rc, _ := e.suiteCensus.SoleResidentRecaptureState(suiteID); rc != 1 {
		return neighborResult{} // N-3 gate-OFF：非单 resident → 不压、不记 stale（归因不安全，非本特性 gap）
	}

	inWindow := func(evtMs int64) bool { // 有向 + 极近：邻房事件晚于本房丢轨，仅 jitter 反向余量
		d := evtMs - lostSeenMs
		return d >= -jit && d <= win
	}

	const roomEnterConf = 0.8 // firmware 过门事件可信度
	r := neighborResult{}
	var staleEvtMs int64
	consider := func(evtMs int64, conf float64, src string) {
		if inWindow(evtMs) {
			if conf > r.conf {
				r.hit, r.conf, r.src = true, conf, src
			}
		} else if !r.staleOcc || evtMs > staleEvtMs { // 最近的窗外占用事件（留驻 deferred gap）
			r.staleOcc, staleEvtMs, r.staleSrc = true, evtMs, src
		}
	}
	for _, tm := range siblings {
		if ts, inBed, c := tm.NeighborBedHandoff(nowMs); inBed {
			consider(ts, c, "bed") // 接触式：sleepad 0.9 / radar-only 0.2
		}
		if ts, occupied := tm.NeighborRoomEnterMs(); occupied {
			consider(ts, roomEnterConf, "room")
		}
	}
	if r.hit {
		return neighborResult{hit: true, conf: r.conf, src: r.src} // fresh hit：不再报 stale
	}
	if r.staleOcc {
		r.staleGap = lostSeenMs - staleEvtMs
	}
	return r
}
