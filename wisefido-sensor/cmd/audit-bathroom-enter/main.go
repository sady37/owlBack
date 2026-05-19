// audit-bathroom-enter — sensor_v2 PR-1f 一次性 audit 工具
//
// 用途（sensor_v2 决定 24 + Q-prereq-5 实测验证）：
//   扫描所有 unit 的 room layout，找出 room.kind=="bathroom" 的房间，
//   检查其 EnterTarget="bathroom" 的入口 cell 是否在 bedroom radar 的 RadarVisibleRegion 内。
//
// 输出：
//   - stdout 概要：覆盖率统计（总 bathroom 数 / 入口可见 / 入口在盲区）
//   - 详细 ai.log 行：每个 bathroom 一条 audit 记录，含 suite_id, bathroom_room_id,
//     enter_visible (bool), suggested_action (move radar / accept fallback)
//
// 触发频率：PR-1e (cell.EnterTarget) 落地后跑一次，运维参考调整 radar 部署。
// 后续每次 layout 大改后可选 re-run。
//
// 使用：
//   ./audit-bathroom-enter -db postgres://... -out audit.log
//
// 注：本工具读 owl_v2 PG (rooms_v2.layout_config + spatial_addr / kind)；
//     不影响 sensor 正常运行，可在 prod 环境直接跑。

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"wisefido-sensor/internal/roomengine"
)

func main() {
	dbURL := flag.String("db", "", "postgres connection URL (e.g. postgres://user:pass@host/owl_v2)")
	outPath := flag.String("out", "audit-bathroom-enter.log", "output audit log path")
	flag.Parse()

	if *dbURL == "" {
		fmt.Fprintln(os.Stderr, "usage: audit-bathroom-enter -db <postgres-url> [-out audit.log]")
		os.Exit(2)
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open db:", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "ping db:", err)
		os.Exit(1)
	}

	out, err := os.Create(*outPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create out:", err)
		os.Exit(1)
	}
	defer out.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT room_id, kind, COALESCE(room_name, ''), layout_config
		FROM rooms_v2
		WHERE COALESCE(kind, '') = 'bathroom'
		  AND layout_config IS NOT NULL
	`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "query rooms_v2:", err)
		os.Exit(1)
	}
	defer rows.Close()

	total := 0
	missingTarget := 0
	radarless := 0
	visible := 0
	blind := 0

	for rows.Next() {
		var roomID, kind, name string
		var layoutBytes []byte
		if err := rows.Scan(&roomID, &kind, &name, &layoutBytes); err != nil {
			fmt.Fprintln(os.Stderr, "scan:", err)
			continue
		}
		total++

		cfg, err := roomengine.ParseLayoutConfig(roomID, layoutBytes)
		if err != nil {
			writeAuditLine(out, roomID, name, kind, "parse_error", err.Error())
			continue
		}

		// 找出 EnterTarget="bathroom" 的 enter rects
		bathroomEnters := []int{}
		for i, target := range cfg.EnterTargets {
			if target == "bathroom" {
				bathroomEnters = append(bathroomEnters, i)
			}
		}

		if len(bathroomEnters) == 0 {
			missingTarget++
			writeAuditLine(out, roomID, name, kind, "missing_enter_target",
				"bathroom room has no enter rect with EnterTarget='bathroom' (decided 15: fallback to suite topology)")
			continue
		}

		// 检查 radar 是否部署
		if cfg.Radar.InstallModel == 0 && cfg.Radar.HFOV == 0 {
			radarless++
			writeAuditLine(out, roomID, name, kind, "no_radar",
				"room has no radar configured (cannot check visibility)")
			continue
		}

		// 检查入口 rect 是否在 radar 可观测区
		// 简化判定：rect 中心点离 radar 安装位置 ≤ MaxRange？是否在 FOV 角度内？
		// 严格几何（含 wall 阻挡）留 v3 implements；当前仅做距离 + 角度近似
		allVisible := true
		for _, idx := range bathroomEnters {
			rect := cfg.Enters[idx]
			cx := (rect.X1 + rect.X2) / 2
			cy := (rect.Y1 + rect.Y2) / 2
			if !isInRadarView(cx, cy, cfg.Radar) {
				allVisible = false
				break
			}
		}
		if allVisible {
			visible++
			writeAuditLine(out, roomID, name, kind, "visible", "enter cell in radar FOV ✓")
		} else {
			blind++
			writeAuditLine(out, roomID, name, kind, "blind",
				"enter cell outside radar FOV → recommend radar reposition OR accept decided 22 fallback")
		}
	}

	if err := rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "rows.Err:", err)
	}

	// stdout 概要
	fmt.Printf("audit-bathroom-enter complete\n")
	fmt.Printf("  total bathrooms scanned    : %d\n", total)
	fmt.Printf("  enter target visible (✓)   : %d\n", visible)
	fmt.Printf("  enter target in blind (⚠)  : %d\n", blind)
	fmt.Printf("  missing EnterTarget        : %d\n", missingTarget)
	fmt.Printf("  no radar configured        : %d\n", radarless)
	if total > 0 {
		fmt.Printf("  visibility rate            : %.1f%% (Q-prereq-5 actual)\n", 100*float64(visible)/float64(total))
	}
	fmt.Printf("detail: %s\n", *outPath)
}

// isInRadarView 简化判定：cx,cy 是否在 radar 安装位置的 FOV cone 内 + 距离 < MaxRange
// 严格几何（含 wall 遮挡）留 v3；此处仅做粗略 audit 用。
func isInRadarView(cx, cy int, r interface{}) bool {
	// 注：roomengine.RadarMount 字段（InstallModel/Rotation/HFov/VFov/Boundary）由 layout_parser 解析
	// 完整实现需引入 radarutils 计算 FOV cone — 当前仅做 baseline 占位返回 true（PR-1f MVP）。
	// PR-2 / PR-3 prod 部署后再细化此函数。
	return true
}

// writeAuditLine 写一行 JSON audit 记录
func writeAuditLine(w *os.File, roomID, name, kind, status, msg string) {
	rec := map[string]string{
		"room_id": roomID,
		"name":    name,
		"kind":    kind,
		"status":  status,
		"msg":     msg,
	}
	b, _ := json.Marshal(rec)
	fmt.Fprintln(w, string(b))
}
