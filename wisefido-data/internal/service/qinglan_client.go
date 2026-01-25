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

// NewQinglanClient 创建 Qinglan 客户端
func NewQinglanClient(apiBaseURL string, logger *zap.Logger) *QinglanClient {
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

	return out.Data, nil
}

// SetDeviceProperties 设置设备属性（PUT /api/v1/radar/devices/{uid}/properties）。
// 设备要求必须分组设置，不能一次性设置所有属性。
// 分组规则：
//   1. 工作模式组：radar_func_ctrl, radar_install_style, radar_install_height（3个key一组）
//   2. 呼吸心率参数：heart_breath_param（单独1组）
//   3. 跌倒参数：fall_param（单独1组）
//   4. 边界：rectangle（单独1组）
//   5. 区域：declare_area（一次只能设置一个区域）
func (c *QinglanClient) SetDeviceProperties(ctx context.Context, deviceUID string, properties map[string]interface{}) error {
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
			return fmt.Errorf("marshal properties group %d: %w", i+1, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonData))
		if err != nil {
			return fmt.Errorf("create request for group %d: %w", i+1, err)
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
			return fmt.Errorf("request group %d failed: %w", i+1, err)
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
			return fmt.Errorf("qinglan API status %d for group %d (%s): %s", resp.StatusCode, i+1, groupName, string(body))
		}

		var out struct {
			Success bool   `json:"success"`
			Error   string `json:"error,omitempty"`
		}
		_ = json.Unmarshal(body, &out)
		if !out.Success && out.Error != "" {
			c.logger.Warn("Qinglan SetDeviceProperties: API returned error",
				zap.String("uid", deviceUID),
				zap.Int("group_index", i+1),
				zap.String("group_name", groupName),
				zap.String("error", out.Error),
			)
			return fmt.Errorf("qinglan API error for group %d (%s): %s", i+1, groupName, out.Error)
		}

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
	)

	return nil
}

// groupProperties 将属性按设备要求分组
func (c *QinglanClient) groupProperties(properties map[string]interface{}) []map[string]interface{} {
	groups := make([]map[string]interface{}, 0)

	// 组1：工作模式相关（3个key一组）
	group1 := make(map[string]interface{})
	if v, ok := properties["radar_func_ctrl"]; ok {
		group1["radar_func_ctrl"] = v
	}
	if v, ok := properties["radar_install_style"]; ok {
		group1["radar_install_style"] = v
	}
	if v, ok := properties["radar_install_height"]; ok {
		group1["radar_install_height"] = v
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

	// 其他未分类的属性，每个单独一组
	knownKeys := map[string]bool{
		"radar_func_ctrl":     true,
		"radar_install_style": true,
		"radar_install_height": true,
		"heart_breath_param":  true,
		"fall_param":          true,
		"rectangle":           true,
		"declare_area":        true,
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
