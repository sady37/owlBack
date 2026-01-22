package models

import "encoding/json"

// ReceivedMessage Sleepace MQTT 消息结构（v1.0 格式）
type ReceivedMessage struct {
	DeviceId  string          `json:"deviceId"`  // 设备代码（device_code）
	DataKey   string          `json:"dataKey"`   // 数据类型：realtime, connectionStatus, sleepStage, alarmNotify 等
	TimeStamp int64           `json:"timestamp"` // 时间戳
	Data      json.RawMessage `json:"data"`      // 数据内容（JSON）
}

// RealtimeData 实时数据
type RealtimeData struct {
	CommonData
	LeftRight     int `json:"leftRight"`
	Breath        int `json:"breath"`        // 呼吸率
	Heart         int `json:"heart"`         // 心率
	TurnOver      int `json:"turnOver"`      // 翻身
	BodyMove      int `json:"bodyMove"`      // 体动
	SitUp         int `json:"sitUp"`          // 坐起
	InitStatus    int `json:"initStatus"`    // 初始化状态
	BedStatus     int `json:"bedStatus"`     // 床状态：0=在床, 1=离床
	SignalQuality int `json:"signalQuality"` // 信号质量
}

// SleepStageData 睡眠阶段数据
type SleepStageData struct {
	CommonData
	LeftRight  int `json:"leftRight"`
	SleepStage int `json:"sleepStage"` // 0=清醒, 1=浅睡眠, 2=深睡眠, 3=REM睡眠
}

// ConnectionStatusData 连接状态数据
type ConnectionStatusData struct {
	CommonData
	ConnectionStatus int `json:"connectionStatus"` // 0=离线, 1=在线
}

// AlarmNotifyData 报警通知数据
type AlarmNotifyData struct {
	CommonData
	Id            int64  `json:"id"`
	Type          string `json:"type"`          // 报警类型
	Status        int    `json:"status"`       // 0=触发, 1=解除
	UserId        string `json:"userId"`        // 用户ID
	RelieveReason string `json:"relieveReason"` // 解除原因
	RelieveTime   int64  `json:"relieveTime"`  // 解除时间
}

// CommonData 通用数据字段
type CommonData struct {
	DeviceId  string `json:"deviceId"`
	TimeStamp int64  `json:"timestamp"`
}

