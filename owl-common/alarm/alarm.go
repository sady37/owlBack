package alarm

import (
	"fmt"
	"strconv"
	"strings"
)

// ========== Constants (常量) ==========

// Common Alarm default values (通用报警默认值)
// 这些是所有设备类型都支持的通用报警项，存储在 alarm_cloud 表的 OfflineAlarm, LowBattery, DeviceFailure 字段
// 使用统一的报警级别常量，保持一致性
const (
	DefaultOfflineAlarm  = AlarmLevelErr // 设备离线报警默认级别
	DefaultLowBattery    = AlarmLevelErr // 低电量报警默认级别（设备类统一 ERROR，便于通过 color 区分）
	DefaultDeviceFailure = AlarmLevelErr // 设备故障报警默认级别
)

// Vital Sign Alert Thresholds default values (生理指标阈值默认值)
const (
	DefaultRespiratoryRateMin = 8  // 呼吸率nomal最小值默认值
	DefaultRespiratoryRateMax = 24 // 呼吸率nomal最大值默认值
	DefaultHeartRateMin       = 50 // 心率正nomal小值默认值
	DefaultHeartRateMax       = 95 // 心率nomal最大值默认值
)

// 事件 payload 字段名（iot:event:stream data 的 key，由 alarm 域定义）
const (
	FieldEventName   = "event_name"
	FieldEventStatus = "event_status"
)

// ========== ProcessType 常量定义 ==========
const (
	ProcessTypeImmediate            = "immediate"              // 立即触发
	ProcessTypeTimeBased            = "time_based"             // 基于时间阈值
	ProcessTypeStateBased           = "state_based"            // 基于状态变化
	ProcessTypeActivityMonitoring   = "activity_monitoring"    // 活动监控
	ProcessTypeConditionalTimeBased = "conditional_time_based" // 条件时间型
	ProcessTypeBedStateChange       = "bed_state_change"       // 床位状态变化
	ProcessTypeRoomStateChange      = "room_state_change"      // 房间状态变化
	ProcessTypeDataCategory         = "data_category"          // 数据类别（仅路由，不直接产生报警：track/vital/sleep_stage/number-people/activity）
)

// ========== 报警类型 ==========

const (
	AlarmTypeOffline        = "Offline"
	AlarmTypeOfflineRecover = "OfflineRecover"
	AlarmTypeDeviceFailure  = "DeviceFailure"
	AlarmTypeDeviceRecover  = "DeviceRecover"
	AlarmTypeUnknown        = "Unknown"
	SignalPoor              = "SignalPoor"
	SingalPoorRecover       = "SignalPoorRecover"
	AngleException          = "AngleException"
	AngleExceptionRecover   = "AngleExceptionRecover"
	SensorDetached          = "SensorDetached"
	SensorDetachedRecover   = "SensorDetachedRecover"
	PressureSensor          = "pressureSensor" // Sleepad 压力传感器状态事件

	// 通用报警类型（不区分设备，设备类型通过 device_type 字段区分）
	ApneaHypopnea         = "ApneaHypopnea"
	HeartRateAlert        = "HeartRateAlert"
	HeartRateAlertHigh    = "HeartRateAlert.High"
	HeartRateAlertLow     = "HeartRateAlert.Low"
	RespRateAlert         = "RespRateAlert"
	RespRateAlertHigh     = "RespRateAlert.High"
	RespRateAlertLow      = "RespRateAlert.Low"
	HeartRateNormal       = "HeartRateNormal"
	RespiratoryRateNormal = "RespiratoryRateNormal"
	LeftBed               = "LeftBed"
	AbnormalBodyMovement  = "AbnormalBodyMovement"
	NoBodyMove            = "NoBodyMove"
	NoTurnOver            = "NoTurnOver"
	ResetTime             = "ResetTime"
	NapTime               = "NapTime"
	SleepadSetting        = "SleepadSetting"
	MaterialSetting       = "MaterialSetting"

	Fall                     = "Fall"
	SuspectedFall            = "SuspectedFall"
	SittingOnGround          = "SittingOnGround"
	SuspectedSittingOnGround = "SuspectedSittingOnGround"
	BedSitUp                 = "BedSitUp"
	SuspectedBedSitUp        = "SuspectedBedSitUp"
	WeakBiometricSignal      = "WeakBiometricSignal"
	InBed                    = "InBed"
	Stay                     = "Stay"
	NightAbsence             = "NightAbsence"    // Radar 用：整夜不在 room（房间层面）
	BedNightAbsence          = "BedNightAbsence" // Sleepad 用：整夜不在床（床层面；床上没人 ≠ 房间没人）
	WarningArea              = "WarningArea"
	EnterRoom                = "EnterRoom"
	ExitRoom                 = "ExitRoom"
	EnterSensingArea         = "EnterSensingArea"
	ExitSensingArea          = "ExitSensingArea"
	EnterMonitor             = "EnterMonitor"
	ExitMonitor              = "ExitMonitor"
	MonitoringMode           = "MonitoringMode"
	PostureDetection         = "PostureDetection"
	Initialization           = "Initialization"
	Walking                  = "Walking"
	Sitting                  = "Sitting"
	Standing                 = "Standing"
	Lying                    = "Lying"

	// 事件/数据类别名（event_name 或 data_category，与 gateway 一致）
	NumberPeople = "number_people"
	Sleep        = "sleep"
	SleepStage   = "sleep-stage"
	Track        = "track"
	Activity     = "activity" // 老人活动性统计：walk_distance, walk_duration, lie_duration, stand_duration 等
)

// EndPolicy 声明 alarm_type 收到 iot:alarm:stream 中 event_status="end"（持续事件解除）的处理策略。
// EndPolicyIgnore（零值默认）：协议层无 end 配对，end 信号来了直接丢弃记日志。
// EndPolicyAutoResolve：设备侧 end = 行为信号已消失，等价 device-driven recovery，调 HandleRecoveryWithTypes 解除同 (device, event_type) active 行。
// EndPolicyManualAck：临床高风险信号（如长时无体动），end ≠ 病情已评估，必须保持 active 等护理人员人工 ack；end 仅记日志。
// 路由由 AlarmHandler 入口集中处理，case 分支不再各自判断 event_status。
type EndPolicy string

const (
	EndPolicyIgnore      EndPolicy = ""
	EndPolicyAutoResolve EndPolicy = "auto_resolve"
	EndPolicyManualAck   EndPolicy = "manual_ack"
)

// AlarmDef 单条报警/事件元数据，消费方通过 Key 查表获取完整配置（类似 observation.Registry）。
type AlarmDef struct {
	Key          string                 // 与 AlarmType 一致，全局唯一
	ProcessType  string                 // ProcessTypeImmediate | ProcessTypeTimeBased | ...
	DurationSec  int                    // 持续型阈值（秒），0=即时
	UpgradeTo    string                 // 时间型升级目标 alarm_type
	DefaultLevel string                 // 默认报警级别
	AlarmParams  map[string]interface{} // 扩展参数，如 duration_sec、duration_min、min、max
	Description  string                 // 规则说明
	Display      string                 // UI 展示用友好名称（可选，空则用 Key）
	// DedupWhileActive=true：同 (device_id, event_type) 已有 active 行时，
	// InsertAlarmAndUpdateCard 不再插入新行，返回 Deduped=true 由调用方短路通知。
	// 设备类（Offline/SignalPoor/AngleException/SensorDetached/DeviceFailure）开启；
	// 事件型（Fall/SuspectedFall/NightAbsence/Stay/InBed/LeftBed/...）保持 false 各自独立成行。
	DedupWhileActive bool
	// SkipUnhandledCount=true：alarm_events 仍按 'active' 落库（审计），但**不**计入
	// cards.unhandled_alarm_N、**不**写 card.pop_alarm、**不**触发推送通知。
	// 用于设备健康类（Offline/SignalPoor/AngleException/SensorDetached/DeviceFailure 及其 Recover）：
	// UI 上的设备状态由 device:status hash 独立通道驱动（卡片 sleepace-offline / radar-offline 图标），
	// 不污染老人状态 alarm 计数 / 弹窗。AutoResolveDeviceAlarms 也按本字段同步过滤 -count。
	SkipUnhandledCount bool
	// EndPolicy 收到 event_status=end 的策略；零值=Ignore（默认协议无 end 配对）。
	// AlarmHandler 入口集中按本字段路由 end 信号，确保不会被当成新报警插入 alarm_events。
	EndPolicy EndPolicy
}

