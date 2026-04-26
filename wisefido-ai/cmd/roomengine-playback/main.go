// roomengine-playback：从 iot_timeseries 拉指定设备的近 N 小时 radar monitor 数据，
// 喂给 RoomEngine 流水线，生成单文件 HTML 滑动查看器。
//
// 用法：
//
//	go run wisefido-ai/cmd/roomengine-playback \
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
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"

	"wisefido-ai/internal/playback"
	"wisefido-ai/internal/roomengine"
)

func main() {
	var (
		layoutPath = flag.String("layout", "doc/layout-09E7-room101.json", "layout JSON 路径")
		roomID     = flag.String("room", "09E7-room101", "room id（仅作显示标识）")
		deviceUID  = flag.String("uid", "9D8A326309E7", "device_uid（雷达硬件 UID）")
		tenantID   = flag.String("tenant", "", "tenant_id（缺省自动查 devices 表）")
		hours      = flag.Int("hours", 48, "回放最近 N 小时")
		snapMin    = flag.Int("snap", 30, "快照间隔（模拟分钟）")
		outHTML    = flag.String("out", "doc/playback-09E7.html", "输出 HTML 路径")
		chunkHours = flag.Int("chunk", 6, "DB 分块查询大小（小时）")
		rowLimit   = flag.Int("row-limit", 30000, "单块最大行数")
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
	end := time.Now()
	start := end.Add(time.Duration(-*hours) * time.Hour)

	res, err := playback.Run(ctx, playback.Options{
		RoomID:    *roomID,
		Cfg:       cfg,
		DB:        db,
		DeviceUID: *deviceUID,
		TenantID:  *tenantID,
		Start:     start,
		End:       end,
		SnapMin:   *snapMin,
		ChunkHrs:  *chunkHours,
		RowLimit:  *rowLimit,
	})
	if err != nil {
		log.Fatalf("playback run: %v", err)
	}
	log.Printf("processed %d rows, %d valid frames, %d snapshots (grid %d×%d)",
		res.TotalRows, res.TotalFrames, len(res.Snapshots), res.GridW, res.GridH)
	log.Printf("silent-fall stats: pending_created=%d cancelled=%d reported=%d outstanding=%d",
		res.SilentFallPendingCreated, res.SilentFallPendingCancelled,
		res.SilentFallReported, res.SilentFallOutstanding)

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
