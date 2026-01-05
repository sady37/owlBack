package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"github.com/lib/pq"
)

// CardRepository 卡片仓库
type CardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewCardRepository 创建卡片仓库
func NewCardRepository(db *sql.DB, logger *zap.Logger) *CardRepository {
	return &CardRepository{
		db:     db,
		logger: logger,
	}
}

// GetCardByDeviceID 根据设备ID获取关联的卡片
// 
// ⚠️ 重要依赖：
// - 本函数依赖 PostgreSQL cards 表，需要 wisefido-card-aggregator 服务先创建卡片
// - 如果 cards 表为空或设备未绑定到卡片，会返回错误
// - 当前 wisefido-card-aggregator 的卡片创建功能还未实现，需要优先实现
//
// 该方法根据设备的绑定关系（绑定到 Bed 或 Room）查询对应的卡片。
// 
// 查询逻辑（根据前端绑定规则和卡片创建规则）：
// 
// 前端确保：
// - 设备不能直接绑定到 Unit，必须绑定到 Room 或 Bed
// - 当设备绑定到 Unit 时，前端会先创建 unit_room（room_name === unit_name），然后绑定到 room
// - 所有 Bed 都绑定在 Room 下
// 
// 查询场景：
// 1. 如果设备绑定到 bed（bound_bed_id IS NOT NULL）：
//    - 查询 ActiveBed 类型的卡片（cards.bed_id = bound_bed_id）
// 2. 如果设备绑定到 room（bound_room_id IS NOT NULL）且未绑床：
//    - 通过 room.unit_id 查询 Location 类型的卡片（cards.unit_id = room.unit_id）
// 
// 注意：
// - 设备只能绑定到 Bed 或 Room 之一（互斥约束）
// - 如果设备未绑定或绑定关系不存在，返回错误
// - ⚠️ 如果 cards 表为空（卡片管理层未实现），所有查询都会失败
// 
// 参数:
//   - tenantID: 租户 ID（UUID 格式）
//   - deviceID: 设备 ID（UUID 格式）
// 
// 返回:
//   - *CardInfo: 卡片信息，包含 card_id、card_type、tenant_id 等
//   - error: 如果设备不存在、未绑定或查询失败
// 
// 示例:
//   card, err := repo.GetCardByDeviceID("tenant-123", "device-123")
//   if err != nil {
//       return nil, fmt.Errorf("获取卡片失败: %w", err)
//   }
func (r *CardRepository) GetCardByDeviceID(tenantID, deviceID string) (*CardInfo, error) {
	query := `
		WITH device_info AS (
			SELECT 
				d.device_id,
				d.tenant_id,
				d.bound_bed_id,
				d.bound_room_id
			FROM devices d
			WHERE d.device_id = $1 AND d.tenant_id = $2
		),
		bed_card AS (
			-- 场景 1：设备绑定到床，查询 ActiveBed 卡片
			-- 前端确保：所有 Bed 都绑定在 Room 下，设备绑定到 Bed 时，bound_bed_id IS NOT NULL
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.bed_id = di.bound_bed_id AND c.tenant_id = di.tenant_id
			WHERE di.bound_bed_id IS NOT NULL
			  AND c.card_type = 'ActiveBed'
			LIMIT 1
		),
		room_card AS (
			-- 场景 2：设备绑定到房间，通过 room.unit_id 查询 Location 卡片
			-- 前端确保：设备不能直接绑定到 Unit，必须绑定到 Room 或 Bed
			-- 当设备绑定到 Unit 时，前端会先创建 unit_room，然后绑定到 room
			-- 所以设备总是通过 bound_room_id 绑定，通过 room.unit_id 查询 Location 卡片
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.unit_id = (
				SELECT r.unit_id FROM rooms r WHERE r.room_id = di.bound_room_id AND r.tenant_id = di.tenant_id
			) AND c.tenant_id = di.tenant_id
			WHERE di.bound_room_id IS NOT NULL
			  AND di.bound_bed_id IS NULL
			  AND c.card_type = 'Location'
			LIMIT 1
		)
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM bed_card
		UNION ALL
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM room_card
		LIMIT 1
	`
	
	card := &CardInfo{}
	var bedID, unitID sql.NullString
	
	err := r.db.QueryRow(query, deviceID, tenantID).Scan(
		&card.CardID,
		&card.TenantID,
		&card.CardType,
		&bedID,
		&unitID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found for device: %s", deviceID)
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}
	
	if bedID.Valid {
		card.BedID = &bedID.String
	}
	if unitID.Valid {
		card.UnitID = &unitID.String
	}
	
	return card, nil
}

