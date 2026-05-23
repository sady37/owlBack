package httpapi

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"

	"github.com/xuri/excelize/v2"
)

func deviceStoreImportHeaderToField() map[string]string {
	return map[string]string{
		"Device Type":                 "device_type",
		"Device Model":                "device_model",
		"Device Code":                 "device_code",
		"Device UID":                  "device_uid",
		"UID":                         "device_uid",
		"MAC":                         "mac",
		"IMEI":                        "imei",
		"Comm Mode":                   "comm_mode",
		"MCU Model":                   "mcu_model",
		"Firmware Version":            "firmware_version",
		"OTA Target Firmware Version": "ota_target_firmware_version",
		"OTA Target MCU Model":        "ota_target_mcu_model",
		"Tenant ID":                   "tenant_id",
		"Tenant Name":                 "tenant_name",
		"Allow Access":                "access",
		"Import Date":                 "import_date",
		"Allocate Time":               "allocate_time",
	}
}

// loadDeviceStoreImportRows reads the first sheet as rows (header + data). .csv uses encoding/csv; otherwise excelize.
func loadDeviceStoreImportRows(fileBytes []byte, filename string) ([][]string, error) {
	name := strings.ToLower(strings.TrimSpace(filename))
	if strings.HasSuffix(name, ".csv") {
		return readDeviceStoreImportCSV(fileBytes)
	}
	return readDeviceStoreImportExcel(fileBytes)
}

func readDeviceStoreImportCSV(data []byte) ([][]string, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	r := csv.NewReader(bytes.NewReader(data))
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	for i := range rows {
		for j := range rows[i] {
			rows[i][j] = strings.TrimSpace(rows[i][j])
		}
	}
	return rows, nil
}

func readDeviceStoreImportExcel(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(&bytesReader{data: fileBytes})
	if err != nil {
		return nil, fmt.Errorf("failed to parse Excel file: %w", err)
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel file has no sheets")
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}
	return rows, nil
}

// parseDeviceStoreImportRowsToItems maps spreadsheet rows to domain items (rows[0] = header).
func parseDeviceStoreImportRowsToItems(rows [][]string) []*domain.DeviceStore {
	if len(rows) < 2 {
		return nil
	}
	headerRow := rows[0]
	headerMap := make(map[string]int)
	for i, h := range headerRow {
		headerMap[strings.TrimSpace(h)] = i
	}
	headerToField := deviceStoreImportHeaderToField()
	items := make([]*domain.DeviceStore, 0, len(rows)-1)
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		itemMap := make(map[string]any)
		for colName, colIdx := range headerMap {
			if colIdx < len(row) && row[colIdx] != "" {
				fieldName := headerToField[colName]
				if fieldName == "" {
					fieldName = strings.ToLower(colName)
				}
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
			items = append(items, payloadToDeviceStore(itemMap))
		}
	}
	return items
}
