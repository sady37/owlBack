package repository

import (
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

// GetRoomInfo 获取房间信息
func (r *RoomRepository) GetRoomInfo(tenantID, roomID string) (*RoomInfo, error) {
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
	err := r.db.QueryRow(query, roomID, tenantID).Scan(
		&info.RoomID,
		&info.RoomName,
		&info.UnitID,
		&info.UnitName,
		&info.BedCount,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found: %s", roomID)
		}
		return nil, fmt.Errorf("failed to query room: %w", err)
	}
	
	return &info, nil
}

// IsBathroom 判断房间是否为卫生间
// 通过 room_name 或 unit_name 中是否包含以下词（不区分大小写）：
// - bathroom
// - restroom
// - toilet
func (r *RoomRepository) IsBathroom(tenantID, roomID string) (bool, error) {
	query := `
		SELECT 
			LOWER(r.room_name) as room_name,
			LOWER(u.unit_name) as unit_name
		FROM rooms r
		JOIN units u ON r.unit_id = u.unit_id AND r.tenant_id = u.tenant_id
		WHERE r.room_id = $1 AND r.tenant_id = $2
	`
	
	var roomName, unitName string
	err := r.db.QueryRow(query, roomID, tenantID).Scan(&roomName, &unitName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("room not found: %s", roomID)
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

// GetRoomByBedID 根据 bed_id 获取房间信息
func (r *RoomRepository) GetRoomByBedID(tenantID, bedID string) (*RoomInfo, error) {
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
	err := r.db.QueryRow(query, bedID, tenantID).Scan(
		&info.RoomID,
		&info.RoomName,
		&info.UnitID,
		&info.UnitName,
		&info.BedCount,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found for bed: %s", bedID)
		}
		return nil, fmt.Errorf("failed to query room by bed: %w", err)
	}
	
	return &info, nil
}

