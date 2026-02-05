package models

import "time"

// BedStateTracker 床状态追踪（按 bed 维度；事件可来自雷达或床垫）。
// 在床/离床两状态一一对应：同一时刻只运行当前状态对应的 timer，另一个为 nil 或已 stop。
// LeftBedTooLong 等由 Wisefido-AI 基于此计算。卡片聚合仅做简单计算（雷达人数、姿态）；不在此做时间序列推断。
type BedStateTracker struct {
	UnitID       string     // 单元
	BedID        string     // 床
	CurrentState string     // "in_bed" | "out_of_bed"
	StateSince   time.Time  // 当前状态起始时间
	LeftBedTimer *time.Timer // 离床时长 → LeftBedTooLong 等（仅运行时，不持久化）
	InBedTimer   *time.Timer // 在床时长 → 在床相关规则（仅运行时，不持久化）
}

// RoomStateTracker 房间状态追踪（按 room 维度；人数比 person_count/postures 更准确）。
// Stay、NoActivity24h 由 Wisefido-AI 计算：如雷达无人但床有人 → 房间有人但无活动，而非简单报房间无人。
type RoomStateTracker struct {
	UnitID        string     // 单元
	RoomID        string     // 房间
	PeopleCount   int        // 人数
	LastLeaveTime time.Time  // 最后离开时间
	StayTimer     *time.Timer // Stay 报警计时器（仅运行时，不持久化）
	ActivityTimer *time.Timer // NoActivity24h 计时器（仅运行时，不持久化）
}
