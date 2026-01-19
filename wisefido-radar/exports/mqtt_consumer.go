package exports

// MQTTConsumer MQTT消费者接口（用于导出）
type MQTTConsumer struct{}

// getOutputStreamNameForTopicTypeStr 根据 topic_type 字符串获取输出 Redis Stream 名称
// 用于处理经过报警判断后的最终 topic_type（可能是 "alarm"）
func (c *MQTTConsumer) getOutputStreamNameForTopicTypeStr(topicTypeStr string) string {
	switch topicTypeStr {
	case "monitor":
		return "iot:monitor:stream"  // 实时数据
	case "stat":
		return "iot:stat:stream"     // 统计数据
	case "event":
		return "iot:event:stream"    // 事件数据
	case "alarm":
		return "iot:alarm:stream"    // 告警数据
	default:
		return "iot:data:stream"     // 默认
	}
}