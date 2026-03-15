package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// QinglanClient 调用 wisefido-qinglan HTTP API，提供雷达设备交互功能（属性读写、实时数据订阅、设备功能调用）。
// 替代原 wisefido-radar 内部 API，统一通过 wisefido-qinglan 与设备通信。
type QinglanClient struct {
	apiBaseURL string
	httpClient *http.Client
	logger     *zap.Logger
}

const defaultQinglanBaseURL = "http://127.0.0.1:8081"

// NewQinglanClient 创建 Qinglan 客户端。若 apiBaseURL 为空则使用 defaultQinglanBaseURL，避免 unsupported protocol scheme ""。
func NewQinglanClient(apiBaseURL string, logger *zap.Logger) *QinglanClient {
	if apiBaseURL == "" {
		apiBaseURL = defaultQinglanBaseURL
		if logger != nil {
			logger.Info("QinglanClient: apiBaseURL empty, using default", zap.String("default", apiBaseURL))
		}
	}
	return &QinglanClient{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}
}

// GetDeviceProperties 读取设备属性（GET /api/v1/radar/devices/{uid}/properties?keys=key1,key2）。
// keys 为空时读取所有属性。
func (c *QinglanClient) GetDeviceProperties(ctx context.Context, deviceUID string, keys []string) (map[string]interface{}, error) {
	urlStr := fmt.Sprintf("%s/api/v1/radar/devices/%s/properties", c.apiBaseURL, deviceUID)
	if len(keys) > 0 {
		u, err := url.Parse(urlStr)
		if err == nil {
			q := u.Query()
			q.Set("keys", strings.Join(keys, ","))
			u.RawQuery = q.Encode()
			urlStr = u.String()
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan GetDeviceProperties non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
		Error   string                 `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if !out.Success && out.Error != "" {
		return nil, fmt.Errorf("qinglan API error: %s", out.Error)
	}
	if out.Data == nil {
		return map[string]interface{}{}, nil
	}
	return out.Data, nil
}

// SetDeviceProperties 设置设备属性（PUT /api/v1/radar/devices/{uid}/properties）。
// 设备要求必须分组设置，不能一次性设置所有属性。
// 返回设备响应码（200=成功，500/777/778=失败）和错误，发起端可透传 device_code 给前端。
// 分组规则：
//   1. 工作模式组：radar_func_ctrl, radar_install_style, radar_install_height（3个key一组）
//   2. 呼吸心率参数：heart_breath_param（单独1组）
//   3. 跌倒参数：fall_param（单独1组）
//   4. 边界：rectangle（单独1组）
//   5. 区域：declare_area（一次只能设置一个区域）
func (c *QinglanClient) SetDeviceProperties(ctx context.Context, deviceUID string, properties map[string]interface{}) (deviceCode int, err error) {
	url := fmt.Sprintf("%s/api/v1/radar/devices/%s/properties", c.apiBaseURL, deviceUID)

	// 分组属性
	groups := c.groupProperties(properties)
	
	c.logger.Info("Qinglan SetDeviceProperties: grouping properties for device",
		zap.String("uid", deviceUID),
		zap.Int("total_properties", len(properties)),
		zap.Int("groups", len(groups)),
	)

	// 按组依次发送
	for i, group := range groups {
		if len(group) == 0 {
			continue
		}

		groupName := c.getGroupName(group)
		c.logger.Info("Qinglan SetDeviceProperties: sending group",
			zap.String("uid", deviceUID),
			zap.Int("group_index", i+1),
			zap.String("group_name", groupName),
			zap.Any("group_properties", group),
		)

		reqBody := map[string]interface{}{"properties": group}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return 0, fmt.Errorf("marshal properties group %d: %w", i+1, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonData))
		if err != nil {
			return 0, fmt.Errorf("create request for group %d: %w", i+1, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Error("Qinglan SetDeviceProperties: HTTP request failed",
				zap.String("uid", deviceUID),
				zap.Int("group_index", i+1),
				zap.String("group_name", groupName),
				zap.Error(err),
			)
			return 0, fmt.Errorf("request group %d failed: %w", i+1, err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			c.logger.Warn("Qinglan SetDeviceProperties non-200",
				zap.String("uid", deviceUID),
				zap.Int("group_index", i+1),
				zap.String("group_name", groupName),
				zap.Int("status", resp.StatusCode),
				zap.String("body", string(body)),
			)
			return 0, fmt.Errorf("qinglan API status %d for group %d (%s): %s", resp.StatusCode, i+1, groupName, string(body))
		}

		var out struct {
			Success    bool   `json:"success"`
			Error      string `json:"error,omitempty"`
			DeviceCode int    `json:"device_code"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			c.logger.Warn("Qinglan SetDeviceProperties: failed to parse response",
				zap.String("uid", deviceUID),
				zap.Int("group_index", i+1),
				zap.String("group_name", groupName),
				zap.String("body", string(body)),
				zap.Error(err),
			)
			return 0, fmt.Errorf("qinglan API invalid response for group %d (%s): %w", i+1, groupName, err)
		}
		if !out.Success {
			errorMsg := out.Error
			if errorMsg == "" {
				errorMsg = "unknown error"
			}
			c.logger.Warn("Qinglan SetDeviceProperties: API returned error",
				zap.String("uid", deviceUID),
				zap.Int("group_index", i+1),
				zap.String("group_name", groupName),
				zap.String("error", errorMsg),
				zap.Int("device_code", out.DeviceCode),
			)
			return out.DeviceCode, fmt.Errorf("qinglan API error for group %d (%s): %s", i+1, groupName, errorMsg)
		}
		deviceCode = out.DeviceCode

		c.logger.Info("Qinglan SetDeviceProperties: group completed successfully",
			zap.String("uid", deviceUID),
			zap.Int("group_index", i+1),
			zap.String("group_name", groupName),
			zap.Int("status", resp.StatusCode),
		)

		// 每组之间稍作延迟，避免设备处理不过来
		if i < len(groups)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	c.logger.Info("Qinglan SetDeviceProperties: all groups completed successfully",
		zap.String("uid", deviceUID),
		zap.Int("total_groups", len(groups)),
		zap.Int("device_code", deviceCode),
	)

	return deviceCode, nil
}

// groupProperties 将属性按设备要求分组。支持规范 key（work_model, install_model, install_height cm）与设备 key（radar_*），规范 key 会转为设备 key 且 install_height cm→dm。
func (c *QinglanClient) groupProperties(properties map[string]interface{}) []map[string]interface{} {
	groups := make([]map[string]interface{}, 0)

	// 组1：工作模式相关（3个key一组）
	group1 := make(map[string]interface{})
	if v, ok := properties["radar_func_ctrl"]; ok && v != nil {
		group1["radar_func_ctrl"] = toIntForProp(v)
	} else if v, ok := properties["work_model"]; ok && v != nil {
		group1["radar_func_ctrl"] = toIntForProp(v)
	}
	if v, ok := properties["radar_install_style"]; ok && v != nil {
		group1["radar_install_style"] = v
	} else if v, ok := properties["install_model"]; ok && v != nil {
		group1["radar_install_style"] = installModelToDeviceStyle(v)
	}
	if v, ok := properties["radar_install_height"]; ok && v != nil {
		group1["radar_install_height"] = toIntForProp(v)
	} else if v, ok := properties["install_height"]; ok && v != nil {
		if n := toIntForProp(v); n > 0 {
			group1["radar_install_height"] = n / 10 // cm → dm
		}
	}
	if len(group1) > 0 {
		groups = append(groups, group1)
	}

	// 组2：呼吸心率参数（单独1组）
	if v, ok := properties["heart_breath_param"]; ok {
		groups = append(groups, map[string]interface{}{
			"heart_breath_param": v,
		})
	}

	// 组3：跌倒参数（单独1组）
	if v, ok := properties["fall_param"]; ok {
		groups = append(groups, map[string]interface{}{
			"fall_param": v,
		})
	}

	// 组4：边界（单独1组）
	if v, ok := properties["rectangle"]; ok {
		groups = append(groups, map[string]interface{}{
			"rectangle": v,
		})
	}

	// 组5：区域（一次只能设置一个区域）
	// 注意：declare_area 可能包含多个区域，需要分别发送
	// 这里假设 declare_area 是一个数组或对象，需要根据实际格式处理
	if v, ok := properties["declare_area"]; ok {
		// 如果 declare_area 是数组，每个元素单独发送
		if areas, ok := v.([]interface{}); ok {
			for _, area := range areas {
				groups = append(groups, map[string]interface{}{
					"declare_area": []interface{}{area},
				})
			}
		} else {
			// 如果是单个区域，直接发送
			groups = append(groups, map[string]interface{}{
				"declare_area": v,
			})
		}
	}

	// 组6：_alarm_items_json（特殊字段，wisefido-qinglan 会从中构建 fall_param 和 heart_breath_param）
	if v, ok := properties["_alarm_items_json"]; ok {
		groups = append(groups, map[string]interface{}{
			"_alarm_items_json": v,
		})
	}

	// 其他未分类的属性，每个单独一组（含规范 key，避免重复下发）
	knownKeys := map[string]bool{
		"radar_func_ctrl":     true,
		"radar_install_style":  true,
		"radar_install_height": true,
		"work_model":          true,
		"install_model":       true,
		"install_height":      true,
		"heart_breath_param":   true,
		"fall_param":          true,
		"rectangle":           true,
		"declare_area":        true,
		"_alarm_items_json":   true,
	}
	for k, v := range properties {
		if !knownKeys[k] {
			groups = append(groups, map[string]interface{}{
				k: v,
			})
		}
	}

	return groups
}

// getGroupName 获取组名（用于日志）
func (c *QinglanClient) getGroupName(group map[string]interface{}) string {
	keys := make([]string, 0, len(group))
	for k := range group {
		keys = append(keys, k)
	}
	if len(keys) == 1 {
		return keys[0]
	}
	return fmt.Sprintf("%v", keys)
}

// toIntForProp 从 interface{} 取整数值（用于 work_model、install_height 等）
func toIntForProp(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	case string:
		var n int
		if _, err := fmt.Sscanf(strings.TrimSpace(x), "%d", &n); err == nil {
			return n
		}
	}
	return 0
}

