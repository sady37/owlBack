package snomed

// SNOMEDMapping SNOMED 映射结构（统一定义，供中后台使用）
// 用于设备数据到 SNOMED 标准格式的转换
//
// 统一约定：display_en / snomed_display 全小写；SNOMED 仅在与外部系统对接时使用。
// 数据源：radar_convert_table.json（Radar）、sleepace_convert_table.json（Sleepace）。
type SNOMEDMapping struct {
	SNOMEDCode    *string `json:"snomed_code"`    // SNOMED CT 编码（对外对接用）
	SNOMEDDisplay string  `json:"snomed_display"` // SNOMED 展示名，全小写
	Category      string  `json:"category"`       // FHIR Category (activity, vital-signs, safety, clinical, behavioral, device)
	DisplayEn     string  `json:"display_en"`     // 英文展示值，全小写，对内/前端用
}