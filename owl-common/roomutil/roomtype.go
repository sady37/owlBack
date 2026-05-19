// Package roomutil 房间分类相关共享工具。
//
// 提供 room_name → room_type 的统一判定。两个使用方：
//   - wisefido-cardagg：决定设备是否走卫生间事件 / Stay 检测路径
//   - wisefido-sensor：still fall 触发位置语义并集（cell.AreaToilet/AreaShower ∪ room.bathroom）
//
// 单源避免漂移：将来加新模式（如 "WC1F"、"powderRoom"）只需改本文件，两服务都生效。
package roomutil

import (
	"strings"

	"owl-common/card"
)

// ClassifyRoomType 根据 room_name 推断 room_type（不区分大小写，子串匹配）。
// 返回 owl-common/card.RoomType*（0=Default / 1=Bathroom / 2=Kitchen）。
//
//	wc / bathroom / restroom / toilet → Bathroom
//	kitchen                           → Kitchen
//	其余（含 bedroom / living / dining） → Default
//
// 设计意图：用户在 layout / unit 编辑器输入的 room_name 是英文/中英混合自由文本，
// 用子串匹配比硬编码 enum 灵活；空名 → Default。
func ClassifyRoomType(roomName string) int {
	lower := strings.ToLower(roomName)
	switch {
	case strings.Contains(lower, "wc"),
		strings.Contains(lower, "bathroom"),
		strings.Contains(lower, "restroom"),
		strings.Contains(lower, "toilet"):
		return card.RoomTypeBathroom
	case strings.Contains(lower, "kitchen"):
		return card.RoomTypeKitchen
	default:
		return card.RoomTypeDefault
	}
}

// IsBathroom 便捷判定（少打字）。
func IsBathroom(roomName string) bool {
	return ClassifyRoomType(roomName) == card.RoomTypeBathroom
}
