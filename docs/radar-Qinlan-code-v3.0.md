## reamtime track
```json
type RadarPosition struct {
    Id            int    `json:"-" xorm:"pk autoincr"`           // 主键，不序列化到JSON
    DeviceCode    string `json:"deviceCode,omitempty"`           // 设备编码
    PersonIndex   int    `json:"personIndex"`                    // 人员索引（88表示无人）
    CoodinateX    int    `json:"coodinateX"`                     // X坐标
    CoodinateY    int    `json:"coodinateY"`                     // Y坐标
    CoodinateZ    int    `json:"coodinateZ"`                     // Z坐标（高度）
    RemainingTime int    `json:"remainingTime"`                  // 剩余时间
    Posture       int    `json:"posture"`                        // 姿势（0-6）
    Event         int    `json:"event"`                          // 事件类型
    AreaId        int    `json:"areaId"`                         // 区域ID
    Timestamp     int64  `json:"timestamp,omitempty"`            // 时间戳
}

type RadarPosition struct {
	Id            int    `json:"-" xorm:"pk autoincr"`           // 主键
	DeviceCode    string `json:"deviceCode,omitempty"`           // 设备编码
	PersonIndex   int    `json:"personIndex"`                    // 人员索引（0-7 或 无人88）
	CoodinateX    int    `json:"coodinateX"`                     // X坐标 (分米, 有符号 -127-127)
	CoodinateY    int    `json:"coodinateY"`                     // Y坐标 (分米, 有符号-127-127)
	CoodinateZ    int    `json:"coodinateZ"`                     // Z坐标 (厘米, 无符号0-255)
	RemainingTime int    `json:"remainingTime"`                  // 剩余时间 (0-60秒)
	Posture       int    `json:"posture"`                        // 人员姿态 (0-11)
	Event         int    `json:"event"`                          // 人员事件 (0-4)
	AreaId        int    `json:"areaId"`                         // 区域ID
	Timestamp     int64  `json:"timestamp,omitempty"`            // 时间戳
}


func DecodeRadarTrack(buf []byte, deviceCode string) []*RadarPosition {
	const step = 16
	// 校验长度是否为 16 的倍数
	if len(buf) == 0 || len(buf)%step != 0 {
		return nil
	}

	n := len(buf) / step
	results := make([]*RadarPosition, 0, n)
	now := time.Now().Unix()

	for i := 0; i < n; i++ {
		offset := i * step
		data := buf[offset : offset+step]

		// 1. 处理坐标 (X, Y 为有符号 int8)
		// Go 默认 uint8 强转 int 会保持正数，需要通过 int8 强制解析负数
		x := int(int8(data[1]))
		y := int(int8(data[2]))
		z := int(data[3]) // Z 是无符号 0-255

		// 2. 处理人员事件 (过滤不在定义内的事件值)
		event := int(data[14])
		if event > 4 {
			event = 0 // 协议建议忽略，这里归零处理
		}

		pos := &RadarPosition{
			DeviceCode:    deviceCode,
			PersonIndex:   int(data[0]),
			CoodinateX:    x,
			CoodinateY:    y,
			CoodinateZ:    z,
			RemainingTime: int(data[12]),
			Posture:       int(data[13]),
			Event:         event,
			AreaId:        int(data[15]),
			Timestamp:     now,
		}

		results = append(results, pos)
	}

	return results
}

```


