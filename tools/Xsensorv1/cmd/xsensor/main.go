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
	"time"

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
		roomTZ:   map[string]*time.Location{},
		mm:       map[string]*roomengine.RoomMM{},
		logger:   logger,
	}

	eng, err := startRoomEngine(ctx, cfg, db, rdb, router, logger)
	if err != nil {
		logger.Fatal("roomengine startup failed", zap.Error(err))
	}
	router.eng = eng // P1：onRoomFrame 每 tick 回注 radar 折叠减量给 RealPeopleInRoom

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

// fwBed 固件 area_id 是否命中床区（N；与 roomengine.RadarBedStates 同判据，forensic 用）。0/255=哨兵不算。
func fwBed(areaID int, bedAreaIDs []int) bool {
	for _, id := range bedAreaIDs {
		if id != 0 && id != 255 && areaID == id {
			return true
		}
	}
	return false
}

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
	rooms    map[string]*engine.Room       // roomID → Room（bootstrap 建，供 NewUnit 分组）
	units    map[string]*engine.Unit       // unitKey(suiteID) → 多房编排器（跨房 hand-off）
	roomUnit map[string]string             // roomID → unitKey
	roomType map[string]int                // roomID → card.RoomType（1=Bathroom）→ UD timer deadline
	roomTZ   map[string]*time.Location     // roomID → 时区（IsNightTime 算 risktime，缩短 floor tFloor）
	mm       map[string]*roomengine.RoomMM // roomID → 房级静态 MM（samebed prior 权威，吸纳读）；nil=无床/无设备
	eng      *roomengine.Engine            // 回注 radar 折叠减量给 P1 占用人数（RealPeopleInRoom，cutover 后服务 zoneengine）
	logger   *zap.Logger
}

