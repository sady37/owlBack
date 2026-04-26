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
}

// Result 回放结果（snapshots + 统计）
type Result struct {
	Snapshots   []Snapshot `json:"snapshots"`
	TotalRows   int        `json:"total_rows"`
	TotalFrames int        `json:"total_frames"`
	GridW       int        `json:"grid_w"` // 优化后的 grid 尺寸
	GridH       int        `json:"grid_h"`
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

			frames := roomengine.ParseRadarTracks(row.DataValue, row.DeviceID, cfg.Radar, ts)
			if len(frames) > 0 {
				outputs := tm.ProcessFrame(frames)
				totalFrames += len(frames)
				realIDs := make(map[int]bool, len(outputs))
				for _, o := range outputs {
					if o.Verdict == roomengine.VerdictReal {
						realIDs[o.TrackID] = true
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

	return &Result{
		Snapshots:   snapshots,
		TotalRows:   totalRows,
		TotalFrames: totalFrames,
		GridW:       grid.Width,
		GridH:       grid.Height,
	}, nil
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
