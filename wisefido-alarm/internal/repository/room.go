package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// RoomRepository 房间仓库（用于报警评估）
type RoomRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewRoomRepository 创建房间仓库
func NewRoomRepository(db *sql.DB, logger *zap.Logger) *RoomRepository {
	return &RoomRepository{
		db:     db,
		logger: logger,
	}
}

// RoomInfo 房间信息
type RoomInfo struct {
	RoomID   string
	RoomName string
	UnitID   string
	UnitName string
	BedCount int // 房间内的床数量
}

// GetRoomInfo 获取房间信息（需验证 tenant_id）
func (r *RoomRepository) GetRoomInfo(ctx context.Context, tenantID, roomID string) (*RoomInfo, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if roomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	query := `
		SELECT 
			r.room_id,
			r.room_name,
			r.unit_id,
			u.unit_name,
			(SELECT COUNT(*) FROM beds b WHERE b.room_id = r.room_id AND b.tenant_id = r.tenant_id) as bed_count
		FROM rooms r
		JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE r.room_id = $1 AND r.tenant_id = $2
	`
	
	var info RoomInfo
	err := r.db.QueryRowContext(ctx, query, roomID, tenantID).Scan(
		&info.RoomID,
		&info.RoomName,
		&info.UnitID,
		&info.UnitName,
		&info.BedCount,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found: room_id=%s, tenant_id=%s", roomID, tenantID)
		}
		return nil, fmt.Errorf("failed to query room: %w", err)
	}
	
	return &info, nil
}

// IsBathroom 判断房间是否为卫生间（需验证 tenant_id）
// 通过 room_name 或 unit_name 中是否包含以下词（不区分大小写）：
// - bathroom
// - restroom
// - toilet
func (r *RoomRepository) IsBathroom(ctx context.Context, tenantID, roomID string) (bool, error) {
	if tenantID == "" {
		return false, fmt.Errorf("tenant_id is required")
	}
	if roomID == "" {
		return false, fmt.Errorf("room_id is required")
	}

	query := `
		SELECT 
			LOWER(r.room_name) as room_name,
			LOWER(u.unit_name) as unit_name
		FROM rooms r
		JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE r.room_id = $1 AND r.tenant_id = $2
	`
	
	var roomName, unitName string
	err := r.db.QueryRowContext(ctx, query, roomID, tenantID).Scan(&roomName, &unitName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("room not found: room_id=%s, tenant_id=%s", roomID, tenantID)
		}
		return false, fmt.Errorf("failed to query room: %w", err)
	}
	
	// 检查是否包含卫生间关键词
	bathroomKeywords := []string{"bathroom", "restroom", "toilet"}
	
	for _, keyword := range bathroomKeywords {
		if strings.Contains(roomName, keyword) || strings.Contains(unitName, keyword) {
			return true, nil
		}
	}
	
	return false, nil
}

// GetRoomByBedID 根据 bed_id 获取房间信息（需验证 tenant_id）
func (r *RoomRepository) GetRoomByBedID(ctx context.Context, tenantID, bedID string) (*RoomInfo, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if bedID == "" {
		return nil, fmt.Errorf("bed_id is required")
	}

	query := `
		SELECT 
			r.room_id,
			r.room_name,
			r.unit_id,
			u.unit_name,
			(SELECT COUNT(*) FROM beds b WHERE b.room_id = r.room_id AND b.tenant_id = r.tenant_id) as bed_count
		FROM rooms r
		JOIN beds b ON r.room_id = b.room_id AND r.tenant_id = b.tenant_id
		JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE b.bed_id = $1 AND b.tenant_id = $2
	`
	
	var info RoomInfo
	err := r.db.QueryRowContext(ctx, query, bedID, tenantID).Scan(
		&info.RoomID,
		&info.RoomName,
		&info.UnitID,
		&info.UnitName,
		&info.BedCount,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found for bed: bed_id=%s, tenant_id=%s", bedID, tenantID)
		}
		return nil, fmt.Errorf("failed to query room by bed: %w", err)
	}
	
	return &info, nil
}

