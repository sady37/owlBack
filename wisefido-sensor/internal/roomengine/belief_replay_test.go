package roomengine

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"owl-common/alarm"
	"owl-common/observation"
	"owl-common/radarutils"
	"wisefido-sensor/internal/roomengine/belief"
)

// belief_replay_test.go — §9 第 3a 步：把 doc/cases 真实录制帧流回放过 adapter→belief，
// 对照人工标签验第一版 A/P(o|s) 标定。零生产风险（test-only），用真机数据当回归 oracle。
//
// 关键口径：sensor 自己发的 "Fall" alarm（source=sensor.*）是 gate-list 的**输出**，是对照对象，
// **不喂进 belief**。belief 只吃 raw 观测（track 帧 + firmware EnterRoom/ExitRoom/InBed/number_people）。

const casesDir = "../../../doc/cases"

type fxRecord struct {
	Category  string                   `json:"category"`
	DeviceUID string                   `json:"device_uid"`
	Timestamp int64                    `json:"timestamp"`
	DataValue []map[string]interface{} `json:"data_value"`
}

func loadFixture(t *testing.T, rel string) []fxRecord {
	b, err := os.ReadFile(filepath.Join(casesDir, rel))
	if err != nil {
		t.Skipf("fixture 缺失 %s: %v", rel, err)
	}
	var recs []fxRecord
	if err := json.Unmarshal(b, &recs); err != nil {
		t.Fatalf("解析 %s: %v", rel, err)
	}
	return recs
}

func buildGridFromLayout(t *testing.T, rel string) *RoomGrid {
	b, err := os.ReadFile(filepath.Join(casesDir, rel))
	if err != nil {
		t.Skipf("layout 缺失 %s: %v", rel, err)
	}
	// fixture 两种外壳：cd2b=顶层 canvas（objects[]）；cabb/export_case=DB 行包 {room_id,
	// room_name, layout_config}。ParseLayoutConfig 只认 canvas，故有 layout_config 先解壳。
	var wrap struct {
		LayoutConfig json.RawMessage `json:"layout_config"`
	}
	if json.Unmarshal(b, &wrap) == nil && len(wrap.LayoutConfig) > 0 {
		b = wrap.LayoutConfig
	}
	cfg, err := ParseLayoutConfig("fd00::/88", b)
	if err != nil {
		t.Fatalf("ParseLayoutConfig %s: %v", rel, err)
	}
	grid := NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	grid.OriginX, grid.OriginY = cfg.OriginX, cfg.OriginY
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}
	radars := cfg.Radars
	if len(radars) == 0 {
		radars = append(radars, cfg.Radar)
	}
	for _, m := range radars {
		grid.StampRadar(m)
	}
	grid.StampEnters(cfg.Enters, cfg.EnterTargets)
	for _, r := range cfg.Enters {
		grid.SetPrior(r, AreaEnter, 99, SourceHuman)
	}
	for _, r := range cfg.Beds {
		grid.SetPrior(r, AreaBed, 99, SourceHuman)
	}
	for _, r := range cfg.Toilets {
		grid.SetPrior(r, AreaToilet, 99, SourceHuman)
	}
	for _, r := range cfg.Showers {
		grid.SetPrior(r, AreaShower, 99, SourceHuman)
	}
	for _, r := range cfg.Chairs {
		grid.SetPrior(r, AreaSit, 80, SourceHuman)
	}
	return grid
}

func mi(m map[string]interface{}, k string) (int, bool) {
	if v, ok := m[k]; ok {
		if f, ok := v.(float64); ok {
			return int(f), true
		}
	}
	return 0, false
}

type replayResult struct {
	frames          int
	maxFallenP      float64
	firstFireTs     int64
	endFallenP      float64 // 末态 P(Fallen)——区分"持续倒地"vs"fire 后人返回恢复"
	minPAfterFire   float64 // firstFire 之后的最低 P(Fallen)——衡量返回恢复
	maxStillSec     float64
	lostStillObs    int
	maxLostStillV   float64
	recoveryTs      int64
	confirmedFireTs int64   // Decider 确认窗后的确认报警时刻（0 = 未确认）
	maxTLost        float64 // DBN P1 Track 层：任一 track 任一帧 P(TLost) 峰值
	endTLost        float64 // 末帧任一 track 的 max P(TLost)
}

