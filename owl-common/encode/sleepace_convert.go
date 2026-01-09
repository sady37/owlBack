package encode

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed config/sleepace_convert_table.json
var sleepaceConvertTableFS embed.FS

// SleepaceConvertTable Sleepace 转换表结构
type SleepaceConvertTable struct {
	Version     string                                       `json:"version"`
	LastUpdated string                                       `json:"last_updated"`
	Description string                                       `json:"description"`
	Conversions map[string]map[string]map[string]interface{} `json:"conversions"`
}

var (
	sleepaceConvertTable     *SleepaceConvertTable
	sleepaceConvertTableOnce sync.Once
	sleepaceConvertTableErr  error
)

// loadSleepaceConvertTable 加载 Sleepace 转换表
func loadSleepaceConvertTable() (*SleepaceConvertTable, error) {
	sleepaceConvertTableOnce.Do(func() {
		data, err := sleepaceConvertTableFS.ReadFile("config/sleepace_convert_table.json")
		if err != nil {
			sleepaceConvertTableErr = fmt.Errorf("failed to read Sleepace convert table: %w", err)
			return
		}

		var table SleepaceConvertTable
		if err := json.Unmarshal(data, &table); err != nil {
			sleepaceConvertTableErr = fmt.Errorf("failed to parse Sleepace convert table: %w", err)
			return
		}

		sleepaceConvertTable = &table
	})

	if sleepaceConvertTableErr != nil {
		return nil, sleepaceConvertTableErr
	}

	return sleepaceConvertTable, nil
}

// GetSleepaceFieldConversion 获取 Sleepace 字段转换规则
// fieldPath: 字段路径，如 "realtime.bedStatus", "realtime.sleepStage", "sleepStage.sleepStage"
func GetSleepaceFieldConversion(fieldPath string) (*FieldConversion, error) {
	table, err := loadSleepaceConvertTable()
	if err != nil {
		return nil, err
	}

	// 解析字段路径
	parts := splitFieldPath(fieldPath)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid field path: %s (expected at least 2 parts)", fieldPath)
	}

	// 从转换表中查找
	var current interface{} = table.Conversions
	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			if next, exists := m[part]; exists {
				current = next
			} else {
				return nil, fmt.Errorf("field path not found: %s (missing: %s)", fieldPath, part)
			}
		} else {
			return nil, fmt.Errorf("invalid structure at path: %s", fieldPath)
		}
	}

	// 转换为 FieldConversion
	convJSON, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conversion: %w", err)
	}

	var conv FieldConversion
	if err := json.Unmarshal(convJSON, &conv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversion: %w", err)
	}

	return &conv, nil
}

// GetSleepaceSNOMEDMappingByFieldPath 根据字段路径和原始值获取 Sleepace SNOMED 映射
func GetSleepaceSNOMEDMappingByFieldPath(fieldPath string, rawValue interface{}) (*SNOMEDMapping, error) {
	conv, err := GetSleepaceFieldConversion(fieldPath)
	if err != nil {
		return nil, err
	}

	// 处理枚举类型
	if conv.FieldType == "enum" && conv.Mappings != nil {
		valueStr := fmt.Sprintf("%v", rawValue)
		if mapping, ok := conv.Mappings[valueStr]; ok {
			return &mapping, nil
		}
		return nil, fmt.Errorf("no mapping found for value: %s in field: %s", valueStr, fieldPath)
	}

	return nil, fmt.Errorf("field %s does not support SNOMED mapping", fieldPath)
}
