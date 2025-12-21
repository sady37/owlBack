package httpapi

import (
	"net/http"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// TagsHandler 标签管理 Handler
type TagsHandler struct {
	tagService *service.TagService
	logger     *zap.Logger
}

// NewTagsHandler 创建标签管理 Handler
func NewTagsHandler(tagService *service.TagService, logger *zap.Logger) *TagsHandler {
	return &TagsHandler{
		tagService: tagService,
		logger:     logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *TagsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/tags" && r.Method == http.MethodGet:
		h.ListTags(w, r)
	case r.URL.Path == "/admin/api/v1/tags" && r.Method == http.MethodPost:
		h.CreateTag(w, r)
	case r.URL.Path == "/admin/api/v1/tags" && r.Method == http.MethodDelete:
		h.DeleteTag(w, r)
	case r.URL.Path == "/admin/api/v1/tags/types" && r.Method == http.MethodDelete:
		h.DeleteTagType(w, r)
	case r.URL.Path == "/admin/api/v1/tags/for-object" && r.Method == http.MethodGet:
		h.GetTagsForObject(w, r)
	case strings.HasSuffix(r.URL.Path, "/objects") && r.Method == http.MethodPost:
		h.AddTagObjects(w, r)
	case strings.HasSuffix(r.URL.Path, "/objects") && r.Method == http.MethodDelete:
		h.RemoveTagObjects(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/tags/") && r.Method == http.MethodPut:
		h.UpdateTag(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ListTags 查询标签列表
func (h *TagsHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")
	tagType := strings.TrimSpace(r.URL.Query().Get("tag_type"))
	includeSystemStr := r.URL.Query().Get("include_system_tag_types")
	includeSystem := includeSystemStr != "false" // 默认为 true
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 20)

	// 2. 调用 Service
	req := service.ListTagsRequest{
		TenantID:          tenantID,
		UserRole:          userRole,
		TagType:           tagType,
		IncludeSystemTags: includeSystem,
		Page:              page,
		Size:              size,
	}

	resp, err := h.tagService.ListTags(ctx, req)
	if err != nil {
		h.logger.Error("ListTags failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// CreateTag 创建标签
func (h *TagsHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		TagName string `json:"tag_name"`
		TagType string `json:"tag_type"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 调用 Service
	req := service.CreateTagRequest{
		TenantID: tenantID,
		UserRole: userRole,
		TagName:  payload.TagName,
		TagType:  payload.TagType,
	}

	resp, err := h.tagService.CreateTag(ctx, req)
	if err != nil {
		h.logger.Error("CreateTag failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// UpdateTag 更新标签名称
func (h *TagsHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	tagID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/tags/")
	if tagID == "" || strings.Contains(tagID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 2. 调用 Service
	req := service.UpdateTagRequest{
		TenantID: tenantID,
		UserRole: userRole,
		TagID:    tagID,
		TagName:  payload.TagName,
	}

	err := h.tagService.UpdateTag(ctx, req)
	if err != nil {
		h.logger.Error("UpdateTag failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// DeleteTag 删除标签
func (h *TagsHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")
	tagName := strings.TrimSpace(r.URL.Query().Get("tag_name"))
	if tagName == "" {
		writeJSON(w, http.StatusOK, Fail("tag_name is required"))
		return
	}

	// 2. 调用 Service
	req := service.DeleteTagRequest{
		TenantID: tenantID,
		UserRole: userRole,
		TagName:  tagName,
	}

	err := h.tagService.DeleteTag(ctx, req)
	if err != nil {
		h.logger.Error("DeleteTag failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// DeleteTagType 删除标签类型
func (h *TagsHandler) DeleteTagType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		TagType string `json:"tag_type"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	if payload.TagType == "" {
		writeJSON(w, http.StatusOK, Fail("tag_type is required"))
		return
	}

	// 2. 调用 Service
	req := service.DeleteTagTypeRequest{
		TenantID: tenantID,
		UserRole: userRole,
		TagType:  payload.TagType,
	}

	err := h.tagService.DeleteTagType(ctx, req)
	if err != nil {
		h.logger.Error("DeleteTagType failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// AddTagObjects 添加标签对象
func (h *TagsHandler) AddTagObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/tags/")
	tagID := strings.TrimSuffix(path, "/objects")
	if tagID == "" || strings.Contains(tagID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		ObjectType string              `json:"object_type"`
		Objects    []service.TagObject `json:"objects"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	if payload.ObjectType == "" || len(payload.Objects) == 0 {
		writeJSON(w, http.StatusOK, Fail("object_type and objects are required"))
		return
	}

	// 2. 调用 Service
	req := service.AddTagObjectsRequest{
		TenantID:   tenantID,
		UserRole:   userRole,
		TagID:      tagID,
		ObjectType: payload.ObjectType,
		Objects:    payload.Objects,
	}

	err := h.tagService.AddTagObjects(ctx, req)
	if err != nil {
		h.logger.Error("AddTagObjects failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// RemoveTagObjects 删除标签对象
func (h *TagsHandler) RemoveTagObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析
	path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/tags/")
	tagID := strings.TrimSuffix(path, "/objects")
	if tagID == "" || strings.Contains(tagID, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	userRole := r.Header.Get("X-User-Role")

	var payload struct {
		ObjectType string                    `json:"object_type"`
		ObjectIDs  []string                  `json:"object_ids"`
		Objects    []service.TagObject `json:"objects"`
	}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	if payload.ObjectType == "" {
		writeJSON(w, http.StatusOK, Fail("object_type is required"))
		return
	}

	if len(payload.ObjectIDs) == 0 && len(payload.Objects) == 0 {
		writeJSON(w, http.StatusOK, Fail("object_ids or objects is required"))
		return
	}

	// 2. 调用 Service
	req := service.RemoveTagObjectsRequest{
		TenantID:   tenantID,
		UserRole:   userRole,
		TagID:      tagID,
		ObjectType: payload.ObjectType,
		ObjectIDs:  payload.ObjectIDs,
		Objects:    payload.Objects,
	}

	err := h.tagService.RemoveTagObjects(ctx, req)
	if err != nil {
		h.logger.Error("RemoveTagObjects failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// GetTagsForObject 查询对象标签
// 从源表查询标签（tag_objects 字段已删除）：
// - user: 从 users.tags JSONB 字段查询
// - resident: 从 residents.family_tag 查询
// - unit: 从 units.branch_tag 和 units.area_tag 查询
func (h *TagsHandler) GetTagsForObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 参数解析和验证
	tenantID, ok := h.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	objectType := r.URL.Query().Get("object_type")
	objectID := r.URL.Query().Get("object_id")
	if objectType == "" || objectID == "" {
		writeJSON(w, http.StatusOK, Fail("object_type and object_id are required"))
		return
	}

	// 2. 调用 Service
	req := service.GetTagsForObjectRequest{
		TenantID:   tenantID,
		ObjectType: objectType,
		ObjectID:   objectID,
	}

	resp, err := h.tagService.GetTagsForObject(ctx, req)
	if err != nil {
		h.logger.Error("GetTagsForObject failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 3. 返回响应
	writeJSON(w, http.StatusOK, Ok(resp))
}

// tenantIDFromReq 从请求中获取 tenant_id
func (h *TagsHandler) tenantIDFromReq(w http.ResponseWriter, r *http.Request) (string, bool) {
	// 复用 StubHandler 的逻辑
	if tid := r.URL.Query().Get("tenant_id"); tid != "" && tid != "null" {
		return tid, true
	}
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