// installModelToDeviceStyle 规范 install_model（0/1/2）→ 设备 radar_install_style 字符串 "0"/"1"/"2"。兼容字符串 "ceiling"/"wall"/"corn" 入参。
func installModelToDeviceStyle(v interface{}) string {
	if s, ok := v.(string); ok {
		switch s {
		case "ceiling":
			return "0"
		case "wall":
			return "1"
		case "corn":
			return "2"
		}
	}
	n := toIntForProp(v)
	if n >= 0 && n <= 2 {
		return fmt.Sprintf("%d", n)
	}
	return "1"
}

// SubscribeRealtimeData 订阅实时数据（POST /api/v1/radar/devices/{uid}/subscribe）。
// content: 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
// duration: 订阅时长（秒），最大3600
func (c *QinglanClient) SubscribeRealtimeData(ctx context.Context, deviceUID string, content interface{}, duration int) error {
	url := fmt.Sprintf("%s/api/v1/radar/devices/%s/subscribe", c.apiBaseURL, deviceUID)

	reqBody := map[string]interface{}{
		"content":  content,
		"duration": duration,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan SubscribeRealtimeData non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	_ = json.Unmarshal(body, &out)
	if !out.Success && out.Error != "" {
		return fmt.Errorf("qinglan API error: %s", out.Error)
	}

	return nil
}

// GetDeviceStatus 获取设备在线状态（GET /api/v1/radar/devices/{uid}/status）。
// 返回设备状态：online, offline, 或 unsubscribed
func (c *QinglanClient) GetDeviceStatus(ctx context.Context, deviceUID string) (string, error) {
	urlStr := fmt.Sprintf("%s/api/v1/radar/devices/%s/status", c.apiBaseURL, deviceUID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan GetDeviceStatus non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return "", fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool `json:"success"`
		Data    struct {
			DeviceUID string `json:"device_uid"`
			Status    string `json:"status"`
		} `json:"data"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if !out.Success && out.Error != "" {
		return "", fmt.Errorf("qinglan API error: %s", out.Error)
	}

	return out.Data.Status, nil
}

// CallDeviceFunction 调用设备功能（POST /api/v1/radar/devices/{uid}/function）。
// dev: 0-重启雷达和主控，1-只重启雷达，2-只重启主控
func (c *QinglanClient) CallDeviceFunction(ctx context.Context, deviceUID string, dev int) error {
	url := fmt.Sprintf("%s/api/v1/radar/devices/%s/function", c.apiBaseURL, deviceUID)

	reqBody := map[string]interface{}{
		"dev": dev,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("Qinglan CallDeviceFunction non-200",
			zap.String("uid", deviceUID),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("qinglan API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	_ = json.Unmarshal(body, &out)
	if !out.Success && out.Error != "" {
		return fmt.Errorf("qinglan API error: %s", out.Error)
	}

	return nil
}
