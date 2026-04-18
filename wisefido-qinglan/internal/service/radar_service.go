// Package service 提供雷达设备 MQTT 读写与控制。
//
// requestID 格式约定（用于与设备请求/响应对齐，前缀区分操作类型，uid4=设备 UID 后 4 位）：
//   - R{uid4}.{ts}  Read 读属性：readOneBatch，ts=UnixMilli，保证分笔读每笔唯一
//   - S{uid4}.{ts}  Set 写属性：SetDeviceProperties，ts=Unix
//   - C{uid4}.{ts}  Control 调功能：CallDeviceFunction（重启、清数据等），ts=Unix
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"owl-common/alarm"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/consumer"
	"wisefido-qinglan/internal/decode"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/mqtt"
	"wisefido-qinglan/internal/repository"

	"github.com/go-redis/redis/v8"
)

// 设备响应错误码常量
const (
	DeviceResponseCodeSuccess      = 200 // 成功
	DeviceResponseCodeFailure      = 500 // 失败
	DeviceResponseCodeOffline      = 777 // 设备离线
	DeviceResponseCodeNotSupported = 778 // 该设备不适用该模式
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
// requestId 必须每笔唯一，否则多 key 分笔读时同一秒内会复用，waitForResponse 会立即命中上一笔的旧响应导致 merge 丢 key（如 rectangle）
func (s *RadarService) readOneBatch(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error) {
	uidLast4 := getLast4Chars(deviceUID)
	requestID := fmt.Sprintf("R%s.%d", uidLast4, time.Now().UnixMilli())
	command := map[string]interface{}{"cmd": "read", "requestId": requestID}
	// 即使 keys 为空（读取所有属性），也添加 data 字段，但 key 为空数组
	// 设备可能期望 data 字段存在，即使 key 为空数组也表示读取所有属性
	if len(keys) > 0 {
		command["data"] = map[string]interface{}{"key": keys}
	} else {
		// keys 为空时，添加空的 data 字段，表示读取所有属性
		command["data"] = map[string]interface{}{"key": []string{}}
	}
	log.Printf("📤 Read Command: device=%s, requestId=%s, command=%+v", deviceUID, requestID, command)
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		return nil, fmt.Errorf("failed to publish command: %w", err)
	}
	log.Printf("Send MQTT: device=%s, requestId=%s, keys: %v", deviceUID, requestID, keys)
	log.Printf("Wait device response: ⏳ device=%s, requestId=%s, GetDeviceProperties", deviceUID, requestID)
	response, err := s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		log.Printf("❌ GetDeviceProperties: failed to get response - device: %s, requestId: [%s], error: %v", deviceUID, requestID, err)
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	if data, ok := response["data"].(map[string]interface{}); ok {
		if b, err := json.Marshal(data); err == nil {
			log.Printf("✅ GetDeviceProperties: device=%s, requestId=%s, data=%s", deviceUID, requestID, string(b))
		} else {
			log.Printf("✅ GetDeviceProperties: device=%s, requestId=%s, data=%+v", deviceUID, requestID, data)
		}
		return data, nil
	}
	return nil, fmt.Errorf("no property data in response")
}