// Registry 全量报警/事件注册表，Key = alarm_type。与 EventStatToAlarmMap 语义一致，便于按类型查元数据与 AlarmParams。
var Registry = map[string]*AlarmDef{
	Fall:                     {Key: Fall, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelEmerg, Description: "Confirmed fall", Display: "Fall Alert"},
	SuspectedFall:            {Key: SuspectedFall, ProcessType: ProcessTypeTimeBased, DurationSec: 60, UpgradeTo: Fall, DefaultLevel: AlarmLevelAlert, AlarmParams: map[string]interface{}{ParamDurationSec: 60}, Description: "Suspected fall, upgrade after 60s", Display: "Potential Fall"},
	SittingOnGround:          {Key: SittingOnGround, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelAlert, Description: "Sitting on ground", Display: "On Floor Alert"},
	SuspectedSittingOnGround: {Key: SuspectedSittingOnGround, ProcessType: ProcessTypeTimeBased, DurationSec: 60, UpgradeTo: SittingOnGround, DefaultLevel: AlarmLevelWarn, AlarmParams: map[string]interface{}{ParamDurationSec: 60}, Description: "Suspected sitting on ground", Display: "Potential On Floor Alert"},
	BedSitUp:                 {Key: BedSitUp, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelWarn, Description: "Bed sit-up", Display: "Bed Sit-up", EndPolicy: EndPolicyAutoResolve},
	SuspectedBedSitUp:        {Key: SuspectedBedSitUp, ProcessType: ProcessTypeTimeBased, DurationSec: 60, UpgradeTo: BedSitUp, DefaultLevel: AlarmLevelWarn, AlarmParams: map[string]interface{}{ParamDurationSec: 60}, Description: "Suspected bed sit-up", Display: "Potential Bed Sit-up"},
	WarningArea:              {Key: WarningArea, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelAlert, Description: "Warning area", Display: "Warning Area"},
	SignalPoor:               {Key: SignalPoor, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelErr, Description: "Signal poor", Display: "Weak Signal", DedupWhileActive: true, SkipUnhandledCount: true},
	AngleException:           {Key: AngleException, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelErr, Description: "Angle exception", Display: "Device Tilted", DedupWhileActive: true, SkipUnhandledCount: true},
	AlarmTypeOffline:         {Key: AlarmTypeOffline, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelErr, Description: "Offline", Display: "Device Offline", DedupWhileActive: true, SkipUnhandledCount: true},
	AlarmTypeOfflineRecover:  {Key: AlarmTypeOfflineRecover, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Offline recovery", Display: "Device Restored", SkipUnhandledCount: true},
	AlarmTypeDeviceFailure:   {Key: AlarmTypeDeviceFailure, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelErr, Description: "Device failure", Display: "Device Fault", DedupWhileActive: true, SkipUnhandledCount: true},
	AlarmTypeDeviceRecover:   {Key: AlarmTypeDeviceRecover, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Device recovery", Display: "Device Restored", SkipUnhandledCount: true},
	AlarmTypeUnknown:         {Key: AlarmTypeUnknown, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelWarn, Description: "Unknown alarm type"},
	Stay:                     {Key: Stay, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelWarn, AlarmParams: map[string]interface{}{ParamDurationSec: 45 * 60}, Description: "Stay (e.g. 45min)", Display: "Prolonged Stay"},
	NightAbsence:             {Key: NightAbsence, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelAlert, AlarmParams: map[string]interface{}{ParamDurationSec: 24 * 60 * 60}, Description: "Night-long absence from the room (Radar)", Display: "Night Absence"},
	BedNightAbsence:          {Key: BedNightAbsence, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelAlert, AlarmParams: map[string]interface{}{ParamDurationSec: 24 * 60 * 60}, Description: "Night-long absence from the bed (Sleepad)", Display: "Bed Night Absence"},
	WeakBiometricSignal:      {Key: WeakBiometricSignal, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelWarn, Description: "Weak vitals", Display: "Weak Biometric Signal"},
	LeftBed:                  {Key: LeftBed, ProcessType: ProcessTypeTimeBased, DefaultLevel: AlarmLevelWarn, AlarmParams: map[string]interface{}{ParamDurationSec: 0}, Description: "Left bed: duration_sec=0 immediate, >=30min timer", Display: "Bed Exit", EndPolicy: EndPolicyAutoResolve},
	InBed:                    {Key: InBed, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "In bed", Display: "In Bed", EndPolicy: EndPolicyAutoResolve},
	EnterRoom:                {Key: EnterRoom, ProcessType: ProcessTypeRoomStateChange, DefaultLevel: AlarmLevelNotice, Description: "Enter room", Display: "Room Entry"},
	ExitRoom:                 {Key: ExitRoom, ProcessType: ProcessTypeRoomStateChange, DefaultLevel: AlarmLevelNotice, Description: "Exit room", Display: "Room Exit"},
	EnterSensingArea:         {Key: EnterSensingArea, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Enter sensing area", Display: "Entered Sensing Area"},
	ExitSensingArea:          {Key: ExitSensingArea, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Exit sensing area", Display: "Exited Sensing Area"},
	EnterMonitor:             {Key: EnterMonitor, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Enter monitor", Display: "Monitoring Started"},
	ExitMonitor:              {Key: ExitMonitor, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Exit monitor", Display: "Monitoring Ended"},
	HeartRateAlert:           {Key: HeartRateAlert, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelCrit, Description: "Heart rate alert (base; gateway may send High/Low)", Display: "Heart Rate Alert"},
	HeartRateAlertHigh:       {Key: HeartRateAlertHigh, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelCrit, Description: "Heart rate high", Display: "Heart Rate Alert"},
	HeartRateAlertLow:        {Key: HeartRateAlertLow, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelCrit, Description: "Heart rate low", Display: "Heart Rate Alert"},
	HeartRateNormal:          {Key: HeartRateNormal, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Heart rate normal (recovery)", Display: "Heart Rate Normal"},
	RespRateAlert:            {Key: RespRateAlert, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelCrit, Description: "Resp rate alert (base; gateway may send High/Low)", Display: "Resp. Rate Alert"},
	RespRateAlertHigh:        {Key: RespRateAlertHigh, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelCrit, Description: "Resp rate high", Display: "Resp. Rate Alert"},
	RespRateAlertLow:         {Key: RespRateAlertLow, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelCrit, Description: "Resp rate low", Display: "Resp. Rate Alert"},
	RespiratoryRateNormal:    {Key: RespiratoryRateNormal, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Respiratory rate normal (recovery)", Display: "Resp. Rate Normal"},
	ApneaHypopnea:            {Key: ApneaHypopnea, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelCrit, Description: "Apnea/hypopnea", Display: "Apnea/Hypopnea Event"},
	AbnormalBodyMovement:     {Key: AbnormalBodyMovement, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelWarn, Description: "Abnormal body movement", Display: "Excessive Movement", EndPolicy: EndPolicyManualAck},
	NoBodyMove:               {Key: NoBodyMove, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelWarn, Description: "No body move", Display: "No Movement", EndPolicy: EndPolicyManualAck},
	NoTurnOver:               {Key: NoTurnOver, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelWarn, Description: "No turn over", Display: "No Turn-over", EndPolicy: EndPolicyAutoResolve},
	SensorDetached:           {Key: SensorDetached, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelErr, Description: "Sensor detached", Display: "Sensor Detached", DedupWhileActive: true, SkipUnhandledCount: true},
	PressureSensor:           {Key: PressureSensor, ProcessType: ProcessTypeImmediate, DefaultLevel: AlarmLevelNotice, Description: "Pressure sensor status (Sleepad)", Display: "Pressure Sensor"},

	// 数据类别（radar stat / gateway 拆出的 category，仅用于路由；具体报警由子逻辑或 EventStatToAlarmMap 产生）
	Track:        {Key: Track, ProcessType: ProcessTypeDataCategory, Description: "Track data category; pose/enter2out → Fall, Stay, LeftBed, etc."},
	Sleep:        {Key: Sleep, ProcessType: ProcessTypeDataCategory, Description: "Vital data category; threshold → HeartRateAlert*, RespRateAlert*, ApneaHypopnea"},
	SleepStage:   {Key: SleepStage, ProcessType: ProcessTypeDataCategory, Description: "Sleep stage data category"},
	NumberPeople: {Key: NumberPeople, ProcessType: ProcessTypeDataCategory, Description: "Number_people data category; enter/exit room"},
	Activity:     {Key: Activity, ProcessType: ProcessTypeDataCategory, Description: "Activity data category; 老人活动性统计 walk_distance, walk_duration, lie_duration, stand_duration"},
}

// LookupAlarm 按 alarm_type 查 Registry，未找到返回 nil
func LookupAlarm(alarmType string) *AlarmDef {
	return Registry[alarmType]
}

// GetAlarmParams 按 alarmType 返回 Registry 中该类型的 AlarmParams（duration_sec、min、max 等）的副本，未找到或无参数返回 nil。
func GetAlarmParams(alarmType string) map[string]interface{} {
	def := LookupAlarm(alarmType)
	if def == nil || def.AlarmParams == nil {
		return nil
	}
	out := make(map[string]interface{}, len(def.AlarmParams))
	for k, v := range def.AlarmParams {
		out[k] = v
	}
	return out
}

// HeartRate WARNING 区间常量（与 DefaultCloudVitalAlarmThreshold.HeartRate.WARNING.Ranges 一致）
const (
	HRWARNING1Min = 45
	HRWARNING1Max = 54
	HRWARNING2Min = 96
	HRWARNING2Max = 109
)

// IsHRInWARNINGRange 判断心率是否落在 HR WARNING 区间内（45–54、96–109）。
func IsHRInWARNINGRange(hr int) bool {
	return (hr >= HRWARNING1Min && hr <= HRWARNING1Max) ||
		(hr >= HRWARNING2Min && hr <= HRWARNING2Max)
}

// 方向后缀：与 AlarmType 组合使用，如 HeartRateAlert.High
const (
	DirHigh = ".High"
	DirLow  = ".Low"
)

// SplitAlarmQualifier 拆分带方向后缀的报警类型
// "HeartRateAlert.High" → ("HeartRateAlert", "High")
// "Fall" → ("Fall", "")
func SplitAlarmQualifier(qualified string) (baseType, qualifier string) {
	if i := strings.LastIndex(qualified, "."); i > 0 {
		q := qualified[i+1:]
		if q == "High" || q == "Low" {
			return qualified[:i], q
		}
	}
	return qualified, ""
}

// ========== FHIR Category 映射（alarm_events.category 列） ==========
// DB CHECK 约束只允许: "safety", "clinical", "behavioral", "device"

const (
	FHIRCategorySafety     = "safety"
	FHIRCategoryClinical   = "clinical"
	FHIRCategoryBehavioral = "behavioral"
	FHIRCategoryDevice     = "device"
)

// AlarmTypeToFHIRCategory 根据 alarmType 返回 FHIR category
// 新增 alarm type 时在此维护
var AlarmTypeToFHIRCategory = map[string]string{
	// safety: 跌倒、坐地、离床、警告区域等人身安全类
	Fall:                     FHIRCategorySafety,
	SuspectedFall:            FHIRCategorySafety,
	SittingOnGround:          FHIRCategorySafety,
	SuspectedSittingOnGround: FHIRCategorySafety,
	Stay:                     FHIRCategorySafety,
	NightAbsence:             FHIRCategorySafety,
	BedNightAbsence:          FHIRCategorySafety,

	// clinical: 生命体征异常
	HeartRateAlert:       FHIRCategoryClinical,
	HeartRateAlertHigh:   FHIRCategoryClinical,
	HeartRateAlertLow:    FHIRCategoryClinical,
	RespRateAlert:        FHIRCategoryClinical,
	RespRateAlertHigh:    FHIRCategoryClinical,
	RespRateAlertLow:     FHIRCategoryClinical,
	ApneaHypopnea:        FHIRCategoryClinical,
	WeakBiometricSignal:  FHIRCategoryClinical,
	NoBodyMove:           FHIRCategoryClinical, //为老人2H翻身
	AbnormalBodyMovement: FHIRCategoryClinical,

	// behavioral: 行为/作息类
	InBed:             FHIRCategoryBehavioral,
	LeftBed:           FHIRCategoryBehavioral,
	BedSitUp:          FHIRCategoryBehavioral,
	SuspectedBedSitUp: FHIRCategoryBehavioral,
	NoTurnOver:        FHIRCategoryBehavioral,
	WarningArea:       FHIRCategoryBehavioral,

	EnterRoom:        FHIRCategoryBehavioral,
	ExitRoom:         FHIRCategoryBehavioral,
	EnterSensingArea: FHIRCategoryBehavioral,
	ExitSensingArea:  FHIRCategoryBehavioral,
	Initialization:   FHIRCategoryBehavioral,
	Walking:          FHIRCategoryBehavioral,
	Sitting:          FHIRCategoryBehavioral,
	Standing:         FHIRCategoryBehavioral,
	Lying:            FHIRCategoryBehavioral,
	EnterMonitor:     FHIRCategoryBehavioral,
	ExitMonitor:      FHIRCategoryBehavioral,

	// device: 设备自身状态
	AlarmTypeOffline:        FHIRCategoryDevice,
	AlarmTypeOfflineRecover: FHIRCategoryDevice,
	AlarmTypeDeviceFailure:  FHIRCategoryDevice,
	AlarmTypeDeviceRecover:  FHIRCategoryDevice,
	SignalPoor:              FHIRCategoryDevice,
	SingalPoorRecover:       FHIRCategoryDevice,
	AngleException:          FHIRCategoryDevice,
	AngleExceptionRecover:   FHIRCategoryDevice,
	SensorDetached:          FHIRCategoryDevice,
	SensorDetachedRecover:   FHIRCategoryDevice,
	PressureSensor:          FHIRCategoryDevice,
}

// GetFHIRCategory 获取 alarmType 对应的 FHIR category，未匹配返回 "safety"
func GetFHIRCategory(alarmType string) string {
	if cat, ok := AlarmTypeToFHIRCategory[alarmType]; ok {
		return cat
	}
	return FHIRCategorySafety
}

// ========== 配置结构：alarm_type + is_enabled(0/1) + alarm_level + alarm_params + display_setting ==========

const (
	IsEnabledOff = 0 // 关闭报警
	IsEnabledOn  = 1 // 开启
)

const (
	DisplayNone                = 0 // 不显示
	DisplayAlarmCloud          = 1 // alarm_cloud
	DisplayAlarmDevice         = 2 // alarm_device
	DisplayAlarmCloudAndDevice = 3 // alarm_cloud + alarm_device
)

const (
	ParamDurationSec = "duration_sec"
	ParamDurationMin = "duration_min"
	ParamMin         = "min"
	ParamMax         = "max"
	ParamSensitivity = "sensitivity"
)

const (
	MonitoringModePeopleTracking  = 3  // People Tracking
	MonitoringModeFallMonitoring  = 7  // Fall Monitoring
	MonitoringModeSleepMonitoring = 11 // Sleep Monitoring
	MonitoringModeFullFunction    = 15 // Full Function
)

// 报警级别：字符串常量（event_level/alarm_level）+ 优先级数值（0-8，越小越优先，参考 TDPv2-1122.md）
const (
	AlarmLevelEmerg  = "EMERG"
	AlarmLevelAlert  = "ALERT"
	AlarmLevelCrit   = "CRITICAL"
	AlarmLevelErr    = "ERROR"
	AlarmLevelWarn   = "WARNING"
	AlarmLevelNotice = "NOTICE"
	AlarmLevelInfo   = "INFO"
	AlarmLevelDebug  = "DEBUG"
	AlarmLevelCancel = "CANCEL"

	AlarmLevelIntEmerg   = 0
	AlarmLevelIntAlert   = 1
	AlarmLevelIntCrit    = 2
	AlarmLevelIntErr     = 3
	AlarmLevelIntWarning = 4
	AlarmLevelIntNotice  = 5
	AlarmLevelIntInfo    = 6
	AlarmLevelIntDebug   = 7
	AlarmLevelIntCancel  = 8
)

// AlarmLevelPriority 级别字符串→优先级数字；入库与统计统一用全名（CRITICAL/ERROR），见 NormalizeAlarmLevel。
var AlarmLevelPriority = map[string]int{
	AlarmLevelEmerg:  AlarmLevelIntEmerg,
	AlarmLevelAlert:  AlarmLevelIntAlert,
	AlarmLevelCrit:   AlarmLevelIntCrit,
	AlarmLevelErr:    AlarmLevelIntErr,
	AlarmLevelWarn:   AlarmLevelIntWarning,
	AlarmLevelNotice: AlarmLevelIntNotice,
	AlarmLevelInfo:   AlarmLevelIntInfo,
	AlarmLevelDebug:  AlarmLevelIntDebug,
	"CANCEL":         AlarmLevelIntCancel,
}

// NormalizeAlarmLevel 将历史别名 CRIT/ERR 规范为全名 CRITICAL/ERROR（与 AlarmLevelCrit/AlarmLevelErr 一致）。
func NormalizeAlarmLevel(s string) string {
	switch strings.TrimSpace(s) {
	case "CRIT":
		return AlarmLevelCrit
	case "CRITIAL": // 历史拼写错误
		return AlarmLevelCrit
	case "ERR":
		return AlarmLevelErr
	default:
		return s
	}
}

func IsAlarmLevel(s string) bool {
	_, ok := AlarmLevelPriority[NormalizeAlarmLevel(s)]
	return ok
}

// ========== Types (类型) ==========

// CloudVitalAlarmThreshold 生理指标阈值配置
// 存储在 alarm_cloud 表的 conditions JSONB 字段中
// 用于定义生理指标类报警的阈值范围
// 包含完整的 CRITICAL/WARNING/Normal ranges
type CloudVitalAlarmThreshold struct {
	// 完整的 conditions 结构（包含 CRITICAL/WARNING/Normal ranges）
	// Heart Rate: CRITICAL (0-44, 116+), WARNING (45-54, 96-109), Normal (55-95)
	// Respiratory Rate: CRITICAL (0-7, 25+),  Normal (8-24)
	Conditions *VitalAlarmConditions `json:"conditions,omitempty"`
}

// VitalAlarmConditions 完整的生理指标报警条件（包含所有级别的 ranges）
type VitalAlarmConditions struct {
	HeartRate *struct {
		CRITICAL *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"CRITICAL,omitempty"`
		WARNING *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"WARNING,omitempty"`
		Normal *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"Normal,omitempty"`
	} `json:"heart_rate,omitempty"`
	RespiratoryRate *struct {
		CRITICAL *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"CRITICAL,omitempty"`
		// WARNING: 已简化，不包含 WARNING 级别
		Normal *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"Normal,omitempty"`
	} `json:"respiratory_rate,omitempty"`
}

// TenantResetTime 租户级别的作息时间配置
// 存储在 alarm_cloud 表的 metadata JSONB 字段中，或者从 device_alarms 中的 ResetTime/NapTime 项提取
type TenantResetTime struct {
	ResetTime             ResetTimeParams `json:"reset_time,omitempty"`
	NapTime               NapTimeParams   `json:"nap_time,omitempty"`
	TenantSleepReportTime int             `json:"tenant_sleepreport_time,omitempty"` // 1-24 整点，tenant 级睡眠报告生成时刻默认值；设备级可覆盖
}

// DefaultSleepReportTime 代码硬编码的最终 fallback（tenant 无配置、device 无覆盖时用）。
const DefaultSleepReportTime = 8

// ResetTimeParams 重置时间参数
type ResetTimeParams struct {
	InBedTime  string `json:"InBedTime"`  // 格式: "21:30"
	OutBedTime string `json:"OutBedTime"` // 格式: "07:30"
}

// NapTimeParams 午睡时间参数
type NapTimeParams struct {
	InBedTime  string `json:"InBedTime"`  // 格式: "13:00"
	OutBedTime string `json:"OutBedTime"` // 格式: "14:00"
}

// CommonAlarm 通用报警配置（所有设备类型都支持的报警项）
type CommonAlarm struct {
	OfflineAlarm  string `json:"OfflineAlarm,omitempty"`  // 设备离线报警级别
	LowBattery    string `json:"LowBattery,omitempty"`    // 低电量报警级别
	DeviceFailure string `json:"DeviceFailure,omitempty"` // 设备故障报警级别
}

// AlarmCloudConfig 完整的 alarm_cloud 配置
// 所有配置保存在 metadata 中，用于保存到 config_versions 表
type AlarmCloudConfig struct {
	TenantResetTime TenantResetTime `json:"TenantResetTime,omitempty"`
	CommonAlarm     CommonAlarm     `json:"common_alarm,omitempty"`
	AlarmSetting    struct {
		Sleepad []AlarmItem `json:"sleepad,omitempty"`
		Radar   []AlarmItem `json:"radar,omitempty"`
	} `json:"AlarmSetting,omitempty"`
	CloudVitalAlarmThreshold CloudVitalAlarmThreshold `json:"CloudVitalAlarmThreshold,omitempty"`
}

// AlarmItem 单条报警项
// alarm_type: 报警类型 | is_enabled: 0=关闭 1=开启 | alarm_level: nil=无报警级别 | alarm_params: nil=无参数 | display_setting: 显示设置
type AlarmItem struct {
	AlarmType      string                 `json:"alarm_type"`
	IsEnabled      *int                   `json:"is_enabled,omitempty"`   // 0=关闭报警 1=开启
	AlarmLevel     *string                `json:"alarm_level,omitempty"`  // nil=无报警级别
	AlarmParams    map[string]interface{} `json:"alarm_params,omitempty"` // nil=无参数
	DisplaySetting int                    `json:"display_setting"`
}

// AlarmEnablementMapRadar Radar 设备的报警使能配置表：map[alarm_type]0/1 (0=禁用 1=启用)
// 注意：此类型已废弃，建议使用 AlarmEnablementItem
//type AlarmEnablementMapRadar map[string]int

// AlarmEnablementMapSleepad Sleepad 设备的报警使能配置表：map[alarm_type]0/1 (0=禁用 1=启用)
// 注意：此类型已废弃，建议使用 AlarmEnablementItem
//type AlarmEnablementMapSleepad map[string]int

// AlarmEnablementItem 报警使能配置项
// 用于 GetAlarmEnablementMap 函数的返回值
type AlarmEnablementItem struct {
	AlarmType  string // 报警类型
	IsEnabled  int    // 是否启用：0=禁用 1=启用
	AlarmLevel string // 报警级别（不为空）
}

// ========== Variables (变量) ==========

// DefaultCloudVitalAlarmThreshold 默认的生理指标阈值配置
// 包含完整的 conditions 结构，包含所有级别的 ranges
var DefaultCloudVitalAlarmThreshold = CloudVitalAlarmThreshold{
	Conditions: &VitalAlarmConditions{
		HeartRate: &struct {
			CRITICAL *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"CRITICAL,omitempty"`
			WARNING *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"WARNING,omitempty"`
			Normal *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"Normal,omitempty"`
		}{
			CRITICAL: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(0), Max: intPtr(44)},
					{Min: intPtr(116), Max: nil},
				},
				DurationSec: 60,
			},
			WARNING: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(45), Max: intPtr(54)},
					{Min: intPtr(96), Max: intPtr(109)},
				},
				DurationSec: 300,
			},
			Normal: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(55), Max: intPtr(95)},
				},
				DurationSec: 0,
			},
		},
		RespiratoryRate: &struct {
			CRITICAL *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"CRITICAL,omitempty"`
			// WARNING: 已简化，不包含 WARNING 级别
			Normal *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"Normal,omitempty"`
		}{
			CRITICAL: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(0), Max: intPtr(7)},
					{Min: intPtr(25), Max: nil},
				},
				DurationSec: 60,
			},
			Normal: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(8), Max: intPtr(24)},
				},
				DurationSec: 0,
			},
		},
	},
}

