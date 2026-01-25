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
// 协议：/prefix/prop/productId/UID/get
// 当 len(keys)>1 时：分笔依次 MQTT 读每个 key，每个指令间额外 50ms，再合并返回（设备分笔读取，间隔由本服务控制）
func (s *RadarService) GetDeviceProperties(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error) {
	if len(keys) <= 1 {
		return s.readOneBatch(ctx, deviceUID, keys)
	}
	merged := make(map[string]interface{})
	for i, k := range keys {
		batch, err := s.readOneBatch(ctx, deviceUID, []string{k})
		if err != nil {
			return nil, fmt.Errorf("read key %q: %w", k, err)
		}
		for kk, v := range batch {
			merged[kk] = v
		}
		if i < len(keys)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return merged, nil
}

// readOneBatch 发一笔 MQTT read（keys 可为空表示读全部），等响应后返回 data
func (s *RadarService) readOneBatch(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error) {
	uidLast4 := getLast4Chars(deviceUID)
	requestID := fmt.Sprintf("G%s.%d", uidLast4, time.Now().Unix())
	command := map[string]interface{}{"cmd": "read", "requestId": requestID}
	if len(keys) > 0 {
		command["data"] = map[string]interface{}{"key": keys}
	}
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		return nil, fmt.Errorf("failed to publish command: %w", err)
	}
	log.Printf("Property read command sent: %s, requestId: %s, keys: %v", deviceUID, requestID, keys)
	response, err := s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	if data, ok := response["data"].(map[string]interface{}); ok {
		return data, nil
	}
	return nil, fmt.Errorf("no property data in response")
}

// SetDeviceProperties 设置设备属性
// 协议：/prefix/prop/productId/UID/get
func (s *RadarService) SetDeviceProperties(ctx context.Context, deviceUID string, properties map[string]interface{}) error {
	// 生成请求ID：S + uid最后4位 + . + timestamp
	uidLast4 := getLast4Chars(deviceUID)
	timestamp := time.Now().Unix()
	requestID := fmt.Sprintf("S%s.%d", uidLast4, timestamp)

	log.Printf("🔧 SetDeviceProperties: preparing command for device %s, requestId: %s", deviceUID, requestID)
	log.Printf("   Properties: %+v", properties)

	// 检查是否是模式修改（radar_func_ctrl）
	if mode, ok := properties["radar_func_ctrl"]; ok {
		log.Printf("🔄 MODE CHANGE: Setting radar_func_ctrl = %v for device %s", mode, deviceUID)
		modeDesc := map[interface{}]string{
			3:  "轨迹模式",
			7:  "呼吸心率模式",
			11: "轨迹+呼吸心率模式",
			15: "轨迹+呼吸心率+跌倒模式",
		}
		if desc, exists := modeDesc[mode]; exists {
			log.Printf("   Mode Description: %s", desc)
		}
	}

	// 将所有属性值转换为字符串（厂家要求所有值都使用字符串格式）
	stringProperties := convertPropertiesToStrings(properties)

	// 构建属性设置命令（根据协议文档 3.4 节）
	// 协议格式：/prop/productId/UID/get
	// {
	//   "cmd": "update",
	//   "requestId": "sadibaiubd123",
	//   "data": {
	//     "key1": "value1",
	//     "key2": "value2"
	//   }
	// }
	command := map[string]interface{}{
		"cmd":       "update",
		"requestId": requestID,
		"data":      stringProperties,
	}

	// 发送命令
	commandJSON, err := json.Marshal(command)
	if err != nil {
		log.Printf("❌ SetDeviceProperties: failed to marshal command for device %s: %v", deviceUID, err)
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	log.Printf("📤 SetDeviceProperties: publishing to MQTT")
	log.Printf("   MQTT Topic: %s", topic)
	log.Printf("   MQTT Payload (JSON): %s", string(commandJSON))
	log.Printf("   MQTT Format: {\"cmd\":\"update\",\"requestId\":\"%s\",\"data\":{...}} (使用cmd/data外层结构 ✅)", requestID)
	log.Printf("   Protocol: /prop/productId/UID/get (符合协议规范 ✅)")

	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		log.Printf("❌ SetDeviceProperties: failed to publish to MQTT for device %s, topic: %s, error: %v", deviceUID, topic, err)
		return fmt.Errorf("failed to publish command: %w", err)
	}

	log.Printf("✅ SetDeviceProperties: MQTT command published successfully - device: %s, requestId: %s, topic: %s", deviceUID, requestID, topic)

	// 等待响应
	log.Printf("⏳ SetDeviceProperties: waiting for device response - device: %s, requestId: [%s] (length: %d), timeout: 10s", deviceUID, requestID, len(requestID))
	log.Printf("   Redis key to check: cmd:response:%s", requestID)
	_, err = s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		log.Printf("❌ SetDeviceProperties: failed to get response from device %s, requestId: [%s] (length: %d), error: %v", deviceUID, requestID, len(requestID), err)
		// 检查 Redis 中是否有类似的 key
		log.Printf("   Checking Redis for similar keys...")
		return fmt.Errorf("failed to get response: %w", err)
	}

	log.Printf("✅ SetDeviceProperties: received response from device %s, requestId: [%s] (length: %d)", deviceUID, requestID, len(requestID))
	return nil
}

