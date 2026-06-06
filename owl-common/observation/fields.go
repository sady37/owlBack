package observation

import "owl-common/alarm"

// FieldDef defines a single observation field — the atomic unit of the system.
// All gateways translate manufacturer-specific data into these standard fields.
type FieldDef struct {
	Key     string   // 全局唯一标准名, snake_case
	Type    DataType // 数据类型
	Unit    string   // 物理单位, 无则空
	Persist bool     // true=状态(持久保留) false=瞬时(过期可丢)
	Min     int      // 合理范围下限, 0 表示不限
	Max     int      // 合理范围上限, 0 表示不限
	Enum    []string // 枚举值列表, index=value, 仅 TEnum 有效
}

// Field name constants — 全系统统一, 禁止 gateway 自定义.
const (
	// ---- Vital (生理信号) ----
	FieldHeartRate       = "heart_rate"
	FieldRespiratoryRate = "respiratory_rate"
	FieldBloodOxygen     = "blood_oxygen"
	FieldTemperature     = "temperature"
	FieldVitalConfidence = "vital_confidence"

	// ---- Motion (动作) ----
	FieldBodyMove         = "body_move"
	FieldTurnOver         = "turn_over"
	FieldSitUp            = "sit_up"
	FieldLeftRight        = "left_right"
	FieldMotionConfidence = "motion_confidence"

	// ---- Track (空间轨迹) ----
	FieldPositionX = "position_x"
	FieldPositionY = "position_y"
	FieldPositionZ = "position_z"

	FieldTrackID             = "track_id" // 0-8=人, 9=未知人, 10=空间, 11=设备,88=无人 实际radar 0-7
	FieldLogicID             = "logic_id" // Card 分配的逻辑目标 ID，cardagg 写入
	FieldTrackCount          = "track_count"
	FieldAreaID              = "area_id"
	FieldAreaType            = "area_type" // 区域类型 (2=床, 4=门, 5=监护床, 6=感应区)
	FieldTrackConfidence     = "track_confidence"
	FieldRemainingTime       = "remaining_time"
	FieldEvent               = "event"                 // track 进出事件: 0=无, 1=进房, 2=离房, 3=进区, 4=离区, 5=进入监护, 6=退出监护（与 Event 常量一致)
	FieldWalkDistance        = "walk_distance"         // 分钟级统计: 行走距离 (米)
	FieldWalkDuration        = "walk_duration"         // 分钟级统计: 行走时长 (秒)
	FieldLieDuration         = "lie_duration"          // 分钟级统计: 躺卧时长 (秒)
	FieldStandDuration       = "stand_duration"        // 分钟级统计: 站立时长 (秒)
	FieldMultiPersonDuration = "multi_person_duration" // 分钟级统计: 多人存在时长 (秒)
	FieldNumberPeople        = "number_people"         // type=3 人数事件 payload 取值字段，event_name 用 alarm.NumberPeople

	FieldPose           = "pose"
	FieldPoseConfidence = "pose_confidence"

	// CategoryTrack iot:monitor:stream 消息的 category/data_category 值，表示 track 数据（与 field/type 一致）
	CategoryTrack = "track"
	// CategoryHeart monitor 流在 monitoring 关闭时仅作在线心跳（如 track_id=11），与真实轨迹区分
	CategoryHeart = "heart"
	// CategorySleep stat 解码的 sleep 块 data_category 值（睡眠/生命体征统计）
	CategorySleep = "sleep"


	// ---- Confidence (置信度, 0-100) ----

	// ---- Device (设备质量) ----
	FieldSignalQuality  = "signal_quality"
	FieldBattery        = "battery"
	FieldMeasureQuality = "measure_quality" // 测量稳定度 (0-100%, 网关归一化, 雷达原始0-3)

	// ---- Status（周期≤5min 上报时视为实时，可出现在 monitor/track 中） ----
	FieldOffline             = "offline"     // 1=离线(异常) 0=在线
	FieldBedStatus           = "bed_status"  // int: 0=在床, 1=离床, 8=未改变/待机
	FieldSleepStage          = "sleep_stage" // 睡眠阶段：0=无数据/已清, 1=清醒, 2=浅睡, 4=深睡, 8=未知。厂家 MQTT 入口仅 1/2/4/8；0 是 sensor 内部投影"已清"语义（initial publish / OnBedVacant 清零，详 card.SleepStageInitial）。
	FieldInitStatus          = "init_status"
	FieldDeviceFailure       = "device_failure"
	FieldDetached            = "detached"
	FieldAngleAbnormal       = "angle_abnormal"        // 倾角异常 (雷达)
	FieldSignalPoor          = "signal_poor"           // 信号差 (bool, 区别于 signal_quality %)
	FieldWeakBiometricSignal = "weak_biometric_signal" // 生命体征信号弱风险度 0-100 百分制；设备原始 0-3 转为 raw*20（3→60）；30 分钟大周期内可与 hr/rr 报警次数相加
	FieldActivity = "activity" // 老人活动性统计：walk_distance, walk_duration, lie_duration, stand_duration 等	

	// ---- Event (事件与报警统一模型, 5W要素) ----
	// Who  = Header.CardID
	// When = event_since (物理发生时间, 区别于 Header.Timestamp 传输时间)
	// What = event_name  (语义标签: Fall, InBed, HeartRateHigh, ...)
	// Status = event_status (生命周期: instant/start/end)
	// Level  = event_level  (严重等级: INFO → CRITICAL)
	FieldTopic       = "topic"
	FieldCategory    = "category"
	FieldDataCategory = "data_category"
	FieldEventID     = "event_id"     // 唯一标识 (关联/闭环用)
	FieldEventName   = "event_name"   // What: 事件语义名称
	FieldEventSince  = "event_since"  // When: 发生时间戳 (毫秒)
	FieldEventStatus = "event_status" // Status: 生命周期阶段
	FieldEventLevel  = "event_level"  // Level: 严重等级
	FieldEventValue  = "event_value"  // 触发时的测量值 (如 heart_rate=135)
	FieldEventEnd    = "event_end"    // 结束时间戳 (毫秒, 仅 status=end 时填)
	FieldEventReason = "event_reason" // 解除/触发原因 (auto_resolved, manual, timeout)

	// ---- Command (控制面ACK) ----
	FieldCommandID       = "cmd_id"       // 指令唯一ID (关联下发记录)
	FieldCommandAction   = "cmd_action"   // 指令动作名 (如 light_on, alarm_mute)
	FieldCommandResult   = "cmd_result"   // 执行结果: success/fail/timeout
	FieldCommandDetail   = "cmd_detail"   // 失败原因或附加信息
	FieldCommandProgress = "cmd_progress" // 执行进度 0-100% (如固件升级)
)



