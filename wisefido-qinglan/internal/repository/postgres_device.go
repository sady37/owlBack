package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"
	"sync"

	"owl-common/alarm"
	"wisefido-qinglan/internal/domain"
)

// PostgresDeviceRepository 设备 Repository 的 PostgreSQL 实现（v2）
//
// v2 数据源（详 owlRD/dbv2/21_device_factory_meta.sql / 22_device_runtime_state.sql / 23_devices.sql）：
//   - device_factory_meta (dfm)：出厂元数据（不变）device_uid/device_type/device_model/imei/comm_mode/mcu_model
//   - device_runtime_state (drs)：运行时（高频 UPDATE）online/firmware_version/sensor_detached/rssi
//   - devices (d)：业务绑定（PK = device_ipv6 INET /128） access/monitoring_enabled
//
// 1:1 关系（device_id UUID 三表 join）；UID/Type 永远来自 dfm；online/firmware 来自 drs。
type PostgresDeviceRepository struct {
	db              *sql.DB
	enablementCache *sync.Map // legacy 缓存槽位（v2 spatial_config 迁移后可移除）
}

// NewPostgresDeviceRepository 创建 PostgreSQL 设备 Repository
func NewPostgresDeviceRepository(db *sql.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{
		db:              db,
		enablementCache: &sync.Map{},
	}
}

var _ DeviceRepository = (*PostgresDeviceRepository)(nil)

// systemTenantPrefix v2 系统 tenant /48 池（详 owlRD/dbv2/90_seed_tenants.sql）
// 自动注册的未授权设备落入此池，待 platform_admin 调拨。
const systemTenantPrefix = "fd00:0:1::/48"

// deviceSelectV2 v2 三表 join 的标准列出参顺序；scanDeviceRow 与之配对。
const deviceSelectV2 = `
		dfm.device_id::text                                              AS device_id,
		dfm.device_uid                                                    AS device_uid,
		COALESCE(host(d.device_ipv6), '')                                 AS device_ipv6,
		COALESCE(host(network(set_masklen(d.device_ipv6, 48))), '')       AS tenant_id,
		COALESCE(dfm.device_code, dfm.device_uid)                         AS device_name,
		CASE WHEN d.device_ipv6 IS NOT NULL
		     THEN host(network(set_masklen(d.device_ipv6, 88))) || '/88'
		     ELSE NULL END                                                AS bound_room_id,
		CASE WHEN d.device_ipv6 IS NOT NULL
		     THEN host(network(set_masklen(d.device_ipv6, 96))) || '/96'
		     ELSE NULL END                                                AS bound_bed_id,
		CASE WHEN COALESCE(drs.online, false) THEN 'online' ELSE 'offline' END AS status,
		CASE WHEN COALESCE(d.access, false) THEN 'approved' ELSE 'pending' END AS business_access,
		COALESCE(d.access, false)                                          AS allow_access,
		COALESCE(d.monitoring_enabled, false)                              AS monitoring_enabled,
		dfm.device_type::text                                              AS device_type,
		dfm.device_model                                                   AS device_model,
		dfm.imei                                                           AS imei,
		dfm.comm_mode                                                      AS comm_mode,
		dfm.mcu_model                                                      AS mcu_model,
		drs.firmware_version                                               AS firmware_version
	FROM device_factory_meta dfm
	LEFT JOIN devices d ON d.device_id = dfm.device_id
	LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id`

// scanDeviceRow 与 deviceSelectV2 列序对齐的 scan helper
func scanDeviceRow(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Device, error) {
	var d domain.Device
	var deviceIPv6, status, businessAccess string
	var boundRoomID, boundBedID sql.NullString
	var deviceType sql.NullString
	var deviceModel, imei, commMode, mcuModel, firmwareVersion sql.NullString
	err := row.Scan(
		&d.DeviceID,
		&d.DeviceUID,
		&deviceIPv6,
		&d.TenantID,
		&d.DeviceName,
		&boundRoomID,
		&boundBedID,
		&status,
		&businessAccess,
		&d.AllowAccess,
		&d.MonitoringEnabled,
		&deviceType,
		&deviceModel,
		&imei,
		&commMode,
		&mcuModel,
		&firmwareVersion,
	)
	if err != nil {
		return nil, err
	}
	d.DeviceIPv6 = deviceIPv6
	d.BoundRoomID = boundRoomID
	d.BoundBedID = boundBedID
	d.Status = status
	d.BusinessAccess = businessAccess
	d.DeviceType = deviceType
	d.DeviceModel = deviceModel
	d.IMEI = imei
	d.CommMode = commMode
	d.MCUModel = mcuModel
	d.FirmwareVersion = firmwareVersion
	return &d, nil
}

