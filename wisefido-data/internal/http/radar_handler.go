package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"
	"wisefido-data/internal/store"
	"wisefido-data/internal/subscriber"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RadarHandler 雷达设备 API Handler，对接 owlFront radarDeviceApi。
//
// =====================================================================
// 设计：device_addr / device_uid / spatial_prefix 三种 identity 各司其职
// =====================================================================
//
// 三类不同的"被指向物"，三类 endpoint 路径段：
//
//   1. business 层 — 「谁绑这张卡 / 在哪个空间集合」
//       :device_addr   INET /128 host text (e.g., "fd00:0:3:411:2:100:7418:b267")
//       device_addr 随 unit/room/bed 重绑改变 — 这正是想要的：跟"当前位置"对齐
//       endpoint: /device/:device_addr/{card-context, init-subscriptions, stream, realtime}
//
//   2. hardware 层 — vendor MQTT 直接读写设备硬件
//       :device_uid    12 hex MAC (e.g., "E598A2ACD523" — dfm.device_uid)
//       device_uid 是 firmware 烧录的不变量 — qinglan MQTT topic 用它寻址
//       endpoint: /device/:device_uid/{bind, unbind, original-properties, config, control}
//
//   3. spatial 层 — layout container（room_visual_layout PK）
//       :spatial_prefix INET CIDR (/80 unit, /88 room, /128 device)
//       endpoint: /room/:spatial_prefix/layout
//
// 融合点 (canvas 内部)：
//   canvas.objects[] 里的 Device 条目应锚 device_uid（physical install identity），
//   设备移出此 spatial_prefix 时 layout 残留 ghost 引用，由 read 时 lazy 过滤
//   (device.device_addr <<= spatial_prefix 不成立 = ghost)。
//   注：当前 radar_install_service.injectDeviceBindingsIntoLayout 仍注入 device_addr，
//   待该层 follow-up 改为注入 device_uid + lazy filter；本 handler 已按目标 contract 命名。
//
// 见 memory feedback_api_ids_ipv6_only / layout_scope_by_entry_point。
// =====================================================================
type RadarHandler struct {
	radarInstall         *service.RadarInstall
	stubHandler          *StubHandler
	kv                   store.KV
	redisClient          *redis.Client
	dataStreamSubscriber interface{} // DataStreamSubscriber
	logger               *zap.Logger
}

func NewRadarHandler(radarInstall *service.RadarInstall, stubHandler *StubHandler, kv store.KV, redisClient *redis.Client, logger *zap.Logger) *RadarHandler {
	return &RadarHandler{
		radarInstall: radarInstall,
		stubHandler:  stubHandler,
		kv:           kv,
		redisClient:  redisClient,
		logger:       logger,
	}
}

func (h *RadarHandler) SetDataStreamSubscriber(subscriber interface{}) {
	h.dataStreamSubscriber = subscriber
}

// =====================================================================
// path identifier parsers — 按语义拆 INET vs MAC vs CIDR
// =====================================================================

// parseRadarDeviceAddr 校验 path :id 段为 device_addr (IPv6 host 或 /128 CIDR)；
// 返 canonical /128 CIDR (与其他 spatial endpoint 一致)。
func parseRadarDeviceAddr(raw string) (string, bool) {
	return parseDeviceAddrFromPath(raw) // 复用 device_addr_helper.go
}

// parseRadarDeviceUID 校验 path :id 段为 12 hex char MAC (dfm.device_uid 格式)；返 upper-case canonical。
func parseRadarDeviceUID(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" || s == "undefined" || s == "null" {
		return "", false
	}
	if len(s) != 12 {
		return "", false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return "", false
		}
	}
	return strings.ToUpper(s), true
}

// extractRadarPathSeg 纯 path slicing（prefix+suffix 包夹的中间段），类型校验由 caller 用 parseRadar* 做。
func extractRadarPathSeg(path, prefix, suffix string) string {
	if len(path) < len(prefix)+len(suffix) {
		return ""
	}
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	return path[len(prefix) : len(path)-len(suffix)]
}

// extractRadarCardIDFromPath 从 /card/{cardId}/devices 中提取 cardId
func extractRadarCardIDFromPath(path, prefix, suffix string) string {
	return extractRadarPathSeg(path, prefix, suffix)
}

// extractRadarRoomIDFromPath 从 /room/{roomId}/layout 中提取 roomId
func extractRadarRoomIDFromPath(path, prefix, suffix string) string {
	return extractRadarPathSeg(path, prefix, suffix)
}

// tenantIDFromReq 从 X-Tenant-Id header 获取 tenant_id（AuthMiddleware 已注入）
func (h *RadarHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenantID := r.Header.Get("X-Tenant-Id")
	if tenantID == "" || tenantID == "null" {
		writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
		return "", false
	}
	return tenantID, true
}

