package encode

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"owl-common/snomed"
)

// SNOMEDMapping 已迁移到 owl-common/snomed，保留别名以保持向后兼容
type SNOMEDMapping = snomed.SNOMEDMapping

// FieldConversion 字段转换规则
type FieldConversion struct {
	FieldPath      string                   `json:"field_path"`
	FieldType      string                   `json:"field_type"`               // "enum", "bit_field", "numeric", "base64_array", "array"
	BytePosition   *int                     `json:"byte_position,omitempty"`  // 单字节位置
	BytePositions  []int                    `json:"byte_positions,omitempty"` // 多字节位置
	ByteOrder      string                   `json:"byte_order,omitempty"`     // "big_endian" 或 "little_endian"
	BitPosition    *string                  `json:"bit_position,omitempty"`   // 如 "7:6", "1:0"
	UnitConversion *UnitConversion          `json:"unit_conversion,omitempty"`
	Mappings       map[string]snomed.SNOMEDMapping `json:"mappings,omitempty"`
	// 数组类型相关字段
	ArrayItemType     string `json:"array_item_type,omitempty"`    // "coordinate_pair", "area_definition"
	Format            string `json:"format,omitempty"`             // 格式字符串，如 "{x1, y1; x2, y2, x3, y3, x4, y4}"
	CoordinateIndices []int  `json:"coordinate_indices,omitempty"` // 坐标值的索引（用于 declare_area，跳过 area-id 和 area-type）
}

// UnitConversion 单位转换规则
type UnitConversion struct {
	Formula      string `json:"formula,omitempty"`       // 单向转换公式，如 "value * 10", "value * 100"
	ReadFormula  string `json:"read_formula,omitempty"`  // 双向转换：读取公式（设备 → Server）
	WriteFormula string `json:"write_formula,omitempty"` // 双向转换：写入公式（Server → 设备）
	FromUnit     string `json:"from_unit"`               // 如 "dm", "m", "10秒单位"
	ToUnit       string `json:"to_unit"`                // 如 "cm", "秒"
	Direction    string `json:"direction,omitempty"`     // "unidirectional" 或 "bidirectional"
	ApplyTo      string `json:"apply_to,omitempty"`      // "all_coordinates" 或 "coordinate_values_only"（用于数组类型）
}

// splitFieldPath 分割字段路径
func splitFieldPath(path string) []string {
	return strings.Split(path, ".")
}

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
