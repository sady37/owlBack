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

// GetRealtimeData 获取卡片实时数据 (使用 hash key 模式，每个字段独立 TTL)
func (r *RedisCache) GetRealtimeData(ctx context.Context, cardID string) (*card.RealtimeData, error) {
	data := &card.RealtimeData{
		CardID: cardID,
	}

	// 1. 获取 bed_state (TTL: 12H)
	bedStateKey := fmt.Sprintf("vital-focus:card:%s:bed_state", cardID)
	if val, err := r.client.Get(ctx, bedStateKey).Result(); err == nil {
		var bedState card.BedState
		if err := json.Unmarshal([]byte(val), &bedState); err == nil {
			data.BedState = &bedState
		}
	}

	// 2. 获取 room_state (TTL: 12H)
	roomStateKey := fmt.Sprintf("vital-focus:card:%s:room_state", cardID)
	if val, err := r.client.Get(ctx, roomStateKey).Result(); err == nil {
		var roomState card.RoomState
		if err := json.Unmarshal([]byte(val), &roomState); err == nil {
			data.RoomState = &roomState
		}
	}

	// 3. 获取 active_alarms (TTL: 5s)
	alarmsKey := fmt.Sprintf("vital-focus:card:%s:active_alarms", cardID)
	if val, err := r.client.Get(ctx, alarmsKey).Result(); err == nil {
		var alarms card.ActiveAlarmState
		if err := json.Unmarshal([]byte(val), &alarms); err == nil {
			data.ActiveAlarms = &alarms
		}
	}

	// 4. 获取所有 vital 数据 (TTL: 5s)
	vitalPattern := fmt.Sprintf("vital-focus:card:%s:vital:*", cardID)
	vitalKeys, _ := r.client.Keys(ctx, vitalPattern).Result()
	vitals := make([]card.VitalSimplified, 0, len(vitalKeys))
	for _, key := range vitalKeys {
		if val, err := r.client.Get(ctx, key).Result(); err == nil {
			var vital card.VitalSimplified
			if err := json.Unmarshal([]byte(val), &vital); err == nil {
				vitals = append(vitals, vital)
			}
		}
	}
	if len(vitals) > 0 {
		data.Vital = vitals
	}

	// 5. 获取所有 posture 数据，组织成数组 (TTL: 5s)
	posturePattern := fmt.Sprintf("vital-focus:card:%s:posture:*", cardID)
	postureKeys, _ := r.client.Keys(ctx, posturePattern).Result()
	postures := make([]card.DevicePosture, 0, len(postureKeys))
	for _, key := range postureKeys {
		if val, err := r.client.Get(ctx, key).Result(); err == nil {
			var posture card.DevicePosture
			if err := json.Unmarshal([]byte(val), &posture); err == nil {
				postures = append(postures, posture)
			}
		}
	}
	if len(postures) > 0 {
		data.Postures = postures
	}

	// 如果没有任何数据，返回 nil（兼容性）
	if data.BedState == nil && data.RoomState == nil && data.ActiveAlarms == nil &&
		len(vitals) == 0 && len(postures) == 0 {
		return nil, nil
	}

	// 获取最新的时间戳（取各字段中最新的）
	timestamps := []int64{}
	if data.BedState != nil {
		timestamps = append(timestamps, int64(data.BedState.Timestamp))
	}
	if data.RoomState != nil {
		timestamps = append(timestamps, int64(data.RoomState.Timestamp))
	}
	if data.ActiveAlarms != nil {
		timestamps = append(timestamps, data.ActiveAlarms.Timestamp)
	}
	for _, vital := range vitals {
		timestamps = append(timestamps, vital.Timestamp)
	}
	for _, posture := range postures {
		timestamps = append(timestamps, posture.Timestamp)
	}

	if len(timestamps) > 0 {
		maxTs := timestamps[0]
		for _, ts := range timestamps {
			if ts > maxTs {
				maxTs = ts
			}
		}
		data.Timestamp = maxTs
	}

	return data, nil
}