// Enum value lists
var (
	EnumLeftRight  = []string{"none", "left", "right"}
	EnumBedStatus  = []string{"out", "in"}
	// Sleepad 标准：取值 1=清醒, 2=浅睡, 4=深睡, 8=未知；Enum 下标=取值，用于展示
	EnumSleepStage = []string{"", "awake", "light", "", "deep", "", "", "", "unknown"}
	EnumInitStatus = []string{"uninit", "initializing", "ready", "error"}
	// EnumPose 与 Pose* 常量下标一致：0-11=雷达姿态，12=跑步（body_move/turn_over 为独立 int 次数字段，不入 pose）
	// 0:初始化 1:行走 2:疑似跌倒 3:蹲坐 4:站立 5:跌倒确认 6:卧床 7:疑似坐地 8:确认坐地 9:普通床上坐起 10:疑似床上坐起 11:确认床上坐起 12:running
	EnumPose = []string{
		"unknown", "walking", "suspected_fall", "sitting", "standing", "fallen", "lying",
		"suspected_sitting_ground", "sitting_ground", "bed_sit_up", "suspected_bed_sit_up", "confirmed_bed_sit_up",
		"running",
	}
	EnumEventStatus = []string{"instant", "start", "end", "pending", "tracking"}
	// EnumEventLevel 与 gateway/使能表 event_level 一致；级别常量以 owl-common/alarm 为准，此处仅作校验枚举。
	EnumEventLevel    = []string{"INFO", "NOTICE", "WARNING", "ERROR", "CRITICAL", "EMERG", "ALERT"}
	EnumCommandResult = []string{"success", "fail", "timeout"}
	// EnumAreaType 区域类型（track/declare 协议 0-6），与 FieldAreaType 一致
	EnumAreaType = []string{"none", "custom", "bed", "interfer", "door", "monitor_bed", "sensing"}
	// EnumPoseDisplay Pose 0-12 的 display 字符串（PascalCase）
	EnumPoseDisplay = []string{"Initialization", "Walking", "SuspectedFall", "Sitting", "Standing", "Fall", "Lying", "SuspectedSittingOnGround", "SittingOnGround", "BedSitUp", "SuspectedBedSitUp", "ConfirmedBedSitUp", "Running"}
)

