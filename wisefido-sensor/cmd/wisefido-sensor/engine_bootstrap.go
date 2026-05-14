// engine_bootstrap.go
//
// 把 RoomEngine 接入 wisefido-sensor 实时流：
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

	"wisefido-sensor/internal/config"
	"wisefido-sensor/internal/roomengine"
)

// startRoomEngine 创建 + 配置 + 注册房间 + 启动主循环；返回的 cancelFunc 用于优雅关闭。
func startRoomEngine(ctx context.Context, cfg *config.Config, db *sql.DB,
	rdb *redis.Client, logger *zap.Logger) (*roomengine.Engine, error) {

	engine := roomengine.NewEngine(rdb, logger)

	// 1. 注入 yaml 运行时参数 + Persister
	engine.Configure(buildRuntimeConfig(cfg, db))

	// 1b. 注入 AI publish 单点配置（mode + source）；
	//     mode="log" 仅写 ai.log，"log&publish" 推到 iot:event/alarm:stream。
	engine.SetAIPublishConfig(cfg.AIPublish.Mode, cfg.AIPublish.Source)
	logger.Info("roomengine: ai_publish configured",
		zap.String("mode", cfg.AIPublish.Mode),
		zap.String("source", cfg.AIPublish.Source))

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

	// 3a. 注入 Stay alarm 启用状态（still fall 第三条 bathroom 路径）
	stayRooms, err := loadStayAlarmEnablement(ctx, engine, db, logger)
	if err != nil {
		logger.Warn("roomengine: load Stay alarm enablement failed", zap.Error(err))
	} else {
		logger.Info("roomengine: rooms with Stay alarm enabled", zap.Int("count", stayRooms))
	}

	// 3b. PR-15：每日 22:00 (local) 重读 layout，hash 变即重置该 room → 从 0 重学
	engine.SetDailyLayoutReload(22, db)

	// 3c. 路由表周期热加载（60s）——修复"启动后才绑定的 device 永远沉默"bug。
	// reloader 闭包共享同一个 db handle，在线运行时安全调用。
	engine.SetRoutesReloader(func(rctx context.Context) error {
		_, err := mapDevicesToRooms(rctx, engine, db, logger)
		return err
	}, 60*time.Second)

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
		RiskTime: roomengine.RiskTimeConfig{
			NightStartH: r.RiskTime.NightStartH,
			NightStartM: r.RiskTime.NightStartM,
			NightEndH:   r.RiskTime.NightEndH,
			NightEndM:   r.RiskTime.NightEndM,
		},
		BedsideFall: roomengine.BedsideFallConfig{
			WindowSec:       r.BedsideFall.WindowSec,
			BedsideMarginCm: r.BedsideFall.BedsideMarginCm,
			StillTimeoutSec: r.BedsideFall.StillTimeoutSec,
		},
		// alarm_events false_alarm 反馈链：用同一个 *sql.DB；间隔 5min（默认）。
		FeedbackDB:       db,
		FeedbackInterval: 5 * time.Minute,
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
		// History persister：复用同 DB；写到 37_*.sql 历史表（每天 11:50 归档，保留 365 天）
		rc.HistoryPersister = roomengine.NewPostgresHistoryPersister(db, "")
	}

	return rc
}

// registerAllRooms 扫 rooms 表所有 room，逐个 ParseLayoutConfig + Optimize + RegisterRoom；
// layout 为空或解析失败的 room 仍以"无 layout placeholder"注册，让 sleepad-only 流量
// （ProcessSleepadBedEvent 不依赖 grid）能正常路由到 TrackManager，避免被 dropped_unrouted_message 沉默。
//
// v2 schema: rooms 表无 tenant_id/unit_id/layout_config 列；layout 在 room_visual_layout 表（PK=spatial_prefix）；
// tenant_id/unit_id 由 room_id INET prefix 派生（/48 / /80）；timezone 从 units LPM (unit /80 contains room /88) 取。
func registerAllRooms(ctx context.Context, engine *roomengine.Engine, db *sql.DB,
	logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text,
		       r.room_name,
		       COALESCE(rvl.canvas::text, '') AS layout_config,
		       COALESCE(u.timezone, '') AS timezone,
		       host(set_masklen(r.room_id, 48))::text || '/48' AS tenant_pref
		FROM rooms r
		LEFT JOIN units u ON u.unit_id >>= r.room_id
		LEFT JOIN room_visual_layout rvl ON rvl.spatial_prefix = r.room_id
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	placeholderCount := 0
	for rows.Next() {
		var roomID, roomName, tenantID string
		var layoutStr sql.NullString
		var timezone string
		if err := rows.Scan(&roomID, &roomName, &layoutStr, &timezone, &tenantID); err != nil {
			logger.Warn("scan rooms row", zap.Error(err))
			continue
		}

		var cfg roomengine.RoomConfig
		if !layoutStr.Valid || layoutStr.String == "" {
			// 无 layout：minimal placeholder。RegisterRoom 内部 RoomW/RoomH=0 会兜底成 Max；
			// WallPolygon/Radar/Beds 等全空 → grid 全是空 cell，雷达派生事件天然走不通；
			// sleepad ProcessSleepadBedEvent 仅用 tm 内部 maps，不查 grid，可正常工作。
			cfg = roomengine.RoomConfig{RoomID: roomID, RoomName: roomName, Timezone: timezone}
			placeholderCount++
		} else {
			cfg, err = roomengine.ParseLayoutConfig(roomID, []byte(layoutStr.String))
			if err != nil {
				// 解析失败也降级成 placeholder，与"无 layout"同等待遇——保 sleepad 能通
				logger.Warn("parse layout failed; registering as placeholder",
					zap.String("room_id", roomID), zap.Error(err))
				cfg = roomengine.RoomConfig{RoomID: roomID, RoomName: roomName, Timezone: timezone}
				placeholderCount++
			} else {
				cfg.RoomName = roomName
				cfg.Timezone = timezone
				roomengine.ApplyOptimizedExtent(&cfg)
			}
		}
		engine.RegisterRoom(cfg)
		// PR-8: 注入 roomID → tenant_id（AI 派生 alarm 发布需要）
		engine.SetRoomTenant(roomID, tenantID)
		count++
	}
	if placeholderCount > 0 {
		logger.Info("roomengine: registered rooms without layout (sleepad-only fallback)",
			zap.Int("placeholder_count", placeholderCount))
	}
	return count, rows.Err()
}

