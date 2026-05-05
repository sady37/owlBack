package card

const (
	hashPrefix         = "card:state:"
	deviceStatusPrefix = "device:status:"
	realtimeStream     = "card:realtime:stream"
	statusStream       = "card:status:stream"
)

func HashKey(cardID string) string {
	return hashPrefix + cardID
}

// DeviceStatusHashKey 单设备运行时状态独立 Hash（Phase A）：
//   - 与 card:state:{card_id} 解耦：device-level 真相不再附属任何 card
//   - 一个设备共享给多张卡（业务模型当前禁止，但保留单独 key 避免冗余）
//   - 字段：online / signal_poor / angle_abnormal / sensor_detached / last_seen_ms 等
func DeviceStatusHashKey(deviceID string) string {
	return deviceStatusPrefix + deviceID
}

func RealtimeStreamName() string {
	return realtimeStream
}

func StatusStreamName() string {
	return statusStream
}

// TsKey returns the internal timestamp key for TTL-based expiry check.
func TsKey(fieldKey string) string {
	return fieldKey + "_ts"
}

// Stream 消息 type 字段：与 wisefido-cardagg 写入、wisefido-data 订阅一致
const (
	MsgTypeKey     = "type" // stream Values 中 message_type 的 key
	MsgTypeEvent   = "event"
	MsgTypeMonitor = "monitor"
)
