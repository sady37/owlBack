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
func fallParamFromItems(m map[string]alarm.AlarmItem) ([]byte, error) {
	b := make([]byte, 16)

	// byte 3: 跌倒告警时间 (10s 单位), 默认 6 => 60s
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
	b[3] = byte(dur / 10)

	// byte 4: bit0 坐地, bit1 姿态, bit2 床上坐起
	// byte 5: 坐地告警时间 (10s 单位), 默认 10 => 100s
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
	b[5] = byte(sitDur / 10)

	return b, nil
}

// heart_breath_param: 16 bytes. 0=upper breath, 1=upper heart, 2=lower breath, 3=lower heart;
// 4=intensive care 0/1, 5=weak vital duration (min), 6=sensitivity. 7–15 reserved.
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

// V1ConfigToRadarDeviceProps 将 v1.0 配置格式转换为 Radar 设备属性格式。
// v1.0 格式：install_model, height, boundary_left, boundary_right, boundary_front, boundary_rear, area_*_*
// Radar 属性格式：radar_install_style, radar_install_height, rectangle, declare_area
// 用于 RadarHandler.UpdateConfig（安装方式/高度/边界配置，可能触发重启）。
func V1ConfigToRadarDeviceProps(config map[string]interface{}) map[string]interface{} {
	properties := make(map[string]interface{})

	// 安装方式：install_model (wall/ceiling) → radar_install_style (0=顶装, 1=侧装)
	if installModel, ok := config["install_model"].(string); ok {
		if installModel == "ceiling" {
			properties["radar_install_style"] = "0"
		} else if installModel == "wall" {
			properties["radar_install_style"] = "1"
		}
	}

	// 安装高度：height (dm) → radar_install_height
	if height, ok := config["height"].(float64); ok {
		properties["radar_install_height"] = fmt.Sprintf("%.0f", height)
	}

	// 检测边界：boundary_left, boundary_right, boundary_front, boundary_rear → rectangle
	// 格式：{x1, y1; x2, y2, x3, y3, x4, y4}
	if boundaryLeft, ok := config["boundary_left"].(float64); ok {
		if boundaryRight, ok := config["boundary_right"].(float64); ok {
			if boundaryFront, ok := config["boundary_front"].(float64); ok {
				if boundaryRear, ok := config["boundary_rear"].(float64); ok {
					// 根据安装方式构建边界坐标
					installModel := "ceiling"
					if im, ok := config["install_model"].(string); ok {
						installModel = im
					}

					rectangle := buildRectangle(boundaryLeft, boundaryRight, boundaryFront, boundaryRear, installModel)
					properties["rectangle"] = rectangle
				}
			}
		}
	}

	// 区域配置：area_*_* → declare_area
	// TODO: 实现区域配置转换（需要根据实际的前端格式）

	return properties
}

// buildRectangle 构建边界矩形坐标字符串
// 格式：{x1, y1; x2, y2, x3, y3, x4, y4}
// 顶装：X 轴 ±left/right，Y 轴 ±front/rear
// 侧装：X 轴 ±left/right，Y 轴 0 到 front/rear
func buildRectangle(left, right, front, rear float64, installModel string) string {
	if installModel == "ceiling" {
		return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
			int(-left), int(-front),
			int(right), int(-front),
			int(-left), int(rear),
			int(right), int(rear))
	} else {
		// 侧装
		return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
			int(-left), 0,
			int(right), 0,
			int(-left), int(rear),
			int(right), int(rear))
	}
}
