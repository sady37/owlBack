// Package playback 把"读取 monitor_stream/event_log → 喂给 RoomEngine grid → 抓 SVG 快照"的完整回放流程
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
	"wisefido-sensor/internal/roomengine"
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
	Start     time.Time
	End       time.Time
	SnapMin   int // 快照间隔（分钟），默认 10
	ChunkHrs  int // DB 分块查询大小（小时），默认 6
	RowLimit  int // 单块最大行数，默认 30000

	// DumpRect debug：跑完后把这个矩形（画布坐标 cm）内每个 InRoom cell 的统计灌进 Result.RectDump。
	// 4 个 0 = 不 dump。用于定位"为什么这片 cell 学成 X 类型"。
	DumpRect [4]int // [x1, y1, x2, y2]

	// LogTrackVerdicts debug：true = 每帧记录所有 track 的位置快照到 Result.TrackVerdicts。
	// 用于 ghost / fall verify 算法验证（realness/verdict 判定已迁出 TrackOutput → census/DBN）。
	LogTrackVerdicts bool

	// AlarmInjector：每个 monitor 行处理后调用，按时间戳注入 firmware Fall 报警到 engine。
	// 用于 PR-5 fall verifier 离线测试（cmd/fall-score-replay）。nil = 不注入。
	AlarmInjector AlarmInjector

	// IngestHistoricalFeedback PR-9.1：playback 启动时跑一次历史 alarm_events 反馈解析，
	// 把 ☑ 勾选直接灌进 grid.cells（GhostCount/RestZoneConfirmed/RealFallCount/FakeAlarmCount）。
	// 让 SVG 反馈层 overlay 能展示真实人类反馈。默认 false（向后兼容）。
	IngestHistoricalFeedback bool

	// Logger：自定义 zap logger 注入 TrackManager；缺省走 zap.NewDevelopment 写 stderr。
	// 用于 fall-score-replay 通过 zapcore.Tee 同时捕获 verifier 结果。
	Logger *zap.Logger

	// Cycles 重放循环数（默认 1）。> 1 时用同一段 DB 数据连续跑 N 次，simT 单调递增
	// （cycle k 时间戳 += k * windowDuration），grid 状态在 cycle 间保留累积。
	// 用途：检查 PR-15.3 时间门控（15 天）等长期累积规则；3 cycles × 3 天 ≈ 9 天等效数据。
	Cycles int

	// FromDate 非零时，从 roomengine_grid_snapshot_history (spatial_prefix /88, archive_date=FromDate)
	// 加载历史 snapshot 并 DecodeSnapshot 灌回 grid，作为演化起点（不再是空白 baseline）。
	// layout_hash 不匹配（中途人工编辑过 layout）→ 警告 + fallback baseline。
	FromDate time.Time
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

	// Grid 跑完学到的 grid（供 --persist 落 grid_snapshot 暖 live；不进 JSON/HTML）。
	Grid *roomengine.RoomGrid `json:"-"`

	// RectDump：Options.DumpRect 设置时，跑完后的 cell 统计列表（用于定位学习异常）
	RectDump []roomengine.CellStat `json:"rect_dump,omitempty"`

	// TrackVerdicts：Options.LogTrackVerdicts 启用时，记录每帧 ProcessFrame 后所有 track 的位置快照
	// 用途：debug ghost 检测——看 engine 对真假 track 的位置/still 演化。
	TrackVerdicts []TrackVerdictRecord `json:"track_verdicts,omitempty"`
}