// DefaultTenantResetTime 默认的租户作息时间配置
var DefaultTenantResetTime = TenantResetTime{
	ResetTime: ResetTimeParams{
		InBedTime:  "21:30",
		OutBedTime: "07:30",
	},
	NapTime: NapTimeParams{
		InBedTime:  "13:00",
		OutBedTime: "14:00",
	},
	TenantSleepReportTime: DefaultSleepReportTime,
}

// DefaultAlarmSetting 各设备报警项默认值；直接改此变量后引用，json.Marshal(DefaultAlarmSetting) 可得 JSON
// 各报警项默认值（合并顺序：本包 < alarm_cloud < vendor_config < monitor_config）
var DefaultAlarmSetting = struct {
	Sleepad []AlarmItem `json:"sleepad"`
	Radar   []AlarmItem `json:"radar"`
}{
	Sleepad: []AlarmItem{
		{
			AlarmType:  ResetTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "21:30",
				"OutBedTime": "07:30",
				"TenantSleepReportTime": 8,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NapTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "13:00",
				"OutBedTime": "14:00",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepadSetting,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"SensorNumber":         0,  // //  单人Single 0，two 1   (左=0, 右=1)
				"realtime_interval":    2,  //测量间隔2秒/ 4G 30sec  1-254
				"Bed_Exit_Sensitivity": 1,  // 0:Low(15-20s) 1:Medium(5-8s) 2:High(3-5s)
				"report_upload_type":   0,  //0 24H 1=自动结束
				"report_upload_time":   8,  //AM8出报告（统一字段名，与厂家 setReportUploadTime / wisefido-sleepace YAML 一致；雷达 sleep monitor 也复用此名）
				"Empty_Bed_Monitor":    1,  // 0:report 1:no report
				"light_mode":           0,  // 0:开启 1:关闭
				"timezone":             "", // 留空：UI 初次打开时由后端 resp.timezone（= unit.timezone）填充；用户改过后覆盖此字段
			},
			DisplaySetting: DisplayNone, // 设备配置项，不在报警列表中显示
		},
		{
			AlarmType:  MaterialSetting,
			IsEnabled:  intPtr(IsEnabledOff),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"material_type": 1,  //床垫材质 (material): select (1、海绵Foam，2、弹簧Spring 3、乳胶Latex，4、气垫 Air Mattress 5、其他 other)
				"thickness":     2,  //床垫厚度 (thickness): 厚度值1-3(1:5-10cm / 2-4"，2:1-20cm / 4-8"，3:21-30cm / 8-12")
				"bedhight":      60, //床的高度，为以后雷达测量准备 Bed Hight:  60cm/24"
			},
			DisplaySetting: DisplayNone,
		},
		{
			AlarmType:  HeartRateAlert,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelCrit),
			AlarmParams: map[string]interface{}{
				ParamMin:            50,
				ParamMax:            100,
				"slow_duration_sec": 120,
				"fast_duration_sec": 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RespRateAlert,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelCrit),
			AlarmParams: map[string]interface{}{
				ParamMin:            8,
				ParamMax:            24,
				"slow_duration_sec": 120,
				"fast_duration_sec": 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  ApneaHypopnea,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelCrit),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{ // BedNightAbsence(Sleepad): 9PM-7AM 床上无人；实际由早上7点（或 RestTime 结束后）查 IoT 统计得出
			AlarmType:      BedNightAbsence,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{ // RestTime 期间，上床5分钟后，启用离床检测：0立即， 30min,45min,1h,1h30h,2h
			AlarmType:  LeftBed,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 2700,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{ // InBed: 实时在床（如上床、在床满5分钟）。9PM-7AM 床上有没有由 NightAbsence 在 7 点查 IoT 统计判断
			AlarmType:  InBed,
			IsEnabled:  intPtr(IsEnabledOff),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 5,
			},
			DisplaySetting: DisplayNone,
		},
		{
			AlarmType:      BedSitUp,
			IsEnabled:      intPtr(IsEnabledOff),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  AbnormalBodyMovement,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 10, //频繁体动时长和无体动时长需要修改改为实时数据上报间隔的倍数
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NoBodyMove,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 45, //频繁体动时长和无体动时长需要修改改为实时数据上报间隔的倍数
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NoTurnOver,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SensorDetached,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelErr),
			AlarmParams: map[string]interface{}{
				"duration_min": 120,
			},
			DisplaySetting: DisplayAlarmCloud,
		},
	},
	Radar: []AlarmItem{
		{
			AlarmType:  ResetTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "21:30",
				"OutBedTime": "07:30",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NapTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "13:00",
				"OutBedTime": "14:00",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  MonitoringMode,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"mode": 15,
			},
			DisplaySetting: DisplayAlarmDevice,
		},
		{
			AlarmType:  HeartRateAlert,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelCrit),
			AlarmParams: map[string]interface{}{
				ParamMin: 50,
				ParamMax: 95,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RespRateAlert,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelCrit),
			AlarmParams: map[string]interface{}{
				ParamMin: 8,
				ParamMax: 24,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  ApneaHypopnea,
			IsEnabled:  intPtr(IsEnabledOff),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				"apnea_60s_min_events":    2, //15sec无呼呼=呼吸暂停event, ≥2次报警
				"apnea_120sec_min_events": 4, //`120s内呼吸暂停≥4次报警
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  WeakBiometricSignal,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelCrit),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 10,
				ParamSensitivity: 35,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      PostureDetection,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     nil,
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayNone,
		},
		{
			AlarmType:  Fall,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelCrit),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 60,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      SuspectedFall,
			IsEnabled:      intPtr(IsEnabledOff),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayNone,
		},
		{
			AlarmType:  SittingOnGround,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 90,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      BedSitUp,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayNone,
		},
		{
			AlarmType:  LeftBed,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 2700,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      NightAbsence,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  Stay,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      SignalPoor,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelErr),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloud,
		},
		{
			AlarmType:      AngleException,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelErr),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloud,
		},
	},
}

