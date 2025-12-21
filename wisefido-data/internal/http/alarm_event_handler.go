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

// ServeHTTP 实现 http.Handler 接口
func (h *AlarmEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	path := r.URL.Path
	switch {
	// ListAlarmEvents
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

// ListAlarmEvents 查询报警事件列表
func (h *AlarmEventHandler) ListAlarmEvents(w http.ResponseWriter, r *http.Request) {
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

	currentUserRole := r.Header.Get("X-User-Role")
	if currentUserRole == "" {
		writeJSON(w, http.StatusOK, Fail("user role is required"))
		return
	}

	// 解析查询参数
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	// 时间范围过滤
	var alarmTimeStart, alarmTimeEnd *int64
	if startStr := strings.TrimSpace(r.URL.Query().Get("alarm_time_start")); startStr != "" {
		if start, err := strconv.ParseInt(startStr, 10, 64); err == nil {
			alarmTimeStart = &start
		}
	}
	if endStr := strings.TrimSpace(r.URL.Query().Get("alarm_time_end")); endStr != "" {
		if end, err := strconv.ParseInt(endStr, 10, 64); err == nil {
			alarmTimeEnd = &end
		}
	}

	// 搜索参数
	resident := strings.TrimSpace(r.URL.Query().Get("resident"))
	branchTag := strings.TrimSpace(r.URL.Query().Get("branch_tag"))
	unitName := strings.TrimSpace(r.URL.Query().Get("unit_name"))
	deviceName := strings.TrimSpace(r.URL.Query().Get("device_name"))

	// 过滤参数（多选）
	var eventTypes, categories, alarmLevels []string
	if eventTypesStr := strings.TrimSpace(r.URL.Query().Get("event_types")); eventTypesStr != "" {
		eventTypes = strings.Split(eventTypesStr, ",")
	}
	if categoriesStr := strings.TrimSpace(r.URL.Query().Get("categories")); categoriesStr != "" {
		categories = strings.Split(categoriesStr, ",")
	}
	if alarmLevelsStr := strings.TrimSpace(r.URL.Query().Get("alarm_levels")); alarmLevelsStr != "" {
		alarmLevels = strings.Split(alarmLevelsStr, ",")
	}

	// 关联过滤
	cardID := strings.TrimSpace(r.URL.Query().Get("card_id"))
	var deviceIDs []string
	if deviceIDsStr := strings.TrimSpace(r.URL.Query().Get("device_ids")); deviceIDsStr != "" {
		deviceIDs = strings.Split(deviceIDsStr, ",")
	}

	// 构建请求
	req := service.ListAlarmEventsRequest{
		TenantID:        tenantID,
		CurrentUserID:   currentUserID,
		CurrentUserRole: currentUserRole,
		Status:          status,
		AlarmTimeStart:  alarmTimeStart,
		AlarmTimeEnd:    alarmTimeEnd,
		Resident:        resident,
		BranchTag:       branchTag,
		UnitName:        unitName,
		DeviceName:      deviceName,
		EventTypes:      eventTypes,
		Categories:      categories,
		AlarmLevels:     alarmLevels,
		CardID:          cardID,
		DeviceIDs:       deviceIDs,
		Page:            page,
		PageSize:        pageSize,
	}

	// 调用 Service
	resp, err := h.alarmEventService.ListAlarmEvents(ctx, req)
	if err != nil {
		h.logger.Error("ListAlarmEvents failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为旧 Handler 格式（对齐 GetAlarmEventsResult）
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
			"triggered_at": item.TriggeredAt,
		}

		// 处理信息
		if item.HandledAt != nil {
			itemMap["handled_at"] = *item.HandledAt
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

		// 关联数据
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

		// 地址信息
		if item.BranchTag != nil {
			itemMap["branch_tag"] = *item.BranchTag
		}
		if item.Building != nil {
			itemMap["building"] = *item.Building
		}
		if item.Floor != nil {
			itemMap["floor"] = *item.Floor
		}
		if item.AreaTag != nil {
			itemMap["area_tag"] = *item.AreaTag
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

		// JSONB 字段
		if item.TriggerData != nil {
			itemMap["trigger_data"] = item.TriggerData
		}
		if item.NotifiedUsers != nil {
			itemMap["notified_users"] = item.NotifiedUsers
		}
		if item.Metadata != nil {
			itemMap["metadata"] = item.Metadata
		}

		items = append(items, itemMap)
	}

	// 分页信息
	pagination := map[string]any{
		"size":  resp.Pagination.Size,
		"page":  resp.Pagination.Page,
		"count": resp.Pagination.Count,
		"total": resp.Pagination.Total,
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items":      items,
		"pagination": pagination,
	}))
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

	// 转换为旧 Handler 格式（对齐旧响应）
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": resp.Success,
	}))
}

