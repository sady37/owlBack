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

// DeviceStatusHashKey 单设备运行时状态独立 Hash key（device:status:{ipv6}）。
// 入参 deviceAddr = /128 IPv6 canonical host text（device_ipv6 单程票）。
// Hash 字段：online / signal_poor / angle_abnormal / sensor_detached / last_seen_ms 等。
func DeviceStatusHashKey(deviceAddr string) string {
	return deviceStatusPrefix + deviceAddr
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
