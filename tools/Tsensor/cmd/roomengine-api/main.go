// roomengine-api：开发期 HTTP 工具服务，封装 internal/playback 提供 web 接口。
//
// 用法：
//
//	go run wisefido-sensor/cmd/roomengine-api --listen :7788 --out-dir ../out
//
// 路径前缀刻意避开 /api/*（owlBack 的 auth middleware 拦截一切 /api/*），
// 用 /roomengine/* 让 nginx/caddy 默认转给 vite，vite 再 proxy 到 :7788。
//
// Endpoint:
//
//	GET /roomengine/playback?uid=<device_uid>&start=<RFC3339>&end=<RFC3339>&snap_min=10[&format=html|json][&layout=<path>]
//	  - uid       必填，雷达 device_uid
//	  - start     必填，开始时间（RFC3339；带时区，例 2026-04-25T13:00:00-07:00）
//	  - end       必填，结束时间
//	  - snap_min  可选，快照间隔分钟，默认 10
//	  - format    可选，html (默认) / json
//	  - layout    可选，layout JSON 文件路径；缺省时从 rooms.layout_config 查
//
//	  生成的 HTML 同时写到 --out-dir/playback_<uid>_<startMMDD>_<endMMDDHHMM>.html。
//	  若该文件已存在 → 直接复用（避免重生成几百 M HTML）。
//	  format=json 时不流回 HTML，只返回 {filename, url, cached, size, ...}。
//	  format=html 时既写盘也流回（io.MultiWriter）。
//
//	GET /roomengine/playback/files?uid=<uid>   列出 out-dir 中 playback_<uid>_*.html（uid 可省=全部）
//	GET /roomengine/files/<filename>           静态文件（来自 out-dir）
//	GET /roomengine/health                     "ok"
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
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"wisefido-sensor/internal/playback"
	"wisefido-sensor/internal/roomengine"
)

const requestTimeout = 60 * time.Second

// outDir 由 --out-dir flag 设置；存放生成的 HTML（跨进程复用，避免重生成）
var outDir string

