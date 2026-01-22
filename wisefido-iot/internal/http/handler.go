package http

import (
	"encoding/json"
	"net/http"

	"wisefido-iot/internal/repository"

	"go.uber.org/zap"
)

// Handler HTTP 处理器
type Handler struct {
	iotRepo *repository.IoTTimeSeriesRepository
	logger  *zap.Logger
}

// NewHandler 创建 HTTP 处理器
func NewHandler(iotRepo *repository.IoTTimeSeriesRepository, logger *zap.Logger) *Handler {
	return &Handler{
		iotRepo: iotRepo,
		logger:  logger,
	}
}

// InvalidateLocationCacheRequest 清除位置信息缓存请求
type InvalidateLocationCacheRequest struct {
	DeviceID string `json:"device_id"`
}

// InvalidateLocationCacheResponse 清除位置信息缓存响应
type InvalidateLocationCacheResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// InvalidateLocationCache 清除设备位置信息缓存
// POST /internal/api/v1/iot-timeseries/cache/invalidate
func (h *Handler) InvalidateLocationCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InvalidateLocationCacheRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode request",
			zap.Error(err),
		)
		writeJSON(w, http.StatusBadRequest, InvalidateLocationCacheResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.DeviceID == "" {
		writeJSON(w, http.StatusBadRequest, InvalidateLocationCacheResponse{
			Success: false,
			Message: "device_id is required",
		})
		return
	}

	// 清除位置信息缓存
	h.iotRepo.InvalidateLocationCache(req.DeviceID)

	h.logger.Info("Location cache invalidated",
		zap.String("device_id", req.DeviceID),
	)

	writeJSON(w, http.StatusOK, InvalidateLocationCacheResponse{
		Success: true,
		Message: "Location cache invalidated successfully",
	})
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