// SubscribeRealtimeData 订阅实时数据
// 协议：/prefix/monitor/productId/UID/get
func (s *RadarService) SubscribeRealtimeData(ctx context.Context, deviceUID string, content interface{}, duration int) error {
	// 验证参数
	if duration < 1 || duration > 3600 {
		return fmt.Errorf("invalid duration: %d (must be between 1 and 3600)", duration)
	}

	// 构建订阅命令（根据协议文档 3.7.1 节）
	// 注意：content 字段必须使用字符串格式（兼容当前固件）
	contentStr := fmt.Sprintf("%v", content)

	command := map[string]interface{}{
		"cmd": "subscription",
		"data": map[string]interface{}{
			"content":  contentStr, // 使用字符串格式
			"duration": duration,
		},
	}

	// 发送命令（注意：monitor 订阅不需要 requestId，因为设备会持续推送数据）
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	topic := s.mqttClient.BuildCommandTopic("monitor", deviceUID)

	// 记录发送的订阅命令格式
	log.Printf("📤 Monitor Subscription: sending to device %s", deviceUID)
	log.Printf("   MQTT Topic: %s", topic)
	log.Printf("   MQTT Payload (JSON): %s", string(commandJSON))
	log.Printf("   MQTT Format: {\"cmd\":\"subscription\",\"data\":{\"content\":\"%s\",\"duration\":%d}} (使用cmd/data外层结构 ✅)", contentStr, duration)

	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		log.Printf("❌ Monitor Subscription: failed to publish to device %s, topic: %s, error: %v", deviceUID, topic, err)
		return fmt.Errorf("failed to publish command: %w", err)
	}

	log.Printf("✅ Realtime data subscription command sent successfully: device=%s, topic=%s, content=%s, duration=%d", deviceUID, topic, contentStr, duration)

	// monitor 订阅不需要等待响应，设备会持续在 /monitor/.../post 主题上推送数据
	// 这些数据会被 mqtt_consumer 的 handleMonitorMessage 处理并发布到 Redis Stream

	return nil
}

