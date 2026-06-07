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
	"os"
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
	"go.uber.org/zap"
)

// qlVerbose 与 consumer.isQinglanVerboseLog 同源；service 包内的热路径用它做闸门
// ——属性轮询每 30s × 设备数 = 频次很高，默认应静默。
func qlVerbose() bool { return os.Getenv("QINGLAN_VERBOSE_LOG") == "true" }

// 设备响应错误码常量
const (
	DeviceResponseCodeSuccess      = 200 // 成功
	DeviceResponseCodeFailure      = 500 // 失败
	DeviceResponseCodeOffline      = 777 // 设备离线
	DeviceResponseCodeNotSupported = 778 // 该设备不适用该模式
)

// HealthRefresher 接收成功的设备属性读响应，用于"正向恢复"加速：
// 任何 GetDeviceProperties 成功都会触发，把外部观察到的健康证据同步给
// DeviceSubscriptionManager，让 alarm 在几秒内 auto-recover，而不是等下一个
// 10min 健康检查 tick。
//
// 不接管反向告警：onset (value=1) 路径仍由 healthCheckMonitor 独占——外部查询
// 失败原因（瞬时超时/设备 app 抖动/上游 MQTT 抖动）不能可靠推断"设备真离线"。
type HealthRefresher interface {
	RefreshHealthFromProps(ctx context.Context, deviceUID string, props map[string]interface{})
}

// RadarService 雷达服务
type RadarService struct {
	config          *config.Config
	mqttClient      *mqtt.Client
	redisClient     *redis.Client
	deviceRepo      repository.DeviceRepository
	streamPublisher *consumer.StreamPublisher
	mqttConsumer    *consumer.MQTTConsumer
	healthRefresher HealthRefresher // 由 main 在 subscriptionManager 创建后注入；nil 时行为退化为旧路径
	logger          *zap.Logger
}

// NewRadarService 创建雷达服务
func NewRadarService(
	cfg *config.Config,
	mqttClient *mqtt.Client,
	redisClient *redis.Client,
	deviceRepo repository.DeviceRepository,
	streamPublisher *consumer.StreamPublisher,
	mqttConsumer *consumer.MQTTConsumer,
	logger *zap.Logger,
) (*RadarService, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RadarService{
		config:          cfg,
		mqttClient:      mqttClient,
		redisClient:     redisClient,
		deviceRepo:      deviceRepo,
		streamPublisher: streamPublisher,
		mqttConsumer:    mqttConsumer,
		logger:          logger,
	}, nil
}

// SetHealthRefresher 注入健康刷新器（由 main 在 subscriptionManager 创建后调用）。
// 未注入时 GetDeviceProperties 行为不变，向后兼容；测试场景可省略。
func (s *RadarService) SetHealthRefresher(refresher HealthRefresher) {
	s.healthRefresher = refresher
}

// Start 启动服务
func (s *RadarService) Start(ctx context.Context) error {
	if err := s.mqttConsumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MQTT consumer: %w", err)
	}
	s.logger.Info("radar service started")
	return nil
}

// Stop 停止服务
func (s *RadarService) Stop(ctx context.Context) error {
	if err := s.mqttConsumer.Stop(ctx); err != nil {
		s.logger.Warn("stop mqtt consumer", zap.Error(err))
	}
	s.logger.Info("radar service stopped")
	return nil
}