// GetCardByID 根据卡片ID获取卡片信息
func (r *CardRepository) GetCardByID(cardID string) (*CardInfo, error) {
	query := `
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM cards
		WHERE card_id = $1
	`
	
	card := &CardInfo{}
	var bedID, unitID sql.NullString
	
	err := r.db.QueryRow(query, cardID).Scan(
		&card.CardID,
		&card.TenantID,
		&card.CardType,
		&bedID,
		&unitID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}
	
	if bedID.Valid {
		card.BedID = &bedID.String
	}
	if unitID.Valid {
		card.UnitID = &unitID.String
	}
	
	return card, nil
}

// GetCardDevices 获取卡片关联的所有设备信息
//
// ⚠️ 重要依赖：
// - 本函数依赖 PostgreSQL cards 表，需要 wisefido-card-aggregator 服务先创建卡片
// - 从 cards.devices JSONB 字段读取设备列表（预计算结果）
// - 如果 cards 表为空或卡片不存在，会返回错误
// - 当前 wisefido-card-aggregator 的卡片创建功能还未实现，需要优先实现
//
// 参数:
//   - cardID: 卡片 ID（UUID 格式）
//
// 返回:
//   - []DeviceInfo: 卡片绑定的设备列表
//   - error: 如果卡片不存在或查询失败
func (r *CardRepository) GetCardDevices(cardID string) ([]DeviceInfo, error) {
	query := `
		SELECT devices
		FROM cards
		WHERE card_id = $1
	`
	
	var devicesJSON []byte
	err := r.db.QueryRow(query, cardID).Scan(&devicesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card devices: %w", err)
	}
	
	var devices []DeviceInfo
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}
	
	return devices, nil
}

// CardInfo 卡片信息
type CardInfo struct {
	CardID   string
	TenantID string
	CardType string // "ActiveBed" 或 "Location"
	BedID    *string
	UnitID   *string
}

// DeviceInfo 设备信息（从 cards.devices JSONB 解析）
type DeviceInfo struct {
	DeviceID    string  `json:"device_id"`
	DeviceName  string  `json:"device_name"`
	DeviceType  string  `json:"device_type"` // "Radar" 或 "Sleepace"
	DeviceModel string  `json:"device_model"`
	BedID       *string `json:"bed_id,omitempty"`       // 设备绑定的床ID（如果绑定到床）
	BedName     *string `json:"bed_name,omitempty"`     // 床名称（如果绑定到床）
	RoomID      *string `json:"room_id,omitempty"`      // 设备绑定的房间ID（如果绑定到房间）
	RoomName    *string `json:"room_name,omitempty"`    // 房间名称（如果绑定到房间）
	UnitID      string  `json:"unit_id"`                // 设备绑定的单元ID
}

