package observation

// Track 表示一条观测轨迹（一人/一占位）的实时状态。
// 设备按能力填充字段；设备级属性在 DeviceSnapshot，不在此。
// LogicID 由 cardagg 收到 DT 后从 Card 申请并写入，用于统一逻辑目标。
type Track struct {
	TrackID  int    `json:"track_id"`  // Radar 0-8, Sleepad 0=left
	LogicID  string `json:"logic_id,omitempty"` // Card 分配的逻辑目标 ID
	Ts       int64  `json:"ts"`

	PositionX       *int   `json:"position_x"`
	PositionY       *int   `json:"position_y"`
	PositionZ       *int   `json:"position_z"`
	AreaID          *int   `json:"area_id,omitempty"` // firmware 区域号(0-indexed,0-15 均有效区)；255=声明区域之外；nil=未上报/设备无区域概念(sleepad)。指针区分 nil(无) vs &0(Area 0)
	Event         int    `json:"event"`          // 事件类型: 0=无, 1=进房, 2=离房, 3=进区, 4=离区
	TrackConfidence int    `json:"track_confidence,omitempty"`

	Pose           int `json:"pose"`
	PoseConfidence int `json:"pose_confidence,omitempty"`

	HeartRate        int     `json:"heart_rate,omitempty"`
	RespiratoryRate  int     `json:"respiratory_rate,omitempty"`
	BloodOxygen      int     `json:"blood_oxygen,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	VitalConfidence  int     `json:"vital_confidence,omitempty"`

	// 周期≤5min 上报视为实时；Sleepad 持续更新。
	// 取值 BedStatusInBed=0 / BedStatusLeftBed=1 / BedStatusUnchanged=8（无床概念设备如 radar 即此）。
	BedStatus int `json:"bed_status,omitempty"`

	// 次数（int），非动作枚举：body_move 适用于 BM8701-2 固件≥5.x 及 M901L；turn_over 适用于 BM8701-2 固件≤2.x
	BodyMove int `json:"body_move,omitempty"` // 体动次数
	TurnOver int `json:"turn_over,omitempty"` // 翻身次数
}

// FromFieldMap 从扁平 map 填充 Track（如 MonitorBuffer 的 fields）。
func (t *Track) FromFieldMap(m map[string]any) {
	if v, ok := intVal(m, FieldTrackID); ok {
		t.TrackID = v
	}
	if v, ok := m[FieldLogicID].(string); ok {
		t.LogicID = v
	}
	if v, ok := intVal(m, FieldPositionX); ok {
		t.PositionX = &v
	}
	if v, ok := intVal(m, FieldPositionY); ok {
		t.PositionY = &v
	}
	if v, ok := intVal(m, FieldPositionZ); ok {
		t.PositionZ = &v
	}
	if v, ok := intVal(m, FieldTrackConfidence); ok {
		t.TrackConfidence = v
	}
	if v, ok := intVal(m, FieldEvent); ok {
		t.Event = v
	}
	if v, ok := intVal(m, FieldAreaID); ok {
		t.AreaID = &v
	}
	if v, ok := intVal(m, FieldPose); ok {
		t.Pose = v
	}
	if v, ok := intVal(m, FieldPoseConfidence); ok {
		t.PoseConfidence = v
	}
	if v, ok := intVal(m, FieldHeartRate); ok {
		t.HeartRate = v
	}
	if v, ok := intVal(m, FieldRespiratoryRate); ok {
		t.RespiratoryRate = v
	}
	if v, ok := intVal(m, FieldBloodOxygen); ok {
		t.BloodOxygen = v
	}
	if v, ok := intVal(m, FieldVitalConfidence); ok {
		t.VitalConfidence = v
	}
	if v, ok := intVal(m, FieldBedStatus); ok {
		t.BedStatus = v
	} else {
		t.BedStatus = BedStatusUnchanged
	}
	if v, ok := intVal(m, FieldBodyMove); ok {
		t.BodyMove = v
	}
	if v, ok := intVal(m, FieldTurnOver); ok {
		t.TurnOver = v
	}
}

// ToFieldMap 转为扁平 field map，供 MonitorBuffer.Write。
func (t *Track) ToFieldMap() map[string]any {
	m := make(map[string]any)
	m[FieldTrackID] = t.TrackID
	if t.LogicID != "" {
		m[FieldLogicID] = t.LogicID
	}
	if t.PositionX != nil {
		m[FieldPositionX] = *t.PositionX
	}
	if t.PositionY != nil {
		m[FieldPositionY] = *t.PositionY
	}
	if t.PositionZ != nil {
		m[FieldPositionZ] = *t.PositionZ
	}
	if t.TrackConfidence != 0 {
		m[FieldTrackConfidence] = t.TrackConfidence
	}
	if t.Event != 0 {
		m[FieldEvent] = t.Event
	}
	if t.AreaID != nil {
		m[FieldAreaID] = *t.AreaID
	}
	m[FieldPose] = t.Pose // 0=初始化(PoseUnknown)是有效枚举,无条件发(不可 omit:与"无 pose"无二义)
	if t.PoseConfidence != 0 {
		m[FieldPoseConfidence] = t.PoseConfidence
	}
	if t.HeartRate != 0 {
		m[FieldHeartRate] = t.HeartRate
	}
	if t.RespiratoryRate != 0 {
		m[FieldRespiratoryRate] = t.RespiratoryRate
	}
	if t.BloodOxygen != 0 {
		m[FieldBloodOxygen] = t.BloodOxygen
	}
	if t.VitalConfidence != 0 {
		m[FieldVitalConfidence] = t.VitalConfidence
	}
	// bed_status：Unchanged(8) = 无床数据 → omit；InBed(0) / LeftBed(1) → 写出。
	if t.BedStatus != BedStatusUnchanged {
		m[FieldBedStatus] = t.BedStatus
	}
	if t.BodyMove != 0 {
		m[FieldBodyMove] = t.BodyMove
	}
	if t.TurnOver != 0 {
		m[FieldTurnOver] = t.TurnOver
	}
	return m
}

func intVal(m map[string]any, key string) (int, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case int64:
		return int(val), true
	}
	return 0, false
}