// GetDeviceByUID v2：dfm.device_uid 查询
func (r *PostgresDeviceRepository) GetDeviceByUID(ctx context.Context, uid string) (*domain.Device, error) {
	query := `SELECT ` + deviceSelectV2 + `
		WHERE dfm.device_uid = $1
		LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, uid)
	d, err := scanDeviceRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", uid)
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}
	return d, nil
}

// UpdateDeviceMonitoring v2：UPDATE devices.monitoring_enabled WHERE device_id (派生自 dfm.device_uid)
func (r *PostgresDeviceRepository) UpdateDeviceMonitoring(ctx context.Context, uid string, enabled bool) error {
	query := `
		UPDATE devices d
		SET monitoring_enabled = $1, updated_at = NOW()
		FROM device_factory_meta dfm
		WHERE d.device_id = dfm.device_id AND dfm.device_uid = $2`
	result, err := r.db.ExecContext(ctx, query, enabled, uid)
	if err != nil {
		return fmt.Errorf("failed to update device monitoring: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("device not found: %s", uid)
	}
	return nil
}

// GetDeviceProperties v2：firmware_version / sensor_detached 等 drs 字段；不再有 JSONB metadata
//
// 兼容 v1 调用：返回 drs.firmware_version 等已知键；未知键返回 nil。
func (r *PostgresDeviceRepository) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	query := `
		SELECT drs.firmware_version, dfm.mcu_model, dfm.device_model, dfm.imei, dfm.comm_mode
		FROM device_factory_meta dfm
		LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id
		WHERE dfm.device_uid = $1
		LIMIT 1`
	var fw, mcuModel, deviceModel, imei, commMode sql.NullString
	if err := r.db.QueryRowContext(ctx, query, uid).Scan(&fw, &mcuModel, &deviceModel, &imei, &commMode); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", uid)
		}
		return nil, fmt.Errorf("failed to query device properties: %w", err)
	}
	all := map[string]interface{}{
		"firmware_version": nullableString(fw),
		"mcu_model":        nullableString(mcuModel),
		"device_model":     nullableString(deviceModel),
		"imei":             nullableString(imei),
		"comm_mode":        nullableString(commMode),
	}
	if len(keys) == 0 {
		return all, nil
	}
	out := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		if v, ok := all[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

// SetDeviceProperties v2：写 device_runtime_state（firmware_version 等）
//
// v1 把任意 JSON 写 devices.metadata；v2 不再有该列。已知键写 drs，未知键忽略并 warn。
func (r *PostgresDeviceRepository) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
	if len(properties) == 0 {
		return nil
	}
	// 解析支持的键
	var fw sql.NullString
	for k, v := range properties {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if k == "firmware_version" {
			fw = sql.NullString{String: s, Valid: true}
		}
		// 其它键暂不持久化（mcu_model/imei/comm_mode/device_model 由 device_factory_meta 出厂时写入）
	}
	if !fw.Valid {
		// 没有 v2 持久化字段需要更新，noop
		return nil
	}
	// UPSERT device_runtime_state
	query := `
		INSERT INTO device_runtime_state (device_id, firmware_version, updated_at)
		SELECT dfm.device_id, $2, NOW()
		FROM device_factory_meta dfm
		WHERE dfm.device_uid = $1
		ON CONFLICT (device_id) DO UPDATE
		SET firmware_version = EXCLUDED.firmware_version, updated_at = NOW()`
	result, err := r.db.ExecContext(ctx, query, uid, fw.String)
	if err != nil {
		return fmt.Errorf("failed to upsert device runtime state: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("device not found: %s", uid)
	}
	return nil
}

// GetAllDeviceStoreInfo v2：从 dfm + drs + devices 派生，启动时拉全集
func (r *PostgresDeviceRepository) GetAllDeviceStoreInfo(ctx context.Context) ([]*DeviceStoreInfo, error) {
	query := `
		SELECT
		  dfm.device_id::text,
		  dfm.device_uid,
		  dfm.device_code,
		  dfm.device_type::text,
		  dfm.device_model,
		  dfm.mac_wifi,
		  dfm.imei,
		  dfm.comm_mode,
		  dfm.mcu_model,
		  drs.firmware_version,
		  COALESCE(host(network(set_masklen(d.device_ipv6, 48))), 'fd00:0:1::') AS tenant_id,
		  COALESCE(d.access, false) AS allow_access
		FROM device_factory_meta dfm
		LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id
		LEFT JOIN devices d ON d.device_id = dfm.device_id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query device_factory_meta: %w", err)
	}
	defer rows.Close()
	var results []*DeviceStoreInfo
	for rows.Next() {
		var d DeviceStoreInfo
		var deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
		var deviceCode sql.NullString
		if err := rows.Scan(&d.DeviceID, &d.DeviceUID, &deviceCode, &d.DeviceType, &deviceModel, &mac, &imei, &commMode, &mcuModel, &firmwareVersion, &d.TenantID, &d.AllowAccess); err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}
		d.DeviceCode = deviceCode
		d.DeviceModel = deviceModel
		d.MAC = mac
		d.IMEI = imei
		d.CommMode = commMode
		d.MCUModel = mcuModel
		d.FirmwareVersion = firmwareVersion
		results = append(results, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("device factory rows iteration error: %w", err)
	}
	return results, nil
}

