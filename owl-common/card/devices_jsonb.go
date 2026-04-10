package card

import (
	"encoding/json"
	"fmt"
)

// ParseDevicesFromCardsJSONB 解析 cards.devices JSONB，与存盘键名一致（bed_id、room_id、device_uid 等）。
// cards 表 JSON 由 convertDevicesToJSON 写入 bed_id/room_id，与 DeviceInfo 的 JSON 标签 bound_* 不一致，
// 不能直接 json.Unmarshal 到 []DeviceInfo。与 wisefido-data/repository/postgres_card.deserializeDevices 对齐。
func ParseDevicesFromCardsJSONB(data []byte) ([]DeviceInfo, error) {
	if len(data) == 0 || string(data) == "[]" {
		return []DeviceInfo{}, nil
	}

	type DeviceJSON struct {
		DeviceID            string  `json:"device_id"`
		DeviceUID           string  `json:"device_uid,omitempty"`
		DeviceCode          string  `json:"device_code,omitempty"`
		DeviceName          string  `json:"device_name"`
		DeviceType          any     `json:"device_type"`
		DeviceModel         string  `json:"device_model"`
		BedID               *string `json:"bed_id,omitempty"`
		RoomID              *string `json:"room_id,omitempty"`
		BoundBedID          *string `json:"bound_bed_id,omitempty"`
		BoundRoomID         *string `json:"bound_room_id,omitempty"`
		UnitID              string  `json:"unit_id"`
		MonitoringEnabled   *bool   `json:"monitoring_enabled,omitempty"`
		Status              string  `json:"status,omitempty"`
	}

	var deviceJSONs []DeviceJSON
	if err := json.Unmarshal(data, &deviceJSONs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices: %w", err)
	}

	devices := make([]DeviceInfo, 0, len(deviceJSONs))
	for _, dj := range deviceJSONs {
		var deviceType string
		switch v := dj.DeviceType.(type) {
		case string:
			deviceType = v
		case float64:
			deviceType = fmt.Sprintf("%.0f", v)
		default:
			deviceType = fmt.Sprint(v)
		}
		bedPtr := coalesceNonEmptyPtr(dj.BedID, dj.BoundBedID)
		roomPtr := coalesceNonEmptyPtr(dj.RoomID, dj.BoundRoomID)
		mon := false
		if dj.MonitoringEnabled != nil {
			mon = *dj.MonitoringEnabled
		}
		devices = append(devices, DeviceInfo{
			DeviceID:          dj.DeviceID,
			DeviceUID:         dj.DeviceUID,
			DeviceCode:        dj.DeviceCode,
			DeviceName:        dj.DeviceName,
			DeviceType:        deviceType,
			DeviceModel:       dj.DeviceModel,
			BoundBedID:        bedPtr,
			BoundRoomID:       roomPtr,
			UnitID:            dj.UnitID,
			MonitoringEnabled: mon,
			Status:            dj.Status,
		})
	}
	return devices, nil
}

func coalesceNonEmptyPtr(a, b *string) *string {
	if a != nil && *a != "" {
		return a
	}
	if b != nil && *b != "" {
		return b
	}
	if a != nil {
		return a
	}
	return b
}
