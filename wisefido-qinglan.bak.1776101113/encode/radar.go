package encode

import (
	"fmt"
	"strings"

	"owl-common/alarm"
	"wisefido-qinglan/internal/decode"
)

// EncodeAlarmItemsToDeviceProps 将雷达监控设置 AlarmItem[] 转为设备属性（仅工作模式+跌倒+呼吸心率）
// 用于 radar-monitor-settings 页「工作模式、跌倒和呼吸心率参数」一齐提交
// 协议：3.4.1 radar_func_ctrl；3.4.4 fall_param；3.4.5 heart_breath_param
func EncodeAlarmItemsToDeviceProps(items []alarm.AlarmItem) (map[string]interface{}, error) {
	return decode.EncodeAlarmItemsToDeviceProps(items)
}

// EncodeV1ConfigToDeviceProps 将 v1 / MQTT 风格配置转为 Radar 设备属性
// 入参与 radarMqttConfig 输出对齐：cm（前端统一单位），install_model 统一 0/1/2，
// boundary_left/right/front/rear，area_{i}_id、area_{i}_type、area_{i}_x1..y4（cm）；area_{i}_id=-1 表示删除
// 出参：radar_install_style, radar_install_height, rectangle, declare_area
// 注意：单位转换统一在 wisefido-qinglan 中处理，此处只做格式转换
func EncodeV1ConfigToDeviceProps(config map[string]interface{}) map[string]interface{} {
	return decode.EncodeV1ConfigToDeviceProps(config)
}

// ToStr 将 interface{} 转为字符串，数字均为字符型
func ToStr(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// MaskSsidPassword 对 SSID:password 格式的字符串，将冒号后部分改为 *******，返回给前端时使用
// 若为空或不含冒号（即无 ssid: 部分），不替换，原样返回
func MaskSsidPassword(s string) string {
	if s == "" {
		return s
	}
	if idx := strings.Index(s, ":"); idx >= 0 {
		return s[:idx+1] + "*******"
	}
	return s
}