// NormalizeEventLevel 将 CRIT/ERR 等别名转为标准全名；实现见 owl-common/alarm.NormalizeAlarmLevel。
func NormalizeEventLevel(level string) string {
	return alarm.NormalizeAlarmLevel(level)
}

// event_name 取值唯一定义在 owl-common/alarm 包，此处仅保留字段名 FieldEventName。

// Registry — 全量字段注册表. 消费方通过 Key 查表获取完整元数据.
var Registry = map[string]*FieldDef{
	// ---- Vital ----
	FieldHeartRate:       {Key: FieldHeartRate, Type: TInt, Unit: "bpm", Persist: false, Min: 30, Max: 220},
	FieldRespiratoryRate: {Key: FieldRespiratoryRate, Type: TInt, Unit: "rpm", Persist: false, Min: 5, Max: 60},
	FieldBloodOxygen:     {Key: FieldBloodOxygen, Type: TInt, Unit: "%", Persist: false, Min: 70, Max: 100},
	FieldTemperature:     {Key: FieldTemperature, Type: TFloat, Unit: "°C", Persist: false, Min: 34, Max: 42},

	// ---- Motion（body_move/turn_over 为次数 int，非动作枚举） ----
	FieldBodyMove:  {Key: FieldBodyMove, Type: TInt, Persist: false},  // 体动次数，适用于 BM8701-2 固件≥5.x 及 M901L
	FieldTurnOver:  {Key: FieldTurnOver, Type: TInt, Persist: false},  // 翻身次数，适用于 BM8701-2 固件≤2.x
	FieldSitUp:     {Key: FieldSitUp, Type: TBool, Persist: false},
	FieldLeftRight: {Key: FieldLeftRight, Type: TEnum, Persist: false, Enum: EnumLeftRight},

	// ---- Track ----
	FieldPositionX:           {Key: FieldPositionX, Type: TInt, Unit: "cm", Persist: false, Max: 2000},
	FieldPositionY:           {Key: FieldPositionY, Type: TInt, Unit: "cm", Persist: false, Max: 2000},
	FieldPositionZ:           {Key: FieldPositionZ, Type: TInt, Unit: "cm", Persist: false, Max: 500},
	FieldPose:                {Key: FieldPose, Type: TEnum, Persist: true, Enum: EnumPose},
	FieldTrackID:             {Key: FieldTrackID, Type: TInt, Persist: false, Max: 88},
	FieldLogicID:             {Key: FieldLogicID, Type: TStr, Persist: false},
	FieldEvent:               {Key: FieldEvent, Type: TInt, Persist: false, Max: 6}, // 0=无, 1-6=EventInRoom/OutRoom/InArea/OutArea/EnterMonitor/ExitMonitor
	FieldTrackCount:          {Key: FieldTrackCount, Type: TInt, Persist: true, Max: 8},
	FieldAreaID:              {Key: FieldAreaID, Type: TInt, Persist: false},
	FieldAreaType:            {Key: FieldAreaType, Type: TEnum, Persist: false, Enum: EnumAreaType},
	FieldRemainingTime:       {Key: FieldRemainingTime, Type: TInt, Unit: "s", Persist: false},
	FieldWalkDistance:        {Key: FieldWalkDistance, Type: TInt, Unit: "m", Persist: false},
	FieldWalkDuration:        {Key: FieldWalkDuration, Type: TInt, Unit: "s", Persist: true, Max: 900},
	FieldLieDuration:         {Key: FieldLieDuration, Type: TInt, Unit: "s", Persist: true},
	FieldStandDuration:       {Key: FieldStandDuration, Type: TInt, Unit: "s", Persist: true, Max: 600},
	FieldMultiPersonDuration: {Key: FieldMultiPersonDuration, Type: TInt, Unit: "s", Persist: true},
	FieldNumberPeople:        {Key: FieldNumberPeople, Type: TInt, Persist: true, Max: 8}, // type=3 人数事件取值，与 alarm.NumberPeople 同键 "number-people"

	// ---- Confidence ----
	FieldTrackConfidence:  {Key: FieldTrackConfidence, Type: TInt, Unit: "%", Persist: false, Max: 100},
	FieldPoseConfidence:   {Key: FieldPoseConfidence, Type: TInt, Unit: "%", Persist: false, Max: 100},
	FieldVitalConfidence:  {Key: FieldVitalConfidence, Type: TInt, Unit: "%", Persist: false, Max: 100},
	FieldMotionConfidence: {Key: FieldMotionConfidence, Type: TInt, Unit: "%", Persist: false, Max: 100},

	// ---- Device ----
	FieldSignalQuality:       {Key: FieldSignalQuality, Type: TInt, Unit: "%", Persist: true, Max: 100},
	FieldBattery:             {Key: FieldBattery, Type: TInt, Unit: "%", Persist: true, Max: 100},
	FieldMeasureQuality:      {Key: FieldMeasureQuality, Type: TInt, Unit: "%", Persist: false, Max: 100},
	FieldDeviceFailure:       {Key: FieldDeviceFailure, Type: TBool, Persist: true},
	FieldDetached:            {Key: FieldDetached, Type: TBool, Persist: true},
	FieldAngleAbnormal:       {Key: FieldAngleAbnormal, Type: TBool, Persist: true},
	FieldSignalPoor:          {Key: FieldSignalPoor, Type: TBool, Persist: true},
	FieldWeakBiometricSignal: {Key: FieldWeakBiometricSignal, Type: TInt, Persist: true, Max: 100}, // 0-100 百分制

	// ---- Status（bed_status/sleep_stage 周期上报时在 monitor/track 中作为实时量） ----
	FieldOffline:     {Key: FieldOffline, Type: TBool, Persist: true},
	FieldBedStatus:   {Key: FieldBedStatus, Type: TInt, Persist: true}, // 0=在床, 1=离床, 8=未改变
	FieldSleepStage:  {Key: FieldSleepStage, Type: TEnum, Persist: true, Enum: EnumSleepStage},
	FieldInitStatus:  {Key: FieldInitStatus, Type: TEnum, Persist: true, Enum: EnumInitStatus},
	FieldTopic:	  {Key: FieldTopic, Type: TStr, Persist: true},
	FieldCategory:   {Key: FieldCategory, Type: TStr, Persist: true},
	FieldDataCategory: {Key: FieldDataCategory, Type: TStr, Persist: true},

	FieldEventID:     {Key: FieldEventID, Type: TStr, Persist: true},
	FieldEventName:   {Key: FieldEventName, Type: TStr, Persist: true},
	FieldEventSince:  {Key: FieldEventSince, Type: TInt, Unit: "ms", Persist: true},
	FieldEventStatus: {Key: FieldEventStatus, Type: TEnum, Persist: true, Enum: EnumEventStatus},
	FieldEventLevel:  {Key: FieldEventLevel, Type: TEnum, Persist: true, Enum: EnumEventLevel},
	FieldEventValue:  {Key: FieldEventValue, Type: TFloat, Persist: true},
	FieldEventEnd:    {Key: FieldEventEnd, Type: TInt, Unit: "ms", Persist: true},
	FieldEventReason: {Key: FieldEventReason, Type: TStr, Persist: true},

	// ---- Command (控制面ACK) ----
	FieldCommandID:       {Key: FieldCommandID, Type: TStr, Persist: true},
	FieldCommandAction:   {Key: FieldCommandAction, Type: TStr, Persist: true},
	FieldCommandResult:   {Key: FieldCommandResult, Type: TEnum, Persist: true, Enum: EnumCommandResult},
	FieldCommandDetail:   {Key: FieldCommandDetail, Type: TStr, Persist: false},
	FieldCommandProgress: {Key: FieldCommandProgress, Type: TInt, Unit: "%", Persist: false, Max: 100},
}

