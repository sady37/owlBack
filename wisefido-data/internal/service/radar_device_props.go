package service

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"owl-common/alarm"
)

// AlarmItemsToRadarDeviceProps 将雷达监控设置 AlarmItem[] 转为设备属性（仅工作模式+跌倒+呼吸心率）。
// 用于 radar-monitor-settings 页「工作模式、跌倒和呼吸心率参数」一齐提交，不涉及安装方式/高度/边界，不下发重启。
// 协议：3.4.1 radar_func_ctrl；3.4.4 fall_param；3.4.5 heart_breath_param。
func AlarmItemsToRadarDeviceProps(items []alarm.AlarmItem) (map[string]interface{}, error) {
	m := indexAlarmItemsByType(items)

	props := make(map[string]interface{})

	// 1. 工作模式 radar_func_ctrl (3/7/11/15)
	if v := modeFromItems(m); v >= 0 {
		props["radar_func_ctrl"] = v
	}

	// 2. 跌倒参数 fall_param (16-byte BASE64)
	if b, err := fallParamFromItems(m); err == nil && len(b) == 16 {
		props["fall_param"] = base64.StdEncoding.EncodeToString(b)
	} else if err != nil {
		return nil, fmt.Errorf("fall_param: %w", err)
	}

	// 3. 呼吸心率参数 heart_breath_param (16-byte BASE64)
	if b, err := heartBreathParamFromItems(m); err == nil && len(b) == 16 {
		props["heart_breath_param"] = base64.StdEncoding.EncodeToString(b)
	} else if err != nil {
		return nil, fmt.Errorf("heart_breath_param: %w", err)
	}

	return props, nil
}

func indexAlarmItemsByType(items []alarm.AlarmItem) map[string]alarm.AlarmItem {
	out := make(map[string]alarm.AlarmItem)
	for _, it := range items {
		if it.AlarmType != "" {
			out[it.AlarmType] = it
		}
	}
	return out
}

