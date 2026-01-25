package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"wisefido-qinglan/internal/service"

	"github.com/gorilla/mux"
)

// APIHandler HTTP API处理器
type APIHandler struct {
	radarService *service.RadarService
}

// NewAPIHandler 创建API处理器
func NewAPIHandler(radarService *service.RadarService) *APIHandler {
	return &APIHandler{
		radarService: radarService,
	}
}

// RegisterRoutes 注册路由
func (h *APIHandler) RegisterRoutes(router *mux.Router) {
	// 设备管理API
	deviceRouter := router.PathPrefix("/api/v1/radar/devices").Subrouter()
	deviceRouter.HandleFunc("/{uid}/properties", h.GetDeviceProperties).Methods("GET")
	deviceRouter.HandleFunc("/{uid}/properties", h.SetDeviceProperties).Methods("PUT")
	deviceRouter.HandleFunc("/{uid}/subscribe", h.SubscribeRealtimeData).Methods("POST")
	deviceRouter.HandleFunc("/{uid}/function", h.CallDeviceFunction).Methods("POST")
	deviceRouter.HandleFunc("/{uid}/info", h.GetDeviceInfo).Methods("GET")
	deviceRouter.HandleFunc("/{uid}/location", h.GetDeviceLocationInfo).Methods("GET")
	
	// 租户设备列表
	router.HandleFunc("/api/v1/tenants/{tenantId}/devices", h.GetDevicesByTenant).Methods("GET")
	
	// 健康检查
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
}

// GetDeviceProperties 读取设备属性
func (h *APIHandler) GetDeviceProperties(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	
	// 解析查询参数
	keysParam := r.URL.Query().Get("keys")
	var keys []string
	if keysParam != "" {
		keys = strings.Split(keysParam, ",")
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	properties, err := h.radarService.GetDeviceProperties(ctx, uid, keys)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get device properties: %v", err))
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    properties,
	})
}

// SetDeviceProperties 设置设备属性
func (h *APIHandler) SetDeviceProperties(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	
	// 解析请求体
	var request struct {
		Properties map[string]interface{} `json:"properties"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	
	if len(request.Properties) == 0 {
		h.sendError(w, http.StatusBadRequest, "Properties cannot be empty")
		return
	}
	
	// 记录 HTTP API 接收到的请求格式（发给 client-qinglan 的格式）
	requestJSON, _ := json.Marshal(request)
	log.Printf("📥 HTTP API SetDeviceProperties: received from client-qinglan")
	log.Printf("   Device UID: %s", uid)
	log.Printf("   HTTP Request Body: %s", string(requestJSON))
	log.Printf("   HTTP Request Format: {\"properties\": {\"key1\": \"value1\", \"key2\": \"value2\", ...}}")
	
	// 检查是否是模式修改（radar_func_ctrl）
	if mode, ok := request.Properties["radar_func_ctrl"]; ok {
		log.Printf("🔄 MODE CHANGE DETECTED: radar_func_ctrl = %v", mode)
		log.Printf("   Mode values: 3=轨迹, 7=呼吸心率, 11=轨迹+呼吸心率, 15=轨迹+呼吸心率+跌倒")
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	if err := h.radarService.SetDeviceProperties(ctx, uid, request.Properties); err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to set device properties: %v", err))
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Device properties set successfully",
	})
}

// SubscribeRealtimeData 订阅实时数据
func (h *APIHandler) SubscribeRealtimeData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	
	// 解析请求体
	var request struct {
		Content  interface{} `json:"content"`  // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
		Duration int         `json:"duration"` // 订阅时长（秒），最大3600
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	
	if request.Duration <= 0 || request.Duration > 3600 {
		h.sendError(w, http.StatusBadRequest, "Duration must be between 1 and 3600 seconds")
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	if err := h.radarService.SubscribeRealtimeData(ctx, uid, request.Content, request.Duration); err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to subscribe realtime data: %v", err))
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Realtime data subscription started",
		"data": map[string]interface{}{
			"uid":      uid,
			"duration": request.Duration,
		},
	})
}

// CallDeviceFunction 调用设备功能
func (h *APIHandler) CallDeviceFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	
	// 解析请求体
	var request struct {
		Dev int `json:"dev"` // 0-重启雷达和主控，1-只重启雷达，2-只重启主控
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	
	if request.Dev < 0 || request.Dev > 2 {
		h.sendError(w, http.StatusBadRequest, "Dev must be 0, 1, or 2")
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	if err := h.radarService.CallDeviceFunction(ctx, uid, request.Dev); err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to call device function: %v", err))
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Device function called successfully",
	})
}

// GetDeviceInfo 获取设备信息
func (h *APIHandler) GetDeviceInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	device, err := h.radarService.GetDeviceInfo(ctx, uid)
	if err != nil {
		h.sendError(w, http.StatusNotFound, fmt.Sprintf("Device not found: %v", err))
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    device,
	})
}

// GetDeviceLocationInfo 获取设备位置信息
func (h *APIHandler) GetDeviceLocationInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	locationInfo, err := h.radarService.GetDeviceLocationInfo(ctx, uid)
	if err != nil {
		h.sendError(w, http.StatusNotFound, fmt.Sprintf("Device location not found: %v", err))
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    locationInfo,
	})
}

// GetDevicesByTenant 根据租户获取设备列表
func (h *APIHandler) GetDevicesByTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	devices, err := h.radarService.GetDevicesByTenant(ctx, tenantID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get devices: %v", err))
		return
	}
	
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    devices,
		"count":   len(devices),
	})
}

// HealthCheck 健康检查
func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "wisefido-qinglan",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// sendJSON 发送JSON响应
func (h *APIHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// sendError 发送错误响应
func (h *APIHandler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}