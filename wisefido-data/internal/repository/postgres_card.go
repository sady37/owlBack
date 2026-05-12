package repository

// postgres_card.go — v2 (IPv6 spatial) 卡片仓储
//
// v2 schema 重写要点：
//   - cards 主键 card_id UUID；业务身份 spatial_prefix INET（mask 决定 card_type）
//   - 删除字段：tenant_id / branch_id / bed_id / unit_id / card_address / timezone /
//     devices JSONB / residents JSONB / unhandled_alarm_* / pop_alarm_*
//   - device 反查走 INET LPM：c.spatial_prefix >>= d.device_ipv6 / d.device_ipv6 <<= c.spatial_prefix
//   - tenant/branch/unit/bed 全部从 spatial_prefix 派生（set_masklen 截断），不再做 ID JOIN
//   - "tenantID" 参数语义：v2 中是 INET CIDR (/48) 字符串；保留参数位以最小化 caller 改动
//
// 同步记录器（CardSyncAffected）：UnitID 字段沿用 string 但内容现在是 /80 INET CIDR

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"owl-common/card"
	"wisefido-data/internal/domain"

	"go.uber.org/zap"
)

// RepositoryInterface 卡片仓储接口（建卡/同步用），由 PostgresCardRepository 实现。
// v2: 所有 ID 参数语义变更为 INET CIDR 字符串（tenantID=/48, unitID=/80, bedID=/96）
type RepositoryInterface interface {
	GetUnitInfo(tenantID, unitID string) (*card.UnitInfo, error)
	GetActiveBedsByUnit(tenantID, unitID string) ([]card.ActiveBedRow, error)
	GetDevicesByBed(tenantID, bedID string) ([]card.DeviceInfo, error)
	GetUnboundDevicesByUnit(tenantID, unitID string) ([]card.DeviceInfo, error)
	GetResidentByBed(tenantID, bedID string) (*card.ResidentInfo, error)
	GetResidentsByUnit(tenantID, unitID string) ([]card.ResidentInfo, error)
	GetCardsByUnit(tenantID, unitID string) ([]domain.Card, error)
	DeleteCard(tenantID, cardID string) error
	DeleteCardsByUnit(tenantID, unitID string) error
	CreateCard(spatialPrefix, cardType, cardName, dnsShortName string, residentID *string) (string, error)
	UpdateCard(cardID string, cardName *string, residentID *string, isActive *bool) error
	GetAllUnits(tenantID string) ([]string, error)
	GetUnitIDByBedID(tenantID, bedID string) (string, error)
	CountCardsByTenant(tenantID string) (int, error)
}

// PostgresCardRepository v2 实现：仅依赖 cards + spatial 层（units/rooms/beds/branches）
// + devices + device_factory_meta + device_runtime_state + residents + resident_unit。
type PostgresCardRepository struct {
	db         *sql.DB
	logger     *zap.Logger
	recorded   []domain.CardSyncAffected
	recordedMu sync.Mutex
}

// NewPostgresCardRepository creates a new card repository
func NewPostgresCardRepository(db *sql.DB, logger *zap.Logger) *PostgresCardRepository {
	return &PostgresCardRepository{
		db:     db,
		logger: logger,
	}
}

// =====================================================================
// 同步记录器
// =====================================================================

// ClearRecorded 清空本次同步记录
func (r *PostgresCardRepository) ClearRecorded() {
	r.recordedMu.Lock()
	defer r.recordedMu.Unlock()
	r.recorded = nil
}

// GetRecordedAndClear 返回本次同步记录并清空
func (r *PostgresCardRepository) GetRecordedAndClear() []domain.CardSyncAffected {
	r.recordedMu.Lock()
	defer r.recordedMu.Unlock()
	out := r.recorded
	r.recorded = nil
	return out
}

func (r *PostgresCardRepository) appendRecorded(op, tenantID, cardID, unitPrefix string, affectedDeviceUIDs []string) {
	r.recordedMu.Lock()
	defer r.recordedMu.Unlock()
	var u []string
	if len(affectedDeviceUIDs) > 0 {
		u = dedupeNonEmptyStrings(affectedDeviceUIDs)
	}
	r.recorded = append(r.recorded, domain.CardSyncAffected{
		TenantID:           tenantID,
		CardID:             cardID,
		UnitID:             unitPrefix,
		Op:                 op,
		AffectedDeviceUIDs: u,
	})
}

// =====================================================================
// Cache row（CardStatic 静态缓存使用）
// =====================================================================