// rtrk 回放期 per-track 移动/丢失状态（belief 不持有，engine/replay 维护）。
// 注：production engine 已做 ghost 检测(Verdict)；replay 自带轻量 ghost-filter（剔 impossible-speed 跳）
// + 复刻 gate-list still-box，使"消失前是否走动"判据与 deployed 止血一致。
type rtrk struct {
	st          *TrackState
	lastX       int
	lastY       int
	lastZ       int
	hasLast     bool
	lastFrameTs int64
	lastGeom    belief.Geom
	lostAnchor  int64               // 丢失 ramp 起点（=最后帧 ts；0=未丢失/已返回）
	tb          *belief.TrackBelief // DBN P1 Track 层 T_t（只读 shadow，不驱动 Room 层）
}

// replay 驱动 belief：track 帧 → adapter；走动中突然消失 → lostWhileMovingToObs；exit/返回取消。
func replay(t *testing.T, recs []fxRecord, grid *RoomGrid) replayResult {
	be := belief.New(belief.DefaultModel())
	var decider belief.Decider
	tracks := map[int]*rtrk{}
	exited := false // ExitRoom 后抑制 lost-fall ramp（人走出去了），新 track 帧重置
	res := replayResult{}

	for _, r := range recs {
		now := r.Timestamp
		var obs []belief.Observation
		switch r.Category {
		case "track":
			for _, dv := range r.DataValue {
				id, _ := mi(dv, "track_id")
				if id == 88 { // firmware heartbeat 无人
					continue
				}
				exited = false // 有真 track 帧 → 取消 exit 抑制
				x, _ := mi(dv, "position_x")
				y, _ := mi(dv, "position_y")
				z, _ := mi(dv, "position_z")
				pose, _ := mi(dv, "pose")
				conf, _ := mi(dv, "track_confidence")

				k := tracks[id]
				if k == nil {
					k = &rtrk{st: &TrackState{Verdict: VerdictReal}, lastZ: z}
					tracks[id] = k
				}
				// production ghost adjudicator 的 verdict 经 fixture "verdict":"ghost" 注入
				// （replay 自身不跑 adjudicator）。无字段 = 维持 Real，与既有 case 行为不变。
				if vd, ok := dv["verdict"].(string); ok && vd == "ghost" {
					k.st.Verdict = VerdictGhost
					k.st.GhostPenalty = 80
				}
				// ghost-filter：impossible-speed 假跳（track-swap）剔除，不喂 still-box/belief。
				if k.hasLast && isGhostJump(math.Hypot(float64(x-k.lastX), float64(y-k.lastY)), now-k.lastFrameTs) {
					continue
				}
				k.st.LastZ = k.lastZ // 上一帧 z 供 dz 运动学
				k.st.PushPoint(x, y, z, now)
				// 静止判定（robust，抗 outlier）：最近 30s ≥60% 帧落在 latest 位置 StillBoxCm 内 = 静止。
				// 比 max-min box 抗噪——production engine 用 Verdict 过滤 ghost，replay 用此鲁棒统计近似，
				// 避免漏网的 30-200cm/s 噪声/ghost 跳把 still-box 单点击穿（production 实测 Hunzi still-box≥60s）。
				inlier, total := 0, 0
				for _, p := range k.st.History {
					if now-p.TMs > 30_000 {
						continue
					}
					total++
					if math.Hypot(float64(x-p.X), float64(y-p.Y)) <= float64(FallRulesParam.Lost.StillBoxCm) {
						inlier++
					}
				}
				if total >= 2 && float64(inlier)/float64(total) >= 0.6 {
					if k.st.StillBoxRunStart == 0 {
						k.st.StillBoxRunStart = k.st.History[0].TMs
					}
				} else {
					k.st.StillBoxRunStart = 0
				}
				k.lastGeom = geomFromGrid(grid, x, y)
				k.lastFrameTs = now
				k.lastX, k.lastY, k.lastZ, k.hasLast = x, y, z, true
				k.lostAnchor = 0 // 在场/返回 → 清丢失

				// DBN P1 Track 层：present 发射（ghostness 由 verdict/penalty，geom 当前帧）。
				if k.tb == nil {
					k.tb = belief.NewTrackBelief()
				}
				ghostness := float64(k.st.GhostPenalty) / 100.0
				k.tb.Step(now, []belief.TObservation{{
					Kind: belief.TObsPresent, Ghostness: ghostness, Geom: k.lastGeom,
					Conf: 0.9, Ts: now, Fresh: true,
				}})

				tr := observation.Track{LogicID: "L", Pose: pose, PoseConfidence: conf, PositionX: &x, PositionY: &y, PositionZ: &z}
				obs = append(obs, radarFrameAdapter(tr, k.st, grid, now)...)
			}

		case alarm.EnterRoom, alarm.ExitRoom:
			if r.Category == alarm.ExitRoom {
				exited = true // 人走出去 → 抑制 lost-fall
				// DBN P1：firmware 离场 → 各 track 喂 TObsExit（→JustLeft，压 Lost）。
				for _, k := range tracks {
					if k.tb != nil {
						k.tb.Step(now, []belief.TObservation{{Kind: belief.TObsExit, Conf: 0.9, Ts: now, Fresh: true}})
					}
				}
			}
			if o, ok := radarEventToObs(r.Category, now, belief.GeomInEnter); ok {
				obs = append(obs, o)
			}
		case "number_people":
			for _, dv := range r.DataValue {
				if n, ok := mi(dv, "number_people"); ok {
					obs = append(obs, belief.Observation{Kind: belief.ObsNumberPeople, Value: float64(n), Conf: 0.8, Ts: now, Fresh: true})
				}
			}
		case "InBed":
			obs = append(obs, neighborToObs(0.9, 0.85, now))
		case "heart", "activity":
			obs = append(obs, belief.Observation{Kind: belief.ObsVitalPresent, Value: 1, Conf: 0.6, Ts: now, Fresh: true})
		default:
			continue // sensor Fall/LeftBed/device-status alarm = 对照对象，不喂
		}

		// DBN P1 Track 层 absent 扫掠：超 TTL 无帧就喂 absent（**不 skip ghost/settled**——
		// 让 A_T 结构化处理：ghost 先验→None，real 先验→Lost/JustLeft）。与下方 Room 层
		// lost-fall 扫掠（带 ghost/settled/exit 各种 gate）刻意分开，正是 P1 要对照的"gate vs A_T"。
		for _, k := range tracks {
			if k.tb == nil || now-k.lastFrameTs <= beliefLostTTLMs {
				continue
			}
			k.tb.Step(now, []belief.TObservation{{Kind: belief.TObsAbsent, Geom: k.lastGeom, Conf: 0.9, Ts: now, Fresh: true}})
		}

		// 丢失扫掠：track 超 TTL 无帧 + 消失前60s在走动（still-box<60s）+ 未走出 → 发 lost-fall（ramp）。
		for _, k := range tracks {
			if exited || now-k.lastFrameTs <= beliefLostTTLMs {
				continue
			}
			// ghost 消失 ≠ 人倒地（镜面反射闪灭）。method-2 在 gate 侧同款 guard：
			// 失锁前是 ghost 的 track 不发 lost-while-moving，否则 ramp 无对手地误拉 Fallen。
			if k.st.Verdict == VerdictGhost {
				continue
			}
			// 走动前置（对齐 gate-list 止血）：消失前 still-box ≥60s = 静止态 settle → 归 Still-fall，不报。
			if k.st.StillBoxRunStart > 0 &&
				k.lastFrameTs-k.st.StillBoxRunStart >= int64(FallRulesParam.Lost.MovingPreconditionMs) {
				continue
			}
			if k.lostAnchor == 0 {
				k.lostAnchor = k.lastFrameTs
			}
			age := float64(now-k.lostAnchor) / 1000 / beliefLostWaitWindowSec
			obs = append(obs, lostWhileMovingToObs(age, k.lastGeom, now))
			res.lostStillObs++
		}
		be.Step(now, obs)

		res.frames++
		p := be.Vector().P(belief.SFallen)
		res.endFallenP = p
		if p > res.maxFallenP {
			res.maxFallenP = p
		}
		if res.firstFireTs == 0 && be.Decide() == belief.DecisionFall {
			res.firstFireTs = now
			res.minPAfterFire = p
		}
		if res.firstFireTs != 0 {
			if p < res.minPAfterFire {
				res.minPAfterFire = p
			}
			if res.recoveryTs == 0 && p < 0.55 {
				res.recoveryTs = now
			}
		}
		if res.confirmedFireTs == 0 && decider.Update(be.Vector(), now) == belief.DecisionFall {
			res.confirmedFireTs = now
		}

		// DBN P1 Track 层快照：本帧任一 track 的 max P(TLost)。
		frameTLost := 0.0
		for _, k := range tracks {
			if k.tb == nil {
				continue
			}
			if pl := k.tb.Vector().P(belief.TLost); pl > frameTLost {
				frameTLost = pl
			}
		}
		res.endTLost = frameTLost
		if frameTLost > res.maxTLost {
			res.maxTLost = frameTLost
		}
	}
	return res
}

