package roomengine

import (
	"owl-common/radarutils"
)

// ParseRadarTracks 把"data_value"段（雷达 monitor JSON）解析成 TrackFrame 列表。
// 调用者：engine.parseTrackFrames（流式消费）+ cmd/roomengine-playback（DB 回放）。
//
// 输入：
//
//	dv          可能是 []interface{}（多 track 一帧）或 map[string]interface{}（单 track）
//	deviceAddr    设备 ID（写入 TrackFrame.DeviceAddr）
//	mount       雷达安装参数（坐标转换用）
//	fallbackTs  本帧 timestamp 缺失时的兜底
//
// 行为：
//   - 过滤无效帧（track_id 越界 / area_id=255+全 0 / 全 0）
//   - 调 RadarToCanvas 把雷达本地坐标转画布坐标
func ParseRadarTracks(dv interface{}, deviceAddr string, mount radarutils.RadarMount, fallbackTs int64) []TrackFrame {
	if dv == nil {
		return nil
	}

	var items []map[string]interface{}
	switch v := dv.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, m)
			}
		}
	case map[string]interface{}:
		items = append(items, v)
	default:
		return nil
	}

	var frames []TrackFrame
	for _, item := range items {
		trackID := intFromAny(item["track_id"])
		// 部分链路用 tracking_id（DB 直存）
		if trackID == 0 {
			if v := intFromAny(item["tracking_id"]); v != 0 {
				trackID = v
			}
		}

		// 无效帧过滤
		if trackID < 0 || trackID > 8 {
			continue
		}

		localX := intFromAny(item["position_x"])
		localY := intFromAny(item["position_y"])
		localZ := intFromAny(item["position_z"])
		areaID := intFromAny(item["area_id"])

		if localX == 0 && localY == 0 && localZ == 0 && areaID == 255 {
			continue
		}
		if localX == 0 && localY == 0 && localZ == 0 {
			continue
		}

		canvas := radarutils.RadarToCanvas(
			radarutils.RadarPoint{H: localX, V: localY, Z: localZ},
			mount,
		)

		pose := intFromAny(item["pose"])
		frameTs := int64FromAny(item["timestamp"])
		if frameTs == 0 {
			frameTs = fallbackTs
		}

		frames = append(frames, TrackFrame{
			TrackID:         trackID,
			DeviceAddr:      deviceAddr,
			X:               canvas.X,
			Y:               canvas.Y,
			Z:               canvas.Z,
			RawH:            localX,
			RawV:            localY,
			RawZ:            localZ,
			Pose:            pose,
			AreaType:        areaID,
			Event:           intFromAny(item["event"]),
			TMs:             frameTs,
			TrackConfidence: intFromAny(item["track_confidence"]),
			RemainingTime:   intFromAny(item["remaining_time"]),
		})
	}
	return frames
}
