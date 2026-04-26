// roomengine-api：开发期 HTTP 工具服务，封装 internal/playback 提供 web 接口。
//
// 用法：
//
//	go run wisefido-ai/cmd/roomengine-api --listen :7788
//
// Endpoint:
//
//	GET /api/playback?uid=<device_uid>&start=<RFC3339>&end=<RFC3339>&snap_min=10[&format=html|json][&layout=<path>]
//	  - uid       必填，雷达 device_uid
//	  - start     必填，开始时间（RFC3339；带时区，例 2026-04-25T13:00:00-07:00）
//	  - end       必填，结束时间
//	  - snap_min  可选，快照间隔分钟，默认 10
//	  - format    可选，html (默认) / json
//	  - layout    可选，layout JSON 文件路径；缺省时从 rooms.layout_config 查
//
//	GET /api/health   返回 "ok"
//
// 限制：
//   - 单请求 hard timeout 60s
//   - 无 auth、无速率限制（开发用工具）
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"

	"wisefido-ai/internal/playback"
	"wisefido-ai/internal/roomengine"
)

const requestTimeout = 60 * time.Second

func main() {
	listen := flag.String("listen", ":7788", "HTTP listen addr")
	flag.Parse()

	db, err := playback.OpenDB()
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/playback", func(w http.ResponseWriter, r *http.Request) {
		handlePlayback(w, r, db)
	})

	log.Printf("roomengine-api listening on %s (endpoints: /api/health, /api/playback)", *listen)
	srv := &http.Server{
		Addr:              *listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// handlePlayback parses query params, runs replay, returns HTML or JSON
func handlePlayback(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	q := r.URL.Query()
	uid := q.Get("uid")
	startStr := q.Get("start")
	endStr := q.Get("end")
	if uid == "" || startStr == "" || endStr == "" {
		http.Error(w, "uid, start, end are required (start/end RFC3339)", http.StatusBadRequest)
		return
	}
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, "bad start (RFC3339): "+err.Error(), http.StatusBadRequest)
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		http.Error(w, "bad end (RFC3339): "+err.Error(), http.StatusBadRequest)
		return
	}
	if !end.After(start) {
		http.Error(w, "end must be after start", http.StatusBadRequest)
		return
	}
	snapMin := 10
	if s := q.Get("snap_min"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 {
			http.Error(w, "snap_min must be positive int", http.StatusBadRequest)
			return
		}
		snapMin = v
	}
	format := q.Get("format")
	if format == "" {
		format = "html"
	}
	if format != "html" && format != "json" {
		http.Error(w, "format must be html or json", http.StatusBadRequest)
		return
	}
	layoutPath := q.Get("layout")

	// 解析 layout：要么从文件读，要么从 DB 查
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	roomID, cfg, err := loadLayout(ctx, db, uid, layoutPath)
	if err != nil {
		http.Error(w, "load layout: "+err.Error(), http.StatusBadRequest)
		return
	}
	roomengine.ApplyOptimizedExtent(&cfg)

	// 跑回放
	res, err := playback.Run(ctx, playback.Options{
		RoomID:    roomID,
		Cfg:       cfg,
		DB:        db,
		DeviceUID: uid,
		Start:     start,
		End:       end,
		SnapMin:   snapMin,
	})
	if err != nil {
		// ctx timeout/cancel 走 504；其他走 500
		if ctx.Err() != nil {
			http.Error(w, "playback timeout (60s) — narrow time range", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "playback: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("playback ok uid=%s span=%s~%s rows=%d frames=%d snaps=%d format=%s",
		uid, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"),
		res.TotalRows, res.TotalFrames, len(res.Snapshots), format)

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"room_id":      roomID,
			"total_rows":   res.TotalRows,
			"total_frames": res.TotalFrames,
			"grid_w":       res.GridW,
			"grid_h":       res.GridH,
			"snapshots":    res.Snapshots,
		})
		return
	}
	// HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := playback.WriteHTML(w, roomID, res.Snapshots); err != nil {
		log.Printf("write html: %v", err)
	}
}

// loadLayout 优先用 layoutPath 指定的文件；否则从 rooms.layout_config 查（用 device_uid 找绑定房间）。
// 返回 roomID（显示用）+ 已 ParseLayoutConfig 的 cfg。
func loadLayout(ctx context.Context, db *sql.DB, deviceUID, layoutPath string) (string, roomengine.RoomConfig, error) {
	if layoutPath != "" {
		raw, err := os.ReadFile(layoutPath)
		if err != nil {
			return "", roomengine.RoomConfig{}, fmt.Errorf("read %s: %w", layoutPath, err)
		}
		// roomID 用文件名（去 .json 后缀）
		roomID := layoutPath
		if i := lastSep(roomID); i >= 0 {
			roomID = roomID[i+1:]
		}
		if i := lastDot(roomID); i > 0 {
			roomID = roomID[:i]
		}
		cfg, err := roomengine.ParseLayoutConfig(roomID, raw)
		if err != nil {
			return "", roomengine.RoomConfig{}, err
		}
		return roomID, cfg, nil
	}

	roomID, raw, err := playback.LookupRoomLayout(ctx, db, deviceUID)
	if err != nil {
		return "", roomengine.RoomConfig{}, err
	}
	cfg, err := roomengine.ParseLayoutConfig(roomID, raw)
	if err != nil {
		return "", roomengine.RoomConfig{}, err
	}
	return roomID, cfg, nil
}

func lastSep(s string) int {
	i := -1
	for j := 0; j < len(s); j++ {
		if s[j] == '/' || s[j] == '\\' {
			i = j
		}
	}
	return i
}

func lastDot(s string) int {
	i := -1
	for j := 0; j < len(s); j++ {
		if s[j] == '.' {
			i = j
		}
	}
	return i
}
