// xsensor — replay-重算 sensor（替换 Tsensor）：消费 test:* 流，复用 wisefido-sensor 馈送派生
// （track_manager/layout/cell/mirror），把 DBN 四轴 engine.Room 当**顶层唯一决策**，fire 落 Xsensor.log。
//
// 生产 = wisefido-sensor（线上）；本进程在 replay-重算道（非生产）：喂 test:*（仿 iot:*），
// sensor 以为在跑生产，实把重放当生产重算。layout/rooms/devices 启动从 DB 只读加载（同 Tsensor）。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"owlBack/tools/Xsensorv1/internal/config"
	"owlBack/tools/Xsensorv1/internal/roomengine"
	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
	"owlBack/tools/Xsensorv1/internal/roomengine/engine"

	"owl-common/card"
	"owl-common/radarutils"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}

	logCfg := zap.NewProductionConfig()
	if os.Getenv("LOG_LEVEL") == "debug" {
		logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logCfg.OutputPaths = []string{"/home/wisefido/owl/log/Xsensor.log"}
	logCfg.ErrorOutputPaths = []string{"/home/wisefido/owl/log/Xsensor.log"}
	logger, err := logCfg.Build()
	if err != nil {
		panic(fmt.Sprintf("init logger: %v", err))
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, rdb, err := openDeps(cfg)
	if err != nil {
		logger.Fatal("deps init failed — refusing to start", zap.Error(err))
	}

	router := &dbnRouter{
		geom:     map[string]*roomGeom{},
		rooms:    map[string]*engine.Room{},
		units:    map[string]*engine.Unit{},
		roomUnit: map[string]string{},
		roomType: map[string]int{},
		unitPub:  map[string]belief.UnitPublicness{},
		logger:   logger,
	}

	eng, err := startRoomEngine(ctx, cfg, db, rdb, router, logger)
	if err != nil {
		logger.Fatal("roomengine startup failed", zap.Error(err))
	}
	_ = eng

	logger.Info("xsensor started — DBN 顶层裁决 over test:*; fire → Xsensor.log")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Info("shutting down", zap.String("signal", sig.String()))
	cancel()
}

func openDeps(cfg *config.Config) (*sql.DB, *redis.Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, nil, fmt.Errorf("ping db: %w", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}
	return db, rdb, nil
}

// poseLying observation.PoseLying = 6（sleepad-only 房合成 bed-track 的姿态）。
const poseLying = 6

// roomGeom 一房的 config-static 几何（RegisterRoom 时从 RoomConfig 派生，OnRoomFrame 构造 FrameInput 用）。
type roomGeom struct {
	beds           []adapter.Rect
	bedAreaIDs     []int // 与 beds 一一对齐：firmware radar.areas 床 area_id（N 判定）
	walls          []adapter.Rect
	entrances      []adapter.Rect // §9.3① enter 区（门，areaType=4）→ 出生地距门 D 软发射
	radarPos       adapter.Point
	sleepadPresent bool
	radarLess      bool // 无雷达 layout（sleepad-only 房）→ 无 S 轴，B 轴靠合成 bed-track
	nb             int
}

// dbnRouter 把 Engine.OnRoomFrame 的 per-room bases 构造成 adapter.FrameInput → engine.Room.Tick（DBN 顶层）。
// 每房一份 engine.Room（懒建）；单房 rhoXroom=0（Stage A，多房 neighbor 走 engine.Unit 留 Stage B）。
type dbnRouter struct {
	mu       sync.Mutex
	geom     map[string]*roomGeom
	rooms    map[string]*engine.Room          // roomID → Room（bootstrap 建，供 NewUnit 分组）
	units    map[string]*engine.Unit          // unitKey(suiteID) → 多房编排器（跨房 hand-off）
	roomUnit map[string]string                // roomID → unitKey
	roomType map[string]int                   // roomID → card.RoomType（1=Bathroom）→ UD timer deadline
	unitPub  map[string]belief.UnitPublicness // unitKey(suiteID) → 公共度（units.unit_type，定找人窗 W）
	logger   *zap.Logger
}

func (d *dbnRouter) onRoomFrame(roomID string, bases []roomengine.TrackStatusBase, bed card.BedState, nowMs int64, exitLogOdds func(trackID int, atMs int64) float64) (fired, dropped []int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	g := d.geom[roomID]
	if g == nil {
		return nil, nil // 无信号源(无 layout 无 sleepad)→ 不进 DBN
	}

	// B 轴读数 = room 级权威 bed 状态(sleepad+radar 床事件融合，带时戳)。治本 bed-reading：
	//   BedConfidence=0 无数据=NoReport（区别于 LeftBed）/ BedStatus=1=LeftBed / 否则=InBed。
	reading := belief.BedNoReport
	if g.sleepadPresent {
		switch {
		case bed.BedConfidence == 0:
			reading = belief.BedNoReport
		case bed.BedStatus == 1:
			// LeftBed 来源差异化(BedOccupancyState 用 conf 编码):sleepad 接触(90)果断清/radar 几何(20)弱清。
			reading = belief.BedLeftBed
			if bed.BedConfidence < 50 {
				reading = belief.BedLeftBedRadar
			}
		default:
			reading = belief.BedInBed
		}
	}

	tracks := make([]adapter.TrackObs, 0, len(bases)+1)
	for _, b := range bases {
		tracks = append(tracks, adapter.TrackObs{LogicID: b.LogicID, RadarTrack: adapter.RadarTrack{
			TrackID: b.TrackID, // logicID↔track_id 反查源（ExitRoom 按号反查丢轨人）
			Online:  b.Present, Pose: b.Pose, X: b.X, Y: b.Y, Z: b.Z,
			StillSec:  float64(b.StillBoxSec), // still-box raw 时长 → FloorGuard 纯计时器（直立折扣已移 emission 压 SFallen）

			AreaType: int(b.CellAreaType), // 每帧读活的 cell area（emission 正向压制 + floor 阈）
			RoomType: d.roomType[roomID],  // 房型 → still CDF room×cell 保守合并(bathroom 未画 toilet 也用 bathsec)
			FwAreaID: b.FwAreaID,          // firmware area_id（present=本帧/lost=冻结）→ 命中床 areaId = N（在床）
		}})
	}
	// sleepad-only 房(无雷达 track)：InBed 合成一条 bed-track 作 B 轴载体(engine.Room track-centric，
	//   无 track 无载体)。pose=Lying + 床占用 → S=Bed（无摔判定，sleepad 看不到姿态）；
	//   LeftBed → 不合成 → 上一帧 bed-track 缺席 → blind→S→Left → LostReal hand-off 源（人离本房床去隔壁）。
	if g.radarLess && reading == belief.BedInBed {
		fwArea := 0
		if len(g.bedAreaIDs) > 0 {
			fwArea = g.bedAreaIDs[0] // 自洽：合成 bed-track 的 area_id = 床声明 id → 命中 N → 压 SBed
		}
		tracks = append(tracks, adapter.TrackObs{LogicID: "sleepad-bed", RadarTrack: adapter.RadarTrack{
			Online: true, Pose: poseLying, X: 0, Y: 0, Z: 0,
			FwAreaID: fwArea, // sleepad-only 床占用 → N 命中 → SBed（无摔判定，sleepad 看不到姿态）
		}})
	}

	sleepads := make([]adapter.SleepadFrame, g.nb)
	covers := make([]float64, g.nb)
	onbed := make([]float64, g.nb)
	overlap := make([]float64, g.nb)
	for j := 0; j < g.nb; j++ {
		sleepads[j] = adapter.SleepadFrame{Present: g.sleepadPresent, Reading: reading}
		covers[j], onbed[j], overlap[j] = 1, 1, 1
	}

	fi := adapter.FrameInput{
		NowMs:      nowMs,
		Tracks:     tracks,
		Sleepads:   sleepads,
		Beds:       g.beds,
		BedAreaIDs: g.bedAreaIDs,
		RadarLess:  g.radarLess,
		Covers:     covers,
		Onbed:      onbed,
		Overlap:    overlap,
		Walls:      g.walls,
		RadarPos:   g.radarPos,
		Entrances:  g.entrances,
		ExitLogOdds: exitLogOdds,
	}

	u := d.units[d.roomUnit[roomID]]
	if u == nil {
		return nil, nil // room 不属任何 unit（无 layout placeholder）
	}
	fr := u.Tick(roomID, fi) // 多房编排：算 ρ_xroom（兄弟房守恒+时间窗）→ Room.Tick（单房无兄弟 ρ=0）
	rho := u.LastRho(roomID)
	unitHasTrack, hasNeighbor := u.UnitState(nowMs)

	top, tp := fr.Probe.MarginalS.Max()

	// 每 tick 全景 X 光：S 9 态全分布 + B 轴 bed + room(N_r/presentCount) + per-track(raw + DBN realness/门控)。
	sDist := map[string]float64{}
	for i, v := range fr.Probe.MarginalS {
		sDist[sName(i)] = v
	}
	raw := make([]map[string]interface{}, 0, len(bases))
	for _, b := range bases {
		raw = append(raw, map[string]interface{}{
			"tid": b.TrackID, "present": b.Present, "pose": b.Pose,
			"x": b.X, "y": b.Y, "z": b.Z, "stillbox": b.StillBoxSec,
			"area": int(b.CellAreaType), "verdict": int(b.Verdict),
		})
	}
	dbn := make([]map[string]interface{}, 0, len(fr.Tracks))
	for _, t := range fr.Tracks {
		dbn = append(dbn, map[string]interface{}{
			"lid": t.LogicID, "present": t.Present, "p_real": t.PReal, "p_mirror": t.PMirror,
			"is_refl": t.IsReflection, "p_fallen": t.PFallen,
			"fire": t.Fire, "band": t.Band,
			"x": t.X, "y": t.Y, "sep": t.Sep, "wall_margin": t.WallMargin, "rho": t.Rho, "later": t.LaterBorn,
			"door_d": t.DoorD, "track_confidence": int(t.PReal*100 + 0.5),
			"repeat_r": t.RepeatR, "self_recovered": t.SelfRecovered,
		})
	}
	walls := make([][4]int, 0, len(g.walls))
	for _, w := range g.walls {
		walls = append(walls, [4]int{w.X1, w.Y1, w.X2, w.Y2})
	}
	d.logger.Info("xsensor_xray",
		zap.String("room", roomID), zap.Int64("ts", nowMs),
		zap.Int("n_r", fr.Decision.PeopleCount), zap.Int("present_cnt", fr.PresentCount),
		zap.Int("raw_tracks", len(bases)),
		zap.Float64("p_fallen", fr.Probe.PFallen), zap.String("band", fr.Decision.Band),
		zap.Bool("fire", fr.Decision.Fire), zap.Float64("lambda", fr.Probe.Lambda),
		zap.String("top_s", sName(int(top))), zap.Float64("top_p", tp),
		zap.Float64("rho_xroom", rho),
		zap.Bool("unit_has_track", unitHasTrack), zap.Bool("has_neighbor", hasNeighbor),
		zap.Bool("lost_exited", fr.LostExited),
		zap.String("bed_reading", bedReadingName(reading)), zap.Bool("bed_present", g.sleepadPresent),
		zap.Float64s("covers", covers), zap.Float64s("onbed", onbed),
		zap.Any("s_dist", sDist), zap.Any("target", raw), zap.Any("dbn", dbn),
		zap.Any("walls", walls), zap.Int("radar_x", g.radarPos.X), zap.Int("radar_y", g.radarPos.Y))

	return fr.FiredTrackIDs, fr.DroppedTrackIDs // fired→复位 still-box；dropped(确认离场/空)→evict track,停 coast re-feed
}

// bedReadingName B 轴 sleepad 读数名（X 光可读）。
func bedReadingName(r belief.BedReading) string {
	switch r {
	case belief.BedInBed:
		return "InBed"
	case belief.BedLeftBed:
		return "LeftBed"
	default:
		return "NoReport"
	}
}

// sName S 轴态名（9 态）。
func sName(i int) string {
	n := []string{"Empty", "Bed", "Sit", "OpenFloor", "Bath", "Fallen", "BlindRest", "BlindOpen", "Left"}
	if i >= 0 && i < len(n) {
		return n[i]
	}
	return fmt.Sprintf("S%d", i)
}

// wallsFromPolygon 闭合多边形顶点 → 墙段矩形（adapter.IsReflection 用：radar→ghost 连线求交）。
func wallsFromPolygon(poly []radarutils.Point) []adapter.Rect {
	if len(poly) < 2 {
		return nil
	}
	out := make([]adapter.Rect, 0, len(poly))
	for i := 0; i+1 < len(poly); i++ {
		out = append(out, adapter.Rect{X1: poly[i].X, Y1: poly[i].Y, X2: poly[i+1].X, Y2: poly[i+1].Y})
	}
	if last := poly[len(poly)-1]; last != poly[0] {
		out = append(out, adapter.Rect{X1: last.X, Y1: last.Y, X2: poly[0].X, Y2: poly[0].Y})
	}
	return out
}

// ones 长度 n 的全 1 切片（NewRoom 种子 BedGeom 的 Covers/Onbed/Overlap；真 covers 走 RadarBedReachCount 后续）。
func ones(n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = 1
	}
	return s
}

func rectsFrom(rs []radarutils.Rect) []adapter.Rect {
	out := make([]adapter.Rect, len(rs))
	for i, r := range rs {
		out[i] = adapter.Rect{X1: r.X1, Y1: r.Y1, X2: r.X2, Y2: r.Y2}
	}
	return out
}
