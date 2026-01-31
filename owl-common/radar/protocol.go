package radar

import (
	"fmt"
	"strconv"
	"strings"
)

// Radar Qinglan version 3.0.1 2025-11-8
// MonitorTrackData 实时轨迹单条，与 Radar_MQTT_v3.0 3.7.2 track 一致。16 字节×N，N 为人数。
// 字节 0 目标 ID；1~3 轨迹坐标 x/y/z(分米,分米,厘米)；4-11 预留字段；12 剩余时间(秒)；13 人员姿态；14 人员事件；15 区域 ID
type MonitorTrackData struct {
	TargetID      int      `json:"target_id"`       // 0: 有人 0-7，无人 88
	PositionX     int      `json:"position_x"`      // 1: x 分米
	PositionY     int      `json:"position_y"`      // 2: y 分米
	PositionZ     int      `json:"position_z"`      // 3: z 厘米
	Reserved      [8]byte  `json:"reserved"`         // 4-11 预留调试使用
	RemainingTime int      `json:"remaining_time"`   // 12: 秒 0-60，自动测边界时用
	Pose          int      `json:"pose"`             // 13: 0-11
	Event         int      `json:"event"`            // 14: 0 无 1 进房 2 离房 3 进区 4 离区
	AreaID        int      `json:"area_id"`          // 15: 14 为 3 或 4 时有效
}

// VitalData 实时呼吸心率，与 Radar_MQTT_v3.0 3.7.2 bh 一致。定长 16 字节。
// 字节 0 标识符(0)；1 呼吸值 次/分；2 心率值 次/分；3-12 预留；13 状态事件 bit7:6 睡眠状态；14 稳定度 bit1:0；15 预留
type MonitorVitalData struct {
	VitalFlag       int      `json:"vital_flag"`        // 0: 固定 0 表示实时呼吸心率
	RespiratoryRate int      `json:"respiratory_rate"`  // 1: 次/分钟
	HeartRate       int      `json:"heart_rate"`        // 2: 次/分钟
	Reserved        [10]byte `json:"reserved"`          // 3-12 预留调试使用
	SleepStatus     int      `json:"sleep_status"`      // 13: bit 7:6 睡眠状态
	Stability       int      `json:"stability"`         // 14: bit 1:0 动作干扰测量
	Reserved15      byte     `json:"reserved_15"`       // 15
}

// SleepData 统计睡眠，与 Radar_MQTT_v3.0 3.6 sleep 一致。定长 16 字节。
// 字节 0 标识符(0xff)；1 实时呼吸 2 实时心率；3-4 睡眠统计信息不公开；5 分钟级平均呼吸 6 分钟级平均心率；7-12 睡眠统计信息不公开；13 呼吸心率事件；14-15 睡眠统计信息不公开
type StatSleepData struct {
	SleepFlag           int     `json:"sleep_flag"`            // 0: 0xff 睡眠统计
	RespiratoryRate     int     `json:"respiratory_rate"`       // 1: 次/分钟
	HeartRate           int     `json:"heart_rate"`             // 2: 次/分钟
	Reserved_3_4        [2]byte `json:"reserved_3_4"`          // 3-4
	AvgRespiratoryRate   int     `json:"avg_respiratory_rate"`  // 5
	AvgHeartRate         int     `json:"avg_heart_rate"`       // 6
	Reserved_7_12        [6]byte `json:"reserved_7_12"`         // 7-12
	BreathStatus         int     `json:"breath_status"`          // 13 bit 1:0
	HeartStatus         int     `json:"heart_status"`           // 13 bit 3:2
	VitalSignsStatus    int     `json:"vital_signs_status"`    // 13 bit 5:4
	SleepStatus         int     `json:"sleep_status"`           // 13 bit 7:6
	Reserved_14_15      [2]byte `json:"reserved_14_15"`         // 14-15
}