// SetRealtimeData 设置卡片实时数据 (使用 hash key 模式，每个字段独立 TTL)
// - bed_state: TTL 12H
// - room_state: TTL 12H
// - active_alarms: TTL 12H
// - vital/{deviceID}: TTL 5s
// - posture/{deviceID}: TTL 5s
func (r *RedisCache) SetRealtimeData(ctx context.Context, cardID string, data *card.RealtimeData, ttl time.Duration) error {
	// 1. 设置 bed_state (使用传入的 TTL，通常为 12H)
	if data.BedState != nil {
		bedStateKey := fmt.Sprintf("vital-focus:card:%s:bed_state", cardID)
		jsonData, err := json.Marshal(data.BedState)
		if err != nil {
			return err
		}
		if err := r.client.Set(ctx, bedStateKey, string(jsonData), ttl).Err(); err != nil {
			return err
		}
	}

	// 2. 设置 room_state (使用传入的 TTL，通常为 12H)
	if data.RoomState != nil {
		roomStateKey := fmt.Sprintf("vital-focus:card:%s:room_state", cardID)
		jsonData, err := json.Marshal(data.RoomState)
		if err != nil {
			return err
		}
		if err := r.client.Set(ctx, roomStateKey, string(jsonData), ttl).Err(); err != nil {
			return err
		}
	}

	// 3. 设置 active_alarms (TTL 12H，来自 handleAlarm)
	if data.ActiveAlarms != nil {
		alarmsKey := fmt.Sprintf("vital-focus:card:%s:active_alarms", cardID)
		jsonData, err := json.Marshal(data.ActiveAlarms)
		if err != nil {
			return err
		}
		// 报警数据使用 12H TTL
		alarmTTL := 12 * time.Hour
		if ttl < alarmTTL {
			alarmTTL = ttl // 如果传入的 TTL 更短，就使用传入的
		}
		if err := r.client.Set(ctx, alarmsKey, string(jsonData), alarmTTL).Err(); err != nil {
			return err
		}
	}

	// 4. 设置 vital 数据 (TTL 5s)
	for _, vital := range data.Vital {
		vitalKey := fmt.Sprintf("vital-focus:card:%s:vital:%s", cardID, vital.DeviceID)
		jsonData, err := json.Marshal(vital)
		if err != nil {
			return err
		}
		if err := r.client.Set(ctx, vitalKey, string(jsonData), 5*time.Second).Err(); err != nil {
			return err
		}
	}

	// 5. 设置 posture 数据 (TTL 5s)
	for _, posture := range data.Postures {
		postureKey := fmt.Sprintf("vital-focus:card:%s:posture:%s", cardID, posture.DeviceID)
		jsonData, err := json.Marshal(posture)
		if err != nil {
			return err
		}
		if err := r.client.Set(ctx, postureKey, string(jsonData), 5*time.Second).Err(); err != nil {
			return err
		}
	}

	return nil
}