// 注意：设备上报的报警类型到报警级别的映射应该从 alarm_cloud 配置中动态获取，而不是硬编码
// 报警级别应该由 alarm_cloud.device_alarms 中的配置决定，每个租户可以有不同的配置
// 因此不再提供 DeviceAlarmLevelMapSleepad 和 DeviceAlarmLevelMapRadar 硬编码映射表
// 如果需要获取报警级别，应该：
// 1. 从 alarm_cloud.device_alarms[device_type][alarm_type] 获取配置的 alarm_level
// 2. 如果配置不存在，使用 DefaultAlarmSetting 中的默认值
// 3. 将字符串级别的 alarm_level（如 "EMERG", "WARNING"）转换为整数级别（0-8）

// ========== Functions (函数) ==========

// strPtr 返回字符串指针
func strPtr(s string) *string { return &s }

// intPtr 返回整数指针
func intPtr(i int) *int { return &i }

// GetRadarAlarmTypes 获取 Radar 设备的报警类型列表
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
func GetRadarAlarmTypes() []string {
	return []string{AlarmTypeOffline, AlarmTypeDeviceFailure, Fall, SuspectedFall, SittingOnGround, SuspectedSittingOnGround, SuspectedBedSitUp, ApneaHypopnea, HeartRateAlert, RespRateAlert, WeakBiometricSignal, LeftBed, BedSitUp, Stay, NightAbsence, SignalPoor, AngleException, MonitoringMode, NapTime, ResetTime, PostureDetection}
}

