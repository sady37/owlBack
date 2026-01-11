package http

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/models"
	"wisefido-radar/internal/publisher"
	"wisefido-radar/internal/repository"
	
	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// CommandService 命令服务
// 封装 Radar 设备命令发送逻辑，实现请求-响应机制
type CommandService struct {
	config       *config.Config
	publisher    *publisher.MQTTPublisher
	deviceRepo   *repository.DeviceRepository
	redisClient  *redis.Client
	logger       *zap.Logger
	responseChan chan *models.PropertyResponse // 属性响应通道（临时，后续改为 Redis）
}

// NewCommandService 创建命令服务
func NewCommandService(
	cfg *config.Config,
	pub *publisher.MQTTPublisher,
	deviceRepo *repository.DeviceRepository,
	redisClient *redis.Client,
	logger *zap.Logger,
) *CommandService {
	return &CommandService{
		config:       cfg,
		publisher:    pub,
		deviceRepo:   deviceRepo,
		redisClient:  redisClient,
		logger:       logger,
		responseChan: make(chan *models.PropertyResponse, 100),
	}
}

// GetDeviceProperties 读取设备属性
// 参考协议文档 3.4 节
// 如果 keys 为空，读取所有属性
func (s *CommandService) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	// 1. 验证设备是否存在
	device, err := s.deviceRepo.GetDeviceByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	
	// 2. 生成请求 ID
	requestID := s.generateRequestID()
	
	// 3. 发送命令
	if err := s.publisher.PublishPropertyReadCommand(ctx, uid, keys, requestID); err != nil {
		return nil, fmt.Errorf("failed to publish property read command: %w", err)
	}
	
	// 4. 等待响应（使用 Redis 存储响应）
	// TODO: 实现基于 Redis 的请求-响应机制
	// 当前实现：简单超时等待（临时方案）
	response, err := s.waitForPropertyResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get property response: %w", err)
	}
	
	if response.Code != 200 {
		return nil, fmt.Errorf("device returned error code: %d", response.Code)
	}
	
	s.logger.Info("Property read command completed",
		zap.String("device_id", device.DeviceID),
		zap.String("uid", uid),
		zap.String("request_id", requestID),
		zap.Int("properties_count", len(response.Data)),
	)
	
	return response.Data, nil
}

// SetDeviceProperties 设置设备属性
// 参考协议文档 3.4 节
func (s *CommandService) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) error {
	// 1. 验证设备是否存在
	device, err := s.deviceRepo.GetDeviceByUID(uid)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}
	
	// 2. 生成请求 ID
	requestID := s.generateRequestID()
	
	// 3. 发送命令
	if err := s.publisher.PublishPropertyUpdateCommand(ctx, uid, properties, requestID); err != nil {
		return fmt.Errorf("failed to publish property update command: %w", err)
	}
	
	// 4. 等待响应
	response, err := s.waitForPropertyResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get property response: %w", err)
	}
	
	if response.Code != 200 {
		return fmt.Errorf("device returned error code: %d", response.Code)
	}
	
	s.logger.Info("Property update command completed",
		zap.String("device_id", device.DeviceID),
		zap.String("uid", uid),
		zap.String("request_id", requestID),
		zap.Int("properties_count", len(properties)),
	)
	
	return nil
}

// SubscribeRealtimeData 订阅实时数据
// 参考协议文档 3.7.1 节
// content: 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
// duration: 订阅时长（秒），最大 3600 秒
func (s *CommandService) SubscribeRealtimeData(ctx context.Context, uid string, content interface{}, duration int) error {
	// 1. 验证设备是否存在
	device, err := s.deviceRepo.GetDeviceByUID(uid)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}
	
	// 2. 验证参数
	if duration < 1 || duration > 3600 {
		return fmt.Errorf("invalid duration: %d (must be between 1 and 3600)", duration)
	}
	
	// 3. 发送订阅命令
	if err := s.publisher.PublishMonitorSubscriptionCommand(ctx, uid, content, duration); err != nil {
		return fmt.Errorf("failed to publish monitor subscription command: %w", err)
	}
	
	// 4. 记录订阅状态到 Redis（供订阅管理器使用）
	// 注意：这里不直接依赖 SubscriptionManager，避免循环依赖
	// 订阅管理器会从 Redis 读取状态
	if err := s.saveSubscriptionInfo(uid, content, duration); err != nil {
		s.logger.Warn("Failed to save subscription info to Redis",
			zap.String("uid", uid),
			zap.Error(err),
		)
		// 不返回错误，订阅已发送
	}
	
	s.logger.Info("Monitor subscription command sent",
		zap.String("device_id", device.DeviceID),
		zap.String("uid", uid),
		zap.Any("content", content),
		zap.Int("duration", duration),
	)
	
	return nil
}

