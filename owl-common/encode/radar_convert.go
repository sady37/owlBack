package encode

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

//go:embed config/radar_convert_table.json
var radarConvertTableFS embed.FS

// SNOMEDMapping SNOMED 映射结构
type SNOMEDMapping struct {
	SNOMEDCode    *string `json:"snomed_code"`
	SNOMEDDisplay string  `json:"snomed_display"`
	Category      string  `json:"category"`
	DisplayEn     string  `json:"display_en"`
}

// FieldConversion 字段转换规则
type FieldConversion struct {
	FieldPath      string                   `json:"field_path"`
	FieldType      string                   `json:"field_type"`               // "enum", "bit_field", "numeric", "base64_array", "array"
	BytePosition   *int                     `json:"byte_position,omitempty"`  // 单字节位置
	BytePositions  []int                    `json:"byte_positions,omitempty"` // 多字节位置
	ByteOrder      string                   `json:"byte_order,omitempty"`     // "big_endian" 或 "little_endian"
	BitPosition    *string                  `json:"bit_position,omitempty"`   // 如 "7:6", "1:0"
	UnitConversion *UnitConversion          `json:"unit_conversion,omitempty"`
	Mappings       map[string]SNOMEDMapping `json:"mappings,omitempty"`
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
	ToUnit       string `json:"to_unit"`                 // 如 "cm", "秒"
	Direction    string `json:"direction,omitempty"`     // "unidirectional" 或 "bidirectional"
	ApplyTo      string `json:"apply_to,omitempty"`      // "all_coordinates" 或 "coordinate_values_only"（用于数组类型）
}

// RadarConvertTable Radar 转换表结构
type RadarConvertTable struct {
	Version     string                                       `json:"version"`
	LastUpdated string                                       `json:"last_updated"`
	Description string                                       `json:"description"`
	Conversions map[string]map[string]map[string]interface{} `json:"conversions"`
}

var (
	radarConvertTable     *RadarConvertTable
	radarConvertTableOnce sync.Once
	radarConvertTableErr  error
)

// loadRadarConvertTable 加载 Radar 转换表
func loadRadarConvertTable() (*RadarConvertTable, error) {
	radarConvertTableOnce.Do(func() {
		data, err := radarConvertTableFS.ReadFile("config/radar_convert_table.json")
		if err != nil {
			radarConvertTableErr = fmt.Errorf("failed to read Radar convert table: %w", err)
			return
		}

		var table RadarConvertTable
		if err := json.Unmarshal(data, &table); err != nil {
			radarConvertTableErr = fmt.Errorf("failed to parse Radar convert table: %w", err)
			return
		}

		radarConvertTable = &table
	})

	if radarConvertTableErr != nil {
		return nil, radarConvertTableErr
	}

	return radarConvertTable, nil
}

