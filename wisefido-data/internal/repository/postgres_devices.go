package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgresDevicesRepository 设备Repository实现（强类型）
// 遵循"bottom-up"设计原则，替代已删除的数据库触发器
type PostgresDevicesRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresDevicesRepository 创建设备Repository
func NewPostgresDevicesRepository(db *sql.DB) *PostgresDevicesRepository {
	return &PostgresDevicesRepository{db: db}
}

// SetLogger 设置日志记录器（可选，用于记录设备连接事件）
func (r *PostgresDevicesRepository) SetLogger(logger *zap.Logger) {
	r.logger = logger
}

// orderByClause 白名单排序字段，防止 SQL 注入；direction 仅允许 asc/desc
func orderByClause(sort, direction string) string {
	dir := "ASC"
	if strings.ToUpper(strings.TrimSpace(direction)) == "DESC" {
		dir = "DESC"
	}
	col := strings.TrimSpace(strings.ToLower(sort))
	switch col {
	case "device_uid":
		return "d.device_uid " + dir
	case "device_type":
		return "ds.device_type " + dir
	case "device_model":
		return "ds.device_model " + dir
	case "status":
		return "d.status " + dir
	case "business_access":
		return "d.business_access " + dir
	case "device_id":
		return "d.device_id " + dir
	case "device_code":
		return "ds.device_code " + dir
	case "device_name":
	default:
	}
	return "d.device_name " + dir
}

// orderByClauseDevicesV2 v2 表名映射的排序白名单（dfm/drs/d/o/t）。
// 默认 import_date DESC（保持 v1 行为：新导入的库存先排在前面）。
func orderByClauseDevicesV2(sortKey, direction string) string {
	dir := "ASC"
	if strings.TrimSpace(strings.ToUpper(direction)) == "DESC" {
		dir = "DESC"
	}
	col := strings.TrimSpace(strings.ToLower(sortKey))
	switch col {
	case "device_uid", "device_code", "device_type", "device_model", "imei", "comm_mode", "mcu_model":
		return "dfm." + col + " " + dir
	case "mac":
		return "dfm.mac_wifi " + dir
	case "firmware_version":
		return "drs.firmware_version " + dir
	case "import_date":
		return "dfm.import_date " + dir
	case "allocate_time":
		return "d.created_at " + dir
	case "tenant_name":
		return "t.tenant_name " + dir
	case "device_id":
		return "dfm.device_id " + dir
	case "ota_target_firmware_version":
		return "o.target_firmware_version " + dir
	case "ota_target_mcu_model":
		return "o.target_mcu_model " + dir
	case "ota_status":
		return "o.status " + dir
	case "ota_progress":
		return "o.progress " + dir
	case "online_status":
		return "drs.online " + dir
	case "device_name":
		// v2 没 device_name 列；让 device_uid 充当 device_name 的排序代理
		return "dfm.device_uid " + dir
	case "status":
		// v1 的 status 是 Enabled/Disabled/Error；v2 用 devices 行存在性近似
		return "(CASE WHEN d.spatial_addr IS NOT NULL THEN 0 ELSE 1 END) " + dir + ", dfm.device_uid"
	default:
		return "dfm.import_date DESC, dfm.device_uid"
	}
}

