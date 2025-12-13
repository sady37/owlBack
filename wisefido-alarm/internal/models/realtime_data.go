package models

// RealtimeData 融合后的实时数据（从 Redis 读取，与 sensor-fusion 保持一致）
type RealtimeData struct {
	// 生命体征
	Heart        *int    `json:"heart"`         // 融合后的心率
	Breath       *int    `json:"breath"`       // 融合后的呼吸率
	HeartSource  string  `json:"heart_source"`  // 数据来源："Sleepace" 或 "Radar"
	BreathSource string  `json:"breath_source"` // 数据来源："Sleepace" 或 "Radar"
	HeartTimestamp *int64 `json:"heart_timestamp,omitempty"` // 心率数据的时间戳
	BreathTimestamp *int64 `json:"breath_timestamp,omitempty"` // 呼吸率数据的时间戳
	
	// 睡眠状态
	SleepStage   *string `json:"sleep_stage"`   // SNOMED 编码
	BedStatus    *string `json:"bed_status"`    // SNOMED 编码
	SleepStageSource string `json:"sleep_stage_source,omitempty"` // 睡眠状态数据来源
	BedStatusSource string `json:"bed_status_source,omitempty"` // 床状态数据来源
	SleepStageTimestamp *int64 `json:"sleep_stage_timestamp,omitempty"` // 睡眠状态数据的时间戳
	BedStatusTimestamp *int64 `json:"bed_status_timestamp,omitempty"` // 床状态数据的时间戳
	
	// 姿态数据（来自 Radar）
	PersonCount  int     `json:"person_count"`  // 人数（tracking_id 数量）
	Postures     []Posture `json:"postures"`   // 姿态列表
	
	// 时间戳
	Timestamp    int64   `json:"timestamp"`    // Unix 时间戳（融合结果的时间戳）
}

// Posture 姿态数据
type Posture struct {
	TrackingID   string `json:"tracking_id"`   // Radar tracking_id
	PostureCode  string `json:"posture_code"`  // SNOMED 编码
	PostureDisplay string `json:"posture_display"` // 显示名称
}