// loadStayAlarmEnablement 列出所有 Stay alarm 启用的 room_id 下发到对应 TrackManager。
//
// 用途：still fall position 的并集第三条 — 运维显式启用 Stay alarm 的房间，
// 即使既不是 cell.AreaToilet/Shower 也不是 room.name=bathroom，也按 bathroom 处理。
//
// v2 schema: 旧 alarm_device.monitor_config 已退役；alarm 启用配置在 spatial_config 表
// (config_key='alarm.cloud_config'，按 tenant /48 存放 device_alarms.Radar.Stay.is_enabled)。
// LPM 解析后，rooms 落入启用 Stay 的 tenant prefix → 全部 room 算启用。
//
// 该状态只在启动时灌一次；运行时配置变更需重启或后续接 Redis 通道（暂未做）。
func loadStayAlarmEnablement(ctx context.Context, engine *roomengine.Engine, db *sql.DB,
	logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT r.room_id::text
		FROM rooms r
		JOIN spatial_config sc
		  ON sc.spatial_prefix >>= r.room_id
		 AND sc.config_key = 'alarm.cloud_config'
		WHERE COALESCE(
		    (sc.config_value->'device_alarms'->'Radar'->'Stay'->>'is_enabled')::int,
		    0
		) = 1
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var roomID string
		if err := rows.Scan(&roomID); err != nil {
			logger.Warn("scan stay enablement row", zap.Error(err))
			continue
		}
		if roomID == "" {
			continue
		}
		engine.SetRoomStayAlarmEnabled(roomID, true)
		count++
	}
	return count, rows.Err()
}

// mapDevicesToRooms 建 device_addr → room 路由表。
//
// v2 schema: 旧 devices.bound_room_id / bound_bed_id 列已退役；device 与 room 关系
// 由 device_ipv6 INET prefix 派生（room_id /88 contains device_ipv6 /128）。
// device_factory_meta (dfm) 提供 device_type；旧 device_store 表已退役。
//
// device_ipv6 单程票（doc/device_ipv6_migration_checklist.md Phase D）后 engine 内部
// map key 为 canonical IPv6 字符串（addr.String()），与 envelope DeviceAddr 一致。
func mapDevicesToRooms(ctx context.Context, engine *roomengine.Engine, db *sql.DB,
	logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT host(d.device_ipv6)::text AS device_addr,
		       r.room_id::text          AS room_id,
		       COALESCE(dfm.device_type::text, '') AS device_type
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		JOIN rooms r                  ON r.room_id >>= d.device_ipv6
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var deviceAddr, roomID, deviceType string
		if err := rows.Scan(&deviceAddr, &roomID, &deviceType); err != nil {
			logger.Warn("scan devices row", zap.Error(err))
			continue
		}
		if roomID == "" || deviceAddr == "" {
			continue
		}
		// 单程票：addr 是唯一 device key（替代旧的 deviceID UUID + deviceUID MAC 双键）
		engine.MapDeviceToRoom(deviceAddr, roomID)
		// 注册 device_addr → device_type，AI publish 用作 envelope.DeviceType
		if deviceType != "" {
			engine.MapDeviceIDToType(deviceAddr, deviceType)
		}
		// MapDeviceIDToUID 单程票后是 identity（addr→addr），无意义，跳过
		count++
	}
	return count, rows.Err()
}

