package repository

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
)

// SNOMEDRepository SNOMED 映射仓库
type SNOMEDRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSNOMEDRepository 创建 SNOMED 映射仓库
func NewSNOMEDRepository(db *sql.DB, logger *zap.Logger) *SNOMEDRepository {
	return &SNOMEDRepository{
		db:     db,
		logger: logger,
	}
}

// PostureMapping 姿态映射结果
type PostureMapping struct {
	SNOMEDCode    *string
	SNOMEDDisplay string
	Category      string
}

// GetPostureMapping 获取姿态映射
// sourceValue: 设备原始姿态值（如 "0", "1", "2", ...）
// firmwareVersion: 固件版本（可选，用于版本特定的映射）
func (r *SNOMEDRepository) GetPostureMapping(sourceValue string, firmwareVersion *string) (*PostureMapping, error) {
	query := `
		SELECT 
			snomed_code,
			snomed_display,
			category
		FROM snomed_mapping
		WHERE mapping_type = 'posture'
		  AND source_value = $1
		  AND (firmware_version IS NULL OR firmware_version = $2)
		ORDER BY firmware_version DESC NULLS LAST
		LIMIT 1
	`
	
	mapping := &PostureMapping{}
	var snomedCode sql.NullString
	
	err := r.db.QueryRow(query, sourceValue, firmwareVersion).Scan(
		&snomedCode,
		&mapping.SNOMEDDisplay,
		&mapping.Category,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("posture mapping not found: source_value=%s, firmware_version=%v", sourceValue, firmwareVersion)
		}
		return nil, fmt.Errorf("failed to query posture mapping: %w", err)
	}
	
	if snomedCode.Valid {
		mapping.SNOMEDCode = &snomedCode.String
	}
	
	return mapping, nil
}

// EventMapping 事件映射结果
type EventMapping struct {
	SNOMEDCode    *string
	SNOMEDDisplay string
	Category      string
}

// GetEventMapping 获取事件映射
// sourceValue: 标准事件类型标识符（如 "ENTER_ROOM", "LEFT_BED", "FALL"）
func (r *SNOMEDRepository) GetEventMapping(sourceValue string) (*EventMapping, error) {
	query := `
		SELECT 
			snomed_code,
			snomed_display,
			category
		FROM snomed_mapping
		WHERE mapping_type = 'event'
		  AND source_value = $1
		LIMIT 1
	`
	
	mapping := &EventMapping{}
	var snomedCode sql.NullString
	
	err := r.db.QueryRow(query, sourceValue).Scan(
		&snomedCode,
		&mapping.SNOMEDDisplay,
		&mapping.Category,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event mapping not found: source_value=%s", sourceValue)
		}
		return nil, fmt.Errorf("failed to query event mapping: %w", err)
	}
	
	if snomedCode.Valid {
		mapping.SNOMEDCode = &snomedCode.String
	}
	
	return mapping, nil
}

