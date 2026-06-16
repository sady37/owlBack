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
		geom:    map[string]*roomGeom{},
		rooms:   map[string]*engine.Room{},
		firstMs: map[string]int64{},
		logger:  logger,
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

// roomGeom 一房的 config-static 几何（RegisterRoom 时从 RoomConfig 派生，OnRoomFrame 构造 FrameInput 用）。
type roomGeom struct {
	beds           []adapter.Rect
	walls          []adapter.Rect
	radarPos       adapter.Point
	sleepadPresent bool
	nb             int
}

// dbnRouter 把 Engine.OnRoomFrame 的 per-room bases 构造成 adapter.FrameInput → engine.Room.Tick（DBN 顶层）。
// 每房一份 engine.Room（懒建）；单房 rhoXroom=0（Stage A，多房 neighbor 走 engine.Unit 留 Stage B）。
type dbnRouter struct {
	mu      sync.Mutex
	geom    map[string]*roomGeom
	rooms   map[string]*engine.Room
	firstMs map[string]int64
	logger  *zap.Logger
}

func (d *dbnRouter) onRoomFrame(roomID string, bases []roomengine.TrackStatusBase, nowMs int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	g := d.geom[roomID]
	if g == nil {
		return // 无 layout placeholder 房：无几何不进 DBN
	}

	tracks := make([]adapter.TrackObs, 0, len(bases))
	sleepadInBed := false
	for _, b := range bases {
		tracks = append(tracks, adapter.TrackObs{RadarTrack: adapter.RadarTrack{
			Online: true, Pose: b.Pose, X: b.X, Y: b.Y, Z: b.Z,
			StillSec: float64(b.StillBoxSec),
		}})
		if b.SleepadInBed {
			sleepadInBed = true
		}
	}

	reading := belief.BedNoReport
	if g.sleepadPresent {
		if sleepadInBed {
			reading = belief.BedInBed
		} else {
			reading = belief.BedLeftBed
		}
	}
	sleepads := make([]adapter.SleepadFrame, g.nb)
	covers := make([]float64, g.nb)
	onbed := make([]float64, g.nb)
	overlap := make([]float64, g.nb)
	for j := 0; j < g.nb; j++ {
		sleepads[j] = adapter.SleepadFrame{Present: g.sleepadPresent, Reading: reading}
		covers[j], onbed[j], overlap[j] = 1, 1, 1
	}

	if _, ok := d.firstMs[roomID]; !ok {
		d.firstMs[roomID] = nowMs
	}
	aloneMin := float64(nowMs-d.firstMs[roomID]) / 60000.0

	fi := adapter.FrameInput{
		NowMs:    nowMs,
		Tracks:   tracks,
		Sleepads: sleepads,
		Beds:     g.beds,
		Covers:   covers,
		Onbed:    onbed,
		Overlap:  overlap,
		Walls:    g.walls,
		RadarPos: g.radarPos,
		Census:   adapter.Census{AloneContinuousMin: aloneMin},
	}

	room := d.rooms[roomID]
	if room == nil {
		room = engine.NewRoom(adapter.BedGeoms(fi), g.nb)
		d.rooms[roomID] = room
	}
	fr := room.Tick(fi, 0)

	top, tp := fr.Probe.MarginalS.Max()
	if fr.Decision.Fire {
		d.logger.Info("xsensor_dbn_fire",
			zap.String("room_id", roomID), zap.Int64("ts", nowMs),
			zap.Float64("p_fallen", fr.Probe.PFallen), zap.Int("n_r", fr.Decision.PeopleCount),
			zap.Int("top_s", int(top)), zap.Float64("top_p", tp),
			zap.String("band", fr.Decision.Band), zap.Float64("lambda", fr.Probe.Lambda))
	}
	d.logger.Debug("xsensor_belief",
		zap.String("room_id", roomID), zap.Int64("ts", nowMs),
		zap.Int("tracks", len(tracks)), zap.Float64("p_fallen", fr.Probe.PFallen),
		zap.Bool("fire", fr.Decision.Fire), zap.String("band", fr.Decision.Band),
		zap.Int("top_s", int(top)), zap.Float64("lambda", fr.Probe.Lambda))
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

func rectsFrom(rs []radarutils.Rect) []adapter.Rect {
	out := make([]adapter.Rect, len(rs))
	for i, r := range rs {
		out[i] = adapter.Rect{X1: r.X1, Y1: r.Y1, X2: r.X2, Y2: r.Y2}
	}
	return out
}