// StatTrackData 统计轨迹，与 Radar_MQTT_v3.0 3.6 track 一致。定长 16 字节，一分钟一次。
// 字节 0 版本标识符；1 人数；2~3 行走距离(米 Big-Endian)；4 行走时长(秒)；5 静坐时长未开放；6 躺卧时长 7 站立时长(秒)；8 多人时长(秒)；9~15 rsv
type StatTrackData struct {
	Version              int     `json:"version"`                // 0: 1 或 2
	PeopleCount          int     `json:"people_count"`           // 1: 0-8
	WalkDistance         int     `json:"walk_distance"`          // 2~3: 米 Big-Endian 0-32767
	WalkDuration         int     `json:"walk_duration"`          // 4: 秒
	Reserved5            byte    `json:"reserved_5"`             // 5: 静坐时长 未开放
	LieDuration          int     `json:"lie_duration"`           // 6: 秒
	StandDuration        int     `json:"stand_duration"`         // 7: 秒
	MultiPersonDuration  int     `json:"multi_person_duration"`  // 8: 秒
	Reserved_9_15        [7]byte `json:"reserved_9_15"`          // 9~15 rsv
}

// EventEnter2OutData 3.5 进出事件 type=1 单条。data 为数组时同事件可含多条轨迹，用 []EventEnter2OutData。
type EventEnter2OutData struct {
	TrackID  int `json:"track_id"`   // track-id：人员轨迹 ID
	Event    int `json:"event"`      // 1 进房 2 离房 3 进区 4 离区 5 进监护 6 退监护
	AreaType int `json:"area_type"`  // 2 普通床 5 监护床 6 感应区
}

// EventEnter2OutPayload type=1 时 data 字段，多条轨迹
type EventEnter2OutPayload []EventEnter2OutData

// EventPoseData 3.5 姿态变化事件 type=2 单条。data 为数组时同事件可含多条轨迹，用 []EventPoseData。
type EventPoseData struct {
	TrackID int `json:"track_id"`
	Pose    int `json:"pose"`   // 2 疑似跌倒 5 确认跌倒 7 疑似坐地 8 确认坐地 10 疑似床上坐起 11 确认床上坐起
}

// EventPosePayload type=2 时 data 字段，多条轨迹
type EventPosePayload []EventPoseData

// EventNumberPeopleData 3.5 人数变化事件 type=3，data 为对象 number-people
type EventNumberPeopleData struct {
	NumberPeople int `json:"number_people"` // 协议字段 number-people
}

// Event type=5 设备在线 data。IsOnline: 0=online, 1=offline
type EventIsOnlineData struct {
	IsOnline int `json:"isOnline"`
}

// Event type=7 信号差 data。Recovery: 0=signal_poor, 1=signal_recovery
type EventSignalPoorData struct {
	Recovery int `json:"recovery"`
}

// Event type=8 倾角异常 data。Recovery: 0=angle_abnormal, 1=angle_recovery
type EventAngleAbnormalData struct {
	Recovery int `json:"recovery"`
}

// Event type=9 其他告警 data
type EventOtherData struct {
	AlarmType string            `json:"alarmType"`
	DeviceUID string            `json:"device_uid"`
	Extra     map[string]string `json:"-"` // 未定义字段
}

// MQTT 事件/告警报文顶层。Data 按 Type 解析为上述 Event*Data
type EventMessage struct {
	Cmd  string      `json:"cmd"`
	Type int         `json:"type"`
	Data interface{} `json:"data"`
}

// PropResponse /prop 响应。RequestID 归一后；Data 透传。
type PropResponse struct {
	RequestID string                 `json:"request_id"`
	Data      map[string]interface{} `json:"data"`
}

// FuncResponse /func 响应。RequestID 归一后。
type FuncResponse struct {
	RequestID string `json:"request_id"`
}

// RectangleData 检测边界，与 3.4.2 一致。4 顶点顺序与协议相同；单位 cm（规范），设备为 dm。
// 设备格式：{x1,y1;x2,y2;x3,y3;x4,y4} 分米
type RectangleData struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
	X3 int `json:"x3"`
	Y3 int `json:"y3"`
	X4 int `json:"x4"`
	Y4 int `json:"y4"`
}

