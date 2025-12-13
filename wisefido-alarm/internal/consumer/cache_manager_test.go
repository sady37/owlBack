package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"wisefido-alarm/internal/config"
	"wisefido-alarm/internal/models"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client, *CacheManager) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cfg := &config.Config{}
	cfg.Alarm.Cache.RealtimeKeyPrefix = "vital-focus:card:"
	cfg.Alarm.Cache.RealtimeSuffix = ":realtime"
	cfg.Alarm.Cache.AlarmKeyPrefix = "vital-focus:card:"
	cfg.Alarm.Cache.AlarmSuffix = ":alarms"
	cfg.Alarm.Cache.AlarmTTL = 30

	logger := zap.NewNop()
	cacheManager := NewCacheManager(cfg, redisClient, logger)

	return mr, redisClient, cacheManager
}

func TestCacheManager_GetRealtimeData_Success(t *testing.T) {
	_, _, cacheManager := setupTestRedis(t)

	cardID := "card-123"
	realtimeData := &models.RealtimeData{
		Heart:        intPtr(72),
		Breath:       intPtr(18),
		HeartSource:  "Sleepace",
		BreathSource: "Sleepace",
		PersonCount:  1,
		Timestamp:    time.Now().Unix(),
	}

	// 先写入数据
	key := "vital-focus:card:" + cardID + ":realtime"
	jsonData, err := json.Marshal(realtimeData)
	require.NoError(t, err)

	ctx := context.Background()
	err = cacheManager.redisClient.Set(ctx, key, jsonData, time.Minute).Err()
	require.NoError(t, err)

	// 读取数据
	data, err := cacheManager.GetRealtimeData(cardID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, intPtr(72), data.Heart)
	assert.Equal(t, intPtr(18), data.Breath)
	assert.Equal(t, "Sleepace", data.HeartSource)
	assert.Equal(t, 1, data.PersonCount)
}

func TestCacheManager_GetRealtimeData_NotFound(t *testing.T) {
	_, _, cacheManager := setupTestRedis(t)

	cardID := "card-not-exist"

	_, err := cacheManager.GetRealtimeData(cardID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "realtime data not found")
}

func TestCacheManager_UpdateAlarmCache_Success(t *testing.T) {
	_, _, cacheManager := setupTestRedis(t)

	cardID := "card-123"
	alarms := []models.AlarmEvent{
		{
			EventID:     "event-1",
			EventType:   "Fall",
			AlarmLevel:  "ALERT",
			AlarmStatus: "active",
		},
		{
			EventID:     "event-2",
			EventType:   "AbnormalHeartRate",
			AlarmLevel:  "WARNING",
			AlarmStatus: "active",
		},
	}

	err := cacheManager.UpdateAlarmCache(cardID, alarms)

	require.NoError(t, err)

	// 验证数据已写入
	key := "vital-focus:card:" + cardID + ":alarms"
	ctx := context.Background()
	val, err := cacheManager.redisClient.Get(ctx, key).Result()
	require.NoError(t, err)

	var cachedAlarms []models.AlarmEvent
	err = json.Unmarshal([]byte(val), &cachedAlarms)
	require.NoError(t, err)
	assert.Len(t, cachedAlarms, 2)
	assert.Equal(t, "event-1", cachedAlarms[0].EventID)
}

func TestStateManager_SetState_GetState(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cfg := &config.Config{}
	cfg.Alarm.Cache.StateKeyPrefix = "alarm:state:"

	logger := zap.NewNop()
	stateManager := NewStateManager(cfg, redisClient, logger)

	ctx := context.Background()
	key := stateManager.GetStateKey("card-123", "track-456", "event1")

	state := &Event1State{
		TrackID:     "track-456",
		LeftBedTime: int64Ptr(time.Now().Unix()),
	}

	err := stateManager.SetState(ctx, key, state, time.Minute)
	require.NoError(t, err)

	// 读取状态
	var retrievedState Event1State
	err = stateManager.GetState(ctx, key, &retrievedState)
	require.NoError(t, err)
	assert.Equal(t, "track-456", retrievedState.TrackID)
	assert.NotNil(t, retrievedState.LeftBedTime)
}

func TestStateManager_ExistsState(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cfg := &config.Config{}
	cfg.Alarm.Cache.StateKeyPrefix = "alarm:state:"

	logger := zap.NewNop()
	stateManager := NewStateManager(cfg, redisClient, logger)

	ctx := context.Background()
	key := "test-key"

	// 状态不存在
	exists, err := stateManager.ExistsState(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// 设置状态
	err = stateManager.SetState(ctx, key, map[string]string{"test": "value"}, time.Minute)
	require.NoError(t, err)

	// 状态存在
	exists, err = stateManager.ExistsState(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
}

// 辅助函数
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
