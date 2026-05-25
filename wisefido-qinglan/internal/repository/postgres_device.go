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
// v2.5 数据源（详 owlRD/dbv2/21_device_factory_meta.sql / 23_devices.sql；drs 已退役）：
//   - device_factory_meta (dfm)：出厂元数据 + firmware_version；device_uid/device_type/device_model/imei/comm_mode/mcu_model
//   - devices (d)：业务绑定（PK = device_ipv6 INET /128） access/monitoring_enabled
//
// 1:1 关系（device_id UUID 双表 join）；online 走 Redis device:status:{ipv6}（cardagg 维护）。
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

// deviceSelectV2 device_uid identity + device_addr 业务寻址。
const deviceSelectV2 = `
		dfm.device_uid                                                    AS device_uid,
		COALESCE(host(d.device_addr), '')                                 AS device_addr,
		COALESCE(host(network(set_masklen(d.device_addr, 48))), '')       AS tenant_id,
		COALESCE(dfm.device_code, dfm.device_uid)                         AS device_name,
		CASE WHEN d.device_addr IS NOT NULL
		     THEN host(network(set_masklen(d.device_addr, 88))) || '/88'
		     ELSE NULL END                                                AS bound_room_id,
		CASE WHEN d.device_addr IS NOT NULL
		     THEN host(network(set_masklen(d.device_addr, 96))) || '/96'
		     ELSE NULL END                                                AS bound_bed_id,
		'offline'::text                                                   AS status,
		COALESCE(d.access, false)                                          AS access,
		COALESCE(d.monitoring_enabled, false)                              AS monitoring_enabled,
		dfm.device_type::text                                              AS device_type,
		dfm.device_model                                                   AS device_model,
		dfm.imei                                                           AS imei,
		dfm.comm_mode                                                      AS comm_mode,
		dfm.mcu_model                                                      AS mcu_model,
		dfm.firmware_version                                               AS firmware_version
	FROM device_factory_meta dfm
	LEFT JOIN devices d ON d.device_uid = dfm.device_uid`

