package decode

// applySNOMedMappingInternal 应用 SNOMED 映射到编码后的数据（内部通用函数）
// encoded: 编码后的数据 map
// fieldName: 字段名称
// fieldPath: 字段路径（用于查询转换表）
// rawValue: 原始字段值
// getMappingFunc: 获取 SNOMED 映射的函数（GetSNOMEDMappingByFieldPath）
//
// 内部使用简化格式：
// - 字段名保持不变：{fieldName}
// - 字段值使用 display_en（如果有映射且 display_en 不为空），否则使用原始值
// - 不添加额外的 SNOMED 相关字段（_snomed_code, _snomed_display, _category, _display_en）
func applySNOMedMappingInternal(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}, getMappingFunc func(string, interface{}) (*SNOMEDMapping, error)) {
	// 获取 SNOMED 映射
	mapping, err := getMappingFunc(fieldPath, rawValue)
	if err != nil || mapping == nil {
		// 如果没有找到映射，使用原始值
		encoded[fieldName] = rawValue
		return
	}

	// 使用 display_en 作为字段值（内部使用简化格式）
	// 如果 display_en 为空，使用原始值
	if mapping.DisplayEn != "" {
		encoded[fieldName] = mapping.DisplayEn
	} else {
		// display_en 为空时，保留原始值（如 stability 的 bit 值 "11"）
		encoded[fieldName] = rawValue
	}
}

// applyRadarSNOMedMapping 应用 Radar SNOMED 映射（便捷函数）
func applyRadarSNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	applySNOMedMappingInternal(encoded, fieldName, fieldPath, rawValue, GetSNOMEDMappingByFieldPath)
}

// applySNOMedMapping 应用 Radar SNOMED 映射（Radar 编码器使用的便捷函数）
// 为了保持向后兼容，保留这个函数名
func applySNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	applyRadarSNOMedMapping(encoded, fieldName, fieldPath, rawValue)
}
