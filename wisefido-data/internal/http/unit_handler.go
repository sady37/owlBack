package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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
	case r.URL.Path == "/admin/api/v1/units/with-availability" && r.Method == http.MethodGet:
		h.ListUnitsWithAvailability(w, r)
	case r.URL.Path == "/admin/api/v1/units" && r.Method == http.MethodGet:
		h.ListUnits(w, r)
	case r.URL.Path == "/admin/api/v1/units/with-hierarchy" && r.Method == http.MethodGet:
		h.ListUnitsWithFullHierarchy(w, r)
	case r.URL.Path == "/admin/api/v1/units" && r.Method == http.MethodPost:
		h.CreateUnit(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/") && r.Method == http.MethodGet:
		h.GetUnit(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/") && r.Method == http.MethodPut:
		h.UpdateUnit(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/units/") && r.Method == http.MethodDelete:
		h.DeleteUnit(w, r)

	// Rooms
	case r.URL.Path == "/admin/api/v1/rooms/by-branch" && r.Method == http.MethodGet:
		h.ListRoomsByBranch(w, r)
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
	case r.URL.Path == "/admin/api/v1/beds-with-details" && r.Method == http.MethodGet:
		h.ListBedsWithDetails(w, r)
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

	// 优先使用 branch_id，如果没有则使用 branch_name（向后兼容 branch_tag）
	branchID := r.URL.Query().Get("branch_id")
	branchName := r.URL.Query().Get("branch_name")
	if branchName == "" {
		// 向后兼容：也支持 branch_tag
		branchName = r.URL.Query().Get("branch_tag")
	}
	if branchID != "" {
		// 如果提供了 branch_id，忽略 branch_name（service 层会通过 branch_id 查找）
		branchName = ""
	}

	req := service.ListBuildingsRequest{
		TenantID:   tenantID,
		BranchID:   branchID,
		BranchName: branchName,
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
	if buildingID == "" || isMultiSegmentPath(buildingID) {
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

	// 构建响应，包含 building 和 units
	response := buildingToJSON(resp.Building)
	// 添加 units 列表
	unitsJSON := make([]map[string]any, 0, len(resp.Units))
	for _, u := range resp.Units {
		unitMap := map[string]any{
			"unit_id":   u.UnitID,
			"unit_name": u.UnitName,
		}
		if u.Floor != "" {
			unitMap["floor"] = u.Floor
		}
		unitsJSON = append(unitsJSON, unitMap)
	}
	response["units"] = unitsJSON

	writeJSON(w, http.StatusOK, Ok(response))
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

	// branch_id 是必填项，不再支持 branch_name
	branchID := getString(payload, "branch_id")
	if branchID == "" {
		writeJSON(w, http.StatusOK, Fail("branch_id is required and cannot be empty"))
		return
	}

	req := service.CreateBuildingRequest{
		TenantID:     tenantID,
		BranchID:     branchID,
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

	// 构建响应，包含 building 和 units
	response := buildingToJSON(getResp.Building)
	// 添加 units 列表
	unitsJSON := make([]map[string]any, 0, len(getResp.Units))
	for _, u := range getResp.Units {
		unitMap := map[string]any{
			"unit_id":   u.UnitID,
			"unit_name": u.UnitName,
		}
		if u.Floor != "" {
			unitMap["floor"] = u.Floor
		}
		unitsJSON = append(unitsJSON, unitMap)
	}
	response["units"] = unitsJSON

	writeJSON(w, http.StatusOK, Ok(response))
}

// UpdateBuilding 更新楼栋
// 注意：只能修改 building_name，不能修改 branch_id
func (h *UnitHandler) UpdateBuilding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	buildingID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
	if buildingID == "" || isMultiSegmentPath(buildingID) {
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

	// UpdateBuilding 只能修改 building_name，不能修改 branch_id
	buildingName := getString(payload, "building_name")
	if buildingName == "" {
		writeJSON(w, http.StatusOK, Fail("building_name is required and cannot be empty"))
		return
	}

	req := service.UpdateBuildingRequest{
		TenantID:     tenantID,
		BuildingID:   buildingID,
		BuildingName: buildingName,
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

	// 构建响应，包含 building 和 units
	response := buildingToJSON(getResp.Building)
	// 添加 units 列表
	unitsJSON := make([]map[string]any, 0, len(getResp.Units))
	for _, u := range getResp.Units {
		unitMap := map[string]any{
			"unit_id":   u.UnitID,
			"unit_name": u.UnitName,
		}
		if u.Floor != "" {
			unitMap["floor"] = u.Floor
		}
		unitsJSON = append(unitsJSON, unitMap)
	}
	response["units"] = unitsJSON

	writeJSON(w, http.StatusOK, Ok(response))
}

// DeleteBuilding 删除楼栋
func (h *UnitHandler) DeleteBuilding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	buildingID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/buildings/")
	if buildingID == "" || isMultiSegmentPath(buildingID) {
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
	// 优先使用 branch_id，如果没有提供则使用 branch_name（向后兼容）
	branchID := r.URL.Query().Get("branch_id")
	branchName := ""
	if branchID == "" {
		// 如果 branch_id 未提供，则使用 branch_name（向后兼容）
		branchName = r.URL.Query().Get("branch_name")
	}

	// building_id: 优先使用 building_id，如果提供则忽略 building
	buildingID := r.URL.Query().Get("building_id")
	buildingName := ""
	if buildingID == "" {
		// 如果 building_id 未提供，则使用 building（向后兼容，通过 building_name 过滤）
		buildingName = r.URL.Query().Get("building")
	}

	residentIDParam := r.URL.Query().Get("resident_id")
	var residentIDPtr *string
	if r.URL.Query().Has("resident_id") {
		residentIDPtr = &residentIDParam
	}
	req := service.ListUnitsRequest{
		TenantID:   tenantID,
		BranchID:   stringPtrOrNil(branchID),
		BranchName: stringPtrOrNil(branchName),
		BuildingID: stringPtrOrNil(buildingID),
		Building:   stringPtrOrNil(buildingName),
		Floor:      stringPtrOrNil(r.URL.Query().Get("floor")),
		UnitName:   stringPtrOrNil(r.URL.Query().Get("unit_name")),
		UnitType:   stringPtrOrNil(r.URL.Query().Get("unit_type")),
		Search:     stringPtrOrNil(r.URL.Query().Get("search")),
		Page:       parseInt(r.URL.Query().Get("page"), 1),
		Size:       parseInt(r.URL.Query().Get("size"), 100),
		ResidentID: residentIDPtr,
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

// ListUnitsWithAvailability 查询 Units 并返回 has_available_bed、is_bound（供前端 (full) 灰行红字、橙/绿）
func (h *UnitHandler) ListUnitsWithAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	branchID := r.URL.Query().Get("branch_id")
	branchName := ""
	if branchID == "" {
		branchName = r.URL.Query().Get("branch_name")
	}
	buildingID := r.URL.Query().Get("building_id")
	buildingName := ""
	if buildingID == "" {
		buildingName = r.URL.Query().Get("building")
	}
	residentIDParam := r.URL.Query().Get("resident_id")
	var residentIDPtr *string
	if r.URL.Query().Has("resident_id") {
		residentIDPtr = &residentIDParam
	}
	req := service.ListUnitsRequest{
		TenantID:   tenantID,
		BranchID:   stringPtrOrNil(branchID),
		BranchName: stringPtrOrNil(branchName),
		BuildingID: stringPtrOrNil(buildingID),
		Building:   stringPtrOrNil(buildingName),
		Floor:      stringPtrOrNil(r.URL.Query().Get("floor")),
		UnitName:   stringPtrOrNil(r.URL.Query().Get("unit_name")),
		UnitType:   stringPtrOrNil(r.URL.Query().Get("unit_type")),
		Search:     stringPtrOrNil(r.URL.Query().Get("search")),
		Page:       parseInt(r.URL.Query().Get("page"), 1),
		Size:       parseInt(r.URL.Query().Get("size"), 100),
		ResidentID: residentIDPtr,
	}
	resp, err := h.unitService.ListUnitsWithAvailability(ctx, req)
	if err != nil {
		h.logger.Error("ListUnitsWithAvailability failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	out := make([]any, 0, len(resp.Items))
	for _, u := range resp.Items {
		m := unitToJSON(u.Unit)
		m["has_available_bed"] = u.HasAvailableBed
		m["is_bound"] = u.IsBound
		out = append(out, m)
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": out,
		"total": resp.Total,
	}))
}

// ListUnitsWithFullHierarchy 查询 Units 及其完整的层级结构（Rooms, Beds, Devices）
func (h *UnitHandler) ListUnitsWithFullHierarchy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 获取 user_id：从已验证的会话中读取（由 AuthMiddleware 注入到 context）
	currentUserID, _, _, _, ok := service.MustSession(ctx)
	if !ok || currentUserID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}

	// 获取 branch_id：优先从 query 参数，其次从 header（可选，如果提供则优先使用，否则 Service 层会从 user_branches 查询）
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		branchID = r.Header.Get("X-Branch-Id")
	}
	if branchID == "null" {
		branchID = ""
	}

	// 根据用户要求：查询时应该只有 tenant_id 和 branch_id，其他参数为空
	// branch_id 如果未提供，Service 层会从 user_branches 表根据 user_id 查询

	req := service.ListUnitsWithFullHierarchyRequest{
		TenantID:      tenantID,
		CurrentUserID: currentUserID,            // 传递 user_id，Service 层会查询用户的 branch_id
		BranchID:      stringPtrOrNil(branchID), // 可选，如果提供则优先使用
		BranchName:    nil,                      // 不使用 branch_name
		BuildingID:    nil,                      // 不使用 building_id
		Building:      nil,                      // 不使用 building
		Floor:         nil,                      // 不使用 floor
		UnitType:      nil,                      // 不使用 unit_type
		Search:        nil,                      // 不使用 search
	}

	resp, err := h.unitService.ListUnitsWithFullHierarchy(ctx, req)
	if err != nil {
		h.logger.Error("ListUnitsWithFullHierarchy failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换响应格式
	out := make([]any, 0, len(resp.Items))
	for _, unitWithHierarchy := range resp.Items {
		out = append(out, unitWithFullHierarchyToJSON(unitWithHierarchy))
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
	if unitID == "" || isMultiSegmentPath(unitID) || unitID == "with-hierarchy" {
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

	// 优先使用 branch_id，如果没有则使用 branch_name（向后兼容 branch_tag）
	branchID := getString(payload, "branch_id")
	branchName := getString(payload, "branch_name")
	if branchName == "" {
		// 向后兼容：也支持 branch_tag
		branchName = getString(payload, "branch_tag")
	}
	if branchID != "" {
		// 如果提供了 branch_id，忽略 branch_name（service 层会通过 branch_id 查找）
		branchName = ""
	}

	// 优先使用 building_id，如果没有则使用 building（向后兼容）
	buildingID := getString(payload, "building_id")
	buildingName := getString(payload, "building")
	if buildingID != "" {
		// 如果提供了 building_id，忽略 building（service 层会通过 building_id 查找）
		buildingName = ""
	}

	// v2 双维度：unit_property (0=Home, 1=Facility default) + unit_type (0=unknown, 1=single, 2=share, 3=public)
	// FE 直接发整数；不再用 is_public_space / is_shared_unit / unit_type 字符串
	req := service.CreateUnitRequest{
		TenantID:     tenantID,
		BranchID:     branchID,
		BranchName:   branchName,
		UnitName:     getString(payload, "unit_name"),
		BuildingID:   buildingID,
		BuildingName: buildingName,
		Floor:        getString(payload, "floor"),
		AreaName:     getString(payload, "area_name"),
		UnitNumber:   getString(payload, "unit_number"),
		LayoutConfig: getString(payload, "layout_config"),
		UnitProperty: getInt8(payload, "unit_property"),
		UnitType:     getInt8(payload, "unit_type"),
		Timezone:     getString(payload, "timezone"),
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
	if unitID == "" || isMultiSegmentPath(unitID) || unitID == "with-hierarchy" {
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

	// 优先使用 branch_id，如果没有则使用 branch_name（向后兼容 branch_tag）
	branchID := getString(payload, "branch_id")
	branchName := getString(payload, "branch_name")
	if branchName == "" {
		// 向后兼容：也支持 branch_tag
		branchName = getString(payload, "branch_tag")
	}
	if branchID != "" {
		// 如果提供了 branch_id，忽略 branch_name（service 层会通过 branch_id 查找）
		branchName = ""
	}

	// 优先使用 building_id，如果没有则使用 building（向后兼容）
	buildingID := getString(payload, "building_id")
	buildingName := getString(payload, "building")
	if buildingID != "" {
		// 如果提供了 building_id，忽略 building（service 层会通过 building_id 查找）
		buildingName = ""
	}

	req := service.UpdateUnitRequest{
		TenantID:     tenantID,
		UnitID:       unitID,
		BranchID:     branchID,
		BranchName:   branchName,
		UnitName:     getString(payload, "unit_name"),
		BuildingID:   buildingID,
		BuildingName: buildingName,
		Floor:        getString(payload, "floor"),
		AreaName:     getString(payload, "area_name"),
		UnitNumber:   getString(payload, "unit_number"),
		LayoutConfig: getString(payload, "layout_config"),
		UnitProperty: getInt8Ptr(payload, "unit_property"),
		UnitType:     getInt8Ptr(payload, "unit_type"),
		Timezone:     getString(payload, "timezone"),
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
	if unitID == "" || isMultiSegmentPath(unitID) || unitID == "with-hierarchy" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 获取 user_id：从 header 获取（用于权限验证，Service 层会从 user_branches 表查询用户的 branch_id）
	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required for permission validation"))
		return
	}

	// 获取 branch_id：可选，如果提供则验证用户是否有权限访问该 branch
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		branchID = r.Header.Get("X-Branch-Id")
	}
	if branchID == "null" {
		branchID = ""
	}

	req := service.DeleteUnitRequest{
		TenantID:      tenantID,
		UnitID:        unitID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会查询用户的 branch_id
		BranchID:      branchID,      // 可选，如果提供则验证用户是否有权限访问该 branch
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

	// 获取 user_id：从 header 获取（可选，用于日志记录）
	currentUserID := r.Header.Get("X-User-Id")

	search := r.URL.Query().Get("search")
	if search == "null" {
		search = ""
	}
	residentID := strings.TrimSpace(r.URL.Query().Get("resident_id"))

	req := service.ListRoomsWithBedsRequest{
		TenantID:      tenantID,
		UnitID:        unitID,
		ResidentID:    residentID,
		CurrentUserID: currentUserID,
		BranchID:      "",
		Search:        search,
	}

	resp, err := h.unitService.ListRoomsWithBeds(ctx, req)
	if err != nil {
		h.logger.Error("ListRoomsWithBeds failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换响应格式（与旧 Handler 一致），含 device_types 供前端 RoomName(R) 展示
	out := make([]any, 0, len(resp.Items))
	for _, item := range resp.Items {
		out = append(out, roomWithBedsItemToJSON(item))
	}

	writeJSON(w, http.StatusOK, Ok(out))
}

// ListRoomsByBranch 按 branch 列出所有 room（带 is_full、is_bound、facility_type）
func (h *UnitHandler) ListRoomsByBranch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}
	branchID := strings.TrimSpace(r.URL.Query().Get("branch_id"))
	if branchID == "" {
		writeJSON(w, http.StatusOK, Fail("branch_id is required"))
		return
	}
	req := service.ListRoomsByBranchRequest{TenantID: tenantID, BranchID: branchID}
	resp, err := h.unitService.ListRoomsByBranch(ctx, req)
	if err != nil {
		h.logger.Error("ListRoomsByBranch failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	out := make([]any, 0, len(resp.Items))
	for _, item := range resp.Items {
		m := map[string]any{
			"room_id":        item.RoomID,
			"tenant_id":      item.TenantID,
			"unit_id":        item.UnitID,
			"unit_name":      item.UnitName,
			"building_name":  item.BuildingName,
			"floor":          item.Floor,
			"room_name":      item.RoomName,
			"unit_type":      item.UnitType,
			"facility_type":  item.FacilityType,
			"is_full":        item.IsFull,
			"is_bound":       item.IsBound,
		}
		if len(item.DeviceTypes) > 0 {
			m["device_types"] = item.DeviceTypes
		}
		out = append(out, m)
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

	// 获取 user_id：从 header 获取（可选，用于日志记录）
	currentUserID := r.Header.Get("X-User-Id")

	req := service.CreateRoomRequest{
		TenantID:      tenantID,
		UnitID:        unitID,
		CurrentUserID: currentUserID, // 可选，用于日志记录
		BranchID:      "",            // 不再使用
		RoomName:      getString(payload, "room_name"),
		LayoutConfig:  getString(payload, "layout_config"),
	}

	resp, err := h.unitService.CreateRoom(ctx, req)
	if err != nil {
		h.logger.Error("CreateRoom failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 room 对象（与旧 Handler 格式一致）
	getReq := service.GetRoomRequest{
		TenantID:      tenantID,
		RoomID:        resp.RoomID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会查询用户的所有 branch_id
		BranchID:      "",            // 不再使用，Service 层会通过 CurrentUserID 查找用户的所有 branch_id
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
	if roomID == "" || isMultiSegmentPath(roomID) {
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

	// 获取 user_id：从 header 获取（用于权限验证，Service 层会从 user_branches 表查询用户的 branch_id）
	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required for permission validation"))
		return
	}

	// 获取 branch_id：可选，如果提供则验证用户是否有权限访问该 branch
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		branchID = r.Header.Get("X-Branch-Id")
	}
	if branchID == "null" {
		branchID = ""
	}

	req := service.UpdateRoomRequest{
		TenantID:      tenantID,
		RoomID:        roomID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会查询用户的 branch_id
		BranchID:      branchID,      // 可选，如果提供则验证用户是否有权限访问该 branch
		RoomName:      getString(payload, "room_name"),
		LayoutConfig:  getString(payload, "layout_config"),
	}

	_, err := h.unitService.UpdateRoom(ctx, req)
	if err != nil {
		h.logger.Error("UpdateRoom failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 获取完整的 room 对象（与旧 Handler 格式一致）
	getReq := service.GetRoomRequest{
		TenantID:      tenantID,
		RoomID:        roomID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会查询用户的 branch_id
		BranchID:      branchID,      // 可选，如果提供则验证用户是否有权限访问该 branch
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
	if roomID == "" || isMultiSegmentPath(roomID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 获取 user_id：从 header 获取（用于权限验证，Service 层会从 user_branches 表查询用户的 branch_id）
	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required for permission validation"))
		return
	}

	// 获取 branch_id：可选，如果提供则验证用户是否有权限访问该 branch
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		branchID = r.Header.Get("X-Branch-Id")
	}
	if branchID == "null" {
		branchID = ""
	}

	req := service.DeleteRoomRequest{
		TenantID:      tenantID,
		RoomID:        roomID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会查询用户的 branch_id
		BranchID:      branchID,      // 可选，如果提供则验证用户是否有权限访问该 branch
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

	// 获取 user_id：从 header 获取（可选，用于日志记录）
	currentUserID := r.Header.Get("X-User-Id")

	search := r.URL.Query().Get("search")
	if search == "null" {
		search = ""
	}
	var residentIDPtr *string
	if r.URL.Query().Has("resident_id") {
		v := strings.TrimSpace(r.URL.Query().Get("resident_id"))
		residentIDPtr = &v
	}

	req := service.ListBedsRequest{
		TenantID:      tenantID,
		RoomID:        roomID,
		ResidentID:    residentIDPtr,
		CurrentUserID: currentUserID,
		BranchID:      "",
		Search:        search,
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

// ListBedsWithDetails 返回床位列表（含 resident_id、设备类型及 monitor 状态）
func (h *UnitHandler) ListBedsWithDetails(w http.ResponseWriter, r *http.Request) {
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
	search := r.URL.Query().Get("search")
	if search == "null" {
		search = ""
	}
	var residentIDPtr *string
	if r.URL.Query().Has("resident_id") {
		v := strings.TrimSpace(r.URL.Query().Get("resident_id"))
		residentIDPtr = &v
	}
	req := service.ListBedsWithDetailsRequest{
		TenantID:   tenantID,
		RoomID:     roomID,
		ResidentID: residentIDPtr,
		Search:     search,
	}
	resp, err := h.unitService.ListBedsWithDetails(ctx, req)
	if err != nil {
		h.logger.Error("ListBedsWithDetails failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	out := make([]any, 0, len(resp.Items))
	for _, it := range resp.Items {
		m := map[string]any{"bed_id": it.BedID, "bed_name": it.BedName, "devices": it.Devices}
		if it.ResidentID != nil {
			m["resident_id"] = *it.ResidentID
		}
		out = append(out, m)
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

	// 获取 user_id 和 role：从 header 获取（用于权限验证）
	currentUserID := r.Header.Get("X-User-Id")
	currentUserRole := r.Header.Get("X-User-Role")

	req := service.CreateBedRequest{
		TenantID:      tenantID,
		RoomID:        roomID,
		CurrentUserID: currentUserID, // 可选，用于日志记录
		BranchID:      "",            // 不再使用
		BedName:       getString(payload, "bed_name"),
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
		MattressMaterial:  getString(payload, "mattress_material"),
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
		TenantID:        tenantID,
		BedID:           resp.BedID,
		CurrentUserID:   currentUserID,   // 必填，用于权限验证
		CurrentUserRole: currentUserRole, // 用于权限验证，检查权限 scope
		BranchID:        "",              // 不再使用
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
	if bedID == "" || isMultiSegmentPath(bedID) {
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

	// 获取 user_id：从 header 获取（用于权限验证，Service 层会从 user_branches 表查询用户的 branch_id）
	currentUserID := r.Header.Get("X-User-Id")
	if currentUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user ID is required for permission validation"))
		return
	}

	// 获取 branch_id：可选，如果提供则验证用户是否有权限访问该 branch
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		branchID = r.Header.Get("X-Branch-Id")
	}
	if branchID == "null" {
		branchID = ""
	}

	req := service.UpdateBedRequest{
		TenantID:      tenantID,
		BedID:         bedID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会查询用户的 branch_id
		BranchID:      branchID,      // 可选，如果提供则验证用户是否有权限访问该 branch
		BedName:       getString(payload, "bed_name"),
		// 注意：BedType 字段已删除，ActiveBed 判断由应用层动态计算
		MattressMaterial:  getString(payload, "mattress_material"),
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
		TenantID:      tenantID,
		BedID:         bedID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会查询用户的 branch_id
		BranchID:      branchID,      // 可选，如果提供则验证用户是否有权限访问该 branch
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
// DeleteBed 删除床位
func (h *UnitHandler) DeleteBed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bedID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/beds/")
	if bedID == "" || isMultiSegmentPath(bedID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 获取 user_id：从 header 获取（用于日志记录，Service 层不再强制用于权限验证）
	currentUserID := r.Header.Get("X-User-Id")
	// currentUserID is optional for logging, not for permission validation here

	req := service.DeleteBedRequest{
		TenantID:      tenantID,
		BedID:         bedID,
		CurrentUserID: currentUserID, // 传递 user_id，Service 层会用于日志记录（可选）
		BranchID:      "",            // 不再使用，Service 层已简化权限验证逻辑
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

// 辅助函数：从 map 中获取 int8（用于 unit_property/unit_type 等枚举字段）
// 兼容数字（float64/int）和数字字符串
func getInt8(payload map[string]any, key string) int8 {
	if v, ok := payload[key]; ok {
		switch n := v.(type) {
		case float64:
			return int8(n)
		case int:
			return int8(n)
		case int64:
			return int8(n)
		case string:
			if n == "" {
				return 0
			}
			var i int
			_, _ = fmt.Sscanf(n, "%d", &i)
			return int8(i)
		}
	}
	return 0
}

// 辅助函数：从 map 中获取 int8 指针（可选字段）
func getInt8Ptr(payload map[string]any, key string) *int8 {
	if _, ok := payload[key]; !ok {
		return nil
	}
	v := getInt8(payload, key)
	return &v
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

// 辅助函数：从 map 中获取布尔值指针（支持多个 key，按优先级查找）
func getBoolPtrFromMultipleKeys(payload map[string]any, keys ...string) *bool {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if b, ok := v.(bool); ok {
				return &b
			}
		}
	}
	return nil
}

// 辅助函数：转换 Building 为 JSON
func buildingToJSON(b *domain.Building) map[string]any {
	m := map[string]any{
		"building_id":   b.BuildingID,
		"tenant_id":     b.TenantID,
		"building_name": b.BuildingName,
	}
	// 返回 branch_id（前端需要 ID 来选择对象）
	if b.BranchID.Valid {
		m["branch_id"] = b.BranchID.String
	}
	// 返回 branch_name（前端需要名称来显示，如 "Branch_name-Building_name"）
	if b.BranchName.Valid {
		m["branch_name"] = b.BranchName.String
	}
	// 注意：不再返回 created_at 和 updated_at 字段
	return m
}

// 辅助函数：转换 Unit 为 JSON（复用 repository.Unit.ToJSON 的逻辑）
func unitToJSON(u *domain.Unit) map[string]any {
	// v2 双维度：FE 直接消费 unit_property + unit_type 整数 enum
	m := map[string]any{
		"unit_id":       u.UnitID,
		"tenant_id":     u.TenantID,
		"unit_name":     u.UnitName,
		"unit_property": u.UnitProperty, // 0=Home, 1=Facility
		"unit_type":     u.UnitType,     // 0=unknown, 1=VIP, 2=Share, 3=Public
		"timezone":      u.Timezone,
	}
	// 返回 branch_id（前端需要 ID 来选择对象）
	if u.BranchID.Valid {
		m["branch_id"] = u.BranchID.String
	}
	// 返回 branch_name（用于显示）
	if u.BranchName.Valid {
		m["branch_name"] = u.BranchName.String
	}
	// 返回 building_id（前端需要 ID 来进行业务逻辑处理）
	if u.BuildingID.Valid {
		m["building_id"] = u.BuildingID.String
	}
	// building_name: 如果为 NULL，不包含在 JSON 中（前端会收到 undefined）
	if u.BuildingName.Valid {
		m["building_name"] = u.BuildingName.String
	}
	// floor: 如果为 NULL，不包含在 JSON 中（前端会收到 undefined）
	// 将 "1F" 格式转换为数字（提取数字部分）
	if u.Floor.Valid && u.Floor.String != "" {
		// 提取数字部分（如 "1F" -> 1, "2F" -> 2）
		floorStr := u.Floor.String
		var floorNum int
		if matched := regexp.MustCompile(`\d+`).FindString(floorStr); matched != "" {
			if n, err := strconv.Atoi(matched); err == nil {
				floorNum = n
			}
		}
		if floorNum > 0 {
			m["floor"] = floorNum
		}
	}
	if u.LayoutConfig.Valid {
		m["layout_config"] = jsonRawOrString(u.LayoutConfig.String)
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
	// 返回 unit_name（前端需要名称来显示）
	if r.UnitName.Valid {
		m["unit_name"] = r.UnitName.String
	}
	// 返回 floor（从 unit 继承，转换为数字类型，默认值为 1）
	if r.Floor.Valid && r.Floor.String != "" {
		// 从 "1F" 格式提取数字，如 "1F" -> 1, "2F" -> 2
		floorNum := parseFloorToNumber(r.Floor.String)
		if floorNum > 0 {
			m["floor"] = floorNum
		} else {
			m["floor"] = 1 // 默认值
		}
	} else {
		m["floor"] = 1 // 默认值
	}
	if r.LayoutConfig.Valid {
		m["layout_config"] = jsonRawOrString(r.LayoutConfig.String)
	}
	return m
}

// parseFloorToNumber 从 "1F" 格式提取数字，如 "1F" -> 1, "2F" -> 2
func parseFloorToNumber(floorStr string) int {
	if floorStr == "" {
		return 1
	}
	// 提取数字部分
	re := regexp.MustCompile(`\d+`)
	matches := re.FindString(floorStr)
	if matches == "" {
		return 1
	}
	num, err := strconv.Atoi(matches)
	if err != nil {
		return 1
	}
	return num
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

// roomWithBedsItemToJSON 转换 RoomWithBedsItem 为 JSON，含 room 级 device_types、每 bed 的 devices（letter + monitoring_enabled）
func roomWithBedsItemToJSON(item *service.RoomWithBedsItem) map[string]any {
	m := roomToJSON(item.Room)
	beds := make([]any, 0, len(item.Beds))
	for _, bwd := range item.Beds {
		bm := bedToJSON(bwd.Bed)
		devs := make([]any, 0, len(bwd.Devices))
		for _, d := range bwd.Devices {
			devs = append(devs, map[string]any{"letter": d.Letter, "monitoring_enabled": d.MonitoringEnabled})
		}
		bm["devices"] = devs
		beds = append(beds, bm)
	}
	m["beds"] = beds
	if len(item.DeviceTypes) > 0 {
		m["device_types"] = item.DeviceTypes
	}
	return m
}

// 辅助函数：转换 Bed 为 JSON
func bedToJSON(b *domain.Bed) map[string]any {
	m := map[string]any{
		"bed_id":    b.BedID,
		"tenant_id": b.TenantID,
		"room_id":   b.RoomID,
		"bed_name":  b.BedName,
		// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	}
	// 返回 room_name（前端需要名称来显示）
	if b.RoomName.Valid {
		m["room_name"] = b.RoomName.String
	}
	if b.MattressMaterial.Valid {
		m["mattress_material"] = b.MattressMaterial.String
	}
	if b.MattressThickness.Valid {
		m["mattress_thickness"] = b.MattressThickness.String
	}
	return m
}

// ============================================
// ListUnitsWithFullHierarchy 相关 JSON 转换函数
// ============================================

// unitWithFullHierarchyToJSON 转换 UnitWithFullHierarchy 为 JSON
func unitWithFullHierarchyToJSON(unitWithHierarchy *service.UnitWithFullHierarchy) map[string]any {
	// 先转换 Unit 基本信息
	m := unitToJSON(unitWithHierarchy.Unit)

	// 添加 Rooms 数组
	rooms := make([]any, 0, len(unitWithHierarchy.Rooms))
	for _, roomWithBeds := range unitWithHierarchy.Rooms {
		rooms = append(rooms, roomWithBedsAndDevicesToJSON(roomWithBeds))
	}
	m["rooms"] = rooms

	return m
}

// roomWithBedsAndDevicesToJSON 转换 RoomWithBedsAndDevices 为 JSON
func roomWithBedsAndDevicesToJSON(roomWithBeds *service.RoomWithBedsAndDevices) map[string]any {
	// 先转换 Room 基本信息
	m := roomToJSON(roomWithBeds.Room)

	// 添加 Beds 数组
	beds := make([]any, 0, len(roomWithBeds.Beds))
	for _, bedWithDevices := range roomWithBeds.Beds {
		beds = append(beds, bedWithDevicesToJSON(bedWithDevices))
	}
	m["beds"] = beds

	// 添加 Device IDs 和 Names
	m["device_ids"] = roomWithBeds.DeviceIDs
	m["device_names"] = roomWithBeds.DeviceNames

	return m
}

// bedWithDevicesToJSON 转换 BedWithDevices 为 JSON
func bedWithDevicesToJSON(bedWithDevices *service.BedWithDevices) map[string]any {
	// 先转换 Bed 基本信息
	m := bedToJSON(bedWithDevices.Bed)

	// 添加 Device IDs 和 Names
	m["device_ids"] = bedWithDevices.DeviceIDs
	m["device_names"] = bedWithDevices.DeviceNames

	return m
}