func main() {
	listen := flag.String("listen", ":7788", "HTTP listen addr")
	// 默认 ../out：典型从 owlBack/ 运行时落到 owl/out，与 kms 的 owlBack/out 隔离避免误删
	outDirFlag := flag.String("out-dir", "../out", "directory for generated HTML files (kept across runs)")
	flag.Parse()

	abs, err := filepath.Abs(*outDirFlag)
	if err != nil {
		log.Fatalf("resolve out-dir: %v", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		log.Fatalf("mkdir out-dir: %v", err)
	}
	outDir = abs

	db, err := playback.OpenDB()
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/roomengine/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/roomengine/playback", func(w http.ResponseWriter, r *http.Request) {
		handlePlayback(w, r, db)
	})
	mux.HandleFunc("/roomengine/playback/files", handleListFiles)
	mux.Handle("/roomengine/files/", http.StripPrefix("/roomengine/files/", http.FileServer(http.Dir(outDir))))

	log.Printf("roomengine-api listening on %s out-dir=%s", *listen, outDir)
	srv := &http.Server{
		Addr:              *listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// playbackFilename: playback_<uid>_<startMMDD>_<endMMDDHHMM>.html
// 不带年份：业务上同 (uid, startMMDD, endMMDDHHMM) 的同年请求被视为同一份产物（避免重生成几百 M）。
func playbackFilename(uid string, start, end time.Time, from time.Time) string {
	suffix := ""
	if !from.IsZero() {
		suffix = "_from" + from.Format("0102")
	}
	return fmt.Sprintf("playback_%s_%s_%s%s.html",
		uid,
		start.Format("0102"),
		end.Format("01021504"),
		suffix,
	)
}

func playbackURL(filename string) string { return "/roomengine/files/" + filename }

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

	// from=YYYY-MM-DD：从 history 表加载该日期 snapshot 作为演化起点（PR2）
	var fromDate time.Time
	if s := q.Get("from"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			http.Error(w, "bad from (YYYY-MM-DD): "+err.Error(), http.StatusBadRequest)
			return
		}
		fromDate = t
	}

	// dump_rect=x1,y1,x2,y2（画布坐标，cm）—— debug：跑完后 dump 矩形内 cell 统计到 JSON
	var dumpRect [4]int
	if s := q.Get("dump_rect"); s != "" {
		parts := strings.Split(s, ",")
		if len(parts) != 4 {
			http.Error(w, "dump_rect must be 'x1,y1,x2,y2'", http.StatusBadRequest)
			return
		}
		for i, p := range parts {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				http.Error(w, "dump_rect parse: "+err.Error(), http.StatusBadRequest)
				return
			}
			dumpRect[i] = v
		}
	}

	// Cache hit：文件已存在直接复用，跳过重生成（HTML 可能几百 M）
	filename := playbackFilename(uid, start, end, fromDate)
	fpath := filepath.Join(outDir, filename)
	if st, err := os.Stat(fpath); err == nil && !st.IsDir() {
		log.Printf("playback CACHE-HIT uid=%s file=%s size=%d", uid, filename, st.Size())
		if format == "json" {
			respondJSONCached(w, filename, st.Size(), start, end)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, fpath)
		return
	}

	// 解析 layout：要么从文件读，要么从 DB 查
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	roomID, cfg, err := loadLayout(ctx, db, uid, layoutPath)
	if err != nil {
		http.Error(w, "load layout: "+err.Error(), http.StatusBadRequest)
		return
	}
	roomengine.ApplyOptimizedExtent(&cfg)

	logVerdicts := q.Get("log_verdicts") == "1"

	res, err := playback.Run(ctx, playback.Options{
		RoomID:           roomID,
		Cfg:              cfg,
		DB:               db,
		DeviceUID:        uid,
		Start:            start,
		End:              end,
		SnapMin:          snapMin,
		DumpRect:         dumpRect,
		LogTrackVerdicts: logVerdicts,
		FromDate:         fromDate,
	})
	if err != nil {
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
		// 只写盘，不流回 HTML，返回 metadata
		f, ferr := os.Create(fpath)
		if ferr != nil {
			http.Error(w, "create out file: "+ferr.Error(), http.StatusInternalServerError)
			return
		}
		if err := playback.WriteHTML(f, roomID, res.Snapshots); err != nil {
			f.Close()
			os.Remove(fpath) // 写失败清掉部分文件，避免下次误命中 cache
			http.Error(w, "write html: "+err.Error(), http.StatusInternalServerError)
			return
		}
		f.Close()
		var size int64
		if st, _ := os.Stat(fpath); st != nil {
			size = st.Size()
		}
		respondJSONFresh(w, filename, size, roomID, start, end, len(res.Snapshots))
		return
	}

	// format=html：同时写盘 + 流回（MultiWriter）
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	f, ferr := os.Create(fpath)
	if ferr != nil {
		log.Printf("create out file: %v (degrading to inline-only)", ferr)
		if err := playback.WriteHTML(w, roomID, res.Snapshots); err != nil {
			log.Printf("write html: %v", err)
		}
		return
	}
	defer f.Close()
	if err := playback.WriteHTML(io.MultiWriter(w, f), roomID, res.Snapshots); err != nil {
		log.Printf("write html (multi): %v", err)
	}
}

func respondJSONCached(w http.ResponseWriter, filename string, size int64, start, end time.Time) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"filename": filename,
		"url":      playbackURL(filename),
		"cached":   true,
		"size":     size,
		"start":    start.Format(time.RFC3339),
		"end":      end.Format(time.RFC3339),
	})
}

func respondJSONFresh(w http.ResponseWriter, filename string, size int64, roomID string, start, end time.Time, snaps int) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"filename": filename,
		"url":      playbackURL(filename),
		"cached":   false,
		"size":     size,
		"room_id":  roomID,
		"start":    start.Format(time.RFC3339),
		"end":      end.Format(time.RFC3339),
		"snaps":    snaps,
	})
}

// handleListFiles: GET /roomengine/playback/files?uid=<uid>
// 列出 out-dir 中 playback_<uid>_*.html；uid 省略 → 全部 playback_*.html。
type listedFile struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	ModTime  string `json:"mod_time"`
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimSpace(r.URL.Query().Get("uid"))
	prefix := "playback_"
	if uid != "" {
		prefix += uid + "_"
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		http.Error(w, "read out-dir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]listedFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".html") {
			continue
		}
		st, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, listedFile{
			Filename: name,
			URL:      playbackURL(name),
			Size:     st.Size(),
			ModTime:  st.ModTime().Format(time.RFC3339),
		})
	}
	// 倒序：新文件在前
	sort.Slice(out, func(i, j int) bool { return out[i].ModTime > out[j].ModTime })
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"files": out})
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

	return playback.LoadDeviceRoomConfig(ctx, db, deviceUID)
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
