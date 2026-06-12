// roomengine-playback：从 iot_timeseries 拉指定设备的近 N 小时 radar monitor 数据，
// 喂给 RoomEngine 流水线，生成单文件 HTML 滑动查看器。
//
// 用法：
//
//	go run wisefido-sensor/cmd/roomengine-playback \
//	   --layout doc/layout-09E7-room101.json \
//	   --uid    9D8A326309E7 \
//	   --hours  48 \
//	   --snap   30 \
//	   --out    doc/playback-09E7.html
//
// 环境变量：DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME / DB_SSLMODE
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

	"wisefido-sensor/internal/playback"
	"wisefido-sensor/internal/roomengine"
)

func main() {
	var (
		layoutPath = flag.String("layout", "doc/layout-09E7-room101.json", "layout JSON 路径")
		roomID     = flag.String("room", "09E7-room101", "room id（仅作显示标识）")
		deviceUID  = flag.String("uid", "9D8A326309E7", "device_uid（雷达硬件 UID）")
		hours      = flag.Int("hours", 48, "回放最近 N 小时（被 --start-ms/--end-ms 覆盖）")
		startMs    = flag.Int64("start-ms", 0, "回放起始 epoch ms（与 --end-ms 一起使用）")
		endMs      = flag.Int64("end-ms", 0, "回放结束 epoch ms（与 --start-ms 一起使用）")
		snapMin    = flag.Int("snap", 30, "快照间隔（模拟分钟）")
		outHTML    = flag.String("out", "doc/playback-09E7.html", "输出 HTML 路径")
		chunkHours = flag.Int("chunk", 6, "DB 分块查询大小（小时）")
		rowLimit   = flag.Int("row-limit", 30000, "单块最大行数")
		feedback   = flag.Bool("feedback", true, "灌入历史 alarm_events 人类反馈到 grid（PR-9.1）")
		startISO   = flag.String("start", "", "回放起始（RFC3339 / 2006-01-02 15:04:05 local），与 --end 配合")
		endISO     = flag.String("end", "", "回放结束")
		cycles     = flag.Int("cycles", 1, "重放循环次数（>1 时同段数据连跑 N 遍，simT 单调递增）")
	)
	flag.Parse()

	ctx := context.Background()

	// 1. 读 layout
	raw, err := os.ReadFile(*layoutPath)
	if err != nil {
		log.Fatalf("read layout: %v", err)
	}
	cfg, err := roomengine.ParseLayoutConfig(*roomID, raw)
	if err != nil {
		log.Fatalf("parse layout: %v", err)
	}
	rawW, rawH := cfg.RoomW, cfg.RoomH
	roomengine.ApplyOptimizedExtent(&cfg)
	log.Printf("layout loaded: room %s  raw=%d×%d cm → grid=%d×%d cm  origin=(%d,%d)  radar=(%d,%d,%d) install=%d",
		*roomID, rawW, rawH, cfg.RoomW, cfg.RoomH, cfg.OriginX, cfg.OriginY,
		cfg.Radar.Center.X, cfg.Radar.Center.Y, cfg.Radar.Center.Z, cfg.Radar.InstallModel)

	// 2. 连 DB
	db, err := playback.OpenDB()
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// 3. 跑回放
	var start, end time.Time
	parseTime := func(s string) (time.Time, error) {
		layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04"}
		for _, ly := range layouts {
			if t, err := time.ParseInLocation(ly, s, time.Local); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse %q", s)
	}
	if *startISO != "" && *endISO != "" {
		var err error
		if start, err = parseTime(*startISO); err != nil {
			log.Fatalf("parse start: %v", err)
		}
		if end, err = parseTime(*endISO); err != nil {
			log.Fatalf("parse end: %v", err)
		}
		log.Printf("playback window (--start/--end): %s ~ %s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	} else if *startMs > 0 && *endMs > 0 {
		start = time.UnixMilli(*startMs)
		end = time.UnixMilli(*endMs)
		log.Printf("playback window (explicit): %s ~ %s", start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339))
	} else {
		end = time.Now()
		start = end.Add(time.Duration(-*hours) * time.Hour)
		log.Printf("playback window (last %dh): %s ~ %s", *hours, start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339))
	}
	if *cycles > 1 {
		log.Printf("playback cycles: %d (simT monotonically increasing across cycles)", *cycles)
	}

	res, err := playback.Run(ctx, playback.Options{
		RoomID:                   *roomID,
		Cfg:                      cfg,
		DB:                       db,
		DeviceUID:                *deviceUID,
		Start:                    start,
		End:                      end,
		SnapMin:                  *snapMin,
		ChunkHrs:                 *chunkHours,
		RowLimit:                 *rowLimit,
		IngestHistoricalFeedback: *feedback,
		Cycles:                   *cycles,
	})
	if err != nil {
		log.Fatalf("playback run: %v", err)
	}
	log.Printf("processed %d rows, %d valid frames, %d snapshots (grid %d×%d)",
		res.TotalRows, res.TotalFrames, len(res.Snapshots), res.GridW, res.GridH)

	// 4. 写 HTML
	if err := os.MkdirAll(filepath.Dir(*outHTML), 0755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}
	f, err := os.Create(*outHTML)
	if err != nil {
		log.Fatalf("create %s: %v", *outHTML, err)
	}
	defer f.Close()
	if err := playback.WriteHTML(f, *roomID, res.Snapshots); err != nil {
		log.Fatalf("write html: %v", err)
	}
	abs, _ := filepath.Abs(*outHTML)
	log.Printf("HTML viewer: %s", abs)
}