// ListDevices 查询设备列表。
//
// 业务规则（v1 R-001 / BR-001 保留）：
//   - tenantID == ""：不加 tenant 过滤（platform_admin 视图，跨 tenant 可见）
//   - tenantID != ""：WHERE devices.spatial_addr <<= $tenantID::INET
//     调用方（DeviceHandler）负责决定是否传 ""，本层不做 admin 检测。
//
// v2 schema JOIN（已纯位掩码化，详见 owlBack/doc/spatial_query_patterns.md）：
//
//	device_factory_meta dfm  (PK device_id, 出厂元数据：device_uid/device_code/device_type/mac_wifi/imei/...)
//	    LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id  (online/firmware_version/...)
//	    LEFT JOIN devices d              ON d.device_id = dfm.device_id      (spatial_addr → tenant/room/bed 反推)
//	    LEFT JOIN device_ota o           ON o.spatial_addr = d.spatial_addr  (OTA 计划)
//
// rooms/beds 不再 JOIN：legacy slot=0 已迁移，byte 10/11 != 0 唯一含义 = 绑定到该层。
//
// v1 ⇄ v2 字段映射：
//   - tenant_id        ← host(network(set_masklen(d.spatial_addr,48))) || '/48'
//   - bound_room_id    ← /88 prefix when byte 10 != 0
//   - bound_bed_id     ← /96 prefix when byte 11 != 0
//   - status           ← devices 行存在 → 'Enabled'，否则 'Disabled'（FE 软删用）
//   - business_access  ← v2 已删；统一回 'approved' 让 FE filter 不过滤掉
//   - device_name      ← v2 无；project as device_uid
//   - OTA approve_way  ← 反映射 v1 的 (ota_permit, ota_way, ota_tenant_approved)
func (r *PostgresDevicesRepository) ListDevices(ctx context.Context, tenantID string, filters DeviceFilters, page, size int, sort, direction string) ([]*domain.Device, int, error) {
	where := []string{}
	args := []any{}
	argN := 1

	// tenant scope：caller 决定（admin 传 "" 不过滤；其他身份传 /48 prefix 限本 tenant）
	if t := strings.TrimSpace(tenantID); t != "" {
		if !looksLikeINETPrefix(t) {
			// 防御：非法 INET prefix 直接返回空（避免 SQL 报错把 500 抛给 FE）
			return []*domain.Device{}, 0, nil
		}
		where = append(where, fmt.Sprintf("d.spatial_addr <<= $%d::INET", argN))
		args = append(args, t)
		argN++
	}

	if dt := strings.TrimSpace(filters.DeviceType); dt != "" {
		where = append(where, fmt.Sprintf("dfm.device_type::text = $%d", argN))
		args = append(args, dt)
		argN++
	}

	// 搜索关键词：device_uid / device_code / mac_wifi / imei / device_id::text 模糊匹配
	if kw := strings.TrimSpace(filters.SearchKeyword); kw != "" {
		where = append(where, fmt.Sprintf(
			"(dfm.device_uid ILIKE $%d OR dfm.device_code ILIKE $%d OR dfm.mac_wifi ILIKE $%d OR dfm.imei ILIKE $%d OR dfm.device_id::text ILIKE $%d)",
			argN, argN, argN, argN, argN))
		args = append(args, "%"+kw+"%")
		argN++
	}

	// online/offline 状态过滤：drs.online IS TRUE / FALSE
	if len(filters.Status) > 0 {
		ors := []string{}
		for _, st := range filters.Status {
			s := strings.ToLower(strings.TrimSpace(st))
			switch s {
			case "online":
				ors = append(ors, "drs.online IS TRUE")
			case "offline":
				ors = append(ors, "(drs.online IS FALSE OR drs.online IS NULL)")
			}
		}
		if len(ors) > 0 {
			where = append(where, "("+strings.Join(ors, " OR ")+")")
		}
	}

	// business_access：v2 没有 allow_access 列；统一视为 approved，FE filter 'approved' 不丢数据。
	// filters.BusinessAccess 在 v2 直接忽略（详见函数注释）。

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// COUNT(*) — 分页前的总数，必须 JOIN devices d 因为 tenant 过滤靠 d.spatial_addr
	countQ := `
		SELECT COUNT(*)
		  FROM device_factory_meta dfm
		  LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id
		  LEFT JOIN devices d                ON d.device_id = dfm.device_id
		  LEFT JOIN device_ota o             ON o.spatial_addr = d.spatial_addr
		  ` + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count devices: %w", err)
	}
	if total == 0 {
		return []*domain.Device{}, 0, nil
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	dataArgs := append([]any{}, args...)
	dataArgs = append(dataArgs, size, offset)
	limitN := argN
	offsetN := argN + 1

	q := `
		SELECT
		  dfm.device_id::text                                                                  AS device_id,
		  dfm.device_uid                                                                       AS device_uid,
		  dfm.device_code                                                                      AS device_code,
		  dfm.device_type::text                                                                AS device_type,
		  dfm.device_model                                                                     AS device_model,
		  dfm.mac_wifi                                                                         AS mac,
		  dfm.imei                                                                             AS imei,
		  dfm.comm_mode                                                                        AS comm_mode,
		  dfm.mcu_model                                                                        AS mcu_model,
		  drs.firmware_version                                                                 AS firmware_version,
		  -- tenant /48 反推（无 devices 行时为 NULL，由扫描层兜底为 unbound prefix）
		  CASE WHEN d.spatial_addr IS NOT NULL
		       THEN host(network(set_masklen(d.spatial_addr, 48))) || '/48'
		  END                                                                                  AS tenant_id,
		  -- bound_room_id / bound_bed_id：纯位掩码 + **最长掩码优先**互斥
		  --   byte 11 != 0           ⇒ bound to bed (bound_bed_id)，bound_room_id = NULL
		  --   byte 10 != 0 且 11 = 0 ⇒ bound to room (bound_room_id)
		  --   byte 10 = 0            ⇒ unbound (兩者皆 NULL)
		  -- 这样 FE tree 只会在 device 真正归属的最深层节点显示一次
		  CASE WHEN d.spatial_addr IS NOT NULL
		        AND (d.spatial_addr & '::ff00:0:0'::INET) <> '::'::INET
		        AND (d.spatial_addr & '::ff:0:0'::INET)   = '::'::INET
		       THEN host(network(set_masklen(d.spatial_addr, 88))) || '/88'
		  END                                                                                  AS bound_room_id,
		  CASE WHEN d.spatial_addr IS NOT NULL
		        AND (d.spatial_addr & '::ff:0:0'::INET) <> '::'::INET
		       THEN host(network(set_masklen(d.spatial_addr, 96))) || '/96'
		  END                                                                                  AS bound_bed_id,
		  -- status 软删：devices 行存在 → Enabled，否则 Disabled
		  CASE WHEN d.spatial_addr IS NOT NULL THEN 'Enabled' ELSE 'Disabled' END              AS status,
		  COALESCE(d.monitoring_enabled, FALSE)                                                AS monitoring_enabled,
		  -- OTA 字段（v2 device_ota → v1 形）
		  o.target_firmware_version                                                            AS ota_target_firmware_version,
		  o.target_mcu_model                                                                   AS ota_target_mcu_model,
		  CASE WHEN o.approve_way IS NULL THEN 'false' ELSE 'true' END                         AS ota_permit,
		  CASE
		    WHEN o.approve_way IS NULL                THEN NULL
		    WHEN o.approve_way LIKE '%_schedule'      THEN 'schedule'
		    WHEN o.approve_way LIKE 'tenant_%'        THEN 'tenant'
		    WHEN o.approve_way LIKE '%_manual'        THEN 'manual'
		    ELSE NULL
		  END                                                                                  AS ota_way,
		  to_char(o.schedule, 'YYYY-MM-DD HH24:MI:SS')                                         AS ota_schedule,
		  o.status                                                                             AS ota_status,
		  o.progress                                                                           AS ota_progress,
		  CASE WHEN o.approve_way LIKE 'tenant_%' THEN TRUE ELSE FALSE END                     AS ota_tenant_approved
		FROM device_factory_meta dfm
		LEFT JOIN device_runtime_state drs ON drs.device_id = dfm.device_id
		LEFT JOIN devices d                ON d.device_id = dfm.device_id
		LEFT JOIN device_ota o             ON o.spatial_addr = d.spatial_addr
		` + whereClause + `
		ORDER BY ` + orderByClauseDevicesV2(sort, direction) + `
		LIMIT $` + fmt.Sprintf("%d", limitN) + ` OFFSET $` + fmt.Sprintf("%d", offsetN)

	rows, err := r.db.QueryContext(ctx, q, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	out := []*domain.Device{}
	for rows.Next() {
		var d domain.Device
		var (
			deviceCode, deviceType, deviceModel, mac, imei, commMode, mcuModel, firmwareVersion sql.NullString
			tenantID, boundRoomID, boundBedID                                                   sql.NullString
			status                                                                              string
			monitoringEnabled                                                                   bool
			otaTargetFW, otaTargetMCU, otaPermit, otaWay, otaSchedule, otaStatus                sql.NullString
			otaProgress                                                                         sql.NullInt32
			otaTenantApproved                                                                   sql.NullBool
		)
		if err := rows.Scan(
			&d.DeviceID,
			&d.DeviceUID,
			&deviceCode,
			&deviceType,
			&deviceModel,
			&mac,
			&imei,
			&commMode,
			&mcuModel,
			&firmwareVersion,
			&tenantID,
			&boundRoomID,
			&boundBedID,
			&status,
			&monitoringEnabled,
			&otaTargetFW,
			&otaTargetMCU,
			&otaPermit,
			&otaWay,
			&otaSchedule,
			&otaStatus,
			&otaProgress,
			&otaTenantApproved,
		); err != nil {
			return nil, 0, fmt.Errorf("scan device row: %w", err)
		}

		d.DeviceCode = deviceCode
		d.DeviceType = deviceType
		d.DeviceModel = deviceModel
		d.MAC = mac
		d.IMEI = imei
		d.CommMode = commMode
		d.MCUModel = mcuModel
		d.FirmwareVersion = firmwareVersion

		if tenantID.Valid {
			d.TenantID = tenantID.String
		} else {
			d.TenantID = defaultUnboundTenantPrefix
		}
		d.BoundRoomID = boundRoomID
		d.BoundBedID = boundBedID
		// room_id / unit_id 计算字段 — bound_* 即可代表当前绑定层
		d.RoomID = boundRoomID
		// unit_id 暂不展开（需要再 JOIN units 表；FE 列表场景一般不需要 unit_id，详情页另查）
		d.Status = status
		d.MonitoringEnabled = monitoringEnabled
		// v2 没 allow_access 列；统一 'approved' 让 FE filter 不过滤掉
		d.BusinessAccess = "approved"
		// v2 没 device_name 列；用 device_uid 占位（FE 可自行渲染）
		d.DeviceName = d.DeviceUID

		d.OTATargetFW = otaTargetFW
		d.OTATargetMCU = otaTargetMCU
		d.OTAPermit = otaPermit
		d.OTAWay = otaWay
		d.OTASchedule = otaSchedule
		d.OTAStatus = otaStatus
		d.OTAProgress = otaProgress
		if otaTenantApproved.Valid {
			d.OTATenantApproved = otaTenantApproved.Bool
		}

		out = append(out, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate device rows: %w", err)
	}
	return out, total, nil
}

// GetDevice 查询单个设备
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) GetDevice(ctx context.Context, tenantID, deviceID string) (*domain.Device, error) {
	return nil, fmt.Errorf("device not found: tenant_id=%s, device_id=%s", tenantID, deviceID)
}

// GetDeviceByUID 根据 device_uid 和 tenant_id 获取设备信息
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) GetDeviceByUID(ctx context.Context, tenantID, deviceUID string) (*domain.Device, error) {
	return nil, fmt.Errorf("device not found: tenant_id=%s, device_uid=%s", tenantID, deviceUID)
}

