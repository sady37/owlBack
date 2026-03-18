package card

const (
	hashPrefix     = "card:state:"
	realtimeStream = "card:realtime:stream"
	statusStream   = "card:status:stream"
)

func HashKey(cardID string) string {
	return hashPrefix + cardID
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
