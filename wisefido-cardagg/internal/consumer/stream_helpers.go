package consumer

import (
	"encoding/json"
	"strconv"
)

// ConvertRedisValues 将 Redis Stream 返回的全 string map 转换为正确 Go 类型
// Redis Stream field values 全部是 string，需要：
//   - timestamp: string → int64
//   - data_value: JSON string → 解析后的值（数组或对象）
func ConvertRedisValues(raw map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		strVal, isStr := v.(string)
		if !isStr {
			out[k] = v
			continue
		}
		switch k {
		case "timestamp":
			if ts, err := strconv.ParseInt(strVal, 10, 64); err == nil {
				out[k] = ts
			} else {
				out[k] = strVal
			}
		case "data_value":
			var arr []interface{}
			if err := json.Unmarshal([]byte(strVal), &arr); err == nil {
				out[k] = arr
			} else {
				var obj map[string]interface{}
				if err := json.Unmarshal([]byte(strVal), &obj); err == nil {
					out[k] = obj
				} else {
					out[k] = strVal
				}
			}
		default:
			out[k] = strVal
		}
	}
	return out
}
