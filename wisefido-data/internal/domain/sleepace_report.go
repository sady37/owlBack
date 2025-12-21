package domain

// SleepaceReport Sleepace 睡眠报告领域模型
// 参考：wisefido-backend/wisefido-sleepace/models/report.go
type SleepaceReport struct {
	ReportID    string `json:"report_id"`    // UUID
	TenantID    string `json:"tenant_id"`   // UUID
	DeviceID    string `json:"device_id"`    // UUID（如果为空，可通过 DeviceCode 匹配 devices 表获取）
	DeviceCode  string `json:"device_code"`  // 设备编码（来自厂家，等价于 devices.serial_number 或 devices.uid）
	
	// 报告基本信息
	RecordCount int    `json:"record_count"` // 记录数量
	StartTime   int64  `json:"start_time"`   // 开始时间（Unix 时间戳，秒）
	EndTime     int64  `json:"end_time"`      // 结束时间（Unix 时间戳，秒）
	Date        int    `json:"date"`         // 日期（YYYYMMDD 格式，如 20240820）
	
	// 报告配置
	StopMode    int    `json:"stop_mode"`    // 停止模式
	TimeStep    int    `json:"time_step"`    // 时间步长（秒）
	Timezone    int    `json:"timezone"`     // 时区偏移（秒）
	
	// 报告数据
	SleepState  string `json:"sleep_state"` // 睡眠状态数组（JSON 字符串，如 "[1,2,1,1,1,...]"）
	Report      string `json:"report"`      // 完整报告数据（JSON 字符串）
	
	// 时间戳
	CreatedAt   int64  `json:"created_at"`   // 创建时间（Unix 时间戳，秒）
	UpdatedAt   int64  `json:"updated_at"`   // 更新时间（Unix 时间戳，秒）
}

