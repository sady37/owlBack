package exports

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/http"
	"wisefido-radar/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// SubscriptionManager 订阅管理器
// 负责管理雷达设备的实时数据订阅，包括自动订阅和自动续订
type SubscriptionManager struct {
	config             *config.Config
	commandService     *http.CommandService
	deviceRepo         *repository.DeviceRepository
	redisClient        *redis.Client
	logger             *zap.Logger
	subscriptionCtx    context.Context
	subscriptionCancel context.CancelFunc
}

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	UID          string      `json:"uid"`
	Content      interface{} `json:"content"`       // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
	Duration     int         `json:"duration"`      // 订阅时长（秒）
	SubscribedAt time.Time   `json:"subscribed_at"` // 订阅时间
	ExpiresAt    time.Time   `json:"expires_at"`    // 过期时间
}

// NewSubscriptionManager 创建订阅管理器
func NewSubscriptionManager(
	cfg *config.Config,
	commandService *http.CommandService,
	deviceRepo *repository.DeviceRepository,
	redisClient *redis.Client,
	logger *zap.Logger,
) *SubscriptionManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SubscriptionManager{
		config:             cfg,
		commandService:     commandService,
		deviceRepo:         deviceRepo,
		redisClient:        redisClient,
		logger:             logger,
		subscriptionCtx:    ctx,
		subscriptionCancel: cancel,
	}
}

// Start 启动订阅管理器
// 启动后台 goroutine 定期检查并续订即将过期的订阅
func (m *SubscriptionManager) Start(ctx context.Context) {
	// 合并上下文，当任一上下文取消时停止
	mergedCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-m.subscriptionCtx.Done():
			cancel()
		case <-ctx.Done():
			cancel()
		}
	}()

	// 使用配置的续订间隔（默认50分钟）
	// 订阅时长为3600秒（1小时），在过期前续订，确保数据不中断
	renewalInterval := time.Duration(m.config.Radar.Subscription.RenewalInterval) * time.Minute
	ticker := time.NewTicker(renewalInterval)
	defer ticker.Stop()

	m.logger.Info("Subscription manager started",
		zap.Int("renewal_interval_minutes", m.config.Radar.Subscription.RenewalInterval),
		zap.Int("subscription_duration", m.config.Radar.Subscription.DefaultDuration),
		zap.Bool("auto_subscribe", m.config.Radar.Subscription.AutoSubscribe),
	)

	// 立即执行一次检查（服务启动时）
	m.checkAndRenewSubscriptions(mergedCtx)

	for {
		select {
		case <-mergedCtx.Done():
			m.logger.Info("Subscription manager stopped")
			return
		case <-ticker.C:
			m.checkAndRenewSubscriptions(mergedCtx)
		}
	}
}

// Stop 停止订阅管理器
func (m *SubscriptionManager) Stop() {
	if m.subscriptionCancel != nil {
		m.subscriptionCancel()
	}
}

// AutoSubscribe 自动订阅设备实时数据
// 在设备首次连接时调用
func (m *SubscriptionManager) AutoSubscribe(ctx context.Context, uid string) error {
	// 检查是否已经订阅
	if m.isSubscribed(uid) {
		m.logger.Debug("Device already subscribed, skipping auto-subscribe",
			zap.String("uid", uid),
		)
		return nil
	}

	// 使用配置的默认值订阅
	content := m.config.Radar.Subscription.DefaultContent   // 默认 0（同时订阅）
	duration := m.config.Radar.Subscription.DefaultDuration // 默认 3600（1小时）

	// 调用订阅服务
	if err := m.commandService.SubscribeRealtimeData(ctx, uid, content, duration); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// 记录订阅状态到 Redis
	if err := m.saveSubscriptionInfo(uid, content, duration); err != nil {
		m.logger.Warn("Failed to save subscription info to Redis",
			zap.String("uid", uid),
			zap.Error(err),
		)
		// 不返回错误，订阅已发送
	}

	m.logger.Info("Auto-subscribed to real-time data",
		zap.String("uid", uid),
		zap.Int("duration", duration),
	)

	return nil
}