// GetDevicesByTenant v2：按 tenant /48 prefix-match
//
// tenantID 期望 v2 格式：CIDR "fd00:0:T::/48" 或 host "fd00:0:T::"。若收到 v1 UUID 字符串
// 则视为非法（v1 cutover 后不允许），返回空结果而非报错以兼容 caller。
func (r *PostgresDeviceRepository) GetDevicesByTenant(ctx context.Context, tenantID string) ([]*domain.Device, error) {
	if !strings.Contains(tenantID, ":") {
		return nil, nil
	}
	prefix := tenantID
	if !strings.Contains(prefix, "/") {
		prefix = prefix + "/48"
	}
	query := `SELECT ` + deviceSelectV2 + `
		WHERE d.device_ipv6 <<= $1::INET
		ORDER BY COALESCE(dfm.device_code, dfm.device_uid)`
	rows, err := r.db.QueryContext(ctx, query, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()
	var devices []*domain.Device
	for rows.Next() {
		d, err := scanDeviceRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}
	return devices, nil
}

// CreateDevice v2：写 device_factory_meta + devices；自动注册未授权设备
//
// 用法：auth_service 在设备首次连上来且 dfm 无记录时调用，分配到 system tenant pool
// (fd00:0:1::/48)，access=FALSE 等 admin 在前端调拨。
//
// device_ipv6 派生：bytes 0-5=system /48 + bytes 6-11=0 (未分配) + bytes 12-15=device_uid 末 32 bit
func (r *PostgresDeviceRepository) CreateDevice(ctx context.Context, device *domain.Device) error {
	if device.DeviceUID == "" {
		return fmt.Errorf("device_uid is required")
	}
	if device.DeviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	deviceType := "Radar"
	if device.DeviceType.Valid && device.DeviceType.String != "" {
		deviceType = device.DeviceType.String
	}
	// INSERT device_factory_meta（出厂表）
	dfmQuery := `
		INSERT INTO device_factory_meta (device_id, device_uid, device_type)
		VALUES ($1::uuid, $2, $3::device_type_enum)
		ON CONFLICT (device_id) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, dfmQuery, device.DeviceID, device.DeviceUID, deviceType); err != nil {
		return fmt.Errorf("failed to insert device_factory_meta: %w", err)
	}
	// 派生 device_ipv6：system /48 + 0 bytes + MAC 末 32 bit
	deviceIPv6, err := deriveSystemDeviceIPv6(device.DeviceUID)
	if err != nil {
		return fmt.Errorf("failed to derive device_ipv6: %w", err)
	}
	// INSERT devices（业务绑定表）
	devQuery := `
		INSERT INTO devices (device_ipv6, device_id, access, monitoring_enabled)
		VALUES ($1::INET, $2::uuid, FALSE, FALSE)
		ON CONFLICT (device_ipv6) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, devQuery, deviceIPv6, device.DeviceID); err != nil {
		return fmt.Errorf("failed to insert devices: %w", err)
	}
	device.DeviceIPv6 = deviceIPv6
	device.TenantID = systemTenantHost
	device.AllowAccess = false
	device.BusinessAccess = "pending"
	device.Status = "offline"
	return nil
}

// systemTenantHost 等价 'fd00:0:1::' 的 host 字符串
const systemTenantHost = "fd00:0:1::"