// SetDeviceProperties 设置设备属性
// 协议：/prefix/prop/productId/UID/get
// 返回设备响应码（200=成功，500/777/778=失败）和错误；发起端可透传 device_code 给前端。
func (s *RadarService) SetDeviceProperties(ctx context.Context, deviceUID string, properties map[string]interface{}) (deviceCode int, err error) {
	convertedProperties := make(map[string]interface{})

	// 处理 _alarm_items_json：从 wisefido-data 传递的原始 AlarmItem[] 数据
	// 构建 fall_param 和 heart_breath_param 的 base64，并应用单位转换
	if alarmItemsJSON, ok := properties["_alarm_items_json"].(string); ok {
		var alarmItems []alarm.AlarmItem
		if err := json.Unmarshal([]byte(alarmItemsJSON), &alarmItems); err == nil {
			// 构建 fall_param base64（应用转换：秒 → 10秒单位）
			if fallParamBase64, err := decode.EncodeFallParam(alarmItems); err == nil {
				convertedProperties["fall_param"] = fallParamBase64
				log.Printf("[CONFIG_WRITE] ✅ Built fall_param base64 from AlarmItems (applied unit conversion)")
			} else {
				log.Printf("[CONFIG_WRITE] ❌ Failed to build fall_param: %v", err)
			}

			// 构建 heart_breath_param base64（无需转换，单位已一致）
			if heartBreathParamBase64, err := decode.EncodeHeartBreathParam(alarmItems); err == nil {
				convertedProperties["heart_breath_param"] = heartBreathParamBase64
				log.Printf("[CONFIG_WRITE] ✅ Built heart_breath_param base64 from AlarmItems")
			} else {
				log.Printf("[CONFIG_WRITE] ❌ Failed to build heart_breath_param: %v", err)
			}
		} else {
			log.Printf("[CONFIG_WRITE] ❌ Failed to unmarshal _alarm_items_json: %v", err)
		}
	}

	for key, value := range properties {
		// 跳过特殊字段 _alarm_items_json（已在上面的逻辑中处理）
		if key == "_alarm_items_json" {
			continue
		}

		// 如果 fall_param 或 heart_breath_param 已经在 convertedProperties 中（从 _alarm_items_json 构建），跳过
		// 否则，如果直接传递了这些字段，直接添加到 convertedProperties
		if key == "fall_param" || key == "heart_breath_param" {
			if _, exists := convertedProperties[key]; !exists {
				// 直接传递的 base64 值，直接使用
				convertedProperties[key] = value
				log.Printf("[CONFIG_WRITE] ✅ Using %s directly from properties (already base64 encoded)", key)
			}
			// 如果已存在（从 _alarm_items_json 构建），跳过
			continue
		}

		convertedProperties[key] = value
	}
	stringProperties := convertPropertiesToStrings(convertedProperties)

	// 设备接口一次只能设置一个区域：declare_area 含多区域时拆成多次下发，每次只发一个区域。多区域用逗号分隔 {area1},{area2}
	declareAreaVal, hasDeclareArea := stringProperties["declare_area"].(string)
	if hasDeclareArea && declareAreaVal != "" && strings.Contains(declareAreaVal, "},{") {
		parts := splitDeclareAreaOnePerRequest(declareAreaVal)
		if len(parts) > 0 {
			// 1) 若有其他属性，先发一条（不含 declare_area）
			rest := make(map[string]interface{})
			for k, v := range stringProperties {
				if k != "declare_area" {
					rest[k] = v
				}
			}
			if len(rest) > 0 {
				code, err := s.sendOneSetProperties(ctx, deviceUID, rest)
				if err != nil {
					return code, err
				}
				time.Sleep(300 * time.Millisecond)
			}
			// 2) 按区域逐条下发，每次只发一个区域
			for i, one := range parts {
				code, err := s.sendOneSetProperties(ctx, deviceUID, map[string]interface{}{"declare_area": one})
				if err != nil {
					return code, err
				}
				if i < len(parts)-1 {
					time.Sleep(300 * time.Millisecond)
				}
			}
			return DeviceResponseCodeSuccess, nil
		}
	}

	return s.sendOneSetProperties(ctx, deviceUID, stringProperties)
}

// sendOneSetProperties 发送单次 MQTT 属性设置并等待响应（一次一条命令）
func (s *RadarService) sendOneSetProperties(ctx context.Context, deviceUID string, stringProperties map[string]interface{}) (deviceCode int, err error) {
	uidLast4 := getLast4Chars(deviceUID)
	requestID := fmt.Sprintf("S%s.%d", uidLast4, time.Now().UnixMilli())

	command := map[string]interface{}{
		"cmd":       "update",
		"requestId": requestID,
		"data":      stringProperties,
	}
	commandJSON, err := json.Marshal(command)
	if err != nil {
		log.Printf("❌ SetDeviceProperties: failed to marshal command, device=%s: %v", deviceUID, err)
		return 0, fmt.Errorf("failed to marshal command: %w", err)
	}
	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	log.Printf("📤 SetDeviceProperties send MQTT: device=%s, requestId=%s, payload=%s", deviceUID, requestID, string(commandJSON))
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		log.Printf("❌ SetDeviceProperties: failed to publish, device=%s, topic=%s: %v", deviceUID, topic, err)
		return 0, fmt.Errorf("failed to publish command: %w", err)
	}
	response, err := s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		log.Printf("❌ SetDeviceProperties: no response, device=%s, requestId=%s: %v", deviceUID, requestID, err)
		return 0, fmt.Errorf("failed to get response: %w", err)
	}
	code := 0
	if c, ok := response["code"].(float64); ok {
		code = int(c)
	}
	if code != DeviceResponseCodeSuccess {
		msg := "unknown error"
		if m, ok := response["msg"].(string); ok {
			msg = m
		}
		return code, fmt.Errorf("device returned code %d: %s", code, msg)
	}
	return code, nil
}

