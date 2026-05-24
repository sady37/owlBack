package card

// ========== ID-Name 标识符结构体（避免重复定义）==========

// BedIdentifier 床标识符（ID-Name对）
type BedIdentifier struct {
	BedID   string `json:"bed_id"`
	BedName string `json:"bed_name,omitempty"`
}

// RoomIdentifier 房间标识符（ID-Name对）
type RoomIdentifier struct {
	RoomID   string `json:"room_id"`
	RoomName string `json:"room_name,omitempty"`
}

// UnitIdentifier 单元标识符（ID-Name对）
type UnitIdentifier struct {
	UnitID   string `json:"unit_id"`
	UnitName string `json:"unit_name,omitempty"`
}

// BranchIdentifier 院区标识符（ID-Name对）
type BranchIdentifier struct {
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`
}

// ========== 静态数据结构体 ==========

// ActiveBedRow 活跃床位信息（用于卡片创建）
type ActiveBedRow struct {
	BedID      string  `json:"bed_id"`
	BedName    string  `json:"bed_name"`
	RoomID     string  `json:"room_id,omitempty"`
	RoomName   string  `json:"room_name"`
	ResidentID *string `json:"resident_id,omitempty"`
	Nickname   *string `json:"nickname,omitempty"`
}

// ShortCodeOf — 6 位 base36 短码：SHA256(spatial_prefix) → mod 36^6 → 0-padded
// 用于 UI Card 显示 + DDNS FQDN（替代冗长的 hierarchy 编号）。
// 性质：
//   - 纯函数：同 spatial_prefix 永得同 code，无 DB 状态
//   - 空间 36^6 = 21.8 亿，万卡级 tenant 碰撞概率 < 1e-4（生日悖论）
//   - 不可逆：UI 显示用，不替代 spatial_prefix 主键
//
// 例: "fd00:0:3:111:3::/80" → "j0yvp3"
func ShortCodeOf(spatialPrefix string) string {
	if spatialPrefix == "" {
		return ""
	}
	h := sha256Sum([]byte(spatialPrefix))
	// 前 8 byte 作 uint64 big-endian
	var v uint64
	for i := 0; i < 8; i++ {
		v = (v << 8) | uint64(h[i])
	}
	const base36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	const modulo = uint64(36 * 36 * 36 * 36 * 36 * 36) // 36^6
	v %= modulo
	out := []byte("000000")
	for i := 5; i >= 0; i-- {
		out[i] = base36[v%36]
		v /= 36
	}
	return string(out)
}

// sha256Sum — 简单包装，避免在 type 文件直接 import crypto（避免循环风险）。
// 实现见 short_code.go
func sha256Sum(b []byte) [32]byte { return sha256SumImpl(b) }

// UnitType 居住单位类型 — 与 units.unit_type SMALLINT 一一对应。
const (
	UnitTypeUnknown = 0
	UnitTypePrivate = 1
	UnitTypeShare   = 2
	UnitTypePublic  = 3
)

// UnitProperty 业务属性 — 与 units.unit_property SMALLINT 一一对应。B2C/B2B 分流由此决定。
const (
	UnitPropertyHome     = 0 // B2C
	UnitPropertyFacility = 1 // B2B
)

// UnitInfo — Unit (/80) 静态属性（2026-05-20 重设计：扁平化，仅 unit 级字段；不含 RoomName/BedName）。
// 跨层关系（room/bed/device）由 IPv6 prefix containment 派生，不在 struct 里嵌套。
type UnitInfo struct {
	UnitID       string `json:"unit_id,omitempty"`       // INET CIDR /80
	UnitName     string `json:"unit_name,omitempty"`
	UnitType     int    `json:"unit_type,omitempty"`     // 0=Unknown 1=Private 2=Share 3=Public
	UnitProperty int    `json:"unit_property,omitempty"` // 0=Home (B2C) 1=Facility (B2B)
	BranchID     string `json:"branch_id,omitempty"`     // INET CIDR /56
	BranchName   string `json:"branch_name,omitempty"`
	BuildingID   string `json:"building_id,omitempty"`   // INET CIDR /60 (building 段；跨 floor 稳定)
	BuildingName string `json:"building_name,omitempty"`
	FloorID      string `json:"floor_id,omitempty"`      // INET CIDR /64 (= site_id；per-floor 唯一)
	FloorName    string `json:"floor_name,omitempty"`    // "1F"/"2F"…
	Floor        int    `json:"floor,omitempty"`         // sites.floor (1..14)
	Timezone     string `json:"timezone,omitempty"`      // IANA
	IsPublic     bool   `json:"is_public,omitempty"`     // = (UnitType == UnitTypePublic)
	IsSharedUnit bool   `json:"is_shared_unit,omitempty"`// = (UnitType == UnitTypeShare)
}

// RoomInfo — Room (/88) 静态属性（2026-05-20 新增）。
// RoomType 从 RoomState 挪到这里 —— room_type 是 room 的固有静态属性（不动态变化），
// 跟"当前房内多少人"这种状态正交。
type RoomInfo struct {
	RoomID   string `json:"room_id,omitempty"`   // INET CIDR /88
	RoomName string `json:"room_name,omitempty"`
	RoomType int    `json:"room_type,omitempty"` // 0=Default 1=Bathroom 2=Kitchen
}

// BedInfo — Bed (/96) 静态属性（2026-05-20 新增）。
type BedInfo struct {
	BedID   string `json:"bed_id,omitempty"`   // INET CIDR /96
	BedName string `json:"bed_name,omitempty"`
}

// DeviceInfo device information（Phase 2 一刀切后）。
//
// 命名收口：
//   - DeviceAddr = INET /128 canonical text，业务侧寻址 (devices.device_addr，spatial 可重分配)
//   - DeviceUID  = logMAC，硬件 identity 不变量 (HIPAA 不向前端暴露)
//
// 删除字段（IPv6 prefix containment 自然派生，不需要存）：
//   - BoundBedID / BoundRoomID：FE 用 device_addr 跟 RoomInfo.RoomID/BedInfo.BedID 做 prefix.Contains() 判
//   - UnitID：同上，从 device_addr /80 mask 派生
type DeviceInfo struct {
	DeviceAddr        string `json:"device_addr"` // INET /128 canonical text，业务侧主键
	DeviceUID         string `json:"-"`           // logMAC，HIPAA 不向前端暴露
	DeviceCode        string `json:"device_code"`
	DeviceName        string `json:"device_name"`
	DeviceType        string `json:"device_type"` // "Radar" / "Sleepad"
	DeviceModel       string `json:"device_model"`
	MonitoringEnabled bool   `json:"monitoring_enabled"`
	Status            string `json:"status"` // "online" | "offline" | "error" | "disabled"
}

type ServiceLevelInfo struct {
	LevelCode     string `json:"level_code"`
	DisplayName   string `json:"display_name"`
	DisplayNameCN string `json:"display_name_cn,omitempty"`
	ColorTag      string `json:"color_tag"`
	ColorHex      string `json:"color_hex"`
	Priority      int    `json:"priority"`
}

// ResidentInfo resident information
type ResidentInfo struct {
	ResidentID       string            `json:"resident_id"`
	DNSShortName     string            `json:"dns_short_name,omitempty"` // 6 位 base36，URL 友好替代 IPv6 hoa（与 cards.dns_short_name 同语义）
	LastName         string            `json:"last_name,omitempty"`
	FirstName        string            `json:"first_name,omitempty"`
	Nickname         string            `json:"nickname,omitempty"`
	ServiceLevel     string            `json:"service_level,omitempty"`
	ServiceLevelInfo *ServiceLevelInfo `json:"service_level_info,omitempty"`
	BedID            *string           `json:"bed_id,omitempty"`
	BedName          *string           `json:"bed_name,omitempty"`
}

// CardStatic 卡片静态视图（v2.5）。
//
// 设计：
//   - card_id 即 cards.card_id（/88 INET，住 unit 下 card slot 段 [0xF0..0xFE]）
//   - 跟 spatial 实体（room/bed/device）的关系靠 *.card_id FK 反查 + 卡内 snapshot bool 派生
//   - struct 平铺 *Info（unit/room/bed），不嵌套；FE 按需取非空
//
// CardType v2.5：只区分 'card'（业务） / 'public'（公共）二态；由 cards_display view 派生
type CardStatic struct {
	// 基础信息（cards 表权威）
	CardID       string `json:"card_id"`
	CardType     string `json:"card_type"`                  // 派生自 masklen + card_name：'tenant'|'branch'|'site'|'unit'|'public'|'room'|'bed'|'device'
	CardName     string `json:"card_name"`                  // nickname / NoOne / public
	DNSShortName string `json:"card_dns,omitempty"`         // = cards.card_dns（DB 列已重命名；Go 字段名保留）

	// 静态属性（平铺；按 card 下 has_bed/has_bathroom/has_kitchen 决定哪个非空）：
	//   bed card  → Bed 非空 + Room 非空 (含 Unit)
	//   room card → Room 非空 (含 Unit)
	//   unit/public card → Unit 非空
	Unit *UnitInfo `json:"unit,omitempty"`
	Room *RoomInfo `json:"room,omitempty"`
	Bed  *BedInfo  `json:"bed,omitempty"`

	// 住户和设备（实时查询，非 JSONB snapshot）
	Residents []ResidentInfo `json:"residents"`
	Devices   []DeviceInfo   `json:"devices"`

	// 护理人员（从 resident_caregivers 聚合）
	CaregiverGroups []string        `json:"caregiver_groups,omitempty"`
	Caregivers      []CaregiverInfo `json:"caregivers,omitempty"`

	// 报警显示控制（阈值 = syslog 级别；图标/弹出条件）
	IconAlarmLevel *int `json:"icon_alarm_level,omitempty"`
	PopAlarm       *int `json:"pop_alarm,omitempty"`
}

// CaregiverInfo 护理人员信息
type CaregiverInfo struct {
	UserID      string `json:"user_id"`
	Nickname    string `json:"nickname,omitempty"`
	UserAccount string `json:"user_account,omitempty"`
	Role        string `json:"role,omitempty"`
}

// ========== Realtime Monitor 聚合（card:realtime:stream） ==========

// TrackFields 单条 track 的平铺字段（仅含有效字段 + ts）
type TrackFields map[string]interface{}

// DeviceRealtimeTracks 单设备 monitor 流：track_id (str) -> TrackFields
type DeviceRealtimeTracks map[string]TrackFields

// CardRealTime 卡片实时数据（monitor 格式）
type CardRealTime struct {
	CardID    string                          `json:"card_id"`
	Timestamp int64                           `json:"timestamp,omitempty"`
	Devices   map[string]DeviceRealtimeTracks `json:"devices,omitempty"` // device_uid -> track_id -> fields+ts
}

// ========== Card Status (card:state / card:status:stream) ==========

// DeviceStatus 单个设备运行时真相（独立 Hash device:status:{addr}）。
// 仅 backend 内部使用——data 层把字段重 pack 给 FE，不直接 JSON marshal 本 struct。
type DeviceStatus struct {
	DeviceUID  string `json:"-"`
	DeviceAddr string `json:"device_addr"`
	DeviceType string `json:"device_type"`

	UpdatedAt  int64 `json:"updated_at,omitempty"`
	LastSeenMs int64 `json:"last_seen_ms,omitempty"`

	Offline        int `json:"offline"`
	SignalPoor     int `json:"signal_poor,omitempty"`
	AngleAbnormal  int `json:"angle_abnormal,omitempty"`
	SensorDetached int `json:"sensor_detached,omitempty"`
}

const BedStateDurationNotSet int = -1

// BedState 在/离床状态。
// 注：OOB 不构成独立 risk（单源不够）；OOB 作为输入证据喂到 sensor fall 链路
// （BedsideFall / BedroomLostFall / Still fall），不再有 BedState.RiskLevel 字段。
//
// **字段-时刻配对约定（2026-05-20 重设计）**：
//
//   每个值 `<field>` 配 ts `<field>_ts`（unix ms）— 该值变更到当前内容的时刻。
//   **state-change-anchored**：
//     - 值不变 → ts 不变（重复观测不刷新）
//     - 值翻转 → ts 更新为翻转时刻
//   消费者派生：
//     - 时长   = now - <field>_ts （不存 DurationSec）
//     - "整体 UpdatedAt" = max(各 *_ts) （不存独立字段）
//     - 用户活动 anchor = TargetState.LastActiveTs（不在 BedState 里）
//   omitempty: ts=0 表示"该字段从未被观测过"（startup placeholder 等）。
type BedState struct {
	// BedStatus — 在/离床二态。**直接信号，不反推**。
	//   0 = InBed（在床；sleepad 压感或 radar firmware bed-area track 触发）
	//   1 = NotInBed/LeftBed（离床）
	BedStatus int `json:"bed_status"`
	// BedStatusTs — BedStatus **翻转到当前值** 时刻；值不变时不刷新。
	// FE 显 "InBed Xm" / "LeftBed Xm" 的 anchor = now - BedStatusTs。
	BedStatusTs int64 `json:"bed_status_ts,omitempty"`

	// TrackNumber — bed FSM 当前 Count（zoneengine ZoneState.Count）。
	//   bed zone：0 或 1（不建模"多人共床"）
	TrackNumber int `json:"track_number"`
	// TrackNumberTs — TrackNumber 变到当前值时刻；值不变不刷新。
	TrackNumberTs int64 `json:"track_number_ts,omitempty"`

	// BedConfidence — 驱动当前 BedStatus 的信号源信任档：
	//   0  = 无数据 / 60 = radar 派生 / 90 = sleepad 原生
	BedConfidence int `json:"bed_confidence"`
	// BedConfidenceTs — BedConfidence 变到当前档时刻；值不变不刷新。
	BedConfidenceTs int64 `json:"bed_confidence_ts,omitempty"`

	// BedEvent — 最近 transition **event** 类型（非 state；每次 transition 都更新 Ts）：
	//   0 = Occupied/Returned（刚 InBed）
	//   1 = Vacant（刚 LeftBed）
	//   8 = None（Leaving / CountChange 中间过渡，不算明确 event）
	// state-change-anchored 例外：BedEvent 是 event，每次发生都刷新 BedEventTs。
	BedEvent int `json:"bed_event"`
	// BedEventTs — 最近 transition event 发生时刻。
	BedEventTs int64 `json:"bed_event_ts,omitempty"`

	// SleepStage — 睡眠阶段（仅 sleepad 自带睡眠分类时有）：
	//   0 = Initial / 1 = Awake / 2 = LightSleep / 4 = DeepSleep / 8 = Unknown
	// 由 sensor SleepStageConsumer 走 bed.sleepstage 独立 category 写（confidence ladder）。
	// **bed.state category 不动此字段**（projector 字段级 merge 保留 prev）。
	SleepStage int `json:"sleep_stage"`
	// SleepStageTs — SleepStage **变到当前值** 时刻；值不变不刷新。
	SleepStageTs int64 `json:"sleep_stage_ts,omitempty"`

	// SleepConfidence — SleepStage 的置信度（信号源信任档）：
	//   0  = 无数据 / 60 = radar 推断（不准）/ 90 = sleepad 原生（权威）
	// confidence ladder 详 [[cardagg_v1_to_v2_migration_audit]] S4。
	SleepConfidence int `json:"sleep_confidence"`
	// SleepConfidenceTs — SleepConfidence 变到当前档时刻。
	SleepConfidenceTs int64 `json:"sleep_confidence_ts,omitempty"`
}

const (
	SleepStageInitial = 0
	SleepStageAwake   = 1
	SleepStageLight   = 2
	SleepStageDeep    = 4
	SleepStageUnknown = 8
)

// MaxTs BedState 内所有 per-field ts 的最大值 — 派生"最新更新时刻"。
// 用途：consumer 端 stale 判定 / "整体 UpdatedAt" 派生。
func (b *BedState) MaxTs() int64 {
	if b == nil {
		return 0
	}
	m := b.BedStatusTs
	if b.TrackNumberTs > m {
		m = b.TrackNumberTs
	}
	if b.BedConfidenceTs > m {
		m = b.BedConfidenceTs
	}
	if b.BedEventTs > m {
		m = b.BedEventTs
	}
	if b.SleepStageTs > m {
		m = b.SleepStageTs
	}
	if b.SleepConfidenceTs > m {
		m = b.SleepConfidenceTs
	}
	return m
}

// MaxTs RoomState 内所有 per-field ts 的最大值。
func (r *RoomState) MaxTs() int64 {
	if r == nil {
		return 0
	}
	m := r.TotalPeopleTs
	if r.LastEnterTs > m {
		m = r.LastEnterTs
	}
	if r.LastExitTs > m {
		m = r.LastExitTs
	}
	if r.AloneSinceTs > m {
		m = r.AloneSinceTs
	}
	if r.RiskLevelTs > m {
		m = r.RiskLevelTs
	}
	return m
}

// RiskLevel 风险等级 — RoomState.RiskLevel 用；FE 据此配色。
// 由 sensor zoneengine 按 room_type + risk-time + 多人陪伴评估。
const (
	RiskNormal    = 0 // 绿
	RiskMuted     = 1 // 灰
	RiskAttention = 2 // 黄
	RiskRisk      = 3 // 红
)

// RoomType 房间语义类型；只标"特殊高风险位置"——其余 room（bedroom/living/dining/lobby/etc.）
// 统一走 Default，BE 不细分。
//   - Default(0)：默认 standing-only risk 阈值
//   - Bathroom(1)：stay+standing 严苛阈值 + bathroom_fall + bathroom ghost adjudicator
//   - Kitchen(2)：占位给未来 lost-fall 例外（老人不下厨）；当前与 Default 同处理
//
// 注意：public 不是 room 类型，是 unit 类型（unit_type=3）；不要混进来。
const (
	RoomTypeDefault  = 0
	RoomTypeBathroom = 1
	RoomTypeKitchen  = 2
)

// RoomState per-physical-room (/88) 聚合状态。所有 kind 共用一个 shape，risk 语义由 Kind 区分。
//
// **字段-时刻配对约定（同 BedState，2026-05-20 重设计）**：
//   每个状态值配 `<field>_ts`，state-change-anchored（值不变 ts 不变）。
//   消费者派生时长 = now - <field>_ts；不存 StaySec。
//   max(各 *_ts) = "整体 UpdatedAt"，不存独立字段。
//
// 注：unit /80 没有 unit_state — unit 卡 FE 自己按 /80 prefix 列子 /88 room，挑最新 ts 的显示。
// sensor 内部 unit-scope risk（如 NightAbsence of Room）走自己的 counter，不入此 hash。
type RoomState struct {
	// RoomIdentifier — sensor producer 端填 RoomID = /88 prefix (envelope.subject_entity)；
	// cardagg 派生 display 时按 RoomID 反查 entry 决定 room_icon_kind 等（保证 latest event 同源）。
	// RoomName 由 cardagg 反查 rooms.room_name 填。
	RoomIdentifier

	// 注：RoomType 是静态属性 → 已迁出本 hash，由 cardagg CardMeta (rooms.room_type LPM /88) 派生，
	// FE 从 CardStatic.Room.RoomType 取。sensor / cardagg 内部走各自 cache，不进 RoomState。

	// TotalPeople — 当前房内人数（radar number_people + sleepad in_bed 融合推断）；
	// 不是绝对真实总人数，监控盲区/识别误差会偏移。
	TotalPeople int `json:"total_people"`
	// TotalPeopleTs — TotalPeople **数值变到当前值** 时刻；任意值变化都刷新（含 1→2 / 2→1）。
	// **不是 stay anchor**！stay 时长用 LastEnterTs 算。本 ts 仅供"最近一次人数变化"独立查询。
	TotalPeopleTs int64 `json:"total_people_ts,omitempty"`

	// LastEnterTs — **最近一次 0→N+ 进房 transition** 时刻（房间从空 (0) 变占用 (>0) 瞬间）。
	// 用途：判 "今日有人进过房" / visitor 派生 / 进入冷启动后第一次进入等。
	// 1→2 / 2→1 等占用期内人数波动 **不更新此字段**。
	// **不是 stay 锚点**（stay 锚 = AloneSinceTs，独居语义）。
	LastEnterTs int64 `json:"last_enter_ts,omitempty"`

	// LastExitTs — 最近一次 N→0 离房 transition 时刻（房间变空瞬间）。
	// 房间再变占用后 LastExitTs 保持上次离开的时刻（"上一段空房结束时刻"语义）。
	LastExitTs int64 `json:"last_exit_ts,omitempty"`

	// AloneSinceTs — **最近一次 TotalPeople 变到 1 的时刻**（独居开始锚点）。
	// **stay 时长唯一锚点（独居语义）**：
	//   消费者 stay = (TotalPeople==1) ? now - AloneSinceTs : 0
	//
	// 触发更新的 transition：0→1（独自进入）/ 2→1（同伴离开剩单人）/ N→1（任意降到 1）
	// 不触发更新：1→2 / 1→3 等"变多"（独居结束）/ 1→0 / 2→3（始终非 1）
	//
	// 业务意义：bathroom "alone too long" 风险评估、卧室"独睡太久"等都基于此锚点。
	// 1 人独居 + 时长超阈值 = 风险；多人陪伴（count>1）期间 stay 不计。
	AloneSinceTs int64 `json:"alone_since_ts,omitempty"`

	// LastExitToOutside — 最近 Vacant 是否由 EnterArea==outside 触发；不参与 SceneState 派生，
	// 仅留作 risk/alarm 原始信号。bool 字段无 ts（变更频率低，由 LastExitTs 同期判定）。
	LastExitToOutside bool `json:"last_exit_to_outside,omitempty"`

	// RiskLevel — 风险等级：0=Normal 1=Muted 2=Attention 3=Risk；
	// kind-specific 阈值 + night/multi-people 由 sensor 评估。
	RiskLevel int `json:"risk_level,omitempty"`
	// RiskLevelTs — RiskLevel 变到当前等级时刻；值不变不刷新。
	RiskLevelTs int64 `json:"risk_level_ts,omitempty"`
}

// TargetState 单 Target 汇总（老人维度）。
//
// per-device 维度（v2 拍板 [[target_state_per_device]]）：sensor 每个 radar /128 维护自己
// 一份；cardagg max-merge 到单 card.target Hash 写入。
//
// StandingContinuousMin 2026-05-18 已从 RoomState 挪到此处：单 device 内（属同 radar 的物理
// 占用）的连续站立分钟，sensor 累加封顶 8；cardagg max-merge across devices in card。
type TargetState struct {
	UpdatedAt             int64  `json:"updated_at,omitempty"`
	SpatialAnchor         string `json:"spatial_anchor"`		  //描述当前目标的空间锚点， CDIR 格式
	TrackID               int    `json:"track_id"`
	LogicID               string `json:"logic_id,omitempty"`
	LastActiveTs          int64  `json:"last_active_ts,omitempty"`
	StandingContinuousMin int    `json:"standing_continuous_min,omitempty"`
	WeakBiometricSignal   int    `json:"weak_biometric_signal"`
	VisitorStartTs        int64  `json:"visitor_start_ts,omitempty"`
	HasVisitorToday       bool   `json:"has_visitor_today,omitempty"`
}

// AlarmState 告警摘要（v2：cards 表无 alarm 列，counter/pop 由 alarm_events 实时聚合得出）
//
// 计算规则按 sensor_v2.md §6.7 三级告警状态机（决定 17）—— alarm_status 与级别分级处理：
//
//	┌──────────────────────┬─────────────────────────────────────────────────────────────┐
//	│ 级别                  │ 计入 Active* counter 的 alarm_status 集合                    │
//	├──────────────────────┼─────────────────────────────────────────────────────────────┤
//	│ Critical (Emerg/     │ {active, acked, auto_resolved}                              │
//	│  Alert/Crit, lvl0-2) │ — 必须人工 ack 才离开 Pending+AlarmBell；                    │
//	│                      │   auto_resolved 不离开（等待人工 handle）；                   │
//	│                      │   终态 = resolved → 不计入                                  │
//	├──────────────────────┼─────────────────────────────────────────────────────────────┤
//	│ Error (lvl 3)        │ {active} 仅                                                  │
//	│ Warning (lvl 4)      │ {active} 仅                                                  │
//	│                      │ — auto_resolved 直接归 Resolved（不强制人工 ack）            │
//	└──────────────────────┴─────────────────────────────────────────────────────────────┘
//
// PopAlarm 选择规则（§6.7.2）：
//   - 仅看 alarm_status='active' 的行（acked/auto_resolved 不参与 popAlarm 选秀）
//   - 高级别 cover 低（level int ASC：Emerg=0 优先）
//   - 同级别 新 cover 旧（triggered_at DESC）
//   - 结果取 1 条 → "<LEVEL>.<event_type>"（如 "EMERG.Fall"）
//   - 无 active 行 → PopAlarm="" + EventID=""
//
// UI 渲染三件套（owlFront Overview.vue 消费）：
//   - PopAlarm bar：仅当 PopAlarm != "" 显示（active 才上 bar）
//   - AlarmBell 着色：由 Active* counter 任一 > 0 触发（Critical auto_resolved 保持 Bell 红）
//   - Pending 列表：详情页查询，覆盖 active + acked + auto_resolved（未到 resolved 终态）
type AlarmState struct {
	UpdatedAt   int64 `json:"updated_at,omitempty"`
	TriggeredAt int64 `json:"triggered_at,omitempty"`

	// Active*：未达 Resolved 终态的告警计数（Critical 含 acked/auto_resolved，Warning/Error 仅 active）
	ActiveEmerg   int `json:"active_emerg"`   // lvl 0 — Critical 语义
	ActiveAlert   int `json:"active_alert"`   // lvl 1 — Critical 语义
	ActiveCrit    int `json:"active_crit"`    // lvl 2 — Critical 语义
	ActiveErr     int `json:"active_err"`     // lvl 3 — auto_resolved 即归 Resolved，不计入
	ActiveWarning int `json:"active_warning"` // lvl 4 — auto_resolved 即归 Resolved，不计入

	// PopAlarm：仅 status=active 行参与；空字符串表示当前无 active 告警（bar 不显示）
	PopAlarm string `json:"pop_alarm"` // "<LEVEL>.<event_type>" 如 "EMERG.Fall" / ""
	EventID  string `json:"event_id"`  // popAlarm 对应 alarm_events.event_id（UI dedup）
}

// ========== CardDisplay：FE Overview dumb-render 契约 ==========
//
// cardagg 投影到 card:state.display Hash 子字段的显示态；FE 单读它就能渲 Overview，
// 不需要也不应读 room_state/bed_state/target/alarm_state（除 alarm_state.pop_alarm 用作
// Section1.up/down.right 自派生）。
//
// DisplayTime 算法：FE 用 `serverTimeRef - <Section>AnchorMs` 算时长，不用本地时钟。
// serverTimeRef 由 SSE `event: tick` 每 ~30s 推送的 server_ts 维护。
//
// 字段对应 UI 4 区：
//   - Section1.up.right / Section1.down.right：alarmBell + alarmEvent — 不在 display，FE 从 alarm_state 派生
//   - Section1.down.left：active room 名 — 不在 display，FE 仅 card.bed_id==null 时按业务渲（暂留空）
//   - Section2.left：Section2LeftMode + bed_status + sleep_stage / room_*
//   - Section2.middle/right：vital/posture — 不在 display，来自 realtime stream
//   - Section3.up.left：ActiveState + ActiveAnchorMs
//   - Section3.up.right：SceneState + SceneAnchorMs
//   - Section3.down：VisitorState + VisitorAnchorMs
//
// 详 owlBack/doc/card_display.md。

// BedStatus 二态枚举（保留 — sensor/cardagg 内部 BedState.BedStatus 仍用）。
const (
	BedStatusInBed    = 0 // 在床
	BedStatusNotInBed = 1 // 离床
)

// Section2LeftIcon — FE 单 switch 决定 base icon；cardagg 派生填这个数。
// 床 sub-variant（Awake/LightSleep/DeepSleep/Empty）由 FE 端从 BedState 或 monitor stream
// 自家派生（sleep_stage / bed_status），不在 display schema 暴露。
const (
	Section2IconEmpty             = 0 // 无 event 且无床 — FE 渲灰底
	Section2IconRoomNormal        = 1 // 普通房间有人，无 risk
	Section2IconBedInUse          = 2 // 床卡（FE 自渲 InBed/LeftBed/Sleep 变体）
	Section2IconBathroomNormal    = 3 // 卫生间占用，无 risk
	Section2IconRoomAttention     = 4 // 普通房间 + Risk=2 黄
	Section2IconBathroomAttention = 5 // 卫生间 + Risk=2 黄
	Section2IconRoomRisk          = 6 // 普通房间 + Risk=3 红
	Section2IconBathroomRisk      = 7 // 卫生间 + Risk=3 红Section2IconEmpty
)

const (
	ActiveStateInactive = 0 // FE 渲 "Active <DisplayTime> ago"
	ActiveStateNow      = 1 // FE 渲 "Active now"（DisplayTime 不参与）
)

const (
	SceneStateOOR    = 0 // 房间空（total_people==0 + 无人在床）
	SceneStateInRoom = 1
	SceneStateInBath = 2 // 含 Stay / Stand 子状态由 SceneAnchorMs 携带
	SceneStateInBed  = 3
	SceneStateOOB    = 4 // 离床（在房）
)

const (
	VisitorStateNone  = 0 // "No visitor today"
	VisitorStateNow   = 1 // "Visitor now · <DisplayTime>"
	VisitorStateToday = 2 // "Visitor today · <DisplayTime>"
)

// VitalTrendLevel — Section3.down.right 横条配色（S4/W1 WeakBio 风险描述符 UI）。
// 由 cardagg card_display_builder 按 Target.WeakBiometricSignal score 阈值派生
// （详 [[target_state_weak_bio_signal_design]]）：
//
//	0-29  → None (hide 横条)
//	30-59 → Gray (Attention)
//	60-79 → Yellow (Watch)
//	80-100→ Red (Alert) — 不独立触发 alarm，由风险放大消费者按 ≥80 提级
const (
	VitalTrendLevelNone   = 0
	VitalTrendLevelGray   = 1
	VitalTrendLevelYellow = 2
	VitalTrendLevelRed    = 3
)

// CardDisplay — 写入 card:state hash 的 `display` JSON 字段。
type CardDisplay struct {
	UpdatedAt int64 `json:"updated_at"`

	// Section1.down.left — 空间 scope 标签（"GuestRoom" / "Bedroom" / "Bathroom" / "Whole Unit"）
	Section1DownLeft string `json:"section1_down_left,omitempty"`

	// Section2.left — FE dumb-renderer 3+1 字段；cardagg 完成全部派生
	Section2LeftIcon int    `json:"section2_left_icon"`           // 单 switch 决定 icon 资源（枚举见 Section2Icon*）
	Section2LeftUp string `json:"section2_left_uptext,omitempty"` // icon 顶部小标签：room source X>1 "RoomName:XP" / X<=1 "RoomName"；bed source ""（隐私）
	Section2LeftDown string `json:"section2_left_downtext,omitempty"` //icon 顶部小标签 显示room内 StandingContinuousMin静止站立时长 /Alonetime 独处时长 
	Section2LeftRisk int    `json:"section2_left_risk,omitempty"` // 0/1/2/3 → FE 配色（复用 RiskLevel*）

	// Section3.up.left
	ActiveState    int   `json:"active_state"`
	ActiveAnchorMs int64 `json:"active_anchor_ms,omitempty"`

	// Section3.up.right
	SceneState    int   `json:"scene_state"`
	SceneAnchorMs int64 `json:"scene_anchor_ms,omitempty"`

	// Section3.down.left (Visitor / Bed timing fallback — bed_anchor_ms 在 VisitorState=None 时配合 BedStatus 渲 InBed/LeftBed Xm)
	VisitorState    int   `json:"visitor_state"`
	VisitorAnchorMs int64 `json:"visitor_anchor_ms,omitempty"`
	BedAnchorMs     int64 `json:"bed_anchor_ms,omitempty"` // = bs.start_time；leaf 卡 BedState 存在时填

	// Section3.down.right (WeakBio 横条配色；VitalTrendLevel* 0=hide/1=Gray/2=Yellow/3=Red)
	VitalTrendLevel int `json:"vital_trend_level,omitempty"`
}

// CardStatus 卡片状态数据
//
// Display 字段 by cardagg card_display_projector 投影；FE Overview 卡列表 dumb render 读此字段，
// 不再派生 picker/scene/tier。RoomState/BedState 给 Detail/RadarCanvas 等诊断视图保留。
type CardStatus struct {
	CardID     string                 `json:"card_id"`
	Target     *TargetState           `json:"target,omitempty"`
	RoomState  *RoomState             `json:"room_state,omitempty"`
	BedState   *BedState              `json:"bed_state,omitempty"`
	AlarmState *AlarmState            `json:"alarm_state,omitempty"`
	Display    *CardDisplay           `json:"display,omitempty"`
	Message    map[string]interface{} `json:"message,omitempty"`
}

// CardType v2.5：cards 表不再存 card_type 列（masklen + card_name 派生）。
// 业务层若需要"卡型字符串"调 view cards_display.card_type；Go 代码用下面两个常量
// 区分 public 卡 vs 业务卡。
const (
	CardTypeCard   = "card"   // 业务卡（active_bed / room / etc.）
	CardTypePublic = "public" // 公共区卡（card_name = 'public'）
)
