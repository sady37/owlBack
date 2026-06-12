// simulate-make —— lego 合成 fixture（window.json + meta.json），AI 灵活造 fall/ghost 测试场景。
// 与 export（PG 真实数据）+ replay（消费灌 live）共用同一 fixture 契约。**不连 DB/redis**。
//
// 场景 = lego 块（walk/lie/fall/jump 片段）拼装成 1Hz radar track 帧序列。加新场景 = 往 scenarios 注册一个函数。
//
// 用法:
//   go run ./simulate-make/ --list
//   go run ./simulate-make/ --scenario open-floor-fall --uid SIM00000CD2B \
//       --addr fd00:0:3:112:3:100:32a1:cd2b --out ../doc/cases/sim-fall --dur-s 20
//
// 输出 <out>/：window.json（合成 track 帧）+ meta.json（uid→addr→type）。
// 注：--addr 须是 live sensor 已绑定（有 DB layout）的真实设备 addr，否则 live 不认这条流。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"owl-common/observation"
)

// radarFrame 对齐生产 radar track data_value（firmware 上报，radar 系坐标 cm）。
type radarFrame struct {
	Pose            int `json:"pose"`
	AreaID          int `json:"area_id"`
	TrackID         int `json:"track_id"`
	PositionX       int `json:"position_x"`
	PositionY       int `json:"position_y"`
	PositionZ       int `json:"position_z"`
	RemainingTime   int `json:"remaining_time"`
	TrackConfidence int `json:"track_confidence"`
}

// fxRec 统一 fixture 记录（与 export / replay 同格式）。
type fxRec struct {
	Category  string        `json:"category"`
	DeviceUID string        `json:"device_uid"`
	Timestamp int64         `json:"timestamp"`
	DataValue []interface{} `json:"data_value"`
}

func trackRec(uid string, ts int64, fs ...radarFrame) fxRec {
	dv := make([]interface{}, len(fs))
	for i, f := range fs {
		dv[i] = f
	}
	return fxRec{Category: "track", DeviceUID: uid, Timestamp: ts, DataValue: dv}
}

// scenario 生成 1Hz radar track 帧序列。lego 块组合。
type scenario func(uid string, startMs int64, durSec int) []fxRec

// scenOpenFloorFall：走进(walk 0-4s, AreaID=4 开阔地)→ 倒地(fallen@5s)→ 持续躺地。验证 open-floor fall 检测。
func scenOpenFloorFall(uid string, startMs int64, durSec int) []fxRec {
	var out []fxRec
	const x0 = -150
	for i := 0; i < durSec; i++ {
		ts := startMs + int64(i)*1000
		f := radarFrame{TrackID: 0, AreaID: 4, PositionY: 100, TrackConfidence: 80}
		switch {
		case i < 5:
			f.Pose, f.PositionX, f.PositionZ = observation.PoseWalking, x0+i*25, 165
		case i == 5:
			f.Pose, f.PositionX, f.PositionZ = observation.PoseFallen, x0+100, 20
		default:
			f.Pose, f.PositionX, f.PositionZ = observation.PoseLying, x0+100, 15
		}
		out = append(out, trackRec(uid, ts, f))
	}
	return out
}

// scenGhostJump：track 每帧瞬移(>120cm/s 室内天花板)= shadow realness jump-ghost。验证 ghost_veto(reason=shadow_realness_jump)。
func scenGhostJump(uid string, startMs int64, durSec int) []fxRec {
	var out []fxRec
	for i := 0; i < durSec; i++ {
		ts := startMs + int64(i)*1000
		x := -200
		if i%2 == 1 {
			x = 200 // 每秒 400cm 瞬移
		}
		out = append(out, trackRec(uid, ts, radarFrame{
			Pose: observation.PoseStanding, AreaID: 4, TrackID: 0,
			PositionX: x, PositionY: 100, PositionZ: 165, TrackConfidence: 80,
		}))
	}
	return out
}

var scenarios = map[string]scenario{
	"open-floor-fall": scenOpenFloorFall,
	"ghost-jump":      scenGhostJump,
}

func main() {
	var (
		scen    = flag.String("scenario", "", "场景名（--list 看全部）")
		uid     = flag.String("uid", "", "合成 device_uid")
		addr    = flag.String("addr", "", "device_addr（须 live 已绑定，有 DB layout）")
		startMs = flag.Int64("start-ms", 0, "起始 ms（0=now）")
		durS    = flag.Int("dur-s", 20, "时长秒")
		out     = flag.String("out", "", "输出 case 目录")
		list    = flag.Bool("list", false, "列出可用场景")
	)
	flag.Parse()

	if *list {
		fmt.Println("可用场景:")
		for k := range scenarios {
			fmt.Println("  " + k)
		}
		return
	}
	gen, ok := scenarios[*scen]
	if !ok {
		fmt.Fprintf(os.Stderr, "未知场景 %q（--list 看可用）\n", *scen)
		os.Exit(1)
	}
	if *uid == "" || *addr == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "必填: --uid --addr --out")
		os.Exit(1)
	}
	start := *startMs
	if start == 0 {
		start = time.Now().UnixMilli()
	}
	recs := gen(*uid, start, *durS)

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	writeJSON(filepath.Join(*out, "window.json"), recs)
	writeJSON(filepath.Join(*out, "meta.json"), map[string]interface{}{
		"devices": []map[string]string{{"device_uid": *uid, "device_addr": *addr, "device_type": "Radar"}},
	})
	fmt.Printf("合成 %s: %d 帧 → %s（uid=%s addr=%s）\n回放: cd tools && go run ./replay/ --fixture %s --speed 30\n",
		*scen, len(recs), *out, *uid, *addr, *out)
}

func writeJSON(path string, v interface{}) {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
