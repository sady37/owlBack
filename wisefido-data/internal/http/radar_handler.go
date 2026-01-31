package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"
	"wisefido-data/internal/store"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RadarHandler 雷达设备 API Handler
// 对接 owlFront radarDeviceApi，路径 :id 为 device_id（devices.device_id），
// 内部通过 radar_install_service 解析 device_uid 后调 qinglan_client → wisefido-qinglan。
// 路由：/radar-device/api/v1/radar-device/device/:id/{realtime|original-properties|config}
type RadarHandler struct {
	radarInstall *service.RadarInstall
	stubHandler  *StubHandler
	kv           store.KV
	redisClient  *redis.Client
	logger       *zap.Logger
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

	// 验证订阅状态
	isSubscribed := h.radarInstall.IsDeviceSubscribed(deviceID)
	h.logger.Info("[RADAR_BIND_API] bind success",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
		zap.Bool("is_subscribed", isSubscribed))
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

// GetDeviceStatus 获取设备在线状态
// GET /radar-device/api/v1/radar-device/device/:id/status
// 支持传入 device_id (UUID格式) 或 device_uid (短字符串格式)
// 如果是 device_uid，直接调用 wisefido-qinglan API，避免数据库查询
func (h *RadarHandler) GetDeviceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	identifier := extractRadarDeviceIDFromPath(r.URL.Path, "/radar-device/api/v1/radar-device/device/", "/status")
	h.logger.Debug("[RADAR_GET_STATUS] Request received",
		zap.String("path", r.URL.Path),
		zap.String("identifier", identifier))
	if !validateRadarDeviceID(identifier) {
		h.logger.Warn("[RADAR_GET_STATUS] Invalid device ID",
			zap.String("identifier", identifier),
			zap.String("path", r.URL.Path))
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		h.logger.Warn("[RADAR_GET_STATUS] Failed to get tenant_id",
			zap.String("identifier", identifier))
		return
	}
	ctx := r.Context()

	var deviceUID string
	// 判断是 device_id (UUID格式，36个字符，包含连字符) 还是 device_uid (短字符串格式，通常12个字符)
	// UUID格式: 8-4-4-4-12 (例如: 791fc634-69de-4987-b7eb-803c17e545a5)
	// device_uid格式: 通常12个字符，如 E598A2ACD523
	if len(identifier) == 36 && strings.Count(identifier, "-") == 4 {
		// 是 device_id (UUID格式)，需要查询数据库转换为 device_uid
		h.logger.Debug("[RADAR_GET_STATUS] Identifier is device_id (UUID), querying database",
			zap.String("device_id", identifier))
		var err error
		deviceUID, err = h.radarInstall.GetDeviceUID(ctx, tenantID, identifier)
		if err != nil {
			h.logger.Error("GetDeviceStatus: GetDeviceUID failed",
				zap.String("device_id", identifier),
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			writeJSON(w, http.StatusOK, Fail(err.Error()))
			return
		}
	} else {
		// 是 device_uid (短字符串格式)，直接使用
		h.logger.Debug("[RADAR_GET_STATUS] Identifier is device_uid, using directly",
			zap.String("device_uid", identifier))
		deviceUID = identifier
	}

	// 通过 qinglanClient 查询设备状态（wisefido-qinglan 会直接返回 offline 如果设备已断电）
	status, err := h.radarInstall.GetDeviceStatus(ctx, deviceUID)
	if err != nil {
		h.logger.Error("GetDeviceStatus: qinglanClient.GetDeviceStatus failed",
			zap.String("identifier", identifier),
			zap.String("device_uid", deviceUID),
			zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"device_uid": deviceUID,
		"status":     status,
	}))
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
// 查询参数 source=stream：直接从 Redis stream (iot:monitor:stream) 读取最新 track 数据（用于 radar-trajectory）
// 默认：从 Redis cache 读取（用于卡片汇总，2秒/次）
// 返回 { positions: [{x,y,z?,timestamp,remaining_time?}], vital: { heart, breath, heart_source, breath_source } }。
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
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	cardID, _, _, _, err := h.radarInstall.ListCardDevicesByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		h.logger.Error("GetRealtimeData: ListCardDevicesByDeviceID", zap.String("device_id", deviceID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	var positions []map[string]interface{}

	// 检查是否从 stream 读取（radar-trajectory 使用）
	if r.URL.Query().Get("source") == "stream" {
		positions = h.getDataFromStream(ctx, deviceID, tenantID)
	} else {
		// 默认从 Redis cache 读取（卡片汇总使用）
		positions = h.getPositionsFromCache(ctx, cardID, deviceID)
	}

	// vital: from vital-focus:card:{card_id}:realtime (radar/sleepad 分源，Sleepad 优先)
	var vital map[string]interface{}
	realtimeKey := "vital-focus:card:" + cardID + ":realtime"
	realtimeVal, err := h.kv.Get(ctx, realtimeKey)
	if err == nil {
		var rt struct {
			Radar *struct {
				Heart  *int `json:"heart,omitempty"`
				Breath *int `json:"breath,omitempty"`
			} `json:"radar,omitempty"`
			Sleepad *struct {
				Heart  *int `json:"heart,omitempty"`
				Breath *int `json:"breath,omitempty"`
			} `json:"sleepad,omitempty"`
		}
		if json.Unmarshal([]byte(realtimeVal), &rt) == nil {
			var sH, rH, sB, rB *int
			if rt.Sleepad != nil {
				sH, sB = rt.Sleepad.Heart, rt.Sleepad.Breath
			}
			if rt.Radar != nil {
				rH, rB = rt.Radar.Heart, rt.Radar.Breath
			}
			heart, heartSrc := pickVital(sH, rH)
			breath, breathSrc := pickVital(sB, rB)
			vital = map[string]interface{}{"heart": heart, "breath": breath, "heart_source": heartSrc, "breath_source": breathSrc}
		}
	} else if !errors.Is(err, store.ErrMiss) {
		h.logger.Debug("GetRealtimeData: realtime cache get", zap.String("key", realtimeKey), zap.Error(err))
	}

	response := map[string]interface{}{
		"positions": positions,
		"vital":     vital,
	}

	// 日志：打印最终返回给前端的数据
	h.logger.Info("[RADAR_REALTIME_API] response",
		zap.String("device_id", deviceID),
		zap.String("card_id", cardID),
		zap.String("source", r.URL.Query().Get("source")),
		zap.Int("positions_count", len(positions)),
		zap.Any("positions", positions),
		zap.Any("vital", vital))

	writeJSON(w, http.StatusOK, Ok(response))
}

// getPositionsFromCache 从 Redis cache 读取 positions（卡片汇总使用）
func (h *RadarHandler) getPositionsFromCache(ctx context.Context, cardID, deviceID string) []map[string]interface{} {
	positions := []map[string]interface{}{}
	deviceKey := "vital-focus:card:" + cardID + ":device:" + deviceID + ":data"
	deviceVal, err := h.kv.Get(ctx, deviceKey)
	if err == nil {
		var dev struct {
			PositionX *int   `json:"position_x,omitempty"`
			PositionY *int   `json:"position_y,omitempty"`
			PositionZ *int   `json:"position_z,omitempty"`
			Height    *int   `json:"height,omitempty"` // 兼容旧字段名
			Timestamp string `json:"timestamp,omitempty"`
		}
		if json.Unmarshal([]byte(deviceVal), &dev) == nil && dev.PositionX != nil && dev.PositionY != nil {
			ts := int64(0)
			if dev.Timestamp != "" {
				if t, e := time.Parse(time.RFC3339Nano, dev.Timestamp); e == nil {
					ts = t.Unix()
				} else if t, e := time.Parse(time.RFC3339, dev.Timestamp); e == nil {
					ts = t.Unix()
				}
			}
			p := map[string]interface{}{"x": *dev.PositionX, "y": *dev.PositionY, "timestamp": ts}
			// 优先使用 position_z，
			if dev.PositionZ != nil {
				p["z"] = *dev.PositionZ
			}
			positions = append(positions, p)

			// 日志：打印返回给前端的 positions
			var zVal interface{} = nil
			if dev.PositionZ != nil {
				zVal = *dev.PositionZ
			}
			h.logger.Info("[RADAR_REALTIME_API] positions output",
				zap.String("device_id", deviceID),
				zap.String("card_id", cardID),
				zap.Int("x", *dev.PositionX),
				zap.Int("y", *dev.PositionY),
				zap.Any("z", zVal),
				zap.Int64("timestamp", ts),
				zap.String("redis_key", deviceKey),
				zap.String("redis_raw", deviceVal))
		} else {
			h.logger.Warn("[RADAR_REALTIME_API] failed to parse device data",
				zap.String("device_id", deviceID),
				zap.String("card_id", cardID),
				zap.String("redis_key", deviceKey),
				zap.String("redis_raw", deviceVal),
				zap.Error(err))
		}
	} else if !errors.Is(err, store.ErrMiss) {
		h.logger.Debug("GetRealtimeData: device cache get", zap.String("key", deviceKey), zap.Error(err))
	} else {
		// 日志：Redis key 不存在
		h.logger.Info("[RADAR_REALTIME_API] device cache miss",
			zap.String("device_id", deviceID),
			zap.String("card_id", cardID),
			zap.String("redis_key", deviceKey))
	}
	return positions
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

	// 手动获取 tenant_id，避免 tenantIDFromReq 调用 writeJSON 覆盖 Content-Type
	// 直接检查 header 和 query，不调用可能写入响应的函数
	var tenantID string
	if h.stubHandler != nil {
		// 先检查 query 参数
		if tid := r.URL.Query().Get("tenant_id"); tid != "" && tid != "null" {
			tenantID = tid
		} else if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
			tenantID = tid
		} else if h.stubHandler != nil && h.stubHandler.DB != nil {
			// 尝试从 DB 查询（如果可用）
			userID := r.Header.Get("X-User-Id")
			if userID != "" {
				var dbTenantID string
				err := h.stubHandler.DB.QueryRowContext(r.Context(), "SELECT tenant_id::text FROM users WHERE user_id = $1", userID).Scan(&dbTenantID)
				if err == nil && dbTenantID != "" {
					tenantID = dbTenantID
				}
			}
		}

		// SystemAdmin 回退到 System tenant
		if tenantID == "" && strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
			// SystemTenantID 的固定值（与 stub_handler_base.go 保持一致）
			tenantID = "00000000-0000-0000-0000-000000000001"
		}

		if tenantID == "" {
			// 使用 SSE 格式返回错误
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "event: error\ndata: {\"error\":\"tenant_id is required\"}\n\n")
			return
		}
	} else {
		if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
			tenantID = tid
		} else {
			// 使用 SSE 格式返回错误
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "event: error\ndata: {\"error\":\"tenant_id is required\"}\n\n")
			return
		}
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
	
	// 将 deviceID 转换为 deviceUID 和 deviceID（用于订阅检查和 Redis Stream 匹配）
	// 无论前端传入的是 device_id 还是 device_uid，都同时获取两者
	var deviceUIDForMatch string
	var deviceIDForMatch string = deviceID // 默认使用传入的 deviceID
	
	if len(deviceID) == 36 && strings.Count(deviceID, "-") == 4 {
		// 是 device_id (UUID格式)，需要查询数据库转换为 device_uid
		uid, err := h.radarInstall.GetDeviceUID(ctx, tenantID, deviceID)
		if err != nil {
			h.logger.Warn("[RADAR_STREAM_SSE] Failed to get device_uid, will only match by device_id",
				zap.String("device_id", deviceID),
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			deviceUIDForMatch = "" // 如果转换失败，只使用 device_id 匹配
		} else {
			deviceUIDForMatch = uid
		}
	} else {
		// 是 device_uid (短字符串格式)，需要查询数据库转换为 device_id（用于订阅检查）
		deviceUIDForMatch = deviceID
		// 通过 device_uid 查询 device_id（用于订阅状态检查，因为 subscribedDevices 使用 device_id 作为 key）
		device, err := h.radarInstall.GetDeviceByUID(ctx, tenantID, deviceID)
		if err != nil {
			h.logger.Warn("[RADAR_STREAM_SSE] Failed to get device_id from device_uid, subscription check may fail",
				zap.String("device_uid", deviceID),
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			deviceIDForMatch = "" // 如果转换失败，订阅检查会失败
		} else {
			deviceIDForMatch = device.DeviceID
		}
	}
	
	// 检查设备是否需要订阅（通过 radar_install_service 的订阅管理器）
	// 注意：subscribedDevices 使用 device_id (UUID) 作为 key，所以需要用 deviceIDForMatch 检查
	isSubscribed := false
	if deviceIDForMatch != "" {
		isSubscribed = h.radarInstall.IsDeviceSubscribed(deviceIDForMatch)
	}
	h.logger.Info("[RADAR_STREAM_SSE] subscription check",
		zap.String("device_id", deviceIDForMatch),
		zap.String("device_uid", deviceUIDForMatch),
		zap.String("original_identifier", deviceID),
		zap.Bool("is_subscribed", isSubscribed))

	if !isSubscribed {
		// 设备未绑定，使用 SSE 格式返回错误
		h.logger.Warn("[RADAR_STREAM_SSE] device not subscribed, connection rejected",
			zap.String("device_id", deviceIDForMatch),
			zap.String("device_uid", deviceUIDForMatch),
			zap.String("tenant_id", tenantID))
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Device not subscribed. Please bind the device first.\"}\n\n")
		return
	}
	
	// 监听 stream：monitor（track数据）、alarm（设备状态变更：isOnline/OfflineAlarm）
	// 注意：isOnline 和 OfflineAlarm 已统一发送到 alarm stream，不再需要订阅 event stream
	monitorStream := "iot:monitor:stream"
	alarmStream := "iot:alarm:stream"
	// 使用 "$" 从最新消息开始（只读取新消息，不读取历史消息）
	// 如果 stream 不存在，第一次读取会报错，然后自动降级为 "0"
	monitorLastID := "$"
	alarmLastID := "$"

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
		zap.String("monitor_stream", monitorStream),
		zap.String("alarm_stream", alarmStream),
		)

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

	// stream 读取结果（monitor / alarm 先到先处理，避免等 2s 拖慢 track 的 1s 推送）
	type streamResult struct {
		streamName string
		streams    []redis.XStream
		err        error
	}
	resultChan := make(chan streamResult, 4)
	var lastIDMu sync.Mutex

	// 常驻：只读 monitor stream（Block 1s），使 track 能约 1s 推到前端
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lastIDMu.Lock()
				mid := monitorLastID
				lastIDMu.Unlock()
				streams, err := h.redisClient.XRead(ctx, &redis.XReadArgs{
					Streams: []string{monitorStream, mid},
					Count:   1,
					Block:   1 * time.Second,
				}).Result()
				resultChan <- streamResult{streamName: monitorStream, streams: streams, err: err}
			}
		}
	}()

	// 常驻：只读 alarm stream（Block 2s）
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lastIDMu.Lock()
				aid := alarmLastID
				lastIDMu.Unlock()
				streams, err := h.redisClient.XRead(ctx, &redis.XReadArgs{
					Streams: []string{alarmStream, aid},
					Count:   10,
					Block:   2 * time.Second,
				}).Result()
				resultChan <- streamResult{streamName: alarmStream, streams: streams, err: err}
			}
		}
	}()

	// 主循环：先到先处理，不再等两个都返回（原逻辑等 2 个导致整轮 2s）
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-resultChan:
				if result.err != nil {
					if result.err == redis.Nil {
						// 超时，继续等待
						continue
					}
					// 如果 stream 不存在或 ID 无效，降级为 "0"（从最早的消息开始）
					// 这通常发生在 stream 不存在时使用 "$" 的情况
					errStr := result.err.Error()
					if errStr == "ERR Invalid stream ID specified as stream command argument" ||
						errStr == "ERR The XADD id specified in XREADGROUP must be greater than 0-0" {
						// Stream 不存在，降级为 "0" 以便后续读取（如果 stream 被创建）
						lastIDMu.Lock()
						if result.streamName == monitorStream && monitorLastID == "$" {
							monitorLastID = "0"
							h.logger.Debug("[RADAR_STREAM_SSE] stream doesn't exist, using 0 as fallback",
								zap.String("device_id", deviceID),
								zap.String("stream", result.streamName))
						} else if result.streamName == alarmStream && alarmLastID == "$" {
							alarmLastID = "0"
							h.logger.Debug("[RADAR_STREAM_SSE] stream doesn't exist, using 0 as fallback",
								zap.String("device_id", deviceID),
								zap.String("stream", result.streamName))
						}
						lastIDMu.Unlock()
						continue
					}
					h.logger.Warn("[RADAR_STREAM_SSE] failed to read stream",
						zap.String("device_id", deviceID),
						zap.String("stream", result.streamName),
						zap.Error(result.err))
					continue
				}

				// 处理读取到的消息
				for _, stream := range result.streams {
					streamName := stream.Stream
					for _, msg := range stream.Messages {
						// 更新对应 stream 的 lastID（与读 goroutine 同步）
						lastIDMu.Lock()
						if streamName == monitorStream {
							monitorLastID = msg.ID
						} else if streamName == alarmStream {
							alarmLastID = msg.ID
						}
						lastIDMu.Unlock()

						// 检查 device_id/device_uid 和 tenant_id
						// Redis Stream 中可能包含 device_id（UUID）或 device_uid（短字符串）
						// 需要同时匹配两者，因为 wisefido-qinglan 发布时可能使用不同的字段
						msgDeviceID, _ := msg.Values["device_id"].(string)
						msgDeviceUID, _ := msg.Values["device_uid"].(string)
						msgTenantID, _ := msg.Values["tenant_id"].(string)
						category, _ := msg.Values["category"].(string)
						
						// 匹配逻辑：tenant_id 必须匹配，优先使用 device_uid 匹配，如果 device_uid 不匹配再尝试 device_id
						tenantMatched := msgTenantID == tenantID
						deviceMatched := false
						
						// 优先匹配 device_uid（Redis Stream 中 device_uid 字段更可靠）
						if deviceUIDForMatch != "" && msgDeviceUID != "" {
							deviceMatched = msgDeviceUID == deviceUIDForMatch
						}
						// 如果 device_uid 不匹配，尝试匹配 device_id（作为后备）
						if !deviceMatched && deviceIDForMatch != "" && msgDeviceID != "" {
							deviceMatched = msgDeviceID == deviceIDForMatch
						}
						
						if !tenantMatched || !deviceMatched {
							continue
						}

						// 处理设备状态变更（来自 alarm stream：isOnline 和 OfflineAlarm 已统一发送到 alarm stream）
						if streamName == alarmStream {
							// 处理 OfflineAlarm（设备离线告警）
							if category == "OfflineAlarm" {
								// 通过 SSE 推送设备状态变更
								statusData := map[string]interface{}{
									"device_id": deviceID,
									"status":    "offline",
								}
								statusJSON, _ := json.Marshal(statusData)
								io.WriteString(w, "event: device_status\ndata: "+string(statusJSON)+"\n\n")
								flusher.Flush()

								h.logger.Info("[RADAR_STREAM_SSE] device status pushed (offline)",
									zap.String("device_id", deviceID),
									zap.String("stream", streamName),
									zap.String("category", category))
								continue
							}

							// 处理 isOnline 事件（设备在线/离线状态）
							if category == "isOnline" {
								// 解析 data_value 获取 device_status
								var deviceStatus string = "offline"
								if dataValueStr, ok := msg.Values["data_value"].(string); ok {
									var dataValueArray []map[string]interface{}
									if err := json.Unmarshal([]byte(dataValueStr), &dataValueArray); err == nil && len(dataValueArray) > 0 {
										if status, ok := dataValueArray[0]["device_status"].(string); ok {
											deviceStatus = status
										}
									}
								} else if dataValueArray, ok := msg.Values["data_value"].([]interface{}); ok && len(dataValueArray) > 0 {
									if firstItem, ok := dataValueArray[0].(map[string]interface{}); ok {
										if status, ok := firstItem["device_status"].(string); ok {
											deviceStatus = status
										}
									}
								}

								// 只推送 offline 状态（online 状态通过 Query 成功来更新）
								if deviceStatus == "offline" {
									statusData := map[string]interface{}{
										"device_id": deviceID,
										"status":    deviceStatus,
									}
									statusJSON, _ := json.Marshal(statusData)
									io.WriteString(w, "event: device_status\ndata: "+string(statusJSON)+"\n\n")
									flusher.Flush()

									h.logger.Info("[RADAR_STREAM_SSE] device status pushed",
										zap.String("device_id", deviceID),
										zap.String("status", deviceStatus),
										zap.String("stream", streamName),
										zap.String("category", category))
								}
								continue
							}
						}

						// 只处理 monitor stream 的 track 数据
						if streamName != monitorStream || category != "track" {
							continue
						}

						// 解析 data_value
						var dataValue map[string]interface{}
						if dataValueStr, ok := msg.Values["data_value"].(string); ok {
							if err := json.Unmarshal([]byte(dataValueStr), &dataValue); err != nil {
								continue
							}
						} else if dataValueMap, ok := msg.Values["data_value"].(map[string]interface{}); ok {
							dataValue = dataValueMap
						} else {
							continue
						}

						// 提取关键字段用于日志
						positionX, _ := dataValue["position_x"]
						positionY, _ := dataValue["position_y"]
						positionZ, _ := dataValue["position_z"]
						targetID, _ := dataValue["target_id"]
						pose, _ := dataValue["pose"] // 提取 pose 字段用于日志

						// 去掉 raw_original 字段
						delete(dataValue, "raw_original")

						// 添加 timestamp
						var timestamp int64
						if tsStr, ok := msg.Values["timestamp"].(string); ok {
							if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
								timestamp = ts
								dataValue["timestamp"] = ts
							}
						}

						// 序列化为 JSON
						jsonData, err := json.Marshal(dataValue)
						if err != nil {
							h.logger.Warn("[RADAR_STREAM_SSE] failed to marshal data",
								zap.String("device_id", deviceID),
								zap.Error(err))
							continue
						}

						// 发送 SSE 消息
						io.WriteString(w, "data: "+string(jsonData)+"\n\n")
						flusher.Flush()

						// 记录推送日志（Debug 级别，避免高频日志占用空间）
						h.logger.Debug("[RADAR_STREAM_SSE] message pushed",
							zap.String("device_id", deviceID),
							zap.String("stream_id", msg.ID),
							zap.String("category", category),
							zap.Any("position_x", positionX),
							zap.Any("position_y", positionY),
							zap.Any("position_z", positionZ),
							zap.Any("target_id", targetID),
							zap.Any("pose", pose),
							zap.Int64("timestamp", timestamp))
					}
				}
		}
	}
}

