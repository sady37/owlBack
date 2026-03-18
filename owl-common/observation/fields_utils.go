package observation

import (
	"encoding/json"
	"fmt"
)

// ValidationResult 按字段校验结果（硬边界违规、未注册字段、通过校验的 Cleaned）
type ValidationResult struct {
	Invalid []FieldError   // 硬边界违规 (物理不可能值)
	Unknown []string       // 未注册字段
	Cleaned map[string]any // 通过校验的有效数据
}

// FieldError 单字段校验错误
type FieldError struct {
	Key    string
	Value  any
	Reason string
}

// Validate 按 Registry 硬边界校验 payload，不修改原 Data。data 通常为 dataValue 首项（单条 payload 时）。
func Validate(data map[string]interface{}) *ValidationResult {
	if data == nil {
		return &ValidationResult{Cleaned: make(map[string]any)}
	}
	r := &ValidationResult{Cleaned: make(map[string]any, len(data))}
	for k, v := range data {
		def := Lookup(k)
		if def == nil {
			r.Unknown = append(r.Unknown, k)
			continue
		}
		if reason := checkRange(def, v); reason != "" {
			r.Invalid = append(r.Invalid, FieldError{Key: k, Value: v, Reason: reason})
			continue
		}
		r.Cleaned[k] = v
	}
	return r
}

func (r *ValidationResult) OK() bool {
	return len(r.Invalid) == 0 && len(r.Unknown) == 0
}

func checkRange(def *FieldDef, v any) string {
	if def.Min == 0 && def.Max == 0 {
		return ""
	}
	var num float64
	switch n := v.(type) {
	case int:
		num = float64(n)
	case int64:
		num = float64(n)
	case float64:
		num = n
	case json.Number:
		num, _ = n.Float64()
	default:
		return ""
	}
	if def.Min != 0 && num < float64(def.Min) {
		return fmt.Sprintf("below min %d", def.Min)
	}
	if def.Max != 0 && num > float64(def.Max) {
		return fmt.Sprintf("above max %d", def.Max)
	}
	return ""
}

// PersistentFields 返回 payload 中 Persist=true 的字段。data 通常为 dataValue 首项。
func PersistentFields(data map[string]interface{}) map[string]any {
	result := make(map[string]any)
	if data == nil {
		return result
	}
	for k, v := range data {
		if f := Lookup(k); f != nil && f.Persist {
			result[k] = v
		}
	}
	return result
}

// EphemeralFields 返回 payload 中 Persist=false 的字段。data 通常为 dataValue 首项。
func EphemeralFields(data map[string]interface{}) map[string]any {
	result := make(map[string]any)
	if data == nil {
		return result
	}
	for k, v := range data {
		if f := Lookup(k); f != nil && !f.Persist {
			result[k] = v
		}
	}
	return result
}
