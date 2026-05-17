package decode

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"owl-common/alarm"
)

// RadarEncoder 编码 Radar 数据（Server → 设备）
// 负责将 Server 端统一单位的数据转换为设备协议格式，包括单位转换和 base64 编码

// EncodeFallParam 从 AlarmItem[] 构建 fall_param base64（应用单位转换：秒 → 10秒单位）
// 协议：16-byte BASE64，byte 3 = fall_alarm_duration/10，byte 4 = bits，byte 5 = sitting_alarm_duration/10
func EncodeFallParam(items []alarm.AlarmItem) (string, error) {
	b := make([]byte, 16)

	// 索引 AlarmItem 按类型
	m := indexAlarmItemsByType(items)

	// byte 3: 跌倒告警时间（秒 → 10秒单位，直接除以10）
	dur := 60 // 默认 60 秒
	if it, ok := m[alarm.Fall]; ok && it.AlarmParams != nil {
		if v, ok := it.AlarmParams["duration_sec"]; ok {
			durVal := extractIntValue(v)
			if durVal >= 0 && durVal <= 2500 {
				dur = durVal
			}
		}
	}
	b[3] = byte(dur / 10)

	// byte 4: bit0 坐地, bit1 姿态, bit2 床上坐起
	// byte 5: 坐地告警时间（秒 → 10秒单位，直接除以10）
	sitDur := 100 // 默认 100 秒
	if it, ok := m[alarm.SittingOnGround]; ok {
		if it.IsEnabled != nil && *it.IsEnabled == 1 {
			b[4] |= 1
			if it.AlarmParams != nil {
				if v, ok := it.AlarmParams["duration_sec"]; ok {
					durVal := extractIntValue(v)
					if durVal >= 10 && durVal <= 2500 {
						sitDur = durVal
					}
				}
			}
		}
	}
	if it, ok := m[alarm.PostureDetection]; ok && it.IsEnabled != nil && *it.IsEnabled == 1 {
		b[4] |= 2
	}
	if it, ok := m[alarm.LeftBed]; ok && it.IsEnabled != nil && *it.IsEnabled == 1 {
		b[4] |= 4
	}
	if b[4]&(1|4) != 0 {
		b[4] |= 2
	}
	if sitDur < 10 {
		sitDur = 10
	}
	b[5] = byte(sitDur / 10)

	return base64.StdEncoding.EncodeToString(b), nil
}