// ParseRectangleFromDM 从设备字符串（分米）解析，坐标 *10 填入结构体（cm）。格式 {x1,y1;x2,y2;x3,y3;x4,y4}
func ParseRectangleFromDM(s string) (*RectangleData, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return nil, false
	}
	if s[0] == '{' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '}' {
		s = s[:len(s)-1]
	}
	points := strings.Split(s, ";")
	if len(points) != 4 {
		return nil, false
	}
	r := &RectangleData{}
	for i, p := range points {
		parts := strings.Split(strings.TrimSpace(p), ",")
		if len(parts) != 2 {
			return nil, false
		}
		x, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		y, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return nil, false
		}
		x, y = x*10, y*10 // dm → cm
		switch i {
		case 0:
			r.X1, r.Y1 = x, y
		case 1:
			r.X2, r.Y2 = x, y
		case 2:
			r.X3, r.Y3 = x, y
		case 3:
			r.X4, r.Y4 = x, y
		}
	}
	return r, true
}

// String 输出规范格式 "{x1,y1;x2,y2;x3,y3;x4,y4}"，单位 cm
func (r *RectangleData) String() string {
	return fmt.Sprintf("{%d,%d;%d,%d;%d,%d;%d,%d}",
		r.X1, r.Y1, r.X2, r.Y2, r.X3, r.Y3, r.X4, r.Y4)
}

// ToDMString 输出设备格式（分米），坐标 /10
func (r *RectangleData) ToDMString() string {
	return fmt.Sprintf("{%d,%d;%d,%d;%d,%d;%d,%d}",
		r.X1/10, r.Y1/10, r.X2/10, r.Y2/10, r.X3/10, r.Y3/10, r.X4/10, r.Y4/10)
}

// DeclareAreaType 区域类型，与 3.4.3 一致（项目内统一用数字）
const (
	DeclareAreaInvalid  = 0 // 无效/删除
	DeclareAreaCustom   = 1 // 自定义区域
	DeclareAreaBed      = 2 // 床（中心在雷达下时系统自动改为 5）
	DeclareAreaNoise    = 3 // 干扰区域
	DeclareAreaDoor     = 4 // 门
	DeclareAreaMonitor  = 5 // 监护床
)

// DeclareAreaData 设置区域单条，与 3.4.3 一致。传输格式 {area-id, area-type, x1, y1; x2, y2, x3, y3, x4, y4}
// AreaID 0-15；AreaType 0-5；顶点坐标单位 cm（规范），设备为 dm
type DeclareAreaData struct {
	AreaID   int `json:"area_id"`   // 0-15
	AreaType int `json:"area_type"` // 0-5，见 DeclareArea* 常量
	X1       int `json:"x1"`
	Y1       int `json:"y1"`
	X2       int `json:"x2"`
	Y2       int `json:"y2"`
	X3       int `json:"x3"`
	Y3       int `json:"y3"`
	X4       int `json:"x4"`
	Y4       int `json:"y4"`
}

// ParseDeclareAreaFromDM 从设备字符串（分米）解析单条区域，坐标 *10 填入结构体（cm）
// 设备可能返回：(1) 全逗号 {id,type,x1,y1,x2,y2,x3,y3,x4,y4} (2) 分号分隔 {id,type,x1,y1;x2,y2;x3,y3;x4,y4}。规范与 Redis 统一用全逗号。
func ParseDeclareAreaFromDM(s string) (*DeclareAreaData, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return nil, false
	}
	if s[0] == '{' {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == '}' || s[len(s)-1] == ',' || s[len(s)-1] == ';') {
		s = s[:len(s)-1]
	}
	// 先尝试全逗号格式：10 个值 id, type, x1, y1, x2, y2, x3, y3, x4, y4（设备读回常见）
	commaParts := strings.Split(s, ",")
	if len(commaParts) == 10 {
		var nums [10]int
		ok := true
		for i := 0; i < 10; i++ {
			n, err := strconv.Atoi(strings.TrimSpace(commaParts[i]))
			if err != nil {
				ok = false
				break
			}
			nums[i] = n
		}
		if ok && nums[0] >= 0 && nums[0] <= 15 && nums[1] >= 0 && nums[1] <= 5 {
			return &DeclareAreaData{
				AreaID: nums[0], AreaType: nums[1],
				X1: nums[2] * 10, Y1: nums[3] * 10,
				X2: nums[4] * 10, Y2: nums[5] * 10,
				X3: nums[6] * 10, Y3: nums[7] * 10,
				X4: nums[8] * 10, Y4: nums[9] * 10,
			}, true
		}
	}
	// 标准格式：分号分隔 4 段
	parts := strings.Split(s, ";")
	if len(parts) != 4 {
		return nil, false
	}
	seg0 := strings.Split(parts[0], ",")
	if len(seg0) != 4 {
		return nil, false
	}
	id, _ := strconv.Atoi(strings.TrimSpace(seg0[0]))
	typ, _ := strconv.Atoi(strings.TrimSpace(seg0[1]))
	x1, _ := strconv.Atoi(strings.TrimSpace(seg0[2]))
	y1, _ := strconv.Atoi(strings.TrimSpace(seg0[3]))
	d := &DeclareAreaData{AreaID: id, AreaType: typ, X1: x1 * 10, Y1: y1 * 10}
	for i, seg := range parts[1:] {
		xy := strings.Split(strings.TrimSpace(seg), ",")
		if len(xy) != 2 {
			return nil, false
		}
		x, _ := strconv.Atoi(strings.TrimSpace(xy[0]))
		y, _ := strconv.Atoi(strings.TrimSpace(xy[1]))
		x, y = x*10, y*10
		switch i {
		case 0:
			d.X2, d.Y2 = x, y
		case 1:
			d.X3, d.Y3 = x, y
		case 2:
			d.X4, d.Y4 = x, y
		}
	}
	return d, true
}

