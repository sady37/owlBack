// engine_bootstrap.go
//
// 把 RoomEngine 接入 wisefido-ai 实时流：
//   1. 从 rooms 表加载所有 layout_config，注册到 Engine（自动 ApplyOptimizedExtent）
//   2. 从 devices 表读 device_uid/device_id → bound_room_id 路由表
//   3. Configure RuntimeConfig（DecayParams + LearnParams + ParamSets + Persister）
//   4. Engine.Run(ctx) 在 goroutine 跑 —— 自动消费 iot:monitor:stream + iot:event:stream
//
// 故意不接 Phase 5（家属 ground-truth 反馈）：RecordGroundTruth 只是被动 API，
// 不调它就不跑；winner 选择会按 baseline 默认 balanced，不影响实时学习。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/go-redis/redis/v8"

	"wisefido-ai/internal/config"
	"wisefido-ai/internal/roomengine"
)

// startRoomEngine 创建 + 配置 + 注册房间 + 启动主循环；返回的 cancelFunc 用于优雅关闭。
func startRoomEngine(ctx context.Context, cfg *config.Config, db *sql.DB,
	rdb *redis.Client, logger *zap.Logger) (*roomengine.Engine, error) {

	engine := roomengine.NewEngine(rdb, logger)

	// 1. 注入 yaml 运行时参数 + Persister
	engine.Configure(buildRuntimeConfig(cfg, db))

	// 2. 注册所有有 layout 的房间
	registered, err := registerAllRooms(ctx, engine, db, logger)
	if err != nil {
		return nil, fmt.Errorf("register rooms: %w", err)
	}
	logger.Info("roomengine: rooms registered", zap.Int("count", registered))

	// 3. 建立 device_uid / device_id → room_id 路由表
	mapped, err := mapDevicesToRooms(ctx, engine, db, logger)
	if err != nil {
		return nil, fmt.Errorf("map devices: %w", err)
	}
	logger.Info("roomengine: devices mapped", zap.Int("count", mapped))

	// 4. 启动主循环（消费 monitor + event 流，跑学习+持久化定时器）
	go func() {
		if err := engine.Run(ctx); err != nil {
			logger.Error("roomengine.Run exited with error", zap.Error(err))
		}
	}()

	return engine, nil
}

// buildRuntimeConfig 把 yaml RoomEngineConfig 转换为 engine 内部 RuntimeConfig
func buildRuntimeConfig(cfg *config.Config, db *sql.DB) roomengine.RuntimeConfig {
	r := cfg.RoomEngine

	rc := roomengine.RuntimeConfig{
		Decay: roomengine.DecayParams{
			ImmediateSec: float64(r.Decay.ImmediateSec),
			WalkSec:      float64(r.Decay.WalkSec),
			SitSec:       float64(r.Decay.SitSec),
			LieSec:       float64(r.Decay.LieSec),
			EventSec:     float64(r.Decay.EventSec),
		},
		Learn: roomengine.LearnParams{
			WalkActiveX10:     r.Learn.WalkActiveX10,
			WalkTraverse:      r.Learn.WalkTraverse,
			SitActiveX10:      r.Learn.SitActiveX10,
			LieAnomalyX10:     r.Learn.LieAnomalyX10,
			BedToleranceCm:    r.Learn.BedToleranceCm,
			ToiletToleranceCm: r.Learn.ToiletToleranceCm,
			ShowerToleranceCm: r.Learn.ShowerToleranceCm,
			ConfFloor:         r.Belief.ConfidenceFloor,
			ConfFull:          r.Belief.ConfidenceFull,
			MoveSpeedCms:      r.Learn.MoveSpeedCms,
			NearTraverseWalk:  r.Learn.NearTraverseWalk,
			NearTraverseDeny:  r.Learn.NearTraverseDeny,
		},
		DecayInterval:      time.Duration(r.Belief.DecayIntervalSec) * time.Second,
		BeliefScanInterval: time.Duration(r.Belief.ScanIntervalSec) * time.Second,
		WinnerEvalInterval: time.Duration(r.Belief.WinnerEvalSec) * time.Second,
		SnapshotInterval:   time.Duration(r.Persist.SnapshotIntervalSec) * time.Second,
	}

	// ParamSets：3 组并行参数（保守/中庸/激进）
	if len(r.ParamSets) >= 3 {
		for i := 0; i < 3; i++ {
			rc.ParamSets[i] = roomengine.ParamSet{
				Name:   r.ParamSets[i].Name,
				Alpha:  r.ParamSets[i].Alpha,
				Beta:   r.ParamSets[i].Beta,
				FlipTh: r.ParamSets[i].FlipTh,
			}
		}
	}

	// Persister：开启时建 PostgresPersister
	if r.Persist.Enabled && r.Persist.Storage == "postgres" {
		rc.Persister = roomengine.NewPostgresPersister(db, r.Persist.Table)
	}

	return rc
}

// registerAllRooms 扫 rooms 表所有 layout_config，逐个 ParseLayoutConfig + Optimize + RegisterRoom。
// 解析失败的房间跳过并记日志，不阻塞其他房间。
func registerAllRooms(ctx context.Context, engine *roomengine.Engine, db *sql.DB,
	logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT room_id::text, layout_config::text
		FROM rooms
		WHERE layout_config IS NOT NULL
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var roomID string
		var layoutStr sql.NullString
		if err := rows.Scan(&roomID, &layoutStr); err != nil {
			logger.Warn("scan rooms row", zap.Error(err))
			continue
		}
		if !layoutStr.Valid || layoutStr.String == "" {
			continue
		}

		cfg, err := roomengine.ParseLayoutConfig(roomID, []byte(layoutStr.String))
		if err != nil {
			logger.Warn("parse layout failed", zap.String("room_id", roomID), zap.Error(err))
			continue
		}
		roomengine.ApplyOptimizedExtent(&cfg)
		engine.RegisterRoom(cfg)
		count++
	}
	return count, rows.Err()
}

// mapDevicesToRooms 建 device → room 路由表。
//
// 路由解析顺序（per device）：
//  1. devices.bound_room_id 直接绑（雷达常见）
//  2. fallback: devices.bound_bed_id → beds.room_id（sleepad 常见，自己绑床而非房间）
//
// **不**用 card_id 路由：一个 card 可能跨多 room（例：sleepad + 2 雷达分属不同房间），
// 用 card 推 room 会路由到错误房间。
func mapDevicesToRooms(ctx context.Context, engine *roomengine.Engine, db *sql.DB,
	logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT d.device_id::text,
		       d.device_uid,
		       COALESCE(d.bound_room_id, b.room_id)::text AS room_id
		FROM devices d
		LEFT JOIN beds b ON b.bed_id = d.bound_bed_id
		WHERE d.bound_room_id IS NOT NULL
		   OR (d.bound_bed_id IS NOT NULL AND b.room_id IS NOT NULL)
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var deviceID, deviceUID, roomID string
		if err := rows.Scan(&deviceID, &deviceUID, &roomID); err != nil {
			logger.Warn("scan devices row", zap.Error(err))
			continue
		}
		if roomID == "" {
			continue
		}
		if deviceID != "" {
			engine.MapDeviceToRoom(deviceID, roomID)
		}
		if deviceUID != "" {
			engine.MapDeviceToRoom(deviceUID, roomID)
		}
		count++
	}
	return count, rows.Err()
}