// GetSleepPadAlarmTypes 获取 SleepPad 设备的报警类型列表
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
// Sleepad 用 BedNightAbsence（床层面）；Radar 用 NightAbsence（房间层面）
func GetSleepPadAlarmTypes() []string {
	return []string{AlarmTypeOffline, AlarmTypeDeviceFailure, SensorDetached, ApneaHypopnea, HeartRateAlert, RespRateAlert, LeftBed, BedSitUp, InBed, BedNightAbsence, AbnormalBodyMovement, NoBodyMove, NoTurnOver, ResetTime, NapTime, SleepadSetting, MaterialSetting}
}

// GetSupportedAlarmTypes 根据设备类型获取支持的报警类型列表
func GetSupportedAlarmTypes(deviceType string) []string {
	switch deviceType {
	case "radar":
		return GetRadarAlarmTypes()
	case "Sleepad", "sleepad":
		return GetSleepPadAlarmTypes()
	default:
		return []string{AlarmTypeOffline, AlarmTypeDeviceFailure}
	}
}

// IsDisplayableAlarmType 判断报警类型是否可显示
func IsDisplayableAlarmType(alarmType string) bool { return alarmType != AlarmTypeUnknown }

// GetDefaultAlarmItemsSleepPad 获取 SleepPad 设备的默认报警项列表
func GetDefaultAlarmItemsSleepPad() []AlarmItem { return DefaultAlarmSetting.Sleepad }

// GetDefaultAlarmItemsRadar 获取 Radar 设备的默认报警项列表
func GetDefaultAlarmItemsRadar() []AlarmItem { return DefaultAlarmSetting.Radar }

// GetDefaultAlarmLevel 从 DefaultAlarmSetting 查找 alarmType 的默认 alarm_level
// 遍历 Radar + Sleepad 所有默认项，返回首个匹配且 is_enabled=1 的 alarm_level
// 未找到返回 ""
func GetDefaultAlarmLevel(alarmType string) string {
	all := append(DefaultAlarmSetting.Radar, DefaultAlarmSetting.Sleepad...)
	for _, item := range all {
		if item.AlarmType == alarmType && item.IsEnabled != nil && *item.IsEnabled == 1 && item.AlarmLevel != nil {
			return *item.AlarmLevel
		}
	}
	return ""
}

// GetDefaultAlarmItems 根据设备类型获取默认报警项列表
func GetDefaultAlarmItems(deviceType string) []AlarmItem {
	switch deviceType {
	case "Sleepad", "sleepad":
		return GetDefaultAlarmItemsSleepPad()
	case "Radar", "radar":
		return GetDefaultAlarmItemsRadar()
	default:
		return nil
	}
}

// ParseTimeString 解析时间字符串 "HH:MM" 为 hour 和 minute
// 例如: "06:20" -> (6, 20, nil), "21:30" -> (21, 30, nil)
func ParseTimeString(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s, expected HH:MM", timeStr)
	}
	hour, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}
	minute, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}
	return hour, minute, nil
}

// GetAlarmEnablementMap 获取设备的报警使能配置表
// 参数：deviceType - 设备类型（"Sleepad"/"Sleepad"/"sleepad" 或 "radar"）
// 参数：alarmItems - 报警项列表（从 alarm_cloud.device_alarms 获取，如果为 nil 则使用 DefaultAlarmSetting）
// 返回：[]AlarmEnablementItem，只包含 isEnabled=1 且 alarm_level 不为空的项
// 用于过滤决定是否将事件发布到 iot:alarm:stream
func GetAlarmEnablementMap(deviceType string, alarmItems []AlarmItem) []AlarmEnablementItem {
	// 如果 alarmItems 为 nil，使用默认配置
	if alarmItems == nil {
		switch deviceType {
		case "Sleepad", "sleepad":
			alarmItems = DefaultAlarmSetting.Sleepad
		case "radar":
			alarmItems = DefaultAlarmSetting.Radar
		default:
			return []AlarmEnablementItem{}
		}
	}

	result := make([]AlarmEnablementItem, 0)
	for _, item := range alarmItems {
		// 过滤条件：isEnabled=1 且 alarm_level 不为空
		if item.IsEnabled == nil || *item.IsEnabled != IsEnabledOn {
			continue
		}
		if item.AlarmLevel == nil || *item.AlarmLevel == "" {
			continue
		}

		result = append(result, AlarmEnablementItem{
			AlarmType:  item.AlarmType,
			IsEnabled:  *item.IsEnabled,
			AlarmLevel: *item.AlarmLevel,
		})
	}

	return result
}

// MQTTToAlarmTypeMapSleepad Sleepad MQTT 消息到 alarm_type 的映射表
// 基于文档 radar-Qinlan-code-v3.0.md (598-610)
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
var MQTTToAlarmTypeMapSleepad = map[string]string{
	"alarmLeftBed":         LeftBed,
	"alarmHeartRateFast":   HeartRateAlert + DirHigh,
	"alarmHeartRateSlow":   HeartRateAlert + DirLow,
	"alarmBreathRateFast":  RespRateAlert + DirHigh,
	"alarmBreathRateSlow":  RespRateAlert + DirLow,
	"alarmBreathRatePause": ApneaHypopnea,
	"alarmBodymove":        AbnormalBodyMovement,
	"alarmNoBodymove":      NoBodyMove,
	"alarmNoTurnOver":      NoTurnOver,
	"alarmBedSitup":        BedSitUp,
	"alarmSitup":           BedSitUp, // 疑似坐起，与 alarmBedSitup 同义
	"alarmInBed":           InBed,
	"alarmOnBed":           InBed, // 在床报警，与 alarmInBed 同义
	"alarmSensorFall":      SensorDetached,
	"offLine":              AlarmTypeOffline,
}

// ========== Radar Event numeric_code → AlarmType 映射（event/alarm topic） ==========
// 一条 MQTT event 可包含多个 track，每个 track 生成一个 numeric_code，查此表获取 AlarmType。
// 多 track 时取 Priority 最小（最高优先级）的作为顶层 category。

// ========== Radar Event 映射表（event/alarm topic） ==========
// 直接使用协议结构化字段（protocol.go）查表，不拼字符串。
// 映射表只负责 protocol → AlarmType，是否报警由业务层（enablement 配置）决定。

// Enter2OutMap type=1 进出事件：[event][area_type] → AlarmType
// 报警映射独立于 protocol.go，直接使用协议数值。
// event: 1=进房 2=离房 3=进区域 4=离区域 5=进入监护 6=退出监护
// area_type: 2=床 4=门 5=监护床 6=感应区
var Enter2OutMap = map[int]map[int]string{
	1: { // EventInRoom
		4: EnterRoom, // AreaTypeDoor
	},
	2: { // EventOutRoom
		4: ExitRoom, // AreaTypeDoor
	},
	3: { // EventInArea
		2: InBed, // AreaTypeBed
		5: InBed,
		6: EnterSensingArea, // AreaTypeSensing
	},
	4: { // EventOutArea
		2: LeftBed, // AreaTypeBed
		5: LeftBed,
		6: ExitSensingArea, // AreaTypeSensing
	},
	5: { // EventEnterMonitor
		5: EnterMonitor, // AreaTypeMonitorBed
	},
	6: { // EventExitMonitor
		5: ExitMonitor, // AreaTypeMonitorBed
	},
}

// PoseMap type=2 姿态变化：[pose] → AlarmType
// 报警映射独立于 protocol.go 的 PoseNumToDisplay（显示用途），
// 此处根据业务语义映射：厂家 pose=2(SuspectedFall) 在 event topic 发出时
// 已经过固件内部等待，应直接映射为 Fall alarm。
var PoseMap = map[int]string{
	0:  Initialization,
	1:  Walking,
	2:  Fall, // 厂家 SuspectedFall → 我方 Fall
	3:  Sitting,
	4:  Standing,
	5:  Fall,
	6:  Lying,
	7:  SittingOnGround,
	8:  SittingOnGround,
	9:  BedSitUp,
	10: BedSitUp, // BedSitUpConfirm → BedSitUp
	11: BedSitUp,
}

// DeviceEventMap type=5/7/8 设备状态：[eventType] → AlarmType
var DeviceEventMap = map[int]string{
	5: AlarmTypeOffline, // 设备离线
	7: SignalPoor,       // 信号差
	8: AngleException,   // 倾角异常
}

// ========== Radar Stat numeric_code → AlarmType 映射（stat topic） ==========
// stat 每分钟 1 条（sleep + track），无多 track 问题，用简单 string→string。
// key 格式: "sleep_" + type + "_" + state（2-bit 二进制），如 "sleep_breath_01"

// StatSleepMap stat sleep base64 解码后的 numeric_code → AlarmType
// key 格式: "sleep_{type}_{state}"，state 为 byte13 的 2-bit 二进制
var StatSleepMap = map[string]string{
	// byte13 bit[1:0] 呼吸: 00=正常 01=过低 10=过高 11=暂停
	"sleep_breath_00": RespiratoryRateNormal,
	"sleep_breath_01": RespRateAlert + DirLow,
	"sleep_breath_10": RespRateAlert + DirHigh,
	"sleep_breath_11": ApneaHypopnea,
	// byte13 bit[3:2] 心率: 00=正常 01=过低 10=过高 11=未定义
	"sleep_heart_00": HeartRateNormal,
	"sleep_heart_01": HeartRateAlert + DirLow,
	"sleep_heart_10": HeartRateAlert + DirHigh,
	// byte13 bit[5:4] 生命体征: 00=正常 01=未定义 10=未定义 11=弱
	"sleep_vitals_11": WeakBiometricSignal,
}

