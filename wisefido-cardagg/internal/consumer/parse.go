package consumer

import (
	"owl-common/redis"
)

// ParseMessage 将 Redis 一条消息的 Values（map）解析为 redis.IoTStreamMessage。
func ParseMessage(raw map[string]interface{}) (*redis.IoTStreamMessage, error) {
	return redis.FromStreamMap(raw)
}