// =====================================================================
// 业务层 endpoint — 路径段 = device_addr (INET)
// =====================================================================

// GetCardDevices 获取卡片上所有设备列表
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
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		uid := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyCardInScope(h.stubHandler.DB, r.Context(), uid, role, cardID); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	list, err := h.radarInstall.ListCardDevices(r.Context(), tenantID, cardID)
	if err != nil {
		h.logger.Error("GetCardDevices", zap.String("card_id", cardID), zap.String("tenant_id", tenantID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(list))
}

// GetCardContext 通过 device_addr 查所属卡片 + spatial 元信息 + layout config (纯读，无 side effect)。
// GET /radar-device/api/v1/radar-device/device/:device_addr/card-context
//
// 返回：{ card_id, room_id (/88 fallback 锚), spatial_prefix (本 canvas /128 save target), devices[], layout_config }
//
// 历史路径 /card-devices 已重命名为 /card-context — 旧 endpoint 的 init-subscriptions side effect
// 拆到独立 POST /init-subscriptions 端点（HTTP 语义合规）。
func (h *RadarHandler) GetCardContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawAddr := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/card-context")
	deviceAddr, ok := parseRadarDeviceAddr(rawAddr)
	if !ok {
		http.Error(w, "Invalid device_addr (expect IPv6 host or /128 CIDR)", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		uid := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), uid, role, deviceAddr); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	cardID, roomID, spatialPrefix, list, layoutConfig, err := h.radarInstall.ListCardDevicesByDeviceAddr(r.Context(), tenantID, deviceAddr)
	if err != nil {
		h.logger.Error("GetCardContext", zap.String("device_addr", deviceAddr), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(struct {
		CardID        string                      `json:"card_id"`
		RoomID        string                      `json:"room_id"`
		SpatialPrefix string                      `json:"spatial_prefix"`
		Devices       []repository.CardDeviceItem `json:"devices"`
		LayoutConfig  json.RawMessage             `json:"layout_config"`
	}{CardID: cardID, RoomID: roomID, SpatialPrefix: spatialPrefix, Devices: list, LayoutConfig: layoutConfig}))
}

// InitSubscriptionsFromLayout 从 layout 配置自动订阅已绑定设备 (side effect endpoint)。
// POST /radar-device/api/v1/radar-device/device/:device_addr/init-subscriptions
//
// FE 进 wave 页时显式调用一次；从 layout.canvas 解析 device 列表并触发 SubscribeRealtimeData。
// 拆出独立 POST 避免在 GET /card-context 里搞 side effect (HTTP 语义合规)。
func (h *RadarHandler) InitSubscriptionsFromLayout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawAddr := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/init-subscriptions")
	deviceAddr, ok := parseRadarDeviceAddr(rawAddr)
	if !ok {
		http.Error(w, "Invalid device_addr", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		uid := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), uid, role, deviceAddr); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	_, _, _, _, layoutConfig, err := h.radarInstall.ListCardDevicesByDeviceAddr(r.Context(), tenantID, deviceAddr)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if len(layoutConfig) == 0 {
		writeJSON(w, http.StatusOK, Ok(map[string]any{"initialized": 0, "reason": "empty layout"}))
		return
	}
	h.logger.Info("[RADAR_INIT_API] initializing subscriptions from layout",
		zap.String("device_addr", deviceAddr), zap.String("tenant_id", tenantID))
	if err := h.radarInstall.InitializeSubscriptionsFromLayout(r.Context(), tenantID, layoutConfig); err != nil {
		h.logger.Warn("[RADAR_INIT_API] failed to initialize subscriptions",
			zap.String("device_addr", deviceAddr), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"initialized": 1}))
}

