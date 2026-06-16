// replay-d5f7 — 把 case-d5f7 真实 firmware 数据按雷达周期(同 ts 分组多 track)喂进**真实** Xsensorv1
// 引擎 engine.Room（census 多 track + 桶二 walls → 每 track 一份 belief DBN → presentCount 共存门控 → OR 聚合），
// 打印逐周期真实 DBN 计算（N_r / P^F / 顶态 S / fire / Λ）。非合成 fixture、非单层 census——这是 Xsensor
// 替换 Tsensor 后生产路径的真实计算。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
	"owlBack/tools/Xsensorv1/internal/roomengine/engine"
)

type rec struct {
	Timestamp int64  `json:"timestamp"`
	Category  string `json:"category"`
	DataValue []struct {
		Pose    int `json:"pose"`
		TrackID int `json:"track_id"`
		X       int `json:"position_x"`
		Y       int `json:"position_y"`
		Z       int `json:"position_z"`
	} `json:"data_value"`
}

func sName(i int) string {
	n := []string{"Empty", "Bed", "Sit", "OpenFloor", "Bath", "Fallen", "BlindRest", "BlindOpen", "Left"}
	if i >= 0 && i < len(n) {
		return n[i]
	}
	return fmt.Sprintf("S%d", i)
}

func main() {
	path := "doc/cases/case-d5f7-0425-172017350-ghost/2026-04-25_17-20_to_17-35_MDT.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	var raw []rec
	if err := json.Unmarshal(b, &raw); err != nil {
		fmt.Fprintln(os.Stderr, "unmarshal:", err)
		os.Exit(1)
	}

	// 0425 浴室墙（areaType==-1 四段，雷达 h,v 系，雷达=原点）。房 box h∈[-120,180] v∈[-80,70]。
	walls := []adapter.Rect{
		{X1: -120, Y1: -80, X2: 180, Y2: -80},
		{X1: 180, Y1: -80, X2: 180, Y2: 70},
		{X1: -40, Y1: 70, X2: 180, Y2: 70},
		{X1: -120, Y1: -80, X2: -120, Y2: 70},
	}
	radar := adapter.Point{X: 0, Y: 0}

	// 按 ts 分组 → 还原雷达每周期 payload（排 track88 空占位）。
	type tk struct {
		pose, x, y, z, fwid int
	}
	byts := map[int64][]tk{}
	var order []int64
	for _, r := range raw {
		if r.Category != "track" {
			continue
		}
		for _, d := range r.DataValue {
			if d.TrackID == 88 {
				continue
			}
			if _, ok := byts[r.Timestamp]; !ok {
				order = append(order, r.Timestamp)
			}
			byts[r.Timestamp] = append(byts[r.Timestamp], tk{d.Pose, d.X, d.Y, d.Z, d.TrackID})
		}
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })

	room := engine.NewRoom(nil, 0) // 浴室无床 nb=0
	base := order[0]
	fmt.Printf("=== 真实 DBN replay：case-d5f7-0425（%d 周期，雷达 h,v 系）===\n", len(order))
	fmt.Printf("%-8s %-14s %-4s %-6s %-16s %-6s %-6s\n", "t(+s)", "fw-tracks", "N_r", "P^F", "顶态S(P)", "fire", "Λ")
	for _, ts := range order {
		cyc := byts[ts]
		obs := make([]adapter.TrackObs, 0, len(cyc))
		fwids := make([]int, 0, len(cyc))
		for _, t := range cyc {
			obs = append(obs, adapter.TrackObs{RadarTrack: adapter.RadarTrack{
				Online: true, Pose: t.pose, X: t.x, Y: t.y, Z: t.z,
			}})
			fwids = append(fwids, t.fwid)
		}
		fi := adapter.FrameInput{
			NowMs:    ts,
			Tracks:   obs,
			Walls:    walls,
			RadarPos: radar,
		}
		fr := room.Tick(fi, 0)
		top, tp := fr.Probe.MarginalS.Max()
		fmt.Printf("+%-7.1f %-14v %-4d %-6.3f %-16s %-6v %-6.2f\n",
			float64(ts-base)/1000, fwids, fr.Decision.PeopleCount, fr.Probe.PFallen,
			fmt.Sprintf("%s(%.2f)", sName(int(top)), tp), fr.Decision.Fire, fr.Probe.Lambda)
	}
	_ = belief.SFallen
}