// String 输出规范格式（cm），全逗号，用于 Redis/API
func (d *DeclareAreaData) String() string {
	return fmt.Sprintf("{%d,%d,%d,%d,%d,%d,%d,%d,%d,%d}",
		d.AreaID, d.AreaType, d.X1, d.Y1, d.X2, d.Y2, d.X3, d.Y3, d.X4, d.Y4)
}

// ToDMString 输出设备格式（分米），全逗号，与规范一致
func (d *DeclareAreaData) ToDMString() string {
	return fmt.Sprintf("{%d,%d,%d,%d,%d,%d,%d,%d,%d,%d}",
		d.AreaID, d.AreaType, d.X1/10, d.Y1/10, d.X2/10, d.Y2/10, d.X3/10, d.Y3/10, d.X4/10, d.Y4/10)
}

// FallParamData 跌倒参数，与 Radar_MQTT_v3.0 3.4.4 fall_param 一致。16 字节 base64。
// 字节 0-2 保留；3 跌倒告警时间(×10秒)；4 功能开关 bit0 坐地 bit1 姿态 bit2 床上坐起；5 坐地告警时间(×10秒)；6-15 保留
type FallParamData struct {
	Reserved_0_2     [3]byte `json:"reserved_0_2"`
	FallDuration10s  int     `json:"fall_duration_10s"`  // 3: 0-250，实际秒=×10
	SittingEnabled   bool    `json:"sitting_enabled"`    // 4 bit0
	PostureEnabled   bool    `json:"posture_enabled"`    // 4 bit1
	BedSitUpEnabled  bool    `json:"bed_situp_enabled"`  // 4 bit2
	SittingDuration10s int     `json:"sitting_duration_10s"` // 5: 1-250，实际秒=×10
	Reserved_6_15     [10]byte `json:"reserved_6_15"`
}

// HeartBreathParamData 呼吸心率参数，与 Radar_MQTT_v3.0 3.4.5 heart_breath_param 一致。16 字节 base64。
// 字节 0 呼吸高阈值 1 心率高阈值 2 呼吸低 3 心率低；4 持续检测；5 生命体征弱时间(分钟)；6 灵敏度；7-15 预留
type HeartBreathParamData struct {
	RespRateMax      int     `json:"resp_rate_max"`       // 0: 次/分 [10,100]
	HeartRateMax      int     `json:"heart_rate_max"`      // 1: [10,255]
	RespRateMin      int     `json:"resp_rate_min"`       // 2: [1,20]
	HeartRateMin     int     `json:"heart_rate_min"`      // 3: [1,100]
	ContinuousDetect int     `json:"continuous_detect"`    // 4: 0/1
	VitalsWeakDurMin int     `json:"vitals_weak_dur_min"` // 5: 分钟 [1-15]
	VitalsWeakSens   int     `json:"vitals_weak_sens"`    // 6: [1,99]
	Reserved_7_15    [9]byte `json:"reserved_7_15"`
}
