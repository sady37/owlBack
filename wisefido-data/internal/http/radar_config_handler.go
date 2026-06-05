package httpapi

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// GetRadarConfig 内部端点：sensor 按 device_addr 取 radar 配置（read-through：库空则现读固件落库）。
// GET /internal/radar-config/{device_addr}  — /internal/ 前缀跳过 auth（服务间内网调用）。
func (h *RadarHandler) GetRadarConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	raw := strings.TrimPrefix(r.URL.Path, "/internal/radar-config/")
	deviceAddr, ok := parseDeviceAddrFromPath(raw)
	if !ok {
		http.Error(w, "invalid device_addr", http.StatusBadRequest)
		return
	}
	snap, err := h.radarInstall.GetRadarConfigSnapshot(r.Context(), deviceAddr)
	if err != nil {
		h.logger.Error("GetRadarConfig", zap.String("device_addr", deviceAddr), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}