// TrackID 哨兵值：0-8=已识别的人, 9=未知人, 10=空间/床/房间级, 11=设备级, 88=无效上界
const (
	TrackUnknownPerson = 9
	TrackSpace         = 10
	TrackDevice        = 11
	TrackInvalid       = 88
)

// BedStatus 取值（bed_status field；firmware + sensor 投影）。无床概念设备（radar）= Unchanged。
const (
	BedStatusInBed     = 0
	BedStatusLeftBed   = 1
	BedStatusUnchanged = 8
)

func IsKnownPerson(id int) bool   { return id >= 0 && id <= 8 }
func IsUnknownPerson(id int) bool { return id == TrackUnknownPerson }
func IsPerson(id int) bool        { return id >= 0 && id <= TrackUnknownPerson }
func IsSpace(id int) bool         { return id == TrackSpace }
func IsDevice(id int) bool        { return id == TrackDevice }

// Pose 数值 0-12：0-11=雷达姿态，12=跑步（body_move/turn_over 为独立次数字段）
const (
	PoseUnknown = iota      // 0: 初始化
	PoseWalking             // 1: 行走
	PoseSuspectedFall       // 2: 疑似跌倒
	PoseSitting             // 3: 蹲坐
	PoseStanding            // 4: 站立
	PoseFallen              // 5: 跌倒确认
	PoseLying               // 6: 卧床
	PoseSuspectedSitGround  // 7: 疑似坐地
	PoseSitGround           // 8: 确认坐地
	PoseBedSitUp            // 9: 普通床上坐起，Sleepad sit_up 可映射到此
	PoseSuspectedBedSitUp   // 10: 疑似床上坐起
	PoseConfirmedBedSitUp   // 11: 确认床上坐起
	PoseRunning             // 12: 跑步
)