// splitDeclareAreaOnePerRequest 将 declare_area 按 },{ 拆成多条，每条一个区域（设备一次只能设置一个区域）。多区域用逗号分隔。
func splitDeclareAreaOnePerRequest(declareArea string) []string {
	declareArea = strings.TrimSpace(declareArea)
	parts := strings.Split(declareArea, "},{")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "{") {
			p = "{" + p
		}
		if !strings.HasSuffix(p, "}") {
			p = p + "}"
		}
		out = append(out, p)
	}
	if len(out) == 0 && declareArea != "" {
		// 单区域无 },{，整条作为一条
		if !strings.HasPrefix(declareArea, "{") {
			declareArea = "{" + declareArea
		}
		if !strings.HasSuffix(declareArea, "}") {
			declareArea = declareArea + "}"
		}
		out = append(out, declareArea)
	}
	return out
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

	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		log.Printf("❌ Monitor Subscription: failed to publish to device %s, topic: %s, error: %v", deviceUID, topic, err)
		return fmt.Errorf("failed to publish command: %w", err)
	}

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
				return response, nil
			}

			// 如果完整 requestId 找不到，且长度超过 19，尝试用截断版本查找
			if err == redis.Nil && len(requestID) > 19 && truncatedRequestID != requestID {
				response, err = s.streamPublisher.GetCommandResponse(ctx, truncatedRequestID)
				if err == nil {
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

// declareAreaToDeviceString 将 declare_area 转为设备协议字符串：id,type,x1,y1,x2,y2,x3,y3,x4,y4（多区域用 ; 分隔）
// 入参已是标准格式（单位由上游保证），此处只做拼接不转换
func declareAreaToDeviceString(v interface{}) string {
	var arr []interface{}
	switch val := v.(type) {
	case []interface{}:
		arr = val
	case string:
		if err := json.Unmarshal([]byte(val), &arr); err != nil || len(arr) == 0 {
			return ""
		}
	default:
		return ""
	}
	if len(arr) == 0 {
		return ""
	}
	var parts []string
	for _, it := range arr {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		id := numFromMap(m, "id", "area_id")
		typ := ""
		if t := m["type"]; t != nil {
			typ = fmt.Sprintf("%v", t)
		}
		x1, y1 := numFromMap(m, "x1", ""), numFromMap(m, "y1", "")
		x2, y2 := numFromMap(m, "x2", ""), numFromMap(m, "y2", "")
		x3, y3 := numFromMap(m, "x3", ""), numFromMap(m, "y3", "")
		x4, y4 := numFromMap(m, "x4", ""), numFromMap(m, "y4", "")
		// 上游已是标准格式（只做一次 cm→dm），此处不再 /10
		parts = append(parts, fmt.Sprintf("%d,%s,%d,%d,%d,%d,%d,%d,%d,%d", id, typ, x1, y1, x2, y2, x3, y3, x4, y4))
	}
	return strings.Join(parts, ";")
}

func numFromMap(m map[string]interface{}, keys ...string) int {
	for _, k := range keys {
		if v := m[k]; v != nil {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case string:
				var i int
				if _, err := fmt.Sscanf(n, "%d", &i); err == nil {
					return i
				}
			}
		}
	}
	return 0
}

// convertPropertiesToStrings 将所有属性值转换为字符串（厂家要求所有值都使用字符串格式）
func convertPropertiesToStrings(properties map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range properties {
		if k == "declare_area" {
			if s, ok := v.(string); ok && s != "" {
				result[k] = s
			} else {
				result[k] = declareAreaToDeviceString(v)
			}
			continue
		}
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
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
