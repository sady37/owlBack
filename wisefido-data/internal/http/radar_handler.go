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

// RadarHandler 雷达设备 API Handler
// 对接 owlFront radarDeviceApi，路径 :id 为 device_id（devices.device_id），
// 内部通过 radar_install_service 解析 device_uid 后调 qinglan_client → wisefido-qinglan。
// 路由：/radar-device/api/v1/radar-device/device/:id/{realtime|original-properties|config}
type RadarHandler struct {
	radarInstall         *service.RadarInstall
	stubHandler          *StubHandler
	kv                   store.KV
	redisClient          *redis.Client
	dataStreamSubscriber interface{} // DataStreamSubscriber
	logger               *zap.Logger
}

// NewRadarHandler 创建 RadarHandler
func NewRadarHandler(radarInstall *service.RadarInstall, stubHandler *StubHandler, kv store.KV, redisClient *redis.Client, logger *zap.Logger) *RadarHandler {
	return &RadarHandler{
		radarInstall: radarInstall,
		stubHandler:  stubHandler,
		kv:           kv,
		redisClient:  redisClient,
		logger:       logger,
	}
}

// SetDataStreamSubscriber 设置数据流订阅器
func (h *RadarHandler) SetDataStreamSubscriber(subscriber interface{}) {
	h.dataStreamSubscriber = subscriber
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

	// 初始化时，从 layout 配置自动订阅已绑定的设备
	if len(layoutConfig) > 0 {
		h.logger.Info("[RADAR_INIT_API] initializing subscriptions from layout",
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.String("room_id", roomID))
		if err := h.radarInstall.InitializeSubscriptionsFromLayout(r.Context(), tenantID, layoutConfig); err != nil {
			h.logger.Warn("[RADAR_INIT_API] failed to initialize subscriptions from layout",
				zap.String("device_id", deviceID),
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			// 不阻塞返回，继续返回数据
		}
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

// BindDevice 绑定设备（通知需要订阅该设备的数据）
// POST /radar-device/api/v1/radar-device/device/:id/bind
func (h *RadarHandler) BindDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/bind")
	if !validateRadarDeviceID(deviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	h.logger.Info("[RADAR_BIND_API] bind request received",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))

	if err := h.radarInstall.BindDevice(r.Context(), tenantID, deviceID); err != nil {
		h.logger.Error("[RADAR_BIND_API] bind failed",
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	h.logger.Info("[RADAR_BIND_API] bind success",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))
	writeJSON(w, http.StatusOK, Ok("ok"))
}

// UnbindDevice 解绑设备（取消订阅该设备的数据）
// POST /radar-device/api/v1/radar-device/device/:id/unbind
func (h *RadarHandler) UnbindDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/unbind")
	if !validateRadarDeviceID(deviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	h.logger.Info("[RADAR_UNBIND_API] unbind request received",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))

	if err := h.radarInstall.UnbindDevice(r.Context(), tenantID, deviceID); err != nil {
		h.logger.Error("[RADAR_UNBIND_API] unbind failed",
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	h.logger.Info("[RADAR_UNBIND_API] unbind success",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))
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

// SubscribeRealtimeStream SSE 订阅实时数据流
// GET /radar-device/api/v1/radar-device/device/:id/stream
// 使用 Server-Sent Events (SSE) 实时推送 Redis stream 数据
func (h *RadarHandler) SubscribeRealtimeStream(w http.ResponseWriter, r *http.Request) {
	// 必须在函数最开始就设置 SSE 响应头，防止任何后续代码覆盖
	// 注意：一旦调用了 WriteHeader，响应头就不能再修改
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")          // 禁用 nginx 缓冲
	w.Header().Set("Access-Control-Allow-Origin", "*") // 允许 CORS（如果需要）

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Method not allowed\"}\n\n")
		return
	}

	deviceID := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/stream")
	if !validateRadarDeviceID(deviceID) {
		// 使用 SSE 格式返回错误
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Invalid device ID\"}\n\n")
		return
	}

	// AuthMiddleware 已经验证了 token 并注入了 X-Tenant-Id header
	tenantID := r.Header.Get("X-Tenant-Id")
	if tenantID == "" || tenantID == "null" {
		// 使用 SSE 格式返回错误
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "event: error\ndata: {\"error\":\"tenant_id is required\"}\n\n")
		return
	}

	// 记录连接尝试
	h.logger.Info("[RADAR_STREAM_SSE] connection attempt",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	if h.redisClient == nil {
		// 使用 SSE 格式返回错误
		h.logger.Error("[RADAR_STREAM_SSE] Redis client not available",
			zap.String("device_id", deviceID))
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Redis client not available\"}\n\n")
		return
	}

	ctx := r.Context()

	// 转换 deviceID：如果是 device_uid，转换为 device_id
	// device_id 是 UUID 格式，device_uid 是短字符串
	// ListCardDevicesByDeviceID 需要 device_id
	var actualDeviceID string = deviceID

	// 判断是否为 UUID 格式（device_id）
	if len(deviceID) != 36 || strings.Count(deviceID, "-") != 4 {
		// 不是 UUID，可能是 device_uid，需要转换
		device, err := h.radarInstall.GetDeviceByUID(ctx, tenantID, deviceID)
		if err != nil {
			h.logger.Warn("[RADAR_STREAM_SSE] Failed to convert device_uid to device_id",
				zap.String("device_uid", deviceID),
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "event: error\ndata: {\"error\":\"Device not found\"}\n\n")
			return
		}
		actualDeviceID = device.DeviceID
	}

	// 获取该设备所属的 card_id
	// SSE 推送基于 card_id，不需要检查 device 的绑定状态
	cardID, _, _, _, err := h.radarInstall.ListCardDevicesByDeviceID(ctx, tenantID, actualDeviceID)
	if err != nil {
		h.logger.Error("[RADAR_STREAM_SSE] failed to get card_id",
			zap.String("device_id", actualDeviceID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Failed to get card info\"}\n\n")
		return
	}

	// 获取 DataStreamSubscriber
	sub, ok := h.dataStreamSubscriber.(*subscriber.DataStreamSubscriber)
	if !ok || sub == nil {
		h.logger.Error("[RADAR_STREAM_SSE] DataStreamSubscriber not available",
			zap.String("device_id", deviceID),
			zap.String("card_id", cardID))
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"DataStreamSubscriber not available\"}\n\n")
		return
	}

	// 获取 cardRealtimeUpdated 通知通道
	cardRealtimeUpdatedChan := sub.GetCardRealtimeUpdatedChan()

	// 发送初始连接确认
	flusher, ok := w.(http.Flusher)
	if !ok {
		// 使用 SSE 格式返回错误
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Streaming not supported\"}\n\n")
		return
	}

	// 发送初始连接确认消息（SSE 格式）
	io.WriteString(w, ": SSE connection established\n")
	io.WriteString(w, "event: connected\ndata: {\"device_id\":\""+deviceID+"\",\"status\":\"connected\"}\n\n")
	flusher.Flush()

	h.logger.Info("[RADAR_STREAM_SSE] connection established",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID))

	// 定期发送心跳（每 30 秒）
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 发送心跳注释（SSE 规范：以 : 开头的行是注释）
				io.WriteString(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}()

	// 主循环：监听 cardRealtimeUpdated 通知（事件驱动模式）
	// 当该 card 的缓存数据更新时，会收到通知
	// 相比轮询模式：零延迟、零 Redis 压力
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("[RADAR_STREAM_SSE] connection closed",
				zap.String("device_id", deviceID),
				zap.String("card_id", cardID))
			return
		case notifiedCardID := <-cardRealtimeUpdatedChan:
			// 拒绝不是本连接关注的 card 的通知
			if notifiedCardID != cardID {
				continue
			}

			// 从 DataStreamSubscriber 缓存中获取该 card 的最新实时数据
			cachedData := sub.GetCardRealtimeData(cardID)
			if cachedData == nil {
				continue
			}

			// 序列化为 JSON
			jsonData, err := json.Marshal(cachedData)
			if err != nil {
				h.logger.Warn("[RADAR_STREAM_SSE] failed to marshal card realtime data",
					zap.String("device_id", deviceID),
					zap.String("card_id", cardID),
					zap.Error(err))
				continue
			}

			// 通过 SSE 推送给前端
			io.WriteString(w, "data: "+string(jsonData)+"\n\n")
			flusher.Flush()

			// 调试日志（Debug 级别避免高频占用磁盘空间）
			h.logger.Debug("[RADAR_STREAM_SSE] card realtime data pushed",
				zap.String("device_id", deviceID),
				zap.String("card_id", cardID))
		}
	}
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
// Body: v1 格式 { install_model, height, boundary_left/right/front/rear }，由 encode.EncodeV1ConfigToDeviceProps 转成 qinglan 属性后写入。
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
		h.logger.Warn("UpdateConfig: invalid body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if config == nil {
		config = make(map[string]interface{})
	}
	// 兼容 body 被包成 { "data": { declare_area, ... } } 的情况
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
	logFields := []zap.Field{zap.String("device_id", deviceID), zap.Int("keys", len(config)), zap.Strings("keys_list", configKeys)}
	if v := config["declare_area"]; v != nil {
		if s, ok := v.(string); ok {
			logFields = append(logFields, zap.String("declare_area", s))
		} else {
			logFields = append(logFields, zap.Any("declare_area", v))
		}
	}
	if v := config["rectangle"]; v != nil {
		logFields = append(logFields, zap.Any("rectangle", v))
	}
	h.logger.Info("UpdateConfig: received config", logFields...)
	deviceCode, err := h.radarInstall.UpdateConfig(r.Context(), tenantID, deviceID, config)
	if err != nil {
		h.logger.Error("Failed to update device config",
			zap.String("device_id", deviceID),
			zap.Int("device_code", deviceCode),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Result[any]{Code: ResultError, Type: "error", Message: err.Error(), Result: map[string]interface{}{"device_code": deviceCode}})
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{"device_code": deviceCode}))
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

// tenantIDFromReq 从 X-Tenant-Id header 获取 tenant_id
// AuthMiddleware 已经验证了 token 并注入了 X-Tenant-Id header
func (h *RadarHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenantID := r.Header.Get("X-Tenant-Id")
	if tenantID == "" || tenantID == "null" {
		writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
		return "", false
	}
	return tenantID, true
}