// GetDeviceProperties 读取设备属性
// 协议：/prefix/prop/productId/UID/get
// 当 len(keys)>1 时：分笔依次 MQTT 读每个 key，每个指令间额外 50ms，再合并返回（设备分笔读取，间隔由本服务控制）
func (s *RadarService) GetDeviceProperties(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if len(keys) <= 1 {
		r, err := s.readOneBatch(ctx, deviceUID, keys)
		if err != nil {
			return nil, err
		}
		result = r
	} else {
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
		result = merged
	}

	// 正向恢复 hook：成功读到任意属性都是"设备在线"证据；若 props 含 wifi_rssi/accelera+install_style
	// 还能恢复 SignalPoor/AngleAbnormal。同步调用——publishHealthIfChanged 内部 transition-only +
	// Redis XADD 量级几 ms，相对 MQTT 等待（多秒）可忽略。
	if s.healthRefresher != nil && len(result) > 0 {
		s.healthRefresher.RefreshHealthFromProps(ctx, deviceUID, result)
	}
	return result, nil
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
	if qlVerbose() {
		s.logger.Debug("read command",
			zap.String("device", deviceUID),
			zap.String("request_id", requestID),
			zap.Any("command", command),
		)
	}
	commandJSON, err := json.Marshal(command)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		return nil, fmt.Errorf("failed to publish command: %w", err)
	}
	if qlVerbose() {
		s.logger.Debug("read command sent, waiting response",
			zap.String("device", deviceUID),
			zap.String("request_id", requestID),
			zap.Strings("keys", keys),
		)
	}
	response, err := s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		s.logger.Warn("get device properties no response",
			zap.String("device", deviceUID),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	if data, ok := response["data"].(map[string]interface{}); ok {
		if qlVerbose() {
			s.logger.Debug("get device properties ok",
				zap.String("device", deviceUID),
				zap.String("request_id", requestID),
				zap.Any("data", data),
			)
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
	// 用 EncodeAlarmItemsToDeviceProps 一次拿齐 radar_func_ctrl + fall_param + heart_breath_param
	// （之前只单独调 fall/heart 两个 sub-func，遗漏了 work_model 的写入，导致 settings 改 Sleep 也落不到 firmware）
	if alarmItemsJSON, ok := properties["_alarm_items_json"].(string); ok {
		var alarmItems []alarm.AlarmItem
		if err := json.Unmarshal([]byte(alarmItemsJSON), &alarmItems); err == nil {
			built, encErr := decode.EncodeAlarmItemsToDeviceProps(alarmItems)
			if encErr != nil {
				s.logger.Warn("config_write encode alarm items", zap.Error(encErr))
			}
			if v, ok := built["radar_func_ctrl"]; ok {
				convertedProperties["radar_func_ctrl"] = v
				s.logger.Info("config_write built radar_func_ctrl", zap.Any("value", v))
			}
			if v, ok := built["fall_param"]; ok {
				convertedProperties["fall_param"] = v
				s.logger.Info("config_write built fall_param")
			}
			if v, ok := built["heart_breath_param"]; ok {
				convertedProperties["heart_breath_param"] = v
				s.logger.Info("config_write built heart_breath_param")
			}
		} else {
			s.logger.Warn("config_write unmarshal _alarm_items_json", zap.Error(err))
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
				convertedProperties[key] = value
				s.logger.Info("config_write using key directly", zap.String("key", key))
			}
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
		s.logger.Warn("set device properties marshal", zap.String("device", deviceUID), zap.Error(err))
		return 0, fmt.Errorf("failed to marshal command: %w", err)
	}
	topic := s.mqttClient.BuildCommandTopic("prop", deviceUID)
	s.logger.Info("set device properties sent",
		zap.String("device", deviceUID),
		zap.String("request_id", requestID),
		zap.String("payload", string(commandJSON)),
	)
	if err := s.mqttClient.Publish(topic, 1, false, commandJSON); err != nil {
		s.logger.Warn("set device properties publish",
			zap.String("device", deviceUID),
			zap.String("topic", topic),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to publish command: %w", err)
	}
	response, err := s.waitForResponse(ctx, requestID, 10*time.Second)
	if err != nil {
		s.logger.Warn("set device properties no response",
			zap.String("device", deviceUID),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
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
// 每段统一 Trim 掉首尾杂散的 { } , 再重新包 {}：split 后首段残留前 {、末段残留后 }、整串尾逗号会让末段成 "..},"——
// 旧实现只补不去，会拼出畸形 "{..},}"（设备虽容忍但不规范）；此处规范化根除。段内坐标逗号在中间，Trim 只动两端不受影响。
func splitDeclareAreaOnePerRequest(declareArea string) []string {
	parts := strings.Split(strings.TrimSpace(declareArea), "},{")
	var out []string
	for _, p := range parts {
		p = strings.Trim(p, "{}, \t\r\n")
		if p == "" {
			continue
		}
		out = append(out, "{"+p+"}")
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
		s.logger.Warn("monitor subscription publish",
			zap.String("device", deviceUID),
			zap.String("topic", topic),
			zap.Error(err),
		)
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

	s.logger.Info("device function command sent",
		zap.String("device", deviceUID),
		zap.String("request_id", requestID),
		zap.String("topic", topic),
		zap.Int("dev", dev),
	)

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
		s.logger.Warn("requestId may be truncated by device",
			zap.String("original", requestID),
			zap.Int("orig_len", len(requestID)),
			zap.String("truncated", truncatedRequestID),
		)
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
