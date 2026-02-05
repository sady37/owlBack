package snomed

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed sleepace_convert_table.json
var sleepaceConvertTable embed.FS

var (
	sleepaceConvertTableOnce sync.Once
	sleepaceConvertTableMap  map[string]map[string]interface{}
)

// loadSleepaceConvertTable 加载 Sleepace 转换表（单例）
func loadSleepaceConvertTable() {
	if sleepaceConvertTableMap != nil {
		return
	}
	data, err := sleepaceConvertTable.ReadFile("sleepace_convert_table.json")
	if err != nil {
		panic(fmt.Errorf("failed to read sleepace_convert_table.json: %w", err))
	}
	err = json.Unmarshal(data, &sleepaceConvertTableMap)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal sleepace_convert_table.json: %w", err))
	}
}

// GetSleepaceMappingByFieldPath 根据字段路径获取 Sleepace 映射（含 SNOMED 信息）
// fieldPath: 字段路径，如 "data.sleepStage"；rawValue: 原始值
// 返回: *SNOMEDMapping 或 nil（无映射时）
func GetSleepaceMappingByFieldPath(fieldPath string, rawValue interface{}) (*SNOMEDMapping, error) {
	sleepaceConvertTableOnce.Do(loadSleepaceConvertTable)

	// 特殊处理：connectionStatus 的 rawValue 是布尔值
	if fieldPath == "data.connectionStatus" {
		boolVal, ok := rawValue.(bool)
		if !ok {
			return nil, fmt.Errorf("connectionStatus value is not bool: %v", rawValue)
		}
		key := "true"
		if !boolVal {
			key = "false"
		}
		if mapping, exists := sleepaceConvertTableMap[fieldPath][key]; exists {
			mappingMap, ok := mapping.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid mapping format for %s.%s", fieldPath, key)
			}
			return &SNOMEDMapping{
				SNOMEDCode:    toStringPtr(mappingMap["snomed_code"]),
				SNOMEDDisplay: mappingMap["snomed_display"].(string),
				Category:      mappingMap["category"].(string),
				DisplayEn:     mappingMap["display_en"].(string),
			}, nil
		}
		return nil, nil
	}

	// 其他字段：rawValue 通常是字符串
	strVal, ok := rawValue.(string)
	if !ok {
		// 尝试转换为字符串（如数字等）
		strVal = fmt.Sprintf("%v", rawValue)
	}

	// 查找映射（支持大小写不敏感）
	lowerStrVal := strings.ToLower(strVal)
	if mapping, exists := sleepaceConvertTableMap[fieldPath][lowerStrVal]; exists {
		mappingMap, ok := mapping.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid mapping format for %s.%s", fieldPath, lowerStrVal)
		}
		return &SNOMEDMapping{
			SNOMEDCode:    toStringPtr(mappingMap["snomed_code"]),
			SNOMEDDisplay: mappingMap["snomed_display"].(string),
			Category:      mappingMap["category"].(string),
			DisplayEn:     mappingMap["display_en"].(string),
		}, nil
	}

	// 未找到映射，返回 nil
	return nil, nil
}

// toStringPtr 安全地将 interface{} 转为 *string（nil-aware）
func toStringPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	if str, ok := v.(string); ok {
		return &str
	}
	str := fmt.Sprintf("%v", v)
	return &str
}
