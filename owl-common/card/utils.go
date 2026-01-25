package card

import (
	"encoding/json"
)

// DeviceJSON device JSON format (for cards.devices JSONB field).
// 写 cards.devices 时：DeviceJSON 需含 device_uid；组装 DeviceInfo 的查询（如 GetDevicesByBed、GetUnboundDevicesByUnit）从 devices 表选 d.device_uid 填入。
type DeviceJSON struct {
	DeviceID    string  `json:"device_id"`
	DeviceUID   string  `json:"device_uid,omitempty"`   // 与 card-overview、GetCardDevices 对齐
	DeviceCode  string  `json:"device_code,omitempty"`  // 与 card-overview、GetCardDevices 对齐
	DeviceName  string  `json:"device_name"`
	DeviceType  string  `json:"device_type"`
	DeviceModel string  `json:"device_model"`
	BedID       *string `json:"bed_id,omitempty"`
	BedName     *string `json:"bed_name,omitempty"`
	RoomID      *string `json:"room_id,omitempty"`
	RoomName    *string `json:"room_name,omitempty"`
	UnitID      string  `json:"unit_id"`
}

// ResidentJSON resident JSON format (for cards.residents JSONB field)
type ResidentJSON struct {
	ResidentID string `json:"resident_id"`
	Nickname   string `json:"nickname"`
}

// ConvertDevicesToJSON converts device list to JSON（含 device_uid、device_code，与 card-overview、GetCardDevices 对齐）
func ConvertDevicesToJSON(devices []DeviceInfo) ([]byte, error) {
	var deviceJSONs []DeviceJSON
	for _, device := range devices {
		deviceJSONs = append(deviceJSONs, DeviceJSON{
			DeviceID:    device.DeviceID,
			DeviceUID:   device.DeviceUID,
			DeviceCode:  device.DeviceCode,
			DeviceName:  device.DeviceName,
			DeviceType:  device.DeviceType,
			DeviceModel: device.DeviceModel,
			BedID:       device.BoundBedID,
			BedName:     device.BedName,
			RoomID:      device.BoundRoomID,
			RoomName:    device.RoomName,
			UnitID:      device.UnitID,
		})
	}
	return json.Marshal(deviceJSONs)
}

// ConvertResidentsToJSON converts resident list to JSON
func ConvertResidentsToJSON(residents []ResidentInfo) ([]byte, error) {
	var residentJSONs []ResidentJSON
	for _, resident := range residents {
		residentJSONs = append(residentJSONs, ResidentJSON{
			ResidentID: resident.ResidentID,
			Nickname:   resident.Nickname,
		})
	}
	return json.Marshal(residentJSONs)
}

