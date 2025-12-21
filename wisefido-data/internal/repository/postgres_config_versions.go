package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresConfigVersionsRepository 配置版本Repository实现（强类型版本）
type PostgresConfigVersionsRepository struct {
	db *sql.DB
}

// NewPostgresConfigVersionsRepository 创建配置版本Repository
func NewPostgresConfigVersionsRepository(db *sql.DB) *PostgresConfigVersionsRepository {
	return &PostgresConfigVersionsRepository{db: db}
}

// 确保实现了接口
var _ ConfigVersionsRepository = (*PostgresConfigVersionsRepository)(nil)

// GetConfigVersion 获取配置版本
func (r *PostgresConfigVersionsRepository) GetConfigVersion(ctx context.Context, tenantID, versionID string) (*domain.ConfigVersion, error) {
	if tenantID == "" || versionID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			version_id::text,
			tenant_id::text,
			config_type,
			entity_id::text,
			current_entity_id::text,
			config_data,
			valid_from,
			valid_to
		FROM config_versions
		WHERE tenant_id = $1 AND version_id = $2
	`

	var configVersion domain.ConfigVersion
	var currentEntityID sql.NullString
	var configData sql.NullString
	var validTo sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, versionID).Scan(
		&configVersion.VersionID,
		&configVersion.TenantID,
		&configVersion.ConfigType,
		&configVersion.EntityID,
		&currentEntityID,
		&configData,
		&configVersion.ValidFrom,
		&validTo,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("config version not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get config version: %w", err)
	}

	if currentEntityID.Valid {
		configVersion.CurrentEntityID = currentEntityID.String
	}
	if configData.Valid {
		configVersion.ConfigData = []byte(configData.String)
	}
	if validTo.Valid {
		configVersion.ValidTo = &validTo.Time
	}

	return &configVersion, nil
}

// GetConfigVersionAtTime 查询某个时间点的配置（用于回放）
func (r *PostgresConfigVersionsRepository) GetConfigVersionAtTime(ctx context.Context, tenantID, configType, entityID string, atTime time.Time) (*domain.ConfigVersion, error) {
	if tenantID == "" || configType == "" || entityID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			version_id::text,
			tenant_id::text,
			config_type,
			entity_id::text,
			current_entity_id::text,
			config_data,
			valid_from,
			valid_to
		FROM config_versions
		WHERE tenant_id = $1 
			AND config_type = $2 
			AND entity_id = $3
			AND valid_from <= $4
			AND (valid_to IS NULL OR valid_to > $4)
		ORDER BY valid_from DESC
		LIMIT 1
	`

	var configVersion domain.ConfigVersion
	var currentEntityID sql.NullString
	var configData sql.NullString
	var validTo sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, configType, entityID, atTime).Scan(
		&configVersion.VersionID,
		&configVersion.TenantID,
		&configVersion.ConfigType,
		&configVersion.EntityID,
		&currentEntityID,
		&configData,
		&configVersion.ValidFrom,
		&validTo,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("config version not found at time %v: %w", atTime, err)
		}
		return nil, fmt.Errorf("failed to get config version at time: %w", err)
	}

	if currentEntityID.Valid {
		configVersion.CurrentEntityID = currentEntityID.String
	}
	if configData.Valid {
		configVersion.ConfigData = []byte(configData.String)
	}
	if validTo.Valid {
		configVersion.ValidTo = &validTo.Time
	}

	return &configVersion, nil
}

// ListConfigVersions 查询某个实体的所有配置历史（支持分页、时间范围过滤）
func (r *PostgresConfigVersionsRepository) ListConfigVersions(ctx context.Context, tenantID, configType, entityID string, filters *ConfigVersionFilters, page, size int) ([]*domain.ConfigVersion, int, error) {
	if tenantID == "" || configType == "" || entityID == "" {
		return []*domain.ConfigVersion{}, 0, nil
	}

	where := []string{"tenant_id = $1", "config_type = $2", "entity_id = $3"}
	args := []any{tenantID, configType, entityID}
	argN := 4

	if filters != nil {
		if filters.StartTime != nil {
			where = append(where, fmt.Sprintf("valid_from >= $%d", argN))
			args = append(args, *filters.StartTime)
			argN++
		}
		if filters.EndTime != nil {
			where = append(where, fmt.Sprintf("(valid_to IS NULL OR valid_to <= $%d)", argN))
			args = append(args, *filters.EndTime)
			argN++
		}
	}

	// 查询总数
	queryCount := `
		SELECT COUNT(*)
		FROM config_versions
		WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count config versions: %w", err)
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
			version_id::text,
			tenant_id::text,
			config_type,
			entity_id::text,
			current_entity_id::text,
			config_data,
			valid_from,
			valid_to
		FROM config_versions
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY valid_from DESC
		LIMIT $` + fmt.Sprintf("%d", argN) + ` OFFSET $` + fmt.Sprintf("%d", argN+1)

	rows, err := r.db.QueryContext(ctx, query, argsList...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list config versions: %w", err)
	}
	defer rows.Close()

	var configVersions []*domain.ConfigVersion
	for rows.Next() {
		var configVersion domain.ConfigVersion
		var currentEntityID sql.NullString
		var configData sql.NullString
		var validTo sql.NullTime

		if err := rows.Scan(
			&configVersion.VersionID,
			&configVersion.TenantID,
			&configVersion.ConfigType,
			&configVersion.EntityID,
			&currentEntityID,
			&configData,
			&configVersion.ValidFrom,
			&validTo,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan config version: %w", err)
		}

		if currentEntityID.Valid {
			configVersion.CurrentEntityID = currentEntityID.String
		}
		if configData.Valid {
			configVersion.ConfigData = []byte(configData.String)
		}
		if validTo.Valid {
			configVersion.ValidTo = &validTo.Time
		}

		configVersions = append(configVersions, &configVersion)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate config versions: %w", err)
	}

	return configVersions, total, nil
}

