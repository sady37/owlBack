package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/consumer"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
)

// RadarService 雷达服务
type RadarService struct {
	config          *config.Config
	mqttClient      *mqtt.Client
	redisClient     *redis.Client
	deviceRepo      repository.DeviceRepository
	streamPublisher *consumer.StreamPublisher
	mqttConsumer    *consumer.MQTTConsumer
}

// NewRadarService 创建雷达服务
func NewRadarService(
	cfg *config.Config,
	mqttClient *mqtt.Client,
	redisClient *redis.Client,
	deviceRepo repository.DeviceRepository,
	streamPublisher *consumer.StreamPublisher,
	mqttConsumer *consumer.MQTTConsumer,
) (*RadarService, error) {
	return &RadarService{
		config:          cfg,
		mqttClient:      mqttClient,
		redisClient:     redisClient,
		deviceRepo:      deviceRepo,
		streamPublisher: streamPublisher,
		mqttConsumer:    mqttConsumer,
	}, nil
}

// Start 启动服务
func (s *RadarService) Start(ctx context.Context) error {
	log.Println("Starting radar service...")

	// 启动MQTT消费者
	if err := s.mqttConsumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MQTT consumer: %w", err)
	}

	log.Println("Radar service started successfully")
	return nil
}

// Stop 停止服务
func (s *RadarService) Stop(ctx context.Context) error {
	log.Println("Stopping radar service...")

	// 停止MQTT消费者
	if err := s.mqttConsumer.Stop(ctx); err != nil {
		log.Printf("Error stopping MQTT consumer: %v", err)
	}

	log.Println("Radar service stopped")
	return nil
}

// GetDeviceProperties 读取设备属性
func (s *RadarService) GetDeviceProperties(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error) {
	// 生成请求ID
	requestID := fmt.Sprintf("prop_%s_%d", deviceUID, time.Now().Unix())

	// 构建属性读取命令
	command := map[string]interface{}{
		"requestId": requestID,
		"method":    "thing.property.get",
		"version":   "1.0",
	}

	if len(keys) > 0 {
		command["data"] = map[string]interface{}{
			"keys": keys,
		}
	}

	// 发送命令
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		return nil, fmt.Errorf("failed to publish command: %w", err)
	}

	log.Printf("Property read command sent: %s, requestId: %s", deviceUID, requestID)

	// 等待响应
	response, err := s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}

	// 提取属性数据
	if data, ok := response["data"].(map[string]interface{}); ok {
		return data, nil
	}

	return nil, fmt.Errorf("no property data in response")
}

// SetDeviceProperties 设置设备属性
func (s *RadarService) SetDeviceProperties(ctx context.Context, deviceUID string, properties map[string]interface{}) error {
	// 生成请求ID
	requestID := fmt.Sprintf("prop_set_%s_%d", deviceUID, time.Now().Unix())

	// 构建属性设置命令
	command := map[string]interface{}{
		"requestId": requestID,
		"method":    "thing.property.set",
		"version":   "1.0",
		"data":      properties,
	}

	// 发送命令
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		return fmt.Errorf("failed to publish command: %w", err)
	}

	log.Printf("Property set command sent: %s, requestId: %s", deviceUID, requestID)

	// 等待响应
	_, err = s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	return nil
}

// SubscribeRealtimeData 订阅实时数据
func (s *RadarService) SubscribeRealtimeData(ctx context.Context, deviceUID string, content interface{}, duration int) error {
	// 生成请求ID
	requestID := fmt.Sprintf("monitor_%s_%d", deviceUID, time.Now().Unix())

	// 构建订阅命令
	command := map[string]interface{}{
		"requestId": requestID,
		"method":    "thing.function.invoke",
		"version":   "1.0",
		"data": map[string]interface{}{
			"identifier": "monitor",
			"inputData": map[string]interface{}{
				"content":  content,
				"duration": duration,
			},
		},
	}

	// 发送命令
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	topic := s.mqttClient.BuildCommandTopic("func", deviceUID)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		return fmt.Errorf("failed to publish command: %w", err)
	}

	log.Printf("Realtime data subscription command sent: %s, requestId: %s, duration: %d", deviceUID, requestID, duration)

	// 等待响应
	_, err = s.waitForResponse(ctx, "func:"+requestID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	return nil
}

// CallDeviceFunction 调用设备功能
func (s *RadarService) CallDeviceFunction(ctx context.Context, deviceUID string, dev int) error {
	// 生成请求ID
	requestID := fmt.Sprintf("func_%s_%d", deviceUID, time.Now().Unix())

	// 构建功能调用命令
	command := map[string]interface{}{
		"requestId": requestID,
		"method":    "thing.function.invoke",
		"version":   "1.0",
		"data": map[string]interface{}{
			"identifier": "reboot",
			"inputData": map[string]interface{}{
				"dev": dev,
			},
		},
	}

	// 发送命令
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	topic := s.mqttClient.BuildCommandTopic("func", deviceUID)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		return fmt.Errorf("failed to publish command: %w", err)
	}

	log.Printf("Device function command sent: %s, requestId: %s, dev: %d", deviceUID, requestID, dev)

	// 等待响应
	_, err = s.waitForResponse(ctx, "func:"+requestID, 30*time.Second) // 重启需要更长时间
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	return nil
}

// GetDeviceInfo 获取设备信息
func (s *RadarService) GetDeviceInfo(ctx context.Context, deviceUID string) (*domain.Device, error) {
	device, err := s.deviceRepo.GetDeviceByUID(ctx, deviceUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	return device, nil
}

// GetDeviceLocationInfo 获取设备位置信息
func (s *RadarService) GetDeviceLocationInfo(ctx context.Context, deviceUID string) (*domain.DeviceLocationInfo, error) {
	locationInfo, err := s.deviceRepo.GetDeviceLocationInfo(ctx, deviceUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device location info: %w", err)
	}

	return locationInfo, nil
}

// GetDevicesByTenant 根据租户获取设备列表
func (s *RadarService) GetDevicesByTenant(ctx context.Context, tenantID string) ([]*domain.Device, error) {
	devices, err := s.deviceRepo.GetDevicesByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices by tenant: %w", err)
	}

	return devices, nil
}

// waitForResponse 等待命令响应
func (s *RadarService) waitForResponse(ctx context.Context, requestID string, timeout time.Duration) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 轮询Redis获取响应
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for response: %s", requestID)
		case <-ticker.C:
			response, err := s.streamPublisher.GetCommandResponse(ctx, requestID)
			if err == nil {
				return response, nil
			}
			// 如果错误是"response not found"，继续等待
			if err.Error() != fmt.Sprintf("response not found: %s", requestID) {
				return nil, err
			}
		}
	}
}
