package service

import "strings"

// ============================================================================
// Weights — derive 与 affinity 所用权重/常量集中在此，便于查找和调参
//
// 1. Derive 周期与 tick（素数散列，避免多卡与 1s 周期谐波）
// 2. 设备信任权重（derive 时多设备竞选）
//    Pose: Sleepad(8) > Radar(6)，Sleepad 仅 InBed 有效
//    Vital/HR/RR/SleepStage: Sleepad(8) > Radar(4)
// 3. 亲和性权重（AI/设备关系）
//    空间: 同床(5) > 同房(3) > 同单元(1)
//    功能: 同类型全能力(15) > 同类型睡眠(7) > 异构互补(4)
// ============================================================================

// ---- 1. Derive 周期与 tick（素数） ----

const DeriveCycleLen = 331 // 周期秒数，素数

var deriveSchedule = map[int]bool{
	17: true, 53: true, 89: true,
	127: true, 191: true, 307: true,
}

// IsDeriveTick 是否为本轮 derive 定时点（tick 为从进程启动起的秒数）。
func IsDeriveTick(tick int) bool {
	return deriveSchedule[tick%DeriveCycleLen]
}

// ---- 2. Derive 设备信任权重 ----

const (
	PoseWeightSleepad  = 8
	PoseWeightRadar    = 6
	VitalWeightSleepad = 8
	VitalWeightRadar   = 4
)

// PoseWeight 返回该设备类型对 Pose 的信任权重。
func PoseWeight(deviceType string) int {
	switch strings.ToLower(deviceType) {
	case "sleepad", "sleeppad":
		return PoseWeightSleepad
	case "radar":
		return PoseWeightRadar
	default:
		return PoseWeightRadar
	}
}

// VitalWeight 返回该设备类型对 Vital/HR/RR/SleepStage 的信任权重。
func VitalWeight(deviceType string) int {
	switch strings.ToLower(deviceType) {
	case "sleepad", "sleeppad":
		return VitalWeightSleepad
	case "radar":
		return VitalWeightRadar
	default:
		return VitalWeightRadar
	}
}

// ---- 3. Affinity 亲和性权重 ----

const (
	SpatialSameBed  = 5
	SpatialSameRoom = 3
	SpatialSameUnit = 1

	FuncFullMatch  = 15 // same-type full-capability (Radar vs Radar)
	FuncSleepMatch = 7  // same-type sleep mode
	FuncComplement = 4  // heterogeneous (Radar vs SleepPad)
)

// ---- 4. Activity / UpdateStateFromActivity 阈值（Target、RoomState 写回条件） ----

const (
	WalkSecThreshold            = 10  // walk_duration >= 且 walk_distance >= 下值 → 更新 Target.LastActiveTs
	WalkDistanceMetersThreshold = 2   // 米
	StandingContinuousSec      = 55  // stand_duration >= 55 则累计 StandingContinuousMin，否则清零
	StandingContinuousMin      = 5   // 累计 >= 5 分钟则设 StandingContinuousMin=5 并 push RoomState
	MultiPersonDurationSec     = 30  // multi_person_duration >= 30 → HasMulti/访客逻辑
	VisitorMinThreshold        = 10  // 访客停留 >= 10 分钟才计 HasVisitorToday
)

// ---- 5. LeftBedFall（离床跌倒）累计与判定 ----

const (
	LeftBedFallMaxActivityCount  = 5   // 只处理离床后前 N 个 Activity
	LeftBedFallSumStandThreshold = 260 // sum_stand >= 260 秒 → 报 SuspectedFall Warning
	LeftBedFallSumMoveUpdateSec = 10  // sum_move >= 10 秒则更新 target lastActive
	LeftBedFallDanceUpdateMin   = 1   // dance >= 1 分钟则更新 target lastActive
)

// ---- 6. Derive 时间点 / 阈值（在床/离床推导） ----

const (
	VitalDeriveRoundsNeeded = 5   // bed_status!=0/1 时累计 N 轮 track 再判定
	DeriveBedStateThreshold = 350 // 累加置信度 >= 判定在床，<= -350 判定离床（Sleepad）；否则 5 轮后按 sum 判离床
)
