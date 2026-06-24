// engine_bootstrap.go
//
// 把 RoomEngine 接入 wisefido-sensor 实时流：
//  1. 从 rooms 表加载所有 layout_config，注册到 Engine（自动 ApplyOptimizedExtent）
//  2. 从 devices 表读 device_uid/device_addr → bound_room_id 路由表
//  3. Configure RuntimeConfig（DecayParams + LearnParams + ParamSets + Persister）
//  4. Engine.Run(ctx) 在 goroutine 跑 —— 自动消费 iot:monitor:stream + iot:event:stream
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
	"wisefido-sensor/internal/roomengine/adapter"
	dbnengine "wisefido-sensor/internal/roomengine/engine"
)

// startRoomEngine 创建 + 配置 + 注册房间 + 启动主循环；返回的 cancelFunc 用于优雅关闭。
//
// sensor_v2 wiring（详 sensor_v2.md §6.A 设计原则 #3 + sensor_v2_known_limitations.md L4）：
//   - SuiteCensus: PR-2 单例进程级 census
//   - GhostAdjudicators: PR-4 接口 + PR-6 bathroom 实现 + PR-7 General Noop
//   - BathroomFallRules: PR-10 §6.A 4 类规则
//   - BedroomFallRules: PR-11 §6.B 11b/11c
//
// 顺序：census 必须先注入（adjudicator/fall rules 都依赖），然后注册 rooms（rooms 注入时
// PR-7 检测 IsPublicBathroom 调 census.MarkPublicBathroom），最后接 adjudicator + fall rules。
func startRoomEngine(ctx context.Context, cfg *config.Config, db *sql.DB,
	rdb *redis.Client, logger *zap.Logger) (*roomengine.Engine, error) {

	engine := roomengine.NewEngine(rdb, logger)

	// S0.c-4b：DBN 顶层裁决路由（纯裁决，engine.OnRoomFrame seam）。registerAllRooms 填 geom/rooms，
	// 注册后按 suiteID 分组 engine.Unit 并接 engine.OnRoomFrame。
	router := newDBNRouter(logger)
	router.eng = engine

	// 1. 注入 yaml 运行时参数 + Persister
	engine.Configure(buildRuntimeConfig(cfg, db))

	// 1b. 注入 AI publish 单点配置（mode + source）
	engine.SetAIPublishConfig(cfg.AIPublish.Mode, cfg.AIPublish.Source)
	logger.Info("roomengine: ai_publish configured",
		zap.String("mode", cfg.AIPublish.Mode),
		zap.String("source", cfg.AIPublish.Source))

	// 1c. PR-2 SuiteCensusManager：单例 + Redis 持久化。先注入再注册 rooms
	//     （PR-7 RegisterRoom 内检测 IsPublicBathroom 调 census.MarkPublicBathroom 需要 census 在场）。
	census := roomengine.NewSuiteCensusManager(rdb, roomengine.DefaultSuiteCensusConfig(), logger)
	engine.SetSuiteCensus(census)

	// 2. 注册所有有 layout 的房间（PR-7：IsPublicBathroom=true 自动 MarkPublicBathroom）
	registered, err := registerAllRooms(ctx, engine, db, logger, router, cfg.DataBaseURL)
	if err != nil {
		return nil, fmt.Errorf("register rooms: %w", err)
	}
	logger.Info("roomengine: rooms registered", zap.Int("count", registered))

	// 2b. S0.c-4b：按 suiteID 分组 rooms → engine.Unit（跨房 hand-off neighbor 轴）；
	//     单房 unit 无兄弟 → ρ_xroom≡0 → 与单房 Room 逐帧等价。接 engine.OnRoomFrame = 纯裁决路由。
	unitRooms := map[string]map[string]*dbnengine.Room{}
	for roomID, unitKey := range router.roomUnit {
		if unitRooms[unitKey] == nil {
			unitRooms[unitKey] = map[string]*dbnengine.Room{}
		}
		unitRooms[unitKey][roomID] = router.rooms[roomID]
	}
	for unitKey, rooms := range unitRooms {
		router.units[unitKey] = dbnengine.NewUnit(rooms, 1)
	}
	engine.OnRoomFrame = router.onRoomFrame
	logger.Info("roomengine: DBN router wired", zap.Int("units", len(router.units)), zap.Int("rooms", len(router.rooms)))

	// 3. 建立 device_uid / device_addr → room_id 路由表
	mapped, err := mapDevicesToRooms(ctx, engine, db, logger)
	if err != nil {
		return nil, fmt.Errorf("map devices: %w", err)
	}
	logger.Info("roomengine: devices mapped", zap.Int("count", mapped))

	// 3b. PR-15：每日 22:00 (local) 重读 layout，hash 变即重置该 room → 从 0 重学
	engine.SetDailyLayoutReload(22, db)

	// 3c. config:card:stream 事件驱动 reload —— 见 main.go configCardConsumer wire。
	// (旧 60s SetRoutesReloader 已废弃；事件驱动延迟 < 1s，10 次/天可忽略)

	// 3d. 旧 gate-list ghost 裁决（GhostAdjudicator）+ Bathroom/BedroomFallRules 已随 DBN cutover 删除：
	// DBN（belief/engine，经 OnRoomFrame seam）自有 ghost 检测（motion/mirror）+ fall 裁决，单一 fire 权威。
	logger.Info("roomengine: gate-list ghost/fall retired; DBN(belief/engine) sole fire authority via OnRoomFrame")

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
			SitScoreSec:  float64(r.Decay.SitScoreSec),
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
			SitPromoteTau:     float64(r.Learn.SitPromoteTauX10) / 10,
			SitSpreadCm:       r.Learn.SitSpreadCm,
		},
		DecayInterval:      time.Duration(r.Belief.DecayIntervalSec) * time.Second,
		BeliefScanInterval: time.Duration(r.Belief.ScanIntervalSec) * time.Second,
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

