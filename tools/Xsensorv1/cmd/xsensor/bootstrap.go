package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"
	"time"

	"owlBack/tools/Xsensorv1/internal/config"
	"owlBack/tools/Xsensorv1/internal/roomengine"
	"owlBack/tools/Xsensorv1/internal/roomengine/adapter"
	"owlBack/tools/Xsensorv1/internal/roomengine/belief"
	"owlBack/tools/Xsensorv1/internal/roomengine/engine"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// normAddr 归一 IPv6 字符串（canvas RadarAddrs vs DB host() 两形式对齐）。解析失败→原样。
func normAddr(s string) string {
	if a, err := netip.ParseAddr(s); err == nil {
		return a.String()
	}
	return s
}

// uidLast4 device_uid hex 末 4 字符 = per-device 床耦合键（同 makeLogicID 前缀）。
func uidLast4(uid string) string {
	if len(uid) > 4 {
		return uid[len(uid)-4:]
	}
	return uid
}

// deviceBedGeom 某雷达设备的 per-bed 几何：covers=1 当房床 j 是本设备**自己画的床**（rect 归属），否则 0。
//
//	Onbed/Overlap=1 与 seed 一致（零回归：单雷达拥有全部床→covers 全 1=改造前）。
func deviceBedGeom(roomBeds, ownBeds []adapter.Rect) []belief.BedGeom {
	g := make([]belief.BedGeom, len(roomBeds))
	for j, rb := range roomBeds {
		covers := 0.0
		for _, ob := range ownBeds {
			if ob == rb {
				covers = 1
				break
			}
		}
		g[j] = belief.BedGeom{Covers: covers, Onbed: 1, Overlap: 1}
	}
	return g
}

