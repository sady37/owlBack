package replay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// replay.go — cd2b 三态 replay harness（集成步骤2，最瘦路径，用户裁定 replay harness）。
// 复用已验证三块：adapter（decoupled FrameInput）+ belief（黑盒）+ probe §9。本层只做
//   window.json / window_sleepad.json / room_layout.json → []adapter.FrameInput 解析 + 驱动。
// 不碰 DB、不克隆引擎（Tsensor playback 是 DB 耦合，不可复用作 harness）。
// AC-1 归因边界：belief 已验证，若 cd2b 不分离/不 fire，缺陷在 adapter 派生或本解析，不在 belief。

const sleepadTTLMs = 35_000 // sleepad 离线判定窗（与生产 beliefSleepadTTLMs 一致）

// ---- fixture JSON 结构 ----

type radarFile []struct {
	DeviceUID string `json:"device_uid"`
	Timestamp int64  `json:"timestamp"`
	DataValue []struct {
		Pose      int `json:"pose"`
		TrackID   int `json:"track_id"`
		PositionX int `json:"position_x"`
		PositionY int `json:"position_y"`
		PositionZ int `json:"position_z"`
	} `json:"data_value"`
}

type sleepadFile []struct {
	Timestamp int64 `json:"timestamp"`
	DataValue []struct {
		BedStatus       int `json:"bed_status"`
		HeartRate       int `json:"heart_rate"`
		RespiratoryRate int `json:"respiratory_rate"`
	} `json:"data_value"`
}

type layoutFile struct {
	Objects []struct {
		TypeValue int `json:"typeValue"`
		Geometry  struct {
			Data struct {
				Vertices []struct {
					X float64 `json:"x"`
					Y float64 `json:"y"`
				} `json:"vertices"`
			} `json:"data"`
		} `json:"geometry"`
	} `json:"objects"`
}