// StatTrackMap stat track base64 解码后的行为统计 → AlarmType
// key 格式: "track_{type}"
var StatTrackMap = map[string]string{
	"track_stay":        Stay,
	"track_no_activity": NightAbsence,
}

// StatCodeMap 合并 StatSleepMap + StatTrackMap（兼容旧接口）
var StatCodeMap = func() map[string]string {
	m := make(map[string]string, len(StatSleepMap)+len(StatTrackMap))
	for k, v := range StatSleepMap {
		m[k] = v
	}
	for k, v := range StatTrackMap {
		m[k] = v
	}
	return m
}()

// ========== 结构化查表函数 ==========

// LookupEnter2Out type=1 查表：(event, areaType) → AlarmType
func LookupEnter2Out(event, areaType int) (string, bool) {
	if areas, ok := Enter2OutMap[event]; ok {
		if alarmType, ok := areas[areaType]; ok {
			return alarmType, true
		}
	}
	return "", false
}

// LookupPose type=2 查表：pose → AlarmType
func LookupPose(pose int) (string, bool) {
	alarmType, ok := PoseMap[pose]
	return alarmType, ok
}

// LookupDeviceEvent type=5/7/8 查表：eventType → AlarmType
func LookupDeviceEvent(eventType int) (string, bool) {
	alarmType, ok := DeviceEventMap[eventType]
	return alarmType, ok
}

// MQTTToAlarmTypeMapRadar 兼容旧接口：从三个结构化 map 生成 string→string
// 供 numericCodeToAlarmTypeMap init() 使用，后续逐步迁移到直接使用 Lookup 函数
var MQTTToAlarmTypeMapRadar = func() map[string]string {
	m := make(map[string]string)
	for event, areas := range Enter2OutMap {
		for area, alarmType := range areas {
			m[fmt.Sprintf("1%d%d", event, area)] = alarmType
		}
	}
	for pose, alarmType := range PoseMap {
		m[fmt.Sprintf("2%d", pose)] = alarmType
	}
	for eventType, alarmType := range DeviceEventMap {
		m[fmt.Sprintf("%d", eventType)] = alarmType
	}
	for code, alarmType := range StatCodeMap {
		m[code] = alarmType
	}
	return m
}()

// ConvertMQTTToAlarmTypeSleepad 将 Sleepad MQTT 消息转换为 alarm_type
// 参数：mqttKey - MQTT 消息中的键（如 "alarmLeftBed", "alarmHeartRateFast" 等）
// 返回：对应的 alarm_type，如果未找到则返回空字符串
func ConvertMQTTToAlarmTypeSleepad(mqttKey string) string {
	if alarmType, ok := MQTTToAlarmTypeMapSleepad[mqttKey]; ok {
		return alarmType
	}
	return ""
}

// numericCodeToAlarmTypeMap 从 MQTTToAlarmTypeMapRadar 自动生成的反向映射
// key=numeric code, value=去掉方向后缀的 base alarm type 列表
var numericCodeToAlarmTypeMap map[string][]string

func init() {
	numericCodeToAlarmTypeMap = make(map[string][]string)
	for code, qualified := range MQTTToAlarmTypeMapRadar {
		baseType, _ := SplitAlarmQualifier(qualified)
		if baseType == "" {
			continue
		}
		existing := numericCodeToAlarmTypeMap[code]
		found := false
		for _, t := range existing {
			if t == baseType {
				found = true
				break
			}
		}
		if !found {
			numericCodeToAlarmTypeMap[code] = append(existing, baseType)
		}
	}
}

// LookupStatCode 查询 StatCodeMap，返回 AlarmType 和是否命中
func LookupStatCode(numericCode string) (string, bool) {
	alarmType, ok := StatCodeMap[numericCode]
	return alarmType, ok
}

// GetAlarmTypesFromNumericCode 根据数字组合获取对应的报警类型列表
func GetAlarmTypesFromNumericCode(numericCode string) []string {
	if types, ok := numericCodeToAlarmTypeMap[numericCode]; ok {
		return types
	}
	return nil
}

// CheckAlarmEnabledByNumericCode 根据数字组合检查报警是否启用
// 参数：numericCode - 数字组合字符串（如"142"）
// 参数：enablementMap - 报警使能配置表
// 返回：是否启用（true=任一匹配的报警类型启用，false=全部未启用或未找到）
// 注意：如果数字组合对应多个报警类型，任一启用即返回true
func CheckAlarmEnabledByNumericCode(numericCode string, enablementMap []AlarmEnablementItem) bool {
	alarmTypes := GetAlarmTypesFromNumericCode(numericCode)
	if len(alarmTypes) == 0 {
		return false
	}
	// 检查任一报警类型是否启用
	for _, alarmType := range alarmTypes {
		for _, item := range enablementMap {
			if item.AlarmType == alarmType && item.IsEnabled == 1 {
				return true
			}
		}
	}
	return false
}

// GetAlarmEnabledMapByNumericCodes 根据数字组合列表获取启用状态映射表
// 参数：numericCodes - 数字组合列表（如["142", "145", "25"]）
// 参数：enablementMap - 报警使能配置表
// 返回：数字组合 -> 启用状态的映射表，例如：{"142": 0, "145": 1, "25": 0}
func GetAlarmEnabledMapByNumericCodes(numericCodes []string, enablementMap []AlarmEnablementItem) map[string]int {
	result := make(map[string]int)
	for _, code := range numericCodes {
		enabled := CheckAlarmEnabledByNumericCode(code, enablementMap)
		if enabled {
			result[code] = 1
		} else {
			result[code] = 0
		}
	}
	return result
}

// ExtractNumericCodesFromEvent 从 event 数据中提取数字组合
// 参数：eventData - 解码后的事件数据（map[string]interface{}）
// 返回：数字组合列表，例如：["142", "25", "7"]
func ExtractNumericCodesFromEvent(eventData map[string]interface{}) []string {
	var codes []string

	// 获取 event_type
	eventType, _ := eventData["event_type"].(string)
	if eventType == "" {
		// 如果没有 event_type，尝试从 type 字段获取
		if typeInt, ok := eventData["type"].(int); ok {
			eventType = fmt.Sprintf("%d", typeInt)
		}
	}

	// Event type=1 (进出事件)
	if eventType == "1" {
		// 获取 area_type
		areaType, _ := eventData["area_type_raw"].(string)
		if areaType == "" {
			// 如果没有 area_type_raw，尝试从 area_type 获取（可能是 int）
			if areaTypeInt, ok := eventData["area_type"].(int); ok {
				areaType = fmt.Sprintf("%d", areaTypeInt)
			}
		}

		// 获取 event 的原始值（优先从 event_raw 获取）
		var eventValue string
		if eventRaw, ok := eventData["event_raw"].(string); ok {
			eventValue = eventRaw
		} else if eventRawInt, ok := eventData["event_raw"].(int); ok {
			eventValue = fmt.Sprintf("%d", eventRawInt)
		} else if eventDisplay, ok := eventData["event"].(string); ok {
			// 如果没有 event_raw，从 display_en 反向查找原始值（与 radar_convert_table 全小写一致，不区分大小写）
			// 5=进入监护模式(EnterMonitor) 6=退出监护模式(ExitMonitor)
			switch strings.ToLower(strings.TrimSpace(eventDisplay)) {
			case "enter room":
				eventValue = "1"
			case "leave room":
				eventValue = "2"
			case "enter area":
				eventValue = "3"
			case "leave area":
				eventValue = "4"
			case "enter monitor":
				eventValue = "5"
			case "exit monitor":
				eventValue = "6"
			}
		}

		// 5/6 为监护模式，缺省 area_type 用 5（监护床）以便生成 code
		if (eventValue == "5" || eventValue == "6") && areaType == "" {
			areaType = "5"
		}
		// 生成数字组合：eventType + event + areaType
		if eventValue != "" && areaType != "" {
			code := fmt.Sprintf("%s%s%s", eventType, eventValue, areaType)
			codes = append(codes, code)
		} else if areaType == "6" {
			// 进入警告区域：eventType=1, event=1, area_type=6 -> "116"
			codes = append(codes, "116")
		}
	}

	// Event type=2 (姿态变化)
	if eventType == "2" {
		// 获取 pose 的原始值（decoder 应设置 pose_raw；JSON 解析后可能为 float64）
		pose, _ := eventData["pose_raw"].(string)
		if pose == "" {
			if poseInt, ok := eventData["pose"].(int); ok {
				pose = fmt.Sprintf("%d", poseInt)
			} else if poseFloat, ok := eventData["pose"].(float64); ok {
				pose = fmt.Sprintf("%.0f", poseFloat)
			}
		}
		if pose != "" {
			// 生成数字组合：eventType + pose -> "25", "210" 等
			code := fmt.Sprintf("%s%s", eventType, pose)
			codes = append(codes, code)
		}
	}

	// Event type=5 (离线/在线)
	if eventType == "5" {
		codes = append(codes, "5")
	}

	// Event type=7 (信号差)
	if eventType == "7" {
		codes = append(codes, "7")
	}

	// Event type=8 (倾角异常)
	if eventType == "8" {
		codes = append(codes, "8")
	}

	return codes
}

