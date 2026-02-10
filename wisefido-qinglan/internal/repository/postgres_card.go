package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// PostgresCardRepository 实现 CardRepository
type PostgresCardRepository struct {
	db *sql.DB
}

// NewPostgresCardRepository 创建Postgres Card repository
func NewPostgresCardRepository(db *sql.DB) *PostgresCardRepository {
	return &PostgresCardRepository{db: db}
}

// parseDevicesColumn 尝试将cards.devices列解析为字符串切片，支持 JSON 数组 和 Postgres text[] 表示法
func parseDevicesColumn(raw sql.NullString) []string {
	if !raw.Valid {
		return nil
	}
	s := raw.String
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err == nil {
		return arr
	}
	// fallback: Postgres text[] like {a,b}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.Trim(parts[i], " \"'")
	}
	return parts
}

// resolveDevice 查询 device_store 表，支持传入 device_id 或 device_uid
// batchResolveByUID 批量解析一组 device_uid，返回 map[device_uid]device_id（device_id 可能为空）
func (r *PostgresCardRepository) batchResolveByUID(ctx context.Context, uids []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(uids) == 0 {
		return result, nil
	}

	// 去重
	uniq := make(map[string]struct{})
	vals := make([]string, 0, len(uids))
	for _, v := range uids {
		if v == "" {
			continue
		}
		if _, ok := uniq[v]; ok {
			continue
		}
		uniq[v] = struct{}{}
		vals = append(vals, v)
	}
	if len(vals) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(vals))
	args := make([]interface{}, len(vals))
	for i, v := range vals {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = v
	}
	ph := strings.Join(placeholders, ",")

	query := fmt.Sprintf("SELECT device_uid, device_id FROM device_store WHERE device_uid IN (%s)", ph)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch query device_store by uid: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var duid, did sql.NullString
		if err := rows.Scan(&duid, &did); err != nil {
			return nil, fmt.Errorf("failed to scan device_store row: %w", err)
		}
		if duid.Valid {
			result[duid.String] = did.String
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("device_store rows iteration error: %w", err)
	}
	return result, nil
}


// GetDeviceCardMappings 全量获取所有设备与卡片的映射关系
func (r *PostgresCardRepository) GetDeviceCardMappings(ctx context.Context) ([]CardDeviceInfo, error) {
	// 移除了 WHERE 子句中对 tenant_id 和 branch_id 的过滤
	query := `
		SELECT 
			devices_data.device_uid, 
			c.tenant_id, 
			c.branch_id, 
			devices_data.device_id, 
			c.card_id::text
		FROM cards c, 
		     jsonb_to_recordset(c.devices) AS devices_data(
		         device_id uuid, 
		         device_uid varchar, 
		         device_name varchar, 
		         device_type varchar, 
		         device_model varchar, 
		         bed_id uuid, 
		         room_id uuid, 
		         unit_id varchar
		     )
		WHERE c.branch_id IS NOT NULL
		  AND devices_data.device_uid IS NOT NULL
		  AND devices_data.device_uid != ''
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all device card mappings: %w", err)
	}
	defer rows.Close()

	var results []CardDeviceInfo

	for rows.Next() {
		var cdi CardDeviceInfo
		if err := rows.Scan(&cdi.DeviceUID, &cdi.TenantID, &cdi.BranchID, &cdi.DeviceID, &cdi.CardID); err != nil {
			return nil, fmt.Errorf("failed to scan device info: %w", err)
		}
		results = append(results, cdi)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetDeviceListByBranch 获取特定租户和分支下的设备-卡片映射
func (r *PostgresCardRepository) GetDeviceCardMappingsByBranch(ctx context.Context, tenantID, branchID string) ([]CardDeviceInfo, error) {
	// 严格匹配传入的 tenant_id 和 branch_id
	query := `
		SELECT 
			devices_data.device_uid, 
			c.tenant_id, 
			c.branch_id, 
			devices_data.device_id, 
			c.card_id::text
		FROM cards c, 
		     jsonb_to_recordset(c.devices) AS devices_data(
		         device_id uuid, 
		         device_uid varchar, 
		         device_name varchar, 
		         device_type varchar, 
		         device_model varchar, 
		         bed_id uuid, 
		         room_id uuid, 
		         unit_id varchar
		     )
		WHERE c.tenant_id = $1 
		  AND c.branch_id = $2
		  AND devices_data.device_uid IS NOT NULL
		  AND devices_data.device_uid != ''
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices for branch %s: %w", branchID, err)
	}
	defer rows.Close()

	var results []CardDeviceInfo
	for rows.Next() {
		var cdi CardDeviceInfo
		if err := rows.Scan(&cdi.DeviceUID, &cdi.TenantID, &cdi.BranchID, &cdi.DeviceID, &cdi.CardID); err != nil {
			return nil, fmt.Errorf("failed to scan device info: %w", err)
		}
		results = append(results, cdi)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}