// PoseToLabel 根据厂家 pose 数值 (0-11) 返回 EnumPose 语义标签，如 5 → "fallen"。越界返回空串。
func PoseToLabel(pose int) string {
	if pose >= 0 && pose < len(EnumPose) {
		return EnumPose[pose]
	}
	return ""
}

// IsSuspectedPose returns true for pose values that indicate uncertain detection.
func IsSuspectedPose(pose int) bool {
	return pose == PoseSuspectedFall || pose == PoseSuspectedSitGround || pose == PoseSuspectedBedSitUp
}

// PoseNumToDisplay returns display string (PascalCase) for pose value 0-12.
func PoseNumToDisplay(n int) string {
	if n >= 0 && n < len(EnumPoseDisplay) {
		return EnumPoseDisplay[n]
	}
	return ""
}

// PoseDisplayToNum returns pose value 0-12 for display string, or (0, false) if unknown.
func PoseDisplayToNum(s string) (int, bool) {
	for i, v := range EnumPoseDisplay {
		if v == s {
			return i, true
		}
	}
	return 0, false
}

// AreaType 字段值（track/declare_area type），与协议一致
const (
	AreaTypeNone       = 0
	AreaTypeCustom     = 1
	AreaTypeBed        = 2
	AreaTypeInterfer   = 3
	AreaTypeDoor       = 4
	AreaTypeMonitorBed = 5
	AreaTypeSensing    = 6
)

// DeclareAreaType 区域类型（declare_area 0-5），与协议 3.4.3 一致
const (
	DeclareAreaInvalid    = 0
	DeclareAreaCustom     = 1
	DeclareAreaBed        = 2
	DeclareAreaNoise      = 3
	DeclareAreaDoor       = 4
	DeclareAreaMonitorBed = 5
)

// Event 进出事件字段值（type=1 track 字节 14），与协议 3.5 一致
const (
	EventInRoom       = 1
	EventOutRoom      = 2
	EventInArea       = 3
	EventOutArea      = 4
	EventEnterMonitor = 5
	EventExitMonitor  = 6
)

// Lookup returns the FieldDef for a key, or nil if not registered.
func Lookup(key string) *FieldDef {
	return Registry[key]
}