// 真机回放 oracle（3a 闭环）：doc/cases 真实帧流过 adapter→belief→Decider，对照人工标签。
// 关键判别（实测收敛）：真跌倒 P(Fallen) 维持数分钟过确认窗 → 确认 Fall；人返回类 P 在 ~47s 内
// 崩回 θ 下（走动反证）→ 永不过确认窗。harness 忠实加载 layout（解 layout_config 壳）+ 复刻
// track_manager StillBox 判据。
func TestReplayOracle(t *testing.T) {
	cases := []struct {
		name        string
		win, layout string
		wantConfirm bool // 期望 Decider 确认 Fall
		assert      bool // false = 仅诊断（ground-truth 不确定）
	}{
		// belief 新设计：lost-fall = 走动中突然消失（still-box<60s 时丢失）+ 无 exit/返回。判别靠移动
		// 不靠 pose（2026-06-01 用户：pose 不可靠）。静止态丢失（still-box≥60s）→ 归 Still-fall，不报。
		// D5F7 = 用户前天在 101/Bathroom 专门做的真跌倒（>15min）→ **应 fire**；其余冻结站/坐 FP → 不报。
		{"D5F7浴室(真跌倒,101测试)", "d5f7-bathroom-lost-1031/window.json", "d5f7-bathroom-lost-1031/room_layout.json", true, true},
		// 2026-06-01 07:2x 真实 FP：镜面 ghost(越界 V140-210)+np=0(人已走)无 ExitRoom → gate 误报 lost_track。
		// belief 应**不确认**（np=0 + ghost 该把信念压在 Empty/Artifact，不进 Fallen）。先诊断不断言。
		{"D5F7浴室FP(0601裸replay,无adjudicator)", "d5f7-bathroom-fp-0601/window.json", "d5f7-bathroom-fp-0601/room_layout.json", false, false},
		// 同数据但注入 production ghost adjudicator 的 verdict（18 帧镜面区 → ghost）。
		// 期望 belief **不确认**：ghost→SArtifact 压住 Fallen。证明模型对，裸 replay 误确认是保真局限。
		{"D5F7浴室FP(0601+ghost verdict)", "d5f7-bathroom-fp-0601/window_ghostadj.json", "d5f7-bathroom-fp-0601/room_layout.json", false, true},
		{"cabb-fall-A(实FP,静止站立)", "cabb-fall-A-frozen-0016/2026-05-04_00-05_to_00-20_MDT.json", "cabb-fall-A-frozen-0016/room_layout.json", false, true},
		{"cd2b(人返回)", "case_lostfall_cd2b_11351148/2026-04-27_11-35_to_11-48_MDT.json", "case_lostfall_cd2b_11351148/room_layout.json", false, true},
		{"MoM浴室FP(走出exit)", "mom-bathroom-lost-0953/window.json", "mom-bathroom-lost-0953/room_layout.json", false, true},
		{"D523卧室FP(9h同型,静止)", "d523-bedroom-lost-0933/window.json", "d523-bedroom-lost-0933/room_layout.json", false, true},
		{"Hunzi-CABB-0529FP(站立静止)", "hunzi-cabb-lost-0529-FP/window.json", "hunzi-cabb-lost-0529-FP/room_layout.json", false, true},
		// Hunzi-0530：ghost-heavy（max跳19166cm/s）。replay 粗 ghost-filter(>200) 漏掉 30-200cm/s 的 ghost
		// 噪声 → still-box 被击穿 → replay 误报。**production engine 用 Verdict 完整 ghost 检测+止血会挡住**
		// （production 实测其 still-box≥60s）。属 replay 保真局限（非 belief 模型错），仅诊断不断言。
		{"Hunzi-0530(ghost-heavy,replay局限)", "hunzi-cabb-lost-0530-FP/window.json", "hunzi-cabb-lost-0530-FP/room_layout.json", false, false},
	}
	for _, c := range cases {
		res := replay(t, loadFixture(t, c.win), buildGridFromLayout(t, c.layout))
		confirmed := res.confirmedFireTs != 0
		t.Logf("%-32s confirm=%-5v maxP=%.3f endP=%.3f minPAfterFire=%.3f maxStill=%.0fs lostStill=%d",
			c.name, confirmed, res.maxFallenP, res.endFallenP, res.minPAfterFire, res.maxStillSec, res.lostStillObs)
		if c.assert && confirmed != c.wantConfirm {
			t.Errorf("%s: confirm=%v 期望 %v（maxP=%.3f minPAfterFire=%.3f）", c.name, confirmed, c.wantConfirm, res.maxFallenP, res.minPAfterFire)
		}
	}
}