// GetDevicePosture 获取设备姿态数据
func (r *RedisCache) GetDevicePosture(ctx context.Context, cardID string, deviceID string) (*card.DevicePosture, error) {
	key := fmt.Sprintf("vital-focus:card:%s:posture:%s", cardID, deviceID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redislib.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var posture card.DevicePosture
	if err := json.Unmarshal([]byte(val), &posture); err != nil {
		return nil, err
	}

	return &posture, nil
}

// SetDevicePosture 设置设备姿态数据
func (r *RedisCache) SetDevicePosture(ctx context.Context, cardID string, posture *card.DevicePosture) error {
	key := fmt.Sprintf("vital-focus:card:%s:posture:%s", cardID, posture.DeviceID)
	jsonData, err := json.Marshal(posture)
	if err != nil {
		return err
	}

	// 姿态数据 TTL 为 5 秒
	return r.client.Set(ctx, key, string(jsonData), 5*time.Second).Err()
}

// GetVitalSimplified 获取设备生命体征
func (r *RedisCache) GetVitalSimplified(ctx context.Context, cardID string, deviceID string) (*card.VitalSimplified, error) {
	key := fmt.Sprintf("vital-focus:card:%s:vital:%s", cardID, deviceID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redislib.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var vital card.VitalSimplified
	if err := json.Unmarshal([]byte(val), &vital); err != nil {
		return nil, err
	}

	return &vital, nil
}

// SetVitalSimplified 设置设备生命体征
func (r *RedisCache) SetVitalSimplified(ctx context.Context, cardID string, vital *card.VitalSimplified) error {
	key := fmt.Sprintf("vital-focus:card:%s:vital:%s", cardID, vital.DeviceID)
	jsonData, err := json.Marshal(vital)
	if err != nil {
		return err
	}

	// 生命体征数据 TTL 为 5 秒
	return r.client.Set(ctx, key, string(jsonData), 5*time.Second).Err()
}

// DeleteVitalSimplified 删除生命体征缓存
func (r *RedisCache) DeleteVitalSimplified(ctx context.Context, cardID string, deviceID string) error {
	key := fmt.Sprintf("vital-focus:card:%s:vital:%s", cardID, deviceID)
	return r.client.Del(ctx, key).Err()
}

// DeleteDevicePosture 删除姿态缓存
func (r *RedisCache) DeleteDevicePosture(ctx context.Context, cardID string, deviceID string) error {
	key := fmt.Sprintf("vital-focus:card:%s:posture:%s", cardID, deviceID)
	return r.client.Del(ctx, key).Err()
}

// GetAllCardIds 获取所有卡片 ID (扫描所有 active_alarms key)
func (r *RedisCache) GetAllCardIds(ctx context.Context) ([]string, error) {
	// 查找所有符合模式 "vital-focus:card:*:active_alarms" 的键
	pattern := "vital-focus:card:*:active_alarms"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	// 从键中提取 cardID
	cardIDs := make([]string, 0, len(keys))
	seen := make(map[string]bool)
	for _, key := range keys {
		// 键的格式为 "vital-focus:card:{cardID}:active_alarms"
		cardID := extractCardIDFromHashKey(key, "active_alarms")
		if cardID != "" && !seen[cardID] {
			cardIDs = append(cardIDs, cardID)
			seen[cardID] = true
		}
	}

	return cardIDs, nil
}

// extractCardIDFromKey 从 realtime key 中提取 cardID (已废弃，保留兼容性)
func extractCardIDFromKey(key string) string {
	// 键的格式为 "vital-focus:card:{cardID}:realtime"
	const prefix = "vital-focus:card:"
	const suffix = ":realtime"

	if len(key) <= len(prefix)+len(suffix) {
		return ""
	}

	if !startsWith(key, prefix) || !endsWith(key, suffix) {
		return ""
	}

	// 提取中间部分
	cardID := key[len(prefix) : len(key)-len(suffix)]
	return cardID
}

// extractCardIDFromHashKey 从 hash key 中提取 cardID
// 例如："vital-focus:card:{cardID}:active_alarms" 或 "vital-focus:card:{cardID}:bed_state"
func extractCardIDFromHashKey(key string, fieldName string) string {
	const prefix = "vital-focus:card:"
	if !startsWith(key, prefix) {
		return ""
	}

	// 移除前缀
	remaining := key[len(prefix):]

	// 查找字段后缀
	suffix := ":" + fieldName
	if !endsWith(remaining, suffix) {
		return ""
	}

	// 提取 cardID
	cardID := remaining[:len(remaining)-len(suffix)]
	return cardID
}

// startsWith 检查字符串是否以指定前缀开头
func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

// endsWith 检查字符串是否以指定后缀结尾
func endsWith(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