// deriveSystemDeviceIPv6 给系统 tenant 池中的新设备派生 /128
//
// 格式：fd00:0:1:0:0:0:HHHH:HHHH，其中 HHHH:HHHH 是 device_uid 末 32 bit (8 hex chars)
// 例：UID="9923003AB17F" → suffix="003AB17F" → "fd00:0:1:0:0:0:003A:B17F/128"
func deriveSystemDeviceIPv6(deviceUID string) (string, error) {
	hex := strings.ReplaceAll(strings.TrimSpace(deviceUID), ":", "")
	if len(hex) < 8 {
		return "", fmt.Errorf("device_uid too short: %q", deviceUID)
	}
	suffix := hex[len(hex)-8:]
	return fmt.Sprintf("fd00:0:1:0:0:0:%s:%s/128", suffix[:4], suffix[4:]), nil
}

// UpdateDevice v2：仅支持 v2 业务可变字段（access / monitoring_enabled）
//
// device_uid/device_type 等出厂字段属 dfm，不可改；BoundRoomID/BoundBedID 属空间绑定，
// 需走 spatial API（device_ipv6 整段重置），不在此简单 UPDATE。
func (r *PostgresDeviceRepository) UpdateDevice(ctx context.Context, device *domain.Device) error {
	if device.DeviceUID == "" {
		return fmt.Errorf("device_uid is required")
	}
	updates := []string{}
	args := []interface{}{}
	argIdx := 1
	// access (legacy: business_access)
	if device.BusinessAccess != "" {
		approved := device.BusinessAccess == "approved" || device.BusinessAccess == "enable"
		updates = append(updates, fmt.Sprintf("access = $%d", argIdx))
		args = append(args, approved)
		argIdx++
	}
	// monitoring_enabled 默认总写（v1 行为：bool 字段每次 UPDATE 都覆盖）
	updates = append(updates, fmt.Sprintf("monitoring_enabled = $%d", argIdx))
	args = append(args, device.MonitoringEnabled)
	argIdx++
	updates = append(updates, "updated_at = NOW()")
	// WHERE device_uid → dfm → device_id
	args = append(args, device.DeviceUID)
	query := fmt.Sprintf(`
		UPDATE devices d
		SET %s
		FROM device_factory_meta dfm
		WHERE d.device_id = dfm.device_id AND dfm.device_uid = $%d`, strings.Join(updates, ", "), argIdx)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("device not found: %s", device.DeviceUID)
	}
	return nil
}

// SearchDevices v2：按 v2 列过滤；支持 tenant_id / status / device_type / search_keyword
//
// v1 criteria business_access 字符串改为 access bool；status 字符串改 drs.online 派生；
// metadata 字段 v2 不支持。
func (r *PostgresDeviceRepository) SearchDevices(ctx context.Context, criteria map[string]interface{}) ([]*domain.Device, error) {
	query := `SELECT ` + deviceSelectV2
	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1
	if tenantID, ok := criteria["tenant_id"].(string); ok && tenantID != "" && strings.Contains(tenantID, ":") {
		prefix := tenantID
		if !strings.Contains(prefix, "/") {
			prefix = prefix + "/48"
		}
		whereClauses = append(whereClauses, fmt.Sprintf("d.device_ipv6 <<= $%d::INET", argIdx))
		args = append(args, prefix)
		argIdx++
	}
	if status, ok := criteria["status"].(string); ok && status != "" {
		if status == "online" {
			whereClauses = append(whereClauses, "COALESCE(drs.online, false) = true")
		} else if status == "offline" {
			whereClauses = append(whereClauses, "COALESCE(drs.online, false) = false")
		}
	}
	if businessAccess, ok := criteria["business_access"].(string); ok && businessAccess != "" {
		approved := businessAccess == "approved" || businessAccess == "enable"
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(d.access, false) = $%d", argIdx))
		args = append(args, approved)
		argIdx++
	}
	if monitoringEnabled, ok := criteria["monitoring_enabled"].(bool); ok {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(d.monitoring_enabled, false) = $%d", argIdx))
		args = append(args, monitoringEnabled)
		argIdx++
	}
	if deviceType, ok := criteria["device_type"].(string); ok && deviceType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("dfm.device_type::text = $%d", argIdx))
		args = append(args, deviceType)
		argIdx++
	}
	if searchKeyword, ok := criteria["search_keyword"].(string); ok && searchKeyword != "" {
		searchType, _ := criteria["search_type"].(string)
		switch searchType {
		case "device_name":
			whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(dfm.device_code, dfm.device_uid) ILIKE $%d", argIdx))
		case "device_uid":
			whereClauses = append(whereClauses, fmt.Sprintf("dfm.device_uid ILIKE $%d", argIdx))
		default:
			whereClauses = append(whereClauses, fmt.Sprintf("(COALESCE(dfm.device_code, dfm.device_uid) ILIKE $%d OR dfm.device_uid ILIKE $%d)", argIdx, argIdx))
		}
		args = append(args, "%"+searchKeyword+"%")
		argIdx++
	}
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY COALESCE(dfm.device_code, dfm.device_uid)"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search devices: %w", err)
	}
	defer rows.Close()
	var devices []*domain.Device
	for rows.Next() {
		d, err := scanDeviceRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}
	return devices, nil
}

