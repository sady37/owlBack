package httpapi

import (
	"bytes"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// DeviceStoreImportHeader 导入模板表头（只包含物理属性和固件版本）
var DeviceStoreImportHeader = []string{
	"Device Type",
	"Device Model",
	"Serial Number",
	"UID",
	"IMEI",
	"Comm Mode",
	"MCU Model",
	"Firmware Version",
}

// DeviceStoreExportHeader 导出表头（包含所有字段）
var DeviceStoreExportHeader = []string{
	"Device Type",
	"Device Model",
	"Serial Number",
	"UID",
	"IMEI",
	"Comm Mode",
	"MCU Model",
	"Firmware Version",
	"OTA Target Firmware Version",
	"OTA Target MCU Model",
	"Tenant Name",
	"Allow Access",
	"Import Date",
	"Allocate Time",
}

// GenerateDeviceStoreImportTemplate 生成导入模板 Excel 文件（只包含物理属性字段）
func GenerateDeviceStoreImportTemplate() ([]byte, error) {
	return generateDeviceStoreExcel(DeviceStoreImportHeader, []map[string]any{}, false)
}

// GenerateDeviceStoreExport 生成设备库存导出 Excel 文件（包含所有字段）
// data: 设备数据列表，如果为空则只生成表头
func GenerateDeviceStoreExport(data []map[string]any) ([]byte, error) {
	return generateDeviceStoreExcel(DeviceStoreExportHeader, data, true)
}

// generateDeviceStoreExcel 生成设备库存 Excel 文件的通用函数
// headers: 表头列表
// data: 设备数据列表，如果为空则只生成表头
// includeAllFields: 是否包含所有字段（用于区分导入和导出）
func generateDeviceStoreExcel(headers []string, data []map[string]any, includeAllFields bool) ([]byte, error) {
	f := excelize.NewFile()
	// Note: Don't defer Close() here, because WriteTo needs the file to be open

	sheetName := "Device Store"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		f.Close() // Close on error
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}

	// 删除默认的 Sheet1
	f.DeleteSheet("Sheet1")

	// 设置默认活动工作表
	f.SetActiveSheet(index)

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6F3FF"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	// 写入表头
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to convert coordinates: %w", err)
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to set header cell %s: %w", cell, err)
		}
		if err := f.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to set header style: %w", err)
		}
	}

	// 设置列宽（根据字段数量动态设置）
	columnWidths := []float64{
		15, // Device Type
		20, // Device Model
		20, // Serial Number
		20, // UID
		20, // IMEI
		15, // Comm Mode
		15, // MCU Model
		20, // Firmware Version
		25, // OTA Target Firmware Version (if included)
		20, // OTA Target MCU Model (if included)
		38, // Tenant ID (if included)
		15, // Allow Access (if included)
		20, // Import Date (if included)
		20, // Allocate Time (if included)
	}
	for i := 0; i < len(headers); i++ {
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to convert column number: %w", err)
		}
		if i < len(columnWidths) && columnWidths[i] > 0 {
			if err := f.SetColWidth(sheetName, col, col, columnWidths[i]); err != nil {
				f.Close()
				return nil, fmt.Errorf("failed to set column width: %w", err)
			}
		}
	}

	// 写入数据
	for rowIdx, item := range data {
		row := rowIdx + 2 // 从第2行开始（第1行是表头）
		colIdx := 0

		// 根据表头顺序写入数据
		for _, header := range headers {
			colIdx++
			var value interface{}

			switch header {
			case "Device Type":
				value = getStringValue(item, "device_type")
			case "Device Model":
				value = getStringValue(item, "device_model")
			case "Serial Number":
				value = getStringValue(item, "serial_number")
			case "UID":
				value = getStringValue(item, "uid")
			case "IMEI":
				value = getStringValue(item, "imei")
			case "Comm Mode":
				value = getStringValue(item, "comm_mode")
			case "MCU Model":
				value = getStringValue(item, "mcu_model")
			case "Firmware Version":
				value = getStringValue(item, "firmware_version")
			case "OTA Target Firmware Version":
				value = getStringValue(item, "ota_target_firmware_version")
			case "OTA Target MCU Model":
				value = getStringValue(item, "ota_target_mcu_model")
			case "Tenant Name":
				value = getStringValue(item, "tenant_name")
			case "Allow Access":
				if val, ok := item["allow_access"].(bool); ok {
					if val {
						value = "Yes"
					} else {
						value = "No"
					}
				}
			case "Import Date":
				value = getTimeValue(item, "import_date")
			case "Allocate Time":
				value = getTimeValue(item, "allocate_time")
			}

			if value != nil && value != "" {
				if err := setCellValue(f, sheetName, colIdx, row, value); err != nil {
					f.Close()
					return nil, fmt.Errorf("failed to set cell value at row %d, col %d: %w", row, colIdx, err)
				}
			}
		}
	}

	// 冻结表头
	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to freeze panes: %w", err)
	}

	// Write to bytes buffer
	// Note: File must remain open during Write operation
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to write to buffer: %w", err)
	}

	// Close file after writing
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	return buf.Bytes(), nil
}

// setCellValue 设置单元格值
func setCellValue(f *excelize.File, sheet string, col, row int, value interface{}) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return err
	}
	return f.SetCellValue(sheet, cell, value)
}

// getStringValue 从 map 中获取字符串值
func getStringValue(item map[string]any, key string) string {
	if val, ok := item[key].(string); ok && val != "" {
		return val
	}
	return ""
}

// getTimeValue 从 map 中获取时间值并格式化为字符串
func getTimeValue(item map[string]any, key string) string {
	if val, ok := item[key].(string); ok && val != "" {
		return val
	}
	if val, ok := item[key].(time.Time); ok {
		return val.Format("2006-01-02 15:04:05")
	}
	return ""
}