func modeFromItems(m map[string]alarm.AlarmItem) int {
	it, ok := m["Radar_MonitoringMode"]
	if !ok || it.AlarmParams == nil {
		return -1
	}
	switch v := it.AlarmParams["mode"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return -1
}

// fall_param: 16 bytes. Byte 3 = suspected_fall_duration/10 (0–250);
// byte 4 = bits: bit0 sit ground, bit1 posture, bit2 left bed;
// byte 5 = sitting_on_ground_duration/10 (1–250). Rest reserved.
// 注意：设备协议中时间单位为 10 秒，需要将 AlarmItem 中的 duration_sec（秒）除以 10 转换
func fallParamFromItems(m map[string]alarm.AlarmItem) ([]byte, error) {
	b := make([]byte, 16)

	// byte 3: 跌倒告警时间
	// AlarmItem 使用 duration_sec（秒），设备协议使用 10 秒单位
	// 转换：duration_sec / 10，默认 60s => 6 (10s单位)
	dur := 60
	if it, ok := m["Radar_Fall"]; ok && it.AlarmParams != nil {
		if v, ok := it.AlarmParams["duration_sec"]; ok {
			dur = int(num(v))
			if dur < 0 {
				dur = 0
			}
			if dur > 2500 {
				dur = 2500
			}
		}
	}
	b[3] = byte(dur / 10) // 转换为 10 秒单位

	// byte 4: bit0 坐地, bit1 姿态, bit2 床上坐起
	// byte 5: 坐地告警时间
	// AlarmItem 使用 duration_sec（秒），设备协议使用 10 秒单位
	// 转换：duration_sec / 10，默认 100s => 10 (10s单位)
	sitDur := 100
	if it, ok := m["Radar_SittingOnGround"]; ok {
		if enabled(it) {
			b[4] |= 1
			if it.AlarmParams != nil {
				if v, ok := it.AlarmParams["duration_sec"]; ok {
					d := int(num(v))
					if d >= 10 && d <= 2500 {
						sitDur = d
					}
				}
			}
		}
	}
	if it, ok := m["Radar_PostureDetection"]; ok && enabled(it) {
		b[4] |= 2
	}
	if it, ok := m["Radar_LeftBed"]; ok && enabled(it) {
		b[4] |= 4
	}
	if b[4]&(1|4) != 0 {
		b[4] |= 2
	}
	if sitDur < 10 {
		sitDur = 10
	}
	if sitDur > 2500 {
		sitDur = 2500
	}
	b[5] = byte(sitDur / 10) // 转换为 10 秒单位

	return b, nil
}

// heart_breath_param: 16 bytes. 0=upper breath, 1=upper heart, 2=lower breath, 3=lower heart;
// 4=intensive care 0/1, 5=weak vital duration (min), 6=sensitivity. 7–15 reserved.
// 注意：byte 5 的生命体征弱持续时间单位为分钟，与 AlarmItem 的 duration_min（分钟）一致，无需转换
func heartBreathParamFromItems(m map[string]alarm.AlarmItem) ([]byte, error) {
	b := make([]byte, 16)

	upperBreath, lowerBreath := 24, 8
	upperHeart, lowerHeart := 90, 50
	intensiveCare := 0
	weakDuration := 10
	weakSensitivity := 35

	if it, ok := m["Radar_AbnormalRespiratoryRate"]; ok && it.AlarmParams != nil {
		if v, ok := it.AlarmParams["max"]; ok {
			upperBreath = int(num(v))
		}
		if v, ok := it.AlarmParams["min"]; ok {
			lowerBreath = int(num(v))
		}
	}
	if it, ok := m["Radar_AbnormalHeartRate"]; ok && it.AlarmParams != nil {
		if v, ok := it.AlarmParams["max"]; ok {
			upperHeart = int(num(v))
		}
		if v, ok := it.AlarmParams["min"]; ok {
			lowerHeart = int(num(v))
		}
	}
	if it, ok := m["Radar_VitalsWeak"]; ok && it.AlarmParams != nil {
		// duration_min（分钟）直接使用，设备协议 byte 5 也是分钟单位
		if v, ok := it.AlarmParams["duration_min"]; ok {
			weakDuration = int(num(v))
		}
		if v, ok := it.AlarmParams["sensitivity"]; ok {
			weakSensitivity = int(num(v))
		}
	}

	clamp := func(v, lo, hi int) byte {
		if v < lo {
			v = lo
		}
		if v > hi {
			v = hi
		}
		return byte(v)
	}
	b[0] = clamp(upperBreath, 10, 100)
	b[1] = clamp(upperHeart, 10, 255)
	b[2] = clamp(lowerBreath, 1, 20)
	b[3] = clamp(lowerHeart, 1, 100)
	b[4] = byte(intensiveCare)
	b[5] = clamp(weakDuration, 1, 15)
	b[6] = clamp(weakSensitivity, 1, 99)

	return b, nil
}

// RadarDevicePropsToAlarmItems 从设备属性更新 AlarmItems（用于设备写入失败时同步实际值）
// 支持工作模式（radar_func_ctrl）、跌倒参数（fall_param）和呼吸心率参数（heart_breath_param）的反向解析
// 参考：wisefido-backend/wisefido-radar/modules/radar_service.go::GetDeviceAllAttributes
func RadarDevicePropsToAlarmItems(items []alarm.AlarmItem, deviceProps map[string]interface{}) ([]alarm.AlarmItem, error) {
	result := make([]alarm.AlarmItem, len(items))
	copy(result, items)

	// 更新工作模式 radar_func_ctrl
	if radarFuncCtrl, ok := deviceProps["radar_func_ctrl"]; ok {
		// 查找 Radar_MonitoringMode
		for i := range result {
			if result[i].AlarmType == alarm.RadarMonitoringMode {
				if result[i].AlarmParams == nil {
					result[i].AlarmParams = make(map[string]interface{})
				}
				// 转换类型
				var mode int
				switch v := radarFuncCtrl.(type) {
				case float64:
					mode = int(v)
				case int:
					mode = v
				case string:
					// 尝试解析字符串
					if parsed, err := strconv.Atoi(v); err == nil {
						mode = parsed
					} else {
						continue
					}
				default:
					continue
				}
				result[i].AlarmParams["mode"] = mode
				break
			}
		}
	}

	// 解析 fall_param（跌倒参数）
	if fallParam, ok := deviceProps["fall_param"]; ok {
		if fallParamStr, ok := fallParam.(string); ok {
			fallParamBytes, err := base64.StdEncoding.DecodeString(fallParamStr)
			if err == nil && len(fallParamBytes) >= 6 {
				// byte 3: 跌倒告警时间（10秒单位）→ Radar_Fall.duration_sec（秒）
				fallDuration := int(fallParamBytes[3]) * 10
				updateAlarmItemParam(&result, alarm.RadarFall, "duration_sec", fallDuration)

				// byte 4: bit0 坐地, bit1 姿态, bit2 床上坐起
				byte4 := fallParamBytes[4]
				// bit 0: 坐地告警使能 → Radar_SittingOnGround.is_enabled
				if byte4&1 != 0 {
					updateAlarmItemEnabled(&result, alarm.RadarSittingOnGround, 1)
					// byte 5: 坐地告警时间（10秒单位）→ Radar_SittingOnGround.duration_sec（秒）
					sitDuration := int(fallParamBytes[5]) * 10
					updateAlarmItemParam(&result, alarm.RadarSittingOnGround, "duration_sec", sitDuration)
				} else {
					updateAlarmItemEnabled(&result, alarm.RadarSittingOnGround, 0)
				}
				// bit 1: 姿态检测使能 → Radar_PostureDetection.is_enabled
				if (byte4>>1)&1 != 0 {
					updateAlarmItemEnabled(&result, alarm.RadarPostureDetection, 1)
				} else {
					updateAlarmItemEnabled(&result, alarm.RadarPostureDetection, 0)
				}
				// bit 2: 离床使能 → Radar_LeftBed.is_enabled
				if (byte4>>2)&1 != 0 {
					updateAlarmItemEnabled(&result, alarm.RadarLeftBed, 1)
				} else {
					updateAlarmItemEnabled(&result, alarm.RadarLeftBed, 0)
				}
			}
		}
	}

	// 解析 heart_breath_param（呼吸心率参数）
	if hrbrParam, ok := deviceProps["heart_breath_param"]; ok {
		if hrbrParamStr, ok := hrbrParam.(string); ok {
			hrbrParamBytes, err := base64.StdEncoding.DecodeString(hrbrParamStr)
			if err == nil && len(hrbrParamBytes) >= 7 {
				// byte 0: 呼吸率上限 → Radar_AbnormalRespiratoryRate.max
				updateAlarmItemParam(&result, alarm.RadarAbnormalRespiratoryRate, "max", int(hrbrParamBytes[0]))
				// byte 1: 心率上限 → Radar_AbnormalHeartRate.max
				updateAlarmItemParam(&result, alarm.RadarAbnormalHeartRate, "max", int(hrbrParamBytes[1]))
				// byte 2: 呼吸率下限 → Radar_AbnormalRespiratoryRate.min
				updateAlarmItemParam(&result, alarm.RadarAbnormalRespiratoryRate, "min", int(hrbrParamBytes[2]))
				// byte 3: 心率下限 → Radar_AbnormalHeartRate.min
				updateAlarmItemParam(&result, alarm.RadarAbnormalHeartRate, "min", int(hrbrParamBytes[3]))
				// byte 4: 重症监护（intensive care）- 在 AlarmItem 中无对应项，跳过
				// byte 5: 生命体征弱持续时间（分钟）→ Radar_VitalsWeak.duration_min（分钟，单位一致）
				updateAlarmItemParam(&result, alarm.RadarVitalsWeak, "duration_min", int(hrbrParamBytes[5]))
				// byte 6: 生命体征弱灵敏度 → Radar_VitalsWeak.sensitivity
				updateAlarmItemParam(&result, alarm.RadarVitalsWeak, "sensitivity", int(hrbrParamBytes[6]))
			}
		}
	}

	return result, nil
}

// updateAlarmItemParam 更新 AlarmItem 的参数值
func updateAlarmItemParam(items *[]alarm.AlarmItem, alarmType string, paramKey string, paramValue interface{}) {
	for i := range *items {
		if (*items)[i].AlarmType == alarmType {
			if (*items)[i].AlarmParams == nil {
				(*items)[i].AlarmParams = make(map[string]interface{})
			}
			(*items)[i].AlarmParams[paramKey] = paramValue
			break
		}
	}
}

// updateAlarmItemEnabled 更新 AlarmItem 的启用状态
func updateAlarmItemEnabled(items *[]alarm.AlarmItem, alarmType string, enabled int) {
	for i := range *items {
		if (*items)[i].AlarmType == alarmType {
			(*items)[i].IsEnabled = &enabled
			break
		}
	}
}

func enabled(it alarm.AlarmItem) bool {
	if it.IsEnabled != nil && *it.IsEnabled == 1 {
		return true
	}
	return false
}

func num(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}

// V1ConfigToRadarDeviceProps 将 v1 / MQTT 风格配置转为 Radar 设备属性。
// 入参与 radarMqttConfig 输出对齐：dm（画布 cm/10），install_model 0/1/2 或 "ceiling"/"wall"/"corn"，
// boundary_left/right/front/rear，area_{i}_id、area_{i}_type、area_{i}_x1..y4（dm）；area_{i}_id=-1 表示删除。
// 出参：radar_install_style, radar_install_height, rectangle, declare_area。
func V1ConfigToRadarDeviceProps(config map[string]interface{}) map[string]interface{} {
	properties := make(map[string]interface{})

	installModelStr := resolveInstallModelString(config)

	// 安装方式 → radar_install_style (0=顶装, 1=侧装, 2=角装)，仅当 config 含 install_model 时写入
	if _, has := config["install_model"]; has {
		switch installModelStr {
		case "ceiling":
			properties["radar_install_style"] = "0"
		case "wall":
			properties["radar_install_style"] = "1"
		case "corn":
			properties["radar_install_style"] = "2"
		}
	}

	// 安装高度：height 已为 dm → radar_install_height
	if _, ok := config["height"]; ok {
		properties["radar_install_height"] = fmt.Sprintf("%.0f", num(config["height"]))
	}

	// 边界：boundary_* 已为 dm → rectangle
	// 3.8.2 自动测量：-127 开始 127 结束，不下发 buildRectangle，直接下发特殊 rectangle
	if _, hasL := config["boundary_left"]; hasL {
		if _, hasR := config["boundary_right"]; hasR {
			if _, hasF := config["boundary_front"]; hasF {
				if _, hasB := config["boundary_rear"]; hasB {
					left := num(config["boundary_left"])
					right := num(config["boundary_right"])
					front := num(config["boundary_front"])
					rear := num(config["boundary_rear"])
					li, ri, fi, rei := int(left), int(right), int(front), int(rear)
					if li == -127 && ri == -127 && fi == -127 && rei == -127 {
						properties["rectangle"] = "{-127,-127;-127,-127;-127,-127;-127,-127}"
					} else if li == 127 && ri == 127 && fi == 127 && rei == 127 {
						properties["rectangle"] = "{127,127;127,127;127,127;127,127}"
					} else {
						rectangle := buildRectangle(left, right, front, rear, installModelStr)
						properties["rectangle"] = rectangle
					}
				}
			}
		}
	}

	// 区域：area_{i}_id/type/x1..y4 (dm) → declare_area；area_{i}_id=-1 表示删除，不加入
	if decl := buildDeclareAreaFromConfig(config); len(decl) > 0 {
		properties["declare_area"] = decl
	}

	// 3.4.6 角度校准：run_Horizontal（HC2 支持，TK2 不支持），数字均为字符型
	if v, ok := config["run_Horizontal"]; ok && v != nil {
		properties["run_Horizontal"] = toStr(v)
	}

	// 202405 读/写：ip_port、ssid_password；只写：voice_end_tip；均为字符型
	if v, ok := config["ip_port"]; ok && v != nil {
		properties["ip_port"] = toStr(v)
	}
	if v, ok := config["ssid_password"]; ok && v != nil {
		properties["ssid_password"] = toStr(v)
	}
	if v, ok := config["voice_end_tip"]; ok && v != nil {
		properties["voice_end_tip"] = toStr(v)
	}

	return properties
}

// toStr 将 interface{} 转为字符串，数字均为字符型
func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// resolveInstallModelString 从 config 解析 install_model，返回 "ceiling"|"wall"|"corn"。
// 支持字符串或 radarMqttConfig 的 0/1/2。
func resolveInstallModelString(config map[string]interface{}) string {
	if s, ok := config["install_model"].(string); ok {
		switch s {
		case "ceiling", "wall", "corn":
			return s
		}
	}
	n := num(config["install_model"])
	switch int(n) {
	case 0:
		return "ceiling"
	case 1:
		return "wall"
	case 2:
		return "corn"
	}
	return "wall"
}

// buildDeclareAreaFromConfig 从 area_{i}_id/type/x1..y4 构建 declare_area（dm）。
// 与 radarMqttConfig 的 area_{id}_* 命名一致；area_{i}_id=-1 表示删除，不加入。
func buildDeclareAreaFromConfig(config map[string]interface{}) []interface{} {
	var out []interface{}
	for i := 0; i < 16; i++ {
		pre := fmt.Sprintf("area_%d_", i)
		idVal := config[pre+"id"]
		if idVal == nil {
			continue
		}
		id := int(num(idVal))
		if id < 0 {
			continue
		}
		areaType := ""
		if v := config[pre+"type"]; v != nil {
			if s, ok := v.(string); ok {
				areaType = s
			} else {
				areaType = fmt.Sprintf("%v", v)
			}
		}
		el := map[string]interface{}{
			"id": id, "type": areaType,
			"x1": int(num(config[pre+"x1"])), "y1": int(num(config[pre+"y1"])),
			"x2": int(num(config[pre+"x2"])), "y2": int(num(config[pre+"y2"])),
			"x3": int(num(config[pre+"x3"])), "y3": int(num(config[pre+"y3"])),
			"x4": int(num(config[pre+"x4"])), "y4": int(num(config[pre+"y4"])),
		}
		out = append(out, el)
	}
	return out
}

// buildRectangle 构建边界矩形坐标字符串
// 格式：{x1, y1; x2, y2, x3, y3, x4, y4}，单位 dm。
// 顶装(ceiling)：X 轴 ±left/right，Y 轴 ±front/rear；侧装(wall)、角装(corn)：Y 从 0 到 rear。
func buildRectangle(left, right, front, rear float64, installModel string) string {
	if installModel == "ceiling" {
		return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
			int(-left), int(-front),
			int(right), int(-front),
			int(-left), int(rear),
			int(right), int(rear))
	}
	// 侧装、角装
	return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
		int(-left), 0,
		int(right), 0,
		int(-left), int(rear),
		int(right), int(rear))
}