// GetFieldConversion 获取字段转换规则
// fieldPath: 字段路径，如 "monitor.track.pose", "stat.sleep.hr_breath_event.breath_state"
func GetFieldConversion(fieldPath string) (*FieldConversion, error) {
	table, err := loadRadarConvertTable()
	if err != nil {
		return nil, err
	}

	// 解析字段路径
	parts := splitFieldPath(fieldPath)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid field path: %s (expected at least 3 parts)", fieldPath)
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

// ConvertFieldValue 转换字段值（数据上报：设备 → Server）
// fieldPath: 字段路径
// rawValue: 原始值（可以是数字或字符串）
// 返回: 转换后的值、SNOMED 映射信息、错误
func ConvertFieldValue(fieldPath string, rawValue interface{}) (interface{}, *SNOMEDMapping, error) {
	return ConvertConfigValue(fieldPath, rawValue, "read")
}

// ConvertConfigValue 转换配置参数值（支持双向转换）
// fieldPath: 字段路径，如 "config.radar_install_height", "config.rectangle", "config.declare_area"
// rawValue: 原始值（可以是数字、字符串或数组）
// direction: "read" (设备 → Server) 或 "write" (Server → 设备)
// 返回: 转换后的值、SNOMED 映射信息、错误
func ConvertConfigValue(fieldPath string, rawValue interface{}, direction string) (interface{}, *SNOMEDMapping, error) {
	conv, err := GetFieldConversion(fieldPath)
	if err != nil {
		return nil, nil, err
	}

	// 处理枚举类型（通常不需要双向转换）
	if conv.FieldType == "enum" || conv.FieldType == "bit_field" {
		valueStr := fmt.Sprintf("%v", rawValue)
		if mapping, ok := conv.Mappings[valueStr]; ok {
			return rawValue, &mapping, nil
		}
		return rawValue, nil, fmt.Errorf("no mapping found for value: %s in field: %s", valueStr, fieldPath)
	}

	// 处理数组类型（如 rectangle, declare_area）
	if conv.FieldType == "array" && conv.UnitConversion != nil {
		return convertArrayValue(conv, rawValue, direction)
	}

	// 处理数值类型（单位转换）
	if conv.FieldType == "numeric" && conv.UnitConversion != nil {
		return convertNumericValue(conv, rawValue, direction)
	}

	return rawValue, nil, nil
}

// convertNumericValue 转换数值（支持双向转换）
func convertNumericValue(conv *FieldConversion, rawValue interface{}, direction string) (interface{}, *SNOMEDMapping, error) {
	var formula string

	if direction == "read" {
		// 读取：设备 → Server
		if conv.UnitConversion.ReadFormula != "" {
			formula = conv.UnitConversion.ReadFormula
		} else if conv.UnitConversion.Formula != "" {
			formula = conv.UnitConversion.Formula
		} else {
			return rawValue, nil, nil
		}
	} else {
		// 写入：Server → 设备
		if conv.UnitConversion.WriteFormula != "" {
			formula = conv.UnitConversion.WriteFormula
		} else if conv.UnitConversion.Formula != "" {
			// 如果没有 write_formula，尝试反向计算
			// 例如：read 是 "value * 10"，write 应该是 "value / 10"
			return nil, nil, fmt.Errorf("write_formula required for bidirectional conversion: %s", conv.FieldPath)
		} else {
			return rawValue, nil, nil
		}
	}

	result, err := applyFormula(rawValue, formula)
	return result, nil, err
}

// convertArrayValue 转换数组值（如 rectangle, declare_area）
func convertArrayValue(conv *FieldConversion, rawValue interface{}, direction string) (interface{}, *SNOMEDMapping, error) {
	// 如果是字符串，尝试解析为数组
	var values []interface{}

	switch v := rawValue.(type) {
	case string:
		// 解析格式："{x1, y1; x2, y2, x3, y3, x4, y4}" 或 "{area-id, area-type, x1, y1; ...}"
		parsed, err := parseCoordinateString(v)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse coordinate string: %w", err)
		}
		values = parsed
	case []interface{}:
		values = v
	case []int:
		// 转换为 []interface{}
		for _, val := range v {
			values = append(values, val)
		}
	case []float64:
		for _, val := range v {
			values = append(values, val)
		}
	default:
		return rawValue, nil, fmt.Errorf("unsupported array type: %T", rawValue)
	}

	// 确定需要转换的索引
	indicesToConvert := make(map[int]bool)
	if len(conv.CoordinateIndices) > 0 {
		// declare_area: 只转换坐标值（跳过 area-id 和 area-type）
		for _, idx := range conv.CoordinateIndices {
			if idx >= 0 && idx < len(values) {
				indicesToConvert[idx] = true
			}
		}
	} else {
		// rectangle: 转换所有值
		for i := 0; i < len(values); i++ {
			indicesToConvert[i] = true
		}
	}

	// 应用转换
	converted := make([]interface{}, len(values))
	for i, val := range values {
		if indicesToConvert[i] {
			convertedVal, _, err := convertNumericValue(conv, val, direction)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to convert value at index %d: %w", i, err)
			}
			converted[i] = convertedVal
		} else {
			converted[i] = val // 不转换（如 area-id, area-type）
		}
	}

	// 如果原始值是字符串，返回字符串格式
	if _, ok := rawValue.(string); ok {
		return formatCoordinateString(converted, conv.Format), nil, nil
	}

	return converted, nil, nil
}

