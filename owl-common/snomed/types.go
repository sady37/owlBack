package snomed

// SNOMEDMapping SNOMED 映射结构（统一定义，供中后台使用）
// 用于设备数据到 SNOMED 标准格式的转换
type SNOMEDMapping struct {
	SNOMEDCode    *string `json:"snomed_code"`     // SNOMED CT 编码
	SNOMEDDisplay string  `json:"snomed_display"`  // SNOMED CT 显示名称
	Category      string  `json:"category"`        // FHIR Category (safety, clinical, behavioral, device)
	DisplayEn     string  `json:"display_en"`       // 英文显示值（中后台统一使用此值）
}