## realtime vital
``` json 
type RadarHeartBreath struct {
    Id           int    `json:"-" xorm:"pk autoincr"`      // 主键
    DeviceCode   string `json:"deviceCode,omitempty"`      // 设备编码
    BreathRate   int    `json:"breathRate"`                // 实时呼吸率（byte1）
    HeartRate    int    `json:"heartRate"`                 // 实时心率（byte2)
    SleepStage   int    `json:"sleepStage"`                // 睡眠阶段（byte13 bit7-6）
    Stability    int    `json:"stability"`                 // 稳定度（byte14 bit1-0，>=202501 固件）
    Rsv3_12      []byte `json:"rsv3_12,omitempty" xorm:"-"`// byte3~12 原始保留字段（长度10）
    Rsv15        int    `json:"rsv15,omitempty"`           // byte15 保留
    Timestamp    int64  `json:"timestamp,omitempty"`       // 时间戳
}

// 解码示例（16字节 bh）
func DecodeRadarHeartBreath(buf []byte) *RadarHeartBreath {
    if len(buf) != 16 {
        return nil
    }

    return &RadarHeartBreath{
        BreathRate: buf[1],
        HeartRate:  buf[2],
        SleepStage: (buf[13] >> 6) & 0x03,
        Stability:  buf[14] & 0x03,
        Rsv3_12:    append([]byte{}, buf[3:13]...),
        Rsv15:      int(buf[15]),
    }
}


```





## Statistics sleep 

``` json

type RadarHeartBreathStatistics struct {
    Id                int    `json:"id" xorm:"pk autoincr"`      // 主键
    DeviceCode        string `json:"deviceCode"`                 // 设备编码
    BreathRate        int    `json:"breathRate"`                 // 当前呼吸率
    HeartRate         int    `json:"heartRate"`                  // 当前心率
    AverageBreathRate int    `json:"averageBreathRate"`          // 平均呼吸率
    AverageHeartRate  int    `json:"averageHeartRate"`           // 平均心率
    BreathStatus      int    `json:"breathStatus"`               // 呼吸状态
    HeartStatus       int    `json:"heartStatus"`                // 心脏状态
    Vitals            int    `json:"vitals"`                     // 生命体征状态
    SleepStage        int    `json:"sleepStage"`                 // 睡眠阶段
    Timestamp         int64  `json:"timestamp"`                  // 时间戳
}

package decoder

import ("time")

type RadarHeartBreathStatistics struct {
	Id                int    `json:"id" xorm:"pk autoincr"`      // 主键
	DeviceCode        string `json:"deviceCode"`                 // 设备编码
	BreathRate        int    `json:"breathRate"`                 // 当前呼吸率
	HeartRate         int    `json:"heartRate"`                  // 当前心率
	AverageBreathRate int    `json:"averageBreathRate"`          // 平均呼吸率
	AverageHeartRate  int    `json:"averageHeartRate"`           // 平均心率
	BreathStatus      int    `json:"breathStatus"`               // 呼吸状态 (0-3)
	HeartStatus       int    `json:"heartStatus"`                // 心脏状态 (0-3)
	Vitals            int    `json:"vitals"`                     // 生命体征状态 (0-3)
	SleepStage        int    `json:"sleepStage"`                 // 睡眠阶段 (0-3)
	Timestamp         int64  `json:"timestamp"`                  // 时间戳
}

func DecodeRadarSleep(buf []byte, deviceCode string) *RadarHeartBreathStatistics {
	// 1. 校验长度与标识符 (0xff)
	if len(buf) != 16 || buf[0] != 0xff {
		return nil
	}

	// 2. 提取第 13 字节进行位运算
	// 字节二进制结构: [SleepStage(7-6) | Vitals(5-4) | HeartStatus(3-2) | BreathStatus(1-0)]
	eventByte := buf[13]

	return &RadarHeartBreathStatistics{
		DeviceCode:        deviceCode,
		BreathRate:        int(buf[1]),
		HeartRate:         int(buf[2]),
		AverageBreathRate: int(buf[5]),
		AverageHeartRate:  int(buf[6]),
		
		// 位运算解析
		BreathStatus: int(eventByte & 0x03),          // 取 bit 0, 1
		HeartStatus:  int((eventByte >> 2) & 0x03),   // 右移 2 位，取 bit 2, 3
		Vitals:       int((eventByte >> 4) & 0x03),   // 右移 4 位，取 bit 4, 5
		SleepStage:   int((eventByte >> 6) & 0x03),   // 右移 6 位，取 bit 6, 7

		Timestamp: time.Now().Unix(),
	}
}

2. 第 13 字节位分布图示

位 (Bit),作用域,二进制掩码,对应字段
7 & 6,睡眠状态,(b >> 6) & 0x03,"0:未定义, 1:浅睡, 2:深睡, 3:清醒"
5 & 4,生命体征,(b >> 4) & 0x03,"0:正常, 3:弱 (1, 2未定义)"
3 & 2,心率状态,(b >> 2) & 0x03,"0:正常, 1:过低, 2:过高"
1 & 0,呼吸状态,b & 0x03,"0:正常, 1:过低, 2:过高, 3:暂停"

```


