package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"owl-common/card"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// DeviceStoreHandler 设备库存管理 Handler。业务逻辑在 DeviceStoreService，Handler 只做请求解析与响应。
type DeviceStoreHandler struct {
	deviceStoreRepo repository.DeviceStoreRepository
	deviceStoreSvc  *service.DeviceStoreService // Sleepace 绑定/InitialAll 等
	stateReader     *card.Reader
	db              *sql.DB
	logger          *zap.Logger
}

// NewDeviceStoreHandler 创建设备库存管理 Handler
func NewDeviceStoreHandler(deviceStoreRepo repository.DeviceStoreRepository, stateReader *card.Reader, db *sql.DB, logger *zap.Logger) *DeviceStoreHandler {
	return &DeviceStoreHandler{
		deviceStoreRepo: deviceStoreRepo,
		stateReader:     stateReader,
		db:              db,
		logger:          logger,
	}
}

// SetDeviceStoreService 注入设备库存 Service（Sleepace 绑定、InitialAll 等）
func (h *DeviceStoreHandler) SetDeviceStoreService(svc *service.DeviceStoreService) {
	h.deviceStoreSvc = svc
}

// ServeHTTP 实现 http.Handler 接口
func (h *DeviceStoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 权限检查：只允许 SystemAdmin 或 SystemOperator 访问设备库存管理
	userRole := r.Header.Get("X-User-Role")
	if userRole != "SystemAdmin" && userRole != "SystemOperator" {
		h.logger.Warn("DeviceStore access denied", zap.String("user_role", userRole), zap.String("path", r.URL.Path))
		writeJSON(w, http.StatusOK, Fail("permission denied: only SystemAdmin and SystemOperator can access device store"))
		return
	}

	// 路由分发
	switch {
	case r.URL.Path == "/admin/api/v1/device-store" && r.Method == http.MethodGet:
		h.ListDeviceStores(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/batch" && r.Method == http.MethodPut:
		h.BatchUpdateDeviceStores(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/import" && r.Method == http.MethodPost:
		h.ImportDeviceStores(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/import-template" && r.Method == http.MethodGet:
		h.GetImportTemplate(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/export" && r.Method == http.MethodGet:
		h.ExportDeviceStores(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/sync-sleepace-bind" && r.Method == http.MethodPost:
		h.SyncSleepaceBind(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/initial-all-sleepad" && r.Method == http.MethodPost:
		h.InitialAllSleepad(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/bind-sleepad" && r.Method == http.MethodPost:
		h.BindSleepadOne(w, r)
	case r.URL.Path == "/admin/api/v1/device-store/unbind-sleepad" && r.Method == http.MethodPost:
		h.UnbindSleepadOne(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/device-store/") && r.Method == http.MethodDelete:
		deviceUID := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/device-store/")
		deviceUID = strings.Trim(deviceUID, "/")
		if deviceUID == "" {
			writeJSON(w, http.StatusOK, Fail("device_uid is required"))
			return
		}
		h.DeleteDeviceStore(w, r, deviceUID)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ListDeviceStores 查询设备库存列表
// tenant_id 为可选过滤；不传时不过滤租户。SystemAdmin / SystemOperator 无 tenant 限制，可不传以查看全部。
func (h *DeviceStoreHandler) ListDeviceStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters := repository.DeviceStoreFilters{
		Search:     r.URL.Query().Get("search"),
		TenantID:   r.URL.Query().Get("tenant_id"),
		DeviceType: r.URL.Query().Get("device_type"),
	}
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 100)
	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	direction := strings.TrimSpace(r.URL.Query().Get("direction"))
	if direction != "desc" {
		direction = "asc"
	}

	items, total, err := h.deviceStoreRepo.ListDeviceStores(ctx, filters, page, size, sort, direction)
	if err != nil {
		h.logger.Error("ListDeviceStores failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list device stores: %v", err)))
		return
	}

	// 在线状态：只从 cardagg 读取（与 device_service.fillDeviceOnlineStatus 一致）
	onlineMap := service.FillDeviceOnlineStatusFromCardagg(ctx, h.stateReader, h.db, deviceStoreDeviceIDs(items), h.logger)
	for _, s := range items {
		s.OnlineStatus = onlineMap[s.DeviceID]
		if s.OnlineStatus == "" {
			s.OnlineStatus = "offline"
		}
	}

	out := make([]any, 0, len(items))
	for _, d := range items {
		out = append(out, d.ToJSON())
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": out,
		"total": total,
	}))
}

func deviceStoreDeviceIDs(stores []*domain.DeviceStore) []string {
	ids := make([]string, 0, len(stores))
	for _, s := range stores {
		if s.DeviceID != "" {
			ids = append(ids, s.DeviceID)
		}
	}
	return ids
}

// BatchUpdateDeviceStores 批量更新设备库存
func (h *DeviceStoreHandler) BatchUpdateDeviceStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload map[string]any
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	updatesRaw, ok := payload["updates"].([]any)
	if !ok {
		writeJSON(w, http.StatusOK, Fail("updates field is required and must be an array"))
		return
	}

	updates := make([]*domain.DeviceStore, 0, len(updatesRaw))
	for _, u := range updatesRaw {
		if m, ok := u.(map[string]any); ok {
			deviceUID, _ := m["device_uid"].(string)
			data, _ := m["data"].(map[string]any)
			if deviceUID != "" {
				updateItem := payloadToDeviceStorePatch(data)
				updateItem.DeviceUID = deviceUID
				updates = append(updates, updateItem)
			}
		}
	}

	if h.deviceStoreSvc == nil {
		writeJSON(w, http.StatusOK, Fail("device store service not configured"))
		return
	}
	if err := h.deviceStoreSvc.BatchUpdateDeviceStoresNotify(ctx, updates); err != nil {
		h.logger.Error("BatchUpdateDeviceStores failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update device stores: %v", err)))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": true,
		"updated": len(updates),
	}))
}

