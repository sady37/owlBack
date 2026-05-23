package httpapi

import (
	"database/sql"
	"io"

	"wisefido-data/internal/domain"
)

// payloadToDeviceStore 将 map[string]any 转为 domain.DeviceStore（导入等）
func payloadToDeviceStore(payload map[string]any) *domain.DeviceStore {
	ds := &domain.DeviceStore{}

	if v, ok := payload["device_type"].(string); ok {
		ds.DeviceType = v
	}
	if v, ok := payload["device_uid"].(string); ok && v != "" {
		ds.DeviceUID = v
	}
	if v, ok := payload["device_code"].(string); ok && v != "" {
		ds.DeviceCode = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["device_model"].(string); ok && v != "" {
		ds.DeviceModel = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["mac"].(string); ok && v != "" {
		ds.MAC = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["imei"].(string); ok && v != "" {
		ds.IMEI = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["comm_mode"].(string); ok && v != "" {
		ds.CommMode = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["mcu_model"].(string); ok && v != "" {
		ds.MCUModel = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["firmware_version"].(string); ok && v != "" {
		ds.FirmwareVersion = sql.NullString{String: v, Valid: true}
	}
	if v, ok := payload["ota_target_firmware_version"]; ok {
		if s, isStr := v.(string); isStr && s != "" {
			ds.OTATargetFirmwareVersion = sql.NullString{String: s, Valid: true}
		} else {
			ds.OTATargetFirmwareVersion = sql.NullString{Valid: false} // null or empty → clear
		}
	}
	if v, ok := payload["ota_target_mcu_model"]; ok {
		if s, isStr := v.(string); isStr && s != "" {
			ds.OTATargetMCUModel = sql.NullString{String: s, Valid: true}
		} else {
			ds.OTATargetMCUModel = sql.NullString{Valid: false} // null or empty → clear
		}
	}
	if v, ok := payload["tenant_id"]; ok {
		if v == nil {
			ds.TenantID = "00000000-0000-0000-0000-000000000002"
		} else if str, ok := v.(string); ok {
			ds.TenantID = str
		}
	}
	if v, ok := payload["access"].(bool); ok {
		ds.Access = v
	}
	if v, ok := payload["ota_permit"]; ok {
		if s, isStr := v.(string); isStr && s != "" {
			ds.OTAPermit = sql.NullString{String: s, Valid: true}
		} else {
			ds.OTAPermit = sql.NullString{Valid: false}
		}
	}
	if v, ok := payload["ota_way"]; ok {
		if s, isStr := v.(string); isStr && s != "" {
			ds.OTAWay = sql.NullString{String: s, Valid: true}
		} else {
			ds.OTAWay = sql.NullString{Valid: false}
		}
	}
	if v, ok := payload["ota_schedule"]; ok {
		if s, isStr := v.(string); isStr && s != "" {
			ds.OTASchedule = sql.NullString{String: s, Valid: true}
		} else {
			ds.OTASchedule = sql.NullString{Valid: false}
		}
	}
	if v, ok := payload["ota_status"].(string); ok {
		if v != "" {
			ds.OTAStatus = sql.NullString{String: v, Valid: true}
		} else {
			ds.OTAStatus = sql.NullString{Valid: false}
		}
	}

	return ds
}

// payloadToDeviceStorePatch 批量更新：仅当 JSON 显式出现 device_code / access 时才更新对应列；
// device_code 可为 null 或 "" 表示置为 NULL。
func payloadToDeviceStorePatch(payload map[string]any) *domain.DeviceStore {
	ds := payloadToDeviceStore(payload)
	if payload == nil {
		return ds
	}
	if _, ok := payload["device_code"]; ok {
		ds.DeviceCodeSet = true
		if payload["device_code"] == nil {
			ds.DeviceCode = sql.NullString{Valid: false}
		} else if v, ok := payload["device_code"].(string); ok {
			if v == "" {
				ds.DeviceCode = sql.NullString{Valid: false}
			} else {
				ds.DeviceCode = sql.NullString{String: v, Valid: true}
			}
		}
	}
	if _, ok := payload["access"]; ok {
		ds.AccessSet = true
		if v, ok := payload["access"].(bool); ok {
			ds.Access = v
		}
	}
	if _, ok := payload["ota_permit"]; ok {
		ds.OTAPermitSet = true
	}
	if _, ok := payload["ota_way"]; ok {
		ds.OTAWaySet = true
	}
	if _, ok := payload["ota_schedule"]; ok {
		ds.OTAScheduleSet = true
	}
	if _, ok := payload["ota_status"]; ok {
		ds.OTAStatusSet = true
	}
	if _, ok := payload["ota_target_firmware_version"]; ok {
		ds.OTATargetFWSet = true
	}
	if _, ok := payload["ota_target_mcu_model"]; ok {
		ds.OTATargetMCUSet = true
	}
	return ds
}

// bytesReader 供 excelize OpenReader 使用
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
