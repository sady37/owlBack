package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// UnitHandler 单元管理 Handler（Building, Unit, Room, Bed）
type UnitHandler struct {
	unitService service.UnitService
	logger      *zap.Logger
}

// NewUnitHandler 创建单元管理 Handler
func NewUnitHandler(unitService service.UnitService, logger *zap.Logger) *UnitHandler {
	return &UnitHandler{
		unitService: unitService,
		logger:      logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *UnitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	// Buildings
	case r.URL.Path == "/admin/api/v1/buildings" && r.Method == http.MethodGet:
		h.ListBuildings(w, r)
	case r.URL.Path == "/admin/api/v1/buildings" && r.Method == http.MethodPost:
		h.CreateBuilding(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/buildings/") && r.Method == http.MethodGet:
		h.GetBuilding(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/buildings/") && r.Method == http.MethodPut:
		h.UpdateBuilding(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/buildings/") && r.Method == http.MethodDelete:
		h.DeleteBuilding(w, r)

	// Units
	case r.URL.Path == "/admin/api/v1/units" && r.Method == http.MethodGet:
		h.ListUnits(w, r)
	case r.URL.Path == "/admin/api/v1/units" && r.Method == http.MethodPost:
		h.CreateUnit(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/") && r.Method == http.MethodGet:
		h.GetUnit(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/") && r.Method == http.MethodPut:
		h.UpdateUnit(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/") && r.Method == http.MethodDelete:
		h.DeleteUnit(w, r)

	// Rooms
	case r.URL.Path == "/admin/api/v1/rooms" && r.Method == http.MethodGet:
		h.ListRoomsWithBeds(w, r)
	case r.URL.Path == "/admin/api/v1/rooms" && r.Method == http.MethodPost:
		h.CreateRoom(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/rooms/") && r.Method == http.MethodPut:
		h.UpdateRoom(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/rooms/") && r.Method == http.MethodDelete:
		h.DeleteRoom(w, r)

	// Beds
	case r.URL.Path == "/admin/api/v1/beds" && r.Method == http.MethodGet:
		h.ListBeds(w, r)
	case r.URL.Path == "/admin/api/v1/beds" && r.Method == http.MethodPost:
		h.CreateBed(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/beds/") && r.Method == http.MethodPut:
		h.UpdateBed(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/beds/") && r.Method == http.MethodDelete:
		h.DeleteBed(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ============================================
// Building 方法
// ============================================

// ListBuildings 查询楼栋列表
func (h *UnitHandler) ListBuildings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	branchTag := r.URL.Query().Get("branch_tag")

	req := service.ListBuildingsRequest{
		TenantID:  tenantID,
		BranchName: branchTag,
	}

	resp, err := h.unitService.ListBuildings(ctx, req)
	if err != nil {
		h.logger.Error("ListBuildings failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换响应格式（与旧 Handler 一致）
	out := make([]any, 0, len(resp.Items))
	for _, b := range resp.Items {
		out = append(out, buildingToJSON(b))
	}

	writeJSON(w, http.StatusOK, Ok(out))
}

// GetBuilding 获取单个楼栋详情
func (h *UnitHandler) GetBuilding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	buildingID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
	if buildingID == "" || strings.Contains(buildingID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	req := service.GetBuildingRequest{
		TenantID:   tenantID,
		BuildingID: buildingID,
	}

	resp, err := h.unitService.GetBuilding(ctx, req)
	if err != nil {
		h.logger.Error("GetBuilding failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(buildingToJSON(resp.Building)))
}

// CreateBuilding 创建楼栋
func (h *UnitHandler) CreateBuilding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.CreateBuildingRequest{
		TenantID:     tenantID,
		BranchName:    getString(payload, "branch_tag"),
		BuildingName: getString(payload, "building_name"),
	}

	resp, err := h.unitService.CreateBuilding(ctx, req)
	if err != nil {
		h.logger.Error("CreateBuilding failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 building 对象（与旧 Handler 格式一致）
	getReq := service.GetBuildingRequest{
		TenantID:   tenantID,
		BuildingID: resp.BuildingID,
	}
	getResp, err := h.unitService.GetBuilding(ctx, getReq)
	if err != nil {
		h.logger.Error("GetBuilding after CreateBuilding failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(buildingToJSON(getResp.Building)))
}

// UpdateBuilding 更新楼栋
func (h *UnitHandler) UpdateBuilding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	buildingID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
	if buildingID == "" || strings.Contains(buildingID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.UpdateBuildingRequest{
		TenantID:     tenantID,
		BuildingID:   buildingID,
		BranchName:    getString(payload, "branch_tag"),
		BuildingName: getString(payload, "building_name"),
	}

	_, err := h.unitService.UpdateBuilding(ctx, req)
	if err != nil {
		h.logger.Error("UpdateBuilding failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 building 对象（与旧 Handler 格式一致）
	getReq := service.GetBuildingRequest{
		TenantID:   tenantID,
		BuildingID: buildingID,
	}
	getResp, err := h.unitService.GetBuilding(ctx, getReq)
	if err != nil {
		h.logger.Error("GetBuilding after UpdateBuilding failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(buildingToJSON(getResp.Building)))
}

// DeleteBuilding 删除楼栋
func (h *UnitHandler) DeleteBuilding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	buildingID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
	if buildingID == "" || strings.Contains(buildingID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	req := service.DeleteBuildingRequest{
		TenantID:   tenantID,
		BuildingID: buildingID,
	}

	_, err := h.unitService.DeleteBuilding(ctx, req)
	if err != nil {
		h.logger.Error("DeleteBuilding failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 与旧 Handler 格式一致：返回 null
	writeJSON(w, http.StatusOK, Ok[any](nil))
}

// ============================================
// Unit 方法
// ============================================

// ListUnits 查询单元列表
func (h *UnitHandler) ListUnits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 构建过滤器
	// 空字符串视为 null（nil），nil 表示匹配 NULL 或未提供
	req := service.ListUnitsRequest{
		TenantID: tenantID,
		// branch_tag: 如果 query 参数不存在或为空字符串，设置为 nil（表示匹配 NULL）
		BranchName:  stringPtrOrNil(r.URL.Query().Get("branch_tag")),
		Building:   stringPtrOrNil(r.URL.Query().Get("building")),
		Floor:      stringPtrOrNil(r.URL.Query().Get("floor")),
		AreaName:    stringPtrOrNil(r.URL.Query().Get("area_name")),
		UnitNumber: stringPtrOrNil(r.URL.Query().Get("unit_number")),
		UnitName:   stringPtrOrNil(r.URL.Query().Get("unit_name")),
		UnitType:   stringPtrOrNil(r.URL.Query().Get("unit_type")),
		Search:     stringPtrOrNil(r.URL.Query().Get("search")),
		Page:       parseInt(r.URL.Query().Get("page"), 1),
		Size:       parseInt(r.URL.Query().Get("size"), 100),
	}

	resp, err := h.unitService.ListUnits(ctx, req)
	if err != nil {
		h.logger.Error("ListUnits failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换响应格式（与旧 Handler 一致）
	out := make([]any, 0, len(resp.Items))
	for _, u := range resp.Items {
		out = append(out, unitToJSON(u))
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": out,
		"total": resp.Total,
	}))
}

// GetUnit 获取单个单元详情
func (h *UnitHandler) GetUnit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	unitID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/units/")
	if unitID == "" || strings.Contains(unitID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	req := service.GetUnitRequest{
		TenantID: tenantID,
		UnitID:   unitID,
	}

	resp, err := h.unitService.GetUnit(ctx, req)
	if err != nil {
		h.logger.Error("GetUnit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(unitToJSON(resp.Unit)))
}

// CreateUnit 创建单元
func (h *UnitHandler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.CreateUnitRequest{
		TenantID:          tenantID,
		BranchName:         getString(payload, "branch_tag"),
		UnitName:          getString(payload, "unit_name"),
		Building:          getString(payload, "building"),
		Floor:             getString(payload, "floor"),
		AreaName:           getString(payload, "area_name"),
		UnitNumber:        getString(payload, "unit_number"),
		LayoutConfig:      getString(payload, "layout_config"),
		UnitType:          getString(payload, "unit_type"),
		IsPublicSpace:     getBool(payload, "is_public_space"),
		IsMultiPersonRoom: getBool(payload, "is_multi_person_room"),
		Timezone:          getString(payload, "timezone"),
	}

	resp, err := h.unitService.CreateUnit(ctx, req)
	if err != nil {
		h.logger.Error("CreateUnit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 unit 对象（与旧 Handler 格式一致）
	getReq := service.GetUnitRequest{
		TenantID: tenantID,
		UnitID:   resp.UnitID,
	}
	getResp, err := h.unitService.GetUnit(ctx, getReq)
	if err != nil {
		h.logger.Error("GetUnit after CreateUnit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(unitToJSON(getResp.Unit)))
}

// UpdateUnit 更新单元
func (h *UnitHandler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	unitID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/units/")
	if unitID == "" || strings.Contains(unitID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.UpdateUnitRequest{
		TenantID:          tenantID,
		UnitID:            unitID,
		BranchName:         getString(payload, "branch_tag"),
		UnitName:          getString(payload, "unit_name"),
		Building:          getString(payload, "building"),
		Floor:             getString(payload, "floor"),
		AreaName:           getString(payload, "area_name"),
		UnitNumber:        getString(payload, "unit_number"),
		LayoutConfig:      getString(payload, "layout_config"),
		UnitType:          getString(payload, "unit_type"),
		IsPublicSpace:     getBoolPtr(payload, "is_public_space"),
		IsMultiPersonRoom: getBoolPtr(payload, "is_multi_person_room"),
		Timezone:          getString(payload, "timezone"),
	}

	_, err := h.unitService.UpdateUnit(ctx, req)
	if err != nil {
		h.logger.Error("UpdateUnit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 unit 对象（与旧 Handler 格式一致）
	getReq := service.GetUnitRequest{
		TenantID: tenantID,
		UnitID:   unitID,
	}
	getResp, err := h.unitService.GetUnit(ctx, getReq)
	if err != nil {
		h.logger.Error("GetUnit after UpdateUnit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(unitToJSON(getResp.Unit)))
}

// DeleteUnit 删除单元
func (h *UnitHandler) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	unitID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/units/")
	if unitID == "" || strings.Contains(unitID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	req := service.DeleteUnitRequest{
		TenantID: tenantID,
		UnitID:   unitID,
	}

	_, err := h.unitService.DeleteUnit(ctx, req)
	if err != nil {
		h.logger.Error("DeleteUnit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 与旧 Handler 格式一致：返回 null
	writeJSON(w, http.StatusOK, Ok[any](nil))
}

// ============================================
// Room 方法
// ============================================

// ListRoomsWithBeds 查询房间及其床位列表
func (h *UnitHandler) ListRoomsWithBeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	unitID := r.URL.Query().Get("unit_id")
	if unitID == "" {
		writeJSON(w, http.StatusOK, Fail("unit_id is required"))
		return
	}

	req := service.ListRoomsWithBedsRequest{
		TenantID: tenantID,
		UnitID:   unitID,
	}

	resp, err := h.unitService.ListRoomsWithBeds(ctx, req)
	if err != nil {
		h.logger.Error("ListRoomsWithBeds failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换响应格式（与旧 Handler 一致）
	out := make([]any, 0, len(resp.Items))
	for _, rwb := range resp.Items {
		out = append(out, roomWithBedsToJSON(rwb))
	}

	writeJSON(w, http.StatusOK, Ok(out))
}

// CreateRoom 创建房间
func (h *UnitHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	unitID := getString(payload, "unit_id")
	if unitID == "" {
		writeJSON(w, http.StatusOK, Fail("unit_id is required"))
		return
	}

	req := service.CreateRoomRequest{
		TenantID:     tenantID,
		UnitID:       unitID,
		RoomName:     getString(payload, "room_name"),
		LayoutConfig: getString(payload, "layout_config"),
	}

	resp, err := h.unitService.CreateRoom(ctx, req)
	if err != nil {
		h.logger.Error("CreateRoom failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 room 对象（与旧 Handler 格式一致）
	getReq := service.GetRoomRequest{
		TenantID: tenantID,
		RoomID:   resp.RoomID,
	}
	getResp, err := h.unitService.GetRoom(ctx, getReq)
	if err != nil {
		h.logger.Error("GetRoom after CreateRoom failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(roomToJSON(getResp.Room)))
}

// UpdateRoom 更新房间
func (h *UnitHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/rooms/")
	if roomID == "" || strings.Contains(roomID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.UpdateRoomRequest{
		TenantID:     tenantID,
		RoomID:       roomID,
		RoomName:     getString(payload, "room_name"),
		LayoutConfig: getString(payload, "layout_config"),
	}

	_, err := h.unitService.UpdateRoom(ctx, req)
	if err != nil {
		h.logger.Error("UpdateRoom failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 room 对象（与旧 Handler 格式一致）
	getReq := service.GetRoomRequest{
		TenantID: tenantID,
		RoomID:   roomID,
	}
	getResp, err := h.unitService.GetRoom(ctx, getReq)
	if err != nil {
		h.logger.Error("GetRoom after UpdateRoom failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(roomToJSON(getResp.Room)))
}

// DeleteRoom 删除房间
func (h *UnitHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/rooms/")
	if roomID == "" || strings.Contains(roomID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	req := service.DeleteRoomRequest{
		TenantID: tenantID,
		RoomID:   roomID,
	}

	_, err := h.unitService.DeleteRoom(ctx, req)
	if err != nil {
		h.logger.Error("DeleteRoom failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 与旧 Handler 格式一致：返回 null
	writeJSON(w, http.StatusOK, Ok[any](nil))
}

// ============================================
// Bed 方法
// ============================================

// ListBeds 查询床位列表
func (h *UnitHandler) ListBeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		writeJSON(w, http.StatusOK, Fail("room_id is required"))
		return
	}

	req := service.ListBedsRequest{
		TenantID: tenantID,
		RoomID:   roomID,
	}

	resp, err := h.unitService.ListBeds(ctx, req)
	if err != nil {
		h.logger.Error("ListBeds failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换响应格式（与旧 Handler 一致）
	out := make([]any, 0, len(resp.Items))
	for _, b := range resp.Items {
		out = append(out, bedToJSON(b))
	}

	writeJSON(w, http.StatusOK, Ok(out))
}

// CreateBed 创建床位
func (h *UnitHandler) CreateBed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	roomID := getString(payload, "room_id")
	if roomID == "" {
		writeJSON(w, http.StatusOK, Fail("room_id is required"))
		return
	}

	req := service.CreateBedRequest{
		TenantID:         tenantID,
		RoomID:           roomID,
		BedName:          getString(payload, "bed_name"),
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
		MattressMaterial: getString(payload, "mattress_material"),
		MattressThickness: getString(payload, "mattress_thickness"),
	}

	resp, err := h.unitService.CreateBed(ctx, req)
	if err != nil {
		h.logger.Error("CreateBed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 bed 对象（与旧 Handler 格式一致）
	getReq := service.GetBedRequest{
		TenantID: tenantID,
		BedID:    resp.BedID,
	}
	getResp, err := h.unitService.GetBed(ctx, getReq)
	if err != nil {
		h.logger.Error("GetBed after CreateBed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(bedToJSON(getResp.Bed)))
}

// UpdateBed 更新床位
func (h *UnitHandler) UpdateBed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bedID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/beds/")
	if bedID == "" || strings.Contains(bedID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	req := service.UpdateBedRequest{
		TenantID:         tenantID,
		BedID:            bedID,
		BedName:          getString(payload, "bed_name"),
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
		MattressMaterial: getString(payload, "mattress_material"),
		MattressThickness: getString(payload, "mattress_thickness"),
	}

	_, err := h.unitService.UpdateBed(ctx, req)
	if err != nil {
		h.logger.Error("UpdateBed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 bed 对象（与旧 Handler 格式一致）
	getReq := service.GetBedRequest{
		TenantID: tenantID,
		BedID:    bedID,
	}
	getResp, err := h.unitService.GetBed(ctx, getReq)
	if err != nil {
		h.logger.Error("GetBed after UpdateBed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(bedToJSON(getResp.Bed)))
}

// DeleteBed 删除床位
func (h *UnitHandler) DeleteBed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bedID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/beds/")
	if bedID == "" || strings.Contains(bedID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	req := service.DeleteBedRequest{
		TenantID: tenantID,
		BedID:    bedID,
	}

	_, err := h.unitService.DeleteBed(ctx, req)
	if err != nil {
		h.logger.Error("DeleteBed failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 与旧 Handler 格式一致：返回 null
	writeJSON(w, http.StatusOK, Ok[any](nil))
}

// ============================================
// 辅助方法
// ============================================

// stringPtrOrNil 将空字符串转换为 nil，非空字符串转换为指针
// 用于区分"未提供"和"空字符串"（空字符串视为 null）
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// tenantIDFromReq 从请求中获取 tenant_id（复用 DeviceHandler 的逻辑）
func (h *UnitHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	if tid := r.URL.Query().Get("tenant_id"); tid != "" {
		return tid, true
	}
	// Prefer tenant header (owlFront axios injects it for all requests after login)
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" && tid != "null" {
		return tid, true
	}
	// Convenience: SystemAdmin without tenant header falls back to System tenant
	if strings.EqualFold(r.Header.Get("X-User-Role"), "SystemAdmin") {
		return SystemTenantID(), true
	}
	writeJSON(w, http.StatusOK, Fail("tenant_id is required"))
	return "", false
}

// 辅助函数：从 map 中获取字符串值
func getString(payload map[string]any, key string) string {
	if v, ok := payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		// 如果是 number 类型，转换为 string（用于 floor 字段）
		if num, ok := v.(float64); ok {
			return fmt.Sprintf("%.0f", num)
		}
		if num, ok := v.(int); ok {
			return fmt.Sprintf("%d", num)
		}
	}
	return ""
}

// 辅助函数：从 map 中获取布尔值
func getBool(payload map[string]any, key string) bool {
	if v, ok := payload[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// 辅助函数：从 map 中获取布尔值指针（用于可选字段）
func getBoolPtr(payload map[string]any, key string) *bool {
	if v, ok := payload[key]; ok {
		if b, ok := v.(bool); ok {
			return &b
		}
	}
	return nil
}

// 辅助函数：转换 Building 为 JSON
func buildingToJSON(b *domain.Building) map[string]any {
	m := map[string]any{
		"building_id":   b.BuildingID,
		"tenant_id":    b.TenantID,
		"building_name": b.BuildingName,
	}
	if b.BranchTag.Valid {
		m["branch_tag"] = b.BranchTag.String
	}
	if b.CreatedAt.Valid {
		m["created_at"] = b.CreatedAt.Time
	}
	if b.UpdatedAt.Valid {
		m["updated_at"] = b.UpdatedAt.Time
	}
	return m
}

// 辅助函数：转换 Unit 为 JSON（复用 repository.Unit.ToJSON 的逻辑）
func unitToJSON(u *domain.Unit) map[string]any {
	m := map[string]any{
		"unit_id":              u.UnitID,
		"tenant_id":            u.TenantID,
		"unit_name":            u.UnitName,
		"unit_number":          u.UnitNumber,
		"unit_type":            u.UnitType,
		"is_public_space":      u.IsPublicSpace,
		"is_multi_person_room": u.IsMultiPersonRoom,
		"timezone":             u.Timezone,
	}
	// building: 如果为 NULL，不包含在 JSON 中（前端会收到 undefined）
	if u.Building.Valid {
		m["building"] = u.Building.String
	}
	// floor: 如果为 NULL，不包含在 JSON 中（前端会收到 undefined）
	if u.Floor.Valid && u.Floor.String != "" {
		m["floor"] = u.Floor.String
	}
	if u.BranchName.Valid {
		m["branch_name"] = u.BranchName.String
	}
	if u.AreaName.Valid {
		m["area_name"] = u.AreaName.String
	}
	if u.LayoutConfig.Valid {
		m["layout_config"] = jsonRawOrString(u.LayoutConfig.String)
	}
	if u.GroupList.Valid {
		m["groupList"] = jsonRawOrString(u.GroupList.String)
	}
	if u.UserList.Valid {
		m["userList"] = jsonRawOrString(u.UserList.String)
	}
	return m
}

// 辅助函数：转换 Room 为 JSON
func roomToJSON(r *domain.Room) map[string]any {
	m := map[string]any{
		"room_id":   r.RoomID,
		"tenant_id": r.TenantID,
		"unit_id":   r.UnitID,
		"room_name": r.RoomName,
	}
	if r.LayoutConfig.Valid {
		m["layout_config"] = jsonRawOrString(r.LayoutConfig.String)
	}
	return m
}

// jsonRawOrString 辅助函数：尝试解析 JSON，如果成功返回 RawMessage，否则返回字符串
func jsonRawOrString(s string) any {
	if s == "" {
		return s
	}
	var tmp any
	if err := json.Unmarshal([]byte(s), &tmp); err == nil {
		return json.RawMessage([]byte(s))
	}
	return s
}

// 辅助函数：转换 RoomWithBeds 为 JSON
func roomWithBedsToJSON(rwb *repository.RoomWithBeds) map[string]any {
	m := roomToJSON(rwb.Room)
	beds := make([]any, 0, len(rwb.Beds))
	for _, bed := range rwb.Beds {
		beds = append(beds, bedToJSON(bed))
	}
	m["beds"] = beds
	return m
}

// 辅助函数：转换 Bed 为 JSON
func bedToJSON(b *domain.Bed) map[string]any {
	m := map[string]any{
		"bed_id":   b.BedID,
		"tenant_id": b.TenantID,
		"room_id":   b.RoomID,
		"bed_name":  b.BedName,
		// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	}
	if b.MattressMaterial.Valid {
		m["mattress_material"] = b.MattressMaterial.String
	}
	if b.MattressThickness.Valid {
		m["mattress_thickness"] = b.MattressThickness.String
	}
	return m
}