// CardRowForCache 卡片缓存视图（v2：字段化，不再有 JSONB）
//
// v2 派生：
//   - TenantID = set_masklen(spatial_prefix, 48)  CIDR
//   - BranchID = set_masklen(spatial_prefix, 56)  CIDR（仅 mask >= 56 时有意义）
//   - UnitID   = set_masklen(spatial_prefix, 80)  CIDR（仅 mask >= 80 时有意义）
//   - BedID    = set_masklen(spatial_prefix, 96)  CIDR（仅 mask = 96 时有意义）
type CardRowForCache struct {
	CardID        string
	SpatialPrefix string  // INET CIDR
	CardType      string
	CardName      string
	DNSShortName  string
	ResidentID    *string
	IsActive      bool

	// LPM 派生 / JOIN 派生（service 层渲染用）
	TenantID   string
	BranchID   string
	BranchName string
	UnitID     string
	UnitName   string
	RoomID     *string
	RoomName   *string
	BedID      *string
	BedName    *string
	Building   string
	Timezone   string
}

// GetCardRowForCache 取卡片 + 关联 unit/branch/room/bed 信息（v2 spatial LPM）
//
// 注意：tenantID 参数在 v2 是 /48 INET CIDR，仅做 sanity check（cards 表无 tenant_id 列）。
func (r *PostgresCardRepository) GetCardRowForCache(ctx context.Context, tenantID, cardID string) (*CardRowForCache, error) {
	// v2 unified: card_id ≡ spatial_prefix；按 prefix 查
	query := `
		SELECT
			c.spatial_prefix::text AS card_id,
			c.spatial_prefix::text,
			c.card_type,
			COALESCE(c.card_name, ''),
			COALESCE(c.dns_short_name, ''),
			c.resident_id::text,
			c.is_active,
			-- 派生 prefix
			set_masklen(c.spatial_prefix, 48)::text AS tenant_prefix,
			CASE WHEN masklen(c.spatial_prefix) >= 56 THEN set_masklen(c.spatial_prefix, 56)::text ELSE '' END AS branch_prefix,
			CASE WHEN masklen(c.spatial_prefix) >= 80 THEN set_masklen(c.spatial_prefix, 80)::text ELSE '' END AS unit_prefix,
			CASE WHEN masklen(c.spatial_prefix) >= 88 THEN set_masklen(c.spatial_prefix, 88)::text ELSE NULL END AS room_prefix,
			CASE WHEN masklen(c.spatial_prefix) = 96 THEN set_masklen(c.spatial_prefix, 96)::text ELSE NULL END AS bed_prefix,
			-- 反查名字
			COALESCE(br.branch_name, ''),
			COALESCE(u.unit_name, ''),
			COALESCE(u.timezone, 'UTC'),
			rm.room_name,
			b.bed_name
		FROM cards c
		LEFT JOIN branches br ON br.branch_id = (CASE WHEN masklen(c.spatial_prefix) >= 56 THEN set_masklen(c.spatial_prefix, 56) ELSE NULL END)
		LEFT JOIN units    u  ON u.unit_id   = (CASE WHEN masklen(c.spatial_prefix) >= 80 THEN set_masklen(c.spatial_prefix, 80) ELSE NULL END)
		LEFT JOIN rooms    rm ON rm.room_id  = (CASE WHEN masklen(c.spatial_prefix) >= 88 THEN set_masklen(c.spatial_prefix, 88) ELSE NULL END)
		LEFT JOIN beds     b  ON b.bed_id    = (CASE WHEN masklen(c.spatial_prefix) = 96  THEN set_masklen(c.spatial_prefix, 96) ELSE NULL END)
		WHERE c.spatial_prefix = $1::INET
	`

	// 防止 unused variable warning（保留 tenantID 参数位）
	_ = tenantID

	var row CardRowForCache
	var residentID, roomID, bedID, roomName, bedName sql.NullString
	err := r.db.QueryRowContext(ctx, query, cardID).Scan(
		&row.CardID,
		&row.SpatialPrefix,
		&row.CardType,
		&row.CardName,
		&row.DNSShortName,
		&residentID,
		&row.IsActive,
		&row.TenantID,
		&row.BranchID,
		&row.UnitID,
		&roomID,
		&bedID,
		&row.BranchName,
		&row.UnitName,
		&row.Timezone,
		&roomName,
		&bedName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if residentID.Valid {
		row.ResidentID = &residentID.String
	}
	if roomID.Valid {
		row.RoomID = &roomID.String
	}
	if bedID.Valid {
		row.BedID = &bedID.String
	}
	if roomName.Valid {
		row.RoomName = &roomName.String
	}
	if bedName.Valid {
		row.BedName = &bedName.String
	}
	return &row, nil
}

// =====================================================================
// Branch / Unit / Bed 派生查询
// =====================================================================

// GetBranchIDByCard 由 card_id 反查 branch /56 INET CIDR 字符串。
func (r *PostgresCardRepository) GetBranchIDByCard(ctx context.Context, tenantID, cardID string) (string, error) {
	_ = tenantID
	if cardID == "" {
		return "", nil
	}
	var branchPrefix sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT CASE WHEN masklen(spatial_prefix) >= 56
		             THEN set_masklen(spatial_prefix, 56)::text
		             ELSE ''
		        END
		   FROM cards WHERE spatial_prefix = $1::INET`,
		cardID,
	).Scan(&branchPrefix)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if branchPrefix.Valid {
		return branchPrefix.String, nil
	}
	return "", nil
}

// GetBranchIDByUnit 由 unit_id INET /80 反查 branch /56 INET CIDR 字符串。
func (r *PostgresCardRepository) GetBranchIDByUnit(ctx context.Context, tenantID, unitID string) (string, error) {
	_ = tenantID
	if unitID == "" {
		return "", nil
	}
	var branchPrefix sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT set_masklen(unit_id, 56)::text FROM units WHERE unit_id = $1::inet`,
		unitID,
	).Scan(&branchPrefix)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if branchPrefix.Valid {
		return branchPrefix.String, nil
	}
	return "", nil
}