// SubscribeRealtimeStream SSE 实时数据流。
// GET /radar-device/api/v1/radar-device/device/:device_addr/stream
func (h *RadarHandler) SubscribeRealtimeStream(w http.ResponseWriter, r *http.Request) {
	// 必须最先设 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Method not allowed\"}\n\n")
		return
	}

	rawAddr := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/stream")
	deviceAddr, ok := parseRadarDeviceAddr(rawAddr)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Invalid device_addr\"}\n\n")
		return
	}

	tenantID := r.Header.Get("X-Tenant-Id")
	if tenantID == "" || tenantID == "null" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "event: error\ndata: {\"error\":\"tenant_id is required\"}\n\n")
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		uid := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), uid, role, deviceAddr); err != nil {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "event: error\ndata: {\"error\":\""+err.Error()+"\"}\n\n")
			return
		}
	}

	h.logger.Info("[RADAR_STREAM_SSE] connection attempt",
		zap.String("device_addr", deviceAddr), zap.String("tenant_id", tenantID))

	if h.redisClient == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Redis client not available\"}\n\n")
		return
	}

	ctx := r.Context()

	cardID, _, _, _, _, err := h.radarInstall.ListCardDevicesByDeviceAddr(ctx, tenantID, deviceAddr)
	if err != nil {
		h.logger.Error("[RADAR_STREAM_SSE] failed to get card_id",
			zap.String("device_addr", deviceAddr), zap.Error(err))
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Failed to get card info\"}\n\n")
		return
	}

	sub, ok := h.dataStreamSubscriber.(*subscriber.DataStreamSubscriber)
	if !ok || sub == nil {
		h.logger.Error("[RADAR_STREAM_SSE] DataStreamSubscriber not available",
			zap.String("device_addr", deviceAddr), zap.String("card_id", cardID))
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"DataStreamSubscriber not available\"}\n\n")
		return
	}

	cardRealtimeUpdatedChan := sub.GetCardRealtimeUpdatedChan()

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Streaming not supported\"}\n\n")
		return
	}

	io.WriteString(w, ": SSE connection established\n")
	io.WriteString(w, "event: connected\ndata: {\"device_addr\":\""+deviceAddr+"\",\"status\":\"connected\"}\n\n")
	flusher.Flush()

	h.logger.Info("[RADAR_STREAM_SSE] connection established",
		zap.String("device_addr", deviceAddr), zap.String("tenant_id", tenantID), zap.String("card_id", cardID))

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				io.WriteString(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("[RADAR_STREAM_SSE] connection closed",
				zap.String("device_addr", deviceAddr), zap.String("card_id", cardID))
			return
		case notifiedCardID := <-cardRealtimeUpdatedChan:
			if notifiedCardID != cardID {
				continue
			}
			cachedData := sub.GetCardRealtimeData(cardID)
			if cachedData == nil {
				continue
			}
			jsonData, err := json.Marshal(cachedData)
			if err != nil {
				h.logger.Warn("[RADAR_STREAM_SSE] failed to marshal", zap.Error(err))
				continue
			}
			io.WriteString(w, "data: "+string(jsonData)+"\n\n")
			flusher.Flush()
		}
	}
}

// =====================================================================
// spatial 层 endpoint — 路径段 = spatial_prefix (CIDR)
// =====================================================================

// PutRoomLayout 保存房间布局到 room_visual_layout (spatial_prefix 三档作用域)。
// PUT /radar-device/api/v1/radar-device/room/:roomId/layout
//
// 三档作用域优先级：
//  1. body.spatial_prefix 字段 (FE 显式 CIDR；/80 unit 与 /128 device 明确指定)
//  2. URL path roomID (host-only IPv6；backend 按 host bits 推断 /88 vs /128)
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
	saveTarget := roomID
	var bodyPeek struct {
		SpatialPrefix string `json:"spatial_prefix"`
	}
	if jerr := json.Unmarshal(body, &bodyPeek); jerr == nil && bodyPeek.SpatialPrefix != "" {
		saveTarget = bodyPeek.SpatialPrefix
	}
	if err := h.radarInstall.SaveRoomLayout(r.Context(), tenantID, saveTarget, body); err != nil {
		h.logger.Error("PutRoomLayout", zap.String("save_target", saveTarget), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok("ok"))
}

// =====================================================================
// hardware 层 endpoint — 路径段 = device_uid (MAC)
// =====================================================================

