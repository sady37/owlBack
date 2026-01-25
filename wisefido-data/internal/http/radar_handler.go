package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// RadarHandler 雷达设备 API Handler
// 对接 owlFront radarDeviceApi，路径 :id 为 device_id（devices.device_id），
// 内部通过 radar_install_service 解析 device_uid 后调 qinglan_client → wisefido-qinglan。
// 路由：/radar-device/api/v1/radar-device/device/:id/{realtime|original-properties|config}
type RadarHandler struct {
	radarInstall *service.RadarInstall
	stubHandler  *StubHandler
	logger       *zap.Logger
}

// NewRadarHandler 创建 RadarHandler
func NewRadarHandler(radarInstall *service.RadarInstall, stubHandler *StubHandler, logger *zap.Logger) *RadarHandler {
	return &RadarHandler{
		radarInstall: radarInstall,
		stubHandler:  stubHandler,
		logger:       logger,
	}
}

// GetCardDevices 获取卡片上所有设备（device_id, device_type, device_uid, device_name），供 vue-radar 画布 Bind 使用
// GET /radar-device/api/v1/radar-device/card/:cardId/devices
func (h *RadarHandler) GetCardDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cardID := extractRadarCardIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/card/", "/devices")
	if cardID == "" || strings.Contains(cardID, "/") {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	list, err := h.radarInstall.ListCardDevices(r.Context(), tenantID, cardID)
	if err != nil {
		h.logger.Error("GetCardDevices", zap.String("card_id", cardID), zap.String("tenant_id", tenantID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(list))
}

// extractRadarCardIDFromPath 从 /radar-device/.../card/{cardId}/devices 中提取 cardId
func extractRadarCardIDFromPath(path, prefix, suffix string) string {
	if len(path) < len(prefix)+len(suffix) {
		return ""
	}
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	return path[len(prefix) : len(path)-len(suffix)]
}

// GetCardDevicesByDeviceID 通过 device_id 查所属卡片并返回该卡设备列表
// GET /radar-device/api/v1/radar-device/device/:deviceId/card-devices
func (h *RadarHandler) GetCardDevicesByDeviceID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/card-devices")
	if !validateRadarDeviceID(deviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	cardID, roomID, list, layoutConfig, err := h.radarInstall.ListCardDevicesByDeviceID(r.Context(), tenantID, deviceID)
	if err != nil {
		h.logger.Error("GetCardDevicesByDeviceID", zap.String("device_id", deviceID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(struct {
		CardID       string                      `json:"card_id"`
		RoomID       string                      `json:"room_id"`
		Devices      []repository.CardDeviceItem `json:"devices"`
		LayoutConfig json.RawMessage             `json:"layout_config"`
	}{CardID: cardID, RoomID: roomID, Devices: list, LayoutConfig: layoutConfig}))
}

// PutRoomLayout 保存房间布局到 config_versions（config_type=room_layout, entity_id=room_id）。Body 为 config_data JSON。
// PUT /radar-device/api/v1/radar-device/room/:roomId/layout
func (h *RadarHandler) PutRoomLayout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	roomID := extractRadarRoomIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/room/", "/layout")
	if roomID == "" || strings.Contains(roomID, "/") {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid or empty body", http.StatusBadRequest)
		return
	}
	if err := h.radarInstall.SaveRoomLayout(r.Context(), tenantID, roomID, body); err != nil {
		h.logger.Error("PutRoomLayout", zap.String("room_id", roomID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok("ok"))
}

func extractRadarRoomIDFromPath(path, prefix, suffix string) string {
	if len(path) < len(prefix)+len(suffix) {
		return ""
	}
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	return path[len(prefix) : len(path)-len(suffix)]
}

// GetRealtimeData 获取实时数据
// GET /radar-device/api/v1/radar-device/device/:id/realtime
// TODO: 从 Redis Streams / 订阅 qinglan 实时流 获取 positions、vital；当前返回空。
func (h *RadarHandler) GetRealtimeData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/realtime")
	if !validateRadarDeviceID(deviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	if _, ok := h.tenantIDFromReq(w, r); !ok {
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"positions": []interface{}{},
		"vital":     nil,
	}))
}

// GetOriginalProperties 获取设备原始属性
// GET /radar-device/api/v1/radar-device/device/:id/original-properties
// 返回 qinglan 原始属性 JSON（radar_install_style, rectangle 等），供 radarDeviceApi.devicePropsToMqttReadFormat 转换。
// 查询参数 source=db：从 DB 读（设备查失败时的回退），当前无持久化则返回 {}
func (h *RadarHandler) GetOriginalProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/original-properties")
	if !validateRadarDeviceID(deviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if strings.TrimSpace(r.URL.Query().Get("source")) == "db" {
		propertiesJSON, err := h.radarInstall.GetOriginalPropertiesFromDB(r.Context(), tenantID, deviceID)
		if err != nil {
			h.logger.Error("GetOriginalPropertiesFromDB", zap.String("device_id", deviceID), zap.Error(err))
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(propertiesJSON))
		return
	}
	var keys []string
	if q := strings.TrimSpace(r.URL.Query().Get("keys")); q != "" {
		for _, k := range strings.Split(q, ",") {
			if t := strings.TrimSpace(k); t != "" {
				keys = append(keys, t)
			}
		}
	}
	propertiesJSON, err := h.radarInstall.GetOriginalProperties(r.Context(), tenantID, deviceID, keys)
	if err != nil {
		h.logger.Error("Failed to get device original properties",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(propertiesJSON))
}

// UpdateConfig 更新设备安装配置
// PUT /radar-device/api/v1/radar-device/device/:id/config
// Body: v1 格式 { install_model, height, boundary_left/right/front/rear }，由 radar_install_service.V1ConfigToRadarDeviceProps 转成 qinglan 属性后写入。
func (h *RadarHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/config")
	if !validateRadarDeviceID(deviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if config == nil {
		config = make(map[string]interface{})
	}
	if err := h.radarInstall.UpdateConfig(r.Context(), tenantID, deviceID, config); err != nil {
		h.logger.Error("Failed to update device config",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok([]string{}))
}

// Control 3.8.1 重启/控制：POST /radar-device/api/v1/radar-device/device/:id/control
// Body: { "dev": 0|1|2|100|101|102 }，0=雷达+主控 1=仅雷达 2=仅主控 100/101/102=清除数据
func (h *RadarHandler) Control(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/control")
	if !validateRadarDeviceID(deviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	var body struct {
		Dev *int `json:"dev"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Dev == nil {
		http.Error(w, "Invalid request body, require { \"dev\": number }", http.StatusBadRequest)
		return
	}
	dev := *body.Dev
	valid := map[int]bool{0: true, 1: true, 2: true, 100: true, 101: true, 102: true}
	if !valid[dev] {
		http.Error(w, "Invalid dev, must be 0|1|2|100|101|102", http.StatusBadRequest)
		return
	}
	uid, err := h.radarInstall.GetDeviceUID(r.Context(), tenantID, deviceID)
	if err != nil {
		h.logger.Error("Control: get device UID", zap.String("device_id", deviceID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if err := h.radarInstall.CallDeviceFunction(r.Context(), uid, dev); err != nil {
		h.logger.Error("Control: call device function", zap.String("device_id", deviceID), zap.Int("dev", dev), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok([]string{}))
}

// extractRadarDeviceIDFromPath 从 URL 路径中提取 device_id（Radar 专用）
// 例：/radar-device/api/v1/radar-device/device/{id}/config → {id}
func extractRadarDeviceIDFromPath(path, prefix, suffix string) string {
	if len(path) < len(prefix)+len(suffix) {
		return ""
	}
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	deviceID := path[len(prefix) : len(path)-len(suffix)]
	if deviceID == "" || strings.Contains(deviceID, "/") {
		return ""
	}
	return deviceID
}

// validateRadarDeviceID 校验 device_id 有效，与 devicesRepo.GetDevice 约定一致
func validateRadarDeviceID(deviceID string) bool {
	return deviceID != "" && deviceID != "undefined" && deviceID != "null"
}

// tenantIDFromReq 从 X-Tenant-Id 或 stubHandler 获取 tenant_id
func (h *RadarHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.stubHandler != nil {
		return h.stubHandler.tenantIDFromReq(w, r)
	}
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