// GetActiveBedsByUnit 列出 unit 下所有"活跃床位"：
// 床下 device_ipv6 <<= bed_id (/96)，monitoring_enabled=TRUE，且有 device 命中。
func (r *PostgresCardRepository) GetActiveBedsByUnit(tenantID, unitID string) ([]card.ActiveBedRow, error) {
	_ = tenantID
	query := `
		SELECT
			b.bed_id::text,
			COALESCE(b.bed_name, '') AS bed_name,
			COALESCE(rm.room_name, '') AS room_name,
			MAX(ru.resident_id::text) AS resident_id
		FROM beds b
		INNER JOIN rooms rm ON rm.room_id = set_masklen(b.bed_id, 88)
		INNER JOIN devices d
		   ON d.device_ipv6 <<= b.bed_id
		  AND d.monitoring_enabled = TRUE
		LEFT JOIN resident_unit ru
		   ON ru.spatial_prefix = b.bed_id
		  AND ru.valid_to IS NULL
		WHERE set_masklen(b.bed_id, 80) = $1::inet
		GROUP BY b.bed_id, b.bed_name, rm.room_name
		HAVING COUNT(DISTINCT d.device_id) > 0
		ORDER BY bed_name ASC, b.bed_id ASC
	`

	rows, err := r.db.Query(query, unitID)
	if err != nil {
		return nil, fmt.Errorf("query active beds: %w", err)
	}
	defer rows.Close()

	var beds []card.ActiveBedRow
	for rows.Next() {
		var bedID, bedName, roomName string
		var residentID sql.NullString
		if err := rows.Scan(&bedID, &bedName, &roomName, &residentID); err != nil {
			return nil, fmt.Errorf("scan bed: %w", err)
		}
		bed := card.ActiveBedRow{
			BedID:    bedID,
			BedName:  bedName,
			RoomName: roomName,
		}
		if residentID.Valid {
			bed.ResidentID = &residentID.String
		}
		beds = append(beds, bed)
	}
	return beds, rows.Err()
}

// GetUnitInfo 由 unit_id INET /80 拿 unit + branch 信息。
func (r *PostgresCardRepository) GetUnitInfo(tenantID, unitID string) (*card.UnitInfo, error) {
	_ = tenantID
	query := `
		SELECT
			u.unit_id::text,
			u.unit_name,
			COALESCE(br.branch_name, '') AS branch_name,
			-- v2 无 buildings 表，留空
			''::text AS building,
			-- v2 unit_type: 1=single, 2=share, 3=public; is_public = (unit_type=3)
			(u.unit_type = 3) AS is_public,
			(u.unit_type = 2) AS is_shared_unit,
			CASE u.unit_property WHEN 0 THEN 'home' ELSE 'facility' END AS unit_type_kind,
			COALESCE(u.timezone, 'UTC') AS timezone
		FROM units u
		LEFT JOIN branches br ON br.branch_id = set_masklen(u.unit_id, 56)
		WHERE u.unit_id = $1::inet
	`

	var unit card.UnitInfo
	err := r.db.QueryRow(query, unitID).Scan(
		&unit.UnitID,
		&unit.UnitName,
		&unit.BranchName,
		&unit.Building,
		&unit.IsPublic,
		&unit.IsSharedUnit,
		&unit.UnitType,
		&unit.Timezone,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit not found: %s", unitID)
		}
		return nil, fmt.Errorf("failed to query unit: %w", err)
	}
	return &unit, nil
}

