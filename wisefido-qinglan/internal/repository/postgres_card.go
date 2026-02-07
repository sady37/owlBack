package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// PostgresCardRepository PostgreSQL 实现的卡片仓库
type PostgresCardRepository struct {
	db *sql.DB
}

// NewPostgresCardRepository 创建 PostgreSQL 卡片仓库实例
func NewPostgresCardRepository(db *sql.DB) CardRepository {
	return &PostgresCardRepository{db: db}
}

// GetDeviceCardMappings 从 cards.devices JSONB 中提取所有设备-卡片映射
// 返回指定租户的所有设备关联的完整信息
func (r *PostgresCardRepository) GetDeviceCardMappings(ctx context.Context, tenantID string) ([]CardDeviceInfo, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// SQL 查询：使用 jsonb_to_recordset 展开设备数组
	// 直接查出扁平化的结果：device_id, device_uid, card_id, branch_id, tenant_id
	query := `
		SELECT 
			d.device_id,
			d.device_uid,
			c.card_id,
			c.branch_id,
			c.tenant_id
		FROM cards c,
			 jsonb_to_recordset(c.devices) AS d(device_id UUID, device_uid TEXT)
		WHERE c.tenant_id = $1
		  AND c.devices IS NOT NULL
		  AND jsonb_typeof(c.devices) = 'array'
		  AND jsonb_array_length(c.devices) > 0
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards and devices: %w", err)
	}
	defer rows.Close()

	var devices []CardDeviceInfo

	for rows.Next() {
		var deviceID sql.NullString
		var deviceUID sql.NullString
		var cardID string
		var branchID sql.NullString
		var tenantID string

		if err := rows.Scan(&deviceID, &deviceUID, &cardID, &branchID, &tenantID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 只添加有效的设备信息
		if deviceID.Valid && deviceID.String != "" && deviceUID.Valid && deviceUID.String != "" && branchID.Valid {
			devices = append(devices, CardDeviceInfo{
				DeviceID:  deviceID.String,
				DeviceUID: deviceUID.String,
				CardID:    cardID,
				BranchID:  branchID.String,
				TenantID:  tenantID,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return devices, nil
}

// GetDeviceCardMappingsByBranch 获取指定租户指定分支的设备-卡片映射（租户隔离）
// tenantID 和 branchID 都是必填的
// 注：使用 sql.NullString 处理 NULLABLE 列，但由于 tenant_id、card_id 都是 NOT NULL，
// 而 branch_id 是清除缓存的唯一凭证，所以当前的 Valid 检查是安全的
func (r *PostgresCardRepository) GetDeviceCardMappingsByBranch(ctx context.Context, tenantID string, branchID string) ([]CardDeviceInfo, error) {
	if tenantID == "" || branchID == "" {
		return nil, fmt.Errorf("tenant_id and branch_id are both required")
	}

	// SQL 查询：使用 jsonb_to_recordset 展开设备数组，限制于指定租户和分支
	query := `
		SELECT 
			d.device_id,
			d.device_uid,
			c.card_id,
			c.branch_id,
			c.tenant_id
		FROM cards c,
			 jsonb_to_recordset(c.devices) AS d(device_id UUID, device_uid TEXT)
		WHERE c.tenant_id = $1
		  AND c.branch_id = $2
		  AND c.devices IS NOT NULL
		  AND jsonb_typeof(c.devices) = 'array'
		  AND jsonb_array_length(c.devices) > 0
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards and devices by branch: %w", err)
	}
	defer rows.Close()

	var devices []CardDeviceInfo

	for rows.Next() {
		var deviceID sql.NullString
		var deviceUID sql.NullString
		var cardID string
		var branchIDVal sql.NullString
		var scannedTenantID string

		if err := rows.Scan(&deviceID, &deviceUID, &cardID, &branchIDVal, &scannedTenantID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 只添加有效的设备信息
		if deviceID.Valid && deviceID.String != "" && deviceUID.Valid && deviceUID.String != "" && branchIDVal.Valid {
			devices = append(devices, CardDeviceInfo{
				DeviceID:  deviceID.String,
				DeviceUID: deviceUID.String,
				CardID:    cardID,
				BranchID:  branchIDVal.String,
				TenantID:  scannedTenantID,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)

	}

	return devices, nil
}

// 确保实现了 CardRepository 接口
var _ CardRepository = (*PostgresCardRepository)(nil)
