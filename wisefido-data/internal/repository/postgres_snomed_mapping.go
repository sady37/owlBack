package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
)

// PostgresSNOMEDMappingRepository SNOMED编码映射Repository实现（强类型版本）
type PostgresSNOMEDMappingRepository struct {
	db *sql.DB
}

// NewPostgresSNOMEDMappingRepository 创建SNOMED编码映射Repository
func NewPostgresSNOMEDMappingRepository(db *sql.DB) *PostgresSNOMEDMappingRepository {
	return &PostgresSNOMEDMappingRepository{db: db}
}

// 确保实现了接口
var _ SNOMEDMappingRepository = (*PostgresSNOMEDMappingRepository)(nil)

// GetMapping 获取映射（按mapping_type和source_value）
func (r *PostgresSNOMEDMappingRepository) GetMapping(ctx context.Context, mappingType, sourceValue string) (*domain.SNOMEDMapping, error) {
	if mappingType == "" || sourceValue == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			mapping_id::text,
			mapping_type,
			source_value,
			snomed_code,
			snomed_display,
			category,
			loinc_code,
			display_en,
			firmware_version,
			duration_threshold_minutes
		FROM snomed_mapping
		WHERE mapping_type = $1 AND source_value = $2
		LIMIT 1
	`

	var mapping domain.SNOMEDMapping
	var snomedCode, loincCode, displayEn, firmwareVersion sql.NullString
	var durationThreshold sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, mappingType, sourceValue).Scan(
		&mapping.MappingID,
		&mapping.MappingType,
		&mapping.SourceValue,
		&snomedCode,
		&mapping.SNOMEDDisplay,
		&mapping.Category,
		&loincCode,
		&displayEn,
		&firmwareVersion,
		&durationThreshold,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("snomed mapping not found: mapping_type=%s, source_value=%s", mappingType, sourceValue)
		}
		return nil, fmt.Errorf("failed to get snomed mapping: %w", err)
	}

	if snomedCode.Valid {
		mapping.SNOMEDCode = snomedCode.String
	}
	if loincCode.Valid {
		mapping.LOINCCode = loincCode.String
	}
	if displayEn.Valid {
		mapping.DisplayEn = displayEn.String
	}
	if firmwareVersion.Valid {
		mapping.FirmwareVersion = firmwareVersion.String
	}
	if durationThreshold.Valid {
		duration := int(durationThreshold.Int64)
		mapping.DurationThresholdMinutes = &duration
	}

	return &mapping, nil
}

// GetPostureMapping 获取姿态映射（支持固件版本）
func (r *PostgresSNOMEDMappingRepository) GetPostureMapping(ctx context.Context, sourceValue string, firmwareVersion *string) (*domain.SNOMEDMapping, error) {
	if sourceValue == "" {
		return nil, sql.ErrNoRows
	}

	// 如果指定了firmwareVersion，优先匹配该版本，否则匹配通用版本（firmware_version IS NULL）
	query := `
		SELECT 
			mapping_id::text,
			mapping_type,
			source_value,
			snomed_code,
			snomed_display,
			category,
			loinc_code,
			display_en,
			firmware_version,
			duration_threshold_minutes
		FROM snomed_mapping
		WHERE mapping_type = 'posture'
		  AND source_value = $1
		  AND (firmware_version IS NULL OR firmware_version = $2)
		ORDER BY firmware_version DESC NULLS LAST
		LIMIT 1
	`

	var mapping domain.SNOMEDMapping
	var snomedCode, loincCode, displayEn, firmwareVersionDB sql.NullString
	var durationThreshold sql.NullInt64

	var firmwareVersionValue interface{}
	if firmwareVersion != nil {
		firmwareVersionValue = *firmwareVersion
	}

	err := r.db.QueryRowContext(ctx, query, sourceValue, firmwareVersionValue).Scan(
		&mapping.MappingID,
		&mapping.MappingType,
		&mapping.SourceValue,
		&snomedCode,
		&mapping.SNOMEDDisplay,
		&mapping.Category,
		&loincCode,
		&displayEn,
		&firmwareVersionDB,
		&durationThreshold,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("posture mapping not found: source_value=%s, firmware_version=%v", sourceValue, firmwareVersion)
		}
		return nil, fmt.Errorf("failed to get posture mapping: %w", err)
	}

	if snomedCode.Valid {
		mapping.SNOMEDCode = snomedCode.String
	}
	if loincCode.Valid {
		mapping.LOINCCode = loincCode.String
	}
	if displayEn.Valid {
		mapping.DisplayEn = displayEn.String
	}
	if firmwareVersionDB.Valid {
		mapping.FirmwareVersion = firmwareVersionDB.String
	}
	if durationThreshold.Valid {
		duration := int(durationThreshold.Int64)
		mapping.DurationThresholdMinutes = &duration
	}

	return &mapping, nil
}

// GetEventMapping 获取事件映射
func (r *PostgresSNOMEDMappingRepository) GetEventMapping(ctx context.Context, sourceValue string) (*domain.SNOMEDMapping, error) {
	return r.GetMapping(ctx, "event", sourceValue)
}

// ListMappings 列表查询（支持按类型、category、固件版本过滤）
func (r *PostgresSNOMEDMappingRepository) ListMappings(ctx context.Context, mappingType string, filters *SNOMEDMappingFilters, page, size int) ([]*domain.SNOMEDMapping, int, error) {
	if mappingType == "" {
		return []*domain.SNOMEDMapping{}, 0, nil
	}

	where := []string{"mapping_type = $1"}
	args := []any{mappingType}
	argN := 2

	if filters != nil {
		if filters.Category != "" {
			where = append(where, fmt.Sprintf("category = $%d", argN))
			args = append(args, filters.Category)
			argN++
		}
		if filters.FirmwareVersion != "" {
			where = append(where, fmt.Sprintf("(firmware_version IS NULL OR firmware_version = $%d)", argN))
			args = append(args, filters.FirmwareVersion)
			argN++
		}
	}

	// 查询总数
	queryCount := `
		SELECT COUNT(*)
		FROM snomed_mapping
		WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count snomed mappings: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// 查询列表
	argsList := append(args, size, offset)
	query := `
		SELECT 
			mapping_id::text,
			mapping_type,
			source_value,
			snomed_code,
			snomed_display,
			category,
			loinc_code,
			display_en,
			firmware_version,
			duration_threshold_minutes
		FROM snomed_mapping
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY source_value
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list snomed mappings: %w", err)
	}
	defer rows.Close()

	var mappings []*domain.SNOMEDMapping
	for rows.Next() {
		var mapping domain.SNOMEDMapping
		var snomedCode, loincCode, displayEn, firmwareVersion sql.NullString
		var durationThreshold sql.NullInt64

		if err := rows.Scan(
			&mapping.MappingID,
			&mapping.MappingType,
			&mapping.SourceValue,
			&snomedCode,
			&mapping.SNOMEDDisplay,
			&mapping.Category,
			&loincCode,
			&displayEn,
			&firmwareVersion,
			&durationThreshold,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan snomed mapping: %w", err)
		}

		if snomedCode.Valid {
			mapping.SNOMEDCode = snomedCode.String
		}
		if loincCode.Valid {
			mapping.LOINCCode = loincCode.String
		}
		if displayEn.Valid {
			mapping.DisplayEn = displayEn.String
		}
		if firmwareVersion.Valid {
			mapping.FirmwareVersion = firmwareVersion.String
		}
		if durationThreshold.Valid {
			duration := int(durationThreshold.Int64)
			mapping.DurationThresholdMinutes = &duration
		}

		mappings = append(mappings, &mapping)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate snomed mappings: %w", err)
	}

	return mappings, total, nil
}

// CreateMapping 创建映射
func (r *PostgresSNOMEDMappingRepository) CreateMapping(ctx context.Context, mapping *domain.SNOMEDMapping) error {
	if mapping.MappingType == "" || mapping.SourceValue == "" {
		return fmt.Errorf("mapping_type and source_value are required")
	}
	if mapping.SNOMEDDisplay == "" {
		return fmt.Errorf("snomed_display is required")
	}
	if mapping.Category == "" {
		return fmt.Errorf("category is required")
	}

	query := `
		INSERT INTO snomed_mapping (
			mapping_type,
			source_value,
			snomed_code,
			snomed_display,
			category,
			loinc_code,
			display_en,
			firmware_version,
			duration_threshold_minutes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (mapping_type, source_value) DO UPDATE SET
			snomed_code = EXCLUDED.snomed_code,
			snomed_display = EXCLUDED.snomed_display,
			category = EXCLUDED.category,
			loinc_code = EXCLUDED.loinc_code,
			display_en = EXCLUDED.display_en,
			firmware_version = EXCLUDED.firmware_version,
			duration_threshold_minutes = EXCLUDED.duration_threshold_minutes
		RETURNING mapping_id::text
	`

	var snomedCode, loincCode, displayEn, firmwareVersion interface{}
	if mapping.SNOMEDCode != "" {
		snomedCode = mapping.SNOMEDCode
	}
	if mapping.LOINCCode != "" {
		loincCode = mapping.LOINCCode
	}
	if mapping.DisplayEn != "" {
		displayEn = mapping.DisplayEn
	}
	if mapping.FirmwareVersion != "" {
		firmwareVersion = mapping.FirmwareVersion
	}

	var durationThreshold interface{}
	if mapping.DurationThresholdMinutes != nil {
		durationThreshold = *mapping.DurationThresholdMinutes
	}

	var mappingID string
	err := r.db.QueryRowContext(ctx, query, mapping.MappingType, mapping.SourceValue,
		snomedCode, mapping.SNOMEDDisplay, mapping.Category, loincCode, displayEn,
		firmwareVersion, durationThreshold).Scan(&mappingID)
	if err != nil {
		return fmt.Errorf("failed to create snomed mapping: %w", err)
	}

	mapping.MappingID = mappingID
	return nil
}

// UpdateMapping 更新映射
func (r *PostgresSNOMEDMappingRepository) UpdateMapping(ctx context.Context, mappingID string, mapping *domain.SNOMEDMapping) error {
	if mappingID == "" {
		return fmt.Errorf("mapping_id is required")
	}

	query := `
		UPDATE snomed_mapping
		SET
			snomed_code = $2,
			snomed_display = $3,
			category = $4,
			loinc_code = $5,
			display_en = $6,
			firmware_version = $7,
			duration_threshold_minutes = $8
		WHERE mapping_id = $1
	`

	var snomedCode, loincCode, displayEn, firmwareVersion interface{}
	if mapping.SNOMEDCode != "" {
		snomedCode = mapping.SNOMEDCode
	}
	if mapping.LOINCCode != "" {
		loincCode = mapping.LOINCCode
	}
	if mapping.DisplayEn != "" {
		displayEn = mapping.DisplayEn
	}
	if mapping.FirmwareVersion != "" {
		firmwareVersion = mapping.FirmwareVersion
	}

	var durationThreshold interface{}
	if mapping.DurationThresholdMinutes != nil {
		durationThreshold = *mapping.DurationThresholdMinutes
	}

	result, err := r.db.ExecContext(ctx, query, mappingID, snomedCode, mapping.SNOMEDDisplay,
		mapping.Category, loincCode, displayEn, firmwareVersion, durationThreshold)
	if err != nil {
		return fmt.Errorf("failed to update snomed mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("snomed mapping not found")
	}

	return nil
}

// DeleteMapping 删除映射
func (r *PostgresSNOMEDMappingRepository) DeleteMapping(ctx context.Context, mappingID string) error {
	if mappingID == "" {
		return fmt.Errorf("mapping_id is required")
	}

	query := `
		DELETE FROM snomed_mapping
		WHERE mapping_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, mappingID)
	if err != nil {
		return fmt.Errorf("failed to delete snomed mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("snomed mapping not found")
	}

	return nil
}

