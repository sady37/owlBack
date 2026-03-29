package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// APNSHandler iOS 设备 Token 注册/注销
// POST /data/api/v1/auth/devices/apns
// DELETE /data/api/v1/auth/devices/apns
type APNSHandler struct {
	apnsSvc *service.APNSDeviceService
	logger  *zap.Logger
}

func NewAPNSHandler(apnsSvc *service.APNSDeviceService, logger *zap.Logger) *APNSHandler {
	return &APNSHandler{apnsSvc: apnsSvc, logger: logger}
}

// RegisterDevice POST /data/api/v1/auth/devices/apns
// iOS App 登录后或 token 刷新时调用
func (h *APNSHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, tenantID, userType, _, ok := service.MustSession(ctx)
	if !ok || userID == "" || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}

	var req struct {
		DeviceToken string `json:"device_token"`
		Environment string `json:"environment"` // "sandbox" | "production"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceToken == "" {
		writeJSON(w, http.StatusBadRequest, Fail("device_token required"))
		return
	}
	if req.Environment == "" {
		req.Environment = "production"
	}

	// 报警推送仅面向 staff，避免住户/联系人接收带来的责任边界问题
	if !strings.EqualFold(userType, "staff") {
		writeJSON(w, http.StatusForbidden, Fail("APNs device registration is only available for staff"))
		return
	}

	if err := h.apnsSvc.Register(ctx, tenantID, userID, userType, req.DeviceToken, req.Environment); err != nil {
		h.logger.Error("[APNS] register device failed",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("register failed"))
		return
	}

	h.logger.Info("[APNS] device registered",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.String("env", req.Environment))
	writeJSON(w, http.StatusOK, Ok(map[string]string{"status": "registered"}))
}

// UnregisterDevice DELETE /data/api/v1/auth/devices/apns
// iOS App 登出时调用
func (h *APNSHandler) UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, tenantID, _, _, ok := service.MustSession(ctx)
	if !ok || userID == "" || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}

	var req struct {
		DeviceToken string `json:"device_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceToken == "" {
		writeJSON(w, http.StatusBadRequest, Fail("device_token required"))
		return
	}

	_ = h.apnsSvc.Unregister(ctx, tenantID, req.DeviceToken)
	writeJSON(w, http.StatusOK, Ok(map[string]string{"status": "unregistered"}))
}