// getDataFromStream 从 Redis stream 直接转发数据（radar-trajectory 使用）
// 完全不做任何操作，只去掉 raw_original 字段，直接转发给前端
func (h *RadarHandler) getDataFromStream(ctx context.Context, deviceID, tenantID string) []map[string]interface{} {
	positions := []map[string]interface{}{}

	if h.redisClient == nil {
		h.logger.Warn("[RADAR_REALTIME_API] redis client not available for stream read")
		return positions
	}

	streamName := "iot:monitor:stream"
	// 读取最新的 100 条消息（足够覆盖最近几秒的数据）
	messages, err := h.redisClient.XRevRangeN(ctx, streamName, "+", "-", 100).Result()
	if err != nil {
		if err != redis.Nil {
			h.logger.Warn("[RADAR_REALTIME_API] failed to read from stream",
				zap.String("stream", streamName),
				zap.String("device_id", deviceID),
				zap.Error(err))
		}
		return positions
	}

	// 注意：getDataFromStream 用于轮询接口，不检查订阅状态
	// 订阅状态检查只在 SSE 端点中进行

	// 从最新消息开始查找，找到第一个匹配指定 device_id 的数据
	for _, msg := range messages {
		// 检查 device_id 和 tenant_id
		msgDeviceID, _ := msg.Values["device_id"].(string)
		msgTenantID, _ := msg.Values["tenant_id"].(string)
		if msgDeviceID != deviceID || msgTenantID != tenantID {
			continue
		}

		// 解析 data_value（可能是 JSON 字符串或 map）
		var dataValue map[string]interface{}
		if dataValueStr, ok := msg.Values["data_value"].(string); ok {
			if err := json.Unmarshal([]byte(dataValueStr), &dataValue); err != nil {
				continue
			}
		} else if dataValueMap, ok := msg.Values["data_value"].(map[string]interface{}); ok {
			dataValue = dataValueMap
		} else {
			continue
		}

		// 去掉 raw_original 字段
		delete(dataValue, "raw_original")

		// 解析 timestamp（从 stream message 中获取）
		if tsStr, ok := msg.Values["timestamp"].(string); ok {
			if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
				dataValue["timestamp"] = ts
			}
		}

		// 直接转发整个 data_value（去掉 raw_original 后）
		positions = append(positions, dataValue)

		// 只返回第一个匹配的数据（最新的）
		break
	}

	return positions
}

// pickVital 按 Sleepad 优先选 display 与 source
func pickVital(sPtr, rPtr *int) (interface{}, string) {
	if sPtr != nil {
		return *sPtr, "sleepad"
	}
	if rPtr != nil {
		return *rPtr, "radar"
	}
	return nil, ""
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
