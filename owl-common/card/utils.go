package card

import (
	"encoding/json"
)

// DeviceJSON device JSON format (for cards.devices JSONB field)
type DeviceJSON struct {
	DeviceID    string  `json:"device_id"`
	DeviceName  string  `json:"device_name"`
	DeviceType  string  `json:"device_type"`
	DeviceModel string  `json:"device_model"`
	BedID       *string `json:"bed_id,omitempty"`    // Bed ID where device is bound (if bound to bed)
	BedName     *string `json:"bed_name,omitempty"`  // Bed name (if bound to bed)
	RoomID      *string `json:"room_id,omitempty"`   // Room ID where device is bound (if bound to room)
	RoomName    *string `json:"room_name,omitempty"` // Room name (if bound to room)
	UnitID      string  `json:"unit_id"`             // Unit ID where device is bound
}

// ResidentJSON resident JSON format (for cards.residents JSONB field)
type ResidentJSON struct {
	ResidentID string `json:"resident_id"`
	Nickname   string `json:"nickname"`
}

// ConvertDevicesToJSON converts device list to JSON
func ConvertDevicesToJSON(devices []DeviceInfo) ([]byte, error) {
	var deviceJSONs []DeviceJSON
	for _, device := range devices {
		deviceJSONs = append(deviceJSONs, DeviceJSON{
			DeviceID:    device.DeviceID,
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

