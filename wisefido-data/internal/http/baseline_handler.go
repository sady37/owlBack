package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"owl-common/card"

	"go.uber.org/zap"
)

type BaselineHandler struct {
	cardDB *card.CardDB
	logger *zap.Logger
}

func NewBaselineHandler(cardDB *card.CardDB, logger *zap.Logger) *BaselineHandler {
	return &BaselineHandler{cardDB: cardDB, logger: logger}
}

// GetDeviceBaseline handles GET /internal/device-baseline/{deviceKey}
func (h *BaselineHandler) GetDeviceBaseline(w http.ResponseWriter, r *http.Request) {
	deviceKey := strings.TrimPrefix(r.URL.Path, "/internal/device-baseline/")
	deviceKey = strings.TrimSpace(deviceKey)
	if deviceKey == "" {
		http.Error(w, "missing device key", http.StatusBadRequest)
		return
	}
	b, ok := h.cardDB.ResolveDeviceBaseline(r.Context(), deviceKey)
	if !ok {
		http.Error(w, "device not found", http.StatusNotFound)
		return
	}
	if strings.TrimSpace(b.CardID) == "" {
		b.CardID = strings.TrimSpace(b.DeviceID)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

// ListDeviceBaselines handles GET /internal/device-baselines?device_type=Sleepad,Radar
func (h *BaselineHandler) ListDeviceBaselines(w http.ResponseWriter, r *http.Request) {
	dtParam := r.URL.Query().Get("device_type")
	var deviceTypes []string
	if dtParam != "" && dtParam != "all" {
		for _, dt := range strings.Split(dtParam, ",") {
			if t := strings.TrimSpace(dt); t != "" {
				deviceTypes = append(deviceTypes, t)
			}
		}
	}
	baselines, err := h.cardDB.ListAllBaselines(r.Context(), deviceTypes)
	if err != nil {
		h.logger.Error("list baselines failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	for i := range baselines {
		if strings.TrimSpace(baselines[i].CardID) == "" {
			baselines[i].CardID = strings.TrimSpace(baselines[i].DeviceID)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(baselines)
}
