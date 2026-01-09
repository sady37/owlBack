package mqtt

import (
	"fmt"
	"strings"
)

// TopicType MQTT 主题类型
type TopicType string

const (
	// TopicTypeProp 属性主题（prop）
	TopicTypeProp TopicType = "prop"
	// TopicTypeMonitor 实时数据主题（monitor）
	TopicTypeMonitor TopicType = "monitor"
	// TopicTypeFunc 功能调用主题（func）
	TopicTypeFunc TopicType = "func"
	// TopicTypeStat 统计数据主题（stat）
	TopicTypeStat TopicType = "stat"
	// TopicTypeEvent 事件主题（event）
	TopicTypeEvent TopicType = "event"
	// TopicTypeAlarm 告警主题（alarm）
	TopicTypeAlarm TopicType = "alarm"
)

// Direction 主题方向
type Direction string

const (
	// DirectionGet 服务器发布，设备接收（命令）
	DirectionGet Direction = "get"
	// DirectionPost 设备发布，服务器接收（响应/数据）
	DirectionPost Direction = "post"
)

// TopicInfo 主题信息
type TopicInfo struct {
	Prefix    string     // 主题前缀（可为空）
	Type      TopicType  // 主题类型（prop, monitor, func, stat, event, alarm）
	ProductID string     // 产品 ID (0-255)
	UID       string     // 设备 UID
	Direction Direction  // 方向（get, post）
}

// BuildTopic 构建 MQTT 主题
// 根据协议文档 3.3.2 和 3.3.3 节
// 格式：/prefix/{type}/productId/UID/{direction}
// 如果 prefix 为空，格式：/{type}/productId/UID/{direction}
func BuildTopic(info *TopicInfo) string {
	if info.Prefix == "" {
		// 无前缀格式：/prop/88/414D7418B267/get
		return fmt.Sprintf("/%s/%s/%s/%s", info.Type, info.ProductID, info.UID, info.Direction)
	}
	// 有前缀格式：/prefix/prop/88/414D7418B267/get
	return fmt.Sprintf("/%s/%s/%s/%s/%s", info.Prefix, info.Type, info.ProductID, info.UID, info.Direction)
}

// ParseTopic 解析 MQTT 主题
// 从主题字符串中提取 prefix, type, productId, UID, direction
func ParseTopic(topic string) (*TopicInfo, error) {
	// 移除开头的斜杠
	topic = strings.TrimPrefix(topic, "/")
	
	parts := strings.Split(topic, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid topic format: %s (expected at least 4 parts)", topic)
	}
	
	info := &TopicInfo{}
	
	// 判断是否有 prefix
	// 如果有 prefix，格式：prefix/type/productId/UID/direction (5 parts)
	// 如果没有 prefix，格式：type/productId/UID/direction (4 parts)
	if len(parts) == 5 {
		// 有 prefix
		info.Prefix = parts[0]
		info.Type = TopicType(parts[1])
		info.ProductID = parts[2]
		info.UID = parts[3]
		info.Direction = Direction(parts[4])
	} else if len(parts) == 4 {
		// 无 prefix
		info.Prefix = ""
		info.Type = TopicType(parts[0])
		info.ProductID = parts[1]
		info.UID = parts[2]
		info.Direction = Direction(parts[3])
	} else {
		return nil, fmt.Errorf("invalid topic format: %s (expected 4 or 5 parts, got %d)", topic, len(parts))
	}
	
	// 验证类型
	validTypes := map[TopicType]bool{
		TopicTypeProp:    true,
		TopicTypeMonitor: true,
		TopicTypeFunc:    true,
		TopicTypeStat:    true,
		TopicTypeEvent:   true,
		TopicTypeAlarm:   true,
	}
	if !validTypes[info.Type] {
		return nil, fmt.Errorf("invalid topic type: %s", info.Type)
	}
	
	// 验证方向
	validDirections := map[Direction]bool{
		DirectionGet:  true,
		DirectionPost: true,
	}
	if !validDirections[info.Direction] {
		return nil, fmt.Errorf("invalid topic direction: %s", info.Direction)
	}
	
	return info, nil
}

// BuildPropertyGetTopic 构建属性命令主题（服务器发布）
func BuildPropertyGetTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeProp,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionGet,
	})
}

// BuildPropertyPostTopic 构建属性响应主题（设备发布）
func BuildPropertyPostTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeProp,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionPost,
	})
}

// BuildMonitorGetTopic 构建实时数据订阅主题（服务器发布）
func BuildMonitorGetTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeMonitor,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionGet,
	})
}

// BuildMonitorPostTopic 构建实时数据主题（设备发布）
func BuildMonitorPostTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeMonitor,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionPost,
	})
}

// BuildFunctionGetTopic 构建功能调用主题（服务器发布）
func BuildFunctionGetTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeFunc,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionGet,
	})
}

// BuildFunctionPostTopic 构建功能响应主题（设备发布）
func BuildFunctionPostTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeFunc,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionPost,
	})
}

// BuildStatPostTopic 构建统计数据主题（设备发布）
func BuildStatPostTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeStat,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionPost,
	})
}

// BuildEventPostTopic 构建事件主题（设备发布）
func BuildEventPostTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeEvent,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionPost,
	})
}

// BuildAlarmPostTopic 构建告警主题（设备发布）
func BuildAlarmPostTopic(prefix, productID, uid string) string {
	return BuildTopic(&TopicInfo{
		Prefix:    prefix,
		Type:      TopicTypeAlarm,
		ProductID: productID,
		UID:       uid,
		Direction: DirectionPost,
	})
}

