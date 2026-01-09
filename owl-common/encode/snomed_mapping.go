package encode

// applySNOMedMappingInternal 应用 SNOMED 映射到编码后的数据（内部通用函数）
// encoded: 编码后的数据 map
// fieldName: 字段名称（用于生成 SNOMED 相关字段名）
// fieldPath: 字段路径（用于查询转换表）
// rawValue: 原始字段值
// getMappingFunc: 获取 SNOMED 映射的函数（可以是 GetSNOMEDMappingByFieldPath 或 GetSleepaceSNOMEDMappingByFieldPath）
func applySNOMedMappingInternal(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}, getMappingFunc func(string, interface{}) (*SNOMEDMapping, error)) {
	// 保留原始值
	encoded[fieldName] = rawValue

	// 获取 SNOMED 映射
	mapping, err := getMappingFunc(fieldPath, rawValue)
	if err != nil {
		// 如果没有找到映射，只保留原始值
		return
	}

	// 添加 SNOMED 映射字段
	encoded[fieldName+"_snomed_code"] = mapping.SNOMEDCode
	encoded[fieldName+"_snomed_display"] = mapping.SNOMEDDisplay
	encoded[fieldName+"_category"] = mapping.Category
	encoded[fieldName+"_display_en"] = mapping.DisplayEn
}

// applyRadarSNOMedMapping 应用 Radar SNOMED 映射（便捷函数）
func applyRadarSNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	applySNOMedMappingInternal(encoded, fieldName, fieldPath, rawValue, GetSNOMEDMappingByFieldPath)
}

// applySleepaceSNOMedMapping 应用 Sleepace SNOMED 映射（便捷函数）
func applySleepaceSNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	applySNOMedMappingInternal(encoded, fieldName, fieldPath, rawValue, GetSleepaceSNOMEDMappingByFieldPath)
}

// applySNOMedMapping 应用 Radar SNOMED 映射（Radar 编码器使用的便捷函数）
// 为了保持向后兼容，保留这个函数名
func applySNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}) {
	applyRadarSNOMedMapping(encoded, fieldName, fieldPath, rawValue)
}
