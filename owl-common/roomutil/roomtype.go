// Package roomutil 房间分类相关共享工具。
//
// 提供 room_name → 语义类型的统一判定。两个使用方：
//   - wisefido-cardagg：决定设备是否走卫生间事件 / Stay 检测路径
//   - wisefido-ai：still fall 触发位置语义并集（cell.AreaToilet/AreaShower ∪ room.bathroom）
//
// 单一来源避免漂移：将来加新模式（如 "WC1F"、"powderRoom"）只需改本文件，两服务都生效。
package roomutil

import "strings"

// 房间语义类型常量
const (
	RoomTypeBathroom = "bathroom"
	RoomTypeBedroom  = "bedroom"
	RoomTypeKitchen  = "kitchen"
	RoomTypeOther    = "other"
)

// ClassifyRoomType 根据 room_name 推断房间语义类型（不区分大小写，子串匹配）。
//
//	wc / bathroom / restroom / toilet → "bathroom"
//	bedroom / bed room                → "bedroom"
//	kitchen                           → "kitchen"
//	其余                               → "other"
//
// 设计意图：用户在 layout / unit 编辑器输入的 room_name 是英文/中英混合自由文本，
// 用子串匹配比硬编码 enum 灵活；空名 → "other"。
func ClassifyRoomType(roomName string) string {
	lower := strings.ToLower(roomName)
	switch {
	case strings.Contains(lower, "wc"),
		strings.Contains(lower, "bathroom"),
		strings.Contains(lower, "restroom"),
		strings.Contains(lower, "toilet"):
		return RoomTypeBathroom
	case strings.Contains(lower, "bedroom"),
		strings.Contains(lower, "bed room"):
		return RoomTypeBedroom
	case strings.Contains(lower, "kitchen"):
		return RoomTypeKitchen
	default:
		return RoomTypeOther
	}
}

// IsBathroom 便捷判定（少打字）。
func IsBathroom(roomName string) bool {
	return ClassifyRoomType(roomName) == RoomTypeBathroom
}
