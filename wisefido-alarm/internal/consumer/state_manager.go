package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-alarm/internal/config"

	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
)

// StateManager 报警状态管理器（用于管理事件1-4的状态）
type StateManager struct {
	config      *config.Config
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewStateManager 创建状态管理器
func NewStateManager(
	cfg *config.Config,
	redisClient *redis.Client,
	logger *zap.Logger,
) *StateManager {
	return &StateManager{
		config:      cfg,
		redisClient: redisClient,
		logger:      logger,
	}
}

// GetStateKey 构建状态键
func (s *StateManager) GetStateKey(cardID, trackID, stateType string) string {
	return fmt.Sprintf("%s%s:track_%s:%s",
		s.config.Alarm.Cache.StateKeyPrefix,
		cardID,
		trackID,
		stateType,
	)
}

// SetState 设置状态（带 TTL）
func (s *StateManager) SetState(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 序列化值
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// 写入 Redis
	err = s.redisClient.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	return nil
}

// GetState 获取状态
func (s *StateManager) GetState(ctx context.Context, key string, dest interface{}) error {
	// 从 Redis 读取
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("state not found: %s", key)
		}
		return fmt.Errorf("failed to get state: %w", err)
	}

	// 反序列化
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return nil
}

// DeleteState 删除状态
func (s *StateManager) DeleteState(ctx context.Context, key string) error {
	err := s.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}
	return nil
}

// ExistsState 检查状态是否存在
func (s *StateManager) ExistsState(ctx context.Context, key string) (bool, error) {
	count, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check state existence: %w", err)
	}
	return count > 0, nil
}

// SetStateTTL 设置状态的 TTL
func (s *StateManager) SetStateTTL(ctx context.Context, key string, ttl time.Duration) error {
	err := s.redisClient.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set state TTL: %w", err)
	}
	return nil
}

// Event1State 事件1的状态数据
type Event1State struct {
	// 阶段1：lying基线
	LyingHeight   *float64 `json:"lying_height,omitempty"`   // lying时的高度值
	LyingPosition *struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"lying_position,omitempty"` // lying时的位置
	LyingTime *int64 `json:"lying_time,omitempty"` // lying开始时间

	// 阶段2：跌落检测
	LeftBedTime *int64 `json:"left_bed_time,omitempty"` // T0：离床时间
	TrackID     string `json:"track_id"`                // 跟踪的 track_id
}

// Event2State 事件2的状态数据
type Event2State struct {
	LastHRTime    *int64 `json:"last_hr_time,omitempty"`    // 最后一次HR时间
	LastRRTime    *int64 `json:"last_rr_time,omitempty"`    // 最后一次RR时间
	RadarDetected bool   `json:"radar_detected"`            // 雷达是否检测到人
}

// Event3State 事件3的状态数据
type Event3State struct {
	TrackID       string    `json:"track_id"`                // 跟踪的 track_id
	StandingTime  *int64    `json:"standing_time,omitempty"` // 开始站立的时间
	LastPosition  *struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"last_position,omitempty"` // 最后位置
	PositionChange float64 `json:"position_change"` // 位置变化（cm）
}

// Event4State 事件4的状态数据
type Event4State struct {
	TrackID         string   `json:"track_id"`                  // 跟踪的 track_id
	LastHeight      *float64 `json:"last_height,omitempty"`    // 消失前的高度
	LastPosition    *struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"last_position,omitempty"` // 消失前的位置
	DisappearTime   *int64 `json:"disappear_time,omitempty"`   // 消失时间
	NoActivitySince *int64 `json:"no_activity_since,omitempty"` // 无活动时间
}

