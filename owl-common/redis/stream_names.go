package redis

import (
	commonconfig "owl-common/config"
)

// StreamDefinition Stream 定义（名称 + 配置）
type StreamDefinition struct {
	Name             string
	MaxLen           int64
	RetentionSeconds int
}

// Stream 定义常量
var (
	// 设备数据流（iot:*:stream）
	// 根据架构文档配置MaxTime（RetentionSeconds）
	StreamMonitor = StreamDefinition{
		Name:             "iot:monitor:stream",
		MaxLen:           1000,        // 可根据实际需求调整
		RetentionSeconds: 30,          // 30秒（架构要求：Maxtime: 30秒）
	}
	StreamStat = StreamDefinition{
		Name:             "iot:stat:stream",
		MaxLen:           1000,        // 可根据实际需求调整
		RetentionSeconds: 300,         // 5分钟（架构要求：Maxtime: 5min，仅Radar支持，1分钟/次）
	}
	StreamEvent = StreamDefinition{
		Name:             "iot:event:stream",
		MaxLen:           500,         // event数量有限，减小MaxLen
		RetentionSeconds: 86400,       // 24小时（架构要求：Maxtime: 24H）
	}
	StreamAlarm = StreamDefinition{
		Name:             "iot:alarm:stream",
		MaxLen:           500,         // alarm数量有限，减小MaxLen
		RetentionSeconds: 86400,       // 24小时（Maxtime: 24H，事件触发）
	}
	StreamAuth = StreamDefinition{
		Name:             "iot:auth:stream",
		MaxLen:           1000,
		RetentionSeconds: 86400,
	}
	StreamOther = StreamDefinition{
		Name:             "iot:other:stream",
		MaxLen:           1000,
		RetentionSeconds: 86400,
	}

	// 配置变更流（config:*:stream）
	StreamConfigAlarmCloud = StreamDefinition{
		Name:             "config:alarm_cloud:stream",
		MaxLen:           1000,
		RetentionSeconds: 0, // 不限制保留时间
	}
	StreamConfigAlarmDevice = StreamDefinition{
		Name:             "config:alarm_device:stream",
		MaxLen:           1000,
		RetentionSeconds: 0, // 不限制保留时间
	}
	// 设备在线状态流（统一使用 config:device_status:stream）
	StreamConfigDeviceStatus = StreamDefinition{
		Name:             "config:device_status:stream",
		MaxLen:           1000,
		RetentionSeconds: 0, // 不限制保留时间
	}
)

// GetStreamConfig 获取 stream 配置（支持从配置覆盖）
func GetStreamConfig(stream StreamDefinition, streamsConfig *commonconfig.StreamsConfig) (maxLen int64, retentionSeconds int) {
	maxLen = stream.MaxLen
	retentionSeconds = stream.RetentionSeconds

	if streamsConfig != nil {
		if streamConfig, ok := streamsConfig.Streams[stream.Name]; ok {
			if streamConfig.MaxLen > 0 {
				maxLen = streamConfig.MaxLen
			}
			if streamConfig.RetentionSeconds > 0 {
				retentionSeconds = streamConfig.RetentionSeconds
			}
		} else if streamsConfig.Default.MaxLen > 0 {
			maxLen = streamsConfig.Default.MaxLen
		} else if streamsConfig.Default.RetentionSeconds > 0 {
			retentionSeconds = streamsConfig.Default.RetentionSeconds
		}
	}

	return maxLen, retentionSeconds
}
