package repository

import (
	"context"
	"time"

	"owl-common/card"
)

// CacheRepository 缓存仓库接口
type CacheRepository interface {
	// GetCardRealTime 获取卡片realtime数据（TrackData/VitalData）
	GetCardRealTime(ctx context.Context, cardID string) (*card.CardRealTime, error)
	// GetCardStatus 获取卡片status数据（DeviceStatus/BedState/RoomState/ActiveAlarms）
	GetCardStatus(ctx context.Context, cardID string) (*card.CardStatus, error)
	// SetCardRealTime 设置卡片realtime数据（TrackData/VitalData，300秒TTL）
	SetCardRealTime(ctx context.Context, data *card.CardRealTime, ttl time.Duration) error
	// SetCardStatus 设置卡片status数据（DeviceStatus/BedState/RoomState/ActiveAlarms，12小时TTL）
	SetCardStatus(ctx context.Context, data *card.CardStatus) error
	// GetAllCardIds 获取所有卡片 ID
	GetAllCardIds(ctx context.Context) ([]string, error)
}
