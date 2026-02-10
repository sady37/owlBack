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
		MaxLen:           1000, // 可根据实际需求调整
		RetentionSeconds: 30,   // 30秒（架构要求：Maxtime: 30秒）
	}
	StreamStat = StreamDefinition{
		Name:             "iot:stat:stream",
		MaxLen:           1000, // 可根据实际需求调整
		RetentionSeconds: 300,  // 5分钟（架构要求：Maxtime: 5min，仅Radar支持，1分钟/次）
	}
	StreamEvent = StreamDefinition{
		Name:             "iot:event:stream",
		MaxLen:           500,   // event数量有限，减小MaxLen
		RetentionSeconds: 86400, // 24小时（架构要求：Maxtime: 24H）
	}
	StreamAlarm = StreamDefinition{
		Name:             "iot:alarm:stream",
		MaxLen:           500,   // alarm数量有限，减小MaxLen
		RetentionSeconds: 86400, // 24小时（Maxtime: 24H，事件触发）
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
	// ⭐ 新增：卡片动态数据流
	StreamIoTCard = StreamDefinition{
		Name:             "iot:card:stream",
		MaxLen:           1000,
		RetentionSeconds: 86400, // 24小时
	}
	// ⭐ 新增：设备状态流（从 config:device_status:stream 改为 iot:deviceStatus:stream）
	StreamIoTDeviceStatus = StreamDefinition{
		Name:             "iot:deviceStatus:stream",
		MaxLen:           1000,
		RetentionSeconds: 0, // 不限制保留时间
	}

    //⭐ 新增：设备告警设置变更流
	StreamConfigAlarmDevice = StreamDefinition{
		Name:             "config:alarmDevice:stream",
		MaxLen:           1000,
		RetentionSeconds: 0, // 不限制保留时间
	}
	// ⭐ 新增：报警处理流
	StreamConfigAlarmProcess = StreamDefinition{
		Name:             "config:alarmProcess:stream",    // 重命名
		MaxLen:           1000,
		RetentionSeconds: 0, // 不限制保留时间
	}
	// ⭐ 新增：卡片配置变更流
	StreamConfigCard = StreamDefinition{
		Name:             "config:card:stream",
		MaxLen:           2000,
		RetentionSeconds: 0, // 不限制保留时间
	}
	// ⭐ 新增：卡片实时数据流（wisefido-data消费，6Hz更新）
	// 来源：wisefido-qinglan → cardagg去重处理(分离track/vital) → 此流
	// 用途：前端实时显示 track/vital 数据
	StreamCardRealTime = StreamDefinition{
		Name:             "card:realtime:stream",
		MaxLen:           5000,
		RetentionSeconds: 6, // 6秒（处理延迟和重复消费, 确保实时性）
	}
	// ⭐ 新增：卡片状态流（BedStatus, RoomStatus, ActiveAlarms, DeviceStatus）
	// 来源：wisefido-AI 或 cardagg处理后
	// 用途：前端显示床位/房间状态、告警信息、设备状态等
	StreamCardStatus = StreamDefinition{
		Name:             "card:status:stream",
		MaxLen:           2000,
		RetentionSeconds: 43200, // 12小时（长生命周期数据，告警保持）
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