// startRoomEngine 建 + 配 + 注册房间 + 接 DBN 顶层回调 + 启动消费循环。
// 适配版（vs wisefido-sensor engine_bootstrap）：去掉已剥离的 AI publish / 每日 reload /
// ghost adjudicator / Persister（馈送层无持久化无投影，决策归 DBN）。
func startRoomEngine(ctx context.Context, cfg *config.Config, db *sql.DB, rdb *redis.Client,
	router *dbnRouter, logger *zap.Logger) (*roomengine.Engine, error) {

	eng := roomengine.NewEngine(rdb, logger)
	eng.Configure(buildRuntimeConfig(cfg))

	census := roomengine.NewSuiteCensusManager(rdb, roomengine.DefaultSuiteCensusConfig(), logger)
	eng.SetSuiteCensus(census)

	dac := newDeclareAreaClient(cfg.DataBaseURL, logger)
	registered, err := registerAllRooms(ctx, eng, db, router, dac, logger)
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
		// residentCount=1（单住户测试；多住户后续）。hand-off 核不随公共度（见 neighbor.go）。
		router.units[unitKey] = engine.NewUnit(rooms, 1)
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
	router *dbnRouter, dac *declareAreaClient, logger *zap.Logger) (int, error) {

	rows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text                                       AS room_id,
		       r.room_name,
		       COALESCE(r.room_type, 0)                              AS room_type,
		       COALESCE(r.is_public_bathroom, FALSE)                 AS is_public_bathroom,
		       CASE
		         WHEN r.room_type = 1 AND COALESCE(r.is_public_bathroom, FALSE)
		           THEN r.room_id::text
		         ELSE host(network(set_masklen(r.room_id, 80)))::text || '/80'
		       END                                                   AS suite_id,
		       COALESCE(u.timezone, '')                              AS timezone,
		       host(network(set_masklen(r.room_id, 48)))::text || '/48' AS tenant_pref,
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

	// 有 sleepad 设备的房（bed_state 不依赖 layout；sleepad-only 房=有sleepad无雷达layout 也要进 DBN）。
	sleepadRooms := map[string]bool{}
	srows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text
		FROM rooms r
		JOIN devices d ON r.room_id >>= d.device_addr
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE dfm.device_type::text = 'Sleepad'
		GROUP BY r.room_id`)
	if err != nil {
		return 0, err
	}
	for srows.Next() {
		var rid string
		if srows.Scan(&rid) == nil {
			sleepadRooms[rid] = true
		}
	}
	srows.Close()

	// sleepad 设备 addr（/128，per room）→ 静态 MM ObjSleepad（onbed=sleepad/128∈bed/96 前缀确定绑定）。
	sleepadAddrsByRoom := map[string][]netip.Addr{}
	sarows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text, host(d.device_addr)::text
		FROM rooms r
		JOIN devices d                ON r.room_id >>= d.device_addr
		JOIN device_factory_meta dfm  ON dfm.device_uid = d.device_uid
		WHERE dfm.device_type::text = 'Sleepad'`)
	if err != nil {
		return 0, err
	}
	for sarows.Next() {
		var rid, addr string
		if sarows.Scan(&rid, &addr) == nil {
			if a, perr := netip.ParseAddr(addr); perr == nil {
				sleepadAddrsByRoom[rid] = append(sleepadAddrsByRoom[rid], a)
			}
		}
	}
	sarows.Close()

	// DB beds /96 prefix（静态 MM onbed/covers 用；mirror SpatialCache bed 装载，唯一 DB 只读归口）。
	bedsByRoom, err := roomengine.LoadRoomBeds(ctx, db, logger)
	if err != nil {
		return 0, err
	}

	// 每房雷达 device uid（hex）→ 取固件活体 declare_area 算床区 area_id；并存 addr→uid（per-device covers 注入用）。
	radarUIDsByRoom := map[string][]string{}
	radarUIDByAddr := map[string]map[string]string{} // roomID → 归一 addr → device_uid（MM per-device geom keying）
	rdrows, err := db.QueryContext(ctx, `
		SELECT r.room_id::text, host(d.device_addr)::text, dfm.device_uid
		FROM rooms r
		JOIN devices d                ON r.room_id >>= d.device_addr
		JOIN device_factory_meta dfm  ON dfm.device_uid = d.device_uid
		WHERE dfm.device_type::text = 'Radar'`)
	if err != nil {
		return 0, err
	}
	for rdrows.Next() {
		var rid, addr, uid string
		if rdrows.Scan(&rid, &addr, &uid) == nil && uid != "" {
			radarUIDsByRoom[rid] = append(radarUIDsByRoom[rid], uid)
			if radarUIDByAddr[rid] == nil {
				radarUIDByAddr[rid] = map[string]string{}
			}
			radarUIDByAddr[rid][normAddr(addr)] = uid
		}
	}
	rdrows.Close()

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

		// 床区 area_id 单源 = 固件活体 declare_area（type∈{2,5}）。治 canvas 下发区域 vs 固件漂移：
		// area_id 是固件槽位号，与上报 track FwAreaID 同源。覆盖 canvas 产出的 cfg.BedAreaIDs，
		// 让 TrackManager.fwIsBed（membership）与 DBN bedHitMask（per-bed）同源。
		var fwBeds []int
		if hasLayout {
			// per-radar 容错（治"一台离线全房瞎"）：某台 declare-area 拉不到 → **只丢这台的床区贡献**，
			//   **不跳整房**——同房其余在线雷达 + sleepad 照常注册监控（多雷达房=一台坏不该拖垮在线的那台）。
			//   ⚠️ 失败台无床区 → 它的 N(MN) 不命中 → 它单独不抬 SBed；房整体仍进 DBN 跑 fall。终态须 retry/canvas-fallback。
			for _, uid := range radarUIDsByRoom[roomID] {
				ids, ferr := dac.bedAreaIDs(ctx, uid)
				if ferr != nil {
					logger.Error("declare_area 拉取失败 → 跳过该雷达床区（**不跳整房**；该台 N 不命中，房用其余在线雷达+sleepad 监控；终态须 retry/canvas-fallback）",
						zap.String("room", roomID), zap.String("radar", uid), zap.Error(ferr))
					continue // 仅跳过本台，继续收其余雷达床区
				}
				fwBeds = append(fwBeds, ids...)
			}
			cfg.BedAreaIDs = fwBeds
			if len(fwBeds) == 0 {
				logger.Warn("declare_area: no firmware bed area for layout room (N never hits → no SBed boost)",
					zap.String("room", roomID))
			}
		}

		eng.RegisterRoom(cfg)
		eng.SetRoomTenant(roomID, tenantID)
		// 进 DBN 判据 = 有任一信号源(雷达 layout OR sleepad)。bed_state 不依赖 layout：
		//   sleepad-only 房(GuestRoom 等)无雷达几何但有 B 轴，照样进 DBN(radarLess)。
		hasSleepad := sleepadRooms[roomID]
		if hasLayout || hasSleepad {
			nb := len(cfg.Beds)
			if nb == 0 && hasSleepad {
				nb = 1 // sleepad-only 房：一张床(sleepad 对应)；几何无关(无雷达 covers)
			}
			beds := rectsFrom(cfg.Beds)
			for len(beds) < nb {
				beds = append(beds, adapter.Rect{}) // 补 nominal 床(radar-less 无几何意义)
			}
			// FrameInput.BedAreaIDs：per-bed 长度 nb，槽位填固件床区 area_id（与 beds 同序，bedHitMask 用）；
			//   sleepad-only 房无雷达声明区 → 虚拟非零 id 让合成 bed-track 命中 N。固件源同 cfg.BedAreaIDs(fwBeds)。
			bedAreaIDs := make([]int, nb)
			if hasLayout {
				for j := 0; j < nb && j < len(fwBeds); j++ {
					bedAreaIDs[j] = fwBeds[j]
				}
				if len(fwBeds) > nb {
					logger.Warn("declare_area: firmware bed areas exceed canvas beds (truncated)",
						zap.String("room", roomID), zap.Int("fw_beds", len(fwBeds)), zap.Int("nb", nb))
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
				sleepadPresent: hasSleepad, // 用 DB 设备(非 cfg.Sleepads 画layout)→ 修 radar 房 sleepad 漏判
				radarLess:      !hasLayout,
				nb:             nb,
			}
			seed := adapter.FrameInput{Beds: beds, Covers: ones(nb), Onbed: ones(nb), Overlap: ones(nb)}
			room := engine.NewRoom(adapter.BedGeoms(seed), nb)
			// MM per-(设备×床) covers 注入：每雷达设备只盖它**自己 canvas 画的床**（rect 归属）→ covers=1，
			//   别台没画的床=0（D523 没画床→全 0→跟床解耦）。defGeom(covers=1) 兜未注册/sleepad-only 合成 track。
			//   key=uid_last4（与 track LogicID[:4] 同源 device_uid → 天然匹配）。Onbed/Overlap=1 同 seed（零回归）。
			if hasLayout {
				for _, addr := range cfg.RadarAddrs {
					uid := radarUIDByAddr[roomID][normAddr(addr)]
					if uid == "" {
						logger.Warn("MM covers: radar addr 无 device_uid（跳过 per-device geom，回退 covers=1）",
							zap.String("room", roomID), zap.String("addr", addr))
						continue
					}
					room.SetDeviceGeom(uidLast4(uid), deviceBedGeom(beds, rectsFrom(cfg.DeviceBeds[addr])))
				}
			}
			router.rooms[roomID] = room
			router.roomType[roomID] = cfg.RoomType // UD timer per-room deadline（Bathroom=1→D=20min，余 UD=90min）
			if tz, terr := time.LoadLocation(cfg.Timezone); terr == nil {
				router.roomTZ[roomID] = tz // risktime(IsNightTime)算 floor 阈用；解析失败→nil→IsNightTime 退 UTC
			}
			// unitKey = suiteID（SQL 已 network() zero 主机位 → 同 /80 房共享；public bathroom=自身/128 独立）。
			router.roomUnit[roomID] = cfg.SuiteID
			// 静态 MM（samebed prior 权威，吸纳读）：room/88 + DB beds/96 + radar/sleepad /128，covers=RadarBedReachCount。
			//   在 RegisterRoom 之后建（RadarBedReachCount 需 deviceBeds 已注入）。prefix 空间算 samebed→化掉床 index（§8.2）。
			if roomPfx, perr := netip.ParsePrefix(roomID); perr == nil {
				var radarAddrs []netip.Addr
				for a := range radarUIDByAddr[roomID] {
					if pa, e := netip.ParseAddr(a); e == nil {
						radarAddrs = append(radarAddrs, pa)
					}
				}
				if mm := roomengine.BuildRoomMM(roomPfx, bedsByRoom[roomID], radarAddrs,
					sleepadAddrsByRoom[roomID], eng.RadarBedReachCount); mm != nil {
					router.mm[roomID] = mm
				}
			}
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