// engineReloader 把 registerAllRooms + mapDevicesToRooms 包装成 consumer.ConfigChangedReloader。
// 用于 config_card_consumer 收到 config.changed 时触发 sensor 端 incremental reload。
type engineReloader struct {
	ctx         context.Context
	engine      *roomengine.Engine
	db          *sql.DB
	logger      *zap.Logger
	dataBaseURL string
}

func (r *engineReloader) ReloadRooms(ctx context.Context) error {
	// reload 只刷 layout/cell 重学；DBN router geom 不在 reload 重建（router=nil 跳过），由进程重启重建。
	_, err := registerAllRooms(ctx, r.engine, r.db, r.logger, nil, r.dataBaseURL)
	return err
}

func (r *engineReloader) ReloadDevices(ctx context.Context) error {
	_, err := mapDevicesToRooms(ctx, r.engine, r.db, r.logger)
	return err
}

// registerAllRooms 扫 rooms 表所有 room，逐个 ParseLayoutConfig + Optimize + RegisterRoom；
// layout 为空或解析失败的 room 仍以"无 layout placeholder"注册，让 sleepad-only 流量
// （ProcessSleepadBedEvent 不依赖 grid）能正常路由到 TrackManager，避免被 dropped_unrouted_message 沉默。
//
// v2 schema: rooms 表无 tenant_id/unit_id/layout_config 列；layout 在 room_visual_layout 表（PK=spatial_prefix）；
// tenant_id/unit_id 由 room_id INET prefix 派生（/48 / /80）；timezone 从 units LPM (unit /80 contains room /88) 取。
//
// sensor_v2 PR-Bootstrap：SQL 加 4 个 v2 RoomConfig 字段：
//   - room_type → RoomKind（决定 16 复用 PG 列；空串 = 默认 bedroom）
//   - is_public_bathroom → IsPublicBathroom（PR-7 决定 24 / migration 2026-05-17）
//   - SuiteID 计算（PR-7 SuiteID 取值约定）：
//     public bathroom (room_type=bathroom + is_public_bathroom=true) → 自身 /128
//     其他（含 suite bathroom） → unit /80
//   - ResidentID LPM 查询：resident_unit valid_to IS NULL + spatial_prefix 与 room_id 重叠
//     LPM order by masklen DESC 取最 specific 绑定（/96 bed > /88 room > /80 unit）
//     multi-resident 场景下返回任意一个（决定 19 留 PR-X，详 sensor_v2.md §10.4 注释）。
func registerAllRooms(ctx context.Context, engine *roomengine.Engine, db *sql.DB,
	logger *zap.Logger, router *dbnRouter, dataBaseURL string) (int, error) {

	// SQL 注：multi-resident bedroom 场景（同 /80 多 resident 各绑 /96 bed）下，
	// LIMIT 1 返回任意一个 — 决定 19 留 PR-X，需要 sleepad device_uid → resident.id 静态映射。
	// 单 resident 部署（v2 主目标）下结果确定。
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

	// 预取所有 /128 device layout canvas，按所属 /88 room 分组。
	// 根因修复：layout 全是 /128 device 级，旧 JOIN `rvl.spatial_prefix = r.room_id`(/88) 0 命中。
	canvasesByRoom, err := roomengine.LoadRoomCanvases(ctx, db, logger)
	if err != nil {
		return 0, err
	}

	// 每房雷达 device uid（hex）→ 取固件活体 declare_area 算床区 area_id（area_type∈{2,5}）。
	// 床 area_id 单源=固件活体（与上报 track FwAreaID 同源），非 canvas/baseline。
	radarUIDsByRoom := map[string][]string{}
	rdrows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text, dfm.device_uid
		FROM rooms r
		JOIN devices d                ON r.room_id >>= d.device_addr
		JOIN device_factory_meta dfm  ON dfm.device_uid = d.device_uid
		WHERE dfm.device_type::text = 'Radar'`)
	if err != nil {
		return 0, err
	}
	for rdrows.Next() {
		var rid, uid string
		if rdrows.Scan(&rid, &uid) == nil && uid != "" {
			radarUIDsByRoom[rid] = append(radarUIDsByRoom[rid], uid)
		}
	}
	rdrows.Close()
	dac := newDeclareAreaClient(dataBaseURL, logger)

	count := 0
	placeholderCount := 0
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
			placeholderCount++
		}
		// sensor_v2 v2 字段填入：RoomType / IsPublicBathroom / SuiteID / ResidentID
		// （RegisterRoom 内部检测 IsPublicBathroom + 调 MarkPublicBathroom — 见 PR-7）
		cfg.RoomType = roomType
		cfg.IsPublicBathroom = isPublicBathroom
		cfg.SuiteID = suiteID
		cfg.ResidentID = residentID

		// 床区 area_id 单源=固件活体 declare_area（type∈{2,5}）→ cfg.BedAreaIDs，让
		// TrackManager.fwIsBed（NewTrackManager 在 RegisterRoom 内建）与 DBN bedHitMask 同源。
		// per-radar 容错：某台拉不到只丢这台床区贡献，不跳整房（其余在线雷达+sleepad 照常监控）。
		if hasLayout {
			var fwBeds []int
			for _, uid := range radarUIDsByRoom[roomID] {
				ids, ferr := dac.bedAreaIDs(ctx, uid)
				if ferr != nil {
					logger.Error("declare_area 拉取失败 → 跳过该雷达床区（不跳整房；该台 N 不命中，房用其余在线雷达+sleepad 监控）",
						zap.String("room", roomID), zap.String("radar", uid), zap.Error(ferr))
					continue
				}
				fwBeds = append(fwBeds, ids...)
			}
			cfg.BedAreaIDs = fwBeds
			if len(fwBeds) == 0 {
				logger.Warn("declare_area: no firmware bed area for layout room (fwIsBed never hits → no bed-occupancy SFallen suppression)",
					zap.String("room", roomID))
			}
		}

		engine.RegisterRoom(cfg)
		engine.SetRoomTenant(roomID, tenantID)

		// S0.c-4b：建 DBN 路由 geom + engine.Room（懒建，单房 unit ρ=0）。
		// 进 DBN 判据 = 有信号源（雷达 layout OR canvas sleepad）。
		// 单雷达房 seed covers=1 = 零回归正确（StageA cd2b 适用）。
		// 🔴 StageB(多雷达) 前置阻塞（双雷达 FN 守卫本体，禁 silent，见 B·R11.1）：
		//   declare_area→BedAreaIDs 单源 + SetDeviceGeom per-device covers MM + BuildRoomMM(sleepad 吸纳)。
		//   缺这些时 09e7/D523 等双雷达 case 必重现历史 FN，严禁进 StageA。
		if router != nil {
			sleepadPresent := len(cfg.Sleepads) > 0
			if hasLayout || sleepadPresent {
				nb := len(cfg.Beds)
				if nb == 0 && sleepadPresent {
					nb = 1
				}
				beds := rectsFrom(cfg.Beds)
				for len(beds) < nb {
					beds = append(beds, adapter.Rect{})
				}
				bedAreaIDs := make([]int, nb)
				if hasLayout {
					for j := 0; j < nb && j < len(cfg.BedAreaIDs); j++ {
						bedAreaIDs[j] = cfg.BedAreaIDs[j]
					}
				} else {
					for j := 0; j < nb; j++ {
						bedAreaIDs[j] = j + 1
					}
				}
				router.geom[roomID] = &roomGeom{
					beds:           beds,
					bedAreaIDs:     bedAreaIDs,
					walls:          wallsFromPolygon(cfg.WallPolygon),
					entrances:      rectsFrom(cfg.Enters),
					radarPos:       adapter.Point{X: cfg.Radar.Center.X, Y: cfg.Radar.Center.Y},
					sleepadPresent: sleepadPresent,
					radarLess:      !hasLayout,
					nb:             nb,
				}
				seed := adapter.FrameInput{Beds: beds, Covers: ones(nb), Onbed: ones(nb), Overlap: ones(nb)}
				router.rooms[roomID] = dbnengine.NewRoom(adapter.BedGeoms(seed), nb)
				router.roomType[roomID] = cfg.RoomType
				if tz, terr := time.LoadLocation(cfg.Timezone); terr == nil {
					router.roomTZ[roomID] = tz
				}
				router.roomUnit[roomID] = cfg.SuiteID
			}
		}
		count++
	}
	if placeholderCount > 0 {
		logger.Info("roomengine: registered rooms without layout (sleepad-only fallback)",
			zap.Int("placeholder_count", placeholderCount))
	}
	return count, rows.Err()
}

// PR-Bootstrap: loadStayAlarmEnablement 已删除（v1 dead code）。
// 原用途："运维显式启用 Stay alarm 的房间按 bathroom 处理" — PR-10 BathroomStillFall 用
// room.kind=="bathroom" 分支替代。zonealarm.Supervisor 的 Stay rule（state-anchored，
// 消费 RoomState.AloneSinceTs，day 45min / night 30min）由 DefaultRules + yaml 加载，
// 与本函数无关。

// mapDevicesToRooms 建 device_addr → room 路由表。
//
// v2 schema: 旧 devices.bound_room_id / bound_bed_id 列已退役；device 与 room 关系
// 由 device_ipv6 INET prefix 派生（room_id /88 contains device_ipv6 /128）。
// device_factory_meta (dfm) 提供 device_type；旧 device_store 表已退役。
//
// （doc/device_ipv6_migration_checklist.md Phase D）后 engine 内部
// map key 为 canonical IPv6 字符串（addr.String()），与 envelope DeviceAddr 一致。
func mapDevicesToRooms(ctx context.Context, engine *roomengine.Engine, db *sql.DB,
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
		// 单程票：addr 是唯一 device key（替代旧的 deviceAddr UUID + deviceUID MAC 双键）
		engine.MapDeviceToRoom(deviceAddr, roomID)
		// 注册 device_addr → device_type，AI publish 用作 envelope.DeviceType
		if deviceType != "" {
			engine.MapDeviceAddrToType(deviceAddr, deviceType)
		}
		// MapDeviceAddrToUID 单程票后语义重定向：addr (IPv6) → device_uid hex MAC
		// 用于 log 双格式 + 人眼可读（zap field "device_uid_hex"）
		if deviceUIDHex != "" {
			engine.MapDeviceAddrToUID(deviceAddr, deviceUIDHex)
		}
		count++
	}
	return count, rows.Err()
}
