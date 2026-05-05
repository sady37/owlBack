// feedback-once: 一次性跑 alarm_events false_alarm 反馈同步（PR-4 实测用）。
//
// 启动时：
//  1. 加载所有有 layout_config 的房间到 engine
//  2. 建 device → room 路由
//  3. 调用 ingester.IngestOnce
//  4. 打印结果（处理条数 + 当前 cursor）
//
// 不启动 Redis 消费循环；不长驻。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"wisefido-sensor/internal/playback"
	"wisefido-sensor/internal/roomengine"
)

func main() {
	resetCursor := flag.Bool("reset-cursor", false, "把 last_hand_time 设回 epoch 后再跑（全量回放历史）")
	flag.Parse()

	ctx := context.Background()

	db, err := playback.OpenDB()
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if *resetCursor {
		if _, err := db.ExecContext(ctx,
			`UPDATE engine_alarm_feedback_cursor SET last_hand_time='epoch', last_event_id=NULL, processed_total=0 WHERE id=1`,
		); err != nil {
			log.Fatalf("reset cursor: %v", err)
		}
		log.Printf("cursor reset to epoch")
	}

	logger, _ := zap.NewDevelopment()

	engine := roomengine.NewEngine(nil, logger)

	// 加载所有 room（同 wisefido-sensor bootstrap 简化版）
	rows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text, r.room_name, r.layout_config::text, COALESCE(u.timezone, '')
		FROM rooms r
		LEFT JOIN units u ON u.unit_id = r.unit_id
		WHERE r.layout_config IS NOT NULL
	`)
	if err != nil {
		log.Fatalf("query rooms: %v", err)
	}
	roomCount := 0
	for rows.Next() {
		var roomID, roomName, layout, tz string
		if err := rows.Scan(&roomID, &roomName, &layout, &tz); err != nil {
			continue
		}
		cfg, err := roomengine.ParseLayoutConfig(roomID, []byte(layout))
		if err != nil {
			continue
		}
		cfg.RoomName = roomName
		cfg.Timezone = tz
		roomengine.ApplyOptimizedExtent(&cfg)
		engine.RegisterRoom(cfg)
		roomCount++
	}
	rows.Close()
	log.Printf("rooms registered: %d", roomCount)

	// 建 device → room 路由
	rows, err = db.QueryContext(ctx, `
		SELECT d.device_id::text, d.device_uid,
		       COALESCE(d.bound_room_id, b.room_id)::text AS room_id
		FROM devices d
		LEFT JOIN beds b ON b.bed_id = d.bound_bed_id
		WHERE d.bound_room_id IS NOT NULL OR (d.bound_bed_id IS NOT NULL AND b.room_id IS NOT NULL)
	`)
	if err != nil {
		log.Fatalf("query devices: %v", err)
	}
	devCount := 0
	for rows.Next() {
		var devID, devUID, roomID string
		if err := rows.Scan(&devID, &devUID, &roomID); err != nil {
			continue
		}
		if roomID == "" {
			continue
		}
		if devID != "" {
			engine.MapDeviceToRoom(devID, roomID)
		}
		if devUID != "" {
			engine.MapDeviceToRoom(devUID, roomID)
		}
		devCount++
	}
	rows.Close()
	log.Printf("devices mapped: %d", devCount)

	// 跑反馈
	ingester := roomengine.NewAlarmFeedbackIngester(db, engine, logger)
	start := time.Now()
	processed, err := ingester.IngestOnce(ctx)
	dur := time.Since(start)
	if err != nil {
		log.Fatalf("ingest: %v", err)
	}
	log.Printf("ingest done: processed=%d duration=%v", processed, dur)

	// 当前 cursor 状态
	var lastHandTime time.Time
	var processedTotal int64
	_ = db.QueryRowContext(ctx,
		`SELECT last_hand_time, processed_total FROM engine_alarm_feedback_cursor WHERE id=1`,
	).Scan(&lastHandTime, &processedTotal)
	log.Printf("cursor: last_hand_time=%s processed_total=%d", lastHandTime.UTC().Format(time.RFC3339), processedTotal)

	_ = filepath.Join // silence unused
	_ = fmt.Println
	_ = os.Args
}
