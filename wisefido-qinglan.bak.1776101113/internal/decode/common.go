package decode

import "fmt"

// copyOtherFields 复制其他字段（排除指定字段）
func copyOtherFields(source, target map[string]interface{}, exclude []string) {
	excludeMap := make(map[string]bool)
	for _, k := range exclude {
		excludeMap[k] = true
	}

	for k, v := range source {
		if !excludeMap[k] {
			target[k] = v
		}
	}
}

// parseInt 将值转换为整数
func parseInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		var result int
		_, err := fmt.Sscanf(v, "%d", &result)
		return result, err
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// parseFloat 将值转换为浮点数
func parseFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		var result float64
		_, err := fmt.Sscanf(v, "%f", &result)
		return result, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}


