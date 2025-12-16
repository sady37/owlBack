package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"wisefido-data/internal/repository"

	"github.com/xuri/excelize/v2"
)

// -------- Device Store impl --------

func (a *AdminAPI) getDeviceStores(w http.ResponseWriter, r *http.Request) {
	filters := map[string]any{
		"search":      r.URL.Query().Get("search"),
		"tenant_id":   r.URL.Query().Get("tenant_id"),
		"device_type": r.URL.Query().Get("device_type"),
		"page":        parseInt(r.URL.Query().Get("page"), 1),
		"size":        parseInt(r.URL.Query().Get("size"), 100),
	}

	items, total, err := a.DeviceStore.ListDeviceStores(r.Context(), filters)
	if err != nil {
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

func (a *AdminAPI) batchUpdateDeviceStores(w http.ResponseWriter, r *http.Request) {
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

	updates := make([]map[string]any, 0, len(updatesRaw))
	for _, u := range updatesRaw {
		if m, ok := u.(map[string]any); ok {
			deviceStoreID, _ := m["device_store_id"].(string)
			data, _ := m["data"].(map[string]any)
			if deviceStoreID != "" {
				updateItem := map[string]any{
					"device_store_id": deviceStoreID,
				}
				// Copy data fields
				for k, v := range data {
					updateItem[k] = v
				}
				updates = append(updates, updateItem)
			}
		}
	}

	if err := a.DeviceStore.BatchUpdateDeviceStores(r.Context(), updates); err != nil {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update device stores: %v", err)))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success": true,
		"updated": len(updates),
	}))
}

func (a *AdminAPI) getImportTemplate(w http.ResponseWriter, r *http.Request) {
	excelData, err := GenerateDeviceStoreImportTemplate()
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to generate template: %v", err)))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=device-store-import-template.xlsx")
	w.WriteHeader(http.StatusOK)
	w.Write(excelData)
}

func (a *AdminAPI) exportDeviceStores(w http.ResponseWriter, r *http.Request) {
	// Query data with same filters
	filters := map[string]any{
		"search":      r.URL.Query().Get("search"),
		"tenant_id":   r.URL.Query().Get("tenant_id"),
		"device_type": r.URL.Query().Get("device_type"),
		"page":        1,
		"size":        10000, // Export all (adjust if needed)
	}

	items, _, err := a.DeviceStore.ListDeviceStores(r.Context(), filters)
	if err != nil {
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
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to generate export: %v", err)))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=device-store-export.xlsx")
	w.WriteHeader(http.StatusOK)
	w.Write(excelData)
}

func (a *AdminAPI) importDeviceStores(w http.ResponseWriter, r *http.Request) {
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
	items := make([]map[string]any, 0, len(rows)-1)
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		item := make(map[string]any)

		for colName, colIdx := range headerMap {
			if colIdx < len(row) && row[colIdx] != "" {
				// Convert Excel header name to database field name
				fieldName := headerToFieldMap[colName]
				if fieldName == "" {
					// If no mapping found, use the header name as-is (lowercase)
					fieldName = colName
				}

				// Special handling for "Allow Access" field: convert "Yes"/"No" to boolean
				if colName == "Allow Access" {
					value := row[colIdx]
					if value == "Yes" || value == "yes" || value == "TRUE" || value == "true" || value == "1" {
						item[fieldName] = true
					} else {
						item[fieldName] = false
					}
				} else {
					item[fieldName] = row[colIdx]
				}
			}
		}

		if len(item) > 0 {
			items = append(items, item)
		}
	}

	// Import using repository (need to add ImportDeviceStores method)
	// For now, use batch insert approach
	successCount, errors, skipped, err := a.importDeviceStoresData(r.Context(), items)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to import: %v", err)))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"success":       true,
		"total":         len(items),
		"success_count": successCount,
		"failed_count":  len(errors),
		"skipped_count": len(skipped),
		"errors":        errors,
		"skipped":       skipped,
	}))
}

// importDeviceStoresData imports device stores data
func (a *AdminAPI) importDeviceStoresData(ctx context.Context, items []map[string]any) (int, []map[string]any, []map[string]any, error) {
	// Use repository method
	if repo, ok := a.DeviceStore.(*repository.PostgresDeviceStoreRepo); ok {
		return repo.ImportDeviceStores(ctx, items)
	}

	// Fallback: not implemented
	return 0, nil, nil, fmt.Errorf("import not supported for this repository type")
}

// bytesReader implements io.ReaderAt and io.Reader for excelize
type bytesReader struct {
	data []byte
	pos  int64
}

func (br *bytesReader) Read(p []byte) (n int, err error) {
	if br.pos >= int64(len(br.data)) {
		return 0, io.EOF
	}
	n = copy(p, br.data[br.pos:])
	br.pos += int64(n)
	return n, nil
}

func (br *bytesReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(br.data)) {
		return 0, io.EOF
	}
	n = copy(p, br.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

func (br *bytesReader) Size() int64 {
	return int64(len(br.data))
}
