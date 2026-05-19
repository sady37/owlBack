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

// Stream 定义常量（名称为本 var 唯一出处）
var (
	// 设备数据流（iot:*:stream）
	// 根据架构文档配置MaxTime（RetentionSeconds）
	StreamMonitor = StreamDefinition{
		Name:             "iot:monitor:stream",
		MaxLen:           1000, // 可根据实际需求调整
		RetentionSeconds: 30,   // 30秒（架构要求：Maxtime: 30秒）
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
	// 新增：卡片动态数据流
	StreamIoTCard = StreamDefinition{
		Name:             "iot:card:stream",
		MaxLen:           1000,
		RetentionSeconds: 86400, // 24小时
	}
	//  设备告警设置变更流
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
	//  新增：卡片配置变更流
	StreamConfigCard = StreamDefinition{
		Name:             "config:card:stream",
		MaxLen:           2000,
		RetentionSeconds: 0, // 不限制保留时间
	}
	// 主动 device 探测请求流。前端 refresh 按钮 → wisefido-data 写入；
	// wisefido-qinglan / wisefido-sleepace 各自 filter device_type 后调用现有 health_check 入口。
	StreamProbeDevice = StreamDefinition{
		Name:             "iot:probe:device:stream",
		MaxLen:           500,
		RetentionSeconds: 60, // 60s 足够：probe 请求是即时事件，过期就丢
	}
	//  新增：卡片实时数据流（wisefido-data消费，6Hz更新）
	// 来源：wisefido-qinglan → cardagg去重处理(分离track/vital) → 此流
	// 用途：前端实时显示 track/vital 数据
	StreamCardRealTime = StreamDefinition{
		Name:             "card:realtime:stream",
		MaxLen:           5000,
		RetentionSeconds: 6, // 6秒（处理延迟和重复消费, 确保实时性）
	}
	//  新增：卡片状态流（BedStatus, RoomStatus, ActiveAlarms, DeviceStatus）
	// 来源：wisefido-AI 或 cardagg处理后
	// 用途：前端显示床位/房间状态、告警信息、设备状态等
	StreamCardStatus = StreamDefinition{
		Name:             "card:status:stream",
		MaxLen:           2000,
		RetentionSeconds: 43200, // 12小时（长生命周期数据，告警保持）
	}

	// Deprecated: StreamCardUpdate replaced by StreamCardRealTime + StreamCardStatus.
	StreamCardUpdate = StreamDefinition{
		Name:             "card:update:stream",
		MaxLen:           5000,
		RetentionSeconds: 30,
	}

	// StreamAITrackVerdict sensor 派生的 track verdict（ghost 判定等）。
	// 来源：wisefido-sensor (RoomEngine 内 PublishAIEvent category="track_verdict")
	// 消费：wisefido-cardagg (ai_verdict_handler) 喂入 aiOverrides cache → monitor_handler.Apply 合并到 realtime
	// 切走原因：sensor → cardagg 反向桥独立化（不再走 iot:event:stream 通用通道）；
	//          为 event 流"cardagg 不再订阅"扫清障碍（B 组迁移前置）。
	StreamAITrackVerdict = StreamDefinition{
		Name:             "ai:track:verdict:stream",
		MaxLen:           500,
		RetentionSeconds: 30, // 短 TTL：verdict 是瞬时事实，超时即过期；cardagg cache 本地保留 TTL 独立
	}

	// StreamSensorDerived sensor 派生的 per-card 输出（bed/room/target/...），多 category 共流。
	// 来源：wisefido-sensor stream_publisher（payload 已是 card.BedState/RoomState/TargetState 格式；
	//       bathroom 归 room.state 带 Kind=bathroom）
	// 消费：wisefido-cardagg sensor_state_projector → card:status:<addr> hash 单 writer 投影
	// 不入库：瞬态投影，CLAUDE.md 规则 #2.1 = 专用流 + 零持久化（iot 不订阅）
	// category：bed.state / room.state / target.state（未来可扩展 vital 派生 / motion 派生）
	// 命名：'derived' 反映 sensor 在 cardagg_sensor_responsibility_split 里的 "派生 + 融合 + 时间窗" 职责。
	StreamSensorDerived = StreamDefinition{
		Name:             "sensor:derived:stream",
		MaxLen:           2000,
		RetentionSeconds: 300, // 5min：短时窗回放（cardagg 重启冷启动 replay）；超时即过期
	}

	// StreamSensorTrackStatus sensor v2 Layer 1 → Layer 2 契约：每帧每 track 的 TrackStatus 投影。
	// 来源：wisefido-sensor roomengine（每帧 ProcessFrame 后派生）
	// 消费：dev playback / 未来 zoneengine v2 / fall verifier
	// 不入库：CLAUDE.md 规则 #2.1 瞬态投影，无审计价值；高频写 → 高频 trim
	// payload schema 详 wisefido-sensor/internal/roomengine/track_status.go TrackStatus.ToStreamMap
	StreamSensorTrackStatus = StreamDefinition{
		Name:             "sensor:track:status:stream",
		MaxLen:           5000,
		RetentionSeconds: 30, // 30s：与 monitor 流同档，瞬态投影超时即丢
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
