// fall-score-replay：把指定窗口的 firmware Fall 报警按时间注入 engine，
// 同时回放 monitor + event 流，跑 PR-5 verifier，输出每条 alarm 的评分表。
//
// 用法：
//
//	go run ./cmd/fall-score-replay \
//	   --layout doc/cases/case_lostfall_cd2b_11351148/room_layout.json \
//	   --uid 9D8A32A1CD2B \
//	   --start-ms 1777401600000 \
//	   --end-ms 1777417200000
//
// 输出：
//   - 实时 ai.log（structured zap dev format）
//   - 终止表：每条 firmware Fall alarm + verifier 评分 + 主导信号
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"owl-common/radarutils"
	"wisefido-ai/internal/playback"
	"wisefido-ai/internal/roomengine"
)

type fallEvent struct {
	EventID      string
	DeviceID     string
	DeviceUID    string
	TriggeredMs  int64
	TrackID      int
	Operation    string
	HandledLabel string
}

type verifyRecord struct {
	TMs            int64
	TrackID        int
	Score          int
	Verdict        string
	DominantSignal string
	Breakdown      map[string]int
}

func main() {
	layoutPath := flag.String("layout", "", "layout JSON 路径（可选；缺省时不构建 grid）")
	roomID := flag.String("room", "Bedroom", "room id")
	deviceUID := flag.String("uid", "", "device_uid 必填")
	startMs := flag.Int64("start-ms", 0, "回放窗起始 epoch ms")
	endMs := flag.Int64("end-ms", 0, "回放窗结束 epoch ms")
	flag.Parse()

	if *deviceUID == "" || *startMs == 0 || *endMs == 0 || *layoutPath == "" {
		log.Fatalf("required: --layout --uid --start-ms --end-ms")
	}

	ctx := context.Background()

	// 1. layout
	rawLayout, err := os.ReadFile(*layoutPath)
	if err != nil {
		log.Fatalf("read layout: %v", err)
	}
	cfg, err := roomengine.ParseLayoutConfig(*roomID, rawLayout)
	if err != nil {
		log.Fatalf("parse layout: %v", err)
	}
	roomengine.ApplyOptimizedExtent(&cfg)

	// 2. DB
	db, err := playback.OpenDB()
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// 3. 抓 firmware Fall 报警（含 verified / false_alarm 全部 operation）
	falls, err := loadFirmwareFalls(ctx, db, *deviceUID, *startMs, *endMs)
	if err != nil {
		log.Fatalf("load falls: %v", err)
	}
	log.Printf("firmware Fall events in window: %d", len(falls))

	// 4. captures verifier output via custom zap core
	captured := make(map[int64]*verifyRecord) // key = ts_ms
	captureCore := newVerifyCaptureCore(captured)

	devLogger, _ := zap.NewDevelopment()
	logger := zap.New(zapcore.NewTee(devLogger.Core(), captureCore))

	// 5. 跑 playback Run（喂 monitor + event 流给 engine）
	res, err := playback.Run(ctx, playback.Options{
		RoomID:    *roomID,
		Cfg:       cfg,
		DB:        db,
		DeviceUID: *deviceUID,
		Start:     time.UnixMilli(*startMs),
		End:       time.UnixMilli(*endMs),
		SnapMin:   60,
		ChunkHrs:  6,
		RowLimit:  30000,
		// 在 ProcessFrame 之前每帧也喂 alarm（按时间戳）
		AlarmInjector: makeAlarmInjector(falls, cfg.Radar),
		Logger:        logger,
	})
	if err != nil {
		log.Fatalf("playback: %v", err)
	}

	// 6. 输出表
	fmt.Println("\n=================== Verify Result Table ===================")
	fmt.Printf("%-20s | %-3s | %-5s | %-7s | %s\n",
		"firmware_alarm_at(MDT)", "tid", "score", "verdict", "dominant_signal")
	fmt.Println("------------------------------------------------------------")
	loc, _ := time.LoadLocation("America/Denver")

	// 按时间排序
	sort.Slice(falls, func(i, j int) bool { return falls[i].TriggeredMs < falls[j].TriggeredMs })

	stat := struct{ ghost, suspect, real, missed int }{}
	for _, f := range falls {
		rec, ok := captured[f.TriggeredMs]
		if !ok {
			// 备选：±2s 内匹配（event_since vs alarm 落账时刻可能差几 ms）
			for k, r := range captured {
				if abs64(k-f.TriggeredMs) <= 2000 && r.TrackID == f.TrackID {
					rec = r
					ok = true
					break
				}
			}
		}
		t := time.UnixMilli(f.TriggeredMs).In(loc).Format("2006-01-02 15:04:05")
		if !ok {
			fmt.Printf("%-20s | %-3d | %-5s | %-7s | %s (firmware %s)\n",
				t, f.TrackID, "?", "?", "(no engine record)", f.Operation)
			stat.missed++
			continue
		}
		fmt.Printf("%-20s | %-3d | %-5d | %-7s | %s (firmware %s)\n",
			t, f.TrackID, rec.Score, rec.Verdict, rec.DominantSignal, f.Operation)
		switch rec.Verdict {
		case "ghost":
			stat.ghost++
		case "suspect":
			stat.suspect++
		case "real":
			stat.real++
		}
	}
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("Verifier verdict counts: ghost=%d suspect=%d real=%d (missed=%d)\n",
		stat.ghost, stat.suspect, stat.real, stat.missed)

	// 7. 也输出 engine 自己 silent/lost/still 报警，作为 "潜在漏报" 的反向证据
	fmt.Println("\n----- Engine independent fall events in same window -----")
	fmt.Printf("silent_fall (old 60s):   pending=%d cancelled=%d reported=%d\n",
		res.SilentFallPendingCreated, res.SilentFallPendingCancelled, res.SilentFallReported)
	fmt.Printf("lost_fall:                pending=%d cancelled=%d reported=%d\n",
		res.LostFallPendingCreated, res.LostFallPendingCancelled, res.LostFallReported)
	fmt.Printf("silent_fall (leftbed v2): reported=%d cancelled=%d\n",
		res.SilentFallLeftBedReported, res.SilentFallLeftBedCancelled)
	fmt.Printf("still_fall:               reported=%d\n", res.StillFallReported)
	fmt.Printf("(每条 engine 独立报警 = 与 firmware Fall 不同源 → 候选「漏报」)\n")
}