// GetImportTemplate 获取导入模板
func (h *DeviceStoreHandler) GetImportTemplate(w http.ResponseWriter, r *http.Request) {
	excelData, err := GenerateDeviceStoreImportTemplate()
	if err != nil {
		h.logger.Error("GenerateDeviceStoreImportTemplate failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to generate template: %v", err)))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=device-store-import-template.xlsx")
	w.WriteHeader(http.StatusOK)
	w.Write(excelData)
}

// ExportDeviceStores 导出设备库存
func (h *DeviceStoreHandler) ExportDeviceStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Query data with same filters
	filters := repository.DeviceStoreFilters{
		Search:     r.URL.Query().Get("search"),
		TenantID:   r.URL.Query().Get("tenant_id"),
		DeviceType: r.URL.Query().Get("device_type"),
	}

	items, _, err := h.deviceStoreRepo.ListDeviceStores(ctx, filters, 1, 1000, "", "asc")
	if err != nil {
		h.logger.Error("ListDeviceStores failed for export", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list device stores: %v", err)))
		return
	}
	onlineMap := service.FillDeviceOnlineStatusFromCardagg(ctx, h.stateReader, h.db, deviceStoreDeviceIDs(items), h.logger)
	for _, s := range items {
		s.OnlineStatus = onlineMap[s.DeviceID]
		if s.OnlineStatus == "" {
			s.OnlineStatus = "offline"
		}
	}

	// Convert to map format for Excel generation
	data := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data = append(data, item.ToJSON())
	}

	excelData, err := GenerateDeviceStoreExport(data)
	if err != nil {
		h.logger.Error("GenerateDeviceStoreExport failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to generate export: %v", err)))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=device-store-export.xlsx")
	w.WriteHeader(http.StatusOK)
	w.Write(excelData)
}