func loadJSON(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// Beds 从 layout 抽 typeValue==2(Bed) 的矩形（顶点 min/max，canvas cm）。
func Beds(layoutPath string) ([]adapter.Rect, error) {
	var lf layoutFile
	if err := loadJSON(layoutPath, &lf); err != nil {
		return nil, err
	}
	var rects []adapter.Rect
	for _, o := range lf.Objects {
		if o.TypeValue != 2 { // 2=Bed
			continue
		}
		vs := o.Geometry.Data.Vertices
		if len(vs) == 0 {
			continue
		}
		x1, y1, x2, y2 := vs[0].X, vs[0].Y, vs[0].X, vs[0].Y
		for _, v := range vs[1:] {
			x1, x2 = min(x1, v.X), max(x2, v.X)
			y1, y2 = min(y1, v.Y), max(y2, v.Y)
		}
		rects = append(rects, adapter.Rect{X1: int(x1), Y1: int(y1), X2: int(x2), Y2: int(y2)})
	}
	return rects, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// sleepadAt 取 ≤ now 的最近 sleepad 帧 + 是否 fresh。
type sleepadState struct {
	ts        int64
	bedStatus int
	hasFrame  bool
}

// BuildTimeline 装一段 cd2b replay 时间线：cd2b 雷达帧为主 tick，sleepad 按 TTL 合入，床矩形从 layout。
// census：单住户独处（cd2b 场景），PeopleCount=1，AloneContinuousMin 随时长增长，Night=false（保守）。
func BuildTimeline(caseDir, radarUID string) ([]adapter.FrameInput, error) {
	var rf radarFile
	if err := loadJSON(filepath.Join(caseDir, "window.json"), &rf); err != nil {
		return nil, err
	}
	var sf sleepadFile
	if err := loadJSON(filepath.Join(caseDir, "window_sleepad.json"), &sf); err != nil {
		return nil, err
	}
	beds, err := Beds(filepath.Join(caseDir, "room_layout.json"))
	if err != nil {
		return nil, err
	}

	// sleepad 帧按 ts 排序。
	sort.Slice(sf, func(i, j int) bool { return sf[i].Timestamp < sf[j].Timestamp })
	sleepadBefore := func(now int64) sleepadState {
		st := sleepadState{}
		for _, s := range sf {
			if s.Timestamp > now {
				break
			}
			if len(s.DataValue) > 0 {
				st = sleepadState{ts: s.Timestamp, bedStatus: s.DataValue[0].BedStatus, hasFrame: true}
			}
		}
		return st
	}

	// cd2b 雷达帧（过滤 + 排序）。
	type rframe struct {
		ts  int64
		dv0 struct{ pose, x, y, z int }
	}
	var radar []rframe
	for _, f := range rf {
		if f.DeviceUID != radarUID || len(f.DataValue) == 0 {
			continue
		}
		d := f.DataValue[0]
		radar = append(radar, rframe{ts: f.Timestamp, dv0: struct{ pose, x, y, z int }{d.Pose, d.PositionX, d.PositionY, d.PositionZ}})
	}
	sort.Slice(radar, func(i, j int) bool { return radar[i].ts < radar[j].ts })
	if len(radar) == 0 {
		return nil, nil
	}

	covers := []float64{1}
	onbed := []float64{1}
	overlap := []float64{1}
	startMs := radar[0].ts

	out := make([]adapter.FrameInput, 0, len(radar))
	for _, r := range radar {
		ss := sleepadBefore(r.ts)
		fresh := ss.hasFrame && r.ts-ss.ts <= sleepadTTLMs
		inBed := ss.bedStatus == 0 // 0=InBed
		out = append(out, adapter.FrameInput{
			NowMs: r.ts,
			Track: adapter.RadarTrack{
				Online: true, Pose: r.dv0.pose,
				X: r.dv0.x, Y: r.dv0.y, Z: r.dv0.z,
				HR: 0, RR: 0, // 雷达帧无 vital 字段（cd2b：radar enter-gate 不返 HR/RR）
			},
			Sleepads: []adapter.SleepadFrame{{InBed: inBed, Fresh: fresh}},
			Beds:     beds,
			Covers:   covers, Onbed: onbed, Overlap: overlap,
			Census: adapter.Census{
				AloneContinuousMin: float64(r.ts-startMs) / 60000.0,
				PeopleCount:        1,
			},
		})
	}
	return out, nil
}

// Frame replay 单帧结果（probe + 裁决）。
type Frame struct {
	Probe    belief.FrameProbe
	Decision belief.Decision
}

// Run 驱动时间线 → adapter→belief→probe→decide。返回逐帧结果 + 最终 marginal S。
func Run(timeline []adapter.FrameInput) ([]Frame, belief.Vector) {
	p := adapter.DefaultParams()
	nb := 0
	if len(timeline) > 0 {
		nb = len(timeline[0].Beds)
	}
	geom := adapter.BedGeoms(timeline[0])
	f := belief.NewFilter(belief.DefaultModel(), nb)
	js := f.Space()
	cp := belief.NewCoupling(geom)
	em := belief.NewEmission(geom)
	dec := belief.NewDecider()

	frames := make([]Frame, 0, len(timeline))
	for _, fi := range timeline {
		logPsi := cp.LogPsi(js, adapter.Gxy(fi, p))
		logPhi := em.LogPhi(js, adapter.BuildObservation(fi, p))
		f.Step(fi.NowMs, adapter.Online(fi), logPsi, logPhi)
		pF := js.PFallen(f.Alpha())
		lam := belief.ComputeLambda(js, logPsi, logPhi)
		d := dec.Step(fi.NowMs, pF, lam, adapter.BuildRiskContext(fi))
		frames = append(frames, Frame{
			Probe:    belief.Snapshot(js, f, cp, d, lam, fi.NowMs),
			Decision: d,
		})
	}
	return frames, js.MarginalS(f.Alpha())
}
