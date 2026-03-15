package snomed

import (
	"fmt"
	"strconv"
)

// SleepaceEncode 编码 Sleepace 数据
// data: 包含 device_id, tenant_id, device_type, data_key, timestamp 和原始数据字段
// dataKey: 数据类型 ("realtime", "sleepStage", "connectionStatus", "alarmNotify")
// 返回: 编码后的数据（格式统一、字段重命名、SNOMED 映射）
func SleepaceEncode(data map[string]interface{}, dataKey string) (map[string]interface{}, error) {
	encoded := make(map[string]interface{})

	// 1. 保留元数据
	encoded["device_id"] = data["device_id"]
	encoded["tenant_id"] = data["tenant_id"]
	encoded["device_type"] = data["device_type"]
	encoded["data_key"] = dataKey
	encoded["timestamp"] = data["timestamp"]

	// 2. 根据 dataKey 进行不同的编码转换
	switch dataKey {
	case "realtime":
		return encodeSleepaceRealtime(data, encoded)
	case "sleepStage":
		return encodeSleepaceSleepStage(data, encoded)
	case "connectionStatus":
		return encodeSleepaceConnectionStatus(data, encoded)
	case "alarmNotify":
		return encodeSleepaceAlarmNotify(data, encoded)
	default:
		// 默认：直接复制其他字段
		copyOtherFields(data, encoded, []string{"device_id", "tenant_id", "device_type", "data_key", "timestamp"})
		return encoded, nil
	}
}

// encodeSleepaceRealtime 编码实时数据
func encodeSleepaceRealtime(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// 字段重命名：统一使用标准字段名（各厂家统一）
	// Sleepace: breath → respiratory_rate, heart → heart_rate
	if breath, ok := data["breath"]; ok {
		encoded["respiratory_rate"] = breath  // 标准字段名（各厂家统一）
	}
	if heart, ok := data["heart"]; ok {
		encoded["heart_rate"] = heart  // 标准字段名（各厂家统一）
	}

	// bedStatus: 0=在床, 1=离床 - 需要 SNOMED 映射
	if bedStatus, ok := data["bedStatus"]; ok {
		applySleepaceSNOMedMapping(encoded, "bedStatus", "realtime.bedStatus", bedStatus)
	}

	// sitUp: 床上坐起/躺下 - 需要 SNOMED 映射
	// sitUp = 0: 床上躺下 (Lying position, 109030009)
	// sitUp > 0: 床上坐起 (Sitting up in bed, 225698008)
	if sitUp, ok := data["sitUp"]; ok {
		if sitUpVal, err := parseInt(sitUp); err == nil {
			// 始终进行 SNOMED 映射：sitUp = 0 映射为躺下，sitUp > 0 映射为坐起
			if sitUpVal > 0 {
				applySleepaceSNOMedMapping(encoded, "sitUp", "realtime.sitUp", 1) // 映射为坐起
			} else {
				applySleepaceSNOMedMapping(encoded, "sitUp", "realtime.sitUp", 0) // 映射为躺下
			}
			encoded["sitUp"] = sitUpVal // 保留原始值
		} else {
			encoded["sitUp"] = sitUp
		}
	}

	// turnOver 和 bodyMove: 体动相关，不进行 SNOMED 映射，直接保留
	if turnOver, ok := data["turnOver"]; ok {
		encoded["turnOver"] = turnOver
	}
	if bodyMove, ok := data["bodyMove"]; ok {
		encoded["bodyMove"] = bodyMove
	}

	// 复制其他字段（包括 signalQuality, initStatus, leftRight 等）
	// 注意：realtime 数据中不包含 sleepStage 字段（sleepStage 只在 sleepStage 数据类型中存在）
	excludeFields := []string{"device_id", "tenant_id", "device_type", "data_key", "timestamp", "breath", "heart", "bedStatus", "sitUp", "turnOver", "bodyMove"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// encodeSleepaceSleepStage 编码睡眠阶段数据（用于 SNOMED/报告等，与 realtime 统一码 1/2/4/8 不同）
func encodeSleepaceSleepStage(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// realtime 已统一为 1=awake,2=light,4=deep,8=unknown
	if sleepStage, ok := data["sleepStage"]; ok {
		applySleepaceSNOMedMapping(encoded, "sleepStage", "sleepStage.sleepStage", sleepStage)
	}

	// 复制其他字段（如 leftRight）
	excludeFields := []string{"device_id", "tenant_id", "device_type", "data_key", "timestamp", "sleepStage"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// encodeSleepaceConnectionStatus 编码连接状态数据
func encodeSleepaceConnectionStatus(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// connectionStatus: 0=离线, 1=在线 - 需要 SNOMED 映射
	if connectionStatus, ok := data["connectionStatus"]; ok {
		applySleepaceSNOMedMapping(encoded, "connectionStatus", "connectionStatus.connectionStatus", connectionStatus)
	}

	// 复制其他字段
	excludeFields := []string{"device_id", "tenant_id", "device_type", "data_key", "timestamp", "connectionStatus"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// encodeSleepaceAlarmNotify 编码告警通知数据
func encodeSleepaceAlarmNotify(data, encoded map[string]interface{}) (map[string]interface{}, error) {
	// alarmType: 报警类型 - 需要 SNOMED 映射
	// 优先使用 alarmType 字段（从 handleAlarmNotify 中设置）
	if alarmType, ok := data["alarmType"]; ok {
		applySleepaceSNOMedMapping(encoded, "alarmType", "alarmNotify.type", alarmType)
	} else if alarmType, ok := data["type"]; ok {
		// 如果只有 type 字段，也进行映射（兼容性处理）
		applySleepaceSNOMedMapping(encoded, "alarmType", "alarmNotify.type", alarmType)
		encoded["type"] = alarmType // 保留原始字段名
	}

	// 复制其他字段（status, userId, relieveReason, relieveTime 等）
	// 注意：alarmStatus, userId, relieveReason, relieveTime 等字段不需要 SNOMED 映射
	excludeFields := []string{"device_id", "tenant_id", "device_type", "data_key", "timestamp", "type", "alarmType"}
	copyOtherFields(data, encoded, excludeFields)

	return encoded, nil
}

// copyOtherFields 复制源map中除excludeFields外的所有字段到目标map
func copyOtherFields(src, dst map[string]interface{}, excludeFields []string) {
	excludeMap := make(map[string]bool)
	for _, field := range excludeFields {
		excludeMap[field] = true
	}

	for key, value := range src {
		if !excludeMap[key] {
			dst[key] = value
		}
	}
}

// applySleepaceSNOMedMapping 应用 Sleepace SNOMED 映射
func applySleepaceSNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	mapping, err := GetSleepaceMappingByFieldPath(fieldPath, rawValue)
	if err != nil || mapping == nil {
		encoded[fieldName] = rawValue
		return
	}
	if mapping.DisplayEn != "" {
		encoded[fieldName] = mapping.DisplayEn
	} else {
		encoded[fieldName] = rawValue
	}
}

// parseInt 将 interface{} 安全地解析为整数
func parseInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot parse %T to int", value)
	}
}