## Statistics  track
```json

type RadarPositionStatistics struct {
    Id                     int    `json:"id" xorm:"pk autoincr"` // 主键
    DeviceCode             string `json:"deviceCode"`            // 设备编码
    Version                int    `json:"version"`               // 版本
    PersonCount            int    `json:"personCount"`           // 人员数量
    MoveDistance           int    `json:"moveDistance"`          // 移动距离
    MoveDuration           int    `json:"moveDuration"`          // 移动时长
    SitDuration            int    `json:"sitDuration"`           // 坐姿时长
    InBedDuration          int    `json:"inBedDuration"`         // 在床时长
    StandDuration          int    `json:"standDuration"`         // 站立时长
    MultiplePersonDuration int    `json:"multiplePersonDuration"`// 多人时长
    MeasuringBreathHeart   int    `json:"measuringBreathHeart"`  // 测量心率和呼吸时长
    Timestamp              int64  `json:"timestamp"`             // 时间戳
}

package decoder

import (
	"encoding/binary"
	"time"
)

type RadarPositionStatistics struct {
	Id                     int    `json:"id" xorm:"pk autoincr"` // 主键
	DeviceCode             string `json:"deviceCode"`            // 设备编码
	Version                int    `json:"version"`               // 版本
	PersonCount            int    `json:"personCount"`           // 人员数量
	MoveDistance           int    `json:"moveDistance"`          // 移动距离 (米)
	MoveDuration           int    `json:"moveDuration"`          // 移动时长 (秒)
	SitDuration            int    `json:"sitDuration"`           // 坐姿时长 (未开放)
	InBedDuration          int    `json:"inBedDuration"`         // 卧床时长 (秒)
	StandDuration          int    `json:"standDuration"`         // 站立时长 (秒)
	MultiplePersonDuration int    `json:"multiplePersonDuration"`// 多人时长 (秒)
	Timestamp              int64  `json:"timestamp"`             // 时间戳
}

func DecodeRadarTrackStatistics(buf []byte, deviceCode string) *RadarPositionStatistics {
	// 校验长度是否为 16 字节
	if len(buf) != 16 {
		return nil
	}

	// 字节 2~3 是行走距离，采用 Big-Endian (大端序)
	// 例如：buf[2]=0x01, buf[3]=0x02 -> 0x0102 = 258
	moveDist := binary.BigEndian.Uint16(buf[2:4])

	return &RadarPositionStatistics{
		DeviceCode:             deviceCode,
		Version:                int(buf[0]),
		PersonCount:            int(buf[1]),
		MoveDistance:           int(moveDist),
		MoveDuration:           int(buf[4]),
		SitDuration:            int(buf[5]), // 协议标注未开放，通常为0
		InBedDuration:          int(buf[6]),
		StandDuration:          int(buf[7]),
		MultiplePersonDuration: int(buf[8]),
		Timestamp:              time.Now().Unix(),
	}
}

2. 核心技术点解析
大端序 (Big-Endian) 处理
在处理第 2~3 字节的 行走距离 时，协议明确要求使用 Big-Endian。

逻辑：高位字节在前（buf[2]），低位字节在后（buf[3]）。

代码：使用 binary.BigEndian.Uint16(buf[2:4]) 是最标准且安全的方法。如果手动计算，等同于 int(buf[2])<<8 | int(buf[3])。

数据统计逻辑注意点
多人冲突：根据协议，当 MultiplePersonDuration (字节 8) 有数值时，该分钟内的行走距离、行走时长、静止时长（卧床/站立）通常是不做统计或不准确的。在做数据报表时，建议先判断人员数量或多人时长。
单位一致性：
距离：单位是米。
时长：单位都是秒。因为这是分钟级上报，所以这些时长的总和理论上不应超过 60 秒。

3. 字段映射总结表

字节,结构体字段,解析方式,业务备注
0,Version,buf[0],1 或 2
1,PersonCount,buf[1],0-8 人
2-3,MoveDistance,Uint16 (BE),过去 1 分钟总和
4,MoveDuration,uint8,过去 1 分钟总和
6,InBedDuration,uint8,过去 1 分钟总和
7,StandDuration,uint8,过去 1 分钟总和
8,MultiplePersonDuration,uint8,触发时其他统计项失效

```