// CallDeviceFunction 调用设备功能（重启等）
// 参考协议文档 3.8.1 节
// dev: 0-重启雷达和主控，1-只重启雷达，2-只重启主控，100-清除设备数据，101-清除雷达数据，102-清除主控数据
func (s *CommandService) CallDeviceFunction(ctx context.Context, uid string, dev int) error {
	// 1. 验证设备是否存在
	device, err := s.deviceRepo.GetDeviceByUID(uid)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}
	
	// 2. 验证参数
	validDevValues := map[int]bool{
		0:   true, // 重启雷达和主控
		1:   true, // 只重启雷达
		2:   true, // 只重启主控
		100: true, // 清除设备数据
		101: true, // 清除雷达数据
		102: true, // 清除主控数据
	}
	if !validDevValues[dev] {
		return fmt.Errorf("invalid dev value: %d", dev)
	}
	
	// 3. 生成请求 ID
	requestID := s.generateRequestID()
	
	// 4. 发送命令
	if err := s.publisher.PublishFunctionControlCommand(ctx, uid, dev, requestID); err != nil {
		return fmt.Errorf("failed to publish function control command: %w", err)
	}
	
	// 5. 等待响应
	response, err := s.waitForFunctionResponse(ctx, requestID, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get function response: %w", err)
	}
	
	if response.Code != 200 {
		return fmt.Errorf("device returned error code: %d", response.Code)
	}
	
	s.logger.Info("Function control command completed",
		zap.String("device_id", device.DeviceID),
		zap.String("uid", uid),
		zap.String("request_id", requestID),
		zap.Int("dev", dev),
	)
	
	return nil
}

// generateRequestID 生成请求 ID
func (s *CommandService) generateRequestID() string {
	return uuid.New().String()
}

// waitForPropertyResponse 等待属性响应
// TODO: 实现基于 Redis 的请求-响应机制
// 当前实现：简单超时等待（临时方案）
func (s *CommandService) waitForPropertyResponse(ctx context.Context, requestID string, timeout time.Duration) (*models.PropertyResponse, error) {
	// 临时实现：从 Redis 中轮询响应
	// 响应应该由 MQTT Consumer 接收到后存储到 Redis
	key := fmt.Sprintf("radar:response:%s", requestID)
	
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// 从 Redis 获取响应
			val, err := s.redisClient.Get(ctx, key).Result()
			if err == redis.Nil {
				// 响应尚未到达，继续等待
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("failed to get response from Redis: %w", err)
			}
			
			// 解析响应
			var response models.PropertyResponse
			if err := json.Unmarshal([]byte(val), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}
			
			// 删除 Redis 中的响应（避免重复使用）
			_ = s.redisClient.Del(ctx, key)
			
			return &response, nil
		}
	}
	
	return nil, fmt.Errorf("timeout waiting for property response: %s", requestID)
}

// waitForFunctionResponse 等待功能调用响应
func (s *CommandService) waitForFunctionResponse(ctx context.Context, requestID string, timeout time.Duration) (*models.FunctionControlResponse, error) {
	key := fmt.Sprintf("radar:response:func:%s", requestID)
	
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			val, err := s.redisClient.Get(ctx, key).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("failed to get response from Redis: %w", err)
			}
			
			var response models.FunctionControlResponse
			if err := json.Unmarshal([]byte(val), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}
			
			_ = s.redisClient.Del(ctx, key)
			
			return &response, nil
		}
	}
	
	return nil, fmt.Errorf("timeout waiting for function response: %s", requestID)
}

// saveSubscriptionInfo 保存订阅信息到 Redis
// 供订阅管理器使用，避免循环依赖
// 使用与 SubscriptionManager 相同的格式，确保兼容性
func (s *CommandService) saveSubscriptionInfo(uid string, content interface{}, duration int) error {
	now := time.Now()
	info := map[string]interface{}{
		"uid":          uid,
		"content":      content,
		"duration":     duration,
		"subscribed_at": now.Format(time.RFC3339),
		"expires_at":   now.Add(time.Duration(duration) * time.Second).Format(time.RFC3339),
	}
	
	key := fmt.Sprintf("radar:subscription:%s", uid)
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription info: %w", err)
	}
	
	// 保存到 Redis，TTL 设置为订阅时长 + 10分钟（作为缓冲）
	ttl := time.Duration(duration)*time.Second + 10*time.Minute
	if err := s.redisClient.Set(context.Background(), key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save to Redis: %w", err)
	}
	
	return nil
}
