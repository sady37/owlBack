package http

import (
	"encoding/json"
	"net/http"
	"wisefido-radar/internal/models"
	
	"go.uber.org/zap"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// ServeHTTP 处理 HTTP 请求
// 参考 radar-ql-v3/simple-https.py 的 do_POST 方法
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 只处理 POST 请求
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	// 读取请求体
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode request body",
			zap.Error(err),
			zap.String("remote_addr", r.RemoteAddr),
		)
		h.sendErrorResponse(w, "请求格式错误", 400)
		return
	}
	
	// 记录请求日志（参考 simple-https.py 的日志输出）
	h.logger.Info("=== Request Received ===",
		zap.String("uid", req.UID),
		zap.Int("type", req.Type),
		zap.String("mcu_hw", req.MCU.HW),
		zap.String("mcu_sw", req.MCU.SW),
		zap.String("radar_hw", req.Radar.HW),
		zap.String("radar_sw", req.Radar.SW),
		zap.String("remote_addr", r.RemoteAddr),
	)
	
	// 调用认证服务
	response, err := h.authService.AuthenticateDevice(r.Context(), &req)
	if err != nil {
		h.logger.Error("Authentication service error",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
		h.sendErrorResponse(w, "服务器内部错误", 500)
		return
	}
	
	// 记录响应日志（参考 simple-https.py 的日志输出）
	h.logger.Info("=== Response Sent ===",
		zap.String("uid", req.UID),
		zap.Int("code", response.Code),
		zap.String("msg", response.Msg),
	)
	
	// 发送响应
	h.sendJSONResponse(w, response, http.StatusOK)
}

// sendJSONResponse 发送 JSON 响应
func (h *AuthHandler) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response",
			zap.Error(err),
		)
	}
}

// sendErrorResponse 发送错误响应
func (h *AuthHandler) sendErrorResponse(w http.ResponseWriter, msg string, code int) {
	response := &models.AuthResponse{
		Msg:  msg,
		Code: code,
		Data: nil,
	}
	h.sendJSONResponse(w, response, code)
}