// checkAndRenewSubscriptions 检查并续订即将过期的订阅
func (m *SubscriptionManager) checkAndRenewSubscriptions(ctx context.Context) {
	m.logger.Debug("Checking subscriptions for renewal")

	// 从 Redis 获取所有活跃订阅
	subscriptions, err := m.getAllActiveSubscriptions(ctx)
	if err != nil {
		m.logger.Error("Failed to get active subscriptions",
			zap.Error(err),
		)
		return
	}

	now := time.Now()
	renewCount := 0

	for uid, info := range subscriptions {
		// 在过期前续订（使用配置的提前时间，默认10分钟）
		advanceTime := time.Duration(m.config.Radar.Subscription.RenewalAdvanceTime) * time.Minute
		renewTime := info.ExpiresAt.Add(-advanceTime)

		if now.After(renewTime) {
			// 需要续订
			if err := m.renewSubscription(ctx, uid, info); err != nil {
				m.logger.Error("Failed to renew subscription",
					zap.String("uid", uid),
					zap.Error(err),
				)
				// 继续处理其他订阅
				continue
			}
			renewCount++
		}
	}

	if renewCount > 0 {
		m.logger.Info("Renewed subscriptions",
			zap.Int("count", renewCount),
		)
	}
}

// renewSubscription 续订订阅
func (m *SubscriptionManager) renewSubscription(ctx context.Context, uid string, info *SubscriptionInfo) error {
	m.logger.Info("Renewing subscription",
		zap.String("uid", uid),
		zap.Time("expires_at", info.ExpiresAt),
	)

	// 调用订阅服务
	if err := m.commandService.SubscribeRealtimeData(ctx, uid, info.Content, info.Duration); err != nil {
		return fmt.Errorf("failed to renew subscription: %w", err)
	}

	// 更新订阅状态到 Redis
	if err := m.saveSubscriptionInfo(uid, info.Content, info.Duration); err != nil {
		return fmt.Errorf("failed to update subscription info: %w", err)
	}

	m.logger.Info("Successfully renewed subscription",
		zap.String("uid", uid),
	)

	return nil
}

// saveSubscriptionInfo 保存订阅信息到 Redis
func (m *SubscriptionManager) saveSubscriptionInfo(uid string, content interface{}, duration int) error {
	now := time.Now()
	info := &SubscriptionInfo{
		UID:          uid,
		Content:      content,
		Duration:     duration,
		SubscribedAt: now,
		ExpiresAt:    now.Add(time.Duration(duration) * time.Second),
	}

	key := m.getSubscriptionKey(uid)
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription info: %w", err)
	}

	// 保存到 Redis，TTL 设置为订阅时长 + 10分钟（作为缓冲）
	ttl := time.Duration(duration)*time.Second + 10*time.Minute
	if err := m.redisClient.Set(context.Background(), key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save to Redis: %w", err)
	}

	return nil
}

// isSubscribed 检查设备是否已订阅
func (m *SubscriptionManager) isSubscribed(uid string) bool {
	key := m.getSubscriptionKey(uid)
	val, err := m.redisClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		m.logger.Warn("Failed to check subscription status",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return false
	}

	// 检查是否过期
	var info SubscriptionInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return false
	}

	return time.Now().Before(info.ExpiresAt)
}

// getAllActiveSubscriptions 获取所有活跃订阅
func (m *SubscriptionManager) getAllActiveSubscriptions(ctx context.Context) (map[string]*SubscriptionInfo, error) {
	// 使用 Redis SCAN 查找所有订阅 key
	pattern := "radar:subscription:*"
	subscriptions := make(map[string]*SubscriptionInfo)

	iter := m.redisClient.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		val, err := m.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			m.logger.Warn("Failed to get subscription info",
				zap.String("key", key),
				zap.Error(err),
			)
			continue
		}

		var info SubscriptionInfo
		if err := json.Unmarshal([]byte(val), &info); err != nil {
			m.logger.Warn("Failed to unmarshal subscription info",
				zap.String("key", key),
				zap.Error(err),
			)
			continue
		}

		// 只返回未过期的订阅
		if time.Now().Before(info.ExpiresAt) {
			subscriptions[info.UID] = &info
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan subscriptions: %w", err)
	}

	return subscriptions, nil
}

// getSubscriptionKey 获取订阅信息的 Redis key
func (m *SubscriptionManager) getSubscriptionKey(uid string) string {
	return fmt.Sprintf("radar:subscription:%s", uid)
}
