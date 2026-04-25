// roomengine-playback：从 iot_timeseries 拉指定设备的近 N 小时 radar monitor 数据，
// 喂给 RoomEngine 的 Grid+TrackManager+cell_learning 流水线，每隔 snapMin 分钟"模拟时间"
// 抓一帧 SVG 快照，最后生成单文件 HTML 滑动查看器。
//
// 用法（默认参数即可跑 09E7-room101 近 48h）：
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
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"owl-common/radarutils"
	"wisefido-ai/internal/roomengine"
)

func main() {
	var (
		layoutPath  = flag.String("layout", "doc/layout-09E7-room101.json", "layout JSON 路径")
		roomID      = flag.String("room", "09E7-room101", "room id（仅作显示标识）")
		deviceUID   = flag.String("uid", "9D8A326309E7", "device_uid（雷达硬件 UID）")
		tenantID    = flag.String("tenant", "", "tenant_id（缺省自动查 devices 表）")
		hours       = flag.Int("hours", 48, "回放最近 N 小时")
		snapMin     = flag.Int("snap", 30, "快照间隔（模拟分钟）")
		outHTML     = flag.String("out", "doc/playback-09E7.html", "输出 HTML 路径")
		chunkHours  = flag.Int("chunk", 6, "DB 分块查询大小（小时）")
		rowLimit    = flag.Int("row-limit", 30000, "单块最大行数")
	)
	flag.Parse()

	ctx := context.Background()

	// ---- 1. 读 layout ----
	raw, err := os.ReadFile(*layoutPath)
	if err != nil {
		log.Fatalf("read layout: %v", err)
	}
	cfg, err := roomengine.ParseLayoutConfig(*roomID, raw)
	if err != nil {
		log.Fatalf("parse layout: %v", err)
	}
	log.Printf("layout loaded: room %s  %d×%d cm  origin=(%d,%d)  radar=(%d,%d,%d) install=%d",
		*roomID, cfg.RoomW, cfg.RoomH, cfg.OriginX, cfg.OriginY,
		cfg.Radar.Center.X, cfg.Radar.Center.Y, cfg.Radar.Center.Z, cfg.Radar.InstallModel)

	// ---- 2. 构建 Grid + TrackManager（与 engine.RegisterRoom 一致）----
	grid := roomengine.NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	if cfg.OriginX != 0 || cfg.OriginY != 0 {
		grid.OriginX = cfg.OriginX
		grid.OriginY = cfg.OriginY
	}
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}
	grid.StampRadar(cfg.Radar)
	grid.StampEnters(cfg.Enters)
	for _, r := range cfg.Enters {
		grid.SetPrior(r, roomengine.AreaEnter, 99, roomengine.SourceHuman)
	}
	for _, r := range cfg.Beds {
		grid.SetPrior(r, roomengine.AreaBed, 99, roomengine.SourceHuman)
	}
	for _, r := range cfg.Toilets {
		grid.SetPrior(r, roomengine.AreaToilet, 99, roomengine.SourceHuman)
	}
	for _, r := range cfg.Showers {
		grid.SetPrior(r, roomengine.AreaShower, 99, roomengine.SourceHuman)
	}
	for _, r := range cfg.Chairs {
		grid.SetPrior(r, roomengine.AreaSit, 80, roomengine.SourceHuman)
	}
	for _, r := range cfg.Furnitures {
		grid.SetPrior(r, roomengine.AreaDeny, 99, roomengine.SourceHuman)
	}
	for _, r := range cfg.Interferes {
		grid.SetPrior(r, roomengine.AreaDeny, 99, roomengine.SourceHuman)
	}
	tm := roomengine.NewTrackManager(*roomID, grid)

	// ---- 3. 连 DB ----
	db, err := openDB()
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// 自动查 tenant_id（如果未提供）
	if *tenantID == "" {
		t, err := lookupTenantID(ctx, db, *deviceUID)
		if err != nil {
			log.Fatalf("lookup tenant: %v", err)
		}
		*tenantID = t
		log.Printf("tenant_id auto: %s", t)
	}

	// ---- 4. 起始快照（baseline）----
	snapshots := []frame{
		{
			TsMs: 0,
			SVG: roomengine.BuildRoomSVG(grid, cfg.Radar, cfg.WallPolygon, cfg.Enters, *roomID,
				roomengine.RoomSVGOptions{ShowFOV: true, Sleepads: cfg.Sleepads,
					TitleSuffix: " | T=0 baseline"}),
			Label: "T=0 baseline (layout prior only)",
		},
	}

	// ---- 5. 分块拉取 + 回放 ----
	end := time.Now()
	start := end.Add(time.Duration(-*hours) * time.Hour)

	// 模拟时钟：随帧时间戳推进
	var simT int64 // ms
	var nextSnapAt, nextDecayAt, nextScanAt int64

	totalRows, totalFrames := 0, 0

	chunkStart := start
	for chunkStart.Before(end) {
		chunkEnd := chunkStart.Add(time.Duration(*chunkHours) * time.Hour)
		if chunkEnd.After(end) {
			chunkEnd = end
		}
		rows, err := queryRows(ctx, db, *tenantID, *deviceUID, chunkStart, chunkEnd, *rowLimit)
		if err != nil {
			log.Fatalf("query rows: %v", err)
		}
		log.Printf("chunk %s ~ %s: %d rows",
			chunkStart.Format("01-02 15:04"), chunkEnd.Format("01-02 15:04"), len(rows))
		totalRows += len(rows)

		for _, row := range rows {
			ts := row.TimestampMs

			// 首帧初始化
			if simT == 0 {
				simT = ts
				nextScanAt = ts + 5*60*1000          // 5 分钟后跑一次 cell_learning
				nextDecayAt = ts + 60*60*1000        // 1 小时后衰减
				nextSnapAt = ts + int64(*snapMin)*60_000
			}

			// 推进模拟时钟到 ts：每轮找最近的事件时间（scan/decay/snap/ts 取 min）。
			// 触发后把对应 next*At 加上各自的间隔，下轮自然不会再命中同一事件。
			for simT < ts {
				nextEvent := ts
				if nextScanAt < nextEvent {
					nextEvent = nextScanAt
				}
				if nextDecayAt < nextEvent {
					nextEvent = nextDecayAt
				}
				if nextSnapAt < nextEvent {
					nextEvent = nextSnapAt
				}
				if nextEvent <= simT {
					// 防御：万一 next*At 落到 simT 之前，强行推进 1ms 避免死循环
					nextEvent = simT + 1
				}
				simT = nextEvent

				if simT >= nextScanAt {
					grid.LearnCellAreas()
					grid.LearnLyingAnomalies()
					nextScanAt += 5 * 60 * 1000
				}
				if simT >= nextDecayAt {
					grid.DecayAll(3600, float64(roomengine.HalfLifeShort))
					nextDecayAt += 60 * 60 * 1000
				}
				if simT >= nextSnapAt {
					snapshots = append(snapshots, takeSnap(grid, cfg, *roomID, simT))
					nextSnapAt += int64(*snapMin) * 60 * 1000
				}
			}

			// 解析本行 → ProcessFrame
			frames := roomengine.ParseRadarTracks(row.DataValue, row.DeviceID, cfg.Radar, ts)
			if len(frames) > 0 {
				tm.ProcessFrame(frames)
				totalFrames += len(frames)
			}
			simT = ts
		}

		chunkStart = chunkEnd.Add(time.Millisecond)
	}

	// 终态：再跑一次 cell_learning + 末帧快照
	grid.LearnCellAreas()
	grid.LearnLyingAnomalies()
	snapshots = append(snapshots, takeSnap(grid, cfg, *roomID, simT))

	log.Printf("processed %d rows, %d valid frames, %d snapshots",
		totalRows, totalFrames, len(snapshots))

	// ---- 6. 写 HTML 查看器 ----
	if err := writeHTML(*outHTML, *roomID, snapshots); err != nil {
		log.Fatalf("write html: %v", err)
	}
	abs, _ := filepath.Abs(*outHTML)
	log.Printf("HTML viewer: %s", abs)
}