// CallDeviceFunction 调用设备功能（重启、清除数据等）
// 协议：/prefix/func/productId/UID/get
// dev 参数：0-重启雷达和主控，1-只重启雷达，2-只重启主控，100-清除设备数据，101-清除雷达数据，102-清除主控数据
func (s *RadarService) CallDeviceFunction(ctx context.Context, deviceUID string, dev int) error {
	// 验证参数
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

	// 生成请求ID：C + uid最后4位 + . + timestamp
	uidLast4 := getLast4Chars(deviceUID)
	timestamp := time.Now().Unix()
	requestID := fmt.Sprintf("C%s.%d", uidLast4, timestamp)

	// 构建功能调用命令（根据协议文档 3.8 节）
	command := map[string]interface{}{
		"cmd":       "control",
		"requestId": requestID,
		"data": map[string]interface{}{
			"dev": dev,
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

	log.Printf("Device function command sent: %s, requestId: %s, topic: %s, dev: %d", deviceUID, requestID, topic, dev)

	// 等待响应（重启操作需要更长时间）
	timeout := 10 * time.Second
	if dev == 0 || dev == 1 || dev == 2 {
		timeout = 30 * time.Second // 重启操作需要更长时间
	}

	_, err = s.waitForResponse(ctx, requestID, timeout)
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

	// 设备端可能将 requestId 截断到 19 字符，准备截断版本用于查找
	truncatedRequestID := requestID
	if len(requestID) > 19 {
		truncatedRequestID = requestID[:19]
		log.Printf("⚠️ waitForResponse: original requestId [%s] (length: %d) may be truncated by device to [%s] (length: %d)", requestID, len(requestID), truncatedRequestID, len(truncatedRequestID))
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for response: %s", requestID)
		case <-ticker.C:
			// 先尝试用完整 requestId 查找
			response, err := s.streamPublisher.GetCommandResponse(ctx, requestID)
			if err == nil {
				log.Printf("✅ waitForResponse: found response for requestId [%s] (length: %d)", requestID, len(requestID))
				// 检查响应中是否有错误信息
				if code, ok := response["code"].(float64); ok && code != 200 {
					msg := "unknown error"
					if m, ok := response["msg"].(string); ok {
						msg = m
					}
					return nil, fmt.Errorf("device returned error: code=%.0f, msg=%s", code, msg)
				}
				return response, nil
			}

			// 如果完整 requestId 找不到，且长度超过 19，尝试用截断版本查找
			if err == redis.Nil && len(requestID) > 19 && truncatedRequestID != requestID {
				response, err = s.streamPublisher.GetCommandResponse(ctx, truncatedRequestID)
				if err == nil {
					log.Printf("✅ waitForResponse: found response using truncated requestId [%s] (length: %d) for original [%s] (length: %d)", truncatedRequestID, len(truncatedRequestID), requestID, len(requestID))
					// 检查响应中是否有错误信息
					if code, ok := response["code"].(float64); ok && code != 200 {
						msg := "unknown error"
						if m, ok := response["msg"].(string); ok {
							msg = m
						}
						return nil, fmt.Errorf("device returned error: code=%.0f, msg=%s", code, msg)
					}
					return response, nil
				}
			}

			// 如果错误是 Redis key 不存在（响应未找到），继续等待
			// redis.Nil 错误表示 key 不存在
			if err == redis.Nil {
				continue
			}
			// 其他错误直接返回
			return nil, fmt.Errorf("failed to get response: %w", err)
		}
	}
}

// getLast4Chars 获取字符串的最后4个字符，如果长度不足4位则返回整个字符串
func getLast4Chars(s string) string {
	if len(s) <= 4 {
		return s
	}
	return s[len(s)-4:]
}

// convertPropertiesToStrings 将所有属性值转换为字符串（厂家要求所有值都使用字符串格式）
func convertPropertiesToStrings(properties map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range properties {
		switch val := v.(type) {
		case string:
			result[k] = val
		case int, int8, int16, int32, int64:
			result[k] = fmt.Sprintf("%d", val)
		case uint, uint8, uint16, uint32, uint64:
			result[k] = fmt.Sprintf("%d", val)
		case float32, float64:
			result[k] = fmt.Sprintf("%.0f", val)
		case bool:
			if val {
				result[k] = "1"
			} else {
				result[k] = "0"
			}
		default:
			// 其他类型转换为字符串
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