// GetDevicesBoundToRoom 查询绑定到指定 room 的设备（仅 id/name，用于删除前检查）
// v2: roomID 是 /88 prefix；device.spatial_addr <<= room.spatial_prefix 即视为绑定到该 room（含 bed 下层）。
// devices 表无 device_name 列；用 device_id 字符串占位回填 d.DeviceName 让上层错误信息能显示。
func (r *PostgresDevicesRepository) GetDevicesBoundToRoom(ctx context.Context, tenantID, roomID string) ([]*domain.Device, error) {
	if tenantID == "" || roomID == "" || !looksLikeINETPrefix(roomID) {
		return nil, nil
	}
	q := `SELECT device_id::text, host(spatial_addr) FROM devices WHERE spatial_addr <<= $1::INET`
	rows, err := r.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Device
	for rows.Next() {
		d := &domain.Device{}
		if err := rows.Scan(&d.DeviceID, &d.DeviceName); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetRoomBoundDeviceTypeLetters 返回绑定到 room 的设备类型字母（R=Radar, S=Sleepad），供前端 RoomName(R) 展示
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) GetRoomBoundDeviceTypeLetters(ctx context.Context, tenantID, roomID string) ([]string, error) {
	return nil, nil
}

// GetDevicesBoundToBedsWithDetails 返回多个 bed 上绑定的设备类型字母及 monitor 状态
func (r *PostgresDevicesRepository) GetDevicesBoundToBedsWithDetails(ctx context.Context, tenantID string, bedIDs []string) (map[string][]DeviceTypeDetail, error) {
	// v2 stub：device_store 表已删；返回空让上层不爆 SQL。Phase E 重写时改用 device_runtime_state。
	return make(map[string][]DeviceTypeDetail), nil
}

// GetDevicesBoundToBed 查询绑定到指定 bed 的设备（仅 id/name，用于删除前检查）
// v2: bedID 是 /96 prefix；device.spatial_addr <<= bed.spatial_prefix 视为绑定到该 bed。
func (r *PostgresDevicesRepository) GetDevicesBoundToBed(ctx context.Context, tenantID, bedID string) ([]*domain.Device, error) {
	if tenantID == "" || bedID == "" || !looksLikeINETPrefix(bedID) {
		return nil, nil
	}
	q := `SELECT device_id::text, host(spatial_addr) FROM devices WHERE spatial_addr <<= $1::INET`
	rows, err := r.db.QueryContext(ctx, q, bedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Device
	for rows.Next() {
		d := &domain.Device{}
		if err := rows.Scan(&d.DeviceID, &d.DeviceName); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetDevicesBoundToRoomsOrBeds 查询绑定到给定 room 或 bed 的设备（仅 id/name，用于 DeleteUnit 前检查）
// v2 stub：返回空让 DeleteUnit 视作"无设备绑定"放行。Phase E 重写。
func (r *PostgresDevicesRepository) GetDevicesBoundToRoomsOrBeds(ctx context.Context, tenantID string, roomIDs, bedIDs []string) ([]*domain.Device, error) {
	return []*domain.Device{}, nil
}

// CreateDevice 手动创建设备与位置的绑定关系（出库操作）
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) CreateDevice(ctx context.Context, tenantID string, device *domain.Device) (string, error) {
	return "", fmt.Errorf("CreateDevice not implemented in v2 yet (Phase E.2)")
}

// UpdateDevice 更新设备绑定（v2: 通过改 spatial_addr 实现 bind/unbind/migrate）
//
// FE 期望（v1 shape）：传 bound_room_id (/88 prefix) 或 bound_bed_id (/96 prefix) 或都 NULL（unbind）
// v2 实现：构造新 spatial_addr，bytes 0-N 取目标 prefix（room/bed/tenant 池），bytes (N+1)..11 清零，bytes 12-15 保留 MAC
//   - bound_bed_id 给定 → addr = bed_prefix(0..11) | (orig_addr & '::ffff:ffff'::INET) → bed-bound
//   - bound_room_id 给定 → addr = room_prefix(0..10) | zeros(11) | mac → room-bound
//   - 都 NULL → reset_device_prefix(addr, fd00::/32, 32, 'branch') 退回 tenant unbound pool
//   - 同时 monitoring_enabled 字段直接 update
func (r *PostgresDevicesRepository) UpdateDevice(ctx context.Context, tenantID, deviceID string, device *domain.Device) error {
	return r.UpdateDeviceWithFlags(ctx, tenantID, deviceID, device, true, true, true, true)
}

// UpdateDeviceWithFlags 同 UpdateDevice，但用 flag 区分字段是否要写
// 当前实现：bound_room_id / bound_bed_id 任一 flag 为 true 即重新计算 spatial_addr；monitoring_enabled flag 控制其更新
func (r *PostgresDevicesRepository) UpdateDeviceWithFlags(ctx context.Context, tenantID, deviceID string, device *domain.Device, updateBoundRoomID, updateBoundBedID, updateBusinessAccess, updateMonitoringEnabled bool) error {
	if deviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. 拿当前 spatial_addr
	var currentAddr sql.NullString
	if err := tx.QueryRowContext(ctx,
		`SELECT host(spatial_addr) FROM devices WHERE device_id = $1::UUID`, deviceID,
	).Scan(&currentAddr); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("device not found: %s", deviceID)
		}
		return fmt.Errorf("get current spatial_addr: %w", err)
	}
	if !currentAddr.Valid || currentAddr.String == "" {
		return fmt.Errorf("device has no spatial_addr (factory-only?)")
	}

	// 2. 计算新 spatial_addr（仅在 bind/unbind flag 触发时）
	bindFlag := updateBoundRoomID || updateBoundBedID
	if bindFlag {
		var newAddr string
		switch {
		case device.BoundBedID.Valid && device.BoundBedID.String != "":
			// bind 到 bed：bed_prefix(/96) | (current_addr & MAC32 mask)
			if err := tx.QueryRowContext(ctx, `
				SELECT host(
				  set_masklen(network(set_masklen($1::INET, 96)), 128)
				  | ($2::INET & '::ffff:ffff'::INET)
				) || '/128'
			`, device.BoundBedID.String, currentAddr.String).Scan(&newAddr); err != nil {
				return fmt.Errorf("compose bed-bound addr: %w", err)
			}
		case device.BoundRoomID.Valid && device.BoundRoomID.String != "":
			// bind 到 room：room_prefix(/88) | (current_addr & MAC32 mask)；byte 11 (bed) 自然为 0
			if err := tx.QueryRowContext(ctx, `
				SELECT host(
				  set_masklen(network(set_masklen($1::INET, 88)), 128)
				  | ($2::INET & '::ffff:ffff'::INET)
				) || '/128'
			`, device.BoundRoomID.String, currentAddr.String).Scan(&newAddr); err != nil {
				return fmt.Errorf("compose room-bound addr: %w", err)
			}
		default:
			// unbind：reset 到 branch 层（清 byte 6..11，保留 MAC + tenant）
			if err := tx.QueryRowContext(ctx, `
				SELECT host(reset_device_prefix(
				  $1::INET, 'fd00::/32'::INET, 32::SMALLINT, 'branch'
				)) || '/128'
			`, currentAddr.String).Scan(&newAddr); err != nil {
				return fmt.Errorf("reset to tenant pool: %w", err)
			}
		}
		// FK device_ota_spatial_addr_fkey 已 ON UPDATE CASCADE，device_ota.spatial_addr 自动同步
		if _, err := tx.ExecContext(ctx, `
			UPDATE devices SET spatial_addr = $2::INET, updated_at = NOW()
			 WHERE device_id = $1::UUID
		`, deviceID, newAddr); err != nil {
			return fmt.Errorf("update devices.spatial_addr: %w", err)
		}
	}

	// 3. monitoring_enabled
	if updateMonitoringEnabled {
		if _, err := tx.ExecContext(ctx,
			`UPDATE devices SET monitoring_enabled = $2, updated_at = NOW() WHERE device_id = $1::UUID`,
			deviceID, device.MonitoringEnabled); err != nil {
			return fmt.Errorf("update monitoring_enabled: %w", err)
		}
	}

	// 4. business_access — v2 已删该列；忽略写入
	_ = updateBusinessAccess

	return tx.Commit()
}

// DeleteDevice 删除设备（v2 软删通过 reset_device_prefix() 把 spatial_addr 退回 trash /48）
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) DeleteDevice(ctx context.Context, tenantID, deviceID string) error {
	return nil
}

