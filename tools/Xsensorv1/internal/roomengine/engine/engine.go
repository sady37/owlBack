package engine

import (
	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
)

// engine.go — 单房 roomengine 主循环（build order ③ W3.2，DBN-wire-roadmap §6）。
// 四隐轴 S/B/realness/neighbor 全经 belief.Filter.Step 融合（避免 gate）：
//   S/B  ← adapter 译 FrameInput（Observation/Gxy/Online）。
//   realness ← pFallReal（W3.3 由上游出生档案/几何译入；单房中性=1）。
//   neighbor ← rhoXroom（W3.4 由跨房 census 读出；单房=0）。
// 多房编排（W3.4a）在此之上持 N 个 Room + 跨房 census 产 rhoXroom。

// Frame 单帧引擎输出（probe + 裁决）。
type Frame struct {
	Probe    belief.FrameProbe
	Decision belief.Decision
}

// Room 单房四轴引擎：belief 滤波器 + 派生层（coupling/emission）+ 裁决器。
type Room struct {
	f   *belief.Filter
	js  *belief.JointSpace
	cp  *belief.Coupling
	em  *belief.Emission
	dec *belief.Decider
	p   adapter.Params
}

// NewRoom 建单房引擎。geom = 床几何（adapter.BedGeoms 从 layout 派生）；nb = 床数。
func NewRoom(geom []belief.BedGeom, nb int) *Room {
	f := belief.NewFilter(belief.DefaultModel(), nb)
	return &Room{
		f:   f,
		js:  f.Space(),
		cp:  belief.NewCoupling(geom),
		em:  belief.NewEmission(geom),
		dec: belief.NewDecider(),
		p:   adapter.DefaultParams(),
	}
}

// Tick 一帧推进。rhoXroom（neighbor，单房=0）、pFallReal（realness，中性=1）。
func (r *Room) Tick(fi adapter.FrameInput, rhoXroom, pFallReal float64) Frame {
	logPsi := r.cp.LogPsi(r.js, adapter.Gxy(fi, r.p))
	logPhi := r.em.LogPhi(r.js, adapter.BuildObservation(fi, r.p))
	r.f.Step(fi.NowMs, adapter.Online(fi), logPsi, logPhi, rhoXroom, pFallReal)
	pF := r.js.PFallen(r.f.Alpha())
	lam := belief.ComputeLambda(r.js, logPsi, logPhi)
	d := r.dec.Step(fi.NowMs, pF, lam, adapter.BuildRiskContext(fi))
	return Frame{
		Probe:    belief.Snapshot(r.js, r.f, r.cp, d, lam, fi.NowMs),
		Decision: d,
	}
}

// MarginalS 当前 S 轴边缘后验。
func (r *Room) MarginalS() belief.Vector { return r.js.MarginalS(r.f.Alpha()) }