func (d *dbnRouter) onRoomFrame(roomID string, bases []roomengine.TrackStatusBase, bed card.BedState, nowMs int64, exitLogOdds, ghostLeftLogOdds func(logicID string, atMs int64) float64) (fired, dropped []string) {
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
			StillSec: float64(b.StillBoxSec), // still-box raw 时长 → FloorGuard 纯计时器（直立折扣已移 emission 压 SFallen）

			AreaType: int(b.CellAreaType), // 每帧读活的 cell area（emission 正向压制 + floor 阈）
			InChair:  b.InChair, ChairMu: b.ChairMu, ChairSigma: b.ChairSigma, // chair 区 dwell 分布 → floor 连续 tFloor
			RoomType: d.roomType[roomID],  // 房型 → still CDF room×cell 保守合并(bathroom 一律 20min)
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
	vitalPresent := false // 房级:任一 sleepad 在床 + HR/RR fresh(SnapshotTrackStatuses 算,塞每 base)
	for _, b := range bases {
		if b.SleepadVitalPresent {
			vitalPresent = true
			break
		}
	}
	for j := 0; j < g.nb; j++ {
		sleepads[j] = adapter.SleepadFrame{Present: g.sleepadPresent, Reading: reading, VitalPresent: vitalPresent}
	}

	// covers/onbed/overlap 不再走 per-tick FrameInput（Tick 路径不消费）：真 covers 现 per-(设备×床)，
	//   bootstrap SetDeviceGeom 一次注入冻结进各设备 Coupling/Emission（MM 床耦合，§K）。
	// risktime(夜间)：缩短 floor 时长兜底阈(纯时间轴,不进 C_FN)。IsNightTime 用本房时区。
	isRiskTime := roomengine.IsNightTime(nowMs, d.roomTZ[roomID])
	fi := adapter.FrameInput{
		NowMs:       nowMs,
		Tracks:      tracks,
		Sleepads:    sleepads,
		Beds:        g.beds,
		BedAreaIDs:  g.bedAreaIDs,
		RadarLess:   g.radarLess,
		Walls:       g.walls,
		RadarPos:    g.radarPos,
		Entrances:   g.entrances,
		ExitLogOdds:      exitLogOdds,
		GhostLeftLogOdds: ghostLeftLogOdds,
		Census:      adapter.Census{Night: isRiskTime},
	}

	u := d.units[d.roomUnit[roomID]]
	if u == nil {
		return nil, nil // room 不属任何 unit（无 layout placeholder）
	}
	fr := u.Tick(roomID, fi) // 多房编排：算 ρ_xroom（兄弟房守恒+时间窗）→ Room.Tick（单房无兄弟 ρ=0）
	// P1 占用人数：把本房 census 折叠后真人数（fr.Decision.PeopleCount=Nr）回注 Engine
	//   （RealPeopleInRoom 优先用；cutover 后服务 zoneengine total_people）。
	realPeople := -1
	var padAbs []roomengine.PadAbsorption
	if d.eng != nil {
		pads := d.eng.SnapshotSleepads(roomID, nowMs)            // sleepad 占用身份（addr/InBed/Fresh/Stale）
		radars := roomengine.RadarBedStates(bases, g.bedAreaIDs) // 每台 radar N-in-bed（MN/FwAreaID，§9.1 raw）
		var uncovered int
		uncovered, padAbs = roomengine.AbsorbSleepads(pads, radars, d.mm[roomID], roomengine.SamebedAbsorbThresh)
		// P1 单 setter（契约 §3 / §9.3 裁定）：Nr 折叠 + uncovered-sleepad（撑占用，修缺陷①漏算）。
		//   只进 RealPeopleInRoom(P1 占用/alone)；P2 census.Nr→C_FN 不动（睡着的人不进 fall 风险，§9.4）。
		d.eng.SetRoomRadarPeople(roomID, fr.Decision.PeopleCount+uncovered)
		for _, t := range fr.Tracks {
			d.eng.SetTrackPReal(roomID, t.LogicID, t.PReal) // ghost 单源回灌：census PReal → TrackState（cell 占用/床区学习读）
		}
		realPeople = d.eng.RealPeopleInRoom(roomID)
		// de-absorption no-silent-caps（§9.4）：主 radar 离床+垫仍 InBed 时 LOG，标明 V5/V6 stale/fresh 仲裁
		//   在 09e7 **未被真 case 覆盖**（铁律 [[validate_real_case_no_unit_tests]] 不靠推理）。
		for _, p := range padAbs {
			if p.RadarLeftBed {
				d.logger.Warn("de_absorption: 主 radar 离床+sleepad 仍 InBed（V5/V6 stale/fresh 仲裁=启发式，09e7 未覆盖，no-silent-caps）",
					zap.String("room", roomID), zap.String("pad_lid", p.LogicID),
					zap.Bool("fresh", p.Fresh), zap.Bool("stale", p.Stale), zap.Bool("counted_uncovered", p.Uncovered))
			}
		}
	}
	handoffL := u.LastHandoffL(roomID)
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
			"area": int(b.CellAreaType), "bf_preal": b.PReal,
			"fw_area": b.FwAreaID, "fw_bed": fwBed(b.FwAreaID, g.bedAreaIDs), // 固件 area_id + 是否命中床区(N,抬 SBed 那条腿)
			"vital": b.SleepadVitalPresent, // 该轨 sleepad 接触 vital(InBed+HR/RR fresh)→ couplesAnyBed 时抬 SBed
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
			"repeat_r": t.RepeatR, "self_recovered": t.SelfRecovered, "still_sec": t.StillSec,
			"s_marg": t.SMarg, "covers": t.Covers, // per-track S 边缘 + per-device covers(MM)：看 D523 自身 SBed + covers=0
		})
	}
	pad := make([]map[string]interface{}, 0, len(padAbs))
	for _, p := range padAbs {
		pad = append(pad, map[string]interface{}{
			"pad_lid": p.LogicID, "pad_present": p.InBed, "pad_fresh": p.Fresh, "pad_stale": p.Stale,
			"pad_absorbed_by": p.AbsorbedBy, "pad_samebed": p.Samebed,
			"pad_uncovered": p.Uncovered, "pad_radar_left_bed": p.RadarLeftBed,
		})
	}
	fuse := make([]map[string]interface{}, 0, len(fr.FuseSync))
	for _, f := range fr.FuseSync {
		fuse = append(fuse, map[string]interface{}{
			"a": f.A, "b": f.B, "agree": f.Agree, "moves": f.Moves, "same": f.Same, // 双雷达同人运动同步检查
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
		zap.Int("real_people", realPeople),               // P1：折叠后占用人数（cutover 后 → zoneengine total_people）
		zap.Int("rescuable", fr.Decision.RescuableCount), // 可救援数（present-real ∧ S≠Bed，folded）；forensic，不门控 fire
		zap.Float64("p_fallen", fr.Probe.PFallen), zap.String("band", fr.Decision.Band),
		zap.Bool("fire", fr.Decision.Fire), zap.Float64("lambda", fr.Probe.Lambda),
		zap.String("top_s", sName(int(top))), zap.Float64("top_p", tp),
		zap.Float64("handoff_logodds", handoffL),              // §7.7 v2 hand-off SLeft 注入对数似然（0=无接力；>0=矩形核命中）
		zap.Bool("lost_real", fr.LostReal),                    // 本帧 confirmed 真人 track 消失 = hand-off 源候选（lostAt 将被设）
		zap.Float64("gained_real", fr.GainedReal),             // 本帧新现真人后验 = hand-off 宿候选（守恒重现）
		zap.Int64("pending_lost_ms", u.PendingLostMs(roomID)), // 本房待解析 lost 时戳（>0=已注册,handoffLFor 在找接力）
		zap.Int("sibling_gains", u.SiblingGainCount()),        // unit 内当前窗口跨房 gain 数（0=无人在别房现身）
		zap.Bool("unit_has_track", unitHasTrack), zap.Bool("has_neighbor", hasNeighbor),
		zap.Bool("lost_exited", fr.LostExited),
		zap.String("bed_reading", bedReadingName(reading)), zap.Bool("bed_present", g.sleepadPresent),
		zap.Bool("vital_present", vitalPresent), // 房级 sleepad 接触 vital(HR/RR fresh)→ covers=1 设备抬 SBed
		zap.Bool("risktime", isRiskTime),        // 夜间风险时段(缩短 floor tFloor;不进 C_FN)
		zap.Any("s_dist", sDist), zap.Any("target", raw), zap.Any("dbn", dbn), zap.Any("pad", pad),
		zap.Any("fuse", fuse), // 双雷达同人运动同步对(agree/same/moves)
		zap.Any("walls", walls), zap.Int("radar_x", g.radarPos.X), zap.Int("radar_y", g.radarPos.Y))

	return fr.FiredLogicIDs, fr.DroppedLogicIDs // fired→复位 still-box；dropped(确认离场/空)→evict track,停 coast re-feed
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
