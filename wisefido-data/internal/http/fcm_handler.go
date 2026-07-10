package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// FCMHandler Android FCM 设备 Token 注册/注销
// POST   /data/api/v1/auth/devices/fcm
// DELETE /data/api/v1/auth/devices/fcm
// 复用 APNSDeviceService.Register/Unregister，写入口与 iOS 同源（apns_devices 表）。
type FCMHandler struct {
	svc    *service.APNSDeviceService
	logger *zap.Logger
}

func NewFCMHandler(svc *service.APNSDeviceService, logger *zap.Logger) *FCMHandler {
	return &FCMHandler{svc: svc, logger: logger}
}

// RegisterDevice POST /data/api/v1/auth/devices/fcm
// Android App 登录后或 token 刷新时调用；platform 固定 android，environment 固定 production。
func (h *FCMHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, tenantID, userType, _, ok := service.MustSession(ctx)
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

	if !strings.EqualFold(userType, "staff") {
		writeJSON(w, http.StatusForbidden, Fail("FCM device registration is only available for staff"))
		return
	}

	if err := h.svc.Register(ctx, userID, req.DeviceToken, "android", "production"); err != nil {
		h.logger.Error("[FCM] register device failed",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		writeJSON(w, http.StatusOK, Fail("register failed"))
		return
	}

	h.logger.Info("[FCM] device registered",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))
	writeJSON(w, http.StatusOK, Ok(map[string]string{"status": "registered"}))
}

// UnregisterDevice DELETE /data/api/v1/auth/devices/fcm
// Android App 登出时调用
func (h *FCMHandler) UnregisterDevice(w http.ResponseWriter, r *http.Request) {
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

	_ = h.svc.Unregister(ctx, userID, req.DeviceToken)
	writeJSON(w, http.StatusOK, Ok(map[string]string{"status": "unregistered"}))
}