// TrackVerdictRecord 单个 track 在某一帧的位置快照（realness/verdict 判定已迁出 TrackOutput → census/DBN）
type TrackVerdictRecord struct {
	TMs      int64 `json:"t_ms"`
	TrackID  int   `json:"track_id"`
	X        int   `json:"x"`
	Y        int   `json:"y"`
	Z        int   `json:"z"`
	StillSec int   `json:"still_sec,omitempty"`
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
	// 快照全量驻留内存（每张全房 SVG≈190KB），长窗口按固定 snap_min 会撑爆内存/磁盘
	// （7d×10min≈1008 张≈190MB）。按窗口时长抬高有效 snap_min 使快照数封顶 maxSnapshots，
	// 短窗口（用户 snap_min 已够稀）不受影响。
	const maxSnapshots = 200
	loops := opts.Cycles
	if loops < 1 {
		loops = 1
	}
	simMinutes := (opts.End.UnixMilli() - opts.Start.UnixMilli()) / 60000 * int64(loops)
	if byCap := int(simMinutes/maxSnapshots) + 1; byCap > opts.SnapMin {
		opts.SnapMin = byCap
	}

	cfg := opts.Cfg

	// 1. device_uid → device_addr（v2 时序数据按 INET /128 寻址）
	deviceAddr, err := LookupDeviceAddr(ctx, opts.DB, opts.DeviceUID)
	if err != nil {
		return nil, fmt.Errorf("lookup device_addr for %s: %w", opts.DeviceUID, err)
	}

	// 2. 构建 grid + TrackManager（与 engine.RegisterRoom 一致）
	grid := roomengine.NewRoomGrid(cfg.RoomW, cfg.RoomH, radarutils.CellSize)
	grid.OriginX = cfg.OriginX
	grid.OriginY = cfg.OriginY
	if len(cfg.WallPolygon) >= 3 {
		grid.StampRoomPolygon(cfg.WallPolygon)
	}
	grid.StampRadar(cfg.Radar)
	grid.StampEnters(cfg.Enters, cfg.EnterTargets)
	for _, z := range cfg.AreaZones {
		grid.SetPrior(z.Rect, z.AreaType, z.Conf, roomengine.SourceHuman)
	}

	tm := roomengine.NewTrackManager(opts.RoomID, grid, nil)
	decayParams := roomengine.DefaultDecayParams()
	learnParams := roomengine.DefaultLearnParams()
	tm.SetMoveSpeedCms(learnParams.MoveSpeedCms)
	tm.SetRoomName(cfg.RoomName)
	// playback 是测试场景：把 startupMs 设到回放窗起点前 10min，
	// 让"5min startup grace"完全不命中（否则窗口前 5min 的 track 全被豁免成 Real，掩盖 ghost 检测能力）。
	tm.SetStartupMs(opts.Start.UnixMilli() - 10*60*1000)
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

	// PR-2: 从 history 表灌入"从某天起步"的 grid 状态（区别于空白 baseline）
	startLabel := "T=0 baseline (layout prior only)"
	startTitleSuffix := " | T=0 baseline"
	if !opts.FromDate.IsZero() && opts.DB != nil {
		// 保证 from 注入路径无论调用方是否传 Logger 都能写出可观察日志（否则静默回退到 baseline 难排查）
		plog := opts.Logger
		if plog == nil {
			plog, _ = zap.NewDevelopment()
		}
		payload, snapHash, found, err := LookupHistorySnapshot(ctx, opts.DB, opts.RoomID, opts.FromDate)
		expectHash := roomengine.LayoutHash(cfg)
		switch {
		case err != nil:
			plog.Warn("playback history snapshot lookup failed; fallback to baseline",
				zap.String("room_id", opts.RoomID),
				zap.String("from_date", opts.FromDate.Format("2006-01-02")),
				zap.Error(err))
		case !found:
			plog.Warn("playback history snapshot not found; fallback to baseline",
				zap.String("room_id", opts.RoomID),
				zap.String("from_date", opts.FromDate.Format("2006-01-02")))
		case snapHash != expectHash:
			plog.Warn("playback history snapshot layout_hash mismatch; fallback to baseline",
				zap.String("room_id", opts.RoomID),
				zap.String("from_date", opts.FromDate.Format("2006-01-02")),
				zap.String("snap_hash", snapHash),
				zap.String("expect_hash", expectHash))
		default:
			snap, derr := roomengine.UnmarshalSnapshot(payload)
			if derr != nil {
				plog.Warn("playback history snapshot unmarshal failed; fallback to baseline",
					zap.String("room_id", opts.RoomID), zap.Error(derr))
				break
			}
			if derr := roomengine.DecodeSnapshot(snap, grid); derr != nil {
				plog.Warn("playback history snapshot decode failed; fallback to baseline",
					zap.String("room_id", opts.RoomID), zap.Error(derr))
				break
			}
			startLabel = "T=0 history " + opts.FromDate.Format("2006-01-02")
			startTitleSuffix = " | from " + opts.FromDate.Format("2006-01-02")
			plog.Info("playback started from history snapshot",
				zap.String("room_id", opts.RoomID),
				zap.String("from_date", opts.FromDate.Format("2006-01-02")))
		}
	}

	// PR-9.1: 灌入历史人类反馈到 grid（让 SVG overlay 可展示真实反馈层）
	if opts.IngestHistoricalFeedback && opts.DB != nil {
		ingestLogger := opts.Logger
		if ingestLogger == nil {
			ingestLogger, _ = zap.NewDevelopment()
		}
		if succ, total, err := IngestHistoricalFeedback(ctx, opts.DB, opts.DeviceUID,
			opts.Start, opts.End, grid, cfg.Radar, cfg.Chairs, learnParams.SitSpreadCm, ingestLogger); err != nil {
			if ingestLogger != nil {
				ingestLogger.Warn("ingest historical feedback", zap.Error(err))
			}
		} else if ingestLogger != nil {
			ingestLogger.Info("playback historical feedback ingested",
				zap.String("device_uid", opts.DeviceUID),
				zap.Int("success", succ), zap.Int("total", total))
		}
	}

	// 3. 起始快照（FromDate 注入成功时显示 history 标签；否则保持 baseline）
	snapshots := []Snapshot{
		{
			TsMs: 0,
			SVG: roomengine.BuildRoomSVG(grid, cfg.Radar, cfg.WallPolygon, cfg.Enters, opts.RoomID,
				roomengine.RoomSVGOptions{
					ShowFOV: true, Sleepads: cfg.Sleepads,
					TitleSuffix: startTitleSuffix,
					RoomName:    cfg.RoomName,
				}),
			Label: startLabel,
		},
	}

	// 4. 主回放循环
	var pathBuf []pathPt
	var simT int64
	var nextSnapAt, nextDecayAt, nextScanAt int64
	var trackVerdicts []TrackVerdictRecord
	totalRows, totalFrames := 0, 0

	cycles := opts.Cycles
	if cycles < 1 {
		cycles = 1
	}
	windowDurationMs := opts.End.UnixMilli() - opts.Start.UnixMilli()

	for cycle := 0; cycle < cycles; cycle++ {
		cycleOffsetMs := int64(cycle) * windowDurationMs
		chunkStart := opts.Start
		for chunkStart.Before(opts.End) {
			chunkEnd := chunkStart.Add(time.Duration(opts.ChunkHrs) * time.Hour)
			if chunkEnd.After(opts.End) {
				chunkEnd = opts.End
			}
			rows, err := QueryMonitorTracks(ctx, opts.DB, deviceAddr, opts.DeviceUID, chunkStart, chunkEnd, opts.RowLimit)
			if err != nil {
				return nil, fmt.Errorf("query rows %s~%s: %w",
					chunkStart.Format("01-02 15:04"), chunkEnd.Format("01-02 15:04"), err)
			}
			eventRows, eErr := QueryEvents(ctx, opts.DB, deviceAddr, opts.DeviceUID, chunkStart, chunkEnd, opts.RowLimit)
			if eErr != nil {
				return nil, fmt.Errorf("query events %s~%s: %w",
					chunkStart.Format("01-02 15:04"), chunkEnd.Format("01-02 15:04"), eErr)
			}
			rows = mergeRowsByTime(rows, eventRows)
			totalRows += len(rows)

			for _, row := range rows {
				ts := row.TimestampMs + cycleOffsetMs // PR-15.3: cycle 内时间戳 += offset，simT 单调递增

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
						grid.LearnCellAreas(learnParams, simT)
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
					for _, evt := range roomengine.ParseRadarTrackEvents(row.DataValue, row.DeviceUID, row.Category, ts) {
						tm.RecordRadarEvent(evt)
					}
					if opts.AlarmInjector != nil {
						opts.AlarmInjector(tm, ts)
					}
					tm.Tick(ts)
					simT = ts
					continue
				}

				frames := roomengine.ParseRadarTracks(row.DataValue, row.DeviceAddr, cfg.Radar, ts)
				// PR-15.3 cycle 支持：强制 frame.TMs = shifted ts（覆盖 firmware 自带 timestamp）
				for i := range frames {
					frames[i].TMs = ts
				}
				if len(frames) == 0 {
					// firmware tid=88 heartbeat 帧被 ParseRadarTracks 过滤后，仍要 tick engine
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
						realIDs[o.TrackID] = true // realness 判定单源已迁出 TrackOutput → census/DBN；轨迹 path 显示全部 track
						if opts.LogTrackVerdicts {
							trackVerdicts = append(trackVerdicts, TrackVerdictRecord{
								TMs:      ts,
								TrackID:  o.TrackID,
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
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
	}

	// 5. 终态快照
	grid.LearnCellAreas(learnParams, simT)
	grid.LearnLyingAnomalies(learnParams)
	snapshots = append(snapshots, takeSnapWithPaths(grid, cfg, opts.RoomID, simT,
		buildTrackPaths(pathBuf, simT, trackLookbackMs)))

	// 推断型 fall（lost / silent-leftbed / still）统计已随 gate-list 退役迁出 → DBN belief shadow 日志重建。
	res := &Result{
		Snapshots:   snapshots,
		TotalRows:   totalRows,
		TotalFrames: totalFrames,
		GridW:       grid.Width,
		GridH:       grid.Height,
		Grid:        grid,
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
				ShowFOV:             true,
				ShowFeedbackOverlay: true, // PR-9：渲染 cell 反馈层
				Sleepads:            cfg.Sleepads,
				TrackPaths:          paths,
				TitleSuffix:         " | " + t,
				RoomName:            cfg.RoomName,
			}),
		Label: t,
	}
}

// 给 HTML writer 用的 strings helper（避免 import strings 仅为一个 byte）
var _ = strings.TrimSpace