// scanDeviceRow 与 deviceSelectV2 列序对齐的 scan helper
func scanDeviceRow(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Device, error) {
	var d domain.Device
	var deviceAddr, status string
	var boundRoomID, boundBedID sql.NullString
	var deviceType sql.NullString
	var deviceModel, imei, commMode, mcuModel, firmwareVersion sql.NullString
	err := row.Scan(
		&d.DeviceUID,
		&deviceAddr,
		&d.TenantID,
		&d.DeviceName,
		&boundRoomID,
		&boundBedID,
		&status,
		&d.Access,
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
	d.DeviceAddr = deviceAddr
	d.BoundRoomID = boundRoomID
	d.BoundBedID = boundBedID
	d.Status = status
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

// UpdateDeviceMonitoring 直接按 device_uid 关联 devices。
func (r *PostgresDeviceRepository) UpdateDeviceMonitoring(ctx context.Context, uid string, enabled bool) error {
	query := `
		UPDATE devices
		SET monitoring_enabled = $1, updated_at = NOW()
		WHERE device_uid = $2`
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

// GetDeviceProperties v2.5：firmware_version 已迁到 dfm；返回 dfm 列即可。
//
// 兼容 v1 调用：返回已知键；未知键返回 nil。
func (r *PostgresDeviceRepository) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	query := `
		SELECT dfm.firmware_version, dfm.mcu_model, dfm.device_model, dfm.imei, dfm.comm_mode
		FROM device_factory_meta dfm
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

// SetDeviceProperties v2.5：firmware_version 写 device_factory_meta（drs 已退役）。
//
// v1 把任意 JSON 写 devices.metadata；v2 不再有该列。已知键写 dfm，未知键忽略并 warn。
func (r *PostgresDeviceRepository) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
	if len(properties) == 0 {
		return nil
	}
	var fw sql.NullString
	for k, v := range properties {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if k == "firmware_version" {
			fw = sql.NullString{String: s, Valid: true}
		}
	}
	if !fw.Valid {
		return nil
	}
	result, err := r.db.ExecContext(ctx,
		`UPDATE device_factory_meta SET firmware_version = $2 WHERE device_uid = $1`,
		uid, fw.String)
	if err != nil {
		return fmt.Errorf("failed to update dfm firmware_version: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("device not found: %s", uid)
	}
	return nil
}

// GetAllDeviceStoreInfo dfm + devices 派生，启动时拉全集
func (r *PostgresDeviceRepository) GetAllDeviceStoreInfo(ctx context.Context) ([]*DeviceStoreInfo, error) {
	query := `
		SELECT
		  dfm.device_uid,
		  dfm.device_code,
		  dfm.device_type::text,
		  dfm.device_model,
		  dfm.mac_wifi,
		  dfm.imei,
		  dfm.comm_mode,
		  dfm.mcu_model,
		  dfm.firmware_version,
		  COALESCE(host(network(set_masklen(d.device_addr, 48))), 'fd00:0:1::') AS tenant_id,
		  COALESCE(d.access, false) AS access,
		  COALESCE(host(d.device_addr), '') AS device_addr
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_uid = dfm.device_uid`
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
		var addrStr sql.NullString
		if err := rows.Scan(&d.DeviceUID, &deviceCode, &d.DeviceType, &deviceModel, &mac, &imei, &commMode, &mcuModel, &firmwareVersion, &d.TenantID, &d.Access, &addrStr); err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}
		d.DeviceCode = deviceCode
		d.DeviceModel = deviceModel
		d.MAC = mac
		d.IMEI = imei
		d.CommMode = commMode
		d.MCUModel = mcuModel
		d.FirmwareVersion = firmwareVersion
		if addrStr.Valid && addrStr.String != "" {
			if a, perr := netip.ParseAddr(addrStr.String); perr == nil {
				d.DeviceAddr = a
			}
		}
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
		WHERE d.device_addr <<= $1::INET
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

// CreateDevice 写 device_factory_meta + devices；自动注册未授权设备
//
// 业务约定：未知 UID 从 MQTT 来认证 → 落 Trash pool (fd00:0:2::/48)，access=FALSE，
// 待 platform_admin 决定接受（迁 System / 真 tenant）或保持丢弃。
//
// 区分于 wisefido-data 导入路径——后者是已知设备 (CSV/Excel/UI)，落 System。
//
// device_addr 派生：bytes 0-5=trash /48 + bytes 6-11=0 (未分配) + bytes 12-15=device_uid 末 32 bit
func (r *PostgresDeviceRepository) CreateDevice(ctx context.Context, device *domain.Device) error {
	if device.DeviceUID == "" {
		return fmt.Errorf("device_uid is required")
	}
	deviceType := "Radar"
	if device.DeviceType.Valid && device.DeviceType.String != "" {
		deviceType = device.DeviceType.String
	}
	// INSERT device_factory_meta（出厂表）
	dfmQuery := `
		INSERT INTO device_factory_meta (device_uid, device_type)
		VALUES ($1, $2::device_type_enum)
		ON CONFLICT (device_uid) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, dfmQuery, device.DeviceUID, deviceType); err != nil {
		return fmt.Errorf("failed to insert device_factory_meta: %w", err)
	}
	// 派生 device_addr：trash /48 + 0 bytes + UID 末 32 bit
	deviceAddr, err := deriveTrashDeviceAddr(device.DeviceUID)
	if err != nil {
		return fmt.Errorf("failed to derive device_addr: %w", err)
	}
	// INSERT devices（业务绑定表）
	devQuery := `
		INSERT INTO devices (device_addr, device_uid, access, monitoring_enabled)
		VALUES ($1::INET, $2, FALSE, FALSE)
		ON CONFLICT (device_addr) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, devQuery, deviceAddr, device.DeviceUID); err != nil {
		return fmt.Errorf("failed to insert devices: %w", err)
	}
	device.DeviceAddr = deviceAddr
	device.TenantID = trashTenantHost
	device.Access = false
	device.Status = "offline"
	return nil
}

// trashTenantHost 等价 'fd00:0:2::' 的 host 字符串
const trashTenantHost = "fd00:0:2::"

// deriveTrashDeviceAddr 给 Trash tenant 池中的新设备派生 /128。
// UID 末 32 bit 派生规则 doc/datagram_envelope.md §2.2：strip 非 hex + 末 8 hex + 不足左补 0。
func deriveTrashDeviceAddr(deviceUID string) (string, error) {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
			return r
		default:
			return -1
		}
	}, deviceUID)
	if len(clean) < 8 {
		clean = strings.Repeat("0", 8-len(clean)) + clean
	}
	last8 := strings.ToLower(clean[len(clean)-8:])
	return fmt.Sprintf("fd00:0:2:0:0:0:%s:%s/128", last8[:4], last8[4:]), nil
}

// UpdateDevice v2：仅支持业务可变字段（access / monitoring_enabled）
//
// device_uid/device_type 等出厂字段属 dfm，不可改；BoundRoomID/BoundBedID 属空间绑定，
// 需走 spatial API（device_ipv6 整段重置），不在此简单 UPDATE。
//
// access / monitoring_enabled 均为 bool 字段，每次 UPDATE 总是覆盖（不做 partial update）。
func (r *PostgresDeviceRepository) UpdateDevice(ctx context.Context, device *domain.Device) error {
	if device.DeviceUID == "" {
		return fmt.Errorf("device_uid is required")
	}
	updates := []string{}
	args := []interface{}{}
	argIdx := 1
	// access：platform_admin 审批位
	updates = append(updates, fmt.Sprintf("access = $%d", argIdx))
	args = append(args, device.Access)
	argIdx++
	updates = append(updates, fmt.Sprintf("monitoring_enabled = $%d", argIdx))
	args = append(args, device.MonitoringEnabled)
	argIdx++
	updates = append(updates, "updated_at = NOW()")
	// WHERE device_uid → dfm → device_addr
	args = append(args, device.DeviceUID)
	query := fmt.Sprintf(`
		UPDATE devices
		SET %s
		WHERE device_uid = $%d`, strings.Join(updates, ", "), argIdx)
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

// SearchDevices 按 criteria 过滤；支持 tenant_id / access / monitoring_enabled / device_type / search_keyword。
// status 不在 SQL 层（drs.online 已删；status 在 Redis），caller 想按 status 过滤自己读 Redis 后处理。
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
		whereClauses = append(whereClauses, fmt.Sprintf("d.device_addr <<= $%d::INET", argIdx))
		args = append(args, prefix)
		argIdx++
	}
	_ = criteria["status"]
	if access, ok := criteria["access"].(bool); ok {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(d.access, false) = $%d", argIdx))
		args = append(args, access)
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

// CountDevicesByStatus v2.5：按 tenant /48 prefix 聚合 — online status 在 Redis 不在 SQL；
// 此 SQL 路径退化为返回 {"offline": <total>}；caller 需 Redis 统计真 online 数请自行实现。
func (r *PostgresDeviceRepository) CountDevicesByStatus(ctx context.Context, tenantID string) (map[string]int, error) {
	if !strings.Contains(tenantID, ":") {
		return map[string]int{}, nil
	}
	prefix := tenantID
	if !strings.Contains(prefix, "/") {
		prefix = prefix + "/48"
	}
	query := `
		SELECT 'offline'::text AS status, COUNT(*) AS count
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE d.device_addr <<= $1::INET
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

// GetDeviceStoreInfo Phase 2：auth 热路径，按 device_uid 查；返回派生 device_addr。
func (r *PostgresDeviceRepository) GetDeviceStoreInfo(ctx context.Context, deviceUID string) (*DeviceStoreInfo, error) {
	return r.queryDeviceStoreInfo(ctx, "WHERE dfm.device_uid = $1", deviceUID)
}

// GetDeviceStoreByDeviceAddr device_id UUID 退役，业务键改 device_addr。
func (r *PostgresDeviceRepository) GetDeviceStoreByDeviceAddr(ctx context.Context, deviceAddr string) (*DeviceStoreInfo, error) {
	if deviceAddr == "" {
		return nil, fmt.Errorf("device_addr is required")
	}
	return r.queryDeviceStoreInfo(ctx, "WHERE d.device_addr = $1::INET", deviceAddr)
}

func (r *PostgresDeviceRepository) queryDeviceStoreInfo(ctx context.Context, whereClause, key string) (*DeviceStoreInfo, error) {
	query := `
		SELECT
		  dfm.device_uid,
		  dfm.device_code,
		  dfm.device_type::text,
		  dfm.device_model,
		  dfm.mac_wifi,
		  dfm.imei,
		  dfm.comm_mode,
		  dfm.mcu_model,
		  dfm.firmware_version,
		  COALESCE(host(network(set_masklen(d.device_addr, 48))) || '/48', 'fd00:0:1::/48') AS tenant_pref,
		  COALESCE(d.access, false) AS access,
		  COALESCE(host(d.device_addr), '') AS device_addr
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_uid = dfm.device_uid
		` + whereClause + `
		LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, key)
	var ds DeviceStoreInfo
	var deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
	var deviceCode, addrStr sql.NullString
	if err := row.Scan(&ds.DeviceUID, &deviceCode, &ds.DeviceType, &deviceModel, &mac, &imei, &commMode, &mcuModel, &firmwareVersion, &ds.TenantID, &ds.Access, &addrStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
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
		JOIN devices d ON d.device_uid = dfm.device_uid
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
