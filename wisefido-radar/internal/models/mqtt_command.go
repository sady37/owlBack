package models

// PropertyReadCommand 属性读取命令
// 参考协议文档 3.4 节
type PropertyReadCommand struct {
	Cmd       string   `json:"cmd"`       // "read"
	RequestID string   `json:"requestId"` // 请求 ID
	Data      struct { // 数据
		Key []string `json:"key"` // 要读取的属性 key 列表，空数组表示读取所有属性
	} `json:"data"`
}

// PropertyUpdateCommand 属性设置命令
// 参考协议文档 3.4 节
type PropertyUpdateCommand struct {
	Cmd       string                 `json:"cmd"`       // "update"
	RequestID string                 `json:"requestId"` // 请求 ID
	Data      map[string]interface{} `json:"data"`     // 属性 key-value 对
}

// PropertyResponse 属性响应
// 参考协议文档 3.4 节
type PropertyResponse struct {
	Cmd       string                 `json:"cmd"`       // "read" 或 "update"
	Code      int                    `json:"code"`     // 状态码，200 表示成功
	RequestID string                 `json:"requestId"` // 请求 ID（与命令中的 requestId 对应）
	Data      map[string]interface{} `json:"data"`     // 属性 key-value 对
}

// MonitorSubscriptionCommand 实时数据订阅命令
// 参考协议文档 3.7.1 节
type MonitorSubscriptionCommand struct {
	Cmd  string `json:"cmd"` // "subscription"
	Data struct {
		Content  interface{} `json:"content"`  // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率（注意：协议文档中是数字，但测试程序中使用字符串）
		Duration int         `json:"duration"` // 订阅时长（秒），最大 3600 秒
	} `json:"data"`
}

// MonitorData 实时数据
// 参考协议文档 3.7.2 节
type MonitorData struct {
	Cmd  string `json:"cmd"` // "subscription"（注意：协议文档中拼写为 "subscriotion"）
	Data struct {
		Track string `json:"track"` // base64 编码的轨迹数据
		Bh    string `json:"bh"`    // base64 编码的呼吸心率数据
	} `json:"data"`
}

// FunctionControlCommand 功能调用命令（重启等）
// 参考协议文档 3.8.1 节
type FunctionControlCommand struct {
	Cmd       string `json:"cmd"`       // "control"
	RequestID string `json:"requestId"` // 请求 ID
	Data      struct {
		Dev int `json:"dev"` // 0-重启雷达和主控，1-只重启雷达，2-只重启主控，100-清除设备数据，101-清除雷达数据，102-清除主控数据
	} `json:"data"`
}

// FunctionControlResponse 功能调用响应
// 参考协议文档 3.8.1 节
type FunctionControlResponse struct {
	Cmd       string                 `json:"cmd"`       // "control_return"
	RequestID string                 `json:"requestId"` // 请求 ID
	Code      int                    `json:"code"`     // 状态码，200 表示成功
	Data      map[string]interface{} `json:"data"`     // 响应数据（通常为空）
}

// StatData 统计数据
// 参考协议文档 3.6 节
type StatData struct {
	Cmd  string `json:"cmd"` // "sleep_trajectory"
	Data struct {
		Sleep string `json:"sleep"` // base64 编码的睡眠统计数据（16 字节）
		Track string `json:"track"` // base64 编码的轨迹统计数据（16 字节）
	} `json:"data"`
}

// EventData 事件数据
// 参考协议文档 3.5 节
type EventData struct {
	Cmd  string      `json:"cmd"`  // "event" 或 "log"
	Type int         `json:"type"` // 事件类型：1-进出事件，2-姿态变化事件，3-人数变化事件
	Data interface{} `json:"data"` // 事件数据（格式根据 type 不同而不同）
}

// AlarmData 告警数据
// 参考协议文档 3.5 节（告警通过 /prefix/alarm/productId/UID/post 主题上报）
type AlarmData struct {
	Cmd  string      `json:"cmd"`  // 告警命令
	Type int         `json:"type"` // 告警类型
	Data interface{} `json:"data"` // 告警数据
}