// EncodeHeartBreathParam 从 AlarmItem[] 构建 heart_breath_param base64
// 注意：heart_breath_param 的单位已一致（分钟），无需转换
// 协议：16-byte BASE64，0=upper breath, 1=upper heart, 2=lower breath, 3=lower heart,
//
//	4=intensive care, 5=weak vital duration (min), 6=sensitivity
func EncodeHeartBreathParam(items []alarm.AlarmItem) (string, error) {
	b := make([]byte, 16)

	// 索引 AlarmItem 按类型
	m := indexAlarmItemsByType(items)

	upperBreath, lowerBreath := 24, 8
	upperHeart, lowerHeart := 90, 50
	intensiveCare := 0
	weakDuration := 10
	weakSensitivity := 35

	if it, ok := m[alarm.RespRateAlert]; ok && it.AlarmParams != nil {
		if v, ok := it.AlarmParams["max"]; ok {
			upperBreath = extractIntValue(v)
		}
		if v, ok := it.AlarmParams["min"]; ok {
			lowerBreath = extractIntValue(v)
		}
	}
	if it, ok := m[alarm.HeartRateAlert]; ok && it.AlarmParams != nil {
		if v, ok := it.AlarmParams["max"]; ok {
			upperHeart = extractIntValue(v)
		}
		if v, ok := it.AlarmParams["min"]; ok {
			lowerHeart = extractIntValue(v)
		}
	}
	if it, ok := m[alarm.WeakBiometricSignal]; ok && it.AlarmParams != nil {
		// duration_min（分钟）直接使用，设备协议 byte 5 也是分钟单位
		if v, ok := it.AlarmParams["duration_min"]; ok {
			weakDuration = extractIntValue(v)
		}
		if v, ok := it.AlarmParams["sensitivity"]; ok {
			weakSensitivity = extractIntValue(v)
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

	return base64.StdEncoding.EncodeToString(b), nil
}

// indexAlarmItemsByType 按 AlarmType 索引 AlarmItem[]
func indexAlarmItemsByType(items []alarm.AlarmItem) map[string]alarm.AlarmItem {
	m := make(map[string]alarm.AlarmItem)
	for _, it := range items {
		if it.AlarmType != "" {
			m[it.AlarmType] = it
		}
	}
	return m
}

// EncodeAlarmItemsToDeviceProps 将雷达监控设置 AlarmItem[] 转为设备属性（仅工作模式+跌倒+呼吸心率）
// 用于 radar-monitor-settings 页「工作模式、跌倒和呼吸心率参数」一齐提交
// 协议：3.4.1 radar_func_ctrl；3.4.4 fall_param；3.4.5 heart_breath_param
func EncodeAlarmItemsToDeviceProps(items []alarm.AlarmItem) (map[string]interface{}, error) {
	m := indexAlarmItemsByType(items)

	props := make(map[string]interface{})

	// 1. 工作模式 radar_func_ctrl (3/7/11/15)
	if v := modeFromItems(m); v >= 0 {
		props["radar_func_ctrl"] = v
	}

	// 2. 构建 fall_param base64（应用单位转换：秒 → 10秒单位）
	if fallParamBase64, err := EncodeFallParam(items); err == nil {
		props["fall_param"] = fallParamBase64
	} else {
		return props, fmt.Errorf("encode fall_param: %w", err)
	}

	// 3. 构建 heart_breath_param base64（无需转换，单位已一致）
	if heartBreathParamBase64, err := EncodeHeartBreathParam(items); err == nil {
		props["heart_breath_param"] = heartBreathParamBase64
	} else {
		return props, fmt.Errorf("encode heart_breath_param: %w", err)
	}

	return props, nil
}

// modeFromItems 从 AlarmItem[] 提取工作模式
func modeFromItems(m map[string]alarm.AlarmItem) int {
	it, ok := m[alarm.MonitoringMode]
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

// EncodeV1ConfigToDeviceProps 将规范配置转为设备属性。入参可为规范 key（install_height cm、install_model、work_model、rectangle cm）或设备 key（radar_install_height dm）；规范 key 会做单位换算（cm→dm）。
// 入参：install_model 0/1/2（统一数值），boundary_left/right/front/rear，area_{i}_* 等；area_{i}_id=-1 表示删除
func EncodeV1ConfigToDeviceProps(config map[string]interface{}) map[string]interface{} {
	properties := make(map[string]interface{})

	installModel := resolveInstallModel(config) // 0/1/2

	// 安装方式 → radar_install_style (0=顶装, 1=侧装, 2=角装)
	if _, has := config["install_model"]; has {
		properties["radar_install_style"] = fmt.Sprintf("%d", installModel)
	}

	// 安装高度：规范 install_height(cm) → radar_install_height(dm)；设备 key 透传
	if v, ok := config["radar_install_height"]; ok && v != nil {
		properties["radar_install_height"] = int(extractFloat64Value(v))
	} else if v, ok := config["install_height"]; ok && v != nil {
		cm := extractFloat64Value(v)
		properties["radar_install_height"] = int(cm / 10) // cm → dm
	} else if v, ok := config["height"]; ok && v != nil {
		cm := extractFloat64Value(v)
		properties["radar_install_height"] = int(cm / 10)
	}

	// 工作模式 → radar_func_ctrl（规范 work_model 或设备 radar_func_ctrl）
	if v, ok := config["work_model"]; ok && v != nil {
		properties["radar_func_ctrl"] = int(extractFloat64Value(v))
	} else if v, ok := config["radar_func_ctrl"]; ok && v != nil {
		properties["radar_func_ctrl"] = int(extractFloat64Value(v))
	}

	// 边界：规范 rectangle(cm) 会转 dm；若缺 rectangle 则由 boundary_* buildRectangle（单位 dm）
	var rect string
	if r, ok := config["rectangle"].(string); ok && r != "" {
		rect = rectangleCMToDM(r)
		if rect == "" {
			rect = r
		}
	} else if _, hasL := config["boundary_left"]; hasL {
		if _, hasR := config["boundary_right"]; hasR {
			if _, hasF := config["boundary_front"]; hasF {
				if _, hasB := config["boundary_rear"]; hasB {
					left := extractFloat64Value(config["boundary_left"])
					right := extractFloat64Value(config["boundary_right"])
					front := extractFloat64Value(config["boundary_front"])
					rear := extractFloat64Value(config["boundary_rear"])
					li, ri, fi, rei := int(left), int(right), int(front), int(rear)
					if li == -127 && ri == -127 && fi == -127 && rei == -127 {
						rect = "{-127,-127;-127,-127;-127,-127;-127,-127}"
					} else if li == 127 && ri == 127 && fi == 127 && rei == 127 {
						rect = "{127,127;127,127;127,127;127,127}"
					} else {
						rect = buildRectangle(left, right, front, rear, installModel)
					}
				}
			}
		}
	}
	if rect != "" {
		properties["rectangle"] = rect
	}

	// 区域：前端可传 declare_area 字符串（qinglan 格式），直接透传；否则从 area_{i}_* 构建
	if s, ok := config["declare_area"].(string); ok && s != "" {
		properties["declare_area"] = s
	} else if decl := buildDeclareAreaFromConfigCm(config); len(decl) > 0 {
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

// APIToDeviceProperties 透传 HTTP API 传入的 properties，不做单位转换；radar_install_height/rectangle 由前端按设备格式（dm）传入。
func APIToDeviceProperties(properties map[string]interface{}) map[string]interface{} {
	if properties == nil {
		return nil
	}
	out := make(map[string]interface{}, len(properties))
	for k, v := range properties {
		out[k] = v
	}
	return out
}

// resolveInstallModel 从 config 解析 install_model，返回 0/1/2（项目内统一数值）
func resolveInstallModel(config map[string]interface{}) int {
	if v, ok := config["install_model"]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			n := int(x)
			if n >= 0 && n <= 2 {
				return n
			}
		case int:
			if x >= 0 && x <= 2 {
				return x
			}
		case string:
			switch x {
			case "ceiling":
				return 0
			case "wall":
				return 1
			case "corn":
				return 2
			}
		}
	}
	return 1
}

// buildDeclareAreaFromConfigCm 从 area_{i}_id/type/x1..y4 构建 declare_area；入参 cm，出参 dm（只在此处 /10 一次，qinglan 不再除）
func buildDeclareAreaFromConfigCm(config map[string]interface{}) []interface{} {
	var out []interface{}
	for i := 0; i < 16; i++ {
		pre := fmt.Sprintf("area_%d_", i)
		idVal := config[pre+"id"]
		if idVal == nil {
			continue
		}
		id := int(extractFloat64Value(idVal))
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
		x1 := int(extractFloat64Value(config[pre+"x1"]) / 10)
		y1 := int(extractFloat64Value(config[pre+"y1"]) / 10)
		x2 := int(extractFloat64Value(config[pre+"x2"]) / 10)
		y2 := int(extractFloat64Value(config[pre+"y2"]) / 10)
		x3 := int(extractFloat64Value(config[pre+"x3"]) / 10)
		y3 := int(extractFloat64Value(config[pre+"y3"]) / 10)
		x4 := int(extractFloat64Value(config[pre+"x4"]) / 10)
		y4 := int(extractFloat64Value(config[pre+"y4"]) / 10)
		el := map[string]interface{}{
			"id": id, "type": areaType,
			"x1": x1, "y1": y1, "x2": x2, "y2": y2, "x3": x3, "y3": y3, "x4": x4, "y4": y4,
		}
		out = append(out, el)
	}
	return out
}

// rectangleCMToDM 将规范 rectangle 字符串（cm）转为设备格式（dm）。格式 {x1,y1;x2,y2;x3,y3;x4,y4}，解析失败返回空串。
func rectangleCMToDM(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return ""
	}
	if s[0] == '{' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '}' {
		s = s[:len(s)-1]
	}
	points := strings.Split(s, ";")
	if len(points) != 4 {
		return ""
	}
	var nums []int
	for _, p := range points {
		parts := strings.Split(strings.TrimSpace(p), ",")
		if len(parts) != 2 {
			return ""
		}
		x, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		y, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return ""
		}
		nums = append(nums, x/10, y/10)
	}
	return fmt.Sprintf("{%d,%d;%d,%d;%d,%d;%d,%d}",
		nums[0], nums[1], nums[2], nums[3], nums[4], nums[5], nums[6], nums[7])
}

// buildRectangle 构建边界矩形坐标字符串，格式 {h1,v1;h2,v2;h3,v3;h4,v4}（单位由调用方负责）
// installModel: 0=吸顶 1=贴墙 2=墙角
func buildRectangle(left, right, front, rear float64, installModel int) string {
	if installModel == 0 {
		return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
			int(-right), int(-rear),
			int(left), int(-rear),
			int(-right), int(front),
			int(left), int(front))
	}
	// 侧装、角装：与前端 Wall 顺序一致 (-right,-rear),(left,-rear),(-right,front),(left,front)
	return fmt.Sprintf("{%d, %d; %d, %d; %d, %d; %d, %d}",
		int(-right), int(-rear),
		int(left), int(-rear),
		int(-right), int(front),
		int(left), int(front))
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

// extractIntValue 从 interface{} 提取 int 值
func extractIntValue(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	default:
		return 0
	}
}

// extractFloat64Value 从 interface{} 提取 float64 值
func extractFloat64Value(v interface{}) float64 {
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