// UpdateCardAlarmCounts 更新卡片的未处理报警计数
// 通过 card_id 找到该卡片关联的所有 device_id（从 cards.devices JSONB 中提取）
// 统计这些设备的未处理报警（alarm_status = 'active' 且 alarm_level IN ('0'-'4', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING')）
// 按 alarm_level 分组统计（映射到 0-4），更新 cards 表的 unhandled_alarm_0 到 unhandled_alarm_4 字段
func (r *CardRepository) UpdateCardAlarmCounts(ctx context.Context, tenantID, cardID string) error {
	// 1. 查询卡片关联的所有 device_id（从 cards.devices JSONB 中提取）
	query := `
		SELECT devices
		FROM cards
		WHERE tenant_id = $1 AND card_id = $2
	`
	
	var devicesJSON []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, cardID).Scan(&devicesJSON)
	if err != nil {
		return fmt.Errorf("failed to get card devices: %w", err)
	}
	
	// 2. 解析 devices JSONB，提取 device_id 列表
	var devices []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}
	
	if len(devices) == 0 {
		// 没有设备，将所有计数设为 0
		return r.updateCardAlarmCountsToZero(ctx, tenantID, cardID)
	}
	
	// 提取 device_id 列表
	deviceIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		if deviceID, ok := device["device_id"].(string); ok && deviceID != "" {
			deviceIDs = append(deviceIDs, deviceID)
		}
	}
	
	if len(deviceIDs) == 0 {
		// 没有有效的 device_id，将所有计数设为 0
		return r.updateCardAlarmCountsToZero(ctx, tenantID, cardID)
	}
	
	// 3. 统计这些设备的未处理报警（alarm_status = 'active'）
	// 按 alarm_level 分组统计（映射到 0-4）
	countQuery := `
		SELECT 
			COUNT(*) FILTER (WHERE alarm_level IN ('0', 'EMERG')) as count_0,
			COUNT(*) FILTER (WHERE alarm_level IN ('1', 'ALERT')) as count_1,
			COUNT(*) FILTER (WHERE alarm_level IN ('2', 'CRIT')) as count_2,
			COUNT(*) FILTER (WHERE alarm_level IN ('3', 'ERR')) as count_3,
			COUNT(*) FILTER (WHERE alarm_level IN ('4', 'WARNING')) as count_4
		FROM alarm_events
		WHERE tenant_id = $1
		  AND device_id = ANY($2::uuid[])
		  AND alarm_status = 'active'
		  AND alarm_level IN ('0', '1', '2', '3', '4', 'EMERG', 'ALERT', 'CRIT', 'ERR', 'WARNING')
		  AND (metadata->>'deleted_at' IS NULL)
	`
	
	var count0, count1, count2, count3, count4 int
	err = r.db.QueryRowContext(ctx, countQuery, tenantID, pq.Array(deviceIDs)).Scan(
		&count0, &count1, &count2, &count3, &count4,
	)
	if err != nil {
		return fmt.Errorf("failed to count alarm events: %w", err)
	}
	
	// 4. 更新 cards 表的 unhandled_alarm_0 到 unhandled_alarm_4 字段
	updateQuery := `
		UPDATE cards
		SET 
			unhandled_alarm_0 = $3,
			unhandled_alarm_1 = $4,
			unhandled_alarm_2 = $5,
			unhandled_alarm_3 = $6,
			unhandled_alarm_4 = $7
		WHERE tenant_id = $1 AND card_id = $2
	`
	
	_, err = r.db.ExecContext(ctx, updateQuery, tenantID, cardID, count0, count1, count2, count3, count4)
	if err != nil {
		return fmt.Errorf("failed to update card alarm counts: %w", err)
	}
	
	return nil
}

// updateCardAlarmCountsToZero 将卡片的报警计数全部设为 0
func (r *CardRepository) updateCardAlarmCountsToZero(ctx context.Context, tenantID, cardID string) error {
	updateQuery := `
		UPDATE cards
		SET 
			unhandled_alarm_0 = 0,
			unhandled_alarm_1 = 0,
			unhandled_alarm_2 = 0,
			unhandled_alarm_3 = 0,
			unhandled_alarm_4 = 0
		WHERE tenant_id = $1 AND card_id = $2
	`
	
	_, err := r.db.ExecContext(ctx, updateQuery, tenantID, cardID)
	if err != nil {
		return fmt.Errorf("failed to update card alarm counts to zero: %w", err)
	}
	
	return nil
}