// ImportDeviceStores 导入设备库存
func (h *DeviceStoreHandler) ImportDeviceStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		writeJSON(w, http.StatusOK, Fail("failed to parse form"))
		return
	}

	file, fileHdr, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("file not found in request"))
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to read file"))
		return
	}

	rows, err := loadDeviceStoreImportRows(fileBytes, fileHdr.Filename)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	if len(rows) < 2 {
		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"success":       true,
			"total":         0,
			"success_count": 0,
			"failed_count":  0,
			"skipped_count": 0,
			"errors":        []any{},
			"skipped":       []any{},
		}))
		return
	}

	items := parseDeviceStoreImportRowsToItems(rows)

	if h.deviceStoreSvc == nil {
		writeJSON(w, http.StatusOK, Fail("device store service not configured"))
		return
	}
	successCount, inserted, skipped, errors, err := h.deviceStoreSvc.ImportDeviceStoresNotify(ctx, items)
	if err != nil {
		h.logger.Error("ImportDeviceStores failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to import: %v", err)))
		return
	}

	if len(inserted) > 0 {
		h.deviceStoreSvc.PostImportSleepadBind(ctx, inserted)
	}

	// Convert errors and skipped to JSON format
	errorsJSON := make([]any, 0, len(errors))
	for _, e := range errors {
		errorsJSON = append(errorsJSON, e.ToJSON())
	}
	skippedJSON := make([]any, 0, len(skipped))
	for _, s := range skipped {
		skippedJSON = append(skippedJSON, s.ToJSON())
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success":       true,
		"total":         len(items),
		"success_count": successCount,
		"failed_count":  len(errors),
		"skipped_count": len(skipped),
		"errors":        errorsJSON,
		"skipped":       skippedJSON,
	}))
}

// SyncSleepaceBind 先查 bindInfo，仅对未绑定的 Sleepad 执行绑定。逻辑在 DeviceStoreService。
func (h *DeviceStoreHandler) SyncSleepaceBind(w http.ResponseWriter, r *http.Request) {
	if h.deviceStoreSvc == nil {
		writeJSON(w, http.StatusOK, Fail("sleepace gateway not configured"))
		return
	}
	res, err := h.deviceStoreSvc.SyncSleepaceBind(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"synced": res.Synced, "skipped": res.Skipped, "failed": res.Failed, "errors": res.Errors,
	}))
}

// InitialAllSleepad 1) 查询、回写 2) 未绑定单独绑定。逻辑在 DeviceStoreService。
func (h *DeviceStoreHandler) InitialAllSleepad(w http.ResponseWriter, r *http.Request) {
	if h.deviceStoreSvc == nil {
		writeJSON(w, http.StatusOK, Fail("sleepace gateway not configured"))
		return
	}
	res, err := h.deviceStoreSvc.InitialAllSleepad(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"synced": res.Synced, "failed": res.Failed, "no_data": res.NoData, "no_data_list": res.NoDataList,
		"no_data_bound": res.NoDataBound, "no_data_bind_failed": res.NoDataBindFailed,
		"errors": res.Errors, "success_details": res.SuccessDetails,
	}))
}

// BindSleepadOne 单条 Sleepad 绑定。逻辑在 DeviceStoreService。
func (h *DeviceStoreHandler) BindSleepadOne(w http.ResponseWriter, r *http.Request) {
	if h.deviceStoreSvc == nil {
		writeJSON(w, http.StatusOK, Fail("sleepace gateway not configured"))
		return
	}
	var body struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DeviceID == "" {
		writeJSON(w, http.StatusOK, Fail("device_id required"))
		return
	}
	if err := h.deviceStoreSvc.BindSleepadOne(r.Context(), body.DeviceID); err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// UnbindSleepadOne 单条 Sleepad 解绑。逻辑在 DeviceStoreService。
func (h *DeviceStoreHandler) UnbindSleepadOne(w http.ResponseWriter, r *http.Request) {
	if h.deviceStoreSvc == nil {
		writeJSON(w, http.StatusOK, Fail("sleepace gateway not configured"))
		return
	}
	var body struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DeviceID == "" {
		writeJSON(w, http.StatusOK, Fail("device_id required"))
		return
	}
	if err := h.deviceStoreSvc.UnbindSleepadOne(r.Context(), body.DeviceID); err != nil {
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// DeleteDeviceStore 删除设备库存（仅允许未分配且未出库的记录）
func (h *DeviceStoreHandler) DeleteDeviceStore(w http.ResponseWriter, r *http.Request, deviceUID string) {
	ctx := r.Context()
	if err := h.deviceStoreRepo.DeleteDeviceStore(ctx, deviceUID); err != nil {
		h.logger.Warn("DeleteDeviceStore failed", zap.String("device_uid", deviceUID), zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}