// ExtractNumericCodesFromStat 从 stat 数据中提取数字组合
// 参数：statData - 解码后的统计数据（map[string]interface{}）
// 返回：数字组合列表，例如：["sleep_breath_01", "sleep_heart_10", "sleep_vitals_11"]
func ExtractNumericCodesFromStat(statData map[string]interface{}) []string {
	var codes []string

	// 检查是否有 stat_numeric_codes 字段（decoder 已生成）
	if numericCodes, ok := statData["stat_numeric_codes"].([]interface{}); ok {
		for _, code := range numericCodes {
			if codeStr, ok := code.(string); ok {
				codes = append(codes, codeStr)
			}
		}
		return codes
	}

	// 如果没有 stat_numeric_codes，从各个字段提取（dataCategory 优先，兼容 category）
	var category string
	if c, _ := statData["dataCategory"].(string); c != "" {
		category = c
	} else if c, _ := statData["category"].(string); c != "" {
		category = c
	}
	if category == "sleep" {
		// 检查是否有 stat_numeric_code_breath
		if breathCode, ok := statData["stat_numeric_code_breath"].(string); ok && breathCode != "" {
			codes = append(codes, breathCode)
		}
		// 检查是否有 stat_numeric_code_heart
		if heartCode, ok := statData["stat_numeric_code_heart"].(string); ok && heartCode != "" {
			codes = append(codes, heartCode)
		}
		// 检查是否有 stat_numeric_code_vitals
		if vitalsCode, ok := statData["stat_numeric_code_vitals"].(string); ok && vitalsCode != "" {
			codes = append(codes, vitalsCode)
		}
	}

	return codes
}

// ShouldConvertToAlarm 根据使能表决定是否将 event/stat 转换为 alarm
// 参数：dataValue - 解码后的数据（event 或 stat 的单个对象）
// 参数：topicType - 主题类型（"event" 或 "stat"）
// 参数：enablementMap - 报警使能配置表
// 返回：是否应该转换为 alarm（true=转换为 alarm，false=保持原 topic）
func ShouldConvertToAlarm(dataValue map[string]interface{}, topicType string, enablementMap []AlarmEnablementItem) bool {
	var numericCodes []string

	// 根据 topicType 提取数字组合
	if topicType == "event" {
		numericCodes = ExtractNumericCodesFromEvent(dataValue)
	} else if topicType == "stat" {
		numericCodes = ExtractNumericCodesFromStat(dataValue)
	}

	if len(numericCodes) == 0 {
		return false
	}

	// 生成使能表
	enabledMap := GetAlarmEnabledMapByNumericCodes(numericCodes, enablementMap)

	// 如果任一数字组合对应的报警启用，则转换为 alarm
	for _, enabled := range enabledMap {
		if enabled == 1 {
			return true
		}
	}

	return false
}

// ========== Event/Stat 到 Alarm_Type 映射表（多厂家共用） ==========

// EventStatToAlarmMapping Event/Stat 到 Alarm_Type 的映射规则
// 用于 card-aggregator 处理 event/stat 到 alarm_type 的映射（Radar 等厂家共用）
type EventStatToAlarmMapping struct {
	// 匹配条件
	TopicType string  // "event" | "stat" | "monitor" | "alarm"
	Category  string  // category 字段值（如 "enter2out", "track", "sleep", "Fall", "SuspectedFall"）
	EventType *int    // event_type 字段值（1=进出, 2=姿态, 7=信号差, 8=倾角）
	AreaType  *int    // area_type 字段值（2=普通床, 5=监护床, 6=警告区）
	Pose      *int    // pose 字段值（2=疑似跌倒, 5=确认跌倒, 7=疑似坐地, 8=确认坐地等）
	StatType  *string // stat 类型（"sleep"）
	StatCode  *string // stat 代码（"breath_01", "heart_10", "stay", "no_activity" 等）

	// 映射结果
	AlarmType   string // 对应的 alarm_type（可为空，如床位/房间状态变化仅更新状态不产生报警）
	ProcessType string // ProcessTypeImmediate | ProcessTypeTimeBased | ...

	// 时间相关参数（仅 ProcessType=ProcessTypeTimeBased 时使用）
	DurationSec *int    // 持续时间阈值（秒），如 60 表示持续60秒后触发
	UpgradeTo   *string // 升级目标 alarm_type（如 SuspectedFall 持续60秒升级为 Fall）

	// AlarmParams 扩展参数，与 Registry[AlarmType].AlarmParams 一致或本条规则覆盖（如 duration_sec、min、max）
	AlarmParams map[string]interface{}

	Description string // 规则说明，便于维护与文档
}

// ========== 完整的 Radar Event/Stat 到 Alarm 映射表 ==========
/*
TopicType — MQTT topic 分类："event"/"stat"/"monitor"/"alarm"
Category — 消息的 category 字段，如 "sleep"、"vital"
EventType — event_type 数值（1=进出、2=姿态）
AreaType — 区域类型（2=普通床、6=警告区）
Pose — 姿态值（5=跌倒）
StatType — stat/monitor 的数据子类型，如 "heart_rate"、"sleep"。用来区分同一个 Category="vital" 下的不同指标，匹配时和传入消息的 stat_type 字段比对
StatCode — stat 具体编码（"breath_01"）
映射结果（匹配后怎么处理）：
AlarmType — 触发哪个报警常量
ProcessType — Immediate（立即触发）还是 TimeBased（持续超时才触发）
DurationSec — TimeBased 时的持续时间阈值（60 = 异常持续60秒才报）
UpgradeTo — 升级目标（如疑似跌倒持续60秒升级为确认跌倒）
Description — 纯注释性质，给开发者看的。UI 展示用 Registry[AlarmType].Display。
*/

