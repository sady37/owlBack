package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// DeviceStoreHandler 设备库存管理 Handler
// 注意：根据架构设计，DeviceStore 不需要 Service 层（业务逻辑简单），直接使用 Repository
type DeviceStoreHandler struct {
	deviceStoreRepo repository.DeviceStoreRepository
	logger          *zap.Logger
}

// NewDeviceStoreHandler 创建设备库存管理 Handler
func NewDeviceStoreHandler(deviceStoreRepo repository.DeviceStoreRepository, logger *zap.Logger) *DeviceStoreHandler {
	return &DeviceStoreHandler{
		deviceStoreRepo: deviceStoreRepo,
		logger:          logger,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *DeviceStoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// ListDeviceStores 查询设备库存列表
func (h *DeviceStoreHandler) ListDeviceStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters := repository.DeviceStoreFilters{
		Search:     r.URL.Query().Get("search"),
		TenantID:   r.URL.Query().Get("tenant_id"),
		DeviceType: r.URL.Query().Get("device_type"),
	}
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("size"), 100)

	items, total, err := h.deviceStoreRepo.ListDeviceStores(ctx, filters, page, size)
	if err != nil {
		h.logger.Error("ListDeviceStores failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list device stores: %v", err)))
		return
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
			deviceStoreID, _ := m["device_store_id"].(string)
			data, _ := m["data"].(map[string]any)
			if deviceStoreID != "" {
				updateItem := payloadToDeviceStore(data)
				updateItem.DeviceStoreID = deviceStoreID
				updates = append(updates, updateItem)
			}
		}
	}

	if err := h.deviceStoreRepo.BatchUpdateDeviceStores(ctx, updates); err != nil {
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

	items, _, err := h.deviceStoreRepo.ListDeviceStores(ctx, filters, 1, 10000)
	if err != nil {
		h.logger.Error("ListDeviceStores failed for export", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list device stores: %v", err)))
		return
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

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("file not found in request"))
		return
	}
	defer file.Close()

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to read file"))
		return
	}

	// Parse Excel file
	f, err := excelize.OpenReader(&bytesReader{data: fileBytes})
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to parse Excel file: %v", err)))
		return
	}
	defer f.Close()

	// Read first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		writeJSON(w, http.StatusOK, Fail("Excel file has no sheets"))
		return
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to read rows: %v", err)))
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

	// Parse header row
	headerRow := rows[0]
	headerMap := make(map[string]int)
	for i, h := range headerRow {
		headerMap[h] = i
	}

	// Map Excel header names to database field names
	headerToFieldMap := map[string]string{
		"Device Type":                 "device_type",
		"Device Model":                "device_model",
		"Serial Number":               "serial_number",
		"UID":                         "uid",
		"IMEI":                        "imei",
		"Comm Mode":                   "comm_mode",
		"MCU Model":                   "mcu_model",
		"Firmware Version":            "firmware_version",
		"OTA Target Firmware Version": "ota_target_firmware_version",
		"OTA Target MCU Model":        "ota_target_mcu_model",
		"Tenant ID":                   "tenant_id",
		"Tenant Name":                 "tenant_name",
		"Allow Access":                "allow_access",
		"Import Date":                 "import_date",
		"Allocate Time":               "allocate_time",
	}

	// Parse data rows
	items := make([]*domain.DeviceStore, 0, len(rows)-1)
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		itemMap := make(map[string]any)

		for colName, colIdx := range headerMap {
			if colIdx < len(row) && row[colIdx] != "" {
				// Convert Excel header name to database field name
				fieldName := headerToFieldMap[colName]
				if fieldName == "" {
					// If no mapping found, use the header name as-is (lowercase)
					fieldName = strings.ToLower(colName)
				}

				// Special handling for "Allow Access" field: convert "Yes"/"No" to boolean
				if colName == "Allow Access" {
					value := row[colIdx]
					if value == "Yes" || value == "yes" || value == "TRUE" || value == "true" || value == "1" {
						itemMap[fieldName] = true
					} else {
						itemMap[fieldName] = false
					}
				} else {
					itemMap[fieldName] = row[colIdx]
				}
			}
		}

		if len(itemMap) > 0 {
			item := payloadToDeviceStore(itemMap)
			items = append(items, item)
		}
	}

	// Import using repository
	successCount, skipped, errors, err := h.deviceStoreRepo.ImportDeviceStores(ctx, items)
	if err != nil {
		h.logger.Error("ImportDeviceStores failed", zap.Error(err))
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to import: %v", err)))
		return
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

// payloadToDeviceStore 函数已在 admin_device_store_impl.go 中定义，直接使用

