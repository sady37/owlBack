package card

import "time"

// StateOwner 标识 card:state Hash 字段的归属方
type StateOwner uint8

const (
	OwnerCardAgg StateOwner = iota
	OwnerAI
)

var stateOwnerNames = [...]string{"cardagg", "ai"}

func (o StateOwner) String() string {
	if int(o) < len(stateOwnerNames) {
		return stateOwnerNames[o]
	}
	return "unknown"
}

// StateField card:state:{card_id} Hash 单字段定义
type StateField struct {
	Key       string
	Owner     StateOwner
	HasTTL    bool
	Default   string
	Indexed   bool
	Immutable bool
}

const AIFieldTTL = 10 * time.Minute

// StateFields — card:state:{card_id} Hash 全量字段注册表。
// cardagg 层：JSON blob，对应 CardStatus 的各个块。
// AI 层：扁平 string 字段，带 TTL 自动过期回退。
var StateFields = []*StateField{
	// ---- cardagg ----
	{Key: "target", Owner: OwnerCardAgg, Default: ""}, // 单 Target，当前仅一个
	{Key: "room_state", Owner: OwnerCardAgg, Default: "{}"},
	{Key: "bed_state", Owner: OwnerCardAgg, Default: "{}"},
	// device_status 不在 card:state；独立到 device:status:{ipv6} Hash（详 keys.go）
	{Key: "alarm_state", Owner: OwnerCardAgg, Default: "{}", Indexed: true},
	{Key: "display", Owner: OwnerCardAgg, Default: "{}"}, // CardDisplay 投影（FE Overview dumb render）
	{Key: "message", Owner: OwnerCardAgg, Default: "{}"}, // 与 CardStatus.Message 对应

	// ---- AI 层（带 TTL） ----
	{Key: "risk_level", Owner: OwnerAI, HasTTL: true, Default: "unknown"},
	{Key: "intent_state", Owner: OwnerAI, HasTTL: true, Default: "unknown"},
	{Key: "care_priority", Owner: OwnerAI, HasTTL: true, Default: "unknown"},
	{Key: "activity_zscore", Owner: OwnerAI, HasTTL: true, Default: "0"},
	{Key: "routine_adherence", Owner: OwnerAI, HasTTL: true, Default: "unknown"},
}

// ActiveAlarm 单条活跃告警（alarm_service 内部使用，从 DB 查询后聚合为 AlarmState 写入 Hash）
type ActiveAlarm struct {
	EventID     string `json:"event_id"`
	EventType   string `json:"event_type"`
	AlarmLevel  string `json:"alarm_level"`
	DeviceUID   string `json:"-"`            // logMAC, HIPAA 不向前端暴露
	DeviceAddr  string `json:"device_addr"`  // INET /128 canonical text, 前端展示用
	TriggeredAt int64  `json:"triggered_at"`
}

var stateFieldMap map[string]*StateField

func init() {
	stateFieldMap = make(map[string]*StateField, len(StateFields))
	for _, f := range StateFields {
		stateFieldMap[f.Key] = f
	}
}

// StateFieldLookup 按字段 key 查找 StateField
func StateFieldLookup(key string) *StateField {
	return stateFieldMap[key]
}

// IsStateFieldRegistered 是否已注册的 state 字段
func IsStateFieldRegistered(key string) bool {
	_, ok := stateFieldMap[key]
	return ok
}

// StateFieldHasTTL 是否带 TTL 的 state 字段
func StateFieldHasTTL(key string) bool {
	if f := stateFieldMap[key]; f != nil {
		return f.HasTTL
	}
	return false
}
