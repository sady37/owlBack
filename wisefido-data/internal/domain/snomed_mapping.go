package domain

// SNOMEDMapping SNOMED编码映射领域模型（对应 snomed_mapping 表）
// 统一管理所有 SNOMED CT 编码映射，包括姿态映射和事件映射
type SNOMEDMapping struct {
	// 主键
	MappingID string `db:"mapping_id"` // UUID, PRIMARY KEY

	// 映射类型
	MappingType string `db:"mapping_type"` // VARCHAR(20), NOT NULL - 'posture'/'event'

	// 源值
	SourceValue string `db:"source_value"` // VARCHAR(50), NOT NULL

	// SNOMED CT 编码和显示名称
	SNOMEDCode    string `db:"snomed_code"`    // VARCHAR(50), nullable
	SNOMEDDisplay string `db:"snomed_display"`  // VARCHAR(100), NOT NULL

	// FHIR Category
	Category string `db:"category"` // VARCHAR(50), NOT NULL

	// LOINC 编码（用于 FHIR，可选）
	LOINCCode string `db:"loinc_code"` // VARCHAR(50), nullable

	// 显示名称（英文）
	DisplayEn string `db:"display_en"` // VARCHAR(100), nullable

	// 固件版本（用于姿态映射）
	FirmwareVersion string `db:"firmware_version"` // VARCHAR(50), nullable

	// 持续时间阈值（用于事件映射）
	DurationThresholdMinutes *int `db:"duration_threshold_minutes"` // INTEGER, nullable
}

