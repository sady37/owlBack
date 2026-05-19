package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ConfigVersionRepository 配置版本仓库
type ConfigVersionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewConfigVersionRepository 创建配置版本仓库
func NewConfigVersionRepository(db *sql.DB, logger *zap.Logger) *ConfigVersionRepository {
	return &ConfigVersionRepository{
		db:     db,
		logger: logger,
	}
}

// GetBedAreaID 获取床的区域ID（area_id）
// 从 config_version 表的 room_layout 配置中提取
func (r *ConfigVersionRepository) GetBedAreaID(
	ctx context.Context,
	tenantID, roomID string,
) (int, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}
	if roomID == "" {
		return 0, fmt.Errorf("room_id is required")
	}

	// 查询当前生效的 room_layout 配置
	query := `
		SELECT config_data
		FROM config_versions
		WHERE tenant_id = $1
		  AND config_type = 'room_layout'
		  AND entity_id = $2
		  AND valid_from <= NOW()
		  AND (valid_to IS NULL OR valid_to > NOW())
		ORDER BY valid_from DESC
		LIMIT 1
	`

	var configDataJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, roomID).Scan(&configDataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("room layout not found for room_id: %s", roomID)
		}
		return 0, fmt.Errorf("failed to query room layout: %w", err)
	}

	// 解析 JSONB 配置数据
	var configData map[string]interface{}
	if err := json.Unmarshal(configDataJSON, &configData); err != nil {
		return 0, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	// 从 layout 中提取床的 area_id
	// 注意：layout 结构需要根据实际前端数据结构来确定
	// 这里假设 layout 中有 objects 数组，每个 object 有 type="bed" 和 area_id
	layout, ok := configData["layout"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("layout not found in config data")
	}

	objects, ok := layout["objects"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("objects not found in layout")
	}

	// 查找床对象（type="bed"）
	for _, obj := range objects {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			continue
		}

		objType, ok := objMap["type"].(string)
		if !ok || objType != "bed" {
			continue
		}

		// 提取 area_id
		areaID, ok := objMap["area_id"]
		if !ok {
			continue
		}

		// 转换为 int
		var areaIDInt int
		switch v := areaID.(type) {
		case float64:
			areaIDInt = int(v)
		case int:
			areaIDInt = v
		case int64:
			areaIDInt = int(v)
		default:
			continue
		}

		if areaIDInt > 0 {
			r.logger.Debug("Found bed area_id from layout",
				zap.String("room_id", roomID),
				zap.Int("area_id", areaIDInt),
			)
			return areaIDInt, nil
		}
	}

	return 0, fmt.Errorf("bed area_id not found in layout for room_id: %s", roomID)
}