// loadFirmwareFalls 拉指定窗口内 firmware Fall 报警（不论 operation）。
func loadFirmwareFalls(ctx context.Context, db *sql.DB, deviceUID string,
	startMs, endMs int64) ([]fallEvent, error) {

	q := `
		SELECT
			ae.event_id::text,
			ae.device_id::text,
			d.device_uid,
			extract(epoch from ae.triggered_at)*1000 AS trigger_ms,
			ae.operation,
			ae.alarm_status,
			ae.trigger_data->>'event_payload' AS payload
		FROM alarm_events ae
		JOIN devices d ON d.device_id = ae.device_id
		WHERE d.device_uid = $1
		  AND ae.event_type = 'Fall'
		  AND ae.triggered_at >= to_timestamp($2 / 1000.0)
		  AND ae.triggered_at <= to_timestamp($3 / 1000.0)
		ORDER BY ae.triggered_at
	`
	rows, err := db.QueryContext(ctx, q, deviceUID, startMs, endMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []fallEvent
	for rows.Next() {
		var e fallEvent
		var ms float64
		var payload string
		if err := rows.Scan(&e.EventID, &e.DeviceID, &e.DeviceUID, &ms,
			&e.Operation, &e.HandledLabel, &payload); err != nil {
			return nil, err
		}
		e.TriggeredMs = int64(ms)
		// payload: {"event_name":"Fall","event_type":2,"pose":2,"track_id":0}
		if payload != "" {
			var p map[string]interface{}
			if err := json.Unmarshal([]byte(payload), &p); err == nil {
				if v, ok := p["track_id"].(float64); ok {
					e.TrackID = int(v)
				}
			}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// makeAlarmInjector 返回 playback 每个 monitor frame 处理后调用的注入函数。
// 每次 nowMs 推进时，把 [prev, nowMs] 之间所有 firmware Fall 注入到 engine。
func makeAlarmInjector(falls []fallEvent, mount radarutils.RadarMount) playback.AlarmInjector {
	sort.Slice(falls, func(i, j int) bool { return falls[i].TriggeredMs < falls[j].TriggeredMs })
	idx := 0
	return func(tm *roomengine.TrackManager, nowMs int64) {
		for idx < len(falls) && falls[idx].TriggeredMs <= nowMs {
			f := falls[idx]
			tm.RecordRadarAlarm(roomengine.RadarFallAlarm{
				DeviceUID: f.DeviceUID,
				TMs:       f.TriggeredMs,
				TrackID:   f.TrackID,
				Pose:      5,
				EventType: 2,
				Status:    "start",
			})
			idx++
		}
	}
}

// === capture core ===

type verifyCaptureCore struct {
	zapcore.Core
	captured map[int64]*verifyRecord
}

func newVerifyCaptureCore(captured map[int64]*verifyRecord) zapcore.Core {
	return &verifyCaptureCore{
		Core:     zapcore.NewNopCore(),
		captured: captured,
	}
}

func (c *verifyCaptureCore) Enabled(_ zapcore.Level) bool { return true }
func (c *verifyCaptureCore) With(_ []zapcore.Field) zapcore.Core {
	return c
}
func (c *verifyCaptureCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if e.Message == "radar_fall_verify" {
		return ce.AddCore(e, c)
	}
	return ce
}
func (c *verifyCaptureCore) Write(_ zapcore.Entry, fields []zapcore.Field) error {
	rec := &verifyRecord{Breakdown: map[string]int{}}
	for _, f := range fields {
		switch f.Key {
		case "ts_ms":
			rec.TMs = f.Integer
		case "track_id":
			rec.TrackID = int(f.Integer)
		case "score":
			rec.Score = int(f.Integer)
		case "verdict":
			rec.Verdict = f.String
		case "dominant_signal":
			rec.DominantSignal = f.String
		}
		if len(f.Key) > 3 && f.Key[:3] == "bd_" {
			rec.Breakdown[f.Key[3:]] = int(f.Integer)
		}
	}
	if rec.TMs > 0 {
		c.captured[rec.TMs] = rec
	}
	return nil
}
func (c *verifyCaptureCore) Sync() error { return nil }

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