// CountDevicesByStatus v2：按 tenant /48 prefix 聚合 drs.online → "online"/"offline"
func (r *PostgresDeviceRepository) CountDevicesByStatus(ctx context.Context, tenantID string) (map[string]int, error) {
	if !strings.Contains(tenantID, ":") {
		return map[string]int{}, nil
	}
	prefix := tenantID
	if !strings.Contains(prefix, "/") {
		prefix = prefix + "/48"
	}
	query := `
		SELECT CASE WHEN COALESCE(drs.online, false) THEN 'online' ELSE 'offline' END AS status,
		       COUNT(*) AS count
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		LEFT JOIN device_runtime_state drs ON drs.device_id = d.device_id
		WHERE d.device_ipv6 <<= $1::INET
		GROUP BY 1`
	rows, err := r.db.QueryContext(ctx, query, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to count devices by status: %w", err)
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count row: %w", err)
		}
		counts[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating count rows: %w", err)
	}
	return counts, nil
}

// GetDeviceStoreInfo v2：auth 热路径，按 device_uid 查。
//
// device_ipv6 单程票后 SELECT 加上 host(d.device_ipv6) 派生 DeviceAddr，
// publisher 端无须再二次反查 cards/dfm。
func (r *PostgresDeviceRepository) GetDeviceStoreInfo(ctx context.Context, deviceUID string) (*DeviceStoreInfo, error) {
	query := `
		SELECT
		  dfm.device_id::text,
		  dfm.device_uid,
		  dfm.device_code,
		  dfm.device_type::text,
		  dfm.device_model,
		  dfm.mac_wifi,
		  dfm.imei,
		  dfm.comm_mode,
		  dfm.mcu_model,
		  drs.firmware_version,
		  COALESCE(host(network(set_masklen(d.device_ipv6, 48))) || '/48', 'fd00:0:1::/48') AS tenant_pref,
		  COALESCE(d.access, false) AS allow_access,
		  host(d.device_ipv6)::text AS device_addr
		FROM device_factory_meta dfm
		LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id
		LEFT JOIN devices d ON d.device_id = dfm.device_id
		WHERE dfm.device_uid = $1
		LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, deviceUID)
	var ds DeviceStoreInfo
	var deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
	var deviceCode, addrStr sql.NullString
	if err := row.Scan(&ds.DeviceID, &ds.DeviceUID, &deviceCode, &ds.DeviceType, &deviceModel, &mac, &imei, &commMode, &mcuModel, &firmwareVersion, &ds.TenantID, &ds.AllowAccess, &addrStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found in device_store")
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}
	ds.DeviceCode = deviceCode
	ds.DeviceModel = deviceModel
	ds.MAC = mac
	ds.IMEI = imei
	ds.CommMode = commMode
	ds.MCUModel = mcuModel
	ds.FirmwareVersion = firmwareVersion
	if addrStr.Valid && addrStr.String != "" {
		if a, perr := netip.ParseAddr(addrStr.String); perr == nil {
			ds.DeviceAddr = a
		}
	}
	return &ds, nil
}

// GetDeviceStoreByDeviceID v2：按 dfm.device_id 查（device_store 变化信号回调用）
func (r *PostgresDeviceRepository) GetDeviceStoreByDeviceID(ctx context.Context, deviceID string) (*DeviceStoreInfo, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	query := `
		SELECT
		  dfm.device_id::text,
		  dfm.device_uid,
		  dfm.device_code,
		  dfm.device_type::text,
		  dfm.device_model,
		  dfm.mac_wifi,
		  dfm.imei,
		  dfm.comm_mode,
		  dfm.mcu_model,
		  drs.firmware_version,
		  COALESCE(host(network(set_masklen(d.device_ipv6, 48))) || '/48', 'fd00:0:1::/48') AS tenant_pref,
		  COALESCE(d.access, false) AS allow_access,
		  host(d.device_ipv6)::text AS device_addr
		FROM device_factory_meta dfm
		LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id
		LEFT JOIN devices d ON d.device_id = dfm.device_id
		WHERE dfm.device_id = $1::uuid
		LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, deviceID)
	var ds DeviceStoreInfo
	var deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
	var deviceCode, addrStr sql.NullString
	if err := row.Scan(&ds.DeviceID, &ds.DeviceUID, &deviceCode, &ds.DeviceType, &deviceModel, &mac, &imei, &commMode, &mcuModel, &firmwareVersion, &ds.TenantID, &ds.AllowAccess, &addrStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found in device_store")
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}
	ds.DeviceCode = deviceCode
	ds.DeviceModel = deviceModel
	ds.MAC = mac
	ds.IMEI = imei
	ds.CommMode = commMode
	ds.MCUModel = mcuModel
	ds.FirmwareVersion = firmwareVersion
	if addrStr.Valid && addrStr.String != "" {
		if a, perr := netip.ParseAddr(addrStr.String); perr == nil {
			ds.DeviceAddr = a
		}
	}
	return &ds, nil
}

