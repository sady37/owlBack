package main

import (
	"context"
	"database/sql"
	"fmt"

	"owlBack/tools/Xsensorv1/internal/config"
	"owlBack/tools/Xsensorv1/internal/roomengine"
	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/engine"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// startRoomEngine 建 + 配 + 注册房间 + 接 DBN 顶层回调 + 启动消费循环。
// 适配版（vs wisefido-sensor engine_bootstrap）：去掉已剥离的 AI publish / 每日 reload /
// ghost adjudicator / Persister（馈送层无持久化无投影，决策归 DBN）。
func startRoomEngine(ctx context.Context, cfg *config.Config, db *sql.DB, rdb *redis.Client,
	router *dbnRouter, logger *zap.Logger) (*roomengine.Engine, error) {

	eng := roomengine.NewEngine(rdb, logger)
	eng.Configure(buildRuntimeConfig(cfg))

	census := roomengine.NewSuiteCensusManager(rdb, roomengine.DefaultSuiteCensusConfig(), logger)
	eng.SetSuiteCensus(census)

	registered, err := registerAllRooms(ctx, eng, db, router, logger)
	if err != nil {
		return nil, fmt.Errorf("register rooms: %w", err)
	}
	logger.Info("xsensor: rooms registered", zap.Int("count", registered))

	mapped, err := mapDevicesToRooms(ctx, eng, db, logger)
	if err != nil {
		return nil, fmt.Errorf("map devices: %w", err)
	}
	logger.Info("xsensor: devices mapped", zap.Int("count", mapped))

	// 按 unitKey(suiteID) 分组 rooms → 建 engine.Unit（跨房 hand-off neighbor 轴）。
	// 单房 unit 无兄弟 → ρ_xroom≡0 → 与单房 Room 逐帧等价（cd2b 零回归保持）。
	unitRooms := map[string]map[string]*engine.Room{}
	for roomID, unitKey := range router.roomUnit {
		if unitRooms[unitKey] == nil {
			unitRooms[unitKey] = map[string]*engine.Room{}
		}
		unitRooms[unitKey][roomID] = router.rooms[roomID]
	}
	for unitKey, rooms := range unitRooms {
		router.units[unitKey] = engine.NewUnit(rooms, 1) // residentCount=1（单住户测试；多住户后续）
	}
	logger.Info("xsensor: units built", zap.Int("units", len(router.units)), zap.Int("rooms", len(router.rooms)))

	eng.OnRoomFrame = router.onRoomFrame

	go func() {
		if err := eng.Run(ctx); err != nil {
			logger.Error("roomengine.Run exited", zap.Error(err))
		}
	}()
	return eng, nil
}

func buildRuntimeConfig(cfg *config.Config) roomengine.RuntimeConfig {
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
	}
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
	return rc
}

func registerAllRooms(ctx context.Context, eng *roomengine.Engine, db *sql.DB,
	router *dbnRouter, logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text                                       AS room_id,
		       r.room_name,
		       COALESCE(r.room_type, 0)                              AS room_type,
		       COALESCE(r.is_public_bathroom, FALSE)                 AS is_public_bathroom,
		       CASE
		         WHEN r.room_type = 1 AND COALESCE(r.is_public_bathroom, FALSE)
		           THEN r.room_id::text
		         ELSE host(set_masklen(r.room_id, 80))::text || '/80'
		       END                                                   AS suite_id,
		       COALESCE(u.timezone, '')                              AS timezone,
		       host(set_masklen(r.room_id, 48))::text || '/48'       AS tenant_pref,
		       (SELECT ru.resident_id::text
		        FROM resident_unit ru
		        WHERE ru.valid_to IS NULL
		          AND (ru.spatial_prefix >>= r.room_id
		               OR ru.spatial_prefix <<= r.room_id
		               OR ru.spatial_prefix = r.room_id)
		        ORDER BY masklen(ru.spatial_prefix) DESC
		        LIMIT 1)                                              AS resident_id
		FROM rooms r
		LEFT JOIN units u ON u.unit_id >>= r.room_id
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	canvasesByRoom, err := roomengine.LoadRoomCanvases(ctx, db, logger)
	if err != nil {
		return 0, err
	}

	count := 0
	for rows.Next() {
		var roomID, roomName, suiteID, tenantID string
		var roomType int
		var isPublicBathroom bool
		var residentIDOpt sql.NullString
		var timezone string
		if err := rows.Scan(&roomID, &roomName, &roomType, &isPublicBathroom,
			&suiteID, &timezone, &tenantID, &residentIDOpt); err != nil {
			logger.Warn("scan rooms row", zap.Error(err))
			continue
		}
		residentID := ""
		if residentIDOpt.Valid {
			residentID = residentIDOpt.String
		}

		cfg, hasLayout := roomengine.BuildRoomConfigFromCanvases(roomID, canvasesByRoom[roomID], logger)
		if hasLayout {
			cfg.RoomName = roomName
			cfg.Timezone = timezone
			roomengine.ApplyOptimizedExtent(&cfg)
		} else {
			cfg = roomengine.RoomConfig{RoomID: roomID, RoomName: roomName, Timezone: timezone}
		}
		cfg.RoomType = roomType
		cfg.IsPublicBathroom = isPublicBathroom
		cfg.SuiteID = suiteID
		cfg.ResidentID = residentID

		eng.RegisterRoom(cfg)
		eng.SetRoomTenant(roomID, tenantID)
		if hasLayout {
			nb := len(cfg.Beds)
			router.geom[roomID] = &roomGeom{
				beds:           rectsFrom(cfg.Beds),
				walls:          wallsFromPolygon(cfg.WallPolygon),
				radarPos:       adapter.Point{X: cfg.Radar.Center.X, Y: cfg.Radar.Center.Y},
				sleepadPresent: len(cfg.Sleepads) > 0,
				nb:             nb,
			}
			seed := adapter.FrameInput{Beds: rectsFrom(cfg.Beds), Covers: ones(nb), Onbed: ones(nb), Overlap: ones(nb)}
			router.rooms[roomID] = engine.NewRoom(adapter.BedGeoms(seed), nb)
			unitKey := cfg.SuiteID // 同 unit(/80) 的房互为兄弟（跨房 hand-off）；public bathroom suiteID=自身/128 独立
			if unitKey == "" {
				unitKey = roomID
			}
			router.roomUnit[roomID] = unitKey
		}
		count++
	}
	return count, rows.Err()
}

func mapDevicesToRooms(ctx context.Context, eng *roomengine.Engine, db *sql.DB,
	logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT host(d.device_addr)::text AS device_addr,
		       r.room_id::text          AS room_id,
		       COALESCE(dfm.device_type::text, '') AS device_type,
		       dfm.device_uid AS device_uid_hex
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		JOIN rooms r                  ON r.room_id >>= d.device_addr
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var deviceAddr, roomID, deviceType, deviceUIDHex string
		if err := rows.Scan(&deviceAddr, &roomID, &deviceType, &deviceUIDHex); err != nil {
			logger.Warn("scan devices row", zap.Error(err))
			continue
		}
		if roomID == "" || deviceAddr == "" {
			continue
		}
		eng.MapDeviceToRoom(deviceAddr, roomID)
		if deviceType != "" {
			eng.MapDeviceAddrToType(deviceAddr, deviceType)
		}
		if deviceUIDHex != "" {
			eng.MapDeviceAddrToUID(deviceAddr, deviceUIDHex)
		}
		count++
	}
	return count, rows.Err()
}