// CreateConfigVersion 创建新版本
// 自动设置valid_from，将旧版本的valid_to设置为当前时间
func (r *PostgresConfigVersionsRepository) CreateConfigVersion(ctx context.Context, tenantID string, configVersion *domain.ConfigVersion) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if configVersion.ConfigType == "" {
		return "", fmt.Errorf("config_type is required")
	}
	if configVersion.EntityID == "" {
		return "", fmt.Errorf("entity_id is required")
	}
	if len(configVersion.ConfigData) == 0 {
		return "", fmt.Errorf("config_data is required")
	}
	if configVersion.ValidFrom.IsZero() {
		configVersion.ValidFrom = time.Now()
	}

	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 将旧版本的valid_to设置为当前时间
	updateOldQuery := `
		UPDATE config_versions
		SET valid_to = $4
		WHERE tenant_id = $1 
			AND config_type = $2 
			AND entity_id = $3
			AND (valid_to IS NULL OR valid_to > $4)
	`
	_, err = tx.ExecContext(ctx, updateOldQuery, tenantID, configVersion.ConfigType, configVersion.EntityID, configVersion.ValidFrom)
	if err != nil {
		return "", fmt.Errorf("failed to update old config versions: %w", err)
	}

	// 创建新版本
	insertQuery := `
		INSERT INTO config_versions (
			tenant_id,
			config_type,
			entity_id,
			current_entity_id,
			config_data,
			valid_from,
			valid_to
		) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
		RETURNING version_id::text
	`

	var currentEntityID interface{}
	if configVersion.CurrentEntityID != "" {
		currentEntityID = configVersion.CurrentEntityID
	}

	var validTo interface{}
	if configVersion.ValidTo != nil {
		validTo = *configVersion.ValidTo
	}

	var versionID string
	err = tx.QueryRowContext(ctx, insertQuery, tenantID, configVersion.ConfigType, configVersion.EntityID,
		currentEntityID, string(configVersion.ConfigData), configVersion.ValidFrom, validTo).Scan(&versionID)
	if err != nil {
		return "", fmt.Errorf("failed to create config version: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return versionID, nil
}

// UpdateConfigVersion 更新配置版本
func (r *PostgresConfigVersionsRepository) UpdateConfigVersion(ctx context.Context, tenantID, versionID string, configVersion *domain.ConfigVersion) error {
	if tenantID == "" || versionID == "" {
		return fmt.Errorf("tenant_id and version_id are required")
	}

	query := `
		UPDATE config_versions
		SET
			current_entity_id = $3,
			config_data = $4::jsonb,
			valid_from = $5,
			valid_to = $6
		WHERE tenant_id = $1 AND version_id = $2
	`

	var currentEntityID interface{}
	if configVersion.CurrentEntityID != "" {
		currentEntityID = configVersion.CurrentEntityID
	}

	var validTo interface{}
	if configVersion.ValidTo != nil {
		validTo = *configVersion.ValidTo
	}

	var configData interface{}
	if len(configVersion.ConfigData) > 0 {
		configData = string(configVersion.ConfigData)
	} else {
		configData = "{}"
	}

	result, err := r.db.ExecContext(ctx, query, tenantID, versionID, currentEntityID, configData,
		configVersion.ValidFrom, validTo)
	if err != nil {
		return fmt.Errorf("failed to update config version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("config version not found")
	}

	return nil
}

// DeleteConfigVersion 删除配置版本
func (r *PostgresConfigVersionsRepository) DeleteConfigVersion(ctx context.Context, tenantID, versionID string) error {
	if tenantID == "" || versionID == "" {
		return fmt.Errorf("tenant_id and version_id are required")
	}

	query := `
		DELETE FROM config_versions
		WHERE tenant_id = $1 AND version_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, versionID)
	if err != nil {
		return fmt.Errorf("failed to delete config version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("config version not found")
	}

	return nil
}

