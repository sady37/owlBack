package snomed

// GetSNOMEDMappingByFieldPath 无转换表时返回 nil，使用原始值（备用，可后续接入 radar_convert_table 等）
func GetSNOMEDMappingByFieldPath(fieldPath string, rawValue interface{}) (*SNOMEDMapping, error) {
	return nil, nil
}

// ApplySNOMEDMappingInternal 应用 SNOMED 映射到编码后的数据（内部通用函数）
// encoded: 编码后的数据 map；fieldName/fieldPath: 字段名与路径；rawValue: 原始值
// getMappingFunc: 获取 SNOMED 映射的函数（如 GetSNOMEDMappingByFieldPath）
// 有映射且 display_en 非空时写入 display_en，否则保留原始值
func ApplySNOMEDMappingInternal(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}, getMappingFunc func(string, interface{}) (*SNOMEDMapping, error)) {
	mapping, err := getMappingFunc(fieldPath, rawValue)
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

// ApplyRadarSNOMEDMapping 应用 Radar SNOMED 映射（便捷函数，备用）
func ApplyRadarSNOMEDMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	ApplySNOMEDMappingInternal(encoded, fieldName, fieldPath, rawValue, GetSNOMEDMappingByFieldPath)
}

// ApplySNOMEDMapping 应用 SNOMED 映射（向后兼容别名）
func ApplySNOMEDMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	ApplyRadarSNOMEDMapping(encoded, fieldName, fieldPath, rawValue)
}