// applyFormula 应用转换公式
func applyFormula(value interface{}, formula string) (interface{}, error) {
	val, err := parseFloat(value)
	if err != nil {
		return nil, fmt.Errorf("cannot convert value to number: %w", err)
	}

	// 简单的公式解析：支持 "value * X" 或 "value / X"
	if len(formula) > 10 && formula[:6] == "value " {
		if formula[6] == '*' {
			var multiplier float64
			if _, err := fmt.Sscanf(formula[7:], "%f", &multiplier); err != nil {
				return nil, fmt.Errorf("invalid multiplier in formula: %s", formula)
			}
			result := val * multiplier
			// 如果是整数结果，返回 int
			if result == float64(int(result)) {
				return int(result), nil
			}
			return result, nil
		} else if formula[6] == '/' {
			var divisor float64
			if _, err := fmt.Sscanf(formula[7:], "%f", &divisor); err != nil {
				return nil, fmt.Errorf("invalid divisor in formula: %s", formula)
			}
			if divisor == 0 {
				return nil, fmt.Errorf("division by zero in formula: %s", formula)
			}
			result := val / divisor
			// 如果是整数结果，返回 int
			if result == float64(int(result)) {
				return int(result), nil
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("unsupported formula format: %s", formula)
}

// parseCoordinateString 解析坐标字符串
// 格式："{x1, y1; x2, y2, x3, y3, x4, y4}" 或 "{area-id, area-type, x1, y1; ...}"
func parseCoordinateString(s string) ([]interface{}, error) {
	// 移除花括号和空格
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	// 按分号分割
	parts := strings.Split(s, ";")
	var values []interface{}

	for _, part := range parts {
		// 按逗号分割每个部分
		coords := strings.Split(part, ",")
		for _, coord := range coords {
			coord = strings.TrimSpace(coord)
			if coord == "" {
				continue
			}
			// 尝试解析为整数
			var val int
			if _, err := fmt.Sscanf(coord, "%d", &val); err == nil {
				values = append(values, val)
			} else {
				// 如果解析失败，保持为字符串
				values = append(values, coord)
			}
		}
	}

	return values, nil
}

// formatCoordinateString 格式化坐标数组为字符串
func formatCoordinateString(values []interface{}, format string) string {
	if format == "" {
		// 如果没有格式，使用默认格式
		return fmt.Sprintf("{%v}", values)
	}

	// 根据 format 格式化
	// 简化处理：直接使用值列表
	var parts []string
	for _, val := range values {
		parts = append(parts, fmt.Sprintf("%v", val))
	}

	// 尝试匹配原始格式
	if strings.Contains(format, ";") {
		// 有分号：每两个值一组
		var result strings.Builder
		result.WriteString("{")
		for i := 0; i < len(parts); i += 2 {
			if i > 0 {
				result.WriteString("; ")
			}
			if i+1 < len(parts) {
				result.WriteString(fmt.Sprintf("%s, %s", parts[i], parts[i+1]))
			} else {
				result.WriteString(parts[i])
			}
		}
		result.WriteString("}")
		return result.String()
	}

	// 没有分号：所有值用逗号分隔
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

// GetSNOMEDMappingByFieldPath 根据字段路径和原始值获取 SNOMED 映射
func GetSNOMEDMappingByFieldPath(fieldPath string, rawValue interface{}) (*SNOMEDMapping, error) {
	_, mapping, err := ConvertFieldValue(fieldPath, rawValue)
	return mapping, err
}

// ParseBitField 从字节值中解析位字段
// byteValue: 原始字节值（0-255）
// bitPosition: 位位置字符串，如 "7:6" (高位:低位)
// 返回: 位字段的二进制字符串（如 "00", "01", "10", "11"）
func ParseBitField(byteValue int, bitPosition string) (string, error) {
	parts := strings.Split(bitPosition, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid bit position format: %s, expected 'high:low'", bitPosition)
	}

	highBit, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid high bit: %w", err)
	}
	lowBit, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid low bit: %w", err)
	}

	if highBit < lowBit || highBit > 7 || lowBit < 0 {
		return "", fmt.Errorf("invalid bit range: high=%d, low=%d", highBit, lowBit)
	}

	// 计算掩码和位数
	bitWidth := highBit - lowBit + 1
	mask := (1<<bitWidth - 1) << lowBit
	extractedBits := (byteValue & mask) >> lowBit

	// 格式化为二进制字符串（固定宽度）
	return fmt.Sprintf("%0*b", bitWidth, extractedBits), nil
}

// splitFieldPath 分割字段路径
func splitFieldPath(path string) []string {
	var parts []string
	var current string

	for _, char := range path {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