## category 自动映射
针对 iot_timeseries 与 alarm_events 的分发逻辑

```json
func GetFHIRCategory(dataType string, isAlarm bool) string {
    if isAlarm {
        switch dataType {
        case "Fall", "Stay", "NoActivity24h":
            return "safety"
        case "HeartAbnormal", "Apnea", "PressureUlcerRisk": // 2H不翻身
            return "clinical"
        case "Offline", "LowBattery":
            return "device"
        default:
            return "behavioral"
        }
    } else {
        // 原始观测数据
        switch dataType {
        case "HeartRate", "BreathRate":
            return "vital-signs"
        case "SleepStage", "Posture", "Movement":
            return "activity"
        default:
            return "activity"
        }
    }
}



// 基础告警结构（对应 alarm_events 表）
type AlarmEvent struct {
    DeviceCode   string `json:"deviceCode"`
    Category     string `json:"category"`     // FHIR Category: clinical, behavioral, safety, device
    ResourceType string `json:"resourceType"` // Flag
    AlarmType    string `json:"alarmType"`    // 如: Fall, HeartAbnormal
    DangerLevel  string `json:"dangerLevel"`  // EMERGENCY, WARNING, etc.
    Timestamp    int64  `json:"timestamp"`
}

// 基础观测结构（对应 iot_timeseries 表）
type ObservationData struct {
    DeviceCode   string      `json:"deviceCode"`
    Category     string      `json:"category"`     // FHIR Category: vital-signs, activity
    ResourceType string      `json:"resourceType"` // Observation
    Value        interface{} `json:"value"`        // 原始数值或状态
    Timestamp    int64       `json:"timestamp"`
}


func GetFHIRMapping(alarmType string, isAlarm bool) (category string, resource string) {
    if isAlarm {
        resource = "Flag"
        switch alarmType {
        // 安全类：生命威胁与空间安全
        case "Fall", "SuspectedFall", "Stay", "NoActivity24h":
            category = "safety"
            
        // 临床类：生理参数越限
        case "AbnormalHeartRate", "AbnormalRespiratoryRate", "ApneaHypopnea", "VitalsWeak":
            category = "clinical"
            
        // 行为健康类：长期习惯异常
        case "AbnormalBodyMovement", "LongTimeNoTurnOver":
            category = "behavioral"
            
        // 设备类
        case "OfflineAlarm", "LowBattery", "DeviceFailure":
            category = "device"
            
        default:
            category = "behavioral"
        }
    } else {
        resource = "Observation"
        switch alarmType {
        // 基础体征测量
        case "HeartRate", "BreathRate":
            category = "vital-signs"
            
        // 行为与姿态监测
        case "SleepState", "Posture", "MoveDistance", "InBedStatus":
            category = "activity"
            
        default:
            category = "activity"
        }
    }
    return
}


func ProcessSleepData(buf []byte, deviceCode string) ([]ObservationData, []AlarmEvent) {
    // 假设调用之前的 DecodeRadarSleep 解析出原始数据
    raw := DecodeRadarSleep(buf, deviceCode)
    if raw == nil { return nil, nil }

    var observations []ObservationData
    var alarms []AlarmEvent

    // 1. 生成观测数据 (vital-signs)
    cat, res := GetFHIRMapping("HeartRate", false)
    observations = append(observations, ObservationData{
        DeviceCode: deviceCode,
        Category: cat, ResourceType: res,
        Value: raw.AverageHeartRate, Timestamp: raw.Timestamp,
    })

    // 2. 判定并生成告警 (clinical)
    if raw.BreathStatus == 3 { // 呼吸暂停
        catA, resA := GetFHIRMapping("ApneaHypopnea", true)
        alarms = append(alarms, AlarmEvent{
            DeviceCode: deviceCode,
            Category: catA, ResourceType: resA,
            AlarmType: "Radar_ApneaHypopnea",
            DangerLevel: "EMERGENCY",
            Timestamp: raw.Timestamp,
        })
    }
    
    return observations, alarms
}



## Radar 报警类型 

package alarm

// 通用报警类型
const (
    OfflineAlarm    = "OfflineAlarm"
    LowBattery      = "LowBattery"
    DeviceFailure   = "DeviceFailure"
)

// Radar 报警类型
const (
    RadarFall                = "Fall"
    RadarSuspectedFall       = "SuspectedFall"
    RadarApneaHypopnea       = "Radar_ApneaHypopnea"
    RadarAbnormalHeartRate   = "Radar_AbnormalHeartRate"
    RadarAbnormalRespiratoryRate = "Radar_AbnormalRespiratoryRate"
    RadarLeftBed             = "Radar_LeftBed"
    RadarVitalsWeak          = "VitalsWeak"
    RadarStay                = "Stay"
    RadarNoActivity24h       = "NoActivity24h"
    RadarWarningArea         = "WarningArea"

    isOffline                ="OfflineAlarm"   //掉线    
    RadarSigalPoor           = "SignalPoor"    //雷达信号弱，非wifi
    RadarAngleException      = "AngleException"
    RadarSittingOnGround     = "SittingOnGround"
    DeviceFailure            ="DeviceFailure"
    Unkonw                   ="Unkonw"
)

// SleepPad 报警类型
const (

    SleepPadApneaHypopnea       = "SleepPad_ApneaHypopnea"               //呼吸暂停
    SleepPadAbnormalHeartRate   = "SleepPad_AbnormalHeartRate"          //异常心率
    SleepPadAbnormalRespiratoryRate = "SleepPad_AbnormalRespiratoryRate"  //异常呼吸率

    SleepPadOnBed               = "SleepPad_OnBed"                        //在床报警  
    SleepPadLeftBed             = "SleepPad_LeftBed"                     //离床报警  
    SleepPadSitUp               = "SleepPad_SitUp"                      //疑似床上坐起报警    

    SleepPadAbnormalBodyMovement = "SleepPad_AbnormalBodyMovement"                  //频繁体动报警   
    SleepPadNoBodymove	            =“”         //无体动报警
    SleepPadNoTurnOver             =“”         //久未翻身报警

    isOffline                   ="OfflineAlarm"   //掉线
    alarmSensorFall	            =“”         //传感器脱落报警
    DeviceFailure            ="DeviceFailure"
    Unkonw                   ="Unkonw"
)

// GetAlarmTypesByDeviceType 根据设备类型获取报警类型列表
func GetAlarmTypesByDeviceType(deviceType string) []string {
    // ...
}

## sleep 报警类型
类型	描述
alarmSensorFall	传感器脱落报警（BM8701-2、M901L支持，SDC100、M8701W、Z400TWP-3不支持）
alarmLeftBed	离床报警
alarmHeartRateFast	心率过速报警
alarmHeartRateSlow	心率过缓报警
alarmBreathRateFast	呼吸过速报警
alarmBreathRateSlow	呼吸过缓报警
alarmBreathRatePause	呼吸暂停报警
alarmBodymove	频繁体动报警
alarmNoBodymove	无体动报警
alarmNoTurnOver	久未翻身报警
alarmSitup	疑似坐起报警（BM8701-2/M901L硬板传感器、M8701W支持，SDC100、BM8701-2/M901L压电传感器、Z400TWP-3不支持）
alarmOnBed	在床报警


获取用户的报警配置信息
http(s)://domain{:port}/sleepace/getalarmnotifyconfig
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx"
}
}

参数说明
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
userid	string	合作方用户的唯一标识

响应
{
"status": 0,
"msg": null,
"data": {
"fallFlag": 0,
"leftBedFlag": 1,
"leftBedDuration": 10,
"leftBedStartHour": 17,
"leftBedStartMinute": 0,
"leftBedEndHour": 21,
"leftBedEndMinute": 0,
"heartRateFastFlag": 0,
"heartRateFastDuration": 600,
"maxHeartRate": 120,
"heartRateSlowFlag": 0,
"heartRateSlowDuration": 1200,
"minHeartRate": 45,
"breathRateFastFlag": 0,
"breathRateFastDuration": 1200,
"maxBreathRate": 26,
"breathRateSlowFlag": 0,
"breathRateSlowDuration": 1200,
"minBreathRate": 10,
"breathPauseFlag": 0,
"breathPauseDuration": 60,
"bodyMoveFlag": 0,
"bodyMoveDuration": 10,
"nobodyMoveFlag": 0,
"nobodyMoveDuration": 60,
"noTurnOverFlag":0,
"noTurnOverDuration":60,
"situpFlag": 0,
"onbedFlag": 0,
"onbedDuration": 600
}
}

字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
fallFlag	int	传感器脱落报警开关，0关1开
leftBedFlag	int	离床报警开关，0关1开
leftBedDuration	int	离床时长，单位秒
leftBedStartHour	int	离床报警范围，开始小时[0,23]
leftBedStartMinute	int	离床报警范围，开始分钟[0,59]
leftBedEndHour	int	离床报警范围，结束小时[0,23]
leftBedEndMinute	int	离床报警范围，结束分钟[0,59]
heartRateFastFlag	int	心率过速报警开关，0关 1开
heartRateFastDuration	int	心率过速持续时长，单位秒
maxHeartRate	int	最大心率值，>=该值触发报警
heartRateSlowFlag	int	心率过缓报警开关，0关 1开
heartRateSlowDuration	int	心率过缓持续时长，单位秒
minHeartRate	int	最小心率值，<=该值触发报警
breathRateFastFlag	int	呼吸过速报警开关，0关 1开
breathRateFastDuration	int	呼吸过速持续时长，单位秒
maxBreathRate	int	最大呼吸值，>=该值触发报警
breathRateSlowFlag	int	呼吸过缓报警开关，0关 1开
breathRateSlowDuration	int	呼吸过缓持续时长，单位秒
minBreathRate	int	最小呼吸值，<=该值触发报警
breathPauseFlag	int	呼吸暂停报警开关，0关 1开
breathPauseDuration	int	呼吸暂停持续时长，单位秒
bodyMoveFlag	int	频繁体动报警开关，0关 1开
bodyMoveDuration	int	频繁体动持续时长，单位分钟，（时长必须是实时数据上报间隔的倍数，如：实时数据上报间隔设置为2分钟，则体动持续时长可以设置为：2分钟，4分钟，6分钟，... ...）
nobodyMoveFlag	int	无体动报警开关，0关 1开
nobodyMoveDuration	int	无体动持续时长，单位分钟
（时长必须是实时数据上报间隔的倍数，如：实时数据上报间隔设置为2分钟，则体动持续时长可以设置为：2分钟，4分钟，6分钟，... ...）
noTurnOverFlag	int	无翻身报警开关，0关 1开
noTurnOverDuration	int	无翻身持续时长，单位分钟
（时长必须是实时数据上报间隔的倍数，如：实时数据上报间隔设置为2分钟，则体动持续时长可以设置为：2分钟，4分钟，6分钟，... ...）
situpFlag	int	坐起报警开关，0关 1开
onbedFlag	int	在床报警开关，0关 1开
onbedDuration	int	在床持续时长，单位秒


	AlarmTypeOfflineAlarm  =  掉线
	AlarmTypeDeviceFailure = "DeviceFailure"
	AlarmTypeUnknown       = "Unknown"

	SleepPadApneaHypopnea           = "SleepPad_ApneaHypopnea"
	SleepPadAbnormalHeartRate       = "SleepPad_AbnormalHeartRate"
	SleepPadAbnormalRespiratoryRate = "SleepPad_AbnormalRespiratoryRate"
	SleepPadLeftBed                 = "SleepPad_LeftBed"
	SleepPadLeftBedTooLong          = "SleepPad_LeftBedTooLong"
	SleepPadnBed                   = "SleepPad_InBed"
	SleepPadBedSitUp                = "SleepPad_BedSitUp"
	SleepPadAbnormalBodyMovement    = "SleepPad_AbnormalBodyMovement"
	SleepPadNoBodyMove              = "SleepPad_NoBodyMove"
	SleepPadNoTurnOver              = "SleepPad_NoTurnOver"
	SleepPadResetTime               = "SleepPad_ResetTime"
	SleepPadNapTime                 = "SleepPad_NapTime"
	SleepPadSensorDetached          = "SleepPad_SensorDetached"


    ## sleep 报警类型
类型	描述
alarmLeftBed	离床报警        :SleepPadLeftBed或SleepPadLeftBedTooLong 均可
alarmHeartRateFast	心率过速报警 : SleepPadAbnormalHeartRate
alarmHeartRateSlow	心率过缓报警 : SleepPadAbnormalHeartRate
alarmBreathRateFast	呼吸过速报警 : SleepPadAbnormalRespiratoryRate
alarmBreathRateSlow	呼吸过缓报警 : SleepPadAbnormalRespiratoryRate
alarmBreathRatePause	呼吸暂停报警:SleepPadApneaHypopnea
alarmBodymove	频繁体动报警:SleepPadAbnormalBodyMovement
alarmNoBodymove	无体动报警:SleepPadNoBodyMove
alarmNoTurnOver	久未翻身报警:SleepPadNoTurnOver
alarmSitup	疑似坐起报警:SleepPadBedSitUp
alarmOnBed	在床报警:SleepPadOnBed
connectionStatus:  offLine
alarmSensorFall	传感器脱落报警  : SleepPadSensorDetached

Radar MQTT
Event类：
根据事件列表：
进出事件Event_type=1
进出房间                   ： RadarStay
进出区域+Area_type={2||5}  ：RadarInBed,RadarLeftBedTooLong,RadarnBed
进入区域+Aarea_type=6      : RadarWarningArea

姿态变化Pose=2
5-确认跌倒             : RadarFall
2-疑似跌倒             : RadarSuspectedFall
7-疑似坐地 8-确认坐地：  ：RadarSittingOnGround
10-疑似床上坐起 11-确认床上坐起 ： RadarBedSitUp

event_type=7
信号差事件"signal_poor" : RadarSignalPoor
event_type=8
倾角异常事件"angle_abnormal" : RadarAngleException

统计类statsitsc
sleep
01: 呼吸过低 10: 呼吸过高  ：RadarAbnormalRespiratoryRate
01: 心率过低 10: 心率过高  ：RadarAbnormalHeartRate
11: 呼吸暂停              ：RadarApneaHypopnea
11: 生命体征弱            ：RadarVitalsWeak











