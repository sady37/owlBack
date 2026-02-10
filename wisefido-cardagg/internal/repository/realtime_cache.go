package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"owl-common/card"

	redislib "github.com/go-redis/redis/v8"
)

// RedisCache Redis 缓存实现
type RedisCache struct {
	client *redislib.Client
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(client *redislib.Client) *RedisCache {
	return &RedisCache{client: client}
}

// GetCardRealTime 获取卡片realtime数据（TrackData/VitalData）
// 对应 StreamCardRealTime，由前端频繁调用
func (r *RedisCache) GetCardRealTime(ctx context.Context, cardID string) (*card.CardRealTime, error) {
	key := fmt.Sprintf("vital-focus:card:%s:realtime", cardID)

	val, err := r.client.Get(ctx, key).Result()
	if err == redislib.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data card.CardRealTime
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// GetCardStatus 获取卡片status数据（DeviceStatus/BedState/RoomState/ActiveAlarms）
// 对应 StreamCardStatus，低频变化数据
func (r *RedisCache) GetCardStatus(ctx context.Context, cardID string) (*card.CardStatus, error) {
	key := fmt.Sprintf("vital-focus:card:%s:status", cardID)

	val, err := r.client.Get(ctx, key).Result()
	if err == redislib.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data card.CardStatus
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// GetRealtimeData 获取卡片完整数据（合并monitor和status）
// 仅用于内部或系统初始化，前端应分别调用GetCardRealTime和GetCardStatus
// SetCardRealTime 设置卡片realtime数据（TrackData/VitalData）
// 对应 StreamCardRealTime，5秒TTL，每秒更新
func (r *RedisCache) SetCardRealTime(ctx context.Context, data *card.CardRealTime, ttl time.Duration) error {
	key := fmt.Sprintf("vital-focus:card:%s:realtime", data.CardID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, string(jsonData), ttl).Err()
}

// SetCardStatus 设置卡片status数据（DeviceStatus/BedState/RoomState/ActiveAlarms）
// 对应 StreamCardStatus，12小时TTL，变化时更新，不被realtime更新干扰
func (r *RedisCache) SetCardStatus(ctx context.Context, data *card.CardStatus) error {
	key := fmt.Sprintf("vital-focus:card:%s:status", data.CardID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 只写status key，12小时TTL
	return r.client.Set(ctx, key, string(jsonData), 12*time.Hour).Err()
}

// GetAllCardIds 获取所有卡片 ID (扫描 realtime key 提取唯一cardID)
func (r *RedisCache) GetAllCardIds(ctx context.Context) ([]string, error) {
	// 只扫描 realtime key，避免重复（status key会有相同的cardID）
	pattern := "vital-focus:card:*:realtime"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	// 从键中提取 cardID
	cardIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		// key format: "vital-focus:card:{cardID}:realtime"
		// 提取 cardID（从第 18 个字符到倒数 9 个字符）
		if len(key) > 27 { // "vital-focus:card:".len(18) + ":realtime".len(9) = 27
			cardID := key[18 : len(key)-9]
			cardIDs = append(cardIDs, cardID)
		}
	}

	return cardIDs, nil
}
