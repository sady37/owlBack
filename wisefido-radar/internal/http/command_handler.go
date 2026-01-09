package http

import (
	"encoding/json"
	"net/http"
	
	"go.uber.org/zap"
)

// CommandHandler 命令处理器
// 提供内部 HTTP API，供 wisefido-data 调用
type CommandHandler struct {
	commandService *CommandService
	logger         *zap.Logger
}

// NewCommandHandler 创建命令处理器
func NewCommandHandler(commandService *CommandService, logger *zap.Logger) *CommandHandler {
	return &CommandHandler{
		commandService: commandService,
		logger:         logger,
	}
}

// GetProperties 读取设备属性
// POST /internal/api/v1/radar/devices/{uid}/properties/get
func (h *CommandHandler) GetProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径提取 UID
	// 路径格式：/internal/api/v1/radar/devices/{uid}/properties/get
	uid := extractUIDFromPath(r.URL.Path, "/internal/api/v1/radar/devices/", "/properties/get")
	if uid == "" {
		http.Error(w, "Invalid device UID", http.StatusBadRequest)
		return
	}
	
	// 解析请求体
	var req struct {
		Keys []string `json:"keys"` // 要读取的属性 key 列表，空数组表示读取所有属性
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 调用命令服务
	properties, err := h.commandService.GetDeviceProperties(r.Context(), uid, req.Keys)
	if err != nil {
		h.logger.Error("Failed to get device properties",
			zap.String("uid", uid),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 200,
		"msg":  "操作成功",
		"data": properties,
	})
}

// SetProperties 设置设备属性
// POST /internal/api/v1/radar/devices/{uid}/properties/set
func (h *CommandHandler) SetProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径提取 UID
	uid := extractUIDFromPath(r.URL.Path, "/internal/api/v1/radar/devices/", "/properties/set")
	if uid == "" {
		http.Error(w, "Invalid device UID", http.StatusBadRequest)
		return
	}
	
	// 解析请求体
	var req struct {
		Properties map[string]interface{} `json:"properties"` // 属性 key-value 对
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 调用命令服务
	if err := h.commandService.SetDeviceProperties(r.Context(), uid, req.Properties); err != nil {
		h.logger.Error("Failed to set device properties",
			zap.String("uid", uid),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 200,
		"msg":  "操作成功",
		"data": nil,
	})
}

// SubscribeRealtime 订阅实时数据
// POST /internal/api/v1/radar/devices/{uid}/realtime/subscribe
func (h *CommandHandler) SubscribeRealtime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径提取 UID
	uid := extractUIDFromPath(r.URL.Path, "/internal/api/v1/radar/devices/", "/realtime/subscribe")
	if uid == "" {
		http.Error(w, "Invalid device UID", http.StatusBadRequest)
		return
	}
	
	// 解析请求体
	var req struct {
		Content  interface{} `json:"content"`  // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
		Duration int         `json:"duration"` // 订阅时长（秒），最大 3600
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 调用命令服务
	if err := h.commandService.SubscribeRealtimeData(r.Context(), uid, req.Content, req.Duration); err != nil {
		h.logger.Error("Failed to subscribe realtime data",
			zap.String("uid", uid),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 200,
		"msg":  "操作成功",
		"data": nil,
	})
}

// SendCommand 发送设备命令（重启等）
// POST /internal/api/v1/radar/devices/{uid}/commands
func (h *CommandHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径提取 UID
	uid := extractUIDFromPath(r.URL.Path, "/internal/api/v1/radar/devices/", "/commands")
	if uid == "" {
		http.Error(w, "Invalid device UID", http.StatusBadRequest)
		return
	}
	
	// 解析请求体
	var req struct {
		Dev int `json:"dev"` // 0-重启雷达和主控，1-只重启雷达，2-只重启主控，100-清除设备数据，101-清除雷达数据，102-清除主控数据
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 调用命令服务
	if err := h.commandService.CallDeviceFunction(r.Context(), uid, req.Dev); err != nil {
		h.logger.Error("Failed to send device command",
			zap.String("uid", uid),
			zap.Int("dev", req.Dev),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 200,
		"msg":  "操作成功",
		"data": nil,
	})
}

// extractUIDFromPath 从 URL 路径中提取 UID
// 例如：/internal/api/v1/radar/devices/25A859B8333B/properties/get
// 返回：25A859B8333B
func extractUIDFromPath(path, prefix, suffix string) string {
	if len(path) < len(prefix)+len(suffix) {
		return ""
	}
	if !startsWith(path, prefix) || !endsWith(path, suffix) {
		return ""
	}
	uid := path[len(prefix) : len(path)-len(suffix)]
	return uid
}

// startsWith 检查字符串是否以指定前缀开始（内部辅助函数）
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

