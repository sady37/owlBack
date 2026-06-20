package roomengine

// layout_load.go — 从 DB 加载 /128 device layout canvas 并按 /88 room 合并成 RoomConfig。
//
// 根因修复：room_visual_layout.spatial_prefix 全是 /128 device 级；旧 bootstrap JOIN
// `rvl.spatial_prefix = r.room_id`(/88) 等值匹配 0 命中 → layout 从不加载，几何门失效。
// 此处按 /128 device 加载，按 /88 room 聚合（单雷达房 = 唯一 canvas；多雷达房 best-effort
// union + 收集每台 mount，device-local 帧未配准 = 治本⑧ seam）。
// registerAllRooms + runDailyLayoutReload 共用，避免加载逻辑双写漂移。

import (
	"context"
	"database/sql"

	"owl-common/radarutils"

	"go.uber.org/zap"
)

// DeviceCanvas 一台 device 的 layout canvas（deviceAddr /128 + canvas JSON 文本）。
type DeviceCanvas struct {
	DeviceAddr string
	Canvas     string
}

// LoadRoomCanvases 一次扫全部 /128 device layout canvas，按所属 /88 room 分组。
// key = room_id /88 文本。
func LoadRoomCanvases(ctx context.Context, db *sql.DB, logger *zap.Logger) (map[string][]DeviceCanvas, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT network(set_masklen(spatial_prefix, 88))::text AS room_id,
		       host(spatial_prefix)::text                     AS device_addr,
		       canvas::text                                   AS canvas
		FROM room_visual_layout
		WHERE masklen(spatial_prefix) = 128
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]DeviceCanvas)
	for rows.Next() {
		var roomID, deviceAddr, canvas string
		if err := rows.Scan(&roomID, &deviceAddr, &canvas); err != nil {
			logger.Warn("scan room_visual_layout row", zap.Error(err))
			continue
		}
		if canvas == "" {
			continue
		}
		out[roomID] = append(out[roomID], DeviceCanvas{DeviceAddr: deviceAddr, Canvas: canvas})
	}
	return out, rows.Err()
}

// BuildRoomConfigFromCanvases 把一个 /88 room 下所有 /128 device canvas 解析+合并成单个 RoomConfig。
// 单雷达房（绝大多数，含 CABB）= 唯一 canvas，cfg 即该 canvas 几何，完全正确。
// 多雷达房 = 各 canvas 独立 device-local 帧、无共享房间坐标系：best-effort union 各自几何 +
// 收集每台 mount/deviceAddr（cfg.Radars/RadarAddrs，供 per-device 坐标转换）。
// hasLayout=false 表示无任何可解析 canvas（caller 退化 placeholder）。
func BuildRoomConfigFromCanvases(roomID string, canvases []DeviceCanvas, logger *zap.Logger) (RoomConfig, bool) {
	var merged RoomConfig
	merged.RoomID = roomID
	merged.DeviceBeds = make(map[string][]radarutils.Rect)
	merged.DeviceBedHeights = make(map[string][]int)
	parsed := 0
	for _, dc := range canvases {
		cfg, err := ParseLayoutConfig(dc.DeviceAddr, []byte(dc.Canvas))
		if err != nil {
			logger.Warn("parse device layout failed; skipping device",
				zap.String("room_id", roomID), zap.String("device_addr", dc.DeviceAddr), zap.Error(err))
			continue
		}
		parsed++
		// per-device 床：留住每台自己 canvas 的床（covers 只用本台 layout，不跨系借别台的床）
		merged.DeviceBeds[dc.DeviceAddr] = cfg.Beds
		merged.DeviceBedHeights[dc.DeviceAddr] = cfg.BedHeights
		merged.Radars = append(merged.Radars, cfg.Radar)
		merged.RadarAddrs = append(merged.RadarAddrs, dc.DeviceAddr)
		if len(merged.WallPolygon) == 0 {
			merged.WallPolygon = cfg.WallPolygon
		}
		merged.Enters = append(merged.Enters, cfg.Enters...)
		merged.EnterTargets = append(merged.EnterTargets, cfg.EnterTargets...)
		merged.EnterHeights = append(merged.EnterHeights, cfg.EnterHeights...)
		merged.Beds = append(merged.Beds, cfg.Beds...)
		merged.BedHeights = append(merged.BedHeights, cfg.BedHeights...)
		merged.BedAreaIDs = append(merged.BedAreaIDs, cfg.BedAreaIDs...)
		merged.Toilets = append(merged.Toilets, cfg.Toilets...)
		merged.ToiletHeights = append(merged.ToiletHeights, cfg.ToiletHeights...)
		merged.Showers = append(merged.Showers, cfg.Showers...)
		merged.ShowerHeights = append(merged.ShowerHeights, cfg.ShowerHeights...)
		merged.Chairs = append(merged.Chairs, cfg.Chairs...)
		merged.ChairHeights = append(merged.ChairHeights, cfg.ChairHeights...)
		merged.Furnitures = append(merged.Furnitures, cfg.Furnitures...)
		merged.FurnitureHeights = append(merged.FurnitureHeights, cfg.FurnitureHeights...)
		merged.Interferes = append(merged.Interferes, cfg.Interferes...)
		merged.InterfereHeights = append(merged.InterfereHeights, cfg.InterfereHeights...)
		merged.Sleepads = append(merged.Sleepads, cfg.Sleepads...)
	}
	if parsed == 0 {
		return RoomConfig{}, false
	}
	merged.Radar = merged.Radars[0] // 主雷达（boundaryPolygon 兜底 + L1 mirror tiebreaker）
	if parsed > 1 {
		logger.Warn("multi-radar room: layout geometry best-effort (device-local frames not co-registered; mount-pose alignment pending)",
			zap.String("room_id", roomID), zap.Int("radar_count", parsed))
	}
	return merged, true
}
