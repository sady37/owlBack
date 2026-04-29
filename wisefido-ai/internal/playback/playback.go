// Package playback 把"读取 iot_timeseries → 喂给 RoomEngine grid → 抓 SVG 快照"的完整回放流程
// 抽成可复用单元。被 cmd/roomengine-playback（写 HTML 文件）和 cmd/roomengine-api（HTTP 返回）共用。
//
// 设计：函数式接口（Options in / Result out），不维护任何运行时状态；可在 HTTP handler 里安全并发调用。
package playback

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"owl-common/radarutils"
	"wisefido-ai/internal/roomengine"
)

// Snapshot 一帧快照（SVG 字符串 + 元信息）
type Snapshot struct {
	TsMs  int64  `json:"ts_ms"`
	SVG   string `json:"svg"`
	Label string `json:"label"`
}

// Options 回放配置
type Options struct {
	RoomID    string             // 显示标识（HTML 标题等）
	Cfg       roomengine.RoomConfig // 已解析的 layout（调用方负责 ParseLayoutConfig）
	DB        *sql.DB
	DeviceUID string
	TenantID  string // 可空；空时从 devices 表查
	Start     time.Time
	End       time.Time
	SnapMin   int // 快照间隔（分钟），默认 10
	ChunkHrs  int // DB 分块查询大小（小时），默认 6
	RowLimit  int // 单块最大行数，默认 30000

	// DumpRect debug：跑完后把这个矩形（画布坐标 cm）内每个 InRoom cell 的统计灌进 Result.RectDump。
	// 4 个 0 = 不 dump。用于定位"为什么这片 cell 学成 X 类型"。
	DumpRect [4]int // [x1, y1, x2, y2]

	// LogTrackVerdicts debug：true = 每帧记录所有 track 的 verdict / anomaly / score 到 Result.TrackVerdicts。
	// 用于 ghost / fall verify 算法验证。
	LogTrackVerdicts bool

	// AlarmInjector：每个 monitor 行处理后调用，按时间戳注入 firmware Fall 报警到 engine。
	// 用于 PR-5 fall verifier 离线测试（cmd/fall-score-replay）。nil = 不注入。
	AlarmInjector AlarmInjector

	// Logger：自定义 zap logger 注入 TrackManager；缺省走 zap.NewDevelopment 写 stderr。
	// 用于 fall-score-replay 通过 zapcore.Tee 同时捕获 verifier 结果。
	Logger *zap.Logger
}

// AlarmInjector playback 在每帧 nowMs 推进后调用；
// impl 内部按 nowMs 把 [prev, now] 之间所有 firmware Fall 注入 tm.RecordRadarAlarm。
type AlarmInjector func(tm *roomengine.TrackManager, nowMs int64)

// Result 回放结果（snapshots + 统计）
type Result struct {
	Snapshots   []Snapshot `json:"snapshots"`
	TotalRows   int        `json:"total_rows"`
	TotalFrames int        `json:"total_frames"`
	GridW       int        `json:"grid_w"` // 优化后的 grid 尺寸
	GridH       int        `json:"grid_h"`
	// Silent Fall 60s 挂起机制统计（由 track_manager 暴露）
	SilentFallPendingCreated   int `json:"silent_fall_pending_created"`
	SilentFallPendingCancelled int `json:"silent_fall_pending_cancelled"`
	SilentFallReported         int `json:"silent_fall_reported"`
	SilentFallOutstanding      int `json:"silent_fall_outstanding"`

	// Lost Fall 挂起机制统计（cell-area-typed wait + ExitRoom + 多人 cancel）
	LostFallPendingCreated   int `json:"lost_fall_pending_created"`
	LostFallPendingCancelled int `json:"lost_fall_pending_cancelled"`
	LostFallReported         int `json:"lost_fall_reported"`
	LostFallOutstanding      int `json:"lost_fall_outstanding"`

	// Silent Fall LeftBed（PR-2c 新版：sleepad LeftBed + radar 仍在 Bed 邻域）
	SilentFallLeftBedReported  int `json:"silent_fall_leftbed_reported"`
	SilentFallLeftBedCancelled int `json:"silent_fall_leftbed_cancelled"`

	// Still Fall（PR-3：bathroom + Stand 静止 ≥ 15/18min）
	StillFallReported int `json:"still_fall_reported"`

	// RectDump：Options.DumpRect 设置时，跑完后的 cell 统计列表（用于定位学习异常）
	RectDump []roomengine.CellStat `json:"rect_dump,omitempty"`

	// TrackVerdicts：Options.LogTrackVerdicts 启用时，记录每帧 ProcessFrame 后所有 track 的 verdict / anomaly
	// 用途：debug ghost 检测——看 engine 对真假 track 是否给出正确判定。
	TrackVerdicts []TrackVerdictRecord `json:"track_verdicts,omitempty"`
}