// GetDeviceRelations 获取设备关联关系（设备、地址、住户）
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) GetDeviceRelations(ctx context.Context, tenantID, deviceID string) (*DeviceRelations, error) {
	return nil, fmt.Errorf("device not found: tenant_id=%s, device_id=%s", tenantID, deviceID)
}

// GetOrCreateDeviceFromStore 首次连接时自动创建设备记录
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) GetOrCreateDeviceFromStore(ctx context.Context, identifier string, mqttTopic string) (*domain.Device, error) {
	return nil, fmt.Errorf("unauthorized device: not registered in device_store")
}

// ============================================
// 批量查询（用于 ListUnitsWithFullHierarchy）
// ============================================

// GetDevicesByRoomIDs 返回每个 room 直接绑定的 devices（**不含**绑到 bed 的设备 — 避免重复计数）。
// roomIDs 是 v2 /88 prefix 字符串数组（如 'fd00:0:3:111::/88'）
// "直接绑定 room" = devices.spatial_addr 的 /88 网络等于 room.spatial_prefix AND byte 11 = 0
func (r *PostgresDevicesRepository) GetDevicesByRoomIDs(ctx context.Context, tenantID string, roomIDs []string) (map[string][]DeviceInfo, error) {
	out := make(map[string][]DeviceInfo, len(roomIDs))
	if len(roomIDs) == 0 {
		return out, nil
	}
	prefixes := make([]string, 0, len(roomIDs))
	for _, id := range roomIDs {
		if looksLikeINETPrefix(id) {
			prefixes = append(prefixes, id)
		}
	}
	if len(prefixes) == 0 {
		return out, nil
	}
	// byte 11 = 0：room-level（不绑到具体 bed）；用 inet bitwise AND 检测
	rows, err := r.db.QueryContext(ctx, `
		SELECT host(network(set_masklen(d.spatial_addr, 88))) || '/88' AS room_id,
		       d.device_id::text, dfm.device_uid
		  FROM devices d
		  JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		 WHERE network(set_masklen(d.spatial_addr, 88))::text = ANY($1::text[])
		   AND (d.spatial_addr & '::ff:0:0'::INET) = '::'::INET`,
		pq.Array(prefixes))
	if err != nil {
		return nil, fmt.Errorf("query devices by room IDs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var roomID, devID, devUID string
		if err := rows.Scan(&roomID, &devID, &devUID); err != nil {
			return nil, err
		}
		out[roomID] = append(out[roomID], DeviceInfo{ID: devID, Name: devUID})
	}
	return out, rows.Err()
}

// GetDevicesByBedIDs 返回每个 bed 绑定的 devices
// bedIDs 是 v2 /96 prefix 字符串数组（如 'fd00:0:3:111:0:200:0:1/96'）
func (r *PostgresDevicesRepository) GetDevicesByBedIDs(ctx context.Context, tenantID string, bedIDs []string) (map[string][]DeviceInfo, error) {
	out := make(map[string][]DeviceInfo, len(bedIDs))
	if len(bedIDs) == 0 {
		return out, nil
	}
	prefixes := make([]string, 0, len(bedIDs))
	for _, id := range bedIDs {
		if looksLikeINETPrefix(id) {
			prefixes = append(prefixes, id)
		}
	}
	if len(prefixes) == 0 {
		return out, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT host(network(set_masklen(d.spatial_addr, 96))) || '/96' AS bed_id,
		       d.device_id::text, dfm.device_uid
		  FROM devices d
		  JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		 WHERE network(set_masklen(d.spatial_addr, 96))::text = ANY($1::text[])`,
		pq.Array(prefixes))
	if err != nil {
		return nil, fmt.Errorf("query devices by bed IDs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var bedID, devID, devUID string
		if err := rows.Scan(&bedID, &devID, &devUID); err != nil {
			return nil, err
		}
		out[bedID] = append(out[bedID], DeviceInfo{ID: devID, Name: devUID})
	}
	return out, rows.Err()
}

// GetDeviceLocationInfo 获取设备的完整位置信息
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) GetDeviceLocationInfo(ctx context.Context, tenantID, deviceID string) (*DeviceLocationInfo, error) {
	return nil, fmt.Errorf("device not found: tenant_id=%s, device_id=%s", tenantID, deviceID)
}

// GetDeviceLocationInfoByIdentifier 通过设备标识符（device_uid）获取位置信息
// v2 stub: Phase E.2 will rewrite using devices.spatial_addr + reset_device_prefix()
func (r *PostgresDevicesRepository) GetDeviceLocationInfoByIdentifier(ctx context.Context, identifier string) (*DeviceLocationInfo, error) {
	return nil, fmt.Errorf("device not found: identifier=%s", identifier)
}