// BindDevice 标记订阅 radar 数据（vendor MQTT subscribe）。
// POST /radar-device/api/v1/radar-device/device/:device_uid/bind
func (h *RadarHandler) BindDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawUID := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/bind")
	deviceUID, ok := parseRadarDeviceUID(rawUID)
	if !ok {
		http.Error(w, "Invalid device_uid (expect 12 hex chars)", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	h.logger.Info("[RADAR_BIND_API] bind",
		zap.String("device_uid", deviceUID), zap.String("tenant_id", tenantID))
	if err := h.radarInstall.BindDevice(r.Context(), tenantID, deviceUID); err != nil {
		h.logger.Error("[RADAR_BIND_API] bind failed",
			zap.String("device_uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok("ok"))
}

// UnbindDevice 取消订阅 radar 数据。
// POST /radar-device/api/v1/radar-device/device/:device_uid/unbind
func (h *RadarHandler) UnbindDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawUID := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/unbind")
	deviceUID, ok := parseRadarDeviceUID(rawUID)
	if !ok {
		http.Error(w, "Invalid device_uid", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	h.logger.Info("[RADAR_UNBIND_API] unbind",
		zap.String("device_uid", deviceUID), zap.String("tenant_id", tenantID))
	if err := h.radarInstall.UnbindDevice(r.Context(), tenantID, deviceUID); err != nil {
		h.logger.Error("[RADAR_UNBIND_API] unbind failed",
			zap.String("device_uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok("ok"))
}

// GetOriginalProperties 获取设备原始属性 (vendor MQTT read)。
// GET /radar-device/api/v1/radar-device/device/:device_uid/original-properties[?source=db&keys=...]
func (h *RadarHandler) GetOriginalProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawUID := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/original-properties")
	deviceUID, ok := parseRadarDeviceUID(rawUID)
	if !ok {
		http.Error(w, "Invalid device_uid", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	device, err := h.radarInstall.GetDeviceByUID(r.Context(), tenantID, deviceUID)
	if err != nil {
		h.logger.Warn("GetOriginalProperties: device not found", zap.String("uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		callerUID := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), callerUID, role, device.DeviceAddr); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	if strings.TrimSpace(r.URL.Query().Get("source")) == "db" {
		propertiesJSON, err := h.radarInstall.GetOriginalPropertiesFromDB(r.Context(), tenantID, device.DeviceAddr)
		if err != nil {
			h.logger.Error("GetOriginalPropertiesFromDB", zap.String("uid", deviceUID), zap.Error(err))
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
	propertiesJSON, err := h.radarInstall.GetOriginalProperties(r.Context(), device.DeviceAddr, deviceUID, keys)
	if err != nil {
		h.logger.Error("GetOriginalProperties failed",
			zap.String("uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(propertiesJSON))
}

// UpdateConfig 更新设备安装配置 (vendor MQTT write)。
// PUT /radar-device/api/v1/radar-device/device/:device_uid/config
// Body: v1 格式 { install_model, height, boundary_left/right/front/rear }
func (h *RadarHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawUID := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/config")
	deviceUID, ok := parseRadarDeviceUID(rawUID)
	if !ok {
		http.Error(w, "Invalid device_uid", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	device, err := h.radarInstall.GetDeviceByUID(r.Context(), tenantID, deviceUID)
	if err != nil {
		h.logger.Warn("UpdateConfig: device not found", zap.String("uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		callerUID := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), callerUID, role, device.DeviceAddr); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	}
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if config == nil {
		config = make(map[string]interface{})
	}
	// 兼容 body 被包成 { "data": { ... } } 的情况
	if len(config) == 1 {
		if dataVal, ok := config["data"]; ok {
			if dataMap, ok := dataVal.(map[string]interface{}); ok {
				config = dataMap
			}
		}
	}
	configKeys := make([]string, 0, len(config))
	for k := range config {
		configKeys = append(configKeys, k)
	}
	logFields := []zap.Field{zap.String("uid", deviceUID), zap.Int("keys", len(config)), zap.Strings("keys_list", configKeys)}
	if v := config["declare_area"]; v != nil {
		logFields = append(logFields, zap.Any("declare_area", v))
	}
	if v := config["rectangle"]; v != nil {
		logFields = append(logFields, zap.Any("rectangle", v))
	}
	h.logger.Info("UpdateConfig", logFields...)
	deviceCode, err := h.radarInstall.UpdateConfig(r.Context(), device.DeviceAddr, deviceUID, config)
	if err != nil {
		h.logger.Error("UpdateConfig failed",
			zap.String("uid", deviceUID), zap.Int("device_code", deviceCode), zap.Error(err))
		writeJSON(w, http.StatusOK, Result[any]{Code: ResultError, Type: "error", Message: err.Error(), Result: map[string]interface{}{"device_code": deviceCode}})
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{"device_code": deviceCode}))
}

// Control 设备重启/控制 (vendor MQTT command)。
// POST /radar-device/api/v1/radar-device/device/:device_uid/control
// Body: { "dev": 0|1|2|100|101|102 }
//   0=雷达+主控  1=仅雷达  2=仅主控  100/101/102=清除数据
func (h *RadarHandler) Control(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rawUID := extractRadarPathSeg(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/control")
	deviceUID, ok := parseRadarDeviceUID(rawUID)
	if !ok {
		http.Error(w, "Invalid device_uid", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	device, err := h.radarInstall.GetDeviceByUID(r.Context(), tenantID, deviceUID)
	if err != nil {
		h.logger.Warn("Control: device not found", zap.String("uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	if h.stubHandler != nil && h.stubHandler.DB != nil {
		callerUID := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		if err := VerifyDeviceInScope(h.stubHandler.DB, r.Context(), callerUID, role, device.DeviceAddr); err != nil {
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
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
	if err := h.radarInstall.CallDeviceFunction(r.Context(), deviceUID, dev); err != nil {
		h.logger.Error("Control failed", zap.String("uid", deviceUID), zap.Int("dev", dev), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok([]string{}))
}