// TrackVerdictRecord 单个 track 在某一帧的判定快照
type TrackVerdictRecord struct {
	TMs      int64  `json:"t_ms"`
	TrackID  int    `json:"track_id"`
	Verdict  string `json:"verdict"` // pending / real / ghost
	Score    int    `json:"score"`
	Risk     int    `json:"risk"`
	Anomaly  string `json:"anomaly,omitempty"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Z        int    `json:"z"`
	StillSec int    `json:"still_sec,omitempty"`
}

// 历史轨迹窗口（与原 playback 一致）
const (
	trackLookbackMs int64 = 30 * 60 * 1000 // 渲染最近 30 min 轨迹
	pathBufKeepMs   int64 = 60 * 60 * 1000 // pathBuf 内存上限保 1h
)

// Run 执行一次回放：从 DB 拉数据 → grid + cell_learning → 收集 snapshots
//
// 调用前提：opts.Cfg 已 ApplyOptimizedExtent 过（grid 范围已优化）；调用方负责
// （cmd/roomengine-playback 和 HTTP handler 都需要做）。
func Run(ctx context.Context, opts Options) (*Result, error) {
	if opts.SnapMin <= 0 {
		opts.SnapMin = 10
	}
	if opts.ChunkHrs <= 0 {
		opts.ChunkHrs = 6
	}
	if opts.RowLimit <= 0 {
		opts.RowLimit = 30000
	}

	cfg := opts.Cfg

	// 1. 自动查 tenant_id
	if opts.TenantID == "" {
		t, err := LookupTenantID(ctx, opts.DB, opts.DeviceUID)
		if err != nil {
			return nil, fmt.Errorf("lookup tenant_id for %s: %w", opts.DeviceUID, err)
		}
		opts.TenantID = t
	}

	// 2. 构建 grid + TrackManager（与 engine.RegisterRoom 一致）
	grid := roomengine.NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	grid.OriginX = cfg.OriginX
	grid.OriginY = cfg.OriginY
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

	tm := roomengine.NewTrackManager(opts.RoomID, grid)
	decayParams := roomengine.DefaultDecayParams()
	learnParams := roomengine.DefaultLearnParams()
	tm.SetMoveSpeedCms(learnParams.MoveSpeedCms)
	tm.SetRoomName(cfg.RoomName)
	// 让 engine 内部 ai.log 走 stderr，方便 playback 实测看 ghost / fall 决策细节
	if opts.Logger != nil {
		tm.SetLogger(opts.Logger)
	} else if logger, err := zap.NewDevelopment(); err == nil {
		tm.SetLogger(logger)
	}
	// 注入时区（IsNightTime 用）；playback 调用方应在 opts.Cfg.Timezone 中设置 IANA 字符串。
	if cfg.Timezone != "" {
		if loc, err := time.LoadLocation(cfg.Timezone); err == nil {
			tm.SetTimezone(loc)
		}
	}

	// 3. 起始快照
	snapshots := []Snapshot{
		{
			TsMs: 0,
			SVG: roomengine.BuildRoomSVG(grid, cfg.Radar, cfg.WallPolygon, cfg.Enters, opts.RoomID,
				roomengine.RoomSVGOptions{
					ShowFOV: true, Sleepads: cfg.Sleepads,
					TitleSuffix: " | T=0 baseline",
				}),
			Label: "T=0 baseline (layout prior only)",
		},
	}

	// 4. 主回放循环
	var pathBuf []pathPt
	var simT int64
	var nextSnapAt, nextDecayAt, nextScanAt int64
	var trackVerdicts []TrackVerdictRecord
	totalRows, totalFrames := 0, 0

	chunkStart := opts.Start
	for chunkStart.Before(opts.End) {
		chunkEnd := chunkStart.Add(time.Duration(opts.ChunkHrs) * time.Hour)
		if chunkEnd.After(opts.End) {
			chunkEnd = opts.End
		}
		rows, err := QueryRows(ctx, opts.DB, opts.TenantID, opts.DeviceUID, chunkStart, chunkEnd, opts.RowLimit)
		if err != nil {
			return nil, fmt.Errorf("query rows %s~%s: %w",
				chunkStart.Format("01-02 15:04"), chunkEnd.Format("01-02 15:04"), err)
		}
		// 同时拉 event 流（radar EnterRoom/ExitRoom/InBed/LeftBed），按 timestamp merge 进 monitor 流。
		// birth filter 检查 EnterRoom 配对依赖此数据。
		eventRows, eErr := QueryEvents(ctx, opts.DB, opts.TenantID, opts.DeviceUID, chunkStart, chunkEnd, opts.RowLimit)
		if eErr != nil {
			return nil, fmt.Errorf("query events %s~%s: %w",
				chunkStart.Format("01-02 15:04"), chunkEnd.Format("01-02 15:04"), eErr)
		}
		rows = mergeRowsByTime(rows, eventRows)
		totalRows += len(rows)

		for _, row := range rows {
			ts := row.TimestampMs

			if simT == 0 {
				simT = ts
				nextScanAt = ts + 5*60*1000
				nextDecayAt = ts + 60*60*1000
				nextSnapAt = ts + int64(opts.SnapMin)*60_000
			}

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
					nextEvent = simT + 1
				}
				simT = nextEvent

				if simT >= nextScanAt {
					grid.LearnCellAreas(learnParams)
					grid.LearnLyingAnomalies(learnParams)
					nextScanAt += 5 * 60 * 1000
				}
				if simT >= nextDecayAt {
					grid.DecayAll(3600, decayParams)
					nextDecayAt += 60 * 60 * 1000
				}
				if simT >= nextSnapAt {
					snapshots = append(snapshots, takeSnapWithPaths(grid, cfg, opts.RoomID, simT,
						buildTrackPaths(pathBuf, simT, trackLookbackMs)))
					nextSnapAt += int64(opts.SnapMin) * 60 * 1000
				}
			}

			// 分发：event 流（radar EnterRoom/ExitRoom/InBed/LeftBed）落账 + Tick；不参与 ProcessFrame
			if row.TopicType == "event" {
				for _, evt := range roomengine.ParseRadarTrackEvents(row.DataValue, row.DeviceUID, ts) {
					tm.RecordRadarEvent(evt)
				}
				if opts.AlarmInjector != nil {
					opts.AlarmInjector(tm, ts)
				}
				tm.Tick(ts)
				simT = ts
				continue
			}

			frames := roomengine.ParseRadarTracks(row.DataValue, row.DeviceID, cfg.Radar, ts)
			if len(frames) == 0 {
				// firmware tid=88 heartbeat 帧被 ParseRadarTracks 过滤后，仍要 tick engine
				// 否则之前活着的 track 的 MissCount 不递增，永远不会进入消失判定 → lost-fall pending 也不会创建。
				// 与 prod engine.go 同 bug；此处先在 playback 层修，后续再看 prod 是否同改。
				if opts.AlarmInjector != nil {
					opts.AlarmInjector(tm, ts)
				}
				tm.Tick(ts)
				simT = ts
				continue
			}
			{
				if opts.AlarmInjector != nil {
					opts.AlarmInjector(tm, ts)
				}
				outputs := tm.ProcessFrame(frames)
				totalFrames += len(frames)
				realIDs := make(map[int]bool, len(outputs))
				for _, o := range outputs {
					if o.Verdict == roomengine.VerdictReal {
						realIDs[o.TrackID] = true
					}
					if opts.LogTrackVerdicts {
						trackVerdicts = append(trackVerdicts, TrackVerdictRecord{
							TMs:      ts,
							TrackID:  o.TrackID,
							Verdict:  verdictName(o.Verdict),
							Score:    o.Score,
							Risk:     o.Risk,
							Anomaly:  anomalyName(o.Anomaly),
							X:        o.X,
							Y:        o.Y,
							Z:        o.Z,
							StillSec: o.StillSec,
						})
					}
				}
				for _, f := range frames {
					if !realIDs[f.TrackID] {
						continue
					}
					pathBuf = append(pathBuf, pathPt{tid: f.TrackID, x: f.X, y: f.Y, tms: f.TMs})
				}
				if cutoff := ts - pathBufKeepMs; len(pathBuf) > 0 && pathBuf[0].tms < cutoff {
					i := 0
					for i < len(pathBuf) && pathBuf[i].tms < cutoff {
						i++
					}
					pathBuf = pathBuf[i:]
				}
			}
			simT = ts
		}

		chunkStart = chunkEnd.Add(time.Millisecond)
		// 主动检查 ctx 取消（HTTP 客户端断开等）
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	// 5. 终态快照
	grid.LearnCellAreas(learnParams)
	grid.LearnLyingAnomalies(learnParams)
	snapshots = append(snapshots, takeSnapWithPaths(grid, cfg, opts.RoomID, simT,
		buildTrackPaths(pathBuf, simT, trackLookbackMs)))

	stats := tm.SilentFallStatsSnapshot()
	lostStats := tm.LostFallStatsSnapshot()
	leftBedStats := tm.SilentFallLeftBedStatsSnapshot()
	stillStats := tm.StillFallStatsSnapshot()
	res := &Result{
		Snapshots:                  snapshots,
		TotalRows:                  totalRows,
		TotalFrames:                totalFrames,
		GridW:                      grid.Width,
		GridH:                      grid.Height,
		SilentFallPendingCreated:   stats.PendingCreated,
		SilentFallPendingCancelled: stats.PendingCancelled,
		SilentFallReported:         stats.Reported,
		SilentFallOutstanding:      stats.Outstanding,
		LostFallPendingCreated:     lostStats.PendingCreated,
		LostFallPendingCancelled:   lostStats.PendingCancelled,
		LostFallReported:           lostStats.Reported,
		LostFallOutstanding:        lostStats.Outstanding,
		SilentFallLeftBedReported:  leftBedStats.Reported,
		SilentFallLeftBedCancelled: leftBedStats.Cancelled,
		StillFallReported:          stillStats.Reported,
	}
	// 可选：dump 矩形内 cell 统计（debug 学习异常）
	if opts.DumpRect != [4]int{0, 0, 0, 0} {
		res.RectDump = grid.DumpRect(opts.DumpRect[0], opts.DumpRect[1],
			opts.DumpRect[2], opts.DumpRect[3])
	}
	if opts.LogTrackVerdicts {
		res.TrackVerdicts = trackVerdicts
	}
	return res, nil
}

// verdictName / anomalyName：把 enum 转可读字符串（JSON 输出用）
func verdictName(v roomengine.TrackVerdict) string {
	switch v {
	case roomengine.VerdictPending:
		return "pending"
	case roomengine.VerdictReal:
		return "real"
	case roomengine.VerdictGhost:
		return "ghost"
	}
	return "unknown"
}

func anomalyName(a roomengine.Anomaly) string {
	switch a {
	case roomengine.AnomalyNone:
		return ""
	case roomengine.AnomalyFall:
		return "fall"
	case roomengine.AnomalyStillTooLong:
		return "still_too_long"
	case roomengine.AnomalyPathBreak:
		return "path_break"
	case roomengine.AnomalyPoseMismatch:
		return "pose_mismatch"
	case roomengine.AnomalyBedFall:
		return "bed_fall"
	case roomengine.AnomalyBedsideFall:
		return "bedside_fall"
	}
	return "unknown"
}

// pathPt 单帧 track 位置记录
type pathPt struct {
	tid  int
	x, y int
	tms  int64
}

func buildTrackPaths(buf []pathPt, nowMs, lookbackMs int64) []roomengine.TrackPath {
	if len(buf) == 0 {
		return nil
	}
	cutoff := nowMs - lookbackMs
	byTID := make(map[int][]radarutils.Point)
	for _, p := range buf {
		if p.tms < cutoff || p.tms > nowMs {
			continue
		}
		byTID[p.tid] = append(byTID[p.tid], radarutils.Point{X: p.x, Y: p.y})
	}
	out := make([]roomengine.TrackPath, 0, len(byTID))
	for tid, pts := range byTID {
		if len(pts) < 2 {
			continue
		}
		out = append(out, roomengine.TrackPath{TrackID: tid, Points: pts})
	}
	return out
}

func takeSnapWithPaths(grid *roomengine.RoomGrid, cfg roomengine.RoomConfig,
	roomID string, simTMs int64, paths []roomengine.TrackPath) Snapshot {
	t := time.UnixMilli(simTMs).Local().Format("2006-01-02 15:04")
	return Snapshot{
		TsMs: simTMs,
		SVG: roomengine.BuildRoomSVG(grid, cfg.Radar, cfg.WallPolygon, cfg.Enters, roomID,
			roomengine.RoomSVGOptions{
				ShowFOV:     true,
				Sleepads:    cfg.Sleepads,
				TrackPaths:  paths,
				TitleSuffix: " | " + t,
			}),
		Label: t,
	}
}

// 给 HTML writer 用的 strings helper（避免 import strings 仅为一个 byte）
var _ = strings.TrimSpace