// GetAlarmEnablement v2：spatial_config longest-prefix-match resolution 替代 alarm_device.monitor_config
//
// 当前为占位实现，返回空 enablement（caller 已知接受空集合 = 默认全部使能/默认禁用，
// 视 alarm 流处理路径而定）。后续 Phase 2 把此函数改为 spatial_config 查询 +
// resolve_config() PG function 调用。
func (r *PostgresDeviceRepository) GetAlarmEnablement(ctx context.Context, tenantID, deviceUID string) ([]alarm.AlarmEnablementItem, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceUID == "" {
		return nil, fmt.Errorf("device_uid is required")
	}
	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceUID)
	if cached, ok := r.enablementCache.Load(cacheKey); ok {
		if enablementMap, ok := cached.([]alarm.AlarmEnablementItem); ok {
			return enablementMap, nil
		}
	}
	emptyEnablement := []alarm.AlarmEnablementItem{}
	r.enablementCache.Store(cacheKey, emptyEnablement)
	return emptyEnablement, nil
}

// ClearAlarmEnablementCache 清除缓存
func (r *PostgresDeviceRepository) ClearAlarmEnablementCache(tenantID, deviceUID string) {
	if tenantID == "" || deviceUID == "" {
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", tenantID, deviceUID)
	r.enablementCache.Delete(cacheKey)
}

// PreloadAlarmEnablement 预加载
func (r *PostgresDeviceRepository) PreloadAlarmEnablement(ctx context.Context, tenantID, deviceUID string) error {
	_, err := r.GetAlarmEnablement(ctx, tenantID, deviceUID)
	return err
}

// GetAllAccessibleDevices 启动时主动订阅可访问的 Radar 设备 UID 列表
//
// v2 条件：dfm.device_type='Radar' AND devices.access=TRUE
//
// 必须按 device_type 过滤，否则 Sleepad 等会被订阅，污染 publishOnlineForConnectedAfterStartup
// 写入的 device_type 标签（详 v1 BM87225200672 事件分析）。
func (r *PostgresDeviceRepository) GetAllAccessibleDevices(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT dfm.device_uid
		FROM device_factory_meta dfm
		JOIN devices d ON d.device_id = dfm.device_id
		WHERE dfm.device_type = 'Radar'
		  AND d.access = TRUE
		ORDER BY dfm.device_uid`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query accessible devices: %w", err)
	}
	defer rows.Close()
	var deviceUIDs []string
	for rows.Next() {
		var deviceUID string
		if err := rows.Scan(&deviceUID); err != nil {
			return nil, fmt.Errorf("failed to scan device_uid: %w", err)
		}
		deviceUIDs = append(deviceUIDs, deviceUID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}
	return deviceUIDs, nil
}

// nullableString 把 sql.NullString 转 interface{}（nil-safe，便于 JSON 序列化）
func nullableString(n sql.NullString) interface{} {
	if n.Valid {
		return n.String
	}
	return nil
}

// 防止未使用导入告警；后续接 spatial_config 时移除
var _ = json.Marshal