var EventStatToAlarmMap = []EventStatToAlarmMapping{
	// ========== 即时触发型报警（已在 qinglan 转换为 alarm）==========
	{
		TopicType:   "alarm",
		Category:    "Fall",
		AlarmType:   Fall,
		ProcessType: ProcessTypeImmediate,
		Description: "Confirmed fall alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "SuspectedFall",
		AlarmType:   SuspectedFall,
		ProcessType: ProcessTypeImmediate,
		Description: "Suspected fall alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "SittingOnGround",
		AlarmType:   SittingOnGround,
		ProcessType: ProcessTypeImmediate,
		Description: "Confirmed sitting on ground alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "SuspectedSittingOnGround",
		AlarmType:   SuspectedSittingOnGround,
		ProcessType: ProcessTypeImmediate,
		Description: "Suspected sitting on ground alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "BedSitUp",
		AlarmType:   BedSitUp,
		ProcessType: ProcessTypeImmediate,
		Description: "Confirmed bed sit-up alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "SuspectedBedSitUp",
		AlarmType:   SuspectedBedSitUp,
		ProcessType: ProcessTypeImmediate,
		Description: "Suspected bed sit-up alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "WarningArea",
		AlarmType:   WarningArea,
		ProcessType: ProcessTypeImmediate,
		Description: "Enter warning area alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "SignalPoor",
		AlarmType:   SignalPoor,
		ProcessType: ProcessTypeImmediate,
		Description: "Radar signal poor alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "AngleException",
		AlarmType:   AngleException,
		ProcessType: ProcessTypeImmediate,
		Description: "Radar angle exception alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "Offline",
		AlarmType:   AlarmTypeOffline,
		ProcessType: ProcessTypeImmediate,
		Description: "Device offline alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "Stay",
		AlarmType:   Stay,
		ProcessType: ProcessTypeImmediate,
		Description: "Stay alarm (evaluated in qinglan)",
	},
	{
		TopicType:   "alarm",
		Category:    "NightAbsence",
		AlarmType:   NightAbsence,
		ProcessType: ProcessTypeImmediate,
		Description: "24h no activity alarm (evaluated in qinglan)",
	},
	{
		TopicType:   "alarm",
		Category:    "WeakBiometricSignal",
		AlarmType:   WeakBiometricSignal,
		ProcessType: ProcessTypeImmediate,
		Description: "Weak vitals alarm",
	},
	{
		TopicType:   "alarm",
		Category:    "LeftBed",
		AlarmType:   LeftBed,
		ProcessType: ProcessTypeTimeBased,
		Description: "Left bed alarm: duration_sec=0 immediate, >=30min timer",
	},
	{
		TopicType:   "alarm",
		Category:    "InBed",
		AlarmType:   InBed,
		ProcessType: ProcessTypeImmediate,
		Description: "In bed alarm (evaluated in qinglan)",
	},

	// ========== 设备在线状态事件 ==========
	{
		TopicType:   "alarm",
		Category:    "OfflineRecover",
		AlarmType:   AlarmTypeOfflineRecover,
		ProcessType: ProcessTypeStateBased,
		Description: "Device recovered online, update device state",
	},

	// ========== Event Stream 中的事件 ==========
	// Event type=1 (进出事件)
	{
		TopicType:   "event",
		Category:    "enter2out",
		EventType:   intPtr(1),
		AreaType:    intPtr(6), // 警告区域
		AlarmType:   WarningArea,
		ProcessType: ProcessTypeImmediate,
		Description: "Enter warning area event",
	},
	{
		TopicType:   "event",
		Category:    "enter2out",
		EventType:   intPtr(1),
		AreaType:    intPtr(2), // 普通床
		ProcessType: ProcessTypeBedStateChange,
		AlarmType:   "",
		Description: "Bed area enter/exit, update bed state",
	},
	{
		TopicType:   "event",
		Category:    "enter2out",
		EventType:   intPtr(1),
		AreaType:    intPtr(5), // 监护床
		ProcessType: ProcessTypeBedStateChange,
		AlarmType:   "",
		Description: "Monitor bed enter/exit, update bed state",
	},
	{
		TopicType:   "event",
		Category:    "enter2out",
		EventType:   intPtr(1),
		AreaType:    intPtr(4), // 房间
		ProcessType: ProcessTypeRoomStateChange,
		AlarmType:   "",
		Description: "Room enter/exit, update room state",
	},

	// Event type=2 (姿态变化) - 时间相关处理
	{
		TopicType:   "event",
		Category:    "pose",
		EventType:   intPtr(2),
		Pose:        intPtr(2), // 疑似跌倒
		AlarmType:   SuspectedFall,
		ProcessType: ProcessTypeTimeBased,
		DurationSec: intPtr(60),
		UpgradeTo:   strPtr(Fall),
		AlarmParams: map[string]interface{}{ParamDurationSec: 60},
		Description: "Suspected fall, upgrade to confirmed fall after 60s",
	},
	{
		TopicType:   "event",
		Category:    "pose",
		EventType:   intPtr(2),
		Pose:        intPtr(7), // 疑似坐地
		AlarmType:   SuspectedSittingOnGround,
		ProcessType: ProcessTypeTimeBased,
		DurationSec: intPtr(60),
		UpgradeTo:   strPtr(SittingOnGround),
		AlarmParams: map[string]interface{}{ParamDurationSec: 60},
		Description: "Suspected sitting on ground, upgrade after 60s",
	},
	{
		TopicType:   "event",
		Category:    "pose",
		EventType:   intPtr(2),
		Pose:        intPtr(10), // 疑似床上坐起
		AlarmType:   SuspectedBedSitUp,
		ProcessType: ProcessTypeTimeBased,
		DurationSec: intPtr(60),
		UpgradeTo:   strPtr(BedSitUp),
		AlarmParams: map[string]interface{}{ParamDurationSec: 60},
		Description: "Suspected bed sit-up, upgrade after 60s",
	},
	// Event type=2 (姿态变化) - 即时触发
	{
		TopicType:   "event",
		Category:    "pose",
		EventType:   intPtr(2),
		Pose:        intPtr(5), // 确认跌倒
		AlarmType:   Fall,
		ProcessType: ProcessTypeImmediate,
		Description: "Confirmed fall event",
	},
	{
		TopicType:   "event",
		Category:    "pose",
		EventType:   intPtr(2),
		Pose:        intPtr(8), // 确认坐地
		AlarmType:   SittingOnGround,
		ProcessType: ProcessTypeImmediate,
		Description: "Confirmed sitting on ground event",
	},
	{
		TopicType:   "event",
		Category:    "pose",
		EventType:   intPtr(2),
		Pose:        intPtr(11), // 确认床上坐起
		AlarmType:   BedSitUp,
		ProcessType: ProcessTypeImmediate,
		Description: "Confirmed bed sit-up event",
	},

	// Event type=3 (number-people) - 人数变化
	{
		TopicType:   "event",
		Category:    "number-people",
		EventType:   intPtr(3),
		ProcessType: ProcessTypeActivityMonitoring,
		AlarmType:   "",
		Description: "Number of people change, for activity monitoring",
	},

	// Event type=7 (信号差)
	{
		TopicType:   "event",
		Category:    "signal",
		EventType:   intPtr(7),
		AlarmType:   SignalPoor,
		ProcessType: ProcessTypeImmediate,
		Description: "Signal poor event",
	},

	// Event type=8 (倾角异常)
	{
		TopicType:   "event",
		Category:    "angle",
		EventType:   intPtr(8),
		AlarmType:   AngleException,
		ProcessType: ProcessTypeImmediate,
		Description: "Angle exception event",
	},

	// ========== Stat Stream 中的统计 ==========
	// 睡眠统计相关
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("breath_01"),
		AlarmType:   RespRateAlert,
		ProcessType: ProcessTypeImmediate,
		Description: "Low respiratory rate stat",
	},
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("breath_10"),
		AlarmType:   RespRateAlert,
		ProcessType: ProcessTypeImmediate,
		Description: "High respiratory rate stat",
	},
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("breath_11"),
		AlarmType:   ApneaHypopnea,
		ProcessType: ProcessTypeImmediate,
		Description: "Apnea/hypopnea stat",
	},
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("heart_01"),
		AlarmType:   HeartRateAlert,
		ProcessType: ProcessTypeImmediate,
		Description: "Low heart rate stat",
	},
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("heart_10"),
		AlarmType:   HeartRateAlert,
		ProcessType: ProcessTypeImmediate,
		Description: "High heart rate stat",
	},
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("vitals_11"),
		AlarmType:   WeakBiometricSignal,
		ProcessType: ProcessTypeImmediate,
		Description: "Weak vitals stat",
	},
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("stay"),
		AlarmType:   Stay,
		ProcessType: ProcessTypeTimeBased,
		DurationSec: intPtr(45 * 60), // 45分钟
		AlarmParams: map[string]interface{}{ParamDurationSec: 45 * 60},
		Description: "Stay stat, trigger after 45min",
	},
	{
		TopicType:   "stat",
		Category:    "sleep",
		StatType:    strPtr("sleep"),
		StatCode:    strPtr("no_activity"),
		AlarmType:   NightAbsence,
		ProcessType: ProcessTypeTimeBased,
		DurationSec: intPtr(24 * 60 * 60), // 24小时
		AlarmParams: map[string]interface{}{ParamDurationSec: 24 * 60 * 60},
		Description: "No activity stat, trigger after 24h",
	},

	// ========== Monitor Stream 中的监控数据 ==========
	{
		TopicType:   "monitor",
		Category:    "vital",
		StatType:    strPtr("respiratory_rate"),
		AlarmType:   RespRateAlert,
		ProcessType: ProcessTypeTimeBased,
		DurationSec: intPtr(60),
		AlarmParams: map[string]interface{}{ParamDurationSec: 60},
		Description: "Respiratory rate monitor",
	},
	{
		TopicType:   "monitor",
		Category:    "vital",
		StatType:    strPtr("heart_rate"),
		AlarmType:   HeartRateAlert,
		ProcessType: ProcessTypeTimeBased,
		DurationSec: intPtr(60),
		AlarmParams: map[string]interface{}{ParamDurationSec: 60},
		Description: "Heart rate monitor",
	},
}

// MatchRadarEventStatToAlarm 匹配 Radar Event/Stat 到 Alarm_Type
// 参数：
//   - topicType: "event" | "stat" | "monitor" | "alarm"
//   - category: category 字段值
//   - eventType: event_type 字段值（可为 nil）
//   - areaType: area_type 字段值（可为 nil）
//   - pose: pose 字段值（可为 nil）
//   - statType: stat 类型（可为 nil）
//   - statCode: stat 代码（可为 nil）
//
// 返回：匹配的映射规则列表（可能有多个匹配，按优先级返回第一个）
func MatchRadarEventStatToAlarm(
	topicType, category string,
	eventType, areaType, pose *int,
	statType, statCode *string,
) []EventStatToAlarmMapping {
	var matches []EventStatToAlarmMapping

	for _, mapping := range EventStatToAlarmMap {
		// 匹配 topic_type
		if mapping.TopicType != "" && mapping.TopicType != topicType {
			continue
		}
		// 匹配 category
		if mapping.Category != "" && mapping.Category != category {
			continue
		}
		// 匹配 event_type
		if mapping.EventType != nil && (eventType == nil || *mapping.EventType != *eventType) {
			continue
		}
		// 匹配 area_type
		if mapping.AreaType != nil && (areaType == nil || *mapping.AreaType != *areaType) {
			continue
		}
		// 匹配 pose
		if mapping.Pose != nil && (pose == nil || *mapping.Pose != *pose) {
			continue
		}
		// 匹配 stat_type
		if mapping.StatType != nil && (statType == nil || *mapping.StatType != *statType) {
			continue
		}
		// 匹配 stat_code
		if mapping.StatCode != nil && (statCode == nil || *mapping.StatCode != *statCode) {
			continue
		}

		matches = append(matches, mapping)
	}

	return matches
}

// GetAlarmTypeFromRadarEventStat 根据 Event/Stat 获取对应的 Alarm_Type（Radar 等厂家共用）
// 返回第一个匹配的 alarm_type，如果未找到则返回空字符串
func GetAlarmTypeFromRadarEventStat(
	topicType, category string,
	eventType, areaType, pose *int,
	statType, statCode *string,
) string {
	matches := MatchRadarEventStatToAlarm(topicType, category, eventType, areaType, pose, statType, statCode)
	if len(matches) > 0 {
		return matches[0].AlarmType
	}
	return ""
}

// ========== UI 层报警归类 ==========
// 将细分报警类型归类到 UI 展示组，便于前端分组显示
// 未在此映射中的类型，UI 组名等于自身

var AlarmTypeUIGroup = map[string]string{
	SuspectedFall:            Fall,
	SuspectedSittingOnGround: SittingOnGround,
	SuspectedBedSitUp:        BedSitUp,
	HeartRateAlert + DirHigh: HeartRateAlert,
	HeartRateAlert + DirLow:  HeartRateAlert,
	RespRateAlert + DirHigh:  RespRateAlert,
	RespRateAlert + DirLow:   RespRateAlert,
}

func GetAlarmUIGroup(alarmType string) string {
	if group, ok := AlarmTypeUIGroup[alarmType]; ok {
		return group
	}
	if base, q := SplitAlarmQualifier(alarmType); q != "" {
		return base
	}
	return alarmType
}

// GetAlarmDisplayName 返回 alarm_type 的 UI 展示名，无则返回 alarmType 自身
func GetAlarmDisplayName(alarmType string) string {
	if def := LookupAlarm(alarmType); def != nil && def.Display != "" {
		return def.Display
	}
	return alarmType
}

/*
心率呼吸率模式获取（BM8701-2专用
http(s)://domain{:port}/sleepace/heartModeGet
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
mode	int	0：模式0(默认)，心率/呼吸率 计算不出时，保留之前的值不变或者加一定的波动
1：模式，心率/呼吸率计算不出，上报无效值 255

状态码	描述
-1	未登录或未认证
0	成功
1	设备离线
2	服务器错误
3	超时
4	设备未绑定
5	设备不存在
8	操作命令不存在
9	参数错误
10	device not belong your channel
该设备不在你的渠道下
11	认证错误


*/
