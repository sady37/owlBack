package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// AlarmEventHandler 报警事件 Handler
type AlarmEventHandler struct {
	alarmEventService service.AlarmEventService
	logger            *zap.Logger
	base              *StubHandler // 用于 tenantIDFromReq
}

// NewAlarmEventHandler 创建报警事件 Handler
func NewAlarmEventHandler(alarmEventService service.AlarmEventService, logger *zap.Logger) *AlarmEventHandler {
	return &AlarmEventHandler{
		alarmEventService: alarmEventService,
		logger:            logger,
		base:              &StubHandler{}, // 用于 tenantIDFromReq
	}
}

// parseArrayParam 从 query 解析数组：优先 key/altKey 多值，否则按逗号分割单值
func parseArrayParam(q map[string][]string, key, altKey string) []string {
	var out []string
	for _, k := range []string{key, altKey} {
		for _, v := range q[k] {
			for _, part := range strings.Split(v, ",") {
				if t := strings.TrimSpace(part); t != "" {
					out = append(out, t)
				}
			}
		}
	}
	return out
}

// ServeHTTP 实现 http.Handler 接口
func (h *AlarmEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	//
	// 两个独立入口（防 cross-card event leak）：
	//   GET /admin/api/v1/alarm-events           全局列表，按 role 过滤；服务 AlarmRecord.vue
	//   GET /admin/api/v1/alarm-events/by-card   严格 scope（5W 直接 / card 反查）；缺 scope → 400
	//                                            服务 Detail / WaveMonitor / Overview-RePlay
	path := r.URL.Path
	switch {
	case path == "/admin/api/v1/alarm-events/by-card" && r.Method == http.MethodGet:
		h.ScopedListAlarmEvents(w, r)
	case path == "/admin/api/v1/alarm-events" && r.Method == http.MethodGet:
		h.ListAlarmEvents(w, r)
	// HandleAlarmEvent
	case strings.HasSuffix(path, "/handle") && r.Method == http.MethodPut:
		eventID := strings.TrimSuffix(path, "/handle")
		eventID = strings.TrimPrefix(eventID, "/admin/api/v1/alarm-events/")
		if eventID != "" && !strings.Contains(eventID, "/") {
			h.HandleAlarmEvent(w, r, eventID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ============================================
// ListAlarmEvents 查询报警事件列表
// ============================================

// parseListAlarmEventsRequest 解析查询参数（两个入口共用）。
// 第二返回值 ok=false 表示已经向 client 写入 4xx，调用方直接 return。
func (h *AlarmEventHandler) parseListAlarmEventsRequest(w http.ResponseWriter, r *http.Request) (service.ListAlarmEventsRequest, bool) {
	var req service.ListAlarmEventsRequest

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return req, false
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return req, false
	}

	currentUserRole := r.Header.Get("X-User-Role")
	if currentUserRole == "" {
		writeJSON(w, http.StatusOK, Fail("user role is required"))
		return req, false
	}

	q := r.URL.Query()
	status := strings.TrimSpace(q.Get("status"))
	page := parseInt(q.Get("page"), 1)
	pageSize := parseInt(q.Get("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	var alarmTimeStart, alarmTimeEnd *int64
	if startStr := strings.TrimSpace(q.Get("alarm_time_start")); startStr != "" {
		if start, err := strconv.ParseInt(startStr, 10, 64); err == nil {
			alarmTimeStart = &start
		}
	}
	if endStr := strings.TrimSpace(q.Get("alarm_time_end")); endStr != "" {
		if end, err := strconv.ParseInt(endStr, 10, 64); err == nil {
			alarmTimeEnd = &end
		}
	}

	resident := strings.TrimSpace(q.Get("resident"))
	branchTag := strings.TrimSpace(q.Get("branch_tag"))
	if branchTag == "" {
		branchTag = strings.TrimSpace(q.Get("branch_name"))
	}
	unitName := strings.TrimSpace(q.Get("unit_name"))
	deviceName := strings.TrimSpace(q.Get("device_name"))

	eventTypes := parseArrayParam(q, "event_types", "event_types[]")
	categories := parseArrayParam(q, "categories", "categories[]")
	alarmLevels := parseArrayParam(q, "alarm_levels", "alarm_levels[]")

	cardID := strings.TrimSpace(q.Get("card_id"))
	roomID := strings.TrimSpace(q.Get("room_id"))
	var deviceIDs []string
	if deviceIDsStr := strings.TrimSpace(q.Get("device_ids")); deviceIDsStr != "" {
		deviceIDs = strings.Split(deviceIDsStr, ",")
	}

	scopeUnitID := strings.TrimSpace(q.Get("scope_unit_id"))
	scopeRoomID := strings.TrimSpace(q.Get("scope_room_id"))
	scopeDeviceID := strings.TrimSpace(q.Get("scope_device_id"))

	req = service.ListAlarmEventsRequest{
		TenantID:        tenantID,
		CurrentUserID:   currentUserID,
		CurrentUserRole: currentUserRole,
		Status:          status,
		AlarmTimeStart:  alarmTimeStart,
		AlarmTimeEnd:    alarmTimeEnd,
		Resident:        resident,
		BranchName:      branchTag,
		UnitName:        unitName,
		DeviceName:      deviceName,
		EventTypes:      eventTypes,
		Categories:      categories,
		AlarmLevels:     alarmLevels,
		CardID:          cardID,
		DeviceIDs:       deviceIDs,
		RoomID:          roomID,
		ScopeUnitID:     scopeUnitID,
		ScopeRoomID:     scopeRoomID,
		ScopeDeviceID:   scopeDeviceID,
		Page:            page,
		PageSize:        pageSize,
	}
	return req, true
}

// ScopedListAlarmEvents：按 5W 严格 scope 查询。优先级：
//   scope_unit_id + scope_room_id + scope_device_id   (alarm_events 直接列过滤，最快最精确)
// > scope_unit_id + scope_room_id                     (alarm_events 直接列过滤，room 内全部设备)
// > card_id / room_id / device_ids                    (反查 devices 兼容旧调用)
//
// scope 全空 → 400 Bad Request（不允许穿透到 role-based 宽过滤，这是 cross-card leak 根因）。
func (h *AlarmEventHandler) ScopedListAlarmEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, ok := h.parseListAlarmEventsRequest(w, r)
	if !ok {
		return
	}

	hasDirectScope := (req.ScopeUnitID != "" && req.ScopeRoomID != "") || req.ScopeDeviceID != ""
	hasFallbackScope := req.CardID != "" || req.RoomID != "" || len(req.DeviceIDs) > 0
	if !hasDirectScope && !hasFallbackScope {
		writeJSON(w, http.StatusBadRequest, Fail(
			"scope required: scope_unit_id+scope_room_id (+scope_device_id) or card_id or room_id or device_ids",
		))
		return
	}

	resp, err := h.alarmEventService.ListAlarmEvents(ctx, req)
	if err != nil {
		h.logger.Error("ScopedListAlarmEvents failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(buildAlarmEventsListPayload(resp)))
}

// buildAlarmEventsListPayload 把 service response 转换为前端 GetAlarmEventsResult 形态。
func buildAlarmEventsListPayload(resp *service.ListAlarmEventsResponse) map[string]any {
	items := make([]any, 0, len(resp.Items))
	for _, item := range resp.Items {
		itemMap := map[string]any{
			"event_id":     item.EventID,
			"tenant_id":    item.TenantID,
			"device_id":    item.DeviceID,
			"event_type":   item.EventType,
			"category":     item.Category,
			"alarm_level":  item.AlarmLevel,
			"alarm_status": item.AlarmStatus,
			"triggered_at":    item.TriggeredAt,
			"triggered_at_ms": item.TriggeredAtMs,
		}
		if item.HandledAt != nil {
			itemMap["handled_at"] = *item.HandledAt
		}
		if item.HandledAtMs != nil {
			itemMap["handled_at_ms"] = *item.HandledAtMs
		}
		if item.HandlingState != nil {
			itemMap["handling_state"] = *item.HandlingState
		}
		if item.HandlingDetails != nil {
			itemMap["handling_details"] = *item.HandlingDetails
		}
		if item.HandlerID != nil {
			itemMap["handler_id"] = *item.HandlerID
		}
		if item.HandlerName != nil {
			itemMap["handler_name"] = *item.HandlerName
		}
		if item.CardID != nil {
			itemMap["card_id"] = *item.CardID
		}
		if item.DeviceName != nil {
			itemMap["device_name"] = *item.DeviceName
		}
		if item.ResidentID != nil {
			itemMap["resident_id"] = *item.ResidentID
		}
		if item.ResidentName != nil {
			itemMap["resident_name"] = *item.ResidentName
		}
		if item.ResidentGender != nil {
			itemMap["resident_gender"] = *item.ResidentGender
		}
		if item.ResidentAge != nil {
			itemMap["resident_age"] = *item.ResidentAge
		}
		if item.ResidentNetwork != nil {
			itemMap["resident_network"] = *item.ResidentNetwork
		}
		if item.BranchID != nil {
			itemMap["branch_id"] = *item.BranchID
		}
		if item.BranchName != nil {
			itemMap["branch_name"] = *item.BranchName
		}
		if item.BuildingID != nil {
			itemMap["building_id"] = *item.BuildingID
		}
		if item.BuildingName != nil {
			itemMap["building_name"] = *item.BuildingName
		}
		if item.Floor != nil {
			itemMap["floor"] = *item.Floor
		}
		if item.UnitID != nil {
			itemMap["unit_id"] = *item.UnitID
		}
		if item.UnitName != nil {
			itemMap["unit_name"] = *item.UnitName
		}
		if item.RoomName != nil {
			itemMap["room_name"] = *item.RoomName
		}
		if item.BedName != nil {
			itemMap["bed_name"] = *item.BedName
		}
		if item.AddressDisplay != nil {
			itemMap["address_display"] = *item.AddressDisplay
		}
		if item.UnitTimezone != nil {
			itemMap["unit_timezone"] = *item.UnitTimezone
		}
		if item.TriggerData != nil {
			itemMap["trigger_data"] = item.TriggerData
		}
		if item.Metadata != nil {
			itemMap["metadata"] = item.Metadata
		}
		items = append(items, itemMap)
	}
	return map[string]any{
		"items": items,
		"pagination": map[string]any{
			"size":  resp.Pagination.Size,
			"page":  resp.Pagination.Page,
			"count": resp.Pagination.Count,
			"total": resp.Pagination.Total,
		},
	}
}

// ListAlarmEvents 查询报警事件列表（全局入口）
//
// 不强制 scope；admin 看 tenant 全部，staff 按 branch/unit，resident 按 bed/unit。
// 注意：scope 参数（card_id/room_id/device_ids/scope_*）即使携带也仅生效在
// service.ListAlarmEvents 内部已有的 fail-secure 反查路径上——本入口的语义
// 是"全局列表"。需要严格 scope 的调用方请走 /admin/api/v1/alarm-events/by-card。
func (h *AlarmEventHandler) ListAlarmEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, ok := h.parseListAlarmEventsRequest(w, r)
	if !ok {
		return
	}

	resp, err := h.alarmEventService.ListAlarmEvents(ctx, req)
	if err != nil {
		h.logger.Error("ListAlarmEvents failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(buildAlarmEventsListPayload(resp)))
}

// ============================================
// HandleAlarmEvent 处理报警事件
// ============================================

// HandleAlarmEvent 处理报警事件（确认或解决）
func (h *AlarmEventHandler) HandleAlarmEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required"))
		return
	}

	currentUserType := r.Header.Get("X-User-Type")
	if currentUserType == "" {
		// 默认为 "staff"
		currentUserType = "staff"
	}

	currentUserRole := r.Header.Get("X-User-Role")
	if currentUserRole == "" {
		writeJSON(w, http.StatusOK, Fail("user role is required"))
		return
	}

	// 解析请求体
	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 解析参数
	alarmStatus, _ := payload["alarm_status"].(string)
	handleType, _ := payload["handle_type"].(string)
	remarks, _ := payload["remarks"].(string)

	if alarmStatus == "" {
		writeJSON(w, http.StatusOK, Fail("alarm_status is required"))
		return
	}

	// 构建请求
	req := service.HandleAlarmEventRequest{
		TenantID:        tenantID,
		EventID:         eventID,
		CurrentUserID:   currentUserID,
		CurrentUserType: currentUserType,
		CurrentUserRole: currentUserRole,
		AlarmStatus:     alarmStatus,
		HandleType:      handleType,
		Remarks:         remarks,
	}

	// 调用 Service
	resp, err := h.alarmEventService.HandleAlarmEvent(ctx, req)
	if err != nil {
		h.logger.Error("HandleAlarmEvent failed",
			zap.String("event_id", eventID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success":  resp.Success,
		"event_id": eventID,
	}))
}