// =====================================================================
// 设备查询（按床 / 按 unit 未绑床）
// =====================================================================

// GetDevicesByBed 床上所有 monitoring_enabled=TRUE 设备。
func (r *PostgresCardRepository) GetDevicesByBed(tenantID, bedID string) ([]card.DeviceInfo, error) {
	_ = tenantID
	query := `
		SELECT
			d.device_id::text,
			COALESCE(dfm.device_uid, '')   AS device_uid,
			COALESCE(dfm.device_code, '')  AS device_code,
			''::text                       AS device_name,
			COALESCE(dfm.device_type::text, '') AS device_type,
			COALESCE(dfm.device_model, '') AS device_model,
			set_masklen(d.device_ipv6, 96)::text AS bound_bed,
			CASE WHEN masklen(d.device_ipv6) >= 88
			     THEN set_masklen(d.device_ipv6, 88)::text
			     ELSE NULL END             AS bound_room,
			set_masklen(d.device_ipv6, 80)::text AS unit_prefix,
			d.monitoring_enabled,
			CASE WHEN drs.online IS TRUE THEN 'online' ELSE 'offline' END AS status
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		LEFT JOIN device_runtime_state drs ON drs.device_id = d.device_id
		WHERE d.device_ipv6 <<= $1::inet
		  AND d.monitoring_enabled = TRUE
		ORDER BY d.device_ipv6
	`
	rows, err := r.db.Query(query, bedID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	return scanDevices(rows)
}

// GetUnboundDevicesByUnit unit 下未在任何 card 命中的 devices（即孤儿）
// v2: 用 find_card_by_device_addr IS NULL 判断没有 LPM 命中的 card。
func (r *PostgresCardRepository) GetUnboundDevicesByUnit(tenantID, unitID string) ([]card.DeviceInfo, error) {
	_ = tenantID
	query := `
		SELECT
			d.device_id::text,
			COALESCE(dfm.device_uid, '')   AS device_uid,
			COALESCE(dfm.device_code, '')  AS device_code,
			''::text                       AS device_name,
			COALESCE(dfm.device_type::text, '') AS device_type,
			COALESCE(dfm.device_model, '') AS device_model,
			CASE WHEN masklen(d.device_ipv6) = 128
			     THEN set_masklen(d.device_ipv6, 96)::text
			     ELSE NULL END             AS bound_bed,
			CASE WHEN masklen(d.device_ipv6) >= 88
			     THEN set_masklen(d.device_ipv6, 88)::text
			     ELSE NULL END             AS bound_room,
			set_masklen(d.device_ipv6, 80)::text AS unit_prefix,
			d.monitoring_enabled,
			CASE WHEN drs.online IS TRUE THEN 'online' ELSE 'offline' END AS status
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		LEFT JOIN device_runtime_state drs ON drs.device_id = d.device_id
		WHERE d.device_ipv6 <<= $1::inet
		  AND d.monitoring_enabled = TRUE
		  AND find_card_by_device_addr(d.device_ipv6) IS NULL
		ORDER BY d.device_ipv6
	`
	rows, err := r.db.Query(query, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unbound devices: %w", err)
	}
	defer rows.Close()
	return scanDevices(rows)
}

// scanDevices 公共 device 扫描逻辑
func scanDevices(rows *sql.Rows) ([]card.DeviceInfo, error) {
	var devices []card.DeviceInfo
	for rows.Next() {
		var di card.DeviceInfo
		var boundBed, boundRoom, unitPrefix sql.NullString
		if err := rows.Scan(
			&di.DeviceID,
			&di.DeviceUID,
			&di.DeviceCode,
			&di.DeviceName,
			&di.DeviceType,
			&di.DeviceModel,
			&boundBed,
			&boundRoom,
			&unitPrefix,
			&di.MonitoringEnabled,
			&di.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		if boundBed.Valid {
			di.BoundBedID = &boundBed.String
		}
		if boundRoom.Valid {
			di.BoundRoomID = &boundRoom.String
		}
		if unitPrefix.Valid {
			di.UnitID = unitPrefix.String
		}
		devices = append(devices, di)
	}
	return devices, rows.Err()
}

// =====================================================================
// Resident 查询（current episode = valid_to IS NULL）
// =====================================================================

// GetResidentByBed 床上当前 resident（valid_to IS NULL 的 episode）。
// 不返回 PHI（last_name 加密在 resident_phi）。
func (r *PostgresCardRepository) GetResidentByBed(tenantID, bedID string) (*card.ResidentInfo, error) {
	_ = tenantID
	query := `
		SELECT
			r.resident_id::text,
			COALESCE(r.nickname, '') AS nickname,
			ru.spatial_prefix::text  AS bed_prefix
		FROM residents r
		JOIN resident_unit ru ON ru.resident_id = r.resident_id
		WHERE ru.spatial_prefix = $1::inet
		  AND ru.valid_to IS NULL
		ORDER BY ru.valid_from DESC
		LIMIT 1
	`
	var resident card.ResidentInfo
	var bedPrefix sql.NullString
	err := r.db.QueryRow(query, bedID).Scan(
		&resident.ResidentID,
		&resident.Nickname,
		&bedPrefix,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query resident: %w", err)
	}
	if bedPrefix.Valid {
		resident.BedID = &bedPrefix.String
	}
	return &resident, nil
}

// GetResidentsByUnit unit 下所有 current residents（valid_to IS NULL）。
func (r *PostgresCardRepository) GetResidentsByUnit(tenantID, unitID string) ([]card.ResidentInfo, error) {
	_ = tenantID
	query := `
		SELECT
			r.resident_id::text,
			COALESCE(r.nickname, ''),
			ru.spatial_prefix::text
		FROM residents r
		JOIN resident_unit ru ON ru.resident_id = r.resident_id
		WHERE set_masklen(ru.spatial_prefix, 80) = $1::inet
		  AND ru.valid_to IS NULL
		ORDER BY r.nickname
	`
	rows, err := r.db.Query(query, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query residents: %w", err)
	}
	defer rows.Close()

	var residents []card.ResidentInfo
	for rows.Next() {
		var resident card.ResidentInfo
		var bedPrefix sql.NullString
		if err := rows.Scan(&resident.ResidentID, &resident.Nickname, &bedPrefix); err != nil {
			return nil, fmt.Errorf("failed to scan resident: %w", err)
		}
		if bedPrefix.Valid {
			resident.BedID = &bedPrefix.String
		}
		residents = append(residents, resident)
	}
	return residents, rows.Err()
}

// =====================================================================
// Unit / Bed 列表
// =====================================================================

// GetAllUnits 返回 tenant 下所有 unit 的 INET CIDR 字符串列表。
// tenantID = /48 prefix CIDR；v2 用 unit_id <<= tenant_prefix 反查。
func (r *PostgresCardRepository) GetAllUnits(tenantID string) ([]string, error) {
	var rows *sql.Rows
	var err error
	if tenantID == "" {
		rows, err = r.db.Query(`SELECT unit_id::text FROM units ORDER BY unit_name`)
	} else {
		rows, err = r.db.Query(
			`SELECT unit_id::text FROM units WHERE unit_id <<= $1::inet ORDER BY unit_name`,
			tenantID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query units: %w", err)
	}
	defer rows.Close()

	var unitIDs []string
	for rows.Next() {
		var unitID string
		if err := rows.Scan(&unitID); err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		unitIDs = append(unitIDs, unitID)
	}
	return unitIDs, rows.Err()
}

// GetUnitIDByBedID 从 bed_id INET /96 派生 unit /80 CIDR。
func (r *PostgresCardRepository) GetUnitIDByBedID(tenantID, bedID string) (string, error) {
	_ = tenantID
	if bedID == "" {
		return "", fmt.Errorf("bed_id is required")
	}
	var unitID string
	err := r.db.QueryRow(
		`SELECT set_masklen(bed_id, 80)::text FROM beds WHERE bed_id = $1::inet LIMIT 1`,
		bedID,
	).Scan(&unitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("bed not found: %s", bedID)
		}
		return "", fmt.Errorf("failed to query unit_id: %w", err)
	}
	return unitID, nil
}

// =====================================================================
// Cards 查询（CRUD）
// =====================================================================

// GetCardsByUnit unit 下所有 cards（v2 改返回 []domain.Card，去 JSONB）
// 包含 unit 自身 (/80) 和其后代 (/88 /96 /128)
func (r *PostgresCardRepository) GetCardsByUnit(tenantID, unitID string) ([]domain.Card, error) {
	_ = tenantID
	// v2 unified: card_id ≡ spatial_prefix
	query := `
		SELECT spatial_prefix::text AS card_id, spatial_prefix::text, card_type,
		       card_name, dns_short_name, resident_id::text,
		       is_active, enabled_at, disabled_at, created_at, updated_at
		FROM cards
		WHERE spatial_prefix <<= $1::inet
		ORDER BY masklen(spatial_prefix), spatial_prefix
	`
	rows, err := r.db.Query(query, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var cards []domain.Card
	for rows.Next() {
		var c domain.Card
		var residentID sql.NullString
		if err := rows.Scan(
			&c.CardID, &c.SpatialPrefix, &c.CardType,
			&c.CardName, &c.DNSShortName, &residentID,
			&c.IsActive, &c.EnabledAt, &c.DisabledAt,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		if residentID.Valid {
			c.ResidentID = sql.NullString{String: residentID.String, Valid: true}
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// DeleteCard 删除卡片；同时 expire 该 card LPM 命中的所有 device 的 active alarms。
func (r *PostgresCardRepository) DeleteCard(tenantID, cardID string) error {
	_ = tenantID

	// v2 unified: cardID is INET CIDR (= spatial_prefix)
	spatialPrefix := cardID
	// Sanity check exists
	var existsRow sql.NullString
	err := r.db.QueryRow(`SELECT spatial_prefix::text FROM cards WHERE spatial_prefix = $1::INET`, cardID).Scan(&existsRow)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("query card before delete: %w", err)
	}
	uids := r.getCardDeviceUIDList(tenantID, cardID)
	deviceIDs := r.getCardDeviceIDList(cardID)

	if len(deviceIDs) > 0 {
		if err := card.ExpireAlarmsByDeviceIDs(context.Background(), r.db, tenantID, deviceIDs, "card_delete"); err != nil {
			r.logger.Warn("DeleteCard: failed to expire alarms",
				zap.String("card_id", cardID), zap.Error(err))
		}
	}

	if _, err := r.db.Exec(`DELETE FROM cards WHERE spatial_prefix = $1::INET`, cardID); err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	// unit prefix = /80 截断（若 mask < 80 则原样）
	unitPrefix := spatialPrefix
	r.appendRecorded("delete", tenantID, cardID, unitPrefix, uids)
	return nil
}

// ListAllCardsForClear 列出全部 cards（用于清表前发 deleted 通知）。
// v2: 无 tenant_id 列；TenantID 字段由 spatial_prefix /48 派生。
func (r *PostgresCardRepository) ListAllCardsForClear(ctx context.Context) ([]domain.CardSyncAffected, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT set_masklen(spatial_prefix, 48)::text AS tenant_prefix,
		       spatial_prefix::text AS card_id,
		       spatial_prefix::text
		FROM cards
	`)
	if err != nil {
		return nil, fmt.Errorf("list cards for clear: %w", err)
	}
	defer rows.Close()
	var out []domain.CardSyncAffected
	for rows.Next() {
		var a domain.CardSyncAffected
		if err := rows.Scan(&a.TenantID, &a.CardID, &a.UnitID); err != nil {
			return nil, err
		}
		a.Op = "deleted"
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListOrphanCards orphan = 没有任何 device LPM 命中的 card（且不是 tenant/branch/site 这种宏卡）。
func (r *PostgresCardRepository) ListOrphanCards(ctx context.Context) ([]domain.CardSyncAffected, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT set_masklen(c.spatial_prefix, 48)::text AS tenant_prefix,
		       c.spatial_prefix::text AS card_id,
		       c.spatial_prefix::text
		FROM cards c
		LEFT JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		WHERE c.card_type IN ('active_bed','room','device','public','unit')
		GROUP BY c.spatial_prefix
		HAVING COUNT(d.device_id) = 0
	`)
	if err != nil {
		return nil, fmt.Errorf("list orphan cards: %w", err)
	}
	defer rows.Close()
	var out []domain.CardSyncAffected
	for rows.Next() {
		var a domain.CardSyncAffected
		if err := rows.Scan(&a.TenantID, &a.CardID, &a.UnitID); err != nil {
			return nil, err
		}
		a.Op = "deleted"
		out = append(out, a)
	}
	return out, rows.Err()
}

// ClearAllCards 全删（admin/灾难恢复）。
func (r *PostgresCardRepository) ClearAllCards() error {
	result, err := r.db.Exec(`DELETE FROM cards`)
	if err != nil {
		return fmt.Errorf("clear all cards: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Cleared all cards globally", zap.Int64("rows_affected", rowsAffected))
	r.ClearRecorded()
	return nil
}

// DeleteCardsByUnit 删除 unit prefix 下所有 cards（/80 + 后代）
func (r *PostgresCardRepository) DeleteCardsByUnit(tenantID, unitID string) error {
	// 先收集 card_id 列表（v2: ≡ spatial_prefix），记录到 recorded（含 device_uid）
	rows, err := r.db.Query(
		`SELECT spatial_prefix::text FROM cards WHERE spatial_prefix <<= $1::inet`,
		unitID,
	)
	if err != nil {
		return fmt.Errorf("failed to list cards: %w", err)
	}
	var cardIDs []string
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			rows.Close()
			return err
		}
		cardIDs = append(cardIDs, cid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, cid := range cardIDs {
		r.appendRecorded("deleted", tenantID, cid, unitID, r.getCardDeviceUIDList(tenantID, cid))
	}

	if _, err := r.db.Exec(`DELETE FROM cards WHERE spatial_prefix <<= $1::inet`, unitID); err != nil {
		return fmt.Errorf("failed to delete cards: %w", err)
	}
	return nil
}

// CountCardsByTenant tenant /48 下 card 总数。
func (r *PostgresCardRepository) CountCardsByTenant(tenantID string) (int, error) {
	var count int
	var err error
	if tenantID == "" {
		err = r.db.QueryRow(`SELECT COUNT(*) FROM cards`).Scan(&count)
	} else {
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM cards WHERE spatial_prefix <<= $1::inet`,
			tenantID,
		).Scan(&count)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to count cards: %w", err)
	}
	return count, nil
}

// UpdateCard 改 card_name / resident_id pointer / is_active。
//   - cardName nil 不动；非 nil 即使空字符串也写入
//   - residentID nil 不动；空字符串 / "NULL" 清空；其他视为 INET 写入
//   - isActive nil 不动；非 nil 时同步 enabled_at/disabled_at
func (r *PostgresCardRepository) UpdateCard(
	cardID string,
	cardName *string,
	residentID *string,
	isActive *bool,
) error {
	if cardID == "" {
		return fmt.Errorf("card_id is required")
	}

	// 动态拼 SET 子句
	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	idx := 1

	if cardName != nil {
		sets = append(sets, fmt.Sprintf("card_name = $%d", idx))
		args = append(args, *cardName)
		idx++
	}
	if residentID != nil {
		s := strings.TrimSpace(*residentID)
		if s == "" || strings.EqualFold(s, "NULL") {
			sets = append(sets, "resident_id = NULL")
		} else {
			sets = append(sets, fmt.Sprintf("resident_id = $%d::inet", idx))
			args = append(args, s)
			idx++
		}
	}
	if isActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", idx))
		args = append(args, *isActive)
		idx++
		if *isActive {
			sets = append(sets, "enabled_at = NOW()", "disabled_at = NULL")
		} else {
			sets = append(sets, "disabled_at = NOW()")
		}
	}

	// v2 unified: cardID is INET CIDR (= spatial_prefix)
	args = append(args, cardID)
	query := fmt.Sprintf(`UPDATE cards SET %s WHERE spatial_prefix = $%d::INET`,
		strings.Join(sets, ", "), idx)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("card not found: %s", cardID)
	}

	// spatial_prefix ≡ cardID
	spatialPrefix := cardID

	// derive tenant prefix
	tenantPrefix := ""
	if spatialPrefix != "" {
		_ = r.db.QueryRow(
			`SELECT set_masklen($1::inet, 48)::text`, spatialPrefix,
		).Scan(&tenantPrefix)
	}

	// 重算 alarm 计数（v2 stub，安全调用）
	if _, err := card.RecalcCardAlarmState(context.Background(), r.db, cardID, tenantPrefix); err != nil {
		r.logger.Warn("UpdateCard: failed to recalc alarm counts",
			zap.String("card_id", cardID), zap.Error(err))
	}

	uids := r.getCardDeviceUIDList(tenantPrefix, cardID)
	r.appendRecorded("updated", tenantPrefix, cardID, spatialPrefix, uids)
	return nil
}

// GetResidentCardIDByPrefix — 反查给定 spatial_prefix 上承载 resident 的 card_id。
// v2 unified: card_id ≡ spatial_prefix；本函数仅做存在性校验，命中即返回 prefix。
//
// 设计：一个 spatial_prefix PK 下只有一张 card，masklen↔card_type 1:1（CardTypeForMasklen）。
func (r *PostgresCardRepository) GetResidentCardIDByPrefix(ctx context.Context, prefix string) (string, error) {
	if strings.TrimSpace(prefix) == "" {
		return "", nil
	}
	var found sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT spatial_prefix::text
		  FROM cards
		 WHERE spatial_prefix = $1::INET
		 LIMIT 1
	`, prefix).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("lookup card by prefix: %w", err)
	}
	if !found.Valid {
		return "", nil
	}
	return found.String, nil
}

// CreateCard 创建卡片（v2 签名）。
//
//	spatialPrefix: INET CIDR (e.g. "fd00:0:1:112:42:301::/96")
//	cardType:      'active_bed' / 'unit' / 'public' / 'room' / 'device' / ...
//	cardName:      显示名（可空）
//	dnsShortName:  永久 DNS 短域名（可空；UNIQUE）
//	residentID:    INET HoA 字符串指针；nil 或空 = NULL
//
// 同 spatial_prefix + card_type 命中 UNIQUE 约束时走 UPDATE 路径（idempotent）。
func (r *PostgresCardRepository) CreateCard(
	spatialPrefix string,
	cardType string,
	cardName string,
	dnsShortName string,
	residentID *string,
) (string, error) {
	if spatialPrefix == "" || cardType == "" {
		return "", fmt.Errorf("spatial_prefix and card_type are required")
	}

	var residentArg interface{}
	if residentID != nil {
		s := strings.TrimSpace(*residentID)
		if s != "" && !strings.EqualFold(s, "NULL") {
			residentArg = s
		}
	}

	var dnsArg interface{}
	if strings.TrimSpace(dnsShortName) != "" {
		dnsArg = dnsShortName
	}

	// v2 unified: card_id ≡ spatial_prefix；UPSERT 用 ON CONFLICT (spatial_prefix)
	query := `
		INSERT INTO cards (
			spatial_prefix, card_type, card_name, dns_short_name, resident_id, is_active, enabled_at
		) VALUES ($1::inet, $2, $3, $4, $5::inet, TRUE, NOW())
		ON CONFLICT (spatial_prefix) DO UPDATE SET
			card_name      = COALESCE(EXCLUDED.card_name, cards.card_name),
			dns_short_name = COALESCE(EXCLUDED.dns_short_name, cards.dns_short_name),
			resident_id    = EXCLUDED.resident_id,
			is_active      = TRUE,
			enabled_at     = COALESCE(cards.enabled_at, NOW()),
			updated_at     = NOW()
		RETURNING spatial_prefix::text
	`

	var cardID string
	err := r.db.QueryRow(query, spatialPrefix, cardType, cardName, dnsArg, residentArg).Scan(&cardID)
	if err != nil {
		return "", fmt.Errorf("failed to create card: %w", err)
	}

	// derive tenant prefix
	var tenantPrefix string
	_ = r.db.QueryRow(`SELECT set_masklen($1::inet, 48)::text`, spatialPrefix).Scan(&tenantPrefix)

	uids := r.getCardDeviceUIDList(tenantPrefix, cardID)
	r.appendRecorded("created", tenantPrefix, cardID, spatialPrefix, uids)
	return cardID, nil
}

// =====================================================================
// 内部 helpers
// =====================================================================

// getCardDeviceIDList card LPM 命中的所有 device_id（UUID string list）。
func (r *PostgresCardRepository) getCardDeviceIDList(cardID string) []string {
	rows, err := r.db.Query(`
		SELECT d.device_id::text
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		WHERE c.spatial_prefix = $1::INET
	`, cardID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil && id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// getCardDeviceUIDList card LPM 命中的所有 device_uid（来自 device_factory_meta）。
func (r *PostgresCardRepository) getCardDeviceUIDList(tenantID, cardID string) []string {
	_ = tenantID
	if cardID == "" {
		return nil
	}
	rows, err := r.db.Query(`
		SELECT dfm.device_uid
		FROM cards c
		JOIN devices d ON d.device_ipv6 <<= c.spatial_prefix
		JOIN device_factory_meta dfm ON dfm.device_id = d.device_id
		WHERE c.spatial_prefix = $1::INET
	`, cardID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var uids []string
	for rows.Next() {
		var uid sql.NullString
		if err := rows.Scan(&uid); err == nil && uid.Valid {
			if s := strings.TrimSpace(uid.String); s != "" {
				uids = append(uids, s)
			}
		}
	}
	return dedupeNonEmptyStrings(uids)
}

// GetDeviceUIDsForCard 公开版（去重）。
func (r *PostgresCardRepository) GetDeviceUIDsForCard(tenantID, cardID string) []string {
	return r.getCardDeviceUIDList(tenantID, cardID)
}

// dedupeNonEmptyStrings 通用 helper：去空白 + 去重。
func dedupeNonEmptyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