// --------------------------------------------------------------------------
// DB
// --------------------------------------------------------------------------

func openDB() (*sql.DB, error) {
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASSWORD", "postgres")
	name := getenv("DB_NAME", "owlrd")
	sslMode := getenv("DB_SSLMODE", "disable")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, sslMode)
	return sql.Open("postgres", dsn)
}

func lookupTenantID(ctx context.Context, db *sql.DB, deviceUID string) (string, error) {
	var tid string
	row := db.QueryRowContext(ctx,
		`SELECT tenant_id::text FROM devices WHERE device_uid = $1 LIMIT 1`, deviceUID)
	if err := row.Scan(&tid); err != nil {
		return "", err
	}
	return tid, nil
}

type rowRec struct {
	ID          int64
	DeviceID    string
	DeviceUID   string
	TimestampMs int64
	DataValue   interface{}
}

func queryRows(ctx context.Context, db *sql.DB, tenantID, deviceUID string,
	start, end time.Time, limit int) ([]rowRec, error) {

	q := `
SELECT its.id, its.device_id::text, its.device_uid, its."timestamp", its.data_value
FROM iot_timeseries its
WHERE its.tenant_id::text = $1
  AND its.device_uid = $2
  AND its.topic_type = 'monitor'
  AND its."timestamp" >= $3
  AND its."timestamp" <= $4
ORDER BY its."timestamp" ASC
LIMIT $5`
	rows, err := db.QueryContext(ctx, q, tenantID, deviceUID,
		start.UnixMilli(), end.UnixMilli(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []rowRec
	for rows.Next() {
		var (
			id    int64
			did   sql.NullString
			duid  sql.NullString
			tsMs  int64
			dvRaw []byte
		)
		if err := rows.Scan(&id, &did, &duid, &tsMs, &dvRaw); err != nil {
			return nil, err
		}
		var dv interface{}
		if len(dvRaw) > 0 {
			_ = json.Unmarshal(dvRaw, &dv)
		}
		out = append(out, rowRec{
			ID:          id,
			DeviceID:    did.String,
			DeviceUID:   duid.String,
			TimestampMs: tsMs,
			DataValue:   dv,
		})
	}
	return out, rows.Err()
}

// --------------------------------------------------------------------------
// 快照 + HTML
// --------------------------------------------------------------------------

type frame struct {
	TsMs  int64
	SVG   string
	Label string
}

func takeSnap(grid *roomengine.RoomGrid, cfg roomengine.RoomConfig,
	roomID string, simTMs int64) frame {
	t := time.UnixMilli(simTMs).Local().Format("2006-01-02 15:04")
	return frame{
		TsMs: simTMs,
		SVG: roomengine.BuildRoomSVG(grid, cfg.Radar, cfg.WallPolygon, cfg.Enters, roomID,
			roomengine.RoomSVGOptions{
				ShowFOV:     true,
				Sleepads:    cfg.Sleepads,
				TitleSuffix: " | " + t,
			}),
		Label: t,
	}
}

func writeHTML(path, roomID string, frames []frame) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>RoomEngine Playback: `)
	sb.WriteString(html.EscapeString(roomID))
	sb.WriteString(`</title><style>
body{margin:0;background:#1d1f23;color:#ddd;font-family:system-ui,-apple-system,sans-serif}
header{padding:8px 14px;background:#26282d;display:flex;align-items:center;gap:14px;flex-wrap:wrap}
h1{font-size:14px;margin:0;font-weight:600}
button{background:#3a8;color:#fff;border:0;padding:6px 14px;border-radius:4px;cursor:pointer;font-size:13px}
button:hover{background:#4b9}
.label{font-size:13px;font-variant-numeric:tabular-nums;color:#aac}
input[type=range]{flex:1;min-width:300px;accent-color:#3a8}
#stage{display:flex;justify-content:center;padding:10px 0;background:#1d1f23}
#stage svg{max-width:96vw;max-height:88vh}
.hint{font-size:11px;color:#888}
</style></head><body>
<header>
  <h1>Playback: `)
	sb.WriteString(html.EscapeString(roomID))
	sb.WriteString(`</h1>
  <button id="play">▶ Play</button>
  <input type="range" id="slider" min="0" max="`)
	fmt.Fprintf(&sb, "%d", len(frames)-1)
	sb.WriteString(`" value="0"/>
  <div class="label" id="counter"></div>
  <div class="label" id="ts"></div>
  <div class="hint">键盘 ← → 步进；空格 播放/暂停</div>
</header>
<div id="stage"></div>
<script>
const FRAMES = [`)
	for i, f := range frames {
		if i > 0 {
			sb.WriteString(",\n")
		}
		// 把 SVG 当字符串写入 JS（Go 的 strconv.Quote 等价 JSON.stringify on string）
		sb.WriteString(strconv.Quote(f.SVG))
	}
	sb.WriteString(`];
const LABELS = [`)
	for i, f := range frames {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Quote(f.Label))
	}
	sb.WriteString(`];
const stage=document.getElementById('stage');
const slider=document.getElementById('slider');
const counter=document.getElementById('counter');
const tsEl=document.getElementById('ts');
const playBtn=document.getElementById('play');
let playing=false;
let timer=null;

function render(i){
  i=Math.max(0,Math.min(FRAMES.length-1,i|0));
  stage.innerHTML=FRAMES[i];
  slider.value=i;
  counter.textContent='frame '+(i+1)+' / '+FRAMES.length;
  tsEl.textContent=LABELS[i];
}
slider.addEventListener('input',e=>render(+e.target.value));
playBtn.addEventListener('click',()=>{
  playing=!playing;
  playBtn.textContent=playing?'⏸ Pause':'▶ Play';
  if(playing){
    timer=setInterval(()=>{
      let v=+slider.value+1;
      if(v>=FRAMES.length){playing=false;playBtn.textContent='▶ Play';clearInterval(timer);return;}
      render(v);
    },200);
  } else { clearInterval(timer); }
});
document.addEventListener('keydown',e=>{
  if(e.key==='ArrowLeft'){render(+slider.value-1);}
  else if(e.key==='ArrowRight'){render(+slider.value+1);}
  else if(e.key===' '){playBtn.click();e.preventDefault();}
});
render(0);
</script>
</body></html>`)

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