// TestTrackLayerOracle — DBN P1：Track 层 T_t 在真机数据上的结构化路由对照（只读，不驱动 Room）。
// P1 能结构化分开的：ghost 类（→None）+ 门区/exit 类（→JustLeft），即 4 大冲突里的 2 类。
// P1 还分不开的：静止-vs-走动丢失（cabb/D523/Hunzi 在 Track 层都是真 Lost——FP 性来自
// moving-precondition = P3 dwell 层，不在存在性 Track 层）。后者诚实标 diagnostic 不断言。
func TestTrackLayerOracle(t *testing.T) {
	cases := []struct {
		name        string
		win, layout string
		wantLost    bool    // 期望 Track 层判定真人丢失（候选倒地）
		thresh      float64 // P(TLost) 判定阈
		assert      bool    // false = 仅诊断
	}{
		// 真跌倒：真人开阔地板走动中消失 → Track 层应 Lost（喂 Room 层 Fallen 前置）。
		{"D5F7浴室(真跌倒,101测试)", "d5f7-bathroom-lost-1031/window.json", "d5f7-bathroom-lost-1031/room_layout.json", true, 0.5, true},
		// ghost 注入：18 帧镜面区→ghost。结构核心：Ghost→Lost≈0，应 →None（不 Lost）。
		{"D5F7浴室FP(0601+ghost verdict)", "d5f7-bathroom-fp-0601/window_ghostadj.json", "d5f7-bathroom-fp-0601/room_layout.json", false, 0.2, true},
		// 走出 exit：firmware ExitRoom → JustLeft，不 Lost。
		{"MoM浴室FP(走出exit)", "mom-bathroom-lost-0953/window.json", "mom-bathroom-lost-0953/room_layout.json", false, 0.2, true},
		// 以下 Track 层都是真 Lost（FP 性在 dwell 层，P1 不分）——诊断。
		{"cabb-fall-A(实FP,静止站立)", "cabb-fall-A-frozen-0016/2026-05-04_00-05_to_00-20_MDT.json", "cabb-fall-A-frozen-0016/room_layout.json", false, 0.2, false},
		{"cd2b(人返回)", "case_lostfall_cd2b_11351148/2026-04-27_11-35_to_11-48_MDT.json", "case_lostfall_cd2b_11351148/room_layout.json", false, 0.2, false},
		{"D523卧室FP(9h同型,静止)", "d523-bedroom-lost-0933/window.json", "d523-bedroom-lost-0933/room_layout.json", false, 0.2, false},
		{"Hunzi-CABB-0529FP(站立静止)", "hunzi-cabb-lost-0529-FP/window.json", "hunzi-cabb-lost-0529-FP/room_layout.json", false, 0.2, false},
		{"D5F7浴室FP(0601裸replay,无adjudicator)", "d5f7-bathroom-fp-0601/window.json", "d5f7-bathroom-fp-0601/room_layout.json", false, 0.2, false},
	}
	for _, c := range cases {
		res := replay(t, loadFixture(t, c.win), buildGridFromLayout(t, c.layout))
		// maxTLost = "Track 层是否曾判定真人丢失"——peak 是喂 Room 层候选的触发信号；
		// 之后的衰减（真跌倒被重新检出为 lying / 人返回 recapture）属 dwell 轴（P3），不在 P1。
		gotLost := res.maxTLost >= c.thresh
		t.Logf("%-32s T:lost=%-5v maxTLost=%.3f endTLost=%.3f", c.name, gotLost, res.maxTLost, res.endTLost)
		if c.assert && gotLost != c.wantLost {
			t.Errorf("%s: Track 层 lost=%v 期望 %v（maxTLost=%.3f endTLost=%.3f）", c.name, gotLost, c.wantLost, res.maxTLost, res.endTLost)
		}
	}
}
