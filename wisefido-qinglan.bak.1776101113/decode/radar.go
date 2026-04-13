package decode

import (
	"owl-common/alarm"
	internalDecode "wisefido-qinglan/internal/decode"
)

// DecodeDevicePropsToAlarmItems 从设备属性更新 AlarmItems（用于设备写入失败时同步实际值）
// 支持工作模式（radar_func_ctrl）、跌倒参数（fall_param）和呼吸心率参数（heart_breath_param）的反向解析
func DecodeDevicePropsToAlarmItems(items []alarm.AlarmItem, deviceProps map[string]interface{}) ([]alarm.AlarmItem, error) {
	return internalDecode.DecodeDevicePropsToAlarmItems(items, deviceProps)
}

// DecodeFallParam 从 fall_param base64 解码为 AlarmItem 参数
func DecodeFallParam(fallParamBase64 string) (map[string]interface{}, error) {
	return internalDecode.DecodeFallParam(fallParamBase64)
}

// DecodeHeartBreathParam 从 heart_breath_param base64 解码为 AlarmItem 参数
func DecodeHeartBreathParam(heartBreathParamBase64 string) (map[string]interface{}, error) {
	return internalDecode.DecodeHeartBreathParam(heartBreathParamBase64)
}
