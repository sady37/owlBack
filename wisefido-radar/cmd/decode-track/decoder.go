package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// TrackData 解码后的 track 数据
type TrackData struct {
	TargetID      int    // 字节 0: 目标 ID (0-7 或 88 表示无人)
	PositionX     int    // 字节 1: x 坐标（分米，已转换为厘米）
	PositionY     int    // 字节 2: y 坐标（分米，已转换为厘米）
	PositionZ     int    // 字节 3: z 坐标（厘米）
	RemainingTime int    // 字节 12: 剩余时间（秒，0-60）
	Pose          int    // 字节 13: 姿态值 (0-11)
	Event         int    // 字节 14: 事件值 (0-4)
	AreaID        int    // 字节 15: 区域 ID
}

// DecodeMonitorTrack 解码 monitor track 字段
// base64Track: base64 编码的 track 字符串
// 返回: 解码后的 track 数据数组（每个人一个元素）
func DecodeMonitorTrack(base64Track string) ([]TrackData, error) {
	// 1. Base64 解码
	trackBytes, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 track: %w", err)
	}

	// 2. 检查长度是否为 16 的倍数
	if len(trackBytes)%16 != 0 {
		return nil, fmt.Errorf("invalid track length: %d (must be multiple of 16)", len(trackBytes))
	}

	// 3. 按 16 字节分段处理（每个人）
	personCount := len(trackBytes) / 16
	results := make([]TrackData, personCount)

	for i := 0; i < personCount; i++ {
		offset := i * 16
		personData := trackBytes[offset : offset+16]

		// 字节 0: target_id
		targetID := int(personData[0])

		// 字节 1: position_x (有符号数，分米，转换为厘米)
		positionX := int(int8(personData[1])) // 转换为有符号数
		positionXCm := positionX * 10         // 分米 → 厘米

		// 字节 2: position_y (有符号数，分米，转换为厘米)
		positionY := int(int8(personData[2])) // 转换为有符号数
		positionYCm := positionY * 10         // 分米 → 厘米

		// 字节 3: position_z (厘米)
		positionZ := int(personData[3])

		// 字节 12: remaining_time (秒)
		remainingTime := int(personData[12])

		// 字节 13: pose (姿态值)
		pose := int(personData[13])

		// 字节 14: event (事件值)
		event := int(personData[14])

		// 字节 15: area_id
		areaID := int(personData[15])

		results[i] = TrackData{
			TargetID:      targetID,
			PositionX:     positionXCm,
			PositionY:     positionYCm,
			PositionZ:     positionZ,
			RemainingTime: remainingTime,
			Pose:          pose,
			Event:         event,
			AreaID:        areaID,
		}
	}

	return results, nil
}

// GetPoseDisplay 获取姿态显示名称
func GetPoseDisplay(pose int) string {
	poseMap := map[int]string{
		0:  "初始化",
		1:  "行走",
		2:  "疑似跌倒",
		3:  "蹲坐",
		4:  "站立",
		5:  "跌倒确认",
		6:  "卧床",
		7:  "疑似坐地",
		8:  "确认坐地",
		9:  "普通床上坐起",
		10: "疑似床上坐起",
		11: "确认床上坐起",
	}
	if display, ok := poseMap[pose]; ok {
		return display
	}
	return fmt.Sprintf("未知(%d)", pose)
}

// GetEventDisplay 获取事件显示名称
func GetEventDisplay(event int) string {
	eventMap := map[int]string{
		0: "无事件",
		1: "进入房间",
		2: "离开房间",
		3: "进入区域",
		4: "离开区域",
	}
	if display, ok := eventMap[event]; ok {
		return display
	}
	return fmt.Sprintf("未知(%d)", event)
}

// StatTrackData 解码后的 stat track 数据
type StatTrackData struct {
	Version            int // 字节 0: 版本标识符 (1 或 2)
	PeopleCount        int // 字节 1: 人数 (0-8)
	WalkDistance       int // 字节 2-3: 行走距离（米，已转换为厘米）
	WalkDuration       int // 字节 4: 行走时长（秒）
	SitDuration        int // 字节 5: 静坐时长（未开放使用）
	LieDuration        int // 字节 6: 躺卧时长（秒）
	StandDuration      int // 字节 7: 站立时长（秒）
	MultiPersonDuration int // 字节 8: 多人时长（秒）
}

// DecodeStatTrack 解码 stat track 字段
// base64Track: base64 编码的 track 字符串
// 返回: 解码后的 stat track 数据
func DecodeStatTrack(base64Track string) (*StatTrackData, error) {
	// 1. Base64 解码
	trackBytes, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 track: %w", err)
	}

	// 2. 检查长度是否为 16 字节
	if len(trackBytes) != 16 {
		return nil, fmt.Errorf("invalid stat track length: %d (must be 16)", len(trackBytes))
	}

	// 3. 解析字段
	version := int(trackBytes[0])
	peopleCount := int(trackBytes[1])

	// 字节 2-3: 行走距离（Big-Endian，米 → 厘米）
	walkDistanceM := int(trackBytes[2])<<8 | int(trackBytes[3])
	walkDistanceCm := walkDistanceM * 100 // 米 → 厘米

	walkDuration := int(trackBytes[4])
	sitDuration := int(trackBytes[5])
	lieDuration := int(trackBytes[6])
	standDuration := int(trackBytes[7])
	multiPersonDuration := int(trackBytes[8])

	return &StatTrackData{
		Version:             version,
		PeopleCount:         peopleCount,
		WalkDistance:        walkDistanceCm,
		WalkDuration:        walkDuration,
		SitDuration:         sitDuration,
		LieDuration:         lieDuration,
		StandDuration:       standDuration,
		MultiPersonDuration: multiPersonDuration,
	}, nil
}

// FormatTrackBytes 格式化 track 字节数组为十六进制字符串
func FormatTrackBytes(base64Track string) (string, error) {
	trackBytes, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 track: %w", err)
	}
	return hex.EncodeToString(trackBytes), nil
}
